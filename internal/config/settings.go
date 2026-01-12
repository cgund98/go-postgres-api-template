package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/cgund98/go-postgres-api-template/internal/observability"
)

var logger = observability.Logger

// Config holds the application configuration
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`

	AWS AWSConfig `mapstructure:"aws"`

	Events EventsConfig `mapstructure:"events"`

	Server ServerConfig `mapstructure:"server"`

	Environment string `mapstructure:"environment"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type AWSConfig struct {
	Region        string `mapstructure:"region"`
	UseLocalstack bool   `mapstructure:"use_localstack"`
	Endpoint      string `mapstructure:"endpoint"`
}

type EventsConfig struct {
	TopicARN            string `mapstructure:"events_topic_arn"`
	QueueURLUserCreated string `mapstructure:"queue_url_user_created"`
	QueueURLUserUpdated string `mapstructure:"queue_url_user_updated"`
	QueueURLUserDeleted string `mapstructure:"queue_url_user_deleted"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

// LoadConfig loads configuration from file, environment variables, or defaults
func LoadConfig() (*Config, error) {
	// Enable environment variable support
	// No prefix - use standard env var names (DATABASE_URL, AWS_REGION, etc.)
	// Replace dots with underscores for nested keys (e.g., database.url -> DATABASE_URL)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set defaults (this makes AutomaticEnv work for those keys)
	setDefaults()

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

	// Note: Environment variables have the highest priority and will override
	// both .env.local and .env files
	// Explicitly bind environment variables AFTER loading .env files to ensure they're read correctly
	// BindEnv maps viper keys to environment variable names
	if err := viper.BindEnv("events.events_topic_arn", "EVENTS_TOPIC_ARN"); err != nil {
		return nil, fmt.Errorf("error binding env var EVENTS_TOPIC_ARN: %w", err)
	}
	if err := viper.BindEnv("events.queue_url_user_created", "EVENTS_QUEUE_URL_USER_CREATED"); err != nil {
		return nil, fmt.Errorf("error binding env var EVENTS_QUEUE_URL_USER_CREATED: %w", err)
	}
	if err := viper.BindEnv("events.queue_url_user_updated", "EVENTS_QUEUE_URL_USER_UPDATED"); err != nil {
		return nil, fmt.Errorf("error binding env var EVENTS_QUEUE_URL_USER_UPDATED: %w", err)
	}
	if err := viper.BindEnv("events.queue_url_user_deleted", "EVENTS_QUEUE_URL_USER_DELETED"); err != nil {
		return nil, fmt.Errorf("error binding env var EVENTS_QUEUE_URL_USER_DELETED: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Database defaults
	viper.SetDefault("database.url", "")

	// AWS defaults
	viper.SetDefault("aws.region", "us-east-1")
	viper.SetDefault("aws.use_localstack", false)
	viper.SetDefault("aws.endpoint", "")

	// Events defaults
	viper.SetDefault("events.events_topic_arn", "")
	viper.SetDefault("events.queue_url_user_created", "")
	viper.SetDefault("events.queue_url_user_updated", "")
	viper.SetDefault("events.queue_url_user_deleted", "")

	// Server defaults
	viper.SetDefault("server.port", "8080")

	// Environment defaults
	viper.SetDefault("environment", "development")
}

// LoadSettings is an alias for LoadConfig to maintain backward compatibility
// Deprecated: Use LoadConfig instead
func LoadSettings() (*Config, error) {
	return LoadConfig()
}
