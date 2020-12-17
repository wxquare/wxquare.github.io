---
title: 操作系统基础
categories: 
- C/C++
---


## CPU任务调度，进程/线程/协程
1. 进程和线程的区别，了解协程吗？CPU调度，数据共享。
2. 复杂系统中通常融合了多进程编程，多线程编程，协程编程
3. 进程之间怎么通信，线程通信通信，协程怎么通信
4. 进程之间怎么同步（信号量，自旋锁，屏障），线程之间怎么同步（锁），协程怎么同步。进程之间通过共享内存、管道、消息队列消息队列等方式通信，通过信号和信号量进行同步。线程在进程内部，全部变量时共享的，通过锁机制来同步。
5. 死锁：产生的四个条件、四个解决方法,死锁检测
6. 守护进程，linux系统编程实现守护进程
7. 在Linux上，对于多进程，子进程继承了父进程的下列哪些？堆栈、文件描述符、进程组、会话、环境变量、共享内存 
7. 僵尸进程和孤儿进程。孤儿进程：一个父进程退出，而它的一个或多个子进程还在运行，那么那些子进程将成为孤儿进程。孤儿进程将被init进程(进程号为1)所收养，并由init进程对它们完成状态收集工作。僵尸进程：一个进程使用fork创建子进程，如果子进程退出，而父进程并没有调用wait或waitpid获取子进程的状态信息，那么子进程的进程描述符仍然保存在系统中。这种进程称之为僵死进程。
8. 进程的状态。
	- TASK_RUNNING（运行态）：进bai程是可执行du的；或者正在执行，zhi或者在运行队列中等待执行。
	- TASK_INTERRUPTIBLE（可中断睡眠态）：进程被阻塞，等待某些条件的完成。一旦完成这些条件，内核就会将该进程的状态设置为运行态。
	- TASK_UNINTERRUPTIBLE（不可中断睡眠态）：进程被阻塞，等待某些条件的完成。与可中断睡眠态不同的是，该状态进程不可被信号唤醒。
	- TASK_ZOMBIE（僵死态）：该进程已经结束，但是其父进程还没有将其回收。
	- TASK_STOP（终止态）：进程停止执行。通常进程在收到SIGSTOP、SIGTTIN、SIGTTOU等信号的时候会进入该状态。
9. linux的CFS调度机制是什么？时间片/policy（进程类别）/priority（优先级）/counter。linux的任务调度机制是什么？在每个进程的task_struct结构中有以下四项：policy、priority、counter、rt_priority。这四项是选择进程的依据。其中，policy是进程的调度策略，用来区分实时进程和普通进程，实时进程优先于普通进程运行；priority是进程(包括实时和普通)的静态优先级；counter是进程剩余的时间片，它的起始值就是priority的值；由于counter在后面计算一个处于可运行状态的进程值得运行的程度goodness时起重要作用，因此，counter 也可以看作是进程的动态优先级。rt_priority是实时进程特有的，用于实时进程间的选择。 Linux用函数goodness()来衡量一个处于可运行状态的进程值得运行的程度。该函数综合了以上提到的四项，还结合了一些其他的因素，给每个处于可运行状态的进程赋予一个权值(weight)，调度程序以这个权值作为选择进程的唯一依据。
10. goroutine的GPM，没有时间片和优先级的概念，但也支持“抢占式调度”。 goroutine的主要状态grunnable、grunning、gwaiting
11. 线程的状态
	- runnable
	- running
	- blocked
	- dead
