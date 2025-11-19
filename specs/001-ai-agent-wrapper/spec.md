# Feature Specification: AI Agent Wrapper

**Feature Branch**: `1-ai-agent-wrapper`
**Created**: 2025-11-18
**Status**: Draft
**Input**: User description: "我想创建一个程序包装市面上所有 cli AI Agent， 提供 A2A Endpoint ， CLI AI Agent 类型包括：claude code, codex, gemini cli 等. algonius-supervisor 同时可以配置多个Agent, 也支持配置定时任务来驱动 Agent 完成任务"

## Clarifications

### Session 2025-11-18

- Q: Should A2A endpoints require authentication? → A: Authentication required for all A2A endpoints
- Q: Should A2A endpoints have rate limiting? → A: No rate limiting (unlimited requests)
- Q: What level of logging is required for agent executions? → A: Comprehensive logging required for all agent executions
- Q: How should concurrent executions be handled? → A: Distinguish between read-write and read-only agents; read-write only allow 1 concurrent execution, read-only allow multiple concurrent executions
- Q: How should sensitive data in agent inputs/outputs be handled? → A: System MUST NOT store or log sensitive data from agent inputs/outputs

### Session 2025-11-18 (Additional)

- Q: What A2A protocol specification should be implemented? → A: Implement A2A Protocol v0.3.0 using github.com/a2aproject/a2a-go library with support for JSON-RPC 2.0, gRPC, and HTTP+JSON/REST transport protocols as defined in docs/researchs/A2A_Protocol_Endpoints_Spec.md

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure and Execute CLI AI Agents (Priority: P1)

As a user, I want to configure multiple CLI AI agents in the algonius-supervisor system using a pattern-based configuration so that I can execute any command-line AI agent through a unified interface. The system should allow me to specify the agent's input/output patterns and execution parameters without requiring code changes for new agents.

**Why this priority**: This is the core functionality that enables the primary value proposition of wrapping multiple CLI AI agents in a unified system.

**Independent Test**: The system can be tested by configuring a single agent type (e.g., Claude Code) with specific parameters and successfully executing it with input commands, verifying that the agent runs and produces expected output.

**Acceptance Scenarios**:

1. **Given** I have configured an agent like Claude Code in the system, **When** I provide an input command, **Then** the system executes the agent with that command and returns the output
2. **Given** I have configured multiple different agents, **When** I select a specific agent and provide input, **Then** only the selected agent runs with the input
3. **Given** I have configured an agent with specific parameters, **When** I execute the agent, **Then** the agent uses those parameters as configured

---

### User Story 2 - A2A Endpoint Access (Priority: P2)

As a user, I want to access configured CLI AI agents through A2A (Agent-to-Agent) endpoints so that other agents or systems can communicate with my AI agents programmatically.

**Why this priority**: This enables the A2A (Agent-to-Agent) capability mentioned in the requirements and allows for integration with other systems.

**Independent Test**: The system can be tested by configuring an agent and then making an authenticated API call to the A2A endpoint, verifying that the agent executes and returns the appropriate response.

**Acceptance Scenarios**:

1. **Given** I have configured an agent with an A2A endpoint, **When** I make an authenticated API call to that endpoint, **Then** the corresponding agent executes and returns a response
2. **Given** I have multiple agents configured with A2A endpoints, **When** I call a specific endpoint with valid authentication, **Then** only the corresponding agent executes
3. **Given** I call an A2A endpoint without valid authentication, **When** I make the API call, **Then** the request is rejected with an authentication error

---

### User Story 3 - Schedule Automated Agent Tasks (Priority: P3)

As a user, I want to configure scheduled tasks that automatically execute my CLI AI agents so that routine or periodic tasks can be accomplished without manual intervention.

**Why this priority**: This provides automation capabilities for routine tasks, as mentioned in the requirements about "configuring scheduled tasks to drive agents to complete tasks."

**Independent Test**: The system can be tested by setting up a scheduled task for an agent with a specific time interval, waiting for that time, and verifying that the agent executed automatically as scheduled.

**Acceptance Scenarios**:

1. **Given** I have configured a scheduled task for an agent, **When** the scheduled time arrives, **Then** the system automatically executes the specified agent
2. **Given** I have scheduled tasks configured, **When** I modify the schedule, **Then** the system updates the execution timing accordingly

---

### Edge Cases

