package a2a

import (
	"context"
	"time"

	"github.com/a2aproject/a2a-go/a2a"
)

// Protocol represents the A2A protocol interface
type Protocol interface {
	// SendMessage sends a message to an agent
	SendMessage(ctx context.Context, params *pkg.MessageSendParams) (*pkg.MessageSendResponse, error)
	
	// StreamMessage sends a message and returns a stream of updates
	StreamMessage(ctx context.Context, params *pkg.MessageSendParams) (<-chan pkg.Message, error)
	
	// GetTask retrieves a task by ID
	GetTask(ctx context.Context, params *pkg.TaskGetParams) (*pkg.TaskGetResponse, error)
	
	// ListTasks lists tasks with optional filters
	ListTasks(ctx context.Context, params *pkg.TaskListParams) (*pkg.TaskListResponse, error)
	
	// CancelTask cancels a running task
	CancelTask(ctx context.Context, params *pkg.TaskCancelParams) (*pkg.TaskCancelResponse, error)
}

// A2ARequest represents the internal representation of an A2A request
type A2ARequest struct {
	Protocol  string      `json:"protocol"`
	Version   string      `json:"version"`
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Context   A2AContext  `json:"context"`
	Payload   A2APayload  `json:"payload"`
}

// A2AContext represents the context of an A2A message
type A2AContext struct {
	From          string `json:"from"`
	To            string `json:"to"`
	ConversationID string `json:"conversation_id"`
	MessageID     string `json:"message_id"`
}

// A2APayload represents the payload of an A2A message
type A2APayload struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
	Result interface{}            `json:"result,omitempty"`
	Error  *A2AError              `json:"error,omitempty"`
}

// A2AError represents an A2A error
type A2AError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// A2AEndpoint represents an A2A endpoint configuration
type A2AEndpoint struct {
	AgentID            string           `json:"agent_id" yaml:"agent_id"`
	BasePath           string           `json:"base_path" yaml:"base_path"`
	TransportProtocols []string         `json:"transport_protocols" yaml:"transport_protocols"`
	Authentication     *A2AAuth         `json:"authentication" yaml:"authentication"`
	RateLimiting       *A2ARateLimit    `json:"rate_limiting" yaml:"rate_limiting"`
	Capabilities       *A2ACapabilities `json:"capabilities" yaml:"capabilities"`
}

// A2AAuth represents A2A authentication configuration
type A2AAuth struct {
	Type       string   `json:"type" yaml:"type"`
	Token      string   `json:"token" yaml:"token"`
	HeaderName string   `json:"header_name" yaml:"header_name"`
	Required   bool     `json:"required" yaml:"required"`
}

// A2ARateLimit represents A2A rate limiting configuration
type A2ARateLimit struct {
	RequestsPerSecond int `json:"requests_per_second" yaml:"requests_per_second"`
	BurstSize         int `json:"burst_size" yaml:"burst_size"`
}

// A2ACapabilities represents A2A agent capabilities
type A2ACapabilities struct {
	SupportedMethods     []string `json:"supported_methods" yaml:"supported_methods"`
	MaxInputSize         int64    `json:"max_input_size" yaml:"max_input_size"`
	MaxOutputSize        int64    `json:"max_output_size" yaml:"max_output_size"`
	StreamingSupport     bool     `json:"streaming_support" yaml:"streaming_support"`
	ConcurrentExecution  bool     `json:"concurrent_execution" yaml:"concurrent_execution"`
	SupportedContentTypes []string `json:"supported_content_types" yaml:"supported_content_types"`
}