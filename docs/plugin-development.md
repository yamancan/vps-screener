# Plugin Development Guide

This guide explains how to create custom plugins for the VPS Screener agent.

## Plugin Types

VPS Screener supports two types of plugins:

1. **Python Scripts**: Simple, text-based plugins that output JSON
2. **Go Plugins**: Compiled plugins that integrate directly with the agent

## Python Plugin Development

### Basic Structure

A Python plugin is a simple script that outputs JSON to stdout. Here's a basic example:

```python
#!/usr/bin/env python3

import json
import os

# Get project name from environment
project_name = os.getenv("VPS_PROJECT_NAME", "unknown_project")

# Collect metrics
data = {
    "custom_metric": 42,
    "status": "ok",
    "project": project_name,
    "version": "1.0.0"
}

# Output JSON to stdout
print(json.dumps(data))
```

### Requirements

1. **Executable**: Make your script executable:
```bash
chmod +x my_plugin.py
```

2. **JSON Output**: Your plugin must output valid JSON to stdout
3. **Error Handling**: Handle errors gracefully and output error information in JSON
4. **Timeout**: Complete execution within 10 seconds

### Best Practices

1. **Error Handling**:
```python
try:
    # Your code here
    result = {"status": "ok", "data": data}
except Exception as e:
    result = {
        "status": "error",
        "error": str(e),
        "timestamp": time.time()
    }
print(json.dumps(result))
```

2. **Input Validation**:
```python
def validate_input(data):
    required_fields = ["metric_name", "value"]
    for field in required_fields:
        if field not in data:
            raise ValueError(f"Missing required field: {field}")
```

3. **Logging**:
```python
import logging

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)
```

## Go Plugin Development

### Basic Structure

A Go plugin implements the `Plugin` interface:

```go
type Plugin interface {
    Collect() (map[string]interface{}, error)
    Health() (string, error)
}
```

Example implementation:

```go
package main

import (
    "encoding/json"
    "os"
)

type MyPlugin struct {
    ProjectName string
}

func (p *MyPlugin) Collect() (map[string]interface{}, error) {
    return map[string]interface{}{
        "custom_metric": 42,
        "status": "ok",
        "project": p.ProjectName,
        "version": "1.0.0",
    }, nil
}

func (p *MyPlugin) Health() (string, error) {
    return "ok", nil
}
```

### Building Go Plugins

1. Create a new directory for your plugin:
```bash
mkdir -p agent/plugins/my_plugin
cd agent/plugins/my_plugin
```

2. Initialize Go module:
```bash
go mod init github.com/yamancan/vps-screener/plugins/my_plugin
```

3. Build the plugin:
```bash
go build -buildmode=plugin -o my_plugin.so
```

## Plugin Configuration

Add your plugin to the agent's `config.yaml`:

```yaml
projects:
  my_project:
    match:
      systemd_service: my-app.service
    plugin: plugins/my_plugin.py  # or my_plugin.so for Go plugins
```

## Testing Plugins

### Python Plugins

1. Test the script directly:
```bash
VPS_PROJECT_NAME=test_project ./my_plugin.py
```

2. Verify JSON output:
```bash
./my_plugin.py | python3 -m json.tool
```

### Go Plugins

1. Write unit tests:
```go
func TestMyPlugin_Collect(t *testing.T) {
    plugin := &MyPlugin{ProjectName: "test"}
    data, err := plugin.Collect()
    if err != nil {
        t.Errorf("Collect() error = %v", err)
    }
    if data["status"] != "ok" {
        t.Errorf("Expected status 'ok', got %v", data["status"])
    }
}
```

2. Run tests:
```bash
go test -v
```

## Security Considerations

1. **Input Validation**:
   - Validate all inputs
   - Sanitize data
   - Handle errors gracefully

2. **Resource Limits**:
   - Limit memory usage
   - Set timeouts
   - Restrict file system access

3. **Error Handling**:
   - Never expose sensitive information
   - Log errors appropriately
   - Return meaningful error messages

## Best Practices

1. **Code Organization**:
   - Keep plugins focused and small
   - Use clear naming conventions
   - Document your code

2. **Error Handling**:
   - Always handle errors
   - Provide meaningful error messages
   - Log errors appropriately

3. **Testing**:
   - Write unit tests
   - Test error conditions
   - Verify output format

4. **Documentation**:
   - Document plugin purpose
   - List requirements
   - Provide usage examples

## Example Plugins

### System Metrics Plugin

```python
#!/usr/bin/env python3

import json
import psutil
import os

def get_system_metrics():
    return {
        "cpu_percent": psutil.cpu_percent(),
        "memory_percent": psutil.virtual_memory().percent,
        "disk_usage": psutil.disk_usage('/').percent
    }

def main():
    try:
        metrics = get_system_metrics()
        print(json.dumps({
            "status": "ok",
            "metrics": metrics,
            "project": os.getenv("VPS_PROJECT_NAME", "unknown")
        }))
    except Exception as e:
        print(json.dumps({
            "status": "error",
            "error": str(e)
        }))

if __name__ == "__main__":
    main()
```

### Custom Application Plugin

```python
#!/usr/bin/env python3

import json
import requests
import os

def check_application_health(url):
    try:
        response = requests.get(url, timeout=5)
        return {
            "status": "ok" if response.status_code == 200 else "error",
            "response_time": response.elapsed.total_seconds(),
            "status_code": response.status_code
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e)
        }

def main():
    app_url = os.getenv("APP_URL", "http://localhost:8080/health")
    health_data = check_application_health(app_url)
    print(json.dumps({
        "status": health_data["status"],
        "metrics": health_data,
        "project": os.getenv("VPS_PROJECT_NAME", "unknown")
    }))

if __name__ == "__main__":
    main()
```

## Troubleshooting

### Common Issues

1. **Plugin Not Found**:
   - Check file permissions
   - Verify plugin path in config.yaml
   - Ensure plugin is executable

2. **Invalid JSON Output**:
   - Validate JSON format
   - Check for syntax errors
   - Verify all required fields

3. **Timeout Errors**:
   - Optimize plugin performance
   - Reduce data collection
   - Implement timeouts

### Debugging

1. **Enable Debug Logging**:
```python
import logging
logging.basicConfig(level=logging.DEBUG)
```

2. **Test Plugin Directly**:
```bash
VPS_PROJECT_NAME=test ./my_plugin.py
```

3. **Check Agent Logs**:
```bash
journalctl -u vps-agent -f
```

## Next Steps

- [Installation Guide](installation.md)
- [API Documentation](api.md)
- [Contributing Guide](../CONTRIBUTING.md) 