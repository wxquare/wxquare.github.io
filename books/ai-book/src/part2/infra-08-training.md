# 第8章 训练平台与数据管线

## 如何稳定、经济地生产模型能力？

上一章建立了大模型 Infra 的全局资源模型。本章聚焦训练阶段：如何把海量数据稳定地送到 GPU，如何选择数据并行、张量并行、流水线并行和专家并行，如何让通信与计算重叠，以及如何在节点故障、版本变化和硬件变化后恢复到一个可解释的训练状态。对于后端工程师，训练平台可以理解为一个持续运行很久、拥有复杂状态、对吞吐极其敏感的分布式任务系统。

训练 Infra 的目标不是让脚本在一台机器上运行，而是让“有效 token 数/时间”可预测、可扩展、可恢复。一个训练 step 的耗时可以拆成数据读取、host 到 device 拷贝、forward、backward、梯度同步、optimizer update 和 checkpoint。任何阶段成为瓶颈，继续增加 GPU 都只会增加成本。Megatron-LM 的模型并行研究、DeepSpeed 与 ZeRO 的状态分片，以及 PyTorch FSDP 的工程实现，都说明大模型训练本质上是模型状态、通信拓扑和故障恢复共同决定的系统问题 [1][3][5][6]。

本章按训练作业生命周期组织：先定义作业对象和控制面，再让数据以可复现方式进入 GPU，随后选择并行策略和 kernel/通信优化，接着用 checkpoint 与弹性恢复保护长时间运行，最后用调度、成本、观测和发布门禁把训练结果转换成可交付制品。

| 生命周期阶段 | 核心对象 | 主要风险 | 交付证据 |
|:---|:---|:---|:---|
| 提交 | job spec、experiment、run | 配置可变、身份不清 | spec hash、owner、预算 |
| 取数 | dataset manifest、token shard | 数据漂移、重复、坏 shard | manifest、统计、样本 hash |
| 计算 | rank、parallel group、step | 通信瓶颈、NaN、OOM | step trace、loss、资源曲线 |
| 保存 | checkpoint、optimizer state | 半写入、缺元数据、不可恢复 | COMMITTED 标记、校验和、恢复 probe |
| 调度 | allocation、quota、priority | 资源碎片、抢占、成本失控 | 调度事件、配额、成本报告 |
| 发布 | model artifact、eval report | 训练态和推理态混淆 | 制品 manifest、评估门禁、注册记录 |

## 8.1 训练平台的对象模型与控制面

### 从训练脚本到训练作业

训练脚本只描述计算图，而平台还要管理 job、experiment、run、checkpoint、dataset、artifact 和 allocation。experiment 表示一个研究问题，run 是一次具体配置执行，allocation 表示实际申请到的节点与 GPU，checkpoint 是某个逻辑 step 的可恢复状态。若这几个概念都用一个目录名代替，稍后就无法回答“这个模型由哪个数据版本训练、失败重试了几次、最终使用了哪些 GPU”。

训练控制面应维护一份不可变的 job spec，至少包括模型代码版本、tokenizer、数据 manifest、采样权重、global batch、sequence length、优化器、精度、随机种子、并行度、容器镜像、依赖锁定文件、checkpoint 策略和预算。提交时生成 job_id 与 spec hash；相同 hash 的重试可以复用实验身份，但不应覆盖已有 run。修改学习率、数据混合或并行度，都应产生新的 spec 和新的 run。

Ray 把任务执行、资源声明和分布式运行时分开，为控制面设计提供了有用的抽象 [16]。Kubernetes 的 GPU device plugin 可作为资源分配基础，但训练平台还要补充 gang scheduling、拓扑亲和性、节点健康和优先级 [18]。Borg 的实践说明，配额、优先级、抢占和故障域不应由每个任务自己实现，而应由集群控制面统一管理 [17]。

### 控制面与数据面的边界

控制面负责提交、排队、分配、启动、心跳、停止、重试、恢复和归档；数据面负责 batch 生成、forward、backward、通信和状态写入。控制面需要幂等与可审计，数据面需要低延迟与高吞吐。用户重复提交相同 idempotency key 时，控制面应返回已有 job，而不是再次申请一组 GPU；worker 失联时，控制面应通过 lease 和 heartbeat 判断是否可以回收资源。

控制面不能把所有运行时细节写进数据库事务。一个训练 step 可能持续数秒，checkpoint 可能持续数分钟，控制面应通过事件记录阶段和状态，而不是长时间持有锁。推荐的状态包括 CREATED、QUEUED、ALLOCATING、STARTING、RUNNING、CHECKPOINTING、RECOVERING、SUCCEEDED、FAILED、CANCELLED。状态迁移需要校验前置状态，重复的完成事件必须幂等，旧 worker 迟到的心跳不能把已经回收的作业改回 RUNNING。

### 训练运行时的身份与配置

每个 rank 都要知道 global rank、local rank、world size、node rank、数据分片范围和通信组。配置不能只依赖环境变量，因为环境变量容易在重试或跨集群迁移中变化。平台应生成一个 rank manifest，记录 rank 到节点、GPU、网卡和并行维度的映射，并将它与 checkpoint 和性能报告关联。

运行时配置要区分用户意图和平台推导值。用户声明 global batch、目标 sequence length 和模型并行上限，平台推导 micro-batch、gradient accumulation、通信组和数据 worker 数。推导过程应可解释：为什么某个 batch 被降低，为什么 tensor parallel 只能取 8，为什么某个节点被排除。否则调度器虽然自动化，却让训练结果变得不可理解。

## 8.2 数据管线：让 GPU 持续获得有效 token

### 数据的分层与版本

数据不应直接从原始文档流入 GPU。常见分层是 raw document、clean document、deduplicated corpus、token shard、packed sample 和 batch。raw 层保留来源与授权信息；clean 层完成编码、语言识别、格式过滤和敏感信息处理；deduplicated 层处理精确与近似重复；token shard 适合高吞吐读取；packed sample 负责把多个短样本拼成固定窗口；batch 则绑定某个 run 的 seed 和采样策略。

每一层都应有 manifest、schema、统计和 hash。统计至少包含文档数、有效 token、语言比例、长度分布、重复率、过滤率、错误率和采样权重。数据版本变化不能只改一个路径名，因为路径可能指向可变对象存储。训练 job 保存 manifest hash 后，任何人都能重新定位具体 shard；数据删除或授权撤回时，也能找到受影响的 run。

### 数据清洗与去重的系统代价

清洗规则会影响质量、数据量和训练成本。过强过滤可能丢失少数语言、代码和长文档；过弱过滤会引入乱码、模板重复、广告和隐私。去重也不是只做整文档 hash：同一网页的导航、模板、转录和代码片段可能构成近重复，需要在 token n-gram、文档 embedding 或 MinHash 层面建立策略。不同数据源的去重阈值应记录在 manifest，而不是藏在一次性脚本里。

数据质量检查可以在写入 token shard 前做，也可以在训练时抽样做。前者节省训练资源，后者能发现分片损坏和分布漂移。建议把抽样样本的统计摘要写入训练报告，原始内容只在有权限的调试流程中访问。数据处理应有失败隔离：一个坏 shard 进入 quarantine，不应让整个预处理 DAG 重做。RDD 提出的分区、血缘和容错思想可用于理解这种数据层设计 [15]。

### 读取路径与预取

GPU 训练的读取路径通常经过对象存储或分布式文件系统、节点缓存、CPU 内存、Pinned Memory 和 GPU 显存。每一层都要明确缓存大小、并发、超时和失败策略。直接从远端对象存储读取小文件会产生过多请求；把所有 token shard 复制到每台节点又会放大存储成本。更合理的方案是大 shard、顺序读取、节点级缓存和有限的预取队列。

