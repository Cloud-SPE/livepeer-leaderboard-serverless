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
	db   *sql.DB
	mock *MockStore
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

	return &ClickhouseStore{
		db:   db,
		mock: NewMockStore(),
	}, nil
}

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

	sqlQuery := `
		WITH
			arrayJoin(JSONExtractArrayRaw(data)) AS orch_raw,
			arrayElement(JSONExtractArrayRaw(orch_raw, 'hardware'), 1) AS hw,
			JSONExtractRaw(hw, 'gpu_info') AS gpu_info_raw,
			JSONExtractRaw(gpu_info_raw, '0') AS gpu0
		SELECT
			event_timestamp,
			JSONExtractString(orch_raw, 'address') AS orch_addr,
			JSONExtractString(orch_raw, 'local_address') AS local_addr,
			JSONExtractString(orch_raw, 'orch_uri') AS orch_uri,
			JSONExtractString(hw, 'pipeline') AS pipeline,
			JSONExtractString(gpu0, 'id') AS gpu_id,
			JSONExtractFloat(gpu0, 'memory_total') AS mem_total,
			JSONExtractFloat(gpu0, 'memory_free') AS mem_free
		FROM streaming_events
		WHERE type = 'network_capabilities'
			AND event_timestamp >= ?
			AND event_timestamp <= ?
	`

	args := []interface{}{start, end}
	if query.Workflow != "" {
		sqlQuery += " AND JSONExtractString(hw, 'pipeline') = ?"
		args = append(args, query.Workflow)
	}
	if query.GPUId != "" {
		sqlQuery += " AND JSONExtractString(gpu0, 'id') = ?"
		args = append(args, query.GPUId)
	}
	if query.OrchestratorWallet != "" {
		sqlQuery += ` AND (
			JSONExtractString(orch_raw, 'address') = ?
			OR JSONExtractString(orch_raw, 'local_address') = ?
		)`
		args = append(args, query.OrchestratorWallet, query.OrchestratorWallet)
	}
	sqlQuery += " ORDER BY event_timestamp DESC LIMIT 200"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	common.Logger.Debug("ClickHouse GPUMetrics query start=%v end=%v", start, end)
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		common.Logger.Warn("ClickHouse GPUMetrics query failed, using mock: %v", err)
		return s.mock.GPUMetrics(query)
	}
	defer rows.Close()

	results := make([]*models.GPUMetric, 0, 50)
	for rows.Next() {
		var (
			timestamp time.Time
			orchAddr  string
			localAddr string
			pipeline  string
			gpuID     string
			memTotal  float64
			memFree   float64
		)
		var orchURI string
		if err := rows.Scan(&timestamp, &orchAddr, &localAddr, &orchURI, &pipeline, &gpuID, &memTotal, &memFree); err != nil {
			return s.mock.GPUMetrics(query)
		}

		orchestratorWallet := firstNonEmpty(orchAddr, localAddr, query.OrchestratorWallet)
		workflow := firstNonEmpty(query.Workflow, pipeline, "inference")
		region := firstNonEmpty(query.Region, "global")

		memUsed := memTotal - memFree
		utilization := 0.0
		if memTotal > 0 {
			utilization = (memUsed / memTotal) * 100.0
		}

		results = append(results, &models.GPUMetric{
			OrchestratorWallet: orchestratorWallet,
			GPUId:              gpuID,
			Region:             region,
			Workflow:           workflow,
			Timestamp:          timestamp,
			UtilizationPct:     utilization,
			MemoryUsedMB:       bytesToMB(memUsed),
			MemoryTotalMB:      bytesToMB(memTotal),
			TemperatureC:       0,
			PowerWatts:         0,
			SuccessRate:        1.0,
			ErrorRate:          0.0,
		})
	}

	if len(results) == 0 {
		common.Logger.Debug("ClickHouse GPUMetrics returned 0 rows, using mock")
		return s.mock.GPUMetrics(query)
	}
	common.Logger.Debug("ClickHouse GPUMetrics returned %d rows", len(results))
	return results, nil
}

func (s *ClickhouseStore) NetworkDemand(query *models.NetworkDemandQuery) (*models.NetworkDemand, error) {
	if query == nil {
		return nil, errors.New("network demand query cannot be nil")
	}

	interval := query.Interval
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	end := time.Now().UTC().Truncate(interval)
	start := end.Add(-interval * 12)

	streamCounts, err := s.queryStreamCounts(start, end, interval, query.Gateway)
	if err != nil {
		common.Logger.Warn("ClickHouse NetworkDemand stream query failed, using mock: %v", err)
		return s.mock.NetworkDemand(query)
	}
	inferCounts, err := s.queryInferenceCounts(start, end, interval, query.Gateway, query.Workflow)
	if err != nil {
		common.Logger.Warn("ClickHouse NetworkDemand inference query failed, using mock: %v", err)
		return s.mock.NetworkDemand(query)
	}

	points := make([]models.NetworkDemandPoint, 0, 12)
	for i := 11; i >= 0; i-- {
		bucket := end.Add(-time.Duration(i) * interval)
		streamMinutes := float64(streamCounts[bucket]) * interval.Minutes()
		inferMinutes := float64(inferCounts[bucket]) * interval.Minutes()
		points = append(points, models.NetworkDemandPoint{
			Timestamp:     bucket,
			StreamMinutes: streamMinutes,
			InferMinutes:  inferMinutes,
		})
	}

	return &models.NetworkDemand{
		Gateway:  firstNonEmpty(query.Gateway, "public"),
		Region:   firstNonEmpty(query.Region, "global"),
		Workflow: firstNonEmpty(query.Workflow, "inference"),
		Interval: interval.String(),
		Points:   points,
	}, nil
}