**Agent Execution Failures**
- **When an agent fails to execute or returns an error**: System categorizes errors as transient (network issues, resource constraints) or permanent (invalid configuration, missing dependencies). Transient errors trigger automatic retry up to 3 times with exponential backoff (1s, 2s, 4s). Permanent errors fail immediately without retry. All errors are logged with sanitized messages and execution context.
- **Agent crashes or hangs**: System implements timeout mechanism with default 5-minute execution limit. Agents exceeding timeout are forcibly terminated and marked as "Timeout" status. Timeout value is configurable per agent (1 minute minimum, 60 minutes maximum).

**Concurrency Management**
- **Read-write agent execution conflict**: When a read-write agent is already executing, new execution requests are queued and executed sequentially. Queue depth is limited to 10 pending requests per agent. Queue timeout is 30 minutes - requests exceeding this are rejected with 503 Service Unavailable.
- **Multiple concurrent executions**: Read-only agents support up to 10 concurrent executions. Read-write agents support exactly 1 execution with queue-based sequential processing. System monitors resource usage and rejects new executions if system resources (CPU > 90% or Memory > 85%) are exhausted.
- **Same agent multiple requests**: System tracks execution state per agent instance. Duplicate requests with identical input within 60 seconds are deduplicated and return the same execution result.

**Execution Characteristics**
- **User interaction requirements**: CLI agents requiring interactive input are not supported. System fails such executions with clear error message indicating "Interactive input required - execution not supported". Configuration validation detects and warns about potentially interactive commands during agent setup.
- **Variable execution times**: System implements no minimum execution time. Quick executions (< 1 second) are logged with microsecond precision. Long-running executions are monitored with heartbeat checks every 30 seconds. Partial results are not streamed - only complete execution results are returned.

**Security and Authentication**
- **A2A endpoint authentication failures**: Failed authentication attempts return 401 Unauthorized without revealing valid authentication methods. After 5 failed attempts from same IP within 10 minutes, IP is temporarily blocked for 15 minutes. Authentication uses Bearer token with minimum length of 32 characters.
- **Unlimited request handling**: System does not implement rate limiting but monitors request patterns. Sustained high request rates (> 100 requests/minute per endpoint) trigger performance monitoring alerts. No automatic throttling - manual intervention required for abuse cases.
- **Sensitive data handling**: System implements pattern-based sanitization for: API keys (AKIA..., sk-...), credentials (password=..., token=...), and email addresses. Sanitization happens before logging and result storage. Original sensitive data is never persisted - only masked versions (e.g., "sk-...1234").

**Configuration and State**
- **Agent access type determination**: Configuration explicitly specifies access_type as "read-only" or "read-write". Read-write agents are defined as those modifying files, system state, or external resources. Read-only agents only retrieve information without modification. Default is "read-only" if not specified.
- **Resource exhaustion**: When disk space falls below 1GB or system load exceeds 4.0, new execution requests are rejected with 503 Service Unavailable. Ongoing executions continue but are monitored for graceful degradation.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support configuration of multiple CLI AI agents through a generic configuration schema that works with any command-line AI agent without code changes
- **FR-002**: System MUST provide A2A endpoints that implement A2A Protocol v0.3.0 specification with support for JSON-RPC 2.0, gRPC, and HTTP+JSON/REST transport protocols
- **FR-003**: Users MUST be able to configure scheduled tasks that automatically execute agents at specified intervals or times
- **FR-004**: System MUST capture and return the output from executed CLI AI agents to the caller
- **FR-005**: System MUST maintain separate configuration profiles for different agent instances
- **FR-006**: System MUST support passing parameters and context to agents during execution
- **FR-007**: Users MUST be able to view execution history and logs for each agent
- **FR-008**: System MUST handle agent execution failures gracefully and provide error information
- **FR-009**: System MUST support concurrent execution of multiple different agents
- **FR-010**: System MUST allow configuration of agent-specific parameters and settings
- **FR-011**: System MUST require authentication for all A2A endpoints to ensure secure access
- **FR-012**: System MUST provide comprehensive logging for all agent executions including inputs, outputs, errors, and execution timing
- **FR-013**: System MUST distinguish between read-write and read-only agents in configuration, allowing only 1 concurrent execution for read-write agents while permitting multiple concurrent executions for read-only agents
- **FR-014**: System MUST NOT store or log sensitive data from agent inputs/outputs (e.g., API keys, credentials, personal information)

### Key Entities