预取深度要由实测决定。队列太浅，网络抖动会直接暴露到 GPU；队列太深，会占用 CPU 内存并把错误延迟到很久以后。数据 worker 应暴露等待时间、读取吞吐、解压时间、队列长度、坏样本数和重试次数。训练 worker 应能区分“没有数据”“数据损坏”“数据服务超时”和“GPU 处理慢”，否则所有问题都会被归类为 GPU 利用率下降。

### Packing、padding 与有效 token

固定 sequence length 便于 kernel 和 batch，但短样本 padding 会浪费计算。packing 可以把多个样本拼进一个窗口，提高有效 token 比例，却要求保存样本边界、attention mask 和 loss mask。若样本跨窗口被截断，还要明确是否允许跨样本上下文。训练报告不应只记录 token 数，应同时记录 padded token、有效 token 和 loss token。

有效 token/s 是训练 Infra 的重要指标：

```text
effective_tokens_per_second
  = loss_tokens / wall_clock_seconds
```

如果一个优化把 padding 降低但引入复杂的数据准备，它可能提高理论有效率却降低真实吞吐。评估时要同时展示 device tokens/s、loss tokens/s、GPU busy 和数据等待占比。PaLM 与 Llama 3 的规模化训练经验说明，数据配方、训练时间和系统吞吐必须共同考虑，不能只看模型参数量 [9][10]。

## 8.3 并行策略：把模型放进集群而不是把集群堆在模型外面

### 数据并行与梯度同步

数据并行让每个 rank 持有模型副本，处理不同 micro-batch，然后通过 all-reduce 同步梯度。它实现简单、扩展直接，但完整复制参数、梯度和 optimizer state 的内存成本很高。随着模型变大，单纯增加数据并行 rank 会先遇到单卡显存上限，再遇到梯度同步和小 batch 的效率问题。

global batch 等于 micro-batch、数据并行度和梯度累积步数的乘积。改变任何一个因素都可能改变优化轨迹。平台必须记录真实的 global batch 和每步有效 token，而不是只记录 dataloader batch size。NCCL 的 all-reduce 适合梯度聚合，但通信是否能与 backward 重叠，取决于 bucket 大小、梯度 ready 顺序和拓扑 [13]。

### 张量并行与流水线并行

张量并行把单个线性层或 attention 的矩阵切到多个 GPU，单层计算需要 collective 通信。它适合高速互联的同机 GPU，tensor parallel degree 通常受 NVLink、PCIe 和显存约束。Megatron-LM 通过对 Transformer 层做模型并行，使单卡无法容纳的模型能够训练 [1]。但 TP 越大，每个 token 的通信次数和同步敏感性也越高。

流水线并行把不同层放到不同 stage，输入 micro-batch 依次通过 stage。它减少单卡参数占用，却产生 pipeline bubble、调度复杂度和 stage 不均衡。1F1B 等调度可以降低 bubble，但不能消除最长 stage 决定吞吐的事实。层划分应考虑参数、激活、通信和 kernel 时间，而不是简单平均层数。若某些层因 MoE 或视觉模块更重，静态平均会造成严重 straggler。

### ZeRO 与 FSDP

ZeRO 逐步分片 optimizer state、gradient 和 parameter，减少每个 rank 的冗余；FSDP 在 PyTorch 中以参数分片、all-gather 和 reduce-scatter 实现类似目标 [3][6]。分片让模型能在更少的显存中训练，却把参数获取和释放加入了关键路径。正确选择需要测量通信与计算重叠、参数 prefetch、CPU offload 和 activation checkpoint 的组合。

ZeRO-Infinity 进一步利用 CPU 和 NVMe 扩展可用内存 [4]。卸载能解决容量问题，但 PCIe、NVMe 和网络带宽可能使 step 时间显著增加。平台不能把“能启动”当作成功，而应比较每百万有效 token 的成本、恢复时间和尾部抖动。对低利用率实验，offload 可能比增加 GPU 更经济；对大规模生产训练，它可能成为不可接受的同步瓶颈。

### MoE 与专家并行

MoE 用路由器把 token 分发到少数专家，激活参数少于总参数。GShard 和 Switch Transformer 展示了稀疏专家模型的扩展方式 [7][8]。Infra 需要处理 token dispatch、专家容量、all-to-all 通信、负载不均衡和 token drop。专家数增加并不自动提高效率；如果路由把大部分 token 集中到少数专家，热门 expert 成为瓶颈，其他 GPU 却空闲。

专家并行的容量因子决定每个 expert 能接收多少 token。容量太小会丢 token，容量太大则浪费显存并增加通信。训练平台要记录每个 expert 的 token 数、溢出率、路由熵、all-to-all 时间和 rank 方差。MoE checkpoint 还要保存专家布局和路由相关配置，不能只保存最终权重，否则恢复时的并行映射可能不一致。

### 并行度搜索

并行度选择可以视为约束优化。先满足单层参数和激活能放入显存，再在目标拓扑上选择 TP；根据层数和 bubble 选择 PP；剩余 GPU 用 DP 扩展吞吐；MoE 再选择 EP。每个候选组合都需要小规模真实 workload 测试，测量 step time、通信比例、显存峰值和扩展效率。理论 FLOPs 只能给出上界，不能替代实测。

并行配置一旦写入 checkpoint，就成为恢复契约的一部分。支持 reshard 的平台应保存逻辑参数名、分片范围和布局版本，恢复时由转换器生成新的 shard；不支持 reshard 时，应让调度器保证相同的 world size 和并行度。静默改变并行配置会让 optimizer state 与参数分片错位，是最危险的恢复错误之一。

## 8.4 Kernel、通信与数值稳定性

### 从 FLOPs 到 IO

Transformer 训练不是只受 FLOPs 限制。中间张量写入 HBM、kernel launch、非连续 layout、padding 和同步都可能成为瓶颈。FlashAttention 用 IO-aware tiling 降低 attention 的 HBM 读写，并保持精确计算 [11]；FlashAttention-2 进一步改善并行和 work partitioning [12]。采用新 kernel 时，要验证端到端 step、显存峰值、不同长度、不同 batch、不同 GPU 架构和数值误差。

优化应遵循基线—假设—变更—测量—回滚链路。先用 profiler 找到热点，再决定融合、重排、编译、量化或缓存。一个 kernel benchmark 提速，不代表真实训练提速；它可能把瓶颈转移到通信或数据处理。性能结果必须与模型版本、输入 shape、CUDA、driver、编译器和环境变量绑定。

### 集合通信与拓扑

all-reduce 用于同步梯度，all-gather 用于取得分片参数，reduce-scatter 可把聚合结果直接保留为分片。NCCL 会根据 NVLink、PCIe、InfiniBand 和网卡拓扑选择通信路径 [13]。训练平台应保存拓扑快照，集群变更后重新测量集合通信。通信日志至少包括 collective 类型、字节数、耗时、rank 方差、是否重试和是否与计算重叠。

当通信时间上升时，要区分带宽不足、延迟主导、消息过小、最慢 rank、网络拥塞和同步顺序错误。单机 microbenchmark 只能说明局部能力，必须用真实 batch 验证是否能解释 step 时间。跨节点 TP 往往比同机 TP 更敏感；若通信不能隐藏在计算之后，减少并行度可能比扩大集群更快。

### 混合精度与异常处理

BF16、FP16、FP8 和更低精度可以降低显存、通信和计算成本，但会改变数值误差。训练系统需要监控 loss spike、梯度范数、overflow、underflow、NaN、Inf 和异常 rank。master weights、optimizer moments、归一化和输出层可能需要更高精度；不能把 dtype 当作一个全局开关。

