from typing import Iterable


def average(values: Iterable[float]) -> float:
    items = list(values)
    if not items:
        return 0
    return sum(items) / len(items)
