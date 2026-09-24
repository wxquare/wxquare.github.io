---
title: TVM 算子优化实战：从 Relay 图优化到 GEMM 与 INT8 量化
date: 2026-09-22
categories:
  - AI 与 Agent
tags:
  - tvm
  - deep-learning-compiler
  - gemm
  - model-quantization
mathjax: true
---

本文把过去几篇 TVM 学习笔记合并成一条完整的实践主线：先建立 TensorFlow 基线，再观察 TVM 如何导入和优化计算图，接着深入 GEMM 算子调度，最后讨论 INT8 量化和综合测试。文章不把“编译成功”当作“性能一定更好”，而是把优化拆成可测量、可解释的实验。

## 1. 为什么从推理引擎转向深度学习编译器

在模型部署中，模型结构、算子实现、内存布局、线程调度和目标硬件往往同时影响延迟。仅修改网络结构并不能保证端到端加速：非结构化稀疏需要运行时和硬件真正跳过零值，量化需要整数算子和合适的校准，剪枝也需要后端识别新的张量形状。过去的实验中，直接使用 TensorFlow 推理、尝试权重量化、稀疏和通道剪枝，都能说明一个事实：优化手段必须和执行引擎配套。

TVM 的价值在于把模型表示、图级变换、算子程序生成和运行时部署放在同一套编译流程中。原始 TVM 论文把这种思路描述为面向深度学习的端到端优化编译器[1]。这并不意味着 TVM 自动解决所有性能问题，而是把很多过去需要手写内核、适配不同平台的工作，转化为可以检查和搜索的中间表示与调度问题。

本文的实验问题有三个：

1. 直接把模型交给 TVM，收益来自哪里？
2. GEMM 为什么是最适合观察调度优化的算子？
3. 图优化、算子优化和 INT8 量化叠加后，怎样判断收益是否真实？

## 2. 实验设计：先固定基线，再逐层增加优化

性能测试最容易犯的错误是比较了不同的输入、线程数、预热方式或编译目标。每次实验都应记录硬件、操作系统、编译器、TVM 版本、模型版本、输入形状、数据类型、线程数和测量方法。旧文章中的测试来自 2020 年的 x86_64 Ubuntu 环境，不能直接当成当前机器上的结论；本文保留它们作为历史基线，并建议在当前环境重新运行。

推荐将实验分成六组：

| 组别 | 执行路径 | 目的 |
|---|---|---|
| A | TensorFlow 原始推理 | 建立业务基线 |
| B | TVM 直接导入和编译 | 观察编译器基础收益 |
| C | B + Relay 图级优化 | 分离算子融合和常量优化收益 |
| D | C + GEMM 调度优化 | 观察内核级优化收益 |
| E | C + INT8 量化 | 观察数据类型变化的收益和精度代价 |
| F | D + E | 评估组合优化是否存在重复收益或新瓶颈 |

单次测量应包含预热阶段，正式阶段至少重复多次并报告中位数和离散程度。对 CPU 测试，还要固定线程数和 CPU 亲和性；对 GPU 测试，要区分首次编译、首次运行和稳定运行。端到端延迟应包括数据准备和输出同步，内核微基准则可以单独报告，以免把两种指标混为一谈。

## 3. TVM 的编译执行链路

TVM 可以粗略看成两层。第一层接收模型计算图，完成前端导入、类型推导、图级 Pass 和算子拆分；第二层把算子表达成 TensorIR 或底层函数，经过调度、目标相关的代码生成，最后交给运行时模块执行。Relay 是一种面向机器学习程序的函数式中间表示，适合表达张量计算、函数调用和部分控制流[2][3]。

在旧版 API 中，`tvm.build` 常被用于从 Tensor Expression 或低层 IR 生成算子模块，`tvm.relay.build` 则负责从 Relay 模块生成模型运行时。新版本将 IRModule、TensorIR 和统一编译接口进一步整合，因此不能机械照搬 0.6/0.7 时代的代码。阅读旧代码时应先确认 API 版本，再检查生成的 IR 和目标代码。

最有用的调试路径不是只看最终性能，而是逐层打印：

```python
print(mod.astext())
lib = tvm.build(schedule_or_ir, target=target)
print(lib.get_source())
```

前者帮助确认 Pass 是否真的改变了计算图，后者帮助确认是否生成了期望的循环、向量指令或目标平台代码。TVM 的 IRModule、Runtime Module、PrimFunc 和 Relay Function 分别处在不同抽象层，不应把它们当成同一个对象[4][5][6][7]。

## 4. Relay 图级优化：先减少不必要的计算

Relay 的图级优化通常包括常量折叠、算子融合、死代码消除、布局转换和类型相关的变换。图优化的主要目标不是让单个加法指令更快，而是减少中间张量写回、降低函数调用和调度开销，并把一串适合融合的操作交给后端共同生成。算子融合是否有效取决于张量尺寸、内存带宽、目标硬件和后端实现，不能只看 Pass 名称。

一个典型流程可以写成：

```python
with tvm.transform.PassContext(opt_level=3):
    built = relay.build(mod, target=target, params=params)
```

实际项目中应把每个 Pass 的输入和输出保存下来，对比以下问题：

- 是否消除了恒定表达式和无用分支？
- 是否把连续的逐元素操作融合了？
- 融合后是否产生过大的函数，导致寄存器压力上升？
- 布局转换是否抵消了融合收益？
- 动态形状或控制流是否限制了优化？

XLA 的 HLO、MLIR 的 Linalg/Tensor 方言和 Halide 的调度模型提供了有价值的参照：计算描述与执行策略应该分离，编译器才能在目标硬件变化时重新选择实现[8][9][10][11]。TVM 的 Relay 不应被理解为“另一个 Python 网络框架”，它更接近可以被分析和变换的程序表示。

## 5. GEMM 为什么适合做优化实验

GEMM 表示通用矩阵乘法，基本形式是 `C = alpha * A * B + beta * C`。深度学习中的全连接层、卷积降低后的矩阵乘法、注意力中的部分计算，都可能归约为 GEMM。矩阵乘法的计算量容易估算，结果容易校验，性能也会明显受到缓存、SIMD、线程并行和内存布局影响，因此适合作为编译器优化的观察窗口。

对 `M x K` 与 `K x N` 的矩阵相乘，主要计算量约为 `2*M*K*N` 次浮点运算。理论峰值可以帮助判断结果是否合理，但理论峰值不是实际可达性能。实际效率还受缓存容量、访存带宽、指令宽度、线程同步、矩阵边界和库调用开销影响。旧实验以 2.4 GHz、AVX2、FMA 的 CPU 为例估算单核峰值，这个估算只能作为当时机器的上界，重新测试时必须根据当前 CPU 的 SIMD 宽度和核心频率更新。

GEMM 优化通常围绕四件事展开：

