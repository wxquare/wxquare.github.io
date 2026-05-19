# 第26章 微调、量化与部署：LoRA、QLoRA、Serving Engine

企业真正落地大模型时，常见问题不是“怎么训练一个 GPT”，而是：已有模型怎么适配业务，怎么降低成本，怎么稳定部署，怎么在质量、延迟和显存之间取舍。

## 26.1 宏观理解：适配模型的几条路

面对一个业务需求，通常有五种手段：

```text
Prompt -> RAG / Tool -> SFT -> LoRA / QLoRA -> Full Fine-tuning / Continued Pretraining
```

越往右，成本越高、风险越大、对数据和评估要求越高。

不要把微调当成默认选项。很多问题通过 Prompt、RAG、工具和 eval 就能解决。

## 26.2 Full Fine-tuning

Full fine-tuning 更新模型全部参数。

优点：

- 适配能力强；
- 可以改变模型深层行为；
- 适合大规模、高价值、稳定任务。

缺点：

- 显存和计算成本高；
- 容易灾难性遗忘；
- 需要高质量数据和严格 eval；
- 部署和版本管理成本高。

## 26.3 PEFT：参数高效微调

PEFT（Parameter-Efficient Fine-Tuning）只训练少量新增参数或低秩参数，冻结大部分原模型权重。

常见方法包括：

- Adapter；
- Prefix tuning；
- Prompt tuning；
- LoRA；
- QLoRA。

工业界最常见的是 LoRA / QLoRA，因为它们成本低、生态成熟、部署方便。

## 26.4 LoRA：低秩适配

LoRA 的核心思想是：不直接更新原始大矩阵，而是学习一个低秩增量。

简化理解：

```text
W' = W + BA
```

其中 `W` 是冻结的原始权重，`BA` 是低秩可训练参数。

LoRA 适合：

- 学习特定输出格式；
- 学习领域表达风格；
- 适配固定任务；
- 降低训练成本；
- 多租户模型适配。

LoRA 不适合：

- 注入大量动态事实；
- 修复基础模型严重能力缺陷；
- 替代权限系统；
- 解决没有 eval 的模糊质量问题。

## 26.5 QLoRA：量化后再微调

QLoRA 把基础模型加载为低精度量化权重，同时训练 LoRA adapter，从而显著降低显存需求。

它让单卡或少量 GPU 上微调较大模型成为可能。典型关键点包括：

- 4-bit NormalFloat；
- double quantization；
- paged optimizer；
- LoRA adapter 训练。

工程上，QLoRA 的重点不是“能不能跑起来”，而是：

- 数据是否干净；
- eval 是否可靠；
- 量化误差是否影响目标任务；
- adapter 合并或动态加载是否适配部署系统。

## 26.6 量化：用精度换成本

量化把模型权重、激活或 KV cache 从 FP16/BF16 降到 INT8、INT4、FP8 等更低精度。

常见类别：

- **Weight-only quantization**：只量化权重，部署相对简单。
- **Weight + activation quantization**：进一步提升推理效率，但校准和 kernel 要求更高。
- **KV cache quantization**：降低长上下文显存，但可能影响质量。
- **FP8 training / inference**：在新硬件上越来越重要。

常见方法包括 GPTQ、AWQ、SmoothQuant、bitsandbytes、FP8、NF4 等。

量化的核心 trade-off：

```text
显存 / 吞吐 / 延迟 / 质量 / 硬件兼容 / 工程复杂度
```

## 26.7 Serving Engine：模型如何在线服务

Serving engine 负责把模型变成可在线调用的服务。

常见能力包括：

- OpenAI-compatible API；
- batching；
- streaming；
- paged KV cache；
- prefix caching；
- tensor parallel；
- pipeline parallel；
- LoRA adapter 动态加载；
- speculative decoding；
- structured output；
- metrics 和 tracing。

常见选择：

- **vLLM**：高吞吐、PagedAttention、OpenAI API 兼容、生态活跃。
- **TensorRT-LLM**：NVIDIA 硬件上高性能推理，适合深度优化。
- **SGLang**：强调结构化生成、RadixAttention、agentic serving。
- **Hugging Face TGI / Transformers**：生态友好，适合模型实验和标准部署。
- **llama.cpp**：本地 CPU/GPU 混合、量化生态强，适合边缘和个人设备。

