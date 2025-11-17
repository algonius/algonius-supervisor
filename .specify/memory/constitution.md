<!--
Sync Impact Report:
- Version change: 1.1.1 → 1.2.0 (added MCP and A2A Protocol libraries)
- Modified sections: Technology Stack section updated to include MCP and A2A Protocol
- Added sections: None
- Removed sections: None
- Templates requiring updates: ✅ plan-template.md, spec-template.md, tasks-template.md
- No follow-up TODOs (all placeholders resolved)
-->
# algonius-supervisor Constitution

## Core Principles

### I. Go 1.23 Development Standards
All code must be written using Go 1.23 with modern Go practices. The codebase must follow Go idioms, standard formatting (go fmt), and pass go vet and static analysis tools. All new features should leverage Go 1.23's latest capabilities and performance improvements.

### II. Test-Driven Development (NON-NEGOTIABLE)
All code must follow TDD practices: write tests first using github.com/stretchr/testify for assertions and mocking, ensure they fail, then implement functionality. Unit tests, integration tests, and end-to-end tests must be written for all features. Code coverage should remain high, and all tests must pass before merging.

### III. Dependency Inversion Principle
All modules must depend on abstractions rather than concretions. Interfaces should be defined at high levels and implemented at lower levels. This enables better testability, maintainability, and allows for proper mocking in tests.

### IV. Interface-Driven Design
System design should start with well-defined interfaces that specify contracts between components. All public APIs should be interface-based to enable flexibility, testability, and loose coupling between system components.

### V. Code as Documentation
Code must be self-explanatory with clear variable names, function names, and package structures. Comprehensive comments should explain the "why" when the "what" isn't obvious. All public functions and types must have godoc-compliant documentation.

## Technology Stack

All development must use the following technology stack:
- Configuration: github.com/spf13/viper for configuration management
- Testing: github.com/stretchr/testify for assertions and mocking
- Environment Variables: github.com/subosito/gotenv for environment configuration
- Logging: go.uber.org/zap for structured logging
- Error Handling: github.com/pkg/errors for error wrapping and stack traces
- Web Framework: github.com/gin-gonic/gin for HTTP services
- HTTP Assertions: github.com/gavv/httpexpect for HTTP testing and assertions
- MCP (Model Context Protocol): github.com/modelcontextprotocol/go-sdk for context protocol implementation
- A2A (Agent-to-Agent Protocol): github.com/a2aproject/a2a-go for agent communication

## Development Workflow

All contributions require code review, passing tests, and adherence to Go coding standards. Features must include documentation and examples. Breaking changes require deprecation periods and migration paths. Code reviews must verify compliance with all constitution principles.

## Governance

This constitution governs all development and operations of the algonius-supervisor project. All feature specifications and implementation plans must validate against these principles. Amendments require documentation of impact assessment and approval from maintainers.

**Version**: 1.2.0 | **Ratified**: 2025-11-18 | **Last Amended**: 2025-11-18
