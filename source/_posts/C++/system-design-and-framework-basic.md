---
title: 系统方案设计
categories: 
- C/C++
---


##  Gin web 
- https://github.com/jincheng9/go-tutorial/tree/main/workspace/gin/01
- https://github.com/jincheng9/go-tutorial/tree/main/workspace/gin/02
- https://github.com/gin-gonic/gin

## 服务发现 zk
- https://blog.csdn.net/zyhlwzy/article/details/101847565


## 定时任务调度 cron task
- https://github.com/robfig/cron


## 源码阅读
1. 如何实现定时任务? https://github.com/go-co-op/gocron
2. 协程池的实现? https://strikefreedom.top/high-performance-implementation-of-goroutine-pool
3. database/sql连接池的实现,mysql链接池的实现? github.com/go-sql-driver/mysql




## 系统设计
1. protobuf 为什么这么快,tlv编码 https://blog.csdn.net/carson_ho/article/details/70568606s
2. [阿里云，救火必备！问题排查与系统优化手册](https://zhuanlan.51cto.com/art/202007/620840.htm)
    - 常见的问题及其应对办法
    - 怎么做系统优化
3. 怎么设计一个分布式调度系统（滴滴）
4. [使用redis实现微信步数排行榜](https://www.cnblogs.com/zwwhnly/p/13041641.html)
5. https://leetcode-cn.com/circle/discuss/ej0oh6/
6. 虚拟机与容器的区别？虚拟机需要多一层guestos，隔离更好，一把是用户级别的隔离。而docker则是应用级别的隔离，共享宿主机操作系统。
7. docker和k8s之间的关系：官方定义1：Docker是一个开源的应用容器引擎，开发者可以打包他们的应用及依赖到一个可移植的容器中，发布到流行的Linux机器上，也可实现虚拟化。官方定义2：k8s是一个开源的容器集群管理系统，可以实现容器集群的自动化部署、自动扩缩容、维护等功能。
9. 负载均衡与l5名字服务？https://blog.csdn.net/qq_18144747/article/details/86672206

## go
1. [Golang调度器GPM原理与调度全分析](https://zhuanlan.zhihu.com/p/323271088)
2. [为什么要使用 Go 语言？Go 语言的优势在哪里？](https://www.zhihu.com/question/21409296/answer/1040884859)
3. [Go内置数据结构原理](https://zhuanlan.zhihu.com/p/341945051)
4. [从 bug 中学习：六大开源项目告诉你 go 并发编程的那些坑](https://zhuanlan.zhihu.com/p/352589023)
5. [Go runtime剖析系列（一）：内存管理](https://zhuanlan.zhihu.com/p/323915446)
6. [Go 内存泄露三宗罪](http://km.oa.com/group/19253/articles/show/460278?kmref=home_headline)

## redis
6. [Redis 多线程网络模型全面揭秘](https://zhuanlan.zhihu.com/p/356059845)

## db
1. [https://zhuanlan.zhihu.com/p/329865336](https://zhuanlan.zhihu.com/p/329865336)
2. 

## 
1. [Kubernetes 入门&进阶实战](https://zhuanlan.zhihu.com/p/339008746)
2. Lambda 和 Kappa 架构简介：https://libertydream.github.io/2020/04/12/lambda-%E5%92%8C-kappa-%E7%AE%80%E4%BB%8B/
3. https://blog.csdn.net/weixin_39471249/article/details/79585231


## 微服务架构
