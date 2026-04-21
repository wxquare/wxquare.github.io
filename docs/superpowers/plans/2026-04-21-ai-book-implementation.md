# AI工程实践书籍 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将7篇AI工程博客（16,700行）系统化整理为完整的电子书，包含11章内容和完整目录结构。

**Architecture:** 基于mdBook构建静态电子书，通过GitHub Actions自动部署到GitHub Pages。内容分4部分：AI编程实战(3章)、Agent系统设计(4章)、实战案例(2章)、基础理论(2章)。

**Tech Stack:** mdBook, mdbook-mermaid, GitHub Actions, GitHub Pages

---

## 文件结构

```
ai-book/
├── book.toml                    # mdBook配置
├── src/
│   ├── SUMMARY.md               # 目录结构（关键文件）
│   ├── README.md                # 书籍介绍
│   ├── part1/                   # 第一部分：AI编程实战
│   │   ├── chapter1.md          # 从Vibe Coding到Spec Coding
│   │   ├── chapter2.md          # Claude Code实践指南
│   │   └── chapter3.md          # Harness Engineering方法论
│   ├── part2/                   # 第二部分：Agent系统设计
│   │   ├── chapter4.md          # Agent架构设计与决策框架
│   │   ├── chapter5.md          # 工具系统与MCP协议
│   │   ├── chapter6.md          # 多Agent协作与工作流编排
│   │   └── chapter7.md          # 可观测性与成本优化
│   ├── part3/                   # 第三部分：实战案例
│   │   ├── chapter8.md          # DoD Agent案例
│   │   └── chapter9.md          # 个人知识管理Agent
│   ├── part4/                   # 第四部分：基础理论
│   │   ├── chapter10.md         # LLM能力边界与工程化
│   │   └── chapter11.md         # RAG与上下文工程基础
│   ├── appendix/
│   │   ├── glossary.md          # 术语表
│   │   ├── references.md        # 参考资料
│   │   └── tools.md             # 常用工具与框架
│   └── images/                  # 图片资源
├── theme/
│   └── custom.css               # 自定义样式
├── mermaid.min.js               # Mermaid库
├── mermaid-init.js              # Mermaid初始化
└── .github/workflows/
    └── deploy-ai-book.yml       # 自动部署工作流
```

---

## Task 1: 项目结构初始化

**Files:**
- Modify: `ai-book/book.toml`
- Modify: `ai-book/src/SUMMARY.md`
- Modify: `ai-book/src/README.md`
- Create: `.github/workflows/deploy-ai-book.yml`

- [ ] **Step 1: 更新 book.toml 配置**

更新书籍元信息（书名、描述、作者）

- [ ] **Step 2: 创建完整的 SUMMARY.md**

创建包含11章完整结构的目录文件

- [ ] **Step 3: 更新 README.md 书籍介绍**

更新书籍介绍，说明目标读者和内容特色

- [ ] **Step 4: 创建 GitHub Actions 工作流**

创建自动部署配置文件

- [ ] **Step 5: 本地构建测试**

运行命令验证配置正确：
```bash
cd ai-book
mdbook build
```
预期：构建成功，无错误

- [ ] **Step 6: 提交结构初始化**

```bash
git add ai-book/book.toml ai-book/src/SUMMARY.md ai-book/src/README.md .github/workflows/deploy-ai-book.yml
git commit -m "feat(ai-book): initialize book structure with 11 chapters"
```

---

## Task 2: 创建章节文件骨架

**Files:**
- Create: `ai-book/src/part1/chapter1.md`
- Create: `ai-book/src/part1/chapter2.md`
- Create: `ai-book/src/part1/chapter3.md`
- Create: `ai-book/src/part2/chapter4.md`
- Create: `ai-book/src/part2/chapter5.md`
- Create: `ai-book/src/part2/chapter6.md`
- Create: `ai-book/src/part2/chapter7.md`
- Create: `ai-book/src/part3/chapter8.md`
- Create: `ai-book/src/part3/chapter9.md`
- Create: `ai-book/src/part4/chapter10.md`
- Create: `ai-book/src/part4/chapter11.md`

