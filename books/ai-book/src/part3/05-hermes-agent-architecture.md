# 第16章 Hermes Agent 架构解析：自我进化、记忆与多入口 Agent Gateway

> Hermes Agent 的核心价值，不是“多一个聊天入口”，而是把长期运行的 Agent 做成一个会积累记忆、沉淀技能、跨入口工作、可扩展工具并能生成训练轨迹的个人运行时。

## 引言

前几章分别分析了 Agent 平台、Coding Agent、Pi Runtime 和 OpenClaw。Pi 让我们看到终端原生 Coding Agent 可以被拆成可嵌入、可扩展的 Runtime；OpenClaw 让我们看到个人 AI 助手可以从聊天机器人升级为 Gateway：它把 WhatsApp、Telegram、Slack、WebChat、CLI 等入口接到同一个 Agent Runtime。

Hermes Agent 和 OpenClaw 处在相近的问题空间，但设计重心不同：

- OpenClaw 更像一个个人 Agent Gateway，强调多渠道接入、插件生态和本地优先；
- Hermes Agent 更像一个长期运行、自我进化的 Agent Runtime，强调记忆、技能、工具、子 Agent、执行后端和研究数据闭环。

如果用一句话概括：

```text
OpenClaw 解决“Agent 如何到达用户所在的地方”
Hermes 解决“Agent 如何在长期使用中变得更懂用户、更会做事”
```

本章基于 2026 年 5 月 15 日可访问的 Hermes Agent 官方 README 与文档进行分析。Hermes Agent 正在快速演进，工具注册表、平台适配器和闭环学习能力仍在持续变化，因此本章尽量使用“数十个内置工具”“持续增长的工具集”这类稳健表述，而不是绑定某个容易过期的精确数量。

本章关注六个问题：

1. Hermes Agent 为什么要做成长运行的多入口 Agent？
2. 它的底层架构由哪些核心组件组成？
3. 用户输入、规划、工具执行、外部环境和记忆回流如何形成完整数据流？
4. Memory、Skills、Session Search 和 Context Files 如何协作？
5. 工具中心、执行引擎、Gateway、Cron、ACP、Profiles 如何把 Agent 从聊天扩展成工作平台？
6. 如果我们自研个人或团队 Agent，可以借鉴哪些设计？

---

## 16.1 系统定位：从一次性会话到长期运行的 Agent

很多 AI 产品仍然是“一次性会话”：

```text
User Prompt -> LLM -> Answer
```

这种形态适合问答，但不适合真正的助手。真实助手需要具备连续性：

- 记得用户偏好；
- 记得项目背景；
- 记得过去解决过什么问题；
- 能把一次复杂任务沉淀成可复用技能；
- 能从 CLI、Telegram、Slack、Discord、WhatsApp、Email 等入口继续同一类工作；
- 能在本地、Docker、SSH、Modal、Daytona、Singularity 等环境里执行任务；
- 能把运行轨迹导出，用于评估、微调或强化学习。

Hermes Agent 的定位可以抽象成：

```text
Hermes Agent
  = Long-running Agent Runtime
  + Persistent Memory
  + Procedural Skill System
  + Multi-platform Gateway
  + Tool / Toolset Registry
  + Execution Backends
  + Research Trajectory Pipeline
```

它和普通聊天机器人的本质区别是：普通聊天机器人围绕“单次回复”设计，Hermes 围绕“长期能力增长”设计。

### 关键判断：Agent 的能力不只来自模型

Hermes 的设计隐含了一个重要判断：

> 长期 Agent 的能力，不只来自模型参数，而来自模型、记忆、技能、工具、入口、执行环境和历史轨迹共同组成的系统。

同一个模型，如果每次都从空白上下文开始，就是普通聊天；如果它能读取项目规则、调用工具、搜索旧会话、更新记忆、创建技能、定时执行任务，并在不同平台保持身份连续性，就开始接近真正的个人 Agent。

---

## 16.2 底层架构总览：六大核心组件

如果把 Hermes Agent 抽象成一套长期运行的 Agent Runtime，它的底层架构可以先用六个核心组件理解：

```mermaid
flowchart TB
    Input["用户输入 / 平台事件 / 定时任务"] --> Brain["大脑中枢<br/>LLM / Prompt / 推理"]
    Brain --> Planner["小脑<br/>规划 / 状态 / 工作流 / 反思"]
    Planner --> Tools["工具中心<br/>Tool Registry / Toolsets / MCP / Skills"]
    Tools --> Action["执行引擎<br/>解析 / 调度 / 结果处理 / 重试"]
    Action --> Env["外部环境<br/>Gateway / Cron / ACP / Backends"]
    Env --> Action
    Action --> Memory["记忆系统<br/>Memory / Sessions / Skills / Profiles"]
    Memory --> Brain
    Memory --> Planner
```

这张图比源码模块更抽象，但更适合理解 Hermes 的设计意图。这里的“大脑中枢”“小脑”“工具中心”等说法是分析性抽象，不是 Hermes 源码里的官方模块命名：

| 组件 | 解决的问题 | Hermes 中的代表实现 |
|:---|:---|:---|
| 大脑中枢 | 理解输入、生成推理、决定下一步行动 | LLM Provider、Prompt Builder、Context Compressor、Callbacks |
| 记忆系统 | 保存长期事实、历史会话、用户偏好和可复用经验 | Persistent Memory / User Profile、SQLite Sessions + FTS5、Skills、Profiles |
| 小脑 | 把复杂任务拆成步骤，维持状态，必要时反思和重规划 | AIAgent Loop、Cron 任务配置、脚本化/服务化调用入口、Context Compressor |
| 工具中心 | 定义 Agent 能使用哪些能力，以及这些能力如何注册和治理 | Tool Registry、Toolsets、Plugins、MCP Tools、Skills |
| 执行引擎 | 把模型输出的工具调用变成真实执行，并处理结果、异常和回退 | Tool Dispatch、Execution Backends、Streaming Callbacks、Result Persistence |
| 外部环境 | 让 Agent 接入真实世界，包括消息平台、IDE、文件系统、远程环境和自动化任务 | CLI / TUI、Messaging Gateway、ACP、Cron、local / Docker / SSH / Modal |

### 16.2.1 大脑中枢：LLM、Prompt 与推理核心

大脑中枢不是单个模型调用，而是模型、系统提示、上下文压缩、Provider 选择和流式回调的组合。它负责理解用户意图、抽取关键信息、生成任务计划、选择工具，并把下一步行动表达成 Runtime 能执行的结构。

在 Hermes 里，这一层最重要的工程点是 Prompt Builder。它不是把所有材料拼到一起，而是把人格、长期记忆、用户画像、项目上下文、相关 Skills、工具说明和会话状态组织成一个稳定的上下文快照。这样模型每次进入推理时，看到的是经过筛选和分层的信息，而不是随机堆叠的文本。

### 16.2.2 记忆系统：Memory、Session Search 与长期上下文

记忆系统解决的是“Agent 如何连续存在”。Hermes 不是只依赖当前对话窗口，而是把长期事实、用户偏好、历史会话和可复用操作流程分层保存：

