# A2A API Contracts: Standard A2A Protocol Implementation

## Overview
API contracts for Agent-to-Agent (A2A) endpoints that fully implement the standard A2A protocol specification (https://a2a-protocol.org/latest/specification/). This ensures interoperability with other A2A-compliant systems and provides a complete implementation of the A2A protocol as specified.

## A2A Protocol Compliance

The implementation supports the three primary transport protocols defined in the A2A specification:
- HTTP+JSON/REST (primary implementation)
- JSON-RPC 2.0 (alternative)
- gRPC (alternative)

All endpoints comply with A2A protocol version 1.0 and use HTTPS for all communications.

## HTTP Transport Mappings

### Base Configuration
- Base path pattern: `/agents/{agentId}/v1`
- Content-Type: `application/json` (requests/responses)
- Authentication: Bearer token via `Authorization` header (handled at API layer)
- All endpoints require authentication

### Multi-Agent Path Prefix Routing
Following the path prefix routing approach for multi-agent support:

```
https://api.example.com/
├── /agents/agent1/v1/          # Agent 1's endpoints
│   ├── /message:send
│   ├── /tasks/{id}
│   └── /.well-known/agent-card.json
├── /agents/agent2/v1/          # Agent 2's endpoints
│   ├── /message:send
│   ├── /tasks/{id}
│   └── /.well-known/agent-card.json
└── /agents/                    # Agent registration and discovery
```

### Endpoint Mappings
Based on A2A Protocol specification section 2.1 with agent-specific prefixes:

| Function | REST Endpoint | Method | Description |
|----------|---------------|--------|-------------|
| Send message | `/agents/{agentId}/v1/message:send` | `POST` | Send message to specific agent, create or continue task |
| Stream message | `/agents/{agentId}/v1/message:stream` | `POST` | Send message to specific agent and subscribe to real-time updates |
| Get task | `/agents/{agentId}/v1/tasks/{id}` | `GET` | Get task status and details for specific agent |
| List tasks | `/agents/{agentId}/v1/tasks` | `GET` | List tasks with filtering and pagination for specific agent |
| Cancel task | `/agents/{agentId}/v1/tasks/{id}:cancel` | `POST` | Cancel an ongoing task for specific agent |
| Resubscribe | `/agents/{agentId}/v1/tasks/{id}:subscribe` | `POST` | Reconnect to task stream for specific agent |
| Set notification config | `/agents/{agentId}/v1/tasks/{id}/pushNotificationConfigs` | `POST` | Configure task push notifications for specific agent |
| Get notification config | `/agents/{agentId}/v1/tasks/{id}/pushNotificationConfigs/{configId}` | `GET` | Get push notification configuration for specific agent |
| List notification configs | `/agents/{agentId}/v1/tasks/{id}/pushNotificationConfigs` | `GET` | List task push notification configurations for specific agent |
| Delete notification config | `/agents/{agentId}/v1/tasks/{id}/pushNotificationConfigs/{configId}` | `DELETE` | Delete push notification configuration for specific agent |
| Get extended card | `/agents/{agentId}/v1/card` | `GET` | Get authenticated extended agent card for specific agent |

## Core Data Structures

### Message Definition
```json
{
  "content": "The actual message content",
  "role": "user|assistant|system",
  "id": "string (optional)",
  "timestamp": "RFC3339 formatted timestamp (optional)"
}
```

### PushNotificationConfig Definition
```json
{
  "id": "string",
  "url": "string (webhook url)",
  "token": "string (authentication token)",
  "authentication": {
    "schemes": ["bearer", "basic", "api-key"],
    "credentials": "string (credentials format depends on scheme)"
  }
}
```

### Task State Enum
- `created`: Task has been created but not yet started
- `running`: Task is currently executing
- `succeeded`: Task completed successfully
- `failed`: Task failed to complete
- `cancelled`: Task was cancelled
- `expired`: Task expired without completion

## Detailed Endpoint Specifications

### 1. Send Message Endpoint
**Endpoint**: `POST /agents/{agentId}/v1/message:send`
**Description**: Send a message to a specific agent, creating a new task or continuing an existing task.

#### Path Parameters
- `agentId`: The identifier of the target agent

#### Request
```json
{
  "message": {
    "content": "Your message here",
    "role": "user"
  },
  "configuration": {
    "acceptedOutputModes": ["text", "json"],
    "historyLength": 10,
    "pushNotificationConfig": {
      "id": "notification-config-1",
      "url": "https://your-webhook.example.com/task-update",
      "token": "your-auth-token",
      "authentication": {
        "schemes": ["bearer"],
        "credentials": "auth-credentials"
      }
    },
    "blocking": true
  },
  "metadata": {
    "customField": "customValue"
  }
}
```

#### Response - Success
```json
{
  "id": "agent1:task-12345",  // Task ID includes agent prefix for global uniqueness
  "status": "running|succeeded|failed|cancelled|expired",
  "messages": [
    {
      "content": "Response from agent",
      "role": "assistant",
      "id": "msg-67890",
      "timestamp": "2025-11-18T10:00:00.000Z"
    }
  ],
  "artifacts": [
    {
      "id": "artifact-1",
      "type": "code",
      "description": "Generated code file",
      "uri": "file://path/to/generated/code"
    }
  ],
  "history": [
    {
      "id": "hist-1",
      "timestamp": "2025-11-18T10:00:00.000Z",
      "message": {
        "content": "Previous message",
        "role": "user"
      }
    }
  ],
  "createdAt": "2025-11-18T10:00:00.000Z",
  "modifiedAt": "2025-11-18T10:00:30.000Z"
}
```

#### Response - Error
Standard A2A protocol error response:
```json
{
  "error": {
    "code": -32004,
    "message": "Unsupported operation error",
    "data": "Additional error details if applicable"
  }
}
```

#### HTTP Status Codes
- 200: Success
- 400: Invalid parameters
- 401: Unauthorized
- 404: Task not found
- 429: Rate limited
- 500: Internal server error

### 2. Stream Message Endpoint
**Endpoint**: `POST /agents/{agentId}/v1/message:stream`
**Description**: Send a message to a specific agent and subscribe to real-time task updates via Server-Sent Events (SSE).

#### Path Parameters
- `agentId`: The identifier of the target agent

#### Request
Same as the Send Message endpoint.

#### Response
Content-Type: `text/event-stream`
Events follow the format:
```
event: taskUpdate
data: {"id":"agent1:task-12345","status":"running","messages":[...],"artifacts":[...],"history":[...],"final":false}
id: event-1

event: taskUpdate
data: {"id":"agent1:task-12345","status":"succeeded","messages":[...],"artifacts":[...],"history":[...],"final":true}
id: event-2

event: done
data: {}
id: event-3
```

Each data payload follows the same structure as the successful Send Message response.

### 3. Get Task Endpoint
**Endpoint**: `GET /agents/{agentId}/v1/tasks/{id}`
**Description**: Get the current status and details of a specific task for a specific agent.

#### Path Parameters
- `agentId`: The identifier of the agent that owns the task
- `id`: Task ID (URL parameter)

#### Query Parameters
- `historyLength`: Number of history items to return (optional)

#### Response - Success
```json
{
  "id": "agent1:task-12345",
  "status": "running",
  "messages": [
    {
      "content": "Response from agent",
      "role": "assistant",
      "id": "msg-67890",
      "timestamp": "2025-11-18T10:00:00.000Z"
    }
  ],
  "artifacts": [
    {
      "id": "artifact-1",
      "type": "code",
      "description": "Generated code file",
      "uri": "file://path/to/generated/code"
    }
  ],
  "history": [
    {
      "id": "hist-1",
      "timestamp": "2025-11-18T10:00:00.000Z",
      "message": {
        "content": "Previous message",
        "role": "user"
      }
    }
  ],
  "createdAt": "2025-11-18T10:00:00.000Z",
  "modifiedAt": "2025-11-18T10:00:30.000Z"
}
```

### 4. List Tasks Endpoint
**Endpoint**: `GET /agents/{agentId}/v1/tasks`
**Description**: List tasks for a specific agent with filtering, pagination, and sorting capabilities.

#### Path Parameters
- `agentId`: The identifier of the agent whose tasks to list

#### Query Parameters
- `contextId`: Filter by context ID
- `status`: Filter by task status (created|running|succeeded|failed|cancelled|expired)
- `pageSize`: Number of results per page (1-100, default 50)
- `pageToken`: Pagination token
- `historyLength`: Number of history items to return (default 0)
- `lastUpdatedAfter`: Filter tasks updated after timestamp (milliseconds since epoch)
- `includeArtifacts`: Include artifacts in response (default false)
- `metadata`: Filter by metadata values

#### Response - Success
```json
{
  "tasks": [
    {
      "id": "agent1:task-12345",
      "status": "running",
      "messages": [...],
      "artifacts": [...],
      "history": [...],
      "createdAt": "2025-11-18T10:00:00.000Z",
      "modifiedAt": "2025-11-18T10:00:30.000Z"
    }
  ],
  "totalSize": 100,
  "pageSize": 50,
  "nextPageToken": "next-page-token-string"
}
```

### 5. Cancel Task Endpoint
**Endpoint**: `POST /agents/{agentId}/v1/tasks/{id}:cancel`
**Description**: Request cancellation of an ongoing task for a specific agent.

#### Path Parameters
- `agentId`: The identifier of the agent that owns the task
- `id`: Task ID (URL parameter)

#### Response - Success
```json
{
  "id": "agent1:task-12345",
  "status": "cancelled",  // Will be cancelled if operation is successful
  "messages": [...],
  "artifacts": [...],
  "history": [...],
  "createdAt": "2025-11-18T10:00:00.000Z",
  "modifiedAt": "2025-11-18T10:00:30.000Z"
}
```

### 6. Resubscribe Endpoint
**Endpoint**: `POST /agents/{agentId}/v1/tasks/{id}:subscribe`
**Description**: Reconnect to a task's Server-Sent Events (SSE) stream after disconnection for a specific agent.

#### Path Parameters
- `agentId`: The identifier of the agent that owns the task
- `id`: Task ID (URL parameter)

#### Response
Same SSE format as the Stream Message endpoint.

### 7. Push Notification Configuration Endpoints

#### Set Push Notification Configuration
**Endpoint**: `POST /agents/{agentId}/v1/tasks/{id}/pushNotificationConfigs`
**Description**: Set push notification configuration for a task for a specific agent.

##### Path Parameters
- `agentId`: The identifier of the agent that owns the task
- `id`: Task ID (URL parameter)

##### Request Body
```json
{
  "taskId": "agent1:task-12345",
  "pushNotificationConfig": {
    "id": "config-1",
    "url": "https://your-webhook.example.com/task-update",
    "token": "your-auth-token",
    "authentication": {
      "schemes": ["bearer"],
      "credentials": "auth-credentials"
    }
  }
}
```

##### Response - Success
```json
{
  "taskId": "agent1:task-12345",
  "pushNotificationConfig": {
    "id": "config-1",
    "url": "https://your-webhook.example.com/task-update",
    "token": "your-auth-token",
    "authentication": {
      "schemes": ["bearer"],
      "credentials": "auth-credentials"
    }
  }
}
```

#### Get Push Notification Configuration
**Endpoint**: `GET /agents/{agentId}/v1/tasks/{id}/pushNotificationConfigs/{configId}`
**Description**: Get specific push notification configuration for a task for a specific agent.

##### Path Parameters
- `agentId`: The identifier of the agent that owns the task
- `id`: Task ID (URL parameter)
- `configId`: Configuration ID (URL parameter)

##### Response - Success
```json
{
  "taskId": "agent1:task-12345",
  "pushNotificationConfig": {
    "id": "config-1",
    "url": "https://your-webhook.example.com/task-update",
    "token": "your-auth-token",
    "authentication": {
      "schemes": ["bearer"],
      "credentials": "auth-credentials"
    }
  }
}
```

#### List Push Notification Configurations
**Endpoint**: `GET /agents/{agentId}/v1/tasks/{id}/pushNotificationConfigs`
**Description**: List all push notification configurations for a task for a specific agent.

##### Path Parameters
- `agentId`: The identifier of the agent that owns the task
- `id`: Task ID (URL parameter)

##### Response - Success
```json
[
  {
    "taskId": "agent1:task-12345",
    "pushNotificationConfig": {
      "id": "config-1",
      "url": "https://your-webhook.example.com/task-update",
      "token": "your-auth-token",
      "authentication": {
        "schemes": ["bearer"],
        "credentials": "auth-credentials"
      }
    }
  }
]
```

#### Delete Push Notification Configuration
**Endpoint**: `DELETE /agents/{agentId}/v1/tasks/{id}/pushNotificationConfigs/{configId}`
**Description**: Delete a specific push notification configuration for a task for a specific agent.

##### Path Parameters
- `agentId`: The identifier of the agent that owns the task
- `id`: Task ID (URL parameter)
- `configId`: Configuration ID (URL parameter)

##### Response - Success
Status code: 200, no response body

### 8. Get Agent Card Endpoint
**Endpoint**: `GET /agents/{agentId}/v1/card`
**Description**: Get the specific agent's extended card with capabilities and metadata.

#### Path Parameters
- `agentId`: The identifier of the agent whose card to retrieve

#### Response - Success
```json
{
  "id": "claude-agent",
  "name": "Claude Code Agent",
  "description": "CLI Claude Code agent wrapped by algonius-supervisor",
  "version": "1.0.0",
  "protocols": {
    "a2a": {
      "supported": true,
      "version": "1.0"
    }
  },
  "capabilities": {
    "pushNotifications": true,
    "streaming": true
  },
  "endpoints": {
    "a2a": "https://your-agent-domain.com/agents/claude-agent/v1"
  },
  "supportedOutputModes": ["text", "json"],
  "supportsAuthenticatedExtendedCard": true,
  "metadata": {
    "agentType": "claude-code",
    "wrappedAgents": ["claude-code"]
  }
}
```

## Agent Discovery and Registration Endpoints

### 1. Agent Directory Endpoint
**Endpoint**: `GET /agents/list.json`
**Description**: Retrieve a list of all registered agents with their endpoints and card URLs.

#### Response - Success
```json
{
  "agents": [
    {
      "id": "claude-agent",
      "name": "Claude Code Agent",
      "description": "CLI Claude Code agent",
      "cardUrl": "/agents/claude-agent/v1/card",
      "endpoint": "/agents/claude-agent/v1",
      "status": "active"
    },
    {
      "id": "gemini-agent",
      "name": "Gemini CLI Agent",
      "description": "CLI Gemini agent",
      "cardUrl": "/agents/gemini-agent/v1/card",
      "endpoint": "/agents/gemini-agent/v1",
      "status": "active"
    }
  ]
}
```

### 2. Global Agent Cards Registry
**Endpoint**: `GET /.well-known/agent-cards/{agentId}.json`
**Description**: Retrieve a specific agent's card through the global registry.

#### Path Parameters
- `agentId`: The identifier of the agent whose card to retrieve

#### Response - Success
Same format as the individual agent card endpoint.

### 3. Well-known Agent Card Endpoint Per Agent
**Endpoint**: `GET /agents/{agentId}/.well-known/agent-card.json`
**Description**: Retrieve a specific agent's card via the standard .well-known path.

#### Path Parameters
- `agentId`: The identifier of the agent whose card to retrieve

#### Response - Success
Same format as the individual agent card endpoint.
```

## Authentication

All A2A endpoints require authentication using the HTTP Authorization header, which is handled at the API layer:

```
Authorization: Bearer <token>
```

Where `<token>` is a valid authentication token as configured at the API level.

## Error Handling

### Standard JSON-RPC Error Format
All errors follow the JSON-RPC 2.0 specification format:

```json
{
  "error": {
    "code": -32600,  // Error code
    "message": "Error message",  // Human-readable error message
    "data": "any"  // Optional error details
  }
}
```

### Standard JSON-RPC Errors
| Code | Name | Description |
|------|------|-------------|
| -32700 | Parse error | Invalid JSON format |
| -32600 | Invalid Request | Invalid JSON-RPC request |
| -32601 | Method not found | Method doesn't exist |
| -32602 | Invalid params | Invalid parameters |
| -32603 | Internal error | Server internal error |

### A2A-Specific Errors
| Code | Name | Description |
|------|------|-------------|
| -32001 | TaskNotFoundError | Task doesn't exist or has expired |
| -32002 | TaskNotCancelableError | Task cannot be cancelled (already terminated) |
| -32003 | PushNotificationNotSupportedError | Agent doesn't support push notifications |
| -32004 | UnsupportedOperationError | Operation not supported by agent |
| -32005 | ContentTypeNotSupportedError | Content type not supported |

## Go Implementation Structure
The A2A endpoints implementation uses the a2a-go library (github.com/a2aproject/a2a-go) to ensure A2A protocol compliance and follows Go standard practices with proper separation of concerns for multi-agent support:

### Dependencies
- **github.com/a2aproject/a2a-go** - Core A2A protocol implementation
- **google.golang.org/grpc** - gRPC server implementation (for gRPC transport)
- **github.com/gin-gonic/gin** - HTTP server and routing (for HTTP transport)

### Protocol Layer
- **pkg/a2a/protocol.go**: Protocol interfaces and common types, using a2a-go types
- **pkg/a2a/errors.go**: A2A-specific error types and codes, compliant with a2a-go
- **pkg/a2a/structures.go**: Core A2A data structures, wrapping a2a-go types where needed
- **pkg/a2a/messages.go**: Message handling utilities based on a2a-go message types

### A2A Core Implementation
- **internal/a2a/config.go**: Custom configuration for A2A request handler options
- **internal/a2a/middleware.go**: Custom middleware for the A2A request handler

### Service Layer
- **internal/services/a2a_service.go**: Service interface and implementation for A2A operations using a2a-go, coordinates between handlers and agent executors
- **internal/services/agent_executor.go**: Implements a2asrv.AgentExecutor interface for handling A2A requests, integrates with internal agent implementations
- **internal/services/agent_registry_service.go**: Service for managing registered agents and their discovery, maintains agent configurations and availability
- **internal/services/task_service.go**: Task management service with agent-specific task isolation and lifecycle management
- **internal/services/message_service.go**: Message handling service using and converting a2a-go types
- **internal/services/notification_service.go**: Push notification service for task updates

### AgentExecutor Implementation Details
The **internal/services/agent_executor.go** file implements the a2asrv.AgentExecutor interface from the a2a-go library. This interface requires implementing methods for handling the core A2A operations:

- `SendMessage(ctx context.Context, params *MessageSendParams) (*MessageSendResponse, error)`: Handles sending messages to specific agents and returning responses
- `StreamMessage(ctx context.Context, params *MessageSendParams, stream MessageStream) error`: Handles streaming messages to agents with real-time updates
- `GetTask(ctx context.Context, params *TaskGetParams) (*TaskGetResponse, error)`: Retrieves the status and details of a specific task
- `ListTasks(ctx context.Context, params *TaskListParams) (*TaskListResponse, error)`: Lists tasks with filtering and pagination
- `CancelTask(ctx context.Context, params *TaskCancelParams) (*TaskCancelResponse, error)`: Cancels an ongoing task
- `GetAgentCard(ctx context.Context) (*AgentCard, error)`: Returns the agent card information for discovery

The AgentExecutor implementation will coordinate with other services (task_service, message_service, etc.) to properly handle requests for different agents based on the agent ID extracted from the request context.

### A2A Request Handler Configuration
The implementation uses a2asrv.RequestHandler with custom options as shown in the a2a-go library guide:

- **internal/a2a/config.go**: Contains custom options implementation for the A2A request handler
- **internal/a2a/middleware.go**: Implements custom middleware for the A2A request handler

The server-side initialization follows this pattern:
```go
var options []a2asrv.RequestHandlerOption = newCustomOptions()  // Defined in internal/a2a/config.go
var agentExecutor a2asrv.AgentExecutor = newCustomAgentExecutor()  // Implemented in internal/services/agent_executor.go
requestHandler := a2asrv.NewHandler(agentExecutor, options...)

// Use the handler with specific transport protocols
grpcHandler := a2agrpc.NewHandler(requestHandler)
jsonrpcHandler := a2asrv.NewJSONRPCHandler(requestHandler)
```

### Model Layer
- **internal/models/a2a_endpoint.go**: A2A endpoint configuration model
- **internal/models/agent_config.go**: Agent configuration model with agent ID, type, and routing info
- **internal/models/task.go**: Task data structure with agent ID prefix for global uniqueness
- **internal/models/message.go**: Message data structure and operations, compatible with a2a-go types
- **internal/models/artifact.go**: Artifact data structure and operations
- **internal/models/push_notification_config.go**: Push notification configuration model

### API Layer
- **internal/api/handlers/a2a_handlers.go**: HTTP/REST handlers implementing A2A protocol for multi-agent routing using a2a-go
- **internal/api/handlers/grpc_handlers.go**: gRPC handlers implementing A2A protocol by wrapping a2asrv.Handler with a2a-go
- **internal/api/handlers/jsonrpc_handlers.go**: JSON-RPC handlers implementing A2A protocol by wrapping a2asrv.Handler with a2a-go
- **internal/api/handlers/agent_discovery_handler.go**: Handlers for agent discovery and registration endpoints
- **internal/api/middleware/auth.go**: Authentication middleware for A2A endpoints
- **internal/api/middleware/agent_router.go**: Middleware for agent-based routing using path prefixes
- **internal/api/routes/a2a_routes.go**: Route definitions for A2A endpoints with agent-specific paths

### Transport Protocol Implementation
The implementation uses a2a-go's transport-agnostic architecture:

- **gRPC Transport**: Uses `a2agrpc.NewHandler(requestHandler)` to wrap the core a2asrv.Handler
- **JSON-RPC Transport**: Uses `a2asrv.NewJSONRPCHandler(requestHandler)` to wrap the core a2asrv.Handler
- **HTTP/REST Transport**: Custom implementation using a2a-go types and utilities for protocol compliance

### Client-Side Implementation (for internal agent discovery and communication)
For internal communication between the supervisor and individual agents, as well as for agent-to-agent communication, we'll use a2a-go's client functionality:

- **internal/clients/a2a_client.go**: Implements a2aclient functionality for communication with wrapped agents
- **internal/clients/discovery.go**: Uses `agentcard.DefaultResolver.Resolve(ctx)` to discover agent capabilities
- **internal/clients/message_sender.go**: Uses `a2aclient.NewFromCard(ctx, card, options...)` to create clients from agent cards
- **internal/clients/task_monitor.go**: Uses client to monitor task status via a2a-go client interface

Client implementation follows a2a-go usage:
```go
// Resolving agent card
card, err := agentcard.DefaultResolver.Resolve(ctx)

// Creating client from card
client, err := a2aclient.NewFromCard(ctx, card, options...)

// Sending messages
msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.TextPart{Text: "Hello"})
resp, err := client.SendMessage(ctx, &a2a.MessageSendParams{Message: msg})
```

### Agent Integration Layer
- **internal/agents/agent_interface.go**: Interface for CLI AI agents
- **internal/agents/agent_service.go**: Agent execution service with multi-agent support
- **internal/agents/agent_factory.go**: Factory for creating and managing different agent types
- **internal/agents/types.go**: Agent-specific types and configurations