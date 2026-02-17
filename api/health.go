package handler

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/metrics"
	"github.com/livepeer/leaderboard-serverless/middleware"
)

type componentStatus struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type healthResponse struct {
	Postgres   componentStatus `json:"postgres"`
	Clickhouse componentStatus `json:"clickhouse"`
}

// HealthHandler checks readiness of both Postgres and ClickHouse.
// Returns 200 if both are OK, 503 otherwise.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	middleware.AddStandardHttpHeaders(w)

	resp := healthResponse{}
	healthy := true

	// Check Postgres
	if os.Getenv("POSTGRES") == "" {
		resp.Postgres = componentStatus{OK: false, Error: "POSTGRES env var not configured"}
		healthy = false
	} else if err := db.CacheDB(); err != nil {
		resp.Postgres = componentStatus{OK: false, Error: "connection failed"}
		healthy = false
	} else {
		resp.Postgres = componentStatus{OK: true}
	}

	// Check ClickHouse
	if err := metrics.CacheCH(); err != nil {
		resp.Clickhouse = componentStatus{OK: false, Error: "connection failed"}
		healthy = false
	} else {
		resp.Clickhouse = componentStatus{OK: true}
	}

	encoded, _ := json.Marshal(resp)

	if healthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	w.Write(encoded)
}
