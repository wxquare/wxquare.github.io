# 第7章 可观测性与成本优化

> "You can't improve what you don't measure." （无法衡量，就无法改进）

## 引言

Agent 系统的可观测性与传统系统有本质区别。传统系统监控的是代码执行路径和资源消耗，而 Agent 系统还需要监控：
- LLM 的推理过程
- 工具调用链路
- 上下文使用情况
- Token 消耗和成本

本章将深入探讨 Agent 系统的可观测性设计，以及如何在保证质量的前提下优化成本。

---

## 7.1 Agent可观测性的三个层次

### 层次 1：基础指标（Metrics）

**系统级指标：**
```text
- 请求总数
- 平均响应时间
- 成功率 / 失败率
- P50 / P95 / P99 延迟
- 系统资源使用（CPU、内存）
```

**Agent 级指标：**
```text
- LLM 调用次数
- Token 消耗（input + output）
- 工具调用次数
- 平均迭代轮数
- 最终成功率
```

**实现示例：**
```python
class AgentMetrics:
    def __init__(self):
        self.llm_calls = 0
        self.token_usage = {"input": 0, "output": 0}
        self.tool_calls = {}
        self.iterations = 0
    
    def record_llm_call(self, input_tokens: int, output_tokens: int):
        self.llm_calls += 1
        self.token_usage["input"] += input_tokens
        self.token_usage["output"] += output_tokens
    
    def record_tool_call(self, tool_name: str):
        self.tool_calls[tool_name] = self.tool_calls.get(tool_name, 0) + 1
    
    def record_iteration(self):
        self.iterations += 1
    
    def get_summary(self):
        return {
            "llm_calls": self.llm_calls,
            "token_usage": self.token_usage,
            "tool_calls": self.tool_calls,
            "iterations": self.iterations,
            "estimated_cost": self.calculate_cost()
        }
    
    def calculate_cost(self):
        # GPT-4 定价（示例）
        input_cost = self.token_usage["input"] / 1000 * 0.03
        output_cost = self.token_usage["output"] / 1000 * 0.06
        return input_cost + output_cost
```

### 层次 2：执行追踪（Tracing）

记录 Agent 的完整执行路径。

**追踪内容：**
```json
{
  "trace_id": "trace_abc123",
  "session_id": "sess_xyz789",
  "user_id": "user_456",
  "start_time": "2026-04-21T10:30:00Z",
  "end_time": "2026-04-21T10:30:45Z",
  "duration_ms": 45000,
  "spans": [
    {
      "span_id": "span_1",
      "name": "llm_generate",
      "start_time": "2026-04-21T10:30:00Z",
      "duration_ms": 2000,
      "attributes": {
        "model": "gpt-4",
        "input_tokens": 500,
        "output_tokens": 150
      }
    },
    {
      "span_id": "span_2",
      "name": "tool_call:prometheus_query",
      "start_time": "2026-04-21T10:30:02Z",
      "duration_ms": 300,
      "attributes": {
        "tool": "prometheus_query",
        "args": {"query": "rate(cpu[5m])"}
      }
    }
  ]
}
```

**可视化：**
```text
trace_abc123
├─ llm_generate (2s)
├─ tool:prometheus_query (0.3s)
├─ llm_generate (1.8s)
├─ tool:loki_search (0.5s)
├─ llm_generate (1.5s)
└─ tool:kubernetes_get (0.4s)

Total: 6.5s
```

### 层次 3：行为分析（Behavior Analysis）

分析 Agent 的决策过程和行为模式。

**关键问题：**
- Agent 为什么选择这个工具？
- 推理路径是否合理？
- 在哪个步骤偏离了预期？
- 什么样的输入导致了失败？

**Thought Chain 记录：**
```python
{
  "iteration": 1,
  "thought": "告警显示 CPU 使用率 90%，我需要查看最近的部署记录",
  "action": {
    "tool": "kubernetes_get",
    "args": {"resource": "deployments", "namespace": "production"}
  },
  "observation": "最近 2 小时内有 3 次部署",
  "reasoning": "部署可能导致了 CPU 升高，需要进一步查看日志"
}
```

