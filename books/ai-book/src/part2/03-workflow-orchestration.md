# 第7章 Agent 工作流、状态机与多 Agent 编排

> "Alone we can do so little; together we can do so much." —— Helen Keller

## 引言

单个 Agent 能力有限。复杂任务往往需要多个 Agent 协作完成——有的擅长分析，有的擅长执行，有的擅长审查。

本章将深入探讨多 Agent 协作的设计模式、工作流编排策略，以及如何避免协作中的常见陷阱。

---

## 7.1 为什么需要多Agent协作？

### 单 Agent 的局限性

**局限 1：上下文窗口限制**

单个 Agent 的上下文会随着对话增长而膨胀，最终超出限制或导致性能下降。

**局限 2：能力泛化困难**

一个 Agent 很难同时擅长所有任务：
- 分析型任务需要深度推理
- 执行型任务需要精确操作
- 审查型任务需要批判性思维

**局限 3：缺少制衡机制**

单个 Agent 容易陷入：
- 确认偏误（只寻找支持自己观点的证据）
- 过度自信（高估自己的判断）
- 路径依赖（沿着最初的思路走到底）

### 多 Agent 协作的优势

```text
┌─────────────────────────────────────────────────────────────┐
│                  多 Agent 协作的价值                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. 职责分离（Separation of Concerns）                       │
│     - 每个 Agent 专注自己的领域                              │
│     - 减少单个 Agent 的复杂度                                │
│                                                             │
│  2. 并行执行（Parallel Execution）                           │
│     - 多个 Agent 同时工作                                    │
│     - 大幅缩短总体执行时间                                   │
│                                                             │
│  3. 互相审查（Peer Review）                                  │
│     - Writer Agent 生成内容                                  │
│     - Reviewer Agent 审查质量                                │
│     - 形成制衡机制                                           │
│                                                             │
│  4. 上下文隔离（Context Isolation）                          │
│     - 每个 Agent 在独立的上下文中工作                         │
│     - 避免上下文污染                                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 7.2 多Agent协作的核心模式

### 模式 1：Sequential（顺序执行）

最简单的协作模式，Agent 按顺序依次执行。

```text
┌─────────────────────────────────────────────────────────────┐
│                   Sequential Pattern                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Input → Agent 1 → Agent 2 → Agent 3 → Output               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**适用场景：**
- 任务有明确的先后顺序
- 后续步骤依赖前面的结果

**示例：文章写作流水线**
```text
Outline Agent → Writer Agent → Editor Agent → Publisher Agent
```

**实现示例：**
```python
class SequentialWorkflow:
    def __init__(self, agents: List[Agent]):
        self.agents = agents

    def run(self, initial_input: str):
        result = initial_input

        for agent in self.agents:
            result = agent.run(result)

        return result

# 使用
workflow = SequentialWorkflow([
    OutlineAgent(),
    WriterAgent(),
    EditorAgent(),
    PublisherAgent()
])

article = workflow.run("写一篇关于 AI Agent 的文章")
```

### 模式 2：Parallel（并行执行）

多个 Agent 同时工作，最后汇总结果。

```text
┌─────────────────────────────────────────────────────────────┐
│                    Parallel Pattern                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│            ┌──────► Agent 1 ──────┐                         │
│            │                       │                         │
│  Input ────┼──────► Agent 2 ──────┼────► Aggregator → Output│
│            │                       │                         │
│            └──────► Agent 3 ──────┘                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**适用场景：**
- 任务可以独立执行
- 需要从多个角度分析问题

**示例：多角度代码审查**
```text
Security Reviewer (安全审查)
Performance Reviewer (性能审查)  → Aggregator → 综合报告
Code Quality Reviewer (代码质量审查)
```

**实现示例：**
```python
import asyncio

class ParallelWorkflow:
    def __init__(self, agents: List[Agent]):
        self.agents = agents

    async def run(self, input: str):
        # 并行执行所有 Agent
        tasks = [agent.run_async(input) for agent in self.agents]
        results = await asyncio.gather(*tasks)

        # 汇总结果
        return self.aggregate(results)

    def aggregate(self, results: List[str]) -> str:
        # 合并所有 Agent 的输出
        return "\n\n".join(results)

# 使用
workflow = ParallelWorkflow([
    SecurityReviewer(),
    PerformanceReviewer(),
    CodeQualityReviewer()
])

