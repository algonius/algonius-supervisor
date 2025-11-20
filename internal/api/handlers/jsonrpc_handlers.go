package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/algonius/algonius-supervisor/internal/models"
	"github.com/algonius/algonius-supervisor/internal/services"
	"github.com/algonius/algonius-supervisor/internal/a2a"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// JSONRPCHandlers handles JSON-RPC 2.0 requests for A2A protocol
type JSONRPCHandlers struct {
	agentService     *services.AgentService
	executionService services.IExecutionService
	logger           *zap.Logger
	config           *a2a.A2AConfig
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      interface{} `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

// RPCError represents a JSON-RPC 2.0 error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// NewJSONRPCHandlers creates a new instance of JSONRPCHandlers
func NewJSONRPCHandlers(agentService *services.AgentService, executionService services.IExecutionService, logger *zap.Logger, config *a2a.A2AConfig) *JSONRPCHandlers {
	return &JSONRPCHandlers{
		agentService:     agentService,
		executionService: executionService,
		logger:           logger,
		config:           config,
	}
}

// RegisterJSONRPCRoutes registers JSON-RPC routes
func (jrh *JSONRPCHandlers) RegisterJSONRPCRoutes(router *gin.Engine) {
	// Apply authentication and validation middleware
	jsonrpcGroup := router.Group("/jsonrpc")
	jsonrpcGroup.Use(a2a.AuthenticationMiddleware(jrh.config))
	jsonrpcGroup.Use(a2a.CORSMiddleware(jrh.config))

	// Handle all JSON-RPC requests through a single endpoint
	jsonrpcGroup.POST("", jrh.HandleJSONRPC)
}

// HandleJSONRPC handles incoming JSON-RPC requests
func (jrh *JSONRPCHandlers) HandleJSONRPC(c *gin.Context) {
	// Log the incoming request
	jrh.logger.Info("handling JSON-RPC request",
		zap.String("path", c.Request.URL.Path))

	// Read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		jrh.logger.Error("failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, jrh.createJSONRPCError(nil, -32700, "Parse error", "Failed to parse request body"))
		return
	}

	// Parse the JSON-RPC request
	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		jrh.logger.Error("failed to parse JSON-RPC request", zap.Error(err))
		c.JSON(http.StatusBadRequest, jrh.createJSONRPCError(nil, -32700, "Parse error", "Invalid JSON was received"))
		return
	}

	// Validate the JSON-RPC request
	if req.Jsonrpc != "2.0" {
		response := jrh.createJSONRPCError(req.ID, -32600, "Invalid Request", "JSON-RPC version must be 2.0")
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Process the request based on the method
	response := jrh.processRequest(c, req)

	// Set content type and return response
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, response)
}

// processRequest processes the JSON-RPC request based on the method
func (jrh *JSONRPCHandlers) processRequest(c *gin.Context, req JSONRPCRequest) JSONRPCResponse {
	switch req.Method {
	case "execute-agent":
		return jrh.handleExecuteAgent(c, req)
	case "status":
		return jrh.handleStatus(c, req)
	case "list-agents":
		return jrh.handleListAgents(c, req)
	default:
		return jrh.createJSONRPCError(req.ID, -32601, "Method not found", fmt.Sprintf("Method %s not found", req.Method))
	}
}

// handleExecuteAgent handles execute-agent method
func (jrh *JSONRPCHandlers) handleExecuteAgent(c *gin.Context, req JSONRPCRequest) JSONRPCResponse {
	// Extract parameters
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return jrh.createJSONRPCError(req.ID, -32602, "Invalid params", "Parameters must be an object")
	}

	// Extract agent ID (from context or URL)
	agentID, exists := params["agentId"].(string)
	if !exists {
		// Try to get from a different field name
		agentID, exists = params["agent_id"].(string)
		if !exists {
			// Default to the first path parameter if available in context
			agentID = c.Param("agentId")
			if agentID == "" {
				return jrh.createJSONRPCError(req.ID, -32602, "Invalid params", "agentId is required")
			}
		}
	}

	// Get the agent configuration
	agent, err := jrh.agentService.GetAgent(agentID)
	if err != nil {
		jrh.logger.Error("agent not found", zap.String("agent_id", agentID), zap.Error(err))
		return jrh.createJSONRPCError(req.ID, -32001, "Agent not found", fmt.Sprintf("Agent with ID %s not found", agentID))
	}

	// Extract input from parameters
	input, exists := params["input"].(string)
	if !exists {
		return jrh.createJSONRPCError(req.ID, -32602, "Invalid params", "input is required")
	}

	// Create a simple agent wrapper for execution
	simpleAgent := &SimpleJSONRPCAgent{
		config: agent,
	}

	// Execute the agent
	execution, err := jrh.executionService.ExecuteAgent(c.Request.Context(), simpleAgent, input)
	if err != nil {
		jrh.logger.Error("agent execution failed", zap.Error(err))
		return jrh.createJSONRPCError(req.ID, -32002, "Agent execution failed", err.Error())
	}

	// Create result
	result := map[string]interface{}{
		"execution_id": execution.ID,
		"status":       string(execution.State),
		"output":       "Agent execution completed",
		"agent_id":     agentID,
	}

	return JSONRPCResponse{
		Jsonrpc: "2.0",
		Result:  result,
		ID:      req.ID,
	}
}

// handleStatus handles status method
func (jrh *JSONRPCHandlers) handleStatus(c *gin.Context, req JSONRPCRequest) JSONRPCResponse {
	result := map[string]interface{}{
		"status":       "healthy",
		"uptime":       "running",
		"version":      "1.0.0",
		"capabilities": []string{"execute-agent", "status", "list-agents"},
	}

	return JSONRPCResponse{
		Jsonrpc: "2.0",
		Result:  result,
		ID:      req.ID,
	}
}

// handleListAgents handles list-agents method
func (jrh *JSONRPCHandlers) handleListAgents(c *gin.Context, req JSONRPCRequest) JSONRPCResponse {
	agents, err := jrh.agentService.ListAgents()
	if err != nil {
		jrh.logger.Error("failed to list agents", zap.Error(err))
		return jrh.createJSONRPCError(req.ID, -32603, "Internal error", "Failed to list agents")
	}

	// Convert agents to response format
	agentList := make([]map[string]interface{}, len(agents))
	for i, agent := range agents {
		agentList[i] = map[string]interface{}{
			"id":          agent.ID,
			"name":        agent.Name,
			"agent_type":  agent.AgentType,
			"access_type": string(agent.AccessType),
			"enabled":     agent.Enabled,
		}
	}

	result := map[string]interface{}{
		"agents": agentList,
		"total":  len(agentList),
	}

	return JSONRPCResponse{
		Jsonrpc: "2.0",
		Result:  result,
		ID:      req.ID,
	}
}

// createJSONRPCError creates a JSON-RPC error response
func (jrh *JSONRPCHandlers) createJSONRPCError(id interface{}, code int, message, data string) JSONRPCResponse {
	return JSONRPCResponse{
		Jsonrpc: "2.0",
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}
}

// SimpleJSONRPCAgent is a wrapper to make our agent configuration compatible with the execution service
type SimpleJSONRPCAgent struct {
	config *models.AgentConfiguration
}

func (sjra *SimpleJSONRPCAgent) Execute(ctx context.Context, input string) (*models.ExecutionResult, error) {
	// In a real implementation, this would execute the actual agent process
	// For this implementation, we'll return a mock result
	result := &models.ExecutionResult{
		ID:        "jsonrpc-" + time.Now().Format("20060102-150405"),
		AgentID:   sjra.config.ID,
		Status:    models.SuccessStatus,
		Input:     input,
		Output:    fmt.Sprintf("Executed command: %s with input: %s", sjra.config.ExecutablePath, input),
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}
	return result, nil
}

func (sjra *SimpleJSONRPCAgent) GetID() string {
	return sjra.config.ID
}

func (sjra *SimpleJSONRPCAgent) GetName() string {
	return sjra.config.Name
}

func (sjra *SimpleJSONRPCAgent) GetType() string {
	return sjra.config.AgentType
}

func (sjra *SimpleJSONRPCAgent) IsReadOnly() bool {
	return sjra.config.AccessType == models.ReadOnlyAccessType
}

func (sjra *SimpleJSONRPCAgent) GetConfig() *models.AgentConfiguration {
	return sjra.config
}

func (sjra *SimpleJSONRPCAgent) Validate() error {
	return sjra.config.Validate()
}