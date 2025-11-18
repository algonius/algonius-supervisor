# Data Model: AI Agent Wrapper

**Date**: 2025-11-18
**Feature**: AI Agent Wrapper
**Branch**: 001-ai-agent-wrapper

## Overview

This document defines the data model entities for the AI Agent Wrapper feature, including agent configurations, scheduled tasks, A2A endpoints, and execution results. The model supports the generic pattern-based approach for wrapping any CLI AI agent and integrates with the A2A protocol specification.

## Go Package Structure
The data models are organized following Go standard practices:
- **internal/models/** - Contains all model definitions (not importable by other projects)
- **pkg/types/** - Contains shared types that can be imported by other projects
- **internal/services/** - Contains service interfaces and implementations
- **pkg/a2a/** - Contains A2A protocol interfaces and types

## AgentConfiguration Entity

Represents the settings for a generic CLI AI agent, including execution patterns, input/output handling, and access type (read-only vs read-write).

### Go Type Definition
**File**: `internal/models/agent_config.go`

### Fields
- `ID` (string): Unique identifier for the agent configuration
- `Name` (string): Human-readable name for the agent
- `AgentType` (string): Type of agent based on input/output pattern (e.g., "stdin-stdout", "file-input", "json-rpc")
- `ExecutablePath` (string): Path to the CLI executable for this agent
- `WorkingDirectory` (string): Working directory for agent execution (defaults to current directory if not specified)
- `Envs` (map[string]string): Environment variables to set during agent execution
- `CliArgs` (map[string]string): Default command-line arguments for agent execution
- `Mode` (AgentMode): Enum - "task", "interactive" - determines execution behavior (single execution vs persistent session)
- `InputPattern` (InputPattern): Enum - "stdin", "file", "args", "json-rpc" - how the agent accepts input
- `OutputPattern` (OutputPattern): Enum - "stdout", "file", "json-rpc" - how the agent returns output
- `InputFileTemplate` (string): Template for input file when using file input pattern
- `OutputFileTemplate` (string): Template for output file when using file output pattern
- `AccessType` (AccessType): Enum value - "read-only" or "read-write"
- `MaxConcurrentExecutions` (int): Maximum number of concurrent executions allowed (1 for read-write, unlimited for read-only by default)
- `Timeout` (int): Execution timeout in seconds (for task mode)
- `SessionTimeout` (int): Session timeout in seconds (for interactive mode)
- `KeepAlive` (bool): Whether to maintain process alive between requests (for interactive mode)
- `Enabled` (bool): Whether the agent is currently enabled for execution

### Relationships
- One-to-many with ExecutionResult (one configuration can have many execution results)
- One-to-many with ScheduledTask (one configuration can be used by multiple scheduled tasks)

### Validation Rules
- `Name` must be unique across all configurations
- `AgentType` must be one of the supported agent types
- `AccessType` must be either "read-only" or "read-write"
- `ExecutablePath` must be a valid executable path
- `WorkingDirectory` must be a valid directory path if specified
- `Timeout` must be greater than 0


## ScheduledTask Entity

Represents an automated task that executes an agent at specified intervals or times, including timing configuration and execution context.

### Go Type Definition
**File**: `internal/models/scheduled_task.go`

### Fields
- `ID` (string): Unique identifier for the scheduled task
- `Name` (string): Human-readable name for the task
- `AgentID` (string): Reference to the agent configuration ID to execute
- `CronExpression` (string): CRON expression for scheduling
- `Enabled` (bool): Whether the task is currently enabled
- `LastExecution` (time.Time): Time of last execution
- `NextExecution` (time.Time): Time of next scheduled execution
- `InputParameters` (map[string]interface{}): Parameters to pass to the agent during execution
- `Active` (bool): Whether the task is currently active

### Relationships
- Many-to-one with AgentConfiguration (many scheduled tasks can reference one agent configuration)
- One-to-many with ExecutionResult (one scheduled task can produce many execution results)

### Validation Rules
- `CronExpression` must be a valid CRON expression
- `AgentID` must reference an existing, enabled agent configuration
- `Name` must be unique across all scheduled tasks


## AgentExecution Entity

Represents the lifecycle state and metadata of an active or recent agent execution. This entity tracks the execution state machine and provides real-time execution information.

### Go Type Definition
**File**: `internal/models/agent_execution.go`

### Fields
- `ID` (string): Unique identifier for the execution
- `AgentID` (string): Reference to the agent configuration ID being executed
- `TaskID` (string): Reference to the scheduled task ID that triggered execution (if applicable)
- `State` (AgentState): Current state of the execution (Idle, Starting, Running, Completed, Failed, Cleanup)
- `PreviousState` (AgentState): Previous state for state transition tracking
- `StartTime` (time.Time): Time when execution started
- `EndTime` (time.Time): Time when execution completed/failed (nil if still running)
- `LastStateChange` (time.Time): Time of the most recent state change
- `Input` (string): Input provided to the agent (sanitized of sensitive data)
- `ProcessID` (int): OS process ID of the executed agent (if applicable)
- `ExitCode` (int): Process exit code (if applicable)
- `ErrorMessage` (string): Error message if execution failed
- `ErrorCategory` (ErrorCategory): Category of error for retry logic
- `RetryCount` (int): Number of retry attempts made
- `MaxRetries` (int): Maximum number of retry attempts allowed
- `Timeout` (int): Execution timeout in seconds
- `ResourceUsage` (ResourceUsage): CPU and memory usage information
- `Context` (map[string]interface{}): Additional execution context

### Relationships
- Many-to-one with AgentConfiguration (many executions for one agent configuration)
- Many-to-one with ScheduledTask (many executions for one scheduled task, optional)
- One-to-one with ExecutionResult (one execution produces one result)

### Validation Rules
- `State` must be one of the defined AgentState values
- `State` transitions must follow the valid state machine rules
- `RetryCount` must be less than or equal to `MaxRetries`
- Sensitive data must be sanitized from `Input` field

### State Machine Rules
Valid state transitions:
- `Idle` → `Starting` (execution begins)
- `Starting` → `Running` (process started successfully)
- `Starting` → `Failed` (process failed to start)
- `Running` → `Completed` (execution succeeded)
- `Running` → `Failed` (execution failed)
- `Running` → `Timeout` (execution exceeded timeout)
- `Running` → `Cancelled` (execution was cancelled)
- `Completed`/`Failed`/`Timeout`/`Cancelled` → `Cleanup` (cleanup phase)
- `Cleanup` → `Idle` (cleanup completed, ready for next execution)

## ExecutionResult Entity

Represents the output, status, and metadata from an agent execution, including logs, return values, and error information.

### Go Type Definition
**File**: `internal/models/execution_result.go`

### Fields
- `ID` (string): Unique identifier for the execution result
- `AgentID` (string): Reference to the agent configuration ID that was executed
- `TaskID` (string): Reference to the scheduled task ID that triggered execution (if applicable)
- `StartTime` (time.Time): Time when execution started
- `EndTime` (time.Time): Time when execution completed/failed
- `Status` (ExecutionStatus): Enum value - "success", "failure", "timeout", "cancelled"
- `Input` (string): Input provided to the agent (sanitized of sensitive data)
- `Output` (string): Output from the agent (sanitized of sensitive data)
- `Error` (string): Error message if execution failed
- `ExecutionTime` (int64): Execution duration in milliseconds
- `ProcessID` (int): OS process ID of the executed agent (if applicable)

### Relationships
- Many-to-one with AgentConfiguration (many execution results for one agent configuration)
- Many-to-one with ScheduledTask (many execution results for one scheduled task, optional)

### Validation Rules
- `Status` must be one of the defined execution status values
- `ExecutionTime` must be greater than or equal to 0
- Sensitive data must be sanitized from `Input` and `Output` fields

## Enums and Constants

### Go Type Definition
**File**: `internal/models/enums.go`

### AgentMode
- `Task`: Agent runs in task mode - single execution with clear start/end, suitable for batch operations
- `Interactive`: Agent runs in interactive mode - persistent session allowing multiple exchanges, suitable for ongoing conversations

### AccessType
- `ReadOnly`: Agent performs only read operations, allows multiple concurrent executions
- `ReadWrite`: Agent performs read and write operations, allows only 1 concurrent execution

### InputPattern
- `Stdin`: Agent accepts input via stdin
- `File`: Agent accepts input via file path arguments
- `Args`: Agent accepts input via command-line arguments
- `JsonRpc`: Agent accepts input via JSON-RPC over stdin/stdout

### OutputPattern
- `Stdout`: Agent returns output via stdout
- `File`: Agent returns output via file path
- `JsonRpc`: Agent returns output via JSON-RPC over stdin/stdout

### ExecutionStatus
- `Success`: Execution completed successfully
- `Failure`: Execution failed due to an error
- `Timeout`: Execution exceeded timeout limit
- `Cancelled`: Execution was cancelled externally

### AgentState
- `Idle`: Agent is not currently executing
- `Starting`: Agent is being initialized for execution
- `Running`: Agent is currently executing
- `Completed`: Agent execution completed successfully
- `Failed`: Agent execution failed
- `Cleanup`: Agent is being cleaned up after execution

### ErrorCategory
- `Transient`: Temporary errors that may succeed on retry (network issues, resource constraints)
- `Permanent`: Errors that will not succeed on retry (invalid configuration, missing dependencies)
- `AgentError`: Errors specific to the agent implementation
- `SystemError`: Errors from the algonius-supervisor system

## ResourceUsage Entity

Represents resource usage information for an agent execution.

### Go Type Definition
**File**: `internal/models/resource_usage.go`

### Fields
- `CPUPercent` (float64): CPU usage percentage during execution
- `MemoryMB` (int64): Memory usage in megabytes
- `PeakMemoryMB` (int64): Peak memory usage in megabytes
- `DiskReadMB` (int64): Disk read operations in megabytes
- `DiskWriteMB` (int64): Disk write operations in megabytes

## Service Interfaces

### Go Type Definition
**File**: `internal/services/agent_service.go`

### IAgentService Interface
Interface for managing agent configurations and executions:
- `ExecuteAgent(agentID string, input string) (*ExecutionResult, error)`
- `ExecuteAgentWithParameters(agentID string, input string, parameters map[string]interface{}) (*ExecutionResult, error)`
- `ExecuteAgentWithOptions(agentID string, input string, parameters map[string]interface{}, workingDir string, envVars map[string]string) (*ExecutionResult, error)`
- `GetAgentStatus(agentID string) (*AgentStatus, error)`
- `GetAgentExecution(executionID string) (*AgentExecution, error)`
- `ListAgentExecutions(agentID string, limit int) ([]*AgentExecution, error)`
- `ListActiveExecutions() ([]*AgentExecution, error)`
- `CancelExecution(executionID string) error`
- `ListAgents() ([]*AgentConfiguration, error)`

### Go Type Definition
**File**: `internal/services/scheduler_service.go`

### ISchedulerService Interface
Interface for managing scheduled tasks:
- `ScheduleTask(task *ScheduledTask) error`
- `UnscheduleTask(taskID string) error`
- `ListScheduledTasks() ([]*ScheduledTask, error)`
- `ExecuteTask(taskID string) (*ExecutionResult, error)`

### Go Type Definition
**File**: `internal/services/a2a_service.go`

### IA2AService Interface
Interface that complies with A2A protocol specification:
- `HandleA2ARequest(request A2ARequest) (*A2AResponse, error)`  // Handles standard A2A protocol requests
- `GetA2AStatus() (*A2AStatusResponse, error)`  // Returns standard A2A status according to spec

### Go Type Definition
**File**: `internal/services/execution_service.go`

### IExecutionService Interface
Interface for managing agent execution lifecycle:
- `ExecuteAgent(ctx context.Context, agent IAgent, input string) (*AgentExecution, error)`
- `GetExecution(executionID string) (*AgentExecution, error)`
- `ListExecutions(agentID string) ([]*AgentExecution, error)`
- `CancelExecution(executionID string) error`
- `GetActiveExecutions() ([]*AgentExecution, error)`

### IReadWriteExecutionService Interface
Specialized interface for read-write agents (single concurrent execution):
- `IExecutionService` (embedded)
- `WaitForCompletion(agentID string) error` // Block until current execution completes
- `GetQueueLength(agentID string) (int, error)` // Get number of waiting executions

### IReadOnlyExecutionService Interface
Specialized interface for read-only agents (multiple concurrent executions):
- `IExecutionService` (embedded)
- `GetResourcePoolMetrics() (*ResourcePoolMetrics, error)` // Get pool utilization metrics

### ExecutionResult Extensions
The ExecutionResult entity is extended to support lifecycle management:
- `PreviousRetries` ([]ExecutionResult): References to previous retry attempts
- `ResourceUsage` (*ResourceUsage): Resource consumption during execution
- `StateTransitions` ([]StateTransition): Log of all state changes during execution

### A2A Protocol Types
Types that align with A2A protocol specification (https://a2a-protocol.org/latest/specification):

#### A2AMessage
**File**: `pkg/a2a/protocol.go`
Base message structure as defined in A2A spec v0.3.0:
- `Protocol` (string): Protocol identifier, MUST be "a2a"
- `Version` (string): Protocol version (e.g., "0.3.0")
- `ID` (string): Unique message identifier (UUID v4 recommended)
- `Type` (string): Message type ("request", "response", "error", "stream")
- `Timestamp` (time.Time): Message timestamp in RFC 3339 format
- `InResponseTo` (string): ID of the message this is a response to (for responses only)

#### A2AContext
**File**: `pkg/a2a/protocol.go`
Context structure as defined in A2A spec:
- `From` (string): Source agent identifier
- `To` (string): Target agent identifier
- `ConversationID` (string): Conversation identifier
- `MessageID` (string): Message identifier for correlation

#### A2APayload
**File**: `pkg/a2a/protocol.go`
Payload structure as defined in A2A spec:
- `Method` (string): The method to execute (e.g., "execute-agent", "status") for requests
- `Params` (map[string]interface{}): Method parameters for requests
- `Result` (interface{}): Result of method execution for responses
- `Error` (A2AError): Error information for error responses

#### A2AError
**File**: `pkg/a2a/errors.go`
Error structure following JSON-RPC 2.0 and A2A spec:
- `Code` (int): Error code (-32700 to -32005 for standard errors, -32001 to -32005 for A2A-specific)
- `Message` (string): Human-readable error message
- `Data` (interface{}): Additional error data (optional)

### A2A-Specific Data Models

#### A2AEndpoint Configuration
**File**: `internal/models/a2a_endpoint.go`
Extended A2A endpoint configuration:
- `AgentID` (string): Reference to AgentConfiguration
- `BasePath` (string): Base path for A2A endpoints (e.g., "/agents/{agentId}/v1")
- `TransportProtocols` ([]string): Supported protocols ["http_json", "grpc", "json_rpc"]
- `Authentication` (A2AAuth): Authentication configuration
- `RateLimiting` (A2ARateLimit): Rate limiting configuration
- `Capabilities` (A2ACapabilities): Agent capabilities advertisement

#### A2AAuth
**File**: `internal/models/a2a_auth.go`
Authentication configuration:
- `Type` (string): "bearer_token", "api_key", "none"
- `Token` (string): Authentication token (from environment variable)
- `HeaderName` (string): HTTP header name for token
- `Required` (bool): Whether authentication is required

#### A2ACapabilities
**File**: `internal/models/a2a_capabilities.go`
Agent capabilities for A2A protocol:
- `SupportedMethods` ([]string): Methods this agent supports
- `MaxInputSize` (int64): Maximum input size in bytes
- `MaxOutputSize` (int64): Maximum output size in bytes
- `StreamingSupport` (bool): Whether streaming is supported
- `ConcurrentExecution` (bool): Whether concurrent execution is supported
- `SupportedContentTypes` ([]string): Supported content types

## State Transitions

### AgentExecution State Machine
The AgentExecution entity follows a strict state machine to manage the agent lifecycle:

1. **Idle** → **Starting**: Execution request received, initialization begins
2. **Starting** → **Running**: Process successfully started
3. **Starting** → **Failed**: Process failed to start (invalid config, missing dependencies)
4. **Running** → **Completed**: Execution completed successfully
5. **Running** → **Failed**: Execution failed with error
6. **Running** → **Timeout**: Execution exceeded timeout limit
7. **Running** → **Cancelled**: Execution was cancelled by user/system
8. **Completed**/**Failed**/**Timeout**/**Cancelled** → **Cleanup**: Cleanup phase begins
9. **Cleanup** → **Idle**: Cleanup completed, ready for next execution

**Retry Logic**: If error is categorized as Transient and retry count < max retries:
- **Failed** → **Starting**: Retry attempt initiated

### ExecutionResult State Transitions
1. Created (initial state when execution starts)
2. Running (execution in progress)
3. Success/Failed/Timeout/Cancelled (final states based on execution outcome)

### ScheduledTask State Transitions
1. Created (task defined but not yet scheduled)
2. Active (task is scheduled and waiting for execution time)
3. Executing (task has triggered an agent execution)
4. Paused (task temporarily disabled)