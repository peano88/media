package main

import (
	"github.com/peano88/medias/config"
	"github.com/peano88/medias/internal/adapters/storage/postgres"
)

// application config holds all application configuration
type applicationConfig struct {
	*config.Config
	Server   ServerConfig    `mapstructure:"server"`
	Database postgres.Config `mapstructure:"database"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port                   int `mapstructure:"listen-port"`
	ShutdownTimeoutSeconds int `mapstructure:"shutdown-timeout-seconds"`
}

func LoadConfig() (*applicationConfig, error) {
	baseConfig := config.NewConfig()
	cfgLoader := baseConfig.ConfigLoader()

	cfgLoader.SetDefault("server.shutdown-timeout-seconds", 20)
	cfgLoader.SetDefault("server.listen-port", 8080)
	postgres.SetDefaultConfig(cfgLoader, "database")

	if err := baseConfig.Load(); err != nil {
		return nil, err
	}

	ac := &applicationConfig{
		Config: baseConfig,
	}

	if err := cfgLoader.Unmarshal(&ac); err != nil {
		return nil, err
	}

	return ac, nil
}
