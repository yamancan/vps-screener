#!/bin/bash

echo "Checking VPS for agent installations..."

# Check common locations
echo "Checking common locations..."
for dir in /root/vps-screener /opt/vps-screener /usr/local/vps-screener /home/*/vps-screener; do
    if [ -d "$dir" ]; then
        echo "Found directory: $dir"
        ls -la "$dir"
    fi
done

# Check for running processes
echo "Checking for running vps-agent processes..."
ps aux | grep -i "vps-agent" | grep -v grep

# Check for service files
echo "Checking for service files..."
find /etc/systemd/system /lib/systemd/system -name "*vps*" -o -name "*agent*" 2>/dev/null

# Check for cron jobs
echo "Checking for cron jobs..."
for user in $(cut -f1 -d: /etc/passwd); do
    echo "Checking crontab for user $user:"
    crontab -u "$user" -l 2>/dev/null | grep -i "vps\|agent"
done

# Check for any vps-agent binaries
echo "Checking for vps-agent binaries..."
find / -name "vps-agent" -type f 2>/dev/null

# Check current directory
echo "Current directory: $(pwd)"
echo "Current directory contents:"
ls -la

echo "Check complete. Please review the output above." 