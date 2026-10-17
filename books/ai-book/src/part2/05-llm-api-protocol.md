# 第17章 LLM API 协议：模型能力如何被系统消费

> 模型能力不是直接被业务系统“使用”，而是先被抽象成输入、消息、工具、结构化输出和流式事件，再进入 Runtime、编排器和工具系统。

版本与范围：滚动官方资料核验于2026-10-18；示例是未调用真实模型的教学协议记录，模型名与额度必须从已验证manifest选择。

## 引言

理解 LLM 的原理，只能回答“模型为什么能工作”；理解 LLM API 协议，才能回答“系统怎样把模型接进来”。在工程落地里，这一步经常被低估。很多团队熟悉 Token、上下文窗口、Transformer 和 Sampling，却在真正接入模型时发现一连串新问题：

1. 一次调用里，系统究竟该传什么？
2. `system`、`messages`、`input`、`content blocks` 这些字段本质上分别代表什么？
3. 模型返回的为什么不只是文本，还有工具调用、结构化结果、流式事件和 usage？
4. OpenAI、Anthropic、DeepSeek 看起来都“差不多”，实际差别在哪？
5. 一个 Agent 工作流，怎样被映射成多轮请求、工具回填和状态延续？

这一章讨论的不是 SDK 语法糖，而是模型协议层的统一抽象。你可以把它看成 Agent Runtime 和模型能力之间的接口层：往左连接 Prompt、Context、Tool、Skill 和 Workflow，往右连接 OpenAI、Anthropic、DeepSeek 等不同提供方。

如果说第 11 章定义了持久Runtime，第13章给出了Agent架构总纲，那么本章回答的就是：**Runtime 究竟怎样消费模型能力。**

## 17.1 统一抽象：输入、指令、上下文与输出

学到这里，如果只知道 token、上下文和 Transformer 还不够。工程上真正调用大模型时，还要理解这些能力如何暴露成 API 协议。否则你会知道“模型能做什么”，却不知道“系统该怎么把这些能力接进来”。

这也是为什么 API 协议属于 Agent Runtime 的核心知识。它不是单纯的 SDK 用法，而是模型能力在工程边界上的投影：

- 上下文窗口会表现为 `messages`、`input`、`system`、`conversation state` 等输入结构；
- 结构化输出会表现为 JSON mode、JSON schema、strict schema；
- 工具调用会表现为 `tools`、`tool_choice`、`tool_calls` 或 `tool_use`；
- 长上下文成本会表现为 prompt caching、cache hit 指标和上下文压缩；
- 多模态能力会表现为文本、图片、文件、音频等不同 content block；
- 推理能力会表现为 reasoning / thinking 开关、effort、流式事件和 token 统计。

### 17.1.1 协议抽象：LLM API 本质上在传什么

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

## 17.2 Chat Completions、Responses 与消息协议

### 17.2.1 不同接口保留自己的消息语义

