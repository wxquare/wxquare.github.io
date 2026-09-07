# 第12章 LLM API 协议：模型能力如何被系统消费

> 模型能力不是直接被业务系统“使用”，而是先被抽象成输入、消息、工具、结构化输出和流式事件，再进入 Runtime、编排器和工具系统。

## 引言

理解 LLM 的原理，只能回答“模型为什么能工作”；理解 LLM API 协议，才能回答“系统怎样把模型接进来”。在工程落地里，这一步经常被低估。很多团队熟悉 Token、上下文窗口、Transformer 和 Sampling，却在真正接入模型时发现一连串新问题：

1. 一次调用里，系统究竟该传什么？
2. `system`、`messages`、`input`、`content blocks` 这些字段本质上分别代表什么？
3. 模型返回的为什么不只是文本，还有工具调用、结构化结果、流式事件和 usage？
4. OpenAI、Anthropic、DeepSeek 看起来都“差不多”，实际差别在哪？
5. 一个 Agent 工作流，怎样被映射成多轮请求、工具回填和状态延续？

这一章讨论的不是 SDK 语法糖，而是模型协议层的统一抽象。你可以把它看成 Agent Runtime 和模型能力之间的接口层：往左连接 Prompt、Context、Tool、Skill 和 Workflow，往右连接 OpenAI、Anthropic、DeepSeek 等不同提供方。

如果说第 11 章定义了 Agent Runtime 的总纲，那么本章回答的就是：**Runtime 究竟怎样消费模型能力。**

## 12.1 统一抽象：输入、指令、上下文与输出

学到这里，如果只知道 token、上下文和 Transformer 还不够。工程上真正调用大模型时，还要理解这些能力如何暴露成 API 协议。否则你会知道“模型能做什么”，却不知道“系统该怎么把这些能力接进来”。

这也是为什么 API 协议属于 Agent Runtime 的核心知识。它不是单纯的 SDK 用法，而是模型能力在工程边界上的投影：

- 上下文窗口会表现为 `messages`、`input`、`system`、`conversation state` 等输入结构；
- 结构化输出会表现为 JSON mode、JSON schema、strict schema；
- 工具调用会表现为 `tools`、`tool_choice`、`tool_calls` 或 `tool_use`；
- 长上下文成本会表现为 prompt caching、cache hit 指标和上下文压缩；
- 多模态能力会表现为文本、图片、文件、音频等不同 content block；
- 推理能力会表现为 reasoning / thinking 开关、effort、流式事件和 token 统计。

### 12.1.1 协议抽象：LLM API 本质上在传什么

不管是 OpenAI、Anthropic 还是 DeepSeek，主流 LLM API 本质上都是：

```text
HTTP + JSON
  -> 提交上下文和控制参数
  -> 模型生成文本 / 结构化结果 / 工具调用
  -> 返回 usage、stop reason、可选流式事件
```

从抽象层看，一次调用通常包含六类信息：

| 抽象层 | 典型字段 | 作用 |
|:---|:---|:---|
| 模型选择 | `model` | 选择能力、价格、延迟和上下文窗口 |
| 输入上下文 | `messages`、`input`、`system` | 把用户问题、历史、规则和证据送进模型 |
| 生成控制 | `temperature`、`max_tokens`、`reasoning_effort` | 控制采样、长度和推理预算 |
| 输出约束 | `response_format`、`json_schema`、`strict` | 让结果更像机器可消费契约 |
| 外部能力 | `tools`、`tool_choice` | 让模型提出工具调用，而不是只回答文本 |
| 运行形态 | `stream`、conversation state、cache | 决定是一次性返回、增量返回还是复用上下文 |

可以把它看成一个最小统一心智模型：

```text
Request
  = 模型
  + 上下文
  + 输出约束
  + 工具定义
  + 推理与流式控制

Response
  = 最终文本 / 结构化结果 / 工具调用意图
  + token usage
  + stop reason
  + 可选 reasoning / streaming events
```

## 12.2 Chat Completions、Responses 与消息协议

### 12.2.1 三种常见接口范式

