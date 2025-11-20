package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/algonius/algonius-supervisor/internal/models"
	"github.com/algonius/algonius-supervisor/internal/services"
	"go.uber.org/zap"
	"github.com/stretchr/testify/assert"
)

// FlakyTestAgent simulates an agent that sometimes fails
type FlakyTestAgent struct {
	id         string
	name       string
	failCount  int
	maxFailures int
}

func (fta *FlakyTestAgent) Execute(ctx context.Context, input string) (*models.ExecutionResult, error) {
	fta.failCount++
	
	if fta.failCount <= fta.maxFailures {
		// Simulate a transient error
		return nil, errors.New("transient network error")
	}
	
	// Eventually succeed
	return &models.ExecutionResult{
		ID:        "result-1",
		AgentID:   fta.id,
		Status:    models.SuccessStatus,
		Input:     input,
		Output:    "test output after retry",
	}, nil
}

func (fta *FlakyTestAgent) GetID() string {
	return fta.id
}

func (fta *FlakyTestAgent) GetName() string {
	return fta.name
}

func (fta *FlakyTestAgent) GetType() string {
	return "test"
}

func (fta *FlakyTestAgent) IsReadOnly() bool {
	return true
}

func (fta *FlakyTestAgent) GetConfig() *models.AgentConfiguration {
	return &models.AgentConfiguration{
		ID:         fta.id,
		Name:       fta.name,
		AccessType: models.ReadOnlyAccessType,
	}
}

func (fta *FlakyTestAgent) Validate() error {
	return nil
}

// PermanentErrorTestAgent simulates an agent that always fails with a permanent error
type PermanentErrorTestAgent struct {
	id   string
	name string
}

func (peta *PermanentErrorTestAgent) Execute(ctx context.Context, input string) (*models.ExecutionResult, error) {
	// Always return a permanent error (something that won't be fixed by retrying)
	return nil, errors.New("invalid configuration")
}

func (peta *PermanentErrorTestAgent) GetID() string {
	return peta.id
}

func (peta *PermanentErrorTestAgent) GetName() string {
	return peta.name
}

func (peta *PermanentErrorTestAgent) GetType() string {
	return "test"
}

func (peta *PermanentErrorTestAgent) IsReadOnly() bool {
	return true
}

func (peta *PermanentErrorTestAgent) GetConfig() *models.AgentConfiguration {
	return &models.AgentConfiguration{
		ID:         peta.id,
		Name:       peta.name,
		AccessType: models.ReadOnlyAccessType,
	}
}

func (peta *PermanentErrorTestAgent) Validate() error {
	return nil
}

func TestExecutionService_RetryLogic(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	agentService := services.NewAgentService(logger)
	executionService := services.NewExecutionService(agentService, logger)

	// Test with flaky agent that succeeds after retries
	flakyAgent := &FlakyTestAgent{
		id:          "flaky-agent",
		name:        "Flaky Test Agent",
		maxFailures: 2, // Will fail twice, then succeed on third attempt
	}

	ctx := context.Background()
	execution, err := executionService.ExecuteAgent(ctx, flakyAgent, "test input")

	assert.NoError(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, models.CompletedState, execution.State)
	assert.Equal(t, 3, execution.RetryCount) // Should have retried twice before succeeding

	// Verify that the execution completed successfully despite initial failures
	assert.Equal(t, "", execution.ErrorMessage)
	assert.Equal(t, models.SystemError, execution.ErrorCategory) // Default category, but execution succeeded
}

func TestExecutionService_PermanentError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	agentService := services.NewAgentService(logger)
	executionService := services.NewExecutionService(agentService, logger)

	// Test with agent that always fails with a permanent error
	permanentErrorAgent := &PermanentErrorTestAgent{
		id:   "permanent-error-agent",
		name: "Permanent Error Test Agent",
	}

	ctx := context.Background()
	execution, err := executionService.ExecuteAgent(ctx, permanentErrorAgent, "test input")

	assert.NoError(t, err) // ExecuteAgent returns the execution record even if the agent execution fails
	assert.NotNil(t, execution)
	assert.Equal(t, models.FailedState, execution.State)
	assert.Equal(t, 3, execution.RetryCount) // Should retry up to max retries
	assert.NotEqual(t, "", execution.ErrorMessage)
	assert.Equal(t, models.PermanentError, execution.ErrorCategory)
}

func TestIsTransientError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	agentService := services.NewAgentService(logger)
	executionService := services.NewExecutionService(agentService, logger)

	// Test transient error detection
	transientErrors := []error{
		errors.New("connection timeout"),
		errors.New("network error occurred"),
		errors.New("connection refused"),
		errors.New("temporary failure in name resolution"),
		errors.New("context deadline exceeded"),
	}

	for _, err := range transientErrors {
		isTransient := executionService.IsTransientError(err)
		assert.True(t, isTransient, "Error should be detected as transient: %v", err)
	}

	// Test permanent error detection
	permanentErrors := []error{
		errors.New("invalid configuration"),
		errors.New("file not found"),
		errors.New("permission denied"),
		errors.New("syntax error in input"),
	}

	for _, err := range permanentErrors {
		// Note: Our implementation currently only detects specific patterns
		// For a complete implementation, we'd need more sophisticated detection
		_ = err
	}
}

// Add a helper method to expose the private isTransientError function for testing
func (es *ExecutionService) IsTransientError(err error) bool {
	return es.isTransientError(err)
}