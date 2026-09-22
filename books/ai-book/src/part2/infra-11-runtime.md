# 第11章 Agent Runtime

## 如何让 Agent 行为可执行、可恢复、可回放？

当模型需要读取知识、调用工具、执行代码、等待外部事件或完成多步任务时，单次模型 RPC 已经不足以表达系统。Agent Runtime 需要管理计划、上下文、工具、状态、重试、人工接管、权限和成本；它既不能把所有逻辑交给模型自由生成，也不能把每个业务流程硬编码成不可演进的脚本。本章讨论如何把 ReAct 的思考—行动循环、工具协议、持久化工作流和后端分布式系统原则组合成一个可观测、可恢复、可治理的运行时。

运行时的关键对象不是一段 prompt，而是 task、run、step、message、tool call、approval、artifact、memory 和 event。每个对象有版本、状态、所有者、租户和生命周期。一次 Agent 任务可能调用多个模型、多个工具和多个外部服务，失败后需要从正确的边界恢复，不能简单把整段对话重新发送。Temporal、LangGraph、Ray、AutoGen、MCP、Saga 和幂等消息模式分别提供了工作流、图、分布式 actor、对话编排、工具协议、补偿事务和消息一致性的参考 [5][6][7][8][12][18][19]。

## 11.1 Runtime 的对象模型与边界

### Task、Run 与 Step

Task 是用户期望完成的目标，Run 是一次实际执行，Step 是运行时对模型、工具、检索、审批或子任务的一次调用。一个 Task 可以有多个 Run，例如用户重试、系统恢复或版本对比；一个 Run 由可持久化的 step DAG 或状态图组成。将 Task 与 Run 混为一个 id，会让重试和计费无法区分，也会把不同版本的执行历史覆盖。

每个 Run 保存入口请求、模型与工具版本、租户、预算、deadline、权限、状态、父任务和 trace id。Step 保存输入引用、输出引用、尝试次数、开始结束时间、资源、错误、幂等键和补偿动作。大文本不直接嵌在状态表中，而是存 artifact 或 message store，通过 hash 和版本引用，避免状态膨胀。

### Event 与状态机

运行时的事实由追加式事件组成，例如 RUN_CREATED、STEP_STARTED、MODEL_COMPLETED、TOOL_REQUESTED、TOOL_SUCCEEDED、APPROVAL_REQUIRED、RETRY_SCHEDULED、RUN_PAUSED 和 RUN_COMPLETED。当前状态由事件派生或由带版本的状态快照保存。重复事件、乱序事件和迟到 worker 回报必须通过 sequence、attempt 和状态版本校验。

状态机定义可接受迁移：RUNNING 可以进入 WAITING_TOOL、WAITING_APPROVAL、PAUSED、FAILED 或 COMPLETED；COMPLETED 不能被迟到的失败事件覆盖；CANCELLED 的新工具动作必须被拒绝。AWS Step Functions 和 Temporal 展示了把工作流状态、重试与持久化分离的思路 [7][20]。平台不一定使用它们，但必须具备同等明确的状态语义。

### 控制面与执行面

控制面负责创建 Run、选择模型和工具、持久化状态、发放租约、调度 Step、处理审批、重试和恢复；执行面负责具体模型请求、工具调用、沙箱、检索和流式输出。执行 worker 可以无状态重启，控制面通过 durable state 让它从租约和 step attempt 恢复。worker 不应私自修改 Run 的全局状态，只提交带版本的事件。

控制面和执行面的边界也决定安全。控制面验证用户权限、工具 allowlist、预算和模型能力；执行面使用短期凭证访问具体资源。即使模型生成了一个看似合理的工具调用，也必须经过控制面 schema、租户、目标资源和审批检查，不能由 worker 直接执行。

## 11.2 规划循环与模型调用调度

### ReAct 不是生产状态机

ReAct 把推理和行动交替组织起来，让模型根据观察结果继续计划 [1]。在生产中，模型输出的 thought、action、observation 需要被解析、校验、持久化和限制；不能将自由文本直接当作状态迁移。运行时把模型调用视为一个产生候选动作的 step，动作经过 schema、权限和预算检查后才进入工具执行。

循环要有最大 step、最大 wall-clock、最大 token、最大工具调用、最大并行度和重复检测。模型可能反复调用同一个工具、不断改写计划或把错误观察当成事实。运行时维护 step fingerprint、工具参数 hash 和 observation hash，检测循环并触发重新规划、降级或人工接管。Toolformer 的工具调用学习说明模型可以学会插入 API 调用，但运行时仍需把调用当作不可信输出 [2]。

### 计划与执行分离

复杂任务可以先产生结构化 plan，再由调度器执行；简单任务可以边思考边行动。显式 plan 有利于审计、并行、预算和人工审批，但计划本身可能过时，外部世界会在执行中变化。运行时保存 plan version、依赖、前置条件、可重试性和补偿动作，执行后根据 observation 允许有限重规划。

计划步骤分为纯计算、幂等读、可重试写、不可重试副作用和人工审批。不同类别的 retry、timeout 和 rollback 不同。Saga 的长事务思想指出，跨多个服务的业务动作通常无法使用一个全局事务，需要为每一步定义补偿或接受不可逆状态 [18]。Agent Runtime 必须将这种分类写进工具注册与工作流定义。

### 并行与依赖

无依赖的检索、文件读取或模型候选可以并行，依赖同一状态的写操作必须串行。并行执行降低总延迟，却增加 token、连接、资源和失败组合。运行时为每个分支分配 child step、预算和 trace，只有所有必需分支满足 join 条件才继续；可选分支失败可以降级，但必须记录。

并行限制至少包括全局、租户、Run、工具和目标资源。一个用户让 Agent 启动很多并发工具调用，可能打爆数据库或第三方 API；工具自身也可能有限流。调度器使用 semaphore、租约和 deadline，超过容量进入队列或明确失败。Ray actor 和任务模型可作为分布式执行参考，但状态与副作用仍由 Runtime 负责 [8]。

## 11.3 工具协议、Schema 与权限

### 工具注册表

工具注册表保存 name、version、description、input schema、output schema、side effects、timeout、retry、rate limit、required scopes、data region、owner、sandbox 和 compensation。工具版本不可变，升级生成新版本；模型看到的工具列表来自当前租户和 Run 的权限，而不是全局目录。MRKL 将模型与外部模块结合的思想说明，路由和工具能力需要显式边界 [3]。

工具描述过长会增加上下文和选择歧义，过短会让模型参数错误。注册表为模型生成规范化 schema 与示例，同时保留机器校验的 JSON Schema。OpenAI function calling、MCP 和 JSON Schema 分别提供调用协议、资源/工具发现和参数契约的参考 [11][12][13]。

### 调用前校验

模型生成的名称、参数、目标资源和用户意图都不可信。调用前执行 schema 类型、必填字段、范围、正则、引用存在性、租户权限、目标资源 allowlist、数据区域、预算和审批校验。校验失败可以让模型修复参数，但修复次数有限且不能绕过权限。错误信息返回给模型时要避免泄露内部凭证、系统路径和安全策略细节。

工具调用在执行前生成 operation id 和 idempotency key，持久化请求摘要与状态。重复提交首先查询原 operation，而不是重新执行。Idempotent Receiver 模式适合处理至少一次消息投递 [19]；业务工具还需要自身支持幂等或提供查询接口。没有幂等保证的副作用应强制人工确认或禁止自动重试。

