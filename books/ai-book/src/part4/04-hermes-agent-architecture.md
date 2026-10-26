# 第 22 章 Hermes Agent 系统：长期运行 Agent Runtime 的实现、使用与调优

> Hermes Agent 的核心不是“多一个聊天入口”，而是把一个会积累记忆、沉淀技能、跨入口工作、可扩展工具、并生成训练轨迹的个人 Agent，做成一个长期运行的运行时：Gateway 拥有会话的权威状态，AIAgent 负责一次回合的模型与工具循环，记忆、技能和会话搜索在两者之间形成一条学习闭环。也正因为它是长期在线、能替人行动的系统，安全边界比功能更重要。

**资料核验日期：2026-10-23。** 本章分析 NousResearch/hermes-agent 仓库的固定提交 `907e3224c432647567cc4dd7aebcdf258971896a`（提交时间 2026-10-23 01:48 UTC）。源码与文档都固定到这一提交。本章的行为结论来自阅读源码与文档，没有运行验证、没有故障注入。

| 标记 | 含义 | 使用方式 |
| --- | --- | --- |
| 源码事实 | 固定提交中可定位的文件与符号、行号 | 只引用 `907e3224…`，不引用主分支 |
| 官方文档事实 | 同一提交里 website/docs 的陈述 | 注明核验日期；滚动文档，版本敏感 |
| 系统推导 | 根据公开证据做的架构解释 | 不等于产品内部实现 |
| 本书建议 | 本书依据事实归纳的使用建议 | 与官方建议分开标注 |

两条诚实性限制需要提前说明。第一，Hermes 文档与源码在同一提交内就有多处冲突（`agent.max_turns` 文档写 500、源码默认无限；schema 版本文档写 31、源码是 32；中断是否注入部分响应的说法相反），正文遇到会点名，分别说明源码实现与文档承诺。第二，README 里“only agent with a built-in learning loop”这类厂商自述，只作为厂商说法引用，不作为事实结论。

