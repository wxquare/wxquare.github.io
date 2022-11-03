---
title: Golang 后台常用组件介绍
categories: 
- C/C++
---

## Gin web 
我们先思考下，一个完整的Web开发框架需要做哪些事情
- server，作为server，监听端口，接受请求
- router 路由和分组路由，可以把请求路由到对应的处理函数
- middleware 支持中间件，对外部发过来的http请求经过中间件处理，再给到对应的处理函数。例如http请求的日志记录、请求鉴权(比如校验token)、CORS支持、CSRF校验等。
- Crash-free：崩溃恢复，Gin可以捕捉运行期处理http请求过程中的panic并且做recover操作，让服务一直可用。
- JSON validation：JSON验证。Gin可以解析和验证request里的JSON内容，比如字段必填等。当然开发人员也可以选择使用第三方的JSON validation工具，比如[beego validation](https://github.com/beego/beego/tree/develop/core/validation)。
- Error management：错误管理。Gin提供了一种简单的方式可以收集http request处理过程中的错误，最终中间件可以选择把这些错误写入到log文件、数据库或者发送到其它系统。
- Middleware Extendtable：可以自定义中间件。Gin除了自带的官方中间件之外，还支持用户自定义中间件，甚至可以把自己开发的中间件提交到[官方代码仓库](https://github.com/gin-gonic/contrib)里。

参考：
- [web 框架比较](https://github.com/jincheng9/go-tutorial/tree/main/workspace/gin/01)
- https://github.com/gin-gonic/gin
- https://github.com/gin-gonic/contrib


## 任务调度
### 单点调度：https://github.com/robfig/cron
### 分布式调度：https://github.com/xuxueli/xxl-job
将调度行为抽象形成“调度中心”公共平台，而平台自身并不承担业务逻辑，“调度中心”负责发起调度请求。将任务抽象成分散的JobHandler，交由“执行器”统一管理，“执行器”负责接收调度请求并执行对应的JobHandler中业务逻辑。因此，“调度”和“任务”两部分可以相互解耦，提高系统整体稳定性和扩展性
- 调度模块（调度中心）：
负责管理调度信息，按照调度配置发出调度请求，自身不承担业务代码。调度系统与任务解耦，提高了系统可用性和稳定性，同时调度系统性能不再受限于任务模块；
支持可视化、简单且动态的管理调度信息，包括任务新建，更新，删除，GLUE开发和任务报警等，所有上述操作都会实时生效，同时支持监控调度结果以及执行日志，支持执行器Failover。
- 执行模块（执行器，executor）：
负责接收调度请求并执行任务逻辑。任务模块专注于任务的执行等操作，开发和维护更加简单和高效；
接收“调度中心”的执行请求、终止请求和日志请求等

参考：
- https://www.xuxueli.com/xxl-job/
- https://github.com/mousycoder/xxl-job-go-sdk



## zookeeper
Zookeeper是一个高性能、分布式的开源的协作服务；
提供一系列简单的功能，分布式应用可以在此基础上实现例如数据发布/订阅、负载均衡、命名服务、分布式协调/通知、集群管理、Leader选举、分布式锁和分布式队列等。常用的场景：
- 命名服务（Name Service）
- 配置中心
- 分布式锁
- 集群管理
参考：
1. [zookeeper介绍与使用场景](https://juejin.cn/post/6911981919974457358)
2. [golang 操作zookeeper](https://www.cnblogs.com/zhichaoma/p/12640064.html)
3. https://zookeeper.apache.org/
4. https://github.com/go-zookeeper/zk
5. 服务发现 zk https://blog.csdn.net/zyhlwzy/article/details/101847565


## 延时队列

任务队列跟消息队列在使用场景上最大的区别是： 任务之间是没有顺序约束而消息要求顺序(FIFO)，且可能会对任务的状态更新而消息一般只会消费不会更新。 类似 Kafka 利用消息 FIFO 和不需要更新(不需要对消息做索引)的特性来设计消息存储，将消息读写变成磁盘的顺序读写来实现比较好的性能。而任务队列需要能够任务状态进行更新则需要对每个消息进行索引，如果把两者放到一起实现则很难实现在功能和性能上兼得。比如一下场景：
- 定时任务，如每天早上 8 点开始推送消息，定期删除过期数据等
- 任务流，如自动创建 Redis 流程由资源创建，资源配置，DNS 修改等部分组成，使用任务队列可以简化整体的设计和重试流程
- 重试任务，典型场景如离线图片处理
<img src=https://raw.githubusercontent.com/bitleak/lmstfy/master/doc/lmstfy-internal.png width=600/>

参考：
- 延时队列
- https://juejin.cn/post/7000189281641693192
- https://github.com/bitleak/lmstfy


## backoff 服务异常重试-指数退避算法
在wiki当中对指数退避算法的介绍是：“In a variety of computer networks, binary exponential backoff or truncated binary exponential backoff refers to an algorithm used to space out repeated retransmissions of the same block of data, often as part of network congestion avoidance.”

翻译成中文的意思大概是“在各种的计算机网络中，二进制指数后退或是截断的二进制指数后退使用于一种隔离同一数据块重复传输的算法，常常做为网络避免冲突的一部分”

比如说在我们的服务调用过程中发生了调用失败，系统要对失败的资源进行重试，那么这个重试的时间如何把握，使用指数退避算法我们可以在某一范围内随机对失败的资源发起重试，并且随着失败次数的增加长，重试时间也会随着指数的增加而增加。

当然，指数退避算法并没有人上面说的那么简单，想具体了解的可以具体wiki上的介绍
参考：
- https://en.wikipedia.org/wiki/Exponential_backoff
- github golang 实现：https://github.com/cenkalti/backoff
- https://github.com/cenkalti/backoff


## RPC 框架
- https://github.com/jincheng9/go-tutorial/tree/main/workspace/rpc/02
- trpc


## 监控平台
- prometheus,https://prometheus.io/
- grafna,https://www.google.com.hk/search?q=grafana&rlz=1C5GCEM_enCN985CN985&oq=grafana&aqs=chrome..69i57j69i60l3j69i65l3j69i60.8511j0j7&sourceid=chrome&ie=UTF-8


## 其它

- 限流的设计和实，单机限流，分布式限流
- abtest 平台
- jenkins
- docker
- Kubernetes [Kubernetes 入门&进阶实战](https://zhuanlan.zhihu.com/p/339008746)

# 数据分析和处理
- scala
- spark
- spark streaming
- hive
- presto

# 方案设计与写作
- 方案模版
- 画架构图
- 数据

# 系统和架构设计
- https://wxquare.github.io/2022/05/20/C++/system-design-and-framework-basic/

# 英语能力

# 好用工具
- https://wxquare.github.io/2022/05/20/other/tools/

# 好的博客，站点
- https://catcoding.me/
- https://coderscat.com/
- https://www.zhihu.com/people/wxquare0
- https://leetcode.cn/leetbook/read/leetcode-cookbook/5is6a6/

# 源码和开源社区

# 其它



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
10. bus 中 bitmap使用，[bitmap原理和应用](https://www.jianshu.com/p/970c367460b1)

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
