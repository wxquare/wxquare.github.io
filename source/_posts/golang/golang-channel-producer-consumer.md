---
title: golang 基于channel的生产者消费者模型
categories:
- Golang
---

生产者消费者模式在实际中非常常见，很多问题多可以被抽象成生产者消费者模型；  
通过golang中带缓冲的channel，可以很容易实现；  
同时也能帮助理解channel和协程并发。   

```
    package main
    
    import (
    	"fmt"
    	"sync"
    	"time"
    )
    
    func main() {
    
    	queue := make(chan interface{}, 10)
    	var wg sync.WaitGroup
    	wg.Add(4)
    
    	//producer
    	go func(q chan interface{}) {
    		defer func() {
    			wg.Done()
    			fmt.Println("producer1 exit")
    		}()
    		for i := 0; i < 10; i++ {
    			q <- i
    		}
    	}(queue)
    
    	go func(q chan interface{}) {
    		defer func() {
    			wg.Done()
    			fmt.Println("producer2 exit")
    		}()
    		s := "hello"
    		for _, c := range s {
    			q <- string(c)
    		}
    	}(queue)
    
    	//consumer
    	go func(q chan interface{}) {
    		defer func() {
    			wg.Done()
    			fmt.Println("consumer1 exit")
    		}()
    		for {
    			select {
    			case c := <-q:
    				fmt.Println("consumer1: ", c)
    			case <-time.After(time.Duration(3) * time.Second):
    				return
    			}
    		}
    	}(queue)
    
    	go func(q chan interface{}) {
    		defer func() {
    			wg.Done()
    			fmt.Println("consumer2 exit")
    		}()
    		for {
    			select {
    			case c := <-q:
    				fmt.Println("consumer2: ", c)
    			case <-time.After(time.Duration(3) * time.Second):
    				return
    			}
    		}
    		fmt.Println("end of consumer")
    	}(queue)
    
    	wg.Wait()
    }
```