- Persistent Memory 与 User Profile 保存每次会话都应该知道的关键事实；
- SQLite Sessions + FTS5 保存历史会话细节，供需要时检索；
- Skills 保存“下次怎么做”的程序性经验；
- Profiles 隔离不同身份、项目、客户或环境。

这使 Hermes 的记忆不是一个无限知识库，而是一套有预算、有检索、有边界的上下文系统。

### 16.2.3 小脑：规划、状态、工作流与反思

小脑负责把“我要完成什么”转成“接下来怎么做”。长期 Agent 不能只靠单轮 ReAct，因为真实任务经常包含多步执行、长时间等待、工具失败、用户打断、上下文压缩和跨会话恢复。

Hermes 的规划能力更多是隐式分布在 AIAgent Loop、Cron、脚本化/服务化调用入口、Context Compressor 和 Skills 中：模型负责提出步骤，Runtime 负责维持会话、工具结果和执行状态，Skills 把稳定成功路径沉淀成可复用流程。对企业级 Agent 来说，这一层还需要进一步显式化，例如引入任务状态机、审批节点、暂停恢复、失败补偿和可回放执行轨迹。

### 16.2.4 工具中心：Tool Registry、Toolsets、MCP 与 Skills

工具中心回答“Agent 能做什么，以及谁允许它做”。Hermes 把能力组织成 Tool Registry 和 Toolsets：Registry 管理工具 schema、可用性和分发；Toolsets 把 web、terminal、file、browser、memory、session_search、cronjob、delegation、MCP 等能力按平台、profile 和任务类型打包。

Skills 在这里有双重身份：它既是记忆系统中的程序性经验，也是工具中心中的能力说明。一个好的 Skill 不只是提示词片段，而是包含触发条件、操作步骤、依赖工具、约束和验证方式的可复用任务协议。

### 16.2.5 执行引擎：指令解析、调度、结果处理与失败恢复

执行引擎回答“模型选择了工具以后，系统如何可靠地行动”。它需要解析模型输出的 tool call，校验参数，选择执行后端，调度工具调用，捕获 stdout / stderr / API result，向用户流式展示进度，并把结果写回会话状态。

这层的关键不是“能不能调用工具”，而是失败处理：工具超时怎么办，参数不合法怎么办，执行后端不可用怎么办，危险命令是否需要确认，结果太长是否要截断，失败是否允许重试或降级。没有执行引擎，工具调用只是 demo；有了执行引擎，Agent 才能进入长期运行。

### 16.2.6 外部环境：Gateway、Cron、ACP 与执行后端

外部环境是 Hermes 和真实世界交互的边界。它包括 CLI / TUI、Messaging Gateway、ACP / IDE Integration、Cron、脚本化/服务化调用入口，也包括 local、Docker、SSH、Daytona、Modal、Singularity 等执行后端。

这一层让 Hermes 不再只是“聊天窗口里的 Agent”，而是可以在用户所在的平台里工作、在后台定时执行任务、在 IDE 中协作、在隔离环境里运行命令。它同时也是风险最大的层，因为越靠近真实环境，越需要授权、审计、隔离和人工确认。

### 16.2.7 工程分层视角

如果从源码和运行时模块看，Hermes Agent 可以进一步分成八个核心层：

```mermaid
flowchart TB
    subgraph Entry["Entry Points"]
        CLI["CLI / TUI"]
        Gateway["Messaging Gateway"]
        ACP["ACP / IDE Integration"]
        Cron["Cron Jobs"]
        Batch["脚本化 / 服务化入口"]
    end

    subgraph Core["Agent Core"]
        Agent["AIAgent Loop"]
        Prompt["Prompt Builder"]
        Provider["Provider Resolver"]
        Compressor["Context Compressor"]
        Callbacks["Callbacks / Streaming"]
    end

    subgraph Context["Context & Learning"]
        Memory["Persistent Memory / User Profile"]
        Sessions["SQLite Sessions + FTS5"]
        Skills["Skills / SKILL.md"]
        ContextFiles["AGENTS.md / CLAUDE.md / .cursorrules / 其他项目规则"]
        Profiles["Profiles"]
    end

    subgraph Tools["Tool Runtime"]
        Registry["Tool Registry"]
        Toolsets["Toolsets"]
        MCP["MCP Tools"]
        Plugins["Plugins"]
    end

    subgraph Execution["Execution Backends"]
        Local["Local"]
        Docker["Docker"]
        SSH["SSH"]
        Daytona["Daytona"]
        Modal["Modal"]
        Singularity["Singularity"]
    end

    subgraph Storage["State Storage"]
        Config["config.yaml"]
        StateDB["state.db"]
        SkillStore["~/.hermes/skills"]
        MemoryStore["~/.hermes/memories"]
    end

    Entry --> Agent
    Agent --> Prompt
    Prompt --> Context
    Agent --> Provider
    Agent --> Registry
    Registry --> Toolsets
    Registry --> MCP
    Registry --> Plugins
    Toolsets --> Execution
    Agent --> Compressor
    Agent --> Callbacks
    Context --> Storage
    Agent --> Storage
```

这张图背后有三个核心分离：

| 分离点 | 设计含义 |
|:---|:---|
| Entry 与 Core 分离 | CLI、Gateway、ACP、Cron 都复用同一个 Agent Core |
| Context 与 Tools 分离 | 记忆和技能决定“知道什么”，工具系统决定“能做什么” |
| Toolsets 与 Execution 分离 | 同一个 terminal 工具可以跑在 local、Docker、SSH 或云端后端 |

这种分层让 Hermes 不只是一个 CLI 工具，而是一个可嵌入不同入口、不同执行环境、不同研究工作流的 Agent Runtime。

### 16.2.8 与第5章组件地图的对应关系

Hermes 的重点不是“单次任务执行”，而是长期运行、自我进化和多入口接入。用第 5 章组件地图看，Hermes 对 Memory、Skills、Toolsets、Profiles、Research Pipeline 和 Learning Loop 覆盖很强；但如果要进入企业生产，还需要额外补强审批、合规审计、发布门禁和严格 Eval Harness。

