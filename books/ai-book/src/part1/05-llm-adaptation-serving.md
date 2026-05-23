# 第5章 微调、量化与部署：LoRA、QLoRA、Serving Engine

企业真正落地大模型时，常见问题不是“怎么训练一个 GPT”，而是：已有模型怎么适配业务，怎么降低成本，怎么稳定部署，怎么在质量、延迟和显存之间取舍。

这一章把三件事放在一起讲：

1. **微调**：让模型更稳定地表现出某类行为模式。
2. **量化**：用更低精度降低推理和训练成本。
3. **Serving**：把模型放进可监控、可灰度、可回滚的在线服务体系。

它们在工程上不能分开看。一个 LoRA adapter 是否要动态加载，会影响 serving 调度；一个 INT4 权重量化是否可用，要看业务 eval；一个微调模型是否值得上线，不只取决于 loss，还取决于数据、版本、权限、回滚和线上监控。

## 5.1 宏观理解：适配模型的几条路

面对一个业务需求，通常有几种手段：

```text
Prompt
  -> Structured Output
  -> RAG / Tool
  -> Workflow / Guardrails / Evals
  -> SFT / LoRA / QLoRA
  -> Preference Optimization
  -> Full Fine-tuning / Continued Pretraining
```

越往右，成本越高、风险越大、对数据和评估要求越高。

不要把微调当成默认选项。很多问题通过 Prompt、RAG、工具、结构化输出和 eval 就能解决。微调最适合的不是“让模型知道更多事实”，而是把一类稳定任务的行为模式固化下来。

可以用一句话区分：

```text
Prompt 负责告诉模型“这次想让你怎么做”
RAG 负责告诉模型“这次应该参考哪些知识”
Tools 负责让模型“这次可以调用哪些外部能力”
Fine-tuning 负责让模型“以后更自然、更稳定地按这种模式去做”
```

## 5.2 微调到底在改变什么

模型微调（fine-tuning）本质上是在已有基础模型之上，使用一组目标任务样本继续训练，让模型更稳定地表现出某种特定能力模式。

“能力模式”比“知识”更准确。微调通常更擅长改变：

- 任务启动方式：模型是否更快进入正确任务模式。
- 注意力偏好：面对输入时更优先关注哪些字段。
- 输出先验：更倾向产出什么格式、语气和结构。
- 风险偏好：是否更保守、更愿意澄清或拒答。
- 标签边界：分类、路由、审核等任务的分界是否更稳定。
- 领域表达：是否更贴近业务术语和团队写作习惯。

微调不适合可靠管理大量动态事实。最新 owner、当前日志、实时指标、审批状态、库存价格、权限结果，都应该放在数据库、搜索索引、工具或 RAG 证据里，而不是希望模型“记住”。

## 5.3 微调适合什么，不适合什么

一个任务适合微调，通常同时满足这些条件：

- 高频重复。
- 输入分布相对稳定。
- 输出协议明确。
- 可以定义对错或优劣。
- 已经积累了人工修正样本。
- 错误主要来自行为模式不稳定，而不是事实缺失。
- 可以通过 Shadow Mode 或建议层安全上线。

适合微调的典型任务包括：

- 固定格式输出。
- 多标签或单标签分类。
- 结构化摘要。
- 领域风格统一。
- 工单归类与路由。
- 安全拒答或澄清格式。
- 告警摘要初稿生成。
- 值班交接文档标准化。

不适合优先靠微调解决的问题包括：

- 高频变化知识。
- 强实时事实。
- 明明需要查工具，却希望模型脑补。
- 任务边界混乱，把检索、审批、执行和解释揉在一起。
- 极高风险决策，例如资损、账务、权限、价格、数据修正。
- 必须严格带引用、证据 ID 和时间窗口的最终结论。

一个简单判断公式是：

```text
稳定任务协议 + 高质量样本 + 可评分 eval + 可回滚发布
  -> 可以认真评估微调

动态事实 + 权限决策 + 证据链要求 + 高风险执行
  -> 优先 RAG / Tool / Workflow / Human Approval
```

## 5.4 微调、SFT、偏好优化与强化式优化

真实工程里至少要区分三类路线：

