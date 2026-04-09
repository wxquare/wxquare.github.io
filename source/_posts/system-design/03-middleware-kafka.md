---
title: 中间件 - 异步和消息队列
date: 2024-03-10
categories:
  - 系统设计基础
tags:
- Kafka
- 消息队列
- 异步
- 分布式
toc: true
---

## 速查导航

**阅读时间**: 45 分钟 | **难度**: ⭐⭐⭐⭐ | **面试频率**: 极高

**核心考点速查**:
- [一、Kafka 核心特性](#一kafka-核心特性面试必答) - 3 分钟掌握使用场景与 Kafka vs Redis
- [二、核心概念](#二核心概念5-分钟速记) - Topic/Partition/ISR/ZooKeeper vs KRaft
- [三、数据流与架构](#三数据流与架构高频考点) - Producer → Broker → Consumer 完整流程
- [四、为什么 Kafka 这么快](#四为什么-kafka-这么快面试必问) - 顺序写 + 零拷贝 + Page Cache 三板斧
- [五、消息不丢失全链路保障](#五消息不丢失全链路保障) - At-least-once/Exactly-once 配置清单
- [六、Rebalance 机制](#六rebalance-机制及影响) - 触发条件与优化方案
- [七、文件存储机制](#七文件存储机制) - Segment/Index/HW/LEO
- [八、性能调优实践](#八性能调优实践) - 消费积压/参数调优/Go 代码示例
- [九、生产踩坑实录](#九生产环境踩坑实录) - 5 个真实案例
- [十、面试高频 20 题](#十面试高频-20-题) - 标准答案 + 追问应对

---

## 一、Kafka 核心特性（面试必答）

### 1.1 三句话介绍 Kafka

**标准回答**（45 秒内说完）：
Kafka 是分布式流式消息队列，具有**高吞吐**（百万 TPS）、**低延迟**（ms 级）、**持久化**的特点。采用发布-订阅模式，常用于异步解耦、削峰填谷、日志采集。

**加分项**：提到"顺序写磁盘 + 零拷贝 + Page Cache"性能三板斧。

### 1.2 使用场景（带真实案例）

| 场景 | 痛点 | Kafka 方案 | 示例 |
|------|------|-----------|------|
| **订单处理** | 同步调用慢 | 异步解耦 | 下单 → MQ → 库存/物流并行处理 |
| **秒杀** | 瞬时流量打垮 DB | 削峰填谷 | 请求入队 → 匀速消费 |
| **日志采集** | 海量日志 | 高吞吐 | Filebeat → Kafka → ES |
| **风控系统** | 实时流计算 | 流式数据源 | Kafka → Flink/Storm → 实时告警 |
| **数据同步** | 异构系统集成 | 数据管道 | MySQL Binlog → Kafka → 数仓 |

### 1.3 面试追问：为什么不用 Redis 做消息队列？

**对比表格**：

| 维度 | Kafka | Redis |
|------|-------|-------|
| **吞吐量** | 百万级 TPS | 十万级 |
| **持久化** | 强（磁盘） | 弱（RDB/AOF 有丢失风险） |
| **消息回溯** | 支持（offset） | 不支持 |
| **集群** | 原生支持 | Cluster 模式复杂 |
| **消息大小** | 支持 MB 级 | 推荐 KB 级 |

**结论**：核心业务用 Kafka，轻量级任务可用 Redis List。

---

## 二、核心概念（5 分钟速记）

### 2.1 核心组件速查表

| 组件 | 面试关键点 | 记忆口诀 |
|------|----------|---------|
| **Broker** | 消息代理服务器，负责存储和转发 | "银行柜台"，存取消息 |
| **Topic** | 消息主题/分类，类似数据库表 | 按业务分类 |
| **Partition** | Topic 物理分片，实现并行与扩展 | **分区内有序，全局无序** |
| **Offset** | 消费位点，每条消息的唯一编号 | 类似数组下标 |
| **Producer** | 消息生产者 | 往银行存钱 |
| **Consumer** | 消息消费者 | 从银行取钱 |
| **Consumer Group** | 消费者组，**一个分区只能被组内一个消费者消费** | 多人共同分账单 |

### 2.2 副本机制（高可用核心）

**面试必问点：**

| 概念 | 解释 | 面试话术 |
|------|------|---------|
| **Leader** | 负责读写的主副本 | "所有读写请求都打在 Leader 上" |
| **Follower** | 从 Leader 同步的备份副本 | "只负责同步，不对外服务" |
| **ISR** | In-Sync Replicas，同步副本集合 | "能跟上 Leader 的副本才能入选" |
| **AR** | Assigned Replicas，所有副本 | AR = ISR + OSR |
| **OSR** | Out-of-Sync Replicas，落后的副本 | "掉队的副本会被踢出 ISR" |

**追问：Leader 挂了怎么办？**
- 从 **ISR 中选举**新 Leader（保证数据不丢）
- 如果 ISR 为空，是否允许从 OSR 选举？取决于 `unclean.leader.election.enable`（默认 false，不允许）

### 2.3 Partition 与顺序性

**面试标准答案（30 秒）：**
1. **同一 Partition 内严格有序**（按 offset 递增）
2. **不同 Partition 之间无序**
3. **需要全局有序？** 设置 Partition = 1（牺牲并行度）
4. **常见方案**：按业务 key（如订单 ID）Hash 到同一分区

**代码示例（Partition 策略）：**

```go
// 自定义分区器：按订单 ID 保证同订单消息有序
import "hash/crc32"

func OrderPartitioner(key []byte, numPartitions int) int {
    orderID := string(key)
    return int(crc32.ChecksumIEEE([]byte(orderID))) % numPartitions
}
```

### 2.4 ZooKeeper vs KRaft

| 维度 | ZooKeeper 模式（旧） | KRaft 模式（新） |
|------|---------------------|----------------|
| **元数据存储** | 外部 ZK 集群 | Kafka 内部 Raft 日志 |
| **Controller 选举** | ZK 选举 | Raft 协议选举 |
| **部署复杂度** | 高（需额外维护 ZK） | 低（无外部依赖） |
| **启动速度** | 慢（ZK session 超时） | 快 |
| **生产可用** | 稳定（旧版默认） | Kafka 3.3+ 推荐 |

**面试加分项**：新项目推荐 KRaft，减少外部依赖，简化运维。

---

## 三、数据流与架构（高频考点）

### 3.1 Producer → Broker → Consumer 完整流程

```
┌─────────────┐                ┌─────────────┐                ┌─────────────┐
│  Producer   │──①发送消息───→│   Broker    │──③消费拉取───→│  Consumer   │
│             │                │  (Leader)   │                │             │
└─────────────┘                └─────────────┘                └─────────────┘
                                     │②同步
                                     ↓
                              ┌─────────────┐
                              │  Follower   │
                              │   副本集    │
                              └─────────────┘
```

**详细步骤：**

| 阶段 | 操作 | 关键配置 |
|------|------|---------|
| **① Producer 发送** | 序列化 → 分区路由 → 批量打包 → 发送 | `acks`, `batch.size`, `linger.ms` |
| **② Broker 写入** | 追加到 log → Follower 拉取同步 → 返回 ack | `min.insync.replicas` |
| **③ Consumer 消费** | Fetch 请求 → 反序列化 → 业务处理 → 提交 offset | `enable.auto.commit` |

### 3.2 Consumer Group 与 Rebalance

消费者组内的消费者共同消费一个 Topic，每个 Partition 只能被组内一个消费者消费。

**Rebalance 触发条件：**
- 消费者加入或离开消费者组
- 订阅的 Topic Partition 数量变化
- 消费者心跳超时（`session.timeout.ms`）

**Rebalance 的影响：**
- 数据重复消费：未提交的 offset 导致消息重新投递
- 消费暂停：Rebalance 期间所有消费者停止消费
- 扩散效应：一个消费者退出可能触发整个 Group 的 Rebalance

**减少不必要 Rebalance 的方法：**
- 合理设置 `session.timeout.ms` 和 `heartbeat.interval.ms`
- 增大 `max.poll.interval.ms`，避免消费逻辑超时
- 使用 Static Membership（`group.instance.id`）减少重启引起的 Rebalance

### 3.3 Controller 与协调

Kafka 集群中会选举出一个 **Controller Broker**，负责 Partition Leader 选举、副本管理、集群元数据变更等。

- 早期依赖 ZooKeeper 存储 controller 与部分元数据
- **KRaft** 模式下，元数据以 Raft 日志形式在 controller 节点间复制，去掉外部 ZK，部署与扩缩容更简单

客户端通过 **Bootstrap Server** 列表首次连接后，会拉取集群元数据（Topic、Partition Leader、ISR 等），后续生产与消费都尽量直连对应 Leader Broker，避免所有读写流量经过单点代理。

---

## 四、为什么 Kafka 这么快？（面试必问）

Kafka 虽然是基于磁盘的消息队列，但吞吐量可达**百万 TPS**，延迟低至 **ms 级别**。核心原因是以下三大优化：

### 4.1 顺序写磁盘

**原理**：
- Kafka 的消息追加到 log 文件**末尾**，是**顺序写**（Sequential Write）
- 顺序写避免了磁盘寻道时间，性能接近内存写入（磁盘顺序写 > 内存随机写）

**对比**：
- 顺序写：600 MB/s（SSD）
- 随机写：100 MB/s（SSD）

**面试话术**：
> "Kafka 将消息顺序追加到磁盘，避免随机 IO，磁盘顺序写性能甚至优于内存随机写。"

### 4.2 Page Cache（页缓存）

**原理**：
- Kafka **不使用 JVM 堆内存**管理缓存，而是依赖操作系统的 **Page Cache**
- 操作系统会自动将热点数据缓存到内存，读写操作直接命中缓存

**优势**：
- 减少 GC 压力（无大对象在堆内存）
- 操作系统级别的缓存管理更高效
- 重启 Kafka 后 Page Cache 依然存在（操作系统管理）

**面试话术**：
> "Kafka 依赖操作系统 Page Cache，避免 JVM GC，热数据读写基本都是内存操作。"

### 4.3 零拷贝（Zero Copy）

**传统方式（4 次拷贝）**：

```
磁盘 → 内核缓冲区 → 用户空间 → Socket 缓冲区 → 网卡
```

**零拷贝方式（sendfile 系统调用）**：

```
磁盘 → Page Cache → 网卡（DMA 直接传输）
```

**优势**：
- 减少 2 次 CPU 拷贝（内核 → 用户空间，用户空间 → Socket 缓冲区）
- 减少 2 次上下文切换（用户态 ↔ 内核态）

**Go 代码示例（模拟零拷贝）**：

```go
// Go 标准库的 io.Copy 内部会使用 sendfile（Linux）或 TransmitFile（Windows）
import (
    "io"
    "net"
    "os"
)

func SendFileWithZeroCopy(conn net.Conn, filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // io.Copy 在 Linux 下会自动使用 sendfile 系统调用
    _, err = io.Copy(conn, file)
    return err
}
```

**面试话术**：
> "Kafka 使用 sendfile 系统调用，数据从磁盘通过 DMA 直接传输到网卡，减少 CPU 拷贝和上下文切换。"

### 4.4 批量读写与压缩

**批量发送**：
- Producer 会将多条消息打包成一个 batch 发送
- 配置 `batch.size`（批量大小）和 `linger.ms`（等待时间）

**批量压缩**：
- Kafka 支持 `gzip`、`snappy`、`lz4`、`zstd` 压缩
- 压缩后网络传输量减少，提高吞吐量

**面试话术**：
> "Kafka 通过批量发送和压缩，减少网络 IO 次数，提高吞吐量。"

### 4.5 分区并行

**原理**：
- 一个 Topic 可以有多个 Partition
- 多个 Partition 可以并行写入/读取，充分利用多核 CPU

**面试话术**：
> "Kafka 的 Partition 机制实现了水平扩展，多个 Partition 并行处理，提高吞吐量。"

---

## 五、消息不丢失全链路保障

### 5.1 可靠性语义

**三种语义**：

| 语义 | 生产端配置 | 消费端配置 | 适用场景 |
|------|-----------|-----------|----------|
| **At-most-once** | acks=0 | 自动提交 offset | 日志采集（允许丢失） |
| **At-least-once** | acks=all + 重试 | 手动提交 offset | 订单处理（不允许丢失） |
| **Exactly-once** | 幂等 + 事务 | 事务性消费 | 金融场景 |

### 5.2 生产端配置

| 配置项 | 推荐值 | 说明 |
|--------|--------|------|
| `acks` | `all` (-1) | 等待所有 ISR 副本确认 |
| `retries` | `≥3` | 发送失败重试次数 |
| `max.in.flight.requests.per.connection` | `1` | 配合重试保证消息顺序 |
| `enable.idempotence` | `true` | 开启幂等性，防止重复发送 |

### 5.3 Broker 端配置

| 配置项 | 推荐值 | 说明 |
|--------|--------|------|
| `min.insync.replicas` | `2` | 至少 2 个副本同步才允许写入 |
| `unclean.leader.election.enable` | `false` | 禁止非 ISR 副本成为 Leader |
| `default.replication.factor` | `3` | 默认副本数 |

### 5.4 消费端配置

- 关闭自动提交：`enable.auto.commit=false`
- 消费成功后手动提交 offset
- 消费逻辑实现幂等（唯一键/状态机/版本号）

### 5.5 Exactly-once 实现

**Exactly-once 的两个维度**：
1. **Broker 内部**：Producer 幂等性 + 事务
2. **端到端**：消费逻辑幂等

**Producer 幂等性配置**：

```go
// Go kafka-go 示例
writer := &kafka.Writer{
    Addr:         kafka.TCP("localhost:9092"),
    Topic:        "orders",
    RequiredAcks: kafka.RequireAll,  // acks=all
    Idempotent:   true,               // 开启幂等性
    MaxAttempts:  3,                  // 重试 3 次
}
```

**事务性写入**：

```go
// Go Sarama 示例（kafka-go 不支持事务）
import "github.com/Shopify/sarama"

config := sarama.NewConfig()
config.Producer.Idempotent = true
config.Producer.RequiredAcks = sarama.WaitForAll
config.Producer.Return.Errors = true
config.Producer.Transaction.ID = "my-transaction-id"  // 事务 ID

producer, _ := sarama.NewAsyncProducer(brokers, config)

// 开始事务
producer.BeginTxn()
producer.Input() <- &sarama.ProducerMessage{Topic: "orders", Value: sarama.StringEncoder("msg1")}
producer.Input() <- &sarama.ProducerMessage{Topic: "orders", Value: sarama.StringEncoder("msg2")}
producer.CommitTxn()  // 提交事务
```

### 5.6 检查清单（落地排查）

- **写进日志才算数**：Producer 未收到成功 ack 前，业务层是否错误地当作「已发送成功」并更新状态？
- **ISR 是否退化**：Broker 或副本故障后 ISR 可能暂时只剩 Leader，此时 `acks=all` 在语义上会退化为弱一致场景，需结合副本监控与告警。
- **提交时机**：手动提交 offset 是否在业务落库、调用下游成功之后执行？先提交再处理会导致宕机丢消息。
- **重试与顺序**：提高 `max.in.flight.requests.per.connection` 有利于吞吐，但与 Producer 重试组合时可能影响同分区内消息顺序，需按业务是否强依赖顺序做取舍。

---

## 六、Rebalance 机制及影响

### 6.1 什么是 Rebalance？

Rebalance 是 Kafka 消费者组内 Partition 重新分配的过程。

**触发条件：**
1. 消费者加入或离开消费者组
2. 订阅的 Topic Partition 数量变化
3. 消费者心跳超时（`session.timeout.ms`）

### 6.2 Rebalance 的影响

**面试标准答案**：
1. **数据重复消费**：未提交的 offset 导致消息重新投递
2. **消费暂停**：Rebalance 期间所有消费者停止消费（Stop-the-world）
3. **扩散效应**：一个消费者退出可能触发整个 Group 的 Rebalance

### 6.3 如何减少 Rebalance？

**配置优化**：

| 配置项 | 推荐值 | 说明 |
|--------|--------|------|
| `session.timeout.ms` | `30000`（30 秒） | 心跳超时时间（过短易误判） |
| `heartbeat.interval.ms` | `3000`（3 秒） | 心跳间隔（建议为 session.timeout 的 1/3） |
| `max.poll.interval.ms` | `300000`（5 分钟） | 两次 poll 之间的最大间隔 |
| `group.instance.id` | 设置静态成员 ID | 避免重启时触发 Rebalance |

**代码优化**：
- 消费逻辑异步化：消费时直接返回，启动异步线程处理
- 避免长时间阻塞：确保业务逻辑在 `max.poll.interval.ms` 内完成

### 6.4 监控 Lag 情况

**查看消费积压**：

```bash
kafka-consumer-groups.sh --describe --group <group-name> --bootstrap-server <broker>
```

**关键指标**：
- `CURRENT-OFFSET`：当前消费位点
- `LOG-END-OFFSET`：Partition 最新 offset
- `LAG`：积压消息数（LOG-END-OFFSET - CURRENT-OFFSET）

**告警策略**：
- 核心业务 Topic（如订单）：LAG > 1000 告警
- 日志类 Topic：LAG > 10000 告警

---

## 七、文件存储机制

### 7.1 存储结构

**逻辑上**：Topic 分为多个 Partition
**物理上**：每个 Partition 是一个目录，包含多个 Segment 文件

```
/kafka-logs/orders-0/
├── 00000000000000000000.index  # 索引文件
├── 00000000000000000000.log    # 数据文件
├── 00000000000000000000.timeindex  # 时间索引
├── 00000000000000368769.index
├── 00000000000000368769.log
└── 00000000000000368769.timeindex
```

### 7.2 Segment 滚动策略

**触发条件**：
- 文件大小达到 `log.segment.bytes`（默认 1GB）
- 时间达到 `log.roll.ms`（默认 7 天）

**文件命名**：文件名为该 Segment 起始 offset（如 `00000000000000368769.log`）

### 7.3 索引机制

**稀疏索引**：
- `.index` 文件不是为每条消息建索引，而是按间隔记录（默认每 4KB 建一条索引）
- 查找时先在索引中**二分查找**区间，再在 `.log` 中顺序扫描

**示例**：查找 offset=368800 的消息
1. 通过文件名定位到 `00000000000000368769.log`
2. 在 `.index` 文件中二分查找，找到最接近的索引项（如 offset=368790, position=1024）
3. 从 `.log` 文件的 position=1024 开始顺序扫描，找到 offset=368800

### 7.4 HW 与 LEO

| 概念 | 解释 | 面试话术 |
|------|------|---------|
| **LEO** | Log End Offset，副本本地 log 末尾的 offset | "副本最新写到哪里" |
| **HW** | High Watermark，ISR 中所有副本都复制到的 offset | "消费者能读到哪里" |

**面试追问：为什么消费者只能读到 HW 之前的数据？**
- 保证消息已被多个副本确认，避免读到未充分复制、可能丢失的数据

---

## 八、性能调优实践

### 8.1 消费积压排查步骤

**1. 确认 lag 情况**

```bash
kafka-consumer-groups.sh --describe --group <group-name> --bootstrap-server <broker>
```

**2. 定位原因**
- 消费逻辑慢：查看消费端 DB/网络/外部服务耗时
- Partition 数不够：消费者数超过 Partition 数时多余的消费者空闲
- 消费线程阻塞：检查是否有死锁或长时间 GC

**3. 应急方案**
- 临时增加消费者实例（不超过 Partition 数）
- 消费逻辑异步化：消费时直接返回，启动异步线程处理
- 跳过非关键消息：重置 offset 到最新位置

### 8.2 生产者侧调优

**批量发送**：

```go
writer := &kafka.Writer{
    Addr:         kafka.TCP("localhost:9092"),
    Topic:        "orders",
    BatchSize:    100,              // 批量大小 100 条
    BatchTimeout: 10 * time.Millisecond,  // 最多等待 10ms
}
```

**压缩**：

```go
writer := &kafka.Writer{
    Compression: kafka.Lz4,  // 使用 lz4 压缩
}
```

**分区策略**：

```go
// 无顺序要求：轮询
writer := &kafka.Writer{
    Balancer: &kafka.RoundRobin{},
}

// 需要顺序：按 key Hash
writer := &kafka.Writer{
    Balancer: &kafka.Hash{},
}
```

### 8.3 Broker 与系统层优化

**页缓存**：
- Broker 依赖 OS page cache 做热读热写，机器内存应留足给文件系统缓存
- JVM 堆过大可能与 page cache 争用，需按官方建议调优

**磁盘与 IO**：
- 数据目录尽量使用高性能 SSD
- 避免与高 IO 的其他服务混用同一盘

### 8.4 Go 生产级别 Producer 示例

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/segmentio/kafka-go"
)

// NewProducer 创建生产者
func NewProducer(brokers []string, topic string) *kafka.Writer {
    return &kafka.Writer{
        Addr:                   kafka.TCP(brokers...),
        Topic:                  topic,
        Balancer:               &kafka.Hash{},        // 按 key Hash 分区
        RequiredAcks:           kafka.RequireAll,     // acks=all
        MaxAttempts:            3,                    // 重试 3 次
        BatchSize:              100,                  // 批量 100 条
        BatchTimeout:           10 * time.Millisecond,
        Compression:            kafka.Lz4,            // lz4 压缩
        ReadTimeout:            10 * time.Second,
        WriteTimeout:           10 * time.Second,
        Idempotent:             true,                 // 幂等性
        AllowAutoTopicCreation: false,                // 禁止自动创建 Topic
    }
}

// ProduceMessage 发送消息（带重试和错误处理）
func ProduceMessage(w *kafka.Writer, key, value string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    err := w.WriteMessages(ctx, kafka.Message{
        Key:   []byte(key),
        Value: []byte(value),
        Time:  time.Now(),  // 消息时间戳
    })

    if err != nil {
        log.Printf("Failed to send message: key=%s, error=%v", key, err)
        return err
    }

    log.Printf("Message sent successfully: key=%s", key)
    return nil
}

func main() {
    producer := NewProducer([]string{"localhost:9092"}, "orders")
    defer producer.Close()

    // 发送消息
    for i := 0; i < 100; i++ {
        key := fmt.Sprintf("order-%d", i)
        value := fmt.Sprintf(`{"order_id":"%s","amount":100}`, key)
        if err := ProduceMessage(producer, key, value); err != nil {
            log.Printf("Error: %v", err)
        }
    }
}
```

### 8.5 Go 生产级别 Consumer 示例

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/segmentio/kafka-go"
)

// NewConsumer 创建消费者
func NewConsumer(brokers []string, topic, groupID string) *kafka.Reader {
    return kafka.NewReader(kafka.ReaderConfig{
        Brokers:        brokers,
        GroupID:        groupID,
        Topic:          topic,
        MinBytes:       10e3,   // 10KB
        MaxBytes:       10e6,   // 10MB
        MaxWait:        500 * time.Millisecond,
        CommitInterval: 0,      // 手动提交
        StartOffset:    kafka.LastOffset,  // 从最新位置开始
    })
}

// processMessage 业务处理逻辑（需保证幂等性）
func processMessage(msg kafka.Message) error {
    log.Printf("Received: partition=%d, offset=%d, key=%s, value=%s",
        msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))

    // TODO: 业务逻辑（写 DB、调用下游服务等）
    // 注意：必须实现幂等性！

    return nil
}

// ConsumeLoop 消费循环
func ConsumeLoop(r *kafka.Reader) {
    ctx := context.Background()

    for {
        // 拉取消息
        msg, err := r.FetchMessage(ctx)
        if err != nil {
            log.Printf("Fetch error: %v", err)
            time.Sleep(1 * time.Second)
            continue
        }

        // 业务处理
        if err := processMessage(msg); err != nil {
            log.Printf("Process error: %v", err)
            // 注意：根据业务决定是否重试或跳过
            continue
        }

        // 处理成功后手动提交 offset
        if err := r.CommitMessages(ctx, msg); err != nil {
            log.Printf("Commit error: %v", err)
        }
    }
}

func main() {
    consumer := NewConsumer([]string{"localhost:9092"}, "orders", "order-consumer-group")
    defer consumer.Close()

    log.Println("Start consuming...")
    ConsumeLoop(consumer)
}
```

### 8.6 常用配置参数总结

**Producer 配置**：

```go
c.Producer.MaxMessageBytes = 1000000  // 1MB
c.Producer.RequiredAcks = WaitForLocal  // acks=1（默认）
c.Producer.Timeout = 10 * time.Second
c.Producer.Partitioner = NewHashPartitioner
c.Producer.Retry.Max = 3
c.Producer.Retry.Backoff = 100 * time.Millisecond
c.Producer.Return.Errors = true
c.Producer.CompressionLevel = CompressionLevelDefault
```

**Consumer 配置**：

```go
c.Consumer.Fetch.Min = 1
c.Consumer.Fetch.Default = 1024 * 1024  // 1MB
c.Consumer.Retry.Backoff = 2 * time.Second
c.Consumer.MaxWaitTime = 500 * time.Millisecond
c.Consumer.MaxProcessingTime = 100 * time.Millisecond
c.Consumer.Return.Errors = false
c.Consumer.Offsets.AutoCommit.Enable = true  // 自动提交
c.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
c.Consumer.Offsets.Initial = OffsetNewest  // 从最新位置开始
c.Consumer.Offsets.Retry.Max = 3
```

---

## 九、生产环境踩坑实录

### 9.1 案例1：消费者 Rebalance 导致大量重复消费

**现象**：
- 消费者频繁 Rebalance，导致同一批消息被重复消费 3-5 次
- 数据库出现大量重复订单记录

**原因分析**：
- 消费逻辑耗时长（调用外部 API 3-5 秒），超过 `max.poll.interval.ms`（默认 5 分钟）
- 消费者被 Coordinator 认为"假死"，触发 Rebalance

**解决方案**：
1. 将 `max.poll.interval.ms` 调整为 10 分钟
2. 消费逻辑异步化：消费时直接返回，启动 goroutine 处理
3. 业务层实现幂等性：使用订单 ID 作为唯一键

### 9.2 案例2：Partition 数量不足导致扩容无效

**现象**：
- 消费积压严重（LAG > 10 万），增加消费者实例后 LAG 依然不降

**原因分析**：
- Topic 只有 3 个 Partition，但启动了 10 个消费者实例
- 只有 3 个消费者在工作，其他 7 个空闲

**解决方案**：
1. 增加 Partition 数量到 10（**注意：Partition 只能增加不能减少**）
2. 重启消费者，触发 Rebalance 重新分配 Partition

### 9.3 案例3：acks=1 导致数据丢失

**现象**：
- 生产环境发现部分订单消息丢失（约 0.1%）

**原因分析**：
- Producer 配置 `acks=1`，只等待 Leader 确认
- Leader 写入成功后，Follower 尚未同步，Leader 宕机
- 新选举的 Leader 没有这条消息

**解决方案**：
1. 修改 Producer 配置：`acks=all`
2. 修改 Broker 配置：`min.insync.replicas=2`（至少 2 个副本）
3. 开启 Producer 幂等性：`enable.idempotence=true`

### 9.4 案例4：Page Cache 不足导致性能下降

**现象**：
- Broker 机器内存 32GB，JVM 堆设置为 24GB
- Kafka 读写性能低，大量磁盘 IO

**原因分析**：
- JVM 堆占用过多内存，导致 OS Page Cache 不足（只剩 8GB）
- 热数据无法完全缓存在内存，大量磁盘读

**解决方案**：
1. 将 JVM 堆调整为 6GB（Kafka 官方推荐 6-8GB）
2. 留出 26GB 给 OS Page Cache
3. 性能提升 5 倍

### 9.5 案例5：未设置 retention 导致磁盘爆满

**现象**：
- Broker 磁盘使用率达到 100%，无法写入新消息

**原因分析**：
- Topic 未设置 `retention.ms`（保留时间），消息永久保留
- 日志类 Topic 每天产生 100GB 数据，积累半年后磁盘爆满

**解决方案**：
1. 设置 Topic 级别的 `retention.ms=604800000`（7 天）
2. 手动删除旧数据：`kafka-delete-records.sh`
3. 增加磁盘容量

---

## 十、面试高频 20 题

### 1. 介绍一下 Kafka？

**标准答案**（30 秒）：
Kafka 是分布式流式消息队列，具有高吞吐（百万 TPS）、低延迟（ms 级）、持久化的特点。采用发布-订阅模式，常用于异步解耦、削峰填谷、日志采集。核心组件包括 Producer、Broker、Consumer、Topic、Partition。

**追问应对**：
- **为什么快？** 顺序写磁盘 + Page Cache + 零拷贝
- **如何保证高可用？** 副本机制（ISR）+ Leader 选举

### 2. Kafka 如何保证消息不丢失？

**标准答案**：
分三个环节保障：
1. **Producer**：`acks=all` + 重试 + 幂等性
2. **Broker**：`min.insync.replicas=2` + 禁止非 ISR 选举
3. **Consumer**：手动提交 offset + 业务幂等

**追问应对**：
- **acks=all 就一定不丢吗？** 不一定，如果 ISR 只剩一个 Leader，依然可能丢
- **如何实现 Exactly-once？** Producer 幂等性 + 事务 + Consumer 幂等

### 3. Kafka 如何保证消息顺序？

**标准答案**：
1. **同一 Partition 内严格有序**
2. **不同 Partition 之间无序**
3. **全局有序方案**：设置 Partition=1（牺牲并行度）
4. **常见方案**：按业务 key Hash 到同一分区

**追问应对**：
- **如何保证同一订单的消息有序？** 使用订单 ID 作为 message key

### 4. Kafka 为什么这么快？

**标准答案**（3 点必答）：
1. **顺序写磁盘**：避免随机 IO，性能接近内存写入
2. **Page Cache**：依赖 OS 缓存，避免 JVM GC
3. **零拷贝**：sendfile 系统调用，减少 CPU 拷贝和上下文切换

**追问应对**：
- **零拷贝具体原理？** 磁盘 → Page Cache → 网卡（DMA），避免内核态 ↔ 用户态拷贝

### 5. 什么是 ISR？

**标准答案**：
ISR（In-Sync Replicas）是同步副本集合，包含与 Leader 保持同步的所有副本。

**追问应对**：
- **副本如何被踢出 ISR？** 同步落后超过 `replica.lag.time.max.ms`（默认 10 秒）
- **ISR 为空怎么办？** 取决于 `unclean.leader.election.enable`（默认 false，不允许从 OSR 选举）

### 6. 什么是 Rebalance？如何避免？

**标准答案**：
Rebalance 是消费者组内 Partition 重新分配的过程。触发条件包括消费者加入/离开、Partition 数量变化、心跳超时。

**如何避免**：
1. 合理设置 `session.timeout.ms` 和 `max.poll.interval.ms`
2. 消费逻辑异步化，避免长时间阻塞
3. 使用 Static Membership（`group.instance.id`）

**追问应对**：
- **Rebalance 的影响？** 消费暂停 + 重复消费 + 扩散效应

### 7. Kafka 的存储结构是怎样的？

**标准答案**：
- **逻辑上**：Topic → Partition → Message
- **物理上**：Partition → Segment（.log + .index + .timeindex）

**追问应对**：
- **如何查找某个 offset 的消息？** 通过文件名定位 Segment → 在 .index 中二分查找 → 在 .log 中顺序扫描

### 8. HW 和 LEO 是什么？

**标准答案**：
- **LEO**：Log End Offset，副本本地 log 末尾的 offset
- **HW**：High Watermark，ISR 中所有副本都复制到的 offset

**追问应对**：
- **为什么消费者只能读到 HW 之前的数据？** 保证消息已被多个副本确认，避免读到未复制、可能丢失的数据

### 9. Kafka 如何实现高可用？

**标准答案**：
1. **副本机制**：每个 Partition 有多个副本（Leader + Follower）
2. **ISR 机制**：只有 ISR 中的副本才能参与选举
3. **Controller**：负责 Leader 选举和元数据管理

**追问应对**：
- **Controller 挂了怎么办？** 从其他 Broker 中重新选举 Controller

### 10. Kafka 和 RabbitMQ 有什么区别？

**对比表格**：

| 维度 | Kafka | RabbitMQ |
|------|-------|----------|
| **吞吐量** | 百万级 TPS | 十万级 |
| **延迟** | ms 级 | us 级（更低） |
| **消息顺序** | 分区内有序 | 队列内有序 |
| **消息回溯** | 支持（offset） | 不支持 |
| **适用场景** | 大数据、日志、流式计算 | 实时任务、RPC、微服务 |

### 11. Kafka 消费者如何实现负载均衡？

**标准答案**：
通过 **Consumer Group** 实现：
- 同一 Consumer Group 内的消费者共同消费一个 Topic
- 每个 Partition 只能被组内一个消费者消费
- 多个消费者并行消费不同 Partition，实现负载均衡

### 12. Kafka 如何处理消费积压？

**标准答案**：
1. **确认原因**：消费逻辑慢 / Partition 数不够 / 消费者阻塞
2. **临时方案**：增加消费者实例（不超过 Partition 数）/ 消费逻辑异步化
3. **长期方案**：优化消费逻辑 / 增加 Partition 数 / 调整 `max.poll.records`

### 13. Kafka 的 offset 存储在哪里？

**标准答案**：
- **旧版本（0.9 之前）**：存储在 ZooKeeper
- **新版本（0.9 之后）**：存储在 Kafka 内部 Topic（`__consumer_offsets`）

**追问应对**：
- **为什么从 ZK 迁移到 Kafka？** 减少 ZK 压力，提高性能

### 14. Kafka 如何保证幂等性？

**标准答案**：
1. **Producer 幂等性**：开启 `enable.idempotence=true`，Kafka 会为每条消息分配唯一 ID（PID + Sequence Number）
2. **Consumer 幂等性**：业务层实现（唯一键 / 状态机 / 版本号）

### 15. Kafka 的分区策略有哪些？

**标准答案**：
1. **轮询（Round-Robin）**：依次分配到不同 Partition
2. **Hash**：按 message key 的 Hash 值分配
3. **自定义**：实现 `Partitioner` 接口

### 16. Kafka 的压缩算法有哪些？

**标准答案**：
支持 `gzip`、`snappy`、`lz4`、`zstd` 四种。

**推荐**：
- **高吞吐**：`lz4`（压缩比中等，速度最快）
- **高压缩比**：`zstd`（压缩比最高，速度较慢）

### 17. Kafka 的 acks 参数有哪些值？

**标准答案**：
- **acks=0**：不等待确认，性能最高，可能丢失
- **acks=1**：等待 Leader 确认，可能丢失（Leader 挂掉）
- **acks=all（-1）**：等待所有 ISR 确认，最可靠

### 18. Kafka 的事务是如何实现的？

**标准答案**：
Kafka 事务基于 **事务协调器（Transaction Coordinator）** 实现，支持：
1. **原子性写入**：多个消息要么全部成功，要么全部失败
2. **跨分区事务**：可以跨多个 Topic/Partition

**追问应对**：
- **如何开启事务？** 设置 `transactional.id` + 调用 `beginTransaction()` / `commitTransaction()`

### 19. Kafka 如何实现消息去重？

**标准答案**：
1. **Producer 幂等性**：开启 `enable.idempotence=true`
2. **事务**：使用事务性发送
3. **Consumer 去重**：业务层实现（Redis 缓存消息 ID / 数据库唯一键）

### 20. Kafka 的监控指标有哪些？

**标准答案**：
1. **吞吐量**：`MessagesInPerSec` / `BytesInPerSec`
2. **延迟**：`RequestLatencyAvg`
3. **消费积压**：`ConsumerLag`
4. **副本同步**：`UnderReplicatedPartitions`（未完全同步的 Partition 数）
5. **ISR 变化**：`IsrShrinksPerSec` / `IsrExpandsPerSec`

---

## 常用命令

### 创建 Topic

```bash
kafka-topics.sh --create --topic orders --replication-factor 3 --partitions 10 --bootstrap-server localhost:9092
```

### 查看 Topic 详情

```bash
kafka-topics.sh --describe --topic orders --bootstrap-server localhost:9092
```

### 查看消费组情况

```bash
kafka-consumer-groups.sh --describe --group order-group --bootstrap-server localhost:9092
```

### 重置消费 offset

```bash
# 重置到最新位置
kafka-consumer-groups.sh --group order-group --bootstrap-server localhost:9092 --reset-offsets --all-topics --to-latest --execute

# 重置到指定时间
kafka-consumer-groups.sh --group order-group --bootstrap-server localhost:9092 --reset-offsets --all-topics --to-datetime 2024-03-10T00:00:00.000 --execute
```

### 生产消息（测试）

```bash
kafka-console-producer.sh --topic orders --bootstrap-server localhost:9092
```

### 消费消息（测试）

```bash
kafka-console-consumer.sh --topic orders --from-beginning --bootstrap-server localhost:9092
```

---

## 参考资料

1. [Kafka Consumer Rebalance 机制与影响（知乎）](https://zhuanlan.zhihu.com/p/46963810)
2. [Kafka 为什么吞吐量大、速度快？（CSDN）](https://blog.csdn.net/kzadmxz/article/details/101576401)
3. [Kafka 数据可靠性深度解读 / ISR 与副本（CSDN）](https://blog.csdn.net/u013256816/article/details/71091774)
4. [Shopify Sarama 客户端配置参考（config.go）](https://github.com/Shopify/sarama/blob/v1.37.2/config.go)
5. [Kafka 选举机制（掘金）](https://juejin.im/post/6844903846297206797)
6. [简单理解 Kafka 的消息可靠性策略（腾讯云）](https://cloud.tencent.com/developer/article/1752150)
7. [Bootstrap server vs zookeeper in kafka?（StackOverflow）](https://stackoverflow.com/questions/46173003/bootstrap-server-vs-zookeeper-in-kafka)
8. [Kafka 如何保证顺序消费（CSDN）](https://blog.csdn.net/java_atguigu/article/details/123920233)
9. [Kafka 官方文档](https://kafka.apache.org/documentation/)
10. [Kafka-go GitHub](https://github.com/segmentio/kafka-go)
