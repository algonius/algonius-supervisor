package handlers

import (
	"net/http"

	"github.com/algonius/algonius-supervisor/internal/a2a"
	"github.com/algonius/algonius-supervisor/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// A2AHandlers handles A2A protocol requests
type A2AHandlers struct {
	a2aService *services.A2AService
	logger     *zap.Logger
	config     *a2a.A2AConfig
}

// NewA2AHandlers creates a new instance of A2AHandlers
func NewA2AHandlers(a2aService *services.A2AService, logger *zap.Logger, config *a2a.A2AConfig) *A2AHandlers {
	return &A2AHandlers{
		a2aService: a2aService,
		logger:     logger,
		config:     config,
	}
}

// RegisterA2ARoutes registers all A2A-related routes
func (ah *A2AHandlers) RegisterA2ARoutes(router *gin.Engine) {
	// Apply A2A-specific middleware
	a2aGroup := router.Group("/agents")
	a2aGroup.Use(a2a.AuthenticationMiddleware(ah.config))
	a2aGroup.Use(a2a.CORSMiddleware(ah.config))
	a2aGroup.Use(a2a.RequestValidationMiddleware(ah.config))

	// Register A2A protocol routes
	a2aGroup.POST("/:agentId/v1/message/send", ah.HandleSendMessage)
	a2aGroup.POST("/:agentId/v1/message/stream", ah.HandleStreamMessage) // Note: This would need SSE or WebSocket implementation
	a2aGroup.GET("/:agentId/v1/.well-known/agent-card.json", ah.GetAgentCard)
	a2aGroup.GET("/:agentId/v1/tasks", ah.ListTasks)
	a2aGroup.GET("/:agentId/v1/tasks/:taskId", ah.GetTask)
}

// HandleSendMessage handles sending messages to agents via A2A protocol
func (ah *A2AHandlers) HandleSendMessage(c *gin.Context) {
	agentID := c.Param("agentId")

	// Log the incoming request
	ah.logger.Info("handling A2A send message request",
		zap.String("agent_id", agentID),
		zap.String("path", c.Request.URL.Path))

	// For now, we'll just return a placeholder response
	// In a real implementation, you would validate the request against A2A protocol,
	// call the A2A service to process it, and return the response

	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		ah.logger.Error("failed to parse A2A request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON in request body",
			"code":  -32700, // Parse error according to A2A spec
		})
		return
	}

	// In a real implementation, you would process the request using the A2A service
	// and return an appropriate A2A response
	response := gin.H{
		"protocol":     "a2a",
		"version":      "0.3.0",
		"id":           "response-123",
		"type":         "response",
		"timestamp":    "2025-11-20T10:30:05Z",
		"inResponseTo": c.Param("agentId"),
		"context": gin.H{
			"from": agentID,
			"to":   "requester",
		},
		"payload": gin.H{
			"result": gin.H{
				"status":  "success",
				"message": "Message received and processed",
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// HandleStreamMessage handles streaming messages to agents via A2A protocol
func (ah *A2AHandlers) HandleStreamMessage(c *gin.Context) {
	agentID := c.Param("agentId")

	// Log the incoming request
	ah.logger.Info("handling A2A stream message request",
		zap.String("agent_id", agentID),
		zap.String("path", c.Request.URL.Path))

	// For streaming, we would need to implement Server-Sent Events (SSE) or WebSockets
	// For now, return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message":  "Streaming not implemented yet",
		"agent_id": agentID,
	})
}

// GetAgentCard returns the agent card for discovery (as per A2A spec)
func (ah *A2AHandlers) GetAgentCard(c *gin.Context) {
	agentID := c.Param("agentId")

	// Log the incoming request
	ah.logger.Info("handling A2A agent card request",
		zap.String("agent_id", agentID),
		zap.String("path", c.Request.URL.Path))

	// In a real implementation, you would get the agent details and return them in A2A agent card format
	// For now, return a placeholder
	agentCard := gin.H{
		"agentId": agentID,
		"name":    "Sample Agent",
		"version": "1.0.0",
		"capabilities": gin.H{
			"supportedMethods":      []string{"execute-agent", "status", "list-agents"},
			"streamingSupport":      true,
			"concurrentExecution":   false,
			"supportedContentTypes": []string{"text/plain", "application/json"},
		},
		"endpoints": []gin.H{
			{
				"protocol": "http_json",
				"url":      c.Request.URL.Scheme + "://" + c.Request.Host + "/agents/" + agentID + "/v1",
			},
		},
		"authentication": gin.H{
			"required": true,
			"methods":  []string{"bearer_token"},
		},
	}

	c.JSON(http.StatusOK, agentCard)
}

// ListTasks returns a list of tasks for an agent
func (ah *A2AHandlers) ListTasks(c *gin.Context) {
	agentID := c.Param("agentId")

	// Log the incoming request
	ah.logger.Info("handling A2A list tasks request",
		zap.String("agent_id", agentID),
		zap.String("path", c.Request.URL.Path))

	// For now, return a placeholder response
	tasks := gin.H{
		"tasks":  []gin.H{},
		"total":  0,
		"limit":  10,
		"offset": 0,
	}

	c.JSON(http.StatusOK, tasks)
}

// GetTask returns details for a specific task
func (ah *A2AHandlers) GetTask(c *gin.Context) {
	agentID := c.Param("agentId")
	taskID := c.Param("taskId")

	// Log the incoming request
	ah.logger.Info("handling A2A get task request",
		zap.String("agent_id", agentID),
		zap.String("task_id", taskID),
		zap.String("path", c.Request.URL.Path))

	// For now, return a placeholder response
	task := gin.H{
		"id":        taskID,
		"status":    "completed",
		"input":     "sample input",
		"output":    "sample output",
		"startTime": "2025-11-20T10:30:00Z",
		"endTime":   "2025-11-20T10:30:05Z",
		"duration":  5000,
		"exitCode":  0,
	}

	c.JSON(http.StatusOK, task)
}
