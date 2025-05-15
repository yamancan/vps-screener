#!/bin/bash

# Define the base directory
BASE_DIR="/root/vps-screener"
AGENT_DIR="$BASE_DIR/agent"

echo "Setting up directory structure..."

# Create directories if they don't exist
mkdir -p "$AGENT_DIR"
mkdir -p "$AGENT_DIR/config"

# Move to the agent directory
cd "$AGENT_DIR"

# Clone the repository if it doesn't exist
if [ ! -d ".git" ]; then
    echo "Cloning repository..."
    git clone https://github.com/yamancan/vps-screener.git .
fi

# Make all scripts executable
chmod +x *.sh

echo "Directory structure setup complete."
echo "Current directory: $(pwd)"
echo "Available files:"
ls -la

echo "You can now run the scripts in this order:"
echo "1. ./cleanup.sh"
echo "2. ./fix_versions.sh"
echo "3. ./deploy.sh" 