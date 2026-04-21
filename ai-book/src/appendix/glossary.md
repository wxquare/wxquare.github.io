# 附录A 术语表

本术语表收录了书中出现的核心术语和概念，按字母顺序排列。

---

## A

**Agent（智能体）**  
能够感知环境、做出决策并执行操作的自主系统。在AI领域，通常指基于LLM的自主任务执行系统。

**Attention Mechanism（注意力机制）**  
Transformer架构的核心，通过计算序列内部元素之间的关联权重，捕获长距离依赖关系。

---

## B

**Backlinks（反向链接）**  
在知识管理系统中，指向当前页面的所有链接。用于建立概念之间的双向关联。

**BM25**  
一种基于词频和逆文档频率的关键词检索算法，常用于混合检索策略。

---

## C

**Chain-of-Thought（思维链）**  
一种Prompt Engineering技术，要求LLM逐步展示推理过程，提高复杂推理任务的准确性。

**Chunking（分块）**  
将长文档切分为较小片段的过程，用于RAG系统中的文档索引和检索。

**Context Window（上下文窗口）**  
LLM一次可以处理的最大token数量。现代LLM的上下文窗口从16k到1M tokens不等。

**Cosine Similarity（余弦相似度）**  
衡量两个向量方向相似程度的指标，常用于计算文本Embedding之间的语义相似度。

---

## D

**Decision Tree（决策树）**  
在Agent系统中，用于判断是否应使用Agent的决策框架，通过一系列Yes/No问题评估任务特征。

**DoD（Duty on Duty）**  
值班工程师，负责处理生产环境的告警和突发问题。

---

## E

**Embedding（嵌入向量）**  
将文本转换为固定维度的数值向量的技术，捕获文本的语义信息。

**Early Stopping（提前停止）**  
检测Agent陷入无效循环时及时终止执行的机制。

---

## F

**Few-Shot Learning（少样本学习）**  
在Prompt中提供少量示例，帮助LLM理解任务格式和期望输出。

**Fingerprint（指纹）**  
告警的唯一标识符，用于去重和关联相同的告警。

---

## G

**GAS（Go Application Server）**  
Shopee内部的Go语言应用开发框架。

---

## H

**Hallucination（幻觉）**  
LLM生成看似合理但实际错误的内容，包括事实性幻觉、逻辑性幻觉和引用性幻觉。

**Harness Engineering（驾驭工程）**  
通过构建可靠的基础设施和约束系统来驾驭AI的工程方法论。

**Hybrid Search（混合检索）**  
结合向量检索和关键词检索的方法，提高检索的召回率和精确度。

---

## I

**Inference（推理）**  
模型根据输入数据生成输出的过程。在机器学习中指模型的预测阶段。

---

## L

**LLM（Large Language Model，大语言模型）**  
基于Transformer架构的大规模预训练语言模型，如GPT-4、Claude等。

**Loop Detection（循环检测）**  
检测Agent重复执行相同操作的机制，用于防止无限循环。

---

## M

**MCP（Model Context Protocol）**  
Anthropic提出的协议，用于AI工具连接外部数据源和服务。

**MTTR（Mean Time To Resolution，平均恢复时间）**  
从问题发生到解决的平均时间，运维领域的关键指标。

**Multi-Agent System（多Agent系统）**  
由多个Agent协作完成复杂任务的系统架构。

---

## P

**Plan-and-Execute（计划-执行）**  
一种Agent架构模式，先制定详细计划，再逐步执行。

**Positional Encoding（位置编码）**  
在Transformer中注入序列位置信息的机制。

**Prompt Caching（Prompt缓存）**  
缓存长System Prompt以节省token成本的技术。

**Prompt Engineering（提示工程）**  
设计和优化Prompt以提高LLM输出质量的技术。

---

## R

**RAG（Retrieval-Augmented Generation，检索增强生成）**  
结合信息检索和语言模型生成的方法，通过检索外部知识增强LLM能力。

**ReACT（Reasoning and Acting）**  
一种Agent架构模式，交替进行推理（Thought）和行动（Action）。

**Reciprocal Rank Fusion（倒数排名融合）**  
一种融合多个检索结果的算法，用于混合检索。

**Reranking（重排序）**  
使用专门模型重新排序检索结果，提高精确度。

---

## S

**Self-Consistency（自我一致性）**  
多次采样LLM输出，选择最一致的结果以提高准确性。

**Sliding Window（滑动窗口）**  
保留最近N条消息的上下文管理策略。

**SOP（Standard Operating Procedure，标准操作流程）**  
标准化的问题处理流程。

**Spec Coding（规格编程）**  
先定义清晰规格，再让AI生成代码的编程范式。

**State Machine（状态机）**  
管理系统状态转换的设计模式，用于控制Agent工作流。

---

## T

**Token**  
LLM处理文本的基本单位，通常一个token约等于0.75个英文单词或0.5个中文字符。

**Tool Call（工具调用）**  
Agent调用外部工具（API、数据库、命令行等）执行操作。

**Transformer**  
基于Self-Attention机制的神经网络架构，现代LLM的基础。

---

## V

**Vector Database（向量数据库）**  
专门存储和检索高维向量的数据库，用于RAG系统。

**Vibe Coding（感觉编程）**  
依赖直觉和经验，与AI反复迭代的编程方式。

---

## Z

**Zero-Shot Learning（零样本学习）**  
不提供示例，直接让LLM完成任务。

---

## 中文术语

**分级决策**  
根据风险等级采用不同决策策略的方法。

**幻觉**  
见 Hallucination。

**混合架构**  
结合多种设计模式的系统架构，如状态机 + ReACT。

**可观测性**  
通过指标、日志、追踪等方式了解系统运行状态的能力。

**上下文管理**  
管理LLM上下文窗口中信息的策略和技术。

**状态机**  
见 State Machine。

**向量数据库**  
见 Vector Database。

**知识编译**  
LLM将原始数据转换为结构化知识的过程。

---

*本术语表持续更新中*
