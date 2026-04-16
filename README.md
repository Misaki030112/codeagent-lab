# codeagent-lab

Cross-language analytics report platform — a lab repository for CodeAgent integration testing.

## Overview

This repository implements a cross-language analytics toolkit across Go, Python, and TypeScript, sharing a unified **analytics report contract**. It supports multi-metric CSV input, configurable window aggregation, trend detection, anomaly alerting, and dashboard-ready data transforms.

The shared JSON schema is defined in `fixtures/schema/report_schema.json` and enforced through cross-language contract tests consuming shared fixtures.

## Layout

| Directory | Language | Purpose |
|-----------|----------|---------|
| `go/` | Go | Stats engine, multi-metric CSV parsing, window aggregation, report generation, HTTP API |
| `python/` | Python | Stats module, multi-command CLI, report generation |
| `web/` | TypeScript | Analytics types, formatting, transforms, query builder, string helpers |
| `fixtures/` | JSON/CSV | Shared test fixtures and expected golden outputs |
| `docs/` | Markdown | API docs, test prompts, manual verification steps |

## Quick Start

### Go — run tests

```bash
cd go && go test ./...
```

### Go — run HTTP server

```bash
cd go && go run ./cmd/server
```

### Python — run tests

```bash
cd python && python3 -m unittest discover -s . -p "test_*.py"
```

### TypeScript — run tests

```bash
cd web && npm install && npx jest
```

### TypeScript — type check

```bash
cd web && npx tsc --noEmit
```

## CSV Input Format

### Multi-metric format (primary)

```csv
timestamp,metric,value,dimension,source
2026-04-01T10:00:00Z,revenue,120.5,cn,ads
2026-04-01T10:01:00Z,revenue,80.0,us,organic
2026-04-01T10:02:00Z,latency_ms,240.0,cn,api
```

| Column | Required | Description |
|--------|----------|-------------|
| `timestamp` | Yes | RFC 3339 format (UTC recommended) |
| `metric` | Yes | Metric name (e.g., `revenue`, `latency_ms`) |
| `value` | Yes | Floating-point number |
| `dimension` | No | Grouping dimension (e.g., region, country) |
| `source` | No | Data source label |

### Legacy format (backward compatible)

```csv
timestamp,value
2026-04-01T10:00:00Z,10
```

## Summary Schema (Shared Contract)

All three languages produce identical summary objects with 14 fields:

```json
{
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
}
```

| Field | Description |
|-------|-------------|
| `count` | Number of values |
| `sum` | Sum of all values |
| `min` / `max` | Minimum / maximum value |
| `average` | Arithmetic mean |
| `median` | Median value |
| `variance` | Population variance |
| `std_dev` | Population standard deviation |
| `p90` / `p95` | 90th / 95th percentile (linear interpolation) |
| `first` / `last` | First / last value in time order |
| `delta` | `last - first` |
| `percent_change` | Percentage change from first to last (`null` when first is 0) |

## Report Schema

The report endpoint returns a higher-level analytics report:

```json
{
  "generated_at": "2026-04-01T12:00:00Z",
  "metric": "revenue",
  "dimension": "",
  "window_size": "5m",
  "current_windows": [ ... ],
  "previous_windows": [],
  "overall_summary": { ... },
  "trend": "up",
  "alerts": []
}
```

### Trend Values

| Value | Meaning |
|-------|---------|
| `up` | Last window average > first by >5% |
| `down` | Last window average < first by >5% |
| `flat` | Change within ±5% |
| `insufficient_data` | Fewer than 2 windows |

### Alert Types

| Type | Rule |
|------|------|
| `spike` | Window average > 2× overall average |
| `drop` | Window average < 0.5× overall average |
| `high_variance` | Window std_dev > 2× overall std_dev |

## Go HTTP API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `POST /api/summary` | POST | Upload multi-metric CSV, return filtered summary |
| `POST /api/window-summary` | POST | Upload CSV, return legacy window summaries |
| `POST /api/report` | POST | Upload CSV, return full analytics report |
| `GET /healthz` | GET | Health check |
| `GET /api/meta` | GET | API version and metadata |

### Query Parameters (form fields)

| Parameter | Description | Default |
|-----------|-------------|---------|
| `metric` | Filter by metric name | First metric in CSV |
| `dimension` | Filter by dimension | (none) |
| `window_size` | Window size (`1m`, `5m`, `15m`, `1h`) | `5m` |
| `fill_empty_windows` | Fill gaps with empty windows (`true`/`false`) | `false` |

### Example: Generate a report

```bash
curl -F "file=@fixtures/events/basic.csv" \
     -F "metric=revenue" \
     -F "window_size=5m" \
     http://localhost:8080/api/report
```

### Error Responses

