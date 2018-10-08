package gateway

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/spf13/viper"

	"github.com/gladiusio/gladius-common/pkg/blockchain"
	"github.com/gladiusio/gladius-common/pkg/handlers"
	"github.com/gladiusio/gladius-controld/pkg/p2p/peer"
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
	baseRouter.NotFoundHandler = http.HandlerFunc(handlers.NotFoundHandler)

	peerStruct := peer.New(g.ga)
	p2pRouter := baseRouter.PathPrefix("/p2p").Subrouter().StrictSlash(true)
	// P2P Message Routes
	p2pRouter.HandleFunc("/message/sign", handlers.CreateSignedMessageHandler(g.ga)).
		Methods(http.MethodPost)
	p2pRouter.HandleFunc("/message/verify", handlers.VerifySignedMessageHandler).
		Methods("POST")
	p2pRouter.HandleFunc("/network/join", handlers.JoinHandler(peerStruct)).
		Methods("POST")
	p2pRouter.HandleFunc("/network/leave", handlers.LeaveHandler(peerStruct)).
		Methods("POST")
	p2pRouter.HandleFunc("/state/push_message", handlers.PushStateMessageHandler(peerStruct)).
		Methods("POST")
	p2pRouter.HandleFunc("/state", handlers.GetFullStateHandler(peerStruct)).
		Methods("GET")
	p2pRouter.HandleFunc("/state/node/{node_address}", handlers.GetNodeStateHandler(peerStruct)).
		Methods("GET")
	p2pRouter.HandleFunc("/state/signatures", handlers.GetSignatureListHandler(peerStruct)).
		Methods("GET")
	p2pRouter.HandleFunc("/state/content_diff", handlers.GetContentNeededHandler(peerStruct)).
		Methods("POST")
	p2pRouter.HandleFunc("/state/content_links", handlers.GetContentLinksHandler(peerStruct)).
		Methods("POST")

	// Only enable for testing
	if viper.GetBool("NodeManager.Config.Debug") {
		p2pRouter.HandleFunc("/state/set_state", handlers.SetStateDebugHandler(peerStruct)).
			Methods("POST")
	}

	// Blockchain account management endpoints
	accountRouter := baseRouter.PathPrefix("/account/{address:0[xX][0-9a-fA-F]{40}}").Subrouter().StrictSlash(true)
	accountRouter.HandleFunc("/balance/{symbol:[a-z]{3}}", handlers.AccountBalanceHandler)
	accountRouter.HandleFunc("/transactions", handlers.AccountTransactionsHandler).
		Methods(http.MethodPost)

	// Local wallet management
	walletRouter := baseRouter.PathPrefix("/keystore").Subrouter().StrictSlash(true)
	walletRouter.HandleFunc("/account/create", handlers.KeystoreAccountCreationHandler(g.ga)).
		Methods(http.MethodPost)
	walletRouter.HandleFunc("/account", handlers.KeystoreAccountRetrievalHandler(g.ga))
	walletRouter.HandleFunc("/account/open", handlers.KeystoreAccountUnlockHandler(g.ga)).
		Methods(http.MethodPost)

	// Transaction status endpoints
	statusRouter := baseRouter.PathPrefix("/status").Subrouter().StrictSlash(true)
	statusRouter.HandleFunc("/", handlers.StatusHandler).
		Methods(http.MethodGet, http.MethodPut).
		Name("status")
	statusRouter.HandleFunc("/tx/{tx:0[xX][0-9a-fA-F]{64}}", handlers.StatusTxHandler).
		Methods(http.MethodGet).
		Name("status-tx")

	// Node pool application routes
	nodeApplicationRouter := baseRouter.PathPrefix("/node").Subrouter().StrictSlash(true)
	// Node pool applications
	nodeApplicationRouter.HandleFunc("/applications", handlers.NodeViewAllApplicationsHandler(g.ga)).
		Methods(http.MethodGet)
	// Node application to Pool
	nodeApplicationRouter.HandleFunc("/applications/{poolAddress:0[xX][0-9a-fA-F]{40}}/new", handlers.NodeNewApplicationHandler(g.ga)).
		Methods(http.MethodPost)
	nodeApplicationRouter.HandleFunc("/applications/{poolAddress:0[xX][0-9a-fA-F]{40}}/view", handlers.NodeViewApplicationHandler(g.ga)).
		Methods(http.MethodGet)

	// Pool listing routes
	poolRouter := baseRouter.PathPrefix("/pool").Subrouter().StrictSlash(true)
	// Retrieve owned Pool if available
	poolRouter.HandleFunc("/", nil)
	// Pool Retrieve Data
	poolRouter.HandleFunc("/{poolAddress:0[xX][0-9a-fA-F]{40}}", handlers.PoolPublicDataHandler(g.ga)).
		Methods(http.MethodGet)

	// Market Sub-Routes
	marketRouter := baseRouter.PathPrefix("/market").Subrouter().StrictSlash(true)
	marketRouter.HandleFunc("/pools", handlers.MarketPoolsHandler(g.ga))
}
