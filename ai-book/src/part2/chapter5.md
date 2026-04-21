# 第5章 工具系统与MCP协议

> "Tools are the hands and eyes of an Agent." （工具是Agent的手和眼）

## 引言

Agent 通过工具与外部世界交互。工具系统的设计质量，直接决定了 Agent 的能力边界和可靠性。

本章将深入探讨工具系统的设计原则、MCP（Model Context Protocol）协议的核心机制，以及如何构建可靠、可扩展的工具编排系统。

---

## 5.1 工具系统的核心价值

### 为什么需要工具系统？

LLM 本身只能生成文本，无法直接与外部系统交互。工具系统赋予 Agent **执行能力**：

```
┌─────────────────────────────────────────────────────────────┐
│                 Agent 能力层次                                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Layer 3: 执行层（Tool System）                              │
│  ├─ 查询数据库                                              │
│  ├─ 调用 API                                                 │
│  ├─ 执行命令                                                │
│  └─ 操作文件系统                                            │
│                                                             │
│  Layer 2: 推理层（LLM）                                      │
│  ├─ 理解需求                                                │
│  ├─ 规划步骤                                                │
│  ├─ 选择工具                                                │
│  └─ 生成参数                                                │
│                                                             │
│  Layer 1: 输入层（User Interface）                           │
│  ├─ 自然语言输入                                            │
│  ├─ 结构化输入                                              │
│  └─ 上下文信息                                              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 工具系统的三大作用

**1. 扩展能力边界**

LLM 只能基于训练数据回答问题，无法访问：
- 实时数据（股票价格、天气）
- 私有数据（公司内部文档、数据库）
- 动态信息（服务器状态、日志）

工具系统突破这些限制。

**2. 提供确定性保证**

LLM 的输出是概率性的，但某些操作需要确定性：
- 数学计算：LLM 可能算错，用 Calculator 工具保证准确
- 日期查询：LLM 的训练数据有截止日期，用 API 查询实时日期
- 数据查询：LLM 可能幻觉，用数据库查询保证真实性

**3. 实现闭环反馈**

```
用户需求 → Agent 推理 → 调用工具 → 观察结果 → 调整策略 → 再次调用工具
```

工具返回的结果成为 Agent 的"观察"（Observation），形成 ReACT 循环。

---

## 5.2 工具设计的五大原则

### 原则 1：单一职责

每个工具应该只做一件事，并且做好。

❌ **反例：万能工具**
```python
def universal_tool(action: str, **kwargs):
    if action == "search":
        return search(**kwargs)
    elif action == "calculate":
        return calculate(**kwargs)
    elif action == "query_db":
        return query_db(**kwargs)
    # ... 100 种操作
```

问题：
- LLM 难以理解工具能力
- 参数验证复杂
- 错误处理困难

✅ **正例：专用工具**
```python
def search_web(query: str, limit: int = 10):
    """Search the web for information.
    
    Args:
        query: The search query
        limit: Maximum number of results (default 10)
    
    Returns:
        List of search results with title, url, snippet
    """
    # ...

def calculate(expression: str):
    """Evaluate a mathematical expression.
    
    Args:
        expression: A valid math expression (e.g. "2+2", "sqrt(16)")
    
    Returns:
        The result of the calculation
    """
    # ...
```

优势：
- LLM 容易理解工具用途
- 参数清晰，不易出错
- 易于测试和维护

### 原则 2：明确的输入输出

工具的接口应该清晰、类型安全、有文档。

**好的工具描述：**
```python
{
    "name": "prometheus_query",
    "description": "Query Prometheus for metrics data",
    "parameters": {
        "type": "object",
        "properties": {
            "query": {
                "type": "string",
                "description": "PromQL query expression (e.g. 'rate(http_requests_total[5m])')"
            },
            "time_range": {
                "type": "string",
                "description": "Time range in format like '1h', '30m', '7d'. Default: '1h'",
                "default": "1h"
            }
        },
        "required": ["query"]
    },
    "returns": {
        "type": "object",
        "description": "Query result with timestamp and values"
    }
}
```

**示例输出：**
```json
{
    "status": "success",
    "data": {
        "metric": {"__name__": "http_requests_total"},
        "values": [
            [1704067200, "150.5"],
            [1704067260, "162.3"]
        ]
    }
}
```

### 原则 3：幂等性

相同的输入应该产生相同的结果（对于查询类操作）。

**幂等工具：**
- ✅ `get_user_by_id(123)` - 每次返回相同用户
- ✅ `search_documents(query)` - 相同查询返回相同结果（在数据不变的情况下）

**非幂等工具（需要谨慎设计）：**
- ⚠️ `create_order(items)` - 每次创建新订单
- ⚠️ `send_email(to, content)` - 每次发送新邮件
- ⚠️ `delete_file(path)` - 不可逆操作

对于非幂等工具，应该：
- 明确标记为"写操作"
- 需要额外的确认机制
- 记录完整的审计日志

### 原则 4：错误处理

工具应该返回结构化的错误信息，而不是抛出异常。

❌ **反例：抛出异常**
```python
def query_database(sql: str):
    result = db.execute(sql)  # 可能抛出异常
    return result
