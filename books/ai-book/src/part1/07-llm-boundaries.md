# 第7章 AI Infra 全景

## AI Infra 的全局地图是什么？

前六章讨论了模型如何表示、学习、对齐、推理和适配。本章开始进入大模型 Infra：同一个算法模型，怎样被稳定地训练出来、保存下来、部署出去、服务给用户，并在故障、扩容、升级和成本约束下持续运行。

大模型 Infra 不是“把模型放进一台 GPU 服务器”。它是一条跨越数据、计算、通信、存储、调度、服务、观测和治理的系统链路。训练侧关心有效 token/s、扩展效率、故障恢复和 checkpoint；推理侧关心 TTFT、TPOT、吞吐、KV cache、排队和单位成功成本；平台侧还要处理模型版本、容量预测、租户隔离、权限、审计和回滚。

本章作为 Infra 六章的总览，先建立资源模型和系统边界，后续章节再分别深入训练平台、推理服务、数据与评估、分布式调度以及可靠性与成本治理。算法章节给出的模型结构、并行需求、精度和验证器，是 Infra 的输入契约；Infra 的工作是让这些算法在真实约束下可重复、可观测、可恢复地执行。

## 7.1 为什么大模型需要独立的 Infra 方法论

### 从模型文件到生命周期系统

传统后端服务通常把二进制、配置和数据库迁移打包发布。大模型服务的“可运行单元”更复杂：除了权重，还包括 tokenizer、chat template、量化配置、LoRA adapter、视觉预处理、采样参数、工具 schema、验证器、硬件 kernel 和安全策略。训练得到的 checkpoint 只是生命周期中的一个状态，不是可以直接交付的产品。

可以把大模型生命周期表示为：

```text
数据源
  -> 数据清洗与切分
  -> 训练 / 后训练
  -> checkpoint 与评估
  -> 模型注册与转换
  -> 量化 / 编译 / 打包
  -> serving 部署
  -> 线上观测与灰度
  -> 反馈、回滚与再训练
```

每个箭头都是接口。数据 manifest 不稳定，训练不可复现；checkpoint 不完整，故障无法恢复；模型转换不一致，线上质量会变；服务没有容量指标，流量一升高就排队；没有 trace，模型失败无法归因；没有回滚，任何升级都可能变成事故。

### 算法与 Infra 的边界

算法层主要回答：模型结构是什么，损失如何定义，数据如何采样，推理策略如何选择，验证器如何判定。Infra 层主要回答：这些计算如何分布到 GPU，如何高效读写和通信，如何在多请求下调度，如何保存和恢复状态，如何对外提供稳定接口。

边界不是绝对的。GQA 会改变 KV cache 大小，FlashAttention 同时涉及算法和 kernel，MoE 路由会影响通信，RLVR 的 verifier 会影响 serving 资源，推理预算会改变排队和成本。因此设计时不能把算法指标和系统指标分开到互不沟通，而要把它们连接成一条可测量的链路。

| 算法输入 | Infra 需要回答的问题 | 最终验收 |
|:---|:---|:---|
| 参数量、层数、hidden size | 单卡能否容纳，如何并行和分片 | 显存水位、扩展效率 |
| 上下文长度、KV heads | 每请求 cache 需要多少内存 | 并发、TTFT、TPOT |
| 训练 token、batch、序列长度 | 数据和 GPU 是否持续供给 | 有效 token/s、MFU |
| checkpoint、optimizer state | 保存、恢复和跨版本是否可靠 | 恢复时间、丢失工作量 |
| 量化、LoRA、视觉输入 | 转换和 kernel 是否兼容 | 质量、吞吐、回滚 |
| 验证器、工具和多步推理 | 如何调度额外计算 | 单位成功成本、尾延迟 |

### 为什么“能跑”远远不够

一个训练脚本能在 8 张 GPU 上跑通，只能说明功能闭环成立，不能说明系统可扩展。可能存在 GPU 等待、数据加载不足、通信瓶颈、checkpoint 阻塞、故障后无法恢复和成本不可接受等问题。一个推理服务返回了 token，也不代表可上线：它可能首 token 很慢、并发一高就 OOM、长请求饿死短请求、模型升级无法灰度、工具请求没有审计。

大模型 Infra 的验收必须从“功能正确”扩展到五个维度：性能、容量、可靠性、可观测性和治理。Google SRE 方法论把服务目标、错误预算和发布决策联系起来 [22]；大模型系统还要把 token、上下文、模型版本和生成质量纳入同一套服务目标。

## 7.2 工作负载与资源模型：先算清楚再选组件

### 训练、推理和数据处理是三种负载

训练是长时间、稳定、高吞吐的批处理负载。它可以容忍单请求高延迟，却不能容忍 GPU 长时间空闲；需要保存中间状态，发生节点故障后要恢复。推理是在线或准在线负载，输入和输出长度差异巨大，重点是延迟、并发、容量和质量稳定。数据处理是 I/O、CPU、网络和存储混合负载，容易成为训练前端的隐形瓶颈。

三者的资源画像不同：

| 负载 | 主要资源 | 主要瓶颈 | 关键指标 |
|:---|:---|:---|:---|
| 预训练 | GPU、网络、对象存储 | 通信、数据供给、故障恢复 | token/s、MFU、扩展效率 |
| 后训练 | GPU、CPU、评估环境 | rollout、验证器、数据混合 | 成功率、样本吞吐、成本 |
| 在线推理 | GPU 显存、带宽、调度 | KV cache、排队、batch 形状 | TTFT、TPOT、p95、goodput |
| 数据处理 | CPU、网络、对象存储 | 解码、去重、shuffle、序列化 | 有效 token/s、I/O 利用率 |
| 评估回放 | GPU、工具沙箱 | 并发执行和结果判定 | case/s、通过率、回归时间 |

如果用在线 serving 的 QPS 指标衡量训练，就会忽略训练 GPU 是否被通信和数据拖慢；如果只用训练吞吐选推理 GPU，也会忽略单 token 延迟和显存容量。平台设计首先要建立 workload taxonomy，再决定共享集群还是隔离资源池。

### 参数、激活和状态的内存账本

训练显存不只有模型权重。粗略账本包括参数、梯度、优化器状态、激活、通信 buffer、临时 workspace 和 checkpoint buffer。以 Adam 类优化器为例，参数本身可能以 BF16 保存，梯度和 master weight、动量与方差还需要额外空间；激活随 batch 和 sequence length 增长。ZeRO 通过分片优化器状态、梯度和参数降低单卡冗余 [2]，ZeRO-Infinity 进一步探索 CPU/NVMe offload [3]。

推理显存则主要由权重、KV cache、运行时 workspace 和 batch 中的临时张量组成。对 decoder-only 模型，可以用近似式估算每个请求的 KV：

$$
M_{KV}\approx 2\times L\times B\times T\times H_{KV}\times d_{head}\times bytes,
$$

其中 (L) 为层数，(B) 为序列数，(T) 为输入加输出长度，(H_{KV}) 为 KV head 数。GQA 通过减少 (H_{KV}) 降低 cache 压力，但模型结构和质量已经在算法侧决定。Infra 需要把最大长度、并发和精度转换成 admission control 和容量上限。

### 计算、带宽与通信

大模型性能不能只看 FLOPS。训练时需要矩阵乘法，也需要 GPU 之间的 all-reduce、all-gather、reduce-scatter；推理 decode 可能更受显存带宽和 KV 读取限制；数据处理可能受网络和对象存储吞吐限制。NCCL 为 GPU 集合通信提供了高性能 primitives [19]，但应用层仍要正确选择并行维度、通信时机和拓扑。

一个简单判断是：如果 GPU compute 利用率低、网络利用率高，可能是并行切分或 all-reduce 瓶颈；如果 compute 低、显存带宽接近上限，可能是 decode 或小 batch memory-bound；如果 GPU 和网络都低，而 CPU 或对象存储高，说明数据供给是瓶颈。需要同时观察多个层级的指标，不能只看一个 GPU utilization 百分比。

### 成本单位要与任务结果关联

