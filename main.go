package main

import (
	"net/http"

	handler "github.com/livepeer/leaderboard-serverless/api"
	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/metrics"
)

// this func is for running in local mode.  Vercel does not use this as an entrypoint
// so any logic here should only reflect what is needed for local development
func main() {

	// Explicit ClickHouse init for local mode.
	// Vercel handlers use metrics.CacheCH() lazily per-request instead.
	ch, err := metrics.NewClickhouseStoreFromEnv()
	if err != nil {
		common.Logger.Fatal("Failed to initialise ClickHouse metrics store: %v", err)
	}
	metrics.SetStore(ch)

	http.HandleFunc("/api/health", handler.HealthHandler)
	http.HandleFunc("/api/raw_stats", handler.RawStatsHandler)
	http.HandleFunc("/api/aggregated_stats", handler.AggregatedStatsHandler)
	http.HandleFunc("/api/top_ai_score", handler.TopAiScoreHandler)
	http.HandleFunc("/api/post_stats", handler.PostStatsHandler)
	http.HandleFunc("/api/pipelines", handler.PipelinesHandler)
	http.HandleFunc("/api/regions", handler.RegionsHandler)
	http.HandleFunc("/api/gpu/metrics", handler.GPUMetricsHandler)
	http.HandleFunc("/api/network/demand", handler.NetworkDemandHandler)
	http.HandleFunc("/api/sla/compliance", handler.SLAComplianceHandler)
	http.HandleFunc("/api/datasets", handler.DatasetsHandler)

	common.Logger.Info("Server starting on port 8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		common.Logger.Fatal("Unable to start the server: %v", err)
	}
}
