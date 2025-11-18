# A2A Protocol Integration Research

**Date**: 2025-11-18
**Feature**: AI Agent Wrapper
**Branch**: 001-ai-agent-wrapper

## Overview

The algonius-supervisor system integrates `github.com/a2aproject/a2a-go` library to implement A2A Protocol v0.3.0 specification with support for JSON-RPC 2.0, gRPC, and HTTP+JSON/REST transport protocols. This research document consolidates findings on integration patterns, architecture decisions, and implementation approach.

## Technology Stack Analysis

### Required Dependencies (from Constitution)
- **Go 1.23** with modern practices
- **A2A Protocol**: `github.com/a2aproject/a2a-go` library
- **Web Framework**: `github.com/gin-gonic/gin` for HTTP services
- **Configuration**: `github.com/spf13/viper`
- **Logging**: `go.uber.org/zap`
- **Testing**: `github.com/stretchr/testify` (TDD requirement)

### A2A-Go Library Capabilities
Based on the specification requirements, the a2a-go library provides:
- Core A2A protocol implementation (v0.3.0)
- Multi-transport support (JSON-RPC 2.0, gRPC, HTTP+JSON/REST)
- AgentExecutor interface for custom agent integration
- Request handler abstraction for transport-agnostic processing
- Built-in error handling and protocol compliance

## Decision: Generic Agent Adapter Pattern
**Rationale**: Implement a generic agent adapter that can work with any CLI AI agent based on configuration patterns rather than specific implementations. This allows new agents to be integrated without code changes, just configuration updates. The adapter handles different input/output patterns (stdin/stdout, file-based, JSON-RPC) based on configuration settings.

**Alternatives considered**:
- Direct process execution without abstraction
- Plugin system using separate binaries
- Using containers for each agent

## Decision: Agent Configuration with Working Directory and Environment Variables
**Rationale**: Include working directory and environment variable configuration in agent configurations to allow agents to run in specific contexts with appropriate environment settings. This is essential for agents that depend on specific file system locations or environment variables.

## Decision: Input/Output Pattern Classification for Generic Agent Support
**Rationale**: Classify CLI agents by their input/output patterns (stdin/stdout, file-based, command-line args, JSON-RPC) to enable a single generic implementation that can handle any CLI agent based on its pattern. This allows configuration-based integration of new agents without custom code.

**Alternatives considered**:
- Custom implementation for each agent type
- Plugin system with individual plugins
- Only supporting stdin/stdout agents


