module vps-screener/agent

go 1.21 // Or your desired Go version

require (
	github.com/elastic/go-sysinfo v1.11.1 // For system metrics
	github.com/gorilla/websocket v1.5.1 // For WebSocket communication
)

require (
	github.com/elastic/go-windows v1.0.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/procfs v0.0.0-20190425082905-87a4384522e0 // indirect
	golang.org/x/sys v0.0.0-20191026070338-33540a1f6037 // indirect
) 