review = asyncio.run(workflow.run(code_snippet))
```

### 模式 3：Writer-Reviewer（写作-审查）

一个 Agent 生成内容，另一个 Agent 审查并提供反馈，形成迭代循环。

```text
┌─────────────────────────────────────────────────────────────┐
│                 Writer-Reviewer Pattern                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Input → Writer → Reviewer → 评估                            │
│             ↑         │                                      │
│             │         │ 反馈                                 │
│             └─────────┘                                      │
│                                                             │
│           （迭代直到通过审查）                                │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**适用场景：**
- 需要高质量输出
- 可以接受多轮迭代

**示例：代码生成与审查**
```text
Code Writer Agent: 生成代码
Code Reviewer Agent: 审查代码（bug、性能、风格）
如果审查不通过 → Writer 根据反馈修改
```

**实现示例：**
```python
class WriterReviewerWorkflow:
    def __init__(self, writer: Agent, reviewer: Agent, max_iterations: int = 3):
        self.writer = writer
        self.reviewer = reviewer
        self.max_iterations = max_iterations

    def run(self, input: str):
        content = self.writer.run(input)

        for i in range(self.max_iterations):
            review = self.reviewer.run(content)

            # 检查是否通过审查
            if review["approved"]:
                return content

            # 根据反馈修改
            feedback = review["feedback"]
            content = self.writer.run(f"修改以下内容：\n{content}\n\n反馈：\n{feedback}")

        return content

# 使用
workflow = WriterReviewerWorkflow(
    writer=CodeWriter(),
    reviewer=CodeReviewer(),
    max_iterations=3
)

code = workflow.run("实现一个快速排序算法")
```

### 模式 4：Hierarchical（层级协调）

有一个 Coordinator Agent 负责任务分解和结果汇总，其他 Worker Agent 执行具体任务。

```text
┌─────────────────────────────────────────────────────────────┐
│                  Hierarchical Pattern                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                  ┌─────────────┐                            │
│                  │ Coordinator │                            │
│                  │   (协调者)  │                            │
│                  └──────┬──────┘                            │
│                         │                                    │
│          ┌──────────────┼──────────────┐                    │
│          │              │              │                    │
│    ┌─────▼────┐   ┌────▼─────┐  ┌────▼─────┐              │
│    │ Worker 1 │   │ Worker 2 │  │ Worker 3 │              │
│    └──────────┘   └──────────┘  └──────────┘              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**适用场景：**
- 复杂任务需要分解
- 需要协调多个子任务

**示例：研究报告生成**
```text
Coordinator: 分解任务、分配给 Worker、汇总结果
Worker 1: 搜索相关文献
Worker 2: 分析数据
Worker 3: 生成图表
```

**实现示例：**
```python
class HierarchicalWorkflow:
    def __init__(self, coordinator: Agent, workers: List[Agent]):
        self.coordinator = coordinator
        self.workers = workers

    def run(self, input: str):
        # 1. Coordinator 分解任务
        plan = self.coordinator.run(f"分解任务：{input}")
        tasks = self.parse_plan(plan)

        # 2. Workers 并行执行子任务
        results = []
        for task in tasks:
            worker = self.assign_worker(task)
            result = worker.run(task)
            results.append(result)

        # 3. Coordinator 汇总结果
        final_result = self.coordinator.run(f"汇总结果：\n{results}")
        return final_result

# 使用
workflow = HierarchicalWorkflow(
    coordinator=CoordinatorAgent(),
    workers=[
        LiteratureSearchAgent(),
        DataAnalysisAgent(),
        VisualizationAgent()
    ]
)

report = workflow.run("生成AI Agent市场研究报告")
```

---

## 7.3 Git Worktrees：并行工作的基础设施

Claude Code 通过 Git Worktrees 实现多个 Agent 实例的并行运行。

### 为什么需要并行？

单个 Claude Code session 的工作模式：

```text
你给任务 → Claude 执行（2-5分钟）→ 你 review → 给下一个任务
         等待                                等待
```

大量时间浪费在等待上。

多个 session 并行：

```text
Session 1: 执行任务 A
Session 2: 执行任务 B    你同时 review 多个结果
Session 3: 执行任务 C
Session 4: 执行任务 D
Session 5: 执行任务 E
```

等待时间几乎降到零。

### Worktree 操作

```bash
# 启动一个在独立 worktree 中运行的 Claude session
claude --worktree

# 在 Tmux 会话中启动（可以后台运行）
claude --worktree --tmux

