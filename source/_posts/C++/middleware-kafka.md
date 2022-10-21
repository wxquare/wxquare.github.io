---
title: Middleware - kafka 
categories:
- C/C++
---

## kafka 特点和使用场景
- kafka具有高吞吐、低延迟、分布式容错、持久化、可扩展的特点，常用于系统之间的异步解偶，相比接口调用，减少单个服务的复杂性
- 场景1: 系统间不同模块的异步解偶，例如电商系统的订单和发货
- 场景2：系统或者用户日志的采集、异步分析、持久化
- 场景3: 保存收集流数据，以提供之后对接的Storm或其他流式计算框架进行处理。例如风控系统


## 基本概念和组成
![kafka架构图](source/images/kafka_architecture.png)
- broker
Kafka 集群包含一个或多个服务器，服务器节点称为broker。broker 是消息的代理，Producers往Brokers里面的指定Topic中写消息，Consumers从Brokers里面拉取指定Topic的消息，然后进行业务处理，broker在中间起到一个代理保存消息的中转站。 
- producer和client id
生产者即数据的发布者，该角色将消息发布到Kafka的topic中。broker接收到生产者发送的消息后，broker将该消息追加到当前用于追加数据的segment文件中。生产者发送的消息，存储到一个partition中，生产者也可以指定数据存储的partition。
- Consumer 、Consumer Group 和 group id
消费者可以从broker中读取数据。消费者可以消费多个topic中的数据。每个Consumer属于一个特定的Consumer Group。这是kafka用来实现一个topic消息的广播（发给所有的consumer）和单播（发给任意一个consumer）的手段。一个topic可以有多个CG。topic的消息会复制-给consumer。如果需要实现广播，只要每个consumer有一个独立的CG就可以了。要实现单播只要所有的consumer在同一个CG。用CG还可以将consumer进行自由的分组而不需要多次发送消息到不同的topic。
- topic
topic类似于kafka中表名，每条发布到Kafka集群的消息都有一个类别，这个类别被称为Topic。（物理上不同Topic的消息分开存储，逻辑上一个Topic的消息虽然保存于一个或多个broker上但用户只需指定消息的Topic即可生产或消费数据而不必关心数据存于何处）
- Partition 和 offset
topic中的数据分割为一个或多个partition。每个topic至少有一个partition。每个partition中的数据使用多个segment文件存储。partition中的数据是有序的，不同partition间的数据丢失了数据的顺序。如果topic有多个partition，消费数据时就不能保证数据的顺序。在需要严格保证消息的消费顺序的场景下，需要将partition数目设为1。
- Leader 和 follower
每个partition有多个副本，其中有且仅有一个作为Leader，Leader是当前负责数据的读写的partition。Follower跟随Leader，所有写请求都通过Leader路由，数据变更会广播给所有Follower，Follower与Leader保持数据同步。如果Leader失效，则从Follower中选举出一个新的Leader。当Follower与Leader挂掉、卡住或者同步太慢，leader会把这个follower从“in sync replicas”（ISR）列表中删除，重新创建一个Follower。
- zookeeper
zookeeper 是一个分布式的协调组件，早期版本的kafka用zk做meta信息存储，consumer的消费状态，group的管理以及 offset的值。考虑到zk本身的一些因素以及整个架构较大概率存在单点问题，新版本中逐渐弱化了zookeeper的作用。新的consumer使用了kafka内部的group coordination协议，也减少了对zookeeper的依赖，但是broker依然依赖于ZK，zookeeper 在kafka中还用来选举controller 和 检测broker是否存活等等


## 系统所需要可靠性语义
### 生产者producer
**业务上需要考关注失败、丢失、重复三个问题**：
- 消费发送失败：消息写入失败是否需要ack，是否需要重试
- 消息发送重复：同一条消息重复写入对系统产生的影响
- 消息发送丢失：消息写入成功，但是由于kafka内部的副本、容错机制，导致消息丢失对系统产生的影响

**三种语义**：