1. 分块：让工作集尽量落入 L1/L2/L3 缓存。
2. 循环重排：让连续访问方向与存储布局一致。
3. 向量化：把标量乘加转换为 SIMD 指令。
4. 并行化：把独立的输出块分配给多个线程，同时控制同步开销。

如果矩阵很小，函数调用和线程启动的成本可能超过计算本身；如果矩阵很大，带宽和缓存层级会主导结果。因此必须覆盖小、中、大多组尺寸，而不能只报告一个最漂亮的数字。

## 6. 在 TVM 中调度 GEMM

一个简化的 Tensor Expression 可以表达矩阵乘法：

```python
M, N, K = 1024, 1024, 1024
A = te.placeholder((M, K), name="A")
B = te.placeholder((K, N), name="B")
k = te.reduce_axis((0, K), name="k")
C = te.compute((M, N), lambda i, j: te.sum(A[i, k] * B[k, j], axis=k), name="C")
```

这段代码只描述“算什么”，没有说明循环如何执行。调度负责回答“怎么执行”：

```python
s = te.create_schedule(C.op)
yo, yi = s[C].split(C.op.axis[0], factor=tile_m)
xo, xi = s[C].split(C.op.axis[1], factor=tile_n)
ko, ki = s[C].split(k, factor=tile_k)
s[C].reorder(yo, xo, ko, yi, xi, ki)
```

分块参数不能凭经验固定。`tile_m`、`tile_n` 和 `tile_k` 需要结合缓存、向量宽度、数据类型和线程数测试。过小的块会增加循环控制开销，过大的块会造成缓存冲突或寄存器压力。CPU 后端通常还需要考虑向量化轴的连续性，GPU 后端则要考虑线程块、共享内存、warp 和同步。

TVM 的 AutoTVM、Ansor 和 MetaSchedule 体现了自动搜索的不同阶段。AutoTVM 依赖人工定义模板和搜索空间；Ansor 试图自动生成更广泛的程序和搜索任务[12]；MetaSchedule 则在 TensorIR 和统一搜索流程上进一步发展[13][14]。自动搜索并不是魔法：搜索目标、测量噪声、候选空间、硬件频率波动和正确性检查，都会决定最终结果。

GEMM 测试必须同时做数值校验：将 TVM 输出与 NumPy 或高质量 BLAS 结果比较，设置相对误差和绝对误差阈值。性能更快但结果错误的内核没有工程价值。对于 FP32、FP16 和 INT8，误差阈值不能共用；量化还要报告校准集和准确率变化。

## 7. 代码生成与运行时模块

`build` 的作用不只是“编译一下”。它把高层计算描述、目标信息和调度结果转换为一个可加载的运行时模块。`IRModule` 可能包含多个函数，`PrimFunc` 更接近单个低层张量程序，`runtime.Module` 则负责保存生成后的函数和目标代码。调试时应检查：

- IR 中的循环边界是否符合预期；
- reduce 轴是否被错误地移动或重复计算；
- 目标是否真的启用了所需指令集；
- `get_source()` 输出的代码是否与目标平台一致；
- 运行时输入输出 dtype 和布局是否匹配。

LLVM 是 CPU 代码生成的重要基础，CUDA、OpenCL 和 Metal 等目标则有自己的运行时和内存模型[15][16][17]。同一套 Relay 图在不同目标上获得不同性能是正常现象；跨平台的价值是复用模型和编译流程，不是保证所有平台共享同一组最佳调度参数。

## 8. INT8 量化实践

量化通常把 FP32 激活和权重映射到更低位宽的整数表示。最简单的线性映射可以写成：

`real_value = scale * (integer_value - zero_point)`。

对称量化简化了硬件计算，但可能浪费非对称分布中的数值范围；非对称量化能更充分利用范围，却会引入 zero point 处理。量化收益不仅取决于模型大小减少，还取决于硬件是否有高效的整数向量指令，以及量化算子是否覆盖完整路径。

量化流程至少包含：选择量化粒度、准备校准数据、统计激活范围、插入量化/反量化节点、编译整数算子、验证精度和测试延迟。TensorFlow Lite 和 ONNX Runtime 的官方量化文档都强调校准、算子支持和精度验证的重要性[18][19]。TVM 的量化接口和 RFC 则反映了特定版本的实现状态，不能把 2020 年的 RFC 当成当前 API 文档[20][21]。

量化实验应同时比较：

| 指标 | FP32 | INT8 |
|---|---:|---:|
| 模型或权重大小 | 记录实测 | 记录实测 |
| 单次延迟 | 记录中位数 | 记录中位数 |
| 吞吐 | 记录稳定值 | 记录稳定值 |
| 精度 | 原始精度 | 与基线比较 |
| 算子覆盖率 | 统计 | 统计 |

如果只有部分算子被量化，模型可能频繁在 FP32 和 INT8 之间转换，端到端收益会低于单个整数算子的微基准收益。量化与 GEMM 优化也可能产生重叠收益：整数 GEMM 更快不代表整个模型同样加速，瓶颈可能转移到数据布局转换、内存搬运或未量化算子。

## 9. 综合结果应如何解读

建议把最终结果整理成一张表，而不是只给出“快了几倍”：

| 方案 | 图优化 | GEMM 调度 | 数据类型 | 延迟 | 精度变化 | 主要瓶颈 |
|---|---|---|---|---:|---:|---|
| TensorFlow | - | 库实现 | FP32 | 基线 | 0 | 框架与算子执行 |
| TVM Basic | 基础 | 默认 | FP32 | 实测 | 0 | 默认算子 |
| TVM Relay | 有 | 默认 | FP32 | 实测 | 0 | 未优化热点 |
| TVM GEMM | 有 | 调优 | FP32 | 实测 | 0 | 非 GEMM 算子 |
| TVM INT8 | 有 | 默认/调优 | INT8 | 实测 | 实测 | 量化覆盖率 |
| Combined | 有 | 调优 | INT8 | 实测 | 实测 | 端到端数据流 |

结果分析应该回答“为什么”，例如：Relay 融合减少了中间张量写回，GEMM 分块提升了缓存复用，INT8 减少了带宽和计算成本；也要回答“为什么没有更快”，例如矩阵过小、线程开销占比高、目标 CPU 未启用 AVX2、量化覆盖不足或数据转换成为新瓶颈。

性能测试还需要区分算子微基准和模型端到端基准。前者适合验证调度，后者适合判断业务价值。二者不一致并不矛盾，反而能帮助定位优化收益在链路中的损失位置。

## 10. 技术复盘：TVM 适合解决什么问题

这几篇历史笔记最有价值的地方不是某一次“比 TensorFlow 快两倍”的结果，而是展示了从模型压缩走向编译器优化的学习路径。结论可以归纳为四点。

