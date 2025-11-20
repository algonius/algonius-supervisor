package clients

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/algonius/algonius-supervisor/pkg/a2a"
)

// MessageSenderClient provides methods for sending messages to agents
type MessageSenderClient struct {
	a2aClient      *A2AClient
	discoveryClient *DiscoveryClient
}

// NewMessageSenderClient creates a new message sender client instance
func NewMessageSenderClient(a2aClient *A2AClient, discoveryClient *DiscoveryClient) *MessageSenderClient {
	return &MessageSenderClient{
		a2aClient:      a2aClient,
		discoveryClient: discoveryClient,
	}
}

// SendMessage sends a message to a specific agent
func (msc *MessageSenderClient) SendMessage(ctx context.Context, agentID string, message *a2a.A2AMessage) (*a2a.A2AMessage, error) {
	// Validate the agent exists by checking its card
	_, err := msc.discoveryClient.GetAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("validation failed for agent %s: %w", agentID, err)
	}
	
	// Send the message using the A2A client
	response, err := msc.a2aClient.SendMessage(ctx, agentID, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send message to agent %s: %w", agentID, err)
	}
	
	return response, nil
}

// SendMessageWithRetry sends a message with automatic retry on transient errors
func (msc *MessageSenderClient) SendMessageWithRetry(ctx context.Context, agentID string, message *a2a.A2AMessage, maxRetries int) (*a2a.A2AMessage, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		response, err := msc.SendMessage(ctx, agentID, message)
		if err == nil {
			return response, nil
		}
		
		// Check if the error is a transient error that allows for retry
		a2aErr, ok := a2a.AsA2AError(err)
		if ok && (a2aErr.Code == a2a.InternalError || a2aErr.Code == a2a.AgentExecutionFailedError) {
			// This is a transient error, we can retry
			lastErr = err
			// In a real implementation, we'd add a backoff delay here
			continue
		} else if ok {
			// This is a permanent error, don't retry
			return nil, err
		}
		
		// Non-A2A error, determine if it's transient
		// In a real implementation, we'd have more sophisticated error classification
		lastErr = err
	}
	
	return nil, fmt.Errorf("failed to send message after %d retries: %w", maxRetries, lastErr)
}

// BroadcastMessage sends a message to multiple agents
func (msc *MessageSenderClient) BroadcastMessage(ctx context.Context, agentIDs []string, message *a2a.A2AMessage) (map[string]*a2a.A2AMessage, map[string]error) {
	responses := make(map[string]*a2a.A2AMessage)
	errors := make(map[string]error)
	
	for _, agentID := range agentIDs {
		response, err := msc.SendMessage(ctx, agentID, message)
		if err != nil {
			errors[agentID] = err
		} else {
			responses[agentID] = response
		}
	}
	
	return responses, errors
}

// SendStructuredMessage builds and sends a structured A2A message
func (msc *MessageSenderClient) SendStructuredMessage(ctx context.Context, agentID, method string, params map[string]interface{}) (*a2a.A2AMessage, error) {
	// Build the A2A message
	message := &a2a.A2AMessage{
		Protocol:  "a2a",
		Version:   "0.3.0",
		ID:        GenerateA2AID(), // This would be a function to generate unique IDs
		Type:      "request",
		Timestamp: GetMessageTimestamp(), // This would be a function to get current time in proper format
		Context: &a2a.A2AContext{
			From:          "supervisor", // This would be dynamically determined
			To:            agentID,
			ConversationID: GenerateConversationID(), // This would be a function to generate conversation IDs
			MessageID:     GenerateA2AID(), // This would be a function to generate message IDs
		},
		Payload: &a2a.A2APayload{
			Method: method,
			Params: params,
		},
	}
	
	// Send the message
	response, err := msc.SendMessage(ctx, agentID, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send structured message to agent %s: %w", agentID, err)
	}
	
	return response, nil
}

// SendExecuteAgentRequest sends an execute-agent request to a specific agent
func (msc *MessageSenderClient) SendExecuteAgentRequest(ctx context.Context, agentID, input string) (*a2a.A2AMessage, error) {
	params := map[string]interface{}{
		"command": input,
	}
	
	return msc.SendStructuredMessage(ctx, agentID, "execute-agent", params)
}

// SendStatusRequest sends a status request to a specific agent
func (msc *MessageSenderClient) SendStatusRequest(ctx context.Context, agentID string) (*a2a.A2AMessage, error) {
	params := map[string]interface{}{}
	
	return msc.SendStructuredMessage(ctx, agentID, "status", params)
}

// SendListAgentsRequest sends a list-agents request
func (msc *MessageSenderClient) SendListAgentsRequest(ctx context.Context, agentID string) (*a2a.A2AMessage, error) {
	params := map[string]interface{}{}
	
	return msc.SendStructuredMessage(ctx, agentID, "list-agents", params)
}

// Helper functions for message generation

// GenerateA2AID generates a unique ID for A2A messages
func GenerateA2AID() string {
	// In a real implementation, this would generate a proper UUID
	return "temp-id" // Placeholder
}

// GetMessageTimestamp gets the current time in RFC3339 format for A2A messages
func GetMessageTimestamp() interface{} {
	// In a real implementation, this would return the current time in the proper format
	return "2025-11-20T10:30:00Z" // Placeholder
}

// GenerateConversationID generates a conversation ID for A2A messages
func GenerateConversationID() string {
	// In a real implementation, this would generate a proper conversation ID
	return "temp-conversation-id" // Placeholder
}