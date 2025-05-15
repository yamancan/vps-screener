#!/bin/bash

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

# Install Go from official source
echo "Installing Go..."
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
rm go1.21.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.bashrc

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
cp -r /tmp/vps-screener/agent/* .
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

echo "Setup complete!" 