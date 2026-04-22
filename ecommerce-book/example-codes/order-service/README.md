# Order Service - 三层架构示例

这个示例用于配合《电商系统设计》第一章的“阶段 0：标准三层架构”。

它刻意保持朴素：`service` 直接依赖具体的 `repository` 和 `infra`，没有引入 Clean Architecture、DDD 聚合、CQRS 等更复杂的结构。这样读者可以先看清楚三层架构的自然形态，再理解后续为什么需要演进。

## 目录结构

```text
order-service/
├── cmd/
│   ├── order-server/          # 同步服务入口：HTTP / RPC
│   ├── order-job/             # 定时任务入口
│   └── order-consumer/        # 消息消费入口
├── internal/
│   ├── handler/               # 入口层：接收外部输入，转换成 service 调用
│   ├── service/               # 业务逻辑层：编排创单、支付、关单、发布事件
│   ├── repository/            # 数据访问层：负责订单数据的保存和查询
│   ├── model/                 # 数据结构：Order、Request、Event
│   └── infra/                 # 基础设施：模拟 DB、消息总线、日志
└── go.mod
```

## 分层职责

`handler` 是入口层。HTTP、RPC、Job、Consumer 都只是不同的触发方式，它们不写核心业务逻辑。

`service` 是业务逻辑层。创单、支付成功后改状态、超时关单、发布订单事件都在这里编排。

`repository` 是数据访问层。它隐藏数据读写细节，但在三层架构阶段，`service` 仍然直接依赖具体实现。

`model` 是共享数据结构。这里的 `Order` 还是偏贫血的数据对象，业务规则主要散落在 `service` 中。

`infra` 是基础设施。示例中使用 MySQL 做订单持久化，事件总线仍用内存实现来简化演示。

## 运行示例

```bash
export ORDER_MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/order_service?parseTime=true&charset=utf8mb4&loc=Local'
go run ./cmd/order-server
go run ./cmd/order-job
go run ./cmd/order-consumer
```

`ORDER_MYSQL_DSN` 不配置时会使用上面的默认 DSN。程序启动时会自动建表：

- `orders`
- `order_id_seq`

## 学习重点

这个示例展示了三层架构在早期项目中的优点：结构直观、上手快、调用链短。

它也刻意保留了后续演进的痛点：`service` 直接依赖具体基础设施、`model` 只承载数据、事件发布和事务一致性还比较粗糙。这些正是后面引入 Clean Architecture、DDD 和 CQRS 的原因。
