---
title: 互联网系统设计 - 微服务原理
date: 2024-02-20
categories: 
- 系统设计
---


## RPC 服务开发
### 单体服务、微服务、Service Mesh
<p align="center">
  <img src="/images/rpc_to_service_mesh.png" width=600 height=350>
  <br/>
  <strong><a href="https://www.zhihu.com/question/56125281">什么是服务治理</a></strong>
</p>

- 单体服务（Monolithic Services）：单体服务是指将整个应用程序作为一个单一的、紧密耦合的单元进行开发、部署和运行的架构模式。在单体服务中，应用程序的各个功能模块通常运行在同一个进程中，并共享相同的数据库和资源。单体服务的优点是开发简单、部署方便，但随着业务规模的增长，单体服务可能变得庞大且难以维护。

- 微服务（Microservices）：微服务是一种将应用程序拆分为一组小型、独立部署的服务的架构模式。每个微服务都专注于单个业务功能，并通过轻量级的通信机制（如RESTful API或消息队列）进行相互通信。微服务的优点是灵活性高、可扩展性好，每个微服务可以独立开发、测试、部署和扩展。然而，微服务架构也带来了分布式系统的复杂性和管理的挑战。

- Service Mesh：Service Mesh是一种用于解决微服务架构中服务间通信和治理问题的基础设施层。它通过在服务之间插入一个专用的代理（称为Sidecar）来提供服务间的通信、安全性、可观察性和弹性的功能。Service Mesh可以提供流量管理、负载均衡、故障恢复、安全认证、监控和追踪等功能，而不需要在每个微服务中显式实现这些功能。常见的Service Mesh实现包括Istio、Linkerd和Consul Connect等。


### 微服务
<p align="center">
  <img src="/images/landing-2.svg" width=600 height=350>
  <br/>
  <strong><a href="https://grpc.io/docs/what-is-grpc/introduction">gRPC 概述</a></strong>
</p>

与此讨论相关的话题是 [微服务](https://en.wikipedia.org/wiki/Microservices)，可以被描述为一系列可以独立部署的小型的，模块化服务。每个服务运行在一个独立的线程中，通过明确定义的轻量级机制通讯，共同实现业务目标。<sup><a href=https://smartbear.com/learn/api-design/what-are-microservices>1</a></sup>例如，Pinterest 可能有这些微服务： 用户资料、关注者、Feed 流、搜索、照片上传等。

### 服务发现
**ZooKeeper**
- ZooKeeper是一个开源的分布式协调服务，最初由雅虎开发并后来成为Apache软件基金会的顶级项目。
- ZooKeeper提供了一个分布式的、高可用的、强一致性的数据存储服务。它的设计目标是为构建分布式系统提供可靠的协调机制。
- ZooKeeper使用基于ZAB（ZooKeeper Atomic Broadcast）协议的一致性算法来保证数据的一致性和可靠性。
- ZooKeeper提供了一个类似于文件系统的层次化命名空间（称为ZNode），可以存储和管理数据，并支持对数据的读写操作。
- ZooKeeper还提供了一些特性，如临时节点、顺序节点和观察者机制，用于实现分布式锁、选举算法和事件通知等。

**etcd**
- etcd是一个开源的分布式键值存储系统，由CoreOS开发并后来成为Cloud Native Computing Foundation（CNCF）的项目之一。
- etcd被设计为一个高可用、可靠的分布式存储系统，用于存储和管理关键的配置数据和元数据。
- etcd使用Raft一致性算法来保证数据的一致性和可靠性，Raft是一种强一致性的分布式共识算法。
- etcd提供了一个简单的键值存储接口，可以存储和检索键值对数据，并支持对数据的原子更新操作。
- etcd还提供了一些高级特性，如目录结构、事务操作和观察者机制，用于构建复杂的分布式系统和应用

- [Etcd](https://coreos.com/etcd/docs/latest) 
- [Zookeeper](https://zookeeper.apache.org) 
- [Consul](https://www.consul.io/docs/index.html)
- [grpc](https://grpc.io/docs)

### Service Mesh
<p align="center">
  <img src="/images/istio_service_mesh.svg" width=600 height=600>
  <br/>
  <strong><a href="https://istio.io/latest/about/service-mesh/">service Mesh 是怎么工作的</a></strong>
</p>

### 远程过程调用协议（RPC）
<p align="center">
  <img src="/images/iF4Mkb5.png" width=700 height=400>
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