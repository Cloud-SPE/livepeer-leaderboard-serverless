package metrics

import (
	"github.com/livepeer/leaderboard-serverless/models"
)

type MetricsStore interface {
	GPUMetrics(query *models.GPUMetricsQuery) ([]*models.GPUMetric, error)
	GPUMetricsCount(query *models.GPUMetricsQuery) (int, error)
	NetworkDemand(query *models.NetworkDemandQuery) ([]*models.NetworkDemandRow, error)
	NetworkDemandCount(query *models.NetworkDemandQuery) (int, error)
	SLACompliance(query *models.SLAComplianceQuery) ([]*models.SLAComplianceRow, error)
	SLAComplianceCount(query *models.SLAComplianceQuery) (int, error)
}

var Store MetricsStore

func SetStore(store MetricsStore) {
	if store == nil {
		return
	}
	Store = store
}
