# Agent 博客与书稿内容整合实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将两篇 Agent 博客的独有内容整合进现有 AI 书稿，并把博客移动到 `source/archive/AI/` 后停止发布。

**Architecture:** 采用按章节职责吸收的方式，不新增重复独立章节。第 13 章承接 Agent 选型与后端思维迁移，第 16 章承接 Harness 实证与检查表，第 28 章承接 DoD Agent 的 V1→V3 演进；博客原文作为 `published: false` 的归档副本保留。

**Tech Stack:** Markdown、mdBook 源稿、Hexo front matter、Ruby editorial-integrity 检查脚本、Git。

## Global Constraints

- 不重写两篇博客的全部正文。
- 不重排书稿的章节编号或 SUMMARY 导航。
- 不改动与本次整合无关的博客迁移、分类和归档文件。
- 只补博客独有内容；书稿已覆盖的安全、成本、性能、工具、RAG 和评估内容不整段复制。
- 归档文件保留原有正文和元数据，并设置 `published: false`。
- 保留用户工作区中与本任务无关的文章迁移和归档改动。

---

### Task 1: 补充第 13 章的 Agent 选型与后端迁移视角

**Files:**
- Modify: `books/ai-book/src/part2/01-agent-architecture.md`，在现有 `13.6 场景映射与落地校验` 和 `本章小结` 之间新增内容。

**Interfaces:**
- Consumes: 现有第 13 章的 Agent 选型、Runtime、架构模式和检查清单。
- Produces: 一段与第 14–22 章衔接的 Agent 设计决策框架，不新增 SUMMARY 条目。

- [ ] **Step 1: 读取插入点上下文**

  确认新增段落位于第 13 章场景映射之后、章节小结之前，并复用本章已有术语：Agent Runtime、State Machine、ReAct、Plan-and-Execute、Policy 和 Human Control Plane。

- [ ] **Step 2: 写入“从后端系统设计迁移到 Agent 系统设计”**

  新增一个二级小节，包含以下四个部分：

  1. 用“输入稳定性、信息分散度、路径开放性、动作风险、结果可验证性”判断传统后端、规则引擎、普通 LLM 应用或 Agent；
  2. 明确传统后端负责状态、权限、事务和确定性流程，Agent 负责意图理解、假设生成、证据综合和动态路径；
  3. 用表格比较固定流程、State Machine、Plan-and-Execute、Multi-Agent 和自定义 Runtime 的适用条件，不重复本章已有架构模式定义；
  4. 给出需求、风险、权限、成本、延迟、验证和人工接管的设计检查清单，并指向后续章节。

- [ ] **Step 3: 检查章节衔接**

  确认新内容没有重新解释已有 Runtime 组件，没有引入博客中的“第 N 章”旧编号，也没有新增外部链接或 SUMMARY 条目。

- [ ] **Step 4: 运行 Markdown 结构检查**

  运行：

  ```bash
  ruby books/ai-book/scripts/test_verify_editorial_integrity.rb
  ```

  预期：脚本通过，或只报告实施前已存在的问题；若失败，先检查新增标题、代码围栏和引用格式。

### Task 2: 补充第 16 章的 Harness 实证与工程师角色

**Files:**
- Modify: `books/ai-book/src/part2/04-harness-engineering.md`，在现有 Harness 迭代和最小可行检查表附近补充内容。

**Interfaces:**
- Consumes: 现有第 16 章的六层架构、失败分类、Eval、Observability 和迭代闭环。
- Produces: 面向书稿读者的行业案例对照、最小 Harness 组成和角色转变说明。

- [ ] **Step 1: 对照书稿与博客的重复部分**

  保留书稿已有的 Context Layer、Tool Layer、Workflow Layer、Guardrail Layer、Eval Layer 和 Observability Layer 定义；只补博客中尚未充分表达的证据和总结。

- [ ] **Step 2: 写入 Harness 行业实证**

  增加一个小节，使用表格表达三个案例：OpenAI 的架构约束与可观测性、LangChain 的自验证/循环检测/主动上下文、Anthropic 的规划/生成/评估协作。数字只作为特定案例数据，并保留博客中的来源链接或明确标注为案例，不写成普适基准。

- [ ] **Step 3: 写入最小 Harness 与角色转变**

  增加一个小节，整理最小可行 Harness 的七项能力：精简入口文档、可复现环境、机械化约束、自验证回路、上下文防火墙、最小权限与回滚、定期熵治理；随后用“传统工程师能力 → Harness 时代能力”表格总结职责变化。

- [ ] **Step 4: 检查与现有迭代闭环的一致性**

  确认新检查表中的每项都能落到现有 Context、Tool、Workflow、Guardrail、Eval 或 Observability 层，避免引入一个脱离正文的新框架。

- [ ] **Step 5: 运行 editorial integrity 检查**

  运行：

  ```bash
  ruby books/ai-book/scripts/test_verify_editorial_integrity.rb
  ```

  预期：通过；若失败，修复新增 Markdown 或引用问题后再进入下一任务。

### Task 3: 补充第 28 章的 DoD Agent 架构演进复盘

**Files:**
- Modify: `books/ai-book/src/part3/06-dod-agent-case-study.md`，在 `28.3 总体架构` 的架构分层说明附近补充演进复盘。

