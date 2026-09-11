---
title: 电商 Agent 项目全景调研报告
date: 2026-09-11
categories:
  - AI 与 Agent
tags:
  - ecommerce
  - agentic-commerce
  - MCP
  - UCP
  - ACP
  - research
toc: true
---

# 电商 Agent 项目全景调研报告

## 执行摘要

电商 Agent 已从“对话式商品搜索”进入三个并行演进方向：

1. 消费者侧：Agent 理解自然语言意图，完成发现、比较、个性化推荐，并在越来越多的渠道中将商品加入购物车或引导至商户结账。
2. 商家侧：Agent 汇总经营数据，生成商品、营销和客服动作，并逐步进入受审批的后台操作。
3. 交易基础设施：MCP、UCP、ACP 等协议试图将商品、购物车、结账、支付与售后能力变为 Agent 可调用的标准工具。

结论是，电商项目不应将“自主下单”视为 Agent 的唯一形态。短期最易证明价值的方向通常是 **商家经营、客服售后、商品运营、履约异常处理和 B2B 采购辅助**；消费者侧则以商品发现、比较和受控购物旅程为主。无论场景，写操作都应经过策略校验、模拟预览和人工确认，再调用既有领域服务。库存、核销、支付、退款和价格等确定性约束不能交由模型自行判断。

建议项目团队从一个高频、可回放、低资金风险的工作流开始，例如“经营摘要与异常归因”“商品目录质量修复”“客服售后辅助”或“B2B 复购与订单状态查询”。应以任务正确率、数据一致性、建议采纳率和人工处理时长下降衡量价值，而不是聊天次数。

## 1. 范围与判断框架

本报告截至 2026 年 9 月 11 日，研究与电商相关、能够调用业务能力或影响交易流程的 Agent 项目。范围包括平台能力、商家运营产品、开放协议、开源实现和面向各类电商企业的工程落地。报告不将普通客服机器人、单纯文案生成器或没有工具调用/工作流能力的聊天界面视为 Agent。

判断一个项目是否值得参考，采用五个维度：

| 维度 | 问题 |
|---|---|
| 业务闭环 | 能否从意图理解走到可验证的业务结果，而不仅是生成文本？ |
| 领域控制 | 是否将商品、库存、价格、履约、支付等约束交给确定性服务，而不是模型猜测？ |
| 权限与审计 | 能否定位请求人、商户、工具调用、审批和最终结果？ |
| 平台可移植性 | 是否依赖单一模型/单一渠道，或能通过标准协议替换和扩展？ |
| 落地成熟度 | 有生产用户、官方文档和明确边界，还是仅为演示仓库？ |

## 2. 市场与项目地图

### 2.1 消费者购物 Agent

**ChatGPT 商品发现与商家结账。** OpenAI 的商品发现能力把复杂意图、商品属性比较和商户选择放在对话中完成。2026 年 3 月的官方说明明确，体验主轴是商品发现及跳转/应用内浏览器中的商家自有结账；实施方案不能再假设平台统一代收银是默认路径。[1] 沃尔玛的 ChatGPT 应用则展示了账户、会员和支付能力可以由零售商在 Agent 渠道中整合。[2]

**Shopify Agentic Storefronts。** Shopify 把面向消费者的 Agent 渠道作为销售渠道管理，覆盖 ChatGPT、Google AI Mode/Gemini、Copilot 和 Meta。商户配置可用商品、政策与归因信息；不同渠道的结账能力并不完全相同。[3] 这说明消费者 Agent 的核心不只是检索，而是商品数据质量、可售性、政策解释、跳转/结账状态与订单归因。

**Amazon Shop Direct 与 Buy for Me。** Amazon 的 Shop Direct 将外部商户商品纳入其购物发现流程；对符合条件的商品，Buy for Me 在用户确认价格、运费和支付信息后，可协助完成商户站点购买。[4] 对平台方的启示是：即使拥有支付与账户体系，跨商户交易仍必须明确商户身份、订单确认、履约、退换货责任和失败补偿。

### 2.2 商家运营 Agent

**Shopify Sidekick。** Sidekick 是较成熟的商家侧范式：在后台利用商店上下文回答经营问题，生成内容并辅助商品、订单和设置相关工作。它把“建议”和“执行”分开，并对写操作设置可见的确认边界。[5] 其可借鉴之处不是对话体验，而是以已授权的后台身份、结构化上下文和可审核动作作为产品前提。