训练成本可以按每百万有效 token、每个 checkpoint 或每个实验比较；推理成本可以按每百万输入/输出 token、每个成功请求、每个完成任务或每个工具动作比较。对于 reasoning 或 Agent 系统，输出 token 多不一定坏，关键是额外 token 是否带来成功率增益。

建议建立如下成本表：

```text
总成本 = GPU 时间成本
      + CPU / 存储 / 网络成本
      + 评估与人工标注成本
      + 失败重跑成本
      + 线上人工接管成本

单位成功成本 = 总成本 / 通过业务验收的任务数
```

如果模型吞吐很高但失败率也很高，单位成功成本可能更差；如果低延迟模型需要更多重试，p99 流量下的总成本可能超过慢但稳定的模型。Infra 评审必须把资源指标和质量、成功率、人工修复连接起来。

## 7.3 分布式训练：从单卡程序到 GPU 集群

### 数据并行

数据并行把不同 batch 分配给不同 GPU，每张卡保存完整模型副本，反向后通过 all-reduce 同步梯度。它实现简单，适合模型能放进单卡、通信带宽足够的场景。模型越大，单卡越放不下；batch 越小，梯度同步相对成本越高。经典大规模训练系统会把 data parallel 与 tensor/pipeline parallel 组合使用 [1][5]。

### 张量并行

张量并行把矩阵乘法切分到多张 GPU，例如按 hidden dimension 或 attention heads 分片。每层计算过程中需要 collective communication，适合单层矩阵很大、单卡无法容纳的模型。切分粒度必须与 GPU 拓扑匹配：同机 NVLink 的通信延迟和跨节点网络不同，TP degree 过大可能让通信成为主耗时。

### 流水线并行

流水线并行把不同 Transformer 层放在不同 GPU，输入 micro-batch 依次通过 stage。它减少单卡参数压力，却引入 pipeline bubble 和 stage imbalance。micro-batch 越小，bubble 可能越明显；micro-batch 越大，激活和排队压力增加。流水线切分不能只按层数平均，还要考虑不同层的算子、通信和激活成本。

### ZeRO、FSDP 与分片状态

ZeRO 将优化器状态、梯度和参数的冗余分片，按阶段逐步降低单卡内存 [2]。PyTorch FSDP 提供 fully sharded data parallel 的实现，把参数、梯度和优化器状态在数据并行组中分片，并在计算前 all-gather、计算后释放或重新分片 [18]。它们的共同代价是通信和调度复杂度上升。

选择 DDP、FSDP、ZeRO 或 Megatron 组合时，需要明确：模型是否能完整驻留、通信拓扑是什么、checkpoint 是否支持、优化器状态如何保存、是否需要 CPU offload，以及训练是否要和推理共享权重格式。没有统一最优方案，只有与模型规模、硬件和故障模型相匹配的方案。

### MoE 与专家并行

MoE 通过 router 为每个 token 选择少数专家，参数总量可以大于每个 token 的激活量，但 token 需要在 GPU 间路由。专家负载不均会导致某些 GPU 成为 straggler，capacity factor 过小会丢弃 token，过大又浪费显存。Infra 需要观测专家负载、token dispatch、通信量、溢出率和路由稳定性。

MoE 的扩展效率不等于 dense 模型的扩展效率。专家并行、数据并行和张量并行组合后，通信拓扑更复杂；容错时一个专家节点故障可能影响全局。训练平台应支持专家负载告警、故障重启和 checkpoint 恢复，不要只看总体 FLOPS。

### 并行度如何选择

可以把总并行度写成：

$$
P_{total}=P_{data}\times P_{tensor}\times P_{pipeline}\times P_{expert}.
$$

每个维度都可能带来通信、显存、调度和可恢复性代价。一个实用的选择顺序是：先让单层和单 stage 放得下，再根据机内拓扑确定 TP，按模型深度和 bubble 选择 PP，剩余 GPU 扩展 DP；MoE 再根据专家数量和路由流量选择 EP。小规模 benchmark 必须测真实拓扑和真实 sequence length，不能用理论计算量外推大集群性能。

### 集群扩展效率

理想情况下，GPU 数翻倍，吞吐也翻倍；实际扩展效率受到通信、同步、数据读取和尾部 straggler 影响：

$$
E(N)=\frac{throughput(N)}{N\times throughput(1)}.
$$

当 (E(N)) 随 GPU 数下降时，继续扩容可能不划算。Megatron-LM 的大规模训练研究展示了模型并行与集群训练如何协同 [1][5]；PaLM 的 Pathways 经验也说明，大规模训练需要统一计算、通信和故障管理 [7]。平台团队应把扩展曲线作为发布门禁：在目标规模上测吞吐、通信比例、checkpoint 时间和节点故障恢复。

## 7.4 数据、Checkpoint 与训练可恢复性

### 数据管线是训练系统的前端

训练 GPU 只在拿到有效 batch 时产生价值。数据管线要完成对象存储读取、解压、解码、去重、tokenization、混合采样、shuffle、packing 和 batch 组装。任何一个环节抖动，都会让 GPU 等待。Spark 等数据处理系统展示了大规模 DAG、分区和容错的基本方法 [27]；大模型训练通常还需要面向 token 的流式和分片设计。

数据管线要区分 raw document、clean document、token shard、packed sample 和 training batch。每一层都有版本和统计：文档数、有效 token、语言比例、平均长度、重复率、过滤率、错误率和采样权重。训练时记录 manifest hash 和 shard 顺序，才能在 checkpoint 恢复时继续得到一致或可解释的样本流。

### Shuffle、seed 与数据重复

分布式训练的随机性来自数据顺序、采样器、dropout、kernel 和并行归约。恢复训练时，如果 global step 相同但数据 shard 或随机状态不同，结果可能出现可见差异。严格复现成本很高，但至少要保存 rank 数、epoch、shard cursor、随机种子、采样权重和 tokenizer 版本。

数据重复会浪费训练预算，近重复还会造成评估污染。数据管线应在 shard 生成前去重，并在训练中记录每个数据桶实际消费的 token。若发生数据版本切换，要明确它是继续训练、阶段性配比还是新实验，不能让同一个 checkpoint 名称指向不同 manifest。

### Checkpoint 不只是模型权重

一个可恢复 checkpoint 通常包括：模型参数、优化器状态、学习率调度器、梯度 scaler、随机状态、数据游标、训练配置、并行拓扑、tokenizer、数据 manifest、代码版本和评估结果。只保存权重可以做推理，却不能保证从同一个优化状态继续训练。

Checkpoint 保存会产生大量 I/O 和 GPU/CPU 拷贝。可以使用分片、异步写入、增量 checkpoint、对象存储 multipart 和本地缓存，但必须定义一致性：训练进程崩溃时，不能得到一组模型权重来自 step (t)、optimizer 来自 step (t-1) 的伪恢复点。保存过程应有临时前缀、manifest、校验和、完成标志，读取时只选择完整版本。

### 保存频率与丢失工作量

Checkpoint 越频繁，丢失的训练工作越少，但 I/O 阻塞和存储成本越高。可以用近似成本做决策：

$$
Expected\ Loss\ Cost\approx failure\ rate\times recovery\ time\times compute\ cost.
$$

节点越多、运行时间越长、故障率越高，越应该降低恢复粒度。训练系统还要区分可恢复 checkpoint 和发布 checkpoint：前者服务训练重启，后者经过评估、转换和注册后供推理使用。

### 断点恢复与弹性训练

集群可能发生 GPU 错误、节点重启、网络分区、文件系统超时或容器驱逐。弹性训练需要决定：故障后等待原节点回来、缩小 world size 继续、替换节点重启，还是回滚到最近 checkpoint。每种策略都会影响优化状态、数据顺序和最终模型。

恢复流程应自动做健康检查：权重校验、optimizer state 校验、数据 shard 可读性、通信组重建、loss sanity check 和短步数回归。恢复后如果 loss 突然跳变，要能定位是数据、随机状态、精度、并行拓扑还是 checkpoint 损坏，而不是让训练静默跑几个小时。

### 训练数据安全与权限

