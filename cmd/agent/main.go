// DataLens 2.0 — On-Premise Agent
//
// The agent runs inside the customer's infrastructure. It scans local
// data sources, executes DSR tasks, and reports metadata (never PII)
// back to the Control Centre platform.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/complyark/datalens/internal/config"
	"github.com/complyark/datalens/pkg/logging"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logging.New(cfg.App.Env, cfg.App.LogLevel).
		WithComponent("agent")

	log.Info("Starting DataLens Agent",
		"agent_id", cfg.Agent.ID,
		"control_centre_endpoint", cfg.Agent.ControlCentreEndpoint,
	)

	// =========================================================================
	// Validate Agent Configuration
	// =========================================================================

	if cfg.Agent.ID == "" || cfg.Agent.APIKey == "" {
		log.Error("AGENT_ID and AGENT_API_KEY are required")
		os.Exit(1)
	}

	// =========================================================================
	// Initialize Infrastructure
	// =========================================================================

	// TODO: Initialize local database (SQLite or embedded PostgreSQL)
	// TODO: Initialize connector registry
	// TODO: Initialize scan scheduler
	// TODO: Initialize Control Centre API client

	// =========================================================================
	// Register Connectors
	// =========================================================================

	// TODO: Register database connectors (PostgreSQL, MySQL, etc.)
	// TODO: Register file connectors (filesystem, S3, etc.)
	// TODO: Register application connectors (Salesforce, etc.)

	// =========================================================================
	// Start Agent Services
	// =========================================================================

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: Start scan scheduler
	// TODO: Start DSR task executor
	// TODO: Start health reporter (heartbeat to Control Centre)
	// TODO: Start metric collector

	log.Info("Agent started successfully", "agent_id", cfg.Agent.ID)

	// =========================================================================
	// Wait for Shutdown
	// =========================================================================

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	log.Info("Shutdown signal received", "signal", sig.String())
	cancel()

	// TODO: Gracefully stop scheduler
	// TODO: Wait for in-flight scans to complete
	// TODO: Report final status to Control Centre

	_ = ctx // Suppresses unused variable warning until services are wired up.
	log.Info("Agent stopped gracefully")
}
