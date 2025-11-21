package models

import (
	"sort"
	"time"

	"github.com/algonius/algonius-supervisor/pkg/types"
)

// ExecutionHistory represents the history of task executions
type ExecutionHistory struct {
	ID               string                    `json:"id" yaml:"id"`
	TaskID           string                    `json:"task_id" yaml:"task_id"`
	ExecutionID      string                    `json:"execution_id" yaml:"execution_id"`
	StartTime        time.Time                 `json:"start_time" yaml:"start_time"`
	EndTime          time.Time                 `json:"end_time" yaml:"end_time"`
	Status           types.ExecutionStatus     `json:"status" yaml:"status"`
	Input            string                    `json:"input" yaml:"input"`
	Output           string                    `json:"output" yaml:"output"`
	Error            string                    `json:"error,omitempty" yaml:"error,omitempty"`
	ExecutionTimeMs  int64                     `json:"execution_time_ms" yaml:"execution_time_ms"`
	RetryCount       int                       `json:"retry_count" yaml:"retry_count"`
	TriggerType      types.TaskTriggerType     `json:"trigger_type" yaml:"trigger_type"` // scheduled, manual, api, event
	CreatedAt        time.Time                 `json:"created_at" yaml:"created_at"`
}

// TaskExecutionResult represents the result of a task execution
type TaskExecutionResult struct {
	ID               string                    `json:"id" yaml:"id"`
	TaskID           string                    `json:"task_id" yaml:"task_id"`
	ExecutionID      string                    `json:"execution_id" yaml:"execution_id"`
	StartTime        time.Time                 `json:"start_time" yaml:"start_time"`
	EndTime          time.Time                 `json:"end_time" yaml:"end_time"`
	Status           types.ExecutionStatus     `json:"status" yaml:"status"`
	Input            string                    `json:"input" yaml:"input"`
	Output           string                    `json:"output" yaml:"output"`
	Error            string                    `json:"error,omitempty" yaml:"error,omitempty"`
	ExecutionTimeMs  int64                     `json:"execution_time_ms" yaml:"execution_time_ms"`
	AgentID          string                    `json:"agent_id" yaml:"agent_id"`
	TriggerType      types.TaskTriggerType     `json:"trigger_type" yaml:"trigger_type"`
	CreatedAt        time.Time                 `json:"created_at" yaml:"created_at"`
}

// Validate ExecutionHistory
func (h *ExecutionHistory) Validate() error {
	if h.ID == "" {
		return ValidationError("execution history ID cannot be empty")
	}
	if h.TaskID == "" {
		return ValidationError("execution history task ID cannot be empty")
	}
	if h.ExecutionID == "" {
		return ValidationError("execution history execution ID cannot be empty")
	}
	if h.StartTime.IsZero() {
		return ValidationError("execution history start time cannot be zero")
	}
	return nil
}

// Validate TaskExecutionResult
func (r *TaskExecutionResult) Validate() error {
	if r.ID == "" {
		return ValidationError("task execution result ID cannot be empty")
	}
	if r.TaskID == "" {
		return ValidationError("task execution result task ID cannot be empty")
	}
	if r.ExecutionID == "" {
		return ValidationError("task execution result execution ID cannot be empty")
	}
	if r.StartTime.IsZero() {
		return ValidationError("task execution result start time cannot be zero")
	}
	return nil
}

// ExecutionHistoryRepository interface for storing and retrieving execution history
type ExecutionHistoryRepository interface {
	// StoreExecutionHistory stores an execution history record
	StoreExecutionHistory(history *ExecutionHistory) error
	
	// GetExecutionHistory retrieves execution history for a specific task
	GetExecutionHistory(taskID string, limit int) ([]*ExecutionHistory, error)
	
	// GetExecutionHistoryByTaskAndStatus retrieves execution history for a task with a specific status
	GetExecutionHistoryByTaskAndStatus(taskID string, status types.ExecutionStatus, limit int) ([]*ExecutionHistory, error)
	
	// GetExecutionHistoryByTimeRange retrieves execution history within a time range
	GetExecutionHistoryByTimeRange(start, end time.Time, limit int) ([]*ExecutionHistory, error)
	
	// GetLatestExecutionHistory retrieves the most recent execution history for a task
	GetLatestExecutionHistory(taskID string) (*ExecutionHistory, error)
	
	// DeleteExecutionHistory deletes execution history records for a task
	DeleteExecutionHistory(taskID string) error
}

// InMemoryExecutionHistoryRepository is an in-memory implementation of ExecutionHistoryRepository
type InMemoryExecutionHistoryRepository struct {
	histories map[string][]*ExecutionHistory
}

// NewInMemoryExecutionHistoryRepository creates a new in-memory history repository
func NewInMemoryExecutionHistoryRepository() *InMemoryExecutionHistoryRepository {
	return &InMemoryExecutionHistoryRepository{
		histories: make(map[string][]*ExecutionHistory),
	}
}

