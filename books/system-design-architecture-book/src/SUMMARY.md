# 系统设计与架构实战：原理、工程与电商案例

[前言与使用说明](README.md)

---

# 第一部分：系统设计方法论

- [第 1 章 系统设计完全指南：从问题定义到架构落地](part01/01-system-design-guide.md)
- [第 2 章 技术方案设计方法论：把需求、约束与取舍写清楚](part01/02-technical-design-methodology.md)
- [第 3 章 架构师的组合拳：复杂系统设计的方法论地图](part01/03-architecture-combination.md)
- [第 4 章 业务边界与战略设计：先画对系统的地图](part01/04-business-boundary-strategic-design.md)
- [第 5 章 系统内部结构设计：把复杂业务组织进可演进结构](part01/05-internal-architecture-design.md)
- [第 6 章 系统集成与一致性设计：让系统之间可靠协作](part01/06-integration-consistency-design.md)
- [第 7 章 架构质量保障：用评审、清单和验证守住设计](part01/07-architecture-quality-assurance.md)
- [第 8 章 编码原则与设计模式：用清晰代码承接架构意图](part01/07-coding-principles-design-patterns.md)

---

# 第二部分：核心中间件与基础设施

- [第 9 章 MySQL：存储与数据库](part02/03-mysql-storage-database.md)
- [第 10 章 Redis：缓存原理与实践](part02/04-redis-cache-practice.md)
- [第 11 章 Kafka：消息队列与异步](part02/05-kafka-message-queue-async.md)
- [第 12 章 Elasticsearch：搜索与索引](part02/06-elasticsearch-search-index.md)
- [第 13 章 Kubernetes 与 Docker](part02/07-kubernetes-docker.md)
- [第 14 章 全局 ID 体系与基础服务设计](part02/12-global-id-and-basic-services.md)
- [第 15 章 技术栈选型指南](part02/13-tech-stack-selection.md)

---

# 第三部分：可靠性与工程治理

- [第 16 章 系统可靠性工程](part03/08-system-reliability-engineering.md)
- [第 17 章 对账、补偿、DLQ 与故障恢复](part03/17-reconciliation-compensation-dlq.md)
- [第 18 章 资损防控：资金、库存、优惠与账务安全](part03/18-asset-loss-prevention.md)
- [第 19 章 容量规划、压测、限流、熔断与降级](part03/18-capacity-planning-resilience.md)

---

# 第四部分：电商系统设计实战

- [第 20 章 电商系统全景图](part04/18-ecommerce-overview.md)
- [第 21 章 商品中心系统](part04/19-product-center.md)
- [第 22 章 库存系统](part04/20-inventory-system.md)
- [第 23 章 营销系统](part04/21-marketing-system.md)
- [第 24 章 商品供给管理：运营、库存与生命周期](part04/22-product-supply-ops.md)
- [第 25 章 计价系统设计与实现](part04/23-pricing-system.md)
- [第 26 章 搜索与导购](part04/24-search-discovery.md)
- [第 27 章 购物车与结算](part04/25-cart-checkout.md)
- [第 28 章 订单系统](part04/26-order-system.md)
- [第 29 章 支付系统](part04/27-payment-system.md)
- [第 30 章 供应商数据同步链路](part04/28-supplier-sync.md)
- [第 31 章 商品供给与运营治理平台](part04/29-product-supply-governance.md)
- [第 32 章 B2B2C 平台完整架构](part04/30-b2b2c-platform-architecture.md)
  - [32.1-32.3 业务背景、品类模型与边界设计](part04/30-b2b2c-business-context.md)
  - [32.4-32.5 整体架构与技术选型](part04/30-b2b2c-architecture-decisions.md)
  - [32.6 核心系统设计](part04/30-b2b2c-core-systems.md)
  - [32.7-32.8 完整业务链路与 DDD 战术设计](part04/30-b2b2c-business-flows.md)
  - [32.9-32.14 架构决策、治理与演进](part04/30-b2b2c-governance-evolution.md)

---

# 第五部分：系统设计面试

- [第 33 章 系统设计面试综合](part05/31-system-design-interview-overview.md)
- [第 34 章 中间件与可靠性高频追问](part05/32-middleware-reliability-interview.md)
- [第 35 章 电商架构面试题精选](part05/33-ecommerce-architecture-interview.md)
  - [35.1 电商架构基础题库](part05/33-ecommerce-architecture-interview-foundation.md)
  - [35.2 商品与库存管理题库](part05/33-ecommerce-architecture-interview-product-inventory.md)
    - [35.2.1 商品中心系统题库](part05/33-ecommerce-architecture-interview-product.md)
    - [35.2.2 库存系统题库](part05/33-ecommerce-architecture-interview-inventory.md)
    - [35.2.3 营销与计价系统题库](part05/33-ecommerce-architecture-interview-marketing-pricing.md)
  - [35.3 交易核心链路题库](part05/33-ecommerce-architecture-interview-transaction.md)
    - [35.3.1 搜索与导购题库](part05/33-ecommerce-architecture-interview-search.md)
    - [35.3.2 购物车与结算题库](part05/33-ecommerce-architecture-interview-cart-checkout.md)
    - [35.3.3 订单系统题库](part05/33-ecommerce-architecture-interview-order.md)
    - [35.3.4 支付系统题库](part05/33-ecommerce-architecture-interview-payment.md)
  - [35.4 综合实战案例题库](part05/33-ecommerce-architecture-interview-case-studies.md)
  - [35.5 章节补充、速查与结语](part05/33-ecommerce-architecture-interview-supplement.md)
- [第 36 章 商品、库存、营销与计价专题](part05/34-product-inventory-marketing-pricing-interview.md)
- [第 37 章 搜索、购物车、订单与支付专题](part05/35-search-cart-order-payment-interview.md)
- [第 38 章 白板答辩与容量估算表达](part05/36-whiteboard-capacity-estimation.md)

---

# 第六部分：计算机基础与编码补充

- [第 39 章 操作系统基础](part05/10-operating-system.md)
- [第 40 章 计算机网络实践](part05/11-computer-networking.md)
- [第 41 章 Bash 与 Shell 实用](part05/12-bash-shell-practice.md)
- [第 42 章 Python 实践](part05/13-python-practice.md)
- [第 43 章 C++ 实践](part05/14-cpp-practice.md)
- [第 44 章 Go 语言实践](part05/15-go-practice.md)
- [第 45 章 数据结构与算法题型速查](part06/44-data-structures-and-algorithms.md)

---

# 附录

- [附录 A 术语表](appendix/glossary.md)
- [附录 B 参考文献与外链](appendix/references.md)
- [附录 C 工具与构建说明](appendix/tooling.md)
