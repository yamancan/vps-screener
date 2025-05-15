#!/bin/bash

set -e  # Exit on any error
set -x  # Print commands as they are executed

echo "Starting VPS Agent setup..."

# Wait for any existing apt process to finish
echo "Waiting for any existing package manager processes..."
while fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 ; do
    echo "Waiting for other package manager to finish..."
    sleep 1
done

# Install required packages
echo "Installing required packages..."
apt-get update
apt-get install -y git

# Install Go
echo "Installing Go..."
wget https://go.dev/dl/go1.16.15.linux-amd64.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.16.15.linux-amd64.tar.gz
rm go1.16.15.linux-amd64.tar.gz

# Verify Go installation
echo "Verifying Go installation..."
if [ ! -f /usr/local/go/bin/go ]; then
    echo "Error: Go binary not found at /usr/local/go/bin/go"
    exit 1
fi

# Set up Go environment
echo "Setting up Go environment..."
export GOROOT=/usr/local/go
export GOPATH=/root/go
export PATH=/usr/local/go/bin:$PATH:$GOPATH/bin

# Verify Go version
echo "Verifying Go version..."
/usr/local/go/bin/go version
if [ $? -ne 0 ]; then
    echo "Error: Failed to verify Go installation"
    exit 1
fi

# Clean up any existing installation
echo "Cleaning up old files..."
pkill -f vps-agent || true
rm -rf /root/vps-screener
rm -rf /opt/vps-screener
rm -rf /usr/local/vps-screener
rm -rf /home/*/vps-screener
rm -f vps-agent agent.log

# Create new directory
echo "Creating new directory..."
mkdir -p /root/vps-screener/agent
cd /root/vps-screener/agent

# Download the latest code
echo "Downloading latest code..."
git clone https://github.com/yamancan/vps-screener.git /tmp/vps-screener
cp -r /tmp/vps-screener/agent/README.md /tmp/vps-screener/agent/collector /tmp/vps-screener/agent/config /tmp/vps-screener/agent/config.example.yaml /tmp/vps-screener/agent/config.go /tmp/vps-screener/agent/config.yaml /tmp/vps-screener/agent/executor /tmp/vps-screener/agent/go.mod /tmp/vps-screener/agent/main.go /tmp/vps-screener/agent/mapper /tmp/vps-screener/agent/plugins /tmp/vps-screener/agent/sender .
rm -rf /tmp/vps-screener

# Create config directory
echo "Setting up package structure..."
mkdir -p config
if [ -f config.go ]; then
    mv config.go config/
fi

# Update go.mod
echo "Updating dependencies..."
cat > go.mod << 'EOL'
module vps-agent

go 1.16

require (
	github.com/elastic/go-sysinfo v1.7.1
	github.com/gorilla/websocket v1.4.2
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
)

require (
	github.com/elastic/go-windows v1.0.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/procfs v0.0.11 // indirect
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd // indirect
)
EOL

# Update imports in all Go files
echo "Updating imports in Go files..."
find . -name "*.go" -type f -exec sed -i 's|vps-screener/agent/|vps-agent/|g' {} +

# Build the agent
echo "Building agent..."
echo "Current directory: $(pwd)"
echo "Directory contents:"
ls -la

echo "Running go mod tidy..."
/usr/local/go/bin/go mod tidy

echo "Building with verbose output..."
if ! /usr/local/go/bin/go build -v -o vps-agent .; then
    echo "Error: Build failed"
    echo "Directory contents after failed build:"
    ls -la
    exit 1
fi

echo "Checking if binary was created..."
if [ ! -f vps-agent ]; then
    echo "Error: Binary was not created"
    echo "Directory contents after build:"
    ls -la
    exit 1
fi

echo "Binary details:"
ls -l vps-agent
file vps-agent

# Make binary executable
chmod +x vps-agent

# Start the agent
echo "Starting agent..."
if ! nohup ./vps-agent > agent.log 2>&1 & then
    echo "Error: Failed to start agent"
    exit 1
fi

# Wait a moment for the agent to start
sleep 2

# Check if agent is running
echo "Checking if agent is running..."
if ! ps aux | grep -v grep | grep vps-agent > /dev/null; then
    echo "Error: Agent failed to start"
    echo "Recent logs:"
    tail -n 20 agent.log
    exit 1
fi

# Show logs
echo "Recent logs:"
tail -n 20 agent.log

echo "Setup complete!" 