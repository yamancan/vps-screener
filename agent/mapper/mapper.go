package mapper

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/elastic/go-sysinfo"
	"github.com/elastic/go-sysinfo/types"
	"vps-screener/agent/config" // Importing our own config package
)

const cgroupPathPattern = "/proc/%d/cgroup"

var (
	dockerLabelsCache      = make(map[string]map[string]string)
	dockerLabelsCacheMutex = &sync.RWMutex{}
	dockerCliNotFound      = false
)

func readCgroupFile(pid int32) ([]string, error) {
	cgroupFile := fmt.Sprintf(cgroupPathPattern, pid)
	file, err := os.Open(cgroupFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // File not existing is not an error for this helper
		}
		return nil, fmt.Errorf("error opening cgroup file %s: %w", cgroupFile, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// GetSystemdServiceForPid attempts to find the systemd service name for a PID.
func GetSystemdServiceForPid(pid int32) (string, error) {
	lines, err := readCgroupFile(pid)
	if err != nil {
		return "", err
	}

	// Regex to find systemd service names. Examples:
	// 0::/system.slice/docker.service
	// 0::/system.slice/my-app.service
	// 0::/user.slice/user-1000.slice/user@1000.service/app.slice/some-gui.service
	re := regexp.MustCompile(`.*/(system|user)\.slice/.*([a-zA-Z0-9_.\-]+)\.service`)

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		// The last submatch is usually the most specific service name
		if len(matches) > 2 {
			serviceName := matches[len(matches)-1] // e.g., my-app.service or user@1000.service
			// If it's a user slice like user@1000.service/app.slice/sub.service, get sub.service
			if strings.HasPrefix(serviceName, "user@") && strings.Contains(line, "/app.slice/") {
				parts := strings.Split(line, "/")
				if len(parts) > 0 && strings.HasSuffix(parts[len(parts)-1], ".service") {
					serviceName = parts[len(parts)-1]
				}
			}
			// Avoid generic session or user slice managers themselves unless explicitly targeted
			if serviceName != "user.slice" && !strings.HasPrefix(serviceName, "session-") && serviceName != fmt.Sprintf("user@%s.service", procUsername(pid)) {
				return serviceName, nil
			}
		}
	}
	return "", nil // Not found or not a systemd managed process in a typical way
}

// Helper to get username, used to filter out generic user services
func procUsername(pid int32) string {
	p, err := sysinfo.ProcessByPID(int(pid))
	if err != nil {
		return ""
	}
	info, err := p.Info()
	if err != nil {
		return ""
	}
	return info.Username
}

// GetDockerContainerIDForPid attempts to find the Docker container ID for a PID.
func GetDockerContainerIDForPid(pid int32) (string, error) {
	lines, err := readCgroupFile(pid)
	if err != nil {
		return "", err
	}
	// Example: 12:pids:/docker/ab358028849698799098f07999e89a148318139af341950e831950e7dd059085
	// Example: 8:memory:/docker/actions_job_12345 // (from within actions runner)
	re := regexp.MustCompile(`:/docker/([0-9a-fA-F]{12,64})(?:\.scope)?$`)
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			return matches[1], nil // Return the full or partial container ID
		}
	}
	return "", nil
}

