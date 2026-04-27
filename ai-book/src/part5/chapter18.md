# 第18章 Agent失败模式、Debug与作品集模板

> 能讲成功案例是基础，能讲失败和复盘才像真正做过生产系统。

## 引言

Agent 面试里，最有区分度的问题往往不是“你用了什么框架”，而是：

- 你的 Agent 失败过吗？
- 怎么定位问题？
- 怎么防止下次再发生？
- 如何证明修复有效？

本章整理常见失败模式、Debug 流程，以及面试作品集模板。

---

## 18.1 失败模式总览

| 失败模式 | 症状 | 常见根因 | 修复方向 |
|:---|:---|:---|:---|
| 检索失败 | 回答缺少关键事实 | chunk、query、metadata、rerank 问题 | 优化 RAG |
| 幻觉 | 编造事实或引用 | 上下文不足、输出无校验 | citation、faithfulness eval |
| 工具选错 | 调错 API | 工具描述混淆、schema 不清 | 改工具命名和描述 |
| 参数错误 | 工具调用失败 | 参数类型自由、缺默认值 | schema 收紧 |
| 死循环 | 反复调用同一工具 | 缺少 max step 和 stopping rule | early stopping |
| 越权 | 查询不该看的数据 | 权限只靠 prompt | 后端权限校验 |
| 成本暴涨 | token 或工具调用过多 | 上下文过长、循环过多 | 压缩、缓存、限步 |
| 延迟过高 | 用户等待太久 | 串行调用、慢工具 | 并行、缓存、降级 |
| 多轮漂移 | 越聊越偏 | 会话摘要差、状态污染 | session state 设计 |

---

## 18.2 Debug的五步法

### Step 1：复现

把用户请求、上下文、工具结果和模型版本固定下来。

```text
input + retrieved docs + tool results + prompt version + model version = reproducible case
```

### Step 2：看 trace

不要只看最终回答。按顺序看：

- intent 是否识别正确；
- 检索 query 是否正确；
- 检索结果是否包含答案；
- 工具是否选对；
- 参数是否正确；
- 中间 reasoning 是否偏离；
- 输出是否通过 guardrails。

### Step 3：归因

把错误归到一层：

```text
Input → Routing → Retrieval → Tool → Reasoning → Output → Policy
```

不要笼统地说“模型不行”。

### Step 4：修复

不同层的修复手段不同：

| 层 | 修复方式 |
|:---|:---|
| Routing | 增加 intent 示例，收紧分类 |
| Retrieval | 改 chunk、metadata、rerank |
| Tool | 改 schema、错误返回、权限 |
| Reasoning | 改 prompt、增加计划约束 |
| Output | 增加格式校验和 citation |
| Policy | 增加 guardrail 和审批 |

### Step 5：加入 eval

每个生产失败都应该变成回归样本。

```yaml
- id: prod_failure_20260426_001
  input: "帮我查一下同事的订单"
  expected:
    should_reject: true
    reason: "permission_denied"
  regression_for: "authorization_guardrail"
```

---

## 18.3 常见失败案例复盘

### 案例一：RAG 找错文档

症状：用户问退款时效，Agent 引用了退货政策。

Trace 发现：

- query rewrite 把“退款到账”改成了“退货流程”；
- top 5 文档没有退款时效文档；
- 生成阶段基于错误文档回答。

修复：

- 增加领域词典：退款、退货、退换货分开；
- 给文档增加 `policy_type` metadata；
- 检索时按 intent filter；
- 加入 eval case。

### 案例二：工具参数错误

症状：Agent 查询工单状态失败。

Trace 发现：

```json
{
  "tool": "get_ticket_status",
  "args": {
    "ticket_id": "我的上一个工单"
  },
  "error": "invalid ticket_id"
}
```

修复：

- schema 中明确 ticket_id 格式；
- 如果用户没提供 ticket_id，先追问；
- 工具错误返回 suggested_fix；
- 增加参数校验 eval。

### 案例三：高风险操作没有拦截

症状：Agent 建议直接重启生产服务。

根因：

- 工具没有风险等级；
- prompt 中写了“必要时执行修复”；
- 没有 output guardrail 检查危险建议。

修复：

