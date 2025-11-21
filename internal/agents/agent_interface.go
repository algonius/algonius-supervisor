package agents

import (
	"context"

	"github.com/algonius/algonius-supervisor/internal/models"
)

// IAgent interface representing an executable agent
type IAgent interface {
	// Execute runs the agent with the given input
	Execute(ctx context.Context, input string) (*models.ExecutionResult, error)
	
	// GetID returns the agent's ID
	GetID() string
	
	// GetName returns the agent's name
	GetName() string
	
	// GetType returns the agent's type
	GetType() string
	
	// IsReadOnly returns whether the agent is read-only or read-write
	IsReadOnly() bool
	
	// GetConfig returns the agent's configuration
	GetConfig() *models.AgentConfiguration
	
	// Validate checks if the agent configuration is valid
	Validate() error
}

// IAgentFactory interface for creating agents based on configuration
type IAgentFactory interface {
	// CreateAgent creates an agent based on the provided configuration
	CreateAgent(config *models.AgentConfiguration) (IAgent, error)
	
	// GetSupportedTypes returns a list of supported agent types
	GetSupportedTypes() []string
	
	// ValidateConfig validates an agent configuration
	ValidateConfig(config *models.AgentConfiguration) error
}