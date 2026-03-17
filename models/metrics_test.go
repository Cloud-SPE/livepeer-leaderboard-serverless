package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestGPUMetricNullableFieldsSerialization verifies that nullable pointer fields
// on GPUMetric (cuda_version, model_id, fps_jitter_coefficient, etc.) serialise
// as JSON null rather than being omitted when their values are nil. This covers
// the ClickHouse NULL coercion edge case where the DB returns NULL for these fields.
func TestGPUMetricNullableFieldsSerialization(t *testing.T) {
	m := &GPUMetric{
		WindowStart:         time.Now().UTC(),
		OrchestratorAddress: "0xabc",
		PipelineID:          "test-pipeline",
		// All nullable fields left nil
		ModelID:                        nil,
		GPUID:                          nil,
		Region:                         nil,
		FPSJitterCoefficient:           nil,
		GPUModelName:                   nil,
		GPUMemoryBytesTotal:            nil,
		RunnerVersion:                  nil,
		CudaVersion:                    nil,
		AvgPromptToFirstFrameMs:        nil,
		AvgStartupLatencyMs:            nil,
		AvgE2ELatencyMs:                nil,
		P95PromptToFirstFrameLatencyMs: nil,
		P95StartupLatencyMs:            nil,
		P95E2ELatencyMs:                nil,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	body := string(data)

	nullFields := []string{
		`"cuda_version":null`,
		`"model_id":null`,
		`"gpu_id":null`,
		`"fps_jitter_coefficient":null`,
		`"gpu_model_name":null`,
		`"runner_version":null`,
		`"avg_prompt_to_first_frame_ms":null`,
		`"avg_startup_latency_ms":null`,
		`"avg_e2e_latency_ms":null`,
	}

	for _, field := range nullFields {
		if !strings.Contains(body, field) {
			t.Errorf("expected %q in JSON output, got: %s", field, body)
		}
	}
}