## Decision: A2A Protocol Implementation
**Rationale**: Use the a2aproject/a2a-go library as specified in the constitution and required by the review comment to ensure compliance with the official A2A protocol specification (https://a2a-protocol.org/latest/specification/#323-httpjsonrest-transport) and provide interoperability with other A2A-compliant systems. The implementation strictly follows the A2A specification rather than custom endpoints.

**Alternatives considered**:
- Custom protocol implementation
- Generic RPC mechanisms
- Simple HTTP endpoints without A2A standard

## Decision: Concurrency Control
**Rationale**: Implement a concurrency manager that distinguishes between read-write and read-only agents using channels and mutexes. Read-write agents will use a single execution slot while read-only agents can run concurrently.

**Alternatives considered**:
- Process-level locks
- Database-based locks
- External coordination systems

## Decision: Configuration Management
**Rationale**: Use viper for configuration management to support multiple config sources (files, environment variables, etc.) and provide flexibility for different deployment scenarios.

**Alternatives considered**:
- Simple JSON files
- Database configuration
- Command-line arguments only

## Decision: Authentication for A2A Endpoints
**Rationale**: Implement token-based authentication as required by the feature spec, with authentication handled at the API level rather than in individual agent configurations.

**Alternatives considered**:
- No authentication (not allowed per spec)
- Certificate-based authentication

## Decision: Scheduled Task Implementation
**Rationale**: Use a cron-like scheduler library that integrates with the existing service architecture and allows for dynamic task management.

**Alternatives considered**:
- External cron jobs
- Database-based scheduling
- Event-based triggers

## Decision: Logging Strategy
**Rationale**: Use structured logging with zap to capture all required information about agent executions while ensuring sensitive data is not logged.

**Alternatives considered**:
- Simple text logging
- Log aggregation services
- No comprehensive logging (not allowed per spec)

## Integration Architecture

### Core Integration Pattern
```go
// 1. Implement AgentExecutor interface from a2a-go
type IAgentExecutor interface {
    Execute(ctx context.Context, request *a2a.Message) (*a2a.Message, error)
}

// 2. Create transport-agnostic request handler
requestHandler := a2asrv.NewHandler(agentExecutor, options...)

// 3. Support multiple transport protocols
grpcHandler := a2agrpc.NewHandler(requestHandler)
jsonrpcHandler := a2ajsonrpc.NewHandler(requestHandler)
```

### Multi-Agent Routing Strategy
The system implements path-based routing for multiple agents:
```
/agents/{agentId}/v1/
├── /message:send        # Send message to agent
├── /message:stream      # Stream messages from agent
├── /tasks/{id}          # Get specific task status
├── /tasks               # List tasks
└── /.well-known/agent-card.json  # Agent discovery
```

## Implementation Components

### 1. Protocol Layer (`pkg/a2a/`)
**Decision**: Create dedicated A2A protocol package
**Rationale**: Isolates A2A-specific logic and provides clean abstraction over a2a-go library
**Components**:
- Protocol interfaces wrapping a2a-go types
- Error mapping between internal errors and A2A protocol errors
- Message conversion utilities
- Request/response validation

### 2. Service Layer (`internal/services/`)
**Decision**: Implement service layer for business logic separation
**Rationale**: Follows dependency inversion principle and enables testability
**Components**:
- **A2A Service**: Coordinates between handlers and agent executors
- **Agent Executor Service**: Implements `a2asrv.AgentExecutor` interface
- **Task Manager**: Manages task lifecycle with agent-specific isolation
- **Agent Registry**: Manages registered agents and discovery

### 3. API Layer (`internal/api/`)
**Decision**: Use Gin framework for HTTP services
**Rationale**: Required by constitution and provides robust routing/middleware
**Components**:
- HTTP handlers for REST endpoints using Gin
- gRPC handlers wrapping a2asrv.Handler
- JSON-RPC handlers wrapping a2asrv.Handler
- Authentication middleware for Bearer token validation

### 4. Agent Integration (`internal/agents/`)
**Decision**: Generic pattern-based agent adapter
**Rationale**: Enables support for any CLI AI agent without code changes
**Components**:
- Generic Agent Adapter: Pattern-based CLI agent wrapper
- Process Management: Handles concurrent execution limits
- I/O Pattern Handlers: Support for stdin/stdout, file-based, JSON-RPC patterns

## Security & Compliance

### Authentication Requirements
- **All A2A endpoints require authentication** (Bearer token)
- **No rate limiting** (unlimited requests as per spec)
- **Comprehensive logging** for all agent executions

### Sensitive Data Protection
- **System MUST NOT store or log sensitive data** from agent inputs/outputs
- **HTTPS requirement** for all A2A communications
- **Input sanitization** before logging

### Concurrency Management
- **Read-write agents**: Only 1 concurrent execution allowed
- **Read-only agents**: Multiple concurrent executions permitted
- **Task isolation**: Each agent's tasks must be isolated

## Error Handling Strategy

### A2A-Specific Error Codes
- **-32001**: Agent not found
- **-32002**: Agent execution failed
- **-32003**: Authentication required
- **-32004**: Concurrent execution limit exceeded
- **-32005**: Sensitive data detected

### JSON-RPC Standard Errors
- **-32700**: Parse error
- **-32600**: Invalid request
- **-32601**: Method not found
- **-32602**: Invalid params
- **-32603**: Internal error

## Testing Strategy

### TDD Requirements (Constitution Mandate)
- **All code must follow TDD practices**: Write tests first, ensure they fail, then implement
- **Unit tests**: For all service layer components
- **Integration tests**: For A2A protocol integration
- **End-to-end tests**: For complete agent execution flows
- **High code coverage** requirement

### Test Categories
1. **Protocol compliance tests**: Verify A2A specification adherence
2. **Transport layer tests**: Test JSON-RPC, gRPC, and HTTP endpoints
3. **Authentication tests**: Verify security requirements
4. **Concurrency tests**: Test read-write vs read-only agent limits
5. **Error handling tests**: Verify proper error responses

## Performance Considerations

### Response Time Goals
- **A2A endpoints respond within 2 seconds** for basic agent execution requests
- **Agent execution results available within 10 seconds** of completion

### Scalability Requirements
- **Support concurrent execution of at least 10 agent instances** without degradation
- **Handle unlimited requests** to A2A endpoints (no rate limiting)

## Alternatives Considered

### Alternative 1: Direct Protocol Implementation
**Rejected**: Implementing A2A protocol from scratch would be error-prone and time-consuming. Using `a2a-go` library ensures specification compliance and reduces development effort.

### Alternative 2: Single Transport Protocol
**Rejected**: Requirements specify support for JSON-RPC 2.0, gRPC, and HTTP+JSON/REST. Supporting only one would not meet the specification requirements.

### Alternative 3: Agent-Specific Endpoints
**Rejected**: Generic pattern-based approach allows supporting any CLI AI agent without code changes, which aligns with the core requirement of wrapping "all CLI AI agents."

## Implementation Priority

1. **Phase 1**: Initialize project structure and dependencies
2. **Phase 2**: Implement core A2A protocol layer
3. **Phase 3**: Create AgentExecutor implementation
4. **Phase 4**: Build HTTP handlers with Gin framework
5. **Phase 5**: Add authentication middleware
6. **Phase 6**: Implement streaming support
7. **Phase 7**: Add comprehensive testing

## Key Success Metrics

- **Protocol Compliance**: 100% adherence to A2A Protocol v0.3.0 specification
- **Transport Support**: All three protocols (JSON-RPC 2.0, gRPC, HTTP+JSON/REST) functional
- **Agent Support**: Successfully wrap at least 5 different CLI AI agents
- **Security**: All endpoints properly authenticated with Bearer tokens
- **Performance**: Response times meet specified goals
- **Test Coverage**: High coverage with all tests passing

This research provides the foundation for implementing robust A2A protocol integration using the `github.com/a2aproject/a2a-go` library while maintaining compliance with algonius-supervisor architecture and requirements.

---

# CLI Agent Lifecycle Management Research

## CLI Agent Lifecycle Management Research

### Decision: Process-Based Agent Execution with State Management
**Rationale**: CLI AI agents are external processes that need to be managed through their lifecycle states. Based on the requirements to support multiple agent types (Claude Code, Codex, Gemini CLI) with different execution patterns, a process-based approach with explicit state management provides the most flexibility and reliability.

**Alternatives considered**:
- **Container-based execution**: Would provide better isolation but adds significant complexity and overhead for CLI tools
- **In-process execution**: Not feasible since these are external CLI tools
- **Simple spawn-and-forget**: Lacks the control needed for concurrent execution limits and error handling

### Decision: State Machine Pattern for Agent Lifecycle
**Rationale**: A state machine provides clear lifecycle management with defined transitions:
- `IDLE` → `STARTING` → `RUNNING` → `COMPLETED`/`FAILED` → `CLEANUP` → `IDLE`

This pattern allows for:
- Proper handling of concurrent execution limits (read-write vs read-only agents)
- Clean error handling and recovery
- Resource management and cleanup
- Execution timeout handling

**Alternatives considered**:
- **Simple boolean flags**: Too simplistic for complex state transitions
- **Event-driven approach**: Would add unnecessary complexity for this use case

### Decision: Configuration-Driven Agent Factory Pattern
**Rationale**: The requirement to support "any command-line AI agent without code changes" necessitates a configuration-driven approach where agents are created based on configuration patterns rather than hardcoded implementations.

Key configuration aspects:
- Command pattern and arguments
- Input/output handling patterns
- Environment variables and working directory
- Execution timeouts and resource limits
- Agent type classification (read-write vs read-only)

**Alternatives considered**:
- **Plugin architecture**: Would require agents to implement specific interfaces
- **Code generation**: Would require rebuilding for new agent types
- **Hardcoded agent types**: Doesn't meet the flexibility requirement

### Decision: Separate Execution Services for Different Agent Types
**Rationale**: The requirement to distinguish between read-write (1 concurrent execution) and read-only (multiple concurrent executions) agents requires different execution strategies:

**Read-Write Agents**:
- Single execution queue with exclusive access
- Must wait for previous execution to complete
- Higher priority for resource allocation

**Read-Only Agents**:
- Multiple concurrent executions allowed
- Resource pooling for efficiency
- Can share execution contexts

**Alternatives considered**:
- **Single execution service**: Would require complex logic to handle different concurrency rules
- **Global locking mechanism**: Would be prone to deadlocks and race conditions

### Decision: Structured Logging with Execution Context
**Rationale**: The requirement for "comprehensive logging" and "MUST NOT store or log sensitive data" requires a structured approach:

- Use zap for structured logging
- Implement data sanitization before logging
- Include execution context (agent type, execution ID, timestamps)
- Separate sensitive data handling from logging

**Alternatives considered**:
- **Simple text logging**: Doesn't provide the structure needed for analysis
- **No sanitization**: Would violate the sensitive data requirement

### Decision: Graceful Error Handling with Retry Logic
**Rationale**: CLI agents can fail for various reasons (network issues, resource constraints, agent errors). A robust error handling strategy includes:

- Categorization of error types (transient vs permanent)
- Configurable retry logic for transient errors
- Proper error propagation to callers
- Resource cleanup on failure

**Alternatives considered**:
- **Fail-fast approach**: Would not handle transient failures well
- **Infinite retries**: Could lead to resource exhaustion
- **No error categorization**: Would treat all errors the same way

### Decision: Resource Management with Timeouts
**Rationale**: CLI agents may have varying execution times and resource requirements. Proper resource management includes:

- Configurable execution timeouts per agent type
- Memory and CPU limits where applicable
- Proper cleanup of temporary files and resources
- Graceful shutdown handling

**Alternatives considered**:
- **No timeouts**: Could lead to resource exhaustion from hung processes
- **Fixed timeouts**: Wouldn't accommodate different agent types
- **No resource limits**: Could impact system stability

## Implementation Patterns

### Agent Factory Pattern
The system will use a factory pattern to create agent instances based on configuration:

```go
type IAgentFactory interface {
    CreateAgent(config AgentConfig) (IAgent, error)
    GetAgentTypes() []string
}

type IAgent interface {
    Execute(ctx context.Context, input string) (ExecutionResult, error)
    GetType() string
    IsReadOnly() bool
    GetState() AgentState
}
```

### State Management
Agent states will be managed through a state machine with explicit transitions and validation:

```go
type AgentState string

const (
    StateIdle       AgentState = "IDLE"
    StateStarting   AgentState = "STARTING"
    StateRunning    AgentState = "RUNNING"
    StateCompleted  AgentState = "COMPLETED"
    StateFailed     AgentState = "FAILED"
    StateCleanup    AgentState = "CLEANUP"
)
```

### Execution Services
Separate services for read-write and read-only agents to handle concurrency requirements:

```go
type IExecutionService interface {
    ExecuteAgent(ctx context.Context, agent IAgent, input string) (ExecutionResult, error)
    GetActiveExecutions() []ExecutionInfo
    CancelExecution(executionID string) error
}

type ReadWriteExecutionService struct {
    // Single execution queue
}

type ReadOnlyExecutionService struct {
    // Concurrent execution pool
}
```

### Configuration Schema
Agent configuration will support pattern-based execution:

```yaml
agents:
  - name: "claude-code"
    type: "claude-code"
    command: "claude"
    args: ["--verbose"]
    working_dir: "/workspace"
    env_vars:
      ANTHROPIC_API_KEY: "${ANTHROPIC_API_KEY}"
    input_pattern: "stdin"
    output_pattern: "stdout"
    timeout: 300
    read_only: false
    max_concurrent: 1
```

## Security Considerations

### Sensitive Data Handling
- Sanitize inputs/outputs before logging
- Use environment variables for sensitive configuration
- Implement proper authentication for A2A endpoints
- Secure storage of execution results (avoid storing sensitive data)

### Resource Isolation
- Process isolation for agent execution
- Resource limits (CPU, memory, timeout)
- Proper cleanup of temporary resources
- Sandboxing where appropriate

## Monitoring and Observability

### Metrics
- Execution success/failure rates
- Execution duration
- Resource usage
- Queue lengths for read-write agents

### Logging
- Structured logging with execution context
- Error categorization and tracking
- Performance metrics
- Resource usage tracking