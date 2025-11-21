package agents

import (
	"fmt"

	"github.com/algonius/algonius-supervisor/internal/models"
)

// AgentFactory implements the IAgentFactory interface
type AgentFactory struct {
	// supportedTypes is a list of agent types that can be created
	supportedTypes []string
}

// NewAgentFactory creates a new instance of AgentFactory
func NewAgentFactory() *AgentFactory {
	return &AgentFactory{
		supportedTypes: []string{
			"generic",
			"cli",
			"stdin-stdout",
			"file-io",
			"json-rpc",
			"claude",
			"codex",
			"gemini",
		},
	}
}

// CreateAgent creates an agent based on the provided configuration
func (af *AgentFactory) CreateAgent(config *models.AgentConfiguration) (IAgent, error) {
	if config == nil {
		return nil, fmt.Errorf("agent configuration cannot be nil")
	}

	// Validate the config first
	if err := af.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid agent configuration: %w", err)
	}

	// Create the appropriate agent based on the configuration
	// For now, we'll create a generic agent for all types
	// In a more complex implementation, we might have different agent types
	agent := &GenericAgent{
		config: config,
	}

	return agent, nil
}

// GetSupportedTypes returns a list of supported agent types
func (af *AgentFactory) GetSupportedTypes() []string {
	return af.supportedTypes
}

// ValidateConfig validates an agent configuration
func (af *AgentFactory) ValidateConfig(config *models.AgentConfiguration) error {
	if config == nil {
		return fmt.Errorf("agent configuration cannot be nil")
	}

	// Validate the configuration using the built-in Validate method
	if err := config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Check if the agent type is supported
	isSupported := false
	for _, supportedType := range af.supportedTypes {
		if config.AgentType == supportedType {
			isSupported = true
			break
		}
	}
	
	if !isSupported {
		return fmt.Errorf("unsupported agent type: %s", config.AgentType)
	}

	return nil
}

// IsAgentTypeSupported checks if the specified agent type is supported
func (af *AgentFactory) IsAgentTypeSupported(agentType string) bool {
	for _, supportedType := range af.supportedTypes {
		if supportedType == agentType {
			return true
		}
	}
	return false
}