### 沙箱与执行隔离

代码执行、浏览器、文件系统、数据库和网络工具应按风险使用不同隔离级别。容器、微 VM、进程沙箱、网络 egress allowlist、CPU/内存/时间限制和临时文件系统共同构成执行边界。Kubernetes 可以提供资源和命名空间基础，但不能替代工具级权限、凭证和审计 [9]。

工具 worker 使用短期、最小权限凭证，输出进行大小、类型、敏感信息和恶意内容检查。文件路径不能由模型直接决定宿主机路径；SQL 工具使用只读或参数化接口；浏览器工具限制域名、下载和登录状态。工具返回的文本标记来源和可信级别，进入下一轮上下文时要防止 prompt injection。

### MCP 与协议适配

MCP 将工具和资源暴露为可发现、可调用的协议对象 [12]。Runtime 需要在协议适配层统一超时、取消、权限、审计和错误，而不是把每个 MCP server 的行为直接暴露给模型。server 能力、版本、schema、资源范围和安全声明进入注册表；调用时建立 parent trace 和 operation id。

协议适配要处理 server 断线、能力变化、版本不兼容、流式结果、分页和取消。发现到的工具不能自动获得全局权限，租户和 Run 仍需选择 allowlist。MCP 的可组合性提高了工具生态，但也扩大了供应链与数据边界，运行时需要签名、来源、审批和速率限制。

## 11.4 持久化工作流、重试与补偿

### Durable execution

长任务可能运行几分钟、几小时甚至等待人工或外部事件。进程内状态和内存队列在重启后丢失，Runtime 需要 durable event、checkpoint、timer、signal 和 activity result。Temporal 的工作流模型把持久状态、activity、重试和定时器组合起来 [7]；数据库事件表也可以实现类似语义，但需要自行处理 replay、版本和并发。

持久化不等于每个 token 都写数据库。模型流式输出可以写 artifact、摘要或可恢复 offset；关键状态迁移、工具请求、审批和结果必须持久化。状态快照定期压缩事件，但保留事件版本和 hash，便于回放。大上下文使用 message store 或对象存储，状态表只保存引用。

恢复语义必须先回答“恢复什么”，再选择存储技术。恢复一个纯读取模型 Step，可以重新发起调用或复用录制结果；恢复一个尚未提交的工具操作，可以用同一 operation identity 继续；恢复一个提交后结果未知的写操作，首先要查询外部系统事实；恢复一个等待人工的 Step，只能恢复到等待状态，不能由 worker 推断批准。把这些情形都叫作 retry，会把不同的业务后果压成同一个按钮。Runtime 应在 Step spec 中明确 `replayable`、`idempotent`、`query-before-retry`、`compensable` 与 `human-only` 等恢复类别，并把分类同工具版本一起保存。

checkpoint 也不应只是“每隔 N 分钟存一次”。一个可恢复 checkpoint 至少包含 Run 与 workflow revision、当前状态机节点、已完成和待完成的依赖、每个 attempt、预算消耗、deadline、上下文与 memory 的不可变引用、已发出的 operation、租约/fencing token、审批状态和计时器。它还要标记哪些字段是观察结果、哪些是模型候选、哪些已成为外部业务事实。没有这一界线，恢复 worker 可能把上次模型的自然语言结论当作提交结果，或在重放时漏掉已经开始的补偿。

checkpoint 的写入点由恢复点目标决定。对只读链路，写在一个完整 Step 后通常足够；对会触发外部副作用的链路，必须在提交意图、获得 operation id、收到外部确认、开始补偿和完成补偿这些边界分别记录。写得过密会拖慢执行并放大状态库成本，写得过疏会使重启后的不确定窗口过大。团队应为每种工作流测量恢复点目标和允许丢失工作量，而不是把统一间隔当作正确答案。长时间模型流可用 offset 和 artifact hash 降低存储压力，但最终业务状态仍只能由持久事件确认。

恢复时的顺序同样重要。控制面读取最近一致 checkpoint，验证 workflow、工具、权限和 schema 是否兼容，冻结旧 worker 的提交权，再根据未完成节点的恢复分类安排新 worker。若旧 worker 在网络分区后恢复，它携带的 fencing token 已过期，控制面拒绝其迟到写入。若外部工具不能查询 operation 状态，Runtime 不能假装实现 exactly-once；它应把这项能力缺失暴露为风险，并为该工具要求人工核对或业务侧幂等。持久化编排系统的价值并非让失败消失，而是把失败后的事实、所有权与下一步动作保留下来 [7]。

### 重试分类

重试必须根据错误类型决定。网络超时、限流、临时不可用通常可退避；schema 错误、权限拒绝、用户取消、预算超限和确定性业务错误不应自动重试；模型输出不确定时可以有限次重新采样，但要增加 attempt 和成本。Google API 错误与重试规范强调错误分类、backoff、deadline 和幂等的重要性 [21]。

重试预算按 Run、Step、工具、租户和全局设置。指数退避加随机抖动避免惊群；deadline 传递到每一层，剩余时间不足时直接失败或转人工。重试不能产生新的 root task，也不能重复计费或重复副作用。每次 attempt 保存输入 hash、模型版本、错误、等待、输出和资源。

幂等的 owner 必须明确到边界。Runtime 可以为一次调用生成 operation id，并确保同一 Run/Step/attempt 的消息不被重复投递；工具 adapter 负责把 id 传给下游；真正拥有业务状态的订单、邮件、文件或支付系统则必须决定相同 id 的重复提交如何返回同一事实。若下游只提供“尽力而为”的 HTTP 接口，Runtime 无法凭本地去重表承诺业务幂等，因为网络在提交后的断开会留下未知状态。Enterprise Integration Patterns 的 Idempotent Receiver 模式正是将去重与消息接收边界绑定 [19]；在 Agent 系统中，边界还要覆盖模型重试、队列至少一次投递和人工重放。

operation id 应包含稳定的业务范围，而不是随机 request id。发送同一封通知可以使用 `tenant + business-object + action + version`；修改文件可以使用文件版本和补丁 hash；创建工单可以使用外部系统允许的 client token。若模型每次重试都重新生成操作语义或参数，系统必须先比较 canonical payload：相同意图可查询或去重，不同意图则是新操作，可能需要新审批。把“看起来相似”的自然语言当作幂等键会造成误合并，反过来把相同操作的随机 UUID 当作新请求会造成重复副作用。

幂等记录需要生命周期治理。短期通知可能只需要在业务 SLA 窗口内保留键；财务、权限和数据删除等操作则需要更长的审计与重放窗口。记录至少保存 operation id、payload hash、请求者、状态、外部引用、首次与最后一次观察、响应摘要、过期策略和 owner。过早删除会让延迟消息重新造成副作用，永久保留又会造成隐私和存储问题。保留期应与下游的重复投递上界、合规需要和补偿窗口一起确定，并通过文档而不是隐式代码约定。

### 补偿与人工接管

不可逆副作用不能依靠回滚数据库解决。发送邮件后可以发送撤回或补充通知，创建订单后可以取消，修改文件前保存版本，支付动作通常要求人工或业务事务。工具注册表声明 compensation、是否可逆、是否需要 approval 和失败后 owner。Saga 将一组局部事务及其补偿组织成长事务 [18]。

