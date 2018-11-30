package peer

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/buger/jsonparser"
	"github.com/gladiusio/legion/network"

	"github.com/gladiusio/gladius-network-gateway/pkg/p2p/signature"
	"github.com/gladiusio/gladius-network-gateway/pkg/p2p/state"
)

// StatePlugin handles incoming messages related to the network state
type StatePlugin struct {
	network.GenericPlugin
	peerState *state.State
}

// NewMessage is called every time a new message is received
func (state *StatePlugin) NewMessage(ctx *network.MessageContext) {
	switch ctx.Message.Type() {
	case "state_update":
		sm, err := parseSignedMessage(ctx.Message.Body())
		if err == nil {
			go state.peerState.UpdateState(sm)
		} else {
			fmt.Println(err)
		}
	case "sync_request":
		smList := state.peerState.GetSignatureList()
		smStringList := make([]string, 0)
		for _, sm := range smList {
			smString, _ := json.Marshal(sm)
			smStringList = append(smStringList, string(smString))
		}
		b, _ := json.Marshal(smStringList)
		ctx.Reply(ctx.Legion.NewMessage("sync_response", b))
	case "sync_response":
		smListBytes := ctx.Message.Body()
		jsonparser.ArrayEach(smListBytes, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			sm, err := parseSignedMessage(value)
			if err != nil {
				return
			}
			go state.peerState.UpdateState(sm)
		})
	}
}

// Startup is called once the network is started. Every 60
// seconds we ask a random peer for it's state. This is an anti
// entropy method that might not be entirely needed.
func (state *StatePlugin) Startup(ctx *network.NetworkContext) {
	go func() {
		for {
			time.Sleep(60 * time.Second)
			ctx.Legion.BroadcastRandom(ctx.Legion.NewMessage("sync_request", []byte{}), 1)
		}
	}()
}

// PeerAdded is called when a new peer connects or is added
func (state *StatePlugin) PeerAdded(ctx *network.PeerContext) {
	ctx.Legion.PromotePeer(ctx.Peer.Remote())
}

func parseSignedMessage(smBytes []byte) (*signature.SignedMessage, error) {
	messageBytes, _, _, err := jsonparser.Get(smBytes, "message")
	if err != nil {
		return nil, errors.New("Can't find `message` in body")
	}

	hash, err := jsonparser.GetString(smBytes, "hash")
	if err != nil {
		return nil, errors.New("Can't find `hash` in body")
	}

	signatureString, err := jsonparser.GetString(smBytes, "signature")
	if err != nil {
		return nil, errors.New("Could not find `signature` in body")

	}

	address, err := jsonparser.GetString(smBytes, "address")
	if err != nil {
		return nil, errors.New("Could not find `address` in body")
	}

	parsed, err := signature.ParseSignedMessage(string(messageBytes), hash, signatureString, address)
	if err != nil {
		return nil, errors.New("Couldn't parse body")

	}

	return parsed, nil
}
