---
title: golang 基础知识汇总
categories:
- Golang
---

### golang基础
- golang中的new和make区别？
- golang中的defer？调用时机？调用顺序？预计算值？
- golang中的错误处理方式？error，nil，panic，recover？
- golang select 的用途？
- nil


### 类型系统以及反射相关
- string
- array
- slice的实现
- map的实现
- sync.map的实现
- interface接口
- channel通道的实现原理，
- function函数
- 怎么实现set？
- 空结构体
- 指针类型和unsafe包的使用
- struct 可以比较吗？引用类型不可比较？reflect.DeepEqual的比较？
- golang 中reflect的理解？reflect.DeepEqual()?如何结构体反射取出所有的成员？


### 函数和方法
- main和init函数
- 值接收者和方法接收者
- 匿名函数？闭包？闭包延时绑定问题？用闭包写fibonacci数列？



### 接口interface
- 理解隐式接口的含义
- 有方法的接口和空接口在实现时是不同的结构iface和eface
- 注意使用指针接受者实现接口和使用值接收者实现接口方法的不同
- 空接口类型不是任意类型，而是类型变换
- 接口与类型的互相转换
- 接口类型断言
- 动态派发与多态
- golang没有泛型，通过interface可以实现简单泛型编程，例如的sort的实现
- 接口实现的源码


### 通道channel
- channel的实现原理
- 如何优雅的关闭channel？https://www.jianshu.com/p/d24dfbb33781, channel关闭后读操作会发生什么？写操作会发生什么？


### golang类型系统和反射（type system)
- 如何优雅的关闭channel？https://www.jianshu.com/p/d24dfbb33781, channel关闭后读操作会发生什么？写操作会发生什么？
- golang 中reflect的理解？reflect.DeepEqual()?如何结构体反射取出所有的成员？
- 实现一个hashmap，解决hash冲突的方法，解决hash倾斜的方法
- c++的模板跟go的interface的区别
- 怎么理解go的interface
- unsafe包学习，与指针
- golang bytes



### golang并发编程 (concurrent programming)
- golang中的G-P-M调度模型？
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
- 并发调度
- 用channel实现定时器？（实际上是两个协程同步）


### golang 内存和垃圾回收（memory and gc）
- golang中的三级内存管理？对比C++中的内存管理？
- 堆、栈和逃逸分析？
- 三色标记垃圾回收？
- golang 什么情况下会发生内存泄漏？Goroutinue泄露？
- golang sync.pool 临时对象池


### golang runtime以及内部实现
- [golang中的runtime包教程](golang中的runtime包教程)
- golang slice 的实现？
- golang map的实现，哈希算法，如何解决hash冲突，如何扩容？
- golang sync.map的实现。
- golang channel的实现
- golang interface的设计和实现，接口类型转换、类型断言以及动态派发机制


### 包和库（package)
- golang sql 链接池的实现
- golang http 连接池的实现
- golang 与 kafka
- golang 与 mysql
- [译]Go文件操作大全](https://colobu.com/2016/10/12/go-file-operations/)


### 其它相关
- golang 单元测试，mock
- golang 性能分析？
- golang 的编译过程？
- golang runtime 了解多少？
- [[]byte和string的相互转换和unsafe？](https://go101.org/article/unsafe.html)

参考：
- https://go101.org/article/101.html
- https://colobu.com/
- http://legendtkl.com/about/
- https://draveness.me/
- https://github.com/uber-go/guide 《golang uber style》
- [Effective Go](http://https://golang.org/doc/effective_go.html)