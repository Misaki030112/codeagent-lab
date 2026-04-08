"""Cross-language analytics summary toolkit — Python module.

Provides reusable statistics functions and a CLI entry point.

Usage as module:
    from stats import average, median, percent_change, build_summary

Usage as CLI:
    python3 python/stats.py --values 1,2,3,4,5
"""

from __future__ import annotations

import argparse
import json
import sys
from typing import Iterable


def average(values: Iterable[float]) -> float:
    """Return the arithmetic mean. Returns 0 for empty input."""
    items = list(values)
    if not items:
        return 0
    return sum(items) / len(items)


def median(values: Iterable[float]) -> float:
    """Return the median. Returns 0 for empty input."""
    items = sorted(list(values))
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


def build_summary(values: Iterable[float]) -> dict:
    """Return a dict with count, sum, min, max, average, and median."""
    items = list(values)
    if not items:
        return {
            "count": 0,
            "sum": 0,
            "min": 0,
            "max": 0,
            "average": 0,
            "median": 0,
        }
    return {
        "count": len(items),
        "sum": sum(items),
        "min": min(items),
        "max": max(items),
        "average": average(items),
        "median": median(items),
    }


def main() -> None:
    parser = argparse.ArgumentParser(description="Analytics summary CLI")
    parser.add_argument(
        "--values",
        required=True,
        help="Comma-separated list of numbers, e.g. 1,2,3,4,5",
    )
    args = parser.parse_args()

    try:
        nums = [float(v.strip()) for v in args.values.split(",") if v.strip()]
    except ValueError as exc:
        print(f"Error parsing values: {exc}", file=sys.stderr)
        sys.exit(1)

    result = build_summary(nums)
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
