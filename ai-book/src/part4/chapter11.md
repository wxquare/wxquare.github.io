# 第11章 RAG与上下文工程基础

> "The most powerful models are not those with the most parameters, but those with access to the right information at the right time." （最强大的模型不是参数最多的，而是能在正确的时间获取正确信息的）

## 引言

第10章我们学习了 LLM 的能力边界，其中一个核心限制是：**LLM 的知识截止于训练时间，无法获取实时信息或私有数据**。

RAG（Retrieval-Augmented Generation，检索增强生成）是解决这个问题的关键技术。它通过**外部知识检索**增强 LLM 的能力，让 Agent 系统能够访问最新信息、私有数据和专业知识。

本章将深入探讨 RAG 的核心原理、工程实现和最佳实践。

---

## 11.1 RAG核心原理

### 什么是 RAG？

**定义**：检索增强生成是一种结合信息检索和语言模型的方法，在生成回答前先检索相关知识。

**基本流程：**

```
问题 → 检索相关文档 → 构建上下文 → LLM 生成答案
```

**与传统方法对比：**

```
传统 LLM:
问题 → LLM → 答案
问题：受限于训练数据，可能产生幻觉

RAG:
问题 → 检索 → 相关文档 → LLM + 文档 → 答案
优势：基于事实、可更新、可追溯
```

### 为什么需要 RAG？

**场景 1：实时信息**

```
问题: "2026年4月的 AI 领域有哪些重要新闻？"

纯 LLM:
"我的知识截止于 2023年10月，无法回答"

RAG:
检索 → 新闻数据库 → 最新文章
回答 → "2026年4月，OpenAI 发布了 GPT-5..."
```

**场景 2：私有数据**

```
问题: "我们公司的 Q1 销售数据是多少？"

纯 LLM:
"我无法访问你们公司的私有数据"

RAG:
检索 → 内部数据库 → Q1 报告
回答 → "根据 Q1 报告，销售额为 $5M..."
```

**场景 3：专业领域**

```
问题: "这个 bug 的历史处理记录是什么？"

纯 LLM:
"我不清楚你们系统的具体情况"

RAG:
检索 → Jira + Confluence → 相关 ticket
回答 → "根据 JIRA-1234，这个问题曾在..."
```

### RAG 架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          RAG System                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                   1. 离线索引构建（Indexing）                      │ │
│  │                                                                    │ │
│  │  文档 → 分块 → Embedding → 向量数据库                              │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                   2. 在线检索（Retrieval）                         │ │
│  │                                                                    │ │
│  │  问题 → Embedding → 向量搜索 → Top-K 文档                          │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                   3. 增强生成（Generation）                        │ │
│  │                                                                    │ │
│  │  问题 + 检索文档 → LLM → 答案                                      │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 11.2 核心组件详解

### 组件 1：文档分块（Chunking）

**为什么需要分块？**

- 文档太长无法全部输入 LLM（token 限制）
- 小块更容易匹配相关内容
- 提高检索精度

**分块策略：**

**策略 1：固定大小分块**

```python
def chunk_by_size(text: str, chunk_size: int = 500, overlap: int = 50):
    """按固定字符数分块"""
    chunks = []
    start = 0
    
    while start < len(text):
        end = start + chunk_size
        chunk = text[start:end]
        chunks.append(chunk)
        start = end - overlap  # 重叠部分
    
    return chunks

# 示例
text = "长文本..." * 1000
chunks = chunk_by_size(text, chunk_size=500, overlap=50)
# 结果: 每块 500 字符，相邻块重叠 50 字符
```

**优点**：简单、可控  
**缺点**：可能截断句子或段落

**策略 2：按语义分块**

```python
def chunk_by_semantics(text: str, max_chunk_size: int = 500):
    """按段落和句子分块"""
    paragraphs = text.split('\n\n')
    chunks = []
    current_chunk = ""
    
    for para in paragraphs:
        sentences = para.split('。')
        
        for sent in sentences:
            if len(current_chunk) + len(sent) > max_chunk_size:
                if current_chunk:
                    chunks.append(current_chunk.strip())
                current_chunk = sent
            else:
                current_chunk += sent + '。'
    
    if current_chunk:
        chunks.append(current_chunk.strip())
    
    return chunks
```