第一，优化必须从基线开始。没有固定版本、输入、线程和测量方法，性能数字无法比较。第二，GEMM 优化是硬件相关的工程问题。分块、向量化和并行化有通用原则，但最佳参数依赖缓存、指令集、矩阵尺寸和线程运行时。第三，量化不是把 dtype 改成 int8 就完成了，校准、算子覆盖、精度和数据转换共同决定收益。第四，自动调优降低了手写搜索的成本，但需要正确的搜索空间、可靠的测量和严格的数值校验。

对于固定模型、固定硬件并且愿意投入编译优化的团队，TVM 可以把模型部署和算子调度统一起来。对于动态形状多、模型变化快、硬件已有成熟高性能库的场景，直接使用厂商库或成熟运行时可能更合适。选择 TVM 的依据应是整个生命周期成本，而不是单次 benchmark 的峰值。

## 11. 复现实验清单

重新运行本文实验前，应记录：

- TVM、LLVM、Python、NumPy 和模型框架版本；
- CPU/GPU 型号、核心数、缓存、SIMD 或 CUDA 能力；
- 编译目标、优化级别、线程数和 CPU 亲和性；
- 矩阵尺寸、batch、布局、dtype 和随机种子；
- 预热次数、正式迭代次数、统计口径和异常值处理；
- 输出误差阈值、校准数据和精度评估集。

历史文章中的链接和代码仍放在归档目录中，本文只保留能够支撑主线的内容。旧 API 示例需要在对应 TVM 版本环境中运行；如果无法恢复旧环境，应把它们当作概念示例，并补充当前 API 的迁移说明。

## 12. GEMM 调优中的内存层次

矩阵乘法的难点不在于写出三个循环，而在于让每一级存储层次都能持续提供数据。最直接的实现按照 `i、j、k` 顺序访问，可能反复从较慢的缓存读取同一行或同一列。分块之后，外层循环负责选择一个输出块，内层循环让 A、B 和 C 的局部工作集在一段时间内保持活跃。这样做的本质不是减少数学运算，而是减少昂贵的数据搬运。

在 CPU 上，L1 缓存容量通常很小，适合保存当前微内核使用的 A 和 B 子块；L2 缓存可以承载更大的线程局部块；共享的 L3 缓存则影响多个线程之间的数据复用。块大小过大时，虽然单个块的计算量增加，但缓存淘汰也会增加；块大小过小时，循环控制和边界处理占比上升。实际调优不能只根据缓存总容量除以数据大小计算，还要考虑缓存组冲突、其他线程占用和运行时栈。

矩阵的存储布局同样重要。行主序矩阵中，沿最后一维访问通常是连续的；如果内层循环访问列方向，就会产生较大的步长。转置 B 或者在加载时改变访问顺序，可能比单纯调整循环顺序更有效。对于非方阵和尾部块，不能为了追求整齐而越界读取；可以使用边界判断、填充或单独的尾块内核，但三者会带来不同的开销。

## 13. 向量化、微内核与并行策略

向量化需要满足两个条件：数据访问足够连续，且循环迭代之间没有阻碍向量指令重排的依赖。GEMM 的归约轴通常可以被展开，多个乘加操作同时进行。FMA 指令能够把乘法和加法合成一条指令，但最终性能还要看寄存器数量、加载吞吐和后端是否生成了正确的指令。仅仅在代码中写出“vectorize”并不能证明生成了 SIMD，需要检查目标代码或使用硬件计数器。

常见的高性能 GEMM 会使用微内核。微内核固定计算一个小的 `m_r x n_r` 输出区域，让多个累加器常驻寄存器中，然后循环读取 A 和 B 的一小段数据。外层分块负责把大矩阵切成适合微内核的块，内层微内核负责最大化乘加吞吐。TVM 调度可以表达这种层级，但参数选择通常依赖目标 CPU 的向量宽度、寄存器数量和编译器能力。

多线程切分时，优先把互不重叠的输出块分给不同线程，避免写冲突。线程粒度过细会让调度器、屏障和线程唤醒成本变得明显；粒度过粗则会造成负载不均。矩阵很小的时候，单线程可能反而更快。矩阵很大时，还要观察线程数增加后是否已经受到内存带宽限制。一个好的测试矩阵应该同时包含单线程、物理核心数和超线程数几种配置。

## 14. 自动调优的正确使用方式

自动调优前应先建立可靠的正确性测试和合理的搜索边界。把所有可能的 tile、展开因子、线程数和布局组合全部放进搜索空间，往往会导致搜索时间过长，而且容易在测量噪声中选出偶然的最优值。更好的方法是先用硬件常识排除明显不合理的组合，再让搜索器解决人工难以判断的细节。

搜索日志必须记录目标硬件和 TVM 版本。相同的调度在不同 CPU 上可能完全不同，甚至同一 CPU 在频率策略、散热状态或后台负载变化时也会出现波动。测量前需要预热，测量中要控制环境，失败任务要区分编译失败、运行失败、超时和数值校验失败。把所有失败都当作慢结果会污染搜索模型。

AutoTVM、Ansor 和 MetaSchedule 的设计反映了自动优化的发展过程，但使用者仍然需要理解任务定义。搜索器只能优化被暴露出来的计算和参数，不能自动弥补错误的布局、缺失的算子实现或端到端数据转换开销。最稳妥的流程是先用一个可读的人工调度得到基线，再逐步扩大搜索空间，并保留人工基线用于回归比较。

## 15. Relay 融合的收益和代价

算子融合减少中间张量写入，是图级优化最容易观察到的收益之一。例如，连续的加法、乘法和激活函数可以合并到同一个循环中，避免每一步都把完整张量写回内存。但是融合不是越多越好。过度融合可能产生很大的函数，增加寄存器压力，使指令缓存变差，也可能让某些算子无法调用已经高度优化的厂商库。

融合还会改变调优边界。没有融合时，GEMM 可以独立使用高性能库；融合之后，GEMM 可能成为更大函数中的一部分，后端需要在融合收益和专用内核收益之间做选择。对于带有动态形状、控制流或不规则索引的模型，融合规则也可能更加保守。分析 Relay 图时，应同时看节点数量、张量大小、布局和最终生成函数，而不能只统计 Pass 执行成功次数。

建议为每个重要 Pass 建立前后对照：保存文本 IR、列出算子数量、估算中间张量大小，并记录端到端延迟。这样才能解释一次性能变化到底来自算子融合、常量折叠、布局改变，还是其他编译阶段的副作用。

## 16. 量化误差的定位方法

量化精度下降时，不应只看最终准确率。可以逐层比较 FP32 和 INT8 的输出范围、均值、方差、最大误差和饱和比例，定位误差从哪一层开始放大。激活分布存在少量离群值时，使用全局最小最大值会压缩大多数样本的有效精度；使用百分位数校准或逐通道权重量化，可能改善结果，但也增加部署复杂度。

