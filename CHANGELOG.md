# Changelog

## Breaking Changes

### API parameter renames — ClickHouse endpoints

The three ClickHouse-backed endpoints now use a unified `window` parameter in place of their individual time-range parameters.

| Endpoint | Old parameter | New parameter | Old default | New default |
|---|---|---|---|---|
| `GET /api/gpu/metrics` | `time_range` | `window` | `24h` | `24h` |
| `GET /api/network/demand` | `interval` | `window` | `15m` (effective: `3h`) | `3h` |
| `GET /api/sla/compliance` | `period` | `window` | `24h` | `24h` |

**Migration**: Replace the old parameter name with `window` in your query strings.

```
# Before
GET /api/gpu/metrics?time_range=6h
GET /api/network/demand?interval=15m
GET /api/sla/compliance?period=7d

# After
GET /api/gpu/metrics?window=6h
GET /api/network/demand?window=3h
GET /api/sla/compliance?window=7d
```

### Semantic fix — `/api/network/demand`

The old `interval` parameter had a hidden `× 12` multiplier: `start = end - (interval × 12)`. This meant `interval=1h` returned 12 hours of data rather than 1 hour.

The new `window` parameter is a direct lookback duration with no multiplier: `window=1h` returns exactly the last 1 hour of data.

**Impact on `/api/network/demand` defaults**: The old default `interval=15m` produced an effective 3-hour window (`15m × 12`). The new default `window=3h` preserves this observable behavior.

**Maximum window for `/api/network/demand`**: Changed from `48h` (old `interval` max, effective `576h`) to `30d`, consistent with `/api/sla/compliance`.
