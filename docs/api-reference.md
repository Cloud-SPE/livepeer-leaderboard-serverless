# API Reference

All APIs start with `/api/`

### ClickHouse Platform Metrics APIs

These endpoints are backed by ClickHouse views:
- [`/api/gpu/metrics`](#get-apigpumetrics) -> `v_api_gpu_metrics` (or `v_api_gpu_metrics_by_org` when `org` is provided)
- [`/api/network/demand`](#get-apinetworkdemand) -> `v_api_network_demand` (or `v_api_network_demand_by_org` when `org` is provided)
- [`/api/sla/compliance`](#get-apislacompliance) -> `v_api_sla_compliance` (or `v_api_sla_compliance_by_org` when `org` is provided)

#### `GET /api/gpu/metrics`

| Parameter | Description |
|---|---|
| `time_range` | Duration window to query. Default is `24h`, minimum is `1m`, maximum is `72h`. |
| `orchestrator_address` | Optional orchestrator address filter. |
| `pipeline_id` | Optional pipeline filter. |
| `model_id` | Optional model filter. |
| `gpu_id` | Optional GPU ID filter. |
| `region` | Optional region filter. |
| `gpu_model_name` | Optional GPU model filter. |
| `runner_version` | Optional runner version filter. |
| `cuda_version` | Optional CUDA version filter. |
| `org` | Optional organization filter. When present, data is read from `v_api_gpu_metrics_by_org`; otherwise from `v_api_gpu_metrics`. |
| `page` | Page number (1-indexed). Default is `1`. |
| `page_size` | Number of rows per page. Default is `50`, maximum is `500`. |

Response payload shape:
```json
{
  "metrics": [ ... ],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "total_count": 342,
    "total_pages": 7
  }
}
```

| Field Group | Fields |
|---|---|
| Keys/Dimensions | `window_start`, `org` (when org-filtered), `orchestrator_address`, `pipeline_id`, `model_id`, `gpu_id`, `region`, `gpu_model_name`, `gpu_memory_bytes_total`, `runner_version`, `cuda_version` |
| Performance/Latency | `avg_output_fps`, `p95_output_fps`, `fps_jitter_coefficient`, `avg_prompt_to_first_frame_ms`, `avg_startup_latency_ms`, `avg_e2e_latency_ms`, `p95_prompt_to_first_frame_latency_ms`, `p95_startup_latency_ms`, `p95_e2e_latency_ms` |
| Valid Counts | `prompt_to_first_frame_sample_count`, `startup_latency_sample_count`, `e2e_latency_sample_count`, `status_samples`, `error_status_samples` |
| Reliability | `known_sessions_count`, `startup_success_sessions`, `startup_excused_sessions`, `startup_unexcused_sessions`, `confirmed_swapped_sessions`, `inferred_swap_sessions`, `total_swapped_sessions`, `sessions_ending_in_error`, `health_signal_coverage_ratio` |
| Rates | `startup_unexcused_rate`, `swap_rate` |

Contract notes:
- Grain: one row per `(window_start hour, orchestrator_address, pipeline_id, model_id, gpu_id, region)`.
- When `org` is provided, results are scoped to that org and include the `org` field in row payloads.
- `total_swapped_sessions` is the union of `confirmed_swapped_sessions` and `inferred_swap_sessions`.
- `startup_unexcused_rate = startup_unexcused_sessions / known_sessions_count` (returns `0` when denominator is `0`).
- `swap_rate = total_swapped_sessions / known_sessions_count` (returns `0` when denominator is `0`).
- Latency fields are nullable and may be `null` when no valid samples exist for the bucket.

#### `GET /api/network/demand`

| Parameter | Description |
|---|---|
| `interval` | Aggregation interval duration. Default is `15m`, minimum is `1m`, maximum is `48h`. |
| `gateway` | Optional gateway filter. |
| `region` | Optional region filter. |
| `pipeline_id` | Optional pipeline filter. |
| `model_id` | Optional model filter. |
| `org` | Optional organization filter. When present, data is read from `v_api_network_demand_by_org`; otherwise from `v_api_network_demand`. |
| `page` | Page number (1-indexed). Default is `1`. |
| `page_size` | Number of rows per page. Default is `50`, maximum is `500`. |

Response payload shape:
```json
{
  "demand": [ ... ],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "total_count": 342,
    "total_pages": 7
  }
}
```

| Field Group | Fields |
|---|---|
| Keys/Dimensions | `window_start`, `org` (when org-filtered), `gateway`, `region`, `pipeline_id`, `model_id` |
| Demand/Capacity | `sessions_count`, `total_minutes`, `known_sessions_count`, `served_sessions`, `unserved_sessions`, `total_demand_sessions` |
| Reliability | `startup_unexcused_sessions`, `confirmed_swapped_sessions`, `inferred_swap_sessions`, `total_swapped_sessions`, `sessions_ending_in_error`, `error_status_samples`, `health_signal_coverage_ratio`, `startup_success_rate`, `effective_success_rate` |
| Economics | `ticket_face_value_eth` |

Contract notes:
- Grain: one row per `(window_start hour, gateway, region, pipeline_id, model_id)`.
- When `org` is provided, results are scoped to that org and include the `org` field in row payloads.
- `total_demand_sessions = served_sessions + unserved_sessions`.
- `total_swapped_sessions` is the union of `confirmed_swapped_sessions` and `inferred_swap_sessions`.
- `startup_success_rate` is startup-only success and tracks startup contract reliability.
- `effective_success_rate` is effective output viability (not startup-only) and can include startup-unexcused and no-output behaviors.

#### `GET /api/sla/compliance`

| Parameter | Description |
|---|---|
| `period` | Duration window to query. Default is `24h`, minimum is `1h`, maximum is `30d`. |
| `orchestrator_address` | Optional orchestrator address filter. |
| `pipeline_id` | Optional pipeline filter. |
| `model_id` | Optional model filter. |
| `gpu_id` | Optional GPU ID filter. |
| `region` | Optional region filter. |
| `org` | Optional organization filter. When present, data is read from `v_api_sla_compliance_by_org`; otherwise from `v_api_sla_compliance`. |
| `page` | Page number (1-indexed). Default is `1`. |
| `page_size` | Number of rows per page. Default is `50`, maximum is `500`. |

Response payload shape:
```json
{
  "compliance": [ ... ],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "total_count": 342,
    "total_pages": 7
  }
}
```

| Field Group | Fields |
|---|---|
| Keys/Dimensions | `window_start`, `org` (when org-filtered), `orchestrator_address`, `pipeline_id`, `model_id`, `gpu_id`, `region` |
| Reliability | `known_sessions_count`, `startup_success_sessions`, `startup_excused_sessions`, `startup_unexcused_sessions`, `confirmed_swapped_sessions`, `inferred_swap_sessions`, `total_swapped_sessions`, `sessions_ending_in_error`, `error_status_samples`, `health_signal_coverage_ratio`, `startup_success_rate` |
| SLA Scores | `effective_success_rate`, `no_swap_rate`, `sla_score` |

Contract notes:
- Grain: one row per attributed `(window_start hour, orchestrator_address, pipeline_id, model_id, gpu_id, region)`.
- When `org` is provided, results are scoped to that org and include the `org` field in row payloads.
- `total_swapped_sessions` is the union of `confirmed_swapped_sessions` and `inferred_swap_sessions`.
- `startup_success_rate` is startup-only success and should be used for startup SLA interpretation.
- `effective_success_rate` is effective output viability aligned with demand semantics.
- `no_swap_rate = 1 - (total_swapped_sessions / known_sessions_count)` (returns `0` when denominator is `0`).
- `sla_score` weighting is defined by the upstream view contract and may change with metric hardening.
- Tail artifact suppression is stricter for true rollover cleanup hours, while same-hour failed/no-output sessions are retained for observability.

GPU row behavior note:
- True rollover tail artifacts are filtered using current-hour + previous-hour no-work guards; same-hour failed/no-output attempts are retained.

#### `GET /api/aggregated_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>&until=<timestamp>`

| Parameter         | Description                                                                                                                                                           |
|-------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `orchestrator`     | The orchestrator to get aggregated stats for. If `orchestrator` is not provided, the response will include aggregated scores for all orchestrators.                    |
| `region`          | The region to get aggregated stats for. If `region` is not provided, all regions will be returned in the response. Region must be a registered region in the database.  For example `"FRA", "MDW", "SIN"`.         |
| `since`           | The timestamp to evaluate the query from. If neither `since` nor `until` are provided, it will return the results starting from the time period specified by the environment variable `START_TIME_WINDOW` or its default.                                |
| `until`           | If `until` is provided but `since` is not, it will return all results before the `until` timestamp.                                                                     |


#### Transcoding Response 

```
{
   "<orchAddr>": {
    "MDW": {
      "total_score": 5.5,
      "latency_score": 6.01,
      "success_rate": 91.5
    },
    "FRA": {
    	"total_score": 2.5,
      "latency_score": 2.5,
      "success_rate": 100
    },
    "SIN": {
    	"total_score": 6.6,
      "latency_score": 7.10
			"success_rate": 93
    }
  },
   "<orchAddr2>": {
  		...
  },
	...
}
```

#### AI Request 

This is similar to [`GET /api/aggregated_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>&until=<timestamp>`](#get-apiaggregated_statsorchestratororchaddrregionregion_codesincetimestampuntiltimestamp), with the added parameters documented below.

#### `GET /api/aggregated_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>&until=<timestamp>&model=<model>&pipeline=<pipeline>`

| Parameter         | Description                                                                                                                          |
|-------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| `model`           | The model to check stats for. Model is a required field. If `model` is not provided, you will get an HTTP status code 400 (Bad Request). |
| `pipeline`        | The pipeline to check stats for. Pipeline is a required field. If `pipeline` is not provided, you will get an HTTP status code 400 (Bad Request). |


#### AI Response 

```
{
  "0x10742714f33f3d804e3fa489618b5c3ca12a6df7": {
    "FRA": {
      "success_rate": 1,
      "round_trip_score": 0.742521971754293,
      "score": 0.909882690114002
    },
    "LAX": {
      "success_rate": 1,
      "round_trip_score": 0.844420265972075,
      "score": 0.945547093090226
    },
    "MDW": {
      "success_rate": 1,
      "round_trip_score": 0.797933017387645,
      "score": 0.929276556085676
    }
  }
}
```

#### `GET /api/raw_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>`

| Parameter         | Description                                                                                                                           |
|-------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| `orchestrator`     | The orchestrator's address to check raw stats for. If no parameter for `orchestrator` is provided, the request will return `400 Bad Request`. |
| `region`          | The region to check stats for. If `region` is not provided, all regions will be returned in the response.                               |
| `since`           | The timestamp to evaluate the query from. If `since` is not provided, it will return the results starting from time period specified by the environment variable `START_TIME_WINDOW` or its default.                 |

#### Transcoding Response

For each region return an array of the metrics from the 'metrics gathering' section as a "raw dump"

```
{
 "FRA": [
    {
	    "timestamp": number,
        "segments_sent": number,
        "segments_received": number,
        "success_rate": number,
        "seg_duration": number,
        "upload_time": number,
        "download_time": number,
        "transcode_time": number,
        "round_trip_time": number,
        "errors": Array
      }
   ],
   "MDW": [...],
   "SIN": [...]
}
```

#### AI Request 

This is similar to [`GET /api/raw_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>`](#get-apiraw_statsorchestratororchaddrregionregion_codesincetimestamp), with the added parameters documented below.

#### `GET /api/raw_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>&until=<timestamp>&model=<model>&pipeline=<pipeline>`

| Parameter         | Description                                                                                                                               |
|-------------------|-------------------------------------------------------------------------------------------------------------------------------------------|
| `model`           | The model to check stats for. Model is an optional field.                                                                                 |
| `pipeline`        | The pipeline to check stats for. Pipeline is a required field. If pipeline is not provided, the request falls back to the Transcoding behavior described above. |

#### AI Response 

```
{
  "FRA": [
    {
      "region": "FRA",
      "orchestrator": "0x10742714f33f3d804e3fa489618b5c3ca12a6df7",
      "success_rate": 1,
      "round_trip_time": 7.236450406,
      "errors": [],
      "timestamp": 1726864722,
      "model": "stabilityai/stable-video-diffusion-img2vid-xt-1-1",
      "model_is_warm": true,
      "pipeline": "Image to video",
      "input_parameters": "{\"fps\":8,\"height\":256,\"model_id\":\"stabilityai/stable-video-diffusion-img2vid-xt-1-1\",\"motion_bucket_id\":127,\"noise_aug_strength\":0.065,\"width\":256}",
      "response_payload": "{\"images\":[{\"nsfw\":false,\"seed\":1384909895,\"url\":\"/stream/112b6ad2/772ed708.mp4\"}]}\n"
    },
    {
      "region": "FRA",
      "orchestrator": "0x10742714f33f3d804e3fa489618b5c3ca12a6df7",
      "success_rate": 1,
      "round_trip_time": 7.333097532,
      "errors": [],
      "timestamp": 1726857456,
      "model": "stabilityai/stable-video-diffusion-img2vid-xt-1-1",
      "model_is_warm": true,
      "pipeline": "Image to video",
      "input_parameters": "{\"fps\":8,\"height\":256,\"model_id\":\"stabilityai/stable-video-diffusion-img2vid-xt-1-1\",\"motion_bucket_id\":127,\"noise_aug_strength\":0.065,\"width\":256}",
      "response_payload": "{\"images\":[{\"nsfw\":false,\"seed\":2533618378,\"url\":\"/stream/774f96b9/105469a0.mp4\"}]}\n"
    }
  ],
  "LAX": [
    {
      "region": "LAX",
      "orchestrator": "0x10742714f33f3d804e3fa489618b5c3ca12a6df7",
      "success_rate": 1,
      "round_trip_time": 4.110541139,
      "errors": [],
      "timestamp": 1726866030,
      "model": "stabilityai/stable-video-diffusion-img2vid-xt-1-1",
      "model_is_warm": true,
      "pipeline": "Image to video",
      "input_parameters": "{\"fps\":8,\"height\":256,\"model_id\":\"stabilityai/stable-video-diffusion-img2vid-xt-1-1\",\"motion_bucket_id\":127,\"noise_aug_strength\":0.065,\"width\":256}",
      "response_payload": "{\"images\":[{\"nsfw\":false,\"seed\":689122349,\"url\":\"/stream/bafdeb1f/5f1f1fee.mp4\"}]}\n"
    },
  ]
}
```


#### POST `/api/post_stats`

This accepts a JSON encododed Stats object that maps to the `Stats` struct below.

```
// Stats are the raw stats per test stream
type Stats struct {
	Region        string  `json:"region" bson:"-"`
	Orchestrator  string  `json:"orchestrator" bson:"orchestrator"`
	SuccessRate   float64 `json:"success_rate" bson:"success_rate"`
	RoundTripTime float64 `json:"round_trip_time" bson:"round_trip_time"`
	Errors        []Error `json:"errors" bson:"errors"`
	Timestamp     int64   `json:"timestamp" bson:"timestamp"`

	// Transcoding stats fields
	SegDuration      float64 `json:"seg_duration,omitempty" bson:"seg_duration,omitempty"`
	SegmentsSent     int     `json:"segments_sent,omitempty" bson:"segments_sent,omitempty"`
	SegmentsReceived int     `json:"segments_received,omitempty" bson:"segments_received,omitempty"`
	UploadTime       float64 `json:"upload_time,omitempty" bson:"upload_time,omitempty"`
	DownloadTime     float64 `json:"download_time,omitempty" bson:"download_time,omitempty"`
	TranscodeTime    float64 `json:"transcode_time,omitempty" bson:"transcode_time,omitempty"`

	// AI stats fields
	Model           string `json:"model,omitempty" bson:"model,omitempty"`
	ModelIsWarm     bool   `json:"model_is_warm,omitempty" bson:"model_is_warm,omitempty"`
	Pipeline        string `json:"pipeline,omitempty" bson:"pipeline,omitempty"`
	InputParameters string `json:"input_parameters,omitempty" bson:"input_parameters,omitempty"`
	ResponsePayload string `json:"response_payload,omitempty" bson:"response_payload,omitempty"`
}
```

#### `GET /api/pipelines?region=<region_code>&since=<timestamp>&until=<timestamp>`

| Parameter         | Description                                                                                                                                                      |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `region`          | The region to check stats for. If `region` is not provided, all pipelines will be returned in the response.                                                       |
| `since`           | The timestamp to evaluate the query from. If neither `since` nor `until` are provided, it will return the results starting from the time period specified by the environment variable `START_TIME_WINDOW` or its default.                           |
| `until`           | If `until` is provided but `since` is not, it will return all results before the `until` timestamp.                                                                |


This endpoint outputs the pipelines and models in JSON format.

#### Response 
```
{
  "pipelines": [
    {
      "id": "Audio to text",
      "models": [
        "openai/whisper-large-v3"
      ],
      "regions": [
        "FRA",
        "LAX",
        "MDW"
      ]
    },
    {
      "id": "Image to image",
      "models": [
        "ByteDance/SDXL-Lightning",
        "timbrooks/instruct-pix2pix"
      ],
      "regions": [
        "FRA",
        "LAX",
        "MDW"
      ]
    }
    ...
  ]
}
```


#### `GET /api/regions`

This endpoint outputs the regions in JSON format.  It does not take any parameters.

#### Response 
```
{
  "regions": [
    {
      "id": "TOR",
      "name": "Toronto",
      "type": "transcoding"
    },
    {
      "id": "HND",
      "name": "Tokyo",
      "type": "transcoding"
    },
    {
      "id": "SYD",
      "name": "Sydney",
      "type": "transcoding"
    },
    {
      "id": "STO",
      "name": "Stockholm",
      "type": "transcoding"
    },
  ]
}
```
