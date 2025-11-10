package postgres

import (
	"fmt"
	"os"

	"github.com/peano88/medias/config"
)

type Config struct {
	Host               string `mapstructure:"host"`
	Port               int    `mapstructure:"port"`
	Database           string `mapstructure:"database"`
	User               string `mapstructure:"user"`
	SSLMode            string `mapstructure:"ssl-mode"`
	MaxConns           int32  `mapstructure:"max-conns"`
	MinConns           int32  `mapstructure:"min-conns"`
	MaxConnLifetimeMin int    `mapstructure:"max-conn-lifetime-min"`
	MaxConnIdleTimeMin int    `mapstructure:"max-conn-idle-time-min"`
}

func SetDefaultConfig(loader config.ConfigLoader, prefix string) {
	loader.SetDefault(prefix+".max-conns", 25)
	loader.SetDefault(prefix+".min-conns", 5)
	loader.SetDefault(prefix+".max-conn-lifetime-min", 60)
	loader.SetDefault(prefix+".max-conn-idle-time-min", 30)
	loader.SetDefault(prefix+".ssl-mode", "require")
}

// ConnectionString builds the database connection string
// Password must be provided via DB_PASSWORD environment variable
func (d *Config) connectionString() string {
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "" // Allow empty for development, but should be set in production
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User,
		password,
		d.Host,
		d.Port,
		d.Database,
		d.SSLMode,
	)
}
