---
title: kafka 基础知识
categories:
- other
---

## 一、 kafka 使用场景
1. 数据的接入和处理进行异步解耦，相比接口调用，减少单个服务的复杂性
2. 增加数据处理的灵活性，提高扩展性
3. 面对突发流量，有一定的消除峰作用
4. 其它作用：扩展性、冗余、顺序性（partition中的数据有序）   
实例：广告系统和用户增长项目中经常将用户行为数据和广告投放数据接入消息队列中做后续的处理

## 二、 kafka 基本概念
![kafka架构图](/images/kafka_architecture.jpg)

- broker
Kafka 集群包含一个或多个服务器，服务器节点称为broker。broker 是消息的代理，Producers往Brokers里面的指定Topic中写消息，Consumers从Brokers里面拉取指定Topic的消息，然后进行业务处理，broker在中间起到一个代理保存消息的中转站。  
- topic
topic类似于kafka中表名，每条发布到Kafka集群的消息都有一个类别，这个类别被称为Topic。（物理上不同Topic的消息分开存储，逻辑上一个Topic的消息虽然保存于一个或多个broker上但用户只需指定消息的Topic即可生产或消费数据而不必关心数据存于何处）
- producer和client id
生产者即数据的发布者，该角色将消息发布到Kafka的topic中。broker接收到生产者发送的消息后，broker将该消息追加到当前用于追加数据的segment文件中。生产者发送的消息，存储到一个partition中，生产者也可以指定数据存储的partition。
- Consumer 、Consumer Group 和 group id
消费者可以从broker中读取数据。消费者可以消费多个topic中的数据。每个Consumer属于一个特定的Consumer Group。这是kafka用来实现一个topic消息的广播（发给所有的consumer）和单播（发给任意一个consumer）的手段。一个topic可以有多个CG。topic的消息会复制-给consumer。如果需要实现广播，只要每个consumer有一个独立的CG就可以了。要实现单播只要所有的consumer在同一个CG。用CG还可以将consumer进行自由的分组而不需要多次发送消息到不同的topic。
- Partition 和 offset
topic中的数据分割为一个或多个partition。每个topic至少有一个partition。每个partition中的数据使用多个segment文件存储。partition中的数据是有序的，不同partition间的数据丢失了数据的顺序。如果topic有多个partition，消费数据时就不能保证数据的顺序。在需要严格保证消息的消费顺序的场景下，需要将partition数目设为1。
- Leader 和 follower
每个partition有多个副本，其中有且仅有一个作为Leader，Leader是当前负责数据的读写的partition。Follower跟随Leader，所有写请求都通过Leader路由，数据变更会广播给所有Follower，Follower与Leader保持数据同步。如果Leader失效，则从Follower中选举出一个新的Leader。当Follower与Leader挂掉、卡住或者同步太慢，leader会把这个follower从“in sync replicas”（ISR）列表中删除，重新创建一个Follower。
- zookeeper
zookeeper 是一个分布式的协调组件，早期版本的kafka用zk做meta信息存储，consumer的消费状态，group的管理以及 offset的值。考虑到zk本身的一些因素以及整个架构较大概率存在单点问题，新版本中逐渐弱化了zookeeper的作用。新的consumer使用了kafka内部的group coordination协议，也减少了对zookeeper的依赖，但是broker依然依赖于ZK，zookeeper 在kafka中还用来选举controller 和 检测broker是否存活等等

## 三、kafka是怎么做到高性能
https://blog.csdn.net/kzadmxz/article/details/101576401
Kafka虽然除了具有上述优点之外，还具有高性能、高吞吐、低延时的特点，其吞吐量动辄几十万、上百万。
- 盘顺序写入。Kafka的message是不断追加到本地磁盘文件末尾的，而不是随机的写入。所以Kafka是不会删除数据的，它会把所有的数据都保留下来，每个消费者（Consumer）对每个Topic都有一个offset用来表示 读取到了第几条数据 。
- 减少内核态和用户态的拷贝
- 批量读写和批量压缩
- 权衡，不支持修改删除，只能增加数据
Kafka速度的秘诀在于，它把所有的消息都变成一个批量的文件，并且进行合理的批量压缩，减少网络IO损耗，通过mmap提高I/O速度，写入数据的时候由于单个Partion是末尾添加所以速度最优；读取数据的时候配合sendfile直接暴力输出。

