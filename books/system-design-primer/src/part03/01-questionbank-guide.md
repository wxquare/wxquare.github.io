# 第 37 章 系统设计题库：使用说明、双索引与训练路线

本题库把系统设计面试拆成可限时练习、可追问、可复盘的题卡。候选人用题卡训练稳定的判断和表达；面试官用相同结构观察需求理解、权威状态、链路设计和取舍能力。

## 题卡标签、难度与建议用时约定

每张题卡的元信息使用同一套口径，便于候选人选题和面试官配题：

- `能力标签`标记需要展示的可迁移能力，例如领域建模、状态机、容量规划、数据一致性、可观测性或架构演进；每题保留 2 至 4 个标签，避免只罗列中间件名称。
- `场景标签`标记题目的业务和运行条件，例如商品供给、交易主链路、大促峰值、支付回调、故障演练或运营后台；它用于组合题目，不替代能力标签。
- `范围`区分组件级、领域级、系统级和平台级问题。先从范围相近的题开始，再跨域组合。
- `难度`分为`基础`、`中等`和`进阶`：基础题要求能说清核心概念和边界；中等题要求给出主链路、状态和失败处理；进阶题要求量化约束、权衡一致性与性能，并覆盖恢复和演进。
- `建议用时`是一次完整口述练习的参考：基础题约 15 分钟，中等题约 30 分钟，进阶题约 45 分钟，综合案例可延长至 60 分钟。时间不足时优先保留目标、状态、主链路、失败处理和取舍。

<a id="chapter37-candidate-workflow"></a>
## 候选人训练工作流

每次练习先选一张题卡和时长，按下面顺序作答：

1. 澄清目标、核心用户、成功指标、约束和非目标。
2. 估算峰值请求、读写比例、数据增长、延迟预算和容量余量。
3. 先画高层主路径，再说明关键状态放在哪里。
4. 分开讲读路径如何加速、写路径如何保证正确性。
5. 补充超时、重试、幂等、降级、补偿、监控和人工处理。
6. 收束当前方案的边界、代价与下一阶段演进。

开场可先说明：“我先确认目标和约束，再估算规模，画出主链路并拆开读写路径，随后说明失败处理，最后补充演进和取舍。”这能让面试官知道回答结构，也方便被打断后回到主线。

练习结束后，不要只检查是否提到缓存、消息队列或分库分表；应检查每个组件是否有明确职责、权威事实是否清楚、失败后果是否被覆盖。

## 面试官评估工作流

面试官先给出题干中的必要约束，观察候选人是否主动量化不完整信息，再沿候选人自己的假设追问。评估重点是：

| 观察点 | 合格表现 | 深入表现 |
| --- | --- | --- |
| 问题定义 | 先问目标、规模和边界 | 说明不同约束会改变哪些方案 |
| 主链路 | 能顺序说明请求流转 | 明确同步点、异步点和用户可见结果 |
| 状态与数据 | 能给出存储选择 | 区分权威状态、缓存和派生读模型 |
| 风险处理 | 提到常见故障 | 说明幂等、补偿、对账和恢复责任 |
| 取舍与演进 | 能说出优缺点 | 能给出当前边界和逐步演进条件 |

递进追问应优先围绕候选人主动提出的数字、组件和承诺，例如“峰值再放大十倍怎么办”“缓存失效时谁是事实来源”“重试导致重复副作用如何收敛”。

<a id="chapter37-whiteboard-capacity"></a>
## 白板与容量表达

白板从少到多展开，避免一开始堆满组件：

1. 写出目标、核心指标和非目标。
2. 写出容量假设和峰值放大口径。
3. 画入口、应用服务、权威存储、缓存、异步链路和关键下游。
4. 用不同箭头或编号拆读路径与写路径。
5. 在有风险的边上标记超时、限流、幂等、重试或补偿。
6. 在图旁写当前瓶颈、监控指标和演进方向。

容量估算至少覆盖读请求、写请求、峰值并发、读写比例、数据增长、带宽、热点、异步重试流量和外部依赖。估算不需要假装精确到机器数，但必须说明口径如何影响选择：读高时优先设计缓存和读模型，写高时考虑削峰、分片和异步化，热点集中时考虑预热、隔离和限流，消息量大时考虑积压、回压和消费恢复。

可使用稳定句式：

- “这里先把权威数据和读模型分开。”
- “这条链路允许最终一致，但需要补偿和对账兜底。”
- “这个组件解决吞吐或延迟问题，不替代业务幂等。”
- “规模未到阈值前先保持简单实现，同时保留分片键和归档策略。”
- “故障时优先保护核心交易，非核心推荐、通知和营销能力可以降级。”

## 自我复盘方法

每道题结束后按三层复盘：

| 层次 | 检查问题 |
| --- | --- |
| 结构 | 是否按目标、规模、链路、状态、失败、演进的顺序展开？ |
| 内容 | 每个关键结论是否有业务约束或数量级依据？是否说明权威来源和失败处理？ |
| 表达 | 是否先报结构、避免术语堆叠，并在结尾说明取舍和边界？ |

把失分点记录为下一次可验证的动作，例如“下次先写读写比例”“下次补充支付回调的幂等键”，而不是笼统记为“加强可靠性”。复练时改变一个约束并重答，确认模板能够适应变化而不是死记答案。

## 按能力索引

