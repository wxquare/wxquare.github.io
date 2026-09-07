# 3 分钟短视频脚本：LLM 能力边界与架构约束

## 视频定位

标题建议：**LLM 很强，但边界在哪里？3 分钟讲清 AI Agent 的架构底层逻辑**

视频类型：知识科普型  
目标观众：对 AI、AI Agent、RAG、Prompt Engineering 感兴趣的工程师、产品经理和技术管理者  
核心观点：LLM 负责理解、生成和规划；系统负责检索、计算、执行、权限、验证和审计。

## 成片参数

```text
时长：约 3 分钟
比例：9:16 竖屏，兼容视频号、抖音、B 站竖屏、YouTube Shorts
风格：技术信息图动画，干净、克制、节奏清晰
语速：每分钟约 240-280 中文字
配音：沉稳、有工程感，避免营销腔
字幕：全程双行中文字幕，关键词加粗或高亮
背景音乐：低音量科技感节奏，不能压过人声
```

## 一句话钩子

```text
如果你让 LLM 直接算钱、查实时数据、执行线上命令，它迟早会坑你。
```

## 完整口播稿

```text
如果你让 LLM 直接算钱、查实时数据、执行线上命令，它迟早会坑你。

这不是因为大模型不聪明，而是因为我们经常把它用错了。

LLM 的本质，不是数据库，不是计算器，也不是权限系统。它更像一个在当前上下文下，生成最可能答案的概率模型。

所以它非常擅长什么？

它擅长理解语言，改写内容，总结文档，做分类，抽取结构化信息，生成代码片段，也擅长根据上下文做规划和解释。

比如你给它一段用户反馈，让它判断是物流问题、价格问题还是售后问题，这类任务通常很适合。因为标签有限，输入清楚，输出还能用 JSON Schema 约束。

但它不擅长什么？

它不擅长精确计算，不擅长实时事实，不擅长长期记忆，也不应该独立执行高风险操作。

比如这个问题：本金一万元，年利率百分之五，复利二十年后是多少钱？

你不应该让 LLM 直接给数字。更好的架构是：让 LLM 识别这是复利问题，选择公式，然后调用计算工具。模型负责理解，工具负责计算。

同样，如果你问今天的价格、库存、线上服务状态，它也不应该凭记忆回答。它应该去查数据库、搜索系统、监控平台或业务 API。

这就是 AI Agent 架构的核心。

成熟的 AI 系统，不是让 LLM 做所有事情，而是把它放在一个工程系统里。

这个系统通常包括六个部分：

第一，RAG，用来查外部知识。

第二，Tool，用来计算、搜索、读文件、调用 API。

第三，Memory，用来保存用户偏好、任务状态和长期上下文。

第四，Policy，用来控制权限，决定哪些动作可以自动执行，哪些必须审批。

第五，Eval，用来评估答案是否真的有效。

第六，Trace，用来记录每一步输入、输出、工具调用和错误，方便复盘。

所以，真正的问题不是“这个模型够不够强”，而是“这个任务应该交给模型，还是交给系统”。

让 LLM 做它擅长的事：理解、生成、归纳、规划。

让传统工程做它擅长的事：计算、检索、执行、存储、权限和验证。

这就是 LLM 能力边界，也是 Agent 系统设计的第一原则。

一句话总结：成熟的 AI 架构，不是迷信模型，而是知道什么时候不该让模型单独做决定。
```

## 分镜表

| 时间 | 口播重点 | 画面设计 | 屏幕文字 |
|:---|:---|:---|:---|
| 0:00-0:10 | LLM 直接算钱、查数据、执行命令会出问题 | 黑底代码终端快速闪过：`calculate money`、`query live price`、`restart service`，随后出现红色警示线 | 别把 LLM 当确定性系统 |
| 0:10-0:25 | 不是模型不聪明，而是用错了 | 中央出现一个发光的 LLM 核心，外面有“数据库”“计算器”“权限系统”三个模块被划掉 | 问题不是模型弱，而是边界错 |
| 0:25-0:42 | LLM 是概率生成模型 | 动画展示 token 一个个生成，概率条从左到右滚动 | LLM = 上下文中的概率生成器 |
| 0:42-1:05 | LLM 擅长语言理解、总结、分类、抽取、代码片段 | 六个卡片依次浮现：总结、改写、分类、抽取、代码、规划 | 它擅长语义任务 |
| 1:05-1:25 | 客户反馈分类例子 | 手机聊天气泡：“订单还没发货，等了 3 天”，右侧输出 JSON | 标签有限、输入清楚、输出可约束 |
| 1:25-1:45 | LLM 不擅长计算、实时事实、长期记忆、高风险执行 | 四个红色图标：计算器、时钟、记忆、危险操作 | 这些不该直接交给模型 |
| 1:45-2:05 | 复利例子 | 左侧 LLM 识别问题，右侧 calculator 工具计算公式 | 模型选公式，工具算结果 |
| 2:05-2:28 | 正确的 Agent 架构 | 中央 LLM，周围六个模块连接：RAG、Tool、Memory、Policy、Eval、Trace | Agent = LLM + 工程系统 |
| 2:28-2:45 | 模型负责理解，系统负责执行和验证 | 流程图：User -> LLM -> Tool -> Eval -> Answer | 生成不是证据，验证才是证据 |
| 2:45-3:00 | 总结第一原则 | 所有模块收束成一句话，背景渐亮 | 知道什么时候不让模型单独决定 |

