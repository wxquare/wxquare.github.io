---
title: golang 标准库学习
categories:
- Golang
---


[https://books.studygolang.com/The-Golang-Standard-Library-by-Example/chapter06/06.2.html](https://books.studygolang.com/The-Golang-Standard-Library-by-Example/chapter06/06.2.html)


[golang文件读写三种方式——bufio，ioutil和os.create](https://www.cnblogs.com/bonelee/p/6893398.html)



[https://golangcaff.com/articles/110/two-schemes-for-reading-golang-super-large-files](https://golangcaff.com/articles/110/two-schemes-for-reading-golang-super-large-files)


https://zhuanlan.zhihu.com/p/27050761（golang面试题）
[golang中的runtime包教程](golang中的runtime包教程)

[在腾讯的八年，我的职业思考](https://baijiahao.baidu.com/s?id=1607037562668810273&wfr=spider&for=pc)



1.os包、io、io/ioutil、bufio、path
https://my.oschina.net/solate/blog/719702 文件操作概览
https://my.oschina.net/xinxingegeya/blog/724490 文件读 
https://my.oschina.net/xinxingegeya/blog/725105 文件写
文件操作
目录操作
path操作
IO缓冲
[[译]Go文件操作大全](https://colobu.com/2016/10/12/go-file-operations/)

2.path、path/filepath  
filepath包的功能和path包类似，但是对于不同操作系统提供了更好的支持。filepath包能够自动的根据不同的操作系统文件路径进行转换，所以如果你有跨平台的需求，你需要使用filepath。

    package main
    
    import (
    	"fmt"
    	"path"
    	// "path/filepath"
    )
    
    func main() {
    	fmt.Println(path.Ext("/a/b/c/bar.css"))
    	fmt.Println(path.Base("/a/b/c/"))
    	fmt.Println(path.Dir("/a/b/c"))
    	fmt.Println(path.Clean("/a/b/.."))
    	fmt.Println(path.Join("a/b", "c"))
    	fmt.Println(path.Match("a*/b", "a/c/b"))
    	fmt.Println(path.Split("static/myfile.css"))
    }


3.time包学习 日期和时间  
[https://juejin.im/post/5ae32a8651882567105f7dd3](https://juejin.im/post/5ae32a8651882567105f7dd3)  
- 2006-01-02 15:04:05  
- 获取时间点、格式化为某种格式  
- 时间转为为字符串  
- 字符串转为时间类型
- 时间类型转时间戳
- 时间段Duration,3*time.Second,time.Hour
- Ticker类型和Timer类型

    package main
    
    import (
    	"fmt"
    	"sync"
    	"time"
    )
    
    func main() {
    	fmt.Println(time.Now())
    
    	//strimg to time
    	t, err := time.Parse("2006-01-02 15:04:05", "2018-04-23 12:24:51")
    	if err == nil {
    		fmt.Println(t)
    	}
    
    	t, err = time.ParseInLocation("2006-01-02 15:04:05", "2018-04-23 12:24:51", time.Local)
    	if err == nil {
    		fmt.Println(t)
    	}
    
    	//get time and conver to string
    	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
    
    	//time type to unix stamp
    	fmt.Println(t.Unix())
    
    	time.Sleep(3 * time.Second)
    	time.Sleep(time.Second * 1)
    	time.Sleep(time.Duration(1) * time.Second)
    	// time.Sleep(1 * time.Hour)
    
    	tp, err := time.ParseDuration("1.5s")
    	if err == nil {
    		fmt.Println(tp)
    	}
    
    	//compare time
    	fmt.Println(time.Now().After(t))
    
    	var wg sync.WaitGroup
    	wg.Add(2)
    	//NewTimer 创建一个 Timer，它会在最少过去时间段 d 后到期，向其自身的 C 字段发送当时的时间
    	timer1 := time.NewTimer(2 * time.Second)
    
    	//NewTicker 返回一个新的 Ticker，该 Ticker 包含一个通道字段，并会每隔时间段 d 就向该通道发送当时的时间。它会调
    	//整时间间隔或者丢弃 tick 信息以适应反应慢的接收者。如果d <= 0会触发panic。关闭该 Ticker 可
    	//以释放相关资源。
    	ticker1 := time.NewTicker(2 * time.Second)
    
    	go func(t *time.Ticker) {
    		defer wg.Done()
    		for {
    			<-t.C
    			fmt.Println("get ticker1", time.Now().Format("2006-01-02 15:04:05"))
    		}
    	}(ticker1)
    
    	go func(t *time.Timer) {
    		defer wg.Done()
    		for {
    			<-t.C
    			fmt.Println("get timer", time.Now().Format("2006-01-02 15:04:05"))
    			//Reset 使 t 重新开始计时，（本方法返回后再）等待时间段 d 过去后到期。如果调用时t
    			//还在等待中会返回真；如果 t已经到期或者被停止了会返回假。
    			t.Reset(2 * time.Second)
    		}
    	}(timer1)
    
    	wg.Wait()
    }



4.unsafe包学习 

 
golang指针学习
https://studygolang.com/articles/10953

https://www.jianshu.com/p/c394436ec9e5?utm_campaign=maleskine&utm_content=note&utm_medium=seo_notes&utm_source=recommendation  

https://juejin.im/entry/5829548bd203090054000ab6
- 普通指针  
- unsafe.Pointer (*int) 是int指针类型的一个别名 
- uintptr  
- 出于安全原因，Golang不允许以下之间的直接转换：
- 两个不同指针类型的值，例如 int64和 float64。
- 指针类型和uintptr的值。
- 但是借助unsafe.Pointer，我们可以打破Go类型和内存安全性，并使上面的转换成为可能。这怎么可能发生？让我们阅读unsafe包文档中列出的规则：
- 
- 任何类型的指针值都可以转换为unsafe.Pointer。
- unsafe.Pointer可以转换为任何类型的指针值。
- uintptr可以转换为unsafe.Pointer。
- unsafe.Pointer可以转换为uintptr



5.golang bytes