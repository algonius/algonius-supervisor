# Data Model: supervisorctl Control Program

**Date**: 2025-11-21
**Purpose**: Define entities, relationships, and validation rules for supervisorctl CLI

## Core Entities

### 1. SupervisorctlClient
**Purpose**: Main CLI client interface for communicating with supervisord daemon

**Fields**:
```go
type SupervisorctlClient interface {
    // Agent lifecycle operations
    GetStatus(agentNames ...string) ([]AgentStatus, error)
    StartAgents(agentNames ...string) (*OperationResult, error)
    StopAgents(agentNames ...string) (*OperationResult, error)
    RestartAgents(agentNames ...string) (*OperationResult, error)

    // Real-time monitoring
    TailLogs(agentName string, follow bool) (<-chan LogEntry, error)
    GetEventStream() (<-chan AgentEvent, error)

    // Configuration
    GetServerInfo() (*ServerInfo, error)
    ValidateConfig() (*ConfigValidation, error)
}
```

**Relationships**:
- Uses HTTP client to communicate with supervisord
- Manages connection state and authentication
- Handles pattern matching for agent selection

### 2. AgentStatus
**Purpose**: Runtime state information for supervised agents

**Fields**:
```go
type AgentStatus struct {
    Name         string    `json:"name" yaml:"name"`
    State        string    `json:"state" yaml:"state"` // RUNNING, STOPPED, FAILED, COMPLETED
    PID          int       `json:"pid" yaml:"pid"`
    StartTime    time.Time `json:"start_time" yaml:"start_time"`
    Duration     string    `json:"duration" yaml:"duration"`
    ExitStatus   int       `json:"exit_status" yaml:"exit_status"`
    RestartCount int       `json:"restart_count" yaml:"restart_count"`
    LastError    string    `json:"last_error,omitempty" yaml:"last_error,omitempty"`

    // Runtime metrics
    CPUUsage     float64   `json:"cpu_usage" yaml:"cpu_usage"`
    MemoryUsage  int64     `json:"memory_usage" yaml:"memory_usage"`
    DiskUsage    int64     `json:"disk_usage" yaml:"disk_usage"`
}
```

**Validation Rules**:
- Name: Required, alphanumeric with hyphens/underscores
- State: Must be one of [RUNNING, STOPPED, FAILED, COMPLETED, STARTING, STOPPING]
- PID: >= 0 (0 means not running)
- ExitStatus: 0-255 standard Unix exit codes

**State Transitions**:
```
STOPPED → STARTING → RUNNING
RUNNING → STOPPING → STOPPED
RUNNING → FAILED (on error)
STARTING → FAILED (on start failure)
```

### 3. OperationResult
**Purpose**: Results from batch agent operations

**Fields**:
```go
type OperationResult struct {
    Operation string        `json:"operation"` // start, stop, restart
    Successes []AgentResult `json:"successes"`
    Failures  []AgentResult `json:"failures"`
    Summary   OperationSummary `json:"summary"`
}

type AgentResult struct {
    AgentName string `json:"agent_name"`
    Success   bool   `json:"success"`
    Message   string `json:"message,omitempty"`
    Error     string `json:"error,omitempty"`
}

type OperationSummary struct {
    Total    int     `json:"total"`
    Succeeded int    `json:"succeeded"`
    Failed   int     `json:"failed"`
    Duration string  `json:"duration"`
}
```

### 4. LogEntry
**Purpose**: Individual log entries for tail functionality

**Fields**:
```go
type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    AgentName string    `json:"agent_name"`
    Level     string    `json:"level"` // DEBUG, INFO, WARN, ERROR
    Message   string    `json:"message"`
    Source    string    `json:"source,omitempty"` // stdout, stderr
}
```

### 5. AgentEvent
**Purpose**: Lifecycle events for real-time monitoring

**Fields**:
```go
type AgentEvent struct {
    ID        string    `json:"id"`
    Timestamp time.Time `json:"timestamp"`
    AgentName string    `json:"agent_name"`
    EventType string    `json:"event_type"` // STARTED, STOPPED, FAILED, COMPLETED
    Details   string    `json:"details,omitempty"`
    PID       int       `json:"pid,omitempty"`
}
```

