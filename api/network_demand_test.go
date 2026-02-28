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

	req, err := http.NewRequest("GET", "/network/demand?gateway=cloud-spe-ai-live-video-tester-mdw&pipeline=streamdiffusion-sdxl&model_id=streamdiffusion-sdxl&interval=15m", nil)
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
	if rows[0].Pipeline != "streamdiffusion-sdxl" {
		t.Fatalf("Expected pipeline to match query, got %s", rows[0].Pipeline)
	}
	if rows[0].ModelID == nil || *rows[0].ModelID != "streamdiffusion-sdxl" {
		t.Fatalf("Expected model_id to match query, got %+v", rows[0].ModelID)
	}
	if rows[0].TotalSessions == 0 {
		t.Fatalf("Expected total_sessions to be non-zero")
	}
	if rows[0].TotalStreams == 0 {
		t.Fatalf("Expected total_streams to be non-zero")
	}
	if rows[0].SuccessRatio == 0 {
		t.Fatalf("Expected success_ratio to be non-zero")
	}
	if rows[0].ConfirmedSwappedSessions+rows[0].InferredOrchestratorChangeSessions == 0 {
		t.Fatalf("Expected swapped session breakdown counters to be populated")
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
