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

	req, err := http.NewRequest("GET", "/network/demand?gateway=cloud-spe-ai-live-video-tester-mdw&pipeline_id=streamdiffusion-sdxl&model_id=streamdiffusion-sdxl&window=3h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NetworkDemandHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Handler returned wrong status code: got %v want %v, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var payload struct {
		Demand     []models.NetworkDemandRow `json:"demand"`
		Pagination models.Pagination         `json:"pagination"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	rows := payload.Demand
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

func TestNetworkDemandHandler_ValidationRejectsBadDuration(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/network/demand?window=721h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NetworkDemandHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for out-of-range window, got %v", rr.Code)
	}
}

func TestNetworkDemandHandler_WithOrg(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/network/demand?org=test-org-1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NetworkDemandHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for org filter, got %v", rr.Code)
	}

	var payload struct {
		Demand []models.NetworkDemandRow `json:"demand"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(payload.Demand) == 0 {
		t.Fatalf("Expected at least one demand row")
	}
	if payload.Demand[0].Org == nil || *payload.Demand[0].Org != "test-org-1" {
		t.Fatalf("Expected org field to be present and match query, got %+v", payload.Demand[0].Org)
	}
}

func TestNetworkDemandHandler_PaginationRejectsInvalidPage(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/network/demand?page=0", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NetworkDemandHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for page=0, got %v", rr.Code)
	}
}

func TestNetworkDemandHandler_PaginationRejectsInvalidPageSize(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/network/demand?page_size=501", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NetworkDemandHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for page_size=501, got %v", rr.Code)
	}
}

func TestNetworkDemandHandler_PaginationAcceptsMaxPageSize(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/network/demand?page_size=500", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NetworkDemandHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for page_size=500, got %v", rr.Code)
	}
}

func TestNetworkDemandHandler_AllowsMaxWindow(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/network/demand?window=48h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NetworkDemandHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for window=48h, got %v", rr.Code)
	}
}

func TestNetworkDemandHandler_FiltersByModelID(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/network/demand?model_id=my-custom-model", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NetworkDemandHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for model_id filter, got %v", rr.Code)
	}

	var payload struct {
		Demand []models.NetworkDemandRow `json:"demand"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(payload.Demand) == 0 {
		t.Fatalf("Expected at least one demand row")
	}
	if payload.Demand[0].ModelID == nil || *payload.Demand[0].ModelID != "my-custom-model" {
		t.Fatalf("Expected model_id to be filtered to 'my-custom-model', got %+v", payload.Demand[0].ModelID)
	}
}

// TestNetworkDemandHandler_WindowIsDirectLookback documents the fix for the former
// interval×12 multiplier bug. window=1h must return data for the last 1 hour, not 12 hours.
func TestNetworkDemandHandler_WindowIsDirectLookback(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/network/demand?window=1h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NetworkDemandHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for window=1h, got %v: %s", rr.Code, rr.Body.String())
	}

	var payload struct {
		Demand []models.NetworkDemandRow `json:"demand"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if len(payload.Demand) == 0 {
		t.Fatalf("Expected demand rows for window=1h")
	}
}

func TestNetworkDemandHandler_UnknownParamIgnored(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/network/demand?window=3h&unknown_param=foo", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(NetworkDemandHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 when unknown query params are present, got %v", rr.Code)
	}
}
