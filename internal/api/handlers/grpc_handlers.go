package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/algonius/algonius-supervisor/internal/models"
	"github.com/algonius/algonius-supervisor/internal/services"
	"github.com/algonius/algonius-supervisor/internal/a2a"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"go.uber.org/zap"
)

// GRPCHandlers handles gRPC requests for A2A protocol
type GRPCHandlers struct {
	agentService     *services.AgentService
	executionService services.IExecutionService
	a2aService       *services.A2AService
	logger           *zap.Logger
	config           *a2a.A2AConfig
	
	// Embed the gRPC server interface to satisfy the interface requirements
	// This would typically be a generated interface from a .proto file
	// For this implementation, we'll create a basic interface
	UnimplementedA2AServiceServer
}

// UnimplementedA2AServiceServer is a placeholder for the gRPC service interface
type UnimplementedA2AServiceServer struct{}

// NewGRPCHandlers creates a new instance of GRPCHandlers
func NewGRPCHandlers(agentService *services.AgentService, executionService services.IExecutionService, a2aService *services.A2AService, logger *zap.Logger, config *a2a.A2AConfig) *GRPCHandlers {
	return &GRPCHandlers{
		agentService:     agentService,
		executionService: executionService,
		a2aService:       a2aService,
		logger:           logger,
		config:           config,
	}
}

// RegisterGRPCRoutes registers all gRPC routes and services
func (gh *GRPCHandlers) RegisterGRPCRoutes(server *grpc.Server) {
	// Register the A2A service with the gRPC server
	RegisterA2AServiceServer(server, gh)
	
	// Any other gRPC services would be registered here
}

// SendMessage handles sending messages to agents via gRPC
func (gh *GRPCHandlers) SendMessage(ctx context.Context, req *A2AMessageSendRequest) (*A2AMessageSendResponse, error) {
	// Log the incoming request
	gh.logger.Info("handling gRPC A2A send message request",
		zap.String("agent_id", req.AgentId),
		zap.String("request_id", req.RequestId))

	// Validate the request
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "Agent ID is required")
	}
	
	if req.Message == nil {
		return nil, status.Error(codes.InvalidArgument, "Message is required")
	}

	// Get the agent configuration
	agent, err := gh.agentService.GetAgent(req.AgentId)
	if err != nil {
		gh.logger.Error("agent not found", zap.String("agent_id", req.AgentId), zap.Error(err))
		return nil, status.Error(codes.NotFound, "Agent not found")
	}

	// Validate the agent is enabled
	if !agent.Enabled {
		return nil, status.Error(codes.Unavailable, "Agent is disabled")
	}

	// Create a simple agent wrapper for execution
	simpleAgent := &SimpleGRPCAgent{
		config: agent,
	}

	// Execute the agent with the provided input
	// Extract content from the message payload
	var input string
	if req.Message != nil && req.Message.Payload != nil {
		// In a real implementation, we would properly extract content from the payload
		// For now, we'll take the method or other parameters as input
		input = fmt.Sprintf("%s: %v", req.Message.Payload.Method, req.Message.Payload.Params)
	} else {
		// Default to empty string if no payload
		input = ""
	}

	execution, err := gh.executionService.ExecuteAgent(ctx, simpleAgent, input)
	if err != nil {
		gh.logger.Error("agent execution failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "Agent execution failed")
	}

	// Create response
	response := &A2AMessageSendResponse{
		Message: &A2AMessage{
			Id:          generateA2AID(), // This would be a function to generate A2A IDs
			Type:        "response",
			Timestamp:   getCurrentTimestamp(), // This would be a function to get current timestamp
			InResponseTo: req.Message.Id,
			Context: &A2AContext{
				From:          req.Message.Context.To,
				To:            req.Message.Context.From,
				ConversationId: req.Message.Context.ConversationId,
				MessageId:     generateA2AID(),
			},
			Payload: &A2APayload{
				Result: &A2AResult{
					Status: "success",
					Output: "Agent execution completed",
					ExecutionId: execution.ID,
				},
			},
		},
		Success: true,
	}

	return response, nil
}