**阿里国际站 Smart Assistant Agent。** 阿里国际站将商品发布、询盘处理、商机开发和经营诊断组合为卖家 Agent 工作流。[6] 淘宝开放平台 Q Lab 进一步公开了 Agent、工作流、知识库、插件和 MCP 的集成形态。[7] 这表明国内电商场景的价值重心同样在商家运营自动化，而不是完全自主的消费者交易。

### 2.3 协议与工具接入

**MCP。** Model Context Protocol 是将外部数据和操作暴露给模型的通用协议；它解决的是模型如何发现并调用工具，不解决业务授权和交易安全本身。官方规范强调服务端授权、资源边界与客户端责任。[8] 对电商系统，MCP 应视为适配层，而不能替代网关、权限服务、风控、幂等和审计。

**UCP。** Google 与 Shopify 提出的 Universal Commerce Protocol 覆盖商品发现、购物车、结账、支付能力发现等问题，并可映射到 API、MCP、A2A 与 AP2。[9] UCP 对跨渠道商品展示和交易编排有战略价值，但标准仍处于早期，不应成为首期 MVP 的外部依赖。

**ACP。** OpenAI 与 Stripe 提出的 Agentic Commerce Protocol 定义商品 feed、订单、委托支付等互动接口。[10] 它适合关注外部 Agent 渠道的商户，但实际渠道规则会持续变化，接入必须以当前平台政策和当地支付/消费者保护要求为准。

## 3. 代表项目对比

| 项目/协议 | 主要用户 | 主要能力 | 成熟度 | 可复用的设计要点 |
|---|---|---|---|---|
| ChatGPT 商品发现 | 消费者 | 意图理解、商品比较、商户跳转 | 生产级渠道能力 | 中：未来可作为外部获客入口 |
| Shopify Agentic Storefronts | 商户/消费者 | 商品供给、渠道分发、归因、部分原生结账 | 生产级平台能力 | 高：参考商品事实、政策、可售性和归因建模 |
| Amazon Shop Direct / Buy for Me | 消费者 | 跨站发现与符合条件的代理购买 | 生产级能力 | 中：参考跨站授权、确认与失败处理，不宜直接复制 |
| Shopify Sidekick | 商家运营 | 经营分析、内容与后台辅助 | 生产级产品 | 高：参考审批式后台助手 |
| 阿里 Smart Assistant / Q Lab | 商家运营/开发者 | 运营工作流、知识库、工具与 MCP | 平台级 | 高：参考国内卖家任务和生态接入 |
| MCP | 开发者/Agent 平台 | 工具发现与调用 | 通用标准 | 高：作为内部工具适配接口 |
| UCP / ACP | 跨渠道商户 | 商品、购物车、结账、支付互操作 | 早期标准 | 中：保持兼容性预留，暂不绑定 |
| `nitin27may/e-commerce-agents` | 开发者 | 商品、订单、定价、评论、库存、客服多 Agent | 演示/原型 | 中：参考工具按领域拆分 |
| `felixhuhao/ecommerce-agent` | 开发者 | FastAPI 编排与 Spring Boot MCP 服务分层 | 演示/原型 | 中：参考异构服务的 Adapter 边界 |
| `Maarmapa/storefront-mcp` | 开发者 | Storefront 工具与敏感工具分级 | 演示/原型 | 高：参考敏感写操作隔离 |

上述三个开源项目分别见 [13]、[14]、[15]。它们主要用于学习架构和提示/工具测试，不能直接承担生产交易。其普遍缺少多租户鉴权、审计留存、价格与库存一致性、异步补偿、配额控制、模型输出评测和事故处置机制。

## 4. 关键能力拆解

一个可落地的电商 Agent 应拆成五层，而非把数据库或内部 RPC 直接交给模型：

1. **交互与编排层**：识别任务、维护会话状态、选择工具、组织结果。这里可以使用任意受支持大模型和 Agent SDK。
2. **策略层**：根据角色、市场、商户、风险等级和动作类型决定可调用工具、字段范围、数据时间窗和是否要求审批。
3. **领域工具层**：每个工具只表示一个清晰业务意图，例如 `search_sellable_deals`、`get_deal_stock_snapshot`、`explain_deal_processing_state`、`draft_deal_update`。工具输入输出使用结构化 schema。
4. **既有领域服务层**：库存、Deal、商品、门店、核销、订单和商户权限仍由当前服务负责确定性校验和状态变更。
5. **审计与评测层**：记录模型版本、提示版本、检索数据版本、工具参数、策略判定、审批主体、执行结果与纠正反馈。

