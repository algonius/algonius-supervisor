package models

import (
	"fmt"
	"time"
)

// OperationType represents the type of operation performed on an agent
type OperationType string

const (
	OperationTypeStart   OperationType = "start"
	OperationTypeStop    OperationType = "stop"
	OperationTypeRestart OperationType = "restart"
	OperationTypeStatus  OperationType = "status"
	OperationTypeTail    OperationType = "tail"
	OperationTypeLogs    OperationType = "logs"
)

// String returns the string representation of the operation type
func (ot OperationType) String() string {
	return string(ot)
}

// OperationStatus represents the status of an operation
type OperationStatus string

const (
	OperationStatusSuccess    OperationStatus = "success"
	OperationStatusFailed     OperationStatus = "failed"
	OperationStatusPending    OperationStatus = "pending"
	OperationStatusTimeout    OperationStatus = "timeout"
	OperationStatusCancelled  OperationStatus = "cancelled"
)

// String returns the string representation of the operation status
func (os OperationStatus) String() string {
	return string(os)
}

// IsSuccessful returns true if the operation was successful
func (os OperationStatus) IsSuccessful() bool {
	return os == OperationStatusSuccess
}

// IsFailed returns true if the operation failed
func (os OperationStatus) IsFailed() bool {
	return os == OperationStatusFailed || os == OperationStatusTimeout || os == OperationStatusCancelled
}

// IsPending returns true if the operation is still pending
func (os OperationStatus) IsPending() bool {
	return os == OperationStatusPending
}

// OperationResult represents the result of an operation performed on an agent
type OperationResult struct {
	OperationID     string          `json:"operation_id"`     // Unique identifier for this operation
	AgentName       string          `json:"agent_name"`       // Name of the agent
	Operation       OperationType   `json:"operation"`        // Type of operation performed
	Status          OperationStatus `json:"status"`           // Status of the operation
	Message         string          `json:"message"`          // Human-readable message
	Details         string          `json:"details"`          // Detailed information or error message
	Timestamp       time.Time       `json:"timestamp"`        // When the operation was performed
	Duration        time.Duration   `json:"duration"`         // How long the operation took
	PreviousState   AgentState      `json:"previous_state"`   // Agent state before operation
	NewState        AgentState      `json:"new_state"`        // Agent state after operation
	ProcessID       int             `json:"process_id"`       // Process ID (if applicable)
	ExitCode        int             `json:"exit_code"`        // Exit code (if applicable)
	RetryCount      int             `json:"retry_count"`      // Number of retries attempted
	MaxRetries      int             `json:"max_retries"`      // Maximum number of retries allowed
	Metadata        map[string]interface{} `json:"metadata"`  // Additional operation metadata
}

// NewOperationResult creates a new OperationResult
func NewOperationResult(agentName string, operation OperationType) *OperationResult {
	now := time.Now()
	return &OperationResult{
		OperationID:   generateOperationID(),
		AgentName:     agentName,
		Operation:     operation,
		Status:        OperationStatusPending,
		Message:       fmt.Sprintf("%s operation initiated for agent '%s'", operation, agentName),
		Timestamp:     now,
		Duration:      0,
		PreviousState: AgentStateUnknown,
		NewState:      AgentStateUnknown,
		ProcessID:     0,
		ExitCode:      0,
		RetryCount:    0,
		MaxRetries:    0,
		Metadata:      make(map[string]interface{}),
	}
}

// SetSuccess marks the operation as successful
func (or *OperationResult) SetSuccess(message string) {
	or.Status = OperationStatusSuccess
	or.Message = message
	if !or.Timestamp.IsZero() {
		or.Duration = time.Since(or.Timestamp)
	}
}

// SetFailed marks the operation as failed
func (or *OperationResult) SetFailed(message, details string) {
	or.Status = OperationStatusFailed
	or.Message = message
	or.Details = details
	if !or.Timestamp.IsZero() {
		or.Duration = time.Since(or.Timestamp)
	}
}

// SetTimeout marks the operation as timed out
func (or *OperationResult) SetTimeout(message string) {
	or.Status = OperationStatusTimeout
	or.Message = message
	if !or.Timestamp.IsZero() {
		or.Duration = time.Since(or.Timestamp)
	}
}

// SetCancelled marks the operation as cancelled
func (or *OperationResult) SetCancelled(message string) {
	or.Status = OperationStatusCancelled
	or.Message = message
	if !or.Timestamp.IsZero() {
		or.Duration = time.Since(or.Timestamp)
	}
}

// SetStateTransition sets the state transition information
func (or *OperationResult) SetStateTransition(previousState, newState AgentState) {
	or.PreviousState = previousState
	or.NewState = newState
}

// SetProcessInfo sets process-related information
func (or *OperationResult) SetProcessInfo(pid int, exitCode int) {
	or.ProcessID = pid
	or.ExitCode = exitCode
}

// IncrementRetry increments the retry count
func (or *OperationResult) IncrementRetry() {
	or.RetryCount++
}

// AddMetadata adds metadata to the operation
func (or *OperationResult) AddMetadata(key string, value interface{}) {
	or.Metadata[key] = value
}

