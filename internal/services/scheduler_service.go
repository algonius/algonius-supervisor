package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/algonius/algonius-supervisor/internal/models"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// ISchedulerService interface for managing scheduled tasks
type ISchedulerService interface {
	// ScheduleTask schedules a new task based on the provided configuration
	ScheduleTask(task *models.ScheduledTask) error

	// UnscheduleTask removes a scheduled task by its ID
	UnscheduleTask(taskID string) error

	// ListScheduledTasks returns all currently scheduled tasks
	ListScheduledTasks() ([]*models.ScheduledTask, error)

	// ExecuteTask immediately executes a task regardless of its schedule
	ExecuteTask(taskID string) (*models.ExecutionResult, error)

	// PauseTask pauses a scheduled task (prevents it from executing according to schedule)
	PauseTask(taskID string) error

	// ResumeTask resumes a paused task
	ResumeTask(taskID string) error

	// UpdateTask updates an existing task configuration
	UpdateTask(task *models.ScheduledTask) error

	// GetTask returns a specific task by its ID
	GetTask(taskID string) (*models.ScheduledTask, error)
}

// TaskState represents the state of a scheduled task
type TaskState string

const (
	TaskCreated  TaskState = "created"
	TaskActive   TaskState = "active"
	TaskPaused   TaskState = "paused"
	TaskExecuting TaskState = "executing"
)

// SchedulerService implements ISchedulerService interface
type SchedulerService struct {
	// Internal cron scheduler
	cronScheduler *cron.Cron

	// Map of scheduled tasks by ID
	tasks map[string]*models.ScheduledTask

	// Map of cron entry IDs for tracking
	entryIDs map[string]cron.EntryID

	// Agent service for executing agents
	agentService IAgentService

	// Execution service for executing agents
	executionService IExecutionService

	// Logger for logging
	logger *zap.Logger

	// Mutex for thread safety
	mutex sync.RWMutex

	// Context for cancellation
	ctx context.Context
	cancel context.CancelFunc
}

// NewSchedulerService creates a new instance of SchedulerService
func NewSchedulerService(agentService IAgentService, executionService IExecutionService, logger *zap.Logger) *SchedulerService {
	if logger == nil {
		// Create a fallback logger if none provided
		var err error
		logger, err = zap.NewProduction()
		if err != nil {
			// If we can't create production logger, use development logger
			logger, _ = zap.NewDevelopment()
		}
	}

	// Create a context for the scheduler
	ctx, cancel := context.WithCancel(context.Background())

	// Create the cron scheduler with seconds support
	cronScheduler := cron.New(
		cron.WithSeconds(),          // Include seconds in cron expressions
		cron.WithLogger(&CronLogger{logger: logger}), // Use our custom logger
	)

	service := &SchedulerService{
		cronScheduler:  cronScheduler,
		tasks:          make(map[string]*models.ScheduledTask),
		entryIDs:       make(map[string]cron.EntryID),
		agentService:   agentService,
		executionService: executionService,
		logger:         logger,
		ctx:            ctx,
		cancel:         cancel,
	}

	// Start the cron scheduler
	service.cronScheduler.Start()

	return service
}

// ScheduleTask schedules a new task based on the provided configuration
func (ss *SchedulerService) ScheduleTask(task *models.ScheduledTask) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	// Validate the task
	if err := ss.validateTask(task); err != nil {
		ss.logger.Error("invalid task configuration", zap.Error(err))
		return fmt.Errorf("invalid task configuration: %w", err)
	}

	// Check if task with this ID already exists
	if _, exists := ss.tasks[task.ID]; exists {
		return fmt.Errorf("task with ID %s already exists", task.ID)
	}

	// Set the initial state to active
	task.Active = true
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	// Parse the cron expression to validate it
	_, err := cron.ParseStandard(task.CronExpression)
	if err != nil {
		return fmt.Errorf("invalid cron expression '%s': %w", task.CronExpression, err)
	}

	// Schedule the task with the cron scheduler
	entryID, err := ss.cronScheduler.AddFunc(task.CronExpression, func() {
		ss.executeScheduledTask(task)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule task: %w", err)
	}

	// Store the task and its entry ID
	ss.tasks[task.ID] = task
	ss.entryIDs[task.ID] = entryID

	ss.logger.Info("task scheduled successfully",
		zap.String("task_id", task.ID),
		zap.String("agent_id", task.AgentID),
		zap.String("cron_expression", task.CronExpression))

	return nil
}

// UnscheduleTask removes a scheduled task by its ID
func (ss *SchedulerService) UnscheduleTask(taskID string) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	// Check if task exists
	task, exists := ss.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	// Remove from cron scheduler
	if entryID, found := ss.entryIDs[taskID]; found {
		ss.cronScheduler.Remove(entryID)
		delete(ss.entryIDs, taskID)
	}

	// Remove from internal map
	delete(ss.tasks, taskID)

	ss.logger.Info("task unscheduled successfully",
		zap.String("task_id", taskID),
		zap.String("agent_id", task.AgentID))

	return nil
}

