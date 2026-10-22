# 第17章 成熟 Coding Agent 内部实现：Codex CLI、Claude Code 与 OpenCode

> 成熟的 Coding Agent 不是“会改代码的模型”，而是一个持续组织上下文、调度工具、约束副作用并验证结果的工程运行时。

**资料核验日期：2026-10-22。** 本章从三个产品观察 Coding Agent 的运行机制，但开源状态并不相同：Codex CLI 与 OpenCode 有公开源码；Claude Code 核心程序闭源。本章对 Codex CLI、OpenCode 的实现描述以对应公开仓库为依据，对 Claude Code 只描述 Anthropic 官方文档公开的产品行为与接口。无法由公开源码或官方资料确认的模块边界、数据结构和内部算法，会明确标为“系统推导”，不作为实现事实。[[1]](#ref-1)

## 17.1 先看共同问题：让模型参与工程闭环

### 17.1.1 从代码生成到可审查交付

普通代码生成器以文本片段为主要产物。Coding Agent 的产物则存在于仓库和执行环境：它读取文件、发起模型请求、调用工具修改代码、运行命令，再把工作结果交给人审查。模型回答只是交付说明，不能代替真实的文件差异和验证结果。

对比三个产品时，可以先抽象出一条不绑定具体实现的工作路径：

~~~mermaid
flowchart LR
    A[用户任务] --> B[建立会话和任务上下文]
    B --> C[读取规则与代码证据]
    C --> D[模型选择下一步]
    D --> E{需要工具吗？}
    E -- 是 --> F[权限裁决与工具执行]
    F --> G[记录结果并反馈给模型]
    G --> D
    E -- 否 --> H[候选完成]
    H --> I[检查差异和验证证据]
    I --> J[向用户交付]
~~~

这是本书用于分析产品的共同模型，不代表任何一个产品的源码模块图。它把关注点分成四层：

| 层次 | 核心问题 | 不能混为一谈的事实 |
| --- | --- | --- |
| 任务与上下文 | 这一轮模型要解决什么、能看到哪些信息？ | 磁盘文件存在，不代表已进入上下文 |
| 模型与工具 | 模型提出什么行动，运行时如何执行？ | 模型生成工具调用，不代表工具已经成功 |
| 权限与隔离 | 哪些动作可以执行，动作影响到什么范围？ | 获得审批不等于进程被沙箱约束 |
| 验证与交付 | 哪些证据足以支持“完成”？ | 模型说测试通过，不等于测试进程实际成功 |

### 17.1.2 一轮工具调用不是整项任务

一次模型响应可以只含解释，也可以含一个或多个工具请求。运行时要把工具请求解析为可执行操作，检查参数与权限，调用本地或外部能力，将结果归入会话，再决定是否继续采样。Coding Agent 的主循环因此是一个受控状态机，不是单次提示词调用。

可以把一次任务视作以下状态迁移：

~~~text
输入任务
  → 整理当前上下文
  → 请求模型产生下一步
  → 若有工具调用：授权、执行、记录结果
  → 将观察结果带回后续上下文
  → 继续循环或形成候选完成
  → 对照需求、差异和验证结果审查
~~~

工具可能失败、超时、被用户拒绝或被中断。可靠运行时必须把这些情况作为不同结果回送，而不是把失败输出包装成成功。取消也不一定撤回已经发生的外部副作用；远端请求、数据库写入和命令进程都需要自己的恢复语义。

### 17.1.3 本章采用的证据等级

后面的产品章节会把陈述分成三类：

- **源码事实**：可以定位到公开仓库中的实现或类型定义；
- **官方文档事实**：官方文档描述的功能、调用方式或边界；
- **系统推导**：根据公开证据抽象出的架构解释，便于学习，但不能当作产品内部的精确实现。

产品资料会持续变化。Codex 的链接指向 OpenAI 官方文档和公开仓库主分支；OpenCode 部分采用官方 V2 文档和公开开发分支；Claude Code 部分按官方滚动文档记录。讨论默认值、权限模式和功能开关时，应以对应版本和入口为准，不从一篇文章推断所有部署。

<a id="coding-agent-runtime-startup"></a>

### 17.1.4 运行时组成：界面、会话、模型与执行环境

[第14章](../part3/03-agent-runtime.md#architecture)给出了通用 Runtime 组件参考设计。本章把它对应到三个产品，回答两个具体问题：**哪些组件留在本机或服务进程中，输入第一条任务前后分别发生什么？** 本节以本地编码场景为边界，资料核验于 2026-10-22；Codex 采用公开 App Server 接口，Claude Code 采用官方行为文档，OpenCode 采用 V2 文档。下面分别展开每个产品的实现边界与启动过程。

先分开三个时间尺度：**进程启动**准备服务与连接，**会话初始化**选择工作区、配置并建立或恢复历史，**轮次启动**把新输入送入模型与工具循环。它们可以连续发生，也可以相隔很久。一个后台服务能承载多个会话，一个会话可以执行多轮；恢复历史不会自动恢复旧命令进程，更不会把磁盘文件恢复到历史时刻。

[![Codex、Claude Code 与 OpenCode 的运行时组成对照：客户端、运行时职责、模型服务与执行后端](../images/coding-agent-runtime-comparison.png)](../images/coding-agent-runtime-comparison.svg)

*运行时组成对照。虚线框表示逻辑职责分组，不代表三款产品具有相同的进程划分；框内卡片不表示调用顺序。蓝色双向箭头表示模型请求与响应，橙色箭头表示受控动作。Claude Code 一列采用官方行为文档，不推测私有模块。点击图片查看可缩放、可点击来源标注的 SVG。*

下表按职责归纳产品的公开组成；同一格内的职责不意味着同一个进程或同一个源码模块。

| 职责 | Codex：以本地 App Server 接入为例 | Claude Code：以本地终端会话为例 | OpenCode V2：以本地 CLI 为例 |
| --- | --- | --- | --- |
| 入口与交互 | TUI 或其他客户端；App Server 提供结构化请求、审批与流式事件 | 终端交互入口；Harness 围绕 Claude 模型提供上下文和工具 | TUI、Mini 或非交互 CLI；默认连接当前用户的共享后台服务 |
| 会话与执行对象 | 对外使用 Thread → Turn → Item：会话包含轮次，轮次包含消息、工具调用、命令或文件变化等条目 | Session 保存对话、工具调用与结果；核心内部对象和调度结构未公开 | 后台服务持有 Session、配置、集成、权限与工具执行；客户端发起操作并显示状态 |
| 配置与上下文 | 分层 `config.toml`、项目可信状态、`AGENTS.md` 指令链，以及会话历史 | 分层 Settings、项目指令、Auto memory、会话历史与按需加载的 Skills | 全局、项目与系统配置，`AGENTS.md`，Agent 配置与模型选择 |
| 模型接入 | 模型配置与认证；App Server 的轮次接口驱动生成 | 配置选定的 Claude 模型；本地执行工具不表示本地运行模型 | Provider / Model 目录、凭证与适配包；运行时按选定模型调用供应商 |
| 扩展与工具 | 命令、文件变化、MCP 等工具过程通过事件回传；工具可用性由配置与环境决定 | 内置文件、搜索、Shell 等工具；MCP、Skills、Hooks、Subagents 提供扩展 | 文件、搜索、Shell 等工具，MCP、Skills 和插件；Permissions / Policies 控制调用 |
| 权限与实际执行 | 审批与沙箱策略分别控制动作放行与执行资源边界；可连接独立执行环境 | 权限规则与模式控制动作；Shell、外部服务和可配置沙箱承担实际执行 | 服务端等待客户端审批或拒绝动作；Shell 和外部服务执行副作用，角色权限不能替代系统隔离 |
| 持久化与恢复 | Thread 接口支持读取、恢复、分叉与归档；不等于重放所有外部操作 | 本地 JSONL 会话文件支持继续与分叉；文件 Checkpoint 有独立边界 | Session API 与本地数据存储承载历史；快照和会话恢复不覆盖所有外部副作用 |

Codex 的接口与配置依据官方 App Server、配置及指令发现文档；Claude Code 的组成依据官方 Harness、Settings、Memory 与 Skills 说明；OpenCode 的服务归属、配置和 Provider 依据 V2 文档。表格是跨产品的职责映射，不是三家通用的模块命名规范。[[2]](#ref-2)[[3]](#ref-3)[[4]](#ref-4)[[5]](#ref-5)[[6]](#ref-6)[[7]](#ref-7)[[8]](#ref-8)[[9]](#ref-9)[[10]](#ref-10)[[11]](#ref-11)[[12]](#ref-12)

**图示的文档与源码依据**

两张图使用统一的证据标注。源码固定为 Codex `86e17e98c7f9` 与 OpenCode `c93aaae37157`（均核验于 2026-10-22），避免主分支后续变化使图中入口无法回查。连线、逻辑角色和时序排列是本书依据公开接口和实现归纳的说明；源码事实对应下表的可定位入口，不把图中的每个中文职责名当成产品内部类名。[[13]](#ref-13)[[14]](#ref-14)

| 图中标注 | 外部文档或代码入口 | 支撑的节点与关系 |
| --- | --- | --- |
| C1：Codex 服务与会话 | `app-server/src/lib.rs::run_main_with_transport_options`、`message_processor.rs`、`core/src/thread_manager.rs::ThreadManager` | 服务启动、初始化请求检查、Thread/Turn 请求分发、进程级存储与会话管理；配置和指令规则另见官方文档。[[13]](#ref-13)[[3]](#ref-3)[[4]](#ref-4) |
| C2：Codex 循环与工具 | `core/src/session/turn.rs::run_turn`、`core/src/tools/router.rs::ToolRouter` | 模型采样、工具分发、结果进入下一轮；审批和执行环境在相邻控制路径落实。[[15]](#ref-15)[[16]](#ref-16) |
| O1：OpenCode 连接路径 | `packages/cli/src/services/server-connection.ts::resolve`、`standalone.ts::start` | 指定服务检查、共享服务发现与确保启动、私有子进程及 Ready 输出等待。[[14]](#ref-14) |
| O2：OpenCode 会话与工具 | `packages/core/src/session/runner/llm.ts::drain`、`packages/core/src/tool.ts::snapshot` / `execute` | 待处理输入、SessionStore、模型传输、Step、事件和工具结果归一化；配置、Provider 与权限语义另见官方文档。[[17]](#ref-17)[[10]](#ref-10)[[11]](#ref-11)[[12]](#ref-12)[[18]](#ref-18) |
| D1：Codex 公开协议 | 官方 App Server 文档 | 每连接一次的握手、Thread/Turn/Item、审批与流式事件；不把客户端元数据视为模型服务认证。[[2]](#ref-2) |
| D2：Claude Code 公开职责 | 官方 How Claude Code works、Settings、Memory、Skills、Permissions | Harness、配置与指令、会话历史、按需能力加载、循环和工具权限；不证明私有线程模型或模块边界。[[5]](#ref-5)[[6]](#ref-6)[[7]](#ref-7)[[8]](#ref-8)[[19]](#ref-19) |
| D3：Claude Code 生命周期 | 官方 Hooks reference 的 `SessionStart` 与 `Setup` | 启动和恢复的 Hook、首个模型请求的等待条件，以及 Setup 需要明确触发的边界。[[20]](#ref-20) |
| D4：OpenCode V2 产品约定 | 官方 CLI、Config、Providers、Instructions | 共享服务的职责、工作区配置、Agent/Model 选择与项目指令；不是内部函数调用顺序。[[9]](#ref-9)[[10]](#ref-10)[[11]](#ref-11)[[12]](#ref-12) |

## 17.2 Codex CLI：公开源码中的 Runtime 与协议边界

### 17.2.1 先划清 Codex 的开源范围

OpenAI 将 Codex CLI、Codex SDK 和 App Server 等列为开源组件，同时明确 IDE 扩展与 Codex Cloud 不开源。因此，准确的说法是“Codex CLI 开源”，而不是“Codex 的所有产品和服务都开源”。公开仓库让人能追踪 CLI/runtime 的代码，但不意味着云端模型或托管服务实现也在仓库内。[[1]](#ref-1)

仓库中的 Rust 核心 crate 自述为 Codex 的业务逻辑实现，并面向多种 Rust UI 使用。由此可以从源码分析 CLI 的会话、工具与安全边界；分析范围仍要限制在仓库实际公开的组件。[[21]](#ref-21)

### 17.2.2 Session 与 Turn：运行时的两个时间尺度

Codex 的源码把长期会话和一次执行轮次分开管理。Session 保有线程、历史、配置、上下文和服务等生命周期较长的状态；Turn 则承接当前用户输入、模型采样、工具执行和中断等短期状态。源码中的 Session 与 Turn 结构、相关队列和取消令牌可以直接检视。[[15]](#ref-15)

在核心执行路径中，turn 逻辑反复向模型发起 sampling。若返回工具调用，运行时执行相应工具并把结果送回后续采样；若返回助手消息，则记录并检查该轮是否结束。这个源码注释给出了最小 Agent Loop 的清晰实现证据。[[22]](#ref-22)

这个区分很有用：会话可以继续存在，而某个 turn 已经结束；turn 也可以被中断或接收追加输入，而无需把整段会话重建。对持久化、统计、恢复和取消而言，Session 与 Turn 不是同一个状态对象。

### 17.2.3 Core、TUI 与 App Server：同一 Runtime 的不同接入面

Codex 的协议 crate 定义 UI 与核心之间使用的类型，并描述内部 Core/TUI 通信和外部 App Server 通信。文档还说明，该协议尽量承载数据类型，而不把主要业务逻辑塞入协议层。[[23]](#ref-23)

App Server 暴露结构化协议，供更丰富的界面和客户端驱动 Codex。CLI/TUI 与外部客户端因此可以围绕相对稳定的请求、事件和状态同步协作。App Server 的源码与 README 可用于确认公开的传输方式和协议边界。[[24]](#ref-24)

从系统设计上看，这种边界把“如何呈现任务”与“如何执行任务”解耦：前端不需要复制一套模型循环，Runtime 也不必绑定单一 UI。这里的“解耦”是对公开协议和 crate 职责的架构总结，不代表所有界面共享完全相同的功能。

### 17.2.4 Tool Router：工具声明不等于执行权限

Codex 核心包含 Tool Router，负责把模型可见的工具请求关联到具体执行路径。工具输入在执行前还可能经过审批、环境选择、沙箱和策略检查。相关源码把路由、审批与沙箱分别放在不同文件中，能观察到它们是多个相邻的责任点，而不是一个“工具权限开关”。[[16]](#ref-16)

这体现了两种不同的控制：

1. **审批控制**回答某个动作是否需要用户或策略决策；
2. **执行隔离**回答获准动作实际能读写哪些路径、连接哪些目标或启动什么进程。

同意运行一个命令，不会让该命令自动变安全；沙箱允许某个路径，也不代表命令本身符合业务意图。审查时应分别看决策和执行环境。

### 17.2.5 Context Compaction：保存任务连续性，不是无损记忆

Codex 源码中包含上下文窗口、自动压缩和压缩任务等实现。压缩需要在上下文预算受限时继续推进对话，它对历史做选择或归纳，必然可能丢失细节。公开代码可用于追踪压缩的触发与恢复流程，但不能把摘要当成工作区当前状态的事实来源。[[25]](#ref-25)

因此，压缩后的任务状态应重新对照仓库：文件是否仍是记录中的版本、差异是否变化、测试是否针对当前差异运行。历史摘要可以缩短恢复路径，无法证明磁盘、Git 或远端系统仍与旧描述一致。

### 17.2.6 多 Agent 与共享状态

Codex 仓库公开了 Agent、子任务、协作和消息传递相关代码。理解这类功能时，首先要区分“上下文分开”与“工作区分开”：前者避免共享完整对话，后者避免并行写入同一批文件。一个只读子任务即使与主 Agent 共用目录，也不一定产生写冲突；两个 Agent 即使使用独立上下文，也不因此获得隔离文件系统。[[15]](#ref-15)

跨 Agent 的消息还可能迟到、被取消或落到不同 turn。源码状态里对当前轮是否继续接收子 Agent 消息有明确建模，这说明协作交付不仅是“spawn 后等待”，还涉及消息归属、终止条件和晚到结果的处理。[[15]](#ref-15)

### 17.2.7 从 Codex 源码得到的架构观察

Codex 的公开实现适合观察一种“核心 Runtime + 前端协议 + 受控工具执行”的组织方式。它将会话、turn、协议、工具路由、安全策略、状态持久化和 UI 接入拆为可追踪的 crate 或模块。这个拆分有助于源码审计，也带来跨模块状态协调的复杂度。

不应从源码开放推断模型开放，也不应把主分支今天的实现当作永久稳定接口。源码分析最好记录具体文件和核验日期；如果结论依赖某个 feature gate、协议版本或运行平台，还要把前提写在结论旁边。

### 17.2.8 启动过程：从启动服务到首个 Turn

以客户端启动 `codex app-server` 的公开协议路径为例，可以沿下面的顺序观察。它展示一个可验证的接入边界，不要求所有 Codex 界面都用同一种进程部署方式。

| 阶段 | 初始化或执行动作 | 完成后得到什么 |
| --- | --- | --- |
| 启动服务 | 启动 App Server，选择传输并准备服务；读取相关配置与认证环境 | 可建立连接的本地服务；此时还没有用户任务的 Turn |
| 连接握手 | 客户端发送 `initialize`，收到响应后发送 `initialized`；传入客户端身份及能力声明 | 当前连接完成协议初始化；握手前的业务请求会被拒绝 |
| 建立或恢复会话 | 使用 `thread/start` 新建，或 `thread/resume` 恢复；选择模型、工作目录与执行策略 | Thread 身份及会话状态；配置与项目指令参与后续上下文准备 |
| 启动轮次 | 客户端用 `turn/start` 提交用户输入和可选的本轮覆盖参数 | 活跃 Turn；开始处理输入与后续模型生成 |
| 进入执行循环 | 接收模型输出，处理工具调用及必要审批，运行命令或修改文件，把结果带入后续生成 | 消息、命令、文件变化等 Item 与增量事件 |
| 收束本轮 | 客户端消费 `item/*`、`turn/*` 通知，直到收到带最终状态的 `turn/completed` | 已结束的 Turn；Thread 可继续接收下一轮输入 |

协议硬顺序是**连接 → 握手 → 建立或恢复 Thread → 启动 Turn**。`initialize` 每个连接执行一次，不能把它当成每轮模型调用前的初始化。`turn/completed` 还可能表示中断等状态，不能仅凭事件名认定需求已实现。[[2]](#ref-2)

配置和指令也有各自的加载边界：Codex 从用户及可信项目的 `.codex/config.toml` 等层解析配置；未受信任项目的项目配置层会被跳过。`AGENTS.md` 按全局与项目路径构造指令链，并受大小上限约束。因此，“目录下有配置或说明文件”不等于“本次会话已经采用它”。[[3]](#ref-3)[[4]](#ref-4)

当前 App Server 文档还说明默认会启动本地 Code Mode host，也可指定远端 host。它是 App Server 向执行宿主发起连接的方向；`--listen` 则控制客户端怎样连入 App Server。这两条连接不能混为一条，更不能把模型服务端点当成命令执行宿主。具体执行能力与可用参数应绑定实际安装版本。[[2]](#ref-2)

## 17.3 Claude Code：基于官方文档分析闭源产品

### 17.3.1 GitHub 仓库公开，不代表核心程序开源

Anthropic 的 Claude Code GitHub 仓库是公开可访问的，但仓库许可证写明版权归 Anthropic 所有、保留所有权利，并指向 Anthropic 商业服务条款；它不是可按常见开源许可证使用的核心实现源码仓库。公开仓库中的 README、变更记录、示例和插件材料，不能替代未公开的 CLI 实现。[[26]](#ref-26) [[27]](#ref-27)

因此，本节不从非官方逆向、泄露代码或社区猜测补写内部细节。即使某个细节听起来符合常见 Agent 设计，只要官方文档没有证实，也会标作“系统推导”，而不是 Claude Code 的确切实现。

### 17.3.2 官方公开的 Agent Loop

Anthropic 官方文档把 Claude Code 的工作描述为收集上下文、采取行动、验证结果三个阶段。阶段会按任务需要交错：理解代码的问题可能只需检索，修复任务则会多次穿插编辑和验证。官方也直接将 Claude Code 描述为围绕模型提供工具并管理上下文的 Harness。[[5]](#ref-5)

文档可确认的概念循环如下：

~~~text
理解用户任务
  → 读取项目与代码上下文
  → 选择文件、搜索、命令等工具
  → 观察工具结果并调整方案
  → 验证实现
  → 继续行动或向用户说明结果
~~~

这描述的是公开的产品行为，不是私有函数调用图。可见的工具包括文件读取、编辑、搜索和命令执行；扩展层还支持 Skills、MCP、Hooks 与 Subagents。工具结果成为后续决策的新观察，因此人可以在循环中途纠正方向。[[5]](#ref-5) [[28]](#ref-28)

### 17.3.3 Session、Context 与历史压缩

Claude Code 官方文档说明，对话、工具使用和结果会保存在本地 Session 文件中，用于恢复、继续或分叉；每个新 Session 默认有新的上下文窗口。切换 Git 分支会改变模型能读取的工作区文件，但不会自动改写会话历史。[[5]](#ref-5)

上下文包含对话、文件和命令结果、项目指令、载入的 Skill 与系统指令。接近窗口限制时，Claude Code 会清理旧工具输出，必要时摘要对话；官方提示早期的详细指令可能在压缩后丢失。MCP 工具定义默认可延迟加载，Skills 也可按需载入；Subagent 使用独立上下文，主会话接收摘要。[[5]](#ref-5)

可迁移的工程原则是保存稳定约束与任务事实的区别：项目规范适合放进持久指令，动态代码事实应回到仓库重新读取。会话日志和摘要能帮助恢复工作，但都不能证明当前工作区与历史描述一致。

### 17.3.4 扩展层：Skills、MCP、Hooks 与 Subagents

Anthropic 的公开文档把 Skills 作为可复用知识与工作流，把 MCP 作为外部服务连接，把 Hooks 作为生命周期事件上的自动化，把 Subagents 作为隔离上下文中的委派任务。插件可以把多种扩展打包分发。它们连接到 Agent Loop 的不同位置，不能因名字都叫“扩展”就当成同一类工具。[[28]](#ref-28)

- **Skill**主要把做法和领域知识按需提供给模型；
- **MCP**将外部系统的工具或资源接入；
- **Hook**在某些生命周期事件上触发脚本、提示或其他动作；
- **Subagent**将边界明确的工作交给独立上下文；[[29]](#ref-29)
- **Plugin**是扩展的打包和分发形式。

这些是官方公开的功能语义。内部如何注册、调度和持久化扩展，若没有公开源码或官方实现说明，不在本章中作源码级断言。

### 17.3.5 Permissions 与 Checkpoints：授权、约束和恢复各司其职

权限模式与规则决定哪些工具调用可以执行、哪些需要询问或拒绝。官方文档还说明 Hooks 可以参与工具生命周期。权限配置不是同一层的全部安全机制：它回答工具调用是否被放行，却不单独定义命令进程可访问的所有资源。[[19]](#ref-19)

Claude Code 的官方文档描述了文件编辑前的 Checkpoint，可用于回退被跟踪的文件变化；数据库、API 和部署等远端副作用不能由文件回退撤销。这个边界适合作为一般 Agent 系统的提醒：恢复能力必须与副作用类型对应。[[5]](#ref-5)

### 17.3.6 哪些结论是公开事实，哪些只是系统推导

| 结论 | 证据等级 | 依据与限制 |
| --- | --- | --- |
| 工作过程包含收集上下文、采取行动、验证结果 | 官方文档事实 | 公开的用户与开发者文档 |
| Session 历史用于恢复、继续或分叉 | 官方文档事实 | 文档描述本地 Session 行为 |
| 工具调用会形成循环并把观察结果用于下一步 | 官方文档事实 | 产品行为说明，不是内部源码图 |
| 模型请求前有哪些内部模块、具体类和线程模型 | 不能确认 | 核心程序源码没有按开源许可证公开 |
| 复杂任务可能由多个内部阶段组成 | 系统推导 | 用于解释可见行为，不能据此声称具体实现 |
| Checkpoint 等于完整机器回滚 | 错误结论 | 官方资料限定为文件变化，不覆盖远端副作用 |

这样分析闭源产品仍有价值：官方公开行为足以讨论产品接口、状态边界和工程取舍；但若文章目标是审计内部实现，公开资料就有明确上限。这个证据边界本身也是产品比较的一项重要结果。

### 17.3.7 启动过程：初始化 Harness 与 Session

官方公开的是产品行为、配置与生命周期事件，没有公开足以重建整个核心程序的源码。下面按依赖关系归纳启动过程；表格不是私有函数调用顺序，也不表示所有准备动作严格串行。

| 阶段 | 官方资料能确认的行为 | 需要区分的边界 |
| --- | --- | --- |
| 启动与配置准备 | 在项目目录运行 `claude`，采用相关 Settings 和启动参数，使用配置的模型及访问条件 | 项目目录、模型选择与工具权限是不同配置事实 |
| 指令与能力准备 | 载入适用的项目指令与 Auto memory 索引，发现 Skills 等扩展 | 子目录指令、记忆专题及 Skill 全文可以按需读取；不会预读整个仓库 |
| 建立或恢复 Session | 新会话从新的上下文窗口开始；`--continue` 或 `--resume` 继续已有 Session | 恢复的是会话历史，读取的仍是当前工作区文件 |
| 会话启动事件 | 执行匹配的 `SessionStart` Hook；新建和恢复分别可用 `startup`、`resume` 匹配 | 当前文档允许启动 Hook 在后台运行、先显示输入界面；提交的首条输入仍会等待 Hook 完成后送给模型 |
| 首次输入与循环 | 处理用户输入，组合上下文，进入收集上下文、行动、验证的循环 | 每次工具使用都受相应规则控制；工具结果继续成为模型观察 |
| 记录与后续交互 | 对话、工具使用和结果写入本地 Session 文件，允许继续、恢复或分叉 | 会话持续存在不意味着命令持续运行；文件恢复不撤销远端副作用 |

Settings、项目指令、会话行为与 Skill 渐进加载分别见官方文档。Skill 的名称和描述用于让模型知道有哪些方法，调用时才加入其内容；Auto memory 的索引与详细专题文件也不是同一种加载方式。[[5]](#ref-5)[[6]](#ref-6)[[7]](#ref-7)[[8]](#ref-8)

这里最容易误判的是“界面出现”与“首个模型请求已可发送”。`SessionStart` 可以向上下文补充信息，所以界面先出现时，请求仍可能等待启动 Hook。另一个事件 `Setup` 用于明确触发的初始化或维护，**不是普通每次启动都会执行**。这说明需要依据具体产品定义判断启动完成，不能照搬一个统一的 Ready 步骤。[[20]](#ref-20)

## 17.4 OpenCode：开源、模型提供商无关的 Client/Server Runtime

### 17.4.1 版本边界：区分 OpenCode 1 与 OpenCode 2

OpenCode 官方文档当前同时说明 1.x 和 2.x 版本线。两条线使用相同的命令名，但配置结构与实现已有变化；本节以核验日的 OpenCode 2 官方文档和公开开发分支为准，不把 V1 的旧配置字段套用到 V2。仓库声明项目为开源并采用 MIT 许可证。[[30]](#ref-30)

### 17.4.2 Client/Server：把界面放在服务端之上

OpenCode V2 的 CLI 文档说明，默认会发现或启动当前用户的共享后台服务，本地客户端连接该服务；`--standalone` 使用私有服务，`--server` 连接指定服务。后台服务拥有会话、配置、集成、权限与工具执行，公开 API 支持程序化接入。旧版“启动 TUI 时启动一个服务”的描述不能直接套用于 V2 默认行为。[[31]](#ref-31)

这个结构使会话与模型运行不必由单一终端界面独占。服务端成为会话、请求和事件的接口边界，客户端则负责展示与交互。对比 Codex，二者都公开了服务/协议边界；对比闭源的 Claude Code，OpenCode 的具体协议和代码路径可以直接检视。

### 17.4.3 Provider 抽象：让模型接入成为可替换部分

OpenCode 的模型文档描述了多家 LLM Provider 与本地模型支持；Provider 层负责把选定模型信息和请求交给相应服务。模型目录与接口转换在代码中也有对应实现。[[32]](#ref-32)

这个架构让 Agent Harness 与某一个模型供应商不必完全绑定，但不能保证所有 Provider 的工具调用、上下文长度、流式行为和多模态支持完全一致。适配层把差异归纳为接口，不会消除底层协议差异；系统仍要处理模型能力声明和实际响应之间的不一致。

### 17.4.4 Session Runner 与 Tool Registry：模型循环和工具执行

OpenCode V2 的源码可看到 Session Runner、LLM 运行逻辑和工具服务的组织方式。工具系统负责注册工具、构造模型可见的定义、执行工具并规范化结果；Session Runner 再把模型输出和工具结果纳入会话过程。[[17]](#ref-17)

从源码结构可归纳出一次循环：

~~~text
接收 Session 输入
  → 准备选定 Provider/Model 与工具定义
  → 发送模型请求
  → 识别工具调用或文本输出
  → 执行工具并收集结果
  → 将结果追加到会话并继续，或结束当前回应
~~~

这条链是基于公开实现路径的系统概括。具体工具集合、模型协议转换、输出截断和异常恢复会随版本演进，应回到 V2 源码核对，不能用旧版的文件名或配置细节描述当前实现。

### 17.4.5 Agent 与 Permissions：把角色配置和动作裁决放在一起

OpenCode 文档区分 Primary Agent 与 Subagent。Primary Agent 面向用户承接主会话；Subagent 用于专项委派。Agent 配置可以指定模型、提示、可用工具和权限。权限支持 allow、ask、deny，并可按工具、Shell 命令、文件范围等规则细化。[[33]](#ref-33)

V2 文档进一步描述 Permissions 与 Policies。规则会在匹配资源时允许、拒绝或请求审批；具体判断层次与 V1 的权限字段并不完全相同。比较权限设计时应关注谁作出决定、规则如何匹配、拒绝是否可被更宽的配置覆盖，以及授权后由哪层执行隔离。[[18]](#ref-18)

Agent 角色本身不是安全沙箱。将编辑工具标为 deny 能缩小调用面，但系统还需要确认 Shell 能力、外部目录访问和插件工具是否在另一条路径上提供相同副作用。

### 17.4.6 会话、快照和回退

OpenCode V2 的配置文档说明，Agent 操作期间会通过快照跟踪文件变化，以支持会话内撤销与恢复。服务端也为 Session 提供消息、分叉、终止、差异、总结和回退等公开接口。[[31]](#ref-31) [[34]](#ref-34)

这些接口表明 Session 不只是聊天文本列表；它还要承载消息、工具过程、当前状态和可视化差异。但文件快照仍不能撤销所有外部副作用。数据迁移、API 请求和远端仓库操作需要单独的幂等、补偿或审计设计。

### 17.4.7 OpenCode 的实现特征

从 V2 公开实现看，OpenCode 把 Client/Server、Session Runner、Provider 适配、Tool Registry、权限策略和多种界面组织为开放系统。它的优势是读者可以顺着源码和协议验证设计；代价是跨 Provider 与多客户端支持带来额外适配和兼容工作。

可迁移的观察不是“任何开源 Agent 都天然更安全”，而是源码可见提升了实现可审计性。安全结论仍要检查有效配置、执行边界、依赖和实际运行路径。

### 17.4.8 启动过程：连接服务与装配会话

OpenCode V2 的默认启动方式是**发现或启动当前用户的共享后台服务，然后让本地客户端连接它**；该服务拥有会话、配置、集成、权限和工具执行。`--standalone` 使用私有服务，`--server` 连接指定 URL。不能把旧版本“每次 TUI 启动时附带一个服务”的描述直接套到 V2 默认行为。[[9]](#ref-9)

| 阶段 | 初始化或执行动作 | 完成后得到什么 |
| --- | --- | --- |
| 选择连接方式 | CLI 采用共享服务发现 / 启动，或按参数选择私有服务、指定服务 | 客户端到目标服务的连接；连接已有服务时无需重新启动后台进程 |
| 准备工作区配置 | 服务端按工作区处理生效配置、项目指令与相关集成 | 当前工作区的配置与指令范围；不是所有客户端共享同一项目上下文 |
| 选择 Agent 与模型 | 确定主 Agent 和 Provider / Model，使用已有凭证，缺少连接条件时通过 `/connect` 等入口补齐 | 可用于该会话的角色、模型和访问条件 |
| 建立或继续会话 | 经 Session 接口建立会话或读取已有状态，客户端提交输入 | 会话身份与消息；模型任务从输入提交后展开 |
| 执行与权限交互 | 服务端运行模型和工具过程，应用权限规则，需要决定时等待客户端 | 已执行的工具结果或明确的拒绝；模型据此继续 |
| 事件与持久化 | 客户端通过公开 API 和事件查看消息、状态及变更 | 可继续的会话和可审查结果；关闭某个界面不等于共享服务退出 |

上述顺序是依据服务职责、配置、Provider 与 Session 接口作出的依赖概括，不承诺各子系统的内部构造顺序。配置中声明 Provider 并不证明已持有有效凭证；`/connect` 完成账户接入后，还要选择模型。项目指令采用 `AGENTS.md`，下层目录的说明可随文件读取和目录探索加入；V2 不以 `CLAUDE.md` 作为回退。[[10]](#ref-10)[[11]](#ref-11)[[12]](#ref-12)[[35]](#ref-35)

权限也在服务端执行路径上继续发挥作用。V2 的 `permissions` 规则对动作与资源返回 `allow / deny / ask`，`ask` 等待客户端决定。因此，工具出现在目录中、服务已经启动、某个客户端已经连接，都不足以证明这次工具调用已经获准或完成。[[36]](#ref-36)[[37]](#ref-37)

## 17.5 三种产品的实现边界与设计取舍

### 17.5.1 对照表

| 维度 | Codex CLI | Claude Code | OpenCode |
| --- | --- | --- | --- |
| 核心代码可见性 | CLI/runtime 多个组件公开源码 | 核心程序闭源；官方文档公开产品行为 | 项目源码公开，MIT 许可证 |
| 本章证据 | 官方开源清单与 Rust 仓库 | 官方滚动文档与仓库许可证 | V2 文档与公开开发分支 |
| 主循环 | Session/Turn 驱动采样、工具执行和事件 | 官方描述收集上下文、行动、验证循环 | V2 Session Runner 处理模型和工具过程 |
| 前端边界 | Rust Core/TUI 与 App Server 协议 | CLI、IDE、桌面、云等接口；核心实现不可审计 | 服务端与 TUI/Desktop/IDE 等客户端 |
| 模型接入 | 以 OpenAI/Codex 产品路径为主，仓库还含适配组件 | Claude 模型和官方支持的部署路径 | 多 Provider 与本地模型抽象 |
| 工具和权限 | Tool Router、审批、环境与沙箱分层 | 官方公开工具、权限规则、模式和 Hook 行为 | Tool Registry、Agent 权限和 V2 Policies |
| 会话与上下文 | 会话历史、Turn 状态和压缩逻辑可在仓库中跟踪 | 官方公开 Session、上下文和压缩行为 | Session API、消息事件、快照与恢复可审计 |
| 结论可信度 | 可对照具体代码，但主分支会变化 | 只能对官方公开机制下结论 | V2 代码可见，需注意 V1/V2 差异 |

表格比较的是公开资料能支持的实现边界，不是产品能力排名。尤其是 Claude Code 一栏：文档足以解释外部可见行为，却不足以审计核心算法和私有内部模块。

### 17.5.2 开源能回答“怎么实现”，不能单独回答“是否可靠”

公开源码让人可以检查工具执行、策略配置、恢复行为和错误处理，但代码存在不代表部署时启用了正确配置；默认功能存在也不代表每个版本和入口行为相同。源码审查还不能替代运行时验证、依赖审查、权限测试和真实故障演练。

闭源产品也不是完全不可分析。官方行为文档、协议说明、配置接口和故障边界仍能支持严谨的系统比较。需要做的是把“可观察事实”和“内部猜测”分开，而不是用开源标签替代证据质量。

### 17.5.3 成熟 Coding Agent 的共同设计经验

从三个实现及公开接口中，可以归纳出几条工程原则：

1. **把 Session、Turn 和工作区状态分开。** 会话记录是上下文资产，当前工作区是外部事实，当前轮次是短时执行状态。
2. **把模型计划与工具执行分开。** 模型提出行动，运行时解析、授权、执行并记录结果。
3. **让拒绝与失败成为一等结果。** 被用户拒绝、被策略拒绝、工具失败、网络失败和任务取消都必须被准确反馈。
4. **将审批与隔离分层。** 审批决定是否允许某类操作，隔离限制实际进程的影响范围。
5. **设计有损压缩后的恢复流程。** 摘要保存目标、证据、差异、验证和待办；恢复时仍需回到仓库核实。
6. **委派要明确上下文与写入归属。** 子 Agent 的总结不是最终证据；并行修改还要解决文件冲突与集成验证。
7. **验证当前差异而非接受模型自述。** 报告“测试通过”时要能追溯命令、退出状态和对应代码版本。
8. **让可审查的交付成为主界面。** diff、命令结果、权限决策和未覆盖范围应能被人核对。

### 17.5.4 如何阅读一个 Coding Agent 仓库

如果要分析另一个 Coding Agent，可沿这些问题追源码，而不是从目录树开始随机浏览：

| 问题 | 应找的代码或文档 |
| --- | --- |
| 一轮会话如何开始和结束？ | Session/Turn 入口、请求队列、取消路径 |
| 模型请求如何携带上下文？ | Prompt/context 构造、历史存储、压缩逻辑 |
| 工具如何注册和执行？ | 工具 schema、路由器、执行器、结果格式 |
| 谁能批准或拒绝操作？ | 权限规则、审批请求、Hook、策略优先级 |
| 进程实际受哪些资源限制？ | 沙箱、文件路径控制、网络策略、子进程入口 |
| 多 Agent 如何通信和隔离？ | 子任务创建、消息投递、工作区、取消及晚到结果 |
| 什么证据能证明完成？ | 测试结果、diff、会话状态、交付协议 |
| 哪些功能只在托管环境可用？ | 云端 API、认证层、服务端调用和开源清单 |

这套阅读顺序先定位运行闭环，再追踪关键边界。公开代码会随着版本改变，因此结论应绑定提交、版本或核验日期；没有公开证据时，明确写“未知”比补造一套内部架构更可靠。

### 17.5.5 启动就绪条件与首轮执行边界

[![Codex、Claude Code 与 OpenCode 的启动时序图：参与者生命线、请求与响应、条件分支及模型工具循环](../images/coding-agent-startup-comparison.png)](../images/coding-agent-startup-comparison.svg)

*启动时序对照。三幅时序分别沿生命线向下阅读：实线为调用或请求，虚线为响应或结果，橙色为通知或界面展示；`alt` 表示条件分支，`opt` 表示可选 Hook，`loop` 表示模型与工具的重复交互。参与者是逻辑角色，不保证对应独立进程；Claude Code 仅描述官方可观察顺序。图中正常路径仍可能在认证、连接、Hook 或审批处等待或失败。标注含义见 17.1.4，点击图片查看可缩放 SVG。*

| 观察点 | Codex App Server | Claude Code 本地会话 | OpenCode V2 |
| --- | --- | --- | --- |
| 进程或服务可用 | 服务启动并能接受客户端连接 | CLI / Harness 已启动，界面可以显示 | 共享、私有或指定服务可被客户端连接 |
| 会话可用 | Thread 已建立或恢复，工作目录和执行设置已确定 | Session 已建立或恢复，适用指令与启动 Hook 参与上下文准备 | 服务端已确定工作区、Session、Agent 与模型条件 |
| 本轮可执行 | 连接已握手，`turn/start` 被接受；工具执行仍可能等待审批 | 用户输入经过启动等待与本轮处理后进入循环；工具仍逐次受控 | 输入已提交，模型与工具条件满足；`ask` 仍会暂停动作 |
| 完成的证据 | Turn 状态、Item、文件差异与实际检查结果 | Session 中的工具结果、文件差异与检查结果 | Session 消息、工具结果、事件与实际工作区差异 |

据此，[第14章 14.2.1](../part3/03-agent-runtime.md#execution-loop) 的七步是**建设 Runtime 时的参考职责**，三款产品可以把它们拆在进程、工作区、会话和轮次四个阶段，部分能力延迟准备。不能据此断言每款产品都有七个同名模块、启动时会探测全部外部服务，或都会先运行验证器再接收输入。实践中应分别检查连接、会话、模型请求与工具执行的可用条件。

## 本章小结

Codex CLI、Claude Code 与 OpenCode 都把模型放进了包含上下文、工具和验证的工程循环，但源码透明度不同：Codex CLI 与 OpenCode 可以沿公开代码追踪运行时，Claude Code 只能根据官方文档分析其外部机制与配置边界。

三者共同说明，成熟 Coding Agent 的核心并不是某个特别长的系统提示词，而是把任务、状态、权限、执行、恢复和验证组织成一个能被用户审查的闭环。要评价系统是否可靠，应检查工具真正做了什么、动作被限制在哪里、任务状态如何恢复，以及完成结论有何证据。

## 参考资料

1. <a id="ref-1"></a> OpenAI，[Open Source](https://learn.chatgpt.com/docs/open-source)，Codex 开源组件清单与非开源组件说明，访问日期：2026-10-22。
2. <a id="ref-2"></a> OpenAI，[Codex App Server](https://developers.openai.com/codex/app-server)，连接握手、Thread / Turn / Item、事件及 Code Mode host；持续更新，核验于 2026-10-22。
3. <a id="ref-3"></a> OpenAI，[Config basics](https://developers.openai.com/codex/config-basic)，分层配置与项目可信状态；持续更新，核验于 2026-10-22。
4. <a id="ref-4"></a> OpenAI，[Custom instructions with AGENTS.md](https://developers.openai.com/codex/guides/agents-md)，指令发现与加载边界；持续更新，核验于 2026-10-22。
5. <a id="ref-5"></a> Anthropic，[How Claude Code works](https://code.claude.com/docs/en/how-claude-code-works)，Agent Loop、会话、上下文、压缩、权限与 Checkpoint，访问日期：2026-10-22。
6. <a id="ref-6"></a> Anthropic，[Settings files and precedence](https://code.claude.com/docs/en/settings)，配置作用域与生效关系；持续更新，核验于 2026-10-22。
7. <a id="ref-7"></a> Anthropic，[How Claude remembers your project](https://code.claude.com/docs/en/memory)，项目指令与 Auto memory 的加载；持续更新，核验于 2026-10-22。
8. <a id="ref-8"></a> Anthropic，[Extend Claude with skills](https://code.claude.com/docs/en/skills)，Skill 目录与按需加载；持续更新，核验于 2026-10-22。
9. <a id="ref-9"></a> OpenCode，[V2 CLI](https://opencode.ai/v2/docs/cli/)，共享后台服务及 standalone / server 连接模式；持续更新，核验于 2026-10-22。
10. <a id="ref-10"></a> OpenCode，[V2 Config](https://opencode.ai/v2/docs/config/)，全局、项目、系统配置与默认 Agent；持续更新，核验于 2026-10-22。
11. <a id="ref-11"></a> OpenCode，[V2 Providers](https://opencode.ai/v2/docs/providers/)，Provider 目录、凭证与模型选择；持续更新，核验于 2026-10-22。
12. <a id="ref-12"></a> OpenCode，[V2 Instructions](https://opencode.ai/v2/docs/instructions/)，AGENTS.md 的初始发现与按需加载；持续更新，核验于 2026-10-22。
13. <a id="ref-13"></a> OpenAI，[App Server 启动](https://github.com/openai/codex/blob/86e17e98c7f901277aeeb4649cf1b01462b082f2/codex-rs/app-server/src/lib.rs)、[消息处理与初始化检查](https://github.com/openai/codex/blob/86e17e98c7f901277aeeb4649cf1b01462b082f2/codex-rs/app-server/src/message_processor.rs)与 [ThreadManager](https://github.com/openai/codex/blob/86e17e98c7f901277aeeb4649cf1b01462b082f2/codex-rs/core/src/thread_manager.rs)，运行时组成和启动图的源码依据，固定提交 `86e17e98c7f9`；核验于 2026-10-22。
14. <a id="ref-14"></a> Anomaly Innovations，[ServerConnection.resolve](https://github.com/anomalyco/opencode/blob/c93aaae37157d785f76a8e9a9243a6560e286468/packages/cli/src/services/server-connection.ts) 与 [Standalone.start](https://github.com/anomalyco/opencode/blob/c93aaae37157d785f76a8e9a9243a6560e286468/packages/cli/src/services/standalone.ts)，V2 连接路径、共享服务与私有服务启动，固定提交 `c93aaae37157`；核验于 2026-10-22。
15. <a id="ref-15"></a> OpenAI，[codex-rs/core/src/session/turn.rs](https://github.com/openai/codex/blob/86e17e98c7f901277aeeb4649cf1b01462b082f2/codex-rs/core/src/session/turn.rs) 与 [state/turn.rs](https://github.com/openai/codex/blob/main/codex-rs/core/src/state/turn.rs)，Turn 主循环和轮次状态，访问日期：2026-10-22。
16. <a id="ref-16"></a> OpenAI，[Tool Router](https://github.com/openai/codex/blob/86e17e98c7f901277aeeb4649cf1b01462b082f2/codex-rs/core/src/tools/router.rs)、[Sandboxing](https://github.com/openai/codex/blob/main/codex-rs/core/src/tools/sandboxing.rs) 与 [Approvals](https://github.com/openai/codex/blob/main/codex-rs/core/src/tools/approvals.rs)，工具路由、审批与沙箱职责，访问日期：2026-10-22。
17. <a id="ref-17"></a> Anomaly Innovations，[Session Runner](https://github.com/anomalyco/opencode/blob/c93aaae37157d785f76a8e9a9243a6560e286468/packages/core/src/session/runner/llm.ts) 与 [Core tool service](https://github.com/anomalyco/opencode/blob/c93aaae37157d785f76a8e9a9243a6560e286468/packages/core/src/tool.ts)，V2 开发分支中的模型循环与工具注册/执行，访问日期：2026-10-22。
18. <a id="ref-18"></a> OpenCode，[Policies](https://opencode.ai/v2/docs/policies/) 与 [Permissions](https://opencode.ai/v2/docs/permissions/)，V2 权限和策略规则，访问日期：2026-10-22。
19. <a id="ref-19"></a> Anthropic，[Configure permissions](https://code.claude.com/docs/en/permissions)，权限模式和规则，访问日期：2026-10-22。
20. <a id="ref-20"></a> Anthropic，[Hooks reference](https://code.claude.com/docs/en/hooks)，SessionStart、启动等待与 Setup 的触发边界；持续更新，核验于 2026-10-22。
21. <a id="ref-21"></a> OpenAI，[codex-rs/core/README.md](https://github.com/openai/codex/blob/main/codex-rs/core/README.md)，Codex Core 的职责与支持范围，访问日期：2026-10-22。
22. <a id="ref-22"></a> OpenAI，[codex-rs/core/src/session/turn.rs](https://github.com/openai/codex/blob/86e17e98c7f901277aeeb4649cf1b01462b082f2/codex-rs/core/src/session/turn.rs)，模型采样、工具调用和助手消息的执行循环，访问日期：2026-10-22。
23. <a id="ref-23"></a> OpenAI，[codex-rs/protocol/README.md](https://github.com/openai/codex/blob/main/codex-rs/protocol/README.md)，Core、TUI 与 App Server 共用的协议类型，访问日期：2026-10-22。
24. <a id="ref-24"></a> OpenAI，[codex-rs/app-server/README.md](https://github.com/openai/codex/blob/main/codex-rs/app-server/README.md)，App Server 接入和协议，访问日期：2026-10-22。
25. <a id="ref-25"></a> OpenAI，[Context window](https://github.com/openai/codex/blob/main/codex-rs/core/src/session/context_window.rs)、[Compaction](https://github.com/openai/codex/blob/main/codex-rs/core/src/compact.rs)，上下文窗口和历史压缩，访问日期：2026-10-22。
26. <a id="ref-26"></a> Anthropic，[Claude Code GitHub repository](https://github.com/anthropics/claude-code)，公开仓库中的产品说明与资料范围，访问日期：2026-10-22。
27. <a id="ref-27"></a> Anthropic，[Claude Code LICENSE.md](https://github.com/anthropics/claude-code/blob/main/LICENSE.md)，版权和适用条款，访问日期：2026-10-22。
28. <a id="ref-28"></a> Anthropic，[Extend Claude Code](https://code.claude.com/docs/en/features-overview)，Skills、MCP、Hooks、Subagents 与插件的公开职责，访问日期：2026-10-22。
29. <a id="ref-29"></a> Anthropic，[Create custom subagents](https://code.claude.com/docs/en/sub-agents)，子 Agent 的上下文与配置，访问日期：2026-10-22。
30. <a id="ref-30"></a> Anomaly Innovations，[OpenCode repository](https://github.com/anomalyco/opencode)，OpenCode 开源项目与许可证信息，访问日期：2026-10-22。
31. <a id="ref-31"></a> OpenCode，[V2 CLI](https://opencode.ai/v2/docs/cli/) 与 [V2 API](https://opencode.ai/v2/docs/api)，共享后台服务、Client/Server 结构与 Session API，访问日期：2026-10-22。
32. <a id="ref-32"></a> OpenCode，[Models](https://opencode.ai/v2/docs/models/) 与 [Providers](https://opencode.ai/v2/docs/providers/)，Provider 与模型配置，访问日期：2026-10-22。
33. <a id="ref-33"></a> OpenCode，[Agents](https://opencode.ai/v2/docs/agents/) 与 [Tools](https://opencode.ai/v2/docs/tools/)，Primary Agent、Subagent 与工具能力，访问日期：2026-10-22。
34. <a id="ref-34"></a> OpenCode，[Config](https://opencode.ai/v2/docs/config/) 与 [V2 API](https://opencode.ai/v2/docs/api)，文件快照、Session 差异与恢复接口，访问日期：2026-10-22。
35. <a id="ref-35"></a> OpenCode，[V2 API](https://opencode.ai/v2/docs/api)，服务端 Session、消息及事件接口；持续更新，核验于 2026-10-22。
36. <a id="ref-36"></a> OpenCode，[V2 Permissions](https://opencode.ai/v2/docs/permissions/)，动作与资源裁决及客户端审批；持续更新，核验于 2026-10-22。
37. <a id="ref-37"></a> OpenCode，[V2 Tools](https://opencode.ai/v2/docs/tools/)，文件、搜索、Shell 等工具与权限对应关系；持续更新，核验于 2026-10-22。