All errors use a unified JSON structure:

```json
{
  "error": "bad_request",
  "message": "CSV parsing failed"
}
```

## Python CLI

The CLI supports three subcommands:

### `summary` — Compute summary statistics

```bash
# From values
python3 python/stats.py summary --values 1,2,3,4,5

# From multi-metric CSV
python3 python/stats.py summary --file fixtures/events/basic.csv --metric revenue
```

### `window-summary` — Compute window summaries

```bash
python3 python/stats.py window-summary \
  --file fixtures/events/basic.csv \
  --metric revenue \
  --window-size 5m
```

### `report` — Generate full analytics report

```bash
python3 python/stats.py report \
  --file fixtures/events/basic.csv \
  --metric revenue \
  --window-size 5m \
  --fill-empty-windows
```

### Legacy mode (backward compatible)

```bash
python3 python/stats.py --values 1,2,3,4,5
python3 python/stats.py --file events.csv
```

## TypeScript Analytics Layer

### Types (`web/src/analytics/types.ts`)

TypeScript type definitions mirroring the shared JSON contract: `Summary`, `WindowSummary`, `Report`, `Alert`, `Trend`, `QueryParams`.

### Format (`web/src/analytics/format.ts`)

Display formatting for numbers, percentages, deltas, trends, alerts, timestamps, and window sizes.

### Transform (`web/src/analytics/transform.ts`)

Convert report data into view models: table rows, chart data points, navigation items, alert groups, report comparisons.

### Query (`web/src/analytics/query.ts`)

Build API URLs, parse/serialize query parameters, manage URL hash state.

### String Helpers (`web/src/string.ts`)

| Function | Description | Example |
|----------|-------------|---------|
| `slugify(input)` | URL-safe slug | `"Hello World!"` → `"hello-world"` |
| `formatMetricLabel(key)` | Title-case label | `"total_revenue"` → `"Total Revenue"` |
| `buildSummaryAnchor(section, label)` | HTML anchor | `("revenue", "Rev")` → `<a href="#revenue">Rev</a>` |
| `truncateMiddle(input, max)` | Middle ellipsis | `("abcdefghij", 7)` → `"ab...ij"` |

## Shared Fixtures & Contract Tests

### Fixture Files

| File | Purpose |
|------|---------|
| `fixtures/events/basic.csv` | Single metric (revenue), 6 events, 2 windows |
| `fixtures/events/multi_metric.csv` | Two metrics (revenue + latency_ms), 9 events |
| `fixtures/events/invalid_rows.csv` | Mix of valid and invalid rows |
| `fixtures/events/empty.csv` | Header only, no data rows |
| `fixtures/events/single_metric_single_value.csv` | Single event |
| `fixtures/schema/report_schema.json` | JSON Schema for the report contract |
| `fixtures/expected/basic_report.json` | Expected output for basic.csv report |
| `fixtures/expected/multi_metric_revenue_5m.json` | Expected output for multi_metric.csv revenue |

### Contract Test Strategy

Go and Python both consume the same fixture CSV files and verify:
- Same event count after parsing
- Same window count and boundaries
- Same overall summary values (count, sum, min, max)
- Same trend and alert behavior

TypeScript verifies type compatibility via the shared `Summary` interface (14 fields, matching JSON keys).

## Window Design

| Aspect | Rule |
|--------|------|
| Window type | Configurable tumbling windows (default 5 minutes) |
| Sizes | `1m`, `5m`, `15m`, `1h` |
| Boundary | `[start, end)` — event at `10:05:00` belongs to `10:05–10:10` bucket |
| Out-of-order | Sorted by timestamp before grouping |
| Empty windows | Not emitted by default; `fill_empty_windows` fills gaps |
| Invalid rows | Skipped with structured warnings; valid rows still processed |

## Edge Case Behavior

| Scenario | Behavior |
|----------|----------|
| Empty input | Zero-valued summary, `percent_change: null`, trend: `insufficient_data` |
| Single value | Summary with count=1, variance=0, delta=0 |
| Zero first value | `percent_change: null` (not an error) |
| Negative values | Handled correctly; `percent_change` uses `abs(prev)` |
| Missing timezone | Row skipped with warning |
| Empty metric field | Row skipped with warning |
| Duplicate timestamps | All events included, ordered by occurrence |

## Manual Verification

```bash
# Go tests
cd go && go test ./... -v

# Python tests
cd python && python3 -m unittest discover -s . -p "test_*.py" -v

# TypeScript tests
cd web && npx jest

# CLI report
python3 python/stats.py report --file fixtures/events/basic.csv --metric revenue

# Go API report
cd go && go run ./cmd/server &
curl -F "file=@../fixtures/events/basic.csv" -F "metric=revenue" http://localhost:8080/api/report
```
