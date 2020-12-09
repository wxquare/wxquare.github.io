---
title: Golang 并发编程和GMP调度模型
categories:
- Golang
---


# 并发编程的模式
1. 多进程
2. 多线程
3. 线程池
4. Goroutine 之类的协程
5. 同步和异步
6. 阻塞和非阻塞
## 问题1：进程，线程和协程的区别？
- 进程是系统进行资源分配和调度的一个独立单位,每个进程都有自己的独立内存空间,栈、寄存器、虚拟内存、文件句柄等
- 线程是进程的一个实体,是CPU调度和分派的基本单位,在运行中必不可少的资源,如程序计数器,一组寄存器和栈
- 协程是一种用户态的轻量级线程，协程的调度完全由用户控制,协程拥有自己的寄存器上下文和栈,直接操作栈则基本没有内核切换的开销
## 问题2：进程、线程、协程切换分别包含哪些内容?
- 进程因为有自己独立的地址空间，所以进程切换时需要切换页目录以使用新的地址空间，除此之外也需要切换内核栈和上下文环境
- 线程的调度只有拥有最高权限的内核空间才可以完成，所以线程的切换涉及到用户空间和内核空间的切换,也就是特权模式切换，然后需要操作系统调度模块完成线程调度（taskstruct）.
- 协程切换只涉及基本的CPU上下文切换，所谓的 CPU 上下文，就是一堆寄存器，里面保存了 CPU运行任务所需要的信息：从哪里开始运行（%rip：指令指针寄存器，标识 CPU 运行的下一条指令），栈顶的位置（%rsp： 是堆栈指针寄存器，通常会指向栈顶位置），当前栈帧在哪（%rbp 是栈帧指针，用于标识当前栈帧的起始位置）以及其它的CPU的中间状态或者结果（%rbx，%r12，%r13，%14，%15 等等）。协程切换非常简单，就是把当前协程的 CPU 寄存器状态保存起来，然后将需要切换进来的协程的 CPU 寄存器状态加载的 CPU 寄存器上就 ok 了。而且完全在用户态进行
## 问题3：多进程、多线程、多协程编程有哪些优缺点?
- 多进程的优点是稳定性好，一个子进程崩溃了，不会影响主进程以及其余进程.多进程编程也有不足，即创建进程的代价非常大
- 多线程编程的优点是效率较高一些，适用于批处理任务等功能；不足之处在于，任何一个线程崩溃都可能造成整个进程的崩溃，因为它们共享了进程的内存资源池
- 协程并发粒度小，可以同时进行的并发量比较大。缺点是需要用户自己调度。Golang 自己内置了runtime调度器;创建协程goroutine的代价低，通常为KB级别，因此协程数量大，可达数十万个


# 执行体之间的通信方式
1. 进程间的通信方式
- 共享内存、socket、管道、信号、共享队列
2. 线程间的通信方式
- 共享内存加锁
3. Goroutine 之间的通信方式
- 基于channel的CSP模式
- 共享内存加锁


# 执行体之间的同步原语和锁
## C++
## Golang
1. [互斥锁,sync.Mutex](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-sync-primitives/#mutex)
- 饥饿模式
- 普通模式
- 自旋
2. [读写锁，sync.RWMtex](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-sync-primitives/#rwmutex)
3. sync.Waitgroup
4. 单例模式,[sync.Once](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-sync-primitives/#once)
5. sync.Cond
6. [channel](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-channel/#642-%E6%95%B0%E6%8D%AE%E7%BB%93%E6%9E%84)
7. [context](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-context/)
8. sync.atomic
9. sync.map
- 双map,read 和 dirty
- lock
- https://colobu.com/2017/07/11/dive-into-sync-Map/
- https://segmentfault.com/a/1190000020946989
- https://wudaijun.com/2018/02/go-sync-map-implement/
- load,store,delete 的流程


# 执行体的调度模型GMP
1. 操作系统os的调度
- 非抢占式（nonpreemptive）调度算法：挑选一个进程让它一直执行直至被阻塞或自动释放CPU（在这种情况下，该进程若交出CPU都是自愿的）
- 抢占式（preemptive）调度算法：挑选一个进程让它运行某个固定时间的最大值，结束时会被挂起，调度程序会选择另一个合适的进程来运行（优先级高的先调度），必须要有可用的时钟来发生时钟中断，不然只能用非抢占式调度算法
- 先来先服务（FCFS：first-come first-served）
- 时间片轮转调度（Round Robin，RR）
- 优先级调度（Priority Schedule）
- 多级队列（Multilevel Queue）
2. [golang 运行时GMP的调度]()
- G/M/P
- 基于协作的抢占式调度器


并发日常编程必不可少的内容，可以选择多进程、多线程、多协程、池化技术，golang协程并发具有以下这些优势：
1. 编码比较简单
2. 推崇基于channel的CSP模式，避免了重复加锁出错的概率
3. 同步原语比较丰富
4. 相比进程、线程MB级别的代价，创建协程的代价比较低，通常为KB级别。
5. 相比进程、线程并发调度时的切换，线程切换在非系统调用阻塞下通常只在用户态进行
6. golang GMP实现了通过较少内核线程调度多个协程
7. 因此goroutine具有粒度更小，高效调度且易用的特点。