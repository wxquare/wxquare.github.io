# 第8章 DoD Agent：电商告警自动处理系统

> "A real-world agent is 20% AI magic and 80% engineering discipline." （真实的Agent系统，20%是AI魔法，80%是工程纪律）

## 引言

前面七章我们学习了 AI 编程范式、Claude Code 工具、Harness Engineering、Agent 架构设计、工具系统、多 Agent 协作、可观测性等理论知识。现在是时候看一个完整的实战案例了。

DoD Agent（Developer on Duty Agent）是一个电商告警自动处理系统，用于辅助值班工程师快速诊断和处理生产告警。本章将展示从需求分析到架构设计，从核心实现到上线部署的完整过程。

---

## 8.1 项目背景与痛点

### 业务背景

电商公司日常运维面临大量告警处理工作，包括：
- **基础设施告警**：CPU、内存、磁盘使用率
- **应用告警**：错误率、响应时间、超时
- **业务告警**：订单量异常、支付失败率

### 当前痛点

| 痛点 | 影响 |
|:---|:---|
| **告警量大**（50-200条/天） | 值班人员疲劳，响应延迟 |
| **重复性问题多** | 80% 问题有标准处理流程，但每次都需人工诊断 |
| **知识分散** | Confluence 有200+文档，难以快速定位 |
| **跨部门协作** | 告警升级和分发效率低 |
| **新人上手慢** | 需要2-3个月才能独立值班 |

### 目标与成功标准

**定量目标：**
- 自动诊断率 ≥ 60%
- 诊断准确率 ≥ 85%
- MTTR（平均恢复时间）降低 30%
- 值班人员工作量减少 30%

**定性目标：**
- 知识沉淀和复用
- 新人上手时间缩短到 1 个月
- 值班人员满意度提升

---

## 8.2 设计决策与架构选择

### 为什么需要 Agent？

基于第4章的决策框架：

- ✅ Q1: 告警描述是自然语言
- ✅ Q2: 告警场景复杂多变（50+ 种类型）
- ✅ Q3: 需要多步骤诊断（查指标 → 查日志 → 查 K8s）
- ✅ Q4: 需要整合多个系统（Prometheus、Loki、K8s、Confluence）
- ✅ Q5: 85-95% 准确率可接受（人工兜底）
- ✅ Q6: 诊断时间 10-30s 可接受

**结论**：Agent 是合适的选择。

### 架构演进

**V1：规则引擎方案**
- 需要维护 500+ 规则（50 告警类型 × 10 服务）
- 新增告警需要修改代码
- 无法处理复杂的上下文关联

**V2：纯 ReACT Agent**
- 灵活但难以控制
- 执行路径不可预测
- 缺少状态管理

**V3：状态机 + ReACT 混合架构** ✅
- 状态机管理生命周期（可控）
- ReACT 负责智能诊断（灵活）
- 决策引擎分级处理（安全）

### 整体架构

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                          DoD Agent System                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                      Input Layer (输入层)                        │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐    │   │
│  │  │Alertmanager│  │  Grafana  │  │  Seatalk  │  │  Web API  │    │   │
│  │  │  Webhook  │  │  Webhook  │  │  Message  │  │  Request  │    │   │
│  │  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘    │   │
│  └────────┼──────────────┼──────────────┼──────────────┼──────────┘   │
│           │              │              │              │               │
│           ▼              ▼              ▼              ▼               │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Gateway (API 网关)                            │   │
│  │  • 统一接入  • 认证鉴权  • 消息标准化  • 限流熔断               │   │
│  └─────────────────────────────┬───────────────────────────────────┘   │
│                                │                                       │
│                                ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    DoD Agent 核心                                │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │   │
│  │  │ 状态机控制器 │  │ ReACT 引擎   │  │ 决策引擎     │          │   │
│  │  │ (Lifecycle)  │◄─┤ (智能分析)   │◄─┤ (分级策略)   │          │   │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │   │
│  │         │                 │                 │                    │   │
│  │         └─────────────────┼─────────────────┘                    │   │
│  │                           ▼                                      │   │
│  │                   ┌──────────────┐                               │   │
│  │                   │ 上下文管理器 │                               │   │
│  │                   │  (Memory)    │                               │   │
│  │                   └──────┬───────┘                               │   │
│  │                          │                                       │   │
│  │                          ▼                                       │   │
│  │                   ┌──────────────┐                               │   │
│  │                   │  工具调用层  │                               │   │
│  │                   │  (MCP Tools) │                               │   │
│  │                   └──────────────┘                               │   │
│  └────────────────────┬────────────────────────────────────────────┘   │
│                       │                                                │
│                       ▼                                                │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    RAG 知识库                                    │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐                   │   │
│  │  │Confluence │  │ Runbooks  │  │ 历史案例  │                   │   │
│  │  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘                   │   │
│  │        └───────────────┼───────────────┘                         │   │
│  │                        ▼                                         │   │
│  │              ┌──────────────────┐                                │   │
│  │              │  Vector Database │                                │   │
│  │              │  (Chroma/Milvus) │                                │   │
│  │              └──────────────────┘                                │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 8.3 核心模块实现

