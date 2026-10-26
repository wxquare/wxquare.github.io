# 第 21 章 OpenClaw 系统：个人 AI 助手的 Gateway 实现、使用与调优

> OpenClaw 的核心不是“又一个聊天机器人”，而是把个人 AI 助手抽象成一个长期运行的本地 Gateway：它接入多渠道消息，管理会话与上下文，调度内置的 Agent Runtime，并用工具策略、沙箱、插件和审批控制行动边界。也正因为它是长期在线、可以替人行动的系统，安全边界比功能更重要。

**资料核验日期：2026-10-23。** 本章分析 OpenClaw 仓库 `openclaw/openclaw` 的固定提交 `ff7597025dd22b087b250a4df94caded50e798aa`（提交时间 2026-10-23 01:51 UTC，根 `package.json` 版本 `2026.10.6`）。官方文档页与源码都固定到这一提交，不引用主分支。OpenClaw 是开源项目，但本章的行为结论来自阅读源码与文档，没有运行验证、没有故障注入；正文里“行为是 A”的准确含义是“固定提交的源码（或文档）写了 A”。

| 标记 | 含义 | 使用方式 |
| --- | --- | --- |
| 源码事实 | 固定提交中可定位的文件与符号、行号 | 只引用 `ff7597025dd2…`，不引用主分支 |
| 官方文档事实 | 同一提交里 docs/ 的陈述 | 注明核验日期与版本 |
| 系统推导 | 根据公开证据做的架构解释 | 不等于产品内部实现 |
| 本书建议 | 本书依据事实归纳的使用建议 | 与官方建议分开标注 |

资料边界：源码描述固定提交的实现，文档描述该版本的承诺；两者冲突时并列说明条件。本章没有运行验证，不能把静态阅读推导为故障注入结论。