数据平台必须把训练数据当作高价值资产。对象存储权限、脱敏、密钥扫描、租户隔离、审计和删除机制，不能因为数据进入 GPU 就失去控制。训练日志和样本抽样也可能泄露敏感文本，应限制访问和保留时间。

模型 checkpoint 同样可能包含训练数据记忆和内部能力，下载、复制、转换和发布都需要权限。平台要记录谁创建了 checkpoint、谁下载过、哪个数据版本训练而来，以及是否通过安全评估。模型注册表不是简单文件目录，而是算法、数据和治理元数据的索引。

## 7.5 Kernel、通信与存储：性能优化的正确层次

### 从算子到融合 kernel

Transformer 训练和推理包含 GEMM、attention、normalization、activation、通信和采样等算子。单个算子都正确，不代表组合执行高效：中间张量写入 HBM、kernel launch 次数、非连续 layout 和 padding 都会造成开销。FlashAttention 通过 IO-aware 的 tiling 减少 attention 的 HBM 读写，并保持精确结果 [9]；FlashAttention-2 进一步改善并行和 work partitioning [10]。

工程优化应先用 profiler 找到热点，再决定融合、重排、kernel 替换或模型结构调整。不要为了追求 benchmark 盲目改动算子，尤其要验证不同 sequence length、batch、精度、GPU 架构和边界输入。kernel 变化必须绑定版本和回归集，因为数值微小差异可能在长训练或低精度推理中累积。

### 集合通信与拓扑

all-reduce 适合同步梯度，all-gather 适合参数分片，reduce-scatter 可把聚合结果直接分片。NCCL 根据 GPU、PCIe、NVLink、InfiniBand 和网卡拓扑选择通信路径 [19]。拓扑探测错误、网卡配置不一致、MTU、拥塞或跨机带宽不足，都会让训练吞吐大幅下降。

性能分析不能只看通信总时长，还要看通信是否与计算重叠、是否被最慢 rank 阻塞、消息大小是否匹配、是否存在热点链路。训练系统应记录每种 collective 的耗时、字节数、调用次数、rank 方差和重试情况。集群变更后必须重新测扩展曲线。

### 存储层次与数据局部性

大模型平台通常有本地 NVMe、分布式文件系统、对象存储、缓存和 checkpoint 仓库多个层次。数据 shard 频繁读取时，直接访问远端对象存储会放大网络和请求开销；checkpoint 保存时，单个大文件会导致恢复并发不足。应按访问模式设计：热数据放本地或节点缓存，冷数据放对象存储，发布模型放有校验和版本的制品仓库。

缓存不能只看命中率。错误缓存、旧 tokenizer、错误权限或跨租户复用会产生正确性和安全问题。数据缓存的 key 至少包含数据版本、预处理版本和权限范围；模型缓存要包含权重、量化、kernel、模板和 adapter 兼容条件。

### 精度和数值稳定性

BF16、FP16、FP8、INT8 和 INT4 的选择需要同时考虑硬件支持、训练稳定性、通信量、显存和质量。训练中通常保留必要的高精度 master state，推理中可以使用更低精度但要做校准。混合精度需要监控 overflow、underflow、loss spike、梯度范数和异常 token。

低精度不是全局开关。embedding、归一化、输出层、敏感 attention 或 outlier 通道可能需要更高精度；量化后 kernel 的累加精度也会影响结果。平台应记录精度配置，并让评估系统能够在同一数据集上比较不同精度的质量—成本曲线。

### 编译、缓存和可重复性

Torch compile、CUDA graph、kernel autotune 和 TensorRT-LLM engine build 可以减少运行时开销，但会引入编译时间、shape 约束、设备兼容和缓存失效。TensorRT-LLM 文档覆盖了张量并行、量化和推理优化的工程接口 [16]；Triton Inference Server 提供模型加载、版本和 ensemble 服务能力 [17]。

编译产物必须绑定模型 hash、GPU 架构、CUDA/driver、输入 shape、精度和算子版本。一个 engine 在 A100 上生成，不能默认在 H100 或不同 driver 上复用。构建过程应可重放，发布前进行冷启动、热启动、最大长度和错误输入测试。

## 7.6 推理平台：从一次 forward 到在线服务

### Prefill 与 Decode

推理通常分为 prefill 和 decode。Prefill 处理输入 prompt，计算历史 token 的 attention 状态并产生首个 logits；decode 每次生成一个 token，读取历史 KV cache 并追加新 token。Prefill 计算并行度高，通常影响 TTFT；decode 单步计算小、需要反复读权重和 KV，通常影响 TPOT 和并发。

在线服务必须分别测量：

```text
TTFT = request arrival -> first output token
TPOT = decode time / generated tokens
E2E latency = queue + prefill + decode + postprocess
goodput = successful requests / time under SLO
```

只报告 tokens/s 容易掩盖排队和尾延迟。一个批处理吞吐很高的服务，可能让短请求等待长 prompt；一个单请求延迟低的服务，可能在并发时显存爆炸。平台要按 workload 分布测 p50、p95、p99 和成功率。

### Serving engine 的职责

推理引擎负责模型加载、权重分片、KV cache、batch 调度、sampling、streaming、量化 kernel、prefix cache 和错误处理。vLLM 以 PagedAttention、连续批处理和高效内存管理为核心 [11][12]；TensorRT-LLM 侧重编译和 NVIDIA GPU 优化 [16]；Triton 更像服务编排和模型管理层 [17]。它们可能组合使用，不应把“模型服务框架”当成一个单一组件。

服务接口还要处理 tokenizer 版本、chat template、停止词、结构化输出、工具调用、超时、取消和流式断开。客户端取消后是否立刻释放 KV cache，工具调用中断后是否保留上下文，重试是否导致重复副作用，都属于 Infra 的正确性问题。

### 模型并行和副本扩展

小模型可以单 GPU 部署多个副本，大模型通常需要 tensor parallel 或 pipeline parallel。副本扩展增加吞吐和容灾，但每个副本都占用完整权重或一组分片；模型加载时间和 cache warmup 也会影响扩容速度。路由层要知道副本健康、可用 KV 容量、当前排队和模型版本，不能只按 round-robin 分发。

对于多模型平台，路由还要考虑模型能力、租户、区域、价格、精度和数据驻留。fallback 模型必须经过任务和安全评估，不能在主模型不可用时任意切换到能力不等价的模型。路由决策要写入 trace，便于解释一次请求实际使用了什么模型。

## 7.7 调度、Batching 与 KV Cache

### Continuous Batching

传统 batch 等待一批请求一起开始、一起结束，但生成长度通常差异很大。Orca 提出的 iteration-level scheduling 使请求可以在每个生成迭代加入或退出 batch [13]。Continuous batching 提升了 GPU 利用率，却要求调度器实时管理 prefill、decode、KV 容量、优先级、取消和饥饿。

调度器至少要选择三个对象：本轮处理哪些请求、每个请求分配多少 token、是否允许新请求进入。只按 FIFO 可能让一个超长 prompt 阻塞大量短请求；只按短请求优先可能让长任务饥饿；只追求 batch 最大化可能违反 TTFT SLO。常见策略包括 token budget、deadline、优先级队列、chunked prefill 和公平配额。

### PagedAttention 与内存碎片

PagedAttention 把 KV cache 切成固定大小的 block，像虚拟内存一样由引擎管理，不要求每个请求占用连续大块显存 [11]。这减少了外部碎片，支持不同长度请求和动态 batch，也便于 beam 或 prefix 共享。它把 KV cache 从模型内部数组变成平台资源，带来 block 分配、回收、引用计数和 eviction 的系统问题。

KV block 分配要考虑请求取消、生成结束、异常、重试和多租户隔离。若 block 泄漏，服务会逐渐 OOM；若过度回收，重新计算 prefix 会增加 TTFT；若跨请求共享错误，可能泄露上下文。缓存 key 必须由完整 token prefix、模型版本、模板和权限范围组成。

### Prefix Cache

系统 prompt、工具 schema、few-shot 示例和长文档前缀经常重复。Prefix cache 可以复用 prefill 产生的 KV，减少重复计算。但只有 token 序列、位置编码、模型版本、模板和权限都一致时才安全。工具 schema 更新、系统提示变化、租户隔离或动态上下文插入都可能使旧 cache 失效。

