---
title: Golang 基础知识汇总
categories:
- 计算机基础
---


### Go 和 C++ 语言对比
Go and C++ are two different programming languages with different design goals, syntax, and feature sets. Here's a brief comparison of the two:

Syntax: Go has a simpler syntax than C++. It uses indentation for block structure and has fewer keywords and symbols. C++ has a more complex syntax with a lot of features that can make it harder to learn and use effectively.

Memory Management: C++ gives the programmer more control over memory management through its support for pointers, manual memory allocation, and deallocation. Go, on the other hand, uses a garbage collector to automatically manage memory, making it less error-prone.

Concurrency: Go has built-in support for concurrency through goroutines and channels, which make it easier to write concurrent code. C++ has a thread library that can be used to write concurrent code, but it requires more manual management of threads and locks.

Performance: C++ is often considered a high-performance language, and it can be used for system-level programming and performance-critical applications. Go is also fast but may not be as fast as C++ in some cases.

Libraries and Frameworks: C++ has a vast ecosystem of libraries and frameworks that can be used for a variety of applications, from game development to machine learning. Go's ecosystem is smaller, but it has good support for web development and distributed systems.

Overall, the choice of programming language depends on the project requirements, the available resources, and the developer's expertise. Both Go and C++ have their strengths and weaknesses, and the best choice depends on the specific needs of the project.


