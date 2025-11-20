package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/algonius/algonius-supervisor/internal/a2a"
	"github.com/algonius/algonius-supervisor/internal/api/handlers"
	"github.com/algonius/algonius-supervisor/internal/api/routes"
	"github.com/algonius/algonius-supervisor/internal/config"
	"github.com/algonius/algonius-supervisor/internal/logging"
	"github.com/algonius/algonius-supervisor/internal/services"
)

func main() {
	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := logging.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Create zap logger instance
	zap.ReplaceGlobals(logger)

	// Set Gin mode based on config
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add custom logging middleware
	router.Use(logging.Middleware(logger))

	// Create service instances
	agentService := services.NewAgentService(logger)

	// Create metrics collector
	metricsCollector := services.NewMetricsCollector(logger)

	// Create execution service with metrics and caching
	executionService := services.NewExecutionService(agentService, logger)

	// Create A2A service with required dependencies
	a2aService := services.NewA2AService(agentService, executionService, logger)

	// Create scheduler service
	schedulerService := services.NewSchedulerService(agentService, executionService, logger)

	// Load A2A configuration
	a2aConfig := a2a.DefaultA2AConfig()
	// Override defaults with actual config if available
	if viper.IsSet("a2a") {
		if err := viper.UnmarshalKey("a2a", a2aConfig); err != nil {
			zap.S().Errorf("Failed to unmarshal A2A config: %v", err)
		}
	}

	// Setup A2A routes
	routeConfig := &routes.A2ARouteConfig{
		Router:           router,
		A2AService:       a2aService,
		AgentService:     agentService,
		ExecutionService: executionService,
		A2AConfig:        a2aConfig,
	}
	routes.SetupA2ARoutes(routeConfig)

	// Create and register scheduled task handlers
	taskHandlers := handlers.NewScheduledTaskHandlers(schedulerService, logger)
	taskHandlers.RegisterScheduledTaskRoutes(router)

	// Define basic routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"service": "algonius-supervisor",
			"timestamp": time.Now().UTC(),
		})
	})

	// Add metrics endpoint
	router.GET("/metrics", func(c *gin.Context) {
		metrics := metricsCollector.GetOverallMetrics()
		c.JSON(200, metrics)
	})

	// Start server
	zap.S().Infof("Starting algonius-supervisor on %s:%d", cfg.Host, cfg.Port)
	if err := router.Run(cfg.Host + ":" + fmt.Sprintf("%d", cfg.Port)); err != nil {
		zap.S().Fatalf("Failed to start server: %v", err)
	}
}