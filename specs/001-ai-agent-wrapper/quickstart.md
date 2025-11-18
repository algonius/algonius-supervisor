# Quick Start Guide: A2A Protocol Integration

**Date**: 2025-11-18
**Version**: 0.3.0

## Overview

This guide provides a quick start for integrating with the A2A (Agent-to-Agent) Protocol implementation in algonius-supervisor using the `github.com/a2aproject/a2a-go` library. The system supports multiple transport protocols (HTTP+JSON/REST, gRPC, JSON-RPC 2.0) and provides a unified interface for communicating with CLI AI agents.

## Prerequisites

- Go 1.23+ installed
- Bearer token for authentication
- Agent configured in algonius-supervisor
- curl or HTTP client for testing

## 1. Agent Discovery

First, discover available agents and their capabilities:

```bash
# Get agent card for discovery
curl -X GET \
  "https://api.algonius.local/agents/claude-code-agent/v1/.well-known/agent-card.json" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response**:
```json
{
  "agentId": "claude-code-agent",
  "name": "Claude Code Agent",
  "version": "1.0.0",
  "capabilities": {
    "supportedMethods": ["execute-agent", "status", "list-capabilities"],
    "streamingSupport": true,
    "concurrentExecution": false,
    "supportedContentTypes": ["text/plain", "application/json"]
  },
  "endpoints": [
    {
      "protocol": "http_json",
      "url": "https://api.algonius.local/agents/claude-code-agent/v1"
    }
  ],
  "authentication": {
    "required": true,
    "methods": ["bearer_token"]
  }
}
```

## 2. Basic Agent Execution

Send a message to execute an agent:

```bash
# Execute agent with basic command
curl -X POST \
  "https://api.algonius.local/agents/claude-code-agent/v1/message:send" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "protocol": "a2a",
    "version": "0.3.0",
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "request",
    "timestamp": "2025-11-18T10:30:00Z",
    "context": {
      "from": "your-agent",
      "to": "claude-code-agent"
    },
    "payload": {
      "method": "execute-agent",
      "params": {
        "command": "review code",
        "context": "security review"
      }
    }
  }'
```

**Response**:
```json
{
  "protocol": "a2a",
  "version": "0.3.0",
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "type": "response",
  "timestamp": "2025-11-18T10:30:05Z",
  "inResponseTo": "550e8400-e29b-41d4-a716-446655440000",
  "context": {
    "from": "claude-code-agent",
    "to": "your-agent"
  },
  "payload": {
    "result": {
      "status": "completed",
      "output": "Code review completed. No security issues found.",
      "executionTime": 5234,
      "exitCode": 0
    }
  }
}
```

## 3. Streaming Agent Execution

For long-running tasks, use streaming:

```bash
# Execute agent with streaming response
curl -X POST \
  "https://api.algonius.local/agents/claude-code-agent/v1/message:stream" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "protocol": "a2a",
    "version": "0.3.0",
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "request",
    "timestamp": "2025-11-18T10:30:00Z",
    "context": {
      "from": "your-agent",
      "to": "claude-code-agent"
    },
    "payload": {
      "method": "execute-agent-stream",
      "params": {
        "command": "analyze large codebase",
        "context": "performance analysis"
      }
    }
  }'
```

**Streaming Response** (application/x-ndjson):
```json
{"type": "chunk", "data": "Analyzing file 1 of 100...", "finished": false}
{"type": "chunk", "data": "Analyzing file 50 of 100...", "finished": false}
{"type": "chunk", "data": "Analysis complete. Found 3 performance issues.", "finished": true}
```

## 4. Task Management

### List Tasks

```bash
# List all tasks for an agent
curl -X GET \
  "https://api.algonius.local/agents/claude-code-agent/v1/tasks?status=completed&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response**:
```json
{
  "tasks": [
    {
      "id": "task-123",
      "status": "completed",
      "startTime": "2025-11-18T10:30:00Z",
      "endTime": "2025-11-18T10:30:05Z",
      "duration": 5234
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

### Get Task Details

```bash
# Get detailed information about a specific task
curl -X GET \
  "https://api.algonius.local/agents/claude-code-agent/v1/tasks/task-123" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response**:
