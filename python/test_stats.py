"""Tests for the stats module."""

import json
import math
import os
import subprocess
import tempfile
import unittest
from datetime import datetime, timedelta, timezone

from stats import (
    average,
    build_multi_window_summaries,
    build_report,
    build_summary,
    build_summary_ordered,
    build_window_summaries,
    compute_trend,
    detect_alerts,
    filter_events,
    median,
    parse_csv,
    parse_multi_csv,
    parse_window_size,
    percent_change,
    percentile,
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


FIXTURES_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "fixtures")


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


class TestVariance(unittest.TestCase):
    def test_empty(self):
        self.assertEqual(variance([]), 0)

    def test_single(self):
        self.assertEqual(variance([5]), 0)

    def test_multiple(self):
        # [2, 4, 4, 4, 5, 5, 7, 9] -> mean=5, pop variance=4
        self.assertAlmostEqual(variance([2, 4, 4, 4, 5, 5, 7, 9]), 4.0)

    def test_identical(self):
        self.assertEqual(variance([3, 3, 3]), 0)


class TestStdDev(unittest.TestCase):
    def test_empty(self):
        self.assertEqual(std_dev([]), 0)

    def test_multiple(self):
        self.assertAlmostEqual(std_dev([2, 4, 4, 4, 5, 5, 7, 9]), 2.0)


class TestPercentile(unittest.TestCase):
    def test_empty(self):
        self.assertEqual(percentile([], 90), 0)

    def test_single(self):
        self.assertEqual(percentile([42], 90), 42)

    def test_p50_equals_median(self):
        values = [1, 2, 3, 4, 5]
        self.assertAlmostEqual(percentile(values, 50), median(values))

    def test_p90(self):
        values = list(range(1, 11))
        self.assertAlmostEqual(percentile(values, 90), 9.1)

    def test_p0(self):
        self.assertEqual(percentile([1, 2, 3, 4, 5], 0), 1)

    def test_p100(self):
        self.assertEqual(percentile([1, 2, 3, 4, 5], 100), 5)


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


class TestBuildSummary(unittest.TestCase):
    def test_empty(self):
        s = build_summary([])
        self.assertEqual(s["count"], 0)
        self.assertEqual(s["sum"], 0)
        self.assertEqual(s["variance"], 0)
        self.assertEqual(s["std_dev"], 0)
        self.assertIsNone(s["percent_change"])

    def test_single(self):
        s = build_summary([10])
        self.assertEqual(s["count"], 1)
        self.assertEqual(s["sum"], 10)
        self.assertEqual(s["min"], 10)
        self.assertEqual(s["max"], 10)
        self.assertEqual(s["average"], 10)
        self.assertEqual(s["median"], 10)
        self.assertEqual(s["variance"], 0)
        self.assertEqual(s["std_dev"], 0)
        self.assertEqual(s["first"], 10)
        self.assertEqual(s["last"], 10)
        self.assertEqual(s["delta"], 0)

    def test_multiple(self):
        s = build_summary([1, 2, 3, 4, 5])
        self.assertEqual(s["count"], 5)
        self.assertEqual(s["sum"], 15)
        self.assertEqual(s["min"], 1)
        self.assertEqual(s["max"], 5)
        self.assertAlmostEqual(s["average"], 3.0)
        self.assertEqual(s["median"], 3)
        self.assertAlmostEqual(s["variance"], 2.0)
        self.assertAlmostEqual(s["std_dev"], math.sqrt(2.0))
        self.assertEqual(s["first"], 1)
        self.assertEqual(s["last"], 5)
        self.assertEqual(s["delta"], 4)
        self.assertAlmostEqual(s["percent_change"], 400.0)

    def test_negatives(self):
        s = build_summary([-10, -5, 0, 5, 10])
        self.assertEqual(s["min"], -10)
        self.assertEqual(s["max"], 10)
        self.assertAlmostEqual(s["average"], 0.0)

    def test_keys_present(self):
        s = build_summary([1, 2, 3])
        for key in ("count", "sum", "min", "max", "average", "median",
                     "variance", "std_dev", "p90", "p95", "first", "last",
                     "delta", "percent_change"):
            self.assertIn(key, s)

    def test_generator_input(self):
        s = build_summary(x for x in [2, 4, 6])
        self.assertEqual(s["count"], 3)
        self.assertAlmostEqual(s["average"], 4.0)

    def test_percent_change_zero_first(self):
        s = build_summary([0, 5, 10])
        self.assertIsNone(s["percent_change"])


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


