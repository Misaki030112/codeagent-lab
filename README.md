# codeagent-lab

Cross-language analytics summary toolkit — a lab repository for CodeAgent integration testing.

## Overview

This repository implements a small but realistic analytics toolkit across three languages, sharing a unified **summary** schema (`count`, `sum`, `min`, `max`, `average`, `median`). It is designed to exercise multi-file, cross-language planning, code generation, and review workflows.

## Layout

| Directory | Language | Purpose |
|-----------|----------|---------|
| `go/` | Go | Arithmetic helpers (`calc.go`) and statistical summary functions (`stats.go`) |
| `python/` | Python | Reusable stats module with CLI entry point (`stats.py`) |
| `web/` | TypeScript | Display/string helpers for analytics dashboards (`string.ts`) |
| `docs/` | Markdown | Test prompts and manual verification steps |

## Quick start

### Go — run tests

```bash
cd go && go test ./...
```

### Go — run HTTP server

```bash
cd go && go run ./cmd/server
```

Upload a CSV file to get rolling 5-minute window summaries:

```bash
curl -F "file=@events.csv" http://localhost:8080/api/window-summary
```

### Python — run tests

```bash
cd python && python3 -m unittest discover -s . -p "test_*.py"
```

### Python — CLI demo

```bash
python3 python/stats.py --values 1,2,3,4,5
```

Output (JSON):

```json
{
  "count": 5,
  "sum": 15,
  "min": 1,
  "max": 5,
  "average": 3.0,
  "median": 3
}
```

### Python — CSV rolling window

```bash
python3 python/stats.py --file events.csv
```

Output (JSON):

```json
{
  "windows": [
    {
      "window_start": "2026-04-01T10:00:00Z",
      "window_end": "2026-04-01T10:05:00Z",
      "summary": {
        "count": 2,
        "sum": 30.0,
        "min": 10.0,
        "max": 20.0,
        "average": 15.0,
        "median": 15.0
      }
    }
  ]
}
```

> **Note:** `--values` and `--file` are mutually exclusive.

### TypeScript — function reference

| Function | Description | Example |
|----------|-------------|---------|
| `slugify(input)` | URL-safe slug | `"Hello World!"` → `"hello-world"` |
| `formatMetricLabel(key)` | Title-case label from snake/camel key | `"total_revenue"` → `"Total Revenue"` |
| `buildSummaryAnchor(section, label)` | HTML anchor for summary navigation | `("revenue", "Rev")` → `<a href="#revenue">Rev</a>` |
| `truncateMiddle(input, max)` | Ellipsis in the middle | `("abcdefghij", 7)` → `"ab...ij"` |

#### `slugify` — boundary behaviour

| Input | Output | Notes |
|-------|--------|-------|
| `"Hello  World"` | `"hello-world"` | Consecutive spaces collapsed to one dash |
| `"hello__world"` | `"hello-world"` | Consecutive underscores → single dash |
| `"hello...world"` | `"helloworld"` | Punctuation removed (no separator inserted) |
| `"  --leading-- "` | `"leading"` | Leading/trailing dashes and spaces stripped |
| `"Hello_World!"` | `"hello-world"` | Underscore → dash, punctuation removed |

#### `formatMetricLabel` — boundary behaviour

| Input | Output | Notes |
|-------|--------|-------|
| `"total_revenue"` | `"Total Revenue"` | `snake_case` split on underscore |
| `"avgSessionTime"` | `"Avg Session Time"` | `camelCase` split before each uppercase letter |
| `"totalRevenue"` | `"Total Revenue"` | Single camelCase boundary |
| `"avg_sessionTime"` | `"Avg Session Time"` | Mixed snake + camel both handled |

#### `truncateMiddle` — boundary behaviour

| Call | Output | Notes |
|------|--------|-------|
| `truncateMiddle("short", 10)` | `"short"` | `input.length <= max` → returned unchanged |
| `truncateMiddle("abcd", 4)` | `"abcd"` | Exact length match → returned unchanged |
| `truncateMiddle("abcde", 3)` | `"a..."` | `max < 4` clamped to `4`; `back = 0` so no trailing chars |
| `truncateMiddle("abcdefghij", 7)` | `"ab...ij"` | Normal case: 2 front + `...` + 2 back |

## API summary

### Go (`package calc`)

- `Add(a, b int) int`
- `Divide(a, b int) (int, error)`
- `Sum(values []float64) float64`
- `Average(values []float64) float64`
- `Median(values []float64) float64`
- `PercentChange(prev, current float64) (float64, error)`
- `BuildSummary(values []float64) Summary`
- `ParseCSV(r io.Reader) ([]Event, []error)` — parse `timestamp,value` CSV
- `BuildWindowSummaries(events []Event) []WindowSummary` — fixed 5-min window aggregation
- `HandleWindowSummary(w, r)` — HTTP handler (POST multipart form with `file` field)

### Python (`stats`)

- `average(values)` → `float`
- `median(values)` → `float`
- `percent_change(prev, current)` → `float`
- `build_summary(values)` → `dict`
- `parse_csv(filepath)` → `list[tuple[datetime, float]]`
- `build_window_summaries(events)` → `list[dict]`

## CSV input format

```csv
timestamp,value
2026-04-01T10:00:00Z,10
2026-04-01T10:02:00Z,20
2026-04-01T10:06:00Z,30
```

- `timestamp`: RFC 3339 format (UTC recommended)
- `value`: floating-point number

## Rolling window design

| Aspect | Rule |
|--------|------|
| Window type | Fixed 5-minute time buckets (not sliding) |
| Boundary | `[start, end)` — an event at exactly `10:05:00` belongs to the `10:05–10:10` bucket |
| Out-of-order input | Sorted by timestamp before grouping |
| Empty windows | Not emitted — only windows containing events appear in output |
| Invalid rows | Skipped with warnings; valid rows are still processed |
| Shared schema | `window_start`, `window_end`, `summary` with `count`/`sum`/`min`/`max`/`average`/`median` |
