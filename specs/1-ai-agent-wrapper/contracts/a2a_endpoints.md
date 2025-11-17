# A2A API Contracts: AI Agent Wrapper

## Overview
API contract for Agent-to-Agent (A2A) endpoints that follow the standard A2A protocol specification (https://a2a-protocol.org/latest/specification/#323-httpjsonrest-transport). This ensures interoperability with other A2A-compliant systems.

## A2A Protocol Compliance

### HTTP/JSON/REST Transport
Following the A2A specification section 3.2.3:
- Base path: `/a2a`
- Content-Type: `application/json`
- Methods: `POST` for requests

### Standard A2A Message Structure
Based on A2A protocol specification:

#### A2A Header Block
The `a2a` block contains protocol-specific metadata:
```json
{
  "a2a": {
    "protocol": "a2a",
    "version": "1.0",
    "id": "unique-message-identifier",
    "timestamp": "2025-11-18T10:00:00.000Z",
    "type": "request|response"
  }
}
```

**Fields:**
- `protocol`: MUST be "a2a"
- `version`: Specifies the version of the A2A protocol (currently "1.0")
- `id`: Unique identifier for the message
- `timestamp`: RFC 3339 formatted timestamp
- `type`: Either "request" for requests or "response" for responses

#### A2A Context Block
The `context` block contains addressing and context information:
```json
{
  "context": {
    "from": "sender-agent-identifier",
    "to": "recipient-agent-identifier",
    "conversationId": "conversation-identifier"
  }
}
```

**Fields:**
- `from`: The identifier of the sending agent
- `to`: The identifier of the receiving agent
- `conversationId`: Identifier for grouping related messages in a conversation

#### A2A Payload Block
The `payload` block contains the actual message content:
```json
{
  "payload": {
    "method": "method-name",
    "params": {
      "param1": "value1",
      "param2": "value2"
    }
  }
}
```

**For responses:**
```json
{
  "payload": {
    "result": {
      "resultField1": "value1",
      "resultField2": "value2"
    }
  }
}
```

**Fields:**
- `method`: The method or action to perform
- `params`: Parameters for the method (requests only)
- `result`: Result of the method execution (responses only)

## Standard A2A Protocol Endpoints

### A2A Message Exchange
**Endpoint**: `POST /a2a`
**Description**: Standard A2A protocol endpoint for agent-to-agent communication

#### Request
Standard A2A request format:
```json
{
  "a2a": {
    "protocol": "a2a",
    "version": "1.0",
    "id": "request-unique-id",
    "timestamp": "2025-11-18T10:00:00.000Z",
    "type": "request"
  },
  "context": {
    "from": "source-agent-id",
    "to": "target-agent-id",
    "conversationId": "conversation-id"
  },
  "payload": {
    "method": "execute-agent",
    "params": {
      "agentId": "string",
      "input": "string",
      "workingDirectory": "string",
      "environmentVariables": {
        "VARIABLE_NAME": "value"
      }
    }
  }
}
```

#### Response
Standard A2A response format:
```json
{
  "a2a": {
    "protocol": "a2a",
    "version": "1.0",
    "id": "response-unique-id",
    "inResponseTo": "request-unique-id",
    "timestamp": "2025-11-18T10:00:01.000Z",
    "type": "response"
  },
  "context": {
    "from": "target-agent-id",
    "to": "source-agent-id"
  },
  "payload": {
    "result": {
      "output": "Generated response from agent",
      "executionId": "execution-identifier",
      "executionTime": 1234
    }
  }
}
```

#### HTTP Response
- Status: 200 OK on successful processing
- Content-Type: application/json
- Body: Standard A2A response format

### A2A Status Request
**Endpoint**: `POST /a2a`
**Description**: A2A protocol request to get the status of an agent

#### Request
```json
{
  "a2a": {
    "protocol": "a2a",
    "version": "1.0",
    "id": "status-request-id",
    "timestamp": "2025-11-18T10:00:00.000Z",
    "type": "request"
  },
  "context": {
    "from": "requesting-agent-id",
    "to": "target-agent-id",
    "conversationId": "conversation-id"
  },
  "payload": {
    "method": "status",
    "params": {}
  }
}
```

#### Response
```json
{
  "a2a": {
    "protocol": "a2a",
    "version": "1.0",
    "id": "status-response-id",
    "inResponseTo": "status-request-id",
    "timestamp": "2025-11-18T10:00:01.000Z",
    "type": "response"
  },
  "context": {
    "from": "target-agent-id",
    "to": "requesting-agent-id"
  },
  "payload": {
    "result": {
      "status": "ready|busy|error",
      "capabilities": [
        "execute-agent",
        "status"
      ],
      "agentInfo": {
        "id": "agent-id",
        "name": "Agent Name",
        "version": "1.0.0"
      }
    }
  }
}
```

## Implementation Considerations

### Go Implementation Structure
The A2A endpoints implementation will follow Go standard practices:
- **internal/api/handlers/a2a_handlers.go**: HTTP handlers that conform to A2A protocol
- **internal/services/a2a_service.go**: Service interface and implementation
- **pkg/a2a/protocol.go**: Protocol interfaces and common types
- **internal/models/a2a_endpoint.go**: A2A endpoint configuration model

### Internal API Endpoints (Non-A2A)
The following endpoints are internal to the supervisor system and not part of the A2A protocol:

#### Get Agent Status (Internal API)
**Endpoint**: `GET /api/agents/status`
**Description**: Get status of all configured agents (internal API)

#### Response
```json
{
  "agents": [
    {
      "id": "claude-agent",
      "name": "Claude Code Agent",
      "type": "claude-code",
      "status": "active|inactive",
      "lastExecution": "2025-11-18T09:30:00Z",
      "concurrentExecutions": 1,
      "maxConcurrent": 1
    }
  ]
}
```

#### Get Execution History (Internal API)
**Endpoint**: `GET /api/executions`
**Description**: Get execution history with filtering options (internal API)

#### Query Parameters
- `agentId`: Filter by specific agent
- `status`: Filter by execution status
- `limit`: Number of results to return (default 50)
- `offset`: Offset for pagination

#### Response
```json
{
  "executions": [
    {
      "id": "exec-12345",
      "agentId": "claude-agent",
      "status": "success",
      "startTime": "2025-11-18T09:30:00Z",
      "endTime": "2025-11-18T09:32:00Z",
      "executionTime": 120000
    }
  ],
  "total": 100,
  "limit": 50,
  "offset": 0
}
```