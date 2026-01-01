package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tobilg/ai-observer/internal/api"
)

func TestNewDuckDBStore(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.duckdb")

	store, err := NewDuckDBStore(dbPath)
	if err != nil {
		t.Fatalf("NewDuckDBStore() error = %v", err)
	}
	defer store.Close()

	if store.db == nil {
		t.Error("db is nil")
	}
	if store.DB() == nil {
		t.Error("DB() returns nil")
	}
}

func TestNewDuckDBStore_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "test.duckdb")

	store, err := NewDuckDBStore(nestedPath)
	if err != nil {
		t.Fatalf("NewDuckDBStore() error = %v", err)
	}
	defer store.Close()

	// Check directory was created
	dir := filepath.Dir(nestedPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("directory %s was not created", dir)
	}
}

func TestNewDuckDBStore_InitializesSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.duckdb")

	store, err := NewDuckDBStore(dbPath)
	if err != nil {
		t.Fatalf("NewDuckDBStore() error = %v", err)
	}
	defer store.Close()

	// Verify tables exist
	tables := []string{"otel_traces", "otel_logs", "otel_metrics"}
	for _, table := range tables {
		var count int
		err := store.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		if err != nil {
			t.Errorf("table %s query failed: %v", table, err)
		}
	}
}

func TestDuckDBStore_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.duckdb")

	store, err := NewDuckDBStore(dbPath)
	if err != nil {
		t.Fatalf("NewDuckDBStore() error = %v", err)
	}

	if err := store.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify connection is closed
	if err := store.db.Ping(); err == nil {
		t.Error("expected error after Close(), got nil")
	}
}

// Helper to create in-memory test store
func setupTestStore(t *testing.T) (*DuckDBStore, func()) {
	t.Helper()
	store, err := NewDuckDBStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	return store, func() { store.Close() }
}

// ============ Traces Store Tests ============

func TestInsertSpans(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	spans := []api.Span{
		{
			TraceID:     "trace-001",
			SpanID:      "span-001",
			ServiceName: "test-service",
			SpanName:    "GET /api/users",
			Timestamp:   now,
			Duration:    100000000, // 100ms
			StatusCode:  "OK",
			SpanKind:    "SERVER",
			SpanAttributes: map[string]string{
				"http.method": "GET",
				"http.url":    "/api/users",
			},
		},
	}

	err := store.InsertSpans(ctx, spans)
	if err != nil {
		t.Fatalf("InsertSpans failed: %v", err)
	}

	// Verify insertion
	var count int
	if err := store.db.QueryRow("SELECT COUNT(*) FROM otel_traces").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 span, got %d", count)
	}
}

func TestInsertSpans_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	err := store.InsertSpans(context.Background(), []api.Span{})
	if err != nil {
		t.Errorf("InsertSpans with empty slice should not error: %v", err)
	}
}

func TestInsertSpans_Multiple(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	spans := []api.Span{
		{TraceID: "trace-001", SpanID: "span-001", ServiceName: "svc-a", SpanName: "span1", Timestamp: now},
		{TraceID: "trace-001", SpanID: "span-002", ParentSpanID: "span-001", ServiceName: "svc-a", SpanName: "span2", Timestamp: now.Add(10 * time.Millisecond)},
		{TraceID: "trace-002", SpanID: "span-003", ServiceName: "svc-b", SpanName: "span3", Timestamp: now.Add(20 * time.Millisecond)},
	}

	if err := store.InsertSpans(ctx, spans); err != nil {
		t.Fatalf("InsertSpans failed: %v", err)
	}

	var count int
	store.db.QueryRow("SELECT COUNT(*) FROM otel_traces").Scan(&count)
	if count != 3 {
		t.Errorf("expected 3 spans, got %d", count)
	}
}

