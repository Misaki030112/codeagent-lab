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

### Python — run tests

```bash
python3 -m unittest discover -s python -p "test_*.py"
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

### TypeScript — function reference

| Function | Description | Example |
|----------|-------------|---------|
| `slugify(input)` | URL-safe slug | `"Hello World!"` → `"hello-world"` |
| `formatMetricLabel(key)` | Title-case label from snake/camel key | `"total_revenue"` → `"Total Revenue"` |
| `buildSummaryAnchor(section, label)` | HTML anchor for summary navigation | `("revenue", "Rev")` → `<a href="#revenue">Rev</a>` |
| `truncateMiddle(input, max)` | Ellipsis in the middle | `("abcdefghij", 7)` → `"ab...ij"` |

## API summary

### Go (`package calc`)

- `Add(a, b int) int`
- `Divide(a, b int) (int, error)`
- `Sum(values []float64) float64`
- `Average(values []float64) float64`
- `Median(values []float64) float64`
- `PercentChange(prev, current float64) (float64, error)`
- `BuildSummary(values []float64) Summary`

### Python (`stats`)

- `average(values)` → `float`
- `median(values)` → `float`
- `percent_change(prev, current)` → `float`
- `build_summary(values)` → `dict`
