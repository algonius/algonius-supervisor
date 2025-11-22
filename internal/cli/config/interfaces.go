package config

import (
	"time"
)

// IConfigManager defines the interface for CLI configuration management
type IConfigManager interface {
	// Load loads configuration from file, environment variables, and defaults
	Load() (*CLIConfig, error)

	// Save saves the current configuration to file
	Save() error

	// GetServerURL returns the configured server URL
	GetServerURL() string

	// GetServerTimeout returns the configured server timeout
	GetServerTimeout() time.Duration

	// GetAuthToken returns the authentication token
	GetAuthToken() string

	// SetServerURL updates the server URL
	SetServerURL(url string)

	// SetAuthToken updates the authentication token
	SetAuthToken(token string)

	// Validate validates the current configuration
	Validate() error
}

// CLIConfig represents the configuration for the supervisorctl client
type CLIConfig struct {
	Server   ServerConfig   `yaml:"server" validate:"required"`
	Auth     AuthConfig     `yaml:"auth"`
	Display  DisplayConfig  `yaml:"display"`
	Defaults DefaultsConfig `yaml:"defaults"`
}

// ServerConfig contains server connection configuration
type ServerConfig struct {
	URL     string        `yaml:"url" validate:"required,url"`
	Timeout time.Duration `yaml:"timeout" validate:"min=1s,max=5m"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	Token string `yaml:"token"`
}

// DisplayConfig contains display and formatting configuration
type DisplayConfig struct {
	Format      string `yaml:"format"` // table, json, yaml
	Colors      bool   `yaml:"colors"`
	RefreshRate string `yaml:"refresh_rate"`
}

// DefaultsConfig contains default behavior configuration
type DefaultsConfig struct {
	RestartAttempts int           `yaml:"restart_attempts" validate:"min=0,max=10"`
	WaitTime       time.Duration `yaml:"wait_time" validate:"min=1s,max=5m"`
}