# Agent Interview Portfolio Appendix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand Appendix D into a practical LLM/Agent system design interview and portfolio guide.

**Architecture:** Keep the implementation to one Markdown chapter so the mdBook navigation remains unchanged. Replace the current outline with a longer, structured appendix that combines job-market signals, system design questions, answer templates, portfolio templates, and verification checklists.

**Tech Stack:** Markdown, mdBook, existing Hexo/npm build pipeline.

---

## Files

- Modify: `books/ai-book/src/appendix/system-design-interview-portfolio.md`
- Reference: `docs/superpowers/specs/2026-05-16-agent-interview-portfolio-design.md`
- Verify with: `npm run clean && npm run build`

## Task 1: Replace Appendix D With Expanded Structure

**Files:**
- Modify: `books/ai-book/src/appendix/system-design-interview-portfolio.md`

- [x] **Step 1: Replace the chapter body**

Use `apply_patch` to replace the current appendix with these sections:

```text
# 附录D 系统设计面试题与作品集模板
D.1 LLM / Agent 岗位真实考察点
D.2 Agent 系统设计通用回答框架
D.3 企业知识库问答 Agent
D.4 客服工单处理 Agent
D.5 代码审查 Agent
D.6 生产告警诊断 Agent
D.7 Coding Agent / Agent Harness
D.8 Agent Evals 平台
D.9 企业 Tool Registry / MCP Gateway
D.10 Multi-agent Research Agent
D.11 Prompt Injection 与权限防护系统
D.12 LLM Observability 与 Trace Debugging 平台
D.13 个人知识管理 Agent
D.14 作品集项目材料包
D.15 GitHub README 模板
D.16 面试表达脚本
D.17 失败复盘模板
D.18 面试前检查清单
D.19 小结
```

Expected result: the file keeps the original appendix title but upgrades the content into a complete practical guide.

- [x] **Step 2: Preserve writing conventions**

Check the edited Markdown for:

```text
中文标点
中英文之间有空格
代码块带语言标识，例如 text、markdown、json
标题编号连续
无绝对图片路径
```

Expected result: the chapter follows the repository writing rules.

## Task 2: Add Interview Question Content

**Files:**
- Modify: `books/ai-book/src/appendix/system-design-interview-portfolio.md`

- [x] **Step 1: Add common per-question subsections**

Each major question should include the following where useful:

```text
需求
关键澄清
核心架构
设计重点
Evals
Observability
常见追问
优秀回答信号
常见扣分点
```

Expected result: readers can learn a reusable answer rhythm instead of memorizing isolated diagrams.

- [x] **Step 2: Add production-level signals**

Include these concrete design concerns across the questions:

```text
permission filter before generation
evidence package
tool risk level
sandbox and approval
trace and span
capability eval and regression eval
cost, latency, token usage
idempotency and retries
prompt injection isolation
human-in-the-loop checkpoint
```

Expected result: the appendix reflects real LLM/Agent engineering work rather than chatbot-only design.

## Task 3: Add Portfolio Templates

**Files:**
- Modify: `books/ai-book/src/appendix/system-design-interview-portfolio.md`

- [x] **Step 1: Add project material templates**

Add templates for:

```text
one-page project profile
architecture diagram
agent trace sample
eval dataset sample
eval report sample
failure postmortem
GitHub README
2-minute, 5-minute, 15-minute interview scripts
```

Expected result: a reader can package a project from README to interview narrative.

- [x] **Step 2: Add concrete examples**

Use examples based on Agent projects already covered in the book:

```text
Production Alert Diagnosis Agent
Enterprise Knowledge Assistant
Coding Agent
Personal Knowledge Management Agent
```

Expected result: the templates feel connected to earlier chapters.

## Task 4: Verify

**Files:**
- Verify: `books/ai-book/src/appendix/system-design-interview-portfolio.md`

- [x] **Step 1: Search for markdown mistakes**

Run:

```bash
rg -n "TODO|TBD|FIXME|```$" books/ai-book/src/appendix/system-design-interview-portfolio.md
```

Expected: no TODO/TBD/FIXME. Triple-backtick matches are acceptable and should be manually checked for balanced fences.

- [x] **Step 2: Build the site**

Run:

```bash
npm run clean && npm run build
```

Expected: command exits with status 0.

- [x] **Step 3: Review git diff**

Run:

```bash
git diff -- books/ai-book/src/appendix/system-design-interview-portfolio.md docs/superpowers/plans/2026-05-16-agent-interview-portfolio.md
```

Expected: diff only contains the appendix enhancement and this implementation plan.