### string/[]byte
- string是golang的基本数组类型，s := "hello,world"，一旦初始化后不允许修改其内容
- [内部实现结构](https://draveness.me/golang/docs/part2-foundation/ch03-datastructure/golang-string/)，指向数据的指针data和表示长度的len
- 字符串拼接和格式化四种方式，+=，strings.join,buffer.writestring,fmt.sprintf
- [string 与 []byte的类型转换](https://www.cnblogs.com/shuiyuejiangnan/p/9707066.html)
- <font color=red>字符串与数值类型的不能强制转化，要使用strconv包中的parse和format函数</font>
- 标准库strings提供了许多字符串操作的函数,例如Split、HasPrefix,Trim。

### array 
- 数组array: [3]int{1,2,3}
- <font color=red>**数组是值类型**</font>，数组传参发生拷贝
- 定长
- 数组的创建、初始化、访问和遍历range，len(arr)求数组的长度
  
### slice
- 切片slice初始化: make([]int,len,cap)
- <font color=red>**slice是引用类型**</font>
- 变长，用容量和长度的区别，分别使用cap和len函数获取
- 内存结构和实现：指针、cap、size共24字节
- 常用函数，append，cap，len
- 切片动态扩容
- 深拷贝copy和浅拷贝“=”的区别

### map
- 引用类型，需要初始化 make(map[string]int,5) 
- [内部实现的数据结构，hmap、bmap等](https://draveness.me/golang/docs/part2-foundation/ch03-datastructure/golang-hashmap/)
- 链地址法解决冲突
- hashmap中buckets为什么为通常2的幂次方
- 访问流程，先用低位确定bucket，再用高8位粗选
- 增量扩容，迁移
- 使用map[interface{}]struct{}作为set
- [map遍历是无序且随机的，如何实现顺序遍历？](https://blog.csdn.net/slvher/article/details/44779081)
- [内部hashmap的实现原理](https://ninokop.github.io/2017/10/24/Go-Hashmap%E5%86%85%E5%AD%98%E5%B8%83%E5%B1%80%E5%92%8C%E5%AE%9E%E7%8E%B0/)。内部结构（bucket），扩容与迁移，删除。 

### sync.map
- 双map,read 和 dirty
- 以空间换效率，通过read和dirty两个map来提高读取效率
- 优先从read map中读取(无锁)，否则再从dirty map中读取(加锁)
- 动态调整，当misses次数过多时，将dirty map提升为read map
- 延迟删除，删除只是为value打一个标记，在dirty map提升时才执行真正的删除
- [sync.map 揭秘](https://colobu.com/2017/07/11/dive-into-sync-Map/)
- [sync.map读写流程图](https://segmentfault.com/a/1190000020946989)
- https://wudaijun.com/2018/02/go-sync-map-implement/


### struct
- 空结构体struct{}的用途，节省内存。
- 不支持继承，使用结构体嵌套组合
- struct 可以比较吗？普通struct可以比较，带引用的struc不可比较，需要使用reflect.DeepEqual
- struct没有slice和map类型时可直接判断
- slice和map本身不可比较，需要使用reflect.DeepEqual()。
- struct中包含slice和map等字段时，也要使用reflect.DeepEqual().
- https://stackoverflow.com/questions/24534072/how-to-compare-struct-slice-map-are-equal


### 类型和拷贝方式
- 值类型 ：String，Array，Int，Struct，Float，Bool，pointer（深拷贝）
- 引用类型：Slice，Map （浅拷贝）

### 函数和方法，匿名函数
- init函数
- 值接收和指针接收的区别
- 匿名函数？闭包？闭包延时绑定问题？用闭包写fibonacci数列？

### interface
- https://draveness.me/golang/docs/part2-foundation/ch04-basic/golang-interface/
- **隐式接口**，实现接口的所有方法就隐式地实现了接口；不需要显示申明实现某接口
- **接口也是 Go 语言中的一种类型**，它能够出现在变量的定义、函数的入参和返回值中并对它们进行约束，不过 Go 语言中有两种略微不同的接口，一种是带有一组方法的接口，另一种是不带任何方法的 interface{}：
- interface{} 类型不是任意类型，而是将类型转换成了 interface{} 类型
- 结构体实现接口 vs 结构体指针实现接口 区别？
- runtime.eface 和 runtime.iface 结构？
- **结构体类型转化为接口的类型相互变换，interface类型断言为struct类型** 过程
- **动态派发与多态**。动态派发（Dynamic dispatch）是在运行期间选择具体多态操作（方法或者函数）执行的过程，它是面向对象语言中的常见特性6。Go 语言虽然不是严格意义上的面向对象语言，但是接口的引入为它带来了动态派发这一特性，调用接口类型的方法时，如果编译期间不能确认接口的类型，Go 语言会在运行期间决定具体调用该方法的哪个实现。
- Golang没有泛型，通过interface可以实现简单泛型编程，例如的sort的实现
- 接口实现的源码

### channel
- Go鼓励CSP模型(communicating sequential processes),Goroutin之间通过channel传递数据
- 非缓冲的同步channel和带缓冲的异步channel
- [内部实现结构，带锁的循环队列runtime.hchan](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-channel/#642-%E6%95%B0%E6%8D%AE%E7%BB%93%E6%9E%84)
- channel创建make
- chan <- i
- **向channel发送数据**。在发送数据的逻辑执行之前会先为当前 Channel 加锁，防止多个线程并发修改数据。如果 Channel 已经关闭，那么向该 Channel 发送数据时会报 “send on closed channel” 错误并中止程序。分为的三个部分：
  当存在等待的接收者时，通过 runtime.send 直接将数据发送给阻塞的接收者；
  当缓冲区存在空余空间时，将发送的数据写入 Channel 的缓冲区；
  当不存在缓冲区或者缓冲区已满时，等待其他 Goroutine 从 Channel 接收数据；
- i <- ch，i, ok <- ch
- **从channel接收数据**的五种情况:
  - 如果 Channel 为空，那么会直接调用 runtime.gopark 挂起当前 Goroutine；
  - 如果 Channel 已经关闭并且缓冲区没有任何数据，runtime.chanrecv 会直接返回；
  - 如果 Channel 的 sendq 队列中存在挂起的 Goroutine，会将 recvx 索引所在的数据拷贝到接收变量所在的内存空间上并将 sendq 队列中 Goroutine 的数据拷贝到缓冲区；
  - 如果 Channel 的缓冲区中包含数据，那么直接读取 recvx 索引对应的数据；
  - 在默认情况下会挂起当前的 Goroutine，将 runtime.sudog 结构加入 recvq 队列并陷入休眠等待调度器的唤醒；
- **关闭channel**
- 如何优雅的关闭channel？https://www.jianshu.com/p/d24dfbb33781, channel关闭后读操作会发生什么？写操作会发生什么？

### 指针和unsafe.Pointer
- 相比C/C++，为了安全性考虑，Go指针弱化。不同类型的指针不能相互转化，指针变量不支持运算，不支持c/c++中的++，需要借助unsafe包
- 任何类型的指针都可以被转换成unsafe.Pointer类型，通过unsafe.Pointer实现不同类型指针的转化
- uintptr值可以被转换成unsafe.Pointer类型，通过uintptr实现指针的运算
- unsafe.Pointer是一个指针类型，指向的值不能被解析，类似于C/C++里面的(void *)，只说明这是一个指针，但是指向什么的不知道。
- uintptr 是一个整数类型，这个整数的宽度足以用来存储一个指针类型数据；那既然是整数类类型，当然就可以对其进行运算了
- nil
- [实践string和[]byte的高效转换](https://www.cnblogs.com/shuiyuejiangnan/p/9707066.html)
- 在业务场景中，使用指针虽然方便，但是要注意深拷贝和浅拷贝，这种错误还是比较常见的
- 当你对象是结构体对象的指针时，你想要获取字段属性时，可以直接使用'.'，而不需要解引用

### 集合set
1. golang中本身没有提供set，但可以通过map自己实现
2. 利用map键值不可重复的特性实现set，value为空结构体。 map[interface{}]struct{} 
3. [如何自己实现set？](https://studygolang.com/articles/11179)


### defer
- defer定义的延迟函数参数在defer语句出时就已经确定下来了
- defer定义顺序与实际执行顺序相反
- return不是原子操作，执行过程是: 保存返回值(若有)-->执行defer（若有）-->执行ret跳转
- 申请资源后立即使用defer关闭资源是好习惯
- golang中的defer用途？调用时机？调用顺序？预计算值？
- [defer 实现原理？](https://blog.csdn.net/Tybyqi/article/details/83827140)


### Go 错误处理 error、panic
- 在Go 语言中，错误被认为是一种可以预期的结果；而异常则是一种非预期的结果，发生异常可能表示程序中存在BUG 或发生了其它不可控的问题。 
- Go 语言推荐使用 recover 函数将内部异常转为错误处理，这使得用户可以真正的关心业务相关的错误处理。
- 在Go服务中通常需要自定义粗错误类型，最好能有效区分业务逻辑错误和系统错误，同时需要捕获panic，将panic转化为error，避免某个错误影响server重启
- panic 时需要保留runtime stack
```
  defer func() {
		if x := recover(); x != nil {
			panicReason := fmt.Sprintf("I'm panic because of: %v\n", x)
			logger.LogError(panicReason)
			stk := make([]byte, 10240)
			stkLen := runtime.Stack(stk, false)
			logger.LogErrorf("%s\n", string(stk[:stkLen]))
		}
	}()
 ```

## How does Go handle concurrency? (Goroutine,GMP调度模型，channel)
### what's CSP?
The **Communicating Sequential Processes (CSP) model** is a theoretical model of concurrent programming that was first introduced by Tony Hoare in 1978. The CSP model is based on the idea of concurrent processes that communicate with each other by sending and receiving messages through channels.The Go programming language provides support for the CSP model through its built-in concurrency features, such as goroutines and channels. In Go, concurrent processes are represented by goroutines, which are lightweight threads of execution. The communication between goroutines is achieved through channels, which provide a mechanism for passing values between goroutines in a safe and synchronized manner.

### Which is Goroutine ?
- Goroutines are lightweight, user-level threads of execution that run concurrently with other goroutines within the same process.
- Unlike traditional threads, goroutines are managed by the Go runtime, which automatically schedules and balances their execution across multiple CPUs and makes efficient use of available system resources.

### 比较Goroutine、thread、process
- 比较进程、线程和Goroutine。进程是资源分配的单位，有独立的地址空间，线程是操作系统调度的单位，协程是更细力度的执行单元，需要程序自身调度。Go语言原生支持Goroutine，并提供高效的协程调度模型。
- Goroutines, threads, and processes are all mechanisms for writing concurrent and parallel code, but they have some important differences:
- Goroutines: A goroutine is a lightweight, user-level thread of execution that runs concurrently with other goroutines within the same process. Goroutines are managed by the Go runtime, which automatically schedules and balances their execution across multiple CPUs. Goroutines require much less memory and have much lower overhead compared to threads, allowing for many goroutines to run simultaneously within a single process.
- Threads: A thread is a basic unit of execution within a process. Threads are independent units of execution that share the same address space as the process that created them. This allows threads to share data and communicate with each other, but also introduces the need for explicit synchronization to prevent race conditions and other synchronization issues.
- Processes: A process is a self-contained execution environment that runs in its own address space. Processes are independent of each other, meaning that they do not share memory or other resources. Communication between processes requires inter-process communication mechanisms, such as pipes, sockets, or message queues.
- In general, goroutines provide a more flexible and scalable approach to writing concurrent code compared to threads, as they are much lighter and more efficient, and allow for many more concurrent units of execution within a single process. Processes provide a more secure and isolated execution environment, but have higher overhead and require more explicit communication mechanisms.

### Why is Goroutine lighter and more efficient than thread or process?
- Stack size: Goroutines have a much smaller stack size compared to threads. The stack size of a goroutine is dynamically adjusted by the Go runtime, based on the needs of the goroutine. This allows for many more goroutines to exist simultaneously within a single process, as they require much less memory.
- Scheduling: Goroutines are scheduled by the Go runtime, which automatically balances and schedules their execution across multiple CPUs. This eliminates the need for explicit thread management and synchronization, reducing overhead.
- Context switching: Context switching is the process of saving and restoring the state of a running thread in order to switch to a different thread. Goroutines have a much lower overhead for context switching compared to threads, as they are much lighter and require less state to be saved and restored.
- Resource sharing: Goroutines share resources with each other and with the underlying process, eliminating the need for explicit resource allocation and deallocation. This reduces overhead and allows for more efficient use of system resources.
- Overall, the combination of a small stack size, efficient scheduling, low overhead context switching, and efficient resource sharing makes goroutines much lighter and more efficient than threads or processes, and allows for many more concurrent units of execution within a single process.
- Goroutine 上下文切换只涉及到三个寄存器（PC / SP / DX）的值修改；而对比线程的上下文切换则需要涉及模式切换（从用户态切换到内核态）、以及 16 个寄存器、PC、SP...等寄存器的刷新；内存占用少：线程栈空间通常是 2M，Goroutine 栈空间最小 2K；Golang 程序中可以轻松支持10w 级别的 Goroutine 运行，而线程数量达到 1k 时，内存占用就已经达到 2G。
- 理解G、P、M的含义以及调度模型

### How are goroutines scheduled by runtime?
- **Cooperative** (协作式). The scheduler uses a **cooperative** scheduling model, which means that goroutines voluntarily yield control to the runtime when they are blocked or waiting for an event. 
- **Timer-based preemption**. The scheduler uses a technique called **timer-based preemption** to interrupt the execution of a running goroutine and switch to another goroutine if it exceeds its time slice
- **Work-stealing**. The scheduler uses a work-stealing algorithm, where each CPU has its own local run queue, and goroutines are dynamically moved between run queues to balance the o balance the load and improve performance.
- **no explicit prioritization**. The Go runtime scheduler does not provide explicit support for prioritizing goroutines. Instead, it relies on the cooperative nature of goroutines to ensure that all goroutines make progress. In a well-designed Go program, the program should be designed such that all goroutines make progress in a fair and balanced manner.
- https://blog.csdn.net/sinat_34715587/article/details/124990458
- G 的数量可以远远大于 M 的数量，换句话说，Go 程序可以利用少量的内核级线程来支撑大量 Goroutine 的并发。多个 Goroutine 通过用户级别的上下文切换来共享内核线程 M 的计算资源，但对于操作系统来说并没有线程上下文切换产生的性能损耗，支持任务窃取（work-stealing）策略：为了提高 Go 并行处理能力，调高整体处理效率，当每个 P 之间的 G 任务不均衡时，调度器允许从 GRQ，或者其他 P 的 LRQ 中获取 G 执行。
- 减少因Goroutine创建大量M：
  -  由于原子、互斥量或通道操作调用导致 Goroutine 阻塞，调度器将把当前阻塞的 Goroutine 切换出去，重新调度 LRQ 上的其他 Goroutine；
  -  由于网络请求和 IO 操作导致 Goroutine 阻塞，通过使用 NetPoller 进行网络系统调用，调度器可以防止 Goroutine 在进行这些系统调用时阻塞 M。这可以让 M 执行 P 的 LRQ 中其他的 Goroutines，而不需要创建新的 M。有助于减少操作系统上的调度负载。
  -  当调用一些系统方法的时候，如果系统方法调用的时候发生阻塞，这种情况下，网络轮询器（NetPoller）无法使用，而进行系统调用的 Goroutine 将阻塞当前 M，则创建新的M。阻塞的系统调用完成后：M1 将被放在旁边以备将来重复使用
  -  如果在 Goroutine 去执行一个 sleep 操作，导致 M 被阻塞了。Go 程序后台有一个监控线程 sysmon，它监控那些长时间运行的 G 任务然后设置可以强占的标识符，别的 Goroutine 就可以抢先进来执行。

### What are the states of Goroutine and how do they flow?
- 协程的状态流转？Grunnable、Grunning、Gwaiting
- In Go, a Goroutine can be in one of several states during its lifetime. The states are:
- New: The Goroutine is created but has not started executing yet.
- Running: The Goroutine is executing on a machine-level thread.
- Waiting: The Goroutine is waiting for some external event, such as I/O, channel communication, or a timer.
- Sleeping: The Goroutine is sleeping, or waiting for a specified amount of time.
- Dead: The Goroutine has completed its execution and is no longer running.

In summary, the lifetime of a Goroutine in Go starts when it is created and ends when it completes its execution or encounters a panic, and can be influenced by synchronization mechanisms such as channels and wait groups.

### 生产者、消费者模型，并行计算累加求和
```Go
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func produce(mq chan<- int) {
	rand.Seed(time.Now().UnixNano())
	limitGoroutine := 2
	cnt := 100000
	var wg sync.WaitGroup
	for i := 0; i < limitGoroutine; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := start; j < cnt; j += limitGoroutine {
				num := rand.Intn(100)
				mq <- num
			}
		}(i)
	}
	go func() {
		wg.Wait()
		close(mq)
	}()
}

func consume(nums <-chan int) int {
	limitGoroutine := 4
	resChan := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < limitGoroutine; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var sum int = 0
			for num := range nums {
				sum += num
			}
			resChan <- sum
		}()
	}
	go func() {
		wg.Wait()
		close(resChan)
	}()
	var finalRes int = 0
	for r := range resChan {
		finalRes += r
	}
	return finalRes
}

func main() {
	mq := make(chan int, 10)
	go produce(mq)
	res := consume(mq)
	fmt.Printf("%+v\n", res)
}
```

-  golang context 用于在树形goroutine结构中，通过信号减少资源的消耗，包含Deadline、Done、Error、Value四个接口
-  常用的同步原语：channel、sync.mutex、sync.RWmutex、sync.WaitGroup、sync.Once、atomic
- 协程的状态流转？Grunnable、Grunning、Gwaiting
- sync.Mutex 和 sync.RWMutex 互斥锁和读写锁的使用场景？
- [golang 协程优雅的退出？](https://segmentfault.com/a/1190000017251049)
- 用channel实现定时器？（实际上是两个协程同步）
- 深入理解协程gmp调度模型，以及其发展历史
- 理解操作系统是怎么调度的，golang协程调度的优势，切换代价低，goroutine开销低，并发度高。
- Golang IO 模型和网络轮训器
- [sync.Mutex: “锁”实现背后那些事](http://km.oa.com/articles/show/502088)


## Golang 内存管理和垃圾回收（memory and gc）
### gc 的过程
- Marking phase: In this phase, the Go runtime identifies all objects that are accessible by the program and marks them as reachable. Objects that are not marked as reachable are considered unreachable and eligible for collection.
- Sweeping phase: In this phase, the Go runtime scans the memory heap and frees all objects that are marked as unreachable. The memory space occupied by these objects is now available for future allocation.
- Compacting phase: In this phase, the Go runtime rearranges the remaining objects on the heap to reduce fragmentation and minimize the impact of future allocations and deallocations.


### What are the memory leak scenarios in Go language?
- Goroutine leaks: If a goroutine is created and never terminated, it can result in a memory leak. This can occur when a program creates a goroutine to perform a task but fails to provide a mechanism for the goroutine to terminate, such as a channel to receive a signal to stop.
- Leaked closures: Closures are anonymous functions that capture variables from their surrounding scope. If a closure is created and assigned to a global variable, it can result in a memory leak, as the closure will continue to hold onto the captured variables even after they are no longer needed.
- Incorrect use of channels: Channels are a mechanism for communicating between goroutines. If a program creates a channel but never closes it, it can result in a memory leak. Additionally, if a program receives values from a channel but never discards them, they will accumulate in memory and result in a leak.
- Unclosed resources: In Go, it's important to close resources, such as files and network connections, when they are no longer needed. Failure to do so can result in a memory leak, as the resources and their associated memory will continue to be held by the program.
- Unreferenced objects: In Go, unreferenced objects are objects that are no longer being used by the program but still exist in memory. This can occur when an object is created and never explicitly deleted or when an object is assigned a new value and the old object is not properly disposed of.
By following best practices and being mindful of these common scenarios, you can help to avoid memory leaks in your Go programs. Additionally, you can use tools such as the Go runtime profiler to detect and diagnose memory leaks in your programs.


- **多级缓存**：内存分配器不仅会区别对待大小不同的对象，还会将内存分成不同的级别分别管理，TCMalloc 和 Go 运行时分配器都会引入线程缓存（Thread Cache）、中心缓存（Central Cache）和页堆（Page Heap）三个组件分级管理内存
- **对象大小**：Go 语言的内存分配器会根据申请分配的内存大小选择不同的处理逻辑，运行时根据对象的大小将对象分成微对象、小对象和大对象三种，tiny,small,large
- mspan、mcache、mcentral、mheap
- [深入理解golang GC的演进过程](https://segmentfault.com/a/1190000022030353)
- golang 什么情况下会发生内存泄漏？Goroutinue泄露？
- [Memory Leaking Scenarios](https://go101.org/article/memory-leaking.html)
  - hanging goroutine
  - cgo
  - substring/slice
  - ticker
- golang sync.pool 临时对象池
- [golang 程序启动过程?](https://blog.iceinto.com/posts/go/start/) 
- 当go服务部署到线上了，发现有内存泄露，该怎么处理?

### 包和库（package)
- golang sql 链接池的实现
- golang http 连接池的实现?
- golang 与 kafka
- golang 与 mysql
- context
- json
- reflect
- http http库源码分析
- [Go Http包解析：为什么需要response.Body.Close()](https://segmentfault.com/a/1190000020086816)
- [为什么Response.Body需要被关闭](https://studygolang.com/articles/9887)
- [译]Go文件操作大全](https://colobu.com/2016/10/12/go-file-operations/)
- [Golang调度器GPM原理与调度全分析](https://zhuanlan.zhihu.com/p/323271088)
- [为什么要使用 Go 语言？Go 语言的优势在哪里？](https://www.zhihu.com/question/21409296/answer/1040884859)
- [Go内置数据结构原理](https://zhuanlan.zhihu.com/p/341945051)
- [从 bug 中学习：六大开源项目告诉你 go 并发编程的那些坑](https://zhuanlan.zhihu.com/p/352589023)
- [Go runtime剖析系列（一）：内存管理](https://zhuanlan.zhihu.com/p/323915446)
- [Go 内存泄露三宗罪](http://km.oa.com/group/19253/articles/show/460278?kmref=home_headline)
- [Go 与 C 的桥梁：cgo 入门，剖析与实践](https://www.zhihu.com/org/teng-xun-ji-zhu-gong-cheng),警惕cgo引入导致的性能问题比如线程数量过多，内存泄漏问题
- 定时器的设计和实现原理，golang的定时器是怎么实现的？
- [一次 go 服务大量连接 time_wait 问题排查](http://km.oa.com/group/35228/articles/show/461981?kmref=discovery).一般解决思路：TIME_WAIT排查是不是短链接，即频繁create and close socket CLOSE_WAIT排查自己代码BUG，socket没有close
- [为什么要使用 Go 语言？Go 语言的优势在哪里？](https://www.zhihu.com/question/21409296/answer/1040884859)



### 其它
- golang 单元测试，mock
- golang 性能分析？
- golang 的编译过程？



参考：
- https://go101.org/article/101.html
- https://colobu.com/
- http://legendtkl.com/about/
- https://draveness.me/
- https://github.com/uber-go/guide 《golang uber style》
- [Effective Go](http://https://golang.org/doc/effective_go.html)
- https://github.com/iswbm/golang-interview
- https://github.com/jincheng9/go-tutorial


golang 与闭包


## What's Go closure?
    In Go, a closure is a function that has access to variables from its outer (enclosing) function's scope. The closure "closes over" the variables, meaning that it retains access to them even after the outer function has returned. This makes closures a powerful tool for encapsulating data and functionality and for creating reusable code.


### Encapsulating State

```Go
package main

import "fmt"

func counter() func() int {
    i := 0
    return func() int {
        i++
        return i
    }
}

func main() {
    c := counter()

    fmt.Println(c()) // Output: 1
    fmt.Println(c()) // Output: 2
    fmt.Println(c()) // Output: 3
}

```

### Implementing Callbacks
```Go
package main

import "fmt"

func forEach(numbers []int, callback func(int)) {
    for _, n := range numbers {
        callback(n)
    }
}

func main() {
    numbers := []int{1, 2, 3, 4, 5}

    // Define a callback function to apply to each element of the numbers slice.
    callback := func(n int) {
        fmt.Println(n * 2)
    }

    // Use the forEach function to apply the callback function to each element of the numbers slice.
    forEach(numbers, callback)
}

```

### Fibonacci

```Go
package main

import "fmt"

func memoize(f func(int) int) func(int) int {
    cache := make(map[int]int)
    return func(n int) int {
        if val, ok := cache[n]; ok {
            return val
        }
        result := f(n)
        cache[n] = result
        return result
    }
}

func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
    fib := memoize(fibonacci)
    for i := 0; i < 10; i++ {
        fmt.Println(fib(i))
    }
}
```


### Factorial
```Go
package main

import "fmt"

func main() {
    factorial := func(n int) int {
        if n <= 1 {
            return 1
        }
        return n * factorial(n-1)
    }

    fmt.Println(factorial(5)) // Output: 120
}

```


### Event Handling
```Go
package main

import (
	"fmt"
	"time"
)

type Button struct {
	onClick func()
}

func NewButton() *Button {
	return &Button{}
}

func (b *Button) SetOnClick(f func()) {
	b.onClick = f
}

func (b *Button) Click() {
	if b.onClick != nil {
		b.onClick()
	}
}

func main() {
	button := NewButton()
	button.SetOnClick(func() {
		fmt.Println("Button Clicked!")
	})

	go func() {
		for {
			button.Click()
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Scanln()
}

```



---
title: golang哈希一致性算法实践
---


## 原理介绍
　　最近在项目中用到哈希一致性算法，它的需求是将入库的视频根据id均匀的分配到不同的容器中，当增加或者减少容器时，使得任务状态更改尽可能的少，于是想到了哈希一致性。
　　在做负载均衡时，简单的做法是将请求按照某个规则对服务器数量取模。取模的问题是当服务器数量增加或者减少时，会对原来的取模关系有非常大的影响。这在需要数据迁移或者更改服务状态的情况很难接受，hash一致性能在满足负载均衡的同时，尽可能少的更改服务状态或者数据迁移的工作量。
- 哈希环：用一个环表示0~2^32-1取值范围
- 节点映射： 根据节点标识信息计算出0~2^32-1的值，然后映射到哈希环上
- **虚拟节点**： 当节点数量很少时，映射关系较不确定，会导致节点在哈希环上分布不均匀，无法实现复杂均衡的效果，因此通常会引入虚拟节点。例如假设有3个节点对外提供服务，将3个节点映射到哈希环上很难保证分布均匀，如果将3个节点虚拟成1000个节点甚至更多节点，它们在哈希环上就会相对均匀。有些情况我们还会为每个节点设置权重例如node1、node2、node3的权重分别为1、2、3，假设虚拟节点总数为1200个，那么哈希环上将会有200个node1、400个node2、600个node3节点
- 将key值映射到节点： 以同样的映射关系将key映射到哈希环上，以顺时针的方式找到第一个值比key的哈希大的节点。
- **增加或者删除节点**：关于增加或者删除节点有多种不同的做法，常见的做法是剩余节点的权重值，重新安排虚拟的数量。例如上述的node1，node2和node3中，假设node3节点被下线，新的哈希环上会映射有有400个node1和800个node2。要注意的是原有的200个node1和400个node2会在相同的位置，但是会在之前的空闲区间增加了node1或者node2节点，因为权重的关系有些情况也会导致原有虚拟的节点的减少。
- **任务(数据更新)**：由于哈希环上节点映射更改，需要更新任务的状态。具体的做法是对每个任务映射状态进行检查，可以发现大多数任务的映射关系都保持不变，只有少量任务映射关系发生改变。总体来说就是**全状态检查，少量更改**。
![哈希一致性](https://github.com/wxquare/wxquare.github.io/raw/hexo/source/photos/hash_consistent.jpg)


## 实践
　　目前，Golang关于hash一致性有多种开源实现，因此实践起来也不是很难。这里参考https://github.com/g4zhuj/hashring, 根据自己的理解做了一些修改，并在项目中使用。

### 核心代码：hash_ring.go
```
package hashring

import (
	"crypto/sha1"
	"sync"
	"fmt"
	"math"
	"sort"
	"strconv"
)

/*
	https://github.com/g4zhuj/hashring
	https://segmentfault.com/a/1190000013533592
*/

const (
	//DefaultVirualSpots default virual spots
	DefaultTotalVirualSpots = 1000
)

type virtualNode struct {
	nodeKey   string
	nodeValue uint32
}
type nodesArray []virtualNode

func (p nodesArray) Len() int           { return len(p) }
func (p nodesArray) Less(i, j int) bool { return p[i].nodeValue < p[j].nodeValue }
func (p nodesArray) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p nodesArray) Sort()              { sort.Sort(p) }

//HashRing store nodes and weigths
type HashRing struct {
	total           int            //total number of virtual node
	virtualNodes    nodesArray     //array of virtual nodes sorted by value
	realNodeWeights map[string]int //Node:weight
	mu              sync.RWMutex
}

//NewHashRing create a hash ring with virual spots
func NewHashRing(total int) *HashRing {
	if total == 0 {
		total = DefaultTotalVirualSpots
	}

	h := &HashRing{
		total:           total,
		virtualNodes:    nodesArray{},
		realNodeWeights: make(map[string]int),
	}
	h.buildHashRing()
	return h
}

//AddNodes add nodes to hash ring
func (h *HashRing) AddNodes(nodeWeight map[string]int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for nodeKey, weight := range nodeWeight {
		h.realNodeWeights[nodeKey] = weight
	}
	h.buildHashRing()
}

//AddNode add node to hash ring
func (h *HashRing) AddNode(nodeKey string, weight int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.realNodeWeights[nodeKey] = weight
	h.buildHashRing()
}

//RemoveNode remove node
func (h *HashRing) RemoveNode(nodeKey string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.realNodeWeights, nodeKey)
	h.buildHashRing()
}

//UpdateNode update node with weight
func (h *HashRing) UpdateNode(nodeKey string, weight int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.realNodeWeights[nodeKey] = weight
	h.buildHashRing()
}

func (h *HashRing) buildHashRing() {
	var totalW int
	for _, w := range h.realNodeWeights {
		totalW += w
	}
	h.virtualNodes = nodesArray{}
	for nodeKey, w := range h.realNodeWeights {
		spots := int(math.Floor(float64(w) / float64(totalW) * float64(h.total)))
		for i := 1; i <= spots; i++ {
			hash := sha1.New()
			hash.Write([]byte(nodeKey + ":" + strconv.Itoa(i)))
			hashBytes := hash.Sum(nil)

			oneVirtualNode := virtualNode{
				nodeKey:   nodeKey,
				nodeValue: genValue(hashBytes[6:10]),
			}
			h.virtualNodes = append(h.virtualNodes, oneVirtualNode)

			hash.Reset()
		}
	}
	// sort virtual nodes for quick searching
	h.virtualNodes.Sort()
}

func genValue(bs []byte) uint32 {
	if len(bs) < 4 {
		return 0
	}
	v := (uint32(bs[3]) << 24) | (uint32(bs[2]) << 16) | (uint32(bs[1]) << 8) | (uint32(bs[0]))
	return v
}

//GetNode get node with key
func (h *HashRing) GetNode(s string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.virtualNodes) == 0 {
		fmt.Println("no valid node in the hashring")
		return ""
	}
	hash := sha1.New()
	hash.Write([]byte(s))
	hashBytes := hash.Sum(nil)
	v := genValue(hashBytes[6:10])
	i := sort.Search(len(h.virtualNodes), func(i int) bool { return h.virtualNodes[i].nodeValue >= v })
	//ring
	if i == len(h.virtualNodes) {
		i = 0
	}
	return h.virtualNodes[i].nodeKey
}


```

### 测试：hashring_test.go
```
package hashring

import (
	"fmt"
	"testing"
)

func TestHashRing(t *testing.T) {
	realNodeWeights := make(map[string]int)
	realNodeWeights["node1"] = 1
	realNodeWeights["node2"] = 2
	realNodeWeights["node3"] = 3

	totalVirualSpots := 100

	ring := NewHashRing(totalVirualSpots)
	ring.AddNodes(realNodeWeights)
	fmt.Println(ring.virtualNodes, len(ring.virtualNodes))
	fmt.Println(ring.GetNode("1845"))  //node3
	fmt.Println(ring.GetNode("994"))   //node1
	fmt.Println(ring.GetNode("hello")) //node3

	//remove node
	ring.RemoveNode("node3")
	fmt.Println(ring.GetNode("1845"))  //node2
	fmt.Println(ring.GetNode("994"))   //node1
	fmt.Println(ring.GetNode("hello")) //node2

	//add node
	ring.AddNode("node4", 2)
	fmt.Println(ring.GetNode("1845"))  //node4
	fmt.Println(ring.GetNode("994"))   //node1
	fmt.Println(ring.GetNode("hello")) //node4

	//update the weight of node
	ring.UpdateNode("node1", 3)
	fmt.Println(ring.GetNode("1845"))  //node4
	fmt.Println(ring.GetNode("994"))   //node1
	fmt.Println(ring.GetNode("hello")) //node1
	fmt.Println(ring.realNodeWeights)
}

```


```
    package main
    
    import (
    	"context"
    	"fmt"
    	"sync"
    	"time"
    )
    
    //sync package
    func sync1() {
    	var wg sync.WaitGroup
    	for i := 0; i < 10; i++ {
    		wg.Add(1) //设置协程等待的个数
    		go func(x int) {
    			defer func() {
    				wg.Done()
    			}()
    			fmt.Println("I'm", x)
    		}(i)
    	}
    	wg.Wait()
    }
    
    //chan
    func sync2() {
    	chanSync := make([]chan bool, 10)
    	for i := 0; i < 10; i++ {
    		chanSync[i] = make(chan bool)
    		go func(x int, ch chan bool) {
    			fmt.Println("I'm ", x)
    			ch <- true
    		}(i, chanSync[i])
    	}
    
    	for _, ch := range chanSync {
    		<-ch
    	}
    }
    
    //context
    func sync3() {
    	ctx, cancelFunc := context.WithCancel(context.Background())
    	defer cancelFunc()
    
    	for i := 0; i < 10; i++ {
    		go func(ctx context.Context, i int) {
    			for {
    				select {
    				case <-ctx.Done():
    					fmt.Println(ctx.Err(), i)
    					return
    				case <-time.After(2 * time.Second):
    					fmt.Println("time out", i)
    					return
    				}
    			}
    		}(ctx, i)
    	}
    	time.Sleep(5 * time.Second)
    }
    
    func main() {
    	sync1()
    	sync2()
    	sync3()
    	time.Sleep(10 * time.Second)
    }
```




最近在项目开发中使用http服务与第三方服务交互，感觉golang的http封装得很好，很方便使用但是也有一些坑需要注意，一是自动复用连接，二是Response.Body的读取和关闭

## http客户端自动复用连接
首先用代码直观的体验http客户端自动复用连接特点  
server.go

	func main() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "hello!")
		})
		http.ListenAndServe(":8848", nil)
	}

client.go

	func doReq() {
		resp, err := http.Get("http://127.0.0.1:8848/test")
		if err != nil {
			fmt.Println(err)
			return
		}
		io.Copy(os.Stdout, resp.Body)
		defer resp.Body.Close()
	}
	
	func main() {
		//http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 10
		for {
			go doReq()
			go doReq()
			//	go doReq()
			time.Sleep(300 * time.Millisecond)
		}
	}

测试1：执行`netstat | grep "8848" | wc -l`  结果：一直都是4  
测试2：增加一个go doReq(),继续测试，结果：是一直增大  
测试3：在测试2的基础上设置MaxIdleConnsPerHost = 10，结果：一直都是6

测试1已经能说明golang的http会自动复用连接  
测试2为什么连接数量会一直增加呢？原因是golang中默认只保持两条持久连接，http.Transport没有设置MaxIdleConnPerHost，于是便采用了默认的DefaultMaxIdleConnsPerHost，这个值是2。  
测试3通过加大MaxIdleConnPerHost的值，就能高效的利用http的自动复用机制。

## 读取和关闭Response.Body
将Resonse.Body的读取的代码屏蔽，继续测试。

    func doReq() {
    	resp, err := http.Get("http://127.0.0.1:8848/test")
    	if err != nil {
    		fmt.Println(err)
    		return
    	}
    	//io.Copy(os.Stdout, resp.Body)
    	defer resp.Body.Close()
    }  

测试结果发现，连接数一直增加。    
产生的原因：body实际上是一个嵌套了多层的net.TCPConn，当body没有被完全读取，也没有被关闭是，那么这次的http事物就没有完成，除非连接因为超时终止了，否则相关资源无法被回收。
从实现上看只要body被读完，连接就能被回收，只有需要抛弃body时才需要close，似乎不关闭也可以。但那些正常情况能读完的body，即第一种情况，在出现错误时就不会被读完，即转为第二种情况。而分情况处理则增加了维护者的心智负担，所以始终close body是最佳选择。


参考：  
[1].[https://my.oschina.net/hebaodan/blog/1609245](https://my.oschina.net/hebaodan/blog/1609245)  
[2].[https://www.jianshu.com/p/407fada3cc9d](https://www.jianshu.com/p/407fada3cc9d)  
[3].[https://serholiu.com/go-http-client-keepalive](https://serholiu.com/go-http-client-keepalive)




---
title: golang 基于task编程范式
categories:
- Golang
---




在复杂的业务逻辑中，一个逻辑或者说一个接口通常包含许多子逻辑，为了使得代码比较清晰，以前常使用责任链的设计模式来实现。最近发现有些场景下使用基于task的编程方式可以使得代码很清晰。这里简单记录下这种编程方式的特点以及实现方式。

责任链设计：https://refactoringguru.cn/design-patterns/chain-of-responsibility/go/example
  
主要特点：
1. 将一个复杂的任务分解成多个任务，任务之间支持串行和并行
2. Task执行的独立
3. Task执行的参数和中间数据保存Session中，使得数据能够灵活共享，需要保证协程安全
    
<img src=https://raw.githubusercontent.com/wxquare/wxquare.github.io/hexo/source/images/1ae6b34b-78d9-4e27-b7d8-cb3c89f13d7b.png width=400/> 

实现：
1. task.go定义 itask interface
2. 定义serialtask.go
3. 定义paralleltask.go
4. main主要业务逻辑定义session、任务编排、执行


```Go
package main

import (
	"context"
)

type iTask interface {
	Do(c context.Context) error
}
```


```GO
package main

import "context"

type SerialTask struct {
	tasks []iTask
}

func (s *SerialTask) Add(task iTask) {
	s.tasks = append(s.tasks, task)

}

func (S *SerialTask) Do(c context.Context) error {
	for _, t := range S.tasks {
		if err := t.Do(c); err != nil {
			return err
		}
	}
	return nil
}
```


```Go
package main

import (
	"context"
	"errors"
	"sync"
)

type ParallelTask struct {
	tasks []iTask
}

func (s *ParallelTask) Add(task iTask) {
	s.tasks = append(s.tasks, task)

}

func (S ParallelTask) Do(c context.Context) error {
	Errs := make(chan error, len(S.tasks))
	wg := sync.WaitGroup{}
	for _, t := range S.tasks {
		wg.Add(1)
		go func(i iTask) {
			defer wg.Done()
			if err := i.Do(c); err != nil {
				Errs <- err
			}
		}(t)
	}
	wg.Wait()
	if len(Errs) != 0 {
		return errors.New("parallel task error")
	}
	return nil
}
```

```Go
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Session struct {
	ctx    context.Context
	cancel context.CancelFunc

	Param   string
	Lock    *sync.Mutex
	Errs    chan error
	Timeout time.Duration
}

func (s Session) Deadline() (deadline time.Time, ok bool) {
	return s.ctx.Deadline()
}

func (s Session) Done() <-chan struct{} {
	return s.ctx.Done()
}

func (s Session) Err() error {
	return s.ctx.Err()
}

func (s Session) Value(key interface{}) interface{} {
	return s.ctx.Value(key)
}

type Task1 struct{}

func (t Task1) Do(ctx context.Context) error {
	session, ok := ctx.(*Session)
	if !ok {
		return errors.New("38 type assertion abort")
	}
	for {
		select {
		case <-session.ctx.Done():
			return session.ctx.Err()
		default:
			time.Sleep(time.Duration(rand.Int31n(5)) * time.Second)
			session.Lock.Lock()
			defer session.Lock.Unlock()
			session.Param = "task1"
			fmt.Printf("run task1 param%s\n", session.Param)
			return nil
		}
	}
}

type Task2 struct{}

func (t Task2) Do(ctx context.Context) error {
	session, ok := ctx.(*Session)
	if !ok {
		return errors.New("type assertion abort")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(time.Duration(rand.Int31n(5)) * time.Second)
			session.Lock.Lock()
			defer session.Lock.Unlock()
			session.Param = "task2"
			fmt.Printf("run task2 param%s\n", session.Param)
			return nil
		}
	}
}

type Task3 struct{}

func (t Task3) Do(ctx context.Context) error {
	session, ok := ctx.(*Session)
	if !ok {
		return errors.New("type assertion abort")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(time.Duration(rand.Int31n(5)) * time.Second)
			session.Lock.Lock()
			defer session.Lock.Unlock()
			session.Param = "task3"
			fmt.Printf("run task3 param%s\n", session.Param)
			return nil
		}
	}
}

type Task4 struct{}

func (t Task4) Do(ctx context.Context) error {
	session, ok := ctx.(*Session)
	if !ok {
		return errors.New("type assertion abort")
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(time.Duration(rand.Int31n(5)) * time.Second)
			session.Lock.Lock()
			defer session.Lock.Unlock()
			session.Param = "task4"
			fmt.Printf("run task4 param%s\n", session.Param)
			return nil
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	m := SerialTask{}
	m.Add(Task1{})
	p := ParallelTask{}
	p.Add(Task2{})
	p.Add(Task3{})
	m.Add(p)
	m.Add(Task4{})

	session := Session{
		Param:   "initial",
		Lock:    &sync.Mutex{},
		Timeout: 7 * time.Second,
	}
	session.ctx, session.cancel = context.WithTimeout(context.Background(), session.Timeout)
	defer session.cancel()

	if err := m.Do(&session); err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	fmt.Printf("%+v\n", session.Param)
}
```



---
title: golang sync.pool和连接池
categories:
- Golang
---


## 一、sync.Pool 基本使用
[https://golang.org/pkg/sync/](https://golang.org/pkg/sync/)  
sync.Pool的使用非常简单，它具有以下几个特点：
  
- sync.Pool设计目的是存放已经分配但暂时不用的对象，供以后使用，以减轻gc的代价，提高效率  
- 存储在Pool中的对象会随时被gc自动回收，Pool中对象的缓存期限为两次gc之间  
- 用户无法定义sync.Pool的大小，其大小仅仅受限于内存的大小     
- sync.Pool支持多协程之间共享
  
sync.Pool的使用非常简单，定义一个Pool对象池时，需要提供一个New函数，表示当池中没有对象时，如何生成对象。对象池Pool提供Get和Put函数从Pool中取和存放对象。

下面有一个简单的实例，直接运行是会打印两次“new an object”,注释掉runtime.GC(),发现只会调用一次New函数，表示实现了对象重用。
```
	package main
	
	import (
		"fmt"
		"runtime"
		"sync"
	)
	
	func main() {
		p := &sync.Pool{
			New: func() interface{} {
				fmt.Println("new an object")
				return 0
			},
		}
	
		a := p.Get().(int)
		a = 100
		p.Put(a)
		runtime.GC()
		b := p.Get().(int)
		fmt.Println(a, b)
	}
```
## 二、sync.Pool 如何支持多协程共享？
sync.Pool支持多协程共享，为了尽量减少竞争和加锁的操作，golang在设计的时候为每个P（核）都分配了一个子池，每个子池包含一个私有对象和共享列表。 私有对象只有对应的和核P能够访问，而共享列表是与其它P共享的。  

在golang的GMP调度模型中，我们知道协程G最终会被调度到某个固定的核P上。当一个协程在执行Pool的get或者put方法时，首先对改核P上的子池进行操作，然后对其它核的子池进行操作。因为一个P同一时间只能执行一个goroutine，所以对私有对象存取操作是不需要加锁的，而共享列表是和其他P分享的，因此需要加锁操作。  

一个协程希望从某个Pool中获取对象，它包含以下几个步骤：  
1. 判断协程所在的核P中的私有对象是否为空，如果非常则返回，并将改核P的私有对象置为空    
2. 如果协程所在的核P中的私有对象为空，就去改核P的共享列表中获取对象（需要加锁）  
3. 如果协程所在的核P中的共享列表为空，就去其它核的共享列表中获取对象（需要加锁）  
4. 如果所有的核的共享列表都为空，就会通过New函数产生一个新的对象  

在sync.Pool的源码中，每个核P的子池的结构如下所示：   
  
	// Local per-P Pool appendix.
	type poolLocalInternal struct {
		private interface{}   // Can be used only by the respective P.
		shared  []interface{} // Can be used by any P.
		Mutex                 // Protects shared.
	}
更加细致的sync.Pool源码分析，可参考[http://jack-nie.github.io/go/golang-sync-pool.html](http://jack-nie.github.io/go/golang-sync-pool.html)

## 三、为什么不使用sync.pool实现连接池？
刚开始接触到sync.pool时，很容易让人联想到连接池的概念，但是经过仔细分析后发现sync.pool并不是适合作为连接池，主要有以下两个原因： 
 
- 连接池的大小通常是固定且受限制的，而sync.Pool是无法控制缓存对象的数量，只受限于内存大小，不符合连接池的目标  
- sync.Pool对象缓存的期限在两次gc之间,这点也和连接池非常不符合

golang中连接池通常利用channel的缓存特性实现。当需要连接时，从channel中获取，如果池中没有连接时，将阻塞或者新建连接，新建连接的数量不能超过某个限制。

[https://github.com/goctx/generic-pool](https://github.com/goctx/generic-pool)基于channel提供了一个通用连接池的实现
```
	package pool
	
	import (
		"errors"
		"io"
		"sync"
		"time"
	)
	
	var (
		ErrInvalidConfig = errors.New("invalid pool config")
		ErrPoolClosed    = errors.New("pool closed")
	)
	
	type Poolable interface {
		io.Closer
		GetActiveTime() time.Time
	}
	
	type factory func() (Poolable, error)
	
	type Pool interface {
		Acquire() (Poolable, error) // 获取资源
		Release(Poolable) error     // 释放资源
		Close(Poolable) error       // 关闭资源
		Shutdown() error            // 关闭池
	}
	
	type GenericPool struct {
		sync.Mutex
		pool        chan Poolable
		maxOpen     int  // 池中最大资源数
		numOpen     int  // 当前池中资源数
		minOpen     int  // 池中最少资源数
		closed      bool // 池是否已关闭
		maxLifetime time.Duration
		factory     factory // 创建连接的方法
	}
	
	func NewGenericPool(minOpen, maxOpen int, maxLifetime time.Duration, factory factory) (*GenericPool, error) {
		if maxOpen <= 0 || minOpen > maxOpen {
			return nil, ErrInvalidConfig
		}
		p := &GenericPool{
			maxOpen:     maxOpen,
			minOpen:     minOpen,
			maxLifetime: maxLifetime,
			factory:     factory,
			pool:        make(chan Poolable, maxOpen),
		}
	
		for i := 0; i < minOpen; i++ {
			closer, err := factory()
			if err != nil {
				continue
			}
			p.numOpen++
			p.pool <- closer
		}
		return p, nil
	}
	
	func (p *GenericPool) Acquire() (Poolable, error) {
		if p.closed {
			return nil, ErrPoolClosed
		}
		for {
			closer, err := p.getOrCreate()
			if err != nil {
				return nil, err
			}
			// 如果设置了超时且当前连接的活跃时间+超时时间早于现在，则当前连接已过期
			if p.maxLifetime > 0 && closer.GetActiveTime().Add(time.Duration(p.maxLifetime)).Before(time.Now()) {
				p.Close(closer)
				continue
			}
			return closer, nil
		}
	}
	
	func (p *GenericPool) getOrCreate() (Poolable, error) {
		select {
		case closer := <-p.pool:
			return closer, nil
		default:
		}
		p.Lock()
		if p.numOpen >= p.maxOpen {
			closer := <-p.pool
			p.Unlock()
			return closer, nil
		}
		// 新建连接
		closer, err := p.factory()
		if err != nil {
			p.Unlock()
			return nil, err
		}
		p.numOpen++
		p.Unlock()
		return closer, nil
	}
	
	// 释放单个资源到连接池
	func (p *GenericPool) Release(closer Poolable) error {
		if p.closed {
			return ErrPoolClosed
		}
		p.Lock()
		p.pool <- closer
		p.Unlock()
		return nil
	}
	
	// 关闭单个资源
	func (p *GenericPool) Close(closer Poolable) error {
		p.Lock()
		closer.Close()
		p.numOpen--
		p.Unlock()
		return nil
	}
	
	// 关闭连接池，释放所有资源
	func (p *GenericPool) Shutdown() error {
		if p.closed {
			return ErrPoolClosed
		}
		p.Lock()
		close(p.pool)
		for closer := range p.pool {
			closer.Close()
			p.numOpen--
		}
		p.closed = true
		p.Unlock()
		return nil
	}
```
参考：  
[1].[https://blog.csdn.net/yongjian_lian/article/details/42058893](https://blog.csdn.net/yongjian_lian/article/details/42058893)  
[2].[https://segmentfault.com/a/1190000013089363](https://segmentfault.com/a/1190000013089363)  
[3].[http://jack-nie.github.io/go/golang-sync-pool.html](http://jack-nie.github.io/go/golang-sync-pool.html)


---
title: golang 指针和unsafe
categories:
- Golang
---

## 一、golang指针和unsafe.pointer
1. 不同类型的指针不能相互转化  
2. 指针变量不能进行运算，不支持c/c++中的++，--运算  
3. 任何类型的指针都可以被转换成unsafe.Pointer类型，反之也是  
4. uintptr值可以被转换成unsafe.Pointer类型，反之也是
5. 对unsafe.Pointer和uintptr两种类型单独解释两句：  
	- unsafe.Pointer是一个指针类型，指向的值不能被解析，类似于C/C++里面的(void *)，只说明这是一个指针，但是指向什么的不知道。
	- uintptr 是一个整数类型，这个整数的宽度足以用来存储一个指针类型数据；那既然是整数类类型，当然就可以对其进行运算了
```      
    package main
    import (
    	"fmt"
    	"unsafe"
    )
    func main() {
    	var ii [4]int = [4]int{11, 22, 33, 44}
    	px := &ii[0]
    	fmt.Println(&ii[0], px, *px)
    	//compile error
    	//pf32 := (*float32)(px)
    
    	//compile error
    	// px = px + 8
    	// px++
    
    	var pointer1 unsafe.Pointer = unsafe.Pointer(px)
    	var pf32 *float32 = (*float32)(pointer1)
    
    	var p2 uintptr = uintptr(pointer1)
    	print(p2)
    	p2 = p2 + 8
    	var pointer2 unsafe.Pointer = unsafe.Pointer(p2)
    	var pi32 *int = (*int)(pointer2)
    
    	fmt.Println(*px, *pf32, *pi32)
    
    }
```
## 二、 nil指针
引用类型声明而没有初始化赋值时，其值为nil。golang需要经常判断nil,防止出现panic错误。  
```
    bool  -> false  
    numbers -> 0 
    string-> ""  
    
    pointers -> nil
    slices -> nil
    maps -> nil
    channels -> nil
    functions -> nil
    interfaces -> nil



    package main
    
    import (
    	"fmt"
    )
    
    type Person struct {
    	AgeYears int
    	Name string
    	Friends  []Person
    }
    
    func main() {
    	var p Person
    	fmt.Printf("%v\n", p)
    
    	var slice1 []int
    	fmt.Println(slice1)
    	if slice1 == nil {
    		fmt.Println("slice1 is nil")
    	}
    	// fmt.Println(slice1[0])  panic
    
    	// var c chan int
    	// close(c)  panic
    }
```
参考：  

- https://studygolang.com/articles/10953  
- https://www.jianshu.com/p/dd80f6be7969  



---
title: Golang 基础知识汇总
categories:
- Golang
---


### Go 和 C++ 语言对比
Go and C++ are two different programming languages with different design goals, syntax, and feature sets. Here's a brief comparison of the two:

Syntax: Go has a simpler syntax than C++. It uses indentation for block structure and has fewer keywords and symbols. C++ has a more complex syntax with a lot of features that can make it harder to learn and use effectively.

Memory Management: C++ gives the programmer more control over memory management through its support for pointers, manual memory allocation, and deallocation. Go, on the other hand, uses a garbage collector to automatically manage memory, making it less error-prone.

Concurrency: Go has built-in support for concurrency through goroutines and channels, which make it easier to write concurrent code. C++ has a thread library that can be used to write concurrent code, but it requires more manual management of threads and locks.

Performance: C++ is often considered a high-performance language, and it can be used for system-level programming and performance-critical applications. Go is also fast but may not be as fast as C++ in some cases.

Libraries and Frameworks: C++ has a vast ecosystem of libraries and frameworks that can be used for a variety of applications, from game development to machine learning. Go's ecosystem is smaller, but it has good support for web development and distributed systems.

Overall, the choice of programming language depends on the project requirements, the available resources, and the developer's expertise. Both Go and C++ have their strengths and weaknesses, and the best choice depends on the specific needs of the project.


### string/[]byte
- string是golang的基本数组类型，s := "hello,world"，一旦初始化后不允许修改其内容
- [内部实现结构](https://draveness.me/golang/docs/part2-foundation/ch03-datastructure/golang-string/)，指向数据的指针data和表示长度的len
- 字符串拼接和格式化四种方式，+=，strings.join,buffer.writestring,fmt.sprintf
- [string 与 []byte的类型转换](https://www.cnblogs.com/shuiyuejiangnan/p/9707066.html)
- <font color=red>字符串与数值类型的不能强制转化，要使用strconv包中的parse和format函数</font>
- 标准库strings提供了许多字符串操作的函数,例如Split、HasPrefix,Trim。

### array 
- 数组array: [3]int{1,2,3}
- <font color=red>**数组是值类型**</font>，数组传参发生拷贝
- 定长
- 数组的创建、初始化、访问和遍历range，len(arr)求数组的长度
  
### slice
- 切片slice初始化: make([]int,len,cap)
- <font color=red>**slice是引用类型**</font>
- 变长，用容量和长度的区别，分别使用cap和len函数获取
- 内存结构和实现：指针、cap、size共24字节
- 常用函数，append，cap，len
- 切片动态扩容
- 深拷贝copy和浅拷贝“=”的区别

### map
- 引用类型，需要初始化 make(map[string]int,5) 
- [内部实现的数据结构，hmap、bmap等](https://draveness.me/golang/docs/part2-foundation/ch03-datastructure/golang-hashmap/)
- 链地址法解决冲突
- hashmap中buckets为什么为通常2的幂次方
- 访问流程，先用低位确定bucket，再用高8位粗选
- 增量扩容，迁移
- 使用map[interface{}]struct{}作为set
- [map遍历是无序且随机的，如何实现顺序遍历？](https://blog.csdn.net/slvher/article/details/44779081)
- [内部hashmap的实现原理](https://ninokop.github.io/2017/10/24/Go-Hashmap%E5%86%85%E5%AD%98%E5%B8%83%E5%B1%80%E5%92%8C%E5%AE%9E%E7%8E%B0/)。内部结构（bucket），扩容与迁移，删除。 

### sync.map
- 双map,read 和 dirty
- 以空间换效率，通过read和dirty两个map来提高读取效率
- 优先从read map中读取(无锁)，否则再从dirty map中读取(加锁)
- 动态调整，当misses次数过多时，将dirty map提升为read map
- 延迟删除，删除只是为value打一个标记，在dirty map提升时才执行真正的删除
- [sync.map 揭秘](https://colobu.com/2017/07/11/dive-into-sync-Map/)
- [sync.map读写流程图](https://segmentfault.com/a/1190000020946989)
- https://wudaijun.com/2018/02/go-sync-map-implement/


### struct
- 空结构体struct{}的用途，节省内存。
- 不支持继承，使用结构体嵌套组合
- struct 可以比较吗？普通struct可以比较，带引用的struc不可比较，需要使用reflect.DeepEqual
- struct没有slice和map类型时可直接判断
- slice和map本身不可比较，需要使用reflect.DeepEqual()。
- struct中包含slice和map等字段时，也要使用reflect.DeepEqual().
- https://stackoverflow.com/questions/24534072/how-to-compare-struct-slice-map-are-equal


### 类型和拷贝方式
- 值类型 ：String，Array，Int，Struct，Float，Bool，pointer（深拷贝）
- 引用类型：Slice，Map （浅拷贝）

### 函数和方法，匿名函数
- init函数
- 值接收和指针接收的区别
- 匿名函数？闭包？闭包延时绑定问题？用闭包写fibonacci数列？

### interface
- https://draveness.me/golang/docs/part2-foundation/ch04-basic/golang-interface/
- **隐式接口**，实现接口的所有方法就隐式地实现了接口；不需要显示申明实现某接口
- **接口也是 Go 语言中的一种类型**，它能够出现在变量的定义、函数的入参和返回值中并对它们进行约束，不过 Go 语言中有两种略微不同的接口，一种是带有一组方法的接口，另一种是不带任何方法的 interface{}：
- interface{} 类型不是任意类型，而是将类型转换成了 interface{} 类型
- 结构体实现接口 vs 结构体指针实现接口 区别？
- runtime.eface 和 runtime.iface 结构？
- **结构体类型转化为接口的类型相互变换，interface类型断言为struct类型** 过程
- **动态派发与多态**。动态派发（Dynamic dispatch）是在运行期间选择具体多态操作（方法或者函数）执行的过程，它是面向对象语言中的常见特性6。Go 语言虽然不是严格意义上的面向对象语言，但是接口的引入为它带来了动态派发这一特性，调用接口类型的方法时，如果编译期间不能确认接口的类型，Go 语言会在运行期间决定具体调用该方法的哪个实现。
- Golang没有泛型，通过interface可以实现简单泛型编程，例如的sort的实现
- 接口实现的源码

### channel
- Go鼓励CSP模型(communicating sequential processes),Goroutin之间通过channel传递数据
- 非缓冲的同步channel和带缓冲的异步channel
- [内部实现结构，带锁的循环队列runtime.hchan](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-channel/#642-%E6%95%B0%E6%8D%AE%E7%BB%93%E6%9E%84)
- channel创建make
- chan <- i
- **向channel发送数据**。在发送数据的逻辑执行之前会先为当前 Channel 加锁，防止多个线程并发修改数据。如果 Channel 已经关闭，那么向该 Channel 发送数据时会报 “send on closed channel” 错误并中止程序。分为的三个部分：
  当存在等待的接收者时，通过 runtime.send 直接将数据发送给阻塞的接收者；
  当缓冲区存在空余空间时，将发送的数据写入 Channel 的缓冲区；
  当不存在缓冲区或者缓冲区已满时，等待其他 Goroutine 从 Channel 接收数据；
- i <- ch，i, ok <- ch
- **从channel接收数据**的五种情况:
  - 如果 Channel 为空，那么会直接调用 runtime.gopark 挂起当前 Goroutine；
  - 如果 Channel 已经关闭并且缓冲区没有任何数据，runtime.chanrecv 会直接返回；
  - 如果 Channel 的 sendq 队列中存在挂起的 Goroutine，会将 recvx 索引所在的数据拷贝到接收变量所在的内存空间上并将 sendq 队列中 Goroutine 的数据拷贝到缓冲区；
  - 如果 Channel 的缓冲区中包含数据，那么直接读取 recvx 索引对应的数据；
  - 在默认情况下会挂起当前的 Goroutine，将 runtime.sudog 结构加入 recvq 队列并陷入休眠等待调度器的唤醒；
- **关闭channel**
- 如何优雅的关闭channel？https://www.jianshu.com/p/d24dfbb33781, channel关闭后读操作会发生什么？写操作会发生什么？

### 指针和unsafe.Pointer
- 相比C/C++，为了安全性考虑，Go指针弱化。不同类型的指针不能相互转化，指针变量不支持运算，不支持c/c++中的++，需要借助unsafe包
- 任何类型的指针都可以被转换成unsafe.Pointer类型，通过unsafe.Pointer实现不同类型指针的转化
- uintptr值可以被转换成unsafe.Pointer类型，通过uintptr实现指针的运算
- unsafe.Pointer是一个指针类型，指向的值不能被解析，类似于C/C++里面的(void *)，只说明这是一个指针，但是指向什么的不知道。
- uintptr 是一个整数类型，这个整数的宽度足以用来存储一个指针类型数据；那既然是整数类类型，当然就可以对其进行运算了
- nil
- [实践string和[]byte的高效转换](https://www.cnblogs.com/shuiyuejiangnan/p/9707066.html)
- 在业务场景中，使用指针虽然方便，但是要注意深拷贝和浅拷贝，这种错误还是比较常见的
- 当你对象是结构体对象的指针时，你想要获取字段属性时，可以直接使用'.'，而不需要解引用

### 集合set
1. golang中本身没有提供set，但可以通过map自己实现
2. 利用map键值不可重复的特性实现set，value为空结构体。 map[interface{}]struct{} 
3. [如何自己实现set？](https://studygolang.com/articles/11179)


### defer
- defer定义的延迟函数参数在defer语句出时就已经确定下来了
- defer定义顺序与实际执行顺序相反
- return不是原子操作，执行过程是: 保存返回值(若有)-->执行defer（若有）-->执行ret跳转
- 申请资源后立即使用defer关闭资源是好习惯
- golang中的defer用途？调用时机？调用顺序？预计算值？
- [defer 实现原理？](https://blog.csdn.net/Tybyqi/article/details/83827140)


### Go 错误处理 error、panic
- 在Go 语言中，错误被认为是一种可以预期的结果；而异常则是一种非预期的结果，发生异常可能表示程序中存在BUG 或发生了其它不可控的问题。 
- Go 语言推荐使用 recover 函数将内部异常转为错误处理，这使得用户可以真正的关心业务相关的错误处理。
- 在Go服务中通常需要自定义粗错误类型，最好能有效区分业务逻辑错误和系统错误，同时需要捕获panic，将panic转化为error，避免某个错误影响server重启
- panic 时需要保留runtime stack
```
  defer func() {
		if x := recover(); x != nil {
			panicReason := fmt.Sprintf("I'm panic because of: %v\n", x)
			logger.LogError(panicReason)
			stk := make([]byte, 10240)
			stkLen := runtime.Stack(stk, false)
			logger.LogErrorf("%s\n", string(stk[:stkLen]))
		}
	}()
 ```

## How does Go handle concurrency? (Goroutine,GMP调度模型，channel)
### what's CSP?
The **Communicating Sequential Processes (CSP) model** is a theoretical model of concurrent programming that was first introduced by Tony Hoare in 1978. The CSP model is based on the idea of concurrent processes that communicate with each other by sending and receiving messages through channels.The Go programming language provides support for the CSP model through its built-in concurrency features, such as goroutines and channels. In Go, concurrent processes are represented by goroutines, which are lightweight threads of execution. The communication between goroutines is achieved through channels, which provide a mechanism for passing values between goroutines in a safe and synchronized manner.

### Which is Goroutine ?
- Goroutines are lightweight, user-level threads of execution that run concurrently with other goroutines within the same process.
- Unlike traditional threads, goroutines are managed by the Go runtime, which automatically schedules and balances their execution across multiple CPUs and makes efficient use of available system resources.

### 比较Goroutine、thread、process
- 比较进程、线程和Goroutine。进程是资源分配的单位，有独立的地址空间，线程是操作系统调度的单位，协程是更细力度的执行单元，需要程序自身调度。Go语言原生支持Goroutine，并提供高效的协程调度模型。
- Goroutines, threads, and processes are all mechanisms for writing concurrent and parallel code, but they have some important differences:
- Goroutines: A goroutine is a lightweight, user-level thread of execution that runs concurrently with other goroutines within the same process. Goroutines are managed by the Go runtime, which automatically schedules and balances their execution across multiple CPUs. Goroutines require much less memory and have much lower overhead compared to threads, allowing for many goroutines to run simultaneously within a single process.
- Threads: A thread is a basic unit of execution within a process. Threads are independent units of execution that share the same address space as the process that created them. This allows threads to share data and communicate with each other, but also introduces the need for explicit synchronization to prevent race conditions and other synchronization issues.
- Processes: A process is a self-contained execution environment that runs in its own address space. Processes are independent of each other, meaning that they do not share memory or other resources. Communication between processes requires inter-process communication mechanisms, such as pipes, sockets, or message queues.
- In general, goroutines provide a more flexible and scalable approach to writing concurrent code compared to threads, as they are much lighter and more efficient, and allow for many more concurrent units of execution within a single process. Processes provide a more secure and isolated execution environment, but have higher overhead and require more explicit communication mechanisms.

### Why is Goroutine lighter and more efficient than thread or process?
- Stack size: Goroutines have a much smaller stack size compared to threads. The stack size of a goroutine is dynamically adjusted by the Go runtime, based on the needs of the goroutine. This allows for many more goroutines to exist simultaneously within a single process, as they require much less memory.
- Scheduling: Goroutines are scheduled by the Go runtime, which automatically balances and schedules their execution across multiple CPUs. This eliminates the need for explicit thread management and synchronization, reducing overhead.
- Context switching: Context switching is the process of saving and restoring the state of a running thread in order to switch to a different thread. Goroutines have a much lower overhead for context switching compared to threads, as they are much lighter and require less state to be saved and restored.
- Resource sharing: Goroutines share resources with each other and with the underlying process, eliminating the need for explicit resource allocation and deallocation. This reduces overhead and allows for more efficient use of system resources.
- Overall, the combination of a small stack size, efficient scheduling, low overhead context switching, and efficient resource sharing makes goroutines much lighter and more efficient than threads or processes, and allows for many more concurrent units of execution within a single process.
- Goroutine 上下文切换只涉及到三个寄存器（PC / SP / DX）的值修改；而对比线程的上下文切换则需要涉及模式切换（从用户态切换到内核态）、以及 16 个寄存器、PC、SP...等寄存器的刷新；内存占用少：线程栈空间通常是 2M，Goroutine 栈空间最小 2K；Golang 程序中可以轻松支持10w 级别的 Goroutine 运行，而线程数量达到 1k 时，内存占用就已经达到 2G。
- 理解G、P、M的含义以及调度模型

### How are goroutines scheduled by runtime?
- **Cooperative** (协作式). The scheduler uses a **cooperative** scheduling model, which means that goroutines voluntarily yield control to the runtime when they are blocked or waiting for an event. 
- **Timer-based preemption**. The scheduler uses a technique called **timer-based preemption** to interrupt the execution of a running goroutine and switch to another goroutine if it exceeds its time slice
- **Work-stealing**. The scheduler uses a work-stealing algorithm, where each CPU has its own local run queue, and goroutines are dynamically moved between run queues to balance the o balance the load and improve performance.
- **no explicit prioritization**. The Go runtime scheduler does not provide explicit support for prioritizing goroutines. Instead, it relies on the cooperative nature of goroutines to ensure that all goroutines make progress. In a well-designed Go program, the program should be designed such that all goroutines make progress in a fair and balanced manner.
- https://blog.csdn.net/sinat_34715587/article/details/124990458
- G 的数量可以远远大于 M 的数量，换句话说，Go 程序可以利用少量的内核级线程来支撑大量 Goroutine 的并发。多个 Goroutine 通过用户级别的上下文切换来共享内核线程 M 的计算资源，但对于操作系统来说并没有线程上下文切换产生的性能损耗，支持任务窃取（work-stealing）策略：为了提高 Go 并行处理能力，调高整体处理效率，当每个 P 之间的 G 任务不均衡时，调度器允许从 GRQ，或者其他 P 的 LRQ 中获取 G 执行。
- 减少因Goroutine创建大量M：
  -  由于原子、互斥量或通道操作调用导致 Goroutine 阻塞，调度器将把当前阻塞的 Goroutine 切换出去，重新调度 LRQ 上的其他 Goroutine；
  -  由于网络请求和 IO 操作导致 Goroutine 阻塞，通过使用 NetPoller 进行网络系统调用，调度器可以防止 Goroutine 在进行这些系统调用时阻塞 M。这可以让 M 执行 P 的 LRQ 中其他的 Goroutines，而不需要创建新的 M。有助于减少操作系统上的调度负载。
  -  当调用一些系统方法的时候，如果系统方法调用的时候发生阻塞，这种情况下，网络轮询器（NetPoller）无法使用，而进行系统调用的 Goroutine 将阻塞当前 M，则创建新的M。阻塞的系统调用完成后：M1 将被放在旁边以备将来重复使用
  -  如果在 Goroutine 去执行一个 sleep 操作，导致 M 被阻塞了。Go 程序后台有一个监控线程 sysmon，它监控那些长时间运行的 G 任务然后设置可以强占的标识符，别的 Goroutine 就可以抢先进来执行。

### What are the states of Goroutine and how do they flow?
- 协程的状态流转？Grunnable、Grunning、Gwaiting
- In Go, a Goroutine can be in one of several states during its lifetime. The states are:
- New: The Goroutine is created but has not started executing yet.
- Running: The Goroutine is executing on a machine-level thread.
- Waiting: The Goroutine is waiting for some external event, such as I/O, channel communication, or a timer.
- Sleeping: The Goroutine is sleeping, or waiting for a specified amount of time.
- Dead: The Goroutine has completed its execution and is no longer running.

In summary, the lifetime of a Goroutine in Go starts when it is created and ends when it completes its execution or encounters a panic, and can be influenced by synchronization mechanisms such as channels and wait groups.

### 生产者、消费者模型，并行计算累加求和
```Go
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func produce(mq chan<- int) {
	rand.Seed(time.Now().UnixNano())
	limitGoroutine := 2
	cnt := 100000
	var wg sync.WaitGroup
	for i := 0; i < limitGoroutine; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := start; j < cnt; j += limitGoroutine {
				num := rand.Intn(100)
				mq <- num
			}
		}(i)
	}
	go func() {
		wg.Wait()
		close(mq)
	}()
}

func consume(nums <-chan int) int {
	limitGoroutine := 4
	resChan := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < limitGoroutine; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var sum int = 0
			for num := range nums {
				sum += num
			}
			resChan <- sum
		}()
	}
	go func() {
		wg.Wait()
		close(resChan)
	}()
	var finalRes int = 0
	for r := range resChan {
		finalRes += r
	}
	return finalRes
}

func main() {
	mq := make(chan int, 10)
	go produce(mq)
	res := consume(mq)
	fmt.Printf("%+v\n", res)
}
```

-  golang context 用于在树形goroutine结构中，通过信号减少资源的消耗，包含Deadline、Done、Error、Value四个接口
-  常用的同步原语：channel、sync.mutex、sync.RWmutex、sync.WaitGroup、sync.Once、atomic
- 协程的状态流转？Grunnable、Grunning、Gwaiting
- sync.Mutex 和 sync.RWMutex 互斥锁和读写锁的使用场景？
- [golang 协程优雅的退出？](https://segmentfault.com/a/1190000017251049)
- 用channel实现定时器？（实际上是两个协程同步）
- 深入理解协程gmp调度模型，以及其发展历史
- 理解操作系统是怎么调度的，golang协程调度的优势，切换代价低，goroutine开销低，并发度高。
- Golang IO 模型和网络轮训器
- [sync.Mutex: “锁”实现背后那些事](http://km.oa.com/articles/show/502088)


## Golang 内存管理和垃圾回收（memory and gc）
### gc 的过程
- Marking phase: In this phase, the Go runtime identifies all objects that are accessible by the program and marks them as reachable. Objects that are not marked as reachable are considered unreachable and eligible for collection.
- Sweeping phase: In this phase, the Go runtime scans the memory heap and frees all objects that are marked as unreachable. The memory space occupied by these objects is now available for future allocation.
- Compacting phase: In this phase, the Go runtime rearranges the remaining objects on the heap to reduce fragmentation and minimize the impact of future allocations and deallocations.


### What are the memory leak scenarios in Go language?
- Goroutine leaks: If a goroutine is created and never terminated, it can result in a memory leak. This can occur when a program creates a goroutine to perform a task but fails to provide a mechanism for the goroutine to terminate, such as a channel to receive a signal to stop.
- Leaked closures: Closures are anonymous functions that capture variables from their surrounding scope. If a closure is created and assigned to a global variable, it can result in a memory leak, as the closure will continue to hold onto the captured variables even after they are no longer needed.
- Incorrect use of channels: Channels are a mechanism for communicating between goroutines. If a program creates a channel but never closes it, it can result in a memory leak. Additionally, if a program receives values from a channel but never discards them, they will accumulate in memory and result in a leak.
- Unclosed resources: In Go, it's important to close resources, such as files and network connections, when they are no longer needed. Failure to do so can result in a memory leak, as the resources and their associated memory will continue to be held by the program.
- Unreferenced objects: In Go, unreferenced objects are objects that are no longer being used by the program but still exist in memory. This can occur when an object is created and never explicitly deleted or when an object is assigned a new value and the old object is not properly disposed of.
By following best practices and being mindful of these common scenarios, you can help to avoid memory leaks in your Go programs. Additionally, you can use tools such as the Go runtime profiler to detect and diagnose memory leaks in your programs.


- **多级缓存**：内存分配器不仅会区别对待大小不同的对象，还会将内存分成不同的级别分别管理，TCMalloc 和 Go 运行时分配器都会引入线程缓存（Thread Cache）、中心缓存（Central Cache）和页堆（Page Heap）三个组件分级管理内存
- **对象大小**：Go 语言的内存分配器会根据申请分配的内存大小选择不同的处理逻辑，运行时根据对象的大小将对象分成微对象、小对象和大对象三种，tiny,small,large
- mspan、mcache、mcentral、mheap
- [深入理解golang GC的演进过程](https://segmentfault.com/a/1190000022030353)
- golang 什么情况下会发生内存泄漏？Goroutinue泄露？
- [Memory Leaking Scenarios](https://go101.org/article/memory-leaking.html)
  - hanging goroutine
  - cgo
  - substring/slice
  - ticker
- golang sync.pool 临时对象池
- [golang 程序启动过程?](https://blog.iceinto.com/posts/go/start/) 
- 当go服务部署到线上了，发现有内存泄露，该怎么处理?

### 包和库（package)
- golang sql 链接池的实现
- golang http 连接池的实现?
- golang 与 kafka
- golang 与 mysql
- context
- json
- reflect
- http http库源码分析
- [Go Http包解析：为什么需要response.Body.Close()](https://segmentfault.com/a/1190000020086816)
- [为什么Response.Body需要被关闭](https://studygolang.com/articles/9887)
- [译]Go文件操作大全](https://colobu.com/2016/10/12/go-file-operations/)
- [Golang调度器GPM原理与调度全分析](https://zhuanlan.zhihu.com/p/323271088)
- [为什么要使用 Go 语言？Go 语言的优势在哪里？](https://www.zhihu.com/question/21409296/answer/1040884859)
- [Go内置数据结构原理](https://zhuanlan.zhihu.com/p/341945051)
- [从 bug 中学习：六大开源项目告诉你 go 并发编程的那些坑](https://zhuanlan.zhihu.com/p/352589023)
- [Go runtime剖析系列（一）：内存管理](https://zhuanlan.zhihu.com/p/323915446)
- [Go 内存泄露三宗罪](http://km.oa.com/group/19253/articles/show/460278?kmref=home_headline)
- [Go 与 C 的桥梁：cgo 入门，剖析与实践](https://www.zhihu.com/org/teng-xun-ji-zhu-gong-cheng),警惕cgo引入导致的性能问题比如线程数量过多，内存泄漏问题
- 定时器的设计和实现原理，golang的定时器是怎么实现的？
- [一次 go 服务大量连接 time_wait 问题排查](http://km.oa.com/group/35228/articles/show/461981?kmref=discovery).一般解决思路：TIME_WAIT排查是不是短链接，即频繁create and close socket CLOSE_WAIT排查自己代码BUG，socket没有close
- [为什么要使用 Go 语言？Go 语言的优势在哪里？](https://www.zhihu.com/question/21409296/answer/1040884859)



### 其它
- golang 单元测试，mock
- golang 性能分析？
- golang 的编译过程？



参考：
- https://go101.org/article/101.html
- https://colobu.com/
- http://legendtkl.com/about/
- https://draveness.me/
- https://github.com/uber-go/guide 《golang uber style》
- [Effective Go](http://https://golang.org/doc/effective_go.html)
- https://github.com/iswbm/golang-interview
- https://github.com/jincheng9/go-tutorial


---
title: golang程序启动与init函数
categories:
- Golang
---

　　在golang中，可执行文件的入口函数并不是我们写的main函数，编译器在编译go代码时会插入一段起引导作用的汇编代码，它引导程序进行命令行参数、运行时的初始化，例如内存分配器初始化、垃圾回收器初始化、协程调度器的初始化。golang引导初始化之后就会进入用户逻辑，因为存在特殊的init函数，main函数也不是程序最开始执行的函数。

## 一、golang程序启动流程
　　golang可执行程序由于运行时runtime的存在，其启动过程还是非常复杂的，这里通过gdb调试工具简单查看其启动流程：  
1. 找一个golang编译的可执行程序test，info file查看其入口地址：gdb test，info files
(gdb) info files
Symbols from "/home/terse/code/go/src/learn_golang/test_init/main".
Local exec file:
	/home/terse/code/go/src/learn_golang/test_init/main', 
    file type elf64-x86-64.
	**Entry point: 0x452110**
	.....
2. 利用断点信息找到目标文件信息：
(gdb) b *0x452110
Breakpoint 1 at 0x452110: file /usr/local/go/src/runtime/rt0_linux_amd64.s, line 8.
3. 依次找到对应的文件对应的行数，设置断点，调到指定的行，查看具体的内容：
(gdb) b _rt0_amd64  
(gdb) b b runtime.rt0_go  
至此，由汇编代码针对特定平台实现的引导过程就全部完成了，后续的代码都是用Go实现的。分别实现命令行参数初始化，内存分配器初始化、垃圾回收器初始化、协程调度器的初始化等功能。
```
	CALL	runtime·args(SB)
	CALL	runtime·osinit(SB)
	CALL	runtime·schedinit(SB)

	CALL	runtime·newproc(SB)

	CALL	runtime·mstart(SB)
```

## 二、特殊的init函数
1. init函数先于main函数自动执行，不能被其他函数调用
2. init函数没有输入参数、没有返回值
3. 每个包可以含有多个同名的init函数，每个源文件也可以有多个同名的init函数
4. **执行顺序** 变量初始化 > init函数 > main函数。在复杂项目中，初始化的顺序如下：
	- 先初始化import包的变量，然后先初始化import的包中的init函数，，再初始化main包变量，最后执行main包的init函数
	- 从上到下初始化导入的包（执行init函数），遇到依赖关系，先初始化没有依赖的包
	- 从上到下初始化导入包中的变量，遇到依赖，先执行没有依赖的变量初始化
	- main包本身变量的初始化，main包本身的init函数
	- 同一个包中不同源文件的初始化是按照源文件名称的字典序

util.go
```
package util

import (
	"fmt"
)

var c int = func() int {
	fmt.Println("util variable init")
	return 3
}()

func init() {
	fmt.Println("call util.init")
}
```

main.go
```
package main

import (
	"fmt"
	_ "util"
)

var a int = func() int {
	fmt.Println("main variable init")
	return 3
}()

func init() {
	fmt.Println("call main.init")
}

func main() {
	fmt.Println("call main.main")
}
```
执行结果：  
　　　util variable init
　　　call util.init
　　　main variable init
　　　call main.init
　　　call main.main


参考：《Go语言学习笔记13、14、15章》



---
title: golang 程序性能分析与优化
categories:
- Golang
---


　　最近参加了dave关于高性能golang的论坛，它通过几个case非常清晰的介绍了golang性能分析与优化的技术，非常值得学习。[https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html](https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html)。随着计算机硬件资源越来越受限制，关注程序的性能不仅能提高服务的性能也能降低成本。

## 一、性能测量
### 1、**time** 
　　在Linux中，time命令经常用于统计程序的**耗时(real)、用户态cpu耗时(user)、系统态cpu耗时(sys)**。在操作系统中程序的运行包括用户态和系统态。由于程序有时处于等待状态，在单核程序中，总是real>user+sys的，将(user+sys)/real称为cpu利用率。对于多核程序来说，由于能把多个cpu都利用起来，上面的关系就不成立。

### 2、**benchmarking**
　　有时我们有测试某些函数性能的需求，go testing包内置了非常好用的benchmarks。例如有一个产生斐波那契数列的函数，可以用testing包测试出它的benchmark。
fib.go:
```
	package benchmarkFib

	func Fib(n int) int {
		switch n {
		case 0:
			return 0
		case 1:
			return 1
		case 2:
			return 2
		default:
			return Fib(n-1) + Fib(n-2)
		}
	}
```
fib_test.go
```
	package benchmarkFib

	import (
		"testing"
	)

	func BenchmarkFib20(b *testing.B) {
		for n := 0; n < b.N; n++ {
			Fib(20)
		}
	}
```
运行$go test -bench=.
	goos: linux
	goarch: amd64
	pkg: learn_golang/benchmarkFib
	BenchmarkFib20-4   	  100000	     22912 ns/op
	PASS
	ok  	learn_golang/benchmarkFib	2.526s


### 3、profile
　　benchmark能帮助分析某些函数的性能，但是对于分析整个程序来说还是需要使用profile。golang使用profile是非常方便的，因为很早期的时候就集成到runtime中,它包括两个部分：
- runtime/pprof
- go tool pprof cpu.pprof 分析profile数据

pprof包括四种类型的profile，其中最常用的是cpu profile和memory profile。
- **CPU profile**：最常用，运行时每10ms中断并记录当前运行的goroutine的堆栈跟踪，通过cpu profile可以看出函数调用的次数和所占用时间的百分比。
- **Memory profile**：采样的是分配的堆内存而不是使用的内存
- Block profile
- Mutex contention profile

**收集profile**
为了更方便的产生profile文件，dave封装了runtime/pprof。https://github.com/pkg/profile.git
结合dave的例子分析cpu profile：https://github.com/wxquare/learn_golang/tree/master/pprof
	% go run main.go moby.txt
	2019/05/06 21:26:56 profile: cpu profiling enabled, cpu.pprof
	"moby.txt": 181275 words
	2019/05/06 21:26:57 profile: cpu profiling disabled, cpu.pprof

**分析profile**
a、使用命令分析profile：
	% go tool pprof
	% top 
b、借助浏览器分析profile： go tool pprof -http=:8080
	图模式（Graph mode)
	火焰图模式(Flame Graph mode)
 

## 二、Execution Tracer
   profile是基于采样(sample)的，而Execution Tracer是集成到Go运行时(runtime)中，因此它能知道程序在某个时间点的具体行为。Dave用了一个例子来说明为什么需要tracer，而 go tool pprof执行的效果很差。

1. v1 time ./mandelbrot (原版)
    real    0m1.654s
	user    0m1.630s
	sys     0m0.015s

2. 跑出profile、分析profile
	cd examples/mandelbrot-runtime-pprof
	go run mandelbrot.go > cpu.pprof
    go tool pprof -http=:8080 cpu.pprof

3. 通过profile数据，可以知道fillpixel几乎做了程序所有的工作，但是我们似乎也没有什么可以优化的了？？？这个时候可以考虑引入Execution tracer。运行程序跑出trace数据。
	import "github.com/pkg/profile"

	func main() {
		defer profile.Start(profile.TraceProfile, profile.ProfilePath(".")).Stop()

	然后使用go tool trace trace.out 分析trace数据。

4. 分析trace数据，记住要使用chrome浏览器。
  通过trace数据可以看出只有一个Goroutine在工作，没有利用好机器的资源。

5. 之后的几个优化通过调整使用的gorutine的数量使得程序充分利用CPU计算资源，提高程序的效率。


## 三、编译器优化
1. 逃逸分析（Escape analysis）
	golang在内存分配的时候没有堆(heap)和栈(stack)的区别，由编译器决定是否需要将对象逃逸到堆中。例如：
```
		func Sum() int {
		const count = 100
		numbers := make([]int, count)
		for i := range numbers {
			numbers[i] = i + 1
		}

		var sum int
		for _, i := range numbers {
			sum += i
		}
		return sum
	}

	func main() {
		answer := Sum()
		fmt.Println(answer)
	}
```
$ go build -gcflags=-m test_esc.go 
command-line-arguments
./test_esc.go:9:17: Sum make([]int, count) does not escape
./test_esc.go:23:13: answer escapes to heap
./test_esc.go:23:13: main ... argument does not escape

2. 内敛（Inlining）
   了解C/C++的应该知道内敛，golang编译器同样支持函数内敛，对于较短且重复调用的函数可以考虑使用内敛

3. Dead code elimination/Branch elimination
	编译器会将代码中一些无用的分支进行优化，分支判断，提高效率。例如下面一段代码由于a和b是常量，编译器也可以推导出Max(a,b)，因此最终F函数为空
```	
	func Max(a, b int) int {
		if a > b {
			return a
		}
		return b
	}

	func F() {
		const a, b = 100, 20
		if Max(a, b) == b {
			panic(b)
		}
	}
```
常用的编译器选项： go build -gcflags="-lN" xxx.go
- "-S",编译时查看汇编代码
- "-l",关闭内敛优化
- "-m",打印编译优化的细节
- "-l -N",关闭所有的优化



## 四、内存和垃圾回收
golang支持垃圾回收，gc能减少编程的负担，但与此同时也可能造成程序的性能问题。那么如何测量golang程序使用的内存，以及如何减少golang gc的负担呢？经历了许多版本的迭代，golang gc 沿着低延迟和高吞吐的目标在进化，相比早起版本，目前有了很大的改善，但仍然有可能是程序的瓶颈。因此要学会分析golang 程序的内存和垃圾回收问题。

如何查看程序的gc信息？
1. 通过设置环境变量？env GODEBUG=gctrace=1
例如： env GODEBUG=gctrace=1 godoc -http=:8080
2. import _ "net/http/pprof"，查看/debug/pprof

tips：
1. 减少内存分配，优先使用第二种APIs
	func (r *Reader) Read() ([]byte, error)
	func (r *Reader) Read(buf []byte) (int, error)
2. 尽量避免string 和 []byte之间的转换
3. 尽量减少两个字符串的合并
4. 对slice预先分配大小
5. 尽量不要使用cgo，因为c和go毕竟是两种语言。cgo是个high overhead的操作，调用cgo相当于阻塞IO，消耗一个线程
6. defer is expensive？在性能要求较高的时候，考虑少用
7. 对IO操作设置超时机制是个好习惯SetDeadline, SetReadDeadline, SetWriteDeadline
8. 当数据量很大的时候，考虑使用流式IO(streaming IO)。io.ReaderFrom / io.WriterTo



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

<div align='center'>
<img src="https://github.com/wxquare/wxquare.github.io/raw/hexo/source/images/runtime.png" width="500" height="500">
</div >

## 程序bootstrap过程
如上图所示，Go程序启动大致分为一下一个部分：
- 参数处理，runtime·args(SB)
- 操作系统初始化，runtime·osinit(SB)
- 调度器初始化，runtime·schedinit(SB)
- 运行runtime.main函数，装载用户main函数并运行，runtime.main()
参数处理和osinit逻辑比较简单，代码也较少，这里主要记录下调度器初始化和runtime.main函数两个部分

### runtime·schedinit
  schedinit内容比较多，主要包含：
  - 栈初始化 stackinit() 
  - 堆初始化 mallocinit()
  - gc初始化 gcinit()
  - 初始化resize allp []*p procresize()

#### stack

stackinit() 核心代码用于初始化全局的stackpool和stackLarge两个结构
```GO
var stackpool [_NumStackOrders]struct {
	item stackpoolItem
	_    [cpu.CacheLinePadSize - unsafe.Sizeof(stackpoolItem{})%cpu.CacheLinePadSize]byte
}

//go:notinheap
type stackpoolItem struct {
	mu   mutex
	span mSpanList
}

// Global pool of large stack spans.
var stackLarge struct {
	lock mutex
	free [heapAddrBits - pageShift]mSpanList // free lists by log_2(s.npages)
}

func stackinit() {
	if _StackCacheSize&_PageMask != 0 {
		throw("cache size must be a multiple of page size")
	}
	for i := range stackpool {
		stackpool[i].item.span.init()
		lockInit(&stackpool[i].item.mu, lockRankStackpool)
	}
	for i := range stackLarge.free {
		stackLarge.free[i].init()
		lockInit(&stackLarge.lock, lockRankStackLarge)
	}
}

```

### newproc 需要一个初始的stack
```Go
	if gp.stack.lo == 0 {
		// Stack was deallocated in gfput or just above. Allocate a new one.
		systemstack(func() {
			gp.stack = stackalloc(startingStackSize)
		})
		gp.stackguard0 = gp.stack.lo + _StackGuard
```

goroutine 运行时需要把stack 地址传给m


### 



### runtime.main






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

<div align='center'>
<img src="https://github.com/wxquare/wxquare.github.io/raw/hexo/source/images/runtime.png" width="500" height="500">
</div >

## 程序bootstrap过程
如上图所示，Go程序启动大致分为一下一个部分：
- 参数处理，runtime·args(SB)
- 操作系统初始化，runtime·osinit(SB)
- 调度器初始化，runtime·schedinit(SB)
- 运行runtime.main函数，装载用户main函数并运行，runtime.main()
参数处理和osinit逻辑比较简单，代码也较少，这里主要记录下调度器初始化和runtime.main函数两个部分

### runtime·schedinit
  schedinit内容比较多，主要包含：
  - 栈初始化 stackinit() 
  - 堆初始化 mallocinit()
  - gc初始化 gcinit()
  - 初始化resize allp []*p procresize()

#### stack

stackinit() 核心代码用于初始化全局的stackpool和stackLarge两个结构
```GO
var stackpool [_NumStackOrders]struct {
	item stackpoolItem
	_    [cpu.CacheLinePadSize - unsafe.Sizeof(stackpoolItem{})%cpu.CacheLinePadSize]byte
}

//go:notinheap
type stackpoolItem struct {
	mu   mutex
	span mSpanList
}

// Global pool of large stack spans.
var stackLarge struct {
	lock mutex
	free [heapAddrBits - pageShift]mSpanList // free lists by log_2(s.npages)
}

func stackinit() {
	if _StackCacheSize&_PageMask != 0 {
		throw("cache size must be a multiple of page size")
	}
	for i := range stackpool {
		stackpool[i].item.span.init()
		lockInit(&stackpool[i].item.mu, lockRankStackpool)
	}
	for i := range stackLarge.free {
		stackLarge.free[i].init()
		lockInit(&stackLarge.lock, lockRankStackLarge)
	}
}

```

### newproc 需要一个初始的stack
```Go
	if gp.stack.lo == 0 {
		// Stack was deallocated in gfput or just above. Allocate a new one.
		systemstack(func() {
			gp.stack = stackalloc(startingStackSize)
		})
		gp.stackguard0 = gp.stack.lo + _StackGuard
```

goroutine 运行时需要把stack 地址传给m


### 



### runtime.main






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



---
title: golang 程序测试和优化
categories:
- Golang
---

　　Golang非常注重工程化，提供了非常好用单元测试、性能测试（benchmark）和调优工具（pprof），它们对提高代码的质量和服务的性能非常有帮助。[参考链接](https://tonybai.com/2015/08/25/go-debugging-profiling-optimization)中通过一段http代码非常详细的介绍了golang程序优化的步骤和方便之处。实际工作中，我们很难每次都对代码都有那么高的要求，但是能使用一些工具对程序进行优化程序性能也是golang程序员必备的技能。
- testing 标准库 
- go test 测试工具
- go tool pprof 分析profile数据


## 一、单元测试，测试正确性
1. 为了测试某个文件中的某个函数的性能，在相同目录下定义xxx_test.go文件，使用go build命令编译程序时会忽略测试文件
2. 在测试文件中定义测试某函数的代码，以TestXxxx方式命名，例如TestAdd
3. 在相同目录下运行 go test -v 即可观察代码的测试结果

    	func TestAdd(t *testing.T) {
    		if add(1, 3) != 4 {
    			t.FailNow()
    		}
    	}

	

## 二、性能测试，benchmark
1. 单元测试，测试程序的正确性。benchmark 用户测试代码的效率，执行的时间
2. benchmark测试以BenchMark开头，例如BenchmarkAdd
3. 运行 go test -v -bench=. 程序会运行到一定的测试，直到有比较准备的测试结果

    	func BenchmarkAdd(b *testing.B) {
    		for i := 0; i < b.N; i++ {
    			_ = add(1, 2)
    		}
    	}
    
    	BenchmarkAdd-4  	2000000000	 0.26 ns/op

## 三、pprof性能分析

1. 除了使用使用testing进行单元测试和benchanmark性能测试，golang能非常方便捕获或者监控程序运行状态数据，它包括cpu、内存、和阻塞等，并且非常的直观和易于分析。
2. 有两种捕获方式： a、在测试时输出并保存相关数据；b、在运行阶段，在线采集，通过web接口获得实时数据。
3. Benchamark时输出profile数据：go test -v -bench=. -memprofile=mem.out -cpuprofile=cpu.out
4. 使用go tool pprof xxx.test mem.out 进行交互式查看，例如top5。同理，可以分析其它profile文件。  

(pprof) top5
Showing nodes accounting for 1994.93MB, 63.62% of 3135.71MB total
Dropped 28 nodes (cum <= 15.68MB)
Showing top 5 nodes out of 46
      flat  flat%   sum%        cum   cum%
  475.10MB 15.15% 15.15%   475.10MB 15.15%  regexp/syntax.(*compiler).inst
  455.58MB 14.53% 29.68%   455.58MB 14.53%  regexp.progMachine
  421.55MB 13.44% 43.12%   421.55MB 13.44%  regexp/syntax.(*parser).newRegexp
  328.61MB 10.48% 53.60%   328.61MB 10.48%  regexp.onePassCopy
  314.09MB 10.02% 63.62%   314.09MB 10.02%  net/http/httptest.cloneHeader

- flat：仅当前函数，不包括它调用的其它函数
- cum： 当前函数调用堆栈的累计
- sum： 列表前几行所占百分比的总和

更加详细的golang程序调试和优化请参考：
[1]. https://tonybai.com/2015/08/25/go-debugging-profiling-optimization/
[2]. https://blog.golang.org/profiling-go-programs



---
title: golang channel通道
categories:
- Golang
---

## 一、channel
  channel是golang中的csp并发模型非常重要组成部分，使用起来非常像阻塞队列。
- 通道channel变量本身就是指针，可用“==”操作符判断是否为同一对象
- 未初始化的channel为nil，需要使用make初始化
- 理解初始化的channel和nil channel的区别？读写nil channel都会阻塞，关闭nil channel会出现panic；可以读关闭的channel，写关闭的channel会发出panic，close关闭了的channel会发出panic
- 同步模式的channel必须有配对操作的goroutine出现，否则会一直阻塞，而异步模式在缓冲区未满或者数据未读完前，不会阻塞。
- 内置的cap和len函数返回channel缓冲区大小和当前已缓冲的数量，而对于同步通道则返回0
- 除了使用"<-"发送和接收操作符外，还可以用ok-idom或者range模式处理chanel中的数据。
- 重复关闭和关闭nil channel都会导致pannic
- make可以创建单项通道，但那没有意义，通产使用类型转换来获取单向通道，并分别赋予给操作方
- 无法将单向通道转换成双向通道



## 二、基本用法
1. 协程之间传递数据
2. 用作事件通知，经常使用空结构体channel作为某个事件通知
3. select帮助同时多个通道channel，它会随机选择一个可用的通道做收发操作
4. 使用异步channel（带有缓冲）实现信号量semaphore
5. 标准库提供了timeout和tick的channel实现。
6. 通道并非用来取代锁的，通道和锁有各自不同的使用场景，通道倾向于解决逻辑层次的并发处理架构，而锁则用来保护数据的安全性。
7. channel队列本质上还是使用锁同步机制，单次获取更多的数据（批处理），减少收发的次数，可改善因为频繁加锁造成的性能问题。
8. channel可能会导致goroutine leak问题，是指goroutine处于发送或者接收阻塞状态，但一直未被唤醒，垃圾回收器并不收集此类资源，造成资源的泄露。    
```
		    func main() {
		    	done := make(chan struct{})
		    	s := make(chan int)
		    	go func() {
		    		s <- 1
		    		close(done)
		    	}()
		    	fmt.Println(<-s)
		    	<-done
		    }
		
		    func main() {
		    	sem := make(chan struct{}, 2) //two groutine
		    	var wg sync.WaitGroup
		    	for i := 0; i < 10; i++ {
		    		wg.Add(1)
		    		go func(id int) {
		    			defer wg.Done()
		    			defer func() { <-sem }()
		    			sem <- struct{}{}
		    			time.Sleep(1 * time.Second)
		    			fmt.Println("id=", id)
		    		}(i)
		    	}
		    	wg.Wait()
		    }
		
		
		    func main() {
		    	go func() {
		    		tick := time.Tick(1 * time.Second)
		    		for {
		    			select {
		    			case <-time.After(5 * time.Second):
		    				fmt.Println("time out")
		    			case <-tick:
		    				fmt.Println("time tick 1s")
		    			default:
		    				fmt.Println("default")
		    			}
		    		}
		    	}()
		    	<-(chan struct{})(nil)
		    }

```
参考：《Go语言学习笔记第8章》



---
title: golang http 使用总结
categories:
- Golang
---

最近在项目开发中使用http服务与第三方服务交互，感觉golang的http封装得很好，很方便使用但是也有一些坑需要注意，一是自动复用连接，二是Response.Body的读取和关闭

## http客户端自动复用连接
首先用代码直观的体验http客户端自动复用连接特点  
server.go

	func main() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "hello!")
		})
		http.ListenAndServe(":8848", nil)
	}

client.go

	func doReq() {
		resp, err := http.Get("http://127.0.0.1:8848/test")
		if err != nil {
			fmt.Println(err)
			return
		}
		io.Copy(os.Stdout, resp.Body)
		defer resp.Body.Close()
	}
	
	func main() {
		//http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 10
		for {
			go doReq()
			go doReq()
			//	go doReq()
			time.Sleep(300 * time.Millisecond)
		}
	}

测试1：执行`netstat | grep "8848" | wc -l`  结果：一直都是4  
测试2：增加一个go doReq(),继续测试，结果：是一直增大  
测试3：在测试2的基础上设置MaxIdleConnsPerHost = 10，结果：一直都是6

测试1已经能说明golang的http会自动复用连接  
测试2为什么连接数量会一直增加呢？原因是golang中默认只保持两条持久连接，http.Transport没有设置MaxIdleConnPerHost，于是便采用了默认的DefaultMaxIdleConnsPerHost，这个值是2。  
测试3通过加大MaxIdleConnPerHost的值，就能高效的利用http的自动复用机制。

## 读取和关闭Response.Body
将Resonse.Body的读取的代码屏蔽，继续测试。

    func doReq() {
    	resp, err := http.Get("http://127.0.0.1:8848/test")
    	if err != nil {
    		fmt.Println(err)
    		return
    	}
    	//io.Copy(os.Stdout, resp.Body)
    	defer resp.Body.Close()
    }  

测试结果发现，连接数一直增加。    
产生的原因：body实际上是一个嵌套了多层的net.TCPConn，当body没有被完全读取，也没有被关闭是，那么这次的http事物就没有完成，除非连接因为超时终止了，否则相关资源无法被回收。
从实现上看只要body被读完，连接就能被回收，只有需要抛弃body时才需要close，似乎不关闭也可以。但那些正常情况能读完的body，即第一种情况，在出现错误时就不会被读完，即转为第二种情况。而分情况处理则增加了维护者的心智负担，所以始终close body是最佳选择。


参考：  
[1].[https://my.oschina.net/hebaodan/blog/1609245](https://my.oschina.net/hebaodan/blog/1609245)  
[2].[https://www.jianshu.com/p/407fada3cc9d](https://www.jianshu.com/p/407fada3cc9d)  
[3].[https://serholiu.com/go-http-client-keepalive](https://serholiu.com/go-http-client-keepalive)


---
title: golang 协程同步的三种方式
categories:
- Golang
---
```
    package main
    
    import (
    	"context"
    	"fmt"
    	"sync"
    	"time"
    )
    
    //sync package
    func sync1() {
    	var wg sync.WaitGroup
    	for i := 0; i < 10; i++ {
    		wg.Add(1) //设置协程等待的个数
    		go func(x int) {
    			defer func() {
    				wg.Done()
    			}()
    			fmt.Println("I'm", x)
    		}(i)
    	}
    	wg.Wait()
    }
    
    //chan
    func sync2() {
    	chanSync := make([]chan bool, 10)
    	for i := 0; i < 10; i++ {
    		chanSync[i] = make(chan bool)
    		go func(x int, ch chan bool) {
    			fmt.Println("I'm ", x)
    			ch <- true
    		}(i, chanSync[i])
    	}
    
    	for _, ch := range chanSync {
    		<-ch
    	}
    }
    
    //context
    func sync3() {
    	ctx, cancelFunc := context.WithCancel(context.Background())
    	defer cancelFunc()
    
    	for i := 0; i < 10; i++ {
    		go func(ctx context.Context, i int) {
    			for {
    				select {
    				case <-ctx.Done():
    					fmt.Println(ctx.Err(), i)
    					return
    				case <-time.After(2 * time.Second):
    					fmt.Println("time out", i)
    					return
    				}
    			}
    		}(ctx, i)
    	}
    	time.Sleep(5 * time.Second)
    }
    
    func main() {
    	sync1()
    	sync2()
    	sync3()
    	time.Sleep(10 * time.Second)
    }
```



---
title: golang哈希一致性算法实践
categories:
- Golang
---


## 原理介绍
　　最近在项目中用到哈希一致性算法，它的需求是将入库的视频根据id均匀的分配到不同的容器中，当增加或者减少容器时，使得任务状态更改尽可能的少，于是想到了哈希一致性。
　　在做负载均衡时，简单的做法是将请求按照某个规则对服务器数量取模。取模的问题是当服务器数量增加或者减少时，会对原来的取模关系有非常大的影响。这在需要数据迁移或者更改服务状态的情况很难接受，hash一致性能在满足负载均衡的同时，尽可能少的更改服务状态或者数据迁移的工作量。
- 哈希环：用一个环表示0~2^32-1取值范围
- 节点映射： 根据节点标识信息计算出0~2^32-1的值，然后映射到哈希环上
- **虚拟节点**： 当节点数量很少时，映射关系较不确定，会导致节点在哈希环上分布不均匀，无法实现复杂均衡的效果，因此通常会引入虚拟节点。例如假设有3个节点对外提供服务，将3个节点映射到哈希环上很难保证分布均匀，如果将3个节点虚拟成1000个节点甚至更多节点，它们在哈希环上就会相对均匀。有些情况我们还会为每个节点设置权重例如node1、node2、node3的权重分别为1、2、3，假设虚拟节点总数为1200个，那么哈希环上将会有200个node1、400个node2、600个node3节点
- 将key值映射到节点： 以同样的映射关系将key映射到哈希环上，以顺时针的方式找到第一个值比key的哈希大的节点。
- **增加或者删除节点**：关于增加或者删除节点有多种不同的做法，常见的做法是剩余节点的权重值，重新安排虚拟的数量。例如上述的node1，node2和node3中，假设node3节点被下线，新的哈希环上会映射有有400个node1和800个node2。要注意的是原有的200个node1和400个node2会在相同的位置，但是会在之前的空闲区间增加了node1或者node2节点，因为权重的关系有些情况也会导致原有虚拟的节点的减少。
- **任务(数据更新)**：由于哈希环上节点映射更改，需要更新任务的状态。具体的做法是对每个任务映射状态进行检查，可以发现大多数任务的映射关系都保持不变，只有少量任务映射关系发生改变。总体来说就是**全状态检查，少量更改**。
![哈希一致性](https://github.com/wxquare/wxquare.github.io/raw/hexo/source/photos/hash_consistent.jpg)


## 实践
　　目前，Golang关于hash一致性有多种开源实现，因此实践起来也不是很难。这里参考https://github.com/g4zhuj/hashring, 根据自己的理解做了一些修改，并在项目中使用。

### 核心代码：hash_ring.go
```
package hashring

import (
	"crypto/sha1"
	"sync"
	"fmt"
	"math"
	"sort"
	"strconv"
)

/*
	https://github.com/g4zhuj/hashring
	https://segmentfault.com/a/1190000013533592
*/

const (
	//DefaultVirualSpots default virual spots
	DefaultTotalVirualSpots = 1000
)

type virtualNode struct {
	nodeKey   string
	nodeValue uint32
}
type nodesArray []virtualNode

func (p nodesArray) Len() int           { return len(p) }
func (p nodesArray) Less(i, j int) bool { return p[i].nodeValue < p[j].nodeValue }
func (p nodesArray) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p nodesArray) Sort()              { sort.Sort(p) }

//HashRing store nodes and weigths
type HashRing struct {
	total           int            //total number of virtual node
	virtualNodes    nodesArray     //array of virtual nodes sorted by value
	realNodeWeights map[string]int //Node:weight
	mu              sync.RWMutex
}

//NewHashRing create a hash ring with virual spots
func NewHashRing(total int) *HashRing {
	if total == 0 {
		total = DefaultTotalVirualSpots
	}

	h := &HashRing{
		total:           total,
		virtualNodes:    nodesArray{},
		realNodeWeights: make(map[string]int),
	}
	h.buildHashRing()
	return h
}

//AddNodes add nodes to hash ring
func (h *HashRing) AddNodes(nodeWeight map[string]int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for nodeKey, weight := range nodeWeight {
		h.realNodeWeights[nodeKey] = weight
	}
	h.buildHashRing()
}

//AddNode add node to hash ring
func (h *HashRing) AddNode(nodeKey string, weight int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.realNodeWeights[nodeKey] = weight
	h.buildHashRing()
}

//RemoveNode remove node
func (h *HashRing) RemoveNode(nodeKey string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.realNodeWeights, nodeKey)
	h.buildHashRing()
}

//UpdateNode update node with weight
func (h *HashRing) UpdateNode(nodeKey string, weight int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.realNodeWeights[nodeKey] = weight
	h.buildHashRing()
}

func (h *HashRing) buildHashRing() {
	var totalW int
	for _, w := range h.realNodeWeights {
		totalW += w
	}
	h.virtualNodes = nodesArray{}
	for nodeKey, w := range h.realNodeWeights {
		spots := int(math.Floor(float64(w) / float64(totalW) * float64(h.total)))
		for i := 1; i <= spots; i++ {
			hash := sha1.New()
			hash.Write([]byte(nodeKey + ":" + strconv.Itoa(i)))
			hashBytes := hash.Sum(nil)

			oneVirtualNode := virtualNode{
				nodeKey:   nodeKey,
				nodeValue: genValue(hashBytes[6:10]),
			}
			h.virtualNodes = append(h.virtualNodes, oneVirtualNode)

			hash.Reset()
		}
	}
	// sort virtual nodes for quick searching
	h.virtualNodes.Sort()
}

func genValue(bs []byte) uint32 {
	if len(bs) < 4 {
		return 0
	}
	v := (uint32(bs[3]) << 24) | (uint32(bs[2]) << 16) | (uint32(bs[1]) << 8) | (uint32(bs[0]))
	return v
}

//GetNode get node with key
func (h *HashRing) GetNode(s string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.virtualNodes) == 0 {
		fmt.Println("no valid node in the hashring")
		return ""
	}
	hash := sha1.New()
	hash.Write([]byte(s))
	hashBytes := hash.Sum(nil)
	v := genValue(hashBytes[6:10])
	i := sort.Search(len(h.virtualNodes), func(i int) bool { return h.virtualNodes[i].nodeValue >= v })
	//ring
	if i == len(h.virtualNodes) {
		i = 0
	}
	return h.virtualNodes[i].nodeKey
}


```

### 测试：hashring_test.go
```
package hashring

import (
	"fmt"
	"testing"
)

func TestHashRing(t *testing.T) {
	realNodeWeights := make(map[string]int)
	realNodeWeights["node1"] = 1
	realNodeWeights["node2"] = 2
	realNodeWeights["node3"] = 3

	totalVirualSpots := 100

	ring := NewHashRing(totalVirualSpots)
	ring.AddNodes(realNodeWeights)
	fmt.Println(ring.virtualNodes, len(ring.virtualNodes))
	fmt.Println(ring.GetNode("1845"))  //node3
	fmt.Println(ring.GetNode("994"))   //node1
	fmt.Println(ring.GetNode("hello")) //node3

	//remove node
	ring.RemoveNode("node3")
	fmt.Println(ring.GetNode("1845"))  //node2
	fmt.Println(ring.GetNode("994"))   //node1
	fmt.Println(ring.GetNode("hello")) //node2

	//add node
	ring.AddNode("node4", 2)
	fmt.Println(ring.GetNode("1845"))  //node4
	fmt.Println(ring.GetNode("994"))   //node1
	fmt.Println(ring.GetNode("hello")) //node4

	//update the weight of node
	ring.UpdateNode("node1", 3)
	fmt.Println(ring.GetNode("1845"))  //node4
	fmt.Println(ring.GetNode("994"))   //node1
	fmt.Println(ring.GetNode("hello")) //node1
	fmt.Println(ring.realNodeWeights)
}

```


---
title: Go Closure 使用场景介绍
categories:
- Golang
---

## What's Go closure?
    In Go, a closure is a function that has access to variables from its outer (enclosing) function's scope. The closure "closes over" the variables, meaning that it retains access to them even after the outer function has returned. This makes closures a powerful tool for encapsulating data and functionality and for creating reusable code.


### Encapsulating State

```Go
package main

import "fmt"

func counter() func() int {
    i := 0
    return func() int {
        i++
        return i
    }
}

func main() {
    c := counter()

    fmt.Println(c()) // Output: 1
    fmt.Println(c()) // Output: 2
    fmt.Println(c()) // Output: 3
}

```

### Implementing Callbacks
```Go
package main

import "fmt"

func forEach(numbers []int, callback func(int)) {
    for _, n := range numbers {
        callback(n)
    }
}

func main() {
    numbers := []int{1, 2, 3, 4, 5}

    // Define a callback function to apply to each element of the numbers slice.
    callback := func(n int) {
        fmt.Println(n * 2)
    }

    // Use the forEach function to apply the callback function to each element of the numbers slice.
    forEach(numbers, callback)
}

```

### Fibonacci

```Go
package main

import "fmt"

func memoize(f func(int) int) func(int) int {
    cache := make(map[int]int)
    return func(n int) int {
        if val, ok := cache[n]; ok {
            return val
        }
        result := f(n)
        cache[n] = result
        return result
    }
}

func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
    fib := memoize(fibonacci)
    for i := 0; i < 10; i++ {
        fmt.Println(fib(i))
    }
}
```


### Factorial
```Go
package main

import "fmt"

func main() {
    factorial := func(n int) int {
        if n <= 1 {
            return 1
        }
        return n * factorial(n-1)
    }

    fmt.Println(factorial(5)) // Output: 120
}

```


### Event Handling
```Go
package main

import (
	"fmt"
	"time"
)

type Button struct {
	onClick func()
}

func NewButton() *Button {
	return &Button{}
}

func (b *Button) SetOnClick(f func()) {
	b.onClick = f
}

func (b *Button) Click() {
	if b.onClick != nil {
		b.onClick()
	}
}

func main() {
	button := NewButton()
	button.SetOnClick(func() {
		fmt.Println("Button Clicked!")
	})

	go func() {
		for {
			button.Click()
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Scanln()
}

```
