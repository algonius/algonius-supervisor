package models

import "time"

// ScheduledTask represents an automated task that executes an agent at specified intervals or times,
// including timing configuration and execution context.
type ScheduledTask struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	AgentID          string                 `json:"agent_id"` // Reference to the agent configuration ID to execute
	CronExpression   string                 `json:"cron_expression"`
	Enabled          bool                   `json:"enabled"`
	LastExecution    *time.Time             `json:"last_execution"` // Time of last execution
	NextExecution    *time.Time             `json:"next_execution"` // Time of next scheduled execution
	InputParameters  map[string]interface{} `json:"input_parameters"` // Parameters to pass to the agent during execution
	Active           bool                   `json:"active"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	LastResult       *ExecutionResult       `json:"last_result"` // Result of the last execution
	MaxRetries       int                    `json:"max_retries"` // Maximum number of retry attempts for failed executions
	RetryCount       int                    `json:"retry_count"` // Current retry count
	Timeout          int                    `json:"timeout"` // Execution timeout in seconds
	Description      string                 `json:"description"` // Optional description of the task
	Owner            string                 `json:"owner"` // Optional owner of the task
	Tags             []string               `json:"tags"` // Optional tags for task categorization
}

// Validate validates the scheduled task fields
func (st *ScheduledTask) Validate() error {
	if st.ID == "" {
		return ValidationError("ScheduledTask ID cannot be empty")
	}

	if st.Name == "" {
		return ValidationError("ScheduledTask Name cannot be empty")
	}

	if st.AgentID == "" {
		return ValidationError("ScheduledTask AgentID cannot be empty")
	}

	// Note: We're not validating the cron expression format here to avoid adding a dependency
	// In a real implementation, you might want to use a library like "github.com/robfig/cron"
	if st.CronExpression == "" {
		return ValidationError("ScheduledTask CronExpression cannot be empty")
	}

	// Validate max retries
	if st.MaxRetries < 0 {
		return ValidationError("ScheduledTask MaxRetries cannot be negative")
	}

	// Validate timeout
	if st.Timeout < 0 {
		return ValidationError("ScheduledTask Timeout cannot be negative")
	}

	return nil
}

// IsActive returns true if the task is currently active and enabled
func (st *ScheduledTask) IsActive() bool {
	return st.Enabled && st.Active
}

// ShouldExecute checks if the task should be executed at the given time
func (st *ScheduledTask) ShouldExecute(at time.Time) bool {
	if !st.IsActive() {
		return false
	}

	// In a real implementation, this would parse the cron expression and check if the time matches
	// For now, we'll just return true for testing purposes
	// A proper implementation would use a cron library to evaluate this
	return true
}

// UpdateLastExecution updates the last execution time
func (st *ScheduledTask) UpdateLastExecution(executionTime time.Time) {
	st.LastExecution = &executionTime
	st.UpdatedAt = time.Time{}
}

// UpdateNextExecution updates the next execution time based on the cron expression
func (st *ScheduledTask) UpdateNextExecution() {
	// In a real implementation, this would calculate the next execution time
	// based on the cron expression and the current time
	// For now, we'll just return a zero time
	st.NextExecution = nil
}

// CanRetry returns true if the task can be retried based on the current retry count
func (st *ScheduledTask) CanRetry() bool {
	return st.RetryCount < st.MaxRetries
}

// IncrementRetryCount increments the retry count
func (st *ScheduledTask) IncrementRetryCount() {
	st.RetryCount++
	st.UpdatedAt = time.Time{}
}