| 第5章组件 | Hermes 中的实现方式 | 实现状态与差异 |
|:---|:---|:---|
| Event & Intake Router | CLI / TUI、Messaging Gateway、ACP / IDE、Cron、脚本化/服务化调用入口 | 强实现；多入口是 Hermes 从工具变成服务的关键 |
| Intent Normalizer | Prompt System、Profiles、入口类型和上下文文件共同塑造任务边界 | 部分实现；更偏运行时上下文塑形，不一定显式生成任务契约 |
| Task Planner | AIAgent Loop 中的任务推进，Cron / 脚本化入口可形成固定任务计划 | 隐式到中等实现；计划存在于 loop、prompt 和任务配置中 |
| Context Builder | Prompt Builder、Context Compressor、AGENTS.md / `CLAUDE.md` / `.cursorrules` 等上下文文件、Profiles | 强实现；上下文稳定性是 Hermes 的核心设计 |
| Memory Layer | Persistent Memory、User Profile、SQLite Sessions + FTS5、Profiles、Session Search | 强实现；长期事实、偏好、会话搜索和身份隔离是重点 |
| Execution State & Checkpoint | SQLite state、sessions、callbacks、streaming、tool result persistence | 中等到强；支持回放和研究，但严格 checkpoint / rollback 需要上层治理 |
| Capability Registry | Tool Registry、Toolsets、MCP Tools、Plugins、Skills | 强实现；Toolsets 是权限和能力分组的核心单位 |
| Policy Engine & Human Control Plane | Toolsets、Profiles、Execution Backends、Security 分层 | 中等实现；能力分组清晰，但企业审批和人工接管流程需要强化 |
| Agent Loop | AIAgent Loop 统一多入口、模型调用、工具分发、流式回调和状态持久化 | 强实现；这是 Hermes 的运行核心 |
| Model Router & Handoff Manager | Provider Resolver、Profiles、不同入口和执行后端 | 中等实现；模型选择清晰，专家委派和多 Agent 协作不是主要叙事 |
| Verifier & Eval Harness | Research Pipeline、轨迹数据、失败样本、技能演化约束 | 部分到中等实现；更偏研究与改进数据，生产 release gate 仍需补齐 |
| Review Surface、Trace & Audit | callbacks、streaming、SQLite sessions、Gateway 消息、轨迹资产 | 中等实现；可观测材料丰富，但合规审计口径需要工程化 |
| Learning Loop | Skills 演化、Memory 更新、Research Pipeline、轨迹到训练数据 | 强实现；Hermes 最鲜明的差异点，但必须受验证和 owner review 约束 |

Hermes 因此特别适合作为第 5 章 Learning Loop、Memory Layer 和长期 Agent 的系统案例。它提醒我们：Agent 的能力会随着记忆、技能和轨迹积累而增长，但这种增长必须被验证器、评测集和人工审核约束，否则“自我进化”很容易变成“错误经验自动固化”。

---

## 16.3 核心数据流：从用户输入到记忆回流

从运行时看，Hermes 的一次任务不是简单的 `Prompt -> Answer`，而是一条带状态、工具、外部环境和记忆回流的数据链路：

```text
用户输入 / 平台事件 / Cron
  -> 入口标准化
  -> 大脑中枢理解任务
  -> 小脑拆解计划
  -> 工具中心选择能力
  -> 执行引擎调度工具
  -> 外部环境返回结果
  -> 执行引擎整理观察结果
  -> 大脑中枢继续推理或输出
  -> 记忆系统按需沉淀事实、会话和技能
```

这条链路里有三类数据流：

| 数据流 | 内容 | 关键风险 |
|:---|:---|:---|
| 任务流 | 用户意图、平台事件、Cron 任务、Slash Command | 入口信息不完整，任务边界不清 |
| 执行流 | tool call、执行后端、工具结果、错误和重试 | 高风险命令、超时、参数错误、结果过长 |
| 学习流 | Memory 更新、Session 归档、Skill 候选、Trajectory 数据 | 错误经验固化、隐私泄露、跨 profile 串线 |

Hermes 的关键点是：**Memory 不只是输入层，也在输出后参与回流。** 每次任务完成后，系统可以把稳定事实写入 persistent memory，把会话写入 SQLite，把可复用过程沉淀成 Skill，把执行轨迹交给 Research Pipeline。这样，Agent 的能力增长不依赖模型参数立即改变，而依赖 Runtime 中的上下文、技能和数据资产持续演进。

但记忆回流必须受约束。不是所有结果都应该写入长期记忆，也不是所有成功路径都应该变成 Skill。可靠的长期 Agent 需要在写入前判断：

- 这条信息是否长期有效；
- 是否属于当前 profile；
- 是否包含凭据、隐私或敏感业务数据；
- 是否经过工具结果或用户确认验证；
- 是否应该进入 Memory、Session、Skill，还是只作为本轮临时上下文。

这个判断决定了 Hermes 这类系统能否长期稳定运行。没有回流，Agent 每次都从头开始；没有约束，Agent 会把错误、噪声和越权信息永久化。

### 16.3.1 一个消息任务的端到端路径

如果把抽象数据流落到一个具体例子里，可以把一条 Telegram 消息在 Hermes 中的处理路径简化为：

1. 用户在 Telegram 中发来请求，例如“帮我检查这个仓库今天的 CI 失败原因”；
2. Gateway 适配器接收消息，并根据用户、线程和 profile 把它路由到正确 session；
3. Prompt Builder 组装人格、长期记忆、用户画像、相关 Skills、项目上下文和工具说明；
4. 模型先判断是否需要调用 GitHub、web、terminal 或 file 等工具；
5. Runtime 在对应 toolset 和执行后端上执行工具调用，并通过 callbacks 向用户流式反馈进度；
6. 工具结果回到 Agent Loop，模型继续推理，决定是追加调用、请求确认，还是直接给出答案；
7. 会话内容写入 session store；只有稳定事实才进入 persistent memory，只有经过验证的流程才进入 Skill 候选。

这个例子说明，Hermes 的核心不在“消息平台接进来了”，而在“平台入口、上下文构建、工具执行、状态持久化和能力沉淀”被串成了一条统一链路。

如果要把这条链路画成一张更适合读者快速浏览的时序图，可以简化为：

```mermaid
sequenceDiagram
    participant U as 用户 / 平台入口
    participant G as Gateway / 入口适配层
    participant S as Session Router / 会话路由
    participant P as Prompt Builder / 上下文构建
    participant M as Model / 推理核心
    participant T as Tool Runtime / 工具运行时
    participant E as Backend / 外部环境
    participant D as Session Store / 状态存储
    participant L as Memory & Skills / 能力沉淀层

    U->>G: 发送请求 / 平台事件
    G->>S: 标准化消息 + 识别用户/线程/profile
    S->>P: 加载当前 session 与上下文边界
    P->>M: 注入人格、记忆、技能、上下文文件、工具边界
    M->>T: 产生工具调用 / 或直接回答
    T->>E: 在 local / Docker / SSH / MCP 等环境执行
    E-->>T: 返回结果 / 错误 / 观察
    T-->>M: 结构化观察结果
    M-->>G: 最终回答 / 继续请求工具
    G-->>U: 流式反馈进度与结果
    T->>D: 持久化工具结果与会话状态
    D->>L: 生成 session / memory / skill 候选
    L-->>P: 下次会话按需回流
```

---

## 16.4 Agent Loop：统一多入口的运行核心

Hermes 的核心是一个同步编排引擎，可以理解为：

```text
load profile
  -> load config / memory / skills / context files
  -> assemble system prompt
  -> resolve provider and model
  -> receive user turn
  -> call model
  -> dispatch tool calls
  -> stream progress via callbacks
  -> persist session and tool results
  -> compress context when needed
  -> update memory / skills when appropriate
```

伪代码如下：

