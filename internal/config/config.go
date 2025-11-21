package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
	LogLevel   string `mapstructure:"log_level"`
	
	// A2A Configuration
	A2A struct {
		Enabled     bool          `mapstructure:"enabled"`
		Timeout     time.Duration `mapstructure:"timeout"`
		AuthEnabled bool          `mapstructure:"auth_enabled"`
		AuthToken   string        `mapstructure:"auth_token"`
	} `mapstructure:"a2a"`
	
	// Agent Configuration
	Agents []AgentConfig `mapstructure:"agents"`
	
	// Scheduler Configuration
	Scheduler struct {
		Enabled bool `mapstructure:"enabled"`
	} `mapstructure:"scheduler"`
}

// AgentConfig defines the configuration for an individual agent
type AgentConfig struct {
	ID                  string            `mapstructure:"id"`
	Name                string            `mapstructure:"name"`
	AgentType           string            `mapstructure:"agent_type"`
	ExecutablePath      string            `mapstructure:"executable_path"`
	WorkingDirectory    string            `mapstructure:"working_directory"`
	Envs                map[string]string `mapstructure:"envs"`
	CliArgs             map[string]string `mapstructure:"cli_args"`
	Mode                string            `mapstructure:"mode"` // "task" or "interactive"
	InputPattern        string            `mapstructure:"input_pattern"`
	OutputPattern       string            `mapstructure:"output_pattern"`
	InputFileTemplate   string            `mapstructure:"input_file_template"`
	OutputFileTemplate  string            `mapstructure:"output_file_template"`
	AccessType          string            `mapstructure:"access_type"` // "read-only" or "read-write"
	MaxConcurrentExecutions int           `mapstructure:"max_concurrent_executions"`
	Timeout             int               `mapstructure:"timeout"`
	SessionTimeout      int               `mapstructure:"session_timeout"`
	KeepAlive           bool              `mapstructure:"keep_alive"`
	Enabled             bool              `mapstructure:"enabled"`
}

// LoadConfig loads the application configuration using viper
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/github.com/algonius/algonius-supervisor/")
	
	// Set default values
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 8080)
	viper.SetDefault("environment", "development")
	viper.SetDefault("log_level", "info")
	
	viper.SetDefault("a2a.enabled", true)
	viper.SetDefault("a2a.timeout", "30s")
	viper.SetDefault("a2a.auth_enabled", true)
	
	viper.SetDefault("scheduler.enabled", true)

	// Allow environment variables to override config
	viper.AutomaticEnv()
	
	err := viper.ReadInConfig()
	if err != nil {
		// If config file not found, proceed with defaults and environment variables
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	// Validate and set defaults for agents
	for i := range config.Agents {
		if config.Agents[i].MaxConcurrentExecutions == 0 {
			if config.Agents[i].AccessType == "read-write" {
				config.Agents[i].MaxConcurrentExecutions = 1
			} else {
				config.Agents[i].MaxConcurrentExecutions = 10 // default for read-only
			}
		}

		// Validate access type
		if config.Agents[i].AccessType != "read-only" && config.Agents[i].AccessType != "read-write" {
			config.Agents[i].AccessType = "read-only" // Default to read-only
		}

		// Validate mode
		if config.Agents[i].Mode != "task" && config.Agents[i].Mode != "interactive" {
			config.Agents[i].Mode = "task" // Default to task mode
		}

		// Validate input/output patterns
		if config.Agents[i].InputPattern == "" {
			config.Agents[i].InputPattern = "stdin" // Default input pattern
		}

		if config.Agents[i].OutputPattern == "" {
			config.Agents[i].OutputPattern = "stdout" // Default output pattern
		}

		// Validate timeouts
		if config.Agents[i].Timeout <= 0 {
			config.Agents[i].Timeout = 300 // Default to 5 minutes
		}
	}

	return &config, validateConfig(&config)
}

// validateConfig validates the configuration values
func validateConfig(config *Config) error {
	// Validate port range
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", config.Port)
	}

	// Validate host is not empty
	if config.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// Validate log level
	switch config.LogLevel {
	case "debug", "info", "warn", "error":
		// Valid log levels
	default:
		return fmt.Errorf("log level must be one of: debug, info, warn, error, got %s", config.LogLevel)
	}

	// Validate agent configurations
	agentIds := make(map[string]bool)
	for _, agent := range config.Agents {
		// Validate ID is unique and not empty
		if agent.ID == "" {
			return fmt.Errorf("agent ID cannot be empty")
		}
		if agentIds[agent.ID] {
			return fmt.Errorf("duplicate agent ID found: %s", agent.ID)
		}
		agentIds[agent.ID] = true

		// Validate executable path exists if specified
		if agent.ExecutablePath != "" {
			// We won't check if the file exists here as it might not be available during config loading
		}

		// Validate access type
		if agent.AccessType != "read-only" && agent.AccessType != "read-write" {
			return fmt.Errorf("agent access type must be 'read-only' or 'read-write', got %s for agent %s", agent.AccessType, agent.ID)
		}

		// Validate mode
		if agent.Mode != "task" && agent.Mode != "interactive" {
			return fmt.Errorf("agent mode must be 'task' or 'interactive', got %s for agent %s", agent.Mode, agent.ID)
		}

		// Validate max concurrent executions
		if agent.MaxConcurrentExecutions < 1 {
			return fmt.Errorf("max concurrent executions must be at least 1, got %d for agent %s", agent.MaxConcurrentExecutions, agent.ID)
		}

		if agent.AccessType == "read-write" && agent.MaxConcurrentExecutions > 1 {
			return fmt.Errorf("read-write agents must have max concurrent executions of 1, got %d for agent %s", agent.MaxConcurrentExecutions, agent.ID)
		}
	}

	return nil
}