class TestParseMultiCSV(unittest.TestCase):
    def test_normal(self):
        path = _write_csv(
            self,
            "timestamp,metric,value,dimension,source\n"
            "2026-04-01T10:00:00+00:00,revenue,120.5,cn,ads\n"
            "2026-04-01T10:01:00+00:00,latency_ms,240.0,us,api\n"
        )
        events, warnings = parse_multi_csv(path)
        self.assertEqual(len(warnings), 0)
        self.assertEqual(len(events), 2)
        self.assertEqual(events[0]["metric"], "revenue")
        self.assertAlmostEqual(events[0]["value"], 120.5)

    def test_empty_metric(self):
        path = _write_csv(
            self,
            "timestamp,metric,value,dimension,source\n"
            "2026-04-01T10:00:00+00:00,,120.5,cn,ads\n"
        )
        events, warnings = parse_multi_csv(path)
        self.assertEqual(len(events), 0)
        self.assertEqual(len(warnings), 1)
        self.assertIn("empty metric", warnings[0]["message"])

    def test_optional_columns(self):
        path = _write_csv(
            self,
            "timestamp,metric,value\n"
            "2026-04-01T10:00:00+00:00,revenue,120.5\n"
        )
        events, warnings = parse_multi_csv(path)
        self.assertEqual(len(warnings), 0)
        self.assertEqual(len(events), 1)
        self.assertEqual(events[0]["dimension"], "")
        self.assertEqual(events[0]["source"], "")


class TestFilterEvents(unittest.TestCase):
    def test_by_metric(self):
        events = [
            {"metric": "revenue", "value": 100, "dimension": "cn"},
            {"metric": "latency_ms", "value": 200, "dimension": "us"},
            {"metric": "revenue", "value": 300, "dimension": "us"},
        ]
        filtered = filter_events(events, "revenue")
        self.assertEqual(len(filtered), 2)

    def test_by_metric_and_dimension(self):
        events = [
            {"metric": "revenue", "value": 100, "dimension": "cn"},
            {"metric": "revenue", "value": 200, "dimension": "us"},
            {"metric": "revenue", "value": 300, "dimension": "cn"},
        ]
        filtered = filter_events(events, "revenue", "cn")
        self.assertEqual(len(filtered), 2)


class TestParseWindowSize(unittest.TestCase):
    def test_default(self):
        self.assertEqual(parse_window_size(""), timedelta(minutes=5))

    def test_minutes(self):
        self.assertEqual(parse_window_size("15m"), timedelta(minutes=15))

    def test_hours(self):
        self.assertEqual(parse_window_size("1h"), timedelta(hours=1))

    def test_invalid(self):
        with self.assertRaises(ValueError):
            parse_window_size("abc")


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


class TestBuildMultiWindowSummaries(unittest.TestCase):
    def _evt(self, ts_str, metric, value, dimension=""):
        return {
            "timestamp": datetime.fromisoformat(ts_str),
            "metric": metric,
            "value": value,
            "dimension": dimension,
            "source": "",
        }

    def test_basic(self):
        events = [
            self._evt("2026-04-01T10:00:00+00:00", "revenue", 100),
            self._evt("2026-04-01T10:02:00+00:00", "revenue", 200),
            self._evt("2026-04-01T10:06:00+00:00", "revenue", 300),
        ]
        windows = build_multi_window_summaries(events, timedelta(minutes=5))
        self.assertEqual(len(windows), 2)
        self.assertEqual(windows[0]["summary"]["count"], 2)

    def test_fill_empty(self):
        events = [
            self._evt("2026-04-01T10:00:00+00:00", "revenue", 100),
            self._evt("2026-04-01T10:12:00+00:00", "revenue", 200),
        ]
        windows = build_multi_window_summaries(events, timedelta(minutes=5), fill_empty=True)
        self.assertEqual(len(windows), 3)
        self.assertEqual(windows[1]["summary"]["count"], 0)


