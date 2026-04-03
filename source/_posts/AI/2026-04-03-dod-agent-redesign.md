# DoD Agent 重设计规范

**日期**: 2026-04-03  
**作者**: AI Planner Team  
**版本**: 1.0  
**状态**: Ready for Planning

## 0. 规划范围

本规范涵盖 DoD Agent 的完整设计，包括4个学习能力阶段和5个部署波次。

**首次实施范围**：**Phase 1 - 基础规则引擎（MVP）**
- 状态机 + ReACT 混合工作流
- 基于规则的决策引擎
- 分级自主决策（Low/Medium/High/Critical）
- 可配置策略
- 完整的监控和日志

**后续阶段**：Phase 2-4（模式识别、反馈优化、知识库构建）将作为独立的迭代项目，各自有独立的计划和实施周期。

---

## 1. 执行摘要

### 1.1 项目背景

当前的 DoD Agent 是一个被动的工具服务，仅提供值班人员信息查询功能。随着系统复杂度增加和告警量增长，需要一个更智能、更主动的告警处理系统。

### 1.2 设计目标

重新设计 DoD Agent 为一个**事件驱动的智能协调型 Agent**，具备以下能力：

- **智能分析**：自动分析告警原因，协调多个子系统（告警、Jira、Seatalk、知识库）
- **自主决策**：基于风险等级和配置策略，自动决定处理方式（自动处理/人工确认/升级DoD）
- **可控性**：通过状态机保证流程可控、可监控、可恢复
- **学习能力**：从历史数据和反馈中持续学习和优化

### 1.3 核心架构选择

**方案 A：增强型 ReACT Agent**

- 基于现有 ReACT 框架，增加状态机管理和决策引擎
- 单体 Agent + 工具调用模式
- 状态机 + ReACT 混合工作流
- 分级自主决策 + 可配置策略
- 4阶段渐进式学习能力演进

---

## 2. 系统架构

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    告警输入层                                │
│  (Webhook, Seatalk, 定时任务)                               │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                  DoD Agent 核心                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 状态机控制器 │  │ ReACT 引擎   │  │ 决策引擎     │      │
│  │ (Lifecycle)  │◄─┤ (智能分析)   │◄─┤ (分级策略)   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         │                  │                  │              │
│         └──────────────────┼──────────────────┘              │
│                            ▼                                 │
│                   ┌──────────────┐                           │
│                   │  工具调用层  │                           │
│                   │  (MCP Tools) │                           │
│                   └──────────────┘                           │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                    工具集成层                                │
│  [告警API] [知识库] [Jira] [Seatalk] [SOP执行器] [DoD查询]│
└─────────────────────────────────────────────────────────────┘
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

### 2.3 处理流程

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

## 3. 状态机设计

### 3.1 状态定义

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

### 3.2 状态转换

每个状态转换包含：
- **From/To**：源状态和目标状态
- **Condition**：转换条件函数
- **Action**：转换时执行的动作
- **Timeout**：状态超时时间

#### 3.2.1 完整状态转换表

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
| StateWaitingConfirm | StateDoDNotified | 超时(Critical，虽然通常不经过此状态) | 升级DoD | 10min |
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

### 3.3 超时处理策略

**超时语义说明**：
- Timeout 表示**进入该状态后的最大停留时间**
- 超时后触发状态转换或降级处理
- 状态转换表（§3.2.1）中的 Timeout 列表示**目标状态**的超时时间

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

### 3.4 状态持久化

所有状态变更都会持久化到数据库，包括：
- 当前状态
- 状态历史（包含每个状态的持续时间）
- 完整的上下文信息
- 失败原因和重试次数

---

## 4. 决策引擎设计

### 4.1 风险评估模型

#### 4.1.1 风险等级

```go
type RiskLevel int

const (
    RiskLow      RiskLevel = 1  // 自动处理
    RiskMedium   RiskLevel = 2  // 快速确认（30秒超时）
    RiskHigh     RiskLevel = 3  // 必须确认
    RiskCritical RiskLevel = 4  // 立即升级DoD
)
```

#### 4.1.2 风险因素

风险评估基于以下因素的加权计算：

| 因素 | 权重 | 说明 |
|------|------|------|
| 环境 (environment) | 30% | 生产环境风险更高 |
| 严重程度 (severity) | 25% | Critical级别需要立即关注 |
| 影响范围 (impact_scope) | 20% | 多市场影响风险更高 |
| 历史模式 (historical_pattern) | 15% | 重复告警可能有已知解决方案 |
| 时间因素 (time_factor) | 10% | 高峰期风险更高 |

