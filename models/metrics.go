package models

import "time"

// --- GPU Metrics (v_api_gpu_metrics) ---

type GPUMetricsQuery struct {
	OrchestratorAddress string
	PipelineID          string
	ModelID             string
	GPUID               string
	Region              string
	GPUModelName        string
	RunnerVersion       string
	CudaVersion         string
	TimeRange           time.Duration
}

type GPUMetric struct {
	WindowStart               time.Time `json:"window_start"`
	OrchestratorAddress       string    `json:"orchestrator_address"`
	PipelineID                string    `json:"pipeline_id"`
	ModelID                   *string   `json:"model_id"`
	GPUID                     *string   `json:"gpu_id"`
	Region                    *string   `json:"region"`
	AvgOutputFPS              float64   `json:"avg_output_fps"`
	P95OutputFPS              float32   `json:"p95_output_fps"`
	FPSJitterCoefficient      *float64  `json:"fps_jitter_coefficient"`
	StatusSamples             uint64    `json:"status_samples"`
	ErrorStatusSamples        uint64    `json:"error_status_samples"`
	HealthSignalCoverageRatio float64   `json:"health_signal_coverage_ratio"`

	// Hardware dimensions
	GPUModelName        *string `json:"gpu_model_name"`
	GPUMemoryBytesTotal *uint64 `json:"gpu_memory_bytes_total"`
	RunnerVersion       *string `json:"runner_version"`
	CudaVersion         *string `json:"cuda_version"`

	// Latency metrics
	AvgPromptToFirstFrameMs        *float64 `json:"avg_prompt_to_first_frame_ms"`
	AvgStartupLatencyMs            *float64 `json:"avg_startup_latency_ms"`
	AvgE2ELatencyMs                *float64 `json:"avg_e2e_latency_ms"`
	P95PromptToFirstFrameLatencyMs *float32 `json:"p95_prompt_to_first_frame_latency_ms"`
	P95StartupLatencyMs            *float32 `json:"p95_startup_latency_ms"`
	P95E2ELatencyMs                *float32 `json:"p95_e2e_latency_ms"`

	// Valid counts
	PromptToFirstFrameSampleCount uint64 `json:"prompt_to_first_frame_sample_count"`
	StartupLatencySampleCount     uint64 `json:"startup_latency_sample_count"`
	E2ELatencySampleCount         uint64 `json:"e2e_latency_sample_count"`

	// Session breakdowns
	KnownSessionsCount       uint64 `json:"known_sessions_count"`
	StartupSuccessSessions   uint64 `json:"startup_success_sessions"`
	StartupExcusedSessions   uint64 `json:"startup_excused_sessions"`
	StartupUnexcusedSessions uint64 `json:"startup_unexcused_sessions"`
	ConfirmedSwappedSessions uint64 `json:"confirmed_swapped_sessions"`
	InferredSwapSessions     uint64 `json:"inferred_swap_sessions"`
	TotalSwappedSessions     uint64 `json:"total_swapped_sessions"`
	SessionsEndingInError    uint64 `json:"sessions_ending_in_error"`

	// Rates
	StartupUnexcusedRate float64 `json:"startup_unexcused_rate"`
	SwapRate             float64 `json:"swap_rate"`
}

// --- Network Demand (v_api_network_demand) ---

type NetworkDemandQuery struct {
	Gateway    string
	Region     string
	PipelineID string
	ModelID    string
	Interval   time.Duration
}

type NetworkDemandRow struct {
	WindowStart               time.Time `json:"window_start"`
	Gateway                   string    `json:"gateway"`
	Region                    *string   `json:"region"`
	PipelineID                string    `json:"pipeline_id"`
	ModelID                   *string   `json:"model_id"`
	SessionsCount             uint64    `json:"sessions_count"`
	AvgOutputFPS              float64   `json:"avg_output_fps"`
	TotalMinutes              float64   `json:"total_minutes"`
	KnownSessionsCount        uint64    `json:"known_sessions_count"`
	ServedSessions            uint64    `json:"served_sessions"`
	UnservedSessions          uint64    `json:"unserved_sessions"`
	TotalDemandSessions       uint64    `json:"total_demand_sessions"`
	StartupUnexcusedSessions  uint64    `json:"startup_unexcused_sessions"`
	ConfirmedSwappedSessions  uint64    `json:"confirmed_swapped_sessions"`
	InferredSwapSessions      uint64    `json:"inferred_swap_sessions"`
	TotalSwappedSessions      uint64    `json:"total_swapped_sessions"`
	SessionsEndingInError     uint64    `json:"sessions_ending_in_error"`
	ErrorStatusSamples        uint64    `json:"error_status_samples"`
	HealthSignalCoverageRatio float64   `json:"health_signal_coverage_ratio"`
	StartupSuccessRate        float64   `json:"startup_success_rate"`
	EffectiveSuccessRate      float64   `json:"effective_success_rate"`
	TicketFaceValueEth        float64   `json:"ticket_face_value_eth"`
}

// --- SLA Compliance (v_api_sla_compliance) ---

type SLAComplianceQuery struct {
	OrchestratorAddress string
	Region              string
	PipelineID          string
	ModelID             string
	GPUID               string
	Period              time.Duration
}

type SLAComplianceRow struct {
	WindowStart               time.Time `json:"window_start"`
	OrchestratorAddress       string    `json:"orchestrator_address"`
	PipelineID                string    `json:"pipeline_id"`
	ModelID                   *string   `json:"model_id"`
	GPUID                     *string   `json:"gpu_id"`
	Region                    *string   `json:"region"`
	KnownSessionsCount        uint64    `json:"known_sessions_count"`
	StartupSuccessSessions    uint64    `json:"startup_success_sessions"`
	StartupExcusedSessions    uint64    `json:"startup_excused_sessions"`
	StartupUnexcusedSessions  uint64    `json:"startup_unexcused_sessions"`
	ConfirmedSwappedSessions  uint64    `json:"confirmed_swapped_sessions"`
	InferredSwapSessions      uint64    `json:"inferred_swap_sessions"`
	TotalSwappedSessions      uint64    `json:"total_swapped_sessions"`
	SessionsEndingInError     uint64    `json:"sessions_ending_in_error"`
	ErrorStatusSamples        uint64    `json:"error_status_samples"`
	HealthSignalCoverageRatio float64   `json:"health_signal_coverage_ratio"`
	StartupSuccessRate        *float64  `json:"startup_success_rate"`
	EffectiveSuccessRate      *float64  `json:"effective_success_rate"`
	NoSwapRate                *float64  `json:"no_swap_rate"`
	SLAScore                  *float64  `json:"sla_score"`
}

// --- Datasets (no view yet, hard-coded) ---

type DatasetsQuery struct {
	Workflow string
	Type     string
}

type Dataset struct {
	ID          string    `json:"id"`
	Workflow    string    `json:"workflow"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	SizeMB      int       `json:"size_mb"`
	UpdatedAt   time.Time `json:"updated_at"`
	URI         string    `json:"uri"`
}