// ListScheduledTasks returns all currently scheduled tasks
func (ss *SchedulerService) ListScheduledTasks() ([]*models.ScheduledTask, error) {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	tasks := make([]*models.ScheduledTask, 0, len(ss.tasks))
	for _, task := range ss.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// ExecuteTask immediately executes a task regardless of its schedule
func (ss *SchedulerService) ExecuteTask(taskID string) (*models.ExecutionResult, error) {
	ss.mutex.RLock()
	task, exists := ss.tasks[taskID]
	ss.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}

	// Get the agent configuration
	agentConfig, err := ss.agentService.GetAgent(task.AgentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Create an agent instance
	agent := &ScheduledAgent{
		config: agentConfig,
		logger: ss.logger,
	}

	// Execute the agent with the task's input parameters
	input := ss.buildInputFromParameters(task.InputParameters)
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(agentConfig.Timeout)*time.Second)
	defer cancel()

	execution, err := ss.executionService.ExecuteAgent(ctx, agent, input)
	if err != nil {
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	// Create an execution result based on the execution
	result := &models.ExecutionResult{
		ID:        execution.ID,
		AgentID:   task.AgentID,
		TaskID:    task.ID,
		StartTime: execution.StartTime,
		EndTime:   *execution.EndTime,
		Status:    models.SuccessStatus, // This should be set based on the execution status
		Input:     input,
		Output:    "Execution completed successfully", // In a real implementation, this would come from the result
		ExecutionTime: execution.EndTime.Sub(execution.StartTime).Milliseconds(),
	}

	// Log the task execution
	ss.logger.Info("scheduled task executed",
		zap.String("task_id", taskID),
		zap.String("agent_id", task.AgentID),
		zap.String("execution_id", execution.ID))

	return result, nil
}

// PauseTask pauses a scheduled task (prevents it from executing according to schedule)
func (ss *SchedulerService) PauseTask(taskID string) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	task, exists := ss.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	// If already paused, return error
	if !task.Active {
		return fmt.Errorf("task with ID %s is already paused", taskID)
	}

	// Remove from cron scheduler
	if entryID, found := ss.entryIDs[taskID]; found {
		ss.cronScheduler.Remove(entryID)
		delete(ss.entryIDs, taskID)
	}

	// Update task state
	task.Active = false
	task.UpdatedAt = time.Now()

	ss.logger.Info("task paused",
		zap.String("task_id", taskID),
		zap.String("agent_id", task.AgentID))

	return nil
}

// ResumeTask resumes a paused task
func (ss *SchedulerService) ResumeTask(taskID string) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	task, exists := ss.tasks[taskID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", taskID)
	}

	// If not paused, return error
	if task.Active {
		return fmt.Errorf("task with ID %s is not paused", taskID)
	}

	// Schedule the task again with the cron scheduler
	entryID, err := ss.cronScheduler.AddFunc(task.CronExpression, func() {
		ss.executeScheduledTask(task)
	})
	if err != nil {
		return fmt.Errorf("failed to resume task: %w", err)
	}

	// Store the entry ID
	ss.entryIDs[taskID] = entryID

	// Update task state
	task.Active = true
	task.UpdatedAt = time.Now()

	ss.logger.Info("task resumed",
		zap.String("task_id", taskID),
		zap.String("agent_id", task.AgentID))

	return nil
}

// UpdateTask updates an existing task configuration
func (ss *SchedulerService) UpdateTask(task *models.ScheduledTask) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	// Check if task exists
	existingTask, exists := ss.tasks[task.ID]
	if !exists {
		return fmt.Errorf("task with ID %s not found", task.ID)
	}

	// Validate the updated task
	if err := ss.validateTask(task); err != nil {
		ss.logger.Error("invalid updated task configuration", zap.Error(err))
		return fmt.Errorf("invalid task configuration: %w", err)
	}

	// If the cron expression changed, we need to reschedule
	if existingTask.CronExpression != task.CronExpression {
		// Remove the old schedule
		if entryID, found := ss.entryIDs[task.ID]; found {
			ss.cronScheduler.Remove(entryID)
			delete(ss.entryIDs, task.ID)
		}

		// Add the new schedule if the task is active
		if task.Active {
			entryID, err := ss.cronScheduler.AddFunc(task.CronExpression, func() {
				ss.executeScheduledTask(task)
			})
			if err != nil {
				return fmt.Errorf("failed to reschedule task: %w", err)
			}
			ss.entryIDs[task.ID] = entryID
		}
	}

	// Update the task
	task.UpdatedAt = time.Now()
	ss.tasks[task.ID] = task

	ss.logger.Info("task updated",
		zap.String("task_id", task.ID),
		zap.String("agent_id", task.AgentID))

	return nil
}

