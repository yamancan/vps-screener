package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the agent\'s configuration, loaded from config.yaml
type Config struct {
	ApiGateway APIGatewayConfig         \`yaml:"api_gateway"\`
	Interval   int                      \`yaml:"interval"\`      // Metrics collection interval in seconds
	Projects   map[string]ProjectConfig \`yaml:"projects"\` // Map from project name to its configuration
	// Add other global configurations here if needed, e.g., log level
}

// APIGatewayConfig contains the configuration for connecting to the API gateway
type APIGatewayConfig struct {
	URL   string \`yaml:"url"\`
	Token string \`yaml:"token"\`
}

// ProjectConfig defines the configuration for a single monitored project
type ProjectConfig struct {
	Match      ProjectMatch \`yaml:"match"\`                 // Rules to identify processes belonging to this project
	PluginPath string       \`yaml:"plugin,omitempty"\` // Optional path to a plugin executable for custom metrics
	// Add other project-specific configurations here, e.g., resource limits for plugins/tasks
}

// ProjectMatch specifies the criteria for matching processes to a project.
// At least one field should be non-empty for a valid match.
type ProjectMatch struct {
	SystemdService     string \`yaml:"systemd_service,omitempty"\`     // e.g., "my-app.service"
	DockerLabel        string \`yaml:"docker_label,omitempty"\`        // e.g., "com.mycompany.project=my-app"
	DockerContainerName string \`yaml:"docker_container_name,omitempty"\` // e.g., "my-app-container" (exact match)
	// Consider adding DockerContainerNamePattern for regex matching later.
	Username           string \`yaml:"user,omitempty"\`                // e.g., "appuser"
	ProcessNamePattern string \`yaml:"process_name_pattern,omitempty"\`  // Regex, e.g., "^/opt/my-app/bin/server"
	CgroupPathPattern  string \`yaml:"cgroup_path_pattern,omitempty"\` // Regex, e.g., ".*/docker-[a-f0-9]{64}\\.scope$"
}

// LoadConfig reads the configuration file from the given path and parses it.
// If no path is provided, it defaults to "config.yaml" in the current directory.
// It also looks for AGENT_CONFIG_PATH environment variable.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = os.Getenv("AGENT_CONFIG_PATH")
		if path == "" {
			path = "config.yaml" // Default config file name
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// Expand environment variables in the config file content
	expandedData := os.ExpandEnv(string(data))

	var cfg Config
	err = yaml.Unmarshal([]byte(expandedData), &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	// Basic validation (can be expanded)
	if cfg.Interval <= 0 {
		cfg.Interval = 60 // Default to 60 seconds if not specified or invalid
	}
	if cfg.ApiGateway.URL == "" {
		return nil, fmt.Errorf("api_gateway.url is not set in config")
	}

	return &cfg, nil
} 