// StreamMessage handles streaming messages to agents via gRPC
func (gh *GRPCHandlers) StreamMessage(req *A2AMessageStreamRequest, srv A2AService_StreamMessageServer) error {
	// Log the incoming request
	gh.logger.Info("handling gRPC A2A stream message request",
		zap.String("agent_id", req.AgentId))

	// For now, we'll simulate a streaming response
	// In a real implementation, this would connect to an actual streaming agent

	// Send initial response
	initialResponse := &A2AMessageStreamResponse{
		Message: &A2AMessage{
			Id:          generateA2AID(),
			Type:        "response",
			Timestamp:   getCurrentTimestamp(),
			InResponseTo: req.Message.Id,
			Context: &A2AContext{
				From:          req.Message.Context.To,
				To:            req.Message.Context.From,
				ConversationId: req.Message.Context.ConversationId,
				MessageId:     generateA2AID(),
			},
			Payload: &A2APayload{
				Result: &A2AResult{
					Status: "running",
					Output: "Processing stream request...",
				},
			},
		},
		Finished: false,
	}

	if err := srv.Send(initialResponse); err != nil {
		gh.logger.Error("failed to send stream response", zap.Error(err))
		return err
	}

	// Simulate additional messages in the stream
	// In a real implementation, this would be connected to the actual agent execution
	for i := 0; i < 3; i++ {
		response := &A2AMessageStreamResponse{
			Message: &A2AMessage{
				Id:          generateA2AID(),
				Type:        "response",
				Timestamp:   getCurrentTimestamp(),
				InResponseTo: req.Message.Id,
				Context: &A2AContext{
					From:          req.Message.Context.To,
					To:            req.Message.Context.From,
					ConversationId: req.Message.Context.ConversationId,
					MessageId:     generateA2AID(),
				},
				Payload: &A2APayload{
					Result: &A2AResult{
						Status: "running",
						Output: "Processing step " + string(rune(i+1)),
					},
				},
			},
			Finished: i == 2, // Mark as finished on the last message
		}

		if err := srv.Send(response); err != nil {
			gh.logger.Error("failed to send stream response", zap.Error(err))
			return err
		}
	}

	return nil
}

// GetTask handles getting task information via gRPC
func (gh *GRPCHandlers) GetTask(ctx context.Context, req *A2AGetTaskRequest) (*A2AGetTaskResponse, error) {
	// Log the incoming request
	gh.logger.Info("handling gRPC A2A get task request",
		zap.String("agent_id", req.AgentId),
		zap.String("task_id", req.TaskId))

	// Validate the request
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "Agent ID is required")
	}
	
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "Task ID is required")
	}

	// In a real implementation, this would retrieve the actual task
	// For now, return a placeholder response
	task := &A2ATask{
		Id:     req.TaskId,
		Status: "completed",
		Messages: []*A2AMessage{
			{
				Id:        "msg-1",
				Type:      "response",
				Timestamp: getCurrentTimestamp(),
				Context: &A2AContext{
					From: req.AgentId,
					To:   "requester",
				},
				Payload: &A2APayload{
					Result: &A2AResult{
						Status: "completed",
						Output: "Task execution completed",
					},
				},
			},
		},
		CreatedAt:  getCurrentTimestamp(),
		ModifiedAt: getCurrentTimestamp(),
	}

	response := &A2AGetTaskResponse{
		Task: task,
	}

	return response, nil
}

// ListTasks handles listing tasks via gRPC
func (gh *GRPCHandlers) ListTasks(ctx context.Context, req *A2AListTasksRequest) (*A2AListTasksResponse, error) {
	// Log the incoming request
	gh.logger.Info("handling gRPC A2A list tasks request",
		zap.String("agent_id", req.AgentId))

	// Validate the request
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "Agent ID is required")
	}

	// In a real implementation, this would retrieve the actual tasks
	// For now, return a placeholder response
	tasks := []*A2ATask{
		{
			Id:     "task-1",
			Status: "completed",
			Messages: []*A2AMessage{
				{
					Id:        "msg-1",
					Type:      "response",
					Timestamp: getCurrentTimestamp(),
					Context: &A2AContext{
						From: req.AgentId,
						To:   "requester",
					},
					Payload: &A2APayload{
						Result: &A2AResult{
							Status: "completed",
							Output: "Task execution completed",
						},
					},
				},
			},
			CreatedAt:  getCurrentTimestamp(),
			ModifiedAt: getCurrentTimestamp(),
		},
	}

	response := &A2AListTasksResponse{
		Tasks:  tasks,
		TotalSize: int32(len(tasks)),
	}

	return response, nil
}