**优点**：保持语义完整性  
**缺点**：块大小不均匀

**策略 3：递归分块（LangChain 方式）**

```python
class RecursiveCharacterTextSplitter:
    """递归按分隔符分块"""
    
    def __init__(self, chunk_size: int = 500, chunk_overlap: int = 50):
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap
        self.separators = ["\n\n", "\n", "。", "，", " ", ""]
    
    def split(self, text: str):
        return self._split(text, self.separators)
    
    def _split(self, text: str, separators: list):
        """递归分割"""
        if len(text) <= self.chunk_size:
            return [text]
        
        # 尝试用第一个分隔符分割
        separator = separators[0]
        splits = text.split(separator)
        
        chunks = []
        current_chunk = ""
        
        for split in splits:
            if len(current_chunk) + len(split) <= self.chunk_size:
                current_chunk += split + separator
            else:
                if current_chunk:
                    chunks.append(current_chunk)
                
                # 如果单个 split 太长，用下一级分隔符
                if len(split) > self.chunk_size and len(separators) > 1:
                    chunks.extend(self._split(split, separators[1:]))
                else:
                    current_chunk = split + separator
        
        if current_chunk:
            chunks.append(current_chunk)
        
        return chunks
```

**最佳实践：**

| 文档类型 | 推荐分块大小 | 推荐策略 |
|---------|------------|---------|
| **技术文档** | 300-500 tokens | 按段落 + 递归 |
| **代码** | 100-200 tokens | 按函数/类 |
| **对话记录** | 200-300 tokens | 按轮次 |
| **新闻文章** | 400-600 tokens | 按段落 |

---

### 组件 2：Embedding 模型

**什么是 Embedding？**

将文本转换为固定维度的向量，语义相似的文本在向量空间中距离更近。

```
文本: "Transformer 是一种神经网络架构"
↓
Embedding: [0.23, -0.45, 0.67, ..., 0.12]  # 1536维向量
```

**主流 Embedding 模型对比（2026年）**

| 模型 | 维度 | 性能 | 成本 | 适用场景 |
|------|------|------|------|---------|
| **OpenAI text-embedding-3-small** | 1536 | 优秀 | $0.02/1M tokens | 通用、成本敏感 |
| **OpenAI text-embedding-3-large** | 3072 | 最佳 | $0.13/1M tokens | 高精度要求 |
| **Cohere Embed v3** | 1024 | 优秀 | $0.10/1M tokens | 多语言 |
| **Voyage AI** | 1024 | 优秀 | $0.12/1M tokens | 专业领域 |
| **BGE-M3** | 1024 | 良好 | 免费（自部署） | 中文、本地部署 |
| **E5-large-v2** | 1024 | 良好 | 免费（自部署） | 开源、本地部署 |

**Embedding 实现：**

```python
from openai import OpenAI

class EmbeddingService:
    """Embedding 服务"""
    
    def __init__(self, model: str = "text-embedding-3-small"):
        self.client = OpenAI()
        self.model = model
    
    def encode(self, texts: List[str]) -> List[List[float]]:
        """批量编码文本"""
        
        # OpenAI API 支持批量编码
        response = self.client.embeddings.create(
            model=self.model,
            input=texts
        )
        
        embeddings = [item.embedding for item in response.data]
        return embeddings
    
    def encode_single(self, text: str) -> List[float]:
        """编码单个文本"""
        return self.encode([text])[0]

# 使用示例
embedder = EmbeddingService()
query_embedding = embedder.encode_single("Transformer 架构原理")
doc_embeddings = embedder.encode([
    "Transformer 是一种基于注意力机制的神经网络",
    "卷积神经网络用于图像处理",
    "循环神经网络处理序列数据"
])
```

**相似度计算：**

