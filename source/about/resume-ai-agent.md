# 王贤桂

**AI Agent 工程 / 后端架构**  
电话：137-9858-3528　｜　邮箱：xianguiwang0316@gmail.com　｜　工作经验：8 年

## 职业摘要

8 年后端工程与架构经验，腾讯微信广告 + Shopee 数字电商背景，正在系统化沉淀 AI Agent 工程实践。关注 Agent Runtime、Tool Runtime、RAG、MCP、状态机、多 Agent 编排、Evals、Guardrails 与可观测性；优势不是做 demo，而是把 Agent 放进真实生产系统，处理权限、安全、审计、灰度、监控和故障闭环。

**关键成果**：DoD Agent 自动处理 65%+ 告警｜响应 15 分钟 → <2 分钟｜告警噪音 -70%｜10+ MCP 工具｜5000 万+ 日计价调用｜20,000 QPS 库存｜AI Agent 工程实践 20 章体系化沉淀

## 核心能力

Agent 工程：Agent Runtime、Tool Runtime、MCP、RAG、Memory、状态机、多 Agent 编排｜生产治理：Evals、Guardrails、trace、权限、审计、人工确认、失败回归｜后端底座：Go、Java、MySQL、Redis、Kafka、Elasticsearch、Kubernetes、DDD、Saga、Outbox

## 工作经历

### Shopee｜Digital Purchase 数字电商平台｜Listing & Order 组　2021.07 - 至今

**生产级 AI 告警处理 Agent（DoD Agent）**

面向电商值班和线上故障处理场景，设计 DoD Agent，将告警接入、上下文构建、工具调用、诊断推理、风险分级、人工确认和处置复盘组织成可审计的工程闭环。

- 采用状态机 + ReACT 混合架构：状态机管理 NEW → ANALYZING → WAITING_CONFIRM / ESCALATED → RESOLVED 生命周期，ReACT 负责诊断假设、工具选择和证据汇总，兼顾流程可控与 LLM 推理灵活性。
- 接入 Confluence / Runbook / 历史案例构建 RAG 检索，注册 10+ MCP 工具（Prometheus 指标、K8s Pod、日志检索、Trace、Jira、Seatalk 通知等），要求诊断报告保留来源、证据链和工具调用 trace。
- 按工具风险分级：低风险只读工具自动执行，中风险动作进入人工确认，高风险修复只生成建议；工具层实现 schema 校验、权限校验、超时、幂等、审计和回放，避免把安全边界只写在 Prompt 中。
- 自动处理 65%+ 告警，告警响应时间从 15 分钟降至 <2 分钟，值班工作量降低 60%，通过智能聚合、去重和分级升级将告警噪音减少 70%。

**Agent 生产化工程底座**

将电商系统中长期沉淀的后端工程能力迁移到 Agent Runtime 和 Tool Runtime 设计中，强调可恢复、可追踪、可评估，而不是让多个 Agent 自由对话。

- 使用状态机、任务表、Outbox、Kafka Worker、Saga、幂等键、审计日志和灰度开关处理长任务、工具副作用和失败恢复，为 Agent 执行链路提供确定性控制面。
- 将监控指标、日志、Trace、发布记录、配置变更、订单/库存/支付状态等组织为 Context Package，减少模型凭空判断；通过 eval case、失败 trace 回归和 unsafe action rate 约束上线质量。
- 对多 Agent 协作采用 Coordinator + Worker 思路：Coordinator 负责任务拆解、状态推进和结果汇总，Metrics / Logs / Change / Runbook / Business Worker 各自处理单一职责。

**电商核心系统架构经验**

在 Shopee 负责商品、库存、计价、OTA 与稳定性等核心系统建设，为 AI Agent 生产化提供真实复杂系统背景和工具接入场景。

- 建设统一商品运营平台，面向 10+ 品类抽象 item / sku / 属性 / 库存 / 价格模型，使用策略模式、状态机、Kafka Worker、Outbox 与 Saga 统一上架链路，代码复用率 <10% → 90%+，新品类接入 2 周 → 2 天。
- 设计多品类统一库存系统，Redis Lua 原子扣减支撑秒杀 20,000 QPS、日均 200 万订单；MySQL 权威数据源 + Kafka 持久化 + 小时级对账保障一致性。
- 建设统一价格计算引擎，DDD + 四层计价模型支撑日均 5000 万+ 调用；订单/支付双快照防止价格篡改，空跑比对 + 三阶段灰度迁移 10+ 品类，资损事故 3-5 次/年降至 0。

### 腾讯｜微信广告部｜引擎策略组　2018.09 - 2021.07

- 从 0 到 1 构建 AI 短视频原生广告植入系统，覆盖场景检测、广告位识别、素材融合与投放链路；使用 TVM 做算子融合、量化与推理优化，推理延迟降低 40%，模型体积压缩 60%。
- 优化微信广告调价数据流，将数据上线时间由十分钟级提升至分钟级；建设广告营销中台，对接字节、百度等 10+ 平台，服务内部视频、游戏团队买量需求。

## 技术输出与作品集

- 《AI Agent 工程实践：从大模型基础到生产级智能体系统》：覆盖 LLM 边界、Prompt、Context、Harness、Agent 架构、MCP、RAG、Memory、Evals、Guardrails、可观测性与生产案例。
- DoD Agent、企业知识助手、可观测 Coding Agent、个人知识管理 Agent 等案例沉淀，强调 Agent 与后端系统、权限、审计、知识治理和工具生态的结合。
- 《电商系统架构设计与实现》《程序员系统设计与面试指南》：作为后端架构底座，系统整理电商交易链路、中间件、可靠性工程和系统设计方法论。

## 教育背景

2015.09 - 2018.07　北京大学　计算机应用　硕士  
2011.09 - 2015.07　四川大学　电子信息工程　学士
