# 第19章 OpenClaw 系统：个人 AI 助手的 Gateway 实现、使用与调优

> OpenClaw 的核心不是“又一个聊天机器人”，而是把个人 AI 助手抽象成一个长期运行的本地 Gateway：它接入多渠道消息，管理会话与上下文，调度内置的 Agent Runtime，并用工具策略、沙箱、插件和审批控制行动边界。也正因为它是长期在线、可以替人行动的系统，安全边界比功能更重要。

**资料核验日期：2026-10-23。** 本章分析 OpenClaw 仓库 `openclaw/openclaw` 的固定提交 `ff7597025dd22b087b250a4df94caded50e798aa`（提交时间 2026-10-23 01:51 UTC，根 `package.json` 版本 `2026.10.6`）。官方文档页与源码都固定到这一提交，不引用主分支。OpenClaw 是开源项目，但本章的行为结论来自阅读源码与文档，没有运行验证、没有故障注入；正文里“行为是 A”的准确含义是“固定提交的源码（或文档）写了 A”。

| 标记 | 含义 | 使用方式 |
| --- | --- | --- |
| 源码事实 | 固定提交中可定位的文件与符号、行号 | 只引用 `ff7597025dd2…`，不引用主分支 |
| 官方文档事实 | 同一提交里 docs/ 的陈述 | 注明核验日期与版本 |
| 系统推导 | 根据公开证据做的架构解释 | 不等于产品内部实现 |
| 本书建议 | 本书依据事实归纳的使用建议 | 与官方建议分开标注 |

三条诚实性限制需要提前说明。第一，OpenClaw 文档演进很快：相对现稿的旧提交，docs 目录在两个提交之间变更了 472 个文件（+15,937/−3,552 行），本章对“新旧之间发生了什么变化”的结论绑定这两个提交，不再往后追。第二，文档内部有若干处互相矛盾（技能优先级在两个页面写法不同、会话修剪的持久性说法不一致、`agents.entries` 与旧 `agents.list` 混用），正文遇到会点名，并以源码或更具体的页面为准。第三，官方文档里引用的第三方评测数据（如提示注入攻击成功率）不作为本书结论转述。

## 19.1 组成：OpenClaw 由哪些部分构成

### 19.1.1 定位：一个长期运行的 Gateway