## 26.8 工业实践：如何选择适配方案

可以用下面的判断顺序：

### 1. 只是知识不足

优先 RAG、数据库、工具调用，不要微调。

### 2. 输出格式不稳定

优先结构化输出、JSON schema、few-shot、constrained decoding。仍不稳定时考虑 SFT/LoRA。

### 3. 领域语言风格不对

可以考虑 LoRA/SFT，但要准备真实语料和人工评审。

### 4. 任务模式固定且高频

LoRA 或蒸馏可能有价值，因为能降低 prompt 长度和推理成本。

### 5. 模型太慢或太贵

先评估量化、换小模型、routing、caching、batching、speculative decoding，再考虑蒸馏。

## 26.9 工业实践：部署前必须有 Eval

微调和量化都可能悄悄改变模型行为。上线前至少要有：

- 目标任务准确率；
- 格式遵循率；
- 幻觉率；
- 拒答率；
- 安全 case；
- 长上下文 case；
- 中文/英文/代码分布；
- 延迟和吞吐；
- 成本评估；
- 回滚策略。

没有 eval 的微调，本质上是在给线上系统加不确定性。

## 26.10 科研现状：截至 2026-05

### 1. LoRA 及其变体

LoRA 仍是主流 PEFT 方法之一。研究继续探索更好的 rank 分配、初始化、层选择、多 adapter 组合和持续学习。

### 2. 低比特量化

GPTQ、AWQ、SmoothQuant、NF4、FP8 等路线推动低成本部署。新的研究越来越关注端到端 workload，而不是只看 perplexity。

### 3. KV Cache 压缩

长上下文和 reasoning 模型让 decode token 数增加，KV cache 成为显存瓶颈。KV quantization、cache eviction、paged compression 和 MLA 类结构都在处理这个问题。

### 4. Speculative Decoding 与 Draft Model

用小模型或额外 head 先生成候选，再由大模型验证，是提升吞吐的重要方向。挑战在于接受率、质量稳定性和 serving engine 集成。

### 5. Disaggregated Serving

Prefill 和 decode 的资源特征不同。工业和研究都在探索分离 prefill/decode、远程 KV cache、prefix-aware routing 和多级缓存。

## 26.11 工程清单

做微调、量化或部署前，检查：

- 是能力问题、知识问题、格式问题还是成本问题？
- 是否有训练前 eval？
- 数据是否去重、脱敏、版本化？
- 微调后是否对比基础模型？
- LoRA adapter 是否需要多租户隔离？
- 量化后目标任务质量是否下降？
- serving engine 是否支持目标模型架构？
- 是否监控 TTFT、TPOT、tokens/s、显存水位？
- 是否有回滚和灰度发布？

## 26.12 面试表达

一句话版：

> 微调用于改变模型行为和任务格式，RAG 用于接入动态事实，量化用于降低部署成本，serving engine 用于把模型高吞吐、低延迟地服务出来。LoRA/QLoRA 是常用低成本适配手段，但必须用 eval 验证质量。

展开版：

> 我会先判断问题类型。如果是知识更新，我会优先 RAG 或工具；如果是格式和风格，我会先 Prompt 和 structured output，再考虑 LoRA/SFT；如果是成本问题，我会看量化、换小模型、batching、cache 和 speculative decoding。微调和量化都不是免费优化，必须通过业务 eval、长尾 case 和线上指标验证。

## 26.13 深入理解：LoRA 为什么有效

LoRA 的直觉是：很多下游任务不需要重新学习全部模型能力，只需要在已有能力上做低维方向的调整。

如果基础模型已经会中文、代码、问答和推理，那么领域适配往往只是让它更偏向某种表达方式、输出格式或任务模式。这个增量可能在参数空间中是低秩的，因此可以用很少的可训练参数表达。

这也是 LoRA 的边界：