// GetDockerLabels fetches labels for a given container ID using `docker inspect`.
func GetDockerLabels(containerID string) (map[string]string, error) {
	if containerID == "" {
		return nil, fmt.Errorf("containerID cannot be empty")
	}
	dockerLabelsCacheMutex.RLock()
	if dockerCliNotFound {
		dockerLabelsCacheMutex.RUnlock()
		return nil, fmt.Errorf("docker CLI not found, skipping label fetch")
	}
	labels, found := dockerLabelsCache[containerID]
	dockerLabelsCacheMutex.RUnlock()
	if found {
		return labels, nil
	}

	dockerLabelsCacheMutex.Lock()
	defer dockerLabelsCacheMutex.Unlock()
	// Re-check after acquiring write lock
	labels, found = dockerLabelsCache[containerID]
	if found {
		return labels, nil
	}

	cmd := exec.Command("docker", "inspect", containerID)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("Docker inspect for %s failed: %s, stderr: %s", containerID, err, string(exitErr.Stderr))
		} else if err.Error() == "exec: \"docker\": executable file not found in $PATH" {
			log.Println("Docker CLI not found. Docker label matching will be disabled.")
			dockerCliNotFound = true
			return nil, fmt.Errorf("docker CLI not found: %w", err)
		} else {
			log.Printf("Docker inspect for %s failed: %s", containerID, err)
		}
		return nil, fmt.Errorf("docker inspect command failed for %s: %w", containerID, err)
	}

	var inspectOutput []struct {
		Config struct {
			Labels map[string]string `json:"Labels"`
		} `json:"Config"`
	}
	if err := json.Unmarshal(output, &inspectOutput); err != nil {
		return nil, fmt.Errorf("failed to unmarshal docker inspect output for %s: %w", containerID, err)
	}

	if len(inspectOutput) > 0 && inspectOutput[0].Config.Labels != nil {
		dockerLabelsCache[containerID] = inspectOutput[0].Config.Labels
		return inspectOutput[0].Config.Labels, nil
	}
	dockerLabelsCache[containerID] = map[string]string{} // Cache empty if no labels
	return map[string]string{}, nil
}

// MapPIDToProject determines the project for a given process.
func MapPIDToProject(p types.Process, projectsConfig []config.ProjectConfig) string {
	info, err := p.Info()
	if err != nil {
		return ""
	}

	pinfo := struct {
		Name     string
		Username string
		CmdLine  string
	}{
		Name:     info.Name,
		Username: info.Username,
		CmdLine:  strings.Join(info.Args, " "),
	}

	for _, proj := range projectsConfig {
		match := proj.Match
		// 1. Systemd Unit
		if match.SystemdUnit != "" {
			systemdService, _ := GetSystemdServiceForPid(info.PID)
			if systemdService == match.SystemdUnit {
				log.Printf("PID %d (%s) matched project '%s' by systemd unit: %s", info.PID, pinfo.Name, proj.Name, systemdService)
				return proj.Name
			}
		}

		// 2. Docker Label
		if match.DockerLabel != "" {
			containerID, _ := GetDockerContainerIDForPid(info.PID)
			if containerID != "" {
				labels, err := GetDockerLabels(containerID)
				if err == nil && labels != nil {
					labelKeyVal := strings.SplitN(match.DockerLabel, "=", 2)
					labelKey := labelKeyVal[0]
					expectedValue := ""
					if len(labelKeyVal) > 1 {
						expectedValue = labelKeyVal[1]
					}
					if val, ok := labels[labelKey]; ok && (len(labelKeyVal) == 1 || val == expectedValue) {
						log.Printf("PID %d (%s) matched project '%s' by Docker label: %s=%s", info.PID, pinfo.Name, proj.Name, labelKey, val)
						return proj.Name
					}
				}
			}
		}

		// 3. Process Name
		if match.ProcessName != "" {
			if pinfo.Name == match.ProcessName {
				log.Printf("PID %d (%s) matched project '%s' by process name", info.PID, pinfo.Name, proj.Name)
				return proj.Name
			}
		}

		// 4. Username
		if match.Username != "" {
			if pinfo.Username == match.Username {
				log.Printf("PID %d (%s) matched project '%s' by username: %s", info.PID, pinfo.Name, proj.Name, pinfo.Username)
				return proj.Name
			}
		}

		// 5. Command Line
		if match.CmdLine != "" {
			if strings.Contains(pinfo.CmdLine, match.CmdLine) {
				log.Printf("PID %d (%s) matched project '%s' by command line: %s", info.PID, pinfo.Name, proj.Name, match.CmdLine)
				return proj.Name
			}
		}
	}

	return ""
} 