出现 NaN 时，自动重试同一个 batch 通常没有意义。平台应保存触发 step 的输入摘要、参数统计、梯度范数、精度配置和 kernel 版本，然后暂停或回滚到最近健康 checkpoint。若直接继续运行，坏状态可能覆盖所有可恢复版本。异常检测应与 checkpoint 保留策略绑定，至少保留最后一个已验证健康点。

## 8.5 Checkpoint、断点恢复与弹性

### Checkpoint 的完整状态

可训练 checkpoint 通常包含模型参数、master weights、梯度、optimizer moments、学习率调度器、gradient scaler、随机数状态、数据游标、训练配置、并行拓扑、tokenizer、数据 manifest、代码版本和评估结果。只保存权重足以做推理，却不能保证继续训练等价。对于分片训练，文件名不是状态语义，必须用 manifest 描述逻辑参数、shard 范围、dtype、shape 和校验和。

### 两阶段提交

推荐用临时路径加完成标记实现 checkpoint 的两阶段提交。各 rank 先写带有 job_id、step 和版本号的临时 shard，计算大小与 checksum；协调者确认全部 shard 到齐、元数据完整、抽样读取成功后，再写 COMMITTED 标记。恢复程序只读取已提交版本。训练进程在写入中途退出时，残留临时目录可以异步清理，不会被误认为最新状态。

完成标记应包含 world size、TP、PP、DP、EP、dtype、模型 schema、数据 manifest hash、tokenizer hash 和 global step。恢复后要执行小 batch forward、loss sanity check、参数统计和数据可读性检查。校验和只能证明字节完整，不能证明参数、优化器和数据游标的语义对应。

### 保存频率与丢失工作量

保存越频繁，故障后丢失的训练工作越少，但 I/O 和同步开销越高。可以根据故障率、恢复时间、checkpoint 成本和 GPU 单价估计总成本。大规模训练还要区分可恢复 checkpoint、评估 checkpoint 和发布 checkpoint：可恢复点追求快速重启，评估点需要质量报告，发布点需要转换、签名和兼容性检查。不要用发布制品替代高频恢复点。

异步 checkpoint 可以让训练继续计算，但需要处理内存快照、写入期间的状态一致性和 GPU/CPU 拷贝。若 checkpoint 线程读取正在变化的 optimizer state，得到的可能是跨 step 混合状态。安全做法是使用冻结的 state snapshot、copy-on-write 或在逻辑 step 边界建立一致快照。异步并不意味着可以取消一致性约束。

### 故障分类与恢复策略

常见故障包括 GPU Xid、节点重启、网络分区、NCCL hang、对象存储限流、文件系统不可写、容器驱逐、OOM、NaN 和数据 shard 损坏。单卡故障可能需要替换节点并重建 communicator；存储抖动可以退避重试；NaN 应停止并恢复健康点；rank 卡死需要 watchdog 和全局超时。不同故障必须有不同的 runbook，不能把“自动重启”当作统一方案。

弹性恢复可以等待原节点、替换节点后保持 world size、缩小 world size 继续，或回滚并重新分配。改变 world size 可能改变 global batch、学习率、数据顺序和 optimizer 行为。平台应让恢复策略显式化，并在报告中记录丢失步数、重算 token、恢复耗时和最终质量差异。SRE 的错误预算思想可用于决定何时继续重试、何时停止作业并人工处理 [19]。

## 8.6 训练调度、成本与多租户

### Gang scheduling 与故障域

分布式训练需要一组 GPU 同时可用，部分分配往往只能让 job 长期 pending。调度器应识别 gang，并根据节点、GPU、NVLink、网络和本地数据做拓扑感知。大 job 需要反碎片化，小 job 可以填充空洞；训练、推理、评估和开发应有不同资源池与优先级。抢占时要考虑 checkpoint 是否完成，不能在不可恢复的阶段直接杀死作业。

### 配额、优先级与公平

配额应同时按 GPU 数、GPU 小时、显存、优先级和项目预算表达。一个团队占用少量高端 GPU 训练很久，可能比另一个团队使用更多低端 GPU 更昂贵。成本系统要记录预约、实际使用、空闲等待、通信等待、恢复重算、评估和 checkpoint。按有效 token 和成功实验归因，才能比较不同团队的真实效率。

多租户环境应隔离数据凭证、checkpoint 路径、日志、缓存和调试权限。训练 worker 只获得访问具体 manifest 和 shard 的短期凭证，不能拥有整个 bucket 的写权限；checkpoint 目录与原始数据隔离；样本抽样和错误日志默认脱敏。OpenTelemetry 与 Prometheus 可以提供统一的运行指标和追踪基础，但敏感数据是否记录仍需平台策略 [20][21][22]。

### 成本预算与容量计划

训练计划应在启动前估算 token、step、GPU 小时、checkpoint 次数、评估次数和重试余量。理想扩展效率公式为：

```text
E(N) = throughput(N) / (N × throughput(1))
```

当扩展效率随 GPU 数快速下降时，继续扩大集群可能比优化数据、kernel 或并行策略更贵。平台应把扩展曲线作为门禁，记录有效 token/s、通信比例、数据等待、峰值显存和节点故障恢复。PaLM 和 Megatron-LM 的规模化经验表明，训练规模越大，系统效率和故障管理越是模型质量的一部分 [1][2][9]。

## 8.7 训练观测与验收

### 指标、日志与追踪

训练指标包括 loss、学习率、梯度范数、有效 token、device token、GPU 利用率、显存、kernel 时间、通信、数据等待、checkpoint、恢复和节点健康。指标要按 rank、节点、数据桶和并行维度聚合，同时保留最大值和方差，避免平均值掩盖一个慢 rank。日志记录错误上下文，trace 连接控制面、数据服务、worker 和 checkpoint 服务。

Dapper 的分布式追踪思想适合把训练控制面和数据面关联起来 [22]。训练 trace 不应默认记录完整样本和隐私文本，而应记录 manifest、shard、batch、step、request id、错误类型和 hash。OpenTelemetry 提供跨组件的语义约定，Prometheus 适合时间序列与告警；二者结合后，平台能从“GPU 下降”追到“某类 shard 解压变慢”或“某个 rank 通信抖动”。

### 验收矩阵

训练 Infra 的验收至少包含五类测试：数据正确性、性能扩展、故障恢复、数值稳定和安全治理。数据正确性验证 token 统计、样本边界、loss mask、manifest 和重复率；性能验证单机、单节点、多节点扩展曲线；恢复验证中断、节点替换、checkpoint 损坏和数据不可读；数值验证 loss 连续、梯度无异常和不同精度在容差内；治理验证权限、审计、成本和数据删除。

每项测试都要固定模型 hash、数据 manifest、容器镜像、硬件、驱动、并行度、global batch 和测试时间。报告不仅写“通过”，还要保存原始指标与失败样本。若性能提高但 loss 曲线异常，不能接受；若恢复成功但重复了大量数据，也不能只算功能通过。训练平台的交付标准是可解释、可重现和可恢复，而不是某一次 demo 的最高吞吐。

## 8.8 数据一致性、采样与训练语义

### 训练数据不是普通消息队列

把数据管线简单理解为“从存储读取消息”会遗漏训练语义。消息队列通常关心至少一次、至多一次或恰好一次消费；训练还要关心样本顺序、采样权重、样本边界、token 数和 epoch 定义。一个 batch 被消费后，worker 崩溃并重放它，计算结果可能不同，因为 dropout、混合精度和并行归约都可能变化。平台需要先定义业务可接受的语义：预训练可以容忍少量重复，监督微调可能要求样本不重复，对齐数据则可能需要精确记录每次消费。

