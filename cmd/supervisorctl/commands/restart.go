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
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// RestartCmd represents the restart command
var RestartCmd = &cobra.Command{
	Use:   "restart <agent-name...>",
	Short: "Restart one or more agents",
	Long: `Restart the specified agents. This is equivalent to stopping and then starting each agent.

Examples:
  supervisorctl restart web-server               # Restart specific agent
  supervisorctl restart prefix:web-              # Restart all agents starting with 'web-'
  supervisorctl restart suffix:-worker           # Restart all agents ending with '-worker'
  supervisorctl restart "regex:web-.*"           # Restart agents matching regex pattern
  supervisorctl restart web-* db-*               # Restart agents matching wildcard patterns

Supported pattern types:
  exact:     Exact match (default)
  prefix:    Prefix match
  suffix:    Suffix match
  contains:  Contains match
  regex:     Regular expression match
  wildcard:  Wildcard pattern with * and ?`,

	RunE: runRestart,
}

// restartOptions contains options for the restart command
type restartOptions struct {
	format string
	colors bool
	wait   bool
	force  bool
}

// runRestart executes the restart command
func runRestart(cmd *cobra.Command, args []string) error {
	logger := logging.GetLogger()
	logger.Info("Executing restart command", zap.Strings("args", args))

	if len(args) == 0 {
		return fmt.Errorf("at least one agent name or pattern must be provided")
	}

	// Parse command options
	opts, err := parseRestartOptions(cmd)
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

	// Show what will be restarted
	fmt.Printf("Restarting %d agents: %s\n", len(targetedAgents), strings.Join(targetedAgents, ", "))

	// Confirm operation if force flag is not set and we have multiple agents
	if !opts.force && len(targetedAgents) > 1 {
		fmt.Print("Are you sure? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	// Restart the agents
	result, err := supervisorClient.RestartAgents(targetedAgents...)
	if err != nil {
		return fmt.Errorf("failed to restart agents: %w", err)
	}

	// Display results
	if err := displayRestartResult(result, opts, config.Display.Colors && opts.colors); err != nil {
		return fmt.Errorf("failed to display restart result: %w", err)
	}

	// Wait for agents to be running again if requested
	if opts.wait {
		if err := waitForAgentsRunning(supervisorClient, targetedAgents); err != nil {
			return fmt.Errorf("error while waiting for agents: %w", err)
		}
	}

	logger.Info("Restart command completed", zap.Int("targeted_agents", len(targetedAgents)), zap.Int("successes", result.Summary.Succeeded))
	return nil
}

// parseRestartOptions parses command options from the cobra command
func parseRestartOptions(cmd *cobra.Command) (*restartOptions, error) {
	opts := &restartOptions{
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

// displayRestartResult formats and displays the restart operation result
func displayRestartResult(result *supervisorctl.OperationResult, opts *restartOptions, useColors bool) error {
	switch opts.format {
	case "table":
		tableFormatter := formatter.NewTableFormatter(os.Stdout).
			WithColors(useColors).
			WithTabWriterSettings(0, 8, 1, ' ', 0)
		printRestartSummary(result, useColors)
		return tableFormatter.FormatOperationResult(result)
	case "simple":
		simpleFormatter := formatter.NewSimpleFormatter(os.Stdout).
			WithColors(useColors)
		printRestartSummary(result, useColors)
		return simpleFormatter.FormatOperationResult(result)
	case "json":
		jsonFormatter := formatter.NewJSONFormatter(os.Stdout)
		return jsonFormatter.FormatOperationResult(result)
	default:
		return fmt.Errorf("unsupported format: %s", opts.format)
	}
}

// printRestartSummary prints a summary of the restart operation
func printRestartSummary(result *supervisorctl.OperationResult, useColors bool) {
	fmt.Printf("Restart operation completed: %d total, %d succeeded, %d failed\n",
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
		fmt.Printf("Successfully restarted: %d\n", result.Summary.Succeeded)
		if !useColors {
			for _, success := range result.Successes {
				fmt.Printf("  %s: %s\n", success.AgentName, success.Message)
			}
		}
	}
	fmt.Println()
}

// getWarningColor returns ANSI color code for warning messages
func getWarningColor(useColors bool) string {
	if !useColors {
		return ""
	}
	return "\x1b[33m" // Yellow
}

// getSuccessColor returns ANSI color code for success messages
func getSuccessColor(useColors bool) string {
	if !useColors {
		return ""
	}
	return "\x1b[32m" // Green
}

func init() {
	// Add command flags
	RestartCmd.Flags().StringP("format", "f", "", "Output format (table, simple, json)")
	RestartCmd.Flags().BoolP("colors", "c", false, "Force colored output")
	RestartCmd.Flags().BoolP("wait", "w", false, "Wait for agents to be running after restart")
	RestartCmd.Flags().BoolP("force", "F", false, "Skip confirmation for multiple agents")

	// Bind flags to viper
	viper.BindPFlag("display.format", RestartCmd.Flags().Lookup("format"))
	viper.BindPFlag("display.colors", RestartCmd.Flags().Lookup("colors"))
	viper.BindPFlag("wait", RestartCmd.Flags().Lookup("wait"))
	viper.BindPFlag("force", RestartCmd.Flags().Lookup("force"))
}