---
title: DoD Agent 完整设计文档：电商告警自动处理系统
date: 2026-04-03
categories:
  - AI 与 Agent
tags:
- AI Agent
- DevOps
- 告警处理
- 系统设计
- ReACT
- 状态机
toc: true
version: 2.0
---

<!-- toc -->

> **文档说明**
> 
> 本文档整合了 DoD Agent 的两个设计版本：
> - **v1 (2026-03-09)**：初始设计，基于纯 ReACT 架构，包含详细的技术实现
> - **v2 (2026-04-03)**：重设计版本，采用状态机 + ReACT 混合架构，强调分级决策和渐进式学习
> 
> 本文档结合了两个版本的精华，提供完整的设计方案和实施指南。

---

## Part 1: 执行摘要和架构决策

### 1.1 项目背景

电商公司日常运维面临大量告警处理工作，包括基础设施告警、应用告警和业务告警。

**当前痛点**：

| 痛点 | 影响 |
|:---|:---|
| 告警量大（50-200条/天） | 值班人员疲劳，响应延迟 |
| 重复性问题多 | 80% 问题有标准处理流程 |
| 知识分散 | Confluence 文档难以快速定位 |
| 跨部门协作 | 告警升级和分发效率低 |
| 被动响应 | 现有 DoD Agent 仅提供查询功能 |

### 1.2 设计目标

重新设计 DoD Agent 为一个**事件驱动的智能协调型 Agent**，具备以下能力：

```
┌─────────────────────────────────────────────────────────────┐
│                    DoD Agent 核心目标                        │
├─────────────────────────────────────────────────────────────┤
│  🎯 智能分析     - 自动分析告警原因，协调多个子系统          │
│  🤖 自主决策     - 基于风险等级自动决定处理方式              │
│  📚 智能问答     - 基于 Confluence 知识库回答咨询            │
│  🔄 标准化处理   - 常见问题自动生成处理建议                  │
│  📊 告警聚合     - 关联告警智能聚合，减少噪音                │
│  🔒 可控性       - 状态机保证流程可控、可监控、可恢复        │
│  📈 学习能力     - 从历史数据和反馈中持续学习和优化          │
│  🚀 可扩展       - 支持扩展到其他部门（客服/安全/DBA）       │
└─────────────────────────────────────────────────────────────┘
```

### 1.3 设计原则

1. **只读诊断优先**：第一阶段只做诊断和建议，不执行危险操作
2. **人机协作**：Agent 辅助决策，关键操作人工确认
3. **渐进增强**：从简单场景开始，逐步扩展能力
4. **可观测性**：完整的日志、指标和追踪
5. **状态可控**：通过状态机保证流程可控、可监控、可恢复

### 1.4 核心架构选择

**架构演进**：

- **v1 方案**：纯 ReACT Agent（灵活但难以控制）
- **v2 方案**：状态机 + ReACT 混合架构（可控且智能）✅

**最终选择：增强型 ReACT Agent with 状态机**

- 基于现有 ReACT 框架，增加状态机管理和决策引擎
- 单体 Agent + 工具调用模式
- 状态机 + ReACT 混合工作流
- 分级自主决策 + 可配置策略
- 4阶段渐进式学习能力演进

### 1.5 技术选型

| 组件 | 选型 | 理由 |
|:---|:---|:---|
| **LLM** | OpenAI GPT-4 / Claude-3.5-Sonnet | 推理能力强，工具调用成熟 |
| **实现语言** | Go | 现有系统技术栈，性能优秀 |
| **告警源** | Prometheus + Alertmanager | 已有系统，Webhook 集成 |
| **知识库** | Confluence + RAG | 利用现有文档 |
| **交互渠道** | Seatalk | 团队主要沟通工具 |
| **部署平台** | Kubernetes | 已有基础设施 |
| **向量数据库** | Chroma / Milvus | 本地部署，数据安全 |
| **状态管理** | 数据库 + 内存缓存 | 持久化 + 高性能 |

---

## Part 2: 系统架构

### 2.1 整体架构图

