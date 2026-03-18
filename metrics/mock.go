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
	pipelineID := "streamdiffusion-sdxl"
	if query.PipelineID != "" {
		pipelineID = query.PipelineID
	}
	modelID := "streamdiffusion-sdxl"
	if query.ModelID != "" {
		modelID = query.ModelID
	}
	org := nilIfEmpty(query.Org)

	gpuModelName := "NVIDIA RTX 4090"
	var gpuMemory uint64 = 24576
	runnerVersion := "0.9.1"
	cudaVersion := "12.4"

	metrics := make([]*models.GPUMetric, 0, 6)
	for i := 0; i < 6; i++ {
		jitter := 0.45 + float64(i)*0.08
		promptToFirstFrame := 120.5 + float64(i)*10.0
		startupTimeMs := 250.0 + float64(i)*15.0
		e2eLatency := 350.0 + float64(i)*20.0
		p95Prompt := float32(180.0 + float64(i)*12.0)
		p95Startup := float32(400.0 + float64(i)*18.0)
		p95E2E := float32(500.0 + float64(i)*25.0)

		metrics = append(metrics, &models.GPUMetric{
			WindowStart:               now.Add(-time.Duration(i) * time.Minute),
			Org:                       org,
			OrchestratorAddress:       orchAddr,
			PipelineID:                pipelineID,
			ModelID:                   &modelID,
			GPUID:                     nilIfEmpty(query.GPUID),
			Region:                    nilIfEmpty(query.Region),
			AvgOutputFPS:              14.67 - float64(i)*1.2,
			P95OutputFPS:              float32(18.19 - float64(i)*1.0),
			FPSJitterCoefficient:      &jitter,
			StatusSamples:             uint64(6 - i),
			SessionsEndingInError:     uint64(i % 2),
			ErrorStatusSamples:        uint64(1 + i%2),
			HealthSignalCoverageRatio: 0.95,

			GPUModelName:        &gpuModelName,
			GPUMemoryBytesTotal: &gpuMemory,
			RunnerVersion:       &runnerVersion,
			CudaVersion:         &cudaVersion,

			AvgPromptToFirstFrameMs:        &promptToFirstFrame,
			AvgStartupLatencyMs:            &startupTimeMs,
			AvgE2ELatencyMs:                &e2eLatency,
			P95PromptToFirstFrameLatencyMs: &p95Prompt,
			P95StartupLatencyMs:            &p95Startup,
			P95E2ELatencyMs:                &p95E2E,

			PromptToFirstFrameSampleCount: uint64(5 - i%5),
			StartupLatencySampleCount:     uint64(5 - i%5),
			E2ELatencySampleCount:         uint64(5 - i%5),

			KnownSessionsCount:       uint64(10 - i),
			StartupSuccessSessions:   uint64(8 - i),
			StartupExcusedSessions:   1,
			StartupUnexcusedSessions: uint64(i % 2),
			ConfirmedSwappedSessions: uint64(i % 2),
			InferredSwapSessions:     uint64((i + 1) % 2),
			TotalSwappedSessions:     uint64(i % 3),

			StartupUnexcusedRate: float64(i%2) * 0.05,
			SwapRate:             float64(i%3) * 0.03,
		})
	}
	return metrics, nil
}

func (m *MockStore) GPUMetricsCount(_ *models.GPUMetricsQuery) (int, error) {
	return 6, nil
}

func (m *MockStore) NetworkDemand(query *models.NetworkDemandQuery) ([]*models.NetworkDemandRow, error) {
	now := time.Now().UTC()
	window := query.Window
	if window <= 0 {
		window = 3 * time.Hour
	}
	slotSize := window / 12

	gateway := "cloud-spe-ai-live-video-tester-mdw"
	if query.Gateway != "" {
		gateway = query.Gateway
	}
	pipelineID := "streamdiffusion-sdxl"
	if query.PipelineID != "" {
		pipelineID = query.PipelineID
	}
	modelID := "streamdiffusion-sdxl"
	if query.ModelID != "" {
		modelID = query.ModelID
	}
	org := nilIfEmpty(query.Org)

	rows := make([]*models.NetworkDemandRow, 0, 12)
	for i := 11; i >= 0; i-- {
		rows = append(rows, &models.NetworkDemandRow{
			WindowStart:               now.Add(-time.Duration(i) * slotSize),
			Org:                       org,
			Gateway:                   gateway,
			Region:                    nilIfEmpty(query.Region),
			PipelineID:                pipelineID,
			ModelID:                   &modelID,
			SessionsCount:             uint64(3 + i),
			AvgOutputFPS:              11.99 + float64(i)*0.5,
			TotalMinutes:              45.5 + float64(i)*3.0,
			KnownSessionsCount:        uint64(3 + i),
			ServedSessions:            uint64(2 + i),
			UnservedSessions:          1,
			TotalDemandSessions:       uint64(4 + i),
			StartupUnexcusedSessions:  uint64(i % 2),
			ConfirmedSwappedSessions:  uint64(i % 2),
			InferredSwapSessions:      uint64((i + 1) % 2),
			TotalSwappedSessions:      uint64(i % 3),
			SessionsEndingInError:     uint64(i % 2),
			ErrorStatusSamples:        uint64(1 + i%2),
			HealthSignalCoverageRatio: 0.95,
			StartupSuccessRate:        0.95 + float64(i)*0.002,
			EffectiveSuccessRate:      0.92 + float64(i)*0.005,
			TicketFaceValueEth:        0.0025 + float64(i)*0.0003,
		})
	}
	return rows, nil
}

func (m *MockStore) NetworkDemandCount(_ *models.NetworkDemandQuery) (int, error) {
	return 12, nil
}

func (m *MockStore) SLACompliance(query *models.SLAComplianceQuery) ([]*models.SLAComplianceRow, error) {
	now := time.Now().UTC()
	window := query.Window
	if window <= 0 {
		window = 24 * time.Hour
	}

	orchAddr := "0x5263e0ce3a97b634d8828ce4337ad0f70b30b077"
	if query.OrchestratorAddress != "" {
		orchAddr = query.OrchestratorAddress
	}
	pipelineID := "streamdiffusion-sdxl"
	if query.PipelineID != "" {
		pipelineID = query.PipelineID
	}
	org := nilIfEmpty(query.Org)

	successRatio := 1.0
	startupSuccessRatio := 1.0
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
			WindowStart:               now.Add(-window),
			Org:                       org,
			OrchestratorAddress:       orchAddr,
			PipelineID:                pipelineID,
			ModelID:                   &modelID,
			GPUID:                     &gpuID,
			Region:                    nilIfEmpty(query.Region),
			KnownSessionsCount:        4,
			StartupSuccessSessions:    3,
			StartupExcusedSessions:    1,
			StartupUnexcusedSessions:  0,
			ConfirmedSwappedSessions:  1,
			InferredSwapSessions:      0,
			TotalSwappedSessions:      1,
			SessionsEndingInError:     0,
			ErrorStatusSamples:        1,
			HealthSignalCoverageRatio: 1.0,
			StartupSuccessRate:        &startupSuccessRatio,
			EffectiveSuccessRate:      &successRatio,
			NoSwapRate:                &noSwapRatio,
			SLAScore:                  &slaScore,
		},
	}
	return rows, nil
}

func (m *MockStore) SLAComplianceCount(_ *models.SLAComplianceQuery) (int, error) {
	return 1, nil
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