```json
{
  "id": "task-123",
  "status": "completed",
  "input": "review code",
  "output": "Code review completed. No security issues found.",
  "startTime": "2025-11-18T10:30:00Z",
  "endTime": "2025-11-18T10:30:05Z",
  "duration": 5234,
  "exitCode": 0,
  "logs": [
    {
      "timestamp": "2025-11-18T10:30:00Z",
      "level": "info",
      "message": "Starting code review"
    },
    {
      "timestamp": "2025-11-18T10:30:05Z",
      "level": "info",
      "message": "Code review completed"
    }
  ]
}
```

## 5. Error Handling

### Common Error Responses

**Authentication Error**:
```json
{
  "code": -32003,
  "message": "Authentication required",
  "data": {
    "required": "Bearer token"
  }
}
```

**Agent Not Found**:
```json
{
  "code": -32001,
  "message": "Agent not found",
  "data": {
    "agentId": "unknown-agent"
  }
}
```

**Concurrent Execution Limit**:
```json
{
  "code": -32004,
  "message": "Concurrent execution limit exceeded",
  "data": {
    "currentExecutions": 1,
    "maxExecutions": 1
  }
}
```

**Invalid Parameters**:
```json
{
  "code": -32602,
  "message": "Invalid params",
  "data": {
    "field": "method",
    "reason": "Method not found"
  }
}
```

## 6. GraphQL API (Alternative)

For more complex queries, use the GraphQL endpoint:

```bash
# GraphQL query for agent execution
curl -X POST \
  "https://api.algonius.local/graphql" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "
      mutation ExecuteAgent($agentId: ID!, $input: String!) {
        executeAgent(agentId: $agentId, input: $input) {
          id
          status
          output
          startTime
          endTime
          duration
        }
      }
    ",
    "variables": {
      "agentId": "claude-code-agent",
      "input": "review code"
    }
  }'
```

## 7. gRPC API (Alternative)

For high-performance scenarios, use gRPC:

```protobuf
// Proto definition for gRPC service
service A2AService {
  rpc SendMessage(A2AMessage) returns (A2AMessage);
  rpc SendStreamingMessage(A2AMessage) returns (stream A2AMessage);
  rpc GetAgentCard(AgentCardRequest) returns (AgentCard);
}
```

```bash
# Using grpcurl for testing
grpcurl -plaintext \
  -H "authorization: Bearer YOUR_TOKEN" \
  -d '{
    "protocol": "a2a",
    "version": "0.3.0",
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "REQUEST",
    "context": {
      "from": "your-agent",
      "to": "claude-code-agent"
    },
    "payload": {
      "method": "execute-agent",
      "params": {
        "command": "review code"
      }
    }
  }' \
  api.algonius.local:50051 \
  a2a.A2AService/SendMessage
```

## 8. JSON-RPC 2.0 API (Alternative)

For JSON-RPC 2.0 compatibility:

```bash
# JSON-RPC 2.0 request
curl -X POST \
  "https://api.algonius.local/agents/claude-code-agent/v1/jsonrpc" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "execute-agent",
    "params": {
      "command": "review code",
      "context": "security review"
    },
    "id": 1
  }'
```

**Response**:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "status": "completed",
    "output": "Code review completed. No security issues found.",
    "executionTime": 5234,
    "exitCode": 0
  },
  "id": 1
}
```

## 9. Integration Examples

### Python Example

```python
import requests
import json
from datetime import datetime

def execute_agent(agent_id, command, token):
    url = f"https://api.algonius.local/agents/{agent_id}/v1/message:send"

    message = {
        "protocol": "a2a",
        "version": "0.3.0",
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "type": "request",
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "context": {
            "from": "python-client",
            "to": agent_id
        },
        "payload": {
            "method": "execute-agent",
            "params": {
                "command": command
            }
        }
    }

    response = requests.post(
        url,
        headers={
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        },
        json=message
    )

    return response.json()

# Usage
result = execute_agent("claude-code-agent", "review code", "YOUR_TOKEN")
print(result["payload"]["result"]["output"])
```

### JavaScript/Node.js Example

```javascript
const axios = require('axios');