```python
def run_turn(user_message, profile, entry_point):
    config = load_config(profile)
    memory = load_memory(profile)
    skills = select_relevant_skills(user_message, profile)
    context_files = discover_context_files()
    sessions = load_session_state(entry_point, profile)

    prompt = build_prompt(
        personality=config.personality,
        memory=memory,
        skills=skills,
        context_files=context_files,
        tools=enabled_toolsets(entry_point),
        session=sessions.current,
    )

    while not done:
        response = model.complete(prompt)
        if response.tool_calls:
            results = tool_registry.dispatch(response.tool_calls)
            callbacks.stream_tool_results(results)
            prompt = append_observations(prompt, results)
            persist(results)
        else:
            callbacks.stream_answer(response.text)
            persist(response)
            done = True
```

这个循环和第 13 章 Coding Agent 的循环很像，但 Hermes 多了三个面向长期运行的能力：

- **Prompt Assembly**：每次会话开始时把人格、记忆、技能、项目上下文和工具指南组装成稳定系统提示；
- **Session Persistence**：会话写入 SQLite，并用 FTS5 支持跨会话搜索；
- **Learning Loop**：把经验沉淀到 memory 或 skill，而不是只留在一次对话里。

### 可中断和可观测

长期运行 Agent 必须可中断。用户可能在 CLI 里按 `Ctrl+C`，也可能在消息平台发新消息打断当前任务。Hermes 的设计强调：

- 工具调用过程对用户可见；
- 模型输出可以流式返回；
- 当前任务可以被用户中断或重定向；
- 背景进程可以被查询、等待、查看日志或终止。

这和传统后端的“请求进来、响应出去”不同。Agent 的执行可能持续几十秒甚至几分钟，用户需要知道它正在做什么、卡在哪里、是否可以停止。

---

## 16.5 Prompt System：稳定上下文而不是动态拼贴

Hermes 的 Prompt System 不只是把用户输入发给模型，而是一个上下文控制面。

一次系统提示通常由这些部分组成：

```text
System Prompt
  ├─ Personality / Persona
  ├─ Persistent Memory
  ├─ User Profile
  ├─ Relevant Skills
  ├─ Project Context Files
  │   ├─ AGENTS.md
  │   ├─ CLAUDE.md
  │   ├─ .cursorrules
  │   └─ 其他项目本地规则文件
  ├─ Tool Guidance
  ├─ Model-specific Instructions
  └─ Conversation State
```

这里有一个很关键的工程取舍：**长期记忆和用户画像在会话开始时注入为 frozen snapshot**。

也就是说，Agent 在本轮会话中可以更新 memory store 或 user profile，但这些变更不会立刻改变当前系统提示，而是下一次会话开始时生效。

这个设计看起来“不实时”，但它有两个好处：

- 保持系统提示前缀稳定，利于 prompt caching；
- 避免会话中途系统身份和长期记忆突然变化，降低行为漂移。

很多 Agent 原型会犯一个错误：每次工具调用后都重新拼一个完全不同的系统提示。短任务问题不大，长任务就容易出现行为不一致。Hermes 的 frozen snapshot 是一个值得学习的稳定性设计。

如果把 Hermes 的长期上下文进一步拆开，可以得到一张“上下文分层图”：

```mermaid
flowchart TB
    subgraph Stable["第一层：稳定前缀（会话开始时注入）"]
        Persona["人格 / Persona\n回答风格、行为边界"]
        UserProfile["用户画像 / User Profile\n偏好、角色、时区、习惯"]
        PMemory["长期记忆 / Persistent Memory\n环境事实、项目约定、长期约束"]
    end

    subgraph Selective["第二层：按需召回（减少上下文浪费）"]
        Skills["相关 Skills\n可复用流程、验证步骤、适用边界"]
        ContextFiles["项目上下文文件\nAGENTS.md / CLAUDE.md / .cursorrules"]
        SessionSearch["Session Search\n历史会话细节按需检索"]
    end

    subgraph Live["第三层：运行态上下文（随本轮任务演进）"]
        Conv["当前会话状态\n用户问题、任务目标、最新约束"]
        ToolObs["工具观察结果\nstdout / stderr / API result"]
    end

    Persona --> Conv
    UserProfile --> Conv
    PMemory --> Conv
    Skills --> Conv
    ContextFiles --> Conv
    SessionSearch --> Conv
    ToolObs --> Conv
```

这张图有助于读者理解：Hermes 不是把所有资料一股脑塞进 prompt，而是把“稳定前缀”“按需召回”和“本轮运行态”分开治理。

---

## 16.6 Memory：长期事实的稀缺预算

Hermes 的内置 Memory 不是无限知识库，而是一个很小、很克制的长期事实层。

更准确地说，它通常可以分成两类长期上下文：

| 层 | 作用 | 典型内容 |
|:---|:---|:---|
| Persistent Memory | Agent 的长期事实层 | 环境事实、项目约定、稳定工具经验、跨会话仍然有效的约束 |
| User Profile | 用户画像层 | 沟通偏好、角色、时区、工作习惯、表达偏好 |

这两层内容会以稳定快照的方式注入系统提示，因此必须非常短、非常高密度。Hermes 文档中给出了字符预算，这说明它把 Memory 当成“提示词预算里的稀缺资源”，不是普通文档库。底层实现既可以是内置存储，也可以接外部 memory provider。

### Memory 解决什么问题

Memory 适合保存这些内容：

- 用户偏好，例如“回答要简洁，先给结论”；
- 环境事实，例如“项目在某目录，测试命令是某个 make target”；
- 反复出现的工具坑，例如“这个服务器 SSH 端口不是 22”；
- 项目约定，例如“后端使用 Go + sqlc，迁移脚本在 migrations/”；
- 长期约束，例如“生产环境变更必须先跑 smoke test 并等待人工确认”。

Memory 不适合保存：

- 大段日志；
- 大段代码；
- 原始文档；
- 一次性临时路径；
- 已完成任务日志、PR/Issue 编号、临时工单状态；
- 一次性任务结果摘要；
- 可以随时重新检索到的通用知识。

这和第 8 章讲的 Agent Memory 原则一致：**长期记忆应该保存高价值、低频变化、可执行的信息，而不是把上下文垃圾永久化。**

### Session Search：长期历史的第二层

如果 Persistent Memory 和 User Profile 是“每次都必须知道的关键事实”，Session Search 则是“需要时再搜索的历史记录”。

Hermes 把 CLI 和消息平台的会话存进 SQLite，并用 FTS5 做全文检索。它可以回答类似问题：

```text
我们上周是不是讨论过这个部署失败？
之前那个 Postgres 连接池参数最后怎么改的？
我在哪次会话里让你记住了这个项目约定？
```

这形成了两级记忆：

| 层级 | 注入方式 | 适合内容 |
|:---|:---|:---|
| Memory Snapshot | 每次会话自动注入 | 必须稳定存在的事实 |
| Session Search | 工具按需检索 | 过去会话里的细节 |

这比“把全部历史塞进上下文”更可控，也比“完全不记得历史”更有连续性。

---

## 16.7 Skills：把经验变成可复用程序性记忆

