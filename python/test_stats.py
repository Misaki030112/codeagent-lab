"""Tests for the stats module."""

import unittest

from stats import average, build_summary, median, percent_change


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
        for key in ("count", "sum", "min", "max", "average", "median"):
            self.assertIn(key, s)

    def test_generator_input(self):
        s = build_summary(x for x in [2, 4, 6])
        self.assertEqual(s["count"], 3)
        self.assertAlmostEqual(s["average"], 4.0)


if __name__ == "__main__":
    unittest.main()