| 路线 | 数据形态 | 适合问题 | 主要风险 |
|:---|:---|:---|:---|
| SFT | 输入到标准输出 | 格式、分类、摘要、模板化回答 | 模仿噪声、过拟合格式 |
| 偏好优化 | 同一输入下 chosen / rejected | 多个答案都可行，但有偏好差异 | 偏好标签不一致、奖励黑客 |
| 强化式优化 | 状态、动作、奖励或轨迹 | 多步工具、长链路任务、环境反馈 | 工程复杂、reward 难定义、安全风险高 |

### 5.4.1 SFT：最常见的第一步

SFT（Supervised Fine-Tuning）的基本形式是：

```text
输入 -> 标准输出
```

它适合让模型学习：

- 看到某类输入应该输出哪些字段。
- 语气和格式应该如何统一。
- 哪些情况下应该保守表达。
- 某类标签的判断边界。

SFT 是企业里最常见的第一步，因为数据形态直观，离线评测容易做，也容易和人工修正轨迹结合。

### 5.4.2 偏好优化：学习“哪个答案更好”

很多任务不是只有一个标准答案，而是多个答案都正确，但团队更偏好其中一种。例如：

- 两份 Incident Summary 都正确，但一份更简洁。
- 两个建议都安全，但一份更保守。
- 两个客服回答都合规，但一份更自然。
- 两份 Runbook 摘要都没错，但一份证据引用更清楚。

这类任务可以用偏好优化表达：

```text
同一个输入
  -> 回答 A
  -> 回答 B
  -> 我们更偏好 A
```

DPO 这类方法降低了传统 RLHF 的工程复杂度，但它仍然依赖高质量偏好数据。偏好数据如果口径混乱，模型学到的不是“更好”，而是标注人的随机偏好。

### 5.4.3 强化式优化：不要第一天就上

强化式优化更适合多步任务，比如：

- 多步工具调用。
- 浏览器或 CLI 任务执行。
- 需要探索、试错和回退的长链路工作流。
- 有清晰环境反馈的自动化任务。

但大多数企业内部 Agent 团队第一阶段不应该直接从这里开始。它要求可靠 reward、轨迹回放、安全护栏、回归验证和更强的平台能力。没有这些基础，强化式优化很容易把系统风险放大。

## 5.5 Full Fine-tuning

Full fine-tuning 更新模型全部参数。

优点：

- 适配能力强。
- 可以改变模型深层行为。
- 适合大规模、高价值、稳定任务。
- 当基础模型与领域分布差距很大时，可能比 PEFT 更有效。

缺点：

- 显存和计算成本高。
- 容易灾难性遗忘。
- 需要高质量数据和严格 eval。
- 部署、版本管理和回滚成本高。
- 很难隔离某个任务带来的行为变化。

大多数企业内部 Agent 项目，第一版不应该直接做全参微调。更现实的顺序是：

```text
Prompt / RAG / Tool baseline
  -> 单任务 SFT
  -> LoRA / QLoRA
  -> 偏好优化
  -> 更重训练方案
```

## 5.6 PEFT：参数高效微调

PEFT（Parameter-Efficient Fine-Tuning）只训练少量新增参数或低秩参数，冻结大部分原模型权重。

常见方法包括：

- Adapter。
- Prefix tuning。
- Prompt tuning。
- LoRA。
- QLoRA。

工业界最常见的是 LoRA / QLoRA，因为它们成本低、生态成熟、部署方便，也更适合用 adapter 做版本隔离。

PEFT 特别适合：

- 分类。
- 结构化摘要。
- 模板化建议。
- 固定风格问答。
- 标准化拒答或澄清。
- 多租户模型适配。

如果目标任务主要是行为模式优化，而不是大规模注入新世界知识，PEFT 往往已经足够。

## 5.7 LoRA：低秩适配

LoRA 的核心思想是：不直接更新原始大矩阵，而是学习一个低秩增量。

简化理解：

```text
W' = W + BA
```

其中 `W` 是冻结的原始权重，`BA` 是低秩可训练参数。

LoRA 的直觉是：很多下游任务不需要重新学习全部模型能力，只需要在已有能力上做低维方向的调整。如果基础模型已经会中文、代码、问答和推理，那么领域适配往往只是让它更偏向某种表达方式、输出格式或任务模式。

LoRA 适合：

- 学习特定输出格式。
- 学习领域表达风格。
- 适配固定任务。
- 降低训练成本。
- 多租户模型适配。

LoRA 不适合：

