package formatter

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/algonius/algonius-supervisor/pkg/supervisorctl"
)

// TableFormatter formats data as tables for CLI output
type TableFormatter struct {
	writer    io.Writer
	minWidth  int
	tabWidth  int
	padding   int
	padChar   byte
	flags     uint
	useColors bool
}

// NewTableFormatter creates a new TableFormatter
func NewTableFormatter(writer io.Writer) *TableFormatter {
	return &TableFormatter{
		writer:    writer,
		minWidth:  0,
		tabWidth:  8,
		padding:   1,
		padChar:   ' ',
		flags:     0,
		useColors: false,
	}
}

// WithColors enables or disables color output
func (tf *TableFormatter) WithColors(useColors bool) *TableFormatter {
	tf.useColors = useColors
	return tf
}

// WithTabWriterSettings configures the underlying tabwriter
func (tf *TableFormatter) WithTabWriterSettings(minWidth, tabWidth, padding int, padChar byte, flags uint) *TableFormatter {
	tf.minWidth = minWidth
	tf.tabWidth = tabWidth
	tf.padding = padding
	tf.padChar = padChar
	tf.flags = flags
	return tf
}

// FormatAgentStatus formats a list of agent statuses as a table
func (tf *TableFormatter) FormatAgentStatus(statuses []supervisorctl.AgentStatus) error {
	if len(statuses) == 0 {
		fmt.Fprintln(tf.writer, "No agents found.")
		return nil
	}

	// Create tabwriter
	w := tabwriter.NewWriter(tf.writer, tf.minWidth, tf.tabWidth, tf.padding, tf.padChar, tf.flags)

	// Write header
	header := "NAME\tSTATE\tPID\tUPTIME\tRESTARTS\tCPU\tMEMORY"
	fmt.Fprintln(w, header)

	// Write separator
	separator := strings.Repeat("-", strings.Count(header, "\t")+20)
	fmt.Fprintln(w, separator)

	// Write agent data
	for _, status := range statuses {
		stateColor := tf.getStateColor(status.State)
		pidStr := tf.formatPID(status.PID)
		cpuStr := tf.formatCPU(status.CPUUsage)
		memoryStr := tf.formatMemory(status.MemoryUsage)

		if tf.useColors {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
				status.Name,
				stateColor+status.State+"\x1b[0m",
				pidStr,
				status.Duration,
				status.RestartCount,
				cpuStr,
				memoryStr,
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
				status.Name,
				status.State,
				pidStr,
				status.Duration,
				status.RestartCount,
				cpuStr,
				memoryStr,
			)
		}
	}

	return w.Flush()
}

// FormatOperationResult formats operation results as a table
func (tf *TableFormatter) FormatOperationResult(result *supervisorctl.OperationResult) error {
	w := tabwriter.NewWriter(tf.writer, tf.minWidth, tf.tabWidth, tf.padding, tf.padChar, tf.flags)

	// Write summary
	fmt.Fprintf(w, "Operation: %s\n", result.Operation)
	fmt.Fprintf(w, "Summary: %d total, %d succeeded, %d failed\n",
		result.Summary.Total, result.Summary.Succeeded, result.Summary.Failed)
	fmt.Fprintf(w, "Duration: %s\n", result.Summary.Duration)
	fmt.Fprintln(w)

	if len(result.Successes) > 0 {
		fmt.Fprintln(w, "Successful Operations:")
		fmt.Fprintln(w, "NAME\tMESSAGE")
		for _, success := range result.Successes {
			if success.Message == "" {
				success.Message = "Operation completed successfully"
			}
			fmt.Fprintf(w, "%s\t%s\n", success.AgentName, success.Message)
		}
		fmt.Fprintln(w)
	}

	if len(result.Failures) > 0 {
		fmt.Fprintln(w, "Failed Operations:")
		fmt.Fprintln(w, "NAME\tERROR")
		for _, failure := range result.Failures {
			fmt.Fprintf(w, "%s\t%s\n", failure.AgentName, failure.Error)
		}
	}

	return w.Flush()
}

// FormatServerInfo formats server information as a table
func (tf *TableFormatter) FormatServerInfo(info *supervisorctl.ServerInfo) error {
	w := tabwriter.NewWriter(tf.writer, tf.minWidth, tf.tabWidth, tf.padding, tf.padChar, tf.flags)

	fmt.Fprintln(w, "SERVER INFORMATION")
	fmt.Fprintln(w, "Property\tValue")
	fmt.Fprintln(w, "--------\t-----")
	fmt.Fprintf(w, "Version\t%s\n", info.Version)
	fmt.Fprintf(w, "Uptime\t%s\n", info.Uptime)
	fmt.Fprintf(w, "Total Agents\t%d\n", info.AgentCount)
	fmt.Fprintf(w, "Running Agents\t%d\n", info.RunningAgents)

	if len(info.SystemLoad) >= 3 {
		fmt.Fprintf(w, "System Load\t%.2f, %.2f, %.2f\n",
			info.SystemLoad[0], info.SystemLoad[1], info.SystemLoad[2])
	}

	return w.Flush()
}

// Helper methods

func (tf *TableFormatter) getStateColor(state string) string {
	if !tf.useColors {
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

func (tf *TableFormatter) formatPID(pid int) string {
	if pid == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", pid)
}

func (tf *TableFormatter) formatCPU(cpu float64) string {
	if cpu == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", cpu)
}

func (tf *TableFormatter) formatMemory(memory int64) string {
	if memory == 0 {
		return "-"
	}

	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case memory >= GB:
		return fmt.Sprintf("%.1fGB", float64(memory)/GB)
	case memory >= MB:
		return fmt.Sprintf("%.1fMB", float64(memory)/MB)
	case memory >= KB:
		return fmt.Sprintf("%.1fKB", float64(memory)/KB)
	default:
		return fmt.Sprintf("%dB", memory)
	}
}

