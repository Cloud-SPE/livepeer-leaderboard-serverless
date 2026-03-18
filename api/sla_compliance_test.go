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

	req, err := http.NewRequest("GET", "/sla/compliance?orchestrator_address=0x5263e0ce3a97b634d8828ce4337ad0f70b30b077&pipeline_id=streamdiffusion-sdxl&window=24h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SLAComplianceHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Handler returned wrong status code: got %v want %v, body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var payload struct {
		Compliance []models.SLAComplianceRow `json:"compliance"`
		Pagination models.Pagination         `json:"pagination"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	rows := payload.Compliance
	if len(rows) < 1 {
		t.Fatalf("Expected at least one compliance row, got %d", len(rows))
	}
	if rows[0].OrchestratorAddress != "0x5263e0ce3a97b634d8828ce4337ad0f70b30b077" {
		t.Fatalf("Expected orchestrator_address to match query, got %s", rows[0].OrchestratorAddress)
	}
	if rows[0].EffectiveSuccessRate == nil {
		t.Fatalf("Expected effective_success_rate to be populated")
	}
	if rows[0].StartupSuccessRate == nil {
		t.Fatalf("Expected startup_success_rate to be populated")
	}
	if rows[0].NoSwapRate == nil {
		t.Fatalf("Expected no_swap_rate to be populated")
	}
	if rows[0].SLAScore == nil {
		t.Fatalf("Expected sla_score to be populated")
	}
	if rows[0].StartupSuccessSessions == 0 {
		t.Fatalf("Expected startup_success_sessions to be non-zero")
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

func TestSLAComplianceHandler_ValidationRejectsBadPeriod(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/sla/compliance?window=5m", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SLAComplianceHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for window below 1h, got %v", rr.Code)
	}
}

func TestSLAComplianceHandler_WithOrg(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/sla/compliance?org=test-org-1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SLAComplianceHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for org filter, got %v", rr.Code)
	}

	var payload struct {
		Compliance []models.SLAComplianceRow `json:"compliance"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(payload.Compliance) == 0 {
		t.Fatalf("Expected at least one compliance row")
	}
	if payload.Compliance[0].Org == nil || *payload.Compliance[0].Org != "test-org-1" {
		t.Fatalf("Expected org field to be present and match query, got %+v", payload.Compliance[0].Org)
	}
}

func TestSLAComplianceHandler_PaginationRejectsInvalidPage(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/sla/compliance?page=0", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SLAComplianceHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for page=0, got %v", rr.Code)
	}
}

func TestSLAComplianceHandler_PaginationRejectsInvalidPageSize(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/sla/compliance?page_size=501", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SLAComplianceHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 for page_size=501, got %v", rr.Code)
	}
}

func TestSLAComplianceHandler_PaginationAcceptsMaxPageSize(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/sla/compliance?page_size=500", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SLAComplianceHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for page_size=500, got %v", rr.Code)
	}
}

func TestSLAComplianceHandler_Allows48hPeriod(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/sla/compliance?window=48h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SLAComplianceHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for window=48h, got %v", rr.Code)
	}
}

func TestSLAComplianceHandler_Allows72hPeriod(t *testing.T) {
	metrics.SetStore(metrics.NewMockStore())

	req, err := http.NewRequest("GET", "/sla/compliance?window=72h", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SLAComplianceHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for window=72h, got %v", rr.Code)
	}
}
