package config

import (
	"path/filepath"
	"strings"

	"github.com/gladiusio/gladius-common/pkg/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// SetupConfig sets up viper and adds our config options
func SetupConfig() {
	base, err := utils.GetGladiusBase()
	if err != nil {
		log.Warn().Err(err).Msg("Error retrieving base directory")
	}

	// Add config file name and searching
	viper.SetConfigName("gladius-network-gateway")
	viper.AddConfigPath(base)

	// Setup env variable handling
	viper.SetEnvPrefix("GATEWAY")
	r := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(r)
	viper.AutomaticEnv()

	// Load config
	err = viper.ReadInConfig()
	if err != nil {
		log.Warn().Err(err).Msg("Error reading config file, it may not exist or is corrupted. Using defaults.")
	}

	// Build our config options
	buildOptions(base)
}

func buildOptions(base string) {
	// Log options
	ConfigOption("Log.Level", "debug")
	ConfigOption("Log.Pretty", true)

	// P2P options
	ConfigOption("P2P.BindAddress", "0.0.0.0")
	ConfigOption("P2P.BindPort", 7947)
	ConfigOption("P2P.AdvertiseAddress", "")
	ConfigOption("P2P.AdvertisePort", 7947)
	ConfigOption("P2P.MessageVerifyOverride", false)

	// Blockchain options
	ConfigOption("Blockchain.Provider", "https://mainnet.infura.io/tjqLYxxGIUp0NylVCiWw")
	ConfigOption("Blockchain.MarketAddress", "0x27a9390283236f836a0b3c8dfdbed2ed854322fc")
	ConfigOption("Blockchain.PoolUrl", "http://174.138.111.1/api/")
	ConfigOption("Blockchain.PoolManagerAddress", "0x9717EaDbfE344457135a4f1fA8AE3B11B4CAB0b7")

	// Wallet options
	ConfigOption("Wallet.Directory", filepath.Join(base, "wallet"))

	// API options
	ConfigOption("API.Port", "3001")
	ConfigOption("API.DebugRequests", false)
	ConfigOption("API.RemoteConnectionsAllowed", false)

	// Misc.
	ConfigOption("GladiusBase", base)  // Convenient option to have, not needed though
	ConfigOption("UPNPEnabled", false) // Use UPNP to get external IP and open ports

}

func ConfigOption(key string, defaultValue interface{}) string {
	viper.SetDefault(key, defaultValue)

	return key
}
