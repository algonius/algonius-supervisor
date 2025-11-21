package a2a

import "fmt"

// A2A-specific error codes as defined in the A2A protocol specification
const (
	// Standard JSON-RPC 2.0 errors
	ParseError     = -32700 // Invalid JSON was received by the server
	InvalidRequest = -32600 // The JSON sent is not a valid Request object
	MethodNotFound = -32601 // The method does not exist / is not available
	InvalidParams  = -32602 // Invalid method parameter(s)
	InternalError  = -32603 // Internal JSON-RPC error

	// A2A-specific errors
	TaskNotFoundError              = -32001 // Task doesn't exist or has expired
	TaskNotCancelableError         = -32002 // Task cannot be cancelled (already terminated)
	PushNotificationNotSupportedError = -32003 // Agent doesn't support push notifications
	UnsupportedOperationError      = -32004 // Operation not supported by agent
	ContentTypeNotSupportedError   = -32005 // Content type not supported
	AgentNotFoundError             = -32006 // Agent not found
	AgentExecutionFailedError      = -32007 // Agent execution failed
	AuthenticationRequiredError    = -32008 // Authentication required
	ConcurrentExecutionLimitError  = -32009 // Concurrent execution limit exceeded
	SensitiveDataError             = -32010 // Sensitive data detected in input/output
	AgentConfigurationError        = -32011 // Agent configuration error
	InvalidAgentStateError         = -32012 // Agent is in invalid state to process request
	AgentTimeoutError              = -32013 // Agent execution timed out
)

// A2AError represents an error in the A2A protocol
type A2AError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error returns the error message
func (e *A2AError) Error() string {
	return fmt.Sprintf("A2A Error [%d]: %s", e.Code, e.Message)
}

// NewA2AError creates a new A2A error with the specified code and message
func NewA2AError(code int, message string) *A2AError {
	return &A2AError{
		Code:    code,
		Message: message,
	}
}

// NewA2AErrorWithData creates a new A2A error with the specified code, message, and additional data
func NewA2AErrorWithData(code int, message string, data interface{}) *A2AError {
	return &A2AError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// IsA2AError checks if an error is an A2AError
func IsA2AError(err error) bool {
	_, ok := err.(*A2AError)
	return ok
}

// AsA2AError attempts to convert an error to an A2AError
func AsA2AError(err error) (*A2AError, bool) {
	a2aErr, ok := err.(*A2AError)
	if ok {
		return a2aErr, true
	}
	
	// Check if it's a wrapped error that contains an A2AError
	return nil, false
}

// Common error constructors for convenience
var (
	// Authentication errors
	ErrAuthenticationRequired = &A2AError{
		Code:    AuthenticationRequiredError,
		Message: "Authentication required",
	}

	// Agent errors
	ErrAgentNotFound = &A2AError{
		Code:    AgentNotFoundError,
		Message: "Agent not found",
	}

	ErrAgentExecutionFailed = &A2AError{
		Code:    AgentExecutionFailedError,
		Message: "Agent execution failed",
	}

	ErrAgentConfiguration = &A2AError{
		Code:    AgentConfigurationError,
		Message: "Agent configuration error",
	}

	ErrInvalidAgentState = &A2AError{
		Code:    InvalidAgentStateError,
		Message: "Agent is in invalid state to process request",
	}

	ErrAgentTimeout = &A2AError{
		Code:    AgentTimeoutError,
		Message: "Agent execution timed out",
	}

	// Task errors
	ErrTaskNotFound = &A2AError{
		Code:    TaskNotFoundError,
		Message: "Task not found",
	}

	ErrTaskNotCancelable = &A2AError{
		Code:    TaskNotCancelableError,
		Message: "Task cannot be cancelled",
	}

	// Operation errors
	ErrUnsupportedOperation = &A2AError{
		Code:    UnsupportedOperationError,
		Message: "Operation not supported by agent",
	}

	ErrConcurrentExecutionLimit = &A2AError{
		Code:    ConcurrentExecutionLimitError,
		Message: "Concurrent execution limit exceeded",
	}

	ErrSensitiveData = &A2AError{
		Code:    SensitiveDataError,
		Message: "Sensitive data detected",
	}
)