- **至少一次语义（At least once semantics）**：如果生产者收到了Kafka broker的确认（acknowledgement，ack），并且生产者的acks配置项设置为all（或-1），这就意味着消息已经被精确一次写入Kafka topic了。然而，如果生产者接收ack超时或者收到了错误，它就会认为消息没有写入Kafka topic而尝试重新发送消息。如果broker恰好在消息已经成功写入Kafka topic后，发送ack前，出了故障，生产者的重试机制就会导致这条消息被写入Kafka两次，从而导致同样的消息会被消费者消费不止一次。每个人都喜欢一个兴高采烈的给予者，但是这种方式会导致重复的工作和错误的结果。
- **至多一次语义（At most once semantics）**：如果生产者在ack超时或者返回错误的时候不重试发送消息，那么消息有可能最终并没有写入Kafka topic中，因此也就不会被消费者消费到。但是为了避免重复处理的可能性，我们接受有些消息可能被遗漏处理。
- **精确一次语义（Exactly once semantics）**： 即使生产者重试发送消息，也只会让消息被发送给消费者一次。精确一次语义是最令人满意的保证，但也是最难理解的。因为它需要消息系统本身和生产消息的应用程序还有消费消息的应用程序一起合作。比如，在成功消费一条消息后，你又把消费的offset重置到之前的某个offset位置，那么你将收到从那个offset到最新的offset之间的所有消息。这解释了为什么消息系统和客户端程序必须合作来保证精确一次语义

**实践**
Kafka消息发送有两种方式：同步（sync）和异步（async），默认是同步方式，可通过producer.type属性进行配置。Kafka通过配置request.required.acks属性来确认消息的生产：
- 0 ---表示不进行消息接收是否成功的确认；
- 1 ---表示当Leader接收成功时确认；
- -1---表示Leader和Follower都接收成功时确认

综上所述，有6种消息生产的情况，下面分情况来分析消息丢失的场景：
- acks=0，不和Kafka集群进行消息接收确认，则当网络异常、缓冲区满了等情况时，消息可能丢失；
- acks=1、同步模式下，只有Leader确认接收成功后但挂掉了，副本没有同步，数据可能丢失；

**通常来说，producer 采用at least once方式**

### 消息消费consumer
- **重复消息的幂等性**：由于生产者可能多次投递和消费者commit机制等原因，消费者重复消费是很常见的问题，需要思考系统对于幂等性的要求。在很多场景下， 比如写db、redis是天然的幂等性，某些特殊的场景，可以根据唯一id，借助例如redis判别是否消费过来实现消费者的幂等性
- **消息丢失**：评估消息丢失的影响和容忍度
- **commit**：考虑auto commit 和 mannul commit



## 监控topic消息堆积情况（lag)
在实际业务场景中，由于consumer消费速度慢于producer的速度，会造成消息堆积，最终会导致消息过期删除丢失。业务需要监控这种lag情况，并及时告警出来。

另外需要注意的是，kafka只允许单个分区的数据被一个消费者线程消费，如果消费者越多意味着partition也要越多。

然而在分区数量有限的情况下，消费者数量也就会被限制。在这种约束下，如果消息堆积了该如何处理？

消费消息的时候直接返回，然后启动异步线程去处理消息，消息如果再处理的过程中失败的话，再重新发送到kafka中。
- 增加分区数量
- 优化消费速度
- 增加并行度，找多个人消化


## Rebalance 机制以及可能产生的影响
Rebalance本身是Kafka集群的一个保护设定，用于剔除掉无法消费或者过慢的消费者，然后由于我们的数据量较大，同时后续消费后的数据写入需要走网络IO，很有可能存在依赖的第三方服务存在慢的情况而导致我们超时。Rebalance对我们数据的影响主要有以下几点：
- 数据重复消费: 消费过的数据由于提交offset任务也会失败，在partition被分配给其他消费者的时候，会造成重复消费，数据重复且增加集群压力
- Rebalance扩散到整个ConsumerGroup的所有消费者，因为一个消费者的退出，导致整个Group进行了Rebalance，并在一个比较慢的时间内达到稳定状态，影响面较大
- 频繁的Rebalance反而降低了消息的消费速度，大部分时间都在重复消费和Rebalance
- 数据不能及时消费，会累积lag，在Kafka的超过一定时间后会丢弃数据
- https://zhuanlan.zhihu.com/p/46963810



