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
	if err := common.ValidateDuration("interval", interval, time.Minute, 48*time.Hour); err != nil {
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
	pipelineID, err := common.ValidateOptionalString("pipeline_id", r.URL.Query().Get("pipeline_id"), 256)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	modelID, err := common.ValidateOptionalString("model_id", r.URL.Query().Get("model_id"), 256)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	page, pageSize, err := common.ParsePageParams(r)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	query := &models.NetworkDemandQuery{
		Gateway:    gateway,
		Region:     region,
		PipelineID: pipelineID,
		ModelID:    modelID,
		Interval:   interval,
		Pagination: models.Pagination{Page: page, PageSize: pageSize},
	}

	common.Logger.Debug("NetworkDemandHandler query=%+v store=%T", query, metrics.Store)
	totalCount, err := metrics.Store.NetworkDemandCount(query)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	results, err := metrics.Store.NetworkDemand(query)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	query.Pagination.TotalCount = totalCount
	query.Pagination.TotalPages = (totalCount + query.Pagination.PageSize - 1) / query.Pagination.PageSize
	resultsEncoded, err := json.Marshal(map[string]interface{}{
		"demand":     results,
		"pagination": query.Pagination,
	})
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}