| 本章的判断问题 | Hermes 的差异 | 通用解释位置 |
| --- | --- | --- |
| 输入到底被谁接受 | session authority 与 durable admissions | [Runtime 受理与状态](../part3/04-agent-runtime.md) |
| 崩溃时动作是否发生 | unknown admission、effect_disposition | [外部副作用](../part3/04-agent-runtime.md#failure-side-effects) |
| 生成结果能否送达 | delivery_obligations 与重复提示 | [结果交付](../part3/04-agent-runtime.md#result-delivery) |
| 学习结果是否可信 | memory、skills 与后台反思写入 | [记忆生命周期](../part3/02-context-engineering.md#memory) |


## 22.1 组成：Hermes 由哪些部分构成

### 22.1.1 定位：一个长期运行、自我改进的 Agent Runtime

Hermes 的官方自述是“a self-improving AI agent”，README 强调它“built-in learning loop”（内置学习闭环），并把 Telegram、Discord、Slack、WhatsApp、Signal 和 CLI 全部收进“a single gateway process”[[1]](#ref-1)。这句“单一 gateway 进程承载多入口”是理解它的钥匙：与把入口当旁路的系统不同，Hermes 的 Gateway 是会话的权威写者，所有入口都向它提交输入（见 22.3.1）。

“长期运行”需要精确表述。Gateway 有两种进程形态：常驻服务（systemd/launchd/容器），和 CLI 按需拉起、空闲后自退的 unmanaged 守护——后者带 `--idle-exit`，无附着客户端、无适配器、无可运行 cron/kanban、无排队 admission 时，空闲 `gateway.unmanaged_idle_exit_seconds`（默认 600 秒）后自行退出[[23]](#ref-23)。所以“长期”靠的是持久化状态（state.db、记忆、技能），不是进程寿命[[23]](#ref-23)。

安全定位上，Hermes 的 SECURITY.md 说得非常直白：**“唯一的安全边界是操作系统”**——审批门、输出脱敏、模式扫描器、工具 allowlist 都不是 containment[[30]](#ref-30)。这一条是 22.5 整个实践部分的框架。

### 22.1.2 代码组织：facade 与 siblings

Hermes 的代码组织有一个显式的约定：每个曾经的 god file 是**门面**（公共入口 + 被其他包 import 的名字），`<stem>_<topic>.py` 兄弟文件各管一个主题；项目开发指南要求“按主题找代码，不要读门面”，文件不超过 2000 行[[1]](#ref-1)。所以 `run_agent.py`（约 1500 行）、`gateway/run.py`（约 5200 行）、`hermes_state.py`（约 1700 行）只是门面，机制在 `agent/turn_*.py`（约 30 个阶段文件）、`gateway/run_*.py`、`hermes_state_*.py`（60 个文件）里。读 Hermes 源码时，这是一个必须遵守的地图[[1]](#ref-1)。

顶层目录含 `agent/`（回合循环）、`gateway/`（多入口控制面）、`hermes_cli/`（CLI）、`tools/`（工具与后端）、`cron/`（定时）、`plugins/`（平台适配器与模型 provider）、`skills/`（内置技能）、`optional-skills/`（官方可选）、`plugin-catalog/`（目录条目随版本变化）等[[3]](#ref-3)。

### 22.1.3 核心对象

| 对象 | 位置 | 职责 |
| --- | --- | --- |
| `AIAgent` | `run_agent.py` | 门面；`__init__` 转交 `init_agent()`，`run_conversation()` 由 `TurnFacadeMixin` 提供 |
| `GatewayAgent` | `hermes_cli/gateway_agent.py` | CLI 里代替 `AIAgent` 的“gateway 后端替身”，`run_conversation()` 经 `prompt.submit` 提交 |
| `HermesCLI` | `cli.py` | 17 个 mixin 组合的 CLI 门面 |
| `GatewayRunner` | `gateway/run.py` | gateway 门面，组合 `gateway/run_*.py` 兄弟 mixin |
| `SessionAuthorities` | `gateway/session_authorities.py` | 按 hermes_home 键控的 authority 注册表，多 profile 时每个 served profile 一个 |
| `TurnRunner` | `gateway/run_turn_runner.py` | 一个 gateway 回合的执行体，`run_sync()` → `_run_sync_scoped()` |
| `SessionStore` / `SessionSource` | `gateway/session.py` | 会话键与条目；`sessions.json` 镜像 + SQLite 规范存储 |
| `ToolRegistry` | `tools/registry.py` | 工具自注册表：`register()`/`get_definitions()`/`dispatch()` |
| `ContextCompressor` | `agent/context_compressor.py` | 默认上下文引擎实现 |
| `MemoryStore` | `tools/memory_tool_store.py` | 内置 Markdown 记忆 |
| `SessionDB` | `hermes_state.py` | SQLite 会话库门面 |

### 22.1.4 状态与存储

状态目录默认 `~/.hermes`（`HERMES_HOME` 覆盖）。关键布局：`config.yaml`（行为配置）、`.env`（只放密钥）、`auth.json`（OAuth 凭据）、`SOUL.md`、`memories/`（MEMORY.md、USER.md）、`skills/`、`cron/`、`sessions/`、`logs/`[[5]](#ref-5)。命名 profile 目录是 `<root>/profiles/<name>/`，各带独立 `state.db`、配置、日志[[5]](#ref-5)。一个例外是 Kanban 数据库：默认看板 DB 在共享根 `<root>/kanban.db`，是跨 profile 共享状态，与“profile 是岛”的默认隔离不同[[11]](#ref-11)。

会话的权威状态在 SQLite `state.db`（WAL）。表包括 `sessions`、`messages`、`session_model_usage`、三个 FTS 表（`messages_fts`/`_trigram`/`_cjk`）、`state_meta`、`gateway_routing`、`compression_locks`、`async_delegations`、`delivery_obligations`、`runtime_epoch`、`session_admissions`、`worker_executions`、`worker_receipts`、`schema_version` 等[[7]](#ref-7)。`SCHEMA_VERSION = 32`（文档写 31；源码常量见 22.6.8，不推定安装版本）[[7]](#ref-7)。Gateway 是 canonical session 的唯一写者：CLI、TUI、Desktop、API server、ACP、cron、Kanban worker 都是“向 session authority 提交输入”的客户端[[7]](#ref-7)。

除会话库外，还有一套运行时账本表支撑恢复：`session_admissions`（durable 输入 FIFO，`intent ∈ {queue, steer, redirect}`、`status ∈ {queued, started, unknown, terminal}`）、`worker_executions`（cron/child/compute/kanban）、`worker_receipts`、`runtime_epoch`（epoch + instance_id 单行）、`session_turn_leases`（跨进程回合租约）[[7]](#ref-7)。

### 22.1.5 进程与部署形态

可核对的进程形态有七种：CLI/`hermes chat`/TUI 客户端；gateway 守护进程（常驻服务或按需拉起）；进程外 managed worker（bypass/safe 会话、custom-provider、kanban/cron 等）；TUI（Node/Ink 前端 + Python 后端）；`hermes serve` 无头后端（端口默认 9119）；ACP stdio 服务；cron（gateway 内置 ticker 线程，或外部 `hermes cron tick`）[[23]](#ref-23)。console scripts 有三个：`hermes`、`hermes-agent`（legacy 单查询 runner）、`hermes-acp`[[6]](#ref-6)。

多 profile 复用进程：`gateway.multiplex_profiles` 默认未设，启动时由 `resolve_multiplex_mode` 裁决——无阻塞项则 True；显式 `false` 已退役为“拓扑 opt-out”（只告警，视同未设）；有阻塞项（s6 容器、仍自有 gateway 的次要 profile、重复 bot 凭据）时降级为单 profile[[9]](#ref-9)。多 profile 隔离靠两个 contextvar 作用域（`hermes_home` 与 secret scope）叠加，`get_secret()` 在 multiplex 开启而无作用域时抛 `UnscopedSecretError`（fail closed）[[9]](#ref-9)。

### 22.1.6 模型接入与执行后端

模型接入由共享解析器 `resolve_runtime_provider()` 按“梯子”逐级返回：禁用 provider 守卫 → 快捷名 → 命名自定义 provider → 本地端点 → 显式 `--api-key/--base-url` → 凭证池 → OAuth → 环境变量 → bedrock → OpenRouter/裸 custom 回退 → `model.openai_runtime` 覆盖[[8]](#ref-8)。执行后端（terminal backend）内置 6 个：`local`、`docker`、`ssh`、`modal`、`daytona`、`vercel_sandbox`（另有 `managed_modal`）；Singularity 已从内置后端变为官方插件，这是相对较早实现的变化[[22]](#ref-22)。工具集由 `toolsets.py` 的 `_HERMES_CORE_TOOLS` 与平台预设（`hermes-cli`、`hermes-telegram`、`hermes-acp` 等）定义[[2]](#ref-2)。

provider 梯子的完整顺序是（“梯子的顺序即行为——每级返回或抛出，否则落到下一级”）：禁用 provider 守卫 → `moa`/`anthropic@azure`/`azure-foundry`/`vertex` 快捷名 → 命名自定义 provider → 本地端点旁路 → `auth.resolve_provider`（显式 `--api-key/--base-url`）→ 凭证池 → OAuth（nous/codex/xai/qwen 等）→ 外部进程 → anthropic 环境变量 → bedrock → 注册表 api_key provider → OpenRouter/裸 custom 回退 → `model.openai_runtime` 覆盖（可改写 api_mode 为 `codex_app_server`）[[8]](#ref-8)。bundled 模型 provider 插件目录 `plugins/model-providers/` 承载各模型后端；每个 provider 的 key 只发往其自己的 base URL[[8]](#ref-8)。

### 22.1.7 记忆、技能与学习闭环

记忆分“内置 Markdown 记忆”与“外部 memory provider”两层：内置 provider 把条目写进 `MEMORY.md`（预算 2200 字符）与 `USER.md`（1375 字符）；外部 provider 至多一个，仓库自带 `byterover`、`retaindb`，其余（honcho、mem0 等）在 plugin-catalog 里由 `hermes plugins install` 安装[[12]](#ref-12)。技能遵循 SKILL.md + YAML frontmatter，渐进披露：技能索引（名称 + ≤60 字符描述）每次会话注入，正文按需读取[[4]](#ref-4)。

学习闭环（厂商自述）包括经验写入、后续召回和技能复用；持久 Memory、会话内 prompt 快照与供应商缓存则是三个不同状态层，不能以缓存命中证明知识已长期保存[[1]](#ref-1)。轨迹导出的边界是：`save_trajectories` 默认 False，**没有 CLI/config 开关**，只有 `python run_agent.py --save_trajectories` 或库参数；轨迹格式（ShareGPT 风格）、`trajectory_compressor.py`（默认 15250 目标 token）、`batch_runner.py` 存在，但 eval case 生成、policy refinement、RL 训练集成在本章核查的轨迹导出与批处理入口中未发现集成路径，不能据此否定外部研究系统——README 的“reinforcement learning datasets”只是文档声明，指向外部仓库[[15]](#ref-15)。

架构文档与源码在同提交内有三处不一致，读源码时要小心：文档数据流写 `HermesCLI.process_input() → AIAgent.run_conversation()`，而 CLI 实际用 `GatewayAgent`（见 22.2.2）；文档的 “Terminal (7 backends)” 与本章列出的 6 个内置后端冲突（图未更新）；文档的 “3 API Modes” 与源码 4 个注册 transport 不一致（`chat_completions`、`anthropic_messages`、`codex_responses`、`bedrock_converse`，另有第五个 api_mode `codex_app_server` 整回合委托子进程）[[1]](#ref-1)。平台适配器分布在 `plugins/platforms/` 下多个目录（a2a、discord、feishu、line、matrix、signal、slack、teams、telegram、wechat、whatsapp 等），`gateway/platforms/` 是共享 base 与直连适配器[[3]](#ref-3)。

state.db 的表列值得看一眼，因为它决定了能观测什么：`sessions` 表关键列有 `model、model_config、system_prompt、parent_session_id、started_at/ended_at、end_reason、message_count、tool_call_count、input/output/cache_read/cache_write/reasoning_tokens、estimated_cost_usd、actual_cost_usd、cost_status、pricing_version`，另有网关路由列（session_key、chat_id、chat_type、thread_id、origin_json）与工作区列（cwd、git_branch、git_repo_root）[[7]](#ref-7)。`messages` 表有 `content、tool_calls、tool_name、finish_reason、reasoning、effect_disposition、platform_message_id、compacted、api_content、message_uid` 等；`api_content` 是“发给 API 的精确字节”旁路列，用于缓存稳定重放，`message_uid` 是稳定消息 id[[7]](#ref-7)。并发写有明确预算：SQLite 超时 1 秒加应用层抖动重试，常规写 20 秒、transcript 写 60 秒、仅活动写 0.5 秒预算，`BEGIN IMMEDIATE`，每 50 次成功写做 PASSIVE WAL checkpoint，预算耗尽时回合以 `session_persistence_failed:locked` 结束[[7]](#ref-7)。

技能分三个层级：`skills/`（内置，一直可用）、`optional-skills/`（官方可选，默认不激活，`hermes skills browse/install`）、Skills Hub（社区/小众，安装前审查）；插件在 `plugin-catalog/`（单 YAML、固定 40 位提交 SHA，经人工合并）[[31]](#ref-31)。配置优先级官方写清：CLI 参数 > `config.yaml` > `.env` > 内置默认；`.env` 只放密钥，其余行为配置放 `config.yaml`——开发指南甚至拒收“非 secret 配置用新 `HERMES_*` 环境变量”的贡献[[31]](#ref-31)。





会话路由的镜像文件 `sessions.json` 是 `session_key → session_id` 的映射加条目元数据（原子写），SQLite 为 canonical、`sessions.json` 是镜像；镜像存在不构成 SQLite 故障时仍能按原权威契约执行的保证[[7]](#ref-7)。`get_hermes_home()` 的解析顺序是：context-local override → `HERMES_HOME` 环境变量 → 平台默认（macOS/Linux `~/.hermes`，Windows `%LOCALAPPDATA%/hermes`）[[5]](#ref-5)。

## 22.2 启动：从入口到首轮

Hermes 的启动要回答两层问题：一个入口进程怎样起来，以及一个 AIAgent 回合怎样从“装配”走到“可以发第一个模型请求”。它没有单一入口，所以“启动”要按入口分别看。

### 22.2.1 命令入口与 Profile 解析

`hermes` 控制台脚本指向 `hermes_cli.main:main`[[6]](#ref-6)。Profile 在其余 Hermes 模块导入**之前**解析：`_apply_profile_override()` 在 main.py 模块加载时执行，规则是显式 `--profile/-p` 优先，其次 `HERMES_HOME` 已指向 `profiles/<name>`，再读默认根下 `active_profile`；结果写入 `os.environ["HERMES_HOME"]`，并把 `--profile` 从 argv 剥掉。源码注释强调 profiles 模块内部异常绝不阻止启动（告警并用默认）[[10]](#ref-10)。

`main()` 在分派命令前做一串启动准备：设进程标题、`install_truststore()`（统一 TLS CA）、恢复被中断的更新欠账、非 update 命令时 `maybe_run_boot_bootstrap`、`--yolo` 在插件发现前写 `HERMES_YOLO_MODE=1`（因为 `tools.approval` 在 import 时冻结该值）、`_prepare_agent_startup`、无子命令则默认进 chat[[10]](#ref-10)。

无 provider 配置时的首启保护：`_has_any_provider_configured()` 为假且不是 TUI → `_first_run_setup_guard` 并返回；TUI 自己渲染 “Setup Required”[[10]](#ref-10)。插件/MCP/hook 发现是“按命令类型条件触发”的：对 `chat` 与 TUI，客户端**不做**插件/MCP 发现，因为“gateway 自己发现并拥有 MCP servers”——源码注释给出理由：避免第二份 MCP server 与跨进程发现锁（源码注释报告 2.0 秒（非本书实测），一台 stdio server）[[10]](#ref-10)。

### 22.2.2 经典 CLI 到 gateway 的接入

经典 CLI 的 `AIAgent` 其实是 `GatewayAgent`：`cli_agent_setup_mixin.py` 里 `from hermes_cli.gateway_agent import GatewayAgent as AIAgent  # the gateway runs the turn`[[3]](#ref-3)。会话存储取 `GatewaySessionStore`（“gateway is the single writer”），`_ensure_gateway_session()` 创建/恢复会话并**采用 gateway 返回的 session id**[[8]](#ref-8)。

连接并按需拉起 gateway：`GatewayConnection.start(timeout=90)`；未显式 `HERMES_TUI_GATEWAY_URL` 时调用 `ensure_gateway_runtime(home, idle_exit=True)`；显式远端 URL 失败**永不**本地拉起替代。无回退文案是 “Gateway connection failed; no local fallback”[[23]](#ref-23)。`ensure_gateway_runtime` 的语义是“只发现或请求一次，不安装服务”：观察到 `absent` 才启动，优先已安装服务，否则 `spawn_unmanaged_gateway`；源码注释强调 “A successful service command or Popen is not session readiness.”——READY 由发现协议判定，不是 Popen 成功[[23]](#ref-23)。

### 22.2.3 Gateway 冷启动到 READY

`gateway.run.start_gateway()` → `gateway.run_bootstrap.start_gateway()`，主体在 `GatewayRunner.start()`（`gateway/run_startup.py`）。真实顺序比早期实现的八阶段表多了两项：**PID 文件用 `O_EXCL` 抢在适配器之前**，**控制 socket 失败即中止启动**；MCP 发现发生在 `runner.start()` 之前；cron ticker 与 housekeeping 在 `runner.start()` 之后、READY 之前[[24]](#ref-24)。核心顺序是：①设 `HERMES_EXEC_ASK=1`、记录启动代码指纹；②加载 gateway 配置（“配置裁决先于重复实例守卫”）；③并行启动环境探针；④multiplex 裁决；⑤同 `HERMES_HOME` 重复实例守卫（`--replace` 需证明所有权）；⑥全有或全无预留 profile 运行所有权；⑦抢 PID 文件；⑧恢复离开核心的功能；⑨构造 `GatewayRunner` 并 `initialize_gateway_runtime`；⑩信号处理器；⑪控制 socket（失败即中止）；⑫MCP 发现（执行器线程，失败非致命）→ `runner.start()`；⑬cron ticker 与 housekeeping；⑭`publish_gateway_runtime_ready`[[24]](#ref-24)。

`initialize_gateway_runtime` 为每个 reserved profile 建 authority，生成 `instance_id`，启动 profile 失败则整体抛出、次要 profile 失败则 `_park_reserved_profile` 停放[[24]](#ref-24)。READY 的判定：`publish_gateway_runtime_ready` 只在 listener 未结束、`_running`、未 `_draining` 时置 `state='ready'` 与 `capabilities=['session-authority-v1','durable-admission-v1','event-replay-v1']`；文档强调 “READY only once adapters, cron and housekeeping run”，且**不保证**外部模型、MCP、沙箱已调用成功[[24]](#ref-24)。

启动门控是有界的：恢复期间入站事件排队，`_finish_startup_restore` 对启动恢复任务等待上限 30 秒（可配）、对预热等待上限 20 秒，超时只警告、继续放行入站，后台任务不取消；`finally` 保证闸门无论怎样都放开。源码 docstring 说得很清楚：“availability outranks prompt completeness for a WEDGED init”——可用性优先于提示完备性[[24]](#ref-24)。

### 22.2.4 平台连接失败的分级

平台连接结果分级处理：全部连上 → 正常；有“不可重试”失败但仍有已连平台 → `degraded`；零连接且只有不可重试错误 → `_startup_fail_fatal_config`（退出码 78）；致命与瞬时混合 → 保持存活走重试；零启用平台 → 仍运行（供 cron）[[24]](#ref-24)。重连看门狗指数退避 30 秒 → 300 秒封顶，可重试失败永远重试，持续失败超过 `reconnect_attention_after`（默认 7200 秒）后标 `needs_attention`[[24]](#ref-24)。启动存活看门狗（`hermes_startup_watchdog.py`）在导入 gateway 之前武装，事件循环 300 秒未就绪就转储线程栈、以服务重启码 75 退出；它用“阶段性进度租约 + CPU 进展”避免误杀慢启动（schema 迁移这类低 CPU 阶段靠进度租约证明存活）[[25]](#ref-25)。

### 22.2.5 AIAgent 装配

`AIAgent.__init__` → `agent.agent_init.init_agent()`。装配顺序（源码行号可核）：①safe worker 时强制跳过 context files/memory/后台复盘；②`iteration_budget = iteration_budget or IterationBudget(max_iterations)`；③`_resolve_api_mode` + `_finalize_routing`；④`_build_client` + `_init_fallback_chain`；⑤`_load_tools`（工具快照）；⑥`_init_session_state`；⑦`_init_memory`、`_apply_agent_section`（`api_max_retries` 默认 3、`auto_recovery_cycles` 默认 5）、`_parse_compression_config`、`_resolve_context_length`、`_build_context_engine`、`_enforce_minimum_context`、`_inject_context_engine_tools`、`_init_usage_state`；⑧`_snapshot_primary_runtime`[[3]](#ref-3)。

两个容易误读的点。第一，“子 Agent 可继承共享迭代预算”不准：`init_agent` 注释写 parent creates/children inherit，但 `delegate_task` 构造子 Agent 时传 `iteration_budget=None  # fresh budget per subagent`，即每个子 Agent 是**新预算**[[14]](#ref-14)。第二，工具快照先于记忆与上下文引擎的扩展工具注入，最终 schema 不一定只来自最初 registry 查询——`_inject_context_engine_tools` 在 `_load_tools` 之后[[3]](#ref-3)。

惰性/延迟的部分汇总：built-in 工具与插件发现发生在 `model_tools` 导入期，MCP 不在导入期；`croniter` 惰性导入；终端沙箱在首次需要时创建；预热失败退回首轮惰性初始化；`deferred_token_mint()`（CLI 不做第二次 Vertex JWT 交换）；浏览器 vault 工具默认 deferred；GatewayStartup dashboard 仅对有 supervisor 的 gateway 在 READY 后预热，client 自起 gateway 则首次请求时挂载[[24]](#ref-24)。

`init_agent` 装配的完整顺序（含一些容易遗漏的步骤）：safe worker 强制跳过上下文/记忆/后台复盘 → 绑定透传参数与网关身份参数、`IterationBudget(max_iterations)` → `_resolve_api_mode` + `_finalize_routing` → 默认控制状态、`_init_prompt_cache_config`、`_init_turn_state` → `_build_client` + `_init_fallback_chain` → `_load_tools` → `_init_session_state` → 读一次 `load_config_readonly()` → `_init_memory`、`_apply_agent_section`（`api_max_retries` 默认 3、`auto_recovery_cycles` 默认 5）、`_parse_compression_config`、`_resolve_context_length`、`_build_context_engine`、`_enforce_minimum_context`、`_inject_context_engine_tools`、`_init_usage_state` → `_snapshot_primary_runtime`[[3]](#ref-3)。工具快照 `_load_tools` 的顺序是：先 `recover_at_startup`（离开核心的功能）、`discover_plugins()`（multiplex 下必须在快照前做，因为 HERMES_HOME 可能已换）、`session_disabled_toolsets` → `get_tool_definitions(enabled, disabled)` → `prune_oneshot_tools`（有限的 `-q` 运行去掉技能写入工具）→ `agent_tool_drops`[[3]](#ref-3)。

### 22.2.6 首轮边界

首轮入口 `TurnFacadeMixin.run_conversation`：先取消后台 review，建立 task/session/relay/accounting 作用域，**申请持久回合租约**（`session_turn_leases`，`LEASE_TTL_SECONDS = 300`、等待上限 `LEASE_WAIT_SECONDS = 1800`、刷新间隔默认 60 秒），再进 `conversation_loop.run_conversation`；租约/准入失败会提前结束而不进入模型循环[[3]](#ref-3)。这个跨进程租约的意义在于：同一 state.db 上的多个进程（Desktop、CLI resume、gateway）可能并发，租约续租失败/到期时 `agent.interrupt(hard_cancel=True, tool_reason="session turn lease lost")` 保护 transcript[[16]](#ref-16)。

system prompt 的恢复或构建在 `_restore_or_build_system_prompt`：有历史且存在 DB 会话行 → 取 `sessions.system_prompt`；`_stored_prompt_matches_runtime`（比对 model/provider/运行环境身份）匹配则**原样复用**并恢复钉住的工具顺序；不匹配、为空或无历史 → `agent._build_system_prompt()` 并持久化；持久化失败只告警并说明“后续回合将重建并 miss 前缀缓存”[[17]](#ref-17)。

后端是惰性创建的：`ensure_task_env()` 在首次需要环境的能力调用时才创建本地/Docker/SSH 等后端，命中活跃环境则复用；local 返回 None；失败记 warning 返回 None。这条“首条终端命令才创建后端”是启动耗时与首轮时延的边界[[22]](#ref-22)。

## 22.3 执行：一条消息从进来到出去

`cmd_chat` 有三条子路径：`stream-json` 或“bypass”（safe-mode/ignore-user-config）→ 纯传输客户端 `launch_gateway_chat`；否则判定 TUI（`--tui/--cli` 标志、TTY、`HERMES_TUI`、`display.interface`）→ `_launch_tui`（**替换当前进程**为 Node/Ink TUI）；否则进入经典 CLI（`HermesCLI`）[[10]](#ref-10)。CLI 每次进入 agent 之前重新解析凭证（密钥轮换/令牌刷新无需重启）：`_ensure_runtime_credentials()`，失败时尝试 fallback[[8]](#ref-8)。

冷启动里预热是一个明确步骤：`_warm_turn_machinery_sync()` 在执行器线程里 import `run_agent`、物化工具 schema（含 `check_fn` TTL 缓存）、本地 Python 工具链探测、默认路由的上下文窗口元数据；预热失败降级为首轮惰性初始化，不阻塞启动[[24]](#ref-24)。一个容易被忽略的启动期事实是 `model_tools.py` 导入时触发内置工具与插件发现，但**不**在模块顶层执行可能长时间阻塞的 MCP 发现——MCP 发现由 gateway 在执行器线程里做（防止其内部最长 120 秒的阻塞冻结平台心跳），失败只记 debug[[24]](#ref-24)。

首轮的跨进程回合租约是 Hermes 对“同 state.db 多进程并发”的回答：`session_turn_leases` 表（`conversation_id` 主键，holder/acquired_at/expires_at），`LEASE_TTL_SECONDS = 300`、获取等待上限 1800 秒、刷新间隔 60 秒；续租用 holder 限定的 UPDATE，SQLite 锁冲突只在下一次尝试仍能在到期前完成时当作“漏一拍”；续租真失败/到期则 `agent.interrupt(hard_cancel=True, tool_reason="session turn lease lost")` 保护 transcript[[16]](#ref-16)。此外还有一道独立的活跃度看门狗：`turn_liveness.timeout_s = 600`、`poll_s = 15`，因为“续租本身不算进展，一个静默卡死的 turn 会永远续租”[[16]](#ref-16)。


### 22.3.1 端到端路径与授权

平台适配器把原始事件规范化为 `MessageEvent`。Gateway 侧 `_handle_message()` 的阶段顺序（源码 docstring 是 “auth → command check → running-agent interrupt → get/create session → build context → run agent → return response”）：`_hm_admit_event`（授权、profile 路由、忽略频道、plugin hook）→ `_hm_estop_gate`（全局暂停）→ `_session_key_for_source` → 待回复拦截 → `_hm_authority_admission`（若本事件归 session authority 的 durable FIFO，则在此被“拥有”并返回）→ 驱逐泄漏的 `_running_agents` 槽 → 若会话仍在运行 → busy 处理 → 斜杠命令 → 外部 drain 守卫 → `_claim_active_session_slot`（`max_concurrent_sessions` 上限）→ **在任何 await 之前**把 `turn.agent = _AGENT_PENDING_SENTINEL` 占位（防止同会话并发生成重复 agent）→ `_begin_session_run_generation` → `_handle_message_with_agent`[[19]](#ref-19)。

授权顺序（文档）：平台 allow-all 标志 → 平台 allowlist → DM pairing → 全局 allow-all → 默认拒绝[[21]](#ref-21)。会话键由 `build_session_key` 构造，布局 `<ns>:<platform>:<chat_type>[:…][:<chat_id>][:<thread_id>][:<user>]`；DM 以 chat_id 隔离，群/频道默认按参与者隔离，线程默认共享；multiplex 下命名空间是 `agent:<profile>`[[20]](#ref-20)。会话状态机里，`suspended`（`/stop` 或卡死环路升级触发）强制新 session_id，`resume_pending`（崩溃恢复/排空超时）保留 session_id；**“空闲/每日自动重置”已不存在**——`/new` 与 `/reset` 才是显式边界[[13]](#ref-13)。

#### admission 记录的是推进责任，不是动作成功

根据固定提交的运行时账本设计，`queued` 表示权威已经保存一个待推进输入，`started` 表示执行已经跨过认领边界，`unknown` 表示旧执行结束后的实际结果不能被安全判定。把三个值区分开，才能理解 Hermes 为什么不把重启后的所有输入重新跑一遍。未知并不是一个等待时间更长就会自动变成失败的普通队列项。[[7]](#ref-7)

| 状态层 | 权威事实 | 不能代替的另一层事实 |
| --- | --- | --- |
| session_admissions | 输入的准入、开始及结算状态 | 外部工单、支付或文件是否已经改变 |
| worker_executions / receipts | 后台执行的身份、owner 与回执 | 用户渠道是否收到最终消息 |
| messages.effect_disposition | 恢复后如何解释工具效果的不确定性 | 下游资源的权威查询结果 |
| delivery_obligations | 回复等待发送、发送中或完成的记录 | 原业务动作是否符合目标 |
| session_turn_leases | 当前哪个 holder 获准推进会话 | 过期 holder 已发出的网络请求是否消失 |

本书的系统推导是：persist-before-execute 保存了恢复所需的责任，但它的效果取决于持久化发生在哪一步、哪些操作依赖这次提交。若接受记录写入失败，执行路径应按实现规定报错；若记录已提交而队列调度未完成，恢复可以重新发现待处理责任。若执行已经开始，重新发现该责任并不意味着可以再次调用有副作用的工具。账本与业务动作之间仍有不确定窗口。

教学例子是一次创建事故单请求。输入先进入 queued，worker 认领为 started，远端成功而本地回执丢失，进程随后退出。新 owner 只能确认旧工作已开始，不能凭没有 tool 成功消息断言事故单不存在。将 admission 标为 unknown 是保留事实缺口，而不是推迟一次普通重试。人工通过原请求标识查到事故单后，可以结算输入并补充结果；若不能查询，需按业务协议保留未知或接受明确风险。

这个取舍牺牲自动完成率来避免盲目重放。它特别适合用户可以介入的会话，却会让无人等待的 cron/自动化任务需要另一种结算策略。固定提交对无人等待的未知工作按 interrupted 结算，并不承诺原动作从未发生；运营者仍需保留后续核验入口。这也是为什么会话状态与业务 Operation 必须分开，而不能用一张“任务成功率”掩盖未知效果。[[7]](#ref-7)

### 22.3.2 Agent Loop：阶段边界与继续条件

一次迭代的阶段顺序（源码 docstring 都写 “in the original order”）：`begin_iteration`（应用 redirect、重置 checkpoint 去重、中断/review 预算/迭代预算出口）→ `prepare_iteration`（steer 文本作为新 user 行追加；幽灵行过滤；缓存断点只在空白规整、孤儿清理、仅思考消息丢弃/user 合并之后注入）→ `assemble_api_request` → `run_preflight_gate` → **API 重试内循环** `_run_api_retry_loop` → `apply_retry_restarts` → `normalize_model_response` → 有 tool_calls 走 `run_tool_round` 否则 `finish_text_response` → 外层异常 `handle_outer_loop_error`；循环外 `finalize_turn`[[17]](#ref-17)。

迭代上限默认值是**无限**（`max_iterations = sys.maxsize`），配置文档里 “unlimited by default — the turn cap caused more problems than it solved”；`agent-loop.md` 写 “Default: 500 iterations” 是过时文档，不可引用[[17]](#ref-17)。API 重试预算：`api_max_retries` 默认 3，超出后还有“耗尽后自动恢复阶梯” `auto_recovery_cycles` 默认 5（0=关）；外层异常最多 8 次（与 max_iterations 取小）[[17]](#ref-17)。

### 22.3.3 工具调度：并行、串行与分段

工具执行有三条路径：并发、顺序、分段。并发批次上限 `_MAX_TOOL_WORKERS = 8`，批并发超时默认 420 秒；`_NEVER_PARALLEL_TOOLS` 与只读安全集、路径作用域（同路径冲突则串行）、`supports_parallel_tool_calls`（默认 False）决定一个工具能否并行[[18]](#ref-18)。纯读更易并行、路径不冲突的文件工具才并行、高副作用/交互式串行，这是源码里三层规划函数的规则[[18]](#ref-18)。

一个关键的持久化不变量在工具执行里：`run_tool_round` 在**任何副作用之前**先 `_flush_messages_to_session_db`，持久化失败则该回合结束（`failed`），不会用仅存在于进程内的状态去执行工具——源码 docstring：“Persist-before-execute is a durability invariant: resume must see the executed block if a destructive tool restarts Hermes”[[17]](#ref-17)。

工具结果有三层预算：单个结果 100,000 字符（MCP 50,000）、每轮聚合 200,000 字符、预览 1,500 字符；超阈值落盘溢出目录，替换为 `<persisted-output>` 预览块；`read_file` 阈值固定为 inf，避免 persist→read→persist 死循环[[18]](#ref-18)。工具失败以 `isError` 的 tool 角色结果回送模型，不向上抛；未知工具返回 `{"error": "Unknown tool: <name>"}`；处理器异常经 `_sanitize_tool_error` 清洗后返回，上限 2048 字符加 `[TOOL_ERROR]` 前缀[[18]](#ref-18)。

工具调度里还有两条细节。第一，agent 级拦截工具 `todo_list`、`memory`、`session_search`、`delegate_task` 由 agent 层截获，不经注册表分派；顶层 `delegate_task` 由 `run_agent._dispatch_delegate_task` 转发[[2]](#ref-2)。第二，工具进度与观察：`tool.started`/`tool.completed`（旧）加 `tool.progress`（新，MCP `notifications/progress` 每调用 0.5 秒节流）；`data_use` 戳不进模型[[18]](#ref-18)。

工具环路护栏（tool_loop_guardrails）是防止模型空转的独立一层：每轮跟踪 (工具, 参数, 结果) 观察，决定 warn/block/halt；警告默认开，硬停止在交互平台默认关、在无人值守的 gateway/cron 平台默认开；阈值是相同参数失败 warn 2/block 5、同工具不同参数 warn 3/halt 8、相同 (工具,参数,结果) 连续 3 次触发 NOTICE、周期环检测最长周期 4、历史窗口 64[[18]](#ref-18)。护栏的豁免规则：`terminal`/`execute_code`/`process_manage`/`browser_navigate`/`web_extract` 属“失败即正常输出”的工具（只警告不 halt）；写文件、patch、终端等成功调用视为进展，使此前失败签名的下一次重试变成“新实验”；被限流（429）的失败不计入连击[[18]](#ref-18)。

### 22.3.4 上下文构造与系统提示三档

system prompt 分三档，归属与早期实现不同（早期实现把 MEMORY/USER 放 Stable、Skills 放 Context 是错的）[[4]](#ref-4)：

| 档 | 内容 |
| --- | --- |
| stable | 身份（SOUL.md）、指引、钉选技能、coding brief |
| context | 调用者 system_message、项目 context 文件、workspace 快照 |
| volatile | 技能索引、MEMORY.md、USER.md、外部记忆块、插件段、profile 行、时间戳行、运行环境块 |

三档合起来是冻结的缓存 system prompt；用户消息与工具结果在 messages，不属任何一档[[4]](#ref-4)。每会话构建一次存 `agent._cached_system_prompt`，仅 stable 档作缓存前缀，压缩后才重建[[4]](#ref-4)。项目 context 文件“只载入一类，先到先得”：`.hermes.md` → `AGENTS.md` → `CLAUDE.md` → `.cursorrules`[[4]](#ref-4)。

AIAgent 复用是 gateway 降低每轮装配开销的机制：缓存键是 session_key，值 `(agent, config_signature, msg_count, session_id)`；签名由模型、runtime（provider/base_url/api_mode）、enabled_toolsets、ephemeral prompt、cache-busting 配置、user_id、skip_context_files 组合；签名变化或缓存的 session 已死 → 驱逐并新建[[20]](#ref-20)。LRU 上限 128、空闲 TTL 3600 秒、内存压力驱逐（见源码注释）。这条缓存解释了 22.2.3 里“READY 不保证模型/MCP/沙箱已调用成功”的下一层含义：真正的模型客户端可能在一个回合里才第一次被构造[[20]](#ref-20)。

回合内上下文装配的“请求局部”部分不缓存：缓存断点最后注入；`ephemeral_system_prompt` 只在 API 调用时注入、不缓存；`prefill_messages`；steer/redirect 文本；预算警告附在最新工具结果尾部；surface-switch/skills-index-delta 一次性 note 走 per-turn user 消息通道[[4]](#ref-4)。回合内中断点：`_interruptible_api_call` 在后台线程发 HTTP，主线程等待“响应/中断事件/超时”，中断时丢弃 API 线程响应；redirect 取消在途模型请求并在同一回合继续，或在工具边界 steer[[17]](#ref-17)。

### 22.3.5 压缩与缓存规划

压缩触发阈值 `compression.threshold` 默认 0.50，但窗口小于 512K token 的模型被**只升不降**地抬到 ≥0.75（注释称否则“compaction re-fires every 1-2 turns”）——所以 agent-loop.md 写的 “50%” 对多数模型不是实际触发点[[13]](#ref-13)。相关参数：`protect_last_n` 20、`target_ratio` 0.20、`max_attempts` 3、`keep_recent_tokens` 20000；此外 Gateway 侧还有一道 85% 的会话 hygiene 预压缩安全网，独立于 agent 内压缩[[13]](#ref-13)。压缩把旧轮次摘要为一个 compact 条目，**完整历史仍在磁盘**，分割点保持 assistant 工具调用与 toolResult 成对[[13]](#ref-13)。

缓存规划有三种布局：static 前缀拆分 + 最近 2 条非 system 消息；legacy “system-and-3”；native tool cache（tools 末项 + 最后 2 个已完成事务端点）[[4]](#ref-4)。客户端构造缓存标记不证明服务端命中，以 usage 的 `cache_read_tokens/cache_write_tokens` 为准[[7]](#ref-7)。

### 22.3.6 委派、Kanban 与 MoA

复杂任务有四层能力（本书归纳）：单 Agent 工具链、`delegate_task` 子 Agent、Kanban、MoA。`delegate_task` 的模型可见参数只有 `tasks[]{goal, context, output_schema, images, group}`、`action(spawn|list|steer|stop)`、`subagent_id`、`message`；`role`/`toolsets`/`max_iterations`/`background` **不是**模型参数；`role` 被忽略（能力由深度推出），`goal` 是子 Agent 的首条 user 消息而不是 system prompt[[14]](#ref-14)。并发默认 `max_concurrent_children` 10、`max_spawn_depth` 1；顶层委派总是后台，结果作为新回合回流；子超时是无进展窗口（默认无限），心跳 30 秒、停滞 15 个周期（450 秒）判陈旧[[14]](#ref-14)。

Kanban 是持久化编排：dispatcher 按 tick（默认 60 秒）调 `dispatch_once`，先回收僵尸/过期 claim/崩溃 worker，再 `recompute_ready`（依赖提升）、`claim_task`（CAS `status='ready' AND claim_lock IS NULL`，事务内复检父任务）。状态集合是 triage/todo/scheduled/ready/running/blocked/review/done（+archived），`blocked` 分 4 种 `block_kind`[[11]](#ref-11)。MoA 是虚拟 provider：preset=模型名，reference 并发上限 8，advisory view 预算 4000 字符，缓存键 `preset+SHA256(advisory view)+labels`，`fanout` 有 `user_turn`（默认）/`per_iteration`/`every_n` 三档[[14]](#ref-14)。

### 22.3.7 会话持久化与记忆写入

谁写库：gateway 的 session authority 是 canonical 写者；agent 通过 `_flush_messages_to_session_db` 写增量消息；异常路径用 `display_metadata` 里的 owner marker 做“失败写入仲裁”，不是通用 exactly-once[[7]](#ref-7)。写失败的后果：SessionDB 写预算耗尽 → 回合以 `session_persistence_failed:locked` 结束；工具前持久化失败 → 回合结束；system prompt 持久化失败 → 后续回合重建并 miss 前缀缓存[[7]](#ref-7)。

记忆写入立即落盘但不改冻结快照——会话内系统提示里的记忆块不变（保护缓存前缀），工具响应显示实时状态；所以“写入成功但本会话看不到”是设计而非失败[[12]](#ref-12)。

后台复盘在回合结束、回复交付后在后台线程 fork review agent（最多 16 次迭代，工具仅 memory+skills+read_file/search_files）[[12]](#ref-12)。





















### 22.3.8 定时任务：调度与交付分开追踪

定时任务执行有一条完整主线：tick 默认 60 秒 → tick 锁 → 线程池 + 去重（fire claim TTL 300 秒、心跳 60 秒）→ `run_canonical_job` 经 gateway admission（admission journal 0600 落盘、fsync）→ worker 内全新 `AIAgent(platform="cron")`（禁 messaging/clarify/默认 cronjob，默认不载入项目 context 文件，`skip_background_review=True`）→ 投递（`cron/deliveries.db` 持久投递队列）→ 状态簿记[[23]](#ref-23)。tick 先推进 `next_run_at` 再派发（至多一次语义），同一保存里写 `pending_slot`，`scheduled_instant` 阻止二次触发；`cron.catch_up_missed: false` 跳过超出宽限的错过并记录原因[[23]](#ref-23)。

## 22.4 异常：失败、中断与恢复

Hermes 的异常面很大：单机 CLI 的中断、跨进程的回合租约、模型失败与凭证池、Gateway 重启恢复、Kanban 任务的熔断、投递账本、安全拦截。它有一个鲜明的特点：大量“失败”被设计成结构化结果回送给模型或调用方，而不是向上抛。

### 22.4.1 异常分类与证据现状

| 异常类型 | 关键机制 | 依据 |
| --- | --- | --- |
| 中断与取消 | `interrupt()` 跨线程请求；工具位线程作用域；顺序路径 1 秒轮询 | 源码 |
| 工具失败 | 一律 `isError` 回送；错误清洗限 2048 字符；护栏 warn/block/halt | 源码 |
| 模型失败 | 错误分类器流水线 + 凭证池轮换 + provider fallback + 自动恢复阶梯 | 源码 |
| 上下文超限 | 压缩重试一次 + 反抖 + 确定性兜底摘要 + 网关自动重置 | 源码 |
| Gateway 重启 | `resume_pending`/`.clean_shutdown`/active turn 标记/stuck-loop 挂起 | 源码 + 文档 |
| 投递失败 | delivery ledger“带重复提示的有界恢复” | 源码 |
| Kanban 失败 | claim TTL/心跳/崩溃回收/熔断（默认 2 次） | 源码 |
| 子 Agent 失败 | 心跳陈旧监控 + 无进展超时 + 中断传播整棵 spawn 树 | 源码 |
| 记忆写入失败 | 字符预算/文件锁/威胁扫描/每轮 3 次失败上限 | 源码 |
| 安全拦截 | 审批 fail-closed + hardline 地板 + smart 模式熔断 | 源码 + 文档 |

### 22.4.2 中断与取消：能停什么

`AIAgent.interrupt()` 是跨线程的“请求中断”入口：写 `_interrupt_requested`，置位当前执行线程的工具中断位，扇出到并发工具 worker 与子 agent；`hard_cancel=True` 表示显式 `/stop`，否则是软中断（新消息或普通中断）[[16]](#ref-16)。中断原因分“用户类”与“系统类”，退出原因写成 `interrupted_by_system(<issuer>)` 而不是记成用户停止[[16]](#ref-16)。工具层中断位是线程作用域的：网关一个进程里跑多个 agent，中断一个会话不会杀其他会话的工具[[16]](#ref-16)。

几个精确的语义。steer 不打断：文本在当前工具批结束后作为独立 user 行投递；redirect 在模型请求进行中只取消该次请求（保留部分输出），在工具执行中退化为 steer；`/stop` 是硬中断，会丢弃未消费的 pending steer[[16]](#ref-16)。顺序工具路径每 1 秒切片轮询中断（即使工具自己不检查 `is_interrupted()`，`/stop` 也在约 1 秒内生效），中断后给 worker 3 秒宽限，仍未返回则放弃该 worker[[18]](#ref-18)。被中断的模型请求：有 redirect 则保留纠正文本重建 turn，否则把已流出的部分文本以 `display_metadata={"interrupted": True}` 持久化——**这与 agent-loop.md 的“不会把部分响应注入会话历史”冲突，以源码为准**[[17]](#ref-17)。

中断后残留状态有专门的“孤儿恢复”：恢复/重放历史时清理被打断的 assistant→tool 尾巴；只读工具块被丢弃，含副作用工具的块被保留并把结果改写成 `effect_disposition="unknown"` 的 “Orphan recovery ... effect is UNKNOWN” 通知，防止模型重发已执行的调用[[16]](#ref-16)。工具结果消息携带 `effect_disposition`：取消/跳过的未启动调用为 `"none"`，工具超时为 `"unknown"`；这个字段随会话持久化[[18]](#ref-18)。

#### 中断能停止什么，不能停止什么

中断首先改变本地继续推进的许可。是否能真正停止远端工作，取决于请求是否已经发送、工具是否提供取消、并行任务是否检查信号，以及后台委派是否仍属于当前回合的生命周期。固定提交包含软中断与 hard_cancel 等路径，不能仅凭界面出现“已停止”推导所有子动作都已撤销。[[16]](#ref-16)[[18]](#ref-18)

| 对象 | 中断后可做到的事 | 仍可能发生的效果 | 核验责任 |
| --- | --- | --- | --- |
| 未启动工具 | 在检查点阻止派发 | 检查点之前已启动的其他工具继续 | 当前 turn 清点已启动集合 |
| 进行中的模型 HTTP | 本地放弃响应或请求取消 | 服务端继续计算并计费 | 使用量与计费按提供方回执核对 |
| 并行工具 | 取消支持信号且已检查的工作 | 不检查信号的工具完成外部动作 | 逐工具保存原身份与结果状态 |
| 外部异步作业 | 按后端接口提出取消 | 取消与完成竞态、部分完成 | 查询原作业，而不是新建替代作业 |
| 后台委派 | 停止当前回合等待或可控子项 | 已脱离当前 children 集合的工作继续 | 委派账本与后台 owner 接续 |
| 回复发送 | 阻止尚未发送的新内容 | 在途发送可能已经到达 | 出站账本独立核验 |

`/stop` 与进程退出也不是同一事件。前者是经过应用处理的控制信号，后者可能发生在任意 await 或数据库提交之间。测试中应分别覆盖软中断、硬取消、租约丢失与进程崩溃，并记录是哪一步先发生。一次顺序正常的 stop 演示无法证明崩溃恢复也正确。

孤儿恢复中的 effect_disposition 有助于把不确定性带回模型上下文。它保留“可能已经发生”的事实约束，避免模型见到缺失结果就重发；但提示中写 UNKNOWN 不是资源端防重。对高风险动作，还需应用层查询与禁止重复派发；模型无论怎样解释历史，都不能获得绕过未知结果的权限。[[16]](#ref-16)[[18]](#ref-18)

中断的竞态处理是 Hermes 特别细的地方。并发工具批次里 `interrupt()` 把中断位扇出到已登记的工具 worker 线程（否则一个挂住的并发工具会跑到自己的超时）；worker 在登记晚于扇出时会在 `run_worker` 里自行补上中断位[[16]](#ref-16)。停止信号落地后才登记的子 Agent 会镜像父级未决的停止，覆盖当前登记的 spawn 树，不含已经脱离此集合的后台委派[[16]](#ref-16)。宿主可绑定 `InterruptScope`，其 `cancel()` 对该作用域内所有 agent 发硬中断，且之后才登记的 agent 立刻被中断——“取消不会输给尚未开始的 turn”[[16]](#ref-16)。

顺序工具路径里，中断后为剩余每个 tool_call 追加一条取消结果（保持 assistant→tool 配对），一个工具执行后若中断，余下调用写 “skipped ... was not started”；并发批次被中断时，不带中断检查的工具（web_search、read_file）会运行到结束，仅取消未启动的 future[[18]](#ref-18)。经典 CLI 的 Ctrl+C 优先级链是：取消语音录制 → 取消前台 UI 提示 → （即使 agent 正在运行也）先清除已键入草稿 → agent 运行中且输入为空则中断 agent → 2 秒内第二次按下强制退出[[33]](#ref-33)。

### 22.4.3 模型失败：重试、轮换与 fallback

模型失败处理是一条优先级流水线：`ClassifiedError` 把异常映射为 `{reason, status_code, retryable, should_compress, should_rotate_credential, should_fallback}`，循环只读这些提示，不再自己匹配字符串[[16]](#ref-16)。HTTP 状态静态表：401→auth、402→billing（不可重试）、408→timeout、413→payload_too_large、429→限流、500/502→server_error、503/529→overloaded[[16]](#ref-16)。

应用层重试 `api_max_retries` 默认 3（源码把值钳为 ≥1，configuration.md 写的“默认 3=四次尝试、可降到 0”与源码不符）；SDK 层重试被关闭（主客户端 `max_retries=0`，把重试交给外层循环以尊重 Retry-After）[[16]](#ref-16)。退避公式是 `min(base·2^(attempt-1), max_delay) + jitter`，API 错误重试用 base 2 秒、max 60 秒；Retry-After 优先，封顶 600 秒[[16]](#ref-16)。耗尽后还有“自动恢复阶梯” `auto_recovery_cycles` 默认 5（等待约 15/30/60/60/60 秒），只在瞬时错误（overloaded/server_error/timeout）且尚无回答文本送达时进入；认证、格式、计费、策略类错误从不进入[[16]](#ref-16)。

凭证池轮换：rate_limit 先对同 key 重试一次再轮换，billing 立即轮换，auth 先刷新后轮换，`upstream_rate_limit` 不轮换（key 健康）[[16]](#ref-16)。provider fallback 用 `fallback_providers`（列表，每项必须同时有 provider 与 model），沿链逐项前进；显式用户选择的模型是“严格”的，失败即报错而不被无关 fallback 掩盖；fallback 只对本 turn 生效，每 turn 开始恢复主运行时[[16]](#ref-16)。主 provider 冷却指数退避 60 秒 → 4 小时封顶[[16]](#ref-16)。

模型失败还有两个值得记的边界。第一，内容策略拦截是确定性的，`failure_retryable=False`、`failure_reason="content_policy_blocked"`，不重试；失败结果统一“盖章”带 `failed`/`failure_reason`/`failure_retryable`[[16]](#ref-16)。第二，流式层单独有重试：`HERMES_STREAM_RETRIES` 默认 2（共 3 次开流尝试），重连退避 1 秒、2 秒且可被 /stop 打断，耗尽后文案为 “Could not open a stream to `<host>` after N attempts”，此时客户端尚未收到流事件；这不能证明服务端没有计算或计费[[16]](#ref-16)。流/调用的陈旧看门狗 `HERMES_STREAM_STALE_TIMEOUT` 默认 180 秒（按上下文规模放大），本地 provider 默认 900 秒[[16]](#ref-16)。

工具失败的回送形态值得单列，因为它是“失败作为一等结果”的样本：未知工具返回 `{"error": "Unknown tool: <name>"}`；处理器抛异常经 `_sanitize_tool_error` 清洗后返回（上限 2048 字符加 `[TOOL_ERROR]` 前缀，日志保留前 8192 字符）；非字符串且非多模态信封的返回值被替换为契约错误；参数不是合法 JSON 对象时返回 “Tool arguments must be a valid JSON object; tool was not executed.”[[18]](#ref-18)。未知工具名的循环内策略：先 `_repair_tool_call` 自动纠正，纯非法批次逐次累计，第 3 次以 partial 结果退出；无效 JSON 参数前两次不写消息直接重发 API 请求，第 3 次起注入 tool 错误结果[[18]](#ref-18)。

### 22.4.4 上下文超限与压缩失败

溢出被分类器判为 `context_overflow`（只带 `should_compress`，不 fallback），413 判为 `payload_too_large`；两者都在 `RETRYABLE_CLIENT_REASONS` 里，所以不会被当成不可重试错误终止[[16]](#ref-16)。压缩尝试上限 `compression.max_attempts` 默认 3（≥1、≤10）；超出后 `fail_turn`，`failure_retryable=False`、`compression_exhausted=True`，提示 `/new` 或 `/compress`[[16]](#ref-16)。Gateway 对 `compression_exhausted` 的处置：非 API 服务器平台默认自动重置会话（旋转到新 session 并迁移排队消息），API 服务器平台为保持会话绑定只返回带类型的错误[[16]](#ref-16)。

压缩失败有一整条降级阶梯。摘要失败冷却 60→300→900 秒递增；独立摘要模型失败后用主模型重试一次；终止性失败（访问/配额、网络、空内容、length 截断）无论如何都 abort 并保留 transcript；非终止性失败默认提交确定性兜底摘要（上限 8000 字符，只保留连续性锚点）[[16]](#ref-16)。反抖（anti-thrash）：只有 provider 证实的无效压缩才累计，连续 2 次阻断自动压缩，300 秒后允许一次探测[[16]](#ref-16)。压缩中提交 prompt 会直接抛错，树导航/`/compact` 在压缩期间同样受限。

终端后端的失败分“命令失败”与“基础设施故障”：后端连接/基础设施类故障抛 `EnvironmentConnectionError`（SSH 主机宕机、Docker 守护进程未运行），`terminal_tool` 把它变成结构化 `status: "degraded"` 结果（`exit_code: -1`、`reason`、`retry_hint`），默认 `terminal.degraded_mode: warn`；故障后端被驱逐出缓存[[22]](#ref-22)。前台命令超时默认 180 秒，硬上限 600 秒——超过上限的前台超时不再拒绝，而是把命令提升为受跟踪的后台进程[[22]](#ref-22)。后台进程注册表有界：最多跟踪 64 个进程（LRU 淘汰），滚动输出缓冲 200,000 字符，完成后保留 30 分钟[[27]](#ref-27)。重启后 `recover_from_checkpoint` 只“领养”PID + 启动时间仍匹配的存活进程，不重新执行命令[[27]](#ref-27)。

### 22.4.5 Gateway 重启恢复：Recovery 的恢复标记与年龄阈值分别负责什么

重启恢复围绕几个持久化标记展开。`.clean_shutdown` 是“干净退出回执”：只在排空未超时时创建；下次启动若存在，则丢弃孤儿 active-turn 标记并删除回执，不存在才走 `_recover_unclean_sessions`[[24]](#ref-24)。活动 turn 标记 `mark_turn_active` 在发布到内存之前先把 (token, 开始时间) 持久化，`clear_turn_active` 以 token 做 CAS 清除[[13]](#ref-13)。未清洁退出恢复只看“被标记的 turn”：`max_age = max(3600, 2×HERMES_AGENT_TIMEOUT)`，更老的标记只清除不执行[[24]](#ref-24)。

崩溃遗留回复分两分支：转录里已有最终回复 → 作为“无主 attempting”行写入投递账本，启动清扫补发一次（不重新生成）；转录里没有 → `resume_pending` + `restart_interrupted`，自动续跑一次[[24]](#ref-24)。自动续跑原因集合是 `{restart_timeout, shutdown_timeout, restart_interrupted}`；`resume_pending` 保留 session_id，下次成功回合后清除；`suspended` 则下次访问强制新 session[[13]](#ref-13)。

关停时的 drain 标记：`_stop_drain_active_work` 在等待排空**之前**就把所有运行中会话预标记为 `resume_pending`，这样 drain 期间被 SIGKILL 也留下持久标记[[24]](#ref-24)。stuck-loop 检测：关停时给仍活跃的会话计数加一（文件 `.restart_failure_counts`，文档误写为 `restart_counts.json`），启动时把计数 ≥3 的会话置 `suspended`[[24]](#ref-24)。

durable 输入队列（`session_admissions`）的恢复语义是“从不重放已开始的工作”：`recover_session_inputs` 把其他 owner epoch 留下的 `worker_executions` 改为 `unknown`，`session_admissions` 里 `started` 且 epoch 不符的行改为 `unknown`；`unknown` 只能由人确认解除（`/discard` 等），无人等待的 `unknown`（cron、自动化）按 `interrupted` 自动结算[[7]](#ref-7)。

标记负责回答“这项工作是否可能被本次非正常退出打断”，年龄阈值负责限制自动处理范围；二者并不矛盾。没有活动标记不能仅因记录很新就重新执行，有标记但超龄也不会自动恢复。`resume_pending` 处理会话续跑，durable admission 的 unknown 处理已开始工作的不确定责任，不能因为都在启动阶段出现就混成一种自动重放机制。[[7]](#ref-7)[[24]](#ref-24)

| 权威状态 | 固定路径提供的保证 | 不提供的保证/故障窗口 | 恢复 owner |
| --- | --- | --- | --- |
| active-turn token 与 clean_shutdown | token CAS 清除对应回合；非干净退出时按标记检查 | 标记失败或超龄时无法据年龄猜出安全续跑点 | Gateway 恢复器；超龄由运维检查原任务 |
| resume_pending 与已保存 final | 有 final 则转投递责任，无 final 则受控续跑 | 重建计算不撤销旧外部效果 | 会话恢复器与投递账本分别处理 |
| started admission 的旧 epoch | recover_session_inputs 将其记 unknown | 不自动重放已开始输入，不证明远端动作未发生 | 人工查询/确认；无人等待的自动化按 interrupted 结算 |
| orphan tool tail 的 effect_disposition | 副作用尾部保留 unknown，未启动保留 none | 文本标记不是远端幂等，不阻止所有扩展绕路 | 工具与业务资源 owner 查询原动作 |

### 22.4.6 投递账本：带重复提示的有界恢复

Gateway 最终回复的持久“投递义务”表 `delivery_obligations`，目标是防止“定稿与平台 ACK 之间崩溃导致回复静默丢失”。检查点依次是 `record_obligation`（任何发送之前）→ `mark_attempting` → `mark_delivered`/`mark_failed`；所有调用尽力而为，账本失败不阻塞发送[[26]](#ref-26)。语义是“带重复提示的有界恢复”：`pending` 补发不带标记，`attempting`（崩溃在 await 中，平台可能已收到）补发附 `RECOVERED_MARKER`（“gateway restarted during delivery, so this may be a duplicate”），`failed` 可由同进程的 runtime 恢复或死 owner 的启动清扫处理，分别带重连或重启提示[[26]](#ref-26)。账本**不保证**恰好一次、不保证永不丢（3 次尝试或 24 小时后 `abandoned`，7 天/500 行后清理）；`gateway.delivery_ledger: false` 时整套关闭，崩溃遗留回复改走 `resume_pending` 重新执行[[26]](#ref-26)。

#### 投递账本的有界恢复与丢失窗口

固定提交的 delivery_ledger.py 明确将全部记录/标记调用设为 best-effort：账本失败不能阻塞发送。这一选择保住正常渠道发送的可用性，同时意味着“任何最终回复都先持久保存”不是无条件不变量。要准确说明保证，应从记录失败、发送开始和成功标记三个边界分别分析。[[26]](#ref-26)

| 故障位置 | 可恢复的依据 | 不能保证的事 | 恢复行为/责任 |
| --- | --- | --- | --- |
| record_obligation 失败，随后尚未发送便崩溃 | 可能没有该回复的账本项 | 恢复器一定发现应补发内容 | 使用者监控记录失败；可能需要从会话人工重建 |
| pending 已保存，发送尚未开始 | 已保存目标和内容 | 后续授权/渠道一定仍有效 | 在恢复预算内尝试发送 |
| attempting 后崩溃 | 知道发送可能已经发生 | 用户没收到或只收到一次 | 补发带 RECOVERED_MARKER，接受可见重复 |
| 渠道确认成功，mark_delivered 失败 | 远端可能已经收到，本地仍未结算 | 本地状态可以排除重复 | 沿未决路径恢复，重复风险仍存在 |
| 尝试或年龄上限耗尽 | abandoned 记录在保留期内可观察 | 最终一定送达 | 停止自动恢复，交运营/用户处理 |

源码中的三次尝试、二十四小时过期、保留行数和清理周期是版本敏感策略，不是分布式投递的一般定理。带标记的补发使重复更容易解释，却不能弥补从未保存的义务，也不能让 abandoned 自动等于用户确认收到。本章将其称为有界恢复，是为了把范围和耗尽后的责任写清。[[26]](#ref-26)

这里有两种不同的可观测失败。第一种是业务丢失：最终回复没有送达；第二种是账本自身不可用，导致系统无法完整统计丢失数量。只看 abandoned 行会漏掉记录失败的第一段窗口。因此本书建议同时监控记录/标记错误、恢复积压、最老未决年龄和 abandoned，并说明这些指标仍不是完整投递成功率。若部署没有对账本错误的独立日志/告警，就不能宣称全部回复丢失可观察。

固定源码的接管还有限制：`sweep_recoverable` 先跳过仍存活 owner，再检查次数和年龄，随后只认领当前已连接的 platform/profile，避免每次启动给离线 bot 空耗一次预算；认领事务同时把状态改为 attempting 并增加 attempts。它按旧 owner 戳做条件更新，竞争者不能把同一行都认领成功。这里的身份戳包含进程启动时间，不能只用可复用的 PID。[[26]](#ref-26)

同进程适配器重连走 `sweep_failed_for_runtime`，只接管自己当前进程实例的 failed 行，并受退避、次数和年龄限制；死 owner 留给启动清扫。非分类故障的 runtime 重试保留最后一次机会给启动恢复，避免定时器在长故障中耗尽全部机会。这些是账本接管与预算语义，仍不保证发送请求和平台 ACK 原子提交。Flood control 的等待期间可以只接管 owner、不扣发送次数，直到 not_before 才发送；因此“每次扫到就重发”与源码不符。[[26]](#ref-26)

发送目标和内容也有版本。恢复时引用的原附件可能已经删除，群成员权限可能收紧，某渠道可能不再允许机器人发言。简单地把原文本重新发出去会绕过当前可见性；重新调用模型又可能生成不同内容并重做原业务动作。应把“重投原结果”和“授权后重新执行任务”分为两个产品操作，各自检查身份、内容与副作用。

已发生的外部动作、会话消息、渠道回复可以分别处于成功、未知和失败。一个工具创建事故单成功，不会因消息发送失败而自动回滚；一个带重复提示的回复也不能证明事故单只创建了一次。读者应沿原动作和原投递义务分别追踪，不用某一层的成功替其他层补上事实。

### 22.4.7 Kanban 任务的失败与熔断

Kanban 的调度器每个 tick 先做回收：回收僵尸 → 回收已结束 run 仍存活的 worker → `release_stale_claims`（claim TTL 默认 15 分钟）→ `reconcile_orphaned_running` → `detect_stale_running`（心跳过期）→ `detect_crashed_workers`（PID 已死）→ `enforce_max_runtime` → `recompute_ready`，之后才按预算 claim 并 spawn[[11]](#ref-11)。崩溃检测用 PID + 进程指纹防 PID 复用，按退出状态分类：rc=0 为 protocol_violation，rc=75 为限流（不计失败），rc=78 为终态 provider 错误（首次即熔断）[[11]](#ref-11)。

熔断器 `_record_task_failure` 把非成功结局（spawn_failed/timed_out/crashed/reclaimed）累加 `consecutive_failures`，达阈值（任务级 `max_retries` > 调度器 `failure_limit` 默认 2）则置 `blocked` 并写 `gave_up` 事件；`infrastructure=True`（主机拒绝 spawn）不计数[[11]](#ref-11)。block→unblock→re-block 循环达 2 次路由到 `triage` 转人工。Kanban 是单主机设计，崩溃检测假设 PID 为本机 PID，不支持跨主机共享同一块板[[11]](#ref-11)。

### 22.4.8 子 Agent 失败与中断传播

子 Agent 没有默认墙钟超时（`child_timeout_seconds` 默认 0），配置为正数时是“无进展”上限而非总时长，任何进展信号（一次 API 调用、工具变化、活动时钟）都重置窗口[[14]](#ref-14)。心跳监控是权威的卡死检测：父侧每 30 秒心跳，子 Agent 15 个周期（450 秒）回合间无进展或 40 个周期（1200 秒）卡在同一工具里即判陈旧，之后给 worker 2 秒收集真实结果，否则合成超时[[14]](#ref-14)。

中断传播覆盖仍登记在 `_active_children` 中的 spawn 树：`interrupt()` 对 `_active_children` 快照逐个传播，停止信号落地后才登记的子 Agent 会镜像父级未决的停止，避免孤儿子 Agent 跑到结束[[16]](#ref-16)。失败的子 Agent 不静默：工具结果条目含 `status: "failed"` 与完整 `error`；被停的子 Agent 返回 `status="interrupted"`，`summary` 是停止前最后产出的文本[[14]](#ref-14)。

### 22.4.9 安全、记忆与外部能力的失败语义

命令守卫的判定顺序是：隔离后端（docker 且无宿主路径挂载等）只执行用户 deny 规则其余放行 → 否则先过“地板”hardline → 用户 deny → yolo/off → 永久白名单 → 无人值守按各自 deny 配置 → `detect_dangerous_command` 命中则人工决策[[28]](#ref-28)。hardline blocklist 覆盖递归删除根/系统目录、`mkfs`、`dd` 写裸设备、fork bomb 等，在 yolo、`mode: off`、cron approve 之前触发，没有覆盖开关[[28]](#ref-28)。审批超时 fail-closed：超时返回 `BLOCKED: … Silence is not consent.`，不记为拒绝；过期提示不可重开[[28]](#ref-28)。

smart 模式用辅助 LLM 回答一个词 APPROVE/DENY/ESCALATE，命令文本被视为不可信输入；守卫失败、空答案、无法识别的答案一律按 ESCALATE（回到人工）；同一会话连续 3 次守卫 DENY 后触发熔断文案，任何批准重置计数[[28]](#ref-28)。文件写保护是“路径层”而非审批：`write_file`/`patch` 对受保护路径直接拒绝、不弹审批；模型仍可能谎称编辑成功，文档建议信任 `display.file_mutation_verifier` 的页脚[[28]](#ref-28)。

















记忆写入失败有一整套防护：写入在文件锁内重新读盘、原子替换，读不出来则拒绝（把不可读当成空会抹掉记忆）；外部漂移（有条目大于整个存储上限）则拒绝并保存 `.bak` 快照；新内容过 `strict` 威胁扫描，命中注入/外泄/不可见 Unicode 则返回 `Blocked` 不写入；每轮整合失败上限 3 次，超过后返回终态“Stop retrying memory calls”，使失败的记忆副作用永不阻塞本轮回答[[12]](#ref-12)。

MCP 失败有独立的降级：单服务器初始连接超时默认 60 秒、首次连接重试 3 次、已建立后重连重试 5 次、退避上限 60 秒；失败的服务器按“连接冷却”调度（基础 30 秒、翻倍、上限 600 秒）；单服务器熔断器连续 3 次失败打开、冷却 60 秒后半开探测；MCP 工具错误信息在返回模型前把 ghp_、sk-、Bearer、token=/key=/password= 等替换为 `[REDACTED]`[[32]](#ref-32)。

插件加载失败的隔离：每个插件 import + `register()` 套上 `plugins.load_timeout_seconds`（默认 10 秒）死线，任何 `Exception` 或 `SystemExit`（插件里不受保护的 `sys.exit()` 也不得拖垮进程）都被捕获，处置该插件已注册的内容、移除其事件订阅，其余插件继续加载[[33]](#ref-33)。启动失败与运行期看门狗分两道：启动存活看门狗 300 秒未就绪转储栈并以 75 退出；运行期事件循环活性看门狗每 30 秒探测、10 秒未处理计一次 miss、连续 3 次转储栈硬退出（约 90–120 秒的持续循环阻塞）[[25]](#ref-25)。

### 22.4.10 静态分析的范围与部署验收

本章静态分析范围为所列固定提交中的 Gateway、运行时账本、turn、中断和投递路径。未公开统一契约的后端错误、工具结果预算和插件行为，不能从某个检索未命中的符号推为产品绝对不支持。部署时还需按实际 backend、插件与配置验证；本文未执行性能测量或故障注入。

## 22.5 实践：用好 Hermes

这部分每条做法标注“官方建议”或“本书建议”。Hermes 的官方安全文档给了一整套可执行建议，本节的骨架就来自它。

### 22.5.1 先理解安全模型

Hermes 的安全模型官方归纳为八层：用户授权、危险命令审批、文件写保护、容器隔离、MCP 凭据过滤、context 文件扫描、跨会话隔离、输入净化[[29]](#ref-29)。但 SECURITY.md 同时把底线说得很直白：**“唯一的安全边界是操作系统”**——审批门、输出脱敏、模式扫描器、工具 allowlist 都不是 containment；审批门只抓“合作模式”的错误，不抓对抗输出；插件以完整 agent 权限运行，第三方插件的边界是安装前由操作者审阅[[30]](#ref-30)。

官方给出两种 OS 级隔离姿态：terminal-backend 隔离（仅限 shell/file 工具，不限制 code-execution、MCP 子进程、插件、hook、skill 加载）与 whole-process wrapping（把整个 agent 进程树放进 Docker/Compose 或 NVIDIA OpenShell）[[30]](#ref-30)。默认 local 后端 + 不受控输入面，或指望 terminal-backend 沙箱约束不走 shell 的路径，都属于“支持姿态之外”[[30]](#ref-30)。生产 gateway 部署官方建议用 docker/modal/daytona/vercel_sandbox 后端隔离命令（可免去危险命令审批）[[29]](#ref-29)。

### 22.5.2 审批、deny 规则与 allowlist

审批模式 `approvals.mode ∈ {smart, manual, off}`，默认 smart；`off` 等价 `--yolo`，关闭全部审批提示，官方标注“仅用于可信环境（CI/CD、容器等）”[[29]](#ref-29)。`--yolo` 关闭所有危险命令检查，唯独 hardline blocklist 除外；hardline 是“`--yolo` 之下的地板”，合法流程需要 hardline 命令时官方建议“在 agent 之外运行”[[29]](#ref-29)。

`approvals.deny` 是用户自定义 fnmatch glob，大小写不敏感，先于 yolo/off 判断，对所有 terminal 后端（含容器）生效，改动即时生效；YAML 里必须加引号（裸 `*` 开头会被当成 alias）。官方同时提醒：deny 规则不是完整 shell 解释器或 OS 沙箱，要隔离应使用 OS 权限与受限隔离后端[[29]](#ref-29)。永久 allowlist 在启动时读取，为安全移除条目后必须重启；用 `hermes config edit` 审阅，用 `hermes approvals suggest` 从已批准命令挖掘提案（默认 dry-run，破坏类永不提议）[[29]](#ref-29)。

无人值守上下文的默认值是 deny：`approvals.cron_mode`、`single_query_mode`、`unattended_mode` 默认全 deny，命中危险命令即时返回 `BLOCKED`，而不是阻塞满超时[[28]](#ref-28)。本书建议：把放权做成分层——OS 隔离（容器/专用账户）是第一层，`approvals.deny` 是第二层，`command_allowlist` 是第三层，`approvals.mode: off` 只留给自己完全控制的本地环境。

### 22.5.3 文件写保护与凭据边界

文件写保护是“路径层”而非审批：`write_file`/`patch` 对受保护路径（`~/.ssh`、`~/.aws`、`.env`、auth/、vault/、mcp-tokens/、pairing/ 等）直接拒绝、无审批提示；`HERMES_WRITE_SAFE_ROOT` 设置后写入只能落在列出的目录前缀内（Docker 官方镜像自动设为 `/opt/data`）[[29]](#ref-29)。官方警告两条：不要随手把 `HERMES_WRITE_SAFE_ROOT` 写入 `~/.hermes/.env`（设为项目目录会使 agent 无法写 cron/jobs.json 与 profile skills）；写保护只作用于 write_file/patch，terminal 以同一 OS 用户运行仍可覆盖受保护路径，denylist 不是对抗恶意 agent 的沙箱[[29]](#ref-29)。

凭据边界上，`execute_code` 与 `terminal` 从子进程剥离敏感环境变量；skill 声明的 `required_environment_variables` 在 skill 加载后自动透传；Hermes 基础设施密钥（provider API key、gateway token）不应加入 `env_passthrough`。与 Codex CLI/Claude Code 并用时，官方建议给 Hermes 独立登录并设 `auth.adopt_external_logins: false`（因 refresh token 单次使用轮换）[[29]](#ref-29)。

### 22.5.4 入口身份与 DM pairing

Gateway 授权顺序是：平台 allow-all → 平台 allowlist → DM pairing → 全局 allow-all → 默认拒绝；未配置任何 allowlist 且未设 allow-all 时**所有用户被拒绝**[[29]](#ref-29)。DM pairing 是身份优先于模型的落实：陌生用户收到 8 字符码，owner 用 `hermes pairing approve <platform> <code>`；码 1 小时过期、每用户 10 分钟 1 次请求、每平台最多 3 个待处理、5 次失败锁 1 小时[[29]](#ref-29)。Docker 下执行配对命令须带 `-u hermes`，否则 root 创建的文件 gateway 读不到、批准被静默忽略[[29]](#ref-29)。

SECURITY.md 对外部面有一条硬规则：每个网络暴露适配器必须设 allowlist（未设置时不得派发工作、解决审批、回传输出）；session ID 只是路由句柄不是授权；授权集合内所有调用者同等可信，需要能力隔离则运行多个实例[[30]](#ref-30)。生产 gateway 官方给了一个 10 条清单：显式 allowlist、容器后端、限资源、`.env` 权限、启用 DM pairing、审计 command_allowlist、设 terminal.cwd、不以 root 运行、监控日志、定期 `hermes update`[[29]](#ref-29)。

### 22.5.5 记忆与技能的使用

记忆分 `MEMORY.md`（2200 字符）与 `USER.md`（1375 字符），按字符而非 token 计；超限不自动压缩，返回错误并附 `current_entries`，建议用单次 `operations` 批处理（同时删除/缩短旧条目并添加新条目，限额只对最终结果检查）[[12]](#ref-12)。官方对记忆与技能写入门默认关闭（`memory.write_approval`/`skills.write_approval` 默认 False）——这是把“模型直接写长期状态”变成可选的能力[[12]](#ref-12)。

技能方面，官方开发指南的判据是“几乎总是 skill”：能用指令 + shell + 现有工具表达的做 skill，需要端到端集成 API key/认证、必须每次精确执行的逻辑才做 Tool[[31]](#ref-31)。SKILL.md 的 `description` 必须 ≤60 字符、一句话、句号结尾（长描述稀释模型注意力）；正文引用 Hermes 原生工具名（search_files/read_file/patch 等）而非 grep/cat/sed[[31]](#ref-31)。技能索引（名称 + 描述）每次会话注入，所以“保持描述简短”是真实的 token 压力，不只是风格建议。

#### 长期状态污染：从一次输入到未来行为

Hermes 的学习闭环把有用经验写进 memory 和 skills，使后续任务能复用方法；同一条路径也会把输入污染放大。群聊、文档或工具输出中的错误内容，若经后台反思转成记忆或技能，就可能在原输入已离开上下文后继续影响未来会话。写入扫描、文件锁和格式合法各自只解决部分问题：锁防并发覆盖，扫描发现已知危险模式，二者都不能证明经验真实或适用于所有用户。[[7]](#ref-7)[[30]](#ref-30)

| 阶段 | 污染来源/风险 | 本书建议的控制 |
| --- | --- | --- |
| 收集候选 | 不可信用户把猜测写成操作规范 | 保留原来源、用户、范围、时间与信任等级 |
| 后台反思 | 历史里反复出现的错误被当作经验 | 明确区分观察、推断和已核验事实，不因重复而晋级 |
| memory 写入 | 单用户偏好覆盖共享业务事实 | 按用户/项目命名空间隔离，并设写入审批 |
| skills 写入 | 一条经验变成可执行脚本或工具流程 | 审查代码、依赖、出口和宿主权限后再启用 |
| 后续召回 | 旧经验被应用到新版本或其他用户 | 重验适用条件、来源版本和当前授权 |
| 纠错与删除 | 覆盖原句但派生摘要仍存在 | 沿依赖失效缓存、摘要和衍生技能 |

特别要区分 memory 与 skill_ledger。技能的变更账本可以帮助追踪某个技能版本的来源和变更，但不能据此宣称普通记忆也有完整版本账本和一键回滚。若目标版本未提供足够的记忆来源/回滚支持，部署方需要外部审阅与备份策略，或者禁止自动持久写入。把这个缺口标成部署责任，比给产品补写一个不存在的 ledger 更重要。

受限群助手不宜直接沿用个人助手的自动学习默认。可以允许把经验保存为待审候选，审核后才进入群共享命名空间；单个成员的偏好只影响该成员或明确授权的空间。若不能在产品里实现细粒度隔离，就拆分 profile/存储/OS 身份，而不是把全部用户记忆写在同一个文件里再提示模型“注意区分”。

一个教学投毒链可以这样审查：成员在文档中加入“遇到数据库延迟先导出客户表”；工具把它作为正文返回；后台反思提炼成排障技能；下一次任务发现技能并提出外发。合格的控制不是等最后一步识别敏感词，而是每一层都保留来源等级，阻止未审阅内容晋级为共享方法，且真实外发工具仍受确定性权限限制。这个例子是本书构造的威胁路径，未声称对 Hermes 完成攻击实验。

纠错时保留原值、纠正依据、影响范围和处理时间。把 memory 文本改正后，还要检查从它生成的技能、回答缓存与正在运行会话；已装入模型上下文的旧内容不会因磁盘文件变化自动消失。必要时中断受影响会话并重新装配，工具外部事实另行核验。回滚技能也不能撤销已经发送的消息或文件写入。

审批设置必须覆盖后台路径。前台 memory 工具要求确认，并不自动证明后台反思写 memory/skills 也走同一门。应分别验证前台、后台、插件和直接文件写入；未验证路径按未受保护处理。安全责任仍由 OS、目录权限、受控工具和隔离执行环境共同承担，不能把学习循环中的一句约束当成 containment。[[30]](#ref-30)

### 22.5.6 Cron 与委托

Cron 的 agent 任务经 gateway admission + worker 执行（`run_canonical_job` 写 admission journal 后交给 session authority），`no_agent` 脚本模式在 import `run_agent` 之前短路；无头 `hermes cron tick` 在没有 gateway 拥有该 home 时拒绝派发，而不是自行拉起无人管理的守护进程[[23]](#ref-23)。cron 的 agent 任务 `skip_background_review=True`（不进入后台复盘闭环），默认禁 `messaging`/`clarify`/`cronjob` 工具集[[23]](#ref-23)。官方 cron 指南的立场是：每个周期发生都有交代——tick 先推进 `next_run_at` 再派发（至多一次语义），错过窗口的补跑规则写清[[23]](#ref-23)。

委托方面，`delegate_task` 的模型可见参数有限（`tasks[]{goal, context, output_schema, images, group}`、`action`、`subagent_id`、`message`），`role` 被忽略、`goal` 是子 Agent 首条 user 消息[[14]](#ref-14)。子 Agent 默认禁用 `delegate_task/clarify/memory/send_message/cronjob_manage`，审批在 worker 线程里非交互解析（默认自动拒绝，`delegation.subagent_auto_approve: true` 才自动批准一次）[[14]](#ref-14)。需要“必须跨重启存活”的工作，官方文档明确：后台 `delegate_task` 是进程内的，重启会中断，应改用 cron job 或 Kanban[[14]](#ref-14)。

### 22.5.7 常见问题诊断表

| 现象 | 常见原因 | 对策 | 依据 |
| --- | --- | --- | --- |
| 重启后回复重复 | 崩溃遗留回复补发带 RECOVERED_MARKER | 属于“带重复提示的有界恢复”，看标记判断是否重复 | 22.4.6 |
| 会话里写入 memory 但下一轮看不到 | 冻结快照保护缓存前缀 | 设计如此，下个会话生效 | 22.3.7 |
| 压缩反复触发 | 小窗口模型阈值被抬到 0.75 | 理解有效触发点，用 `/new` 或换大窗口 | 22.3.5 |
| 模型报错不换模型 | 显式用户选择的模型是严格的 | 用 default primary + fallbacks | 22.4.3 |
| 无人值守命令被 BLOCKED | `cron_mode`/`unattended_mode` 默认 deny | 显式配置 approve，或走容器后端 | 22.5.2 |
| `/stop` 后子 agent 还在跑 | 后台委派脱离了 `_active_children` | 用 `interrupt_for_session` 或等回流 | 22.4.8 |
| 会话被 suspended | stuck-loop 连续 3 次重启 | 检查 `.restart_failure_counts`，修复根因 | 22.4.5 |
| write 被直接拒绝 | 受保护路径 / `HERMES_WRITE_SAFE_ROOT` | 路径层保护，在本节给定宿主和执行通道条件下按策略检查审批 | 22.5.3 |

凭据与环境的几个操作细节值得记下（官方说明）。MCP stdio 子进程只继承 PATH/HOME/USER/LANG/LC_ALL/TERM/SHELL/TMPDIR 与 XDG_*，其余（API key、token、secret）被剥离；`docker_forward_env` 加入的变量会注入容器，容器内代码可读取并外泄，Hermes 基础设施密钥不应加入[[29]](#ref-29)。context 文件（AGENTS.md、SOUL.md 等）在进入 system prompt 前被扫描（忽略指令、隐藏 HTML 注释、读取 secrets、curl 外泄、不可见 Unicode），命中则 `[BLOCKED: … Content not loaded.]`；用户自写的 SOUL.md 命中只警告，profile distribution 带来的命中仍阻断[[29]](#ref-29)。SSRF 防护拒绝 RFC1918、回环、链路本地、云元数据主机名/IP，DNS 失败 fail-closed，重定向逐跳重验[[29]](#ref-29)。

生产 gateway 官方清单是：显式 allowlist、容器后端、限资源、`.env` 权限、启用 DM pairing、审计 command_allowlist、设 terminal.cwd、不以 root 运行、监控 `~/.hermes/logs/`、定期 `hermes update`；`chmod 600 ~/.hermes/.env`，不同服务用不同 key，`.env` 不入版本库[[29]](#ref-29)。官方还给了最大隔离姿态：gateway 与执行分机，`terminal.backend: ssh`，主机细节写 `.env`（`TERMINAL_SSH_HOST/USER/KEY`），不随 profile 导出共享[[29]](#ref-29)。

技能与记忆的写入门默认关闭（`memory.write_approval`/`skills.write_approval` 默认 False），这是把“模型直接改长期状态”变成可选能力；技能侧另有 Skills Guard 信任分级与安装策略、curator 的归档（默认 7 天一次、14 天 stale、30 天 archive、只归档不删除、pinned 豁免）与 `skill_ledger` 审计回滚[[12]](#ref-12)。官方对 skill 分发的层级是：bundled `skills/`（人人需要）→ `optional-skills/`（官方、默认不激活）→ Skills Hub（社区，安装前审查）[[31]](#ref-31)。

几条操作层面的官方提醒值得补全。TLS：通过 truststore 使用平台信任库，`ssl_verify: false` 不安全，官方明说“不要把它当作缺失企业根证书的永久修复”[[29]](#ref-29)。gateway event hooks 目录 `<profile home>/hooks/<name>/` 以放置即启用、进程内以 gateway 权限导入，不受 `plugins.enabled` 与 `HERMES_SAFE_MODE` 约束——应像审阅 config.yaml 一样审阅 handler.py[[29]](#ref-29)。供应链公告扫描由 `hermes doctor` 展示，`hermes doctor --ack <advisory-id>` 持久化到 `config.security.acked_advisories`；`security.allow_lazy_installs` 可置 false 关闭按需安装[[29]](#ref-29)。

无 provider 配置时的首启保护与 `hermes doctor` 也值得写进日常工作流：`hermes doctor` 对配置/依赖做诊断，退出码 0 表示没有未解决问题、1 表示至少还有一个（含 `--fix` 修不好的），因此可作健康门禁/CI 步骤；网关仍在运行时不要运行 `hermes doctor --fix`[[33]](#ref-33)。

### 22.5.8 案例：给一个 Telegram 群挂一个受限助手

下面是一个教学构造的案例，未实测，只使用本章已核对过的入口与配置。目标：在 Telegram 群挂一个助手，只读文档、不执行危险命令、不写宿主文件。

| 步骤 | 做法 | 依据 |
| --- | --- | --- |
| 隔离 | terminal backend 用 `docker`（无宿主路径挂载），危险命令检查被跳过是因为容器本身是边界；容器资源设上限 | 22.5.1、22.5.2 |
| 身份 | Telegram 配置 allowlist，不设 allow-all；陌生用户走 DM pairing | 22.5.4 |
| 命令 | `approvals.deny` 加 rm/chmod/mkfs/dd 等破坏类；`approvals.mode` 保持 smart；hardline 保持默认 | 22.5.2 |
| 文件 | 不设 `HERMES_WRITE_SAFE_ROOT` 到项目目录；写保护保持默认 | 22.5.3 |
| 记忆/技能 | 本案例显式设 `memory.write_approval:true` 与 `skills.write_approval:true`；群与后台写入均暂存审核，技能仅载入审阅版本 | 22.5.5 |
| 审计 | 跑 `hermes doctor`，审阅 `~/.hermes/logs/`，定期 `hermes update` | 22.5.4 |

案例从把默认从“个人受信运营者的全权限助手”收窄到“受限群助手”：OS 隔离是第一层，身份 allowlist 决定谁能触发，deny 规则决定哪些命令不能跑，写保护决定能碰哪些文件，记忆/技能的自动写入保持受控，审批开关不得沿用默认关闭，审计把这一切变成可复查的状态。

## 22.6 调优：控制上下文、成本与质量

Hermes 有完整的旋钮与默认值，但**官方没有给出基准、本章没有实测**。这一部分只写“旋钮与机制、作用对象、观测手段”，不写“调到多少好”；数字都是固定提交里的默认值。官方对性能的数字声明（如提示缓存“~75% 节省”、MoA HermesBench 分数）范围不明，只作为厂商说法引用。

受限案例中，DM pairing 只回答谁能开始私聊，群授权仍需单独配置并与数据范围匹配。容器只有在不挂载敏感宿主路径、不过度暴露凭据/网络、设置资源限制时才形成预期边界；容器内选择较少交互审批可以是运营者的显式取舍，不能泛化为共享宿主上也可关闭控制。以上是本书案例前提，不是产品默认安全保证。



### 22.6.1 度量与观测

用量与成本的观测在会话行里：`sessions` 表有 `input/output/cache_read/cache_write/reasoning_tokens`、`estimated_cost_usd`、`actual_cost_usd`、`cost_status` 等列[[7]](#ref-7)。模型选择与 `/model`、`/usage`、`/logs`、`/dump`、`/insights`、`/prompt-size` 等命令提供入口（文档未给 `/cost` 命令）[[31]](#ref-31)。上下文拆解靠 `/context`（分 token 列出上下文与工具定义）。缓存可观测靠 usage 的 `cache_read_tokens/cache_write_tokens`，客户端构造缓存标记不证明服务端命中[[7]](#ref-7)。

### 22.6.2 上下文压缩与预算

压缩旋钮集中在 `compression.*`：`threshold`（默认 0.50，小窗口抬到 0.75）、`protect_last_n` 20、`target_ratio` 0.20、`max_attempts` 3、`keep_recent_tokens` 20000，以及 Gateway 侧独立的 85% hygiene 阈值[[13]](#ref-13)。注入预算有三块：MEMORY 2200 / USER 1375 字符；项目 context 文件上限 20000 字符 floor、0.06×窗口×4 字符、500000 ceiling（头 70%/尾 20%）；技能索引无硬上限[[12]](#ref-12)。工具结果预算三层（单结果 100,000、每轮 200,000、预览 1,500 字符）并按窗口缩放[[18]](#ref-18)。

### 22.6.3 提示缓存

缓存布局默认最多 4 个断点（静态 system 前缀 + system 末尾 + 最后 2 条消息）；`prompt_caching.cache_ttl` 默认 5 分钟，`auto` TTL 按会话来源在 1h/5m 间选[[4]](#ref-4)。换模型必然开启新缓存谱系，改 thinking/推理级别也可能失效。AGENTS.md 的贡献不变量值得使用者记住：每会话 prompt caching “神圣”——会话中途改历史、换 toolset、重载 memory、重建 system prompt 会破坏缓存前缀并增加成本，唯一例外是压缩；改 system prompt 状态的斜杠命令推迟到下个会话（`--now` 才立即）[[1]](#ref-1)。

### 22.6.4 模型与 fallback

模型/provider 选择经 `resolve_runtime_provider` 梯子；fallback 用 `fallback_providers`（列表，每项需 provider+model），沿链前进，显式用户选择严格不回退；主 provider 冷却 60 秒 → 4 小时[[16]](#ref-16)。`fallback.min_switch_reset_seconds`（默认 0=关闭）让重置时间很短的限流主 provider 不切 fallback、由重试退避等过窗口[[16]](#ref-16)。子 Agent 的 fallback 用 `delegation.fallback_providers`（未固定的子 Agent 继承父链）[[14]](#ref-14)。

### 22.6.5 工具集裁剪与委派成本

官方对“裁剪工具集节省多少 token”没有给数字，只给机制：`hermes tools` 启停平台 toolset、`tool_search` 按需加载工具描述、`code_mode` 让模型经单一入口调用工具[[31]](#ref-31)。AGENTS.md 的不变量是“每个模型工具每次 API 调用都会发送”，所以新增核心工具门槛高——这也是 22.5.5 里“能用 skill 就别用 tool”的 token 依据[[1]](#ref-1)。委派成本方面，`delegation.max_concurrent_children` 默认 10、`max_iterations` 250、`max_summary_chars` 24000；子 Agent 用更便宜的模型可经 profile 路由配置[[14]](#ref-14)。MoA 的 reference 并发上限 8，缓存键 `preset+SHA256(advisory view)+labels`[[14]](#ref-14)。

压缩触发点有一个容易误算的细节：`compression.threshold` 默认 0.50，但窗口 <512K 的模型被只升不降地抬到 ≥0.75——所以 200K 窗口的实际触发点约 150,000，不是文档示例的 100,000[[13]](#ref-13)。注入预算的精确数字：MEMORY 2200 / USER 1375 字符；项目 context 文件上限 20000 字符 floor、0.06×窗口×4 字符、500,000 ceiling（头 70%/尾 20%）；技能索引无硬上限，但每个技能约 97 字符加 name/description/location 长度[[12]](#ref-12)。

工具集裁剪的机制是 `hermes tools` 按平台启停、`tool_search` 按需加载工具描述、`code_mode` 让模型经单一入口调用工具；官方没有给“裁剪节省多少 token”的数字[[31]](#ref-31)。委派成本旋钮：`delegation.max_concurrent_children` 默认 10、`max_iterations` 250、`max_summary_chars` 24000（子摘要受父上下文剩余余量动态预算双重约束，裁掉的文本溢出到磁盘并给 head+tail 窗口）[[14]](#ref-14)。MoA 的 reference 并发上限 8、advisory view 预算 4000 字符、缓存键 `preset+SHA256(advisory view)+labels`、`fanout` 三档[[14]](#ref-14)。

可观测性方面，除会话行 usage/cost 列外，`/usage`、`/context`、`/prompt-size`、`/dump`、`/insights`、`/logs` 提供入口；日志分 `agent.log`（默认 5 MB×3 备份）、`errors.log`、`gateway.log`；仓内没有 langfuse/otel 插件（官方 Langfuse 是外部目录插件），也没有 `/cost` 命令[[31]](#ref-31)。官方对性能的数字声明（提示缓存“~75% 节省”、MoA HermesBench 分数）范围不明，只作为厂商说法引用，不作基准[[4]](#ref-4)。

### 22.6.6 不应为省 Token 放弃的东西

有几类开销不能省（本书建议）。命令审批：降低打断要靠容器后端、deny 规则和 allowlist，不能靠把 `approvals.mode` 改成 `off`——hardline 之外的地板是安全语义，不是性能。投递账本：`gateway.delivery_ledger: false` 关掉的代价是崩溃遗留回复从“补发”退化成“重新执行”，可能重复副作用。工具前持久化（persist-before-execute）：关掉它省一次写，但破坏“resume must see the executed block”的耐久性不变量。缓存保温可以关，逐动作的权限边界和恢复语义不能为省 token 而放宽。

### 22.6.7 从故障轨迹判断长期运行是否成立

#### 一条输入跨越两种持久责任

下面是本书构造的验收轨迹，不是 Hermes 的实测记录。一个获准群成员要求分析告警并回复结论，Gateway 将输入保存到 authority 管理的 admission，再认领 worker。模型读取授权指标，形成带证据的结果，回复随后进入 delivery obligation。前一条责任负责“输入有没有推进”，后一条责任负责“结果有没有投递”；中间还可能有独立的工具效果和会话提交。

在 admission 仍 queued 时终止进程，恢复可以依据持久责任安排执行；在 started 后终止，旧 owner 的工作可能已产生副作用，因此恢复为 unknown 不意味着按 queued 重试；在结果已保存、投递标记 pending 时终止，恢复的是发送责任；在 attempting 时终止，接收端可能已收到，补发必须保留重复提示。这四个位置用相同的“重试任务”按钮处理，会把已经完成的工作重复执行。

验收记录应包含事件顺序、会话 ID、输入 ID、worker 身份、效果标记、发送状态和日志引用。把进程杀死再启动，只是故障注入动作；是否满足保证取决于这些事实的核对。若运行环境没有权限进行故障注入，可以做源码路径审查并清楚标明未运行，但不能把审查结论称为实测恢复率。

#### 标记与年龄阈值怎样配合

恢复标记用于识别“这个状态来自什么中断与接续”，年龄阈值用于决定“何时值得扫描、何时停止自动尝试”。一个较老条目可能仍有真实远端效果，一个较新条目也可能已失去 owner。二者解决不同问题，所以“有标记”与 `max(3600,2×timeout)` 等策略可以同时存在；错误在于把年龄本身当作原动作未发生的证明。

例如一条自动化工作记录超过检查年龄，系统可以将它交给恢复扫描；扫描看到旧 epoch、已开始标记和缺失回执，再按实现结算 unknown/interrupted。这不是等待一小时就安全重放的协议。相反，即使任务只开始了数秒，只要确认原 owner 失效且外部状态未知，也不能跳过核验。具体阈值与清理条件只在固定版本范围内解释。

同理，投递义务超过时限进入 abandoned，表示自动处理责任已经停止，不代表消息“必然没到”或“已经送达”。告警应让运营者看到原目标、最后一次尝试和可用回执，再决定是否联系用户或人工重发。过期清理会进一步减少可追查证据，保留期应与业务可接受的发现延迟匹配，不能只按数据库行数优化。

#### Kanban 的持久性与单主机边界

Kanban 把待办与执行状态持久保存，使工作不必依赖某个对话页面一直打开；熔断和停止条件有助于限制反复失败。但单机数据库、文件锁与本机 worker 组合不自动形成多主机一致的调度系统。复制同一个工作目录到两台机器并同时启动，可能让租约、凭据、文件与后台任务产生新的共享状态问题。

本书的部署建议是先确定唯一调度权威，其他入口通过受控接口提交；确需多机时，先验证产品是否提供跨主机的所有权和存储契约。不能因为 SQLite 支持 WAL 就推导网络文件系统上的任意并发安全，也不能把 turn lease 当作全系统每种后台作业的通用分布式锁。具体数据库运行要求仍遵循 SQLite 与目标平台的官方说明。

Kanban 熔断还应区分故障类别。依赖服务暂不可用可以有界重试，权限撤销应停止，原动作未知应查询，学习内容被投毒应隔离。所有失败都统一增加重试次数，只会放大副作用与成本。恢复面板要同时展示被阻断原因、原动作引用、责任人和期限，使“自动化暂停”成为可接续的工作状态。

#### MoA、学习与成本不能只看单轮结果

多模型或多路径聚合可能在某些任务上改善答案，但也增加调用、上下文复制和中间结果信任边界。比较 MoA 时，应把候选模型、提示、聚合器、工具权限和总预算作为一组固定实验配置。某个子路径读取的内容，不应因聚合而扩散到无权用户；聚合结果中引用的证据也要保留原来源，不能把多个模型意见当作独立外部事实。

学习循环的价值同样要跨时间衡量。某条技能让当前任务更快，不代表它对未来版本正确；反复复用可能让错误逐渐成为默认。评估应同时包含未学习基线、审阅后版本、后续留出任务和污染反例，检查成功率、误触发、权限违例与纠错成本。这样才能区分方法改进和对训练过任务的记忆。

后台反思消耗也属于任务总成本。前台模型费用下降，但新增多轮复盘和技能验证，单位成功任务成本可能上升。对无权或无价值候选，尽早拒绝持久化可以同时降低污染与成本；对有价值但高风险技能，额外审阅是明确的安全投入。这里不提供没有负载与测量条件的统一节省比例。

#### 配置变更怎样影响在途会话

修改 memory、system prompt 或 toolset 会改变下一轮可见输入；为保持缓存而延后应用某些变更，可能让活动会话继续使用旧材料。安全撤权和事实纠错不能因为缓存代价而延后。部署方需要区分优化性配置变更与必须立即生效的授权变更，后者要停止受影响派发、失效派生内容，必要时重建会话。

恢复或升级时，不要把整份旧配置复制为永久授权。原审批可能过期，工具实现可能变化，后台任务的账户也可能被收回。保存配置快照用于解释历史，当前继续执行仍需重验。若目标版本没有足够细的动态撤权机制，可以采用停止服务、收紧网络/OS 权限和人工核验的运维措施，并把可用性代价写清楚。

#### 一份可以接手的未决清单

长期运行的最低要求是责任可以被发现和接续。未决项至少记录原输入、当前会话、已开始的 worker、外部效果状态、投递状态、最后证据、当前权限、负责人和期限。它应该能回答“下一步查哪个原对象”，而不是只留下“Agent 报错”。对账得出新事实后，保存依据与处置，不直接修改历史声称原本就成功。

人工核验也不能无限期悬空。先按影响收窄新动作，明确查询入口和升级链；无法确定时向用户报告事实边界。若选择放弃重投，要说明原回复可能丢失；若选择重复发送，要说明可能重复；若选择重新执行，先排除旧副作用或取得明确的新业务授权。用这些具体选择结束一个未决项，才完成长期运行的恢复责任。

#### 受限学习助手的审阅和回滚实例

假设群成员在一次排障后建议“把这次方法保存下来”。这句话只授权讨论候选方法，不能自动授权把整段群聊、个人信息或凭据写入共享技能。助手先给出候选：适用服务和版本、输入条件、只读取证步骤、完成证据、禁止动作，以及它来自哪次经过核验的任务。审阅者可以检查候选是否把相关性误写成因果、是否漏掉当前事件才成立的条件，再决定是否发布。

例如本次发现连接等待增加，候选可以写成“先查询连接池等待与请求量，再检查对应时间窗的慢查询”，但不能写成“延迟高就扩大连接池”。后一句把一种处置泛化成默认行动，既缺容量前提，也改变了只读边界。即使扫描器没有发现危险模式，也应因证据和适用范围不足而拒绝发布。技能审阅因此同时检查事实、方法与权限，不只检查文件格式。

发布后给能力包一个不可变版本，并把本次通过的用例和必要反例关联起来。后续发现该方法对另一数据库版本无效时，先停止新任务选择受影响版本，通知当前使用者，再决定修订或回滚。正在运行的任务可能已经加载旧规则，不能只修改磁盘文件便认为已全部切换；对高风险影响需要停止受影响派发并重新装配上下文。已经发生的外部效果沿原身份另行核验。

这里的回滚操作只改变下一次选择的方法，不倒退业务世界。若旧技能已经创建工单，恢复上一版不能删除工单；若旧技能把无权资料发进群，删除技能也不能收回所有副本。应将数据事件、业务副作用和能力修复分别处理，并记录负责人。这样的例子把学习闭环从“写得进去”延伸到“错了以后如何纠正”，才符合长期运行助手的维护需求。

#### 用户可控的记忆不应只有一个删除按钮

用户可以要求查看或删除自己的记忆，但实现需要先确定“自己的”包括哪些来源与派生物。个人偏好、群共享事实和运营策略拥有不同权威；一个成员不应删除其他成员的独立记忆，也不能用更新个人偏好的请求覆盖全群安全规则。若产品只提供较粗的命名空间，应向运营者说明限制，并通过独立 profile 或存储边界满足隔离目标。

删除后还要处理派生摘要、已生成技能、缓存与活动上下文。保留审计记录时只保存必要的标识和处理事实，不为追溯而继续保存已要求删除的敏感正文。权威记录可以保留“某记忆已撤回”的墓碑，使后续导入旧备份时不把它复活；若目标产品没有这项能力，外部维护流程必须接住，而不能把本书建议写成产品已有保证。

最后，用户界面应区分自动学到的候选、已经审阅的共享方法和运营者强制策略。三者都可能以文字出现，但拥有不同修改与执行权限。模型可以建议修改候选，不能自行晋升权限或决定删除审计事实。将这三个入口分清，既有利于用户理解，也能让后续安全评审定位真实的写入门。

### 22.6.8 版本敏感事实与验收边界

下面逐项绑定 `907e3224c432`；相关固定文件核验于 2026-10-24，未运行服务。目录条目数、模块行数只帮助定位，不作为能力或性能基准。

| 主张 | 固定值与条件 | 证据与保证边界 |
| --- | --- | --- |
| 会话 schema | SCHEMA_VERSION=32；文档另写 31 | 源码事实，schema 常量[[34]](#ref-34)；文档或安装版本自动等于该提交 |
| durable admission 恢复 | 旧 epoch 的 started 工作转 unknown，不自动重放 | 源码事实，输入恢复函数[[34]](#ref-34)；unknown 等于远端未执行 |
| 自动 orphan turn 恢复 | max_age=max(3600, 2×timeout)；超龄标记清除而不执行 | 源码事实，Gateway 孤儿回合恢复路径[[24]](#ref-24)；只凭新时间戳就有恢复资格 |
| 投递重试边界 | MAX_ATTEMPTS=3，STALE_AFTER_SECONDS=24h；清理保留7天/最多500行 | 源码事实，投递账本常量与启动清扫函数[[26]](#ref-26)；每个最终回复必达或所有丢失永久可查 |
| 内存与技能写门 | 两者 write_approval 默认 false；memory 在交互前台询问、其他入口暂存；skill 写一律暂存 | 官方文档事实，固定 Memory 文档的两个 write_approval 章节[[12]](#ref-12)；开门后其他插件或 terminal 自动受同一门控制 |
| 技能回滚 | skill_ledger 记录修改与单次回滚，rollback_entry 的安全捕获失败会阻断 | 源码事实，技能账本回滚入口[[34]](#ref-34)；内建文件 memory 有同样完整账本，或回滚技能撤销既有动作 |

中断是否生效还需按 HTTP、顺序/并行工具与后台委派逐项验收，见 22.4.2；长期状态的来源与污染处理见 22.5。产品自述、源码注释中的测量和本书威胁构造分别标注，不能互相代替实验。

## 本章小结

Hermes 把“长期运行、自我改进”的个人 Agent 做成了一套有明确权威关系的系统：Gateway 是 canonical session 的唯一写者，AIAgent 负责一次回合的模型与工具循环，记忆、技能与后台复盘在两者之间形成学习闭环。它的工程重心有两个：一是“长期在线”带来的可靠性（跨进程回合租约、persist-before-execute、重启恢复的标记机制、投递账本的带重复提示的有界恢复、Kanban 的 claim 与熔断），二是“能替人行动”带来的安全（八层安全模型、审批 fail-closed、hardline 地板、文件写保护、OS 级隔离）。

这两者之间有一条清晰的主线：Hermes 反复强调“唯一的安全边界是操作系统”，进程内的审批、脱敏、allowlist 都是启发式。所以使用者真正要做的，是把默认从“个人受信运营者”收窄到具体场景——用容器后端做隔离，用 allowlist 和 DM pairing 做身份，用 deny 规则和写保护做边界，用 `hermes doctor` 和日志做审计。

调优方面，Hermes 的旋钮（压缩阈值与注入预算、提示缓存布局、fallback 链、工具集裁剪、委派并发）和观测手段（会话行的 usage 与 cost 列、`/context`、`/usage`）都很完整，但官方没有给基准，本章因此只讲机制与默认值，把“调到多少好”留给使用者用自己的工作负载去测。

## 参考资料

官方文档页与源码都固定到提交 `907e3224c432647567cc4dd7aebcdf258971896a`，核验日期 2026-10-23。文档与源码同提交内冲突处分别说明适用范围，不推定已安装二进制行为。

1. <a id="ref-1"></a> NousResearch / Hermes maintainers，[README](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/README.md)。
2. <a id="ref-2"></a> NousResearch / Hermes maintainers，[toolsets.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/toolsets.py)、[model_tools.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/model_tools.py)、[tools/registry.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/registry.py)。
3. <a id="ref-3"></a> NousResearch / Hermes maintainers，[run_agent.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/run_agent.py)、[agent/agent_init.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/agent_init.py)、[agent/turn_facade.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/turn_facade.py)、[hermes_cli/cli_agent_setup_mixin.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/cli_agent_setup_mixin.py)、[hermes_cli/gateway_agent.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/gateway_agent.py)。
4. <a id="ref-4"></a> NousResearch / Hermes maintainers，[agent/system_prompt.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/system_prompt.py)、[agent/prompt_caching.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/prompt_caching.py)、[agent/prompt_cache_scope.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/prompt_cache_scope.py)、[website/docs/developer-guide/prompt-assembly.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/developer-guide/prompt-assembly.md)。
5. <a id="ref-5"></a> NousResearch / Hermes maintainers，[website/docs/user-guide/configuration.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/user-guide/configuration.md)、[website/docs/developer-guide/session-storage.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/developer-guide/session-storage.md)。
6. <a id="ref-6"></a> NousResearch / Hermes maintainers，[pyproject.toml](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/pyproject.toml)。
7. <a id="ref-7"></a> NousResearch / Hermes maintainers，[hermes_state.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_state.py)、[hermes_state_common.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_state_common.py)、[hermes_state_runtime.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_state_runtime.py)、[website/docs/developer-guide/session-storage.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/developer-guide/session-storage.md)。
8. <a id="ref-8"></a> NousResearch / Hermes maintainers，[hermes_cli/runtime_provider.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/runtime_provider.py)、[hermes_cli/cli_init_mixin.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/cli_init_mixin.py)、[website/docs/developer-guide/provider-runtime.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/developer-guide/provider-runtime.md)。
9. <a id="ref-9"></a> NousResearch / Hermes maintainers，[website/docs/developer-guide/multiplexing-gateway.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/developer-guide/multiplexing-gateway.md)、[hermes_cli/gateway_multiplex_mode.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/gateway_multiplex_mode.py)。
10. <a id="ref-10"></a> NousResearch / Hermes maintainers，[hermes_cli/main.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/main.py)。
11. <a id="ref-11"></a> NousResearch / Hermes maintainers，[hermes_cli/kanban_db.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/kanban_db.py)、[hermes_cli/kanban_db_dispatch.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/kanban_db_dispatch.py)、[gateway/kanban_watchers.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/kanban_watchers.py)、[tools/kanban_tools.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/kanban_tools.py)、[website/docs/user-guide/features/kanban.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/user-guide/features/kanban.md)。
12. <a id="ref-12"></a> NousResearch / Hermes maintainers，[tools/memory_tool_store.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/memory_tool_store.py)、[agent/memory_manager.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/memory_manager.py)、[agent/memory_provider.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/memory_provider.py)、[website/docs/user-guide/features/memory.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/user-guide/features/memory.md)。
13. <a id="ref-13"></a> NousResearch / Hermes maintainers，[agent/context_compressor.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/context_compressor.py)、[website/docs/developer-guide/context-compression-and-caching.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/developer-guide/context-compression-and-caching.md)、[website/docs/developer-guide/gateway-session-lifecycle.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/developer-guide/gateway-session-lifecycle.md)。
14. <a id="ref-14"></a> NousResearch / Hermes maintainers，[tools/delegate_tool.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/delegate_tool.py)、[tools/delegate_tool_config.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/delegate_tool_config.py)、[tools/delegate_tool_child_run.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/delegate_tool_child_run.py)、[agent/moa_loop.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/moa_loop.py)、[website/docs/user-guide/features/delegation.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/user-guide/features/delegation.md)。
15. <a id="ref-15"></a> NousResearch / Hermes maintainers，[batch_runner.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/batch_runner.py)、[trajectory_compressor.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/trajectory_compressor.py)、[website/docs/developer-guide/trajectory-format.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/developer-guide/trajectory-format.md)。
16. <a id="ref-16"></a> NousResearch / Hermes maintainers，[agent/interrupt_control.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/interrupt_control.py)、[agent/error_classifier.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/error_classifier.py)、[agent/turn_api_error.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/turn_api_error.py)、[agent/retry_utils.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/retry_utils.py)、[agent/credential_pool.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/credential_pool.py)、[agent/agent_init_fallback.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/agent_init_fallback.py)、[agent/turn_facade_lease.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/turn_facade_lease.py)、[agent/replay_cleanup.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/replay_cleanup.py)。
17. <a id="ref-17"></a> NousResearch / Hermes maintainers，[agent/conversation_loop.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/conversation_loop.py)、[agent/turn_tool_round.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/turn_tool_round.py)、[agent/turn_api_call.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/turn_api_call.py)、[website/docs/developer-guide/agent-loop.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/developer-guide/agent-loop.md)。
18. <a id="ref-18"></a> NousResearch / Hermes maintainers，[agent/tool_executor.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/tool_executor.py)、[agent/tool_guardrails.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/tool_guardrails.py)、[tools/tool_result_storage.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/tool_result_storage.py)、[tools/budget_config.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/budget_config.py)。
19. <a id="ref-19"></a> NousResearch / Hermes maintainers，[gateway/run_inbound.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/run_inbound.py)、[gateway/run_turn.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/run_turn.py)、[gateway/run_turn_runner.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/run_turn_runner.py)。
20. <a id="ref-20"></a> NousResearch / Hermes maintainers，[gateway/session.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/session.py)、[gateway/session_execution.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/session_execution.py)。
21. <a id="ref-21"></a> NousResearch / Hermes maintainers，[website/docs/developer-guide/gateway-internals.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/developer-guide/gateway-internals.md)。
22. <a id="ref-22"></a> NousResearch / Hermes maintainers，[tools/terminal_tool.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/terminal_tool.py)、[tools/terminal_tool_lifecycle.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/terminal_tool_lifecycle.py)、[tools/environments/base.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/environments/base.py)、[website/docs/user-guide/features/tools.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/user-guide/features/tools.md)。
23. <a id="ref-23"></a> NousResearch / Hermes maintainers，[gateway/run.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/run.py)、[gateway/run_idle_exit.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/run_idle_exit.py)、[hermes_cli/gateway_runtime.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/gateway_runtime.py)、[hermes_cli/gateway_runtime_start.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/gateway_runtime_start.py)、[cron/scheduler.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/cron/scheduler.py)、[cron/AGENTS.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/cron/AGENTS.md)。
24. <a id="ref-24"></a> NousResearch / Hermes maintainers，[gateway/run_bootstrap.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/run_bootstrap.py)、[gateway/run_startup.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/run_startup.py)、[gateway/run_shutdown.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/run_shutdown.py)、[gateway/run_runtime.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/run_runtime.py)、[gateway/run_adapters.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/run_adapters.py)。
25. <a id="ref-25"></a> NousResearch / Hermes maintainers，[hermes_startup_watchdog.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_startup_watchdog.py)、[gateway/shutdown_watchdog.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/shutdown_watchdog.py)。
26. <a id="ref-26"></a> NousResearch / Hermes maintainers，[gateway/delivery_ledger.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/gateway/delivery_ledger.py)。
27. <a id="ref-27"></a> NousResearch / Hermes maintainers，[tools/process_registry.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/process_registry.py)、[tools/process_registry_checkpoint.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/process_registry_checkpoint.py)。
28. <a id="ref-28"></a> NousResearch / Hermes maintainers，[tools/approval.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/approval.py)、[tools/approval_detection.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/approval_detection.py)、[tools/approval_smart.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/approval_smart.py)、[agent/file_safety.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/agent/file_safety.py)、[tools/url_safety.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/url_safety.py)、[tools/path_security.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/path_security.py)。
29. <a id="ref-29"></a> NousResearch / Hermes maintainers，[website/docs/user-guide/security.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/user-guide/security.md)。
30. <a id="ref-30"></a> NousResearch / Hermes maintainers，[SECURITY.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/SECURITY.md)。
31. <a id="ref-31"></a> NousResearch / Hermes maintainers，[AGENTS.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/AGENTS.md)、[CONTRIBUTING.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/CONTRIBUTING.md)、[website/docs/reference/cli-commands.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/reference/cli-commands.md)、[website/docs/user-guide/features/skills.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/user-guide/features/skills.md)。
32. <a id="ref-32"></a> NousResearch / Hermes maintainers，[tools/mcp_tool.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/mcp_tool.py)、[tools/mcp_tool_discovery.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/mcp_tool_discovery.py)、[website/docs/user-guide/features/mcp.md](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/website/docs/user-guide/features/mcp.md)。
33. <a id="ref-33"></a> NousResearch / Hermes maintainers，[hermes_cli/plugins_loader.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/plugins_loader.py)、[hermes_cli/config_defaults.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/config_defaults.py)、[hermes_cli/config.py](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_cli/config.py)。

34. <a id="ref-34"></a> NousResearch / Hermes maintainers，[schema 常量](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_state_common.py#L289)、[recover_session_inputs](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/hermes_state_runtime.py#L447)、[skill_ledger](https://github.com/NousResearch/hermes-agent/blob/907e3224c432647567cc4dd7aebcdf258971896a/tools/skill_ledger.py)，固定提交；核验于 2026-10-24。