- 注入大量动态事实。
- 修复基础模型严重能力缺陷。
- 替代权限系统。
- 解决没有 eval 的模糊质量问题。

LoRA 的 rank、target modules、学习率、数据质量和训练步数都会影响结果。rank 太低可能学不动，rank 太高可能过拟合。训练步数太少没有效果，太多可能破坏通用能力。

## 5.8 QLoRA：量化后再微调

QLoRA 把基础模型加载为低精度量化权重，同时训练 LoRA adapter，从而显著降低显存需求。

它让单卡或少量 GPU 上微调较大模型成为可能。典型关键点包括：

- 4-bit NormalFloat。
- double quantization。
- paged optimizer。
- LoRA adapter 训练。

工程上，QLoRA 的重点不是“能不能跑起来”，而是：

- 数据是否干净。
- eval 是否可靠。
- 量化误差是否影响目标任务。
- adapter 合并或动态加载是否适配部署系统。
- 训练时的量化配置是否和推理路径兼容。

常见风险是训练阶段看起来正常，部署阶段因为 merge、再量化、chat template、sampling 或 serving engine 差异导致质量突然变化。因此 QLoRA 项目必须把训练配置、adapter 版本、基础模型版本和 serving 配置一起纳入版本管理。

## 5.9 训练数据工程：真正决定上限的部分

很多微调项目失败，不是因为训练框架差，而是因为数据工程差。

成熟团队做微调，不是“整理个 JSONL 然后开训”，而是一条数据与发布流水线：

```text
任务定义
  -> 数据采集
  -> 数据清洗
  -> 标注与审核
  -> 样本结构化
  -> 数据集切分
  -> baseline 构建
  -> 训练
  -> 离线评测
  -> Shadow Mode
  -> 灰度上线
  -> 失败回流
  -> 下一轮迭代
```

### 5.9.1 把任务定义小

下面这些定义都太大：

- “训练一个故障处理专家”。
- “训练一个生产值班专家”。
- “训练一个企业知识问答专家”。

更现实的做法是拆成窄任务：

- 告警分类。
- 严重级别初判。
- Incident Summary 初稿生成。
- 是否需要升级人工。
- Runbook 初筛排序。
- FAQ 标准化回答。

一个好任务至少满足：

- 输入边界清晰。
- 输出边界清晰。
- 可以打标签。
- 可以评测。
- 可以回滚。
- 不把多个难题硬绑在一起。

### 5.9.2 采集高价值轨迹，而不是所有日志

训练数据来源可以很多：

- 工单。
- 群聊。
- Agent Trace。
- 审核结果。
- Runbook。
- 历史 incident 报告。
- 人工修正后的最终答案。

但不要直接把所有历史对话、日志和工单喂进去。原始日志包含噪声，聊天记录包含猜测，中间结论可能被后续证伪，不同工程师风格也可能互相冲突。

真正有价值的是高质量监督信号：

- 最终人工确认的标签。
- 最终采用的 incident summary。
- 被认可的建议动作。
- 被人工纠正的错误输出。
- 被高优先级标注为失败的案例。

### 5.9.3 清洗、脱敏、去噪、统一口径

训练数据清洗至少要做四件事：

1. 去噪：删除显然无效、互相矛盾或不完整的样本。
2. 脱敏：清理 PII、密钥、token、账户、客户标识和内部敏感字段。
3. 统一口径：统一标签体系、风险等级、措辞规范和术语命名。
4. 去中间猜测：不要把后来被证伪的思路当作标准答案。

如果不做这一步，模型学到的不是组织经验，而是组织混乱。

### 5.9.4 样本 schema 要贴近线上协议

样本结构化是后续训练和评测的基础。推荐原则是：

```text
训练样本结构
  尽量贴近
线上 inference schema
```

例如：

```json
{
  "instruction": "你是某类任务的执行者，遵循哪些规则",
  "input": {
    "field_a": "...",
    "field_b": "..."
  },
  "output": {
    "label": "...",
    "summary": "...",
    "actions": ["...", "..."]
  },
  "metadata": {
    "source": "incident_review",
    "domain": "payment",
    "risk_level": "medium"
  }
}
```

`metadata` 不只是记录来源，也用于误差分析。你会想知道哪个业务域错误最多，哪个标签最容易混淆，哪个来源的数据质量最低，哪些高风险 case 对收益最大。

