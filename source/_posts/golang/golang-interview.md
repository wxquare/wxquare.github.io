---
title: golang 基础知识汇总
categories:
- Golang
---

## golang 常用数据结构以及内部实现
### string/[]byte
- string 内容不可变，只可读
- 字符串拼接的四种方式，+=，strings.join,buffer.writestring,fmt.sprintf
- string 与 []byte的类型转换
- [[]byte和string的相互转换和unsafe？](https://go101.org/article/unsafe.html)
### array 
### slice
- 内部实现
- make，len，cap
- 扩容
### map
- 内部实现的结构
- 链地址法解决冲突
- hashmap中buckets为什么为2的幂次方
- 怎么做的增量扩容
- map按照key顺序输出
- 使用map[interface{}]struct{}
- https://segmentfault.com/a/1190000018632347
### sync.map
- 双map,read 和 dirty
- lock
- https://colobu.com/2017/07/11/dive-into-sync-Map/
- https://segmentfault.com/a/1190000020946989
- https://wudaijun.com/2018/02/go-sync-map-implement/
- load,store,delete 的流程
### channel
- 内部实现，带锁的循环队列
- 非缓冲，可缓冲 
- channel的实现原理
- 如何优雅的关闭channel？https://www.jianshu.com/p/d24dfbb33781, channel关闭后读操作会发生什么？写操作会发生什么？
### interface
- 空接口的实现
- 带函数的interface的实现
- 理解隐式接口的含义
- 有方法的接口和空接口在实现时是不同的结构iface和eface
- 注意使用指针接受者实现接口和使用值接收者实现接口方法的不同
- 空接口类型不是任意类型，而是类型变换
- 接口与类型的互相转换
- 接口类型断言
- 动态派发与多态
- golang没有泛型，通过interface可以实现简单泛型编程，例如的sort的实现
- 接口实现的源码
- 接口类型转换、类型断言以及动态派发机制
### struct
- 空结构体struct{}
- 结构体嵌套
- struct 可以比较吗？普通struct可以比较，带引用的struc不可比较
- reflect.DeepEqual
### 函数和方法，匿名函数
- init函数
- 值接收和指针接收的区别
- 匿名函数？闭包？闭包延时绑定问题？用闭包写fibonacci数列？
### 指针和unsafe.Pointer
- 原生指针
- unsafe.Pointer



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
- goroutine
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


## golang 内存和垃圾回收（memory and gc）
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


参考：
- https://go101.org/article/101.html
- https://colobu.com/
- http://legendtkl.com/about/
- https://draveness.me/
- https://github.com/uber-go/guide 《golang uber style》
- [Effective Go](http://https://golang.org/doc/effective_go.html)