**Event Types**:
- `STARTED`: Agent successfully started
- `STOPPED`: Agent stopped gracefully
- `FAILED`: Agent failed (start failure or runtime error)
- `COMPLETED`: Agent completed successfully
- `RESTARTED`: Agent was restarted

### 6. CLI Configuration
**Purpose**: Configuration for supervisorctl client

**Fields**:
```go
type CLIConfig struct {
    Server   ServerConfig   `yaml:"server" validate:"required"`
    Auth     AuthConfig     `yaml:"auth"`
    Display  DisplayConfig  `yaml:"display"`
    Defaults DefaultsConfig `yaml:"defaults"`
}

type ServerConfig struct {
    URL     string        `yaml:"url" validate:"required,url"`
    Timeout time.Duration `yaml:"timeout" validate:"min=1s,max=5m"`
}

type AuthConfig struct {
    Token string `yaml:"token"`
}

type DisplayConfig struct {
    Format      string `yaml:"format"` // table, json, yaml
    Colors      bool   `yaml:"colors"`
    RefreshRate string `yaml:"refresh_rate"`
}

type DefaultsConfig struct {
    RestartAttempts int           `yaml:"restart_attempts" validate:"min=0,max=10"`
    WaitTime       time.Duration `yaml:"wait_time" validate:"min=1s,max=5m"`
}
```

**Validation Rules**:
- Server.URL: Required, valid URL
- Server.Timeout: Between 1 second and 5 minutes
- Display.Format: One of [table, json, yaml]
- Defaults.RestartAttempts: Between 0 and 10

### 7. Pattern Matcher
**Purpose**: Agent name pattern matching for batch operations

**Fields**:
```go
type PatternMatcher struct {
    Patterns []string `json:"patterns"`
}

type MatchResult struct {
    Matched   []string `json:"matched"`
    Unmatched []string `json:"unmatched"`
    Invalid   []string `json:"invalid"`
}
```

**Pattern Types**:
- `all`: Matches all configured agents
- `name`: Exact agent name match
- `prefix:*`: Prefix match (e.g., `web:*` matches `web-server`, `web-worker`)
- `*suffix`: Suffix match (e.g., `*-service` matches `web-service`, `db-service`)

## Relationships

### Entity Relationship Diagram
```
SupervisorctlClient
    ├── Uses → HTTP Client
    ├── Manages → AgentStatus[]
    ├── Generates → OperationResult
    ├── Streams → LogEntry
    ├── Streams → AgentEvent
    └── Configured By → CLIConfig

AgentStatus
    ├── Has → AgentEvent[] (historical)
    └── Produces → LogEntry[]

OperationResult
    ├── Contains → AgentResult[]
    └── Summarized By → OperationSummary

PatternMatcher
    └── Matches → AgentStatus[]
```

### Data Flow
```
CLI Command → PatternMatcher → Filtered Agents → SupervisorctlClient → HTTP API
    ↓
OperationResult ← API Response ← Supervisord
```

## Validation Rules Summary

### Agent Names
- Required field
- 3-64 characters
- Alphanumeric, hyphens, underscores allowed
- Must start and end with alphanumeric
- Case-sensitive

### States
- Enumerated values only
- Valid transitions enforced
- Timestamp tracking for state changes

### Operations
- Batch operations limited to 100 agents per request
- Operations timeout after 30 seconds
- Rollback not supported (fire-and-forget)

### Configuration
- Server URL validation
- Timeout bounds checking
- Authentication token validation
- Format validation

## Performance Considerations

### Memory Usage
- AgentStatus: ~200 bytes per agent
- LogEntry: ~150 bytes per entry
- AgentEvent: ~100 bytes per event

### Scaling
- Supports up to 1000 agents (as per success criteria)
- Status queries optimized with pagination
- Real-time streams use bounded buffers

### Caching
- Agent status cached for 5 seconds maximum
- Configuration cached until changed
- Authentication token cached for session duration

## Security Considerations

### Data Sanitization
- All user inputs sanitized before API calls
- Error messages sanitized to prevent information disclosure
- Log entries filtered for sensitive information

### Access Control
- All operations authenticated via bearer token
- No privilege escalation beyond existing API permissions
- Operation logging for audit trails

### Error Handling
- Structured error responses with codes
- No stack traces in production output
- Graceful degradation on API failures