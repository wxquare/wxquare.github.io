---
title: golang 面试题汇总
categories:
- Golang
---

### golang基础
- golang中的new和make区别？
- golang中的defer？调用时机？调用顺序？预计算值？
- golang中的main函数和init函数？
- golang中的匿名函数？闭包？闭包延时绑定问题？用闭包写fibonacci数列？
- golang中的错误处理方式？error，nil，panic，recover？
- golang select 的用途？

### golang类型系统和反射（type system)
- 如何优雅的关闭channel？https://www.jianshu.com/p/d24dfbb33781, channel关闭后读操作会发生什么？写操作会发生什么？
- golang 中reflect的理解？reflect.DeepEqual()?如何结构体反射取出所有的成员？
- golang 接口和接口对象断言
- golang map的实现，图解，扩容，哈希冲突？非协程安全？map加sync.Mutex的方案？sync.map减少锁带来的影响，sync.map 实现原理，拓扑关系图？
- sync.Mutex 和 sync.RWMutex 互斥锁和读写锁的使用场景？
- golang struct 可以比较吗？引用类型不可比较？reflect.DeepEqual的比较？
- golang 中的空结构体？
- golang 总的set？
- golang 中的指针和unsafe包？golang指针退化，不支持类型转换和运算，需要使用借助unsafe
- 用channel实现定时器？（实际上是两个协程同步）
- golang channel的内部实现？
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

### golang 内存和垃圾回收（memory and gc）
- golang中的三级内存管理？对比C++中的内存管理？
- 堆、栈和逃逸分析？
- 三色标记垃圾回收？
- golang 什么情况下会发生内存泄漏？Goroutinue泄露？
- golang sync.pool 临时对象池

### golang 内部实现
- [golang中的runtime包教程](golang中的runtime包教程)


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



参考：
- https://go101.org/article/101.html
- https://colobu.com/
- http://legendtkl.com/about/


