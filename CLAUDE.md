# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AI Observer is an OpenTelemetry-compatible observability backend designed for monitoring AI coding tools (Claude Code, OpenAI Codex CLI, Gemini CLI). It provides real-time ingestion of OTLP traces, metrics, and logs with a DuckDB-based storage layer and a React dashboard.

## Build & Development Commands

```bash
# Setup
make setup              # Install all dependencies (go mod download + pnpm install)

# Development
make dev                # Run both backend and frontend in dev mode (parallel)
make backend-dev        # Run only backend (go run ./cmd/server)
make frontend-dev       # Run only frontend (pnpm dev at localhost:5173)

# Build
make all                # Build both backend and frontend
make backend            # Build backend binary to bin/ai-observer
make frontend           # Build frontend for production

# Testing
make test               # Run all tests
make backend-test       # Run Go tests: cd backend && go test -v ./...
make frontend-test      # Run frontend tests: cd frontend && pnpm test

# Run a single Go test
cd backend && go test -v -run TestFunctionName ./path/to/package

# Run a single frontend test
cd frontend && pnpm vitest run src/path/to/file.test.ts

# Linting
make lint               # Run all linters
make backend-lint       # Run golangci-lint
make frontend-lint      # Run ESLint

# Cleanup
make clean              # Remove bin/ and frontend/dist/
```

## Architecture

### Backend (Go)

The backend runs **two HTTP servers** simultaneously:

1. **OTLP Ingestion Server** (port 4318) - Receives telemetry from AI tools
   - `POST /v1/traces` - Trace data (protobuf or JSON)
   - `POST /v1/metrics` - Metrics data
   - `POST /v1/logs` - Log data
   - `POST /` - Auto-detects signal type (Gemini CLI sends to root path instead of `/v1/*`)
   - Supports HTTP/1.1 + h2c (HTTP/2 cleartext), gzip-compressed payloads
   - Auto-detects JSON vs Protobuf format regardless of Content-Type header

2. **API/WebSocket Server** (port 8080) - Serves dashboard and real-time updates
   - `/api/traces`, `/api/metrics`, `/api/logs` - Query endpoints
   - `/api/services`, `/api/stats` - Aggregations
   - `/api/dashboards` - Dashboard CRUD
   - `/ws` - WebSocket for real-time updates to frontend
   - Serves embedded React frontend at root

**Key packages:**
- `internal/otlp/` - Format detection (`format_detector.go`) and decoder interface with proto/JSON implementations
- `internal/storage/` - DuckDB storage layer with separate stores for traces, logs, metrics
- `internal/handlers/` - HTTP handlers for OTLP ingestion (`otlp_*.go`) and query API (`query.go`)
- `internal/websocket/` - Hub/client pattern for real-time broadcasting
- `internal/server/` - Server setup, routing configuration
- `pkg/compression/` - GZIP decompression middleware for incoming OTLP data

**Configuration (environment variables):**
- `AI_OBSERVER_API_PORT` - HTTP server port (default: 8080)
- `AI_OBSERVER_OTLP_PORT` - OTLP ingestion port (default: 4318)
- `AI_OBSERVER_DATABASE_PATH` - DuckDB file path (default: ./data/ai-observer.duckdb)
- `AI_OBSERVER_FRONTEND_URL` - CORS allowed origin (default: http://localhost:5173)

### Frontend (React + TypeScript)

Vite-based React app with Tailwind CSS v4:
- `react-router-dom` for routing (Dashboard, Traces, Metrics, Logs pages)
- `zustand` for state management (`stores/telemetryStore.ts` buffers real-time data)
- `recharts` for visualizations
- `@dnd-kit` for drag-and-drop dashboard widgets
- WebSocket singleton (`hooks/useWebSocket.ts`) for live updates

**Key files:**
- `lib/api.ts` - API client with request deduplication
- `stores/dashboardStore.ts` - Dashboard widget state
- `pages/` - TracesPage, MetricsPage, LogsPage, Dashboard

**Path alias:** `@` maps to `./src`

**Dev proxy:** Vite proxies `/api` and `/ws` to backend at localhost:8080

**UI Components:** Always use [shadcn/ui](https://ui.shadcn.com) components from the registry, see [overview](https://ui.shadcn.com/llms.txt) also. Use the shadcn CLI to add new components:

```bash
cd frontend && pnpm dlx shadcn@latest add [component]  # e.g., button, dialog, table
```

Existing components are in `src/components/ui/`. Never manually create UI primitives - add them via CLI instead.

### Database Schema

DuckDB tables defined in `internal/storage/schema.go`:
- `otel_traces` - Span data with nested events and links arrays
- `otel_logs` - Log records with severity and trace context
- `otel_metrics` - All metric types (gauge, sum, histogram, summary, exponential histogram) unified in one table
- `dashboards` / `dashboard_widgets` - User dashboard persistence

All tables indexed on `Timestamp`, `ServiceName`, and relevant query fields.

## External Reference Projects

The `external-projects/` directory contains local checkouts of the AI coding tools that send telemetry to AI Observer:
- `external-projects/claude-code/` - Claude Code CLI source
- `external-projects/gemini-cli/` - Gemini CLI source
- `external-projects/codex/` - OpenAI Codex CLI source
- `external-projects/CodexBar` - SwiftUI application showing cost and usage metrics from different providers
- `external-projects/ccusage` - CLI tool for analyzing Claude Code/Codex CLI usage from local JSONL files

**Use these local sources** instead of web search when investigating OTLP telemetry formats, metric names, log events, or span structures. These repos show exactly what telemetry data each tool emits.

## AI Coding Tool Integration

Point AI tools to OTLP endpoint `http://localhost:4318`:

```bash
# Claude Code
export CLAUDE_CODE_ENABLE_TELEMETRY=1
export OTEL_METRICS_EXPORTER=otlp
export OTEL_LOGS_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

# Gemini CLI - add to ~/.gemini/settings.json
# { "telemetry": { "enabled": true, "target": "local", "useCollector": true, "otlpEndpoint": "http://localhost:4318" } }

# OpenAI Codex CLI - add to ~/.codex/config.toml
# [otel]
# exporter = { otlp-http = { endpoint = "http://localhost:4318/v1/logs", protocol = "binary" } }
```
