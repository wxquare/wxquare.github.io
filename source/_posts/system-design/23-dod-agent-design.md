---
title: DoD Agent 设计方案：电商告警自动处理系统
date: 2026-03-09
categories:
- 系统设计
tags:
- AI Agent
- DevOps
- 告警处理
- 系统设计
toc: true
---

<!-- toc -->

## 一、项目概述

### 1.1 背景与目标

电商公司日常运维面临大量告警处理工作，包括基础设施告警、应用告警和业务告警。当前痛点：

| 痛点 | 影响 |
|:---|:---|
| 告警量大（50-200条/天） | 值班人员疲劳，响应延迟 |
| 重复性问题多 | 80% 问题有标准处理流程 |
| 知识分散 | Confluence 文档难以快速定位 |
| 跨部门协作 | 告警升级和分发效率低 |

**DoD Agent（DevOps on Duty Agent）** 目标：

```
┌─────────────────────────────────────────────────────────────┐
│                    DoD Agent 目标                           │
├─────────────────────────────────────────────────────────────┤
│  🎯 自动诊断     - 80% 告警自动分析根因                      │
│  📚 智能问答     - 基于 Confluence 知识库回答咨询            │
│  🔄 标准化处理   - 常见问题自动生成处理建议                   │
│  📊 告警聚合     - 关联告警智能聚合，减少噪音                 │
│  🚀 可扩展       - 支持扩展到其他部门（客服/安全/DBA）        │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 设计原则

1. **只读诊断优先**：第一阶段只做诊断和建议，不执行危险操作
2. **人机协作**：Agent 辅助决策，关键操作人工确认
3. **渐进增强**：从简单场景开始，逐步扩展能力
4. **可观测性**：完整的日志、指标和追踪

### 1.3 技术选型

| 组件 | 选型 | 理由 |
|:---|:---|:---|
| **LLM** | OpenAI GPT-4 | 推理能力强，工具调用成熟 |
| **告警源** | Prometheus + Alertmanager | 已有系统，Webhook 集成 |
| **知识库** | Confluence + RAG | 利用现有文档 |
| **交互渠道** | Slack | 团队主要沟通工具 |
| **部署平台** | Kubernetes | 已有基础设施 |
| **向量数据库** | Chroma / Milvus | 本地部署，数据安全 |

---

## 二、系统架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          DoD Agent System                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                      Input Layer (输入层)                        │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐    │   │
│  │  │Alertmanager│  │  Grafana  │  │   Slack   │  │  Web API  │    │   │
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
│  │                    Agent Core (核心引擎)                         │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │   │
│  │  │   Router    │  │ Agent Loop  │  │   Memory    │              │   │
│  │  │  (意图路由)  │─→│  (ReAct)   │←→│  (上下文)   │              │   │
│  │  └─────────────┘  └──────┬──────┘  └─────────────┘              │   │
│  └──────────────────────────┼──────────────────────────────────────┘   │
│                             │                                          │
│           ┌─────────────────┼─────────────────┐                       │
│           ▼                 ▼                 ▼                       │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐                 │
│  │    Tools    │   │     RAG     │   │   Output    │                 │
│  │  (工具系统)  │   │  (知识检索)  │   │  (输出层)   │                 │
│  └──────┬──────┘   └──────┬──────┘   └──────┬──────┘                 │
│         │                 │                 │                         │
│         ▼                 ▼                 ▼                         │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐                 │
│  │• Prometheus │   │• Confluence │   │   • Slack   │                 │
│  │• Kubernetes │   │• Runbook    │   │   • Email   │                 │
│  │• Grafana    │   │• 历史案例   │   │   • Ticket  │                 │
│  │• Log System │   │             │   │             │                 │
│  └─────────────┘   └─────────────┘   └─────────────┘                 │
│                                                                       │
└───────────────────────────────────────────────────────────────────────┘
```

### 2.2 数据流设计

```
告警触发 ──→ Alertmanager Webhook ──→ Gateway ──→ Alert Parser
                                                      │
                                                      ▼
                                              ┌───────────────┐
                                              │  Alert Queue  │
                                              │   (Redis)     │
                                              └───────┬───────┘
                                                      │
                    ┌─────────────────────────────────┼─────────────────────────┐
                    │                                 │                         │
                    ▼                                 ▼                         ▼
            ┌───────────────┐               ┌───────────────┐           ┌───────────────┐
            │ Alert Dedup   │               │ Alert Enrich  │           │Alert Correlate│
            │  (告警去重)    │               │  (告警富化)    │           │  (告警关联)   │
            └───────┬───────┘               └───────┬───────┘           └───────┬───────┘
                    │                               │                           │
                    └───────────────────────────────┼───────────────────────────┘
                                                    │
                                                    ▼
                                            ┌───────────────┐
                                            │  Agent Core   │
                                            │  (诊断分析)    │
                                            └───────┬───────┘
                                                    │
                                    ┌───────────────┴───────────────┐
                                    ▼                               ▼
                            ┌───────────────┐               ┌───────────────┐
                            │ Auto Resolve  │               │ Human Escalate│
                            │  (自动建议)    │               │  (人工升级)    │
                            └───────────────┘               └───────────────┘
```

---

## 三、核心模块设计

### 3.1 Gateway 模块

负责统一接入和消息标准化。

