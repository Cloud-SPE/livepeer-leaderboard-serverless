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

// NetworkDemandHandler handles a request for aggregate network demand data.
func NetworkDemandHandler(w http.ResponseWriter, r *http.Request) {
	middleware.AddStandardHttpHeaders(w)

	interval, err := common.ParseDurationParam(r, "interval", 15*time.Minute)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	query := &models.NetworkDemandQuery{
		Gateway:  r.URL.Query().Get("gateway"),
		Region:   r.URL.Query().Get("region"),
		Workflow: r.URL.Query().Get("workflow"),
		Interval: interval,
	}

	common.Logger.Debug("NetworkDemandHandler query=%+v store=%T", query, metrics.Store)
	results, err := metrics.Store.NetworkDemand(query)
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
