package config

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigManager_Load(t *testing.T) {
	tests := []struct {
		name     string
		viper    *viper.Viper
		wantURL  string
		wantTimeout time.Duration
		wantErr  bool
	}{
		{
			name: "default configuration",
			viper: func() *viper.Viper {
				v := viper.New()
				v.Set("server.url", "http://localhost:8080")
				v.Set("server.timeout", "30s")
				return v
			}(),
			wantURL:     "http://localhost:8080",
			wantTimeout: 30 * time.Second,
			wantErr:     false,
		},
		{
			name: "custom configuration",
			viper: func() *viper.Viper {
				v := viper.New()
				v.Set("server.url", "http://example.com:9000")
				v.Set("server.timeout", "60s")
				v.Set("display.format", "json")
				return v
			}(),
			wantURL:     "http://example.com:9000",
			wantTimeout: 60 * time.Second,
			wantErr:     false,
		},
		{
			name: "missing required fields uses defaults",
			viper: func() *viper.Viper {
				v := viper.New()
				// server.url is missing - will use default from setDefaults()
				v.Set("server.timeout", "30s")
				return v
			}(),
			wantURL:     "http://localhost:8080", // Default value
			wantTimeout: 30 * time.Second,
			wantErr:     false, // Load succeeds, validation would fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewConfigManager(tt.viper)

			config, err := cm.Load()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, cm.GetServerURL())
			assert.Equal(t, tt.wantTimeout, cm.GetServerTimeout())
			assert.NotNil(t, config)
		})
	}
}

func TestConfigManager_Validate(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(*CLIConfig)
		wantErr     bool
		errMsg      string
	}{
		{
			name: "valid configuration",
			setupConfig: func(c *CLIConfig) {
				c.Server.URL = "http://localhost:8080"
				c.Server.Timeout = 30 * time.Second
				c.Display.Format = "table"
			},
			wantErr: false,
		},
		{
			name: "missing server URL",
			setupConfig: func(c *CLIConfig) {
				c.Server.URL = ""
				c.Server.Timeout = 30 * time.Second
			},
			wantErr: true,
			errMsg:  "server URL is required",
		},
		{
			name: "invalid timeout",
			setupConfig: func(c *CLIConfig) {
				c.Server.URL = "http://localhost:8080"
				c.Server.Timeout = 0
			},
			wantErr: true,
			errMsg:  "server timeout must be positive",
		},
		{
			name: "invalid display format",
			setupConfig: func(c *CLIConfig) {
				c.Server.URL = "http://localhost:8080"
				c.Server.Timeout = 30 * time.Second
				c.Display.Format = "invalid"
			},
			wantErr: true,
			errMsg:  "invalid display format: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewConfigManager(viper.New())
			tt.setupConfig(cm.config)

			err := cm.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigFactory_CreateTestConfigManager(t *testing.T) {
	factory := NewConfigFactory()

	cm := factory.CreateTestConfigManager()

	// Load the test configuration
	_, err := cm.Load()
	require.NoError(t, err)

	// Verify test defaults
	assert.Equal(t, "http://localhost:8080", cm.GetServerURL())
	assert.Equal(t, 30*time.Second, cm.GetServerTimeout())

	// Verify the viper instance can be retrieved for testing
	v := cm.(*ConfigManager).GetViper()
	assert.NotNil(t, v)
	assert.Equal(t, "table", v.GetString("display.format"))
	assert.False(t, v.GetBool("display.colors")) // Colors disabled in tests
}