- [ ] **Step 1: 创建第一部分章节骨架**

创建 part1 目录和3个章节文件，每个文件包含章节标题和基本结构

- [ ] **Step 2: 创建第二部分章节骨架**

创建 part2 目录和4个章节文件

- [ ] **Step 3: 创建第三部分章节骨架**

创建 part3 目录和2个章节文件

- [ ] **Step 4: 创建第四部分章节骨架**

创建 part4 目录和2个章节文件

- [ ] **Step 5: 验证构建**

运行命令验证所有文件创建正确：
```bash
cd ai-book
mdbook build
mdbook serve --open
```
预期：所有11章在目录中显示

- [ ] **Step 6: 提交章节骨架**

```bash
git add ai-book/src/part1/ ai-book/src/part2/ ai-book/src/part3/ ai-book/src/part4/
git commit -m "feat(ai-book): create 11 chapter skeleton files"
```

---

## Task 3: 迁移第1章 - Vibe Coding到Spec Coding

**Files:**
- Source: `source/_posts/AI/00-vibe-coding-vs-spec-coding.md`
- Target: `ai-book/src/part1/chapter1.md`

- [ ] **Step 1: 读取源博客内容**

读取完整的博客文件，理解结构和内容

- [ ] **Step 2: 移除 Front Matter**

移除 Hexo 博客的 Front Matter（title, date, categories, tags等）

- [ ] **Step 3: 调整章节结构**

将博客内容重新组织为6个小节：
- 1.1 AI编程工具的三次演进
- 1.2 Vibe Coding的本质与陷阱
- 1.3 Spec Coding：规范驱动的工程化方法
- 1.4 编写高质量Spec的完整指南
- 1.5 Cursor IDE完整实践
- 1.6 Claude Code工作流与最佳实践

- [ ] **Step 4: 添加章节引言**

在章节开头添加引言，说明本章背景和重要性

- [ ] **Step 5: 处理代码块和图表**

检查所有代码块有语言标注，Mermaid图表正确渲染

- [ ] **Step 6: 添加本章小结**

在章节末尾添加核心要点回顾

- [ ] **Step 7: 本地验证**

```bash
cd ai-book
mdbook serve
```
打开浏览器检查第1章渲染正常

- [ ] **Step 8: 提交第1章**

```bash
git add ai-book/src/part1/chapter1.md
git commit -m "feat(ai-book): complete chapter 1 - Vibe Coding to Spec Coding"
```

---

## Task 4: 迁移第2章 - Claude Code实践指南

**Files:**
- Source: `source/_posts/AI/01-claude-code-practices.md`
- Target: `ai-book/src/part1/chapter2.md`

- [ ] **Step 1: 读取源博客内容**

读取完整的博客文件

- [ ] **Step 2: 移除 Front Matter 并调整结构**

移除元数据，重新组织为7个小节

- [ ] **Step 3: 扩充实战案例**

在原有内容基础上，扩充更多实际使用案例

- [ ] **Step 4: 补充最佳实践部分**

添加完整的最佳实践清单和常见陷阱

- [ ] **Step 5: 添加引言和小结**

补充章节引言和核心要点小结

- [ ] **Step 6: 本地验证**

```bash
cd ai-book
mdbook serve
```
检查第2章渲染正常

- [ ] **Step 7: 提交第2章**

```bash
git add ai-book/src/part1/chapter2.md
git commit -m "feat(ai-book): complete chapter 2 - Claude Code practices"
```

---

## Task 5: 迁移第3章 - Harness Engineering

**Files:**
- Source: `source/_posts/AI/06-harness-engineering.md`
- Target: `ai-book/src/part1/chapter3.md`

- [ ] **Step 1: 读取源博客内容**

读取博客内容，理解Harness Engineering核心概念