```python
from dataclasses import dataclass
from enum import Enum
from typing import Optional, Dict, List
from datetime import datetime

class InputSource(Enum):
    ALERTMANAGER = "alertmanager"
    GRAFANA = "grafana"
    SLACK = "slack"
    API = "api"

class AlertSeverity(Enum):
    CRITICAL = "critical"
    WARNING = "warning"
    INFO = "info"

@dataclass
class StandardAlert:
    """统一告警格式"""
    id: str                          # 唯一标识
    source: InputSource              # 来源
    severity: AlertSeverity          # 严重级别
    title: str                       # 告警标题
    description: str                 # 详细描述
    labels: Dict[str, str]           # 标签（service, env, pod等）
    annotations: Dict[str, str]      # 注解（runbook_url等）
    starts_at: datetime              # 开始时间
    fingerprint: str                 # 指纹（用于去重）
    raw_data: Dict                   # 原始数据

class AlertmanagerAdapter:
    """Alertmanager Webhook 适配器"""
    
    def parse(self, payload: Dict) -> List[StandardAlert]:
        alerts = []
        for alert in payload.get("alerts", []):
            alerts.append(StandardAlert(
                id=self._generate_id(alert),
                source=InputSource.ALERTMANAGER,
                severity=self._map_severity(alert["labels"].get("severity", "warning")),
                title=alert["labels"].get("alertname", "Unknown"),
                description=alert["annotations"].get("description", ""),
                labels=alert["labels"],
                annotations=alert["annotations"],
                starts_at=self._parse_time(alert["startsAt"]),
                fingerprint=alert["fingerprint"],
                raw_data=alert
            ))
        return alerts
    
    def _map_severity(self, severity: str) -> AlertSeverity:
        mapping = {
            "critical": AlertSeverity.CRITICAL,
            "warning": AlertSeverity.WARNING,
            "info": AlertSeverity.INFO
        }
        return mapping.get(severity.lower(), AlertSeverity.WARNING)
```

### 3.2 Router 模块（意图识别）

根据输入类型路由到不同处理流程。

```python
class IntentType(Enum):
    ALERT_DIAGNOSIS = "alert_diagnosis"      # 告警诊断
    KNOWLEDGE_QUERY = "knowledge_query"      # 知识查询
    HISTORY_LOOKUP = "history_lookup"        # 历史案例
    STATUS_CHECK = "status_check"            # 状态检查
    ESCALATION = "escalation"                # 升级处理

class IntentRouter:
    """意图路由器"""
    
    def __init__(self, llm):
        self.llm = llm
        self.intent_prompt = """
你是一个运维助手的意图识别模块。根据用户输入，判断意图类型。

意图类型：
1. alert_diagnosis - 告警诊断：用户询问某个告警的原因、影响、处理方法
2. knowledge_query - 知识查询：询问某个系统/服务的工作原理、配置方法
3. history_lookup - 历史案例：查找类似问题的历史处理记录
4. status_check - 状态检查：查询当前系统/服务状态
5. escalation - 升级处理：需要人工介入或升级

用户输入：{input}

请返回 JSON 格式：
{{"intent": "意图类型", "confidence": 0.0-1.0, "entities": {{"service": "", "alert_name": ""}}}}
"""
    
    def route(self, user_input: str, context: Dict = None) -> IntentType:
        # 如果是 Alertmanager Webhook，直接路由到告警诊断
        if context and context.get("source") == InputSource.ALERTMANAGER:
            return IntentType.ALERT_DIAGNOSIS
        
        # 使用 LLM 进行意图识别
        response = self.llm.generate(
            self.intent_prompt.format(input=user_input)
        )
        intent_data = self._parse_response(response)
        return IntentType(intent_data["intent"])
```

### 3.3 Agent Core（核心引擎）

基于 ReAct 模式的诊断引擎。

```python
class DoDAgent:
    """DoD Agent 核心引擎"""
    
    def __init__(self, llm, tools: ToolRegistry, rag: RAGSystem, memory: Memory):
        self.llm = llm
        self.tools = tools
        self.rag = rag
        self.memory = memory
        self.max_iterations = 8
    
    async def diagnose_alert(self, alert: StandardAlert) -> DiagnosisResult:
        """告警诊断主流程"""
        
        # 1. 构建初始上下文
        context = self._build_alert_context(alert)
        
        # 2. 检索相关知识
        knowledge = await self.rag.retrieve(
            query=f"{alert.title} {alert.description}",
            filters={"service": alert.labels.get("service")}
        )
        context += f"\n\n相关知识文档：\n{knowledge}"
        
        # 3. 检索历史案例
        history = await self.memory.search_similar_alerts(alert)
        if history:
            context += f"\n\n历史相似案例：\n{self._format_history(history)}"
        
        # 4. ReAct 诊断循环
        diagnosis_steps = []
        for i in range(self.max_iterations):
            response = await self.llm.generate(
                self._build_diagnosis_prompt(context, diagnosis_steps)
            )
            
            action = self._parse_action(response)
            
            if action.type == "final_diagnosis":
                return DiagnosisResult(
                    alert_id=alert.id,
                    root_cause=action.root_cause,
                    impact=action.impact,
                    suggested_actions=action.suggested_actions,
                    confidence=action.confidence,
                    steps=diagnosis_steps,
                    references=action.references
                )
            
            if action.type == "tool_call":
                result = await self.tools.execute(action.tool, **action.args)
                diagnosis_steps.append({
                    "thought": action.thought,
                    "tool": action.tool,
                    "args": action.args,
                    "result": result
                })
                context += f"\n\nStep {i+1}:\nThought: {action.thought}\nAction: {action.tool}\nResult: {result}"
        
        # 超过最大迭代，返回部分诊断结果
        return self._build_partial_result(alert, diagnosis_steps)
    
    def _build_diagnosis_prompt(self, context: str, steps: List) -> str:
        return f"""
你是一个专业的电商系统运维诊断专家。请根据告警信息和上下文，诊断问题根因。

## 告警上下文
{context}

## 已执行的诊断步骤
{self._format_steps(steps)}

## 可用工具
{self.tools.get_tools_prompt()}

## 诊断要求
1. 首先分析告警的直接原因
2. 使用工具收集更多信息（日志、指标、K8s状态等）
3. 结合知识库和历史案例分析
4. 给出根因分析和处理建议

请使用以下格式回复：
Thought: 你的分析思路
Action: 工具名称（或 "final_diagnosis"）
Action Input: {{"param": "value"}}

如果已完成诊断，使用：
Action: final_diagnosis
Action Input: {{
    "root_cause": "根因分析",
    "impact": "影响范围",
    "suggested_actions": ["建议1", "建议2"],
    "confidence": 0.85,
    "references": ["参考文档链接"]
}}
"""
```

---

## 四、工具系统设计

### 4.1 工具清单