数据服务应使用不可变 shard 和可定位游标。游标不应只表示第几个 batch，因为 worker 数、packing 策略和 shard 版本变化会使 batch 编号失去意义。更稳妥的游标包括 manifest hash、shard id、sample offset、token offset、packing 状态和采样器状态。恢复时如果这些字段与当前运行时不兼容，要明确走 reshard、重放或拒绝恢复，而不是静默从最近位置开始。

### 混合数据配方

预训练通常混合网页、代码、书籍、论文、对话和多语言数据。混合比例可以按文档数、字节、token 或有效 loss token 定义，四种定义结果不同。平台应把权重归一化过程写进 manifest，并在运行中统计实际消费比例。若代码样本平均更长，按文档数设置 20% 并不等于按 token 得到 20%；若大量样本被过滤或截断，声明配方与真实训练配方可能相差很大。

采样器还要处理数据源耗尽、动态增量和阶段切换。一个数据桶耗尽后，是循环采样、重新归一化还是提前结束 epoch，必须有明确策略。数据版本切换时，可以从某个 global step 开始使用新权重，但要把切换事件写入 trace 和 checkpoint。否则训练损失曲线的变化无法解释，出现质量回退时也无法定位是模型代码还是数据配方引起。

### 评估数据隔离

训练、验证和测试数据需要物理或逻辑隔离。去重系统如果只在训练集内部工作，可能把评估集内容保留在训练数据中；如果把所有数据一起去重，又可能泄露测试集的存在和分布。数据 manifest 应记录 split 生成规则、去重边界和时间版本。评估服务读取只读版本，不应让训练 worker 获得写权限。

评估结果要绑定 checkpoint、数据版本、采样配置和代码版本。在线 dashboard 中只显示一个“验证集 loss”很容易把不同实验混在一起。对于长周期训练，建议定期保存小型固定 probe 集和较大的正式评估集：probe 用于快速检测 loss、格式和数值异常，正式集用于能力、偏差和安全门禁。两者都不能被训练程序修改。

## 8.9 通信性能的工程推导

### 通信量账本

每种并行策略都可以建立通信量账本。数据并行主要产生梯度同步；张量并行在层内产生 all-reduce 或 all-gather；流水线并行传输激活和梯度；MoE 产生 token dispatch 的 all-to-all；FSDP 需要在计算前后获取和释放参数。平台应估计每一步的字节数、消息数量和同步次数，再用真实链路带宽与延迟估算下界。

如果一个层的计算时间小于一次 collective 的延迟，就算理论 FLOPs 很低，扩大并行度也不会有收益。相反，大矩阵乘法可以把通信隐藏在计算中。通信量账本应按 sequence length、micro-batch、hidden size、并行度和 dtype 计算，并在 profiler 中验证。出现差异时，重点查 padding、bucket、参数 prefetch、梯度累积和 collective 是否被意外串行化。

### Overlap 的前提

通信与计算重叠需要独立 stream、合理的梯度 bucket、正确的依赖和足够大的计算块。bucket 太小会产生大量 collective 和 launch，太大则要等很久才开始同步；通信 stream 与计算 stream 的依赖设置错误时，表面上有两个 stream，实际仍然串行。平台应在报告中分别列出 kernel 时间、可重叠通信、不可重叠通信和同步等待。

重叠还会影响显存。为了提前 all-gather 参数，系统可能要同时保留更多参数 shard；为了让 backward 与通信并行，梯度 bucket 会延迟释放。显存预算不能只按模型权重计算，而要加入 overlap buffer、activation checkpoint、通信 buffer 和 dataloader staging。一个在显存充裕机器上有效的优化，迁移到更小 GPU 后可能反而触发 OOM。

### NUMA、PCIe 与设备亲和

多 GPU 节点常常有多个 CPU socket、PCIe root complex 和网卡。数据 worker、Pinned Memory、GPU 和网卡绑定不合理，会让 host 到 device 拷贝经过远端 NUMA，吞吐下降且抖动增加。启动器应根据拓扑生成 CPU affinity、GPU affinity 和 NIC affinity，并在运行时记录映射。不能只依赖容器默认的 CPU 集合，因为调度器可能把进程放在与 GPU 不同的 socket。

故障排查时要把拓扑作为版本化环境的一部分。换一批节点、升级 BIOS、改变网卡固件或调整 MIG，都可能改变通信路径。NCCL 提供了拓扑感知能力，但应用还需要保存 NCCL 环境变量、通信算法和日志摘要 [13]。当扩展曲线突然下降时，拓扑快照往往比训练代码更能解释原因。

## 8.10 Checkpoint 转换与模型发布

### 训练状态与推理制品分离

训练 checkpoint 的目标是继续优化，推理制品的目标是快速加载和稳定服务。训练状态包含 optimizer 和分片元数据，推理制品通常只包含参数、tokenizer、模板、量化和运行时配置。两者应使用不同目录、权限和生命周期，但通过一个发布 manifest 关联。直接把训练目录挂载到 serving，容易暴露不必要状态，也容易让线上误读一个未完成版本。

转换器要处理参数命名、分片布局、dtype、权重共享、词表、位置编码和模型 schema。转换后应做参数总量、shape、均值方差、hash 和小 batch 输出对比。对量化模型，还要比较校准集质量、最大误差、生成稳定性和特殊 token。转换脚本要固定版本并产生日志，不能依赖某个开发者机器上的临时命令。

### 发布门禁

发布门禁至少包括：制品完整性、模型加载、tokenizer 与模板兼容、固定 probe 输出、离线质量、显存峰值、冷启动时间和目标硬件性能。一个模型可能 loss 更低，却因为 tokenizer 改变导致线上输入边界不同；量化可能节省显存，却造成结构化输出失败；新的 kernel 可能提高吞吐，却在最大长度下产生 NaN。门禁应同时检查功能、质量、性能和安全。

发布后的 artifact 要签名并登记来源。注册表记录训练 run、数据 manifest、checkpoint、转换器版本、量化配置、engine build、评估结果和审批者。回滚时根据旧 manifest 恢复完整制品，而不是只把 image tag 改回旧值。推理平台的模型仓库与训练平台的 checkpoint 仓库可以有不同接口，但必须共享不可变 hash。

### 迁移训练与分布式重分片

当训练从 64 张卡迁移到 128 张卡，可能需要改变 DP 或 pipeline stage；当 GPU 型号变化，可能需要调整 batch、dtype 或 checkpoint 读取路径。支持迁移的格式应将逻辑参数与物理 shard 分离，转换器根据目标拓扑重新分片。迁移后还要验证 optimizer state 的分片和数据游标是否仍然一致。

若系统不支持任意重分片，应在调度时把 checkpoint 的恢复约束传递给资源调度器，例如 world size、TP、PP、显存下限和 GPU 架构。调度器找不到兼容资源时宁可保持 pending，也不能分配一个表面满足 GPU 数量但无法恢复的集群。恢复兼容性是调度约束，而不是作业启动后的异常处理。

## 8.11 训练平台的故障演练与治理

### 故障注入

训练平台至少要演练五类注入：杀死一个 rank、关闭一台节点、阻断集合通信、让对象存储返回超时、损坏一个 checkpoint shard。每次注入要记录检测时间、作业状态、用户可见影响、自动恢复、重算 token、恢复后的 loss 和资源释放。若系统只在单进程退出时能恢复，而一个 rank 卡死会让其他进程永久等待，那么它并不具备生产级弹性。

故障注入不能只发生在非高峰环境。高峰期的队列、配额和发布动作会改变恢复路径。可以先在隔离资源池做小规模演练，再在生产影子作业验证指标与告警，最后安排可控的真实节点维护。SRE Workbook 强调演练、错误预算和事故复盘之间的闭环，这些方法同样适用于训练平台 [19]。

### 数据与模型治理