// StoreExecutionHistory stores an execution history record
func (r *InMemoryExecutionHistoryRepository) StoreExecutionHistory(history *ExecutionHistory) error {
	if err := history.Validate(); err != nil {
		return err
	}
	
	r.histories[history.TaskID] = append(r.histories[history.TaskID], history)
	return nil
}

// GetExecutionHistory retrieves execution history for a specific task
func (r *InMemoryExecutionHistoryRepository) GetExecutionHistory(taskID string, limit int) ([]*ExecutionHistory, error) {
	histories, exists := r.histories[taskID]
	if !exists {
		return []*ExecutionHistory{}, nil
	}
	
	// If limit is 0 or greater than the available histories, return all
	if limit <= 0 || limit > len(histories) {
		limit = len(histories)
	}
	
	// Return the most recent histories up to the limit
	startIdx := len(histories) - limit
	if startIdx < 0 {
		startIdx = 0
	}
	
	result := make([]*ExecutionHistory, limit)
	copy(result, histories[startIdx:])
	
	return result, nil
}

// GetExecutionHistoryByTaskAndStatus retrieves execution history for a task with a specific status
func (r *InMemoryExecutionHistoryRepository) GetExecutionHistoryByTaskAndStatus(taskID string, status types.ExecutionStatus, limit int) ([]*ExecutionHistory, error) {
	histories, exists := r.histories[taskID]
	if !exists {
		return []*ExecutionHistory{}, nil
	}
	
	var filteredHistories []*ExecutionHistory
	for _, history := range histories {
		if history.Status == status {
			filteredHistories = append(filteredHistories, history)
		}
	}
	
	// If limit is 0 or greater than the available filtered histories, return all
	if limit <= 0 || limit > len(filteredHistories) {
		limit = len(filteredHistories)
	}
	
	// Return the most recent histories up to the limit
	startIdx := len(filteredHistories) - limit
	if startIdx < 0 {
		startIdx = 0
	}
	
	result := make([]*ExecutionHistory, limit)
	copy(result, filteredHistories[startIdx:])
	
	return result, nil
}

// GetExecutionHistoryByTimeRange retrieves execution history within a time range
func (r *InMemoryExecutionHistoryRepository) GetExecutionHistoryByTimeRange(start, end time.Time, limit int) ([]*ExecutionHistory, error) {
	var filteredHistories []*ExecutionHistory
	
	for _, histories := range r.histories {
		for _, history := range histories {
			if history.StartTime.After(start) && history.StartTime.Before(end) {
				filteredHistories = append(filteredHistories, history)
			}
		}
	}
	
	// Sort by start time (newest first)
	sort.Slice(filteredHistories, func(i, j int) bool {
		return filteredHistories[i].StartTime.After(filteredHistories[j].StartTime)
	})
	
	// If limit is 0 or greater than the available filtered histories, return all
	if limit <= 0 || limit > len(filteredHistories) {
		limit = len(filteredHistories)
	}
	
	result := filteredHistories
	if len(filteredHistories) > limit {
		result = filteredHistories[:limit]
	}
	
	return result, nil
}

// GetLatestExecutionHistory retrieves the most recent execution history for a task
func (r *InMemoryExecutionHistoryRepository) GetLatestExecutionHistory(taskID string) (*ExecutionHistory, error) {
	histories, exists := r.histories[taskID]
	if !exists || len(histories) == 0 {
		return nil, nil
	}
	
	// Find the most recent history
	latest := histories[0]
	for _, history := range histories {
		if history.StartTime.After(latest.StartTime) {
			latest = history
		}
	}
	
	return latest, nil
}

// DeleteExecutionHistory deletes execution history records for a task
func (r *InMemoryExecutionHistoryRepository) DeleteExecutionHistory(taskID string) error {
	delete(r.histories, taskID)
	return nil
}

// Additional utility functions for working with execution history

// ExecutionStats provides statistics about task execution
type ExecutionStats struct {
	TaskID          string          `json:"task_id"`
	TotalExecutions int             `json:"total_executions"`
	Successful      int             `json:"successful"`
	Failed          int             `json:"failed"`
	AverageTimeMs   float64         `json:"average_time_ms"`
	LastExecution   *ExecutionHistory `json:"last_execution"`
}

// CalculateExecutionStats calculates execution statistics for a task
func CalculateExecutionStats(histories []*ExecutionHistory) *ExecutionStats {
	if len(histories) == 0 {
		return &ExecutionStats{}
	}

	stats := &ExecutionStats{
		TaskID:          histories[0].TaskID,
		TotalExecutions: len(histories),
		LastExecution:   histories[0], // Will be updated to latest
	}

	var totalTime int64
	var latest *ExecutionHistory

	for _, history := range histories {
		if history.Status == types.SuccessStatus {
			stats.Successful++
		} else {
			stats.Failed++
		}

		totalTime += history.ExecutionTimeMs

		// Find the most recent execution
		if latest == nil || history.StartTime.After(latest.StartTime) {
			latest = history
		}
	}

	stats.LastExecution = latest

	if stats.TotalExecutions > 0 {
		stats.AverageTimeMs = float64(totalTime) / float64(stats.TotalExecutions)
	}

	return stats
}