Hermes 最有代表性的设计是 Skills System。

Memory 保存“事实”，Skills 保存“做法”。一个 Skill 通常是一个 Markdown 文档，描述某类任务的步骤、约束、工具选择、常见失败和验证方法。

可以这样理解：

```text
Memory = 我知道什么
Skill  = 我下次怎么做
Tool   = 我实际能执行什么
```

### 从一次任务到技能

一个典型闭环是：

```mermaid
flowchart LR
    Task["复杂任务"] --> Execute["Agent 执行"]
    Execute --> Trace["工具结果与会话轨迹"]
    Trace --> Reflect["反思哪些步骤可复用"]
    Reflect --> Skill["生成或更新 SKILL.md"]
    Skill --> Future["未来相似任务按需加载"]
    Future --> Execute
```

这比普通“记忆”更强，因为它保存的是可执行流程：

- 什么时候先搜索；
- 什么时候读配置；
- 哪个命令能验证；
- 常见错误怎么修复；
- 产物应该放在哪里；
- 什么操作必须先询问用户。

### Progressive Disclosure

Skills 也会占上下文预算，所以不能每次全部塞进 prompt。更合理的方式是 progressive disclosure：

1. 先只让模型看到 skill 名称和简短描述；
2. 当任务匹配时，再加载对应 `SKILL.md`；
3. 如果 skill 引用脚本、模板或资源，再按需读取。

这和本书前面讲的 Context Engineering 是同一个思想：不是让模型“知道所有东西”，而是让它在需要时拿到正确材料。

### Hermes 的 Skill 演化管道

如果说 OpenClaw 更强调 Skill 的加载优先级和插件生态，那么 Hermes 更值得关注的是：**Skill 如何从长期使用轨迹中演化出来**。

结合第 6 章对 Skills 的定义，Hermes 的成熟实现可以抽象成一条管道：

```text
Session Trace
  │
  ├─ 用户反复要求同类任务
  ├─ 某次任务形成稳定成功路径
  ├─ 工具调用序列可复用
  ├─ 验证命令稳定
  └─ 人工纠正减少
      │
      ▼
Skill Candidate
      │  提取触发条件、步骤、工具、约束、验证方式
      ▼
Review / Eval
      │  检查是否安全、是否过度泛化、是否真的提升质量
      ▼
Skill Registry
      │  保存版本、owner、适用范围和依赖工具
      ▼
Future Sessions
```

这条链路让 Skill 不只是“手写说明”，而是长期 Agent 的能力沉淀机制。一次成功任务本身没有价值，能被压缩成可验证、可复用、可审查的程序性知识，才有价值。

一个 Hermes-style Skill 需要保存的不只是步骤，还应保存这些元信息：

```yaml
skill:
  name: repo_release_check
  source_trace_ids:
    - trace_20260430_001
    - trace_20260430_019
  trigger:
    - "发布前检查"
    - "release validation"
  required_tools:
    - file_search
    - shell
    - git_diff
  verification:
    - "run tests"
    - "check diff"
    - "summarize risk"
  status: reviewed
  version: "0.3.0"
```

`source_trace_ids` 很重要。它让后续 review 能回到原始任务，判断这个 Skill 是从真实成功经验中总结出来的，还是模型凭空概括出来的。

### 风险：技能会固化错误经验

Skills 的风险也很明显：如果一次任务的解法本身是错误的，Agent 把它沉淀成 skill，下次会更稳定地犯同样错误。

因此生产级 Skills System 需要：

- skill 创建前有验证证据；
- skill 更新时保留版本或变更记录；
- skill 里写清适用条件和不适用条件；
- 定期清理过期技能；
- 对高风险技能增加人工 review。

真正可靠的自我进化，不是“做完就记住”，而是“验证后再沉淀”。

进一步说，Skill 还需要生命周期治理：不是越多越好，而是要有版本、owner、适用边界和清理机制。过期 Skill 应该归档，高风险 Skill 应该人工 review，常用 Skill 需要持续修订。只有这样，程序性记忆才会随着使用变得更可靠，而不是越来越臃肿。

---

## 16.8 Tool Runtime：工具注册、工具集与能力边界

Hermes 的工具系统可以拆成三层：

```text
Tool Registry
  ├─ Toolsets
  │   ├─ web
  │   ├─ terminal
  │   ├─ file
  │   ├─ browser
  │   ├─ memory
  │   ├─ session_search
  │   ├─ cronjob
  │   ├─ delegation
  │   └─ mcp-*
  └─ Execution Backends
      ├─ local
      ├─ docker
      ├─ ssh
      ├─ daytona
      ├─ modal
      └─ singularity
```

### Tool Registry

工具注册表负责：

- 收集工具 schema；
- 判断工具是否可用；
- 分发工具调用；
- 包装错误；
- 支持插件或 MCP 扩展；
- 根据平台配置启用不同工具集。

这和第 6 章 Tool Calling 的原则一致：模型不能直接执行任意函数，必须通过 Runtime 暴露的工具边界行动。

### Toolsets

Toolsets 是对工具能力的打包。比如：

- CLI 可以启用 terminal、file、web、browser；
- Telegram 可以启用 web、memory、send_message，但限制危险 terminal；
- 某个 profile 可以只启用 read-only 工具；
- MCP server 可以动态形成 `mcp-<server>` 工具集。

Toolsets 的价值是把“能不能用某类能力”变成配置，而不是散落在 prompt 里。

### Execution Backends

Hermes 的 terminal tool 不只是在本地跑命令，它可以选择多个执行后端：

| 后端 | 适合场景 |
|:---|:---|
| local | 本机开发、可信任务 |
| docker | 隔离执行、可复现环境 |
| ssh | 远程服务器或隔离机器 |
| daytona | 持久云端开发环境 |
| modal | serverless 执行和弹性任务 |
| singularity | HPC 或 rootless 容器场景 |

这说明 Hermes 把“执行环境”作为一等公民。对于长期 Agent 来说，这非常重要：同一个工具调用，在本地执行和在 Docker/SSH 执行，风险完全不同。

---

## 16.9 Action Engine：执行调度、失败处理与结果回传

工具中心定义“有哪些能力”，执行引擎负责“如何把能力可靠地用起来”。在 Hermes 这类长期 Agent 里，Action Engine 至少要处理五个阶段：

```text
model tool call
  -> action parse
  -> permission / policy check
  -> backend selection
  -> execute
  -> result handle
  -> retry / fallback / persist
```

这层的核心职责不是把函数调用出去，而是把模型的不确定输出转成可治理的系统行为。

| 阶段 | 关键问题 | 工程要求 |
|:---|:---|:---|
| 指令解析 | 模型输出是否符合工具 schema | 参数校验、类型转换、缺省值处理 |
| 权限检查 | 当前入口、profile、toolset 是否允许执行 | allowlist、危险命令审批、MCP 凭据过滤 |
| 后端选择 | 应该在本地、容器、SSH 还是云端执行 | 执行环境隔离、资源限制、超时控制 |
| 结果处理 | 工具结果如何回传给模型和用户 | stdout / stderr 分离、结果截断、结构化观察 |
| 失败恢复 | 工具失败后是否重试、降级或请求用户介入 | retry、fallback、human confirmation、session persistence |

