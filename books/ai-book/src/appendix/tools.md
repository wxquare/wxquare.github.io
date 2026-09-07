# 附录C 常用工具与框架

本附录整理了AI Agent开发中常用的工具、框架和服务，方便快速查找和选择。

---

## 🤖 LLM Provider APIs

### OpenAI

**产品**：GPT-4 Turbo, GPT-4, GPT-3.5 Turbo  
**优势**：生态完善、API稳定、工具调用成熟  
**定价**：$0.01-0.03/1K tokens（输入）  
**链接**：https://platform.openai.com/

**推荐场景**：
- 复杂推理任务
- 工具调用密集型Agent
- 需要高稳定性的生产环境

---

### Anthropic

**产品**：Claude 3.5 Sonnet, Claude 3 Opus  
**优势**：代码生成强、上下文200k、Prompt Caching  
**定价**：$0.003-0.015/1K tokens（输入）  
**链接**：https://www.anthropic.com/

**推荐场景**：
- 代码生成和审查
- 长文档处理
- 需要Prompt Caching降低成本

---

### Google

**产品**：Gemini 1.5 Pro, Gemini 1.5 Flash  
**优势**：上下文1M tokens、多模态、价格低  
**定价**：$0.00125-0.005/1K tokens  
**链接**：https://ai.google.dev/

**推荐场景**：
- 超长文档处理
- 多模态任务（图片+文本）
- 成本敏感的应用

---

### 本地部署

**Ollama**  
https://ollama.ai/  
*本地运行Llama、Mistral等开源模型*

**LM Studio**  
https://lmstudio.ai/  
*图形化界面的本地LLM工具*

**vLLM**  
https://github.com/vllm-project/vllm  
*高性能LLM推理引擎*

---

## 🧰 开发框架

### LangChain

**描述**：最流行的LLM应用开发框架  
**语言**：Python, JavaScript  
**链接**：https://www.langchain.com/

**核心功能**：
- Prompt模板和管理
- Agent和工具调用
- RAG系统
- 记忆和上下文管理
- LangSmith（可观测性平台）

**适用场景**：
- 快速原型开发
- 复杂的多步骤工作流
- 需要丰富的集成生态

**示例代码**：
```python
from langchain.agents import create_openai_functions_agent
from langchain_openai import ChatOpenAI
from langchain.tools import Tool

llm = ChatOpenAI(model="gpt-4")
tools = [Tool(name="search", func=search_function, description="...")]
agent = create_openai_functions_agent(llm, tools, prompt)
```

---

### LlamaIndex

**描述**：专注于RAG和数据索引的框架  
**语言**：Python, TypeScript  
**链接**：https://www.llamaindex.ai/

**核心功能**：
- 数据加载和解析
- 索引构建和管理
- 查询引擎
- 多种检索策略

**适用场景**：
- RAG系统
- 企业知识库
- 文档问答

**示例代码**：
```python
from llama_index import VectorStoreIndex, SimpleDirectoryReader

documents = SimpleDirectoryReader('data').load_data()
index = VectorStoreIndex.from_documents(documents)
query_engine = index.as_query_engine()
response = query_engine.query("What is...?")
```

---

### AutoGPT / AutoGen

**AutoGPT**  
https://github.com/Significant-Gravitas/AutoGPT  
*自主Agent实现*

**AutoGen (Microsoft)**  
https://microsoft.github.io/autogen/  
*多Agent协作框架*

**适用场景**：
- 自主任务执行
- 多Agent协作
- 复杂工作流

---

### Semantic Kernel (Microsoft)

**描述**：微软的LLM编排框架  
**语言**：C#, Python, Java  
**链接**：https://learn.microsoft.com/en-us/semantic-kernel/

**适用场景**：
- 企业级应用
- .NET生态集成
- Azure云服务

---

## 📊 向量数据库

### Chroma

**类型**：本地/嵌入式  
**语言**：Python, JavaScript  
**链接**：https://www.trychroma.com/

**特点**：
- 轻量级、易用
- 适合原型开发
- 支持内存和持久化模式

**使用**：
```python
import chromadb
client = chromadb.PersistentClient(path="./chroma_db")
collection = client.create_collection("docs")
collection.add(documents=[...], embeddings=[...])
```

---

### Pinecone

**类型**：云服务（托管）  
**链接**：https://www.pinecone.io/

**特点**：
- 高性能、可扩展
- 托管服务，无需运维
- 付费服务

**定价**：Starter免费，标准版$70/月起

---

### Weaviate

**类型**：本地/云  
**链接**：https://weaviate.io/

**特点**：
- 功能丰富
- 支持GraphQL
- 混合检索（向量+关键词）

---

### Milvus / Zilliz

**类型**：本地/云  
**链接**：https://milvus.io/

**特点**：
- 高性能、分布式
- 支持大规模数据
- 企业级功能

---

### Qdrant

**类型**：本地/云  
**链接**：https://qdrant.tech/

**特点**：
- Rust实现，高性能
- 丰富的过滤功能
- 支持多向量