缓存命中率不是唯一指标。还要测命中带来的 TTFT 降低、cache 内存占用、失效成本、跨模型共享边界和隐私风险。对于高动态请求，强行缓存可能增加管理成本；对于稳定长前缀，多级缓存可能是最有效的优化。

### Chunked Prefill 与长请求公平性

长 prompt 的 prefill 可能占据 GPU 多个迭代，阻塞正在 decode 的请求。Sarathi-Serve 通过 chunked prefill 把长 prompt 切成块，与 decode 交错执行，以改善吞吐和延迟权衡 [14]。切块大小决定计算利用率和抢占粒度：太小会增加 launch 与调度开销，太大仍然阻塞 decode。

长请求还要设置 admission control：最大输入 token、最大输出 token、并发占用、租户配额和超时。超过限制时可以拒绝、截断、压缩、转异步队列或路由到低优先级池。重要的是提前拒绝而不是让请求进入后才 OOM。

### Prefill/Decode Disaggregation

Prefill 和 decode 的资源画像不同：前者偏计算，后者偏内存带宽和 KV。DistServe 把二者分离到不同资源池，允许分别扩容和优化 SLO [15]。分离带来 KV transfer、网络、状态一致性和调度复杂度，适合流量规模和延迟要求足够大、能够摊平额外通信成本的场景。

设计 disaggregation 时要回答：KV 如何传输和序列化，传输期间请求如何取消，decode 节点如何发现 prefill 已完成，失败时是否重做 prefill，跨机网络是否成为瓶颈。小规模部署不一定值得引入这种复杂度；平台应通过 workload 画像和容量模型证明收益。

## 7.8 资源调度、隔离与平台化

### GPU 资源编排

Kubernetes 可以通过 device plugin 和资源请求调度 GPU [20]，但大模型训练通常还需要拓扑感知、gang scheduling、优先级、抢占、节点健康和本地数据。一个分布式训练 job 需要一组 GPU 同时可用，只有部分资源会导致长期 pending；推理服务可以弹性扩容，但模型加载和 warmup 时间使瞬时扩容不一定及时。

集群调度应区分训练、推理、评估和交互式开发资源池。训练追求长时间稳定和大 gang，推理追求低延迟和滚动升级，评估追求大量可并行任务，开发追求快速反馈。混在同一资源池中，低优先级训练可能抢占在线推理，或交互式任务碎片化大 GPU。

控制面首先要维护的不是“有多少张卡”，而是一份可以结算的资源账本。对训练作业，账本记录请求的 GPU 型号、数量、节点与互联约束、CPU、内存、本地 NVMe、网络、对象存储吞吐、预计运行窗口和 checkpoint 保留期；对推理副本，记录权重、量化格式、batching 参数、并发上限、KV 预算、模型加载时间和可服务 token 速率；对评估与开发任务，记录优先级、可抢占性、数据权限和过期时间。控制面把这些请求转换为 reservation、allocation、usage 和 release 四种事实，而不是只保存一份期望配置。这样才能在节点故障、抢占和用户取消后回答资源是否真正释放。

资源账本必须同时拥有“请求量”和“实耗量”。请求量用于准入与公平：一个八卡训练任务若只申报四卡，就会挤占不属于它的容量；实耗量用于容量与成本：长期低利用率的 reservation 可能需要回收或降级。两者之间的差异不是立即的违规证据，因为模型加载、checkpoint、通信同步和长上下文请求都可能造成短时空闲；但平台应按工作负载类型设定观察窗口和解释码。训练作业可在稳定 step 窗口内评估有效 token/s 和通信等待，推理副本可按长度分桶评估排队、TTFT、TPOT、KV 水位和取消率，不能用一个全局 GPU 利用率阈值处罚所有任务。

拓扑是资源账本的一部分。八张跨机 PCIe GPU 与八张同机高速互联 GPU 在训练、TP 和 MoE 场景中的可用能力不同；同样的显存总量也不能保证模型、激活、优化器状态和通信 buffer 同时放得下。控制面应把 topology class、故障域和网络带宽写入可调度属性，调度器只在兼容集合中做 placement。Borg 对资源分配、优先级和集群效率的讨论 [21] 提供了这一控制逻辑的基本框架；Kubernetes 的资源请求和 device plugin 是执行底座 [20]，但它们不能自动判断一组 GPU 是否满足某种并行策略。

资源生命周期还需要处理“僵尸占用”。worker 已失联、训练进程仍在运行、NCCL communicator 未释放或对象存储上传卡住时，调度器看到的状态可能与真实设备不同。每个 allocation 因而需要 lease、心跳、fencing token 和可验证的释放确认。lease 过期只表示控制面可以开始调查或隔离，不表示可以立刻把同一块 GPU 给另一个高风险任务；执行面必须确认旧进程、容器、挂载和网络身份已经被撤销。对于无法确认的节点，宁可暂时摘除也不要超卖，这与数据库中不把失联锁直接视为已释放是同一种保守性。

资源控制面可以用以下最小状态机表达：

```text
requested -> admitted -> reserved -> allocated -> warming -> serving/running
    -> draining -> released
                     ^
                     | failure, preemption, cancellation, lease expiry
```

每次迁移带上 actor、policy version、request revision、资源快照和原因。`admitted` 只表示预算与配额通过；`reserved` 表示容量在计划窗口被预留；`allocated` 表示执行面已获得资源；`warming` 对推理尤其重要，因为权重尚未加载完成时不能把容量当作可服务吞吐；`draining` 则允许已开始的 decode 或 checkpoint 在 deadline 内收敛。把这些状态显式化，能避免“发布看起来完成、其实新副本还没有健康”或“训练被取消、其实仍在消耗 GPU”的常见误判。

### 调度公平与容量预留

Borg 等集群管理系统的经验表明，资源调度要同时处理优先级、配额、隔离、故障域和利用率 [21]。大模型平台还要增加显存、模型权重、KV 容量和通信拓扑维度。GPU 数量相同，不同 GPU 型号和互联拓扑的有效容量不同。

容量规划不能只按平均 QPS。要看输入/输出 token 分布、并发峰值、长尾请求、模型路由比例、缓存命中和故障余量。建议建立容量方程：

```text
required capacity
  = peak token rate / sustainable token rate per GPU
  + failover reserve
  + rolling deployment reserve
  + capacity for long-context tail
```

### 多租户隔离

多租户模型平台要隔离权重、adapter、KV cache、日志、数据和工具权限。GPU 共享可以提高利用率，但可能造成显存争抢、尾延迟相互影响和数据泄露。租户配额应覆盖并发请求、输入 token、输出 token、模型种类、工具调用和缓存大小。

租户级 trace 和成本归因也很重要。每个请求要带 tenant、project、model、version、route 和 cost center；离线训练要带数据拥有者、实验 ID 和 GPU 预算。没有归因，平台无法回答谁消耗了资源、哪个模型最贵、哪类请求造成排队。

成本归因的粒度应与决策粒度一致。训练的最小归因单元通常是 experiment 或 dataset revision，推理是 request、conversation 或 business operation，Agent 则可能需要延伸到 tool、retry、人工审核和最终 outcome。每个用量事件至少包括 usage id、开始与结束时间、资源类型、租户、项目、成本中心、模型/制品版本、区域、分配方式、计量值、计价规则版本和来源。计量事件要幂等：同一段 GPU 使用、同一组 token 或同一次对象存储写入即使被重复消费，也只能结算一次。账单不是观测 dashboard 的导出，而是独立、可重放的事实流。

“每百万 token 的价格”不足以支撑平台优化。一次请求的真实边际成本可能包含 prefill、decode、KV 驻留、cache miss、路由到更大模型、重试、工具调用、网络出口和人工处理；训练成本还包含空闲 reservation、失败 step、checkpoint、数据读取和评估。较有用的单位是每次成功任务、每个被接受的工单、每个通过的测试或每个有效训练 token 的成本，但这些单位必须同时报告质量和延迟。否则团队可能通过缩短输出、拒绝困难任务或降低验证次数来制造“成本下降”的假象。

