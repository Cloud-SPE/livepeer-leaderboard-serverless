package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/livepeer/leaderboard-serverless/metrics"
	"github.com/livepeer/leaderboard-serverless/models"
)

func TestGPUMetricsHandler(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/gpu/metrics?orchestrator_address=0x0abe02f6ef1fa8c29f9b3f9f170c6f3681fd3031&pipeline=streamdiffusion-sdxl-v2v&time_range=1h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GPUMetricsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Handler returned wrong status code: got %v want %v, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var payload map[string][]models.GPUMetric
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	metricsList := payload["metrics"]
	if len(metricsList) < 2 {
		t.Fatalf("Expected more than one metric, got %d", len(metricsList))
	}
	if metricsList[0].OrchestratorAddress != "0x0abe02f6ef1fa8c29f9b3f9f170c6f3681fd3031" {
		t.Fatalf("Expected orchestrator_address to match query, got %s", metricsList[0].OrchestratorAddress)
	}
	if metricsList[0].Pipeline != "streamdiffusion-sdxl-v2v" {
		t.Fatalf("Expected pipeline to match query, got %s", metricsList[0].Pipeline)
	}
	if metricsList[0].GPUName == nil {
		t.Fatalf("Expected gpu_name to be populated")
	}
	if metricsList[0].PromptToFirstFrameMs == nil {
		t.Fatalf("Expected prompt_to_first_frame_ms to be populated")
	}
	if metricsList[0].KnownSessions == 0 {
		t.Fatalf("Expected known_sessions to be non-zero")
	}
	if metricsList[0].StartupSuccessSessions == 0 {
		t.Fatalf("Expected startup_success_sessions to be non-zero")
	}
	if metricsList[0].ConfirmedSwappedSessions+metricsList[0].InferredOrchestratorChangeSessions == 0 {
		t.Fatalf("Expected swapped session breakdown counters to be populated")
	}
}

func TestGPUMetricsHandler_ValidationRejectsBadDuration(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/gpu/metrics?time_range=48h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GPUMetricsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for out-of-range time_range, got %v", rr.Code)
	}
}
