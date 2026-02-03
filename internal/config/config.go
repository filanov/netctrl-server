package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Logging  LoggingConfig  `yaml:"logging"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Gateway  GatewayConfig  `yaml:"gateway"`
}

// ServerConfig contains general server configuration
type ServerConfig struct {
	Environment string `yaml:"environment"`
}

// GRPCConfig contains gRPC server configuration
type GRPCConfig struct {
	EnableReflection bool `yaml:"enable_reflection"`
	Port             int  `yaml:"port"`
}

// GatewayConfig contains HTTP gateway configuration
type GatewayConfig struct {
	EnableCORS bool `yaml:"enable_cors"`
	Port       int  `yaml:"port"`
}

// DatabaseConfig contains PostgreSQL database configuration
type DatabaseConfig struct {
	URL            string `yaml:"url"`
	MinConnections int32  `yaml:"min_connections"`
	MaxConnections int32  `yaml:"max_connections"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Load reads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	applyDefaults(config)

	return config, nil
}

// LoadOrDefault attempts to load configuration from a file, or returns defaults if the file doesn't exist
func LoadOrDefault(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		config := &Config{}
		applyDefaults(config)
		return config, nil
	}

	return Load(path)
}

// applyDefaults sets default values for missing configuration
func applyDefaults(config *Config) {
	if config.Server.Environment == "" {
		config.Server.Environment = "development"
	}

	if config.GRPC.Port == 0 {
		config.GRPC.Port = 9090
	}
	if !config.GRPC.EnableReflection {
		config.GRPC.EnableReflection = true
	}

	if config.Gateway.Port == 0 {
		config.Gateway.Port = 8080
	}
	if !config.Gateway.EnableCORS {
		config.Gateway.EnableCORS = true
	}

	// Database configuration with environment variable override
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.Database.URL = dbURL
	}
	if config.Database.MaxConnections == 0 {
		config.Database.MaxConnections = 100
	}
	if config.Database.MinConnections == 0 {
		config.Database.MinConnections = 20
	}

	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}
}