在这五层中，模型只能承担理解、选择、归纳和草拟；库存扣减、价格校验、核销、退款、支付、资格判断和状态迁移必须由领域服务执行。

## 5. 电商 Agent 应用地图

| 领域 | 典型任务 | 价值与成熟度 | 建议的自动化边界 |
|---|---|---|---|
| 导购与搜索 | 需求澄清、商品比较、搭配、礼品推荐、门店推荐 | 体验价值高，生产案例最多 | 返回可验证商品事实；下单前显示价格、库存、费用和商户 |
| 商品目录运营 | 商品属性补全、分类、图片/文案生成、违规检查、feed 修复 | ROI 清晰，适合先行 | 自动生成与校验；发布和价格变更需规则/审批 |
| 营销运营 | 受众洞察、活动草案、优惠策略、素材变体和效果解释 | 高价值但易产生品牌和利润风险 | 自动分析和草拟；预算、折扣、投放须审批 |
| 客服与售后 | 订单查询、物流解释、退换货资格判断、工单摘要 | 数据和流程明确，适合规模化 | 查询与资料收集可自动；退款、补偿和例外人工确认 |
| 商家经营 | 销售/流量诊断、经营日报、缺货预警、异常归因 | 最适合 B2B 后台 Agent | 优先只读；生成工单或计划，不直接变更经营数据 |
| 供应链与采购 | 补货建议、采购单草拟、需求解释、供应商沟通 | 价值高但依赖预测和主数据质量 | 预测/草拟自动化；PO、合同、付款需强审批 |
| 履约与风控 | 延迟预警、异常分流、欺诈调查摘要 | 可降低人工排查成本 | Agent 提供证据；风控拦截和资金动作由规则系统决定 |

优先顺序应由数据可用性、流程确定性和错误成本决定。拥有高质量订单、库存、物流、客服与商品主数据的企业，应先做售后、经营或目录 Agent；流量平台或零售媒体则更适合从导购和营销 Agent 开始。

## 6. 可复用的项目形态

### 6.1 购物顾问 Agent

输入是用户意图、上下文和偏好；输出是带价格、库存、交付时效、退换货和商户信息的候选商品。核心不是开放式推荐，而是将自然语言约束编译为检索过滤、排序特征和比较维度。购物车、结账和支付应使用确定性服务或渠道协议完成。

### 6.2 商家运营 Copilot

输入是经营数据、活动配置、商品目录和异常事件；输出是数据解释、行动建议和待审批草案。Sidekick、Smart Assistant、Agentforce Merchant Agent 都属于此类。它适合从每日摘要、异常解释和内容草拟切入，再逐步增加创建工单、生成活动和批量修改草案。

### 6.3 客服与售后 Agent

输入是订单、物流、会员、政策和历史会话；输出是事实性答复、资格判断依据、工单摘要和下一步流程。其关键设计是把“政策解释”和“最终裁决”分开：模型可以引用规则，退款、补偿、账户风险和争议结论必须由规则引擎或人工审批决定。

### 6.4 Agentic Commerce 接入层

面向外部 Agent 渠道，向其提供商品 feed、可售性、政策、购物车、结账和订单状态能力。UCP、ACP、MCP 和厂商的 Storefront MCP 均可归为此类。实施上要把外部协议视为边缘适配器，避免将渠道模型、支付令牌或平台规则渗透到内部商品和订单领域模型。

## 7. 通用目标架构

```text
Consumer App / Merchant Console / External Agent Channel
          |
          v
Agent API Gateway
  - session and identity
  - user, merchant, market, role context
  - rate limit and audit ID
          |
          v
Agent Orchestrator
  - intent and plan
  - tool selection
  - response composition
          |
          v
Policy and Approval Service
  - allowlist
  - data scope filter
  - risk classification
  - approval ticket
          |
          +------------------------+
          |                        |
          v                        v
Read-only MCP/HTTP Tools      Draft/Command Tools
  - catalog/search               - validate change
  - price and availability       - create approval ticket
  - order and logistics          - preview bulk action
  - customer/service policy
  - merchant analytics
          |                        |
          +-----------+------------+
                      v
Existing domain systems
  catalog, inventory, pricing, order, payment, OMS, CRM, WMS, ERP
```

工程上建议建立独立的 Agent Gateway，而非把模型 SDK 嵌入库存、价格、订单和支付等领域服务。Gateway 负责协议、鉴权传播、提示模板、审计和工具适配；领域服务只暴露经过评审的查询或“预校验/创建审批单”接口。

