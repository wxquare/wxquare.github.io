# AI Book Methodology Spine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first implementation slice of the redesigned AI book structure by making Prompt Engineering, Context Engineering, and Harness Engineering deep standalone chapters.

**Architecture:** This slice updates the mdBook navigation and README to establish the new methodology spine, then creates or rewrites three focused first-part chapters. Existing source drafts are preserved unless explicitly rewritten; old theory-topic files can remain unlinked until a later cleanup phase.

**Tech Stack:** mdBook Markdown, Mermaid diagrams, existing `ai-book/src` content, local `mdbook build` verification.

---

## File Structure

- Modify: `ai-book/src/SUMMARY.md`
  - Responsibility: mdBook navigation for the new six-part architecture.
- Modify: `ai-book/src/README.md`
  - Responsibility: public-facing book introduction, content structure, and reading paths.
- Create: `ai-book/src/part1/prompt-engineering.md`
  - Responsibility: deep standalone chapter on Prompt Engineering as task protocol design.
- Create: `ai-book/src/part1/context-engineering.md`
  - Responsibility: deep standalone chapter on Context Engineering as information architecture.
- Modify: `ai-book/src/part1/chapter3.md`
  - Responsibility: rewrite existing Harness Engineering chapter so it focuses on Harness as Agent runtime environment, not a compressed three-paradigm overview.
- Modify: `ai-book/src/part1/chapter2.md`
  - Responsibility: renumber Claude Code chapter to chapter 5 and adjust its opening framing.

---

### Task 1: Update mdBook Navigation

**Files:**
- Modify: `ai-book/src/SUMMARY.md`

- [ ] **Step 1: Replace the table of contents with the methodology-spine structure**

Write the new navigation with these exact top-level parts:

```markdown
# 第一部分：AI 工程方法论基础
# 第二部分：Agent 架构与运行时设计
# 第三部分：知识、上下文与记忆系统
# 第四部分：生产级 Agent 治理
# 第五部分：完整架构案例
# 第六部分：Agent 应用工程师面试与作品集
# 附录
```

Use these first-part links:

```markdown
- [第1章 从 Vibe Coding 到 Spec Coding：AI 编程范式演进](part1/chapter1.md)
- [第2章 Prompt Engineering：从提示词到任务协议](part1/prompt-engineering.md)
- [第3章 Context Engineering：从上下文注入到信息架构](part1/context-engineering.md)
- [第4章 Harness Engineering：从模型调用到 Agent 运行环境](part1/chapter3.md)
- [第5章 Claude Code：终端原生的 AI Agent 实践](part1/chapter2.md)
```

- [ ] **Step 2: Keep later links conservative**

For this slice, update later part titles and link labels to the target architecture, but keep existing file paths so mdBook can build. Use existing paths for chapters that have not yet been moved.

- [ ] **Step 3: Run a quick link sanity check**

Run: `rg -n "part1/prompt-engineering|part1/context-engineering|专题：|理论专题：" ai-book/src/SUMMARY.md`

Expected:
- The two new part1 links appear.
- No `专题：` or `理论专题：` appears in `SUMMARY.md`.

---

### Task 2: Update Book Introduction and Reading Paths

**Files:**
- Modify: `ai-book/src/README.md`

- [ ] **Step 1: Rewrite the content structure section**

The new introduction must frame the book as:

```text
Prompt Engineering → Context Engineering → Harness Engineering → Agent Runtime → Knowledge / Memory / Retrieval → Production Governance → Case Studies → Interview / Portfolio
```

The first-part description must explicitly say that Prompt, Context, and Harness form the book's methodology spine.

- [ ] **Step 2: Update reading paths**

Adjust reading paths so:

- Quick-start readers read chapters 1, 2, 3, 4, then one case.
- Systematic readers read parts 1 through 5 before using interview material.
- Project-driven readers start from Agent architecture, then jump to Tool Calling, RAG, Evals, Guardrails, or Observability depending on the problem.
- Interview readers use the final part as a translation layer over the engineering chapters, not as the main body of knowledge.

- [ ] **Step 3: Search for stale positioning**

Run: `rg -n "基础理论补充|专题|第 12 章|第 18 章|第12章|第18章" ai-book/src/README.md`

Expected:
- No stale “基础理论补充” language remains.
- Chapter references match the new direction or are phrased generically.

---

### Task 3: Create Deep Prompt Engineering Chapter

**Files:**
- Create: `ai-book/src/part1/prompt-engineering.md`

- [ ] **Step 1: Write chapter skeleton**

Required sections:

```markdown
# 第2章 Prompt Engineering：从提示词到任务协议

## 引言
## 2.1 Prompt Engineering 的真正边界
## 2.2 Prompt 的四层结构
## 2.3 任务协议设计
## 2.4 结构化输出与后端契约
## 2.5 Few-shot、反例与决策边界
## 2.6 Prompt 与工具 Schema 的协同设计
## 2.7 Prompt 版本管理、评估与回滚
## 2.8 常见失败模式与修复方式
## 本章小结
```

- [ ] **Step 2: Add depth requirements**

The chapter must include:

- A clear claim that Prompt is a runtime protocol, not a magic phrase.
- A Mermaid or text diagram showing `Intent → Task Protocol → Model Output → Validator / Workflow`.
- A concrete production Agent example using alert diagnosis.
- A JSON output contract and schema-validation discussion.
- Prompt version metadata example.
- A failure-mode table covering instruction conflict, output drift, over-refusal, fake citation, tool misuse, and prompt bloat.