# 设置 shell 别名快速跳转
alias za="tmux select-window -t claude:0"
alias zb="tmux select-window -t claude:1"
alias zc="tmux select-window -t claude:2"
```

每次运行 `claude --worktree`，Claude Code 会自动：
1. 创建一个新的 worktree
2. 切到一个新分支
3. 在隔离环境中工作

### 使用场景

**场景 1：独立功能并行开发**
```bash
# Session A: 实现用户认证
# Session B: 实现支付集成
# Session C: 实现订单管理
# Session D: 实现通知系统
# Session E: 编写单元测试
```

**场景 2：大规模重构**
```bash
# 批量迁移 50 个文件from JavaScript 到 TypeScript
for file in $(cat files-to-migrate.txt); do
  claude -p "Migrate $file from JS to TS" \
    --worktree &
done
```

50 个 Claude 实例并行运行，几分钟完成原本需要一整天的工作。

---

## 7.4 Subagents：专家协作

### Subagents vs 并行 Sessions

| 维度 | 并行 Sessions | Subagents |
|------|--------------|-----------|
| **用途** | 互不相关的独立任务 | 当前任务中的子任务 |
| **上下文** | 各自独立 | 继承主 Agent 的部分上下文 |
| **协调** | 人工协调 | 主 Agent 自动协调 |
| **适用场景** | 5 个独立 feature | 一个复杂 feature 的 5 个环节 |

### 定义 Subagent

```markdown
# .claude/agents/security-reviewer.md
---
name: Security Reviewer
tools: [Read, Grep]  # 只读权限，不能改代码
model: opus-4.6      # 使用推理能力更强的模型
---

你是一个安全审查专家。审查代码时重点关注：
1. 认证和授权逻辑
2. 敏感数据处理
3. SQL 注入风险
4. XSS 漏洞
5. CSRF 防护

发现问题时，给出具体的修复建议和代码示例。
```

### 调用 Subagent

```text
> 请审查 src/auth/ 目录下的代码，使用 security-reviewer subagent
```

Claude Code 会：
1. 启动一个新的 subagent
2. 传递必要的上下文
3. 让 subagent 执行审查
4. 将结果返回给主 Agent

### Subagents 的核心价值：独立上下文

每个 subagent 运行在自己的上下文窗口中，不消耗主 session 的上下文空间。

当主 session 的对话已经很长、上下文快要满时，调用 subagent 处理子任务，相当于开了一个新的"思考空间"。

---

## 7.5 Agent Teams：自动协调

Agent Teams 是 Claude Code 最强大的协作模式，核心理念：**不是你来协调多个 agent，而是让 agent 自己协调**。

### Writer/Reviewer 模式

```text
1. Writer Agent 写代码
   - 负责实现功能，按照需求写代码、跑测试

2. Reviewer Agent 审代码
   - review Writer 的输出，指出问题、建议改进

3. Writer 根据反馈修改
   - 收到 review 意见后改进代码，形成迭代循环
```

这个模式比单个 agent 写代码好不少。原因和人类团队一样：写代码的人容易陷入自己的思路，审代码的人能从不同角度发现问题。

### Coordinator Mode：四阶段协调

复杂任务会自动走四个阶段：

**1. Research（调研）**

多个 worker 并行调查代码库：
```text
Worker 1: 搜索认证相关代码
Worker 2: 搜索数据库操作代码
Worker 3: 搜索 API 接口定义
```

**2. Synthesis（综合）**

Coordinator 综合发现，生成规格说明：
```text
基于调研结果，这个功能需要：
- 修改 3 个文件
- 新增 2 个 API 端点
- 更新 1 个数据库表
- 编写 5 个测试用例
```

**3. Implementation（实现）**

Worker 按规格做精准修改：
```text
Worker 1: 修改 auth.go
Worker 2: 修改 api.go
Worker 3: 更新 schema.sql
Worker 4: 编写测试
```

**4. Verification（验证）**

验证结果，确保正确性：
```text
- 运行测试
- 检查 linter
- 验证 API 响应
```

### 自动判断协调模式

你不需要手动配置协调模式，Agent Teams 会根据任务复杂度自动判断：

- 简单任务：单个 Agent 直接执行
- 中等任务：Writer/Reviewer 模式
- 复杂任务：Coordinator Mode（四阶段）

---

## 7.6 Fan-out批处理：人海战术的AI版

### 非交互模式

```bash
# 非交互模式执行单个任务
claude -p "把这个文件从 JavaScript 迁移到 TypeScript"

# 批量迁移一批文件
for file in $(cat files-to-migrate.txt); do
  claude -p "Migrate $file from JS to TS" \
    --allowedTools "Edit,Bash(git commit *)" &
