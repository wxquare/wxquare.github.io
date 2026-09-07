# 第7章 LLM 能力边界与架构约束

> "Understanding the limits of AI is as important as understanding its capabilities." （理解AI的边界与理解它的能力同样重要）

## 引言

在进入 Prompt、Context、Harness 和 Agent 架构之前，我们需要先理解 LLM 的**能力边界**和**工程化要点**。许多 AI 工程问题并不是工具不够强，而是我们把模型当成了确定性系统。

许多 Agent 系统失败的根源，不是架构设计问题，而是对 LLM 能力的**误解**或**过度期待**。本章将系统梳理 LLM 的能力边界、常见陷阱和工程化最佳实践。

---

## 7.1 LLM 的能力边界

### 不要把 LLM 能力写成通用准确率

LLM 的能力不是一个固定百分比，而是一个条件函数：

```text
能力表现 = f(
  模型版本,
  任务分布,
  Prompt / System Prompt,
  上下文质量,
  是否可调用工具,
  解码参数,
  输出约束,
  评测指标,
  人工兜底策略
)
```

同一个模型，在“封闭标签分类”“开放式事实问答”“多文件代码修改”“带工具的告警诊断”上表现会完全不同。即使是同一个任务，换一个 prompt、换一批数据、换一个模型 snapshot、是否允许检索和工具调用，结果也会变化。

因此，本书不再给出“文本生成 90%+、代码生成 85%+”这类通用数字。更工程化的写法是：**先描述任务条件，再定义评测口径，最后给出可复现的本地 eval 结果**。

一个负责任的能力结论应该长这样：

```text
任务：客户反馈分类
模型：model-x-2026-05-01
输入：最近 30 天人工标注的 800 条中文客服反馈
标签集：物流 / 质量 / 价格 / 客服 / 售后 / 其他
Prompt：v3，包含 6 个 few-shot 示例
输出约束：JSON Schema，category 必须来自枚举
指标：macro F1、各类别 precision / recall、拒答率
基线：规则分类器、上一版 prompt、上一版模型
结论：只对这批数据、这个标签集和这个 prompt 生效
```

没有这些条件，准确率数字没有工程意义。

### 从关键论文理解 LLM 的能力来源

理解 LLM 的能力边界，不能只看产品发布会和排行榜，还要回到几个关键研究脉络。下面这些论文不适合作为“历史知识”死记，而应该作为工程判断的底层地图：模型为什么会有上下文学习、为什么需要对齐、为什么会幻觉、为什么 Agent 必须引入工具和外部状态。

