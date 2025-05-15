package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall" // For resource limits, though not fully implemented here yet
	"time"

	"vps-screener/agent/config"
)

// Task defines the structure of a task received from the API.
type Task struct {
	ID        string `json:"id"`       // Assuming task ID is a string (e.g., UUID)
	Cmd       string `json:"cmd"`      // The command to execute
	ProjectID string `json:"project_id,omitempty"` // Optional: for context
}

// TaskResult defines the structure for sending back the outcome of a task.
type TaskResult struct {
	Status string `json:"status"` // e.g., "completed", "failed", "timed_out", "error"
	Output string `json:"output"` // Combined stdout and stderr
}

func getHostname(cfg *config.Config) string {
	nodeHostname := cfg.AgentSettings.NodeIdentifier
	if nodeHostname == "" {
		hn, err := os.Hostname()
		if err != nil {
			log.Printf("Warning: Could not determine OS hostname: %v. Using 'unknown-host'.", err)
			return "unknown-host"
		}
		return hn
	}
	return nodeHostname
}

// FetchTasks retrieves pending tasks for this node from the API gateway.
func FetchTasks(cfg *config.Config) ([]Task, error) {
	apiURL := cfg.APIGateway.URL
	apiToken := cfg.APIGateway.Token
	nodeHostname := getHostname(cfg)

	client := &http.Client{
		Timeout: 10 * time.Second, // Short timeout for fetching tasks
	}

	tasksEndpoint := fmt.Sprintf("%s/tasks?node=%s", apiURL, nodeHostname)
	req, err := http.NewRequest("GET", tasksEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for tasks: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Accept", "application/json")

	// log.Printf("Fetching tasks from %s...", tasksEndpoint) // Can be verbose
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks from %s: %w", tasksEndpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := os.ReadAll(resp.Body)
		log.Printf("Error fetching tasks from %s. Status: %s, Body: %s", tasksEndpoint, resp.Status, string(bodyBytes))
		return nil, fmt.Errorf("API gateway at %s returned error status %s for tasks", tasksEndpoint, resp.Status)
	}

	var tasks []Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks JSON from %s: %w", tasksEndpoint, err)
	}

	if len(tasks) > 0 {
		log.Printf("Fetched %d task(s) from %s", len(tasks), tasksEndpoint)
	}
	return tasks, nil
}

// ExecuteTask runs a single task command.
// WARNING: Directly executing commands via shell (sh -c) is a security risk.
// The Technical Blueprint mentions "forks shell with resource.setrlimit() safeguards".
// Proper sandboxing and resource limiting are complex and crucial for production.
func ExecuteTask(task Task) TaskResult {
	log.Printf("Executing task ID %s: %s", task.ID, task.Cmd)

	// For production, avoid "sh -c" if possible, or heavily sanitize/validate commands.
	// Implement resource limits (CPU, memory, time) using syscall.Setrlimit
	// This is a simplified execution for now.
	cmd := exec.Command("sh", "-c", task.Cmd)

	// TODO: Implement syscall.Setrlimit for the new process
	// This is OS-specific (Linux example below, needs refinement)
	/*
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true, // Important for managing the process group
			// Pdeathsig: syscall.SIGKILL, // Kill child if parent dies
		}
		// Example: Set CPU time limit to 60 seconds
		rlimitCpu := syscall.Rlimit{Cur: 60, Max: 60}
		// In a real scenario, you would apply this to the child process after fork/exec
		// or use cgroups. This is non-trivial with os/exec alone directly.
	*/

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	startTime := time.Now()
	err := cmd.Start() // Start the command
	if err != nil {
		log.Printf("Failed to start command for task %s: %v", task.ID, err)
		return TaskResult{Status: "error", Output: fmt.Sprintf("Failed to start command: %v", err)}
	}

	// Timeout for the command execution (e.g., 5 minutes)
	// This is a simpler timeout than process-level CPU limits.
	cmdTimeout := 5 * time.Minute
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait() // Wait for the command to finish
	}()

	var taskStatus string
	var combinedOutput string

	select {
	case <-time.After(cmdTimeout):
		log.Printf("Command for task %s timed out after %v", task.ID, cmdTimeout)
		if err := cmd.Process.Kill(); err != nil {
			log.Printf("Failed to kill timed-out process for task %s: %v", task.ID, err)
		}
		taskStatus = "timed_out"
		combinedOutput = fmt.Sprintf("Task timed out after %v.\n%s\n%s", cmdTimeout, outb.String(), errb.String())
	case err := <-done:
		stdoutStr := outb.String()
		stderrStr := errb.String()
		combinedOutput = fmt.Sprintf("STDOUT:\n%s\nSTDERR:\n%s", stdoutStr, stderrStr)

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				log.Printf("Command for task %s finished with error: %v. Exit code: %d", task.ID, err, exitErr.ExitCode())
				taskStatus = "failed"
			} else {
				log.Printf("Command for task %s failed to run or finished with unknown error: %v", task.ID, err)
				taskStatus = "error"
			}
		} else {
			log.Printf("Command for task %s completed successfully.", task.ID)
			taskStatus = "completed"
		}
	}
	log.Printf("Task %s finished in %v. Status: %s", task.ID, time.Since(startTime), taskStatus)
	return TaskResult{Status: taskStatus, Output: combinedOutput}
}

// SendTaskResult sends the outcome of a completed task back to the API.
func SendTaskResult(cfg *config.Config, taskID string, result TaskResult) error {
	apiURL := cfg.APIGateway.URL
	apiToken := cfg.APIGateway.Token

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal task result for %s: %w", taskID, err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resultEndpoint := fmt.Sprintf("%s/tasks/%s/result", apiURL, taskID)
	req, err := http.NewRequest("POST", resultEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request for task result %s: %w", taskID, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiToken)

	log.Printf("Sending result for task %s to %s...", taskID, resultEndpoint)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send task result for %s to %s: %w", taskID, resultEndpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated { // Allow 200 or 201
		bodyBytes, _ := os.ReadAll(resp.Body)
		log.Printf("Error sending task result for %s. Status: %s, Body: %s", taskID, resp.Status, string(bodyBytes))
		return fmt.Errorf("API gateway at %s returned error status %s for task result %s", resultEndpoint, resp.Status, taskID)
	}

	log.Printf("Successfully sent result for task %s. Status: %s", taskID, resp.Status)
	return nil
}

// ProcessTasks fetches, executes, and reports results for tasks.
// This function can be called in the main agent loop.
func ProcessTasks(cfg *config.Config) {
	tasks, err := FetchTasks(cfg)
	if err != nil {
		log.Printf("Error fetching tasks: %v", err)
		return
	}

	if len(tasks) == 0 {
		return // No tasks to process
	}

	for _, task := range tasks {
		// Consider running tasks in parallel with goroutines if they are independent
		// and the system can handle it. For now, sequential execution.
		result := ExecuteTask(task)
		if err := SendTaskResult(cfg, task.ID, result); err != nil {
			log.Printf("Error sending task result for task ID %s: %v", task.ID, err)
			// Decide if agent should retry sending result or just log
		}
	}
} 