package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"vps-screener/agent/collector" // For collector.CollectedMetrics type
	"vps-screener/agent/config"
)

// APIPayload defines the structure of the data sent to the API gateway.
// This matches the expected input for the /v1/metrics endpoint.
type APIPayload struct {
	Timestamp    int64                        `json:"timestamp"`      // Unix timestamp (seconds)
	NodeHostname string                       `json:"node_hostname"`
	MetricsData  collector.CollectedMetrics `json:"metrics_data"` // map[string]collector.MetricData
}

// SendMetrics sends the collected metrics to the API gateway.
func SendMetrics(cfg *config.Config, metrics collector.CollectedMetrics) error {
	apiURL := cfg.APIGateway.URL
	apiToken := cfg.APIGateway.Token

	nodeHostname := cfg.AgentSettings.NodeIdentifier
	if nodeHostname == "" {
		hn, err := os.Hostname()
		if err != nil {
			log.Printf("Warning: Could not determine OS hostname: %v. Using 'unknown-host'.", err)
			hn = "unknown-host"
		}
		nodeHostname = hn
	}

	payload := APIPayload{
		Timestamp:    time.Now().Unix(),
		NodeHostname: nodeHostname,
		MetricsData:  metrics,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics payload: %w", err)
	}

	client := &http.Client{
		Timeout: 15 * time.Second, // Configurable timeout
	}

	metricsEndpoint := fmt.Sprintf("%s/metrics", apiURL)
	req, err := http.NewRequest("POST", metricsEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create new HTTP request to %s: %w", metricsEndpoint, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiToken)

	log.Printf("Sending metrics to %s for node %s...", metricsEndpoint, nodeHostname)
	resp, err := client.Do(req)
	if err != nil {
		// TODO: Implement local buffering here as per Technical Blueprint
		log.Printf("Error sending metrics to %s: %v. (Buffering not yet implemented)", metricsEndpoint, err)
		return fmt.Errorf("failed to send HTTP request to %s: %w", metricsEndpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Metrics sent successfully to %s. Status: %s", metricsEndpoint, resp.Status)
		return nil
	} else {
		// TODO: Implement local buffering here as per Technical Blueprint
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Error sending metrics to %s. Status: %s, Body: %s. (Buffering not yet implemented)", metricsEndpoint, resp.Status, string(bodyBytes))
		return fmt.Errorf("API gateway at %s returned error status %s", metricsEndpoint, resp.Status)
	}
} 