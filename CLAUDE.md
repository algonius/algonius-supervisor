# algonius-supervisor Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-11-18

## Active Technologies

- Go 1.23 (001-ai-agent-wrapper)
- Configuration files and execution logs (files) (001-ai-agent-wrapper)
- Web service with CLI components (001-ai-agent-wrapper)

## Project Structure

```text
cmd/
└── supervisor/
    └── main.go              # Application entry point

internal/
├── models/
│   ├── agent_config.go      # Agent configuration models (includes working directory, env vars)
│   ├── agent_execution.go   # Agent execution lifecycle models
│   ├── execution_result.go  # Execution result models
│   ├── resource_usage.go    # Resource usage models
│   └── scheduled_task.go    # Scheduled task models
├── services/
│   ├── agent_service.go     # Main agent service
│   ├── execution_service.go # Agent execution lifecycle service
│   ├── scheduler_service.go # Task scheduler
│   └── a2a_service.go       # A2A endpoint service
├── api/
│   └── handlers/
│       ├── agent_handlers.go   # Agent API handlers
│       └── a2a_handlers.go     # A2A endpoint handlers
├── agents/
│   ├── agent_interface.go     # Agent interface definition
│   ├── generic_agent.go       # Generic agent implementation based on configuration patterns
│   ├── agent_patterns.go      # Pattern-specific handlers for different input/output patterns
│   └── agent_factory.go       # Agent factory for creating agents from configuration
├── config/
│   └── config.go              # Configuration utilities
├── api/
│   └── middleware/
│       └── auth.go            # Authentication middleware
└── logging/
    └── logger.go              # Logging utilities

pkg/
├── a2a/
│   └── protocol.go            # A2A protocol interfaces and types
├── mcp/
│   └── context.go             # MCP context interfaces and types
└── types/
    └── common_types.go        # Common shared types

tests/
├── contract/
├── integration/
└── unit/

contracts/
└── a2a_endpoints.yaml       # A2A API contracts
```

## CLI Agent Lifecycle Management

### Agent Execution State Machine

Agent executions follow a strict state machine:

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

### Key Entities

- **AgentConfiguration**: Configuration for CLI AI agents with working directory, environment variables, and access type (read-only vs read-write)
- **AgentExecution**: Tracks the lifecycle state and metadata of active agent executions
- **ExecutionResult**: Final results and metadata from completed executions
- **ResourceUsage**: CPU, memory, and disk usage metrics during execution

### Concurrency Management

- **Read-Write Agents**: Only 1 concurrent execution allowed (single execution queue)
- **Read-Only Agents**: Multiple concurrent executions allowed (resource pooling)

### Error Handling

- **Transient Errors**: Temporary issues that may succeed on retry (network issues, resource constraints)
- **Permanent Errors**: Issues that will not succeed on retry (invalid configuration, missing dependencies)
- **Agent Errors**: Issues specific to the agent implementation
- **System Errors**: Issues from the algonius-supervisor system

### Security Requirements

- **Sensitive Data**: MUST NOT store or log sensitive data from agent inputs/outputs
- **Authentication**: All A2A endpoints require Bearer token authentication
- **Sanitization**: All inputs/outputs must be sanitized before logging

## Technology Stack

- **Language**: Go 1.23 with modern practices
- **Web Framework**: github.com/gin-gonic/gin for HTTP services
- **Configuration**: github.com/spf13/viper for configuration management
- **Testing**: github.com/stretchr/testify for assertions and mocking (TDD required)
- **Logging**: go.uber.org/zap for structured logging
- **Error Handling**: github.com/pkg/errors for error wrapping and stack traces
- **A2A Protocol**: github.com/a2aproject/a2a-go for agent communication
- **MCP**: github.com/modelcontextprotocol/go-sdk for context protocol implementation

## Architecture Patterns

- **Factory Pattern**: Configuration-driven agent creation
- **State Machine**: Strict execution lifecycle management
- **Interface-Driven Design**: All services defined through interfaces (I-prefix naming convention)
- **Dependency Inversion**: High-level modules depend on abstractions
- **Process-Based Execution**: CLI agents run as external processes with full lifecycle management

## Commands

# Add commands for Go 1.23

## Code Style

Go 1.23: Follow standard conventions

## Recent Changes

- 001-ai-agent-wrapper: Added Go 1.23
- 001-ai-agent-wrapper: Added CLI agent lifecycle management with state machine pattern
- 001-ai-agent-wrapper: Added execution services for read-write and read-only agents
- 001-ai-agent-wrapper: Added comprehensive error handling with retry logic
- 001-ai-agent-wrapper: Added resource usage monitoring and structured logging

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
