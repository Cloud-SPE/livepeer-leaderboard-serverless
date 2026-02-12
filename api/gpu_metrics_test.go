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

	req, err := http.NewRequest("GET", "/gpu/metrics?gpu_id=gpu-9&workflow=inference&region=us-west", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GPUMetricsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	var payload map[string][]models.GPUMetric
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	metricsList := payload["metrics"]
	if len(metricsList) < 2 {
		t.Fatalf("Expected more than one metric, got %d", len(metricsList))
	}
	if metricsList[0].GPUId == "" {
		t.Fatalf("Expected gpu_id to be populated")
	}
}
