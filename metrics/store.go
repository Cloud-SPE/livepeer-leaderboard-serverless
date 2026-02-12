package metrics

import (
	"os"
	"strings"

	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/models"
)

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

func init() {
	storeType := strings.ToLower(os.Getenv("METRICS_STORE"))
	if storeType != "clickhouse" {
		return
	}

	clickhouseStore, err := NewClickhouseStoreFromEnv()
	if err != nil {
		common.Logger.Warn("Failed to start ClickHouse metrics store, using mock: %v", err)
		return
	}
	common.Logger.Info("Metrics store set to ClickHouse")
	Store = clickhouseStore
}