// CancelTask handles cancelling a task via gRPC
func (gh *GRPCHandlers) CancelTask(ctx context.Context, req *A2ACancelTaskRequest) (*A2ACancelTaskResponse, error) {
	// Log the incoming request
	gh.logger.Info("handling gRPC A2A cancel task request",
		zap.String("agent_id", req.AgentId),
		zap.String("task_id", req.TaskId))

	// Validate the request
	if req.AgentId == "" {
		return nil, status.Error(codes.InvalidArgument, "Agent ID is required")
	}
	
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "Task ID is required")
	}

	// In a real implementation, this would cancel the actual task
	// For now, return a placeholder response indicating cancellation
	task := &A2ATask{
		Id:     req.TaskId,
		Status: "cancelled",
		Messages: []*A2AMessage{
			{
				Id:        "msg-1",
				Type:      "response",
				Timestamp: getCurrentTimestamp(),
				Context: &A2AContext{
					From: req.AgentId,
					To:   "requester",
				},
				Payload: &A2APayload{
					Result: &A2AResult{
						Status: "cancelled",
						Output: "Task was cancelled",
					},
				},
			},
		},
		CreatedAt:  getCurrentTimestamp(),
		ModifiedAt: getCurrentTimestamp(),
	}

	response := &A2ACancelTaskResponse{
		Task: task,
	}

	return response, nil
}

// Helper functions (stubs for now, would be properly implemented in a real system)
func generateA2AID() string {
	// In a real implementation, this would generate a proper A2A ID
	return "grpc-generated-id"
}

func getCurrentTimestamp() string {
	// In a real implementation, this would return the current time in RFC3339 format
	return "2025-11-20T10:30:00Z"
}

// Define basic message types that would normally be generated from a .proto file
type A2AMessageSendRequest struct {
	AgentId   string      `json:"agent_id"`
	Message   *A2AMessage `json:"message"`
	RequestId string      `json:"request_id"`
}

type A2AMessageSendResponse struct {
	Message *A2AMessage `json:"message"`
	Success bool        `json:"success"`
}

type A2AMessageStreamRequest struct {
	AgentId string      `json:"agent_id"`
	Message *A2AMessage `json:"message"`
}

type A2AMessageStreamResponse struct {
	Message  *A2AMessage `json:"message"`
	Finished bool        `json:"finished"`
}

type A2AGetTaskRequest struct {
	AgentId string `json:"agent_id"`
	TaskId  string `json:"task_id"`
}

type A2AGetTaskResponse struct {
	Task *A2ATask `json:"task"`
}

type A2AListTasksRequest struct {
	AgentId   string            `json:"agent_id"`
	Filters   map[string]string `json:"filters"`
	PageSize  int32             `json:"page_size"`
	PageToken string            `json:"page_token"`
}

type A2AListTasksResponse struct {
	Tasks     []*A2ATask `json:"tasks"`
	TotalSize int32      `json:"total_size"`
	PageSize  int32      `json:"page_size"`
	PageToken string     `json:"page_token"`
}

type A2ACancelTaskRequest struct {
	AgentId string `json:"agent_id"`
	TaskId  string `json:"task_id"`
}

type A2ACancelTaskResponse struct {
	Task *A2ATask `json:"task"`
}

type A2AMessage struct {
	Id           string      `json:"id"`
	Type         string      `json:"type"`
	Timestamp    string      `json:"timestamp"`
	InResponseTo string      `json:"in_response_to"`
	Context      *A2AContext `json:"context"`
	Payload      *A2APayload `json:"payload"`
}

type A2AContext struct {
	From          string `json:"from"`
	To            string `json:"to"`
	ConversationId string `json:"conversation_id"`
	MessageId     string `json:"message_id"`
}

