package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vps-screener/agent/collector"
	"vps-screener/agent/config"
	"vps-screener/agent/executor"
	"vps-screener/agent/sender"
	// "vps-screener/agent/executor"
)

var ( 
	cfg *config.Config
)

func main() {
	log.Println("Starting VPS Screener Agent (Go version)...")

	var err error
	// Attempt to load config using default path first, then allow override by env var
	configPath := "config.yaml" // Default relative to agent executable
	if envPath := os.Getenv("AGENT_CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	cfg, err = config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v", configPath, err)
	}
	log.Printf("Configuration loaded. Agent settings: %+v", cfg.AgentSettings)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(cfg.AgentSettings.CollectionInterval) * time.Second)
	defer ticker.Stop()

Loop:
	for {
		select {
		case <-ticker.C:
			log.Println("Agent tick: Collecting metrics...")
			collectedMetrics := collector.CollectMetrics(cfg)
			if len(collectedMetrics) > 0 {
				log.Printf("Collected data for %d projects/entities", len(collectedMetrics)-1) // -1 for _system
				err := sender.SendMetrics(cfg, collectedMetrics)
				if err != nil {
					log.Printf("Error sending metrics: %v", err)
				}
			}

			log.Println("Agent tick: Checking for tasks...")
			executor.ProcessTasks(cfg)

			log.Println("Agent tick: Cycle complete.")
		case sig := <-sigs:
			log.Printf("Received signal: %s, shutting down...", sig)
			break Loop
		}
	}

	log.Println("VPS Screener Agent stopped.")
} 