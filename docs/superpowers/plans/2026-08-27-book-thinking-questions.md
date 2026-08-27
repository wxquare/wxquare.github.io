# 全书思考题与设计推演导向 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 移除书稿的阅读路径、面试和作品集导向，将问题、答案和项目材料统一改为思考题、设计推演和项目实践。

**Architecture:** `README.md` 与 `SUMMARY.md` 是导航入口；正文保留技术结论，逐处改写引导语；附录 D/E 迁移到新的思考题路径，并重构其求职语境。目录链接扫描和关键词扫描是本次文档改造的回归测试。

**Tech Stack:** GitBook 风格 Markdown、Git、Node.js、npm/Hexo。

## Global Constraints

- 只修改 `books/ai-book/src/`；不得直接编辑 `books/ai-book/book/`、`public/` 或 `.deploy_git/`。
- 删除 README 中“阅读路径”“快速上手”“系统学习”“项目驱动”和天/周学习时间承诺。
- 除外部来源名称或 URL 的引用保真外，书稿中文源内容不得再出现“面试”或“作品集”。
- 保留技术结论、架构取舍、题目、答案、代码、项目 README、架构图、指标和失败复盘。
- 附录 D 路径改为 `appendix/system-design-thinking-and-practice.md`；附录 E 路径改为 `appendix/llm-agent-thinking-questions.md`。
- 每项任务先运行 `rg` 红灯检查；最终执行目录链接检查、`git diff --check` 和 `npm run clean && npm run build`。

---

### Task 1: 更新导航并迁移附录文件

**Files:**

- Modify: `books/ai-book/src/README.md`
- Modify: `books/ai-book/src/SUMMARY.md`
- Move: `books/ai-book/src/appendix/system-design-interview-portfolio.md` → `books/ai-book/src/appendix/system-design-thinking-and-practice.md`
- Move: `books/ai-book/src/appendix/llm-agent-interview-question-bank.md` → `books/ai-book/src/appendix/llm-agent-thinking-questions.md`

**Produces:** 无阅读路径的 README；附录 D/E 的新路径和目录链接。

- [ ] **Step 1: 写出失败检查并确认旧导航存在**

```bash
rg -n '阅读路径|快速上手|系统学习|项目驱动|面试|作品集' \
  books/ai-book/src/README.md books/ai-book/src/SUMMARY.md
test -f books/ai-book/src/appendix/system-design-thinking-and-practice.md
test -f books/ai-book/src/appendix/llm-agent-thinking-questions.md
```

Expected: 第一项命中旧语境；两个新路径不存在。

- [ ] **Step 2: 移动附录、删除阅读路径并更新目录**

```bash
git mv books/ai-book/src/appendix/system-design-interview-portfolio.md \
  books/ai-book/src/appendix/system-design-thinking-and-practice.md
git mv books/ai-book/src/appendix/llm-agent-interview-question-bank.md \
  books/ai-book/src/appendix/llm-agent-thinking-questions.md
```

删除 README 中 `## 阅读路径` 到下一处 `##` 标题之间的全部内容。将附录说明和 SUMMARY 标题分别改为“系统设计思考题与项目实践模板”和“LLM / Agent 思考题与参考来源”，并使用两个新路径。

- [ ] **Step 3: 运行导航绿灯检查并提交**

```bash
! rg -n '阅读路径|快速上手|系统学习|项目驱动' books/ai-book/src/README.md
test -f books/ai-book/src/appendix/system-design-thinking-and-practice.md
test -f books/ai-book/src/appendix/llm-agent-thinking-questions.md
! rg -n 'system-design-interview-portfolio|llm-agent-interview-question-bank' books/ai-book/src --glob '*.md'
git add books/ai-book/src/README.md books/ai-book/src/SUMMARY.md books/ai-book/src/appendix
git commit -m "docs: reframe book navigation around thinking questions"
```

### Task 2: 改写第一至第三部分的求职语境

**Files:**

- Modify: 命中 `面试` 或 `作品集` 的 `books/ai-book/src/part1/**/*.md`
- Modify: 命中 `面试` 或 `作品集` 的 `books/ai-book/src/part2/**/*.md`
- Modify: 命中 `面试` 或 `作品集` 的 `books/ai-book/src/part3/**/*.md`

**Produces:** 仍有原题目、技术推导与代码，但导向变为思考题、设计推演、设计评审或项目实践。

- [ ] **Step 1: 写出失败检查并枚举受影响正文**

```bash
rg -n '面试|作品集' books/ai-book/src/part1 books/ai-book/src/part2 books/ai-book/src/part3 --glob '*.md'
```

Expected: 输出所有需要逐段处理的正文位置。