补偿本身也可能失败，不能假设它一定成功。Runtime 将主动作、补偿动作和人工任务放进同一状态图，记录部分成功和当前风险。对用户返回明确的 completed、partially_completed、needs_attention 或 failed，不把补偿未完成隐藏成成功。人工处理完成后以事件关闭，而不是直接修改历史状态。

补偿不是把数据库回滚一遍，而是为已提交的业务事实追加一个反向或缓解动作。它必须有自己的前置条件、权限、deadline、幂等键、可观测性和失败路径。例如已发送的邮件只能追加更正或撤回请求，不能抹掉收件人已经看到的内容；已创建的外部工单可以关闭，但不能保证对方没有开始处理；已写入文件可以恢复到一个已知版本，但若之后存在人工编辑，自动覆盖反而可能扩大损失。Saga 提供的是局部事务的协调思路 [18]，不是“所有动作都可逆”的承诺。

因此补偿计划应在执行主动作之前确定。工具注册表声明正常 operation、成功证据、补偿 operation、不能补偿的后果、人工 owner 和最大等待时间；workflow 在跨越不可逆边界前检查这些字段和审批。对于没有安全补偿的高影响动作，设计应改为预览、双人审批、延迟提交或由业务系统提供事务性 API，而不是把风险留给模型临场判断。补偿开始后，用户可见状态应变为 `compensating` 或 `needs_attention`，避免前端把“原动作成功”误展示为最终成功。

补偿完成也要按业务事实验收。HTTP 200、队列接受或模型说“已恢复”都不足够；Runtime 需要保存查询到的版本、外部对象状态或人工确认。若主动作只完成了一部分，补偿可选择继续完成剩余部分、撤销已完成部分，或交给人工，这三种选择各自改变成本、可用性和用户预期。工作流定义应把选择写成版本化决策，不能让新 worker 在事故现场临时猜测。

### 版本化工作流

工作流定义改变时，正在运行的 Run 不能突然使用新节点、重试和补偿语义。每个 Run 绑定 workflow version；新版本只用于新 Run，或通过明确的 migration event 在安全边界切换。Temporal 的版本兼容思想可以帮助处理长时间运行的代码变化 [7]。状态图节点 id、输入输出 schema 和 side effect 要保持稳定或提供迁移函数。

## 11.5 上下文、记忆与状态存储

### Message store 与 Context builder

Runtime 不应把全部历史消息直接拼进每次 prompt。message store 保存原始消息、来源、时间、权限、hash 和 parent；context builder 根据当前 step 的预算、任务、工具、记忆和安全策略选择可见上下文。模型看到的最终 prompt 记录版本和组成摘要，便于回放与审计。

上下文构建需要优先级：系统指令、当前任务约束、工具 schema、必要观察、相关记忆、历史消息、低优先级背景。超出窗口时可以摘要、压缩、检索或失败，不应静默删除安全约束和用户目标。上下文预算同时影响 latency、token 成本、KV 和质量，Runtime 把它作为调度参数。

### 短期状态与长期记忆

短期状态包括当前 Run、step、观察和临时变量；长期记忆包括用户偏好、历史事实、工作成果和可检索文档。两者有不同可靠性和权限，不能因为模型生成了一句话就写入长期记忆。Memory 记录 source、confidence、created_at、updated_at、expiry、tenant、subject 和 consent；冲突时保留证据与版本。

Generative Agents 展示了记忆、反思和计划如何支持长期行为 [4]，但生产 Runtime 还要增加删除、纠错、租户隔离和过期。记忆召回进入上下文前执行权限和相关性检查，用户可以查看、修改或删除。模型对记忆的陈述不是事实来源，重要动作仍需外部系统确认。

### 状态一致与并发

同一 Run 可能由恢复 worker、用户取消、工具回调和定时器同时更新。状态存储使用 optimistic concurrency、版本号或 actor single writer，拒绝旧版本写入。事件处理器幂等，projection 可重建。多 Agent 共享任务状态时，明确哪些字段只能由协调者写，哪些是 append-only message。

分布式锁不能解决所有问题：锁超时、worker 崩溃和外部副作用仍会产生重复。状态机、租约、幂等键和查询接口组合起来，才能安全恢复。每次状态更新返回 version，调用方用 compare-and-set；冲突时重新读取并决定合并或失败。

## 11.6 调度、租户与成本

### 多租户资源

租户配额覆盖模型调用、输入输出 token、工具次数、并行 step、GPU、CPU、内存、存储、网络、长期记忆和人工审核。一个租户可能低 QPS 但运行长 Agent，另一个高 QPS 但短请求，单一 QPS 配额不公平。Runtime 同时维护 request budget、Run budget、project budget 和全局容量，超限返回可解释错误。

租户隔离包括状态、提示、记忆、工具、凭证、trace、cache 和 artifact。模型输出可能包含另一个租户的数据，任何跨租户共享 cache 或记忆都必须有明确的公共范围。Borg 的配额、优先级和隔离经验适合用来设计资源控制面 [10]；Kubernetes 提供执行层隔离基础 [9]。

### 调度优先级

交互式任务、批量任务、生产告警、人工审批和开发实验使用不同 deadline。调度器以可用资源、剩余预算、依赖就绪、租户份额和 deadline 选择 step。高优先级可以抢占等待队列，但不应随意打断不可逆工具；长任务被抢占后要保存 checkpoint 或返回可恢复状态。

Agent 的资源消耗不可只按一次模型调用估算。一个 Run 可能出现规划、检索、工具、反思、验证和重试。调度器维护动态 budget，step 完成后扣除实际 token、GPU 和外部调用，预测剩余成本与 deadline。达到阈值时可以摘要、换小模型、减少候选、暂停等待或请求人工。

### 成本与计费

计费事件按 request、step、tool operation、model token、GPU 时间、存储和人工记录。失败重试是否计费由产品定义，但平台必须保留真实成本。父任务与子任务共享 budget，工具副作用和人工审核单独计量。成本事件幂等并带版本，不能因为重试消费消息而重复计费。

成本归因还要连接成功任务。只按 token 计费会激励输出更长或反复重试；单位成功任务成本更接近业务价值。Runtime 把 task outcome、token、latency、工具失败、人工接管和模型版本放进同一 trace，支持不同模型和策略的比较。

## 11.7 可观测性、调试与回放

### Trace 结构

一个 Agent trace 以 task/run 为 root，包含模型 step、tool operation、retrieval、memory read/write、approval、retry、queue 和 compensation span。每个 span 携带 model revision、prompt hash、input/output token、tool name/version、tenant、budget、deadline、attempt、error 和 state version。Dapper 的分布式追踪方法和 OpenTelemetry 的上下文传播提供基础 [14][16]。

默认不记录完整 prompt、记忆和敏感工具返回；使用 hash、摘要和受控 event。需要调试时通过临时授权、采样和脱敏提升粒度。trace 与日志通过 trace id 关联，指标使用低基数标签。Prometheus 监控 Run active、step latency、retry、tool error、budget、queue 和 SLO [15]。

### 可解释的执行回放

回放读取事件和 artifact，按原始 workflow version 重建状态，不自动执行副作用。模型调用可以使用录制结果、sandbox 或新模型；工具使用录制 response、mock 或只读查询。回放标记哪些结果是真实、录制、模拟和重新生成，避免把新模型输出误当作历史事实。

回放支持三类目的：事故调查、版本比较和测试。事故调查重建失败路径；版本比较在同一输入和外部观察上运行新策略；测试验证状态机、重试、补偿和权限。回放数据带租户和数据等级，访问与导出审计。一个可回放的 Runtime 比仅能查看最后答案的系统更容易演进。

