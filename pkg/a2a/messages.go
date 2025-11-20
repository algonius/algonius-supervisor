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
		ID:          msg.ID,
		Type:        a2a.MessageType(msg.Type),
		Timestamp:   msg.Timestamp,
		InResponseTo: msg.InResponseTo,
		Context: &a2a.Context{
			From:          msg.Context.From,
			To:            msg.Context.To,
			ConversationID: msg.Context.ConversationID,
			MessageID:     msg.Context.MessageID,
		},
		Payload: &a2a.Payload{},
	}

	// Copy payload if it exists
	if msg.Payload != nil {
		a2aGoMsg.Payload = &a2a.Payload{
			Method: msg.Payload.Method,
			Params: msg.Payload.Params,
			Result: msg.Payload.Result,
		}

		if msg.Payload.Error != nil {
			a2aGoMsg.Payload.Error = &a2a.Error{
				Code:    msg.Payload.Error.Code,
				Message: msg.Payload.Error.Message,
				Data:    msg.Payload.Error.Data,
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
		Protocol:    "a2a",
		Version:     "0.3.0",
		ID:          msg.ID,
		Type:        string(msg.Type),
		Timestamp:   msg.Timestamp,
		InResponseTo: msg.InResponseTo,
		Context: &A2AContext{
			From:          msg.Context.From,
			To:            msg.Context.To,
			ConversationID: msg.Context.ConversationID,
			MessageID:     msg.Context.MessageID,
		},
		Payload: &A2APayload{},
	}

	// Copy payload if it exists
	if msg.Payload != nil {
		internalMsg.Payload.Method = msg.Payload.Method
		internalMsg.Payload.Params = msg.Payload.Params
		internalMsg.Payload.Result = msg.Payload.Result

		if msg.Payload.Error != nil {
			internalMsg.Payload.Error = &A2AError{
				Code:    msg.Payload.Error.Code,
				Message: msg.Payload.Error.Message,
				Data:    msg.Payload.Error.Data,
			}
		}
	}

	return internalMsg
}