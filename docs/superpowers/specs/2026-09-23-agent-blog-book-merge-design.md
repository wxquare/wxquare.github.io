# Agent 博客与书稿内容整合设计

## 背景

仓库中的两篇 AI 博客与 `books/ai-book` 的 Agent 工程章节存在较高重合：

- `source/_posts/AI/02-agent-system-design-guid.md` 覆盖 Agent 选型、架构模式、工具、知识、记忆、DoD Agent 实践、成本、安全和面试视角。
- `source/_posts/AI/06-harness-engineering.md` 覆盖 Prompt / Context / Harness 的范式演进、Harness 组件、行业案例、工程师角色变化和实践检查表。

书稿已经有相应的章节主线，因此目标不是追加两篇重复长文，而是提取博客中仍有价值的内容，按书稿章节职责重新组织，并保留博客作为不可发布的历史归档。

## 目标

1. 将两篇博客中书稿尚未充分覆盖的观点、案例和检查清单整合进现有章节。
2. 保持书稿现有五部分结构和章节导航，不新增重复的独立长章节。
3. 将两篇博客移动到 `source/archive/AI/`，并通过 `published: false` 防止其继续作为 Hexo 文章发布。
4. 保留用户工作区中与本任务无关的文章迁移和归档改动。

## 非目标

- 不重写两篇博客的全部正文。
- 不重排书稿的章节编号或 SUMMARY 导航。
- 不改动与本次整合无关的博客迁移、分类和归档文件。
- 不把未经核验的博客数据扩写成新的事实性结论；行业数据保留为案例说明，并在书稿中附上已有来源信息。

## 内容整合方案

### 第 13 章：Agent 的演化与架构总纲

在现有 Agent 选型和 Runtime 主线附近新增一段“从后端系统设计迁移到 Agent 系统设计”，提取博客中的：

- 何时应使用传统后端、规则引擎、普通 LLM 应用或 Agent；
- Agent 与传统后端的混合分工；
- 框架选型应服务于运行时边界，而不是反过来决定架构；
- 面向需求、风险、权限、成本和延迟的设计检查清单。

已存在的 Runtime、ReAct、Plan-and-Execute、State Machine 和 Multi-Agent 说明只做必要的交叉引用，不重复复制。

### 第 16 章：Harness Engineering

在现有六层 Harness 架构和迭代闭环中补入：

- OpenAI、LangChain、Anthropic 案例所体现的“同一模型、不同 Harness”实证；
- Harness 的最小工程组成：上下文入口、机械化约束、自验证回路、上下文防火墙、最小权限和熵治理；
- 工程师职责从编写业务代码扩展到设计约束、验证、观测和反馈环境的变化。

这些内容会被改写为书稿的工程论证，不重复博客的宣传式叙述；数值和外部案例保留来源链接或明确案例属性。

### 第 28 章：DoD Agent 案例

在总体架构说明附近加入 DoD Agent 的演进复盘：

```text
V1：传统后端 / 被动工具
  -> V2：规则引擎和固定流程
  -> V3：受控 Agent Runtime
```

重点说明每一阶段解决了什么、仍然缺什么，以及为什么最终采用“状态机 + 受控 ReAct + Policy + 人工接管”的混合方案。博客中已有、且书稿已经覆盖的安全、成本、性能、工具、RAG 和评估内容不再整段复制。

## 归档方案

执行以下文件移动：

- `source/_posts/AI/02-agent-system-design-guid.md` → `source/archive/AI/02-agent-system-design-guid.md`
- `source/_posts/AI/06-harness-engineering.md` → `source/archive/AI/06-harness-engineering.md`

归档文件保留原有正文和元数据，并在 front matter 中设置 `published: false`。不在 `SUMMARY.md` 中加入归档博客，不调整 `_config.yml` 的正式文章目录配置。

## 验证策略

1. 检查三个书稿章节的 Markdown 标题层级、代码围栏和 Mermaid 围栏闭合。
2. 检查归档文件存在、原博客路径不再存在，并确认 `published: false`。
3. 运行 `ruby books/ai-book/scripts/test_verify_editorial_integrity.rb`。
4. 运行仓库已有的书稿构建或 Markdown 检查命令；如果环境缺少构建依赖，记录具体阻塞原因。
5. 复查 `git diff`，确认只包含本任务的三处书稿内容修改、两处博客归档及必要的设计/计划文档。

## 风险与应对

- **内容重复**：以“只补独有内容”为原则，修改前后对照章节标题和相邻段落。
- **案例数据过时**：把行业数字表述为特定案例，保留来源，不将其写成普适性基准。
- **归档仍被发布**：使用 `published: false`，并验证 Hexo 输入目录中只剩归档副本。
- **用户未提交改动被覆盖**：只编辑列明的目标文件，提交时只暂存本设计文档；实施阶段逐个检查目标文件的当前内容。
