# Agent Patterns: Support for Arbitrary CLI AI Agents

## Overview

This document describes the pattern-based approach that allows the system to support arbitrary CLI AI agents without code changes. Rather than implementing specific adapters for each agent, the system recognizes common patterns in how CLI agents accept input and return output, as well as how they handle execution modes.

## Execution Mode Patterns

The system supports two distinct execution modes for CLI AI agents:

### Task Mode Pattern
- The agent executes a single operation and then terminates
- Suitable for batch operations or single requests
- Process is created, runs to completion, then exits
- Each execution is independent with no shared state
- Example: `claude "Summarize this document"` → gets response → process exits

### Interactive Mode Pattern
- The agent maintains a persistent session with multiple exchanges
- Suitable for ongoing conversations or multi-step workflows
- Process remains alive for the duration of the session
- Maintains context between exchanges
- Example: `claude` → conversation begins → multiple exchanges → session timeout

## Input Patterns

### Stdin Pattern
- The agent accepts input via stdin
- The supervisor sends input directly to the agent's stdin
- For task mode: all input is sent at once, then stdin closed
- For interactive mode: input is sent in chunks as needed
- Example: `echo "Hello" | claude`

### File Pattern
- The agent accepts input by specifying a file path as a command-line argument
- The supervisor creates an input file and passes its path to the agent
- For task mode: file is created, agent processes, file may be deleted after
- For interactive mode: may use multiple files during session
- Example: `custom-agent --input input.txt`

### Args Pattern
- The agent accepts input as command-line arguments
- The supervisor passes input as arguments to the executable
- For task mode: all arguments provided at startup
- For interactive mode: may require special handling for continued interaction
- Example: `openai --prompt "Hello"`

### JSON-RPC Pattern
- The agent communicates using JSON-RPC over stdin/stdout
- The supervisor handles the JSON-RPC protocol for communication
- For task mode: single request-response cycle
- For interactive mode: multiple request-response cycles during session
- Example: Language server protocol implementations

## Output Patterns

### Stdout Pattern
- The agent returns output via stdout
- The supervisor reads the output from the agent's stdout
- For task mode: supervisor reads all output until process exits
- For interactive mode: supervisor processes output in chunks as it appears
- Example: `claude "What is AI?"` → output appears in stdout

### File Pattern
- The agent writes output to a file, typically specified by a command-line argument
- The supervisor reads the output from the specified file
- For task mode: supervisor waits for process to complete, then reads file
- For interactive mode: supervisor may monitor file for updates during session
- Example: `custom-agent --input input.txt --output output.txt`

### JSON-RPC Pattern
- The agent returns output using JSON-RPC over stdout
- The supervisor parses the JSON-RPC responses
- For task mode: single response message
- For interactive mode: multiple response messages during session
- Example: Language server responses

## Configuration Examples

### Task Mode Agent (e.g., Claude for one-off requests)
```yaml
id: claude-task-agent
name: Claude Code Task Agent
agentType: claude-task
executablePath: claude
mode: task
inputPattern: stdin
outputPattern: stdout
envs:
  CLAUDE_API_KEY: ${CLAUDE_API_KEY}
cliArgs:
  model: claude-3-opus
timeout: 300  # 5 minutes
enabled: true
```

### Interactive Mode Agent (e.g., Claude for ongoing conversations)
```yaml
id: claude-interactive-agent
name: Claude Code Interactive Agent
agentType: claude-interactive
executablePath: claude
mode: interactive
inputPattern: stdin
outputPattern: stdout
envs:
  CLAUDE_API_KEY: ${CLAUDE_API_KEY}
cliArgs:
  model: claude-3-opus
sessionTimeout: 1800  # 30 minutes
keepAlive: true
enabled: true
```

### File-based Agent (e.g., Custom processor)
```yaml
id: custom-file-agent
name: Custom File-Based Agent
agentType: file-input-output
executablePath: custom-processor
mode: task  # or interactive depending on agent capabilities
inputPattern: file
outputPattern: file
inputFileTemplate: "input_{{timestamp}}.txt"
outputFileTemplate: "output_{{timestamp}}.txt"
envs:
  CUSTOM_API_KEY: ${CUSTOM_API_KEY}
cliArgs:
  format: json
timeout: 600  # 10 minutes for task mode
sessionTimeout: 3600  # 1 hour for interactive mode
enabled: true
```

## Extensibility

When a new CLI agent is introduced that doesn't match existing patterns, a new pattern implementation can be added to the system. Once the pattern is implemented, all agents that follow this pattern can be configured without further code changes.

This approach allows for:
- Integration of new agents through configuration only
- Support for multiple agents using the same pattern
- Consistent handling of agents with similar interfaces
- Reduced maintenance overhead compared to per-agent implementations
- Clear separation between task and interactive operation modes