---
title: 互联网系统设计 - 代理、负载均衡和高可用性网关
date: 2024-02-02
categories:
- 系统设计
---


## 代理
### 正常代理
正向代理是指对客户端提供的代理服务，在客户端无法直接访问服务端的情况下，通过配置代理服务器的方式访问服务端。
在整个过程中，客户端请求首先发送到代理服务器，代理服务器再将请求发送到服务端后将结果返回给客户端。从服务端角度来看，认为代理服务器才客户端，因此正向代理即代理客户端与服务端进行交互。比如生活中我们通过代购去购买海外商品，代购就是我们的正向代理。
- 提供网络通道：解决客户端由于防火墙或网络限制无法访问服务端的问题，如访问google等国外网站。
- 隐藏客户端身份：服务端只感知代理服务器，无法获取真实客户端，如黑客控制肉鸡
### 反向代理
反向代理是指对服务端提供的代理服务，通常出于安全考虑，真正的服务端只有内网网络，无法直接提供对外服务，为此需要设置反向代理服务器，由代理服务器接收外网请求，然后再转发到内部服务器。从客户端角度看，代理服务器是提供服务的服务端，因此反向代理即代理服务端与客户端交互
- 提供对外服务：代理服务器暴露公网地址，接收请求并转发到内网服务器。
- 负载均衡：根据预设策略将请求分发到多台服务器

### 区别
- 正向代理代理客户端，服务端认为请求来自代理服务器；反向代理代理服务端，客户端认为提供服务的是代理服务器
- 正向代理通常解决访问限制的问题，反向代理通常解决对外服务和负载均衡的问题

## 负载均衡和高可用性网关
### 负载均衡方案
- 基于DNS的负载均衡
在DNS服务器中，可以为多个不同的地址配置相同的名字，最终查询这个名字的客户机将在解析这个名字时得到其中一个地址，所以这种代理方式是通过DNS服务中的随机名字解析域名和IP来实现负载均衡。

- 基于NAT的负载均衡（四层）
该技术通过一个地址转换网关将每个客户端连接转换为不同的内部服务器地址，因此客户端就各自与自己转换得到的地址上的服务器进行通信，从而达到负载均衡的目的，如LVS和Nginx的四层配置形式

- 反向代理负载均衡（7层）
通常的反向代理技术，支持为同一服务配置多个后端服务器地址，以及设定相应的轮询策略。请求到达反向代理服务器后，代理通过既定的轮询策略转发请求到具体服务器，实现负载均衡，如Nginx的七层配置形式。

<p align="center">
  <img src="/images/load_balancer_architecture.jpeg" width=600 height=700>
  <br/>
</p>

### 网络负载均衡器（L4,lvs）
- CIP：客户端ip地址
- VIP：lvs服务器对外发布的ip地址，用户通过vip访问集群
- DIP：LVS连内网的ip地址叫DIP，用于接收用户请求的ip叫做VIP
- RS：提供服务的服务器
用户访问流程：
  客户端通过 CIP--->VIP--->DIP---->RIP
- https://www.cnblogs.com/heyongshen/p/16827111.html

四层负载常用软件有：
- LVS（常用，稳定性最好）
- Nginx（需要额外编译stream模块）
- HaProxy

### 七层负载均衡器和反向代理

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

七层负载常用软件有：
- Nginx
- haproxy

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

## 来源及延伸阅读
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