type A2APayload struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	Result *A2AResult  `json:"result"`
}

type A2AResult struct {
	Status      string `json:"status"`
	Output      string `json:"output"`
	ExecutionId string `json:"execution_id"`
}

type A2ATask struct {
	Id         string        `json:"id"`
	Status     string        `json:"status"`
	Messages   []*A2AMessage `json:"messages"`
	CreatedAt  string        `json:"created_at"`
	ModifiedAt string        `json:"modified_at"`
}

// RegisterA2AServiceServer registers the A2A service with the gRPC server
// This is a placeholder function to satisfy the compilation requirement
func RegisterA2AServiceServer(s *grpc.Server, srv A2AServiceServer) {
	// In a real implementation, this function would be auto-generated from the proto file
	// For now, we provide a stub that should satisfy compilation
	s.RegisterService(&A2AService_ServiceDesc, srv)
}

// A2AService_ServiceDesc is the service description for A2A service
var A2AService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "a2a.A2AService",
	HandlerType: (*A2AServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMessage",
			Handler:    _A2AService_SendMessage_Handler,
		},
		{
			MethodName: "GetTask",
			Handler:    _A2AService_GetTask_Handler,
		},
		{
			MethodName: "ListTasks",
			Handler:    _A2AService_ListTasks_Handler,
		},
		{
			MethodName: "CancelTask",
			Handler:    _A2AService_CancelTask_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamMessage",
			Handler:       _A2AService_StreamMessage_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "a2a.proto",
}

// Stub handler functions
func _A2AService_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	// Actual implementation would be generated
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func _A2AService_GetTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	// Actual implementation would be generated
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func _A2AService_ListTasks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	// Actual implementation would be generated
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func _A2AService_CancelTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	// Actual implementation would be generated
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func _A2AService_StreamMessage_Handler(srv interface{}, stream grpc.ServerStream) error {
	// Actual implementation would be generated
	return status.Error(codes.Unimplemented, "not implemented")
}

// The gRPC service interface that would normally be generated from a .proto file
type A2AServiceServer interface {
	SendMessage(context.Context, *A2AMessageSendRequest) (*A2AMessageSendResponse, error)
	StreamMessage(*A2AMessageStreamRequest, A2AService_StreamMessageServer) error
	GetTask(context.Context, *A2AGetTaskRequest) (*A2AGetTaskResponse, error)
	ListTasks(context.Context, *A2AListTasksRequest) (*A2AListTasksResponse, error)
	CancelTask(context.Context, *A2ACancelTaskRequest) (*A2ACancelTaskResponse, error)
}

type A2AService_StreamMessageServer interface {
	Send(*A2AMessageStreamResponse) error
	grpc.ServerStream
}

// SimpleGRPCAgent is a wrapper to make our agent configuration compatible with the execution service
type SimpleGRPCAgent struct {
	config *models.AgentConfiguration
}

func (sgr *SimpleGRPCAgent) Execute(ctx context.Context, input string) (*models.ExecutionResult, error) {
	// In a real implementation, this would execute the actual agent process
	// For this implementation, we'll return a mock result
	result := &models.ExecutionResult{
		ID:        "grpc-" + time.Now().Format("20060102-150405"),
		AgentID:   sgr.config.ID,
		Status:    models.SuccessStatus,
		Input:     input,
		Output:    fmt.Sprintf("Executed command: %s with input: %s", sgr.config.ExecutablePath, input),
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}
	return result, nil
}

func (sgr *SimpleGRPCAgent) GetID() string {
	return sgr.config.ID
}

func (sgr *SimpleGRPCAgent) GetName() string {
	return sgr.config.Name
}

func (sgr *SimpleGRPCAgent) GetType() string {
	return sgr.config.AgentType
}

func (sgr *SimpleGRPCAgent) IsReadOnly() bool {
	return sgr.config.AccessType == models.ReadOnlyAccessType
}

func (sgr *SimpleGRPCAgent) GetConfig() *models.AgentConfiguration {
	return sgr.config
}

func (sgr *SimpleGRPCAgent) Validate() error {
	return sgr.config.Validate()
}