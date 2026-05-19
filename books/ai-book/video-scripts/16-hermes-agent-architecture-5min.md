# 5 分钟短视频脚本：Hermes Agent 架构解析

## 视频定位

标题建议：**Hermes Agent 架构解析：长期运行的 Agent 如何记忆、进化和跨入口工作？**

视频类型：知识讲解型 + 技术解读型  
目标观众：工程师、技术管理者、AI Agent 学习者  
核心观点：Hermes 的价值不是多一个聊天入口，而是把 Agent 做成长期运行、可记忆、可沉淀技能、可跨平台执行、可产生训练轨迹的个人运行时。

## 成片参数

```text
时长：约 5 分钟
比例：9:16 竖屏，素材结构保留 16:9 迁移可能
风格：技术信息图，架构图优先，文字克制
语速：每分钟约 240-280 中文字
配音：Piper 中文 TTS 首版
字幕：全程中文字幕，关键词高亮
```

## 一句话钩子

```text
真正的 Agent，不是多一个聊天入口，而是一个会长期积累记忆和技能的运行时。
```

## 完整口播稿

```text
真正的 Agent，不是多一个聊天入口，而是一个会长期积累记忆和技能的运行时。

很多 AI 助手看起来都很像：你发一句话，它回一句话；换到 Telegram、Slack 或 CLI，也只是多了几个入口。但 Hermes Agent 想解决的问题更深：Agent 如何在长期使用中变得更懂你、更会做事？

可以把 OpenClaw 和 Hermes 放在一起理解。OpenClaw 更像个人 Agent Gateway，重点是让 Agent 到达用户所在的地方。Hermes 更像长期运行的 Agent Runtime，重点是记忆、技能、工具、执行后端和轨迹闭环。

一句话概括：OpenClaw 解决“Agent 如何到达用户”，Hermes 解决“Agent 如何长期成长”。

普通聊天机器人围绕单次回复设计。用户输入，模型回答，这条链路很短。但真正的个人 Agent 需要连续性：它要记得用户偏好和项目背景，能把复杂任务沉淀成可复用技能，也能从 CLI、消息平台、IDE 或定时任务继续工作。

所以 Hermes 可以抽象成一个长期运行的 Agent Runtime：它包含 Persistent Memory、Skills、Multi-platform Gateway、Tool Registry、Execution Backends 和 Research Trajectory Pipeline。

这背后有一个关键判断：Agent 的能力不只来自模型参数，而来自模型、记忆、技能、工具、入口、执行环境和历史轨迹共同组成的系统。

理解 Hermes，可以先看六个核心组件。

第一是大脑中枢，负责理解意图、生成推理和选择工具。第二是记忆系统，保存长期事实、用户画像、历史会话和 Skills。第三是小脑，负责规划、状态和工作流。第四是工具中心，管理 Tool Registry、Toolsets、MCP 和插件。第五是执行引擎，把工具调用变成真实动作，并处理失败和重试。第六是外部环境，包括 Gateway、Cron、ACP、CLI、本地、Docker、SSH 等后端。

一次 Hermes 任务的数据流也不是 Prompt 到 Answer 这么简单。用户从 Telegram、CLI 或 Cron 发起请求，入口层先标准化消息，Session Router 找到正确的用户、线程和 profile。Prompt Builder 再组装人格、长期记忆、相关 Skills、项目上下文和工具边界。

模型判断是否需要工具。Tool Runtime 在对应 toolset 和执行后端里调用 web、terminal、file、browser、memory 或 MCP 工具。工具结果回到 Agent Loop，模型继续推理，直到给出答案。最后，会话写入 Session Store，稳定事实进入 Memory，稳定流程进入 Skill 候选，执行轨迹进入研究数据。

这就是 Hermes 的重点：Memory 不只是输入，也参与输出后的回流。没有回流，Agent 每次都从头开始；没有约束，Agent 又会把错误、噪声和越权信息永久化。

所以 Prompt System 的设计很关键。Hermes 不是把所有历史塞进上下文，而是分三层治理：稳定前缀、按需召回、本轮运行态。这样既有连续性，又不会让上下文失控。

Memory 保存“我知道什么”，Skills 保存“我下次怎么做”，Tool 保存“我实际能执行什么”。这三者不要混在一起。

比如一次发布检查，如果反复成功，Hermes-style Skill 应该沉淀触发条件、步骤、依赖工具、验证命令和适用边界，而不是只记一句“发布前要检查”。

这也是 Hermes 最值得借鉴的地方：它把经验变成程序性记忆。一次成功任务本身价值有限，能被压缩成可验证、可复用、可审查的流程，才会真正提升长期能力。

但自我进化也有风险。错误步骤可能被固化，隐私可能被写入长期记忆，跨 profile 的信息可能串线，高风险工具可能被错误授权。所以长期 Agent 必须有验证、评测、人工审核、权限边界和可观测轨迹。

Gateway 和 Toolsets 正是边界的一部分。多入口让 Agent 活在用户所在的平台里，Toolsets 则决定它在每个平台、每个 profile、每种任务下能调用哪些能力。长期在线不等于无限授权。

最后总结一下：Hermes 的核心价值，不是多一个聊天入口，而是把 Agent 做成一个长期运行的个人运行时。它会记忆，会沉淀技能，会跨入口工作，会调用工具，也会把轨迹变成未来改进的数据资产。

如果你要自研个人或团队 Agent，最小可行架构不是“接一个模型 API”，而是至少包含五件事：统一入口，稳定上下文，受控工具集，可审计执行，以及被验证约束的学习闭环。

真正的 Agent 能力，来自模型加 Runtime，加长期轨迹闭环。
```