// GetTask returns a specific task by its ID
func (ss *SchedulerService) GetTask(taskID string) (*models.ScheduledTask, error) {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	task, exists := ss.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task with ID %s not found", taskID)
	}

	return task, nil
}

// validateTask validates a task before scheduling
func (ss *SchedulerService) validateTask(task *models.ScheduledTask) error {
	if task.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	if task.AgentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	if task.CronExpression == "" {
		return fmt.Errorf("cron expression cannot be empty")
	}

	// Validate cron expression
	_, err := cron.ParseStandard(task.CronExpression)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Check if agent exists
	_, err = ss.agentService.GetAgent(task.AgentID)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	return nil
}

// executeScheduledTask is called by the cron scheduler to execute a scheduled task
func (ss *SchedulerService) executeScheduledTask(task *models.ScheduledTask) {
	ss.logger.Info("executing scheduled task",
		zap.String("task_id", task.ID),
		zap.String("agent_id", task.AgentID))

	// Get the agent configuration
	agentConfig, err := ss.agentService.GetAgent(task.AgentID)
	if err != nil {
		ss.logger.Error("agent not found for scheduled task",
			zap.String("task_id", task.ID),
			zap.String("agent_id", task.AgentID),
			zap.Error(err))
		return
	}

	// Create an agent instance
	agent := &ScheduledAgent{
		config: agentConfig,
		logger: ss.logger,
	}

	// Execute the agent with the task's input parameters
	input := ss.buildInputFromParameters(task.InputParameters)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(agentConfig.Timeout)*time.Second)
	defer cancel()

	execution, err := ss.executionService.ExecuteAgent(ctx, agent, input)
	if err != nil {
		ss.logger.Error("scheduled task execution failed",
			zap.String("task_id", task.ID),
			zap.String("agent_id", task.AgentID),
			zap.Error(err))
		return
	}

	ss.logger.Info("scheduled task execution completed",
		zap.String("task_id", task.ID),
		zap.String("agent_id", task.AgentID),
		zap.String("execution_id", execution.ID))
}

// buildInputFromParameters builds an input string from task parameters
func (ss *SchedulerService) buildInputFromParameters(params map[string]interface{}) string {
	// In a real implementation, this would construct the input based on the agent type and parameters
	// For now, we'll just return a simple string representation
	if params == nil || len(params) == 0 {
		return ""
	}

	input := ""
	for key, value := range params {
		input += fmt.Sprintf("%s=%v ", key, value)
	}

	return input
}

// ScheduledAgent is a wrapper to make our agent configuration compatible with the execution service
type ScheduledAgent struct {
	config *models.AgentConfiguration
	logger *zap.Logger
}

func (sa *ScheduledAgent) Execute(ctx context.Context, input string) (*models.ExecutionResult, error) {
	// In a real implementation, this would execute the actual agent process
	// For this implementation, we'll return a mock result
	result := &models.ExecutionResult{
		ID:        "sched-" + time.Now().Format("20060102-150405"),
		AgentID:   sa.config.ID,
		Status:    models.SuccessStatus,
		Input:     input,
		Output:    fmt.Sprintf("Executed scheduled command: %s with input: %s", sa.config.ExecutablePath, input),
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}
	return result, nil
}

func (sa *ScheduledAgent) GetID() string {
	return sa.config.ID
}

func (sa *ScheduledAgent) GetName() string {
	return sa.config.Name
}

func (sa *ScheduledAgent) GetType() string {
	return sa.config.AgentType
}

func (sa *ScheduledAgent) IsReadOnly() bool {
	return sa.config.AccessType == models.ReadOnlyAccessType
}

func (sa *ScheduledAgent) GetConfig() *models.AgentConfiguration {
	return sa.config
}

func (sa *ScheduledAgent) Validate() error {
	return sa.config.Validate()
}

// CronLogger adapts zap logger to cron logger interface
type CronLogger struct {
	logger *zap.Logger
}

func (cl *CronLogger) Info(msg string, keysAndValues ...interface{}) {
	fields := make([]zap.Field, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			key = "unknown_key"
		}
		fields[i/2] = zap.Any(key, keysAndValues[i+1])
	}
	cl.logger.Info(msg, fields...)
}

func (cl *CronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	fields := make([]zap.Field, len(keysAndValues)/2+1)
	fields[0] = zap.Error(err)
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			key = "unknown_key"
		}
		fields[i/2+1] = zap.Any(key, keysAndValues[i+1])
	}
	cl.logger.Error(msg, fields...)
}