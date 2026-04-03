from typing import Iterable


def average(values: Iterable[float]) -> float:
    items = list(values)
    if not items:
        return 0
    return sum(items) / len(items)


def summarize(values: Iterable[float]) -> dict[str, float | int]:
    """
    计算数值集合的汇总统计信息。

    参数:
        values: 可迭代的数值集合

    返回:
        包含以下统计信息的字典：
        - count: 元素个数
        - average: 算术平均值
        - min: 最小值
        - max: 最大值
        - median: 中位数（偶数长度时取中间两个值的平均）
    """
    items = list(values)

    # 处理空列表情况
    if not items:
        return {
            'count': 0,
            'average': 0,
            'min': 0,
            'max': 0,
            'median': 0
        }

    # 计算中位数需要排序
    sorted_items = sorted(items)
    n = len(items)

    # 计算中位数
    if n % 2 == 1:
        # 奇数个元素：取中间值
        median_val = sorted_items[n // 2]
    else:
        # 偶数个元素：取中间两个值的平均
        median_val = (sorted_items[n // 2 - 1] + sorted_items[n // 2]) / 2

    return {
        'count': n,
        'average': average(items),  # 复用现有函数
        'min': min(items),
        'max': max(items),
        'median': median_val
    }