长期 Agent 的执行引擎还必须支持可观测。用户需要看到 Agent 正在调用什么工具、在哪个环境执行、执行是否卡住、失败原因是什么、下一步是否需要确认。Hermes 的 callbacks、streaming、session persistence 和 research trajectory 都可以看成这一层的观测材料。

从系统设计角度看，Action Engine 是 Agent 从“会说”走向“会做”的分水岭。没有它，模型只是生成建议；有了它，模型的建议才会变成可审计、可回放、可恢复的行动。

---

## 16.10 Gateway、Cron 与 Profiles：让 Agent 长期在线且身份隔离

Gateway、Cron 和 Profiles 分别解决长期 Agent 的三个问题：用户从哪里触达 Agent，Agent 如何主动运行，以及不同身份和环境如何隔离。

### 16.10.1 Gateway：让 Agent 活在用户所在的平台里

Hermes 的 Gateway 是一个长期运行进程，负责把不同消息平台接入同一个 Agent Core。

```mermaid
flowchart TB
    subgraph Platforms["Messaging Platforms"]
        Telegram[Telegram]
        Discord[Discord]
        Slack[Slack]
        WhatsApp[WhatsApp]
        Signal[Signal]
        Email[Email]
        Other[Other Adapters]
    end

    subgraph Gateway["Gateway Process"]
        Adapter["Platform Adapters"]
        Auth["Allowlist / DM Pairing"]
        Router["Session Routing"]
        Slash["Slash Commands"]
        Hooks["Hooks"]
        CronTick["Cron Tick"]
    end

    subgraph AgentCore["Agent Core"]
        Session["Session State"]
        Runtime["AIAgent Runtime"]
        Tools["Toolsets"]
    end

    Platforms --> Adapter
    Adapter --> Auth
    Auth --> Router
    Router --> Slash
    Slash --> Session
    Router --> Runtime
    CronTick --> Runtime
    Hooks --> Runtime
    Runtime --> Tools
```

Gateway 不是简单 webhook 转发器。它至少承担六类职责：

- 平台适配：把不同平台消息标准化；
- 用户授权：限制谁可以访问 Agent；
- 会话路由：把不同平台、不同用户、不同线程映射到正确 session；
- Slash Commands：支持 `/new`、`/model`、`/skills`、`/stop` 等控制命令；
- Cron：定时触发 Agent 任务并把结果发送到平台；
- Hooks：在平台事件和 Agent Runtime 之间插入扩展逻辑。

这也是为什么 Hermes 不是“聊天机器人包装器”。它把 messaging gateway 做成了 Agent 的长期控制面。

---

### 16.10.2 Cron：Agent Task，而不是 Shell Cron

普通 cron 是执行命令：

```text
0 9 * * * /scripts/report.sh
```

Hermes 的 cron 更接近“定时 Agent 任务”：

```text
每天 9 点：
  读取相关数据
  使用指定 skill
  调用必要工具
  生成报告
  发到 Telegram 或 Slack
```

这类任务和 Shell Cron 的区别是：

| 维度 | Shell Cron | Hermes Cron |
|:---|:---|:---|
| 执行对象 | 命令或脚本 | Agent 任务 |
| 上下文 | 环境变量和文件 | 记忆、技能、工具、模型 |
| 输出 | stdout、文件、邮件 | 多平台消息、报告、行动结果 |
| 失败处理 | 依赖脚本自己处理 | 可通过 Agent 解释和总结 |
| 复用 | 主要复用脚本 | 复用 skill、memory、toolsets |

这对个人自动化很有价值。比如：

- 每天早上总结 GitHub issue；
- 每周检查服务器磁盘和证书；
- 每晚整理当天笔记；
- 定期生成项目风险报告；
- 监控某个网页或数据源变化。

从系统角度看，Cron 把 Agent 从“被动回答”推向“主动运行”。

---

### 16.10.3 Profiles：隔离长期身份

Hermes 支持 profile，每个 profile 有自己的 home、配置、memory、sessions、gateway PID 等。

这个设计看起来像多账户，但工程意义更深：

- 工作项目和个人生活不混在一起；
- 不同客户或团队有不同 memory；
- 不同 profile 可以启用不同工具和模型；
- 高风险 profile 可以只启用受限执行后端；
- 多个 profile 可以并行运行。

长期 Agent 最怕“上下文串线”。如果一个 Agent 既记住公司 A 的规则，又记住公司 B 的凭据，还在同一个 session 里切换任务，迟早会出问题。

Profile Isolation 是防止这种问题的第一道边界。

除了 Gateway、Cron 和 Profiles，Hermes 还在逐步形成一组后台系统来支撑长期运行：例如 delegation 用于把复杂任务拆给子 Agent，curator 用于维护 Skill 生命周期，kanban 用于多 Agent / 多 profile 协作队列。它们不一定每次都出现在用户视角里，但从架构上看，这些机制说明 Hermes 正在从“单 Agent 会话循环”扩展为“可编排的 Agent 工作平台”。

如果把这些后台系统之间的关系再抽象一层，可以得到下面这张关系图：

```mermaid
flowchart TB
    subgraph Entry["用户触达与任务入口"]
        Gateway["Gateway\n多平台消息入口"]
        Cron["Cron\n定时 Agent 任务"]
    end

    subgraph Runtime["统一运行时"]
        Core["Agent Core\nPrompt + Model + Tools + Session"]
        Profiles["Profiles\n身份、配置、记忆、权限隔离"]
    end

    subgraph Orchestration["后台编排与协作系统"]
        Delegation["Delegation\n把复杂任务拆给子 Agent"]
        Kanban["Kanban\n多 Agent / 多 profile 协作队列"]
        Curator["Curator\nSkill 生命周期维护"]
    end

    Gateway --> Core
    Cron --> Core
    Profiles --> Core
    Core --> Delegation
    Core --> Curator
    Core --> Kanban
    Delegation --> Profiles
    Kanban --> Profiles
    Curator --> Core
```

这张图强调的不是源码调用顺序，而是长期运行 Hermes 时几套后台能力之间的分工：Gateway 和 Cron 负责把任务送进 Runtime，Profiles 提供身份边界，Delegation / Kanban / Curator 则把单次会话扩展成可持续编排、可协作、可维护的工作平台。

---

## 16.11 Plugin 与 MCP：扩展能力，但保持边界

Hermes 的扩展可以来自三类路径：

```text
Built-in Tools
  系统自带工具，例如 web、terminal、memory、browser

Plugins
  用户、项目或 pip entry point 提供的扩展

MCP Servers
  外部工具能力通过 MCP 协议接入
```

### Plugin System

Plugin 可以注册：

- tools；
- hooks；
- CLI commands；
- memory provider；
- context engine。

其中 memory provider 和 context engine 是特殊插件类型，通常是单选：同一时间只激活一个外部 memory provider 或一个 context engine。这个约束很重要，因为多个记忆系统同时改写上下文，很容易产生冲突。

