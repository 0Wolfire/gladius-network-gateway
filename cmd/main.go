package main

import (
	"os"
	"strings"

	"github.com/gladiusio/gladius-network-gateway/config"
	"github.com/gladiusio/gladius-network-gateway/pkg/gateway"
	ipify "github.com/rdegges/go-ipify"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	upnp "gitlab.com/NebulousLabs/go-upnp"
)

func main() {
	// Setup config
	config.SetupConfig()

	// Setup logging
	setupLogger()

	if viper.GetString("P2P.AdvertiseAddress") == "" {
		if viper.GetBool("UPNPEnabled") {
			log.Debug().Msg("Using UPNP to detect external IP")
			// connect to router
			d, err := upnp.Discover()
			if err != nil {
				log.Fatal().Err(err).Msg("UPNP is set to enabled, but cannot connect to service on gateway. Try enabling it on your router.")
			}

			// discover external IP
			ip, err := d.ExternalIP()
			if err != nil {
				log.Fatal().Err(err).Msg("Unable to get external IP from UPNP")
			}

			config.ConfigOption("P2P.AdvertiseAddress", ip)

			err = d.Forward(uint16(viper.GetInt("P2P.BindPort")), "Gladius Legion Port")
			if err != nil {
				log.Fatal().Int("port", viper.GetInt("P2P.BindPort")).Err(err).Msg("Error forwarding prot")
			}
		} else {
			log.Debug().Msg("Using remote service to detect external IP")
			ip, err := ipify.GetIp()
			if err != nil {
				log.Error().Err(err).Msg("Error getting IP address from remote service, peer to peer disabled")
			} else {
				config.ConfigOption("P2P.AdvertiseAddress", ip)
			}
		}
	}

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