### 指标与 SLO

Runtime SLO 分为任务完成率、step 可用性、工具成功率、恢复时间、人工接管、总延迟、模型 TTFT/TPOT、成本和安全事件。Agent 任务的 HTTP 成功不代表完成；需要定义业务 outcome。SRE Workbook 的错误预算思想可以决定是否暂停实验、降低并发或切换稳定 workflow [17]。

## 11.8 安全、审批与供应链

### Prompt injection 与工具边界

工具返回、检索文档、网页和文件内容都可能包含让模型越权的指令。Runtime 将外部内容标记为 untrusted，系统指令、用户指令和工具结果分层；模型提出的动作必须经过权限和 schema，而不是让内容直接改变 policy。OWASP LLM Top 10 将提示注入、数据泄露、工具风险和供应链列为重要问题 [25]。

工具权限采用最小范围、短期凭证和目标 allowlist。用户可以允许“读取某项目文件”，但不等于允许写任意路径；允许查询数据库，不等于允许执行任意 SQL。权限决策记录主体、资源、动作、理由、策略版本和结果，供审计与回放。

### Human-in-the-loop

审批适用于高风险写入、外部消息、支付、权限变更、删除、公开发布和不确定安全动作。审批请求包含目标、参数、影响、证据、风险、过期时间和可选修改；批准与拒绝都是事件。审批超时进入明确状态，不自动当作批准。用户取消 Run 后，待审批动作必须失效。

人工可以修改模型生成的参数，但修改结果标记为 human-edited；工具执行后不可通过回写历史假装未发生。人工操作也有租户、角色、双人复核和审计。Runtime 把人工接管作为正常状态，不把它当作异常黑洞。

### Runtime 供应链

workflow 定义、tool server、MCP server、prompt template、memory adapter、模型和 sandbox 都是供应链制品。注册表保存来源、签名、版本、依赖、权限和审批；运行时只加载允许的 hash。第三方工具升级需要重新做 schema、权限、超时、数据流和安全测试。容器和依赖锁定，网络 egress 限制，运行时禁止工具下载未审计代码。

## 11.9 故障、回滚与弹性

### 故障分类

模型故障包括超时、限流、格式错误、上下文超限和质量门禁失败；工具故障包括网络、权限、schema、部分成功和外部状态未知；运行时故障包括 worker crash、状态库不可用、事件乱序、租约过期和队列堆积；安全故障包括凭证泄露、越权、注入和敏感输出。分类决定重试、补偿、回滚和人工。

### 恢复边界

恢复优先从最近 durable checkpoint 或 step 边界开始，而不是重新执行所有步骤。纯读模型调用可以重试，幂等工具可以重试，不可逆工具需要查询 operation 状态，未知状态动作要进入人工。工作流版本、上下文 snapshot、memory version 和工具协议必须与 Run 绑定，恢复不能使用最新默认值。

恢复边界应在 workflow 设计阶段画出来。一个边界前的内容可安全重新计算，一个边界后的内容必须从已确认事实继续；两者之间是“提交已发出但事实未知”的灰区。灰区不是异常分支的细节，而是外部系统交互的常态：支付网关、邮件服务、工单系统、数据库写入和设备控制都可能在客户端看到结果前已经完成。Runtime 在灰区内保存 operation id、请求摘要、下游引用和最后观察时间，优先查询；查询仍不能消除不确定性时，转入显式人工队列，并禁止自动复制请求。这样做可能降低自动完成率，却保留了正确性与可审计性。

不同恢复类别需要不同验收。模型调用的恢复检查输入与配置是否可复现、预算是否还允许；检索恢复检查数据版本与权限是否仍成立；工具恢复检查 operation 状态与幂等窗口；等待事件恢复检查 correlation id、签名、deadline 与取消状态；人工恢复检查决定是否仍有效、操作者是否有权限。把所有 Step 交给一个通用 retry middleware 会遗漏这些业务条件。通用框架应提供状态、租约、超时和事件基础设施，具体恢复分类则属于 workflow 与工具契约。

版本演进时，恢复边界更不能被悄悄移动。旧 Run 的 checkpoint 可能引用已经下线的模型、工具 schema 或 memory 格式；新版本要么提供兼容 adapter，要么把旧 Run 保留在旧执行环境，要么在明确停机点由人工批准迁移。迁移事件记录旧值、新值、转换函数、验证结果和回退路径。若没有证明旧事件在新定义下仍有相同语义，系统应宁可暂停该 Run，也不要让“升级”变成一次未经批准的业务重放。

### 版本发布与回滚

Runtime 发布包括 workflow、模型、prompt、tool schema、policy、memory、router 和 evaluator。灰度按租户、任务类型、区域或实验组；影子运行不执行副作用。指标包括任务成功、步骤成功、重试、工具错误、成本、延迟、安全和人工接管。回滚是状态与路由转换，旧版本必须保留兼容的 schema 与恢复逻辑。

发布与回滚的最小单位是经过签名的 Runtime manifest。它不只包含代码 commit，还包含 workflow 定义、活动 worker 镜像、模型和 prompt revision、工具与 MCP server 版本、JSON Schema、policy、memory adapter、路由与配额、评估集、feature flag 和数据保留规则。Run 在创建时绑定 manifest hash；trace、checkpoint 和成本事件携带同一 hash。这样，团队能解释某次失败到底来自模型、工具、策略还是编排，而不是在多处默认配置之间猜测。

灰度需要把副作用隔离作为硬前提。影子 Run 可以读取同一输入、执行录制工具或只读查询、比较计划和结构化结果，但不能发送消息、修改文件、创建订单或写长期记忆。对真实流量放量时，按风险而非只按百分比选择：先选择可撤销、低权限、可人工观察的任务，再扩大到高影响任务。每一个阶段设定质量、重复操作、未知状态、恢复时延、成本、隐私和人工接管门槛；超过门槛后自动停止扩容并保存可回放证据。灰度不是“新版本可能有问题”的缓冲，而是让风险范围与证据增长同步的控制机制。

回滚也有两类。路由回滚让新 Run 立即使用旧 manifest；状态回滚处理已经在新 manifest 下运行的 Run。前者较快，后者必须尊重已开始的工作流、操作和补偿，不能强制把 checkpoint 解释成旧 schema。发布计划因此要列出：哪些 Run 可以继续完成，哪些必须暂停，哪些可迁移，哪些需要人工；旧依赖会保留多久；回滚演练使用什么录制事件；以及紧急情况下谁拥有冻结工具、撤销凭证和恢复流量的权限。把这些问题留给事故期间解决，往往会把技术回滚扩大为业务中断。

## 11.10 Runtime 验收与交付

### 功能与状态验收

测试正常完成、模型错误、工具超时、权限拒绝、重复事件、迟到事件、worker 重启、状态库故障、人工审批、取消、过期和部分成功。验证每个状态迁移、attempt、幂等键、预算、trace 和资源释放。状态回放后得到的当前状态应与在线状态一致，重复 replay 不产生副作用。

验收用例需要把“未知状态”单独列出。最有价值的故障往往不是工具明确返回失败，而是请求在远端提交之后连接断开、worker 在写入本地事件之前崩溃、回调到达两次或乱序到达。测试应断言 Runtime 先查询 operation，再决定完成、重试、补偿或人工，而不是根据超时猜测失败。还要在 checkpoint 前后分别杀死 worker，检查恢复出的 workflow revision、输入 hash、attempt 和预算是否与事件历史一致。若无法构造这类演练，就不能宣称某个工具具有可靠恢复语义。