### 状态机设计

告警处理生命周期由状态机管理：

```go
type AlertState string

const (
    // 初始状态
    StateNew AlertState = "NEW"
    
    // 分析中
    StateAnalyzing AlertState = "ANALYZING"
    
    // 等待确认（中高风险）
    StateWaitingConfirm AlertState = "WAITING_CONFIRM"
    
    // 自动处理中（低风险）
    StateAutoResolving AlertState = "AUTO_RESOLVING"
    
    // 执行SOP中
    StateExecutingSOP AlertState = "EXECUTING_SOP"
    
    // 已通知DoD
    StateDoDNotified AlertState = "DOD_NOTIFIED"
    
    // 已解决
    StateResolved AlertState = "RESOLVED"
    
    // 已关闭
    StateClosed AlertState = "CLOSED"
    
    // 失败
    StateFailed AlertState = "FAILED"
)
```

**状态转换规则：**

```text
NEW → ANALYZING → (决策引擎判断)
    ├─ 低风险 → AUTO_RESOLVING → RESOLVED
    ├─ 中风险 → WAITING_CONFIRM → EXECUTING_SOP → RESOLVED
    ├─ 高风险 → WAITING_CONFIRM → DOD_NOTIFIED
    └─ 严重 → DOD_NOTIFIED
```

### ReACT 诊断引擎

```python
class ReACTEngine:
    """ReACT 诊断引擎"""
    
    def __init__(self, llm, tools: ToolRegistry, rag: RAGSystem):
        self.llm = llm
        self.tools = tools
        self.rag = rag
        self.max_iterations = 10
    
    async def diagnose(self, alert: StandardAlert, state: AlertState) -> DiagnosisResult:
        """诊断告警"""
        
        # 1. 构建初始上下文
        context = self._build_alert_context(alert)
        
        # 2. 检索相关知识
        knowledge = await self.rag.retrieve(
            query=f"{alert.title} {alert.description}",
            filters={"service": alert.labels.get("service")}
        )
        context += f"\n\n## 相关知识文档\n{knowledge}"
        
        # 3. ReACT 循环
        diagnosis_steps = []
        allowed_tools = self._get_allowed_tools(state)
        
        for i in range(self.max_iterations):
            # 生成下一步行动
            response = await self.llm.generate(
                self._build_diagnosis_prompt(context, diagnosis_steps, state, allowed_tools)
            )
            
            action = self._parse_action(response)
            
            # 检查是否完成诊断
            if action.type == "final_diagnosis":
                return DiagnosisResult(
                    alert_id=alert.id,
                    root_cause=action.root_cause,
                    impact=action.impact,
                    suggested_actions=action.suggested_actions,
                    confidence=action.confidence,
                    steps=diagnosis_steps
                )
            
            # 执行工具调用
            if action.type == "tool_call":
                if action.tool not in allowed_tools:
                    diagnosis_steps.append({
                        "thought": action.thought,
                        "error": f"工具 {action.tool} 不可用"
                    })
                    continue
                
                result = await self.tools.execute(action.tool, **action.args)
                diagnosis_steps.append({
                    "thought": action.thought,
                    "tool": action.tool,
                    "args": action.args,
                    "result": result
                })
                
                # 更新上下文
                context += f"\n\n观察: {result}"
        
        # 达到最大迭代，返回部分结果
        return self._build_partial_result(alert, diagnosis_steps)
```

