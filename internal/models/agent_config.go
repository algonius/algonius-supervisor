package models

import (
	"github.com/algonius/algonius-supervisor/pkg/types"
	"time"
)

// AgentConfiguration represents the settings for a generic CLI AI agent, including execution patterns,
// input/output handling, and access type (read-only vs read-write).
type AgentConfiguration struct {
	ID                    string            `json:"id"`
	Name                  string            `json:"name"`
	AgentType             string            `json:"agent_type"`
	ExecutablePath        string            `json:"executable_path"`
	WorkingDirectory      string            `json:"working_directory"`
	Envs                  map[string]string `json:"envs"`
	CliArgs               map[string]string `json:"cli_args"`
	Mode                  types.AgentMode   `json:"mode"`
	InputPattern          types.InputPattern `json:"input_pattern"`
	OutputPattern         types.OutputPattern `json:"output_pattern"`
	InputFileTemplate     string            `json:"input_file_template"`
	OutputFileTemplate    string            `json:"output_file_template"`
	AccessType            types.AgentAccessType `json:"access_type"`
	MaxConcurrentExecutions int             `json:"max_concurrent_executions"`
	Timeout               int               `json:"timeout"` // seconds
	SessionTimeout        int               `json:"session_timeout"` // seconds
	KeepAlive             bool              `json:"keep_alive"`
	Enabled               bool              `json:"enabled"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
}

// Validate validates the agent configuration fields
func (ac *AgentConfiguration) Validate() error {
	if ac.ID == "" {
		return ValidationError("AgentConfiguration ID cannot be empty")
	}

	if ac.Name == "" {
		return ValidationError("AgentConfiguration Name cannot be empty")
	}

	if ac.ExecutablePath == "" {
		return ValidationError("AgentConfiguration ExecutablePath cannot be empty")
	}

	if ac.AccessType != types.ReadOnlyAccessType && ac.AccessType != types.ReadWriteAccessType {
		return ValidationError("AgentConfiguration AccessType must be 'read-only' or 'read-write'")
	}

	if ac.MaxConcurrentExecutions < 1 {
		return ValidationError("AgentConfiguration MaxConcurrentExecutions must be at least 1")
	}

	if ac.AccessType == types.ReadWriteAccessType && ac.MaxConcurrentExecutions > 1 {
		return ValidationError("ReadWrite agents must have MaxConcurrentExecutions of 1")
	}

	// Validate mode
	if ac.Mode != types.TaskMode && ac.Mode != types.InteractiveMode {
		return ValidationError("AgentConfiguration Mode must be 'task' or 'interactive'")
	}

	// Validate input pattern
	switch ac.InputPattern {
	case types.StdinPattern, types.FilePattern, types.ArgsPattern, types.JsonRpcPattern:
		// Valid
	default:
		return ValidationError("AgentConfiguration InputPattern must be 'stdin', 'file', 'args', or 'json-rpc'")
	}

	// Validate output pattern
	switch ac.OutputPattern {
	case types.StdoutPattern, types.FilePatternOut, types.JsonRpcPatternOut:
		// Valid
	default:
		return ValidationError("AgentConfiguration OutputPattern must be 'stdout', 'file', or 'json-rpc'")
	}

	return nil
}

// ValidationError represents an error during validation
type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}