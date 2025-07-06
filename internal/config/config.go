package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DatabasePath       string `json:"database_path,omitempty"`
	DefaultView        string `json:"default_view,omitempty"`
}

func GetDefaultConfig() *Config {
	return &Config{
		DatabasePath:       "", // Empty means use default path
		DefaultView:        "week", // Default to week view
	}
}

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return GetDefaultConfig(), err
	}

	configDir := filepath.Join(homeDir, ".config", "chronos")
	configPath := filepath.Join(configDir, "config.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := GetDefaultConfig()
		err := SaveConfig(config)
		if err != nil {
			return config, err
		}
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return GetDefaultConfig(), err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return GetDefaultConfig(), err
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "chronos")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetDatabasePath returns the database path from config or default path
func GetDatabasePath(config *Config) string {
	if config.DatabasePath != "" {
		return config.DatabasePath
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "chronos.db" // fallback to current directory
	}
	
	return filepath.Join(homeDir, ".local", "share", "chronos", "data.db")
}

// GetDefaultView returns the default view mode from config, defaulting to "week" if not set
func GetDefaultView(config *Config) string {
	if config.DefaultView != "" {
		switch config.DefaultView {
		case "week", "month", "agenda":
			return config.DefaultView
		default:
			return "week" // fallback to week if invalid value
		}
	}
	return "week"
}