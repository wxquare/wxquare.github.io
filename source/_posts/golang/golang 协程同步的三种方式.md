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
