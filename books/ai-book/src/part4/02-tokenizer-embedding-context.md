# 第22章 Token、Tokenizer、Embedding 与上下文窗口

LLM 不能直接处理“文字”。它处理的是 token 序列。理解 token，是理解上下文长度、成本、延迟、RAG 切块、Prompt 预算和长上下文限制的第一步。

## 22.1 宏观理解：文本如何进入模型

从用户输入到模型输入，大致经过下面几步：

```mermaid
flowchart LR
    A["原始文本"] --> B["Tokenizer"]
    B --> C["Token 序列"]
    C --> D["Token IDs"]
    D --> E["Embedding 向量"]
    E --> F["Transformer 层"]
```

Tokenizer 负责把文本切成模型词表里的 token。Embedding 层负责把每个 token ID 映射成向量。

注意：token 不等于汉字，也不等于英文单词。

例如：

- 一个英文单词可能被切成一个或多个 token。
- 一个中文词可能按字、词或子词切分。
- 空格、换行、标点、缩进、代码符号都可能影响 tokenization。
- 同一句话在不同模型的 tokenizer 下 token 数可能不同。

这就是为什么同样 10 页文档，换一个模型后 token 数、价格和延迟都可能变化。

## 22.2 Tokenizer 的几类常见方法

### 22.2.1 BPE

BPE（Byte Pair Encoding）从字符或字节开始，反复合并高频片段，得到词表。它常见于 GPT 系模型。

优点是简单、压缩率高、能处理未登录词；缺点是切分结果不一定符合自然语言词边界。

### 22.2.2 WordPiece

WordPiece 常见于 BERT 系模型。它同样使用子词思想，但合并策略与 BPE 不完全相同。

工程上你不需要记住所有训练细节，关键是理解：WordPiece 和 BPE 都是在词和字符之间找一个折中，让模型既能表达常见词，又能处理罕见词。

### 22.2.3 SentencePiece

SentencePiece 把输入当成 Unicode 字符序列，不依赖预先分词，因此对多语言更友好。很多开源模型使用 SentencePiece 或类似方案。

### 22.2.4 Byte-level Tokenizer

Byte-level tokenizer 从字节层面处理文本，理论上可以覆盖任意输入，不容易遇到 unknown token。但它可能让某些语言或特殊文本变成长 token 序列。

## 22.3 Embedding：从离散符号到连续向量

Tokenizer 输出的是 token ID，例如：

```text
["KV", " cache", " 是", " 什么"] -> [12345, 6789, 3456, 7890]
```

模型不能直接在 ID 上做语义计算。Embedding 层会把每个 ID 映射成一个向量：

```text
token_id -> dense vector
```

这些向量不是人工写的词典，而是在训练中学习出来的参数。相似语境中的 token 往往会形成相似的表示，但不要把 embedding 简化成“词义坐标”。在深层 Transformer 中，hidden state 会随着上下文不断变化，同一个 token 在不同句子里的表示也会不同。

## 22.4 上下文窗口是什么

上下文窗口指模型一次调用中最多能处理的 token 数。

它包括：

- system prompt；
- developer / instruction 信息；
- 用户输入；
- few-shot 示例；
- 工具说明；
- 检索片段；
- 历史对话；
- 模型已经生成的输出。

上下文窗口不是数据库。它只是本次推理时模型能看到的 token 序列。超过窗口的内容要么被截断，要么需要压缩、检索、分层加载或重新组织。

## 22.5 上下文窗口的三个成本

长上下文不只是“能放更多字”，它有三个成本。

### 1. 价格成本

大多数模型按输入 token 和输出 token 计费。长文档、长历史、长工具结果都会增加输入成本。

### 2. 延迟成本

输入越长，prefill 阶段越慢。用户感受到的首 token 延迟会增加。

### 3. 显存成本

推理时需要保存历史 token 的 KV cache。上下文越长，KV cache 越大，并发能力越受限。第24章会专门解释这个问题。

## 22.6 工业实践：Token 预算怎么做

生产系统不会无脑把所有信息塞进 prompt。常见做法是给不同上下文分配预算：

```text
总预算 = 系统指令 + 用户输入 + 会话状态 + 检索证据 + 工具结果 + 输出预留
```

