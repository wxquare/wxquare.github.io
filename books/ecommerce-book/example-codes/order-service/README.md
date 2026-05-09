# Order Service - 三层架构示例

这个示例用于配合《电商系统设计》第一章的“阶段 0：标准三层架构”。

它刻意保持朴素：`application/service` 直接依赖 `infrastructure/persistence` 的具体实现，没有引入 Clean Architecture、DDD 聚合、CQRS 等更复杂的结构。这样读者可以先看清楚三层架构的自然形态，再理解后续为什么需要演进。

## 目录结构

```text
order-service/
├── cmd/
│   ├── order-server/          # 同步服务入口：HTTP / RPC
│   ├── order-job/             # 定时任务入口
│   └── order-consumer/        # 消息消费入口
├── internal/
│   ├── interfaces/            # 表现层：接收外部输入，转换成 application 调用
│   ├── application/           # 业务逻辑层：编排创单、支付、关单、发布事件
│   ├── model/                 # 数据结构：Order、Request、Event
│   ├── infrastructure/        # 数据访问层与基础设施：MySQL、事件总线、日志
│   └── bootstrap/             # 程序启动与依赖组装
└── go.mod
```

## 分层职责

`interfaces` 是入口层。HTTP、RPC、Job、Event 都只是不同的触发方式，它们不写核心业务逻辑。

`application` 是业务逻辑层。创单、支付成功后改状态、超时关单、发布订单事件都在这里编排。

`model` 是共享数据结构。这里的 `Order` 还是偏贫血的数据对象，业务规则主要散落在 `service` 中。

`infrastructure/persistence` 是数据访问层。它隐藏数据读写细节，但在三层架构阶段，`application/service` 仍然直接依赖具体实现。

`infrastructure` 里还放了日志、MySQL 连接和事件总线这类技术组件。

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

它也刻意保留了后续演进的痛点：`application/service` 直接依赖具体基础设施、`model` 只承载数据、事件发布和事务一致性还比较粗糙。这些正是后面引入 Clean Architecture、DDD 和 CQRS 的原因。