---

## 7.2 成本分析与优化

### Token 消耗分析

**成本构成：**
```text
总成本 = Input Token 成本 + Output Token 成本
```

**Input Token 来源：**
- System Prompt（固定）
- 用户输入
- 工具返回结果（累积）
- 历史对话（累积）

**Output Token 来源：**
- LLM 生成的内容
- 工具调用的参数
- 推理过程（Thought）

**DoD Agent 的成本分析：**
```text
单次告警诊断：
- System Prompt: 500 tokens
- 告警信息: 200 tokens
- 知识库检索结果: 2000 tokens
- 工具调用结果: 1500 tokens（累积）
- LLM 生成: 800 tokens
- 迭代 3 轮

Input: 500 + 200 + 2000 + 1500 * 3 = 7200 tokens
Output: 800 * 3 = 2400 tokens

成本: 7200 * 0.00003 + 2400 * 0.00006 = $0.36
```

每月 200 个告警，总成本约 $72/月。

### 优化策略 1：Prompt 压缩

**技巧 1：精简 System Prompt**

❌ **冗长版本（800 tokens）：**
```text
你是一个经验丰富的运维专家，拥有 10 年以上的大型电商系统运维经验。
你精通 Kubernetes、Prometheus、Grafana 等工具...
（详细描述 500 字）
```

✅ **精简版本（100 tokens）：**
```text
你是运维专家，分析告警并提供诊断建议。
```

节省 87.5% 的 System Prompt Token。

**技巧 2：上下文裁剪**

只保留最相关的信息：

```python
def build_context(alert, max_tokens=2000):
    # 1. 检索相关文档
    docs = retrieve_docs(alert, top_k=5)
    
    # 2. 裁剪到 token 限制
    context = []
    total_tokens = 0
    
    for doc in docs:
        doc_tokens = count_tokens(doc)
        if total_tokens + doc_tokens > max_tokens:
            break
        context.append(doc)
        total_tokens += doc_tokens
    
    return "\n\n".join(context)
```

**技巧 3：工具返回结果压缩**

```python
def prometheus_query(query: str, time_range: str):
    # 原始结果可能有 100+ 数据点
    raw_result = prom_api.query(query, time_range)
    
    # 压缩：只返回关键统计信息
    return {
        "avg": calculate_avg(raw_result),
        "max": calculate_max(raw_result),
        "current": raw_result[-1],
        "trend": calculate_trend(raw_result)  # "increasing" / "stable" / "decreasing"
    }
```

从 1000+ tokens 压缩到 50 tokens。

### 优化策略 2：缓存机制

**工具结果缓存：**
```python
class CachedTool:
    def __init__(self, tool: Tool, ttl: int = 300):
        self.tool = tool
        self.cache = {}
        self.ttl = ttl
    
    def execute(self, args):
        cache_key = self.generate_cache_key(args)
        
        # 检查缓存
        if cache_key in self.cache:
            cached = self.cache[cache_key]
            if time.time() - cached["timestamp"] < self.ttl:
                return cached["result"]
        
        # 执行工具
        result = self.tool.execute(args)
        
        # 写入缓存
        self.cache[cache_key] = {
            "result": result,
            "timestamp": time.time()
        }
        
        return result
```

**相似查询缓存：**

对于知识库检索，相似的查询可以返回缓存结果：

```python
def cached_retrieval(query: str, similarity_threshold: float = 0.95):
    # 计算查询的 embedding
    query_embedding = get_embedding(query)
    
    # 查找缓存中相似的查询
    for cached_query, cached_result in cache.items():
        cached_embedding = get_embedding(cached_query)
        similarity = cosine_similarity(query_embedding, cached_embedding)
        
        if similarity > similarity_threshold:
            return cached_result
    
    # 未命中缓存，执行实际检索
    result = vector_db.search(query)
    cache[query] = result
    return result
```

