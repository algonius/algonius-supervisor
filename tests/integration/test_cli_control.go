package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSupervisorIntegrationAPI creates a more comprehensive mock API for integration testing
func MockSupervisorIntegrationAPI() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Authentication middleware
	router.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "Bearer test-token" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized: invalid or missing token",
			})
			c.Abort()
			return
		}
		c.Next()
	})

	// Mock agent storage
	type Agent struct {
		Name      string    `json:"name"`
		State     string    `json:"state"`
		PID       int       `json:"pid"`
		Uptime    string    `json:"uptime"`
		Memory    string    `json:"memory"`
		CPU       string    `json:"cpu"`
		StartTime time.Time `json:"start_time"`
	}

	agents := make(map[string]*Agent)

	// Initialize with some default agents
	agents["web-server"] = &Agent{
		Name:   "web-server",
		State:  "RUNNING",
		PID:    1001,
		Uptime: "2:15:30",
		Memory: "45.2MB",
		CPU:    "3.1%",
		StartTime: time.Now().Add(-2 * time.Hour),
	}

	agents["db-worker"] = &Agent{
		Name:   "db-worker",
		State:  "STOPPED",
		Uptime: "0:00:00",
		Memory: "0MB",
		CPU:    "0.0%",
		StartTime: time.Time{},
	}

	agents["api-gateway"] = &Agent{
		Name:   "api-gateway",
		State:  "FATAL",
		PID:    0,
		Uptime: "0:05:12",
		Memory: "12.8MB",
		CPU:    "0.0%",
		StartTime: time.Now().Add(-5 * time.Minute),
	}

	// Status endpoints
	router.GET("/api/v1/agents/status", func(c *gin.Context) {
		var agentList []Agent
		for _, agent := range agents {
			agentList = append(agentList, *agent)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    agentList,
		})
	})

	router.GET("/api/v1/agents/:name/status", func(c *gin.Context) {
		name := c.Param("name")
		agent, exists := agents[name]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   fmt.Sprintf("agent '%s' not found", name),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    agent,
		})
	})

	// Lifecycle endpoints
	router.POST("/api/v1/agents/:name/start", func(c *gin.Context) {
		name := c.Param("name")
		agent, exists := agents[name]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   fmt.Sprintf("agent '%s' not found", name),
			})
			return
		}

		if agent.State == "RUNNING" {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   fmt.Sprintf("agent '%s' is already running", name),
			})
			return
		}

		// Simulate agent start
		agent.State = "STARTING"
		agent.StartTime = time.Now()

		// Simulate transition to running after a delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			agent.State = "RUNNING"
			agent.PID = 2000 + len(agents)
			agent.Uptime = "0:00:01"
			agent.Memory = "8.5MB"
			agent.CPU = "1.2%"
		}()

		result := map[string]interface{}{
			"agent_name":    name,
			"operation":     "start",
			"success":       true,
			"message":       fmt.Sprintf("Agent '%s' start initiated", name),
			"timestamp":     time.Now().Format(time.RFC3339),
			"previous_state": "STOPPED",
			"new_state":     "STARTING",
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	router.POST("/api/v1/agents/:name/stop", func(c *gin.Context) {
		name := c.Param("name")
		agent, exists := agents[name]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   fmt.Sprintf("agent '%s' not found", name),
			})
			return
		}

		if agent.State == "STOPPED" {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   fmt.Sprintf("agent '%s' is already stopped", name),
			})
			return
		}

		previousState := agent.State
		agent.State = "STOPPED"
		agent.PID = 0
		agent.Uptime = "0:00:00"
		agent.Memory = "0MB"
		agent.CPU = "0.0%"
		agent.StartTime = time.Time{}

		result := map[string]interface{}{
			"agent_name":    name,
			"operation":     "stop",
			"success":       true,
			"message":       fmt.Sprintf("Agent '%s' stopped successfully", name),
			"timestamp":     time.Now().Format(time.RFC3339),
			"previous_state": previousState,
			"new_state":     "STOPPED",
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	router.POST("/api/v1/agents/:name/restart", func(c *gin.Context) {
		name := c.Param("name")
		agent, exists := agents[name]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   fmt.Sprintf("agent '%s' not found", name),
			})
			return
		}

		previousState := agent.State
		agent.State = "STARTING"
		agent.StartTime = time.Now()

		// Simulate restart process
		go func() {
			time.Sleep(200 * time.Millisecond)
			agent.State = "RUNNING"
			agent.PID = 3000 + len(agents)
			agent.Uptime = "0:00:01"
			agent.Memory = "10.2MB"
			agent.CPU = "2.1%"
		}()

		result := map[string]interface{}{
			"agent_name":    name,
			"operation":     "restart",
			"success":       true,
			"message":       fmt.Sprintf("Agent '%s' restart initiated", name),
			"timestamp":     time.Now().Format(time.RFC3339),
			"previous_state": previousState,
			"new_state":     "STARTING",
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	return router
}

