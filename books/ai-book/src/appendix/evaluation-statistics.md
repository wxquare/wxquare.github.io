# 附录D 评估统计复算

本附录保存第 10 章统计例的完整可运行参考。它帮助复算二项零失败上界、配对二元结果的保守区间和样本量数量级，不构成通用统计软件包。使用者必须先确定独立单位、配对关系、抽样方式、停止规则和假设族；聚类样本、反复查看或数据后选需要另行设计。

## 方法选择

| 问题 | 建议起点 | 不能省略的前提 |
|:---|:---|:---|
| n 次独立验收零失败，失败率有多大 | 二项精确单侧上界 | 独立同分布、固定窗口 |
| 同一批 case 比较候选与基线 | 配对四格、配对精确或 score 区间 | case 一一对应、方向预先定义 |
| 只需一个分布无关保守参照 | 有界配对差的 Hoeffding 界 | 独立单位；接受区间可能很宽 |
| 规划独立两组差异检测 | 正态近似样本量或模拟 | 这不是配对非劣效规划 |
| 多切片或重复查看 | 预先声明假设族与错误预算 | 不能事后选择有利切片 |

## 可运行 Python

```python
from math import exp, fsum, lgamma, log, log1p

def _binom_cdf(k, n, p):
    if k < 0:
        return 0.0
    if k >= n or p == 0:
        return 1.0
    if p == 1:
        return 0.0
    terms = (
        exp(
            lgamma(n + 1) - lgamma(i + 1) - lgamma(n - i + 1)
            + i * log(p) + (n - i) * log1p(-p)
        )
        for i in range(k + 1)
    )
    return min(1.0, fsum(terms))

def _cdf_root(k, n, target):
    lo, hi = 0.0, 1.0
    for _ in range(70):
        mid = (lo + hi) / 2
        if _binom_cdf(k, n, mid) > target:
            lo = mid
        else:
            hi = mid
    return (lo + hi) / 2

def _cell_bounds(k, n, tail):
    low = 0.0 if k == 0 else _cdf_root(k - 1, n, 1 - tail)
    high = 1.0 if k == n else _cdf_root(k, n, tail)
    return low, high

def paired_binary_interval(pairs, alpha=0.05):
    if not 0 < alpha < 1 or not pairs:
        raise ValueError("nonempty independent pairs and valid alpha required")
    if any(a not in (0, 1) or b not in (0, 1) for a, b in pairs):
        raise ValueError("binary outcomes required")
    n = len(pairs)
    gain = sum(a == 0 and b == 1 for a, b in pairs)
    loss = sum(a == 1 and b == 0 for a, b in pairs)
    gain_lo, gain_hi = _cell_bounds(gain, n, alpha / 4)
    loss_lo, loss_hi = _cell_bounds(loss, n, alpha / 4)
    return {
        "method": "paired-binomial-bonferroni",
        "alpha": alpha,
        "unit": "independent_case",
        "n": n,
        "gain": gain,
        "loss": loss,
        "delta": (gain - loss) / n,
        "low": max(-1.0, gain_lo - loss_hi),
        "high": min(1.0, gain_hi - loss_lo),
    }

def zero_event_bound(n, alpha=0.05):
    if not isinstance(n, int) or n <= 0 or not 0 < alpha < 1:
        raise ValueError("positive independent sample count and valid alpha required")
    return {
        "method": "zero-event-binomial-upper",
        "alpha": alpha,
        "n": n,
        "upper": 1 - alpha ** (1 / n),
    }
```

## 交叉校验

- 零失败：n=100、95% 单侧上界为约 2.95%。
- 配对零翻转：n=320、alpha=.05 时，上述 Bonferroni 单元格组合约为 ±1.36 个百分点；这不是任意配对区间都共有的结论。
- NIST 二项例：n=20、事件数 4、90% 区间约为 [0.071354, 0.401029]。
- 独立两组近似：p=.8、绝对差 .05、双侧 alpha=.05、功效 .8 时每组约 1004；不能把它移作配对单侧非劣效样本量。

来源：NIST/SEMATECH 的[比例置信区间](https://www.itl.nist.gov/div898/handbook/prc/section2/prc241.htm)与 [Bonferroni 方法](https://www.itl.nist.gov/div898/handbook/prc/section4/prc473.htm)，核验日期 2026-10-25。配对单元格组合为本书教学推导。