- [ ] **Step 2: 移除 Front Matter 并重组**

移除元数据，重新组织为7个小节

- [ ] **Step 3: 扩充核心组件详解**

详细展开Harness的七大核心组件（上下文、工具、验证、约束、可观测性等）

- [ ] **Step 4: 补充设计清单**

添加完整的Harness设计和实施清单

- [ ] **Step 5: 添加实战案例**

补充Harness在实际项目中的应用案例

- [ ] **Step 6: 添加引言和小结**

补充章节引言和核心要点小结

- [ ] **Step 7: 本地验证和提交**

```bash
cd ai-book
mdbook serve
# 验证通过后
git add ai-book/src/part1/chapter3.md
git commit -m "feat(ai-book): complete chapter 3 - Harness Engineering"
```

---

## Task 6: 迁移第4章 - Agent架构设计

**Files:**
- Source: `source/_posts/AI/02-agent-system-design-guid.md` (第1-8章部分)
- Target: `ai-book/src/part2/chapter4.md`

- [ ] **Step 1: 提取架构设计相关内容**

从8900+行的agent-guide中提取第1-8章：
- Agent vs 传统后端系统
- 主流框架对比
- 需求分析框架
- 架构设计方法论
- ReACT模式
- Plan-and-Execute模式
- 状态机与混合架构
- 数据流与状态管理

- [ ] **Step 2: 重新组织为8个小节**

按照新的章节结构重新编排内容

- [ ] **Step 3: 补充决策框架**

添加架构决策的完整思维框架

- [ ] **Step 4: 添加对比表格和图表**

使用表格和Mermaid图清晰展示架构对比

- [ ] **Step 5: 添加引言和小结**

补充章节引言和核心要点小结

- [ ] **Step 6: 本地验证和提交**

```bash
cd ai-book
mdbook serve
# 验证通过后
git add ai-book/src/part2/chapter4.md
git commit -m "feat(ai-book): complete chapter 4 - Agent architecture design"
```

---

## Task 7: 迁移第5章 - 工具系统与MCP

**Files:**
- Source: `source/_posts/AI/02-agent-system-design-guid.md` (第11章部分)
- Target: `ai-book/src/part2/chapter5.md`

- [ ] **Step 1: 提取工具系统内容**

从agent-guide提取第11章工具系统设计部分

- [ ] **Step 2: 重组为7个小节**

按照新结构编排：工具抽象、接口设计、发现注册、MCP协议、编排组合、错误处理、安全控制

- [ ] **Step 3: 补充MCP协议细节**

详细展开MCP（Model Context Protocol）的概念、架构和使用方法

- [ ] **Step 4: 添加代码示例**

补充工具接口定义、MCP使用的完整代码示例

- [ ] **Step 5: 添加最佳实践**

总结工具系统设计的最佳实践和常见陷阱

- [ ] **Step 6: 添加引言和小结**

补充章节引言和核心要点小结

- [ ] **Step 7: 本地验证和提交**

```bash
cd ai-book
mdbook serve
# 验证通过后
git add ai-book/src/part2/chapter5.md
git commit -m "feat(ai-book): complete chapter 5 - Tool system and MCP"
```

---

## Task 8: 迁移第6章 - 多Agent协作

**Files:**
- Source: `source/_posts/AI/02-agent-system-design-guid.md` (第6章部分)
- Target: `ai-book/src/part2/chapter6.md`

- [ ] **Step 1: 提取多Agent协作内容**

从agent-guide提取第6章多Agent协作模式部分

- [ ] **Step 2: 重组为6个小节**

按照新结构编排多Agent协作内容

- [ ] **Step 3: 补充工作流编排策略**

详细展开不同场景下的工作流编排方法

- [ ] **Step 4: 添加协作模式图**

使用Mermaid绘制不同协作模式的架构图

- [ ] **Step 5: 添加引言和小结**

补充章节引言和核心要点小结

- [ ] **Step 6: 本地验证和提交**

