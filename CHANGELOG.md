# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-01-01

### Added

- Initial release of AI Observer
- OpenTelemetry-compatible OTLP ingestion (HTTP/JSON and HTTP/Protobuf)
- Support for traces, metrics, and logs
- DuckDB-powered storage for fast analytics
- Real-time dashboard with WebSocket updates
- Customizable drag-and-drop dashboard widgets
- Multi-tool support:
  - Claude Code
  - Gemini CLI
  - OpenAI Codex CLI
- Query API for traces, metrics, logs, and dashboards
- Single binary with embedded React frontend
- Multi-arch Docker images (linux/amd64, linux/arm64)
- Homebrew formula for macOS (Apple Silicon)

### Technical Details

- Backend: Go 1.24, chi router, DuckDB, gorilla/websocket
- Frontend: React 19, TypeScript, Vite, Tailwind CSS v4, Zustand, Recharts
- OTLP ingestion on port 4318
- API/Dashboard on port 8080

[Unreleased]: https://github.com/tobilg/ai-observer/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/tobilg/ai-observer/releases/tag/v0.1.0
