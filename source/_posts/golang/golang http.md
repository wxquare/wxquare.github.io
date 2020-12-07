---
title: golang http 使用总结
categories:
- Golang
---

最近在项目开发中使用http服务与第三方服务交互，感觉golang的http封装得很好，很方便使用但是也有一些坑需要注意，一是自动复用连接，二是Response.Body的读取和关闭

## http客户端自动复用连接
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

## 读取和关闭Response.Body
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