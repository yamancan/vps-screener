#!/bin/bash

echo "Updating repository..."
git pull origin main

echo "Making all scripts executable..."
chmod +x *.sh

# Check if cleanup.sh exists
if [ ! -f "cleanup.sh" ]; then
    echo "Downloading cleanup.sh..."
    curl -O https://raw.githubusercontent.com/yamancan/vps-screener/main/agent/cleanup.sh
    chmod +x cleanup.sh
fi

# Check if fix_versions.sh exists
if [ ! -f "fix_versions.sh" ]; then
    echo "Downloading fix_versions.sh..."
    curl -O https://raw.githubusercontent.com/yamancan/vps-screener/main/agent/fix_versions.sh
    chmod +x fix_versions.sh
fi

echo "Running scripts in order..."

echo "1. Running cleanup.sh..."
./cleanup.sh

echo "2. Running fix_versions.sh..."
./fix_versions.sh

echo "3. Running deploy.sh..."
./deploy.sh

echo "Checking if agent is running..."
ps aux | grep vps-agent

echo "Checking logs..."
tail -f agent.log 