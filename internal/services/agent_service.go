package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/algonius/algonius-supervisor/internal/models"
	"github.com/algonius/algonius-supervisor/pkg/types"
	"go.uber.org/zap"
)

// IAgentService interface for managing agent configurations and executions
type IAgentService interface {
	// ExecuteAgent executes an agent with the specified ID and input string
	ExecuteAgent(ctx context.Context, agentID string, input string) (*models.ExecutionResult, error)

	// ExecuteAgentWithParameters executes an agent with the specified ID, input string, and parameters
	ExecuteAgentWithParameters(ctx context.Context, agentID string, input string, parameters map[string]interface{}) (*models.ExecutionResult, error)

	// ExecuteAgentWithOptions executes an agent with the specified ID, input string, parameters, working directory, and environment variables
	ExecuteAgentWithOptions(ctx context.Context, agentID string, input string, parameters map[string]interface{}, workingDir string, envVars map[string]string) (*models.ExecutionResult, error)

	// GetAgentStatus returns the status of an agent with the specified ID
	GetAgentStatus(agentID string) (*AgentStatus, error)

	// GetAgentExecution returns the execution details for the specified execution ID
	GetAgentExecution(executionID string) (*models.AgentExecution, error)

	// ListAgentExecutions returns a list of executions for the specified agent ID
	ListAgentExecutions(agentID string, limit int) ([]*models.AgentExecution, error)

	// ListActiveExecutions returns a list of all currently active executions
	ListActiveExecutions() ([]*models.AgentExecution, error)

	// CancelExecution cancels the execution with the specified ID
	CancelExecution(executionID string) error

	// ListAgents returns a list of all available agent configurations
	ListAgents() ([]*models.AgentConfiguration, error)

	// GetAgent returns the configuration for an agent with the specified ID
	GetAgent(agentID string) (*models.AgentConfiguration, error)

	// RegisterAgent registers a new agent configuration
	RegisterAgent(config *models.AgentConfiguration) error

	// UpdateAgent updates an existing agent configuration
	UpdateAgent(config *models.AgentConfiguration) error

	// DeleteAgent deletes an agent configuration with the specified ID
	DeleteAgent(agentID string) error
}

// AgentStatus represents the current status of an agent
type AgentStatus struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Status      string                    `json:"status"`      // "idle", "running", "error", "disabled"
	Mode        types.AgentMode          `json:"mode"`
	LastRun     *time.Time                `json:"last_run"`    // Time of last execution
	NextRun     *time.Time                `json:"next_run"`    // Time of next scheduled execution (if applicable)
	ActiveTasks int                       `json:"active_tasks"` // Number of active tasks
	Executions  []*models.AgentExecution  `json:"executions"`
	Health      AgentHealthStatus         `json:"health"`      // Health status of the agent
}

// AgentHealthStatus represents the health status of an agent
type AgentHealthStatus string

const (
	AgentHealthy     AgentHealthStatus = "healthy"     // Agent is operating normally
	AgentDegraded    AgentHealthStatus = "degraded"    // Agent is operating with reduced capacity
	AgentUnhealthy   AgentHealthStatus = "unhealthy"   // Agent is not operating correctly
	AgentUnknown     AgentHealthStatus = "unknown"     // Agent health status is unknown
)

// AgentService provides a concrete implementation of IAgentService
type AgentService struct {
	// Agents is a map of agent configurations by ID
	Agents map[string]*models.AgentConfiguration

	// ActiveExecutions tracks currently running executions
	ActiveExecutions map[string]*models.AgentExecution

	// ExecutionResults stores the results of completed executions
	ExecutionResults map[string]*models.ExecutionResult

	// logger for logging
	logger *zap.Logger
}

// NewAgentService creates a new instance of AgentService
func NewAgentService(logger *zap.Logger) *AgentService {
	if logger == nil {
		// Create a fallback logger if none provided
		var err error
		logger, err = zap.NewProduction()
		if err != nil {
			// If we can't create production logger, use development logger
			logger, _ = zap.NewDevelopment()
		}
	}

	return &AgentService{
		Agents:           make(map[string]*models.AgentConfiguration),
		ActiveExecutions: make(map[string]*models.AgentExecution),
		ExecutionResults: make(map[string]*models.ExecutionResult),
		logger:           logger,
	}
}

// RegisterAgent registers a new agent configuration
func (as *AgentService) RegisterAgent(config *models.AgentConfiguration) error {
	if config == nil {
		return errors.New("agent configuration cannot be nil")
	}

	// Validate configuration comprehensively
	if err := as.ValidateAgentConfiguration(config); err != nil {
		as.logger.Error("invalid agent configuration", zap.Error(err))
		return fmt.Errorf("invalid agent configuration: %w", err)
	}

	// Check if agent with this ID already exists
	if _, exists := as.Agents[config.ID]; exists {
		return fmt.Errorf("agent with ID %s already exists", config.ID)
	}

	// Set timestamps
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	// Store the agent configuration
	as.Agents[config.ID] = config

	as.logger.Info("agent registered successfully",
		zap.String("agent_id", config.ID),
		zap.String("agent_name", config.Name))

	return nil
}