权重和激活的量化粒度不同。权重通常可以按张量或按输出通道量化，激活则常常依赖运行时统计。对称量化便于整数乘加，非对称量化能覆盖偏移明显的分布。选择方案时要看后端指令和算子实现是否支持，而不能只根据理论误差公式判断。

量化后的性能还要检查数据转换。一个模型如果在每个小算子之间频繁执行反量化，可能比完全使用 FP32 更慢。真正有价值的 INT8 路径需要较高的算子覆盖率，并尽量保持中间张量在整数域中流动。对于无法量化的算子，应明确列出原因和占比，而不是把整个模型标记为“INT8”。

## 17. 版本迁移与旧代码阅读

历史笔记使用的是 2020 年前后的 TVM 接口，今天阅读时最容易遇到的是模块路径、Relay API、量化接口和调度抽象变化。迁移旧示例时，第一步应固定旧版本环境并运行原始代码，第二步记录实际依赖的行为，第三步用当前官方文档寻找对应 API。不要只根据函数名机械替换，因为同名接口的输入类型和返回对象可能已经改变。

旧代码中还可能存在 Markdown 缩进代码块、过期文档链接、错误拼写和硬件名称问题。重构时应把概念性代码改成带语言标记的代码块，给出当前版本说明；对于无法在当前版本重现的片段，保留其历史背景但不要把它作为当前推荐方案。性能数字必须标注采集日期和环境，避免读者把历史结果理解为保证值。

## 18. 从算子微基准到端到端服务

算子微基准回答的是“这个内核在固定输入上有多快”，服务端性能还受到模型加载、内存分配、输入预处理、线程池、批处理、请求排队和输出后处理影响。TVM 编译出的模块可能降低算子执行时间，但如果预处理占了大部分延迟，整体收益就不会按比例增长。

端到端测试至少要区分冷启动、热运行、单请求、固定批次和动态批次。在线服务还应关注 P50、P95 和 P99，而不是只报告平均值。批处理可能提高吞吐，却增加单请求延迟；线程数增加可能提高并发能力，却造成缓存争用。最终方案应根据业务目标选择指标，并把算子优化放进完整的性能预算中。

## 19. 可复现性与结果可信度

一条性能结论至少需要三类证据：代码或配置、测试环境、重复测量结果。只有截图没有命令，只有单次耗时没有统计分布，只有理论峰值没有实际计数器，都不足以支撑强结论。建议保存编译目标、调度参数、输入生成方式、输出校验结果和原始测量日志。

如果实验无法重新运行，应在正文中使用“在历史环境中观察到”而不是“TVM 一定能够”。如果某个公开参考源只说明算法原理，不说明当前 TVM API，也应把它放在原理章节而不是实现章节。通过区分观察、推断和建议，文章可以同时保留实践价值和技术可信度。

## 20. 历史实验数据：从 TensorFlow 到 TVM

早期实验是在 Ubuntu 19.04、x86_64 环境中完成的。原始文章记录的 TensorFlow 推理时间为 0.586195 秒，TVM 推理时间为 0.277053 秒，约为 2.12 倍的速度差异。两套结果的分类输出基本一致，例如最高结果都是 African elephant 和 tusker。这个结果说明当时的 TVM 导入、编译和运行时路径具有明显收益，但它只代表当时的模型、输入图片、线程设置和软件版本，不能直接推广到今天的 TVM 或所有模型。

原始实验还记录了模型编译过程：安装 LLVM，递归克隆 TVM，复制 `cmake/config.cmake`，通过 `USE_LLVM` 指向 `llvm-config`，再配置 Python 环境。旧版示例曾遇到 antlr 运行时缺失和 HTTPS 下载问题。当前版本的安装方法已经变化，因此这些内容保留为历史排障记录，新的实验应优先遵循 Apache TVM 官方安装文档。

```bash
sudo apt-get install llvm
git clone --recursive https://github.com/apache/tvm.git
cd tvm
mkdir build
cp cmake/config.cmake build
cd build
cmake ..
cmake --build . --parallel 4
```

旧实验还遇到 AutoTVM 找不到 LLVM target 配置的问题。日志中的典型提示是 `Cannot find config for target=llvm`，随后使用 fallback configuration，这可能造成明显的性能回退。这个问题本身很重要：自动调优没有命中有效配置时，测到的不是“TVM 的最佳性能”，而是回退实现的性能。新文章应把 target、调优日志和 fallback 状态一并记录。

## 21. 原始 GEMM 基线数据

原始 GEMM 实验以一台 2.4 GHz、支持 AVX2 和 FMA 的多核服务器为背景，使用 NumPy 测试不同矩阵规模。按照当时的估算，FP32 单核理论峰值为 `2.4G * (8+8) * 2 = 76.8 GFLOPS`，FP64 单核理论峰值为 `2.4G * (4+4) * 2 = 38.4 GFLOPS`。这里的 8 和 4 反映每条向量指令能够处理的元素数量，最后的 2 反映 FMA 的乘加操作。这个估算是解释性能上限的工具，不是实测值。

原始 NumPy 测试数据如下，单位沿用旧文章的 GFLOPS 表达方式：

| 配置 | 32 | 128 | 1024 | 2048 | 4096 | 10240 | 硬件利用率 |
|---|---:|---:|---:|---:|---:|---:|---:|
| 单核 FP32 | 1.82 | 36.16 | 67.99 | 67.94 | 68.88 | 69.88 | 91.0% |
| 单核 FP64 | 1.67 | 19.49 | 35.56 | 35.40 | 36.11 | 36.90 | 96.1% |
| 四核 FP32 | 6.60 | 52.20 | 225.42 | 246.20 | 244.20 | 256.00 | 83.8% |
| 四核 FP64 | 5.56 | 37.62 | 116.42 | 120.39 | 127.03 | 141.15 | 91.9% |

这些数据呈现出两个值得保留的现象。第一，小矩阵受到函数调用、初始化和循环开销影响，硬件利用率明显较低；第二，四核 FP32 的利用率低于单核，说明增加线程后不一定线性扩展，内存、线程调度和共享缓存都会成为限制。重新测试时应补充 NumPy 使用的底层 BLAS 库和线程配置，否则不同环境的 NumPy 结果不能直接比较。

## 22. 原始 GEMM 手工优化阶梯

旧文章没有只给出最终结果，而是记录了从朴素实现逐步加入优化的过程。测试矩阵约为 1024×1024，表中的时间和硬件利用率沿用历史记录：

