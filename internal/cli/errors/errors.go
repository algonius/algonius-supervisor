package errors

import (
	"fmt"
)

// SupervisorError represents a structured error from supervisor operations
type SupervisorError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *SupervisorError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewSupervisorError creates a new supervisor error
func NewSupervisorError(code, message string, details interface{}) *SupervisorError {
	return &SupervisorError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Common error codes
const (
	ErrorCodeConnection    = "CONNECTION_ERROR"
	ErrorCodeAuth          = "AUTH_ERROR"
	ErrorCodeAPI           = "API_ERROR"
	ErrorCodeProcess       = "PROCESS_ERROR"
	ErrorCodeValidation    = "VALIDATION_ERROR"
	ErrorCodeTimeout       = "TIMEOUT_ERROR"
	ErrorCodeNotFound      = "NOT_FOUND"
	ErrorCodeAlreadyExists = "ALREADY_EXISTS"
)

// ConnectionError creates a connection error
func ConnectionError(message string, details interface{}) *SupervisorError {
	return NewSupervisorError(ErrorCodeConnection, message, details)
}

// AuthError creates an authentication error
func AuthError(message string, details interface{}) *SupervisorError {
	return NewSupervisorError(ErrorCodeAuth, message, details)
}

// APIError creates an API error
func APIError(message string, details interface{}) *SupervisorError {
	return NewSupervisorError(ErrorCodeAPI, message, details)
}

// ProcessError creates a process operation error
func ProcessError(message string, details interface{}) *SupervisorError {
	return NewSupervisorError(ErrorCodeProcess, message, details)
}

// ValidationError creates a validation error
func ValidationError(message string, details interface{}) *SupervisorError {
	return NewSupervisorError(ErrorCodeValidation, message, details)
}

// TimeoutError creates a timeout error
func TimeoutError(message string, details interface{}) *SupervisorError {
	return NewSupervisorError(ErrorCodeTimeout, message, details)
}

// NotFoundError creates a not found error
func NotFoundError(message string, details interface{}) *SupervisorError {
	return NewSupervisorError(ErrorCodeNotFound, message, details)
}

// AlreadyExistsError creates an already exists error
func AlreadyExistsError(message string, details interface{}) *SupervisorError {
	return NewSupervisorError(ErrorCodeAlreadyExists, message, details)
}