package metrics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/livepeer/leaderboard-serverless/common"
	"github.com/livepeer/leaderboard-serverless/models"
)

type ClickhouseStore struct {
	db *sql.DB
}

func NewClickhouseStoreFromEnv() (*ClickhouseStore, error) {
	host := envOrDefault("CLICKHOUSE_HOST", "localhost")
	port := envOrDefault("CLICKHOUSE_PORT", "8123")
	database := envOrDefault("CLICKHOUSE_DB", "livepeer_analytics")
	user := envOrDefault("CLICKHOUSE_USER", "analytics_user")
	password := envOrDefault("CLICKHOUSE_PASS", "analytics_password")
	protocol := strings.ToLower(envOrDefault("CLICKHOUSE_PROTOCOL", "http"))

	addr := fmt.Sprintf("%s:%s", host, port)
	common.Logger.Info("Connecting ClickHouse metrics store to %s/%s as %s (protocol=%s)", addr, database, user, protocol)
	db := clickhouse.OpenDB(&clickhouse.Options{
		Addr:     []string{addr},
		Protocol: protocolFromString(protocol),
		Auth: clickhouse.Auth{
			Database: database,
			Username: user,
			Password: password,
		},
		DialTimeout: 5 * time.Second,
		Compression: &clickhouse.Compression{Method: clickhouse.CompressionLZ4},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		common.Logger.Warn("ClickHouse ping failed: %v", err)
		return nil, err
	}
	common.Logger.Info("ClickHouse metrics store connected")

	return &ClickhouseStore{db: db}, nil
}

// --- GPU Metrics ---

func (s *ClickhouseStore) GPUMetrics(query *models.GPUMetricsQuery) ([]*models.GPUMetric, error) {
	if query == nil {
		return nil, errors.New("gpu metrics query cannot be nil")
	}

	end := time.Now().UTC()
	timeRange := query.TimeRange
	if timeRange <= 0 {
		timeRange = time.Hour
	}
	start := end.Add(-timeRange)

	sqlQuery := `SELECT
		window_start, orchestrator_address, pipeline_id,
		model_id, gpu_id, region,
		avg_output_fps, p95_output_fps, fps_jitter_coefficient, status_samples,
		sessions_ending_in_error, error_status_samples, health_signal_coverage_ratio,
		gpu_model_name, gpu_memory_bytes_total, runner_version, cuda_version,
		avg_prompt_to_first_frame_ms, avg_startup_latency_ms, avg_e2e_latency_ms,
		p95_prompt_to_first_frame_latency_ms, p95_startup_latency_ms, p95_e2e_latency_ms,
		prompt_to_first_frame_sample_count, startup_latency_sample_count, e2e_latency_sample_count,
		known_sessions_count, startup_success_sessions, startup_excused_sessions, startup_unexcused_sessions,
		confirmed_swapped_sessions, inferred_swap_sessions, total_swapped_sessions,
		startup_unexcused_rate, swap_rate
	FROM v_api_gpu_metrics
	WHERE window_start >= ? AND window_start <= ?`

	args := []interface{}{start, end}

	if query.OrchestratorAddress != "" {
		sqlQuery += " AND orchestrator_address = ?"
		args = append(args, query.OrchestratorAddress)
	}
	if query.PipelineID != "" {
		sqlQuery += " AND pipeline_id = ?"
		args = append(args, query.PipelineID)
	}
	if query.ModelID != "" {
		sqlQuery += " AND model_id = ?"
		args = append(args, query.ModelID)
	}
	if query.GPUID != "" {
		sqlQuery += " AND gpu_id = ?"
		args = append(args, query.GPUID)
	}
	if query.Region != "" {
		sqlQuery += " AND region = ?"
		args = append(args, query.Region)
	}
	if query.GPUModelName != "" {
		sqlQuery += " AND gpu_model_name = ?"
		args = append(args, query.GPUModelName)
	}
	if query.RunnerVersion != "" {
		sqlQuery += " AND runner_version = ?"
		args = append(args, query.RunnerVersion)
	}
	if query.CudaVersion != "" {
		sqlQuery += " AND cuda_version = ?"
		args = append(args, query.CudaVersion)
	}

	pageSize := query.Pagination.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := (query.Pagination.Page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	sqlQuery += " ORDER BY window_start DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	common.Logger.Debug("ClickHouse GPUMetrics query start=%v end=%v", start, end)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("gpu metrics query failed: %w", err)
	}
	defer rows.Close()

	results := make([]*models.GPUMetric, 0, 50)
	for rows.Next() {
		m := &models.GPUMetric{}
		// clickhouse-go returns Nullable columns as pointers
		var modelID, gpuID, region *string
		var fpsJitterCoefficient *float64
		var avgOutputFPS sql.NullFloat64
		var p95OutputFPS sql.NullFloat64
		var gpuModelName, runnerVersion, cudaVersion *string
		var gpuMemoryBytesTotal *uint64
		var avgPromptToFirstFrameMs, avgStartupLatencyMs, avgE2ELatencyMs *float64
		var p95PromptToFirstFrameLatencyMs, p95StartupLatencyMs, p95E2ELatencyMs *float32

		if err := rows.Scan(
			&m.WindowStart, &m.OrchestratorAddress, &m.PipelineID,
			&modelID, &gpuID, &region,
			&avgOutputFPS, &p95OutputFPS, &fpsJitterCoefficient, &m.StatusSamples,
			&m.SessionsEndingInError, &m.ErrorStatusSamples, &m.HealthSignalCoverageRatio,
			&gpuModelName, &gpuMemoryBytesTotal, &runnerVersion, &cudaVersion,
			&avgPromptToFirstFrameMs, &avgStartupLatencyMs, &avgE2ELatencyMs,
			&p95PromptToFirstFrameLatencyMs, &p95StartupLatencyMs, &p95E2ELatencyMs,
			&m.PromptToFirstFrameSampleCount, &m.StartupLatencySampleCount, &m.E2ELatencySampleCount,
			&m.KnownSessionsCount, &m.StartupSuccessSessions, &m.StartupExcusedSessions, &m.StartupUnexcusedSessions,
			&m.ConfirmedSwappedSessions, &m.InferredSwapSessions, &m.TotalSwappedSessions,
			&m.StartupUnexcusedRate, &m.SwapRate,
		); err != nil {
			return nil, fmt.Errorf("gpu metrics scan failed: %w", err)
		}

		m.ModelID = modelID
		m.GPUID = gpuID
		m.Region = region
		if avgOutputFPS.Valid {
			m.AvgOutputFPS = avgOutputFPS.Float64
		}
		m.FPSJitterCoefficient = fpsJitterCoefficient
		if p95OutputFPS.Valid {
			m.P95OutputFPS = float32(p95OutputFPS.Float64)
		}
		m.GPUModelName = gpuModelName
		m.GPUMemoryBytesTotal = gpuMemoryBytesTotal
		m.RunnerVersion = runnerVersion
		m.CudaVersion = cudaVersion
		m.AvgPromptToFirstFrameMs = avgPromptToFirstFrameMs
		m.AvgStartupLatencyMs = avgStartupLatencyMs
		m.AvgE2ELatencyMs = avgE2ELatencyMs
		m.P95PromptToFirstFrameLatencyMs = p95PromptToFirstFrameLatencyMs
		m.P95StartupLatencyMs = p95StartupLatencyMs
		m.P95E2ELatencyMs = p95E2ELatencyMs

		results = append(results, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gpu metrics rows iteration: %w", err)
	}

	common.Logger.Debug("ClickHouse GPUMetrics returned %d rows", len(results))
	return results, nil
}

func (s *ClickhouseStore) GPUMetricsCount(query *models.GPUMetricsQuery) (int, error) {
	if query == nil {
		return 0, errors.New("gpu metrics query cannot be nil")
	}

	end := time.Now().UTC()
	timeRange := query.TimeRange
	if timeRange <= 0 {
		timeRange = time.Hour
	}
	start := end.Add(-timeRange)

	sqlQuery := `SELECT count() FROM v_api_gpu_metrics WHERE window_start >= ? AND window_start <= ?`
	args := []interface{}{start, end}

	if query.OrchestratorAddress != "" {
		sqlQuery += " AND orchestrator_address = ?"
		args = append(args, query.OrchestratorAddress)
	}
	if query.PipelineID != "" {
		sqlQuery += " AND pipeline_id = ?"
		args = append(args, query.PipelineID)
	}
	if query.ModelID != "" {
		sqlQuery += " AND model_id = ?"
		args = append(args, query.ModelID)
	}
	if query.GPUID != "" {
		sqlQuery += " AND gpu_id = ?"
		args = append(args, query.GPUID)
	}
	if query.Region != "" {
		sqlQuery += " AND region = ?"
		args = append(args, query.Region)
	}
	if query.GPUModelName != "" {
		sqlQuery += " AND gpu_model_name = ?"
		args = append(args, query.GPUModelName)
	}
	if query.RunnerVersion != "" {
		sqlQuery += " AND runner_version = ?"
		args = append(args, query.RunnerVersion)
	}
	if query.CudaVersion != "" {
		sqlQuery += " AND cuda_version = ?"
		args = append(args, query.CudaVersion)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var total int
	if err := s.db.QueryRowContext(ctx, sqlQuery, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("gpu metrics count query failed: %w", err)
	}
	return total, nil
}

// --- Network Demand ---

func (s *ClickhouseStore) NetworkDemand(query *models.NetworkDemandQuery) ([]*models.NetworkDemandRow, error) {
	if query == nil {
		return nil, errors.New("network demand query cannot be nil")
	}

	end := time.Now().UTC()
	interval := query.Interval
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	start := end.Add(-interval * 12)

	sqlQuery := `SELECT
		window_start, gateway, region, pipeline_id, model_id,
		sessions_count, avg_output_fps,
		total_minutes, known_sessions_count, served_sessions, unserved_sessions,
		total_demand_sessions, startup_unexcused_sessions,
		confirmed_swapped_sessions, inferred_swap_sessions, total_swapped_sessions,
		sessions_ending_in_error, error_status_samples, health_signal_coverage_ratio,
		startup_success_rate, effective_success_rate, ticket_face_value_eth
	FROM v_api_network_demand
	WHERE window_start >= ? AND window_start <= ?`

	args := []interface{}{start, end}

	if query.Gateway != "" {
		sqlQuery += " AND gateway = ?"
		args = append(args, query.Gateway)
	}
	if query.Region != "" {
		sqlQuery += " AND region = ?"
		args = append(args, query.Region)
	}
	if query.PipelineID != "" {
		sqlQuery += " AND pipeline_id = ?"
		args = append(args, query.PipelineID)
	}
	if query.ModelID != "" {
		sqlQuery += " AND model_id = ?"
		args = append(args, query.ModelID)
	}

	pageSize := query.Pagination.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := (query.Pagination.Page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	sqlQuery += " ORDER BY window_start DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	common.Logger.Debug("ClickHouse NetworkDemand query start=%v end=%v", start, end)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("network demand query failed: %w", err)
	}
	defer rows.Close()

	results := make([]*models.NetworkDemandRow, 0, 50)
	for rows.Next() {
		r := &models.NetworkDemandRow{}
		// clickhouse-go returns Nullable columns as pointers
		var region, modelID *string

		if err := rows.Scan(
			&r.WindowStart, &r.Gateway, &region, &r.PipelineID, &modelID,
			&r.SessionsCount, &r.AvgOutputFPS,
			&r.TotalMinutes, &r.KnownSessionsCount, &r.ServedSessions, &r.UnservedSessions,
			&r.TotalDemandSessions, &r.StartupUnexcusedSessions,
			&r.ConfirmedSwappedSessions, &r.InferredSwapSessions, &r.TotalSwappedSessions,
			&r.SessionsEndingInError, &r.ErrorStatusSamples, &r.HealthSignalCoverageRatio,
			&r.StartupSuccessRate, &r.EffectiveSuccessRate, &r.TicketFaceValueEth,
		); err != nil {
			return nil, fmt.Errorf("network demand scan failed: %w", err)
		}

		r.Region = region
		r.ModelID = modelID
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("network demand rows iteration: %w", err)
	}

	common.Logger.Debug("ClickHouse NetworkDemand returned %d rows", len(results))
	return results, nil
}

func (s *ClickhouseStore) NetworkDemandCount(query *models.NetworkDemandQuery) (int, error) {
	if query == nil {
		return 0, errors.New("network demand query cannot be nil")
	}

	end := time.Now().UTC()
	interval := query.Interval
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	start := end.Add(-interval * 12)

	sqlQuery := `SELECT count() FROM v_api_network_demand WHERE window_start >= ? AND window_start <= ?`
	args := []interface{}{start, end}

	if query.Gateway != "" {
		sqlQuery += " AND gateway = ?"
		args = append(args, query.Gateway)
	}
	if query.Region != "" {
		sqlQuery += " AND region = ?"
		args = append(args, query.Region)
	}
	if query.PipelineID != "" {
		sqlQuery += " AND pipeline_id = ?"
		args = append(args, query.PipelineID)
	}
	if query.ModelID != "" {
		sqlQuery += " AND model_id = ?"
		args = append(args, query.ModelID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var total int
	if err := s.db.QueryRowContext(ctx, sqlQuery, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("network demand count query failed: %w", err)
	}
	return total, nil
}

// --- SLA Compliance ---

func (s *ClickhouseStore) SLACompliance(query *models.SLAComplianceQuery) ([]*models.SLAComplianceRow, error) {
	if query == nil {
		return nil, errors.New("sla compliance query cannot be nil")
	}

	end := time.Now().UTC()
	period := query.Period
	if period <= 0 {
		period = 24 * time.Hour
	}
	start := end.Add(-period)

	sqlQuery := `SELECT
		window_start, orchestrator_address, pipeline_id,
		model_id, gpu_id, region,
		known_sessions_count, startup_success_sessions, startup_excused_sessions,
		startup_unexcused_sessions, confirmed_swapped_sessions, inferred_swap_sessions, total_swapped_sessions,
		sessions_ending_in_error, error_status_samples, health_signal_coverage_ratio,
		startup_success_rate, effective_success_rate, no_swap_rate, sla_score
	FROM v_api_sla_compliance
	WHERE window_start >= ? AND window_start <= ?`

	args := []interface{}{start, end}

	if query.OrchestratorAddress != "" {
		sqlQuery += " AND orchestrator_address = ?"
		args = append(args, query.OrchestratorAddress)
	}
	if query.PipelineID != "" {
		sqlQuery += " AND pipeline_id = ?"
		args = append(args, query.PipelineID)
	}
	if query.ModelID != "" {
		sqlQuery += " AND model_id = ?"
		args = append(args, query.ModelID)
	}
	if query.GPUID != "" {
		sqlQuery += " AND gpu_id = ?"
		args = append(args, query.GPUID)
	}
	if query.Region != "" {
		sqlQuery += " AND region = ?"
		args = append(args, query.Region)
	}

	pageSize := query.Pagination.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := (query.Pagination.Page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	sqlQuery += " ORDER BY window_start DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	common.Logger.Debug("ClickHouse SLACompliance query start=%v end=%v", start, end)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("sla compliance query failed: %w", err)
	}
	defer rows.Close()

	results := make([]*models.SLAComplianceRow, 0, 50)
	for rows.Next() {
		r := &models.SLAComplianceRow{}
		// clickhouse-go returns Nullable columns as pointers
		var modelID, gpuID, region *string
		var startupSuccessRate, effectiveSuccessRate, noSwapRate, slaScore *float64

		if err := rows.Scan(
			&r.WindowStart, &r.OrchestratorAddress, &r.PipelineID,
			&modelID, &gpuID, &region,
			&r.KnownSessionsCount, &r.StartupSuccessSessions, &r.StartupExcusedSessions,
			&r.StartupUnexcusedSessions, &r.ConfirmedSwappedSessions, &r.InferredSwapSessions, &r.TotalSwappedSessions,
			&r.SessionsEndingInError, &r.ErrorStatusSamples, &r.HealthSignalCoverageRatio,
			&startupSuccessRate, &effectiveSuccessRate, &noSwapRate, &slaScore,
		); err != nil {
			return nil, fmt.Errorf("sla compliance scan failed: %w", err)
		}

		r.ModelID = modelID
		r.GPUID = gpuID
		r.Region = region
		r.StartupSuccessRate = startupSuccessRate
		r.EffectiveSuccessRate = effectiveSuccessRate
		r.NoSwapRate = noSwapRate
		r.SLAScore = slaScore

		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sla compliance rows iteration: %w", err)
	}

	common.Logger.Debug("ClickHouse SLACompliance returned %d rows", len(results))
	return results, nil
}

func (s *ClickhouseStore) SLAComplianceCount(query *models.SLAComplianceQuery) (int, error) {
	if query == nil {
		return 0, errors.New("sla compliance query cannot be nil")
	}

	end := time.Now().UTC()
	period := query.Period
	if period <= 0 {
		period = 24 * time.Hour
	}
	start := end.Add(-period)

	sqlQuery := `SELECT count() FROM v_api_sla_compliance WHERE window_start >= ? AND window_start <= ?`
	args := []interface{}{start, end}

	if query.OrchestratorAddress != "" {
		sqlQuery += " AND orchestrator_address = ?"
		args = append(args, query.OrchestratorAddress)
	}
	if query.PipelineID != "" {
		sqlQuery += " AND pipeline_id = ?"
		args = append(args, query.PipelineID)
	}
	if query.ModelID != "" {
		sqlQuery += " AND model_id = ?"
		args = append(args, query.ModelID)
	}
	if query.GPUID != "" {
		sqlQuery += " AND gpu_id = ?"
		args = append(args, query.GPUID)
	}
	if query.Region != "" {
		sqlQuery += " AND region = ?"
		args = append(args, query.Region)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var total int
	if err := s.db.QueryRowContext(ctx, sqlQuery, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("sla compliance count query failed: %w", err)
	}
	return total, nil
}

// --- Datasets (hard-coded, no view yet) ---

func (s *ClickhouseStore) Datasets(query *models.DatasetsQuery) ([]*models.Dataset, error) {
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

	if query == nil {
		return datasets, nil
	}

	filtered := make([]*models.Dataset, 0, len(datasets))
	for _, d := range datasets {
		if query.Workflow != "" && d.Workflow != query.Workflow {
			continue
		}
		if query.Type != "" && d.Type != query.Type {
			continue
		}
		filtered = append(filtered, d)
	}
	return filtered, nil
}

// --- helpers ---

func envOrDefault(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func protocolFromString(value string) clickhouse.Protocol {
	switch strings.ToLower(value) {
	case "native":
		return clickhouse.Native
	case "http", "https":
		return clickhouse.HTTP
	default:
		return clickhouse.HTTP
	}
}