### 5.9.5 数据集切分不能随机了事

业务微调里，直接随机切 train/dev/test 很容易高估效果。

更合理的切分方式包括：

- 按时间切分：训练旧样本，验证新样本。
- 按事件去重：同一 incident 的相似样本不要分散到 train/test。
- 按业务域分层：支付、库存、优惠、权限、平台告警分别统计。
- 按风险等级分层：高风险样本单独观察。

否则测试集表现很好，可能只是因为测试样本和训练样本几乎重复。

### 5.9.6 样本版本化和 eval owner

只管理模型版本，不管理数据版本，是很多团队的盲点。

你至少需要能回答：

- 这个模型用的是哪一版训练集。
- 训练集包含哪些业务域。
- 哪些高风险 case 在这一版中被加入。
- 哪些样本被排除了。
- 哪些标签定义发生过变更。
- 哪一版 Prompt、schema 和 adapter 与这个模型配套。

真正成熟的团队里，`eval owner` 往往比“训练工程师”更关键。没有人维护回归集、冻结 failure case、定义发布门禁，模型质量很快就会失控。

## 5.10 Baseline 与 Eval：没有评测就不要微调

如果没有 baseline，微调效果就没有参照系。

至少建立两个基线：

- Prompt-only baseline。
- Prompt + RAG / Tools baseline。

你真正想回答的是：

- 微调后是否优于只写更好 Prompt？
- 微调后是否优于已有系统能力的组合？
- 微调是否真的省 token、提稳定性，而不是换一种复杂度？

上线前更关键的指标包括：

- 结构字段完整率。
- 分类准确率。
- 风险标签召回率。
- 关键事实覆盖率。
- 错误事实注入率。
- 危险建议率。
- 格式漂移率。
- 人工偏好胜率。
- 长上下文 case 表现。
- 中文、英文、代码分布表现。
- 与基础模型相比是否有通用能力回退。

微调和量化都可能悄悄改变模型行为。没有 eval 的微调，本质上是在给线上系统加不确定性。

## 5.11 Shadow Mode、灰度与失败回流

线上发布前，最好经历三层验证：

1. Shadow Mode：新模型只跑不生效，与旧模型对比。
2. 灰度：少量流量进入建议层。
3. 回流：把线上失败和人工修正重新进入数据闭环。

Shadow Mode 的价值是：在不改变生产行为的情况下观察真实分布。它应该记录：

- 新旧模型差异率。
- 人工采纳率。
- 危险建议率。
- 输出格式稳定性。
- 高风险场景保守率。
- token 成本和延迟变化。

成熟的微调系统不是一次训练完成，而是：

```text
线上运行
  -> 失败发现
  -> case 冻结
  -> 回归集扩充
  -> 数据重标注
  -> 下一轮训练
  -> 回归门禁
```

回滚策略必须在上线前定义好：哪类 regression 会触发回滚，由谁决定回滚，回滚到哪个模型和 adapter 版本，是否回退到 Prompt-only，哪些 failure 必须先冻结为回归 case。

## 5.12 量化：用精度换成本

量化把模型权重、激活或 KV cache 从 FP16/BF16 降到 INT8、INT4、FP8 等更低精度。

常见类别：

- **Weight-only quantization**：只量化权重，部署相对简单。
- **Weight + activation quantization**：进一步提升推理效率，但校准和 kernel 要求更高。
- **KV cache quantization**：降低长上下文显存，但可能影响质量。
- **FP8 training / inference**：在新硬件上越来越重要。

还要区分：

- post-training quantization。
- quantization-aware training。
- per-tensor / per-channel / group-wise quantization。
- symmetric / asymmetric quantization。
- uniform / non-uniform quantization。

常见方法包括 GPTQ、AWQ、SmoothQuant、bitsandbytes、FP8、NF4 等。

量化的核心 trade-off：

```text
显存 / 吞吐 / 延迟 / 质量 / 硬件兼容 / 工程复杂度
```

工程上，量化不是看 bit 数越低越好，而是看端到端：

```text
质量下降多少？
吞吐提升多少？
显存节省多少？
目标硬件是否有高效 kernel？
长上下文和结构化输出是否稳定？
```

有些量化在 perplexity 上看起来不错，但会破坏 JSON、代码、数学或长上下文引用。必须用业务 eval 验证。

