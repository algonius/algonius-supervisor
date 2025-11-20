package services

import (
	"net/http"

	"github.com/a2aproject/a2a-go/a2asrv"
	"go.uber.org/zap"
)

// A2AService interfaces with the a2a-go library
type A2AService struct {
	agentService    IAgentService
	executionService IExecutionService
	logger          *zap.Logger
	agentExecutor   *AgentExecutor
}

// NewA2AService creates a new instance of A2AService
func NewA2AService(agentService IAgentService, executionService IExecutionService, logger *zap.Logger) *A2AService {
	if logger == nil {
		// Create a fallback logger if none provided
		var err error
		logger, err = zap.NewProduction()
		if err != nil {
			// If we can't create production logger, use development logger
			logger, _ = zap.NewDevelopment()
		}
	}

	// Create an agent executor that will handle the actual agent execution
	agentExecutor := &AgentExecutor{
		agentService:    agentService,
		executionService: executionService,
		logger:          logger,
	}

	return &A2AService{
		agentService:    agentService,
		executionService: executionService,
		logger:          logger,
		agentExecutor:   agentExecutor,
	}
}

// GetAgentExecutor returns the agent executor for use with the a2a-go library
func (as *A2AService) GetAgentExecutor() *AgentExecutor {
	return as.agentExecutor
}

// GetLogger returns the logger used by the A2A service
func (as *A2AService) GetLogger() *zap.Logger {
	return as.logger
}

// StartHTTPServer starts the A2A HTTP server
func (as *A2AService) StartHTTPServer(address string) error {
	// Create the A2A server using the a2a-go library's JSON-RPC handler
	handler := a2asrv.NewHandler(as.agentExecutor)
	jsonRpcHandler := a2asrv.NewJSONRPCHandler(handler)

	// Start the HTTP server
	return http.ListenAndServe(address, jsonRpcHandler)
}