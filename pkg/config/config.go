package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	// Editor settings
	FontSize    int    `json:"fontSize"`
	Theme       string `json:"theme"`
	TabSize     int    `json:"tabSize"`
	WordWrap    bool   `json:"wordWrap"`
	LineNumbers bool   `json:"lineNumbers"`

	// Window settings
	WindowWidth  int `json:"windowWidth"`
	WindowHeight int `json:"windowHeight"`

	// PowerShell settings
	ExecutionTimeout int    `json:"executionTimeout"` // in seconds
	PowerShellPath   string `json:"powerShellPath"`

	// Recent files
	RecentFiles []string `json:"recentFiles"`
}

// Default returns default configuration
func Default() *Config {
	return &Config{
		FontSize:         12,
		Theme:            "monokai",
		TabSize:          4,
		WordWrap:         false,
		LineNumbers:      true,
		WindowWidth:      900,
		WindowHeight:     700,
		ExecutionTimeout: 30,
		PowerShellPath:   "pwsh",
		RecentFiles:      []string{},
	}
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return Default(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// AddRecentFile adds a file to recent files list
func (c *Config) AddRecentFile(path string) {
	// Remove if already exists
	for i, f := range c.RecentFiles {
		if f == path {
			c.RecentFiles = append(c.RecentFiles[:i], c.RecentFiles[i+1:]...)
			break
		}
	}

	// Add to beginning
	c.RecentFiles = append([]string{path}, c.RecentFiles...)

	// Keep only last 10
	if len(c.RecentFiles) > 10 {
		c.RecentFiles = c.RecentFiles[:10]
	}
}

// GetConfigPath returns the default config file path
func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".ps-ide-config.json"
	}
	return filepath.Join(home, ".config", "ps-ide", "config.json")
}
