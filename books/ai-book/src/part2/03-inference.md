# 第10章 推理服务与资源调度

本章的推理请求、KV cache 与 GPU 调度是 Infra 领域特殊案例，因为它们描述承载 ALR 等上层任务的服务机制，而不是业务诊断本身；适用范围仅限本章。请求/制品、运行观测、调度动作、流式状态和 SLO 验收，分别对照 ALR 的任务输入与资源版本、Evidence、受控动作、Run 状态和诊断交付。

训练完成并不意味着模型已经可以被用户使用。训练输出是一个需要转换、装载、调度和观测的制品；在线推理则是一个同时面对突发流量、长尾上下文、流式连接、动态 batch、显存限制和版本发布的分布式服务。本章聚焦从一次 `forward` 到生产级 serving 的完整链路，重点解释 prefill 与 decode 的资源差异、KV cache 的内存管理、连续批处理、模型并行、长上下文、公平调度、量化、投机解码以及服务 SLO。

推理 Infra 的核心指标不应只有 tokens/s。交互请求关心首 token 延迟 TTFT、每个输出 token 的延迟 TPOT 和端到端尾延迟；离线任务关心 goodput 和单位成本；多步任务还要关心跨多次调用之后的最终成功率与总等待时间。一个系统可能平均吞吐很高，却让短请求被长上下文阻塞；也可能单请求很快，却在并发增加时因为 KV cache 失控而 OOM。Orca、PagedAttention、Sarathi-Serve 和 DistServe 等工作说明，LLM serving 的关键突破来自把模型执行拆成可调度的资源和状态，而不是简单把模型放进 HTTP 服务 [[1]](#ref-1)[[2]](#ref-2)[[3]](#ref-3)[[4]](#ref-4)。

训练制品不能直接交付，是因为用户面对的是请求生命周期而不是 checkpoint。推理平台必须回答五个问题：制品是否完整，容量是否够，状态是否可回收，调度是否公平，质量与成本是否仍在门限内。

| 交付问题 | Serving 对象 | 主要指标 | 失败时优先检查 |
|:---|:---|:---|:---|
| 制品能否加载 | manifest、tokenizer、template、engine | 加载成功率、兼容矩阵 | artifact hash、模板、量化配置 |
| 容量能否满足 SLO | request、queue、prefill、decode | TTFT、TPOT、goodput、p99 | 长度分布、队列、GPU 水位 |
| 状态能否回收 | KV block、stream、operation | KV 利用率、取消释放、OOM | block owner、租约、引用计数 |
| 调度是否公平 | batch、priority、tenant quota | 饥饿率、配额命中、尾延迟 | batching 策略、长请求、离线任务 |
| 成本是否可解释 | route、cache、retry、fallback | 单位成功任务成本 | cache 命中、重试、降级模型 |

## 10.1 请求生命周期与服务契约

### 模型制品不是一个权重文件

一个可服务模型至少由权重、配置、tokenizer、chat template、特殊 token、量化配置、并行配置、采样默认值、停止条件、结构化输出约束和安全策略组成。视觉或音频模型还要包含预处理器、图像尺寸限制和模态路由。模型注册中心应为这些 artifact 生成不可变版本和 hash，serving engine 只加载经过完整性、兼容性和质量门禁的 manifest。

把权重单独发布会产生很多线上差异：tokenizer 版本改变了输入长度，chat template 改变了系统提示，量化配置改变了显存与质量，停止 token 不一致导致输出过长，adapter 没有加载导致能力回退。生产发布应把 artifact manifest 当作运行时契约，并记录构建镜像、GPU 架构、CUDA/driver、engine 版本和评估结果。TensorRT 的构建/运行时架构与 Triton 的模型管理说明了制品兼容和模型生命周期的不同层次 [[5]](#ref-5)[[6]](#ref-6)。

### 请求生命周期

请求的计算状态可以是 ACCEPTED、QUEUED、PREFILLING、DECODING、COMPLETED、CANCELLED 或 FAILED；交付状态另有连接中、STREAMING、已断开和已确认。STREAMING 与 DECODING 可以同时成立，不能排成互斥的单线状态。排队占用队列和配额，prefill 占用计算与临时激活，decode 持有 KV block 和执行槽，交付轴持有连接与发送缓冲。状态迁移应幂等，重复取消不能把已完成请求改成失败。

请求的输入也不能只用一个字符串表示。平台需要知道消息、token ids、输入 token 数、最大输出 token、停止条件、工具 schema、优先级、租户、模型版本、deadline、trace id 和幂等键。若网关提前做一次 tokenization，而 engine 再做一次，输入长度与计费可能不一致；若模板在不同组件中各自拼接，cache 命中与安全审计也会出错。建议在进入调度器前生成规范化 request manifest。

### 同步、流式与异步接口

同步接口适合短请求，但不能让连接超时决定 GPU 状态；流式接口需要定义事件顺序、心跳、断线、重连和取消；异步接口需要持久化 job、查询状态、获取结果和过期清理。客户端断开时，服务端可以取消生成、继续完成并缓存结果，或转入异步任务，但必须通过产品契约决定，不能依赖某个 HTTP 框架的默认行为。

流式响应的一个常见错误是只返回文本，不返回 usage、finish reason、模型版本和错误状态。客户端可能收到部分 token 后连接断开，却无法知道是正常停止、服务失败还是用户取消。每个事件应包含 request id 和单调序号，服务端保存最小必要状态，重连时可以检测是否重复发送。工具调用的参数生成和执行更需要独立 operation id，避免重试造成重复副作用。

### 推理引擎的执行图

一个 serving engine 通常包含 tokenizer、输入规范化、请求队列、调度器、模型执行、采样器、输出流和资源回收器。每一层都可能成为瓶颈，也可能改变语义。tokenizer 在 CPU 上处理大量短请求时会耗尽线程；输入规范化如果重复复制字符串，会增加内存；采样器如果在 CPU 上逐 token 处理，会把 GPU 的收益抵消；输出流的慢客户端会占用发送缓冲和请求状态。

因此，engine 需要定义数据结构和所有权。规范化后的 token ids 由请求上下文持有，KV block 由 cache manager 持有，GPU stream 由执行器持有，流式事件由 output channel 持有；请求状态只保存引用和状态，而不是复制所有大对象。取消或失败时按照反向顺序释放：停止生成、等待 GPU stream 安全点、减少 block 引用、关闭 output channel、释放 token buffer。没有明确所有权的实现容易出现 cache 泄漏或 use-after-free。

### Sampling 与结构化输出

temperature、top-k、top-p、重复惩罚、logit bias、停止词和随机 seed 都会影响 decode 路径。某些参数可以在 GPU 上融合，某些结构化输出约束需要有限状态机、正则或 grammar mask。结构化输出会改变可选 token 集，导致采样 kernel、cache 和性能与普通文本不同。平台必须把采样配置作为请求版本的一部分，并在 usage 与 trace 中记录实际生效值。

JSON、函数参数和工具调用不能仅靠 prompt 要求。服务端应校验语法、字段类型、必需字段、最大长度和 schema 版本；失败时可以重采样、修复或返回明确错误。重采样会增加 target model 调用和 KV 保留，必须计入成本与超时。修复模型输出时不能默默改变用户意图或执行副作用，工具调用要先进入待执行状态，经过 schema 和权限检查后再提交。

### Tokenizer 服务化

大型模型 tokenizer 可能包含复杂正则、词表和特殊 token，冷启动与内存并不小。多模型服务可以共享 tokenizer worker，也可以把 tokenizer 与 engine 放在同一进程。共享降低重复内存，但模型版本、模板和租户配置更容易混淆；同进程隔离更简单，却可能在高并发短请求下成为 CPU 瓶颈。压测应分别测 tokenization、模板拼接、H2D 拷贝和 GPU prefill。

tokenizer 的 batch 化也要考虑长短混合。将很多短请求拼在一起可能提高 CPU 吞吐，但一个超长请求会延迟整个 batch。可以按输入长度分桶、设置最大等待时间并让长输入单独处理。tokenizer 输出的 token count 必须与计费、admission、KV 估算和 trace 一致，任何一个组件自行重新 tokenize 都可能造成容量与账单不一致。

### 事件时间与指标口径

中文 Transformer 教材解释了自回归生成如何反复使用已有表示 [[7]](#ref-7)；服务指标则还要计入排队和发送。以下采用客户端观察的首尾输出时间，输入到达点固定为网关接受完整请求之时；跨机事件需注明时钟误差。一个请求生成 N 个可见 token，TPOT 是首 token 之后的 N−1 个输出间隔均值，N≤1 时记为不适用，不能除零或伪造零延迟。客户端 token 事件若一次合并多 token，不能把事件数当 token 数。内部 GPU iteration time、网络 token 间隔和用户可见 TPOT 是不同口径，必须同时注明边界。

## 10.2 显存与容量模型

### 两类完全不同的计算

Prefill 一次处理全部输入 token，矩阵并行度高，主要决定 TTFT；decode 每次生成一个 token，反复读取权重与历史 KV，主要决定 TPOT、并发和输出吞吐。输入很长时，prefill 可能占据大量计算和激活显存；输出很长时，decode 会长期持有 KV block。把二者混合成一个平均 latency，会掩盖真正的瓶颈。

在线指标可写为：

```text
TTFT = first_token_timestamp - request_arrival
TPOT = (last_token_timestamp - first_token_timestamp) / (generated_tokens - 1)  # only for N > 1
E2E = queue_time + prefill_time + decode_time + postprocess_time
goodput = successful_requests_within_SLO / wall_clock
```

报告时还应按输入长度、输出长度、并发、模型版本、cache 命中和租户分桶。平均值无法表达长尾；p99 TTFT 可能由一个超长 prefill 造成，p99 TPOT 可能由 KV 水位接近上限造成。容量模型要从请求分布和服务目标出发，而不是从单请求 benchmark 外推。

### 从 token 速率到 GPU 数量

设到达率为 λ，平均输入 token 为 P，平均输出 token 为 D，单 GPU 在给定负载下的 prefill、decode 有效吞吐为 R_p、R_d，利用率预算为 u。若两阶段共享同一串行计算预算，粗估需要把工作相加：

```text
shared_pool_GPU_equivalent ≈ (λ × P / R_p + λ × D / R_d) / u
PD_pool_prefill ≥ λ × P / (R_p × u_p)
PD_pool_decode  ≥ λ × D / (R_d × u_d)
```

对独立瓶颈取 max 只能表示下界，不能视为混部所需总 GPU。PD 两池还要分别向上取整、满足拓扑与通信约束；实际混部并非线性叠加，需压测干扰。以上吞吐不是硬件峰值，必须绑定模型、长度与 batch 分布。这只是初始估算，还要加入 KV cache 容量、模型权重、临时激活、网络传输、故障余量、灰度余量和长尾请求。若输出长度有明显长尾，使用平均 D 会低估 decode slot；若输入长度不断增加，prefill 与 KV 同时增长。应分别压测短、中、长请求混合，并以满足 SLO 的 goodput 作为容量基准。

SRE 的容量和错误预算思想可以延伸到推理平台：不能把同一份剩余容量同时承诺给节点故障、扩容和滚动发布 [[8]](#ref-8)。至少应保留 failover reserve、deployment reserve 和 burst reserve，并在容量接近阈值时限制长上下文或转入低优先级队列，而不是等到 OOM 后再降级。

### Pre-allocate 与 admission control

请求进入 engine 前应先估算输入 token、最大输出 token、潜在 KV block、优先级和预计执行时间。超过租户并发、最大上下文、GPU cache 或预算时，可以拒绝、截断、转异步、降低优先级或路由到其他模型。提前 admission control 比让请求进入 decode 后才 OOM 更容易解释，也避免已经消耗大量计算后才失败。

admission control 不能只看当前 GPU memory。还要看已经分配的 KV block、正在 prefill 的激活、等待中的 token budget、权重加载、通信 buffer 和 prefix cache。对每个请求建立保守的上界可能降低利用率，因此可以使用分层策略：短请求使用精确估计，超长请求要求预留，未知长度请求采用较低优先级并设置硬上限。

### GPU 内存分区

推理节点应明确权重、KV、activation、workspace、通信 buffer 和临时 tensor 的内存上限。权重加载后，剩余显存不能全部交给 KV，因为 kernel workspace、CUDA graph、量化 scale 和异常路径仍需要空间。建议建立静态保留区与动态 cache 区，动态区根据水位触发 admission、驱逐或降级。

显存水位应有 soft limit 和 hard limit。soft limit 触发降低并发、停止接受长请求或淘汰低价值 prefix；hard limit 触发明确 OOM 保护和实例摘除。只依靠 CUDA OOM 异常通常太晚，可能导致多个请求同时失败并破坏 engine 状态。内存指标还要区分 allocated、reserved、active blocks、free blocks、fragmentation 和 pending allocation。

### Qwen3-8B：从配置到候选并发

这里做可复算的教学估计，不报告实际设备测试。来源是 Qwen 官方 `Qwen3-8B/config.json`，核验日期 2026-10-17，页面所示文件提交 `47719a2`：36 层、32 个 query heads、8 个 KV heads、head_dim 为 128、hidden_size 为 4096，默认 BF16 [[9]](#ref-9)。不能把 32 个 query heads 代入 GQA 的 KV 账本。部署时还需冻结完整模型 revision、检查点文件哈希和运行时；这里引用的主分支地址可能变化，因此这些字段与日期一起构成推演输入。

按名义 8×10^9 参数和每参数 2 字节，裸权重约 16,000,000,000 字节，即 16 GB 或 14.901 GiB。模型名的 8B 是近似，不代表检查点精确参数数；实际权重还要逐 tensor 统计，不能将此值当下载文件大小。普通完整注意力 KV 每序列每 token 为 2×36×8×128×2=147,456 字节；第一个 2 代表 K 与 V，最后的 2 是 BF16 字节数。8192 个已缓存 token 占 1,207,959,552 字节，恰为 1.125 GiB；输入与已生成输出都计入长度，预留最大输出时应使用预计总缓存长度。

| 24 GiB 单卡教学预算 | 数值 | 性质与边界 |
|:---|:---|:---|
| 名义 BF16 裸权重 | 14.901 GiB | 用8B近似，实际需检查点核算 |
| 激活、logits、通信、图和工作区 | 2 GiB | 人为给定的联合上界，待目标负载测量 |
| 分页尾块/分配器不可用空间/安全余量 | 1 GiB | 预算预留，不能与实际reserved重复相加 |
| 可用于KV的预算 | 约6.099 GiB | 24减去前三项 |
| 五条8192-token序列KV | 5.625 GiB | 没有prefix共享和beam复制 |
| 第六条加入后的KV | 6.750 GiB | 超过预算，不能准入 |

因此只得到候选上限 floor(6.099/1.125)=5；不代表五并发满足 TPOT，也不保证真实工作区低于 2 GiB。若每条总缓存长度翻倍，上限至少重新计算；若使用 block pool，应按实际 block 尾部取整。INT4 裸权重名义约4 GB，仍有scale、zero point、padding和反量化工作区；权重降精度并不会自动把 BF16 KV 缩小四倍。多卡时按各 rank 的权重/KV实际分片和复制建账，不把所有项机械除以 TP 数。这里的容量边界同时适用于更大模型，换模型先换结构与精度输入，不能只按70B与8B的比例缩放全部账本。

## 10.3 KV 与缓存的层次

### KV 的内存账本

自回归解码会保存每层每个历史 token 的 key 和 value。粗略地说，KV bytes 与层数、KV heads、head dimension、序列长度、batch 和 dtype 成正比。使用 GQA 或 MQA 可以减少 KV heads，但不会消除长上下文的线性增长。服务端内存账本至少包括权重、KV cache、activation、通信 buffer、CUDA graph workspace、tokenizer buffer 和 allocator 碎片。

一个请求的 KV 生命周期从 prefill 产生开始，随着 decode 追加 token，直到完成、取消、超时或被驱逐。若模型使用 beam、并行采样或多候选推理，KV 可能被复制或引用共享。服务端要记录 block owner、引用计数、模型版本、租户和过期时间，不能只依赖 Python 对象回收。PagedAttention 将 KV 切成固定大小 block，让动态请求不再需要连续大块显存 [[1]](#ref-1)。

缓存省去旧 token 的投影和重复前向，新 token 仍要读取历史 K/V 做注意力，因此不把完整注意力的成本自动降为线性。滑动窗口、MLA 等结构须按实际表示重建账本；KV 是可重算的中间状态，不等同于业务的持久记忆。

### 分页分配与碎片

连续分配容易产生外部碎片：短请求结束后留下许多小洞，长请求仍无法获得连续空间。分页分配把 cache 切成 block，逻辑 token 序列通过 block table 映射到物理显存。申请、追加、释放和回收类似虚拟内存，但要考虑 GPU allocator、block size、引用计数、copy-on-write 和跨 stream 同步。

block 太小会增加 table、调度和 kernel 访问开销，太大会增加内部碎片，尤其对短请求不划算。block size 要在真实请求长度分布上压测，不应只用固定长度 benchmark。指标包括 cache 利用率、内部碎片、block 分配耗时、回收延迟、OOM 次数和重新计算比例。vLLM 的实现将这类内存管理与连续批处理结合，是一个重要的工程参考 [[10]](#ref-10)。

### Prefix Cache

系统提示、工具 schema、few-shot 示例和共享文档经常形成重复前缀。Prefix cache 可以复用已计算的 KV，减少 prefill 时间，但 cache key 必须包含完整 token prefix、模型版本、tokenizer、模板、位置编码、租户和权限范围。只按字符串前缀匹配而忽略模板或权限，会造成错误结果甚至跨租户泄露。

prefix cache 需要租约、淘汰、失效和安全清理。模型升级、系统提示变更、工具 schema 变更、量化变化和 LoRA adapter 变化都可能使旧 cache 不再兼容。缓存命中率不是唯一指标，还要看命中带来的 TTFT 降低、内存占用、失效成本、冷启动和跨模型边界。高动态请求不一定适合缓存；稳定长前缀才可能覆盖管理成本。

### Cache 与取消、重试的交互

请求取消后，已经完成的 prefix KV 可以保留，也可以立即回收；选择取决于命中概率、租户策略和隐私要求。decode 中途取消时，部分输出对应的 KV 不能被另一个请求直接复用，除非它是明确的公共前缀并通过权限校验。服务重试时，如果新请求使用了同一个 idempotency key，应该查询原请求状态，而不是无条件重新分配一份 KV。

cache 泄漏通常不是显式的数据返回，而是命中率、延迟和错误模式泄露。高敏感租户可以禁用跨请求 prefix cache，或使用加密、独立 namespace 和更短 TTL。观测系统也不应把完整 token 序列作为 cache key 日志记录。OpenTelemetry 和 Prometheus 可以记录 cache hit、block 数和释放时间等摘要指标，但样本内容仍需按数据等级保护 [[11]](#ref-11)[[12]](#ref-12)。

### Block table 的一致性

PagedAttention 依赖逻辑 token 到物理 block 的映射。调度器增加 decode token 时，先检查 block capacity，再更新 block table，最后提交 kernel；释放时先阻止新的 kernel 引用，再减少引用计数。这个顺序不是实现细节，而是并发正确性的基础。多个请求共享 prefix 时，copy-on-write 要确保一个请求追加 token 不会修改另一个请求的只读 block。

block table 更新可以与 GPU stream 异步进行，但需要版本号或事件 fence 防止 kernel 读取旧映射。服务取消、超时和实例 drain 时，仍在执行的 stream 可能持有 block；立即释放会造成内存错误。可靠的 cache manager 应支持 deferred free，并在 trace 中记录 block 从 active 到 reclaimable 再到 free 的时间。

### Cache 驱逐

当 KV 水位达到 soft limit，系统需要选择驱逐对象。候选因素包括最近访问、前缀长度、租户优先级、重算成本、剩余 TTL 和请求是否仍在 decode。驱逐一个长 prefix 可能释放很多 block，却会让后续请求重复 prefill；驱逐短且高频 prefix 可能释放较少空间但降低命中率。LRU 只是起点，不一定适合多租户和长上下文。

驱逐不能影响正在执行的请求。可以先标记新请求不可引用，等待引用计数为零，再回收 block；若需要立即释放，应把请求迁移或重新计算。驱逐事件要计入成本和 TTFT，避免平台为了追求内存利用率而造成隐性重算。对敏感数据，TTL 和显式清理优先于命中率。

### 实例重启与 KV 丢失

KV cache 通常是可重建状态，实例重启后不必持久化全部 KV，但必须处理用户体验和容量。正在 streaming 的请求可以失败并返回可重试错误，也可以迁移到新实例并重做 prefill；重做会增加 TTFT 和 GPU 负载。是否迁移取决于输入可重新获得、请求是否包含敏感数据、已输出 token 数和业务 deadline。

如果服务支持 prefix cache 的持久化或跨实例共享，还要处理版本、拓扑、dtype、权限和损坏。跨节点传输 KV 可能比重新 prefill 更快，也可能因为网络拥塞更慢。DistServe 的 prefill/decode 分离说明，状态传输本身是新的系统边界，必须以 workload、网络和 SLO 证明收益 [[4]](#ref-4)。任何共享 cache 都要有校验、过期和安全隔离。

### 四种缓存不能共用一套命中语义

| 层次 | 保存与复用什么 | 命中需要什么 | 主要收益与风险 |
|:---|:---|:---|:---|
| 请求内KV | 当前序列每层K/V | 该请求的有效历史与位置一致 | 避免历史重复计算；总token增长消耗显存 |
| Prefix/Prompt | 已有前缀计算状态 | 完整token前缀、模型及执行配置兼容，授权域允许 | 主要省prefill；失效、侧信道与显存驻留 |
| 精确响应 | 已完成的答案或纯查询结果 | 规范化请求及上下文、权限、策略、知识版本一致 | 可跳过生成；过期事实或随机语义不符 |
| 语义响应 | 相似请求对应的候选答案 | 相似候选还须通过答案等价、新鲜度与权限检查 | 可能覆盖措辞变化；误命中可直接交付错答案 |

vLLM APC 官方说明复用共享前缀的计算，收益集中在 prefill，不能据此承诺 decode 等比例加速 [[13]](#ref-13)。其 V1 设计将前块hash与当前block内容等信息组合，支持在首块加入 `cache_salt` 来限定复用范围 [[14]](#ref-14)。平台应由服务端认证后的租户域生成或约束salt；客户端自己填写某个salt不是访问授权。模型、adapter与多模态内容也须按实现支持的边界进入兼容判断。稳定规则、工具序列和公共参考置于动态问题前有助命中，但不能改变指令优先级，也不能为了稳定前缀保留已撤权内容。商业API的缓存资格、费用和保留时间按各自版本另行核验；本章不把一个自托管实现等同于所有供应商的实现。

GPTCache 是基于语义相似性复用响应的公开实现参考 [[15]](#ref-15)，它不构成业务等价性证明。例如“我可以报销吗”与“我不可以报销吗”高度相似，却可能需要相反回答；同一句问题换主体、政策日期或金额也会改变结果。本书建议先以租户、授权范围、知识与政策版本做硬过滤，再在该域内检索候选；对否定、条件、实体、时间逐项校验，拿近似但答案不同的负例校准阈值。高风险审批、余额和实时库存可以禁用语义响应复用；有副作用的操作不能因缓存命中重新执行。指标需有抽样误命中率、人工投诉、过期命中和每成功任务成本，命中率上升而质量下降必须收紧或停用。

撤权、删除、知识发布和模板升级触发失效事件；删除索引项还应删除载荷和派生缓存，失败时阻止旧项继续命中。精确缓存也不是天然安全：若key遗漏用户授权或上下文，字节相同的文本仍可能对应不同答案。缓存结果绑定来源和生成时间，返回前重新检查当前访问权限。少发无关历史有时比提高命中率更省，压缩改写会损失旧prefix命中，验收仍以相同质量下的总费用与端到端时延为准。

## 10.4 Batching、背压与公平调度

### 从静态 batch 到 iteration-level

传统 batch 等待一组请求同时开始、同时结束；生成长度不同会导致短请求完成后 GPU 空闲。Orca 提出的 iteration-level scheduling 允许请求在每次迭代加入或退出 batch，continuous batching 由此提高了动态 workload 的利用率 [[2]](#ref-2)。但调度器需要在每个迭代点同时处理 prefill、decode、KV 分配、取消、优先级、deadline 和错误。

一次调度决策至少选择：本轮处理哪些请求、每个请求分配多少 token、是否插入新 prefill、是否保留 decode slot、是否驱逐低优先级任务。只按 FIFO 会让长 prompt 阻塞短请求；只按短请求优先会让长任务饥饿；只追求最大 batch 会违反 TTFT。实际策略通常使用 token budget、deadline、优先级、租户配额和公平轮转的组合。

### Prefill 与 decode 的混合调度

Prefill 的计算密集程度高，decode 的迭代频率高。若一个长 prefill 独占 GPU，正在 decode 的请求会停顿；若完全优先 decode，新的请求 TTFT 会变差。调度器可以限制每次 prefill 的 token budget，把长输入切成 chunk，与 decode 交替执行。Sarathi-Serve 的研究说明，chunked prefill 能在吞吐与延迟之间取得更好折中 [[3]](#ref-3)。

chunk size 需要在模型、硬件和 workload 上调优。太小增加 kernel launch 和调度开销，太大仍会阻塞 decode；不同长度请求可能需要不同 chunk。调度器要记录每次 chunk 的等待、执行、抢占和 cache 分配，才能解释 TTFT 和 TPOT 的变化。长请求还应有最大输入、最大输出、总时间和租户 token 预算。

### 公平性与优先级

交互式问答、批量离线、评估和内部开发对延迟的要求不同。可以按 tenant、product、request class 和 priority 建立队列，但高优先级不等于无限资源。每个队列需要并发上限、token budget、最大等待时间和借用规则；否则一个高优先级租户的突发会耗尽所有 KV block。

公平策略还要关注请求的资源大小。按请求数公平会偏向长请求，按 token 公平可能让大量短请求被饿死。可以使用虚拟时间、加权 token 份额或 deadline-aware scheduling，并把实际资源消费写入成本归因。Borg 的优先级、配额和故障域经验说明，公平调度是控制面职责，不应由调用方通过反复重试来争抢 [[16]](#ref-16)。

### 取消、超时与背压

取消需要从网关传到调度器、engine、GPU stream 和 cache manager。重复取消必须幂等，已完成请求不能被覆盖；超时要区分排队超时、prefill 超时、decode 超时和下游发送超时。连接断开不一定意味着任务取消，若产品允许异步完成，服务端应转移状态并释放流式连接。

背压可以发生在多个层面：网关限制并发和请求体，调度器限制 queued token，engine 限制 cache block，结果流限制发送缓冲，计费和观测限制写入速率。资源不足要返回可重试或不可重试的明确错误码，安全拒绝不能被无限重试。重试前必须考虑生成非确定性和工具副作用。

### 批处理与离线推理

离线推理可以使用更大 batch、更长队列等待和更高 GPU 利用率，但不能与在线服务共享一个无界队列。离线任务应有 job id、输入 manifest、输出 manifest、checkpoint 或断点游标、重试和幂等写入。输出写入对象存储时按 shard 提交完成标记，避免下游读取半成品。批任务的失败不能让在线队列被拖慢。

在线与离线共享权重时，权重加载和 cache 仍会相互影响。可以使用独立 GPU 池、低优先级抢占、时间窗口或限制离线 token budget。若必须共享，调度器按 token、deadline 和显存水位做隔离，并在离线任务启动前确认不会影响在线错误预算。离线吞吐高不代表系统成本低，必须计入等待、重试、输出存储和失败数据。

### 连接池与慢客户端

流式服务的网络连接、HTTP/2 stream、WebSocket、发送 buffer 和代理超时都会占用资源。慢客户端如果持续不读取，会阻塞服务端发送并保留 KV；平台需要写超时、发送 buffer 上限、心跳和明确的断开语义。连接数、活跃流、平均发送速度和断开原因应进入指标。

网关和 engine 的超时不能随意叠加。网关 timeout 小于 engine 生成时间会导致客户端断开但 GPU 继续工作；engine timeout 小于工具或下游 timeout 会产生重复重试。请求应携带 deadline，所有组件根据剩余时间决定是否接纳、继续 prefill、停止 decode 或返回 partial result。deadline 需要在 trace 中传播，避免各层使用不一致的本地超时。

### 防饥饿与跨层预算

只有优先级没有最低服务份额时，高优流量持续到达会使低优队列永久等待。可为各等级保留可验证的份额并按等待时间老化；超出容量时返回排队超时，不承诺所有等级都准时完成。出队还受可用 KV、token budget 和下游依赖容量约束。若一次逻辑请求会扇出为多个并行的下游调用，扇出必须由调用方显式限制：只读且彼此独立的调用可以并行，存在依赖或冲突的步骤必须串行；不能把 GPU 的 batch 策略直接当作外部调用的调度算法。

## 10.5 执行优化与解耦

### 副本扩展与模型并行

模型能放进单 GPU 时，副本扩展通常简单且容易隔离故障；模型放不进单卡时，需要 tensor parallel、pipeline parallel 或 expert parallel。TP 把矩阵计算切到多个 GPU，每个 token 需要 collective；PP 把层切成 stage，带来 pipeline bubble；EP 为 MoE 路由 token，带来 all-to-all。并行度越大不一定越快，必须结合拓扑、batch、序列长度和通信占比测量。

路由层需要知道副本健康、模型版本、GPU 可用 KV、队列、拓扑和租户。round-robin 不能处理一个副本正在加载模型、另一个副本 cache 已满的情况。模型并行副本的故障域也不同：同一 TP group 中一张卡故障可能使整个 group 不可用，副本级扩展则可以摘除一个副本。发布和容量控制应以逻辑副本为单位。

### Tensor parallel 的通信边界

TP 适合同机高速互联，因为每层可能需要 all-reduce 或 all-gather。NCCL 根据 GPU、PCIe、NVLink、网卡和网络拓扑选择路径，但服务端仍需记录通信耗时、消息大小和 rank 方差 [[17]](#ref-17)。跨节点 TP 在低并发 decode 里尤其容易把通信延迟暴露到每个 token。若模型已经能通过量化或 offload 放入单节点，减少 TP 可能比扩展到更多节点更适合在线 serving。

### 量化的系统影响

INT8、INT4、FP8 等量化减少权重显存和带宽，但会改变 kernel、校准、激活 outlier、精度和输出质量。量化不是一个简单配置开关：不同层、不同通道、KV cache 和 embedding 可能需要不同精度。TensorRT 的精度与执行路径受构建和硬件约束；TGI 在核验日已进入维护模式，历史功能存在不等于持续新增支持。无论使用哪个 runtime，发布前均需固定质量集和目标 workload [[5]](#ref-5)[[18]](#ref-18)。

量化验收至少包括权重大小、显存峰值、冷启动、TTFT、TPOT、吞吐、长上下文、结构化输出、代码任务和安全拒答。不能只看平均困惑度；小概率 token 错误可能导致工具 schema 或 JSON 失败。量化制品要和校准数据版本、算法、dtype、硬件架构和 engine build 绑定，回滚时同时回滚这些对象。

反量化同样消耗计算与访存：只有节省的权重读取超过反量化和额外kernel代价时，特定负载才可能变快；小batch或不支持的shape可能退化。量化算法与校准原理不在本章重推。

### 编译与 CUDA Graph

Torch compile、CUDA Graph、kernel autotune 和 TensorRT engine 可以降低 launch overhead，但通常要求固定或有限的 shape、内存地址、batch 和控制流。动态 batch、长短请求混合、流式生成和取消会增加图捕获与回退复杂度。固定 shape 的 graph 适合稳定 decode bucket；不规则 prefill 可能需要 eager 或多个 graph。

编译缓存必须包含模型 hash、GPU 架构、driver、CUDA、输入 shape、精度、engine 版本和环境变量。冷启动和热启动要分别测量，避免第一批用户承担编译成本。编译失败应回退到已验证的 runtime，但回退路径需要独立监控，否则性能悄悄下降却只表现为 GPU 利用率变低 [[19]](#ref-19)。

### Draft 与 verify

投机解码使用较小的 draft model 生成候选，再由 target model 一次验证多个 token。若验证接受率较高，target model 可以减少逐 token 迭代；若接受率低，额外的 draft 计算和 KV 管理会抵消收益。Leviathan 等人的方法说明，在保持目标分布等价的条件下，投机解码可以加速生成 [[20]](#ref-20)；speculative sampling 进一步讨论了采样场景的接受规则 [[21]](#ref-21)。

服务端需要为 draft 和 target 管理权重、KV、调度和 GPU。draft 可以与 target 共卡、独立副本或使用 CPU，选择取决于模型大小、接受率、网络和并发。每个请求还要记录候选长度、接受 token 数、拒绝位置、draft 时间、verify 时间和额外显存。低接受率、短输出和高并发场景可能不适合投机解码。

### 与 batching 的交互

投机解码改变了一个 iteration 生成 token 的数量，调度器不能再假设每个请求每轮只追加一个 token。KV block 需要预留候选空间，并在拒绝后回滚未接受 token；多个请求的候选长度不同，batch shape 更不规则。若 engine 没有高效的 block rollback 和 verify kernel，理论加速会被内存与调度开销吃掉。

投机解码也会改变容量指标。输出 token 数没有改变，但 target forward 次数下降、draft GPU 计算上升、内存带宽与通信模式变化。平台应分别报告目标模型调用、draft 调用、接受率、每个成功输出 token 的总 GPU 时间和质量一致性。不能把 target tokens/s 直接与未使用 draft 的 tokens/s 比较。

### 为什么分离

prefill 偏计算，decode 偏内存带宽和长期状态；混合部署时，长 prompt 的突发会抢占 decode，decode 的长尾又会占用 prefill 的资源。DistServe 将二者放到不同资源池，允许分别扩容和优化 goodput [[4]](#ref-4)。分离不是免费的，它增加 KV 传输、请求协调、故障重试、路由和网络容量。

是否分离取决于负载规模和请求分布。小规模、低并发或短输入 workload 可能无法摊平控制面和网络开销；输入长、输出长、SLO 分层明显的服务更可能受益。平台应通过混合部署基线、分离部署、不同网络带宽和故障注入比较真实 goodput，而不是只比较 GPU 数。

### KV 传输协议

prefill 节点完成后，需要传输模型版本、请求 id、token 长度、position 信息、KV dtype、layer layout、block table 和数据校验。接收端确认后才开始 decode；传输中取消时要释放发送和接收 buffer。协议应支持大小、版本、压缩、重试、超时和 partial failure，不应把一个 Python 对象直接跨进程传递。

KV 传输可以使用高速网络、RDMA 或共享内存，但每种路径都有故障模型。网络拥塞时，传输队列会反过来占用 prefill 资源；接收端不可用时，发送端需要限流；序列化与反序列化会消耗 CPU。监控要包括 KV bytes、传输时间、等待时间、重试、丢弃和每个请求的端到端收益。

### 分离后的调度

前端调度器需要同时选择 prefill pool 和 decode pool。选择 prefill 时看计算队列、输入长度、模型版本和网络路径；选择 decode 时看可用 KV、decode slot、输出 deadline 和租户。prefill 完成但 decode 没有空间时，不能无限堆积已产生的 KV；可以延迟 prefill、转回混合池或暂存到受控内存。

这种调度本质上是一个有状态的流量分配问题。只按 GPU 空闲比例路由会忽略 cache 和传输；只按队列长度路由会把长请求集中到一处。路由决策要记录在 trace，便于分析某类请求的 TTFT 变差是因为 prefill、传输、decode 还是队列。

### 权重量化与 KV 量化

权重量化主要减少模型权重占用和读取带宽，KV 量化主要减少长上下文内存。两者误差来源不同：权重量化影响所有层计算，KV 量化影响历史上下文表示。KV 量化可能让更高并发成为可能，却在长对话、检索上下文或代码任务上产生质量变化。评估应按上下文长度、语言、任务和输出长度分层。

量化 scale、zero point、group size、校准数据和累加精度都需要写入制品 manifest。不同硬件可能使用不同 kernel，不能只在一种 GPU 上校验。若量化路径失败，回退到 FP16 可能导致显存不足；因此要预留清晰的降级资源或直接阻止不兼容模型进入该节点池。

### Kernel 融合与 IO

FlashAttention 说明 attention 的性能受 IO 和片上存储影响，FlashAttention-2 进一步优化了并行划分 [[22]](#ref-22)[[23]](#ref-23)。serving 中还要考虑 decode 的小 batch、动态长度和 KV layout，训练时有效的 kernel 不一定适合逐 token decode。性能实验必须包含 prefill、decode、混合 batch、cache 命中和长尾，而不是只测一个 attention kernel。

融合 kernel 可能改变错误传播、精度和调试能力。发布前记录 CUDA、driver、GPU 架构、编译选项和输入 shape；遇到 NaN、非法内存或输出差异时，可以复现构建。engine 应保留 eager 或非融合回退路径，用于诊断和灰度。回退路径的性能差异要有告警，避免所有请求悄悄进入慢模式。

### 静态图和动态请求

CUDA Graph 适合重复执行相同 shape 和内存布局，动态请求需要 padding、bucket 或多图缓存。bucket 太少会产生 padding 浪费，太多会增加编译和缓存，图切换还可能带来同步。应根据真实长度分布选择 bucket，并设置图缓存上限。请求在 graph 中执行时，取消和异常需要安全退出点，不能直接释放仍被 graph 引用的 buffer。

图缓存 key 不仅包括 batch 和 sequence length，还包括 dtype、模型版本、采样路径、量化、adapter 和设备。LoRA adapter 数量多时，为每个 adapter 编译图不可行，可以选择 eager、分组或只对高频 adapter 建图。发布新版本要原子切换图缓存，旧图在 drain 完成后清理。

### PD与AF：两种不同的解耦边界

本节 AF 明确定义为 Attention/FFN disaggregation，即把注意力和前馈/专家模块放到不同资源。它不是“任意异步框架”的简称。MegaScale-Infer 是具体的 MoE 参考：decode 中 attention节点持有注意力权重和KV，expert节点持有专家；各节点内部可以用TP，节点间使用M2N通信，并以ping-pong微批流水掩盖依赖等待 [[24]](#ref-24)。这是该论文的实现范围，不意味着任何模型任意拆层都能获益。

PD改变阶段的放置：prefill完成后迁移已有历史KV，再由decode池迭代；AF改变每层模块的放置，在解码阶段内部仍需反复交换激活。PD解决输入计算与逐token服务互相干扰，不能消除decode内部attention与FFN的资源不均衡，尤其MoE会把batch分散到不同专家。AF允许独立聚合与扩容，但每层增加网络往返和同步，网络拥塞会直接出现在TPOT里。二者可组合，也都可以不采用。

粗略交通量可帮助判断瓶颈：PD的一次输入KV迁移按2L×H_kv×d×S×b字节增长；一个未分片、无专家复制的AF边界，batch中B个token送出与返回的激活粗估2B×hidden_size×b，每层重复。后者只是维度推导，MoE top-k路由、TP/EP分片、共享或重复传输会改变真实字节，不能照搬为论文实际网络量。比较时记录每层延迟、大块带宽、队列、KV驻留、微批大小与故障重试，在同SLO和质量下核算总GPU/网络成本。低负载、低接受率或慢网络时，继续共置可能更合适。

### 大词表选读：先分清投影、归约和采样

词表超过十万时，hidden到vocab的投影、logits写回、softmax归约、采样与跨卡候选合并都可能占时；先profile再选优化，不能断言瓶颈必为softmax。精确路径可用融合、向量化、稳定max/sum归约和词表分片，减少无用logprob落显存。贪心argmax不必显式归一化，但通常仍需投影；top-k/top-p改变采样分布，也不自动省去全部logits。层次或采样softmax常改变训练目标或模型结构，不能无训练地替换现有checkpoint并承诺精确等价。这里只限定这些设计边界，不交付新 kernel 或性能结论。

## 10.6 多模态与长上下文负载

### 长上下文的 admission

长上下文同时增加 prefill 计算、KV 容量、排队时间和输出阶段的内存保留。平台应在入口限制最大输入、最大输出、总 token、并发和租户预算，并根据请求类型选择同步、异步或低优先级队列。截断策略要保留消息边界、系统指令和工具 schema，不能简单从字符串尾部截断。

长上下文请求的公平性可以用 token budget 和 deadline 控制。一个请求占用大量 prefill 时间时，调度器应切块或暂时挂起；一个 decode 很慢的请求应有最大 wall-clock 或输出 token 限制。指标区分 queue、prefill、decode、cache wait 和 network wait，才能知道用户是在等待计算还是等待资源。

### 多模态输入

图像、音频和视频会先转换为视觉或音频 token，输入 token 数量和显存不一定与文本长度线性对应。图片分辨率、视频帧数、压缩、patch 数和视觉 encoder batch 都会改变 prefill 成本；不同模态还可能有不同 cache 可复用边界。serving 接口要记录原始模态大小、预处理版本、视觉 token 数和编码时间，不能只计费文本 token。

多模态模型可能共享语言模型权重，却需要额外的 encoder、预处理 CPU、GPU workspace 和跨模态 cache。调度器应区分 text-only、image、audio 和 video workload，避免大视频请求阻塞短文本请求。模型并行和设备放置要考虑 encoder 与 decoder 的通信，发布门禁要加入图片损坏、超分辨率、超长视频和错误 MIME 类型。

### 多轮对话与上下文压缩

多轮对话会不断增长输入和 KV。平台可以把历史消息重新 tokenize，使用 prefix cache，摘要压缩，检索相关片段或把旧轮次转为外部状态。压缩改变模型可见上下文，不能只视为性能优化；应记录压缩算法、摘要版本、保留消息和 token 预算。安全和权限信息不能在压缩时丢失，工具 schema 和系统指令必须按固定规则保留。

上下文压缩有不同成本：摘要需要额外模型调用，检索需要存储与索引，滑动窗口丢失旧信息，KV 复用占用显存。路由器可以依据任务类型、上下文长度、deadline 和租户预算选择策略。评估要比较压缩前后任务成功、引用正确、工具调用和延迟，而不只是输入 token 减少比例。后续数据与评估章节会把这些策略纳入回归集。

## 10.7 平台路由、隔离与发布

### 模型仓库与实例生命周期

模型实例经历 DOWNLOADING、LOADING、WARMING、READY、DRAINING、UNHEALTHY 和 TERMINATED。加载权重、构建 engine、预热 tokenizer、分配 KV pool 和运行 probe 都属于启动过程；只有 READY 才能接收流量。滚动发布时，新旧实例同时存在，平台需要预留双份权重和 cache，容量模型必须计入这段峰值。

Triton 的模型仓库与版本管理提供了服务编排参考 [[6]](#ref-6)。生产平台还应把 instance id、模型 hash、engine hash、GPU、拓扑、启动耗时、预热结果和健康状态写入注册表。实例被摘除后，应停止接收新请求，等待或迁移正在执行的请求，释放 cache，再回收 GPU。强制终止会造成部分流和重试，必须进入故障指标。

### 路由与降级

路由依据可以包括模型能力、版本、区域、租户、数据驻留、价格、当前队列、KV 水位、GPU 类型和 deadline。fallback 模型不能只满足接口兼容，还要经过任务质量、安全和成本评估。所有路由结果、降级原因和策略版本写入 trace，方便解释一次请求为什么没有使用主模型。

流量突发时可以采用限流、排队、降级采样参数、降低最大输出、切换小模型或转异步。降级顺序应由产品定义，安全策略和工具权限不能因为资源不足而被跳过。对高价值任务，可以保留容量池；对低优先级离线任务，可以使用抢占式资源。Kubernetes 的 GPU 调度只解决资源发现和分配，真正的业务降级仍需平台控制面 [[25]](#ref-25)。

### 灰度、影子与回滚

模型灰度可以按租户、地区、请求类型、流量比例或实验分组进行。影子流量执行新模型但不触发工具副作用，用于比较延迟、长度、格式和质量；正式灰度才承担真实响应。灰度指标包括 HTTP 错误、TTFT、TPOT、p99、OOM、cache 命中、停止原因、结构化输出成功率、工具调用成功率、拒答率和单位成本。

回滚需要同时回滚权重、tokenizer、模板、量化 engine、adapter、路由、cache 版本和安全策略。新版本创建的 prefix cache 不能默认给旧版本使用；旧版本在当前 driver 和 GPU 上也必须仍能启动。发布系统应保留最近可用制品、健康 probe 和回滚演练记录，不应只把一个 tag 改回旧值。

### 租户级资源账本

租户配额应覆盖请求数、输入 token、输出 token、KV block、并发模型、优先级、工具调用、cache 占用和预算。只限制 QPS 无法防止一个租户发送超长上下文；只限制 token 无法防止大量短请求占用连接和 CPU。计量系统要记录 accepted、rejected、queued、generated、cancelled 和 failed 的资源，明确是否计费。

共享权重通常可以只读共享，adapter、KV、prefix cache、日志和工具状态则需要隔离。租户 id 不能只作为日志字段，还要进入 cache key、路由、凭证和访问控制。不同数据地域或合规等级的请求不能因为同一模型副本空闲而跨区域路由。权限检查应发生在 admission 和工具执行前两个阶段。

### 配额借用与公平

固定配额容易产生碎片：某租户空闲时，其他租户却排队。可以允许短期借用空闲 token budget，但设置最大借用、归还优先级和高峰抢回。借用的资源要可追溯，避免月底账单无法解释。高优先级请求可以抢占排队中的低优先级请求，但不应粗暴杀死已经输出大量 token 的 decode，除非产品明确允许。

公平调度还要避免通过拆分请求绕过配额。平台按 request id、parent task、tenant 和 operation 统计，同一逻辑任务的多步调用共享预算。超预算时返回结构化错误，让上层可以结束任务或转人工；不要默默降低安全检查、扩大重试或切换到未经评估的模型。

### Serving runtime 的选择边界

不同 runtime 解决的问题不同。vLLM 是执行和KV管理的工程参考；TGI 是已有服务实现，核验日处于维护模式；TensorRT体现编译与执行引擎层；Triton负责模型管理和服务层 [[10]](#ref-10)[[18]](#ref-18)[[5]](#ref-5)[[6]](#ref-6)。这些对象不在同一层级，不能当四个等价替代项。选型时不能只比较启动命令或单一 benchmark，而要比较动态 batch、长上下文、量化、adapter、流式、工具协议、观测和回滚。

一个实际平台可以分层组合：网关负责鉴权、限流和协议；路由负责模型版本、租户和区域；serving engine 负责执行与 cache；模型仓库负责 artifact；Kubernetes 或其他调度器负责 GPU；观测系统负责指标、日志与 trace。组合的代价是接口和故障边界增多，因此必须定义 request manifest、model manifest、错误码和状态事件。若把多个组件拼在一起却没有统一契约，替换其中任意一个都会改变行为。

### 多进程与多模型共存

同一节点部署多个小模型可以提高利用率，但会引入权重竞争、KV 竞争、编译缓存竞争和尾延迟相互影响。多进程隔离较强，却可能重复加载 CUDA context 和权重；单进程多模型共享资源效率高，却需要更复杂的模型选择和内存回收。模型路由必须把可用 cache 和切换成本纳入决策，不能只看当前请求数。

模型切换包括加载权重、初始化 kernel、分配 KV pool、预热和释放旧模型。如果加载期间占用同一 GPU，会造成在线请求抖动；如果使用双份显存，又需要容量余量。可以把冷门模型放到独立池，或采用异步加载和模型驻留策略。每个模型要有最小驻留时间和最大空闲时间，避免流量抖动导致反复加载。

### Adapter 与 LoRA serving

一个基础模型可能服务多个 LoRA adapter。共享 base weight 能降低显存，但每个请求需要选择 adapter、加载或缓存 adapter，并确保 batch 中不同 adapter 的执行路径正确。adapter cache 有容量、淘汰、租户隔离和版本兼容问题；将不同 adapter 合并进同一个 batch 可能增加 kernel 或权重访问开销。

adapter manifest 应包含 base model hash、adapter hash、训练数据等级、rank、dtype、目标层和安全标签。base model 升级后旧 adapter 不一定兼容，不能只按 adapter 名称加载。灰度时同时比较 base-only、adapter 和 fallback 的质量、延迟、显存和工具成功率。adapter 的下载权限与缓存权限也必须分开，避免低权限租户读取其他租户的制品。

### 生成质量与系统回退

系统回退不仅是换一台 GPU 或换一个副本，也可能改变模型、精度、采样、上下文、工具和安全策略。每种回退路径都需要能力声明和质量基线，例如小模型是否支持 JSON、视觉输入、工具 schema 和最大上下文。路由器根据请求能力选择兼容 fallback，不能把任意文本模型作为通用后备。

回退事件要对调用方透明且可追踪。响应中可以返回 model revision、degraded flag 或 warning，trace 中记录原因和候选模型。对安全敏感请求，若没有满足策略的 fallback，应快速失败并引导人工，而不是使用未经审核的模型。质量回退、延迟回退和成本回退应分别计入指标，便于决定是否扩大容量或修复主路径。

### 版本兼容矩阵

生产平台应维护模型版本、tokenizer、template、engine、量化、adapter、GPU 架构和驱动的兼容矩阵。矩阵不是静态文档，而是发布控制器可以读取的约束。例如某 adapter 只兼容特定 base hash，某 engine 只支持特定 compute capability，某 cache 只支持某种 KV dtype，某模板版本改变了 special token。加载前做校验可以把复杂错误提前到部署阶段。

兼容矩阵还要包括 API 和质量能力。一个小模型可能支持文本和 JSON，却不支持图像或工具；一个降级模型可能没有同样的上下文长度；一个量化版本可能在代码任务上低于质量门限。调用方通过能力声明选择模型，路由器在请求前验证能力，不应等到生成中途才失败。发布登记时记录矩阵版本，回滚时恢复相同矩阵。

### 数据驻留与跨区域路由

多区域服务需要在路由前判断数据驻留、租户合规、模型制品可用、网络延迟和容量。把请求路由到更近区域可以降低 TTFT，却可能违反数据区域；把请求路由到有空闲 GPU 的区域可以提高 goodput，却增加跨区域传输和故障面。请求 manifest 应携带 region、data class 和允许的 fallback 区域，路由决策写入审计。

跨区域故障时，系统应区分可迁移和不可迁移请求。纯文本、无敏感数据且没有副作用的请求可以重试到备用区域；包含敏感上下文或外部工具状态的请求可能只能失败或转人工。区域间共享 prefix cache 和 KV 通常有更高风险，默认不共享，除非具备加密、权限、版本和删除保证。

### 推理服务的安全降级

资源压力不应导致安全校验被跳过。限流、降级模型、截断和异步化都必须经过相同的输入安全、输出安全和工具权限策略。安全服务本身不可用时，平台应选择拒绝、有限能力或人工，而不是默默放行。安全决策的版本、结果、超时和 fallback 写入 trace，敏感内容只保留受控摘要。

结构化输出和工具执行尤其需要安全边界。模型生成的 JSON 只是未可信输入，必须做 schema、权限、参数范围和目标资源校验；工具返回的内容也可能进入下一轮上下文，需要标记来源和可信级别。推理 Infra 负责提供执行前的拦截点、审计和幂等，不能把所有安全责任推给 prompt。

### 运行时升级的逐级验证

runtime 升级可分为离线兼容、单实例 canary、影子流量、小比例正式流量和全量。离线阶段验证加载、输出、量化、最大长度和错误语义；单实例阶段验证冷启动、显存、cache、batch 和 drain；影子阶段比较真实长度与路由；正式灰度阶段观察 SLO、质量、成本和租户差异。每一步都有停止条件和回滚版本。

如果新版本只在特定 GPU 或特定长度变差，整体平均可能看不出来。因此灰度指标按模型、硬件、长度桶、租户、输入语言、请求类型和 adapter 分层。监控系统要避免标签爆炸，但可以把高维分层放在 trace 和离线聚合中。发布控制器依据分层门禁，而不是单一全局平均。

### 网关的尝试、额度与部分输出

公共协议统一消息与错误外形，能力矩阵仍要显式登记结构化输出、工具消息、窗口、媒体、拒绝语义和区域许可。网关数据面鉴权、准入、适配和转发；控制面维护凭据及版本化策略，凭据不能进入模型可见prompt。配额存储不可用时，只有持有未过期、可审计本地额度且策略允许的请求可继续；需要强一致金额或敏感校验的请求应拒绝或排队。配置快照有有效期，不设置无保护全量旁路。

一个逻辑request可以有多个attempt。流式前原子检查并预留输入、最大输出和必要下游预算，同账户并发预留必须同时扣除在途金额；每次预留、扣减和最终退款都有稳定ID并幂等。完成后按usage结算，断流未知用量保留待对账，不能记成零费用。诊断费用翻倍时，按任务版本比较轮次、输入输出、cached read/write、Embedding、Rerank、工具、存储和judge用量，再看重试、路由和失败任务；QPS不变不代表工作量不变。

主模型未输出时可在预算、能力及数据策略允许下切换；已输出一半时，旧结果标记中断，备用结果作为新版本或独立结果交付，不能无缝拼接互相冲突的答案；工具已经执行时沿用原业务operation核查事实，不重新发明副作用。超预算停止新增步骤并交付明确partial状态，不能把截断JSON当成功。缓存、压缩和批处理优化都需要质量门禁，目标是每成功任务成本。

## 10.8 性能诊断与事故

本节处理“已经出问题”的情形：怎么确认瓶颈在哪一层，怎么判断一次降级是否安全，以及怎样把事故变成可复现的证据。验收标准放在下一节，两者分开，是为了避免在排障时被门禁条款分散注意力。

### 推理 SLO

服务 SLO 应按产品场景分层。交互式请求关心可用性、TTFT、TPOT、p99 和流式完成率；批量任务关心完成时间、goodput 和单位成本；多步任务关心最终成功、中间状态一致性和总时延。HTTP 200 不代表生成成功，部分流、错误 JSON、工具参数不完整和安全策略失败都应有明确结果状态。

错误预算可以决定是否继续灰度、是否暂停性能实验、是否保留更多容量。SLO 还需要质量门禁：量化或 batching 如果提高吞吐却让代码任务、结构化输出或安全指标回退，不能算达标。SRE 方法提供了把可靠性与发布决策连接起来的框架 [[8]](#ref-8)。

### Metrics、Logs 与 Traces

Metrics 记录吞吐、队列、TTFT、TPOT、KV 使用、block 分配、cache 命中、GPU 显存、OOM、取消、重试和路由；logs 记录具体错误、模型 hash、engine、节点和状态；traces 连接网关、tokenizer、调度器、prefill、decode、工具和计费。Dapper 的追踪思想和 OpenTelemetry 的语义约定适合把一次请求跨组件串起来 [[26]](#ref-26)[[11]](#ref-11)。

默认不记录完整 prompt 和输出。使用 prompt hash、token 数、长度分桶、采样摘要和脱敏后的错误上下文，必要时通过受控诊断流程临时提升采样。Prometheus 指标标签要避免完整 request id、租户自由字符串和 token 内容造成高基数；详细信息放结构化日志或 trace event [[12]](#ref-12)。

### 成本归因

在线成本不应只按 GPU 小时计费。一次请求的成本包括权重占用、prefill、decode、draft、KV 保留、cache miss、网络、CPU tokenizer、工具调用和重试。可以按输入 token、输出 token、GPU 时间和成功任务组合归因，并区分模型版本、租户、项目和业务场景。prefix cache 命中需要记录节省的计算，同时也要计入 cache 内存和失效成本。

成本优化要和质量、SLO 同时看。降低最大输出 token 可以降低成本，却可能降低任务成功；激进 batching 可以提高平均吞吐，却增加长尾；更低精度可以节省显存，却增加重试和人工审核。因此，系统优化不能脱离输出质量和任务目标，任务成功率与资源报告必须并列。

### 典型故障

常见事故包括模型加载失败、权重损坏、tokenizer 不匹配、量化 kernel crash、KV OOM、GPU Xid、NCCL hang、节点网络故障、对象存储不可用、流式连接堆积和慢客户端。每种故障应定义检测信号、影响范围、自动动作、人工动作、恢复时间和是否丢失请求。把所有故障都归类为 5xx，会让容量、质量和数据泄露问题无法区分。

### 降级与恢复

实例故障时先摘除健康检查失败的副本，停止新请求并 drain；网关将新请求路由到兼容副本或异步队列。正在 decode 的请求可以失败重试、迁移重算或继续等待，决策取决于 deadline 和副作用。模型主版本不可用时，fallback 需要保持接口、安全、工具权限和质量边界。降级事件进入错误预算与事故报告。

恢复后要检查 cache 是否释放、队列是否回落、p99 是否恢复、GPU 是否降频、工具状态是否一致、重复请求是否产生副作用。仅看到 Pod Running 不代表服务恢复。Dapper/OpenTelemetry trace 能帮助确认故障影响是否从网关传播到 engine 和下游 [[26]](#ref-26)[[11]](#ref-11)。

### 演练设计

每月可以执行一次小规模节点故障、一次 KV OOM、一次模型回滚和一次流式断开演练。注入动作应可撤销，先在影子流量或隔离租户验证，再进入生产候选池。演练结束保存请求样本摘要、指标、trace、资源释放、恢复时间和人工操作。失败的演练要转为发布门禁，而不是只写在复盘文档中。

### 压测模型

推理压测应使用真实长度分布而不是固定 prompt。至少构造短输入短输出、长输入短输出、短输入长输出、长输入长输出、prefix 高命中、prefix 低命中、结构化输出、工具调用和突发流量八类 workload。每类记录 arrival process、并发、deadline、取消比例和租户混合，并分别比较混合部署与 prefill/decode 分离。

压测工具要区分客户端等待、网关排队、调度等待、prefill、decode 和网络发送。若只从客户端测总延迟，无法知道瓶颈；若只测 engine，不包含真实连接和 tokenization，又会高估服务能力。压测完成后保留配置、模型制品、原始事件和环境，保证下一次升级能做回归，而不是重新猜测基线。

### 推理平台的状态机回放

事故排查最有效的方法之一是回放一个 request 的状态机：何时被接收，排队多久，何时分配 KV，经历几次 prefill chunk，生成多少 token，是否发生 cache 命中、迁移、重试、取消和回退。回放数据来自 request event、metrics 和 trace，必须使用同一个 request id 和单调序列。没有状态回放，工程师只能从多组件日志按时间猜测，容易把因果顺序弄错。

状态回放也可用于测试。给定一组事件，验证重复取消不会泄漏 block，失败迁移不会重复工具，实例 drain 会停止新请求，cache 版本不兼容会被拒绝，deadline 到期会释放资源。把这些规则写成状态机测试，比只做 HTTP happy path 更能覆盖 serving 的真实复杂度。

### 可回放的OOM教学调查

下面是故意构造的诊断事件，不是GPU实测日志。固定模型revision、单卡进程、BF16 KV、请求长度和并发；让多轮对话逐轮追加历史，记录每轮输入/输出token、active KV block、cache条目、allocated、reserved、设备总占用和安全点后的值。PyTorch 2.9把 `memory_allocated` 定义为张量占用，把 `memory_reserved` 定义为分配器管理总量；unused缓存仍可能显示在设备工具中，`empty_cache()`只释放未使用缓存，不释放仍被引用的活跃张量 [[19]](#ref-19)。其他引擎自建KV池可能不完整反映在PyTorch指标内，应再对照engine账本和设备级占用。

| 教学事件顺序 | 观察字段 | 假设检验与下一步 |
|:---|:---|:---|
| A：总token逐轮增长 | 活跃KV随之增长，结束后回到有界基线 | 正常历史成本，不能直接判泄漏；限制总token或压缩 |
| B：固定总token重复同一批 | 完成后allocated仍逐轮增，cache条目不增 | 查输出/logits列表、回调引用、autograd图及后台任务 |
| C：allocated稳定而reserved阶梯增 | snapshot中的分块、空闲区、分配失败大小 | 可能合法池扩张或碎片；比较稳定shape，不能以差值直接量化碎片 |
| D：取消率提高 | active owner消失但pending-free堆积 | 查stream事件、block引用和deferred free是否推进 |
| E：首次非法访存后报OOM | 最早CUDA错误、请求shape、kernel版本 | 先定位越界或异步错误，后面的异常可能只是传播 |

分配器reserved包含allocated及池内空间，不能把二者都加进总显存账本；reserved−allocated中有可复用空闲块，也可能有不可满足当前申请的碎片，并非全部无用。预分配KV池、prefix与CUDA Graph驻留都可能合法，关键是能否对应到有owner、容量和回收规则的对象。检查取消后先停止新引用，在实际stream安全点后回收，等待后台释放完成再采样；过早释放会造成随机崩溃，永久等待会造成泄漏 [[27]](#ref-27)。

可复现步骤是先建立固定长度/并发基线，再一次只增加轮次、取消、prefix或图捕获中的一个变量，保存显存快照、最早错误、事件序列与版本。若固定负载在缓存饱和后仍无限增长，追引用链和持有张量；若只有形状变化触发，比较allocator统计和padding/graph bucket；若长prefill触发峰值，调小chunk及准入预算。清cache只能作为受控对照，不能把循环empty_cache当修复。最终验收核对请求终态、待回收block归零、合法缓存有界、重跑相同负载无新增增长和质量回归；本章仅给方法，未声称这些条件已在GPU上通过。

### 性能剖析要形成因果假设

Nsight Compute提供Roofline和内核资源分析方法 [[28]](#ref-28)。先用请求时间线确认瓶颈属于CPU预处理、排队、H2D、prefill、decode、collective还是发送，再选相应工具。算术强度低且带宽接近上限时考虑KV布局和权重读量；算力低且带宽也低，检查launch、同步、形状与依赖空洞，不能直接归因为“GPU太弱”。只在单kernel好看而端到端变差时，核对融合寄存器压力、padding、编译冷启动及调度干扰。报告固定shape、硬件和版本，将质量、尾延迟和成功任务成本一起对照。

## 10.9 交付验收与运行手册

验收回答的是“上线前要证明什么”。与诊断不同，验收要求每一项结论都绑定模型制品、硬件、engine 与负载画像，并且能被独立重跑；只报告一个漂亮的平均吞吐不构成验收证据。

### 性能回归门禁

回归门禁可以设为多目标约束：TTFT p95 不超过基线某个比例，TPOT p95 不超过阈值，goodput 不下降，显存峰值不超过容量，质量指标不低于容差，错误和取消率不增加。不同 workload 可以有不同权重；不能用一个总体平均掩盖长上下文或高价值任务退化。

当新版本提升平均吞吐但降低 p99，应根据业务 SLO 判断是否接受；当量化降低成本但工具成功率下降，应检查单位成功任务成本；当 prefix cache 提升命中但隐私风险增加，应优先修复隔离。门禁结果要附带解释和失败样本，发布控制器只依据明确策略自动决策，复杂回退进入人工评审。

### Engine 的兼容性测试

升级 runtime 时，功能测试要覆盖单请求、动态 batch、流式、取消、最大长度、结构化输出、工具调用和多模态；性能测试要覆盖不同输入输出长度和并发；资源测试要覆盖显存、KV block、编译缓存和冷启动；故障测试要覆盖 engine crash、GPU reset、模型加载失败和滚动 drain。测试结果绑定 runtime、模型、driver 和硬件 hash。

兼容性还包括错误语义。某个版本把超长输入返回 400，另一个版本可能截断后返回 200；某个版本把客户端断开视为取消，另一个版本继续生成；某个版本在 tool call 失败时返回 partial output，另一个版本返回 error。上游服务依赖这些差异，不能只用“文本看起来一样”判断兼容。API contract test 和状态机 test 是 serving 平台的必要组成。

### 请求迁移

请求迁移可以发生在实例 drain、节点故障、负载均衡或 prefill/decode 分离时。迁移分为重新发送完整输入、传输已有 KV、只迁移请求状态和放弃当前请求。重新 prefill 简单但延迟高；传输 KV 需要网络和版本兼容；只迁移状态而不迁移 cache 会导致状态与资源不一致。迁移策略必须声明可接受的重复计算和输出连续性。

流式请求迁移更复杂。旧实例可能已经发送部分 token，新实例继续生成时必须保持上下文、采样随机状态和停止条件。若产品只保证最终文本，不保证 token 级连续，可以在事件中标记重连和重算；若要求严格连续，需要持久化更多 decode 状态。大多数系统应优先让迁移发生在 prefill 完成前或请求边界，避免把所有状态都做成可迁移。

### 功能验收

固定验证集需要覆盖普通生成、空输入、超长输入、最大输出、停止词、流式、取消、重复请求、结构化 JSON、工具调用、多模态错误和非法参数。每个场景验证状态迁移、错误码、资源释放和审计字段。对 streaming，要测试客户端中途断开、服务端重启、重连和部分结果；对工具，要测试超时、重试、重复提交和副作用幂等。

### 性能与容量验收

性能测试固定模型、硬件、engine、tokenizer、输入输出分布和并发曲线，分别报告 cold start、warm start、TTFT、TPOT、E2E、p50/p95/p99、goodput、显存峰值、KV 水位、cache 命中和 GPU 利用率。容量测试加入长尾长度、突发流量、节点故障和滚动发布，验证 failover reserve 与 deployment reserve 是否真实存在。

### 质量、可靠性与治理验收

质量测试比较原始精度、量化、kernel、batching、投机解码和降级模型的任务指标；可靠性测试注入 worker OOM、节点故障、网络抖动、模型加载失败、cache 损坏、对象存储不可用和流式断开；治理测试检查权限、租户隔离、数据驻留、脱敏、审计、成本和回滚。所有结果绑定 artifact hash 和报告版本。

### 交付后的运行手册

推理服务需要一份可执行 runbook

### 验收证据的最小集合

把上面几类验收合起来看，一份可交付的推理平台至少要留下五组证据，缺任何一组都意味着某个结论无法复核。

| 证据组 | 回答的问题 | 关键字段 |
|:---|:---|:---|
| 制品与兼容 | 这个版本能在目标环境跑起来吗 | manifest hash、tokenizer/模板版本、engine build、GPU 架构、驱动 |
| 容量与性能 | 在目标负载下能达到什么指标 | 长度分布、并发曲线、TTFT/TPOT 分位数、显存与 KV 水位 |
| 质量回归 | 系统优化有没有伤害输出 | 原始精度与优化后精度的配对比较、结构化输出合法率、关键任务切片 |
| 可靠性与降级 | 故障时行为是否可预期 | 故障注入记录、摘除与恢复时间、降级路径的能力声明、丢失请求数 |
| 治理与成本 | 谁能用、花了多少、能否回滚 | 租户配额、数据驻留、审计字段、单位成功任务成本、回滚演练记录 |

五组证据之间不能互相替代。一个版本可能容量漂亮但质量回退，可能质量持平却无法回滚，也可能全部达标但成本翻倍。验收的产物不是一句"通过"，而是这五组证据连同它们的环境与版本一起归档——下一次升级时，这份归档就是唯一可比的基线。

需要强调的是，本章给出的所有验收表格与阈值示例都是本书建议，不是从某个生产平台测得的结论。真实阈值必须在目标硬件和目标负载上重新测量，并且随模型、量化格式和 engine 版本变化重新确定。：如何判断是队列、prefill、decode、KV、网络还是客户端；如何摘除实例；如何清理 cache；如何回滚 engine；如何切换 fallback；如何处理流式部分结果；如何核对工具副作用；如何恢复成本与容量。每条命令或操作都应有权限、预期指标、风险和回滚，避免事故中临时尝试造成二次损害。

runbook 要通过演练更新。每次事故或压测发现新的边界，就增加诊断字段、告警或自动化；如果一个故障需要人工在机器上删除目录才能恢复，说明平台缺少状态清理接口。后端团队熟悉的健康检查、租约、熔断、重试和数据迁移经验在这里都适用，但需要把 GPU、KV 和生成质量纳入考虑。

### 显存账本练习卡（Q275）

输入是 10.2 节列出的 Qwen3-8B 配置字段、检查点 tensor 清单、权重/KV精度、单卡容量、block大小、并发和每条输入/最大输出长度；工作区与安全余量必须显式给定，缺失项列为未知，不猜成零。输出包括GB/GiB双单位权重、每token及总KV、逐rank驻留账本、候选并发上限、主导项与OOM风险解释，并保存manifest和计算假设。

边界检查包含GQA的KV头与query头区别、空序列、尾块取整、beam复制、INT4元数据、权重与KV不同精度、共享prefix的引用、TP不均匀分片及prefill峰值；再用固定长度重复、正常多轮增长、取消延迟释放三个事件轨迹判断是否泄漏。验收标准是账本各项不重计、单位正确、未知项显式且不把静态上限当SLO承诺。交付状态为练习设计，未编写实验程序、未运行GPU、未宣称实测通过。

## 10.10 本章小结

推理服务的工程边界可以归纳为四条：第一，模型制品、tokenizer、模板、engine 和安全策略必须版本化；第二，prefill、decode、KV、batch 和网络必须作为资源被测量和调度；第三，取消、重试、流式和工具调用必须有明确状态与幂等语义；第四，SLO、质量、成本、隔离、灰度和回滚必须形成同一条证据链。只优化某个 kernel 或某个吞吐数字，都不足以证明服务成熟。

对于后端工程师，最值得迁移的是状态机、租约、背压、幂等、容量模型和故障演练；最需要补齐的是 GPU 内存、KV layout、collective、kernel 与生成质量之间的联系。真正生产级的 serving engine 不只是把模型 forward 封装为 RPC，而是管理一组动态请求、共享中间状态和有限硬件资源，并在任何版本、流量和故障变化下保持可解释。

请求能够稳定在线运行之后，下一个问题是“如何知道模型是否仍然正确”：数据治理、离线评估、回归集、在线反馈和质量门禁会与本章的 trace、usage、版本和成本数据连接，形成从数据到生产的评估闭环。

## 10.11 参考资料

<a id="ref-1"></a>[1] Kwon, W., et al. *Efficient Memory Management for Large Language Model Serving with PagedAttention*. SOSP, 2023. https://arxiv.org/abs/2309.06180 访问日期：2026-10-17

<a id="ref-2"></a>[2] Yu, G.-I., et al. *Orca: A Distributed Serving System for Transformer-Based Generative Models*. OSDI, 2022. https://www.usenix.org/conference/osdi22/presentation/yu 访问日期：2026-10-17

<a id="ref-3"></a>[3] Agrawal, A., et al. *Taming Throughput-Latency Tradeoff in LLM Inference with Sarathi-Serve*. OSDI, 2024. https://www.usenix.org/conference/osdi24/presentation/agrawal 访问日期：2026-10-17

<a id="ref-4"></a>[4] Zhong, Y., et al. *DistServe: Disaggregating Prefill and Decoding for Goodput-optimized Large Language Model Serving*. OSDI, 2024. https://arxiv.org/abs/2401.09670 核验日期：2026-10-17

<a id="ref-5"></a>[5] NVIDIA. *TensorRT Architecture Overview*. online checked 2026-10-17. https://docs.nvidia.com/deeplearning/tensorrt/latest/architecture/architecture-overview.html 核验日期：2026-10-17

<a id="ref-6"></a>[6] NVIDIA. *Triton Inference Server: Model Management*. 在线文档，核验2026-10-17. https://docs.nvidia.com/deeplearning/triton-inference-server/user-guide/docs/user_guide/model_management.html

<a id="ref-7"></a>[7] 《动手学深度学习》作者. *11.7 Transformer*. 中文在线2.0.0. https://zh.d2l.ai/chapter_attention-mechanisms-and-transformers/transformer.html 核验2026-10-17

<a id="ref-8"></a>[8] Beyer, B., et al. *The Site Reliability Workbook: Implementing SLOs*. O’Reilly, 2018. https://sre.google/workbook/implementing-slos/ 核验日期：2026-10-17

<a id="ref-9"></a>[9] Qwen Team. *Qwen3-8B config.json*. 2026-10-17所见文件提交47719a2. https://huggingface.co/Qwen/Qwen3-8B/blob/main/config.json

<a id="ref-10"></a>[10] vLLM Team. *vLLM: A High-Throughput and Memory-Efficient Inference Engine*. https://github.com/vllm-project/vllm 访问日期：2026-10-17

<a id="ref-11"></a>[11] OpenTelemetry Authors. *Signals*. 在线文档，核验2026-10-17. https://opentelemetry.io/docs/concepts/signals/

<a id="ref-12"></a>[12] Prometheus Authors. *Prometheus Documentation*. https://prometheus.io/docs/introduction/overview/ 访问日期：2026-10-17

<a id="ref-13"></a>[13] vLLM Team. *Automatic Prefix Caching*. 官方用户指南，核验2026-10-17. https://docs.vllm.ai/en/latest/features/automatic_prefix_caching/

<a id="ref-14"></a>[14] vLLM Team. *Automatic Prefix Caching: Design*. V1设计，核验2026-10-17. https://docs.vllm.ai/en/latest/design/prefix_caching/

<a id="ref-15"></a>[15] Bang, F. *GPTCache: An Open-Source Semantic Cache for LLM Applications Enabling Faster Answers and Cost Savings*. NLP-OSS, 2023, pp.212–218. https://aclanthology.org/2023.nlposs-1.24/

<a id="ref-16"></a>[16] Verma, A., et al. *Large-scale Cluster Management at Google with Borg*. EuroSys, 2015. https://research.google/pubs/large-scale-cluster-management-at-google-with-borg/ 访问日期：2026-10-17

<a id="ref-17"></a>[17] NVIDIA. *NCCL: Collective Operations*. 2.32.3文档，核验2026-10-17. https://docs.nvidia.com/deeplearning/nccl/user-guide/docs/usage/collectives.html

<a id="ref-18"></a>[18] Hugging Face. *Text Generation Inference*. https://github.com/huggingface/text-generation-inference 访问日期：2026-10-17

<a id="ref-19"></a>[19] PyTorch. *CUDA Semantics*. 2.9. https://docs.pytorch.org/docs/2.9/notes/cuda.html 核验2026-10-17

<a id="ref-20"></a>[20] Leviathan, Y., Kalman, M., & Matias, Y. *Fast Inference from Transformers via Speculative Decoding*. ICML, 2023. https://arxiv.org/abs/2211.17192 访问日期：2026-10-17

<a id="ref-21"></a>[21] Chen, C., et al. *Accelerating Large Language Model Decoding with Speculative Sampling*. 2023. https://arxiv.org/abs/2302.01318 访问日期：2026-10-17

<a id="ref-22"></a>[22] Dao, T., et al. *FlashAttention*. NeurIPS, 2022. https://arxiv.org/abs/2205.14135 访问日期：2026-10-17

<a id="ref-23"></a>[23] Dao, T. *FlashAttention-2*. ICLR, 2024. https://arxiv.org/abs/2307.08691 访问日期：2026-10-17

<a id="ref-24"></a>[24] Zhu, R., et al. *MegaScale-Infer: Serving Mixture-of-Experts at Scale with Disaggregated Expert Parallelism*. arXiv, 2025. https://arxiv.org/abs/2504.02263

<a id="ref-25"></a>[25] Kubernetes. *Schedule GPUs*. https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/ 访问日期：2026-10-17

<a id="ref-26"></a>[26] Sigelman, B. H., et al. *Dapper, a Large-Scale Distributed Systems Tracing Infrastructure*. Google Research, 2010. https://research.google/pubs/dapper-a-large-scale-distributed-systems-tracing-infrastructure/ 访问日期：2026-10-17

<a id="ref-27"></a>[27] NVIDIA. *CUDA Programming Guide: Asynchronous Execution*. 2026-09-10. https://docs.nvidia.com/cuda/cuda-programming-guide/02-basics/asynchronous-execution.html 核验日期：2026-10-17

<a id="ref-28"></a>[28] NVIDIA. *Nsight Compute Profiling Guide*. online checked 2026-10-17. https://docs.nvidia.com/nsight-compute/ProfilingGuide/index.html 核验日期：2026-10-17