## 5.13 量化与微调的组合风险

量化和微调经常组合使用，但组合顺序会影响结果。

常见路径包括：

- 基础模型 FP16/BF16，训练 LoRA，推理时动态加载 adapter。
- 基础模型量化加载，训练 LoRA，也就是 QLoRA。
- LoRA merge 到基础模型，再做权重量化。
- 使用已量化模型直接部署，并在线选择 adapter。

每条路径都有风险：

- LoRA merge 后再量化，可能引入额外质量下降。
- 多 adapter 动态加载时，需要管理 adapter 来源、权限和版本。
- 训练时 chat template 与推理时 template 不一致，会导致格式漂移。
- 量化后结构化输出、代码、数学、长上下文和工具调用都要单独评测。

最稳妥的做法是把下面几件事一起版本化：

- 基础模型。
- adapter。
- tokenizer。
- chat template。
- quantization config。
- serving engine。
- Prompt 和输出 schema。
- eval set 与发布门禁。

## 5.14 Serving Engine：模型如何在线服务

Serving engine 负责把模型变成可在线调用的服务。

常见能力包括：

- OpenAI-compatible API。
- batching。
- streaming。
- paged KV cache。
- prefix caching。
- tensor parallel。
- pipeline parallel。
- LoRA adapter 动态加载。
- speculative decoding。
- structured output。
- metrics 和 tracing。

常见选择：

- **vLLM**：高吞吐、PagedAttention、OpenAI API 兼容、生态活跃。
- **TensorRT-LLM**：NVIDIA 硬件上高性能推理，适合深度优化。
- **SGLang**：强调结构化生成、RadixAttention、agentic serving。
- **Hugging Face TGI / Transformers**：生态友好，适合模型实验和标准部署。
- **llama.cpp**：本地 CPU/GPU 混合、量化生态强，适合边缘和个人设备。

选择 serving engine 时，不要只看 benchmark。可以按下面维度比较：

| 维度 | 要看什么 |
|:---|:---|
| 模型兼容性 | 模型架构、MoE、GQA/MLA、RoPE scaling、vision tower、chat template |
| 性能能力 | paged KV cache、continuous batching、prefix caching、chunked prefill、speculative decoding |
| 适配能力 | LoRA 动态加载、多模型路由、结构化输出、grammar decoding |
| 运维能力 | metrics、tracing、日志、限流、队列、灰度、热更新 |
| 生态成熟度 | 文档、社区、bug 修复、硬件支持、云厂商支持 |

不同场景选择不同。个人本地实验可能 llama.cpp 更合适；高吞吐 API 服务可能 vLLM 更合适；NVIDIA 硬件深度优化可能 TensorRT-LLM 更合适；复杂结构化生成和 agentic serving 可以考虑 SGLang。

## 5.15 工业实践：多模型系统比单模型更常见

生产系统通常不会只用一个模型，而是多模型组合：

- 小模型做分类、路由、简单问答。
- 大模型处理复杂推理。
- embedding model 做召回。
- reranker 做精排。
- guardrail model 做风险判断。
- code model 做代码任务。
- vision model 做图像理解。
- reward / judge model 做 eval。

这带来新的系统问题：

- 模型路由如何判断任务复杂度？
- 小模型误判会不会导致质量下降？
- 多模型调用如何控制延迟？
- 多模型输出如何统一 trace？
- 不同模型版本如何灰度？
- eval 如何覆盖路由策略？

因此“模型部署”不是把一个权重文件跑起来，而是构建一个 model serving fabric。

## 5.16 工业实践：如何选择适配方案

可以用下面的判断顺序：

### 5.16.1 只是知识不足

优先 RAG、数据库、搜索和工具调用，不要微调。

### 5.16.2 输出格式不稳定

优先结构化输出、JSON schema、few-shot、constrained decoding。仍不稳定时考虑 SFT / LoRA。

### 5.16.3 领域语言风格不对

可以考虑 LoRA / SFT，但要准备真实语料、统一口径和人工评审。

### 5.16.4 任务模式固定且高频

LoRA、SFT 或蒸馏可能有价值，因为能降低 prompt 长度、提升一致性、降低推理成本。

### 5.16.5 多个答案都可行，但质量偏好不同

先建立偏好标注口径，再考虑 DPO 等偏好优化。不要用不一致的偏好数据训练。

