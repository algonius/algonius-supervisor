# Agent Patterns: Support for Arbitrary CLI AI Agents

## Overview

This document describes the pattern-based approach that allows the system to support arbitrary CLI AI agents without code changes. Rather than implementing specific adapters for each agent, the system recognizes common patterns in how CLI agents accept input and return output.

## Input Patterns

### Stdin Pattern
- The agent accepts input via stdin
- The supervisor sends input directly to the agent's stdin
- Example: `echo "Hello" | claude`

### File Pattern
- The agent accepts input by specifying a file path as a command-line argument
- The supervisor creates an input file and passes its path to the agent
- Example: `custom-agent --input input.txt`

### Args Pattern
- The agent accepts input as command-line arguments
- The supervisor passes input as arguments to the executable
- Example: `openai --prompt "Hello"`

### JSON-RPC Pattern
- The agent communicates using JSON-RPC over stdin/stdout
- The supervisor handles the JSON-RPC protocol for communication
- Example: Language server protocol implementations

## Output Patterns

### Stdout Pattern
- The agent returns output via stdout
- The supervisor reads the output from the agent's stdout
- Example: `claude "What is AI?"` → output appears in stdout

### File Pattern
- The agent writes output to a file, typically specified by a command-line argument
- The supervisor reads the output from the specified file
- Example: `custom-agent --input input.txt --output output.txt`

### JSON-RPC Pattern
- The agent returns output using JSON-RPC over stdout
- The supervisor parses the JSON-RPC responses
- Example: Language server responses

## Configuration Examples

### Stdin/Stdout Agent (e.g., Claude, Ollama)
```yaml
id: claude-agent
name: Claude Code Agent
agentType: stdin-stdout
executablePath: claude
inputPattern: stdin
outputPattern: stdout
envs:
  CLAUDE_API_KEY: ${CLAUDE_API_KEY}
cliArgs:
  model: claude-3-opus
```

### File-based Agent (e.g., Custom processor)
```yaml
id: custom-file-agent
name: Custom File-Based Agent
agentType: file-input-output
executablePath: custom-processor
inputPattern: file
outputPattern: file
inputFileTemplate: "input_{{timestamp}}.txt"
outputFileTemplate: "output_{{timestamp}}.txt"
envs:
  CUSTOM_API_KEY: ${CUSTOM_API_KEY}
cliArgs:
  format: json
```

## Extensibility

When a new CLI agent is introduced that doesn't match existing patterns, a new pattern implementation can be added to the system. Once the pattern is implemented, all agents that follow this pattern can be configured without further code changes.

This approach allows for:
- Integration of new agents through configuration only
- Support for multiple agents using the same pattern
- Consistent handling of agents with similar interfaces
- Reduced maintenance overhead compared to per-agent implementations