package models

import (
	"time"
)

// A2AEndpoint represents an A2A endpoint configuration
type A2AEndpoint struct {
	ID                   string           `json:"id" yaml:"id"`
	AgentID              string           `json:"agent_id" yaml:"agent_id"`
	BasePath             string           `json:"base_path" yaml:"base_path"`
	TransportProtocols   []string         `json:"transport_protocols" yaml:"transport_protocols"`
	Authentication       *A2AAuth         `json:"authentication" yaml:"authentication"`
	RateLimiting         *A2ARateLimit    `json:"rate_limiting" yaml:"rate_limiting"`
	Capabilities         *A2ACapabilities `json:"capabilities" yaml:"capabilities"`
	Enabled              bool             `json:"enabled" yaml:"enabled"`
	CreatedAt            time.Time        `json:"created_at" yaml:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at" yaml:"updated_at"`
}

// A2AAuth represents A2A authentication configuration
type A2AAuth struct {
	Type       string   `json:"type" yaml:"type"`              // "bearer_token", "api_key", "none"
	Token      string   `json:"token" yaml:"token"`            // Authentication token (from environment variable)
	HeaderName string   `json:"header_name" yaml:"header_name"` // HTTP header name for token
	Required   bool     `json:"required" yaml:"required"`      // Whether authentication is required
}

// A2ARateLimit represents A2A rate limiting configuration
type A2ARateLimit struct {
	RequestsPerSecond int `json:"requests_per_second" yaml:"requests_per_second"`
	BurstSize         int `json:"burst_size" yaml:"burst_size"`
}

// A2ACapabilities represents A2A agent capabilities for A2A protocol
type A2ACapabilities struct {
	SupportedMethods      []string `json:"supported_methods" yaml:"supported_methods"`
	MaxInputSize          int64    `json:"max_input_size" yaml:"max_input_size"`
	MaxOutputSize         int64    `json:"max_output_size" yaml:"max_output_size"`
	StreamingSupport      bool     `json:"streaming_support" yaml:"streaming_support"`
	ConcurrentExecution   bool     `json:"concurrent_execution" yaml:"concurrent_execution"`
	SupportedContentTypes []string `json:"supported_content_types" yaml:"supported_content_types"`
}

// Validate A2AEndpoint
func (e *A2AEndpoint) Validate() error {
	if e.ID == "" {
		return ValidationError("A2A endpoint ID cannot be empty")
	}
	if e.AgentID == "" {
		return ValidationError("A2A endpoint agent ID cannot be empty")
	}
	if e.BasePath == "" {
		return ValidationError("A2A endpoint base path cannot be empty")
	}
	if e.Authentication == nil {
		return ValidationError("A2A endpoint authentication config cannot be nil")
	}
	return nil
}

// Validate A2AAuth
func (a *A2AAuth) Validate() error {
	if a.Type == "" {
		return ValidationError("A2A auth type cannot be empty")
	}
	if a.Type != "bearer_token" && a.Type != "api_key" && a.Type != "none" {
		return ValidationError("A2A auth type must be 'bearer_token', 'api_key', or 'none'")
	}
	return nil
}

// Validate A2ACapabilities
func (c *A2ACapabilities) Validate() error {
	if c.SupportedMethods == nil || len(c.SupportedMethods) == 0 {
		return ValidationError("A2A capabilities must have at least one supported method")
	}
	if c.MaxInputSize <= 0 {
		return ValidationError("A2A capabilities max input size must be greater than 0")
	}
	if c.MaxOutputSize <= 0 {
		return ValidationError("A2A capabilities max output size must be greater than 0")
	}
	return nil
}