```
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
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    工具集成层                                    │   │
│  │  [告警API] [知识库] [Jira] [Seatalk] [SOP执行器] [DoD查询]    │   │
│  │  [Prometheus] [Kubernetes] [Grafana] [Log System]              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 核心组件

#### 2.2.1 DoD Agent 核心结构

```go
type DoDAgent struct {
    // 状态机：管理告警处理生命周期
    stateMachine *AlertStateMachine
    
    // ReACT引擎：智能分析和工具调用
    reactEngine *ReACTEngine
    
    // 决策引擎：分级决策和策略配置
    decisionEngine *DecisionEngine
    
    // 上下文管理：维护会话状态
    contextManager *ContextManager
    
    // 工具注册表：所有可用的MCP工具
    toolRegistry *ToolRegistry
    
    // RAG系统：知识检索
    ragSystem *RAGSystem
    
    // 学习模块（Phase 2+）
    learningModule *LearningModule
}
```

#### 2.2.2 告警上下文

```go
type AlertContext struct {
    // 基础信息
    AlertID      string
    Alert        *Alert
    Team         string
    StartTime    time.Time
    
    // 分析结果
    RiskAssessment *RiskAssessment
    RiskLevel      RiskLevel
    HasKnownSolution bool
    SuggestedSOP   *SOP
    
    // 决策信息
    Decision       *DecisionResult
    RequireConfirm bool
    ConfirmTimeout time.Duration
    
    // 处理记录
    Actions        []Action
    StateHistory   []StateHistoryEntry
    
    // DoD信息
    DoDInfo        *DoDData
    DoDNotified    bool
    
    // 事件信息
    EventID        string
    EventCreated   bool
    
    // 失败信息
    FailureReason  string
    RetryCount     int
}
```

### 2.3 数据流设计

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
                                            │  状态机初始化 │
                                            │  (State: NEW) │
                                            └───────┬───────┘
                                                    │
                                                    ▼
                                            ┌───────────────┐
                                            │  ReACT 分析   │
                                            │ (ANALYZING)   │
                                            └───────┬───────┘
                                                    │
                                                    ▼
                                            ┌───────────────┐
                                            │  决策引擎     │
                                            │ (风险评估)    │
                                            └───────┬───────┘
                                                    │
                    ┌───────────────────────────────┼───────────────────────────┐
                    │                               │                           │
                    ▼                               ▼                           ▼
            ┌───────────────┐               ┌───────────────┐           ┌───────────────┐
            │ Auto Resolve  │               │ Wait Confirm  │           │ Escalate DoD  │
            │  (自动处理)    │               │  (等待确认)    │           │  (立即升级)    │
            └───────────────┘               └───────────────┘           └───────────────┘
```

### 2.4 处理流程

```
告警接收 → [状态:NEW] → ReACT分析 → 决策引擎判断 →
    ├─ 低风险：自动处理 → [状态:AUTO_RESOLVING]
    ├─ 中风险：建议+确认 → [状态:WAITING_CONFIRM]
    ├─ 高风险：必须确认 → [状态:WAITING_CONFIRM]
    └─ 严重：立即升级 → [状态:DOD_NOTIFIED]
                              ↓
                    用户确认/超时/自动
                              ↓
    ├─ 可自动解决 → 执行SOP → [状态:RESOLVED]
    ├─ 需要DoD → 查找DoD → 通知 → [状态:DOD_NOTIFIED]
    └─ 创建事件 → [状态:EVENT_CREATED]
```

---

## Part 3: 核心模块详细设计

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
    SEATALK = "seatalk"
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

### 3.3 Agent Core（ReACT 引擎）

基于 ReAct 模式的诊断引擎，与状态机集成。

```python
class DoDAgent:
    """DoD Agent 核心引擎"""
    
    def __init__(self, llm, tools: ToolRegistry, rag: RAGSystem, memory: Memory):
        self.llm = llm
        self.tools = tools
        self.rag = rag
        self.memory = memory
        self.max_iterations = 10
    
    async def diagnose_alert(self, alert: StandardAlert, state: AlertState) -> DiagnosisResult:
        """告警诊断主流程（状态感知）"""
        
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
        
        # 4. 获取状态特定的工具列表
        allowed_tools = self._get_allowed_tools_for_state(state)
        
        # 5. ReACT 诊断循环（状态约束）
        diagnosis_steps = []
        for i in range(self.max_iterations):
            response = await self.llm.generate(
                self._build_diagnosis_prompt(context, diagnosis_steps, state, allowed_tools)
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
                # 验证工具是否在当前状态允许使用
                if action.tool not in allowed_tools:
                    diagnosis_steps.append({
                        "thought": action.thought,
                        "error": f"工具 {action.tool} 在当前状态 {state} 不可用"
                    })
                    continue
                
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
    
    def _get_allowed_tools_for_state(self, state: AlertState) -> List[str]:
        """根据状态返回允许的工具列表"""
        tool_map = {
            AlertState.ANALYZING: [
                "search_knowledge_base",
                "query_alert_history",
                "analyze_logs",
                "check_metrics",
                "search_similar_alerts"
            ],
            AlertState.AUTO_RESOLVING: [
                "execute_sop",
                "restart_service",
                "clear_cache",
                "update_config"
            ],
            AlertState.EXECUTING_SOP: [
                "execute_sop",
                "check_sop_status",
                "verify_resolution"
            ],
            AlertState.DOD_NOTIFIED: [
                "get_dod_info",
                "send_seatalk_message",
                "create_jira_ticket"
            ]
        }
        return tool_map.get(state, [])
    
    def _build_diagnosis_prompt(self, context: str, steps: List, state: AlertState, allowed_tools: List[str]) -> str:
        return f"""
你是一个专业的电商系统运维诊断专家。请根据告警信息和上下文，诊断问题根因。

## 当前状态
{state.value}

## 告警上下文
{context}

## 已执行的诊断步骤
{self._format_steps(steps)}

## 可用工具（当前状态限制）
{self._format_allowed_tools(allowed_tools)}

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

## Part 4: 状态机和决策引擎

### 4.1 状态定义

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
    
    // 已创建事件
    StateEventCreated AlertState = "EVENT_CREATED"
    
    // 已解决
    StateResolved AlertState = "RESOLVED"
    
    // 已关闭
    StateClosed AlertState = "CLOSED"
    
    // 失败/需要人工介入
    StateFailed AlertState = "FAILED"
)
```