### 决策引擎

基于诊断结果进行分级决策：

```python
class RiskLevel(Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"

class DecisionEngine:
    """决策引擎"""
    
    def __init__(self):
        self.rules = self._load_decision_rules()
    
    def decide(self, diagnosis: DiagnosisResult, alert: StandardAlert) -> Decision:
        """根据诊断结果做出决策"""
        
        # 1. 评估风险等级
        risk = self._assess_risk(diagnosis, alert)
        
        # 2. 根据风险等级决策
        if risk == RiskLevel.LOW:
            return Decision(
                action="auto_resolve",
                risk_level=risk,
                require_confirm=False,
                message="低风险，自动处理"
            )
        
        elif risk == RiskLevel.MEDIUM:
            return Decision(
                action="suggest_and_confirm",
                risk_level=risk,
                require_confirm=True,
                confirm_timeout=300,  # 5 分钟
                message="中风险，建议处理方案，等待确认"
            )
        
        elif risk == RiskLevel.HIGH:
            return Decision(
                action="escalate_to_dod",
                risk_level=risk,
                require_confirm=True,
                message="高风险，需要 DoD 介入"
            )
        
        else:  # CRITICAL
            return Decision(
                action="immediate_escalation",
                risk_level=risk,
                require_confirm=False,
                message="严重告警，立即升级"
            )
    
    def _assess_risk(self, diagnosis: DiagnosisResult, alert: StandardAlert) -> RiskLevel:
        """评估风险等级"""
        
        # 因素 1：告警严重级别
        if alert.severity == AlertSeverity.CRITICAL:
            base_risk = 3
        elif alert.severity == AlertSeverity.WARNING:
            base_risk = 2
        else:
            base_risk = 1
        
        # 因素 2：诊断置信度
        if diagnosis.confidence < 0.7:
            base_risk += 1  # 置信度低，提升风险等级
        
        # 因素 3：影响范围
        if "核心服务" in diagnosis.impact or "支付" in diagnosis.impact:
            base_risk += 1
        
        # 因素 4：是否有已知解决方案
        if diagnosis.has_known_solution:
            base_risk -= 1
        
        # 映射到风险等级
        risk_map = {
            1: RiskLevel.LOW,
            2: RiskLevel.MEDIUM,
            3: RiskLevel.HIGH,
            4: RiskLevel.CRITICAL
        }
        
        return risk_map.get(max(1, min(4, base_risk)), RiskLevel.MEDIUM)
```

---

## 8.4 工具系统实现

### 工具注册表

```go
type Tool struct {
    Name        string
    Description string
    RiskLevel   RiskLevel
    Execute     func(args map[string]interface{}) (interface{}, error)
    Schema      *ToolSchema
}

type ToolRegistry struct {
    tools map[string]*Tool
}

func NewToolRegistry() *ToolRegistry {
    registry := &ToolRegistry{
        tools: make(map[string]*Tool),
    }
    
    // 注册只读工具（低风险）
    registry.Register(&Tool{
        Name: "prometheus_query",
        Description: "查询 Prometheus 指标",
        RiskLevel: RiskLevelLow,
        Execute: executePrometheusQuery,
        Schema: &ToolSchema{
            Parameters: map[string]interface{}{
                "query": "PromQL 查询表达式",
                "time_range": "时间范围（如 1h, 30m）",
            },
        },
    })
    
    registry.Register(&Tool{
        Name: "loki_search",
        Description: "搜索日志",
        RiskLevel: RiskLevelLow,
        Execute: executeLokiSearch,
    })
    
    // 注册写操作工具（高风险）
    registry.Register(&Tool{
        Name: "restart_service",
        Description: "重启服务",
        RiskLevel: RiskLevelHigh,
        Execute: executeRestartService,
    })
    
    return registry
}
```

### 关键工具实现

**Prometheus 查询工具：**

