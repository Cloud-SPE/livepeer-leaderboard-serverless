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

	if !requireClickhouse(w) {
		return
	}

	window, err := common.ParseDurationParam(r, "window", 24*time.Hour)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	if err := common.ValidateDuration("window", window, time.Hour, 30*24*time.Hour); err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	orchAddr, err := common.ValidateOptionalString("orchestrator_address", r.URL.Query().Get("orchestrator_address"), 256)
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
	gpuID, err := common.ValidateOptionalString("gpu_id", r.URL.Query().Get("gpu_id"), 256)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	org, err := common.ValidateOptionalString("org", r.URL.Query().Get("org"), 256)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	page, pageSize, err := common.ParsePageParams(r)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	query := &models.SLAComplianceQuery{
		Org:                 org,
		OrchestratorAddress: orchAddr,
		Region:              region,
		PipelineID:          pipelineID,
		ModelID:             modelID,
		GPUID:               gpuID,
		Window:              window,
		Pagination:          models.Pagination{Page: page, PageSize: pageSize},
	}

	common.Logger.Debug("SLAComplianceHandler query=%+v store=%T", query, metrics.Store)
	totalCount, err := metrics.Store.SLAComplianceCount(query)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	results, err := metrics.Store.SLACompliance(query)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	query.Pagination.TotalCount = totalCount
	query.Pagination.TotalPages = (totalCount + query.Pagination.PageSize - 1) / query.Pagination.PageSize
	resultsEncoded, err := json.Marshal(map[string]interface{}{
		"compliance": results,
		"pagination": query.Pagination,
	})
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}
