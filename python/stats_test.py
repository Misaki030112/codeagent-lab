import unittest
from stats import average, summarize


class TestStatsAverage(unittest.TestCase):
    """现有 average() 函数的测试"""

    def test_average_empty_list(self):
        """测试空列表返回 0"""
        self.assertEqual(average([]), 0)

    def test_average_normal_case(self):
        """测试正常情况下的平均值计算"""
        self.assertEqual(average([1, 2, 3, 4, 5]), 3.0)


class TestStatsSummarize(unittest.TestCase):
    """新增 summarize() 函数的测试"""

    def test_empty_list(self):
        """测试空列表：所有值应为 0"""
        result = summarize([])
        self.assertEqual(result['count'], 0)
        self.assertEqual(result['average'], 0)
        self.assertEqual(result['min'], 0)
        self.assertEqual(result['max'], 0)
        self.assertEqual(result['median'], 0)

    def test_single_element(self):
        """测试单元素：count=1，所有统计值等于该元素"""
        result = summarize([5])
        self.assertEqual(result['count'], 1)
        self.assertEqual(result['average'], 5)
        self.assertEqual(result['min'], 5)
        self.assertEqual(result['max'], 5)
        self.assertEqual(result['median'], 5)

    def test_odd_length_list(self):
        """测试奇数长度列表：中位数为中间值"""
        result = summarize([1, 2, 3, 4, 5])
        self.assertEqual(result['count'], 5)
        self.assertEqual(result['average'], 3.0)
        self.assertEqual(result['min'], 1)
        self.assertEqual(result['max'], 5)
        self.assertEqual(result['median'], 3)

    def test_even_length_list(self):
        """测试偶数长度列表：中位数为中间两值的平均"""
        result = summarize([1, 2, 3, 4])
        self.assertEqual(result['count'], 4)
        self.assertEqual(result['average'], 2.5)
        self.assertEqual(result['min'], 1)
        self.assertEqual(result['max'], 4)
        self.assertEqual(result['median'], 2.5)

    def test_float_values(self):
        """测试浮点数处理"""
        result = summarize([1.5, 2.5, 3.5])
        self.assertEqual(result['count'], 3)
        self.assertAlmostEqual(result['average'], 2.5)
        self.assertEqual(result['min'], 1.5)
        self.assertEqual(result['max'], 3.5)
        self.assertEqual(result['median'], 2.5)

    def test_negative_values(self):
        """测试负数处理"""
        result = summarize([-5, -2, 0, 3, 10])
        self.assertEqual(result['count'], 5)
        self.assertAlmostEqual(result['average'], 1.2)
        self.assertEqual(result['min'], -5)
        self.assertEqual(result['max'], 10)
        self.assertEqual(result['median'], 0)


if __name__ == '__main__':
    unittest.main()