// GetAgent returns the configuration for an agent with the specified ID
func (as *AgentService) GetAgent(agentID string) (*models.AgentConfiguration, error) {
	config, exists := as.Agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}

	return config, nil
}

// ListAgents returns a list of all available agent configurations
func (as *AgentService) ListAgents() ([]*models.AgentConfiguration, error) {
	var configs []*models.AgentConfiguration
	for _, config := range as.Agents {
		configs = append(configs, config)
	}

	return configs, nil
}

// UpdateAgent updates an existing agent configuration
func (as *AgentService) UpdateAgent(config *models.AgentConfiguration) error {
	if config == nil {
		return errors.New("agent configuration cannot be nil")
	}

	// Validate configuration comprehensively
	if err := as.ValidateAgentConfiguration(config); err != nil {
		as.logger.Error("invalid agent configuration", zap.Error(err))
		return fmt.Errorf("invalid agent configuration: %w", err)
	}

	// Check if agent with this ID exists
	_, exists := as.Agents[config.ID]
	if !exists {
		return fmt.Errorf("agent with ID %s does not exist", config.ID)
	}

	// Update timestamps
	config.UpdatedAt = time.Now()

	// Update the agent configuration
	as.Agents[config.ID] = config

	as.logger.Info("agent updated successfully",
		zap.String("agent_id", config.ID),
		zap.String("agent_name", config.Name))

	return nil
}

// ValidateAgentConfiguration performs comprehensive validation of an agent configuration
func (as *AgentService) ValidateAgentConfiguration(config *models.AgentConfiguration) error {
	if config == nil {
		return errors.New("agent configuration cannot be nil")
	}

	// Use the built-in Validate method first
	if err := config.Validate(); err != nil {
		return fmt.Errorf("basic validation failed: %w", err)
	}

	// Perform additional validation for working directory and environment variables (T036)
	if err := as.validateWorkingDirectoryAndEnvVars(config); err != nil {
		return fmt.Errorf("working directory or environment variable validation failed: %w", err)
	}

	// Perform input/output pattern validation (T037)
	if err := as.validateInputOutputPatterns(config); err != nil {
		return fmt.Errorf("input/output pattern validation failed: %w", err)
	}

	// Perform access type validation (T038)
	if err := as.validateAccessType(config); err != nil {
		return fmt.Errorf("access type validation failed: %w", err)
	}

	return nil
}

// validateWorkingDirectoryAndEnvVars validates the working directory and environment variables (T036)
func (as *AgentService) validateWorkingDirectoryAndEnvVars(config *models.AgentConfiguration) error {
	// Validate working directory exists if specified
	if config.WorkingDirectory != "" {
		// In a real implementation, we would check if the directory exists
		// For now, we'll just log it
		as.logger.Debug("validating working directory",
			zap.String("agent_id", config.ID),
			zap.String("working_directory", config.WorkingDirectory))
	}

	// Validate environment variables don't contain sensitive data in their keys
	for key := range config.Envs {
		if as.isPotentialSensitiveKey(key) {
			return fmt.Errorf("environment variable key '%s' might contain sensitive information", key)
		}
	}

	return nil
}

// validateInputOutputPatterns validates the input and output patterns (T037)
func (as *AgentService) validateInputOutputPatterns(config *models.AgentConfiguration) error {
	// Validate input/output patterns are compatible
	// For example, if using file patterns, ensure the templates are properly formatted

	if config.InputPattern == models.FilePattern && config.InputFileTemplate == "" {
		return fmt.Errorf("input file pattern requires an input file template")
	}

	if config.OutputPattern == models.FilePatternOut && config.OutputFileTemplate == "" {
		return fmt.Errorf("output file pattern requires an output file template")
	}

	// Additional pattern compatibility checks can be added here
	switch {
	case config.InputPattern == models.JsonRpcPattern && config.OutputPattern != models.JsonRpcPatternOut:
		as.logger.Warn("input/output patterns might not be compatible",
			zap.String("agent_id", config.ID),
			zap.String("input_pattern", string(config.InputPattern)),
			zap.String("output_pattern", string(config.OutputPattern)),
		)
	case config.OutputPattern == models.JsonRpcPatternOut && config.InputPattern != models.JsonRpcPattern:
		as.logger.Warn("input/output patterns might not be compatible",
			zap.String("agent_id", config.ID),
			zap.String("input_pattern", string(config.InputPattern)),
			zap.String("output_pattern", string(config.OutputPattern)),
		)
	}

	return nil
}

