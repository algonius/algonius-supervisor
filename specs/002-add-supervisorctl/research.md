# Research: supervisorctl CLI Implementation

**Date**: 2025-11-21
**Purpose**: Research best practices for implementing supervisorctl CLI tool

## CLI Framework Decision

### Selected: github.com/spf13/cobra
**Decision**: Use Cobra for CLI framework
**Rationale**:
- Excellent for supervisor-like tools with hierarchical commands (`supervisorctl status`, `supervisorctl restart service`)
- Seamless Viper integration for configuration management
- POSIX compliant flag behavior users expect
- Rich ecosystem (used by kubectl, docker CLI, gh CLI)
- Built-in auto-completion support for bash/zsh/fish

**Alternatives considered**:
- urfave/cli: Simpler but less sophisticated command hierarchy
- Standard library flag: Only suitable for very simple tools

## Configuration Management Strategy

### Approach: Viper with 12-Factor App Principles
**Decision**: Use Viper for configuration with precedence: CLI flags > env vars > config files > defaults

**Configuration Structure**:
```yaml
server:
  url: "http://localhost:8080"
  timeout: 30s
  auth:
    token: ""

defaults:
  restart_attempts: 3
  wait_time: 5s
```

**Environment Variables**: `SUPERVISOR_SERVER_URL`, `SUPERVISOR_SERVER_TIMEOUT`, `SUPERVISOR_AUTH_TOKEN`

## HTTP Client Integration Patterns

### Client Architecture
**Decision**: Structured HTTP client with interface abstraction for testing

**Key Patterns**:
- Interface-based client design for mocking
- Proper authentication header handling
- Retry logic for transient errors
- Configurable timeouts
- Structured error types with codes

**Authentication**: Leverage existing Bearer token authentication from current API

## Error Handling Strategy

### Structured Error Types
**Decision**: Implement SupervisorError type with error codes for different failure modes

**Error Categories**:
- `CONNECTION_ERROR`: Cannot reach supervisord
- `AUTH_ERROR`: Authentication failures
- `API_ERROR`: HTTP API errors
- `PROCESS_ERROR`: Agent operation failures

**Exit Codes**: Consistent exit codes for different error types (1=connection, 2=auth, 3=general)

## Command Structure

### Primary Commands (from feature spec):
- `status [agent...]` - Show agent status
- `start <agent>...` - Start agents
- `stop <agent>...` - Stop agents
- `restart <agent>...` - Restart agents
- `tail -f <agent>` - Follow agent logs
- `events` - Show lifecycle events

### Pattern Matching:
- Support `all` for all agents
- Support glob patterns like `agent-name:*`
- Individual agent names

## Testing Strategy

### Multi-Layer Testing Approach
**Decision**: Unit tests + Integration tests + Contract tests

**Unit Tests**:
- Command parsing and validation
- Client interface mocking with testify
- Table-driven tests for command variations

**Integration Tests**:
- Against running supervisord instance
- End-to-end command execution
- Real API communication

**Contract Tests**:
- Verify CLI matches API capabilities
- Test error response handling

## Performance Considerations

### Requirements Compliance:
- **2-second command response**: Implement client timeouts and connection pooling
- **1000+ agent support**: Efficient status rendering with pagination/search
- **Sub-second status queries**: Optimize API calls and response parsing

### Optimization Strategies:
- HTTP client connection reuse
- Efficient table rendering for large agent lists
- Lazy loading for detailed agent information
- Background polling for real-time features

## Integration with Existing Codebase

### Leverage Existing Infrastructure:
- Current HTTP API endpoints for agent management
- Existing authentication system (Bearer tokens)
- Current agent status models and services
- Existing configuration patterns with Viper

### Minimal Impact Approach:
- Rename cmd/supervisor to cmd/supervisord only
- Add new cmd/supervisorctl entry point
- Share internal models and service interfaces
- Extend existing API endpoints as needed

## Security Considerations

### Secure Communication:
- Use HTTPS in production environments
- Secure token storage (environment variables, config files)
- Validate all user inputs
- Sanitize error messages to avoid information disclosure

### Access Control:
- Respect existing authentication system
- No elevation of privileges beyond current API
- Audit logging for CLI operations

## Summary

All research questions resolved. The supervisorctl implementation will use:
- **Framework**: Cobra + Viper (Go ecosystem standard)
- **Communication**: HTTP client with existing API endpoints
- **Configuration**: 12-factor app pattern with environment variables
- **Testing**: Comprehensive unit + integration + contract tests
- **Performance**: Optimized for 2-second response time and 1000+ agents

The approach ensures consistency with existing codebase while following Go CLI best practices.