## kafka是怎么做到高性能
Kafka虽然除了具有上述优点之外，还具有高性能、高吞吐、低延时的特点，其吞吐量动辄几十万、上百万。
- **磁盘顺序写入**。Kafka的message是不断追加到本地磁盘文件末尾的，而不是随机的写入。所以Kafka是不会删除数据的，它会把所有的数据都保留下来，每个消费者（Consumer）对每个Topic都有一个offset用来表示 读取到了第几条数据 。
- **操作系统page cache**，使得kafka的读写操作基本基于内存，提高读写的性能
- **零拷贝**，操作系统将数据从Page Cache 直接发送socket缓冲区，减少内核态和用户态的拷贝
-  消息topic分区partition、segment存储，提高数据操作的并行度。
-  **批量读写和批量压缩**
Kafka速度的秘诀在于，它把所有的消息都变成一个批量的文件，并且进行合理的批量压缩，减少网络IO损耗，通过mmap提高I/O速度，写入数据的时候由于单个Partion是末尾添加所以速度最优；读取数据的时候配合sendfile直接暴力输出。
- https://blog.csdn.net/kzadmxz/article/details/101576401


## Kafka文件存储机制
- 逻辑上以topic进行分类和分组
- 物理上topic以partition分组，一个topic分成若干个partition，物理上每个partition为一个目录，名称规则为topic名称+partition序列号
- 每个partition又分为多个segment（段），segment文件由两部分组成，.index文件和.log文件。通过将partition划分为多个segment，避免单个partition文件无限制扩张，方便旧的消息的清理。


## kafka partition 副本ISR机制保障高可用性
- 为了保障消息的可靠性，kafka中每个partition会设置大于1的副本数。
- 每个patition都有唯一的leader
- partition的所有副本称为AR。所有的副本（replicas）统称为Assigned Replicas，即AR。ISR是AR中的一个子集，由leader维护ISR列表，follower从leader同步数据有一些延迟（包括延迟时间replica.lag.time.max.ms和延迟条数replica.lag.max.messages两个维度, 当前最新的版本0.10.x中只支持replica.lag.time.max.ms这个维度），任意一个超过阈值都会把follower剔除出ISR, 存入OSR（Outof-Sync Replicas）列表，新加入的follower也会先存放在OSR中。AR=ISR+OSR
- partition 副本同步机制。Kafka的复制机制既不是完全的同步复制，也不是单纯的异步复制。事实上，同步复制要求所有能工作的follower都复制完，这条消息才会被commit，这种复制方式极大的影响了吞吐率。而异步复制方式下，follower异步的从leader复制数据，数据只要被leader写入log就被认为已经commit，这种情况下如果follower都还没有复制完，落后于leader时，突然leader宕机，则会丢失数据。而Kafka的这种使用ISR的方式则很好的均衡了确保数据不丢失以及吞吐率
当producer向leader发送数据时，可以通过request.required.acks参数来设置数据可靠性的级别：
  - 1（默认）：这意味着producer在ISR中的leader已成功收到数据并得到确认。如果leader宕机了，则会丢失数据。
  - 0：这意味着producer无需等待来自broker的确认而继续发送下一批消息。这种情况下数据传输效率最高，但是数据可靠性确是最低的。
  - -1：producer需要等待ISR中的所有follower都确认接收到数据后才算一次发送完成，可靠性最高。但是这样也不能保证数据不丢失，比如当ISR中只有leader时（前面ISR那一节讲到，ISR中的成员由于某些情况会增加也会减少，最少就只剩一个leader），这样就变成了acks=1的情况。
- ISR 副本选举leader
- https://blog.csdn.net/u013256816/article/details/71091774



## kafka 命令
- 创建topic
```
bin/kafka-topics.sh --create --topic topic-name --replication-factor 2 --partitions 3 --bootstrap-server ip:port
```

- 查看topic情况
```
bin/kafka-topics.sh --topic topic_name --describe --bootstrap-server broker 

```

- 查看消费组情况
```
./bin/kafka-consumer-groups.sh --describe --group group_name  --bootstrap-server brokers
```

## 参考
1. Kafka Golang Sarama的使用demo,https://github.com/wxquare/programming/blob/master/golang/util/kafka_util.go
2. [kafka数据可靠性深度解读](https://blog.csdn.net/u013256816/article/details/71091774)
3. [kafka 选举](https://juejin.im/post/6844903846297206797)
4. [Kafka为什么吞吐量大、速度快？](https://blog.csdn.net/kzadmxz/article/details/101576401)
5. [简单理解 Kafka 的消息可靠性策略](https://cloud.tencent.com/developer/article/1752150)
6. [Bootstrap server vs zookeeper in kafka?](https://stackoverflow.com/questions/46173003/bootstrap-server-vs-zookeeper-in-kafka)