- [ ] **Step 3: Cross-link to next chapter**

End with a transition explaining why good Prompt Engineering still fails without Context Engineering: the model can follow instructions only when it has the right information.

---

### Task 4: Create Deep Context Engineering Chapter

**Files:**
- Create: `ai-book/src/part1/context-engineering.md`

- [ ] **Step 1: Write chapter skeleton**

Required sections:

```markdown
# 第3章 Context Engineering：从上下文注入到信息架构

## 引言
## 3.1 Context Engineering 的本质
## 3.2 上下文的类型系统
## 3.3 上下文优先级与可信度
## 3.4 上下文预算：token、成本、延迟与信息密度
## 3.5 上下文压缩：摘要、滑动窗口与事件化
## 3.6 上下文检索：RAG、metadata、query rewrite 与 rerank
## 3.7 上下文污染与 Context Rot
## 3.8 上下文防火墙：子代理、新会话与任务隔离
## 3.9 项目级上下文：CLAUDE.md、AGENTS.md、docs/ 与 specs/
## 本章小结
```

- [ ] **Step 2: Add depth requirements**

The chapter must include:

- A context assembly pipeline diagram.
- A table for context types, sources, lifecycle, and trust level.
- A priority rule that distinguishes current user input, tool results, authoritative docs, memory, and historical cases.
- A section explaining Context Rot and why long sessions degrade.
- A project-level context layout example.
- A checklist for deciding what enters context.

- [ ] **Step 3: Cross-link to next chapter**

End with a transition explaining that Prompt and Context still need Harness to provide tools, workflow, guardrails, evals, and observability.

---

### Task 5: Rewrite Harness Engineering Chapter

**Files:**
- Modify: `ai-book/src/part1/chapter3.md`

- [ ] **Step 1: Change title and framing**

Change title to:

```markdown
# 第4章 Harness Engineering：从模型调用到 Agent 运行环境
```

Opening must frame Harness as the runtime environment around the model, not just an evolution from Prompt.

- [ ] **Step 2: Restructure sections**

Required section structure:

```markdown
## 引言
## 4.1 Harness 的定义：Agent = Model + Harness
## 4.2 Harness 的六层架构
## 4.3 工具系统：Tool Registry、MCP、权限、超时、重试、审计
## 4.4 工作流控制：状态机、DAG、Router 与 Plan-and-Execute
## 4.5 验证回路：测试、lint、eval、review 与 human-in-the-loop
## 4.6 安全护栏：输入、上下文、工具、输出四层 Guardrails
## 4.7 可观测性：trace、step、tool call、cost 与 failure category
## 4.8 Harness 迭代：把失败沉淀成系统改造
## 本章小结
```

- [ ] **Step 3: Preserve useful existing material**

Keep and adapt:

- `Agent = Model + Harness`
- Harness six-component diagram, updated to six layers from the spec.
- Verification loop discussion.
- Context firewall idea, but make it a cross-reference to chapter 3 rather than the main focus.
- Industry proof points if they fit the new flow.

- [ ] **Step 4: Avoid over-claiming**

Remove exact benchmark percentages unless the chapter gives a source and context. Prefer qualitative claims unless sourced later in references.

---

### Task 6: Renumber Claude Code Chapter

**Files:**
- Modify: `ai-book/src/part1/chapter2.md`

- [ ] **Step 1: Change chapter title**

Change title to:

```markdown
# 第5章 Claude Code：终端原生的 AI Agent 实践
```

- [ ] **Step 2: Update section numbering**

Replace headings `2.1` through `2.7` with `5.1` through `5.7`.

- [ ] **Step 3: Adjust introduction**

Add one short paragraph after the opening quote:

```markdown
前面三章已经分别讨论了 Prompt、Context 和 Harness。本章把这些方法放进 Claude Code 这个具体工具里，观察它们如何变成日常开发中的 Plan 模式、CLAUDE.md、Skills、Hooks、MCP 和多会话协作。
```

---

### Task 7: Verify Build and Structural Integrity

**Files:**
- Verify: `ai-book/src/SUMMARY.md`
- Verify: `ai-book/src/part1/prompt-engineering.md`
- Verify: `ai-book/src/part1/context-engineering.md`
- Verify: `ai-book/src/part1/chapter3.md`
- Verify: `ai-book/src/part1/chapter2.md`

- [ ] **Step 1: Search for stale labels**

Run:

```bash
rg -n "专题：|理论专题：|基础理论补充|第2章 Claude|第3章 Harness Engineering：驾驭AI" ai-book/src/SUMMARY.md ai-book/src/README.md ai-book/src/part1
```

Expected:
- No stale labels in `SUMMARY.md`, `README.md`, or part 1.

- [ ] **Step 2: Build mdBook**

Run:

```bash
cd ai-book && mdbook build
```

Expected:
- Exit code 0.
- No missing-file errors for `part1/prompt-engineering.md` or `part1/context-engineering.md`.

- [ ] **Step 3: Run Hexo build if mdBook build passes**

Run:

```bash
npm run clean && npm run build
```

Expected:
- Exit code 0.
- This satisfies the repository instruction that content changes must be build-verified.

