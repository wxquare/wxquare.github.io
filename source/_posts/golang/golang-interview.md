---
title: Golang 基础知识汇总
categories:
- Golang
---

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


### 如何golang处理程序中的error、panic
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

### Golang并发编程 (concurrent programming)
- 比较进程、线程和Goroutine。进程是资源分配的单位，有独立的地址空间，线程是操作系统调度的单位，协程是更细力度的执行单元，需要程序自身调度。Go语言原生支持Goroutine，并提供高效的协程调度模型。
- 参考：[为什么要使用 Go 语言？Go 语言的优势在哪里？](https://www.zhihu.com/question/21409296/answer/1040884859)
- Goroutine 上下文切换只涉及到三个寄存器（PC / SP / DX）的值修改；而对比线程的上下文切换则需要涉及模式切换（从用户态切换到内核态）、以及 16 个寄存器、PC、SP...等寄存器的刷新；内存占用少：线程栈空间通常是 2M，Goroutine 栈空间最小 2K；Golang 程序中可以轻松支持10w 级别的 Goroutine 运行，而线程数量达到 1k 时，内存占用就已经达到 2G。
- 理解G、P、M的含义以及调度模型
- G 的数量可以远远大于 M 的数量，换句话说，Go 程序可以利用少量的内核级线程来支撑大量 Goroutine 的并发。多个 Goroutine 通过用户级别的上下文切换来共享内核线程 M 的计算资源，但对于操作系统来说并没有线程上下文切换产生的性能损耗
- 支持任务窃取（work-stealing）策略：为了提高 Go 并行处理能力，调高整体处理效率，当每个 P 之间的 G 任务不均衡时，调度器允许从 GRQ，或者其他 P 的 LRQ 中获取 G 执行。
- 减少因Goroutine创建大量M：
  -  由于原子、互斥量或通道操作调用导致 Goroutine 阻塞，调度器将把当前阻塞的 Goroutine 切换出去，重新调度 LRQ 上的其他 Goroutine；
  -  由于网络请求和 IO 操作导致 Goroutine 阻塞，通过使用 NetPoller 进行网络系统调用，调度器可以防止 Goroutine 在进行这些系统调用时阻塞 M。这可以让 M 执行 P 的 LRQ 中其他的 Goroutines，而不需要创建新的 M。有助于减少操作系统上的调度负载。
  -  当调用一些系统方法的时候，如果系统方法调用的时候发生阻塞，这种情况下，网络轮询器（NetPoller）无法使用，而进行系统调用的 Goroutine 将阻塞当前 M，则创建新的M。阻塞的系统调用完成后：M1 将被放在旁边以备将来重复使用
  -  如果在 Goroutine 去执行一个 sleep 操作，导致 M 被阻塞了。Go 程序后台有一个监控线程 sysmon，它监控那些长时间运行的 G 任务然后设置可以强占的标识符，别的 Goroutine 就可以抢先进来执行。
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


### Golang 内存管理和垃圾回收（memory and gc）
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
