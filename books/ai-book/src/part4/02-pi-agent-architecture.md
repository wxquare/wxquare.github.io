# 第 20 章 Pi Agent 系统：终端原生 Coding Agent Runtime 的实现、使用与调优

> Pi 的核心价值不是“又一个终端里的 AI 编程助手”，而是把 Coding Agent 做成一个小内核、强扩展、可嵌入的 Agent Harness：模型调用、工具执行、上下文构建、会话树、扩展系统、Skills、Prompt Templates 和 SDK 都围绕同一套运行时组织起来。也正因为内核克制，安全边界、逐次审批、沙箱这类能力要靠使用者和嵌入方自己补上。

**资料核验日期：2026-10-23。** 本章分析的对象是 Pi（npm 包 `@earendil-works/pi-coding-agent`，版本 1.4.0，发布于 2026-10-21）。源码固定在仓库 `earendil-works/pi` 的主分支提交 `936e5f40aede9327b9142b63506ed2b68bc6b4ac`（提交时间 2026-10-22 21:32:38 UTC）；官方文档站 pi.dev 是滚动页面，本章引用的文档页与固定提交里 `packages/coding-agent/docs/` 的文件绑定所列固定版本，因此文档类引用同时给出站点地址与固定提交副本，二者可以互相回查。

Pi 的公开程度让本章可以同时使用两种证据，并在正文中分开标注：

| 标记 | 含义 | 使用方式 |
| --- | --- | --- |
| 源码事实 | 固定提交中可定位的文件与符号 | 只引用 `936e5f40aede…` 这一提交，不引用主分支 |
| 官方文档事实 | 官方文档页与 README、CHANGELOG 的陈述 | 注明核验日期；文档写明版本门槛的保留 |
| 系统推导 | 根据公开证据做的架构解释 | 帮助理解，不等于产品的精确实现 |
| 本书建议 | 本书依据前述事实归纳的使用建议 | 只出现在实践与调优部分，与官方建议分开标注 |

官方自报数字属于官方文档事实的子类型，保留自述与条件，不作通用基准。

一条限制需要提前说明：本章的行为结论来自阅读源码与官方文档，没有运行验证、没有故障注入。正文里所有“行为是 A”的陈述，准确含义是“固定提交的源码（或官方文档）写了 A”，除非另有标注。文档与源码不一致的地方，分别说明该提交实现与文档承诺，不推定读者安装的二进制与任意主分支提交一致；只在源码里能找到、文档没有承诺的行为，标注“源码事实，非文档承诺”。命令名、配置键、默认值、模型名与别名都是版本敏感内容，一律注明核验日期，正文不硬写模型名。

## 20.1 组成：Pi 由哪些部分构成

### 20.1.1 定位与边界