虽然抽象相似，但主流厂商在协议层已经分化出三种常见范式。

**第一种是 OpenAI 风格的通用 Responses / Chat 接口。** 这类接口强调统一输入抽象，既能接文本，也能接图片、文件、工具和结构化输出。OpenAI 当前主推 `Responses API`，核心思路是用 `input` 承载多模态内容，并在输出中统一返回 message、tool call 和 usage。它仍保留 `Chat Completions` 兼容路径，因此很多第三方也会优先兼容这套形态。OpenAI 官方文档还提供结构化输出、函数调用、提示缓存、图片输入、推理 effort 和 realtime 等能力。[OpenAI Responses API](https://developers.openai.com/api/reference/resources/responses/methods/create) [OpenAI Structured Outputs](https://developers.openai.com/api/docs/guides/structured-outputs) [OpenAI Prompt Caching](https://developers.openai.com/api/docs/guides/prompt-caching)

**第二种是 Anthropic 的 Messages / content blocks 范式。** 它没有把接口设计成“OpenAI Chat 的一个镜像”，而是更强调内容块和工具块。输入主体是 `messages`，每条消息的 `content` 可以是文本、图片等 block；工具使用也会回到消息内容流里。Anthropic 还把 extended thinking、prompt caching、strict tool use、parallel tool use 等能力直接纳入 Claude 平台文档体系。它的协议风格通常更适合把“消息内容”和“工具内容”统一看成一个事件流。 [Anthropic Messages API](https://platform.claude.com/docs/en/api/messages) [Anthropic Tool Use](https://platform.claude.com/docs/en/agents-and-tools/tool-use/overview) [Anthropic Prompt Caching](https://platform.claude.com/docs/en/build-with-claude/prompt-caching)

**第三种是兼容层范式。** DeepSeek 的定位很典型：它同时提供 OpenAI 兼容和 Anthropic 兼容入口。你可以继续使用 OpenAI SDK，把 `base_url` 改成 `https://api.deepseek.com`；也可以用 Anthropic 兼容入口 `https://api.deepseek.com/anthropic`。这种模式的价值在于迁移成本低，但也意味着“兼容”不一定等于“协议细节完全相同”，尤其是 reasoning 输出、缓存语义和模型专有参数。 [DeepSeek 首次调用 API](https://api-docs.deepseek.com/zh-cn/) [DeepSeek Anthropic API](https://api-docs.deepseek.com/zh-cn/guides/anthropic_api)

### 12.2.2 输入输出协议长什么样

把复杂的 SDK 名词去掉后，三家接口都可以还原成下面这两类模式。

**模式 A：对话消息数组**

```json
{
  "model": "xxx",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Explain KV cache."}
  ]
}
```

这类协议最常见于 OpenAI Chat Completions、Anthropic Messages 和 DeepSeek 的 OpenAI 兼容接口。它的优点是简单直观，缺点是当内容类型越来越多时，文本消息数组会逐渐演化成更复杂的 block 结构。

**模式 B：结构化输入块**

```json
{
  "model": "xxx",
  "input": [
    {
      "role": "user",
      "content": [
        {"type": "input_text", "text": "What is in this image?"},
        {"type": "input_image", "image_url": "..."}
      ]
    }
  ]
}
```

这类协议更像“统一内容容器”，适合多模态和工具混排。OpenAI 的 Responses API 就是这个方向。

输出侧也有两个常见层次：

- **最终回答层**：模型给你文本结果、JSON 结果或多模态输出。
- **中间动作层**：模型不直接给最终答案，而是先给出 tool call、thinking、stop reason、streaming delta。

所以一个工程上更完整的理解是：

```text
LLM API 输出 ≠ 只有一段字符串
LLM API 输出 = 文本结果 + 结构化状态 + 工具调用意图 + usage + stop reason
```

## 12.3 流式输出、结构化输出与 Tool Calling

结构化输出、流式返回和工具调用并不是三个孤立功能，而是模型被系统消费时最常见的三种“可编排输出形态”：

- **流式输出**解决交互延迟和事件驱动渲染；
- **结构化输出**解决机器可消费结果；
- **Tool Calling**解决模型向外部世界发起行动请求。

在 API 层，它们分别表现为：

- `stream` / streaming events
- `response_format` / JSON schema / strict schema
- `tools` / `tool_choice` / `tool_calls` / `tool_use`

真正的系统不会把这三者拆开看，而是放进同一个 Runtime 控制面：模型先返回文本、JSON 或工具调用意图，系统再根据当前任务协议决定是直接交付、继续追问、执行工具，还是把工具结果回填给模型。

### 12.3.1 大模型 API 的流式输出是什么

大模型 API 的流式输出（streaming），简单来说，就是让模型的回答像打字机一样逐段“跳”出来，而不是等完整答案全部生成后再一次性返回。

如果把非流式模式看成“厨师把整桌菜做完再一起上桌”，那么流式模式更像“旋转寿司”：模型每生成一小段 token，就立刻通过持久连接发送给客户端，前端马上就能渲染出来。

两种模式在工程上的区别可以概括为：

| 特性 | 非流式输出 | 流式输出 |
|:---|:---|:---|
| API 行为 | 服务端生成完整结果后一次性返回大 JSON | 服务端边生成边推送增量事件 |
| 首字延迟（TTFT）感知 | 高，用户会长时间盯着空白或 loading | 低，通常几百毫秒内就能看到首个 token |
| 用户感受 | 像“系统在等待” | 像“系统正在实时工作” |
| 长连接风险 | 内容过长时更容易超时 | 持续传输可维持连接活跃 |

所以，流式输出的核心价值不是“更酷”，而是把模型生成过程从“黑盒等待”变成“可感知、可消费、可中途响应的输出过程”。

### 12.3.2 为什么流式输出几乎是现代 LLM Apps 的标配

只要场景同时满足两个条件，流式输出几乎就是必选项：

1. 内容生成需要明显时间。
2. 前端有人类用户在等待结果。

在现代 LLM 应用里，最典型的四类场景如下。

**第一类是实时人机对话与助理。**  
这是最经典的场景，代表应用包括 ChatGPT、Claude、Kimi，以及企业内部问答助手和客服机器人。问题不在于模型能否最终答对，而在于用户是否愿意盯着一个静止加载动画等 5 到 10 秒。流式输出把等待拆成连续反馈，显著降低“网页卡死了”的焦虑感。

**第二类是长文本与内容创作。**  
当模型需要写长文、报告、翻译稿、分析说明甚至长段代码时，生成时间可能从数十秒上升到数分钟。流式输出的价值有两层：一是用户可以边生成边阅读、边检查；二是持续传输能降低网关超时和浏览器超时的风险。

**第三类是 AI 编程辅助与代码生成。**  
在 Copilot、Cursor、IDE 插件这类场景里，用户不是被动等待完整结果，而是在和模型做异步协同。模型流式输出函数后半段时，开发者可能已经开始阅读前半段结构并准备下一步操作。这种“人先看，模型继续写”的重叠过程，是代码场景中非常重要的效率来源。

**第四类是语音协同与实时交互。**  
在语音助手、实时翻译、电话助理和 Realtime Agent 场景里，流式输出几乎不是优化项，而是基础能力。只有把模型增量输出和流式 TTS 串起来，才能让系统做到“边想边说”；如果必须等完整文本生成完再播报，语音交互会出现明显而不自然的停顿。

从系统设计角度看，流式输出真正优化的不是模型本身的推理速度，而是**用户感知延迟**和**前后端协同方式**。

### 12.3.3 什么场景不需要，甚至不应该使用流式输出

流式输出虽然常见，但并不是所有场景都值得开启。对于纯后端消费、自动化链路或严格结构化结果，流式过程往往没有业务价值，反而会增加协议处理复杂度。

常见反例如下：

**数据结构化提取。**  
如果任务目标是从简历、合同、票据或日志中抽取结构化 JSON，后端真正需要的是最终那个可解析的完整对象。中间零碎 token 既不能直接 `JSON.parse()`，也不利于稳定重试，通常没有必要让消费方处理流式碎片。

**离线批处理。**  
例如夜间批量审核评论、批量改写标题、批量生成标签。这类任务关注的是吞吐量、稳定性和成本，而不是人类是否正在盯着屏幕。此时比起流式渲染，更重要的是队列调度、失败重试和批量并发控制。

**Agent 自动化工作流。**  
在多步骤 Agent 链路里，后续步骤经常依赖前一步的完整结论、完整 JSON 或完整工具结果。比如 Step 2 必须消费 Step 1 的最终结构化输出，才能决定下一步动作。这种场景下，增量 token 本身并没有稳定语义，过早消费反而容易让流程进入不确定状态。

所以，一个实用判断标准是：

```text
有人在等 + 可以边看边用
  -> 优先流式

机器在等 + 必须拿完整结果再继续
  -> 优先非流式
```

### 12.3.4 流式输出在协议层意味着什么

从 API 协议角度看，流式输出不是“把完整响应切碎”，而是把模型生成过程显式暴露成一个事件流。

工程上常见的流式事件包括：

- 文本增量 delta
- reasoning / thinking 增量
- tool call 增量或工具调用完成事件
- message 完成事件
- usage 或 stop reason 结束事件

也就是说，流式模式下客户端处理的不再是“一个最终 JSON”，而是：

```text
start
  -> delta
  -> delta
  -> delta
  -> tool_call / message / reasoning events
  -> completed
```

这也是为什么流式输出一旦进入系统设计，就不只是前端打字机效果，而是会影响：

- 前端如何增量渲染；
- 服务端如何转发和聚合事件；
- Runtime 如何处理中途中断、取消和超时；
- 日志和 trace 如何记录“过程输出”而不只是最终结果。

## 12.4 OpenAI、Anthropic、DeepSeek 等厂商差异

### 12.4.1 OpenAI、Anthropic、DeepSeek 的能力与协议差异

下面这张表总结了三家截至 **2026 年 7 月 1 日** 官方文档可确认的公开能力与协议形态。这里说的“支持”指官方文档明确提供对应能力；并不意味着所有模型、所有 SDK 或所有兼容层都完全等价。

| 维度 | OpenAI | Anthropic | DeepSeek |
|:---|:---|:---|:---|
| 主要接口范式 | `Responses API` 为主，也保留 `Chat Completions` | `Messages API` | OpenAI 兼容 + Anthropic 兼容 |
| 典型输入结构 | `input` 或 `messages` | `messages` + content blocks | 以 `messages` 为主，兼容两套 SDK |
| 工具调用 | 支持 `tools`、function calling、strict schema | 支持 tool use、`tool_choice`、strict tool use、parallel tool use | 支持 OpenAI 风格 `tools` / `tool_calls` |
| 结构化输出 | 支持 JSON schema、strict structured outputs | 支持 structured outputs 与 strict tool schema | 支持 JSON Output，`response_format={type: json_object}` |
| 多模态输入 | 官方文档支持文本、图片、文件、音频等多种输入路径 | 官方文档支持文本、图片、文件相关能力，围绕 messages/content blocks 展开 | 公开文档重点是文本、工具、思考模式；能力表述更偏兼容层 |
| 推理控制 | 官方文档提供 reasoning models 与 `reasoning_effort` | 官方文档提供 extended thinking / effort | 官方文档提供 thinking 开关和 `reasoning_effort` |
| 流式输出 | 支持 streaming events | 支持 streaming messages | 支持流式，thinking 和 content 可分别增量返回 |
| 上下文缓存 | 官方文档提供 prompt caching，`usage.prompt_tokens_details.cached_tokens` 可观测 | 官方文档提供 prompt caching，支持 `cache_control` 与显式 breakpoint | 官方文档称上下文硬盘缓存默认开启，返回 `prompt_cache_hit_tokens` / `prompt_cache_miss_tokens` |
| 兼容 SDK 迁移 | 原生 | 原生 | 迁移成本最低，适合复用 OpenAI / Anthropic SDK |
| 协议风格 | 统一平台型 | 内容块 / 事件流型 | 兼容层型 |

### 12.4.2 几个最值得工程上关注的差异

**第一，OpenAI 更像统一平台协议。**  
它把文本、图片、文件、结构化输出、工具调用、推理 effort、conversation state、realtime 和 prompt caching 放进一个越来越统一的平台接口体系里。对新项目来说，这种协议更适合做“一个入口接多种能力”的平台化设计。[OpenAI Responses API](https://developers.openai.com/api/reference/resources/responses/methods/create)

**第二，Anthropic 更强调消息内容块与工具事件流。**  
它的文档体系把 tool use、strict tool use、parallel tool use、prompt caching、extended thinking 组织得非常清楚。对于需要自己实现 Agent runtime 的团队，这种“消息内容即事件流”的风格比较容易映射到状态机和 orchestrator 上。[Anthropic Tool Use](https://platform.claude.com/docs/en/agents-and-tools/tool-use/overview)

**第三，DeepSeek 的核心价值是兼容和迁移友好。**  
它官方明确说明可通过修改 `base_url` 使用 OpenAI / Anthropic SDK，并提供 OpenAI 兼容的 tool calls、JSON output、thinking mode 与默认开启的上下文硬盘缓存。这对已有 OpenAI 风格代码库非常友好，但也要注意兼容层上的专有行为，例如思考模式下 `reasoning_content` 的回传规则。 [DeepSeek 首次调用 API](https://api-docs.deepseek.com/zh-cn/) [DeepSeek 思考模式](https://api-docs.deepseek.com/zh-cn/guides/thinking_mode)

### 12.4.3 工程选型怎么判断

如果你的系统目标是**统一接入多模态、结构化输出和平台级能力**，OpenAI 风格接口通常更适合作为默认抽象层。

如果你的系统目标是**围绕 tool use、长对话缓存、消息事件流自己做 Agent runtime**，Anthropic 的 Messages 范式会更值得认真学习。

如果你的系统目标是**快速兼容已有 OpenAI/Anthropic SDK，降低迁移成本，或在成本与能力之间做替代路线**，DeepSeek 这种兼容层模式会很有吸引力。

但不要把“兼容某家 SDK”误以为“所有协议细节和所有模型能力完全等价”。真正需要验证的至少包括：

1. 工具调用返回格式是否一致。
2. 严格 JSON 或 schema 约束是否真的可靠。
3. thinking / reasoning 字段是否需要单独处理。
4. streaming 事件粒度是否一致。
5. 缓存语义和 usage 字段是否一致。
6. 多模态输入、文件输入和 server-side tools 是否都被兼容。

一个成熟的做法不是直接把厂商 SDK 暴露到业务层，而是在你自己的系统里再包一层 provider adapter：

```text
Business Logic
  -> Provider Adapter
      -> OpenAI / Anthropic / DeepSeek
```

这样，真正稳定的不是供应商字段，而是你自己定义的抽象契约：输入消息、工具定义、结构化输出、usage 统计、错误类型和重试策略。

## 12.5 Agent 工作流如何映射为模型请求

### 12.5.1 Skill、Tool 与多轮 Brainstorming：如何把 Agent 工作流映射到模型请求

到这里还有一个常见误区：很多开发者已经在 Agent 框架里使用了 Skill、Tool、Workflow，于是会自然地以为这些概念都可以直接作为底层模型 API 的字段传进去。真实情况并不是这样。

在底层大模型协议里，**Tool 通常是原生字段**，例如 `tools`、`tool_choice`、`tool_calls`；但 **Skill 往往不是原生协议字段**。像 `superpowers:brainstorming` 这类 Skill，本质上更像一份“工作方法说明书”，它约束的是模型做事的顺序、提问方式、停止条件和交付格式，而不是一个底层 API 枚举值。

因此，一个更准确的映射关系是：

```text
Skill
  -> system prompt / developer prompt / injected context

Tool
  -> tools schema + tool_choice + tool result messages

Multi-turn conversation
  -> messages history
```

也就是说：

- **Skill 解决“模型应该按什么方法工作”**；
- **Tool 解决“模型可以调用哪些外部能力”**；
- **Messages history 解决“模型当前已经知道哪些对话上下文”**。

如果把一个 Agent 请求拆开看，它实际更像：

```text
system: 工作流规则、角色、边界
user / assistant: 多轮历史对话
tools: 可调用能力定义
current user turn: 本轮新任务
```

这也是为什么 Agent Runtime 不能只做一个简单的 `prompt -> completion` 封装。它至少还要负责三件事：

1. 把 Skill 压缩成适合当前任务的系统约束，而不是把整份文档原封不动塞进 prompt。
2. 维护多轮消息历史，保留澄清问题、用户回答、阶段性判断和工具结果。
3. 在模型返回 `tool_calls` 时执行工具，再把工具结果作为后续消息喂回模型。

### 12.5.2 一个具体例子：把 `superpowers:brainstorming` 映射进请求

以 `superpowers:brainstorming` 为例，它原始描述很长，但工程上真正要传给模型的，不是整份 Skill 文档，而是其中对当前任务最重要的几条约束：

- 先理解项目上下文；
- 每次只问一个澄清问题；
- 在设计被批准前，不进入实现；
- 信息足够后，先给出 2 到 3 个方案并说明 trade-off；
- 必要时才调用工具读取项目文件或目录。

这些内容可以被压缩成一个 `system` 消息。例如：

```json
{
  "role": "system",
  "content": "你是一个严格遵循 brainstorming 工作流的设计助手。规则：1. 先理解项目上下文。2. 每次只问一个澄清问题。3. 在用户批准设计前，不要写代码，不要进入实现。4. 信息足够后，先提出 2 到 3 个方案并说明 trade-off。5. 只有在缺少项目上下文时，才调用工具读取相关文件。"
}
```

注意这里传给模型的是**Skill 的运行约束摘要**，而不是 Skill 名称本身。换句话说，模型并不知道 `superpowers:brainstorming` 这个标识符意味着什么；真正起作用的是你注入进去的规则文本。

### 12.5.3 模拟任务：三轮 Brainstorming 之后的一次模型请求

下面这份请求模拟了一个真实场景：用户要为现有博客系统设计“PDF 自动摘要”能力；对话已经进行了三轮澄清；当前这一轮模型既要遵守 brainstorming 工作流，又可以按需调用工具读取项目目录和文件。

```json
{
  "model": "deepseek-v4-pro",
  "messages": [
    {
      "role": "system",
      "content": "你是一个严格遵循 brainstorming 工作流的设计助手。规则：1. 先理解项目上下文。2. 每次只问一个澄清问题。3. 在用户批准设计前，不要写代码，不要进入实现。4. 信息足够后，先提出 2 到 3 个方案并说明 trade-off。5. 只有在缺少项目上下文时，才调用工具读取相关文件。"
    },
    {
      "role": "user",
      "content": "我想给博客系统增加一个 PDF 自动摘要功能，请先帮我 brainstorm。"
    },
    {
      "role": "assistant",
      "content": "先确认一个关键问题：这个 PDF 摘要功能是给站点管理员离线生成文章摘要，还是给终端访客在线上传 PDF 实时生成摘要？"
    },
    {
      "role": "user",
      "content": "给站点管理员使用，主要是在写文章时辅助生成摘要，不是给终端访客开放上传。"
    },
    {
      "role": "assistant",
      "content": "明白了。再确认一个问题：你希望这个能力直接集成到现有博客构建流程里，还是先做成一个独立脚本或命令行工具？"
    },
    {
      "role": "user",
      "content": "先做成独立脚本，但最好后面能接进博客工作流。项目是 Hexo 博客，文章在 source/_posts 下面。"
    },
    {
      "role": "assistant",
      "content": "信息已经足够开始收敛方案，但我还缺少项目当前文章结构和脚本组织方式。接下来我会先读取相关目录或现有脚本，再决定给出哪种方案最合适。"
    },
    {
      "role": "user",
      "content": "可以，继续。"
    }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "list_project_files",
        "description": "列出项目中与当前任务相关的文件，用于理解目录结构和现有脚本组织。",
        "parameters": {
          "type": "object",
          "properties": {
            "path": {
              "type": "string",
              "description": "要查看的目录路径"
            },
            "pattern": {
              "type": "string",
              "description": "可选的文件匹配模式"
            }
          },
          "required": ["path"]
        }
      }
    },
    {
      "type": "function",
      "function": {
        "name": "read_project_file",
        "description": "读取项目文件内容，用于理解现有实现、脚本或文档。",
        "parameters": {
          "type": "object",
          "properties": {
            "path": {
              "type": "string",
              "description": "要读取的文件路径"
            }
          },
          "required": ["path"]
        }
      }
    }
  ],
  "tool_choice": "auto",
  "stream": false
}
```

这份请求体现了三层不同职责：

- `system` 承载的是 Skill 约束；
- `messages` 承载的是三轮 brainstorming 历史；
- `tools` 承载的是本轮可调用能力。

### 12.5.4 模型在这类请求后可能返回什么

收到上面的请求后，模型通常不会直接进入实现，而会在三种动作中选择其一：

1. **继续追问一个澄清问题**：如果它认为信息仍不足；
2. **直接给出 2 到 3 个设计方案**：如果它认为上下文已经足够；
3. **先发起 tool call**：如果它认为必须先读目录、读文件或查项目结构。

例如，如果模型决定先读取目录，它可能返回：

```json
{
  "choices": [
    {
      "message": {
        "role": "assistant",
        "tool_calls": [
          {
            "id": "call_1",
            "type": "function",
            "function": {
              "name": "list_project_files",
              "arguments": "{\"path\":\"source/_posts\",\"pattern\":\"*.md\"}"
            }
          }
        ]
      },
      "finish_reason": "tool_calls"
    }
  ]
}
```

这时 Runtime 需要执行 `list_project_files`，再把结果作为 `tool` 消息继续喂回模型。也就是说，**多轮 brainstorming + tool calling 并不是一次性请求，而是一个运行闭环**：

```text
Skill rules
  + conversation history
  + tools schema
  -> model proposes tool call
  -> runtime executes tool
  -> tool result returns to model
  -> model continues brainstorming
```

这个例子也说明了为什么 Skill、Tool 和底层模型 API 不能混为一谈。Skill 决定行为策略，Tool 决定能力边界，Messages 决定当前上下文，Runtime 决定如何把这些要素编排成可重复、可治理的对话事务。

### 12.5.5 真实样例：思考模式下的多轮对话响应

前面的例子偏向 Agent 工作流：有 Skill 约束，有 Tool schema，有可能触发 `tool_calls`。但在很多日常场景里，请求并不会走到工具调用，而只是一个带历史上下文的普通多轮对话。这个时候，思考模式的协议行为会更接近“模型先内部推理，再输出面向用户的自然回答”。

例如，先有这样一段多轮对话历史：

```text
system:
  你是一个实用主义的生活助手。回答要更自然、更完整、更像聊天助手，更适合普通用户阅读。先直接给结论，再用简洁易懂的方式解释原因，语气友好，不要过于生硬。

user:
  我想去洗车，洗车店距离我家 50 米。你说我应该开车过去还是走过去？

assistant:
  你应该开车过去。因为你是去洗车，车得一起到洗车店，走过去没有意义。50 米非常近，启动一下车开过去也不会造成什么额外负担。

user:
  如果我只是先去问价格，还没决定洗，那我应该走过去还是开车过去？
```

对应到一次实际请求，请求体可以写成：

```json
{
  "model": "deepseek-v4-pro",
  "messages": [
    {
      "role": "system",
      "content": "你是一个实用主义的生活助手。回答要更自然、更完整、更像聊天助手，更适合普通用户阅读。先直接给结论，再用简洁易懂的方式解释原因，语气友好，不要过于生硬。"
    },
    {
      "role": "user",
      "content": "我想去洗车，洗车店距离我家 50 米。你说我应该开车过去还是走过去？"
    },
    {
      "role": "assistant",
      "content": "你应该开车过去。因为你是去洗车，车得一起到洗车店，走过去没有意义。50 米虽然近，但这不影响结论。"
    },
    {
      "role": "user",
      "content": "如果我担心刚开过去 50 米，车还没热就到了，会不会伤车？"
    },
    {
      "role": "assistant",
      "content": "一般不用担心，这么短的距离不会因为“没热车”就对车辆造成什么明显影响。相比之下，更关键的是你本来就是要把车送去洗，所以把车开过去本身就是合理动作。"
    },
    {
      "role": "user",
      "content": "那如果我只是想先去问价格，再决定洗不洗呢？"
    }
  ],
  "thinking": {
    "type": "enabled"
  },
  "reasoning_effort": "high",
  "stream": false
}
```

这里可以看到，普通多轮对话在协议层并不复杂：只需要把前面的 `user` 和 `assistant` 消息继续保存在 `messages` 里，再追加当前轮的 `user` 问题即可。对于这个不涉及工具调用的场景，历史上下文的核心载体仍然是 `messages`，而不是额外的工作流状态对象。

在 DeepSeek 的思考模式下，请求模型后，可能会得到下面这样的返回：

```json
{
  "id": "3a0689fc-197d-4e07-b89e-f8347aefc1a6",
  "object": "chat.completion",
  "created": 1782898620,
  "model": "deepseek-v4-pro",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "那走过去更合适。\\n既然只是去问价格、还没决定要洗，就走几步路过去方便又省事，不用特意发动车子、找停车位。问完如果觉得合适，再回去把车开过来也就一分钟的事儿，完全来得及。",
        "reasoning_content": "我们需要理解用户的新问题：用户之前问的是去洗车，现在问如果只是想去问价格再决定洗不洗，那应该开车还是走过去。之前我建议开车过去，因为车也要洗。但现在场景变了，只是去问价格，可能不需要车在场。\\n\\n作为生活助手，我需要给出实用的建议。结论应该是：走过去更合适。因为距离只有50米，走路几分钟就到，方便又省事，还省得启动车辆、找车位。如果决定洗车了，再回去开车过来也不迟。这样既环保又高效。\\n\\n语气要友好自然，先直接给出结论，再解释原因。不需要过度复杂化。"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 185,
    "completion_tokens": 191,
    "total_tokens": 376,
    "completion_tokens_details": {
      "reasoning_tokens": 136
    }
  }
}
```

这个样例很适合说明思考模式下的三个关键点。

**第一，`content` 和 `reasoning_content` 是两层不同输出。**  
`content` 是真正给终端用户看的自然语言回答；`reasoning_content` 是模型在输出前的思考轨迹。你可以看到，这个例子里给用户的最终回答很短，但 `reasoning_content` 明确记录了模型如何利用上一轮上下文修正结论：前一轮因为“要洗车”所以建议开车，这一轮因为“只是去问价格”所以建议走过去。

**第二，多轮对话真正生效的是历史消息，而不是把整段思维链反复塞回去。**  
这个例子没有发生工具调用，因此下一轮继续对话时，通常只需要把前面的 `user` 和 `assistant.content` 放回 `messages`。按照 DeepSeek 文档，在“无工具调用”的场景里，之前轮次的 `reasoning_content` 后续传回去会被忽略。也就是说，这类请求的核心是“历史对话状态”，而不是“长期保存全部思维链”。

**第三，思考模式会显著增加 token 消耗。**  
这个例子里 `prompt_tokens` 是 185，但 `completion_tokens` 达到 191，其中 `reasoning_tokens` 就占了 136。说明最终展示给用户的短回答背后，模型实际上做了更长的内部推理。工程上这意味着：一旦开启 thinking，不仅输出更稳定，成本和延迟也会相应上升。

如果把这个例子和前面的 brainstorming/tool calling 示例对照起来，可以得到一个更完整的结论：

- **普通多轮对话**：重点是维护 `messages` 历史；
- **思考模式**：重点是区分 `content` 与 `reasoning_content`；
- **工具调用场景**：重点是追加 `tool_calls`、`tool` 消息，并在需要时回传 `reasoning_content`；
- **Skill 场景**：重点是把工作流规则压缩成 `system` 或注入上下文。