func TestCLIControl_Integration_BasicWorkflow(t *testing.T) {
	// Setup mock server
	mockServer := httptest.NewServer(MockSupervisorIntegrationAPI())
	defer mockServer.Close()

	// Create temporary config file for CLI
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	configContent := fmt.Sprintf(`
server:
  url: %s
  timeout: 30s
  auth_token: test-token

display:
  format: table
  colors: false

defaults:
  restart_attempts: 3
  wait_time: 5s
`, mockServer.URL)

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test workflow: Check status -> Start agent -> Verify status -> Stop agent -> Verify status
	t.Run("complete agent lifecycle workflow", func(t *testing.T) {
		ctx := context.Background()

		// Step 1: Check initial status
		initialStatus, err := getAgentStatus(ctx, mockServer.URL, "db-worker", "test-token")
		require.NoError(t, err)
		assert.Equal(t, "STOPPED", initialStatus.State)

		// Step 2: Start the agent
		startResult, err := startAgent(ctx, mockServer.URL, "db-worker", "test-token")
		require.NoError(t, err)
		assert.True(t, startResult.Success)
		assert.Equal(t, "start", startResult.Operation)
		assert.Equal(t, "STOPPED", startResult.PreviousState)
		assert.Equal(t, "STARTING", startResult.NewState)

		// Step 3: Wait for agent to be fully started and check status
		time.Sleep(200 * time.Millisecond) // Wait for mock async transition
		runningStatus, err := getAgentStatus(ctx, mockServer.URL, "db-worker", "test-token")
		require.NoError(t, err)
		assert.Equal(t, "RUNNING", runningStatus.State)
		assert.Greater(t, runningStatus.PID, 0)

		// Step 4: Stop the agent
		stopResult, err := stopAgent(ctx, mockServer.URL, "db-worker", "test-token")
		require.NoError(t, err)
		assert.True(t, stopResult.Success)
		assert.Equal(t, "stop", stopResult.Operation)
		assert.Equal(t, "RUNNING", stopResult.PreviousState)
		assert.Equal(t, "STOPPED", stopResult.NewState)

		// Step 5: Verify final status
		finalStatus, err := getAgentStatus(ctx, mockServer.URL, "db-worker", "test-token")
		require.NoError(t, err)
		assert.Equal(t, "STOPPED", finalStatus.State)
		assert.Equal(t, 0, finalStatus.PID)
	})
}

