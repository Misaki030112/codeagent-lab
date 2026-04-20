"""Cross-language analytics summary toolkit — Python module.

Provides reusable statistics functions and a multi-command CLI entry point.

Usage as module:
    from stats import average, median, percent_change, build_summary
    from stats import parse_multi_csv, build_report

Usage as CLI:
    python3 python/stats.py summary --values 1,2,3,4,5
    python3 python/stats.py window-summary --file events.csv --metric revenue --window-size 5m
    python3 python/stats.py report --file events.csv --metric revenue --window-size 5m
"""

from __future__ import annotations

import argparse
import csv
import json
import math
import sys
from datetime import datetime, timedelta, timezone
from typing import Iterable


DEFAULT_WINDOW_SIZE = timedelta(seconds=300)  # 5 minutes
WINDOW_SIZE = DEFAULT_WINDOW_SIZE  # backward compat

# Thresholds for trend detection and anomaly alerting.
TREND_THRESHOLD_PCT = 5.0
SPIKE_MULTIPLIER = 2.0
DROP_MULTIPLIER = 0.5
VARIANCE_MULTIPLIER = 2.0


# ---------------------------------------------------------------------------
# Core statistics functions
# ---------------------------------------------------------------------------

def average(values: Iterable[float]) -> float:
    """Return the arithmetic mean. Returns 0 for empty input."""
    items = list(values)
    if not items:
        return 0
    return sum(items) / len(items)


