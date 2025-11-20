package models

import (
	"github.com/algonius/algonius-supervisor/pkg/types"
	"time"
)

// AgentExecution represents the lifecycle state and metadata of an active or recent agent execution.
// This entity tracks the execution state machine and provides real-time execution information.
type AgentExecution struct {
	ID               string                 `json:"id"`
	AgentID          string                 `json:"agent_id"`
	TaskID           string                 `json:"task_id"` // Reference to the scheduled task ID that triggered execution (if applicable)
	State            types.AgentState       `json:"state"`
	PreviousState    types.AgentState       `json:"previous_state"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          *time.Time             `json:"end_time"` // nil if still running
	LastStateChange  time.Time              `json:"last_state_change"`
	Input            string                 `json:"input"` // sanitized of sensitive data
	ProcessID        int                    `json:"process_id"` // OS process ID of the executed agent (if applicable)
	ExitCode         int                    `json:"exit_code"`
	ErrorMessage     string                 `json:"error_message"`
	ErrorCategory    types.ErrorCategory    `json:"error_category"`
	RetryCount       int                    `json:"retry_count"`
	MaxRetries       int                    `json:"max_retries"`
	Timeout          int                    `json:"timeout"` // seconds
	ResourceUsage    *ResourceUsage         `json:"resource_usage"`
	Context          map[string]interface{} `json:"context"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// Validate validates the agent execution fields
func (ae *AgentExecution) Validate() error {
	if ae.ID == "" {
		return ValidationError("AgentExecution ID cannot be empty")
	}

	if ae.AgentID == "" {
		return ValidationError("AgentExecution AgentID cannot be empty")
	}

	if !isValidAgentState(ae.State) {
		return ValidationError("AgentExecution State must be a valid state value")
	}

	if ae.RetryCount > ae.MaxRetries {
		return ValidationError("AgentExecution RetryCount cannot exceed MaxRetries")
	}

	return nil
}

// isValidAgentState checks if the state is one of the valid AgentState values
func isValidAgentState(state types.AgentState) bool {
	switch state {
	case types.IdleState, types.StartingState, types.RunningState, types.CompletedState,
	     types.FailedState, types.CleanupState, types.TimeoutState, types.CancelledState:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if the current state can transition to the target state
func (ae *AgentExecution) CanTransitionTo(targetState types.AgentState) bool {
	allowedTransitions := map[types.AgentState][]types.AgentState{
		types.IdleState:      {types.StartingState},
		types.StartingState:  {types.RunningState, types.FailedState},
		types.RunningState:   {types.CompletedState, types.FailedState, types.TimeoutState, types.CancelledState},
		types.CompletedState: {types.CleanupState},
		types.FailedState:    {types.CleanupState, types.StartingState}, // Allow retry from Failed state
		types.TimeoutState:   {types.CleanupState},
		types.CancelledState: {types.CleanupState},
		types.CleanupState:   {types.IdleState},
	}

	allowed, exists := allowedTransitions[ae.State]
	if !exists {
		return false
	}

	for _, validState := range allowed {
		if targetState == validState {
			return true
		}
	}

	return false
}

// UpdateState updates the state and timestamp for the execution
func (ae *AgentExecution) UpdateState(newState types.AgentState) error {
	if !ae.CanTransitionTo(newState) {
		return ValidationError("Invalid state transition from " + string(ae.State) + " to " + string(newState))
	}

	ae.PreviousState = ae.State
	ae.State = newState
	ae.LastStateChange = time.Now()
	ae.UpdatedAt = time.Now()

	return nil
}

// IsComplete returns true if the execution is in a completed state
func (ae *AgentExecution) IsComplete() bool {
	switch ae.State {
	case types.CompletedState, types.FailedState, types.TimeoutState, types.CancelledState, types.CleanupState:
		return true
	default:
		return false
	}
}

// IsRunning returns true if the execution is currently running
func (ae *AgentExecution) IsRunning() bool {
	return ae.State == types.RunningState
}