为了支持审计，成本链应能从总账一路钻取到单个动作。例如一个业务 Run 的根成本为 12 元，其中模型 token 6 元、GPU reservation 分摊 2 元、搜索工具 1 元、重试 1 元、人工批准 2 元；若任务失败，系统还应显示是模型质量、工具故障还是预算策略导致失败。这样，负责人能判断应该调模型路由、修复工具、优化 cache 还是限制输入长度，而不是把所有异常归咎于某个团队。归因也必须支持共享成本：基础模型 warm pool、公共向量索引和观测系统可以按预先公布的规则分摊，不能在月末按临时印象摊派。

多租户隔离的公平不是平均分卡。交互式生产请求通常有严格 deadline，训练可以接受等待但需要连续 gang，评估适合在空档批量填充。控制面可同时采用 hard quota、reservation、借用和回收：hard quota 防止一个租户耗尽集群；reservation 为关键服务保留容量；借用让空闲容量被低优先级任务使用；回收在 reservation 真正需要资源时让可检查点的任务有序退出。借用必须有最大期限、可抢占标记和 checkpoint 协议，不能把“空闲”误解为永久所有权。

公平策略应对外可解释。租户看到 pending 时，应能知道是配额、预算、拓扑、容量、优先级、数据权限还是发布冻结造成，而不是只得到“no GPU”。管理员也应能模拟策略：新增一个八卡训练 job 会影响哪些在线 SLO，允许某租户临时借用会挤掉多少评估，某模型上线后需要多少滚动余量。模拟结果不是承诺，但能把容量争论从头衔和直觉转回显式假设。

### 发布、灰度与回滚

模型发布包括权重制品、tokenizer、模板、量化 engine、服务配置和评估门禁。灰度可以按租户、区域、请求类型或流量比例进行；影子流量可以只执行推理而不触发副作用工具。发布过程中要监控错误率、质量回归、TTFT、TPOT、OOM、cache 命中和成本。

回滚必须是可执行的状态转换，不是把一个 tag 改回旧值。需要保留旧权重、旧 engine、旧模板、旧 schema、旧路由和旧安全策略，确认旧版本仍能在当前硬件和依赖上启动。每次发布记录变更和指标，便于事故复盘。

发布门禁应把“制品可启动”与“变更可接受”分开。前者验证签名、依赖、权重完整性、tokenizer、模板、量化 engine、模型卡、权限和回滚包；后者验证离线质量、格式兼容、红队与安全评估、容量模型、成本变化、长上下文、故障降级和业务 SLO。只有两类证据都满足，才有资格进入影子和灰度。一个能加载却使结构化输出失效的模型不能发布；一个离线分数提升却让 p99 或成本超过预算的模型也不能直接扩大流量。

灰度门禁应按阶段定义停止条件，而不是只写“观察一段时间”。例如先用录制流量和只读 shadow 验证请求解析、路由和资源画像，再对内部或低风险租户放出有限比例；每阶段比较新旧版本在质量、拒绝率、TTFT、TPOT、OOM、取消率、cache 命中、单位成功任务成本和安全事件上的差异。阈值、统计窗口、owner 和自动/人工处置动作必须在发布前写定。若某项指标越界，系统冻结扩容、保留现场证据并按已验证路径回滚，而不是等待模型团队解释后再决定。

一个成熟的发布控制面还要处理配置组合。模型权重本身可能不变，但 tokenizer、chat template、工具 schema、采样默认值、prefix cache、路由规则、GPU driver 或 kernel 的变化都可能改变用户可见行为。发布单元应因此是不可变 manifest，而不是单一模型 tag：manifest 列出所有输入制品及其 hash、资源配置、评估集版本、批准记录和已知风险。请求 trace 记录实际命中的 manifest，事故发生后才能准确复现；回滚也回滚到经过验证的 manifest，而不是拼凑“看起来相同”的旧配置。

可以把关键决策写成轻量 ADR，避免平台在事故后才追问为什么这样取舍：

| ADR 维度 | 示例：将长上下文流量分到独立资源池 |
|:---|:---|
| 背景 | 长请求挤压普通请求的 KV，造成共享副本 p99 抖动 |
| 决策驱动 | 普通请求 SLO、长请求完成率、隔离、成本、运维复杂度 |
| 候选 | 共享池限流；统一扩大副本；按上下文长度分池 |
| 决策 | 按长度分池，并保留共享池的降级路由 |
| 获得 | 更可预测的 KV 容量与普通请求延迟 |
| 主动牺牲 | 资源碎片和较低的瞬时总体利用率 |
| 接受风险 | 长池冷启动和跨池路由失败 |
| 验证/重评估 | 比较 p99、拒绝率、闲置率、成本和长任务完成率 |

ADR 不替代实验数据，但它规定在什么约束下解释实验数据。发布后若观察到闲置成本高于隔离收益，团队可以根据既定重评估条件合并资源池；若发生跨租户 cache 风险，则应立即提高隔离优先级。将获得与牺牲放在同一维度比较，能防止“吞吐提升”掩盖安全、可运维性或恢复能力的下降。

## 7.9 可靠性、观测与 Infra 验收

### SLO 不只是一条延迟线

在线 LLM 服务的 SLO 至少包括可用性、TTFT、TPOT、完成率、错误率、超时率和质量门禁。对流式请求，“HTTP 200”不代表成功：可能只返回了部分 token、工具调用格式损坏或用户中途取消。对 Agent，最终任务成功和工具状态一致性比单次模型响应更重要。

可以把 SLO 分层：平台层保证服务可用和延迟，模型层保证格式和能力回归，业务层保证任务成功和人工接管率。错误预算用于决定是否继续发布、是否暂停实验、是否把资源投入稳定性。SRE Workbook 提供了服务等级和错误预算的系统方法 [22]。

### Metrics、Logs 与 Traces

Metrics 适合时间序列和聚合趋势，logs 记录单次错误与上下文，traces 记录跨服务和跨工具的因果链。Dapper 展示了大规模分布式 tracing 如何把请求在多个服务中的路径串起来 [23]；OpenTelemetry 提供统一的 trace、metric 和 log 语义 [24]；Prometheus 提供指标采集、查询和告警基础 [25]。

大模型 trace 需要额外字段：model、model version、tokenizer、prompt hash、input/output token、TTFT、TPOT、queue time、cache hit、sampling、tool calls、verifier、finish reason、safety decision 和 estimated cost。默认不要记录完整敏感 prompt 和隐私输出；应做脱敏、采样、访问审计和保留时间控制。

观测设计的第一原则是把控制面决定与数据面结果连起来。一次请求的 trace 不能只知道“模型慢”，还要知道它命中了哪个 routing rule、在哪个队列等待、由哪个 scheduler admission、分配了哪种 GPU、是否发生 prefill/decode 迁移、KV 是否因配额被驱逐、以及最终归因到哪个 tenant 和成本中心。没有这些字段，SRE 只能看到症状，平台团队无法判断根因是容量不足、调度不公平、缓存策略、模型版本还是上游输入变化。OpenTelemetry 的统一上下文传播可以承载关联 id [24]，但字段的语义、脱敏范围和基数控制必须由平台契约定义。

指标也要避免两个极端。只记录全局平均值会掩盖长上下文、低频模型、特定区域和单一租户的退化；为每个 request id、prompt 或用户建立指标标签又会导致时序系统失控。正确做法是把高基数事实放 trace 或日志，把可聚合维度放 metrics：例如按模型、版本、长度桶、区域、硬件池、路由原因和结果类别聚合 TTFT、TPOT、完成率、OOM、queue time、cache hit 和成本；对异常样本保留可检索 trace。指标字典应写清名称、单位、分母、窗口、延迟、owner 与允许标签，避免不同 dashboard 用“成功率”指代不同东西。

观测数据本身有成本与风险。全量记录 prompt、输出、工具参数或 KV 事件既昂贵又可能泄露隐私；过度采样又会遗漏罕见故障。平台可采用分层留存：所有安全、失败、超预算、回滚和人工接管事件保留完整受控证据；成功请求保留摘要和少量采样；聚合指标长期保留；原始敏感 artifact 按租户政策加密、最小化访问并设置删除期限。采样规则同样需要版本化，因为若发布后只改变采样率，趋势图可能看似质量下降或上升而实际上只是观测口径变化。