| 工具 | 功能 | 权限级别 | 说明 |
|:---|:---|:---|:---|
| `prometheus_query` | 查询 Prometheus 指标 | 只读 | 查询 CPU/内存/请求量等 |
| `kubernetes_get` | 查询 K8s 资源状态 | 只读 | Pod/Deployment/Service 状态 |
| `log_search` | 搜索日志 | 只读 | 搜索 ES/Loki 日志 |
| `grafana_snapshot` | 获取 Grafana 面板截图 | 只读 | 生成监控截图 |
| `confluence_search` | 搜索 Confluence 文档 | 只读 | 搜索 Runbook 和文档 |
| `alert_history` | 查询历史告警 | 只读 | 查询相似告警处理记录 |
| `service_topology` | 查询服务依赖 | 只读 | 获取上下游依赖关系 |
| `slack_notify` | 发送 Slack 消息 | 写入 | 通知和升级 |

### 4.2 工具实现示例

```python
from typing import Dict, Any
import httpx

class PrometheusQueryTool(Tool):
    """Prometheus 查询工具"""
    
    name = "prometheus_query"
    description = "查询 Prometheus 监控指标，支持 PromQL"
    parameters = {
        "type": "object",
        "properties": {
            "query": {
                "type": "string",
                "description": "PromQL 查询语句"
            },
            "time_range": {
                "type": "string",
                "description": "时间范围，如 '5m', '1h', '24h'",
                "default": "15m"
            }
        },
        "required": ["query"]
    }
    
    def __init__(self, prometheus_url: str):
        self.prometheus_url = prometheus_url
        self.client = httpx.AsyncClient()
    
    async def execute(self, query: str, time_range: str = "15m") -> str:
        """执行 PromQL 查询"""
        try:
            # 范围查询
            response = await self.client.get(
                f"{self.prometheus_url}/api/v1/query_range",
                params={
                    "query": query,
                    "start": f"now-{time_range}",
                    "end": "now",
                    "step": "1m"
                }
            )
            data = response.json()
            
            if data["status"] == "success":
                return self._format_result(data["data"]["result"])
            else:
                return f"查询失败: {data.get('error', 'Unknown error')}"
        except Exception as e:
            return f"Prometheus 查询异常: {str(e)}"
    
    def _format_result(self, results: list) -> str:
        """格式化查询结果"""
        if not results:
            return "无数据"
        
        formatted = []
        for result in results[:5]:  # 限制返回数量
            metric = result["metric"]
            values = result["values"]
            
            # 获取最新值和趋势
            latest = float(values[-1][1]) if values else 0
            avg = sum(float(v[1]) for v in values) / len(values) if values else 0
            
            formatted.append(
                f"指标: {metric}\n"
                f"  最新值: {latest:.2f}\n"
                f"  平均值: {avg:.2f}\n"
                f"  数据点数: {len(values)}"
            )
        
        return "\n\n".join(formatted)


class KubernetesTool(Tool):
    """Kubernetes 查询工具"""
    
    name = "kubernetes_get"
    description = "查询 Kubernetes 资源状态，包括 Pod、Deployment、Service 等"
    parameters = {
        "type": "object",
        "properties": {
            "resource_type": {
                "type": "string",
                "enum": ["pod", "deployment", "service", "node", "event"],
                "description": "资源类型"
            },
            "namespace": {
                "type": "string",
                "description": "命名空间",
                "default": "default"
            },
            "name": {
                "type": "string",
                "description": "资源名称（可选，支持前缀匹配）"
            },
            "labels": {
                "type": "string",
                "description": "标签选择器，如 'app=order-service'"
            }
        },
        "required": ["resource_type"]
    }
    
    async def execute(
        self, 
        resource_type: str, 
        namespace: str = "default",
        name: str = None,
        labels: str = None
    ) -> str:
        """查询 K8s 资源"""
        from kubernetes import client, config
        
        try:
            config.load_incluster_config()  # 集群内部署
        except:
            config.load_kube_config()  # 本地开发
        
        v1 = client.CoreV1Api()
        apps_v1 = client.AppsV1Api()
        
        if resource_type == "pod":
            return await self._get_pods(v1, namespace, name, labels)
        elif resource_type == "deployment":
            return await self._get_deployments(apps_v1, namespace, name, labels)
        elif resource_type == "event":
            return await self._get_events(v1, namespace, name)
        else:
            return f"不支持的资源类型: {resource_type}"
    
    async def _get_pods(self, v1, namespace, name, labels) -> str:
        """获取 Pod 状态"""
        pods = v1.list_namespaced_pod(
            namespace=namespace,
            label_selector=labels
        )
        
        results = []
        for pod in pods.items:
            if name and not pod.metadata.name.startswith(name):
                continue
            
            # 容器状态
            container_statuses = []
            for cs in (pod.status.container_statuses or []):
                status = "Running" if cs.ready else "NotReady"
                restarts = cs.restart_count
                container_statuses.append(f"{cs.name}: {status} (restarts: {restarts})")
            
            results.append(
                f"Pod: {pod.metadata.name}\n"
                f"  Phase: {pod.status.phase}\n"
                f"  Node: {pod.spec.node_name}\n"
                f"  Containers: {', '.join(container_statuses)}"
            )
        
        return "\n\n".join(results[:10]) if results else "未找到匹配的 Pod"


class LogSearchTool(Tool):
    """日志搜索工具"""
    
    name = "log_search"
    description = "搜索应用日志，支持关键字和时间范围筛选"
    parameters = {
        "type": "object",
        "properties": {
            "service": {
                "type": "string",
                "description": "服务名称"
            },
            "keywords": {
                "type": "string",
                "description": "搜索关键字，如 'error timeout'"
            },
            "time_range": {
                "type": "string",
                "description": "时间范围",
                "default": "15m"
            },
            "level": {
                "type": "string",
                "enum": ["error", "warn", "info", "debug"],
                "description": "日志级别"
            },
            "limit": {
                "type": "integer",
                "description": "返回条数",
                "default": 20
            }
        },
        "required": ["service"]
    }
    
    async def execute(
        self,
        service: str,
        keywords: str = None,
        time_range: str = "15m",
        level: str = None,
        limit: int = 20
    ) -> str:
        """搜索日志"""
        # 这里以 Loki 为例
        query = f'{{app="{service}"}}'
        
        if level:
            query += f' |= "{level.upper()}"'
        if keywords:
            for kw in keywords.split():
                query += f' |= "{kw}"'
        
        # 调用 Loki API
        logs = await self._query_loki(query, time_range, limit)
        
        if not logs:
            return f"未找到 {service} 的相关日志"
        
        return self._format_logs(logs)
```

