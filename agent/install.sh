#!/bin/bash

echo "Starting VPS Agent installation..."

# Stop any existing processes
echo "Stopping any existing processes..."
pkill -f vps-agent || true

# Clean up old files
echo "Cleaning up old files..."
rm -f vps-agent agent.log

# Update go.mod with correct versions
echo "Updating dependencies..."
cat > go.mod << 'EOL'
module vps-agent

go 1.21

require (
	github.com/elastic/go-sysinfo v1.11.1
	github.com/gorilla/websocket v1.5.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/elastic/go-windows v1.0.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
)
EOL

# Clean Go module cache
echo "Cleaning Go module cache..."
go clean -modcache

# Update imports in all Go files
echo "Updating imports in Go files..."
find . -name "*.go" -type f -exec sed -i 's|vps-screener/agent/|vps-agent/|g' {} +

# Build the agent
echo "Building agent..."
go mod tidy
go build -o vps-agent .

# Start the agent
echo "Starting agent..."
nohup ./vps-agent > agent.log 2>&1 &

# Wait a moment for the agent to start
sleep 2

# Check if agent is running
echo "Checking if agent is running..."
ps aux | grep vps-agent

# Show logs
echo "Recent logs:"
tail -n 20 agent.log

echo "Installation complete!" 