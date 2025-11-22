package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/algonius/algonius-supervisor/internal/cli/client"
	"github.com/algonius/algonius-supervisor/internal/cli/config"
	"github.com/algonius/algonius-supervisor/internal/cli/formatter"
	"github.com/algonius/algonius-supervisor/internal/cli/patterns"
	"github.com/algonius/algonius-supervisor/internal/logging"
	"github.com/algonius/algonius-supervisor/pkg/supervisorctl"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// StartCmd represents the start command
var StartCmd = &cobra.Command{
	Use:   "start <agent-name...>",
	Short: "Start one or more agents",
	Long: `Start the specified agents.

Examples:
  supervisorctl start web-server                 # Start specific agent
  supervisorctl start prefix:web-                # Start all agents starting with 'web-'
  supervisorctl start suffix:-worker             # Start all agents ending with '-worker'
  supervisorctl start "regex:web-.*"             # Start agents matching regex pattern
  supervisorctl start web-* db-*                 # Start agents matching wildcard patterns

Supported pattern types:
  exact:     Exact match (default)
  prefix:    Prefix match
  suffix:    Suffix match
  contains:  Contains match
  regex:     Regular expression match
  wildcard:  Wildcard pattern with * and ?`,

	RunE: runStart,
}

// startOptions contains options for the start command
type startOptions struct {
	format string
	colors bool
	wait   bool
}