### 4.3 工具注册

```python
class ToolRegistry:
    """工具注册中心"""
    
    def __init__(self):
        self._tools: Dict[str, Tool] = {}
    
    def register(self, tool: Tool):
        self._tools[tool.name] = tool
    
    def get_tools_prompt(self) -> str:
        """生成工具描述供 LLM 使用"""
        descriptions = []
        for tool in self._tools.values():
            descriptions.append(
                f"### {tool.name}\n"
                f"描述: {tool.description}\n"
                f"参数: {json.dumps(tool.parameters, ensure_ascii=False, indent=2)}"
            )
        return "\n\n".join(descriptions)
    
    async def execute(self, name: str, **kwargs) -> str:
        if name not in self._tools:
            raise ValueError(f"未知工具: {name}")
        return await self._tools[name].execute(**kwargs)

# 初始化工具注册
def create_tool_registry(config: Config) -> ToolRegistry:
    registry = ToolRegistry()
    
    registry.register(PrometheusQueryTool(config.prometheus_url))
    registry.register(KubernetesTool())
    registry.register(LogSearchTool(config.loki_url))
    registry.register(ConfluenceSearchTool(config.confluence_url, config.confluence_token))
    registry.register(AlertHistoryTool(config.database_url))
    registry.register(SlackNotifyTool(config.slack_webhook))
    
    return registry
```

---

## 五、RAG 知识库设计

### 5.1 知识来源

```
┌─────────────────────────────────────────────────────────────┐
│                    Knowledge Sources                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ Confluence  │  │  Runbooks   │  │  历史案例   │        │
│  │  技术文档   │  │  处理手册   │  │  诊断记录   │        │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘        │
│         │                │                │                │
│         ▼                ▼                ▼                │
│  ┌─────────────────────────────────────────────────────┐  │
│  │              Document Processor                       │  │
│  │  • 文档解析  • 分块  • 清洗  • 元数据提取           │  │
│  └─────────────────────────┬───────────────────────────┘  │
│                            │                               │
│                            ▼                               │
│  ┌─────────────────────────────────────────────────────┐  │
│  │              Embedding + Index                        │  │
│  │  • OpenAI Embedding  • Milvus Vector DB              │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 5.2 文档处理流水线

```python
from typing import List
from dataclasses import dataclass
import re

@dataclass
class DocumentChunk:
    """文档块"""
    id: str
    content: str
    metadata: Dict[str, str]  # source, title, service, type
    embedding: List[float] = None

class ConfluenceLoader:
    """Confluence 文档加载器"""
    
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.token = token
        self.client = httpx.AsyncClient(
            headers={"Authorization": f"Bearer {token}"}
        )
    
    async def load_space(self, space_key: str) -> List[Dict]:
        """加载整个空间的文档"""
        documents = []
        start = 0
        limit = 50
        
        while True:
            response = await self.client.get(
                f"{self.base_url}/wiki/rest/api/content",
                params={
                    "spaceKey": space_key,
                    "type": "page",
                    "status": "current",
                    "expand": "body.storage,metadata.labels",
                    "start": start,
                    "limit": limit
                }
            )
            data = response.json()
            
            for page in data.get("results", []):
                documents.append({
                    "id": page["id"],
                    "title": page["title"],
                    "content": self._clean_html(page["body"]["storage"]["value"]),
                    "labels": [l["name"] for l in page.get("metadata", {}).get("labels", {}).get("results", [])],
                    "url": f"{self.base_url}/wiki{page['_links']['webui']}"
                })
            
            if len(data.get("results", [])) < limit:
                break
            start += limit
        
        return documents
    
    def _clean_html(self, html: str) -> str:
        """清理 HTML，提取纯文本"""
        from bs4 import BeautifulSoup
        soup = BeautifulSoup(html, "html.parser")
        
        # 移除脚本和样式
        for script in soup(["script", "style"]):
            script.decompose()
        
        return soup.get_text(separator="\n", strip=True)


class DocumentChunker:
    """文档分块器"""
    
    def __init__(self, chunk_size: int = 500, chunk_overlap: int = 50):
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap
    
    def chunk(self, document: Dict) -> List[DocumentChunk]:
        """将文档分块"""
        content = document["content"]
        chunks = []
        
        # 按段落分割
        paragraphs = self._split_paragraphs(content)
        
        current_chunk = ""
        for para in paragraphs:
            if len(current_chunk) + len(para) > self.chunk_size:
                if current_chunk:
                    chunks.append(self._create_chunk(document, current_chunk, len(chunks)))
                current_chunk = para
            else:
                current_chunk += "\n\n" + para if current_chunk else para
        
        if current_chunk:
            chunks.append(self._create_chunk(document, current_chunk, len(chunks)))
        
        return chunks
    
    def _split_paragraphs(self, text: str) -> List[str]:
        """按段落分割，保持代码块完整"""
        paragraphs = re.split(r'\n{2,}', text)
        return [p.strip() for p in paragraphs if p.strip()]
    
    def _create_chunk(self, document: Dict, content: str, index: int) -> DocumentChunk:
        return DocumentChunk(
            id=f"{document['id']}_{index}",
            content=content,
            metadata={
                "source": "confluence",
                "title": document["title"],
                "url": document["url"],
                "labels": ",".join(document.get("labels", []))
            }
        )