### 优化策略 3：模型分层

不是所有任务都需要最强的模型。

**模型选择策略：**

| 任务类型 | 推荐模型 | 成本 |
|---------|---------|------|
| **简单分类** | GPT-3.5 | $0.0015/1K tokens |
| **一般推理** | GPT-4 | $0.03/1K tokens |
| **复杂推理** | GPT-4 Turbo | $0.06/1K tokens |
| **代码生成** | Claude 3.5 Sonnet | $0.015/1K tokens |

**实现示例：**
```python
class AdaptiveAgent:
    def __init__(self):
        self.models = {
            "simple": GPT35(),
            "medium": GPT4(),
            "complex": GPT4Turbo()
        }
    
    def classify_complexity(self, task: str) -> str:
        # 简单分类：用便宜的模型
        complexity = self.models["simple"].generate(
            f"评估任务复杂度（simple/medium/complex）：{task}"
        )
        return complexity
    
    def run(self, task: str):
        # 根据复杂度选择模型
        complexity = self.classify_complexity(task)
        model = self.models[complexity]
        
        return model.generate(task)
```

### 优化策略 4：Early Stopping

当 Agent 陷入无效循环时及时停止。

**循环检测：**
```python
class LoopDetector:
    def __init__(self, max_repeats: int = 3):
        self.action_history = []
        self.max_repeats = max_repeats
    
    def is_looping(self, action: str) -> bool:
        self.action_history.append(action)
        
        # 检查最近的 N 次操作是否重复
        if len(self.action_history) >= self.max_repeats:
            recent = self.action_history[-self.max_repeats:]
            if len(set(recent)) == 1:  # 所有操作相同
                return True
        
        return False

# 使用
detector = LoopDetector(max_repeats=3)

for iteration in range(max_iterations):
    action = agent.decide_action()
    
    if detector.is_looping(action):
        print("检测到循环，提前停止")
        break
    
    execute_action(action)
```

---

## 7.3 可观测性实现方案

### 方案 1：结构化日志

**日志内容：**
```python
import logging
import json

class AgentLogger:
    def __init__(self, session_id: str):
        self.session_id = session_id
        self.logger = logging.getLogger("agent")
    
    def log_llm_call(self, prompt: str, response: str, tokens: dict):
        self.logger.info(json.dumps({
            "type": "llm_call",
            "session_id": self.session_id,
            "timestamp": time.time(),
            "prompt_length": len(prompt),
            "response_length": len(response),
            "tokens": tokens
        }))
    
    def log_tool_call(self, tool_name: str, args: dict, result: any):
        self.logger.info(json.dumps({
            "type": "tool_call",
            "session_id": self.session_id,
            "timestamp": time.time(),
            "tool": tool_name,
            "args": args,
            "result_type": type(result).__name__
        }))
    
    def log_error(self, error: Exception, context: dict):
        self.logger.error(json.dumps({
            "type": "error",
            "session_id": self.session_id,
            "timestamp": time.time(),
            "error": str(error),
            "context": context
        }))
```

### 方案 2：分布式追踪

使用 OpenTelemetry 实现分布式追踪：

```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

# 初始化 tracer
tracer = trace.get_tracer("agent")

class TracedAgent:
    def run(self, input: str):
        with tracer.start_as_current_span("agent_run") as span:
            span.set_attribute("input_length", len(input))
            
            # LLM 调用
            with tracer.start_as_current_span("llm_generate") as llm_span:
                response = self.llm.generate(input)
                llm_span.set_attribute("tokens", response.usage)
            
            # 工具调用
            with tracer.start_as_current_span("tool_execute") as tool_span:
                result = self.tool.execute(args)
                tool_span.set_attribute("tool_name", "prometheus_query")
            
            span.set_attribute("success", True)
            return result
```

### 方案 3：实时监控面板

