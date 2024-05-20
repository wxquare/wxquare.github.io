---
title: 一文记录计算机网络
date: 2023-09-13
categories: 
- 计算机基础
---


## TCP和UDP协议
1. [tcp头格式，其20个字节包含哪些内容？](https://www.cnblogs.com/xiaolincoding/p/12638546.html) udp头部格式，其8个字节分别包含哪些内容？ 
2. 为什么 UDP 头部没有「首部长度」字段，而 TCP 头部有「首部长度」字段呢？原因是 TCP 有可变长的「选项」字段，而 UDP 头部长度则是不会变化的，无需多一个字段去记录 UDP 的首部长度
3. **tcp和udp的区别以及应用场景**
	- TCP是面向连接的，而UDP是不需要建立连接的
	- TCP 是一对一的两点服务，UDP 支持一对一、一对多、多对多的交互通信
	- 可靠性，TCP 是可靠交付数据的，数据可以无差错、不丢失、不重复、按需到达。UDP 是尽最大努力交付，不保证可靠交付数据。
	- TCP有拥塞控制、流量控制
	- 首部开销，TCP 首部长度较长，会有一定的开销，首部在没有使用「选项」字段时是 20 个字节，如果使用了「选项」字段则会变长的。UDP 首部只有 8 个字节，并且是固定不变的，开销较小。
	- 传输方式，TCP 是流式传输，没有边界，但保证顺序和可靠。UDP 是一个包一个包的发送，是有边界的，但可能会丢包和乱序
   
   TCP 和 UDP 应用场景：由于 TCP 是面向连接，能保证数据的可靠性交付，因此经常用于，FTP 文件传输HTTP / HTTPS，由于 UDP 面向无连接，它可以随时发送数据，再加上UDP本身的处理既简单又高效，因此经常用于：包总量较少的通信，如 DNS 、SNMP 等视频、音频等多媒体通信广播通信
4. **TCP协议如何保证可靠传输？**
	- 三次握手四次挥手确保连接的建立和释放
	- 超时重发：数据切块发送，等待确认，超时未确认会重发
	- 数据完整性校验：TCP首部中数据有端到端的校验和，接收方会校验，一旦出错将丢弃且不确认收到此报文
	- 根据序列码进行数据的排序和去重
	- 根据接收端缓冲区大小做流量控制
	- 根据网络环境做拥塞控制。当网络拥塞时，会减少数据的发送
	
5. **TCP怎么通过三次握手和四次挥手建立可靠连接以及需要注意的问题**
	- [分别准确画出三次握手和四次挥手状态转换图](https://www.cnblogs.com/xiaolincoding/p/12638546.html) 从上面的过程可以发现第三次握手是可以携带数据的，前两次握手是不可以携带数据的，这也是面试常问的题
	- 为什么需要三次握手？ 通过三次握手实现了同步序列号和避免了旧的重复连接初始化造成混乱，浪费服务器资源，两个作用
	- 为什么需要四次挥手？全双工通信
	- time_wait状态什么作用？ 防止之前的报文造成新连接数据混乱，通过2msl使前一连接数据失效；确保ack报文发送给服务端。
	
6. **超时重传和快速重传**
	- 客户端通过定时器在指定时间内未发现会收到ack信息就认为进行超时重传
	- 客户端收到连续三个重复ack信息就会发起快速重传而不用等待超时重传
	
7. **如何解决可能出现的乱序和重复数据问题**
	- 三次握手双方约定ISN
	- [TCP建立链接时ISN是怎么产生的，为什么需要每次都不相同？](https://www.cnblogs.com/xiaolincoding/p/12638546.html)
	- 根据序列号调整顺序
	
8. **[TCP流量控制和滑动窗口](https://www.cnblogs.com/xiaolincoding/p/12732052.html)**
	- 为了提高数据传输的小路，tcp避免了一问一答式的消息传输策略
	- 通过累积确认ACK的方式提高效率
	- 在累积确认时通过接收窗口进行流量控制	
	
9. **tcp拥塞控制和拥塞窗口？**
   ![TCP拥塞控制](/images/tcp-network-congestion.jpg)
	- tcp在数据发送时会结合整个网络环境调整数据发送的速率
	- 发送者如何判断拥塞已经发生的？发送超时，或者说TCP重传定时器溢出；接收到重复的确认报文段
	- 快重传算法（接收端到失序的报文段立即重传、发送端一旦接收三个重复的确认报文段，立即重传，不用等定时器）


10. TCP 的连接状态查看，在 Linux 可以通过 netstat -napt 命令查看
11. 什么是SYN攻击，怎么避免SYN攻击？ 
- SYN攻击属于DOS攻击的一种，它利用TCP协议缺陷，通过发送大量的半连接请求，耗费CPU和内存资源。SYN攻击除了能影响主机外，还可以危害路由器、防火墙等网络系统，事实上SYN攻击并不管目标是什么系统，只要这些系统打开TCP服务就可以实施。从上图可看到，服务器接收到连接请求（syn=j），将此信息加入未连接队列，并发送请求包给客户（syn=k,ack=j+1），此时进入SYN_RECV状态。当服务器未收到客户端的确认包时，重发请求包，一直到超时，才将此条目从未连接队列删除。配合IP欺骗，SYN攻击能达到很好的效果，通常，客户端在短时间内伪造大量不存在的IP地址，向服务器不断地发送syn包，服务器回复确认包，并等待客户的确认，由于源地址是不存在的，服务器需要不断的重发直至超时，这些伪造的SYN包将长时间占用未连接队列，正常的SYN请求被丢弃，目标系统运行缓慢，严重者引起网络堵塞甚至系统瘫痪。
12. 如何解决close_wait和time_wait过多的问题？
	- CLOSE_WAIT，只会发生在客户端先关闭连接的时候，但已经收到客户端的fin包，但服务器还没有关闭的时候会产生这个状态，如果服务器产生大量的这种连接一般是程序问题导致的，如部分情况下不会执行socket的close方法，解决方法是查程序
	- TIME_WAIT，time_wait是一个需要特别注意的状态，他本身是一个正常的状态，只在主动断开那方出现，每次tcp主动断开都会有这个状态的，维持这个状态的时间是2个msl周期（2分钟），设计这个状态的目的是为了防止我发了ack包对方没有收到可以重发。那如何解决出现大量的time_wait连接呢？千万不要把tcp_tw_recycle改成1，这个我再后面介绍，正确的姿势应该是降低msl周期，也就是tcp_fin_timeout值，同时增加time_wait的队列（tcp_max_tw_buckets），防止满了。
	
13. 什么是TCP粘包，应用层怎么解决，http是怎么解决的。tcp是字节流，需要根据特殊字符和长度信息将消息分开

14. **udp协议怎么做可靠传输？**
	由于在传输层UDP已经是不可靠的连接，那就要在应用层自己实现一些保障可靠传输的机制，简单来讲，要使用UDP来构建可靠的面向连接的数据传输，就要实现类似于TCP协议的，超时重传（定时器），有序接受 （添加包序号），应答确认 （Seq/Ack应答机制），滑动窗口流量控制等机制 （滑动窗口协议），等于说要在传输层的上一层（或者直接在应用层）实现TCP协议的可靠数据传输机制，比如使用UDP数据包+序列号，UDP数据包+时间戳等方法。目前已经有一些实现UDP可靠传输的机制，比如UDT（UDP-based Data Transfer Protocol）基于UDP的数据传输协议（UDP-based Data Transfer Protocol，简称UDT）是一种互联网数据传输协议。UDT的主要目的是支持高速广域网上的海量数据传输，而互联网上的标准数据传输协议TCP在高带宽长距离网络上性能很差。 顾名思义，UDT建于UDP之上，并引入新的拥塞控制和数据可靠性控制机制。UDT是面向连接的双向的应用层协议。它同时支持可靠的数据流传输和部分可靠的数据报传输。 由于UDT完全在UDP上实现，它也可以应用在除了高速数据传输之外的其它应用领域，例如点到点技术（P2P），防火墙穿透，多媒体数据传输等等
	
14. **TCP 保活机制KeepAlive？其局限性？Http的keep-alive？为什么应用层也经常做心跳检查？**
	- TCP KeepAlive 的基本原理是，隔一段时间给连接对端发送一个探测包，如果收到对方回应的 ACK，则认为连接还是存活的，在超过一定重试次数之后还是没有收到对方的回应，则丢弃该 TCP 连接。TCP-Keepalive-HOWTO 有对 TCP KeepAlive 特性的详细介绍，有兴趣的同学可以参考。
	- TCP KeepAlive 的局限。首先 TCP KeepAlive 监测的方式是发送一个 probe 包，会给网络带来额外的流量，另外 TCP KeepAlive 只能在内核层级监测连接的存活与否，而连接的存活不一定代表服务的可用。例如当一个服务器 CPU 进程服务器占用达到 100%，已经卡死不能响应请求了，此时 TCP KeepAlive 依然会认为连接是存活的。因此 TCP KeepAlive 对于应用层程序的价值是相对较小的。需要做连接保活的应用层程序，例如 QQ，往往会在应用层实现自己的心跳功能。
除了TCP自带的Keeplive机制，实现业务中经常在业务层面定制**“心跳”**功能，主要有以下几点考虑：
	- TCP自带的keepalive使用简单，仅提供连接是否存活的功能  
	- 应用层心跳包不依赖于传输协议，支持tcp和udp  
	- 应用层心跳包可以定制，可以应对更加复杂的情况或者传输一些额外的消息  
	- Keepalive仅仅代表连接保持着，而心跳往往还表示服务正常工作
在 HTTP 1.0 时期，每个 TCP 连接只会被一个 HTTP Transaction（请求加响应）使用，请求时建立，请求完成释放连接。当网页内容越来越复杂，包含大量图片、CSS 等资源之后，这种模式效率就显得太低了。所以，在 HTTP 1.1 中，引入了 HTTP persistent connection 的概念，也称为 HTTP keep-alive，目的是复用TCP连接，在一个TCP连接上进行多次的HTTP请求从而提高性能。HTTP1.0中默认是关闭的，需要在HTTP头加入"Connection: Keep-Alive"，才能启用Keep-Alive；HTTP1.1中默认启用Keep-Alive，加入"Connection: close "，才关闭。两者在写法上不同，http keep-alive 中间有个"-"符号。 **HTTP协议的keep-alive 意图在于连接复用**，同一个连接上串行方式传递请求-响应数据。**TCP的keepalive机制意图在于保活、心跳，检测连接错误。**


15. [TCP 协议性能问题分析？](https://draveness.me/whys-the-design-tcp-performance/)
	- TCP 的拥塞控制在发生丢包时会进行退让，减少能够发送的数据段数量，但是丢包并不一定意味着网络拥塞，更多的可能是网络状况较差；
	- TCP 的三次握手带来了额外开销，这些开销不只包括需要传输更多的数据，还增加了首次传输数据的网络延迟；
	- TCP 的重传机制在数据包丢失时可能会重新传输已经成功接收的数据段，造成带宽的浪费；

16. [QUIC 是如何解决TCP 性能瓶颈的？](https://blog.csdn.net/m0_37621078/article/details/106506532)
17. [科普：QUIC协议原理分析](https://zhuanlan.zhihu.com/p/32553477)
18. 

## http和https
1. [HTTP协议协议格式详解](https://www.jianshu.com/p/8fe93a14754c)
    - 请求行(request line)。请求方法、域名、协议版本。
    - 请求头部(header)从第二行起为请求头部，Host指出请求的目的地（主机域名）；User-Agent是客户端的信息，它是检测浏览器类型的重要信息，由浏览器定义，并且在每个请求中自动发送
    - 空行
    - 请求数据
2. http 常见的状态码有哪些？
	- 200 成功
	- 3xx重定向相关，301 永久重定向，302临时重定向
	- 4xx客户端错误，400请求报文有问题，403服务器禁止访问资源,404资源不存在
	- 5xx服务器内部错误,501 请求的功能暂不支持，502 服务器逻辑有问题，503 服务器繁忙
3. get 和 post 区别
	- GET参数通过URL传递，POST放在Request body中
	- GET请求只能进行url编码，而POST支持多种编码方式
	- GET请求在URL中传送的参数是有长度限制的，而POST没有
	- GET比POST更不安全，因为参数直接暴露在URL上，所以不能用来传递敏感信息。
	- GET请求参数会被完整保留在浏览器历史记录里，而POST中的参数不会被保留。
4. [https的工作原理和流程](https://segmentfault.com/a/1190000021494676)

5. http和https的区别
	- http采用明文传输，http+ssl的加密传输
	- http是80端口，https是443端口
	- HTTP的连接很简单，是无状态的；HTTPS协议是由SSL+HTTP协议构建的可进行加密传输、身份认证的网络协议，比HTTP协议安全

6. [浏览器输入http://www.baidu.com](https://www.nowcoder.com/questionTerminal/f09d6db0077d4731ac5b34607d4431ee)
	事件顺序
	(1) 浏览器获取输入的域名www.baidu.com
	(2) 浏览器向DNS请求解析www.baidu.com的IP地址
	(3) 域名系统DNS解析出百度服务器的IP地址
	(4) 浏览器与该服务器建立TCP连接(默认端口号80)
	(5) 浏览器发出HTTP请求，请求百度首页
	(6) 服务器通过HTTP响应把首页文件发送给浏览器
	(7) TCP连接释放
	(8) 浏览器将首页文件进行解析，并将Web页显示给用户。

7. http长连接和短连接？http长连接和短连接以及keep-Alive的含义，HTTP 长连接不可能一直保持，例如 Keep-Alive: timeout=5, max=100，表示这个TCP通道可以保持5秒，max=100，表示这个长连接最多接收100次请求就断开。
8. http cookie和session
	- Cookie和Session都是客户端与服务器之间保持状态的解决方案，具体来说，cookie机制采用的是在客户端保持状态的方案，而session机制采用的是在服务器端保持状态的方案
	- Cookie实际上是一小段的文本信息。客户端请求服务器，如果服务器需要记录该用户状态，就使用response向客户端浏览器颁发一个Cookie，而客户端浏览器会把Cookie保存起来。当浏览器再请求该网站时，浏览器把请求的网址连同该Cookie一同提交给服务器，服务器检查该Cookie，以此来辨认用户状态。服务器还可以根据需要修改Cookie的内容
9. http1.0,tttp1.1,http2.0,http 3.0各有什么变化
	- http 1.0
	- http 1.1, 长连接
	- http 2.0，二进制压缩+连接复用
	- http QUIC，udp+ssl
10. [HTTP/3 竟然基于 UDP，HTTP 协议这些年都经历了啥？](https://zhuanlan.zhihu.com/p/68012355)
10. 使用curl
11. [https中间人攻击原理以及防御措施](https://zh.wikipedia.org/wiki/%E4%B8%AD%E9%97%B4%E4%BA%BA%E6%94%BB%E5%87%BB)
12. [如何理解http的无连接和无状态的特点？](https://blog.csdn.net/tennysonsky/article/details/44562435)
13. [半链接和Sync 攻击原理及防范技术](https://www.cnblogs.com/mafeng/p/7615230.html)



<p align="center">
  <img src="/images/5KeocQs.jpg">
  <br/>
  <strong><a href=http://www.escotal.com/osilayer.html>资料来源：OSI 7层模型</a></strong>
</p>

### 超文本传输协议（HTTPS/HTTP1.1/HTTP2/HTTP3）
https://aws.amazon.com/cn/compare/the-difference-between-https-and-http/

HTTP 是一种在客户端和服务器之间编码和传输数据的方法。它是一个请求/响应协议：客户端和服务端针对相关内容和完成状态信息的请求和响应。HTTP 是独立的，允许请求和响应流经许多执行负载均衡，缓存，加密和压缩的中间路由器和服务器。

一个基本的 HTTP 请求由一个动词（方法）和一个资源（端点）组成。 以下是常见的 HTTP 动词：

| 动词     | 描述             | *幂等  | 安全性  | 可缓存            |
| ------ | -------------- | ---- | ---- | -------------- |
| GET    | 读取资源           | Yes  | Yes  | Yes            |
| POST   | 创建资源或触发处理数据的进程 | No   | No   | Yes，如果回应包含刷新信息 |
| PUT    | 创建或替换资源        | Yes  | No   | No             |
| DELETE | 删除资源           | Yes  | No   | No             |


<p align="center">
  <img src="/images/http.png" width=500 height=100>
</p>

- HTTPS 是基于 HTTP 的安全版本，通过使用 SSL 或 TLS 加密和身份验证通信。
- HTTP/1.1 是 HTTP 的第一个主要版本，引入了持久连接、管道化请求等特性。
- HTTP/2 是 HTTP 的第二个主要版本，使用二进制协议，引入了多路复用、头部压缩、服务器推送等特性。
- HTTP/3 是 HTTP 的第三个主要版本，基于 QUIC 协议，使用 UDP，提供更快的传输速度和更好的性能


**多次执行不会产生不同的结果**。

HTTP 是依赖于较低级协议（如 **TCP** 和 **UDP**）的应用层协议。

#### 来源及延伸阅读：HTTP

* [README](https://www.quora.com/What-is-the-difference-between-HTTP-protocol-and-TCP-protocol)    +
* [HTTP 是什么？](https://www.nginx.com/resources/glossary/http/)
* [HTTP 和 TCP 的区别](https://www.quora.com/What-is-the-difference-between-HTTP-protocol-and-TCP-protocol)
* [PUT 和 PATCH的区别](https://laracasts.com/discuss/channels/general-discussion/whats-the-differences-between-put-and-patch?page=1)

### 传输控制协议（TCP）

<p align="center">
  <img src="/images/JdAsdvG.jpg">
  <br/>
  <strong><a href="http://www.wildbunny.co.uk/blog/2012/10/09/how-to-make-a-multi-player-game-part-1/">资料来源：如何制作多人游戏</a></strong>
</p>

TCP 是通过 [IP 网络](https://en.wikipedia.org/wiki/Internet_Protocol)的面向连接的协议。 使用[握手](https://en.wikipedia.org/wiki/Handshaking)建立和断开连接。 发送的所有数据包保证以原始顺序到达目的地，用以下措施保证数据包不被损坏：

- 每个数据包的序列号和[校验码](https://en.wikipedia.org/wiki/Transmission_Control_Protocol#Checksum_computation)。
- [确认包](https://en.wikipedia.org/wiki/Acknowledgement_(data_networks))和自动重传

如果发送者没有收到正确的响应，它将重新发送数据包。如果多次超时，连接就会断开。TCP 实行[流量控制](https://en.wikipedia.org/wiki/Flow_control_(data))和[拥塞控制](https://en.wikipedia.org/wiki/Network_congestion#Congestion_control)。这些确保措施会导致延迟，而且通常导致传输效率比 UDP 低。

为了确保高吞吐量，Web 服务器可以保持大量的 TCP 连接，从而导致高内存使用。在 Web 服务器线程间拥有大量开放连接可能开销巨大，消耗资源过多，也就是说，一个 [memcached](#memcached) 服务器。[连接池](https://en.wikipedia.org/wiki/Connection_pool) 可以帮助除了在适用的情况下切换到 UDP。

TCP  对于需要高可靠性但时间紧迫的应用程序很有用。比如包括 Web 服务器，数据库信息，SMTP，FTP 和 SSH。

以下情况使用 TCP 代替 UDP：

- 你需要数据完好无损。
- 你想对网络吞吐量自动进行最佳评估。

### 用户数据报协议（UDP）

<p align="center">
  <img src="/images/yzDrJtA.jpg">
  <br/>
  <strong><a href="http://www.wildbunny.co.uk/blog/2012/10/09/how-to-make-a-multi-player-game-part-1">资料来源：如何制作多人游戏</a></strong>
</p>

UDP 是无连接的。数据报（类似于数据包）只在数据报级别有保证。数据报可能会无序的到达目的地，也有可能会遗失。UDP 不支持拥塞控制。虽然不如 TCP 那样有保证，但 UDP 通常效率更高。

UDP 可以通过广播将数据报发送至子网内的所有设备。这对 [DHCP](https://en.wikipedia.org/wiki/Dynamic_Host_Configuration_Protocol) 很有用，因为子网内的设备还没有分配 IP 地址，而 IP 对于 TCP 是必须的。

UDP 可靠性更低但适合用在网络电话、视频聊天，流媒体和实时多人游戏上。

以下情况使用 UDP 代替 TCP：

* 你需要低延迟
* 相对于数据丢失更糟的是数据延迟
* 你想实现自己的错误校正方法

#### 来源及延伸阅读：TCP 与 UDP

* [游戏编程的网络](http://gafferongames.com/networking-for-game-programmers/udp-vs-tcp/)
* [TCP 与 UDP 的关键区别](http://www.cyberciti.biz/faq/key-differences-between-tcp-and-udp-protocols/)
* [TCP 与 UDP 的不同](http://stackoverflow.com/questions/5970383/difference-between-tcp-and-udp)
* [传输控制协议](https://en.wikipedia.org/wiki/Transmission_Control_Protocol)
* [用户数据报协议](https://en.wikipedia.org/wiki/User_Datagram_Protocol)
* [Memcache 在 Facebook 的扩展](http://www.cs.bu.edu/~jappavoo/jappavoo.github.com/451/papers/memcache-fb.pdf)

### 远程过程调用协议（RPC）

<p align="center">
  <img src="/images/iF4Mkb5.png">
  <br/>
  <strong><a href="http://www.puncsky.com/blog/2016/02/14/crack-the-system-design-interview">Source: Crack the system design interview</a></strong>
</p>

在 RPC 中，客户端会去调用另一个地址空间（通常是一个远程服务器）里的方法。调用代码看起来就像是调用的是一个本地方法，客户端和服务器交互的具体过程被抽象。远程调用相对于本地调用一般较慢而且可靠性更差，因此区分两者是有帮助的。热门的 RPC 框架包括 [Protobuf](https://developers.google.com/protocol-buffers/)、[Thrift](https://thrift.apache.org/) 和 [Avro](https://avro.apache.org/docs/current/)。

RPC 是一个“请求-响应”协议：

* **客户端程序** ── 调用客户端存根程序。就像调用本地方法一样，参数会被压入栈中。
* **客户端 stub 程序** ── 将请求过程的 id 和参数打包进请求信息中。
* **客户端通信模块** ── 将信息从客户端发送至服务端。
* **服务端通信模块** ── 将接受的包传给服务端存根程序。
* **服务端 stub 程序** ── 将结果解包，依据过程 id 调用服务端方法并将参数传递过去。

RPC 调用示例：

```
GET /someoperation?data=anId

POST /anotheroperation
{
  "data":"anId";
  "anotherdata": "another value"
}
```

RPC 专注于暴露方法。RPC 通常用于处理内部通讯的性能问题，这样你可以手动处理本地调用以更好的适应你的情况。


当以下情况时选择本地库（也就是 SDK）：

* 你知道你的目标平台。
* 你想控制如何访问你的“逻辑”。
* 你想对发生在你的库中的错误进行控制。
* 性能和终端用户体验是你最关心的事。

遵循 **REST** 的 HTTP API 往往更适用于公共 API。

#### 缺点：RPC

* RPC 客户端与服务实现捆绑地很紧密。
* 一个新的 API 必须在每一个操作或者用例中定义。
* RPC 很难调试。
* 你可能没办法很方便的去修改现有的技术。举个例子，如果你希望在 [Squid](http://www.squid-cache.org/) 这样的缓存服务器上确保 [RPC 被正确缓存](http://etherealbits.com/2012/12/debunking-the-myths-of-rpc-rest/)的话可能需要一些额外的努力了。

### 表述性状态转移（REST）

REST 是一种强制的客户端/服务端架构设计模型，客户端基于服务端管理的一系列资源操作。服务端提供修改或获取资源的接口。所有的通信必须是无状态和可缓存的。

RESTful 接口有四条规则：

* **标志资源（HTTP 里的 URI）** ── 无论什么操作都使用同一个 URI。
* **表示的改变（HTTP 的动作）** ── 使用动作, headers 和 body。
* **可自我描述的错误信息（HTTP 中的 status code）** ── 使用状态码，不要重新造轮子。
* **[HATEOAS](http://restcookbook.com/Basics/hateoas/)（HTTP 中的HTML 接口）** ── 你的 web 服务器应该能够通过浏览器访问。

REST 请求的例子：

```
GET /someresources/anId

PUT /someresources/anId
{"anotherdata": "another value"}
```

REST 关注于暴露数据。它减少了客户端／服务端的耦合程度，经常用于公共 HTTP API 接口设计。REST 使用更通常与规范化的方法来通过 URI 暴露资源，[通过 header 来表述](https://github.com/for-GET/know-your-http-well/blob/master/headers.md)并通过 GET、POST、PUT、DELETE 和 PATCH 这些动作来进行操作。因为无状态的特性，REST 易于横向扩展和隔离。

#### 缺点：REST

* 由于 REST 将重点放在暴露数据，所以当资源不是自然组织的或者结构复杂的时候它可能无法很好的适应。举个例子，返回过去一小时中与特定事件集匹配的更新记录这种操作就很难表示为路径。使用 REST，可能会使用 URI 路径，查询参数和可能的请求体来实现。
* REST 一般依赖几个动作（GET、POST、PUT、DELETE 和 PATCH），但有时候仅仅这些没法满足你的需要。举个例子，将过期的文档移动到归档文件夹里去，这样的操作可能没法简单的用上面这几个 verbs 表达。
* 为了渲染单个页面，获取被嵌套在层级结构中的复杂资源需要客户端，服务器之间多次往返通信。例如，获取博客内容及其关联评论。对于使用不确定网络环境的移动应用来说，这些多次往返通信是非常麻烦的。
* 随着时间的推移，更多的字段可能会被添加到 API 响应中，较旧的客户端将会接收到所有新的数据字段，即使是那些它们不需要的字段，结果它会增加负载大小并引起更大的延迟。

### RPC 与 REST 比较

| 操作          | RPC                                      | REST                                     |
| ----------- | ---------------------------------------- | ---------------------------------------- |
| 注册          | **POST** /signup                         | **POST** /persons                        |
| 注销          | **POST** /resign<br/>{<br/>"personid": "1234"<br/>} | **DELETE** /persons/1234                 |
| 读取用户信息      | **GET** /readPerson?personid=1234        | **GET** /persons/1234                    |
| 读取用户物品列表    | **GET** /readUsersItemsList?personid=1234 | **GET** /persons/1234/items              |
| 向用户物品列表添加一项 | **POST** /addItemToUsersItemsList<br/>{<br/>"personid": "1234";<br/>"itemid": "456"<br/>} | **POST** /persons/1234/items<br/>{<br/>"itemid": "456"<br/>} |
| 更新一个物品      | **POST** /modifyItem<br/>{<br/>"itemid": "456";<br/>"key": "value"<br/>} | **PUT** /items/456<br/>{<br/>"key": "value"<br/>} |
| 删除一个物品      | **POST** /removeItem<br/>{<br/>"itemid": "456"<br/>} | **DELETE** /items/456                    |

<p align="center">
  <strong><a href="https://apihandyman.io/do-you-really-know-why-you-prefer-rest-over-rpc">资料来源：你真的知道你为什么更喜欢 REST 而不是 RPC 吗</a></strong>
</p>

## 其它
1. https://blog.csdn.net/justloveyou_/article/details/78303617
2. 图解https的过程:https://segmentfault.com/a/1190000021494676
3. [35 张图解：被问千百遍的 TCP 三次握手和四次挥手面试题](https://www.cnblogs.com/xiaolincoding/p/12638546.html)
4. [30张图解： TCP 重传、滑动窗口、流量控制、拥塞控制](https://www.cnblogs.com/xiaolincoding/p/12732052.html)
5. [硬核！30 张图解 HTTP 常见的面试题](https://www.cnblogs.com/xiaolincoding/p/12442435.html)
