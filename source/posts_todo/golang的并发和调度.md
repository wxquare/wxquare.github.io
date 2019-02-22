##1.C/C++操作系统线程调度的缺点
- 创建线程和切换线程代价较大，线程数量不能太多，经常采用线程池或者网络IO复用技术，因此线程调度难以扩展  
- 线程的同步和通信较为麻烦
- 加锁易犯错且易效率低


##2.Golang运行时的协程调度的特点  
- 创建协程goroutine的代价低
- 协程数量大，可达数十万个
- 协程的同步和通信机制简单，基于channel  
- G-M-P调度模型较为高效，实现协程阻塞、抢占式调度、stealing等情况，具有较高的调度效率


##3.Golang运行时调度器
golang运行时调度器位于用户golang代码和操作系统os之间，它决定何时哪个goroutine将获得资源开始执行、哪个goroutine应该停止执行让出资源、哪个goroutine应该被唤醒恢复执行等。由于操作系统是以线程为调度的单位，因此golang运行时调度器实际上是将协程调度到具体的线程上。

随着golang版本的更新，其调度模型也在不断的优化，goalng 1.1版本中的G-P-M模型使其调度模型基本成型，也具有较高的效率。为了实现调度的可扩展性（scalable），在协程和线程之间增加了一个逻辑层P。

- goroutine 都由一个G结构表示，它管理着goroutine的栈和状态
- 运行时管理着G，并将它们映射到Logical Processor P上。P可以看作是一个抽象的资源或者一个上下文
- 为了运行goroutine，M需要持有上下文P，M会从P的queue弹出一个goutine并执行



##4.其它概念：
###4.1抢占式调度
和操作系统按时间片调度线程不同，Go并没有时间片的概念。如果某个G没有进行system call调用、没有进行I/O操作、没有阻塞在一个channel操作上，那么m是如何让G停下来并调度下一个runnable G的呢？答案是：G是被抢占调度的
###4.2channel阻塞或者network I/O情况下的调度
如果G被阻塞在某个channel操作或network I/O操作上时，G会被放置到某个wait队列中，而M会尝试运行下一个runnable的G；如果此时没有runnable的G供m运行，那么m将解绑P，并进入sleep状态。当I/O available或channel操作完成，在wait队列中的G会被唤醒，标记为runnable，放入到某P的队列中，绑定一个M继续执行。
###4.3system call阻塞状态下的调度
如果G被阻塞在某个system call操作上，那么不光G会阻塞，执行该G的M也会解绑P(实质是被sysmon抢走了)，与G一起进入sleep状态。如果此时有idle的M，则P与其绑定继续执行其他G；如果没有idle M，但仍然有其他G要去执行，那么就会创建一个新M。

当阻塞在syscall上的G完成syscall调用后，G会去尝试获取一个可用的P，如果没有可用的P，那么G会被标记为runnable，之前的那个sleep的M将再次进入sleep。

##5.golang调度器的跟踪调试
https://colobu.com/2016/04/19/Scheduler-Tracing-In-Go/

参考：  
- https://tonybai.com/2017/06/23/an-intro-about-goroutine-scheduler/  
- https://colobu.com/2017/05/04/go-scheduler/  