```python
import numpy as np

def cosine_similarity(vec1: List[float], vec2: List[float]) -> float:
    """余弦相似度"""
    vec1 = np.array(vec1)
    vec2 = np.array(vec2)
    
    dot_product = np.dot(vec1, vec2)
    norm1 = np.linalg.norm(vec1)
    norm2 = np.linalg.norm(vec2)
    
    return dot_product / (norm1 * norm2)

# 示例
query = embedder.encode_single("Transformer 架构")
doc1 = embedder.encode_single("Attention is All You Need 论文")
doc2 = embedder.encode_single("CNN 图像分类")

print(f"Query vs Doc1: {cosine_similarity(query, doc1):.3f}")  # 0.85
print(f"Query vs Doc2: {cosine_similarity(query, doc2):.3f}")  # 0.42
```

---

### 组件 3：向量数据库

**主流向量数据库对比**

| 数据库 | 类型 | 特点 | 适用场景 |
|--------|------|------|---------|
| **Chroma** | 本地/嵌入式 | 轻量、易用 | 原型开发、小规模 |
| **Pinecone** | 云服务 | 高性能、托管 | 生产环境、大规模 |
| **Weaviate** | 本地/云 | 功能丰富、GraphQL | 复杂查询 |
| **Milvus** | 本地/云 | 高性能、分布式 | 大规模生产 |
| **Qdrant** | 本地/云 | Rust 实现、高效 | 性能敏感 |
| **pgvector** | PostgreSQL 扩展 | 与 SQL 集成 | 已有 PG 数据库 |

**Chroma 实现示例：**

```python
import chromadb
from chromadb.config import Settings

class ChromaVectorStore:
    """Chroma 向量数据库封装"""
    
    def __init__(self, persist_dir: str = "./chroma_db"):
        self.client = chromadb.PersistentClient(
            path=persist_dir,
            settings=Settings(anonymized_telemetry=False)
        )
        self.collection = None
    
    def create_collection(self, name: str):
        """创建集合"""
        self.collection = self.client.get_or_create_collection(
            name=name,
            metadata={"hnsw:space": "cosine"}  # 使用余弦相似度
        )
    
    def add_documents(
        self, 
        ids: List[str],
        documents: List[str],
        embeddings: List[List[float]],
        metadatas: List[dict] = None
    ):
        """添加文档"""
        self.collection.add(
            ids=ids,
            documents=documents,
            embeddings=embeddings,
            metadatas=metadatas
        )
    
    def search(
        self,
        query_embedding: List[float],
        top_k: int = 5,
        filters: dict = None
    ) -> dict:
        """向量搜索"""
        results = self.collection.query(
            query_embeddings=[query_embedding],
            n_results=top_k,
            where=filters  # 元数据过滤
        )
        
        return {
            "ids": results["ids"][0],
            "documents": results["documents"][0],
            "distances": results["distances"][0],
            "metadatas": results["metadatas"][0]
        }

# 使用示例
vector_store = ChromaVectorStore()
vector_store.create_collection("knowledge_base")

# 添加文档
embedder = EmbeddingService()
docs = [
    "Transformer 使用 Self-Attention 机制",
    "BERT 是基于 Transformer 的预训练模型",
    "GPT 采用自回归方式生成文本"
]
embeddings = embedder.encode(docs)
vector_store.add_documents(
    ids=["doc1", "doc2", "doc3"],
    documents=docs,
    embeddings=embeddings,
    metadatas=[
        {"category": "architecture"},
        {"category": "model"},
        {"category": "model"}
    ]
)

# 搜索
query = "什么是 Transformer？"
query_emb = embedder.encode_single(query)
results = vector_store.search(query_emb, top_k=2)

print("检索结果:")
for i, doc in enumerate(results["documents"]):
    print(f"{i+1}. {doc} (相似度: {1 - results['distances'][i]:.3f})")
```

---

## 11.3 检索策略

### 策略 1：基础向量检索

```python
def basic_retrieval(query: str, top_k: int = 5):
    """基础向量检索"""
    query_emb = embedder.encode_single(query)
    results = vector_store.search(query_emb, top_k=top_k)
    return results["documents"]
```

