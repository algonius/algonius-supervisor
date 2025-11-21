package config

import (
	"github.com/spf13/viper"
)

// ConfigFactory provides factory methods for creating configuration managers
type ConfigFactory struct{}

// NewConfigFactory creates a new configuration factory
func NewConfigFactory() *ConfigFactory {
	return &ConfigFactory{}
}

// CreateConfigManager creates a new configuration manager with default viper
func (f *ConfigFactory) CreateConfigManager() IConfigManager {
	v := viper.New()
	return NewConfigManager(v)
}

// CreateConfigManagerWithViper creates a new configuration manager with custom viper
func (f *ConfigFactory) CreateConfigManagerWithViper(v *viper.Viper) IConfigManager {
	return NewConfigManager(v)
}

// CreateTestConfigManager creates a configuration manager for testing
func (f *ConfigFactory) CreateTestConfigManager() IConfigManager {
	v := viper.New()
	v.Set("server.url", "http://localhost:8080")
	v.Set("server.timeout", "30s")
	v.Set("display.format", "table")
	v.Set("display.colors", false) // Disable colors in tests
	v.Set("defaults.restart_attempts", 3)
	v.Set("defaults.wait_time", "5s")

	return NewConfigManager(v)
}