工具 schema 需要携带并强制校验 `market`、`merchant_id`、`actor_id`、`request_id`、数据时间范围和分页限制。不能由模型任意填写或覆盖这些上下文字段；它们应来自登录态和网关策略。

## 8. 控制与安全要求

| 风险 | 控制措施 |
|---|---|
| Prompt injection | 将商品描述、评论、商户自定义文本、外部网页视为不可信数据；禁止其改变系统策略和工具权限 |
| 越权读数 | 网关从身份上下文注入地区、商户和角色；领域服务再次鉴权，不信任 Agent 传入的 ID |
| 写操作误执行 | “草案 - 预校验 - 审批 - 执行”四阶段；高风险操作要求二次确认和审批人分离 |
| 库存/价格竞态 | Agent 只读快照仅用于解释；真正提交时由领域服务重读、加锁/乐观锁和幂等校验 |
| 工具过度授权 | 工具按动作、字段和资源范围细粒度拆分；默认拒绝；禁止通用 RPC/SQL 工具 |
| 幻觉和过期数据 | 所有结论附带数据时间、工具来源和不确定性；关键数值必须来自结构化工具输出 |
| 敏感信息泄露 | 响应脱敏；提示、工具参数和日志进入分级存储；最小化向模型发送 PII 和券码 |
| 异步操作不可追踪 | 将 `request_id` 与业务任务、审批单、Kafka 事件和最终状态串联；保留可回放审计链 |

OWASP 对 LLM 应用的提示注入、敏感信息泄露和过度自主性风险，与电商 Agent 的风险模型高度相关。[11] MCP 的授权最佳实践同样要求资源服务自行验证令牌受众、范围和调用上下文。[12]

## 9. 评测与运营指标

上线前应先建设离线评测集，至少覆盖 200 个脱敏历史任务。每条样本要有自然语言问题、可调用工具、事实答案、允许的动作、拒绝条件和人工判定规则。WebShop 等基准可用于借鉴任务成功率和工具交互评测方法，但不能替代以本地业务规则、权限和异常路径构建的回放集。[16]

| 指标 | 定义 | MVP 建议门槛 |
|---|---|---|
| 工具选择准确率 | 是否调用正确的领域工具和参数范围 | >= 95% |
| 事实正确率 | 数字、状态、价格、库存、地域和商户归属是否与工具结果一致 | >= 98% |
| 高风险动作拦截率 | 对无权限、缺确认、危险动作的拒绝率 | 100% |
| 建议采纳率 | 人工批准并实际执行的草案占比 | 先建立基线，目标 >= 30% |
| 人工处理时长 | 异常定位或经营分析的中位处理时间 | 相比基线下降 >= 30% |
| 失败可解释率 | 失败响应是否给出可验证原因、下一步和责任域 | >= 90% |
| 单任务成本/时延 | 模型、工具调用成本和端到端延迟 | 以业务 SLA 设上限，按场景分级 |

不应以“回答满意度”单独作为成功指标。对于价格、库存、订单和售后场景，错误但流畅的回答比明确拒绝更危险。

## 10. 后续值得关注的方向

### 10.1 从创业角度

**优先关注“Agent 就绪的商业基础设施”。** 消费端入口、模型和支付网络会被大型平台快速占据，但商品事实、实时可售性、履约约束、商家政策、身份授权与争议记录存在于大量异构企业系统中。为商家提供可验证的商品 feed、价格/库存同步、政策结构化、MCP/UCP/ACP Adapter 和渠道归因，属于可嵌入既有业务系统的基础设施机会。行业调查也表明，商家已在准备 Agent 渠道，但对支付、争议和消费者信任保持高度关注。[17][18]

**面向垂直流程，而不是泛化购物机器人。** B2B 复购、工业备件采购、跨境贸易询盘、企业礼赠、复杂售后和本地服务预约的规则密度高、客单价或人工成本高，适合构建领域 Agent。创业公司应选择具有封闭数据、清晰审批人和可衡量结果的流程，例如“将缺货导致的取消率降低”“将售后工单处理时间缩短”或“提高商品目录合规率”，而不是以“替用户浏览网页”为主要卖点。

**Agent 评测、可观测和治理是独立机会。** 企业真正需要的是可回放的任务集、工具调用追踪、事实归因、权限策略测试、提示注入防护和审批审计。通用 LLM 可观测产品会覆盖一部分需求，但电商特有的价格有效期、库存竞态、促销规则、订单状态、退货资格和支付争议仍需要垂直语义模型。

