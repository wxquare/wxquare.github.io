---
title: golang 基础知识汇总
categories:
- Golang
---

## golang 常用数据结构以及内部实现

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
- 拷贝copy和“=”的区别

### map
- 引用类型，需要初始化 make(map[string]int,5) 
- [内部实现的数据结构，hmap、bmap等](https://draveness.me/golang/docs/part2-foundation/ch03-datastructure/golang-hashmap/)
- 链地址法解决冲突
- hashmap中buckets为什么为通常2的幂次方
- 访问流程，先用低位确定bucket，再用高8位粗选
- 增量扩容，迁移
- 使用map[interface{}]struct{}作为set
- [如何实现顺序遍历？](https://blog.csdn.net/slvher/article/details/44779081)
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
- 
- channel的实现原理
- 如何优雅的关闭channel？https://www.jianshu.com/p/d24dfbb33781, channel关闭后读操作会发生什么？写操作会发生什么？

### 指针和unsafe.Pointer
- 原生指针
- unsafe.Pointer
- 实践string和[]byte的高效转换

### 集合set
1. golang中本身没有提供set，但可以通过map自己实现
2. 利用map键值不可重复的特性实现set，value为空结构体。 map[interface{}]struct{} 
3. [如何自己实现set？](https://studygolang.com/articles/11179)

  
### 容器container/heap、list、ring
1. heap与优先队列，最小堆
2. 链表list，双向列表
3. 循环队列ring
4. <font color=red>golang没有提供stack，可自己实现</font>
5. <font color=red>golang没有提供queue，但可以通过channel替换或者自己实现</font>


### 延伸问题：
#### 1.如何比较struct/slice/map?
- struct没有slice和map类型时可直接判断
- slice和map本身不可比较，需要使用reflect.DeepEqual()。
- truct中包含slice和map等字段时，也要使用reflect.DeepEqual().
- [https://stackoverflow.com/questions/24534072/how-to-compare-struct-slice-map-are-equal](https://stackoverflow.com/questions/24534072/how-to-compare-struct-slice-map-are-equal)
- [https://studygolang.com/articles/11342](https://studygolang.com/articles/11342)



## golang 关键字
### defer
- golang中的defer用途？调用时机？调用顺序？预计算值？
### select
- 用途和实现
### range
### make/new
- make和new的区别
### panic/recover
### nil



## golang并发编程 (concurrent programming)
- channel、sync.mutex,sync.RWmutext,sync.WaitGroup,sync.Once,atomic 原子操作
- goroutine的实现以及其调度模型
- golang中的G-P-M调度模型？协程的状态?gwaiting和Gsyscall?抢占式调度?
- 协程的状态流转？Grunnable、Grunning、Gwaiting
- golang怎么做Goroutine之间的同步？channel、sync.mutex、sync.WaitGroup、context，锁怎么实现，用了什么cpu指令?
- [goroutine交替执行,使其能顺序输出1-20的自然数code](https://github.com/wxquare/programming/blob/master/golang/learn_golang/goroutine_example1.go)
- [生产者消费者模式code](https://github.com/wxquare/programming/blob/master/golang/learn_golang/producer_consumer.go)
- sync.Mutex 和 sync.RWMutex 互斥锁和读写锁的使用场景？
- golang context 包的用途？
- [golang 协程优雅的退出？](https://segmentfault.com/a/1190000017251049)
- golang 为什么高并发好？讲了go的调度模型
- sync.Mutex 和 sync.RWMutex 互斥锁和读写锁的使用场景？
- 怎么做协程同步
- 主协程如何等其余协程完再操作
- 并发调度
- 用channel实现定时器？（实际上是两个协程同步）
- 深入理解协程gmp调度模型，以及其发展历史
- 理解操作系统是怎么调度的，golang协程调度的优势


## golang 内存管理和垃圾回收（memory and gc）
- golang中的三级内存管理？对比C++中的内存管理？
- [golang GC](https://segmentfault.com/a/1190000022030353)
- golang 什么情况下会发生内存泄漏？Goroutinue泄露？
- golang sync.pool 临时对象池
- [golang 程序启动过程?](https://blog.iceinto.com/posts/go/start/) 
- golang 内存模型与C++的比较?
- golang IO 模型和网络轮训器


## 包和库（package)
- golang sql 链接池的实现
- golang http 连接池的实现?
- golang 与 kafka
- golang 与 mysql
- context
- json
- reflect
- http http库源码分析
- [Go Http包解析：为什么需要response.Body.Close()](https://segmentfault.com/a/1190000020086816)
- [译]Go文件操作大全](https://colobu.com/2016/10/12/go-file-operations/)


## 其它
- golang 单元测试，mock
- golang 性能分析？
- golang 的编译过程？
- 当go服务部署到线上了，发现有内存泄露，该怎么处理?
- 微服务架构中名字服务，服务注册，服务发现，复杂均衡，心跳，路由等
- golang 单例模式，mutext，sync.once


## 一、解释并发与并行
1. 并行： 物理上的并行，程序能否利用多核物理设备同一时刻执行多个任务，并行依赖多核的支持
2. 并发： 逻辑上的并发，程序在同一时刻执行过个任务，并发不需要多核的支持，在单核处理器上能以间隔方式切换不同的任务


## 二、解释进程，线程和协程的区别。协程有哪些优势？
1. 进程
2. 线程
3. 协程

参考：
- https://go101.org/article/101.html
- https://colobu.com/
- http://legendtkl.com/about/
- https://draveness.me/
- https://github.com/uber-go/guide 《golang uber style》
- [Effective Go](http://https://golang.org/doc/effective_go.html)
