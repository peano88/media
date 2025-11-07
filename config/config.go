package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	viper  *viper.Viper
	env    string
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port                   int `mapstructure:"listen-port"`
	ShutdownTimeoutSeconds int `mapstructure:"shutdown-timeout-seconds"`
}

// Load reads configuration from file based on ENV environment variable
// Defaults to "dev" environment if ENV is not set
// Config files should be named: config.{env}.yaml
func Load() (*Config, error) {
	env := getEnv()

	v := viper.New()

	// Set config file name and path
	v.SetConfigName(fmt.Sprintf("config.%s", env))
	v.SetConfigType("yaml")

	// Check if CONFIG_PATH environment variable is set
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		v.AddConfigPath(configPath)
	} else {
		// Default search paths
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// Enable environment variable overrides
	// e.g., SERVER_LISTEN_PORT will override server.listen-port
	v.SetEnvPrefix("") // No prefix
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv() // TODO

	//TODO: move to default function
	v.SetDefault("server.shutdown-timeout-seconds", 20)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.viper = v
	cfg.env = env

	return &cfg, nil
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

// AllSettings returns all configuration settings as a map
func (c *Config) AllSettings() map[string]any {
	if c.viper == nil {
		return nil
	}
	return c.viper.AllSettings()
}

// Env returns the current environment
func (c *Config) Env() string {
	return c.env
}