```python
class PrometheusQueryTool(Tool):
    """Prometheus 指标查询工具"""
    
    name = "prometheus_query"
    description = "查询 Prometheus 指标数据"
    
    def __init__(self, prometheus_url: str):
        self.url = prometheus_url
        self.client = httpx.AsyncClient()
    
    async def execute(self, query: str, time_range: str = "1h") -> str:
        """执行查询"""
        try:
            response = await self.client.get(
                f"{self.url}/api/v1/query",
                params={"query": query, "time": self._parse_time_range(time_range)}
            )
            data = response.json()
            
            if data["status"] != "success":
                return f"查询失败: {data.get('error', 'Unknown error')}"
            
            # 格式化结果
            return self._format_results(data["data"]["result"])
            
        except Exception as e:
            return f"查询异常: {str(e)}"
    
    def _format_results(self, results: List) -> str:
        """格式化查询结果为可读文本"""
        if not results:
            return "没有查询到数据"
        
        formatted = []
        for metric in results:
            labels = ", ".join([f"{k}={v}" for k, v in metric["metric"].items()])
            value = metric["value"][1]
            formatted.append(f"{labels}: {value}")
        
        return "\n".join(formatted)
```

**知识库检索工具：**

```python
class KnowledgeSearchTool(Tool):
    """知识库检索工具"""
    
    name = "search_knowledge_base"
    description = "从 Confluence 知识库检索相关文档"
    
    def __init__(self, rag_system: RAGSystem):
        self.rag = rag_system
    
    async def execute(self, query: str, filters: Dict = None) -> str:
        """执行检索"""
        docs = await self.rag.retrieve(
            query=query,
            filters=filters,
            top_k=3
        )
        
        return self._format_docs(docs)
    
    def _format_docs(self, docs: List[DocumentChunk]) -> str:
        """格式化文档为可读文本"""
        formatted = []
        for i, doc in enumerate(docs):
            formatted.append(
                f"文档 {i+1}: {doc.metadata['title']}\n"
                f"链接: {doc.metadata['url']}\n"
                f"内容: {doc.content[:500]}...\n"
            )
        return "\n---\n".join(formatted)
```

---

## 8.5 RAG知识库系统

### 文档加载与处理

```python
class DocumentProcessor:
    """文档处理流水线"""
    
    def __init__(self):
        self.loader = ConfluenceLoader()
        self.chunker = DocumentChunker(chunk_size=500, overlap=50)
        self.embedder = OpenAIEmbedding()
    
    async def process_space(self, space_key: str) -> List[DocumentChunk]:
        """处理整个 Confluence 空间"""
        
        # 1. 加载文档
        documents = await self.loader.load_space(space_key)
        print(f"加载了 {len(documents)} 个文档")
        
        # 2. 分块
        all_chunks = []
        for doc in documents:
            chunks = self.chunker.chunk(doc)
            all_chunks.extend(chunks)
        print(f"生成了 {len(all_chunks)} 个文档块")
        
        # 3. 生成 Embedding
        for chunk in all_chunks:
            chunk.embedding = await self.embedder.encode(chunk.content)
        
        return all_chunks
```

### 向量数据库集成

```python
class VectorDatabase:
    """向量数据库抽象"""
    
    async def insert(self, chunks: List[DocumentChunk]):
        """批量插入文档块"""
        pass
    
    async def search(
        self,
        vector: List[float],
        top_k: int = 5,
        filters: Dict = None
    ) -> List[DocumentChunk]:
        """向量相似度搜索"""
        pass
    
    async def delete_by_metadata(self, filters: Dict):
        """根据元数据删除"""
        pass

# Chroma 实现
class ChromaVectorDB(VectorDatabase):
    """Chroma 向量数据库实现"""
    
    def __init__(self, persist_dir: str):
        self.client = chromadb.PersistentClient(path=persist_dir)
        self.collection = self.client.get_or_create_collection(
            name="dod_knowledge",
            metadata={"hnsw:space": "cosine"}
        )
    
    async def insert(self, chunks: List[DocumentChunk]):
        """插入文档块"""
        self.collection.add(
            ids=[chunk.id for chunk in chunks],
            embeddings=[chunk.embedding for chunk in chunks],
            documents=[chunk.content for chunk in chunks],
            metadatas=[chunk.metadata for chunk in chunks]
        )
    
    async def search(
        self,
        vector: List[float],
        top_k: int = 5,
        filters: Dict = None
    ) -> List[DocumentChunk]:
        """搜索相似文档"""
        results = self.collection.query(
            query_embeddings=[vector],
            n_results=top_k,
            where=filters
        )
        
        chunks = []
        for i in range(len(results["ids"][0])):
            chunks.append(DocumentChunk(
                id=results["ids"][0][i],
                content=results["documents"][0][i],
                metadata=results["metadatas"][0][i]
            ))
        
        return chunks
```

