package metrics

import (
	"fmt"
	"time"

	"github.com/livepeer/leaderboard-serverless/models"
)

type MockStore struct{}

func NewMockStore() *MockStore {
	return &MockStore{}
}

func (m *MockStore) GPUMetrics(query *models.GPUMetricsQuery) ([]*models.GPUMetric, error) {
	now := time.Now().UTC()
	gpuID := query.GPUId
	if gpuID == "" {
		gpuID = "gpu-0"
	}
	workflow := query.Workflow
	if workflow == "" {
		workflow = "inference"
	}
	region := query.Region
	if region == "" {
		region = "us-west"
	}
	wallet := query.OrchestratorWallet
	if wallet == "" {
		wallet = "0x0000000000000000000000000000000000000000"
	}
	metrics := make([]*models.GPUMetric, 0, 6)
	for i := 0; i < 6; i++ {
		metrics = append(metrics, &models.GPUMetric{
			OrchestratorWallet: wallet,
			GPUId:              gpuID,
			Region:             region,
			Workflow:           workflow,
			Timestamp:          now.Add(-time.Duration(i) * time.Minute),
			UtilizationPct:     55.0 + float64(i)*3.2,
			MemoryUsedMB:       11000 + float64(i)*420,
			MemoryTotalMB:      24576,
			TemperatureC:       68.0 + float64(i)*0.9,
			PowerWatts:         190.0 + float64(i)*4.1,
			SuccessRate:        0.995 - float64(i)*0.001,
			ErrorRate:          0.005 + float64(i)*0.001,
		})
	}
	return metrics, nil
}

func (m *MockStore) NetworkDemand(query *models.NetworkDemandQuery) (*models.NetworkDemand, error) {
	now := time.Now().UTC()
	interval := query.Interval
	points := make([]models.NetworkDemandPoint, 0, 12)
	for i := 11; i >= 0; i-- {
		pointTime := now.Add(-time.Duration(i) * interval)
		points = append(points, models.NetworkDemandPoint{
			Timestamp:     pointTime,
			StreamMinutes: 120 + float64(i)*6,
			InferMinutes:  60 + float64(i)*3,
		})
	}
	return &models.NetworkDemand{
		Gateway:  defaultString(query.Gateway, "public"),
		Region:   defaultString(query.Region, "global"),
		Workflow: defaultString(query.Workflow, "inference"),
		Interval: interval.String(),
		Points:   points,
	}, nil
}

func (m *MockStore) SLACompliance(query *models.SLAComplianceQuery) (*models.SLACompliance, error) {
	now := time.Now().UTC()
	start := now.Add(-query.Period)
	return &models.SLACompliance{
		OrchestratorID: defaultString(query.OrchestratorID, "orch-0"),
		Period:         query.Period.String(),
		Score:          97.5,
		WindowStart:    start,
		WindowEnd:      now,
	}, nil
}

func (m *MockStore) Datasets(query *models.DatasetsQuery) ([]*models.Dataset, error) {
	now := time.Now().UTC()
	datasets := []*models.Dataset{
		{
			ID:          "dataset-good-001",
			Workflow:    "streaming",
			Type:        "good",
			Description: "Stable network load test sample set.",
			SizeMB:      512,
			UpdatedAt:   now.Add(-48 * time.Hour),
			URI:         "s3://livepeer/datasets/streaming/good-001",
		},
		{
			ID:          "dataset-good-002",
			Workflow:    "inference",
			Type:        "good",
			Description: "Low-latency inference benchmark set.",
			SizeMB:      1536,
			UpdatedAt:   now.Add(-36 * time.Hour),
			URI:         "s3://livepeer/datasets/inference/good-002",
		},
		{
			ID:          "dataset-random-002",
			Workflow:    "inference",
			Type:        "random",
			Description: "Mixed inference prompts for baseline variance.",
			SizeMB:      2048,
			UpdatedAt:   now.Add(-24 * time.Hour),
			URI:         "s3://livepeer/datasets/inference/random-002",
		},
		{
			ID:          "dataset-random-003",
			Workflow:    "streaming",
			Type:        "random",
			Description: "Randomized bitrate and segment sizes.",
			SizeMB:      640,
			UpdatedAt:   now.Add(-12 * time.Hour),
			URI:         "s3://livepeer/datasets/streaming/random-003",
		},
		{
			ID:          "dataset-bad-003",
			Workflow:    "streaming",
			Type:        "bad",
			Description: "Adversarial network conditions for failure testing.",
			SizeMB:      768,
			UpdatedAt:   now.Add(-72 * time.Hour),
			URI:         "s3://livepeer/datasets/streaming/bad-003",
		},
		{
			ID:          "dataset-bad-004",
			Workflow:    "inference",
			Type:        "bad",
			Description: "Adversarial prompts causing retries.",
			SizeMB:      980,
			UpdatedAt:   now.Add(-96 * time.Hour),
			URI:         "s3://livepeer/datasets/inference/bad-004",
		},
	}

	filtered := make([]*models.Dataset, 0, len(datasets))
	for _, dataset := range datasets {
		if query.Workflow != "" && dataset.Workflow != query.Workflow {
			continue
		}
		if query.Type != "" && dataset.Type != query.Type {
			continue
		}
		filtered = append(filtered, dataset)
	}
	return filtered, nil
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func (m *MockStore) String() string {
	return fmt.Sprintf("MockStore")
}
