package main

import (
	"os"
	"strings"

	"github.com/gladiusio/gladius-network-gateway/config"
	"github.com/gladiusio/gladius-network-gateway/pkg/gateway"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	// Setup config
	config.SetupConfig()

	// Setup logging
	setupLogger()

	g := gateway.New(viper.GetString("API.Port"))
	g.Start()

	select {}
}

func setupLogger() {
	// Setup logging level
	switch loglevel := viper.GetString("Log.Level"); strings.ToLower(loglevel) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "disabled":
		zerolog.SetGlobalLevel(zerolog.Disabled)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if viper.GetBool("Log.Pretty") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