- [ ] **Step 2: 对每处命中进行语义转换**

按以下映射逐段改写，不能用全局替换破坏句意：

```text
面试表达 → 思考题 或 设计推演
面试加分表达 → 进阶设计推演
面试官问 → 设计评审或复盘时可追问
回答 → 说明、推导或给出设计依据
作品集 → 项目实践 或 项目材料
```

使用“设计清单”收束能力模型；使用“思考题”引导独立推导；使用“设计推演”保留架构权衡和失败模式。不得删除原有技术结论或代码块。

- [ ] **Step 3: 运行正文绿灯检查并提交**

```bash
! rg -n '面试|作品集' books/ai-book/src/part1 books/ai-book/src/part2 books/ai-book/src/part3 --glob '*.md'
rg -n '思考题|设计推演|设计评审|项目实践' books/ai-book/src/part1 books/ai-book/src/part2 books/ai-book/src/part3 --glob '*.md'
git diff --check
git add books/ai-book/src/part1 books/ai-book/src/part2 books/ai-book/src/part3
git commit -m "docs: replace interview framing in book chapters"
```

### Task 3: 重构附录 D、E 的叙述语境

**Files:**

- Modify: `books/ai-book/src/appendix/system-design-thinking-and-practice.md`
- Modify: `books/ai-book/src/appendix/llm-agent-thinking-questions.md`

**Produces:** 面向设计评审、复盘和项目实践的附录题目、参考答案和材料。

- [ ] **Step 1: 写出失败检查**

```bash
rg -n '面试|作品集|候选人|面试官|面试信号' \
  books/ai-book/src/appendix/system-design-thinking-and-practice.md \
  books/ai-book/src/appendix/llm-agent-thinking-questions.md
```

Expected: 两个附录仍出现旧的求职导向。

- [ ] **Step 2: 重写附录 D 的标题和引导语**

将标题改为 `# 附录D 系统设计思考题与项目实践模板`。统一使用“思考材料”“设计题”“评审关注点”“设计阐述框架”“设计评审前检查清单”“项目实践材料包”“项目实践提示”。保留 README、架构图、指标和失败复盘，但不再把它们表述为求职展示。

- [ ] **Step 3: 重写附录 E 的标题和引导语**

将标题改为 `# 附录E LLM / Agent 思考题与参考来源`。统一使用“问题来源”“学习材料”“社区实践反馈”“主题思考题”“反思与追问”“快速推演参考”。保留外部 URL 和外部原始英文标题；书稿自身中文叙述不能保留旧术语。

- [ ] **Step 4: 运行附录绿灯检查并提交**

```bash
! rg -n '面试|作品集|候选人|面试官|面试信号' \
  books/ai-book/src/appendix/system-design-thinking-and-practice.md \
  books/ai-book/src/appendix/llm-agent-thinking-questions.md
rg -n '^# 附录D 系统设计思考题与项目实践模板$|^# 附录E LLM / Agent 思考题与参考来源$' \
  books/ai-book/src/appendix/system-design-thinking-and-practice.md \
  books/ai-book/src/appendix/llm-agent-thinking-questions.md
git add books/ai-book/src/appendix/system-design-thinking-and-practice.md books/ai-book/src/appendix/llm-agent-thinking-questions.md
git commit -m "docs: reframe appendices as engineering exercises"
```

### Task 4: 全书验收与构建

**Files:**

- Verify: `books/ai-book/src/SUMMARY.md`
- Verify: `books/ai-book/src/**/*.md`

**Produces:** 可构建、无旧导向术语的书稿源码。

- [ ] **Step 1: 验证全书术语和目录链接**

```bash
node - <<'NODE'
const fs = require('fs');
const path = require('path');
const summary = fs.readFileSync('books/ai-book/src/SUMMARY.md', 'utf8');
const links = [...summary.matchAll(/\]\(([^)]+\.md)\)/g)].map(m => m[1]);
const missing = links.filter(link => !fs.existsSync(path.join('books/ai-book/src', link)));
if (missing.length) throw new Error(`Missing summary links: ${missing.join(', ')}`);
console.log(`Validated ${links.length} summary Markdown links.`);
NODE
! rg -n '面试|作品集' books/ai-book/src --glob '*.md'
! rg -n '阅读路径|快速上手|系统学习|项目驱动' books/ai-book/src/README.md
```

Expected: 目录链接均存在，书稿自身中文源内容没有旧导向术语。

- [ ] **Step 2: 构建并检查边界**

```bash
npm run clean
npm run build
git diff --check
git log --name-only --format='' -3
git status --short
```

Expected: 构建成功；无空白错误；书稿提交只修改 `books/ai-book/src/`；生成物不纳入提交。