done
```

注意末尾的 `&`：这让每个 Claude 实例在后台并行运行。

如果有 50 个文件要迁移，50 个 Claude 同时跑，可能几分钟就完成了原本需要一整天的工作。

### /batch 命令

```text
1. 交互式规划
   告诉 Claude 你想做什么（比如"把所有 React 类组件迁移到函数组件"）
   Claude 会分析项目，列出所有需要处理的文件

2. 确认执行
   你 review 计划，确认后 Claude 启动数十个 agent 并行执行

3. 汇总结果
   所有 agent 完成后，Claude 汇总成功/失败情况
   你只需要处理少数失败的 case
```

这种模式特别适合：
- 大规模重构
- 代码迁移
- 批量修复

---

## 7.7 协作模式的选择指南

```text
┌─────────────────────────────────────────────────────────────┐
│                  协作模式选择决策树                           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Q1: 子任务之间有依赖关系吗？                                │
│      是 → Sequential（顺序执行）                             │
│      否 → 继续                                              │
│                                                             │
│  Q2: 需要从多个角度分析吗？                                  │
│      是 → Parallel（并行执行）                               │
│      否 → 继续                                              │
│                                                             │
│  Q3: 需要反复优化质量吗？                                    │
│      是 → Writer-Reviewer（写作-审查）                       │
│      否 → 继续                                              │
│                                                             │
│  Q4: 任务需要分解协调吗？                                    │
│      是 → Hierarchical（层级协调）                           │
│      否 → 单 Agent 足够                                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 7.8 协作陷阱与最佳实践

### 陷阱 1：过度协作

不是所有任务都需要多 Agent。

❌ **过度协作：**
```text
Task: 写一个 Hello World 函数
Agent 1: 写代码
Agent 2: 审查
Agent 3: 测试
Agent 4: 文档
```

成本 4 倍，收益接近 0。

✅ **合理协作：**
```text
Task: 实现一个完整的认证系统
Agent 1: 实现核心逻辑
Agent 2: 安全审查
Agent 3: 性能测试
```

### 陷阱 2：上下文丢失

Agent 之间传递信息时，上下文容易丢失。

解决方案：
- 明确定义 Agent 间的接口
- 使用结构化的数据格式
- 记录完整的决策过程

### 陷阱 3：无限循环

Writer-Reviewer 模式可能陷入无限循环：
```text
Writer: 生成代码 v1
Reviewer: 不满意，要求修改
Writer: 生成代码 v2
Reviewer: 还是不满意
...
```

解决方案：
- 设置最大迭代次数
- 定义明确的验收标准
- 记录每次迭代的改进点

### 最佳实践

**1. 明确职责边界**

每个 Agent 应该有清晰的职责定义。

**2. 结构化通信**

Agent 间通信使用结构化格式，避免自然语言歧义。

**3. 增量验证**

每完成一个子任务就验证，而不是等全部完成。

**4. 成本控制**

多 Agent 协作成本高，需要权衡收益。

---

## 本章小结

### 核心要点回顾

**1. 为什么需要多 Agent**
- 职责分离、并行执行
- 互相审查、上下文隔离

**2. 四种核心模式**
- **Sequential**：顺序执行，适合有依赖的任务
- **Parallel**：并行执行，适合独立的任务
- **Writer-Reviewer**：迭代优化，适合需要高质量输出
- **Hierarchical**：层级协调，适合复杂任务分解

**3. 实现机制**
- **Git Worktrees**：并行运行多个独立任务
- **Subagents**：当前任务中调用专家
- **Agent Teams**：自动协调（Writer/Reviewer、Coordinator）
- **Fan-out**：批量处理（人海战术）

**4. 协作陷阱**
- 过度协作（成本不合理）
- 上下文丢失（接口不清晰）
- 无限循环（缺少退出条件）

### 关键洞察

> **不是你来协调多个 Agent，而是让 Agent 自己协调。好的协作架构应该让协调成本趋近于零。**

### 下一章预告

第8章我们将进入 RAG 与检索系统工程，讨论 Agent 如何从外部知识中构建可靠上下文和证据链。

---

## 参考资料

1. **CrewAI: Multi-Agent Framework** - https://github.com/joaomdmoura/crewAI
2. **LangGraph Multi-Agent Architectures** - https://langchain-ai.github.io/langgraph/
3. **Anthropic Claude Code Documentation** - Agent Teams
4. **Multi-Agent Systems: Survey** - Wooldridge & Jennings, 2023
5. **Distributed AI: From Agents to Teams** - AAAI Tutorial, 2025
