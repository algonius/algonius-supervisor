package a2a

import (
	"context"
	"time"
)

// MessageSendParams represents parameters for sending a message
type MessageSendParams struct {
	AgentID string                 `json:"agent_id"`
	Message *A2AMessage            `json:"message"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// MessageSendResponse represents the response for sending a message
type MessageSendResponse struct {
	MessageID string      `json:"message_id"`
	Status    string      `json:"status"`
	Data      interface{} `json:"data,omitempty"`
}

// TaskGetParams represents parameters for getting a task
type TaskGetParams struct {
	TaskID string `json:"task_id"`
}

// TaskGetResponse represents the response for getting a task
type TaskGetResponse struct {
	Task *A2ATask `json:"task"`
}

// TaskListParams represents parameters for listing tasks
type TaskListParams struct {
	AgentID string `json:"agent_id,omitempty"`
	Status  string `json:"status,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
}

// TaskListResponse represents the response for listing tasks
type TaskListResponse struct {
	Tasks []A2ATask `json:"tasks"`
	Total int       `json:"total"`
}

// TaskCancelParams represents parameters for cancelling a task
type TaskCancelParams struct {
	TaskID string `json:"task_id"`
	Reason string `json:"reason,omitempty"`
}

// TaskCancelResponse represents the response for cancelling a task
type TaskCancelResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

// Protocol represents the A2A protocol interface
type Protocol interface {
	// SendMessage sends a message to an agent
	SendMessage(ctx context.Context, params *MessageSendParams) (*MessageSendResponse, error)

	// StreamMessage sends a message and returns a stream of updates
	StreamMessage(ctx context.Context, params *MessageSendParams) (<-chan A2AMessage, error)

	// GetTask retrieves a task by ID
	GetTask(ctx context.Context, params *TaskGetParams) (*TaskGetResponse, error)

	// ListTasks lists tasks with optional filters
	ListTasks(ctx context.Context, params *TaskListParams) (*TaskListResponse, error)

	// CancelTask cancels a running task
	CancelTask(ctx context.Context, params *TaskCancelParams) (*TaskCancelResponse, error)
}

// A2ARequest represents the internal representation of an A2A request
type A2ARequest struct {
	Protocol  string     `json:"protocol"`
	Version   string     `json:"version"`
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Timestamp time.Time  `json:"timestamp"`
	Context   A2AContext `json:"context"`
	Payload   A2APayload `json:"payload"`
}

// A2AContext represents the context of an A2A message
type A2AContext struct {
	From           string `json:"from"`
	To             string `json:"to"`
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
}

// A2APayload represents the payload of an A2A message
type A2APayload struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
	Result interface{}            `json:"result,omitempty"`
	Error  *A2AError              `json:"error,omitempty"`
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
	Type       string `json:"type" yaml:"type"`
	Token      string `json:"token" yaml:"token"`
	HeaderName string `json:"header_name" yaml:"header_name"`
	Required   bool   `json:"required" yaml:"required"`
}

// A2ARateLimit represents A2A rate limiting configuration
type A2ARateLimit struct {
	RequestsPerSecond int `json:"requests_per_second" yaml:"requests_per_second"`
	BurstSize         int `json:"burst_size" yaml:"burst_size"`
}

// A2ACapabilities represents A2A agent capabilities
type A2ACapabilities struct {
	SupportedMethods      []string `json:"supported_methods" yaml:"supported_methods"`
	MaxInputSize          int64    `json:"max_input_size" yaml:"max_input_size"`
	MaxOutputSize         int64    `json:"max_output_size" yaml:"max_output_size"`
	StreamingSupport      bool     `json:"streaming_support" yaml:"streaming_support"`
	ConcurrentExecution   bool     `json:"concurrent_execution" yaml:"concurrent_execution"`
	SupportedContentTypes []string `json:"supported_content_types" yaml:"supported_content_types"`
}
