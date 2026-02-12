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

// SLAComplianceHandler handles a request for SLA compliance scores.
func SLAComplianceHandler(w http.ResponseWriter, r *http.Request) {
	middleware.AddStandardHttpHeaders(w)

	period, err := common.ParseDurationParam(r, "period", 24*time.Hour)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	query := &models.SLAComplianceQuery{
		OrchestratorID: r.URL.Query().Get("orchestrator_id"),
		Period:         period,
	}

	results, err := metrics.Store.SLACompliance(query)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	resultsEncoded, err := json.Marshal(results)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}
