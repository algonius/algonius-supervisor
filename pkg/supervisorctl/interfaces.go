package supervisorctl

import (
	"time"
)

// ISupervisorctlClient defines the interface for communicating with the supervisord daemon
type ISupervisorctlClient interface {
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

// AgentStatus represents the runtime state of a supervised agent
type AgentStatus struct {
	Name         string    `json:"name"`
	State        string    `json:"state"`
	PID          int       `json:"pid"`
	StartTime    time.Time `json:"start_time"`
	Duration     string    `json:"duration"`
	ExitStatus   int       `json:"exit_status"`
	RestartCount int       `json:"restart_count"`
	LastError    string    `json:"last_error,omitempty"`
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  int64     `json:"memory_usage"`
	DiskUsage    int64     `json:"disk_usage"`
}

// OperationResult represents the result of batch agent operations
type OperationResult struct {
	Operation string        `json:"operation"`
	Successes []AgentResult `json:"successes"`
	Failures  []AgentResult `json:"failures"`
	Summary   OperationSummary `json:"summary"`
}

// AgentResult represents the result of an operation on a single agent
type AgentResult struct {
	AgentName string `json:"agent_name"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}

// OperationSummary provides summary statistics for batch operations
type OperationSummary struct {
	Total     int    `json:"total"`
	Succeeded int    `json:"succeeded"`
	Failed    int    `json:"failed"`
	Duration  string `json:"duration"`
}

// LogEntry represents a single log entry from an agent
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	AgentName string    `json:"agent_name"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source,omitempty"`
}

// AgentEvent represents a lifecycle event for an agent
type AgentEvent struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	AgentName string    `json:"agent_name"`
	EventType string    `json:"event_type"`
	Details   string    `json:"details,omitempty"`
	PID       int       `json:"pid,omitempty"`
}

// ServerInfo provides information about the supervisord server
type ServerInfo struct {
	Version     string  `json:"version"`
	Uptime      string  `json:"uptime"`
	AgentCount  int     `json:"agent_count"`
	RunningAgents int   `json:"running_agents"`
	SystemLoad  []float64 `json:"system_load"`
}

// ConfigValidation represents the result of configuration validation
type ConfigValidation struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}