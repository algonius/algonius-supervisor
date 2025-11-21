package agents

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/algonius/algonius-supervisor/internal/models"
	"go.uber.org/zap"
)

// GenericAgent implements the IAgent interface for a generic CLI agent
type GenericAgent struct {
	config *models.AgentConfiguration
	logger *zap.Logger
}

// NewGenericAgent creates a new instance of GenericAgent
func NewGenericAgent(config *models.AgentConfiguration, logger *zap.Logger) *GenericAgent {
	return &GenericAgent{
		config: config,
		logger: logger,
	}
}

// Execute runs the agent with the given input
func (ga *GenericAgent) Execute(ctx context.Context, input string) (*models.ExecutionResult, error) {
	if ga.config == nil {
		return nil, fmt.Errorf("agent configuration is nil")
	}

	startTime := time.Now()

	// Create execution result
	result := &models.ExecutionResult{
		ID:        generateExecutionID(),
		AgentID:   ga.config.ID,
		StartTime: startTime,
		Input:     input, // This will be sanitized later
	}

	// Validate the agent configuration
	if err := ga.Validate(); err != nil {
		ga.logger.Error("agent validation failed", zap.Error(err))
		result.Status = models.FailureStatus
		result.Error = err.Error()
		result.EndTime = time.Now()
		result.SanitizeInput()
		return result, err
	}

	// Prepare command execution based on input pattern
	cmd, stdin, err := ga.prepareCommand(input)
	if err != nil {
		ga.logger.Error("failed to prepare command", zap.Error(err))
		result.Status = models.FailureStatus
		result.Error = err.Error()
		result.EndTime = time.Now()
		result.SanitizeInput()
		return result, err
	}

	// Set context with timeout
	if ga.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(ga.config.Timeout)*time.Second)
		defer cancel()
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		ga.logger.Error("failed to start command", zap.Error(err))
		result.Status = models.FailureStatus
		result.Error = err.Error()
		result.EndTime = time.Now()
		result.SanitizeInput()
		return result, err
	}

	// Write input to stdin if needed
	if stdin != nil {
		_, err := stdin.Write([]byte(input))
		if err != nil {
			ga.logger.Error("failed to write to stdin", zap.Error(err))
			// Continue execution even if input write fails
		}
		stdin.Close()
	}

	// Wait for command to finish or context to be cancelled
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		// Context was cancelled (timeout or cancellation)
		if ctx.Err() == context.DeadlineExceeded {
			ga.logger.Info("agent execution timed out", zap.String("agent_id", ga.config.ID))
			// Attempt to kill the process
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			result.Status = models.TimeoutStatus
			result.Error = "execution timed out"
		} else {
			ga.logger.Info("agent execution cancelled", zap.String("agent_id", ga.config.ID))
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
			result.Status = models.CancelledStatus
			result.Error = "execution cancelled"
		}
	case err := <-done:
		// Command completed
		if err != nil {
			ga.logger.Info("agent execution completed with error", 
				zap.String("agent_id", ga.config.ID), 
				zap.Error(err))
			result.Status = models.FailureStatus
			result.Error = err.Error()
		} else {
			ga.logger.Info("agent execution completed successfully", 
				zap.String("agent_id", ga.config.ID))
			result.Status = models.SuccessStatus
		}
	}

	// Capture the end time
	result.EndTime = time.Now()
	
	// Capture process ID if available
	if cmd.ProcessState != nil {
		result.ProcessID = cmd.ProcessState.Pid()
	}

	// Get output based on output pattern
	output, err := ga.getOutput()
	if err != nil {
		ga.logger.Warn("failed to get output", zap.Error(err))
		// Continue with empty output
	} else {
		result.Output = output
	}

	// Sanitize sensitive data
	result.SanitizeInput()
	result.SanitizeOutput()

	return result, nil
}