错误预算应连接发布与资源控制。若某模型路由在最近窗口内消耗了 latency 或可用性错误预算，控制面可以冻结扩容、降低其流量、增加可用副本或把新实验转为 shadow；若质量 gate 消耗过快，则应停止灰度，即使基础设施指标健康。SRE Workbook 对错误预算的核心启发是将可靠性目标转换为决策边界 [22]，而不是只在月报中展示一个百分比。对 AI Infra 而言，质量、成本和安全也可以有各自的预算，但它们不能被任意相互抵消：更便宜的模型不能用安全事件换取，吞吐提升也不能抵消结构化输出错误。

故障演练应在控制面与数据面两个层次同时验证。数据面演练包括杀掉副本、注入 OOM、限制对象存储、使 KV transfer 失败和模拟 token 流中断；控制面演练包括错误路由、过期 reservation、计费事件重复、策略服务不可用、发布 manifest 不兼容和区域容量突然下降。每项演练记录检测信号、告警延迟、自动处置、人工升级、恢复时间、丢失或重复的请求、成本上限和剩余风险。容器能自动重启只证明一个局部机制存在，不能证明用户任务、资源账本与发布状态能收敛。

### 训练观测

训练侧要观测 loss、学习率、梯度范数、吞吐、有效 token、GPU 利用率、显存、通信、数据等待、checkpoint、节点故障和恢复时间。按数据桶、语言、长度和任务分层的 loss，常常比全局平均更早发现数据或能力退化。

训练平台还应生成 experiment manifest，把代码、配置、数据、环境、并行拓扑、随机状态、checkpoint 和评估关联起来。一个 dashboard 只能告诉你当前状态，manifest 才能支持复现、审计和比较。

训练的资源效率也需要与结果绑定。一次训练 run 的 GPU 小时可能因为 batch size、数据等待、通信、重算、checkpoint、节点替换和评估而不同；只比较总 GPU 小时无法区分“更快收敛”与“更快浪费”。报告应同时包含有效 token、有效 step、训练/验证曲线、恢复次数、数据版本、失败原因、最终评估和成本区间。对于可中断或弹性训练，还要记录重新 placement 后的吞吐与数值一致性，避免通过频繁重启把真实的调度碎片隐藏在平均值中。

控制面应把训练作业的最终状态写回制品和成本目录。一个 checkpoint 若没有对应的代码、数据、tokenizer、并行配置、随机状态和评估证据，就不应被当作可发布模型；一个实验若被取消，也应记录取消时已消耗的资源和是否留下可复用 checkpoint。这样，研究团队可以从结果回溯成本与数据，平台团队可以从容量异常回溯具体实验，治理团队可以从制品回溯训练证据。可复现不是单纯的科研习惯，而是让资源、质量和责任能够闭环的系统接口。

### 推理观测

推理侧不能只看平均 tokens/s。至少需要输入长度分桶、输出长度分桶、并发、队列等待、prefill 时间、decode 时间、KV cache 使用、cache 命中、GPU 显存水位、OOM、取消、重试和路由结果。p99 长请求往往决定用户体验和容量，而不是平均请求。

性能回归要与质量回归同时做。量化后吞吐增加但代码通过率下降，不能算成功；prefix cache 命中后 TTFT 降低但跨租户隔离失败，更不能上线；更激进的 batching 提高平均吞吐但长请求饿死，也不符合 SLO。

### 故障模型与演练

要提前列出故障：单 GPU、单节点、网络、对象存储、模型加载、KV OOM、kernel crash、tokenizer 不兼容、上游限流、工具超时、checkpoint 损坏和区域不可用。每种故障都要定义检测、降级、重试、隔离、回滚和人工处理。

训练故障可回到最近 checkpoint；推理故障可摘除副本、路由到备用模型或降级为异步；工具故障可返回明确的 partial result；不可重试的副作用动作必须依赖幂等键和状态查询。故障演练要验证恢复时间、丢失请求、数据泄露和成本，不只是看容器能否重启。

### Infra 验收清单

交付一个大模型 Infra 平台前，可以用以下清单评审：

```text
资源：模型、激活、KV、optimizer、checkpoint 的内存账本是否清楚？
训练：并行策略、通信拓扑、数据供给、恢复和扩展曲线是否验证？
推理：TTFT、TPOT、吞吐、队列、cache、长上下文和 OOM 是否有指标？
服务：模型版本、模板、量化、adapter、schema 是否作为发布契约？
调度：租户、优先级、配额、拓扑、故障域和滚动发布是否隔离？
观测：metrics、logs、traces 是否能关联一次请求的全链路？
治理：成本、权限、脱敏、审计、灰度、回滚和错误预算是否可执行？
```

对后端工程师而言，最重要的思维变化是：模型不是一个函数调用，而是一个需要资源、状态和生命周期治理的分布式系统。每一次模型能力提升，都会重新改变显存、网络、调度、评估和成本边界；Infra 的职责不是隐藏这些复杂性，而是把复杂性变成可测量、可控制的接口。

在上线前，可以把“资源与治理”作为独立 gate 来复核。先确认资源账本是否覆盖 reservation、allocation、实际 usage 和释放，是否能解释拓扑与 warmup 造成的差异；再确认成本是否能从租户、项目、模型和请求钻取到 token、GPU、工具与人工；随后确认发布 manifest、评估证据、灰度阈值和回滚包是否一一对应；最后在故障与高峰组合负载下验证租户隔离、错误预算和人工升级。任何一个问题没有证据，都应缩小影响范围而不是扩大流量。

这种 gate 还要求明确责任分工。模型团队对能力、评估集、已知限制和制品签名负责；Infra 团队对调度、容量、路由、计量、隔离和恢复负责；业务团队对成功定义、风险接受和人工路径负责；安全与治理团队对数据、权限、审计和发布政策负责。责任边界并不意味着问题可以被转交：一次发布需要这些 owner 在同一个 manifest 与 ADR 上留下证据。没有 owner 的指标、没有回滚路径的制品、没有成本中心的任务都不应进入生产资源池。

## 7.10 生命周期、五平面与决策地图

本章只建立总览，不重复第8到第12章的细节。后续章节会分别展开训练作业、推理服务、数据评估、Agent Runtime 和治理控制；这里保留一张跨章节地图，帮助读者在设计评审中判断问题应落在哪一层。

### 生命周期：从实验到退役

大模型 Infra 的生命周期可以压缩成七个阶段：

```text
数据进入 -> 训练运行 -> 制品注册 -> 推理交付
  -> 评估反馈 -> 运行治理 -> 退役归档
```

数据进入阶段关注来源、授权、清洗、去重、版本和统计；训练运行阶段关注并行、通信、checkpoint、恢复和成本；制品注册阶段关注权重、tokenizer、模板、量化、adapter、评估和签名；推理交付阶段关注 SLO、KV、batching、路由、隔离和发布；评估反馈阶段关注回归集、judge、线上信号和漂移；运行治理阶段关注安全、审计、灾备、成本和事故；退役归档阶段关注模型下线、数据删除、证据留存和依赖清理。

### 五个平面

| 平面 | 管什么 | 典型问题 | 后续章节 |
|:---|:---|:---|:---|
| 数据平面 | 训练数据、评估数据、线上反馈、血缘 | 数据是否可追溯、可授权、可重算 | 第8、10章 |
| 计算平面 | GPU、网络、存储、kernel、通信 | 资源是否足够、拓扑是否匹配、成本是否可解释 | 第8、9章 |
| 状态平面 | checkpoint、KV、请求、workflow、工具操作 | 失败后从哪里恢复，重试是否重复副作用 | 第8、9、11章 |
| 控制平面 | 调度、路由、发布、配额、回滚 | 谁可以启动、升级、降级和停止 | 第8、9、11、12章 |
| 治理平面 | SLO、安全、审计、供应链、责任 | 证据在哪里，风险由谁接受 | 第10、12章 |