| 实现 | 64 | 256 | 512 | 1024 | 硬件利用率 | 主要优化 |
|---|---:|---:|---:|---:|---:|---|
| MMult0 | 1.51 | 0.79 | 0.66 | 0.65 | 1.69% | 基线 |
| MMult_1x4_5 | 2.15 | 1.08 | 0.72 | 0.716 | 2.6% | 一次计算 1×4 |
| MMult_1x4_9 | 4.90 | 3.15 | 3.10 | 3.14 | 8.18% | 1×4、寄存器 |
| MMult_4x4_5 | 2.76 | 1.53 | 1.26 | 1.26 | 3.28% | 一次计算 4×4 |
| MMult_4x4_9 | 5.19 | 2.92 | 2.88 | 2.87 | 7.47% | 4×4、寄存器 |
| MMult_4x4_10 | 5.95 | 4.16 | 4.04 | 4.01 | 10.4% | 4×4、寄存器、SSE |
| MMult_4x4_10_1 | 10.0 | 6.60 | 6.35 | 6.40 | 16.7% | 4×4、寄存器、FMA |
| MMult_4x4_11_1 | 14.5 | 8.95 | 7.16 | 7.08 | 18.4% | FMA、分块、缓存 |
| MMult_4x4_15_1 | 11.3 | 11.6 | 11.7 | 11.7 | 30.4% | FMA、分块、内存顺序 |

这组数据最适合用来解释优化的因果关系：寄存器块提高了同一数据的复用，SSE/FMA 提高了单位时间的计算量，分块改善了缓存局部性，内存访问顺序则决定了加载是否连续。最后一行比前一行更快并不意味着每一步都单调改善；参数变化会改变缓存和寄存器压力，调优必须保留完整的实验阶梯。

## 23. TVM GEMM 优化的阶段性结果

原始 TVM 实验在 1024×1024 矩阵上记录了从默认调度到 AutoTVM 的阶段结果：

| 阶段 | TVM 时间 | NumPy 时间 | 阶段含义 |
|---|---:|---:|---|
| baseline | 2.49 s | 0.0135 s | 默认循环 |
| blocking | 1.73 s | 0.0120 s | 分块 |
| vectorization | 0.411 s | 0.0117 s | 向量化 |
| loop permutation | 0.104 s | 0.0116 s | 循环重排 |
| packing | 0.0987 s | 0.0103 s | 打包布局 |
| write cache | 0.0926 s | 0.01158 s | 写缓存 |
| parallel | 0.018 s | 0.0120 s | 并行 |
| AutoTVM | 0.014 s | 0.0112 s | 自动调优 |

这张表应该保留在合并文章中，因为它清楚展示了“写出 GEMM”与“写出高性能 GEMM”的差别。baseline 到 blocking 只获得有限改善，向量化和循环重排带来更明显的下降，并行化让结果接近 NumPy，AutoTVM 在这个特定 workload 上进一步缩小差距。表中数值来自历史环境，文章不应把它们当作当前版本的保证值；更严谨的做法是用同样的阶段重新测量，并增加 P50/P95、线程数和目标信息。

## 24. Relay 原始示例和 Pass 清单

原始 Relay 笔记首先用标量变量、常量和加法构造 Hello Relay，再用卷积单元说明 Relay 如何表达神经网络计算。它强调 Relay 的两个特点：一是计算图可以以文本 IR 的形式检查，二是图优化通过一组 Pass 实现，而不是把所有优化硬编码在模型定义中。

历史笔记列出的 Pass 包括 `Legalize`、`SimplifyInference`、`EliminateCommonSubexpr`、`CombineParallelConv2D`、`CombineParallelDense`、`FoldConstant`、`FoldScaleAxis`、`CanonicalizeCast`、`CanonicalizeOps`、`AlterOpLayout` 和 `FuseOps`。这些名称和当前版本可能不同，但覆盖的优化意图仍然值得保留：目标相关合法化、推理简化、公共子表达式消除、并行算子合并、常量折叠、布局变换和算子融合。

一个更完整的检查方式是保存每个阶段的 `IRModule`：

```python
from tvm import relay

seq = tvm.transform.Sequential([
    relay.transform.InferType(),
    relay.transform.FoldConstant(),
    relay.transform.FuseOps(),
])
with tvm.transform.PassContext(opt_level=3):
    optimized = seq(mod)
print(optimized.astext(show_meta_data=False))
```

实际使用时应以当前版本的 Relay API 为准，并检查每个 Pass 的适用条件。原始文章在 TVM 0.6 环境中验证过代码，因此本文把它作为概念和历史记录，而不是当前版本的可直接复制脚本。

## 25. 原始代码生成示例

旧文章把代码生成接口概括为两个 build、两个 module 和两个 function。算子路径从 Tensor Expression 构造 GEMM，调用 `tvm.lower` 查看低层 IR，再调用 `tvm.build` 生成模块；Relay 路径则先取得 ResNet18 workload，调用 Relay build，再检查 Relay IR 和生成代码。

```python
M = K = N = 1024
k = te.reduce_axis((0, K), 'k')
A = te.placeholder((M, K), name='A')
B = te.placeholder((K, N), name='B')
C = te.compute((M, N), lambda x, y: te.sum(A[x, k] * B[k, y], axis=k), name='C')
s = te.create_schedule(C.op)
ir_m = tvm.lower(s, [A, B, C], simple_mode=True, name='mmult')
rt_m = tvm.build(ir_m, [A, B, C], target='c', name='mmult')
print(ir_m.astext(show_meta_data=False))
print(rt_m.get_source())
```

旧文章还记录了 C、LLVM、CUDA、OpenCL、OpenGL、Metal 和 Vulkan 等 target build 路径，并指出可以通过 Bring Your Own Codegen 扩展自定义后端。合并后应保留这一点，但强调目标后端的实现和调试方式不同：CPU 代码通常依赖 LLVM，GPU 需要处理设备函数和内存空间，移动端还需要结合平台运行时。

## 26. 原始量化和学习路径

量化笔记的动机来自两个方向：一是非结构化量化和稀疏在没有推理引擎支持时难以转化成真实加速，二是同一个团队需要面向 x86、GPU、ARM 等多个平台。文章还记录了模型从 PyTorch、ONNX、TFLite 到 Android 的迁移经历，以及移动端单帧处理达到 4 秒的反例。这些材料应保留，因为它们解释了为什么需要同时考虑模型表示、编译器和硬件，而不是只在训练侧压缩模型。

原始资料列表包括 TVM 安装、向量相加、TensorFlow 模型编译、INT8 quantization proposal 和 quantization story。重构后把它们分为两类：当前官方文档作为正文引用，旧社区 RFC 作为历史背景。这样既保留当时的学习路径，也避免把已经过时的接口当成当前使用建议。

## 27. 一份可直接执行的实验脚本结构

为了把历史实验变成可复现流程，可以按以下结构组织脚本：

