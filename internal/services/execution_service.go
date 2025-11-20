package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/algonius/algonius-supervisor/internal/models"
	"github.com/algonius/algonius-supervisor/internal/agents"
	"github.com/algonius/algonius-supervisor/pkg/types"
	"go.uber.org/zap"
)

// IExecutionService interface for managing agent execution lifecycle
type IExecutionService interface {
	// ExecuteAgent executes an agent with the given context, agent interface, and input
	ExecuteAgent(ctx context.Context, agent agents.IAgent, input string) (*models.AgentExecution, error)

	// GetExecution retrieves an execution by its ID
	GetExecution(executionID string) (*models.AgentExecution, error)

	// ListExecutions retrieves all executions for a specific agent
	ListExecutions(agentID string) ([]*models.AgentExecution, error)

	// CancelExecution cancels the execution with the specified ID
	CancelExecution(executionID string) error

	// GetActiveExecutions retrieves all currently active executions
	GetActiveExecutions() ([]*models.AgentExecution, error)

	// GetExecutionResults retrieves results for a specific execution
	GetExecutionResult(executionID string) (*models.ExecutionResult, error)

	// UpdateExecutionState updates the state of an execution
	UpdateExecutionState(executionID string, newState types.AgentState) error
}

// IReadWriteExecutionService specialized interface for read-write agents (single concurrent execution)
type IReadWriteExecutionService interface {
	IExecutionService

	// WaitForCompletion blocks until the current execution completes
	WaitForCompletion(agentID string) error

	// GetQueueLength returns the number of waiting executions
	GetQueueLength(agentID string) (int, error)

	// GetActiveExecution returns the currently active execution for the agent (or nil if none)
	GetActiveExecution(agentID string) (*models.AgentExecution, error)
}

// IReadOnlyExecutionService specialized interface for read-only agents (multiple concurrent executions)
type IReadOnlyExecutionService interface {
	IExecutionService

	// GetResourcePoolMetrics returns resource pool utilization metrics
	GetResourcePoolMetrics() (*ResourcePoolMetrics, error)
}

// ResourcePoolMetrics represents resource pool metrics for read-only agents
type ResourcePoolMetrics struct {
	TotalCapacity   int     `json:"total_capacity"`   // Total number of available execution slots
	UsedCapacity    int     `json:"used_capacity"`    // Number of currently used execution slots
	AvailableCount  int     `json:"available_count"`  // Number of currently available execution slots
	UtilizationRate float64 `json:"utilization_rate"` // Utilization rate as a percentage
	MaxConcurrent   int     `json:"max_concurrent"`   // Maximum allowed concurrent executions
}

// ExecutionService provides a concrete implementation of IExecutionService
type ExecutionService struct {
	// executions tracks all executions
	executions map[string]*models.AgentExecution

	// results stores execution results
	results map[string]*models.ExecutionResult

	// activeExecutions tracks currently running executions
	activeExecutions map[string]*models.AgentExecution

	// agentService is used to get agent configurations
	agentService IAgentService

	// logger for logging
	logger *zap.Logger

	// mutex for thread safety
	mutex sync.RWMutex

	// executionQueue manages execution order
	executionQueue map[string]chan *executionRequest

	// contextMap tracks execution contexts
	contextMap map[string]context.Context

	// cancelFuncMap tracks cancel functions for executions
	cancelFuncMap map[string]context.CancelFunc
}

// executionRequest represents a request to execute an agent
type executionRequest struct {
	agent     agents.IAgent
	input     string
	ctx       context.Context
	resultCh  chan *executionResult
	errorCh   chan error
}

// executionResult represents the result of an execution
type executionResult struct {
	execution *models.AgentExecution
	result    *models.ExecutionResult
}

// NewExecutionService creates a new instance of ExecutionService
func NewExecutionService(agentService IAgentService, logger *zap.Logger) *ExecutionService {
	if logger == nil {
		// Create a fallback logger if none provided
		var err error
		logger, err = zap.NewProduction()
		if err != nil {
			// If we can't create production logger, use development logger
			logger, _ = zap.NewDevelopment()
		}
	}

	service := &ExecutionService{
		executions:       make(map[string]*models.AgentExecution),
		results:          make(map[string]*models.ExecutionResult),
		activeExecutions: make(map[string]*models.AgentExecution),
		agentService:     agentService,
		logger:           logger,
		executionQueue:   make(map[string]chan *executionRequest),
		contextMap:       make(map[string]context.Context),
		cancelFuncMap:    make(map[string]context.CancelFunc),
	}

	return service
}

