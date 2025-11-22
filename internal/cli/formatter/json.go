package formatter

import (
	"encoding/json"
	"io"

	"github.com/algonius/algonius-supervisor/pkg/supervisorctl"
)

// JSONFormatter provides JSON output formatting
type JSONFormatter struct {
	writer io.Writer
}

// NewJSONFormatter creates a new JSONFormatter
func NewJSONFormatter(writer io.Writer) *JSONFormatter {
	return &JSONFormatter{
		writer: writer,
	}
}

// FormatAgentStatus formats agent statuses as JSON
func (jf *JSONFormatter) FormatAgentStatus(statuses []supervisorctl.AgentStatus) error {
	encoder := json.NewEncoder(jf.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(statuses)
}

// FormatOperationResult formats operation results as JSON
func (jf *JSONFormatter) FormatOperationResult(result *supervisorctl.OperationResult) error {
	encoder := json.NewEncoder(jf.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// FormatServerInfo formats server information as JSON
func (jf *JSONFormatter) FormatServerInfo(info *supervisorctl.ServerInfo) error {
	encoder := json.NewEncoder(jf.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}