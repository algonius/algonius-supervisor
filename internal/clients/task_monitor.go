package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/algonius/algonius-supervisor/pkg/a2a"
)

// TaskMonitorClient provides methods for monitoring tasks
type TaskMonitorClient struct {
	a2aClient         *A2AClient
	pollingInterval   time.Duration
	maxPollingRetries int
}

// TaskMonitorCallback is a function type for task monitoring callbacks
type TaskMonitorCallback func(task *a2a.A2ATask, err error)

// NewTaskMonitorClient creates a new task monitor client instance
func NewTaskMonitorClient(a2aClient *A2AClient) *TaskMonitorClient {
	return &TaskMonitorClient{
		a2aClient:         a2aClient,
		pollingInterval:   5 * time.Second, // Default polling interval
		maxPollingRetries: 3,               // Default max retries for polling failures
	}
}

// GetTask retrieves the status and details of a specific task
func (tmc *TaskMonitorClient) GetTask(ctx context.Context, agentID, taskID string) (*a2a.A2ATask, error) {
	task, err := tmc.a2aClient.GetTask(ctx, agentID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task %s for agent %s: %w", taskID, agentID, err)
	}
	
	return task, nil
}

// WaitForTaskCompletion waits for a task to reach a terminal state (completed, failed, cancelled)
func (tmc *TaskMonitorClient) WaitForTaskCompletion(ctx context.Context, agentID, taskID string) (*a2a.A2ATask, error) {
	ticker := time.NewTicker(tmc.pollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			task, err := tmc.GetTask(ctx, agentID, taskID)
			if err != nil {
				// Check if it's a temporary error that allows for retries
				if tmc.isTransientError(err) {
					continue // Retry on transient errors
				}
				return nil, err
			}
			
			// Check if task is in a terminal state
			if isTerminalState(task.Status) {
				return task, nil
			}
		}
	}
}

// isTransientError checks if an error is transient and allows for retry
func (tmc *TaskMonitorClient) isTransientError(err error) bool {
	// In a real implementation, we'd check the specific error types
	// For now, we'll implement a basic check
	if a2aErr, ok := a2a.AsA2AError(err); ok {
		// These error codes might indicate temporary issues
		return a2aErr.Code == a2a.InternalError
	}
	return false
}

// isTerminalState checks if a task status is a terminal state
func isTerminalState(status string) bool {
	return status == "succeeded" || status == "failed" || status == "cancelled" || status == "expired"
}

// MonitorTaskAsync monitors a task asynchronously and calls the callback when the task completes
func (tmc *TaskMonitorClient) MonitorTaskAsync(ctx context.Context, agentID, taskID string, callback TaskMonitorCallback) {
	go func() {
		task, err := tmc.WaitForTaskCompletion(ctx, agentID, taskID)
		callback(task, err)
	}()
}

// ListTasks lists tasks with optional filters for a specific agent
func (tmc *TaskMonitorClient) ListTasks(ctx context.Context, agentID string, filters map[string]string) ([]*a2a.A2ATask, error) {
	tasks, err := tmc.a2aClient.ListTasks(ctx, agentID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks for agent %s: %w", agentID, err)
	}
	
	return tasks, nil
}

// CancelTask cancels a running task
func (tmc *TaskMonitorClient) CancelTask(ctx context.Context, agentID, taskID string) (*a2a.A2ATask, error) {
	task, err := tmc.a2aClient.CancelTask(ctx, agentID, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel task %s for agent %s: %w", taskID, agentID, err)
	}
	
	return task, nil
}

// WaitForTaskWithTimeout waits for a task to complete or timeout
func (tmc *TaskMonitorClient) WaitForTaskWithTimeout(ctx context.Context, agentID, taskID string, timeout time.Duration) (*a2a.A2ATask, error) {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	return tmc.WaitForTaskCompletion(timeoutCtx, agentID, taskID)
}

// MonitorAgentTasks monitors all tasks for a specific agent
func (tmc *TaskMonitorClient) MonitorAgentTasks(ctx context.Context, agentID string, statusFilter string) ([]*a2a.A2ATask, error) {
	filters := make(map[string]string)
	if statusFilter != "" {
		filters["status"] = statusFilter
	}
	
	tasks, err := tmc.ListTasks(ctx, agentID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to monitor tasks for agent %s: %w", agentID, err)
	}
	
	return tasks, nil
}

// GetTaskExecutionTime returns the execution time of a completed task
func (tmc *TaskMonitorClient) GetTaskExecutionTime(task *a2a.A2ATask) (time.Duration, error) {
	if task == nil {
		return 0, fmt.Errorf("task is nil")
	}
	
	if task.CreatedAt.IsZero() || task.ModifiedAt.IsZero() {
		return 0, fmt.Errorf("task timestamps are not set")
	}
	
	return task.ModifiedAt.Sub(task.CreatedAt), nil
}

// IsTaskCompleted returns whether a task has reached a completed state
func (tmc *TaskMonitorClient) IsTaskCompleted(task *a2a.A2ATask) bool {
	if task == nil {
		return false
	}
	
	return task.Status == "succeeded" || task.Status == "failed" || task.Status == "cancelled" || task.Status == "expired"
}

// GetTaskResult retrieves the result of a completed task
func (tmc *TaskMonitorClient) GetTaskResult(task *a2a.A2ATask) (interface{}, error) {
	if task == nil {
		return nil, fmt.Errorf("task is nil")
	}
	
	if !tmc.IsTaskCompleted(task) {
		return nil, fmt.Errorf("task is not completed, current status: %s", task.Status)
	}
	
	// In a real implementation, the result would be extracted from the task
	// For now, we'll return the entire task as the result
	return task, nil
}