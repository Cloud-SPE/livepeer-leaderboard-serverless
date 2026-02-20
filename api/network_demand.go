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

	if !requireClickhouse(w) {
		return
	}

	interval, err := common.ParseDurationParam(r, "interval", 15*time.Minute)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	if err := common.ValidateDuration("interval", interval, time.Minute, 24*time.Hour); err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	gateway, err := common.ValidateOptionalString("gateway", r.URL.Query().Get("gateway"), 256)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	region, err := common.ValidateOptionalString("region", r.URL.Query().Get("region"), 64)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	pipeline, err := common.ValidateOptionalString("pipeline", r.URL.Query().Get("pipeline"), 256)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	query := &models.NetworkDemandQuery{
		Gateway:  gateway,
		Region:   region,
		Pipeline: pipeline,
		Interval: interval,
	}

	common.Logger.Debug("NetworkDemandHandler query=%+v store=%T", query, metrics.Store)
	results, err := metrics.Store.NetworkDemand(query)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	resultsEncoded, err := json.Marshal(map[string][]*models.NetworkDemandRow{"demand": results})
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}