状态验收还应包含并发竞态：取消事件与工具成功同时到达，人工批准发生在 deadline 之后，旧 worker 在 lease 过期后提交结果，升级后的控制面读取旧版本 snapshot。每种竞态都要有唯一的可见终态和审计说明。建议把状态机转移表变成自动化属性测试：任意重复 event、乱序 event 或 replay 不得产生第二次业务提交；任意终态不得回到执行态；任意 `needs_attention` 都能找到 owner、证据和下一步期限。这些性质比只覆盖一条 happy path 更接近 Runtime 的正确性定义。

### 任务与工具验收

测试 schema 错误修复、工具重试、不可逆动作、补偿失败、MCP server 断线、server 能力变化、提示注入、恶意返回、文件路径、SQL、网络 egress 和敏感输出。每个工具有 owner、版本、权限、限流、超时、side effect 和 compensation 证据。工具成功不能只看 HTTP 200，还要验证业务状态与幂等。

工具验收应建立“契约样本”而非只依赖现场服务。每个版本保留正常返回、可重试错误、确定性错误、部分成功、畸形 schema、超大输出、延迟回调与权限拒绝的录制样本；adapter 对这些样本验证 schema、canonical payload、operation id 传播、脱敏和状态映射。MCP 或第三方 server 的能力变更要被视为兼容性事件：新增字段未必安全，删除字段、修改枚举或改变副作用语义必须阻断发布，直到 owner 重新签署契约。协议规范和 JSON Schema 是输入边界 [12][13]，但业务成功证据仍由工具 owner 定义。

对有副作用工具，验收报告要明确 idempotency 的实现位置。若由下游业务系统保证，应附上重复请求、未知状态查询和过期键的证据；若由 Runtime adapter 保证，应说明状态库故障与多区域重试下的边界；若只能人工处理，则不得在产品中标成“自动重试”。这种诚实的能力分级比把所有工具都贴上“reliable”标签更利于编排器作出保守选择。

### 资源与租户验收

测试多租户并发、优先级、公平、预算耗尽、长 Run、并行分支、GPU/CPU 配额、存储、记忆和 trace 隔离。检查一个租户无法读取另一个租户的 prompt、memory、cache、artifact 或工具结果；成本事件与 task outcome 能对账；超预算有明确降级和人工路径。

资源与成本验收应从“账能否对上”推进到“决策是否可复现”。对一组混合负载，报告每个 tenant、project、Run、Step 和 operation 的 token、GPU、存储、工具、重试与人工成本，并证明汇总值与原始不可变计量事件一致。随后改变一个策略参数，例如租户配额、模型路由或最大并行度，验证控制面能预测哪些任务会被延迟、降级或拒绝，以及结果是否符合政策。只在月末发现账目不平，说明成本系统仍是事后报表而不是运行时控制器。

隔离验收还应覆盖 side channel：不同租户的 cache 命中、错误文本长度、队列等待、artifact URL 和指标标签是否泄露存在性或内容；高负载租户是否能通过重试或长上下文把其他租户挤出 SLO；被删除的 memory、凭证与 artifact 是否仍出现在 replay 或备份索引。安全测试不应为了方便而使用同一租户下的不同项目替代真正的跨租户边界，因为两者的授权路径可能不同。

### 可观测、恢复和发布验收

使用一个完整 Run 验证 trace 从网关到模型、工具、审批、存储和计费；杀死 worker 后恢复；切换 workflow 和模型版本；回滚；执行影子和灰度；回放失败任务。报告包含成功、成本、延迟、重试、工具副作用、安全、SLO 和数据保留。只有这些证据齐全，Runtime 才可从实验池进入生产。

验收指标应有预先约定的门槛和重评估条件。核心集合可包括：任务成功率及其分层、已知副作用的重复率、未知 operation 的人工收敛时间、从 worker 故障到可恢复状态的时间、checkpoint 恢复正确率、补偿完成率、p95/p99 端到端时延、单位成功任务成本、预算超限率、跨租户隔离失败数、关键 trace 完整率，以及发布后回滚时间。阈值由任务风险和业务目标确定，不能把某个团队的历史数字写成通用标准；但每个阈值都要有数据来源、owner、观察窗口和超过时的处置。

可以把生产准入分为四个 gate：语义 gate 验证状态机、幂等与补偿；安全 gate 验证权限、隔离、秘密与人工边界；容量 gate 验证混合负载、预算、队列和故障余量；发布 gate 验证影子、灰度、可回放和回滚。一个 gate 未通过时，系统仍可在受限实验环境运行，但不得扩大可执行权限或影响范围。这样，团队能把“模型 demo 可以运行”和“具备生产行动资格”清楚地区分。

运行时 Infra 的核心结论是：模型负责生成候选，Runtime 负责决定候选能否成为动作；工具负责执行局部能力，Runtime 负责状态、权限、重试和补偿；工作流负责描述流程，Runtime 负责持久化、调度、观测和回滚。Agent 的智能表现依赖模型，但 Agent 的可靠行为依赖这些系统边界。

### 验收结论

运行时 Infra 的核心结论是：模型负责生成候选，Runtime 负责决定候选能否成为动作；工具负责执行局部能力，Runtime 负责状态、权限、重试和补偿；工作流负责描述流程，Runtime 负责持久化、调度、观测和回滚。Agent 的智能表现依赖模型，但 Agent 的可靠行为依赖这些系统边界。

接口、状态、工具生命周期、模型路由、上下文预算、记忆写入和多 Agent 协作都应回到同一个最小契约：每个动作有版本、权限、预算、幂等、证据和回滚语义。11.15 会把这份最小契约单独收束，后续第13到第21章再展开具体 Prompt、工具、记忆和编排方法。

## 11.11 运行时安全与隔离细节

### 不可信观察

检索文档、网页、邮件、代码、工具返回和用户上传文件都可能包含指令。Runtime 保存来源与信任等级，把 observation 与 system policy 分开；模型可以读取 observation，但不能通过 observation 修改权限、工具 allowlist、预算或审批规则。上下文构建器对外部内容做标记和长度限制，敏感字段脱敏。

攻击者可能利用工具错误、跨 Run 记忆、cache 命中、错误回放或日志导出获取数据。权限检查在每次读取和写入时执行，不能只在任务开始时授权。工具输出如果进入另一个租户的任务，需要明确数据共享策略；默认不共享。OWASP LLM 风险清单可用于建立测试与威胁模型 [25]。

### 凭证与秘密

Runtime 不把 API key、数据库密码或云凭证放进 prompt、事件或模型输出。工具 worker 通过短期 token、工作负载身份或秘密管理服务获得最小权限，token 绑定 tenant、run、tool、resource 和 expiration。重试使用同一 operation identity 或查询原状态，不能每次发放新宽权限凭证。

工具返回中的秘密扫描、日志脱敏和输出过滤需要独立层。即使模型无法理解秘密，后续 trace、缓存和人工标注也可能泄露。发生凭证疑似泄露时，Runtime 触发 revoke、隔离 Run、停止相关工具、保存审计和通知；不能只删除一条日志。

### 网络与文件隔离

