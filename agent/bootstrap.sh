#!/bin/bash

echo "Starting bootstrap process..."

# Create a temporary directory
echo "Creating temporary directory..."
mkdir -p /tmp/vps-setup
cd /tmp/vps-setup

# Download the fresh start script
echo "Downloading fresh start script..."
curl -L -o fresh_start.sh https://raw.githubusercontent.com/yamancan/vps-screener/main/agent/fresh_start.sh

# Make it executable
echo "Making script executable..."
chmod +x fresh_start.sh

# Run the script
echo "Running fresh start script..."
./fresh_start.sh

# Clean up
echo "Cleaning up temporary files..."
cd /
rm -rf /tmp/vps-setup

echo "Bootstrap complete!" 