class TestComputeTrend(unittest.TestCase):
    def test_insufficient_data(self):
        windows = [{"summary": {"average": 100, "count": 1}}]
        self.assertEqual(compute_trend(windows), "insufficient_data")

    def test_up(self):
        windows = [
            {"summary": {"average": 100, "count": 2}},
            {"summary": {"average": 150, "count": 2}},
        ]
        self.assertEqual(compute_trend(windows), "up")

    def test_down(self):
        windows = [
            {"summary": {"average": 200, "count": 2}},
            {"summary": {"average": 100, "count": 2}},
        ]
        self.assertEqual(compute_trend(windows), "down")

    def test_flat(self):
        windows = [
            {"summary": {"average": 100, "count": 2}},
            {"summary": {"average": 102, "count": 2}},
        ]
        self.assertEqual(compute_trend(windows), "flat")


class TestDetectAlerts(unittest.TestCase):
    def test_no_alerts(self):
        windows = [
            {"window_start": "t1", "summary": {"average": 100, "std_dev": 5, "count": 2}},
            {"window_start": "t2", "summary": {"average": 110, "std_dev": 6, "count": 2}},
        ]
        alerts = detect_alerts(windows)
        self.assertEqual(len(alerts), 0)

    def test_spike(self):
        windows = [
            {"window_start": "t1", "summary": {"average": 10, "std_dev": 1, "count": 2}},
            {"window_start": "t2", "summary": {"average": 10, "std_dev": 1, "count": 2}},
            {"window_start": "t3", "summary": {"average": 500, "std_dev": 1, "count": 2}},
        ]
        alerts = detect_alerts(windows)
        types = [a["type"] for a in alerts]
        self.assertIn("spike", types)


class TestBuildReport(unittest.TestCase):
    def _evt(self, ts_str, metric, value, dimension=""):
        return {
            "timestamp": datetime.fromisoformat(ts_str),
            "metric": metric,
            "value": value,
            "dimension": dimension,
            "source": "",
        }

    def test_basic(self):
        events = [
            self._evt("2026-04-01T10:00:00+00:00", "revenue", 100, "cn"),
            self._evt("2026-04-01T10:01:00+00:00", "revenue", 200, "cn"),
            self._evt("2026-04-01T10:06:00+00:00", "revenue", 300, "cn"),
        ]
        report = build_report(events, "revenue", ws=timedelta(minutes=5))
        self.assertEqual(report["metric"], "revenue")
        self.assertEqual(report["window_size"], "5m")
        self.assertEqual(len(report["current_windows"]), 2)
        self.assertEqual(report["overall_summary"]["count"], 3)
        self.assertEqual(report["trend"], "up")

    def test_empty(self):
        report = build_report([], "revenue")
        self.assertEqual(report["overall_summary"]["count"], 0)
        self.assertEqual(report["trend"], "insufficient_data")


# ---------------------------------------------------------------------------
# Fixture-based golden / contract tests
# ---------------------------------------------------------------------------

class TestFixtureBasicCSV(unittest.TestCase):
    def test_basic_fixture(self):
        path = os.path.join(FIXTURES_DIR, "events", "basic.csv")
        events, warnings = parse_multi_csv(path)
        self.assertEqual(len(warnings), 0)
        self.assertEqual(len(events), 6)

        report = build_report(events, "revenue", ws=timedelta(minutes=5))
        self.assertEqual(len(report["current_windows"]), 2)
        self.assertEqual(report["overall_summary"]["count"], 6)
        self.assertAlmostEqual(report["overall_summary"]["sum"], 756.0)
        self.assertAlmostEqual(report["overall_summary"]["min"], 80.0)
        self.assertAlmostEqual(report["overall_summary"]["max"], 200.0)


