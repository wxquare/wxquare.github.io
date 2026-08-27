# Agent 演化融入第 8 章 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不改变目录路径和后续章节编号的前提下，将 Agent 演化的双主线融入第 8 章，并保持书籍元数据与构建输出一致。

**Architecture:** 第 8 章以新增 8.1 建立“技术形态演化 + 工程能力演化”框架；原有 8.1 至 8.5 的所有层级标题顺延为 8.2 至 8.6。`SUMMARY.md` 与 `README.md` 只更新对第二部分和第 8 章的叙述，不改变链接路径和第 9 至第 26 章编号。

**Tech Stack:** Markdown、mdBook、Hexo/npm build scripts、ripgrep。

## Global Constraints

- 只编辑 `books/ai-book/src/` 下的书稿源文件；不编辑 `books/ai-book/book/` 生成产物。
- 第 8 章文件路径保持 `books/ai-book/src/part2/01-agent-architecture.md`。
- 第 9 至第 26 章的编号、路径、内容均不因本次改动而调整。
- “自我演化”只能描述为受 Evals、审核、灰度发布与回滚约束的改进闭环。
- 以“企业告警与知识答疑”作为 8.1 的升级案例，并明确详细案例仍属于第 22 章。

---

### Task 1: 将 Agent 演化双主线写入第 8 章

**Files:**
- Modify: `books/ai-book/src/part2/01-agent-architecture.md`
- Test: `rg -n '^## 8\\.[1-6]|^### 8\\.[1-6]\\.' books/ai-book/src/part2/01-agent-architecture.md`

**Interfaces:**
- Consumes: 现有第 8 章的 Runtime 架构总纲与第 9 至第 17 章的章节职责。
- Produces: `8.1 Agent 的演化：从回答到受控完成工作`，作为后续 8.2 至 8.6 的导航和边界定义。

- [ ] **Step 1: 记录现有标题层级，确认顺延范围**

Run:

```bash
rg -n '^## 8\\.|^### 8\\.' books/ai-book/src/part2/01-agent-architecture.md
```

Expected: 现有二级标题从 `8.1` 到 `8.5`，以及对应三级标题；它们是唯一需要顺延的标题。

- [ ] **Step 2: 先写编号结构检查，确认新增后没有跳号**

Run:

```bash
rg -n '^## 8\\.[1-6] ' books/ai-book/src/part2/01-agent-architecture.md
```

Expected before edit: FAIL，因为尚不存在 `8.6` 且 `8.1` 仍不是演化章节。

- [ ] **Step 3: 写入新增 8.1 并顺延原章节标题**

在引言后插入 `## 8.1 Agent 的演化：从回答到受控完成工作`，包含：

```text
Chatbot → Prompt Application → Tool-using Agent → Workflow Agent
→ Runtime Agent → Multi-Agent → 受控持续改进系统

Prompt → Context → Harness → API / Tools → Knowledge / Memory
→ Orchestration → Evals / Guardrails / Observability
```

新增映射表的列固定为“阶段、主要能力、解决的问题、新增风险、需要补齐的工程能力”；新增企业告警与知识答疑的七阶段升级案例；增加“何时不应升级为 Agent”检查；把“自我演化”限定为评测、复盘、审核、灰度和回滚的闭环。

将原 `## 8.1` 至 `## 8.5` 及其所有 `### 8.x.y` 标题分别顺延为 `8.2` 至 `8.6`。同步更新章节引言对章节分段的说明，确保不再声称本章只从 8.1 至 8.5 展开。

- [ ] **Step 4: 运行标题和内容断言**

Run:

```bash
rg -n '^## 8\\.[1-6] ' books/ai-book/src/part2/01-agent-architecture.md
rg -n 'Chatbot|Workflow Agent|企业告警与知识答疑|灰度发布|何时不应升级为 Agent' books/ai-book/src/part2/01-agent-architecture.md
```