训练数据需要来源、授权、保留期、删除流程和审计。数据处理 worker 获得最小读权限，输出 shard 写入独立路径；日志中的样本文本默认脱敏；调试抽样需要临时授权。模型 checkpoint 可能记忆训练数据或包含内部能力，因此下载、复制、转换和对外发布都应登记。

治理信息不应只存在文档中。manifest、job spec、checkpoint manifest 和发布记录都应机器可读，评估系统和权限系统可以据此阻断不合规任务。例如数据 license 不允许某种用途时，控制面在分配 GPU 前拒绝 job；checkpoint 使用了已撤回的数据版本时，注册表阻止发布。治理越晚介入，返工成本越高。

### 人工接管边界

自动恢复不是越多越好。短暂的对象存储超时可以自动退避；连续 NCCL hang、NaN、checkpoint 校验失败和数据权限错误应暂停并通知负责人。平台要定义最大重试次数、指数退避、预算上限和升级联系人。重试次数超过阈值后，系统保留现场、停止覆盖健康 checkpoint，并生成包含 job spec、拓扑、最后健康 step 和错误 trace 的诊断包。

人工接管也要有明确操作。允许负责人选择回滚到某个 checkpoint、跳过坏 shard、改变资源池或终止作业，但每个动作都产生审计事件。不能为了“先跑起来”直接修改共享存储或删除现场。事后复盘应区分代码缺陷、数据缺陷、平台缺陷和容量规划缺陷，并把修复加入下一次验收。

## 8.12 一个可执行的训练 Infra 交付模板

### 交付前配置

交付前应生成一份配置摘要：模型和代码 hash、tokenizer、数据 manifest、数据配方、global batch、sequence length、dtype、并行度、节点类型、网络拓扑、checkpoint 频率、恢复策略、预算和告警阈值。摘要由控制面生成并随 run 保存，避免用户填写一份配置、启动器再隐式覆盖另一份配置。

### 小规模到目标规模的阶梯验证

第一阶验证单 GPU 的数据、loss、checkpoint 和恢复；第二阶验证单节点的并行组、通信和显存；第三阶验证多节点的扩展曲线、数据供给、故障替换和 reshard；第四阶才运行目标规模。每一阶都应使用相同的模型和数据语义，区别只在资源规模。这样出现问题时可以判断是算法、单机 kernel、跨机网络还是调度器导致。

### 交付后的持续检查

上线后持续检查有效 token/s、GPU 利用率、通信比例、数据等待、loss、checkpoint 成功率、恢复时间和成本。每次驱动、CUDA、网络、文件系统、数据格式、模型代码或并行配置升级都触发一轮基线对比。若只在初次验收时测量，系统随环境变化退化却不会被发现。

训练 Infra 的最终验收不是“脚本跑完”，而是：给定相同的 job spec，平台可以在可接受的时间内分配资源、持续供给数据、达到可解释吞吐，在故障后从一致状态恢复，在发布时生成可验证制品，并且每个资源和质量变化都能追溯到具体版本。这个闭环建立后，后续推理 Infra 才能可靠地消费训练产物。

### 训练数据的访问控制实现

数据访问控制要覆盖控制面和数据面。控制面在创建 job 时校验项目、用途、数据 license、地域和保留期限；数据面在 worker 获取 shard 时发放短期、最小范围的凭证。凭证不应允许列出整个 bucket，也不应允许 worker 将训练数据写回原始路径。对象存储、缓存和本地临时目录的权限要保持一致，任务结束后回收临时凭证并清理不再需要的缓存。

训练日志常常包含样本文本、异常 token 和模型输出，必须按照数据等级处理。默认记录 hash、长度、语言和错误类型；完整文本只在经过授权的诊断任务中短暂保存。将 prompt 或样本直接作为 metric label 更危险，会造成高基数、日志泄露和成本膨胀。观测字段应有允许列表、脱敏器和保留策略，OpenTelemetry 的统一上下文只能解决关联问题，不能替代隐私控制 [20]。

### 数据漂移与训练漂移

数据管线在长周期训练中可能发生漂移：上游抓取站点变化、语言比例变化、过滤规则更新、代码仓库增加或某个数据源失效。平台应按时间窗口统计有效 token、语言、长度、来源、重复率和过滤率，并与 job spec 中的目标配方比较。偏差超过阈值时，可以暂停消费、重新生成 shard 或标记为新阶段，不应让训练静默改变分布。

训练指标也会漂移。全局 loss 下降不代表每种语言、代码、长文档和少数任务都改善。将 loss 按数据桶和长度分层，可以发现某个源异常、padding 变多或 tokenizer 处理错误。固定 probe 集用于检测运行时变化，分层正式评估用于决定是否继续训练；两者都需要与 checkpoint 版本绑定。

### 资源泄漏与作业清理

训练作业失败后，常见残留包括 Kubernetes Pod、NCCL 进程、共享内存、Pinned Memory、对象存储临时目录、租约、GPU reservation 和告警。控制面应有终止流程：先发送取消事件，等待 worker 上报 checkpoint 或停止原因，再清理通信组、释放分配、关闭数据流、删除临时文件并写最终状态。清理流程需要超时和人工兜底，不能无限等待一个失联 worker。

资源回收要幂等。重复执行清理不会删除其他 job 的路径，不会释放错误的 GPU，也不会覆盖成功的最终状态。临时目录命名必须包含不可猜测的 run_id 和版本；清理器根据 manifest 校验归属，而不是使用宽泛的通配符。Prometheus 可以监测 reservation 与实际 Pod 的差异，及时发现资源泄漏 [21]。

### 训练作业的可观测事件

除了时间序列指标，还需要结构化事件：JOB_ACCEPTED、ALLOCATION_READY、DATA_STAGE_READY、WORKER_STARTED、STEP_HEALTHY、CHECKPOINT_COMMITTED、NODE_LOST、RECOVERY_STARTED、RECOVERY_SUCCEEDED 和 RUN_FINISHED。每个事件携带 run_id、spec hash、global step、world size、checkpoint version 和时间。这样事故复盘可以重建状态变化，而不是从分散的 stdout 中猜测。

事件系统要处理重复、乱序和延迟。worker 可能在网络恢复后发送旧的 checkpoint 完成事件，控制面必须通过单调 step、版本号和状态机拒绝过期事件。重要事件需要持久化，普通 profiler 数据可以采样。训练控制面不应依赖某个 dashboard 的实时状态作为事实来源，dashboard 只是从事件和指标派生视图。

### 训练性能报告的最小字段

一个可比较的性能报告至少包括：模型参数量、层数、hidden size、序列长度、global batch、micro-batch、有效 token、padding token、GPU 型号、GPU 数量、节点数、TP、PP、DP、EP、dtype、CUDA 和 driver 版本。还要包括 step time、forward、backward、通信、数据等待、checkpoint、MFU 或等价计算效率、显存峰值和扩展效率。

报告必须说明是否包含编译、warmup、验证、checkpoint 和失败重试。不同团队经常使用不同口径，导致“吞吐提升”无法复核。建议把原始 trace、汇总 JSON 和人类可读表格一起保存，并将报告 hash 写入 run metadata。后续优化如果只保留一个漂亮的数字，而没有 workload 与环境，就不应作为平台基线。

### 训练与后训练的资源差异

预训练通常以稳定的大 batch 和长时间运行换取高吞吐；监督微调和偏好优化的 batch 更小、序列长度更不稳定，评估和生成式 verifier 可能成为瓶颈。统一的训练平台可以复用控制面、数据 manifest 和 checkpoint，但应允许 workload profile 声明不同的资源需求。后训练不能直接套用预训练的扩展曲线，因为生成、采样、奖励模型和验证器会改变 GPU 与 CPU 的比例。