例如一个 32K token 窗口的企业知识助手，可以这样分配：

- 2K：系统指令、角色、风格和安全边界；
- 4K：用户问题、会话摘要和任务状态；
- 18K：检索证据；
- 4K：工具结果；
- 4K：输出预留。

实际系统还要动态调整：简单问题少取证据，复杂问题多取证据；短回答少预留输出，报告生成多预留输出。

## 22.7 工业实践：常见坑

### 22.7.1 用字符数估算 token

字符数只能粗略估计，不能用于严格预算。中文、英文、代码、Markdown 表格、JSON、日志的 token 密度差异很大。

### 22.7.2 忽略 Chat Template

很多开源模型有自己的 chat template。system/user/assistant/tool 消息会被包装成特殊格式。你看到的文本长度，不等于真实 token 长度。

### 22.7.3 RAG chunk 太大

chunk 太大时，召回结果看似包含答案，但真正相关的信息密度低。模型要在大量噪声中找答案，效果可能更差。

### 22.7.4 工具结果无限追加

Agent 每轮调用工具后都把完整结果塞回上下文，会让上下文迅速膨胀。应该保留结构化状态和关键证据，而不是保留所有原始日志。

## 22.8 科研现状：Tokenizer 的问题与新方向

截至 2026-05，tokenizer 仍然是大模型系统里的基础组件，但它有明显问题：

- 多语言 token 效率不均衡；
- 代码、数学、表格和特殊符号切分不稳定；
- token 边界不一定符合语义边界；
- 不同模型 tokenizer 不兼容；
- 长上下文成本受 tokenization 强烈影响。

研究方向包括：

- **Tokenizer-free / byte-level 模型**：直接在字节或更细粒度上建模，减少手工 tokenizer 的偏置。
- **Byte Latent Transformer（BLT）**：尝试用动态 byte patch 替代固定 token 切分，在鲁棒性和推理效率之间寻找新平衡。
- **多语言 tokenizer 优化**：减少非英语语言的 token 膨胀。
- **面向代码和结构化文本的 tokenizer**：让缩进、语法符号和结构边界更稳定。

短期工业界仍会大量使用成熟 tokenizer，因为它们和已有模型、推理引擎、训练数据、评测体系高度绑定。长期看，tokenizer-free 或动态切分会继续挑战传统 tokenization。

## 22.9 工程清单

设计 LLM 系统时，至少要问：

- 当前模型的 tokenizer 是什么？
- 同一批文本在目标模型下平均 token 密度是多少？
- 是否统计过 system prompt、工具 schema 和 chat template 的真实 token 数？
- 输出是否预留了足够 token？
- 超预算时是截断、摘要、检索还是分层加载？
- RAG chunk 大小是否按 token 而不是字符设计？
- 是否监控真实线上请求的 token 分布？
- 是否区分输入 token、输出 token 和 cache token 成本？

## 22.10 面试表达

一句话版：

> LLM 处理的不是字符或单词，而是 tokenizer 切出来的 token。Token 会被映射成 embedding 向量进入 Transformer。上下文窗口限制的是一次推理能看到的 token 序列，它同时影响价格、延迟和 KV cache 显存。

展开版：

> 我会把 token 预算当成 LLM 系统设计的一等资源。因为 system prompt、工具 schema、历史对话、检索证据和输出都共享同一个上下文窗口。长上下文虽然提升了可见信息量，但会增加 prefill 延迟和 KV cache 显存。因此在生产系统里，我会统计真实 token 分布，给不同上下文类型分配预算，并在超预算时做摘要、检索、分层加载或状态化，而不是简单截断。

## 22.11 专家视角：Tokenizer 是模型能力的一部分

Tokenizer 经常被当成预处理工具，但它实际上是模型能力边界的一部分。模型训练时看到的是 token 序列，而不是原始字符序列。tokenizer 的切分方式会影响模型学到的统计模式。

例如，中文如果被切得更碎，同样一段话会占用更多 token，模型要用更多位置才能表达同样语义。这不仅增加成本，也改变 attention 距离。代码也是如此：缩进、括号、换行、变量名和特殊符号如果切分不稳定，模型学习代码结构会更困难。