```

✅ **正例：结构化错误**
```python
def query_database(sql: str) -> ToolResult:
    try:
        result = db.execute(sql)
        return ToolResult(
            success=True,
            data=result,
            message="Query executed successfully"
        )
    except SQLSyntaxError as e:
        return ToolResult(
            success=False,
            error_code="SQL_SYNTAX_ERROR",
            message=f"Invalid SQL syntax: {str(e)}",
            details={"sql": sql}
        )
    except PermissionError as e:
        return ToolResult(
            success=False,
            error_code="PERMISSION_DENIED",
            message="Insufficient permissions to execute this query"
        )
```

LLM 可以根据错误信息调整策略：
- `SQL_SYNTAX_ERROR` → 修正 SQL 语句重试
- `PERMISSION_DENIED` → 切换到其他工具或通知用户

### 原则 5：可观测性

工具调用应该被完整记录，便于调试和优化。

**记录内容：**
```python
{
    "tool_name": "prometheus_query",
    "timestamp": "2026-04-21T10:30:00Z",
    "inputs": {
        "query": "rate(http_requests_total[5m])",
        "time_range": "1h"
    },
    "outputs": {
        "status": "success",
        "data_points": 12,
        "execution_time_ms": 145
    },
    "agent_session_id": "sess_abc123",
    "user_id": "user_456"
}
```

**监控指标：**
- 调用次数
- 成功率
- 平均延迟
- 错误分布

---

## 5.3 MCP：Model Context Protocol

### 什么是 MCP？

MCP（Model Context Protocol）是 Anthropic 推出的开放标准，用于统一 AI 应用与外部数据源和工具的连接方式。

**核心理念：**
```
AI 应用不应该为每个工具写专门的集成代码。
应该有一个统一的协议，让工具"即插即用"。
```

类比：USB 接口之于硬件设备。

### MCP 架构

```
┌─────────────────────────────────────────────────────────────┐
│                     MCP Architecture                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐                                           │
│  │ AI Application│                                          │
│  │  (Client)    │                                           │
│  └──────┬───────┘                                           │
│         │                                                    │
│         │ MCP Protocol                                       │
│         │                                                    │
│         ├────────► ┌──────────────┐                         │
│         │          │ MCP Server 1 │                         │
│         │          │  (Slack)     │                         │
│         │          └──────────────┘                         │
│         │                                                    │
│         ├────────► ┌──────────────┐                         │
│         │          │ MCP Server 2 │                         │
│         │          │  (GitHub)    │                         │
│         │          └──────────────┘                         │
│         │                                                    │
│         └────────► ┌──────────────┐                         │
│                    │ MCP Server 3 │                         │
│                    │  (Database)  │                         │
│                    └──────────────┘                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### MCP 的三大核心能力

**1. Resources（资源）**

只读的数据源，如文件、数据库记录、API 响应。

```json
{
  "resources": [
    {
      "uri": "slack://channel/general",
      "name": "Slack #general channel",
      "mimeType": "application/json",
      "description": "Messages from the #general channel"
    }
  ]
}
```

**2. Tools（工具）**

可执行的操作，如发送消息、创建文件、查询数据库。

```json
{
  "tools": [
    {
      "name": "send_slack_message",
      "description": "Send a message to a Slack channel",
      "inputSchema": {
        "type": "object",
        "properties": {
          "channel": {"type": "string"},
          "message": {"type": "string"}
        },
        "required": ["channel", "message"]
      }
    }
  ]
}
```

**3. Prompts（提示词模板）**

预定义的提示词模板，便于复用。

```json
{
  "prompts": [
    {
      "name": "analyze_incident",
      "description": "Analyze a production incident",
      "arguments": [
        {
          "name": "incident_id",
          "description": "The incident ID to analyze"
        }
      ]
    }
  ]
}
```

### 添加 MCP Server

