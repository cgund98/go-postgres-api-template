package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

var logger = observability.Logger

// Config holds the application configuration
type Config struct {
	DatabaseURL string `split_words:"true"`

	AwsRegion        string `default:"us-east-1"`
	AwsUseLocalstack bool   `split_words:"true"`
	AwsEndpoint      string `split_words:"true"`

	EventsTopicArn            string `split_words:"true"`
	EventsQueueURLUserCreated string `split_words:"true"`
	EventsQueueURLUserUpdated string `split_words:"true"`
	EventsQueueURLUserDeleted string `split_words:"true"`

	ServerPort string `split_words:"true" default:"8080"`

	Environment string
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

// LoadConfig loads configuration from file, environment variables, or defaults
func LoadConfig() (*Config, error) {

	// Load .env.local file first (if it exists)
	// Lowest priority - will be overridden by .env and environment variables
	// Use godotenv to load as environment variables so standard naming (DATABASE_URL) works
	if _, err := os.Stat(".env.local"); err == nil {
		if err := godotenv.Load(".env.local"); err != nil {
			return nil, fmt.Errorf("error loading .env.local file: %w", err)
		}
		logger.Info("loaded .env.local file")
	}

	// Load .env file (if it exists)
	// Higher priority - will override .env.local but be overridden by environment variables
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
		logger.Info("loaded .env file")
	}

	logger.Info("loading config", "env", os.Getenv("ENVIRONMENT"))

	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}
