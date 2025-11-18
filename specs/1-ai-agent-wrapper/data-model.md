# Data Model: AI Agent Wrapper

## Overview
Entity definitions for the AI Agent Wrapper feature based on the key entities identified in the feature specification.

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
- `InputPattern` (InputPattern): Enum - "stdin", "file", "args", "json-rpc" - how the agent accepts input
- `OutputPattern` (OutputPattern): Enum - "stdout", "file", "json-rpc" - how the agent returns output
- `InputFileTemplate` (string): Template for input file when using file input pattern
- `OutputFileTemplate` (string): Template for output file when using file output pattern
- `AccessType` (AccessType): Enum value - "read-only" or "read-write"
- `MaxConcurrentExecutions` (int): Maximum number of concurrent executions allowed (1 for read-write, unlimited for read-only by default)
- `Timeout` (int): Execution timeout in seconds
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

## Service Interfaces

### Go Type Definition
**File**: `internal/services/agent_service.go`

### AgentService Interface
Interface for managing agent configurations and executions:
- `ExecuteAgent(agentID string, input string) (*ExecutionResult, error)`
- `ExecuteAgentWithParameters(agentID string, input string, parameters map[string]interface{}) (*ExecutionResult, error)`
- `ExecuteAgentWithOptions(agentID string, input string, parameters map[string]interface{}, workingDir string, envVars map[string]string) (*ExecutionResult, error)`
- `GetAgentStatus(agentID string) (*AgentStatus, error)`
- `ListAgents() ([]*AgentConfiguration, error)`

### Go Type Definition
**File**: `internal/services/scheduler_service.go`

### SchedulerService Interface
Interface for managing scheduled tasks:
- `ScheduleTask(task *ScheduledTask) error`
- `UnscheduleTask(taskID string) error`
- `ListScheduledTasks() ([]*ScheduledTask, error)`
- `ExecuteTask(taskID string) (*ExecutionResult, error)`

### Go Type Definition
**File**: `internal/services/a2a_service.go`

### A2AService Interface
Interface that complies with A2A protocol specification:
- `HandleA2ARequest(request A2ARequest) (*A2AResponse, error)`  // Handles standard A2A protocol requests
- `GetA2AStatus() (*A2AStatusResponse, error)`  // Returns standard A2A status according to spec

### A2A Protocol Types
Types that align with A2A protocol specification (https://a2a-protocol.org/latest/specification):

#### A2AMessage
**File**: `pkg/a2a/protocol.go`
Base message structure as defined in A2A spec:
- `Protocol` (string): Protocol identifier, MUST be "a2a"
- `Version` (string): Protocol version (e.g., "1.0")
- `ID` (string): Unique message identifier
- `Type` (string): Message type ("request" or "response")
- `Timestamp` (time.Time): Message timestamp in RFC 3339 format
- `InResponseTo` (string): ID of the message this is a response to (for responses only)

#### A2AContext
**File**: `pkg/a2a/protocol.go`
Context structure as defined in A2A spec:
- `From` (string): Source agent identifier
- `To` (string): Target agent identifier
- `ConversationID` (string): Conversation identifier

#### A2APayload
**File**: `pkg/a2a/protocol.go`
Payload structure as defined in A2A spec:
- `Method` (string): The method to execute (e.g., "execute-agent", "status") for requests
- `Params` (map[string]interface{}): Method parameters for requests
- `Result` (interface{}): Result of method execution for responses

## State Transitions

### ExecutionResult State Transitions
1. Created (initial state when execution starts)
2. Running (execution in progress)
3. Success/Failed/Timeout/Cancelled (final states based on execution outcome)

### ScheduledTask State Transitions
1. Created (task defined but not yet scheduled)
2. Active (task is scheduled and waiting for execution time)
3. Executing (task has triggered an agent execution)
4. Paused (task temporarily disabled)