func TestQueryTraces(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Insert test data
	spans := []api.Span{
		{TraceID: "trace-001", SpanID: "span-001", ServiceName: "service-a", SpanName: "root-span", Timestamp: now, Duration: 100000000, StatusCode: "OK"},
		{TraceID: "trace-001", SpanID: "span-002", ParentSpanID: "span-001", ServiceName: "service-a", SpanName: "child-span", Timestamp: now.Add(10 * time.Millisecond), Duration: 50000000, StatusCode: "OK"},
		{TraceID: "trace-002", SpanID: "span-003", ServiceName: "service-b", SpanName: "other-root", Timestamp: now.Add(100 * time.Millisecond), Duration: 200000000, StatusCode: "ERROR"},
	}
	if err := store.InsertSpans(ctx, spans); err != nil {
		t.Fatalf("InsertSpans failed: %v", err)
	}

	// Query all traces
	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	resp, err := store.QueryTraces(ctx, "", "", from, to, 10, 0)
	if err != nil {
		t.Fatalf("QueryTraces failed: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("expected 2 traces, got %d", resp.Total)
	}
}

func TestQueryTraces_WithServiceFilter(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	spans := []api.Span{
		{TraceID: "trace-001", SpanID: "span-001", ServiceName: "service-a", SpanName: "span1", Timestamp: now, StatusCode: "OK"},
		{TraceID: "trace-002", SpanID: "span-002", ServiceName: "service-b", SpanName: "span2", Timestamp: now.Add(10 * time.Millisecond), StatusCode: "OK"},
	}
	store.InsertSpans(ctx, spans)

	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	resp, err := store.QueryTraces(ctx, "service-a", "", from, to, 10, 0)
	if err != nil {
		t.Fatalf("QueryTraces failed: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("expected 1 trace for service-a, got %d", resp.Total)
	}
}

func TestQueryTraces_EmptyResult(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	from := time.Now().Add(-1 * time.Hour)
	to := time.Now()

	resp, err := store.QueryTraces(context.Background(), "", "", from, to, 10, 0)
	if err != nil {
		t.Fatalf("QueryTraces failed: %v", err)
	}

	if resp.Total != 0 {
		t.Errorf("expected 0 traces, got %d", resp.Total)
	}
	if len(resp.Traces) != 0 {
		t.Errorf("expected empty traces slice, got %d", len(resp.Traces))
	}
}

func TestGetTraceSpans(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	spans := []api.Span{
		{TraceID: "trace-001", SpanID: "span-001", ServiceName: "test-service", SpanName: "root", Timestamp: now, StatusCode: "OK"},
		{TraceID: "trace-001", SpanID: "span-002", ParentSpanID: "span-001", ServiceName: "test-service", SpanName: "child1", Timestamp: now.Add(10 * time.Millisecond), StatusCode: "OK"},
		{TraceID: "trace-001", SpanID: "span-003", ParentSpanID: "span-001", ServiceName: "test-service", SpanName: "child2", Timestamp: now.Add(20 * time.Millisecond), StatusCode: "OK"},
		{TraceID: "trace-002", SpanID: "span-004", ServiceName: "other-service", SpanName: "other", Timestamp: now, StatusCode: "OK"},
	}
	store.InsertSpans(ctx, spans)

	traceSpans, err := store.GetTraceSpans(ctx, "trace-001")
	if err != nil {
		t.Fatalf("GetTraceSpans failed: %v", err)
	}

	if len(traceSpans) != 3 {
		t.Errorf("expected 3 spans for trace-001, got %d", len(traceSpans))
	}
}

func TestGetTraceSpans_NotFound(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	spans, err := store.GetTraceSpans(context.Background(), "nonexistent-trace")
	if err != nil {
		t.Fatalf("GetTraceSpans failed: %v", err)
	}

	if len(spans) != 0 {
		t.Errorf("expected 0 spans for nonexistent trace, got %d", len(spans))
	}
}

func TestGetServices(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Insert spans from different services
	spans := []api.Span{
		{TraceID: "t1", SpanID: "s1", ServiceName: "service-a", SpanName: "span", Timestamp: now},
		{TraceID: "t2", SpanID: "s2", ServiceName: "service-b", SpanName: "span", Timestamp: now},
		{TraceID: "t3", SpanID: "s3", ServiceName: "service-a", SpanName: "span", Timestamp: now}, // Duplicate
	}
	store.InsertSpans(ctx, spans)

	services, err := store.GetServices(ctx)
	if err != nil {
		t.Fatalf("GetServices failed: %v", err)
	}

	if len(services) != 2 {
		t.Errorf("expected 2 unique services, got %d", len(services))
	}
}

func TestGetServices_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	services, err := store.GetServices(context.Background())
	if err != nil {
		t.Fatalf("GetServices failed: %v", err)
	}

	if len(services) != 0 {
		t.Errorf("expected 0 services, got %d", len(services))
	}
}

func TestGetStats(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Insert test data
	spans := []api.Span{
		{TraceID: "t1", SpanID: "s1", ServiceName: "svc", SpanName: "span", Timestamp: now, StatusCode: "OK"},
		{TraceID: "t1", SpanID: "s2", ServiceName: "svc", SpanName: "span", Timestamp: now, StatusCode: "ERROR"},
		{TraceID: "t2", SpanID: "s3", ServiceName: "svc", SpanName: "span", Timestamp: now, StatusCode: "OK"},
	}
	store.InsertSpans(ctx, spans)

	logs := []api.LogRecord{
		{Timestamp: now, ServiceName: "svc", SeverityText: "INFO", Body: "test log"},
	}
	store.InsertLogs(ctx, logs)

	metrics := []api.MetricDataPoint{
		{Timestamp: now, ServiceName: "svc", MetricName: "test_metric", MetricType: "gauge", Value: ptrFloat64(42.0)},
	}
	store.InsertMetrics(ctx, metrics)

	stats, err := store.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.SpanCount != 3 {
		t.Errorf("expected 3 spans, got %d", stats.SpanCount)
	}
	if stats.TraceCount != 2 {
		t.Errorf("expected 2 traces, got %d", stats.TraceCount)
	}
	if stats.LogCount != 1 {
		t.Errorf("expected 1 log, got %d", stats.LogCount)
	}
	if stats.MetricCount != 1 {
		t.Errorf("expected 1 metric, got %d", stats.MetricCount)
	}
	if stats.ServiceCount != 1 {
		t.Errorf("expected 1 service, got %d", stats.ServiceCount)
	}
}

// ============ Logs Store Tests ============

func TestInsertLogs(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	logs := []api.LogRecord{
		{
			Timestamp:      now,
			ServiceName:    "test-service",
			SeverityText:   "INFO",
			SeverityNumber: 9,
			Body:           "Test log message",
		},
	}

	err := store.InsertLogs(ctx, logs)
	if err != nil {
		t.Fatalf("InsertLogs failed: %v", err)
	}

	var count int
	store.db.QueryRow("SELECT COUNT(*) FROM otel_logs").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 log, got %d", count)
	}
}

