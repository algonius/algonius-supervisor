package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAgentStatus represents the expected response format for agent status
type MockAgentStatus struct {
	Name    string `json:"name"`
	State   string `json:"state"`
	PID     int    `json:"pid,omitempty"`
	Uptime  string `json:"uptime"`
	Mem     string `json:"memory"`
	CPU     string `json:"cpu"`
}

// MockSupervisorAPI creates a mock supervisor HTTP API for testing
func MockSupervisorAPI() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock agents data
	mockAgents := map[string]MockAgentStatus{
		"agent1": {
			Name:   "agent1",
			State:  "RUNNING",
			PID:    1234,
			Uptime: "0:10:30",
			Mem:    "15.2MB",
			CPU:    "2.1%",
		},
		"agent2": {
			Name:   "agent2",
			State:  "STOPPED",
			Uptime: "0:00:00",
			Mem:    "0MB",
			CPU:    "0.0%",
		},
		"test-agent": {
			Name:   "test-agent",
			State:  "FATAL",
			PID:    0,
			Uptime: "0:05:15",
			Mem:    "8.7MB",
			CPU:    "1.2%",
		},
	}

	// Status endpoint - GET /api/v1/agents/status
	router.GET("/api/v1/agents/status", func(c *gin.Context) {
		var agents []MockAgentStatus
		for _, agent := range mockAgents {
			agents = append(agents, agent)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    agents,
		})
	})

	// Single agent status endpoint - GET /api/v1/agents/:name/status
	router.GET("/api/v1/agents/:name/status", func(c *gin.Context) {
		name := c.Param("name")

		agent, exists := mockAgents[name]
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

	return router
}

func TestStatusCommand_Contract_AllAgents(t *testing.T) {
	// Setup mock server
	mockServer := httptest.NewServer(MockSupervisorAPI())
	defer mockServer.Close()

	// Test client request to get all agents status
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/agents/status", mockServer.URL))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Success bool               `json:"success"`
		Data    []MockAgentStatus  `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Len(t, response.Data, 3) // Should have 3 mock agents

	// Verify agent data structure
	for _, agent := range response.Data {
		assert.NotEmpty(t, agent.Name)
		assert.NotEmpty(t, agent.State)
		assert.NotEmpty(t, agent.Uptime)
	}

	// Verify specific agents exist
	agentNames := make(map[string]bool)
	for _, agent := range response.Data {
		agentNames[agent.Name] = true
	}

	assert.True(t, agentNames["agent1"])
	assert.True(t, agentNames["agent2"])
	assert.True(t, agentNames["test-agent"])
}

func TestStatusCommand_Contract_SingleAgent(t *testing.T) {
	// Setup mock server
	mockServer := httptest.NewServer(MockSupervisorAPI())
	defer mockServer.Close()

	t.Run("existing agent", func(t *testing.T) {
		// Test client request to get specific agent status
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/agents/agent1/status", mockServer.URL))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response struct {
			Success bool          `json:"success"`
			Data    MockAgentStatus `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "agent1", response.Data.Name)
		assert.Equal(t, "RUNNING", response.Data.State)
		assert.Equal(t, 1234, response.Data.PID)
	})

	t.Run("non-existent agent", func(t *testing.T) {
		// Test client request to non-existent agent
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/agents/nonexistent/status", mockServer.URL))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify response
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var response struct {
			Success bool `json:"success"`
			Error   string `json:"error"`
		}

		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "agent 'nonexistent' not found")
	})
}

func TestStatusCommand_Contract_Authentication(t *testing.T) {
	// Setup mock server with authentication
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "Bearer valid-token" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized: invalid or missing token",
			})
			c.Abort()
			return
		}
		c.Next()
	})

	// Protected status endpoint
	router.GET("/api/v1/agents/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []MockAgentStatus{},
		})
	})

	mockServer := httptest.NewServer(router)
	defer mockServer.Close()

	t.Run("with valid token", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/agents/status", mockServer.URL), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer valid-token")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("without token", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/agents/status", mockServer.URL))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("with invalid token", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/agents/status", mockServer.URL), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestStatusCommand_Contract_ErrorHandling(t *testing.T) {
	// Setup mock server that returns errors
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Simulate internal server error
	router.GET("/api/v1/agents/status", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error: failed to retrieve agent status",
		})
	})

	mockServer := httptest.NewServer(router)
	defer mockServer.Close()

	// Test client request
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/agents/status", mockServer.URL))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify error response
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var response struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "internal server error")
}

func TestStatusCommand_Contract_ResponseFormat(t *testing.T) {
	// Setup mock server
	mockServer := httptest.NewServer(MockSupervisorAPI())
	defer mockServer.Close()

	resp, err := http.Get(fmt.Sprintf("%s/api/v1/agents/status", mockServer.URL))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify content type
	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

	// Verify response is valid JSON
	var jsonResponse interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	require.NoError(t, err)

	// Verify response structure
	responseMap, ok := jsonResponse.(map[string]interface{})
	require.True(t, ok)

	// Must have success field (boolean)
	_, hasSuccess := responseMap["success"]
	require.True(t, hasSuccess)

	// Must have data field (array)
	_, hasData := responseMap["data"]
	require.True(t, hasData)
}