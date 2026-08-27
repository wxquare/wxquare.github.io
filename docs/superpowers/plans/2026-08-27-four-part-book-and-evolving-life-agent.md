# 四部分书稿与持续进化生活 Agent Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 AI Agent 书稿重组为四部分，新增第 26 章“持续进化的生活 Agent”，并将研究综述迁移为第 27 章。

**Architecture:** 保持 Markdown 源码为唯一事实源。目录与阅读路径由 `SUMMARY.md`、`README.md` 描述；第 26 章负责把既有工程能力组合为受控的生活 Agent 演进闭环；第 27 章保留既有研究综述内容，仅调整章节归属、路径和编号。

**Tech Stack:** GitBook 风格 Markdown、Git、npm/Hexo 构建。

## Global Constraints

- 只编辑 `books/ai-book/src/` 的书稿源码；不得直接编辑 `books/ai-book/book/`、`public/` 或 `.deploy_git/`。
- 第 26 章只使用模拟数据、伪接口和审批边界；不得加入真实凭据或可执行的支付、购买、邮件发送、医疗或金融操作。
- 将原研究章节移至 `part4/` 并编号为第 27 章；所有受影响的内部引用必须同步。
- 仅替换第 25 章 Markdown 标题行里的 `24.x` 为 `25.x`，不得改变正文数值、代码或数据样例。
- 新章必须使用原创中文表述，并以参考书第 9 章的方法为概念输入：证据聚合、独立回归、人工审阅、渐进发布与回滚。
- 每项书稿变更后执行 Markdown 结构检查；完成后执行 `npm run clean && npm run build` 与 `git diff --check`。

---

### Task 1: 建立结构变更的失败检查

**Files:**
- Modify: `books/ai-book/src/SUMMARY.md`
- Modify: `books/ai-book/src/README.md`
- Move: `books/ai-book/src/part3/09-research-agent-overview.md` to `books/ai-book/src/part4/01-agent-frontier-research.md`

**Consumes:** 设计说明中定义的四部分目录与第 27 章路径。

**Produces:** 可重复执行的目录与引用验收命令；后续任务依赖其检查模式。

- [ ] **Step 1: 运行结构红灯检查**

```bash
test -f books/ai-book/src/part3/10-daily-life-evolving-agent.md
rg -n '^# 第四部分：前沿与研究$' books/ai-book/src/SUMMARY.md
rg -n '^# 第27章 ' books/ai-book/src/part4/01-agent-frontier-research.md
```

Expected: 三项均失败，因为新章节、第四部分和新研究路径尚不存在。

- [ ] **Step 2: 创建第四部分目录并迁移研究综述**

```bash
mkdir -p books/ai-book/src/part4
git mv books/ai-book/src/part3/09-research-agent-overview.md \
  books/ai-book/src/part4/01-agent-frontier-research.md
```

将迁移文件的一级标题由 `第26章` 改为 `第27章`；保留研究综述正文与参考资料。

- [ ] **Step 3: 更新书籍目录和阅读路径**

在 `SUMMARY.md` 中：

```markdown
# 第三部分：Agent 应用与实战
...
- [第26章 持续进化的生活 Agent：从日常反馈到可信能力闭环](part3/10-daily-life-evolving-agent.md)

# 第四部分：前沿与研究

- [第27章 AI 智能体研究现状、工程瓶颈与未来理想能力架构报告](part4/01-agent-frontier-research.md)
```

在 `README.md` 中增加“第四部分：前沿与研究”，并将两处研究型阅读建议从第 26 章更新为第 27 章；补充生活 Agent 的阅读入口。

- [ ] **Step 4: 运行结构绿灯检查**

```bash
test -f books/ai-book/src/part4/01-agent-frontier-research.md
test ! -e books/ai-book/src/part3/09-research-agent-overview.md
rg -n '^# 第四部分：前沿与研究$|^# 第27章 ' \
  books/ai-book/src/SUMMARY.md books/ai-book/src/part4/01-agent-frontier-research.md
```

Expected: 所有检查成功，且目录链接路径指向存在的文件。

- [ ] **Step 5: 提交结构迁移**

```bash
git add books/ai-book/src/SUMMARY.md books/ai-book/src/README.md \
  books/ai-book/src/part3/09-research-agent-overview.md \
  books/ai-book/src/part4/01-agent-frontier-research.md
git commit -m "docs: split book into four parts"
```

### Task 2: 编写第 26 章持续进化生活 Agent

**Files:**
- Create: `books/ai-book/src/part3/10-daily-life-evolving-agent.md`
- Modify: `books/ai-book/src/SUMMARY.md`
- Modify: `books/ai-book/src/README.md`

**Consumes:** 第 10–17 章的上下文、Harness、工具、知识、记忆、编排和治理内容；第 25 章的个人知识管理案例。

**Produces:** 第 26 章完整 Markdown，并从目录和阅读路径可访问。

- [ ] **Step 1: 运行章节红灯检查**

```bash
test -f books/ai-book/src/part3/10-daily-life-evolving-agent.md
rg -n '^## 26\.[1-7] ' books/ai-book/src/part3/10-daily-life-evolving-agent.md
```

Expected: 命令失败，因为第 26 章尚未创建。

- [ ] **Step 2: 创建章节骨架**

创建文件并写入以下标题：

