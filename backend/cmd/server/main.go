package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/tobilg/ai-observer/internal/config"
	"github.com/tobilg/ai-observer/internal/logger"
	"github.com/tobilg/ai-observer/internal/server"
	"github.com/tobilg/ai-observer/internal/version"
)

func main() {
	// Define flags
	var showVersion bool
	var setupTool string

	flag.BoolVar(&showVersion, "v", false, "Show version information")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.StringVar(&setupTool, "s", "", "Show setup instructions for TOOL (claude, gemini, codex)")
	flag.StringVar(&setupTool, "setup", "", "Show setup instructions for TOOL (claude, gemini, codex)")

	// Custom usage function
	flag.Usage = func() {
		fmt.Println(`AI Observer - OpenTelemetry-compatible observability backend for AI coding tools

Usage: ai-observer [options]

Options:
  -h, --help           Show this help message
  -v, --version        Show version information
  -s, --setup TOOL     Show setup instructions for TOOL (claude, gemini, codex)

Environment Variables:
  AI_OBSERVER_API_PORT       API server port (default: 8080)
  AI_OBSERVER_OTLP_PORT      OTLP ingestion port (default: 4318)
  AI_OBSERVER_DATABASE_PATH  DuckDB database path (default: ./data/ai-observer.duckdb)
  AI_OBSERVER_FRONTEND_URL   Frontend URL for CORS (default: http://localhost:5173)
  AI_OBSERVER_LOG_LEVEL      Log level: DEBUG, INFO, WARN, ERROR (default: INFO)`)
	}

	flag.Parse()

	// Handle version flag
	if showVersion {
		fmt.Printf("AI Observer %s\n", version.Version)
		fmt.Printf("Git Commit: %s\n", version.GitCommit)
		fmt.Printf("Build Date: %s\n", version.BuildDate)
		os.Exit(0)
	}

	// Handle setup flag
	if setupTool != "" {
		printSetupInstructions(setupTool)
		os.Exit(0)
	}

	// Initialize structured logging (text format for development readability)
	logLevel := parseLogLevel(os.Getenv("AI_OBSERVER_LOG_LEVEL"))
	logger.InitializeText(logLevel)
	log := logger.Logger()

	cfg := config.Load()

	srv, err := server.New(cfg)
	if err != nil {
		log.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info("Received shutdown signal")
		cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Error("Error during shutdown", "error", err)
		}
		os.Exit(0)
	}()

	log.Info("AI Observer starting",
		"database", cfg.DatabasePath,
		"api_port", cfg.APIPort,
		"otlp_port", cfg.OTLPPort,
	)

	if err := srv.ListenAndServe(); err != nil {
		log.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func printSetupInstructions(tool string) {
	switch tool {
	case "claude":
		fmt.Print(`Claude Code Setup
=================

Add to ~/.bashrc or ~/.zshrc:

# 1. Enable telemetry
export CLAUDE_CODE_ENABLE_TELEMETRY=1

# 2. Choose exporters (both are optional - configure only what you need)
export OTEL_METRICS_EXPORTER=otlp       # Options: otlp, prometheus, console
export OTEL_LOGS_EXPORTER=otlp          # Options: otlp, console

# 3. Configure OTLP endpoint (for OTLP exporter)
export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

# 4. For faster visibilty: reduce export intervals
export OTEL_METRIC_EXPORT_INTERVAL=10000  # 10 seconds (default: 60000ms)
export OTEL_LOGS_EXPORT_INTERVAL=5000     # 5 seconds (default: 5000ms)

# 5. Log prompts
export OTEL_LOG_USER_PROMPTS=1
`)
	case "gemini":
		fmt.Print(`Gemini CLI Setup
================

1. Add to ~/.gemini/settings.json:

{
  "telemetry": {
    "enabled": true,
    "target": "local",
    "otlpProtocol": "http",
    "otlpEndpoint": "http://localhost:4318",
    "logPrompts": true,
    "useCollector": true
  }
}

2. Add to ~/.bashrc or ~/.zshrc:

# Mitigate Gemini CLI bug
export OTEL_METRIC_EXPORT_TIMEOUT=10000
export OTEL_LOGS_EXPORT_TIMEOUT=5000
`)
	case "codex":
		fmt.Print(`OpenAI Codex CLI Setup
======================

Add to ~/.codex/config.toml:

[otel]
exporter = { otlp-http = { endpoint = "http://localhost:4318/v1/logs", protocol = "binary" } }
trace_exporter = { otlp-http = { endpoint = "http://localhost:4318/v1/traces", protocol = "binary" } }
log_user_prompt = true
`)
	default:
		fmt.Printf("Unknown tool: %s\n\nSupported tools: claude, gemini, codex\n", tool)
		os.Exit(1)
	}
}
