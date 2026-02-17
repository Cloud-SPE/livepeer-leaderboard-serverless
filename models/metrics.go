package models

import "time"

// --- GPU Metrics (v_api_gpu_metrics) ---

type GPUMetricsQuery struct {
	OrchestratorAddress string
	Pipeline            string
	PipelineID          string
	ModelID             string
	GPUID               string
	Region              string
	TimeRange           time.Duration
}

type GPUMetric struct {
	WindowStart         time.Time `json:"window_start"`
	OrchestratorAddress string    `json:"orchestrator_address"`
	Pipeline            string    `json:"pipeline"`
	PipelineID          string    `json:"pipeline_id"`
	ModelID             *string   `json:"model_id"`
	GPUID               *string   `json:"gpu_id"`
	Region              *string   `json:"region"`
	AvgOutputFPS        float64   `json:"avg_output_fps"`
	P95OutputFPS        float32   `json:"p95_output_fps"`
	JitterCoeffFPS      *float64  `json:"jitter_coeff_fps"`
	StatusSamples       uint64    `json:"status_samples"`
}

// --- Network Demand (v_api_network_demand) ---

type NetworkDemandQuery struct {
	Gateway    string
	Region     string
	Pipeline   string
	PipelineID string
	Interval   time.Duration
}

type NetworkDemandRow struct {
	WindowStart    time.Time `json:"window_start"`
	Gateway        string    `json:"gateway"`
	Region         *string   `json:"region"`
	Pipeline       string    `json:"pipeline"`
	PipelineID     string    `json:"pipeline_id"`
	ActiveSessions uint64    `json:"active_sessions"`
	ActiveStreams  uint64    `json:"active_streams"`
	AvgOutputFPS   float64   `json:"avg_output_fps"`
}

// --- SLA Compliance (v_api_sla_compliance) ---

type SLAComplianceQuery struct {
	OrchestratorAddress string
	Region              string
	Pipeline            string
	PipelineID          string
	ModelID             string
	GPUID               string
	Period              time.Duration
}

type SLAComplianceRow struct {
	WindowStart         time.Time `json:"window_start"`
	OrchestratorAddress string    `json:"orchestrator_address"`
	Pipeline            string    `json:"pipeline"`
	PipelineID          string    `json:"pipeline_id"`
	ModelID             *string   `json:"model_id"`
	GPUID               *string   `json:"gpu_id"`
	Region              *string   `json:"region"`
	KnownSessions       uint64    `json:"known_sessions"`
	UnexcusedSessions   uint64    `json:"unexcused_sessions"`
	SwappedSessions     uint64    `json:"swapped_sessions"`
	SuccessRatio        *float64  `json:"success_ratio"`
	NoSwapRatio         *float64  `json:"no_swap_ratio"`
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
