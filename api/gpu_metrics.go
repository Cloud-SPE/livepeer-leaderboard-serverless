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

	if !requireClickhouse(w) {
		return
	}

	timeRange, err := common.ParseDurationParam(r, "time_range", time.Hour)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	if err := common.ValidateDuration("time_range", timeRange, time.Minute, 24*time.Hour); err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	orchAddr, err := common.ValidateOptionalString("orchestrator_address", r.URL.Query().Get("orchestrator_address"), 256)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	gpuID, err := common.ValidateOptionalString("gpu_id", r.URL.Query().Get("gpu_id"), 256)
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
	gpuModelName, err := common.ValidateOptionalString("gpu_model_name", r.URL.Query().Get("gpu_model_name"), 256)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	runnerVersion, err := common.ValidateOptionalString("runner_version", r.URL.Query().Get("runner_version"), 64)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	cudaVersion, err := common.ValidateOptionalString("cuda_version", r.URL.Query().Get("cuda_version"), 32)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	query := &models.GPUMetricsQuery{
		OrchestratorAddress: orchAddr,
		GPUID:               gpuID,
		Region:              region,
		PipelineID:          pipelineID,
		ModelID:             modelID,
		GPUModelName:        gpuModelName,
		RunnerVersion:       runnerVersion,
		CudaVersion:         cudaVersion,
		TimeRange:           timeRange,
	}

	common.Logger.Debug("GPUMetricsHandler query=%+v store=%T", query, metrics.Store)
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
