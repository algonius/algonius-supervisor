package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/algonius/algonius-supervisor/internal/cli/errors"
	"github.com/algonius/algonius-supervisor/internal/models"
	"github.com/algonius/algonius-supervisor/pkg/supervisorctl"
)

// SupervisorClient implements the ISupervisorctlClient interface
type SupervisorClient struct {
	httpClient IHTTPClient
	serverURL  string
	authToken  string
	timeout    time.Duration
}

// NewSupervisorClient creates a new SupervisorClient instance
func NewSupervisorClient(httpClient IHTTPClient, serverURL, authToken string, timeout time.Duration) *SupervisorClient {
	return &SupervisorClient{
		httpClient: httpClient,
		serverURL:  serverURL,
		authToken:  authToken,
		timeout:    timeout,
	}
}

// GetStatus retrieves the status of specified agents or all agents if no names provided
func (c *SupervisorClient) GetStatus(agentNames ...string) ([]supervisorctl.AgentStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	if len(agentNames) == 0 {
		// Get all agents status
		return c.getAllAgentsStatus(ctx)
	}

	// Get specific agents status
	return c.getSpecificAgentsStatus(ctx, agentNames)
}

// StartAgents starts the specified agents
func (c *SupervisorClient) StartAgents(agentNames ...string) (*supervisorctl.OperationResult, error) {
	if len(agentNames) == 0 {
		return nil, errors.ValidationError("at least one agent name must be provided", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	operationResult := models.NewBatchOperationResult(models.OperationTypeStart, len(agentNames), false, 1)

	for _, agentName := range agentNames {
		result, err := c.startAgent(ctx, agentName)
		if err != nil {
			opResult := models.NewOperationResult(agentName, models.OperationTypeStart)
			opResult.SetFailed("Failed to start agent", err.Error())
			operationResult.AddResult(*opResult)
			continue
		}
		operationResult.AddResult(*result)
	}

	operationResult.Complete()
	return c.convertBatchOperationResult(operationResult), nil
}

// StopAgents stops the specified agents
func (c *SupervisorClient) StopAgents(agentNames ...string) (*supervisorctl.OperationResult, error) {
	if len(agentNames) == 0 {
		return nil, errors.ValidationError("at least one agent name must be provided", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	operationResult := models.NewBatchOperationResult(models.OperationTypeStop, len(agentNames), false, 1)

	for _, agentName := range agentNames {
		result, err := c.stopAgent(ctx, agentName)
		if err != nil {
			opResult := models.NewOperationResult(agentName, models.OperationTypeStop)
			opResult.SetFailed("Failed to stop agent", err.Error())
			operationResult.AddResult(*opResult)
			continue
		}
		operationResult.AddResult(*result)
	}

	operationResult.Complete()
	return c.convertBatchOperationResult(operationResult), nil
}

// RestartAgents restarts the specified agents
func (c *SupervisorClient) RestartAgents(agentNames ...string) (*supervisorctl.OperationResult, error) {
	if len(agentNames) == 0 {
		return nil, errors.ValidationError("at least one agent name must be provided", nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	operationResult := models.NewBatchOperationResult(models.OperationTypeRestart, len(agentNames), false, 1)

	for _, agentName := range agentNames {
		result, err := c.restartAgent(ctx, agentName)
		if err != nil {
			opResult := models.NewOperationResult(agentName, models.OperationTypeRestart)
			opResult.SetFailed("Failed to restart agent", err.Error())
			operationResult.AddResult(*opResult)
			continue
		}
		operationResult.AddResult(*result)
	}

	operationResult.Complete()
	return c.convertBatchOperationResult(operationResult), nil
}

// TailLogs streams logs from an agent (placeholder for User Story 2)
func (c *SupervisorClient) TailLogs(agentName string, follow bool) (<-chan supervisorctl.LogEntry, error) {
	// This will be implemented in User Story 2
	return nil, errors.APIError("Log tailing not yet implemented", nil)
}

// GetEventStream streams agent events (placeholder for User Story 2)
func (c *SupervisorClient) GetEventStream() (<-chan supervisorctl.AgentEvent, error) {
	// This will be implemented in User Story 2
	return nil, errors.APIError("Event streaming not yet implemented", nil)
}

// GetServerInfo retrieves server information
func (c *SupervisorClient) GetServerInfo() (*supervisorctl.ServerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	url := fmt.Sprintf("%s/api/v1/server/info", c.serverURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.ConnectionError("Failed to create request", err)
	}

	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.ConnectionError("Failed to get server info", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp, "Failed to get server info")
	}

	var response struct {
		Success bool                   `json:"success"`
		Data    supervisorctl.ServerInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.APIError("Failed to parse server info response", err)
	}

	if !response.Success {
		return nil, errors.APIError("Server info request failed", resp.StatusCode)
	}

	return &response.Data, nil
}

// ValidateConfig validates the current configuration
func (c *SupervisorClient) ValidateConfig() (*supervisorctl.ConfigValidation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	url := fmt.Sprintf("%s/api/v1/config/validate", c.serverURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return nil, errors.ConnectionError("Failed to create request", err)
	}

	c.setAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.ConnectionError("Failed to validate config", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp, "Failed to validate config")
	}

	var response struct {
		Success bool                        `json:"success"`
		Data    supervisorctl.ConfigValidation `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.APIError("Failed to parse config validation response", err)
	}

	if !response.Success {
		return nil, errors.APIError("Config validation request failed", resp.StatusCode)
	}

	return &response.Data, nil
}

// Helper methods

func (c *SupervisorClient) getAllAgentsStatus(ctx context.Context) ([]supervisorctl.AgentStatus, error) {
	url := fmt.Sprintf("%s/api/v1/agents/status", c.serverURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.ConnectionError("Failed to create request", err)
	}

	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.ConnectionError("Failed to get agents status", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp, "Failed to get agents status")
	}

	var response struct {
		Success bool               `json:"success"`
		Data    []models.AgentStatus `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.APIError("Failed to parse status response", err)
	}

	if !response.Success {
		return nil, errors.APIError("Status request failed", resp.StatusCode)
	}

	return c.convertAgentStatuses(response.Data), nil
}

func (c *SupervisorClient) getSpecificAgentsStatus(ctx context.Context, agentNames []string) ([]supervisorctl.AgentStatus, error) {
	var statuses []supervisorctl.AgentStatus

	for _, agentName := range agentNames {
		status, err := c.getSingleAgentStatus(ctx, agentName)
		if err != nil {
			// Continue with other agents even if one fails
			continue
		}
		statuses = append(statuses, *status)
	}

	return statuses, nil
}

func (c *SupervisorClient) getSingleAgentStatus(ctx context.Context, agentName string) (*supervisorctl.AgentStatus, error) {
	url := fmt.Sprintf("%s/api/v1/agents/%s/status", c.serverURL, agentName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.ConnectionError("Failed to create request", err)
	}

	c.setAuthHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.ConnectionError("Failed to get agent status", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleHTTPError(resp, fmt.Sprintf("Failed to get status for agent '%s'", agentName))
	}

	var response struct {
		Success bool               `json:"success"`
		Data    models.AgentStatus `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.APIError("Failed to parse agent status response", err)
	}

	if !response.Success {
		return nil, errors.APIError("Agent status request failed", resp.StatusCode)
	}

	converted := c.convertAgentStatus(response.Data)
	return &converted, nil
}

func (c *SupervisorClient) startAgent(ctx context.Context, agentName string) (*models.OperationResult, error) {
	url := fmt.Sprintf("%s/api/v1/agents/%s/start", c.serverURL, agentName)
	return c.performLifecycleOperation(ctx, "POST", url, agentName, models.OperationTypeStart)
}

func (c *SupervisorClient) stopAgent(ctx context.Context, agentName string) (*models.OperationResult, error) {
	url := fmt.Sprintf("%s/api/v1/agents/%s/stop", c.serverURL, agentName)
	return c.performLifecycleOperation(ctx, "POST", url, agentName, models.OperationTypeStop)
}

func (c *SupervisorClient) restartAgent(ctx context.Context, agentName string) (*models.OperationResult, error) {
	url := fmt.Sprintf("%s/api/v1/agents/%s/restart", c.serverURL, agentName)
	return c.performLifecycleOperation(ctx, "POST", url, agentName, models.OperationTypeRestart)
}

func (c *SupervisorClient) performLifecycleOperation(ctx context.Context, method, url, agentName string, operation models.OperationType) (*models.OperationResult, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return nil, errors.ConnectionError("Failed to create request", err)
	}

	c.setAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.ConnectionError("Failed to perform operation", err)
	}
	defer resp.Body.Close()

	var apiResponse struct {
		Success bool                 `json:"success"`
		Error   string               `json:"error,omitempty"`
		Data    models.OperationResult `json:"data,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, errors.APIError("Failed to parse operation response", err)
	}

	if resp.StatusCode != http.StatusOK {
		if apiResponse.Error != "" {
			result := models.NewOperationResult(agentName, operation)
			result.SetFailed("Operation failed", apiResponse.Error)
			return result, nil
		}
		return nil, c.handleHTTPError(resp, fmt.Sprintf("Failed to perform %s operation on agent '%s'", operation, agentName))
	}

	if !apiResponse.Success {
		result := models.NewOperationResult(agentName, operation)
		result.SetFailed("Operation failed", apiResponse.Error)
		return result, nil
	}

	return &apiResponse.Data, nil
}

func (c *SupervisorClient) setAuthHeader(req *http.Request) {
	if c.authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))
	}
}

func (c *SupervisorClient) handleHTTPError(resp *http.Response, message string) error {
	var errorResponse struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}

	json.NewDecoder(resp.Body).Decode(&errorResponse)

	details := map[string]interface{}{
		"status_code": resp.StatusCode,
		"error":       errorResponse.Error,
	}

	if errorResponse.Error != "" {
		details["message"] = errorResponse.Error
		return errors.APIError(message, details)
	}

	return errors.APIError(message, details)
}

