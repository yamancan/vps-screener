#!/bin/bash

echo "Getting latest changes..."

# Pull latest changes
git pull origin main

# Download any missing scripts
echo "Downloading missing scripts..."
for script in update_and_run.sh cleanup.sh fix_versions.sh; do
    if [ ! -f "$script" ]; then
        echo "Downloading $script..."
        curl -O "https://raw.githubusercontent.com/yamancan/vps-screener/main/agent/$script"
        chmod +x "$script"
    fi
done

echo "Making all scripts executable..."
chmod +x *.sh

echo "Latest changes obtained. You can now run:"
echo "./update_and_run.sh" 