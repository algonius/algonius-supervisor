# Quickstart: AI Agent Wrapper

## Overview
Quickstart guide for setting up and using the AI Agent Wrapper to configure and execute CLI AI agents via A2A endpoints.

## Prerequisites
- Go 1.23 installed
- CLI AI agents (Claude Code, Codex, Gemini CLI, etc.) installed and accessible in PATH
- Valid API keys or authentication credentials for each agent

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/algonius/algonius-supervisor.git
   cd algonius-supervisor
   git checkout 1-ai-agent-wrapper
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

## Project Structure

The project follows Go standard practices:
```
algonius-supervisor/
├── cmd/
│   └── supervisor/
│       └── main.go          # Application entry point
├── internal/
│   ├── models/              # Data models (not importable by other projects)
│   ├── services/            # Service interfaces and implementations
│   ├── api/                 # API handlers
│   ├── agents/              # Agent implementations
│   ├── config/              # Configuration utilities
│   ├── auth/                # Authentication utilities
│   └── logging/             # Logging utilities
├── pkg/                     # Importable packages
│   ├── a2a/                 # A2A protocol interfaces and types
│   ├── mcp/                 # MCP context interfaces and types
│   └── types/               # Common shared types
├── tests/
├── contracts/
└── specs/
```

## Configuration

1. Create a configuration file (`config.yaml`):
   ```yaml
   agents:
     - id: claude-agent
       name: Claude Code Agent
       agentType: stdin-stdout  # Generic type based on input/output pattern
       executablePath: claude
       inputPattern: stdin      # Input via stdin
       outputPattern: stdout    # Output via stdout
       workingDirectory: /path/to/working/directory  # Optional: defaults to current directory
       envs:
         CLAUDE_API_KEY: ${CLAUDE_API_KEY}  # Can reference environment variables
         CUSTOM_VAR: "custom value"         # Or set static values
       cliArgs:
         model: claude-3-opus
       accessType: read-write
       timeout: 300
       enabled: true

     - id: gemini-agent
       name: Gemini CLI Agent
       agentType: stdin-stdout  # Generic type based on input/output pattern
       executablePath: gemini
       inputPattern: stdin      # Input via stdin
       outputPattern: stdout    # Output via stdout
       workingDirectory: /path/to/gemini/working  # Optional: defaults to current directory
       envs:
         GEMINI_API_KEY: ${GEMINI_API_KEY}  # Can reference environment variables
         GOOGLE_CLOUD_PROJECT: "my-project" # Additional project-specific variables
       cliArgs:
         model: gemini-pro
       accessType: read-only
       timeout: 300
       enabled: true

     - id: ollama-agent
       name: Ollama Agent
       agentType: stdin-stdout  # Generic type based on input/output pattern
       executablePath: ollama
       inputPattern: stdin      # Input via stdin
       outputPattern: stdout    # Output via stdout
       workingDirectory: /path/to/ollama/working  # Optional: defaults to current directory
       envs:
         OLLAMA_HOST: "http://localhost:11434"
       cliArgs:
         model: llama3
       accessType: read-only
       timeout: 300
       enabled: true

     - id: custom-file-agent
       name: Custom File-Based Agent
       agentType: file-input-output  # Generic type based on input/output pattern
       executablePath: custom-processor
       inputPattern: file           # Input via file
       outputPattern: file          # Output via file
       inputFileTemplate: "input_{{timestamp}}.txt"  # Template for input file
       outputFileTemplate: "output_{{timestamp}}.txt"  # Template for output file
       workingDirectory: /path/to/custom/working
       envs:
         CUSTOM_API_KEY: ${CUSTOM_API_KEY}
       cliArgs:
         format: json
       accessType: read-only
       timeout: 600
       enabled: true

   scheduledTasks:
     - id: daily-report
       name: Daily Report Generator
       agentId: claude-agent
       cronExpression: "0 9 * * *"  # Daily at 9 AM
       enabled: true
       inputParameters:
         task: Generate daily report
   ```

2. Set up environment variables:
   ```bash
   export CLAUDE_API_KEY="your-claude-api-key"
   export GEMINI_API_KEY="your-gemini-api-key"
   export A2A_AUTH_TOKEN="your-auth-token"
   ```

## Running the Service

1. Start the supervisor:
   ```bash
   cd cmd/supervisor
   go run main.go
   ```

2. The service will start on port 8080 by default.

## A2A Protocol Usage

### Standard A2A Request

The supervisor implements the standard A2A protocol as defined in the specification. To execute an agent:

```bash
curl -X POST http://localhost:8080/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "a2a": {
      "protocol": "a2a",
      "version": "1.0",
      "id": "req-12345",
      "timestamp": "2025-11-18T10:00:00.000Z",
      "type": "request"
    },
    "context": {
      "from": "client-agent",
      "to": "claude-agent",
      "conversationId": "conv-67890"
    },
    "payload": {
      "method": "execute-agent",
      "params": {
        "agentId": "claude-agent",
        "input": "Explain how neural networks work",
        "workingDirectory": "/path/to/working/dir",
        "envs": {
          "CUSTOM_VAR": "value"
        }
      }
    }
  }'
```

### Check A2A Status

```bash
curl -X POST http://localhost:8080/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "a2a": {
      "protocol": "a2a",
      "version": "1.0",
      "id": "status-req-67890",
      "timestamp": "2025-11-18T10:00:00.000Z",
      "type": "request"
    },
    "context": {
      "from": "client-agent",
      "to": "target-agent",
      "conversationId": "conv-67891"
    },
    "payload": {
      "method": "status",
      "params": {}
    }
  }'
```

## Internal API Usage

### Check Agent Status (Internal API)

```bash
curl -X GET http://localhost:8080/api/agents/status \
  -H "Authorization: Bearer $A2A_AUTH_TOKEN"
```

### Check Agent Status

```bash
curl -X GET http://localhost:8080/api/agents/status \
  -H "Authorization: Bearer $A2A_AUTH_TOKEN"
```

### List Scheduled Tasks

```bash
curl -X GET http://localhost:8080/api/tasks \
  -H "Authorization: Bearer $A2A_AUTH_TOKEN"
```

## Advanced Usage

### Configure Concurrency Settings

In your agent configuration, you can control concurrent executions:

```yaml
- id: high-priority-agent
  accessType: read-write  # Only allows 1 concurrent execution
  maxConcurrentExecutions: 1

- id: low-priority-agent
  accessType: read-only   # Allows multiple concurrent executions
  maxConcurrentExecutions: 5  # Optional: limit to 5 concurrent executions
```

### Schedule a Task

Add a scheduled task to your configuration to run agents automatically:

```yaml
scheduledTasks:
  - id: weekly-summary
    name: Weekly Summary
    agentId: claude-agent
    cronExpression: "0 0 * * 1"  # Every Monday at midnight
    enabled: true
    inputParameters:
      prompt: "Summarize the week's activities"
```

## Security Considerations

1. Always use HTTPS in production
2. Protect your API keys and authentication tokens
3. Use IP whitelisting for A2A endpoints if needed
4. Monitor logs for unauthorized access attempts
5. Ensure sensitive data is not stored or logged

## Troubleshooting

### Agent Not Found
- Check that the CLI agent is installed and accessible in PATH
- Verify the `executablePath` in the configuration

### Authentication Errors
- Ensure the `A2A_AUTH_TOKEN` is correctly set
- Check that authentication headers are properly formatted

### Execution Timeouts
- Increase the `timeout` value in the agent configuration
- Check if the agent is responding properly to CLI commands