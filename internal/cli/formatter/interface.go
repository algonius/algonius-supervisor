package formatter

import "github.com/algonius/algonius-supervisor/pkg/supervisorctl"

// OutputFormatter defines the interface for output formatters
type OutputFormatter interface {
	FormatAgentStatus(statuses []supervisorctl.AgentStatus) error
	FormatOperationResult(result *supervisorctl.OperationResult) error
	FormatServerInfo(info *supervisorctl.ServerInfo) error
}