### 4.2 状态转换表

| From | To | Condition | Action | Timeout |
|------|----|-----------| -------|---------|
| StateNew | StateAnalyzing | alert != nil | 开始ReACT分析 | 30s |
| StateAnalyzing | StateAutoResolving | risk_level == Low && has_known_solution | 执行自动处理 | 5min |
| StateAnalyzing | StateWaitingConfirm | risk_level == Medium \|\| risk_level == High | 发送确认请求 | 30s (Medium) / 无超时 (High) |
| StateAnalyzing | StateDoDNotified | risk_level == Critical | 立即升级DoD | 10min |
| StateAnalyzing | StateFailed | 分析超时或失败 | 记录失败原因 | - |
| StateWaitingConfirm | StateAutoResolving | 用户确认 && action == auto_resolve | 执行自动处理 | 5min |
| StateWaitingConfirm | StateExecutingSOP | 用户确认 && action == execute_sop | 执行SOP | 5min |
| StateWaitingConfirm | StateDoDNotified | 用户拒绝 | 升级DoD | 10min |
| StateWaitingConfirm | StateAutoResolving | 超时(仅Medium风险) | 自动执行建议操作 | 5min |
| StateAutoResolving | StateResolved | 自动处理成功 | 记录解决方案 | - |
| StateAutoResolving | StateExecutingSOP | 自动处理失败，尝试SOP | 执行SOP | 5min |
| StateAutoResolving | StateDoDNotified | 自动处理失败，无SOP | 升级DoD | 10min |
| StateExecutingSOP | StateResolved | SOP执行成功 | 记录解决方案 | - |
| StateExecutingSOP | StateDoDNotified | SOP执行失败 \|\| 超时 | 升级DoD | 10min |
| StateDoDNotified | StateEventCreated | DoD响应超时 | 创建事件跟踪 | - |
| StateDoDNotified | StateResolved | DoD解决问题 | 记录解决方案 | - |
| StateEventCreated | StateClosed | 事件关闭 | 归档 | - |
| StateResolved | StateClosed | 人工确认关闭 OR 自动关闭（24小时无新告警） | 归档 | 24h（自动关闭） |
| StateFailed | StateDoDNotified | 需要人工介入 | 升级DoD | 10min |
| StateFailed | StateClosed | 标记为无法处理 | 归档 | - |

### 4.3 超时处理策略

| 状态 | 超时时间 | 超时处理 |
|------|---------|---------|
| StateAnalyzing | 30s | 标记失败，通知管理员 |
| StateWaitingConfirm (中风险) | 30s | 自动执行建议操作 |
| StateWaitingConfirm (高风险) | 无超时 | 持续等待人工确认 |
| StateAutoResolving | 5min | 转到ExecutingSOP或升级DoD |
| StateExecutingSOP | 5min | 标记失败，升级DoD |
| StateDoDNotified | 10min | 创建事件，升级团队负责人 |
| StateResolved | 24h | 自动关闭（如无新告警） |

**注意**：高风险告警的 `StateWaitingConfirm` 无超时机制，必须等待人工确认。

### 4.4 决策引擎设计

#### 4.4.1 风险等级

```go
type RiskLevel int

const (
    RiskLow      RiskLevel = 1  // 自动处理
    RiskMedium   RiskLevel = 2  // 快速确认（30秒超时）
    RiskHigh     RiskLevel = 3  // 必须确认
    RiskCritical RiskLevel = 4  // 立即升级DoD
)
```

#### 4.4.2 风险评估模型

风险评估基于以下因素的加权计算：

| 因素 | 权重 | 说明 |
|------|------|------|
| 环境 (environment) | 30% | 生产环境风险更高 |
| 严重程度 (severity) | 25% | Critical级别需要立即关注 |
| 影响范围 (impact_scope) | 20% | 多市场影响风险更高 |
| 历史模式 (historical_pattern) | 15% | 重复告警可能有已知解决方案 |
| 时间因素 (time_factor) | 10% | 高峰期风险更高 |

#### 4.4.3 风险阈值

| 分数范围 | 风险等级 | 处理方式 |
|---------|---------|---------|
| 0-30 | Low | 自动处理 |
| 31-60 | Medium | 建议+快速确认（30s超时） |
| 61-85 | High | 必须人工确认 |
| 86-100 | Critical | 立即升级DoD |

