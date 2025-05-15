#!/bin/bash

# Create config directory if it doesn't exist
mkdir -p config

# Move config.go to config directory if it exists in root
if [ -f config.go ]; then
    echo "Moving config.go to config directory..."
    mv config.go config/
fi

# Update go.mod
echo "Updating go.mod..."
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
	github.com/prometheus/procfs v0.12.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
)
EOL

# Update imports in all Go files
echo "Updating imports in Go files..."
find . -name "*.go" -type f -exec sed -i 's|vps-screener/agent/|vps-agent/|g' {} +

# Clean and rebuild
echo "Cleaning and rebuilding..."
go mod tidy
go build -o vps-agent .

echo "Package organization fix complete. Try running ./deploy.sh again." 