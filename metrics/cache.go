package metrics

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/livepeer/leaderboard-serverless/common"
)

var cacheMu sync.Mutex

var requiredEnvVars = []string{
	"CLICKHOUSE_HOST",
	"CLICKHOUSE_PORT",
	"CLICKHOUSE_DB",
	"CLICKHOUSE_USER",
	"CLICKHOUSE_PASS",
}

// CacheCH lazily initialises the ClickHouse MetricsStore on the first call.
// Subsequent calls are no-ops if the store is already set.
// Returns an error when required env vars are missing or the connection fails.
func CacheCH() error {
	if Store != nil {
		return nil
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	// Double-check after acquiring the lock.
	if Store != nil {
		return nil
	}

	var missing []string
	for _, v := range requiredEnvVars {
		if strings.TrimSpace(os.Getenv(v)) == "" {
			missing = append(missing, v)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required ClickHouse env vars: %s", strings.Join(missing, ", "))
	}

	ch, err := NewClickhouseStoreFromEnv()
	if err != nil {
		return fmt.Errorf("clickhouse init failed: %w", err)
	}

	Store = ch
	common.Logger.Info("ClickHouse metrics store cached successfully")
	return nil
}
