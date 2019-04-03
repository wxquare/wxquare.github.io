---
title: golang http 使用总结
categories:
- Golang
---

最近在项目开发中使用http服务与第三方服务交互，感觉golang的http封装得很好，很方便使用但是也有一些坑需要注意，一是自动复用连接，二是Response.Body的读取和关闭

## 1.TCP keepalive 和 http keep-alive
  keepalive虽然不是TCP协议规范的内容， 但是Linux和windows中都实现了keepalive功能。因为在使用TCP长连接的时候，需要对TCP连接进行保活。操作系统通过在TCP连接定时发送keepalive探测包，实现**连接保活、检测连接**的有效性和**自动关闭无效连接**的作用。
  TCP的keepalive是默认关闭的，可以通过内核设置或者SO_KEEPALIVE才能生效。

  除了TCP自带的Keeplive机制，实现业务中经常在业务层面定制**“心跳”**功能，主要有以下几点考虑：  
- TCP自带的keepalive使用简单，仅提供连接是否存活的功能  
- 应用层心跳包不依赖于传输协议，支持tcp和udp  
- 应用层心跳包可以定制，可以应对更加复杂的情况或者传输一些额外的消息  
- Keepalive仅仅代表连接保持着，而心跳往往还表示服务正常工作  
	
在 HTTP 1.0 时期，每个 TCP 连接只会被一个 HTTP Transaction（请求加响应）使用，请求时建立，请求完成释放连接。当网页内容越来越复杂，包含大量图片、CSS 等资源之后，这种模式效率就显得太低了。所以，在 HTTP 1.1 中，引入了 HTTP persistent connection 的概念，也称为 HTTP keep-alive，目的是复用TCP连接，在一个TCP连接上进行多次的HTTP请求从而提高性能。HTTP1.0中默认是关闭的，需要在HTTP头加入"Connection: Keep-Alive"，才能启用Keep-Alive；HTTP1.1中默认启用Keep-Alive，加入"Connection: close "，才关闭。

两者在写法上不同，http keep-alive 中间有个"-"符号。 **HTTP协议的keep-alive 意图在于连接复用**，同一个连接上串行方式传递请求-响应数据。**TCP的keepalive机制意图在于保活、心跳，检测连接错误。**

## 2.http客户端自动复用连接
首先用代码直观的体验http客户端自动复用连接特点  
server.go

	func main() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "hello!")
		})
		http.ListenAndServe(":8848", nil)
	}

client.go

	func doReq() {
		resp, err := http.Get("http://127.0.0.1:8848/test")
		if err != nil {
			fmt.Println(err)
			return
		}
		io.Copy(os.Stdout, resp.Body)
		defer resp.Body.Close()
	}
	
	func main() {
		//http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 10
		for {
			go doReq()
			go doReq()
			//	go doReq()
			time.Sleep(300 * time.Millisecond)
		}
	}

测试1：执行`netstat | grep "8848" | wc -l`  结果：一直都是4  
测试2：增加一个go doReq(),继续测试，结果：是一直增大  
测试3：在测试2的基础上设置MaxIdleConnsPerHost = 10，结果：一直都是6

测试1已经能说明golang的http会自动复用连接  
测试2为什么连接数量会一直增加呢？原因是golang中默认只保持两条持久连接，http.Transport没有设置MaxIdleConnPerHost，于是便采用了默认的DefaultMaxIdleConnsPerHost，这个值是2。  
测试3通过加大MaxIdleConnPerHost的值，就能高效的利用http的自动复用机制。

## 3.读取和关闭Response.Body
将Resonse.Body的读取的代码屏蔽，继续测试。

    func doReq() {
    	resp, err := http.Get("http://127.0.0.1:8848/test")
    	if err != nil {
    		fmt.Println(err)
    		return
    	}
    	//io.Copy(os.Stdout, resp.Body)
    	defer resp.Body.Close()
    }  

测试结果发现，连接数一直增加。    
产生的原因：body实际上是一个嵌套了多层的net.TCPConn，当body没有被完全读取，也没有被关闭是，那么这次的http事物就没有完成，除非连接因为超时终止了，否则相关资源无法被回收。
从实现上看只要body被读完，连接就能被回收，只有需要抛弃body时才需要close，似乎不关闭也可以。但那些正常情况能读完的body，即第一种情况，在出现错误时就不会被读完，即转为第二种情况。而分情况处理则增加了维护者的心智负担，所以始终close body是最佳选择。


参考：  
[1].[https://my.oschina.net/hebaodan/blog/1609245](https://my.oschina.net/hebaodan/blog/1609245)  
[2].[https://www.jianshu.com/p/407fada3cc9d](https://www.jianshu.com/p/407fada3cc9d)  
[3].[https://serholiu.com/go-http-client-keepalive](https://serholiu.com/go-http-client-keepalive)