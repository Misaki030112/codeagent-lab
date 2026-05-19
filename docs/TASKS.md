# Test prompts

## Simple prompts

- Issue comment: `/dev:plan explain this repository structure`
- Issue comment: `/dev:code add zero-division handling to go/calc.go and tests`
- PR comment: `/dev:review`
- Mention: `@misaki-codeagent-dev summarize the risk in this change`

## Complex planning prompt (`/dev:plan`)

> Design a streaming pipeline that reads a CSV file of timestamped numeric
> events, computes rolling 5-minute window summaries using the existing
> Go `BuildSummary` function, and exposes the results via a minimal HTTP
> endpoint that returns JSON. The Python CLI should also gain a
> `--file` flag that reads the same CSV format and prints per-window
> summaries. Outline the new files, shared schema, error handling
> strategy, and test plan.

## Implementation prompt (`/dev:code`)

> Add a `Variance(values []float64) float64` and
> `StdDev(values []float64) float64` to `go/stats.go`, with full
> table-driven tests. Then add matching `variance()` and `std_dev()`
> to `python/stats.py` (with tests), and include both new fields in
> each language's `build_summary` output. Update the README to reflect
> the new fields.

## Review prompt (`/dev:review`)

> Review the latest PR for correctness, edge-case coverage, and
> cross-language consistency. Verify that the summary schema
> (`count`, `sum`, `min`, `max`, `average`, `median`) is identical
> across Go structs, Python dicts, and documentation. Flag any
> missing tests for empty input, negative numbers, or single-element
> edge cases.

## Manual verification steps

### Go

```bash
cd go && go test -v ./...
```

Expected: all tests pass, including `stats_test.go` cases for empty
input, negatives, duplicates, and non-mutation of input slices.

### Python — tests

```bash
python3 -m unittest discover -s python -p "test_*.py" -v
```

Expected: all tests pass, covering average, median, percent_change,
and build_summary with edge cases.

### Python — CLI

```bash
python3 python/stats.py --values 10,20,30
```

Expected JSON output:

```json
{
  "count": 3,
  "sum": 60.0,
  "min": 10.0,
  "max": 30.0,
  "average": 20.0,
  "median": 20.0
}
```

### Python — CSV rolling window

```bash
python3 python/stats.py --file events.csv
```

Where `events.csv` contains:

```csv
timestamp,value
2026-04-01T10:00:00Z,10
2026-04-01T10:02:00Z,20
2026-04-01T10:06:00Z,30
```

Expected JSON output:

```json
{
  "windows": [
    {
      "window_start": "2026-04-01T10:00:00Z",
      "window_end": "2026-04-01T10:05:00Z",
      "summary": { "count": 2, "sum": 30.0, "min": 10.0, "max": 20.0, "average": 15.0, "median": 15.0 }
    },
    {
      "window_start": "2026-04-01T10:05:00Z",
      "window_end": "2026-04-01T10:10:00Z",
      "summary": { "count": 1, "sum": 30.0, "min": 30.0, "max": 30.0, "average": 30.0, "median": 30.0 }
    }
  ]
}
```

### Go — HTTP endpoint

```bash
cd go && go run ./cmd/server
# In another terminal:
curl -F "file=@events.csv" http://localhost:8080/api/window-summary
```

Expected: JSON output matching the same window schema as the Python CLI.

### TypeScript — spot-check examples

#### `slugify`

| Expression | Expected output | What is being checked |
|---|---|---|
| `slugify("Hello World!")` | `"hello-world"` | Basic lowercase + punctuation removal |
| `slugify("Hello  World")` | `"hello-world"` | Consecutive spaces collapsed to one dash |
| `slugify("hello__world")` | `"hello-world"` | Consecutive underscores → single dash |
| `slugify("Hello__World!! ")` | `"hello-world"` | Combined: double underscore + punctuation + trailing space |
| `slugify("hello...world")` | `"helloworld"` | Punctuation removed without inserting a separator |
| `slugify("  --leading-- ")` | `"leading"` | Leading/trailing dashes and spaces stripped |
| `slugify("_hello_")` | `"hello"` | Leading/trailing underscores (→ dashes) stripped |

#### `formatMetricLabel`

| Expression | Expected output | What is being checked |
|---|---|---|
| `formatMetricLabel("total_revenue")` | `"Total Revenue"` | `snake_case` split on underscore |
| `formatMetricLabel("avg_session_time")` | `"Avg Session Time"` | Multiple underscores |
| `formatMetricLabel("avgSessionTime")` | `"Avg Session Time"` | `camelCase` split before each uppercase letter |
| `formatMetricLabel("totalRevenue")` | `"Total Revenue"` | Single camelCase boundary |
| `formatMetricLabel("avg_sessionTime")` | `"Avg Session Time"` | Mixed `snake_case` + `camelCase` |

#### `truncateMiddle`

| Expression | Expected output | What is being checked |
|---|---|---|
| `truncateMiddle("abcdefghij", 7)` | `"ab...ij"` | Normal truncation: 2 front + `...` + 2 back |
| `truncateMiddle("short", 10)` | `"short"` | `input.length <= max` → returned unchanged |
| `truncateMiddle("abcd", 4)` | `"abcd"` | Exact length match → returned unchanged |
| `truncateMiddle("abcde", 3)` | `"a..."` | `max < 4` clamped to `4`; `back = 0` so no trailing chars |
| `truncateMiddle("abcde", 4)` | `"a..."` | Minimum valid `max`; single front char + `...` |