```bash
cd ai-book
mdbook serve
# 验证通过后
git add ai-book/src/part2/chapter6.md
git commit -m "feat(ai-book): complete chapter 6 - Multi-agent collaboration"
```

---

## Task 9: 迁移第7章 - 可观测性与成本优化

**Files:**
- Source: `source/_posts/AI/02-agent-system-design-guid.md` (第12章部分)
- Target: `ai-book/src/part2/chapter7.md`

- [ ] **Step 1: 提取可观测性内容**

从agent-guide提取第12章可观测性与成本优化部分

- [ ] **Step 2: 重组为6个小节**

按照新结构编排：挑战、日志指标追踪、LLM监控、成本分析、性能优化、调试排查

- [ ] **Step 3: 补充监控实践**

添加完整的Agent系统监控和告警实践

- [ ] **Step 4: 补充成本优化策略**

详细展开LLM成本控制的具体方法

- [ ] **Step 5: 添加引言和小结**

补充章节引言和核心要点小结

- [ ] **Step 6: 本地验证和提交**

```bash
cd ai-book
mdbook serve
# 验证通过后
git add ai-book/src/part2/chapter7.md
git commit -m "feat(ai-book): complete chapter 7 - Observability and cost optimization"
```

---

## Task 10: 迁移第8章 - DoD Agent实战案例

**Files:**
- Source: `source/_posts/AI/03-dod-agent-design.md`
- Target: `ai-book/src/part3/chapter8.md`

- [ ] **Step 1: 读取DoD Agent完整设计文档**

读取2370行的完整案例内容

- [ ] **Step 2: 移除 Front Matter 并重组**

移除元数据，重新组织为9个小节

- [ ] **Step 3: 优化案例呈现结构**

突出从需求分析→架构设计→实现→部署→评估的完整流程

- [ ] **Step 4: 补充代码示例**

添加关键组件的代码实现示例

- [ ] **Step 5: 添加架构演进说明**

说明从v1到v2的架构演进过程和决策

- [ ] **Step 6: 添加引言和小结**

补充章节引言和核心要点小结

- [ ] **Step 7: 本地验证和提交**

```bash
cd ai-book
mdbook serve
# 验证通过后
git add ai-book/src/part3/chapter8.md
git commit -m "feat(ai-book): complete chapter 8 - DoD Agent case study"
```

---

## Task 11: 迁移第9章 - 个人知识管理Agent

**Files:**
- Source: `source/_posts/AI/04-karpathy-evolving-knowledge-base.md`
- Source: `source/_posts/AI/05-openclaw-research-and-practice.md`
- Target: `ai-book/src/part3/chapter9.md`

- [ ] **Step 1: 读取两篇源博客**

读取Karpathy知识管理(948行)和OpenClaw实践(1006行)

- [ ] **Step 2: 融合知识管理概念**

将Karpathy的理论框架和OpenClaw的实践经验融合

- [ ] **Step 3: 重组为7个小节**

按照新结构编排：自进化知识库、系统架构、OpenClaw框架、技能系统、记忆管理、多平台集成、从零搭建

- [ ] **Step 4: 形成完整实践指南**

补充从理论到实践的完整路径

- [ ] **Step 5: 添加搭建步骤**

详细的个人助手搭建步骤和配置说明

- [ ] **Step 6: 添加引言和小结**

补充章节引言和核心要点小结

- [ ] **Step 7: 本地验证和提交**

```bash
cd ai-book
mdbook serve
# 验证通过后
git add ai-book/src/part3/chapter9.md
git commit -m "feat(ai-book): complete chapter 9 - Personal knowledge management agent"
```

---

## Task 12: 撰写第10章 - LLM基础理论

**Files:**
- Target: `ai-book/src/part4/chapter10.md`
- Reference: 从已有博客提取LLM相关片段

- [ ] **Step 1: 从现有博客提取LLM概念**

在vibe-coding、harness、agent-guide等文章中提取LLM相关内容

- [ ] **Step 2: 设计章节结构**

