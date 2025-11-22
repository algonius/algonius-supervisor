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

// StopCmd represents the stop command
var StopCmd = &cobra.Command{
	Use:   "stop <agent-name...>",
	Short: "Stop one or more agents",
	Long: `Stop the specified agents.

Examples:
  supervisorctl stop web-server                  # Stop specific agent
  supervisorctl stop prefix:web-                 # Stop all agents starting with 'web-'
  supervisorctl stop suffix:-worker              # Stop all agents ending with '-worker'
  supervisorctl stop "regex:web-.*"              # Stop agents matching regex pattern
  supervisorctl stop web-* db-*                  # Stop agents matching wildcard patterns

Supported pattern types:
  exact:     Exact match (default)
  prefix:    Prefix match
  suffix:    Suffix match
  contains:  Contains match
  regex:     Regular expression match
  wildcard:  Wildcard pattern with * and ?`,

	RunE: runStop,
}

// stopOptions contains options for the stop command
type stopOptions struct {
	format string
	colors bool
	wait   bool
	force  bool
}

// runStop executes the stop command
func runStop(cmd *cobra.Command, args []string) error {
	logger := logging.GetLogger()
	logger.Info("Executing stop command", zap.Strings("args", args))

	if len(args) == 0 {
		return fmt.Errorf("at least one agent name or pattern must be provided")
	}

	// Parse command options
	opts, err := parseStopOptions(cmd)
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

	// Confirm operation if force flag is not set and we have multiple agents
	if !opts.force && len(targetedAgents) > 1 {
		fmt.Printf("About to stop %d agents: %s\n", len(targetedAgents), strings.Join(targetedAgents, ", "))
		fmt.Print("Are you sure? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	// Stop the agents
	result, err := supervisorClient.StopAgents(targetedAgents...)
	if err != nil {
		return fmt.Errorf("failed to stop agents: %w", err)
	}

	// Display results
	if err := displayStopResult(result, opts, config.Display.Colors && opts.colors); err != nil {
		return fmt.Errorf("failed to display stop result: %w", err)
	}

	// Wait for agents to be stopped if requested
	if opts.wait {
		if err := waitForAgentsStopped(supervisorClient, targetedAgents); err != nil {
			return fmt.Errorf("error while waiting for agents: %w", err)
		}
	}

	logger.Info("Stop command completed", zap.Int("targeted_agents", len(targetedAgents)), zap.Int("successes", result.Summary.Succeeded))
	return nil
}

// parseStopOptions parses command options from the cobra command
func parseStopOptions(cmd *cobra.Command) (*stopOptions, error) {
	opts := &stopOptions{
		format: viper.GetString("display.format"),
		colors: viper.GetBool("display.colors"),
		wait:   viper.GetBool("wait"),
		force:  viper.GetBool("force"),
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
	if force, err := cmd.Flags().GetBool("force"); err == nil {
		opts.force = force
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

// displayStopResult formats and displays the stop operation result
func displayStopResult(result *supervisorctl.OperationResult, opts *stopOptions, useColors bool) error {
	// Print summary
	printStopSummary(result, useColors)

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

// printStopSummary prints a summary of the stop operation
func printStopSummary(result *supervisorctl.OperationResult, useColors bool) {
	fmt.Printf("Stop operation completed: %d total, %d succeeded, %d failed\n",
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
		fmt.Printf("Successfully stopped: %d\n", result.Summary.Succeeded)
		if !useColors {
			for _, success := range result.Successes {
				fmt.Printf("  %s: %s\n", success.AgentName, success.Message)
			}
		}
	}
	fmt.Println()
}

// waitForAgentsStopped waits for the specified agents to be in STOPPED state
func waitForAgentsStopped(supervisorClient supervisorctl.ISupervisorctlClient, agentNames []string) error {
	fmt.Printf("Waiting for agents to be stopped...\n")

	const maxWaitTime = 30 * time.Second
	const checkInterval = 500 * time.Millisecond
	startTime := time.Now()

	for time.Since(startTime) < maxWaitTime {
		statuses, err := supervisorClient.GetStatus(agentNames...)
		if err != nil {
			return fmt.Errorf("failed to check agent status: %w", err)
		}

		allStopped := true
		for _, status := range statuses {
			if status.State != "STOPPED" && status.State != "FATAL" && status.State != "FAILED" && status.State != "EXITED" {
				allStopped = false
				break
			}
		}

		if allStopped {
			fmt.Printf("All agents are stopped\n")
			return nil
		}

		// Show progress
		stoppedCount := 0
		for _, status := range statuses {
			if status.State == "STOPPED" || status.State == "FATAL" || status.State == "FAILED" || status.State == "EXITED" {
				stoppedCount++
			}
		}
		fmt.Printf("\rProgress: %d/%d agents stopped", stoppedCount, len(statuses))
		time.Sleep(checkInterval)
	}

	return fmt.Errorf("timeout waiting for agents to be stopped after %v", maxWaitTime)
}

func init() {
	// Add command flags
	StopCmd.Flags().StringP("format", "f", "", "Output format (table, simple, json)")
	StopCmd.Flags().BoolP("colors", "c", false, "Force colored output")
	StopCmd.Flags().BoolP("wait", "w", false, "Wait for agents to be stopped before returning")
	StopCmd.Flags().BoolP("force", "F", false, "Skip confirmation for multiple agents")

	// Bind flags to viper
	viper.BindPFlag("display.format", StopCmd.Flags().Lookup("format"))
	viper.BindPFlag("display.colors", StopCmd.Flags().Lookup("colors"))
	viper.BindPFlag("wait", StopCmd.Flags().Lookup("wait"))
	viper.BindPFlag("force", StopCmd.Flags().Lookup("force"))
}