```markdown
# 第26章 持续进化的生活 Agent：从日常反馈到可信能力闭环

## 26.1 长期运行的生活 Agent：范围、授权与非目标
## 26.2 以证据为先的运行数据模型
## 26.3 三层评估：结果、过程与质量验证
## 26.4 更新路由：知识、Prompt/Skill、工作流与 Harness
## 26.5 受控发布闭环
## 26.6 案例：每周生活回顾 Agent
## 26.7 上线检查清单与度量
```

- [ ] **Step 3: 填充受控进化闭环与案例**

围绕“日历、待办、邮件、购物清单”的模拟每周回顾案例，写清：

1. 建议、草稿、外部动作三层权限；
2. 不可变运行轨迹和结构化经验卡；
3. 结果、过程、质量三层评估；
4. 知识、Prompt/Skill、工作流、Harness 四路更新；
5. 多轨迹证据聚合、留出集、人工审阅、小流量发布、版本记录和回滚；
6. 个人隐私、数据最小化、保留期和删除机制；
7. 成功率、人工接管率、越权拦截率、用户修订率、回归通过率、回滚恢复时间等度量。

使用相对 Markdown 链接连接第 10、11、13–17 和第 25 章；不得复制参考书原文，不得加入真实服务凭据或可执行外部动作。

- [ ] **Step 4: 运行章节绿灯检查**

```bash
rg -n '^## 26\.[1-7] ' books/ai-book/src/part3/10-daily-life-evolving-agent.md
rg -n '人工审阅|留出集|回滚|最小权限|经验卡' \
  books/ai-book/src/part3/10-daily-life-evolving-agent.md
```

Expected: 七个编号小节和五类受控演进要素都存在。

- [ ] **Step 5: 提交新章节**

```bash
git add books/ai-book/src/part3/10-daily-life-evolving-agent.md \
  books/ai-book/src/SUMMARY.md books/ai-book/src/README.md
git commit -m "docs: add evolving daily life agent chapter"
```

### Task 3: 修复第 25 章遗留编号与全书引用

**Files:**
- Modify: `books/ai-book/src/part3/07-pkm-agent-case-study.md`
- Modify: 所有 `books/ai-book/src/**/*.md` 中仍指向旧研究第 26 章或旧路径的引用（由扫描结果决定）

**Consumes:** Task 1 产生的第 27 章路径，Task 2 产生的第 26 章。

**Produces:** 无重排遗留的标题编号和无旧研究路径的书稿源文件。

- [ ] **Step 1: 运行编号与引用红灯检查**

```bash
rg -n '^#{2,6} 24\.' books/ai-book/src/part3/07-pkm-agent-case-study.md
rg -n 'part3/09-research-agent-overview|第 ?26 章' books/ai-book/src --glob '*.md'
```

Expected: 第 25 章出现 `24.x` 标题；扫描结果中包含迁移前的研究型引用或目录文本。

- [ ] **Step 2: 仅修复 Markdown 标题编号**

使用只匹配行首 Markdown 标题的替换，将 `part3/07-pkm-agent-case-study.md` 中 `24.` 改为 `25.`。完成后人工检查 diff，确认正文、代码块和数据数值没有变化。

- [ ] **Step 3: 更新语义明确的旧引用**

将指向研究综述的 `第 26 章` 改为 `第 27 章`，并把 Markdown 路径换为 `part4/01-agent-frontier-research.md`。第 26 章生活 Agent 的新引用不得被误改。

- [ ] **Step 4: 运行编号与引用绿灯检查**

```bash
! rg -n '^#{2,6} 24\.' books/ai-book/src/part3/07-pkm-agent-case-study.md
! rg -n 'part3/09-research-agent-overview' books/ai-book/src --glob '*.md'
rg -n '^#{2,6} 25\.' books/ai-book/src/part3/07-pkm-agent-case-study.md
```

Expected: 没有 `24.x` 标题和旧路径，且第 25 章标题均以 `25.` 开头。

- [ ] **Step 5: 提交编号与引用修复**

```bash
git add books/ai-book/src/part3/07-pkm-agent-case-study.md
git commit -m "docs: fix application chapter references"
```

### Task 4: 全书结构验证与构建

**Files:**
- Verify: `books/ai-book/src/SUMMARY.md`
- Verify: `books/ai-book/src/**/*.md`

**Consumes:** Tasks 1–3 的全部书稿变动。

**Produces:** 已验证的四部分书稿提交。

- [ ] **Step 1: 验证目录链接和章节编号**

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
rg -n '^# 第2[5-7]章 ' books/ai-book/src/part3 books/ai-book/src/part4
```

Expected: 所有目录 Markdown 链接存在；第 25、26、27 章各有正确一级标题。

- [ ] **Step 2: 构建书稿站点**

```bash
npm run clean
npm run build
```

Expected: 成功完成 Hexo 构建。构建产生的 `books/ai-book/book/` 文件为生成物，不纳入提交。

- [ ] **Step 3: 检查变更边界和空白错误**

```bash
git diff --check
git log --name-only --format='' -3
```

Expected: 无空白错误；最近书稿提交只涉及 `books/ai-book/src/`，且不包含 `books/ai-book/book/`。

- [ ] **Step 4: 提交最终验证记录（仅在产生源码修复时）**

```bash
git status --short
```

Expected: 书稿源码工作区干净；现有或构建产生的 `books/ai-book/book/` 变更保持未暂存且不修改。
