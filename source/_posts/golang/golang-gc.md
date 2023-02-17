---
title: Go 内存管理与垃圾回收
categories:
- Golang
---

　　为了避开直接通过系统调用分配内存而导致的性能开销，通常会通过预分配、内存池等操作自主管理内存。golang由运行时runtime管理内存，完成初始化、分配、回收和释放操作。目前主流的内存管理器有glibc和tcmolloc，tcmolloc由Google开发，具有更好的性能，兼顾内存分配的速度和内存利用率。golang也是使用类似tcmolloc的方法进行内存管理。建议参考下面链接学习tcmalloc的原理，其内存管理的方法也是golang内存分配的方法。另外一个原因，golang自主管理也是为了更好的配合垃圾回收。

【1】.https://zhuanlan.zhihu.com/p/29216091  
【2】.http://goog-perftools.sourceforge.net/doc/tcmalloc.html 


## What is the Go runtime?
  The Go runtime is a collection of software components that provide essential services for Go programs, including memory management, garbage collection, scheduling, and low-level system interaction. The runtime is responsible for managing the execution of Go programs and for providing a consistent, predictable environment for Go code to run in.

At a high level, the Go runtime is responsible for several core tasks:
- Memory management: The runtime manages the allocation and deallocation of memory used by Go programs, including the stack, heap, and other data structures.
- Garbage collection: The runtime automatically identifies and frees memory that is no longer needed by a program, preventing memory leaks and other related issues.
- Scheduling: The runtime manages the scheduling of Goroutines, the lightweight threads used by Go programs, to ensure that they are executed efficiently and fairly.
- Low-level system interaction: The runtime provides an interface for Go programs to interact with low-level system resources, including system calls, I/O operations, and other low-level functionality.

The Go runtime is an essential component of the Go programming language, and it is responsible for many of the language's unique features and capabilities. By providing a consistent, efficient environment for Go code to run in, the runtime enables developers to write high-performance, scalable software that can run on a wide range of platforms and architectures.

