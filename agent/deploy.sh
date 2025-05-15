#!/bin/bash

# Stop the current agent if it's running
if pgrep vps-agent > /dev/null; then
    echo "Stopping current agent..."
    pkill vps-agent
    sleep 2
fi

# Backup the current config
if [ -f config.yaml ]; then
    echo "Backing up current config..."
    cp config.yaml config.yaml.bak
fi

# Remove old files
echo "Removing old files..."
rm -f vps-agent

# Download and build the new version
echo "Building new version..."
go build -o vps-agent .

# Restore config if it was backed up
if [ -f config.yaml.bak ]; then
    echo "Restoring config..."
    mv config.yaml.bak config.yaml
fi

# Start the agent
echo "Starting agent..."
nohup ./vps-agent > agent.log 2>&1 &

echo "Deployment complete! Check agent.log for any issues." 