## 分镜表

| 时间 | 口播重点 | 画面设计 | 屏幕文字 |
|:---|:---|:---|:---|
| 0:00-0:15 | Hook | 大标题 + 多入口图标汇入 Runtime | Agent 不是聊天入口 |
| 0:15-0:37 | Hermes 定位 | OpenClaw 与 Hermes 对比 | Gateway 解决入口，Runtime 解决成长 |
| 0:37-0:57 | Runtime 总览 | Memory、Skills、Gateway、Tools、Backends 环绕 Runtime | Long-running Agent Runtime |
| 0:57-1:17 | 核心判断 | 模型与 Runtime 关系图 | Agent 能力不只来自模型 |
| 1:17-1:39 | 六大组件 | 六组件架构图 | 大脑、记忆、小脑、工具、执行、环境 |
| 1:39-1:59 | 大脑中枢 | Prompt Builder 组装上下文 | LLM + Prompt + Provider |
| 1:59-2:19 | 记忆系统 | Memory、Profile、Session Search、Skills 分层 | 让 Agent 连续存在 |
| 2:19-2:39 | 工具中心 | Tool Registry 到 Toolsets | 决定 Agent 能做什么 |
| 2:39-2:59 | 执行引擎 | Tool Call 到 Backend | 把调用变成真实动作 |
| 2:59-3:21 | 数据流 | 入口、上下文、工具、观察、回流 | 不只是 Prompt 到 Answer |
| 3:21-3:43 | Prompt System | 稳定前缀、按需召回、本轮运行态 | 稳定上下文，而不是动态拼贴 |
| 3:43-4:05 | Skills | 任务轨迹到 SKILL.md | 把经验变成程序性记忆 |
| 4:05-4:25 | Gateway 与 Toolsets | 多入口加权限边界 | 长期在线不等于无限授权 |
| 4:25-4:45 | 风险 | 错误经验、隐私、串线、危险授权 | 自我进化必须被验证约束 |
| 4:45-5:07 | 自研启示 | 最小可行架构清单 | 不要只接模型 API |
| 5:07-5:25 | 总结 | 模型、Runtime、轨迹闭环合并 | 真正能力来自系统闭环 |

## 字幕版文案

```text
真正的 Agent，不是多一个聊天入口，而是一个会长期积累记忆和技能的运行时。
很多 AI 助手看起来都很像：你发一句话，它回一句话；换到 Telegram、Slack 或 CLI，也只是多了几个入口。
但 Hermes Agent 想解决的问题更深一层：Agent 如何在长期使用中变得更懂你、更会做事？
OpenClaw 更像个人 Agent Gateway，重点是让 Agent 到达用户所在的地方。
Hermes 更像长期运行的 Agent Runtime，重点是记忆、技能、工具、执行后端和轨迹闭环。
OpenClaw 解决“Agent 如何到达用户”，Hermes 解决“Agent 如何长期成长”。
普通聊天机器人围绕单次回复设计，但真正的个人 Agent 需要连续性。
它要记得用户偏好，记得项目背景，能把复杂任务沉淀成可复用技能。
Hermes 可以抽象成一个长期运行的 Agent Runtime。
它包含 Persistent Memory、Skills、Multi-platform Gateway、Tool Registry、Execution Backends 和 Research Trajectory Pipeline。
Agent 的能力不只来自模型参数，而来自模型、记忆、技能、工具、入口、执行环境和历史轨迹共同组成的系统。
理解 Hermes，可以先看六个核心组件：大脑、记忆、小脑、工具中心、执行引擎和外部环境。
一次 Hermes 任务的数据流也不是 Prompt 到 Answer 这么简单。
入口层标准化消息，Prompt Builder 组装上下文，Tool Runtime 执行工具，Agent Loop 继续推理。
最后，会话写入 Session Store，稳定事实进入 Memory，稳定流程进入 Skill 候选，执行轨迹进入研究数据。
Memory 不只是输入，也参与输出后的回流。
没有回流，Agent 每次都从头开始；没有约束，Agent 会把错误和噪声永久化。
Prompt System 要分三层治理：稳定前缀、按需召回、本轮运行态。
Memory 保存“我知道什么”，Skills 保存“我下次怎么做”，Tool 保存“我实际能执行什么”。
这三者不要混在一起。
Hermes 最值得借鉴的地方，是把经验变成程序性记忆。
一次成功任务本身价值有限，能被压缩成可验证、可复用、可审查的流程，才会真正提升长期能力。
但自我进化必须被验证约束。
错误步骤、隐私泄露、跨 profile 串线和高风险授权，都会让长期 Agent 变危险。
Gateway 让 Agent 活在用户所在的平台里，Toolsets 决定它能调用哪些能力。
长期在线不等于无限授权。
Hermes 的核心价值，是把 Agent 做成一个长期运行的个人运行时。
如果你要自研个人或团队 Agent，最小可行架构不是接一个模型 API。
它至少需要统一入口、稳定上下文、受控工具集、可审计执行，以及被验证约束的学习闭环。
真正的 Agent 能力，来自模型加 Runtime，加长期轨迹闭环。
```

## 发布标题备选

1. Hermes Agent 架构解析：长期运行的 Agent 如何自我进化？
2. 真正的 AI Agent，不只是聊天入口
3. Memory、Skills、Gateway：Hermes Agent 的架构启示
4. 从聊天机器人到长期运行时：Hermes Agent 讲清楚了什么？
5. Agent 能力不只来自模型：Hermes 的 Runtime 架构

## 封面文案

```text
Hermes Agent
长期运行的 Agent Runtime
```

副标题：

```text
记忆、技能、多入口与轨迹闭环
```
