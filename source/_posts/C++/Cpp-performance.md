---
title: C++ 程序性能分析
categories: 
- C/C++
---


　　C++代码编译测试完成功能之后，有时会遇到一些性能问题，此时需要学会使用一些工具对其进行性能分析，找出程序的性能瓶颈，然后进行优化，基本需要掌握下面几个命令：
1. time分析程序的执行时间
2. top观察程序资源使用情况
3. perf/gprof进一步分析程序的性能
4. 内存问题与valgrind
5. 自己写一个计时器，计算局部函数的时间



## 一、time
1. shell time。 time非常方便获取程序运行的时间，包括用户态时间user、内核态时间sys和实际运行的时间real。我们可以通过(user+sys)/real计算程序CPU占用率，判断程序时CPU密集型还是IO密集型程序。
$time ./kcf2.0 ../data/bag.mp4 312 146 106 98 1 196 result.csv 1
real	0m2.065s
user	0m4.598s
sys	    0m0.907s
cpu使用率：(4.598+0.907)/2.065=267%
视频帧数196，196/2.065=95


2. Linux中除了shell time，还有/usr/bin/time，它能获取程序运行更多的信息，通常带有-v参数。
```
$ /usr/bin/time -v  ./kcf2.0 ../data/bag.mp4 312 146 106 98 1 196 result.csv 1
    User time (seconds): 4.28                                  # 用户态时间
	System time (seconds): 1.11                                # 内核态时间
	Percent of CPU this job got: 279%                          # CPU占用率
	Elapsed (wall clock) time (h:mm:ss or m:ss): 0:01.93   
	Average shared text size (kbytes): 0
	Average unshared data size (kbytes): 0
	Average stack size (kbytes): 0
	Average total size (kbytes): 0
	Maximum resident set size (kbytes): 63980                  # 最大内存分配
	Average resident set size (kbytes): 0
	Major (requiring I/O) page faults: 0
	Minor (reclaiming a frame) page faults: 19715              # 缺页异常
	Voluntary context switches: 3613                           # 上下文切换
	Involuntary context switches: 295682
	Swaps: 0
	File system inputs: 0
	File system outputs: 32
	Socket messages sent: 0
	Socket messages received: 0
	Signals delivered: 0
	Page size (bytes): 4096
	Exit status: 0
```


## 二、top
top是linux系统的任务管理器，它既能看系统所有任务信息，也能帮助查看单个进程资源使用情况。
主要有以下几个功能：
1. 查看系统任务信息：
 Tasks:  87 total,   1 running,  86 sleeping,   0 stopped,   0 zombie
2. 查看CPU使用情况
 Cpu(s):  0.0%us,  0.2%sy,  0.0%ni, 99.7%id,  0.0%wa,  0.0%hi,  0.0%si,  0.2%st
3. 查看内存使用情况
 Mem:    377672k total,   322332k used,    55340k free,    32592k buffers
4. 查看单个进程资源使用情况 
	- PID：进程的ID
	- USER：进程所有者
	- PR：进程的优先级别，越小越优先被执行
	- NInice：值
	- VIRT：进程占用的虚拟内存
	- RES：进程占用的物理内存
	- SHR：进程使用的共享内存
	- S：进程的状态。S表示休眠，R表示正在运行，Z表示僵死状态，N表示该进程优先值为负数
	- %CPU：进程占用CPU的使用率
	- %MEM：进程使用的物理内存和总内存的百分比
	- TIME+：该进程启动后占用的总的CPU时间，即占用CPU使用时间的累加值。
	- COMMAND：进程启动命令名称
5. 除此之外top还提供了一些交互命令：
	- q:退出
	- 1:查看每个逻辑核
	- H：查看线程
	- P：按照CPU使用率排序
	- M：按照内存占用排序

参考：https://linuxtools-rst.readthedocs.io/zh_CN/latest/tool/top.html


## 三、perf
参考：https://www.ibm.com/developerworks/cn/linux/l-cn-perf1/index.html
参考：https://zhuanlan.zhihu.com/p/22194920

