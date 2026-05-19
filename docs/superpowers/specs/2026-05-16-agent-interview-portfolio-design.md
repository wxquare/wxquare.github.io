# Agent 系统设计面试与作品集附录增强设计

## 背景

`books/ai-book/src/appendix/system-design-interview-portfolio.md` 目前提供了 Agent 系统设计回答框架、4 道典型题和若干作品集模板。章节方向正确，但内容更像提纲，缺少真实 LLM/Agent 岗位常考信号、追问方式、候选人回答层次和作品集包装细节。

## 目标

- 将附录 D 增强为“面试实战 + 作品集表达”的综合章节。
- 覆盖 LLM/Agent 岗位高频能力：RAG、tool use、agent harness、multi-agent、evals、observability、guardrails、sandbox、cost/latency 和 failure debugging。
- 每道题提供可直接复用的回答结构、架构图、追问点、优秀回答信号和常见扣分点。
- 给读者一套把 demo 讲成生产级 Agent 项目的作品集模板。

## 信息来源策略

不写通用爬虫。公开网页质量差异很大，直接爬取容易引入 SEO 面试题和低质量二手资料。本次增强采用“深度检索 + 官方/一手资料优先 + 原创改写”的方式：

- OpenAI Codex Agents 岗位描述：抽取 agent harness、工具执行、长任务、安全执行、model evals 等岗位信号。
- OpenAI Agents SDK 文档：参考 tracing、guardrails、handoffs、tools 等生产组件。
- Anthropic Building Effective Agents：参考 workflow 与 agent 的边界、routing、parallelization、evaluator-optimizer、orchestrator-workers 等模式。
- Anthropic Multi-agent Research System：参考 orchestrator-worker、多子 Agent 并行、citation agent、确定性 checkpoint 和 retry。
- Anthropic agent evals 文章与 LangSmith/Braintrust/Arize 等岗位或文档：参考 agent eval、trace grading、observability、生产反馈闭环。

正文不搬运原题，而是沉淀成原创中文面试题和回答模板。

## 章节结构

增强后的附录 D 保留现有主线，并扩展为以下结构：

```text
D.1 岗位真实考察点
D.2 通用回答框架
D.3-D.12 系统设计题库
D.13 作品集项目模板
D.14 GitHub README 模板
D.15 失败复盘与面试表达模板
D.16 面试前检查清单
```

系统设计题库计划覆盖：

- 企业知识库问答 Agent；
- 客服工单处理 Agent；
- 代码审查 Agent；
- 生产告警诊断 Agent；
- Coding Agent / Agent Harness；
- Agent Evals 平台；
- 企业 Tool Registry / MCP Gateway；
- Multi-agent Research Agent；
- Prompt Injection 与权限防护系统；
- LLM Observability 与 Trace Debugging 平台；
- 个人知识管理 Agent 或工作流自动化 Agent。

## 单题模板

每道题尽量使用同一结构，便于读者形成肌肉记忆：

```text
需求
关键澄清
核心架构
关键组件
数据流 / 状态流
Guardrails
Evals
Observability
常见追问
优秀回答信号
常见扣分点
```

架构图优先使用文本图或 Mermaid。由于当前章节已使用文本架构图，为保持 mdBook 渲染稳定，新增图以文本图为主，必要时使用 Mermaid。

## 作品集模板

作品集部分补齐以下材料：

- 一页项目介绍模板；
- 项目架构图模板；
- Agent trace 示例模板；
- eval dataset 示例；
- eval report 示例；
- failure postmortem 模板；
- GitHub README 结构；
- 面试 2 分钟、5 分钟、15 分钟讲法。

## 验证

- 修改 `books/ai-book/src/appendix/system-design-interview-portfolio.md`。
- 保持中文标点、中英文之间空格、代码块语言标注。
- 不修改 `themes/`、`db.json`、`.deploy_git`、`node_modules`。
- 完成后运行 `npm run clean && npm run build`。
- 如构建失败，修复后重新验证。

## 风险与处理

- 篇幅过长：控制为附录实战手册，不展开框架源码级实现。
- 内容像清单而非教学：每道题加入回答层次、追问和扣分点。
- 过度依赖互联网材料：只抽象岗位信号，正文保持原创表达。
- 链接过多影响阅读：正文少放外链，必要来源可集中放在“参考信号”小节。