OpenAI Responses使用输入/输出项表达文本与工具调用；工具结果按call_id回填，执行仍由应用负责。[[1]](#ref-1)Anthropic Messages的流式表达以message/content block为单位。[[2]](#ref-2)DeepSeek提供兼容形态，但strict工具模式还要求其beta入口和受支持Schema，不能只改base_url后假定全部参数相同。[[3]](#ref-3)

适配层保留原始provider、API形态、模型版本、finish/error和usage字段，再导出稳定应用类型。未识别能力必须明确拒绝或按已评估策略降级，不悄悄丢掉system规则、工具定义或图像输入。

### 17.2.2 输入输出协议长什么样

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

这是Chat风格示意；不能原样发送给Anthropic Messages，因为它的system指令位于顶层，具体结构由对应适配器生成。它的优点是简单直观，缺点是当内容类型越来越多时，文本消息数组会逐渐演化成更复杂的 block 结构。

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

## 17.3 流式输出、结构化输出与 Tool Calling

结构化输出、流式返回和工具调用并不是三个孤立功能，而是模型被系统消费时最常见的三种“可编排输出形态”：

- **流式输出**解决交互延迟和事件驱动渲染；
- **结构化输出**解决机器可消费结果；
- **Tool Calling**解决模型向外部世界发起行动请求。

在 API 层，它们分别表现为：

- `stream` / streaming events
- `response_format` / JSON schema / strict schema
- `tools` / `tool_choice` / `tool_calls` / `tool_use`

真正的系统不会把这三者拆开看，而是放进同一个 Runtime 控制面：模型先返回文本、JSON 或工具调用意图，系统再根据当前任务协议决定是直接交付、继续追问、执行工具，还是把工具结果回填给模型。

### 17.3.1 大模型 API 的流式输出是什么

大模型 API 的流式输出（streaming），简单来说，就是让模型的回答像打字机一样逐段“跳”出来，而不是等完整答案全部生成后再一次性返回。

如果把非流式模式看成“厨师把整桌菜做完再一起上桌”，那么流式模式更像“旋转寿司”：模型每生成一小段 token，就立刻通过持久连接发送给客户端，前端马上就能渲染出来。

两种模式在工程上的区别可以概括为：

| 特性 | 非流式输出 | 流式输出 |
|:---|:---|:---|
| API 行为 | 服务端生成完整结果后一次性返回大 JSON | 服务端边生成边推送增量事件 |
| 首字延迟（TTFT）感知 | 高，用户会长时间盯着空白或 loading | 可以先看到增量；实际TTFT取决于排队、模型与网络 |
| 用户感受 | 像“系统在等待” | 像“系统正在实时工作” |
| 长连接风险 | 内容过长时更容易超时 | 持续传输可维持连接活跃 |

所以，流式输出的核心价值不是“更酷”，而是把模型生成过程从“黑盒等待”变成“可感知、可消费、可中途响应的输出过程”。

### 17.3.2 为什么流式输出几乎是现代 LLM Apps 的标配

只要场景同时满足两个条件，流式输出通常值得评估：

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

### 17.3.3 什么场景不需要，甚至不应该使用流式输出

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

### 17.3.4 流式输出在协议层意味着什么

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

### 17.3.5 SSE分帧与完成判断

SSE是UTF-8文本，空行结束一帧，event给类型，data给载荷，id可用于EventSource的Last-Event-ID机制。TCP读取块不是帧边界；同一JSON甚至UTF-8字符可能跨读取块，必须增量解码、组帧再解析。WHATWG并未规定某模型服务必须保存重放日志。[[4]](#ref-4)

OpenAI Responses使用response.created、response.output_text.delta、response.completed等typed events；中途error或连接断开不能当正常完成。[[5]](#ref-5)Anthropic还区分content_block_stop、message_delta与message_stop；工具块结束不等于业务动作成功。[[2]](#ref-2)

下列是本书**应用任务流**的合法SSE教学样例，应用字段由第11章约定，不能冒充任何供应商原始流：

```text
id: run-demo:41
event: step.completed
data: {"run_id":"run-demo","event_seq":41,"step_id":"s1","state_version":7,"artifact_ref":"answer-v1"}

id: run-demo:42
event: run.completed
data: {"run_id":"run-demo","event_seq":42,"state_version":8,"result_ref":"answer-v1","verification_ref":"check-v1"}

```

每帧后的空行不可省。前端按run_id/event_seq去重更新视图，再查询结果接口核对当前状态；这项回放保证来自应用的持久事件与保留策略，不来自SSE语法。

## 17.4 供应商适配与错误分类

### 17.4.1 能力矩阵必须绑定版本

| 维度 | OpenAI Responses | Anthropic Messages | DeepSeek工具接口 | 适配验收 |
| --- | --- | --- | --- | --- |
| 输出组织 | typed output item/event | message/content block | 按目标API形态返回 | 保留provider原始类型，未知项不能误判完成 |
| 工具关联 | function_call与function_call_output按call_id | tool_use与tool_result关联 | tool_calls与结果关联 | 允许并行、错序结果按ID归并，不能按位置配对 |
| 参数增量 | function_call_arguments.delta/done | input_json_delta后完成块 | 按所选端点文档组装 | 全量参数完整后解析并校验，片段不执行 |
| 严格结构 | 官方受支持Schema子集 | 按目标模型/端点核验 | strict beta需beta入口及支持子集 | 不支持时能力错误，不静默放宽 |
| 结束/错误 | completed/failed/incomplete等分别处理 | stop_reason及message_stop、error分别处理 | 按响应finish/error原字段 | 保留原原因，不把连接EOF当成功 |
| 续接 | 具体API状态功能另验 | 文本续写策略随模型版本变化 | 不假定原流游标重放 | 生成续写和业务恢复分别测试 |

上述事实分别依据三家已读协议页。[[5]](#ref-5)[[1]](#ref-1)[[6]](#ref-6)[[2]](#ref-2)[[3]](#ref-3)矩阵表达核对维度，**没有进行真实端点兼容性测试**；具体model/API/SDK版本、schema hash和支持能力写入第8–9章manifest，未测项不能填“测试通过”。原生工具、缓存、多模态和推理参数不得在未核对时泛化为所有模型都支持。

### 17.4.2 稳定错误类型与重试责任

| 应用类别 | 例子 | 下一步 |
| --- | --- | --- |
| capability_unsupported | 模型不支持工具或Schema关键字 | 调整已评估适配组合，不能盲重试 |
| request_invalid / auth_denied | 消息不合法、凭据/权限拒绝 | 修配置或授权，禁止模型绕过 |
| rate_limited / transient_provider | 限流、暂态服务错误 | 单一责任层在总deadline和预算内退避 |
| stream_incomplete | 中断、长度截断、半个工具参数 | 标未完成；保留可展示内容，不执行参数 |
| protocol_invalid | JSON损坏、关联ID缺失、事件不相容 | 停止该路径，隔离样本与版本，不能猜关联 |
| model_refusal | 明确拒绝 | 交付拒绝或允许的替代，不循环改写绕过 |
| tool_result_unknown | 工具请求已发但结果未知 | 回第11章Operation查证/对账，禁止直接再执行 |

记录原HTTP状态、provider错误码、request/response/call ID、attempt及可脱敏上下文。重试模型生成可能计费，也可能生成不同工具提案；不能因生成重试创建新的业务幂等键。参数/授权变化必须重新校验。

## 17.5 多轮工具关联与应用重连

### 17.5.1 调用ID不是业务Operation

教学Responses输出项与回填项分别如下，output里的JSON字符串是工具已返回的观察，不能由模型编造。[[1]](#ref-1)

```json
{"type":"function_call","id":"fc_demo","call_id":"call_demo","name":"lookup_document","arguments":"{\"document_id\":\"policy-7\"}"}
```

```json
{"type":"function_call_output","call_id":"call_demo","output":"{\"status\":\"ok\",\"version\":\"v3\",\"evidence_ref\":\"ev7\"}"}
```

应用先校验完整arguments、工具版本与当前权限，再执行并关联结果；同一call_id冲突内容应拒绝而非覆盖。保留协议要求的相关输出项，缺结果不能伪装为已完成对话。call_id只关联一轮模型调用，业务Operation的稳定幂等键由第11章另行维护。Skill是应用加载的程序性上下文，不自动变成供应商原生工具或权限。

### 17.5.2 断线重连的两个层面

教学时序：客户端订阅应用run-demo，确认事件40后断线；服务端继续按权威Run状态执行，先提交Step/产物/Event 41，再形成终态Event 42；客户端重连携带40，授权检查后重放41、42并去重，最后查询run结果。游标过期时使用第11章快照/重新订阅契约，不伪造供应商续流。

如果断的是模型供应商连接且工具参数只收到一半，adapter把生成attempt记为stream_incomplete，不能据半个JSON执行工具。若已执行工具但结果未知，先查询Operation而不是重新生成后再执行一次。取消只表达停止意图，迟到结果仍需要写入事实账本。

Anthropic文档的续写策略随模型版本区分，部分工具/思考块不能按文本方式恢复；这种“带部分输出重新请求”不是从旧SSE事件ID重放，也不保证续写逐字相同。[[2]](#ref-2)应用事件回放的保留期、权限、重复和过期语义仍由[第11章](infra-11-runtime.md)负责；第27章Gateway的具体实现需独立核验。

## 本章小结

Provider Adapter把消息、工具关联、完整输出与错误映射成应用可消费契约，同时保留供应商差异。SSE只提供传输分帧；参数完整、结构/语义/业务校验、授权和Operation才决定可否执行。模型生成中断、任务执行中断和用户连接断开必须分别处理。兼容声明要有版本和测试，教学样例不能冒充端点实测。


## 参考资料

1. <a id="ref-1"></a> OpenAI，[Function calling](https://developers.openai.com/api/docs/guides/function-calling)，rolling; 2026-10-18。
2. <a id="ref-2"></a> Anthropic，[Streaming messages](https://platform.claude.com/docs/en/build-with-claude/streaming)，rolling; 2026-10-18。
3. <a id="ref-3"></a> DeepSeek，[Tool Calls](https://api-docs.deepseek.com/guides/tool_calls/)，rolling; 2026-10-18。
4. <a id="ref-4"></a> WHATWG，[HTML Standard: Server-sent events](https://html.spec.whatwg.org/multipage/server-sent-events.html)，living; 2026-10-16。
5. <a id="ref-5"></a> OpenAI，[Streaming API responses](https://developers.openai.com/api/docs/guides/streaming-responses)，rolling; 2026-10-18。
6. <a id="ref-6"></a> OpenAI，[Structured model outputs](https://developers.openai.com/api/docs/guides/structured-outputs)，rolling; 2026-10-17。
