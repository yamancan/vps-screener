# Go Agent for VPS Screener

This document provides an overview of the Go-based agent for the VPS Screener project, including its structure, dependencies, and development guidelines.

## Overview

The agent is a Go application designed to run on each monitored VPS. Its primary responsibilities are:
- Collecting system and per-project resource metrics (CPU, RAM, Disk, Network).
- Mapping running processes to defined "projects" based on configurable rules.
- Sending collected metrics to a central API Gateway.
- Fetching and executing tasks (commands) received from the API Gateway and reporting results.
- Operating with a minimal resource footprint.

## Directory & Package Structure

The agent code resides in the `agent/` directory of the project.

```
agent/
├── go.mod                # Go module definition and dependencies
├── go.sum                # Checksums for dependencies
├── main.go               # Main application entry point, agent loop
├── config.yaml           # Agent configuration (metrics interval, API endpoint, project rules)
├── config.example.yaml   # Example configuration file
├── config/               # Package for loading and managing config.yaml
│   └── config.go
├── mapper/               # Package for mapping PIDs to projects
│   └── mapper.go
├── collector/            # Package for collecting system and per-project metrics
│   └── collector.go
├── sender/               # Package for sending data to the API Gateway
│   └── sender.go
├── executor/             # Package for fetching and executing tasks
│   └── executor.go
└── plugins/              # Directory for custom metric plugins
    ├── sample_plugin.py  # Example Python plugin
    └── .gitkeep
```

## Core Packages & Responsibilities

- **`main.go`**: Initializes the agent, loads configuration, sets up the main operational loop (ticker for metrics, task processing), and handles graceful shutdown via OS signals.
- **`config/config.go`**: Defines Go structs corresponding to `config.yaml` and provides the `LoadConfig` function to parse it.
- **`mapper/mapper.go`**: Contains logic to map Process IDs (PIDs) to project names. It uses rules defined in `config.yaml` (e.g., systemd service name, Docker labels, username, process name patterns) and employs techniques like cgroup parsing and `docker inspect` (via `os/exec`).
- **`collector/collector.go`**: Responsible for gathering metrics. It collects overall system metrics (CPU, RAM, Disk) and iterates through running processes, using the `mapper` to attribute resource usage (CPU, RAM) to specific projects. It also executes custom plugins for project-specific metrics.
- **`sender/sender.go`**: Handles the transmission of collected metrics to the API Gateway. It formats the data as JSON and sends it via HTTP POST, including the necessary authentication token.
- **`executor/executor.go`**: Manages the task lifecycle. It fetches pending tasks from the API Gateway, executes the specified commands (currently using `sh -c` with a note on future security enhancements like `setrlimit`), and sends the task results (stdout, stderr, status) back to the API.

## Plugin System

The agent supports custom metric collection through plugins. Plugins are executable scripts or binaries that output JSON to stdout. They are configured per project in `config.yaml`.

### Plugin Requirements

1. **Executable**: The plugin must be executable (`chmod +x plugin.py`).
2. **JSON Output**: The plugin must output a valid JSON object to stdout.
3. **Timeout**: Plugins have a 10-second execution timeout.
4. **Environment**: The project name is passed as `VPS_PROJECT_NAME` environment variable.

### Example Plugin (Python)

```python
#!/usr/bin/env python3
import json
import os
import random

# Access the project name from environment
project_name = os.getenv("VPS_PROJECT_NAME", "unknown_project")

# Collect custom metrics
data = {
    "custom_metric_A": random.randint(1, 1000),
    "custom_status": "ok",
    "processed_for_project": project_name,
    "plugin_version": "1.0.1"
}

# Output JSON to stdout
print(json.dumps(data))
```

### Plugin Configuration

In `config.yaml`, specify the plugin path for a project:

```yaml
# VPS Agent Configuration Example

# API Gateway connection details
api_gateway:
  url: http://localhost:3000/v1/metrics  # Test için localhost
  token: test-token-123  # Test için basit bir token

# Metrics collection interval in seconds
interval: 30

# Project configurations
projects:
  # Örnek: nginx servisi
  nginx:
    match:
      systemd_service: nginx.service
      user: www-data
    plugin: plugins/sample_plugin.py

  # Örnek: Docker container
  docker_app:
    match:
      docker_label: com.mycompany.project=docker-app
      docker_container_name: app-container
    plugin: plugins/sample_plugin.py
```

### Plugin Execution

- Plugins are executed in their own directory (working directory set to plugin's location).
- If a plugin fails (timeout, non-zero exit, invalid JSON), the error is logged and included in the metrics.
- Each plugin is executed only once per metrics collection cycle, even if multiple processes match the project.

## Key External Dependencies

As defined in `go.mod`:

- `gopkg.in/yaml.v3`: For parsing the `config.yaml` file.
- ` Acomprehensive cross-platform library for retrieving system and process information (CPU, memory, disk, network, process details, etc.).

Standard Go library packages are used for HTTP communication, JSON handling, OS interaction, etc.

## Building and Running

1.  **Navigate to the agent directory:**
    ```bash
    cd path/to/your/project/agent
    ```
2.  **Ensure Go is installed** (version 1.21 or as specified in `go.mod`).
3.  **Tidy dependencies (optional, good practice):
    ```bash
    go mod tidy
    ```
4.  **Build the agent:**
    ```bash
    go build -o vps-agent .
    ```
    This will create an executable named `vps-agent` (or `vps-agent.exe` on Windows) in the `agent` directory.

5.  **Configure `config.yaml`:**
    - Copy `config.example.yaml` to `config.yaml`.
    - Update `api_gateway.url` and `api_gateway.token`.
    - Define your `projects` mapping rules and plugins.

6.  **Run the agent:**
    ```bash
    ./vps-agent
    ```
    By default, it looks for `config.yaml` in the same directory. You can also specify a different config path using the `AGENT_CONFIG_PATH` environment variable:
    ```bash
    AGENT_CONFIG_PATH=/path/to/your/custom_config.yaml ./vps-agent
    ```

## Configuration

The agent's behavior is primarily controlled by `config.yaml`. Refer to the comments within the sample `config.yaml` and the structs in `config/config.go` for details on available options.

## Next Steps for Development

Key areas for further development and enhancement include:

- **Plugin System Enhancements:**
    - Add support for Go-native plugins (compiled with the agent).
    - Implement plugin versioning and auto-update mechanism.
    - Add plugin health checks and monitoring.
- **Executor Security & Resource Limits:** Enhancing `executor/executor.go` to apply resource limits (e.g., CPU, memory, time via `syscall.Setrlimit` or cgroups) to executed tasks and to minimize risks associated with `os/exec`.
- **Sender Buffering:** Implementing a local data buffering mechanism in `sender/sender.go` to store metrics locally if the API Gateway is unreachable and send them when connectivity is restored.
- **Mapper Enhancements:**
    - Implementing `container_name_pattern` matching in `mapper/mapper.go`.
    - Improving the robustness and error handling of cgroup parsing and `docker inspect` interactions.
- **Comprehensive Error Handling & Retries:** Adding more robust error handling throughout the agent, including retries for network operations where appropriate.
- **Configuration:** Making more parameters (e.g., HTTP timeouts, Docker command timeout) configurable via `config.yaml`.
- **Testing:** Adding unit and integration tests for the various packages.

This document should provide a good starting point for continuing development on the Go agent. 