class RAGSystem:
    """RAG 检索系统"""
    
    def __init__(self, embedding_model, vector_db):
        self.embedding = embedding_model
        self.vector_db = vector_db
    
    async def retrieve(
        self, 
        query: str, 
        filters: Dict = None,
        top_k: int = 5
    ) -> str:
        """检索相关文档"""
        # 1. Query Embedding
        query_embedding = await self.embedding.encode(query)
        
        # 2. Vector Search
        results = await self.vector_db.search(
            vector=query_embedding,
            top_k=top_k * 2,  # 检索更多用于重排
            filters=filters
        )
        
        # 3. Rerank (可选)
        if len(results) > top_k:
            results = await self._rerank(query, results, top_k)
        
        # 4. Format Results
        return self._format_results(results)
    
    def _format_results(self, results: List[DocumentChunk]) -> str:
        """格式化检索结果"""
        formatted = []
        for i, chunk in enumerate(results):
            formatted.append(
                f"### 文档 {i+1}: {chunk.metadata['title']}\n"
                f"来源: {chunk.metadata['url']}\n"
                f"内容:\n{chunk.content}\n"
            )
        return "\n---\n".join(formatted)
```

### 5.3 知识库更新策略

```python
class KnowledgeBaseUpdater:
    """知识库增量更新"""
    
    def __init__(self, loader: ConfluenceLoader, chunker: DocumentChunker, rag: RAGSystem):
        self.loader = loader
        self.chunker = chunker
        self.rag = rag
        self.last_sync: Dict[str, datetime] = {}
    
    async def sync_incremental(self, space_key: str):
        """增量同步 Confluence 空间"""
        # 获取上次同步后更新的文档
        last_sync = self.last_sync.get(space_key, datetime.min)
        
        updated_docs = await self.loader.load_updated_since(space_key, last_sync)
        
        for doc in updated_docs:
            # 删除旧的 chunks
            await self.rag.vector_db.delete_by_metadata({"source_id": doc["id"]})
            
            # 创建新的 chunks
            chunks = self.chunker.chunk(doc)
            
            # 生成 embedding 并索引
            for chunk in chunks:
                chunk.embedding = await self.rag.embedding.encode(chunk.content)
            
            await self.rag.vector_db.insert(chunks)
        
        self.last_sync[space_key] = datetime.now()
        
        return len(updated_docs)
```

---

## 六、工作流设计

### 6.1 告警处理工作流

```python
from enum import Enum
from typing import Optional

class AlertWorkflowState(Enum):
    RECEIVED = "received"           # 接收
    DEDUPED = "deduped"             # 去重
    ENRICHED = "enriched"           # 富化
    DIAGNOSING = "diagnosing"       # 诊断中
    DIAGNOSED = "diagnosed"         # 已诊断
    NOTIFIED = "notified"           # 已通知
    ESCALATED = "escalated"         # 已升级
    RESOLVED = "resolved"           # 已解决
    CLOSED = "closed"               # 已关闭

class AlertWorkflow:
    """告警处理工作流"""
    
    def __init__(self, agent: DoDAgent, notifier: SlackNotifier):
        self.agent = agent
        self.notifier = notifier
        self.state_machine = self._build_state_machine()
    
    async def process(self, alert: StandardAlert) -> WorkflowResult:
        """处理告警"""
        ctx = WorkflowContext(alert=alert, state=AlertWorkflowState.RECEIVED)
        
        try:
            # 1. 去重检查
            if await self._is_duplicate(alert):
                ctx.state = AlertWorkflowState.DEDUPED
                return WorkflowResult(ctx, action="deduplicated")
            
            # 2. 告警富化
            ctx = await self._enrich_alert(ctx)
            ctx.state = AlertWorkflowState.ENRICHED
            
            # 3. AI 诊断
            ctx.state = AlertWorkflowState.DIAGNOSING
            diagnosis = await self.agent.diagnose_alert(alert)
            ctx.diagnosis = diagnosis
            ctx.state = AlertWorkflowState.DIAGNOSED
            
            # 4. 根据诊断结果决定下一步
            if diagnosis.confidence >= 0.8 and diagnosis.severity != "critical":
                # 高置信度 + 非严重：自动通知 + 建议
                await self._notify_with_suggestion(ctx)
                ctx.state = AlertWorkflowState.NOTIFIED
            else:
                # 低置信度或严重告警：升级到人工
                await self._escalate(ctx)
                ctx.state = AlertWorkflowState.ESCALATED
            
            # 5. 记录诊断结果
            await self._save_diagnosis(ctx)
            
            return WorkflowResult(ctx, action="processed")
            
        except Exception as e:
            # 异常情况：升级到人工
            await self._escalate_with_error(ctx, e)
            return WorkflowResult(ctx, action="error", error=str(e))
    
    async def _enrich_alert(self, ctx: WorkflowContext) -> WorkflowContext:
        """富化告警信息"""
        alert = ctx.alert
        
        # 添加服务依赖信息
        if service := alert.labels.get("service"):
            ctx.dependencies = await self._get_service_dependencies(service)
        
        # 添加最近部署信息
        ctx.recent_deployments = await self._get_recent_deployments(
            alert.labels.get("namespace", "default")
        )
        
        # 添加关联告警
        ctx.related_alerts = await self._get_related_alerts(alert)
        
        return ctx
    
    async def _notify_with_suggestion(self, ctx: WorkflowContext):
        """发送诊断结果和建议"""
        message = self._build_diagnosis_message(ctx)
        await self.notifier.send(
            channel=self._get_channel(ctx.alert),
            message=message,
            attachments=self._build_attachments(ctx)
        )
    
    def _build_diagnosis_message(self, ctx: WorkflowContext) -> str:
        """构建诊断消息"""
        d = ctx.diagnosis
        return f"""
🔔 *告警诊断报告*

*告警*: {ctx.alert.title}
*严重级别*: {ctx.alert.severity.value}
*服务*: {ctx.alert.labels.get('service', 'Unknown')}

---

📋 *根因分析* (置信度: {d.confidence:.0%})
{d.root_cause}

⚠️ *影响范围*
{d.impact}

✅ *建议处理步骤*
{self._format_suggestions(d.suggested_actions)}

📚 *参考文档*
{self._format_references(d.references)}

---
_诊断由 DoD Agent 自动生成，如有疑问请 @oncall_
"""
```

### 6.2 咨询问答工作流

```python
class QueryWorkflow:
    """知识咨询工作流"""
    
    async def process(self, query: str, user: str, channel: str) -> str:
        """处理咨询问题"""
        # 1. 意图识别
        intent = await self.router.route(query)
        
        # 2. 根据意图处理
        if intent == IntentType.KNOWLEDGE_QUERY:
            return await self._handle_knowledge_query(query)
        elif intent == IntentType.STATUS_CHECK:
            return await self._handle_status_check(query)
        elif intent == IntentType.HISTORY_LOOKUP:
            return await self._handle_history_lookup(query)
        else:
            return await self._handle_general_query(query)
    
    async def _handle_knowledge_query(self, query: str) -> str:
        """处理知识查询"""
        # RAG 检索
        docs = await self.rag.retrieve(query, top_k=3)
        
        # LLM 生成回答
        prompt = f"""
基于以下文档回答用户问题。如果文档中没有相关信息，请明确说明。

## 相关文档
{docs}

## 用户问题
{query}

## 要求
1. 直接回答问题，不要重复问题
2. 引用具体文档来源
3. 如果不确定，说明并建议咨询相关负责人
"""
        response = await self.llm.generate(prompt)
        return response
