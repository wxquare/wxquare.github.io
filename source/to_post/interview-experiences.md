---
title: 面试经验
categories:
- other
---


## golang 
1. 相比线程，协程有哪些优势？
1. 如何控制Goroutine的数量？
2. Goroutine复用
3. Goroutine 池？
4. 如何解决系统并发量大的问题
5. G-P-M调度（https://juejin.im/entry/6844903621969215495）
6. 解释golang GC


## 操作系统相关


## 数据库
1. 事务隔离级别，mysql默认的事务隔离级别RR（repeated read）
2. 有两个事务，事务A要要修改一个记录，未提交。事务B也要修改该记录，此时数据库会表现为什么？ commit之后会发生什么？
3. 什么是聚簇索引
4. 什么情况会发生索引失效
5. 有数据库优化的经验


## redis

## kafka


## 网路协议
1. 为了可靠传输，tcp相比udp做了哪些工作
2. 了解https吗，https是不是一种新协议，对加密算法的认识，每一步具体的含义


## linux命令
1. netstat 找出tcp链接的数据量 
2. 找不哪个进程占用了8080端口
3. grep找出关键字的前十行和后十行


## 系统设计


### bytedance
1. https://blog.csdn.net/luolianxi/article/details/105592179
2. https://www.jianshu.com/p/d424dcb6637f
3. https://www.nowcoder.com/discuss/471541
4. https://leetcode-cn.com/circle/discuss/A0YstA/(2021,1月，2week）

### tencent
1. https://blog.csdn.net/luolianxi/article/details/105606741

### microsoft
1. https://leetcode-cn.com/circle/discuss/TEXcH1/
2. https://zhuanlan.zhihu.com/p/95836541
3. [英特尔 Intel｜面经分享｜2021｜](https://leetcode-cn.com/circle/discuss/5tFRIM/)


1. 如何实现定时任务? https://github.com/go-co-op/gocron
2. 协程池的实现? https://strikefreedom.top/high-performance-implementation-of-goroutine-pool
3. database/sql连接池的实现,mysql链接池的实现? github.com/go-sql-driver/mysql
3. protobuf 为什么这么快,tlv编码 https://blog.csdn.net/carson_ho/article/details/70568606s
4. [阿里云，救火必备！问题排查与系统优化手册](https://zhuanlan.51cto.com/art/202007/620840.htm)
    - 常见的问题及其应对办法
    - 怎么做系统优化
3. 怎么设计一个分布式调度系统（滴滴）
4. [使用redis实现微信步数排行榜](https://www.cnblogs.com/zwwhnly/p/13041641.html)
5. https://leetcode-cn.com/circle/discuss/ej0oh6/
6. 虚拟机与容器的区别？虚拟机需要多一层guestos，隔离更好，一把是用户级别的隔离。而docker则是应用级别的隔离，共享宿主机操作系统。
7. docker和k8s之间的关系：官方定义1：Docker是一个开源的应用容器引擎，开发者可以打包他们的应用及依赖到一个可移植的容器中，发布到流行的Linux机器上，也可实现虚拟化。官方定义2：k8s是一个开源的容器集群管理系统，可以实现容器集群的自动化部署、自动扩缩容、维护等功能。
9. 负载均衡与l5名字服务？https://blog.csdn.net/qq_18144747/article/details/86672206

10. [Golang调度器GPM原理与调度全分析](https://zhuanlan.zhihu.com/p/323271088)
1. [为什么要使用 Go 语言？Go 语言的优势在哪里？](https://www.zhihu.com/question/21409296/answer/1040884859)
3. [Go内置数据结构原理](https://zhuanlan.zhihu.com/p/341945051)
4. [从 bug 中学习：六大开源项目告诉你 go 并发编程的那些坑](https://zhuanlan.zhihu.com/p/352589023)
5. [Go runtime剖析系列（一）：内存管理](https://zhuanlan.zhihu.com/p/323915446)
6. [Go 内存泄露三宗罪](http://km.oa.com/group/19253/articles/show/460278?kmref=home_headline)
6. [Redis 多线程网络模型全面揭秘](https://zhuanlan.zhihu.com/p/356059845)
1. [https://zhuanlan.zhihu.com/p/329865336](https://zhuanlan.zhihu.com/p/329865336)
1. [Kubernetes 入门&进阶实战](https://zhuanlan.zhihu.com/p/339008746)
2. Lambda 和 Kappa 架构简介：https://libertydream.github.io/2020/04/12/lambda-%E5%92%8C-kappa-%E7%AE%80%E4%BB%8B/
3. https://blog.csdn.net/weixin_39471249/article/details/79585231