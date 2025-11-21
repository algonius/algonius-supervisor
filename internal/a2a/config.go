package a2a

import (
	"time"

	"github.com/algonius/algonius-supervisor/internal/models"
)

// A2AConfig represents the configuration for A2A protocol services
type A2AConfig struct {
	// Server configuration
	ServerAddress string `json:"server_address" yaml:"server_address"`
	ServerPort    int    `json:"server_port" yaml:"server_port"`

	// Authentication configuration
	Authentication A2AAuthenticationConfig `json:"authentication" yaml:"authentication"`

	// Protocol configuration
	Protocol A2AProtocolConfig `json:"protocol" yaml:"protocol"`

	// Transport configuration
	Transports A2ATransportConfig `json:"transports" yaml:"transports"`

	// Agent configuration
	Agents map[string]*models.AgentConfiguration `json:"agents" yaml:"agents"`
}

// A2AAuthenticationConfig contains authentication-related configuration
type A2AAuthenticationConfig struct {
	Required     bool     `json:"required" yaml:"required"`
	HeaderName   string   `json:"header_name" yaml:"header_name"`
	ValidTokens  []string `json:"valid_tokens" yaml:"valid_tokens"`
	TokenEnvVar  string   `json:"token_env_var" yaml:"token_env_var"` // Environment variable name for the token
}

// A2AProtocolConfig contains protocol-related configuration
type A2AProtocolConfig struct {
	Version       string        `json:"version" yaml:"version"`
	Timeout       time.Duration `json:"timeout" yaml:"timeout"`
	MaxMessageSize int64        `json:"max_message_size" yaml:"max_message_size"`
}

// A2ATransportConfig contains transport protocol configuration
type A2ATransportConfig struct {
	HTTPEnabled   bool `json:"http_enabled" yaml:"http_enabled"`
	GRPCEnabled   bool `json:"grpc_enabled" yaml:"grpc_enabled"`
	JSONRPCEnabled bool `json:"jsonrpc_enabled" yaml:"jsonrpc_enabled"`
	
	// Specific configuration for each transport
	HTTPConfig   A2AHTTPConfig   `json:"http_config" yaml:"http_config"`
	GRPCConfig   A2AGRPCConfig   `json:"grpc_config" yaml:"grpc_config"`
	JSONRPCConfig A2AJSONRPCConfig `json:"jsonrpc_config" yaml:"jsonrpc_config"`
}

// A2AHTTPConfig contains HTTP transport configuration
type A2AHTTPConfig struct {
	EnableCORS     bool     `json:"enable_cors" yaml:"enable_cors"`
	AllowedOrigins []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedHeaders []string `json:"allowed_headers" yaml:"allowed_headers"`
}

// A2AGRPCConfig contains gRPC transport configuration
type A2AGRPCConfig struct {
	EnableReflection bool `json:"enable_reflection" yaml:"enable_reflection"`
	MaxRecvMsgSize   int  `json:"max_recv_msg_size" yaml:"max_recv_msg_size"`
	MaxSendMsgSize   int  `json:"max_send_msg_size" yaml:"max_send_msg_size"`
}

// A2AJSONRPCConfig contains JSON-RPC transport configuration
type A2AJSONRPCConfig struct {
	// Currently just a placeholder - specific JSON-RPC config can be added as needed
	MaxBatchSize int `json:"max_batch_size" yaml:"max_batch_size"`
}

// DefaultA2AConfig returns a default A2A configuration
func DefaultA2AConfig() *A2AConfig {
	return &A2AConfig{
		ServerAddress: ":8080",
		ServerPort:    8080,
		Authentication: A2AAuthenticationConfig{
			Required:   true,
			HeaderName: "Authorization",
			ValidTokens: []string{}, // Tokens should be loaded from environment or config
			TokenEnvVar: "A2A_AUTH_TOKEN",
		},
		Protocol: A2AProtocolConfig{
			Version:       "0.3.0",
			Timeout:       30 * time.Second,
			MaxMessageSize: 10 * 1024 * 1024, // 10MB
		},
		Transports: A2ATransportConfig{
			HTTPEnabled:   true,
			GRPCEnabled:   true,
			JSONRPCEnabled: true,
			HTTPConfig: A2AHTTPConfig{
				EnableCORS:     true,
				AllowedOrigins: []string{"*"},
				AllowedHeaders: []string{"*"},
			},
			GRPCConfig: A2AGRPCConfig{
				EnableReflection: true,
				MaxRecvMsgSize:   4 * 1024 * 1024, // 4MB
				MaxSendMsgSize:   4 * 1024 * 1024, // 4MB
			},
			JSONRPCConfig: A2AJSONRPCConfig{
				MaxBatchSize: 10,
			},
		},
		Agents: make(map[string]*models.AgentConfiguration),
	}
}