#### 4.1.3 风险阈值

| 分数范围 | 风险等级 | 处理方式 |
|---------|---------|---------|
| 0-30 | Low | 自动处理 |
| 31-60 | Medium | 建议+快速确认（30s超时） |
| 61-85 | High | 必须人工确认 |
| 86-100 | Critical | 立即升级DoD |

### 4.2 决策策略配置

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

### 4.3 决策流程

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

## 5. ReACT 引擎集成

### 5.1 状态感知的 ReACT

ReACT 引擎与状态机紧密集成：

- **状态约束**：每个状态只允许调用特定的工具
- **超时控制**：ReACT 循环受状态超时限制
- **状态回调**：ReACT 可以触发状态转换

### 5.2 状态与工具的映射

| 状态 | 允许的工具 | ReACT模式 |
|------|-----------|----------|
| StateAnalyzing | search_knowledge_base, query_alert_history, analyze_logs, check_metrics, search_similar_alerts | 完整ReACT循环 |
| StateAutoResolving | execute_sop, restart_service, clear_cache, update_config | 限制ReACT（仅执行工具） |
| StateExecutingSOP | execute_sop, check_sop_status, verify_resolution | 限制ReACT（仅SOP相关） |
| StateDoDNotified | get_dod_info, send_seatalk_message, create_jira_ticket | 限制ReACT（仅通知相关） |
| StateWaitingConfirm | 不允许工具调用 | 无ReACT（等待外部输入） |
| 其他状态 | 不允许工具调用 | 无ReACT |

**说明**：
- `StateExecutingSOP` 允许有限的工具调用，用于执行和验证SOP
- `StateWaitingConfirm` 完全由外部交互驱动（Seatalk按钮回调），不运行ReACT
- ReACT引擎在不同状态有不同的约束级别

### 5.3 Prompt 工程

每个状态有专门的 Prompt 模板，包括：
- 当前状态和告警信息
- 状态特定的目标和任务
- 可用工具列表和使用说明
- 推理格式要求
- 历史记录（观察、思考、行动）

### 5.4 ReACT 循环控制

- **最大迭代次数**：10次（防止无限循环）
- **超时控制**：继承状态超时设置
- **提前终止**：当得出最终结论时停止

---

## 6. 学习能力迭代路线

### 6.1 Phase 1：基础规则引擎（MVP）

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
- 能够基于规则正确处理告警
- 准确率 > 85%（基于人工标注的测试集）
- 所有状态转换正常工作
- 决策过程可解释（能输出决策依据）

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

### 6.2 Phase 2：模式识别学习

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
- 能够识别重复模式
- 推荐准确率 > 75%
- 相似告警匹配准确率 > 80%

### 6.3 Phase 3：反馈驱动优化

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
- 根据反馈优化后，误判率下降 > 20%
- 阈值调整收敛（不再频繁变化）
- 用户满意度提升

### 6.4 Phase 4：知识库自动构建

**时间**：4-5周  
**目标**：从成功的处理案例中自动生成知识库条目和SOP

**功能**：
- 识别值得沉淀的案例（DoD介入 + 快速解决 + 重复出现）
- 提取关键信息（问题、原因、解决方案）
- 使用LLM生成结构化知识库条目
- 自动生成SOP（如果有明确步骤）
- 待审核状态，需要人工确认

**验收标准**：
- 自动生成的知识库条目，人工审核通过率 > 60%
- 自动生成的SOP，可执行率 > 70%
- 知识库覆盖率提升 > 30%

### 6.5 迭代总结

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

## 7. 数据模型

### 7.1 告警状态记录

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

### 7.2 决策记录

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

### 7.3 反馈记录

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

### 7.4 模式记录（Phase 2+）

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

## 8. 工具集成

### 8.1 现有工具复用

以下现有 MCP 工具将被复用：

- `get_dod_info` - 获取DoD信息
- `get_dod_by_team_id` - 按团队ID查询DoD
- `get_dod_by_sub_team_name` - 按子团队名称查询DoD
- `send_seatalk_message` - 发送Seatalk消息
- `create_jira_ticket` - 创建Jira工单
- `search_knowledge_base` - 搜索知识库
- `execute_sop` - 执行SOP