## 字幕版文案

```text
如果你让 LLM 直接算钱、查实时数据、执行线上命令，它迟早会坑你。
这不是因为大模型不聪明，而是因为我们经常把它用错了。
LLM 不是数据库，不是计算器，也不是权限系统。
它是在当前上下文下生成最可能答案的概率模型。
它擅长理解语言、总结文档、做分类、抽取结构化信息、生成代码片段和规划任务。
但它不擅长精确计算、实时事实、长期记忆和高风险执行。
比如复利计算，不应该让 LLM 直接给数字。
正确做法是让 LLM 识别问题和公式，再调用计算工具。
模型负责理解，工具负责计算。
这就是 AI Agent 架构的核心。
成熟系统通常包括 RAG、Tool、Memory、Policy、Eval 和 Trace。
RAG 查外部知识，Tool 做确定性操作，Memory 保存状态，Policy 控制权限，Eval 验证质量，Trace 记录过程。
真正的问题不是模型够不够强，而是这个任务应该交给模型，还是交给系统。
让 LLM 做理解、生成、归纳和规划。
让传统工程做计算、检索、执行、存储、权限和验证。
成熟的 AI 架构，不是迷信模型，而是知道什么时候不该让模型单独做决定。
```

## 画面素材提示词

### 主视觉：Agent 架构图

```text
Clean technical infographic, vertical 9:16 layout, central glowing node labeled LLM, six surrounding modules labeled RAG, Tool, Memory, Policy, Eval, Trace, thin connection lines, dark background, blue and white accents, professional software architecture style, high contrast, minimal, no cartoon characters
```

### 概率生成模型画面

```text
Vertical technical animation style, tokens appearing one by one from left to right, probability bars under each token, dark interface, clean typography, subtle blue glow, modern AI engineering explainer visual
```

### 复利计算例子画面

```text
Split screen technical diagram, left side LLM reads compound interest question, middle formula FV = P * (1 + r) ^ n, right side calculator tool outputs result, clean UI, dark background, blue highlights, professional educational video style
```

### 错误边界画面

```text
Minimal technical warning visual, four icons representing calculator, realtime clock, memory storage, dangerous command execution, red caution outlines, dark background, clean vector style, no mascot, professional AI safety explainer
```

## 剪辑节奏建议

```text
0:00-0:10：强钩子，快速切换，字幕大
0:10-0:42：解释本质，动画慢一点，让观众理解
0:42-1:25：能力清单，卡片式快节奏
1:25-2:05：边界和复利例子，制造“原来如此”的转折
2:05-2:45：架构图登场，是全片核心画面
2:45-3:00：收束，给一句可传播的总结
```

## 制作工具建议

```text
脚本润色：ChatGPT / Claude / Codex
画面：Canva / Figma / Keynote / Excalidraw
配音：真人录音优先，也可以用 ElevenLabs / 剪映配音
剪辑：剪映 / CapCut / Premiere
字幕：剪映自动字幕后人工校对
架构图：Mermaid / Excalidraw / Figma
```

## 发布标题备选

1. LLM 很强，但边界在哪里？3 分钟讲清 AI Agent 的底层逻辑
2. 为什么不能让大模型直接算钱、查库存、执行命令？
3. AI Agent 架构第一原则：什么时候不该相信 LLM？
4. 大模型不是系统，真正的 AI Agent 要这样设计
5. LLM 能做什么，不能做什么？工程师必须知道的边界

## 封面文案

```text
LLM 很强
但别让它做所有事
```

副标题：

```text
3 分钟讲清 AI Agent 的能力边界
```

## 结尾引导

```text
如果你想系统学习 AI Agent 工程实践，下一期我们继续讲：Prompt 为什么不是提问技巧，而是任务协议。
```