// Conversion methods

func (c *SupervisorClient) convertAgentStatuses(statuses []models.AgentStatus) []supervisorctl.AgentStatus {
	var converted []supervisorctl.AgentStatus
	for _, status := range statuses {
		converted = append(converted, c.convertAgentStatus(status))
	}
	return converted
}

func (c *SupervisorClient) convertAgentStatus(status models.AgentStatus) supervisorctl.AgentStatus {
	return supervisorctl.AgentStatus{
		Name:         status.Name,
		State:        status.State.String(),
		PID:          status.ProcessID,
		StartTime:    status.StartTime,
		Duration:     status.GetFormattedUptime(),
		ExitStatus:   status.ExitCode,
		RestartCount: status.RestartCount,
		LastError:    status.ExitMessage,
		CPUUsage:     0, // Will be set from ResourceUsage if available
		MemoryUsage:  0, // Will be set from ResourceUsage if available
		DiskUsage:    0, // Will be set from ResourceUsage if available
	}
}

func (c *SupervisorClient) convertBatchOperationResult(result *models.BatchOperationResult) *supervisorctl.OperationResult {
	var successes []supervisorctl.AgentResult
	var failures []supervisorctl.AgentResult

	for _, opResult := range result.Results {
		agentResult := supervisorctl.AgentResult{
			AgentName: opResult.AgentName,
			Success:   opResult.IsSuccessful(),
		}

		if opResult.IsSuccessful() {
			agentResult.Message = opResult.Message
			successes = append(successes, agentResult)
		} else {
			agentResult.Error = opResult.Details
			if opResult.Details == "" {
				agentResult.Error = opResult.Message
			}
			failures = append(failures, agentResult)
		}
	}

	return &supervisorctl.OperationResult{
		Operation: result.Operation.String(),
		Successes: successes,
		Failures:  failures,
		Summary: supervisorctl.OperationSummary{
			Total:     result.TotalAgents,
			Succeeded: result.SuccessfulOps,
			Failed:    result.FailedOps,
			Duration:  result.Duration.String(),
		},
	}
}