**Interfaces:**
- Consumes: 现有 DoD Agent 的双模式定位、风险分层、Harness 总体架构、Policy、Workflow、Tool Runtime 和上线阶段。
- Produces: 一个从 V1 到 V3 的可解释架构决策链，供本章后续数据模型、工作流和端到端案例引用。

- [ ] **Step 1: 写入 V1→V3 演进表**

  增加“传统后端 / 被动工具 → 规则引擎 / 固定流程 → 受控 Agent Runtime”的对照表，逐阶段写明解决的问题、剩余缺陷、适用边界和迁移触发条件。

- [ ] **Step 2: 写入关键决策复盘**

  用连续段落说明：为什么不能把所有逻辑交给 LLM，为什么不能一开始就自动执行，为什么最终采用“状态机 + 受控 ReAct + Policy + 人工接管”，并与本章现有 `28.7`、`28.8`、`28.13` 和 `28.14` 交叉引用。

- [ ] **Step 3: 去重并保持案例一致**

  不重复博客中已有的成本公式、Prompt 示例、工具实现和安全代码；统一 DoD Agent 的术语、业务域和风险分级，确保新段落与本章现有端到端案例一致。

- [ ] **Step 4: 运行 editorial integrity 检查**

  运行：

  ```bash
  ruby books/ai-book/scripts/test_verify_editorial_integrity.rb
  ```

  预期：通过，且第 28 章仍能被 SUMMARY 正常引用。

### Task 4: 移动并标记两篇博客为归档

**Files:**
- Move: `source/_posts/AI/02-agent-system-design-guid.md` → `source/archive/AI/02-agent-system-design-guid.md`
- Move: `source/_posts/AI/06-harness-engineering.md` → `source/archive/AI/06-harness-engineering.md`
- Modify: 两个归档文件的 front matter，增加 `published: false`。

**Interfaces:**
- Consumes: 两篇博客原文和仓库现有 `source/archive/AI/` 归档约定。
- Produces: 保留原文内容但不再生成博客页面的两个归档文件。

- [ ] **Step 1: 确认目标目录和源文件状态**

  确认 `source/archive/AI/` 已存在，两个源文件仍在 `source/_posts/AI/`，且目标文件名不存在；不覆盖该目录中的现有用户文件。

- [ ] **Step 2: 移动文件并补充归档元数据**

  使用版本可追踪的移动操作，将两个文件移到目标目录；在 YAML front matter 中加入：

  ```yaml
  published: false
  ```

  保留原有标题、日期、分类、标签、正文、代码和参考资料。

- [ ] **Step 3: 验证归档结果**

  运行：

  ```bash
  test ! -e source/_posts/AI/02-agent-system-design-guid.md
  test ! -e source/_posts/AI/06-harness-engineering.md
  test -f source/archive/AI/02-agent-system-design-guid.md
  test -f source/archive/AI/06-harness-engineering.md
  rg -n '^published: false$' source/archive/AI/02-agent-system-design-guid.md source/archive/AI/06-harness-engineering.md
  ```

  预期：源路径不存在、目标路径存在，两个归档文件均包含 `published: false`。

### Task 5: 全量验证与交付检查

**Files:**
- Test: `books/ai-book/scripts/test_verify_editorial_integrity.rb`
- Inspect: `books/ai-book/src/SUMMARY.md`
- Inspect: `git diff` and `git status`

**Interfaces:**
- Consumes: Tasks 1–4 的书稿修改和博客归档。
- Produces: 可审查的最终 diff 和验证结果，不修改 SUMMARY。

- [ ] **Step 1: 运行完整 editorial integrity 测试**

  运行：

  ```bash
  ruby books/ai-book/scripts/test_verify_editorial_integrity.rb
  ```

  预期：测试全部通过；若环境或仓库已有问题导致失败，记录失败命令、错误位置和是否与本次修改相关。

- [ ] **Step 2: 检查三章结构和归档元数据**

  运行：

  ```bash
  rg -n '^## |^### ' books/ai-book/src/part2/01-agent-architecture.md books/ai-book/src/part2/04-harness-engineering.md books/ai-book/src/part3/06-dod-agent-case-study.md
  rg -n '^published: false$' source/archive/AI/02-agent-system-design-guid.md source/archive/AI/06-harness-engineering.md
  ```

  预期：新增内容保持正确标题层级，SUMMARY 中现有三章链接无需变化，归档文件均为不可发布状态。

- [ ] **Step 3: 检查 diff 范围**

  运行：

  ```bash
  git status --short
  git diff --stat
  git diff -- books/ai-book/src/part2/01-agent-architecture.md books/ai-book/src/part2/04-harness-engineering.md books/ai-book/src/part3/06-dod-agent-case-study.md
  ```

  预期：除已存在的用户改动外，本任务只新增三处书稿内容、两处博客归档和必要的计划/设计文档，不出现其他文件的内容变化。

- [ ] **Step 4: 提交本任务文件**

  只暂存本任务的三个书稿文件、两个归档文件和计划文档；不暂存工作区原有的文章迁移改动。提交信息：

  ```bash
  git add books/ai-book/src/part2/01-agent-architecture.md \
    books/ai-book/src/part2/04-harness-engineering.md \
    books/ai-book/src/part3/06-dod-agent-case-study.md \
    source/archive/AI/02-agent-system-design-guid.md \
    source/archive/AI/06-harness-engineering.md \
    docs/superpowers/plans/2026-09-23-agent-blog-book-merge.md
  git commit -m "docs: merge agent blogs into ai book"
  ```

