package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/livepeer/leaderboard-serverless/metrics"
	"github.com/livepeer/leaderboard-serverless/models"
)

func TestNetworkDemandHandler(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/network/demand?gateway=cloud-spe-ai-live-video-tester-mdw&pipeline_id=streamdiffusion-sdxl&model_id=streamdiffusion-sdxl&interval=15m", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NetworkDemandHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Handler returned wrong status code: got %v want %v, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var payload map[string][]models.NetworkDemandRow
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	rows := payload["demand"]
	if len(rows) < 2 {
		t.Fatalf("Expected more than one row, got %d", len(rows))
	}
	if rows[0].Gateway != "cloud-spe-ai-live-video-tester-mdw" {
		t.Fatalf("Expected gateway to match query, got %s", rows[0].Gateway)
	}
	if rows[0].PipelineID != "streamdiffusion-sdxl" {
		t.Fatalf("Expected pipeline_id to match query, got %s", rows[0].PipelineID)
	}
	if rows[0].ModelID == nil || *rows[0].ModelID != "streamdiffusion-sdxl" {
		t.Fatalf("Expected model_id to match query, got %+v", rows[0].ModelID)
	}
	if rows[0].SessionsCount == 0 {
		t.Fatalf("Expected sessions_count to be non-zero")
	}
	if rows[0].TotalMinutes == 0 {
		t.Fatalf("Expected total_minutes to be non-zero")
	}
	if rows[0].StartupSuccessRate == 0 {
		t.Fatalf("Expected startup_success_rate to be non-zero")
	}
	if rows[0].EffectiveSuccessRate == 0 {
		t.Fatalf("Expected effective_success_rate to be non-zero")
	}
	if rows[0].ConfirmedSwappedSessions+rows[0].InferredSwapSessions == 0 {
		t.Fatalf("Expected swapped session breakdown counters to be populated")
	}
	if rows[0].ErrorStatusSamples == 0 {
		t.Fatalf("Expected error_status_samples to be populated")
	}
	if rows[0].HealthSignalCoverageRatio == 0 {
		t.Fatalf("Expected health_signal_coverage_ratio to be populated")
	}
}

func TestNetworkDemandHandler_ValidationRejectsBadDuration(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/network/demand?interval=48h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NetworkDemandHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for out-of-range interval, got %v", rr.Code)
	}
}