### MCP Integration

MCP 的价值是把外部工具系统标准化接入 Hermes，例如：

- GitHub；
- 数据库；
- 内部服务；
- 日志平台；
- 浏览器工具；
- 文档系统。

但 MCP 也扩大了权限面。一个成熟系统必须做：

- MCP server 级别授权；
- 工具 allowlist / denylist；
- secret 环境变量过滤；
- tool schema 审查；
- tool result 截断；
- 高风险工具人工确认。

这和第 13 章 Coding Agent 里的 MCP + 日志工作流是同一个问题：MCP 不是魔法，它只是把“外部能力”接进 Agent Runtime。真正可靠的是 Runtime 的权限、审计和上下文边界。

---

## 16.12 Security：长期 Agent 的攻击面更大

Hermes 官方文档把安全模型拆成多层，包括用户授权、危险命令审批、容器隔离、MCP 凭据过滤、上下文文件扫描、跨会话隔离和输入清洗。

这非常合理，因为长期运行 Agent 的风险比普通 Chatbot 大得多：

- 它连接消息平台；
- 它可能能执行命令；
- 它有长期记忆；
- 它能读项目文件；
- 它能调用浏览器；
- 它能运行 cron；
- 它能通过 MCP 访问外部系统；
- 它可能在服务器上 24 小时运行。

### 防御分层

| 层 | 典型风险 | 防御方式 |
|:---|:---|:---|
| 用户授权 | 陌生人给 bot 发消息触发工具 | allowlist、DM pairing |
| 命令执行 | 删除文件、泄露 secret、破坏系统 | 危险命令审批、工具策略 |
| 执行环境 | 本地权限过大 | Docker、SSH、Modal、Singularity |
| MCP | 外部 server 读取不该读取的凭据 | 环境变量过滤、工具白名单 |
| Context Files | 项目文件里藏 prompt injection | 扫描和隔离 |
| Session | 平台或 profile 间串线 | session/profile isolation |
| Cron | 后台任务误触发高风险动作 | 审批、审计、可暂停 |

安全设计的核心原则是：

> 不能因为 Agent 长期运行，就默认它值得长期信任。

它必须在每个入口、每个工具、每个执行后端、每次持久化时重新经过边界检查。

---

## 16.13 Research Pipeline：从使用轨迹到训练数据

Hermes 的另一个重要价值，是把 Agent 运行过程视为可积累的数据资产。它不仅是产品形态的 Agent，也明显在朝面向模型训练和研究的平台方向演进。

从当前公开能力看，它已经支持或正在围绕以下能力建设：

- batch trajectory generation；
- tool-calling 轨迹压缩；
- ShareGPT 格式导出；
- RL environments；
- Atropos 相关训练集成。

这说明 Hermes 把 Agent 运行看成一种可积累的数据资产：

```text
真实任务
  -> Agent 工具调用轨迹
  -> 成功/失败/人工修正
  -> 压缩与标注
  -> eval dataset
  -> fine-tuning / RL
  -> 更好的 agent behavior
```

这对团队自研 Agent 很有启发。很多团队只关心“Agent 当前能不能完成任务”，但忽略了“Agent 的失败能不能变成下一轮改进的数据”。真正长期可进化的系统，需要把每次执行都变成可复盘、可评估、可训练的材料。

换句话说，Hermes 的价值不在于它已经是一个完备的训练数据工厂，而在于它把“运行轨迹资产化”明确纳入了 Runtime 设计：任务过程不仅服务于当下完成率，也服务于后续 eval、微调和策略改进。

---

## 16.14 关键设计原则：模块化、标准化、扩展性、可靠性、安全性、可观测

Hermes 的价值不只在于功能多，而在于它把长期 Agent 的复杂性拆成了可以治理的工程边界。结合前面的架构，可以提炼出六个设计原则。

| 原则 | Hermes 中的体现 | 对自研 Agent 的启发 |
|:---|:---|:---|
| 模块化 | Entry Points、Agent Core、Context & Learning、Tool Runtime、Execution Backends、State Storage 分离 | 不要把入口、模型调用、工具执行和记忆写在一个大循环里 |
| 标准化 | Tool Schema、Toolsets、MCP、Skills、Context Files 都有相对稳定的接口 | Agent 能力必须通过结构化协议接入，不能只靠 prompt 描述 |
| 可扩展 | Plugins、MCP、Execution Backends、Gateway Adapters、Memory Provider 可替换 | 新平台、新工具、新模型应该是注册和配置问题，而不是重写核心循环 |
| 可靠性 | Session Persistence、Context Compression、Retry / Fallback、Streaming Callbacks | 长任务必须可中断、可恢复、可解释失败原因 |
| 安全性 | Profiles、Toolsets、Allowlist、危险命令审批、容器隔离、MCP 凭据过滤 | 长期 Agent 的默认姿态应该是最小权限，而不是默认信任 |
| 可观测 | callbacks、streaming、sessions、tool results、trajectory datasets | 每次行动都要能追踪：为什么调用、调用了什么、结果是什么、是否验证 |

这六个原则可以作为评估 Agent Runtime 的检查清单。一个系统即使接入了很多工具，如果没有 profile 隔离、工具权限、执行审计和失败恢复，也只能算能力演示；只有当这些能力被模块化、标准化、可观测地纳入 Runtime，才适合长期运行。

更重要的是，Hermes 把“可进化”建立在可治理之上。Memory、Skills 和 Research Pipeline 让 Agent 可以增长能力；Toolsets、Profiles、Security 和 Eval 约束让这种增长不会失控。这是长期 Agent 和普通 Chatbot 最大的工程差异。

---

## 16.15 Hermes 与 OpenClaw 的架构对比

Hermes 和 OpenClaw 很容易被放在一起比较，因为它们都强调个人 AI 助手、多渠道入口、本地运行和工具生态。但它们的重心不同。

| 维度 | OpenClaw | Hermes Agent |
|:---|:---|:---|
| 核心定位 | Personal Agent Gateway | Self-improving Long-running Agent |
| 入口 | 多聊天平台、WebChat、CLI、节点 | CLI/TUI、Messaging Gateway、ACP、Cron、脚本化/服务化入口 |
| 记忆 | 个人上下文和长期配置 | Persistent Memory、User Profile、SQLite Session Search、外部 Memory Provider |
| 技能 | 技能与插件生态 | 支持创建/更新 Skills，并将验证过的经验沉淀为可复用流程，兼容 agentskills.io |
| 工具 | 工具、技能、插件、MCP | Tool Registry、Toolsets、MCP、Plugins、执行后端 |
| 执行环境 | 本地和沙箱为主 | local、Docker、SSH、Daytona、Modal、Singularity |
| 研究闭环 | 更偏产品使用 | 轨迹生成、压缩、RL/eval 数据 |
| 架构气质 | Gateway-first | Learning-loop-first |

两者不是谁替代谁，而是代表两种长期 Agent 的设计方向：