def median(values: Iterable[float]) -> float:
    """Return the median. Returns 0 for empty input."""
    items = sorted(values)
    n = len(items)
    if n == 0:
        return 0
    if n % 2 == 1:
        return items[n // 2]
    return (items[n // 2 - 1] + items[n // 2]) / 2


def variance(values: Iterable[float]) -> float:
    """Return the population variance. Returns 0 for empty input."""
    items = list(values)
    n = len(items)
    if n == 0:
        return 0
    avg = average(items)
    return sum((x - avg) ** 2 for x in items) / n


def std_dev(values: Iterable[float]) -> float:
    """Return the population standard deviation. Returns 0 for empty input."""
    return math.sqrt(variance(values))


def percentile(values: Iterable[float], p: float) -> float:
    """Return the p-th percentile (0-100) using linear interpolation.
    Returns 0 for empty input.
    """
    items = sorted(values)
    n = len(items)
    if n == 0:
        return 0
    if n == 1:
        return items[0]
    rank = (p / 100.0) * (n - 1)
    lower = int(math.floor(rank))
    upper = int(math.ceil(rank))
    if lower == upper:
        return items[lower]
    frac = rank - lower
    return items[lower] * (1 - frac) + items[upper] * frac


def percent_change(prev: float, current: float) -> float:
    """Return the percentage change from prev to current.

    Raises ValueError if prev is zero.
    """
    if prev == 0:
        raise ValueError("previous value cannot be zero")
    return ((current - prev) / abs(prev)) * 100


def build_summary(values: Iterable[float]) -> dict:
    """Return a dict with all summary statistics (shared contract).

    For ordered fields (first, last, delta, percent_change), the input
    order is used.
    """
    return build_summary_ordered(values, None)


def build_summary_ordered(values: Iterable[float], ordered_values: list[float] | None) -> dict:
    """Return a dict with all summary statistics.

    ordered_values provides time-ordered sequence for first/last/delta/percent_change.
    If None, values order is used.
    """
    items = list(values)
    if not items:
        return {
            "count": 0,
            "sum": 0,
            "min": 0,
            "max": 0,
            "average": 0,
            "median": 0,
            "variance": 0,
            "std_dev": 0,
            "p90": 0,
            "p95": 0,
            "first": 0,
            "last": 0,
            "delta": 0,
            "percent_change": None,
        }

    if ordered_values is None:
        ordered_values = items

    first_val = ordered_values[0]
    last_val = ordered_values[-1]
    delta = last_val - first_val
    try:
        pct = percent_change(first_val, last_val)
    except ValueError:
        pct = None

    return {
        "count": len(items),
        "sum": sum(items),
        "min": min(items),
        "max": max(items),
        "average": average(items),
        "median": median(items),
        "variance": variance(items),
        "std_dev": std_dev(items),
        "p90": percentile(items, 90),
        "p95": percentile(items, 95),
        "first": first_val,
        "last": last_val,
        "delta": delta,
        "percent_change": pct,
    }


# ---------------------------------------------------------------------------
# CSV parsing
# ---------------------------------------------------------------------------

def _window_start_for(ts: datetime, ws: timedelta) -> datetime:
    """Return the start of the bucket containing ts for the given window size."""
    ts_utc = ts.astimezone(timezone.utc)
    epoch = int(ts_utc.timestamp())
    ws_sec = int(ws.total_seconds())
    bucket = epoch - (epoch % ws_sec)
    return datetime.fromtimestamp(bucket, tz=timezone.utc)


def _window_start(ts: datetime) -> datetime:
    """Return the start of the 5-minute bucket containing ts (legacy)."""
    return _window_start_for(ts, WINDOW_SIZE)


def parse_csv(filepath: str) -> list[tuple[datetime, float]]:
    """Parse a CSV file with timestamp,value rows (legacy format).

    Returns a list of (datetime, float) tuples.
    Prints warnings to stderr for invalid rows but continues parsing.
    Raises SystemExit if the file cannot be opened or has an invalid header.
    """
    try:
        f = open(filepath, newline="")
    except OSError as exc:
        print(f"Error opening file: {exc}", file=sys.stderr)
        sys.exit(1)

    with f:
        reader = csv.reader(f)
        try:
            header = next(reader)
        except StopIteration:
            print("Error: CSV file is empty", file=sys.stderr)
            sys.exit(1)

        if len(header) < 2 or header[0] != "timestamp" or header[1] != "value":
            print(
                f"Error: invalid CSV header: expected [timestamp, value], got {header}",
                file=sys.stderr,
            )
            sys.exit(1)

        events: list[tuple[datetime, float]] = []
        for line_num, row in enumerate(reader, start=2):
            if len(row) < 2:
                print(f"Warning: line {line_num}: expected 2 columns, got {len(row)}", file=sys.stderr)
                continue
            try:
                ts = datetime.fromisoformat(row[0])
            except ValueError:
                print(f"Warning: line {line_num}: invalid timestamp {row[0]!r}", file=sys.stderr)
                continue
            if ts.tzinfo is None:
                print(f"Warning: line {line_num}: timestamp {row[0]!r} has no timezone offset", file=sys.stderr)
                continue
            try:
                val = float(row[1])
            except ValueError:
                print(f"Warning: line {line_num}: invalid value {row[1]!r}", file=sys.stderr)
                continue
            events.append((ts, val))

    return events


def parse_multi_csv(filepath: str) -> tuple[list[dict], list[dict]]:
    """Parse a multi-metric CSV file.

    Expected header: timestamp,metric,value,dimension,source
    metric is required; dimension and source may be empty.

    Returns (events, warnings) where:
    - events: list of dicts with keys: timestamp, metric, value, dimension, source
    - warnings: list of dicts with keys: row, message
    """
    try:
        f = open(filepath, newline="")
    except OSError as exc:
        return [], [{"row": 0, "message": f"Error opening file: {exc}"}]

    with f:
        reader = csv.reader(f)
        try:
            header = next(reader)
        except StopIteration:
            return [], [{"row": 1, "message": "CSV file is empty"}]

        if len(header) < 3 or header[0] != "timestamp" or header[1] != "metric" or header[2] != "value":
            return [], [{"row": 1, "message": f"invalid CSV header: expected [timestamp,metric,value,...], got {header}"}]

        events: list[dict] = []
        warnings: list[dict] = []

        for line_num, row in enumerate(reader, start=2):
            if len(row) < 3:
                warnings.append({"row": line_num, "message": f"expected at least 3 columns, got {len(row)}"})
                continue

            try:
                ts = datetime.fromisoformat(row[0])
            except ValueError:
                warnings.append({"row": line_num, "message": f"invalid timestamp {row[0]!r}"})
                continue

            if ts.tzinfo is None:
                warnings.append({"row": line_num, "message": f"timestamp {row[0]!r} has no timezone offset"})
                continue

            metric = row[1].strip()
            if not metric:
                warnings.append({"row": line_num, "message": "empty metric field"})
                continue

            try:
                val = float(row[2])
            except ValueError:
                warnings.append({"row": line_num, "message": f"invalid value {row[2]!r}"})
                continue

            evt = {
                "timestamp": ts,
                "metric": metric,
                "value": val,
                "dimension": row[3].strip() if len(row) > 3 else "",
                "source": row[4].strip() if len(row) > 4 else "",
            }
            events.append(evt)

    return events, warnings


def filter_events(events: list[dict], metric: str, dimension: str = "") -> list[dict]:
    """Filter multi-events by metric and optional dimension."""
    result = []
    for e in events:
        if e["metric"] != metric:
            continue
        if dimension and e.get("dimension", "") != dimension:
            continue
        result.append(e)
    return result


# ---------------------------------------------------------------------------
# Window summaries
# ---------------------------------------------------------------------------

def build_window_summaries(
    events: list[tuple[datetime, float]],
) -> list[dict]:
    """Group events into fixed 5-minute windows and compute summaries (legacy).

    Uses [start, end) semantics. Out-of-order input is handled gracefully.
    Empty windows are not emitted.
    """
    if not events:
        return []

    buckets: dict[datetime, list[float]] = {}
    for ts, val in events:
        ws = _window_start(ts)
        buckets.setdefault(ws, []).append(val)

    window_starts = sorted(buckets.keys())
    result = []
    for ws in window_starts:
        we = ws + WINDOW_SIZE
        result.append(
            {
                "window_start": ws.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "window_end": we.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "summary": build_summary(buckets[ws]),
            }
        )
    return result


def build_multi_window_summaries(
    events: list[dict],
    ws: timedelta = DEFAULT_WINDOW_SIZE,
    fill_empty: bool = False,
) -> list[dict]:
    """Group multi-events into configurable windows.

    events: list of dicts with timestamp and value keys.
    ws: window size duration.
    fill_empty: if True, fills empty windows between min and max timestamps.
    """
    if not events:
        return []

    # Sort by timestamp
    sorted_events = sorted(events, key=lambda e: e["timestamp"])

    buckets: dict[datetime, list[float]] = {}
    for e in sorted_events:
        start = _window_start_for(e["timestamp"], ws)
        buckets.setdefault(start, []).append(e["value"])

    starts = sorted(buckets.keys())

    if fill_empty and len(starts) > 1:
        min_start = starts[0]
        max_start = starts[-1]
        all_starts = []
        t = min_start
        while t <= max_start:
            all_starts.append(t)
            if t not in buckets:
                buckets[t] = []
            t += ws
        starts = all_starts

    result = []
    for s in starts:
        vals = buckets[s]
        result.append({
            "window_start": s.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "window_end": (s + ws).strftime("%Y-%m-%dT%H:%M:%SZ"),
            "summary": build_summary_ordered(vals, vals),
        })
    return result


# ---------------------------------------------------------------------------
# Report generation
# ---------------------------------------------------------------------------

def compute_trend(windows: list[dict]) -> str:
    """Determine overall trend from window summaries."""
    if len(windows) < 2:
        return "insufficient_data"

    first_avg = windows[0]["summary"]["average"]
    last_avg = windows[-1]["summary"]["average"]

    if first_avg == 0 and last_avg == 0:
        return "flat"

    if first_avg != 0:
        pct = ((last_avg - first_avg) / first_avg) * 100
    else:
        return "up" if last_avg > 0 else "down"

    if pct > TREND_THRESHOLD_PCT:
        return "up"
    if pct < -TREND_THRESHOLD_PCT:
        return "down"
    return "flat"


def detect_alerts(windows: list[dict]) -> list[dict]:
    """Detect anomalies in windows using heuristic rules.

    Spike: window average > 2x overall average
    Drop: window average < 0.5x overall average
    High variance: window std_dev > 2x overall std_dev
    """
    if len(windows) < 2:
        return []

    avgs = [w["summary"]["average"] for w in windows if w["summary"]["count"] > 0]
    if not avgs:
        return []

    overall_avg = average(avgs)
    overall_std = std_dev(avgs)

    alerts = []
    for w in windows:
        if w["summary"]["count"] == 0:
            continue

        w_avg = w["summary"]["average"]
        w_std = w["summary"]["std_dev"]

        spike_threshold = overall_avg * SPIKE_MULTIPLIER
        drop_threshold = overall_avg * DROP_MULTIPLIER
        var_threshold = overall_std * VARIANCE_MULTIPLIER

        if overall_avg > 0 and w_avg > spike_threshold:
            alerts.append({
                "type": "spike",
                "window_start": w["window_start"],
                "message": "window average is more than 2x the overall average",
                "value": w_avg,
                "threshold": spike_threshold,
            })

        if overall_avg > 0 and w_avg < drop_threshold:
            alerts.append({
                "type": "drop",
                "window_start": w["window_start"],
                "message": "window average is less than 0.5x the overall average",
                "value": w_avg,
                "threshold": drop_threshold,
            })

        if overall_std > 0 and w_std > var_threshold:
            alerts.append({
                "type": "high_variance",
                "window_start": w["window_start"],
                "message": "window standard deviation is more than 2x the overall",
                "value": w_std,
                "threshold": var_threshold,
            })

    return alerts


def build_report(
    events: list[dict],
    metric: str,
    dimension: str = "",
    ws: timedelta = DEFAULT_WINDOW_SIZE,
    fill_empty: bool = False,
) -> dict:
    """Generate a full analytics report for filtered events."""
    filtered = filter_events(events, metric, dimension)
    filtered.sort(key=lambda e: e["timestamp"])

    windows = build_multi_window_summaries(filtered, ws, fill_empty)

    values = [e["value"] for e in filtered]
    overall = build_summary_ordered(values, values)

    trend = compute_trend(windows)
    alerts = detect_alerts(windows)

    # Format window size
    total_min = int(ws.total_seconds()) // 60
    if total_min >= 60 and total_min % 60 == 0:
        ws_str = f"{total_min // 60}h"
    else:
        ws_str = f"{total_min}m"

    return {
        "generated_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
        "metric": metric,
        "dimension": dimension,
        "window_size": ws_str,
        "current_windows": windows,
        "previous_windows": [],
        "overall_summary": overall,
        "trend": trend,
        "alerts": alerts,
    }


# ---------------------------------------------------------------------------
# Window size parsing
# ---------------------------------------------------------------------------

def parse_window_size(s: str) -> timedelta:
    """Parse a window size string like '1m', '5m', '15m', '1h'.

    Returns DEFAULT_WINDOW_SIZE for empty string.
    """
    if not s:
        return DEFAULT_WINDOW_SIZE
    s = s.strip()
    if len(s) < 2:
        raise ValueError(f"invalid window size: {s!r}")
    unit = s[-1]
    try:
        num = int(s[:-1])
    except ValueError:
        raise ValueError(f"invalid window size: {s!r}")
    if num <= 0:
        raise ValueError(f"invalid window size: {s!r}")
    if unit == "m":
        return timedelta(minutes=num)
    if unit == "h":
        return timedelta(hours=num)
    raise ValueError(f"invalid window size unit {unit!r}: expected 'm' or 'h'")


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def _parse_values_arg(raw: str) -> list[float]:
    """Parse a comma-separated string of numbers into a list of floats."""
    try:
        return [float(v.strip()) for v in raw.split(",") if v.strip()]
    except ValueError as exc:
        print(f"Error parsing values: {exc}", file=sys.stderr)
        sys.exit(1)


def main() -> None:
    parser = argparse.ArgumentParser(description="Analytics summary CLI")
    subparsers = parser.add_subparsers(dest="command")

    # summary subcommand
    sum_parser = subparsers.add_parser("summary", help="Compute summary statistics")
    sum_group = sum_parser.add_mutually_exclusive_group(required=True)
    sum_group.add_argument("--values", help="Comma-separated list of numbers")
    sum_group.add_argument("--file", help="Path to a multi-metric CSV file")
    sum_parser.add_argument("--metric", help="Metric to filter by")
    sum_parser.add_argument("--dimension", default="", help="Dimension to filter by")

    # window-summary subcommand
    ws_parser = subparsers.add_parser("window-summary", help="Compute window summaries")
    ws_parser.add_argument("--file", required=True, help="Path to a multi-metric CSV file")
    ws_parser.add_argument("--metric", help="Metric to filter by")
    ws_parser.add_argument("--dimension", default="", help="Dimension to filter by")
    ws_parser.add_argument("--window-size", default="5m", help="Window size (e.g. 1m, 5m, 15m, 1h)")
    ws_parser.add_argument("--fill-empty-windows", action="store_true", help="Fill empty windows")

    # report subcommand
    rpt_parser = subparsers.add_parser("report", help="Generate analytics report")
    rpt_parser.add_argument("--file", required=True, help="Path to a multi-metric CSV file")
    rpt_parser.add_argument("--metric", help="Metric to filter by")
    rpt_parser.add_argument("--dimension", default="", help="Dimension to filter by")
    rpt_parser.add_argument("--window-size", default="5m", help="Window size (e.g. 1m, 5m, 15m, 1h)")
    rpt_parser.add_argument("--fill-empty-windows", action="store_true", help="Fill empty windows")

    # Legacy: support old --values/--file style without subcommand
    parser.add_argument("--values", help=argparse.SUPPRESS)
    parser.add_argument("--file", help=argparse.SUPPRESS)

    args = parser.parse_args()

    # Handle legacy mode (no subcommand)
    if args.command is None:
        if getattr(args, "values", None) is not None:
            result = build_summary(_parse_values_arg(args.values))
            print(json.dumps(result, indent=2))
            return
        elif getattr(args, "file", None) is not None:
            events = parse_csv(args.file)
            windows = build_window_summaries(events)
            result = {"windows": windows}
            print(json.dumps(result, indent=2))
            return
        else:
            parser.print_help()
            sys.exit(1)

    if args.command == "summary":
        if args.values is not None:
            result = build_summary(_parse_values_arg(args.values))
            print(json.dumps(result, indent=2))
        else:
            events, warnings = parse_multi_csv(args.file)
            metric = args.metric
            if not metric and events:
                metric = events[0]["metric"]
            filtered = filter_events(events, metric, args.dimension)
            values = [e["value"] for e in filtered]
            result = build_summary(values)
            output = {"metric": metric, "dimension": args.dimension, "summary": result}
            if warnings:
                output["warnings"] = warnings
            print(json.dumps(output, indent=2))

    elif args.command == "window-summary":
        events, warnings = parse_multi_csv(args.file)
        ws = parse_window_size(args.window_size)
        metric = args.metric
        if not metric and events:
            metric = events[0]["metric"]
        filtered = filter_events(events, metric, args.dimension)
        windows = build_multi_window_summaries(filtered, ws, args.fill_empty_windows)
        output = {"windows": windows}
        if warnings:
            output["warnings"] = warnings
        print(json.dumps(output, indent=2))

    elif args.command == "report":
        events, warnings = parse_multi_csv(args.file)
        ws = parse_window_size(args.window_size)
        metric = args.metric
        if not metric and events:
            metric = events[0]["metric"]
        report = build_report(events, metric, args.dimension, ws, args.fill_empty_windows)
        output = {"report": report}
        if warnings:
            output["warnings"] = warnings
        print(json.dumps(output, indent=2))


if __name__ == "__main__":
    main()
