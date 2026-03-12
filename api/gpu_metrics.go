package handler

import (
	"encoding/json"
	"net/http"
	"strings"
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
	gpuIDs, err := parseGPUIDFilters(r)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	// Preserve backwards compatibility for callers that only send gpu_id once.
	if gpuID == "" && len(gpuIDs) == 1 {
		gpuID = gpuIDs[0]
	}
	// Only enforce the strict time window when no GPU filter is provided.
	if gpuID == "" && len(gpuIDs) == 0 {
		if err := common.ValidateDuration("time_range", timeRange, time.Minute, 24*time.Hour); err != nil {
			common.HandleBadRequest(w, err)
			return
		}
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

	page, pageSize, err := common.ParsePageParams(r)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	query := &models.GPUMetricsQuery{
		OrchestratorAddress: orchAddr,
		GPUID:               gpuID,
		GPUIDs:              gpuIDs,
		Region:              region,
		PipelineID:          pipelineID,
		ModelID:             modelID,
		GPUModelName:        gpuModelName,
		RunnerVersion:       runnerVersion,
		CudaVersion:         cudaVersion,
		TimeRange:           timeRange,
		Pagination:          models.Pagination{Page: page, PageSize: pageSize},
	}

	common.Logger.Debug("GPUMetricsHandler query=%+v store=%T", query, metrics.Store)
	totalCount, err := metrics.Store.GPUMetricsCount(query)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	results, err := metrics.Store.GPUMetrics(query)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	query.Pagination.TotalCount = totalCount
	query.Pagination.TotalPages = (totalCount + query.Pagination.PageSize - 1) / query.Pagination.PageSize
	resultsEncoded, err := json.Marshal(map[string]interface{}{
		"metrics":    results,
		"pagination": query.Pagination,
	})
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}

func parseGPUIDFilters(r *http.Request) ([]string, error) {
	values := append([]string{}, r.URL.Query()["gpu_id"]...)
	values = append(values, r.URL.Query()["gpu_id[]"]...)

	ids := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		for _, candidate := range strings.Split(raw, ",") {
			trimmed, err := common.ValidateOptionalString("gpu_id", candidate, 256)
			if err != nil {
				return nil, err
			}
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			ids = append(ids, trimmed)
		}
	}
	return ids, nil
}