![runtime](https://github.com/wxquare/wxquare.github.io/raw/hexo/source/images/runtime.png)

## 程序启动过程bootstrap

## 内存分配和管理策略mallocgc

## 垃圾回收garbage collector

## 程序并发Goroutine调度



## 一、内存管理基本策略
为了兼顾内存分配的速度和内存利用率，大多数都采用以下策略进行内存管理：
1. **申请**：每次从操作系统申请一大块内存（比如1MB），以减少系统调用
2. **切分**：为了兼顾大小不同的对象，将申请到的内存按照一定的策略切分成小块，使用链接相连
3. **分配**：为对象分配内存时，只需从大小合适的链表中提取一块即可。
4. **回收复用**: 对象不再使用时，将该小块内存归还到原链表
5. **释放**： 如果闲置内存过多，则尝试归凡部分内存给操作系统，减少内存开销。



## 二、golang内存管理
　golang内存管理基本继承了tcmolloc成熟的架构，因此也符合内存管理的基本策略。
1. 分三级管理，线程级的thread cache，中央center cache，和管理span的center heap。
2. 每一级都采用链表管理不同size空闲内存，提高内存利用率
3. 线程级的tread local cache能够减少竞争和加锁操作，提高效率。中央center cache为所有线程共享。
4. 小对象直接从本地cache获取，大对象从center heap获取，提高内存利用率
5. 每一级内存不足时，尝试从下一级内存获取
![内存三级管理](https://github.com/wxquare/wxquare.github.io/raw/hexo/source/photos/threelayer.jpg)
![线程cache](https://github.com/wxquare/wxquare.github.io/raw/hexo/source/photos/threadheap.gif)
![大对象span管理](https://github.com/wxquare/wxquare.github.io/raw/hexo/source/photos/pageheap.gif)



## 三、垃圾回收算法概述

　　golang是近几年出现的带有垃圾回收的现代语言，其垃圾回收算法自然也相互借鉴。因此在学习golang gc之前有必要了解目前主流的垃圾回收方法。
1. **引用计数**：熟悉C++智能指针应该了解引用计数方法。它对每一个分配的对象增加一个计数的域，当对象被创建时其值为1。每次有指针指向该对象时，其引用计数增加1，引用该对象的对象被析构时，其引用计数减1。当该对象的引用计数为0时，该对象也会被析构回收。引用对象对于C++这类没有垃圾回收器，对于便于对象管理的是不错的工具，但是维护引用计数会造成程序运行效率下降。
2. **标记-清扫**： 标记清扫是古老的垃圾回收算法，出现在70年代。通过指定每个内存阈值或者时间长度，垃圾回收器会挂起用户程序，也称为STW（stop the world）。垃圾回收器gc会对程序所涉及的所有对象进行一次遍历以确定哪些内存单元可以回收，因此分为标记（mark）和清扫（sweep），标记阶段标明哪些内存在使用不能回收，清扫阶段将不需要的内存单元释放回收。标记清扫法最大的问题是需要STW，当程序使用的内存较多时，其性能会比较差，延时较高。
3. **三色标记法**： 三色标记法是对标记清扫的改进，也是golang gc的主要算法，其最大的的优点是能够让部分gc和用户程序并发进行。它将对象分为白色、灰色和黑色：
	- 开始时所有的对象都是白色
	- 从根出发，将所有可到达对象标记为灰色，放入待处理队列
	- 从待处理队列中取出灰色对象，并将其引用的对象标记为灰色放入队列中，其自身标记为黑色。
	- 重复步骤3，直到灰色对象队列为空。最终只剩下白色对象和黑色对象，对白色对象尽心gc。
4. 另外，还有一些在此基础上进行优化改进的gc算法，例如分代收集，节点复制等，它会考虑到对象的生命周期的长度，减少扫描标记的操作，相对来说效率会高一些。


## 四、golang垃圾回收
　　**golang gc是使用三色标记清理法**，为了对用户对象进行标记需要将用户程序所有线程全部冻结（STW），当程序中包含很多对象时，暂停时间会很长，用户逻辑对用户的反应就会中止。那么如何缩短这个过程呢?一种自然的想法，在三色标记法扫描之后，只会存在黑色和白色两种对象，黑色是程序正在使用的对象不可回收，白色对象是此时不会被程序的对象，也是gc的要清理的对象。那么回收白色对象肯定不会和用户程序造成竞争冲突，因此回收操作和用户程序是可以并发的，这样可以缩短STW的时间。

　　**写屏障**使得扫描操作和回收操作都可以和用户程序并发。我们试想一下，刚把一个对象标记为白色，用户程序突然又引用了它，这种扫描操作就比较麻烦，于是引入了屏障技术。内存扫描和用户逻辑也可以并发执行，用户新建的对象认为是黑色的，已经扫描过的对象有可能因为用户逻辑造成对象状态发生改变。所以**对扫描过后的对象使用操作系统写屏障功能用来监控用户逻辑这段内存，一旦这段内存发生变化写屏障会发生一个信号，gc捕获到这个信号会重新扫描改对象，查看它的引用或者被引用是否发生改变，从而判断该对象是否应该被清理。因此通过写屏障技术，是的扫描操作也可以合用户程序并发执行。


　　**gc控制器**：gc算法并不万能的，针对不同的场景可能需要适当的设置。例如大数据密集计算可能不在乎内存使用量，甚至可以将gc关闭。golang 通过百分比来控制gc触发的时机，设置的百分比指的是程序新分配的内存与上一次gc之后剩余的内存量，例如上次gc之后程序占有2MB，那么下一次gc触发的时机是程序又新分配了2MB的内存。我们可以通过*SetGCPercent*函数动态设置，默认值为100，当百分比设置为负数时例如-1，表明关闭gc。
![SetGCPercent](https://github.com/wxquare/wxquare.github.io/raw/hexo/source/photos/gc_setGCPercent.jpg)


## 五、golang gc调优实例
gc 是golang程序性能优化非常重要的一部分，建议依照下面两个实例实践golang程序优化。
- https://tonybai.com/2015/08/25/go-debugging-profiling-optimization/
- https://blog.golang.org/profiling-go-programs
　　
	

参考：
- http://legendtkl.com/2017/04/28/golang-gc/
- https://www.jianshu.com/p/9c8e56314164
- https://blog.golang.org/ismmkeynote
- http://goog-perftools.sourceforge.net/doc/tcmalloc.html
- https://zhuanlan.zhihu.com/p/29216091

