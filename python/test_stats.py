"""Tests for the stats module."""

import json
import math
import os
import subprocess
import tempfile
import unittest
from datetime import datetime, timezone

from stats import (
    average,
    build_summary,
    build_window_summaries,
    median,
    parse_csv,
    percent_change,
    std_dev,
    variance,
)


def _write_csv(test_case: unittest.TestCase, content: str) -> str:
    """Write content to a temporary CSV file, register cleanup, and return its path."""
    f = tempfile.NamedTemporaryFile(
        mode="w", suffix=".csv", delete=False, newline=""
    )
    f.write(content)
    f.close()
    test_case.addCleanup(os.unlink, f.name)
    return f.name


class TestAverage(unittest.TestCase):
    def test_empty(self):
        self.assertEqual(average([]), 0)

    def test_single(self):
        self.assertEqual(average([5]), 5)

    def test_multiple(self):
        self.assertAlmostEqual(average([1, 2, 3, 4, 5]), 3.0)

    def test_negatives(self):
        self.assertAlmostEqual(average([-10, 10]), 0.0)

    def test_generator(self):
        self.assertAlmostEqual(average(x for x in [2, 4, 6]), 4.0)


class TestMedian(unittest.TestCase):
    def test_empty(self):
        self.assertEqual(median([]), 0)

    def test_single(self):
        self.assertEqual(median([42]), 42)

    def test_odd_count(self):
        self.assertEqual(median([3, 1, 2]), 2)

    def test_even_count(self):
        self.assertEqual(median([4, 1, 3, 2]), 2.5)

    def test_negatives(self):
        self.assertEqual(median([-5, -1, -3]), -3)

    def test_duplicates(self):
        self.assertEqual(median([7, 7, 7, 7]), 7)

    def test_generator(self):
        self.assertEqual(median(x for x in [5, 3, 1]), 3)


class TestPercentChange(unittest.TestCase):
    def test_increase(self):
        self.assertAlmostEqual(percent_change(100, 150), 50.0)

    def test_decrease(self):
        self.assertAlmostEqual(percent_change(200, 100), -50.0)

    def test_no_change(self):
        self.assertAlmostEqual(percent_change(42, 42), 0.0)

    def test_zero_prev_raises(self):
        with self.assertRaises(ValueError):
            percent_change(0, 50)

    def test_negative_prev(self):
        self.assertAlmostEqual(percent_change(-100, -50), 50.0)


class TestVariance(unittest.TestCase):
    def test_empty(self):
        self.assertEqual(variance([]), 0)

    def test_single(self):
        self.assertEqual(variance([42]), 0)

    def test_two_equal(self):
        self.assertEqual(variance([3, 3]), 0)

    def test_two_different(self):
        self.assertAlmostEqual(variance([2, 4]), 1.0)

    def test_known_values(self):
        self.assertAlmostEqual(variance([2, 4, 4, 4, 5, 5, 7, 9]), 4.0)

    def test_negatives(self):
        self.assertAlmostEqual(variance([-3, -1, 1, 3]), 5.0)

    def test_generator(self):
        self.assertAlmostEqual(variance(x for x in [2, 4, 4, 4, 5, 5, 7, 9]), 4.0)


class TestStdDev(unittest.TestCase):
    def test_empty(self):
        self.assertEqual(std_dev([]), 0)

    def test_single(self):
        self.assertEqual(std_dev([5]), 0)

    def test_two_equal(self):
        self.assertEqual(std_dev([7, 7]), 0)

    def test_known_values(self):
        self.assertAlmostEqual(std_dev([2, 4, 4, 4, 5, 5, 7, 9]), 2.0)

    def test_negatives(self):
        self.assertAlmostEqual(std_dev([-3, -1, 1, 3]), math.sqrt(5))

    def test_generator(self):
        self.assertAlmostEqual(std_dev(x for x in [2, 4, 4, 4, 5, 5, 7, 9]), 2.0)


