package unit

import (
	"context"
	"testing"
	"time"

	"github.com/algonius/algonius-supervisor/internal/models"
	"github.com/algonius/algonius-supervisor/internal/services"
	"go.uber.org/zap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAgent for testing
type MockAgent struct {
	mock.Mock
	ID      string
	Name    string
	ReadOnly bool
}

func (m *MockAgent) Execute(ctx context.Context, input string) (*models.ExecutionResult, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*models.ExecutionResult), args.Error(1)
}

func (m *MockAgent) GetID() string {
	return m.ID
}

func (m *MockAgent) GetName() string {
	return m.Name
}

func (m *MockAgent) GetType() string {
	return "test-agent"
}

func (m *MockAgent) IsReadOnly() bool {
	return m.ReadOnly
}

func (m *MockAgent) GetConfig() *models.AgentConfiguration {
	return &models.AgentConfiguration{ID: m.ID, Name: m.Name}
}

func (m *MockAgent) Validate() error {
	return nil
}

func TestAgentService_RegisterAgent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	agentService := services.NewAgentService(logger)

	// Create a valid agent configuration
	config := &models.AgentConfiguration{
		ID:                    "test-agent-1",
		Name:                  "Test Agent",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadOnlyAccessType,
		MaxConcurrentExecutions: 5,
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	// Test successful registration
	err := agentService.RegisterAgent(config)
	assert.NoError(t, err)

	// Test duplicate registration
	err = agentService.RegisterAgent(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test invalid configuration
	invalidConfig := &models.AgentConfiguration{
		ID: "test-agent-2",
		// Missing required fields
	}
	err = agentService.RegisterAgent(invalidConfig)
	assert.Error(t, err)
}

func TestAgentService_GetAgent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	agentService := services.NewAgentService(logger)

	// Create and register an agent
	config := &models.AgentConfiguration{
		ID:                    "test-agent-3",
		Name:                  "Test Agent 3",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadOnlyAccessType,
		MaxConcurrentExecutions: 5,
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	err := agentService.RegisterAgent(config)
	assert.NoError(t, err)

	// Test getting existing agent
	retrievedConfig, err := agentService.GetAgent("test-agent-3")
	assert.NoError(t, err)
	assert.Equal(t, config.ID, retrievedConfig.ID)
	assert.Equal(t, config.Name, retrievedConfig.Name)

	// Test getting non-existing agent
	_, err = agentService.GetAgent("non-existing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAgentService_ListAgents(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	agentService := services.NewAgentService(logger)

	// Create and register multiple agents
	agent1 := &models.AgentConfiguration{
		ID:                    "test-agent-4",
		Name:                  "Test Agent 4",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadOnlyAccessType,
		MaxConcurrentExecutions: 5,
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	agent2 := &models.AgentConfiguration{
		ID:                    "test-agent-5",
		Name:                  "Test Agent 5",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadWriteAccessType,
		MaxConcurrentExecutions: 1,
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	err := agentService.RegisterAgent(agent1)
	assert.NoError(t, err)

	err = agentService.RegisterAgent(agent2)
	assert.NoError(t, err)

	// Test listing agents
	agents, err := agentService.ListAgents()
	assert.NoError(t, err)
	assert.Len(t, agents, 2)

	// Check that both agents are in the list
	found1, found2 := false, false
	for _, a := range agents {
		if a.ID == agent1.ID {
			found1 = true
		}
		if a.ID == agent2.ID {
			found2 = true
		}
	}
	assert.True(t, found1)
	assert.True(t, found2)
}

func TestAgentService_UpdateAgent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	agentService := services.NewAgentService(logger)

	// Create and register an agent
	config := &models.AgentConfiguration{
		ID:                    "test-agent-6",
		Name:                  "Test Agent 6",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadOnlyAccessType,
		MaxConcurrentExecutions: 5,
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	err := agentService.RegisterAgent(config)
	assert.NoError(t, err)

	// Update the agent
	updatedConfig := &models.AgentConfiguration{
		ID:                    "test-agent-6",
		Name:                  "Updated Test Agent 6",
		AgentType:             "updated-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadOnlyAccessType,
		MaxConcurrentExecutions: 10,
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               60,
		Enabled:               true,
	}

	err = agentService.UpdateAgent(updatedConfig)
	assert.NoError(t, err)

	// Verify the update
	retrievedConfig, err := agentService.GetAgent("test-agent-6")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Test Agent 6", retrievedConfig.Name)
	assert.Equal(t, 60, retrievedConfig.Timeout)
}

func TestAgentService_DeleteAgent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	agentService := services.NewAgentService(logger)

	// Create and register an agent
	config := &models.AgentConfiguration{
		ID:                    "test-agent-7",
		Name:                  "Test Agent 7",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadOnlyAccessType,
		MaxConcurrentExecutions: 5,
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	err := agentService.RegisterAgent(config)
	assert.NoError(t, err)

	// Verify it exists
	_, err = agentService.GetAgent("test-agent-7")
	assert.NoError(t, err)

	// Delete the agent
	err = agentService.DeleteAgent("test-agent-7")
	assert.NoError(t, err)

	// Verify it no longer exists
	_, err = agentService.GetAgent("test-agent-7")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Try to delete non-existing agent
	err = agentService.DeleteAgent("non-existing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAgentService_ValidateAgentConfiguration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	agentService := services.NewAgentService(logger)

	// Test valid configuration
	validConfig := &models.AgentConfiguration{
		ID:                    "test-valid",
		Name:                  "Valid Agent",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadOnlyAccessType,
		MaxConcurrentExecutions: 5,
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	err := agentService.ValidateAgentConfiguration(validConfig)
	assert.NoError(t, err)

	// Test read-write access type with max concurrent > 1 (should fail)
	invalidConfig1 := &models.AgentConfiguration{
		ID:                    "test-invalid-1",
		Name:                  "Invalid Agent 1",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadWriteAccessType,
		MaxConcurrentExecutions: 2, // Should be 1 for read-write
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Timeout:               30,
		Enabled:               true,
	}

	err = agentService.ValidateAgentConfiguration(invalidConfig1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MaxConcurrentExecutions equal to 1")

	// Test config with file patterns but no templates (should fail)
	invalidConfig2 := &models.AgentConfiguration{
		ID:                    "test-invalid-2",
		Name:                  "Invalid Agent 2",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadOnlyAccessType,
		MaxConcurrentExecutions: 5,
		Mode:                  models.TaskMode,
		InputPattern:          models.FilePattern, // Requires InputFileTemplate
		OutputPattern:         models.FilePatternOut, // Requires OutputFileTemplate
		Timeout:               30,
		Enabled:               true,
	}

	err = agentService.ValidateAgentConfiguration(invalidConfig2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input file pattern requires an input file template")

	// Test config with sensitive environment variable (should fail)
	invalidConfig3 := &models.AgentConfiguration{
		ID:                    "test-invalid-3",
		Name:                  "Invalid Agent 3",
		AgentType:             "test-type",
		ExecutablePath:        "/bin/echo",
		AccessType:            models.ReadOnlyAccessType,
		MaxConcurrentExecutions: 5,
		Mode:                  models.TaskMode,
		InputPattern:          models.StdinPattern,
		OutputPattern:         models.StdoutPattern,
		Envs:                  map[string]string{"API_PASSWORD": "secret123"},
		Timeout:               30,
		Enabled:               true,
	}

	err = agentService.ValidateAgentConfiguration(invalidConfig3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "might contain sensitive information")
}