**命令行添加：**
```bash
# 添加 Slack MCP
claude mcp add slack -- npx -y @modelcontextprotocol/server-slack

# 添加 GitHub MCP
claude mcp add github -- npx -y @modelcontextprotocol/server-github

# 添加数据库 MCP
claude mcp add postgres -- npx -y @modelcontextprotocol/server-postgres
```

**配置文件：**
```json
// .mcp.json
{
  "mcpServers": {
    "slack": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-slack"],
      "env": {
        "SLACK_TOKEN": "${SLACK_TOKEN}"
      }
    },
    "postgres": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-postgres",
        "${DATABASE_URL}"
      ]
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  }
}
```

### 实用 MCP Server 推荐

| MCP Server | 能力 | 适用场景 |
|------------|------|----------|
| **Slack** | 搜索/发送消息 | 团队协作，通知机制 |
| **GitHub** | 操作仓库/Issue/PR | 项目管理，代码审查 |
| **Postgres** | 查询数据库 | 数据分析，报表生成 |
| **Figma** | 读取设计稿 | 设计转代码 |
| **Sentry** | 获取错误日志 | 错误诊断，性能分析 |
| **Confluence** | 搜索文档 | 知识库检索 |

---

## 5.4 工具编排最佳实践

### 实践 1：优先使用标准工具

LLM 在标准工具上有海量训练数据，执行可靠性远高于自定义工具。

**标准工具：**
- ✅ `git`、`docker`、`kubectl`
- ✅ `npm`、`pip`、`cargo`
- ✅ `curl`、`jq`、`grep`

**自定义工具：**
- ⚠️ 自研的 CLI 工具
- ⚠️ 内部 API

如果必须使用自定义工具：
- 提供详细的文档和示例
- 保持接口简单一致
- 模仿标准工具的风格

### 实践 2：工具分层设计

将工具按抽象层次分层：

```
┌─────────────────────────────────────────────────────────────┐
│                    工具层次结构                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Layer 3: 业务层工具（High-Level）                           │
│  ├─ diagnose_alert: 诊断告警                                │
│  ├─ create_incident: 创建事故单                             │
│  └─ send_notification: 发送通知                              │
│                                                             │
│  Layer 2: 集成层工具（Mid-Level）                            │
│  ├─ prometheus_query: 查询 Prometheus                       │
│  ├─ loki_search: 搜索日志                                    │
│  └─ kubernetes_get: 查询 K8s 资源                           │
│                                                             │
│  Layer 1: 基础层工具（Low-Level）                            │
│  ├─ http_request: HTTP 请求                                 │
│  ├─ shell_exec: 执行命令                                    │
│  └─ file_read: 读取文件                                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**设计原则：**
- 高层工具基于低层工具构建
- Agent 优先调用高层工具（封装复杂度）
- 低层工具保持通用性

### 实践 3：最小权限原则

只给 Agent 完成任务所需的最少权限。

**风险分级：**
```
┌─────────────────────────────────────────────────────────────┐
│                    工具风险分级                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  低风险（自动放行）                                          │
│  ├─ 读取文件                                                │
│  ├─ 查询数据库                                              │
│  └─ 搜索日志                                                │
│                                                             │
│  中风险（需要确认）                                          │
│  ├─ 重启服务                                                │
│  ├─ 修改配置                                                │
│  └─ 执行脚本                                                │
│                                                             │
│  高风险（人工审批）                                          │
│  ├─ 删除数据                                                │
│  ├─ 数据库变更                                              │
│  └─ 生产环境部署                                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**实现机制：**
```go
type Tool struct {
    Name        string
    Description string
    RiskLevel   RiskLevel  // Low / Medium / High
    Execute     func(args map[string]interface{}) (interface{}, error)
}

func (a *Agent) executeTool(tool Tool, args map[string]interface{}) (interface{}, error) {
    // 风险评估
    if tool.RiskLevel == RiskHigh {
        // 需要人工审批
        approved := a.requestApproval(tool, args)
        if !approved {
            return nil, errors.New("execution not approved")
        }
    }
    
    // 执行前记录审计日志
    a.auditLog(tool.Name, args)
    
    // 执行工具
    result, err := tool.Execute(args)
    
    // 执行后记录结果
    a.auditLog(tool.Name, result, err)
    
    return result, err
}
```

### 实践 4：工具文档即代码

将工具描述写成代码，确保文档和实现同步。

