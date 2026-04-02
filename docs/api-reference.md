# API Reference

All APIs start with `/api/`

#### `GET /api/health`

Returns the readiness status of Postgres. Returns `200` if healthy, `503` otherwise.

Response payload shape:
```json
{ "postgres": { "ok": true } }
```

When unhealthy:
```json
{ "postgres": { "ok": false, "error": "POSTGRES env var not configured" } }
```

The `error` field is only present when `ok` is `false`.

---

#### `GET /api/top_ai_score?orchestrator=<orchAddr>`

Returns the top regional AI score for a given orchestrator — the best-performing region, model, and pipeline combination based on aggregated stats.

| Parameter | Description |
|---|---|
| `orchestrator` | The orchestrator address to look up. If not provided, returns `{}`. |

Response payload shape:
```json
{
  "orchestrator": "0x...",
  "region": "MDW",
  "value": 0.94,
  "model": "stabilityai/stable-diffusion-xl-base-1.0",
  "pipeline": "text-to-image"
}
```

Returns `{}` if no stats are found.

---

#### `GET /api/aggregated_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>&until=<timestamp>`

| Parameter | Description |
|---|---|
| `orchestrator` | The orchestrator to get aggregated stats for. If not provided, results include all orchestrators. |
| `region` | The region to get aggregated stats for. If not provided, all regions are returned. Must be a registered region in the database (e.g. `FRA`, `MDW`, `SIN`). |
| `since` | The timestamp to evaluate the query from. If neither `since` nor `until` are provided, defaults to the period set by the `START_TIME_WINDOW` env var. |
| `until` | If `until` is provided but `since` is not, returns all results before the `until` timestamp. |

#### Transcoding Response

```json
{
  "<orchAddr>": {
    "MDW": {
      "score": 5.5,
      "round_trip_score": 6.01,
      "success_rate": 0.915
    },
    "FRA": {
      "score": 2.5,
      "round_trip_score": 2.5,
      "success_rate": 1.0
    }
  }
}
```

#### AI Request

Same as above with additional required parameters.

#### `GET /api/aggregated_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>&until=<timestamp>&model=<model>&pipeline=<pipeline>`

| Parameter | Description |
|---|---|
| `model` | The model to check stats for. Required when using AI job type — returns `400` if `pipeline` is provided but `model` is not. |
| `pipeline` | The pipeline to check stats for. Required when using AI job type — returns `400` if `model` is provided but `pipeline` is not. |

#### AI Response

```json
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
    }
  }
}
```

---

#### `GET /api/raw_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>`

| Parameter | Description |
|---|---|
| `orchestrator` | The orchestrator address to check raw stats for. **Required** — returns `400` if not provided. |
| `region` | The region to check stats for. If not provided, all regions are returned. |
| `since` | The timestamp to evaluate the query from. If not provided, defaults to the period set by `START_TIME_WINDOW`. |

#### Transcoding Response

```json
{
  "FRA": [
    {
      "timestamp": 1726864722,
      "segments_sent": 10,
      "segments_received": 10,
      "success_rate": 1.0,
      "seg_duration": 2.0,
      "upload_time": 0.1,
      "download_time": 0.05,
      "transcode_time": 0.3,
      "round_trip_time": 0.45,
      "errors": []
    }
  ]
}
```

#### AI Request

Same as above with additional optional parameters.

#### `GET /api/raw_stats?orchestrator=<orchAddr>&region=<region_code>&since=<timestamp>&until=<timestamp>&model=<model>&pipeline=<pipeline>`

| Parameter | Description |
|---|---|
| `model` | The model to check stats for. Optional. |
| `pipeline` | The pipeline to check stats for. Optional. If not provided, falls back to transcoding behavior. |

#### AI Response

```json
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
      "input_parameters": "{\"fps\":8,\"height\":256}",
      "response_payload": "{\"images\":[{\"nsfw\":false,\"seed\":1384909895}]}"
    }
  ]
}
```

---

#### POST `/api/post_stats`

Accepts a JSON-encoded `Stats` object. Requires an `Authorization` header containing an HMAC-SHA256 hex signature of the request body, computed using the `SECRET` env var.

```json
{
  "region": "FRA",
  "orchestrator": "0x...",
  "success_rate": 1.0,
  "round_trip_time": 0.45,
  "errors": [],
  "timestamp": 1726864722,

  "seg_duration": 2.0,
  "segments_sent": 10,
  "segments_received": 10,
  "upload_time": 0.1,
  "download_time": 0.05,
  "transcode_time": 0.3,

  "model": "stabilityai/stable-diffusion-xl-base-1.0",
  "model_is_warm": true,
  "pipeline": "text-to-image",
  "input_parameters": "{}",
  "response_payload": "{}"
}
```

Transcoding-only fields (`seg_duration`, `segments_sent`, `segments_received`, `upload_time`, `download_time`, `transcode_time`) and AI-only fields (`model`, `model_is_warm`, `pipeline`, `input_parameters`, `response_payload`) are optional.

Returns `200 ok` on success, `400` for invalid JSON or unrecognised region, `403` if the request cannot be authenticated.

---

#### `GET /api/pipelines?region=<region_code>&since=<timestamp>&until=<timestamp>`

| Parameter | Description |
|---|---|
| `region` | Optional. If not provided, all pipelines are returned. |
| `since` | Optional. Defaults to the period set by `START_TIME_WINDOW`. |
| `until` | Optional. |

#### Response

```json
{
  "pipelines": [
    {
      "id": "Audio to text",
      "models": ["openai/whisper-large-v3"],
      "regions": ["FRA", "LAX", "MDW"]
    },
    {
      "id": "Image to image",
      "models": ["ByteDance/SDXL-Lightning", "timbrooks/instruct-pix2pix"],
      "regions": ["FRA", "LAX", "MDW"]
    }
  ]
}
```

---

#### `GET /api/regions`

No parameters. Returns all registered regions.

#### Response

```json
{
  "regions": [
    { "id": "TOR", "name": "Toronto", "type": "transcoding" },
    { "id": "HND", "name": "Tokyo", "type": "transcoding" },
    { "id": "FRA", "name": "Frankfurt", "type": "ai" }
  ]
}
```