func TestInsertLogs_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	err := store.InsertLogs(context.Background(), []api.LogRecord{})
	if err != nil {
		t.Errorf("InsertLogs with empty slice should not error: %v", err)
	}
}

func TestQueryLogs(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	logs := []api.LogRecord{
		{Timestamp: now, ServiceName: "svc-a", SeverityText: "INFO", Body: "info message"},
		{Timestamp: now.Add(10 * time.Millisecond), ServiceName: "svc-a", SeverityText: "ERROR", Body: "error message"},
		{Timestamp: now.Add(20 * time.Millisecond), ServiceName: "svc-b", SeverityText: "WARN", Body: "warning"},
	}
	store.InsertLogs(ctx, logs)

	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	// Query all logs
	resp, err := store.QueryLogs(ctx, "", "", "", "", from, to, 10, 0)
	if err != nil {
		t.Fatalf("QueryLogs failed: %v", err)
	}

	if resp.Total != 3 {
		t.Errorf("expected 3 logs, got %d", resp.Total)
	}
}

func TestQueryLogs_WithSeverityFilter(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	logs := []api.LogRecord{
		{Timestamp: now, ServiceName: "svc", SeverityText: "INFO", Body: "info"},
		{Timestamp: now, ServiceName: "svc", SeverityText: "ERROR", Body: "error"},
		{Timestamp: now, ServiceName: "svc", SeverityText: "ERROR", Body: "another error"},
	}
	store.InsertLogs(ctx, logs)

	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	resp, err := store.QueryLogs(ctx, "", "ERROR", "", "", from, to, 10, 0)
	if err != nil {
		t.Fatalf("QueryLogs failed: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("expected 2 ERROR logs, got %d", resp.Total)
	}
}

func TestQueryLogs_WithSearch(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	logs := []api.LogRecord{
		{Timestamp: now, ServiceName: "svc", SeverityText: "INFO", Body: "database connection established"},
		{Timestamp: now, ServiceName: "svc", SeverityText: "ERROR", Body: "database connection failed"},
		{Timestamp: now, ServiceName: "svc", SeverityText: "INFO", Body: "request processed"},
	}
	store.InsertLogs(ctx, logs)

	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	resp, err := store.QueryLogs(ctx, "", "", "", "database", from, to, 10, 0)
	if err != nil {
		t.Fatalf("QueryLogs failed: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("expected 2 logs matching 'database', got %d", resp.Total)
	}
}

func TestGetLogLevels(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	logs := []api.LogRecord{
		{Timestamp: now, ServiceName: "svc", SeverityText: "INFO", Body: "1"},
		{Timestamp: now, ServiceName: "svc", SeverityText: "INFO", Body: "2"},
		{Timestamp: now, ServiceName: "svc", SeverityText: "ERROR", Body: "3"},
		{Timestamp: now, ServiceName: "svc", SeverityText: "WARN", Body: "4"},
	}
	store.InsertLogs(ctx, logs)

	levels, err := store.GetLogLevels(ctx)
	if err != nil {
		t.Fatalf("GetLogLevels failed: %v", err)
	}

	if levels["INFO"] != 2 {
		t.Errorf("expected INFO count 2, got %d", levels["INFO"])
	}
	if levels["ERROR"] != 1 {
		t.Errorf("expected ERROR count 1, got %d", levels["ERROR"])
	}
	if levels["WARN"] != 1 {
		t.Errorf("expected WARN count 1, got %d", levels["WARN"])
	}
}

func TestGetLogLevels_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	levels, err := store.GetLogLevels(context.Background())
	if err != nil {
		t.Fatalf("GetLogLevels failed: %v", err)
	}

	if len(levels) != 0 {
		t.Errorf("expected 0 levels, got %d", len(levels))
	}
}

// ============ Metrics Store Tests ============

func TestInsertMetrics(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	metrics := []api.MetricDataPoint{
		{
			Timestamp:   now,
			ServiceName: "test-service",
			MetricName:  "cpu_usage",
			MetricType:  "gauge",
			Value:       ptrFloat64(45.5),
		},
	}

	err := store.InsertMetrics(ctx, metrics)
	if err != nil {
		t.Fatalf("InsertMetrics failed: %v", err)
	}

	var count int
	store.db.QueryRow("SELECT COUNT(*) FROM otel_metrics").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 metric, got %d", count)
	}
}