**优点**：简单、快速  
**缺点**：可能遗漏关键信息

---

### 策略 2：混合检索（Hybrid Search）

结合向量检索和关键词检索：

```python
def hybrid_retrieval(query: str, top_k: int = 5, alpha: float = 0.7):
    """混合检索：向量 + 关键词"""
    
    # 1. 向量检索
    vector_results = vector_search(query, top_k=top_k * 2)
    
    # 2. 关键词检索（BM25）
    keyword_results = bm25_search(query, top_k=top_k * 2)
    
    # 3. 融合排序（Reciprocal Rank Fusion）
    fused_results = reciprocal_rank_fusion(
        [vector_results, keyword_results],
        weights=[alpha, 1 - alpha]
    )
    
    return fused_results[:top_k]

def reciprocal_rank_fusion(results_list: List, weights: List[float], k: int = 60):
    """倒数排名融合算法"""
    scores = {}
    
    for results, weight in zip(results_list, weights):
        for rank, doc_id in enumerate(results):
            if doc_id not in scores:
                scores[doc_id] = 0
            scores[doc_id] += weight / (rank + k)
    
    # 按分数排序
    sorted_docs = sorted(scores.items(), key=lambda x: x[1], reverse=True)
    return [doc_id for doc_id, score in sorted_docs]
```

**效果对比：**

```
问题: "Transformer 的训练技巧"

纯向量检索:
1. Transformer 架构介绍（相似度 0.85）
2. Attention 机制原理（相似度 0.82）
3. BERT 预训练（相似度 0.75）

混合检索:
1. Transformer 训练技巧详解（综合分数 0.92）← 包含关键词 "训练技巧"
2. Transformer 架构介绍（综合分数 0.88）
3. 优化 Transformer 性能（综合分数 0.85）
```

---

### 策略 3：重排序（Reranking）

使用专门的 Reranker 模型重新排序检索结果：

```python
from sentence_transformers import CrossEncoder

class Reranker:
    """重排序器"""
    
    def __init__(self, model_name: str = "cross-encoder/ms-marco-MiniLM-L-12-v2"):
        self.model = CrossEncoder(model_name)
    
    def rerank(
        self,
        query: str,
        documents: List[str],
        top_k: int = 5
    ) -> List[dict]:
        """重新排序文档"""
        
        # 计算 query-document 对的相关性分数
        pairs = [[query, doc] for doc in documents]
        scores = self.model.predict(pairs)
        
        # 排序
        results = [
            {"document": doc, "score": score}
            for doc, score in zip(documents, scores)
        ]
        results.sort(key=lambda x: x["score"], reverse=True)
        
        return results[:top_k]

# 使用
def retrieval_with_reranking(query: str, initial_k: int = 20, final_k: int = 5):
    """检索 + 重排序"""
    
    # 1. 初始检索（多一些候选）
    candidates = hybrid_retrieval(query, top_k=initial_k)
    
    # 2. 重排序（精选最相关的）
    reranker = Reranker()
    final_results = reranker.rerank(query, candidates, top_k=final_k)
    
    return final_results
```

**性能提升：**

- 纯向量检索：NDCG@5 = 0.72
- 混合检索：NDCG@5 = 0.78
- 混合 + 重排序：NDCG@5 = 0.85

---

### 策略 4：查询扩展（Query Expansion）

扩展用户查询以提高召回率：

```python
def query_expansion_with_llm(query: str) -> List[str]:
    """用 LLM 扩展查询"""
    
    prompt = f"""
    原始问题：{query}
    
    生成 3 个语义相似但表达不同的问题，用于扩展检索：
    1. ...
    2. ...
    3. ...
    """
    
    response = llm.generate(prompt)
    expanded_queries = parse_queries(response)
    
    return [query] + expanded_queries

def retrieval_with_expansion(query: str, top_k: int = 5):
    """查询扩展 + 检索"""
    
    # 1. 扩展查询
    queries = query_expansion_with_llm(query)
    
    # 2. 对每个查询检索
    all_results = []
    for q in queries:
        results = vector_store.search(embedder.encode_single(q), top_k=10)
        all_results.extend(results["documents"])
    
    # 3. 去重 + 重排序
    unique_docs = list(set(all_results))
    reranker = Reranker()
    final_results = reranker.rerank(query, unique_docs, top_k=top_k)
    
    return final_results
```