**商家数据治理将成为分发能力。** Agent 不能稳定消费模糊、过期或彼此矛盾的商品信息。围绕商品属性、图片/规格、适配关系、区域税费、配送时效、退换货政策和评论可信度的治理工具，既服务传统搜索与转化，也服务未来 Agent 渠道。WooCommerce/IDC 对平台领导者的调研同样把开放数据架构和平台柔性列为扩展 Agent 的关键条件。[19]

**谨慎进入三类高噪声赛道。** 通用“购物 Agent”前台会受大型平台的账户、支付、流量和信任优势挤压；无领域数据壁垒的客服机器人容易同质化；直接替用户支付、议价或处理退款的 Agent 则面临更高的牌照、损失赔付、风控和消费者保护成本。Forrester 在 2026 年中指出，大多数所谓 Agentic Commerce 体验仍主要停留在对话式发现与推荐，结账仍多由人主导。[20]

### 10.2 从求职角度

未来 12-24 个月，电商 Agent 相关岗位更可能以既有职能的升级形式出现，而不是统一命名为“Agent Engineer”。重点关注以下角色：

| 角色方向 | 工作重点 | 应建立的能力 |
|---|---|---|
| AI/Agent 产品经理 | 选择高 ROI 工作流、定义人机边界和业务指标 | 电商漏斗、订单/售后规则、实验设计、风险分级 |
| 应用 AI / Agent 工程师 | 编排模型、工具调用、记忆、回退和成本治理 | LLM API、结构化输出、MCP、工作流、异步可靠性 |
| 电商后端与集成工程师 | 将目录、库存、订单、CRM、OMS/WMS 变成受控工具 | API 设计、鉴权、幂等、事件驱动、领域建模 |
| 数据/知识工程师 | 建设商品事实、业务知识、检索与评测数据集 | 数据质量、实体解析、检索、数据血缘、SQL/分析 |
| Agent QA / Evals / AI Ops | 构建回放集、红队测试、监控与事故闭环 | 评测设计、可观测、提示注入测试、统计分析 |
| AI 治理与风控 | 定义权限、审计、隐私、资金与例外处理控制 | 安全工程、合规、流程控制、风控策略 |

求职竞争力来自“把 Agent 放入真实系统后仍可控”的能力，而不是单独展示一个聊天界面。LinkedIn 2026 年劳动力报告显示，要求 AI 素养的职位增长快于整体岗位需求，AI、数据与数字素养正在成为跨职能基础能力。[21]

建议用一个可复核作品证明能力：选择一个公开电商流程，例如订单查询、退货资格判断、目录修复或采购单草拟；实现 5-8 个有 schema 的工具；为每个高风险动作设置拒绝或审批；提供 50-100 条评测样本，并展示事实正确率、工具成功率、成本和失败案例。这样的作品同时覆盖产品、工程、数据和治理，是比提示词集合更有说服力的求职材料。

### 10.3 应持续跟踪的信号

| 信号 | 为什么重要 |
|---|---|
| UCP、ACP、MCP 的版本和渠道支持矩阵 | 决定外部 Agent 能否发现、调用和结账 |
| 支付网络的 Agent 身份、授权与争议规则 | 决定代理购买能否规模化且可追责 |
| 商品 feed、价格/库存、政策和订单状态的标准化 | 决定商户能否被 Agent 正确理解和引用 |
| Agent 归因与激励机制 | 决定品牌是否愿意为 Agent 渠道投入预算 |
| 隐私、消费者保护和跨境合规要求 | 决定个人数据、推荐、自动决策和退款的可行边界 |
| 人类保留控制权与可解释性 | 用户对 Agent 过程可见性和可撤销性的要求会直接影响信任；Visa 的 2026 年研究将其列为扩展的关键条件。[22] |

## 11. 选型建议

| 决策 | 建议 | 理由 |
|---|---|---|
| 编排方式 | 单 Agent + 窄领域工具 | 工具边界清晰，最适合审计、测试和逐步放权 |
| 工具协议 | 内部 API 优先，同时设计 MCP Adapter | 不绑定既有技术栈；为未来模型/客户端兼容预留 |
| 检索 | 保留现有搜索、主数据、规则和分析系统作为事实层 | 结构化约束、实时价格和关系规则不应由向量检索或模型替代 |
| 写操作 | 审批单/命令模式，禁止直写 | 保持领域服务的状态机、幂等、事务和风控语义 |
| 记忆 | 只保存会话和用户显式偏好；不保存未验证业务事实 | 避免过期状态污染交易判断 |
| 外部协议 | 跟踪 UCP/ACP，第二阶段评估 | 标准与渠道政策仍快速变化，首期价值不依赖它们 |
| 开源参考 | 只借鉴分层和工具隔离 | 演示仓库通常未处理生产交易控制面 |

