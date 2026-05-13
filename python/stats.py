"""Cross-language analytics summary toolkit — Python module.

Provides reusable statistics functions and a CLI entry point.

Usage as module:
    from stats import average, median, percent_change, build_summary

Usage as CLI:
    python3 python/stats.py --values 1,2,3,4,5
    python3 python/stats.py --file events.csv
"""

from __future__ import annotations

import argparse
import csv
import json
import math
import sys
from datetime import datetime, timedelta, timezone
from typing import Iterable

WINDOW_SIZE = timedelta(seconds=300)  # 5 minutes


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


def percent_change(prev: float, current: float) -> float:
    """Return the percentage change from prev to current.

    Raises ValueError if prev is zero.
    """
    if prev == 0:
        raise ValueError("previous value cannot be zero")
    return ((current - prev) / abs(prev)) * 100


def variance(values: Iterable[float]) -> float:
    """Return the population variance. Returns 0 for empty or single-element input."""
    items = list(values)
    n = len(items)
    if n < 2:
        return 0
    mean = average(items)
    return sum((v - mean) ** 2 for v in items) / n


def std_dev(values: Iterable[float]) -> float:
    """Return the population standard deviation. Returns 0 for empty or single-element input."""
    return math.sqrt(variance(values))


def build_summary(values: Iterable[float]) -> dict:
    """Return a dict with count, sum, min, max, average, median, variance, and std_dev."""
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
        }
    var = variance(items)
    return {
        "count": len(items),
        "sum": sum(items),
        "min": min(items),
        "max": max(items),
        "average": average(items),
        "median": median(items),
        "variance": var,
        "std_dev": math.sqrt(var),
    }


def _window_start(ts: datetime) -> datetime:
    """Return the start of the 5-minute bucket containing ts."""
    ts_utc = ts.astimezone(timezone.utc)
    epoch = int(ts_utc.timestamp())
    bucket = epoch - (epoch % int(WINDOW_SIZE.total_seconds()))
    return datetime.fromtimestamp(bucket, tz=timezone.utc)


def parse_csv(filepath: str) -> list[tuple[datetime, float]]:
    """Parse a CSV file with timestamp,value rows.

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


def build_window_summaries(
    events: list[tuple[datetime, float]],
) -> list[dict]:
    """Group events into fixed 5-minute windows and compute summaries.

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


def main() -> None:
    parser = argparse.ArgumentParser(description="Analytics summary CLI")
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument(
        "--values",
        help="Comma-separated list of numbers, e.g. 1,2,3,4,5",
    )
    group.add_argument(
        "--file",
        help="Path to a CSV file with timestamp,value columns",
    )
    args = parser.parse_args()

    if args.values is not None:
        try:
            nums = [float(v.strip()) for v in args.values.split(",") if v.strip()]
        except ValueError as exc:
            print(f"Error parsing values: {exc}", file=sys.stderr)
            sys.exit(1)

        result = build_summary(nums)
        print(json.dumps(result, indent=2))
    else:
        events = parse_csv(args.file)
        windows = build_window_summaries(events)
        result = {"windows": windows}
        print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