规划6个小节：概览、能力边界、推理生成、模型选择、Prompt基础、工程化实践

- [ ] **Step 3: 撰写LLM概览**

简明介绍大语言模型的发展和基本原理

- [ ] **Step 4: 撰写能力边界**

说明LLM擅长和不擅长的任务，避免误用

- [ ] **Step 5: 撰写工程化要点**

总结LLM在生产环境的关键工程化考虑

- [ ] **Step 6: 控制篇幅**

精简内容，控制在5,000-8,000字

- [ ] **Step 7: 添加引言和小结**

补充章节引言和核心要点小结

- [ ] **Step 8: 本地验证和提交**

```bash
cd ai-book
mdbook serve
# 验证通过后
git add ai-book/src/part4/chapter10.md
git commit -m "feat(ai-book): complete chapter 10 - LLM fundamentals"
```

---

## Task 13: 撰写第11章 - RAG基础理论

**Files:**
- Target: `ai-book/src/part4/chapter11.md`
- Reference: 从karpathy文章提取RAG部分

- [ ] **Step 1: 从karpathy文章提取RAG内容**

提取知识管理系统中的RAG相关部分

- [ ] **Step 2: 设计章节结构**

规划6个小节：系统架构、文档处理、向量检索、上下文管理、混合检索、评估优化

- [ ] **Step 3: 撰写RAG系统架构**

说明RAG的基本原理和架构组件

- [ ] **Step 4: 撰写检索优化**

详细说明向量检索和混合检索策略

- [ ] **Step 5: 撰写实践指南**

总结RAG系统的工程实践和优化方法

- [ ] **Step 6: 控制篇幅**

精简内容，控制在5,000-8,000字

- [ ] **Step 7: 添加引言和小结**

补充章节引言和核心要点小结

- [ ] **Step 8: 本地验证和提交**

```bash
cd ai-book
mdbook serve
# 验证通过后
git add ai-book/src/part4/chapter11.md
git commit -m "feat(ai-book): complete chapter 11 - RAG fundamentals"
```

---

## Task 14: 完善附录内容

**Files:**
- Modify: `ai-book/src/appendix/glossary.md`
- Modify: `ai-book/src/appendix/references.md`
- Create: `ai-book/src/appendix/tools.md`

- [ ] **Step 1: 编写附录A - 术语表**

整理书中所有重要术语（Agent、LLM、RAG、MCP、ReACT、Harness等）并添加简明定义

- [ ] **Step 2: 编写附录B - 参考资料**

整理所有引用的论文、文章、工具文档的链接

- [ ] **Step 3: 编写附录C - 工具与框架**

列出书中提到的所有工具和框架（Cursor、Claude Code、LangChain、OpenClaw等）及官方链接

- [ ] **Step 4: 本地验证和提交**

```bash
cd ai-book
mdbook serve
# 验证通过后
git add ai-book/src/appendix/
git commit -m "feat(ai-book): complete appendix - glossary, references, and tools"
```

---

## Task 15: 内容优化与统一

**Files:**
- Modify: All chapter files in `ai-book/src/`

- [ ] **Step 1: 统一术语**

全书搜索并统一关键术语：
- Agent（不用agent、AI Agent混用）
- LLM（不用大模型、大语言模型混用）
- RAG（统一用全称或缩写）
- 等等

- [ ] **Step 2: 检查代码块格式**

