package a2a

import (
	"time"

	"github.com/a2aproject/a2a-go/a2a"
)

// A2AMessage represents the base message structure as defined in A2A spec v0.3.0
type A2AMessage struct {
	Protocol    string      `json:"protocol"`
	Version     string      `json:"version"`
	ID          string      `json:"id"`
	Type        string      `json:"type"` // "request", "response", "error", "stream"
	Timestamp   time.Time   `json:"timestamp"`
	InResponseTo string     `json:"inResponseTo,omitempty"`
	Context     *A2AContext `json:"context"`
	Payload     *A2APayload `json:"payload"`
}

// A2ATask represents an A2A task
type A2ATask struct {
	ID          string      `json:"id"`
	Status      string      `json:"status"` // created, running, succeeded, failed, cancelled, expired
	Messages    []A2AMessage `json:"messages"`
	Artifacts   []A2AArtifact `json:"artifacts,omitempty"`
	History     []A2AMessage `json:"history,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	ModifiedAt  time.Time   `json:"modifiedAt"`
}

// A2AArtifact represents an artifact produced by an A2A task
type A2AArtifact struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Description string `json:"description"`
	URI         string `json:"uri"`
	Size        int64  `json:"size,omitempty"`
	Hash        string `json:"hash,omitempty"`
}

// A2AMessagePart represents a part of an A2A message (for content)
type A2AMessagePart struct {
	Type  string `json:"type"`  // "text", "code", "file", etc.
	Value string `json:"value"` // The actual content value
}

// A2APushNotificationConfig represents push notification configuration for A2A tasks
type A2APushNotificationConfig struct {
	ID       string                 `json:"id"`
	URL      string                 `json:"url"`
	Token    string                 `json:"token"`
	Authentication *A2AAuth         `json:"authentication"`
}

// A2ATaskConfiguration represents configuration for an A2A task
type A2ATaskConfiguration struct {
	AcceptedOutputModes      []string                    `json:"acceptedOutputModes,omitempty"`
	HistoryLength            int                         `json:"historyLength,omitempty"`
	PushNotificationConfig   *A2APushNotificationConfig  `json:"pushNotificationConfig,omitempty"`
	Blocking                 bool                        `json:"blocking"`
}

// A2AExtendedCard represents an extended A2A agent card
type A2AExtendedCard struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Version         string            `json:"version"`
	Protocols       map[string]interface{} `json:"protocols"`
	Capabilities    *A2ACapabilities  `json:"capabilities"`
	Endpoints       map[string]string `json:"endpoints"`
	SupportedOutputModes []string     `json:"supportedOutputModes"`
	SupportsAuthenticatedExtendedCard bool `json:"supportsAuthenticatedExtendedCard"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ConvertMessageToA2AGo converts an internal A2A message to the a2a-go library's message format
func ConvertMessageToA2AGo(msg *A2AMessage) *a2a.Message {
	if msg == nil {
		return nil
	}

	// Create the a2a-go message format
	a2aGoMsg := &a2a.Message{
		ID:        msg.ID,
		ContextID: msg.Context.ConversationID,
		Role:      a2a.MessageRole(msg.Type), // Map our Type to their Role
		Metadata: map[string]any{
			"timestamp":     msg.Timestamp,
			"inResponseTo":  msg.InResponseTo,
			"from":          msg.Context.From,
			"to":            msg.Context.To,
			"messageId":     msg.Context.MessageID,
			"protocol":      msg.Protocol,
			"version":       msg.Version,
		},
	}

	// Add payload information to metadata if it exists
	if msg.Payload != nil {
		a2aGoMsg.Metadata["payload"] = map[string]any{
			"method": msg.Payload.Method,
			"params": msg.Payload.Params,
			"result": msg.Payload.Result,
		}

		if msg.Payload.Error != nil {
			a2aGoMsg.Metadata["error"] = map[string]any{
				"code":    msg.Payload.Error.Code,
				"message": msg.Payload.Error.Message,
				"data":    msg.Payload.Error.Data,
			}
		}
	}

	return a2aGoMsg
}

// ConvertMessageFromA2AGo converts an a2a-go library message to internal A2A message format
func ConvertMessageFromA2AGo(msg *a2a.Message) *A2AMessage {
	if msg == nil {
		return nil
	}

	// Create internal message format from a2a-go message
	internalMsg := &A2AMessage{
		Protocol:  "a2a",
		Version:   "0.3.0",
		ID:        msg.ID,
		Type:      string(msg.Role), // Map their Role to our Type
		Timestamp: time.Now(), // Default to current time since a2a-go doesn't have timestamp
		Context: &A2AContext{
			ConversationID: msg.ContextID,
			MessageID:      msg.ID,
		},
		Payload: &A2APayload{},
	}

	// Extract additional context from metadata if available
	if msg.Metadata != nil {
		if timestamp, ok := msg.Metadata["timestamp"].(time.Time); ok {
			internalMsg.Timestamp = timestamp
		}
		if inResponseTo, ok := msg.Metadata["inResponseTo"].(string); ok {
			internalMsg.InResponseTo = inResponseTo
		}
		if from, ok := msg.Metadata["from"].(string); ok {
			internalMsg.Context.From = from
		}
		if to, ok := msg.Metadata["to"].(string); ok {
			internalMsg.Context.To = to
		}
		if messageId, ok := msg.Metadata["messageId"].(string); ok {
			internalMsg.Context.MessageID = messageId
		}

		// Extract payload from metadata if available
		if payloadData, ok := msg.Metadata["payload"].(map[string]any); ok {
			if method, ok := payloadData["method"].(string); ok {
				internalMsg.Payload.Method = method
			}
			if params, ok := payloadData["params"].(map[string]any); ok {
				internalMsg.Payload.Params = params
			}
			if result, ok := payloadData["result"]; ok {
				internalMsg.Payload.Result = result
			}
		}

		// Extract error from metadata if available
		if errorData, ok := msg.Metadata["error"].(map[string]any); ok {
			if code, ok := errorData["code"].(int); ok {
				if message, ok := errorData["message"].(string); ok {
					if data, ok := errorData["data"]; ok {
						internalMsg.Payload.Error = &A2AError{
							Code:    code,
							Message: message,
							Data:    data,
						}
					}
				}
			}
		}
	}

	return internalMsg
}