Pi 官方对自己的描述是“a minimal, extensible agent harness that you can make your own”（一个最小、可扩展、可以由用户改造成自己想要的形态的 Agent Harness）[[29]](#ref-29)。文档首页把它定位为“在终端里工作的可扩展 AI Agent”，用途列了软件开发、研究笔记、写作和数据文件处理[[1]](#ref-1)；npm 包的描述则是“带 read、bash、edit、write 工具和会话管理的 Coding Agent CLI”[[46]](#ref-46)。综合起来看，“Coding Agent”的限定在文档层面已经比早期放宽，“终端原生、可扩展、可嵌入”才是更稳定的定位。

“小”不是功能少，而是核心职责克制。README 明确写了一句边界：Pi “skips features like sub-agents and plan mode”——不做子 Agent、不做内置 plan mode[[29]](#ref-29)。这两项能力在官方示例扩展里以另一种形式出现（见 20.5.8），但不在内核里。另一条更硬的边界是权限：Pi 官方安全页明确说它不逐次询问，工具以启动 Pi 的账户权限读写和执行[[6]](#ref-6)。这一条不是“缺一个功能”，而是整个安全模型的起点，20.1.6 和 20.4 会反复回到它。

本章要回答三个选型问题：哪些复杂度留在内核，哪些必须由扩展或宿主承担，以及会话恢复能否覆盖业务动作的恢复。后文按组成、启动、执行、异常、实践与调优六个方面检验这些取舍。

### 20.1.2 包结构与分层

Pi 的仓库是一个 npm workspaces monorepo，`packages/` 下有 11 个包，本核验日版本均为 1.4.0[[29]](#ref-29)[[30]](#ref-30)：

| 包 | 包名 | 职责 |
| --- | --- | --- |
| `agent` | `@earendil-works/pi-agent-core` | 通用 Agent 核心：转录状态、事件、工具执行、steering/follow-up 队列 |
| `ai` | `pi-ai` | 统一 LLM API：模型目录、provider、认证、流式事件、重试与溢出判定、消息变换 |
| `coding-agent` | `pi-coding-agent` | `AgentSession`、会话树持久化、资源加载、扩展宿主、内置工具、四种运行模式与 CLI |
| `tui` | `pi-tui` | 终端界面 |
| `codemode` | `pi-codemode` | 让模型写 JavaScript 在沙箱里执行的工具 |
| `mcp` | `pi-mcp` | MCP 客户端（stdio 与 Streamable HTTP） |
| `durable` | `pi-durable` | 长期对话、任务与文档运行时（README 标为 Experimental，API 随时变化） |
| `chord` | `@earendil-works/chord` | 底层协同/同步库 |
| `env` | `pi-env` | 环境隔离相关 |
| `telemetry` | `pi-telemetry` | 遥测 |
| `evals` | `pi-evals` | 私有评测包 |

依赖方向（取自各包 `package.json`）给出了一条清晰的分层[[30]](#ref-30)：`coding-agent` 依赖 `agent-core`、`ai`、`codemode`、`mcp`、`tui`；`agent-core` 依赖 `ai`；`ai` 依赖 `telemetry` 和各 provider SDK；`durable`、`chord`、`env` 之间另有依赖，但 `coding-agent` 不依赖 `durable`、`chord`、`env`。也就是说，本章讨论的这条“Coding Agent 主链路”自下而上是三层：

```text
pi-ai（统一 LLM API）
  → pi-agent-core（Agent 与 Agent Loop）
    → pi-coding-agent（AgentSession、资源加载、扩展宿主、内置工具、CLI）
```

代码规模可以对“小”做一个数量级上的校准（以下行数是本书用 `cat | wc -l` 统计、排除测试后的口径，不是官方数字，仅作参考）：agent-core 约 2,500 行，ai 约 27,000 行，coding-agent 约 76,000 行；coding-agent 内最大的文件是 `interactive-mode.ts`（约 7,100 行）和 `agent-session.ts`（约 4,500 行）。真正“小”的是最底层的 agent-core；越往上，会话、扩展、模式与 UI 的复杂度越集中[[30]](#ref-30)。

### 20.1.3 核心对象

运行时的骨架由六个对象构成，它们各自的权威职责可以在源码里定位。

`Agent`（agent-core 包）是“低层 agent loop 之上的有状态封装”：它拥有当前转录与工具、发出生命周期事件、执行工具，并暴露 steering 与 follow-up 的排队接口。对外的核心方法只有 `prompt`、`continue`、`steer`、`followUp`、`abort`、`waitForIdle`、`reset`、`subscribe`[[30]](#ref-30)。

`AgentSession`（coding-agent 包）是使用者真正接触的对象。官方 SDK 文档对它的一句话定义是：一个会话拥有“一段对话、它的模型与工具、排队消息、压缩状态和扩展运行时”[[11]](#ref-11)。它在构造时订阅底层 `Agent` 的事件用于持久化、扩展、自动压缩与重试，并安装五组钩子（工具执行前后、下一回合刷新、请求投影、回合边界、隐藏声明与强制提示投影），再构建工具注册表与系统提示[[31]](#ref-31)。

`SessionManager` 是“已定稿模型上下文的权威”[[11]](#ref-11)：它拥有持久化或内存的条目树并追踪活动叶子，提供静态工厂（`create`、`open`、`continueRecent`、`inMemory`、`forkFrom`、`findById`、`list`）、各类 `append*` 方法、树操作（`branch`、`branchWithSummary`、`resetLeaf`、`getBranch`）和上下文构造（`buildSessionProjection`、`buildSessionContext`）[[32]](#ref-32)。

`ModelRuntime` 是模型目录、可用性、认证解析与流式请求的统一入口。它支持运行时 API key、虚拟模型和扩展注册的 provider；`create` 默认不联网刷新目录（`allowNetwork` 默认为 false）[[36]](#ref-36)。

`DefaultResourceLoader` 发现并持有扩展、Skills、Prompt Templates、主题、上下文文件、`SYSTEM.md` 与 `APPEND_SYSTEM.md`；它的接口可由宿主完全自实现，这是 SDK 嵌入时替换资源来源的入口[[33]](#ref-33)。

`ExtensionRunner` 负责按扩展加载顺序分发事件、执行工具调用与结果钩子、注入上下文，并把处理器错误上报到统一通道[[38]](#ref-38)。

在这些对象之外还有一个 `AgentSessionRuntime`：它负责“替换活动会话”，`newSession()`、`switchSession()`、`fork()`、`importFromJsonl()` 每次替换都会先发 `session_before_*`，再 `session_shutdown`，最后用新的工作目录重建服务并创建新的 `AgentSession`。SDK 嵌入方如果要在同一进程里切换会话，需要经由它，而不是反复 `createAgentSession`[[31]](#ref-31)。

### 20.1.4 数据目录与文件格式

Pi 的状态分成两处：全局代理目录，和每个项目的 `.pi/` 目录。

代理目录默认是 `~/.pi/agent`，可由环境变量 `PI_CODING_AGENT_DIR` 或 SDK 的 `agentDir` 改变。里面放 `settings.json`、`keybindings.json`、`mcp.json`、`models.json`、`auth.json`、上下文文件（`AGENTS.override.md`、`AGENTS.md`、`CLAUDE.md` 等）、`SYSTEM.md`、`APPEND_SYSTEM.md`，以及 `extensions/`、`skills/`、`prompts/`、`themes/` 和 `sessions/`；另有运行时写入的 `trust.json`、`crashes.json`、`mcp.log`、`pi-debug.log`[[7]](#ref-7)。

项目目录用 `.pi/` 前缀：`.pi/settings.json`、`.pi/mcp.json`、`.pi/SYSTEM.md`、`.pi/APPEND_SYSTEM.md`、`.pi/extensions/`、`.pi/skills/`、`.pi/prompts/`、`.pi/themes/`，另外项目还支持 `.agents/skills`（从当前目录向上到祖先，有仓库根则止于仓库根）[[7]](#ref-7)[[19]](#ref-19)。上下文文件（AGENTS.md 等）与 `.pi` 配置分开：前者不需要项目信任，后者需要（见 20.2 的信任机制）[[6]](#ref-6)。

会话文件是 JSONL，每条一行一个条目，用 `id`/`parentId` 形成一棵树；文件头是 `type:"session"`、`version:3`、`cwd`，可选 `parentSession`。条目类型共 11 种：`message`、`thinking_level_change`、`model_change`、`usage`、`compaction`、`branch_summary`、`custom`、`custom_message`、`context_edit`、`label`、`session_info`。其中 `custom` 不进模型上下文，`custom_message` 进；`usage` 条目用于缓存预热这类非消息计费，不进上下文、在树中隐藏[[16]](#ref-16)[[32]](#ref-32)。

值得强调的是“权威状态”的组织方式。模型每次看到的上下文不是把文件顺序重放一遍，而是由活动分支重建：取最近一个 `compaction` 条目，从 `firstKeptEntryId` 起保留，再应用此后的 `context_edit`。系统提示与工具声明则作为 **system 消息**写入转录——首个请求持久化完整的 sections 与 `toolsAdded`，之后以按 section 名的补丁记录变化。官方文档的原话是：按顺序重放它们就得到当前的提示与工具，没有独立的“提示状态条目”[[16]](#ref-16)。这一点决定了 20.4 里“会话文件损坏时能恢复什么”的边界。

会话文件路径是 `<agent-dir>/sessions/--<cwd 去掉前导分隔符并把 / \ : 替换成 ->--/<时间戳>_<session-id>.jsonl`，`session-id` 默认 UUID；文件版本 v3（v1 线性 → v2 树 → v3 改名 `custom`），旧文件加载时自动迁移并重写整个文件[[16]](#ref-16)[[32]](#ref-32)。

`auth.json` 单独有更严的写保护：文件权限 `0o600`，并发写用 `proper-lockfile` 加锁；`models.json` 里的 `apiKey` 可以写 `$NAME`、字面量或以 `!` 开头的命令，命令在请求时执行且不缓存[[36]](#ref-36)[[22]](#ref-22)。

四个内置扩展（llama.cpp、codemode、tool-search、mcp）在设置里以 `builtin:mcp` 这类名字启停，`--no-mcp` 禁用内置 MCP；其中后三个标记 `replaceable: true`，扩展若注册同名工具/命令则取代内置实现[[9]](#ref-9)。codemode 值得一提：它让模型写 JavaScript 在 QuickJS（WebAssembly）沙箱里运行，沙箱没有 Node API、文件系统、网络和定时器，只能通过 `tools.<name>()` 访问外界，只有脚本输出进入模型上下文；默认**不启用**，限制是每脚本 VM 内存 256 MB、全进程最多 2 个 worker、等不到 worker 60 秒失败、`max_output_tokens` 默认 10000[[25]](#ref-25)。

### 20.1.5 模型接入层

`pi-ai` 把各家 API 归一成一套流式语义。已知 API 协议 10 种（`openai-completions`、`anthropic-messages`、`google-generative-ai` 等），README 的 Supported Providers 列了 35 项，源码 `src/providers/` 下有 40 余个 provider 模块；README 声明目录“只包含支持工具调用的模型”[[40]](#ref-40)。
已知 API 协议的十种名称是 `openai-completions`、`mistral-conversations`、`openai-responses`、`azure-openai-responses`、`openai-codex-responses`、`anthropic-messages`、`bedrock-converse-stream`、`google-generative-ai`、`google-vertex`、`pi-messages`；provider 模块数与模型目录随版本变化，本核验日分别为 40 余个 provider 模块与 42 个目录文件，正文不写死为接口承诺[[40]](#ref-40)。

统一的流式契约是：请求失败（含中止与工具调用校验失败）不抛异常，而是发出 `error` 事件，最终消息的 `stopReason` 为 `error` 或 `aborted` 并带部分内容与部分用量；`StopReason` 的取值是 `pending/stop/length/toolUse/error/aborted/deferred`。注意边界：直接调用 API 层 `streamSimple` 且缺认证时，会同步抛错，而不是走事件流[[40]](#ref-40)。

用量结构 `Usage` 记录 `input`、`output`、`cacheRead`、`cacheWrite`、可选 `cacheWrite1h`、可选 `reasoning`（是 `output` 的子集）、`totalTokens` 和 `cost`；`cost` 是按模型目录价格算出来的 input/output/cacheRead/cacheWrite 与总计[[40]](#ref-40)。

认证适配可能包含提供方专属的兼容处理；这些实现细节不构成第三方客户端的通用接入承诺，使用者应遵循正式授权与支持范围。[[22]](#ref-22) [[40]](#ref-40)

### 20.1.6 内置工具与权限边界

内置工具共 8 个，默认启用 `read`、`bash`、`edit`、`write` 四个，`grep`、`find`、`ls`、`powershell` 可通过选项启用；CLI 的 `--tools` 用 `+name`/`-name` 增减[[9]](#ref-9)[[37]](#ref-37)。关键行为与数字如下（都来自固定提交源码）：

| 工具 | 行为要点 |
| --- | --- |
| `read` | 读文本与图片；文本截断到 2000 行或 50 KB（先到为准），用 `offset`/`limit` 续读；路径解析处理 `~` 与绝对路径、剥离 `@` 前缀，**不限定在工作目录内** |
| `bash` | 起一个 shell 子进程（非 Windows 为 detached，即独立进程组）；stdout/stderr 合并流式输出；`timeout` 参数可选、**无默认**；输出保留末尾 2000 行/50 KB，超出时完整输出落临时文件；非零退出码→`isError:true`；结果另带 `structuredContent` 供 codemode 用 |
| `edit` | `path` + `edits[]`（每项 `oldText`/`newText`）；每个 `oldText` 必须在原文件唯一匹配、编辑不得重叠；整个读-改-写走文件互斥队列 |
| `write` | 创建或覆盖文件并建父目录；无工作目录限制 |
| `grep`/`find`/`ls` | 默认上限 100 条匹配（单行 500 字符）/1000 条/500 条，另受 50 KB 限制；grep 遵守 `.gitignore` 并固定带 `--no-config` 调 `rg` |

`bash` 的失败文案分得很细：非零退出码追加 “Command exited with code N”，中止追加 “Command aborted”，超时追加 “Command timed out after N seconds”，无退出码追加 “Command terminated without an exit code”，工作目录不存在直接报错[[37]](#ref-37)。同文件的写入由 `withFileMutationQueue` 按 `realpath` 串行化，不同文件并行[[37]](#ref-37)。

权限边界是 Pi 最容易被误解的地方，也是它与多数同类产品差别最大的地方。**内置工具没有任何确认、审批、拒绝逻辑**：在工具源码与 `agent-session.ts` 里检索 confirm/approve/permission（排除 trust 相关）没有命中；路径解析也不限定在工作目录内；`bash` 子进程继承完整 `process.env`，并额外注入 `PI_SESSION_ID`、`PI_SESSION_FILE`、`PI_PROVIDER`、`PI_MODEL`、`PI_REASONING_LEVEL` 这几个环境变量[[37]](#ref-37)。官方安全页的原话是：Pi 不逐次询问，工具以启动它的账户权限读写和执行[[6]](#ref-6)。唯一的内置拦截点是扩展的 `tool_call` 钩子——确认与审批要靠使用者自己去写扩展（见 20.5.6），这不是缺陷，而是一个明确的设计选择。

### 20.1.7 扩展点清单

Pi 的“强扩展”体现在一组清晰的扩展点。按从静态到可执行排序：

| 扩展点 | 形态 | 是否可执行代码 |
| --- | --- | --- |
| 上下文文件（AGENTS.md / CLAUDE.md / AGENTS.override.md） | 静态指令，累加加载 | 否 |
| `SYSTEM.md` / `APPEND_SYSTEM.md`、`--system-prompt` / `--append-system-prompt` | 替换默认前缀 / 追加 | 否 |
| Skills（`SKILL.md`，遵循 Agent Skills 规范） | 渐进披露，`/skill:name` | 内容可指示模型运行命令 |
| Prompt Templates（`.md`→`/command`） | 输入阶段展开 | 否 |
| Extensions（TypeScript，`jiti` 即时编译加载） | 注册工具/命令/快捷键/CLI flag/provider/虚拟模型/MCP 服务器，订阅事件，拦截工具调用 | 是 |
| Packages（npm/git/本地，`pi` 字段） | 打包 extensions+skills+prompts+themes | 是（可含扩展） |
| `models.json` / 自定义 provider / 虚拟模型 | 兼容端点、模型覆盖、按请求路由 | 否（配置） |
| MCP 服务器（`mcp.json`、`pi.registerMcpServer`） | 外部工具 | 是（外部进程） |
| SDK 替换点 | `ResourceLoader`、`SettingsManager`、`SessionManager`、`ModelRuntime`、`customTools`、`extensionFactories` | 视宿主 |

官方对“该用哪个”给过一条原则：“从满足需求的最弱机制开始”——持久指令用 AGENTS.md，复用提示用模板，任务专用指令与支持文件用 Skill，可执行工具/命令/事件处理用 Extension，定制终端组件用 TUI 组件，连接不支持的模型服务用自定义 provider，分发安装多个资源用 Package[[2]](#ref-2)。这条原则是 20.5 的骨架。

扩展的加载方式是“会话启动前由 `jiti` 即时编译”，扩展以 Pi 进程的 OS 权限运行[[18]](#ref-18)。这一点在 20.1.8 和 20.4 里是安全分析的支点：扩展是真正的代码边界，Packages 不形成权限沙箱[[21]](#ref-21)。

Packages 的资源声明支持过滤（`!pattern`、`+path`、`-path`），过滤只能收窄清单、不能暴露包未声明的资源；带版本的 npm 规格与 git tag/commit 视为钉死，更新只对账检出、不移动已配置的 ref[[21]](#ref-21)。

### 20.1.8 与 OpenClaw 的关系：一个需要绑定版本的事实

Pi 官方 README 至今写着一句 “See OpenClaw for a real-world integration”[[29]](#ref-29)。这句描述与 OpenClaw 所列提交的依赖状态不一致，需要分别记录 README 的集成表述和固定提交 `ff7597025dd2…`（2026-10-23）的实现观察。

OpenClaw 自己的文档声明：它拥有内置的 agent runtime，“不再有外部 Agent 框架包依赖”；`pi` 只是 `openclaw` 的 legacy 别名；`@earendil-works/pi-tui` 仍是第三方依赖，用于本地 TUI 与会话工具渲染[[47]](#ref-47)。`package.json` 里 Pi 系依赖确实只剩 `@earendil-works/pi-tui` 1.2.0，没有 `pi-coding-agent`、`pi-agent-core`、`pi-ai`；OpenClaw 自己有 `createAgentSession` 的同名实现[[47]](#ref-47)。

依赖史可以用 OpenClaw 自己的 CHANGELOG 串起来（逐条是官方发布说明）：2026.2.1 “对齐内嵌 runner 与 pi-coding-agent 的 API 更新（pi 0.51.0）”；2026.2.21 把四个 `@mariozechner/pi-*` 升级到 0.54.0；2026.3.13 升到 0.58.0；2026.4.20 为对齐 pi-coding-agent 0.66.1 的事件命名重命名压缩事件；2026.5.12 迁移到 `@earendil-works` 命名空间；2026.5.22 “把此前的 Pi agent runtime 内化进 OpenClaw，移除 legacy 包依赖，只保留 Pi 命名的 SDK 别名作为弃用插件的兼容”，同一版本还关闭了内嵌 pi-coding-agent 的自动重试，以免 OpenClaw 自己的重试与故障转移循环把失败的工具调用重放一遍[[48]](#ref-48)。

因此可以写成一句有版本绑定的结论：OpenClaw 在 2026.5.22 之前通过 npm 依赖嵌入 Pi 的 agent-core、ai、coding-agent、tui 四个包（已查发布说明覆盖 0.51.0 至 0.66.1，命名空间先 `@mariozechner` 后 `@earendil-works`），2026.5.22 起内化为自有 runtime，固定提交只剩 `pi-tui 1.2.0` 这个 TUI 组件。Pi README 的集成表述不能单独证明所列 OpenClaw 提交仍依赖完整运行时[[29]](#ref-29)[[48]](#ref-48)。这个例子本身也是一个教训：两个项目之间的关系是会随时间变化的事实，引用时必须绑定版本。
## 20.2 启动：从 `pi` 到首个模型请求

启动要回答的问题，是从敲下 `pi`（或 SDK 调 `createAgentSession`）到第一个模型请求发出之间发生了什么，以及每一步失败时系统停在哪里。Pi 的启动路径在源码里可以逐步定位，这是它作为开源对象最有教学价值的部分之一。

启动阶段体现的取舍是“可替换资源加载器”与“统一可信入口”不能自动同时获得。CLI 替使用者解析项目信任，SDK 则把宿主当作可信编排者；同样一个工作目录，经两个入口加载的资源不一定相同。嵌入方应在创建会话前完成配置与扩展来源决策，而不是模型开始工作后再用提示词补救。这里决定的是哪些代码进入进程，后续 tool_call 的确认无法撤销导入时已经发生的代码执行。

### 20.2.1 CLI 主路径

CLI 入口只有两行：`setupCli()` 和 `main(process.argv.slice(2))`。`setupCli` 设进程标题、置 `PI_CODING_AGENT=true` 与 `AI_AGENT=pi` 两个进程标记、把 `process.emitWarning` 置空，并提前配置 HTTP dispatcher[[43]](#ref-43)。随后 `main()` 按顺序做以下事情（行号都在固定提交）：

1. 把四个内置扩展工厂（llama.cpp、codemode、tool-search、mcp）并入扩展工厂列表；`--offline` 或 `PI_OFFLINE` 为真时同时设 `PI_OFFLINE=1` 与 `PI_SKIP_VERSION_CHECK=1`。
2. 子命令分流：`auth`、Windows 自更新残留清理、包命令（install/remove/uninstall/update/list）、`config`、`mcp` 命中即返回，不创建 `AgentSession`。
3. 参数解析：有 error 诊断就 `exit(1)`，warning 只打印；`--version`、`--export` 打印后退出。
4. 模式判定 `resolveAppMode`：`--mode rpc`→rpc，`--mode json`→json，`--print` 或 stdin/stdout 任一非 TTY→print，否则 interactive。非交互模式立即“接管 stdout”；RPC 与 `@file` 互斥；`--fork`、`--session-id` 各自与恢复类参数互斥[[9]](#ref-9)。
5. 迁移旧凭据到 `auth.json` 等。
6. 用一个“引导用”的 `SettingsManager` 读全局 `httpProxy` 并配置 HTTP dispatcher；再解析会话位置（`sessionDir` 优先级是 `--session-dir` > `PI_CODING_AGENT_SESSION_DIR` > 启动设置里的 `sessionDir`）。
7. 创建或恢复 `SessionManager`：`--no-session`/`--help`/`--list-models`→内存会话；`--fork`→`forkFrom`；`--session`→按路径/ID（精确，其次前缀，当前项目优先，跨项目匹配会询问是否 fork）；`--resume`→交互选择器；`--continue`→`continueRecent`；`--session-id`→已存在则打开、否则以该 ID 新建；默认新建。找不到会话打印 “No session found matching…” 并 `exit(1)`。
8. 构建运行时工厂与服务层：`ModelRuntime.create` → `SettingsManager` → `new DefaultResourceLoader` → **`await resourceLoader.reload(...)`** → 把扩展在加载期登记的 provider/虚拟模型注册进 `modelRuntime` → `modelRuntime.refresh({allowNetwork:false})` → 应用扩展 CLI flag[[31]](#ref-31)。
9. 解析模型范围、会话选项（`--model`、`--thinking`、`--tools`/`--exclude-tools` 等）、`--api-key` 运行时覆盖。
10. `createAgentSessionFromServices` → `createAgentSession`：恢复会话记录的模型（需有凭据，否则给回退提示）→ `findInitialModel` → 思考强度（显式 > 会话记录 > 每模型设置 > 全局默认 `medium`，再按模型能力夹紧）→ 默认工具 → 构造 `Agent` → 新会话追加初始 `model_change` 与 `thinking_level_change` → `new AgentSession(...)`[[11]](#ref-11)[[34]](#ref-34)。
11. 启动后检查：运行时诊断含 error 则 `exit(1)`；若其中有 “Failed to load extension” 额外提示 `Start without extensions using "pi -ne".`；非 RPC 模式读 stdin；处理 `@file`；非交互模式下会话模型为空则打印“无可用模型”并 `exit(1)`。
12. 模式分发：RPC 后台刷新模型目录（15 秒超时）后进入 `runRpcMode`；interactive 启动 UI；print/json 走 `runPrintMode`，结束后退出码非 0 才设 `process.exitCode`[[43]](#ref-43)。

交互式还有两个顺序上的讲究：它先注册信号处理、**先启动 UI 再初始化扩展**（让 `session_start` 处理器能弹交互对话）；进入主循环前异步做目录刷新、版本检查、包更新检查，这些都不阻塞输入[[41]](#ref-41)。

### 20.2.2 资源加载顺序与项目信任

`reload()` 的实际顺序（可选的）是：解析项目信任 → `settingsManager.reload()` → 包管理器解析 → CLI 扩展路径 → **扩展先于 Skills 加载** → Skills → Prompt Templates → Themes → 上下文文件 → `SYSTEM.md` → `APPEND_SYSTEM.md`[[33]](#ref-33)。系统提示的构造不在资源加载器里，而在 `AgentSession._rebuildSystemPrompt` 与 `system-prompt.ts`（见 20.3.4）。

项目信任是需要区分、但最值得补的机制。需要信任决定的项目资源集合是源码常量给出的：`.pi/` 下的 `settings.json`、`mcp.json`、`extensions`、`skills`、`prompts`、`themes`、`SYSTEM.md`、`APPEND_SYSTEM.md`，以及当前目录到祖先的 `.agents/skills`（排除 `~/.agents/skills`）；裸的 `.pi` 目录不需要决定。决定顺序：命令行 `--approve/--no-approve` → 用户级/命令行扩展的 `project_trust` 事件（第一个返回 yes/no 的扩展拥有决定权）→ `trust.json` 里最近的已保存决定 → 全局 `defaultProjectTrust`（默认 `"ask"`）[[6]](#ref-6)[[35]](#ref-35)。

两个关键边界。第一，项目信任是“资源加载”门，不是工具权限：拒绝信任会跳过 `.pi` 下受保护的设置与资源，但 AGENTS.md/CLAUDE.md 仍加载，信任之后工具权限不变[[6]](#ref-6)。第二，非交互模式下无法弹窗，`defaultProjectTrust` 为 `always` 才加载受保护的项目资源，`ask` 或 `never` 都跳过；自动化场景应显式 `--approve`/`--no-approve`[[6]](#ref-6)。

为了让扩展能处理 `project_trust` 事件，资源加载器只允许**用户级与命令行**扩展在信任决定之前加载，内置扩展与项目扩展在决定之后才加载[[18]](#ref-18)。这解释了为什么内置扩展不能参与信任决定。

### 20.2.3 SDK 创建会话的路径

SDK 的默认路径比 CLI 短，也少一环：`createAgentSession()` 不带参数时，cwd 取 `process.cwd()`，`SettingsManager.create`、`SessionManager.create`（**持久会话**）、`DefaultResourceLoader.reload()`，工具取设置里的 `defaultTools` 或 `read/bash/edit/write`[[11]](#ref-11)。

一个需要嵌入方自己注意的点：SDK 默认路径**没有项目信任环节**。`SettingsManager.create` 的 `projectTrusted` 选项默认是 `true`，而 CLI 的信任解析逻辑在 `main.ts` 里、不在 SDK 里。也就是说 SDK 嵌入默认直接加载 cwd 下 `.pi/` 里的项目扩展与 `.pi/mcp.json`。这是由三处源码拼出的推论（官方 SDK 页没有提到），标为“源码事实，非文档承诺”[[34]](#ref-34)。嵌入方若要沿用 CLI 的信任语义，需要自己传 `projectTrusted:false` 的设置管理器并实现决定。

另一个差异是内置扩展：codemode、tool_search、MCP 在 SDK 里不自动加载，需通过 `DefaultResourceLoader` 的 `extensionFactories` 显式加入，并调用 `session.bindExtensions()`（MCP 在 `session_start` 才连接）。文档对这一点的表述是“SDK sessions do not”[[11]](#ref-11)。

安装与更新也有几个版本敏感的事实。官方安装脚本 `curl -fsSL https://pi.dev/install.sh | sh`（安装器“pins all dependencies”），或 `npm install -g --ignore-scripts @earendil-works/pi-coding-agent`（不钉传递依赖，要求 Node ≥ 22.19）；Nix 版用 `nix profile upgrade pi`，`pi update` 无法更新 Nix 安装；`pi update` 无目标即更新 Pi 本身[[2]](#ref-2)[[9]](#ref-9)。官方强调，正常的 npm 安装不需要依赖生命周期脚本——这条与 20.5.5 里“包安装未见 `--ignore-scripts`”是两回事，前者是 Pi 自身的安装方式，后者是 Pi 安装第三方 package 时用的命令。

SDK 的官方示例一共 14 个，覆盖了嵌入时的常见组合：minimal、custom-model、custom-prompt、skills、tools、extensions、context-files、prompt-templates、api-keys-and-oauth、settings、sessions、full-control、session-runtime、codemode-mcp[[11]](#ref-11)。其中三个最值得对照：`full-control` 里自实现的 `ResourceLoader` 对象需要提供 11 个成员（`getExtensions/getSkills/getPrompts/getThemes/getAgentsFiles/getSystemPrompt/getAppendSystemPrompt/extendResources/reload` 等），`ModelRuntime.create` 加 `setRuntimeApiKey` 注入运行时密钥，`SettingsManager.inMemory` 里可以直接关压缩或改重试次数；`tools` 示例给出只读会话（`tools:["read","grep","find","ls"]`，内存会话）；`sessions` 示例说明 `SessionManager` 是已定稿模型上下文的权威，恢复外部历史要构造包含这些条目的 manager，而不是给 `session.agent.state.messages` 赋值[[45]](#ref-45)。

### 20.2.4 启动失败的行为

启动失败的处理并不一致：有的终止，有的降级，有的只打印警告。下表按来源逐项列出。

| 情形 | 行为 | 依据 |
| --- | --- | --- |
| 命令行参数解析出 error | 打印并 `exit(1)`；warning 仅打印 | 源码 |
| 未知选项（不是扩展登记的 flag） | 记为 error 诊断 → 启动终止 | 源码 |
| `--provider` 未配 `--model` | error → 终止（1.0.0 起，此前静默忽略） | 源码 + CHANGELOG |
| 扩展加载失败 | error 诊断 `Failed to load extension "<path>": …`，**启动终止**，提示 `pi -ne`；四种模式一致 | 源码 |
| `settings.json` 解析失败 | 该作用域按空设置继续，warning 诊断，不终止 | 源码 |
| `models.json` 出错 | 交互式启动后显示 `models.json error: …`，不终止；非交互式处理未追到 | 源码（中） |
| 会话文件非空但首个有效行不是 header | 抛 “Session file is not a valid pi session”，CLI 转成 error + `exit(1)`；空文件初始化新 header；个别损坏行静默跳过 | 源码 |
| 非交互模式无可用模型/凭据 | 打印提示并 `exit(1)`；交互式允许进 UI 后 `/login` | 源码 |
| MCP 配置/连接失败 | 启动后一次性报告，不阻塞启动 | 文档 |
| 非 TTY 下的项目信任 | 无覆盖/扩展决定/已存决定时，仅 `always` 加载受保护资源 | 文档 + 源码 |

可以看出一个总的原则：**配置解析错误、扩展加载失败、无可用模型这三类“无法安全继续”的问题会让启动终止；资源层面的可选项失败（单个 settings 文件、models.json、MCP）则降级继续，并通过诊断或启动头告知**。这是“fail fast 与 fail soft 分界”的一个具体例子，也回答了 20.4.8 里“哪些失败算启动失败”的口径。

### 20.2.5 首个请求之前：从 `prompt()` 到 `streamFn`

`prompt()` 的前置处理是一串顺序明确的检查：若正在发 `agent_settled` 则排队延后；以 `/` 开头先尝试扩展命令（即使正在流式输出也立即执行）；若正在压缩则**抛错** “Cannot submit a prompt while compaction is in progress”；触发 `input` 事件（可被扩展处理或变换）；展开 `/skill:name` 与提示模板；若正在流式输出，必须带 `streamingBehavior`（`steer`/`followUp`），否则抛错；无模型抛错；无可用凭据抛错；最后 `_startRun`[[31]](#ref-31)。

`_startRun` 先把会话标记为 busy（此后到来的消息进入本次 run 而不是另起一个），再 `_prepareRun`：对前一条 assistant 消息做压缩检查 → 触发扩展 `before_agent_start` → 拼 user 消息 → 并入挂起的 `nextTurn` 消息与扩展消息 → 生成 system 补丁消息。

进入 `agent.prompt(messages)` 后，`runAgentLoop` 先 `declareToolChanges`（把“可执行工具集 − 转录里已声明的工具集”的差异写成 system 消息的 `toolsAdded/toolsRemoved`），依次发 `agent_start`、`turn_start`、user 消息的 `message_start/end`，再进入循环[[30]](#ref-30)。每个模型请求前，`prepareRequest` 把请求上下文替换为会话投影，再做虚拟模型路由与阈值检查；随后经过扩展的 `context`/`context_with_system` 变换、图片处理、消息归一化，最终到 `streamFn`——也就是 `modelRuntime.streamSimple(model, context, requestOptions)`，这是第一个真正发出的模型请求[[31]](#ref-31)。

## 20.3 执行：一次任务从输入走到交付

执行部分关注三个内核契约：事件是否代表定稿事实，工具是否经过同一拦截路径，会话投影怎样决定下一次模型输入。小内核把这些过程保持为可组合机制，宿主则承担跨事件的任务身份和完成判断。例如一次 prompt 被接受后，重试和压缩还可能继续；只订阅第一个 agent_end 就关闭任务，会截断内核本来准备进行的恢复。通用循环与副作用状态以[第16章](../part3/04-agent-runtime.md#execution-loop)为主，本章只追踪 Pi 的对象和事件边界。

### 20.3.1 Agent Loop：双层的循环

Pi 的循环是双层的。外层 `while(true)` 只在“排队的 follow-up 到达”或“扩展要求显式继续”时再转一圈；内层每圈是一个 turn：发 `turn_start` → 把准备好的消息与排队的 steering 写入上下文 → 流式取 assistant 响应 → 有 toolCall 则执行 → `finishTurn` → 发 `turn_end` → 再取 steering[[30]](#ref-30)。官方对 turn 的定义是“一次 assistant 响应加上由它产生的工具调用与工具结果”；一个 run 重复 turn，之后检查是否还有更多工作（重试、恢复、压缩后重试、排队消息）再决定是否开始下一个低层 run[[3]](#ref-3)[[15]](#ref-15)。

停止条件在源码里是穷举的：①assistant 响应 `stopReason` 为 `error` 或 `aborted`：发 `turn_end`、`agent_end` 后**立即返回**，不执行工具；②`finishTurn` 返回 `action:"end"`；③一个回合没有（可执行的）工具调用、`terminate` 批次结束、且无 steering/follow-up/显式继续。**没有最大轮数、没有 token 或费用上限**——这是在 `agent/src` 与 `coding-agent/src` 里检索 `maxTurns` 与预算类标识无命中的结论[[30]](#ref-30)。多数同类产品都会设步数或预算上限，Pi 没有，这是它的一个重要边界。

`stopReason:"length"`（输出被 token 上限截断）且含工具调用时有个刻意设计：**所有工具调用都不执行**，每个都返回错误结果 “Tool call "x" was not executed: the response hit the output token limit, so its arguments may be truncated.”。理由是流式参数用了尽力而为的 JSON 修补解析，截断时可能得到“能解析、能过校验、但不完整”的调用[[30]](#ref-30)。

steering 与 follow-up 是两个独立的队列，各自有模式 `"all"`（一次取全部）或 `"one-at-a-time"`（一次取一条，**默认**），由设置 `steeringMode`、`followUpMode` 配置[[8]](#ref-8)。两者的语义不同：steering 是“当前回合的工具调用全部执行完之后、下一次模型调用之前”进入的纠正性输入，用于在生成中途调整方向；follow-up 只在内层循环将要结束（无工具调用、无 steering）时才取，用于在当前工作收尾后追加一件新事。steering 的轮询点有三个：run 开始时、每个 turn 结束后、`prepareNextTurn`（可能很慢，比如压缩）之后若此前没取到再取一次[[30]](#ref-30)。

交互式里这组语义落到四个键：`Enter` 发 steering，`Alt+Enter` 发 follow-up，`Alt+Up` 把排队消息取回编辑器，`Escape` 中止并把排队消息退回编辑器；SDK/RPC 用 `steer()`、`followUp()` 或 `prompt` 的 `streamingBehavior`，返回值是 `"queued"` 或 `"handled"`[[4]](#ref-4)[[11]](#ref-11)。`AgentSession` 另维护两份文本队列用于 UI，入队时发 `queue_update` 事件；扩展发的自定义消息可选 `deliverAs: steer/followUp/nextTurn`，流式期间 `triggerTurn:false` 的消息会被延后到本回合结束再写入——因为中途插入会落在 assistant 的 tool call 与其 result 之间，会被校验消息顺序的 provider 在重放时拒绝[[31]](#ref-31)。

### 20.3.2 事件模型：两个层级的语义

事件分两层。底层 `Agent` 事件：`agent_start`、`agent_end`、`turn_start`、`turn_end`、`message_start`、`message_update`（仅 assistant）、`message_end`、`tool_execution_start`、`tool_execution_update`、`tool_execution_end`。`AgentSession` 在其上加一层：`agent_end.willRetry`、`agent_settled`、`queue_update`、`compaction_start/end`、`entry_appended`、`session_info_changed`、`thinking_level_changed`、`auto_retry_*`、`summarization_retry_*`、`bash_execution_update`[[15]](#ref-15)[[30]](#ref-30)。

两个语义区分特别重要。第一，`agent_end` 只表示**一个低层 run 结束**，后面仍可能有自动重试、溢出恢复、压缩后重试、steering、follow-up；`agent_settled` 才表示“Pi 不会再自动继续”，其 `aborted` 字段区分被取消与自然结束。官方对使用者的建议是：宿主需要知道“Pi 不会自动继续”时，看 `agent_settled`[[11]](#ref-11)。第二，没有独立的 `error`/`interrupt` 事件：失败体现在 assistant 消息的 `stopReason: "error"/"aborted"`、`auto_retry_*`、`compaction_end.errorMessage`，RPC 另有 `extension_error`[[15]](#ref-15)。另外，在所列固定来源中 JSON/RPC 事件里**没有 run id** 字段，只有 session id、toolCallId[[15]](#ref-15)。

JSON/RPC 线路上的 `message_update` 只传增量，`message_end` 是权威最终消息；顶层 `usage` 是当前响应的最新累计用量[[15]](#ref-15)。

一次简单问答的事件顺序（官方 JSONL 示例）是固定的：`agent_start` → `turn_start` → user 的 `message_start/end` → assistant 的 `message_start`（`stopReason:"pending"`）→ 若干 `message_update` → assistant 的 `message_end` → `turn_end` → `agent_end{messages,willRetry}` → `agent_settled{aborted}`。带工具的回合，在 assistant 的 `message_end` 之后依次是每个调用的 `tool_execution_start` → `tool_execution_update`* → `tool_execution_end`，随后每个 toolResult 的 `message_start/end`，再 `turn_end`[[15]](#ref-15)。`agent_end.willRetry` 为真时，后面还会跟着 `auto_retry_*`，所以读事件流时“见到 `agent_end` 就认为结束”是错的，要等 `agent_settled`（见 20.3.2 的第一点）[[11]](#ref-11)。

### 20.3.3 系统提示的构成

系统提示由 `buildSystemPromptSections` 按固定顺序生成：`preamble`（默认文本是 “You are an expert coding assistant operating inside pi, a coding agent harness.”）、`tools`（每个工具一行 snippet）、`rules`、`docs`（Pi 自带文档与示例的路径）、`addendum`（APPEND_SYSTEM.md 与 `--append-system-prompt`）、`project_context`（`<project_instructions path="…">` 列表）、`skills`（`<available_skills>`）、`cwd`，再加扩展自定义 section；除 preamble 外每段包在同名标签里[[33]](#ref-33)。`SYSTEM.md` 只替换 `preamble/tools/rules/docs` 这一段默认前缀，`addendum`、`project_context`、`skills`、`cwd` 段仍会追加[[33]](#ref-33)。

上下文文件在系统提示里渲染成 `<project_instructions path="…">` 列表。每个目录按固定候选顺序只取第一个命中：`AGENTS.override.md`、`AGENTS.md`、`AGENTS.MD`、`CLAUDE.md`、`CLAUDE.MD`；向上走到文件系统根而不是 git 根；linked worktree 里的重复 AGENTS 会去重[[33]](#ref-33)。也就是说，全局 AGENTS.md 和外层到内层的所有上下文文件是**累加**的，SYSTEM/APPEND_SYSTEM 才是受信项目同名覆盖全局。

第一个请求会把完整的 sections 与 `toolsAdded` 持久化为转录首个 system 消息；之后只持久化按 section 名的补丁（`null` 表示删除某段）与工具增减[[16]](#ref-16)。这个设计让会话文件可以独立重放，也是“会话文件是权威状态”的实现基础。

### 20.3.4 工具执行：并行、串行与嵌套

工具执行策略默认是 `parallel`：模型一次返回多个工具调用时并发执行；但只要批次里**任一**工具声明 `executionMode:"sequential"`，整批就串行；内置工具均未声明。并行批次的实际顺序是先对每个调用依次发 `tool_execution_start` 并串行执行 prepare（含 `tool_call` 钩子，可能等待用户确认），全部 prepare 完才对可执行的调用 `Promise.all` 并发执行；结果消息按**调用在 assistant 消息中的顺序**写回[[30]](#ref-30)。同文件的写入由互斥队列串行，不同文件并行[[37]](#ref-37)。

单个工具调用的流水线是：按名找工具（找不到→“Tool X not found”）→ `prepareArguments` → TypeBox schema 校验 → `beforeToolCall` 钩子（可 `block`）→ `tool.execute(..., signal, onUpdate)` → `afterToolCall` 钩子。任何一步抛错都被转成 `isError:true` 的结果，**不向上抛**；源码注释的原话是“Never rejects for tool failures”[[30]](#ref-30)。工具之间可以嵌套调用：`ctx.executeTool(name, args, {signal, onUpdate})`，与模型发起的调用走同一套校验与钩子，事件带 `parentToolCallId`；嵌套调用不进转录，只在调用者结果的 `nestedCalls` 里留有限记录（参数 >8 KiB、结果 >32 KiB 不记，最多 256 条）[[18]](#ref-18)。

工具结果的语义还有两条要写清：`execute()` 里 `throw` 才产生失败结果，**返回对象不会被当作错误**，需要带数据的失败要显式设 `isError:true`；`terminate:true` 仅当同一批所有工具都 terminate 才跳过自动续跑。官方要求工具自己截断大结果并告诉模型去哪里读完整输出，有数据结果时声明 `outputSchema` 并返回匹配的 `structuredContent`，脱敏的 `tool_result` 处理器要同时替换 `structuredContent`[[18]](#ref-18)。

工具输出进入上下文前有统一预算：`read` 2000 行/50 KB，`bash` 取末尾 2000 行/50 KB，`grep` 100 条/单行 500 字符，`find` 1000 条，`ls` 500 条；超出时模型会看到“如何续读”的提示；压缩时每个工具结果再截到 2000 字符[[37]](#ref-37)[[17]](#ref-17)。工具调用没有循环层的统一超时——`bash` 有可选 `timeout` 参数、MCP 每请求默认 60 秒、codemode 脚本默认不设，这些是各工具自己的约束[[9]](#ref-9)[[24]](#ref-24)。

### 20.3.5 扩展对执行的拦截

扩展在请求链和工具链上各有挂载点。请求链顺序是 `context` → `context_with_system` → `before_provider_headers` → `before_provider_request` → `after_provider_response` → `provider_stream_event`；工具链是 `tool_execution_start` → `tool_call` → `tool_execution_update` → `tool_result` → `tool_execution_end`[[18]](#ref-18)。

处理器的组合语义分四类：通知类忽略返回值；就地变更类后者看见前者的修改；链式变换类（`input`、`context`、`tool_result`、`before_provider_request` 等）后者接收前者返回值；门类（被处理的输入、被拦截的工具调用、被取消的会话操作）阻止默认路径[[18]](#ref-18)。两个与可靠性直接相关的细节：`turn_end`、`agent_before_settle` 是两个“可行动边界”：扩展可以在此追加 `custom`/`custom_message`/`context_edit`/`compaction` 条目并置 `continue:true` 续跑一次。官方明确警告：无条件续跑会死循环，续跑条件要带保护[[18]](#ref-18)。`before_agent_start` 则可以改本次 run 的系统提示结构（sections、工具、guidelines）或追加消息；返回 `systemPrompt` 会替换整段提示，但官方更推荐改 sections/工具/guidelines，以便 Pi 追加转录增量[[18]](#ref-18)。

`tool_call` 处理器按扩展加载顺序调用，第一个返回 `{block:true}` 立即短路，**处理器抛错不捕获**——抛错沿 `prepareToolCall` 变成错误结果，即“失败即阻止”（fail-safe）；而 `tool_result` 处理器逐个 try/catch，出错只上报不影响结果[[38]](#ref-38)。官方对门类处理器的要求是：无法做出安全决定时应“fail closed”[[18]](#ref-18)。

工具负载（active tools）的中途变化会写入转录里的 system 消息；对无法表达这种中途变更的 provider，Pi 会发送完整转录检查点，**可能使缓存前缀失效**。因此官方建议先注册全部工具、把可选工具保持非激活，再用 `setActiveTools()` 选择[[18]](#ref-18)。这条建议在 20.6 的缓存调优里会再次出现。

### 20.3.6 上下文构造与压缩

模型每次看到的消息是一条流水线的产物：会话投影（活动分支，最近压缩摘要 + 自 `firstKeptEntryId` 起的条目 + 之后条目，再应用 `context_edit`）→ 扩展 `context` 变换 → `convertToLlm`（`bashExecution`、`custom`、`branchSummary`、`compactionSummary` 都转成 user 角色文本，压缩摘要带前缀 “The conversation history before this point was compacted into the following summary:”）→ provider 层丢弃 `error`/`aborted` 的 assistant 消息，为孤儿 tool call 合成错误结果 “No result provided”[[16]](#ref-16)[[40]](#ref-40)。

压缩是阈值触发的：`contextTokens > contextWindow - reserveTokens`，`reserveTokens` 默认 16384。检查点有四处：①新 prompt 之前（对上一条 assistant，含 aborted）；②多回合 run 中、工具结果追加之后、下一次 assistant 响应之前；③run 结束后（溢出、长度截断、阈值）；④手动 `/compact`[[17]](#ref-17)。token 的估计法是用最近一条有效 assistant 的 `usage.totalTokens`，加上其后消息的字符数/4（图片按 4800 字符计）；压缩刚结束、尚无新响应时 `contextUsage.tokens/percent` 为 `null`[[17]](#ref-17)。

压缩的动作是：从后往前累计估计 token 直到 `keepRecentTokens`（默认 20000）找切点；切点只能落在 user/assistant/bashExecution/custom 消息上，**绝不落在 tool result 上**（tool result 必须跟它的 tool call 在一起）；单个 user 消息跨度超过预算时在内部切并生成“历史摘要 + 跨度前缀摘要”两份；然后用 LLM 生成结构化摘要（Goal、Constraints & Preferences、Progress、Key Decisions、Next Steps、Critical Context，附 `<read-files>`/`<modified-files>`），追加 `CompactionEntry` 并重建上下文[[17]](#ref-17)。摘要请求把消息序列化成 `[User]:`/`[Assistant]:`/`[Tool result]:` 文本、省略 thinking 块、工具结果截到 2000 字符，并且**关闭 prompt-cache 写入**（这类一次性请求不太可能复用）[[17]](#ref-17)。

### 20.3.7 会话树：分叉、导航与它的边界

Pi 的会话是树。三个操作各有含义：`/tree` 在**同一文件内**移动叶子（旧分支不删除）；`/fork` 从更早的 user 消息创建新 session（文件）；`/clone` 把活动分支复制成新 session[[5]](#ref-5)。树导航的算法是：收集“旧叶子→公共祖先”的条目作为待摘要内容 → 触发 `session_before_tree`（可取消、可给摘要）→ 可选用 LLM 生成 `BranchSummaryEntry` → 落到目标位置 → `_restoreToolsFromTranscript()` 恢复该分支记录的工具负载[[31]](#ref-31)。

两个关键边界。第一，`context_edit` 与摘要是**分支相对**的——导航到编辑之前的位置会让被省略的原始内容重新可见；扩展的状态应在 `session_start` 里用 `getBranch()` 重建，不要遍历全部文件条目，否则会把已放弃的分支当成有效历史[[16]](#ref-16)。第二，**树浏览不会撤销已执行的副作用**：转录只是消息树，shell、网络、文件修改已经发生；官方文档没有对导航/fork 后外部状态做任何回滚承诺，在本章检查的 AgentSession、SessionManager 与官方会话文档中未见内建文件检查点/撤销契约。官方只在示例扩展里给了一种做法：`git-checkpoint.ts` 每个 turn 前 `git stash create`，fork 时询问是否恢复代码[[5]](#ref-5)[[45]](#ref-45)。这与多数同类产品“文件快照 ≠ 外部副作用”的原则一致，本章所查的会话接口没有把文件快照纳入导航事务，工作区恢复由版本控制或扩展处理。

### 20.3.8 四种运行形态与 RPC 协议

四种使用面共用同一 `AgentSession`：交互式 TUI、print/json（一个 `pi` 进程）、RPC（长驻子进程）、SDK（宿主进程内）。进程边界不同：interactive/print/json/rpc 各自是一个 `pi` 进程；SDK 在宿主 Node/Bun 进程内；RPC 的子进程由宿主 spawn，通过 stdin/stdout 的 JSONL 通信[[3]](#ref-3)[[12]](#ref-12)。

RPC 协议有四点值得记录。第一，成帧是严格 JSONL，一条记录一个 JSON 对象并以 LF 结尾，**不能用 Node 的 `readline`**（它还会按 U+2028/U+2029 切分，而这两个字符可合法出现在 JSON 字符串里）；客户端必须持续读 stdout，否则管道写满会让 Pi 阻塞[[12]](#ref-12)。第二，每条命令可带字符串 `id`，响应复述；命令异步处理，应按 `id` 关联而非顺序关联[[12]](#ref-12)。第三，`prompt` 的响应只表示“已接受/已排队/已被处理”（`disposition = started/queued/handled`），**不表示模型工作已完成**，完成要等 `agent_settled`；订阅事件必须先于发 prompt[[12]](#ref-12)。第四，输入流 `end` 触发有序 shutdown：先 abort 再发 `session_shutdown` 再退出；SIGTERM/SIGHUP 同样走 shutdown，退出码 143/129[[13]](#ref-13)。

RPC 的命令按宿主责任分组：运行控制负责提交、引导与取消；会话操作负责分支和上下文；模型配置影响后续请求；观测接口提供用量与最终消息。宿主只需选择自己承担的控制面，不必把全部终端命令复制到客户端。完整命令表查固定提交的 RPC 文档；本章关注接受、完成与取消的语义。[[13]](#ref-13)

RPC 下扩展也有 UI 通道：`ctx.mode === "rpc"` 且 `ctx.hasUI === true`，对话类 UI 请求（select/confirm/input/editor）由客户端响应，带 `timeout` 的由 Pi 侧到期自动按默认值解析，客户端不必自己计时；通知类（notify/setStatus/setTitle 等）不需要响应[[14]](#ref-14)。

官方对两种嵌入方式的边界表是：SDK = 进程内、直接 TypeScript 方法与事件，适合需要完整 API 的 Node/Bun 宿主；RPC = 子进程、JSONL 命令/响应/事件，适合其他语言、隔离进程、IDE、自定义客户端。文档还特别注明，“隔离”的含义**仅限进程边界**，Pi 并不因此提供沙箱[[12]](#ref-12)。
## 20.4 异常：失败、中断与恢复

Pi 的异常处理按本章固定提交的文件与符号定位。检索范围内未出现某项实现，不足以证明所有版本都不支持它；产品保证仍需与文档及所用发行版一起判断。

对小内核而言，错误被转成观察不等于业务被恢复。一次 bash 超时可以形成工具失败消息，却不能证明子进程没有向外部服务写入；一次 aborted assistant 可以被过滤出下一请求，却不应抹掉宿主的对账责任。下文因此分别检查工具结果、进程停止、会话文件和退出状态，避免把一个层面的成功推出另一个层面的保证。

### 20.4.1 异常分类与证据现状

| 异常类型 | 证据 | 关键结论 |
| --- | --- | --- |
| 中止（abort） | 源码 | 只触发 AbortController；已执行的外部副作用不回滚 |
| 工具失败 | 源码 | 一律变成 `isError:true` 回送模型，不向上抛 |
| 模型请求失败 | 源码 + 文档 | 两层重试，默认只开 agent 级 3 次 |
| 上下文超限与压缩失败 | 源码 + 文档 | 溢出后压缩重试一次；压缩失败分自动/手动两条路径 |
| 会话文件损坏 | 源码 | 逐行容错，损坏行静默跳过；无 fsync、无跨进程锁 |
| 扩展抛错 | 源码 + 文档 | `tool_call` 失败即阻止，其余上报后继续；无超时 |
| 权限与审批 | 文档 | 无逐次审批，全靠扩展自己实现 |
| 进程清理 | 源码 | 进程组 SIGKILL + 跟踪表；无“无遗留进程”保证 |

### 20.4.2 中止：能停什么，停不了什么

底层 `Agent.abort()` 只是触发当前 run 的 `AbortController`，这个 signal 同时传给 provider 请求与每个工具的 `execute`。provider 流被中止时，以 `stopReason:"aborted"` 的 assistant 消息结束（带已收到的部分内容与部分用量）；循环遇到 `error`/`aborted` 响应就发 `turn_end`、`agent_end` 并**直接返回**，不再执行该消息里的工具调用[[30]](#ref-30)[[40]](#ref-40)。

bash 的失败文案在源码里是穷举的，值得原文记下，因为它是“失败作为一等结果回送模型”的具体样本：非零退出码在输出后追加 “Command exited with code N”，被中止追加 “Command aborted”，超时追加 “Command timed out after N seconds”，无退出码追加 “Command terminated without an exit code”，被信号杀死的 shell 退出码按 `128+信号号` 报告，工作目录不存在直接报错 “Working directory does not exist: …”[[37]](#ref-37)。模型能拿到的不只是“失败”两个字，而是可以据此决定下一步的失败类别。

工具阶段的中止分两种：串行批次在每个工具之后检查 `signal.aborted` 并 `break`（后面的调用既不发 `tool_execution_start` 也没有结果）；并行批次在 prepare 阶段遇到 abort 就停止登记后续调用，已登记但尚未开始的调用得到 “Operation aborted” 错误结果。因此转录末尾可能出现“有 tool call 无 tool result”的 assistant 消息，由 provider 层在下一次请求时补合成结果[[30]](#ref-30)。

中止之后迟到的工具 partial 更新会被丢弃：`executePreparedToolCall` 用 `acceptingUpdates` 标志，工具返回或抛错后不再接收 `onUpdate`；`write` 工具刻意不在 abort 监听器里 reject，而是在每步之后检查 `signal.aborted`，让文件互斥队列一直持有到当前文件操作结束，避免队列提前释放造成并发写[[30]](#ref-30)[[37]](#ref-37)。

`AgentSession.abort()` 更高一层：取消重试等待、压缩、分支摘要，调用 `agent.abort()`，**等待会话回到 idle**，最后在 `finally` 里发 `agent_settled{aborted:true}`[[31]](#ref-31)。被中止的 assistant 消息会持久化到会话文件（含部分内容），但下一次请求前 provider 层会整体跳过 `error`/`aborted` 的 assistant 消息——它留在文件里，不会重放给模型[[40]](#ref-40)。

中止停不了什么？两个“未找到”要写清楚：官方文档没有“取消后无遗留子进程”的保证——被模型启动、又用 `setsid`/`nohup`/双 fork 脱离进程组的后台进程，不在 `kill(-pid)` 的覆盖范围内；print 模式只对 SIGTERM/SIGHUP 注册了清理，SIGINT 走 Node 默认退出，此时 detached 子进程是否被清理没有保证（系统推导，低置信度）[[42]](#ref-42)。同样，`/tree`、`/fork` 后外部副作用不随会话回滚（见 20.3.7）。

进程清理本身是尽力而为但结构清晰的：POSIX 下 `spawn` 用 `detached:true` 让命令自成进程组，中止、超时、父进程收到 SIGTERM/SIGHUP 时对进程组 `kill(-pid,"SIGKILL")`，失败回退单进程；Windows 用 `taskkill /F /T /PID`；所有活动的 detached 子进程登记在一个集合里，三种模式的 SIGTERM/SIGHUP 处理中统一清理。本章不把未发布分支的清理变化并入所列发行版行为[[42]](#ref-42)。

### 20.4.3 模型请求失败：两层重试

Pi 的重试分两层，默认值不同，这是最容易配置错的地方。

**agent 级自动重试**默认开启：`retry.enabled=true`、`maxRetries=3`、`baseDelayMs=2000`、`maxAgentDelayMs=60000`；退避是 `baseDelayMs × 2^(attempt-1)` 且上限 `maxAgentDelayMs`，代码里**没有抖动**（函数上方的注释写 “before jitter”，但函数体没有实现）[[34]](#ref-34)[[40]](#ref-40)。重试开始时发 `auto_retry_start`，成功发 `auto_retry_end{success:true}`，耗尽发 `auto_retry_end{success:false,finalError}`。重试时先把失败那次 assistant 消息及其 tool results 用 `context_edit(null)` **持久地从模型上下文省略**（原始记录仍在文件里），再 `agent.continue()` 重新发请求[[31]](#ref-31)。

哪些错误会重试，是正则匹配 `errorMessage`：可重试模式包含 overloaded、rate limit、too many requests、状态码 429/500/502/503/504/520/524、service unavailable、server error、internal error、网关转发失败等；明确不重试的“限额类”模式包括 `insufficient_quota`、quota exceeded、billing、Monthly usage limit reached、ChatGPT 订阅共享/工作区用量上限（这些以小时而非秒恢复）[[40]](#ref-40)。上下文溢出先于重试判定，交给压缩处理。

**provider 级重试**默认关闭：`retry.provider.maxRetries=0`，官方明确建议保持 0（否则会拖延 Pi 自己处理配额/用量上限错误）[[8]](#ref-8)。开启后由 `retryProviderRequest` 实现：可重试状态 408/409/429/≥500 或响应头 `x-should-retry`；优先遵从 `retry-after-ms`，若服务端要求的等待超过 `maxRetryDelayMs`（默认 60000）则**立即失败**；否则指数退避 `min(0.5×2^i, 8)` 秒并乘抖动；内部强制把 SDK 自带重试设为 0（因为 SDK 的重试定时器不响应 AbortSignal）[[40]](#ref-40)。

两层叠加的风险是：agent 级最多 3 次 × provider 级每次再多次，总延迟可观。一个真实的佐证是 OpenClaw 在 2026.5.22 的更新里，为避免“自己的重试与故障转移循环”和内嵌 Pi 的自动重试叠加而关闭了内嵌 runtime 的自动重试[[48]](#ref-48)。这证明嵌入方必须显式决定重试归属，不能默认两层都开。

### 20.4.4 上下文超限与压缩失败

上下文溢出的判定是对 `errorMessage` 做各 provider 错误文案的正则匹配（Anthropic “prompt is too long”、OpenAI “exceeds the context window”、Gemini、xAI、Groq、Mistral 等十余类），排除限流类文案；另有两种“静默溢出”也被识别：`stop` 但 `usage.input+cacheRead > contextWindow`，`length` 且 `output===0` 且输入占满窗口 99% 以上。**自定义 provider 若错误文案不匹配则检测不到**（官方注释要求把模式加进 `OVERFLOW_PATTERNS`）；Ollama 的静默截断无法检测[[40]](#ref-40)。

溢出恢复是“压缩后重试一次”：命中溢出或“可恢复的长度截断”后，若已重试过则不再重试，发 `compaction_end{reason:"overflow", errorMessage:"Context overflow recovery failed after one compact-and-retry attempt. …"}`；否则先持久省略失败尝试，再 `_runAutoCompaction("overflow", willRetry=true)`，成功后以新 run 重试。标志在一次成功响应或新的 user 消息开始时复位[[31]](#ref-31)[[17]](#ref-17)。一个特别的情形：响应**成功**但超窗时只压缩、不重试，因为 `agent.continue()` 无法从已完成的 assistant 响应继续[[31]](#ref-31)。

手动压缩支持 `/compact [instructions]`，可以指定摘要要保留的主题；`compaction_end.result.estimatedTokensAfter` 给出压缩后的估计规模[[17]](#ref-17)。

压缩失败的外显分两条路：手动压缩抛错并发 `compaction_end{aborted, errorMessage:"Compaction failed: …"}`；自动压缩发 `compaction_end{errorMessage:"Auto-compaction failed: …"}` 加 `session_compact_failed`，**自动路径吞掉异常、返回 false**，本轮不再续跑。恢复压缩失败或被取消时，保留已写入的省略编辑、不追加压缩条目、不安排内部重试。官方给的处理是：修好 provider 问题后再运行 `/compact`；关闭自动压缩不会关闭手动命令[[31]](#ref-31)[[17]](#ref-17)。压缩进行中提交 prompt 会直接抛错；树导航在流式或压缩期间同样抛错[[31]](#ref-31)。

1.4.0 的 CHANGELOG 里有三个官方自述的失败案例，比任何抽象说明都更具体：推理占满固定摘要上限导致压缩/分支摘要失败（1.4.0 起改用模型输出上限）；从“输出被截断”恢复后，被省略的 tool call 的 tool result 仍保留，导致之后每个请求都被 provider 拒绝；GitHub Copilot 大请求体 HTTP 413 原先不触发压缩[[44]](#ref-44)。这三个案例说明，溢出恢复这类跨组件机制最容易在边界条件上出问题。

### 20.4.5 会话文件损坏与恢复

会话文件的写入是每条 `appendFileSync` 追加一行 JSON，**没有 fsync、没有跨进程文件锁**（`auth.json` 才用 `proper-lockfile`，会话文件没有）；新文件在首个 user/assistant 消息出现时用 `openSync(...,"wx")` 一次性写出（“wx”防止覆盖已存在文件）；迁移时以 `"w"` 重写整个文件[[32]](#ref-32)。读取的容错是：逐行解析，空行与无法解析的行被**静默跳过**；首个有效条目不是合法 session header 则整个文件视为无效；文件末尾缺换行时补一个 `\n`；非空但无效的文件被显式打开时抛错且**不修改文件**[[32]](#ref-32)。

恢复的语义有一个容易踩的细节：加载后叶子 = 文件中**最后一个条目**；`branch(id)` 只改内存里的叶子指针，**不写入文件**（下一次 append 才形成新分支）。所以在 `/tree` 里切到旧分支后若没有新写入就退出，下次 `--continue`/`/resume` 会回到文件里最后追加的那条路径，而不是退出时所在的分支。`createBranchedSession(leafId)` 可把根到某叶子的单条路径抽成新会话文件[[32]](#ref-32)。

崩溃点分析（系统推导）：流式中的 assistant 消息在 `message_end` 前不落盘，崩溃会丢失该条；工具副作用发生在结果落盘之前，崩溃可能留下“副作用已发生、转录里无记录”；恢复时若末尾是带 tool call 的 assistant 消息而没有 tool result，provider 层会合成 “No result provided”，模型会看到“这些调用没有结果”。本章核查的 AgentSession 与 provider 恢复路径未发现自动重放未完成工具的契约，不能据此推断任意扩展的行为[[31]](#ref-31)[[40]](#ref-40)。崩溃本身会记录到 `<agent-dir>/crashes.json`（最多 5 条，超过 7 天不再提示），下次启动提示 “Pi crashed on … Run /bug to report it”[[43]](#ref-43)。

### 20.4.6 扩展抛错：fail-safe 与它的边界

扩展处理器抛错的语义按事件类型不同。总则是“上报并尽量继续，`tool_call` 失败则阻止工具”。细节：`tool_call` 抛错→工具被阻止（fail-safe）；`tool_result` 逐个 try/catch→上报后继续；`context`/`context_with_system` 逐个 try/catch→上报并沿用上一版消息；`input`、`before_agent_start` 同样上报后继续；`user_bash` 处理器失败→**阻止命令**而不是回落到本地执行；`project_trust` 处理器出错→记为 warning，决定权回落到已存决定/默认值[[18]](#ref-18)[[38]](#ref-38)。

错误上报通道是 `emitError`，把 `{extensionPath, event, error, stack}` 派发出去：print/json 打印到 stderr，RPC 输出 `{"type":"extension_error", …}` 记录[[15]](#ref-15)[[38]](#ref-38)。两个“未找到”要写清：在所查 ExtensionRunner 事件分派路径中未见统一超时装配；这不证明其他宿主不能施加期限，官方只提醒流式事件处理器里不要做慢工作；`ctx` 在 reload/会话替换后失效，继续用旧 `ctx` 会抛 stale 错误[[18]](#ref-18)。官方的写作规范里还有一条对异常直接相关：工厂函数里不要启动进程、socket、监听器或定时器，长驻资源应在 `session_start` 启动、在**幂等**的 `session_shutdown` 里释放，因为取消、reload、会话替换和进程退出可能汇聚到同一条路径[[18]](#ref-18)。

### 20.4.7 MCP 接入的失败边界

MCP 的失败语义要单独说，因为它和内置工具不同：HTTP 瞬态错误（408/429/5xx）会重试 2 次，但**工具调用本身不重试**——官方给出的理由是服务器可能已经执行过该调用；每个请求的超时默认 60 秒，有进度通知时重置[[24]](#ref-24)。这条“调用不重试”与 20.4.3 里“副作用要先判断是否已执行”是同一个原则在 MCP 层的体现。

### 20.4.8 各模式的错误通道与退出码

RPC 的错误分三类：命令级失败（`{"type":"response",...,"success":false,"error":…}`）；非法 JSON（`command:"parse"`、无 `id` 的失败响应）；接受后的失败（provider 错误、中止）只出现在消息与事件流里，**不会**为同一请求 id 再发第二个 response[[12]](#ref-12)。

退出码语义在三种非交互模式里不一致，这是脚本化的一个坑：print 模式最终 assistant 响应为 `error`/`aborted` 时 stderr 打印错误并退出非 0，`length` 也退出非 0；json 模式失败或中止的响应出现在事件流里但**不会**使退出码非 0，只有调用抛错才非 0，成败要读事件[[10]](#ref-10)[[41]](#ref-41)。`pi auth check` 以 0/1/2 表示 ready/not_ready/invalid；`pi mcp list` 有无效配置或未连接的启用服务器时退出 1[[9]](#ref-9)。

### 20.4.9 静态分析范围与尚未承诺的行为

下面这些，在已读范围内没有找到，或只有间接证据。写出来是为了防止读者把空白当成默认行为：

- 最大轮数、token/费用预算上限（检索无命中）。
- 扩展处理器超时。
- 会话文件的跨进程并发写保护与 fsync；多进程同时写同一会话文件会发生什么，官方没有说明。
- 恢复时重放或补偿未完成的工具调用（只有 provider 层补 “No result provided”）。
- “取消后无遗留子进程”的保证。
- 内置 secret masking、命令 allowlist/denylist、命令分类器。
- 内核级文件检查点/撤销，`/tree` 或 `/fork` 后外部副作用的回滚。
- 非交互模式下 `models.json` 出错的完整处理路径。
- 切换模型后提示缓存的官方说明。

“未找到”要按“本章在固定提交与官方文档中未找到”来写，不能写成“Pi 保证没有”。其中前两条、第六条来自对源码的全文检索，置信度较高；其余多为“来源没有陈述”。
## 20.5 实践：用好 Pi

这一部分的每条做法都标注依据：“官方建议”指官方文档明确写出；“本书建议”指本书依据前文机制归纳。Pi 的官方“建议”高度集中在安全与扩展写法上，日常使用文档多是功能说明，这一点会如实呈现，不把功能说明包装成官方建议。

### 20.5.1 先理解一个前提：Pi 不替你设边界

用好 Pi 的第一步是接受它的默认：Pi 不逐次询问，模型提出的 `read/bash/edit/write`（含 `rm`、网络访问）在交互、print、json、rpc 四种模式下都直接执行，用 Pi 进程的 OS 权限[[6]](#ref-6)。官方安全页的三句话必须放在一起读：“把模型生成的命令与代码当作不可信”“安全来自限制 Pi 能访问和影响的文件、凭据、进程与网络服务”“看对话记录、用项目信任、审阅改动都不构成安全边界”[[6]](#ref-6)。所以 Pi 的使用本质是：**在运行环境层面为它造边界**，而不是指望它在运行时问你。

官方给了三种运行方式的保护范围：直接运行（受限于 OS 用户权限，专用账户可收窄但仍共享 OS 与网络）；把整个 Pi 放进容器/虚拟机/沙箱（官方称“通常最强的实用选项”）；Pi 在外、只把内置工具放进隔离环境（更窄，Pi 本身与其他扩展仍在边界之外）[[6]](#ref-6)。无论哪种，官方建议只提供任务所需的文件与服务、凭据尽量放在环境之外或用窄权限短期凭据、命令不需要网络时限制网络[[6]](#ref-6)。这是官方建议，也是 20.5.6 里扩展层做“逐次确认”的语境：那个确认发生在工具层，不是系统边界。

### 20.5.2 项目信任：先决定，再自动

项目信任是 Pi 里最接近“入口闸门”的机制，但它只控制资源加载，不控制工具权限[[6]](#ref-6)。需要信任决定的资源是 `.pi/` 下的 settings、mcp.json、extensions、skills、prompts、themes、SYSTEM.md、APPEND_SYSTEM.md，以及项目 `.agents/skills`[[35]](#ref-35)。决定顺序是命令行 `--approve/--no-approve` → 用户级扩展的 `project_trust` 事件 → `trust.json` 里最近的已保存决定 → 全局 `defaultProjectTrust`（默认 `"ask"`）。

对实践者有三条直接结论（本书建议）。第一，非交互模式（CI、脚本、无人值守）无法弹窗，`ask`/`never` 都不会加载受保护资源，必须显式 `--approve` 或 `--no-approve`，否则会得到“静默不加载”而不是报错[[6]](#ref-6)。第二，拒绝信任只挡住 `.pi` 资源，AGENTS.md/CLAUDE.md 仍会加载，官方提醒“即使拒绝信任，也要把文件夹里的指令当不可信输入”[[6]](#ref-6)。第三，信任会写进 `~/.pi/agent/trust.json`，审阅项目里的 package 声明之后再授予信任，因为信任后项目扩展和设置可以在 Pi 进程里执行[[21]](#ref-21)。

### 20.5.3 上下文文件与 SYSTEM.md 怎么写

官方对上下文文件只说明加载规则，没有给“该写什么”的写作指南（限本章所列文档检索范围）：上下文文件在它所在目录及以下运行时生效，`AGENTS.override.md` 只替换同目录的 AGENTS.md/CLAUDE.md，不抑制代理目录或其他目录的上下文文件[[7]](#ref-7)。所以这里的内容是本书建议：上下文文件放模型猜不到、且从代码里读不出来的事实（构建命令、测试约定、风格、陷阱），因为它们每次请求都进上下文（见 20.6.2），不该放能被代码推断的信息；多个祖先目录的文件是累加的，写之前要想清楚每一层承担什么。

`SYSTEM.md` 替换的是默认前缀（preamble/tools/rules/docs），`addendum`、`project_context`、`skills`、`cwd` 段仍会追加[[33]](#ref-33)。官方给过一条选择原则——“从满足需求的最弱机制开始”：持久指令用 AGENTS.md，任务专用指令与支持文件用 Skill，复用提示用模板，可执行工具/命令/事件处理用 Extension，分发多个资源用 Package[[2]](#ref-2)。本书据此补一条：想改角色/语气用 `APPEND_SYSTEM.md` 或 Skill，想整体换掉 Pi 的默认行为准则时才动 `SYSTEM.md`，因为它会连同 tools/rules/docs 一起被替换。

### 20.5.4 Skills 与 Prompt Templates

Skill 遵循 Agent Skills 规范，渐进披露：系统提示里只放 name/description/location，模型需要时用 `read` 读完整 `SKILL.md`[[19]](#ref-19)。官方给的关键建议是：`description` 决定模型何时考虑加载，要写“做什么 + 何时适用”，避免 “Helps with PDFs” 这类缺乏路由信息的一句话；并且“模型可能没加载相关 skill”，需要强制时用 `/skill:name`；完整的斜杠命令清单见官方 slash-commands 页[[19]](#ref-19)[[28]](#ref-28)。规格约束：`name` 小写字母数字连字符、≤64 字符；`description` ≤1024；`disable-model-invocation: true` 表示只能显式调用；Pi 不要求 `name` 与目录名一致，但保持一致更可移植[[19]](#ref-19)。

Prompt Template 是输入阶段的机制：文件名即命令名，在模型拿到输入之前展开，参数语法支持 `$1`、`$@`、`${1:-default}` 等[[20]](#ref-20)。一个容易混淆的点：模板**不进入系统提示**，它在输入阶段展开成用户文本[[3]](#ref-3)。所以模板适合“复用一段固定的任务框架”，不适合“持久的行为规则”——后者该进 AGENTS.md。

### 20.5.5 Packages 与供应链

Packages 用 npm/git/本地来源打包 extensions+skills+prompts+themes，`package.json` 里的 `pi` 字段声明资源[[21]](#ref-21)。官方的两条建议直接指向供应链风险：审阅第三方 package 源码后再安装；写 package 时把宿主提供的包（`pi-ai`、`pi-agent-core`、`pi-coding-agent`、`pi-tui`、`typebox`）写进 `peerDependencies` 而不是 `dependencies`，否则会复制出重复的类/注册表[[21]](#ref-21)。一个源码事实要记住：包安装命令是 `npm install <spec> --prefix <root> --legacy-peer-deps`，**未见 `--ignore-scripts`**，而 Pi 自己安装文档推荐 `npm install -g --ignore-scripts …`。也就是说，第三方包的依赖生命周期脚本是否执行取决于包管理器的默认值，这不是 Pi 单独能消除的风险[[21]](#ref-21)[[2]](#ref-2)。

### 20.5.6 权限与确认：用扩展补上 Pi 没给的东西

Pi 没有逐次审批，确认与拦截只能由扩展的 `tool_call` 钩子 + `ctx.ui.confirm/select` 实现。官方示例给了四个起点：`permission-gate.ts`（正则匹配 `rm -rf`、`sudo`、`chmod 777`，无 UI 时直接拦截）、`protected-paths.ts`（对 `write`/`edit` 命中 `.env`、`.git/`、`node_modules/` 的路径直接 `block`）、`confirm-destructive.ts`（清空/切换会话前确认）、`timed-confirm.ts`[[45]](#ref-45)。官方扩展文档里的一个正式要求是：门类处理器无法做出安全决定时应“fail closed”，而 `tool_call` 处理器抛错本身就是 fail-safe（错误会变成工具的失败结果）[[18]](#ref-18)[[38]](#ref-38)。

还有几个示例扩展体现了可迁移的工程模式：`git-checkpoint.ts` 每个 turn 前 `git stash create`，`/fork` 时询问是否恢复代码，用版本控制补上“会话树不回滚工作区”的缺口；`auto-commit-on-exit.ts` 在 `session_shutdown` 里做收尾提交；`ssh.ts` 复用工具的 operations 接口把内置工具路由到另一台机器[[45]](#ref-45)。这三个例子的共同点是：都借助了 20.1.6 提到的“工具 operations 接口可注入”与“shutdown 钩子幂等”这两个内核边界，而不是另起一套执行路径。

OS 级的隔离另有官方示例 `sandbox`：覆盖内置 `bash`，用 `@anthropic-ai/sandbox-runtime`（macOS `sandbox-exec`、Linux `bubblewrap`）施加文件系统与网络限制，配置来自 `~/.pi/agent/extensions/sandbox.json` 与 `.pi/sandbox.json`[[45]](#ref-45)。这和 20.5.1 的“第三种运行方式”（工具放进隔离环境）是同一个思路，但边界更细：文件系统 denyRead/allowWrite/denyWrite、网络 allowedDomains。官方容器化文档里有一条专门的警告：不要在沙箱里运行 `/login`，那会把真实凭据写进沙箱[[26]](#ref-26)。

本书建议把放权做成分层：第一层是运行环境（容器/专用账户/网络限制），第二层是项目信任，第三层是扩展层的确认与路径保护。三层各管各的，不要指望任何一层单独承担全部。

观察入口只保留与验收直接相关的三项：工具输出可展开检查，页脚提供上下文与累计用量，`/session` 给出会话文件和费用；完整键位查交互文档。一次性 `--no-session` 会话不能恢复，不能把这种运行模式用于需要后续追责或恢复的持久任务。[[4]](#ref-4)[[5]](#ref-5)

### 20.5.7 会话、诊断与工作节奏

会话的日常操作是功能说明，但有几个值得单独说的边界。`/tree`、`/fork`、`/clone` 的选择规则官方写得清楚：相关备选留在同一文件用 `/tree`，备选应成为独立工作用 `/fork`，需要当前状态的独立副本用 `/clone`[[5]](#ref-5)。本书补充两条从机制推出的建议：树导航不撤销外部副作用（20.3.7），所以“回退到一个旧分支”不等于“回退工作区”，代码回退要靠版本控制或 `git-checkpoint.ts` 这类扩展[[45]](#ref-45)；会话文件里最后追加的条目决定下次 `--continue` 恢复到哪里，`/tree` 切走后如果没有新写入就退出，下次会回到最后写入的那条路径（20.4.5）[[32]](#ref-32)。

诊断与分享的边界也要注意：`/debug` 把渲染行与会话消息写进 `pi-debug.log`，`/bug` 生成可含会话文本或摘要的报告，`/share` 上传会话——官方都提醒分享前要审阅，因为会话可能包含提示、工具参数、命令输出、文件内容与暴露过的凭据[[6]](#ref-6)[[4]](#ref-4)。这条对任何 Agent 产品都成立，Pi 因为默认不设边界而尤其如此。

扩展的状态该放在哪里，官方给过一张选择表：随活动分支变化的工具状态放工具结果的 `details`；不进模型上下文的持久数据用 `pi.appendEntry()`；要发给模型的自定义内容用 `pi.sendMessage()`；会话外数据用外部存储；分支敏感状态在 `session_start` 里用 `getBranch()` 重建，不要遍历所有文件条目，因为已放弃的分支是“替代历史”[[18]](#ref-18)。

### 20.5.8 SDK 嵌入：把 Pi 当运行时而不是当 CLI

SDK 嵌入时，宿主是边界的所有者。官方示例给出了最小形态（`createAgentSession` → `prompt` → `dispose`）、流式订阅（先 `subscribe` 再 `prompt`，在 `message_update` 里取 `text_delta`）、内存会话、完全控制（自实现 `ResourceLoader`、传 `ModelRuntime`/`SettingsManager`）、只读会话（`tools:["read","grep","find","ls"]`）和会话运行时（`createAgentSessionRuntime` 做 newSession/switchSession/fork）[[11]](#ref-11)[[45]](#ref-45)。

三条从源码推出的边界，官方 SDK 文档没有写，嵌入方要自己决定（本书建议，标注“源码推论”）：SDK 默认视项目为已信任（`projectTrusted` 默认 `true`），若要沿用 CLI 的信任语义需自己传 `false` 并实现决定[[34]](#ref-34)；SDK 不自动加载内置扩展（codemode/tool_search/MCP），需要显式加 `extensionFactories` 并调用 `bindExtensions()`[[11]](#ref-11)；会话替换（`switchSession`/`fork`）后旧 `AgentSession` 已被 dispose，订阅必须重绑，扩展的 `ctx` 也失效[[11]](#ref-11)[[18]](#ref-18)。官方还提醒，`SessionManager` 是已定稿模型上下文的权威，直接给 `session.agent.state.messages` 赋值不会替换持久化上下文，要恢复外部历史应构造包含这些条目的 manager[[11]](#ref-11)。

### 20.5.9 常见问题诊断表

| 现象 | 常见原因 | 对策 | 依据 |
| --- | --- | --- | --- |
| 项目 `.pi` 设置没生效 | 项目未受信任，或非交互模式下未 `--approve` | 显式 `--approve`，或 `/trust` 保存决定 | 20.2.2 |
| 加了 Skill 但模型从不读它 | description 缺“何时适用”，或模型漏加载 | 重写 description；需要强制时用 `/skill:name` | 官方建议 |
| 改完设置/指令不生效 | 未 `/reload` | 改完 `/reload` | 官方建议 |
| `/tree` 切走退出后，`--continue` 回到旧分支 | `branch()` 不落盘，叶子=文件最后条目 | 切走后要有新写入，或用 `/clone` 导出 | 20.4.5 |
| 无人值守脚本里工具直接执行了危险命令 | Pi 不逐次询问 | 容器/专用账户 + 扩展层拦截（permission-gate 等） | 官方建议 |
| JSON 模式退出码是 0 但任务失败了 | JSON 失败不能仅靠进程退出码识别，应读取事件语义 | 脚本里读事件流判断，别只看退出码 | 20.4.7 |
| 压缩反复失败 | 摘要请求用的模型不可用 | 修好 provider 再 `/compact` | 官方建议 |
| 启动报 “Failed to load extension” | 某扩展抛错 | 修扩展，或 `pi -ne` 先不带扩展启动 | 20.2.4 |
| 上下文很快就满 | 上下文文件太多 / Skill 描述太多 / 工具声明太多 | 精简上下文文件、`disable-model-invocation`、缩小工具集 | 20.6.2 |

### 20.5.10 案例：在受信项目里修一个失败的测试

下面是一个教学构造的案例，未实测，只使用本章已核对过的入口与命令。任务设定是：用户报告“会话超时之后登录失败”，仓库已有测试但没有覆盖这个情形。

| 阶段 | 做法 | 依据 |
| --- | --- | --- |
| 准备与信任 | 在项目目录里启动；交互式下对项目信任弹窗选决定，非交互式（CI/脚本）显式 `--approve`，决定写进 `~/.pi/agent/trust.json` | 20.2.2 |
| 交代任务 | 写清症状、大致位置和完成条件：“先写一个能复现问题的失败测试，修复后这个测试与既有测试都通过；不要改公共接口。”完成条件是测试通过，而不是“修好” | 本书建议；20.5 开头 |
| 执行 | 默认工具 `read/bash/edit/write` 直接执行；改文件走 `edit`（每处 `oldText` 唯一匹配），跑测试走 `bash` 并看退出码 | 20.1.6、20.4.2 |
| 边界 | Pi 不逐次询问，所以这次运行放在容器或专用目录里，改动只在项目目录内；若要危险命令拦截，挂 `permission-gate` 扩展 | 20.5.1、20.5.6 |
| 验证 | 要求模型运行测试命令并出示退出码与输出，而不是“测试通过了”这句话；审查 `git diff` | 20.4.2（bash 退出码是失败类别） |
| 收尾 | `/session` 看用量；下一个不相关的任务开新会话，需要试另一条路用 `/fork`，不污染当前会话 | 20.5.7 |

这个案例里没有哪一步依赖更聪明的模型。每一步改善的都是运行时拿到的输入和边界：受信决定决定了哪些项目资源被加载，完成条件让“完成”可以被 `bash` 的退出码检验，容器/专用目录和扩展补上了 Pi 不内置的隔离，`/session` 让成本可核对。这正是 Pi 式使用的特征：内核只负责把机制跑起来，边界和验证是使用者自己搭的。

## 20.6 调优：控制上下文、成本与质量

本章核查的固定文档提供参数和默认值，未提供能直接用于当前任务的成本/质量对照实验；厂商自报的特定 token 算例也不能代替本地负载测量。所以这一部分只写“旋钮与机制”，不写“调到多少好”；所有数字都是固定提交里的默认值，不是推荐值。

### 20.6.1 度量：先看字段，再看口径

调优之前要能读数据。事件流与会话文件提供不同时间尺度的观测：`message_update.usage` 是流式累计用量，`message_end` 带权威最终 `usage` 和 `stopReason`，`tool_execution_end.durationMs` 是工具耗时，`compaction_*` 与 `auto_retry_*` 记录自动动作；RPC 的 `get_session_stats` 返回 `tokens{input,output,cacheRead,cacheWrite,total}`、`cost`、`contextUsage{tokens,contextWindow,percent}`[[15]](#ref-15)[[13]](#ref-13)。会话文件里每条 assistant 消息都有 `usage`（含 `cost`）、`provider`、`model`、`thinkingLevel`，压缩条目有 `tokensBefore`[[16]](#ref-16)。

口径上有两个要点。第一，`tokens` 和 `cost` 覆盖整个会话文件（含已被压缩掉的历史、工具自报的用量、压缩与分支摘要生成、缓存预热的 `usage` 条目），所以反映真实计费总量，而不是当前上下文的大小[[13]](#ref-13)。第二，`contextUsage.tokens/percent` 在压缩刚结束、尚无新响应时为 `null`[[17]](#ref-17)。分清楚“会话累计”与“当前上下文”两个口径，是避免误判的第一步（本书建议）。

图片也占上下文，并且有独立预算：`images.autoResize` 默认 true（最大 2000×2000 后再发给模型），`images.blockImages` 默认 false（打开则替换为 “Image reading is disabled.”）；上下文估计里一张图按 4800 字符计[[8]](#ref-8)[[17]](#ref-17)。

### 20.6.2 上下文预算：固定开销先于增长开销

上下文里的固定开销是结构性的（未测量具体 token 数）：系统提示的默认段、`<project_instructions>`（全局 + 外层到内层所有上下文文件，每次请求都带）、`<available_skills>`（每个 skill 一条 name/description/location）、cwd，加上每个已声明工具的 API 工具声明[[33]](#ref-33)。减少固定开销的开关是明确的：`disable-model-invocation: true` 的 skill 不进 `<available_skills>`；`-nc/--no-context-files`、`-ns/--no-skills`、`-ne/--no-extensions`、`--no-mcp`、`--tools`/`--no-tools` 缩小工具声明[[9]](#ref-9)[[33]](#ref-33)。

增长开销靠工具输出预算控制：`read`/`bash` 2000 行/50 KB、`grep` 100 条/单行 500 字符、`find` 1000 条、`ls` 500 条，压缩时工具结果再截到 2000 字符[[37]](#ref-37)[[17]](#ref-17)。codemode 是另一个降声明占用的手段：嵌套工具调用不进模型上下文，只有脚本输出进入；官方 1.0.0 changelog 自报“默认工具加 codemode 时，一个 GPT-5.6 请求从约 5,300 缩到约 3,300 token”——这是官方自报数字、模型与版本敏感，只能当量级参考，不能当基准[[25]](#ref-25)[[44]](#ref-44)。

压缩参数是 `reserveTokens`（默认 16384，决定何时开始）和 `keepRecentTokens`（默认 20000，不摘要的最近 token）；触发线是 `contextWindow − reserveTokens`，所以窗口越大、同一 `reserveTokens` 下触发越晚。官方只给了“按模型覆盖”的机制和一个 1M 窗口配 `reserveTokens:400000` 的示例，**没有给推荐值**（限本章所列文档检索范围）[[8]](#ref-8)[[17]](#ref-17)。摘要默认用当前会话模型，内置设置里没有“摘要专用模型”，想换更便宜的模型要用扩展 `session_before_compact`（官方示例 `custom-compaction.ts`）[[17]](#ref-17)。

### 20.6.3 模型与思考强度

思考强度七档：`off/minimal/low/medium/high/xhigh/max`，全局默认 `medium`；解析优先级是显式传入 > 会话记录 > 每模型设置 > 全局默认，最后按模型能力夹紧[[8]](#ref-8)[[34]](#ref-34)。预算型思考有 `thinkingBudgets`（minimal/low/medium/high 默认 1024/2048/8192/16384，仅对按预算控制思考的 provider 有意义）；`models.json` 的 `thinkingLevelMap` 为每个档位指定发给 provider 的值，`null` 表示不支持[[8]](#ref-8)[[23]](#ref-23)。切换模型写 `model_change` 条目，恢复会话时还原而不改全局默认；选择器只显示有可用认证的模型[[22]](#ref-22)。

交互式里模型切换的入口是 `/model`（`Ctrl+L`）选择、`Ctrl+S` 存为新会话默认、`Ctrl+P` 在“范围模型”里循环；范围由 `--models`、`enabledModels` 或 `/scoped-models` 决定，这样可以把日常可用的模型圈在一个小子集里，避免 Ctrl+P 遍历全部[[22]](#ref-22)[[9]](#ref-9)。

一个与压缩判断相关的细节：压缩检查会跳过“来自另一个模型的溢出错误”，避免从小窗口模型切到大窗口模型后用旧错误触发压缩；虚拟模型按路由后模型的限制判断压缩[[31]](#ref-31)。切换模型对提示缓存的影响，官方没有直接陈述（限本章所列文档检索范围）；缓存预热器明确“只预热选中模型的请求”[[31]](#ref-31)。

### 20.6.4 缓存与成本

Pi 的缓存预热是默认开启的（`cacheWarming` 默认 `"streaming"`）：在运行期间定时重放最近的请求以刷新 provider 提示缓存，仅当模型声明了缓存寿命且 Pi 估计避免的缓存未命中成本 ≥ $0.05 才发送，刷新点取 TTL 的 90% 且至少留 10 秒余量；重放请求限制 1 个输出 token，用量记为 `usage` 条目（`kind:"cache_warm"`）计入会话总计但不进上下文[[8]](#ref-8)[[39]](#ref-39)。**这意味着默认配置会产生额外的真实请求与费用**——官方说明了“只在预计划算时发送”，但使用者应当知道这条默认行为（本书建议：`cacheWarming:"off"` 可以关掉）。

缓存的失效因素有四处：工具负载中途变化且 provider 无法表达时发完整检查点、压缩改写前缀、空闲超过 TTL（源码里 `CACHE_TTL_MS` 是 5 分钟，注释“Anthropic's default cache TTL is 5 minutes”）、切换模型[[39]](#ref-39)。这解释了 20.3.5 的那条官方建议：先注册全部工具、可选工具保持非激活，用 `setActiveTools()` 选择，避免缓存失效。`showCacheMissNotices`（默认 false）可以打开，显示显著缓存未命中、成功预热等通知；`cache-stats.ts` 的算法把低于 1024 token 的前缀差异当作断点粒度噪声[[39]](#ref-39)[[8]](#ref-8)。

还有一些启动与网络旋钮值得记录：队列模式 `steeringMode`/`followUpMode`（`one-at-a-time` 默认 / `all`）、传输 `transport`（`auto/sse/websocket/websocket-cached`）、`httpIdleTimeoutMs`（默认 300000）、`websocketConnectTimeoutMs`（默认 15000）；启动开销可用 `--offline`（关目录刷新与版本/包更新检查）、`-ne/-ns/-np/-nc/--no-mcp` 关闭对应资源；`PI_TIMING=1` 输出各启动阶段耗时[[8]](#ref-8)[[9]](#ref-9)[[27]](#ref-27)[[43]](#ref-43)。MCP 还有一个启动细节：首个 prompt 最多等 10 秒，且只等有 `direct` 工具的服务器[[24]](#ref-24)。

### 20.6.5 可观测与外部遥测

除事件流与会话文件外，`pi-ai` 在调用方传入 `telemetry` 时会创建 `pi.ai.request` span 与指标（duration、first_chunk、tokens、cost），不记录提示与生成内容。但 `coding-agent` 构造请求选项时**没有传 `telemetry` 字段**（源码推论），所以 CLI 默认不产生这些 span；`enableInstallTelemetry`（默认 true）是另一回事，只涉及安装/更新上报与 provider 归因头[[40]](#ref-40)[[8]](#ref-8)。想自己看透过程，最直接的是 `PI_TIMING=1`（打印各启动阶段耗时）、`/debug`（渲染行与会话消息）、以及 `provider_stream_event`/`before_provider_request` 这些扩展钩子[[18]](#ref-18)[[43]](#ref-43)。

### 20.6.6 不应为省 Token 放弃的东西

官方如何评估 harness 的改动，是调优时可以参考的一个视角。Pi 自带一个评测包 `packages/evals`（私有包，不随主包分发），用 `vitest-evals` 分别在“本机”与“隔离容器”两个环境里跑 `without_docs` 与 `with_docs` 两个臂，计算文档带来的“提升（lift）”，需要 `PI_PROVIDER`、`PI_MODEL` 环境变量[[29]](#ref-29)。这里只有评测设计、没有公开的评测数值——它说明的是“怎么验证一个 harness 改动没有让任务变差”，而不是“哪个模型更好”。对自建 Pi-like 运行时或深度定制扩展的团队，这种“同一组任务、有/无某项能力”的双臂对比，比单点看成功率高更接近因果。

有几类开销是不能省的（本书建议）。验证：Pi 没有内置验证机制，但省掉验证的代价在 20.5 的语境里尤其高——因为 Pi 不逐次询问，错误的代价靠环境边界和人工审查兜底，而不是靠模型自觉。隔离：降低打断不能靠关掉扩展层的拦截或放宽容器边界，只能靠把可预判的动作写成规则、把需要判断的留给人。项目信任：`--approve` 不能当作“跳过安全”的开关，它决定的是加载哪些可执行资源。缓存预热可以关（那是省费用），但逐动作的边界不能因为省 token 而放宽。

### 20.6.7 小内核的四项成本由谁承担

小内核的价值是缩短模型、工具、消息和事件之间的路径，让使用者更容易理解与替换机制。少一个内建控制层，也意味着对应责任需要由宿主、扩展或运营者承担。本节是根据本章已列接口作出的系统推导，不宣称 Pi 默认实现了下面所有企业治理能力。

| 选择 | 内核获得什么 | 使用者承担什么 | 适用边界 |
| --- | --- | --- | --- |
| 少量核心工具 | 可见接口小，调用链清晰 | 自行提供业务契约、数据 ACL 与写入防重 | 个人可信项目优于敌对多租户 |
| 扩展即代码 | 可以注册/拦截更多生命周期行为 | 审查依赖、更新与宿主权限 | 不可信扩展必须外部隔离 |
| 会话树保留分支 | 可以回看和从旧点重新推理 | 识别外部动作已经发生 | 分支不是数据库回滚 |
| 灵活循环预算 | 可处理未知长度调查 | 显式约束时间、调用、成本和停止 | 不以无限上下文或步数解释可靠性 |

#### 一次只读分析怎样演进到业务写入

假设先用 Pi 阅读一个项目并提出性能修复方案，宿主只开放项目读取与受限测试。这个阶段可以依赖较少内建机制，因为目标是形成可审查建议。若接下来允许改代码，宿主要定义写入目录、差异与测试验收；若再允许创建工单或发送通知，业务动作就跨出了文件树，必须增加稳定操作身份、授权与结果核验。不能因为读文件任务顺利便沿用同一权限配置处理外部写入。

扩展能让这条路径接入业务工具，却不会自动成为隔离边界。它在 Pi 进程的 OS 权限下运行时，工具列表中没有 shell 也不能证明扩展不能调用进程 API。审查扩展应覆盖导入时执行、环境变量、网络出口、文件读写和安装脚本；只看注册出来的工具 Schema 会漏掉这些路径。Package 对资源清单的过滤解决分发范围，不等于内核对代码的权限裁剪。[[18]](#ref-18)[[21]](#ref-21)

可以将高风险能力放进独立服务，通过受限 RPC 或 MCP 工具暴露；服务端仍按请求者检查资源范围，凭据不返回到模型或扩展上下文。这样内核保持简洁，复杂业务约束由真正拥有数据的服务强制执行。代价是多一个连接、版本和运维边界，错误与超时也需映射回结构化观察。这个设计适合需要复用能力且能承担服务维护成本的团队，不是所有个人任务都必须采用。

#### 会话分支与外部世界的时间差

用户从旧节点分支后，模型看到的历史可以回到“尚未创建事故单”的位置，外部服务中的事故单却不会消失。若分支再次提出创建操作，宿主应识别是否仍属同一个业务意图：相同意图沿原身份查询，真正的新意图需明确授权。仅通过消息内容相同判断重复，会错误合并用户两次独立操作；每次分支生成新随机键，则可能重复一次旧操作。

因此会话树适合管理推理与交互的版本，业务事实必须另外核验。显示一条旧成功结果也不证明它现在仍有效：工单可能已关闭，权限可能撤销，文件可能被其他用户修改。交付前的查询与当前授权检查应以资源端为准。第16章 Operation 模型提供可借鉴的应用设计，不能反向说 Pi 默认内建了同样账本。

#### 无内建上限不等于无需预算

对有开放结束条件的研究任务，灵活循环可以避免过早停止；对无人值守任务，缺少上限会使无效工具调用、反复压缩或外部故障不断消耗资源。宿主应分别定义模型调用数、工具时限、总墙钟时间、输出成本、外部动作次数和任务完成条件。不能只限制单次模型 token，因为重试、摘要和子任务会把成本转移到其他调用。

停止策略也需区分完成、缺证据、权限不足、预算耗尽与结果未知。预算耗尽可以终止新增计算，却不能删除已经发出的写入责任；取消时保留原操作标识，并把核验入口交给用户。对只读研究可以返回带缺口说明的部分报告，对修改代码必须说明哪些验证没有运行，对外部写入则明确已确认与未知事实。

#### 什么时候选择这种内核

如果团队的任务边界集中在可信代码仓库，希望直接控制事件与提示组织，并愿意维护扩展，Pi 的小内核有利于建立容易理解的工作流。若需要陌生用户共享服务、长期业务对账、强审计或托管式权限控制，选择 Pi 仍然可行，但要把这些要求计入宿主设计和维护成本。选型的核心问题是“省下的内建层，是否由一个有责任人、可验证的边界接住”，不能仅比较默认工具数量。

验收时可让一个不了解扩展实现的人回答：当前主体能读取什么、能写什么、模型提案在哪被拒、会话分支后怎样查询既有动作、扩展更新后怎样撤销或回滚。若这些答案依赖模型自觉或隐含约定，说明宿主契约还没有完成。用这种问题评估小内核，比穷举所有快捷键更能迁移到其他 Runtime。

### 20.6.8 把最终输出变成宿主可以验收的结果

终端显示一段 final、RPC 返回一条结果、JSON 流正常关闭，分别描述产品交互层的结束条件。宿主仍需区分已验证完成、部分完成、缺少输入、预算耗尽、权限拒绝和外部结果未知。若任务是改代码，输出应关联差异与实际测试；若任务是读文档，输出应关联证据与范围；若任务跨外部动作，还需原操作的权威结果。不能通过模式名推导业务已经完成。

这也解释了为什么相同提示在 interactive、print、JSON 和 RPC 入口中需要分别验收：可交互审批、事件订阅和退出码由入口契约决定。选型时先固定目标入口，给正常与失败各一个用例，检查消费方是否能从输出区分上述状态。事件订阅断开不应触发整个业务动作重做，连接身份与动作身份分开保存。

### 20.6.9 宿主验收：先锁定事件，再检查业务证据

下面是本书建议的宿主消费流程，不是 Pi 新增的原生状态机。宿主在发送 prompt 前建立事件订阅，记录应用侧 task_id、目标会话和消费游标；命令响应只改变“提交受理”状态。随后将 message_end 的定稿消息与工具结果归档，等待 agent_settled，再运行任务自身的验收。模型的最后一段文字不是唯一输入，aborted、压缩错误与测试未执行都可能改变交付结论。[[11]](#ref-11)[[13]](#ref-13)[[15]](#ref-15)

| 观察 | 宿主可确认的事实 | 仍需核对 |
| --- | --- | --- |
| prompt disposition 为 queued | 消息已进入队列 | 是否最终被处理、是否因关闭而丢弃 |
| 单次 agent_end | 低层 run 已结束 | 后续重试、压缩、follow-up 是否继续 |
| agent_settled 且未 aborted | 本次不会自动继续 | 代码差异、测试与业务结果是否满足任务 |
| RPC 断开而没有定稿事件 | 观察链路已中断 | 子进程存活、会话条目和外部效果，不能直接重发写入 |
| 切换会话成功 | 活动 AgentSession 已替换 | 旧订阅解除、新订阅建立和宿主任务映射 |

重连时应先读取已有条目并按稳定条目标识去重，再决定是否提交新 prompt。由于所查事件协议不提供应用所需的完整业务动作账本，宿主不能凭文本相似度推断“之前没有执行”；若涉及外部写入，查询原操作身份的结果优先于重新指示模型执行。反过来，纯读取任务可以允许重复计算，但应将重复费用计入原任务，而不是偷偷新建一项任务隐藏成本。

验收可以从三个具体故障开始：受理后断开客户端；收到 agent_end 后发生自动恢复；切换会话后旧处理器收到迟到事件。每次记录预期责任和最终证据，确认宿主不会重复应用结果、漏掉失败或把旧会话消息归入新任务。这些是待部署团队执行的集成测试，本章没有宣称已运行。它们直接检验小内核交给宿主的责任，比为每个命令做一遍界面点击更有意义。

### 20.6.10 版本敏感事实的核验入口

| 事实 | 值与适用条件 | 固定依据与范围 |
| --- | --- | --- |
| npm 发布 | pi-coding-agent 1.4.0，2026-10-21 | npm 版本元数据[[46]](#ref-46)；与次日源码提交分别记录，不声称字节一致 |
| 源码分析点 | `936e5f40aede`，2026-10-22 | 本章固定文件/符号；不以任意主分支实现覆盖安装版本 |
| RPC 完成信号 | prompt 响应只表示受理；agent_settled 才表示不再自动继续 | 固定 RPC/SDK 文档[[11]](#ref-11)[[12]](#ref-12)；宿主仍验收业务结果 |
| 压缩默认预算 | reserveTokens=16384，keepRecentTokens=20000 | 固定设置与压缩文档[[8]](#ref-8)[[17]](#ref-17)；不是所有窗口的推荐值 |
| 工具结果预算 | read/bash 2000行/50KB；摘要每个结果截2000字符 | 固定内置工具与压缩实现[[37]](#ref-37)[[17]](#ref-17)；自定义工具另验收 |
| 缓存预热 | 默认 streaming；预估避免的缓存未命中成本至少0.05美元才发起 | 固定设置与cache-warmer[[8]](#ref-8)[[39]](#ref-39)；额外请求计入费用，不是免费能力 |
| OpenClaw 关系 | 所列提交仅依赖pi-tui 1.2.0，runtime已内化 | 固定package与历史发布说明[[47]](#ref-47)[[48]](#ref-48)；遗留别名与ACP目标分开 |


表中原始资料核验日期沿用章首记录，本文未运行产品。文档页面改用固定提交副本后，读者可以在同一版本上下文核对文件；后续升级按实际安装发行版重新验收。

## 本章小结

Pi 把 Coding Agent 拆成一条可定位的分层：`pi-ai` 归一模型协议，`pi-agent-core` 承担循环与工具执行，`pi-coding-agent` 负责会话、资源、扩展与四种运行形态。它的“小”是职责克制——不做子 Agent、不做 plan mode、不内置审批——而不是能力缺失：会话树、渐进披露的 Skills、可编译加载的扩展、SDK 与 RPC 两种嵌入方式，把可定制性留给了宿主和使用者。

这套设计的两面都很清楚。正面是可控、可嵌入、可审计：启动路径、工具流水线、重试与压缩的判定可沿固定提交的文件与符号追踪。反面是安全责任外移：Pi 不设逐次审批、不限定路径、不内置沙箱，边界要靠运行环境、项目信任和扩展三层去补；会话文件没有 fsync 和跨进程锁，崩溃后不会重放未完成的工具调用，树导航不回滚外部副作用。这些结论受本章固定源码和文档范围限制，不能据此宣称所有扩展都没有相应能力。需要开箱即用的多用户权限与业务对账时，应把补齐宿主能力的成本计入选型。

调优方面，Pi 提供了完整的旋钮（思考档位、压缩参数、重试、缓存预热、工具声明预算）和丰富的观测字段，但没有官方实测数据，也没有官方推荐值。本章因此只讲机制与默认值，把“调到多少好”留给使用者用自己的工作负载去测。

## 参考资料

官方文档是滚动页面；本章引用的文档页与固定提交 `936e5f40aede9327b9142b63506ed2b68bc6b4ac` 中 `packages/coding-agent/docs/` 的文件逐字一致，故文档类引用以站点地址给出、可用固定提交副本回查。源码一律钉到固定提交。核验日期均为 2026-10-23。

1. <a id="ref-1"></a> Pi / Earendil，[Pi（文档首页）](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/index.md)。
2. <a id="ref-2"></a> Pi / Earendil，[Quickstart](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/quickstart.md)。
3. <a id="ref-3"></a> Pi / Earendil，[How Pi works](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/how-pi-works.md)。
4. <a id="ref-4"></a> Pi / Earendil，[Usage](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/usage.md)。
5. <a id="ref-5"></a> Pi / Earendil，[Sessions](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/sessions.md)。
6. <a id="ref-6"></a> Pi / Earendil，[Security](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/security.md)。
7. <a id="ref-7"></a> Pi / Earendil，[Configuration](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/configuration.md)。
8. <a id="ref-8"></a> Pi / Earendil，[Settings](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/settings.md)。
9. <a id="ref-9"></a> Pi / Earendil，[CLI](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/cli.md)。
10. <a id="ref-10"></a> Pi / Earendil，[CLI integration](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/cli-integration.md)。
11. <a id="ref-11"></a> Pi / Earendil，[SDK](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/sdk.md)。
12. <a id="ref-12"></a> Pi / Earendil，[RPC](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/rpc.md)。
13. <a id="ref-13"></a> Pi / Earendil，[RPC commands](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/rpc-commands.md)。
14. <a id="ref-14"></a> Pi / Earendil，[RPC extension UI](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/rpc-extension-ui.md)。
15. <a id="ref-15"></a> Pi / Earendil，[JSON mode](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/json.md)。
16. <a id="ref-16"></a> Pi / Earendil，[Session format](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/session-format.md)。
17. <a id="ref-17"></a> Pi / Earendil，[Compaction](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/compaction.md)。
18. <a id="ref-18"></a> Pi / Earendil，[Extensions](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/extensions.md)。
19. <a id="ref-19"></a> Pi / Earendil，[Skills](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/skills.md)。
20. <a id="ref-20"></a> Pi / Earendil，[Prompt Templates](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/prompt-templates.md)。
21. <a id="ref-21"></a> Pi / Earendil，[Packages](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/packages.md)。
22. <a id="ref-22"></a> Pi / Earendil，[Models](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/models.md)。
23. <a id="ref-23"></a> Pi / Earendil，[Models configuration](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/models-configuration.md)。
24. <a id="ref-24"></a> Pi / Earendil，[MCP](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/mcp.md)。
25. <a id="ref-25"></a> Pi / Earendil，[Codemode](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/codemode.md)。
26. <a id="ref-26"></a> Pi / Earendil，[Containerization](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/containerization.md)。
27. <a id="ref-27"></a> Pi / Earendil，[Environment variables](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/environment-variables.md)。
28. <a id="ref-28"></a> Pi / Earendil，[Slash commands](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/docs/slash-commands.md)。
29. <a id="ref-29"></a> Pi / Earendil，[README](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/README.md)；[coding-agent CHANGELOG](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/CHANGELOG.md)，固定提交 `936e5f40aede…`。
30. <a id="ref-30"></a> Pi / Earendil，agent-core 源码：[agent.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/agent/src/agent.ts)、[agent-loop.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/agent/src/agent-loop.ts)、[types.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/agent/src/types.ts)，固定提交 `936e5f40aede…`。
31. <a id="ref-31"></a> Pi / Earendil，coding-agent 会话核心源码：[agent-session.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/agent-session.ts)、[agent-session-services.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/agent-session-services.ts)、[agent-session-runtime.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/agent-session-runtime.ts)、[sdk.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/sdk.ts)，固定提交 `936e5f40aede…`。
32. <a id="ref-32"></a> Pi / Earendil，[session-manager.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/session-manager.ts)、[auth-storage.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/auth-storage.ts)，固定提交 `936e5f40aede…`。
33. <a id="ref-33"></a> Pi / Earendil，[resource-loader.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/resource-loader.ts)、[system-prompt.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/system-prompt.ts)、[skills.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/skills.ts)，固定提交 `936e5f40aede…`。
34. <a id="ref-34"></a> Pi / Earendil，[settings-manager.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/settings-manager.ts)、[settings-defaults.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/settings-defaults.ts)、[settings-diagnostics.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/settings-diagnostics.ts)、[defaults.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/defaults.ts)，固定提交 `936e5f40aede…`。
35. <a id="ref-35"></a> Pi / Earendil，[project-trust.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/project-trust.ts)、[trust-manager.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/trust-manager.ts)，固定提交 `936e5f40aede…`。
36. <a id="ref-36"></a> Pi / Earendil，[model-runtime.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/model-runtime.ts)、[messages.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/messages.ts)，固定提交 `936e5f40aede…`。
37. <a id="ref-37"></a> Pi / Earendil，内置工具源码：[tools/index.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/tools/index.ts)、[bash.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/tools/bash.ts)、[read.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/tools/read.ts)、[write.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/tools/write.ts)、[edit.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/tools/edit.ts)、[grep.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/tools/grep.ts)、[find.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/tools/find.ts)、[ls.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/tools/ls.ts)、[truncate.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/tools/truncate.ts)、[file-mutation-queue.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/tools/file-mutation-queue.ts)，固定提交 `936e5f40aede…`。
38. <a id="ref-38"></a> Pi / Earendil，扩展宿主源码：[extensions/runner.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/extensions/runner.ts)、[extensions/loader.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/extensions/loader.ts)、[extensions/types.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/extensions/types.ts)，固定提交 `936e5f40aede…`。
39. <a id="ref-39"></a> Pi / Earendil，[cache-warmer.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/cache-warmer.ts)、[cache-stats.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/cache-stats.ts)、[compaction.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/compaction/compaction.ts)，固定提交 `936e5f40aede…`。
40. <a id="ref-40"></a> Pi / Earendil，pi-ai 源码：[utils/retry.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/ai/src/utils/retry.ts)、[utils/provider-retry.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/ai/src/utils/provider-retry.ts)、[utils/overflow.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/ai/src/utils/overflow.ts)、[api/transform-messages.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/ai/src/api/transform-messages.ts)、[api/anthropic-messages.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/ai/src/api/anthropic-messages.ts)、[types.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/ai/src/types.ts)，固定提交 `936e5f40aede…`。
41. <a id="ref-41"></a> Pi / Earendil，模式源码：[modes/print-mode.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/modes/print-mode.ts)、[modes/rpc/rpc-mode.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/modes/rpc/rpc-mode.ts)、[modes/json-event.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/modes/json-event.ts)、[modes/interactive/interactive-mode.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/modes/interactive/interactive-mode.ts)，固定提交 `936e5f40aede…`。
42. <a id="ref-42"></a> Pi / Earendil，[utils/shell.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/utils/shell.ts)、[utils/child-process.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/utils/child-process.ts)，固定提交 `936e5f40aede…`。
43. <a id="ref-43"></a> Pi / Earendil，[main.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/main.ts)、[cli.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/cli.ts)、[cli/setup.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/cli/setup.ts)、[config.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/config.ts)、[crash-log.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/crash-log.ts)、[timings.ts](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/src/core/timings.ts)，固定提交 `936e5f40aede…`。
44. <a id="ref-44"></a> Pi / Earendil，[coding-agent CHANGELOG](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/CHANGELOG.md)，固定提交 `936e5f40aede…`；1.4.0 发布于 2026-10-21。
45. <a id="ref-45"></a> Pi / Earendil，官方示例：[examples/sdk](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/examples/sdk/01-minimal.ts) 与 [examples/extensions](https://github.com/earendil-works/pi/blob/936e5f40aede9327b9142b63506ed2b68bc6b4ac/packages/coding-agent/examples/extensions/permission-gate.ts)（permission-gate、protected-paths、git-checkpoint、sandbox、subagent、plan-mode），固定提交 `936e5f40aede…`。
46. <a id="ref-46"></a> npm，[@earendil-works/pi-coding-agent](https://www.npmjs.com/package/@earendil-works/pi-coding-agent/v/1.4.0)，版本 1.4.0，发布于 2026-10-21。
47. <a id="ref-47"></a> OpenClaw maintainers，[Agent runtime architecture](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/docs/agent-runtime-architecture.md)、[package.json](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/package.json)、[THIRD_PARTY_NOTICES.md](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/THIRD_PARTY_NOTICES.md)、[src/agents/sessions/sdk.ts](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/src/agents/sessions/sdk.ts)，固定提交 `ff7597025dd2…`。
48. <a id="ref-48"></a> OpenClaw maintainers，CHANGELOG：[OpenClaw 2026.2.1 changelog](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.2.1.md)、[OpenClaw 2026.2.21 changelog](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.2.21.md)、[OpenClaw 2026.3.13 changelog](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.3.13.md)、[OpenClaw 2026.4.20 changelog](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.4.20.md)、[OpenClaw 2026.5.12 changelog](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.5.12.md)、[OpenClaw 2026.5.22 changelog](https://github.com/openclaw/openclaw/blob/ff7597025dd22b087b250a4df94caded50e798aa/CHANGELOG/2026.5.22.md)，固定提交 `ff7597025dd2…`。