// runStart executes the start command
func runStart(cmd *cobra.Command, args []string) error {
	logger := logging.GetLogger()
	logger.Info("Executing start command", zap.Strings("args", args))

	if len(args) == 0 {
		return fmt.Errorf("at least one agent name or pattern must be provided")
	}

	// Parse command options
	opts, err := parseStartOptions(cmd)
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

	// Validate patterns
	for _, arg := range args {
		if err := patterns.ValidatePatternString(arg); err != nil {
			return fmt.Errorf("invalid pattern '%s': %w", arg, err)
		}
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

	// First, get current status to resolve patterns and filter agents
	currentStatuses, err := supervisorClient.GetStatus()
	if err != nil {
		return fmt.Errorf("failed to get current agent status: %w", err)
	}

	// Filter agents by patterns
	targetedAgents, err := filterAgentNamesByPatterns(currentStatuses, args)
	if err != nil {
		return fmt.Errorf("failed to filter agents: %w", err)
	}

	if len(targetedAgents) == 0 {
		fmt.Printf("No agents found matching patterns: %s\n", strings.Join(args, ", "))
		return nil
	}

	// Start the agents
	result, err := supervisorClient.StartAgents(targetedAgents...)
	if err != nil {
		return fmt.Errorf("failed to start agents: %w", err)
	}

	// Display results
	if err := displayStartResult(result, opts, config.Display.Colors && opts.colors); err != nil {
		return fmt.Errorf("failed to display start result: %w", err)
	}

	// Wait for agents to be running if requested
	if opts.wait {
		if err := waitForAgentsRunning(supervisorClient, targetedAgents); err != nil {
			return fmt.Errorf("error while waiting for agents: %w", err)
		}
	}

	logger.Info("Start command completed", zap.Int("targeted_agents", len(targetedAgents)), zap.Int("successes", result.Summary.Succeeded))
	return nil
}

// parseStartOptions parses command options from the cobra command
func parseStartOptions(cmd *cobra.Command) (*startOptions, error) {
	opts := &startOptions{
		format: viper.GetString("display.format"),
		colors: viper.GetBool("display.colors"),
		wait:   viper.GetBool("wait"),
	}

	// Override with command-line flags if provided
	if format, err := cmd.Flags().GetString("format"); err == nil && format != "" {
		opts.format = format
	}
	if colors, err := cmd.Flags().GetBool("colors"); err == nil {
		opts.colors = colors
	}
	if wait, err := cmd.Flags().GetBool("wait"); err == nil {
		opts.wait = wait
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

// filterAgentNamesByPatterns filters agent names by patterns
func filterAgentNamesByPatterns(statuses []supervisorctl.AgentStatus, patternStrings []string) ([]string, error) {
	if len(patternStrings) == 0 {
		return nil, fmt.Errorf("no patterns provided")
	}

	// Create matcher for patterns
	matcher, err := patterns.NewMatcher(patternStrings)
	if err != nil {
		return nil, fmt.Errorf("failed to create pattern matcher: %w", err)
	}

	var agentNames []string
	for _, status := range statuses {
		if matcher.Matches(status.Name) {
			agentNames = append(agentNames, status.Name)
		}
	}

	return agentNames, nil
}

// displayStartResult formats and displays the start operation result
func displayStartResult(result *supervisorctl.OperationResult, opts *startOptions, useColors bool) error {
	// Print summary
	printStartSummary(result, useColors)

	// Format the detailed result
	switch opts.format {
	case "table":
		tableFormatter := formatter.NewTableFormatter(os.Stdout).
			WithColors(useColors).
			WithTabWriterSettings(0, 8, 1, ' ', 0)
		return tableFormatter.FormatOperationResult(result)
	case "simple":
		simpleFormatter := formatter.NewSimpleFormatter(os.Stdout).
			WithColors(useColors)
		return simpleFormatter.FormatOperationResult(result)
	case "json":
		jsonFormatter := formatter.NewJSONFormatter(os.Stdout)
		return jsonFormatter.FormatOperationResult(result)
	default:
		return fmt.Errorf("unsupported format: %s", opts.format)
	}
}

// printStartSummary prints a summary of the start operation
func printStartSummary(result *supervisorctl.OperationResult, useColors bool) {
	fmt.Printf("Start operation completed: %d total, %d succeeded, %d failed\n",
		result.Summary.Total, result.Summary.Succeeded, result.Summary.Failed)
	fmt.Printf("Duration: %s\n", result.Summary.Duration)

	if result.Summary.Failed > 0 {
		fmt.Printf("Failed operations: %d\n", result.Summary.Failed)
		for _, failure := range result.Failures {
			color := getErrorColor(useColors)
			fmt.Printf("  %s%s: %s\x1b[0m\n", color, failure.AgentName, failure.Error)
		}
	}

	if result.Summary.Succeeded > 0 {
		fmt.Printf("Successfully started: %d\n", result.Summary.Succeeded)
		if !useColors {
			for _, success := range result.Successes {
				fmt.Printf("  %s: %s\n", success.AgentName, success.Message)
			}
		}
	}
	fmt.Println()
}

// waitForAgentsRunning waits for the specified agents to be in RUNNING state
func waitForAgentsRunning(supervisorClient supervisorctl.ISupervisorctlClient, agentNames []string) error {
	fmt.Printf("Waiting for agents to be running...\n")

	const maxWaitTime = 30 * time.Second
	const checkInterval = 1 * time.Second
	startTime := time.Now()

	for time.Since(startTime) < maxWaitTime {
		statuses, err := supervisorClient.GetStatus(agentNames...)
		if err != nil {
			return fmt.Errorf("failed to check agent status: %w", err)
		}

		allRunning := true
		for _, status := range statuses {
			if status.State != "RUNNING" {
				allRunning = false
				break
			}
		}

		if allRunning {
			fmt.Printf("All agents are running\n")
			return nil
		}

		// Show progress
		runningCount := 0
		for _, status := range statuses {
			if status.State == "RUNNING" {
				runningCount++
			}
		}
		fmt.Printf("\rProgress: %d/%d agents running", runningCount, len(statuses))
		time.Sleep(checkInterval)
	}

	return fmt.Errorf("timeout waiting for agents to be running after %v", maxWaitTime)
}

// getErrorColor returns ANSI color code for error messages
func getErrorColor(useColors bool) string {
	if !useColors {
		return ""
	}
	return "\x1b[31m" // Red
}

func init() {
	// Add command flags
	StartCmd.Flags().StringP("format", "f", "", "Output format (table, simple, json)")
	StartCmd.Flags().BoolP("colors", "c", false, "Force colored output")
	StartCmd.Flags().BoolP("wait", "w", false, "Wait for agents to be running before returning")

	// Bind flags to viper
	viper.BindPFlag("display.format", StartCmd.Flags().Lookup("format"))
	viper.BindPFlag("display.colors", StartCmd.Flags().Lookup("colors"))
	viper.BindPFlag("wait", StartCmd.Flags().Lookup("wait"))
}