```

---

## 七、Slack 集成

### 7.1 Slack Bot 设计

```python
from slack_bolt.async_app import AsyncApp
from slack_bolt.adapter.socket_mode.async_handler import AsyncSocketModeHandler

class DoDSlackBot:
    """DoD Agent Slack Bot"""
    
    def __init__(self, config: Config, agent: DoDAgent, workflow: AlertWorkflow):
        self.app = AsyncApp(token=config.slack_bot_token)
        self.agent = agent
        self.workflow = workflow
        
        self._register_handlers()
    
    def _register_handlers(self):
        """注册消息处理器"""
        
        # 响应 @mention
        @self.app.event("app_mention")
        async def handle_mention(event, say):
            text = event.get("text", "")
            user = event.get("user")
            channel = event.get("channel")
            
            # 移除 @mention
            query = self._extract_query(text)
            
            # 显示处理中
            await say(f"🤔 正在分析您的问题...")
            
            # 处理查询
            response = await self.agent.handle_query(query, user, channel)
            
            await say(response)
        
        # 响应私信
        @self.app.event("message")
        async def handle_dm(event, say):
            if event.get("channel_type") == "im":
                query = event.get("text", "")
                user = event.get("user")
                
                response = await self.agent.handle_query(query, user)
                await say(response)
        
        # 快捷命令
        @self.app.command("/dod")
        async def handle_command(ack, command, say):
            await ack()
            
            subcommand = command.get("text", "").split()[0]
            
            if subcommand == "status":
                status = await self._get_system_status()
                await say(status)
            elif subcommand == "alerts":
                alerts = await self._get_active_alerts()
                await say(alerts)
            elif subcommand == "help":
                await say(self._get_help_message())
            else:
                await say(f"未知命令: {subcommand}，使用 /dod help 查看帮助")
        
        # 交互式按钮
        @self.app.action("escalate_alert")
        async def handle_escalate(ack, body, say):
            await ack()
            
            alert_id = body["actions"][0]["value"]
            user = body["user"]["id"]
            
            await self.workflow.manual_escalate(alert_id, user)
            await say(f"✅ 告警 {alert_id} 已升级，通知相关负责人")
        
        @self.app.action("mark_resolved")
        async def handle_resolve(ack, body, say):
            await ack()
            
            alert_id = body["actions"][0]["value"]
            user = body["user"]["id"]
            
            await self.workflow.mark_resolved(alert_id, user)
            await say(f"✅ 告警 {alert_id} 已标记为解决")
    
    def _get_help_message(self) -> str:
        return """
🤖 *DoD Agent 使用指南*

*直接对话*
• @DoD Agent 查询某服务为什么报警
• @DoD Agent 如何重启 order-service
• @DoD Agent 查看最近的部署记录

*快捷命令*
• `/dod status` - 查看系统整体状态
• `/dod alerts` - 查看当前活跃告警
• `/dod help` - 显示帮助信息

*自动诊断*
• 收到告警后自动分析并推送诊断报告
• 点击「查看详情」了解更多
• 点击「升级」通知值班人员
"""
```

### 7.2 告警消息格式

```python
class SlackMessageBuilder:
    """Slack 消息构建器"""
    
    def build_alert_notification(self, ctx: WorkflowContext) -> Dict:
        """构建告警通知消息"""
        alert = ctx.alert
        diagnosis = ctx.diagnosis
        
        # 严重级别对应的颜色
        colors = {
            AlertSeverity.CRITICAL: "#FF0000",
            AlertSeverity.WARNING: "#FFA500",
            AlertSeverity.INFO: "#36A64F"
        }
        
        blocks = [
            {
                "type": "header",
                "text": {
                    "type": "plain_text",
                    "text": f"🔔 {alert.title}",
                    "emoji": True
                }
            },
            {
                "type": "section",
                "fields": [
                    {"type": "mrkdwn", "text": f"*严重级别:*\n{alert.severity.value}"},
                    {"type": "mrkdwn", "text": f"*服务:*\n{alert.labels.get('service', 'N/A')}"},
                    {"type": "mrkdwn", "text": f"*环境:*\n{alert.labels.get('env', 'N/A')}"},
                    {"type": "mrkdwn", "text": f"*时间:*\n{alert.starts_at.strftime('%Y-%m-%d %H:%M:%S')}"}
                ]
            },
            {"type": "divider"},
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"*📋 AI 诊断结果* (置信度: {diagnosis.confidence:.0%})\n\n{diagnosis.root_cause}"
                }
            },
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": f"*✅ 建议处理*\n{self._format_actions(diagnosis.suggested_actions)}"
                }
            },
            {"type": "divider"},
            {
                "type": "actions",
                "elements": [
                    {
                        "type": "button",
                        "text": {"type": "plain_text", "text": "📊 查看监控"},
                        "url": self._get_grafana_url(alert)
                    },
                    {
                        "type": "button",
                        "text": {"type": "plain_text", "text": "📖 查看文档"},
                        "url": diagnosis.references[0] if diagnosis.references else "#"
                    },
                    {
                        "type": "button",
                        "text": {"type": "plain_text", "text": "🚨 升级"},
                        "style": "danger",
                        "action_id": "escalate_alert",
                        "value": alert.id
                    },
                    {
                        "type": "button",
                        "text": {"type": "plain_text", "text": "✅ 已解决"},
                        "style": "primary",
                        "action_id": "mark_resolved",
                        "value": alert.id
                    }
                ]
            }
        ]
        
        return {
            "blocks": blocks,
            "attachments": [{
                "color": colors.get(alert.severity, "#808080")
            }]
        }
