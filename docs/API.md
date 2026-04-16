# Analytics API Reference

## Base URL

```
http://localhost:8080
```

## Endpoints

### POST /api/summary

Upload a multi-metric CSV and receive a statistical summary for a filtered metric.

**Request:** Multipart form with:
- `file` (required): CSV file
- `metric` (optional): Metric name to filter. Defaults to first metric in CSV.
- `dimension` (optional): Dimension to filter by.

**Response (200):**

```json
{
  "metric": "revenue",
  "dimension": "",
  "summary": {
    "count": 6,
    "sum": 756.0,
    "min": 80.0,
    "max": 200.0,
    "average": 126.0,
    "median": 115.25,
    "variance": 1808.92,
    "std_dev": 42.53,
    "p90": 190.0,
    "p95": 195.0,
    "first": 120.5,
    "last": 150.0,
    "delta": 29.5,
    "percent_change": 24.48
  },
  "warnings": []
}
```

### POST /api/window-summary

Legacy endpoint. Upload a `timestamp,value` CSV and receive per-window summaries.

**Request:** Multipart form with `file` field.

**Response (200):**

```json
{
  "windows": [
    {
      "window_start": "2026-04-01T10:00:00Z",
      "window_end": "2026-04-01T10:05:00Z",
      "summary": { ... }
    }
  ],
  "warnings": ["line 3: invalid timestamp ..."]
}
```

### POST /api/report

Upload a multi-metric CSV and receive a full analytics report with trend detection and anomaly alerts.

**Request:** Multipart form with:
- `file` (required): CSV file
- `metric` (optional): Metric name to filter
- `dimension` (optional): Dimension to filter by
- `window_size` (optional): Window size (`1m`, `5m`, `15m`, `1h`). Default: `5m`
- `fill_empty_windows` (optional): `true` to fill gaps. Default: `false`

**Response (200):**

```json
{
  "report": {
    "generated_at": "2026-04-01T12:00:00Z",
    "metric": "revenue",
    "dimension": "",
    "window_size": "5m",
    "current_windows": [ ... ],
    "previous_windows": [],
    "overall_summary": { ... },
    "trend": "up",
    "alerts": [
      {
        "type": "spike",
        "window_start": "2026-04-01T10:05:00Z",
        "message": "window average is more than 2x the overall average",
        "value": 350.0,
        "threshold": 252.0
      }
    ]
  },
  "warnings": []
}
```

### GET /healthz

Health check endpoint.

**Response (200):**

```json
{
  "status": "ok"
}
```

### GET /api/meta

API version and metadata.

**Response (200):**

```json
{
  "version": "2.0.0",
  "name": "codeagent-lab analytics API"
}
```

## Error Responses

All errors use a unified JSON structure:

```json
{
  "error": "bad_request",
  "message": "Human-readable description of what went wrong"
}
```

### Error Codes

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `bad_request` | 400 | Invalid input (bad CSV, missing file) |
| `parse_error` | 400 | CSV parsing failed completely |
| `invalid_param` | 400 | Invalid query parameter value |
| `method_not_allowed` | 405 | Wrong HTTP method |

## Warning Structure

Warnings from CSV parsing are returned as structured objects:

```json
{
  "row": 3,
  "message": "invalid timestamp \"not-a-date\""
}
```

Note: The legacy `/api/window-summary` endpoint returns warnings as plain strings for backward compatibility.

## Summary Fields

| Field | Type | Description |
|-------|------|-------------|
| `count` | integer | Number of values |
| `sum` | number | Sum of all values |
| `min` | number | Minimum value |
| `max` | number | Maximum value |
| `average` | number | Arithmetic mean |
| `median` | number | Median value |
| `variance` | number | Population variance |
| `std_dev` | number | Population standard deviation |
| `p90` | number | 90th percentile (linear interpolation) |
| `p95` | number | 95th percentile (linear interpolation) |
| `first` | number | First value in time order |
| `last` | number | Last value in time order |
| `delta` | number | `last - first` |
| `percent_change` | number \| null | Percentage change from first to last. `null` when first is 0. |

## Trend Values

| Value | Condition |
|-------|-----------|
| `up` | Last window average > first window average by >5% |
| `down` | Last window average < first window average by >5% |
| `flat` | Change within ±5% |
| `insufficient_data` | Fewer than 2 windows |

## Alert Types

| Type | Detection Rule |
|------|---------------|
| `spike` | Window average > 2× overall window average |
| `drop` | Window average < 0.5× overall window average |
| `high_variance` | Window std_dev > 2× overall window std_dev |
