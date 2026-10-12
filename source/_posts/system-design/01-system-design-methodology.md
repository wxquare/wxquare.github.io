---
title: 系统设计：技术方案写作、架构选型与设计评审
date: 2025-04-01
updated: 2026-10-12
permalink: /2025/04/01/system-design/01-system-design-methodology/
description: 配合《系统设计：技术方案与架构设计》课件，从技术方案写作、架构选型、接口、存储、缓存与异步设计到评审，建立可执行的系统设计方法。
categories:
  - 系统设计基础
tags:
  - 技术方案
  - 架构设计
  - API设计
  - 微服务
toc: true
---

技术方案需要说明解决什么问题、选择了什么方案、为什么这样选择，以及如何验证结果。本文配合《系统设计：技术方案与架构设计》课件使用：课件梳理设计步骤与常见选型，正文展开判断依据、实现约束和评审问题。

适合正在编写后端技术方案、准备设计评审，或希望建立系统设计知识框架的工程师。建议先浏览课件，再按当前任务查阅正文；写方案时从目标、约束和验收条件开始，结合接口、存储与失败处理补全设计。

## 配套课件与阅读路径

[在线阅读《系统设计：技术方案与架构设计》公开版课件（PDF，38 页）](/library/presentations/system-design-overview-public.pdf) · [下载可编辑课件（PPTX）](/library/presentations/system-design-overview-public.pptx)

![《系统设计：技术方案与架构设计》公开版课件封面](/diagrams/01-system-design-methodology/31-public-deck-cover.png)

课件署名为 wxquare，原分享日期为 2024-06-12，公开版于 2026-10-12 整理。PDF 由本站直接托管，点击课件链接即可在浏览器中阅读，也可使用浏览器的下载按钮保存。PPTX 为可编辑源文件，PDF 为对应的阅读导出版本。正文持续维护，课件保留原分享时的组织方式。旧课件中的厂商 API 示例、数据库能力对比和经验阈值，需结合正文的修订说明与当前官方资料判断。

| 阅读任务 | 课件页码 | 正文入口与使用方法 |
|---|---|---|
| 明确方案内容与质量标准 | 第 2–5 页 | 从“技术方案应该包含哪些内容”到“技术方案模板”，填写目标、约束、风险与验收条件 |
| 选择架构视图与职责边界 | 第 7–12 页 | 阅读“整体架构设计”，区分业务、应用、数据与技术视图；绘图方法见 {% post_link system-design/07-engineering-diagrams-ai-workflow %} |
| 定义接口契约与失败语义 | 第 14–18 页 | 阅读“接口与应用网关设计”中的 API、错误码与幂等性；明确超时、重试和结果未知时的处理 |
| 根据访问模式选择存储 | 第 20–25 页 | 阅读“中间件和存储”，结合事务、一致性、查询、分片与副本条件评估选型 |
| 设计缓存与回源保护 | 第 27–30 页 | 阅读“缓存设计：位置、更新与回源保护”，明确可接受的陈旧程度与缓存失效时的容量 |
| 设计异步流程与故障恢复 | 第 32–33 页 | 阅读“异步设计：任务状态、重试与补偿”，区分请求受理、业务完成和结果通知 |
| 评估非功能需求与方案取舍 | 第 35–36 页 | 阅读“如何评估技术方案的质量”与“系统设计的权衡”，再用 {% post_link system-design/05-architecture-code-review-checklist %} 检查发布准备 |

课件中的建议应结合具体约束使用。例如，“文档数据库不支持事务”“关系数据库不能水平扩展”不能作为类别级结论；“单表超过 500 万行就拆分”也不是通用门槛。应核对具体产品与版本，依据访问模式、压测结果、数据增长和运维成本决定。接口错误的 HTTP 状态语义与业务错误类型，以本文的“API 错误码设计规范”为准。

## 速查导航

**任务速查**：