```text
benchmarks/tvm/
  configs/
    cpu-fp32.json
    cpu-int8.json
  models/
  run_baseline.py
  run_relay.py
  run_gemm.py
  run_quantization.py
  validate_outputs.py
  summarize_results.py
```

`run_baseline.py` 只负责原始框架推理，`run_relay.py` 负责导入和图级 Pass，`run_gemm.py` 只测算子，`run_quantization.py` 负责校准和精度比较。把职责拆开后，任何一次失败都能定位到模型导入、编译、运行时、调度或量化，而不是在一个长脚本里猜测问题来源。

输出结果建议使用 JSON 或 CSV，至少包含 `model`、`shape`、`dtype`、`target`、`threads`、`warmup`、`repeat`、`p50_ms`、`p95_ms`、`error_max` 和 `accuracy`。后续表格由脚本生成，避免手工复制数字时发生错位。

## 28. 合并后的结论

六篇旧文章并不是六个互不相关的主题：初识 TVM 解释了问题背景，Relay 文章解释了图级表示和 Pass，代码生成文章解释了 IR 到模块的路径，GEMM 文章解释了算子级性能，量化文章解释了低精度部署，学习资料文章记录了当时的资料入口。重构的关键不是删除这些内容，而是为它们建立因果顺序。

因此，本文最终采用“基线—图—算子—代码—量化—端到端—复盘”的叙事。读者先看到为什么要优化，再看到 TVM 如何表示模型，然后通过 GEMM 理解调度，最后用量化和完整测试检验优化是否真的有业务价值。历史数字全部保留并标注来源，新实验可以沿相同表格重新运行，得到当前硬件上的可比结果。

## 29. 原始 Relay 代码：从标量表达式到卷积单元

旧文章中的 Hello Relay 不是抽象描述，而是一段可以看到 IR 输出的最小代码。它使用变量、常量和加法构造函数，再调用 build 观察图、参数和目标代码：

```python
from tvm import relay
import tvm.relay.op

x = relay.expr.var('x', relay.scalar_type('int64'), dtype='int64')
one = relay.expr.const(1, dtype='int64')
add = relay.op.tensor.add(x, one)
func = relay.expr.Function([x], add, relay.scalar_type('int64'))
mod = relay.Module.from_expr(func)
print(mod.astext(show_meta_data=False))
graph, lib, params = tvm.relay.build(mod, 'llvm', params={})
print(graph)
print(params)
print(lib.get_source())
```

旧版本的 `relay.Module.from_expr`、`relay.expr.Function` 和 `relay.scalar_type` 在当前 TVM 中可能已经迁移，但这段代码表达的实验方法仍然有效：先构造最小 IR，再检查编译产物，而不是直接把复杂模型交给编译器。

卷积单元实验还定义了 `batch_norm_infer`、`conv2d` 和 `conv_block` 三个辅助函数，将卷积、BatchNorm 和 ReLU 组合起来。输入形状为 `(1, 3, 224, 224)`，卷积核形状为 `(32, 3, 3, 3)`，步幅为 `(2, 2)`，输出形状为 `(1, 32, 112, 112)`。权重由固定随机种子生成，再通过 `relay.build` 编译并用 graph runtime 执行。这个例子说明图优化的输入不是一组孤立算子，而是带参数、布局和形状信息的完整函数。

```python
def conv_block(data, name, channels, kernel_size=(3, 3),
               strides=(1, 1), padding=(1, 1), epsilon=1e-5):
    weight = relay.var(name + '_weight')
    conv = relay.nn.conv2d(
        data, weight, channels=channels, kernel_size=kernel_size,
        strides=strides, padding=padding, data_layout='NCHW')
    bn = relay.nn.batch_norm(
        conv, relay.var(name + '_gamma'), relay.var(name + '_beta'),
        relay.var(name + '_moving_mean'), relay.var(name + '_moving_var'),
        epsilon=epsilon)[0]
    return relay.nn.relu(bn)

data = relay.var('data', shape=(1, 3, 224, 224), dtype='float32')
act = conv_block(data, 'graph', 32, strides=(2, 2))
mod = tvm.IRModule.from_expr(relay.Function(relay.analysis.free_vars(act), act))
mod = relay.transform.InferType()(mod)
```

原始文章还把优化前后的 IR 打印出来。优化前，BatchNorm 仍以独立调用存在；绑定参数、常量折叠和规范化后，它会变成卷积之后的乘法、加法和 ReLU。这个变化是理解图优化最重要的一手证据：不要只说“做了融合”，而要展示 IR 中哪些节点消失、哪些常量被内联、哪些布局发生改变。

## 30. Relay Pass 的原始对照

旧版 `my_optimize` 函数使用 `Sequential` 串联 `SimplifyInference`、`FoldConstant`、`FoldScaleAxis`、`CanonicalizeOps` 和再次 `FoldConstant`。这段代码体现了一个实用原则：Pass 有顺序关系，前一个 Pass 产生的结构可能成为后一个 Pass 的输入。当前版本的 Pass 名称或参数可能变化，因此重构时不直接保证旧代码可运行，而是保留其优化意图和 IR 对照。

```python
def my_optimize(mod, params=None):
    optimize = tvm.transform.Sequential([
        relay.transform.InferType(),
        relay.transform.SimplifyInference(),
        relay.transform.FoldConstant(),
        relay.transform.FoldScaleAxis(),
        relay.transform.CanonicalizeOps(),
        relay.transform.FoldConstant(),
    ])
    return optimize(mod)

optimized_mod = my_optimize(mod)
print(optimized_mod.astext(show_meta_data=False))
```

原始 Pass 清单中的 `Legalize` 负责把高层表达变成目标相关的等价形式，`SimplifyInference` 会简化推理阶段的数据流，例如处理 BatchNorm 和 Dropout，`EliminateCommonSubexpr` 消除公共子表达式，`CombineParallelConv2D` 和 `CombineParallelDense` 合并具有相同输入的并行计算，`FoldConstant` 做常量传播，`CanonicalizeCast` 和 `CanonicalizeOps` 规范化表达式，`AlterOpLayout` 改变布局，`FuseOps` 按规则形成更大的算子。重构文章会把这些 Pass 分成“语义简化、常量和表达式、布局、融合”四组，便于读者理解。

## 31. 原始量化测试结果与解释

量化笔记给出的测试表必须保留，因为它体现了一个经常被忽略的事实：INT8 并不自动带来端到端加速。原始结果如下：

| 模型 | 原始框架 | 原始框架时间 | TVM FP32 | TVM INT8 | TVM INT8 + AutoTVM |
|---|---|---:|---:|---:|---:|
| ResNet18 v1 | MXNet 1.5.1 | 27.8 ms | 46.9 ms | 51.10 ms | 25.83 ms |
| Inception v1 | TensorFlow 1.13 | 560 ms | 164 ms | 185 ms | 116 ms |

