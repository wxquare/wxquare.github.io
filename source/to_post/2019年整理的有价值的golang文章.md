
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