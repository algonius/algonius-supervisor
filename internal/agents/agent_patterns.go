package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/algonius/algonius-supervisor/internal/models"
	"github.com/algonius/algonius-supervisor/pkg/types"
)

// InputPatternHandler handles different input patterns for agents
type InputPatternHandler interface {
	// PrepareInput prepares the input for the agent based on the pattern
	PrepareInput(input string, config *models.AgentConfiguration) ([]string, io.Reader, error)
	
	// ProcessOutput processes the output from the agent based on the pattern
	ProcessOutput(output []byte, config *models.AgentConfiguration) (string, error)
}

// StdinHandler handles stdin input pattern
type StdinHandler struct{}

// PrepareInput for StdinHandler
func (h *StdinHandler) PrepareInput(input string, config *models.AgentConfiguration) ([]string, io.Reader, error) {
	// For stdin pattern, we pass arguments and use stdin for the actual input
	args := h.buildArgs(config)
	return args, strings.NewReader(input), nil
}

// ProcessOutput for StdinHandler
func (h *StdinHandler) ProcessOutput(output []byte, config *models.AgentConfiguration) (string, error) {
	return string(output), nil
}

// buildArgs builds command line arguments for stdin handler
func (h *StdinHandler) buildArgs(config *models.AgentConfiguration) []string {
	var args []string

	// Add default arguments from configuration
	if config.CliArgs != nil {
		for arg, value := range config.CliArgs {
			args = append(args, arg)
			if value != "" {
				args = append(args, value)
			}
		}
	}

	return args
}

// FileHandler handles file input pattern
type FileHandler struct{}

// PrepareInput for FileHandler
func (h *FileHandler) PrepareInput(input string, config *models.AgentConfiguration) ([]string, io.Reader, error) {
	if config.InputFileTemplate == "" {
		return nil, nil, fmt.Errorf("input file template not specified in configuration")
	}

	// Generate the input filename based on the template
	inputFilename := processTemplate(config.InputFileTemplate, map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"agent_id":  config.ID,
	})

	// Write input to the file
	if err := os.WriteFile(inputFilename, []byte(input), 0644); err != nil {
		return nil, nil, fmt.Errorf("failed to write input to file: %w", err)
	}

	// Build arguments that include the input file path
	args := h.buildArgs(config, inputFilename)

	return args, nil, nil
}

// ProcessOutput for FileHandler
func (h *FileHandler) ProcessOutput(output []byte, config *models.AgentConfiguration) (string, error) {
	// For file output, we need to read the output file
	if config.OutputFileTemplate == "" {
		return string(output), nil
	}

	// Generate the output filename based on the template
	outputFilename := processTemplate(config.OutputFileTemplate, map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"agent_id":  config.ID,
	})

	// Read the output file
	outputContent, err := os.ReadFile(outputFilename)
	if err != nil {
		return "", fmt.Errorf("failed to read output file: %w", err)
	}

	return string(outputContent), nil
}

// buildArgs builds command line arguments for file handler
func (h *FileHandler) buildArgs(config *models.AgentConfiguration, inputFilename string) []string {
	var args []string

	// Add default arguments from configuration
	if config.CliArgs != nil {
		for arg, value := range config.CliArgs {
			args = append(args, arg)
			if value != "" {
				args = append(args, value)
			}
		}
	}

	// Add the input file as an argument
	args = append(args, inputFilename)

	return args
}

// ArgsHandler handles command-line arguments input pattern
type ArgsHandler struct{}

// PrepareInput for ArgsHandler
func (h *ArgsHandler) PrepareInput(input string, config *models.AgentConfiguration) ([]string, io.Reader, error) {
	// For args pattern, we add the input as command-line arguments
	args := h.buildArgs(config, input)
	return args, nil, nil
}

// ProcessOutput for ArgsHandler
func (h *ArgsHandler) ProcessOutput(output []byte, config *models.AgentConfiguration) (string, error) {
	return string(output), nil
}

// buildArgs builds command line arguments for args handler
func (h *ArgsHandler) buildArgs(config *models.AgentConfiguration, input string) []string {
	var args []string

	// Add default arguments from configuration
	if config.CliArgs != nil {
		for arg, value := range config.CliArgs {
			args = append(args, arg)
			if value != "" {
				args = append(args, value)
			}
		}
	}

	// Add the input as an argument
	args = append(args, input)

	return args
}

// JSONRPCHandler handles JSON-RPC input/output pattern
type JSONRPCHandler struct{}