| 能力 | 题卡 |
| --- | --- |
| 需求澄清与范围 | [Q-GEN-SCOPE-01](02-general-questionbank.md#q-gen-scope-01系统设计面试与算法面试的区别是什么)、[Q-GEN-SCOPE-02](02-general-questionbank.md#q-gen-scope-02设计秒杀系统前需要澄清什么) |
| 容量估算 | [Q-GEN-CAP-01](02-general-questionbank.md#q-gen-cap-01如何把容量估算转化为设计依据) |
| 权威状态与一致性 | [Q-GEN-DATA-01](02-general-questionbank.md#q-gen-data-01为什么核心交易状态要有权威来源)、[Q-GEN-DATA-02](02-general-questionbank.md#q-gen-data-02系统设计题中如何回答一致性问题)、[Q-GEN-DATA-03](02-general-questionbank.md#q-gen-data-03为什么核心交易状态通常放在-mysql)、[Q-GEN-DATA-04](02-general-questionbank.md#q-gen-data-04库存扣减如何防止超卖)、[Q-GEN-DATA-05](02-general-questionbank.md#q-gen-data-05缓存和数据库不一致怎么办)、[Q-GEN-DATA-06](02-general-questionbank.md#q-gen-data-06分布式锁能解决所有并发问题吗) |
| 可靠性 | [Q-GEN-REL-01](02-general-questionbank.md#q-gen-rel-01如何解释高可用)、[Q-GEN-REL-02](02-general-questionbank.md#q-gen-rel-02超时重试与幂等如何协作)、[Q-GEN-REL-03](02-general-questionbank.md#q-gen-rel-03什么时候用熔断什么时候用降级) |
| 中间件选型 | [Q-GEN-MW-01](02-general-questionbank.md#q-gen-mw-01什么场景应该引入消息队列)、[Q-GEN-MW-02](02-general-questionbank.md#q-gen-mw-02检索场景为什么不总是使用-elasticsearch)、[Q-GEN-MW-03](02-general-questionbank.md#q-gen-mw-03如何降低-kafka-消息丢失风险)、[Q-GEN-MW-04](02-general-questionbank.md#q-gen-mw-04消息重复时如何保证结果正确)、[Q-GEN-MW-05](02-general-questionbank.md#q-gen-mw-05何时选择-elasticsearch-而不是-mysql) |
| 架构抽象与演进 | [Q-GEN-EVO-01](02-general-questionbank.md#q-gen-evo-01高频系统设计题的底层共性是什么)、[Q-GEN-EVO-02](02-general-questionbank.md#q-gen-evo-02系统设计面试后如何复盘并改进) |

## 按场景索引

| 场景 | 题卡 |
| --- | --- |
| 面试方法与白板表达 | [Q-GEN-SCOPE-01](02-general-questionbank.md#q-gen-scope-01系统设计面试与算法面试的区别是什么)、[Q-GEN-CAP-01](02-general-questionbank.md#q-gen-cap-01如何把容量估算转化为设计依据)、[Q-GEN-EVO-02](02-general-questionbank.md#q-gen-evo-02系统设计面试后如何复盘并改进) |
| 秒杀、库存与交易 | [Q-GEN-SCOPE-02](02-general-questionbank.md#q-gen-scope-02设计秒杀系统前需要澄清什么)、[Q-GEN-DATA-03](02-general-questionbank.md#q-gen-data-03为什么核心交易状态通常放在-mysql)、[Q-GEN-DATA-04](02-general-questionbank.md#q-gen-data-04库存扣减如何防止超卖) |
| 缓存与并发控制 | [Q-GEN-DATA-05](02-general-questionbank.md#q-gen-data-05缓存和数据库不一致怎么办)、[Q-GEN-DATA-06](02-general-questionbank.md#q-gen-data-06分布式锁能解决所有并发问题吗) |
| 异步消息 | [Q-GEN-MW-01](02-general-questionbank.md#q-gen-mw-01什么场景应该引入消息队列)、[Q-GEN-MW-03](02-general-questionbank.md#q-gen-mw-03如何降低-kafka-消息丢失风险)、[Q-GEN-MW-04](02-general-questionbank.md#q-gen-mw-04消息重复时如何保证结果正确) |
| 搜索与复杂查询 | [Q-GEN-MW-02](02-general-questionbank.md#q-gen-mw-02检索场景为什么不总是使用-elasticsearch)、[Q-GEN-MW-05](02-general-questionbank.md#q-gen-mw-05何时选择-elasticsearch-而不是-mysql) |
| 故障治理与演进 | [Q-GEN-REL-01](02-general-questionbank.md#q-gen-rel-01如何解释高可用)、[Q-GEN-REL-02](02-general-questionbank.md#q-gen-rel-02超时重试与幂等如何协作)、[Q-GEN-REL-03](02-general-questionbank.md#q-gen-rel-03什么时候用熔断什么时候用降级)、[Q-GEN-EVO-01](02-general-questionbank.md#q-gen-evo-01高频系统设计题的底层共性是什么) |

## 训练路线

先完成 `Q-GEN-SCOPE-01` 和 `Q-GEN-CAP-01`，建立答题顺序；再选择数据、可靠性和中间件题卡练习权威状态、异常和取舍；最后用 `Q-GEN-EVO-01` 做综合归纳，并以 `Q-GEN-EVO-02` 把失分点转成下一轮训练动作。短时练习先保证结构完整，长时练习再增加容量、追问和演进约束。