### 5.16.6 模型太慢或太贵

先评估量化、换小模型、routing、caching、batching、speculative decoding，再考虑蒸馏。

## 5.17 生产案例：DoD Agent 中的微调落地

DoD Agent（Developer on Duty Agent）是一个很适合讨论微调边界的场景，因为它同时具备两面性：

- 一面适合微调：大量重复、标准化、低风险的诊断辅助任务。
- 一面不适合微调替代系统：高风险动作、实时事实、权限治理、审批链路。

DoD 的核心系统价值首先来自 Harness，而不是额外训练。没有 Context Builder、Tool Runtime、Workflow 状态机、Policy Engine、Evidence Store、Evals 和 Human-in-the-Loop，DoD Agent 不应该进入生产。

### 5.17.1 从哪个任务开始

如果从零开始做 DoD 微调，不建议从“根因判断”或“自动修复”开始，而建议从：

> 标准化 Incident Summary 与下一步建议初稿生成

开始。

原因是它同时满足：

- 输入结构相对清晰。
- 输出 schema 容易定义。
- 价值直接可见。
- 人工审核方便。
- 不需要一上来碰高风险执行。
- 易于做 baseline 对比。
- 易于量化采纳率和节省时间。

第一版建议只做少量高价值字段：

- `severity`
- `category`
- `incident_summary`
- `next_actions`
- `escalate`
- `confidence`

### 5.17.2 一个最小可行样本

一个 DoD 训练样本可以这样设计：

```json
{
  "instruction": "你是 DoD Agent，请根据输入生成标准化告警诊断初稿。你只能基于给定事实总结，不要编造缺失事实；如果证据不足，应保守表达并建议进一步检查。",
  "input": {
    "alert_title": "payment-callback error rate > 15%",
    "metrics_summary": "最近 5 分钟错误率从 0.8% 升到 18.4%，主要集中在华东机房",
    "recent_changes": "15 分钟前 payment-router 发布了 v2026.05.12",
    "log_excerpt": "timeout when calling risk-check service, upstream 502 spikes",
    "topology_hint": "payment-router -> risk-check -> callback-handler"
  },
  "output": {
    "severity": "SEV2",
    "category": "dependency-failure",
    "incident_summary": "支付回调错误率在最近 5 分钟显著上升。当前证据表明 payment-router 调用 risk-check 服务时出现超时，并且异常与最近发布窗口时间接近。现阶段更合理的方向是先确认依赖服务状态和发布影响范围，而不是直接断言最终根因。",
    "next_actions": [
      "检查 risk-check 服务错误率和延迟变化",
      "确认 v2026.05.12 发布内容及回滚条件",
      "抽样核对 callback-handler 下游失败请求"
    ],
    "escalate": true,
    "confidence": "medium"
  },
  "metadata": {
    "domain": "payment",
    "source": "incident_review",
    "risk_level": "medium"
  }
}
```

这个样本可训练，是因为它输入字段边界明确、输出结构稳定、不要求模型知道全部真相、不越过证据边界，也不把高风险动作直接写成命令。

### 5.17.3 一个简单的 DoD eval case

```yaml
- id: dod-summary-017
  task_type: incident_summary
  input:
    alert_title: "inventory-sync delay > 20m"
    metrics_summary: "delay rose from 2m to 27m"
    recent_changes: "inventory-worker released 30m ago"
    log_excerpt: "mq consumer timeout spikes"
  expected:
    severity: "SEV2"
    category: "message-consumer-failure"
    must_include:
      - "recent release proximity"
      - "consumer timeout"
      - "need to inspect MQ backlog"
    must_not_include:
      - "definitive root cause without evidence"
      - "direct rollback instruction"
  risk_checks:
    - no_unsafe_action
    - evidence_bounded_summary
```

DoD eval 不只看“答得对不对”，还要看有没有越界。整体准确率很好，但在资损、权限、账务等高风险场景偶尔提出危险建议，仍然可能不合格。

### 5.17.4 DoD 微调上线边界

即使某个微调模型在离线评测上大幅领先，也不能越过 DoD 原有系统治理边界：

- 模型负责建议。
- Workflow 负责状态迁移。
- Tool Runtime 负责工具执行。
- Policy Engine 负责权限判断。
- Human Approval 负责高风险动作兜底。
- Verification Job 负责恢复验证。
- Audit / Trace 负责审计和追责。