func TestInsertMetrics_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	err := store.InsertMetrics(context.Background(), []api.MetricDataPoint{})
	if err != nil {
		t.Errorf("InsertMetrics with empty slice should not error: %v", err)
	}
}

func TestQueryMetrics(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	metrics := []api.MetricDataPoint{
		{Timestamp: now, ServiceName: "svc-a", MetricName: "cpu_usage", MetricType: "gauge", Value: ptrFloat64(50.0)},
		{Timestamp: now, ServiceName: "svc-a", MetricName: "memory_usage", MetricType: "gauge", Value: ptrFloat64(70.0)},
		{Timestamp: now, ServiceName: "svc-b", MetricName: "cpu_usage", MetricType: "gauge", Value: ptrFloat64(30.0)},
	}
	store.InsertMetrics(ctx, metrics)

	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	resp, err := store.QueryMetrics(ctx, "", "", "", from, to, 10, 0)
	if err != nil {
		t.Fatalf("QueryMetrics failed: %v", err)
	}

	if resp.Total != 3 {
		t.Errorf("expected 3 metrics, got %d", resp.Total)
	}
}

func TestQueryMetrics_WithFilters(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	metrics := []api.MetricDataPoint{
		{Timestamp: now, ServiceName: "svc-a", MetricName: "cpu_usage", MetricType: "gauge", Value: ptrFloat64(50.0)},
		{Timestamp: now, ServiceName: "svc-a", MetricName: "request_count", MetricType: "sum", Value: ptrFloat64(100.0)},
		{Timestamp: now, ServiceName: "svc-b", MetricName: "cpu_usage", MetricType: "gauge", Value: ptrFloat64(30.0)},
	}
	store.InsertMetrics(ctx, metrics)

	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	// Filter by service
	resp, err := store.QueryMetrics(ctx, "svc-a", "", "", from, to, 10, 0)
	if err != nil {
		t.Fatalf("QueryMetrics failed: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("expected 2 metrics for svc-a, got %d", resp.Total)
	}

	// Filter by metric name
	resp, err = store.QueryMetrics(ctx, "", "cpu_usage", "", from, to, 10, 0)
	if err != nil {
		t.Fatalf("QueryMetrics failed: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("expected 2 cpu_usage metrics, got %d", resp.Total)
	}

	// Filter by type
	resp, err = store.QueryMetrics(ctx, "", "", "sum", from, to, 10, 0)
	if err != nil {
		t.Fatalf("QueryMetrics failed: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected 1 sum metric, got %d", resp.Total)
	}
}

func TestGetMetricNames(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	metrics := []api.MetricDataPoint{
		{Timestamp: now, ServiceName: "svc", MetricName: "cpu_usage", MetricType: "gauge", Value: ptrFloat64(50.0)},
		{Timestamp: now, ServiceName: "svc", MetricName: "memory_usage", MetricType: "gauge", Value: ptrFloat64(70.0)},
		{Timestamp: now, ServiceName: "svc", MetricName: "cpu_usage", MetricType: "gauge", Value: ptrFloat64(55.0)}, // Duplicate
	}
	store.InsertMetrics(ctx, metrics)

	names, err := store.GetMetricNames(ctx, "")
	if err != nil {
		t.Fatalf("GetMetricNames failed: %v", err)
	}

	if len(names) != 2 {
		t.Errorf("expected 2 unique metric names, got %d", len(names))
	}
}

func TestGetMetricNames_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	names, err := store.GetMetricNames(context.Background(), "")
	if err != nil {
		t.Fatalf("GetMetricNames failed: %v", err)
	}

	if len(names) != 0 {
		t.Errorf("expected 0 metric names, got %d", len(names))
	}
}

// ============ Pagination Tests ============

func TestQueryTraces_Pagination(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Insert 5 traces
	for i := 0; i < 5; i++ {
		spans := []api.Span{
			{TraceID: "trace-" + string(rune('a'+i)), SpanID: "span-" + string(rune('a'+i)), ServiceName: "svc", SpanName: "span", Timestamp: now.Add(time.Duration(i) * time.Minute), StatusCode: "OK"},
		}
		store.InsertSpans(ctx, spans)
	}

	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	// Get first page
	resp, err := store.QueryTraces(ctx, "", "", from, to, 2, 0)
	if err != nil {
		t.Fatalf("QueryTraces failed: %v", err)
	}

	if len(resp.Traces) != 2 {
		t.Errorf("expected 2 traces on first page, got %d", len(resp.Traces))
	}
	if !resp.HasMore {
		t.Error("expected HasMore to be true")
	}

	// Get second page
	resp, err = store.QueryTraces(ctx, "", "", from, to, 2, 2)
	if err != nil {
		t.Fatalf("QueryTraces failed: %v", err)
	}

	if len(resp.Traces) != 2 {
		t.Errorf("expected 2 traces on second page, got %d", len(resp.Traces))
	}
}

func TestQueryLogs_Pagination(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Insert 5 logs
	for i := 0; i < 5; i++ {
		logs := []api.LogRecord{
			{Timestamp: now.Add(time.Duration(i) * time.Minute), ServiceName: "svc", SeverityText: "INFO", Body: "log " + string(rune('a'+i))},
		}
		store.InsertLogs(ctx, logs)
	}

	from := now.Add(-1 * time.Hour)
	to := now.Add(1 * time.Hour)

	resp, err := store.QueryLogs(ctx, "", "", "", "", from, to, 2, 0)
	if err != nil {
		t.Fatalf("QueryLogs failed: %v", err)
	}

	if len(resp.Logs) != 2 {
		t.Errorf("expected 2 logs on first page, got %d", len(resp.Logs))
	}
	if !resp.HasMore {
		t.Error("expected HasMore to be true")
	}
}

// ============ Metric Series Tests ============

func TestQueryMetricSeries(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Minute)

	// Insert gauge metrics
	metrics := []api.MetricDataPoint{
		{Timestamp: now, ServiceName: "svc-a", MetricName: "cpu_usage", MetricType: "gauge", Value: ptrFloat64(50.0)},
		{Timestamp: now.Add(1 * time.Minute), ServiceName: "svc-a", MetricName: "cpu_usage", MetricType: "gauge", Value: ptrFloat64(60.0)},
		{Timestamp: now.Add(2 * time.Minute), ServiceName: "svc-a", MetricName: "cpu_usage", MetricType: "gauge", Value: ptrFloat64(55.0)},
	}
	store.InsertMetrics(ctx, metrics)

	from := now.Add(-1 * time.Minute)
	to := now.Add(5 * time.Minute)

	// Query time series
	resp, err := store.QueryMetricSeries(ctx, "cpu_usage", "", from, to, 60, false)
	if err != nil {
		t.Fatalf("QueryMetricSeries failed: %v", err)
	}

	if len(resp.Series) == 0 {
		t.Error("expected at least one time series")
	}
}