// prepareCommand prepares the command based on the agent configuration
func (ga *GenericAgent) prepareCommand(input string) (*exec.Cmd, io.WriteCloser, error) {
	// Split the executable path and arguments
	executable := ga.config.ExecutablePath
	args := ga.buildArgs(input)

	// Create the command
	cmd := exec.CommandContext(context.Background(), executable, args...)
	
	// Set working directory if specified
	if ga.config.WorkingDirectory != "" {
		cmd.Dir = ga.config.WorkingDirectory
	}

	// Set environment variables
	if ga.config.Envs != nil {
		envVars := make([]string, 0, len(ga.config.Envs))
		for key, value := range ga.config.Envs {
			envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = append(os.Environ(), envVars...)
	}

	// Handle input based on pattern
	var stdin io.WriteCloser
	var err error

	switch ga.config.InputPattern {
	case models.StdinPattern:
		// Use stdin for input
		stdin, err = cmd.StdinPipe()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create stdin pipe: %w", err)
		}
	case models.FilePattern:
		// Create input file based on template
		if ga.config.InputFileTemplate != "" {
			filename := ga.processTemplate(ga.config.InputFileTemplate, map[string]interface{}{
				"input": input,
				"agent_id": ga.config.ID,
				"execution_id": generateExecutionID(),
			})
			
			// Write input to the file
			err := os.WriteFile(filename, []byte(input), 0644)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to write input to file: %w", err)
			}
			
			// Add the file as an argument
			args = append(args, filename)
			cmd = exec.CommandContext(context.Background(), executable, args...)
		}
	case models.ArgsPattern:
		// Input is passed as command line arguments, already handled in buildArgs
	case models.JsonRpcPattern:
		// For JSON-RPC, we'll use stdin but format the input appropriately
		stdin, err = cmd.StdinPipe()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create stdin pipe for JSON-RPC: %w", err)
		}
	}

	return cmd, stdin, nil
}

// buildArgs builds command line arguments based on the configuration
func (ga *GenericAgent) buildArgs(input string) []string {
	var args []string

	// Add default arguments from configuration
	if ga.config.CliArgs != nil {
		for arg, value := range ga.config.CliArgs {
			args = append(args, arg)
			if value != "" {
				args = append(args, value)
			}
		}
	}

	// Add input as argument if using ArgsPattern
	if ga.config.InputPattern == models.ArgsPattern {
		args = append(args, input)
	}

	return args
}

// getOutput gets the output based on the output pattern
func (ga *GenericAgent) getOutput() (string, error) {
	// This would need to be implemented based on the output pattern
	// For now, we'll return an empty string as output capture
	// would require more complex handling based on the output pattern
	return "", nil
}

// processTemplate processes a template string with the given variables
func (ga *GenericAgent) processTemplate(template string, vars map[string]interface{}) string {
	result := template
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// GetID returns the agent's ID
func (ga *GenericAgent) GetID() string {
	return ga.config.ID
}

// GetName returns the agent's name
func (ga *GenericAgent) GetName() string {
	return ga.config.Name
}

// GetType returns the agent's type
func (ga *GenericAgent) GetType() string {
	return ga.config.AgentType
}

// IsReadOnly returns whether the agent is read-only or read-write
func (ga *GenericAgent) IsReadOnly() bool {
	return ga.config.AccessType == models.ReadOnlyAccessType
}

// GetConfig returns the agent's configuration
func (ga *GenericAgent) GetConfig() *models.AgentConfiguration {
	return ga.config
}

// Validate checks if the agent configuration is valid
func (ga *GenericAgent) Validate() error {
	if ga.config == nil {
		return fmt.Errorf("agent configuration is nil")
	}

	// Validate the configuration using the built-in Validate method
	return ga.config.Validate()
}

// generateExecutionID generates a unique execution ID (placeholder implementation)
func generateExecutionID() string {
	// In a real implementation, this would generate a proper UUID
	// For example, using github.com/google/uuid
	return "exec-" + time.Now().Format("20060102-150405")
}