### 1. perf stat
　　做任何事都最好有条有理。老手往往能够做到不慌不忙，循序渐进，而新手则往往东一下，西一下，不知所措。面对一个问题程序，最好采用自顶向下的策略。先整体看看该程序运行时各种统计事件的大概，再针对某些方向深入细节。而不要一下子扎进琐碎细节，会一叶障目的。有些程序慢是因为计算量太大，其多数时间都应该在使用 CPU 进行计算，这叫做 CPU bound 型；有些程序慢是因为过多的 IO，这种时候其 CPU 利用率应该不高，这叫做 IO bound 型；对于 CPU bound 程序的调优和 IO bound 的调优是不同的。如果您认同这些说法的话，Perf stat 应该是您最先使用的一个工具。它通过概括精简的方式提供被调试程序运行的整体情况和汇总数据。虚拟机上面有些参数不全面，cycles、instructions、branches、branch-misses。下面的测试数据来自服务器。                                          
**$time ./kcf2.0 ../data/bag.mp4 312 146 106 98 1 196 result.csv 1**
```
     25053.120420      task-clock (msec)         #   17.196 CPUs utilized          
         1,509,877      context-switches          #    0.060 M/sec                  
             3,427      cpu-migrations            #    0.137 K/sec                  
            34,025      page-faults               #    0.001 M/sec                  
    65,242,918,152      cycles                    #    2.604 GHz                    
                 0      stalled-cycles-frontend   #    0.00% frontend cycles idle   
                 0      stalled-cycles-backend    #    0.00% backend  cycles idle   
    64,695,693,541      instructions              #    0.99  insns per cycle        
     8,049,836,066      branches                  #  321.311 M/sec                  
        42,734,371      branch-misses             #    0.53% of all branches        

       1.456907056 seconds time elapsed
```
### 2. perf top
　　Perf top 用于实时显示当前系统的性能统计信息。该命令主要用来观察整个系统当前的状态，比如可以通过查看该命令的输出来查看当前系统最耗时的内核函数或某个用户进程。
### 3. perf record/perf report
　　使用 top 和 stat 之后，这时对程序基本性能有了一个大致的了解，为了优化程序，便需要一些粒度更细的信息。比如说您已经断定目标程序计算量较大，也许是因为有些代码写的不够精简。那么面对长长的代码文件，究竟哪几行代码需要进一步修改呢？这便需要使用 perf record 记录单个函数级别的统计信息，并使用 perf report 来显示统计结果。您的调优应该将注意力集中到百分比高的热点代码片段上，假如一段代码只占用整个程序运行时间的 0.1%，即使您将其优化到仅剩一条机器指令，恐怕也只能将整体的程序性能提高 0.1%。俗话说，好钢用在刀刃上，要优化热点函数。