例如 RLVR 训练可能需要并行生成多个候选，再执行 verifier；如果只给训练 worker 配 GPU，CPU verifier、队列和网络会成为隐性瓶颈。平台应把辅助模型、评估模型、tokenizer、验证器和数据服务纳入 job graph，统一记录版本与成本。否则看到的“训练吞吐”可能只是主模型 forward 速度，真实的每个有效样本成本却不断上升。

### 训练平台的演进顺序

搭建平台时，优先顺序应是可复现、可观测、可恢复，然后才是极限性能。没有稳定 manifest 和 checkpoint，增加复杂的并行策略只会扩大排错空间；没有扩展曲线和通信 trace，增加 GPU 只会增加账单；没有模型制品与发布门禁，训练结果也不能安全进入 serving。一个小规模但能完整生成 spec、运行、评估、保存、恢复和归档的闭环，比一个只能启动大集群的脚手架更有价值。

在闭环之上再逐步加入 FSDP、MoE、offload、异步 checkpoint、拓扑调度和弹性 world size。每次引入一个复杂机制，都保留一个简单基线，做固定 workload 的 A/B 测试。若新机制只在一个 benchmark 上有效，就不应替代默认路径；若它改善吞吐但增加恢复风险，应把风险写进发布门禁和错误预算。这样平台可以不断吸收新的算法与硬件，而不必反复推倒重来。

### 数据处理 DAG 的重试边界

预处理往往是一个包含下载、解析、过滤、去重、分词、分片和统计的 DAG。每个节点都应有输入版本、输出版本、执行镜像、参数和结果 manifest。失败时只重试失败分区，不重做已经完成且校验通过的分区；上游输出发生变化时，通过血缘关系确定哪些下游需要失效。Spark 的 RDD 和分区容错思想说明，数据处理的可靠性来自可重建的中间结果与明确的血缘，而不是依赖某个常驻进程 [15]。

重试要区分确定性错误和瞬时错误。对象存储超时、节点暂时不可用可以退避；解析器遇到未知编码、schema 不匹配或越过内存上限，应隔离样本并记录原因。无限重试会掩盖坏数据并持续消耗配额。每个分区应有最大重试次数、错误样本上限和 quarantine 路径，数据负责人可以在不影响其他分区的情况下修复并重新运行。

### Tokenizer 的版本契约

tokenizer 不是训练前的普通工具，而是决定样本长度、词表、特殊 token、padding、截断和模型输入的运行时契约。升级 tokenizer 会改变有效 token 数、sequence packing、embedding shape 和训练成本。job spec 必须保存 tokenizer 文件或不可变 artifact hash、normalization 规则、special token、chat template 和最大长度。checkpoint 恢复时如果 tokenizer 不一致，应明确拒绝或执行经过验证的迁移。

tokenizer 处理异常字符时要有统计。未知字符、超长 Unicode 序列、二进制内容、空样本和控制字符都可能造成 token 爆炸或 loss 异常。数据质量报告应包含字符到 token 的膨胀比例，并按语言、数据源和长度分桶。对代码、多语言和数学公式，通用过滤器可能误删重要信息，训练 Infra 需要把这些边界交给数据和算法团队共同确认。

### Activation checkpoint 与内存时间交换

activation checkpoint 通过不保存全部中间激活、在 backward 时重新计算来降低显存。它把空间成本换成计算成本，适合模型大、激活占比高的训练。平台应记录 checkpoint 粒度、重算 FLOPs、显存峰值、step time 和质量。层间切分过粗会导致峰值仍高，切分过细会增加 kernel 和调度开销；不同 sequence length 下最佳粒度也可能不同。

activation checkpoint 与流水线并行、FSDP、混合精度和异步通信会互相影响。重新计算时需要相同的随机状态，否则 dropout 或随机算子会改变梯度；参数已经被释放时，需要再次 all-gather。恢复和性能报告必须保存这些运行时开关，不能只保存一个 enable_activation_checkpoint 的布尔值。对于后训练和生成式训练，还要评估重新计算对采样一致性的影响。

### GPU 利用率的正确解释

GPU busy 只是设备上有 kernel 执行，不代表 kernel 高效。大量小 kernel、低 occupancy、访存瓶颈、等待同步和无效 padding 都可能让 busy 较高而有效吞吐较低。相反，某些通信或数据等待期间 GPU 利用率下降，可能是系统正常的阶段性行为。应结合 SM 利用率、显存带宽、Tensor Core 利用率、kernel occupancy、通信和有效 token 分析。

训练 dashboard 需要同时展示全局和局部。全局 step time 便于看趋势，rank 最大耗时便于看 straggler，数据桶 loss 便于看质量，checkpoint 和恢复指标便于看可用性。指标如果没有 run、step、rank、node、model 和数据版本这些低基数字段，就无法定位；字段过多又会造成监控成本，因而应使用固定标签加结构化日志承载细节。

### 集群升级与基线保护

驱动、CUDA、NCCL、内核、固件和容器基础镜像的升级都可能改变训练结果与性能。升级前应保存固定 workload 的基线，包括吞吐、通信、显存、loss 曲线、异常率和 checkpoint 恢复。升级后用相同 job spec 在同一批或等价节点上运行，并比较允许范围。若只比较速度，不比较 loss 和恢复，可能把数值变化误认为性能提升。

集群升级需要分批和可回退。先用影子作业验证数据和通信，再让低优先级实验迁移，最后处理生产训练。节点标签应标识硬件、driver、NCCL 和固件版本，调度器可以按兼容性选择节点。发生性能回退时，保留旧节点池或旧镜像一段时间，保证正在运行的长任务有安全迁移路径。

### 训练预算的动态控制

作业预算不能只在提交时检查。训练过程中失败重试、恢复重算、评估、checkpoint 和数据重处理都会增加成本。控制面应实时累计 GPU 小时、有效 token、对象存储请求、网络流量和失败次数，并在接近预算时提醒或暂停。用户可以选择“质量优先”“时间优先”或“成本上限”策略，平台据此调整评估频率、checkpoint 频率和资源池，而不是无条件继续运行。

动态预算还需要防止部分成功被误判。一个实验如果只完成了计划 token 的 20%，不能与完整实验直接比较；一个训练 job 超预算后被强制停止，应保存最后健康 checkpoint、最终 loss 和未完成原因。成本报表要把有效训练、恢复重算和无效等待分开，否则优化方向会被错误数据引导。

### 训练平台与数据平台的接口

训练平台向数据平台提出的不是“给我一个路径”，而是一个数据合同：数据版本、schema、授权用途、split、采样权重、tokenizer、分片大小、可用区域、删除策略和质量阈值。数据平台返回 manifest、统计、访问凭证和血缘。合同变更需要版本化，训练平台根据 manifest hash 决定是否允许复用缓存和 checkpoint。

接口还应支持数据不可用和部分可用。一个区域对象存储故障时，训练可以切换到镜像副本，但必须记录数据路径和复制版本；某个 shard 损坏时，可以隔离并跳过，但要报告有效 token 损失。数据平台不应为了满足训练吞吐静默返回旧版本，训练平台也不应为了继续运行绕过授权检查。双方通过显式版本和错误码协作。

### 训练平台与调度平台的接口

训练作业向调度器声明 GPU 数量只是最小信息，还应声明 GPU 类型、拓扑、互联、CPU、内存、网络、临时盘、gang 约束、优先级、可抢占性、预计时长和恢复限制。调度器返回 allocation、rank manifest、租约和故障域。作业启动后若发现拓扑不满足通信要求，应拒绝运行并释放资源，而不是勉强开始再产生低吞吐。