Expected: 恰好出现连续的 8.1 至 8.6；新增演化阶段、案例和受控改进边界均可检索。

- [ ] **Step 5: 提交书稿主体变更**

```bash
git add books/ai-book/src/part2/01-agent-architecture.md
git commit -m "docs: add agent evolution to architecture chapter"
```

### Task 2: 同步目录和全书阅读叙述

**Files:**
- Modify: `books/ai-book/src/SUMMARY.md`
- Modify: `books/ai-book/src/README.md`
- Test: `rg -n '第8章 Agent 的演化与架构总纲|先理解 Agent 的演化' books/ai-book/src/SUMMARY.md books/ai-book/src/README.md`

**Interfaces:**
- Consumes: Task 1 确定的第 8 章标题和第二部分的阅读顺序。
- Produces: 目录与书籍介绍均将“演化”作为 Agent 工程入口，但不更改既有 Markdown 链接目标。

- [ ] **Step 1: 写入目录标题检查，确认当前描述待更新**

Run:

```bash
rg -n '第8章' books/ai-book/src/SUMMARY.md
```

Expected before edit: 显示旧标题 `Agent 架构总纲：边界、Runtime 与模式选择`。

- [ ] **Step 2: 更新显示标题与第二部分说明**

把 `SUMMARY.md` 中第 8 章显示标题更新为 `第8章 Agent 的演化与架构总纲：从对话应用到可治理 Runtime`，链接继续指向 `part2/01-agent-architecture.md`。

把 `README.md` 中第二部分的说明更新为：先理解 Agent 如何从对话和 Prompt 应用演化为受控 Runtime，再依次展开任务协议、上下文、Harness、模型协议、工具、知识、记忆、编排和生产治理。

- [ ] **Step 3: 验证标题、叙述和链接保持一致**

Run:

```bash
rg -n '第8章 Agent 的演化与架构总纲|先理解 Agent 如何从对话' books/ai-book/src/SUMMARY.md books/ai-book/src/README.md
rg -n 'part2/01-agent-architecture.md' books/ai-book/src/SUMMARY.md
```

Expected: 两处叙述均包含演化定位；目录链接仍唯一指向原有第 8 章源文件。

- [ ] **Step 4: 提交目录与介绍变更**

```bash
git add books/ai-book/src/SUMMARY.md books/ai-book/src/README.md
git commit -m "docs: surface agent evolution in book navigation"
```

### Task 3: 书稿链接与构建验证

**Files:**
- Verify: `books/ai-book/src/`
- Verify: repository build output only; do not stage or edit `books/ai-book/book/`

**Interfaces:**
- Consumes: Tasks 1 and 2 的书稿源文件。
- Produces: 可构建书稿，以及对标题编号和目录链接有效性的验证记录。

- [ ] **Step 1: 检查目录链接目标存在**

Run:

```bash
awk -F'[()]' '/^- \\[/ {print $2}' books/ai-book/src/SUMMARY.md | while read -r path; do test -f "books/ai-book/src/$path" || echo "missing: $path"; done
```

Expected: 无输出。

- [ ] **Step 2: 扫描旧标题和异常章节编号**

Run:

```bash
rg -n '第8章 Agent 架构总纲：边界、Runtime 与模式选择|本章按五层展开：5\\.1|^## 8\\.[7-9]' books/ai-book/src
```

Expected: 无输出。

- [ ] **Step 3: 执行构建验证**

Run:

```bash
npm run clean
npm run build
git diff --check
```

Expected: 两个 npm 命令退出码为 0，`git diff --check` 无输出。构建可能刷新 `books/ai-book/book/`，但不得将该目录加入暂存区或提交。

- [ ] **Step 4: 提交验证之外的源文件变更（若 Task 1、2 已分别提交则无需额外提交）**

Run:

```bash
git status --short
```

Expected: 书稿源文件均已提交；如仅剩 `books/ai-book/book/` 变动，保留它们供主工作区所有者处理。
