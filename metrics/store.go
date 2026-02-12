package metrics

import "github.com/livepeer/leaderboard-serverless/models"

type MetricsStore interface {
	GPUMetrics(query *models.GPUMetricsQuery) ([]*models.GPUMetric, error)
	NetworkDemand(query *models.NetworkDemandQuery) (*models.NetworkDemand, error)
	SLACompliance(query *models.SLAComplianceQuery) (*models.SLACompliance, error)
	Datasets(query *models.DatasetsQuery) ([]*models.Dataset, error)
}

var Store MetricsStore = NewMockStore()

func SetStore(store MetricsStore) {
	if store == nil {
		return
	}
	Store = store
}