**示例：**

```
原始查询: "Transformer 为什么比 RNN 快？"

扩展后:
1. "Transformer 为什么比 RNN 快？"
2. "Transformer 和 RNN 的速度对比"
3. "Transformer 并行化优势"
4. "RNN 顺序处理的局限性"

召回提升: 从 5 个相关文档 → 12 个相关文档
```

---

## 11.4 上下文构建

### 挑战：Token 限制

检索到的文档可能很长，需要合理构建上下文：

```
检索到 10 个文档，每个 500 tokens = 5000 tokens
但 LLM 上下文有限，需要优化
```

### 策略 1：Top-K 截断

```python
def build_context_topk(query: str, top_k: int = 3):
    """只使用 Top-K 文档"""
    results = retrieval_with_reranking(query, final_k=top_k)
    
    context = "\n\n---\n\n".join([
        f"文档 {i+1}:\n{doc['document']}"
        for i, doc in enumerate(results)
    ])
    
    return context
```

### 策略 2：Token 预算分配

```python
def build_context_with_budget(query: str, max_tokens: int = 2000):
    """根据 token 预算构建上下文"""
    results = retrieval_with_reranking(query, final_k=10)
    
    context_parts = []
    total_tokens = 0
    
    for doc in results:
        doc_tokens = count_tokens(doc["document"])
        
        if total_tokens + doc_tokens <= max_tokens:
            context_parts.append(doc["document"])
            total_tokens += doc_tokens
        else:
            # 预算用完，停止
            break
    
    return "\n\n---\n\n".join(context_parts)
```

### 策略 3：摘要压缩

```python
async def build_context_with_summary(query: str, top_k: int = 10):
    """压缩文档为摘要"""
    results = retrieval_with_reranking(query, final_k=top_k)
    
    summaries = []
    for doc in results:
        # 如果文档太长，生成摘要
        if count_tokens(doc["document"]) > 500:
            summary = await llm.generate(
                f"总结以下内容（100字以内）：\n{doc['document']}"
            )
            summaries.append(summary)
        else:
            summaries.append(doc["document"])
    
    return "\n\n".join(summaries)
```

---

## 11.5 完整 RAG 实现

### 端到端 RAG 系统