- [技术方案模板](#技术方案模板)：问题、目标、决策与发布准备。
- [数学估算](#数学估算)：流量、内存、带宽和存储。
- [API 错误码设计规范](#API-错误码设计规范)：协议状态与业务错误契约。
- [缓存设计](#缓存设计：位置、更新与回源保护)：缓存位置、更新策略与源站保护。
- [异步设计](#异步设计：任务状态、重试与补偿)：任务恢复、重复消费与顺序边界。


---

## 技术方案的目标与写作方法
### 为什么要做设计方案
<p align="center">
  <img src="/diagrams/01-system-design-methodology/01-why-technical-design.png" width=600 height=400>
</p>

技术方案把业务目标转换成可讨论、可验证的设计决策。它需要让业务、研发、测试和运维对系统边界、关键规则及失败处理形成共同理解，也让后续维护者知道当初为什么这样选择。

课件第 3 页强调沟通与协作。实际评审时，图表用于定位组件与流程，文字用于解释约束和取舍，两者需要互相对应。仅列技术栈不足以回答“为什么选择这个方案”，仅描述正常流程也不足以回答“出错后怎样恢复”。

方案还能帮助拆分工作、发现依赖和安排验证，但不能消除估算中的不确定性。对尚未验证的容量、接口能力或迁移成本，应记录假设、验证责任人与未通过时的备选路径。重要决策可以随代码一起维护，避免文档和实现长期脱节。

### 技术方案应该包含哪些内容
<p align="center">
  <img src="/diagrams/01-system-design-methodology/02-what-tech-design-contain.png" width=600 height=400>
  <br/>
</p>

1. 背景：
   - 解决的问题：明确要解决的技术问题和产品问题的具体描述。
   - 难点和挑战：列出可能遇到的难点、挑战和限制条件。
   - 目标和关键指标：明确解决方案的目标和关键指标，例如性能要求、用户体验等。

2. 外部依赖调研
   - 外部服务和组件：列出系统所依赖的外部服务、组件或系统，并描述其功能和接口。
   - 管理和集成策略：说明如何管理和集成外部依赖，包括版本控制、接口规范等。

3. 业界方案调研和对比：
   - 调研结果：调研现有的业界解决方案，并总结其优缺点。
   - 对比分析：比较不同方案之间的特点、适用性和可行性。

4. 整体设计：
   - 业务流程架构图：展示系统的业务流程和组件之间的关系。
   - 系统调用拓扑图：显示系统内部和外部的调用关系。
   - 技术架构图：描述系统的技术架构，包括各个模块、组件和数据流之间的关系。

5. 功能设计：
   - 存储设计：定义系统中数据的存储方式和结构。
   - 接口设计：定义系统的各个模块之间的接口和通信方式。
   - 流程设计：描述系统的各个功能模块的流程和交互方式。
   - 缓存设计：确定系统中需要使用的缓存策略和机制。

6. 非功能设计：
   - 兼容性设计：考虑系统与不同平台、浏览器或设备的兼容性。
   - 稳定性设计：定义系统的容错和恢复机制，确保系统的稳定性和可用性。
   - 扩展性设计：考虑系统的可扩展性，以便在需要时能够方便地扩展功能和容量。
   - 安全设计：定义系统的安全策略和机制，保护用户数据和系统资源。
   - 性能设计：考虑系统的性能需求，并设计相应的优化措施。
   - 部署设计：定义系统的部署架构和流程，包括服务器配置、网络拓扑等。
   - 可维护性设计：考虑系统的可维护性，包括日志记录、错误处理和调试功能。
   - 测试策略和方案：定义系统的测试策略和测试计划，包括单元测试、集成测试和系统测试等。
   - 部署和运维设计：描述系统的部署和运维策略，包括自动化部署、监控和故障处理等。
   - 风险点：识别系统设计中的潜在风险和问题，并提供相应的应对措施。
   - 监控设计和异常处理机制：
   - 监控需求：定义系统的监控需求，包括日志记录、性能监控和错误监控等。
   - 异常处理机制：描述系统对异常情况的处理方式和机制，包括错误提示、异常捕获和处理流程等。

7. 资源清单：
   - 硬件资源：列出系统所需的硬件资源，例如服务器、存储设备等。
   - 软件资源：列出系统所需的软件资源，例如操作系统、数据库等。
   - 人力资源：确定系统开发和维护所需的人力资源，包括开发人员、测试人员等。

8. 任务拆分和排期：
   - 任务拆分：将系统开发和实施过程分解为具体的任务和子任务。
   - 排期计划：为每个任务和子任务确定时间表和优先级。

9. 评审记录：
    - 评审会议记录：记录技术方案评审会议的讨论和决策结果。
    - 修改和改进建议：记录评审过程中提出的修改和改进建议，并记录其处理状态。


### 如何评估技术方案的质量

课件第 5、35 页分别从质量属性和利益相关者角度讨论设计质量。写方案时，应把“高性能”“高可用”“易维护”转换成可验证的场景，并说明指标由谁确认、如何测量。下表中的检查问题是本文整理的评审框架。

| 质量维度 | 方案需要回答的问题 | 验证材料 |
|---|---|---|
| 功能与正确性 | 正常流程、边界条件、并发冲突与权限检查是否完整？ | 业务规则、状态机与验收用例 |
| 可靠性与可用性 | 依赖失效时能提供什么服务？已确认的数据会不会丢失？ | 故障演练、恢复流程与数据核对 |
| 性能与容量 | 在什么流量、数据量与请求分布下满足目标？ | 压测报告、尾延迟与资源使用率 |
| 安全与隐私 | 谁能读写哪些数据？日志、缓存与导出是否泄露敏感信息？ | 威胁分析、授权矩阵与审计记录 |
| 可维护性与演进 | 增加规则、扩容、迁移和回滚分别影响哪些模块？ | 职责边界、兼容策略与发布方案 |
| 成本与体验 | 资源和运维成本能否接受？等待、失败与恢复是否可理解？ | 成本估算、交互契约与结果查询流程 |

其中，可扩展性（scalability）关注负载增长时的容量，可扩充性（extensibility）关注增加功能的难度，两者应分别论证。SLI、SLO 和 SLA 的定义与统计口径见“技术设计基础”。[[1]](#ref-1)

### 技术方案模板
设计文档没有定式。即使如此，笔者参考谷歌设计文档的结构和格式，并结合实际工作经验加以完善。在此提供一个可供新手参考的设计文档模板，您可以使用此文档模板作为思考的基础。通常，无须事无巨细地填写每一部分，不相关的内容直接略过即可。

设计文档应该体现设计决策的合理性，同时也有助于日后迭代设计时，检查最初的假设是否仍然成立。

#### 背景与问题

用简短文字描述现状、需求动机、已有系统的局限和影响范围，让不了解历史的读者也能理解问题。可以链接需求或历史方案，但关键约束仍需在本文交代。

例如：订单创建接口在超时重试后可能生成重复订单，调用方不能确认第一次请求的结果。本次设计要明确同一次业务意图的识别方式、并发控制和结果查询；更换数据库不是预设目标。

#### 难点和挑战

“解决这个问题的难点和挑战”

用几句话说明该设计文档的关键目的，让读者能够一眼得知自己是否对该设计文档感兴趣。 如：“本文描述 Spanner 的顶层设计”


#### 目标和关键指标

继而，使用 Bullet Points 描述该设计试图达到的重要目标，如：

- 可扩展性
- 多版本
- 全球分布
- 同步复制
非目标也可能很重要。非目标并非单纯目标的否定形式，也不是与解决问题无关的其它目标，而是一些可能是读者非预期的、本可作为目标但并没有的目标，如：
- 高可用性
- 高可靠性 如果可能，解释是基于哪些方面的考虑将之作为非目标。如：
- 可维护性： 本服务只是过渡方案，预计寿命三个月，待 XX 上线运行后即可下线
设计不是试图达到完美，而是试图达到平衡。 显式地声明哪些是目标，哪些是非目标，有助于帮助读者理解下文中的设计取舍。

#### 总体设计

“我们如何解决这个问题？”

用一页描述高层设计。说明系统的主要组成部分，以及一些关键设计决策。应该说明该系统的模块和决策如何满足前文所列出的目标。

本设计文档的评审人应该能够根据该总体设计理解你的设计思路并做出评价。描述应该让新加入项目、不了解历史实现的工程师也能理解。

推荐使用系统关系图描述设计。它可以使读者清晰地了解文中的新系统和已经熟悉的系统间的关系。它也可以包含新系统内部概要的组成模块。

注意：不要只放一个图而不做任何说明，请根据上面小节的要求用文字描述设计思想。

建议在总体设计中给出系统边界、核心模块职责与主要交互，并用文字解释关键取舍；具体绘图方法见 {% post_link system-design/07-engineering-diagrams-ai-workflow %}。
不要在这里描述细节，放在下一章节中；不要在这里描述背景，放在上一章节中。

#### 详细设计

在这一节中，除了介绍设计方案的细节，还应该包括在产生最终方案过程中，主要的设计思想及权衡（tradeoff）。这一节的结构和内容因设计对象（系统，API，流程等）的不同可以自由决定，可以划分一些小节来更好地组织内容，尽可能以简洁明了的结构阐明整个设计。

不要过多写实现细节。就像我们不推荐添加只是说明”代码做了什么”的注释，我们也不推荐在设计文档中只说明你具体要怎么实现该系统。否则，为什么不直接实现呢？ 以下内容可能是实现细节例子，不适合在设计文档中讨论：

- **API 的所有细节**
- **存储系统的 Data Schema**
- **具体代码或伪代码**
- **该系统各模块代码的存放位置、各模块代码的布局**
- **该系统使用的编译器版本**

通常可以包含以下内容（注意，小节的命名可以更改为更清晰体现内容的标题）：

**各子模块的设计**
阐明一些复杂模块内部的细节，可以包含一些模块图、流程图来帮助读者理解。可以借助时序图进行展现，如一次调用在各子模块中的运行过程。每个子模块需要说明自己存在的意义。如无必要，勿添模块。如果没有特殊情况（例如该设计文档是为了描述并实现一个核心算法），不要在系统设计加入代码或者伪代码。

**API 接口**

如果设计的系统会暴露 API 接口，那么简要地描述一下 API 会帮助读者理解系统的边界。避免将整个接口复制粘贴到文档中，因为在特定编程语言中的接口通常包含一些语言细节而显得冗长，并且有一些细节也会很快变化。着重表现 API 接口跟设计最相关的主要部分即可。

**存储设计**
介绍系统依赖的存储设计。该部分内容应该回答以下问题，如果答案并非显而易见：

**该系统对数据/存储有哪些要求？**
- 该系统会如何使用数据？
- 数据是什么类型的？
- 数据规模有多大？
- 读写比是多少？读写频率有多高？
- 对可扩展性是否有要求？
- 对原子性要求是什么？
- 对一致性要求是什么？是否需要支持事务？
- 对可用性要求是什么？
- 对性能的要求是什么？

**基于上面的事实，数据库应该如何选型？**
- 选用关系型数据库还是非关系型数据库？是否有合适的中间件可以使用？
- 如何分片？是否需要分库分表？是否需要副本？
- 是否需要异地容灾？
- 是否需要冷热分离？
数据的抽象以及数据间关系的描述至关重要。可以借助 ER 图(Entity Relationshiop) 的方式展现数据关系。

回答上述问题时，尽可能提供数据，将数据作为答案或作为辅助。 不要回答“数据规模很大，读写频繁”，而是回答“预计数据规模为 300T， 3M 日读出， 0.3M 日写入， 巅峰 QPS 为 300”。这样才能为下一步的具体数据库造型提供详细的决策依据，并让读者信服。 注意：在选型时也应包括可能会造成显著影响的非技术因素，如费用。

避免将所有数据定义（data schema）复制粘贴到文档中，因为 data schema 更偏实现细节。

#### 其他方案
“我们为什么不用另一种方式解决问题？”

在介绍了最终方案后，可以有一节介绍一下设计过程中考虑过的其他设计方案（Alternatives Considered）、它们各自的优缺点和权衡点、以及导致选择最终方案的原因等。通常，有经验的读者（尤其是方案的审阅者）会很自然地想到一些其他设计方案，如果这里的介绍描述了没有选择这些方案的原因，就避免读者带着疑问看完整个设计再来询问作者。这一节可以体现设计的严谨性和全面性。

#### 交叉关注点

#### 基础设施
如果基础设施的选用需要特殊考量，则应该列出。如果该系统的实现需要对基础设施进行增强或变更，也应该在此讨论。

#### 可扩展性

你的系统如何扩展？横向扩展还是纵向扩展？注意数据存储量和流量都可能会需要扩展。

#### 安全 & 隐私
项目通常需要在设计期即确定对安全性的保证，而难以事后补足。不同于其它部分是可选的，安全部分往往是必需的。即使你的系统不需要考虑安全和隐私，也需要显式地在本章说明为何是不必要的。

安全性如何保证？
- 系统如何授权、鉴权和审计（Authorization, Authentication and Auditing, AAA）？
- 是否需要破窗（break-glass）机制？
- 有哪些已知漏洞和潜在的不安全依赖关系？
- 是否应该与专业安全团队讨论安全性设计评审？

#### 数据完整性
如何保证数据完整性（Data Integrity）？如何发现存储数据的损坏或丢失？如何恢复？由数据库保证即可，还是需要额外的安全措施？为了数据完整性，需要对稳定性、性能、可复用性、可维护性造成哪些影响？

#### 延迟

声明延迟的预期目标。描述预期延迟可能造成的影响，以及相关的应对措施。

#### 冗余 & 可靠性

是否需要容灾？是否需要过载保护、有损降级、接口熔断、轻重分离？是否需要备份？备份策略是什么？如何修复？在数据丢失和恢复之间会发生什么？

#### 稳定性
SLA 目标是什么？如何监控？如何保证？

#### 外部依赖

你的外部依赖的可靠性（如 SLA）如何？会对你的系统的可靠性造成何种影响？如果你的外部依赖不可用，会对你的系统造成何种影响？除了服务级的依赖外，不要忘记一些隐含的依赖，如 DNS 服务、时间协议服务、运行集群等。

#### 任务拆分和研发排期
描述时间及人力安排（如里程碑）。这有利于相关人员了解预期，调整工作计划。

#### 遗留的问题与未来计划

未来可能的计划会方便读者更好地理解该设计以及其定位。列出当前设计中已知但暂未解决的问题，以及未来可能的优化方向和扩展计划。

[技术方案设计的规范与模板](https://juejin.cn/post/7294443057835425855)


## 技术设计基础
### 如何量化系统指标：SLI、SLO 与 SLA

SLI 是可观测指标，SLO 是在明确统计窗口与有效请求口径下设定的目标，SLA 是包含责任或后果约定的服务协议。内部性能目标不应一概称为 SLA；成功率、延迟达标比例和正确性需要分别定义。[[1]](#ref-1)
#### 可靠性与恢复

分别定义“处理结果正确”“已确认写入不丢失”和“故障后能恢复”的要求。若使用 RPO（可接受的数据丢失窗口）与 RTO（恢复时间目标），应写明故障范围、计时起点和恢复验收方式；副本存在并不等于备份可恢复。
<p align="center">
  <img src="/diagrams/01-system-design-methodology/03-mean-time-between-failures.png" width=600 height=300>
  <br/>
</p>

#### 可用性

优先按有效请求统计成功比例，并说明哪些业务拒绝属于正常结果。维护窗口、客户端取消与依赖超时是否计入，需要在指标契约中明确；仅统计进程存活不能代表用户能够完成业务。[[1]](#ref-1)
<p align="center">
  <img src="/diagrams/01-system-design-methodology/04-available-metric.png" width=600 height=300>
  <br/>
</p>

#### 延迟、吞吐与资源效率

同时观察延迟分位数、吞吐量与资源使用率。平均延迟可能掩盖慢请求；脱离请求大小、读写比例、缓存命中率与并发量的 QPS 也不能直接比较。估算用于提出假设，压测用于检验假设。

#### 可运维性

说明告警触发后由谁处理、如何定位、如何降级或回滚，以及怎样证明恢复完成。可观测性应覆盖业务结果、依赖调用和积压状态。
<p align="center">
  <img src="/diagrams/01-system-design-methodology/05-manageability.png" width=600 height=300>
  <br/>
</p>

### 系统设计的权衡

- **性能与可扩展性的权衡**：提高性能可能需要牺牲一部分可扩展性，因为某些优化可能会引入复杂性或限制系统的扩展性。
- **可维护性与性能的权衡**：某些优化措施可能会降低代码的可读性和可维护性，因此需要在维护性和性能之间进行权衡。
- **时间与成本的权衡**：系统设计需要考虑开发时间和成本，以确保在给定资源限制下实现最佳的设计方案。
- **安全性与用户体验的权衡**：强大的安全措施可能会增加用户的身份验证和授权过程，从而影响用户体验。

#### 用质量属性场景开展评审

课件第 36 页介绍 ATAM（Architecture Tradeoff Analysis Method）。它通过业务驱动、质量属性场景和架构策略识别风险、敏感点与权衡点，而不是给方案算出一个万能总分。正式方法还包括利益相关者参与和场景优先级分析；日常评审可借用这一思路，不能仅凭一张评分表就宣称完成 ATAM。[[2]](#ref-2)

例如，假设电商商品详情允许短暂陈旧，但下单必须校验当前价格与可售库存。评审缓存方案时，就需要分别讨论详情展示和交易确认：前者可以用陈旧数据维持阅读体验，后者需要访问权威事实。缓存保留时间是敏感参数；增加保留时间可能减少回源，同时扩大陈旧窗口，形成需要业务确认的权衡。

#### 用 ADR 留下决策依据

重要选型可单独记录为 ADR（Architecture Decision Record）：写清背景、决策、状态和后果，包含接受的限制。后续替换方案时，保留旧记录并注明被哪项新决策取代。[[3]](#ref-3)

针对上面的例子，一份简短决策可以说明：商品详情采用旁路缓存，结算查询权威数据；选择依据是访问模式和业务容忍度；代价是展示与结算可能存在价格差异，需要交互提示。重新评估的条件是缓存失效后的源站容量不足，或业务要求变为展示即承诺价格。方案比较表应支持这项判断，不能代替判断。

架构取舍的更多背景可参考 [[4]](#ref-4)；图的层次组织见 [[5]](#ref-5)。


### 面向对象系统设计的原则

#### SOLID 原则概述
SOLID 用于检查代码职责与依赖关系，不能单独决定服务如何拆分：

- **单一职责**：聚合由同一类变化驱动的职责，避免不同业务参与者的变化互相牵连。
- **开闭原则**：为已识别的变化提供扩展点，保护稳定部分；不要求所有代码永远不能修改，也不等同于对象不可变。
- **里式替换**：子类型应满足调用方依赖的行为契约，包括前置条件、后置条件与不变量；语法上可继承还不够。
- **接口隔离**：按调用方需要组织接口，避免强迫它依赖无关能力。
- **依赖反转**：高层策略与底层细节依赖合适的抽象，抽象不依赖具体实现；接口应表达策略需要，而不只是包装某个 SDK。[[6]](#ref-6)

#### 其他架构设计原则

此外，我们做架构设计时也会尽量遵循如下一些原则（与上述 SOLID 原则在本质上也是相通的）：

- **正交性**：架构同一层次拆分出的各组件之间，应该尽量保持正交，即彼此职责独立，边界清晰，没有重叠。
- **高内聚**：同一组件内部应该是高度内聚的（cohesive），像是一个不可分割的整体（否则就应该拆开）。
- **低耦合**：不同组件之间应该尽量减少耦合（coupling），既降低相互的变化影响，也能增强组件可复用性。
- **隔离变化**：许多架构原则与模式的本质都是在隔离变化 —— 将预期可能变化的部分都隔离到一块，减少发生变化时受影响（需要修改代码、重新测试或产生故障隐患）的其他稳定部分。

#### 参考阅读

可进一步阅读 [[6]](#ref-6)。


### 互联网系统八大谬论
<p align="center">
  <img src="/diagrams/01-system-design-methodology/06-8-fallacies.png" width=600 height=400>
  <br/>
</p>

- 网络是可靠的；
- 通信时延为0；
- 带宽是无限的；
- 网络是安全的；
- 拓扑不会改变；
- 只有一个管理者；
- 传输成本为0；
- 网络是同构的；[[7]](#ref-7)



### 数学估算
#### 延迟数

下面保留经典系统设计资料中的历史数量级示意，用于理解不同操作的相对成本。实际值会随硬件、负载、网络拓扑和缓存状态变化，不代表当前机器基准；容量与延迟预算应以目标环境测量为准。[[8]](#ref-8)

```text
Latency Comparison Numbers
--------------------------
L1 cache reference                           0.5 ns
Branch mispredict                            5   ns
L2 cache reference                           7   ns                      14x L1 cache
Mutex lock/unlock                           25   ns
Main memory reference                      100   ns                      20x L2 cache, 200x L1 cache
Compress 1K bytes with Zippy            10,000   ns       10 us
Send 1 KB bytes over 1 Gbps network     10,000   ns       10 us
Read 4 KB randomly from SSD*           150,000   ns      150 us          ~1GB/sec SSD
Read 1 MB sequentially from memory     250,000   ns      250 us
Round trip within same datacenter      500,000   ns      500 us
Read 1 MB sequentially from SSD*     1,000,000   ns    1,000 us    1 ms  ~1GB/sec SSD, 4X memory
Disk seek                           10,000,000   ns   10,000 us   10 ms  20x datacenter roundtrip
Read 1 MB sequentially from 1 Gbps  10,000,000   ns   10,000 us   10 ms  40x memory, 10X SSD
Read 1 MB sequentially from disk    30,000,000   ns   30,000 us   30 ms 120x memory, 30X SSD
Send packet CA->Netherlands->CA    150,000,000   ns  150,000 us  150 ms
```
基于上述数字的指标：
* 从磁盘以 30 MB/s 的速度顺序读取
* 以 100 MB/s 从 1 Gbps 的以太网顺序读取
* 从 SSD 以 1 GB/s 的速度读取
* 以 4 GB/s 的速度从主存读取
* 数据中心内每秒有 2,000 次往返

#### traffic estimates
<p align="center">
  <img src="/diagrams/01-system-design-methodology/07-traffic-estimate.png" width=600 height=300>
  <br/>
</p>


#### memory estimates

<p align="center">
  <img src="/diagrams/01-system-design-methodology/08-memory.png" width=600 height=300>
  <br/>
</p>


#### bandwidth estimates
<p align="center">
  <img src="/diagrams/01-system-design-methodology/09-bandwidth.png" width=600 height=300>
  <br/>
</p>


#### storage estimates
<p align="center">
  <img src="/diagrams/01-system-design-methodology/10-storage.png" width=600 height=300>
  <br/>
</p>


### 系统设计核心概念

阅读概念时，优先把它们映射到方案中的实际问题，而不是逐个收集中间件名称。[[8]](#ref-8)

| 设计问题 | 需要理解的概念 | 在方案中的落点 |
|---|---|---|
| 数据能否同时成功或失败？ | 事务、隔离级别、业务不变量 | 事务范围与并发约束 |
| 超时后怎样确认结果？ | 幂等性、重试、状态查询 | 接口契约与恢复流程 |
| 读到旧数据是否可接受？ | 一致性、复制、缓存 | 场景化的陈旧容忍度 |
| 负载增加时哪里先到瓶颈？ | 扩容、分区、热点、背压 | 容量估算与压测 |
| 组件失效后怎样恢复？ | 故障域、降级、备份、观测 | 发布与运维方案 |

## 整体架构设计
### 架构模式与适用边界

架构模式用于组织职责和交互，不应被当作复杂度或先进程度的排名。模块化单体、微服务、分层和事件驱动可以在不同范围内组合使用。[[8]](#ref-8)

| 选择 | 适用判断 | 新增约束 |
|---|---|---|
| 模块化单体 | 业务边界仍在演进，团队希望集中部署 | 仍需约束模块依赖与数据访问 |
| 微服务 | 边界稳定，确有独立发布或扩容需要 | 跨服务契约、观测、失败与数据协调 |
| 分层结构 | 希望分离业务策略与实现细节 | 控制跨层依赖，避免业务规则散落 |
| 事件驱动 | 允许处理延后，存在解耦或多订阅需求 | 发布可靠性、幂等、顺序与积压治理 |

选型记录需说明当前约束与验证结果。不要因为采用微服务就省略模块设计，也不要因为采用单体就忽略容量和故障处理。

### 架构 EA+4A
#### 什么是架构 EA+4A

课件第 7–12 页用 EA（企业架构）及业务、应用、数据、技术四类视图组织说明。本文沿用这一教学分类，帮助读者分开回答不同问题；它不表示所有架构框架都采用相同名称或层级，也不等同于 C4 模型的系统上下文、容器、组件与代码层次。C4 更侧重软件系统的结构缩放，两者可配合使用。[[5]](#ref-5)

| 视图 | 要回答的问题 | 电商订单示例 |
|---|---|---|
| 业务架构（BA） | 业务能力、参与者和规则是什么？ | 创建订单、确认支付、履约与取消的业务边界 |
| 应用架构（AA） | 哪些应用负责这些能力，怎样协作？ | 订单、支付、库存服务的职责与接口 |
| 数据架构（DA） | 哪些数据是权威事实，怎样组织和流转？ | 订单状态、支付记录、库存事实及分析投影 |
| 技术架构（TA） | 系统怎样运行、扩容和恢复？ | 网关、数据库、消息系统与部署故障域 |
<p align="center">
  <img src="/diagrams/01-system-design-methodology/11-what-is-architecture.png" width=600 height=400>
</p>

#### 业务架构

从业务流程和能力边界出发，说明角色、输入、结果和关键规则。先确定“支付成功由谁确认”，再决定是否拆成独立服务；不要直接用数据库或中间件名称代替业务能力。
<p align="center">
  <img src="/diagrams/01-system-design-methodology/12-business-architecture.png" width=600 height=400>
</p>

#### 应用架构

明确模块职责、调用方向和跨应用契约。图区分同步调用与异步事件，并指出超时、失败或重复通知的处理方；服务数量多并不能证明边界合理。
<p align="center">
  <img src="/diagrams/01-system-design-methodology/13-application-architecture.png" width=600 height=400>
</p>


#### 技术架构

说明部署单元、网络边界、容量瓶颈和故障域。图上的“多副本”需要对应复制方式、切换机制与恢复验证，不能直接推导出高可用。
<p align="center">
  <img src="/diagrams/01-system-design-methodology/14-technical-architecture.png" width=600 height=400>
</p>

#### 数据架构

说明数据所有者、事务边界、读写模型与派生数据更新方式。课件的 ODS、DWD、DIM、APP 是数据仓库分层示例，不是每个在线交易系统都必须具备的层；订单服务首先要解决交易事实和一致性，分析链路再按需要建设。

<p align="center">
  <img src="/diagrams/01-system-design-methodology/15-data-architecture.png" width=600 height=400>
</p>

#### 架构设计原则

对每个模块检查职责是否完整、依赖是否必要、变化是否能局部处理。评审至少用一条正常流程和一条失败流程贯穿各视图，确保它们描述同一个系统。具体图型选择见 {% post_link system-design/07-engineering-diagrams-ai-workflow %}。

<p align="center">
  <img src="/diagrams/01-system-design-methodology/16-tech-design-principles.png" width=600 height=400>
</p>

#### 扩展阅读
- [互联网系统设计原则 ](https://www.cnblogs.com/zpbolgs/p/15270837.html)
- [什么是架构绘图？](https://aws.amazon.com/cn/what-is/architecture-diagramming)
- [架构图分类详解](https://juejin.cn/post/7137250779928199181)
- [The Art of Crafting Architectural Diagrams](https://www.infoq.com/articles/crafting-architectural-diagrams/)
- [去哪儿网架构课](https://www.bilibili.com/video/BV1MR4y1F7r7/?vd_source=04e82c25f6ff3d6a1a08a7f1c343987c)



### 单体服务、微服务、Service Mesh
<p align="center">
  <img src="/diagrams/01-system-design-methodology/17-rpc-to-service-mesh.png" width=600 height=350>
  <br/>
  <strong><a href="https://www.zhihu.com/question/56125281">什么是服务治理</a></strong>
</p>

- 单体服务（Monolithic Services）：单体服务是指将整个应用程序作为一个单一的、紧密耦合的单元进行开发、部署和运行的架构模式。在单体服务中，应用程序的各个功能模块通常运行在同一个进程中，并共享相同的数据库和资源。单体服务的优点是开发简单、部署方便，但随着业务规模的增长，单体服务可能变得庞大且难以维护。

- 微服务（Microservices）：微服务是一种将应用程序拆分为一组小型、独立部署的服务的架构模式。每个微服务都专注于单个业务功能，并通过轻量级的通信机制（如RESTful API或消息队列）进行相互通信。微服务的优点是灵活性高、可扩展性好，每个微服务可以独立开发、测试、部署和扩展。然而，微服务架构也带来了分布式系统的复杂性和管理的挑战。

- Service Mesh：Service Mesh是一种用于解决微服务架构中服务间通信和治理问题的基础设施层。它通过在服务之间插入一个专用的代理（称为Sidecar）来提供服务间的通信、安全性、可观察性和弹性的功能。Service Mesh可以提供流量管理、负载均衡、故障恢复、安全认证、监控和追踪等功能，而不需要在每个微服务中显式实现这些功能。常见的Service Mesh实现包括Istio、Linkerd和Consul Connect等。


### 微服务
<p align="center">
  <img src="/diagrams/01-system-design-methodology/18-grpc-overview.svg" width=600 height=350>
  <br/>
  <strong><a href="https://grpc.io/docs/what-is-grpc/introduction">gRPC 概述</a></strong>
</p>

与此讨论相关的话题是 [微服务](https://en.wikipedia.org/wiki/Microservices)，可以被描述为一系列可以独立部署的小型的，模块化服务。每个服务运行在一个独立的线程中，通过明确定义的轻量级机制通讯，共同实现业务目标。<sup><a href=https://smartbear.com/learn/api-design/what-are-microservices>1</a></sup>例如，Pinterest 可能有这些微服务： 用户资料、关注者、Feed 流、搜索、照片上传等。

### 服务发现
**ZooKeeper**
- ZooKeeper是一个开源的分布式协调服务，最初由雅虎开发并后来成为Apache软件基金会的顶级项目。
- ZooKeeper提供了一个分布式的、高可用的、强一致性的数据存储服务。它的设计目标是为构建分布式系统提供可靠的协调机制。
- ZooKeeper使用基于ZAB（ZooKeeper Atomic Broadcast）协议的一致性算法来保证数据的一致性和可靠性。
- ZooKeeper提供了一个类似于文件系统的层次化命名空间（称为ZNode），可以存储和管理数据，并支持对数据的读写操作。
- ZooKeeper还提供了一些特性，如临时节点、顺序节点和观察者机制，用于实现分布式锁、选举算法和事件通知等。

**etcd**
- etcd是一个开源的分布式键值存储系统，由CoreOS开发并后来成为Cloud Native Computing Foundation（CNCF）的项目之一。
- etcd被设计为一个高可用、可靠的分布式存储系统，用于存储和管理关键的配置数据和元数据。
- etcd使用Raft一致性算法来保证数据的一致性和可靠性，Raft是一种强一致性的分布式共识算法。
- etcd提供了一个简单的键值存储接口，可以存储和检索键值对数据，并支持对数据的原子更新操作。
- etcd还提供了一些高级特性，如目录结构、事务操作和观察者机制，用于构建复杂的分布式系统和应用

- [Etcd](https://coreos.com/etcd/docs/latest) 
- [Zookeeper](https://zookeeper.apache.org) 
- [Consul](https://www.consul.io/docs/index.html)
- [grpc](https://grpc.io/docs)

### Service Mesh
<p align="center">
  <img src="/diagrams/01-system-design-methodology/19-istio-service-mesh.svg" width=600 height=600>
  <br/>
  <strong><a href="https://istio.io/latest/about/service-mesh/">service Mesh 是怎么工作的</a></strong>
</p>

### 远程过程调用协议（RPC）
<p align="center">
  <img src="/diagrams/01-system-design-methodology/20-rpc-request-response-flow.png" width=700 height=400>
  <br/>
  <strong><a href="http://www.puncsky.com/blog/2016/02/14/crack-the-system-design-interview">Source: Crack the system design interview</a></strong>
</p>

在 RPC 中，客户端会去调用另一个地址空间（通常是一个远程服务器）里的方法。调用代码看起来就像是调用的是一个本地方法，客户端和服务器交互的具体过程被抽象。远程调用相对于本地调用一般较慢而且可靠性更差，因此区分两者是有帮助的。热门的 RPC 框架包括 [Protobuf](https://developers.google.com/protocol-buffers/)、[Thrift](https://thrift.apache.org/) 和 [Avro](https://avro.apache.org/docs/current/)。

RPC 是一个“请求-响应”协议：

* **客户端程序** ── 调用客户端存根程序。就像调用本地方法一样，参数会被压入栈中。
* **客户端 stub 程序** ── 将请求过程的 id 和参数打包进请求信息中。
* **客户端通信模块** ── 将信息从客户端发送至服务端。
* **服务端通信模块** ── 将接受的包传给服务端存根程序。
* **服务端 stub 程序** ── 将结果解包，依据过程 id 调用服务端方法并将参数传递过去。


## 接入层设计：域名/代理/负载均衡
### 域名系统
#### Amazon Route 53域名系统
<p align="center">
  <img src="/diagrams/01-system-design-methodology/21-aws-route-53.png" width=600 height=400>
  <br/>
  <strong><a href="https://aws.amazon.com/cn/route53">Amazon Route 53 工作原理</a></strong>
</p>
### 域名解析的过程
<p align="center">
  <img src="/diagrams/01-system-design-methodology/22-dns-resolution-process.jpg" width=400 height=400>
  <br/>
  <strong><a href="http://www.slideshare.net/srikrupa5/dns-security-presentation-issa">来源：DNS 安全介绍</a></strong>
</p>

域名系统是把 www.example.com 等域名转换成 IP 地址。域名系统是分层次的，一些 DNS 服务器位于顶层。当查询（域名） IP 时，路由或 ISP 提供连接 DNS 服务器的信息。较底层的 DNS 服务器缓存映射，它可能会因为 DNS 传播延时而失效。DNS 结果可以缓存在浏览器或操作系统中一段时间，时间长短取决于[存活时间 TTL](https://en.wikipedia.org/wiki/Time_to_live)。
* **A 记录（地址）** ─ 指定域名对应的 IP 地址记录。
* **CNAME（规范）** ─ 一个域名映射到另一个域名或 `CNAME` 记录（ example.com 指向 www.example.com ）或映射到一个 `A` 记录。
* **NS 记录（域名服务）** ─ 指定解析域名或子域名的 DNS 服务器。
* **MX 记录（邮件交换）** ─ 指定接收信息的邮件服务.

#### 域名管理服务
- [Route 53](https://aws.amazon.com/route53/)
- [CloudFlare](https://www.cloudflare.com/dns/)

#### 常用命令
- nslookup
- dig

#### 来源及延伸阅读
* [DNS 架构](https://technet.microsoft.com/en-us/library/dd197427(v=ws.10).aspx)
* [Wikipedia](https://en.wikipedia.org/wiki/Domain_Name_System)
* [关于 DNS 的文章](https://support.dnsimple.com/categories/dns/)
* [DNS Technical Reference](https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2008-r2-and-2008/dd197461(v=ws.10))


#### 代理+负载均衡器
#### 正向forward proxy
#### 反向reverse proxy
<p align="center">
  <img src="/diagrams/01-system-design-methodology/23-proxy-server.png" width=800 height=500>
  <br/>
</p>

#### 负载均衡器和反向代理

<p align="center">
  <img src="/diagrams/01-system-design-methodology/24-lvs-nginx-load-balancer.png" width=800 height=800>
</p>

<p align="center">
  <img src="/diagrams/01-system-design-methodology/25-load-balancer.png" width=800 height=400>
  <br/>
</p>

<p align="center">
  <img src="/diagrams/01-system-design-methodology/26-scalable-system-design-patterns.png" width=600 height=400>
  <br/>
  <strong><a href="http://horicky.blogspot.com/2010/10/scalable-system-design-patterns.html">来源：可扩展的系统设计模式</a></strong>
</p>

负载均衡器将传入的请求分发到应用服务器和数据库等计算资源。无论哪种情况，负载均衡器将从计算资源来的响应返回给恰当的客户端。负载均衡器的效用在于:

* 防止请求进入不好的服务器
* 防止资源过载
* 帮助消除单一的故障点
* **SSL 终结** ─ 解密传入的请求并加密服务器响应，这样的话后端服务器就不必再执行这些潜在高消耗运算了。
* 不需要再每台服务器上安装 [X.509 证书](https://en.wikipedia.org/wiki/X.509)。
* **Session 留存** ─ 如果 Web 应用程序不追踪会话，发出 cookie 并将特定客户端的请求路由到同一实例。
* 多个负载均衡器可采用主备或双活方式部署；故障切换还需健康检查、连接处理与剩余容量验证，不能只靠增加实例保证可用。相关判断见 {% post_link system-design/06-system-reliability-engineering %}。

负载均衡器能基于多种方式来路由流量:
* 随机
* 最少负载
* Session/cookie
* [轮询调度或加权轮询调度算法](http://g33kinfo.com/info/archives/2657)
* [四层负载均衡](#四层负载均衡)
* [七层负载均衡](#七层负载均衡器)

#### 四层负载均衡

四层负载均衡通常按 IP、端口与协议等网络层和传输层信息分发连接，不依赖 HTTP 路径等应用层内容。转发可以采用 NAT、直接路由或代理等方式，不能把 NAT 视为所有四层负载均衡的必要实现。网络分层背景见 {% post_link fundamentals/02-computer-network-fundamentals %}。

#### 七层负载均衡器

七层负载均衡器根据[应用层](#接口与应用网关设计)的协议信息决定怎样分发请求。这会涉及路径、请求头和 cookie。它在支持的协议与终止配置下解析请求，再转发给后端；例如根据路径将视频请求和账单请求路由到不同服务。

以损失灵活性为代价，四层负载均衡比七层负载均衡花费更少时间和计算资源，虽然这对现代商用硬件的性能影响甚微。


#### 水平扩展

负载均衡器还能帮助水平扩展，提高性能和可用性。使用商业硬件的性价比更高，并且比在单台硬件上**垂直扩展**更贵的硬件具有更高的可用性。相比招聘特定企业系统人才，招聘商业硬件方面的人才更加容易。

#### 缺陷：水平扩展
* 水平扩展引入了复杂度并涉及服务器复制
* 服务器应该是无状态的:它们也不该包含像 session 或资料图片等与用户关联的数据。
* session 可以集中存储在数据库或缓存中；Redis 的持久化能力取决于配置，Memcached 本身不提供持久化。失效、丢失或驱逐后的处理应由会话政策决定，相关机制见 {% post_link fundamentals/08-redis-principles-and-practice %}。
* 缓存和数据库等下游服务器需要随着上游服务器进行扩展，以处理更多的并发连接。

#### 缺陷：负载均衡器
* 如果没有足够的资源配置或配置错误，负载均衡器会变成一个性能瓶颈。
* 引入负载均衡器以帮助消除单点故障但导致了额外的复杂性。
* 单个负载均衡器会导致单点故障，但配置多个负载均衡器会进一步增加复杂性。

#### 反向代理（web 服务器）
<p align="center">
  <img src="/diagrams/01-system-design-methodology/27-reverse-proxy-architecture.png">
  <br/>
  <strong><a href="https://upload.wikimedia.org/wikipedia/commons/6/67/Reverse_proxy_h2g2bob.svg">资料来源：维基百科</a></strong>
  <br/>
</p>

反向代理是一种可以集中地调用内部服务，并提供统一接口给公共客户的 web 服务器。来自客户端的请求先被反向代理服务器转发到可响应请求的服务器，然后代理再把服务器的响应结果返回给客户端。

带来的好处包括：
- **增加安全性** - 隐藏后端服务器的信息，屏蔽黑名单中的 IP，限制每个客户端的连接数。
- **提高可扩展性和灵活性** - 客户端只能看到反向代理服务器的 IP，这使你可以增减服务器或者修改它们的配置。
- **本地终结 SSL 会话** - 解密传入请求，加密服务器响应，这样后端服务器就不必完成这些潜在的高成本的操作。免除了在每个服务器上安装 [X.509](https://en.wikipedia.org/wiki/X.509) 证书的需要
- **压缩** - 压缩服务器响应
- **缓存** - 直接返回命中的缓存结果
- **静态内容** - 直接提供静态内容
  - HTML/CSS/JS
  - 图片
  - 视频
  - 等等

#### 负载均衡器与反向代理

- 当你有多个服务器时，部署负载均衡器非常有用。通常，负载均衡器将流量路由给一组功能相同的服务器上。
- 即使只有一台 web 服务器或者应用服务器时，反向代理也有用，可以参考上一节介绍的好处。
- NGINX 和 HAProxy 等解决方案可以同时支持第七层反向代理和负载均衡。

#### 不利之处：反向代理
- 引入反向代理会增加系统的复杂度。
- 单独一个反向代理服务器仍可能发生单点故障，配置多台反向代理服务器（如[故障转移](https://en.wikipedia.org/wiki/Failover)）会进一步增加复杂度。

### 来源及延伸阅读
- [反向代理与负载均衡](https://www.nginx.com/resources/glossary/reverse-proxy-vs-load-balancer/)
- [NGINX 架构](https://www.nginx.com/blog/inside-nginx-how-we-designed-for-performance-scale/)
- [HAProxy 架构指南](http://www.haproxy.org/download/1.2/doc/architecture.txt)
- [Wikipedia](https://en.wikipedia.org/wiki/Reverse_proxy)
- [NGINX 架构](https://www.nginx.com/blog/inside-nginx-how-we-designed-for-performance-scale/)
- [HAProxy 架构指南](http://www.haproxy.org/download/1.2/doc/architecture.txt)
- [可扩展性](http://www.lecloud.net/post/7295452622/scalability-for-dummies-part-1-clones)
- [Wikipedia](https://en.wikipedia.org/wiki/Load_balancing_(computing))
- [四层负载平衡](https://www.nginx.com/resources/glossary/layer-4-load-balancing/)
- [七层负载平衡](https://www.nginx.com/resources/glossary/layer-7-load-balancing/)
- [ELB 监听器配置](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-listener-config.html)


## 接口与应用网关设计
<p align="center">
  <img src="/diagrams/01-system-design-methodology/28-meituan-gateway.png" width=600 height=400>
  <br/>
  <strong><a href="https://tech.meituan.com/2021/05/20/shepherd-api-gateway.html">百亿规模API网关服务Shepherd的设计与实现</a></strong>
</p>

将 Web 服务层与应用层（也被称作平台层）分离，可以独立缩放和配置这两层。添加新的 API 只需要添加应用服务器，而不必添加额外的 web 服务器。用于完成基础的:
- 参数校验
- 协议转换等
- 鉴权
- 限流
- 监控、日志
- 熔断（错误率较高时，熔断机制）
- 降级策略（比如电商搜索时正常情况下提供搜索+个性化服务，高负载时仅提供搜索服务，非核心功能降级）
- [Shopee Games API 网关设计与实现](https://www.modb.pro/db/474513)
- [百亿规模API网关服务Shepherd的设计与实现](https://tech.meituan.com/2021/05/20/shepherd-api-gateway.html)
- [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)

### API 架构风格

根据调用方和交互方式选择资源式 HTTP API、GraphQL 或 RPC，并统一认证、版本、错误类型、超时和兼容策略。SOA 描述服务组织方式，不是与 REST、RPC 同一维度的接口协议选项。

### RESTful API

课件第 14 页用资源名称和 HTTP 方法说明接口设计。路径使用名词、统一单复数与大小写，是便于协作的团队约定；HTTP 标准并没有禁止 URI 中出现动词或下划线。需要表达业务动作时，可以建立动作资源，例如订单取消请求，避免把复杂状态转换伪装成随意更新字段。

| 方法 | 典型用途 | 重试与并发注意事项 |
|---|---|---|
| GET | 查询资源或集合 | 保持只读，定义分页与过滤 |
| POST | 创建资源或提交处理请求 | 不天然幂等；重试需要业务幂等契约 |
| PUT | 创建或完整替换指定资源 | 方法语义幂等；仍需处理权限与并发覆盖 |
| PATCH | 按约定补丁格式修改资源 | 不天然幂等；递增等操作重复执行会改变结果 |
| DELETE | 删除指定资源或解除其关联 | 方法语义幂等；重复调用的状态码可以不同 |

HTTP 幂等性针对多次相同请求的预期效果，不要求每次响应相同。PATCH 的具体行为还取决于补丁媒体类型。[[9]](#ref-9)[[10]](#ref-10)

下面是假设接口契约的示例：使用示例域名，PATCH 约定 `application/merge-patch+json`，`If-Match` 用于拒绝基于旧版本的覆盖。幂等键是应用契约，不能仅添加一个头就认为服务端实现了去重。JSON Merge Patch 的格式定义见 [[11]](#ref-11)。

```bash
# 查询订单
curl 'https://api.example.com/v1/orders/order-123'

# 创建订单；重试相同业务意图时复用该键
curl -X POST 'https://api.example.com/v1/orders' \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: checkout-attempt-123' \
  -d '{"cart_id":"cart-123","shipping_address_id":"addr-456"}'

# 修改允许变更的配送备注；实际版本取自查询响应
curl -X PATCH 'https://api.example.com/v1/orders/order-123' \
  -H 'Content-Type: application/merge-patch+json' \
  -H 'If-Match: "order-version-3"' \
  -d '{"delivery_note":"工作日送达"}'
```

对于“取消订单”，还需定义允许取消的状态、重复请求的结果，以及退款或释放库存是否同步完成；URI 命名不能替代业务契约。

### API 错误码设计规范

HTTP 状态码用于表达协议层面的结果，响应体提供稳定的业务错误类型和详情。不要将“所有失败都返回 200”写成通用规范，也不要把旧版厂商示例泛化为当前 API 行为。状态语义应以 HTTP 标准和实际接口契约为准。[[9]](#ref-9)

| 状态码 | 典型语义 | 调用方处理 |
|---|---|---|
| 200 / 201 / 202 | 成功、创建成功、已接受但尚未完成 | 202 需有结果查询或通知机制 |
| 400 | 请求格式或参数非法 | 修正请求后再提交 |
| 401 / 403 | 未提供有效认证 / 拒绝访问 | 按认证或权限政策处理 |
| 404 | 目标资源不存在，或服务端不披露其存在 | 核对资源与访问范围 |
| 409 | 与当前资源状态冲突，例如版本冲突 | 刷新事实并重新决定业务意图 |
| 422 | 内容格式可理解，但无法按请求语义处理 | 按稳定错误类型修正业务输入 |
| 429 | 请求频率受到限制 | 遵守服务端退避要求与客户端预算 |
| 500 / 502 / 503 / 504 | 服务端异常、网关响应无效、服务不可用、网关超时 | 是否重试取决于错误契约与副作用，不能一概重试 |

可采用 RFC 9457 Problem Details，用 `type` 提供稳定问题类别，`title` 提供简述，`status` 与实际 HTTP 状态一致，`detail` 提供本次问题说明。业务错误码无需编码服务组织结构；数字或字符串均可，重点是唯一、稳定、文档化。[[12]](#ref-12)

```http
HTTP/1.1 409 Conflict
Content-Type: application/problem+json

{
  "type": "https://example.com/problems/cart-version-conflict",
  "title": "Cart version conflict",
  "status": 409,
  "detail": "Refresh the cart before requesting a new quote.",
  "currentVersion": 8
}
```

上述 URL 和版本值均为教学示例。错误响应不应暴露堆栈、数据库详情或敏感数据；超时表示结果可能未知，应结合幂等键查询或安全重试，而不能据此认定写入失败。

### 接口幂等性设计

<p align="center">
  <img src="/diagrams/01-system-design-methodology/29-api-idempotence.png">
</p>

#### 幂等性的重要性
- 提高可靠性：在网络不稳定的情况下，客户端可能会重试请求。幂等性确保重复请求不会导致意外的副作用。
- 简化客户端代码：客户端不需要担心重复请求的副作用，从而简化了错误处理逻辑。
- 改善用户体验：确保用户操作的可预测性，避免因重复提交表单等操作导致的错误或重复数据。
#### 怎么实现幂等性

课件第 16 页从源头、请求处理和数据约束三个层面展开。按钮置灰能减少误操作，但网络重试、多实例并发与重复消息仍需服务端保障。

以创建订单为例，本文建议检查以下契约：

1. **识别同一次业务意图**：调用方在重试时复用幂等键，服务端按租户、操作与键限定作用域，并比较请求摘要；同一个键不能悄悄用于不同订单内容。
2. **原子占用与业务写入**：通过唯一约束及事务或受控状态转换处理并发请求，避免“先查询不存在，再分别创建”的竞争。仅有内存去重不能保护多实例。
3. **处理执行中和结果未知**：区分执行中、已成功、已失败与待核实。超时只表示调用方未及时拿到结果，不能据此断言订单或支付失败。带副作用的操作需要明确幂等契约后才适合重试。[[13]](#ref-13)
4. **定义保存与清理策略**：说明键的保留范围、重复请求返回方式和到期后的行为；不能让过期去重记录绕过仍需保持的业务唯一约束。

条件更新和版本号可以阻止不合法状态迁移，但使用锁本身不等于业务幂等。对于支付、发货等外部副作用，还需把业务操作标识传给下游，或通过查询与对账核实结果。



## 中间件和存储
### 如何选择存储组件
<p align="center">
  <img src="/diagrams/01-system-design-methodology/30-how-to-choose-storage.png" width=600 height=400>
</p>

课件第 20–25 页从存储类别、建模、分片和副本展开。选型顺序应从业务不变量和访问模式开始，再验证具体产品的能力。“SQL / NoSQL / NewSQL”只能帮助分类，不能直接推出事务、一致性和可用性结论。例如 MongoDB 支持多文档事务，但配置与代价需要单独评估。[[14]](#ref-14)

| 设计问题 | 需要留下的判断依据 |
|---|---|
| 哪些写入必须同时成功？ | 事务边界、唯一约束、并发冲突与跨系统恢复方案 |
| 数据怎样被读取？ | 查询条件、排序、分页、聚合、访问局部性与索引 |
| 哪些副本或索引可以滞后？ | 权威源、允许陈旧的场景、更新链路与核对方法 |
| 何时需要拆分数据？ | 工作集、读写负载、索引维护、备份恢复时间与增长趋势 |
| 扩容或故障切换影响什么？ | 路由变化、数据迁移、复制延迟和应用兼容性 |

建模可沿用“概念模型 → 逻辑模型 → 物理模型”的顺序：先定义订单、支付及库存事实的关系和规则，再落到表、文档、约束与索引。选择文档嵌套时，要考虑更新粒度和文档增长；选择关系模型时，要考虑关联查询与索引成本。

#### 分区、分片与副本的边界

“超过 500 万行就分表”只是课件中的经验提示，不是容量标准。先用查询计划与压测识别瓶颈，再比较索引优化、归档、扩容和数据拆分。以 PostgreSQL 为例，表分区用于将逻辑表拆成物理分区，是否受益依赖查询与维护方式；分区本身不等于跨节点分片。[[15]](#ref-15)

分片键要同时考虑查询路由、数据倾斜、事务范围与迁移成本。一致性哈希有助于控制节点变化时的键迁移范围，但不能自动消除热点，也不能替代数据迁移协议。

副本设计应回答：从哪里读、复制延迟是否可接受、何时确认写入、谁负责切换、切换后如何避免旧主继续写。不要仅凭“有多个副本”就承诺不丢数据；读写分离、跨区恢复与备份恢复应分别验证。

### 各中间件详细专题

- {% post_link fundamentals/07-mysql-database-practice %}：建表规范、索引优化与数据拆分。
- {% post_link fundamentals/08-redis-principles-and-practice %}：数据结构、缓存策略与锁的边界。
- {% post_link fundamentals/09-kafka-messaging-practice %}：消息语义与性能调优。
- {% post_link fundamentals/10-elasticsearch-search-practice %}：索引设计、查询与检索性能。
- {% post_link fundamentals/11-docker-kubernetes-practice %}：容器网络、部署与运维。

## 缓存设计：位置、更新与回源保护

课件第 27–30 页关注缓存位置、组织方式、读写模式与失效保护。首先明确缓存中的数据能陈旧多久，以及哪些操作必须读取权威数据；再决定是否引入缓存。

### 缓存放在哪里

| 位置 | 适用前提 | 需要承担的代价 |
|---|---|---|
| 本地缓存 | 热点读多写少，能接受实例间短暂差异 | 实例内存、失效广播、冷启动与扩容回源 |
| 共享远程缓存 | 多实例需要复用缓存，网络成本可接受 | 网络延迟、缓存集群故障与源站一致性 |
| 多级缓存 | 已验证单层缓存无法满足需求 | 多层失效顺序、容量治理与更复杂的故障定位 |

共享缓存减少实例间差异，并不保证与数据库一致；本地缓存还会在每个新实例启动时产生冷读。是否采用多级缓存，应结合负载验证。[[16]](#ref-16)

课件中的“双 buffer”和 LRU 解决不同问题：前者用于准备新快照并切换读取版本，后者用于容量不足时选择淘汰对象，可以组合使用。LRU 不负责缓存失效通知，也不能保证数据新鲜。

### 缓存怎样更新

以 Cache-Aside（旁路缓存）为例：读请求先查缓存，未命中时读取数据库并回填；写请求先更新数据库，再使相关缓存失效。即使遵循这个顺序，并发读回填和失效失败仍可能留下旧值，因此旁路缓存不保证强一致性。[[17]](#ref-17)

方案需说明失效失败如何重试、旧版本如何识别、TTL 如何兜底，以及关键业务如何核对事实。“延迟双删”不能脱离读写耗时和失败假设被当作通用一致性保证。对于商品详情，可以接受有界陈旧；对于库存扣减、支付确认，应在权威写入路径检查业务约束。

Write-Through 将缓存更新与存储写入串在一条受控路径上，但仍需明确中途失败如何处理；Write-Behind 先接收写入再异步持久化，还需承担队列可靠性、顺序与恢复责任。选择读写模式之前，应先说明成功响应究竟确认了什么。

### 缓存失效时怎样保护源站

缓存会降低正常回源量，也可能让数据库容量被低估。应测试冷启动、批量过期和缓存不可用时的行为，并限制最大回源负载。[[16]](#ref-16)

- **热点失效**：合并同一键的并发回源，必要时按业务允许的范围返回旧值。
- **集中失效**：错开过期或刷新时间，配合限流与降级；随机 TTL 只能缓解同步过期。
- **不存在的数据**：校验输入，按需要缓存短期空结果；布隆过滤器有误判，并需维护数据变化。
- **热点键与大值**：分别检查访问频率和对象大小；增加分片不能自动拆开一个热点键。

除命中率外，还应观察回源 QPS、回源错误与延迟、热点分布、缓存内存和淘汰情况。更完整的实现细节见 {% post_link fundamentals/08-redis-principles-and-practice %}。

## 异步设计：任务状态、重试与补偿

课件第 32–33 页介绍异步回调、任务队列、发布订阅及恢复机制。异步化能把工作移出同步响应路径，但不会自动增加下游处理能力；持续生产快于消费时，积压仍会增长。

### 区分受理、完成与通知

例如支付请求已经被受理，业务处理可能尚未完成；业务已经完成，回调也可能没有送达。接口、任务状态与用户展示应区分这些阶段。HTTP 202 只代表已接受处理，需要另行定义结果查询或通知契约。[[9]](#ref-9)

外部调用超时或回调丢失时，先把结果记为待核实，再通过业务操作标识查询权威状态。只有确认失败且满足业务条件时，才能决定重试或补偿；不要因为“没收到成功”就重新支付。

### 让任务可以恢复

本文建议在任务设计中明确以下责任：

| 失败位置 | 恢复措施 | 验证信号 |
|---|---|---|
| 业务提交后，消息未发布 | 将业务事实与待发布事件写入同一事务，再由发布器恢复发送 | 未发布事件的数量与最老等待时间 |
| 消费结果已提交，确认失败 | 消费者用业务唯一约束或去重记录识别再次投递 | 重复处理次数与业务结果核对 |
| 外部操作结果未知 | 查询权威状态、对账，必要时进入人工核实 | 待核实任务的年龄与处理记录 |
| 永久业务错误 | 进入明确终态或人工处理流程 | 失败原因分布与人工处理进度 |

第一种方式对应 Transactional Outbox：本地事务保证业务事实和待发布事件一起提交，不代表消息只投递一次；发布器仍可能重复发送，消费者仍需幂等。[[18]](#ref-18)

### 控制重试、顺序与积压

区分可恢复的暂时错误和需要修改输入的业务拒绝。重试应有超时、次数或总时限预算，使用退避和抖动，避免调用链各层同时重试放大流量。重试耗尽后的任务应有可审计的人工处理或补偿入口；跨系统补偿通常是新业务动作，不是数据库回滚。[[13]](#ref-13)

Kafka 的分区日志提供分区内顺序，不等于所有业务消息全局有序。应用还需检查分区键、消费并发、重试路径与业务版本。Kafka 的事务处理语义也不能直接推导出外部支付、数据库和发货副作用都只执行一次。[[19]](#ref-19)

监控除了消费 lag，还要覆盖最老任务年龄、处理速率、重试次数、失败终态和业务结果。只有记录异常日志，无法证明任务最终完成。可靠发布、投影和去重的进一步讨论见 {% post_link system-design/02-clean-architecture-ddd-cqrs %}。

## 框架和引擎
### 工作流引擎与任务编排

引擎用于实现任务控制，不会自动提供业务正确性。选择前需核对持久化、崩溃恢复、版本兼容与运维能力；下面的项目作为调研入口，是否满足需求应以采用版本的官方文档与验证结果为准。

#### goflow
- **技术原理**：goflow 是基于数据流编程模型的工作流引擎，采用有向无环图（DAG）描述任务节点之间的依赖关系。每个节点可独立执行，节点间通过通道（channel）传递数据，充分利用 Go 的并发特性（goroutine、channel）实现高效的任务编排和并行处理。
- **使用场景**：适用于需要复杂任务编排、数据流转、并发处理的场景，如数据处理流水线、ETL、自动化运维流程等。适合对任务依赖关系有明确建模需求的系统。

参考：[goflow GitHub](https://github.com/s8sg/goflow)

#### go-workflow
- **技术原理**：go-workflow 是一款轻量级的工作流引擎，支持任务的定义、调度、状态管理和持久化。通过状态机管理任务流转，支持任务重试、超时、失败回调等机制。底层可集成多种存储后端（如 MySQL、Redis）以保证任务可靠性和可恢复性。
- **使用场景**：适合需要可靠任务编排、长流程管理、任务状态追踪的业务，如订单处理、审批流、异步任务调度等。适用于对任务持久化和容错有较高要求的系统。

参考：[go-workflow GitHub](https://github.com/go-workflow/go-workflow)

---

### 规则引擎与风控、资损、校验

#### gengine
- **技术原理**：gengine 是一款高性能、轻量级的 Go 规则引擎，采用自定义 DSL（领域特定语言）编写规则，支持动态加载和热更新。底层通过 AST（抽象语法树）解析和高效的规则匹配算法，实现复杂业务规则的快速执行。支持规则优先级、条件判断、动作执行等特性。
- **使用场景**：广泛应用于风控决策、资损防控、数据校验、营销策略等需要灵活配置和频繁变更规则的场景。适合对规则实时性和可维护性有较高要求的金融、电商等行业。

参考：[gengine GitHub](https://github.com/bilibili/gengine)

---

### 脚本执行引擎与低代码平台

#### tengo
- **技术原理**：tengo 是一个用 Go 实现的嵌入式脚本语言，语法类似 JavaScript。支持类型安全、垃圾回收、闭包、模块化等特性。可将业务逻辑以脚本形式动态加载和执行，便于扩展和热更新。适合嵌入到 Go 应用中作为业务自定义脚本引擎。
- **使用场景**：适用于低代码平台、动态业务规则、用户自定义脚本、插件系统等场景。可用于实现灵活的业务扩展和快速迭代。

参考：[tengo GitHub](https://github.com/d5/tengo)

#### anko
- **技术原理**：anko 是一个简洁的 Go 脚本解释器，支持基本的脚本语法、变量、函数、流程控制等。可与 Go 代码无缝集成，支持在运行时动态执行脚本。适合对脚本功能要求不高但需要快速集成的场景。
- **使用场景**：适合低代码平台、配置驱动、动态表达式计算、简单自动化脚本等。适用于对性能和安全性有一定要求但业务逻辑相对简单的系统。

参考：[anko GitHub](https://github.com/mattn/anko)


## 好用的规范和工具
### 规范：
- Go编码规范
- api 设计规范
- git 使用规范

### 工具：
- 绘图工具：https://app.diagrams.net/
- plantuml: https://plantuml.com/
- Postman: https://www.postman.com/
- Charles: https://www.charlesproxy.com/
- API 文档生成工具：Swagger
- DBeaver：开源的数据库管理工具，支持多种数据库

## 云原生和服务部署CI/CD
- docker
- Kubernetes [Kubernetes 入门&进阶实战](https://zhuanlan.zhihu.com/p/339008746)
- Jenkins

## 大数据存储和计算
- spark
- spark streaming
- hive
- presto
- Lambda 和 Kappa 架构简介：https://libertydream.github.io/2020/04/12/lambda-%E5%92%8C-kappa-%E7%AE%80%E4%BB%8B/


## 系统稳定性建设

> 详见 {% post_link system-design/06-system-reliability-engineering %}

## 系统设计实践：把设计串成一份订单方案

以下是本文构造的教学案例，不代表某个线上系统的既有实现。假设目标是支持商品浏览、创建订单和异步通知，详情展示允许短暂陈旧，订单金额与支付结果需要按权威事实确认。具体流量、陈旧窗口和恢复目标应由业务与压测共同确定。

1. **明确目标与边界**：把浏览、下单、支付确认分别定义为业务场景，列出权限、重复请求、价格变化和订单取消的验收条件。
2. **组织架构视图**：业务图表达状态变化，应用图表达职责和调用，数据图标明订单与支付事实，技术图标明部署与故障域。
3. **定义接口与存储**：创单支持业务幂等键；本地约束防止重复创建；支付结果未知时进入核实流程，不直接重付。
4. **选择缓存与异步边界**：详情按业务要求缓存，交易路径校验事实；通知失败不否定订单已创建，已提交事件可以恢复发布。
5. **验证失败与发布方案**：覆盖缓存不可用、数据库切换、重复消息、回调丢失和发布回滚，记录预期状态及核对方式。

评审时，先走通“订单创建并完成通知”，再走一遍“订单已创建但响应丢失、客户端重试、通知重复投递”的流程。检查每一步谁持有事实、谁负责恢复、如何观察结果，并把关键取舍记入 ADR。

扩展练习可从短链接、文件同步和消息服务中选择一个，复用同样的输出要求；题目与历史案例线索见 System Design Primer。[[8]](#ref-8) 方案进入开发或发布前，可使用 {% post_link system-design/05-architecture-code-review-checklist %} 复核。

## 参考资料

1. <a id="ref-1"></a> [Implementing SLOs](https://sre.google/workbook/implementing-slos/)；Google SRE，2018。
2. <a id="ref-2"></a> [ATAM: Method for Architecture Evaluation](https://www.sei.cmu.edu/library/atam-method-for-architecture-evaluation/)；Rick Kazman、Mark Klein、Paul Clements，CMU SEI，2000。
3. <a id="ref-3"></a> [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)；Michael Nygard，Cognitect，2011。
4. <a id="ref-4"></a> [系统架构](https://book.douban.com/subject/26938710/)
5. <a id="ref-5"></a> [The C4 model for visualising software architecture](https://c4model.com/)；Simon Brown，官方文档。
6. <a id="ref-6"></a> [The Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)；Robert C. Martin，Clean Coder Blog，2012。
7. <a id="ref-7"></a> [Eight Fallacies of Distributed Computing](https://ably.com/blog/8-fallacies-of-distributed-computing)
8. <a id="ref-8"></a> [System Design Primer](https://github.com/donnemartin/system-design-primer)；Donne Martin，持续维护；历史延迟数字仅用于数量级参考。
9. <a id="ref-9"></a> [RFC 9110: HTTP Semantics](https://www.rfc-editor.org/rfc/rfc9110.html)；IETF，2022。
10. <a id="ref-10"></a> [RFC 5789: PATCH Method for HTTP](https://www.rfc-editor.org/rfc/rfc5789.html)；IETF，2010；访问日期：2026-10-12。
11. <a id="ref-11"></a> [RFC 7396: JSON Merge Patch](https://www.rfc-editor.org/rfc/rfc7396.html)；IETF，2014；访问日期：2026-10-12。
12. <a id="ref-12"></a> [RFC 9457: Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457.html)；IETF，2023。
13. <a id="ref-13"></a> [Timeouts, retries, and backoff with jitter](https://aws.amazon.com/builders-library/timeouts-retries-and-backoff-with-jitter/)；Amazon，Builders’ Library；访问日期：2026-10-12。
14. <a id="ref-14"></a> [Transactions](https://www.mongodb.com/docs/manual/core/transactions/)；MongoDB，官方文档；访问日期：2026-10-12。
15. <a id="ref-15"></a> [Table Partitioning](https://www.postgresql.org/docs/18/ddl-partitioning.html)；PostgreSQL Global Development Group，PostgreSQL 18 文档；访问日期：2026-10-12。
16. <a id="ref-16"></a> [Caching challenges and strategies](https://aws.amazon.com/builders-library/caching-challenges-and-strategies/)；Amazon，Builders’ Library；访问日期：2026-10-12。
17. <a id="ref-17"></a> [Cache-Aside Pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/cache-aside)；Microsoft，Azure Architecture Center；访问日期：2026-10-12。
18. <a id="ref-18"></a> [Transactional outbox pattern](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html)；AWS，Prescriptive Guidance；访问日期：2026-10-12。
19. <a id="ref-19"></a> [Design: Message Delivery Semantics](https://kafka.apache.org/41/design/design/)；Apache Software Foundation，Kafka 4.1 文档；访问日期：2026-10-12。

其他延伸阅读：
- https://github.com/Vonng/ddia
- https://github.com/leewaiho/Clean-Architecture-zh
- https://github.com/donnemartin/system-design-primer
- https://www.oreilly.com/library/view/system-design-on/9781098146887/
- [Building Secure & Reliable Systems](https://sre.google/books/)

- [领域驱动设计DDD在B端营销系统的实践](https://tech.meituan.com/2024/05/27/ddd-in-business.html)