// validateAccessType validates the access type (T038)
func (as *AgentService) validateAccessType(config *models.AgentConfiguration) error {
	// Already validated in basic validation, but we can add more complex logic here
	// For example, ensure MaxConcurrentExecutions is properly set based on AccessType
	switch config.AccessType {
	case models.ReadOnlyAccessType:
		if config.MaxConcurrentExecutions <= 0 {
			config.MaxConcurrentExecutions = 10 // Default to 10 for read-only agents
		}
	case models.ReadWriteAccessType:
		if config.MaxConcurrentExecutions != 1 {
			return fmt.Errorf("read-write agents must have MaxConcurrentExecutions equal to 1, got %d", config.MaxConcurrentExecutions)
		}
	default:
		return fmt.Errorf("invalid access type: %s", config.AccessType)
	}

	return nil
}

// isPotentialSensitiveKey checks if an environment variable key might contain sensitive data
func (as *AgentService) isPotentialSensitiveKey(key string) bool {
	// Convert to lowercase for comparison
	lowerKey := strings.ToLower(key)

	// Check for common sensitive key patterns
	sensitivePatterns := []string{
		"password", "secret", "key", "token", "auth", "credential", "private", "api", "cert", "ssl", "tls",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerKey, pattern) {
			return true
		}
	}

	return false
}

// DeleteAgent deletes an agent configuration with the specified ID
func (as *AgentService) DeleteAgent(agentID string) error {
	// Check if agent with this ID exists
	_, exists := as.Agents[agentID]
	if !exists {
		return fmt.Errorf("agent with ID %s not found", agentID)
	}

	// TODO: Check if agent is currently in use before deletion
	// For now, we'll allow deletion

	// Delete the agent configuration
	delete(as.Agents, agentID)

	as.logger.Info("agent deleted successfully",
		zap.String("agent_id", agentID))

	return nil
}

// ExecuteAgent executes an agent with the specified ID and input string
func (as *AgentService) ExecuteAgent(ctx context.Context, agentID string, input string) (*models.ExecutionResult, error) {
	// This method will be implemented in the execution engine part of Phase 3
	return nil, errors.New("not implemented yet - will be implemented in execution engine section")
}

// ExecuteAgentWithParameters executes an agent with the specified ID, input string, and parameters
func (as *AgentService) ExecuteAgentWithParameters(ctx context.Context, agentID string, input string, parameters map[string]interface{}) (*models.ExecutionResult, error) {
	// This method will be implemented in the execution engine part of Phase 3
	return nil, errors.New("not implemented yet - will be implemented in execution engine section")
}

// ExecuteAgentWithOptions executes an agent with the specified ID, input string, parameters, working directory, and environment variables
func (as *AgentService) ExecuteAgentWithOptions(ctx context.Context, agentID string, input string, parameters map[string]interface{}, workingDir string, envVars map[string]string) (*models.ExecutionResult, error) {
	// This method will be implemented in the execution engine part of Phase 3
	return nil, errors.New("not implemented yet - will be implemented in execution engine section")
}

// GetAgentStatus returns the status of an agent with the specified ID
func (as *AgentService) GetAgentStatus(agentID string) (*AgentStatus, error) {
	config, exists := as.Agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", agentID)
	}

	// For now, we'll set a simple status based on whether there are active executions
	status := "idle"
	activeExecutions, err := as.ListActiveExecutions()
	if err != nil {
		status = "error"
	} else if len(activeExecutions) > 0 {
		status = "running"
	}

	agentStatus := &AgentStatus{
		ID:     config.ID,
		Name:   config.Name,
		Status: status,
		Mode:   config.Mode,
		Health: AgentHealthy, // Default to healthy
	}

	// Find executions for this agent
	var executions []*models.AgentExecution
	for _, exec := range as.ActiveExecutions {
		if exec.AgentID == agentID {
			executions = append(executions, exec)
		}
	}
	agentStatus.Executions = executions

	return agentStatus, nil
}

// GetAgentExecution returns the execution details for the specified execution ID
func (as *AgentService) GetAgentExecution(executionID string) (*models.AgentExecution, error) {
	// Check active executions first
	if exec, exists := as.ActiveExecutions[executionID]; exists {
		return exec, nil
	}

	// For now, we'll return not found
	// In a complete implementation, we'd check a persistence store for completed executions
	return nil, fmt.Errorf("execution with ID %s not found", executionID)
}

// ListAgentExecutions returns a list of executions for the specified agent ID
func (as *AgentService) ListAgentExecutions(agentID string, limit int) ([]*models.AgentExecution, error) {
	var executions []*models.AgentExecution
	count := 0

	// Add active executions
	for _, exec := range as.ActiveExecutions {
		if exec.AgentID == agentID {
			executions = append(executions, exec)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	// For a complete implementation, we'd also load completed executions from storage
	return executions, nil
}

// ListActiveExecutions returns a list of all currently active executions
func (as *AgentService) ListActiveExecutions() ([]*models.AgentExecution, error) {
	var executions []*models.AgentExecution
	for _, exec := range as.ActiveExecutions {
		executions = append(executions, exec)
	}
	return executions, nil
}

// CancelExecution cancels the execution with the specified ID
func (as *AgentService) CancelExecution(executionID string) error {
	// For now, we'll return an error indicating not implemented
	// This would involve actual process cancellation in a complete implementation
	return errors.New("cancel execution not implemented yet")
}