遍历所有章节，确保所有代码块有语言标注：
```bash
cd ai-book/src
# 检查没有语言标注的代码块
rg '```\s*$' -A 1
```

- [ ] **Step 3: 检查中英文空格**

确保中英文之间有空格，符合写作规范

- [ ] **Step 4: 处理图片路径**

检查所有图片引用，确保路径正确（相对路径）

- [ ] **Step 5: 添加章节间衔接**

在每章末尾添加下一章预告，增强连贯性

- [ ] **Step 6: 检查内部链接**

验证所有章节间的交叉引用链接有效

- [ ] **Step 7: 提交优化**

```bash
git add ai-book/src/
git commit -m "refactor(ai-book): unify terminology and improve formatting"
```

---

## Task 16: 最终验证与部署

**Files:**
- All files in `ai-book/`
- `.github/workflows/deploy-ai-book.yml`

- [ ] **Step 1: 本地完整构建**

```bash
cd ai-book
mdbook clean
mdbook build
```
预期：无错误，无警告

- [ ] **Step 2: 本地预览测试**

```bash
mdbook serve --open
```
在浏览器中验证：
- 左侧目录树完整显示
- 所有章节可正常跳转
- 右侧大纲自动生成
- 代码块高亮正常
- Mermaid图表渲染正常
- 搜索功能可用

- [ ] **Step 3: 检查构建产物**

检查 `ai-book/book/` 目录下的HTML文件是否完整

- [ ] **Step 4: 提交最终版本**

```bash
git add .
git commit -m "feat(ai-book): complete v1.0 - 11 chapters with full content"
```

- [ ] **Step 5: 推送触发部署**

```bash
git push origin hexo
```

- [ ] **Step 6: 验证GitHub Actions**

访问 https://github.com/wxquare/wxquare.github.io/actions
等待构建完成（约2-3分钟）

- [ ] **Step 7: 验证在线访问**

访问 https://wxquare.github.io/ai-book/
检查：
- 所有章节正常显示
- 搜索功能正常
- 图片和图表显示正常
- 移动端访问正常

- [ ] **Step 8: 更新项目文档**

更新 `BOOK_PROJECT.md` 或创建 `AI_BOOK_PROJECT.md` 说明文档，记录：
- 项目结构
- 日常更新流程
- 故障排查指南

- [ ] **Step 9: 最终提交**

```bash
git add BOOK_PROJECT.md
git commit -m "docs: add AI book project documentation"
git push origin hexo
```

---

## 验证清单

### 内容完整性
- [ ] 11章全部完成，无占位符
- [ ] 每章包含完整结构（引言、核心内容、最佳实践、小结）
- [ ] 附录A、B、C全部完成
- [ ] 总字数达到10-12万字

### 格式规范
- [ ] 所有代码块指定语言
- [ ] 中英文之间有空格
- [ ] 术语统一
- [ ] 章节标题层级正确
- [ ] 链接格式正确（相对路径）

### 构建与部署
- [ ] 本地构建无错误
- [ ] GitHub Actions自动构建成功
- [ ] 在线访问正常
- [ ] 搜索功能可用
- [ ] 移动端适配良好

### 阅读体验
- [ ] 左侧目录树正常
- [ ] 右侧大纲自动生成
- [ ] 代码块支持复制
- [ ] 图表显示清晰

---

## 时间估算

| 任务组 | 任务数 | 预计时间 |
|--------|--------|---------|
| Task 1-2: 结构搭建 | 2 | 1天 |
| Task 3-5: 第一部分（3章） | 3 | 2天 |
| Task 6-9: 第二部分（4章） | 4 | 2天 |
| Task 10-11: 第三部分（2章） | 2 | 1天 |
| Task 12-13: 第四部分（2章） | 2 | 2天 |
| Task 14-15: 附录与优化 | 2 | 2天 |
| Task 16: 最终验证部署 | 1 | 1天 |
| **总计** | **16** | **11天** |

---

## 执行建议

1. **每日目标明确**：按照任务组划分，每天完成固定的任务
2. **频繁提交**：每完成一章就提交，避免大量修改堆积
3. **持续验证**：每天结束前运行 `mdbook serve` 本地验证
4. **记录问题**：遇到的格式问题、术语不一致等记录下来，统一处理
5. **保持专注**：优先完成有素材的章节（Task 3-11），最后补充新写内容（Task 12-13）

---

**Plan Status:** Ready for execution
**Created:** 2026-04-21
**Estimated Completion:** 2026-05-02 (11 working days)