- 它擅长“偏转已有能力”；
- 不擅长“创造基础模型没有的能力”；
- 不适合“可靠写入大量事实”；
- 不应该替代检索和工具。

LoRA 的 rank、target modules、学习率、数据质量和训练步数都会影响结果。rank 太低可能学不动，rank 太高可能过拟合。训练步数太少没有效果，太多可能破坏通用能力。

## 26.14 深入理解：量化不是单一技术

量化经常被一句“INT4 部署”带过，但实际很复杂。

至少要区分：

- **权重量化**：压缩模型权重，降低显存和带宽。
- **激活量化**：压缩中间激活，对校准和 kernel 要求更高。
- **KV cache 量化**：降低长上下文显存，对 decode 质量敏感。
- **训练量化**：例如 FP8 mixed precision，影响训练效率和稳定性。

还要区分：

- post-training quantization；
- quantization-aware training；
- per-tensor / per-channel / group-wise quantization；
- symmetric / asymmetric quantization；
- uniform / non-uniform quantization。

工程上，量化不是看 bit 数越低越好，而是看端到端：

```text
质量下降多少？
吞吐提升多少？
显存节省多少？
目标硬件是否有高效 kernel？
长上下文和结构化输出是否稳定？
```

有些量化在 perplexity 上看起来不错，但会破坏 JSON、代码、数学或长上下文引用。必须用业务 eval 验证。

## 26.15 工业实践：Serving Engine 选型框架

选择 serving engine 时，不要只看 benchmark。可以按下面维度比较：

### 1. 模型兼容性

是否支持目标模型架构、MoE、GQA/MLA、RoPE scaling、vision tower、tool calling chat template。

### 2. 性能能力

是否支持 paged KV cache、continuous batching、prefix caching、chunked prefill、speculative decoding、tensor parallel。

### 3. 适配能力

是否支持 LoRA adapter 动态加载、多模型路由、结构化输出、grammar decoding。

### 4. 运维能力

是否有 metrics、tracing、日志、限流、队列、灰度、热更新、OpenAI-compatible API。

### 5. 生态成熟度

社区活跃度、文档质量、bug 修复速度、硬件支持、云厂商支持都很重要。

不同场景选择不同。个人本地实验可能 llama.cpp 更合适；高吞吐 API 服务可能 vLLM 更合适；NVIDIA 硬件深度优化可能 TensorRT-LLM 更合适；复杂结构化生成和 agentic serving 可以考虑 SGLang。

## 26.16 工业实践：多模型系统比单模型更常见

生产系统通常不会只用一个模型，而是多模型组合：

- 小模型做分类、路由、简单问答；
- 大模型处理复杂推理；
- embedding model 做召回；
- reranker 做精排；
- guardrail model 做风险判断；
- code model 做代码任务；
- vision model 做图像理解；
- reward/judge model 做 eval。

这带来新的系统问题：

- 模型路由如何判断任务复杂度？
- 小模型误判会不会导致质量下降？
- 多模型调用如何控制延迟？
- 多模型输出如何统一 trace？
- 不同模型版本如何灰度？
- eval 如何覆盖路由策略？

因此“模型部署”不是把一个权重文件跑起来，而是构建一个 model serving fabric。

## 26.17 研究补充：后训练、量化和 Serving 正在融合

过去训练、压缩和部署是分开的流程：先训练，再量化，再部署。现在三者越来越融合。

例如：

- 训练时考虑 FP8 和硬件友好性；
- 后训练时考虑输出长度和推理成本；
- 量化方法关注真实 serving workload；
- KV cache 压缩与 paged attention 结合；
- speculative decoding 需要模型结构或 draft model 配合；
- LoRA adapter 动态加载影响 serving 调度。

未来更常见的不是“最强模型”，而是“在目标成本和延迟下最优的模型系统”。

这要求工程师同时理解模型、训练、推理引擎和业务指标。

## 26.18 微调项目失败模式

常见失败包括：