```python
from typing import TypedDict, Literal

class PrometheusQueryArgs(TypedDict):
    """Arguments for prometheus_query tool."""
    query: str  # PromQL query expression
    time_range: str  # Time range like '1h', '30m', default '1h'

def prometheus_query(args: PrometheusQueryArgs) -> dict:
    """Query Prometheus for metrics data.
    
    This tool allows you to query Prometheus using PromQL.
    
    Examples:
        - Query CPU usage: {"query": "rate(cpu_usage[5m])", "time_range": "1h"}
        - Query memory: {"query": "memory_usage", "time_range": "30m"}
    
    Returns:
        A dict containing:
        - status: 'success' or 'error'
        - data: List of metric data points
        - message: Human-readable message
    """
    # Implementation
    pass
```

类型注解和文档字符串可以自动生成 OpenAI Function Calling 的 Schema。

### 实践 5：工具链组合

将常用的工具调用序列封装成工具链。

**示例：告警诊断工具链**
```python
def diagnose_alert_chain(alert_id: str):
    """Complete alert diagnosis workflow.
    
    This tool chains multiple steps:
    1. Get alert details
    2. Query related metrics (Prometheus)
    3. Search relevant logs (Loki)
    4. Check service status (Kubernetes)
    5. Find similar historical cases
    6. Generate diagnosis report
    """
    # Step 1: Get alert
    alert = get_alert(alert_id)
    
    # Step 2: Query metrics
    metrics = prometheus_query({
        "query": f"rate({alert.metric}[5m])",
        "time_range": "1h"
    })
    
    # Step 3: Search logs
    logs = loki_search({
        "query": f"{{service=\"{alert.service}\"}}",
        "time_range": "30m"
    })
    
    # Step 4: Check K8s status
    k8s_status = kubernetes_get({
        "resource": "pods",
        "namespace": alert.namespace
    })
    
    # Step 5: Find similar cases
    similar_cases = search_similar_cases(alert)
    
    # Step 6: Generate report
    return generate_diagnosis_report({
        "alert": alert,
        "metrics": metrics,
        "logs": logs,
        "k8s_status": k8s_status,
        "similar_cases": similar_cases
    })
```

**优势：**
- 减少 Agent 的推理步骤
- 确保关键步骤不被遗漏
- 提高执行效率

---

## 5.5 工具设计检查清单

```markdown
## 工具设计检查清单

### 接口设计
- [ ] 工具名称清晰，符合命名规范
- [ ] 功能描述简洁明确
- [ ] 参数类型明确，有默认值
- [ ] 返回值结构化，有文档说明
- [ ] 包含使用示例

### 可靠性
- [ ] 实现了错误处理
- [ ] 返回结构化的错误信息
- [ ] 查询类操作是幂等的
- [ ] 写操作有确认机制
- [ ] 有超时控制

### 安全性
- [ ] 实现了权限控制
- [ ] 敏感操作需要审批
- [ ] 记录完整的审计日志
- [ ] 输入参数有验证
- [ ] 防止注入攻击

### 可观测性
- [ ] 记录工具调用日志
- [ ] 暴露关键指标（调用次数、延迟、成功率）
- [ ] 可以追踪到具体的 Agent 会话
- [ ] 错误有详细的上下文信息

### 文档
- [ ] 有清晰的文档和示例
- [ ] 文档和实现同步
- [ ] 有错误码说明
- [ ] 有性能特征说明
```

---

## 本章小结

### 核心要点回顾

**1. 工具系统的核心价值**
- 扩展 Agent 能力边界
- 提供确定性保证
- 实现闭环反馈

**2. 工具设计五大原则**
- 单一职责：每个工具只做一件事
- 明确接口：清晰的输入输出
- 幂等性：相同输入产生相同结果
- 错误处理：结构化错误信息
- 可观测性：完整记录调用日志

**3. MCP 协议**
- 统一的工具连接标准
- 三大核心能力：Resources、Tools、Prompts
- 即插即用，易于扩展

**4. 工具编排最佳实践**
- 优先使用标准工具
- 工具分层设计（业务层/集成层/基础层）
- 最小权限原则（风险分级）
- 工具文档即代码
- 工具链组合（封装常用序列）

### 关键洞察

> **工具不仅决定了 Agent 的能力边界，也决定了它的可靠性边界。好的工具设计是 Agent 系统成功的基础。**

### 下一章预告

第6章我们将探讨多 Agent 协作的设计模式，包括如何设计 Agent 团队、如何协调多个 Agent、以及如何避免常见的协作陷阱。

---

## 参考资料

1. **Model Context Protocol Specification** - Anthropic, 2025
2. **OpenAI Function Calling Guide** - https://platform.openai.com/docs/guides/function-calling
3. **LangChain Tools Documentation** - https://docs.langchain.com/docs/components/tools
4. **MCP Server Examples** - https://github.com/modelcontextprotocol/servers
5. **Tool Design Best Practices** - Anthropic Engineering Blog