- 如果重点是“如何让用户从各种渠道触达 Agent”，OpenClaw 的 Gateway 思路更突出；
- 如果重点是“如何让 Agent 在长期使用中积累能力”，Hermes 的 Memory + Skills + Session + Trajectory 思路更突出。

---

## 16.16 如果自研 Hermes 类系统，最小可行架构是什么

不要一开始就实现完整 Hermes。一个可落地的最小版本可以这样设计：

```text
hermes-like-agent/
├── cli.py
├── agent/
│   ├── loop.py
│   ├── prompt_builder.py
│   ├── provider.py
│   └── callbacks.py
├── context/
│   ├── memory.py
│   ├── skills.py
│   ├── session_store.py
│   └── context_files.py
├── tools/
│   ├── registry.py
│   ├── terminal.py
│   ├── file.py
│   └── web.py
├── gateway/
│   ├── telegram.py
│   └── router.py
├── cron/
│   └── scheduler.py
└── state/
    ├── config.yaml
    ├── state.db
    ├── memories/
    └── skills/
```

优先级建议：

1. **Agent Loop**：先跑通模型、工具、回调、持久化；
2. **Memory**：只做两个小文件，限制字符数；
3. **Session Search**：SQLite + FTS5，先支持按关键词召回；
4. **Skills**：先手工创建技能，再考虑自动生成；
5. **Toolsets**：把工具按平台和 profile 配置；
6. **Gateway**：先接一个平台，例如 Telegram；
7. **Cron**：只允许低风险 read-only 任务；
8. **Security**：命令审批、路径沙箱、allowlist 必须尽早做。

不要过早做的事情：

- 不要一开始接十几个消息平台；
- 不要一开始做复杂 external memory provider；
- 不要自动生成并自动执行高风险 skill；
- 不要给 messaging gateway 默认开放 terminal；
- 不要把所有历史会话无脑注入 prompt。

---

## 16.17 设计启示

Hermes Agent 给 Agent 工程带来几个重要启示。

### 1. 长期 Agent 的核心是连续性

连续性不是“把聊天记录都放进去”，而是分层组织：

- Memory 保存关键事实；
- Session Search 保存历史细节；
- Skills 保存操作流程；
- Context Files 保存项目规范；
- Profiles 保存身份边界。

### 2. 自我进化必须受验证约束

Agent 会记忆、会写 skill、会复用经验，这听起来很强，但如果没有验证和审查，就会把错误经验固化。

更健康的闭环是：

```text
执行 -> 验证 -> 复盘 -> 沉淀 memory/skill -> eval 检查 -> 再复用
```

### 3. Gateway 让 Agent 从工具变成服务

CLI Agent 是工具；Gateway Agent 是服务。

一旦 Agent 进入 Telegram、Slack、Discord、Email，它就不再只服务“坐在电脑前的人”，而变成一个长期在线的数字工作者。此时必须重新设计授权、会话、审计、中断和后台任务。

### 4. Toolsets 是权限治理的基本单位

不要只用 prompt 控制工具。把工具分成 toolsets，再按平台、profile、任务类型启用，是更可靠的设计。

例如：

```text
CLI profile: web + file + terminal + memory
Telegram profile: web + memory + send_message
Cron profile: web + read-only file + send_message
Research profile: batch + trajectory + eval tools
```

### 5. 轨迹是 Agent 的资产

每一次工具调用、失败、修复、用户纠正，都可以成为 eval、fine-tuning 或 RL 的材料。

如果一个团队认真做 Agent，就应该尽早记录：

- task；
- system prompt version；
- tool calls；
- tool results；
- user corrections；
- final output；
- verification result；
- failure reason。

这就是 Agent 系统的“数据飞轮”。

---

## 本章小结

Hermes Agent 展示了长期运行 Agent 的另一条成熟路径：不是只做更强的单次推理，而是围绕模型建立记忆、技能、工具、入口、执行环境和数据闭环。

本章核心结论：

- Hermes 的本质是一个自我进化的长期 Agent Runtime；
- 它的底层架构可以拆成大脑中枢、记忆系统、小脑、工具中心、执行引擎和外部环境六个核心组件；
- Memory 保存关键事实，Session Search 保存历史细节，Skills 保存可复用做法；
- 核心数据流不是单向问答，而是用户输入、规划、工具执行、外部结果和记忆回流组成的闭环；
- Gateway 让同一个 Agent 活在 CLI、消息平台和自动化任务中；
- Tool Registry、Toolsets、Action Engine、Execution Backends 把行动能力拆成可治理的边界；
- Profiles 是长期 Agent 防止上下文串线的重要机制；
- Security 必须覆盖用户授权、命令审批、容器隔离、MCP 凭据过滤、上下文扫描和 session 隔离；
- Research Pipeline 把 Agent 执行轨迹变成 eval、fine-tuning 和 RL 的数据资产；
- 模块化、标准化、可扩展、可靠性、安全性和可观测，是长期 Agent 从 demo 走向 Runtime 的关键设计原则。

如果 OpenClaw 让我们看到“个人 Agent Gateway 如何把用户和模型连起来”，Hermes 则让我们看到“长期 Agent 如何在使用中积累能力”。对自研 Agent 来说，最值得学习的不是某个具体命令，而是它把长期性拆成了可工程化的系统组件。

---

## 参考资料

1. [Hermes Agent GitHub Repository - NousResearch/hermes-agent](https://github.com/NousResearch/hermes-agent)
2. [Hermes Agent Documentation](https://hermes-agent.nousresearch.com/docs/)：官方文档入口，包含 Messaging Gateway、Tools & Toolsets、Skills、Architecture，以及 closed learning loop 的总体说明。
3. [Hermes Agent Features Overview](https://hermes-agent.nousresearch.com/docs/user-guide/features/overview/)：功能总览，覆盖工具集、Skills、Persistent Memory、Context Files 等核心能力。
4. [Hermes Agent Architecture](https://hermes-agent.nousresearch.com/docs/developer-guide/architecture)：开发者架构说明，覆盖 Prompt Builder、Tool Registry、Session Persistence、Gateway、Plugin、Cron、ACP、RL / Trajectory 等内部模块。
5. [Hermes Agent Tools & Toolsets](https://hermes-agent.nousresearch.com/docs/user-guide/features/tools/)：工具与工具集说明，列出 web、terminal、file、browser、memory、session_search、cronjob、delegation、MCP 等常见能力类别。
6. [Hermes Agent Toolsets Reference](https://hermes-agent.nousresearch.com/docs/reference/toolsets-reference)：工具集参考，说明 toolset 如何作为按平台、会话和任务控制能力边界的机制。
7. [Hermes Agent Built-in Tools Reference](https://hermes-agent.nousresearch.com/docs/reference/tools-reference/)：内置工具参考，适合追踪当前代码派生出的工具注册表和 MCP 动态工具能力。
8. [Hermes Agent Persistent Memory](https://hermes-agent.nousresearch.com/docs/user-guide/features/memory/)
9. [Hermes Agent Security](https://hermes-agent.nousresearch.com/docs/user-guide/security)
