package services

import (
	"context"
	"fmt"

	"github.com/algonius/algonius-supervisor/internal/models"
	"github.com/algonius/algonius-supervisor/pkg/types"
	"github.com/a2aproject/a2a-go/a2asrv"
	"github.com/a2aproject/a2a-go/a2asrv/eventqueue"
	"go.uber.org/zap"
	a2a "github.com/a2aproject/a2a-go/a2a"
)

// AgentExecutor implements the A2A protocol's AgentExecutor interface
type AgentExecutor struct {
	agentService    IAgentService
	executionService IExecutionService
	logger          *zap.Logger
}

// Execute implements the a2asrv.AgentExecutor interface
func (ae *AgentExecutor) Execute(ctx context.Context, reqCtx *a2asrv.RequestContext, queue eventqueue.Queue) error {
	ae.logger.Info("executing A2A request",
		zap.String("task_id", string(reqCtx.TaskID)),
		zap.String("context_id", reqCtx.ContextID))

	// Check if this is a new task or continuing an existing one
	if reqCtx.StoredTask == nil {
		event := a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateSubmitted, nil)
		if err := queue.Write(ctx, event); err != nil {
			return fmt.Errorf("failed to write state submitted: %w", err)
		}
	}

	// Update task status to working
	event := a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateWorking, nil)
	if err := queue.Write(ctx, event); err != nil {
			return fmt.Errorf("failed to write state working: %w", err)
	}

	// Process the message parts to extract agent execution parameters
	var input string
	if reqCtx.Message != nil && len(reqCtx.Message.Parts) > 0 {
		// For simplicity, we'll take the text content from the first part
		if textPart, ok := reqCtx.Message.Parts[0].(*a2a.TextPart); ok {
			input = textPart.Text
		}
	}

	// Extract agent ID from context or metadata
	agentID := "default-agent" // Default if no specific agent identified

	// Try to extract agent ID from metadata
	if reqCtx.Metadata != nil {
		if id, ok := reqCtx.Metadata["agent_id"]; ok {
			if agentIDStr, ok := id.(string); ok && agentIDStr != "" {
				agentID = agentIDStr
			}
		}
	}

	// Get the agent configuration
	agentConfig, err := ae.agentService.GetAgent(agentID)
	if err != nil {
		ae.logger.Error("agent not found", zap.String("agent_id", agentID), zap.Error(err))
		errorEvent := a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateFailed, ae.createErrorMessage("Agent not found", err.Error()))
		if err := queue.Write(ctx, errorEvent); err != nil {
			return fmt.Errorf("failed to write error event: %w", err)
		}
		return nil
	}

	// Create an agent wrapper for execution
	simpleAgent := &SimpleA2AAgent{
		config: agentConfig,
	}

	// Execute the agent
	execution, err := ae.executionService.ExecuteAgent(ctx, simpleAgent, input)
	if err != nil {
		ae.logger.Error("agent execution failed", zap.Error(err))
		errorEvent := a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateFailed, ae.createErrorMessage("Agent execution failed", err.Error()))
		if err := queue.Write(ctx, errorEvent); err != nil {
			return fmt.Errorf("failed to write error event: %w", err)
		}
		return nil
	}

	// Create response parts
	parts := []a2a.Part{
		&a2a.TextPart{
			Text: fmt.Sprintf("Agent execution completed. Execution ID: %s, Status: %s", execution.ID, string(execution.State)),
		},
	}

	// Write the result to the event queue
	artifactEvent := a2a.NewArtifactEvent(reqCtx, parts...)

	if err := queue.Write(ctx, artifactEvent); err != nil {
		return fmt.Errorf("failed to write artifact event: %w", err)
	}

	// Mark the task as completed
	completedEvent := a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateCompleted, nil)
	completedEvent.Final = true
	if err := queue.Write(ctx, completedEvent); err != nil {
		return fmt.Errorf("failed to write completed event: %w", err)
	}

	return nil
}

// Cancel implements the a2asrv.AgentExecutor interface
func (ae *AgentExecutor) Cancel(ctx context.Context, reqCtx *a2asrv.RequestContext, queue eventqueue.Queue) error {
	ae.logger.Info("canceling A2A request",
		zap.String("task_id", string(reqCtx.TaskID)),
		zap.String("context_id", reqCtx.ContextID))

	// Write a cancellation event to the queue
	cancelEvent := a2a.NewStatusUpdateEvent(reqCtx, a2a.TaskStateCanceled, ae.createErrorMessage("Task cancelled", "Execution was cancelled by request"))
	if err := queue.Write(ctx, cancelEvent); err != nil {
		return fmt.Errorf("failed to write cancellation event: %w", err)
	}

	return nil
}

// createErrorMessage creates an A2A message for error reporting
func (ae *AgentExecutor) createErrorMessage(message, data string) *a2a.Message {
	return &a2a.Message{
		ID: generateA2AID(),
	}
}

// SimpleA2AAgent is a basic implementation of the agents.IAgent interface
type SimpleA2AAgent struct {
	config *models.AgentConfiguration
}

// Execute implements the agents.IAgent interface
func (s *SimpleA2AAgent) Execute(ctx context.Context, input string) (*models.ExecutionResult, error) {
	// In a real implementation, this would execute the actual agent
	// For now, we'll return a simple success result

	result := &models.ExecutionResult{
		ID:      generateA2AID(),
		AgentID: s.config.ID,
		Status:  types.SuccessStatus,
		Input:   input,
		Output:  fmt.Sprintf("Executed agent '%s' with input: %s", s.config.Name, input),
		Error:   "",
	}

	return result, nil
}

// GetID implements the agents.IAgent interface
func (s *SimpleA2AAgent) GetID() string {
	return s.config.ID
}

// GetName implements the agents.IAgent interface
func (s *SimpleA2AAgent) GetName() string {
	return s.config.Name
}

// GetType implements the agents.IAgent interface
func (s *SimpleA2AAgent) GetType() string {
	return string(s.config.AgentType)
}

// IsReadOnly implements the agents.IAgent interface
func (s *SimpleA2AAgent) IsReadOnly() bool {
	return s.config.AccessType == models.ReadOnlyAccessType
}

// GetConfig implements the agents.IAgent interface
func (s *SimpleA2AAgent) GetConfig() *models.AgentConfiguration {
	return s.config
}

// Validate implements the agents.IAgent interface
func (s *SimpleA2AAgent) Validate() error {
	if s.config == nil {
		return fmt.Errorf("agent configuration is nil")
	}

	return s.config.Validate()
}

// Helper functions
func generateA2AID() string {
	// In a real implementation, this would generate a proper A2A ID
	return "generated-id-" + fmt.Sprint(a2a.NewMessageID())
}

func getCurrentTimestamp() string {
	// In a real implementation, this would return the current time in RFC3339 format
	return "2025-11-20T10:30:00Z"
}