---

## 8.6 完整处理流程示例

### 真实案例：CPU使用率告警

**告警信息：**
```yaml
alertname: HighCPUUsage
service: order-service
namespace: production
pod: order-service-7d8f9c-abc123
severity: warning
value: 92%
threshold: 80%
description: order-service CPU 使用率持续高于 80%，当前值 92%
```

**处理流程：**

**Step 1: 状态初始化**
```text
状态: NEW → ANALYZING
```

**Step 2: ReACT 诊断**

```text
迭代 1:
Thought: CPU 高可能是最近有部署，先查看最近的部署记录
Action: kubernetes_get
Action Input: {"resource": "deployments", "namespace": "production"}
Observation: 最近 2 小时内有 3 次部署，最新一次是 15 分钟前

迭代 2:
Thought: 最新部署可能引入了性能问题，查看错误日志
Action: loki_search
Action Input: {"query": "{service=\"order-service\"} |= \"error\"", "time_range": "30m"}
Observation: 发现大量 "database connection timeout" 错误

迭代 3:
Thought: 数据库连接超时导致请求堆积，查看数据库连接池指标
Action: prometheus_query
Action Input: {"query": "db_connection_pool_active{service=\"order-service\"}", "time_range": "1h"}
Observation: 连接池已满（50/50），大量请求等待连接

迭代 4:
Thought: 确认是数据库连接池配置不足，查看历史案例
Action: search_similar_alerts
Action Input: {"alert_name": "HighCPUUsage", "service": "order-service"}
Observation: 找到 2 个历史案例，都是通过增加连接池大小解决

Final Diagnosis:
root_cause: "数据库连接池配置不足（50），导致请求堆积，CPU 等待 I/O"
impact: "订单服务响应变慢，可能影响用户下单"
suggested_actions: [
  "临时方案：重启服务释放连接",
  "长期方案：将连接池大小调整为 100"
]
confidence: 0.9
references: [
  "https://confluence.company.com/wiki/order-service-tuning",
  "历史案例 #1234"
]
```

**Step 3: 决策引擎判断**

```text
评估因素:
- 告警严重级别: WARNING (基础风险 = 2)
- 诊断置信度: 0.9 (高置信，风险 -0)
- 影响范围: 订单服务（核心，风险 +1）
- 有已知方案: 是（风险 -1）

最终风险等级: MEDIUM (2)

决策: 建议处理方案 + 等待确认
状态: ANALYZING → WAITING_CONFIRM
```

**Step 4: 发送通知**

```text
📬 发送到 Seatalk #oncall 频道：

🔔 *告警诊断报告*

*告警*: HighCPUUsage
*严重级别*: WARNING
*服务*: order-service

---

📋 *根因分析* (置信度: 90%)
数据库连接池配置不足（50），导致请求堆积，CPU 等待 I/O

⚠️ *影响范围*
订单服务响应变慢，可能影响用户下单

✅ *建议处理步骤*
1. 临时方案：重启服务释放连接
2. 长期方案：将连接池大小调整为 100

📚 *参考文档*
- [Order Service 性能调优指南](https://confluence...)
- [历史案例 #1234](https://...)

---

💡 *快速操作*
回复 "1" 执行临时方案（重启服务）
回复 "2" 查看详细诊断步骤
回复 "3" 升级给 DoD
```

**Step 5: 等待确认**

```text
用户回复: "1"

状态: WAITING_CONFIRM → EXECUTING_SOP
执行: 重启 order-service
验证: CPU 降到 45%
状态: EXECUTING_SOP → RESOLVED

最终通知:
✅ 告警已处理，order-service 已重启，CPU 恢复正常（45%）
建议后续调整连接池配置避免复发
```

---

## 8.7 关键设计决策与权衡

### 决策 1：状态机 vs 纯 ReACT

**选择**：状态机 + ReACT 混合