五个平面不是组织架构，而是分析工具。一个线上质量事故可能同时跨越数据平面和推理平面；一次模型升级可能同时影响制品、状态、控制和治理。设计评审要把问题拆到这些平面上，再决定应该修改算法、数据、调度、运行时还是治理策略。

五平面还提供了事故的排查顺序。先确认数据面是否改变了输入分布、权限或版本，再确认计算面是否出现容量、拓扑、通信或存储瓶颈；随后检查状态面中的 checkpoint、KV、请求和幂等是否一致，检查控制面是否做出了预期的 admission、路由、发布和回收决定，最后检查治理面中的阈值、审批、审计和责任是否仍然有效。这个顺序不要求每个事故按线性流程处理，但能防止团队只盯模型输出、忽略让输出进入生产的其它四个边界。

例如某次上线后用户反馈答案截断，数据面可能是新模板增加了隐藏上下文，计算面可能是 KV 容量不足，状态面可能是流式请求在迁移时丢失 offset，控制面可能是路由把长请求送进短请求池，治理面则可能是只设置了平均延迟门槛而没有 completion-rate gate。五个假设需要通过同一条请求的 manifest、trace、资源账本和评估证据来比较。若只因为模型“看起来变差”就回滚权重，既可能错失真正的系统问题，也可能破坏已经修复的质量问题。

### 决策地图

| 观察到的问题 | 先看什么 | 不要先做什么 |
|:---|:---|:---|
| 训练越扩越慢 | step 时间线、通信拓扑、数据等待、checkpoint | 直接加 GPU |
| 推理 p99 上升 | 输入/输出长度、队列、prefill/decode、KV 水位 | 只看平均 tokens/s |
| 模型分数上涨但线上变差 | 评估集污染、线上分布、反馈偏差、任务成功率 | 只扩大 benchmark |
| Agent 重试后出现副作用 | operation id、幂等、工具状态、审批事件 | 让模型“更小心” |
| 成本异常 | 单位成功任务成本、重试、cache 命中、人工接管 | 只压低模型单价 |
| 安全或合规不确定 | 数据来源、权限、审计、供应链和责任人 | 只加输出过滤 |

这张地图还可以成为评审会议的输入模板。每个提案先写出它改变的平面、依赖的事实、失败后由谁检测、如何降级、如何回滚、成本由谁承担，以及需要哪些观测来推翻原假设。比如引入新的 speculative decoding 策略，计算平面要证明吞吐与显存假设，状态平面要说明 draft/target 结果的取消与一致性，控制平面要说明何时启用与禁用，治理平面要说明质量与成本门槛；若这些内容缺失，提案仍是性能实验，不是可交付的平台能力。

跨平面决策也要求避免局部最优。扩大 batch 可能改善计算利用率却伤害状态面的长请求公平；更激进的 prefix cache 可能降低成本却提高隔离风险；更严格的发布 gate 可能提高治理质量却延长紧急修复时间。ADR 应把一致性、可用性、延迟、吞吐、成本、复杂度、可运维性、数据新鲜度、恢复能力和团队负担放在同一张对照表中，并明确哪些损失被接受、哪些由 fallback 缓解、何时重新评估。没有主动承认的牺牲，通常会以线上事故的形式出现。

控制平面本身也必须被当作一个会失败、会滞后的产品，而不是永远正确的裁判。调度器、策略服务、制品目录、配额账本和发布审批分别不可用时，数据平面不能临时依赖模型猜测该做什么。设计应预先规定默认态：已经持有租约的请求在多长时间内可继续，尚未开始的高风险操作是拒绝、排队还是仅允许只读，哪些紧急回滚可以绕过常规队列，以及策略恢复后如何把离线期间的计量、审批和状态变化重新对账。默认态应倾向于收缩权限和影响范围；把旧策略无限期缓存为“正常运行”同样会把过期授权和过度承诺带入事故。

控制权的 owner 也要能随决策阶段转移而不丢失责任。容量负责人可以拥有 admission 阈值，模型 owner 可以拥有质量门禁，业务 owner 可以决定一个降级结果是否可接受，安全 owner 可以冻结某类工具；但任何人都不应单独改写另一个边界的事实。一次策略变更应记录提案人、批准人、执行身份、目标范围、有效期、依赖版本、观测窗口和撤销条件。若某个 owner 不在线，系统应进入预先批准的保守策略或明确升级到值班人，而不是由无审计的管理员临时扩大权限。这样，事故中的快速处置仍然有可追溯的授权链。

重新评估不是等到季度复盘才发生。控制面应把触发条件写成可以观察的信号：实际队列或 KV 水位持续偏离预测、单位成功任务成本越过预算、某个 manifest 的人工接管率上升、跨故障域可用容量低于承诺，或策略服务的模拟结果与真实结果长期不一致。触发后先冻结进一步放量，保留当时的输入画像、策略版本和决策事件，再由对应 owner 判断是修正资源模型、缩小适用租户、回滚策略还是追加容量。恢复原有范围也需要新的证据；不能因为告警消失就自动撤销限制。这个闭环使控制平面不仅能下达决定，也能证明决定仍适用于变化后的负载与风险。

为避免评估只停留在会议纪要，控制面还应把每次决定的反事实写入演练：若不收紧 admission、若继续使用旧路由、若提前解除配额，哪些 SLO、成本或隔离边界会先失效。反事实不要求精确预测事故日期，却要求把可观测指标、阈值和责任人关联起来。之后用真实事件校正这些假设，能够区分策略本身失效、观测缺失与执行未落地三种不同问题，避免团队反复用同一个阈值掩盖不同根因。

### 本章小结：把复杂度变成契约

AI Infra 的核心能力不是堆叠组件，而是把算法复杂度转化为资源契约、状态契约、质量契约和治理契约。资源契约说明模型在不同长度、并发和硬件下需要多少计算、显存、网络与存储；状态契约说明请求、KV、checkpoint、数据游标和工具操作如何创建、恢复、过期与回滚；质量契约说明模型升级、量化、kernel、调度和降级对输出质量与用户任务的影响；治理契约说明谁批准、谁负责、证据保存在哪里。

后续第8到第12章会沿着这张地图逐层展开。读者不必先记住所有组件名称，但应始终追问同一个问题：这个算法或平台选择改变了哪项资源、状态、质量或治理约束，系统用什么证据证明它在真实负载下仍然成立？

最终，AI Infra 的可靠性来自可执行的闭环：工作负载进入时有资源与权限准入，执行时有状态、隔离与观测，结果产生后有质量、成本和责任归因，变更发生时有分阶段门禁与可验证回滚，故障出现时有已知的恢复与人工路径。把这些环节拆成组件很容易，把它们通过同一套 identity、manifest、事件和证据连接起来才是平台设计的难点。只有当资源消耗、模型版本、业务 outcome 与风险处置能在同一条链路上被解释时，团队才真正拥有可扩展的 AI Infra，而不仅是一组能够运行的 GPU 服务。

这也给容量计划提出了更严格的要求。计划不应只输出“下季度买多少 GPU”，还要给出需求假设、负载类别、模型与长度分布、每类服务的 SLO、故障和发布余量、可借用容量、数据与网络瓶颈、预算上限以及触发扩容或收缩的指标。假设变化时，控制面用同一套资源账本重新计算，而不是沿用过去峰值乘一个经验系数。对于高度不确定的新模型或 Agent 工作流，应先保守地设置并发、预算和影响范围，用真实的 token、KV、工具与人工数据更新模型，再决定是否扩大 reservation。

平台团队还应区分“可用容量”和“可承诺容量”。可用容量是当前空闲设备；可承诺容量要扣除故障域冗余、滚动发布、模型加载、驱动维护、长上下文尾部和已经授出的 reservation。把二者混在一起会在高峰或事故期间过度承诺。容量承诺的消费者也必须可见：产品团队知道何时可能被降级，研究团队知道何时可能被抢占，财务团队知道共享 warm pool 的成本如何分摊。透明的承诺模型有时会暴露资源不足，但它能让买卡、优化模型、排队和限制范围成为可比较的选择。

