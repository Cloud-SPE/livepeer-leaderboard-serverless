package metrics

import (
	"github.com/livepeer/leaderboard-serverless/models"
)

type MetricsStore interface {
	GPUMetrics(query *models.GPUMetricsQuery) ([]*models.GPUMetric, error)
	NetworkDemand(query *models.NetworkDemandQuery) ([]*models.NetworkDemandRow, error)
	SLACompliance(query *models.SLAComplianceQuery) ([]*models.SLAComplianceRow, error)
	Datasets(query *models.DatasetsQuery) ([]*models.Dataset, error)
}

var Store MetricsStore

func SetStore(store MetricsStore) {
	if store == nil {
		return
	}
	Store = store
}