```

---

## 八、部署架构

### 8.1 Kubernetes 部署

```yaml
# dod-agent-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dod-agent
  namespace: observability
spec:
  replicas: 2
  selector:
    matchLabels:
      app: dod-agent
  template:
    metadata:
      labels:
        app: dod-agent
    spec:
      serviceAccountName: dod-agent
      containers:
      - name: dod-agent
        image: your-registry/dod-agent:latest
        ports:
        - containerPort: 8080
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: dod-agent-secrets
              key: openai-api-key
        - name: SLACK_BOT_TOKEN
          valueFrom:
            secretKeyRef:
              name: dod-agent-secrets
              key: slack-bot-token
        - name: PROMETHEUS_URL
          value: "http://prometheus.monitoring:9090"
        - name: CONFLUENCE_URL
          valueFrom:
            configMapKeyRef:
              name: dod-agent-config
              key: confluence-url
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      
      # Vector DB Sidecar
      - name: chroma
        image: ghcr.io/chroma-core/chroma:latest
        ports:
        - containerPort: 8000
        volumeMounts:
        - name: chroma-data
          mountPath: /chroma/chroma
      
      volumes:
      - name: chroma-data
        persistentVolumeClaim:
          claimName: chroma-pvc

---
# ServiceAccount with K8s read permissions
apiVersion: v1
kind: ServiceAccount
metadata:
  name: dod-agent
  namespace: observability

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: dod-agent-reader
rules:
- apiGroups: [""]
  resources: ["pods", "services", "events", "nodes"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: dod-agent-reader-binding
subjects:
- kind: ServiceAccount
  name: dod-agent
  namespace: observability
roleRef:
  kind: ClusterRole
  name: dod-agent-reader
  apiGroup: rbac.authorization.k8s.io
```

### 8.2 Alertmanager 配置

```yaml
# alertmanager.yaml
route:
  receiver: 'dod-agent'
  group_by: ['alertname', 'service']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 4h
  routes:
  - match:
      severity: critical
    receiver: 'dod-agent-critical'
  - match:
      severity: warning
    receiver: 'dod-agent'

receivers:
- name: 'dod-agent'
  webhook_configs:
  - url: 'http://dod-agent.observability:8080/webhook/alertmanager'
    send_resolved: true

- name: 'dod-agent-critical'
  webhook_configs:
  - url: 'http://dod-agent.observability:8080/webhook/alertmanager?priority=critical'
    send_resolved: true
```

---

## 九、可观测性

### 9.1 监控指标

```python
from prometheus_client import Counter, Histogram, Gauge

# 核心指标
ALERT_RECEIVED = Counter(
    'dod_agent_alerts_received_total',
    'Total alerts received',
    ['severity', 'service']
)

ALERT_DIAGNOSED = Counter(
    'dod_agent_alerts_diagnosed_total',
    'Total alerts diagnosed',
    ['severity', 'result']  # result: auto_resolved, escalated, failed
)

DIAGNOSIS_LATENCY = Histogram(
    'dod_agent_diagnosis_latency_seconds',
    'Alert diagnosis latency',
    buckets=[1, 5, 10, 30, 60, 120, 300]
)

DIAGNOSIS_CONFIDENCE = Histogram(
    'dod_agent_diagnosis_confidence',
    'Diagnosis confidence score',
    buckets=[0.1, 0.3, 0.5, 0.7, 0.8, 0.9, 0.95, 1.0]
)

LLM_TOKENS_USED = Counter(
    'dod_agent_llm_tokens_total',
    'Total LLM tokens used',
    ['model', 'type']  # type: prompt, completion
)

RAG_RETRIEVAL_LATENCY = Histogram(
    'dod_agent_rag_retrieval_latency_seconds',
    'RAG retrieval latency'
)

TOOL_EXECUTION = Counter(
    'dod_agent_tool_executions_total',
    'Tool executions',
    ['tool', 'status']  # status: success, error
)
```

### 9.2 诊断质量追踪

```python
@dataclass
class DiagnosisFeedback:
    """诊断反馈记录"""
    diagnosis_id: str
    alert_id: str
    user_feedback: str  # helpful, not_helpful, incorrect
    actual_root_cause: Optional[str]
    actual_resolution: Optional[str]
    feedback_time: datetime

class DiagnosisQualityTracker:
    """诊断质量追踪"""
    
    def __init__(self, db):
        self.db = db
    
    async def record_feedback(self, feedback: DiagnosisFeedback):
        """记录用户反馈"""
        await self.db.insert("diagnosis_feedback", feedback)
        
        # 更新诊断准确率指标
        if feedback.user_feedback == "helpful":
            DIAGNOSIS_HELPFUL.labels(service=feedback.service).inc()
        elif feedback.user_feedback == "incorrect":
            DIAGNOSIS_INCORRECT.labels(service=feedback.service).inc()
    
    async def get_accuracy_report(self, days: int = 30) -> Dict:
        """生成准确率报告"""
        feedbacks = await self.db.query(
            "SELECT * FROM diagnosis_feedback WHERE feedback_time > ?",
            datetime.now() - timedelta(days=days)
        )
        
        total = len(feedbacks)
        helpful = sum(1 for f in feedbacks if f.user_feedback == "helpful")
        
        return {
            "total_diagnoses": total,
            "helpful_rate": helpful / total if total > 0 else 0,
            "by_service": self._group_by_service(feedbacks),
            "common_misses": self._analyze_misses(feedbacks)
        }
```

---

## 十、扩展性设计

### 10.1 多部门适配

```python
class DepartmentAdapter:
    """部门适配器基类"""
    
    @abstractmethod
    def get_alert_sources(self) -> List[AlertSource]:
        """获取告警源"""
        pass
    
    @abstractmethod
    def get_tools(self) -> List[Tool]:
        """获取部门特定工具"""
        pass
    
    @abstractmethod
    def get_knowledge_spaces(self) -> List[str]:
        """获取知识库空间"""
        pass
    
    @abstractmethod
    def get_notification_channels(self) -> Dict[str, str]:
        """获取通知渠道"""
        pass

class SREAdapter(DepartmentAdapter):
    """SRE 部门适配"""
    
    def get_tools(self) -> List[Tool]:
        return [
            PrometheusQueryTool(),
            KubernetesTool(),
            LogSearchTool(),
            GrafanaTool()
        ]
    
    def get_knowledge_spaces(self) -> List[str]:
        return ["SRE-Runbooks", "Architecture-Docs"]

class DBAAdapter(DepartmentAdapter):
    """DBA 部门适配"""
    
    def get_tools(self) -> List[Tool]:
        return [
            MySQLQueryTool(),
            SlowQueryAnalyzer(),
            DatabaseStatusTool(),
            BackupStatusTool()
        ]
    
    def get_knowledge_spaces(self) -> List[str]:
        return ["DBA-Runbooks", "Database-Best-Practices"]

class SecurityAdapter(DepartmentAdapter):
    """安全部门适配"""
    
    def get_tools(self) -> List[Tool]:
        return [
            WAFLogTool(),
            ThreatIntelTool(),
            AccessLogAnalyzer(),
            VulnerabilityScanTool()
        ]
```

### 10.2 插件系统

```python
class PluginManager:
    """插件管理器"""
    
    def __init__(self):
        self._plugins: Dict[str, Plugin] = {}
    
    def register(self, plugin: Plugin):
        """注册插件"""
        self._plugins[plugin.name] = plugin
        
        # 注册插件的工具
        for tool in plugin.get_tools():
            self.tool_registry.register(tool)
        
        # 注册插件的处理器
        for handler in plugin.get_handlers():
            self.handler_registry.register(handler)
    
    def load_from_config(self, config_path: str):
        """从配置加载插件"""
        config = yaml.safe_load(open(config_path))
        
        for plugin_config in config.get("plugins", []):
            plugin_class = self._load_plugin_class(plugin_config["module"])
            plugin = plugin_class(**plugin_config.get("config", {}))
            self.register(plugin)

# 插件配置示例
# plugins.yaml
plugins:
  - name: mysql-plugin
    module: dod_plugins.mysql.MySQLPlugin
    config:
      host: mysql.default
      readonly_user: dod_readonly
  
  - name: redis-plugin
    module: dod_plugins.redis.RedisPlugin
    config:
      cluster: redis-cluster.default
```

---

## 十一、实施路线图

### Phase 1: MVP（4周）

| 周次 | 目标 | 交付物 |
|:---|:---|:---|
| Week 1 | 基础架构搭建 | Gateway + Alertmanager 集成 |
| Week 2 | Agent Core 开发 | ReAct 诊断引擎 + 基础工具 |
| Week 3 | RAG 集成 | Confluence 同步 + 向量检索 |
| Week 4 | Slack 集成 | Bot + 告警通知 |

**MVP 功能**：
- 接收 Alertmanager 告警
- 自动诊断并推送到 Slack
- 基于 Confluence 回答咨询问题

### Phase 2: 增强（4周）

| 周次 | 目标 | 交付物 |
|:---|:---|:---|
| Week 5-6 | 工具扩展 | K8s / 日志 / Grafana 工具 |
| Week 7 | 告警关联 | 关联分析 + 聚合 |
| Week 8 | 质量优化 | 反馈收集 + 模型调优 |

### Phase 3: 扩展（4周）

| 周次 | 目标 | 交付物 |
|:---|:---|:---|
| Week 9-10 | 多部门支持 | DBA / 安全适配器 |
| Week 11-12 | 插件系统 | 插件框架 + 文档 |

---

## 十二、成本估算

### 12.1 LLM 成本

基于日均 100 次诊断，每次诊断平均 3 轮 Agent Loop：

| 项目 | 估算 |
|:---|:---|
| 每次诊断 Token | ~4000 (prompt) + ~1000 (completion) |
| 日均 Token | 100 × 3 × 5000 = 1.5M tokens |
| 月均 Token | 45M tokens |
| GPT-4 成本 | ~$1350/月（$30/1M input + $60/1M output） |
| GPT-4-turbo 成本 | ~$450/月（$10/1M input + $30/1M output） |

**优化策略**：
- 简单告警使用 GPT-3.5（成本降低 90%）
- 实现 Semantic Cache（相似问题复用）
- 优化 Prompt（减少 token 消耗）

### 12.2 基础设施成本

| 资源 | 规格 | 月成本（估算） |
|:---|:---|:---|
| DoD Agent Pod | 2 × (1C/1G) | ~$40 |
| Chroma Vector DB | 1 × (2C/4G) + 50G SSD | ~$80 |
| Redis (缓存) | 1G | ~$30 |
| **合计** | | ~$150/月 |

---

## 十三、风险与缓解

| 风险 | 影响 | 缓解措施 |
|:---|:---|:---|
| LLM 响应延迟 | 诊断慢，影响 MTTR | 异步处理 + 超时降级 |
| 诊断错误 | 误导处理方向 | 显示置信度 + 人工确认 |
| 知识库过时 | 回答不准确 | 增量同步 + 反馈机制 |
| API 限流 | 服务不可用 | Fallback 模型 + 队列缓冲 |
| 敏感信息泄露 | 安全风险 | PII 过滤 + 审计日志 |

---

## 总结

DoD Agent 通过 AI 能力增强运维效率，核心价值：

1. **降低 MTTR**：自动诊断减少人工分析时间
2. **知识沉淀**：将专家经验转化为可检索知识
3. **标准化处理**：常见问题自动化处理流程
4. **可扩展**：插件化设计支持多部门复用

> 第一阶段聚焦**只读诊断**，验证价值后再逐步扩展自动化操作能力。
