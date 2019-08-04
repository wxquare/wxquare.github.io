---
title: C++ 程序性能分析
categories: 
- C/C++
---


C++ 代码编译测试完成功能之后，有时会遇到性能问题，此时需要学会使用一些工具对其进行性能分析，找出程序的性能瓶颈，然后进行优化：
1. time分析程序的执行时间
2. top观察程序资源使用情况
3. perf/gprof进一步分析程序的性能
4. valgrind分析程序的内存泄露问题


## 一、Linux C/C++
1、会使用Makefile和CMake构建项目
2、会使用gdb调试程序，程序出错、coredump、异常死循环死锁等。
3、会使用time,top,perf,gprof工具分析程序的性能、资源使用情况。
4、会使用Valgrind分析程序内存泄露问题。
5、




## 二、TODO
3、gdb调试相关的经验?基本调试、运行的进程、coredump文件分析？
- -g参数
- 不能strip
- 

4、C++程序性能分析、profile？
5、C++ 如何定位内存泄露？
6、动态链接和静态链接的区别
7、写一个c程序辨别系统是64位 or 32位
8、写一个c程序辨别系统是大端or小端字节序





参考：
- [跟我学些makefile](https://github.com/wxquare/programming/blob/master/document/%E8%B7%9F%E6%88%91%E4%B8%80%E8%B5%B7%E5%86%99Makefile-%E9%99%88%E7%9A%93.pdf)
- [CMake入门实战](https://www.hahack.com/codes/cmake/)


---
title: time/top/perf/gprof程序性能分析
categories:
- C/C++
---


　　会使用time、top、perf和gprof工具分析程序的性能。


## 一、time
1. shell time。 time非常方便获取程序运行的时间，包括用户态时间user、内核态时间sys和实际运行的时间real。我们可以通过(user+sys)/real计算程序CPU占用率，判断程序时CPU密集型还是IO密集型程序。
$time ffmpeg -y -i in.mp4 -vf "crop=w=100:h=100:x=12:y=34" -acodec copy out.mp4
real	0m0.144s
user	0m0.048s
sys	    0m0.124s

2. Linux中除了shell time，还有/usr/bin/time，它能获取程序运行更多的信息，通常带有-v参数。
```
$ /usr/bin/time -v ffmpeg -y -i in.mp4 -vf "crop=w=100:h=100:x=12:y=34" -acodec copy out.mp4

User time (seconds): 0.01                            # 用户态时间
System time (seconds): 0.10                          # 内核态时间
Percent of CPU this job got: 124%                    # CPU占用率
Elapsed (wall clock) time (h:mm:ss or m:ss): 0:00.09 # 实际时间
Average shared text size (kbytes): 0
Average unshared data size (kbytes): 0
Average stack size (kbytes): 0
Average total size (kbytes): 0
Maximum resident set size (kbytes): 20096
Average resident set size (kbytes): 0
Major (requiring I/O) page faults: 0                 # 缺页异常 
Minor (reclaiming a frame) page faults: 2953
Voluntary context switches: 2715                     # 上下文切换
Involuntary context switches: 414
Swaps: 0
File system inputs: 0
File system outputs: 368
Socket messages sent: 0
Socket messages received: 0
Signals delivered: 0
Page size (bytes): 4096
Exit status: 0
```

参考：https://my.oschina.net/yumm007/blog/920412

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



## 四、gprof
参考： https://blog.csdn.net/stanjiang2010/article/details/5655143





---
title: Linux程序常见内存问题和Valgrind
categories: 
- C/C++
---


## Linux程序的内存问题
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

## valgrind内存检测
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