## 12. 最终结论

电商 Agent 的竞争点正在从“能否聊天”转向“能否在严格权限、真实时效、异步流程和交易约束下完成可验证工作”。消费者侧购物 Agent 是长期渠道机会；商家运营、客服售后、目录治理和 B2B 辅助则通常更适合作为企业的首笔投入。

项目应从现有数据质量最好、流程最确定、错误成本可控的工作流切入，在确定性领域系统外构建独立 Agent Gateway。先证明查询、归因和草拟的业务价值，再通过审批式命令安全扩展到后台操作和外部 Agent 渠道。

## Sources

1. OpenAI. “Powering product discovery in ChatGPT.” March 24, 2026. https://openai.com/index/powering-product-discovery-in-chatgpt/
2. OpenAI. “Walmart and ChatGPT.” 2026. https://openai.com/index/walmart-and-chatgpt/
3. Shopify Help Center. “Agentic Storefronts.” Accessed September 11, 2026. https://help.shopify.com/en/manual/online-sales-channels/agentic-storefronts
4. Amazon Staff. “Amazon introduces feeds to make it easier for merchants to reach more customers through AI-powered Shop Direct.” March 11, 2026. https://www.aboutamazon.com/news/retail/amazon-shop-direct-external-stores
5. Shopify Help Center. “Sidekick.” Accessed September 11, 2026. https://help.shopify.com/en/manual/ai-powered-tools/sidekick
6. Alibaba.com Seller Central. “Smart Assistant Agent.” Accessed September 11, 2026. https://seller.alibaba.com/pk/smart-assistant-agent
7. 淘宝开放平台. “AI 生态实验室 / Q Lab.” Accessed September 11, 2026. https://developer.alibaba.com/docs/doc.htm?articleId=121802&docType=1&source=search&treeId=840
8. Model Context Protocol. “Authorization.” Accessed September 11, 2026. https://modelcontextprotocol.io/specification/2025-06-18/basic/authorization
9. Google Developers Blog. “Under the hood: Universal Commerce Protocol.” 2026. https://developers.googleblog.com/under-the-hood-universal-commerce-protocol-ucp/
10. Agentic Commerce Protocol. “Get started.” Accessed September 11, 2026. https://agentic-commerce-protocol.com/docs/commerce/guides/get-started
11. OWASP. “OWASP Top 10 for Large Language Model Applications.” 2025. https://genai.owasp.org/llmrisk/llm01-prompt-injection/
12. Model Context Protocol. “Authorization Best Practices.” Accessed September 11, 2026. https://modelcontextprotocol.io/specification/draft/basic/authorization
13. Nitin May. “e-commerce-agents.” GitHub repository. Accessed September 11, 2026. https://github.com/nitin27may/e-commerce-agents
14. Felix Hu. “ecommerce-agent.” GitHub repository. Accessed September 11, 2026. https://github.com/felixhuhao/ecommerce-agent
15. Maarmapa. “storefront-mcp.” GitHub repository. Accessed September 11, 2026. https://github.com/Maarmapa/storefront-mcp
16. Yao et al. “WebShop: Towards Scalable Real-World Web Interaction with Grounded Language Agents.” NeurIPS 2022. https://arxiv.org/abs/2207.01206
17. Checkout.com. “The State of Commerce 2026.” 2026. https://www.checkout.com/resources/reports/state-of-commerce-2026
18. Visa. “The Trust Opportunity: A New Era of Commerce.” 2026. https://corporate.visa.com/content/dam/VCOM/global/run-your-business/documents/visa-the-trust-opportunity-a-new-era-of-commerce.pdf
19. WooCommerce and IDC. “Agentic Commerce: How Open Architectures Will Shape the Future of Digital Commerce.” 2026. https://woocommerce.com/posts/agentic-commerce/
20. Forrester. “Predictions 2026: Commerce.” 2026. https://www.forrester.com/blogs/predictions-2026-commerce/
21. LinkedIn Economic Graph. “Workforce Report.” 2026. https://economicgraph.linkedin.com/resources/linkedin-workforce-report
22. Visa. “How consumers are thinking about AI and agentic commerce.” 2026. https://corporate.visa.com/en/sites/visa-perspectives/innovation/agentic-commerce.html
