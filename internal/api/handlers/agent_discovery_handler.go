package handlers

import (
	"net/http"

	"github.com/algonius/algonius-supervisor/internal/services"
	"github.com/algonius/algonius-supervisor/internal/a2a"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AgentDiscoveryHandlers handles agent discovery requests
type AgentDiscoveryHandlers struct {
	agentService *services.AgentService
	logger       *zap.Logger
	config       *a2a.A2AConfig
}

// NewAgentDiscoveryHandlers creates a new instance of AgentDiscoveryHandlers
func NewAgentDiscoveryHandlers(agentService *services.AgentService, logger *zap.Logger, config *a2a.A2AConfig) *AgentDiscoveryHandlers {
	return &AgentDiscoveryHandlers{
		agentService: agentService,
		logger:       logger,
		config:       config,
	}
}

// RegisterAgentDiscoveryRoutes registers all agent discovery routes
func (adh *AgentDiscoveryHandlers) RegisterAgentDiscoveryRoutes(router *gin.Engine) {
	// Public routes for discovery (no authentication required for discovery)
	discoveryGroup := router.Group("/discovery")
	
	// Register discovery routes
	discoveryGroup.GET("/agents", adh.ListAgents)
	discoveryGroup.GET("/agents/:agentId", adh.GetAgent)
	discoveryGroup.GET("/capabilities", adh.GetCapabilities)
}

// ListAgents returns a list of all available agents
func (adh *AgentDiscoveryHandlers) ListAgents(c *gin.Context) {
	adh.logger.Info("handling agent list request",
		zap.String("path", c.Request.URL.Path))

	agents, err := adh.agentService.ListAgents()
	if err != nil {
		adh.logger.Error("failed to list agents", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list agents",
		})
		return
	}

	// Convert to response format
	agentList := make([]gin.H, len(agents))
	for i, agent := range agents {
		agentList[i] = gin.H{
			"id":             agent.ID,
			"name":           agent.Name,
			"agent_type":     agent.AgentType,
			"access_type":    string(agent.AccessType),
			"max_concurrent": agent.MaxConcurrentExecutions,
			"enabled":        agent.Enabled,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agentList,
		"total":  len(agentList),
	})
}

// GetAgent returns details for a specific agent
func (adh *AgentDiscoveryHandlers) GetAgent(c *gin.Context) {
	agentID := c.Param("agentId")
	
	adh.logger.Info("handling agent details request",
		zap.String("agent_id", agentID),
		zap.String("path", c.Request.URL.Path))

	agent, err := adh.agentService.GetAgent(agentID)
	if err != nil {
		adh.logger.Error("agent not found", 
			zap.String("agent_id", agentID), 
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent not found",
			"code": -32001, // A2A agent not found error code
		})
		return
	}

	response := gin.H{
		"id":                    agent.ID,
		"name":                  agent.Name,
		"agent_type":            agent.AgentType,
		"access_type":           string(agent.AccessType),
		"max_concurrent":        agent.MaxConcurrentExecutions,
		"mode":                  string(agent.Mode),
		"input_pattern":         string(agent.InputPattern),
		"output_pattern":        string(agent.OutputPattern),
		"enabled":               agent.Enabled,
		"created_at":            agent.CreatedAt,
		"updated_at":            agent.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// GetCapabilities returns the capabilities of the supervisor
func (adh *AgentDiscoveryHandlers) GetCapabilities(c *gin.Context) {
	adh.logger.Info("handling capabilities request",
		zap.String("path", c.Request.URL.Path))

	capabilities := gin.H{
		"supported_protocols": []string{
			"http_json",
			"grpc",
			"json_rpc",
		},
		"supported_methods": []string{
			"execute-agent",
			"status",
			"list-agents",
		},
		"max_concurrent_executions": 100, // This could be configurable
		"a2a_protocol_version":      "0.3.0",
		"service_version":           "1.0.0",
		"features": gin.H{
			"agent_configuration": true,
			"scheduled_tasks":     true,
			"concurrent_execution": true,
			"resource_monitoring":  true,
		},
	}

	c.JSON(http.StatusOK, capabilities)
}