package config // Note: main.go will refer to this as 'config.LoadConfig'

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the entire agent configuration

type Config struct {
	APIGateway     APIGatewaySettings `yaml:"api_gateway"`
	AgentSettings  AgentSettings    `yaml:"agent_settings"`
	Projects       []ProjectConfig  `yaml:"projects"`
	rawConfig      map[string]interface{} // To store the raw map for debugging or direct access if needed
}

// APIGatewaySettings defines the API gateway connection details
type APIGatewaySettings struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

// AgentSettings defines general agent behaviors
type AgentSettings struct {
	CollectionInterval int    `yaml:"collection_interval"`
	NodeIdentifier     string `yaml:"node_identifier,omitempty"` // omitempty if you want to allow it to be absent
}

// ProjectConfig defines a single project's mapping rules and plugin
type ProjectConfig struct {
	Name   string         `yaml:"name"`
	Match  MatchRules     `yaml:"match"`
	Plugin string         `yaml:"plugin,omitempty"`
}

// MatchRules defines the criteria for mapping a process to a project
type MatchRules struct {
	User                 string `yaml:"user,omitempty"`
	SystemdUnit          string `yaml:"systemd_unit,omitempty"`
	DockerLabel          string `yaml:"docker_label,omitempty"`         // e.g., "com.example.project=ProjectA"
	ContainerNamePattern string `yaml:"container_name_pattern,omitempty"`
	ProcessNamePattern   string `yaml:"process_name_pattern,omitempty"`
}

// LoadConfig reads the YAML configuration file and unmarshals it into the Config struct
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML from %s: %w", filePath, err)
	}

	// Store the raw map as well, useful for debugging or complex lookups
	_ = yaml.Unmarshal(data, &cfg.rawConfig) 

	// Basic Validations
	if cfg.APIGateway.URL == "" {
		return nil, fmt.Errorf("api_gateway.url is required in config")
	}
	if cfg.APIGateway.Token == "" {
		return nil, fmt.Errorf("api_gateway.token is required in config")
	}
	if cfg.AgentSettings.CollectionInterval <= 0 {
		cfg.AgentSettings.CollectionInterval = 30 // Default if invalid
		// Or return an error: return nil, fmt.Errorf("agent_settings.collection_interval must be positive")
	}

	return &cfg, nil
}

// GetRawConfig allows access to the unmarshalled map[string]interface{} representation
// This can be useful if you need to access parts of the config that are not strictly typed
// or for more dynamic processing, though direct struct access is preferred.
func (c *Config) GetRawConfig() map[string]interface{} {
	return c.rawConfig
} 