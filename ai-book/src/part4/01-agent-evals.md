# 第14章 Agent Evals：从离线评估到线上质量闭环

> 没有 eval 的 Agent，只是一个看起来聪明的 demo。

## 引言

传统软件可以通过单元测试、集成测试和监控覆盖大部分行为。但 Agent 系统的输出是概率性的，执行路径会随着上下文、工具结果和模型版本变化而变化。你很难只靠“代码没有报错”判断系统是否真的完成了用户目标。

因此，Agent Evals 的核心价值是：把主观体验变成可重复、可比较、可回归的质量信号。

---

## 13.1 为什么 Agent 评估更难

普通后端接口的测试通常是：

```text
给定输入 → 执行函数 → 对比输出
```

Agent 的评估更接近：

```text
给定任务 → 规划步骤 → 调用工具 → 读取观察结果 → 多轮调整 → 输出答案
```

难点主要有五个：

| 难点 | 表现 | 评估策略 |
|:---|:---|:---|
| 输出不唯一 | 多种答案都可能正确 | 使用 rubric 而不是精确匹配 |
| 路径不唯一 | 工具调用顺序可能不同 | 评估最终目标和关键约束 |
| 多轮依赖 | 单轮正确不代表会话成功 | 做 thread-level eval |
| 工具影响大 | 错误可能来自工具或检索 | trace 分层评估 |
| 模型会变化 | 升级模型可能引入回归 | 固定回归集和版本对比 |

---

## 13.2 评估对象分层

不要只评估最终答案。生产级 Agent 至少要分四层评估。

### 1. 检索层评估

关注 RAG 是否找到了正确材料。

指标：

- Recall@K：正确文档是否出现在前 K 个结果；
- MRR：正确文档排名是否靠前；
- Metadata 命中率：服务、产品、时间范围等过滤条件是否正确；
- 引用覆盖率：答案是否引用了相关来源。

### 2. 工具层评估

关注 Agent 是否选择了正确工具，并正确传参。

指标：

- tool selection accuracy；
- argument accuracy；
- tool error recovery rate；
- unnecessary tool call rate；
- high-risk tool approval rate。

### 3. 任务层评估

关注用户目标是否完成。

指标：

- task success rate；
- answer correctness；
- format compliance；
- policy compliance；
- escalation correctness。

### 4. 会话层评估

关注多轮互动是否整体成功。

指标：

- goal completion；
- conversation resolution rate；
- user frustration signals；
- repeated question rate；
- handoff quality。

---

## 13.3 构建离线评估集

离线 eval dataset 是回归测试的基础。不要只收集“正常问题”，一定要覆盖边界情况。

### 数据集结构

```yaml
- id: alert_cpu_001
  user_input: "order-service CPU 使用率 92%，请帮我诊断"
  expected_behavior:
    - 查询最近部署
    - 查询 CPU 指标趋势
    - 搜索 order-service 相关错误日志
    - 给出风险等级
  must_not:
    - 自动重启生产服务
    - 编造不存在的部署记录
  reference_docs:
    - runbook/order-service/cpu-high.md
  expected_final:
    type: diagnosis_report
    severity: warning
```

### 数据来源

优先级从高到低：

1. 生产 trace 中的真实失败案例；
2. 人工标注的高价值任务；
3. 历史工单和客服记录；
4. 专家构造的边界 case；
5. LLM 生成的补充样本。

### 样本分类

一个健康的数据集应该包含：

- happy path；
- edge case；
- adversarial input；
- missing context；
- tool failure；
- permission denied；
- ambiguous request；
- multi-turn correction。

---

## 13.4 LLM-as-Judge 的正确用法

LLM-as-Judge 不是让模型随便打分，而是给它明确 rubric。

### 不好的 judge prompt

```text
请判断这个回答好不好，给 1-5 分。
```

问题：标准模糊，分数不可解释。

### 更好的 judge prompt

```text
你是 Agent 质量评估员。请根据以下 rubric 评估回答。

任务：
用户要求诊断生产告警。

评分维度：
1. 根因分析是否基于给定证据，而不是猜测；
2. 是否引用了相关日志、指标或知识库；
3. 是否给出可执行的下一步；
4. 是否避免执行高风险操作；
5. 输出格式是否符合诊断报告模板。

输出 JSON：
{
  "score": 1-5,
  "passed": true/false,
  "reasons": ["..."],
  "failure_category": "retrieval|tool|reasoning|format|safety|none"
}
```

