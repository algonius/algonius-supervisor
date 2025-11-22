# Quick Start: supervisorctl Control Program

This guide helps you get started with the supervisorctl CLI tool for managing algonius supervisor agents.

## Prerequisites

- Go 1.23 or later
- Running supervisord daemon
- Proper authentication credentials (if configured)

## Installation

### Build from Source
```bash
# Clone the repository
git clone <repository-url>
cd algonius-supervisor

# Build the supervisorctl binary
go build -o supervisorctl ./cmd/supervisorctl

# Move to system path (optional)
sudo mv supervisorctl /usr/local/bin/
```

### Verify Installation
```bash
supervisorctl --version
supervisorctl --help
```

## Basic Configuration

### Create Configuration File
Create `~/.config/supervisorctl/config.yaml`:
```yaml
server:
  url: "http://localhost:8080"  # Your supervisord URL
  timeout: 30s

auth:
  token: "your-bearer-token-here"  # If authentication is required

display:
  format: "table"     # table, json, yaml
  colors: true
  refresh_rate: "5s"

defaults:
  restart_attempts: 3
  wait_time: 5s
```

### Environment Variables
Alternatively, use environment variables:
```bash
export SUPERVISOR_SERVER_URL="http://localhost:8080"
export SUPERVISOR_AUTH_TOKEN="your-bearer-token-here"
export SUPERVISOR_SERVER_TIMEOUT="30s"
```

## Basic Usage

### 1. Check Supervisord Status
```bash
# Check if supervisord is running
supervisorctl health

# Show all agents and their status
supervisorctl status

# Show specific agents
supervisorctl status web-server worker
```

### 2. Agent Management
```bash
# Start an agent
supervisorctl start web-server

# Start multiple agents
supervisorctl start web-server worker database

# Stop an agent
supervisorctl stop web-server

# Restart an agent
supervisorctl restart web-server
```

### 3. Batch Operations
```bash
# Start all agents
supervisorctl start all

# Stop agents matching pattern
supervisorctl stop web-*

# Restart agents with specific prefix
supervisorctl restart *-service
```

### 4. Real-time Monitoring
```bash
# Follow agent logs
supervisorctl tail -f web-server

# Show recent logs (default 100 lines)
supervisorctl logs web-server

# Show more log lines
supervisorctl logs --lines 500 web-server

# Monitor all agent events
supervisorctl events
```

## Command Reference

### Status Commands
```bash
supervisorctl status [agent...]          # Show agent status
supervisorctl health                    # Check supervisord health
supervisorctl info                      # Show system information
```

### Agent Control
```bash
supervisorctl start <agent>...          # Start agents
supervisorctl stop <agent>...           # Stop agents
supervisorctl restart <agent>...        # Restart agents
```

### Monitoring
```bash
supervisorctl tail -f <agent>           # Follow agent logs
supervisorctl logs [agent]              # Show agent logs
supervisorctl events                    # Show lifecycle events
```

### Configuration
```bash
supervisorctl config show               # Show current configuration
supervisorctl config validate           # Validate configuration
```

### Global Options
```bash
--config <file>                         # Configuration file path
--server-url <url>                      # Override server URL
--token <token>                         # Authentication token
--format <format>                       # Output format: table|json|yaml
--no-colors                             # Disable colored output
--verbose, -v                           # Verbose output
--help, -h                              # Show help
--version                               # Show version
```

## Configuration Options

### Server Configuration
- **url**: Supervisord HTTP API endpoint
- **timeout**: Request timeout duration
- **auth.token**: Bearer token for authentication

### Display Configuration
- **format**: Output format (table, json, yaml)
- **colors**: Enable/disable colored output
- **refresh_rate**: Refresh rate for real-time commands

### Default Behavior
- **restart_attempts**: Number of automatic restart attempts
- **wait_time**: Time to wait between operations

## Pattern Matching

The supervisorctl supports pattern matching for batch operations:

### Exact Names
```bash
supervisorctl start web-server          # Exact agent name
```

### All Agents
```bash
supervisorctl restart all               # All configured agents
```

### Prefix Matching
```bash
supervisorctl start web-*               # web-server, web-worker, etc.
```

### Suffix Matching
```bash
supervisorctl stop *-service            # web-service, db-service, etc.
```

## Exit Codes

- **0**: Success
- **1**: Connection or general error
- **2**: Authentication error
- **3**: Agent not found
- **4**: Invalid command syntax

## Troubleshooting

### Connection Issues
```bash
# Check if supervisord is running
supervisorctl health

# Verify server URL
supervisorctl --server-url http://localhost:8080 status

# Check network connectivity
curl -H "Authorization: Bearer <token>" http://localhost:8080/health
```

### Authentication Issues
```bash
# Check token
supervisorctl --token <token> status

# Verify token with curl
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/agents
```

### Agent Not Found
```bash
# List all available agents
supervisorctl status

# Check spelling and pattern matching
supervisorctl status | grep <agent-name>
```

### Configuration Issues
```bash
# Validate configuration
supervisorctl config validate

# Show effective configuration
supervisorctl config show

# Check configuration file syntax
supervisorctl --config ~/.config/supervisorctl/config.yaml status
```

## Integration Examples

### Shell Scripts
```bash
#!/bin/bash
# Monitor agent status and alert on failures

if ! supervisorctl status web-server | grep -q "RUNNING"; then
    echo "Web server is not running!" >&2
    supervisorctl restart web-server
fi
```

### Cron Jobs
```bash
# Restart all agents daily at 2 AM
0 2 * * * /usr/local/bin/supervisorctl restart all

# Check health every 5 minutes
*/5 * * * * /usr/local/bin/supervisorctl health || mail -s "Supervisor alert" admin@example.com
```

### Systemd Integration
```ini
# /etc/systemd/system/supervisorctl-monitor.service
[Unit]
Description=Supervisor Monitor
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/supervisorctl status
User=supervisor
```

## Performance Tips

1. **Batch Operations**: Use batch commands instead of individual agent commands
2. **Output Format**: Use JSON format for script integration
3. **Filtering**: Use specific agent names instead of 'all' when possible
4. **Timeout**: Adjust timeout for large agent sets
5. **Polling**: Avoid frequent status polling in scripts

## Security Considerations

- Store authentication tokens securely (environment variables, not in config files)
- Use HTTPS in production environments
- Validate all user inputs
- Follow principle of least privilege for API access
- Enable audit logging for sensitive operations

## Getting Help

```bash
# Show general help
supervisorctl --help

# Show command-specific help
supervisorctl status --help
supervisorctl start --help

# Check version
supervisorctl --version
```

For additional support, check the documentation or file an issue on the project repository.