**核心指标面板：**
```text
┌─────────────────────────────────────────────────────────────┐
│                   Agent Monitoring Dashboard                 │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  实时指标                                                    │
│  ├─ 当前活跃 Agent 数: 5                                     │
│  ├─ 总请求数（今日）: 156                                    │
│  ├─ 成功率: 92.3%                                           │
│  └─ 平均响应时间: 12.5s                                     │
│                                                             │
│  Token 使用                                                  │
│  ├─ Input Tokens（今日）: 1.2M                              │
│  ├─ Output Tokens（今日）: 350K                             │
│  └─ 估计成本（今日）: $67                                   │
│                                                             │
│  工具调用排行                                                │
│  ├─ prometheus_query: 89 次                                  │
│  ├─ loki_search: 67 次                                       │
│  ├─ kubernetes_get: 45 次                                    │
│  └─ confluence_search: 34 次                                 │
│                                                             │
│  失败分析                                                    │
│  ├─ tool_timeout: 8 次                                       │
│  ├─ llm_error: 3 次                                          │
│  └─ invalid_action: 5 次                                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 7.4 成本优化实战

### 案例：DoD Agent 成本优化历程

**V1：无优化版本**
```text
月告警量: 6000 次
平均 Token: 10000 tokens/次
月 Token 消耗: 60M tokens
月成本: 60M * 0.00003 = $1800
```

成本过高，不可持续。

**优化 1：Prompt 精简**
```text
System Prompt: 800 → 100 tokens (-87.5%)
检索结果: 3000 → 1000 tokens (top-3 instead of top-10)

平均 Token: 10000 → 6500 tokens (-35%)
月成本: $1800 → $1170 (-35%)
```

**优化 2：缓存相似查询**
```text
缓存命中率: 40%
缓存命中时跳过 LLM 调用

实际 LLM 调用: 6000 * 0.6 = 3600 次
月成本: $1170 → $702 (-40%)
```

**优化 3：模型分层**
```text
简单告警（50%）: 用 GPT-3.5 ($0.0015/1K tokens)
复杂告警（50%）: 用 GPT-4 ($0.03/1K tokens)

平均成本: (0.0015 * 0.5 + 0.03 * 0.5) * 6.5 = $0.10/次
月成本: $702 → $360 (-49%)
```

**最终优化：**
```text
V1 成本: $1800/月
V3 成本: $360/月
优化幅度: 80%

同时保持：
- 诊断准确率: 85%+
- 响应时间: < 30s
```

### 优化技巧汇总

| 优化方向 | 具体方法 | 节省比例 | 风险 |
|---------|---------|---------|------|
| **Prompt 精简** | 删除冗余描述 | 30-40% | 低 |
| **上下文裁剪** | 只保留最相关信息 | 20-30% | 中（可能丢失重要信息） |
| **结果缓存** | 相似查询返回缓存 | 30-50% | 低 |
| **模型分层** | 简单任务用便宜模型 | 40-60% | 中（需要准确分类） |
| **Early Stopping** | 检测循环提前终止 | 10-20% | 低 |
| **批量处理** | 合并多个请求 | 20-30% | 中（需要支持批量） |

---

## 7.5 代码熵管理

Agent 产出的代码会随时间积累技术债，**代码熵**会持续增长。

### 什么是代码熵？

```text
代码熵 = 架构不一致 + 重复代码 + 废弃代码 + 文档过时
```

单次任务看起来没问题，但 Agent 生成的代码在不同会话之间缺乏统一的架构愿景。

### 熵管理策略

**策略 1：持续审计**

定期运行审计 Agent：

```bash
# 每周运行一次
/schedule weekly "Audit codebase for architecture violations"
```

审计 Agent 检查：
- 依赖关系是否违反架构分层
- 新代码是否遵循设计模式
- 文档是否与代码同步

**策略 2：自动重构**

```python
def refactor_duplicated_code():
    """Find and refactor duplicated code blocks."""
    
    # 1. 检测重复代码
    duplicates = detect_duplicates(threshold=0.9)
    
    # 2. 提取为公共函数
    for dup in duplicates:
        common_func = extract_common_function(dup)
        replace_with_function_call(dup, common_func)
    
    # 3. 运行测试验证
    run_tests()
