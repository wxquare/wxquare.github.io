
1. [Go程序调试、分析与优化](https://tonybai.com/2015/08/25/go-debugging-profiling-optimization/)
2. https://blog.csdn.net/moxiaomomo/article/details/77096814

2.golang程序的debug和调优  
https://blog.golang.org/profiling-go-programs

3.golang中的sort使用？？  
[https://books.studygolang.com/The-Golang-Standard-Library-by-Example/chapter03/03.1.html](https://books.studygolang.com/The-Golang-Standard-Library-by-Example/chapter03/03.1.html)


4.golang中的文件和IO操作？？




4.关于golang的反射，[Golang的反射reflect深入理解和示例](https://juejin.im/post/5a75a4fb5188257a82110544)
Go的反射包怎么找到对应的方法（这里忘记怎么问的，直接说不会，只用了DeepEqual，简单讲了DeepEqual）

5.[golang标准库](https://blog.csdn.net/preyta/column/info/21866)

6.[Golang工程经验](https://juejin.im/post/5a6873fb518825733e60a1ae)

7.golang 如何实现单例模式？？
github.com/dropbox/godropbox/singleton"


8.[golang Contenxt深入理解？](https://juejin.im/post/5a6873fef265da3e317e55b6)



9.[以sort为例子，写出基于interface的泛型编程？](https://juejin.im/post/5a6873fb518825733e60a1ae)


10.基础资源的封装（mysql、redis、memcache）


11.如何通过信号量或者channel控制协程的数量？
 
12.golang http请求和rpc请求

13.panic和error的处理

14.Golang、tag和json操作。
[golang中使用json，经常会使用到两个函数](https://studygolang.com/articles/9028)
https://zhuanlan.zhihu.com/p/32279896，tag与reflect


15.golang包管理和go module？？？？


##16.golang如何处理错误？
1. error接口
2. 自定义错误类型
- [错误类型断言（type assertion](https://studygolang.com/articles/11419)）
- 字符串匹配  
- 18.defer？？https://colobu.com/2019/01/22/Runtime-overhead-of-using-defer-in-go/
- errors包，errors.New()





17.golang中的select可以用来做什么？和switch的区别？
https://colobu.com/2017/07/07/select-vs-switch-in-golang/
https://yanyiwu.com/work/2014/11/08/golang-select-typical-usage.html


18.defer？？
https://colobu.com/2019/01/22/Runtime-overhead-of-using-defer-in-go/


19.理解golang中的context？？？
https://juejin.im/post/5a6873fef265da3e317e55b6
http://www.opscoder.info/golang_context.html

20.如何在Go中使用Protobuf
https://studygolang.com/articles/7394

21.golang 操作mysql
https://www.cnblogs.com/hanyouchun/p/6708037.html
https://www.cnblogs.com/shiluoliming/p/7904547.html
https://studygolang.com/articles/12509
https://www.cnblogs.com/zhuyp1015/p/3561470.html

21.golang 操作redis
https://blog.csdn.net/u014520797/article/details/54577195
https://blog.csdn.net/wangshubo1989/article/details/75050024

22.golang中make和new的区别？？？
https://my.oschina.net/xinxingegeya/blog/837140

23.golang正确处理http.Response.Body??
https://zhuanlan.zhihu.com/p/23227849

24.聊聊 TCP 中的 KeepAlive 机制??
http://www.importnew.com/27624.html
https://my.oschina.net/hebaodan/blog/1609245
Go HTTP Client 持久连接:https://serholiu.com/go-http-client-keepalive 
[golang]为什么Response.Body需要被关闭?
https://www.jianshu.com/p/407fada3cc9d


25.golang sync.Pool
https://blog.csdn.net/yongjian_lian/article/details/42058893
https://deepzz.com/post/golang-sync-package-usage.html

26.golang标准库学习
https://books.studygolang.com/The-Golang-Standard-Library-by-Example/


27.golang连接池的实现？？
https://segmentfault.com/a/1190000013089363


28.goroutine的调度？
https://tonybai.com/2017/06/23/an-intro-about-goroutine-scheduler/


29.golang超大文件读取策略？？
https://learnku.com/articles/23559/two-schemes-for-reading-golang-super-large-files
https://colobu.com/2016/10/12/go-file-operations/
https://www.cnblogs.com/bonelee/p/6893398.html


31.golang中的匿名函数和闭包？？
https://blog.csdn.net/wangshubo1989/article/details/79217291


32.go toml 配置文件解析
https://github.com/pelletier/go-toml


33.深入了解golang垃圾回收
http://www.opscoder.info/golang_gc.html


34.深入解析golang
https://tiancaiamao.gitbooks.io/go-internals/zh/

35.golang比较好博客  
https://tonybai.com/  
https://colobu.com   
http://legendtkl.com/categories/golang/page/2/
https://cyc2018.github.io/CS-Notes  

36.golang工程经验？？
https://juejin.im/post/5a6873fb518825733e60a1ae

37.golang中的select可以用来做什么，与switch的区别？  
https://colobu.com/2017/07/07/select-vs-switch-in-golang/  


38.golang socket编程？？
https://victoriest.gitbooks.io/golang-tcp-server/content

39.golang http编程？？？  《go语言编程 5.4节》
- http get跟head   
- http 401,403  
- http keep-alive  
- http能不能一次连接多次请求，不等后端返回  
- client如何实现长连接
- http，https
状态码401,301,302,201



43.golang中channel？？（有缓冲和无缓冲）
退出程序时怎么防止channel没有消费完，这里一开始有点没清楚面试官问的，然后说了监听中断信号，做退出前的处理，然后面试官说不是这个意思，然后说发送前先告知长度，长度要是不知道呢？close channel下游会受到0值，可以利用这点（这里也有点跟面试官说不明白）


44.手写生成者消费者模式？？
生产者消费者模式，手写代码（Go直接使用channel实现很简单，还想着面试官会不会不让用channel实现，不用channel的可以使用数组加条件变量），channel缓冲长度怎么决定，怎么控制上游生产速度过快，这里没说出解决方案，只是简单说了channel长度可以与上下游的速度比例成线性关系，面试官说这是一种解决方案

45.手写循环队列  
写的循环队列是不是线程安全，不是，怎么保证线程安全，加锁，效率有点低啊，然后面试官就提醒Go推崇原子操作和channel


46.sync.Pool用过吗，为什么使用，对象池，避免频繁分配对象（GC有关），那里面的对象是固定的吗？不清楚，没看过这个的源码

47.tcp粘包
理粘包断包实现，面试官以为是negle算法有关，解释了下negle跟糊涂窗口综合征有关，然后面试官觉得其他项目是crud就没问了

48.有没有网络编程，有，怎么看连接状态？netstat，有哪些？ESTABLISHED，LISTEN等等，有异常情况吗？TIME_WAIT很多，为什么？大量短链接



1. [Go程序调试、分析与优化](https://tonybai.com/2015/08/25/go-debugging-profiling-optimization/)
2. https://blog.csdn.net/moxiaomomo/article/details/77096814

2.golang程序的debug和调优  
https://blog.golang.org/profiling-go-programs

3.golang中的sort使用？？  
[https://books.studygolang.com/The-Golang-Standard-Library-by-Example/chapter03/03.1.html](https://books.studygolang.com/The-Golang-Standard-Library-by-Example/chapter03/03.1.html)


4.golang中的文件和IO操作？？




4.关于golang的反射，[Golang的反射reflect深入理解和示例](https://juejin.im/post/5a75a4fb5188257a82110544)
Go的反射包怎么找到对应的方法（这里忘记怎么问的，直接说不会，只用了DeepEqual，简单讲了DeepEqual）

5.[golang标准库](https://blog.csdn.net/preyta/column/info/21866)

6.[Golang工程经验](https://juejin.im/post/5a6873fb518825733e60a1ae)

7.golang 如何实现单例模式？？
github.com/dropbox/godropbox/singleton"


8.[golang Contenxt深入理解？](https://juejin.im/post/5a6873fef265da3e317e55b6)



9.[以sort为例子，写出基于interface的泛型编程？](https://juejin.im/post/5a6873fb518825733e60a1ae)


10.基础资源的封装（mysql、redis、memcache）


11.如何通过信号量或者channel控制协程的数量？
 
12.golang http请求和rpc请求

13.panic和error的处理

14.Golang、tag和json操作。
[golang中使用json，经常会使用到两个函数](https://studygolang.com/articles/9028)
https://zhuanlan.zhihu.com/p/32279896，tag与reflect


15.golang包管理和go module？？？？


##16.golang如何处理错误？
1. error接口
2. 自定义错误类型
- [错误类型断言（type assertion](https://studygolang.com/articles/11419)）
- 字符串匹配  
- 18.defer？？https://colobu.com/2019/01/22/Runtime-overhead-of-using-defer-in-go/
- errors包，errors.New()





17.golang中的select可以用来做什么？和switch的区别？
https://colobu.com/2017/07/07/select-vs-switch-in-golang/
https://yanyiwu.com/work/2014/11/08/golang-select-typical-usage.html


18.defer？？
https://colobu.com/2019/01/22/Runtime-overhead-of-using-defer-in-go/


19.理解golang中的context？？？
https://juejin.im/post/5a6873fef265da3e317e55b6
http://www.opscoder.info/golang_context.html

20.如何在Go中使用Protobuf
https://studygolang.com/articles/7394

21.golang 操作mysql
https://www.cnblogs.com/hanyouchun/p/6708037.html
https://www.cnblogs.com/shiluoliming/p/7904547.html
https://studygolang.com/articles/12509
https://www.cnblogs.com/zhuyp1015/p/3561470.html

21.golang 操作redis
https://blog.csdn.net/u014520797/article/details/54577195
https://blog.csdn.net/wangshubo1989/article/details/75050024

22.golang中make和new的区别？？？
https://my.oschina.net/xinxingegeya/blog/837140

23.golang正确处理http.Response.Body??
https://zhuanlan.zhihu.com/p/23227849

24.聊聊 TCP 中的 KeepAlive 机制??
http://www.importnew.com/27624.html
https://my.oschina.net/hebaodan/blog/1609245
Go HTTP Client 持久连接:https://serholiu.com/go-http-client-keepalive 
[golang]为什么Response.Body需要被关闭?
https://www.jianshu.com/p/407fada3cc9d


25.golang sync.Pool
https://blog.csdn.net/yongjian_lian/article/details/42058893
https://deepzz.com/post/golang-sync-package-usage.html

26.golang标准库学习
https://books.studygolang.com/The-Golang-Standard-Library-by-Example/


27.golang连接池的实现？？
https://segmentfault.com/a/1190000013089363


28.goroutine的调度？
https://tonybai.com/2017/06/23/an-intro-about-goroutine-scheduler/


29.golang超大文件读取策略？？
https://learnku.com/articles/23559/two-schemes-for-reading-golang-super-large-files
https://colobu.com/2016/10/12/go-file-operations/
https://www.cnblogs.com/bonelee/p/6893398.html


31.golang中的匿名函数和闭包？？
https://blog.csdn.net/wangshubo1989/article/details/79217291


32.go toml 配置文件解析
https://github.com/pelletier/go-toml


33.深入了解golang垃圾回收
http://www.opscoder.info/golang_gc.html


34.深入解析golang
https://tiancaiamao.gitbooks.io/go-internals/zh/

35.golang比较好博客  
https://tonybai.com/  
https://colobu.com   
http://legendtkl.com/categories/golang/page/2/
https://cyc2018.github.io/CS-Notes  

36.golang工程经验？？
https://juejin.im/post/5a6873fb518825733e60a1ae

37.golang中的select可以用来做什么，与switch的区别？  
https://colobu.com/2017/07/07/select-vs-switch-in-golang/  


38.golang socket编程？？
https://victoriest.gitbooks.io/golang-tcp-server/content

39.golang http编程？？？  《go语言编程 5.4节》
- http get跟head   
- http 401,403  
- http keep-alive  
- http能不能一次连接多次请求，不等后端返回  
- client如何实现长连接
- http，https
状态码401,301,302,201



43.golang中channel？？（有缓冲和无缓冲）
退出程序时怎么防止channel没有消费完，这里一开始有点没清楚面试官问的，然后说了监听中断信号，做退出前的处理，然后面试官说不是这个意思，然后说发送前先告知长度，长度要是不知道呢？close channel下游会受到0值，可以利用这点（这里也有点跟面试官说不明白）


44.手写生成者消费者模式？？
生产者消费者模式，手写代码（Go直接使用channel实现很简单，还想着面试官会不会不让用channel实现，不用channel的可以使用数组加条件变量），channel缓冲长度怎么决定，怎么控制上游生产速度过快，这里没说出解决方案，只是简单说了channel长度可以与上下游的速度比例成线性关系，面试官说这是一种解决方案

45.手写循环队列  
写的循环队列是不是线程安全，不是，怎么保证线程安全，加锁，效率有点低啊，然后面试官就提醒Go推崇原子操作和channel


46.sync.Pool用过吗，为什么使用，对象池，避免频繁分配对象（GC有关），那里面的对象是固定的吗？不清楚，没看过这个的源码

47.tcp粘包
理粘包断包实现，面试官以为是negle算法有关，解释了下negle跟糊涂窗口综合征有关，然后面试官觉得其他项目是crud就没问了

48.有没有网络编程，有，怎么看连接状态？netstat，有哪些？ESTABLISHED，LISTEN等等，有异常情况吗？TIME_WAIT很多，为什么？大量短链接




书籍
《图解http》


## golang 使用http总结

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


## golang sync.pool 和 通用连接池
## 1.sync.Pool 基本使用
[https://golang.org/pkg/sync/](https://golang.org/pkg/sync/)  
sync.Pool的使用非常简单，它具有以下几个特点：
  
- sync.Pool设计目的是存放已经分配但暂时不用的对象，供以后使用，以减轻gc的代价，提高效率  
- 存储在Pool中的对象会随时被gc自动回收，Pool中对象的缓存期限为两次gc之间  
- 用户无法定义sync.Pool的大小，其大小仅仅受限于内存的大小     
- sync.Pool支持多协程之间共享
  
sync.Pool的使用非常简单，定义一个Pool对象池时，需要提供一个New函数，表示当池中没有对象时，如何生成对象。对象池Pool提供Get和Put函数从Pool中取和存放对象。

下面有一个简单的实例，直接运行是会打印两次“new an object”,注释掉runtime.GC(),发现只会调用一次New函数，表示实现了对象重用。

	package main
	
	import (
		"fmt"
		"runtime"
		"sync"
	)
	
	func main() {
		p := &sync.Pool{
			New: func() interface{} {
				fmt.Println("new an object")
				return 0
			},
		}
	
		a := p.Get().(int)
		a = 100
		p.Put(a)
		runtime.GC()
		b := p.Get().(int)
		fmt.Println(a, b)
	}

## 2.sync.Pool 如何支持多协程共享？
sync.Pool支持多协程共享，为了尽量减少竞争和加锁的操作，golang在设计的时候为每个P（核）都分配了一个子池，每个子池包含一个私有对象和共享列表。 私有对象只有对应的和核P能够访问，而共享列表是与其它P共享的。  

在golang的GMP调度模型中，我们知道协程G最终会被调度到某个固定的核P上。当一个协程在执行Pool的get或者put方法时，首先对改核P上的子池进行操作，然后对其它核的子池进行操作。因为一个P同一时间只能执行一个goroutine，所以对私有对象存取操作是不需要加锁的，而共享列表是和其他P分享的，因此需要加锁操作。  

一个协程希望从某个Pool中获取对象，它包含以下几个步骤：  
1. 判断协程所在的核P中的私有对象是否为空，如果非常则返回，并将改核P的私有对象置为空    
2. 如果协程所在的核P中的私有对象为空，就去改核P的共享列表中获取对象（需要加锁）  
3. 如果协程所在的核P中的共享列表为空，就去其它核的共享列表中获取对象（需要加锁）  
4. 如果所有的核的共享列表都为空，就会通过New函数产生一个新的对象  

在sync.Pool的源码中，每个核P的子池的结构如下所示：   
  
	// Local per-P Pool appendix.
	type poolLocalInternal struct {
		private interface{}   // Can be used only by the respective P.
		shared  []interface{} // Can be used by any P.
		Mutex                 // Protects shared.
	}
更加细致的sync.Pool源码分析，可参考[http://jack-nie.github.io/go/golang-sync-pool.html](http://jack-nie.github.io/go/golang-sync-pool.html)

## 3.为什么不使用sync.pool实现连接池？
刚开始接触到sync.pool时，很容易让人联想到连接池的概念，但是经过仔细分析后发现sync.pool并不是适合作为连接池，主要有以下两个原因： 
 
- 连接池的大小通常是固定且受限制的，而sync.Pool是无法控制缓存对象的数量，只受限于内存大小，不符合连接池的目标  
- sync.Pool对象缓存的期限在两次gc之间,这点也和连接池非常不符合

golang中连接池通常利用channel的缓存特性实现。当需要连接时，从channel中获取，如果池中没有连接时，将阻塞或者新建连接，新建连接的数量不能超过某个限制。

[https://github.com/goctx/generic-pool](https://github.com/goctx/generic-pool)基于channel提供了一个通用连接池的实现

	package pool
	
	import (
		"errors"
		"io"
		"sync"
		"time"
	)
	
	var (
		ErrInvalidConfig = errors.New("invalid pool config")
		ErrPoolClosed    = errors.New("pool closed")
	)
	
	type Poolable interface {
		io.Closer
		GetActiveTime() time.Time
	}
	
	type factory func() (Poolable, error)
	
	type Pool interface {
		Acquire() (Poolable, error) // 获取资源
		Release(Poolable) error     // 释放资源
		Close(Poolable) error       // 关闭资源
		Shutdown() error            // 关闭池
	}
	
	type GenericPool struct {
		sync.Mutex
		pool        chan Poolable
		maxOpen     int  // 池中最大资源数
		numOpen     int  // 当前池中资源数
		minOpen     int  // 池中最少资源数
		closed      bool // 池是否已关闭
		maxLifetime time.Duration
		factory     factory // 创建连接的方法
	}
	
	func NewGenericPool(minOpen, maxOpen int, maxLifetime time.Duration, factory factory) (*GenericPool, error) {
		if maxOpen <= 0 || minOpen > maxOpen {
			return nil, ErrInvalidConfig
		}
		p := &GenericPool{
			maxOpen:     maxOpen,
			minOpen:     minOpen,
			maxLifetime: maxLifetime,
			factory:     factory,
			pool:        make(chan Poolable, maxOpen),
		}
	
		for i := 0; i < minOpen; i++ {
			closer, err := factory()
			if err != nil {
				continue
			}
			p.numOpen++
			p.pool <- closer
		}
		return p, nil
	}
	
	func (p *GenericPool) Acquire() (Poolable, error) {
		if p.closed {
			return nil, ErrPoolClosed
		}
		for {
			closer, err := p.getOrCreate()
			if err != nil {
				return nil, err
			}
			// 如果设置了超时且当前连接的活跃时间+超时时间早于现在，则当前连接已过期
			if p.maxLifetime > 0 && closer.GetActiveTime().Add(time.Duration(p.maxLifetime)).Before(time.Now()) {
				p.Close(closer)
				continue
			}
			return closer, nil
		}
	}
	
	func (p *GenericPool) getOrCreate() (Poolable, error) {
		select {
		case closer := <-p.pool:
			return closer, nil
		default:
		}
		p.Lock()
		if p.numOpen >= p.maxOpen {
			closer := <-p.pool
			p.Unlock()
			return closer, nil
		}
		// 新建连接
		closer, err := p.factory()
		if err != nil {
			p.Unlock()
			return nil, err
		}
		p.numOpen++
		p.Unlock()
		return closer, nil
	}
	
	// 释放单个资源到连接池
	func (p *GenericPool) Release(closer Poolable) error {
		if p.closed {
			return ErrPoolClosed
		}
		p.Lock()
		p.pool <- closer
		p.Unlock()
		return nil
	}
	
	// 关闭单个资源
	func (p *GenericPool) Close(closer Poolable) error {
		p.Lock()
		closer.Close()
		p.numOpen--
		p.Unlock()
		return nil
	}
	
	// 关闭连接池，释放所有资源
	func (p *GenericPool) Shutdown() error {
		if p.closed {
			return ErrPoolClosed
		}
		p.Lock()
		close(p.pool)
		for closer := range p.pool {
			closer.Close()
			p.numOpen--
		}
		p.closed = true
		p.Unlock()
		return nil
	}
参考：  
[1].[https://blog.csdn.net/yongjian_lian/article/details/42058893](https://blog.csdn.net/yongjian_lian/article/details/42058893)  
[2].[https://segmentfault.com/a/1190000013089363](https://segmentfault.com/a/1190000013089363)  
[3].[http://jack-nie.github.io/go/golang-sync-pool.html](http://jack-nie.github.io/go/golang-sync-pool.html)

## golang 数据结构
###38.字符串string
1. 基本数组类型s := "hello,world"
2. 一旦初始化后不允许修改字符串的内容
3. 常用函数s1+s2,len(s1)等
4. <font color=red>字符串与数值类型的不能强制转化，要使用strconv包中的函数</font>
5. 标准库strings提供了许多字符串操作的函数,例如Split、HasPrefix,Trim。

###39.数组array: [3]int{1,2,3}
1. <font color=red>**数组是值类型**</font>，数组传参发生拷贝
2. 定长
3. 数组的创建、初始化、访问和遍历range，len(arr)求数组的长度
  
###40.数组切片slice: make([]int,len,cap)
1. <font color=red>**slice是引用类型**</font>
2. 变长，用容量和长度的区别，分别使用cap和len函数获取
3. 内存结构：指针、cap、size共24字节
4. 常用函数，append，cap，len
5. 切片动态扩容，拷贝

###41.存储kv的哈希表map：make(map[string]int,5) 
1.  map的创建，为了避免频繁的扩容和迁移，创建map时应指定适当的大小
2.  无序
3.  赋值，相同键值会覆盖
4.  遍历，range
5.  [如何实现顺序遍历？](https://blog.csdn.net/slvher/article/details/44779081)
6.  [内部hashmap的实现原理](https://ninokop.github.io/2017/10/24/Go-Hashmap%E5%86%85%E5%AD%98%E5%B8%83%E5%B1%80%E5%92%8C%E5%AE%9E%E7%8E%B0/)。内部结构（bucket），扩容与迁移，删除。 
7.  如何保证map的协程安全性？[sync.map](https://colobu.com/2017/07/11/dive-into-sync-Map/)? 


### 42.集合set
1. golang中本身没有提供set，但可以通过map自己实现
2. 利用map键值不可重复的特性实现set，value为空结构体。 map[interface{}]struct{} 
3. [如何自己实现set？](https://studygolang.com/articles/11179)

  
### 43.容器container/heap、list、ring
1. heap与优先队列，最小堆
2. 链表list，双向列表
3. 循环队列ring
4. <font color=red>golang没有提供stack，可自己实现</font>
5. <font color=red>golang没有提供queue，但可以通过channel替换或者自己实现</font>


##延伸问题：
####1.如何比较struct/slice/map?
- struct没有slice和map类型时可直接判断
- slice和map本身不可比较，需要使用reflect.DeepEqual()。
- 当struct中包含slice和map等字段时，也要使用reflect.DeepEqual().
- [https://stackoverflow.com/questions/24534072/how-to-compare-struct-slice-map-are-equal](https://stackoverflow.com/questions/24534072/how-to-compare-struct-slice-map-are-equal)
- [https://studygolang.com/articles/11342](https://studygolang.com/articles/11342)


## golang 错误处理
错误处理是任何编程语言都不可避免的话题，golang错误处理的方式虽然备受争议，但总体是符合工程语言的要求的。熟悉golang错误处理的方式，需要掌握以下五点:

##1.根据error接口自定义错误类型
golang中引入了关于错误处理的标准模式error接口，实际中可以通过实现error结构，自定义错误类型。error接口只有一个Error方法，它返回一个string表示错误的内容。  
error接口： 
 
    type error Interface{  
    	Error() string  
    }  
自定义错误类型：
  
    type MyError struct {
    	ErrorInfo string
    }
    
    func (e *MyError) Error() string {
    	return ErrorInfo
    }
    
##2.通过errors包生成error对象
errors包提供New方法，非常方便生成error对象，例如：

    func foo() error {
    	return errors.New("foo error")
    }

##3.panic和recover  
- 当一个函数抛出panic错误时，正常的函数流程立即终止
- defer关键字延迟执行的语句将正常执行
- 逐层向上执行panic过程，直到所属的goroutine中所有执行的函数终止
- recover用于终止panic的错误处理流程   

例如：   

    func main() {
    	//defer
    	defer func() {
    		fmt.Println("defer func(){}()")
    		if r := recover(); r != nil {
    			fmt.Println("Runtime error caught!", r)
    		}
    	}()
    	panic("throw a panic")
    	fmt.Println("hello,world")
    }
##4.defer
defer是golang中非常好用的一个错误处理方式，函数正常退出和出错时，defer中的语句也会被执行，作用相当于C++中的析构函数，对资源泄露非常有帮助。实际使用时需要注意：  
- defer语句的位置  
- defer语句执行的顺序  
defer语句的调用遵循的顺序是先进后出，即最后一个defer语句最先被执行。


##5.接口对象类型断言
golang中接口对象非常方便，因此提供类型判断，防止出现panic错误。例如：

    type Person struct {
    	Name string
    	age  int
    }
    
    func main() {
    	//Type Assertion
    	var v interface{}
    	v = Person{"bob", 12}
    	if f, ok := v.(Person); ok {
    		fmt.Println(f.Name)
    	}
    }

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


golang 并发和调度

##1.C/C++操作系统线程调度的缺点
- 创建线程和切换线程代价较大，线程数量不能太多，经常采用线程池或者网络IO复用技术，因此线程调度难以扩展  
- 线程的同步和通信较为麻烦
- 加锁易犯错且易效率低


##2.Golang运行时的协程调度的特点  
- 创建协程goroutine的代价低
- 协程数量大，可达数十万个
- 协程的同步和通信机制简单，基于channel  
- G-M-P调度模型较为高效，实现协程阻塞、抢占式调度、stealing等情况，具有较高的调度效率


##3.Golang运行时调度器
golang运行时调度器位于用户golang代码和操作系统os之间，它决定何时哪个goroutine将获得资源开始执行、哪个goroutine应该停止执行让出资源、哪个goroutine应该被唤醒恢复执行等。由于操作系统是以线程为调度的单位，因此golang运行时调度器实际上是将协程调度到具体的线程上。

随着golang版本的更新，其调度模型也在不断的优化，goalng 1.1版本中的G-P-M模型使其调度模型基本成型，也具有较高的效率。为了实现调度的可扩展性（scalable），在协程和线程之间增加了一个逻辑层P。

- goroutine 都由一个G结构表示，它管理着goroutine的栈和状态
- 运行时管理着G，并将它们映射到Logical Processor P上。P可以看作是一个抽象的资源或者一个上下文
- 为了运行goroutine，M需要持有上下文P，M会从P的queue弹出一个goutine并执行



##4.其它概念：
###4.1抢占式调度
和操作系统按时间片调度线程不同，Go并没有时间片的概念。如果某个G没有进行system call调用、没有进行I/O操作、没有阻塞在一个channel操作上，那么m是如何让G停下来并调度下一个runnable G的呢？答案是：G是被抢占调度的
###4.2channel阻塞或者network I/O情况下的调度
如果G被阻塞在某个channel操作或network I/O操作上时，G会被放置到某个wait队列中，而M会尝试运行下一个runnable的G；如果此时没有runnable的G供m运行，那么m将解绑P，并进入sleep状态。当I/O available或channel操作完成，在wait队列中的G会被唤醒，标记为runnable，放入到某P的队列中，绑定一个M继续执行。
###4.3system call阻塞状态下的调度
如果G被阻塞在某个system call操作上，那么不光G会阻塞，执行该G的M也会解绑P(实质是被sysmon抢走了)，与G一起进入sleep状态。如果此时有idle的M，则P与其绑定继续执行其他G；如果没有idle M，但仍然有其他G要去执行，那么就会创建一个新M。

当阻塞在syscall上的G完成syscall调用后，G会去尝试获取一个可用的P，如果没有可用的P，那么G会被标记为runnable，之前的那个sleep的M将再次进入sleep。

##5.golang调度器的跟踪调试
https://colobu.com/2016/04/19/Scheduler-Tracing-In-Go/

参考：  
- https://tonybai.com/2017/06/23/an-intro-about-goroutine-scheduler/  
- https://colobu.com/2017/05/04/go-scheduler/  


#golang命名和编码规范(整理)#

##1.命名##
- **服务名**：建议使用动名词短语比如FeedsAdPlayerServer、KbAdReportServer、KbAdScoreServer等
- **目录与包名**：包名与目录名相同，包名应该为小写单词，不要使用下划线或者混合大小写，建议使用比较短的名词短语，例如l5、util、log、kafka、ppMonitor、dao、wsd等
- **文件名和类名**：有意义的名词短语或者动名词短语，如果文件中中定义的是一个结构体（类），尽量使文件名和类名保持一致，例如KafkaConsumer.go、KafkaConfig.go、WsdReporter.go、FlowControlCache.go、CptTdwReportTask.cpp、NewsDao.h  
- **函数名**：有意义的动词+名词短语，例如GetXxxx()、GetXxxxByXxxx()、SetXxxx()、PullXxxx()、CheckXxxx()、BuildXxxx()、ReportXxxx()等，根据函数在其它包的可见性决定使用大写字母或者小写字母开头

变量名，由于类名和函数名多为大写字母开头，建议变量名在满足可见性的要求下，尽可能使用小写字母开头   
- **全局变量名**：驼峰式，结合是否可导出确定首字母大小写  
- **局部变量名**：驼峰式，小写字母开头  
- **参数传递**：驼峰式，小写字母开头  
- **常量**：采用下划线连接的全大写的名字短语

##2.import
- 对import的包进行分组管理，分别为标准库、程序内部包、第三方包
- 项目中不要使用相对路径引入包，而使用绝对路径   

##3.函数参数传递
- 参数名使用小写开头
- 少量数据使用对象，对于大量数据的struct使用指针
- 传入参数是map、slice和chan不要传递指针，因为map、slice和chan是引用类型，不需要传递指针

##4.错误处理
- 函数返回error是好习惯，使用多返回值 error， 不要使用 c 风格，返回错误码
- 调用函数时，必须首先对函数可能的错误进行处理
- 尽早判断和处理错误
- 尽量不要使用panic，只有在文件无法打开，数据库无法连接导致程序无法正常运行等特殊情况才考虑使用panic。


##5.代码格式化和分析工具
- gofmt
- goimports
- godoc
- go vet（静态分析一些明显的错误）
 
##6.注释##
  除了标注某些变量的含义，注释尽可能使用完整的句子       
- 包注释  
- 文件注释:包含作者，创建时间和简单功能描述   
- 关键类和对象注释  
- 关键函数注释  
- 关键变量含义注释  


##7.其它约定（待补充）
- 文件长度最好不要超过500行，每行最好不超过80字符
- 多返回值最多返回三个，超过三个请使用 struct
- golang的内置类型slice，map，chan都是引用，初次使用前，必须先用make分配好对象，不然会有空指针异常
- 尽量少用命名多返回值，如果要用，必须显示 return结果

#golang内存管理和垃圾回收

##1.为什么需要自主管理内存？
- 完成类似预分配、内存池等操作，以避开系统调用带来的性能问题
- 更好的配合垃圾回收

##1.内存分配器解决哪些问题？
- 内存碎片  
- 多核处理器能够扩展  

##2.如何分配定长记录？

##3.如何分配变长的记录？
取整导致“内存碎片”问题



##4.如何处理小对象<=32k？？
thread cache->central free list->central page allocator

##5.如何处理大对象>32k？？
直接从central page heap中分配。

##5.Central page Heap以span为管理对象。
span list。



##4.大对象如何分配？
分配对象时，大的对象直接分配 Span，小的对象从Span中分配


TCMalloc 是 Google 开发的内存分配器，在不少项目中都有使用，例如在 Golang 中就使用了类似的算法进行内存分配。它具有现代化内存分配器的基本特征：对抗内存碎片、在多核处理器能够 scale。据称，它的内存分配速度是 glibc2.3 中实现的 malloc的数倍。




##7.垃圾回收
- 标记清扫算法：标记阶段和清扫阶段  
- 精确垃圾回收  

golang使用标记清扫的垃圾回收算法，标记位图是非侵入式的

golang实现了精确的垃圾回收，在精确的垃圾回收中，先通过扫描整个内存块区域，定位对象的类型信息，得到该类型信息，得到其中的gc域。然后得到该类型中的垃圾回收的指令码，通过一个状态机解释这段指令码来执行特定类型的垃圾回收工作。

对于堆中的任意地址的对象，先通过它所在的内存页找到它所属的Mspan，然后通过MSpan中的类型信息找到它的类型信息。

golang并行垃圾回收

垃圾回收的触发是由一个gcpercent的变量控制的，当新分配的内存占已在使用中的内存的比例超过gcprecent时就会触发。比如，gcpercent=100，当前使用了4M的内存，那么当内存分配到达8M时就会再次gc。如果回收完毕后，内存的使用量为5M，那么下次回收的时机则是内存分配达到10M的时候。也就是说，并不是内存分配越多，垃圾回收频率越高，这个算法使得垃圾回收的频率比较稳定，适合应用的场景。



gcpercent的值是通过环境变量GOGC获取的，如果不设置这个环境变量，默认值是100。如果将它设置成off，则是关闭垃圾回收。



1.Go的垃圾回收机制在实践中有哪些需要注意的地方？？？
- 尽量不要创建大量的对象，也尽量不要频繁的创建对象  
- gc执行的时间跟数量是相关的
- 1、尽早的用memprof、cpuprof、GCTRACE来观察程序。 
- 2、关注请求处理时间，特别是开发新功能的时候，有助于发现设计上的问题。  
- 3、尽量避免频繁创建对象(&abc{}、new(abc{})、make())，在频繁调用的地方可以做对象重用。    
- 4、尽量不要用go管理大量对象，内存数据库可以完全用c实现好通过cgo来调用。

作者：达达
链接：https://www.zhihu.com/question/21615032/answer/18781477
来源：知乎
著作权归作者所有。商业转载请联系作者获得授权，非商业转载请注明出处。



Golang程序调优：memprof、cpuprof。



【1】.https://zhuanlan.zhihu.com/p/29216091  
【2】.http://goog-perftools.sourceforge.net/doc/tcmalloc.html  
【3】.https://cloud.tencent.com/developer/article/1072602
[4].http://www.opscoder.info/golang_gc.html

https://studygolang.com/articles/9389（现代垃圾回收）
https://segmentfault.com/a/1190000016828394

https://studygolang.com/articles/14497



[https://books.studygolang.com/The-Golang-Standard-Library-by-Example/chapter04/04.4.html](https://books.studygolang.com/The-Golang-Standard-Library-by-Example/chapter04/04.4.html)


1.日期和时间的格式问题

2.定时器
https://blog.csdn.net/u011304970/article/details/72724357

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

#《Golang编程 读书笔记》#

## 绪论 ##
**golang特点、优点：**

1.更加方面的并发编程
  从语言层面支持协程，编写并发程序简单；并发模型上（共享内存和消息传递），golang内置消息队列（channel），推崇消息传递模型。

2.软件工程方向所做的工作，代码风格相对统一，工程目录结构，错误处理方式相对统一。

3.golang编程哲学，融合众多语言的特点（支持oop、类、类成员、类成员函数、支持类的组合而不支持类的继承、支持匿名函数和包；放弃构造函数、析构函数、虚函数；支持接口类型，非侵入式的接口）


## 第一章 ##

1.2009年，以Ken.Thompsonwei为首设计和实现，谷歌开源。
语言特性（静态强类型）
1. 自动垃圾回收（c++忘记或多次释放堆内存，“野指针”问题，智能指针）
1. 内置类型（简单类型、高级类型（数组、字符串、字典类型map、数组切片（slice））
1. 自定义类型和接口（struct、接口相对松散）
1. 函数多返回值（这个厉害，可使用占位符）
1. 错误处理的规范（defer、panic和recover）
1. 匿名函数和闭包
1. 并发编程（协程、消息传递模型、channel、协程是如何调度的？？）
1. 反射？？（帮助快速获取对象的类型）
1. 语言与其它语言的交互性（Cgo）
1. 工程管理（目录结构、build、test、GDB调试）


2.Golang开发环境


实战一：实现一个简单的计算器功能呢，熟悉Golang的语法和工程管理、编译和调试过程？？？


## 第二章顺序编程 ##

1. 变量（声明、初始化、声明和初始化、支持多重赋值和匿名变量），由于Golang存在右值类型推导，使得有点类似动态动态语言，然而Golang是非常严格的静态强类型语言
2. 常量（const、itoa）、枚举
3. 类型（不同类型）
- bool类型只支持true和false，不支持0和1
- 整数类型（不同类型的整型数不同相互比较、运算）
- 浮点数（float32、float64，小数会被自动推到成float6、浮点数的比较）
- 字符串string是内置的基本类型，一旦初始化后不允许修改，当字符创中包含非ANSI的字符时，注意将源码的编码格式设置为UTF-8
- 字符类型byte
- 数组类型（固定长度；注意Golang中的数组是值类型，这意味着传参需要复制产生副本）
- 数组切片slice(动态变长数组，相当于vector，注意它的创建方式make)
- map类型（注意Golang中map是未排序的，而C++中的map是排序的；创建、赋值、删除、查找）


###流程控制：###
1. 条件语句，在有返回值的函数中，不允许将return语句包含在if...else...结构中。
2. 选择语句，在switch...case...结构中，不需要用break来明确退出一个case
3. 循环语句，不支持while和do..while结构,和循环相关结构全部使用for结构代替，包括“无限循环”等
4. Golang支持goto跳转结构

###函数：###
1. 函数的定义，"func 函数名（参数）（返回值）",相邻参数类型相同，可以合并
2. 函数名称的大小写，Golang中大写字母开头的函数能被其它包调用，小写字母开头的函数只能在本包内可见
3. Golang支持不定参数，同种类型的不定参数（args ...int）,任意类型的不定参数（args ...interface{}）
4. 函数多返回值，返回值可以命名也可以不命名，命名返回值会使得代码更加清晰
5. Golang支持随时在代码中定义匿名函数
6. Golang中的闭包，暂时不太理解其实际用途？？？？

###错误处理规范：###
1. error接口，nil，返回错误类型
2. 怎么自定义错误类型？？实现error接口，error接口只包含一个Error方法
3. defer错误，异常的延迟处理
4. panic和recover函数


***实战二***：实现排序算法，熟悉Golang中的类型、变量、流程逻辑、函数和错误处理规范


##第三章 Golang的面向对象特性##

1. Golang类型基于值传递，通过type关键字自定义类型，为类型添加方法
2. 深入理解值语义和引用语义，注意Golang中数组属于值语义
3. Golang放弃class和继承等大量面向对象的特性，只保留了struct和组合等基础特性
4. Golang中的面向对象在语言层面表现为结构体类型struct和类型的成员方法
5. 没有构造函数，通常由全局创建函数NewXXX来完成对象的构建
6. 对象的构造和初始化
7. Golang放弃了类的继承，支持类的组合方法
8. Golang中没有public、private等关键字，通过控制变量或者函数名称的大小写来控制可见性。Golang的可见性是包级别的，而不是类型级别的。
9. **Gloang的接口interface是类型系统的基石**
10. 侵入式的接口和非侵入式的接口，Golang非侵入式接口的优势。可以现有类型，再定义接口。一个类只需要实现接口所要求的所有函数，我们就说这个类实现了该接口
11. 接口赋值（对象实例给接口赋值，接口给接口赋值）
11. 接口查询，判断某一对象是否实现某个接口
11. Golang支持类型查询，后面也可利用反射功能进行类型查询
11. Golang支持多种接口的组合
11. Any类型与空接口，interface{}

***实战三***：熟悉Golang的面向对象的特性，设计并实现一个简单的播放器功能。


##第四章 Golang的并发编程##
**基于goroutine协程加channel的消息传递模型**

1. once类型，保证全局唯一一次性操作，处理某些函数只需要执行一次的情况
2. 互斥锁和读写锁；结合defer的经典锁使用模式
2. channel用于goroutine之间进行消息传递和数据数据共享；
3. channel是类型相关的，channel在定义的时候必须指明类型，一个channel只能传递一种类型的值
4. 多个同类型的channel构成channel数组
5. channel类似管道，分为数据的读和写，没有数据时，读操作会被阻塞，反之有数据时，写操作会被阻塞
6. channel的声明、定义与基本类型相似，在类型之前增加channel
7. channel的缓冲机制
8. channel的超时机制，避免永久死锁，select加超时等待goroutine
9. channel关闭和关闭的判断
10. Golang并发编程原则：协程同步优先考虑使用基于channel的消息通信机制完成，之后再考虑锁机制


实战四：棋牌游戏服务器样例代码


## 1.日志库 ##
初学编程者喜欢使用print等方法将程序运行的结果打印在终端上，然而这种方式在实际运行的服务中是不可行的，本文总结在实际工作中用到的日志库。


## 2.Glog使用简介 ##
C++标准库中没有提供日志库，glog是Google提供的一个应用程序的日志库，使用范围很广泛，值得学习。glog具有以下几个特点：  
- 提供基于C++风格的流的日志API  
- 默认情况下提供INFO、WARING、ERROR和FATAL四种等级的日志，和其它日志库类似，级别更高的日志会在同级别和所有低级别的日志文件中打印。
- 默认情况下glog日志输出到文件 /tmp/<program name>.<hostname>.<user name>.log.<severity level>.<date>-<time>.<pid> （比如 /tmp/hello_world.example.com.hamaji.log.INFO.20080709-222411.10474）。默认情况下，Glog对于 ERROR 和 FATAL 级别的日志会同时输出到stderr。
- log的主要输出方式就是上面两种，一种是向屏幕终端输出log信息，另一种将log信息写入了磁盘文件。



## 3.Glog中常用的标志位 ##
- logtostderr： 将log信息输出了stderr设备，而不将log信息写入到log文件。
- stderrthreshold：将log信息输出到stderr，并且将log信息写入到log文件，minloglevel设置最小的log等级，比如该标志位设置为2，那么只会显示和记录以及FATAL这两种类型的log信息
- log_dir:设置log文件的存储路径
- v：类似于民loglevel，用来设置最大的显示等级， 需要结合VLOG宏一起使用
- vmodule：标志位v的扩充，可以分别设置不同文件的log信息类别  
**在标志位前面加上FLAGS_字符串，即可设置相应的标识：** 例如FLAGS_logtostderr=1

参考：phxpaxos中glog的使用


###golang的日志库####



https://www.processon.com/ 在线制图


ubuntu磁盘空间不足，如何扩容？？
https://blog.csdn.net/u011345885/article/details/73060897

32.无限试用Navicat for MySQL？？
https://my.oschina.net/u/3509764/blog/910748
宿主机 Navicat 连接VMware Ubuntu 虚拟机 的MySQL 实现方法
https://blog.csdn.net/qq_34256348/article/details/78358678






http://www.runoob.com/redis/redis-tutorial.html

##1.Redis简介
- 主要用于key-value缓存
- 支持数据持久化，将内存数据保存在磁盘中，重启后再次加载使用
- 值（value）类型丰富，支持字符串（String）、哈希（Map）、列表（list）、集合（sets）和有序集合（sorted sets）
- 原子性，Redis的所有操作都是原子性的，意思就是要么成功执行要么失败完全不执行。单个操作是原子性的。多个操作也支持事务，即原子性，通过MULTI和EXEC指令包起来


##2.Redis安装和使用
	ubuntu直接使用命令行安装  
	`$sudo apt-get update`  
	`$sudo apt-get install redis-server`  
	启动服务器：`$redis-server`  
	启动命令行客户端：`redis-cli`  

##3.五种数据类型
1. String（字符串）：二进制安全，可以包含任何数据，比如图片或者序列化的对象，一个键最多存储512MB
2. Hash（字典）：键值对集合，编程语言中的map，例如用于存储、读取、修改用户属性
3. List（列表）：链表（双向链表），例如用于记录朋友圈的时间线数据
4. Set（集合）：哈希表实现，常用于求交集，例如共同好友等应用
5. zset(sorted set:有序集合)：数据插入时已经排序，常用于排行榜和带有权重的消息队列



##4.常用命令


##5.C++使用Redis
C++使用redis用hiredis完成，hiredis使用c语言编写。
官网：[https://github.com/redis/hiredis](https://github.com/redis/hiredis)




##6.Golang使用Redis
[https://github.com/go-redis/redis](https://github.com/go-redis/redis)


书籍
《图解http》
