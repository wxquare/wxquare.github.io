---
title: 互联网系统设计 - DNS、负责均衡器和反向代理
date: 2024-02-02
categories:
- 系统设计
---


## 负载均衡和高可用的区别
“Load Balancing”（负载均衡）和“High Availability”（高可用性）是两个重要的概念，它们在系统设计和架构中有不同的侧重点。以下是它们的主要区别：
### 负载均衡（Load Balancing）
1. **定义**: 负载均衡是将流量或请求分配到多个服务器或资源上的技术，以确保没有单个服务器过载，从而提高性能和响应速度。主要目的是优化资源使用，减少响应时间，提高吞吐量。通过分散负载，系统可以处理更多的请求。
2. **实现方式**: 负载均衡可以通过硬件负载均衡器或软件负载均衡器（如 Nginx、HAProxy 等）来实现。它可以根据多种策略进行流量分配，如轮询、最少连接、加权等。
3. **场景**: 常用于需要处理大量并发请求的应用，如网站、API 服务等。
### 高可用性（High Availability）
1. **定义**: 高可用性是指系统在一定时间内保持正常运行和可用的能力，通常通过冗余和故障转移机制来实现。主要目的是确保系统在硬件故障、软件故障或其他问题发生时仍能继续服务，降低停机时间。
2. **实现方式**: 高可用性通常通过冗余（如多个服务器、数据中心等）、故障检测和自动切换等机制来实现。常见的高可用性解决方案包括集群、主从复制等。
3. **场景**: 适用于对可用性要求极高的应用，如金融服务、医疗系统等。
### 总结
- **负载均衡** 关注的是如何有效分配流量和资源，以提高性能。
- **高可用性** 关注的是如何确保系统在故障时仍能保持运行。

这两者可以结合使用，负载均衡可以在高可用性架构中发挥重要作用，确保在多个冗余实例之间分配请求，从而提高整体的可用性和性能。


## 网关整体架构

<p align="center">
  <img src="/images/load_balanging.png" width=800 height=400>
  <br/>
</p>



## 网络负载均衡器（L4）
- https://www.cnblogs.com/heyongshen/p/16827111.html



## 七层负载均衡器


## 负载均衡器和反向代理
<p align="center">
  <img src="/images/load_balanging.png" width=800 height=400>
  <br/>
</p>

<p align="center">
  <img src="/images/h81n9iK.png" width=600 height=400>
  <br/>
  <strong><a href="http://horicky.blogspot.com/2010/10/scalable-system-design-patterns.html">来源：可扩展的系统设计模式</a></strong>
</p>

负载均衡器将传入的请求分发到应用服务器和数据库等计算资源。无论哪种情况，负载均衡器将从计算资源来的响应返回给恰当的客户端。负载均衡器的效用在于:

* 防止请求进入不好的服务器
* 防止资源过载
* 帮助消除单一的故障点
* **SSL 终结** ─ 解密传入的请求并加密服务器响应，这样的话后端服务器就不必再执行这些潜在高消耗运算了。
* 不需要再每台服务器上安装 [X.509 证书](https://en.wikipedia.org/wiki/X.509)。
* **Session 留存** ─ 如果 Web 应用程序不追踪会话，发出 cookie 并将特定客户端的请求路由到同一实例。
* 通常会设置采用[工作─备用](#工作到备用切换active-passive) 或 [双工作](#双工作切换active-active) 模式的多个负载均衡器，以免发生故障。

负载均衡器能基于多种方式来路由流量:
* 随机
* 最少负载
* Session/cookie
* [轮询调度或加权轮询调度算法](http://g33kinfo.com/info/archives/2657)
* [四层负载均衡](#四层负载均衡)
* [七层负载均衡](#七层负载均衡)



### 四层负载均衡
四层负载均衡根据监看[传输层](#通讯)的信息来决定如何分发请求。通常，这会涉及来源，目标 IP 地址和请求头中的端口，但不包括数据包（报文）内容。四层负载均衡执行[网络地址转换（NAT）](https://www.nginx.com/resources/glossary/layer-4-load-balancing/)来向上游服务器转发网络数据包。


### 七层负载均衡器

七层负载均衡器根据监控[应用层](#通讯)来决定怎样分发请求。这会涉及请求头的内容，消息和 cookie。七层负载均衡器终结网络流量，读取消息，做出负载均衡判定，然后传送给特定服务器。比如，一个七层负载均衡器能直接将视频流量连接到托管视频的服务器，同时将更敏感的用户账单流量引导到安全性更强的服务器。

以损失灵活性为代价，四层负载均衡比七层负载均衡花费更少时间和计算资源，虽然这对现代商用硬件的性能影响甚微。


### 水平扩展

负载均衡器还能帮助水平扩展，提高性能和可用性。使用商业硬件的性价比更高，并且比在单台硬件上**垂直扩展**更贵的硬件具有更高的可用性。相比招聘特定企业系统人才，招聘商业硬件方面的人才更加容易。

### 缺陷：水平扩展
* 水平扩展引入了复杂度并涉及服务器复制
* 服务器应该是无状态的:它们也不该包含像 session 或资料图片等与用户关联的数据。
* session 可以集中存储在数据库或持久化[缓存](#缓存)（Redis、Memcached）的数据存储区中。
* 缓存和数据库等下游服务器需要随着上游服务器进行扩展，以处理更多的并发连接。

### 缺陷：负载均衡器
* 如果没有足够的资源配置或配置错误，负载均衡器会变成一个性能瓶颈。
* 引入负载均衡器以帮助消除单点故障但导致了额外的复杂性。
* 单个负载均衡器会导致单点故障，但配置多个负载均衡器会进一步增加复杂性。


### 反向代理（web 服务器）
<p align="center">
  <img src="/images/n41Azff.png">
  <br/>
  <strong><a href="https://upload.wikimedia.org/wikipedia/commons/6/67/Reverse_proxy_h2g2bob.svg">资料来源：维基百科</a></strong>
  <br/>
</p>

反向代理是一种可以集中地调用内部服务，并提供统一接口给公共客户的 web 服务器。来自客户端的请求先被反向代理服务器转发到可响应请求的服务器，然后代理再把服务器的响应结果返回给客户端。

带来的好处包括：
- **增加安全性** - 隐藏后端服务器的信息，屏蔽黑名单中的 IP，限制每个客户端的连接数。
- **提高可扩展性和灵活性** - 客户端只能看到反向代理服务器的 IP，这使你可以增减服务器或者修改它们的配置。
- **本地终结 SSL 会话** - 解密传入请求，加密服务器响应，这样后端服务器就不必完成这些潜在的高成本的操作。免除了在每个服务器上安装 [X.509](https://en.wikipedia.org/wiki/X.509) 证书的需要
- **压缩** - 压缩服务器响应
- **缓存** - 直接返回命中的缓存结果
- **静态内容** - 直接提供静态内容
  - HTML/CSS/JS
  - 图片
  - 视频
  - 等等

### 负载均衡器与反向代理

- 当你有多个服务器时，部署负载均衡器非常有用。通常，负载均衡器将流量路由给一组功能相同的服务器上。
- 即使只有一台 web 服务器或者应用服务器时，反向代理也有用，可以参考上一节介绍的好处。
- NGINX 和 HAProxy 等解决方案可以同时支持第七层反向代理和负载均衡。

### 不利之处：反向代理
- 引入反向代理会增加系统的复杂度。
- 单独一个反向代理服务器仍可能发生单点故障，配置多台反向代理服务器（如[故障转移](https://en.wikipedia.org/wiki/Failover)）会进一步增加复杂度。

### 来源及延伸阅读
- [反向代理与负载均衡](https://www.nginx.com/resources/glossary/reverse-proxy-vs-load-balancer/)
- [NGINX 架构](https://www.nginx.com/blog/inside-nginx-how-we-designed-for-performance-scale/)
- [HAProxy 架构指南](http://www.haproxy.org/download/1.2/doc/architecture.txt)
- [Wikipedia](https://en.wikipedia.org/wiki/Reverse_proxy)
- [NGINX 架构](https://www.nginx.com/blog/inside-nginx-how-we-designed-for-performance-scale/)
- [HAProxy 架构指南](http://www.haproxy.org/download/1.2/doc/architecture.txt)
- [可扩展性](http://www.lecloud.net/post/7295452622/scalability-for-dummies-part-1-clones)
- [Wikipedia](https://en.wikipedia.org/wiki/Load_balancing_(computing))
- [四层负载平衡](https://www.nginx.com/resources/glossary/layer-4-load-balancing/)
- [七层负载平衡](https://www.nginx.com/resources/glossary/layer-7-load-balancing/)
- [ELB 监听器配置](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-listener-config.html)