代码、浏览器和数据工具使用 egress allowlist、DNS 限制、代理审计、文件系统 namespace、CPU/内存/PID 限制和执行时间。模型生成的路径、URL、SQL 和命令先由策略解析，再交给工具。容器 root、宿主机 socket、任意 cloud metadata 和跨租户网络默认禁止。Kubernetes 的 namespace、service account 和 network policy 是基础，但高风险工具可能需要 microVM。

沙箱输出限长、限类型、限存储，并标记 artifact owner。大输出存对象仓库并过期，小输出进入 event；二进制和脚本不直接嵌入 prompt。下载文件再被模型读取时，做 MIME、大小、编码、恶意内容和权限检查。运行时只提供必要的文件路径引用，避免把宿主机布局暴露给模型。

## 11.12 工作流编排的工程模式

### 顺序、分支与循环

顺序工作流适合确定的步骤，分支根据条件选择路径，并行减少等待，循环用于计划—观察—修正。每种结构在持久化图中有明确节点和边，循环有最大次数、进度判定和重复检测。不能让模型通过输出任意节点名称跳转到未经授权的状态；跳转由图和策略校验。

条件判断可以来自规则、工具事实、模型分类或人工。模型判断需要置信度和 fallback，规则判断需要版本，人工判断需要过期和审计。join 节点声明必需、可选、超时和部分结果；否则一个低价值分支失败会阻塞整个任务，或一个关键分支缺失却继续完成。

### 事件驱动与定时器

Agent 常需要等待 webhook、审批、定时、文件上传或外部 job。事件驱动系统用 correlation id、subscription、deadline 和签名匹配回调。重复 webhook 通过 idempotency key 去重，伪造回调通过认证与资源校验拒绝。定时器在控制面持久化，worker 重启不应丢失。

事件到达时 Run 可能已取消、超时或版本迁移。处理器根据状态和 event version 决定忽略、记录或触发补偿，不能盲目继续工具。外部事件的 payload 作为不可信 observation，进入模型前经过 schema 和权限检查。消息 broker 至少一次投递时，状态机和幂等是正确性基础。

### 长任务与租约

worker 领取 Step 时获得带期限 lease，持续心跳或续租；失联后控制面在 lease 过期后重新调度。工具本身也可能继续执行，重新调度前查询 operation 状态。lease 不能替代幂等，网络分区下旧 worker 可能恢复并提交迟到结果，控制面使用 fencing token 拒绝旧 owner。

长任务的心跳包含 progress、resource、attempt、checkpoint 和预计完成；心跳不上传敏感内容。控制面根据 heartbeat 判断卡死、慢任务和预算，必要时暂停或迁移。迁移策略写入 workflow：模型 step 可以重算，副作用 step 先查询，人工 step 等待或转移队列。

### 状态压缩与归档

长 Run 事件很多，状态存储需要 snapshot、事件归档、artifact 引用和 TTL。snapshot 包含状态版本、当前节点、预算、上下文引用、memory 引用、未完成 operation 和 workflow version；事件归档保留审计与 replay 所需的最小字段。用户删除或租户归档时，按照保留策略删除内容并保留合规的摘要。

归档后的 Run 仍可查询摘要、最终结果、成本和失败原因；恢复执行需要明确重新激活和权限。压缩不能丢掉不可逆动作、审批、补偿和安全事件。状态表只保存引用与索引，原始大文本和文件 artifact 由生命周期策略管理。

## 11.13 运行时观测与成本闭环

### 任务级 SLO

Runtime SLO 包括任务完成、步骤完成、首个有意义结果、总时延、工具成功、人工接管、预算超限和安全事件。一个任务可能在单次模型调用都成功的情况下失败，例如工具参数错误、目标状态未改变或最终答案没有证据。业务 outcome 由 workflow 定义，平台记录 outcome type 与证据。

不同任务使用不同 SLO。告警处理 Agent 关心在 deadline 前完成和误操作；知识助手关心引用和答案相关；Coding Agent 关心测试通过和改动安全；批量抽取关心结构化合法和覆盖。统一平台提供指标采集，业务定义成功；不能用平均模型 latency 替代任务质量。

### Trace 与隐私

trace span 连接 task、run、step、model、tool、memory、approval、retry、compensation、queue 和 cost。字段包括 hash、版本、token、耗时、错误和状态，不默认包括完整上下文。敏感 trace 按 tenant、data class、retention 和访问角色隔离。OpenTelemetry 提供传播规范，Dapper 提供大规模 tracing 的参考 [14][16]。

trace sampling 要保留所有失败、超预算、安全和人工接管事件；正常成功可以按比例采样。高价值任务全量保留摘要。采样策略版本化，否则线上质量趋势会因观测比例改变而产生假象。日志和指标通过 trace id 关联，成本与计费使用独立的不可丢失事件。

### 运行时成本

成本由模型 token、GPU、工具调用、网络、存储、记忆、评估、人工和重试组成。一个 Run 的成本树把 parent、child、step 和 operation 聚合，计费事件带 idempotency。重试的模型调用可计费，工具重复提交不能重复产生业务费用；人工接管单独计入任务成本。

成本预算和质量门禁一起工作。若任务已经满足目标，继续反思和多候选只增加成本；若失败原因是外部工具，切换更大模型未必有价值；若输出不确定，增加验证可能降低人工成本。Runtime 暴露剩余预算给策略，但不让模型任意读取其他租户的预算或修改上限。

## 11.14 Runtime 验收案例

### 纯读取任务

读取知识库并生成答案的任务验证检索、上下文、引用、模型、输出和 trace。模拟知识库超时、空结果、权限拒绝、重复文档、过长上下文和模型超时，确认任务有明确状态和可解释降级。回放时不改变线上数据，反馈样本可以进入评估候选。

### 有副作用任务

发送消息、修改文件、创建工单或执行数据库写入的任务验证审批、schema、幂等、查询、补偿、部分成功和人工接管。网络在提交后断开时，重试首先查询 operation；工具返回未知状态时暂停而非再次执行。用户取消后已批准但未提交的动作失效，已提交动作进入补偿或人工。

### 长时 Agent

模拟运行数小时、等待 webhook、worker 重启、状态库切换、workflow 升级、工具版本退役和预算耗尽。确认 Run 可以从 durable state 恢复，旧版本继续执行兼容节点，事件不会重复副作用，最终成本和 trace 完整。归档后查询摘要，重新激活需要权限和新预算。

### 多租户压力

多个租户同时运行短任务、长任务、批任务和高风险工具，验证优先级、公平、配额、内存、连接、trace、记忆、artifact 和费用隔离。一个租户的工具爆发不能阻塞其他租户；一个模型故障不能让所有任务进入无限重试；预算和错误预算超限有明确降级。压测包含冷启动、节点故障、网络延迟和区域切换。

### 交付判断

Runtime 的交付证据包括对象与状态 schema、workflow version、工具注册与权限、事件日志、回放结果、故障演练、租户隔离、成本对账、SLO、灰度与回滚。每个关键操作都有 owner、版本、幂等、超时、重试和补偿；每个外部输入都标记来源和信任；每个失败都能定位到状态、模型、工具、资源或策略。

中文教材中关于序列建模、优化、泛化和不确定性的基础知识，提醒我们不要把模型输出当作确定事实 [22][23][24]。Runtime 的职责正是把概率输出包在确定的协议、状态和权限中：允许模型提出候选，但只有系统验证、预算和业务规则都通过时才执行。

## 11.15 最小 Runtime 接口与状态模型