class TestBuildSummary(unittest.TestCase):
    def test_empty(self):
        s = build_summary([])
        self.assertEqual(s["count"], 0)
        self.assertEqual(s["sum"], 0)

    def test_single(self):
        s = build_summary([10])
        self.assertEqual(s["count"], 1)
        self.assertEqual(s["sum"], 10)
        self.assertEqual(s["min"], 10)
        self.assertEqual(s["max"], 10)
        self.assertEqual(s["average"], 10)
        self.assertEqual(s["median"], 10)

    def test_multiple(self):
        s = build_summary([1, 2, 3, 4, 5])
        self.assertEqual(s["count"], 5)
        self.assertEqual(s["sum"], 15)
        self.assertEqual(s["min"], 1)
        self.assertEqual(s["max"], 5)
        self.assertAlmostEqual(s["average"], 3.0)
        self.assertEqual(s["median"], 3)

    def test_negatives(self):
        s = build_summary([-10, -5, 0, 5, 10])
        self.assertEqual(s["min"], -10)
        self.assertEqual(s["max"], 10)
        self.assertAlmostEqual(s["average"], 0.0)

    def test_keys_present(self):
        s = build_summary([1, 2, 3])
        for key in ("count", "sum", "min", "max", "average", "median", "variance", "std_dev"):
            self.assertIn(key, s)

    def test_variance_and_stddev(self):
        s = build_summary([2, 4, 4, 4, 5, 5, 7, 9])
        self.assertAlmostEqual(s["variance"], 4.0)
        self.assertAlmostEqual(s["std_dev"], 2.0)

    def test_empty_variance_and_stddev(self):
        s = build_summary([])
        self.assertEqual(s["variance"], 0)
        self.assertEqual(s["std_dev"], 0)

    def test_generator_input(self):
        s = build_summary(x for x in [2, 4, 6])
        self.assertEqual(s["count"], 3)
        self.assertAlmostEqual(s["average"], 4.0)


class TestParseCSV(unittest.TestCase):
    def test_normal(self):
        path = _write_csv(
            self,
            "timestamp,value\n2026-04-01T10:00:00+00:00,10\n2026-04-01T10:02:00+00:00,20\n"
        )
        events = parse_csv(path)
        self.assertEqual(len(events), 2)
        self.assertEqual(events[0][1], 10.0)
        self.assertEqual(events[1][1], 20.0)

    def test_empty_file(self):
        path = _write_csv(self, "")
        with self.assertRaises(SystemExit):
            parse_csv(path)

    def test_invalid_header(self):
        path = _write_csv(self, "time,val\n1,2\n")
        with self.assertRaises(SystemExit):
            parse_csv(path)

    def test_header_only(self):
        path = _write_csv(self, "timestamp,value\n")
        events = parse_csv(path)
        self.assertEqual(len(events), 0)

    def test_invalid_timestamp_skipped(self):
        path = _write_csv(
            self,
            "timestamp,value\nnot-a-time,10\n2026-04-01T10:00:00+00:00,20\n"
        )
        events = parse_csv(path)
        self.assertEqual(len(events), 1)
        self.assertEqual(events[0][1], 20.0)

    def test_invalid_value_skipped(self):
        path = _write_csv(
            self,
            "timestamp,value\n2026-04-01T10:00:00+00:00,abc\n2026-04-01T10:01:00+00:00,5\n"
        )
        events = parse_csv(path)
        self.assertEqual(len(events), 1)
        self.assertEqual(events[0][1], 5.0)

    def test_file_not_found(self):
        with self.assertRaises(SystemExit):
            parse_csv("/nonexistent/path.csv")

    def test_naive_timestamp_skipped(self):
        path = _write_csv(
            self,
            "timestamp,value\n2026-04-01T10:00:00,10\n2026-04-01T10:02:00+00:00,20\n"
        )
        events = parse_csv(path)
        self.assertEqual(len(events), 1)
        self.assertEqual(events[0][1], 20.0)