```

**策略 3：定期清理**

```bash
# 清理未使用的代码
claude -p "Find and remove unused code in src/"

# 清理废弃的依赖
claude -p "Remove unused dependencies from package.json"

# 更新过时的文档
claude -p "Update docs/ to match current implementation"
```

---

## 7.6 可观测性检查清单

```markdown
## Agent 可观测性检查清单

### 基础指标
- [ ] 记录请求总数、成功率、失败率
- [ ] 记录平均响应时间和 P99 延迟
- [ ] 记录 LLM 调用次数和 Token 消耗
- [ ] 记录工具调用次数和分布
- [ ] 记录成本（按天/周/月统计）

### 执行追踪
- [ ] 每个请求有唯一的 trace_id
- [ ] 记录完整的执行路径
- [ ] 记录每个步骤的耗时
- [ ] 支持跨服务追踪（如果是分布式）

### 行为分析
- [ ] 记录 Agent 的推理过程（Thought Chain）
- [ ] 记录工具选择的依据
- [ ] 分析失败的根因
- [ ] 识别常见的错误模式

### 可视化
- [ ] 实时监控面板
- [ ] Token 消耗趋势图
- [ ] 成功率趋势图
- [ ] 工具调用热力图

### 告警
- [ ] 成功率低于阈值告警
- [ ] 成本超预算告警
- [ ] 响应时间过长告警
- [ ] 工具调用失败告警
```

---

## 本章小结

### 核心要点回顾

**1. Agent 可观测性的三个层次**
- **基础指标**：请求数、成功率、Token 消耗、成本
- **执行追踪**：完整的执行路径和耗时分析
- **行为分析**：推理过程、决策依据、失败根因

**2. 成本优化五大策略**
- **Prompt 压缩**：精简 System Prompt，节省 30-40%
- **上下文裁剪**：只保留最相关信息，节省 20-30%
- **结果缓存**：相似查询返回缓存，节省 30-50%
- **模型分层**：简单任务用便宜模型，节省 40-60%
- **Early Stopping**：检测循环提前终止，节省 10-20%

**3. DoD Agent 成本优化实战**
- V1 成本：$1800/月
- 通过四轮优化降到 $360/月
- 优化幅度：80%
- 同时保持准确率 85%+ 和响应时间 < 30s

**4. 代码熵管理**
- 持续审计：定期检查架构违规
- 自动重构：提取重复代码
- 定期清理：删除废弃代码和依赖

### 关键洞察

> **可观测性不是可选项，而是必需品。没有可观测性的 Agent 系统，就像盲人开车。**

> **成本优化不是降低功能，而是提高效率。好的优化能在降低成本的同时提升质量。**

### 第二部分总结

第二部分（Agent 系统设计）的四章内容已完成：
- ✅ 第4章：Agent 架构设计与决策框架
- ✅ 第5章：工具系统与 MCP 协议
- ✅ 第6章：多 Agent 协作与工作流编排
- ✅ 第7章：可观测性与成本优化

**核心思想串联：**
1. **第4章** 建立了设计思维：何时需要 Agent，如何选择架构模式
2. **第5章** 深入工具系统：如何设计可靠的工具接口
3. **第6章** 探讨协作模式：如何让多个 Agent 高效协作
4. **第7章** 关注长期运营：如何观测和优化系统

### 下一章预告

第8章我们将进入第三部分"完整实战案例"，通过 DoD Agent（电商告警自动处理系统）的完整案例，展示从需求分析到上线部署的全过程。

---

## 参考资料

1. **OpenTelemetry for LLM Applications** - OpenTelemetry Docs
2. **LangSmith: LLM Application Observability** - LangChain
3. **Cost Optimization Strategies for LLM Applications** - Anthropic Engineering Blog
4. **Monitoring and Debugging LLM Applications** - OpenAI Cookbook
5. **Agent Metrics and KPIs** - AI Engineering Best Practices
