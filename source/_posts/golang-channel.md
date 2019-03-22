---
title: golang channel通道
---

## 一、channel
	channel是golang中的csp并发模型非常重要组成部分，使用起来非常像阻塞队列。
- 通道channel变量本身就是指针，可用“==”操作符判断是否为同一对象
- 未初始化的channel为nil，需要使用make初始化
- 同步模式的channel必须有配对操作的goroutine出现，否则会一直阻塞，而异步模式在缓冲区未满或者数据未读完前，不会阻塞。
- 内置的cap和len函数返回channel缓冲区大小和当前已缓冲的数量，而对于同步通道则返回0
- 除了使用"<-"发送和接收操作符外，还可以用ok-idom或者range模式处理chanel中的数据。




## 二、基本用法
1. 协程之间传递数据
2. 用作事件通知

