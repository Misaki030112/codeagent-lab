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

### TypeScript — spot-check examples

| Expression | Expected output |
|---|---|
| `slugify("Hello__World!! ")` | `"hello-world"` |
| `slugify("  --leading-- ")` | `"leading"` |
| `formatMetricLabel("avg_session_time")` | `"Avg Session Time"` |
| `formatMetricLabel("avgSessionTime")` | `"Avg Session Time"` |
| `truncateMiddle("abcdefghij", 7)` | `"ab...ij"` |
| `truncateMiddle("short", 10)` | `"short"` |