Tokenizer 还会影响安全和鲁棒性。有些 prompt injection、越狱字符串、Unicode 混淆和不可见字符攻击，本质上利用了人类可见文本和模型 token 序列之间的不一致。人看到的是一句正常话，模型看到的可能是异常 token 组合。

因此在生产系统中，tokenizer 至少要进入三个地方：

- 成本估算；
- 上下文预算；
- 安全和输入规范化。

## 22.12 深入理解：Embedding 不是静态语义，也不是向量数据库专属概念

Embedding 这个词容易混淆，因为它在 LLM 内部和 RAG 系统中都出现。

在 LLM 内部，embedding 层把 token ID 映射成初始向量。这些向量随后经过多层 Transformer，不断被上下文改写。第 1 层的 “bank” 可能只是一个 token 表示；到第 20 层时，它可能已经结合上下文变成“河岸”或“银行”的语义状态。

在 RAG 中，embedding model 把一个 query 或文档 chunk 映射成一个向量，用于相似度检索。这个向量通常是整段文本的压缩表示，服务于“找相似内容”。

二者有联系，但不能混为一谈：

```text
LLM token embedding：模型内部表示的起点
RAG text embedding：外部检索系统的语义索引
```

工程上最常见的错误是把 RAG embedding 想象成“精确理解文本”。实际上，一个单向量 embedding 会压缩掉很多细节，尤其是数字、否定、时间、版本、代码符号、表格结构和权限语义。因此生产 RAG 常常需要 hybrid retrieval、metadata filter 和 rerank。

## 22.13 深入理解：上下文窗口是软能力，不是硬承诺

模型标称支持 128K 或 1M token，并不意味着它能同等质量地使用窗口中的所有信息。

长上下文能力至少要拆成四件事：

- **可输入**：模型和 serving engine 允许这么长的 token 序列。
- **可保持**：模型不会在长序列下明显退化或丢失中间信息。
- **可定位**：模型能从长文本中找到关键证据。
- **可整合**：模型能跨多个位置组合信息并生成正确答案。

很多长上下文失败不是窗口超限，而是定位和整合失败。常见现象包括：

- 只引用开头和结尾，忽略中间内容；
- 多个证据冲突时选择更近的证据；
- 对长表格或日志做错误聚合；
- 在多文档中混淆来源；
- 回答看似完整但漏掉关键约束。

这也是为什么长上下文系统仍需要 RAG、目录结构、摘要、引用和 eval。窗口变大只是给系统更多空间，不等于自动解决信息组织问题。

## 22.14 工业实践：Token 预算应该变成可观测指标

成熟团队不会只在开发阶段估算 token，而会在线上持续监控：

- input tokens 分布；
- output tokens 分布；
- system prompt 占比；
- tool schema 占比；
- RAG evidence 占比；
- discarded context 占比；
- 超预算请求比例；
- prompt cache 命中率；
- 长上下文请求的 TTFT 和错误率。

这些指标可以直接指导架构优化。

如果 system prompt 长期占用 30% 以上上下文，说明工具说明或规则需要结构化压缩。如果 RAG evidence 占比很高但答案忠实性不提升，说明检索质量或 context builder 有问题。如果超长请求带来明显排队延迟，就需要 admission control、异步任务或分层摘要。

上下文管理的成熟标志不是“窗口够大”，而是你知道每一类 token 在系统里花在哪里、换来了什么。

## 22.15 研究补充：Byte-level 与 Tokenizer-free 路线

传统 tokenizer 的问题越来越明显：多语言不公平、长尾字符处理差、代码和表格结构不稳定、token 边界和语义边界不一致。因此 byte-level 和 tokenizer-free 路线持续升温。

Byte Latent Transformer（BLT）尝试直接在字节上建模，但不是天真地一个字节一个字节跑完整 Transformer，而是把字节聚合成动态 patch，让模型在信息复杂的地方用更多计算，在可预测的地方用更少计算。

2026 年 Fast Byte Latent Transformer 进一步关注 byte-level 模型的推理瓶颈：如果逐字节自回归生成，forward 次数会很多。它引入 block-wise diffusion 和 self-speculation 等思路，目标是在保留 tokenizer-free 鲁棒性的同时改善生成速度。

