---
title: golang面试题汇总
categories:
- Golang
---


## 一、计算机基础
1. 大文件排序？内存不够的情况下，使用归并排序
2. 网络编程中的http keep-alive，tcp keepalive 和 TIME_WAIT是怎么回事？Time_WAIT有什么作用？
	- https://www.cnblogs.com/yjf512/p/5354055.html
	- http://www.nowamagic.net/academy/detail/23350375
	- https://zhuanlan.zhihu.com/p/40013724
3. [孤儿进程和僵尸进程？](https://monkeysayhi.github.io/2018/12/05/%E6%B5%85%E8%B0%88Linux%E5%83%B5%E5%B0%B8%E8%BF%9B%E7%A8%8B%E4%B8%8E%E5%AD%A4%E5%84%BF%E8%BF%9B%E7%A8%8B/)
4. 死锁的条件


## 二、golang语言基本特性
1. make和new的区别？
2. 协程交替执行,使其能顺序输出1-20的自然数？
3. channe关闭后，读操作会怎么样？如何优雅的关闭channel？
4. golang中的main和init函数？
5. [golang中的defer、panic和recover和错误处理方式？](https://wxquare.github.io/2019/03/06/golang_error_handling/)
6. golang中的select关键字？
7. goalng中的struct可以进行比较吗？了解reflect.DeepEqual吗？
8. golang中的set实现？map[interface{}]struct{}
9. goalng中的生产者消费者模式？
10. golang中的context包的用途？
11. golang的编译过程？
12. golang闭包的概念？
13. golang中可以对只运行一次的函数定义为匿名函数，匿名函数对外部变量使用的是引用
14. 将匿名函数赋值为一个变量，该变量就称为一个闭包，为闭包对外层词法域变量是引用的。
```
	package main

	import (
		"fmt"
	)

	func main() {

		x := 1
		f := func() int {
			x++
			return x
		}

		fmt.Println(f())
		fmt.Println(f())
	}

```
15. golang 逃逸分析。go在一定程度消除了堆和栈的区别，因为go在编译的时候进行逃逸分析，来决定一个对象放栈上还是放堆上，不逃逸的对象放栈上，可能逃逸的放堆上



## 三、高级主题
### 2.1. golang中的协程调度？
 
### 2.2. golang中的context包？
https://juejin.im/post/5a6873fef265da3e317e55b6  
https://www.flysnow.org/2017/05/12/go-in-action-go-context.html  

### 2.3 主协程如何等待其余协程完再操作？协程同步的三种方式？

### 2.4.golang网络编程点点滴滴？
	https://colobu.com/2014/12/02/go-socket-programming-TCP/
#### 2.4.1 client如何实现长连接？














