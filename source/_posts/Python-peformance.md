---
title: Python 程序性能优化
categories:
- Python
---

Python程序为什么慢？
----
　　不同的场景下，代码是有不同的要求，大体有三个等级，“管用、更好、更快”。相比C/C++，Python具有较好的开发系效率，但是程序的性能运行速度会差一些。究其原因是Python为了灵活性，牺牲了效率。
1. **动态类型**。对于C/C++等静态类型语言，由于变量的类型固定，变量之间的运算很容易指定特定的函数。而动态类型在运行的时间需要大量if else判断处理，直到找到符合条件的函数。**动态类型增加语言的易用性，但是牺牲了程序的运行效率**。

2. GIL(Global Interpreter Lock)全局解释锁，CPython在解释执行任何Python代码的时候，首先都需要**they acquire GIL when running，release GIL when blocking for I/O**。如果没有涉及I/O操作，只有CPU密集型操作时，解释器每隔100一段时间（100ticks）就会释放GIL。GIL是实现Python解释器的（Cython)时所引入的一个概念，不是Python的特性。
由于GIL的存在，使得Python对于计算密集型任务，多线程threading模块形同虚设，因为线程在实际运行的时候必须获得GIL，而GIL只有一个，因此无法发挥多核的优势。为了绕过GIL的限制，只能使用multiprocessing等多进程模块，每个进程各有一个CPython解释器，也就各有一个GIL。

3. CPython不支持JIL（Just-In-Time Compiler),JIL 能充分利用程序运行时信息，进行类型推导等优化，对于重复执行的代码段来说加速效果明显。对于CPython如果想使用JIT对计算密集型任务进行优化，可以尝试使用JIT包numba，它能使得相应的函数变成JIT编译。



Python程序优化的思路？
----
　　最近在做一些算法优化方面的工作,简单总结一下思路:
1. 熟悉算法的整体流程，对于算法代码，最开始尽可能不要使用多线程和多进程方法，
2. 在1的基础上跑出算法的CPU profile，整体了解算法耗时分布和瓶颈。Python提供的cProfile模块灵活的针对特定函数或者文件产生profile文件，根据profile数据进行代码性能优化。
 - 可以直接将生成profile代码写在Python脚本
 - 使用命名行方式生成profile
 - 分析工具pstats
 - 图形化工具[Gprof2Dot](https://github.com/jrfonseca/gprof2dot)。python gprof2dot.py -f pstats result.out | dot -Tpng -o result.png
 [https://blog.csdn.net/asukasmallriver/article/details/74356771](https://blog.csdn.net/asukasmallriver/article/details/74356771)
3. 程序（算法）本身的剪枝。比如视频追踪中，考虑是否让每个像素点都参与计算？优化后选择梯度变化最大的1w个像素点参与计算，能提高分辨率大的视频追踪效率。
4. 使用矩阵操作代替循环操作。(get_values())
5. 任务分解，在理解算法的基础上寻找并行机会，利用多线程或者多进程充分利用机器资源。生产者消费者模型，专门的线程负责图像获取和图形变换，专门的线程负责特征提取和追踪。
6. 使用C/C++重写效率低的瓶颈部分
7. 使用GPU计算