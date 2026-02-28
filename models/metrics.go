package models

import "time"

// --- GPU Metrics (v_api_gpu_metrics) ---

type GPUMetricsQuery struct {
	OrchestratorAddress string
	Pipeline            string
	ModelID             string
	GPUID               string
	Region              string
	GPUName             string
	RunnerVersion       string
	CudaVersion         string
	TimeRange           time.Duration
}

type GPUMetric struct {
	WindowStart         time.Time `json:"window_start"`
	OrchestratorAddress string    `json:"orchestrator_address"`
	Pipeline            string    `json:"pipeline"`
	ModelID             *string   `json:"model_id"`
	GPUID               *string   `json:"gpu_id"`
	Region              *string   `json:"region"`
	AvgOutputFPS        float64   `json:"avg_output_fps"`
	P95OutputFPS        float32   `json:"p95_output_fps"`
	JitterCoeffFPS      *float64  `json:"jitter_coeff_fps"`
	StatusSamples       uint64    `json:"status_samples"`

	// Hardware dimensions
	GPUName        *string `json:"gpu_name"`
	GPUMemoryTotal *uint64 `json:"gpu_memory_total"`
	RunnerVersion  *string `json:"runner_version"`
	CudaVersion    *string `json:"cuda_version"`

	// Latency metrics
	PromptToFirstFrameMs    *float64 `json:"prompt_to_first_frame_ms"`
	StartupTimeMs           *float64 `json:"startup_time_ms"`
	StartupTimeS            *float64 `json:"startup_time_s"`
	E2ELatencyMs            *float64 `json:"e2e_latency_ms"`
	P95PromptToFirstFrameMs *float32 `json:"p95_prompt_to_first_frame_ms"`
	P95StartupTimeMs        *float32 `json:"p95_startup_time_ms"`
	P95E2ELatencyMs         *float32 `json:"p95_e2e_latency_ms"`

	// Valid counts
	ValidPromptToFirstFrameCount uint64 `json:"valid_prompt_to_first_frame_count"`
	ValidStartupTimeCount        uint64 `json:"valid_startup_time_count"`
	ValidE2ELatencyCount         uint64 `json:"valid_e2e_latency_count"`

	// Session breakdowns
	KnownSessions                      uint64 `json:"known_sessions"`
	StartupSuccessSessions             uint64 `json:"startup_success_sessions"`
	ExcusedSessions                    uint64 `json:"excused_sessions"`
	UnexcusedSessions                  uint64 `json:"unexcused_sessions"`
	ConfirmedSwappedSessions           uint64 `json:"confirmed_swapped_sessions"`
	InferredOrchestratorChangeSessions uint64 `json:"inferred_orchestrator_change_sessions"`
	SwappedSessions                    uint64 `json:"swapped_sessions"`

	// Rates
	FailureRate float64 `json:"failure_rate"`
	SwapRate    float64 `json:"swap_rate"`
}

// --- Network Demand (v_api_network_demand) ---

type NetworkDemandQuery struct {
	Gateway  string
	Region   string
	Pipeline string
	ModelID  string
	Interval time.Duration
}

type NetworkDemandRow struct {
	WindowStart                        time.Time `json:"window_start"`
	Gateway                            string    `json:"gateway"`
	Region                             *string   `json:"region"`
	Pipeline                           string    `json:"pipeline"`
	ModelID                            *string   `json:"model_id"`
	TotalSessions                      uint64    `json:"total_sessions"`
	TotalStreams                       uint64    `json:"total_streams"`
	AvgOutputFPS                       float64   `json:"avg_output_fps"`
	TotalInferenceMinutes              float64   `json:"total_inference_minutes"`
	KnownSessions                      uint64    `json:"known_sessions"`
	ServedSessions                     uint64    `json:"served_sessions"`
	UnservedSessions                   uint64    `json:"unserved_sessions"`
	TotalDemandSessions                uint64    `json:"total_demand_sessions"`
	UnexcusedSessions                  uint64    `json:"unexcused_sessions"`
	ConfirmedSwappedSessions           uint64    `json:"confirmed_swapped_sessions"`
	InferredOrchestratorChangeSessions uint64    `json:"inferred_orchestrator_change_sessions"`
	SwappedSessions                    uint64    `json:"swapped_sessions"`
	MissingCapacityCount               uint64    `json:"missing_capacity_count"`
	SuccessRatio                       float64   `json:"success_ratio"`
	FeePaymentEth                      float64   `json:"fee_payment_eth"`
}

// --- SLA Compliance (v_api_sla_compliance) ---

type SLAComplianceQuery struct {
	OrchestratorAddress string
	Region              string
	Pipeline            string
	ModelID             string
	GPUID               string
	Period              time.Duration
}

type SLAComplianceRow struct {
	WindowStart                        time.Time `json:"window_start"`
	OrchestratorAddress                string    `json:"orchestrator_address"`
	Pipeline                           string    `json:"pipeline"`
	ModelID                            *string   `json:"model_id"`
	GPUID                              *string   `json:"gpu_id"`
	Region                             *string   `json:"region"`
	KnownSessions                      uint64    `json:"known_sessions"`
	StartupSuccessSessions             uint64    `json:"startup_success_sessions"`
	ExcusedSessions                    uint64    `json:"excused_sessions"`
	UnexcusedSessions                  uint64    `json:"unexcused_sessions"`
	ConfirmedSwappedSessions           uint64    `json:"confirmed_swapped_sessions"`
	InferredOrchestratorChangeSessions uint64    `json:"inferred_orchestrator_change_sessions"`
	SwappedSessions                    uint64    `json:"swapped_sessions"`
	SuccessRatio                       *float64  `json:"success_ratio"`
	NoSwapRatio                        *float64  `json:"no_swap_ratio"`
	SLAScore                           *float64  `json:"sla_score"`
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
