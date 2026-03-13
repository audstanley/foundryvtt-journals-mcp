package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	WorldsPath string `mapstructure:"worlds_path"`
	User       string `mapstructure:"user"`
	LogLevel   string `mapstructure:"log_level"`
}

// Load reads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath("./")
		viper.AddConfigPath("$HOME/.fjm")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("FJM")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Set defaults
	viper.SetDefault("log_level", "info")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults
		} else {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	c := &Config{}
	if err := viper.Unmarshal(c); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate required fields
	if c.WorldsPath == "" {
		return nil, fmt.Errorf("worlds_path is required")
	}

	return c, nil
}

// GetWorldsPath returns the worlds directory path
func GetWorldsPath() (string, error) {
	if path := os.Getenv("FJM_WORLDS_PATH"); path != "" {
		return path, nil
	}

	// Try to load from config
	c, err := Load("")
	if err != nil {
		return "", err
	}
	return c.WorldsPath, nil
}

// GetUser returns the username if configured
func GetUser() string {
	if user := os.Getenv("FJM_USER"); user != "" {
		return user
	}

	c, err := Load("")
	if err != nil {
		return ""
	}
	return c.User
}
