package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/metrics"
	"github.com/livepeer/leaderboard-serverless/middleware"
	"github.com/livepeer/leaderboard-serverless/models"
)

// GPUMetricsHandler handles a request for per-GPU realtime metrics.
func GPUMetricsHandler(w http.ResponseWriter, r *http.Request) {
	middleware.AddStandardHttpHeaders(w)

	timeRange, err := common.ParseDurationParam(r, "time_range", time.Hour)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	query := &models.GPUMetricsQuery{
		OrchestratorWallet: r.URL.Query().Get("o_wallet"),
		GPUId:              r.URL.Query().Get("gpu_id"),
		Region:             r.URL.Query().Get("region"),
		Workflow:           r.URL.Query().Get("workflow"),
		TimeRange:          timeRange,
	}

	results, err := metrics.Store.GPUMetrics(query)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	resultsEncoded, err := json.Marshal(map[string][]*models.GPUMetric{"metrics": results})
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}
