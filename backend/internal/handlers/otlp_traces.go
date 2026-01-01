package handlers

import (
	"bytes"
	"io"
	"net/http"

	"github.com/tobilg/ai-observer/internal/api"
	"github.com/tobilg/ai-observer/internal/logger"
	"github.com/tobilg/ai-observer/internal/otlp"
	"github.com/tobilg/ai-observer/internal/storage"
	"github.com/tobilg/ai-observer/internal/websocket"
)

type Handlers struct {
	store *storage.DuckDBStore
	hub   *websocket.Hub
}

func New(store *storage.DuckDBStore, hub *websocket.Hub) *Handlers {
	return &Handlers{
		store: store,
		hub:   hub,
	}
}

// HandleRoot handles POST / by detecting signal type from body (workaround for Gemini CLI bug)
func (h *Handlers) HandleRoot(w http.ResponseWriter, r *http.Request) {
	log := logger.Logger()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("Failed to read body", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Detect signal type from JSON body
	r.Body = io.NopCloser(bytes.NewReader(body)) // Reset body for handler

	switch {
	case bytes.Contains(body, []byte(`"resourceSpans"`)):
		log.Debug("Routing POST / to traces handler")
		h.HandleTraces(w, r)
	case bytes.Contains(body, []byte(`"resourceMetrics"`)):
		log.Debug("Routing POST / to metrics handler")
		h.HandleMetrics(w, r)
	case bytes.Contains(body, []byte(`"resourceLogs"`)):
		log.Debug("Routing POST / to logs handler")
		h.HandleLogs(w, r)
	default:
		log.Warn("Unknown signal type in POST /", "body_preview", string(body[:min(200, len(body))]))
		w.WriteHeader(http.StatusBadRequest)
	}
}

// HandleTraces handles POST /v1/traces
func (h *Handlers) HandleTraces(w http.ResponseWriter, r *http.Request) {
	log := logger.Logger()
	contentType := r.Header.Get("Content-Type")

	// Use format detection to handle Content-Type mismatches
	decoder, body, _, err := otlp.GetDecoderWithDetection(r.Body, contentType)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	req, err := decoder.DecodeTraces(body)
	if err != nil {
		log.Error("Failed to decode traces", "error", err)
		api.WriteError(w, http.StatusBadRequest, "failed to decode traces: "+err.Error())
		return
	}

	spans := otlp.ConvertTraces(req)

	// Store spans as-is - Codex CLI spans are handled at query time
	if err := h.store.InsertSpans(r.Context(), spans); err != nil {
		log.Error("Failed to store traces", "error", err)
		api.WriteError(w, http.StatusInternalServerError, "failed to store traces")
		return
	}

	// Broadcast to WebSocket clients
	if h.hub != nil && len(spans) > 0 {
		h.hub.Broadcast(websocket.NewTracesMessage(spans))
	}

	log.Debug("Received spans", "count", len(spans))

	// OTLP success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

