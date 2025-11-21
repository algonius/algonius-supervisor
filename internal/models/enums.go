package models

// This file contains internal constants and enums for the models package
// For shared types between packages, see pkg/types/common_types.go

// Constants for agent modes
const (
	// TaskMode represents single execution with clear start/end, suitable for batch operations
	TaskMode = "task"
	// InteractiveMode represents persistent session allowing multiple exchanges, suitable for ongoing conversations
	InteractiveMode = "interactive"
)

// Constants for access types
const (
	// ReadOnlyAccessType agent performs only read operations, allows multiple concurrent executions
	ReadOnlyAccessType = "read-only"
	// ReadWriteAccessType agent performs read and write operations, allows only 1 concurrent execution
	ReadWriteAccessType = "read-write"
)

// Constants for input patterns
const (
	// StdinPattern agent accepts input via stdin
	StdinPattern = "stdin"
	// FilePattern agent accepts input via file path arguments
	FilePattern = "file"
	// ArgsPattern agent accepts input via command-line arguments
	ArgsPattern = "args"
	// JsonRpcPattern agent accepts input via JSON-RPC over stdin/stdout
	JsonRpcPattern = "json-rpc"
)

// Constants for output patterns
const (
	// StdoutPattern agent returns output via stdout
	StdoutPattern = "stdout"
	// FilePatternOut agent returns output via file path
	FilePatternOut = "file"
	// JsonRpcPatternOut agent returns output via JSON-RPC over stdin/stdout
	JsonRpcPatternOut = "json-rpc"
)

// Constants for execution statuses
const (
	// SuccessStatus execution completed successfully
	SuccessStatus = "success"
	// FailureStatus execution failed due to an error
	FailureStatus = "failure"
	// TimeoutStatus execution exceeded timeout limit
	TimeoutStatus = "timeout"
	// CancelledStatus execution was cancelled externally
	CancelledStatus = "cancelled"
)

// Constants for agent states
const (
	// IdleState agent is not currently executing
	IdleState = "idle"
	// StartingState agent is being initialized for execution
	StartingState = "starting"
	// RunningState agent is currently executing
	RunningState = "running"
	// CompletedState agent execution completed successfully
	CompletedState = "completed"
	// FailedState agent execution failed
	FailedState = "failed"
	// CleanupState agent is being cleaned up after execution
	CleanupState = "cleanup"
	// TimeoutState agent execution exceeded timeout limit
	TimeoutState = "timeout"
	// CancelledState agent execution was cancelled externally
	CancelledState = "cancelled"
)

// Constants for error categories
const (
	// TransientError temporary errors that may succeed on retry (network issues, resource constraints)
	TransientError = "transient"
	// PermanentError errors that will not succeed on retry (invalid configuration, missing dependencies)
	PermanentError = "permanent"
	// AgentError errors specific to the agent implementation
	AgentError = "agent_error"
	// SystemError errors from the algonius-supervisor system
	SystemError = "system_error"
)

// A2AProtocolVersion represents the version of the A2A protocol
type A2AProtocolVersion string

const (
	// A2AV030 represents version 0.3.0 of the A2A protocol
	A2AV030 A2AProtocolVersion = "0.3.0"
)

// A2AMessageType represents the type of A2A message
type A2AMessageType string

const (
	// A2ARequest represents an A2A request message
	A2ARequest A2AMessageType = "request"
	// A2AResponse represents an A2A response message
	A2AResponse A2AMessageType = "response"
	// A2AError represents an A2A error message
	A2AError A2AMessageType = "error"
	// A2AStream represents an A2A stream message
	A2AStream A2AMessageType = "stream"
)

// TransportProtocol represents supported transport protocols
type TransportProtocol string

const (
	// HTTPJSON represents HTTP with JSON transport
	HTTPJSON TransportProtocol = "http_json"
	// GRPC represents gRPC transport
	GRPC TransportProtocol = "grpc"
	// JSONRPC represents JSON-RPC transport
	JSONRPC TransportProtocol = "json_rpc"
)

// A2AErrorCodes represents standard A2A error codes
type A2AErrorCodes int

const (
	// A2AErrorCodeParseError represents a parse error (-32700)
	A2AErrorCodeParseError A2AErrorCodes = -32700
	// A2AErrorCodeInvalidRequest represents an invalid request (-32600)
	A2AErrorCodeInvalidRequest A2AErrorCodes = -32600
	// A2AErrorCodeMethodNotFound represents method not found (-32601)
	A2AErrorCodeMethodNotFound A2AErrorCodes = -32601
	// A2AErrorCodeInvalidParams represents invalid parameters (-32602)
	A2AErrorCodeInvalidParams A2AErrorCodes = -32602
	// A2AErrorCodeInternalError represents an internal error (-32603)
	A2AErrorCodeInternalError A2AErrorCodes = -32603
	// A2AErrorCodeAgentNotFound represents agent not found (-32001)
	A2AErrorCodeAgentNotFound A2AErrorCodes = -32001
	// A2AErrorCodeAgentExecutionFailed represents agent execution failed (-32002)
	A2AErrorCodeAgentExecutionFailed A2AErrorCodes = -32002
	// A2AErrorCodeAuthRequired represents authentication required (-32003)
	A2AErrorCodeAuthRequired A2AErrorCodes = -32003
	// A2AErrorCodeConcurrentLimitExceeded represents concurrent execution limit exceeded (-32004)
	A2AErrorCodeConcurrentLimitExceeded A2AErrorCodes = -32004
	// A2AErrorCodeSensitiveDataDetected represents sensitive data detected (-32005)
	A2AErrorCodeSensitiveDataDetected A2AErrorCodes = -32005
)