package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config is a base configuration allowing dynamic loading and alteration
// based on environment variable and config files
type Config struct {
	viper *viper.Viper
	env   string
}

type DecoderConfigOption = viper.DecoderConfigOption

// ConfigLoader defines methods for dynamic alteration of configuration
// (currently only used to set default values)
// as well as unmarshalling and mapping configuration to structs
type ConfigLoader interface {
	SetDefault(string, any)
	Unmarshal(any, ...DecoderConfigOption) error
	AllSettings() map[string]any
}

// ConfigLoader returns the underlying ConfigLoader
func (c *Config) ConfigLoader() ConfigLoader {
	return c.viper
}

func NewConfig() *Config {
	return &Config{
		viper: viper.New(),
		env:   getEnv(),
	}

}

// Load reads configuration from file based on ENV environment variable
// Defaults to "dev" environment if ENV is not set
// Config files should be named: config.{env}.yaml
func (c *Config) Load() error {

	// Set config file name and path
	c.viper.SetConfigName(fmt.Sprintf("config.%s", c.env))
	c.viper.SetConfigType("yaml")

	// Check if CONFIG_PATH environment variable is set
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		c.viper.AddConfigPath(configPath)
	} else {
		// Default search paths
		c.viper.AddConfigPath(".")
		c.viper.AddConfigPath("./config")
	}

	// Enable environment variable overrides
	// e.g., SERVER_LISTEN_PORT will override server.listen-port
	c.viper.SetEnvPrefix("") // No prefix
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	c.viper.AutomaticEnv() // TODO

	// Read config file
	if err := c.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return nil

}

// getEnv returns the current environment from ENV variable
// Defaults to "dev" if not set
func getEnv() string {
	env := os.Getenv("ENV")
	if env == "" {
		return "dev"
	}
	return env
}

// Env returns the current environment
func (c *Config) Env() string {
	return c.env
}
