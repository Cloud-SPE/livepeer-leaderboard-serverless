package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/metrics"
	"github.com/livepeer/leaderboard-serverless/middleware"
	"github.com/livepeer/leaderboard-serverless/models"
)

// gpuMetricsPostBody is the JSON body for POST /api/gpu/metrics.
type gpuMetricsPostBody struct {
	GPUIDs []string `json:"gpu_ids"`
}

// GPUMetricsHandler handles a request for per-GPU realtime metrics.
// GET: supports comma-separated gpu_id query param.
// POST: accepts JSON body with gpu_ids array for long lists that exceed URL limits.
func GPUMetricsHandler(w http.ResponseWriter, r *http.Request) {
	middleware.AddStandardHttpHeaders(w)

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		common.RespondWithError(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
		return
	}

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

	var gpuID string
	var gpuIDs []string
	if r.Method == http.MethodPost {
		gpuIDs, err = parseGPUIDsFromPost(r)
		if err != nil {
			common.HandleBadRequest(w, err)
			return
		}
	} else {
		gpuIDs, err = parseGPUIDsFromQuery(r)
		if err != nil {
			common.HandleBadRequest(w, err)
			return
		}
	}
	if len(gpuIDs) == 1 {
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

// parseGPUIDsFromQuery parses comma-separated gpu_id from GET query params only.
func parseGPUIDsFromQuery(r *http.Request) ([]string, error) {
	raw := r.URL.Query().Get("gpu_id")
	if raw == "" {
		return nil, nil
	}
	return parseAndDedupeGPUIDs(strings.Split(raw, ","))
}

// parseGPUIDsFromPost parses gpu_ids array from POST JSON body.
func parseGPUIDsFromPost(r *http.Request) ([]string, error) {
	var body gpuMetricsPostBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	return parseAndDedupeGPUIDs(body.GPUIDs)
}

func parseAndDedupeGPUIDs(candidates []string) ([]string, error) {
	ids := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, c := range candidates {
		trimmed, err := common.ValidateOptionalString("gpu_id", c, 256)
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
	return ids, nil
}
