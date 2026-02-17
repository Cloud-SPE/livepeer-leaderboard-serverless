package handler

import (
	"encoding/json"
	"net/http"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/metrics"
	"github.com/livepeer/leaderboard-serverless/middleware"
	"github.com/livepeer/leaderboard-serverless/models"
)

// DatasetsHandler handles a request for public load test datasets.
func DatasetsHandler(w http.ResponseWriter, r *http.Request) {
	middleware.AddStandardHttpHeaders(w)

	if !requireClickhouse(w) {
		return
	}

	workflow, err := validateOptionalString("workflow", r.URL.Query().Get("workflow"), 256)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}
	dsType, err := validateOptionalString("type", r.URL.Query().Get("type"), 256)
	if err != nil {
		common.HandleBadRequest(w, err)
		return
	}

	query := &models.DatasetsQuery{
		Workflow: workflow,
		Type:     dsType,
	}

	common.Logger.Debug("DatasetsHandler query=%+v store=%T", query, metrics.Store)
	results, err := metrics.Store.Datasets(query)
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}
	resultsEncoded, err := json.Marshal(map[string][]*models.Dataset{"datasets": results})
	if err != nil {
		common.HandleInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resultsEncoded)
}