### 校准 judge

LLM-as-Judge 必须用人工样本校准：

- 抽样 50-100 条人工评分；
- 对比 judge 与人工的一致率；
- 找出分歧样本；
- 调整 rubric；
- 对高风险任务保留人工复核。

---

## 13.5 Trace驱动的评估

Agent 的错误经常藏在中间步骤，而不是最终答案里。Trace 应该记录：

```json
{
  "trace_id": "tr_123",
  "task_id": "alert_cpu_001",
  "model": "gpt-4.1",
  "prompt_version": "diagnosis-v3",
  "steps": [
    {
      "type": "tool_call",
      "tool": "prometheus_query",
      "args": {"query": "cpu_usage{service='order-service'}"},
      "latency_ms": 420,
      "status": "success"
    },
    {
      "type": "tool_call",
      "tool": "log_search",
      "args": {"service": "order-service", "range": "30m"},
      "latency_ms": 810,
      "status": "success"
    }
  ],
  "final_answer": "...",
  "cost_usd": 0.08,
  "latency_ms": 12400
}
```

基于 trace 可以做三类评估：

- 单步评估：工具是否选对；
- 路径评估：是否遗漏关键步骤；
- 结果评估：最终答案是否满足任务。

---

## 13.6 线上评估闭环

上线后不能只看错误率。Agent 的质量问题往往不会抛异常，而是“回答得很像对，但实际没解决问题”。

线上闭环：

```text
生产请求
  ↓
记录 trace
  ↓
采样进入评估队列
  ↓
自动 judge + 人工复核
  ↓
归类失败模式
  ↓
加入 eval dataset
  ↓
回归测试新版本
```

### 线上指标

| 指标 | 含义 |
|:---|:---|
| Task Success Rate | 用户目标完成率 |
| Escalation Rate | 转人工比例 |
| First Contact Resolution | 首次解决率 |
| Tool Error Rate | 工具调用失败率 |
| Hallucination Rate | 无依据回答比例 |
| Cost per Task | 单任务成本 |
| P95 Latency | 用户体验延迟 |

### 质量报警

应该为以下情况设置报警：

- task success rate 连续下降；
- hallucination rate 升高；
- tool error rate 升高；
- 单任务成本异常；
- 高风险工具调用增加；
- 某类用户问题大量转人工。

---

## 13.7 面试回答模板

问题：你如何评估一个 Agent 系统？

可以这样回答：

```text
我会分层评估。

第一层评估 RAG，确认正确文档是否被召回，引用是否准确。
第二层评估工具调用，确认工具选择、参数和错误恢复是否正确。
第三层评估任务结果，看用户目标是否完成、格式是否符合要求、是否违反业务规则。
第四层评估多轮会话，看整段会话是否真正解决问题。

离线阶段我会维护 golden dataset 和回归集。
线上阶段我会采样 trace，用自动 judge 和人工复核结合，把失败样本持续加入评估集。
每次改 prompt、模型、retriever 或工具 schema，都必须跑回归评估。
```

---

## 13.8 检查清单

```markdown
## Agent Evals 检查清单

### 数据集
- [ ] 有 happy path
- [ ] 有 edge cases
- [ ] 有 adversarial cases
- [ ] 有工具失败样本
- [ ] 有多轮会话样本

### 指标
- [ ] RAG recall
- [ ] tool selection accuracy
- [ ] task success rate
- [ ] hallucination rate
- [ ] cost per task
- [ ] P95 latency

### 流程
- [ ] 每次发布前跑离线 eval
- [ ] 线上 trace 进入采样队列
- [ ] 失败样本进入回归集
- [ ] judge prompt 有人工校准
- [ ] eval 结果按版本记录
```

---

## 本章小结

Agent Evals 的核心不是“打一个分”，而是建立持续改进闭环。

好的评估体系应该回答：

- 哪一层出了问题；
- 新版本是否更好；
- 失败是否可复现；
- 线上质量是否退化；
- 业务目标是否真的改善。

下一章我们进入 Guardrails 与 Agent 安全。Evals 负责发现问题，Guardrails 负责在关键边界提前拦住问题。
