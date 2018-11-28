package gateway

import (
	"net/http"

	"github.com/gladiusio/gladius-common/pkg/routing"

	"github.com/rs/zerolog/log"

	"github.com/spf13/viper"

	"github.com/gladiusio/gladius-common/pkg/blockchain"
	chandlers "github.com/gladiusio/gladius-common/pkg/handlers"
	lhandlers "github.com/gladiusio/gladius-network-gateway/pkg/gateway/handlers"
	"github.com/gladiusio/gladius-network-gateway/pkg/p2p/peer"
	"github.com/gorilla/mux"
)

func New(port string) *Gateway {
	return &Gateway{
		ga:     blockchain.NewGladiusAccountManager(),
		router: mux.NewRouter(),
		port:   port,
	}
}

type Gateway struct {
	ga     *blockchain.GladiusAccountManager
	router *mux.Router
	port   string
}

func (g *Gateway) Start() {
	g.addMiddleware()
	g.addRoutes()

	// Slighlty confusing but will make /test/ redirect to /test (note the no
	// trailing slash)
	g.router.StrictSlash(true)

	// Listen locally and setup CORS
	go func() {
		err := http.ListenAndServe(":"+g.port, g.router)
		if err != nil {
			log.Fatal().Err(err).Msg("Error starting API")
		}
	}()

	log.Info().Msg("Started API at http://localhost:" + g.port)
}

func (g *Gateway) addMiddleware() {
	addLogging(g.router)
	g.router.Use(responseMiddleware) // Add "application/json" if POST request
}

func (g *Gateway) addRoutes() {
	// Create a base router with "/api"
	baseRouter := g.router.PathPrefix("/api").Subrouter().StrictSlash(true)
	baseRouter.NotFoundHandler = http.HandlerFunc(chandlers.NotFoundHandler)

	peerStruct := peer.New(g.ga)
	p2pRouter := baseRouter.PathPrefix("/p2p").Subrouter().StrictSlash(true)
	// P2P Message Routes
	p2pRouter.HandleFunc("/message/sign", lhandlers.CreateSignedMessageHandler(g.ga)).
		Methods(http.MethodPost)
	p2pRouter.HandleFunc("/message/verify", lhandlers.VerifySignedMessageHandler).
		Methods("POST")
	p2pRouter.HandleFunc("/network/join", lhandlers.JoinHandler(peerStruct)).
		Methods("POST")
	p2pRouter.HandleFunc("/network/leave", lhandlers.LeaveHandler(peerStruct)).
		Methods("POST")
	p2pRouter.HandleFunc("/state/push_message", lhandlers.PushStateMessageHandler(peerStruct)).
		Methods("POST")
	p2pRouter.HandleFunc("/state", lhandlers.GetFullStateHandler(peerStruct)).
		Methods("GET")
	p2pRouter.HandleFunc("/state/node/{node_address}", lhandlers.GetNodeStateHandler(peerStruct)).
		Methods("GET")
	p2pRouter.HandleFunc("/state/signatures", lhandlers.GetSignatureListHandler(peerStruct)).
		Methods("GET")
	p2pRouter.HandleFunc("/state/content_diff", lhandlers.GetContentNeededHandler(peerStruct)).
		Methods("POST")
	p2pRouter.HandleFunc("/state/content_links", lhandlers.GetContentLinksHandler(peerStruct)).
		Methods("POST")

	// Only enable for testing
	if viper.GetBool("NodeManager.Config.Debug") {
		p2pRouter.HandleFunc("/state/set_state", lhandlers.SetStateDebugHandler(peerStruct)).
			Methods("POST")
	}

	// Blockchain account management endpoints
	routing.AppendAccountManagementEndpoints(baseRouter)

	// Local wallet management
	routing.AppendWalletManagementEndpoints(baseRouter, g.ga)

	// Transaction status endpoints
	routing.AppendStatusEndpoints(baseRouter)

	// Node pool application routes
	nodeApplicationRouter := baseRouter.PathPrefix("/node").Subrouter().StrictSlash(true)
	// Node pool applications
	nodeApplicationRouter.HandleFunc("/applications", lhandlers.NodeViewAllApplicationsHandler(g.ga)).
		Methods(http.MethodGet)
	// Node application to Pool
	nodeApplicationRouter.HandleFunc("/applications/{poolAddress:0[xX][0-9a-fA-F]{40}}/new", lhandlers.NodeNewApplicationHandler(g.ga)).
		Methods(http.MethodPost)
	nodeApplicationRouter.HandleFunc("/applications/{poolAddress:0[xX][0-9a-fA-F]{40}}/view", lhandlers.NodeViewApplicationHandler(g.ga)).
		Methods(http.MethodGet)

	// Pool listing routes
	poolRouter := baseRouter.PathPrefix("/pool").Subrouter().StrictSlash(true)
	// Retrieve owned Pool if available
	poolRouter.HandleFunc("/", nil)
	// Pool Retrieve Data
	poolRouter.HandleFunc("/{poolAddress:0[xX][0-9a-fA-F]{40}}", chandlers.PoolPublicDataHandler(g.ga)).
		Methods(http.MethodGet)

	// Market Sub-Routes
	marketRouter := baseRouter.PathPrefix("/market").Subrouter().StrictSlash(true)
	marketRouter.HandleFunc("/pools", chandlers.MarketPoolsHandler(g.ga))

	routing.AppendVersionEndpoints(baseRouter, "0.8.0")
}