OpenClaw 的形态是一个单一、长期运行的 Gateway 进程，官方文档的表述是“A single long-lived Gateway owns all messaging surfaces”（一个长驻 Gateway 拥有全部消息面）[[1]](#ref-1)。它默认接入 WhatsApp、Telegram、Slack、Discord、Signal、iMessage 和 WebChat 等渠道，README 称“20+ more”；渠道适配器在仓库的 `extensions/` 目录里，实际数量远多于文档摘要列出的几个[[38]](#ref-38)。

Gateway 同时是控制面。控制面客户端（Control UI、CLI、节点）经 WebSocket 连接，默认监听 `127.0.0.1:18789`；每台主机一个 Gateway，由状态目录锁保证唯一（见 19.2.1）。同一端口还复用 HTTP API（`/v1/models`、`/v1/chat/completions`、`/v1/responses` 等 OpenAI 兼容面）、插件 HTTP 路由、Control UI 静态资源和 hooks[[1]](#ref-1)[[23]](#ref-23)。这个“一个端口承担消息渠道、控制面、HTTP API”的设计，是理解它后续安全模型（19.5.2 的暴露面阶梯）的前提。

定位上要分清两件事：OpenClaw 是个人助手的运行时，不是敌对多租户平台。官方安全页明确写，它“不是一个针对相互敌对的用户的安全多租户边界”，混合信任或对抗用户必须拆分信任边界（独立 Gateway 与凭据，最好独立 OS 用户或主机）[[25]](#ref-25)。这一条不是免责声明，而是 19.5 整个实践部分的框架。

### 19.1.2 仓库、包与进程形态

仓库是 pnpm workspace（根 README 注明“不支持在仓库根直接 `npm install`”），根包 `openclaw` 版本 `2026.10.6`，MIT，`engines.node` 为 `>=24.22.1 <25 || >=26.12.1`（README 写的 Node 版本范围与 package.json 不一致，以 package.json 为准）[[38]](#ref-38)。顶层分几块：`src/`（核心：gateway、agents、auto-reply、channels、routing、sessions、context-engine、cron、plugins、skills、security、llm、tui、node-host）、`extensions/`（捆绑的渠道与 provider 插件）、`packages/`（内部包：agent-core、ai、gateway-protocol、gateway-client、plugin-sdk、worker-runtime、retry 等）、`apps/`（各平台应用）、`ui/`（Control UI 前端）、`docs/`[[3]](#ref-3)。

进程形态有几种：前台 `openclaw gateway`，或经 `openclaw gateway install|start|stop` 管理的原生服务（launchd/systemd/Windows 计划任务）；还有独立的节点宿主（`openclaw node`）与 macOS/iOS/Android/Linux 应用。CPU 密集或隔离任务走 worker（`@openclaw/worker-runtime`），SQLite 通过 reader/writer worker 访问（reader 池上限 2 个），Gateway 主线程不使用 Node 磁盘模块编译缓存以避免同步 flush 阻塞事件循环[[3]](#ref-3)[[23]](#ref-23)。

### 19.1.3 状态与存储

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

一个与恢复直接相关的表，是“哪些对象持久化在哪里”（19.4.6 会用到）：对话历史=每 agent SQLite；已接受的聊天跟进=每 agent SQLite 待处理输入；中断的 main 会话 turn=每 agent SQLite 会话行与转录；子 agent 运行=共享状态库；出站投递队列=SQLite；定时作业=SQLite cron store；重启 continuation=SQLite restart sentinel；后台 exec 进程与 Gateway 终端 PTY=仅进程内存，不恢复[[22]](#ref-22)。

### 19.1.4 Agent Runtime 与 Pi 的关系

OpenClaw 拥有内置的 agent runtime，文档明确写“没有外部 agent framework 包保留”；运行时代码在 `src/agents/`、`src/llm/`，可复用核心在私有包 `packages/agent-core`（`@openclaw/agent-core`，`agent-loop.ts` 导出 `runAgentLoop`）[[3]](#ref-3)。内置 runtime id 是 `openclaw`，遗留别名 `pi` 规范化为 `openclaw`，`codex-app-server` 规范化为 `codex`[[3]](#ref-3)。

与 Pi 的关系必须写成有版本绑定的事实，因为它随时间变化过。固定提交的根 `package.json` 里 Pi 系依赖只有一个：`@earendil-works/pi-tui` 精确版本 1.2.0，仅用于本地 TUI 与会话工具渲染器；`THIRD_PARTY_NOTICES.md` 声明“OpenClaw 的部分代码改编自 Pi / pi-mono”（MIT，版权 Mario Zechner）[[38]](#ref-38)。但在此之前 OpenClaw 确实嵌入过 Pi SDK：CHANGELOG 记载 2026.2.21 把 `@mariozechner/pi-agent-core`/`pi-ai`/`pi-coding-agent`/`pi-tui` 升级到 0.54.0，2026.3.13 升到 0.58.0，2026.5.12 迁到 `@earendil-works` 命名空间，2026.5.22 “把此前的 Pi agent runtime 内化进 OpenClaw、移除 legacy 包依赖、只保留 Pi 命名的 SDK 别名作为弃用插件兼容”，同一版本还关闭了内嵌 pi-coding-agent 的自动重试，以免 OpenClaw 自己的重试与故障转移循环把失败的工具调用重放一遍[[39]](#ref-39)。

所以准确的表述是：**曾嵌入、现已内化，当前只保留 `pi-tui` 作为终端组件依赖，并保留改编代码的许可证声明**。此外，Pi 仍可作为外部 Coding Agent 被调用：ACP/acpx 的 agent 目标表包含 `pi`，节点会话目录能读取 Pi 的 JSONL 会话存储并用 `pi --session <id>` 续跑——这是“读取/拉起外部 Pi CLI”，不是嵌入它的运行时，两者要分开[[36]](#ref-36)。

### 19.1.5 模型接入与 Provider 抽象

模型接入在 `src/llm/`，插件经契约 `openclaw/plugin-sdk/*` 提供 provider。认证用“profile”抽象：每个 provider 可有多个 profile（OAuth、静态 token、API key），`ModelRuntime` 在每次 run 前解析。启动时为每个已配置 agent 构建一个“准备好的模型运行时代”（auth 模板 + 模型注册表 + 模型目录的原子快照），run 从该快照派生可变存储[[3]](#ref-3)。

provider 认证 profile 的轮换顺序与“按会话粘住”是 19.4.3 故障转移的机制基础：未显式配置时按“OAuth → 静态 token → API key”的类型轮询、最久未用优先，但自动选中的 profile 会按会话粘住以保持 provider 缓存温热，只在 `/new`、`/reset` 或 profile 冷却/禁用时才轮换[[11]](#ref-11)。

### 19.1.6 工具、Skills、Plugin 与 Context Engine

扩展能力分几层。**工具**：OpenClaw 自己的工具定义、参数 schema、tool policy 与 before/after tool-call 适配在 `src/agents/agent-tools*.ts`；内置工具按组归类（文件、exec/process、浏览器、web、message、canvas/nodes、cron/gateway、图像/音乐/视频生成、sessions/subagents 等）[[14]](#ref-14)。**Skills**：`SKILL.md` 目录，按优先级发现（workspace → project agent → personal agent → managed → Workshop → bundled/Custodian → extraDirs + plugin skills），skill 是上下文不是执行权限，技能目录视为受信代码[[15]](#ref-15)。**Plugin**：运行在 Gateway 进程内，可注册 channel/provider/tool/skill/speech/media/image-video/web/context engine/hooks 等十几类能力[[17]](#ref-17)。**Context Engine**：内置 `legacy` 引擎默认使用，插件引擎经 `plugins.slots.contextEngine` 选择；memory 插件与 context engine 分离[[6]](#ref-6)。

### 19.1.7 权限与隔离的分层

OpenClaw 的权限不是一个开关，而是一条多层管线。单层内“deny 先匹配，空 allow=放行未被 deny 者”；跨层按运行时顺序逐层收窄：`tools.profile` → `tools.byProvider.profile` → 全局 `tools.allow/deny` → `tools.byProvider.allow` → agent `tools.allow` → agent `tools.byProvider.allow` → 群/会话 group allow → `tools.toolsBySender` → sandbox allow → subagent allow → runtime allow → inherited[[14]](#ref-14)[[43]](#ref-43)。在这之上还有 `tools.gatewayPolicy`（部署所有者的具体工具与动作白名单，在普通授权之后施加、需重启生效）[[14]](#ref-14)。

另外两条独立的边界：**exec 审批**（决定 shell 命令能否执行、由谁批准，默认值见 19.5.4）和**沙箱**（`agents.defaults.sandbox`，mode 默认 `off`，启用后由 Docker/Podman 后端隔离文件系统与进程）。三者解决的问题不同：tool policy 决定“能不能调用这个工具”，exec 审批决定“这条命令要不要人点头”，沙箱决定“放行后进程能碰什么”。19.5.4 会把这三层串起来。

协议层面还有几个组成事实。WebSocket 帧是 `{type:"req",id,method,params}` / `{type:"res",id,ok,payload|error}` / `{type:"event",event,payload,seq?,stateVersion?}`，协议版本 `PROTOCOL_VERSION=4`；入站帧用 JSON Schema 校验，握手首帧必须是 `connect`，预认证帧上限 64 KiB，每 IP 128 个未认证并发连接[[1]](#ref-1)。Nodes 用同一 WS 服务、以 `role: node` 连接并声明 caps/commands（`camera.*`、`screen.record`、`location.get` 等）；Control UI 是 Gateway 同端口托管的浏览器控制台，经同一套 WS API 通信；WebChat 是静态 UI，经 Gateway WS API 取历史和发送[[1]](#ref-1)。同一端口还复用 OpenAI 兼容 HTTP 面（`/v1/models`、`/v1/chat/completions`、`/v1/responses` 等），`/v1/models` 以 agent 为先[[23]](#ref-23)。

代码模式、压缩规划、文件工具规划走 `WorkerTaskPool`，CPU 准入上限 `max(1, availableParallelism()-1)`，每池默认 128 待处理任务与 256 MiB 报告输入，超限 `WorkerTaskError.code="overloaded"`[[3]](#ref-3)。

渠道适配器的真实接口形态值得写清，因为它修正了“ChannelAdapter 有 start/stop/send/onMessage 方法”的教学构造：渠道插件对 Gateway 暴露 `plugin.config.{isEnabled,isConfigured,isLinked}` 与 `plugin.gateway.{startAccount,stopAccount}`，Gateway 的 `ChannelManager` 负责逐账号生命周期、运行时快照与重启退避；渠道账号快照字段含 `running/connected/lastStartAt/lastDisconnect/restartPending/ingressUnavailable/linked/terminalDisconnect` 等[[41]](#ref-41)。控制面的事件/RPC 面由方法注册表组织，WS 连接处理在 `server/ws-connection.ts`；握手需签名 `connect.challenge` nonce，已配对设备元数据变化需修复配对[[1]](#ref-1)。

内部进程模型：SQLite 通过“spawn broker”与 reader/writer worker 访问（reader 池上限 2、compute 池上限 4、writer 与 singleton 串行）；子运行时（child runtime）是拉起 worker/命令所用的 Node 可执行[[3]](#ref-3)。Gateway 主线程不使用 Node 磁盘模块编译缓存（避免同步 flush 阻塞事件循环），缓存默认在系统临时目录，不可用只产生诊断、启动继续[[23]](#ref-23)。捆绑扩展在 `extensions/` 目录里，渠道类含 discord、slack、signal、imessage、googlechat、msteams、matrix、line、irc、feishu、nostr 等，模型提供方类含 anthropic、openai、google、bedrock、deepseek、ollama、openrouter、litellm 等，能力类含 browser、canvas、memory、diagnostics-otel、prometheus 等[[38]](#ref-38)。
## 19.2 启动：从进程到 Ready

OpenClaw 的启动比单个 CLI 复杂，因为要回答的不是“能不能弹界面”，而是“什么时刻可以接纳渠道消息、什么时刻可以接纳新的 chat.send”。它有明确的分层就绪语义。

### 19.2.1 主路径

公开入口 `startGatewayServer(port=18789, opts)` 的主序是（源码）：安装数据库进程角色 → 安装 Gateway 数据库引擎 → 初始化 SQLite runtime → 获取 Gateway 状态目录锁（防止另一个 Gateway 或 `openclaw agent --local` 同时拥有同一状态目录）→ 校验数据库访问 → `startGatewayServerWithRuntime`[[40]](#ref-40)。失去状态所有权时，优先尝试热重载恢复，否则关闭[[40]](#ref-40)。

随后 `startGatewayServerCore` 创建 Gateway kernel → 提升 skills 快照版本（保证重启后刷新技能目录）→ 创建 HTTP transport（含可选 Tailscale 暴露）→ `finishGatewayStartup`[[40]](#ref-40)。`finishGatewayStartup` 的主序是：附加 WS 连接处理器 → `http.listen` → 并行做 request runtime 就绪、early runtime、主会话孤儿标记 → `runtime.post-attach`（加载启动插件、启动渠道、插件服务、sidecars）→ 就绪判定 → 启动配置热重载器与 TLS 续期 → 提升 last-known-good 配置备份 → 安排 post-ready 维护[[40]](#ref-40)。

一个容易误读的时序是：HTTP 监听先于初始会话投影完成打开；期间 `/healthz` 可返回 200，`/startupz`、`/readyz` 保持 pending，其他请求与 WS 升级等待完整元数据后才认证与路由[[24]](#ref-24)。所以“端口能连”和“可以干活”在 OpenClaw 里是明确分开的。

### 19.2.2 就绪语义与探针

三组无认证探针各管一层[[24]](#ref-24)：

| 探针 | 含义 |
| --- | --- |
| `/health`、`/healthz` | HTTP 存活 |
| `/startup`、`/startupz` | startup sidecars 与 agent 数据库检查/准备完成，且未 draining；不看渠道 |
| `/ready`、`/readyz` | 上述条件 + fresh-chat 准入就绪 + 非 draining + 必需 agent DB 通过 + 已配置渠道账号深度检查通过 |

就绪判定失败原因的求值顺序（源码）是：startup pending → `gateway-draining` → `child-runtime` 不可用 → `state-database` 失败 → `agent-database:<id>` → `plugin-reload` → `chat-recovery`（fresh chat.send 准入未放行）→ 渠道账号逐个评估[[40]](#ref-40)。其中 `chat-recovery` 门值得单独说：启动完成后“保存输入恢复”运行，在其释放 fresh `chat.send` 准入之前，`/readyz` 返回 503 且 `failing:["chat-recovery"]`；也就是说，Gateway 会先恢复上次没来得及处理的输入，再对新输入宣布就绪[[24]](#ref-24)。

`coreReady` 字段进一步把“渠道可用性”从“核心就绪”里排除：它保留 startup、fresh-chat 准入、drain、数据库准入、child-runtime、插件替换检查，但渠道告警不阻止核心就绪；HTTP 状态与 `ready` 仍含渠道健康[[24]](#ref-24)。对运维的含义是：监督者要检测“进程活着但不能干活”，应监控 `/readyz` 而不是 `/healthz`[[24]](#ref-24)。

### 19.2.3 启动守卫与拒绝

OpenClaw 在启动这件事上是“拒绝启动优先于带病运行”。配置层面：JSON5 配置经过严格校验，未知键、类型错误、非法值使 Gateway **拒绝启动**；校验失败时只有诊断命令可用（`doctor`、`logs`、`health`、`status`）；配置文件的写入用“重命名覆盖”原子替换[[26]](#ref-26)。

`gateway run` 之前还有源码级的前置守卫：除非 `--allow-unconfigured`、或 `gateway.mode==="local"`、或（`--dev` 且无配置文件），否则缺配置、缺 `gateway.mode`、`mode` 为其他值都拒绝启动，并附“Config write audit”路径——这是为了防“配置被篡改/清空后 Gateway 带着错误配置跑起来”[[40]](#ref-40)。

运行时安全断言（`assertGatewayRuntimeSecurityConfig`）在启动与热重载候选都执行：非 loopback 绑定且无共享密钥且非 trusted-proxy 会被拒（“refusing to bind gateway to … without auth”）；非 loopback 且启用 Control UI 但无 `allowedOrigins`/`publicOrigin` 也会拒；Tailscale 无认证、trusted-proxy 缺代理 IP 等组合同样拒[[40]](#ref-40)。默认绑定：非容器 → `loopback`，容器环境 → `auto`（需配合认证）[[46]](#ref-46)。

最后是两类数据迁移门：发现遗留 session 存储时，Gateway 拒绝就绪并打印当前 profile 的 Doctor 命令，而不是以空历史静默启动；近期 attested 的 workspace 消失/被清空时，启动拒绝静默重新播种 `BOOTSTRAP.md`[[2]](#ref-2)。这些都是“宁可停下来让运维介入”的例子。

### 19.2.4 配置热重载与 last-known-good

配置热重载模式是 `hybrid`（默认：可热更项热应用，需要时自动重启）或 `off`；旧的 `hot`/`restart` 模式已退役。无效热重载被跳过，运行时保持上一份已接受配置[[26]](#ref-26)。last-known-good 机制：每次成功启动后保存受信副本，但启动与热重载都**不自动恢复**，只有 `openclaw doctor --fix` 会；疑似“误覆盖”的写入（丢掉有效 `gateway.mode`、或文件缩小超过一半）被阻止并保存被拒载荷到 `<path>.rejected.<timestamp>`；含 `***`/`[redacted]` 占位符的候选不会提升为 last-known-good[[26]](#ref-26)。

### 19.2.5 渠道账号启动路径

渠道账号的启动在 `ChannelManager` 里逐账号走：解析账号 → 密钥不可用且 `skipUnavailableAccounts` 时标 skipped → `plugin.config.isEnabled`（否则 disabled）→ 凭据可用性诊断 → `plugin.config.isConfigured`（否则 unconfigured）→ `isLinked` → `plugin.gateway.startAccount(...)` 作为长驻任务。任务退出（无论是否报错）进入自动重启阶梯（见 19.4.2）[[41]](#ref-41)。渠道在“会话投影准备 + 主会话孤儿标记”都完成后才启动；崩溃循环断路器触发时（3 次非干净启动/5 分钟），渠道插件与自启动副服务保持关闭，控制面仍启动，可手动 `openclaw gateway call channels.start` 覆盖[[22]](#ref-22)。

一次 agent run 的创建路径是：`agent` RPC 校验并解析会话 → `agentCommand` → `runEmbeddedAgent`（预留 session controller turn → 全局容量准入 → 解析模型与 auth profile → 构建会话 → 订阅事件）→ 底层循环 `runAgentLoop`[[7]](#ref-7)。

### 19.2.6 onboarding、doctor 与启动期诊断

五分钟本地启动路径（官方运行手册）是：`openclaw gateway --port 18789` → `openclaw gateway status`（健康基线：`Runtime: running`、连通性检查成功，`--require-rpc` 证明读范围 RPC）→ `openclaw status`、`openclaw logs --follow` → `openclaw channels status --probe`[[23]](#ref-23)。配置编辑的“两桶规则”是：根级放基础设施与跨 agent 默认，`agents.defaults` 放 agent-loop 行为，`agents.entries.*` 可覆盖；改配置前用 `config.schema.lookup` 或 `openclaw config schema` 取权威字段文档[[26]](#ref-26)。post-ready 工作延迟 500ms 才释放，sidecars 就绪后再激活 cron 等定时服务；SQLite 快照回收每 15 分钟重复[[40]](#ref-40)。

onboarding 的官方原则是“每条路径都先建立推理（inference）”：`openclaw onboard` 先验证一次真实补全，再配置 workspace、Gateway、服务、渠道、agent、插件；`--classic` 走详细向导，`--install-daemon` 同时安装后台服务，`--non-interactive` 供脚本；README 的最短路径是 `openclaw onboard --install-daemon` → `openclaw gateway status` → `openclaw dashboard`[[35]](#ref-35)。

`openclaw doctor` 是启动前后的检查与修复，姿态有：引导式、`--json` 只读、`--fix`/`--repair` 修复、`--lint` CI 门、`--state-sqlite compact`、`--session-sqlite <mode>`；显式修复会停止匹配的托管 Gateway、检查所有权、做只读 schema 快照后再修复，之后重启并验证就绪；数据库只读快照与初始完整性扫描每库 30 秒上限[[33]](#ref-33)。模型运行时准备在大 fleet 超出启动预算时降级：`openclaw health --json` 报 `modelRuntime.degraded:true` 与 `pendingAgents`，Gateway 继续运行，已完成的 agent 可用，后台准备完成后清除 degraded[[24]](#ref-24)。

端口与协议的启动常量也值得记下：端口解析优先级 `--port` → `OPENCLAW_GATEWAY_PORT` → `gateway.port` → `18789`；已安装服务把解析后的 `--port` 记录在监管元数据里，改 `gateway.port` 后需 `doctor --fix` 或 `gateway install --force`[[23]](#ref-23)。握手后的 `hello-ok` 帧含 `snapshot`（presence、health、stateVersion、uptimeMs）与 `policy`（maxPayload、maxBufferedBytes、tickIntervalMs）；源码里 `TICK_INTERVAL_MS = 30_000`、`HEALTH_REFRESH_INTERVAL_MS = 60_000`、`MAX_BUFFERED_BYTES = 50 MiB`；“Events are not replayed. On sequence gaps, refresh state”——事件不重放，序列有缺口时刷新状态[[23]](#ref-23)。

启动期间的请求处理有一个明确的窗口：主会话恢复在 HTTP 监听绑定后与会话投影并行，二者都完成后才启动渠道并报 ready；启动恢复期间新的 `chat.send` 收到可重试的 startup-unavailable 响应，以免后继消息超过已保存输入；启动期间新回复等待其 agent 的模型与插件运行时加载，已就绪 agent 可继续服务[[22]](#ref-22)。服务托管下的“启动期卡住”有恢复手段：Gateway 持有状态所有权超过 5 分钟却未记录监听绑定时，`openclaw gateway stop` 可恢复该停滞启动[[23]](#ref-23)。

启动期 post-attach 的 sidecars 按依赖推进：内部 hooks（best-effort，失败只记录）→ 并行做模型运行时快照发布、回复运行时预热、WS message handler 预加载 → chat-metadata → channels（`channel-start`）→ ACP 运行时就绪与身份对账 → restart-sentinel → control-ui assets → worker environments；启动子系统用 `startupOutcomes.record({subsystem, status: loaded|skipped|failed, reason})` 汇总[[40]](#ref-40)。

## 19.3 执行：一条消息从进来到出去



### 19.3.1 完整流水线

文档版的入站→出站流水线是：**routing/bindings → 会话键 → 去重 + 去抖 → （若已有活动 run 则）排队 → agent run（流式 + 工具）→ 出站回复（渠道限制 + 分块）**[[5]](#ref-5)。这一步里的每一环都有明确的“事实”要管，下面逐段展开。

入站去重是进程内缓存，键 = agent 范围 + 渠道路由（channel+peer+account+thread）+ message id，TTL 20 分钟或 5000 条先到者（源码常量一致）；这是内存去重，不跨重启，与持久 ingress 去重是两层[[45]](#ref-45)。入站去抖只作用于文本消息、按 channel+会话作用域，媒体/附件立即 flush，控制命令绕过去抖；Telegram 普通文本默认 300ms 安静窗口，其他渠道无通用去抖延迟。注意 `messages.inbound.debounceMs`（入站去抖）与 `messages.queue` 的 500ms 队列 debounce 是两个不同的层[[5]](#ref-5)。

### 19.3.2 路由与会话键

路由匹配层级（源码 `resolveAgentRoute`，`matchedBy`）是：`binding.peer` → `binding.peer.parent` → `binding.peer.wildcard` → `binding.guild+roles` → `binding.guild` → `binding.team` → `binding.account` → `binding.channel` → 兜底。兜底取调用方 owner，否则唯一已配置 agent 或遗留 owner；多 agent 无 owner 且无匹配 binding 时报错，不会取名册第一项[[42]](#ref-42)。这条顺序比现稿多了一个 `peer.wildcard`，末项也不是“default agent”。

会话键形状决定了隔离粒度：DM 默认 `agent:<agentId>:main`；群 `agent:<agentId>:<channel>:group:<id>`；频道/房间 `agent:<agentId>:<channel>:channel:<id>`；Slack/Discord 线程追加 `:thread:<threadId>`；Telegram 论坛主题 `:topic:<topicId>`。`session.dmScope`、`session.groupScope` 控制折叠[[13]](#ref-13)。Main DM 的路由是固定的：`dmScope=main` 且 `allowFrom` 恰有一个非通配条目时，不匹配的发送者不会覆盖 main 会话的 `lastRoute`；模型不选渠道，回复路由遵循主机配置的固定规则[[13]](#ref-13)。

### 19.3.3 队列模式与会话控制器

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

### 19.3.4 Agent Loop 与转录写围栏

`agent` RPC 校验参数、解析会话、持久化会话元数据后**立即返回** `{runId, acceptedAt}`；真正的执行在 `runEmbeddedAgent` → `runAgentLoop` 里流式推进，事件经 `agent` 流（`tool`/`assistant`/`lifecycle`）桥接出来；`agent.wait` 等终态，返回 `{status: ok|error|timeout, …}`[[7]](#ref-7)。

转录写围栏是 OpenClaw 对“并发写同一会话”的回答，也是现稿“session write lock”需要修正的地方：admitted run 在流式输出前记录持久 `activeWriterRunId` 声明；每次转录追加/重写必须带 `expectedWriterRunId`，同步提交事务校验其仍匹配；被取代的 run 无法提交过期转录[[7]](#ref-7)。

循环层（`packages/agent-core/src/agent-loop.ts`）与第17、18章是同一套骨架：外层 `while(true)` + 内层 `while(hasMoreToolCalls || pendingMessages.length>0)`；每个 turn 发 `turn_start`，先把 pending（steering）消息提交进上下文，调模型流式取 assistant 消息，有 tool call 则 `executeToolCalls`，发 `turn_end`，检查停止条件，在检查点取 steering；内层结束取 follow-up，再复查 steering 以免 `agent_end` 丢失已接受的 steer[[44]](#ref-44)。

几个与失败相关、但属于正常路径设计的点。工具调用批里，steering 出现时未启动的调用被跳过（原因 `steering`）或因 `admission` 致命错误停止[[44]](#ref-44)。工具失败/被拒绝/被取消作为不同结果回送：未启动的调用得到合成配对错误结果以保持转录合法；三次连续相同的工具错误（同一签名 + 同一结果哈希）总会以 `repeated_tool_error` 结束 turn，即使 `tools.loopDetection.enabled=false`[[44]](#ref-44)。回复整形只认整条 final 修剪后为大写 `NO_REPLY` 的静默回复；空 final 不是静默——恢复耗尽时记录失败结果，被准入的用户请求得到一条简短系统通知而不是伪造的回答[[7]](#ref-7)。

### 19.3.5 上下文引擎与系统提示

Gateway RPC 层还有一层幂等去重：每个连接状态里一个内存 `Map` 按 `DEDUPE_TTL_MS = 5 分钟` 清理、超过 `DEDUPE_MAX = 1000` 淘汰最旧条目，但不清理仍活动或“已接受待处理”的 run 条目；因为它是进程内 Map，不跨 Gateway 重启，跨重启的重复保护来自持久层（chat.send 的提交回执、durable ingress 去重、出站收据）[[45]](#ref-45)。嵌入的 OpenAI Responses 有个边界：`end_turn:false` 时循环即使只有文本也请求下一次响应；输出上限命中且生成中的 tool call **不执行**，已准入工具完成后从其记录结果重试（受既有有界会话重试预算约束）[[7]](#ref-7)。

Context Engine 的生命周期接口比现稿的“四阶段”更全：`bootstrap?`、`maintain?`、`ingest`、`ingestBatch?`、`afterTurn?`、`commitTurn?`、`assemble`、`compact`、`prepareSubagentSpawn?`、`onSubagentEnded?`、`dispose?`。内置 `legacy` 引擎：ingest 是 no-op（会话管理器直接持久化）、assemble 原样返回（sanitize→validate→limit 流水线在别处）、compact 委托内置摘要、没有 afterTurn[[6]](#ref-6)。引擎隔离是设计好的：非 legacy 引擎缺失、契约校验失败或抛错时，在当前 Gateway 进程内“隔离”该引擎并降级为 legacy，错误带操作名记录[[6]](#ref-6)。

首轮 workspace 引导：新 session 第一轮把 `AGENTS.md/SOUL.md/IDENTITY.md/USER.md/BOOTSTRAP.md/MEMORY.md` 注入系统提示的 Project Context；其中 `BOOTSTRAP.md` 只在全新 workspace 创建并在完成后删除，`MEMORY.md` 仅根目录存在时注入；空文件跳过，大文件截断并带标记（单文件 20000 字符、总 60000 字符，源码常量）[[2]](#ref-2)。系统提示由 OpenClaw 基础提示、skills 提示、bootstrap 上下文与逐 run 覆盖构成，官方明确“OpenClaw owns the system prompt for every agent run; there is no runtime default prompt”[[19]](#ref-19)。

### 19.3.6 压缩

压缩把旧轮次摘要为一个 compact 条目写入会话转录，近期消息保持原样；**完整历史仍在磁盘**，压缩只改变模型下一轮看到的内容；分割点保持 assistant 工具调用与其 `toolResult` 成对[[10]](#ref-10)。默认 `mode:"safeguard"`（更严格护栏 + 摘要质量审计），质量审计检查结构、待办请求与精确标识符；耗尽则存“最佳可得摘要”（有意有损），记录 `details.qualityDegraded` 并给可见警告[[10]](#ref-10)。摘要被输出上限截断则拒绝：`mode:"default"` 自动压缩重试一次，手动压缩直接报错不重试；无有效摘要则原历史保留、不写压缩条目[[10]](#ref-10)。

自动压缩默认开启：接近上下文上限，或模型返回上下文溢出错误时压缩并重试（识别数十种 provider 溢出串）；溢出恢复无法适配时，失败回复建议 `/reset`、`/new` 或更大上下文模型[[10]](#ref-10)。自动压缩摘要超时且无可复用摘要时，**“无摘要地提交压缩”**而不是终止 turn——近期消息保持原样，旧细节离开模型上下文（转录仍保留全部消息）[[10]](#ref-10)。

### 19.3.7 工具策略求值

单层内 deny 先匹配（含工具组展开、通配符、归一化名），命中即拒绝；deny 之后若 allow 为空则放行，否则必须命中 allow。文档的 “Global tool allow/deny policy (deny wins)” 限于同一层[[14]](#ref-14)。多层顺序（源码）是逐层收窄：`tools.profile` → `tools.byProvider.profile` → 全局 → `tools.byProvider.allow` → agent → agent byProvider → group → `toolsBySender` → sandbox → subagent → runtime → inherited[[43]](#ref-43)。子 agent 有硬拒绝层：总是失去 `gateway`、`agents_list`、`session_status`、`progress_card`、`cron` 等[[16]](#ref-16)。工具结果里 `toolResult.content` 对模型可见，`details` 只是运行时元数据，在 provider 重放与压缩输入前被剥离；持久化转录只保留有界 `details`[[5]](#ref-5)。

### 19.3.8 子 Agent 与多 Agent

子 agent 由 `sessions_spawn` 创建后台 run，会话键 `agent:<agentId>:subagent:<uuid>`；默认递归深度 5、`maxChildrenPerAgent` 5、`maxConcurrent` 8；默认用自己的上下文，需要请求者当前转录时用 `context:"fork"`；请求者会话被沙箱化时，`sessions_spawn` 拒绝会无沙箱运行的目标[[16]](#ref-16)。多 Agent 的“隔离”含义是：每个 agent 有自己的 workspace、状态目录与会话存储；路由选出 agent 后，该 agent 决定用哪个 workspace 与 session store[[13]](#ref-13)。`agents.entries.*` 是当前的 per-agent 配置键（旧 `agents.list[]` 仍出现在个别文档示例，以 schema 为准）[[2]](#ref-2)。### 19.3.9 系统提示与 hooks

系统提示由 OpenClaw 基础提示、skills 提示、bootstrap 上下文与逐 run 覆盖构成，官方明确“每次 agent run 的系统提示由 OpenClaw 拥有，没有 runtime 默认提示”[[19]](#ref-19)。固定小节含 Tooling、Execution Bias、Promised Work、Care、Runtime Context、Skills、OpenClaw Control、Messaging、Workspace、Sandbox、Assistant Output Directives、Reasoning 等；大块稳定内容（Project Context、固定执行与 provider 策略）放在“提示缓存边界”之上，易变内容（工具目录、Skills、Messaging、Runtime、日期时间）在其下[[19]](#ref-19)。这个分层直接服务于 19.6.5 的缓存：把不常变的放前面，让缓存前缀稳定。

hooks 有两套进程内系统：内部 hooks（`HOOK.md`，如 `agent:bootstrap`、`command:new|reset|stop`）与插件 hooks（`before_model_resolve`、`before_prompt_build`、`before_agent_reply`、`agent_end`、`before/after_compaction`、`before/after_tool_call`、`tool_result_persist`、`message_received/sending/sent`、`session_start/end`、`gateway_start/stop` 等）。`before_tool_call {block:true}` 是终止性的、停止低优先级处理器；`{block:false}` 是 no-op 不清除先前 block[[7]](#ref-7)。

### 19.3.10 回复整形、分块与事件流

回复整形有一组容易误判的边界。静默回复 `NO_REPLY` 仅限连接外部消息渠道的会话，只有整条 final 文本（忽略大小写与周围空白）是 `NO_REPLY` 才请求静默；带引号、JSON 包裹或附在其他文字后都视为可见文本。空 final 不是静默：有界恢复耗尽（仅推理/空输出）时记录失败结果，被准入的用户请求得到一条简短系统通知而不是伪造的回答[[7]](#ref-7)。messaging 工具的发送被追踪，以抑制重复的助手确认；对当前来源会话确认的 final send/reply 结束 run 循环[[7]](#ref-7)。

回复分块默认关闭（`agents.defaults.blockStreamingDefault:"off"`）：`blockStreamingBreak`（`text_end`/`message_end`，默认 `text_end`）、`blockStreamingChunk`（默认 800–1200 字符）、`humanDelay`；非 Telegram 渠道需显式 `*.streaming.block.enabled:true`；分块遵守渠道文本上限并避免拆开围栏代码[[2]](#ref-2)。群聊历史与提示体由渠道插件填充，用 `[Chat messages since your last reply - for context]` 与 `[Current message - respond to this]` 包裹，渲染为“不可信的 user 角色上下文块”[[5]](#ref-5)。工具结果里 `toolResult.content` 对模型可见，`details` 只是运行时元数据，在 provider 重放与压缩输入前剥离[[5]](#ref-5)。

事件流分 `lifecycle`/`assistant`/`tool`；definitive end/error 后事件 owner 抑制该执行的后续客户端 lifecycle 事件；生命周期与工具起止事件投影到“有界、仅元数据”的审计账本（不复制提示词/消息/工具参数/结果/原始错误）[[7]](#ref-7)。chat 通道上，助手 delta 缓冲为 chat `delta`，终态产生 `final|error|aborted`；可重试错误保留 15 秒宽限以等待同一 run 的 fallback 或重启[[7]](#ref-7)。

### 19.3.11 定时作业与调度器

定时作业的定义、运行状态与历史持久在共享 SQLite 状态库，Gateway 必须运行才会触发；每次运行的墙钟预算：`--timeout-seconds`，否则隔离/分离 agent-turn 作业由调度器自带的 60 分钟看门狗约束，command 作业默认 10 分钟，script 载荷默认 5 分钟；启动时逾期且需要 agent 执行的作业被重新排程而不是立即重放，调度器合并错过的 tick[[37]](#ref-37)。Gateway 内核拥有一个 `GatewayScheduler` 管理维护与 cron 唤醒；owner 在启动时重建调度而非持久化第二份调度状态；睡眠后迟到的唤醒对每个到期注册只派发一次[[1]](#ref-1)。

心跳已经从独立子系统迁移为普通自动化作业：`openclaw doctor --fix` 把旧心跳监视器转成可编辑作业（`openclaw automations list --all`），删除 `agents.defaults.heartbeat` 配置，运行时不再解析 `HEARTBEAT.md`；首次设置只一次性创建一个“Proactive check”作业，默认每 30 分钟[[23]](#ref-23)。这是一个“把特殊机制收编成通用机制”的例子，也是版本敏感项（旧提交的心跳还是独立子系统）。
## 19.4 异常：失败、断线与恢复

OpenClaw 作为长期运行、接入外部渠道的系统，异常面比单机 CLI 大得多：渠道会断线，模型会失败，进程会重启，而“已经发生的外部副作用”和“还排着队没做的事”要分开处理。

### 19.4.1 异常分类与证据现状

| 异常类型 | 关键结论 | 依据 |
| --- | --- | --- |
| 渠道断线与重连 | 退避阶梯 5s→5min、最多 10 次；健康监视器做最后 owner；terminal disconnect 不自动重启 | 源码 + 文档 |
| 模型失败与故障转移 | 两阶段：auth profile 轮换 → 模型 fallback；限流 10 次、其他瞬时 8 次/90s | 文档 |
| 超时、取消与停滞 | `timeoutSeconds` 默认 48h；`agent.wait` 30s 不取消 run；看门狗即使诊断关闭也运行 | 文档 |
| 压缩失败与溢出 | safeguard 模式 + 质量审计；无摘要提交压缩；摘要截断拒绝 | 文档 |
| 进程重启恢复 | “Recovery is always on”；24h 窗口、3 次预算；tombstone | 文档 |
| 插件失败 | 加载失败回滚、不阻塞健康插件；崩溃循环断路器 | 文档 + 源码 |
| 投递失败 | 持久出站队列 + 生产者租约；发送后不明确则不重放 | 文档 |

### 19.4.2 渠道断线与重连

渠道账号的自动重启是一个有界阶梯：`RESTART_POLICY = {initialMs:5000, maxMs:300000, factor:2, jitter:0.1}`，`MAX_RESTARTS=10`，用 `RetrySupervisor` 退避；运行满 5 分钟清零重试计数；终止性断连（terminal disconnect，如 WhatsApp 登出）不自动重启[[41]](#ref-41)。

渠道健康监视器在重启阶梯之后作为最后的重启 owner：默认检查间隔 5 分钟、启动宽限 60 秒、冷却 2 个周期（10 分钟）、连接宽限 120 秒、stale 事件阈值 30 分钟；跳过 `terminal-disconnect` 与 `blocked`。健康评估的原因集包括 `unmanaged, not-running, terminal-disconnect, blocked, busy, stuck, startup-connect-grace, reconnect-grace, disconnected, stale-socket, ingress-unavailable`；busy 超过 25 分钟无 run 活动判 `stuck`；stale-socket 只看传输活动（不看应用事件），安静的 Slack/Discord 不因无消息被判不健康[[41]](#ref-41)。

入站 ingress 与传输是两个独立失败域：无法打开 durable ingress 队列的账号“可发不可收”，判 unhealthy（`ingress-unavailable`）[[24]](#ref-24)。终止性断连使 `/readyz` 503 且 `failing:["whatsapp"]`，需运维操作；WhatsApp 重新登录用 `openclaw channels logout && openclaw channels login --verbose`[[24]](#ref-24)。

### 19.4.3 模型失败与故障转移

模型调用恢复分两阶段：先在当前 provider 内做“认证 profile 轮换”，再“模型 fallback”到 `agents.defaults.model.fallbacks` 的下一个；在轮换/切换之前先做有界同模型恢复（延续既有转录、保留部分输出与已完成工作，并指示 agent 先检查被打断的动作再决定是否重复）[[11]](#ref-11)。

重试预算：限流每次连续故障最多 10 次尝试；其他瞬时故障在 90 秒重试窗口内最多 8 次；一次成功的模型响应重置计数/退避/窗口。`retry.provider.maxRetries`（settings.json）可覆盖续跑重试次数，0 禁用但限流仍封顶 10[[12]](#ref-12)。限流冷却只来自 provider 给出的时间（`Retry-After` 等），“OpenClaw does not invent rate-limit cooldowns”；非限流类冷却随 profile 近期错误数递增：30 秒 → 1 分钟 → 5 分钟（上限），计数在失败窗口过后重置，状态存于每 agent SQLite 的 `usageStats`[[11]](#ref-11)。

故障类别映射：5xx 默认 `server_error`，HTTP 529 `overloaded`，408/504/522/524/499 为 `timeout`；格式/无效请求错误通常终止而不轮换。推进 fallback 的错误包括：认证失败、限流与冷却耗尽、过载、超时形、计费 disable、`model_not_found`、以及在活动模型压缩与超大工具结果恢复都无法适配之后的上下文溢出（且该 turn 须可安全重放）[[11]](#ref-11)。计费/额度类失败会标记凭据 disabled（初始 10 分钟）并轮换到下一个可用 profile/provider[[11]](#ref-11)。

一个重要的边界：本地转录写冲突与会话状态变化**不**触发模型重试/换账号/模型 fallback（先重读当前修订并重试一次，重复冲突作为本地错误报告）；`model_fallback_chain_stopped` 是终止性停止[[11]](#ref-11)。另一个易踩的配置：显式用户选择的模型是“严格”的，失败即报错，不会被无关 fallback 掩盖[[11]](#ref-11)。

### 19.4.4 超时、取消与停滞

超时表要分清几个不同的“超时”。`agent.wait` 默认 30s，只等待、不停止 run、不标示阶段；`agents.defaults.timeoutSeconds` 默认 172800s（48h）是每次模型 attempt 的执行预算，每个 fallback 获得新预算故总时长可超过该值；模型 idle timeout 云端 120s / 自托管 300s，模型进度死线是 idle 窗口的 2 倍[[7]](#ref-7)。

取消语义：模型 attempt 超时后，配置的 fallback 在被打断 attempt 落定后以新预算开始；用户取消停止整个 run（含 fallback）；审批等待暂停未用完的 attempt 预算；`ask_user` 提问不暂停预算；压缩可获得一次有界宽限；取消立即撤销该 run 的工具与审批权限[[7]](#ref-7)。终态超时是失败的 turn，不是成功完成；已完成的工具结果保留在转录中[[7]](#ref-7)。

停滞检测：内置两分钟阈值把长时间 `processing` 且无进展的会话分类为 `session.long_running/stalled/stuck`；“controller watchdog”即使诊断关闭也运行，语义停滞阈值至少 5 分钟且至少为告警阈值的 3 倍[[7]](#ref-7)。恢复性继续：当恢复中止一个尚未回复、请求已保存在转录里的交互式 turn 时，尝试一次 continuation（指示不要重复已完成动作），而不是立即要求用户重试[[4]](#ref-4)。

### 19.4.5 压缩失败与上下文溢出

超时的精确数字（本提交文档）：`agent.wait` 默认 30s；`agents.defaults.timeoutSeconds` 默认 172800s（48h），是每次模型 attempt 的执行预算，每个 fallback 获得新预算故总时长可超过该值，进度不重置预算；模型 idle timeout 云端 120s / 自托管 300s，模型进度死线是 idle 窗口的 2 倍[[7]](#ref-7)。限流冷却只来自 provider 给出的时间；非限流类冷却随 profile 近期错误数递增（30s → 1min → 5min 上限），计数在失败窗口过后重置，状态存于每 agent SQLite 的 `usageStats`[[11]](#ref-11)。认证 profile 轮换顺序是：存储的 per-agent 顺序 > `auth.order[provider]` > `auth.profiles` > 存储的 profiles；自动选中的 profile 按会话粘住以保持 provider 缓存温热[[11]](#ref-11)。

自动压缩在接近上限或模型返回溢出错误时触发并重试；溢出恢复无法适配时，失败回复建议 `/reset`、`/new` 或更大上下文模型；恢复压缩本身失败（摘要模型出错/超时）则回复说明并建议重试或 `/compact`[[10]](#ref-10)。取消不是回滚：运行被停止或超时也会停止其溢出/超时恢复；已完成的压缩保留在转录并计数，不发送迟到回复[[10]](#ref-10)。

摘要无时间提交时“无摘要地提交压缩”（见 19.3.6）；若固定指令与工具本身就超过模型窗口，或无更多历史可压缩而请求仍过大，OpenClaw 报告估算的预算组成而非当作无害跳过，用 `/context detail` 找大头[[10]](#ref-10)。手动 `/compact`：turn 活动时给它最多 60 秒完成，之后取消并再等最多 15 秒；Telegram 上先确认命令再中断，避免五分钟采纳看门狗中止并重放[[10]](#ref-10)。

### 19.4.6 重启恢复：Recovery is always on

压缩前的“静默记忆冲刷（memory flush）”把持久上下文写入记忆，是可选维护，失败不重置会话或丢历史；`agents.defaults.compaction.memoryFlush.model` 可指定本地模型[[10]](#ref-10)。活动转录字节守卫（opt-in，`maxActiveTranscriptBytes`）超限时用“有界近期历史视图”继续，省略最旧的整轮（工具调用与结果成对保留），不重写已存消息[[10]](#ref-10)。provider 侧压缩受“上下文引擎压缩所有权”约束，保存的 checkpoint 上限 16 MiB，超限或不兼容时回退客户端压缩[[10]](#ref-10)。

这是现稿遗漏、但最值得补的机制。OpenClaw 的重启恢复**始终开启**：每次 Gateway 进程关闭（含 SIGTERM、systemd/launchd 停止、`openclaw gateway stop`）都保留可恢复的未完成 turn；崩溃、强杀、漏写关机标记后同样进入恢复[[22]](#ref-22)。

恢复的对象与边界（“重启后自动恢复/不恢复”的清单）[[22]](#ref-22)：

- **会恢复**：被重启打断的已准入 turn（含仍在队列或准备执行的）自动继续既有会话；重放保留已记录的工具调用与结果、复用原用户消息；已接受的 `chat.send` 输入在启动时按接受顺序从 SQLite 恢复（过去 24 小时内接受的输入，最多 4 个会话并行重放）；子 agent 由 registry 恢复；cron 会话由调度器恢复；ACP 管理的会话由 IDE/客户端恢复。
- **不会恢复**：从未被准入的工作；Gateway 终端 PTY；后台 exec 进程结束（所属会话在合格时被唤醒一次并告知结果未知）；incognito 输入不能跨进程重启存活。

恢复有预算：main 会话“每个中断周期 3 次计费的自动派发”，三次未能启动后端 turn 耗尽恢复预算，之后 tombstone 直至检查并替换[[22]](#ref-22)。恢复的权限按当前 profile/role/grant/device 策略重新校验原调用者的记录权限；撤销仍停止恢复的 run；无法重建委托发送者权限则终止该 turn、记录一条本地中断通知与终止性 source receipt，防止重放；恢复耗尽后任务与转录保留，WebChat 的 **Resume** 以当前操作者权限继续[[8]](#ref-8)。

文档反复强调的一个口径：“A recovered count at startup means execution resumed. It does not prove that an assistant reply was delivered.”（启动时恢复计数只证明执行续跑，不证明回复已投递）[[22]](#ref-22)。这与“不确定的副作用须先检查再重复”是同一个原则。

优雅重启先排空：请求的重启（`openclaw gateway restart`、需要重启的配置变更、网关更新）不会立刻杀死在途工作；Gateway 停止接纳新工作，等待活动 agent turn 与后台任务完成，上限为排空预算（默认 5 分钟）；排空期间到达的消息被明确的重启错误拒绝（不静默排入将死进程）[[22]](#ref-22)。

### 19.4.7 插件失败隔离与崩溃循环

插件加载失败的隔离：插件注册失败的贡献被回滚，失败记录留在插件诊断中，不进入执行范围，也不阻塞健康插件与核心 context-engine 准入；插件替换/重载失败时保留 last-good 运行时，待工作结束后自动重试一次[[17]](#ref-17)。保留工作与在途调用共享 60 秒的 pre-stop 预算[[17]](#ref-17)。

崩溃循环断路器：3 次非干净启动在 5 分钟内触发；触发后下次启动抑制自启动副服务（渠道插件等保持关闭、控制面仍启动），安全模式同时暂停 main 会话重启恢复（不扣尝试次数、不落定、不发通知、不 tombstone），窗口排空后同进程自动重新武装恢复并只恢复一次[[22]](#ref-22)。

### 19.4.8 投递失败与持久出站

出站投递有持久队列：Gateway 启动时检查一次未完成的出站投递；新的 durable enqueue 在事务提交后唤醒恢复；有独立的投递尝试预算；生产者租约 60 秒、每 20 秒续约；租约过期不抹去“发送已开始”的证据，此类条目重放前仍需调和[[12]](#ref-12)。重试原则：每个 HTTP 请求单独重试而非整条多步流，只重试当前步骤以保序，避免重复可能造成重复副作用的操作[[12]](#ref-12)。

渠道层对“发送后不明确”的处理是：Telegram 新文本消息发送只有在 429 flood control 或传输证明请求未开始时才重试；发送后断连/超时/丢响应属于“不明确”，**不重放**（避免重复消息）；可安全重复的操作（如预览编辑）可重试瞬时网络与 5xx[[12]](#ref-12)。渠道发送重试默认 3 次、10% 抖动；Discord Gateway WS 重连另有循环：最多 50 次，2000ms 指数退避到 30000ms，无抖动[[12]](#ref-12)。

机器人对机器人循环保护是入站回复 runner 强制执行的滑动窗口速率守卫：默认 `enabled:true`、`maxEventsPerWindow:20`、`windowSeconds:60`、`cooldownSeconds:60`；A→B 与 B→A 计为同一对；不影响人类消息[[32]](#ref-32)。

### 19.4.9 基础设施故障与启动期失败

除渠道与模型外，基础设施层的失败有各自的语义。子运行时（child runtime）失效：Node 被升级/移除后，Gateway 进程活着但无法拉起子 worker；确认后 `childRuntime.available:false`，`/readyz` 503 且 `failing:["child-runtime"]`；服务托管的 Gateway 请求一次优雅重启，监管不可验证或五分钟内已因此关停则 report-only[[24]](#ref-24)。共享状态完整性失败：终止性准入失败立即使 `/ready`、`/readyz` 503，`failing:["state-database"]`[[24]](#ref-24)。插件替换/恢复：`/readyz` 503 `failing:["plugin-reload"]`，带 `phase: reloading|recovering|failed` 与 deadline；替换期间健康监视器不消耗渠道重启次数[[24]](#ref-24)。

Gateway 状态锁与端口冲突：状态所有权等待最多 5 分钟（每 250ms 检查）；`EADDRINUSE` 重试 20 次×500ms；服务监管下新进程先查现有进程 `/healthz`，健康则让位（systemd 下以退出码 78 退出）[[23]](#ref-23)。启动时配置冲突被分类为“需运维介入，非瞬时错误”`GatewayEffectiveConfigConflictError`，监管者的重启无济于事[[40]](#ref-40)。上下文准备超时：模型上下文转录读取有三分钟本地 worker 死线；本地死线会停止模型 fallback，且不会把 provider 凭据标为不健康[[3]](#ref-3)。worker 池过载：每池默认 128 待处理任务/256 MiB，超限 `WorkerTaskError.code="overloaded"`[[3]](#ref-3)。引擎清理死线：前台引擎 dispose 共享 agent 清理死线，默认 10 秒[[6]](#ref-6)。

诊断默认启用（`diagnostics.enabled:false` 关闭）；`openclaw gateway stability` 输出有界稳定性记录；致命退出、关停超时、重启启动失败在 `logs/stability/` 落盘；`openclaw gateway diagnostics export` 生成脱敏 zip[[24]](#ref-24)。

### 19.4.10 未找到与待核验

- 官方对 Agent Loop 最大迭代数/最大工具调用步数的设定：未找到（只有工具循环检测阈值与广播组 `maxTurns`）。
- `agents.defaults.runRetries` 在当前 docs 的说明：未找到（仅 CHANGELOG 记载）。
- 官方对队列等待时延、吞吐、并发调优的实测基准或推荐数值：未找到，调优部分只能写“旋钮与观测手段”。
- 重启恢复文档的中段（宿主休眠/冻结、失败更新恢复、自动续跑细节、子 agent 恢复）本次未读，正文未据此展开。
- 官方“个人助手应默认要求高风险动作确认”的表述：未找到；官方默认 `exec.security=full, ask=off`（见 19.5.4），“默认确认”是本书建议，不是产品默认。
## 19.5 实践：用好 OpenClaw

这部分每条做法标注“官方建议”或“本书建议”。OpenClaw 的官方建议集中在安全加固上，这符合它的定位：一个长期在线、能替人行动的系统，安全边界比功能更重要。

### 19.5.1 先理解信任模型

官方安全页把范围划得很清楚：每个 Gateway 一个信任边界（单个运营者或互信团队），OpenClaw“不是针对相互敌对用户的安全多租户边界”；混合信任或对抗用户必须拆分信任边界——独立 Gateway 与凭据，最好独立 OS 用户或主机（官方建议）[[25]](#ref-25)。这条决定了下文所有做法的框架：OpenClaw 的安全是“把不可信的人挡在边界外”，不是“在边界内做细粒度多租户隔离”。

### 19.5.2 暴露面：从 loopback 到公网的阶梯

官方暴露面板给的是一个从窄到宽的阶梯（官方建议）：Loopback + SSH 隧道（个人/运维）→ Loopback + Tailscale Serve → Tailnet/LAN 绑定（认证 + 防火墙白名单 + 不公网转发）→ 受信反向代理（`trusted-proxy` 认证、严格 `trustedProxies`）→ 公网（罕见高风险，身份感知代理、TLS、速率限制、严格白名单、沙箱化非 main 会话）[[28]](#ref-28)。核心原则是“在你能解释谁能到达它、如何认证之前，不要暴露 Gateway”；避免直接公网端口转发[[28]](#ref-28)。

与它配套的是保守默认：常规主机安装绑定 loopback；多数渠道对未知 DM 发送者返回配对码而不处理消息；群访问为白名单加 mention 门控。例外是容器镜像默认暴露绑定（需配合认证）[[25]](#ref-25)。网络页把底线说得更直白：永远不要在 `0.0.0.0` 上无认证暴露；Docker 发布端口会绕过主机 INPUT 规则，须在 `DOCKER-USER` 链实施[[28]](#ref-28)。Gateway 认证默认必需：无有效认证路径时拒绝 WebSocket 连接；明文 `ws://` 只接受 loopback、私有 IP 字面量、`.local`、Tailnet[[28]](#ref-28)。

### 19.5.3 DM 策略、配对与命令授权

身份优先于模型，是 OpenClaw 反复强调的顺序。DM 访问策略 `dmPolicy` 有四个值：`pairing`（默认，未知发送者得配对码，码 1 小时过期，Bot 忽略直到批准）、`allowlist`、`open`（须渠道白名单含 `"*"`）、`disabled`；批准只授予 DM 访问，不授予群访问（官方建议“DM 用 pairing + allowlist”）[[32]](#ref-32)[[25]](#ref-25)。群按成员关系而非渠道类型决策：陌生人可加入/发言的房间保留发送者白名单或 mention 门控；回复机器人消息（隐式 mention）不绕过 `groupAllowFrom`[[25]](#ref-25)。

多人能私聊你的 agent 时，官方建议开启 DM 隔离：`session.dmScope:"per-channel-peer"`（推荐）；这是“消息上下文边界，不是主机管理员边界”[[25]](#ref-25)。命令授权：斜杠命令与指令仅对授权发送者生效；渠道白名单为空或含 `"*"` 则命令实际对该渠道开放，这是需要显式关掉的默认[[25]](#ref-25)。

### 19.5.4 工具策略、exec 审批与沙箱

三个默认值要如实说明，因为它们与“更安全”的直觉不完全一致。

第一，工具策略默认 `tools.profile:"full"`：full 选择工具，但不授予 Full Access 执行权限；`minimal` 只有 `session_status` 和仅 update 的 `gateway`；`coding`/`messaging` 是中间档[[14]](#ref-14)。第二，exec 审批默认 `security:"full"`、`ask:"off"`——即受信运营者模型下，网关/节点主机的 exec 默认放行、不询问；“高风险动作默认确认”是本书建议，不是产品默认；审批只能收紧不能放松，需要提示但无 UI 或提示超时时按 `askFallback`（默认 `deny`）解析[[18]](#ref-18)。第三，沙箱默认关闭：`agents.defaults.sandbox.mode` 默认 `off`，`workspaceAccess` 默认 `none`；启用时 `non-main` 沙箱化除 main 外的所有会话，backend 默认 Docker，`sandbox.docker.network` 默认 `"none"`[[24]](#ref-24)。

官方对“高风险动作怎么管”给的是收紧路径，而不是默认：对处理不可信内容的 agent 默认 `tools.deny:["gateway","cron","sessions_spawn","sessions_send"]`；exec 需要审批时用 `ask:"always"` 加 `security:"deny"` 或严格 allowlist；`tools.elevated` 是显式逃生口，保持 `allowFrom` 收窄且不给陌生人；强制创建者角色沙箱（`sandbox:"required"`）在后端不可用时阻断执行，不回退到宿主机[[26]](#ref-26)。官方还给了三个 access profile 的例子：个人 agent（全权限无沙箱）、家庭/工作 agent（沙箱 + 只读工具）、公开 agent（沙箱 + 无文件系统/shell 工具）[[26]](#ref-26)。

### 19.5.5 多 Agent 与多用户

提示注入的官方立场是：模型选择是“第一道、最便宜的一层”，但仍需硬性执行（工具策略、exec 审批、沙箱）；不可信内容（网页、邮件、附件、粘贴日志）本身是威胁面，不止发送者；对“工具启用/读不可信内容”的 agent 使用最新一代最佳档模型，不要用更弱/更小档；若必须用小模型，则只读工具、强沙箱、最小文件系统访问、严格白名单，并关闭 `web_search/web_fetch/browser`；更稳的做法是用只读或禁用工具的“读者 agent”先摘要不可信内容再交给主 agent[[30]](#ref-30)。

硬化基线给出了一份可复制的收紧配置：`gateway:{mode:"local", bind:"loopback", auth:{mode:"token", …}}`、`session.dmScope:"per-channel-peer"`、`tools:{profile:"messaging", deny:["group:automation","group:runtime","group:fs","sessions_spawn","sessions_send"], fs:{workspaceOnly:true}, exec:{security:"deny", ask:"always"}, elevated:{enabled:false}, sessions:{visibility:"agent"}, agentToAgent:{enabled:false}}`，加上 WhatsApp 的 `dmPolicy:"pairing"` 与群 `requireMention:true`；其中一条内置底线是：聊天驱动的 agent turn 中非 owner 发送者无论配置如何都不能使用 `cron` 或 `gateway` 工具[[31]](#ref-31)。只读模式的组合是 `workspaceAccess:"ro"`（或 `none`）+ 工具 allow/deny 阻断 write/edit/apply_patch/exec/process；`tools.exec.applyPatch.workspaceOnly:true`（默认）即使沙箱关闭也阻止 apply_patch 写工作区外；官方还提醒“避免把 home 目录这类宽根目录作为 agent/沙箱工作区”[[26]](#ref-26)。

一个与“更安全默认”直觉相反的默认要写明：跨 provider 的消息工具默认开放。`tools.message.crossContext.allowAcrossProviders` 与 `allowWithinProvider` 默认都是 `true`（升级采用该默认，显式 `false` 保留阻断）；这些控制只守护发送/回复/编辑/删除/建线程等变更动作，不限制读取与管理动作，也不替代文件系统/shell/沙箱限制[[26]](#ref-26)。

多 agent 的价值是隔离，不是“多模型”。每个 agent 有自己的 workspace、状态目录与会话存储；路由选出 agent 后由它决定 workspace 与 session store[[13]](#ref-13)。公开 agent 的官方示例配置是：沙箱 all + workspaceAccess none + 只给 `sessions_list/history/send/spawn`、`session_status` 与渠道工具，deny `apply_patch/browser/canvas/cron/edit/exec/gateway/...`，`tools.sessions.visibility:"tree"`[[26]](#ref-26)。

会话可见性与 agent-to-agent 消息是两件事：工具 profile 不收窄会话工具可见范围，会话可见性是 Gateway 范围，agent-to-agent 消息默认开启——需要配合 `tools.sessions.visibility` 与 `tools.agentToAgent` 配置[[26]](#ref-26)。共享部署的官方提醒是：如果 Slack 里人人都能私聊 bot，核心风险是“委托的工具权限”；公司共享 agent 仅在同一信任边界内可接受，专用机器/VM/容器、专用 OS 用户，且不要把该运行时登录到个人账号[[29]](#ref-29)。

### 19.5.6 备份、升级与事件响应

节点执行 `system.run` 要当作远程代码执行看待：已配对 macOS 节点可被 Gateway 调用 `system.run`，官方明说“this is remote code execution on that Mac”；配对建立身份与令牌，不是逐命令审批面；全局节点命令策略只匹配精确命令名、不匹配 shell 文本；要完全禁用远程执行，security 设 `deny` 并移除该节点的配对[[26]](#ref-26)。

插件与技能目录都视为受信代码：插件运行在 Gateway 进程内，只从可信来源安装、优先显式 `plugins.allow` 白名单、启用前审查配置、插件代码/元数据/发现根变更后重启；npm/git 插件安装仅在显式 install/update 流程跑包管理器依赖收敛；`security.installPolicy` 让运营者对技能/插件安装返回 allow/warn/block；优先固定精确版本[[26]](#ref-26)。技能目录视为受信代码并限制谁能修改，技能监视器在下一个 agent turn 更新快照[[26]](#ref-26)。

备份的官方硬规则：绝不拷贝活动 `.sqlite`/`-wal`/`-shm`/`-journal` 文件（“raw file copies of a live database can be torn or corrupt”），用 `openclaw backup create --verify`（在线备份 API）；备份含认证配置、渠道与 provider 凭据、会话历史——加密存放、限制目标访问[[34]](#ref-34)。

升级用 `openclaw update`（先校验新版本、再激活并验证）；升级前跑 preflight checklist（已验证的备份、相同服务账号/状态/配置路径、`--dry-run` 检查）[[34]](#ref-34)。降级的官方警告：降级软件包不会反转配置或数据库迁移；受支持的恢复是“用其对应的 OpenClaw 版本恢复升级前备份”；不要降低 schema 版本标记[[34]](#ref-34)。

事件响应 SOP 是 Contain → Rotate → Audit → Collect：Contain（停进程、收紧绑定与 DM 策略）、Rotate（Gateway auth、远程客户端密钥、provider/渠道凭据）、Audit（`openclaw logs`、审阅转录与近期配置变更、重跑 `security audit --deep`）[[29]](#ref-29)。

### 19.5.7 security audit 与 doctor 的实际用法

敏感信息的官方立场是：假设 `~/.openclaw` 下任何内容都可能含秘密/隐私（config、credentials、两个 SQLite、sandboxes 等）；权限 `openclaw.json` 600、`~/.openclaw` 700；全盘加密；共享主机用专用 OS 用户；日志可含工具摘要、错误与 URL，转录可含粘贴的秘密；`logging.redactPatterns` 自定义脱敏（不改变模型上下文或转录）；分享诊断优先用 `openclaw status --all`；修剪不需要长期保留的转录与日志[[29]](#ref-29)。为基于手机号的渠道（WhatsApp/Signal/Telegram），官方建议考虑为助手使用与个人号码分离的独立号码[[31]](#ref-31)。

`openclaw security audit` 是 OpenClaw 把“安全建议”落成可执行检查的机制。官方建议在任何配置变更之后、或暴露网络面之前跑一次：`--deep` 尝试实时 Gateway 检查与插件收集器，`--fix` 应用安全修复，`--json` 输出结构化结果[[27]](#ref-27)。`--fix` 有意收窄：把 open 的群策略翻为 allowlist、收紧 state/config/include 文件权限（文件 600、目录 700，Windows 用 ACL）[[27]](#ref-27)。审计的分诊优先级是：①任何“open”且启用工具先锁 DM/群；②公网暴露立即修；③浏览器控制远程暴露当作 operator 访问；④权限不得 group/world 可读；⑤插件只装明确信任的[[27]](#ref-27)。

`openclaw doctor` 是启动前后的检查与修复：`--json` 只读、`--fix`/`--repair` 修复、`--lint` 做 CI 门；文档提醒 `--non-interactive` 只抑制提示、不抑制写入；显式修复会停止匹配的托管 Gateway、检查所有权、做只读 schema 快照后再修复，之后重启并验证就绪[[33]](#ref-33)。

### 19.5.8 常见问题诊断表

| 现象 | 常见原因 | 对策 | 依据 |
| --- | --- | --- | --- |
| `/readyz` 一直 503 `chat-recovery` | 启动输入恢复还没释放 fresh-chat 准入 | 等恢复完成，或查恢复日志 | 19.2.2 |
| 配置改了不生效 | 无效热重载被跳过；或需要重启的变更 | 看 `appliedConfigHash`；用 `config schema` 核字段 | 19.2.4 |
| 渠道反复重启后 “giving up” | 重启阶梯耗尽（10 次） | 修渠道根因；terminal disconnect 需手动重登 | 19.4.2 |
| 模型报错不换模型 | 显式用户选择的模型是严格的 | 用默认 primary + fallbacks，而非严格选择 | 19.4.3 |
| 重启后回复了重复内容 | 重放保留原消息与工具结果 | 属于恢复语义；不确定副作用先查 | 19.4.6 |
| 工具策略 deny 没挡住 | deny 只在同层；group/sandbox 层可能再放 | 按 19.3.7 的多层顺序逐层查 | 19.3.7 |
| exec 直接执行了命令 | 默认 `security:full, ask:off` | 收紧 exec 审批或沙箱 | 19.5.4 |
| 无人私聊的陌生人也能命令 | 渠道白名单空或含 `"*"` | 收紧 `commands.allowFrom` / 白名单 | 19.5.3 |

### 19.5.9 案例：给一个工作群挂一个只读助手

下面是一个教学构造的案例，未实测，只使用本章已核对过的入口与配置。目标：在 Slack 工作群挂一个助手，只读文档、不执行命令、不改文件、不碰配置。

| 步骤 | 做法 | 依据 |
| --- | --- | --- |
| 信任边界 | 助手跑在专用 OS 用户与独立 Gateway 实例上，不登录个人账号 | 19.5.1 |
| 暴露面 | Gateway 绑定 loopback，需要远程运维时走 Tailscale Serve，不做公网端口转发 | 19.5.2 |
| 群访问 | `groupPolicy` 保留发送者白名单 + mention 门控；回复不绕过白名单 | 19.5.3 |
| 工具策略 | `tools.profile:"minimal"` 起步；`tools.deny` 加 `group:fs`、`group:runtime`、`group:automation`、`sessions_spawn`、`sessions_send` | 19.5.4 |
| exec 与沙箱 | `exec:{security:"deny", ask:"always"}`；`sandbox:{mode:"non-main", workspaceAccess:"none"}`，非 main 会话进沙箱 | 19.5.4 |
| 审计 | 配置后跑 `openclaw security audit --deep`，再跑 `openclaw doctor`，按 critical/warn 逐项处理 | 19.5.7 |

这个案例里没有一步是“让模型更聪明”。每一步都在把 OpenClaw 的默认从“受信运营者的全权限个人助手”收窄到“只读工作群助手”：信任边界决定谁在系统内，暴露面决定谁能到达它，群白名单决定谁能触发它，工具策略与 exec 审批决定它能做什么，沙箱决定放行后的进程能碰什么，审计把这一切变成可复查的检查。

## 19.6 调优：控制上下文、成本与质量

OpenClaw 提供了完整的旋钮、默认值和观测手段，但**没有官方基准、没有本章的实测**。这一部分只写“旋钮与机制、作用对象、观测手段”，不写“调到多少好”；数字都是固定提交里的默认值。

### 19.6.1 度量与观测

用量与成本的观测入口是 `/status`（模型、上下文使用、最后响应 token、成本）、`/usage off|tokens|full|cost|reset`、`openclaw status --usage`（provider 配额窗口）；成本由 `models.providers.<provider>.models[].cost`（USD/百万 token，input/output/cacheRead/cacheWrite）估算，是估算不是账单[[20]](#ref-20)。上下文的观测用 `/context list`、`/context detail` 看每个注入文件/工具/技能/系统提示的体量[[20]](#ref-20)。

缓存可观测靠 `diagnostics.cacheTrace.enabled`（默认 false，落盘 `logs/cache-trace.jsonl`），观测项含 `input/cacheRead/cacheWrite` 与稳定前缀/易变后缀/工具指纹[[21]](#ref-21)。系统级观测还有 Prometheus 指标 `openclaw_session_recovery_total`、`openclaw_session_recovery_age_seconds`，以及 `OPENCLAW_GATEWAY_STARTUP_TRACE=1` 记录每个启动 span[[24]](#ref-24)。注意 cache trace 默认含提示与消息原文，属敏感数据[[21]](#ref-21)。嵌入式 runtime 在启动期做了“就绪才处理”的预热：`prewarmConfigDrivenReplyRuntime`、预加载 WS message handler、`sidecars.model-runtime` 发布模型快照，使首个连接与首轮复用已准备运行时[[40]](#ref-40)。

### 19.6.2 并发与队列旋钮

并发由三个键控制：`agents.defaults.maxConcurrent`（默认 `max(8, CPU×4)`）、`agents.defaults.subagents.maxConcurrent`（默认 8）、`tools.swarm.maxConcurrent`（默认 32）；每个会话仍串行[[4]](#ref-4)。worker 池：compute 池共享 CPU 准入 `max(1, availableParallelism()-1)`（给 Gateway 留一个 CPU），reader 池上限 2，compute 池上限 4，writer 与 singleton 串行[[3]](#ref-3)。队列的 cap（默认 20）与 drop（默认 `summarize`）只作用于排队消息，不改变运行中的 turn。

### 19.6.3 上下文预算

除 `/usage` 外，上下文拆解用 `/context list`、`/context detail` 看每个注入文件/工具/技能/系统提示的体量；`openclaw doctor` 对“bootstrap 文件超过上下文预算”给出“推荐改进”[[20]](#ref-20)。系统级日志与诊断还有 `openclaw logs --follow`、`openclaw status [--all|--deep|--usage]`、`openclaw health [--verbose|--json]`、`openclaw gateway stability`、`openclaw gateway diagnostics export`（脱敏 zip）；`diagnostics.enabled` 默认开[[24]](#ref-24)。

固定开销的旋钮集中在注入量上：`agents.defaults.bootstrapMaxChars`（20000）、`bootstrapTotalMaxChars`（60000）、`skills.limits.maxSkillsPromptChars`、`agents.defaults.imageMaxDimensionPx`（默认 1200）[[20]](#ref-20)。官方给的降 token 清单是：用 `/compact`；裁剪大工具输出；降低图片尺寸；保持技能描述简短（技能列表注入提示）；探索性冗长工作用更小模型[[20]](#ref-20)。

工具结果大小上限由有效上下文窗口推导另一个降工具声明占用的旋钮是 `tools.codeMode`：启用时普通工具移到“guest catalog bridge”之后，模型通常只见 `exec` 与 `wait`；`executor` 默认 `node`（`node:vm`，官方注明是“受信主机执行，不是安全隔离”），`quickjs` 是加固沙箱；MCP 工具目录按服务器名、工具名排序以稳定提示前缀[[14]](#ref-14)。

工具结果大小上限由有效上下文窗口推导：<100K token 为 16000 字符，≥100K 为 32000，≥200K 为 64000；运行时 context-share 守卫把单个工具结果限制在窗口的 30%[[20]](#ref-20)。会话修剪（`agents.defaults.contextPruning`）默认 `mode:"cache-ttl"`、TTL 5 分钟（Anthropic 家族下插件播种 1h），客户端修剪仅在 TTL 过期且占用 ≥约 30% 时开始；最近三个 assistant turn 与首条用户消息之前永不修剪[[10]](#ref-10)[[23]](#ref-23)。

### 19.6.4 模型与 fallback 调优

fallback 的调优要点（19.4.3 的机制）是：显式用户模型选择严格不回退；agent primary 默认严格、需自带 `fallbacks`；fallback 只对本 turn 生效；上下文溢出恢复耗尽后，已知上下文足够的候选优先[[11]](#ref-11)。认证 profile 按会话粘住以保持缓存温热，只在 `/new`/`/reset` 或冷却/禁用时轮换[[11]](#ref-11)。超时旋钮：`agents.defaults.timeoutSeconds`（172800s）、`models.providers.<id>.timeoutSeconds`（慢速本地/自托管 provider 应先调它再调 agent 超时，且保持 agent 超时不低于它）[[7]](#ref-7)。

### 19.6.5 缓存

`cacheRetention` 取 `none|short|long`持久会话中的普通 agent-turn 定时作业可以在缓存窗口内“保温”：例如一小时缓存配 `openclaw automations edit <job-id> --every 55m`；隔离作业每次新会话不保温；官方提醒“每个 turn 都花 token，比较总用量而不是假设命中”[[21]](#ref-21)。

`cacheRetention` 取 `none|short|long`；**换模型必然开启新缓存谱系**，改 thinking/推理级别也可能失效——官方建议“创建会话时选定模型与 thinking 级别并保持稳定”[[21]](#ref-21)。持久会话中的普通 agent-turn 定时作业可在缓存窗口内“保温”（例：一小时缓存配 `automations edit <job-id> --every 55m`）；隔离作业每次新会话不保温；官方提醒“每个 turn 都花 token，比较总用量而不是假设命中”[[21]](#ref-21)。

### 19.6.6 压缩与存储旋钮

压缩调优键：`agents.defaults.compaction.model` 可用本地小模型专做摘要；`keepRecentTokens` 默认 20000；`maxActiveTranscriptBytes`（opt-in 字节守卫）；`memoryFlush.model`；`notifyUser`[[10]](#ref-10)。会话存储维护影响磁盘占用：`session.maintenance.{mode:"enforce"|"warn", pruneAfter:"30d", archiveDashboardAfter:"7d", maxEntries:5000}`，`openclaw sessions cleanup --enforce` 手动执行[[8]](#ref-8)。

### 19.6.7 不应为省 Token 放弃的东西

有几类开销不能省（本书建议）。工具策略的逐层求值：省掉任何一层的 deny 都意味着越权路径变宽，调优只能改“哪些工具可选”，不能改“deny 是否生效”。exec 审批：降低打断要靠 allowlist 与沙箱，不能靠把 `ask:"always"` 改成 `ask:"off"`。重启恢复：关闭恢复或调低预算的代价是“已接受的输入可能永远丢失”，这不是 token 问题，是正确性问题。缓存保温可以关（那省费用），但逐动作的权限边界和恢复语义不能为省 token 而放宽。

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
39. <a id="ref-39"></a> OpenClaw maintainers，CHANGELOG：[2026.2.21](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.2.21.md)、[2026.3.13](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.3.13.md)、[2026.3.22](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.3.22.md)、[2026.4.20](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.4.20.md)、[2026.4.22](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.4.22.md)、[2026.5.12](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.5.12.md)、[2026.5.14](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.5.14.md)、[2026.5.22](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.5.22.md)。
40. <a id="ref-40"></a> OpenClaw maintainers，启动源码：[server.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server.ts)、[server-start.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server-start.ts)、[server-startup-finish.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server-startup-finish.ts)、[server-startup-post-attach.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server-startup-post-attach.ts)、[server/readiness.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server/readiness.ts)、[server-runtime-config.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server-runtime-config.ts)、[pre-bootstrap.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/cli/gateway-cli/pre-bootstrap.ts)。
41. <a id="ref-41"></a> OpenClaw maintainers，渠道源码：[server-channels.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/server-channels.ts)、[channel-health-policy.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/channel-health-policy.ts)、[channel-health-monitor.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/channel-health-monitor.ts)。
42. <a id="ref-42"></a> OpenClaw maintainers，[resolve-route.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/routing/resolve-route.ts)。
43. <a id="ref-43"></a> OpenClaw maintainers，[tool-policy-pipeline.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/tool-policy-pipeline.ts)、[conversation-tool-policy-pipeline.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/conversation-tool-policy-pipeline.ts)、[tool-policy-match.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/tool-policy-match.ts)。
44. <a id="ref-44"></a> OpenClaw maintainers，[agent-core agent-loop.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/packages/agent-core/src/agent-loop.ts)、[tool-loop-detection.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/tool-loop-detection.ts)。
45. <a id="ref-45"></a> OpenClaw maintainers，[context-engine/types.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/context-engine/types.ts)、[legacy.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/context-engine/legacy.ts)、[bootstrap.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/embedded-agent-helpers/bootstrap.ts)、[bootstrap-limits.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/embedded-agent-helpers/bootstrap-limits.ts)、[inbound-dedupe.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/auto-reply/reply/inbound-dedupe.ts)。
46. <a id="ref-46"></a> OpenClaw maintainers，[security-cli.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/cli/security-cli.ts)、[fix.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/security/fix.ts)、[net.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/gateway/net.ts)。