调度器还要向控制面提供可解释事件：pending 的原因是资源不足、拓扑不匹配、配额不足、镜像不可用还是节点不健康。用户看到“排队中”没有帮助，看到“等待 16 张同型号 GPU 且需要同一 NVLink 域”才可以做取舍。透明的 pending reason 能减少人工干预，也能帮助平台发现容量规划和调度策略问题。

### 训练平台与评估平台的接口

评估不是训练结束后手动运行一个脚本，而是 checkpoint 生命周期的一部分。训练平台提交 checkpoint hash、模型 schema、tokenizer、数据 manifest 和安全标签；评估平台返回任务 id、指标、样本报告和质量门禁。评估任务失败、超时或结果不完整时，checkpoint 不能被自动标记为发布候选。

评估结果应区分可比和不可比。改变 tokenizer、模板、数据切分、采样温度或最大输出长度后，指标不应直接与旧版本横向比较。平台保存评估配置和运行时参数，报告中给出质量、延迟、显存和成本的联合视图。只有达到预设门禁且证据完整，模型制品才可以进入推理 Infra 的注册和灰度流程。

### 大规模训练中的尾部任务

平均 step time 不能代表整个训练的体验。某些 step 可能因为 checkpoint、数据刷新、节点抖动、编译或对象存储限流而明显变慢；长尾 step 会决定 wall-clock 完成时间和 GPU 成本。平台应记录 step 的分位数、最大值、慢 step 原因以及慢 step 是否集中在某些节点、数据桶或并行阶段。若只是用平均值计算 ETA，训练可能在最后阶段反复推迟，容量计划也会失真。

尾部还来自 straggler rank。一个 rank 的坏磁盘、远端 NUMA、GPU 降频或网络重传会让整个 collective 等待。训练运行时可以报告每个 rank 的数据准备、forward、backward、通信和 checkpoint 时间，并设置超过阈值的慢 rank 告警。自动摘除慢节点必须谨慎：替换它会触发新的恢复和通信重建，应该比较继续运行与重启的预计成本。

### 训练取消与优雅停止

取消长训练需要定义边界。用户点击停止后，控制面发出取消事件，worker 在安全 step 边界停止接收新 batch，必要时保存一个可恢复 checkpoint，再关闭数据流和通信组。若节点即将被抢占，平台可以提前发送 preemption notice，让作业尽量完成一次保存。直接 kill 进程虽然快，却可能留下无法判断的新旧状态和大量孤儿资源。

优雅停止的时间也要有上限。checkpoint 长时间卡住时，平台要选择等待、切换存储路径或保留最近健康点；不能因为保存新点而让整个集群无限占用。最终状态应区分 USER_CANCELLED、PREEMPTED、BUDGET_EXCEEDED、FAILED_RECOVERABLE 和 FAILED_PERMANENT。不同状态决定是否自动重试、是否计入实验失败、是否需要人工审批。

### 可复现不等于位级相同

分布式低精度训练中，硬件、归约顺序、kernel、通信算法和随机数都会影响最后几位。平台需要区分严格位级复现、统计复现和语义复现。严格复现可能限制并行和性能；统计复现关注 loss 与主要评估指标在容差内；语义复现关注模型能力和安全指标不发生不可接受变化。job spec 应明确目标等级，而不是笼统写“可复现”。

复现报告包含相同的代码、数据、tokenizer、seed、并行度、dtype、硬件和运行时。如果硬件不同，应说明差异并进行多次重复实验，报告均值和方差。不要把一次结果的微小差异当作故障，也不要用随机性解释明显的 loss 跳变。数据顺序、checkpoint 恢复和编译缓存都应在排查清单中。

### 训练 Infra 的安全边界

训练集群通常拥有高权限网络、对象存储和 GPU，容器中的任意代码都可能读取不该读取的数据。镜像要签名、依赖要锁定、网络要按任务隔离，worker 使用短期凭证，checkpoint 和日志加密。开发者调试权限与生产训练权限分离，禁止通过挂载宿主机路径绕过数据控制。任何访问数据、下载 checkpoint、修改 job spec 和改变发布状态的动作都要有审计。

模型和数据的安全还要覆盖供应链。外部模型、tokenizer、预训练数据、量化工具和 kernel 都可能带来恶意代码或不兼容行为。构建阶段扫描依赖、固定来源和生成 SBOM；运行阶段限制网络和文件系统；发布阶段校验 artifact hash。安全检查不能只在平台上线时做一次，因为模型和依赖会持续更新。

### 训练 Infra 的最终检查表

在进入目标规模训练前，负责人可以逐项确认：job spec 是否不可变；数据 manifest 是否可追溯；tokenizer 是否版本化；有效 token 是否能准确统计；读取、解压、packing 和 batch 是否有背压；并行拓扑是否与硬件匹配；通信是否有真实 trace；dtype 和异常是否受监控；checkpoint 是否两阶段提交；恢复是否验证数据游标和 optimizer；失败是否释放资源；成本是否按有效 token 归因；权限和审计是否覆盖 worker；评估和发布是否有门禁。

这些检查不是行政清单，而是系统边界的外化。缺少其中任何一项，都可能把故障推迟到更昂贵的阶段：数据问题会在训练几天后才暴露，checkpoint 问题会在节点故障时暴露，权限问题会在审计时暴露，吞吐问题会在扩大集群后暴露。把它们前置到小规模验证，通常是训练 Infra 最便宜的优化。

### 一个 step 的完整时间线

为了让性能分析能够落地，可以把每个 step 的时间线固定为：数据游标推进、CPU 预取、Pinned Memory 拷贝、H2D 拷贝、forward、activation 保存或重算、backward、梯度 bucket ready、collective、optimizer update、日志和 checkpoint。每个区间记录开始与结束时间，并以 global step、rank 和 node 关联。这样当吞吐下降时，可以直接回答是数据迟到、GPU kernel 变慢、通信不重叠，还是保存状态引起了暂停。

时间线也能帮助比较不同优化。改变 packing 主要影响有效 token 和数据阶段；改变 FlashAttention 主要影响 attention kernel 与显存；改变 FSDP 主要影响 all-gather、reduce-scatter 和显存峰值；改变 checkpoint 主要影响 I/O 与同步。把所有优化都归结为 GPU utilization，会丢失这些因果关系。性能报告应给出端到端收益以及每个区间的变化，方便后续回归。

### 从训练运行到组织能力

当训练作业数量增加，平台还要提供实验比较、配额管理、预算预警、标准化 runbook 和事故复盘。不同团队使用同一套 spec、manifest、checkpoint 和评估接口，结果才能横向比较；同一模型在不同规模的扩展曲线才能积累；失败原因才能形成知识，而不是每次由个人重新排查。技术平台的价值不仅在于少写启动脚本，更在于把隐性经验沉淀为可执行的默认策略。

对具备后端经验的团队，最容易迁移的能力是接口、状态机、幂等、租约、背压、审计和故障演练；最需要补齐的能力是 GPU 内存账本、集合通信、数值稳定性、kernel profiler 和数据配方。两类能力结合后，训练 Infra 才不会陷入“懂模型的人不懂生产，懂平台的人不懂训练语义”的分割。最终交付的不是一套框架名称，而是一条可以在数据、算法、硬件和业务变化下持续工作的训练生产线。

### 训练结果的可审计性

一次训练结果至少需要能回答四个问题：使用了什么数据，执行了什么计算，经历了哪些故障，为什么可以发布。数据由 manifest、授权和统计回答；计算由 job spec、代码 hash、并行拓扑和运行时镜像回答；故障由事件、trace、重试和恢复记录回答；发布由评估报告、质量门禁和审批记录回答。把这些信息分散在聊天记录、机器本地日志和手工表格中，短期看似灵活，长期一定无法复盘。

