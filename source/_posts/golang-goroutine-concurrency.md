---
title: golang 协程与并发
---

## 一、基本概念
1. 并行与串行： 程序能否利用多核物理设备同一时刻执行多个任务，并行依赖多核的支持
2. 并发： 程序在同一时刻执行过个任务，并发不需要多核的支持，在单核处理器上能以间隔方式切换不同的任务
3. 进程
4. 线程
5. 协程，协程更为轻量，协程栈2KB，线程栈MB级别

## 二、Goroutine
1. 通过关键字go创建并发任务单元，而不是执行并发操作。新建的任务会被放置在系统调度队列中，等待调度器安排合适的系统线程去获取执行权。通过go创建并发单元不会导致阻塞，不会等待改任务启动，golang运行时也不保证并发任务的执行顺序，这意味先创建的任务可能会比后创建的任务晚执行。
2. 进程退出时，不会等待goroutine执行结束，需要使用channel或者sync等同步手段。
3. golang运行时可以创建很多线程，但任何时候仅有限个线程参与并发任务执行。该数量通常默认与处理器核数相等，可通过runtime.MAXPROCS函数修改。
4. 与线程不同，goroutine任务无法设置优先级，无法获取编号，甚至无法获取返回值。只能通过在协程外部定义变量，以参数的形式传递给协程，同时需要做并发保护。
5. 操作系统在线程调度时具有时间片抢占的概念，意味着线程不会一直占有某处理器。而协程goroutine一旦被调度，在没有阻塞、系统调用、IO等情况下，将一直占有cpu，不会被其它协程抢占。协程可通过runtime.Gosched函数主动释放线程器质性其它任务，等待下次调度时恢复执行。


## 三、CSP并发模式
　Go鼓励使用CSP协程并发模型，以通信来代替内存共享，而不是以内存共享来通信。因此channel对于golang并发来说至关重要。**Don't communicate by sharing memory,share memory by communicating.** 另外，golang提供sync包，互斥锁、读写锁和原子操作帮助更好的编写并发代码。除此之外golang提供context包管理协程之间的关系。channel不是用来代替锁的，它们有各自不同的应用场景，通道倾向于解决协程之间的逻辑层次，而锁则用来保护局部数据的安全。
- channel：参考
- sync: Mutex和RWMutex的使用并不复杂，有以下几点需要注意：
	a、使用Mutex作为匿名字段时，相关方法必须实现为pointer-receiver,否则会因为复制导致锁失效
	b、应该将锁粒度控制在最小范围，及早释放，考虑到性能，不要一昧的使用defer unlock
	c、mutex不支持递归锁，即使在同一goroutine下也会导致死锁
	d、读写并发时，用RWMutex性能会好一些
	e、对单个数据的读写保护，可使用原子操作
- context：由于任务复杂，常会存在协程嵌套，context能帮助更好的管理协程之间的关系


## 四、协程调度
上文讲过go关键字只是创建协程并发任务，并不是立刻执行，需要等待运行时tuntime的调度。接下来介绍goroutine的GMP调度模型。
### 4.1. 操作系统线程调度与golang协程调度
操作系统线程并发：
- 创建线程和切换线程代价较大，线程数量不能太多，经常采用线程池或者网络IO复用技术，因此线程调度难以扩展
- 线程的同步和通信较为麻烦
- 加锁易犯错且易效率低
协程并发：  
- 创建协程goroutine的代价低
- 协程数量大，可达数十万个
- 协程的同步和通信机制简单，基于channel  
- G-M-P调度模型较为高效，实现协程阻塞、抢占式调度、stealing等情况，具有较高的调度效率  

### 4.2. Golang运行时调度器
golang运行时调度器位于用户golang代码和操作系统os之间，它决定何时哪个goroutine将获得资源开始执行、哪个goroutine应该停止执行让出资源、哪个goroutine应该被唤醒恢复执行等。由于操作系统是以线程为调度的单位，因此golang运行时调度器实际上是将协程调度到具体的线程上。随着golang版本的更新，其调度模型也在不断的优化，goalng 1.1版本中的G-P-M模型使其调度模型基本成型，也具有较高的效率。为了实现调度的可扩展性（scalable），在协程和线程之间增加了一个逻辑层P。
- goroutine 都由一个G结构表示，它管理着goroutine的栈和状态
- 运行时管理着G，并将它们映射到Logical Processor P上。P可以看作是一个抽象的资源或者一个上下文
- 为了运行goroutine，M需要持有上下文P，M会从P的queue弹出一个goutine并执行。

### 4.3 抢占式调度
　和操作系统按时间片调度线程不同，Go并没有时间片的概念。如果某个G没有进行system call调用、没有进行I/O操作、没有阻塞在一个channel操作上，那么m是如何让G停下来并调度下一个runnable G的呢？
### 4.4 channel阻塞或者network I/O情况下的调度
　如果G被阻塞在某个channel操作或network I/O操作上时，G会被放置到某个wait队列中，而M会尝试运行下一个runnable的G；如果此时没有runnable的G供m运行，那么m将解绑P，并进入sleep状态。当I/O available或channel操作完成，在wait队列中的G会被唤醒，标记为runnable，放入到某P的队列中，绑定一个M继续执行。
### 4.5 system call阻塞状态下的调度
如果G被阻塞在某个system call操作上，那么不光G会阻塞，执行该G的M也会解绑P(实质是被sysmon抢走了)，与G一起进入sleep状态。如果此时有idle的M，则P与其绑定继续执行其他G；如果没有idle M，但仍然有其他G要去执行，那么就会创建一个新M。
当阻塞在syscall上的G完成syscall调用后，G会去尝试获取一个可用的P，如果没有可用的P，那么G会被标记为runnable，之前的那个sleep的M将再次进入sleep。

### 4.6 golang调度器的跟踪调试
https://colobu.com/2016/04/19/Scheduler-Tracing-In-Go/

参考：  
- https://tonybai.com/2017/06/23/an-intro-about-goroutine-scheduler/  
- https://colobu.com/2017/05/04/go-scheduler/  