#### 4.4.4 决策策略配置

```go
type DecisionPolicy struct {
    TeamID   string
    Enabled  bool
    
    // 风险阈值配置（可调整）
    RiskThresholds struct {
        LowToMedium      float64  // 默认 30
        MediumToHigh     float64  // 默认 60
        HighToCritical   float64  // 默认 85
    }
    
    // 超时配置（可调整）
    Timeouts struct {
        MediumRiskConfirm time.Duration  // 默认 30s
        HighRiskConfirm   time.Duration  // 默认 无超时
        SOPExecution      time.Duration  // 默认 5min
        DoDResponse       time.Duration  // 默认 10min
    }
    
    // 自动处理规则
    AutoResolveRules []AutoResolveRule
    
    // 强制升级规则
    ForceEscalateRules []EscalateRule
}
```

#### 4.4.5 决策流程

```
1. 获取团队策略配置
2. 执行风险评估（计算风险分数和等级）
3. 检查强制升级规则（如果匹配，立即升级DoD）
4. 检查自动处理规则（如果匹配，自动执行）
5. 基于风险等级决策：
   - Low: 自动处理
   - Medium: 建议+快速确认（30s超时）
   - High: 必须人工确认
   - Critical: 立即升级DoD
```

---

## Part 5: 工具集成和工作流

### 5.1 工具清单

#### 5.1.1 现有工具（复用）

| 工具 | 功能 | 权限级别 |
|:---|:---|:---|
| `get_dod_info` | 获取DoD信息 | 只读 |
| `get_dod_by_team_id` | 按团队ID查询DoD | 只读 |
| `get_dod_by_sub_team_name` | 按子团队名称查询DoD | 只读 |
| `send_seatalk_message` | 发送Seatalk消息 | 写入 |
| `create_jira_ticket` | 创建Jira工单 | 写入 |
| `search_knowledge_base` | 搜索知识库 | 只读 |
| `execute_sop` | 执行SOP | 写入 |

#### 5.1.2 新增工具

| 工具 | 功能 | 权限级别 | 说明 |
|:---|:---|:---|:---|
| `query_alert_history` | 查询告警历史 | 只读 | 查询历史告警记录 |
| `analyze_logs` | 分析日志 | 只读 | 搜索 ES/Loki 日志 |
| `check_metrics` | 检查监控指标 | 只读 | 查询 Prometheus 指标 |
| `search_similar_alerts` | 搜索相似告警 | 只读 | 基于相似度匹配 |
| `restart_service` | 重启服务 | 写入（需确认） | K8s服务重启 |
| `clear_cache` | 清理缓存 | 写入（需确认） | Redis/Memcached清理 |
| `update_config` | 更新配置 | 写入（需确认） | 配置热更新 |
| `kubernetes_get` | 查询 K8s 资源状态 | 只读 | Pod/Deployment/Service 状态 |
| `grafana_snapshot` | 获取 Grafana 面板截图 | 只读 | 生成监控截图 |

### 5.2 工具实现示例

#### 5.2.1 Prometheus 查询工具

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
        for result in results[:5]:
            metric = result["metric"]
            values = result["values"]
            
            latest = float(values[-1][1]) if values else 0
            avg = sum(float(v[1]) for v in values) / len(values) if values else 0
            
            formatted.append(
                f"指标: {metric}\n"
                f"  最新值: {latest:.2f}\n"
                f"  平均值: {avg:.2f}\n"
                f"  数据点数: {len(values)}"
            )
        
        return "\n\n".join(formatted)
```

#### 5.2.2 Kubernetes 工具

```python
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
            config.load_incluster_config()
        except:
            config.load_kube_config()
        
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
```

#### 5.2.3 日志搜索工具

```python
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
        query = f'{{app="{service}"}}'
        
        if level:
            query += f' |= "{level.upper()}"'
        if keywords:
            for kw in keywords.split():
                query += f' |= "{kw}"'
        
        logs = await self._query_loki(query, time_range, limit)
        
        if not logs:
            return f"未找到 {service} 的相关日志"
        
        return self._format_logs(logs)
```

### 5.3 工具注册

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

def create_tool_registry(config: Config) -> ToolRegistry:
    """初始化工具注册"""
    registry = ToolRegistry()
    
    registry.register(PrometheusQueryTool(config.prometheus_url))
    registry.register(KubernetesTool())
    registry.register(LogSearchTool(config.loki_url))
    registry.register(ConfluenceSearchTool(config.confluence_url, config.confluence_token))
    registry.register(AlertHistoryTool(config.database_url))
    registry.register(SeatalkNotifyTool(config.seatalk_webhook))
    
    return registry
```

### 5.4 告警处理工作流

