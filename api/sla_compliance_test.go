package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/livepeer/leaderboard-serverless/metrics"
	"github.com/livepeer/leaderboard-serverless/models"
)

func TestSLAComplianceHandler(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/sla/compliance?orchestrator_id=orch-9&period=48h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SLAComplianceHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var payload models.SLACompliance
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if payload.Score < 0 || payload.Score > 100 {
		t.Fatalf("Expected score within 0-100, got %v", payload.Score)
	}
	if payload.OrchestratorID == "" {
		t.Fatalf("Expected orchestrator_id to be populated")
	}
}
