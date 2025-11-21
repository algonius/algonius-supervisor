# Implementation Plan: Add supervisorctl Control Program

**Branch**: `002-add-supervisorctl` | **Date**: 2025-11-21 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-add-supervisorctl/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Primary requirement: Create a CLI control program (supervisorctl) for managing the algonius supervisor daemon, following the traditional supervisor pattern with daemon (supervisord) and control client separation. Technical approach will leverage existing HTTP API infrastructure for communication while providing a native Go CLI interface for agent lifecycle management.

## Technical Context

**Language/Version**: Go 1.23 (aligned with existing codebase)
**Primary Dependencies**: github.com/spf13/cobra for CLI framework, existing HTTP API client, github.com/spf13/viper for configuration
**Storage**: No additional storage required - uses existing agent management infrastructure
**Testing**: github.com/stretchr/testify for CLI testing and mocking
**Target Platform**: Linux/macOS/Windows (CLI tool)
**Project Type**: CLI application with HTTP client communication
**Performance Goals**: CLI commands must complete within 2 seconds, support 1000+ agent management operations
**Constraints**: Must work with existing supervisord HTTP API, maintain backward compatibility, graceful error handling for daemon connectivity issues
**Scale/Scope**: Support management of up to 1000 concurrent agents, maintain sub-second response times for status queries

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ Go 1.23 Development Standards
- Implementation will use Go 1.23 with modern practices
- CLI application follows Go idioms and tooling standards

### ✅ Test-Driven Development (NON-NEGOTIABLE)
- All CLI functionality must have comprehensive tests using testify
- Unit tests for command parsing, integration tests for API communication
- Must achieve high code coverage before merging

### ✅ Dependency Inversion Principle
- CLI client depends on HTTP API interface abstraction
- Service interfaces already defined in existing codebase

### ✅ Interface-Driven Design
- All new services will define I-prefixed interfaces
- HTTP client communication through interface abstractions

### ✅ Code as Documentation
- CLI commands will have comprehensive godoc documentation
- Clear variable and function names following Go conventions

### ✅ Go Interface Naming Convention
- All new interfaces will use I-prefix (e.g., ISupervisorctlClient)

### ✅ Technology Stack Compliance
- Uses approved stack: Go 1.23, cobra for CLI (standard), viper for config, testify for testing
- Leverages existing dependencies: gin, zap for integration

## Post-Design Constitution Check

*✅ All constitutional requirements validated through Phase 1 design*

### Updated Validation Summary

**Architecture**: Interface-driven design with I-prefixed interfaces (ISupervisorctlClient, IHTTPClient, IConfigManager)

**Testing Strategy**: Comprehensive TDD approach with unit tests (mocking), integration tests (real supervisord), and contract tests (API compliance)

**Dependencies**: All dependencies from approved technology stack, minimal additional dependencies (cobra for CLI)

**Code Quality**: Go 1.23 standards with godoc documentation, structured error handling, and clear naming conventions

**Complexity**: Well-bounded scope extending existing architecture without unnecessary complexity

**No Constitution Violations Detected - Ready for Implementation**

## Project Structure

### Documentation (this feature)

```text
specs/002-add-supervisorctl/
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
├── supervisord/         # Renamed from cmd/supervisor/
│   └── main.go          # Supervisor daemon entry point
└── supervisorctl/       # NEW: CLI control client
    └── main.go          # CLI entry point

internal/
├── cli/                 # NEW: CLI-specific logic
│   ├── commands/        # Command implementations
│   ├── client/          # HTTP client for supervisor API
│   └── config/          # CLI configuration management
├── models/              # Existing: shared data models
├── services/            # Existing: business logic services
└── api/                 # Existing: HTTP API handlers

pkg/
├── supervisorctl/       # NEW: Public CLI interfaces
│   └── interfaces.go    # CLI service interfaces

tests/
├── unit/                # Unit tests for CLI components
├── integration/         # Integration tests with supervisord
└── contract/            # Contract tests for CLI behavior
```

**Structure Decision**: Single project structure extending existing codebase. The CLI client (supervisorctl) is added as a new cmd entry point with shared internal services for HTTP communication and configuration management. This maintains consistency with existing architecture while providing clear separation of concerns between daemon and client components.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