这里的结论不是“INT8 没用”，而是“量化路径需要与算子调优配套”。ResNet18 的 TVM FP32 已经慢于原始框架，直接 INT8 反而更慢；加入 AutoTVM 后才低于原始框架。Inception v1 的 TVM FP32 已经明显快于原始框架，INT8 单独造成回退，调优后才进一步改善。可能原因包括量化算子覆盖率、数据类型转换、目标硬件整数指令、参数布局和默认调度。

这些历史数据还存在版本和环境差异：MXNet 1.5.1、TensorFlow 1.13、早期 TVM 和当时的硬件不能与当前环境直接比较。重测时应增加模型精度、校准集、线程数、算子覆盖率和 P50/P95，并将“原始数字”与“当前复测数字”分成两列，避免覆盖历史事实。

量化代码建议保留成实验接口，而不是把版本相关的旧 API 混入正文：

```python
def compare_quantized(mod, params, target, calibration_data):
    # 当前 TVM 版本的 quantize API 可能变化，先按官方文档确认接口。
    fp32 = build_and_benchmark(mod, params, target, dtype='float32')
    int8_mod, int8_params = quantize_with_calibration(
        mod, params, calibration_data)
    int8 = build_and_benchmark(int8_mod, int8_params, target, dtype='int8')
    error = compare_outputs(fp32.outputs, int8.outputs)
    return {'fp32': fp32, 'int8': int8, 'error': error}
```

这里的 `quantize_with_calibration` 和 `build_and_benchmark` 是实验层封装，实际实现必须绑定具体 TVM 版本；它们不能伪装成跨版本稳定 API。正文应同时给出校准数据来源、量化范围、输出误差和失败算子列表。

## 32. 原始命令、精确数字与保留说明

为了避免重构后丢失一手证据，本文明确保留以下原始命令和精确测量值。旧文章中的构建命令为：

```bash
sudo apt-get install llvm
git clone --recursive https://github.com/dmlc/tvm.git
cd tvm && mkdir build
cp cmake/config.cmake build
# 将 USE_LLVM OFF 改为 set(USE_LLVM /usr/bin/llvm-config)
cd build
cmake ..
cmake -j4
pip install antlr4-python3-runtime
```

原始 TensorFlow/TVM 对比值分别是 `0.5861950877520752` 和 `0.2770531177520752` 秒附近，旧文章记录的完整输出还包括 African elephant、tusker、Indian elephant、banana 和 vault/desk 的分类分数。这里保留原始观察，但由于旧文中的数字存在不同复制版本，当前文章应以归档原文和测试日志为最终证据，不擅自把四舍五入后的数字重新推导成新结论。

GEMM 阶段数据中的 `baseline=2.49s`、`blocking=1.73s`、`vectorization=0.411s`、`loop permutation=0.104s`、`packing=0.0987s`、`write cache=0.0926s`、`parallel=0.018s`、`auto-tvm=0.014s`，以及 NumPy 的 `0.0135s`、`0.012s`、`0.0117s`、`0.0116s`、`0.0103s`、`0.01158s`、`0.012s`、`0.0112s`，均直接来自旧文章的实验表。它们统一标记为历史数据，不与新机器结果混合。

量化表中的 `27.8ms`、`46.9ms`、`51.10ms`、`25.83ms`、`560ms`、`164ms`、`185ms` 和 `116ms` 也原样保留。正文的新增分析只解释这些数据可能反映的调度、覆盖率和转换开销，不替代原始测量。这样，正文说明可以压缩和重排，代码、公式、命令和实验数据仍然能回溯到归档文章。

## 33. 复现实验方法
一个可复现的实验不应该从调度参数开始，而应该从目录和命令开始。可以为每种方案保存独立配置，配置中明确模型文件、输入形状、dtype、目标平台、线程数和测量次数。编译阶段和运行阶段分开记录，避免把编译时间误算进稳定推理延迟，也避免把第一次运行的缓存建立误算进每次请求。

第一步是验证模型输入输出。使用固定随机种子生成输入，保存一份小尺寸样本作为回归数据。第二步是运行原始框架，确认模型能稳定输出，并记录准确率或任务相关指标。第三步是导入 TVM，打印 Relay 模块，确认输入类型、参数数量和布局没有发生意外改变。第四步才是执行图优化和算子调度。每完成一个阶段，都要运行数值校验，防止错误在多个优化叠加后才暴露。

性能测量建议分成三层。第一层是单个 GEMM 的微基准，用来比较调度参数；第二层是模型热点算子，用来确认 GEMM 优化是否真的覆盖了模型主要耗时；第三层是完整模型，用来判断端到端价值。三层的输入尺寸和统计方式可以不同，但必须在表格中明确标注。不能用一个小矩阵的微基准结果推导大模型的服务延迟。

## 34. 常见失败案例

第一个失败案例是把编译时间当成运行时间。TVM 需要生成代码和构建运行时模块，首次编译可能很慢，但在线服务通常只在发布阶段编译一次。报告中应分别列出编译耗时和热运行耗时；如果场景需要动态编译，则应把编译成本纳入冷启动指标。

第二个失败案例是只测试一个矩阵尺寸。某组 tile 参数可能对 1024×1024 最好，却对 128×128 或非方阵很差。真实模型通常包含许多不同形状，应该用模型中的形状分布作为候选集合，并分别报告小矩阵、大矩阵和尾部尺寸的结果。

第三个失败案例是只比较平均值。CPU 频率变化、后台任务和内存分配都会造成长尾。至少应报告中位数、最小值、最大值和 P95；如果测量次数很少，应明确说明结果只是趋势观察。对需要发布的结论，最好在空闲机器上重复多轮，并记录运行环境。

第四个失败案例是忽略内存布局。模型导入后可能使用 NCHW，某个算子实现却更适合 NHWC；布局转换本身会消耗时间和内存带宽。优化前后必须检查布局变化，不能把布局转换的成本隐藏在“编译器优化”中。

第五个失败案例是量化后只看延迟。量化模型必须同时做输出误差、任务精度和算子覆盖检查。若某层出现大量饱和，最终准确率下降可能并不是 TVM 编译错误，而是校准集不能代表实际输入分布。对异常层进行逐层比较，比直接重新搜索所有调度参数更有效。

## 35. CPU、GPU 和移动端的差异

CPU 优化通常围绕缓存、SIMD、线程和成熟 BLAS 库展开。CPU 的优势是控制流灵活、内存层次清晰，适合形状不规则和小批量推理；缺点是并行宽度有限，线程启动和同步成本可能很快显现。对于 CPU，单线程调度和多线程调度应分别调优，不能简单把核心数写进配置。