这条路线还没有完全替代主流 token 模型，但它指出了一个重要方向：未来大模型的“输入单位”不一定永远是固定词表 token，可能会变成动态、数据驱动、甚至按模态自适应的表示单元。

## 22.16 常见误区：Token 与上下文

### 误区 1：按字符数控制上下文

字符数和 token 数没有稳定比例。英文、中文、代码、JSON、Markdown 表格、日志、emoji、不可见字符都会改变 token 密度。生产系统必须用目标模型 tokenizer 计算真实 token 数。

### 误区 2：认为 embedding 能保留所有语义

Embedding 是压缩表示。压缩就会损失细节。版本号、否定、数字、时间、条件、权限、代码符号都可能在向量相似度中被弱化。

### 误区 3：上下文越长越好

更长上下文会增加成本、延迟和 KV cache 显存，还可能引入噪声。好的上下文不是最长，而是信息密度最高、边界最清楚、证据最可靠。

### 误区 4：摘要一定安全

摘要会丢信息，也可能引入解释。用于 Agent 状态时，摘要应该保留决策、约束、开放问题和证据 ID，而不是把原文压成一段自然语言。

## 22.17 诊断题：如何判断是 Token 问题

当系统出现下面现象时，要怀疑 token 和上下文层：

- 本地测试正常，线上长对话后开始答非所问。
- RAG 检索结果正确，但模型没有引用关键证据。
- 工具 schema 很长，压缩后质量反而提升。
- 中文文档成本明显高于预期。
- 代码任务中模型遗漏文件路径、函数名或错误码。
- prompt cache 命中率低，虽然看起来 system prompt 没变。

诊断步骤：

1. 打印真实 chat template 后的 token 序列长度。
2. 分解各类上下文占比。
3. 检查被截断的是哪一部分。
4. 对比短上下文和长上下文下的输出差异。
5. 检查检索证据是否被噪声淹没。
6. 评估是否需要分层加载、摘要或外部工具。

## 22.18 工程案例：设计一个 32K Token 的上下文预算

假设要设计一个企业知识问答 Agent，模型上下文窗口是 32K。一个合理预算不是平均分配，而是按任务价值分配。

```text
System / Policy：2K
Tool Schema：3K
User Query + Conversation State：3K
Retrieved Evidence：16K
Tool Results：4K
Output Reserve：4K
```

这只是初始值。真实系统要动态调节：

- 如果用户问题简单，检索证据可以降到 4K，把预算留给输出。
- 如果用户要求生成报告，输出预留要增加。
- 如果工具 schema 很稳定，可以依赖 prompt caching 或压缩描述。
- 如果 RAG 证据冲突，要保留更多 metadata 和来源说明。
- 如果会话很长，不要保留完整历史，而是保留结构化状态。

预算还要和质量指标绑定。比如 evidence token 从 8K 增加到 16K，如果准确率没有提升，只是延迟和成本增加，就说明检索或上下文构建没有做好。

## 22.19 工程案例：为什么 Chat Template 会改变结果

同样一段用户文本，直接拼进 prompt 和按模型 chat template 编码，可能得到不同 token 序列。

开源模型通常训练时见到的是特定格式：

```text
<system>...</system>
<user>...</user>
<assistant>...
```

或者某种特殊 token 包围的对话模板。如果部署时模板错了，模型可能：

- 把 system 指令当成普通用户文本；
- 不知道哪里开始回答；
- 工具调用格式不稳定；
- 多轮对话角色混乱；
- stop token 不生效。

所以迁移模型时，除了权重和 tokenizer，还必须检查 chat template、special tokens、stop words 和 tool call format。

## 22.20 参考资料

- [Neural Machine Translation of Rare Words with Subword Units](https://arxiv.org/abs/1508.07909)
- [SentencePiece: A simple and language independent subword tokenizer](https://arxiv.org/abs/1808.06226)
- [Hugging Face Tokenizers](https://huggingface.co/docs/tokenizers/index)
- [OpenAI tiktoken](https://github.com/openai/tiktoken)
- [Byte Latent Transformer: Patches Scale Better Than Tokens](https://arxiv.org/abs/2412.09871)
- [Fast Byte Latent Transformer](https://arxiv.org/abs/2605.08044)