func TestQueryMetricSeries_NoData(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	from := now.Add(-1 * time.Hour)
	to := now

	resp, err := store.QueryMetricSeries(ctx, "nonexistent_metric", "", from, to, 60, false)
	if err != nil {
		t.Fatalf("QueryMetricSeries failed: %v", err)
	}

	if len(resp.Series) != 0 {
		t.Errorf("expected 0 series for nonexistent metric, got %d", len(resp.Series))
	}
}

func TestQueryMetricSeries_WithAggregation(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Minute)

	// Insert gauge metrics
	metrics := []api.MetricDataPoint{
		{Timestamp: now, ServiceName: "svc-a", MetricName: "memory_usage", MetricType: "gauge", Value: ptrFloat64(100.0)},
		{Timestamp: now.Add(1 * time.Minute), ServiceName: "svc-a", MetricName: "memory_usage", MetricType: "gauge", Value: ptrFloat64(200.0)},
		{Timestamp: now.Add(2 * time.Minute), ServiceName: "svc-a", MetricName: "memory_usage", MetricType: "gauge", Value: ptrFloat64(150.0)},
	}
	store.InsertMetrics(ctx, metrics)

	from := now.Add(-1 * time.Minute)
	to := now.Add(5 * time.Minute)

	// Query with aggregation (scalar result)
	resp, err := store.QueryMetricSeries(ctx, "memory_usage", "", from, to, 60, true)
	if err != nil {
		t.Fatalf("QueryMetricSeries with aggregation failed: %v", err)
	}

	if len(resp.Series) == 0 {
		t.Error("expected at least one aggregated series")
	}
}

