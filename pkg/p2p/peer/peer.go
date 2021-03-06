package peer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/gladiusio/gladius-common/pkg/blockchain"
	"github.com/gladiusio/gladius-network-gateway/pkg/p2p/message"
	"github.com/gladiusio/gladius-network-gateway/pkg/p2p/signature"
	"github.com/gladiusio/gladius-network-gateway/pkg/p2p/state"
	"github.com/gladiusio/legion"
	"github.com/gladiusio/legion/network"
	"github.com/gladiusio/legion/utils"
	"github.com/rs/zerolog/log"

	"github.com/gladiusio/legion/plugins/simpledisc"
	"github.com/spf13/viper"
)

// New returns a new peer type
func New(ga *blockchain.GladiusAccountManager) *Peer {
	// Setup our state and register accepted fields
	s := state.New()
	s.RegisterNodeSingleFields("ip_address", "content_port", "heartbeat", "http_port")
	s.RegisterNodeListFields("disk_content")

	s.RegisterPoolListFields("required_content")

	conf := legion.DefaultConfig(viper.GetString("P2P.BindAddress"), uint16(viper.GetInt("P2P.BindPort")))
	// Set up the advertise address
	conf.AdvertiseAddress = utils.NewLegionAddress(viper.GetString("P2P.AdvertiseAddress"), uint16(viper.GetInt("P2P.AdvertisePort")))
	l := legion.New(conf)

	disc := new(simpledisc.Plugin)

	l.RegisterPlugin(disc)
	// Create our state plugin
	statePlugin := new(StatePlugin)
	statePlugin.peerState = s
	l.RegisterPlugin(statePlugin)

	go func() {
		err := l.Listen()
		if err != nil {
			log.Error().Err(err).Msg("Error listenting")
		}
	}()

	peer := &Peer{
		ga:        ga,
		discovery: disc,
		peerState: s,
		net:       l,
		running:   true,
		mux:       sync.Mutex{},
	}
	return peer
}

// Peer is a type that represents a peer in the Gladius p2p network.
type Peer struct {
	ga        *blockchain.GladiusAccountManager
	peerState *state.State
	net       *network.Legion
	running   bool
	discovery *simpledisc.Plugin
	mux       sync.Mutex
}

// Join will request to join the network from a specific node
func (p *Peer) Join(addressList []string) error {
	if viper.GetString("P2P.BindAddress") == "" {
		return errors.New("can't join network, bind address is not correctly detected or set")
	}
	if viper.GetString("P2P.AdvertiseAddress") == "" {
		return errors.New("can't join network, advertise address is not correctly detected or set")
	}
	addrs := make([]utils.LegionAddress, 0)
	for _, addrString := range addressList {
		addr := utils.LegionAddressFromString(addrString)
		if !addr.IsValid() {
			return fmt.Errorf("invalid address string provided: %s", addrString)
		}
		addrs = append(addrs, addr)
	}
	err := p.net.PromotePeer(addrs[0])
	if err != nil {
		return err
	}
	p.discovery.Bootstrap()
	go func() {
		time.Sleep(1 * time.Second)
		p.net.Broadcast(p.net.NewMessage("sync_request", []byte{}), addrs...)
	}()
	return nil
}

// UnlockWallet unlocks the local peer's wallet
func (p *Peer) UnlockWallet(password string) error {
	_, err := p.ga.UnlockAccount(password)
	return err
}

// SignMessage signs the message with the peer's internal account manager
func (p *Peer) SignMessage(m *message.Message) (*signature.SignedMessage, error) {
	return signature.CreateSignedMessage(m, p.ga)
}

// Stop will stop the peer
func (p *Peer) Stop() {
	p.net.Stop()
}

// SetState sets the internal state of the peer without validation
func (p *Peer) SetState(s *state.State) {
	p.mux.Lock()
	p.peerState = s
	p.mux.Unlock()
}

// UpdateAndPushState updates the local state and pushes it to several other peers
func (p *Peer) UpdateAndPushState(sm *signature.SignedMessage) error {
	err := p.GetState().UpdateState(sm)
	if err != nil {
		return err
	}

	signedBytes, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	toSend := p.net.NewMessage("state_update", signedBytes)

	p.net.Broadcast(toSend)

	return nil
}

// GetState returns the current local state
func (p *Peer) GetState() *state.State {
	return p.peerState
}

// CompareContent compares the content provided with the content in the state
// and returns a list of the missing files names in the format of:
// website/<"asset" or "route">/filename
func (p *Peer) CompareContent(contentList []string) []interface{} {
	// Convert to an interface array
	cl := make([]interface{}, len(contentList))
	for i, v := range contentList {
		cl[i] = v
	}
	contentWeHaveSet := mapset.NewSetFromSlice(cl)

	contentField := p.GetState().GetPoolField("required_content")
	if contentField == nil {
		return make([]interface{}, 0)
	}
	contentFromPool := contentField.(*state.SignedList).Data

	// Convert to an interface array
	s := make([]interface{}, len(contentFromPool))
	for i, v := range contentFromPool {
		s[i] = v
	}

	// Create a set
	contentWeNeed := mapset.NewSetFromSlice(s)

	// Return the difference of the two
	return contentWeNeed.Difference(contentWeHaveSet).ToSlice()
}

// GetContentLinks returns a map mapping a file name to all the places it can
// be found on the network
func (p *Peer) GetContentLinks(contentList []string) map[string][]string {
	allContent := p.GetState().GetNodeFieldsMap("disk_content")
	toReturn := make(map[string][]string)
	for nodeAddress, diskContent := range allContent {
		ourContent := diskContent.(*state.SignedList).Data
		// Convert to an interface array
		s := make([]interface{}, len(ourContent))
		for i, v := range ourContent {
			s[i] = v
		}
		ourContentSet := mapset.NewSetFromSlice(s)
		// Check to see if the current node we're iterating over has any of the
		// content we want
		for _, contentWanted := range contentList {
			if ourContentSet.Contains(contentWanted) {
				if toReturn[contentWanted] == nil {
					toReturn[contentWanted] = make([]string, 0)
				}
				// Add the URL to the map
				link := p.createContentLink(nodeAddress, contentWanted)
				if link != "" {
					toReturn[contentWanted] = append(toReturn[contentWanted], link)
				}
			}
		}
	}
	return toReturn
}

// Builds a URL to a node
func (p *Peer) createContentLink(nodeAddress, contentFileName string) string {
	nodeIP := p.GetState().GetNodeField(nodeAddress, "ip_address")
	nodePort := p.GetState().GetNodeField(nodeAddress, "http_port")
	if nodeIP == nil || nodePort == nil {
		return ""
	}

	nodeIP = nodeIP.(*state.SignedField).Data
	nodePort = nodePort.(*state.SignedField).Data

	contentData := strings.Split(contentFileName, "/")
	u := url.URL{}
	if nodeIP == nil || nodePort == nil || nodeIP == "" || nodePort == "" {
		return ""
	}
	u.Host = nodeIP.(string) + ":" + nodePort.(string)
	u.Path = "/content"
	u.Scheme = "http"

	if len(contentData) == 2 {
		q := u.Query()
		q.Add("website", contentData[0]) // website name
		q.Add("asset", contentData[1])   // "asset" to name of file
		u.RawQuery = q.Encode()
		return u.String()
	}
	return ""
}
