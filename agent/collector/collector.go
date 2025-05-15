package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/elastic/go-sysinfo"
	"github.com/elastic/go-sysinfo/types"

	"vps-screener/agent/config"
	"vps-screener/agent/mapper" // Import the mapper package
)

// MetricData holds the collected metrics for a single project or the system.
// We can expand this struct as more specific metrics are added.
type MetricData struct {
	CPUPercent    float64            `json:"cpu_percent"`
	RAMBytes      uint64             `json:"ram_bytes,omitempty"`
	RAMPercent    float32            `json:"ram_percent,omitempty"` // for _system
	DiskPercent   float64            `json:"disk_percent,omitempty"`// for _system
	ProcessCount  int                `json:"process_count,omitempty"`
	CustomMetrics map[string]interface{} `json:"custom,omitempty"`
}

// CollectedMetrics is a map of project name to its MetricData.
// The key can be a project name or "_system" for overall system metrics.
type CollectedMetrics map[string]MetricData

// executePlugin runs a plugin executable and returns its JSON output.
// It sets a timeout and passes the project name as an environment variable.
func executePlugin(pluginPath string, projectName string) (map[string]interface{}, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, pluginPath)
	
	// Set working directory to plugin's directory
	cmd.Dir = filepath.Dir(pluginPath)
	
	// Set environment variable for project name
	cmd.Env = append(os.Environ(), fmt.Sprintf("VPS_PROJECT_NAME=%s", projectName))

	// Capture stdout
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("plugin execution timed out after 10 seconds")
		}
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse plugin output as JSON: %w", err)
	}

	return result, nil
}

// CollectMetrics gathers metrics for all configured projects and overall system.
func CollectMetrics(cfg *config.Config) CollectedMetrics {
	metrics := make(CollectedMetrics)

	// 1. Overall System Metrics
	host, err := sysinfo.Host()
	if err != nil {
		log.Printf("Error getting host info: %v", err)
		return metrics
	}

	cpuInfo, err := host.CPUTime()
	if err != nil {
		log.Printf("Error getting CPU info: %v", err)
	} else {
		metrics["_system"] = MetricData{
			CPUPercent: cpuInfo.User + cpuInfo.System,
		}
	}

	memInfo, err := host.Memory()
	if err != nil {
		log.Printf("Error getting memory info: %v", err)
	} else {
		systemMetrics := metrics["_system"]
		systemMetrics.RAMPercent = float32(float64(memInfo.Used) / float64(memInfo.Total) * 100)
		metrics["_system"] = systemMetrics
	}

	// 2. Per-Project Metrics
	processes, err := sysinfo.Processes()
	if err != nil {
		log.Printf("Error getting process list: %v", err)
		return metrics
	}

	// Track which projects we've processed plugins for
	processedPlugins := make(map[string]bool)

	for _, p := range processes {
		projectName := mapper.MapPIDToProject(p, cfg.Projects)

		// Ensure project entry exists
		currentProjectMetrics, ok := metrics[projectName]
		if !ok {
			currentProjectMetrics = MetricData{
				CustomMetrics: make(map[string]interface{}),
			}
		}

		// Get process info
		procInfo, err := p.Info()
		if err != nil {
			continue
		}

		// Get CPU usage
		cpuInfo, err := p.CPUTime()
		if err == nil {
			currentProjectMetrics.CPUPercent += cpuInfo.User + cpuInfo.System
		}

		// Get memory usage
		memInfo, err := p.Memory()
		if err == nil {
			currentProjectMetrics.RAMBytes += memInfo.RSS
		}

		currentProjectMetrics.ProcessCount++

		// Execute plugin if configured and not yet executed for this project
		if projectConfig, exists := cfg.Projects[projectName]; exists && !processedPlugins[projectName] {
			if projectConfig.PluginPath != "" {
				customMetrics, err := executePlugin(projectConfig.PluginPath, projectName)
				if err != nil {
					log.Printf("Error executing plugin for project %s: %v", projectName, err)
					currentProjectMetrics.CustomMetrics["plugin_error"] = err.Error()
				} else {
					currentProjectMetrics.CustomMetrics = customMetrics
				}
				processedPlugins[projectName] = true
			}
		}

		metrics[projectName] = currentProjectMetrics
	}

	log.Printf("Collected metrics for %d projects/entities.", len(metrics))
	return metrics
} 