// PrepareInput for JSONRPCHandler
func (h *JSONRPCHandler) PrepareInput(input string, config *models.AgentConfiguration) ([]string, io.Reader, error) {
	// For JSON-RPC, we format the input as a JSON-RPC request
	rpcRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "execute",
		"params":  input,
		"id":      generateRequestID(),
	}

	jsonData, err := json.Marshal(rpcRequest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	args := h.buildArgs(config)
	return args, bytes.NewReader(jsonData), nil
}

// ProcessOutput for JSONRPCHandler
func (h *JSONRPCHandler) ProcessOutput(output []byte, config *models.AgentConfiguration) (string, error) {
	// For JSON-RPC, we parse the output as a JSON-RPC response
	var rpcResponse map[string]interface{}
	if err := json.Unmarshal(output, &rpcResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON-RPC response: %w", err)
	}

	// Check if it's an error response
	if err, hasError := rpcResponse["error"]; hasError {
		return "", fmt.Errorf("JSON-RPC error response: %v", err)
	}

	// Extract the result
	if result, hasResult := rpcResponse["result"]; hasResult {
		resultBytes, err := json.Marshal(result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}
		return string(resultBytes), nil
	}

	return string(output), nil
}

// buildArgs builds command line arguments for JSON-RPC handler
func (h *JSONRPCHandler) buildArgs(config *models.AgentConfiguration) []string {
	var args []string

	// Add default arguments from configuration
	if config.CliArgs != nil {
		for arg, value := range config.CliArgs {
			args = append(args, arg)
			if value != "" {
				args = append(args, value)
			}
		}
	}

	return args
}

// GetInputPatternHandler returns the appropriate input pattern handler based on the configuration
func GetInputPatternHandler(inputPattern types.InputPattern) InputPatternHandler {
	switch inputPattern {
	case types.StdinPattern:
		return &StdinHandler{}
	case types.FilePattern:
		return &FileHandler{}
	case types.ArgsPattern:
		return &ArgsHandler{}
	case types.JsonRpcPattern:
		return &JSONRPCHandler{}
	default:
		// Default to stdin handler for unknown patterns
		return &StdinHandler{}
	}
}

// processTemplate processes a template string with the given variables
func processTemplate(template string, vars map[string]interface{}) string {
	result := template
	for key, value := range vars {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// generateRequestID generates a unique request ID (placeholder implementation)
func generateRequestID() string {
	// In a real implementation, this would generate a proper ID
	return fmt.Sprintf("req-%d", time.Now().Unix())
}

// ExecuteAgentWithPattern executes an agent using the appropriate pattern handler
func ExecuteAgentWithPattern(config *models.AgentConfiguration, input string) (string, error) {
	// Get the appropriate handler for the input pattern
	handler := GetInputPatternHandler(config.InputPattern)

	// Prepare the command arguments and input
	args, inputReader, err := handler.PrepareInput(input, config)
	if err != nil {
		return "", fmt.Errorf("failed to prepare input: %w", err)
	}

	// Create the command
	cmd := exec.Command(config.ExecutablePath, args...)

	// Set working directory if specified
	if config.WorkingDirectory != "" {
		cmd.Dir = config.WorkingDirectory
	}

	// Set environment variables
	if config.Envs != nil {
		envVars := make([]string, 0, len(config.Envs))
		for key, value := range config.Envs {
			envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = append(os.Environ(), envVars...)
	}

	// Set up stdin if we have an input reader
	var stdinPipe io.WriteCloser
	if inputReader != nil {
		stdinPipe, err = cmd.StdinPipe()
		if err != nil {
			return "", fmt.Errorf("failed to create stdin pipe: %w", err)
		}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Write input to stdin if needed
	if stdinPipe != nil {
		go func() {
			defer stdinPipe.Close()
			_, err := io.Copy(stdinPipe, inputReader)
			if err != nil {
				// Log the error but don't fail the entire operation
				// as the command might have already started successfully
			}
		}()
	}

	// Wait for command to finish and get output
	output, err := cmd.Output()
	if err != nil {
		// Check if the error is due to a non-zero exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			// Use stderr if stdout is empty
			if len(exitError.Stderr) > 0 {
				output = exitError.Stderr
			} else {
				output = []byte(err.Error())
			}
		} else {
			return "", fmt.Errorf("command execution failed: %w", err)
		}
	}

	// Process the output using the handler
	result, err := handler.ProcessOutput(output, config)
	if err != nil {
		return "", fmt.Errorf("failed to process output: %w", err)
	}

	return result, nil
}