async function executeAgent(agentId, command, token) {
    const url = `https://api.algonius.local/agents/${agentId}/v1/message:send`;

    const message = {
        protocol: "a2a",
        version: "0.3.0",
        id: "550e8400-e29b-41d4-a716-446655440000",
        type: "request",
        timestamp: new Date().toISOString(),
        context: {
            from: "nodejs-client",
            to: agentId
        },
        payload: {
            method: "execute-agent",
            params: {
                command: command
            }
        }
    };

    try {
        const response = await axios.post(url, message, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });

        return response.data;
    } catch (error) {
        console.error('Error:', error.response.data);
        throw error;
    }
}

// Usage
executeAgent('claude-code-agent', 'review code', 'YOUR_TOKEN')
    .then(result => {
        console.log(result.payload.result.output);
    })
    .catch(error => {
        console.error('Failed:', error);
    });
```

### Go Example

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
)

type A2AMessage struct {
    Protocol  string      `json:"protocol"`
    Version   string      `json:"version"`
    ID        string      `json:"id"`
    Type      string      `json:"type"`
    Timestamp string      `json:"timestamp"`
    Context   A2AContext  `json:"context"`
    Payload   A2APayload  `json:"payload"`
}

type A2AContext struct {
    From string `json:"from"`
    To   string `json:"to"`
}

type A2APayload struct {
    Method string                 `json:"method"`
    Params map[string]interface{} `json:"params"`
}

func executeAgent(agentID, command, token string) (*A2AMessage, error) {
    url := fmt.Sprintf("https://api.algonius.local/agents/%s/v1/message:send", agentID)

    message := A2AMessage{
        Protocol:  "a2a",
        Version:   "0.3.0",
        ID:        "550e8400-e29b-41d4-a716-446655440000",
        Type:      "request",
        Timestamp: time.Now().Format(time.RFC3339),
        Context: A2AContext{
            From: "go-client",
            To:   agentID,
        },
        Payload: A2APayload{
            Method: "execute-agent",
            Params: map[string]interface{}{
                "command": command,
            },
        },
    }

    jsonData, err := json.Marshal(message)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var result A2AMessage
    err = json.Unmarshal(body, &result)
    if err != nil {
        return nil, err
    }

    return &result, nil
}

func main() {
    result, err := executeAgent("claude-code-agent", "review code", "YOUR_TOKEN")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Printf("Output: %s\n", result.Payload.Params["output"])
}
```

## 10. Best Practices

### Security
- Always use HTTPS for production
- Store tokens securely (environment variables, not code)
- Implement proper token rotation
- Validate all inputs before sending

### Performance
- Use streaming for long-running tasks
- Implement proper timeouts
- Handle rate limiting gracefully
- Use connection pooling for frequent requests

### Error Handling
- Always check HTTP status codes
- Parse and handle A2A error responses
- Implement retry logic with exponential backoff
- Log errors for debugging

### Rate Limiting
- Respect rate limits (even though not enforced by default)
- Implement client-side rate limiting
- Use bulk operations when possible
- Monitor your request patterns

## 11. Troubleshooting

### Common Issues

**401 Unauthorized**:
- Check your bearer token
- Ensure token hasn't expired
- Verify correct Authorization header format

**404 Not Found**:
- Verify agent ID is correct
- Check if agent is enabled
- Ensure A2A endpoint is configured

**429 Too Many Requests**:
- Wait before retrying
- Implement exponential backoff
- Check rate limit configuration

**500 Internal Server Error**:
- Check server logs
- Verify request format
- Ensure all required fields are present

### Debug Mode

Enable debug logging to see detailed request/response information:

```bash
# With curl - verbose mode
curl -v -X POST ...

# With custom headers for debugging
curl -X POST ... \
  -H "X-Debug: true" \
  -H "X-Request-ID: debug-123"
```

## 12. Next Steps

- Review the [A2A Protocol Specification](https://a2a-protocol.org/latest/specification)
- Explore [GraphQL API](./contracts/a2a-graphql.graphql) for complex queries
- Set up [scheduled tasks](./data-model.md#scheduled-task) for automation
- Implement [custom agents](./research.md#generic-agent-adapter-pattern) for your specific needs
- Configure [authentication and security](./research.md#security--compliance) for production use

For more detailed information, refer to:
- [Research Documentation](./research.md) - Comprehensive integration analysis
- [Data Model](./data-model.md) - Entity definitions and relationships
- [OpenAPI Specification](./contracts/a2a-openapi.yaml) - Complete API reference
- [GraphQL Schema](./contracts/a2a-graphql.graphql) - Alternative API interface