class TestFixtureMultiMetricCSV(unittest.TestCase):
    def test_multi_metric_fixture(self):
        path = os.path.join(FIXTURES_DIR, "events", "multi_metric.csv")
        events, warnings = parse_multi_csv(path)
        self.assertEqual(len(warnings), 0)
        self.assertEqual(len(events), 9)

        revenue_report = build_report(events, "revenue", ws=timedelta(minutes=5))
        self.assertEqual(revenue_report["overall_summary"]["count"], 5)

        latency_report = build_report(events, "latency_ms", ws=timedelta(minutes=5))
        self.assertEqual(latency_report["overall_summary"]["count"], 4)


class TestFixtureInvalidRowsCSV(unittest.TestCase):
    def test_invalid_rows_fixture(self):
        path = os.path.join(FIXTURES_DIR, "events", "invalid_rows.csv")
        events, warnings = parse_multi_csv(path)
        self.assertEqual(len(events), 3)
        self.assertGreaterEqual(len(warnings), 3)
        for w in warnings:
            self.assertGreaterEqual(w["row"], 2)


class TestFixtureEmptyCSV(unittest.TestCase):
    def test_empty_fixture(self):
        path = os.path.join(FIXTURES_DIR, "events", "empty.csv")
        events, warnings = parse_multi_csv(path)
        self.assertEqual(len(events), 0)

        report = build_report(events, "revenue")
        self.assertEqual(report["overall_summary"]["count"], 0)


class TestCrossLanguageContract(unittest.TestCase):
    """Verify that Python produces the same core values as Go for the basic fixture."""

    def test_basic_summary_contract(self):
        values = [120.5, 80.0, 95.5, 110.0, 200.0, 150.0]
        s = build_summary(values)
        self.assertEqual(s["count"], 6)
        self.assertAlmostEqual(s["sum"], 756.0)
        self.assertAlmostEqual(s["min"], 80.0)
        self.assertAlmostEqual(s["max"], 200.0)
        self.assertAlmostEqual(s["average"], 126.0)
        self.assertGreater(s["variance"], 0)
        self.assertGreater(s["std_dev"], 0)
        self.assertGreaterEqual(s["p90"], s["average"])
        self.assertLessEqual(s["p90"], s["max"])


class TestCLILegacyMode(unittest.TestCase):
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

    def test_values_still_works(self):
        result = subprocess.run(
            ["python3", self._stats_py_path, "--values", "1,2,3"],
            capture_output=True,
            text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        data = json.loads(result.stdout)
        self.assertEqual(data["count"], 3)


class TestCLISubcommands(unittest.TestCase):
    _stats_py_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "stats.py")

    def test_summary_values(self):
        result = subprocess.run(
            ["python3", self._stats_py_path, "summary", "--values", "1,2,3,4,5"],
            capture_output=True,
            text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        data = json.loads(result.stdout)
        self.assertEqual(data["count"], 5)
        self.assertIn("variance", data)

    def test_summary_file(self):
        path = os.path.join(FIXTURES_DIR, "events", "basic.csv")
        result = subprocess.run(
            ["python3", self._stats_py_path, "summary", "--file", path, "--metric", "revenue"],
            capture_output=True,
            text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        data = json.loads(result.stdout)
        self.assertEqual(data["summary"]["count"], 6)

    def test_window_summary(self):
        path = os.path.join(FIXTURES_DIR, "events", "basic.csv")
        result = subprocess.run(
            ["python3", self._stats_py_path, "window-summary", "--file", path, "--metric", "revenue", "--window-size", "5m"],
            capture_output=True,
            text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        data = json.loads(result.stdout)
        self.assertIn("windows", data)
        self.assertEqual(len(data["windows"]), 2)

    def test_report(self):
        path = os.path.join(FIXTURES_DIR, "events", "basic.csv")
        result = subprocess.run(
            ["python3", self._stats_py_path, "report", "--file", path, "--metric", "revenue"],
            capture_output=True,
            text=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        data = json.loads(result.stdout)
        self.assertIn("report", data)
        report = data["report"]
        self.assertEqual(report["metric"], "revenue")
        self.assertEqual(len(report["current_windows"]), 2)
        self.assertIn("trend", report)
        self.assertIn("alerts", report)


if __name__ == "__main__":
    unittest.main()