```python
class RAGSystem:
    """完整的 RAG 系统"""
    
    def __init__(
        self,
        embedder: EmbeddingService,
        vector_store: ChromaVectorStore,
        llm: LLM,
        reranker: Reranker = None
    ):
        self.embedder = embedder
        self.vector_store = vector_store
        self.llm = llm
        self.reranker = reranker
    
    async def index_documents(self, documents: List[dict]):
        """索引文档"""
        
        # 1. 分块
        all_chunks = []
        for doc in documents:
            chunks = self.chunk_document(doc["content"])
            for i, chunk in enumerate(chunks):
                all_chunks.append({
                    "id": f"{doc['id']}_chunk_{i}",
                    "content": chunk,
                    "metadata": doc.get("metadata", {})
                })
        
        # 2. 生成 Embedding
        contents = [chunk["content"] for chunk in all_chunks]
        embeddings = self.embedder.encode(contents)
        
        # 3. 存入向量数据库
        self.vector_store.add_documents(
            ids=[chunk["id"] for chunk in all_chunks],
            documents=contents,
            embeddings=embeddings,
            metadatas=[chunk["metadata"] for chunk in all_chunks]
        )
    
    async def query(
        self,
        question: str,
        top_k: int = 5,
        max_context_tokens: int = 2000
    ) -> dict:
        """RAG 查询"""
        
        # 1. 检索
        query_emb = self.embedder.encode_single(question)
        initial_results = self.vector_store.search(
            query_emb,
            top_k=top_k * 2  # 检索更多候选
        )
        
        # 2. 重排序（可选）
        if self.reranker:
            reranked = self.reranker.rerank(
                question,
                initial_results["documents"],
                top_k=top_k
            )
            documents = [r["document"] for r in reranked]
        else:
            documents = initial_results["documents"][:top_k]
        
        # 3. 构建上下文
        context = self.build_context(documents, max_context_tokens)
        
        # 4. 生成答案
        prompt = f"""
        基于以下上下文回答问题。如果上下文中没有相关信息，请明确说明。
        
        ## 上下文
        {context}
        
        ## 问题
        {question}
        
        ## 要求
        1. 基于上下文回答，不要编造信息
        2. 引用具体的上下文来源
        3. 如果不确定，说明并建议如何获取更多信息
        """
        
        answer = await self.llm.generate(prompt)
        
        return {
            "question": question,
            "answer": answer,
            "sources": documents,
            "context": context
        }
    
    def chunk_document(self, content: str, chunk_size: int = 500):
        """文档分块"""
        splitter = RecursiveCharacterTextSplitter(
            chunk_size=chunk_size,
            chunk_overlap=50
        )
        return splitter.split(content)
    
    def build_context(self, documents: List[str], max_tokens: int):
        """构建上下文"""
        context_parts = []
        total_tokens = 0
        
        for i, doc in enumerate(documents):
            doc_tokens = count_tokens(doc)
            if total_tokens + doc_tokens <= max_tokens:
                context_parts.append(f"[文档 {i+1}]\n{doc}")
                total_tokens += doc_tokens
            else:
                break
        
        return "\n\n".join(context_parts)

# 使用示例
async def main():
    # 初始化
    embedder = EmbeddingService()
    vector_store = ChromaVectorStore()
    vector_store.create_collection("knowledge_base")
    llm = Claude35Sonnet()
    reranker = Reranker()
    
    rag = RAGSystem(embedder, vector_store, llm, reranker)
    
    # 索引文档
    documents = [
        {
            "id": "doc1",
            "content": "Transformer 论文内容...",
            "metadata": {"source": "paper", "title": "Attention is All You Need"}
        },
        {
            "id": "doc2",
            "content": "BERT 介绍...",
            "metadata": {"source": "blog", "title": "BERT Explained"}
        }
    ]
    await rag.index_documents(documents)
    
    # 查询
    result = await rag.query("Transformer 的核心创新是什么？", top_k=3)
    
    print(f"问题: {result['question']}")
    print(f"答案: {result['answer']}")
    print(f"来源: {len(result['sources'])} 个文档")
```

---

## 11.6 评估与优化

### 评估指标

**1. 检索质量（Retrieval Quality）**

```python
def evaluate_retrieval(queries: List[dict], k: int = 5):
    """评估检索质量"""
    
    metrics = {
        "recall": [],
        "precision": [],
        "ndcg": []
    }
    
    for query in queries:
        question = query["question"]
        relevant_docs = set(query["relevant_doc_ids"])
        
        # 检索
        results = rag.vector_store.search(
            rag.embedder.encode_single(question),
            top_k=k
        )
        retrieved_docs = set(results["ids"])
        
        # 计算指标
        recall = len(retrieved_docs & relevant_docs) / len(relevant_docs)
        precision = len(retrieved_docs & relevant_docs) / len(retrieved_docs)
        
        metrics["recall"].append(recall)
        metrics["precision"].append(precision)
    
    return {
        "avg_recall": np.mean(metrics["recall"]),
        "avg_precision": np.mean(metrics["precision"])
    }
```

**2. 答案质量（Answer Quality）**