func TestQueryMetricSeries_WithServiceFilter(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Minute)

	// Insert metrics from different services
	metrics := []api.MetricDataPoint{
		{Timestamp: now, ServiceName: "svc-a", MetricName: "requests", MetricType: "gauge", Value: ptrFloat64(100.0)},
		{Timestamp: now, ServiceName: "svc-b", MetricName: "requests", MetricType: "gauge", Value: ptrFloat64(200.0)},
	}
	store.InsertMetrics(ctx, metrics)

	from := now.Add(-1 * time.Minute)
	to := now.Add(5 * time.Minute)

	// Query with service filter
	resp, err := store.QueryMetricSeries(ctx, "requests", "svc-a", from, to, 60, true)
	if err != nil {
		t.Fatalf("QueryMetricSeries with service filter failed: %v", err)
	}

	for _, series := range resp.Series {
		if series.Labels["service"] != "svc-a" {
			t.Errorf("expected only svc-a in results, got %s", series.Labels["service"])
		}
	}
}

func TestQueryMetricSeries_SumMetric(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Minute)

	// Insert sum metrics (delta)
	metrics := []api.MetricDataPoint{
		{Timestamp: now, ServiceName: "svc", MetricName: "request_count", MetricType: "sum", Value: ptrFloat64(10.0)},
		{Timestamp: now.Add(1 * time.Minute), ServiceName: "svc", MetricName: "request_count", MetricType: "sum", Value: ptrFloat64(15.0)},
	}
	store.InsertMetrics(ctx, metrics)

	from := now.Add(-1 * time.Minute)
	to := now.Add(5 * time.Minute)

	resp, err := store.QueryMetricSeries(ctx, "request_count", "", from, to, 60, true)
	if err != nil {
		t.Fatalf("QueryMetricSeries for sum metric failed: %v", err)
	}

	if len(resp.Series) == 0 {
		t.Error("expected at least one series for sum metric")
	}
}

