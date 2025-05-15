#!/bin/bash

echo "Starting complete cleanup..."

# Stop all vps-agent processes
echo "Stopping all vps-agent processes..."
pkill -f vps-agent || true

# Remove all vps-screener directories
echo "Removing all vps-screener directories..."
rm -rf /root/vps-screener
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

echo "Cleanup complete. You can now run install.sh for a fresh installation." 