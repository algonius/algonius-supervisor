# Implementation Plan: AI Agent Wrapper

**Branch**: `001-ai-agent-wrapper` | **Date**: 2025-11-18 | **Spec**: [specs/001-ai-agent-wrapper/spec.md](specs/001-ai-agent-wrapper/spec.md)
**Input**: Feature specification from `/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

The AI Agent Wrapper feature provides a unified interface to wrap multiple CLI AI agents (Claude Code, Codex, Gemini CLI, etc.) with A2A endpoints. The system allows configuration of multiple agents, supports scheduled tasks for automated execution, and handles concurrent execution with different rules for read-write vs read-only agents.

## Technical Context

**Language/Version**: Go 1.23
**Primary Dependencies**: 
- github.com/gin-gonic/gin for HTTP services
- github.com/spf13/viper for configuration management
- github.com/stretchr/testify for testing
- go.uber.org/zap for logging
- github.com/pkg/errors for error handling
- github.com/a2aproject/a2a-go for A2A protocol
- github.com/modelcontextprotocol/go-sdk for MCP
**Storage**: Configuration files and execution logs (files)
**Testing**: github.com/stretchr/testify for assertions and mocking
**Target Platform**: Linux server
**Project Type**: Web service with CLI components
**Performance Goals**: A2A endpoints respond within 2 seconds, handle 10+ concurrent agent executions
**Constraints**: <200ms p95 for A2A endpoints, secure handling of sensitive data, authentication required for A2A endpoints
**Scale/Scope**: Support 5+ different CLI AI agent types, 10+ concurrent agent instances

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Compliance Check:
✓ Go 1.23 Development Standards: Using Go 1.23 as required
✓ Test-Driven Development: Tests will be written first using testify
✓ Dependency Inversion Principle: Will implement through interfaces
✓ Interface-Driven Design: Will define interfaces for agent contracts
✓ Code as Documentation: Will include comprehensive documentation

### Technology Stack Compliance:
✓ gin-gonic/gin for HTTP services (A2A endpoints)
✓ spf13/viper for configuration management
✓ testify for testing
✓ uber.org/zap for logging
✓ pkg/errors for error handling
✓ a2aproject/a2a-go for A2A protocol
✓ modelcontextprotocol/go-sdk for MCP

## Project Structure

### Documentation (this feature)

```text
specs/001-ai-agent-wrapper/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
└── supervisor/
    └── main.go              # Application entry point

internal/
├── models/
│   ├── agent_config.go      # Agent configuration models (includes working directory, env vars)
│   ├── execution_result.go  # Execution result models
│   └── scheduled_task.go    # Scheduled task models
├── services/
│   ├── agent_service.go     # Main agent service
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

**Structure Decision**: Single project structure selected following Go standard practices with cmd/, internal/, and pkg/ directories. The cmd/supervisor contains the application entry point, internal/ contains internal application code that should not be imported by other projects, and pkg/ contains reusable libraries that can be imported by other projects. The agent configurations in internal/models include working directory and environment variable settings to allow agents to run with specific execution contexts. This follows Go project layout standards and maintains testability through the interface-driven design required by the constitution.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (none)    | (none)     | (none)                              |