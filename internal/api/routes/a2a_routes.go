package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/algonius/algonius-supervisor/internal/api/handlers"
	"github.com/algonius/algonius-supervisor/internal/a2a"
	"github.com/algonius/algonius-supervisor/internal/services"
)

// A2ARouteConfig holds the configuration for A2A routes
type A2ARouteConfig struct {
	Router              *gin.Engine
	A2AService          *services.A2AService
	AgentService        *services.AgentService
	ExecutionService    services.IExecutionService
	A2AConfig           *a2a.A2AConfig
}

// SetupA2ARoutes sets up all A2A-related routes
func SetupA2ARoutes(config *A2ARouteConfig) {
	// Create handlers
	a2aHandler := handlers.NewA2AHandlers(config.A2AService, config.A2AService.GetLogger(), config.A2AConfig)
	agentDiscoveryHandler := handlers.NewAgentDiscoveryHandlers(config.AgentService, config.A2AService.GetLogger(), config.A2AConfig)
	jsonrpcHandler := handlers.NewJSONRPCHandlers(config.AgentService, config.ExecutionService, config.A2AService.GetLogger(), config.A2AConfig)

	// Register A2A protocol routes
	a2aHandler.RegisterA2ARoutes(config.Router)

	// Register agent discovery routes
	agentDiscoveryHandler.RegisterAgentDiscoveryRoutes(config.Router)

	// Register JSON-RPC routes
	jsonrpcHandler.RegisterJSONRPCRoutes(config.Router)

	// Add a general A2A status endpoint
	config.Router.GET("/a2a/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "A2A service is running",
			"protocol_version": config.A2AConfig.Protocol.Version,
			"supported_transports": []string{
				"http_json",
				"grpc",
				"json_rpc",
			},
		})
	})
}