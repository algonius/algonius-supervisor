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

// MockOperationResult represents the expected response format for lifecycle operations
type MockOperationResult struct {
	AgentName  string `json:"agent_name"`
	Operation  string `json:"operation"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Timestamp  string `json:"timestamp"`
	PreviousState string `json:"previous_state,omitempty"`
	NewState   string `json:"new_state,omitempty"`
}

// MockLifecycleAPI creates a mock supervisor HTTP API for lifecycle operations
func MockLifecycleAPI() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock agent states
	agentStates := map[string]string{
		"running-agent":  "RUNNING",
		"stopped-agent":  "STOPPED",
		"failed-agent":   "FATAL",
		"starting-agent": "STARTING",
	}

	// Start agent endpoint - POST /api/v1/agents/:name/start
	router.POST("/api/v1/agents/:name/start", func(c *gin.Context) {
		name := c.Param("name")

		currentState, exists := agentStates[name]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   fmt.Sprintf("agent '%s' not found", name),
			})
			return
		}

		if currentState == "RUNNING" {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   fmt.Sprintf("agent '%s' is already running", name),
			})
			return
		}

		// Update state
		agentStates[name] = "STARTING"

		result := MockOperationResult{
			AgentName:    name,
			Operation:    "start",
			Success:      true,
			Message:      fmt.Sprintf("Agent '%s' start initiated", name),
			Timestamp:    "2025-11-21T10:30:00Z",
			PreviousState: currentState,
			NewState:     "STARTING",
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	// Stop agent endpoint - POST /api/v1/agents/:name/stop
	router.POST("/api/v1/agents/:name/stop", func(c *gin.Context) {
		name := c.Param("name")

		currentState, exists := agentStates[name]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   fmt.Sprintf("agent '%s' not found", name),
			})
			return
		}

		if currentState == "STOPPED" {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   fmt.Sprintf("agent '%s' is already stopped", name),
			})
			return
		}

		// Update state
		agentStates[name] = "STOPPED"

		result := MockOperationResult{
			AgentName:    name,
			Operation:    "stop",
			Success:      true,
			Message:      fmt.Sprintf("Agent '%s' stopped successfully", name),
			Timestamp:    "2025-11-21T10:30:00Z",
			PreviousState: currentState,
			NewState:     "STOPPED",
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	// Restart agent endpoint - POST /api/v1/agents/:name/restart
	router.POST("/api/v1/agents/:name/restart", func(c *gin.Context) {
		name := c.Param("name")

		currentState, exists := agentStates[name]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   fmt.Sprintf("agent '%s' not found", name),
			})
			return
		}

		// Update state to starting
		agentStates[name] = "STARTING"

		result := MockOperationResult{
			AgentName:    name,
			Operation:    "restart",
			Success:      true,
			Message:      fmt.Sprintf("Agent '%s' restart initiated", name),
			Timestamp:    "2025-11-21T10:30:00Z",
			PreviousState: currentState,
			NewState:     "STARTING",
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	return router
}

func TestLifecycleCommands_Contract_StartAgent(t *testing.T) {
	mockServer := httptest.NewServer(MockLifecycleAPI())
	defer mockServer.Close()

	t.Run("start stopped agent", func(t *testing.T) {
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/agents/stopped-agent/start", mockServer.URL), "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response struct {
			Success bool                `json:"success"`
			Data    MockOperationResult `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "stopped-agent", response.Data.AgentName)
		assert.Equal(t, "start", response.Data.Operation)
		assert.Equal(t, "STARTING", response.Data.NewState)
		assert.Equal(t, "STOPPED", response.Data.PreviousState)
	})

	t.Run("start non-existent agent", func(t *testing.T) {
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/agents/nonexistent/start", mockServer.URL), "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

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

	t.Run("start already running agent", func(t *testing.T) {
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/agents/running-agent/start", mockServer.URL), "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var response struct {
			Success bool `json:"success"`
			Error   string `json:"error"`
		}

		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "already running")
	})
}