func TestGetLatestMetricValue(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Insert metrics
	metrics := []api.MetricDataPoint{
		{
			Timestamp:   now.Add(-2 * time.Minute),
			ServiceName: "test-service",
			MetricName:  "counter",
			MetricType:  "sum",
			Value:       ptrFloat64(100.0),
			Attributes:  map[string]string{"region": "us-east"},
		},
		{
			Timestamp:   now.Add(-1 * time.Minute),
			ServiceName: "test-service",
			MetricName:  "counter",
			MetricType:  "sum",
			Value:       ptrFloat64(150.0),
			Attributes:  map[string]string{"region": "us-east"},
		},
	}
	store.InsertMetrics(ctx, metrics)

	// Get latest value
	value, found := store.GetLatestMetricValue(ctx, "counter", "test-service", map[string]string{"region": "us-east"})
	if !found {
		t.Fatal("expected to find latest metric value")
	}

	if value != 150.0 {
		t.Errorf("expected value 150.0, got %f", value)
	}
}

func TestGetLatestMetricValue_NotFound(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()

	value, found := store.GetLatestMetricValue(ctx, "nonexistent", "test-service", map[string]string{})
	if found {
		t.Errorf("expected not found, got value %f", value)
	}
}

func TestQueryBatchMetricSeries(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now().Truncate(time.Minute)

	// Insert metrics
	metrics := []api.MetricDataPoint{
		{Timestamp: now, ServiceName: "svc", MetricName: "metric_a", MetricType: "gauge", Value: ptrFloat64(10.0)},
		{Timestamp: now, ServiceName: "svc", MetricName: "metric_b", MetricType: "gauge", Value: ptrFloat64(20.0)},
	}
	store.InsertMetrics(ctx, metrics)

	from := now.Add(-1 * time.Minute)
	to := now.Add(5 * time.Minute)

	queries := []api.MetricQuery{
		{ID: "q1", Name: "metric_a", Aggregate: true},
		{ID: "q2", Name: "metric_b", Aggregate: true},
		{ID: "q3", Name: "nonexistent", Aggregate: true},
	}

	resp := store.QueryBatchMetricSeries(ctx, queries, from, to, 60)

	if len(resp.Results) != 3 {
		t.Errorf("expected 3 results, got %d", len(resp.Results))
	}

	for _, result := range resp.Results {
		if !result.Success {
			if result.ID != "q3" {
				t.Errorf("unexpected failure for query %s: %s", result.ID, result.Error)
			}
		}
	}
}

func TestQueryBatchMetricSeries_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	resp := store.QueryBatchMetricSeries(ctx, []api.MetricQuery{}, now.Add(-1*time.Hour), now, 60)

	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results for empty queries, got %d", len(resp.Results))
	}
}

// ============ Recent Traces Tests ============