```python
from enum import Enum
from typing import Optional

class AlertWorkflowState(Enum):
    RECEIVED = "received"
    DEDUPED = "deduped"
    ENRICHED = "enriched"
    DIAGNOSING = "diagnosing"
    DIAGNOSED = "diagnosed"
    NOTIFIED = "notified"
    ESCALATED = "escalated"
    RESOLVED = "resolved"
    CLOSED = "closed"

class AlertWorkflow:
    """告警处理工作流"""
    
    def __init__(self, agent: DoDAgent, notifier: SeatalkNotifier):
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
                await self._notify_with_suggestion(ctx)
                ctx.state = AlertWorkflowState.NOTIFIED
            else:
                await self._escalate(ctx)
                ctx.state = AlertWorkflowState.ESCALATED
            
            # 5. 记录诊断结果
            await self._save_diagnosis(ctx)
            
            return WorkflowResult(ctx, action="processed")
            
        except Exception as e:
            await self._escalate_with_error(ctx, e)
            return WorkflowResult(ctx, action="error", error=str(e))
    
    async def _enrich_alert(self, ctx: WorkflowContext) -> WorkflowContext:
        """富化告警信息"""
        alert = ctx.alert
        
        if service := alert.labels.get("service"):
            ctx.dependencies = await self._get_service_dependencies(service)
        
        ctx.recent_deployments = await self._get_recent_deployments(
            alert.labels.get("namespace", "default")
        )
        
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

### 5.5 咨询问答工作流

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
        docs = await self.rag.retrieve(query, top_k=3)
        
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

## Part 6: RAG 知识库系统

### 6.1 知识来源

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

### 6.2 文档处理流水线

```python
from typing import List
from dataclasses import dataclass
import re

@dataclass
class DocumentChunk:
    """文档块"""
    id: str
    content: str
    metadata: Dict[str, str]
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
            top_k=top_k * 2,
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

### 6.3 知识库更新策略

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
        last_sync = self.last_sync.get(space_key, datetime.min)
        
        updated_docs = await self.loader.load_updated_since(space_key, last_sync)
        
        for doc in updated_docs:
            await self.rag.vector_db.delete_by_metadata({"source_id": doc["id"]})
            
            chunks = self.chunker.chunk(doc)
            
            for chunk in chunks:
                chunk.embedding = await self.rag.embedding.encode(chunk.content)
            
            await self.rag.vector_db.insert(chunks)
        
        self.last_sync[space_key] = datetime.now()
        
        return len(updated_docs)
```

---

## Part 7: 学习能力迭代路线

### 7.1 Phase 1：基础规则引擎（MVP）

**时间**：2-3周  
**目标**：基于固定规则运行，不包含机器学习能力

**功能**：
- 基于规则的风险评估
- 预定义的自动处理规则
- 强制升级规则
- 固定的决策阈值

**学习能力说明**：
- Phase 1 **不包含机器学习**（模式识别、反馈优化、知识库自动构建）
- 所有决策基于预定义规则和配置
- 会记录处理历史数据，为Phase 2做准备

**验收标准**：
- ✅ 能够基于规则正确处理告警
- ✅ 准确率 > 85%（基于人工标注的测试集）
- ✅ 所有状态转换正常工作
- ✅ 决策过程可解释（能输出决策依据）

**准确率定义**：
- **离线测试**：使用100个人工标注的历史告警作为测试集
  - 标注内容：正确的风险等级、应该采取的行动
  - 计算方式：(正确决策数 / 总告警数) × 100%
  - 目标：> 85%
- **线上验证**：灰度发布期间通过以下方式验证
  - 影子模式运行，记录决策但不实际执行
  - 人工抽样审查（每天抽查20条）
  - 收集用户反馈（通过Seatalk交互按钮）
  - 对比现有系统的处理结果
  - 目标：抽样准确率 > 80%，用户负面反馈率 < 15%

### 7.2 Phase 2：模式识别学习

**时间**：4-6周  
**目标**：从历史数据中识别告警模式，自动推荐处理方式

**功能**：
- 特征提取（环境、级别、文本、时间等）
- 告警聚类和模式识别
- 相似告警匹配（相似度 > 80%）
- 基于历史成功率的推荐

**数据结构**：
```go
type AlertPattern struct {
    ID          string
    Features    map[string]interface{}
    Signature   string
    
    // 统计信息
    Occurrences       int
    SuccessRate       float64
    AvgResolutionTime time.Duration
    RequiredDoDRate   float64
    
    // 推荐
    RecommendedAction string
    RecommendedSOP    *SOP
}
```

**验收标准**：
- ✅ 能够识别重复模式
- ✅ 推荐准确率 > 75%
- ✅ 相似告警匹配准确率 > 80%

### 7.3 Phase 3：反馈驱动优化

**时间**：3-4周  
**目标**：根据用户和DoD的反馈，动态调整决策阈值

**功能**：
- 收集用户和DoD的反馈
- 分析反馈模式（误判类型、频率）
- 计算最优阈值
- 渐进式调整（每次最多10%）

**反馈类型**：
- 决策是否正确
- 是否应该升级
- 响应时间是否合理
- 改进建议

**验收标准**：
- ✅ 根据反馈优化后，误判率下降 > 20%
- ✅ 阈值调整收敛（不再频繁变化）
- ✅ 用户满意度提升

### 7.4 Phase 4：知识库自动构建

**时间**：4-5周  
**目标**：从成功的处理案例中自动生成知识库条目和SOP

**功能**：
- 识别值得沉淀的案例（DoD介入 + 快速解决 + 重复出现）
- 提取关键信息（问题、原因、解决方案）
- 使用LLM生成结构化知识库条目
- 自动生成SOP（如果有明确步骤）
- 待审核状态，需要人工确认

**验收标准**：
- ✅ 自动生成的知识库条目，人工审核通过率 > 60%
- ✅ 自动生成的SOP，可执行率 > 70%
- ✅ 知识库覆盖率提升 > 30%

### 7.5 迭代总结

```
Phase 1 (2-3周): 基础规则引擎
    ↓ 验收通过