// GetMetadata retrieves metadata by key
func (or *OperationResult) GetMetadata(key string) (interface{}, bool) {
	value, exists := or.Metadata[key]
	return value, exists
}

// IsSuccessful returns true if the operation was successful
func (or *OperationResult) IsSuccessful() bool {
	return or.Status.IsSuccessful()
}

// IsFailed returns true if the operation failed
func (or *OperationResult) IsFailed() bool {
	return or.Status.IsFailed()
}

// IsPending returns true if the operation is still pending
func (or *OperationResult) IsPending() bool {
	return or.Status.IsPending()
}

// CanRetry returns true if the operation can be retried
func (or *OperationResult) CanRetry() bool {
	return or.RetryCount < or.MaxRetries && or.MaxRetries > 0
}

// Clone creates a deep copy of the OperationResult
func (or *OperationResult) Clone() *OperationResult {
	clone := *or

	// Deep copy metadata
	if or.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range or.Metadata {
			clone.Metadata[k] = v
		}
	}

	return &clone
}

// String returns a string representation of the operation result
func (or *OperationResult) String() string {
	return fmt.Sprintf("OperationResult{ID: %s, Agent: %s, Op: %s, Status: %s, Message: %s}",
		or.OperationID, or.AgentName, or.Operation, or.Status, or.Message)
}

// generateOperationID generates a unique operation ID
func generateOperationID() string {
	return fmt.Sprintf("op_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// BatchOperationResult represents the result of a batch operation on multiple agents
type BatchOperationResult struct {
	OperationID     string            `json:"operation_id"`     // Unique identifier for the batch operation
	Operation       OperationType     `json:"operation"`        // Type of batch operation
	TotalAgents     int               `json:"total_agents"`     // Total number of agents targeted
	SuccessfulOps   int               `json:"successful_ops"`   // Number of successful operations
	FailedOps       int               `json:"failed_ops"`       // Number of failed operations
	SkippedOps      int               `json:"skipped_ops"`      // Number of skipped operations
	Results         []OperationResult `json:"results"`          // Individual operation results
	Timestamp       time.Time         `json:"timestamp"`        // When the batch operation was performed
	Duration        time.Duration     `json:"duration"`         // Total duration of the batch operation
	Parallel        bool              `json:"parallel"`         // Whether operations were run in parallel
	MaxConcurrency  int               `json:"max_concurrency"`  // Maximum concurrent operations
}

// NewBatchOperationResult creates a new BatchOperationResult
func NewBatchOperationResult(operation OperationType, totalAgents int, parallel bool, maxConcurrency int) *BatchOperationResult {
	return &BatchOperationResult{
		OperationID:    generateOperationID(),
		Operation:      operation,
		TotalAgents:    totalAgents,
		SuccessfulOps:  0,
		FailedOps:      0,
		SkippedOps:     0,
		Results:        []OperationResult{},
		Timestamp:      time.Now(),
		Duration:       0,
		Parallel:       parallel,
		MaxConcurrency: maxConcurrency,
	}
}

// AddResult adds an operation result to the batch
func (bor *BatchOperationResult) AddResult(result OperationResult) {
	bor.Results = append(bor.Results, result)

	if result.IsSuccessful() {
		bor.SuccessfulOps++
	} else if result.IsFailed() {
		bor.FailedOps++
	}
}

// AddSkipped increments the skipped operations count
func (bor *BatchOperationResult) AddSkipped() {
	bor.SkippedOps++
}

// Complete marks the batch operation as complete
func (bor *BatchOperationResult) Complete() {
	bor.Duration = time.Since(bor.Timestamp)
}

// GetSuccessRate returns the success rate as a percentage
func (bor *BatchOperationResult) GetSuccessRate() float64 {
	if bor.TotalAgents == 0 {
		return 0
	}
	return float64(bor.SuccessfulOps) / float64(bor.TotalAgents) * 100
}

// IsComplete returns true if all operations have been processed
func (bor *BatchOperationResult) IsComplete() bool {
	return len(bor.Results)+bor.SkippedOps >= bor.TotalAgents
}

// GetFailedResults returns all failed operation results
func (bor *BatchOperationResult) GetFailedResults() []OperationResult {
	var failed []OperationResult
	for _, result := range bor.Results {
		if result.IsFailed() {
			failed = append(failed, result)
		}
	}
	return failed
}

// GetSuccessfulResults returns all successful operation results
func (bor *BatchOperationResult) GetSuccessfulResults() []OperationResult {
	var successful []OperationResult
	for _, result := range bor.Results {
		if result.IsSuccessful() {
			successful = append(successful, result)
		}
	}
	return successful
}

// String returns a string representation of the batch operation result
func (bor *BatchOperationResult) String() string {
	return fmt.Sprintf("BatchOperationResult{ID: %s, Op: %s, Total: %d, Success: %d, Failed: %d, Skipped: %d, Duration: %v}",
		bor.OperationID, bor.Operation, bor.TotalAgents, bor.SuccessfulOps, bor.FailedOps, bor.SkippedOps, bor.Duration)
}