package integration

import (
	"context"
	"testing"
	"time"

	"github.com/algonius/algonius-supervisor/internal/models"
	"github.com/algonius/algonius-supervisor/internal/services"
	"go.uber.org/zap"
	"github.com/stretchr/testify/assert"
)

// Integration test for agent execution flow
func TestAgentExecutionFlow(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Setup agent service
	agentService := services.NewAgentService(logger)

	// Setup execution service
	executionService := services.NewExecutionService(agentService, logger)

	// Register an agent configuration
	agentConfig := &models.AgentConfiguration{
		ID:                    "integration-test-agent",
		Name:                  "Integration Test Agent",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo", // Using echo as a simple test executable
		AccessType:            models.ReadOnlyAccessType,
		MaxConcurrentExecutions: 5,
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	err := agentService.RegisterAgent(agentConfig)
	assert.NoError(t, err)

	// Verify agent was registered
	registeredAgent, err := agentService.GetAgent("integration-test-agent")
	assert.NoError(t, err)
	assert.Equal(t, agentConfig.ID, registeredAgent.ID)

	// Create a test agent implementation to work with our services
	testAgent := &IntegrationTestAgent{
		id:   "integration-test-agent",
		name: "Integration Test Agent",
	}

	// Execute the agent
	ctx := context.Background()
	execution, err := executionService.ExecuteAgent(ctx, testAgent, "hello world")
	assert.NoError(t, err)
	assert.NotNil(t, execution)

	// Verify execution was tracked
	retrievedExecution, err := executionService.GetExecution(execution.ID)
	assert.NoError(t, err)
	assert.Equal(t, execution.ID, retrievedExecution.ID)

	// Verify execution completed successfully
	assert.Equal(t, models.CompletedState, retrievedExecution.State)
	assert.True(t, retrievedExecution.StartTime.Before(time.Now()))
	assert.True(t, retrievedExecution.EndTime != nil)
	assert.True(t, retrievedExecution.StartTime.Before(*retrievedExecution.EndTime))

	// Verify execution result was stored
	executionResult, err := executionService.GetExecutionResult(execution.ID)
	assert.NoError(t, err)
	assert.NotNil(t, executionResult)
	assert.Equal(t, models.SuccessStatus, executionResult.Status)

	// List executions for the agent
	executions, err := executionService.ListExecutions(testAgent.GetID())
	assert.NoError(t, err)
	assert.Len(t, executions, 1)
	assert.Equal(t, execution.ID, executions[0].ID)
}

// Integration test for read-write agent execution with single concurrent execution
func TestReadWriteAgentExecution(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Setup agent service
	agentService := services.NewAgentService(logger)

	// Setup read-write execution service
	readWriteService := services.NewReadWriteExecutionService(agentService, logger)

	// Register a read-write agent configuration
	agentConfig := &models.AgentConfiguration{
		ID:                    "readwrite-integration-test-agent",
		Name:                  "ReadWrite Integration Test Agent",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadWriteAccessType,
		MaxConcurrentExecutions: 1, // Must be 1 for read-write
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	err := agentService.RegisterAgent(agentConfig)
	assert.NoError(t, err)

	// Verify agent was registered
	registeredAgent, err := agentService.GetAgent("readwrite-integration-test-agent")
	assert.NoError(t, err)
	assert.Equal(t, agentConfig.ID, registeredAgent.ID)
	assert.Equal(t, models.ReadWriteAccessType, registeredAgent.AccessType)

	// Create a test read-write agent
	testReadWriteAgent := &IntegrationReadWriteTestAgent{
		id:   "readwrite-integration-test-agent",
		name: "ReadWrite Integration Test Agent",
	}

	// Execute the read-write agent
	ctx := context.Background()
	execution, err := readWriteService.ExecuteAgent(ctx, testReadWriteAgent, "hello read-write world")
	assert.NoError(t, err)
	assert.NotNil(t, execution)

	// Verify execution was tracked
	retrievedExecution, err := readWriteService.GetExecution(execution.ID)
	assert.NoError(t, err)
	assert.Equal(t, execution.ID, retrievedExecution.ID)

	// Verify execution completed successfully
	assert.Equal(t, models.CompletedState, retrievedExecution.State)
	assert.True(t, retrievedExecution.StartTime.Before(time.Now()))
	assert.True(t, retrievedExecution.EndTime != nil)
	assert.True(t, retrievedExecution.StartTime.Before(*retrievedExecution.EndTime))
}

// Integration test for read-only agent execution with multiple concurrent executions
func TestReadOnlyAgentExecution(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Setup agent service
	agentService := services.NewAgentService(logger)

	// Setup read-only execution service with capacity for 3 concurrent executions
	readOnlyService := services.NewReadOnlyExecutionService(agentService, logger, 3)

	// Register a read-only agent configuration
	agentConfig := &models.AgentConfiguration{
		ID:                    "readonly-integration-test-agent",
		Name:                  "ReadOnly Integration Test Agent",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadOnlyAccessType,
		MaxConcurrentExecutions: 3,
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	err := agentService.RegisterAgent(agentConfig)
	assert.NoError(t, err)

	// Verify agent was registered
	registeredAgent, err := agentService.GetAgent("readonly-integration-test-agent")
	assert.NoError(t, err)
	assert.Equal(t, agentConfig.ID, registeredAgent.ID)
	assert.Equal(t, models.ReadOnlyAccessType, registeredAgent.AccessType)

	// Create a test read-only agent
	testReadOnlyAgent := &IntegrationReadOnlyTestAgent{
		id:   "readonly-integration-test-agent",
		name: "ReadOnly Integration Test Agent",
	}

	// Execute the read-only agent multiple times (up to the concurrent limit)
	ctx := context.Background()
	var executions []*models.AgentExecution

	for i := 0; i < 3; i++ {
		execution, err := readOnlyService.ExecuteAgent(ctx, testReadOnlyAgent, "hello read-only world "+string(rune('0'+i)))
		assert.NoError(t, err)
		assert.NotNil(t, execution)
		executions = append(executions, execution)
	}

	// Verify all executions were tracked
	for _, execution := range executions {
		retrievedExecution, err := readOnlyService.GetExecution(execution.ID)
		assert.NoError(t, err)
		assert.Equal(t, execution.ID, retrievedExecution.ID)
		assert.Equal(t, models.CompletedState, retrievedExecution.State)
	}

	// Test resource pool metrics
	metrics, err := readOnlyService.GetResourcePoolMetrics()
	assert.NoError(t, err)
	assert.Equal(t, 3, metrics.TotalCapacity)
	assert.Equal(t, 3, metrics.MaxConcurrent)
	// Note: Since our test executions complete immediately, UsedCapacity might be 0
	// depending on timing. This is acceptable for the test.
}

// Integration test for configuration validation
func TestConfigurationValidationIntegration(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Setup agent service
	agentService := services.NewAgentService(logger)

	// Test valid read-write configuration
	validReadWriteConfig := &models.AgentConfiguration{
		ID:                    "valid-readwrite-agent",
		Name:                  "Valid ReadWrite Agent",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadWriteAccessType,
		MaxConcurrentExecutions: 1, // Correct for read-write
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	err := agentService.ValidateAgentConfiguration(validReadWriteConfig)
	assert.NoError(t, err)

	err = agentService.RegisterAgent(validReadWriteConfig)
	assert.NoError(t, err)

	// Test valid read-only configuration
	validReadOnlyConfig := &models.AgentConfiguration{
		ID:                    "valid-readonly-agent",
		Name:                  "Valid ReadOnly Agent",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadOnlyAccessType,
		MaxConcurrentExecutions: 5, // Valid for read-only
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	err = agentService.ValidateAgentConfiguration(validReadOnlyConfig)
	assert.NoError(t, err)

	err = agentService.RegisterAgent(validReadOnlyConfig)
	assert.NoError(t, err)

	// Test invalid read-write configuration (concurrent executions > 1)
	invalidReadWriteConfig := &models.AgentConfiguration{
		ID:                    "invalid-readwrite-agent",
		Name:                  "Invalid ReadWrite Agent",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadWriteAccessType,
		MaxConcurrentExecutions: 2, // Invalid for read-write
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	err = agentService.ValidateAgentConfiguration(invalidReadWriteConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MaxConcurrentExecutions equal to 1")

	// Register the valid agents and verify they exist
	retrievedReadWriteAgent, err := agentService.GetAgent("valid-readwrite-agent")
	assert.NoError(t, err)
	assert.Equal(t, "valid-readwrite-agent", retrievedReadWriteAgent.ID)

	retrievedReadOnlyAgent, err := agentService.GetAgent("valid-readonly-agent")
	assert.NoError(t, err)
	assert.Equal(t, "valid-readonly-agent", retrievedReadOnlyAgent.ID)
}

// IntegrationTestAgent is a test implementation of the agent interface
type IntegrationTestAgent struct {
	id   string
	name string
}

func (ita *IntegrationTestAgent) Execute(ctx context.Context, input string) (*models.ExecutionResult, error) {
	return &models.ExecutionResult{
		ID:        "result-" + time.Now().Format("20060102-150405"),
		AgentID:   ita.id,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(100 * time.Millisecond),
		Status:    models.SuccessStatus,
		Input:     input,
		Output:    "Processed: " + input,
	}, nil
}

func (ita *IntegrationTestAgent) GetID() string {
	return ita.id
}

func (ita *IntegrationTestAgent) GetName() string {
	return ita.name
}

func (ita *IntegrationTestAgent) GetType() string {
	return "test"
}

func (ita *IntegrationTestAgent) IsReadOnly() bool {
	return true
}

func (ita *IntegrationTestAgent) GetConfig() *models.AgentConfiguration {
	return &models.AgentConfiguration{
		ID:         ita.id,
		Name:       ita.name,
		AccessType: models.ReadOnlyAccessType,
	}
}

func (ita *IntegrationTestAgent) Validate() error {
	return nil
}

// IntegrationReadWriteTestAgent is a test implementation for read-write testing
type IntegrationReadWriteTestAgent struct {
	id   string
	name string
}

func (irwta *IntegrationReadWriteTestAgent) Execute(ctx context.Context, input string) (*models.ExecutionResult, error) {
	return &models.ExecutionResult{
		ID:        "result-" + time.Now().Format("20060102-150405"),
		AgentID:   irwta.id,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(100 * time.Millisecond),
		Status:    models.SuccessStatus,
		Input:     input,
		Output:    "Processed read-write: " + input,
	}, nil
}

func (irwta *IntegrationReadWriteTestAgent) GetID() string {
	return irwta.id
}

func (irwta *IntegrationReadWriteTestAgent) GetName() string {
	return irwta.name
}

func (irwta *IntegrationReadWriteTestAgent) GetType() string {
	return "test-readwrite"
}

func (irwta *IntegrationReadWriteTestAgent) IsReadOnly() bool {
	return false
}

func (irwta *IntegrationReadWriteTestAgent) GetConfig() *models.AgentConfiguration {
	return &models.AgentConfiguration{
		ID:         irwta.id,
		Name:       irwta.name,
		AccessType: models.ReadWriteAccessType,
	}
}

func (irwta *IntegrationReadWriteTestAgent) Validate() error {
	return nil
}

// IntegrationReadOnlyTestAgent is a test implementation for read-only testing
type IntegrationReadOnlyTestAgent struct {
	id   string
	name string
}

func (irota *IntegrationReadOnlyTestAgent) Execute(ctx context.Context, input string) (*models.ExecutionResult, error) {
	return &models.ExecutionResult{
		ID:        "result-" + time.Now().Format("20060102-150405"),
		AgentID:   irota.id,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(100 * time.Millisecond),
		Status:    models.SuccessStatus,
		Input:     input,
		Output:    "Processed read-only: " + input,
	}, nil
}

func (irota *IntegrationReadOnlyTestAgent) GetID() string {
	return irota.id
}

func (irota *IntegrationReadOnlyTestAgent) GetName() string {
	return irota.name
}

func (irota *IntegrationReadOnlyTestAgent) GetType() string {
	return "test-readonly"
}

func (irota *IntegrationReadOnlyTestAgent) IsReadOnly() bool {
	return true
}

func (irota *IntegrationReadOnlyTestAgent) GetConfig() *models.AgentConfiguration {
	return &models.AgentConfiguration{
		ID:         irota.id,
		Name:       irota.name,
		AccessType: models.ReadOnlyAccessType,
	}
}

func (irota *IntegrationReadOnlyTestAgent) Validate() error {
	return nil
}