func TestCLIControl_Integration_ErrorScenarios(t *testing.T) {
	mockServer := httptest.NewServer(MockSupervisorIntegrationAPI())
	defer mockServer.Close()

	ctx := context.Background()

	t.Run("operate on non-existent agent", func(t *testing.T) {
		// Try to get status of non-existent agent
		_, err := getAgentStatus(ctx, mockServer.URL, "non-existent", "test-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent 'non-existent' not found")

		// Try to start non-existent agent
		_, err = startAgent(ctx, mockServer.URL, "non-existent", "test-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent 'non-existent' not found")
	})

	t.Run("invalid authentication", func(t *testing.T) {
		// Try to access with invalid token
		_, err := getAgentStatus(ctx, mockServer.URL, "web-server", "invalid-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")

		_, err = startAgent(ctx, mockServer.URL, "db-worker", "invalid-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("duplicate operations", func(t *testing.T) {
		// Try to start already running agent
		_, err := startAgent(ctx, mockServer.URL, "web-server", "test-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already running")

		// Try to stop already stopped agent
		_, err = stopAgent(ctx, mockServer.URL, "db-worker", "test-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already stopped")
	})
}

func TestCLIControl_Integration_RestartWorkflow(t *testing.T) {
	mockServer := httptest.NewServer(MockSupervisorIntegrationAPI())
	defer mockServer.Close()

	ctx := context.Background()

	t.Run("restart running agent", func(t *testing.T) {
		// Get initial status
		initialStatus, err := getAgentStatus(ctx, mockServer.URL, "web-server", "test-token")
		require.NoError(t, err)
		assert.Equal(t, "RUNNING", initialStatus.State)

		// Restart the agent
		restartResult, err := restartAgent(ctx, mockServer.URL, "web-server", "test-token")
		require.NoError(t, err)
		assert.True(t, restartResult.Success)
		assert.Equal(t, "restart", restartResult.Operation)
		assert.Equal(t, "RUNNING", restartResult.PreviousState)
		assert.Equal(t, "STARTING", restartResult.NewState)

		// Wait for restart to complete
		time.Sleep(300 * time.Millisecond)

		// Verify agent is running again
		finalStatus, err := getAgentStatus(ctx, mockServer.URL, "web-server", "test-token")
		require.NoError(t, err)
		assert.Equal(t, "RUNNING", finalStatus.State)
	})

	t.Run("restart stopped agent", func(t *testing.T) {
		// Ensure agent is stopped first
		_, err := stopAgent(ctx, mockServer.URL, "db-worker", "test-token")
		require.NoError(t, err)

		// Restart the stopped agent
		restartResult, err := restartAgent(ctx, mockServer.URL, "db-worker", "test-token")
		require.NoError(t, err)
		assert.True(t, restartResult.Success)
		assert.Equal(t, "restart", restartResult.Operation)
		assert.Equal(t, "STOPPED", restartResult.PreviousState)

		// Wait for restart to complete
		time.Sleep(300 * time.Millisecond)

		// Verify agent is running
		finalStatus, err := getAgentStatus(ctx, mockServer.URL, "db-worker", "test-token")
		require.NoError(t, err)
		assert.Equal(t, "RUNNING", finalStatus.State)
	})
}

func TestCLIControl_Integration_MultipleAgents(t *testing.T) {
	mockServer := httptest.NewServer(MockSupervisorIntegrationAPI())
	defer mockServer.Close()

	ctx := context.Background()

	t.Run("manage multiple agents simultaneously", func(t *testing.T) {
		// Get all agents status
		allStatus, err := getAllAgentsStatus(ctx, mockServer.URL, "test-token")
		require.NoError(t, err)
		assert.Len(t, allStatus, 3)

		// Verify initial states
		agentStates := make(map[string]string)
		for _, agent := range allStatus {
			agentStates[agent.Name] = agent.State
		}

		assert.Equal(t, "RUNNING", agentStates["web-server"])
		assert.Equal(t, "STOPPED", agentStates["db-worker"])
		assert.Equal(t, "FATAL", agentStates["api-gateway"])

		// Start stopped agent
		_, err = startAgent(ctx, mockServer.URL, "db-worker", "test-token")
		require.NoError(t, err)

		// Stop running agent
		_, err = stopAgent(ctx, mockServer.URL, "web-server", "test-token")
		require.NoError(t, err)

		// Restart failed agent
		_, err = restartAgent(ctx, mockServer.URL, "api-gateway", "test-token")
		require.NoError(t, err)

		// Wait for operations to complete
		time.Sleep(300 * time.Millisecond)

		// Verify final states
		finalStatus, err := getAllAgentsStatus(ctx, mockServer.URL, "test-token")
		require.NoError(t, err)

		finalStates := make(map[string]string)
		for _, agent := range finalStatus {
			finalStates[agent.Name] = agent.State
		}

		assert.Equal(t, "STOPPED", finalStates["web-server"])
		assert.Equal(t, "RUNNING", finalStates["db-worker"])
		assert.Equal(t, "RUNNING", finalStates["api-gateway"])
	})
}

// Helper functions for HTTP API calls

type AgentStatus struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	PID    int    `json:"pid"`
	Uptime string `json:"uptime"`
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}

type OperationResult struct {
	AgentName     string `json:"agent_name"`
	Operation     string `json:"operation"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	Timestamp     string `json:"timestamp"`
	PreviousState string `json:"previous_state"`
	NewState      string `json:"new_state"`
}

func getAgentStatus(ctx context.Context, serverURL, agentName, token string) (*AgentStatus, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/agents/%s/status", serverURL, agentName), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return nil, fmt.Errorf("API error: %s", errorResp.Error)
	}

	var response struct {
		Success bool        `json:"success"`
		Data    AgentStatus `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

func getAllAgentsStatus(ctx context.Context, serverURL, token string) ([]AgentStatus, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/agents/status", serverURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	var response struct {
		Success bool         `json:"success"`
		Data    []AgentStatus `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func startAgent(ctx context.Context, serverURL, agentName, token string) (*OperationResult, error) {
	return performLifecycleOperation(ctx, "POST", fmt.Sprintf("%s/api/v1/agents/%s/start", serverURL, agentName), token)
}

func stopAgent(ctx context.Context, serverURL, agentName, token string) (*OperationResult, error) {
	return performLifecycleOperation(ctx, "POST", fmt.Sprintf("%s/api/v1/agents/%s/stop", serverURL, agentName), token)
}

func restartAgent(ctx context.Context, serverURL, agentName, token string) (*OperationResult, error) {
	return performLifecycleOperation(ctx, "POST", fmt.Sprintf("%s/api/v1/agents/%s/restart", serverURL, agentName), token)
}

func performLifecycleOperation(ctx context.Context, method, url, token string) (*OperationResult, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errorResp)
		return nil, fmt.Errorf("API error: %s", errorResp.Error)
	}

	var response struct {
		Success bool           `json:"success"`
		Data    OperationResult `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}