GPU 优化则更加依赖线程块、共享内存、全局内存合并访问、warp 利用率和同步。一个在 CPU 上合理的循环分块方案，不能直接迁移到 CUDA。GPU 还需要考虑 kernel launch 开销和主机设备同步；如果模型包含许多很小的算子，单个算子的理论吞吐很高，端到端仍可能不理想。CUDA 官方编程指南和最佳实践指南对线程层次、内存和同步提供了完整说明[25][26]。

移动端的约束更复杂。功耗、热降频、内存容量和异构核心都会影响结果。短时间 benchmark 可能得到漂亮数字，但持续运行后由于温度升高而降频。移动端还需要考虑模型加载、内存拷贝和电池消耗，不能只看一次推理延迟。TVM 的跨平台能力可以减少重复开发，但目标平台仍需要独立测试和调度。

## 36. 如何写出可信的性能结论

性能结论应该包含对象、条件和范围。例如，“在某 CPU、某 TVM 版本、固定线程数和 1024×1024 FP32 GEMM 下，人工分块调度的中位延迟低于默认调度”是一个有边界的结论；“TVM 比 TensorFlow 快两倍”则缺少模型、版本、线程和统计口径，容易造成误解。

建议使用三类表述。观察类表述只陈述数据，例如“INT8 路径的 P50 延迟下降，输出准确率下降 0.2 个百分点”。解释类表述说明可能原因，例如“延迟下降与整数算子覆盖增加和权重带宽减少一致”。建议类表述给出适用条件，例如“对于固定硬件和固定形状，可以进一步搜索 tile 和并行参数”。三类表述分开后，读者更容易判断哪些是测量事实，哪些是工程推断。

如果结果与预期相反，也应该保留。GEMM 调度没有提升，可能说明系统瓶颈在内存拷贝或其他算子；INT8 没有端到端收益，可能说明量化覆盖率太低；Relay 融合后变慢，可能说明函数过度融合或寄存器压力升高。失败结果能帮助读者建立边界，通常比单一成功案例更有实践价值。

## 37. 后续优化路线

完成基础 GEMM 调优后，可以沿四条路线继续深入。第一条是自动调优，使用更多真实形状和硬件反馈，让调度参数从固定规则发展为模型化搜索。第二条是算子融合，把 GEMM 前后的偏置、激活和布局转换纳入统一调度，减少中间结果写回。第三条是低精度计算，比较 FP16、BF16、INT8 和混合精度在精度与吞吐之间的取舍。第四条是端到端部署，加入模型加载、服务线程池、批处理和监控指标。

如果模型包含动态形状，应该先分析形状分布，再决定是为常见形状分别编译，还是使用泛化更好的调度。若模型经常更新，编译缓存和调优日志需要纳入发布系统；若硬件平台固定，可以接受更长的离线搜索时间，以换取稳定运行时性能。若硬件平台变化频繁，过度依赖某组参数可能增加维护成本。

最终目标不是让每个算子都达到理论峰值，而是用合理的工程成本，把最主要的端到端瓶颈消除。TVM 提供了表达和搜索这些优化的工具，性能结果仍需要通过实验、代码检查和业务指标共同确认。

## 参考资料

1. [Apache TVM Documentation](https://tvm.apache.org/docs/)
2. [Relay: A High-Level Intermediate Representation for Deep Learning](https://tvm.apache.org/docs/arch/relay_intro.html)
3. [Relay Python API](https://tvm.apache.org/docs/reference/api/python/relay.html)
4. [TVM IRModule API](https://tvm.apache.org/docs/reference/api/python/ir.html)
5. [TVM TensorIR Documentation](https://tvm.apache.org/docs/deep_dive/tensor_ir/index.html)
6. [TVM Tensor Expression Language](https://tvm.apache.org/docs/arch/ir.html)
7. [TVM Runtime Module API](https://tvm.apache.org/docs/reference/api/python/runtime.html)
8. [TVM: An Automated End-to-End Optimizing Compiler for Deep Learning](https://arxiv.org/abs/1802.04799)
9. [Ansor: Generating High-Performance Tensor Programs for Deep Learning](https://arxiv.org/abs/2006.06762)
10. [TensorIR: An Abstraction for Automatic Tensorized Program Optimization](https://arxiv.org/abs/2207.04296)
11. [Learning to Optimize Tensor Programs](https://arxiv.org/abs/2305.17380)
12. [AutoTVM Documentation](https://tvm.apache.org/docs/v0.8.0/tutorial/auto_scheduler_matmul_x86.html)
13. [MetaSchedule Documentation](https://tvm.apache.org/docs/deep_dive/meta_schedule/index.html)
14. [TVM Quantization Documentation](https://tvm.apache.org/docs/v0.9.0/how_to/deploy_models/deploy_quantized.html)
15. [TVM Build API](https://tvm.apache.org/docs/reference/api/python/driver.html)
16. [LLVM Language Reference](https://llvm.org/docs/LangRef.html)
17. [MLIR Linalg Dialect](https://mlir.llvm.org/docs/Dialects/Linalg/)
18. [XLA Architecture](https://openxla.org/xla)
19. [Halide Documentation](https://halide-lang.org/docs/)
20. [Tensor Comprehensions](https://github.com/facebookresearch/TensorComprehensions)
21. [BLAS Technical Forum Standard](https://www.netlib.org/blas/)
22. [OpenBLAS Documentation](https://www.openmathlib.org/OpenBLAS/)
23. [Intel oneMKL GEMM](https://www.intel.com/content/www/us/en/docs/onemkl/developer-reference-c/2024-0/gemm.html)
24. [ARM Compute Library](https://arm-software.github.io/ComputeLibrary/latest/)
25. [CUDA C++ Programming Guide](https://docs.nvidia.com/cuda/cuda-c-programming-guide/)
26. [CUDA C++ Best Practices Guide](https://docs.nvidia.com/cuda/cuda-c-best-practices-guide/)
27. [OpenCL Specification](https://registry.khronos.org/OpenCL/specs/3.0-unified/html/)
28. [TensorFlow Lite 8-bit Quantization Specification](https://www.tensorflow.org/lite/performance/quantization_spec)
29. [TensorFlow Lite Post-training Quantization](https://www.tensorflow.org/lite/performance/post_training_quantization)
30. [ONNX Runtime Quantization](https://onnxruntime.ai/docs/performance/model-optimizations/quantization.html)
31. [Quantizing deep convolutional networks for efficient inference](https://arxiv.org/abs/1806.08342)
32. [Integer Quantization for Deep Learning Inference](https://arxiv.org/abs/2004.09602)
33. [Efficient Processing of Deep Neural Networks](https://arxiv.org/abs/1608.06993)
34. [Efficient GEMM-based convolution algorithms](https://arxiv.org/abs/1509.09308)
35. [Anatomy of High-Performance Matrix Multiplication](https://www.cs.utexas.edu/~flame/pubs/GotoTOMS_rev.pdf)
