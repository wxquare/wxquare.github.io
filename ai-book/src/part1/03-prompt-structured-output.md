# 第3章 Prompt Engineering 与结构化输出

> Prompt 的目标不是“让模型更听话”，而是把任务边界、上下文、输出格式和失败处理说清楚。

## 引言

很多人把 Prompt Engineering 理解成“写一句神奇咒语”。这种理解只适合 demo，不适合生产系统。

在 Agent 系统里，Prompt 更像一份运行时协议：

- 它定义 Agent 的角色和边界；
- 它告诉模型可以使用哪些上下文；
- 它约束输出格式；
- 它规定遇到不确定性时如何处理；
- 它和工具 schema、guardrails、eval dataset 一起组成质量闭环。

好的 Prompt 不追求花哨，而追求可读、可测、可版本化。

---

## 1. Prompt的四层结构

一个生产级 Prompt 通常可以拆成四层：

```text
System Instruction
  ↓
Task Instruction
  ↓
Context
  ↓
Output Contract
```

### System Instruction

定义 Agent 的身份、职责和硬边界。

```text
你是一个生产告警诊断 Agent。
你的职责是帮助值班工程师分析告警原因，整理证据，并给出处理建议。
你不能直接执行高风险修复动作，例如重启服务、回滚发布或修改生产配置。
```

### Task Instruction

定义本次任务要完成什么。

```text
请根据告警信息、指标、日志和 runbook，判断可能原因，输出风险等级和建议动作。
```

### Context

提供模型可以依赖的信息，例如用户输入、检索文档、工具结果、历史对话摘要。

### Output Contract

定义输出格式和字段含义。

```json
{
  "summary": "一句话诊断结论",
  "risk_level": "low|medium|high",
  "evidence": [],
  "recommended_actions": [],
  "need_human_confirm": true
}
```

---

## 2. Prompt不是权限系统

一个常见错误是把安全边界写在 Prompt 里，然后相信模型一定遵守。

```text
请不要查询用户无权访问的数据。
```

这句话有价值，但不能作为真正的权限控制。真正的权限必须放在后端：

- 检索前按用户权限过滤文档；
- 工具调用前校验 user_id 和 permission；
- 高风险工具需要审批 token；
- 输出前做敏感信息脱敏。

Prompt 的作用是引导模型，后端系统的作用是强制约束。

---

## 3. 结构化输出

Agent 系统通常不应该只输出自然语言。自然语言适合给人看，但不适合驱动后续流程。

### 为什么需要结构化输出

- 方便后端解析；
- 方便做格式校验；
- 方便进入 workflow 的下一步；
- 方便做 eval；
- 方便记录 trace 和审计。

### JSON输出示例

```text
请严格输出 JSON，不要输出 Markdown，不要添加解释。

字段要求：
- intent: string，用户意图
- confidence: number，0 到 1
- required_tools: string[]
- need_clarification: boolean
- clarification_question: string|null
```

输出：

```json
{
  "intent": "diagnose_alert",
  "confidence": 0.82,
  "required_tools": ["query_metrics", "search_logs"],
  "need_clarification": false,
  "clarification_question": null
}
```

### 结构化输出的防线

只靠 Prompt 要求 JSON 还不够，生产系统需要三层防线：

```text
Prompt约束 → JSON Schema校验 → 失败重试或降级
```

如果解析失败，可以让模型修复格式：

```text
上一次输出不是合法 JSON。请只修复格式，不要改变字段含义。
```

如果连续失败，则进入 fallback，避免 Agent 卡死。

---

## 4. Few-shot示例

Few-shot 的价值不是“让模型模仿语气”，而是降低任务歧义。

适合放 few-shot 的场景：

- 意图分类；
- 风险分级；
- 工具选择；
- 输出格式复杂；
- 业务术语容易混淆。

### 风险分级示例

```text
示例1：
用户请求：查看 order-service 最近 30 分钟 CPU 指标
风险等级：read_only
原因：只读查询，不改变系统状态

示例2：
用户请求：帮我重启生产环境 order-service
风险等级：high
原因：会影响生产服务，需要人工确认

示例3：
用户请求：创建一个故障跟进工单
风险等级：medium
原因：有写入动作，但风险可控，需要审计
```

Few-shot 不要无限堆。示例太多会增加成本，也可能让模型过拟合少数模式。

---

## 5. Prompt版本管理

Prompt 是生产系统的一部分，应该像代码一样管理。

推荐记录：

```yaml
prompt_id: dod_agent_diagnosis
version: v3
owner: sre-platform
created_at: 2026-04-27
change_log:
  - 增加高风险动作拒绝策略
  - 强制输出 evidence 字段
eval_baseline:
  task_success_rate: 0.84
  unsafe_action_rate: 0.00
```

上线时要能回答：

- 这个 Prompt 改了什么？
- 哪些 eval case 变好了？
- 哪些 case 变差了？
- 是否支持快速回滚？

---

## 6. Prompt与工具Schema的关系

工具调用场景下，Prompt 和工具 schema 要一起设计。

Prompt 负责解释“什么时候用工具”：

```text
如果用户需要实时指标，请使用 query_metrics。
如果用户只是询问概念，不要调用工具。
```

Schema 负责限制“工具参数长什么样”：

```json
{
  "service": {"type": "string"},
  "metric": {"type": "string", "enum": ["cpu", "memory", "error_rate"]},
  "window_minutes": {"type": "integer", "minimum": 1, "maximum": 120}
}
```

Prompt 不能弥补糟糕的 schema。工具名、描述和参数越清晰，模型选错工具的概率越低。

---

## 7. 常见失败模式

| 失败模式 | 表现 | 修复方式 |
|:---|:---|:---|
| 指令冲突 | 前面说要简洁，后面说要详细 | 拆分优先级，减少冲突 |
| 输出漂移 | 偶尔输出 Markdown 或解释 | 加 schema 校验和重试 |
| 工具误用 | 不该查工具时查工具 | 增加 few-shot 和工具描述 |
| 过度拒答 | 明明有依据却拒答 | 区分“不确定”和“无权限” |
| 编造依据 | citation 不存在 | 输出后做 citation checker |
| 成本过高 | system prompt 太长 | 抽取稳定规则，压缩上下文 |

---

## 8. 面试表达

如果面试官问“你怎么做 Prompt 工程”，不要只讲技巧，可以这样回答：

```text
我会把 Prompt 当成运行时协议，而不是一句提示词。

首先拆成 system instruction、task instruction、context 和 output contract。
其次，关键输出必须结构化，后端用 JSON Schema 校验。
第三，Prompt 要和工具 schema、guardrails、eval dataset 一起设计。
最后，Prompt 要版本化，每次上线前跑回归 eval，并支持快速回滚。
```

---

## 小结

Prompt Engineering 的核心不是“会写提示词”，而是能把模型行为收敛进可测试、可观测、可回滚的工程系统。

好的 Prompt 有四个特征：

1. 边界清楚；
2. 上下文明确；
3. 输出可解析；
4. 可以被 eval 验证。