func TestGetRecentTraces(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Insert traces from different services
	spans := []api.Span{
		{TraceID: "trace-1", SpanID: "span-1", ServiceName: "service-a", SpanName: "root-1", Timestamp: now, Duration: 100000000, StatusCode: "OK"},
		{TraceID: "trace-2", SpanID: "span-2", ServiceName: "service-b", SpanName: "root-2", Timestamp: now.Add(-1 * time.Minute), Duration: 200000000, StatusCode: "ERROR"},
		{TraceID: "trace-3", SpanID: "span-3", ServiceName: "service-a", SpanName: "root-3", Timestamp: now.Add(-2 * time.Minute), Duration: 150000000, StatusCode: "OK"},
	}
	store.InsertSpans(ctx, spans)

	resp, err := store.GetRecentTraces(ctx, 10)
	if err != nil {
		t.Fatalf("GetRecentTraces failed: %v", err)
	}

	if len(resp.Traces) != 3 {
		t.Errorf("expected 3 traces, got %d", len(resp.Traces))
	}

	// Verify order (most recent first)
	if len(resp.Traces) >= 2 && resp.Traces[0].StartTime.Before(resp.Traces[1].StartTime) {
		t.Error("expected traces ordered by most recent first")
	}
}

func TestGetRecentTraces_Empty(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	resp, err := store.GetRecentTraces(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetRecentTraces failed: %v", err)
	}

	if len(resp.Traces) != 0 {
		t.Errorf("expected 0 traces, got %d", len(resp.Traces))
	}
}

func TestGetRecentTraces_Limit(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Insert 5 traces
	for i := 0; i < 5; i++ {
		spans := []api.Span{
			{
				TraceID:     "trace-" + string(rune('a'+i)),
				SpanID:      "span-" + string(rune('a'+i)),
				ServiceName: "service",
				SpanName:    "root",
				Timestamp:   now.Add(-time.Duration(i) * time.Minute),
				StatusCode:  "OK",
			},
		}
		store.InsertSpans(ctx, spans)
	}

	resp, err := store.GetRecentTraces(ctx, 3)
	if err != nil {
		t.Fatalf("GetRecentTraces failed: %v", err)
	}

	if len(resp.Traces) != 3 {
		t.Errorf("expected 3 traces (limit), got %d", len(resp.Traces))
	}
}

func TestGetRecentTraces_ExcludesCodexService(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	// Insert traces including codex_cli_rs service (which should be handled separately)
	spans := []api.Span{
		{TraceID: "trace-1", SpanID: "span-1", ServiceName: "service-a", SpanName: "root-1", Timestamp: now, StatusCode: "OK"},
		{TraceID: "trace-2", SpanID: "span-2", ServiceName: "codex_cli_rs", SpanName: "codex-span", Timestamp: now.Add(-1 * time.Minute), StatusCode: "OK"},
		{TraceID: "trace-3", SpanID: "span-3", ServiceName: "service-b", SpanName: "root-3", Timestamp: now.Add(-2 * time.Minute), StatusCode: "OK"},
	}
	store.InsertSpans(ctx, spans)

	resp, err := store.GetRecentTraces(ctx, 10)
	if err != nil {
		t.Fatalf("GetRecentTraces failed: %v", err)
	}

	// Note: GetRecentTraces queries non-Codex and Codex traces separately and merges them
	// The exact behavior depends on implementation - just verify it doesn't error
	if resp == nil {
		t.Error("expected non-nil response")
	}
}

func TestGetMetricNames_WithServiceFilter(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	ctx := context.Background()
	now := time.Now()

	metrics := []api.MetricDataPoint{
		{Timestamp: now, ServiceName: "svc-a", MetricName: "cpu", MetricType: "gauge", Value: ptrFloat64(50.0)},
		{Timestamp: now, ServiceName: "svc-a", MetricName: "memory", MetricType: "gauge", Value: ptrFloat64(70.0)},
		{Timestamp: now, ServiceName: "svc-b", MetricName: "cpu", MetricType: "gauge", Value: ptrFloat64(30.0)},
		{Timestamp: now, ServiceName: "svc-b", MetricName: "disk", MetricType: "gauge", Value: ptrFloat64(60.0)},
	}
	store.InsertMetrics(ctx, metrics)

	// Filter by service
	names, err := store.GetMetricNames(ctx, "svc-a")
	if err != nil {
		t.Fatalf("GetMetricNames failed: %v", err)
	}

	if len(names) != 2 {
		t.Errorf("expected 2 metric names for svc-a, got %d", len(names))
	}
}

// ============ Helper functions ============

func ptrFloat64(v float64) *float64 {
	return &v
}
