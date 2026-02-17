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
		window_start, orchestrator_address, pipeline, pipeline_id,
		model_id, gpu_id, region,
		avg_output_fps, p95_output_fps, jitter_coeff_fps, status_samples
	FROM v_api_gpu_metrics
	WHERE window_start >= ? AND window_start <= ?`

	args := []interface{}{start, end}

	if query.OrchestratorAddress != "" {
		sqlQuery += " AND orchestrator_address = ?"
		args = append(args, query.OrchestratorAddress)
	}
	if query.Pipeline != "" {
		sqlQuery += " AND pipeline = ?"
		args = append(args, query.Pipeline)
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

	sqlQuery += " ORDER BY window_start DESC LIMIT 200"

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
		// clickhouse-go returns Nullable columns as pointers (*string, *float64)
		var modelID, gpuID, region *string
		var jitterCoeff *float64
		var p95 float32

		if err := rows.Scan(
			&m.WindowStart, &m.OrchestratorAddress, &m.Pipeline, &m.PipelineID,
			&modelID, &gpuID, &region,
			&m.AvgOutputFPS, &p95, &jitterCoeff, &m.StatusSamples,
		); err != nil {
			return nil, fmt.Errorf("gpu metrics scan failed: %w", err)
		}

		m.ModelID = modelID
		m.GPUID = gpuID
		m.Region = region
		m.JitterCoeffFPS = jitterCoeff
		m.P95OutputFPS = p95

		results = append(results, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gpu metrics rows iteration: %w", err)
	}

	common.Logger.Debug("ClickHouse GPUMetrics returned %d rows", len(results))
	return results, nil
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
		window_start, gateway, region, pipeline, pipeline_id,
		active_sessions, active_streams, avg_output_fps
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
	if query.Pipeline != "" {
		sqlQuery += " AND pipeline = ?"
		args = append(args, query.Pipeline)
	}
	if query.PipelineID != "" {
		sqlQuery += " AND pipeline_id = ?"
		args = append(args, query.PipelineID)
	}

	sqlQuery += " ORDER BY window_start DESC LIMIT 200"

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
		var region *string

		if err := rows.Scan(
			&r.WindowStart, &r.Gateway, &region, &r.Pipeline, &r.PipelineID,
			&r.ActiveSessions, &r.ActiveStreams, &r.AvgOutputFPS,
		); err != nil {
			return nil, fmt.Errorf("network demand scan failed: %w", err)
		}

		r.Region = region
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("network demand rows iteration: %w", err)
	}

	common.Logger.Debug("ClickHouse NetworkDemand returned %d rows", len(results))
	return results, nil
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
		window_start, orchestrator_address, pipeline, pipeline_id,
		model_id, gpu_id, region,
		known_sessions, unexcused_sessions, swapped_sessions,
		success_ratio, no_swap_ratio
	FROM v_api_sla_compliance
	WHERE window_start >= ? AND window_start <= ?`

	args := []interface{}{start, end}

	if query.OrchestratorAddress != "" {
		sqlQuery += " AND orchestrator_address = ?"
		args = append(args, query.OrchestratorAddress)
	}
	if query.Pipeline != "" {
		sqlQuery += " AND pipeline = ?"
		args = append(args, query.Pipeline)
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

	sqlQuery += " ORDER BY window_start DESC LIMIT 200"

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
		var successRatio, noSwapRatio *float64

		if err := rows.Scan(
			&r.WindowStart, &r.OrchestratorAddress, &r.Pipeline, &r.PipelineID,
			&modelID, &gpuID, &region,
			&r.KnownSessions, &r.UnexcusedSessions, &r.SwappedSessions,
			&successRatio, &noSwapRatio,
		); err != nil {
			return nil, fmt.Errorf("sla compliance scan failed: %w", err)
		}

		r.ModelID = modelID
		r.GPUID = gpuID
		r.Region = region
		r.SuccessRatio = successRatio
		r.NoSwapRatio = noSwapRatio

		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sla compliance rows iteration: %w", err)
	}

	common.Logger.Debug("ClickHouse SLACompliance returned %d rows", len(results))
	return results, nil
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