## 四、kafka消息堆积
在消费者端，kafka只允许单个分区的数据被一个消费者线程消费，如果消费者越多意味着partition也要越多。然而在分区数量有限的情况下，消费者数量也就会被限制。在这种约束下，如果消息堆积了该如何处理？
目前我处理的方法比较粗暴，消费消息的时候直接返回，然后启动异步线程去处理消息，消息如果再处理的过程中失败的话，再重新发送到kafka中。
- 增加分区数量
- 优化消费速度
- 增加并行度，找多个人消化

## 五、Rebalance 机制
https://zhuanlan.zhihu.com/p/46963810  
Rebalance本身是Kafka集群的一个保护设定，用于剔除掉无法消费或者过慢的消费者，然后由于我们的数据量较大，同时后续消费后的数据写入需要走网络IO，很有可能存在依赖的第三方服务存在慢的情况而导致我们超时。Rebalance对我们数据的影响主要有以下几点：
- 数据重复消费: 消费过的数据由于提交offset任务也会失败，在partition被分配给其他消费者的时候，会造成重复消费，数据重复且增加集群压力
- Rebalance扩散到整个ConsumerGroup的所有消费者，因为一个消费者的退出，导致整个Group进行了Rebalance，并在一个比较慢的时间内达到稳定状态，影响面较大
- 频繁的Rebalance反而降低了消息的消费速度，大部分时间都在重复消费和Rebalance
- 数据不能及时消费，会累积lag，在Kafka的超过一定时间后会丢弃数据


## 六、Kafka中的消息是否会丢失和重复消费
要确定Kafka的消息是否丢失或重复，从两个方面分析入手：消息发送和消息消费。

### 1、消息发送
Kafka消息发送有两种方式：同步（sync）和异步（async），默认是同步方式，可通过producer.type属性进行配置。Kafka通过配置request.required.acks属性来确认消息的生产：
0 ---表示不进行消息接收是否成功的确认；
1 ---表示当Leader接收成功时确认；
-1---表示Leader和Follower都接收成功时确认；
综上所述，有6种消息生产的情况，下面分情况来分析消息丢失的场景：
（1）acks=0，不和Kafka集群进行消息接收确认，则当网络异常、缓冲区满了等情况时，消息可能丢失；
（2）acks=1、同步模式下，只有Leader确认接收成功后但挂掉了，副本没有同步，数据可能丢失；

### 2、消息消费
Kafka消息消费有两个consumer接口，Low-level API和High-level API：
- Low-level API：消费者自己维护offset等值，可以实现对Kafka的完全控制
- High-level API：封装了对parition和offset的管理，使用简单  
  如果使用高级接口High-level API，可能存在一个问题就是当消息消费者从集群中把消息取出来、并提交了新的消息offset值后，还没来得及消费就挂掉了，那么下次再消费时之前没消费成功的消息就“诡异”的消失了；解决办法：
针对消息丢失：同步模式下，确认机制设置为-1，即让消息写入Leader和Follower之后再确认消息发送成功；异步模式下，为防止缓冲区满，可以在配置文件设置不限制阻塞超时时间，当缓冲区满时让生产者一直处于阻塞状态；针对消息重复：将消息的唯一标识保存到外部介质中，每次消费时判断是否处理过即可。

消息重复消费及解决参考：https://www.javazhiyin.com/22910.html

## Kafka Golang Sarama的使用demo
https://github.com/wxquare/programming/blob/master/golang/util/kafka_util.go