class TestBuildWindowSummaries(unittest.TestCase):
    def _ts(self, s: str) -> datetime:
        return datetime.fromisoformat(s)

    def test_empty(self):
        self.assertEqual(build_window_summaries([]), [])

    def test_single_window(self):
        events = [
            (self._ts("2026-04-01T10:00:00+00:00"), 10.0),
            (self._ts("2026-04-01T10:02:00+00:00"), 20.0),
            (self._ts("2026-04-01T10:04:59+00:00"), 30.0),
        ]
        windows = build_window_summaries(events)
        self.assertEqual(len(windows), 1)
        self.assertEqual(windows[0]["window_start"], "2026-04-01T10:00:00Z")
        self.assertEqual(windows[0]["window_end"], "2026-04-01T10:05:00Z")
        self.assertEqual(windows[0]["summary"]["count"], 3)
        self.assertEqual(windows[0]["summary"]["sum"], 60.0)

    def test_multiple_windows(self):
        events = [
            (self._ts("2026-04-01T10:00:00+00:00"), 10.0),
            (self._ts("2026-04-01T10:02:00+00:00"), 20.0),
            (self._ts("2026-04-01T10:06:00+00:00"), 30.0),
            (self._ts("2026-04-01T10:11:00+00:00"), 40.0),
        ]
        windows = build_window_summaries(events)
        self.assertEqual(len(windows), 3)
        self.assertEqual(windows[0]["summary"]["count"], 2)
        self.assertEqual(windows[1]["summary"]["count"], 1)
        self.assertEqual(windows[2]["summary"]["count"], 1)

    def test_out_of_order(self):
        events = [
            (self._ts("2026-04-01T10:06:00+00:00"), 30.0),
            (self._ts("2026-04-01T10:00:00+00:00"), 10.0),
            (self._ts("2026-04-01T10:02:00+00:00"), 20.0),
        ]
        windows = build_window_summaries(events)
        self.assertEqual(len(windows), 2)
        self.assertEqual(windows[0]["window_start"], "2026-04-01T10:00:00Z")
        self.assertEqual(windows[0]["summary"]["count"], 2)

    def test_boundary_point(self):
        events = [
            (self._ts("2026-04-01T10:04:59+00:00"), 1.0),
            (self._ts("2026-04-01T10:05:00+00:00"), 2.0),
        ]
        windows = build_window_summaries(events)
        self.assertEqual(len(windows), 2)
        self.assertEqual(windows[0]["summary"]["count"], 1)
        self.assertEqual(windows[0]["summary"]["sum"], 1.0)
        self.assertEqual(windows[1]["summary"]["count"], 1)
        self.assertEqual(windows[1]["summary"]["sum"], 2.0)


class TestCLIFileMode(unittest.TestCase):
    _stats_py_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "stats.py")

    def test_file_flag_produces_windows(self):
        path = _write_csv(
            self,
            "timestamp,value\n2026-04-01T10:00:00+00:00,10\n2026-04-01T10:02:00+00:00,20\n"
        )
        result = subprocess.run(
            ["python3", self._stats_py_path, "--file", path],
            capture_output=True,
            text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        data = json.loads(result.stdout)
        self.assertIn("windows", data)
        self.assertEqual(len(data["windows"]), 1)

    def test_values_and_file_mutually_exclusive(self):
        result = subprocess.run(
            ["python3", self._stats_py_path, "--values", "1,2", "--file", "x.csv"],
            capture_output=True,
            text=True,
        )
        self.assertNotEqual(result.returncode, 0)

    def test_values_still_works(self):
        result = subprocess.run(
            ["python3", self._stats_py_path, "--values", "1,2,3"],
            capture_output=True,
            text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        data = json.loads(result.stdout)
        self.assertEqual(data["count"], 3)


if __name__ == "__main__":
    unittest.main()
