#!/bin/bash

echo "Starting fresh installation process..."

# Create a new directory
echo "Creating new directory..."
mkdir -p /root/vps-screener/agent
cd /root/vps-screener/agent

# Stop all vps-agent processes
echo "Stopping all vps-agent processes..."
pkill -f vps-agent || true

# Remove all vps-screener directories except current
echo "Removing old vps-screener directories..."
rm -rf /opt/vps-screener
rm -rf /usr/local/vps-screener
rm -rf /home/*/vps-screener

# Clean Go module cache
echo "Cleaning Go module cache..."
go clean -modcache

# Remove any remaining vps-agent binaries
echo "Removing any remaining vps-agent binaries..."
find / -name "vps-agent" -type f -delete 2>/dev/null

# Remove any remaining log files
echo "Removing log files..."
find / -name "agent.log" -type f -delete 2>/dev/null

# Download the latest code
echo "Downloading latest code..."
git clone https://github.com/yamancan/vps-screener.git /tmp/vps-screener
cp -r /tmp/vps-screener/agent/* .
rm -rf /tmp/vps-screener

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

echo "Fresh installation complete!" 