package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/models"
)

// mockDB satisfies interfaces.DB with no-op implementations for health check tests.
type mockDB struct{}

func (m *mockDB) InsertStats(_ *models.Stats) error { return nil }
func (m *mockDB) AggregatedStats(_ *models.StatsQuery) (*models.AggregatedStatsResults, error) {
	return &models.AggregatedStatsResults{}, nil
}
func (m *mockDB) MedianRTT(_ *models.StatsQuery) (float64, error)        { return 0, nil }
func (m *mockDB) BestAIRegion(_ string) (*models.Stats, error)           { return nil, nil }
func (m *mockDB) RawStats(_ *models.StatsQuery) ([]*models.Stats, error) { return nil, nil }
func (m *mockDB) Regions() ([]*models.Region, error)                     { return nil, nil }
func (m *mockDB) InsertRegions(_ []*models.Region) (int, int)            { return 0, 0 }
func (m *mockDB) Pipelines(_ *models.StatsQuery) ([]*models.Pipeline, error) {
	return nil, nil
}
func (m *mockDB) Close() {}

func TestHealthHandler_PostgresUnavailable(t *testing.T) {
	t.Setenv("POSTGRES", "")

	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(HealthHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("Expected 503, got %v", rr.Code)
	}

	var resp healthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if resp.Postgres.OK {
		t.Error("Expected postgres.ok=false when POSTGRES env var is empty")
	}
	if resp.Postgres.Error == "" {
		t.Error("Expected postgres.error to be non-empty")
	}
}

func TestRequirePostgresEnv_StoreAlreadySet(t *testing.T) {
	oldStore := db.Store
	db.Store = &mockDB{}
	defer func() { db.Store = oldStore }()

	rr := httptest.NewRecorder()
	if !requirePostgresEnv(rr) {
		t.Error("expected requirePostgresEnv to return true when db.Store is set")
	}
}

func TestRequirePostgresEnv_PostgresEnvSet(t *testing.T) {
	oldStore := db.Store
	db.Store = nil
	defer func() { db.Store = oldStore }()

	t.Setenv("POSTGRES", "postgres://localhost/test")

	rr := httptest.NewRecorder()
	if !requirePostgresEnv(rr) {
		t.Error("expected requirePostgresEnv to return true when POSTGRES env var is set")
	}
}

func TestRequirePostgresEnv_NeitherSet(t *testing.T) {
	oldStore := db.Store
	db.Store = nil
	defer func() { db.Store = oldStore }()

	t.Setenv("POSTGRES", "")

	rr := httptest.NewRecorder()
	if requirePostgresEnv(rr) {
		t.Error("expected requirePostgresEnv to return false when neither db.Store nor POSTGRES is set")
	}
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestHealthHandler_PostgresOK(t *testing.T) {
	oldStore := db.Store
	db.Store = &mockDB{}
	defer func() { db.Store = oldStore }()

	t.Setenv("POSTGRES", "postgres://localhost/test")

	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(HealthHandler).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %v", rr.Code)
	}

	var resp healthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if !resp.Postgres.OK {
		t.Errorf("Expected postgres.ok=true, got error: %s", resp.Postgres.Error)
	}
}