12. 进程、线程与协程的区别
13. 操作系统写时复制:https://juejin.cn/post/6844903702373859335
14. [操作系统为什么设计用户态和内核态，用户态和内核态的权限不同？怎么解决IO频繁发生内核和用户态的态的切换（缓存）？](https://imageslr.github.io/2020/07/07/user-mode-kernel-mode.html)
15. [select、epoll的监听回调机制，红黑树？](https://www.jianshu.com/p/31cdfd6f5a48)
16. [从一道面试题谈linux下fork的运行机制](https://www.cnblogs.com/leoo2sk/archive/2009/12/11/talk-about-fork-in-linux.html)


## 存储系统，内存和存储
1. 寄存器、缓存cache、内存和磁盘
2. 可执行文件的空间结构，进程的空间结构(虚拟地址空间，栈，堆，未初始化变量，初始化区，代码）
3. 查看进程使用的资源，top，ps，cat /proc/pid/status 
4. 进程的虚拟内存机制（虚拟地址-页表-物理地址）。Linux虚拟内存的实现需要6种机制的支持：地址映射机制、内存分配回收机制、缓存和刷新机制、请求页机制、交换机制和内存共享机制,内存管理程序通过映射机制把用户程序的逻辑地址映射到物理地址。当用户程序运行时，如果发现程序中要用的虚地址没有对应的物理内存，就发出了请求页要求。如果有空闲的内存可供分配，就请求分配内存(于是用到了内存的分配和回收)，并把正在使用的物理页记录在缓存中(使用了缓存机制)。如果没有足够的内存可供分配，那么就调用交换机制；腾出一部分内存。另外，在地址映射中要通过TLB(翻译后援存储器)来寻找物理页；交换机制中也要用到交换缓存，并且把物理页内容交换到交换文件中，也要修改页表来映射文件地址。
5. 操作系统内存分配算法常用缓存置换算法（FIFO，LRU，LFU），LRU算法的实现和优化？
6. Linux系统原理之文件系统（磁盘、分区、文件系统、inode表、data block）
7. 在linux执行ls上实际发生了什么
8. [CPU寻址过程](http://www.ssdfans.com/?p=105901),tlb,cache miss.

## 系统编程以及其它注意事项
1. 使用过哪些进程间通讯机制，并详细说明,linux进程之间的通信7种方式
2. 内核函数、系统调用、库函数/API,strace系统调用追踪调试
3. coredump文件产生？内存访问越界、野指针、堆栈溢出等等
4. fork 和 vfork，exec，system（进程的用户空间是在执行系统调用的fork时创建的，基于写时复制的原理，子进程创建的时候继承了父进程的用户空间，仅仅是mm_struc结构的建立、vm_area_struct结构的建立以及页目录和页表的建立，并没有真正地复制一个物理页面，这也是为什么Linux内核能迅速地创建进程的原因之一。）写时复制(Copy-on-write)是一种可以推迟甚至免除拷贝数据的技术。内核此时并不复制整个进程空间，而是让父进程和子进程共享同一个拷贝。只有在需要写入的时候，数据才会被复制，从而使各个进程拥有各自的拷贝。也就是说，资源的复制只有在需要写入的时候才进行，在此之前，以只读方式共享。这种技术使地址空间上的页的拷贝被推迟到实际发生写入的时候。有时共享页根本不会被写入，例如，fork()后立即调用exec()，就无需复制父进程的页了。fork()的实际开销就是复制父进程的页表以及给子进程创建唯一的PCB。这种优化可以避免拷贝大量根本就不会使用的数据
5. 锁？互斥锁的属性设置、多进程共享内存的使用、多线程的使用互斥锁、pshaed和type设置。使用互斥量和条件变脸实现互斥锁 
6. 共享内存的同步机制，使用信号量，无锁数据结构 
7.	多线程里一个线程sleep，实质上是在干嘛，忙等还是闲等。？
8.	exit()函数与_exit()函数最大的区别就在于exit()函数在调用exit系统调用之前要检查文件的打开情况，把文件缓冲区中的内容写回文件，就是"清理I/O缓冲"。
9.  select/epoll https://www.cnblogs.com/anker/p/3265058.html
- select 内核态和用户态重复拷贝
- select 需要遍历遍历查找就绪的socket
- select 有数量限制1024
- epoll 注册时写进内核
- epoll_wait 返回就绪的事件


## 网络编程
1.	简单了解C语言的socket编程api。socket，bind，listen，accept，connect，read/write.
2.	Linux下socket的五种I/O 模式，同步阻塞、同步非阻塞、同步I/O复用、异步I/O、信号驱动I/O
3.	[Linux套接字和I/O模型](https://www.cnblogs.com/wxquare/archive/2004/01/13/6802078.html)
4.	select和epoll的区别
5.	什么是I/O 复用？关于I/O多路复用(又被称为“事件驱动”)，首先要理解的是，操作系统为你提供了一个功能，当你的某个socket可读或者可写的时候，它可以给你一个通知。这样当配合非阻塞的socket使用时，只有当系统通知我哪个描述符可读了，我才去执行read操作，可以保证每次read都能读到有效数据而不做纯返回-1和EAGAIN的无用功。写操作类似。操作系统的这个功能通过select/poll/epoll/kqueue之类的系统调用函数来使用，这些函数都可以同时监视多个描述符的读写就绪状况，这样，多个描述符的I/O操作都能在一个线程内并发交替地顺序完成，这就叫I/O多路复用，这里的“复用”指的是复用同一个线程。

6.	网络分析工具。ping/tcpdump/netstat/lsof