本章到这里应收束为运行时边界，而不是展开工具、记忆、Prompt、多 Agent 和编排策略教程。那些内容会在第13到第21章分层讨论；本章只定义生产级 Agent Runtime 必须具备的最小接口、状态语义和交付证据。

### 最小外部接口

Runtime 的外部 API 至少分成八类：task submit、run query、event stream、approval、cancel、resume、artifact 和 replay。提交接口接受目标、上下文、能力范围、预算、deadline、优先级和幂等键；查询接口返回当前状态、进度、成本摘要、需要的动作和可见结果；事件流提供状态变化但不暴露未授权内容；审批接口把人工决定写成事件；取消和恢复接口必须穿透模型、工具和等待队列；artifact 接口保存大文本、文件和证据；replay 接口用于调试和审计，不能重新触发真实副作用。

输出也要区分 final answer、partial answer、tool result、human decision、failure 和 compensation。客户端不能通过判断文本是否为空来猜测状态。结构化错误包含 code、retryable、retry_after、run_id、step_id 和 remediation；敏感内部错误只在受控 trace 中出现。Google AIP-194 对错误分类和重试语义的强调，适合迁移到 Agent API [21]。

### 状态模型

| 状态对象 | 保存什么 | 关键约束 | 常见事故 |
|:---|:---|:---|:---|
| Task | 用户目标、租户、可见结果 | 可多次运行，不直接承载执行细节 | 重试覆盖历史结果 |
| Run | 一次执行的模型、工具、预算、状态 | 配置不可变，事件追加 | worker 重启后无法恢复 |
| Step | 模型、工具、检索、审批等一次动作 | attempt、超时、幂等、补偿 | 重复执行副作用 |
| Event | 状态事实和因果链 | append-only、版本化、可回放 | 迟到事件覆盖完成状态 |
| Artifact | 大文本、文件、工具结果、证据 | hash、权限、保留期 | 把敏感 payload 写入日志 |
| Operation | 外部副作用动作 | 独立 idempotency key、状态查询 | 网络断开后重复提交 |

模型响应是候选结果，业务状态是经过规则、工具和事务确认后的事实。Runtime 将 assistant message、tool plan、tool result 和 business commit 分开存储。模型说“订单已取消”不代表订单系统已经成功更新；只有工具返回带有订单版本和确认状态，业务状态才可以标记为已取消。这个边界能避免语言看起来成功而外部事实失败。

### 最小实现路径

一个团队不必第一天构建复杂多 Agent 平台。最小闭环可以是：不可变 Run spec、持久化 Step、统一 model/tool adapter、JSON Schema 校验、幂等 operation、基本 retry/deadline、trace、租户预算和人工暂停。先让一个有限 workflow 能在 worker 重启、工具超时和重复回调后恢复，再增加并行、记忆、MCP、复杂计划和多 Agent。

每次扩展保留简单路径作为基线。引入模型自主规划前，比较固定 workflow 的成功、成本和风险；引入记忆前，验证删除、冲突和权限；引入自动补偿前，验证不可逆动作和人工；引入多 Agent 前，验证共享状态和循环。复杂度只有在证据显示它改善任务结果时才值得引入。

### 与后续 Agent 章节的边界

| 后续主题 | 本章只保留的边界 | 后续章节展开 |
|:---|:---|:---|
| Prompt 与结构化输出 | 模型输出必须可解析、可校验 | 第14章讲提示协议和输出设计 |
| Context 与知识 | Context builder 是受控组件 | 第15、19章讲上下文和知识系统 |
| Harness 与工具 | 工具调用必须有 schema、权限、幂等 | 第16、18章讲工具生态和 MCP |
| Memory | 记忆写入需要来源、权限和确认 | 第20章讲记忆系统 |
| Workflow 与多 Agent | Runtime 保存状态和调度边界 | 第21章讲编排模式与框架生态 |
| Evals 与 Guardrails | 运行证据进入评估与治理 | 第22章讲 Agent 质量闭环 |

### 交付证据

交付包包括 Run/Step/Event schema、workflow version、工具注册、权限策略、幂等与补偿说明、状态回放、容量压测、故障演练、租户隔离、SLO、成本对账、评估门禁、灰度和回滚。随机抽一个生产 Run，应能从最终结果反查模型、模板、工具、记忆、审批、预算和所有关键事件；随机重放不能产生真实副作用。

Runtime 的成熟度不体现在“Agent 能完成一次 demo”，而体现在失败、重试、取消、升级、降级和跨租户压力下，历史事实不被覆盖，副作用不被重复，新的执行仍能被评估和审计。这份端到端证据也是第12章生产治理的输入。

最小接口还需要给每种状态一个对外语义。客户端提交 task 后得到的是不可变的 Run identity，而不是“模型已经开始工作”的承诺；查询接口返回当前状态、可见进度和下一步需要的人或系统动作；取消接口返回取消已被接受、仍在收敛，还是因不可逆 operation 转入人工；恢复接口只接受满足权限、预算和 workflow 兼容条件的请求。API 返回的 `completed` 必须表示 workflow 定义的业务证据已经满足，`failed` 表示无法自动收敛，`needs_attention` 表示仍存在明确 owner 与待处理风险。以状态码和结构化证据表达这些差异，才能避免调用方从自然语言答案反推执行结果。

事件模型也要规定因果而非只记录时间。每个事件带 event id、Run/Step id、parent 或 correlation id、workflow revision、状态 version、actor、时间、输入/输出摘要、幂等键和可信来源；控制面按版本拒绝过期写入，projection 可从 append-only 事件重建。对于来自外部系统的事件，签名、资源归属、deadline 和 operation id 是接受条件。事件可以迟到，但不能覆盖已经确认的业务事实；可以重复，但不能产生第二次副作用；可以被归档，但不能失去审计链。这些约束让回放成为判断系统真实行为的工具，而不只是调试界面。

一份可操作的 Runtime ADR 可以这样记录：背景是团队希望让模型自动重试工具以提高完成率；候选方案是无条件重试、仅按 HTTP 状态重试、或按 operation 分类并查询外部状态；最终选择第三种。获得的是副作用边界清楚、未知状态可人工收敛、回放可审计；主动牺牲的是实现复杂度、某些任务的自动完成率和额外查询成本；接受的风险是下游没有查询 API 的工具需要被限制。后续验证指标包括重复业务操作数、未知状态的收敛时间、人工接管率、任务成功率和单位成功任务成本；当这些指标表明查询成本高于风险收益时，再对低影响且真正幂等的工具放宽策略。这样的 ADR 把“可靠”拆成可证伪的取舍。

交付时，团队应能随机抽取一个生产 Run，从最终结果沿着 manifest、event、checkpoint、trace、计量和审批追溯到所有关键输入；也应能随机抽取一次故障演练，从故障注入追溯到检测、自动处置、人工 owner、恢复和剩余风险。两种抽查都通过，才说明 Runtime 的状态、资源和治理并非分散文档，而是在真实执行中连接成闭环。若只能展示成功 demo 或漂亮 trace，而无法解释未知状态、补偿失败与跨版本恢复，系统仍停留在实验框架阶段。

验收指标还要区分领先与滞后信号。重复 operation、数据泄露和业务损失是滞后信号，出现时已经造成影响；checkpoint 写入失败、无 owner 的工具、缺失 operation id、trace 断链、兼容性检查跳过、超出预算的 retry 和过期 lease 则是可提前阻断的领先信号。平台应为领先信号设置自动 gate 或告警，为滞后信号设置事故响应、补偿和复盘。两者都只看总数也不够：要按 workflow、工具版本、租户风险等级和发布 manifest 分层，才能在扩大影响前发现局部退化。

