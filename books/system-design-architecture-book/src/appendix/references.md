# 附录 B：参考文献与延伸阅读

本附录收录全书写作中反复涉及的基础资料和延伸阅读。正文以工程判断和实践模型为主，本附录用于帮助读者继续深入。

## 系统设计与架构方法

- Martin Kleppmann, *Designing Data-Intensive Applications*。
- Eric Evans, *Domain-Driven Design: Tackling Complexity in the Heart of Software*。
- Vaughn Vernon, *Implementing Domain-Driven Design*。
- Sam Newman, *Building Microservices*。
- Martin Fowler: [Patterns of Enterprise Application Architecture](https://martinfowler.com/books/eaa.html)。
- Martin Fowler: [Microservices](https://martinfowler.com/articles/microservices.html)。
- Martin Fowler: [CQRS](https://martinfowler.com/bliki/CQRS.html)。

## 可靠性、SRE 与工程治理

- Google SRE: [Site Reliability Engineering](https://sre.google/sre-book/table-of-contents/)。
- Google SRE: [The Site Reliability Workbook](https://sre.google/workbook/table-of-contents/)。
- AWS Well-Architected Framework: [Reliability Pillar](https://docs.aws.amazon.com/wellarchitected/latest/reliability-pillar/welcome.html)。
- OpenTelemetry: [Documentation](https://opentelemetry.io/docs/)。
- Prometheus: [Documentation](https://prometheus.io/docs/introduction/overview/)。
- Grafana: [Documentation](https://grafana.com/docs/)。

## 数据库、中间件与基础设施

- MySQL: [Reference Manual](https://dev.mysql.com/doc/refman/8.4/en/)。
- Redis: [Documentation](https://redis.io/docs/latest/)。
- Apache Kafka: [Documentation](https://kafka.apache.org/documentation/)。
- Elasticsearch: [Documentation](https://www.elastic.co/docs)。
- Kubernetes: [Documentation](https://kubernetes.io/docs/home/)。
- Docker: [Documentation](https://docs.docker.com/)。

## 分布式事务、消息与一致性

- Chris Richardson: [Microservices Patterns](https://microservices.io/patterns/index.html)。
- Chris Richardson: [Saga Pattern](https://microservices.io/patterns/data/saga.html)。
- Chris Richardson: [Transactional Outbox](https://microservices.io/patterns/data/transactional-outbox.html)。
- Pat Helland, *Life Beyond Distributed Transactions*。
- Nancy Lynch, Seth Gilbert: *Brewer's Conjecture and the Feasibility of Consistent, Available, Partition-Tolerant Web Services*。

## 电商系统与业务架构

- 商品、库存、计价、营销、订单和支付章节中的模型来自通用电商业务抽象，可结合所在公司的品类、供应商、履约和监管要求调整。
- 资金、优惠、库存和退款相关设计应同时咨询财务、法务、风控和客服团队，不能只由研发侧独立决定。
- 所有涉及支付渠道、银行卡、个人信息、发票、税务和跨境业务的落地方案，都应以当地法律法规和支付机构要求为准。

## 代码与工程实践

- Go: [Documentation](https://go.dev/doc/)。
- Python: [Documentation](https://docs.python.org/3/)。
- C++: [cppreference](https://en.cppreference.com/w/)。
- Bash: [GNU Bash Manual](https://www.gnu.org/software/bash/manual/)。
- Mermaid: [Documentation](https://mermaid.js.org/intro/)。

## 维护建议

- 新增章节时，在本附录补充对应的官方文档、经典书籍或工程案例。
- 外部链接优先选择官方文档、原始论文、作者主页或长期维护的资料页。
- 对强时效内容，例如云产品能力、开源组件版本、支付渠道规范，应在正文中注明“以官方最新文档为准”。