### 8.2 新增工具

需要新增以下工具：

#### 8.2.1 告警分析工具

```go
// query_alert_history - 查询告警历史
type QueryAlertHistoryInput struct {
    AlertName   string
    Team        string
    TimeRange   time.Duration
    Limit       int
}

// analyze_logs - 分析日志
type AnalyzeLogsInput struct {
    ServiceName string
    TimeRange   time.Duration
    Keywords    []string
}

// check_metrics - 检查监控指标
type CheckMetricsInput struct {
    ServiceName string
    MetricName  string
    TimeRange   time.Duration
}

// search_similar_alerts - 搜索相似告警
type SearchSimilarAlertsInput struct {
    Alert       *Alert
    Threshold   float64  // 相似度阈值
    Limit       int
}
```

#### 8.2.2 操作工具

```go
// restart_service - 重启服务
type RestartServiceInput struct {
    ServiceName string
    Environment string
    WaitTime    time.Duration
}

// clear_cache - 清理缓存
type ClearCacheInput struct {
    CacheType   string
    CacheKey    string
}

// update_config - 更新配置
type UpdateConfigInput struct {
    ServiceName string
    ConfigKey   string
    ConfigValue string
}
```

---

## 9. 配置管理

### 9.1 团队配置

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

### 9.2 全局配置

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

## 10. 监控和可观测性

### 10.1 关键指标

| 指标 | 说明 | 目标 |
|------|------|------|
| 告警处理成功率 | 成功解决的告警比例 | > 85% |
| 平均处理时间 | 从接收到解决的平均时间 | < 10min |
| DoD升级率 | 需要DoD介入的比例 | < 20% |
| 误判率 | 错误决策的比例 | < 10% |
| 自动处理率 | 无需人工介入的比例 | > 60% |

### 10.2 日志记录

所有关键操作都需要记录结构化日志：

- 状态转换（包含原因和持续时间）
- 决策过程（风险评估、决策结果）
- ReACT循环（观察、思考、行动）
- 工具调用（参数、结果、耗时）
- 错误和异常（堆栈、上下文）

### 10.3 告警和通知

以下情况需要发送告警：

- 状态机超时（分析超时、SOP执行超时）
- 决策失败（无法评估风险、无法做出决策）
- ReACT循环异常（超过最大迭代次数、工具调用失败）
- 学习模块异常（Phase 2+：模式识别失败、反馈处理失败）

---

## 11. 安全和权限

### 11.1 操作权限

不同风险等级的操作需要不同权限：

| 操作 | 风险等级 | 需要权限 |
|------|---------|---------|
| 查询信息 | Low | 所有用户 |
| 执行SOP | Medium | Agent + 确认用户 |
| 重启服务 | High | Agent + 管理员确认 |
| 更新配置 | High | Agent + 管理员确认 |
| 升级DoD | Medium | Agent自动 |

### 11.2 审计日志

所有操作都需要记录审计日志：

- 操作类型和参数
- 执行者（Agent或用户）
- 执行时间和结果
- 影响范围

---

## 12. 测试策略

### 12.1 单元测试

- 状态机转换逻辑
- 风险评估算法
- 决策引擎逻辑
- ReACT循环控制
- 工具调用

### 12.2 集成测试

- 完整的告警处理流程
- 状态机与ReACT的协作
- 工具集成
- 配置加载和应用

### 12.3 端到端测试

模拟真实告警场景：

- 低风险告警自动处理
- 中风险告警快速确认
- 高风险告警人工确认
- 严重告警立即升级DoD
- 超时处理
- 失败恢复

### 12.4 压力测试

- 并发告警处理能力
- ReACT循环性能
- 工具调用延迟
- 数据库查询性能

---

## 13. 部署策略

### 13.1 灰度发布

Phase 1 部署策略：

1. **Week 1-2**: 内部测试团队（10%流量）
2. **Week 3**: 扩大到试点团队（30%流量）
3. **Week 4**: 全量发布（100%流量）

每个阶段需要监控关键指标，出现问题立即回滚。

### 13.2 特性开关

使用特性开关控制新功能：

```go
type FeatureFlags struct {
    EnableDoDAgent      bool
    EnableAutoResolve   bool
    EnableLearning      bool
    LearningPhase       int  // 1, 2, 3, 4
}
```

