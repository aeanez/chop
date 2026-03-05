package filters

import (
	"strings"
	"testing"
)

func TestFilterKubectlLogsRepeatedLines(t *testing.T) {
	var lines []string
	for i := 0; i < 20; i++ {
		lines = append(lines, "2026-03-05 10:00:00 INFO  Processing request for /api/health")
	}
	lines = append(lines, "2026-03-05 10:01:00 ERROR Connection refused to database")
	for i := 0; i < 10; i++ {
		lines = append(lines, "2026-03-05 10:02:00 INFO  Retrying connection...")
	}
	lines = append(lines, "2026-03-05 10:03:00 WARN  High memory usage detected")
	raw := strings.Join(lines, "\n")

	got, err := filterKubectlLogs(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should deduplicate repeated lines
	if !strings.Contains(got, "(x20)") {
		t.Error("expected repeated health check lines to show (x20)")
	}
	if !strings.Contains(got, "(x10)") {
		t.Error("expected repeated retry lines to show (x10)")
	}

	// Should keep ERROR and WARN lines
	if !strings.Contains(got, "ERROR") {
		t.Error("expected ERROR line to be preserved")
	}
	if !strings.Contains(got, "WARN") {
		t.Error("expected WARN line to be preserved")
	}

	// Token savings >= 50%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("Token savings: %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
}

func TestFilterKubectlLogsJSONStructured(t *testing.T) {
	// Generate realistic JSON logs with repeated entries that get deduped
	var lines []string
	lines = append(lines, `{"timestamp": "2026-03-05T10:00:00Z", "level": "info", "message": "Server started on port 8080", "port": 8080, "version": "v2.3.1", "host": "api-server-7d9b6c8f5-abc12"}`)
	// 10 repeated health check lines
	for i := 0; i < 10; i++ {
		lines = append(lines, `{"timestamp": "2026-03-05T10:00:01Z", "level": "info", "message": "Health check passed", "endpoint": "/healthz", "status": 200, "latency_ms": 1, "checker": "kubernetes"}`)
	}
	lines = append(lines, `{"timestamp": "2026-03-05T10:00:05Z", "level": "error", "message": "Failed to process incoming request", "request_id": "abc123-def456", "status": 500, "method": "POST", "path": "/api/v2/users", "stack": "goroutine 1 [running]: main.handler()"}`)
	lines = append(lines, `{"timestamp": "2026-03-05T10:00:06Z", "level": "warning", "message": "Rate limit threshold approaching for API gateway", "current_rps": 950, "limit": 1000, "client": "partner-service"}`)
	// 8 repeated request processed lines
	for i := 0; i < 8; i++ {
		lines = append(lines, `{"timestamp": "2026-03-05T10:00:07Z", "level": "info", "message": "Request processed successfully", "request_id": "def456", "status": 200, "duration_ms": 45, "method": "GET", "path": "/api/v2/users/123"}`)
	}
	lines = append(lines, `{"timestamp": "2026-03-05T10:00:09Z", "level": "debug", "message": "Garbage collection cycle completed", "heap_size": "128MB", "pause_ms": 2, "num_gc": 42}`)
	raw := strings.Join(lines, "\n")

	got, err := filterKubectlLogs(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should extract timestamp, level, message only
	if !strings.Contains(got, "2026-03-05T10:00:00Z") {
		t.Error("expected timestamp to be preserved")
	}

	// Should NOT contain extra JSON fields
	if strings.Contains(got, "latency_ms") {
		t.Error("expected extra JSON fields to be stripped")
	}
	if strings.Contains(got, "goroutine") {
		t.Error("expected stack trace JSON field to be stripped")
	}

	// Should preserve ERROR
	if !strings.Contains(got, "ERROR") || !strings.Contains(got, "Failed to process incoming request") {
		t.Error("expected ERROR entry to be preserved")
	}

	// Should preserve WARNING
	if !strings.Contains(got, "WARNING") || !strings.Contains(got, "Rate limit threshold approaching") {
		t.Error("expected WARNING entry to be preserved")
	}

	// Should deduplicate repeated entries
	if !strings.Contains(got, "(x10)") {
		t.Error("expected health check lines to be deduped with (x10)")
	}
	if !strings.Contains(got, "(x8)") {
		t.Error("expected request processed lines to be deduped with (x8)")
	}

	// Token savings >= 50%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("Token savings: %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
}

func TestFilterKubectlLogsMixedErrorDebug(t *testing.T) {
	// Generate > 100 lines to trigger debug stripping
	var lines []string
	for i := 0; i < 60; i++ {
		lines = append(lines, "2026-03-05 10:00:00 DEBUG  Received heartbeat ping")
	}
	for i := 0; i < 30; i++ {
		lines = append(lines, "2026-03-05 10:00:00 INFO  Processing batch job")
	}
	lines = append(lines, "2026-03-05 10:01:00 ERROR  OutOfMemoryError: heap space exhausted")
	lines = append(lines, "2026-03-05 10:01:01 FATAL  Application shutting down")
	for i := 0; i < 20; i++ {
		lines = append(lines, "2026-03-05 10:00:00 DEBUG  Cleanup in progress")
	}
	raw := strings.Join(lines, "\n")

	got, err := filterKubectlLogs(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ERROR and FATAL must always be kept
	if !strings.Contains(got, "OutOfMemoryError") {
		t.Error("expected ERROR line to be preserved")
	}
	if !strings.Contains(got, "FATAL") {
		t.Error("expected FATAL line to be preserved")
	}

	// DEBUG lines should be stripped (> 100 total lines)
	if strings.Contains(got, "heartbeat") {
		t.Error("expected DEBUG lines to be stripped when > 100 lines")
	}
	if strings.Contains(got, "Cleanup") {
		t.Error("expected DEBUG lines to be stripped when > 100 lines")
	}

	// Token savings >= 50%
	rawTokens := len(strings.Fields(raw))
	filteredTokens := len(strings.Fields(got))
	savings := 100.0 - float64(filteredTokens)/float64(rawTokens)*100.0
	if savings < 50.0 {
		t.Errorf("expected >=50%% token savings, got %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
	}
	t.Logf("Token savings: %.1f%% (raw=%d, filtered=%d)", savings, rawTokens, filteredTokens)
}

func TestFilterKubectlLogsEmpty(t *testing.T) {
	got, err := filterKubectlLogs("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got: %s", got)
	}
}