| 本章问题 | OpenClaw 的特有入口 | 通用机制 |
| --- | --- | --- |
| 谁能代表会话写入 | 会话控制器与转录 writer 声明 | [Runtime 状态与恢复](../part3/04-agent-runtime.md#failure-side-effects) |
| 重启后由谁继续 | 持久输入和恢复责任 | [任务受理](../part3/04-agent-runtime.md) |
| 已生成回复怎样发送 | reply operation 与出站账本 | [交付机制](../part3/04-agent-runtime.md#result-delivery) |
| 共享部署能否放权 | 渠道身份、工具策略和 OS 边界 | [工具沙箱](../part3/05-skills-tools-mcp-connectors.md#sandbox) |


## 21.1 组成：OpenClaw 由哪些部分构成

### 21.1.1 定位：一个长期运行的 Gateway

OpenClaw 的形态是一个单一、长期运行的 Gateway 进程，官方文档的表述是“A single long-lived Gateway owns all messaging surfaces”（一个长驻 Gateway 拥有全部消息面）[[1]](#ref-1)。它默认接入 WhatsApp、Telegram、Slack、Discord、Signal、iMessage 和 WebChat 等渠道，README 称“20+ more”；渠道适配器在仓库的 `extensions/` 目录里，实际数量远多于文档摘要列出的几个[[38]](#ref-38)。

Gateway 同时是控制面。控制面客户端（Control UI、CLI、节点）经 WebSocket 连接，默认监听 `127.0.0.1:18789`；每台主机一个 Gateway，由状态目录锁保证唯一（见 21.2.1）。同一端口还复用 HTTP API（`/v1/models`、`/v1/chat/completions`、`/v1/responses` 等 OpenAI 兼容面）、插件 HTTP 路由、Control UI 静态资源和 hooks[[1]](#ref-1)[[23]](#ref-23)。这个“一个端口承担消息渠道、控制面、HTTP API”的设计，是理解它后续安全模型（21.5.2 的暴露面阶梯）的前提。

定位上要分清两件事：OpenClaw 是个人助手的运行时，不是敌对多租户平台。官方安全页明确写，它“不是一个针对相互敌对的用户的安全多租户边界”，混合信任或对抗用户必须拆分信任边界（独立 Gateway 与凭据，最好独立 OS 用户或主机）[[25]](#ref-25)。这一条不是免责声明，而是 21.5 整个实践部分的框架。

### 21.1.2 仓库、包与进程形态

仓库是 pnpm workspace（根 README 注明“不支持在仓库根直接 `npm install`”），根包 `openclaw` 版本 `2026.10.6`，MIT，`engines.node` 为 `>=24.22.1 <25 || >=26.12.1`（README 写的 Node 版本范围与 package.json 不一致，以 package.json 为准）[[38]](#ref-38)。顶层分几块：`src/`（核心：gateway、agents、auto-reply、channels、routing、sessions、context-engine、cron、plugins、skills、security、llm、tui、node-host）、`extensions/`（捆绑的渠道与 provider 插件）、`packages/`（内部包：agent-core、ai、gateway-protocol、gateway-client、plugin-sdk、worker-runtime、retry 等）、`apps/`（各平台应用）、`ui/`（Control UI 前端）、`docs/`[[3]](#ref-3)。

进程形态有几种：前台 `openclaw gateway`，或经 `openclaw gateway install|start|stop` 管理的原生服务（launchd/systemd/Windows 计划任务）；还有独立的节点宿主（`openclaw node`）与 macOS/iOS/Android/Linux 应用。CPU 密集或隔离任务走 worker（`@openclaw/worker-runtime`），SQLite 通过 reader/writer worker 访问（reader 池上限 2 个），Gateway 主线程不使用 Node 磁盘模块编译缓存以避免同步 flush 阻塞事件循环[[3]](#ref-3)[[23]](#ref-23)。

### 21.1.3 状态与存储

状态目录默认 `~/.openclaw`（`OPENCLAW_STATE_DIR` 覆盖）。关键布局：

| 位置 | 内容 |
| --- | --- |
| `openclaw.json` | 主配置（JSON5），`OPENCLAW_CONFIG_PATH` 覆盖 |
| `state/openclaw.sqlite` | 共享状态库（workspace 设置状态、attestation、cron、定时、子 agent 注册等） |
| `agents/<agentId>/agent/openclaw-agent.sqlite` | 每 agent 的会话行、活动转录、认证 profile、运行时状态 |
| `credentials/` | 渠道与 provider 凭据 |
| `agents/<agentId>/sessions/` | 遗留迁移源/归档（JSONL 转录仅作迁移输入/归档/导入导出） |
| `sandboxes/`、`logs/`、`workspace/` | 沙箱工作区、日志、默认 workspace |

会话的权威状态在 SQLite，不在 JSONL：官方文档明确“活动 agent 历史与 session 行一起存在 SQLite”，session ID 稳定且由 OpenClaw 选择；JSONL 转录只作遗留迁移、归档与导入导出[[2]](#ref-2)。这修正了“OpenClaw 用 JSONL 存会话”的常见误解。数据库引擎除默认 SQLite 外，还声明支持 PostgreSQL（state/agent/workboard/logbook/team-reports 五个家族，schema 版本 state=31、agent=31），PostgreSQL 备份要求先停 Gateway[[38]](#ref-38)。

一个与恢复直接相关的表，是“哪些对象持久化在哪里”（21.4.6 会用到）：对话历史=每 agent SQLite；已接受的聊天跟进=每 agent SQLite 待处理输入；中断的 main 会话 turn=每 agent SQLite 会话行与转录；子 agent 运行=共享状态库；出站投递队列=SQLite；定时作业=SQLite cron store；重启 continuation=SQLite restart sentinel；后台 exec 进程与 Gateway 终端 PTY=仅进程内存，不恢复[[22]](#ref-22)。

### 21.1.4 Agent Runtime 与 Pi 的关系

OpenClaw 拥有内置的 agent runtime，文档明确写“没有外部 agent framework 包保留”；运行时代码在 `src/agents/`、`src/llm/`，可复用核心在私有包 `packages/agent-core`（`@openclaw/agent-core`，`agent-loop.ts` 导出 `runAgentLoop`）[[3]](#ref-3)。内置 runtime id 是 `openclaw`，遗留别名 `pi` 规范化为 `openclaw`，`codex-app-server` 规范化为 `codex`[[3]](#ref-3)。

与 Pi 的关系必须写成有版本绑定的事实，因为它随时间变化过。固定提交的根 `package.json` 里 Pi 系依赖只有一个：`@earendil-works/pi-tui` 精确版本 1.2.0，仅用于本地 TUI 与会话工具渲染器；`THIRD_PARTY_NOTICES.md` 声明“OpenClaw 的部分代码改编自 Pi / pi-mono”（MIT，版权 Mario Zechner）[[38]](#ref-38)。但在此之前 OpenClaw 确实嵌入过 Pi SDK：CHANGELOG 记载 2026.2.21 把 `@mariozechner/pi-agent-core`/`pi-ai`/`pi-coding-agent`/`pi-tui` 升级到 0.54.0，2026.3.13 升到 0.58.0，2026.5.12 迁到 `@earendil-works` 命名空间，2026.5.22 “把此前的 Pi agent runtime 内化进 OpenClaw、移除 legacy 包依赖、只保留 Pi 命名的 SDK 别名作为弃用插件兼容”，同一版本还关闭了内嵌 pi-coding-agent 的自动重试，以免 OpenClaw 自己的重试与故障转移循环把失败的工具调用重放一遍[[39]](#ref-39)。

所以准确的表述是：**曾嵌入、现已内化，当前只保留 `pi-tui` 作为终端组件依赖，并保留改编代码的许可证声明**。此外，Pi 仍可作为外部 Coding Agent 被调用：ACP/acpx 的 agent 目标表包含 `pi`，节点会话目录能读取 Pi 的 JSONL 会话存储并用 `pi --session <id>` 续跑——这是“读取/拉起外部 Pi CLI”，不是嵌入它的运行时，两者要分开[[36]](#ref-36)。

### 21.1.5 模型接入与 Provider 抽象

模型接入在 `src/llm/`，插件经契约 `openclaw/plugin-sdk/*` 提供 provider。认证用“profile”抽象：每个 provider 可有多个 profile（OAuth、静态 token、API key），`ModelRuntime` 在每次 run 前解析。启动时为每个已配置 agent 构建一个“准备好的模型运行时代”（auth 模板 + 模型注册表 + 模型目录的原子快照），run 从该快照派生可变存储[[3]](#ref-3)。

provider 认证 profile 的轮换顺序与“按会话粘住”是 21.4.3 故障转移的机制基础：未显式配置时按“OAuth → 静态 token → API key”的类型轮询、最久未用优先，但自动选中的 profile 会按会话粘住以保持 provider 缓存温热，只在 `/new`、`/reset` 或 profile 冷却/禁用时才轮换[[11]](#ref-11)。

### 21.1.6 工具、Skills、Plugin 与 Context Engine

扩展能力分几层。**工具**：OpenClaw 自己的工具定义、参数 schema、tool policy 与 before/after tool-call 适配在 `src/agents/agent-tools*.ts`；内置工具按组归类（文件、exec/process、浏览器、web、message、canvas/nodes、cron/gateway、图像/音乐/视频生成、sessions/subagents 等）[[14]](#ref-14)。**Skills**：`SKILL.md` 目录，按优先级发现（workspace → project agent → personal agent → managed → Workshop → bundled/Custodian → extraDirs + plugin skills），skill 是上下文不是执行权限，技能目录视为受信代码[[15]](#ref-15)。**Plugin**：运行在 Gateway 进程内，可注册 channel/provider/tool/skill/speech/media/image-video/web/context engine/hooks 等十几类能力[[17]](#ref-17)。**Context Engine**：内置 `legacy` 引擎默认使用，插件引擎经 `plugins.slots.contextEngine` 选择；memory 插件与 context engine 分离[[6]](#ref-6)。

### 21.1.7 权限与隔离的分层

OpenClaw 的权限不是一个开关，而是一条多层管线。单层内“deny 先匹配，空 allow=放行未被 deny 者”；跨层按运行时顺序逐层收窄：`tools.profile` → `tools.byProvider.profile` → 全局 `tools.allow/deny` → `tools.byProvider.allow` → agent `tools.allow` → agent `tools.byProvider.allow` → 群/会话 group allow → `tools.toolsBySender` → sandbox allow → subagent allow → runtime allow → inherited[[14]](#ref-14)[[43]](#ref-43)。在这之上还有 `tools.gatewayPolicy`（部署所有者的具体工具与动作白名单，在普通授权之后施加、需重启生效）[[14]](#ref-14)。

另外两条独立的边界：**exec 审批**（决定 shell 命令能否执行、由谁批准，默认值见 21.5.4）和**沙箱**（`agents.defaults.sandbox`，mode 默认 `off`，启用后由 Docker/Podman 后端隔离文件系统与进程）。三者解决的问题不同：tool policy 决定“能不能调用这个工具”，exec 审批决定“这条命令要不要人点头”，沙箱决定“放行后进程能碰什么”。21.5.4 会把这三层串起来。

协议层面还有几个组成事实。WebSocket 帧是 `{type:"req",id,method,params}` / `{type:"res",id,ok,payload|error}` / `{type:"event",event,payload,seq?,stateVersion?}`，协议版本 `PROTOCOL_VERSION=4`；入站帧用 JSON Schema 校验，握手首帧必须是 `connect`，预认证帧上限 64 KiB，每 IP 128 个未认证并发连接[[1]](#ref-1)。Nodes 用同一 WS 服务、以 `role: node` 连接并声明 caps/commands（`camera.*`、`screen.record`、`location.get` 等）；Control UI 是 Gateway 同端口托管的浏览器控制台，经同一套 WS API 通信；WebChat 是静态 UI，经 Gateway WS API 取历史和发送[[1]](#ref-1)。同一端口还复用 OpenAI 兼容 HTTP 面（`/v1/models`、`/v1/chat/completions`、`/v1/responses` 等），`/v1/models` 以 agent 为先[[23]](#ref-23)。

代码模式、压缩规划、文件工具规划走 `WorkerTaskPool`，CPU 准入上限 `max(1, availableParallelism()-1)`，每池默认 128 待处理任务与 256 MiB 报告输入，超限 `WorkerTaskError.code="overloaded"`[[3]](#ref-3)。

渠道适配器的真实接口形态值得写清，因为它修正了“ChannelAdapter 有 start/stop/send/onMessage 方法”的教学构造：渠道插件对 Gateway 暴露 `plugin.config.{isEnabled,isConfigured,isLinked}` 与 `plugin.gateway.{startAccount,stopAccount}`，Gateway 的 `ChannelManager` 负责逐账号生命周期、运行时快照与重启退避；渠道账号快照字段含 `running/connected/lastStartAt/lastDisconnect/restartPending/ingressUnavailable/linked/terminalDisconnect` 等[[41]](#ref-41)。控制面的事件/RPC 面由方法注册表组织，WS 连接处理在 `server/ws-connection.ts`；握手需签名 `connect.challenge` nonce，已配对设备元数据变化需修复配对[[1]](#ref-1)。

内部进程模型：SQLite 通过“spawn broker”与 reader/writer worker 访问（reader 池上限 2、compute 池上限 4、writer 与 singleton 串行）；子运行时（child runtime）是拉起 worker/命令所用的 Node 可执行[[3]](#ref-3)。Gateway 主线程不使用 Node 磁盘模块编译缓存（避免同步 flush 阻塞事件循环），缓存默认在系统临时目录，不可用只产生诊断、启动继续[[23]](#ref-23)。捆绑扩展在 `extensions/` 目录里，渠道类含 discord、slack、signal、imessage、googlechat、msteams、matrix、line、irc、feishu、nostr 等，模型提供方类含 anthropic、openai、google、bedrock、deepseek、ollama、openrouter、litellm 等，能力类含 browser、canvas、memory、diagnostics-otel、prometheus 等[[38]](#ref-38)。
## 21.2 启动：从进程到 Ready

OpenClaw 的启动比单个 CLI 复杂，因为要回答的不是“能不能弹界面”，而是“什么时刻可以接纳渠道消息、什么时刻可以接纳新的 chat.send”。它有明确的分层就绪语义。

### 21.2.1 主路径

公开入口 `startGatewayServer(port=18789, opts)` 的主序是（源码）：安装数据库进程角色 → 安装 Gateway 数据库引擎 → 初始化 SQLite runtime → 获取 Gateway 状态目录锁（防止另一个 Gateway 或 `openclaw agent --local` 同时拥有同一状态目录）→ 校验数据库访问 → `startGatewayServerWithRuntime`[[40]](#ref-40)。失去状态所有权时，优先尝试热重载恢复，否则关闭[[40]](#ref-40)。

随后 `startGatewayServerCore` 创建 Gateway kernel → 提升 skills 快照版本（保证重启后刷新技能目录）→ 创建 HTTP transport（含可选 Tailscale 暴露）→ `finishGatewayStartup`[[40]](#ref-40)。`finishGatewayStartup` 的主序是：附加 WS 连接处理器 → `http.listen` → 并行做 request runtime 就绪、early runtime、主会话孤儿标记 → `runtime.post-attach`（加载启动插件、启动渠道、插件服务、sidecars）→ 就绪判定 → 启动配置热重载器与 TLS 续期 → 提升 last-known-good 配置备份 → 安排 post-ready 维护[[40]](#ref-40)。

一个容易误读的时序是：HTTP 监听先于初始会话投影完成打开；期间 `/healthz` 可返回 200，`/startupz`、`/readyz` 保持 pending，其他请求与 WS 升级等待完整元数据后才认证与路由[[24]](#ref-24)。所以“端口能连”和“可以干活”在 OpenClaw 里是明确分开的。

### 21.2.2 就绪语义与探针

三组无认证探针各管一层[[24]](#ref-24)：

| 探针 | 含义 |
| --- | --- |
| `/health`、`/healthz` | HTTP 存活 |
| `/startup`、`/startupz` | startup sidecars 与 agent 数据库检查/准备完成，且未 draining；不看渠道 |
| `/ready`、`/readyz` | 上述条件 + fresh-chat 准入就绪 + 非 draining + 必需 agent DB 通过 + 已配置渠道账号深度检查通过 |

就绪判定失败原因的求值顺序（源码）是：startup pending → `gateway-draining` → `child-runtime` 不可用 → `state-database` 失败 → `agent-database:<id>` → `plugin-reload` → `chat-recovery`（fresh chat.send 准入未放行）→ 渠道账号逐个评估[[40]](#ref-40)。其中 `chat-recovery` 门值得单独说：启动完成后“保存输入恢复”运行，在其释放 fresh `chat.send` 准入之前，`/readyz` 返回 503 且 `failing:["chat-recovery"]`；也就是说，Gateway 会先恢复上次没来得及处理的输入，再对新输入宣布就绪[[24]](#ref-24)。

`coreReady` 字段进一步把“渠道可用性”从“核心就绪”里排除：它保留 startup、fresh-chat 准入、drain、数据库准入、child-runtime、插件替换检查，但渠道告警不阻止核心就绪；HTTP 状态与 `ready` 仍含渠道健康[[24]](#ref-24)。对运维的含义是：监督者要检测“进程活着但不能干活”，应监控 `/readyz` 而不是 `/healthz`[[24]](#ref-24)。

### 21.2.3 启动守卫与拒绝

OpenClaw 在启动这件事上是“拒绝启动优先于带病运行”。配置层面：JSON5 配置经过严格校验，未知键、类型错误、非法值使 Gateway **拒绝启动**；校验失败时只有诊断命令可用（`doctor`、`logs`、`health`、`status`）；配置文件的写入用“重命名覆盖”原子替换[[26]](#ref-26)。

`gateway run` 之前还有源码级的前置守卫：除非 `--allow-unconfigured`、或 `gateway.mode==="local"`、或（`--dev` 且无配置文件），否则缺配置、缺 `gateway.mode`、`mode` 为其他值都拒绝启动，并附“Config write audit”路径——这是为了防“配置被篡改/清空后 Gateway 带着错误配置跑起来”[[40]](#ref-40)。

运行时安全断言（`assertGatewayRuntimeSecurityConfig`）在启动与热重载候选都执行：非 loopback 绑定且无共享密钥且非 trusted-proxy 会被拒（“refusing to bind gateway to … without auth”）；非 loopback 且启用 Control UI 但无 `allowedOrigins`/`publicOrigin` 也会拒；Tailscale 无认证、trusted-proxy 缺代理 IP 等组合同样拒[[40]](#ref-40)。默认绑定：非容器 → `loopback`，容器环境 → `auto`（需配合认证）[[46]](#ref-46)。

最后是两类数据迁移门：发现遗留 session 存储时，Gateway 拒绝就绪并打印当前 profile 的 Doctor 命令，而不是以空历史静默启动；近期 attested 的 workspace 消失/被清空时，启动拒绝静默重新播种 `BOOTSTRAP.md`[[2]](#ref-2)。这些都是“宁可停下来让运维介入”的例子。

### 21.2.4 配置热重载与 last-known-good

配置热重载模式是 `hybrid`（默认：可热更项热应用，需要时自动重启）或 `off`；旧的 `hot`/`restart` 模式已退役。无效热重载被跳过，运行时保持上一份已接受配置[[26]](#ref-26)。last-known-good 机制：每次成功启动后保存受信副本，但启动与热重载都**不自动恢复**，只有 `openclaw doctor --fix` 会；疑似“误覆盖”的写入（丢掉有效 `gateway.mode`、或文件缩小超过一半）被阻止并保存被拒载荷到 `<path>.rejected.<timestamp>`；含 `***`/`[redacted]` 占位符的候选不会提升为 last-known-good[[26]](#ref-26)。

### 21.2.5 渠道账号启动路径

渠道账号的启动在 `ChannelManager` 里逐账号走：解析账号 → 密钥不可用且 `skipUnavailableAccounts` 时标 skipped → `plugin.config.isEnabled`（否则 disabled）→ 凭据可用性诊断 → `plugin.config.isConfigured`（否则 unconfigured）→ `isLinked` → `plugin.gateway.startAccount(...)` 作为长驻任务。任务退出（无论是否报错）进入自动重启阶梯（见 21.4.2）[[41]](#ref-41)。渠道在“会话投影准备 + 主会话孤儿标记”都完成后才启动；崩溃循环断路器触发时（3 次非干净启动/5 分钟），渠道插件与自启动副服务保持关闭，控制面仍启动，可手动 `openclaw gateway call channels.start` 覆盖[[22]](#ref-22)。

一次 agent run 的创建路径是：`agent` RPC 校验并解析会话 → `agentCommand` → `runEmbeddedAgent`（预留 session controller turn → 全局容量准入 → 解析模型与 auth profile → 构建会话 → 订阅事件）→ 底层循环 `runAgentLoop`[[7]](#ref-7)。

### 21.2.6 onboarding、doctor 与启动期诊断

五分钟本地启动路径（官方运行手册）是：`openclaw gateway --port 18789` → `openclaw gateway status`（健康基线：`Runtime: running`、连通性检查成功，`--require-rpc` 证明读范围 RPC）→ `openclaw status`、`openclaw logs --follow` → `openclaw channels status --probe`[[23]](#ref-23)。配置编辑的“两桶规则”是：根级放基础设施与跨 agent 默认，`agents.defaults` 放 agent-loop 行为，`agents.entries.*` 可覆盖；改配置前用 `config.schema.lookup` 或 `openclaw config schema` 取权威字段文档[[26]](#ref-26)。post-ready 工作延迟 500ms 才释放，sidecars 就绪后再激活 cron 等定时服务；SQLite 快照回收每 15 分钟重复[[40]](#ref-40)。

onboarding 的官方原则是“每条路径都先建立推理（inference）”：`openclaw onboard` 先验证一次真实补全，再配置 workspace、Gateway、服务、渠道、agent、插件；`--classic` 走详细向导，`--install-daemon` 同时安装后台服务，`--non-interactive` 供脚本；README 的最短路径是 `openclaw onboard --install-daemon` → `openclaw gateway status` → `openclaw dashboard`[[35]](#ref-35)。

`openclaw doctor` 是启动前后的检查与修复，姿态有：引导式、`--json` 只读、`--fix`/`--repair` 修复、`--lint` CI 门、`--state-sqlite compact`、`--session-sqlite <mode>`；显式修复会停止匹配的托管 Gateway、检查所有权、做只读 schema 快照后再修复，之后重启并验证就绪；数据库只读快照与初始完整性扫描每库 30 秒上限[[33]](#ref-33)。模型运行时准备在大 fleet 超出启动预算时降级：`openclaw health --json` 报 `modelRuntime.degraded:true` 与 `pendingAgents`，Gateway 继续运行，已完成的 agent 可用，后台准备完成后清除 degraded[[24]](#ref-24)。

端口与协议的启动常量也值得记下：端口解析优先级 `--port` → `OPENCLAW_GATEWAY_PORT` → `gateway.port` → `18789`；已安装服务把解析后的 `--port` 记录在监管元数据里，改 `gateway.port` 后需 `doctor --fix` 或 `gateway install --force`[[23]](#ref-23)。握手后的 `hello-ok` 帧含 `snapshot`（presence、health、stateVersion、uptimeMs）与 `policy`（maxPayload、maxBufferedBytes、tickIntervalMs）；源码里 `TICK_INTERVAL_MS = 30_000`、`HEALTH_REFRESH_INTERVAL_MS = 60_000`、`MAX_BUFFERED_BYTES = 50 MiB`；“Events are not replayed. On sequence gaps, refresh state”——事件不重放，序列有缺口时刷新状态[[23]](#ref-23)。

启动期间的请求处理有一个明确的窗口：主会话恢复在 HTTP 监听绑定后与会话投影并行，二者都完成后才启动渠道并报 ready；启动恢复期间新的 `chat.send` 收到可重试的 startup-unavailable 响应，以免后继消息超过已保存输入；启动期间新回复等待其 agent 的模型与插件运行时加载，已就绪 agent 可继续服务[[22]](#ref-22)。服务托管下的“启动期卡住”有恢复手段：Gateway 持有状态所有权超过 5 分钟却未记录监听绑定时，`openclaw gateway stop` 可恢复该停滞启动[[23]](#ref-23)。

启动期 post-attach 的 sidecars 按依赖推进：内部 hooks（best-effort，失败只记录）→ 并行做模型运行时快照发布、回复运行时预热、WS message handler 预加载 → chat-metadata → channels（`channel-start`）→ ACP 运行时就绪与身份对账 → restart-sentinel → control-ui assets → worker environments；启动子系统用 `startupOutcomes.record({subsystem, status: loaded|skipped|failed, reason})` 汇总[[40]](#ref-40)。

## 21.3 执行：一条消息从进来到出去


### 21.3.1 完整流水线

文档版的入站→出站流水线是：**routing/bindings → 会话键 → 去重 + 去抖 → （若已有活动 run 则）排队 → agent run（流式 + 工具）→ 出站回复（渠道限制 + 分块）**[[5]](#ref-5)。这一步里的每一环都有明确的“事实”要管，下面逐段展开。

入站去重是进程内缓存，键 = agent 范围 + 渠道路由（channel+peer+account+thread）+ message id，TTL 20 分钟或 5000 条先到者（源码常量一致）；这是内存去重，不跨重启，与持久 ingress 去重是两层[[45]](#ref-45)。入站去抖只作用于文本消息、按 channel+会话作用域，媒体/附件立即 flush，控制命令绕过去抖；Telegram 普通文本默认 300ms 安静窗口，其他渠道无通用去抖延迟。注意 `messages.inbound.debounceMs`（入站去抖）与 `messages.queue` 的 500ms 队列 debounce 是两个不同的层[[5]](#ref-5)。

#### 权威状态与进程内状态的不同寿命

以下是根据固定提交的会话、队列与转录设计作出的系统推导。Gateway 拥有会话的调度责任，但“由 Gateway 管”并不代表所有东西都在一次数据库事务里。转录、输入准入、恢复标记与发送记录可以持久化；队列对象、去重 Map、活动流和 PTY 则可能只存在于进程内。审查恢复时，应逐项问状态在哪里、何时算接受、重启后谁能重新发现，不能只画一个 SQLite 圆柱便宣称全部可恢复。[[7]](#ref-7)[[8]](#ref-8)[[9]](#ref-9)

| 状态 | 权威依据 | 重启后可依赖什么 | 不能推出什么 |
| --- | --- | --- | --- |
| 会话转录 | 已提交的存储记录及 writer 声明 | 可回读已提交消息和有效写者 | 未提交的流式片段一定存在 |
| 已准入输入 | 对应入口的持久准入/恢复记录 | 可发现待履行输入 | 网络收到事件就等于已经持久接受 |
| 活动执行 | 当前 run/attempt、控制器及执行权 | 可判断谁获准推进 | 旧进程已对远端做了什么 |
| 进程内去重/排队 | 当前进程对象 | 正常生命周期内减少重复和组织顺序 | 崩溃后仍保留全部去重历史 |
| Shell/PTY | 子进程及工具后端 | 取决于后端是否提供持久作业身份 | 恢复会话一定重连原 PTY |

假设渠道事件到达后，进程在持久准入前退出：本地还没有能够驱动恢复的责任记录，是否再次收到取决于渠道的重投协议。若持久准入已经提交、内存队列尚未创建，恢复应该从持久责任重新安排，而不是要求还原同一个 JavaScript 队列对象。若工具已经在远端执行、转录没有记下结果，恢复只能先核验工具原身份，不能从转录缺失推出未执行。这三个窗口分别属于入口投递、内部调度与外部副作用，不能合并成一个“支持重启”指标。

因此故障演练需要在不同边界终止进程：接受前、接受后、模型完成后、工具已提交后、通知发送中。每一项都记录输入标识、run、attempt、reply operation 和外部对象。观察到进程重启并产生回复，只验证了一条可达路径；还要检查是否重复计费、重复写入、丢失责任，以及恢复权限是否仍有效。这里的演练步骤是本书建议，未声称已对 OpenClaw 执行。

### 21.3.2 路由与会话键

路由匹配层级（源码 `resolveAgentRoute`，`matchedBy`）是：`binding.peer` → `binding.peer.parent` → `binding.peer.wildcard` → `binding.guild+roles` → `binding.guild` → `binding.team` → `binding.account` → `binding.channel` → 兜底。兜底取调用方 owner，否则唯一已配置 agent 或遗留 owner；多 agent 无 owner 且无匹配 binding 时报错，不会取名册第一项[[42]](#ref-42)。多 agent 路由应显式确定 owner，不能依赖名册顺序。

会话键形状决定了隔离粒度：DM 默认 `agent:<agentId>:main`；群 `agent:<agentId>:<channel>:group:<id>`；频道/房间 `agent:<agentId>:<channel>:channel:<id>`；Slack/Discord 线程追加 `:thread:<threadId>`；Telegram 论坛主题 `:topic:<topicId>`。`session.dmScope`、`session.groupScope` 控制折叠[[13]](#ref-13)。Main DM 的路由是固定的：`dmScope=main` 且 `allowFrom` 恰有一个非通配条目时，不匹配的发送者不会覆盖 main 会话的 `lastRoute`；模型不选渠道，回复路由遵循主机配置的固定规则[[13]](#ref-13)。

### 21.3.3 队列模式与会话控制器

入站 auto-reply 经一个进程内“车道感知 FIFO 队列”串行化。并发默认值：未配置的车道 1；`main` 车道 `max(8, CPU 并行度 × 4)`；普通子 agent 队列每个派生会话默认 8；Swarm collector 每组 32[[4]](#ref-4)。优先级只做有限插队：前台工作最多超越普通工作 15 秒、后台最多 30 秒，运行中的 turn 不被抢占[[4]](#ref-4)。

四种队列模式[[4]](#ref-4)：

| 模式 | 含义 |
| --- | --- |
| `steer`（默认） | 注入活动 runtime（含工具执行期间）；首个可执行调用先启动，其未启动的串行尾部可被跳过；并行批不因 steering 跳过 |
| `followup` | 不 steer，当前 run 后另起一轮 |
| `collect` | 静默窗口后把兼容的排队消息合并为单个 followup turn |
| `interrupt` | 中止该会话活动 run，再运行最新消息，且该取消不进入重启恢复续跑 |

默认只允许同请求者 steering；他人 steering 被拒并提示用 `/followup`。`allowCollaborativeSteering:true` 才允许同会话已准入人类输入 steer，文档警告“Bob 可借 Alice 的权限影响行动”[[4]](#ref-4)。使用 durable ingress 的渠道（含 Discord、Telegram）里，`collect` 退化为 `followup`[[4]](#ref-4)。

会话控制器（`src/sessions/session-controller*`）是“下一个哪个 turn 可以使用会话”的所有者，不是运行时执行权限、数据库写围栏或并发限额的替代。文档给出的概念阶段是 `idle/admitting/preparing/running/finishing/settling/mutating`，并强调两句话：“请求取消不等于取消完成；释放调度槽不等于所有旧 writer 已落定”，以及逻辑 session key、存储会话化身、reply operation、后端 attempt 是四种不同身份[[9]](#ref-9)。

### 21.3.4 Agent Loop 与转录写围栏

`agent` RPC 校验参数、解析会话、持久化会话元数据后**立即返回** `{runId, acceptedAt}`；真正的执行在 `runEmbeddedAgent` → `runAgentLoop` 里流式推进，事件经 `agent` 流（`tool`/`assistant`/`lifecycle`）桥接出来；`agent.wait` 等终态，返回 `{status: ok|error|timeout, …}`[[7]](#ref-7)。

转录写围栏是 OpenClaw 对“并发写同一会话”的回答，其作用边界需要与 session write lock 分开：admitted run 在流式输出前记录持久 `activeWriterRunId` 声明；每次转录追加/重写必须带 `expectedWriterRunId`，同步提交事务校验其仍匹配；被取代的 run 无法提交过期转录[[7]](#ref-7)。

循环层（`packages/agent-core/src/agent-loop.ts`）通用骨架见[第17章执行循环](../part3/04-agent-runtime.md#execution-loop)；此处只分析 OpenClaw 的消息检查点：外层 `while(true)` + 内层 `while(hasMoreToolCalls || pendingMessages.length>0)`；每个 turn 发 `turn_start`，先把 pending（steering）消息提交进上下文，调模型流式取 assistant 消息，有 tool call 则 `executeToolCalls`，发 `turn_end`，检查停止条件，在检查点取 steering；内层结束取 follow-up，再复查 steering 以免 `agent_end` 丢失已接受的 steer[[44]](#ref-44)。

几个与失败相关、但属于正常路径设计的点。工具调用批里，steering 出现时未启动的调用被跳过（原因 `steering`）或因 `admission` 致命错误停止[[44]](#ref-44)。工具失败/被拒绝/被取消作为不同结果回送：未启动的调用得到合成配对错误结果以保持转录合法；固定实现对三次连续相同的工具错误（同一签名 + 同一结果哈希）以 `repeated_tool_error` 结束 turn，即使 `tools.loopDetection.enabled=false`[[44]](#ref-44)。回复整形只认整条 final 修剪后为大写 `NO_REPLY` 的静默回复；空 final 不是静默——恢复耗尽时记录失败结果，被准入的用户请求得到一条简短系统通知而不是伪造的回答[[7]](#ref-7)。

#### 写围栏为什么不能替代外部动作防重

session write lock 负责正常生命周期内的互斥；持久 writer 声明让提交事务核对当前写者，二者解决不同阶段的问题。当旧 run 被替换，它即使晚到也不能按旧声明提交转录。固定提交的 agent-loop 文档明确要求 append/rewrite 携带 expectedWriterRunId，并说明同步提交事务检查 activeWriterRunId。这里被保护的是转录提交，不是所有外部资源。[[7]](#ref-7)

| 机制 | 保证 | 不保证 | 故障窗口与恢复责任 |
| --- | --- | --- | --- |
| 会话控制器 | 组织会话阶段与下一轮调度 | 撤销已发出的 HTTP 请求 | 控制器交接时由当前 run owner 清点在途动作 |
| writer 围栏 | 旧 writer 无法覆盖新转录 | 旧工具调用不再产生业务效果 | 工具后端与当前责任人沿原操作核验 |
| SQLite writer 队列 | 对本地存储工作排序 | 外部 API 与本地提交原子化 | 外部成功/本地未记之间仍需对账 |
| 输入 steer/collect | 决定新输入影响哪一轮 | 自动获得新权限或回滚旧动作 | 入口授权与派发时授权都需成立 |

以一个只读文档问答演变为创建提醒的教学例子看：旧 run 已发出创建请求，随后用户 interrupt 并开启新 run。旧 run 的文本回执会被转录围栏挡住，但远端提醒可能已经存在。新 run 若只看到没有成功消息便再创建一次，就制造了重复副作用。正确的应用责任是保留原动作身份，核验提醒是否已经创建，向用户说明取消范围，再决定是否授权删除。工具若没有查询或幂等能力，则应保留未知并转人工，而不是把围栏当作恰好一次执行的证明。

steer 在同一个获准 turn 的检查点补充输入，仍使用当前转录写者的身份；它不意味着新输入可以借旧请求人的权限任意改变目标。collect 在适用入口中将兼容消息合并为后继 turn，后继 turn 要在交接后取得自己的有效声明。对于 durable ingress，文档说明 collect 退化为 followup，各消息保留独立准入生命周期。把两种情况区分清楚，才能解释为什么不能简单合并所有待处理输入再一次性确认。[[4]](#ref-4)

这里还有一个容易忽略的可观测边界：writer 提交被拒是本地竞争的证据，不是外部未执行的证据。记录拒绝时应保留旧/新 run 的关联和外部动作引用，但不要把拒绝当成普通模型错误重新尝试工具。恢复者首先区分“需要重新生成文本”和“需要核验已发出的动作”，避免对同一个失败现象做相反处理。

### 21.3.5 上下文引擎与系统提示

Gateway RPC 层还有一层幂等去重：每个连接状态里一个内存 `Map` 按 `DEDUPE_TTL_MS = 5 分钟` 清理、超过 `DEDUPE_MAX = 1000` 淘汰最旧条目，但不清理仍活动或“已接受待处理”的 run 条目；因为它是进程内 Map，不跨 Gateway 重启，跨重启的重复保护来自持久层（chat.send 的提交回执、durable ingress 去重、出站收据）[[45]](#ref-45)。嵌入的 OpenAI Responses 有个边界：`end_turn:false` 时循环即使只有文本也请求下一次响应；输出上限命中且生成中的 tool call **不执行**，已准入工具完成后从其记录结果重试（受既有有界会话重试预算约束）[[7]](#ref-7)。

Context Engine 的生命周期接口比本章的“四阶段”更全：`bootstrap?`、`maintain?`、`ingest`、`ingestBatch?`、`afterTurn?`、`commitTurn?`、`assemble`、`compact`、`prepareSubagentSpawn?`、`onSubagentEnded?`、`dispose?`。内置 `legacy` 引擎：ingest 是 no-op（会话管理器直接持久化）、assemble 原样返回（sanitize→validate→limit 流水线在别处）、compact 委托内置摘要、没有 afterTurn[[6]](#ref-6)。引擎隔离是设计好的：非 legacy 引擎缺失、契约校验失败或抛错时，在当前 Gateway 进程内“隔离”该引擎并降级为 legacy，错误带操作名记录[[6]](#ref-6)。

首轮 workspace 引导：新 session 第一轮把 `AGENTS.md/SOUL.md/IDENTITY.md/USER.md/BOOTSTRAP.md/MEMORY.md` 注入系统提示的 Project Context；其中 `BOOTSTRAP.md` 只在全新 workspace 创建并在完成后删除，`MEMORY.md` 仅根目录存在时注入；空文件跳过，大文件截断并带标记（单文件 20000 字符、总 60000 字符，源码常量）[[2]](#ref-2)。系统提示由 OpenClaw 基础提示、skills 提示、bootstrap 上下文与逐 run 覆盖构成，官方明确“OpenClaw owns the system prompt for every agent run; there is no runtime default prompt”[[19]](#ref-19)。

### 21.3.6 压缩

压缩把旧轮次摘要为一个 compact 条目写入会话转录，近期消息保持原样；**完整历史仍在磁盘**，压缩只改变模型下一轮看到的内容；分割点保持 assistant 工具调用与其 `toolResult` 成对[[10]](#ref-10)。默认 `mode:"safeguard"`（更严格护栏 + 摘要质量审计），质量审计检查结构、待办请求与精确标识符；耗尽则存“最佳可得摘要”（有意有损），记录 `details.qualityDegraded` 并给可见警告[[10]](#ref-10)。摘要被输出上限截断则拒绝：`mode:"default"` 自动压缩重试一次，手动压缩直接报错不重试；无有效摘要则原历史保留、不写压缩条目[[10]](#ref-10)。

自动压缩默认开启：接近上下文上限，或模型返回上下文溢出错误时压缩并重试（识别数十种 provider 溢出串）；溢出恢复无法适配时，失败回复建议 `/reset`、`/new` 或更大上下文模型[[10]](#ref-10)。自动压缩摘要超时且无可复用摘要时，**“无摘要地提交压缩”**而不是终止 turn——近期消息保持原样，旧细节离开模型上下文（转录仍保留全部消息）[[10]](#ref-10)。

压缩前的“静默记忆冲刷（memory flush）”把持久上下文写入记忆，是可选维护，失败不重置会话或丢历史；`agents.defaults.compaction.memoryFlush.model` 可指定本地模型[[10]](#ref-10)。活动转录字节守卫（opt-in，`maxActiveTranscriptBytes`）超限时用“有界近期历史视图”继续，省略最旧的整轮（工具调用与结果成对保留），不重写已存消息[[10]](#ref-10)。provider 侧压缩受“上下文引擎压缩所有权”约束，保存的 checkpoint 上限 16 MiB，超限或不兼容时回退客户端压缩[[10]](#ref-10)。

### 21.3.7 工具策略求值

单层内 deny 先匹配（含工具组展开、通配符、归一化名），命中即拒绝；deny 之后若 allow 为空则放行，否则必须命中 allow。文档的 “Global tool allow/deny policy (deny wins)” 限于同一层[[14]](#ref-14)。多层顺序（源码）是逐层收窄：`tools.profile` → `tools.byProvider.profile` → 全局 → `tools.byProvider.allow` → agent → agent byProvider → group → `toolsBySender` → sandbox → subagent → runtime → inherited[[43]](#ref-43)。子 agent 有硬拒绝层：总是失去 `gateway`、`agents_list`、`session_status`、`progress_card`、`cron` 等[[16]](#ref-16)。工具结果里 `toolResult.content` 对模型可见，`details` 只是运行时元数据，在 provider 重放与压缩输入前被剥离；持久化转录只保留有界 `details`[[5]](#ref-5)。

### 21.3.8 子 Agent 与多 Agent

子 agent 由 `sessions_spawn` 创建后台 run，会话键 `agent:<agentId>:subagent:<uuid>`；默认递归深度 5、`maxChildrenPerAgent` 5、`maxConcurrent` 8；默认用自己的上下文，需要请求者当前转录时用 `context:"fork"`；请求者会话被沙箱化时，`sessions_spawn` 拒绝会无沙箱运行的目标[[16]](#ref-16)。多 Agent 的“隔离”含义是：每个 agent 有自己的 workspace、状态目录与会话存储；路由选出 agent 后，该 agent 决定用哪个 workspace 与 session store[[13]](#ref-13)。`agents.entries.*` 是当前的 per-agent 配置键（旧 `agents.list[]` 仍出现在个别文档示例，以 schema 为准）[[2]](#ref-2)。

### 21.3.9 系统提示与 hooks

系统提示由 OpenClaw 基础提示、skills 提示、bootstrap 上下文与逐 run 覆盖构成，官方明确“每次 agent run 的系统提示由 OpenClaw 拥有，没有 runtime 默认提示”[[19]](#ref-19)。固定小节含 Tooling、Execution Bias、Promised Work、Care、Runtime Context、Skills、OpenClaw Control、Messaging、Workspace、Sandbox、Assistant Output Directives、Reasoning 等；大块稳定内容（Project Context、固定执行与 provider 策略）放在“提示缓存边界”之上，易变内容（工具目录、Skills、Messaging、Runtime、日期时间）在其下[[19]](#ref-19)。这个分层直接服务于 21.6.5 的缓存：把不常变的放前面，让缓存前缀稳定。

hooks 有两套进程内系统：内部 hooks（`HOOK.md`，如 `agent:bootstrap`、`command:new|reset|stop`）与插件 hooks（`before_model_resolve`、`before_prompt_build`、`before_agent_reply`、`agent_end`、`before/after_compaction`、`before/after_tool_call`、`tool_result_persist`、`message_received/sending/sent`、`session_start/end`、`gateway_start/stop` 等）。`before_tool_call {block:true}` 是终止性的、停止低优先级处理器；`{block:false}` 是 no-op 不清除先前 block[[7]](#ref-7)。

### 21.3.10 回复整形、分块与事件流

回复整形有一组容易误判的边界。静默回复 `NO_REPLY` 仅限连接外部消息渠道的会话，只有整条 final 文本（忽略大小写与周围空白）是 `NO_REPLY` 才请求静默；带引号、JSON 包裹或附在其他文字后都视为可见文本。空 final 不是静默：有界恢复耗尽（仅推理/空输出）时记录失败结果，被准入的用户请求得到一条简短系统通知而不是伪造的回答[[7]](#ref-7)。messaging 工具的发送被追踪，以抑制重复的助手确认；对当前来源会话确认的 final send/reply 结束 run 循环[[7]](#ref-7)。

回复分块默认关闭（`agents.defaults.blockStreamingDefault:"off"`）：`blockStreamingBreak`（`text_end`/`message_end`，默认 `text_end`）、`blockStreamingChunk`（默认 800–1200 字符）、`humanDelay`；非 Telegram 渠道需显式 `*.streaming.block.enabled:true`；分块遵守渠道文本上限并避免拆开围栏代码[[2]](#ref-2)。群聊历史与提示体由渠道插件填充，用 `[Chat messages since your last reply - for context]` 与 `[Current message - respond to this]` 包裹，渲染为“不可信的 user 角色上下文块”[[5]](#ref-5)。工具结果里 `toolResult.content` 对模型可见，`details` 只是运行时元数据，在 provider 重放与压缩输入前剥离[[5]](#ref-5)。

事件流分 `lifecycle`/`assistant`/`tool`；definitive end/error 后事件 owner 抑制该执行的后续客户端 lifecycle 事件；生命周期与工具起止事件投影到“有界、仅元数据”的审计账本（不复制提示词/消息/工具参数/结果/原始错误）[[7]](#ref-7)。chat 通道上，助手 delta 缓冲为 chat `delta`，终态产生 `final|error|aborted`；可重试错误保留 15 秒宽限以等待同一 run 的 fallback 或重启[[7]](#ref-7)。

### 21.3.11 定时作业与调度器

定时作业的定义、运行状态与历史持久在共享 SQLite 状态库，Gateway 必须运行才会触发；每次运行的墙钟预算：`--timeout-seconds`，否则隔离/分离 agent-turn 作业由调度器自带的 60 分钟看门狗约束，command 作业默认 10 分钟，script 载荷默认 5 分钟；启动时逾期且需要 agent 执行的作业被重新排程而不是立即重放，调度器合并错过的 tick[[37]](#ref-37)。Gateway 内核拥有一个 `GatewayScheduler` 管理维护与 cron 唤醒；owner 在启动时重建调度而非持久化第二份调度状态；睡眠后迟到的唤醒对每个到期注册只派发一次[[1]](#ref-1)。

心跳已经从独立子系统迁移为普通自动化作业：`openclaw doctor --fix` 把旧心跳监视器转成可编辑作业（`openclaw automations list --all`），删除 `agents.defaults.heartbeat` 配置，运行时不再解析 `HEARTBEAT.md`；首次设置只一次性创建一个“Proactive check”作业，默认每 30 分钟[[23]](#ref-23)。这是一个“把特殊机制收编成通用机制”的例子，也是版本敏感项（旧提交的心跳还是独立子系统）。
## 21.4 异常：失败、断线与恢复

OpenClaw 作为长期运行、接入外部渠道的系统，异常面比单机 CLI 大得多：渠道会断线，模型会失败，进程会重启，而“已经发生的外部副作用”和“还排着队没做的事”要分开处理。

### 21.4.1 异常分类与证据现状

| 异常类型 | 关键结论 | 依据 |
| --- | --- | --- |
| 渠道断线与重连 | 退避阶梯 5s→5min、最多 10 次；健康监视器做最后 owner；terminal disconnect 不自动重启 | 源码 + 文档 |
| 模型失败与故障转移 | 两阶段：auth profile 轮换 → 模型 fallback；限流 10 次、其他瞬时 8 次/90s | 文档 |
| 超时、取消与停滞 | `timeoutSeconds` 默认 48h；`agent.wait` 30s 不取消 run；看门狗即使诊断关闭也运行 | 文档 |
| 压缩失败与溢出 | safeguard 模式 + 质量审计；无摘要提交压缩；摘要截断拒绝 | 文档 |
| 进程重启恢复 | “Recovery is always on”；24h 窗口、3 次预算；tombstone | 文档 |
| 插件失败 | 加载失败回滚、不阻塞健康插件；崩溃循环断路器 | 文档 + 源码 |
| 投递失败 | 持久出站队列 + 生产者租约；发送后不明确则不重放 | 文档 |

### 21.4.2 渠道断线与重连

渠道账号的自动重启是一个有界阶梯：`RESTART_POLICY = {initialMs:5000, maxMs:300000, factor:2, jitter:0.1}`，`MAX_RESTARTS=10`，用 `RetrySupervisor` 退避；运行满 5 分钟清零重试计数；终止性断连（terminal disconnect，如 WhatsApp 登出）不自动重启[[41]](#ref-41)。

渠道健康监视器在重启阶梯之后作为最后的重启 owner：默认检查间隔 5 分钟、启动宽限 60 秒、冷却 2 个周期（10 分钟）、连接宽限 120 秒、stale 事件阈值 30 分钟；跳过 `terminal-disconnect` 与 `blocked`。健康评估的原因集包括 `unmanaged, not-running, terminal-disconnect, blocked, busy, stuck, startup-connect-grace, reconnect-grace, disconnected, stale-socket, ingress-unavailable`；busy 超过 25 分钟无 run 活动判 `stuck`；stale-socket 只看传输活动（不看应用事件），安静的 Slack/Discord 不因无消息被判不健康[[41]](#ref-41)。

入站 ingress 与传输是两个独立失败域：无法打开 durable ingress 队列的账号“可发不可收”，判 unhealthy（`ingress-unavailable`）[[24]](#ref-24)。终止性断连使 `/readyz` 503 且 `failing:["whatsapp"]`，需运维操作；WhatsApp 重新登录用 `openclaw channels logout && openclaw channels login --verbose`[[24]](#ref-24)。

### 21.4.3 模型失败与故障转移

模型调用恢复分两阶段：先在当前 provider 内做“认证 profile 轮换”，再“模型 fallback”到 `agents.defaults.model.fallbacks` 的下一个；在轮换/切换之前先做有界同模型恢复（延续既有转录、保留部分输出与已完成工作，并指示 agent 先检查被打断的动作再决定是否重复）[[11]](#ref-11)。

重试预算：限流每次连续故障最多 10 次尝试；其他瞬时故障在 90 秒重试窗口内最多 8 次；一次成功的模型响应重置计数/退避/窗口。`retry.provider.maxRetries`（settings.json）可覆盖续跑重试次数，0 禁用但限流仍封顶 10[[12]](#ref-12)。限流冷却只来自 provider 给出的时间（`Retry-After` 等），“OpenClaw does not invent rate-limit cooldowns”；非限流类冷却随 profile 近期错误数递增：30 秒 → 1 分钟 → 5 分钟（上限），计数在失败窗口过后重置，状态存于每 agent SQLite 的 `usageStats`[[11]](#ref-11)。

故障类别映射：5xx 默认 `server_error`，HTTP 529 `overloaded`，408/504/522/524/499 为 `timeout`；格式/无效请求错误通常终止而不轮换。推进 fallback 的错误包括：认证失败、限流与冷却耗尽、过载、超时形、计费 disable、`model_not_found`、以及在活动模型压缩与超大工具结果恢复都无法适配之后的上下文溢出（且该 turn 须可安全重放）[[11]](#ref-11)。计费/额度类失败会标记凭据 disabled（初始 10 分钟）并轮换到下一个可用 profile/provider[[11]](#ref-11)。

一个重要的边界：本地转录写冲突与会话状态变化**不**触发模型重试/换账号/模型 fallback（先重读当前修订并重试一次，重复冲突作为本地错误报告）；`model_fallback_chain_stopped` 是终止性停止[[11]](#ref-11)。另一个易踩的配置：显式用户选择的模型是“严格”的，失败即报错，不会被无关 fallback 掩盖[[11]](#ref-11)。

### 21.4.4 超时、取消与停滞

超时表要分清几个不同的“超时”。`agent.wait` 默认 30s，只等待、不停止 run、不标示阶段；`agents.defaults.timeoutSeconds` 默认 172800s（48h）是每次模型 attempt 的执行预算，每个 fallback 获得新预算故总时长可超过该值；模型 idle timeout 云端 120s / 自托管 300s，模型进度死线是 idle 窗口的 2 倍[[7]](#ref-7)。

取消语义：模型 attempt 超时后，配置的 fallback 在被打断 attempt 落定后以新预算开始；用户取消停止整个 run（含 fallback）；审批等待暂停未用完的 attempt 预算；`ask_user` 提问不暂停预算；压缩可获得一次有界宽限；取消立即撤销该 run 的工具与审批权限[[7]](#ref-7)。终态超时是失败的 turn，不是成功完成；已完成的工具结果保留在转录中[[7]](#ref-7)。

停滞检测：内置两分钟阈值把长时间 `processing` 且无进展的会话分类为 `session.long_running/stalled/stuck`；“controller watchdog”即使诊断关闭也运行，语义停滞阈值至少 5 分钟且至少为告警阈值的 3 倍[[7]](#ref-7)。恢复性继续：当恢复中止一个尚未回复、请求已保存在转录里的交互式 turn 时，尝试一次 continuation（指示不要重复已完成动作），而不是立即要求用户重试[[4]](#ref-4)。

### 21.4.5 压缩失败与上下文溢出

超时的精确数字（本提交文档）：`agent.wait` 默认 30s；`agents.defaults.timeoutSeconds` 默认 172800s（48h），是每次模型 attempt 的执行预算，每个 fallback 获得新预算故总时长可超过该值，进度不重置预算；模型 idle timeout 云端 120s / 自托管 300s，模型进度死线是 idle 窗口的 2 倍[[7]](#ref-7)。限流冷却只来自 provider 给出的时间；非限流类冷却随 profile 近期错误数递增（30s → 1min → 5min 上限），计数在失败窗口过后重置，状态存于每 agent SQLite 的 `usageStats`[[11]](#ref-11)。认证 profile 轮换顺序是：存储的 per-agent 顺序 > `auth.order[provider]` > `auth.profiles` > 存储的 profiles；自动选中的 profile 按会话粘住以保持 provider 缓存温热[[11]](#ref-11)。

自动压缩在接近上限或模型返回溢出错误时触发并重试；溢出恢复无法适配时，失败回复建议 `/reset`、`/new` 或更大上下文模型；恢复压缩本身失败（摘要模型出错/超时）则回复说明并建议重试或 `/compact`[[10]](#ref-10)。取消不是回滚：运行被停止或超时也会停止其溢出/超时恢复；已完成的压缩保留在转录并计数，不发送迟到回复[[10]](#ref-10)。

摘要无时间提交时“无摘要地提交压缩”（见 21.3.6）；若固定指令与工具本身就超过模型窗口，或无更多历史可压缩而请求仍过大，OpenClaw 报告估算的预算组成而非当作无害跳过，用 `/context detail` 找大头[[10]](#ref-10)。手动 `/compact`：turn 活动时给它最多 60 秒完成，之后取消并再等最多 15 秒；Telegram 上先确认命令再中断，避免五分钟采纳看门狗中止并重放[[10]](#ref-10)。

### 21.4.6 重启恢复：Recovery is always on



这个机制将会话写权限与工具执行分开。OpenClaw 的重启恢复**始终开启**：每次 Gateway 进程关闭（含 SIGTERM、systemd/launchd 停止、`openclaw gateway stop`）都保留可恢复的未完成 turn；崩溃、强杀、漏写关机标记后同样进入恢复[[22]](#ref-22)。

恢复的对象与边界（“重启后自动恢复/不恢复”的清单）[[22]](#ref-22)：

- **会恢复**：被重启打断的已准入 turn（含仍在队列或准备执行的）自动继续既有会话；重放保留已记录的工具调用与结果、复用原用户消息；已接受的 `chat.send` 输入在启动时按接受顺序从 SQLite 恢复（过去 24 小时内接受的输入，最多 4 个会话并行重放）；子 agent 由 registry 恢复；cron 会话由调度器恢复；ACP 管理的会话由 IDE/客户端恢复。
- **不会恢复**：从未被准入的工作；Gateway 终端 PTY；后台 exec 进程结束（所属会话在合格时被唤醒一次并告知结果未知）；incognito 输入不能跨进程重启存活。

恢复有预算：main 会话“每个中断周期 3 次计费的自动派发”，三次未能启动后端 turn 耗尽恢复预算，之后 tombstone 直至检查并替换[[22]](#ref-22)。恢复的权限按当前 profile/role/grant/device 策略重新校验原调用者的记录权限；撤销仍停止恢复的 run；无法重建委托发送者权限则终止该 turn、记录一条本地中断通知与终止性 source receipt，防止重放；恢复耗尽后任务与转录保留，WebChat 的 **Resume** 以当前操作者权限继续[[8]](#ref-8)。

文档反复强调的一个口径：“A recovered count at startup means execution resumed. It does not prove that an assistant reply was delivered.”（启动时恢复计数只证明执行续跑，不证明回复已投递）[[22]](#ref-22)。这与“不确定的副作用须先检查再重复”是同一个原则。

优雅重启先排空：请求的重启（`openclaw gateway restart`、需要重启的配置变更、网关更新）不会立刻杀死在途工作；Gateway 停止接纳新工作，等待活动 agent turn 与后台任务完成，上限为排空预算（默认 5 分钟）；排空期间到达的消息被明确的重启错误拒绝（不静默排入将死进程）[[22]](#ref-22)。

#### 重启恢复的三项独立保证

重启恢复应分别验收输入责任、执行连续性与最终投递。固定提交的恢复文档描述了恢复权限重验、预算和保留窗口；这些设置意味着恢复是有边界的协议，不是承诺任意长时间后都自动完成。所谓至少一次推进，也只针对满足持久准入、仍在恢复范围、后端可启动且当前授权有效的工作，不能推广到所有已到达网络的消息。[[8]](#ref-8)[[22]](#ref-22)

| 问题 | 可以依赖的依据 | 不在保证内 | 恢复责任 |
| --- | --- | --- | --- |
| 已接受输入还在吗 | 持久准入与恢复记录 | 接受记录提交前的进程崩溃 | 入口/渠道按协议重投或报告未接受 |
| 哪一轮应该继续 | 会话身份、run 记录、恢复标记与有效权限 | 原内存对象、未提交 token 和任意 PTY 连续性 | 当前 Gateway 重建执行所需状态 |
| 是否重复执行 | 工具原身份与后端防重/查询 | 仅凭转录缺失推定没有副作用 | 工具适配层与核验责任人 |
| 用户是否收到回复 | 独立出站责任与渠道回执 | 模型完成就自动等于送达 | 出站处理器和渠道适配器 |

固定文档明确，超出 24 小时自动 chat-input 恢复窗口，以及旧版本没有可恢复 custody 的输入，仍供显式 resend；这不是超过 24 小时后自动续跑，也不是超过窗口立即删除全部历史。已开始的有界恢复尝试可以在窗口结束后完成。[[22]](#ref-22) 本书建议保留未完成工作的审计入口，由使用者选择人工检查或重发。重新提交前必须核验旧动作；新 Run 不是绕开旧未知结果的许可。若现有部署没有满足业务保留期的审计入口，这属于部署能力缺口，不能用延长模型超时解决。

权限在恢复时可能已变化。例如原发送者被移出群组、设备授权撤销，或者工具 profile 收紧。保存原发送者标识的作用是重新核对责任，而不是永久保存其旧权限。恢复请求应在当前策略下验证；无法重建委托关系时停止并留下可见中断记录。用管理员身份无条件续跑用户任务会改变授权主体，不能作为默认恢复策略。

恢复预算也要区分模型尝试、工具尝试和渠道投递。一个模型 fallback 可能得到新的调用预算，但外部创建操作不能因此获得新的业务身份；一轮回复重发也不应重复模型推理。运维面板至少要分别显示待恢复输入、恢复耗尽、未知工具效果和未完成投递，防止“所有进程健康”掩盖业务责任仍然悬空。这里是本书建议的观测拆分，指标名称并非声明产品已经原生提供。

### 21.4.7 插件失败隔离与崩溃循环

插件加载失败的隔离：插件注册失败的贡献被回滚，失败记录留在插件诊断中，不进入执行范围，也不阻塞健康插件与核心 context-engine 准入；插件替换/重载失败时保留 last-good 运行时，待工作结束后自动重试一次[[17]](#ref-17)。保留工作与在途调用共享 60 秒的 pre-stop 预算[[17]](#ref-17)。

崩溃循环断路器：3 次非干净启动在 5 分钟内触发；触发后下次启动抑制自启动副服务（渠道插件等保持关闭、控制面仍启动），安全模式同时暂停 main 会话重启恢复（不扣尝试次数、不落定、不发通知、不 tombstone），窗口排空后同进程自动重新武装恢复并只恢复一次[[22]](#ref-22)。

### 21.4.8 投递失败与持久出站

出站投递有持久队列：Gateway 启动时检查一次未完成的出站投递；新的 durable enqueue 在事务提交后唤醒恢复；有独立的投递尝试预算；生产者租约 60 秒、每 20 秒续约；租约过期不抹去“发送已开始”的证据，此类条目重放前仍需调和[[12]](#ref-12)。重试原则：每个 HTTP 请求单独重试而非整条多步流，只重试当前步骤以保序，避免重复可能造成重复副作用的操作[[12]](#ref-12)。

渠道层对“发送后不明确”的处理是：Telegram 新文本消息发送只有在 429 flood control 或传输证明请求未开始时才重试；发送后断连/超时/丢响应属于“不明确”，**不重放**（避免重复消息）；可安全重复的操作（如预览编辑）可重试瞬时网络与 5xx[[12]](#ref-12)。渠道发送重试默认 3 次、10% 抖动；Discord Gateway WS 重连另有循环：最多 50 次，2000ms 指数退避到 30000ms，无抖动[[12]](#ref-12)。

机器人对机器人循环保护是入站回复 runner 强制执行的滑动窗口速率守卫：默认 `enabled:true`、`maxEventsPerWindow:20`、`windowSeconds:60`、`cooldownSeconds:60`；A→B 与 B→A 计为同一对；不影响人类消息[[32]](#ref-32)。

#### 出站发送的提交点和不确定结果

固定提交的 Retry 文档给出了两个需要分别理解的事实：持久队列在提交后唤醒恢复、未完成时继续推进；但“发送已开始”的证据不会因租约到期清除，重放前必须调和。它没有在该段定义所有渠道统一的自动调和 API，不能据此承诺全部消息至少一次或至多一次。Telegram 新文本发送则有更明确的策略：已出门且断连/超时的歧义发送不自动重放，以避免重复，代价是可能丢失。[[12]](#ref-12)

| 边界 | 固定文档支持的行为 | 不保证什么 | 故障后的责任 |
| --- | --- | --- | --- |
| durable enqueue 提交后 | 唤醒队列恢复；启动也检查未完成条目 | 提交前已保存责任 | 生产者确认提交，入口保留原身份 |
| producer claim 失效或被替换 | 预留检查 owner/lease，不扣失效 claim 的尝试预算，可获取新 claim | 拿到新租约就可重复发送 | 恢复器先核对发送开始证据 |
| 已开始发送、未确认结果 | 租约失效不抹掉开始证据；未被替换的 owner 可记录迟到结果 | 默认无条件自动重发 | 具体渠道/恢复器调和；无法判定时由运营核验 |
| Telegram 新文本发送后歧义 | 不自动重放；仅 429 或确定请求未开始才重试 | 必达与恰好一次 | 检查渠道与任务原记录，显式决定是否重发 |
| 多步骤发送部分成功 | 已完成步骤不重做，按请求重试 | 多步骤是跨系统原子事务 | 按已完成/未决步骤确认剩余交付 |

这是有限文档证据下的保证边界。本书建议为无法调和的发送建立人工入口，显示原消息身份、内容、已开始证据和已收回执；人工选择重发需接受重复风险，选择不重发需接受丢失风险。文档中的租约数字只决定何时可接管，不决定远端是否收到；“恢复了 N 项”也不能写成“送达了 N 项”。

多块回复更复杂。第一块成功、第二块失败时，整条回复既不是完全未发，也不是完整送达。应保存每块身份与发送事实，后续只处理未决块，或采用渠道支持的更新/替换协议。恢复后重新发送全部内容可能重复第一块；只发最后一块可能让用户看不到前提。长度分块、附件上传与正文通知若跨不同接口，也需要分别确认，不能用一条总成功日志覆盖。

业务验收可以设置“模型终态、转录已提交、出站已接受、渠道已确认”四个观察点。它们有助于定位责任，但不能伪装为跨系统原子事务。对渠道不提供已读语义的情况，系统最多证明渠道确认或收件端接受，不能写成用户已经看见。


### 21.4.9 基础设施故障与启动期失败

除渠道与模型外，基础设施层的失败有各自的语义。子运行时（child runtime）失效：Node 被升级/移除后，Gateway 进程活着但无法拉起子 worker；确认后 `childRuntime.available:false`，`/readyz` 503 且 `failing:["child-runtime"]`；服务托管的 Gateway 请求一次优雅重启，监管不可验证或五分钟内已因此关停则 report-only[[24]](#ref-24)。共享状态完整性失败：终止性准入失败立即使 `/ready`、`/readyz` 503，`failing:["state-database"]`[[24]](#ref-24)。插件替换/恢复：`/readyz` 503 `failing:["plugin-reload"]`，带 `phase: reloading|recovering|failed` 与 deadline；替换期间健康监视器不消耗渠道重启次数[[24]](#ref-24)。

Gateway 状态锁与端口冲突：状态所有权等待最多 5 分钟（每 250ms 检查）；`EADDRINUSE` 重试 20 次×500ms；服务监管下新进程先查现有进程 `/healthz`，健康则让位（systemd 下以退出码 78 退出）[[23]](#ref-23)。启动时配置冲突被分类为“需运维介入，非瞬时错误”`GatewayEffectiveConfigConflictError`，监管者的重启无济于事[[40]](#ref-40)。上下文准备超时：模型上下文转录读取有三分钟本地 worker 死线；本地死线会停止模型 fallback，且不会把 provider 凭据标为不健康[[3]](#ref-3)。worker 池过载：每池默认 128 待处理任务/256 MiB，超限 `WorkerTaskError.code="overloaded"`[[3]](#ref-3)。引擎清理死线：前台引擎 dispose 共享 agent 清理死线，默认 10 秒[[6]](#ref-6)。

诊断默认启用（`diagnostics.enabled:false` 关闭）；`openclaw gateway stability` 输出有界稳定性记录；致命退出、关停超时、重启启动失败在 `logs/stability/` 落盘；`openclaw gateway diagnostics export` 生成脱敏 zip[[24]](#ref-24)。

### 21.4.10 保证成立的范围

本章对循环上限和恢复行为的结论限定在所列固定提交、配置与入口。没有声明的统一步数上限、吞吐基准和推荐并发值，不能由局部默认参数推导出来。调优时应同时设置调用侧截止时间、预算和停止条件，保留终止后在途动作的核验责任。具体旋钮见 21.6，通用预算与未知结果见[Runtime 可靠性](../part3/04-agent-runtime.md#failure-side-effects)。

## 21.5 实践：用好 OpenClaw

这部分每条做法标注“官方建议”或“本书建议”。OpenClaw 的官方建议集中在安全加固上，这符合它的定位：一个长期在线、能替人行动的系统，安全边界比功能更重要。

### 21.5.1 先理解信任模型

官方安全页把范围划得很清楚：每个 Gateway 一个信任边界（单个运营者或互信团队），OpenClaw“不是针对相互敌对用户的安全多租户边界”；混合信任或对抗用户必须拆分信任边界——独立 Gateway 与凭据，最好独立 OS 用户或主机（官方建议）[[25]](#ref-25)。这条决定了下文所有做法的框架：OpenClaw 的安全是“把不可信的人挡在边界外”，不是“在边界内做细粒度多租户隔离”。

### 21.5.2 暴露面：从 loopback 到公网的阶梯

官方暴露面板给的是一个从窄到宽的阶梯（官方建议）：Loopback + SSH 隧道（个人/运维）→ Loopback + Tailscale Serve → Tailnet/LAN 绑定（认证 + 防火墙白名单 + 不公网转发）→ 受信反向代理（`trusted-proxy` 认证、严格 `trustedProxies`）→ 公网（罕见高风险，身份感知代理、TLS、速率限制、严格白名单、沙箱化非 main 会话）[[28]](#ref-28)。核心原则是“在你能解释谁能到达它、如何认证之前，不要暴露 Gateway”；避免直接公网端口转发[[28]](#ref-28)。

与它配套的是保守默认：常规主机安装绑定 loopback；多数渠道对未知 DM 发送者返回配对码而不处理消息；群访问为白名单加 mention 门控。例外是容器镜像默认暴露绑定（需配合认证）[[25]](#ref-25)。网络页把底线说得更直白：永远不要在 `0.0.0.0` 上无认证暴露；Docker 发布端口会绕过主机 INPUT 规则，须在 `DOCKER-USER` 链实施[[28]](#ref-28)。Gateway 认证默认必需：无有效认证路径时拒绝 WebSocket 连接；明文 `ws://` 只接受 loopback、私有 IP 字面量、`.local`、Tailnet[[28]](#ref-28)。

### 21.5.3 DM 策略、配对与命令授权

身份优先于模型，是 OpenClaw 反复强调的顺序。DM 访问策略 `dmPolicy` 有四个值：`pairing`（默认，未知发送者得配对码，码 1 小时过期，Bot 忽略直到批准）、`allowlist`、`open`（须渠道白名单含 `"*"`）、`disabled`；批准只授予 DM 访问，不授予群访问（官方建议“DM 用 pairing + allowlist”）[[32]](#ref-32)[[25]](#ref-25)。群按成员关系而非渠道类型决策：陌生人可加入/发言的房间保留发送者白名单或 mention 门控；回复机器人消息（隐式 mention）不绕过 `groupAllowFrom`[[25]](#ref-25)。

多人能私聊你的 agent 时，官方建议开启 DM 隔离：`session.dmScope:"per-channel-peer"`（推荐）；这是“消息上下文边界，不是主机管理员边界”[[25]](#ref-25)。命令授权：斜杠命令与指令仅对授权发送者生效；渠道白名单为空或含 `"*"` 则命令实际对该渠道开放，这是需要显式关掉的默认[[25]](#ref-25)。

### 21.5.4 工具策略、exec 审批与沙箱

三个默认值要如实说明，因为它们与“更安全”的直觉不完全一致。

第一，工具策略默认 `tools.profile:"full"`：full 选择工具，但不授予 Full Access 执行权限；`minimal` 只有 `session_status` 和仅 update 的 `gateway`；`coding`/`messaging` 是中间档[[14]](#ref-14)。第二，exec 审批默认 `security:"full"`、`ask:"off"`——即受信运营者模型下，网关/节点主机的 exec 默认放行、不询问；“高风险动作默认确认”是本书建议，不是产品默认；审批只能收紧不能放松，需要提示但无 UI 或提示超时时按 `askFallback`（默认 `deny`）解析[[18]](#ref-18)。第三，沙箱默认关闭：`agents.defaults.sandbox.mode` 默认 `off`，`workspaceAccess` 默认 `none`；启用时 `non-main` 沙箱化除 main 外的所有会话，backend 默认 Docker，`sandbox.docker.network` 默认 `"none"`[[24]](#ref-24)。

官方对“高风险动作怎么管”给的是收紧路径，而不是默认：对处理不可信内容的 agent 默认 `tools.deny:["gateway","cron","sessions_spawn","sessions_send"]`；exec 需要审批时用 `ask:"always"` 加 `security:"deny"` 或严格 allowlist；`tools.elevated` 是显式逃生口，保持 `allowFrom` 收窄且不给陌生人；强制创建者角色沙箱（`sandbox:"required"`）在后端不可用时阻断执行，不回退到宿主机[[26]](#ref-26)。官方还给了三个 access profile 的例子：个人 agent（全权限无沙箱）、家庭/工作 agent（沙箱 + 只读工具）、公开 agent（沙箱 + 无文件系统/shell 工具）[[26]](#ref-26)。

### 21.5.5 不可信输入、委托权限与多用户边界

提示注入的官方立场是：模型选择是“第一道、最便宜的一层”，但仍需硬性执行（工具策略、exec 审批、沙箱）；不可信内容（网页、邮件、附件、粘贴日志）本身是威胁面，不止发送者；对“工具启用/读不可信内容”的 agent 使用最新一代最佳档模型，不要用更弱/更小档；若必须用小模型，则只读工具、强沙箱、最小文件系统访问、严格白名单，并关闭 `web_search/web_fetch/browser`；更稳的做法是用只读或禁用工具的“读者 agent”先摘要不可信内容再交给主 agent[[30]](#ref-30)。

下述硬化配置是本书依据官方安全材料组合的示例，使用前应按目标版本验证工具可达性：`gateway:{mode:"local", bind:"loopback", auth:{mode:"token", …}}`、`session.dmScope:"per-channel-peer"`、`tools:{profile:"messaging", deny:["group:automation","group:runtime","group:fs","sessions_spawn","sessions_send"], fs:{workspaceOnly:true}, exec:{security:"deny", ask:"always"}, elevated:{enabled:false}, sessions:{visibility:"agent"}, agentToAgent:{enabled:false}}`，加上 WhatsApp 的 `dmPolicy:"pairing"` 与群 `requireMention:true`；其中一条内置底线是：聊天驱动的 agent turn 中非 owner 发送者无论配置如何都不能使用 `cron` 或 `gateway` 工具[[31]](#ref-31)。只读模式的组合是 `workspaceAccess:"ro"`（或 `none`）+ 工具 allow/deny 阻断 write/edit/apply_patch/exec/process；`tools.exec.applyPatch.workspaceOnly:true`（默认）即使沙箱关闭也阻止 apply_patch 写工作区外；官方还提醒“避免把 home 目录这类宽根目录作为 agent/沙箱工作区”[[26]](#ref-26)。

一个与“更安全默认”直觉相反的默认要写明：跨 provider 的消息工具默认开放。`tools.message.crossContext.allowAcrossProviders` 与 `allowWithinProvider` 默认都是 `true`（升级采用该默认，显式 `false` 保留阻断）；这些控制只守护发送/回复/编辑/删除/建线程等变更动作，不限制读取与管理动作，也不替代文件系统/shell/沙箱限制[[26]](#ref-26)。

多 agent 的价值是隔离，不是“多模型”。每个 agent 有自己的 workspace、状态目录与会话存储；路由选出 agent 后由它决定 workspace 与 session store[[13]](#ref-13)。公开 agent 的官方示例配置是：沙箱 all + workspaceAccess none + 只给 `sessions_list/history/send/spawn`、`session_status` 与渠道工具，deny `apply_patch/browser/canvas/cron/edit/exec/gateway/...`，`tools.sessions.visibility:"tree"`[[26]](#ref-26)。

会话可见性与 agent-to-agent 消息是两件事：工具 profile 不收窄会话工具可见范围，会话可见性是 Gateway 范围，agent-to-agent 消息默认开启——需要配合 `tools.sessions.visibility` 与 `tools.agentToAgent` 配置[[26]](#ref-26)。共享部署的官方提醒是：如果 Slack 里人人都能私聊 bot，核心风险是“委托的工具权限”；公司共享 agent 仅在同一信任边界内可接受，专用机器/VM/容器、专用 OS 用户，且不要把该运行时登录到个人账号[[29]](#ref-29)。

#### 只读群助手的权限可达性

“只读”应限定具体数据和出口。允许读取文档意味着应用已经向成员开放这些文档；允许群内回复意味着所有可见成员都可能接收回答。不能只用个人 DM 的 pairing 证明群成员均有数据权限。把群绑定到一个单独 agent、一个明确的文档集合和受控出口，可以降低跨会话泄漏，但仍需要群成员变化时的授权撤销与缓存失效。

| 验收用例 | 应观察的结果 | 未满足时怎样处理 |
| --- | --- | --- |
| 获准成员问当前允许文档 | 能读取指定来源并在当前群回复 | 检查 allowlist 是否真的包含读取和回复路径 |
| 文档夹带“运行命令” | 仍只能读取，exec/写工具在宿主拒绝 | 停止启用该能力组合并修策略 |
| 陌生成员或授权撤销 | 不读取受限文档、不借其他成员身份 | 重新核验群及发送者作用域 |
| 文件路径越出允许目录 | 文件系统/工具边界拒绝 | 不靠提示中的“不要越界”验收 |
| 沙箱后端不可用 | 依需要沙箱的策略拒绝执行 | 不能悄悄落回宿主执行 |

因此 minimal profile 不能直接当作该案例成品：若它只开放状态查询，没有目标文档读取能力，任务根本不可达。本案例选择固定工具目录中的 `read` 与 `session_status`，用 `tools.profile:"full"` 提供基表，再以 `tools.allow:["read","session_status"]` 收窄；显式 deny `write/edit/apply_patch/exec/process/message/cron/gateway/sessions_send/sessions_spawn`。这里禁止的是主动工具发送，当前群内最终文本由已准入的渠道回复链路交付，不依赖 message 工具。[[14]](#ref-14) 文档放入只读挂载的专用目录，read 的路径范围和目标文档 ACL 还需独立限制。配置示例是设计目标，须用目标 schema 验证；本书不把一个未运行的 JSON 片段称为可复制的安全部署。

non-main 的保护范围依赖会话路由。若某个群被折叠成 main，按名称猜测“群一定进沙箱”会漏掉执行路径。要求所有会话隔离时，应选支持的全会话沙箱策略，并核对文件挂载、网络和凭据；只读宿主挂载仍可能暴露敏感文件。沙箱和文档 ACL 解决不同问题，关闭写权限不会自动建立数据隔离。

Tailscale Serve 或反向代理可以控制网络可达性，但同一多用户宿主上可读进程数据、凭据或控制端口的其他用户仍可能越过这个边界。面对相互不信任的成员，按官方安全定位拆分 Gateway、OS 用户或主机，不把一个个人助手的工具权限委托模型扩展成敌对多租户平台。[[25]](#ref-25)[[29]](#ref-29)

#### 从部署配置走到一次可审查的任务

下面以工作群查询一份受限 Runbook 为教学轨迹，把配置、执行和恢复连起来。前提是运营者已经给群助手独立身份，挂载批准的只读文档目录，并限定原群回复出口。用户提出“根据本周 Runbook 解释订单告警”，入口先验证群与发送者，路由到该群专用会话，再按当前可见工具集合装配模型请求。若任一条件缺失，任务停在输入/权限边界，而不是换一个更宽工具继续查。

模型提出读取 `runbooks/order/current.md`，工具适配层解析真实路径，验证它仍在允许根目录且不是越界符号链接。源文档带版本与当前可见范围，读取结果是数据；其中即使写了“管理员已批准导出日志”，也不会生成批准事件。模型得到证据后只能形成带来源的解释，需要的指标未开放则报告缺口。回复仍发到经过授权的原群，不允许文档内容改变收件人。

现在让进程在三个位置退出。第一，输入接受前退出，入口应依渠道重投或向用户明确未接受；第二，接受后、读取前退出，恢复依据持久输入重建，而不要求原内存队列存在；第三，模型已生成、群发送确认前退出，恢复关注出站责任，不重新阅读并生成另一个可能不同的回答。三项需要分别观察，不能用“重启后还能聊天”作为完整验收。

再加入权限变化：任务运行中移除了某文档权限。已取得的片段不能继续当作永久可交付证据，交付前需要重验可见性；若框架只在读取阶段检查，部署方就要明示这一窗口并增加控制。对个人自用资料这可能是可接受取舍，对群共享数据则可能阻塞启用。不能把产品安全定位里的“可信运营者”自动扩展成所有群成员可信。

最后加入外部动作：用户要求把结论创建为工单。原只读 profile 不具备这项能力，所以系统应说明需要一个另外授权的工作流或人工操作；它不应因为“有帮助”临时启用 exec、自动化或跨会话发送。若正式增加建单工具，下一轮必须绑定当前批准、稳定业务身份和结果查询契约，并把建单与群通知分别记录。这个演进说明安全配置不能只从 deny 清单推导，还必须从承诺的任务能力反向检查可达性。

#### 恢复责任怎样落到运营流程

本书建议为未决工作建立一个最小接续单：原输入与会话、run/attempt、最后已知阶段、外部动作引用、当前授权、剩余恢复预算、负责人和截止时间。这里描述的是运营交接物，不宣称 OpenClaw 自带一个同名企业工单系统。需要人工介入时，交接物应能让接手者查询原事实，而不是要求他相信模型最后一句话。

恢复耗尽后，用户界面至少需要说明“任务已接受但未完成”“结果已生成但投递未确认”或“动作可能发生、需核验”中的哪一种。三种情况对应不同下一步：重新计算、重投结果、查询业务系统。只显示一个红色失败图标，会诱导用户重复提交；只显示成功文本，又可能掩盖通知或外部结果未知。

运维也要分开恢复成功率与任务成功率。某次恢复成功启动模型，不代表原问题解决；原问题已经解决而通知失败，也不能当成模型质量退化。统计时固定分母，区分输入接受数、开始执行数、业务已验收数、渠道已确认数与人工未决数。此处的口径是部署建议，实际采集能力按产品日志、数据库和渠道接口核对。

#### 版本升级与未知效果的交叉窗口

升级可能同时改变转录格式、工具目录与恢复策略。切换前应保存当前版本与可回滚制品，查清哪些输入尚未开始、哪些工具结果未知、哪些出站正在发送。旧进程停止后，对未知外部效果继续沿原身份查询；不能把新版启动看作业务动作已经重置。若数据库迁移不向后兼容，回滚二进制也不等于恢复旧存储语义。

插件升级还会改变权限和可执行代码。某工具名保持不变，目标账户、默认目录或发送行为却可能变化，恢复任务若直接绑定新实现，就不再是原来的执行条件。应核验工具契约兼容性；不兼容时暂停受影响任务，给人工展示差异。软件包版本与业务动作身份的生命周期不同，前者更新不能抹掉后者。

备份恢复需要同样谨慎。将 SQLite 回到昨天的快照，外部工单和消息不会回到昨天。恢复后的本地状态可能缺少已成功效果，重放输入会再次制造动作。真正的恢复演练应包括恢复前后外部对象对账、当前授权和待投递结果，而不仅是数据库能够打开。若部署无法完成这些核验，就把恢复后写工具先收紧为只读，直到原事实被确认。

### 21.5.6 备份、升级与事件响应

节点执行 `system.run` 要当作远程代码执行看待：已配对 macOS 节点可被 Gateway 调用 `system.run`，官方明说“this is remote code execution on that Mac”；配对建立身份与令牌，不是逐命令审批面；全局节点命令策略只匹配精确命令名、不匹配 shell 文本；要完全禁用远程执行，security 设 `deny` 并移除该节点的配对[[26]](#ref-26)。

插件与技能目录都视为受信代码：插件运行在 Gateway 进程内，只从可信来源安装、优先显式 `plugins.allow` 白名单、启用前审查配置、插件代码/元数据/发现根变更后重启；npm/git 插件安装仅在显式 install/update 流程跑包管理器依赖收敛；`security.installPolicy` 让运营者对技能/插件安装返回 allow/warn/block；优先固定精确版本[[26]](#ref-26)。技能目录视为受信代码并限制谁能修改，技能监视器在下一个 agent turn 更新快照[[26]](#ref-26)。

备份的官方硬规则：绝不拷贝活动 `.sqlite`/`-wal`/`-shm`/`-journal` 文件（“raw file copies of a live database can be torn or corrupt”），用 `openclaw backup create --verify`（在线备份 API）；备份含认证配置、渠道与 provider 凭据、会话历史——加密存放、限制目标访问[[34]](#ref-34)。

升级用 `openclaw update`（先校验新版本、再激活并验证）；升级前跑 preflight checklist（已验证的备份、相同服务账号/状态/配置路径、`--dry-run` 检查）[[34]](#ref-34)。降级的官方警告：降级软件包不会反转配置或数据库迁移；受支持的恢复是“用其对应的 OpenClaw 版本恢复升级前备份”；不要降低 schema 版本标记[[34]](#ref-34)。

事件响应 SOP 是 Contain → Rotate → Audit → Collect：Contain（停进程、收紧绑定与 DM 策略）、Rotate（Gateway auth、远程客户端密钥、provider/渠道凭据）、Audit（`openclaw logs`、审阅转录与近期配置变更、重跑 `security audit --deep`）[[29]](#ref-29)。

### 21.5.7 security audit 与 doctor 的实际用法

敏感信息的官方立场是：假设 `~/.openclaw` 下任何内容都可能含秘密/隐私（config、credentials、两个 SQLite、sandboxes 等）；权限 `openclaw.json` 600、`~/.openclaw` 700；全盘加密；共享主机用专用 OS 用户；日志可含工具摘要、错误与 URL，转录可含粘贴的秘密；`logging.redactPatterns` 自定义脱敏（不改变模型上下文或转录）；分享诊断优先用 `openclaw status --all`；修剪不需要长期保留的转录与日志[[29]](#ref-29)。为基于手机号的渠道（WhatsApp/Signal/Telegram），官方建议考虑为助手使用与个人号码分离的独立号码[[31]](#ref-31)。

`openclaw security audit` 是 OpenClaw 把“安全建议”落成可执行检查的机制。官方建议在任何配置变更之后、或暴露网络面之前跑一次：`--deep` 尝试实时 Gateway 检查与插件收集器，`--fix` 应用安全修复，`--json` 输出结构化结果[[27]](#ref-27)。`--fix` 有意收窄：把 open 的群策略翻为 allowlist、收紧 state/config/include 文件权限（文件 600、目录 700，Windows 用 ACL）[[27]](#ref-27)。审计的分诊优先级是：①任何“open”且启用工具先锁 DM/群；②公网暴露立即修；③浏览器控制远程暴露当作 operator 访问；④权限不得 group/world 可读；⑤插件只装明确信任的[[27]](#ref-27)。

`openclaw doctor` 是启动前后的检查与修复：`--json` 只读、`--fix`/`--repair` 修复、`--lint` 做 CI 门；文档提醒 `--non-interactive` 只抑制提示、不抑制写入；显式修复会停止匹配的托管 Gateway、检查所有权、做只读 schema 快照后再修复，之后重启并验证就绪[[33]](#ref-33)。

### 21.5.8 常见问题诊断表

| 现象 | 常见原因 | 对策 | 依据 |
| --- | --- | --- | --- |
| `/readyz` 一直 503 `chat-recovery` | 启动输入恢复还没释放 fresh-chat 准入 | 等恢复完成，或查恢复日志 | 21.2.2 |
| 配置改了不生效 | 无效热重载被跳过；或需要重启的变更 | 看 `appliedConfigHash`；用 `config schema` 核字段 | 21.2.4 |
| 渠道反复重启后 “giving up” | 重启阶梯耗尽（10 次） | 修渠道根因；terminal disconnect 需手动重登 | 21.4.2 |
| 模型报错不换模型 | 显式用户选择的模型是严格的 | 用默认 primary + fallbacks，而非严格选择 | 21.4.3 |
| 重启后回复了重复内容 | 重放保留原消息与工具结果 | 属于恢复语义；不确定副作用先查 | 21.4.6 |
| 工具策略 deny 没挡住 | deny 只在同层；group/sandbox 层可能再放 | 按 21.3.7 的多层顺序逐层查 | 21.3.7 |
| exec 直接执行了命令 | 默认 `security:full, ask:off` | 收紧 exec 审批或沙箱 | 21.5.4 |
| 无人私聊的陌生人也能命令 | 渠道白名单空或含 `"*"` | 收紧 `commands.allowFrom` / 白名单 | 21.5.3 |

### 21.5.9 案例：给一个工作群挂一个只读助手

下面是一个教学构造的案例，未实测，只使用本章已核对过的入口与配置。目标：在 Slack 工作群挂一个助手，只读文档、不执行命令、不改文件、不碰配置。

| 步骤 | 做法 | 依据 |
| --- | --- | --- |
| 信任边界 | 助手跑在专用 OS 用户与独立 Gateway 实例上，不登录个人账号 | 21.5.1 |
| 暴露面 | Gateway 绑定 loopback，需要远程运维时走 Tailscale Serve，不做公网端口转发 | 21.5.2 |
| 群访问 | `groupPolicy` 保留发送者白名单 + mention 门控；回复不绕过白名单 | 21.5.3 |
| 工具策略 | 固定目录下 allow read/session_status，deny 写文件、exec/process、message、自动化和跨会话工具；正常 final 走已准入渠道回复，不开放主动发送 | 21.5.4 |
| exec 与沙箱 | `exec:{security:"deny", ask:"always"}`；`sandbox:{mode:"all", workspaceAccess:"ro"}` 作为本书受限案例目标，需按该版本 schema 确认；若采用 non-main，先证明群路由确实不是 main | 21.5.4 |
| 审计 | 配置后跑 `openclaw security audit --deep`，再跑 `openclaw doctor`，按 critical/warn 逐项处理 | 21.5.7 |

这个案例的验收对象是当前身份下读取、回复与拒绝越权的具体路径。每一步都在把 OpenClaw 的默认从“受信运营者的全权限个人助手”收窄到“只读工作群助手”：信任边界决定谁在系统内，暴露面决定谁能到达它，群白名单决定谁能触发它，工具策略与 exec 审批决定它能做什么，沙箱决定放行后的进程能碰什么，审计把这一切变成可复查的检查。

## 21.6 调优：控制上下文、成本与质量

OpenClaw 提供了完整的旋钮、默认值和观测手段，但**没有官方基准、没有本章的实测**。这一部分只写“旋钮与机制、作用对象、观测手段”，不写“调到多少好”；数字都是固定提交里的默认值。

### 21.6.1 度量与观测

用量与成本的观测入口是 `/status`（模型、上下文使用、最后响应 token、成本）、`/usage off|tokens|full|cost|reset`、`openclaw status --usage`（provider 配额窗口）；成本由 `models.providers.<provider>.models[].cost`（USD/百万 token，input/output/cacheRead/cacheWrite）估算，是估算不是账单[[20]](#ref-20)。上下文的观测用 `/context list`、`/context detail` 看每个注入文件/工具/技能/系统提示的体量[[20]](#ref-20)。

缓存可观测靠 `diagnostics.cacheTrace.enabled`（默认 false，落盘 `logs/cache-trace.jsonl`），观测项含 `input/cacheRead/cacheWrite` 与稳定前缀/易变后缀/工具指纹[[21]](#ref-21)。系统级观测还有 Prometheus 指标 `openclaw_session_recovery_total`、`openclaw_session_recovery_age_seconds`，以及 `OPENCLAW_GATEWAY_STARTUP_TRACE=1` 记录每个启动 span[[24]](#ref-24)。注意 cache trace 默认含提示与消息原文，属敏感数据[[21]](#ref-21)。嵌入式 runtime 在启动期做了“就绪才处理”的预热：`prewarmConfigDrivenReplyRuntime`、预加载 WS message handler、`sidecars.model-runtime` 发布模型快照，使首个连接与首轮复用已准备运行时[[40]](#ref-40)。

### 21.6.2 并发与队列旋钮

并发由三个键控制：`agents.defaults.maxConcurrent`（默认 `max(8, CPU×4)`）、`agents.defaults.subagents.maxConcurrent`（默认 8）、`tools.swarm.maxConcurrent`（默认 32）；每个会话仍串行[[4]](#ref-4)。worker 池：compute 池共享 CPU 准入 `max(1, availableParallelism()-1)`（给 Gateway 留一个 CPU），reader 池上限 2，compute 池上限 4，writer 与 singleton 串行[[3]](#ref-3)。队列的 cap（默认 20）与 drop（默认 `summarize`）只作用于排队消息，不改变运行中的 turn。

### 21.6.3 上下文预算

除 `/usage` 外，上下文拆解用 `/context list`、`/context detail` 看每个注入文件/工具/技能/系统提示的体量；`openclaw doctor` 对“bootstrap 文件超过上下文预算”给出“推荐改进”[[20]](#ref-20)。系统级日志与诊断还有 `openclaw logs --follow`、`openclaw status [--all|--deep|--usage]`、`openclaw health [--verbose|--json]`、`openclaw gateway stability`、`openclaw gateway diagnostics export`（脱敏 zip）；`diagnostics.enabled` 默认开[[24]](#ref-24)。

固定开销的旋钮集中在注入量上：`agents.defaults.bootstrapMaxChars`（20000）、`bootstrapTotalMaxChars`（60000）、`skills.limits.maxSkillsPromptChars`、`agents.defaults.imageMaxDimensionPx`（默认 1200）[[20]](#ref-20)。官方给的降 token 清单是：用 `/compact`；裁剪大工具输出；降低图片尺寸；保持技能描述简短（技能列表注入提示）；探索性冗长工作用更小模型[[20]](#ref-20)。

工具结果大小上限由有效上下文窗口推导另一个降工具声明占用的旋钮是 `tools.codeMode`：启用时普通工具移到“guest catalog bridge”之后，模型通常只见 `exec` 与 `wait`；`executor` 默认 `node`（`node:vm`，官方注明是“受信主机执行，不是安全隔离”），`quickjs` 是加固沙箱；MCP 工具目录按服务器名、工具名排序以稳定提示前缀[[14]](#ref-14)。

工具结果大小上限由有效上下文窗口推导：<100K token 为 16000 字符，≥100K 为 32000，≥200K 为 64000；运行时 context-share 守卫把单个工具结果限制在窗口的 30%[[20]](#ref-20)。会话修剪（`agents.defaults.contextPruning`）默认 `mode:"cache-ttl"`、TTL 5 分钟（Anthropic 家族下插件播种 1h），客户端修剪仅在 TTL 过期且占用 ≥约 30% 时开始；最近三个 assistant turn 与首条用户消息之前永不修剪[[10]](#ref-10)[[23]](#ref-23)。

### 21.6.4 模型与 fallback 调优

fallback 的调优要点（21.4.3 的机制）是：显式用户模型选择严格不回退；agent primary 默认严格、需自带 `fallbacks`；fallback 只对本 turn 生效；上下文溢出恢复耗尽后，已知上下文足够的候选优先[[11]](#ref-11)。认证 profile 按会话粘住以保持缓存温热，只在 `/new`/`/reset` 或冷却/禁用时轮换[[11]](#ref-11)。超时旋钮：`agents.defaults.timeoutSeconds`（172800s）、`models.providers.<id>.timeoutSeconds`（慢速本地/自托管 provider 应先调它再调 agent 超时，且保持 agent 超时不低于它）[[7]](#ref-7)。

### 21.6.5 缓存

`cacheRetention` 取 `none|short|long`持久会话中的普通 agent-turn 定时作业可以在缓存窗口内“保温”：例如一小时缓存配 `openclaw automations edit <job-id> --every 55m`；隔离作业每次新会话不保温；官方提醒“每个 turn 都花 token，比较总用量而不是假设命中”[[21]](#ref-21)。

`cacheRetention` 取 `none|short|long`；**换模型必然开启新缓存谱系**，改 thinking/推理级别也可能失效——官方建议“创建会话时选定模型与 thinking 级别并保持稳定”[[21]](#ref-21)。持久会话中的普通 agent-turn 定时作业可在缓存窗口内“保温”（例：一小时缓存配 `automations edit <job-id> --every 55m`）；隔离作业每次新会话不保温；官方提醒“每个 turn 都花 token，比较总用量而不是假设命中”[[21]](#ref-21)。

### 21.6.6 压缩与存储旋钮

压缩调优键：`agents.defaults.compaction.model` 可用本地小模型专做摘要；`keepRecentTokens` 默认 20000；`maxActiveTranscriptBytes`（opt-in 字节守卫）；`memoryFlush.model`；`notifyUser`[[10]](#ref-10)。会话存储维护影响磁盘占用：`session.maintenance.{mode:"enforce"|"warn", pruneAfter:"30d", archiveDashboardAfter:"7d", maxEntries:5000}`，`openclaw sessions cleanup --enforce` 手动执行[[8]](#ref-8)。

### 21.6.7 不应为省 Token 放弃的东西

有几类开销不能省（本书建议）。工具策略的逐层求值：省掉任何一层的 deny 都意味着越权路径变宽，调优只能改“哪些工具可选”，不能改“deny 是否生效”。exec 审批：共享宿主或可触达敏感资源时，降低打断应先收窄能力；对无宿主敏感挂载、凭据和网络能力的隔离执行器，可依据已批准的威胁模型选择不同询问策略。重启恢复：关闭恢复或调低预算的代价是“已接受的输入可能永远丢失”，这不是 token 问题，是正确性问题。缓存保温可以关（那省费用），但逐动作的权限边界和恢复语义不能为省 token 而放宽。

### 21.6.8 版本敏感事实与重新评估入口

| 项目 | 本章绑定的依据 | 使用边界 |
| --- | --- | --- |
| 提交与包版本 | `ff7597025dd2…`，根包标识 `2026.10.6` | 包版本字符串不证明所有用户安装了这一提交 |
| writer 围栏 | agent-loop 的 activeWriterRunId/expectedWriterRunId | 限制转录提交，不撤销外部动作 |
| 队列模式 | 本提交 queue 文档，durable ingress 下 collect→followup | 入口类型变化时重新检查 |
| 重启窗口与预算 | 自动chat-input选择过去24h接受的输入，已开始尝试可跨窗口结束；main中断周期恢复派发预算3次[[22]](#ref-22) | 超窗输入供显式resend，当前权限无效则不自动续跑 |
| 工具 profile 与沙箱 | minimal含session_status及仅update的gateway；sandbox默认off，启用Docker网络默认none[[14]](#ref-14)[[24]](#ref-24) | 不同安装版本重验能力与挂载，profile不授予额外权限 |
| 出站生产者租约 | 60秒、每20秒续约，按owner/lease预留；发送开始证据在到期后保留[[12]](#ref-12) | 不能凭新claim或到期时间推断可安全重发 |

版本资料原核验日期见章首；本章对 agent-loop、queue、retry 与 restart-recovery 固定资料核对，未进行运行验证。`pi` 的遗留 CLI 别名与 ACP 配置中名为 `pi` 的独立目标不是同一个概念；与 Pi 的关系需按第20章所列依赖和 CHANGELOG 共同判断。


## 本章小结

OpenClaw 把个人 AI 助手做成了一个长期运行的 Gateway：一个进程拥有全部消息面，用 WebSocket 承担控制面，用 SQLite 承担会话与恢复的权威状态，用内置 runtime 执行 agent run，用工具策略、exec 审批和沙箱三层收窄行动边界。它的工程重心有两个：一是“长期在线”带来的可靠性（渠道重连阶梯、模型故障转移、Recovery is always on 的重启恢复、持久出站队列），二是“能替人行动”带来的安全（保守默认、暴露面阶梯、`security audit` 与 `doctor` 的可执行检查）。

两者之间的张力也写进了默认值：exec 默认放行不询问、工具策略默认 full、agent-to-agent 消息默认开启——这些是“个人受信运营者”模型的默认，搬到共享或暴露场景就必须显式收紧。文档反复把这条说清楚：OpenClaw 不是敌对多租户边界，信任边界要靠部署方式去划。

调优方面，OpenClaw 的旋钮（并发、上下文注入预算、fallback 链、缓存保留、压缩与修剪、会话维护）和观测手段（`/usage`、`/context detail`、cache trace、Prometheus）都很完整，但没有官方基准。本章因此只写机制与默认值，把“调到多少好”留给使用者用自己的工作负载去测。

## 参考资料

官方文档页与源码都固定到提交 `ff7597025dd22b087b250a4df94caded50e798aa`（版本 2026.10.6），核验日期 2026-10-23。文档内部不一致处以正文点名处为准。

1. <a id="ref-1"></a> OpenClaw maintainers，[Architecture](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/architecture.md)。
2. <a id="ref-2"></a> OpenClaw maintainers，[Agent](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/agent.md)、[Storage locations](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/storage-locations.md)。
3. <a id="ref-3"></a> OpenClaw maintainers，[Agent runtime architecture](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/agent-runtime-architecture.md)。
4. <a id="ref-4"></a> OpenClaw maintainers，[Command queue](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/queue.md)。
5. <a id="ref-5"></a> OpenClaw maintainers，[Messages](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/messages.md)。
6. <a id="ref-6"></a> OpenClaw maintainers，[Context engine](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/context-engine.md)。
7. <a id="ref-7"></a> OpenClaw maintainers，[Agent loop](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/agent-loop.md)。
8. <a id="ref-8"></a> OpenClaw maintainers，[Session management](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/session.md)、[Session controller](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/session-controller.md)。
9. <a id="ref-9"></a> OpenClaw maintainers，[Session controller](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/session-controller.md)。
10. <a id="ref-10"></a> OpenClaw maintainers，[Compaction](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/compaction.md)、[Session pruning](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/session-pruning.md)。
11. <a id="ref-11"></a> OpenClaw maintainers，[Model failover](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/model-failover.md)。
12. <a id="ref-12"></a> OpenClaw maintainers，[Retry](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/retry.md)。
13. <a id="ref-13"></a> OpenClaw maintainers，[Channel routing](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/channels/channel-routing.md)、[Multi-agent routing](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/multi-agent.md)。
14. <a id="ref-14"></a> OpenClaw maintainers，[Tool policy](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/config-tools/tool-policy.md)。
15. <a id="ref-15"></a> OpenClaw maintainers，[Skills](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/tools/skills.md)。
16. <a id="ref-16"></a> OpenClaw maintainers，[Subagents](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/tools/subagents.md)、[Nesting](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/tools/subagents/nesting.md)、[Subagent tool policy](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/tools/subagents/tool-policy.md)。
17. <a id="ref-17"></a> OpenClaw maintainers，[Plugin architecture](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/plugins/architecture.md)。
18. <a id="ref-18"></a> OpenClaw maintainers，[Exec approvals](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/tools/exec-approvals.md)。
19. <a id="ref-19"></a> OpenClaw maintainers，[System prompt](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/concepts/system-prompt.md)。
20. <a id="ref-20"></a> OpenClaw maintainers，[Token use](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/reference/token-use.md)。
21. <a id="ref-21"></a> OpenClaw maintainers，[Prompt caching](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/reference/prompt-caching.md)。
22. <a id="ref-22"></a> OpenClaw maintainers，[Gateway restart recovery](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/restart-recovery.md)。
23. <a id="ref-23"></a> OpenClaw maintainers，[Gateway running guide](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/index.md)、[Health](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/health.md)、[Gateway lock](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/gateway-lock.md)、[Heartbeat](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/heartbeat.md)、[Service](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/cli/gateway/service.md)。
24. <a id="ref-24"></a> OpenClaw maintainers，[Health](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/health.md)、[Sandboxing](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/sandboxing.md)、[Sandbox modes & backend](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/sandboxing/modes-scope-and-backend.md)、[Workspace access](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/sandboxing/workspace-access.md)。
25. <a id="ref-25"></a> OpenClaw maintainers，[Security](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/security/index.md)、[Access control](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/security/access-control.md)。
26. <a id="ref-26"></a> OpenClaw maintainers，[Tool permissions](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/security/tool-permissions.md)、[Configuration](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/configuration.md)、[Hot reload](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/configuration/hot-reload.md)。
27. <a id="ref-27"></a> OpenClaw maintainers，[Running the audit](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/security/running-the-audit.md)、[Audit checks](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/security/audit-checks.md)。
28. <a id="ref-28"></a> OpenClaw maintainers，[Network exposure](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/security/network-exposure.md)、[Exposure runbook](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/security/exposure-runbook.md)。
29. <a id="ref-29"></a> OpenClaw maintainers，[Secrets & storage](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/security/secrets-and-storage.md)、[Operator incident response](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/security/operator-incident-response.md)。
30. <a id="ref-30"></a> OpenClaw maintainers，[Prompt injection](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/security/prompt-injection.md)。
31. <a id="ref-31"></a> OpenClaw maintainers，[Hardened baseline](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/gateway/security/hardened-baseline.md)。
32. <a id="ref-32"></a> OpenClaw maintainers，[Pairing](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/channels/pairing.md)、[Bot loop protection](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/channels/bot-loop-protection.md)。
33. <a id="ref-33"></a> OpenClaw maintainers，[Doctor](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/cli/doctor.md)、[Running doctor](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/cli/doctor/running.md)、[Doctor checks](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/cli/doctor/checks.md)。
34. <a id="ref-34"></a> OpenClaw maintainers，[Backups](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/install/backups.md)、[Updating](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/install/updating.md)、[Rollback & recovery](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/install/updating/rollback-and-recovery.md)。
35. <a id="ref-35"></a> OpenClaw maintainers，[Onboarding](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/start/onboarding-overview.md)。
36. <a id="ref-36"></a> OpenClaw maintainers，[Session catalogs](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/nodes/session-catalogs.md)、[ACP agents setup](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/tools/acp-agents-setup.md)。
37. <a id="ref-37"></a> OpenClaw maintainers，[Cron jobs: how it works](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/automation/cron-jobs/how-it-works.md)。
38. <a id="ref-38"></a> OpenClaw maintainers，[package.json](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/package.json)、[README](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/README.md)、[THIRD_PARTY_NOTICES.md](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/THIRD_PARTY_NOTICES.md)。
39. <a id="ref-39"></a> OpenClaw maintainers，CHANGELOG：[2026.2.21.md](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.2.21.md)、[2026.3.13.md](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.3.13.md)、[2026.3.22.md](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.3.22.md)、[2026.4.20.md](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.4.20.md)、[2026.4.22.md](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.4.22.md)、[2026.5.12.md](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.5.12.md)、[2026.5.14.md](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.5.14.md)、[2026.5.22.md](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.5.22.md)。
40. <a id="ref-40"></a> OpenClaw maintainers，启动源码：[server.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server.ts)、[server-start.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server-start.ts)、[server-startup-finish.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server-startup-finish.ts)、[server-startup-post-attach.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server-startup-post-attach.ts)、[server/readiness.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server/readiness.ts)、[server-runtime-config.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server-runtime-config.ts)、[pre-bootstrap.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/cli/gateway-cli/pre-bootstrap.ts)。
41. <a id="ref-41"></a> OpenClaw maintainers，渠道源码：[server-channels.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server-channels.ts)、[channel-health-policy.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/channel-health-policy.ts)、[channel-health-monitor.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/channel-health-monitor.ts)。
42. <a id="ref-42"></a> OpenClaw maintainers，[resolve-route.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/routing/resolve-route.ts)。
43. <a id="ref-43"></a> OpenClaw maintainers，[tool-policy-pipeline.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/tool-policy-pipeline.ts)、[conversation-tool-policy-pipeline.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/conversation-tool-policy-pipeline.ts)、[tool-policy-match.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/tool-policy-match.ts)。
44. <a id="ref-44"></a> OpenClaw maintainers，[agent-core agent-loop.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/packages/agent-core/src/agent-loop.ts)、[tool-loop-detection.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/tool-loop-detection.ts)。
45. <a id="ref-45"></a> OpenClaw maintainers，[context-engine/types.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/context-engine/types.ts)、[legacy.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/context-engine/legacy.ts)、[bootstrap.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/embedded-agent-helpers/bootstrap.ts)、[bootstrap-limits.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/embedded-agent-helpers/bootstrap-limits.ts)、[inbound-dedupe.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/auto-reply/reply/inbound-dedupe.ts)。
46. <a id="ref-46"></a> OpenClaw maintainers，[security-cli.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/cli/security-cli.ts)、[fix.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/security/fix.ts)、[net.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/net.ts)。