**理由：**
- 状态机提供可控性和可观测性
- ReACT 提供灵活性和智能性
- 混合方案兼顾两者优势

### 决策 2：单体 vs 多 Agent

**选择**：单体 Agent

**理由：**
- 告警处理流程相对简单，单体足够
- 避免多 Agent 协调的复杂度
- 降低成本（一个 LLM 调用 vs 多个）

### 决策 3：LLM 选择

**选择**：GPT-4 / Claude 3.5 Sonnet

**理由：**
- 推理能力强，工具调用准确
- 成本可接受（$0.36/次）
- API 稳定可靠

### 决策 4：部署方式

**选择**：Kubernetes Deployment

**理由：**
- 利用现有基础设施
- 易于扩展和升级
- 完善的监控和日志

---

## 8.8 效果评估

### 上线后数据（运行 1 个月）

| 指标 | 目标 | 实际 | 达成情况 |
|------|------|------|---------|
| **自动诊断率** | ≥ 60% | 68% | ✅ 超额完成 |
| **诊断准确率** | ≥ 85% | 87% | ✅ 达成 |
| **MTTR 降低** | 30% | 35% | ✅ 超额完成 |
| **工作量减少** | 30% | 32% | ✅ 达成 |
| **成本** | < $1000/月 | $720/月 | ✅ 达成 |

### 用户反馈

**正面反馈：**
- "诊断速度快，30 秒内就有结果"
- "历史案例检索很有用，能快速参考之前怎么处理的"
- "新人上手快了很多，直接看 Agent 的诊断就能学习"

**改进建议：**
- "有些诊断太冗长，能不能更简洁？"
- "希望能自动执行低风险操作，不用每次都确认"
- "能否支持更多告警类型？"

### 迭代优化

基于反馈进行了 3 轮迭代：

**v1.1**：优化诊断报告格式，更简洁
**v1.2**：增加自动执行低风险操作
**v1.3**：扩展到 30+ 种新告警类型

---

## 本章小结

### 核心要点回顾

**1. 项目背景**
- 电商告警处理系统，解决值班人员疲劳和重复性工作
- 明确的量化目标和成功标准

**2. 架构设计**
- 状态机 + ReACT 混合架构
- 决策引擎分级处理
- 工具系统标准化接口
- RAG 知识库检索

**3. 核心模块**
- **状态机**：管理告警生命周期
- **ReACT 引擎**：智能诊断和工具调用
- **决策引擎**：风险评估和分级决策
- **工具系统**：Prometheus、Loki、K8s 等集成
- **RAG 系统**：Confluence 知识库检索

**4. 关键设计决策**
- 状态机 vs 纯 ReACT → 混合方案
- 单体 vs 多 Agent → 单体足够
- LLM 选择 → GPT-4 / Claude 3.5
- 部署方式 → Kubernetes

**5. 效果评估**
- 所有目标全部达成或超额完成
- 自动诊断率 68%，准确率 87%
- MTTR 降低 35%，成本 $720/月

### 关键洞察

> **一个成功的 Agent 系统，20% 是 AI 魔法（LLM 推理），80% 是工程纪律（架构、工具、验证、监控）。**

### 经验总结

**成功经验：**
1. **明确的需求和目标**：量化指标，可验证
2. **渐进式设计**：从只读诊断开始，逐步增强
3. **完整的验证体系**：状态机 + 决策引擎 + 人工兜底
4. **持续优化**：基于反馈快速迭代

**踩过的坑：**
1. **初期 Prompt 太长**：精简后效果更好
2. **工具返回结果太详细**：压缩后更高效
3. **缺少超时控制**：长时间运行影响体验

### 下一章预告

第9章我们将探讨另一个实战案例：个人知识管理 Agent，展示如何将 Agent 应用于个人生产力场景，包括知识整理、学习路径规划、文档自动化等。

---

## 参考资料

1. **DoD Agent 完整设计文档** - 内部文档
2. **ReACT: Reasoning and Acting** - Yao et al., 2022
3. **LangGraph State Machine Pattern** - LangChain Docs
4. **Prometheus Query API** - https://prometheus.io/docs/prometheus/latest/querying/api/
5. **Confluence REST API** - https://developer.atlassian.com/cloud/confluence/rest/
