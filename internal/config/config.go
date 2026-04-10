package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppConfig stores application configuration.
type AppConfig struct {
	BackupDir         string         `json:"backupDir"`
	ScheduleEnabled   bool           `json:"scheduleEnabled"`
	ScheduleLevels    map[string]int `json:"scheduleLevels"`
	DefaultCategories []string       `json:"defaultCategories"`
	AutoDeleteDays    int            `json:"autoDeleteDays"`
}

var configPath string

func init() {
	home, _ := os.UserHomeDir()
	configPath = filepath.Join(home, ".config", "gnome-cfg-exporter", "config.json")
}

// DefaultConfig returns the default configuration.
func DefaultConfig() AppConfig {
	home, _ := os.UserHomeDir()
	return AppConfig{
		BackupDir:       filepath.Join(home, ".local", "share", "gnome-cfg-exporter"),
		ScheduleEnabled: false,
		ScheduleLevels: map[string]int{
			"hourly": 0, "daily": 5, "weekly": 3, "monthly": 2,
		},
		DefaultCategories: []string{
			"metadata", "dconf", "extensions", "themes", "icons", "fonts",
			"wallpapers", "gtk", "autostart", "terminal", "monitors",
			"nautilus", "mimeapps", "goa", "keyrings", "desktop_files", "xdg_dirs",
		},
		AutoDeleteDays: 0,
	}
}

// Load reads the config from disk or returns defaults.
func Load() (AppConfig, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}
	return cfg, nil
}

// Save writes the config to disk.
func Save(cfg AppConfig) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// Exists returns true if a config file exists on disk.
func Exists() bool {
	_, err := os.Stat(configPath)
	return err == nil
}
