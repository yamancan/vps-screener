#!/bin/bash

echo "Checking for existing vps-agent processes..."

# Find all vps-agent processes
AGENT_PIDS=$(pgrep -f vps-agent)

if [ -n "$AGENT_PIDS" ]; then
    echo "Found running vps-agent processes with PIDs: $AGENT_PIDS"
    echo "Stopping all vps-agent processes..."
    kill -9 $AGENT_PIDS
    sleep 2
else
    echo "No running vps-agent processes found."
fi

# Check for any remaining processes
REMAINING=$(pgrep -f vps-agent)
if [ -n "$REMAINING" ]; then
    echo "Warning: Some processes could not be stopped. PIDs: $REMAINING"
else
    echo "All vps-agent processes have been stopped."
fi

# Clean up old files
echo "Cleaning up old files..."
rm -f vps-agent
rm -f agent.log

echo "Cleanup complete. You can now run ./fix_versions.sh and ./deploy.sh" 