Phase 2 (4-6周): 模式识别学习
    ↓ 验收通过
Phase 3 (3-4周): 反馈驱动优化
    ↓ 验收通过
Phase 4 (4-5周): 知识库自动构建

总计：13-18周（约3-4.5个月）
```

---

## Part 8: 部署和实施

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
        - name: SEATALK_BOT_TOKEN
          valueFrom:
            secretKeyRef:
              name: dod-agent-secrets
              key: seatalk-bot-token
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

### 8.3 灰度发布策略

Phase 1 部署策略：

1. **Week 1-2**: 内部测试团队（10%流量）
   - 影子模式运行
   - 记录决策但不实际执行
   - 收集反馈和调优

2. **Week 3**: 扩大到试点团队（30%流量）
   - 开始处理低风险告警
   - 中高风险告警仍需人工确认
   - 持续监控指标

3. **Week 4**: 全量发布（100%流量）
   - 所有告警由Agent处理
   - 保留人工确认机制
   - 完整的监控和告警

每个阶段需要监控关键指标，出现问题立即回滚。

### 8.4 部署波次（渐进式接管）

**注意**：这里的"部署波次"与学习能力的"Phase 1-4"是不同的概念。

1. **部署波次 1**: 部署新系统，但不接管告警处理（仅记录日志，影子模式）
2. **部署波次 2**: 接管低风险告警（自动处理）
3. **部署波次 3**: 接管中风险告警（快速确认）
4. **部署波次 4**: 接管高风险告警（必须确认）
5. **部署波次 5**: 完全接管所有告警（包括Critical）

### 8.5 特性开关

使用特性开关控制新功能：

```go
type FeatureFlags struct {
    EnableDoDAgent      bool
    EnableAutoResolve   bool
    EnableLearning      bool
    LearningPhase       int  // 1, 2, 3, 4
}
```

### 8.6 回滚计划

如果出现以下情况，立即回滚：

- 告警处理成功率 < 70%
- 误判率 > 20%
- 系统错误率 > 5%
- DoD升级率 > 40%

---

## Part 9: 监控和可观测性

### 9.1 关键指标

| 指标 | 说明 | 目标 |
|------|------|------|
| 告警处理成功率 | 成功解决的告警比例 | > 85% |
| 平均处理时间 | 从接收到解决的平均时间 | < 10min |
| DoD升级率 | 需要DoD介入的比例 | < 20% |
| 误判率 | 错误决策的比例 | < 10% |
| 自动处理率 | 无需人工介入的比例 | > 60% |

### 9.2 监控指标实现

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

### 9.3 诊断质量追踪

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

### 9.4 日志记录

所有关键操作都需要记录结构化日志：

- 状态转换（包含原因和持续时间）
- 决策过程（风险评估、决策结果）
- ReACT循环（观察、思考、行动）
- 工具调用（参数、结果、耗时）
- 错误和异常（堆栈、上下文）

### 9.5 告警和通知

以下情况需要发送告警：

- 状态机超时（分析超时、SOP执行超时）
- 决策失败（无法评估风险、无法做出决策）
- ReACT循环异常（超过最大迭代次数、工具调用失败）
- 学习模块异常（Phase 2+：模式识别失败、反馈处理失败）

---

## Part 10: 数据模型

### 10.1 告警状态记录

```go
type AlertStateRecord struct {
    AlertID      string
    CurrentState AlertState
    StateHistory []StateHistoryEntry
    Context      *AlertContext
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type StateHistoryEntry struct {
    FromState AlertState
    ToState   AlertState
    Reason    string
    Timestamp time.Time
    Duration  time.Duration
}
```

### 10.2 决策记录

```go
type DecisionRecord struct {
    ID             string
    AlertID        string
    Timestamp      time.Time
    
    // 风险评估
    RiskLevel      RiskLevel
    RiskScore      float64
    RiskFactors    []RiskFactor
    
    // 决策结果
    Action         DecisionAction
    Confidence     float64
    Reasoning      string
    RequireConfirm bool
    
    // 反馈
    Feedback       *Feedback
}
```

### 10.3 反馈记录

```go
type Feedback struct {
    AlertID       string
    DecisionID    string
    
    // 反馈来源
    Source        string  // "user", "dod", "system"
    SourceEmail   string
    
    // 反馈内容
    IsCorrect     bool
    ShouldEscalate *bool
    ResponseTime  *time.Duration
    Suggestion    string
    
    Timestamp     time.Time
}
```

### 10.4 模式记录（Phase 2+）

```go
type AlertPattern struct {
    ID          string
    Features    map[string]interface{}
    Signature   string
    
    // 历史记录
    Occurrences int
    Resolutions []ResolutionRecord
    
    // 统计信息
    SuccessRate       float64
    AvgResolutionTime time.Duration
    RequiredDoDRate   float64
    
    // 推荐
    RecommendedAction string
    RecommendedSOP    *SOP
    
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

---

## Part 11: 配置管理

### 11.1 团队配置

每个团队可以独立配置：

```json
{
  "team_id": "dp-be",
  "dod_agent_config": {
    "enabled": true,
    
    "risk_thresholds": {
      "low_to_medium": 30,
      "medium_to_high": 60,
      "high_to_critical": 85
    },
    
    "timeouts": {
      "medium_risk_confirm": "30s",
      "high_risk_confirm": "0s",
      "sop_execution": "5m",
      "dod_response": "10m"
    },
    
    "auto_resolve_rules": [
      {
        "name": "known_database_timeout",
        "conditions": [
          {"field": "name", "operator": "contains", "value": "database timeout"},
          {"field": "has_sop", "operator": "eq", "value": true}
        ],
        "action": "execute_sop",
        "max_attempts": 3
      }
    ],
    
    "force_escalate_rules": [
      {
        "name": "production_critical",
        "conditions": [
          {"field": "env", "operator": "eq", "value": "prod"},
          {"field": "level", "operator": "eq", "value": "critical"}
        ],
        "target": "dod",
        "priority": 1
      }
    ]
  }
}
```

### 11.2 全局配置

```json
{
  "global_config": {
    "react_max_iterations": 10,
    "react_timeout": "5m",
    "llm_model": "claude-3-5-sonnet",
    "llm_temperature": 0.3,
    "enable_learning": false,
    "learning_phase": 1
  }
}
```

**配置说明**：
- `enable_learning`: Phase 1 默认为 `false`（无ML学习能力）
- `learning_phase`: 表示当前代码支持的学习阶段（1=规则引擎，2=模式识别，3=反馈优化，4=知识库构建）
- Phase 2+ 部署时将 `enable_learning` 改为 `true`

---

## Part 12: 扩展性设计

### 12.1 多部门适配

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

### 12.2 插件系统

```python
class PluginManager:
    """插件管理器"""
    
    def __init__(self):
        self._plugins: Dict[str, Plugin] = {}
    
    def register(self, plugin: Plugin):
        """注册插件"""
        self._plugins[plugin.name] = plugin
        
        for tool in plugin.get_tools():
            self.tool_registry.register(tool)
        
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

## Part 13: 安全和权限

### 13.1 操作权限

不同风险等级的操作需要不同权限：

| 操作 | 风险等级 | 需要权限 |
|------|---------|---------|
| 查询信息 | Low | 所有用户 |
| 执行SOP | Medium | Agent + 确认用户 |
| 重启服务 | High | Agent + 管理员确认 |
| 更新配置 | High | Agent + 管理员确认 |
| 升级DoD | Medium | Agent自动 |

### 13.2 审计日志

所有操作都需要记录审计日志：

- 操作类型和参数
- 执行者（Agent或用户）
- 执行时间和结果
- 影响范围

---

## Part 14: 测试策略

### 14.1 单元测试

- 状态机转换逻辑
- 风险评估算法
- 决策引擎逻辑
- ReACT循环控制
- 工具调用

### 14.2 集成测试

- 完整的告警处理流程
- 状态机与ReACT的协作
- 工具集成
- 配置加载和应用

### 14.3 端到端测试

模拟真实告警场景：

- 低风险告警自动处理
- 中风险告警快速确认
- 高风险告警人工确认
- 严重告警立即升级DoD
- 超时处理
- 失败恢复

### 14.4 压力测试

- 并发告警处理能力
- ReACT循环性能
- 工具调用延迟
- 数据库查询性能

---

## Part 15: 成本估算

### 15.1 LLM 成本

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

### 15.2 基础设施成本

| 资源 | 规格 | 月成本（估算） |
|:---|:---|:---|
| DoD Agent Pod | 2 × (1C/1G) | ~$40 |
| Chroma Vector DB | 1 × (2C/4G) + 50G SSD | ~$80 |
| Redis (缓存) | 1G | ~$30 |
| **合计** | | ~$150/月 |

---

## Part 16: 风险和缓解

### 16.1 技术风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| ReACT循环不稳定 | 高 | 中 | 严格的超时控制、最大迭代限制、完善的错误处理 |
| LLM响应延迟 | 中 | 高 | 缓存常见查询、异步处理、超时降级 |
| 决策误判 | 高 | 中 | 人工确认机制、反馈优化、渐进式部署 |
| 状态机死锁 | 高 | 低 | 超时机制、状态监控、手动干预接口 |

### 16.2 业务风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| 用户不信任自动决策 | 中 | 中 | 透明的决策过程、可解释的推理、人工确认选项 |
| DoD不满意自动升级 | 中 | 低 | 可配置的升级策略、反馈机制、人工审核 |
| 告警量激增 | 高 | 中 | 限流机制、优先级队列、降级策略 |

### 16.3 运维风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| 配置错误 | 高 | 中 | 配置校验、灰度发布、快速回滚 |
| 数据丢失 | 高 | 低 | 定期备份、主从复制、事务保证 |
| 性能下降 | 中 | 中 | 性能监控、资源预留、自动扩容 |

---

## Part 17: 成功标准

### 17.1 Phase 1 成功标准

- ✅ 所有状态转换正常工作
- ✅ 决策引擎准确率 > 85%
- ✅ 告警处理成功率 > 80%
- ✅ 平均处理时间 < 15min
- ✅ 系统稳定性 > 99.9%

### 17.2 Phase 2 成功标准

- ✅ 模式识别准确率 > 75%
- ✅ 相似告警匹配准确率 > 80%
- ✅ 推荐采纳率 > 60%
- ✅ 告警处理成功率 > 85%

### 17.3 Phase 3 成功标准

- ✅ 误判率下降 > 20%
- ✅ 用户满意度提升
- ✅ 阈值调整收敛
- ✅ 告警处理成功率 > 90%

### 17.4 Phase 4 成功标准

- ✅ 知识库条目审核通过率 > 60%
- ✅ SOP可执行率 > 70%
- ✅ 知识库覆盖率提升 > 30%
- ✅ DoD升级率下降 > 15%

---

## Part 18: 未来扩展

### 18.1 多Agent协作（未来）

当前设计是单体Agent，未来可以扩展为多Agent协作：

- **Alert Analyzer Agent**: 专门分析告警
- **SOP Executor Agent**: 专门执行SOP
- **DoD Coordinator Agent**: 专门协调DoD
- **Knowledge Builder Agent**: 专门构建知识库

### 18.2 跨团队协作（未来）

支持跨团队的告警处理和DoD协调：

- 自动识别告警涉及的多个团队
- 协调多个团队的DoD
- 跨团队的知识共享

### 18.3 预测性告警（未来）

基于历史数据和机器学习，预测可能发生的告警：

- 趋势分析
- 异常检测
- 提前预警

---

## 总结

DoD Agent 通过 AI 能力增强运维效率，核心价值：

1. **降低 MTTR**：自动诊断减少人工分析时间
2. **知识沉淀**：将专家经验转化为可检索知识
3. **标准化处理**：常见问题自动化处理流程
4. **可控性**：状态机保证流程可控、可监控、可恢复
5. **学习能力**：从历史数据和反馈中持续学习和优化
6. **可扩展**：插件化设计支持多部门复用

> 第一阶段聚焦**只读诊断 + 基础规则引擎**，验证价值后再逐步扩展学习能力和自动化操作能力。

---

## 附录

### 术语表

| 术语 | 说明 |
|------|------|
| DoD | Developer on Duty，值班开发人员 |
| ReACT | Reasoning and Acting，推理-行动循环 |
| MCP | Model Context Protocol，模型上下文协议 |
| SOP | Standard Operating Procedure，标准操作流程 |
| LLM | Large Language Model，大语言模型 |
| RAG | Retrieval-Augmented Generation，检索增强生成 |

### 参考文档

- [当前 DoD Agent 架构文档](../../DOD_AGENT_ARCHITECTURE.md)
- [ReACT 框架文档](../../NewReACT.png)
- [MCP 工具开发指南](../../mcp-tool-development-guide.md)
- [告警处理流程](../../告警消息处理流程.png)

### 变更历史

| 版本 | 日期 | 作者 | 变更说明 |
|------|------|------|---------|
| 1.0 | 2026-03-09 | AI Planner Team | v1 初始版本（纯ReACT架构） |
| 2.0 | 2026-04-03 | AI Planner Team | v2 重设计版本（状态机+ReACT混合架构） |
| 2.0-merged | 2026-04-03 | AI Planner Team | 合并v1和v2，创建完整设计文档 |

---

**文档结束**
