---
title: golang channel通道
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