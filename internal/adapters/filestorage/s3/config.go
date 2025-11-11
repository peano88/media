package s3

import (
	"os"

	"github.com/peano88/medias/config"
)

// Config holds S3 configuration
type Config struct {
	Region         string `mapstructure:"region"`
	Endpoint       string `mapstructure:"endpoint"`
	PublicEndpoint string `mapstructure:"public-endpoint"`
	BucketName     string `mapstructure:"bucket-name"`
	UploadExpiry   int    `mapstructure:"upload-expiry"` // in seconds
}

func SetDefaultConfig(loader config.ConfigLoader, prefix string) {
	loader.SetDefault(prefix+".region", "eu-central-1")
	loader.SetDefault(prefix+".upload-expiry", 3600)
}

// Credentials returns AWS credentials from environment variables
// Uses standard AWS environment variable names
func (c *Config) Credentials() (accessKeyID, secretAccessKey string) {
	accessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	return
}
