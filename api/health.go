package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/middleware"
)

// requirePostgresEnv checks that Postgres is reachable.
// If db.Store is already initialised (e.g. tests), it passes immediately.
// Otherwise it checks for the POSTGRES env var.
// Returns true if ready; writes a 503 and returns false otherwise.
func requirePostgresEnv(w http.ResponseWriter) bool {
	if db.Store != nil {
		return true
	}
	if os.Getenv("POSTGRES") != "" {
		return true
	}
	common.RespondWithError(w, fmt.Errorf("postgres is not configured"), http.StatusServiceUnavailable)
	return false
}

type componentStatus struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type healthResponse struct {
	Postgres componentStatus `json:"postgres"`
}

// HealthHandler checks readiness of Postgres.
// Returns 200 if Postgres is OK, 503 otherwise.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	middleware.AddStandardHttpHeaders(w)

	resp := healthResponse{}

	if os.Getenv("POSTGRES") == "" {
		resp.Postgres = componentStatus{OK: false, Error: "POSTGRES env var not configured"}
	} else if err := db.CacheDB(); err != nil {
		resp.Postgres = componentStatus{OK: false, Error: "connection failed"}
	} else {
		resp.Postgres = componentStatus{OK: true}
	}

	encoded, _ := json.Marshal(resp)

	if resp.Postgres.OK {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	w.Write(encoded)
}
