#!/bin/bash

# Update go.mod
echo "Updating go.mod..."
sed -i 's|module vps-screener/agent|module vps-agent|' go.mod

# Update imports in all Go files
echo "Updating imports in Go files..."
find . -name "*.go" -type f -exec sed -i 's|vps-screener/agent/|vps-agent/|g' {} +

# Clean and rebuild
echo "Cleaning and rebuilding..."
go mod tidy
go build -o vps-agent .

echo "Build fix complete. Try running ./deploy.sh again." 