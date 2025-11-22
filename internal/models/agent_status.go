package models

import (
	"fmt"
	"time"
)

// AgentState represents the possible states of an agent
type AgentState string

const (
	AgentStateIdle       AgentState = "IDLE"
	AgentStateStarting   AgentState = "STARTING"
	AgentStateRunning    AgentState = "RUNNING"
	AgentStateStopping   AgentState = "STOPPING"
	AgentStateStopped    AgentState = "STOPPED"
	AgentStateCompleted  AgentState = "COMPLETED"
	AgentStateFailed     AgentState = "FAILED"
	AgentStateTimeout    AgentState = "TIMEOUT"
	AgentStateCancelled  AgentState = "CANCELLED"
	AgentStateFatal      AgentState = "FATAL"
	AgentStateUnknown    AgentState = "UNKNOWN"
)

// String returns the string representation of the agent state
func (s AgentState) String() string {
	return string(s)
}

// IsRunning returns true if the agent is in a running state
func (s AgentState) IsRunning() bool {
	return s == AgentStateRunning
}

// IsStopped returns true if the agent is in a stopped state
func (s AgentState) IsStopped() bool {
	return s == AgentStateStopped || s == AgentStateCompleted || s == AgentStateFailed || s == AgentStateTimeout || s == AgentStateCancelled || s == AgentStateFatal
}

// IsTransitioning returns true if the agent is in a transitional state
func (s AgentState) IsTransitioning() bool {
	return s == AgentStateStarting || s == AgentStateStopping
}

// ResourceUsage is already defined in resource_usage.go

// AgentStatus represents the current status of an agent
type AgentStatus struct {
	Name            string         `json:"name"`             // Agent name
	State           AgentState     `json:"state"`            // Current state
	Description     string         `json:"description"`      // Human-readable state description
	ProcessID       int            `json:"process_id"`       // Process ID (if running)
	StartTime       time.Time      `json:"start_time"`       // When the agent was started
	LastUpdateTime  time.Time      `json:"last_update_time"` // Last status update timestamp
	ExitCode        int            `json:"exit_code"`        // Exit code (if applicable)
	ExitMessage     string         `json:"exit_message"`     // Exit message (if applicable)
	RestartCount    int            `json:"restart_count"`    // Number of restarts
	ResourceUsage   *ResourceUsage `json:"resource_usage"`   // Current resource usage
	Configuration   map[string]interface{} `json:"configuration"` // Agent configuration
	Tags            []string       `json:"tags"`             // Agent tags for filtering
}

// NewAgentStatus creates a new AgentStatus with default values
func NewAgentStatus(name string) *AgentStatus {
	now := time.Now()
	return &AgentStatus{
		Name:           name,
		State:          AgentStateIdle,
		Description:    "Agent is idle",
		ProcessID:      0,
		StartTime:      time.Time{},
		LastUpdateTime: now,
		ExitCode:       0,
		ExitMessage:    "",
		RestartCount:   0,
		ResourceUsage:  nil,
		Configuration:  make(map[string]interface{}),
		Tags:           []string{},
	}
}

// GetUptime returns the duration the agent has been running
func (as *AgentStatus) GetUptime() time.Duration {
	if as.StartTime.IsZero() || !as.IsRunning() {
		return 0
	}
	return time.Since(as.StartTime)
}

// GetFormattedUptime returns a human-readable uptime string
func (as *AgentStatus) GetFormattedUptime() string {
	uptime := as.GetUptime()
	if uptime == 0 {
		return "0:00:00"
	}

	hours := int(uptime.Hours())
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60

	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}

// IsRunning returns true if the agent is currently running
func (as *AgentStatus) IsRunning() bool {
	return as.State.IsRunning()
}

// IsStopped returns true if the agent is stopped
func (as *AgentStatus) IsStopped() bool {
	return as.State.IsStopped()
}

// IsTransitioning returns true if the agent is in a transitional state
func (as *AgentStatus) IsTransitioning() bool {
	return as.State.IsTransitioning()
}

