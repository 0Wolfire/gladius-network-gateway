package config

import (
	"path/filepath"
	"strings"

	"github.com/gladiusio/gladius-common/pkg/utils"
	"github.com/spf13/viper"
)

// SetupConfig sets up viper and adds our config options
func SetupConfig() (string, error) {
	base, err := utils.GetGladiusBase()
	if err != nil {
		return "Error retrieving base directory", err
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
	var message = "Using provided config file and overriding default"
	if err != nil {
		message = "Error reading config file, it may not exist or is corrupted. Using defaults."
	}

	// Build our config options
	buildOptions(base)

	return message, err
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
	ConfigOption("Blockchain.Provider", "https://mainnet.infura.io/v3/1d3545f907ff4598893997c522e46676")
	ConfigOption("Blockchain.MarketAddress", "0x27a9390283236f836a0b3c8dfdbed2ed854322fc")
	ConfigOption("Blockchain.PoolUrl", "http://174.138.111.1/api/")
	ConfigOption("Blockchain.PoolManagerAddress", "0x9717EaDbfE344457135a4f1fA8AE3B11B4CAB0b7")

	// Wallet options
	ConfigOption("Wallet.Directory", filepath.Join(base, "wallet"))
	ConfigOption("Wallet.Passphrase", "") // Only should be used for automated deployment

	// Pool
	ConfigOption("Pool.AutoJoin", false)
	ConfigOption("Pool.URL", "")
	ConfigOption("Pool.Address", "")

	// API options
	ConfigOption("API.Port", "3001")
	ConfigOption("API.DebugRequests", false)
	ConfigOption("API.RemoteConnectionsAllowed", false)

	// Misc.
	ConfigOption("GladiusBase", base)   // Convenient option to have, not needed though
	ConfigOption("UPNPEnabled", false)  // Use UPNP to get external IP and open ports
	ConfigOption("HTTPProfiler", false) // Enable the profiler on an HTTP endpoint
}

func ConfigOption(key string, defaultValue interface{}) string {
	viper.SetDefault(key, defaultValue)

	return key
}