func (s *ClickhouseStore) SLACompliance(query *models.SLAComplianceQuery) (*models.SLACompliance, error) {
	if query == nil {
		return nil, errors.New("sla compliance query cannot be nil")
	}

	end := time.Now().UTC()
	start := end.Add(-query.Period)

	sqlQuery := `
		SELECT
			countIf(JSONExtractString(data, 'state') = 'ONLINE') AS online,
			count() AS total
		FROM streaming_events
		WHERE type = 'ai_stream_status'
			AND event_timestamp >= ?
			AND event_timestamp <= ?
	`
	args := []interface{}{start, end}
	if query.OrchestratorID != "" {
		sqlQuery += " AND JSONExtractString(JSONExtractRaw(data, 'orchestrator_info'), 'address') = ?"
		args = append(args, query.OrchestratorID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var online uint64
	var total uint64
	common.Logger.Debug("ClickHouse SLACompliance query start=%v end=%v", start, end)
	if err := s.db.QueryRowContext(ctx, sqlQuery, args...).Scan(&online, &total); err != nil {
		common.Logger.Warn("ClickHouse SLACompliance query failed, using mock: %v", err)
		return s.mock.SLACompliance(query)
	}
	common.Logger.Debug("ClickHouse SLACompliance counts online=%d total=%d", online, total)

	score := 100.0
	if total > 0 {
		score = (float64(online) / float64(total)) * 100.0
	}

	return &models.SLACompliance{
		OrchestratorID: firstNonEmpty(query.OrchestratorID, "orch-0"),
		Period:         query.Period.String(),
		Score:          score,
		WindowStart:    start,
		WindowEnd:      end,
	}, nil
}

func (s *ClickhouseStore) Datasets(query *models.DatasetsQuery) ([]*models.Dataset, error) {
	return s.mock.Datasets(query)
}

func (s *ClickhouseStore) queryStreamCounts(start, end time.Time, interval time.Duration, gateway string) (map[time.Time]uint64, error) {
	sqlQuery := `
		SELECT
			toStartOfInterval(event_timestamp, toIntervalSecond(?)) AS bucket,
			countDistinct(JSONExtractString(data, 'request_id')) AS count
		FROM streaming_events
		WHERE type = 'stream_trace'
			AND JSONExtractString(data, 'type') = 'gateway_receive_stream_request'
			AND event_timestamp >= ?
			AND event_timestamp <= ?
	`
	args := []interface{}{int(interval.Seconds()), start, end}
	if gateway != "" {
		sqlQuery += " AND gateway = ?"
		args = append(args, gateway)
	}
	sqlQuery += " GROUP BY bucket ORDER BY bucket"

	return s.queryBucketCounts(sqlQuery, args...)
}

func (s *ClickhouseStore) queryInferenceCounts(start, end time.Time, interval time.Duration, gateway, workflow string) (map[time.Time]uint64, error) {
	sqlQuery := `
		SELECT
			toStartOfInterval(event_timestamp, toIntervalSecond(?)) AS bucket,
			countDistinct(JSONExtractString(data, 'stream_id')) AS count
		FROM streaming_events
		WHERE type = 'ai_stream_status'
			AND event_timestamp >= ?
			AND event_timestamp <= ?
	`
	args := []interface{}{int(interval.Seconds()), start, end}
	if gateway != "" {
		sqlQuery += " AND gateway = ?"
		args = append(args, gateway)
	}
	if workflow != "" {
		sqlQuery += " AND JSONExtractString(data, 'pipeline') = ?"
		args = append(args, workflow)
	}
	sqlQuery += " GROUP BY bucket ORDER BY bucket"

	return s.queryBucketCounts(sqlQuery, args...)
}

func (s *ClickhouseStore) queryBucketCounts(sqlQuery string, args ...interface{}) (map[time.Time]uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	common.Logger.Debug("ClickHouse bucket query: %s", strings.Join(strings.Fields(sqlQuery), " "))
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[time.Time]uint64)
	for rows.Next() {
		var bucket time.Time
		var count uint64
		if err := rows.Scan(&bucket, &count); err != nil {
			return nil, err
		}
		results[bucket] = count
	}
	return results, nil
}

func envOrDefault(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func bytesToMB(value float64) float64 {
	if value <= 0 {
		return 0
	}
	return value / (1024.0 * 1024.0)
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
