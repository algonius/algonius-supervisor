package formatter

import (
	"fmt"
	"io"
	"strings"

	"github.com/algonius/algonius-supervisor/pkg/supervisorctl"
)

// SimpleFormatter provides simple non-tabular formatting
type SimpleFormatter struct {
	writer    io.Writer
	useColors bool
}

// NewSimpleFormatter creates a new SimpleFormatter
func NewSimpleFormatter(writer io.Writer) *SimpleFormatter {
	return &SimpleFormatter{
		writer:    writer,
		useColors: false,
	}
}

// WithColors enables or disables color output
func (sf *SimpleFormatter) WithColors(useColors bool) *SimpleFormatter {
	sf.useColors = useColors
	return sf
}

// FormatAgentStatus formats agent statuses in a simple list format
func (sf *SimpleFormatter) FormatAgentStatus(statuses []supervisorctl.AgentStatus) error {
	if len(statuses) == 0 {
		fmt.Fprintln(sf.writer, "No agents found.")
		return nil
	}

	for _, status := range statuses {
		stateColor := sf.getStateColor(status.State)
		pidStr := sf.formatPID(status.PID)

		if sf.useColors {
			fmt.Fprintf(sf.writer, "%s: %s%s (PID: %s, Uptime: %s, Restarts: %d)\n",
				status.Name,
				stateColor,
				status.State,
				pidStr,
				status.Duration,
				status.RestartCount,
			)
		} else {
			fmt.Fprintf(sf.writer, "%s: %s (PID: %s, Uptime: %s, Restarts: %d)\n",
				status.Name,
				status.State,
				pidStr,
				status.Duration,
				status.RestartCount,
			)
		}
	}

	return nil
}

// FormatOperationResult formats operation results in a simple list format
func (sf *SimpleFormatter) FormatOperationResult(result *supervisorctl.OperationResult) error {
	// Write summary
	fmt.Fprintf(sf.writer, "Operation: %s\n", result.Operation)
	fmt.Fprintf(sf.writer, "Summary: %d total, %d succeeded, %d failed\n",
		result.Summary.Total, result.Summary.Succeeded, result.Summary.Failed)
	fmt.Fprintf(sf.writer, "Duration: %s\n", result.Summary.Duration)
	fmt.Fprintln(sf.writer)

	if len(result.Successes) > 0 {
		fmt.Fprintln(sf.writer, "Successful Operations:")
		for _, success := range result.Successes {
			if success.Message == "" {
				success.Message = "Operation completed successfully"
			}
			fmt.Fprintf(sf.writer, "  %s: %s\n", success.AgentName, success.Message)
		}
		fmt.Fprintln(sf.writer)
	}

	if len(result.Failures) > 0 {
		fmt.Fprintln(sf.writer, "Failed Operations:")
		for _, failure := range result.Failures {
			if sf.useColors {
				color := getErrorColor(true)
				fmt.Fprintf(sf.writer, "  %s%s: %s\x1b[0m\n", color, failure.AgentName, failure.Error)
			} else {
				fmt.Fprintf(sf.writer, "  %s: %s\n", failure.AgentName, failure.Error)
			}
		}
	}

	return nil
}

// FormatServerInfo formats server information in a simple list format
func (sf *SimpleFormatter) FormatServerInfo(info *supervisorctl.ServerInfo) error {
	fmt.Fprintln(sf.writer, "SERVER INFORMATION")
	fmt.Fprintf(sf.writer, "Version: %s\n", info.Version)
	fmt.Fprintf(sf.writer, "Uptime: %s\n", info.Uptime)
	fmt.Fprintf(sf.writer, "Total Agents: %d\n", info.AgentCount)
	fmt.Fprintf(sf.writer, "Running Agents: %d\n", info.RunningAgents)

	if len(info.SystemLoad) >= 3 {
		fmt.Fprintf(sf.writer, "System Load: %.2f, %.2f, %.2f\n",
			info.SystemLoad[0], info.SystemLoad[1], info.SystemLoad[2])
	}

	return nil
}

// Helper methods

func (sf *SimpleFormatter) getStateColor(state string) string {
	if !sf.useColors {
		return ""
	}

	switch strings.ToUpper(state) {
	case "RUNNING":
		return "\x1b[32m" // Green
	case "STOPPED":
		return "\x1b[31m" // Red
	case "STARTING", "STOPPING":
		return "\x1b[33m" // Yellow
	case "FATAL", "FAILED":
		return "\x1b[31m" // Red
	case "BACKOFF":
		return "\x1b[33m" // Yellow
	default:
		return "\x1b[37m" // White
	}
}

func (sf *SimpleFormatter) formatPID(pid int) string {
	if pid == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", pid)
}

// getErrorColor returns ANSI color code for error messages
func getErrorColor(useColors bool) string {
	if !useColors {
		return ""
	}
	return "\x1b[31m" // Red
}