- **Agent Configuration**: Represents the settings for a specific CLI AI agent, including type, parameters, execution settings, and access type (read-only vs read-write)
- **Scheduled Task**: Represents an automated task that executes an agent at specified intervals or times, including timing configuration and execution context
- **A2A Endpoint**: Represents the API endpoint that allows external systems to trigger agent execution, including security and access controls
- **Execution Result**: Represents the output, status, and metadata from an agent execution, including logs, return values, and error information

## Success Criteria *(mandatory)*

### Measurable Outcomes

#### Configuration and Compatibility (SC-001)
- **Goal**: Users can configure and successfully execute at least 5 different CLI AI agent types
- **Test Scenario**: Configure agents for Claude Code, Codex CLI, Gemini CLI, GitHub Copilot CLI, and Continue Dev
- **Success Metrics**: Each agent type executes successfully with 3 different commands, producing expected output
- **Performance Target**: Agent configuration setup < 2 minutes per agent type

#### A2A Endpoint Performance (SC-002)
- **Goal**: A2A endpoints respond within 2 seconds for basic agent execution requests
- **Test Scenario**: Send authenticated HTTP+JSON request with simple echo command to agent endpoint
- **Measurement Points**: Clock starts when request received, ends when response headers sent
- **Performance Targets**:
  - p50 (median): < 500ms
  - p95: < 2 seconds
  - p99: < 5 seconds
- **Load Conditions**: Tested with 10 concurrent requests to same endpoint

#### Task Scheduling Accuracy (SC-003)
- **Goal**: Scheduled tasks execute within 30 seconds of scheduled time
- **Test Scenarios**:
  - Schedule task at 5-minute intervals, measure actual execution times over 24 hours
  - Schedule task at specific wall-clock time, measure deviation
- **Success Metric**: 95% of tasks execute within ±30 seconds of scheduled time
- **Measurement**: Absolute deviation from scheduled time in seconds

#### Concurrent Execution Capacity (SC-004)
- **Goal**: Handle concurrent execution of at least 10 agent instances without performance degradation
- **Test Scenario**: Launch 10 read-only agent instances simultaneously with different tasks
- **Performance Metrics**:
  - CPU usage < 80% during concurrent execution
  - Memory usage < 4GB total for 10 agents
  - No agent execution time increases by >20% compared to single execution
- **Degradation Threshold**: Performance degradation defined as >20% increase in execution time vs baseline

#### Task Reliability (SC-005)
- **Goal**: 95% of scheduled tasks execute successfully without manual intervention
- **Measurement Period**: 30 days continuous operation
- **Success Metrics**:
  - Successful executions: Return code 0 and valid output
  - Failure types tracked: Timeout, agent error, system error, authentication failure
- **Target Breakdown**:
  - Total scheduled tasks: 100% (baseline)
  - Successful execution: ≥ 95%
  - Transient failures (auto-recovered): < 3%
  - Permanent failures (require intervention): < 2%

#### A2A Endpoint Setup Time (SC-006)
- **Goal**: Users can set up and verify A2A endpoints within 5 minutes
- **Test Scenario**: First-time user configures agent with A2A endpoint, starts server, makes test call
- **Measurement Points**:
  - Time to write configuration: < 2 minutes
  - Time to start service: < 1 minute
  - Time to verify endpoint: < 2 minutes (includes authentication setup and test execution)
- **Prerequisites**: User has valid configuration file template and agent executable path

#### Execution Result Availability (SC-007)
- **Goal**: Agent execution results available within 10 seconds of completion
- **Test Scenario**: Execute agent with moderate output (100KB), measure time from process exit to API response availability
- **Measurement Points**: Clock starts when agent process exits, ends when results available via A2A endpoint query
- **Performance Target**: Result availability < 10 seconds for 99% of executions
- **Includes**: Log aggregation, result serialization, and storage completion

#### Error Message Clarity (SC-008)
- **Goal**: System provides clear error messages for 99% of agent execution failures
- **Test Scenarios**: Trigger each error type (timeout, auth failure, agent crash, invalid config, resource exhaustion)
- **Success Criteria**:
  - Error message clearly indicates failure category (timeout, auth, agent error, system error)
  - Error includes actionable guidance or troubleshooting hint
  - No stack traces or technical details exposed to end users
  - Error format consistent across all failure types
- **Validation Method**: Human review of 20+ failure scenarios, score clarity 1-5 (target: ≥ 4 for 99% of cases)