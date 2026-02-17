package handler

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/db"
	"github.com/livepeer/leaderboard-serverless/metrics"
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

// requireClickhouse ensures the ClickHouse metrics store is initialised.
// Returns true if ready; writes a 503 and returns false otherwise.
func requireClickhouse(w http.ResponseWriter) bool {
	if err := metrics.CacheCH(); err != nil {
		common.RespondWithError(w, fmt.Errorf("clickhouse is not available"), http.StatusServiceUnavailable)
		return false
	}
	return true
}

// validateDuration checks that a parsed duration is within [min, max].
// Returns an error suitable for a 400 response if out of bounds.
func validateDuration(name string, d time.Duration, min, max time.Duration) error {
	if d < min || d > max {
		return fmt.Errorf("%s must be between %s and %s, got %s", name, min, max, d)
	}
	return nil
}

// validateOptionalString trims and length-checks a query parameter.
// Returns the trimmed value and an error if it exceeds maxLen.
func validateOptionalString(name, value string, maxLen int) (string, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > maxLen {
		return "", fmt.Errorf("%s exceeds maximum length of %d characters", name, maxLen)
	}
	return trimmed, nil
}
