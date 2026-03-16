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

	req, err := http.NewRequest("GET", "/gpu/metrics?orchestrator_address=0x0abe02f6ef1fa8c29f9b3f9f170c6f3681fd3031&pipeline_id=streamdiffusion-sdxl-v2v&time_range=1h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GPUMetricsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Handler returned wrong status code: got %v want %v, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var payload struct {
		Metrics    []models.GPUMetric `json:"metrics"`
		Pagination models.Pagination  `json:"pagination"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	metricsList := payload.Metrics
	if len(metricsList) < 2 {
		t.Fatalf("Expected more than one metric, got %d", len(metricsList))
	}
	if metricsList[0].OrchestratorAddress != "0x0abe02f6ef1fa8c29f9b3f9f170c6f3681fd3031" {
		t.Fatalf("Expected orchestrator_address to match query, got %s", metricsList[0].OrchestratorAddress)
	}
	if metricsList[0].PipelineID != "streamdiffusion-sdxl-v2v" {
		t.Fatalf("Expected pipeline_id to match query, got %s", metricsList[0].PipelineID)
	}
	if metricsList[0].GPUModelName == nil {
		t.Fatalf("Expected gpu_model_name to be populated")
	}
	if metricsList[0].AvgPromptToFirstFrameMs == nil {
		t.Fatalf("Expected avg_prompt_to_first_frame_ms to be populated")
	}
	if metricsList[0].KnownSessionsCount == 0 {
		t.Fatalf("Expected known_sessions_count to be non-zero")
	}
	if metricsList[0].StartupSuccessSessions == 0 {
		t.Fatalf("Expected startup_success_sessions to be non-zero")
	}
	if metricsList[0].ConfirmedSwappedSessions+metricsList[0].InferredSwapSessions == 0 {
		t.Fatalf("Expected swapped session breakdown counters to be populated")
	}
	if metricsList[0].ErrorStatusSamples == 0 {
		t.Fatalf("Expected error_status_samples to be populated")
	}
	if metricsList[0].HealthSignalCoverageRatio == 0 {
		t.Fatalf("Expected health_signal_coverage_ratio to be populated")
	}

	// Pagination metadata assertions
	if payload.Pagination.Page != 1 {
		t.Fatalf("Expected default page 1, got %d", payload.Pagination.Page)
	}
	if payload.Pagination.PageSize != 50 {
		t.Fatalf("Expected default page_size 50, got %d", payload.Pagination.PageSize)
	}
	if payload.Pagination.TotalCount < 1 {
		t.Fatalf("Expected total_count >= 1, got %d", payload.Pagination.TotalCount)
	}
	if payload.Pagination.TotalPages < 1 {
		t.Fatalf("Expected total_pages >= 1, got %d", payload.Pagination.TotalPages)
	}
}

func TestGPUMetricsHandler_ValidationRejectsBadDuration(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/gpu/metrics?time_range=96h", nil)
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

func TestGPUMetricsHandler_AllowsExtendedDuration(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/gpu/metrics?time_range=72h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GPUMetricsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for time_range=72h, got %v", rr.Code)
	}
}

func TestGPUMetricsHandler_WithOrg(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/gpu/metrics?org=test-org-1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GPUMetricsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for org filter, got %v", rr.Code)
	}

	var payload struct {
		Metrics []models.GPUMetric `json:"metrics"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(payload.Metrics) == 0 {
		t.Fatalf("Expected at least one metric row")
	}
	if payload.Metrics[0].Org == nil || *payload.Metrics[0].Org != "test-org-1" {
		t.Fatalf("Expected org field to be present and match query, got %+v", payload.Metrics[0].Org)
	}
}

func TestGPUMetricsHandler_PaginationRejectsInvalidPage(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/gpu/metrics?page=0", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GPUMetricsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for page=0, got %v", rr.Code)
	}
}

func TestGPUMetricsHandler_PaginationRejectsInvalidPageSize(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/gpu/metrics?page_size=501", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GPUMetricsHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for page_size=501, got %v", rr.Code)
	}
}
