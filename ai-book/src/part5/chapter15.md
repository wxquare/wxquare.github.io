# 第15章 Tool Calling与MCP工程深化

> Agent 的智能上限来自模型，可靠性下限来自工具。

## 引言

很多 Agent demo 看起来很强，是因为模型会说；很多生产 Agent 不稳定，是因为工具设计太粗糙。工具调用不是简单地把一个 API 暴露给模型，而是要把业务能力包装成清晰、可验证、可审计的动作。

本章从面试和工程落地角度，深入讨论 Tool Calling 与 MCP 的设计。

---

## 15.1 好工具的五个特征

### 1. 语义明确

工具名称应该描述业务动作，而不是技术实现。

不推荐：

```text
call_api
execute_sql
run_command
```

推荐：

```text
get_order_status
search_error_logs
create_support_ticket
query_prometheus_metric
```

### 2. 输入结构化

参数应该有类型、枚举、默认值和约束。

```json
{
  "name": "search_error_logs",
  "description": "Search application error logs for a service in a time range.",
  "input_schema": {
    "type": "object",
    "properties": {
      "service": {
        "type": "string",
        "description": "Service name, such as order-service"
      },
      "time_range": {
        "type": "string",
        "enum": ["15m", "30m", "1h", "6h"]
      },
      "keyword": {
        "type": "string",
        "description": "Optional keyword filter"
      }
    },
    "required": ["service", "time_range"]
  }
}
```

### 3. 输出可解释

工具不要只返回原始数据，要返回 Agent 能理解的结构。

```json
{
  "status": "success",
  "summary": "Found 128 timeout errors in the last 30 minutes.",
  "items": [
    {
      "timestamp": "2026-04-26T10:01:00Z",
      "level": "error",
      "message": "payment provider timeout",
      "trace_id": "abc123"
    }
  ],
  "next_actions": [
    "Check payment provider latency",
    "Compare with recent deployment"
  ]
}
```

### 4. 错误可恢复

错误信息要帮助 Agent 判断下一步。

```json
{
  "status": "error",
  "error_code": "TIME_RANGE_TOO_LARGE",
  "message": "time_range cannot exceed 6h",
  "retryable": true,
  "suggested_fix": "Use one of: 15m, 30m, 1h, 6h"
}
```

### 5. 行为可审计

每个工具调用至少记录：

- trace_id；
- user_id；
- tool_name；
- args hash；
- risk_level；
- approval_id；
- latency；
- result status。

---

## 15.2 工具设计反模式

### 反模式 1：万能工具

```json
{"name": "database_query", "description": "Run SQL query"}
```

问题：权限过大、不可控、难审计。

改法：封装成具体只读业务查询。

### 反模式 2：参数太自由

```json
{"time_range": "any string"}
```

问题：模型容易传入不可执行格式。

改法：使用 enum 或严格格式。

### 反模式 3：错误只返回 failed

```json
{"status": "failed"}
```

问题：Agent 不知道是重试、换工具还是转人工。

改法：返回 error_code、retryable、suggested_fix。

### 反模式 4：把权限交给模型判断

错误做法：

```text
请你判断这个用户能不能执行退款。
```

正确做法：权限由后端系统判断，模型只负责解释原因和生成请求。

---

## 15.3 Tool Registry设计

生产系统中应该有统一工具注册表。

```python
from dataclasses import dataclass
from typing import Callable, Literal

RiskLevel = Literal["low", "medium", "high", "critical"]

@dataclass
class ToolSpec:
    name: str
    description: str
    input_schema: dict
    risk_level: RiskLevel
    timeout_ms: int
    handler: Callable

class ToolRegistry:
    def __init__(self):
        self.tools = {}

    def register(self, spec: ToolSpec):
        if spec.name in self.tools:
            raise ValueError(f"duplicate tool: {spec.name}")
        self.tools[spec.name] = spec

    def get_allowed_tools(self, user, task_context):
        return [
            tool for tool in self.tools.values()
            if has_permission(user, tool, task_context)
        ]
```

关键点：

- 工具按用户和任务上下文动态过滤；
- 高风险工具不一定暴露给模型；
- 工具 schema 和实际 handler 必须同步；
- 工具版本变化要纳入 eval 回归。

---

## 15.4 MCP的工程价值

