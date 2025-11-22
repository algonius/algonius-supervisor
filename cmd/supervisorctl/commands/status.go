package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/algonius/algonius-supervisor/internal/cli/client"
	"github.com/algonius/algonius-supervisor/internal/cli/config"
	"github.com/algonius/algonius-supervisor/internal/cli/formatter"
	"github.com/algonius/algonius-supervisor/internal/cli/patterns"
	"github.com/algonius/algonius-supervisor/internal/logging"
	"github.com/algonius/algonius-supervisor/pkg/supervisorctl"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// StatusCmd represents the status command
var StatusCmd = &cobra.Command{
	Use:   "status [agent-name...]",
	Short: "Get the status of agents",
	Long: `Display the status of one or more agents.
If no agent names are provided, shows the status of all agents.

Examples:
  supervisorctl status                    # Show all agents
  supervisorctl status web-server        # Show specific agent
  supervisorctl status prefix:web-       # Show all agents starting with 'web-'
  supervisorctl status suffix:-worker    # Show all agents ending with '-worker'
  supervisorctl status "regex:web-.*"     # Show agents matching regex pattern
  supervisorctl status web-* db-*         # Show agents matching wildcard patterns

Supported pattern types:
  exact:     Exact match (default)
  prefix:    Prefix match
  suffix:    Suffix match
  contains:  Contains match
  regex:     Regular expression match
  wildcard:  Wildcard pattern with * and ?`,

	RunE: runStatus,
}

// statusOptions contains options for the status command
type statusOptions struct {
	format  string
	verbose bool
	colors  bool
}

// runStatus executes the status command
func runStatus(cmd *cobra.Command, args []string) error {
	logger := logging.GetLogger()
	logger.Info("Executing status command", zap.Strings("args", args))

	// Parse command options
	opts, err := parseStatusOptions(cmd)
	if err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Get configuration manager from viper (dependency injection)
	configManager, ok := cmd.Context().Value(config.ConfigManagerKey{}).(config.IConfigManager)
	if !ok {
		return fmt.Errorf("configuration manager not found in command context")
	}

	// Load configuration
	config, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := configManager.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Create HTTP client
	httpClient := client.NewHTTPClient(config.Server.Timeout)

	// Create supervisor client
	var supervisorClient supervisorctl.ISupervisorctlClient = client.NewSupervisorClient(
		httpClient,
		config.Server.URL,
		config.Auth.Token,
		config.Server.Timeout,
	)

	// Parse agent names/patterns
	var agentNames []string
	if len(args) > 0 {
		// Validate and collect agent names/patterns
		for _, arg := range args {
			if err := patterns.ValidatePatternString(arg); err != nil {
				return fmt.Errorf("invalid pattern '%s': %w", arg, err)
			}
		}
		agentNames = args
	}

	// Get agent statuses
	statuses, err := supervisorClient.GetStatus(agentNames...)
	if err != nil {
		return fmt.Errorf("failed to get agent status: %w", err)
	}

	// Filter by patterns if provided
	if len(args) > 0 {
		statuses, err = filterStatusesByPatterns(statuses, args)
		if err != nil {
			return fmt.Errorf("failed to filter agent statuses: %w", err)
		}
	}

	// Format and display output
	if err := displayStatusOutput(statuses, opts, config.Display.Colors && opts.colors); err != nil {
		return fmt.Errorf("failed to display status output: %w", err)
	}

	logger.Info("Status command completed successfully", zap.Int("agent_count", len(statuses)))
	return nil
}

// parseStatusOptions parses command options from the cobra command
func parseStatusOptions(cmd *cobra.Command) (*statusOptions, error) {
	// Get viper instance from root command
	viperInstance := GetViperInstance()

	opts := &statusOptions{
		format:  viperInstance.GetString("display.format"),
		colors:  viperInstance.GetBool("display.colors"),
		verbose: viperInstance.GetBool("verbose"),
	}

	// Override with command-line flags if provided
	if format, err := cmd.Flags().GetString("format"); err == nil && format != "" {
		opts.format = format
	}
	if verbose, err := cmd.Flags().GetBool("verbose"); err == nil {
		opts.verbose = verbose
	}
	if colors, err := cmd.Flags().GetBool("colors"); err == nil {
		opts.colors = colors
	}

	// Validate format
	switch opts.format {
	case "table", "simple", "json":
		// Valid formats
	default:
		return nil, fmt.Errorf("invalid format '%s'. Supported formats: table, simple, json", opts.format)
	}

	return opts, nil
}

// filterStatusesByPatterns filters agent statuses by name patterns
func filterStatusesByPatterns(statuses []supervisorctl.AgentStatus, patternStrings []string) ([]supervisorctl.AgentStatus, error) {
	if len(patternStrings) == 0 {
		return statuses, nil
	}

	// Create matcher for patterns
	matcher, err := patterns.NewMatcher(patternStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to create pattern matcher: %w", err)
	}

	var filtered []supervisorctl.AgentStatus
	for _, status := range statuses {
		if matcher.Matches(status.Name) {
			filtered = append(filtered, status)
		}
	}

	return filtered, nil
}

// displayStatusOutput formats and displays the status output
func displayStatusOutput(statuses []supervisorctl.AgentStatus, opts *statusOptions, useColors bool) error {
	// Add summary if verbose
	if opts.verbose && len(statuses) > 0 {
		printStatusSummary(statuses, useColors)
	}

	// Format the output
	switch opts.format {
	case "table":
		tableFormatter := formatter.NewTableFormatter(os.Stdout).
			WithColors(useColors).
			WithTabWriterSettings(0, 8, 1, ' ', 0)
		return tableFormatter.FormatAgentStatus(statuses)
	case "simple":
		simpleFormatter := formatter.NewSimpleFormatter(os.Stdout).
			WithColors(useColors)
		return simpleFormatter.FormatAgentStatus(statuses)
	case "json":
		jsonFormatter := formatter.NewJSONFormatter(os.Stdout)
		return jsonFormatter.FormatAgentStatus(statuses)
	default:
		return fmt.Errorf("unsupported format: %s", opts.format)
	}
}

// printStatusSummary prints a summary of agent statuses
func printStatusSummary(statuses []supervisorctl.AgentStatus, useColors bool) {
	stateCounts := make(map[string]int)
	runningCount := 0

	for _, status := range statuses {
		state := strings.ToUpper(status.State)
		stateCounts[state]++
		if status.State == "RUNNING" {
			runningCount++
		}
	}

	fmt.Printf("Summary: %d total agents, %d running\n", len(statuses), runningCount)

	if len(stateCounts) > 1 {
		fmt.Printf("States: ")
		var stateList []string
		for state, count := range stateCounts {
			color := getStateColor(state, useColors)
			stateList = append(stateList, fmt.Sprintf("%s%s (%d)\x1b[0m", color, state, count))
		}
		fmt.Println(strings.Join(stateList, ", "))
	}
	fmt.Println()
}

// getStateColor returns ANSI color code for a given state
func getStateColor(state string, useColors bool) string {
	if !useColors {
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


func init() {
	// Add command flags
	StatusCmd.Flags().StringP("format", "f", "", "Output format (table, simple, json)")
	StatusCmd.Flags().BoolP("verbose", "v", false, "Show verbose output with summary")
	StatusCmd.Flags().BoolP("colors", "c", false, "Force colored output")

	// Note: Flag binding moved to root.go to use local viper instance
}