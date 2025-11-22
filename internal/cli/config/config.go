package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// ConfigManager implements the IConfigManager interface
type ConfigManager struct {
	config     *CLIConfig
	viper      *viper.Viper
	configPath string
}

// NewConfigManager creates a new configuration manager with injected viper instance
func NewConfigManager(v *viper.Viper) *ConfigManager {
	return &ConfigManager{
		config: &CLIConfig{},
		viper:  v,
	}
}

// Load loads configuration from file, environment variables, and defaults
func (cm *ConfigManager) Load() (*CLIConfig, error) {
	// Initialize config if nil
	if cm.config == nil {
		cm.config = &CLIConfig{}
	}

	// Set configuration file name and locations
	cm.viper.SetConfigName("supervisorctl")
	cm.viper.AddConfigPath("$HOME/.config/supervisorctl")
	cm.viper.AddConfigPath(".")
	cm.viper.AddConfigPath("/etc/supervisorctl")

	// Set environment variable prefix
	cm.viper.SetEnvPrefix("SUPERVISOR")
	cm.viper.AutomaticEnv()

	// Set default values
	cm.setDefaults()

	// Try to read configuration file if it exists
	if err := cm.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file found but has errors - ignore and use defaults
			// This prevents the CLI from failing entirely due to config file issues
		} else {
			// Config file not found - that's OK, use defaults
		}
	} else {
		cm.configPath = cm.viper.ConfigFileUsed()
		// Unmarshal configuration if file was found
		if err := cm.viper.Unmarshal(cm.config); err != nil {
			// Config file has parsing errors - ignore and use defaults
			// This prevents the CLI from failing entirely due to malformed config
		}
	}

	return cm.config, nil
}

// Save saves the current configuration to file
func (cm *ConfigManager) Save() error {
	if cm.configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		cm.configPath = filepath.Join(home, ".config", "supervisorctl", "supervisorctl.yaml")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	cm.viper.Set("server", cm.config.Server)
	cm.viper.Set("auth", cm.config.Auth)
	cm.viper.Set("display", cm.config.Display)
	cm.viper.Set("defaults", cm.config.Defaults)

	return cm.viper.WriteConfigAs(cm.configPath)
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *CLIConfig {
	return cm.config
}

// SetConfig updates the configuration
func (cm *ConfigManager) SetConfig(config *CLIConfig) {
	cm.config = config
}

// GetServerURL returns the configured server URL
func (cm *ConfigManager) GetServerURL() string {
	return cm.config.Server.URL
}

// GetServerTimeout returns the configured server timeout
func (cm *ConfigManager) GetServerTimeout() time.Duration {
	return cm.config.Server.Timeout
}

// GetAuthToken returns the authentication token
func (cm *ConfigManager) GetAuthToken() string {
	return cm.config.Auth.Token
}

// SetServerURL updates the server URL
func (cm *ConfigManager) SetServerURL(url string) {
	cm.config.Server.URL = url
}

// SetAuthToken updates the authentication token
func (cm *ConfigManager) SetAuthToken(token string) {
	cm.config.Auth.Token = token
}

// Validate validates the current configuration
func (cm *ConfigManager) Validate() error {
	if cm.config.Server.URL == "" {
		return fmt.Errorf("server URL is required")
	}

	if cm.config.Server.Timeout <= 0 {
		return fmt.Errorf("server timeout must be positive")
	}

	if cm.config.Display.Format != "" {
		validFormats := []string{"table", "json", "yaml"}
		valid := false
		for _, format := range validFormats {
			if cm.config.Display.Format == format {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid display format: %s", cm.config.Display.Format)
		}
	}

	return nil
}

// setDefaults sets default configuration values
func (cm *ConfigManager) setDefaults() {
	cm.viper.SetDefault("server.url", "http://localhost:8080")
	cm.viper.SetDefault("server.timeout", "30s")
	cm.viper.SetDefault("display.format", "table")
	cm.viper.SetDefault("display.colors", true)
	cm.viper.SetDefault("display.refresh_rate", "5s")
	cm.viper.SetDefault("defaults.restart_attempts", 3)
	cm.viper.SetDefault("defaults.wait_time", "5s")
}

// GetViper returns the viper instance (useful for testing)
func (cm *ConfigManager) GetViper() *viper.Viper {
	return cm.viper
}

// ConfigManagerKey is the key used to store the configuration manager in context
type ConfigManagerKey struct{}