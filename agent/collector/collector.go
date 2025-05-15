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

	// Overall CPU Time (as a placeholder, not a percentage yet)
	hostCPUTimes, err := host.CPUTime() // MODIFIED: Renamed for clarity
	if err != nil {
		log.Printf("Error getting host CPU times: %v", err) // MODIFIED: Log message
	} else {
		// Store total CPU time in seconds. THIS IS NOT A LIVE PERCENTAGE.
		// Proper percentage calculation requires sampling over an interval.
		// Ensure metrics["_system"] is initialized if it's the first metric being set.
		systemMetricData := metrics["_system"]
		systemMetricData.CPUPercent = hostCPUTimes.Total().Seconds() // MODIFIED: Call Total() as a method
		metrics["_system"] = systemMetricData
	}

	// Overall Memory
	hostMemInfo, err := host.Memory() // MODIFIED: Renamed for clarity
	if err != nil {
		log.Printf("Error getting host memory info: %v", err) // MODIFIED: Log message
	} else {
		systemMetrics := metrics["_system"] // Get existing or newly created
		systemMetrics.RAMPercent = float32(float64(hostMemInfo.Used) / float64(hostMemInfo.Total) * 100)
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
		if projectName == "" { // MODIFIED: Skip processes not mapped to any project
			continue
		}

		// Ensure project entry exists
		currentProjectMetrics, ok := metrics[projectName]
		if !ok {
			currentProjectMetrics = MetricData{
				CustomMetrics: make(map[string]interface{}),
			}
		}

		// Get process CPU usage percentage // MODIFIED BLOCK to use CPUTime().Total().Seconds()
		// (placeholder: total CPU time in seconds, not live percentage)
		procCPUTimes, cpuErr := p.CPUTime()
		if cpuErr == nil {
			currentProjectMetrics.CPUPercent += procCPUTimes.Total().Seconds() // Accumulating total CPU time in seconds
		} else {
			log.Printf("Error getting CPU time for PID %d: %v", p.PID(), cpuErr)
		}

		// Get process memory usage // MODIFIED BLOCK
		procMemInfo, memErr := p.Memory() // Renamed for clarity
		if memErr == nil {
			currentProjectMetrics.RAMBytes += procMemInfo.Resident // Use .Resident
		} else {
			log.Printf("Error getting memory info for PID %d: %v", p.PID(), memErr)
		}

		currentProjectMetrics.ProcessCount++

		// Execute plugin if configured and not yet executed for this project // MODIFIED BLOCK
		var projectRuleForPlugin *config.ProjectConfig
		for i := range cfg.Projects { // Iterate to find the project rule by name
			if cfg.Projects[i].Name == projectName {
				projectRuleForPlugin = &cfg.Projects[i]
				break
			}
		}

		if projectRuleForPlugin != nil && projectRuleForPlugin.Plugin != "" && !processedPlugins[projectName] {
			pluginExecutablePath := filepath.Join("plugins", projectRuleForPlugin.Plugin) // Construct path relative to 'plugins' dir
			
			customMetrics, pluginErr := executePlugin(pluginExecutablePath, projectName)
			if pluginErr != nil {
				log.Printf("Error executing plugin %s for project %s: %v", pluginExecutablePath, projectName, pluginErr)
                if currentProjectMetrics.CustomMetrics == nil { // Ensure map is initialized before adding error
                    currentProjectMetrics.CustomMetrics = make(map[string]interface{})
                }
				// Use a distinct key for plugin errors to avoid overwriting other custom metrics
				currentProjectMetrics.CustomMetrics["plugin_error_"+projectRuleForPlugin.Plugin] = pluginErr.Error()
			} else {
                if currentProjectMetrics.CustomMetrics == nil { // Ensure map is initialized
                    currentProjectMetrics.CustomMetrics = make(map[string]interface{})
                }
				// Merge plugin metrics.
				for k, v := range customMetrics {
					currentProjectMetrics.CustomMetrics[k] = v
				}
			}
			processedPlugins[projectName] = true
		}

		metrics[projectName] = currentProjectMetrics
	}

	log.Printf("Collected metrics for %d projects/entities.", len(metrics))
	return metrics
} 