资源策略变更本身也应按发布处理。调整一个租户配额、抢占规则、cache 额度或路由权重，可能不改任何模型权重，却会立刻改变延迟、成本、公平与风险。控制面因此要为策略保留版本、模拟结果、批准记录、灰度对象、观察指标和回滚版本；上线后将实际排队、拒绝、借用和回收事件与预测比较。若策略模拟无法解释真实结果，问题可能在资源画像、计量口径或隐藏依赖，而不是简单地把阈值调大。

当预算成为硬约束时，降级路径也要有产品语义。系统可以缩短最大上下文、减少候选、降低优先级、转异步、切换到较小模型或等待人工，但必须告诉调用方发生了哪一种降级、对结果有什么影响、是否可以稍后恢复完整服务。静默地省略检索、验证或安全步骤会制造不可审计的质量变化。预算策略的目标不是让每个任务都在限额内“看似完成”，而是在资源紧张时仍使系统的承诺、风险和用户预期一致。

资源治理的验收因此要使用组合情景，而不是孤立压测：在滚动发布期间注入一台节点失效，同时让一个租户发起长上下文峰值、另一个租户执行可抢占训练，再使计量或策略服务短暂不可用。系统应证明在线服务保留承诺容量、训练按 checkpoint 有序让出、计量事件最终对账、策略失效进入保守模式，且每个决策都有 trace 与 owner。真实事故往往正是这些条件叠加；只在平稳环境下得到的吞吐数字不能证明平台具有弹性。

对外部依赖也要建模。云区库存、GPU 驱动、网络配额、对象存储吞吐、模型仓库、鉴权服务和上游模型 API 都可能限制承诺容量。控制面把依赖健康、速率限制、过期时间和 fallback 写进容量模型，在依赖不健康时减少 admission 或切换受控降级，而不是继续接收无法兑现的任务。这样，资源管理从“调度器如何放置容器”扩展为“平台能向用户可靠承诺什么”。

每次容量决策都应留下可复核的假设快照：预测负载、可持续 token 速率、故障余量、预期 cache 命中和所选降级路径。事后比较预测与实际，才知道应修正画像、扩容资源、优化模型，还是收紧承诺范围。

容量模型应定期用真实高峰、发布窗口和故障演练校准，而不是只用离线 benchmark 更新。模型、token 分布或租户结构变化后，旧的吞吐结论可能已经失效；把校准结果写入新的策略版本，才能让资源承诺持续可信。

校准过程还应保留反例：哪些预测在长尾输入、跨区流量、模型冷启动或依赖退化时失效，最终采取了什么更保守的 admission 与降级。反例能约束容量模型的适用范围，避免团队把一次成功压测误写成通用承诺。

当证据不足时，平台应把不确定性显式计入余量，而不是通过提高平均利用率掩盖它。对用户而言，可解释的受限服务优于高峰时无法兑现的无限承诺。

同样重要的是保留决策后的复盘窗口：当实际负载、成本或质量偏离假设时，负责人应能暂停进一步承诺、收集证据并调整策略，而不是让旧配额自动延续。资源治理的目标不是让集群永远满载，而是在变化中持续兑现已说明的服务边界。

最后，控制面应持续验证自身不会成为单点风险。策略、配额、路由、制品目录、计量、发布审批和告警服务失效时，数据面需要明确的保守行为：已有请求能否在租约内完成，新请求是拒绝还是进入受限队列，紧急回滚是否有独立通道，计量延迟如何补账。只有把控制面故障也放入演练，平台才能在“模型服务本身仍可运行、但管理能力暂时不可用”的情况下维持正确边界。

## 参考资料

[1] Shoeybi, M., et al. *Megatron-LM: Training Multi-Billion Parameter Language Models Using Model Parallelism*. arXiv, 2019. https://arxiv.org/abs/1909.08053 访问日期：2026-09-22

[2] Rajbhandari, S., et al. *ZeRO: Memory Optimizations Toward Training Trillion Parameter Models*. SC, 2020. https://arxiv.org/abs/1910.02054 访问日期：2026-09-22

[3] Rajbhandari, S., et al. *ZeRO-Infinity: Breaking the GPU Memory Wall for Extreme Scale Deep Learning*. SC, 2021. https://arxiv.org/abs/2104.07857 访问日期：2026-09-22

[5] Narayanan, D., et al. *Efficient Large-Scale Language Model Training on GPU Clusters Using Megatron-LM*. arXiv, 2021. https://arxiv.org/abs/2104.04473 访问日期：2026-09-22

[7] Chowdhery, A., et al. *PaLM: Scaling Language Modeling with Pathways*. arXiv, 2022. https://arxiv.org/abs/2204.02311 访问日期：2026-09-22

[9] Dao, T., et al. *FlashAttention: Fast and Memory-Efficient Exact Attention with IO-Awareness*. NeurIPS, 2022. https://arxiv.org/abs/2205.14135 访问日期：2026-09-22

[10] Dao, T. *FlashAttention-2: Faster Attention with Better Parallelism and Work Partitioning*. ICLR, 2024. https://arxiv.org/abs/2307.08691 访问日期：2026-09-22

[11] Kwon, W., et al. *Efficient Memory Management for Large Language Model Serving with PagedAttention*. SOSP, 2023. https://arxiv.org/abs/2309.06180 访问日期：2026-09-22

[12] vLLM Team. *vLLM Documentation and Source Repository*. https://github.com/vllm-project/vllm 访问日期：2026-09-22

[13] Yu, G.-I., et al. *Orca: A Distributed Serving System for Transformer-Based Generative Models*. OSDI, 2022. https://www.usenix.org/conference/osdi22/presentation/yu 访问日期：2026-09-22

[14] Agrawal, A., et al. *Taming Throughput-Latency Tradeoff in LLM Inference with Sarathi-Serve*. OSDI, 2024. https://www.usenix.org/conference/osdi24/presentation/agrawal 访问日期：2026-09-22

[15] Zhong, Y., et al. *DistServe: Disaggregating Prefill and Decoding for Goodput-optimized Large Language Model Serving*. OSDI, 2024. https://www.usenix.org/conference/osdi24/presentation/zhong 访问日期：2026-09-22

[16] NVIDIA. *TensorRT-LLM Documentation*. https://nvidia.github.io/TensorRT-LLM/ 访问日期：2026-09-22

[17] NVIDIA. *Triton Inference Server Documentation*. https://docs.nvidia.com/deeplearning/triton-inference-server/ 访问日期：2026-09-22

[18] PyTorch. *Fully Sharded Data Parallel Documentation*. https://pytorch.org/docs/stable/fsdp.html 访问日期：2026-09-22

[19] NVIDIA. *NCCL Documentation*. https://docs.nvidia.com/deeplearning/nccl/ 访问日期：2026-09-22

[20] Kubernetes. *Schedule GPUs*. https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/ 访问日期：2026-09-22

[21] Verma, A., et al. *Large-scale Cluster Management at Google with Borg*. EuroSys, 2015. https://research.google/pubs/large-scale-cluster-management-at-google-with-borg/ 访问日期：2026-09-22

[22] Beyer, B., et al. *The Site Reliability Workbook*. O'Reilly, 2018. https://sre.google/workbook/table-of-contents/ 访问日期：2026-09-22

[23] Sigelman, B. H., et al. *Dapper, a Large-Scale Distributed Systems Tracing Infrastructure*. Google Research, 2010. https://research.google/pubs/dapper-a-large-scale-distributed-systems-tracing-infrastructure/ 访问日期：2026-09-22

[24] OpenTelemetry Authors. *OpenTelemetry Documentation*. https://opentelemetry.io/docs/ 访问日期：2026-09-22

[25] Prometheus Authors. *Prometheus Documentation*. https://prometheus.io/docs/introduction/overview/ 访问日期：2026-09-22

[27] Zaharia, M., et al. *Resilient Distributed Datasets: A Fault-Tolerant Abstraction for In-Memory Cluster Computing*. NSDI, 2012. https://www.usenix.org/legacy/events/nsdi12/tech/full_papers/Zaharia_new.pdf 访问日期：2026-09-22