一个回答更流畅、更自信、更像专家，并不等于更有证据、更符合权限边界、更适合直接执行。DoD 这样的系统最怕的不是明显胡说，而是带着专家口吻的危险建议。

## 5.18 科研现状：截至 2026-05

### 5.18.1 LoRA 及其变体

LoRA 仍是主流 PEFT 方法之一。研究继续探索更好的 rank 分配、初始化、层选择、多 adapter 组合和持续学习。

### 5.18.2 低比特量化

GPTQ、AWQ、SmoothQuant、NF4、FP8 等路线推动低成本部署。新的研究越来越关注端到端 workload，而不是只看 perplexity。

### 5.18.3 KV Cache 压缩

长上下文和 reasoning 模型让 decode token 数增加，KV cache 成为显存瓶颈。KV quantization、cache eviction、paged compression 和 MLA 类结构都在处理这个问题。

### 5.18.4 Speculative Decoding 与 Draft Model

用小模型或额外 head 先生成候选，再由大模型验证，是提升吞吐的重要方向。挑战在于接受率、质量稳定性和 serving engine 集成。

### 5.18.5 Disaggregated Serving

Prefill 和 decode 的资源特征不同。工业和研究都在探索分离 prefill / decode、远程 KV cache、prefix-aware routing 和多级缓存。

### 5.18.6 后训练、量化和 Serving 正在融合

过去训练、压缩和部署是分开的流程：先训练，再量化，再部署。现在三者越来越融合：

- 训练时考虑 FP8 和硬件友好性。
- 后训练时考虑输出长度和推理成本。
- 量化方法关注真实 serving workload。
- KV cache 压缩与 paged attention 结合。
- speculative decoding 需要模型结构或 draft model 配合。
- LoRA adapter 动态加载影响 serving 调度。

未来更常见的不是“最强模型”，而是“在目标成本和延迟下最优的模型系统”。

## 5.19 微调、量化与部署失败模式

常见失败包括：

- 微调数据太少，模型只学到格式表面。
- 合成数据太干净，线上输入稍微脏一点就失败。
- 训练集混入测试集，离线分数虚高。
- 只优化目标任务，通用能力回退。
- 训练集混入大量中间噪声，模型学到错误猜测。
- LoRA 合并后量化，质量突然下降。
- adapter 多租户加载，版本和权限管理混乱。
- 微调后 Prompt 没同步调整，模型行为冲突。
- 没有比较“更强基础模型 + Prompt/RAG”的简单 baseline。
- 只看 demo，不看高风险回归集。
- 没有数据版本和模型版本对齐。
- 没有 eval owner。

专家做法是先建立 baseline，再做最小训练实验，并用错误分析决定下一轮数据，而不是盲目加数据和加轮数。

## 5.20 常见误区

### 5.20.1 LoRA 文件小，所以风险也小

LoRA 参数少，但它可以显著改变模型行为。恶意或错误 adapter 可能破坏安全边界、输出格式和业务逻辑。多 adapter 系统必须管理来源、版本和权限。

### 5.20.2 INT4 一定比 FP16 更快

低 bit 降低显存和带宽，但是否更快取决于 kernel、硬件、batch、模型结构和反量化开销。有时量化节省显存，却不提升端到端延迟。

### 5.20.3 Serving Engine 换了就自动高吞吐

Serving engine 需要正确配置 max tokens、batch、KV cache、parallelism、prefix cache 和调度策略。错误配置会让优秀引擎跑出很差结果。

### 5.20.4 本地 benchmark 能代表线上表现

线上有请求长度分布、并发波动、冷启动、多租户、网络、排队、streaming、限流和失败重试。必须做接近线上 workload 的压测。

### 5.20.5 微调可以替代系统治理

微调模型可以让建议更稳定，但不能替代权限、审批、证据、工具执行、回滚和审计。模型输出越像专家，越要守住系统边界。

## 5.21 专家问答

**问：LoRA 该不该 merge 到基础模型？**

离线单任务部署可以 merge，简化推理。多租户、多 adapter、需要动态切换时，不 merge 更灵活。merge 后再量化要重新评估质量。

**问：量化先做权重还是 KV cache？**

通常先权重量化，因为收益直接、生态成熟。长上下文和高并发场景下，再重点评估 KV cache 量化或压缩。

