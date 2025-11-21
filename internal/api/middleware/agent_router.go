package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AgentRouterMiddleware creates a middleware that handles agent-based routing
// This middleware extracts agent information from the request path and adds it to the context
type AgentRouterMiddleware struct {
	logger *zap.Logger
}

// NewAgentRouterMiddleware creates a new instance of AgentRouterMiddleware
func NewAgentRouterMiddleware(logger *zap.Logger) *AgentRouterMiddleware {
	return &AgentRouterMiddleware{
		logger: logger,
	}
}

// AgentRouter handles the routing of requests to specific agents based on the path
func (arm *AgentRouterMiddleware) AgentRouter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract agent ID from path (assuming format /agents/{agentId}/v1/...)
		path := c.Request.URL.Path
		agentID := extractAgentIDFromPath(path)
		
		if agentID != "" {
			// Add agent ID to context for use by downstream handlers
			c.Set("agent_id", agentID)
			
			if arm.logger != nil {
				arm.logger.Info("routing request to agent",
					zap.String("agent_id", agentID),
					zap.String("path", path),
					zap.String("method", c.Request.Method))
			}
		} else if arm.logger != nil {
			arm.logger.Info("no agent ID found in path",
				zap.String("path", path),
				zap.String("method", c.Request.Method))
		}
		
		// Continue with the request
		c.Next()
	}
}

// extractAgentIDFromPath extracts the agent ID from the path
// Format expected: /agents/{agentId}/v1/... or /agents/{agentId}/v1/.../{other_path}
func extractAgentIDFromPath(path string) string {
	// Split the path into segments
	segments := strings.Split(path, "/")
	
	// Look for the pattern /agents/{agentId}/v1/
	for i := 0; i < len(segments)-2; i++ {
		if segments[i] == "agents" && len(segments[i+2]) >= 2 && segments[i+2][:2] == "v1" {
			// The agent ID is in the next segment
			return segments[i+1]
		}
	}
	
	return ""
}

// AgentValidatorMiddleware creates a middleware that validates agent configuration
// before passing the request to the handler
func (arm *AgentRouterMiddleware) AgentValidatorMiddleware(agentProvider AgentProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID, exists := c.Get("agent_id")
		if !exists {
			if arm.logger != nil {
				arm.logger.Error("agent_id not found in context")
			}
			c.JSON(400, gin.H{"error": "Invalid request path"})
			c.Abort()
			return
		}
		
		agentIDStr, ok := agentID.(string)
		if !ok {
			if arm.logger != nil {
				arm.logger.Error("agent_id is not a string in context")
			}
			c.JSON(400, gin.H{"error": "Invalid agent ID format"})
			c.Abort()
			return
		}
		
		// Validate that the agent exists and is enabled
		agent, err := agentProvider.GetAgent(agentIDStr)
		if err != nil || agent == nil {
			if arm.logger != nil {
				arm.logger.Error("agent not found or invalid", zap.String("agent_id", agentIDStr))
			}
			c.JSON(404, gin.H{"error": "Agent not found", "code": -32001})
			c.Abort()
			return
		}
		
		if !agent.Enabled {
			if arm.logger != nil {
				arm.logger.Warn("agent is disabled", zap.String("agent_id", agentIDStr))
			}
			c.JSON(403, gin.H{"error": "Agent is disabled"})
			c.Abort()
			return
		}
		
		// Add the agent to the context for use by downstream handlers
		c.Set("agent", agent)
		
		// Continue with the request
		c.Next()
	}
}

// AgentProvider interface for retrieving agent configurations
type AgentProvider interface {
	GetAgent(agentID string) (*AgentConfiguration, error)
}

// AgentConfiguration represents the configuration of an agent
// This is a simplified version - in a real implementation, it would match the internal model
type AgentConfiguration struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Enabled        bool   `json:"enabled"`
	AccessType     string `json:"access_type"` // read-only or read-write
	MaxConcurrent  int    `json:"max_concurrent_executions"`
}