MCP 的价值在于标准化“模型应用如何连接外部工具和数据源”。

传统集成问题：

```text
N 个 Agent 应用 × M 个业务系统 = N × M 套集成
```

MCP 目标：

```text
Agent Host ↔ MCP Client ↔ MCP Server ↔ Business System
```

适合用 MCP 的场景：

- 多个 Agent 需要复用同一套工具；
- 企业内部系统很多，集成成本高；
- 工具需要独立部署和版本管理；
- 希望模型供应商无关；
- 需要统一鉴权、审计和工具发现。

---

## 15.5 MCP Server设计建议

一个 MCP server 不应该只是 API 的薄包装，而应该暴露稳定的业务能力。

### 工具粒度

工具粒度太细，Agent 需要多步拼装，容易出错；工具粒度太粗，又会失去灵活性。

建议：

- 查询类工具可以细一些；
- 写操作工具要更业务化；
- 高风险流程封装成带审批的 workflow；
- 不暴露通用 shell、通用 SQL、通用 HTTP 请求。

### 版本管理

工具 schema 变化会影响 Agent 行为。建议：

```text
get_order_status_v1
get_order_status_v2
```

或者在工具元数据中记录版本：

```json
{
  "name": "get_order_status",
  "version": "2.1.0",
  "deprecated": false
}
```

### 权限模型

MCP server 侧必须做权限校验，不能只依赖 Agent host。

```text
用户身份 → Agent Host → MCP Client → MCP Server → 权限校验 → 业务系统
```

---

## 15.6 工具调用的可靠性设计

### 超时

每个工具必须有超时。

```python
async def call_tool_with_timeout(tool, args):
    return await asyncio.wait_for(
        tool.handler(args),
        timeout=tool.timeout_ms / 1000
    )
```

### 重试

只对可重试错误重试。

```text
可重试：网络超时、临时限流、下游 5xx
不可重试：权限失败、参数错误、业务状态不允许
```

### 幂等

写操作必须有 idempotency key。

```json
{
  "tool": "create_support_ticket",
  "args": {
    "title": "Payment failure spike",
    "idempotency_key": "alert_123_create_ticket"
  }
}
```

### 降级

工具不可用时，Agent 应该能降级：

- 明确告诉用户哪些数据不可用；
- 使用已有上下文给出低置信度建议；
- 转人工；
- 记录失败进入 eval dataset。

---

## 15.7 面试回答模板

问题：你如何设计一个工具调用系统？

```text
我会从工具 schema、权限、安全、可靠性和观测五个方面设计。

首先，工具必须是业务语义明确的，比如 get_order_status，而不是 execute_sql。
其次，输入输出要结构化，参数有类型、枚举和默认值，错误返回 error_code、retryable 和 suggested_fix。
第三，工具按风险分级，只读工具可以自动执行，高风险写操作需要人工审批，critical 操作不允许 Agent 执行。
第四，所有工具调用都要有超时、重试、幂等和审计日志。
第五，工具调用进入 trace，用于 debug 和 eval。

如果企业有多个 Agent 或多个系统，我会用 MCP 把工具标准化暴露，但权限校验必须保留在 MCP server 和业务系统侧。
```

---

## 15.8 检查清单

```markdown
## Tool Calling / MCP 检查清单

### Schema
- [ ] 工具名称是业务语义
- [ ] 参数类型明确
- [ ] 枚举和默认值清晰
- [ ] 输出结构化
- [ ] 错误可恢复

### 安全
- [ ] 工具有风险等级
- [ ] 工具按用户权限过滤
- [ ] 高风险工具需要审批
- [ ] 不暴露通用 shell / SQL
- [ ] MCP server 做服务端权限校验

### 可靠性
- [ ] 超时
- [ ] 重试
- [ ] 幂等
- [ ] 降级
- [ ] 审计日志

### 评估
- [ ] 评估 tool selection accuracy
- [ ] 评估 argument accuracy
- [ ] 评估 tool error recovery
- [ ] 工具变更后跑回归测试
```

---

## 本章小结

工具调用系统决定了 Agent 能做什么，也决定了它会怎么失败。

好的工具系统有三个特征：

1. 能力明确；
2. 权限可控；
3. 失败可恢复。

下一章我们把前面学到的内容组合起来，从零设计一个可观测 Mini Agent。