可审计不等于保存所有原始内容。审计记录可以使用不可逆 hash、版本号、长度、来源、权限和时间，敏感文本单独受控。关键事件使用追加式日志，修正通过新事件表达而不是覆盖旧记录。这样既能支持事故调查和合规审查，又不会把训练数据复制到每个监控系统。Dapper 的 trace 关联思想和 OpenTelemetry 的上下文传播可以帮助连接不同服务，但数据最小化仍然是平台责任 [20][22]。

### 训练 Infra 的验收结论

当正文中所有指标都能回到具体的 workload、硬件、代码和数据版本时，训练 Infra 才具备工程可信度。功能验收证明它可以训练，性能验收证明它值得训练，恢复验收证明它敢于训练，治理验收证明它可以被组织长期使用。四者缺一不可：只有性能没有恢复，规模越大风险越大；只有恢复没有数据治理，模型越多审计越难；只有治理没有性能，平台无法支撑真实研发节奏。

训练平台还应提供面向使用者的失败解释。不是简单显示 job failed，而是说明最后健康 step、失败阶段、可能原因、是否存在可恢复 checkpoint、估计重算 token、建议动作和相关 trace。解释可以由结构化事件和规则生成，关键结论由负责人确认。失败解释越清楚，工程师越能快速区分代码 bug、数据坏 shard、资源不足、网络故障和预算停止，也越不容易通过危险的手工操作绕过平台。

对于大规模训练，正确性与效率必须同步演进。一个高效但无法恢复的系统会在故障后浪费更多 GPU；一个严格保存所有状态但每步吞吐极低的系统也无法完成目标。最好的工程选择通常是明确语义后做取舍：哪些状态必须精确保存，哪些数据允许重复，哪些指标必须实时，哪些日志可以采样，哪些故障自动重试，哪些故障必须人工接管。把这些取舍写进 job spec、平台默认值和验收矩阵，训练 Infra 才真正成为可运营的系统。

至此，训练 Infra 的主链路已经闭合：数据以版本化 manifest 进入可观测管线，控制面以不可变 spec 申请符合拓扑的资源，运行时通过并行与通信完成计算，checkpoint 用一致协议保存状态，故障通过事件和恢复策略处理，评估与注册表把可训练状态转换成可发布制品。后续推理章节将继续沿用这套方法，把在线请求、KV cache、批处理和服务 SLO 连接起来。

平台验收时还应随机抽取一个已完成 run，尝试从注册表反向重建它：找到数据版本，拉取镜像，申请兼容 GPU，恢复 checkpoint，运行固定 probe，重现主要性能指标，再生成推理制品。若这个过程依赖某位工程师记忆中的参数，就说明平台仍有隐性状态。可重建演练可以按月执行，并把耗时、缺失字段和差异写回平台 backlog。

当重建演练、故障演练和目标规模压测都通过，训练平台才可以把“完成一次训练”当作稳定能力，而不是一次偶然成功。对于后端工程师，这相当于同时拥有部署流水线、数据库备份、分布式追踪和容量测试；只是状态更多、计算更贵、质量影响更直接，因此需要更严格的版本和证据链。

因此，本章的核心验收对象不是某个框架，而是训练状态从数据入口到模型出口的连续性：数据可定位，计算可解释，通信可测量，checkpoint 可恢复，故障可演练，制品可发布，成本可归因。只要这条连续性成立，未来更换并行框架、GPU 类型、存储系统或调度器时，平台仍然拥有稳定的迁移边界。

这条连续性也应在每次平台升级后重新验证。

这也是训练 Infra 与普通离线脚本的根本区别：它不是一次性执行，而是一个拥有状态、资源、事件、恢复和发布生命周期的生产系统。

训练平台的基础实现还应依赖清晰的分布式抽象与学习理论边界。PyTorch Distributed 的 collective 语义决定了通信组和错误处理，中文深度学习教材帮助校验优化与数值推导，机器学习教材提醒我们不要把训练集指标直接等同于泛化能力 [14][23][24][25]。这些来源不是装饰性的参考文献，而是把平台默认值、性能解释和质量门禁连接到可复核理论与工程接口的依据。

因此，每个训练 run 的验收报告都应同时包含系统证据和学习证据：step 时间、有效 token、通信和恢复记录，以及 loss、验证集、分桶指标和异常样本。只有两类证据同时合格，run 才能进入发布候选；单独的吞吐峰值或单独的验证集分数，都不足以证明训练 Infra 工作正确。

这种联合报告也是跨团队评审和长期回归的共同事实来源。

它应成为训练制品的固定附件。

## 参考资料

[1] Shoeybi, M., et al. *Megatron-LM: Training Multi-Billion Parameter Language Models Using Model Parallelism*. 2019. https://arxiv.org/abs/1909.08053

[2] Narayanan, D., et al. *Efficient Large-Scale Language Model Training on GPU Clusters Using Megatron-LM*. 2021. https://arxiv.org/abs/2104.04473

[3] Rajbhandari, S., et al. *ZeRO: Memory Optimizations Toward Training Trillion Parameter Models*. SC, 2020. https://arxiv.org/abs/1910.02054

[4] Rajbhandari, S., et al. *ZeRO-Infinity*. SC, 2021. https://arxiv.org/abs/2104.07857

[5] Rasley, J., et al. *DeepSpeed*. KDD, 2020. https://arxiv.org/abs/2007.04262

[6] PyTorch. *Fully Sharded Data Parallel Documentation*. https://pytorch.org/docs/stable/fsdp.html

[7] Lepikhin, D., et al. *GShard*. 2020. https://arxiv.org/abs/2006.16668

[8] Fedus, W., et al. *Switch Transformers*. 2021. https://arxiv.org/abs/2101.03961

[9] Chowdhery, A., et al. *PaLM*. 2022. https://arxiv.org/abs/2204.02311

[10] Dubey, A., et al. *The Llama 3 Herd of Models*. 2024. https://arxiv.org/abs/2407.21783

[11] Dao, T., et al. *FlashAttention*. NeurIPS, 2022. https://arxiv.org/abs/2205.14135

[12] Dao, T. *FlashAttention-2*. ICLR, 2024. https://arxiv.org/abs/2307.08691

[13] NVIDIA. *NCCL Documentation*. https://docs.nvidia.com/deeplearning/nccl/

[14] PyTorch. *Distributed Communication Package*. https://pytorch.org/docs/stable/distributed.html

[15] Zaharia, M., et al. *Resilient Distributed Datasets*. NSDI, 2012. https://www.usenix.org/legacy/events/nsdi12/tech/full_papers/Zaharia_new.pdf

[16] Moritz, P., et al. *Ray*. OSDI, 2018. https://www.usenix.org/conference/osdi18/presentation/moritz

[17] Verma, A., et al. *Large-scale Cluster Management at Google with Borg*. EuroSys, 2015. https://research.google/pubs/large-scale-cluster-management-at-google-with-borg/

[18] Kubernetes. *Schedule GPUs*. https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/

[19] Beyer, B., et al. *The Site Reliability Workbook*. 2018. https://sre.google/workbook/table-of-contents/

[20] OpenTelemetry Authors. *OpenTelemetry Documentation*. https://opentelemetry.io/docs/

[21] Prometheus Authors. *Prometheus Documentation*. https://prometheus.io/docs/introduction/overview/

[22] Sigelman, B. H., et al. *Dapper*. 2010. https://research.google/pubs/dapper-a-large-scale-distributed-systems-tracing-infrastructure/

[23] Zhang, A., et al. *Dive into Deep Learning*. https://zh.d2l.ai/

[24] 邱锡鹏：《神经网络与深度学习》。https://nndl.github.io/

[25] 周志华：《机器学习》。https://cs.nju.edu.cn/zhouzh/zhouzh.files/publication/MLbook2016.htm
