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

	query := &models.DatasetsQuery{
		Workflow: r.URL.Query().Get("workflow"),
		Type:     r.URL.Query().Get("type"),
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
