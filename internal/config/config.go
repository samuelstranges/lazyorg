package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DatabasePath            string `json:"database_path,omitempty"`
	DefaultView             string `json:"default_view,omitempty"`
	WeatherLocation         string `json:"weather_location,omitempty"`
	WeatherUnit             string `json:"weather_unit,omitempty"`
	NotificationsEnabled    bool   `json:"notifications_enabled,omitempty"`
	NotificationMinutes     int    `json:"notification_minutes,omitempty"`
	DefaultColor            string `json:"default_color,omitempty"`
	DefaultEventLength      float64 `json:"default_event_length,omitempty"`
}

func GetDefaultConfig() *Config {
	return &Config{
		DatabasePath:            "", // Empty means use default path
		DefaultView:             "week", // Default to week view
		WeatherLocation:         "", // Empty means no weather display
		WeatherUnit:             "celsius", // Default to celsius
		NotificationsEnabled:    false, // Default to disabled
		NotificationMinutes:     15, // Default to 15 minutes before
		DefaultColor:            "", // Empty means auto-generate from event name
		DefaultEventLength:      1.0, // Default to 1 hour
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

// IsWeatherEnabled returns true if weather location is configured
func IsWeatherEnabled(config *Config) bool {
	return config.WeatherLocation != ""
}

// GetWeatherLocation returns the configured weather location
func GetWeatherLocation(config *Config) string {
	return config.WeatherLocation
}

// GetWeatherUnit returns the configured temperature unit (celsius or fahrenheit)
func GetWeatherUnit(config *Config) string {
	if config.WeatherUnit == "fahrenheit" || config.WeatherUnit == "f" {
		return "fahrenheit"
	}
	return "celsius" // Default to celsius
}

// IsNotificationsEnabled returns true if notifications are enabled
func IsNotificationsEnabled(config *Config) bool {
	return config.NotificationsEnabled
}

// GetNotificationMinutes returns the configured notification minutes (0-60)
func GetNotificationMinutes(config *Config) int {
	minutes := config.NotificationMinutes
	if minutes < 0 || minutes > 60 {
		return 15 // Default to 15 minutes if invalid
	}
	return minutes
}

// GetDefaultColor returns the configured default color or empty string for auto-generation
func GetDefaultColor(config *Config) string {
	if config.DefaultColor == "" {
		return "" // Empty means auto-generate from event name
	}
	
	// Validate color name
	switch config.DefaultColor {
	case "Red", "Green", "Yellow", "Blue", "Magenta", "Cyan", "White":
		return config.DefaultColor
	default:
		return "" // Invalid color defaults to auto-generation
	}
}

// GetDefaultEventLength returns the configured default event length in hours
func GetDefaultEventLength(config *Config) float64 {
	length := config.DefaultEventLength
	// Validate that length is a whole number or ends in .5 and is 0.5 < x < 24
	    if length < 0.5 || length > 24 || int(length*10)%5 != 0 {
		return 1.0 // Default to 1 hour if invalid
	}
	return length
}