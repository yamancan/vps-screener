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

# Clean up temporary directory
echo "Cleaning up temporary directory..."
rm -rf /tmp/vps-screener
if [ -d "/tmp/vps-screener" ]; then
    echo "Error: Failed to remove temporary directory"
    exit 1
fi

# Download the latest code
echo "Downloading latest code..."
if ! git clone https://github.com/yamancan/vps-screener.git /tmp/vps-screener; then
    echo "Error: Failed to clone repository"
    exit 1
fi

# Check repository structure
echo "Checking repository structure..."
if [ ! -d "/tmp/vps-screener/agent" ]; then
    echo "Error: agent directory not found in repository"
    echo "Repository contents:"
    ls -la /tmp/vps-screener
    exit 1
fi

# Copy files from repository
echo "Copying files from repository..."
cd /tmp/vps-screener/agent
for file in $(find . -type f -not -name "setup.sh"); do
    if [ -f "$file" ]; then
        # This copies files, flattening subdirectories from agent/ into /root/vps-screener/agent
        # e.g. ./plugins/sample.py becomes /root/vps-screener/agent/sample.py
        echo "Copying $file to /root/vps-screener/agent/$(basename "$file")"
        cp "$file" "/root/vps-screener/agent/$(basename "$file")"
    fi
done

# Clean up
rm -rf /tmp/vps-screener

# Return to agent directory
cd /root/vps-screener/agent

# Setting up package structure
echo "Setting up package structure..."
mkdir -p collector
mkdir -p mapper
mkdir -p executor
mkdir -p sender
mkdir -p config
mkdir -p plugins

# Move Go files to their respective package directories
echo "Moving Go files to their package directories..."
if [ -f config.go ]; then mv config.go config/; fi
if [ -f collector.go ]; then mv collector.go collector/; fi
if [ -f mapper.go ]; then mv mapper.go mapper/; fi
if [ -f executor.go ]; then mv executor.go executor/; fi
if [ -f sender.go ]; then mv sender.go sender/; fi

# Move plugin files
echo "Moving plugin files..."
if [ -f sample_plugin.py ]; then mv sample_plugin.py plugins/; fi
if [ -f .gitkeep ]; then mv .gitkeep plugins/; fi # Assuming .gitkeep belongs to plugins

# Update go.mod
echo "Updating go.mod..."
cat > go.mod << 'EOL'
module vps-agent

go 1.21

require (
	github.com/elastic/go-sysinfo v1.11.1
	github.com/gorilla/websocket v1.5.1
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
)

// Indirect dependencies will be populated by 'go mod tidy'
// It is often better to let 'go mod tidy' manage the specifics of these
// based on the direct dependencies listed above.
// If specific indirect versions are absolutely required due to complex constraints,
// they can be listed, but usually not necessary here.
EOL

# Update imports in all Go files
echo "Updating imports in Go files..."
# Update fully qualified imports from old GitHub path
find . -name "*.go" -type f -exec sed -i 's|github.com/yamancan/vps-screener/agent/config|vps-agent/config|g' {} +
find . -name "*.go" -type f -exec sed -i 's|github.com/yamancan/vps-screener/agent/collector|vps-agent/collector|g' {} +
find . -name "*.go" -type f -exec sed -i 's|github.com/yamancan/vps-screener/agent/executor|vps-agent/executor|g' {} +
find . -name "*.go" -type f -exec sed -i 's|github.com/yamancan/vps-screener/agent/mapper|vps-agent/mapper|g' {} +
find . -name "*.go" -type f -exec sed -i 's|github.com/yamancan/vps-screener/agent/sender|vps-agent/sender|g' {} +
# Update any broader old module paths if they exist (e.g. "vps-screener/agent" without github prefix)
find . -name "*.go" -type f -exec sed -i 's|vps-screener/agent/config|vps-agent/config|g' {} +
find . -name "*.go" -type f -exec sed -i 's|vps-screener/agent/collector|vps-agent/collector|g' {} +
find . -name "*.go" -type f -exec sed -i 's|vps-screener/agent/executor|vps-agent/executor|g' {} +
find . -name "*.go" -type f -exec sed -i 's|vps-screener/agent/mapper|vps-agent/mapper|g' {} +
find . -name "*.go" -type f -exec sed -i 's|vps-screener/agent/sender|vps-agent/sender|g' {} +
# General catch-all for the base path
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