- 微调数据太少，模型只学到格式表面。
- 合成数据太干净，线上输入稍微脏一点就失败。
- 训练集混入测试集，离线分数虚高。
- 只优化目标任务，通用能力回退。
- LoRA 合并后量化，质量突然下降。
- adapter 多租户加载，版本和权限管理混乱。
- 微调后 prompt 没同步调整，模型行为冲突。
- 没有比较“更强基础模型 + Prompt/RAG”的简单 baseline。

专家做法是先建立 baseline，再做最小训练实验，并用错误分析决定下一轮数据，而不是盲目加数据和加轮数。

## 26.19 常见误区：微调、量化与部署

### 误区 1：LoRA 文件小，所以风险也小

LoRA 参数少，但它可以显著改变模型行为。恶意或错误 adapter 可能破坏安全边界、输出格式和业务逻辑。多 adapter 系统必须管理来源、版本和权限。

### 误区 2：INT4 一定比 FP16 更快

低 bit 降低显存和带宽，但是否更快取决于 kernel、硬件、batch、模型结构和反量化开销。有时量化节省显存，却不提升端到端延迟。

### 误区 3：Serving Engine 换了就自动高吞吐

Serving engine 需要正确配置 max tokens、batch、KV cache、parallelism、prefix cache 和调度策略。错误配置会让优秀引擎跑出很差结果。

### 误区 4：本地 benchmark 能代表线上表现

线上有请求长度分布、并发波动、冷启动、多租户、网络、排队、streaming、限流和失败重试。必须做接近线上 workload 的压测。

## 26.20 专家问答

**问：LoRA 该不该 merge 到基础模型？**

离线单任务部署可以 merge，简化推理。多租户、多 adapter、需要动态切换时，不 merge 更灵活。merge 后再量化要重新评估质量。

**问：量化先做权重还是 KV cache？**

通常先权重量化，因为收益直接、生态成熟。长上下文和高并发场景下，再重点评估 KV cache 量化或压缩。

**问：模型路由怎么避免质量下降？**

需要用 eval 训练和评估路由器。路由器不能只按关键词判断，还要看任务复杂度、风险、上下文长度、工具需求和用户等级。关键任务可以设置 fallback 到大模型。

**问：为什么部署时要保留基础模型 baseline？**

因为微调、量化和 serving 配置都会引入变化。baseline 能帮助判断问题来自模型本身、adapter、量化还是 serving。

## 26.21 部署案例：从单卡 Demo 到生产服务

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

- 多用户并发；
- 请求排队；
- 长 prompt OOM；
- streaming 中断；
- LoRA adapter 切换；
- GPU 节点故障；
- 模型版本灰度；
- prompt cache 失效；
- 输出审计；
- 成本归因。

所以部署不是把模型跑起来，而是把模型放进可靠服务体系。

## 26.22 部署案例：企业内部多模型路由

一个企业内部 AI 平台可能这样路由：

- 分类小模型判断任务类型；
- embedding model 做知识召回；
- reranker 精排；
- 轻量 LLM 处理简单问答；
- 强 LLM 处理复杂推理；
- code model 处理代码任务；
- judge model 做离线 eval；
- guardrail model 做安全判断。

路由策略要考虑：

- 质量；
- 成本；
- 延迟；
- 用户权限；
- 数据敏感级别；
- 是否需要工具；
- 是否需要长上下文；
- 是否需要可审计输出。

这种架构的难点是整体 eval。单个模型分数高，不代表路由系统质量高。必须评估端到端任务完成率和错误路由率。

## 26.23 参考资料

- [LoRA: Low-Rank Adaptation of Large Language Models](https://arxiv.org/abs/2106.09685)
- [QLoRA: Efficient Finetuning of Quantized LLMs](https://arxiv.org/abs/2305.14314)
- [GPTQ](https://arxiv.org/abs/2210.17323)
- [AWQ](https://arxiv.org/abs/2306.00978)
- [SmoothQuant](https://arxiv.org/abs/2211.10438)
- [vLLM Documentation](https://docs.vllm.ai/)
- [TensorRT-LLM Documentation](https://nvidia.github.io/TensorRT-LLM/)
- [SGLang Documentation](https://docs.sglang.io/)
- [llama.cpp](https://github.com/ggml-org/llama.cpp)
