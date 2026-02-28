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
	modelID := "streamdiffusion-sdxl"
	if query.ModelID != "" {
		modelID = query.ModelID
	}

	gpuName := "NVIDIA RTX 4090"
	var gpuMemory uint64 = 24576
	runnerVersion := "0.9.1"
	cudaVersion := "12.4"

	metrics := make([]*models.GPUMetric, 0, 6)
	for i := 0; i < 6; i++ {
		jitter := 0.45 + float64(i)*0.08
		promptToFirstFrame := 120.5 + float64(i)*10.0
		startupTimeMs := 250.0 + float64(i)*15.0
		startupTimeS := startupTimeMs / 1000.0
		e2eLatency := 350.0 + float64(i)*20.0
		p95Prompt := float32(180.0 + float64(i)*12.0)
		p95Startup := float32(400.0 + float64(i)*18.0)
		p95E2E := float32(500.0 + float64(i)*25.0)

		metrics = append(metrics, &models.GPUMetric{
			WindowStart:         now.Add(-time.Duration(i) * time.Minute),
			OrchestratorAddress: orchAddr,
			Pipeline:            pipeline,
			ModelID:             &modelID,
			GPUID:               nilIfEmpty(query.GPUID),
			Region:              nilIfEmpty(query.Region),
			AvgOutputFPS:        14.67 - float64(i)*1.2,
			P95OutputFPS:        float32(18.19 - float64(i)*1.0),
			JitterCoeffFPS:      &jitter,
			StatusSamples:       uint64(6 - i),

			GPUName:        &gpuName,
			GPUMemoryTotal: &gpuMemory,
			RunnerVersion:  &runnerVersion,
			CudaVersion:    &cudaVersion,

			PromptToFirstFrameMs:    &promptToFirstFrame,
			StartupTimeMs:           &startupTimeMs,
			StartupTimeS:            &startupTimeS,
			E2ELatencyMs:            &e2eLatency,
			P95PromptToFirstFrameMs: &p95Prompt,
			P95StartupTimeMs:        &p95Startup,
			P95E2ELatencyMs:         &p95E2E,

			ValidPromptToFirstFrameCount: uint64(5 - i%5),
			ValidStartupTimeCount:        uint64(5 - i%5),
			ValidE2ELatencyCount:         uint64(5 - i%5),

			KnownSessions:                      uint64(10 - i),
			StartupSuccessSessions:             uint64(8 - i),
			ExcusedSessions:                    1,
			UnexcusedSessions:                  uint64(i % 2),
			ConfirmedSwappedSessions:           uint64(i % 2),
			InferredOrchestratorChangeSessions: uint64((i + 1) % 2),
			SwappedSessions:                    uint64(i % 3),

			FailureRate: float64(i%2) * 0.05,
			SwapRate:    float64(i%3) * 0.03,
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
	modelID := "streamdiffusion-sdxl"
	if query.ModelID != "" {
		modelID = query.ModelID
	}

	rows := make([]*models.NetworkDemandRow, 0, 12)
	for i := 11; i >= 0; i-- {
		rows = append(rows, &models.NetworkDemandRow{
			WindowStart:                        now.Add(-time.Duration(i) * interval),
			Gateway:                            gateway,
			Region:                             nilIfEmpty(query.Region),
			Pipeline:                           pipeline,
			ModelID:                            &modelID,
			TotalSessions:                      uint64(3 + i),
			TotalStreams:                       uint64(2 + i),
			AvgOutputFPS:                       11.99 + float64(i)*0.5,
			TotalInferenceMinutes:              45.5 + float64(i)*3.0,
			KnownSessions:                      uint64(3 + i),
			ServedSessions:                     uint64(2 + i),
			UnservedSessions:                   1,
			TotalDemandSessions:                uint64(4 + i),
			UnexcusedSessions:                  uint64(i % 2),
			ConfirmedSwappedSessions:           uint64(i % 2),
			InferredOrchestratorChangeSessions: uint64((i + 1) % 2),
			SwappedSessions:                    uint64(i % 3),
			MissingCapacityCount:               uint64(i % 2),
			SuccessRatio:                       0.92 + float64(i)*0.005,
			FeePaymentEth:                      0.0025 + float64(i)*0.0003,
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
	slaScore := 0.95
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
			WindowStart:                        now.Add(-period),
			OrchestratorAddress:                orchAddr,
			Pipeline:                           pipeline,
			ModelID:                            &modelID,
			GPUID:                              &gpuID,
			Region:                             nilIfEmpty(query.Region),
			KnownSessions:                      4,
			StartupSuccessSessions:             3,
			ExcusedSessions:                    1,
			UnexcusedSessions:                  0,
			ConfirmedSwappedSessions:           1,
			InferredOrchestratorChangeSessions: 0,
			SwappedSessions:                    1,
			SuccessRatio:                       &successRatio,
			NoSwapRatio:                        &noSwapRatio,
			SLAScore:                           &slaScore,
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