### 13.3 回滚计划

如果出现以下情况，立即回滚：

- 告警处理成功率 < 70%
- 误判率 > 20%
- 系统错误率 > 5%
- DoD升级率 > 40%

---

## 14. 迁移计划

### 14.1 数据迁移

现有 DoD 配置需要迁移到新的数据模型：

```go
// 旧配置
type OldDoDConfig struct {
    TeamIDs           []DoDTeamInfo
    RelatedDoDTeamIDs []DoDTeamInfo
}

// 新配置
type NewDoDAgentConfig struct {
    DoDConfig      OldDoDConfig  // 保留
    DecisionPolicy DecisionPolicy // 新增
    FeatureFlags   FeatureFlags   // 新增
}
```

### 14.2 API兼容性

保持现有 DoD 查询 API 的兼容性：

- `GetDodOfTeam()` - 保持不变
- `GetDoDConfig()` - 保持不变
- `SendDODMentionMessage()` - 保持不变

新增 API：

- `ProcessAlert()` - 处理告警的主入口
- `GetAlertState()` - 查询告警状态
- `ProvideFeedback()` - 提供反馈（用于Phase 3学习优化）

### 14.3 渐进式迁移（部署波次）

**注意**：这里的"部署波次"与学习能力的"Phase 1-4"是不同的概念。

1. **部署波次 1**: 部署新系统，但不接管告警处理（仅记录日志，影子模式）
2. **部署波次 2**: 接管低风险告警（自动处理）
3. **部署波次 3**: 接管中风险告警（快速确认）
4. **部署波次 4**: 接管高风险告警（必须确认）
5. **部署波次 5**: 完全接管所有告警（包括Critical）

---

## 15. 成功标准

### 15.1 Phase 1 成功标准

- ✅ 所有状态转换正常工作
- ✅ 决策引擎准确率 > 85%
- ✅ 告警处理成功率 > 80%
- ✅ 平均处理时间 < 15min
- ✅ 系统稳定性 > 99.9%

### 15.2 Phase 2 成功标准

- ✅ 模式识别准确率 > 75%
- ✅ 相似告警匹配准确率 > 80%
- ✅ 推荐采纳率 > 60%
- ✅ 告警处理成功率 > 85%

### 15.3 Phase 3 成功标准

- ✅ 误判率下降 > 20%
- ✅ 用户满意度提升
- ✅ 阈值调整收敛
- ✅ 告警处理成功率 > 90%

### 15.4 Phase 4 成功标准

- ✅ 知识库条目审核通过率 > 60%
- ✅ SOP可执行率 > 70%
- ✅ 知识库覆盖率提升 > 30%
- ✅ DoD升级率下降 > 15%

---

## 16. 风险和缓解

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

## 17. 未来扩展

### 17.1 多Agent协作（未来）

当前设计是单体Agent，未来可以扩展为多Agent协作：

- **Alert Analyzer Agent**: 专门分析告警
- **SOP Executor Agent**: 专门执行SOP
- **DoD Coordinator Agent**: 专门协调DoD
- **Knowledge Builder Agent**: 专门构建知识库

### 17.2 跨团队协作（未来）

支持跨团队的告警处理和DoD协调：

- 自动识别告警涉及的多个团队
- 协调多个团队的DoD
- 跨团队的知识共享

### 17.3 预测性告警（未来）

基于历史数据和机器学习，预测可能发生的告警：

- 趋势分析
- 异常检测
- 提前预警

---

## 18. 参考文档

- [当前 DoD Agent 架构文档](../../DOD_AGENT_ARCHITECTURE.md)
- [ReACT 框架文档](../../NewReACT.png)
- [MCP 工具开发指南](../../mcp-tool-development-guide.md)
- [告警处理流程](../../告警消息处理流程.png)

---

## 19. 附录

### 19.1 术语表

| 术语 | 说明 |
|------|------|
| DoD | Developer on Duty，值班开发人员 |
| ReACT | Reasoning and Acting，推理-行动循环 |
| MCP | Model Context Protocol，模型上下文协议 |
| SOP | Standard Operating Procedure，标准操作流程 |
| LLM | Large Language Model，大语言模型 |

### 19.2 变更历史

| 版本 | 日期 | 作者 | 变更说明 |
|------|------|------|---------|
| 1.0 | 2026-04-03 | AI Planner Team | 初始版本 |

---

**文档结束**