```python
def evaluate_answer_quality(test_cases: List[dict]):
    """评估答案质量"""
    
    scores = []
    for case in test_cases:
        question = case["question"]
        reference_answer = case["answer"]
        
        # 生成答案
        result = await rag.query(question)
        generated_answer = result["answer"]
        
        # 用 LLM 评估
        eval_prompt = f"""
        评估以下答案的质量（1-5分）：
        
        问题：{question}
        参考答案：{reference_answer}
        生成答案：{generated_answer}
        
        评分标准：
        - 准确性：是否正确？
        - 完整性：是否完整？
        - 相关性：是否相关？
        
        返回 JSON：{{"score": 1-5, "reason": "..."}}
        """
        
        eval_result = await llm.generate(eval_prompt)
        score = parse_json(eval_result)["score"]
        scores.append(score)
    
    return np.mean(scores)
```

### 优化策略

**1. Embedding 模型微调**

```python
from sentence_transformers import SentenceTransformer, InputExample, losses
from torch.utils.data import DataLoader

def finetune_embedding_model(training_data: List[dict]):
    """微调 Embedding 模型"""
    
    # 加载预训练模型
    model = SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')
    
    # 准备训练数据
    train_examples = []
    for item in training_data:
        train_examples.append(InputExample(
            texts=[item["query"], item["positive_doc"]],
            label=1.0  # 正样本
        ))
        train_examples.append(InputExample(
            texts=[item["query"], item["negative_doc"]],
            label=0.0  # 负样本
        ))
    
    # 训练
    train_dataloader = DataLoader(train_examples, shuffle=True, batch_size=16)
    train_loss = losses.CosineSimilarityLoss(model)
    
    model.fit(
        train_objectives=[(train_dataloader, train_loss)],
        epochs=3,
        warmup_steps=100
    )
    
    return model
```

**2. 动态分块策略**

```python
def adaptive_chunking(document: str, query: str = None):
    """根据查询动态调整分块"""
    
    if query:
        # 如果有查询，围绕关键词分块
        keywords = extract_keywords(query)
        chunks = chunk_around_keywords(document, keywords)
    else:
        # 否则使用标准分块
        chunks = standard_chunking(document)
    
    return chunks
```

---

## 本章小结

### 核心要点回顾

**1. RAG 核心原理**
- 检索增强生成：检索相关知识 + LLM 生成答案
- 解决 LLM 的知识截止、私有数据、专业领域问题
- 基本流程：问题 → 检索 → 上下文 → 生成

**2. 核心组件**
- **文档分块**：固定大小、语义、递归分块
- **Embedding**：OpenAI、Cohere、开源模型
- **向量数据库**：Chroma、Pinecone、Milvus、Weaviate

**3. 检索策略**
- 基础向量检索：快速但可能遗漏
- 混合检索：向量 + 关键词
- 重排序：提高精确度
- 查询扩展：提高召回率

**4. 上下文构建**
- Top-K 截断：简单但可能丢失信息
- Token 预算：动态选择文档
- 摘要压缩：压缩长文档

**5. 评估与优化**
- 检索质量：Recall、Precision、NDCG
- 答案质量：LLM 评估
- 优化：模型微调、动态分块

### 关键洞察

> **RAG 不是简单的"检索 + 生成"，而是一个需要精心设计的端到端系统，涉及分块、Embedding、检索、重排序、上下文构建等多个环节。每个环节的优化都会影响最终效果。**

### 实战建议

1. **从简单开始**：先用基础向量检索，再逐步优化
2. **重视评估**：建立评估数据集，持续优化
3. **选择合适的工具**：根据规模和场景选择向量数据库
4. **关注成本**：Embedding 和向量存储都有成本
5. **混合策略**：结合向量检索和关键词检索

---

## 参考资料

1. **Retrieval-Augmented Generation for Knowledge-Intensive NLP Tasks** - Lewis et al., 2020
2. **Dense Passage Retrieval for Open-Domain Question Answering** - Karpukhin et al., 2020
3. **LlamaIndex Documentation** - https://docs.llamaindex.ai/
4. **LangChain RAG Tutorial** - https://python.langchain.com/docs/tutorials/rag/
5. **Pinecone RAG Guide** - https://www.pinecone.io/learn/retrieval-augmented-generation/
6. **Chroma Documentation** - https://docs.trychroma.com/