---

### pgvector

**类型**：PostgreSQL扩展  
**链接**：https://github.com/pgvector/pgvector

**特点**：
- 与SQL集成
- 适合已有PG数据库的项目
- 开源免费

---

## 🔍 Embedding服务

### OpenAI Embeddings

**模型**：text-embedding-3-small, text-embedding-3-large  
**维度**：1536, 3072  
**定价**：$0.02-0.13/1M tokens  
**链接**：https://platform.openai.com/docs/guides/embeddings

---

### Cohere Embed

**模型**：embed-v3  
**维度**：1024  
**特点**：多语言支持  
**链接**：https://cohere.com/embed

---

### Voyage AI

**特点**：专业领域优化  
**链接**：https://www.voyageai.com/

---

### 开源Embedding模型

**Sentence Transformers**  
https://www.sbert.net/  
*开源Embedding模型库*

**推荐模型**：
- `all-MiniLM-L6-v2`：轻量快速
- `all-mpnet-base-v2`：平衡性能
- `bge-large-zh-v1.5`：中文优化

---

## 🎨 AI编程工具

### Cursor

**描述**：AI优先的代码编辑器  
**基于**：VS Code  
**链接**：https://cursor.com/

**核心功能**：
- Cmd+K 行内编辑
- Composer 多文件编辑
- Chat 对话式编程
- Claude Code Terminal Agent

**定价**：免费试用，Pro $20/月

---

### GitHub Copilot

**描述**：GitHub官方AI编程助手  
**支持**：VS Code, JetBrains等  
**链接**：https://github.com/features/copilot

**定价**：$10/月（个人），$19/月（Pro）

---

### Cline (VS Code Extension)

**描述**：开源的自主编程Agent  
**链接**：https://github.com/cline/cline

**特点**：
- 开源免费
- 支持多种LLM Provider
- 自主编辑文件和执行命令

---

## 📝 知识管理工具

### Obsidian

**描述**：本地优先的笔记软件  
**特点**：双向链接、Graph View、插件生态  
**链接**：https://obsidian.md/

**推荐插件**：
- Dataview：数据查询
- Templater：模板系统
- Excalidraw：画图
- Marp：幻灯片

---

### Notion

**描述**：在线协作知识库  
**特点**：数据库、团队协作、AI助手  
**链接**：https://www.notion.so/

---

### Logseq

**描述**：开源的双向链接笔记  
**特点**：本地优先、大纲式、开源  
**链接**：https://logseq.com/

---

## 🔧 开发工具

### LangSmith

**描述**：LangChain的可观测性平台  
**功能**：追踪、调试、评估  
**链接**：https://smith.langchain.com/

---

### Helicone

**描述**：LLM可观测性平台  
**功能**：日志、缓存、成本追踪  
**链接**：https://www.helicone.ai/

---

### Weights & Biases

**描述**：机器学习实验平台  
**功能**：实验追踪、超参数优化  
**链接**：https://wandb.ai/

---

## 🧪 测试与评估

### Braintrust

**描述**：LLM应用评估平台  
**功能**：数据集管理、自动评估  
**链接**：https://www.braintrustdata.com/

---

### PromptFoo

**描述**：开源Prompt测试工具  
**功能**：批量测试、自动评估  
**链接**：https://promptfoo.dev/

---

## 📦 模型部署

### Replicate

**描述**：模型托管和API服务  
**特点**：按使用付费、丰富的模型库  
**链接**：https://replicate.com/

---

### Modal

**描述**：无服务器Python运行时  
**特点**：GPU支持、自动扩展  
**链接**：https://modal.com/

---

### Together AI

**描述**：开源模型推理API  
**特点**：多种开源模型、价格低  
**链接**：https://www.together.ai/

---

## 🎯 工具选择建议

### 快速原型（个人项目）

- **LLM**：OpenAI GPT-4 或 Claude 3.5
- **框架**：LangChain
- **向量数据库**：Chroma
- **Embedding**：OpenAI text-embedding-3-small
- **编程工具**：Cursor

### 生产环境（小型团队）

- **LLM**：OpenAI + Claude（双provider）
- **框架**：LangChain + 自定义代码
- **向量数据库**：Pinecone 或 Weaviate
- **Embedding**：OpenAI 或 Voyage AI
- **可观测性**：LangSmith + Helicone

### 企业级（大型公司）

- **LLM**：私有部署（Llama）+ 云服务（OpenAI/Azure）
- **框架**：自研框架 + Semantic Kernel
- **向量数据库**：Milvus（自部署）
- **Embedding**：自训练模型 + 商业API
- **可观测性**：自建监控系统

---

## 💡 成本优化建议

### 开发阶段

- 使用 GPT-3.5 Turbo 或 Gemini Flash
- 本地向量数据库（Chroma）
- 开源Embedding模型

### 生产阶段

- 模型分层（简单任务用便宜模型）
- Prompt Caching（Claude）
- Batch API（OpenAI）
- 向量数据库按需选择

---

*本工具清单持续更新，欢迎补充*
