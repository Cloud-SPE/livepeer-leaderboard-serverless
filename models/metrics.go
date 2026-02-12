package models

import "time"

type GPUMetricsQuery struct {
	OrchestratorWallet string
	GPUId              string
	Region             string
	Workflow           string
	TimeRange          time.Duration
}

type GPUMetric struct {
	OrchestratorWallet string    `json:"o_wallet"`
	GPUId              string    `json:"gpu_id"`
	Region             string    `json:"region"`
	Workflow           string    `json:"workflow"`
	Timestamp          time.Time `json:"timestamp"`
	UtilizationPct     float64   `json:"utilization_pct"`
	MemoryUsedMB       float64   `json:"memory_used_mb"`
	MemoryTotalMB      float64   `json:"memory_total_mb"`
	TemperatureC       float64   `json:"temperature_c"`
	PowerWatts         float64   `json:"power_watts"`
	SuccessRate        float64   `json:"success_rate"`
	ErrorRate          float64   `json:"error_rate"`
}

type NetworkDemandQuery struct {
	Gateway  string
	Region   string
	Workflow string
	Interval time.Duration
}

type NetworkDemandPoint struct {
	Timestamp     time.Time `json:"timestamp"`
	StreamMinutes float64   `json:"stream_minutes"`
	InferMinutes  float64   `json:"inference_minutes"`
}

type NetworkDemand struct {
	Gateway  string               `json:"gateway"`
	Region   string               `json:"region"`
	Workflow string               `json:"workflow"`
	Interval string               `json:"interval"`
	Points   []NetworkDemandPoint `json:"points"`
}

type SLAComplianceQuery struct {
	OrchestratorID string
	Period         time.Duration
}

type SLACompliance struct {
	OrchestratorID string    `json:"orchestrator_id"`
	Period         string    `json:"period"`
	Score          float64   `json:"score"`
	WindowStart    time.Time `json:"window_start"`
	WindowEnd      time.Time `json:"window_end"`
}

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
