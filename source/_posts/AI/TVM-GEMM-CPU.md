---
title: TVM学习笔记--GEMM优化及测试数据
date: 2020-08-13
categories:
- AI
mathjax: true
---

　　在[《初识TVM，相比于tensorflow的2倍性能提升》](https://zhuanlan.zhihu.com/p/88369758)之后，最近花了一点业余时间了解TVM及其周边，并进行相应的性能测试。整体感受是计算优化(GEMM)是非常繁杂的工程工作，需要花费大量的时间和精力才能有比较好的效果。numpy非常优秀，大矩阵乘法硬件利用率在90%以上。TVM在GEMM优化上能实现和numpy相当的效果，重要的是它能大大简化工作量。参考了一些文章，这里简单罗列了几个知识点和测试数据。
1. 怎么评估硬件的理论性能？浮点峰值？
2. 简单测试一下numpy的性能数据，硬件利用率
3. 怎么做GEMM优化？
4. TVM怎么做GEMM的优化？及其与numpy性能的比较

### 怎么评估硬件的计算性能
　　对于性能优化来说，了解硬件的性能指标是非常有必要的。在Linux系统上可以通过/proc/cpuinfo文件看看机器的配置。比如CPU主频、CPU核数core、cache大小、是否支持向量指令SSE、AVX2、AVX512等，这些对于计算性能有非常大的影响。[浮点峰值那些事儿](https://zhuanlan.zhihu.com/p/28226956)。通常我们使用浮点计算能力来衡量硬件的性能，对于多核服务器来说，频率为2.4G，支持AVX2，FMA向量指令，单核性能如下：
	对于float32理论峰值为2.4G \* （8+8） \* 2  = 76.8 GFLOPS
	对于float64理论峰值为2.4G \* （4+4） \* 2  = 38.4 GFLOPS

### 测试numpy GEMM硬件利用率
　　numpy非常优秀，我们通过矩阵乘法了解其性能数据。测试机器为一台多核的服务器，主频是2.4G，支持FMA和AVX2向量指令。测试了不同size矩阵相乘的性能数据。分别测试了单核和四核状态下对float32和float64的不同size(32,128,1024,2048等）矩阵相乘的性能数据。测试结果显示numpy在大矩阵相乘中，硬件利用率大概在90%左右。

name | 32 | 128|1024|2048|4096|10240|硬件利用率|  
-|-|-|
单核float32|1.82|36.16|67.99|67.94|68.88|69.88|91.0%
单核float64|1.67|19.49|35.56|35.40|36.11|36.90|96.1%
四核float32|6.6|52.2|225.42|246.2|244.2|256.0|83.8%
四核float64|5.56|37.62|116.42|120.39|127.03|141.15|91.9%
[测试代码](https://github.com/wxquare/programming/blob/master/blog/TVM_CPU_schedule/test_numpy_gemm_performance.py)

### 怎么优化GEMM？
　　通用矩阵乘(GEMM)是计算领域非常基础且核心的工作，目前已有大量的工作，这里就不赘述了。大体上通过**分块来减少访存次数、存储顺序、提高cache命中率、利用寄存器提高效率、利用SSE等向量指令提高计算效率**等方法。https://github.com/flame/how-to-optimize-gemm/wiki 一步一步详细介绍了GEMM优化的过程，这里在此基础上增加FMA指令的使用，测试了其在1024*1204矩阵相乘的硬件利用率：

name | 64 | 256 |512|1024|硬件利用率|主要优化点|  
-|-|-|
MMult0|1.51|0.79|0.66|0.65|1.69%|base
MMult_1x4_5|2.15|1.08|0.72|0.716|2.6%|一次计算1x4个数
MMult_1x4_9|4.90|3.15|3.10|3.14|8.18%|1x4，寄存器
MMult_4x4_5|2.76|1.53|1.26|1.26|3.28%|一次计算4x4个数
MMult_4x4_9|5.19|2.92|2.88|2.87|7.47%|4x4，寄存器
MMult_4x4_10|5.95|4.16|4.04|4.01|10.4%|4x4，寄存器，SSE
MMult_4x4_10_1|10.0|6.6|6.35|6.4|16.7%|4x4，寄存器，FMA
MMult_4x4_11_1|14.5|8.95|7.16|7.08|18.4%|4x4，寄存器，FMA，分块(缓存)
MMult_4x4_15_1|11.3|11.6|11.7|11.7|30.4%|4x4，寄存器，FMA，分块，内存顺序

[测试代码](https://github.com/wxquare/programming/tree/master/blog/TVM_CPU_schedule/HowToOptimizeGemm)

### TVM GEMM优化与numpy性能比较
　　TVM官网上有关于其针对GEMM的优化的schedule，这里也不赘述了，感兴趣的可以参考后面的参考文章进一步学习，这里测试了在1024*1024矩阵乘法的效率以及其和numpy的比较，可以看出TVM在简单编码的基础上能达到和numpy相当的性能。

  | TVM运行时间 | numpy运行时间 |
-|-|-|  
baseline|2.49s|0.0135s
blocking|1.73s|0.012s
vectorization|0.411s|0.0117s
loop permutaion|0.104s|0.0116s
packing|0.0987s|0.0103s
write_cache|0.0926s|0.01158s
parallel|0.018s|0.012s
auto-tvm|0.014s|0.0112s
[每个阶段测试代码](https://github.com/wxquare/programming/tree/master/blog/TVM_CPU_schedule/TVM_GEMM)

参考学习链接：
1、浮点峰值那些事儿https://zhuanlan.zhihu.com/p/28226956
2、通用矩阵乘（GEMM）优化算法，https://jackwish.net/gemm-optimization.html
3、如何利用TVM快速实现超越Numpy(MKL)的GEMM。https://zhuanlan.zhihu.com/p/75203171
4、tutorial：https://docs.tvm.ai/tutorials/optimize/opt_gemm.html
5、d2ltvm:http://tvm.d2l.ai/chapter_cpu_schedules/index.html
6、https://github.com/flame/how-to-optimize-gemm