// UpdateState updates the agent state with timestamp
func (as *AgentStatus) UpdateState(newState AgentState, description string) {
	as.State = newState
	as.Description = description
	as.LastUpdateTime = time.Now()
}

// UpdateResourceUsage updates the resource usage metrics
func (as *AgentStatus) UpdateResourceUsage(cpu float64, memoryMB, diskReadMB, diskWriteMB, networkInMB, networkOutMB int64) {
	as.ResourceUsage = &ResourceUsage{
		CPUPercent:    cpu,
		MemoryMB:      memoryMB,
		DiskReadMB:    diskReadMB,
		DiskWriteMB:   diskWriteMB,
		NetworkInMB:   networkInMB,
		NetworkOutMB:  networkOutMB,
		StartTime:     time.Now().Unix(),
		MeasurementUnit: "MB",
	}
}

// IncrementRestartCount increments the restart counter
func (as *AgentStatus) IncrementRestartCount() {
	as.RestartCount++
}

// SetProcessID sets the process ID and updates state if necessary
func (as *AgentStatus) SetProcessID(pid int) {
	as.ProcessID = pid
	if pid > 0 && as.State == AgentStateStarting {
		as.UpdateState(AgentStateRunning, "Agent is running")
	}
}

// SetExitInfo sets exit information for stopped agents
func (as *AgentStatus) SetExitInfo(exitCode int, message string) {
	as.ExitCode = exitCode
	as.ExitMessage = message
	as.ProcessID = 0
}

// Clone creates a deep copy of the AgentStatus
func (as *AgentStatus) Clone() *AgentStatus {
	clone := *as

	// Deep copy resource usage
	if as.ResourceUsage != nil {
		ru := *as.ResourceUsage
		clone.ResourceUsage = &ru
	}

	// Deep copy configuration
	if as.Configuration != nil {
		clone.Configuration = make(map[string]interface{})
		for k, v := range as.Configuration {
			clone.Configuration[k] = v
		}
	}

	// Deep copy tags
	if as.Tags != nil {
		clone.Tags = make([]string, len(as.Tags))
		copy(clone.Tags, as.Tags)
	}

	return &clone
}

// AgentStatusList represents a collection of agent statuses
type AgentStatusList struct {
	Agents []AgentStatus `json:"agents"`
	Total  int           `json:"total"`
}

// NewAgentStatusList creates a new AgentStatusList
func NewAgentStatusList() *AgentStatusList {
	return &AgentStatusList{
		Agents: []AgentStatus{},
		Total:  0,
	}
}

// Add adds an agent status to the list
func (asl *AgentStatusList) Add(status AgentStatus) {
	asl.Agents = append(asl.Agents, status)
	asl.Total++
}

// FindByName finds an agent status by name
func (asl *AgentStatusList) FindByName(name string) (*AgentStatus, bool) {
	for i := range asl.Agents {
		if asl.Agents[i].Name == name {
			return &asl.Agents[i], true
		}
	}
	return nil, false
}

// FilterByState filters agents by state
func (asl *AgentStatusList) FilterByState(state AgentState) *AgentStatusList {
	filtered := NewAgentStatusList()
	for _, agent := range asl.Agents {
		if agent.State == state {
			filtered.Add(agent)
		}
	}
	return filtered
}

// FilterByTag filters agents by tag
func (asl *AgentStatusList) FilterByTag(tag string) *AgentStatusList {
	filtered := NewAgentStatusList()
	for _, agent := range asl.Agents {
		for _, agentTag := range agent.Tags {
			if agentTag == tag {
				filtered.Add(agent)
				break
			}
		}
	}
	return filtered
}

// GetRunning returns all running agents
func (asl *AgentStatusList) GetRunning() *AgentStatusList {
	return asl.FilterByState(AgentStateRunning)
}

// GetStopped returns all stopped agents
func (asl *AgentStatusList) GetStopped() *AgentStatusList {
	stopped := NewAgentStatusList()
	for _, agent := range asl.Agents {
		if agent.State.IsStopped() {
			stopped.Add(agent)
		}
	}
	return stopped
}