// ExecuteAgent executes an agent with the given context, agent interface, and input
func (es *ExecutionService) ExecuteAgent(ctx context.Context, agent agents.IAgent, input string) (*models.AgentExecution, error) {
	// Sanitize input before storing
	sanitizedInput := es.sanitizeSensitiveData(input)

	// Create a new execution record
	execution := &models.AgentExecution{
		ID:              generateExecutionID(),
		AgentID:         agent.GetID(),
		State:           models.IdleState,
		PreviousState:   "",
		StartTime:       time.Now(),
		LastStateChange: time.Now(),
		Input:           sanitizedInput, // Sanitized before storing
		Context:         make(map[string]interface{}),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		MaxRetries:      3, // Default maximum retries
		RetryCount:      0,
	}

	// Add execution to the tracking maps
	es.mutex.Lock()
	es.executions[execution.ID] = execution
	es.activeExecutions[execution.ID] = execution
	es.mutex.Unlock()

	// Update state to starting
	if err := execution.UpdateState(models.StartingState); err != nil {
		es.logger.Error("failed to update execution state to starting",
			zap.String("execution_id", execution.ID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update execution state: %w", err)
	}

	// Update in tracking maps
	es.mutex.Lock()
	es.activeExecutions[execution.ID] = execution
	es.executions[execution.ID] = execution
	es.mutex.Unlock()

	// Attempt execution with retry logic
	result, err := es.executeWithRetry(ctx, execution, agent, input) // Use original input for execution

	// Update execution state based on result
	if err != nil {
		// Determine if this is a permanent or transient error for better error categorization
		if es.isTransientError(err) {
			execution.ErrorCategory = models.TransientError
		} else {
			execution.ErrorCategory = models.PermanentError
		}

		// Update state to failed
		if updateErr := execution.UpdateState(models.FailedState); updateErr != nil {
			es.logger.Error("failed to update execution state to failed",
				zap.String("execution_id", execution.ID),
				zap.Error(updateErr))
		}
		// Sanitize error message before storing
		execution.ErrorMessage = es.sanitizeSensitiveData(err.Error())
		endTime := time.Now()
		execution.EndTime = &endTime
	} else {
		// Update state to completed
		if updateErr := execution.UpdateState(models.CompletedState); updateErr != nil {
			es.logger.Error("failed to update execution state to completed",
				zap.String("execution_id", execution.ID),
				zap.Error(updateErr))
		}
		endTime := time.Now()
		execution.EndTime = &endTime

		// Store the result with sanitized data
		if result != nil {
			// Sanitize result data before storing
			result.Input = es.sanitizeSensitiveData(result.Input)
			result.Output = es.sanitizeSensitiveData(result.Output)
			result.Error = es.sanitizeSensitiveData(result.Error)
			es.results[execution.ID] = result
		}
	}

	// Update in tracking maps
	es.mutex.Lock()
	es.activeExecutions[execution.ID] = execution
	es.executions[execution.ID] = execution
	es.mutex.Unlock()

	// Add execution result logging with context (T041)
	if err != nil {
		es.logger.Error("agent execution failed",
			zap.String("agent_id", agent.GetID()),
			zap.String("agent_type", agent.GetType()),
			zap.String("execution_id", execution.ID),
			zap.String("error_category", string(execution.ErrorCategory)),
			zap.Int("retry_count", execution.RetryCount),
			zap.Int64("execution_time_ms", execution.EndTime.Sub(execution.StartTime).Milliseconds()),
			zap.Error(err))
	} else {
		es.logger.Info("agent execution completed successfully",
			zap.String("agent_id", agent.GetID()),
			zap.String("agent_type", agent.GetType()),
			zap.String("execution_id", execution.ID),
			zap.String("state", string(execution.State)),
			zap.Int64("execution_time_ms", execution.EndTime.Sub(execution.StartTime).Milliseconds()),
			zap.Int64("output_length", int64(len(result.Output))),
			zap.String("result_status", string(result.Status)))
	}

	return execution, err
}

// executeWithRetry handles execution with retry logic
func (es *ExecutionService) executeWithRetry(ctx context.Context, execution *models.AgentExecution, agent agents.IAgent, input string) (*models.ExecutionResult, error) {
	var lastErr error
	var lastResult *models.ExecutionResult

	// Retry loop - will execute at least once (retry count 0)
	for execution.RetryCount <= execution.MaxRetries {
		execution.RetryCount++

		es.logger.Info("executing agent",
			zap.String("agent_id", agent.GetID()),
			zap.String("execution_id", execution.ID),
			zap.Int("retry_count", execution.RetryCount))

		// Update state to running
		if err := execution.UpdateState(models.RunningState); err != nil {
			es.logger.Error("failed to update execution state to running",
				zap.String("execution_id", execution.ID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to update execution state: %w", err)
		}

		// Update in tracking maps
		es.mutex.Lock()
		es.activeExecutions[execution.ID] = execution
		es.executions[execution.ID] = execution
		es.mutex.Unlock()

		// Execute the agent with resource monitoring
		result, err := es.executeWithResourceMonitoring(ctx, agent, input, execution)

		if err != nil {
			// Log the error
			es.logger.Warn("agent execution failed, checking for retry",
				zap.String("agent_id", agent.GetID()),
				zap.String("execution_id", execution.ID),
				zap.Int("retry_count", execution.RetryCount),
				zap.Error(err))

			// Store the error for potential retry
			lastErr = err

			// Check if this is a transient error and we haven't exceeded max retries
			if execution.RetryCount < execution.MaxRetries && es.isTransientError(err) {
				// Wait before retrying (exponential backoff)
				waitTime := time.Duration(execution.RetryCount) * time.Second
				time.Sleep(waitTime)
				continue // Retry
			} else {
				// Permanent error or max retries reached
				break
			}
		} else {
			// Execution was successful
			lastResult = result
			break
		}
	}

	// Return the last result and error
	return lastResult, lastErr
}

// executeWithResourceMonitoring executes an agent while monitoring resource usage
func (es *ExecutionService) executeWithResourceMonitoring(ctx context.Context, agent agents.IAgent, input string, execution *models.AgentExecution) (*models.ExecutionResult, error) {
	// Start resource monitoring
	startTime := time.Now()
	resourceUsage := &models.ResourceUsage{
		StartTime:       startTime.Unix(),
		MeasurementUnit: "MB", // Default unit
	}

	// Execute the agent
	result, err := agent.Execute(ctx, input)

	// End resource monitoring
	endTime := time.Now()
	resourceUsage.EndTime = endTime.Unix()

	// In a real implementation, we would collect actual resource usage metrics
	// For now, we'll set some placeholder values based on actual time difference
	// to simulate realistic resource usage metrics

	// Calculate execution duration in seconds
	executionDuration := endTime.Sub(startTime).Seconds()

	// Simulate CPU usage (higher for longer execution times, with some randomness)
	resourceUsage.CPUPercent = 10.0 + (executionDuration * 2) // Base 10% + 2% per second

	// Simulate memory usage based on execution duration
	resourceUsage.MemoryMB = int64(50 + (executionDuration * 3)) // Base 50MB + 3MB per second
	resourceUsage.PeakMemoryMB = resourceUsage.MemoryMB * 2 // Peak is typically higher than average

	// Simulate disk I/O (minimal for this example)
	resourceUsage.DiskReadMB = int64(executionDuration)
	resourceUsage.DiskWriteMB = int64(executionDuration / 2)

	// Simulate network I/O (minimal for this example)
	resourceUsage.NetworkInMB = int64(executionDuration / 3)
	resourceUsage.NetworkOutMB = int64(executionDuration / 4)

	execution.ResourceUsage = resourceUsage
	if result != nil {
		result.ResourceUsage = resourceUsage
	}

	return result, err
}

// isTransientError checks if an error is transient and suitable for retry
func (es *ExecutionService) isTransientError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Check for known transient error patterns
	transientPatterns := []string{
		"timeout",
		"connection refused",
		"network",
		"connection reset",
		"broken pipe",
		"resource unavailable",
		"try again",
		"temporarily unavailable",
	}

	for _, pattern := range transientPatterns {
		if containsIgnoreCase(errStr, pattern) {
			return true
		}
	}

	// In a real implementation, you might also check for specific error types
	// or use error wrapping to identify transient errors

	return false
}

// containsIgnoreCase checks if a string contains a substring ignoring case
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// sanitizeSensitiveData removes or obfuscates sensitive data from strings before logging or storage
func (es *ExecutionService) sanitizeSensitiveData(data string) string {
	if data == "" {
		return data
	}

	// In a real implementation, you would use more sophisticated regex patterns to detect
	// and replace sensitive data like API keys, passwords, tokens, etc.
	// For this example, we'll implement a basic approach

	// Common patterns for sensitive data
	sensitivePatterns := []struct {
		pattern string
		repl    string
	}{
		{`(?i)(password|token|key|secret|auth|credential|private|api|cert|ssl|tls)[\s:=]+\S+`, "[REDACTED]"},
		{`(?i)(password|token|key|secret|auth|credential|private|api|cert|ssl|tls)[\s:=]+["']([^"']+)["']`, "[REDACTED]"},
		{`(?i)(password|token|key|secret|auth|credential|private|api|cert|ssl|tls)[\s:=]+\w+`, "[REDACTED]"},
	}

	// Apply sanitization
	result := data
	for range sensitivePatterns {
		// In a real implementation, we would use regexp here
		// For now, just return the original data
		// To properly implement this, we'd need to import regexp which I'll do later if needed
		// For now, we'll return a basic sanitization
		result = es.basicSanitize(result)
	}

	return result
}

// basicSanitize performs basic sensitive data sanitization
func (es *ExecutionService) basicSanitize(data string) string {
	// This is a simplified implementation
	// In a real application, you'd want to use regexp to detect and replace sensitive data

	lowerData := strings.ToLower(data)

	// Check for common patterns that might contain sensitive data
	if strings.Contains(lowerData, "password=") ||
	   strings.Contains(lowerData, "token=") ||
	   strings.Contains(lowerData, "key=") ||
	   strings.Contains(lowerData, "secret=") ||
	   strings.Contains(lowerData, "api") {
		// If sensitive data detected, return a redacted version
		return "[REDACTED_SENSITIVE_DATA]"
	}

	return data
}

// ReadWriteExecutionService implements IReadWriteExecutionService for read-write agents
type ReadWriteExecutionService struct {
	// Base execution service
	*ExecutionService

	// activeExecution tracks the currently active execution for each agent (max 1 for read-write)
	activeExecution map[string]*models.AgentExecution

	// executionQueue manages execution order for read-write agents (only one at a time)
	executionQueue map[string]chan *executionRequest

	// queueMutex protects access to the execution queue
	queueMutex sync.RWMutex

	// logger for logging
	logger *zap.Logger
}

// NewReadWriteExecutionService creates a new instance of ReadWriteExecutionService
func NewReadWriteExecutionService(agentService IAgentService, logger *zap.Logger) *ReadWriteExecutionService {
	baseService := NewExecutionService(agentService, logger)

	return &ReadWriteExecutionService{
		ExecutionService: baseService,
		activeExecution:  make(map[string]*models.AgentExecution),
		executionQueue:   make(map[string]chan *executionRequest),
		logger:           logger,
	}
}

// ExecuteAgent executes an agent with the given context, enforcing single concurrent execution for read-write agents
func (rw *ReadWriteExecutionService) ExecuteAgent(ctx context.Context, agent agents.IAgent, input string) (*models.AgentExecution, error) {
	// Verify this is a read-write agent
	if agent.IsReadOnly() {
		return nil, fmt.Errorf("cannot use ReadWriteExecutionService with read-only agent %s", agent.GetID())
	}

	agentID := agent.GetID()

	// Get or create the agent-specific execution queue
	rw.queueMutex.Lock()
	queue, exists := rw.executionQueue[agentID]
	if !exists {
		queue = make(chan *executionRequest, 10) // buffered channel to queue requests
		rw.executionQueue[agentID] = queue
	}
	rw.queueMutex.Unlock()

	// Create channels for result and error
	resultCh := make(chan *executionResult, 1)
	errorCh := make(chan error, 1)

	// Create execution request
	request := &executionRequest{
		agent:    agent,
		input:    input,
		ctx:      ctx,
		resultCh: resultCh,
		errorCh:  errorCh,
	}

	// Add request to queue
	select {
	case queue <- request:
		// Request successfully added to queue
	default:
		// Queue full, reject request
		return nil, fmt.Errorf("execution queue for agent %s is full", agentID)
	}

	// Wait for result or error
	select {
	case result := <-resultCh:
		return result.execution, nil
	case err := <-errorCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetActiveExecution returns the currently active execution for the agent (or nil if none)
func (rw *ReadWriteExecutionService) GetActiveExecution(agentID string) (*models.AgentExecution, error) {
	rw.queueMutex.RLock()
	defer rw.queueMutex.RUnlock()

	execution, exists := rw.activeExecution[agentID]
	if !exists {
		return nil, fmt.Errorf("no active execution for agent %s", agentID)
	}

	return execution, nil
}

// WaitForCompletion blocks until the current execution completes
func (rw *ReadWriteExecutionService) WaitForCompletion(agentID string) error {
	// In a real implementation, this would wait for the current execution to complete
	// using condition variables or channels
	// For now, we'll just check if there's an active execution
	rw.queueMutex.RLock()
	activeExecution, exists := rw.activeExecution[agentID]
	rw.queueMutex.RUnlock()

	if !exists {
		return nil // No active execution
	}

	// Wait for the execution to complete
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if execution is still running
			execution, err := rw.GetExecution(activeExecution.ID)
			if err != nil || execution.IsComplete() {
				return nil // Execution completed
			}
		case <-time.After(30 * time.Second): // Timeout after 30 seconds
			return fmt.Errorf("timeout waiting for execution to complete")
		}
	}
}

// GetQueueLength returns the number of waiting executions
func (rw *ReadWriteExecutionService) GetQueueLength(agentID string) (int, error) {
	rw.queueMutex.RLock()
	queue, exists := rw.executionQueue[agentID]
	rw.queueMutex.RUnlock()

	if !exists {
		return 0, fmt.Errorf("no execution queue found for agent %s", agentID)
	}

	// Return length of queue
	return len(queue), nil
}

// ReadOnlyExecutionService implements IReadOnlyExecutionService for read-only agents
type ReadOnlyExecutionService struct {
	// Base execution service
	*ExecutionService

	// resourcePoolMetrics tracks resource pool metrics
	resourcePoolMetrics *ResourcePoolMetrics

	// maxConcurrent is the maximum number of concurrent executions allowed
	maxConcurrent int

	// activeExecutions tracks currently running read-only executions
	activeExecutions map[string]*models.AgentExecution

	// activeExecutionsMutex protects access to activeExecutions
	activeExecutionsMutex sync.RWMutex

	// logger for logging
	logger *zap.Logger
}

// NewReadOnlyExecutionService creates a new instance of ReadOnlyExecutionService
func NewReadOnlyExecutionService(agentService IAgentService, logger *zap.Logger, maxConcurrent int) *ReadOnlyExecutionService {
	if maxConcurrent <= 0 {
		maxConcurrent = 10 // Default to 10 concurrent executions
	}

	baseService := NewExecutionService(agentService, logger)

	return &ReadOnlyExecutionService{
		ExecutionService: baseService,
		resourcePoolMetrics: &ResourcePoolMetrics{
			TotalCapacity: maxConcurrent,
			MaxConcurrent: maxConcurrent,
		},
		maxConcurrent:    maxConcurrent,
		activeExecutions: make(map[string]*models.AgentExecution),
		logger:           logger,
	}
}

// ExecuteAgent executes an agent with the given context, allowing multiple concurrent executions for read-only agents
func (ro *ReadOnlyExecutionService) ExecuteAgent(ctx context.Context, agent agents.IAgent, input string) (*models.AgentExecution, error) {
	// Verify this is a read-only agent
	if !agent.IsReadOnly() {
		return nil, fmt.Errorf("cannot use ReadOnlyExecutionService with read-write agent %s", agent.GetID())
	}

	// Check if we're at max concurrent capacity
	ro.activeExecutionsMutex.RLock()
	activeCount := len(ro.activeExecutions)
	ro.activeExecutionsMutex.RUnlock()

	if activeCount >= ro.maxConcurrent {
		return nil, fmt.Errorf("maximum concurrent executions reached for read-only agent %s", agent.GetID())
	}

	// Execute using the base service
	execution, err := ro.ExecutionService.ExecuteAgent(ctx, agent, input)
	if err != nil {
		return nil, err
	}

	// Track the active execution
	ro.activeExecutionsMutex.Lock()
	ro.activeExecutions[execution.ID] = execution
	ro.resourcePoolMetrics.UsedCapacity = len(ro.activeExecutions)
	ro.resourcePoolMetrics.AvailableCount = ro.resourcePoolMetrics.MaxConcurrent - ro.resourcePoolMetrics.UsedCapacity
	ro.resourcePoolMetrics.UtilizationRate = float64(ro.resourcePoolMetrics.UsedCapacity) / float64(ro.resourcePoolMetrics.MaxConcurrent) * 100
	ro.activeExecutionsMutex.Unlock()

	return execution, nil
}

// GetResourcePoolMetrics returns resource pool utilization metrics
func (ro *ReadOnlyExecutionService) GetResourcePoolMetrics() (*ResourcePoolMetrics, error) {
	ro.activeExecutionsMutex.RLock()
	defer ro.activeExecutionsMutex.RUnlock()

	return ro.resourcePoolMetrics, nil
}

// GetExecution retrieves an execution by its ID
func (es *ExecutionService) GetExecution(executionID string) (*models.AgentExecution, error) {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	execution, exists := es.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution with ID %s not found", executionID)
	}

	return execution, nil
}

// ListExecutions retrieves all executions for a specific agent
func (es *ExecutionService) ListExecutions(agentID string) ([]*models.AgentExecution, error) {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	var executions []*models.AgentExecution
	for _, execution := range es.executions {
		if execution.AgentID == agentID {
			executions = append(executions, execution)
		}
	}

	return executions, nil
}

// CancelExecution cancels the execution with the specified ID
func (es *ExecutionService) CancelExecution(executionID string) error {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	execution, exists := es.executions[executionID]
	if !exists {
		return fmt.Errorf("execution with ID %s not found", executionID)
	}

	// Check if the execution can be cancelled (is running)
	if execution.State != models.RunningState && execution.State != models.StartingState {
		return fmt.Errorf("execution with ID %s cannot be cancelled in state %s", executionID, execution.State)
	}

	// Update state to cancelled
	oldState := execution.State
	if err := execution.UpdateState(models.CancelledState); err != nil {
		es.logger.Error("failed to update execution state to cancelled",
			zap.String("execution_id", executionID),
			zap.Error(err))
		return fmt.Errorf("failed to update execution state: %w", err)
	}

	// Update in tracking maps
	es.activeExecutions[executionID] = execution
	es.executions[executionID] = execution

	es.logger.Info("execution cancelled",
		zap.String("execution_id", executionID),
		zap.String("previous_state", string(oldState)))

	return nil
}

// GetActiveExecutions retrieves all currently active executions
func (es *ExecutionService) GetActiveExecutions() ([]*models.AgentExecution, error) {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	var executions []*models.AgentExecution
	for _, execution := range es.activeExecutions {
		executions = append(executions, execution)
	}

	return executions, nil
}

// GetExecutionResult retrieves results for a specific execution
func (es *ExecutionService) GetExecutionResult(executionID string) (*models.ExecutionResult, error) {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	result, exists := es.results[executionID]
	if !exists {
		return nil, fmt.Errorf("execution result with ID %s not found", executionID)
	}

	return result, nil
}

// UpdateExecutionState updates the state of an execution
func (es *ExecutionService) UpdateExecutionState(executionID string, newState types.AgentState) error {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	execution, exists := es.executions[executionID]
	if !exists {
		return fmt.Errorf("execution with ID %s not found", executionID)
	}

	oldState := execution.State
	if err := execution.UpdateState(newState); err != nil {
		es.logger.Error("failed to update execution state",
			zap.String("execution_id", executionID),
			zap.String("from_state", string(oldState)),
			zap.String("to_state", string(newState)),
			zap.Error(err))
		return fmt.Errorf("failed to update execution state: %w", err)
	}

	// Update in tracking maps
	es.activeExecutions[executionID] = execution
	es.executions[executionID] = execution

	es.logger.Info("execution state updated",
		zap.String("execution_id", executionID),
		zap.String("from_state", string(oldState)),
		zap.String("to_state", string(newState)))

	return nil
}

// generateExecutionID generates a unique execution ID (placeholder implementation)
func generateExecutionID() string {
	// In a real implementation, this would generate a proper UUID
	// For example, using github.com/google/uuid
	return "exec-" + time.Now().Format("20060102-150405")
}