| 层次 | 关键论文 | 讲清楚了什么 | 对工程实践的影响 |
|:---|:---|:---|:---|
| 架构底座 | [Attention Is All You Need](https://arxiv.org/abs/1706.03762) | Transformer 用自注意力替代 RNN / CNN，允许模型在上下文内建立 token 之间的关系 | Context 是模型推理的数据平面；长上下文有成本，注意力不是长期记忆 |
| 规模化规律 | [Scaling Laws for Neural Language Models](https://arxiv.org/abs/2001.08361)、[Training Compute-Optimal Large Language Models](https://arxiv.org/abs/2203.15556) | 模型损失会随参数、数据、算力呈规律性下降；Chinchilla 进一步强调模型规模和训练 token 要匹配 | “更大”不等于“更好”；选型要看模型、数据、推理成本和任务分布 |
| 上下文学习 | [Language Models are Few-Shot Learners](https://arxiv.org/abs/2005.14165) | 大模型可以通过自然语言指令和少量示例在推理时适配任务，无需每个任务都 fine-tune | Prompt / few-shot 是运行时任务协议，但不是可靠训练；示例质量直接影响输出 |
| 指令对齐 | [Training Language Models to Follow Instructions with Human Feedback](https://arxiv.org/abs/2203.02155)、[Direct Preference Optimization](https://arxiv.org/abs/2305.18290) | Base model 会续写文本，assistant model 经过 SFT、RLHF 或 DPO 后更倾向于遵循人类意图 | 对齐改善可用性和偏好匹配，但不能消除事实错误、越权风险和不确定性 |
| 推理诱导 | [Chain-of-Thought Prompting](https://arxiv.org/abs/2201.11903) | 对复杂任务，让模型生成中间推理步骤能显著改善部分数学、常识和符号推理任务 | 分步推理可以提升表现和可调试性，但推理链不是证明，仍需 verifier / test |
| 外部知识 | [Retrieval-Augmented Generation](https://arxiv.org/abs/2005.11401) | 参数记忆可以存储知识模式，但对新知识、私有知识和精确引用不可靠 | 企业知识问答、合规回答和事实引用应优先设计 RAG，而不是依赖模型记忆 |
| 工具与行动 | [ReAct](https://arxiv.org/abs/2210.03629)、[Toolformer](https://arxiv.org/abs/2302.04761) | 模型可以交替进行推理和行动，也可以学习何时调用 API、搜索、计算器等工具 | Agent Runtime 的核心是 Thought / Action / Observation 循环，而不是单次问答 |
| 高效适配 | [LoRA](https://arxiv.org/abs/2106.09685)、[QLoRA](https://arxiv.org/abs/2305.14314)、[LLaMA](https://arxiv.org/abs/2302.13971) | 开放模型和参数高效微调降低了私有化、领域适配和本地部署门槛 | 企业落地不一定训练 foundation model，更多是模型路由、RAG、微调、蒸馏和治理 |

这组论文可以归纳成一个三层视角：

```text
Pre-training:
  从海量 token 中学习语言、代码和世界知识的统计结构。

Post-training:
  通过指令数据、偏好数据和安全数据，让模型更像一个可交互助手。

Inference-time system:
  通过 prompt、context、RAG、tool、memory、eval、policy 和 trace，
  把概率生成模型放进可验证、可审计、可回滚的工程系统。
```

这也解释了为什么现代 LLM 系统不能只讨论“模型本身”：

- Transformer 让模型能利用上下文，但上下文不是数据库；
- Scaling Law 让能力提升可预测，但不能保证某个业务任务一定可靠；
- GPT-3 证明了 few-shot 的价值，但 prompt 仍需要评测和版本管理；
- RLHF / DPO 让模型更会听指令，但不等于输出一定真实；
- CoT / ReAct 让模型更会拆解任务，但生产系统还需要工具验证；
- RAG / Toolformer 说明外部知识和外部工具不是补丁，而是 LLM 架构的一部分。

因此，本书讨论的 Prompt、Context、Harness、RAG、Tool、Memory、Eval 和 Agent Runtime，本质上都是围绕同一个目标展开：**把一个强大的概率生成模型，约束成一个可用于生产系统的工程组件**。

### 从能力边界到三层工程控制面

理解 LLM 的能力边界后，接下来最重要的问题是：工程系统应该在哪些层面约束它？

本书把这个问题拆成三层控制面：

```text
Prompt Engineering  -> 任务协议
Context Engineering -> 信息架构
Harness Engineering -> 运行环境
```

它们对应 LLM 的三个核心限制：

- 模型会主动补全模糊目标，所以需要 Prompt Engineering 明确任务协议；
- 模型只能基于输入窗口工作，所以需要 Context Engineering 构建可信工作区；
- 模型不能天然拥有权限、状态、验证和恢复能力，所以需要 Harness Engineering 提供运行环境。

更具体地说：

| 层 | 核心问题 | 主要产物 | 负责什么 | 不负责什么 |
|:---|:---|:---|:---|:---|
| Prompt Engineering | 模型应该怎么做？ | Task Protocol、Output Contract、Reasoning Contract、Tool Contract | 定义角色、任务步骤、输出结构、工具使用规则和失败策略 | 不负责提供真实数据，不负责执行权限和系统验证 |
| Context Engineering | 模型应该看什么、信什么？ | Context Package、Evidence Package、Context Type System、Memory / RAG 策略 | 选择、过滤、标注、压缩和组织任务所需信息 | 不负责工具执行，不替代任务协议，也不做最终权限控制 |
| Harness Engineering | 模型如何安全运行？ | Agent Runtime、Tool Runtime、Policy Engine、Workflow、Verifier、Eval、Trace | 控制工具调用、审批、状态、验证、护栏、观测、回滚和治理闭环 | 不替代 Prompt 的任务表达，也不替代 Context 的信息质量 |

用一个告警诊断任务来看会更直观。

用户请求：

```text
checkout-api P95 延迟升高，请分析可能原因。
```

Prompt Engineering 负责把任务协议说清楚：

```text
必须先收集指标、日志和部署记录。
不能在没有证据时断言根因。
输出必须包含 hypothesis、evidence、confidence 和 next_steps。
高风险动作只能建议，不能直接执行。
```

Context Engineering 负责构建可信工作区：

```yaml
context_package:
  current_metrics:
    source: prometheus
    time_range: "last_30m"
    trust: current_fact
  error_logs:
    source: log_platform
    time_range: "last_30m"
    trust: current_fact
  deployments:
    source: deploy_system
    trust: authoritative
  runbook:
    source: checkout_latency_runbook
    updated_at: "2026-04-01"
    trust: authoritative
  historical_incident:
    usage_rule: "reference_only"
```

Harness Engineering 负责把行动放进运行边界：

```text
只暴露只读诊断工具；
每次工具调用先经过 Policy Engine；
高风险动作必须人工审批；
每一步写入 trace；
最终结论经过 evidence validator；
失败样本进入 regression eval；
发布前由 Release Gate 检查旧失败是否复发。
```

判断边界时，可以用一个简单规则：

- 如果问题是“模型应该按什么协议完成任务”，属于 Prompt Engineering；
- 如果问题是“模型应该拿到哪些信息，以及这些信息是否可信”，属于 Context Engineering；
- 如果问题是“模型能调用什么、谁来审批、如何验证和追踪”，属于 Harness Engineering。

这也是本书第一部分的展开顺序：先理解能力边界，再设计任务协议，再设计信息架构，最后进入运行环境。

### 从任务类型理解 LLM 擅长什么

LLM 的强项来自预训练目标：它擅长在给定上下文中建模语言、代码和知识模式。因此，它更适合处理“语义清晰、答案空间可约束、允许外部验证”的任务。

| 任务类型 | 可靠性判断 | 工程前提 |
|:---|:---|:---|
| 改写、总结、翻译 | 通常较稳定 | 输入材料足够完整，允许人工抽查 |
| 封闭标签分类 | 较容易工程化 | 标签定义清晰，有标注集，有混淆矩阵 |
| 结构化信息抽取 | 较容易工程化 | 使用 schema / enum / validator / retry |
| 代码片段生成 | 中等可靠 | 范围小，有单元测试和 lint |
| 多文件代码修改 | 条件可靠 | 需要 repo map、搜索、编辑工具、测试和 diff review |
| 基于证据的问答 | 条件可靠 | 需要 RAG、引用、证据支持率评估 |
| 规划和头脑风暴 | 有价值但不可直接执行 | 需要人审、约束和后续验证 |

注意这里说的是“工程化难度”，不是模型天然能力排名。封闭标签分类看起来简单，但如果标签定义含糊、训练数据偏斜，效果也会很差；多文件代码修改看起来难，但如果上下文、工具和测试都设计得好，也可以稳定落地。

**实际例子：客户反馈分类**

```text
输入:
"订单号 12345 还没发货，已经等了 3 天了"

期望输出:
{
  "category": "物流",
  "sentiment": "不满",
  "urgency": "中",
  "order_id": "12345"
}
```

这个任务适合 LLM，不是因为“LLM 分类准确率天然很高”，而是因为它满足几个条件：

- 标签集有限；
- 输入短；
- 语义线索明显；
- 输出可用 JSON Schema 约束；
- 可以积累人工标注集做离线评测；
- 错误成本通常可由人工复核兜底。

### 从失败机制理解 LLM 不擅长什么

LLM 不擅长的任务，不是简单的“模型不够聪明”，而是任务需要模型没有内置的能力：精确计算、实时状态、长期状态、可靠记忆、外部事实验证和确定性执行。

| 任务类型 | 失败机制 | 工程补救 |
|:---|:---|:---|
| 精确计算 | 语言模型不是符号计算器 | 调用计算器、SQL、Python、规则引擎 |
| 实时事实 | 训练数据和当前世界不同步 | 检索、数据库、Web、业务 API |
| 长期记忆 | 单次调用无状态 | 外部 Memory、Profile、Session Store |
| 多步推理 | 中间步骤容易漂移 | 分解任务、显式状态机、工具验证 |
| 自我检查 | 模型可能解释自己的错误 | 独立 verifier、测试、规则检查 |
| 高风险决策 | 错误代价高且不可逆 | 人工审批、权限系统、审计 |
| 开放领域事实问答 | 可能编造引用和细节 | RAG、引用、claim verification |

**实际例子：复利计算**

```text
任务:
本金 10000 元，年利率 5%，按年复利 20 年后是多少？

不推荐:
让 LLM 直接给最终数字。

推荐:
让 LLM 识别任务类型和公式，再调用计算工具：
FV = 10000 * (1 + 0.05) ** 20
```

这里的核心不是“LLM 算错了多少”，而是**架构上不应该让概率模型承担确定性计算职责**。模型负责理解问题、选择公式、解释结果；计算交给确定性工具。

### Benchmark 只能回答局部问题

公开 benchmark 很有价值，但不能直接等同于你的产品效果。

| Benchmark 类型 | 能说明什么 | 不能说明什么 |
|:---|:---|:---|
| MMLU / 考试类 | 学科知识和选择题能力 | 真实业务输入、工具使用、长链路稳定性 |
| HumanEval / 代码题 | 小函数生成能力 | 多文件修改、依赖环境、真实测试修复 |
| SWE-bench | 真实仓库 issue 修复能力 | 你的代码库、你的工具权限、你的 review 流程 |
| TruthfulQA | 对常见误解的抗幻觉能力 | RAG 场景、企业私有知识、引用质量 |
| HELM | 多场景、多指标权衡 | 某个团队的具体任务 ROI |

HELM 的重要启发是：评估不能只看 accuracy，还要看 calibration、robustness、fairness、toxicity、efficiency 等维度。OpenAI 的 GPT-4 Technical Report 也明确指出，即使能力强的模型仍会产生事实幻觉和推理错误，在高风险场景需要人审、额外上下文或避免使用。

### 主流 LLM 榜单该怎么看

截至 2026-05-08，主流 LLM 榜单已经从“谁的总分最高”演化成了多个侧重点不同的评测仪表盘。榜单应该用来发现候选模型、理解能力分布和设计本地 eval，而不应该直接替代你的选型实验。

| 榜单 / Benchmark | 主要信号 | 适合回答的问题 | 常见误读 |
|:---|:---|:---|:---|
| [LMArena / Arena](https://arena.ai/leaderboard) | 人类偏好、盲测对比、聊天 / 代码 / 视觉 / 文档等分榜 | 用户主观体验上，哪些模型更容易给出“看起来更好”的答案？ | 把偏好胜率当成事实正确率或生产稳定性 |
| [LiveBench](https://github.com/livebench/livebench) | 动态更新题目，降低测试集污染，覆盖数学、代码、推理、数据分析、语言和指令跟随 | 哪些模型在较新的综合任务上仍有区分度？ | 忽略你的业务数据、工具链和延迟成本 |
| [HELM](https://crfm.stanford.edu/helm/latest/) | 多场景、多指标评测，强调 accuracy 之外的 calibration、robustness、fairness、toxicity、efficiency | 选型时是否只盯住准确率，漏掉安全性、鲁棒性和成本指标？ | 把研究型综合评估直接当成单一产品排名 |
| [Artificial Analysis](https://artificialanalysis.ai/) | 模型质量、输出速度、上下文、价格、提供商能力等工程维度 | 同一任务下，质量、速度和成本如何权衡？ | 忘记价格、限流、缓存策略会随官方 API 政策变化 |
| [SWE-bench](https://www.swebench.com/) | 真实 GitHub issue 修复，关注代码理解、修改和测试修复 | Coding Agent 是否能处理真实仓库中的缺陷修复？ | 忽略 harness、工具权限、上下文压缩和测试环境差异 |
| [Aider Polyglot](https://aider.chat/docs/leaderboards/) | 多语言代码编辑能力，关注模型按指令修改代码的成功率和编辑格式 | 模型是否适合在工程循环中做局部代码编辑？ | 把单工具、单工作流结果泛化到所有 IDE / CLI Agent |
| [τ-bench](https://www.tau-bench.com/) | 多轮对话、工具调用、业务策略遵循和最终状态一致性 | Agent 能否在客服、订单、航旅等规则场景中稳定执行？ | 只看 pass rate，不看多次运行一致性和策略违规类型 |

一个更实用的读法是按任务选择榜单，而不是按总榜选择模型：

```text
通用问答 / 写作体验      -> 先看 LMArena，再做人工偏好评测
复杂推理 / 新题泛化      -> 先看 LiveBench，再做领域题集 eval
代码修复 / Coding Agent  -> 先看 SWE-bench / Aider，再跑你的仓库任务集
工具调用 / 业务流程 Agent -> 先看 τ-bench，再跑端到端状态校验
企业选型 / 成本治理      -> 先看 Artificial Analysis，再核对官方 pricing 和限流
安全 / 鲁棒性 / 治理     -> 先看 HELM，再补充红队测试和合规评估
```

如果不同榜单给出不同结论，这通常不是“榜单冲突”，而是它们测量的能力不同。一个模型可能在 LMArena 的主观偏好上领先，但在 SWE-bench 的真实代码修复中一般；也可能在 LiveBench 的数学推理上很强，却不适合低延迟客服场景。

因此，模型选型可以分成三步：

1. **榜单筛选**：用公开榜单确定 3-5 个候选模型，记录榜单日期、版本、评价维度和任务分布。
2. **本地复现**：用你的真实样本、prompt、工具、schema、延迟预算和成本约束跑离线 eval。
3. **线上灰度**：在可回滚的流量中监控质量、人工接管率、错误类型、token 成本和用户满意度。

真正可靠的模型选择报告，应该同时包含：

```text
候选模型来源：哪些榜单、什么日期、什么维度
本地评测集：多少样本、什么分布、如何标注
运行配置：prompt 版本、工具权限、temperature、schema、retry 策略
质量指标：accuracy / F1 / pass rate / evidence support / human preference
工程指标：p50 / p95 延迟、成本、限流、缓存命中率、失败率
风险指标：幻觉类型、越权工具调用、拒答率、人工兜底比例
结论边界：结论只对哪些任务、语言、输入分布和系统版本生效
```

榜单给你“从哪里开始看”，本地 eval 才能回答“这个模型能不能用于我的系统”。

### 能力边界总结

```text
LLM 的本质：
┌─────────────────────────────────────────────┐
│  条件概率生成器                              │
│  在当前上下文下生成最可能、最符合指令的输出  │
└─────────────────────────────────────────────┘

工程推论：
1. 语言理解、改写、抽取、规划适合交给模型；
2. 计算、检索、执行、持久化、权限要交给系统；
3. 质量不能靠“模型感觉不错”，必须靠 eval 和 trace；
4. 模型能力结论必须绑定任务、数据、版本、prompt 和指标。
```

---

## 7.2 幻觉问题与应对策略

### 什么是幻觉？

**定义**：LLM 生成看似合理但实际错误的内容。

**典型案例：**

```text
问题: "2022年诺贝尔物理学奖得主是谁？"

LLM 幻觉回答:
"2022年诺贝尔物理学奖授予了 John Doe 和 Jane Smith，
表彰他们在量子纠缠领域的贡献。"

问题：
1. 名字是编造的
2. 研究方向是猜测的
3. 表述非常自信

正确答案:
Alain Aspect, John Clauser, Anton Zeilinger
（量子信息科学的奠基实验）
```

### 幻觉的类型

**1. 事实性幻觉（Factual Hallucination）**

```text
输入: "介绍一下 TensorFlow 2.0 的新特性"
幻觉: "TensorFlow 2.0 引入了自动微分功能"
事实: TensorFlow 1.x 就有自动微分
```

**2. 逻辑性幻觉（Logical Hallucination）**

```text
输入: "如果 A > B 且 B > C，那么 A 和 C 的关系？"
幻觉: "无法确定 A 和 C 的关系"
事实: 必然 A > C（传递性）
```

**3. 引用性幻觉（Citation Hallucination）**

```text
输入: "引用一篇关于 Transformer 的论文"
幻觉: "根据 Smith et al. (2023) 的研究..."
事实: 这篇论文不存在
```

### 应对策略

**策略 1：工具调用（Tool Use）**

```python
# ❌ 直接让 LLM 计算
prompt = "计算 sin(45°) × cos(30°)"
result = llm.generate(prompt)  # 不可靠

# ✅ 调用计算工具
prompt = "生成 Python 代码计算 sin(45°) × cos(30°)"
code = llm.generate(prompt)
result = execute_code(code)  # 可靠
```

**策略 2：检索增强（RAG）**

```python
# ❌ 直接询问 LLM
answer = llm.generate("2022年诺贝尔物理学奖得主？")

# ✅ 先检索再回答
docs = search_wikipedia("2022 Nobel Prize Physics")
answer = llm.generate(f"基于以下资料回答：\n{docs}\n问题：...")
```

**策略 3：Self-Consistency（自我一致性）**

```python
# 多次采样，选择一致的答案
answers = []
for _ in range(5):
    answer = llm.generate(question, temperature=0.7)
    answers.append(answer)

# 投票选择最一致的答案
final_answer = most_common(answers)
```

**策略 4：Chain-of-Thought 验证**

```python
prompt = """
问题：{question}

请分步推理：
1. 列出已知条件
2. 列出推理步骤
3. 给出最终答案
4. 验证答案是否合理

如果发现矛盾，请指出并重新推理。
"""
```

**策略 5：External Verification（外部验证）**

```python
class VerifiedAnswer:
    def answer(self, question: str):
        # 1. LLM 生成答案
        answer = self.llm.generate(question)

        # 2. 提取可验证的事实
        claims = self.extract_claims(answer)

        # 3. 外部验证
        for claim in claims:
            if not self.verify_claim(claim):
                # 标记不可靠
                answer = self.add_warning(answer, claim)

        return answer

    def verify_claim(self, claim: str) -> bool:
        # 通过搜索引擎、数据库等验证
        search_results = search(claim)
        return check_consistency(claim, search_results)
```

---

## 7.3 Prompt Engineering 核心原则

### 原则 1：明确性（Clarity）

**❌ 模糊的 Prompt：**
```text
"帮我写个函数"
```

**✅ 明确的 Prompt：**
```text
请用 Python 写一个函数，功能如下：
- 函数名：calculate_discount
- 输入参数：
  - price: float (原价)
  - discount_rate: float (折扣率，0-1之间)
- 返回：float (折后价)
- 要求：
  - 参数验证（价格非负，折扣率在0-1之间）
  - 保留两位小数
  - 添加 docstring
```

**为什么更稳定：**

- 输出边界从“随便写一个函数”变成了明确接口；
- 参数验证、返回格式和文档要求都可检查；
- 后续可以用单元测试验证，而不是靠读者感觉判断质量。

### 原则 2：结构化（Structure）

**❌ 无结构：**
```text
我想知道 Transformer 的工作原理以及它和 RNN 的区别还有它的优缺点
```

**✅ 结构化：**
```text
关于 Transformer 架构，请回答以下问题：

## 1. 工作原理
- Self-Attention 机制如何工作？
- Positional Encoding 的作用是什么？

## 2. 与 RNN 对比
- 主要区别是什么？
- 各自的优势场景？

## 3. 优缺点
- 优点（至少3个）
- 缺点（至少2个）

请用 Markdown 格式回答。
```

### 原则 3：示例驱动（Few-Shot Learning）

**Zero-Shot（无示例）：**
```text
将以下客户反馈分类：
"产品质量不错，但物流太慢了"
```

**Few-Shot（有示例）：**
```text
将客户反馈分类为：物流、产品质量、客服、价格

示例 1:
输入: "订单还没发货，已经等了5天"
分类: 物流

示例 2:
输入: "产品做工粗糙，不值这个价"
分类: 产品质量

示例 3:
输入: "客服态度很好，帮我解决了问题"
分类: 客服

现在分类：
输入: "产品质量不错，但物流太慢了"
分类: ?
```

**为什么示例有效：**

- 示例把标签边界变成了可模仿模式；
- 模型可以学习“物流”和“产品质量”同时出现时如何取舍；
- 但效果提升必须用你的标注集验证，不能把 few-shot 当成固定收益。

### 原则 4：约束条件（Constraints）

**❌ 无约束：**
```text
生成一篇关于 AI 的文章
```

**✅ 有约束：**
```text
生成一篇关于 AI 的技术文章，要求：

格式约束：
- 字数：800-1000字
- 结构：引言 + 3个小节 + 总结
- 使用 Markdown 格式

内容约束：
- 目标读者：有编程基础的工程师
- 深度：中级（不要太基础，不要太学术）
- 必须包含：实际代码示例

风格约束：
- 语言：中文
- 风格：技术准确，表达简洁
- 避免：营销话术、夸大其词
```

### 原则 5：输出格式（Output Format）

**❌ 自由格式：**
```text
提取这段文本中的关键信息
```

**✅ 指定格式：**
```text
从以下文本提取关键信息，返回 JSON 格式：

{
  "name": "人名",
  "email": "邮箱",
  "phone": "电话",
  "company": "公司名称"
}

文本：...
```

### Prompt 模板库

```python
# 模板 1：任务分解
TASK_DECOMPOSITION_TEMPLATE = """
任务：{task}

请将此任务分解为可执行的子任务：

1. 子任务 1
   - 输入：...
   - 输出：...
   - 验证标准：...

2. 子任务 2
   ...

最终输出：...
"""

# 模板 2：错误处理
ERROR_HANDLING_TEMPLATE = """
执行任务时发生错误：

任务：{task}
错误信息：{error}

请分析：
1. 错误原因是什么？
2. 如何修复？
3. 给出修复后的代码/方案

不要重复之前的错误。
"""

# 模板 3：Self-Critique（自我批评）
SELF_CRITIQUE_TEMPLATE = """
你刚才给出的答案是：
{previous_answer}

请批判性地审查这个答案：
1. 是否有事实错误？
2. 逻辑是否严密？
3. 是否遗漏重要信息？
4. 是否有更好的表达方式？

如果有问题，请给出改进后的答案。
"""
```

---

## 7.4 模型选择与权衡

### 不要把模型选择写成静态排行榜

模型名称、上下文长度、价格、工具调用能力和安全策略都会变化。书稿正文不适合维护一张“2026 年主流模型排行榜”。更稳妥的方式，是把模型选择写成一组工程维度，然后在项目里用小型 eval 选择。

| 维度 | 要问的问题 | 典型评测 |
|:---|:---|:---|
| 推理能力 | 能否处理多约束、多步骤任务？ | 任务集成功率、人工评分、错误类型 |
| 代码能力 | 能否读代码、改代码、修测试？ | 单元测试通过率、diff 质量、review 缺陷率 |
| 工具调用 | 能否生成正确参数并根据结果继续？ | tool call success rate、重试率、无效调用率 |
| 结构化输出 | 能否稳定遵守 schema？ | JSON parse rate、schema validation rate |
| 长上下文 | 长文档下是否还能抓住关键事实？ | evidence recall、引用支持率、遗漏率 |
| 多模态 | 是否需要图像、表格、截图、PDF？ | modality-specific eval |
| 延迟 | 是否满足交互体验？ | p50 / p95 latency |
| 成本 | 单任务成本是否可接受？ | cost per successful task |
| 数据治理 | 数据能否出域？是否支持私有部署？ | 安全审查、合规审查、审计能力 |

### 选择决策树

```text
是否需要本地部署？
├─ 是 → 评估开源模型、私有推理服务、硬件和数据治理
└─ 否 ↓

是否需要多模态？
├─ 是 → 选择支持目标模态的模型，并做 modality-specific eval
└─ 否 ↓

是否需要长上下文？
├─ 是 → 同时评估上下文窗口、有效召回、延迟和成本
└─ 否 ↓

任务是否需要工具调用？
├─ 是 → 优先评估 tool schema adherence 和 recovery 能力
└─ 否 ↓

任务风险等级？
├─ 高 → 强模型 + verifier + 人审 + 审计
├─ 中 → 强模型或中等模型 + 局部验证
└─ 低 → 快模型 / 便宜模型 + 抽样质检
```

### 成本优化策略

**策略 1：模型分层（Model Tiering）**

```python
class AdaptiveModelRouter:
    """根据任务复杂度选择模型"""

    def __init__(self, models):
        self.models = models

    def route(self, task: str):
        complexity = self.classify_complexity(task)
        return self.models[complexity]

    def classify_complexity(self, task: str) -> str:
        """用便宜的模型分类任务复杂度"""
        prompt = f"评估任务复杂度（simple/medium/complex）：{task}"
        result = self.models["simple"].generate(prompt)
        return result
```

**策略 2：Prompt 缓存（Prompt Caching）**

```python
# 如果模型供应商支持 prompt caching，可以缓存稳定上下文：
# - 长 System Prompt
# - 工具说明
# - 稳定知识库摘要
# 具体缓存语义和价格以官方文档为准。

response = anthropic.messages.create(
    model=MODEL_NAME,
    system=[
        {
            "type": "text",
            "text": LONG_SYSTEM_PROMPT,  # 缓存这部分
            "cache_control": {"type": "ephemeral"}
        }
    ],
    messages=[{"role": "user", "content": user_input}]
)
```

**策略 3：批量处理（Batch Processing）**

```python
# 批量处理适用于非实时任务：
# - 离线分类
# - 离线摘要
# - eval case 批量跑分
# 具体价格折扣和完成时限以供应商官方文档为准。

batch_jobs = [
    {"custom_id": "task-1", "method": "POST", "url": "/v1/chat/completions", ...},
    {"custom_id": "task-2", ...},
    ...
]

# 提交批量任务
batch = client.batches.create(
    input_file_id=upload_batch_file(batch_jobs),
    endpoint="/v1/chat/completions",
    completion_window="24h"
)
```

---

## 7.5 上下文管理

### Token 限制

上下文窗口和价格是模型 snapshot 的属性，不适合在正文里写死。选型时应该查官方模型卡和 pricing，并同时关注下面几个指标。

| 指标 | 含义 | 为什么重要 |
|:---|:---|:---|
| Context Window | 最大输入上下文长度 | 决定一次调用理论上能放多少材料 |
| Effective Context | 长上下文中真正可利用的信息量 | 窗口大不等于能稳定用好全部上下文 |
| Max Output | 单次最大输出长度 | 影响长报告、代码生成、批处理 |
| Latency | p50 / p95 响应时间 | 影响交互体验和任务超时 |
| Pricing | 输入、输出、缓存、批处理价格 | 影响单任务成本和规模化成本 |
| Cache Semantics | 哪些上下文可缓存、缓存多久 | 影响长 system prompt 和工具说明成本 |

### 上下文溢出问题

**问题场景：**

```python
# Agent 循环执行多次工具调用
context = ""
for i in range(10):
    context += f"Step {i}: {tool_result}\n"  # 累积上下文
    response = llm.generate(context)

# 问题：
# - 第10次迭代时，context 可能超过 token 限制
# - 早期步骤可能不再相关，但仍占用 token
```

### 解决策略

**策略 1：滑动窗口（Sliding Window）**

```python
class SlidingWindowMemory:
    """保留最近 N 条消息"""

    def __init__(self, max_messages: int = 10):
        self.messages = []
        self.max_messages = max_messages

    def add(self, message: dict):
        self.messages.append(message)
        if len(self.messages) > self.max_messages:
            # 保留 system message + 最近的消息
            system_msg = self.messages[0]  # 假设第一条是 system
            self.messages = [system_msg] + self.messages[-self.max_messages+1:]

    def get_context(self):
        return self.messages
```

**策略 2：摘要压缩（Summarization）**

```python
class SummarizingMemory:
    """定期压缩历史对话"""

    def __init__(self, llm, max_tokens: int = 4000):
        self.llm = llm
        self.messages = []
        self.max_tokens = max_tokens

    def add(self, message: dict):
        self.messages.append(message)

        # 检查 token 数量
        if self.estimate_tokens() > self.max_tokens:
            self.compress()

    def compress(self):
        """压缩旧消息"""
        # 保留最近 5 条完整消息
        recent = self.messages[-5:]

        # 压缩更早的消息
        old = self.messages[:-5]
        summary = self.llm.generate(
            f"总结以下对话，保留关键信息：\n{old}"
        )

        # 用摘要替换旧消息
        self.messages = [
            {"role": "system", "content": f"之前的对话摘要：{summary}"}
        ] + recent
```

**策略 3：相关性过滤（Relevance Filtering）**

```python
class RelevanceFilteredMemory:
    """只保留与当前问题相关的历史"""

    def get_relevant_context(self, current_query: str, history: List):
        """检索相关的历史消息"""
        relevant = []

        for msg in history:
            relevance_score = self.calculate_relevance(current_query, msg)
            if relevance_score > 0.7:
                relevant.append(msg)

        return relevant

    def calculate_relevance(self, query: str, message: dict) -> float:
        """计算相关性（简化版，实际可用 embedding）"""
        # 使用 LLM 评估相关性
        prompt = f"""
        问题：{query}
        历史消息：{message['content']}

        这条历史消息与当前问题的相关性（0-1）？
        只返回数字。
        """
        score = float(self.llm.generate(prompt))
        return score
```

---

## 7.6 质量保证与测试

### LLM 系统的测试策略

**1. 单元测试（固定输入输出）**

```python
def test_sentiment_analysis():
    """测试情感分析功能"""

    test_cases = [
        {
            "input": "这个产品太棒了！",
            "expected": "positive"
        },
        {
            "input": "质量很差，非常失望",
            "expected": "negative"
        },
        {
            "input": "还可以吧",
            "expected": "neutral"
        }
    ]

    for case in test_cases:
        result = sentiment_agent.analyze(case["input"])
        assert result == case["expected"], \
            f"Failed: {case['input']} -> {result} (expected {case['expected']})"
```

**2. 基于 LLM 的评估（Evaluation with LLM）**

```python
class LLMEvaluator:
    """用 LLM 评估 LLM 输出"""

    def evaluate_answer(self, question: str, answer: str, reference: str) -> dict:
        """评估答案质量"""

        prompt = f"""
        评估以下答案的质量：

        问题：{question}
        参考答案：{reference}
        待评估答案：{answer}

        评分标准（1-5分）：
        1. 准确性：答案是否正确？
        2. 完整性：是否覆盖所有要点？
        3. 简洁性：表达是否简洁清晰？

        返回 JSON：
        {{
          "accuracy": 1-5,
          "completeness": 1-5,
          "conciseness": 1-5,
          "overall": 1-5,
          "feedback": "具体反馈"
        }}
        """

        result = self.llm.generate(prompt)
        return parse_json(result)
```

**3. A/B 测试（在线评估）**

```python
class ABTestingFramework:
    """A/B 测试框架"""

    def __init__(self):
        self.model_a = GPT4()
        self.model_b = Claude35()
        self.results = []

    def route_request(self, user_id: int, query: str):
        """随机分配用户到不同模型"""

        if hash(user_id) % 2 == 0:
            model, variant = self.model_a, "A"
        else:
            model, variant = self.model_b, "B"

        start_time = time.time()
        response = model.generate(query)
        latency = time.time() - start_time

        # 记录结果
        self.results.append({
            "variant": variant,
            "latency": latency,
            "response": response,
            "user_id": user_id
        })

        return response

    def analyze_results(self):
        """分析 A/B 测试结果"""
        a_results = [r for r in self.results if r["variant"] == "A"]
        b_results = [r for r in self.results if r["variant"] == "B"]

        return {
            "A": {
                "avg_latency": np.mean([r["latency"] for r in a_results]),
                "count": len(a_results)
            },
            "B": {
                "avg_latency": np.mean([r["latency"] for r in b_results]),
                "count": len(b_results)
            }
        }
```

---

## 7.7 生产环境最佳实践

### 1. 错误处理

```python
class RobustLLMClient:
    """健壮的 LLM 客户端"""

    def __init__(self, llm, max_retries: int = 3):
        self.llm = llm
        self.max_retries = max_retries

    async def generate(self, prompt: str, **kwargs):
        """带重试的生成"""

        for attempt in range(self.max_retries):
            try:
                response = await self.llm.generate(prompt, **kwargs)
                return response

            except RateLimitError as e:
                # 速率限制：指数退避
                wait_time = 2 ** attempt
                logger.warning(f"Rate limited, retry in {wait_time}s")
                await asyncio.sleep(wait_time)

            except TimeoutError as e:
                # 超时：重试
                logger.warning(f"Timeout on attempt {attempt+1}")
                if attempt == self.max_retries - 1:
                    raise

            except InvalidRequestError as e:
                # 无效请求：不重试
                logger.error(f"Invalid request: {e}")
                raise

        raise Exception(f"Failed after {self.max_retries} retries")
```

### 2. 监控与日志

```python
class MonitoredLLMClient:
    """带监控的 LLM 客户端"""

    async def generate(self, prompt: str, **kwargs):
        start_time = time.time()

        try:
            response = await self.llm.generate(prompt, **kwargs)

            # 记录成功指标
            self.metrics.record({
                "latency": time.time() - start_time,
                "input_tokens": self.count_tokens(prompt),
                "output_tokens": self.count_tokens(response),
                "model": self.llm.model_name,
                "status": "success"
            })

            return response

        except Exception as e:
            # 记录失败
            self.metrics.record({
                "latency": time.time() - start_time,
                "model": self.llm.model_name,
                "status": "error",
                "error_type": type(e).__name__
            })
            raise
```

### 3. 成本控制

```python
class CostControlledClient:
    """成本控制的 LLM 客户端"""

    def __init__(self, llm, budget_per_day: float):
        self.llm = llm
        self.budget_per_day = budget_per_day
        self.today_cost = 0
        self.last_reset = date.today()

    async def generate(self, prompt: str, **kwargs):
        # 检查预算
        self.check_budget()

        # 估算成本
        estimated_cost = self.estimate_cost(prompt, kwargs.get("max_tokens", 1000))

        if self.today_cost + estimated_cost > self.budget_per_day:
            raise BudgetExceededError(
                f"Daily budget ${self.budget_per_day} exceeded"
            )

        # 生成
        response = await self.llm.generate(prompt, **kwargs)

        # 更新成本
        actual_cost = self.calculate_cost(prompt, response)
        self.today_cost += actual_cost

        return response

    def check_budget(self):
        """重置每日预算"""
        if date.today() > self.last_reset:
            self.today_cost = 0
            self.last_reset = date.today()
```

---

## 本章小结

### 核心要点回顾

**1. LLM 能力边界**
- 不存在脱离模型版本、任务分布、prompt 和评测集的通用准确率
- 擅长：语义理解、改写、抽取、受约束生成、规划草案
- 不擅长：精确计算、实时事实、长期状态、不可逆高风险行动
- 核心：条件概率生成，需要工具、检索、状态和验证系统配合

**2. 幻觉问题**
- 类型：事实性、逻辑性、引用性幻觉
- 应对：工具调用、RAG、Self-Consistency、外部验证

**3. Prompt Engineering**
- 明确性：详细的任务描述和要求
- 结构化：清晰的输入输出格式
- 示例驱动：Few-Shot Learning 明确标签边界和输出风格
- 约束条件：格式、内容、风格的明确要求

**4. 模型选择**
- 不维护静态模型排行榜，用任务 eval 做选择
- 根据推理、代码、工具调用、结构化输出、长上下文、延迟、成本和数据治理综合取舍
- 通过模型分层、缓存和批处理降低规模化成本

**5. 上下文管理**
- 滑动窗口：保留最近消息
- 摘要压缩：压缩历史对话
- 相关性过滤：只保留相关信息

**6. 质量保证**
- 单元测试：固定输入输出
- LLM 评估：用 LLM 评估 LLM
- A/B 测试：在线对比不同模型

**7. 生产最佳实践**
- 错误处理：重试、退避、降级
- 监控日志：性能、成本、错误追踪
- 成本控制：预算管理、成本估算

### 关键洞察

> **成功的 Agent 系统建立在对 LLM 能力边界的深刻理解之上。不是让 LLM 做所有事情，而是让它做它擅长的事情，其余交给传统工程方法。**

> **任何能力数字都必须回答四个问题：在哪个模型版本上、在哪批数据上、用哪个 prompt / harness、按什么指标评估。答不出这四点，数字就只能算印象，不是工程证据。**

### 下一章预告

下一章我们将进入 **Prompt Engineering 与结构化输出**：学习如何把人的意图整理成模型可执行、系统可验证的任务协议。

---

## 参考资料

1. [GPT-4 Technical Report](https://arxiv.org/abs/2303.08774) - OpenAI, 2023
2. [Holistic Evaluation of Language Models](https://crfm.stanford.edu/helm/latest/) - Stanford CRFM, 2023
3. [TruthfulQA: Measuring How Models Mimic Human Falsehoods](https://arxiv.org/abs/2109.07958) - Lin et al., 2021
4. [SWE-bench: Can Language Models Resolve Real-World GitHub Issues?](https://www.swebench.com/) - Jimenez et al., 2023
5. [Language Models are Few-Shot Learners](https://arxiv.org/abs/2005.14165) - Brown et al., 2020
6. [Chain-of-Thought Prompting Elicits Reasoning in Large Language Models](https://arxiv.org/abs/2201.11903) - Wei et al., 2022
7. [Retrieval-Augmented Generation for Knowledge-Intensive NLP Tasks](https://arxiv.org/abs/2005.11401) - Lewis et al., 2020
8. [Introducing Structured Outputs in the API](https://openai.com/index/introducing-structured-outputs-in-the-api/) - OpenAI, 2024
9. [LMArena / Arena Leaderboard](https://arena.ai/leaderboard) - LMArena, dynamic leaderboard
10. [LiveBench](https://github.com/livebench/livebench) - contamination-aware dynamic LLM benchmark
11. [Artificial Analysis](https://artificialanalysis.ai/) - model quality, speed and cost analysis
12. [Aider LLM Leaderboards](https://aider.chat/docs/leaderboards/) - coding and code editing benchmark
13. [τ-bench: Tool-Agent-User Interaction Benchmark](https://www.tau-bench.com/) - Yao et al., 2024
14. [Attention Is All You Need](https://arxiv.org/abs/1706.03762) - Vaswani et al., 2017
15. [Scaling Laws for Neural Language Models](https://arxiv.org/abs/2001.08361) - Kaplan et al., 2020
16. [Training Compute-Optimal Large Language Models](https://arxiv.org/abs/2203.15556) - Hoffmann et al., 2022
17. [Training Language Models to Follow Instructions with Human Feedback](https://arxiv.org/abs/2203.02155) - Ouyang et al., 2022
18. [Direct Preference Optimization: Your Language Model is Secretly a Reward Model](https://arxiv.org/abs/2305.18290) - Rafailov et al., 2023
19. [ReAct: Synergizing Reasoning and Acting in Language Models](https://arxiv.org/abs/2210.03629) - Yao et al., 2022
20. [Toolformer: Language Models Can Teach Themselves to Use Tools](https://arxiv.org/abs/2302.04761) - Schick et al., 2023
21. [LoRA: Low-Rank Adaptation of Large Language Models](https://arxiv.org/abs/2106.09685) - Hu et al., 2021
22. [QLoRA: Efficient Finetuning of Quantized LLMs](https://arxiv.org/abs/2305.14314) - Dettmers et al., 2023
23. [LLaMA: Open and Efficient Foundation Language Models](https://arxiv.org/abs/2302.13971) - Touvron et al., 2023
