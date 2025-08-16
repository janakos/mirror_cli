package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration
type Config struct {
	PeerDBHost string `yaml:"peerdb_host" mapstructure:"peerdb_host"`
	PeerDBPort int    `yaml:"peerdb_port" mapstructure:"peerdb_port"`
	TLS        bool   `yaml:"tls" mapstructure:"tls"`
	Username   string `yaml:"username" mapstructure:"username"`
	Password   string `yaml:"password" mapstructure:"password"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		PeerDBHost: "localhost",
		PeerDBPort: 8112,
		TLS:        false,
		Username:   "",
		Password:   "",
	}
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Set up viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add config search paths
	homeDir, err := os.UserHomeDir()
	if err == nil {
		viper.AddConfigPath(filepath.Join(homeDir, ".mirror_cli"))
	}
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/mirror_cli")

	// Environment variable support
	viper.SetEnvPrefix("MIRROR_CLI")
	viper.AutomaticEnv()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal into struct
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".mirror_cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Address returns the full address for gRPC connection
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.PeerDBHost, c.PeerDBPort)
}