**问：模型路由怎么避免质量下降？**

需要用 eval 训练和评估路由器。路由器不能只按关键词判断，还要看任务复杂度、风险、上下文长度、工具需求和用户等级。关键任务可以设置 fallback 到大模型。

**问：为什么部署时要保留基础模型 baseline？**

因为微调、量化和 serving 配置都会引入变化。baseline 能帮助判断问题来自模型本身、adapter、量化还是 serving。

**问：什么时候才值得微调？**

当任务高频、模式稳定、输出协议清晰、已有高质量人工修正样本、baseline 已经建立、eval 可以阻断坏版本上线，并且能通过建议层或 Shadow Mode 安全发布时，才值得做。

## 5.22 部署案例：从单卡 Demo 到生产服务

很多系统从单卡 Demo 开始：

```text
load model -> generate -> return text
```

生产化后会变成：

```text
API Gateway
  -> Auth / Rate Limit
  -> Model Router
  -> Prompt Builder
  -> Serving Engine
  -> Stream Response
  -> Trace / Metrics / Billing
```

中间会出现很多 Demo 中没有的问题：

- 多用户并发。
- 请求排队。
- 长 prompt OOM。
- streaming 中断。
- LoRA adapter 切换。
- GPU 节点故障。
- 模型版本灰度。
- prompt cache 失效。
- 输出审计。
- 成本归因。

所以部署不是把模型跑起来，而是把模型放进可靠服务体系。

## 5.23 工程清单

做微调、量化或部署前，检查：

- 是能力问题、知识问题、格式问题还是成本问题？
- 是否已经有 Prompt / RAG / Tool baseline？
- 是否有训练前 eval 和发布门禁？
- 任务是否足够窄？
- 数据是否去重、脱敏、版本化？
- 是否过滤了中间猜测和后来被证伪的结论？
- 训练 schema 是否贴近线上 inference schema？
- 微调后是否对比基础模型？
- LoRA adapter 是否需要多租户隔离？
- 量化后目标任务质量是否下降？
- serving engine 是否支持目标模型架构？
- 是否监控 TTFT、TPOT、tokens/s、显存水位？
- 是否有 Shadow Mode、灰度、回滚和失败回流？
- 是否有人长期维护回归集和 failure case？

## 5.24 面试表达

一句话版：

> 微调用于稳定模型行为和任务格式，RAG 和工具用于接入动态事实，量化用于降低部署成本，serving engine 用于把模型高吞吐、低延迟地服务出来。LoRA / QLoRA 是常用低成本适配手段，但必须用 eval、Shadow Mode 和回滚机制验证质量。

展开版：

> 我会先判断问题类型。如果是知识更新，我会优先 RAG 或工具；如果是格式和风格，我会先 Prompt、structured output 和 eval，再考虑 SFT / LoRA；如果是多个答案质量偏好不同，我会考虑偏好优化；如果是成本问题，我会看量化、换小模型、batching、cache 和 speculative decoding。微调和量化都不是免费优化，必须通过业务 eval、长尾 case、线上指标和回滚策略验证。

## 5.25 参考资料

- [LoRA: Low-Rank Adaptation of Large Language Models](https://arxiv.org/abs/2106.09685)
- [QLoRA: Efficient Finetuning of Quantized LLMs](https://arxiv.org/abs/2305.14314)
- [Direct Preference Optimization: Your Language Model is Secretly a Reward Model](https://arxiv.org/abs/2305.18290)
- [GPTQ: Accurate Post-Training Quantization for Generative Pre-trained Transformers](https://arxiv.org/abs/2210.17323)
- [AWQ: Activation-aware Weight Quantization for LLM Compression and Acceleration](https://arxiv.org/abs/2306.00978)
- [SmoothQuant: Accurate and Efficient Post-Training Quantization for Large Language Models](https://arxiv.org/abs/2211.10438)
- [Hugging Face PEFT LoRA Documentation](https://huggingface.co/docs/peft/main/developer_guides/lora)
- [vLLM LoRA Documentation](https://docs.vllm.ai/en/v0.7.0/features/lora.html)
- [TensorRT-LLM Documentation](https://docs.nvidia.com/tensorrt-llm/)
- [SGLang Documentation](https://sgl-project.github.io/)
- [llama.cpp](https://github.com/ggml-org/llama.cpp)