func TestLifecycleCommands_Contract_StopAgent(t *testing.T) {
	mockServer := httptest.NewServer(MockLifecycleAPI())
	defer mockServer.Close()

	t.Run("stop running agent", func(t *testing.T) {
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/agents/running-agent/stop", mockServer.URL), "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response struct {
			Success bool                `json:"success"`
			Data    MockOperationResult `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "running-agent", response.Data.AgentName)
		assert.Equal(t, "stop", response.Data.Operation)
		assert.Equal(t, "STOPPED", response.Data.NewState)
		assert.Equal(t, "RUNNING", response.Data.PreviousState)
	})

	t.Run("stop non-existent agent", func(t *testing.T) {
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/agents/nonexistent/stop", mockServer.URL), "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

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

	t.Run("stop already stopped agent", func(t *testing.T) {
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/agents/stopped-agent/stop", mockServer.URL), "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var response struct {
			Success bool `json:"success"`
			Error   string `json:"error"`
		}

		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.False(t, response.Success)
		assert.Contains(t, response.Error, "already stopped")
	})
}

func TestLifecycleCommands_Contract_RestartAgent(t *testing.T) {
	mockServer := httptest.NewServer(MockLifecycleAPI())
	defer mockServer.Close()

	t.Run("restart running agent", func(t *testing.T) {
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/agents/running-agent/restart", mockServer.URL), "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response struct {
			Success bool                `json:"success"`
			Data    MockOperationResult `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "running-agent", response.Data.AgentName)
		assert.Equal(t, "restart", response.Data.Operation)
		assert.Equal(t, "STARTING", response.Data.NewState)
		assert.Equal(t, "RUNNING", response.Data.PreviousState)
	})

	t.Run("restart stopped agent", func(t *testing.T) {
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/agents/stopped-agent/restart", mockServer.URL), "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response struct {
			Success bool                `json:"success"`
			Data    MockOperationResult `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.True(t, response.Success)
		assert.Equal(t, "stopped-agent", response.Data.AgentName)
		assert.Equal(t, "restart", response.Data.Operation)
		assert.Equal(t, "STARTING", response.Data.NewState)
		assert.Equal(t, "STOPPED", response.Data.PreviousState)
	})

	t.Run("restart non-existent agent", func(t *testing.T) {
		resp, err := http.Post(fmt.Sprintf("%s/api/v1/agents/nonexistent/restart", mockServer.URL), "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

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

func TestLifecycleCommands_Contract_Authentication(t *testing.T) {
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

	// Protected lifecycle endpoints
	router.POST("/api/v1/agents/:name/start", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": MockOperationResult{
				AgentName: c.Param("name"),
				Operation: "start",
				Success:   true,
			},
		})
	})

	router.POST("/api/v1/agents/:name/stop", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": MockOperationResult{
				AgentName: c.Param("name"),
				Operation: "stop",
				Success:   true,
			},
		})
	})

	router.POST("/api/v1/agents/:name/restart", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": MockOperationResult{
				AgentName: c.Param("name"),
				Operation: "restart",
				Success:   true,
			},
		})
	})

	mockServer := httptest.NewServer(router)
	defer mockServer.Close()

	testCases := []struct {
		name        string
		endpoint    string
		authHeader  string
		expectCode  int
		expectError bool
	}{
		{
			name:       "start with valid token",
			endpoint:   "/api/v1/agents/test/start",
			authHeader: "Bearer valid-token",
			expectCode: http.StatusOK,
		},
		{
			name:       "start without token",
			endpoint:   "/api/v1/agents/test/start",
			authHeader: "",
			expectCode: http.StatusUnauthorized,
			expectError: true,
		},
		{
			name:       "stop with invalid token",
			endpoint:   "/api/v1/agents/test/stop",
			authHeader: "Bearer invalid-token",
			expectCode: http.StatusUnauthorized,
			expectError: true,
		},
		{
			name:       "restart with valid token",
			endpoint:   "/api/v1/agents/test/restart",
			authHeader: "Bearer valid-token",
			expectCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", mockServer.URL, tc.endpoint), nil)
			require.NoError(t, err)

			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectCode, resp.StatusCode)

			if tc.expectError {
				var response struct {
					Success bool `json:"success"`
					Error   string `json:"error"`
				}

				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.Contains(t, response.Error, "unauthorized")
			}
		})
	}
}

func TestLifecycleCommands_Contract_ErrorHandling(t *testing.T) {
	// Setup mock server that returns internal server error
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.POST("/api/v1/agents/:name/start", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error: failed to start agent",
		})
	})

	router.POST("/api/v1/agents/:name/stop", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error: failed to stop agent",
		})
	})

	router.POST("/api/v1/agents/:name/restart", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error: failed to restart agent",
		})
	})

	mockServer := httptest.NewServer(router)
	defer mockServer.Close()

	testCases := []struct {
		name     string
		endpoint string
		errorMsg string
	}{
		{
			name:     "start agent error",
			endpoint: "/api/v1/agents/test/start",
			errorMsg: "failed to start agent",
		},
		{
			name:     "stop agent error",
			endpoint: "/api/v1/agents/test/stop",
			errorMsg: "failed to stop agent",
		},
		{
			name:     "restart agent error",
			endpoint: "/api/v1/agents/test/restart",
			errorMsg: "failed to restart agent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Post(fmt.Sprintf("%s%s", mockServer.URL, tc.endpoint), "application/json", nil)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

			var response struct {
				Success bool `json:"success"`
				Error   string `json:"error"`
			}

			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.False(t, response.Success)
			assert.Contains(t, response.Error, tc.errorMsg)
		})
	}
}

func TestLifecycleCommands_Contract_RequestValidation(t *testing.T) {
	// Setup mock server for validation testing
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Test endpoints that validate request format
	router.POST("/api/v1/agents/:name/start", func(c *gin.Context) {
		// Validate content type if provided
		contentType := c.GetHeader("Content-Type")
		if contentType != "" && contentType != "application/json" {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"success": false,
				"error":   "unsupported media type: application/json required",
			})
			return
		}

		// Simulate successful start
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": MockOperationResult{
				AgentName: c.Param("name"),
				Operation: "start",
				Success:   true,
			},
		})
	})

	mockServer := httptest.NewServer(router)
	defer mockServer.Close()

	t.Run("valid content type", func(t *testing.T) {
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/agents/test/start", mockServer.URL), bytes.NewBuffer([]byte("{}")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("invalid content type", func(t *testing.T) {
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/agents/test/start", mockServer.URL), bytes.NewBuffer([]byte("{}")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
	})
}