生产后评价也不能只以一次上线为终点。定期抽取完成、失败、取消、人工接管和补偿中的 Run 回放，检查状态 projection 是否仍能重建、外部引用是否可解释、保留期是否符合策略、旧 manifest 是否仍可处理未完成任务。新的工具、模型或权限策略加入时，重跑代表性历史事件，验证它们不会改变过去事实或绕过当时的审批。这样的持续验收将 Runtime 从“发布时通过一套测试”变成“运行中持续保持可恢复与可审计”的系统。

最后，风险接受必须有明确 owner。技术团队可以说明一个工具没有查询 API、一个补偿动作只具有尽力而为语义、一个旧版本只能暂停不能迁移；业务或治理 owner 决定在何种任务范围内接受这些限制，并记录到 manifest 或 ADR。没有 owner 的风险会在故障时落到一线操作员和最终用户身上。Runtime 的价值不是声称消除了不确定性，而是让不确定性被分类、测量、隔离，并在需要时进入有权限、有期限和有证据的人工决策。

一个完整的验收包应把这些判断固化为可执行材料：状态转移与 operation 分类表、checkpoint schema、工具契约与 owner 清单、幂等和补偿证据、回放样本、混沌演练记录、租户隔离报告、成本账本对账、SLO 与 error budget、发布 manifest、灰度结果和回滚演练。文档不是替代测试，而是把测试的范围、预期与责任保留下来；测试不是替代运行证据，而是证明这些承诺在可重复场景中成立。任何一项缺失都应明确标成限制，连同临时人工流程和到期复审日期。

在此基础上，团队才能逐步增加多 Agent、复杂规划、长期记忆和更强的自主性。每新增一项能力，都需要重新检查它是否引入新的状态写入者、外部副作用、资源账本项目、权限边界、恢复灰区或观测盲点。若答案是肯定的，就先扩展 Runtime 契约与验收，而不是仅让模型多输出几个步骤。可靠的自主性来自边界随能力一起生长，而不是希望更强的模型自动填补系统缺口。

这种增量方式也保护了团队的学习速度：每次新增能力只扩大一组可测假设，失败可被限定、回放和复盘；当证据不足时，系统仍能退回已验证的 workflow 与人工路径。生产级 Runtime 的长期目标不是消灭人工，而是在自动化尚不确定时保留可恢复的控制权。

因此，任何“更自主”的发布都应同时交付更强的证据：新增动作的 owner、幂等边界、恢复分类、补偿计划、资源预算、隔离规则、验收样本和回滚路径。若这些证据尚未准备好，正确的下一步是限制权限和影响范围，而不是把不确定性转嫁给生产用户。

这项原则同样适用于短期试验：即使只向内部用户开放，Run 仍应有可查询状态、取消与暂停语义、最小 trace 和预算上限。实验可以缩小能力与数据范围，但不应取消事实记录与人工接管；否则试验产生的失败样本无法成为下一次可靠扩展的依据。

把最小边界保留在实验阶段，能使后续扩展成为证据的累积，而不是一次重新设计。系统成熟的标志正是每次扩大范围时都仍能解释历史、控制副作用并恢复未完成工作。

当试验暴露了无法自动判定的事实，系统应优先停止该类动作、保留上下文与证据并交给明确的 owner；在修复契约、工具或流程之前，不把同一不确定性通过更多重试放大。这种保守收敛使自主能力的扩展始终以可恢复性为前提。

## 参考资料

[1] Yao, S., et al. *ReAct: Synergizing Reasoning and Acting in Language Models*. ICLR, 2023. https://arxiv.org/abs/2210.03629 访问日期：2026-09-22

[2] Schick, T., et al. *Toolformer*. NeurIPS, 2023. https://arxiv.org/abs/2302.04761 访问日期：2026-09-22

[3] Karpas, E., et al. *MRKL Systems*. 2022. https://arxiv.org/abs/2205.00445 访问日期：2026-09-22

[4] Park, J. S., et al. *Generative Agents*. UIST, 2023. https://arxiv.org/abs/2304.03442 访问日期：2026-09-22

[5] Wu, Q., et al. *AutoGen*. 2023. https://arxiv.org/abs/2308.08155 访问日期：2026-09-22

[6] LangChain. *LangGraph Documentation*. https://langchain-ai.github.io/langgraph/ 访问日期：2026-09-22

[7] Temporal. *Documentation*. https://docs.temporal.io/ 访问日期：2026-09-22

[8] Moritz, P., et al. *Ray*. OSDI, 2018. https://www.usenix.org/conference/osdi18/presentation/moritz 访问日期：2026-09-22

[9] Kubernetes. *Documentation*. https://kubernetes.io/docs/concepts/ 访问日期：2026-09-22

[10] Verma, A., et al. *Large-scale Cluster Management at Google with Borg*. EuroSys, 2015. https://research.google/pubs/large-scale-cluster-management-at-google-with-borg/ 访问日期：2026-09-22

[11] OpenAI. *Function Calling Guide*. https://platform.openai.com/docs/guides/function-calling 访问日期：2026-09-22

[12] Model Context Protocol. *Specification*. https://modelcontextprotocol.io/specification 访问日期：2026-09-22

[13] JSON Schema. *Specification*. https://json-schema.org/specification 访问日期：2026-09-22

[14] OpenTelemetry Authors. *Documentation*. https://opentelemetry.io/docs/ 访问日期：2026-09-22

[15] Prometheus Authors. *Documentation*. https://prometheus.io/docs/introduction/overview/ 访问日期：2026-09-22

[16] Sigelman, B. H., et al. *Dapper*. 2010. https://research.google/pubs/dapper-a-large-scale-distributed-systems-tracing-infrastructure/ 访问日期：2026-09-22

[17] Beyer, B., et al. *The Site Reliability Workbook*. 2018. https://sre.google/workbook/table-of-contents/ 访问日期：2026-09-22

[18] Garcia-Molina, H., & Salem, K. *Sagas*. ACM SIGMOD, 1987. https://www.cs.cornell.edu/andru/cs711/2002fa/reading/sagas.pdf 访问日期：2026-09-22

[19] Hohpe, G., & Woolf, B. *Idempotent Receiver*. Enterprise Integration Patterns. https://www.enterpriseintegrationpatterns.com/patterns/messaging/IdempotentReceiver.html 访问日期：2026-09-22

[20] AWS. *AWS Step Functions Documentation*. https://docs.aws.amazon.com/step-functions/ 访问日期：2026-09-22

[21] Google. *AIP-194: Errors*. https://google.aip.dev/194 访问日期：2026-09-22

[22] 周志华：《机器学习》。https://cs.nju.edu.cn/zhouzh/zhouzh.files/publication/MLbook2016.htm 访问日期：2026-09-22

[23] 邱锡鹏：《神经网络与深度学习》。https://nndl.github.io/ 访问日期：2026-09-22

[24] 张量网络与深度学习：《动手学深度学习》。https://zh.d2l.ai/ 访问日期：2026-09-22

[25] OWASP. *Top 10 for Large Language Model Applications*. https://owasp.org/www-project-top-10-for-large-language-model-applications/ 访问日期：2026-09-22
