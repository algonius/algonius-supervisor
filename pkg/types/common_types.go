package types

// Common shared types for the algonius-supervisor project

// TaskStatus represents the status of a scheduled task
type TaskStatus string

const (
	TaskStatusCreated  TaskStatus = "created"
	TaskStatusActive   TaskStatus = "active"
	TaskStatusExecuting TaskStatus = "executing"
	TaskStatusPaused   TaskStatus = "paused"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed   TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// ExecutionStatus represents the status of an agent execution
type ExecutionStatus string

const (
	// SuccessStatus execution completed successfully
	SuccessStatus ExecutionStatus = "success"
	// FailureStatus execution failed due to an error
	FailureStatus ExecutionStatus = "failure"
	// TimeoutStatus execution exceeded timeout limit
	TimeoutStatus ExecutionStatus = "timeout"
	// CancelledStatus execution was cancelled externally
	CancelledStatus ExecutionStatus = "cancelled"
)

// ResourceType represents the type of resource being managed
type ResourceType string

const (
	ResourceTypeAgent  ResourceType = "agent"
	ResourceTypeTask   ResourceType = "task"
	ResourceTypeResult ResourceType = "result"
)

// ResourceTypeUsage represents resource usage for an execution
type ResourceUsage struct {
	CPUPercent   float64 `json:"cpu_percent"`
	MemoryMB     int64   `json:"memory_mb"`
	PeakMemoryMB int64   `json:"peak_memory_mb"`
	DiskReadMB   int64   `json:"disk_read_mb"`
	DiskWriteMB  int64   `json:"disk_write_mb"`
	NetworkInMB  int64   `json:"network_in_mb"`
	NetworkOutMB int64   `json:"network_out_mb"`
}

// TaskTriggerType represents how a task was triggered
type TaskTriggerType string

const (
	TaskTriggerTypeScheduled TaskTriggerType = "scheduled"
	TaskTriggerTypeManual    TaskTriggerType = "manual"
	TaskTriggerTypeAPI       TaskTriggerType = "api"
	TaskTriggerTypeEvent     TaskTriggerType = "event"
)

// AgentAccessType defines whether an agent performs read-only or read-write operations
type AgentAccessType string

const (
	// ReadOnlyAccessType allows multiple concurrent executions
	ReadOnlyAccessType AgentAccessType = "read-only"
	
	// ReadWriteAccessType restricts to single concurrent execution
	ReadWriteAccessType AgentAccessType = "read-write"
)

// AgentMode defines whether an agent operates in task mode or interactive mode
type AgentMode string

const (
	// TaskMode: Agent runs in task mode - single execution with clear start/end, suitable for batch operations
	TaskMode AgentMode = "task"
	
	// InteractiveMode: Agent runs in interactive mode - persistent session allowing multiple exchanges
	InteractiveMode AgentMode = "interactive"
)

// InputPattern defines how the agent accepts input
type InputPattern string

const (
	// StdinPattern: Agent accepts input via stdin
	StdinPattern InputPattern = "stdin"
	
	// FilePattern: Agent accepts input via file path arguments
	FilePattern InputPattern = "file"
	
	// ArgsPattern: Agent accepts input via command-line arguments
	ArgsPattern InputPattern = "args"
	
	// JsonRpcPattern: Agent accepts input via JSON-RPC over stdin/stdout
	JsonRpcPattern InputPattern = "json-rpc"
)

// OutputPattern defines how the agent returns output
type OutputPattern string

const (
	// StdoutPattern: Agent returns output via stdout
	StdoutPattern OutputPattern = "stdout"

	// FilePatternOut: Agent returns output via file path
	FilePatternOut OutputPattern = "file"

	// JsonRpcPatternOut: Agent returns output via JSON-RPC over stdin/stdout
	JsonRpcPatternOut OutputPattern = "json-rpc"
)

// AgentState represents the lifecycle state of an agent execution
type AgentState string

const (
	// IdleState: Agent is not currently executing
	IdleState AgentState = "idle"

	// StartingState: Agent is being initialized for execution
	StartingState AgentState = "starting"

	// RunningState: Agent is currently executing
	RunningState AgentState = "running"

	// CompletedState: Agent execution completed successfully
	CompletedState AgentState = "completed"

	// FailedState: Agent execution failed
	FailedState AgentState = "failed"

	// TimeoutState: Agent execution timed out
	TimeoutState AgentState = "timeout"

	// CancelledState: Agent execution was cancelled externally
	CancelledState AgentState = "cancelled"

	// CleanupState: Agent is being cleaned up after execution
	CleanupState AgentState = "cleanup"
)

// ErrorCategory represents the category of an error for retry logic
type ErrorCategory string

const (
	// Transient: Temporary errors that may succeed on retry (network issues, resource constraints)
	Transient ErrorCategory = "transient"
	
	// Permanent: Errors that will not succeed on retry (invalid configuration, missing dependencies)
	Permanent ErrorCategory = "permanent"
	
	// AgentError: Errors specific to the agent implementation
	AgentError ErrorCategory = "agent"
	
	// SystemError: Errors from the algonius-supervisor system
	SystemError ErrorCategory = "system"
)

// A2ATransportProtocol defines the protocol used for A2A communication
type A2ATransportProtocol string

const (
	// HTTPJSON: HTTP+JSON transport protocol
	HTTPJSON A2ATransportProtocol = "http_json"
	
	// GRPC: gRPC transport protocol
	GRPC A2ATransportProtocol = "grpc"
	
	// JSONRPC: JSON-RPC 2.0 transport protocol
	JSONRPC A2ATransportProtocol = "json_rpc"
)