```
perf record – e cpu-clock ./t1 
perf report
```
增加-g参数可以获取调用关系
```
perf record – e cpu-clock – g ./t1 
perf report
```
$perf record -e cpu-clock -g ./kcf2.0 ../data/bag.mp4 312 146 106 98 1 196 result.csv 1
$perf report
![](https://github.com/wxquare/wxquare.github.io/raw/hexo/source/photos/perf_kcf2.0.jpg)
	

## 四、gprof
参考： https://blog.csdn.net/stanjiang2010/article/details/5655143




## 五、内存问题与valgrind
### 5.1常见的内存问题
1. 使用未初始化的变量
对于位于程序中不同段的变量，其初始值是不同的，全局变量和静态变量初始值为0，而局部变量和动态申请的变量，其初始值为随机值。如果程序使用了为随机值的变量，那么程序的行为就变得不可预期。
2. 内存访问越界
比如访问数组时越界；对动态内存访问时超出了申请的内存大小范围。
3. 内存覆盖
C 语言的强大和可怕之处在于其可以直接操作内存，C 标准库中提供了大量这样的函数，比如 strcpy, strncpy, memcpy, strcat 等，这些函数有一个共同的特点就是需要设置源地址 (src)，和目标地址(dst)，src 和 dst 指向的地址不能发生重叠，否则结果将不可预期。
4. 动态内存管理错误
常见的内存分配方式分三种：静态存储，栈上分配，堆上分配。全局变量属于静态存储，它们是在编译时就被分配了存储空间，函数内的局部变量属于栈上分配，而最灵活的内存使用方式当属堆上分配，也叫做内存动态分配了。常用的内存动态分配函数包括：malloc, alloc, realloc, new等，动态释放函数包括free, delete。一旦成功申请了动态内存，我们就需要自己对其进行内存管理，而这又是最容易犯错误的。下面的一段程序，就包括了内存动态管理中常见的错误。
a. 使用完后未释放
b. 释放后仍然读写
c. 释放了再释放
5. 内存泄露
内存泄露（Memory leak）指的是，在程序中动态申请的内存，在使用完后既没有释放，又无法被程序的其他部分访问。内存泄露是在开发大型程序中最令人头疼的问题，以至于有人说，内存泄露是无法避免的。其实不然，防止内存泄露要从良好的编程习惯做起，另外重要的一点就是要加强单元测试（Unit Test），而memcheck就是这样一款优秀的工具

### 5.1 valgrind内存检测

```
#include <iostream>
using namespace std;


int main(int argc, char const *argv[])
{
    int a[5];
    a[0] = a[1] = a[3] = a[4] = 0;

    int s=0;
    for(int i=0;i<5;i++){
        s+=a[i];
    }
    if(s == 0){
        std::cout << s << std::endl;
    }
    a[5] = 10;
    std::cout << a[5] << std::endl;


    int *invalid_write = new int[10];
    delete [] invalid_write;
    invalid_write[0] = 3;

    int *undelete = new int[10];
    
    return 0;
}
```
```
==102507== Memcheck, a memory error detector
==102507== Copyright (C) 2002-2017, and GNU GPL'd, by Julian Seward et al.
==102507== Using Valgrind-3.14.0 and LibVEX; rerun with -h for copyright info
==102507== Command: ./a.out
==102507== 
==102507== Conditional jump or move depends on uninitialised value(s)
==102507==    at 0x1091F6: main (learn_valgrind.cpp:14)
==102507== 
10
==102507== Invalid write of size 4
==102507==    at 0x109270: main (learn_valgrind.cpp:23)
==102507==  Address 0x4dc30c0 is 0 bytes inside a block of size 40 free'd
==102507==    at 0x483A55B: operator delete[](void*) (in /usr/lib/x86_64-linux-gnu/valgrind/vgpreload_memcheck-amd64-linux.so)
==102507==    by 0x10926B: main (learn_valgrind.cpp:22)
==102507==  Block was alloc'd at
==102507==    at 0x48394DF: operator new[](unsigned long) (in /usr/lib/x86_64-linux-gnu/valgrind/vgpreload_memcheck-amd64-linux.so)
==102507==    by 0x109254: main (learn_valgrind.cpp:21)
==102507== 
==102507== 
==102507== HEAP SUMMARY:
==102507==     in use at exit: 40 bytes in 1 blocks
==102507==   total heap usage: 4 allocs, 3 frees, 73,808 bytes allocated
==102507== 
==102507== LEAK SUMMARY:
==102507==    definitely lost: 40 bytes in 1 blocks
==102507==    indirectly lost: 0 bytes in 0 blocks
==102507==      possibly lost: 0 bytes in 0 blocks
==102507==    still reachable: 0 bytes in 0 blocks
==102507==         suppressed: 0 bytes in 0 blocks
==102507== Rerun with --leak-check=full to see details of leaked memory
==102507== 
==102507== For counts of detected and suppressed errors, rerun with: -v
==102507== Use --track-origins=yes to see where uninitialised values come from
==102507== ERROR SUMMARY: 2 errors from 2 contexts (suppressed: 0 from 0)

```
1. https://www.ibm.com/developerworks/cn/linux/l-cn-valgrind/index.html
2. http://senlinzhan.github.io/2017/12/31/valgrind/
3. https://www.ibm.com/developerworks/cn/aix/library/au-memorytechniques.html


## 六、自定义timer计时器
```
class timer {
public:
    clock_t start;
    clock_t end;
    string name;
    timer(string n) {
        start = clock();
        name = n;
    }
    ~timer() {
        end = clock();
        printf("%s time: %f \n", name.c_str(), 
            (end - start) * 1.0 / CLOCKS_PER_SEC * 1000);
    }
};
```

