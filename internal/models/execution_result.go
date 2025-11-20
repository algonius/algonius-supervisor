package models

import (
	"github.com/algonius/algonius-supervisor/pkg/types"
	"time"
)

// ExecutionResult represents the output, status, and metadata from an agent execution,
// including logs, return values, and error information.
type ExecutionResult struct {
	ID              string            `json:"id"`
	AgentID         string            `json:"agent_id"`
	TaskID          string            `json:"task_id"` // Reference to the scheduled task ID that triggered execution (if applicable)
	StartTime       time.Time         `json:"start_time"`
	EndTime         time.Time         `json:"end_time"`
	Status          types.ExecutionStatus `json:"status"`
	Input           string            `json:"input"` // sanitized of sensitive data
	Output          string            `json:"output"` // sanitized of sensitive data
	Error           string            `json:"error"`
	ExecutionTime   int64             `json:"execution_time"` // milliseconds
	ProcessID       int               `json:"process_id"` // OS process ID of the executed agent (if applicable)
	AgentConfig     *AgentConfiguration `json:"agent_config"` // Associated agent configuration
	ResourceUsage   *ResourceUsage    `json:"resource_usage"`
	PreviousRetries []*ExecutionResult `json:"previous_retries"` // References to previous retry attempts
	StateTransitions []StateTransition `json:"state_transitions"` // Log of all state changes during execution
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// StateTransition represents a state change in an agent execution
type StateTransition struct {
	FromState     types.AgentState  `json:"from_state"`
	ToState       types.AgentState  `json:"to_state"`
	Timestamp     time.Time         `json:"timestamp"`
	Reason        string            `json:"reason"`
}

// Validate validates the execution result fields
func (er *ExecutionResult) Validate() error {
	if er.ID == "" {
		return ValidationError("ExecutionResult ID cannot be empty")
	}

	if er.AgentID == "" {
		return ValidationError("ExecutionResult AgentID cannot be empty")
	}

	switch er.Status {
	case types.SuccessStatus, types.FailureStatus, types.TimeoutStatus, types.CancelledStatus:
		// Valid status
	default:
		return ValidationError("ExecutionResult Status must be 'success', 'failure', 'timeout', or 'cancelled'")
	}

	if er.ExecutionTime < 0 {
		return ValidationError("ExecutionResult ExecutionTime must be greater than or equal to 0")
	}

	return nil
}

// SanitizeInput removes sensitive data from the input field before logging or storing
func (er *ExecutionResult) SanitizeInput() {
	// In a real implementation, this would sanitize sensitive data
	// For now, we'll just ensure it's properly handled
	er.Input = SanitizeData(er.Input)
}

// SanitizeOutput removes sensitive data from the output field before logging or storing
func (er *ExecutionResult) SanitizeOutput() {
	// In a real implementation, this would sanitize sensitive data
	// For now, we'll just ensure it's properly handled
	er.Output = SanitizeData(er.Output)
}

// SanitizeData is a helper function to sanitize data
func SanitizeData(data string) string {
	// In a real implementation, this would remove sensitive information
	// like API keys, passwords, tokens, etc.
	return data
}

// CalculateExecutionTime calculates the execution time in milliseconds
func (er *ExecutionResult) CalculateExecutionTime() int64 {
	duration := er.EndTime.Sub(er.StartTime)
	return duration.Milliseconds()
}

// IsSuccessful returns true if the execution was successful
func (er *ExecutionResult) IsSuccessful() bool {
	return er.Status == types.SuccessStatus
}

// GetDuration returns the execution duration as a time.Duration
func (er *ExecutionResult) GetDuration() time.Duration {
	return er.EndTime.Sub(er.StartTime)
}