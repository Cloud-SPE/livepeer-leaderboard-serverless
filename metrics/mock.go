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

	orchAddr := "0x0abe02f6ef1fa8c29f9b3f9f170c6f3681fd3031"
	if query.OrchestratorAddress != "" {
		orchAddr = query.OrchestratorAddress
	}
	pipeline := "streamdiffusion-sdxl"
	if query.Pipeline != "" {
		pipeline = query.Pipeline
	}

	metrics := make([]*models.GPUMetric, 0, 6)
	for i := 0; i < 6; i++ {
		jitter := 0.45 + float64(i)*0.08
		metrics = append(metrics, &models.GPUMetric{
			WindowStart:         now.Add(-time.Duration(i) * time.Minute),
			OrchestratorAddress: orchAddr,
			Pipeline:            pipeline,
			PipelineID:          "",
			ModelID:             nilIfEmpty(query.ModelID),
			GPUID:               nilIfEmpty(query.GPUID),
			Region:              nilIfEmpty(query.Region),
			AvgOutputFPS:        14.67 - float64(i)*1.2,
			P95OutputFPS:        float32(18.19 - float64(i)*1.0),
			JitterCoeffFPS:      &jitter,
			StatusSamples:       uint64(6 - i),
		})
	}
	return metrics, nil
}

func (m *MockStore) NetworkDemand(query *models.NetworkDemandQuery) ([]*models.NetworkDemandRow, error) {
	now := time.Now().UTC()
	interval := query.Interval
	if interval <= 0 {
		interval = 15 * time.Minute
	}

	gateway := "cloud-spe-ai-live-video-tester-mdw"
	if query.Gateway != "" {
		gateway = query.Gateway
	}
	pipeline := "streamdiffusion-sdxl"
	if query.Pipeline != "" {
		pipeline = query.Pipeline
	}

	rows := make([]*models.NetworkDemandRow, 0, 12)
	for i := 11; i >= 0; i-- {
		rows = append(rows, &models.NetworkDemandRow{
			WindowStart:    now.Add(-time.Duration(i) * interval),
			Gateway:        gateway,
			Region:         nilIfEmpty(query.Region),
			Pipeline:       pipeline,
			PipelineID:     "",
			ActiveSessions: 1,
			ActiveStreams:  1,
			AvgOutputFPS:   11.99 + float64(i)*0.5,
		})
	}
	return rows, nil
}

func (m *MockStore) SLACompliance(query *models.SLAComplianceQuery) ([]*models.SLAComplianceRow, error) {
	now := time.Now().UTC()
	period := query.Period
	if period <= 0 {
		period = 24 * time.Hour
	}

	orchAddr := "0x5263e0ce3a97b634d8828ce4337ad0f70b30b077"
	if query.OrchestratorAddress != "" {
		orchAddr = query.OrchestratorAddress
	}
	pipeline := "streamdiffusion-sdxl"
	if query.Pipeline != "" {
		pipeline = query.Pipeline
	}

	successRatio := 1.0
	noSwapRatio := 0.75
	gpuID := "GPU-3f93b3ef-7ea7-4480-aa80-75d59014fb74"
	modelID := "streamdiffusion-sdxl"
	if query.ModelID != "" {
		modelID = query.ModelID
	}
	if query.GPUID != "" {
		gpuID = query.GPUID
	}

	rows := []*models.SLAComplianceRow{
		{
			WindowStart:         now.Add(-period),
			OrchestratorAddress: orchAddr,
			Pipeline:            pipeline,
			PipelineID:          "",
			ModelID:             &modelID,
			GPUID:               &gpuID,
			Region:              nilIfEmpty(query.Region),
			KnownSessions:       4,
			UnexcusedSessions:   0,
			SwappedSessions:     1,
			SuccessRatio:        &successRatio,
			NoSwapRatio:         &noSwapRatio,
		},
	}
	return rows, nil
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

func (m *MockStore) String() string {
	return fmt.Sprintf("MockStore")
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