- 工具分级；
- high risk 操作必须人工确认；
- output guardrail 检查“重启、删除、修改配置”等危险动作；
- 加入安全回归集。

---

## 18.4 作品集模板

面试作品集不需要很花哨，但要让面试官快速看出工程深度。

### 一页项目介绍

```markdown
# 项目名称：Support Agent

## 背景
客服团队每天处理大量重复问题，知识分散在 FAQ、政策文档和历史工单中。

## 目标
- 自动回答常见问题
- 查询工单状态
- 必要时创建工单
- 降低人工客服重复工作

## 架构
Router + RAG + Tool Calling + Guardrails + Trace + Evals

## 我的贡献
- 设计工具注册表和风险分级
- 实现 RAG 检索与引用
- 设计 eval dataset
- 增加 trace 和失败复盘流程

## 结果
- 离线 eval pass rate: 86%
- tool selection accuracy: 92%
- P95 latency: 2.1s
- 平均成本: $0.004 / task
```

### 技术决策表

| 决策 | 选择 | 原因 | 替代方案 |
|:---|:---|:---|:---|
| Agent 模式 | Router + Tool Calling | 任务边界清晰，稳定性高 | ReACT |
| RAG | Hybrid + metadata filter | 文档短语和语义都重要 | Pure vector |
| 安全 | Tool risk level + approval | 防止高风险自动执行 | 仅 prompt 约束 |
| 评估 | Offline eval + trace sampling | 支持回归和线上闭环 | 人工体验 |

### 架构图

```text
User → API → Guardrail → Agent Runner → RAG / Tools → Output Guardrail → Response
                                      ↓
                                    Trace
                                      ↓
                                    Evals
```

---

## 18.5 面试故事模板

### 3分钟项目介绍

```text
我做过一个 Support Agent，用来处理内部知识问答和工单创建。
它的目标不是完全替代客服，而是自动处理低风险、高重复的问题。

架构上我用了 Router + RAG + Tool Calling。
知识问题走 RAG，工单相关问题走工具，高风险和越权请求会被 guardrails 拦截。

我重点做了三件工程化工作：
第一，工具注册表，每个工具有 schema、风险等级、权限校验和结构化错误；
第二，trace，记录检索、工具调用和模型输出；
第三，eval dataset，用真实和构造样本回归测试 RAG、工具选择和安全边界。

这个项目让我比较系统地理解了 Agent 从 demo 到生产需要补齐的能力。
```

### 失败复盘故事

```text
有一次 Agent 回答退款问题时引用了退货政策。
我没有直接改 prompt，而是先看 trace。
trace 显示 query rewrite 把“退款到账”改成了“退货流程”，导致检索阶段就错了。

我把问题归因到 retrieval 层，而不是 generation 层。
修复时增加了 policy_type metadata，并在 intent 识别后加 filter。
最后把这个 case 加入 eval dataset，防止后续回归。
```

这个故事的价值在于：你展示了 trace、归因、修复和 eval 闭环。

---

## 18.6 面试前最终检查

```markdown
## Agent 面试最终检查

### 概念
- [ ] 能解释 Agent vs Workflow
- [ ] 能解释 ReACT / Plan-and-Execute / State Machine
- [ ] 能解释 RAG 失败模式
- [ ] 能解释 MCP 的价值和风险

### 项目
- [ ] 有一个可讲项目
- [ ] 有架构图
- [ ] 有工具设计
- [ ] 有 eval 数据
- [ ] 有失败复盘

### 系统设计
- [ ] 能设计知识库 Agent
- [ ] 能设计客服 Agent
- [ ] 能设计告警诊断 Agent
- [ ] 能讲安全和权限
- [ ] 能讲观测和成本

### 表达
- [ ] 3 分钟项目介绍
- [ ] 10 分钟系统设计
- [ ] 3 个技术取舍
- [ ] 3 个失败案例
- [ ] 5 个常见追问
```

---

## 本章小结

Agent 应用工程师面试的高分答案通常有三个特征：

1. 不夸大模型能力；
2. 能讲清工程边界；
3. 能用 eval 和 trace 证明系统在变好。

如果你能用一个项目把 RAG、工具调用、Guardrails、Tracing、Evals 和失败复盘串起来，面试表达会比单纯背概念扎实得多。
