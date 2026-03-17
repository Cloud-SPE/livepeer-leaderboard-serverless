package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/livepeer/leaderboard-serverless/metrics"
)

func TestHealthHandler_BothUnavailable(t *testing.T) {
	// Ensure ClickHouse store is nil so CacheCH() fails on missing env vars.
	oldStore := metrics.Store
	metrics.Store = nil
	defer func() { metrics.Store = oldStore }()

	// Clear POSTGRES so the handler reports it as not configured.
	t.Setenv("POSTGRES", "")

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(HealthHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("Expected 503 when both components unavailable, got %v", rr.Code)
	}

	var resp struct {
		Postgres struct {
			OK    bool   `json:"ok"`
			Error string `json:"error"`
		} `json:"postgres"`
		Clickhouse struct {
			OK    bool   `json:"ok"`
			Error string `json:"error"`
		} `json:"clickhouse"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}
	if resp.Postgres.OK {
		t.Error("Expected postgres.ok=false when POSTGRES env var is empty")
	}
	if resp.Postgres.Error == "" {
		t.Error("Expected postgres.error to be non-empty")
	}
	if resp.Clickhouse.OK {
		t.Error("Expected clickhouse.ok=false when Store is nil and env vars absent")
	}
	if resp.Clickhouse.Error == "" {
		t.Error("Expected clickhouse.error to be non-empty")
	}
}

func TestHealthHandler_ClickhouseOKPostgresMissing(t *testing.T) {
	// Pre-set the ClickHouse store so CacheCH() returns immediately with success.
	oldStore := metrics.Store
	metrics.Store = metrics.NewMockStore()
	defer func() { metrics.Store = oldStore }()

	// Clear POSTGRES so Postgres is reported as not configured.
	t.Setenv("POSTGRES", "")

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(HealthHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("Expected 503 when Postgres is missing, got %v", rr.Code)
	}

	var resp struct {
		Postgres struct {
			OK bool `json:"ok"`
		} `json:"postgres"`
		Clickhouse struct {
			OK bool `json:"ok"`
		} `json:"clickhouse"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}
	if resp.Postgres.OK {
		t.Error("Expected postgres.ok=false when POSTGRES env var is empty")
	}
	if !resp.Clickhouse.OK {
		t.Error("Expected clickhouse.ok=true when MockStore is set")
	}
}
