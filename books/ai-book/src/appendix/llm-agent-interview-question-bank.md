# LLM / Agent 面试题库与参考来源

> 采集日期：2026-05-16；补充日期：2026-05-21  
> 用途：为《附录D 系统设计面试题与作品集模板》提供外部面试信号、题型来源和后续扩展素材。  
> 原则：不复制大段原文，不直接搬运题库答案；只记录来源、摘要、主题标签和可原创改写的面试题。

## 采集方法

本次采集采用“搜索种子 + 页面打开 + 人工筛选”的轻量爬虫流程：

```text
1. 使用关键词搜索 LLM engineer interview、RAG system design、AI agent interview、agent evals、agent observability 等主题；
2. 优先打开官方岗位、官方工程文章、官方文档和公开题库页面；
3. 对每个来源抽取标题、主题、面试信号、可改写题目和质量等级；
4. 将低质量 SEO 内容降级为“题型参考”，不作为正文事实依据；
5. 将重复题型合并成原创中文题目，供附录后续扩展。
```

质量等级说明：

```text
A：官方岗位、官方工程文章、官方文档，可作为高置信岗位信号；
B：公开题库、系统设计教程、GitHub 题库，可作为题型来源；
C：社区帖子、二手经验、SEO 文章，只作为趋势或追问方向参考。
```

---

## 1. 官方岗位信号

### 1.1 OpenAI：AI Systems Engineer, Codex Agents

- URL: <https://openai.com/careers/ai-systems-engineer-codex-agents-san-francisco/>
- 类型：官方岗位描述
- 质量等级：A
- 标签：`agent-harness`、`coding-agent`、`sandbox`、`evals`、`observability`、`latency-cost`
- 摘要：岗位明确强调 Codex Core Agents 团队关注 agent harness、模型交互、推理运行时、沙箱执行、编排、evals、生产可靠性，以及 token、延迟、成本、容量和质量的综合优化。
- 可改写题目：
  - 设计一个 Coding Agent 的 agent harness，它如何解释模型输出、调用工具、执行代码并安全完成长任务？
  - 如果一个 Coding Agent 线上 solve rate 下降，你如何区分是模型、prompt、harness、工具、推理服务还是产品交互的问题？
  - 如何设计 sandbox，让 Agent 能运行测试但不能破坏用户仓库或读取敏感文件？
- 可映射附录章节：D.7 Coding Agent / Agent Harness、D.8 Agent Evals 平台、D.12 LLM Observability。

### 1.2 Braintrust：Eval Engineer

- URL: <https://jobs.ashbyhq.com/Braintrust/929447dd-14cc-4bf6-9f50-30a2fda4e0a0/>
- 类型：官方岗位描述
- 质量等级：A
- 标签：`evals`、`agent-systems`、`tool-workflows`、`benchmark`、`technical-storytelling`
- 摘要：岗位强调将模型、prompt、agent architecture 和 tool workflow 变成可测实验，构造 dataset、scoring logic、evaluation harness，并分析 trace、输出和失败模式。
- 可改写题目：
  - 设计一个 Agent Evals 平台，如何比较两个 agent architecture 的真实任务表现？
  - 如何为一个多工具 Agent 构造能暴露失败模式和边界条件的评估集？
  - 如何把一次线上失败转化成可复现的 eval case？
- 可映射附录章节：D.8 Agent Evals 平台、D.12 LLM Observability、D.14 作品集项目材料包。

### 1.3 LangChain：Deployed Engineer

- URL: <https://jobs.ashbyhq.com/langchain/7ed7ca6f-2b4a-4dcd-8c14-c0f85fcf9ae2/>
- 类型：官方岗位描述
- 质量等级：A
- 标签：`production-agents`、`architecture-review`、`langgraph`、`evals`、`observability`、`guardrails`
- 摘要：岗位侧重和客户共同设计生产级 Agent，覆盖 conversational agents、research agents、multi-step workflows，并强调生产部署、可靠运行、架构评审和技术权衡。
- 可改写题目：
  - 客户已经有一个能跑的 Agent demo，你如何帮它变成可上线系统？
  - 如何评审一个 LangGraph / workflow-based Agent 的架构风险？
  - 生产 Agent 需要哪些 eval、observability 和 guardrails 才能交付给企业客户？
- 可映射附录章节：D.2 通用回答框架、D.4 客服工单处理 Agent、D.10 Multi-agent Research Agent。

### 1.4 LangChain：Product Marketing - Observability

- URL: <https://jobs.ashbyhq.com/langchain/fc61254d-2dd2-4b92-b01a-162e805b921c>
- 类型：官方岗位描述
- 质量等级：A
- 标签：`observability`、`online-evals`、`human-feedback`、`prompt-optimization`
- 摘要：岗位材料把 Agent 生产化能力拆成 observability 和 evaluation 两个核心产品面：前者关注 tracing、production monitoring、automated insights，后者关注 offline / online evals、human feedback 和 prompt optimization。
- 可改写题目：
  - 设计一个平台，让工程团队能追踪 Agent 行为并发现线上退化。
  - 如何把 human feedback 接入 eval 和 prompt 优化闭环？
  - Agent observability 和传统日志监控有什么不同？
- 可映射附录章节：D.12 LLM Observability、D.8 Agent Evals 平台。

### 1.5 Fieldguide：AI Engineer, Quality

- URL: <https://jobs.ashbyhq.com/fieldguide/d86bef91-71ab-494f-a3c8-393ad8c55063/>
- 类型：官方岗位描述
- 质量等级：A
- 标签：`quality-platform`、`production-feedback`、`agent-trace`、`failure-modes`
- 摘要：岗位强调把 evaluation 做成一等工程能力，建设统一评估平台、自动化 pipeline、生产反馈闭环、agent trace 和 failure mode 分析。
- 可改写题目：
  - 如何设计一个统一 eval 平台，让团队能在几小时内评估新模型对关键 workflow 的影响？
  - 如何把生产失败样本转成一等 eval case？
  - 如何让 Agent 的推理和动作对审计场景透明可信？
- 可映射附录章节：D.8 Agent Evals 平台、D.12 LLM Observability。

### 1.6 Judgment Labs：Forward Deploy AI Engineer

- URL: <https://jobs.ashbyhq.com/judgmentlabs/d6613cd7-9b73-4ebb-85be-d6aa0584f1ee>
- 类型：官方岗位描述
- 质量等级：A
- 标签：`agent-behavior-monitoring`、`instruction-drift`、`context-retrieval-loss`、`production-monitoring`
- 摘要：岗位材料突出 Agent Behavior Monitoring，不只看异常和延迟，也关注 instruction drift、context retrieval loss、行为聚类和回归定位。
- 可改写题目：
  - 传统 observability 只能看到错误率和延迟，如何监控 Agent 的行为退化？
  - 如何发现某类用户请求导致 Agent 经常发生 context retrieval loss？
  - 如何把 agent behavior monitoring 和 regression eval 连接起来？
- 可映射附录章节：D.12 LLM Observability、D.8 Agent Evals 平台。

---

## 2. 官方工程文章和文档

### 2.1 Anthropic：Demystifying evals for AI agents

- URL: <https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents>
- 类型：官方工程文章
- 质量等级：A
- 标签：`agent-evals`、`grader`、`transcript`、`trajectory`、`evaluation-harness`、`regression-eval`
- 摘要：文章把 Agent eval 拆成 task、trial、grader、transcript、outcome、evaluation harness、agent harness 和 evaluation suite。它还强调 Agent 评估需要看完整轨迹和最终环境状态，而不是只看最终回答。
- 可改写题目：
  - 如何设计一个 Agent eval case？task、grader、trace 和 outcome 分别是什么？
  - 为什么 Agent 评估不能只看 final answer？
  - 如何区分 capability eval 和 regression eval？
  - Coding Agent 的 eval 为什么适合使用 deterministic grader？
- 可映射附录章节：D.8 Agent Evals 平台、D.7 Coding Agent、D.12 LLM Observability。

### 2.2 Anthropic：Building Effective Agents

- URL: <https://www.anthropic.com/engineering/building-effective-agents>
- 类型：官方工程文章
- 质量等级：A
- 标签：`workflow-vs-agent`、`routing`、`parallelization`、`evaluator-optimizer`、`tool-design`、`sandbox`
- 摘要：文章强调先从简单系统开始，只有当固定 workflow 不够时再引入 Agent。Agent 适合开放、多步、依赖环境反馈的问题，但需要 guardrails、沙箱测试、清晰工具设计和停止条件。
- 可改写题目：
  - 什么时候应该用 workflow，什么时候应该用 Agent？
  - 设计一个客服系统时，哪些路径应该是固定 workflow，哪些路径可以交给 Agent？
  - evaluator-optimizer 适合哪些任务？如何评估迭代是否真的带来改善？
  - Tool description 写不好会导致什么问题？
- 可映射附录章节：D.2 通用回答框架、D.4 客服工单处理 Agent、D.9 Tool Registry。

### 2.3 Anthropic：How we built our multi-agent research system

- URL: <https://www.anthropic.com/engineering/multi-agent-research-system>
- 类型：官方工程文章
- 质量等级：A
- 标签：`multi-agent`、`orchestrator-worker`、`research-agent`、`citation-agent`、`coordination-complexity`
- 摘要：文章介绍 LeadResearcher + Subagents + CitationAgent 的研究系统。关键面试信号包括：多 Agent 不是默认方案；需要明确 delegation；要按任务复杂度控制子 Agent 数量；最终报告需要 citation verification。
- 可改写题目：
  - 设计一个多 Agent 研究系统，如何让主 Agent 分配任务并避免重复搜索？
  - 如何控制子 Agent 数量和 token 预算？
  - 如何验证最终报告中的每个 claim 都有来源支持？
  - 多 Agent 系统相较单 Agent 主要增加了哪些失败模式？
- 可映射附录章节：D.10 Multi-agent Research Agent、D.12 Observability。

### 2.4 OpenAI Agents SDK：Tracing

- URL: <https://openai.github.io/openai-agents-python/tracing/>
- 类型：官方文档
- 质量等级：A
- 标签：`trace`、`span`、`tool-call`、`handoff`、`guardrail-span`、`debugging`
- 摘要：文档说明 Agent tracing 应覆盖 LLM generations、tool calls、handoffs、guardrails 和 custom events。面试中可抽象为“trace 是 Agent 的可解释执行轨迹，而不是普通日志”。
- 可改写题目：
  - 设计 Agent observability 时，trace 和 span 应该记录哪些信息？
  - 如何用 trace 定位一次 Agent 失败发生在 retrieval、generation、tool 还是 handoff 阶段？
  - trace 中含敏感数据时如何脱敏和采样？
- 可映射附录章节：D.12 LLM Observability、D.8 Agent Evals。

### 2.5 OpenAI Agents SDK：Guardrails

- URL: <https://openai.github.io/openai-agents-python/guardrails/>
- 类型：官方文档
- 质量等级：A
- 标签：`guardrails`、`input-guardrail`、`output-guardrail`、`tool-guardrail`、`tripwire`
- 摘要：文档把 guardrails 放在输入、输出和工具调用边界。对于有 managers、handoffs 或 delegated specialists 的 workflow，需要 tool-level guardrails，不能只依赖 Agent 入口和最终输出检查。
- 可改写题目：
  - input guardrail、output guardrail 和 tool guardrail 的边界是什么？
  - 为什么高风险工具要在调用前后都做校验？
  - blocking guardrail 和 parallel guardrail 在成本、延迟和副作用上如何权衡？
- 可映射附录章节：D.11 Prompt Injection 与权限防护、D.9 Tool Registry。

---

## 3. 公开题库和面试教程

### 3.1 Interview AiBox：RAG System Design Interview Questions

- URL: <https://interviewaibox.co/en/blog/rag-system-design-interview-questions>
- 类型：公开面试教程
- 质量等级：B
- 标签：`rag`、`chunking`、`reranking`、`freshness`、`evals`、`failure-modes`
- 摘要：文章将 RAG 面试从“embedding + top-k + prompt”推进到 follow-up：chunking、reranking、freshness、evaluation、failure modes、metadata、cost 和 latency。
- 可改写题目：
  - 设计 RAG 系统时，chunk size 如何影响召回、上下文连贯性和成本？
  - reranking 什么时候值得引入？如何评估它带来的延迟成本？
  - 如果数据有 freshness 要求，索引和 fallback 如何设计？
  - RAG 系统最常见的失败模式有哪些？
- 可映射附录章节：D.3 企业知识库问答 Agent。

### 3.2 AgenticCareers：Top 10 Interview Questions for LLM Engineer Jobs

- URL: <https://agenticcareers.co/blog/top-10-interview-questions-for-llm-engineer-jobs>
- 类型：公开面试教程
- 质量等级：B
- 标签：`rag`、`prompt-injection`、`debugging`、`agent-failures`、`eval-strategy`
- 摘要：文章强调 LLM 工程面试更重视 AI 系统设计、Agent failure debugging、evaluation strategy 和语言模型行为心智模型。题型覆盖 RAG、web browsing agent 的 prompt injection 和 Agent 退化调试。
- 可改写题目：
  - 一个 Agent 上周正常、本周输出变差，如何系统性 debug？
  - 浏览网页的 Agent 如何处理来自网页内容的 prompt injection？
  - 如何评估 RAG 系统中的 retrieval quality，而不是只看最终回答？
- 可映射附录章节：D.3、D.11、D.12。

### 3.3 System Design Interview Handbook：Generative AI & LLM System Design

- URL: <https://www.systemdesigninterview.com/guides/system-design-interview-handbook/72-generative-ai-llm-system-design>
- 类型：系统设计教程
- 质量等级：B
- 标签：`llm-system-design`、`rag-ingestion`、`vector-db`、`chunking`、`reranking`
- 摘要：教程将生产 RAG 拆成 ingestion pipeline、chunk processing、embedding service、vector database、retrieval service 和 LLM service，并讨论 chunking 对检索质量的影响。
- 可改写题目：
  - 请画出生产级 RAG 系统的 ingestion 和 query-time 两条链路。
  - 如何设计文档更新、重建索引和版本回滚？
  - 什么时候选择 semantic chunking，而不是固定长度 chunking？
- 可映射附录章节：D.3 企业知识库问答 Agent。

### 3.4 GitHub：llmgenai / LLMInterviewQuestions

- URL: <https://github.com/llmgenai/LLMInterviewQuestions>
- 类型：GitHub 公开题库
- 质量等级：B
- 标签：`llm-basics`、`rag`、`embedding`、`vector-db`、`agent-system`、`fine-tuning`
- 摘要：仓库包含 100+ LLM 面试题，分类覆盖 prompt engineering、RAG、chunking、embedding、vector search、retrieval metrics、hybrid search、fine-tuning 和 agent-based system。适合作为题型枚举，不适合作为生产设计答案来源。
- 可改写题目：
  - RAG 和 fine-tuning 的边界是什么？给出三个业务场景判断。
  - 如何 benchmark embedding model 在私有数据上的效果？
  - 如果 RAG 系统 retrieval 不准，你会按什么顺序排查？
  - ReAct、Plan-and-Execute 和 function calling 的差异是什么？
- 可映射附录章节：D.2、D.3、D.7、D.9。

### 3.5 Rubduck：AI Interview Question Bank

- URL: <https://rubduck.ai/questions>
- 类型：公开题库入口
- 质量等级：B
- 标签：`ai-system-design`、`rag`、`agents`、`llm-evaluation`、`prompt-engineering`
- 摘要：题库按 AI Agents & Tool Use、AI System Design、LLM Evaluation & Ops、Prompt Engineering、RAG & Retrieval 分类。适合作为附录题型覆盖度检查表。
- 可改写题目：
  - 你的作品集是否覆盖 RAG、Agent Tool Use、LLM Evaluation、Prompt Engineering 和系统设计？
  - 如果只准备三个系统设计题，如何覆盖最多面试信号？
- 可映射附录章节：D.19 面试前检查清单。

---

## 4. 社区面试反馈和趋势

### 4.1 Reddit：Entry-level GenAI / LLM architect role

- URL: <https://www.reddit.com/r/learnmachinelearning/comments/1sp58nv/what_kind_of_interview_questions_should_i_expect/>
- 类型：社区面试反馈
- 质量等级：C
- 标签：`interview-experience`、`rag`、`evals`、`logging`、`cost-latency`、`portfolio`
- 摘要：帖子反馈 GenAI / LLM 架构类面试会偏系统设计和 trade-off，而不只是算法题。常见准备方向包括 RAG、hallucination、evals、logging、retries、cost / latency、vector DB、chunking、embeddings 和 portfolio app。
- 可改写题目：
  - 初级 AI Engineer 如何用 2-3 个项目证明自己理解 evals、guardrails 和 observability？
  - 系统设计面试中如何讲成本、延迟和质量权衡？
- 可映射附录章节：D.14 作品集项目材料包、D.19 面试前检查清单。

### 4.2 Reddit：AI engineer interview questions

- URL: <https://www.reddit.com/r/ArtificialInteligence/comments/1nybfr8/ai_engineer_interview_questions/>
- 类型：社区面试反馈
- 质量等级：C
- 标签：`interview-loop`、`take-home`、`agent-assessment`、`applied-system-design`
- 摘要：帖子中出现了 AI engineer loop 的常见形态：Python coding、LLM system design、如何 ship 真实系统、agent take-home assessment、applied system design 等。
- 可改写题目：
  - 如果面试要求做一个 Agent take-home，你的 README 需要证明哪些生产能力？
  - 如何把一个演示型 Agent 项目讲成可以上线的系统？
- 可映射附录章节：D.15 GitHub README 模板、D.16 面试表达脚本。

### 4.3 Reddit：How are you evaluating agentic systems in production?

- URL: <https://www.reddit.com/r/LLMDevs/comments/1ryzv71/how_are_you_actually_evaluating_agentic_systems/>
- 类型：社区工程讨论
- 质量等级：C
- 标签：`agent-evals`、`multi-turn`、`synthetic-simulation`、`llm-as-judge`、`regression`
- 摘要：讨论集中在 agentic workflow 的评估盲点：手工测试不足、LLM-as-judge 有噪声、多轮路径难覆盖、synthetic user simulation 可以捕捉部分边界，但不能替代真实流量。
- 可改写题目：
  - 多轮客服 Agent 如何做 synthetic user simulation？
  - LLM-as-judge 用在 Agent eval 时有哪些噪声和校准问题？
  - regression eval 能防哪些问题，防不了哪些问题？
- 可映射附录章节：D.8 Agent Evals 平台。

### 4.4 Reddit：Interview questions to check a team’s RAG / LLM maturity

- URL: <https://www.reddit.com/r/dataengineeringjobs/comments/1rbgqix/interview_tip_for_de_roles_questions_to_check_a_team%E2%80%99s_rag_llm_maturity/>
- 类型：社区工程讨论
- 质量等级：C
- 标签：`rag-maturity`、`debugging`、`data-quality`、`retrieval-quality`
- 摘要：帖子关注反向面试问题：团队如何判断 RAG 错误来自 data、embedding、retriever 还是 prompt。这对候选人也很有价值，因为它直接对应 RAG debug 的层次化思维。
- 可改写题目：
  - RAG 系统答错时，你如何判断问题出在数据、embedding、retriever、reranker、prompt 还是模型？
  - 设计 RAG observability 时，每一层应记录哪些信号？
- 可映射附录章节：D.3、D.12。

---

## 5. 主题题库：可原创改写的面试题

### 5.1 RAG / Agentic RAG

- 设计一个企业知识库问答系统，要求权限和源系统一致，答案必须带引用。
- 文档很长、格式复杂、包含表格和列表，如何设计 chunking pipeline？
- RAG 系统答错时，如何判断是 missed retrieval、wrong grounding、stale data、query rewrite 还是 generation 的问题？
- 如何设计 freshness-aware retrieval？
- hybrid search 和 dense-only retrieval 如何取舍？
- reranking 提升质量但增加延迟，你如何决定是否上线？
- 如何评估 retrieval recall、answer faithfulness 和 citation support？
- 如何处理多租户 RAG 的权限过滤和 trace 脱敏？

### 5.2 Agent Harness / Coding Agent

- 设计一个 Coding Agent 的执行循环：它如何读文件、编辑文件、运行测试、记录 trace 并输出 diff？
- Agent 想执行危险 shell 命令，harness 应该如何拦截和审批？
- 如何保护用户未提交改动不被 Agent 覆盖？
- 一个 Coding Agent solve rate 下降，你如何做 ablation？
- 如何设计 checkpoint，让长任务可以恢复、暂停和人工接管？
- 如何评估 Agent 是真的修复了问题，而不是碰巧让测试通过？

### 5.3 Tool Calling / MCP / Tool Registry

- 设计企业 Tool Registry，让多个 Agent 共享内部 API、MCP server 和 SaaS connector。
- 每个工具应该包含哪些 metadata：description、schema、权限、风险等级、版本、审计？
- Agent 经常选错工具，如何判断是 tool description、tool overlap、prompt 还是 planner 的问题？
- 高风险写工具如何设计 approval workflow？
- 如何为工具调用设计 idempotency key、rate limit 和 audit log？
- tool schema 变更如何灰度发布？

### 5.4 Evals

- 设计一个 Agent Evals 平台，支持 capability eval、regression eval 和线上抽样。
- Agent eval case 应该包含哪些字段？
- 什么时候用 code-based grader、model-based grader、human grader？
- 如何评估完整 trace，而不是只评估 final answer？
- 如何从生产 trace 自动挖掘新的 eval case？
- LLM-as-a-judge 评分不稳定时，如何校准？
- 如何用 eval 判断一次 prompt 修改是改善还是回退？

### 5.5 Observability

- 设计 LLM / Agent observability 平台，trace、span 和 metrics 分别记录什么？
- 如何定位一次失败来自 retrieval、tool、model、handoff、guardrail 还是 workflow？
- trace 中包含用户隐私和商业数据，如何脱敏、采样和做访问控制？
- 如何监控 instruction drift、context retrieval loss 和行为退化？
- observability 如何和 eval 形成闭环？
- 如何控制 trace 存储成本？

### 5.6 Guardrails / Security

- Web browsing Agent 如何防 prompt injection？
- input guardrail、output guardrail 和 tool guardrail 各自适合拦什么？
- 为什么权限检查不能交给模型自己判断？
- 检索文档、工具输出和用户输入分别属于什么信任域？
- 如何测试一个 RAG 系统不会泄露其他租户数据？
- blocking guardrail 和 parallel guardrail 如何在延迟、成本和副作用之间权衡？

### 5.7 Multi-agent

- 设计一个 multi-agent research system，如何分工、合并、引用和停止？
- 什么情况下不应该使用多 Agent？
- 子 Agent 重复搜索、互相干扰或无限扩张时如何限制？
- 如何根据 query complexity 分配 agent 数量和工具预算？
- Citation Agent 应该如何验证报告中的 claim？
- 如何观测和评估 multi-agent coordination failure？

### 5.8 Portfolio / Take-home

- 如果你只有一个 RAG demo，如何把它包装成生产级作品集？
- README 中如何展示 architecture、evals、trace 和 failure modes？
- 面试官问“这个项目真的上线了吗”，你如何诚实表达 demo 和 production 的边界？
- 如何讲一个 Agent 项目的失败复盘，而不是只说“优化了 prompt”？
- 如何用 2 分钟、5 分钟和 15 分钟分别介绍同一个 Agent 项目？

---

## 6. 对附录 D 的后续扩展建议

短期可以直接补充：

- 在 D.8 中加入 “eval case schema 来源于 task / grader / transcript / outcome” 的解释；
- 在 D.12 中加入 “behavior monitoring” 相关指标，如 instruction drift 和 context retrieval loss；
- 在 D.3 中加入 RAG debug 分层：data → chunking → embedding → retrieval → rerank → context → generation；
- 在 D.9 中补一个 tool description eval 的小例子；
- 在 D.16 中加入 take-home 项目讲法。

中期可以扩展为一个独立章节：

```text
附录E LLM / Agent 面试题库索引
  E.1 RAG 高频题
  E.2 Agent Harness 高频题
  E.3 Evals 高频题
  E.4 Observability 高频题
  E.5 Security 高频题
  E.6 Portfolio 高频题
```

---

## 7. 采集质量备注

- 官方岗位和工程文章具有最高参考价值，因为它们直接反映团队正在招聘和构建的能力面。
- 公开题库适合补题型覆盖度，但需要重新组织成系统设计问题，避免变成术语问答。
- Reddit 等社区内容只作为趋势参考，不能作为事实来源；其中有价值的是“面试形式”和“常被追问的 trade-off”。
- 本素材库后续可以定期更新，但附录正文应继续保持原创和结构化，而不是堆链接。

---

## 8. 题单参考来源

第一遍学习建议先“广覆盖”，不要过早筛题。下面这些来源用于构造后面的题海题单，题目均做了中文改写和主题重组。

| 来源 | URL | 适合提取的题型 | 质量备注 |
| --- | --- | --- | --- |
| OpenAI Agents SDK：Tracing | <https://openai.github.io/openai-agents-python/tracing/> | trace、span、tool call、handoff、guardrail span、debugging | A 级，适合作为 Agent Observability 题单主参考 |
| OpenAI Agents SDK：Guardrails | <https://openai.github.io/openai-agents-python/guardrails/> | input guardrail、output guardrail、tool guardrail、tripwire | A 级，适合安全和工具调用边界题 |
| Anthropic：Demystifying evals for AI agents | <https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents> | task、trial、grader、transcript、outcome、eval harness | A 级，适合 Agent eval 题 |
| Anthropic：Building Effective Agents | <https://www.anthropic.com/engineering/building-effective-agents> | workflow vs agent、routing、parallelization、evaluator-optimizer、orchestrator-workers | A 级，适合 Agent 架构题 |
| Anthropic：Multi-agent Research System | <https://www.anthropic.com/engineering/multi-agent-research-system> | lead agent、subagent、citation agent、多 Agent 协作 | A 级，适合研究型 Agent 题 |
| GitHub：KalyanKS-NLP/RAG-Interview-Questions-and-Answers-Hub | <https://github.com/KalyanKS-NLP/RAG-Interview-Questions-and-Answers-Hub> | RAG、chunking、retrieval、reranking、RAG metrics | B+，RAG 题量很大，适合刷题 |
| GitHub：llmgenai/LLMInterviewQuestions | <https://github.com/llmgenai/LLMInterviewQuestions> | LLM 基础、RAG、embedding、vector DB、Agent、prompt hacking | B，覆盖面广，答案需自行校验 |
| GitHub：sreekanth-madisetty/Awesome-LLM-Interview-Questions | <https://github.com/sreekanth-madisetty/Awesome-LLM-Interview-Questions> | RAG、agents、fine-tuning、quantization、pretraining | B，分类清楚，可做学习索引 |
| GitHub：DolbyUUU/Awesome-LLM-Interview-Questions-and-Answers | <https://github.com/DolbyUUU/Awesome-LLM-Interview-Questions-and-Answers> | 中文 LLM / Agent / RAG / MCP / Tool Use / 项目经验 | B-，贴近国内面试，但答案需要复核 |
| Rubduck AI Question Bank | <https://rubduck.ai/questions> | AI system design、agents、RAG、LLM evaluation | B，适合作为题型覆盖检查 |
| OpenAI API：Structured Outputs | <https://developers.openai.com/api/docs/guides/structured-outputs> | JSON Schema、function calling、response format、schema adherence | A，适合结构化输出和工具边界题 |
| OpenAI API：Prompt Caching | <https://developers.openai.com/api/docs/guides/prompt-caching> | prompt cache、成本、延迟、静态前缀、缓存保持 | A，适合成本优化和长上下文题 |
| Model Context Protocol 2025-06-18 Specification | <https://modelcontextprotocol.io/specification/2025-06-18> | host/client/server、resources、prompts、tools、sampling、roots、elicitation、安全原则 | A，适合 MCP 和 Tool Registry 题 |
| MCP Server Tools | <https://modelcontextprotocol.io/specification/2025-06-18/server/tools> | tools/list、tools/call、inputSchema、outputSchema、annotations、structuredContent | A，适合工具 schema 和安全题 |
| MCP Server Resources | <https://modelcontextprotocol.io/specification/2025-06-18/server/resources> | resources/list、resources/read、templates、subscribe、listChanged、annotations | A，适合上下文供给和资源权限题 |
| MCP Server Prompts | <https://modelcontextprotocol.io/specification/2025-06-18/server/prompts> | prompts/list、prompts/get、arguments、prompt messages、多模态内容、安全校验 | A，适合 prompt catalog 和工作流模板题 |
| OWASP Top 10 for LLM Applications | <https://owasp.org/www-project-top-10-for-large-language-model-applications/> | prompt injection、insecure output handling、sensitive disclosure、excessive agency、model theft | A，适合安全和风险建模题 |
| OpenTelemetry GenAI Semantic Conventions | <https://opentelemetry.io/docs/specs/semconv/gen-ai/> | GenAI spans、events、metrics、model spans、agent spans、MCP spans | A，适合 observability 标准化题 |
| LangSmith Evaluation | <https://docs.langchain.com/langsmith/evaluation> | offline eval、online eval、dataset、evaluator、production trace feedback loop | A-，产品文档，适合 eval 平台题 |

---

## 9. 第一遍题海战术题单

使用方法：

```text
第一遍：只看题目，标记“会 / 不会 / 模糊”；
第二遍：每类挑 10 道写 5-8 行答案；
第三遍：把系统设计题改写成架构图 + 追问 + trade-off；
第四遍：把高频题回填到附录 D 的正式题单。
```

### 9.1 LLM 基础与 Prompt Engineering

1. LLM 和传统 NLP 模型的核心区别是什么？
2. 自回归语言模型为什么适合做文本生成？
3. token、context window、temperature、top-p 分别影响什么？
4. temperature 和 top-p 同时调整时会发生什么？
5. stop sequence 的典型用途是什么？
6. 为什么同一个 prompt 多次调用可能得到不同结果？
7. system prompt、developer prompt、user prompt 的职责如何区分？
8. few-shot prompting 为什么能提升特定任务表现？
9. chain-of-thought 适合哪些任务？什么时候不应该暴露推理过程？
10. structured output 相比自然语言输出有什么工程价值？
11. JSON mode / schema validation 能解决哪些问题，不能解决哪些问题？
12. prompt template 如何版本化和回滚？
13. 如何评估一次 prompt 修改是否真的变好？
14. prompt 变长会带来哪些成本、延迟和质量风险？
15. 长上下文模型是否意味着不需要 RAG？
16. 模型幻觉有哪些类型？事实性幻觉和格式性幻觉如何区分？
17. 如何用 prompt 降低幻觉？这种方法的边界是什么？
18. 为什么“让模型不要胡说”不是可靠 guardrail？
19. 如何在 prompt 中表达任务边界、拒答条件和输出格式？
20. 面向工具调用的 prompt 和面向问答的 prompt 有什么不同？
21. 如何设计 prompt，让模型先澄清需求而不是直接执行？
22. 如何处理用户输入过短、模糊或多意图的问题？
23. 如何在 prompt 中注入业务规则，同时保持可维护性？
24. Prompt Engineering、Context Engineering、Harness Engineering 的区别是什么？
25. 如果 prompt 在测试集上变好、线上变差，你如何排查？

### 9.2 RAG 基础架构

26. 为什么需要 RAG？它解决了 LLM 的哪些问题？
27. RAG 和 fine-tuning 的边界是什么？
28. 企业知识库问答系统的 ingestion pipeline 怎么设计？
29. 企业知识库问答系统的 query-time pipeline 怎么设计？
30. chunking 为什么重要？chunk 太大和太小分别有什么问题？
31. 固定长度 chunking、语义 chunking、结构化 chunking 如何取舍？
32. PDF、表格、图片和代码文档如何做 chunking？
33. chunk overlap 的作用是什么？过大有什么副作用？
34. chunk metadata 应该包含哪些字段？
35. 文档更新后如何增量重建索引？
36. 如何处理被删除或撤权的文档？
37. embedding model 如何选择？
38. 如何 benchmark embedding model 在企业私有数据上的效果？
39. 向量数据库和传统数据库分别负责什么？
40. 向量检索为什么可能召回语义相关但业务无关的片段？
41. keyword search、vector search、hybrid search 的优缺点是什么？
42. 什么场景必须使用 hybrid search？
43. BM25 在 RAG 中仍然有什么价值？
44. reranker 解决了什么问题？
45. reranker 为什么会增加延迟？如何优化？
46. top-k 取太大和太小分别有什么风险？
47. 如何合并来自多个数据源的检索结果？
48. 如何处理多跳问题和多意图问题？
49. query rewrite 有什么价值？什么时候会伤害检索？
50. HyDE 的核心思路是什么？适合什么场景？
51. 如何设计 freshness-aware retrieval？
52. 对时效性强的数据，索引延迟如何控制？
53. RAG 是否应该把聊天历史也纳入检索？
54. 如何在 RAG 中处理用户权限？
55. 为什么权限过滤必须在生成前完成？
56. 如何设计多租户 RAG 的数据隔离？
57. RAG 如何返回可验证引用？
58. citation 应该精确到文档、段落、句子还是行号？
59. 证据不足时 RAG 应该如何拒答？
60. 多个来源互相冲突时如何生成答案？
61. RAG 如何处理过期文档？
62. 如何检测 retrieved context 是否支持 final answer？
63. 如何为 RAG 增加用户反馈闭环？
64. RAG 系统的缓存应该缓存 query、retrieval result 还是 final answer？
65. RAG 中哪些部分适合异步化或批处理？
66. 如何估算 RAG 系统一次请求的成本？
67. 如何降低 RAG 的 p95 延迟？
68. 如何处理“召回很多但答案仍然错”的问题？
69. 如何处理“召回正确但模型没用上”的问题？
70. 如何处理“模型答案正确但引用不支持”的问题？
71. 如何设计 RAG 的 fallback：搜索失败、模型失败、工具失败分别怎么办？
72. RAG debug 时如何按 data → chunking → embedding → retrieval → rerank → generation 分层排查？
73. 如果用户问一个源系统中不存在的问题，系统应该怎么表现？
74. 如何让 RAG 支持跨语言查询？
75. 如何让 RAG 支持代码仓库问答？

### 9.3 RAG Evaluation 与检索指标

76. RAG eval 应该评估哪些层：retriever、reranker、generator、citation？
77. retrieval recall 和 answer correctness 有什么区别？
78. context precision 衡量什么？
79. context recall 衡量什么？
80. faithfulness 衡量什么？
81. response relevancy 衡量什么？
82. citation support rate 如何计算？
83. 如果 Context Recall 高但 Faithfulness 低，说明什么？
84. 如果 Context Precision 高但答案错，可能是什么原因？
85. 如果 Recall@10 高但 Precision@10 低，会影响什么？
86. MRR、MAP、NDCG 分别适合什么检索评估场景？
87. 为什么只看 top-1 accuracy 不够？
88. 如何构造 RAG golden dataset？
89. 标准答案、标准证据和禁止证据分别有什么作用？
90. 如何评估权限泄露？
91. 如何评估拒答是否合理？
92. 如何评估 reranker 是否值得上线？
93. 如何从线上 bad case 生成 RAG regression case？
94. 如何区分 capability eval 和 regression eval？
95. 如何评估 RAG 对长文档、表格、代码块的支持能力？
96. 如何做中文、英文、混合语言 RAG 的评估？
97. 如何评估多轮 RAG 问答的上下文连续性？
98. 如何人工标注 RAG eval case？
99. LLM-as-a-judge 评估 RAG 有哪些风险？
100. 如何校准 LLM judge 和人工评审的一致性？
101. 如何避免 eval dataset 被 prompt 或系统过拟合？
102. 如何设计 RAG eval dashboard？
103. RAG 线上指标和离线 eval 指标如何对应？
104. 如何给 RAG 系统设计发布门禁？
105. RAG 质量、成本、延迟三者如何一起评估？

### 9.4 Agent 基础与架构

106. 什么是 Agent？它和普通 chatbot 的区别是什么？
107. Agent 和 workflow 的边界是什么？
108. 什么情况下不应该使用 Agent？
109. ReAct 的核心思想是什么？
110. Plan-and-Execute 的核心思想是什么？
111. Routing workflow 适合哪些问题？
112. Parallelization workflow 适合哪些问题？
113. Evaluator-Optimizer workflow 适合哪些问题？
114. Orchestrator-Workers workflow 适合哪些问题？
115. Agent 为什么需要状态机？
116. Agent Runtime 应该包含哪些模块？
117. Agent Harness 是什么？它和模型有什么边界？
118. Agent loop 中 plan、act、observe、reflect、stop 分别做什么？
119. Agent 的停止条件如何设计？
120. 如何防止 Agent 无限循环？
121. Agent 什么时候需要 memory？
122. short-term memory 和 long-term memory 有什么区别？
123. memory 写入为什么需要审核或可撤销？
124. 如何设计用户可见、可编辑、可删除的 memory？
125. Agent 如何处理长任务？
126. 长任务如何 checkpoint？
127. Agent 执行失败后如何恢复？
128. Agent 如何做 retry？哪些错误不应该 retry？
129. Agent 如何向用户请求澄清？
130. Agent 如何把任务交给人工？
131. Human-in-the-loop 的审批点如何设计？
132. 如何设计 Agent 的权限模型？
133. Agent 如何处理多个用户、多个租户、多个身份？
134. Agent 如何记录每一步决策？
135. Agent 什么时候应该输出多个假设而不是单个结论？
136. Agent 如何处理工具返回冲突？
137. Agent 如何处理工具超时？
138. Agent 如何判断自己没有足够信息？
139. Agent 如何避免过度自信？
140. 如何设计一个客服 Agent？
141. 如何设计一个生产告警诊断 Agent？
142. 如何设计一个代码审查 Agent？
143. 如何设计一个研究型 Agent？
144. 如何设计一个个人知识管理 Agent？
145. 如何把一个 demo Agent 改造成生产级系统？

### 9.5 Coding Agent / Agent Harness

146. Coding Agent 如何理解用户任务？
147. Coding Agent 如何选择需要阅读的文件？
148. 大型代码仓库中，Agent 如何做 context selection？
149. Agent 如何避免把整个仓库塞进上下文？
150. Agent 如何编辑文件并保持最小 diff？
151. Agent 修改前如何检查用户未提交改动？
152. Agent 如何运行测试并解释失败？
153. 测试失败时如何区分代码问题、环境问题和 flaky test？
154. Agent 什么时候应该新增测试？
155. Coding Agent 如何做 red-green verification？
156. Agent 如何生成可审查的 PR summary？
157. Agent 如何避免执行危险命令？
158. sandbox 应该限制哪些资源：文件、网络、进程、环境变量？
159. Agent 如何处理依赖安装和网络访问？
160. Agent 如何保护 `.env`、密钥和本地凭据？
161. Agent 如何处理跨文件重构？
162. Agent 如何处理大型迁移任务？
163. Agent 如何暂停并等待用户确认？
164. Agent 如何记录工具调用 trace？
165. 如何评估 Coding Agent 的 solve rate？
166. 如何评估 Coding Agent 的 patch quality？
167. 如何评估 Agent 是否真的理解代码而不是随机改？
168. 如何比较不同模型在 Coding Agent 上的表现？
169. 如何比较不同 harness 版本？
170. Coding Agent 最常见的失败模式有哪些？

### 9.6 Tool Calling / Function Calling / MCP

171. Tool Calling 解决了 LLM 的什么问题？
172. function calling 和普通文本输出有什么区别？
173. 工具 schema 应该如何设计？
174. 工具描述应该写给人看还是写给模型看？
175. 一个工具应该大而全还是小而专？
176. 工具职责重叠会导致什么问题？
177. 如何评估 Agent 是否选对工具？
178. 如何评估工具参数是否正确？
179. 工具返回结果如何进入上下文？
180. 工具返回的内容是否可信？
181. 如何处理工具返回敏感数据？
182. 如何处理工具调用失败、超时和限流？
183. 写工具为什么需要 idempotency key？
184. 高风险工具如何设计审批？
185. 工具风险等级如何划分？
186. read-only 工具和 write 工具的安全策略有什么不同？
187. destructive tool 如何管控？
188. Tool Registry 应该存哪些 metadata？
189. Tool Gateway 应该负责哪些能力？
190. MCP 解决了什么集成问题？
191. MCP server 和普通内部 API 有什么区别？
192. 多个 Agent 共享工具时如何做鉴权和审计？
193. 工具 schema 变更如何兼容旧 Agent？
194. 工具调用日志如何进入 observability？
195. 如何发现工具描述导致的误用？

### 9.7 Agent Evals

196. 为什么 Agent eval 比普通 LLM eval 更难？
197. Agent eval 中 task、trial、transcript、outcome 分别是什么？
198. Agent eval case 应该包含哪些字段？
199. 为什么 Agent eval 需要看完整轨迹？
200. final answer 正确但工具路径错误，算不算通过？
201. 工具路径正确但 final answer 错误，如何评分？
202. code-based grader 适合哪些场景？
203. model-based grader 适合哪些场景？
204. human grader 适合哪些场景？
205. 如何设计 Agent regression suite？
206. 如何从线上 trace 挖掘 eval case？
207. 如何评估 Agent 的 tool-call correctness？
208. 如何评估 Agent 的 unsafe action rate？
209. 如何评估 Agent 的 human override rate？
210. 如何评估 Agent 的 task completion rate？
211. 如何评估 Agent 的 multi-turn consistency？
212. 如何评估 Agent 的 cost per successful task？
213. 如何评估 Agent 是否过度调用工具？
214. 如何评估 Agent 是否漏调用关键工具？
215. 如何评估 Agent 是否遵守审批策略？
216. 如何做 eval dataset 版本化？
217. 如何比较模型版本、prompt 版本、tool 版本和 harness 版本？
218. 如何做 canary release 的 eval gate？
219. 如何防止 eval 被刷题式优化？
220. 如何设计 Agent eval report？

### 9.8 Observability / Tracing

221. 为什么 Agent 需要 tracing？
222. trace 和 log 的区别是什么？
223. trace 和 span 的关系是什么？
224. 一个 Agent trace 应该包含哪些 span？
225. generation span 应该记录什么？
226. tool span 应该记录什么？
227. retrieval span 应该记录什么？
228. guardrail span 应该记录什么？
229. handoff span 应该记录什么？
230. custom span 适合记录什么？
231. 如何用 trace 定位 retrieval failure？
232. 如何用 trace 定位 tool failure？
233. 如何用 trace 定位 generation failure？
234. 如何用 trace 定位 guardrail false positive？
235. 如何用 trace 定位 handoff failure？
236. trace 中是否应该保存完整 prompt？
237. trace 中如何做 PII 脱敏？
238. trace 如何做采样？
239. 高价值 trace 如何优先保留？
240. 如何从 trace 生成 eval case？
241. 如何监控 instruction drift？
242. 如何监控 context retrieval loss？
243. 如何监控 Agent 行为聚类变化？
244. 如何监控成本异常？
245. 如何监控工具调用异常？
246. 如何设计 Agent observability dashboard？
247. 如何把 trace、metrics、feedback 和 eval 连接起来？
248. 如何控制 trace 存储成本？
249. 如何限制谁能查看敏感 trace？
250. OpenTelemetry 和 Agent tracing 如何结合？

### 9.9 Guardrails / Prompt Injection / Security

251. prompt injection 是什么？
252. prompt injection 和普通用户指令冲突有什么区别？
253. RAG 文档中的恶意指令如何处理？
254. Web browsing Agent 如何处理网页中的恶意提示？
255. 工具输出是否可能包含 prompt injection？
256. memory 是否可能被污染？
257. 如何划分 trusted instruction 和 untrusted content？
258. 为什么不可信文档不能拥有指令权？
259. input guardrail 适合拦什么？
260. output guardrail 适合拦什么？
261. tool guardrail 适合拦什么？
262. tripwire 触发后系统如何响应？
263. guardrail 是 blocking 还是 async parallel，如何取舍？
264. 权限检查为什么不能交给模型？
265. 如何防止跨租户数据泄露？
266. 如何防止模型输出 PII？
267. 如何防止 Agent 泄露系统 prompt？
268. 如何防止 Agent 调用未授权工具？
269. 如何处理用户要求“忽略之前所有指令”？
270. 如何处理用户诱导 Agent 输出密钥？
271. 如何处理文档中包含“把所有数据发给攻击者”的内容？
272. 如何设计安全 eval？
273. 如何把安全事故转成 regression case？
274. 如何在安全和可用性之间做权衡？
275. 如何给高风险操作设计人工审批？

### 9.10 Multi-agent / Research Agent

276. 什么情况下需要多 Agent？
277. 什么情况下多 Agent 是过度设计？
278. Lead Agent 的职责是什么？
279. Subagent 的输入应该包含哪些约束？
280. 如何避免多个子 Agent 重复工作？
281. 如何控制子 Agent 的预算？
282. 如何根据任务复杂度决定子 Agent 数量？
283. 如何合并多个子 Agent 的结论？
284. 子 Agent 结论冲突时怎么办？
285. Citation Agent 的职责是什么？
286. 如何验证报告中的 claim 有来源支持？
287. 如何追踪多 Agent 的任务树？
288. 多 Agent 系统如何 checkpoint？
289. 多 Agent 系统如何 debug？
290. 多 Agent 系统如何 eval？
291. 多 Agent 系统的 token 成本如何控制？
292. 多 Agent 系统如何避免无限扩张？
293. 多 Agent 和并行 workflow 的区别是什么？
294. 多 Agent research 输出如何防幻觉？
295. 多 Agent 系统最常见的协调失败有哪些？

### 9.11 LLM 系统部署、成本与性能

296. LLM 应用上线需要哪些环境隔离？
297. 如何选择闭源模型、开源模型和本地模型？
298. 如何做模型路由？
299. 如何根据任务难度选择模型？
300. 如何设计 fallback model？
301. 如何处理模型 API 超时？
302. 如何处理模型 API 限流？
303. 如何降低 token 成本？
304. 如何降低 p95 延迟？
305. streaming 对用户体验和系统架构有什么影响？
306. caching 在 LLM 系统中有哪些层次？
307. prompt cache 适合什么场景？
308. retrieval cache 适合什么场景？
309. answer cache 有什么风险？
310. batch inference 适合什么场景？
311. 如何估算容量？
312. 如何做 rate limiting？
313. 如何做 quota 管理？
314. 如何做 tenant-level 成本归因？
315. 如何监控模型质量退化？
316. 模型版本升级如何灰度？
317. 如何设计模型回滚？
318. 如何做线上 A/B？
319. 如何处理供应商 API 故障？
320. 如何设计 LLM 系统的 SLO？

### 9.12 作品集与项目经验

321. 如何把一个 RAG demo 包装成生产级作品集？
322. 如何把一个客服 Agent demo 包装成生产级作品集？
323. 如何把一个 Coding Agent demo 包装成生产级作品集？
324. 项目 README 应该包含哪些章节？
325. 架构图中必须展示哪些生产组件？
326. 如何展示 eval dataset？
327. 如何展示 eval report？
328. 如何展示 trace 示例？
329. 如何展示 failure postmortem？
330. 如何诚实说明 demo 和 production 的差距？
331. 如何讲项目中的一次失败？
332. 如何说明自己具体贡献？
333. 如何量化 Agent 项目结果？
334. 没有线上数据时如何设计离线指标？
335. take-home 项目如何体现工程深度？
336. 面试官问“为什么不用普通 workflow”，你怎么回答？
337. 面试官问“为什么不用 fine-tuning”，你怎么回答？
338. 面试官问“为什么不用长上下文直接塞文档”，你怎么回答？
339. 面试官问“怎么证明效果变好”，你怎么回答？
340. 面试官问“怎么防止出事故”，你怎么回答？

### 9.13 反向面试问题

341. 你们如何定义 Agent 项目的成功指标？
342. 你们线上是否有 Agent trace 和 eval 闭环？
343. 你们如何区分模型问题、检索问题和工具问题？
344. 你们有没有 regression eval suite？
345. 你们如何处理 prompt injection？
346. 你们的工具调用有没有风险分级和审批？
347. 你们如何做 RAG 权限过滤？
348. 你们如何处理线上用户反馈？
349. 你们如何评估新模型上线风险？
350. 你们团队更看重 demo 速度还是生产可靠性？

### 9.14 题海答案速记

> 这一节用于第一遍快速过题。答案故意写成“面试口头速答”，不是完整讲稿；真正的系统设计题可以再展开成架构图、指标、失败模式和 trade-off。

#### 9.14.1 LLM 基础与 Prompt Engineering

1. LLM 是通用生成模型，传统 NLP 多是任务专用模型；关键差异在预训练规模、上下文学习和生成能力。
2. 自回归模型按 token 条件概率逐步预测下一个 token，天然适合连续文本生成。
3. token 是计算单位；context window 决定可见上下文；temperature 控制随机性；top-p 控制候选概率质量。
4. 两者都提高随机性时输出更发散；通常不要同时大幅调高，先固定一个再调另一个。
5. 用于控制生成停止点，例如 JSON 结束、分隔符、对话轮次和工具参数边界。
6. 采样策略、temperature、top-p、服务端非确定性和上下文细微差异都会导致输出不同。
7. system 定总规则，developer 定应用策略，user 提具体任务；权限和优先级应从高到低。
8. few-shot 提供任务示例和输出分布，让模型在上下文中学习格式、边界和风格。
9. 适合复杂推理和分解任务；涉及隐私、安全或最终用户场景时不应暴露完整推理过程。
10. 便于解析、校验、自动化执行和回归测试，是工程系统接入 LLM 的关键。
11. 能约束格式和字段，不能保证事实正确、业务合规或工具动作安全。
12. 把模板、变量、模型版本、eval 结果一起版本化，失败可回滚。
13. 用固定 eval 集和线上指标比较正确率、拒答率、格式错误率、成本和延迟。
14. 成本更高、延迟更大、噪声更多，还可能稀释关键指令。
15. 不是。长上下文解决“能放下”，RAG 解决“取对、更新、权限、引用和成本”。
16. 事实性幻觉是编造事实；格式性幻觉是格式不合规；也有引用、工具和权限幻觉。
17. 可通过引用约束、拒答条件、结构化输出降低幻觉，但不能替代检索、校验和 eval。
18. 模型指令不是安全边界；权限、工具策略和输出校验必须在系统层实现。
19. 明确角色、输入范围、拒答条件、输出 schema、示例和禁止行为。
20. 工具 prompt 要强调工具选择、参数完整性和风险边界；问答 prompt 更强调证据和表达。
21. 写清“信息不足先问问题”，并定义必须澄清的字段和可直接执行的条件。
22. 先做意图识别和槽位检查；不足则澄清，多意图则拆分或让用户选择。
23. 把规则外置成可版本化 policy/context，不把大量业务逻辑硬编码在 prompt 里。
24. Prompt 管指令表达，Context 管信息供给，Harness 管执行环境、工具、状态和验证。
25. 查数据分布、线上输入、模型版本、工具变化、采样参数和 eval 覆盖是否偏。

#### 9.14.2 RAG 基础架构

26. RAG 让模型接入外部知识，解决知识过期、私有知识、可引用和可更新问题。
27. RAG 适合知识注入和更新，fine-tuning 适合能力、风格和稳定格式学习。
28. 数据接入、解析、清洗、chunk、metadata、embedding、索引、权限同步和增量更新。
29. 查询理解、改写、权限过滤、检索、重排、证据打包、生成、引用和反馈。
30. chunk 决定检索粒度；太大噪声多，太小上下文断裂。
31. 固定长度简单，语义 chunk 保持语义，结构化 chunk 适合标题、表格、代码等文档结构。
32. 先做版面解析和结构抽取；表格保留 schema，图片做 OCR/说明，代码按函数/类切分。
33. overlap 保留跨块上下文；过大会增加重复、成本和检索噪声。
34. source、doc_id、section、timestamp、owner、permission、version、language、tenant、tags。
35. 用变更检测、增量 chunk、增量 embedding、索引版本和回滚机制。
36. 撤权/删除要同步 metadata 和索引，必要时 tombstone、重建索引并清缓存。
37. 看语言、领域、长短文本、成本、延迟、维度、召回 eval 和部署约束。
38. 用企业真实 query-doc pair，比较 Recall@k、MRR、NDCG 和人工相关性。
39. 向量库管语义相似搜索，传统数据库管事务、权限、metadata 和结构化过滤。
40. embedding 捕捉语义相似，不懂业务权限、时效、实体和上下文意图。
41. 关键词精确可解释，向量语义召回强，hybrid 同时兼顾术语和语义。
42. 专有名词、代码、订单号、缩写、法规条款、产品名等必须用 hybrid。
43. BM25 对关键词、稀有词、编号和精确短语很强，也便于解释。
44. reranker 对初召结果重新排序，提升 top-k 相关性和证据质量。
45. cross-encoder 要逐对计算，延迟高；可限候选数、缓存、轻量模型或按需启用。
46. top-k 太小漏证据，太大噪声多、成本高、生成更容易跑偏。
47. 统一 score、source 权重、去重、权限过滤、rerank 和 evidence packaging。
48. 多跳要分解子问题，多意图要拆任务或澄清，避免一次检索混杂多个目标。
49. query rewrite 扩展召回和消歧；错误改写会偏离原意或引入幻觉查询。
50. HyDE 先生成假想答案再检索，适合短 query 或概念性问题，但可能带偏。
51. 查询识别时效需求，优先检索新版本，metadata 过滤时间，并给过期提示。
52. 用流式索引、增量更新、CDC、队列和版本切换控制分钟级或小时级延迟。
53. 可以，但要摘要、过滤和权限处理；不要把全部历史无脑塞进检索。
54. 权限作为 metadata 和检索前过滤条件，生成前模型不能看到无权内容。
55. 生成后过滤已经泄露给模型，无法保证模型不受无权内容影响。
56. tenant_id 强过滤、独立索引或命名空间、独立密钥、trace 脱敏和权限回归测试。
57. Evidence Package 保存片段、来源、时间、权限决策和引用锚点。
58. 越精确越可信；面试中建议至少段落级，代码和法规最好行级。
59. 拒答并说明缺少证据，可建议用户补充信息或升级人工。
60. 明确指出冲突来源、更新时间和可信度，给出条件性结论或请求确认。
61. metadata 标记版本和更新时间，检索优先新文档，过期内容不用于关键答案。
62. 用 claim-to-evidence 检查、LLM judge、规则校验和人工抽检。
63. 收集点赞、纠错、无用原因和人工答案，进入 eval 和索引优化闭环。
64. 低风险可缓存 retrieval result；final answer 缓存要考虑权限、时效和个性化。
65. ingestion、embedding、rerank 批处理，跨源检索并行，摘要和日志异步化。
66. 估算 embedding、检索、rerank、输入/输出 token、缓存命中率和工具调用成本。
67. 并行检索、缓存、轻量 reranker、减少上下文、模型路由和流式输出。
68. 提升 rerank、context compression、证据选择和生成约束，减少无关上下文。
69. 检查 prompt 证据使用规则、上下文排序、引用要求和生成模型能力。
70. 加 citation verifier，要求每个 claim 对齐证据；不支持则拒答或修正。
71. 搜索失败给澄清/人工；模型失败重试/降级；工具失败兜底或排队。
72. 每层记录输入输出和指标，逐层定位召回、排序、上下文、生成或引用问题。
73. 明确拒答，不编造；可说明源系统无记录并建议查询路径。
74. 用多语 embedding、query translation、跨语言 rerank 和语言一致性 eval。
75. 按 repo、文件、函数、符号索引，结合 keyword、AST、依赖图和代码引用。

#### 9.14.3 RAG Evaluation 与检索指标

76. 分层评估 retriever 召回、reranker 排序、generator 忠实度和 citation 支持率。
77. recall 看证据是否找到了；answer correctness 看最终答案是否正确。
78. 检索结果中相关上下文占比和排序质量。
79. 标准答案所需证据有多少被检索出来。
80. 答案中的 claim 是否被上下文支持。
81. 回答是否针对用户问题，而不是答非所问。
82. 有引用且引用真正支持 claim 的答案比例。
83. 证据找到了但生成阶段没忠实使用，可能 prompt 或模型问题。
84. 证据相关但不完整、生成误解、答案需要推理或标准答案定义问题。
85. 下游上下文噪声大，模型可能被无关片段干扰，成本也上升。
86. MRR 看第一个相关结果，MAP 看多相关平均精度，NDCG 看排序和相关度等级。
87. 多证据、多跳和引用场景只看 top-1 会漏掉召回完整性。
88. 收集真实问题、标准答案、必需证据、禁止证据、用户权限和评分规则。
89. 标准答案评 correctness，标准证据评 recall/citation，禁止证据评泄露和误用。
90. 用低权限用户 query 测试高权限内容是否被检索、生成或记录。
91. 评估证据是否足够、拒答是否符合 policy、是否给出合理下一步。
92. 比较上线前后 NDCG/MAP、answer quality、延迟和成本。
93. 线上失败经脱敏、标注、归因后加入 regression suite。
94. capability eval 探索能力边界，regression eval 防止已修问题回退。
95. 分文档类型建 case，分别看解析、chunk、召回、引用和生成。
96. 分语言建 query-doc pair，评估跨语言召回、翻译损失和回答语言。
97. 多轮要评历史理解、引用连续性、纠错和上下文污染。
98. 给标注指南，标问题、答案、证据、权限、失败类型和置信度。
99. judge 可能偏、漂移、被提示影响，且对事实细节不一定可靠。
100. 用双评审、golden set、人工抽检、一致性指标和阈值校准。
101. 留出隐藏集，防止针对固定题调 prompt，并定期加入线上新样本。
102. 展示 recall、precision、faithfulness、citation、拒答、权限、延迟和成本。
103. 离线指标解释质量能力，线上指标捕捉真实分布和用户反馈。
104. 设置质量阈值、无安全回退、成本延迟阈值和关键回归零容忍。
105. 用多目标评估，按业务场景给权重，不单看准确率。

#### 9.14.4 Agent 基础与架构

106. Agent 能根据目标、上下文和工具反馈多步行动；chatbot 主要是对话生成。
107. workflow 路径固定可控，Agent 路径动态；优先 workflow，复杂开放任务再 Agent。
108. 任务简单、规则明确、高风险不可错、无评估兜底时不该用 Agent。
109. ReAct 交替进行 reasoning 和 acting，通过工具观察继续决策。
110. 先制定计划，再逐步执行，适合可分解任务但要处理计划失效。
111. Routing 适合把输入分发给不同模型、工具或流程的场景。
112. Parallelization 适合独立子任务并行，如多源搜索、多评审。
113. Evaluator-Optimizer 适合可迭代改进且有评分器的任务。
114. Orchestrator-Workers 适合复杂任务动态拆分给多个 worker。
115. 状态机让步骤、重试、审批、失败和恢复可控。
116. 需要 planner、context builder、tool dispatcher、memory、policy、trace 和 verifier。
117. Harness 是模型外的执行环境，负责工具、权限、状态、沙箱和验证。
118. plan 定策略，act 调工具，observe 看结果，reflect 修正，stop 判断完成。
119. 用目标完成、预算耗尽、风险触发、无信息、人工接管和最大步数。
120. 限步数、限预算、检测重复状态、失败熔断和人工确认。
121. 多轮、长期偏好、跨任务积累和个人化场景需要 memory。
122. short-term 是会话内状态，long-term 是跨会话持久知识/偏好。
123. 错误 memory 会长期污染行为，所以要可见、可撤销、可审计。
124. 用 memory 面板、来源引用、编辑删除、过期策略和写入审批。
125. 长任务要拆阶段、checkpoint、进度汇报和中断恢复。
126. 保存任务状态、上下文摘要、工具结果、文件 diff 和下一步计划。
127. 根据 checkpoint 恢复，重试幂等步骤，非幂等动作需人工确认。
128. 网络、超时可 retry；权限、参数错误、高风险动作不应盲重试。
129. 定义缺失槽位和澄清问题，先问最少必要信息。
130. 触发高风险、低置信度、用户要求或策略失败时 handoff，并附摘要。
131. 放在写操作、高风险工具、外部发送、删除和生产变更之前。
132. 基于用户身份、工具权限、数据权限、风险等级和审计策略。
133. 用 tenant、user identity、delegated auth、最小权限和隔离 trace。
134. 用 trace 记录 planner、tool call、policy decision、state transition。
135. 证据冲突、低置信度、诊断类任务应输出多假设和验证路径。
136. 标注冲突、比较来源可信度、请求更多证据或人工判断。
137. 超时重试、降级、换工具、返回部分结果或人工接管。
138. 根据 evidence threshold、工具空结果和置信度规则判断。
139. 要求引用证据、输出置信度、列假设，并在不足时拒答。
140. 低风险 FAQ 自动，高风险退款/账号转人工，工单结构化。
141. 只读收集指标/日志/部署/runbook，高风险操作审批，trace 变 eval。
142. 构建 diff 上下文，多 reviewer，finding verifier，行号和置信度。
143. lead agent 拆任务，subagent 搜索，synthesis 汇总，citation 校验。
144. 用户可见 memory、审批队列、笔记/邮件/日历工具和隐私边界。
145. 加权限、eval、trace、guardrails、错误处理、成本延迟和人工兜底。

#### 9.14.5 Coding Agent / Agent Harness

146. 先解析目标、约束、验收标准和可能影响范围。
147. 从入口文件、错误栈、测试、README、依赖图和搜索结果选择文件。
148. 用代码搜索、符号索引、调用图、最近变更和测试相关性。
149. 只取任务相关文件和摘要，必要时分阶段读取。
150. 小步修改、遵循局部风格、避免无关重构，并检查 diff。
151. 先看 git status 和文件 diff，遇到用户改动要避让或确认。
152. 运行相关测试，读取失败信息，按栈、断言和变更定位。
153. 环境问题多为依赖/权限/网络，flaky 有随机和历史不稳定，代码问题与新 diff 相关。
154. 修 bug、加功能、改边界逻辑或缺少回归保护时应新增测试。
155. 写失败测试，确认失败，修复，再确认通过；必要时反证。
156. 说明问题、修改点、测试结果、风险和剩余限制。
157. 命令白名单、风险分类、沙箱、超时和审批。
158. 限制工作目录、网络、进程、环境变量、密钥、系统路径和写权限。
159. 默认禁止或审批网络；依赖安装要隔离环境和锁版本。
160. 不读取 `.env`，脱敏日志，密钥路径加入 denylist。
161. 先建计划，分批改，跑测试，避免跨模块一次性大改。
162. 分阶段 checkpoint，生成迁移脚本和回滚方案，持续验证。
163. 高风险、需求不清、测试失败或影响范围扩大时暂停。
164. 记录工具名、输入摘要、输出摘要、状态码、耗时和错误。
165. 用任务通过率、测试通过、人工验收和回归集。
166. 看正确性、最小 diff、可维护性、风格一致性和测试覆盖。
167. 看是否读取关键上下文、解释合理、修改与根因一致。
168. 固定任务集、同 harness、同工具权限，对比 solve rate、成本和失败类型。
169. 固定模型和任务集，对比工具接口、上下文策略、沙箱和验证流程。
170. 上下文不足、误改文件、测试没跑、危险命令、过度重构和虚假完成。

#### 9.14.6 Tool Calling / Function Calling / MCP

171. 让模型连接外部系统，获取实时数据或执行动作。
172. function calling 输出结构化工具名和参数，普通文本不可靠可执行。
173. schema 要明确字段、类型、必填、约束、示例和错误含义。
174. 写给模型理解，也要让人审计；描述要短、准、无重叠。
175. 小而专更容易选对和授权；大而全易误用但集成简单。
176. 模型选择不稳定，eval 难归因，权限边界模糊。
177. 用 tool selection eval：给任务和可选工具，看是否选对。
178. 用 schema 校验、golden 参数、业务规则和工具返回校验。
179. 作为 observation 进入上下文，但要摘要、脱敏和标注可信度。
180. 不一定可信；工具可能失败、过期、被注入或返回脏数据。
181. 脱敏、最小化、按权限过滤，敏感字段不进模型或 trace。
182. 超时、重试、降级、限流提示、fallback 和人工接管。
183. 防重复写操作，例如重复创建工单、退款或发送消息。
184. 先做风险识别、参数展示、人工确认、审计和可回滚。
185. read-only、write-low-risk、write-high-risk、destructive。
186. 读工具重隐私和权限；写工具还要审批、幂等和回滚。
187. 默认禁用，必须显式授权、二次确认和审计。
188. name、description、schema、owner、version、risk、auth、rate limit、audit。
189. 工具发现、鉴权、schema 校验、策略、限流、审批和日志。
190. MCP 标准化工具/资源/上下文接入，让 Agent 更容易复用外部能力。
191. MCP 更偏 Agent 工具协议和发现，内部 API 是业务服务接口。
192. 用 delegated auth、tenant 隔离、最小权限和统一 audit log。
193. 版本化 schema，保持向后兼容，灰度新版本并跑 tool eval。
194. 每次调用作为 tool span 记录输入摘要、输出摘要、耗时和状态。
195. 分析误用 trace，做 tool description A/B 和 selection eval。

#### 9.14.7 Agent Evals

196. Agent 有多步、工具、副作用和状态，最终答案不能代表过程正确。
197. task 是任务，trial 是一次运行，transcript 是轨迹，outcome 是环境结果。
198. id、input、初始状态、允许工具、期望行为、禁止行为、rubric 和 grader。
199. 轨迹能暴露错工具、越权、绕审批、过度调用和偶然正确。
200. 通常不完全通过；应按 final correctness 和 trajectory safety 分开评分。
201. 标记生成失败，保留工具路径评分，避免一刀切。
202. 代码任务、格式校验、权限校验、确定性业务规则。
203. 开放回答、摘要质量、策略遵守和可操作性评分。
204. 高风险、安全、主观质量和 judge 校准样本。
205. 收集已修 bug、线上失败、安全事故和边界条件，固定版本化。
206. 从失败、人工接管、低评分、高成本和异常轨迹中抽样标注。
207. 比较选择工具、调用顺序、参数和次数是否符合 rubric。
208. 高风险动作未审批或禁止动作发生的比例。
209. 需要人工覆盖、纠正或取消的任务比例。
210. 按任务目标是否达成、环境状态是否正确计分。
211. 测多轮上下文、纠错、记忆和状态一致性。
212. 总成本除以成功任务数，比平均请求成本更有业务意义。
213. 统计无效工具调用、重复调用和可由上下文回答的调用。
214. 看必需工具是否调用，尤其是权限、检索、验证类工具。
215. 检查高风险动作前是否有 approval span 和用户确认。
216. case、rubric、grader、数据和期望输出都要版本化。
217. 固定其他变量，逐一替换做 ablation，并记录版本。
218. 新版本先跑离线 eval，再小流量 canary，触发门禁则回滚。
219. 使用隐藏集、线上新样本、人工抽检和多维指标。
220. 包含数据集、版本、指标、失败分类、示例 trace 和改进计划。

#### 9.14.8 Observability / Tracing

221. Agent 行为多步且不确定，tracing 才能定位每一步发生了什么。
222. log 是事件记录，trace 是带父子关系的端到端执行链路。
223. trace 是一次请求全链路，span 是其中一个步骤。
224. generation、retrieval、tool、guardrail、handoff、planner、custom span。
225. 模型、prompt 摘要、输入输出摘要、token、耗时、错误和版本。
226. 工具名、版本、参数摘要、结果摘要、耗时、状态和风险等级。
227. query、source、top-k、score、filter、rerank 和证据 id。
228. guardrail 类型、输入摘要、决策、tripwire、误杀/漏放标记。
229. handoff 对象、原因、上下文摘要和人工处理结果。
230. 业务状态、审批、checkpoint、预算、缓存和自定义决策。
231. 看 retrieval span 的 query、filter、top-k、score 和空结果。
232. 看 tool span 的参数、状态码、错误、超时和重试。
233. 看 generation span 的上下文、输出、引用和格式错误。
234. 对比被拦输入、policy、人工判断和历史类似样本。
235. 看 handoff 触发条件、传递摘要和人工接手结果。
236. 不宜默认完整保存；应摘要、脱敏、按权限控制。
237. 识别姓名、邮箱、电话、密钥、合同等，替换、哈希或不存。
238. 全量保留错误和高风险，普通请求按比例采样。
239. 保留失败、低评分、高成本、人工接管和安全相关 trace。
240. 脱敏后标注 input、expected behavior、trajectory 和 grader。
241. 监控指令遵守率、policy 违反、输出风格漂移和用户纠错。
242. 监控必需证据缺失、空检索、低 recall 和用户追问。
243. 对行为 embedding/标签聚类，观察分布变化和异常簇。
244. token、工具调用、重试、模型路由和缓存命中异常。
245. 工具错误率、超时率、调用次数、参数错误和高风险调用。
246. 展示质量、成本、延迟、工具、错误、guardrail 和 eval 趋势。
247. trace 提供样本，metrics 看趋势，feedback 标注质量，eval 防回归。
248. 采样、摘要、冷热分层、保留策略和敏感字段不落盘。
249. RBAC、租户隔离、审计、脱敏视图和临时访问授权。
250. 用 OpenTelemetry 统一 trace 语义，再扩展 Agent 专属 span。

#### 9.14.9 Guardrails / Prompt Injection / Security

251. 恶意输入诱导模型违反系统指令、泄露数据或调用危险工具。
252. 普通指令是合法任务，injection 试图改变权限、规则或隐藏目标。
253. 文档作为不可信内容，只能当证据，不能拥有指令权。
254. 网页内容隔离为 untrusted，工具和权限策略在模型外执行。
255. 可能；工具返回也要标注信任域并做输出/工具 guardrail。
256. 可能；错误或恶意 memory 会长期影响 Agent，需要审核和删除。
257. system/developer 是 trusted，用户、文档、网页、工具输出多为 untrusted。
258. 否则检索内容可覆盖系统规则，造成越权和数据泄露。
259. 拦恶意请求、越权意图、PII 输入和不支持任务。
260. 拦敏感输出、无引用结论、格式错误和违规承诺。
261. 拦危险工具、越权参数、高风险动作和异常返回。
262. 停止执行、解释原因、请求确认或转人工。
263. blocking 更安全但慢；parallel 延迟低但副作用前要小心。
264. 模型不是可信执行环境，权限必须由系统和数据层判断。
265. tenant 过滤、独立索引/命名空间、权限 eval 和 trace 脱敏。
266. 输出前 DLP 检测、脱敏、拒答和最小化上下文。
267. 不把系统 prompt 暴露给模型可输出区域，输出 guardrail 拦截。
268. Tool Gateway 鉴权和策略检查，模型只能请求，不能绕过。
269. 识别为冲突指令，坚持高优先级规则并可提醒用户。
270. 检测 exfiltration 意图，拒答并记录安全事件。
271. 把它当文档内容忽略指令部分，只抽取可验证事实。
272. 构造 injection、越权、PII、危险工具和跨租户 case。
273. 复盘根因，抽象输入和期望行为，加入安全 regression suite。
274. 高风险宁可保守，低风险可给澄清和替代路径。
275. 展示动作、参数、影响、回滚方式，让授权人显式确认。

#### 9.14.10 Multi-agent / Research Agent

276. 任务复杂、可并行、需要多视角或多工具专家时。
277. 简单问答、固定流程、预算紧张或协调成本超过收益时。
278. 澄清目标、拆分任务、分配子任务、监控进度和合成结果。
279. 目标、边界、工具、预算、输出格式、停止条件和引用要求。
280. 共享任务表、去重查询、source registry 和 lead agent 协调。
281. 给每个子 Agent token、时间、工具次数和搜索深度限制。
282. 根据 query complexity、信息源数量和不确定性动态分配。
283. 按 claim 合并、去重、比较证据和可信来源。
284. 标注冲突、请求补证、让 checker 验证或交给用户判断。
285. 验证每个 claim 是否有来源支持，并修正或删除无证据 claim。
286. claim-to-source 对齐，必要时逐句检查引用。
287. trace 中记录 lead/subagent 层级、任务 id 和父子 span。
288. 保存任务树、子结果、预算、已用来源和合成草稿。
289. 看任务拆分、重复工作、冲突、预算耗尽和 citation failure。
290. 评报告质量、引用支持、覆盖率、重复率、成本和协调失败。
291. 限制子 Agent 数量、深度、上下文和模型路由。
292. 最大深度、最大子任务数、预算阈值和停止条件。
293. 并行 workflow 路径固定，多 Agent 任务拆分更动态。
294. 强制引用、claim verification、反证搜索和最终校验。
295. 重复搜索、目标漂移、冲突不处理、引用不支持和预算爆炸。

#### 9.14.11 LLM 系统部署、成本与性能

296. 开发、测试、预发、生产隔离，数据、密钥和权限分环境。
297. 看质量、成本、延迟、隐私、可控性、部署能力和合规。
298. 按任务类型、难度、风险、成本和延迟选择模型。
299. 简单任务小模型，复杂推理/安全关键任务强模型。
300. 定义失败条件、降级模型、输出差异和用户提示。
301. 超时重试、降级、排队、返回部分结果或转人工。
302. 指数退避、队列、限流、缓存和供应商 fallback。
303. 精简上下文、缓存、模型路由、压缩、批处理和减少重试。
304. 并行化、缓存、流式输出、轻量模型、减少工具链和优化检索。
305. 提升感知速度，但需要处理中断、部分输出和前端协议。
306. prompt、retrieval、rerank、tool result、answer 和 embedding cache。
307. 系统 prompt 长且重复、支持稳定前缀时适合。
308. 重复 query、热门文档、低时效数据适合。
309. 可能权限错配、过期、个性化错误和引用不一致。
310. 离线评估、批量摘要、embedding 和非实时任务。
311. 用 QPS、token/s、p95、工具耗时、并发和供应商配额估算。
312. 按用户、租户、工具、模型和成本做多维限流。
313. 给租户/用户预算，超限降级、排队或审批。
314. 每个 trace 记录 tenant、model、token、tool 和成本。
315. 线上抽样 eval、用户反馈、任务成功率和关键指标漂移。
316. 离线 eval、canary、小流量、观察窗口和自动回滚。
317. 保留旧模型配置、prompt、tool schema 和 eval 结果，一键切回。
318. 分流、随机化、指标定义、显著性和安全门禁。
319. 多供应商 fallback、降级模式、熔断和用户提示。
320. 定义可用性、延迟、任务成功率、安全违规率和成本边界。

#### 9.14.12 作品集与项目经验

321. 补架构、权限、eval、trace、失败复盘和量化指标。
322. 展示意图分类、工单字段、高风险审批、人工交接和 SLA 边界。
323. 展示 harness、sandbox、diff、测试、trace 和 PR workflow。
324. Problem、Architecture、Design Decisions、Evals、Trace、Failure Modes、Runbook。
325. Gateway、runtime、tools、data source、guardrails、evals、observability。
326. 展示 case schema、输入、期望行为、禁止行为和 grader。
327. 展示数据集、指标、回归、失败样本和下一步。
328. 展示关键 span、工具调用、guardrail、输出和人工决策。
329. 讲现象、影响、trace、根因、修复和防复发。
330. 明确哪些是 demo，哪些生产能力已设计/实现，哪些仍是后续。
331. 用 trace 定位问题，用工程改动修复，用 regression eval 防复发。
332. 说自己负责的模块、决策、权衡、指标和具体产出。
333. 用成功率、准确率、召回、MTTR、采纳率、成本和延迟。
334. 用历史数据回放、合成 case、人工标注和离线 eval。
335. 小而完整：有架构、测试、eval、trace、README 和复盘。
336. 回答：固定流程优先，Agent 用在路径开放、依赖工具反馈的部分。
337. 回答：知识更新和权限用 RAG，fine-tuning 适合风格/能力。
338. 回答：长上下文贵且权限/引用/更新差，RAG 更可控。
339. 回答：通过离线 eval、线上指标、A/B 和失败样本回归证明。
340. 回答：权限、guardrails、审批、sandbox、trace、回滚和 eval。

#### 9.14.13 反向面试问题

341. 看团队是否有明确业务指标，而不是只追 demo。
342. 判断生产化程度；没有 trace/eval，后续排障会困难。
343. 看团队是否能分层 debug，而不是把问题都归因于模型。
344. 判断工程成熟度和发布安全性。
345. 看安全意识和对不可信内容的隔离设计。
346. 看工具系统是否有生产边界和审计。
347. 看 RAG 权限是否在生成前完成。
348. 看是否有反馈到 eval、数据和产品改进的闭环。
349. 看是否有发布门禁、canary、回滚和安全评估。
350. 理想答案是两者平衡：demo 速度用于探索，生产可靠性用于交付。

---

## 10. 第二遍进阶专题题单：带参考答案

> 这一节用于第二遍复习。每题给出可直接口头回答的参考答案，并尽量映射到 2026 年仍然高频的工程面试信号。来源优先参考 OpenAI、Anthropic、MCP、OWASP、OpenTelemetry、Ragas、LangSmith 等官方文档或工程文章。

### 10.1 结构化输出、模型路由与成本

351. **Structured Outputs 和 JSON mode 的核心区别是什么？**  
     答：JSON mode 主要保证输出是合法 JSON；Structured Outputs 进一步要求输出符合给定 JSON Schema。面试里要强调它解决的是接口解析和 schema adherence，不等于保证事实正确、业务正确或安全合规。

352. **什么时候用 function calling，什么时候用 structured response format？**  
     答：如果模型要连接系统能力、数据库、工具或外部动作，用 function calling；如果只是希望最终回复按固定结构返回给应用层或 UI，用 structured response format。前者是“模型请求系统做事”，后者是“模型按结构回答”。

353. **Structured Outputs 能否替代后端校验？**  
     答：不能。它能降低格式错误，但业务约束、权限、幂等、风险等级、库存状态、金额上限等仍必须由后端系统校验。生产系统应把 schema 视为输入边界的一层，不是可信执行环境。

354. **多模型供应商下，结构化输出有什么兼容风险？**  
     答：不同供应商对 JSON Schema 子集、并行工具调用、拒答格式和错误处理的支持不完全一致。工程上要做 provider adapter、schema 子集约束、契约测试和回归 eval，避免在切模型时只测“能解析”而不测“字段语义正确”。

355. **Prompt caching 如何优化成本和延迟？**  
     答：把稳定前缀放在 prompt 开头，例如 system 指令、工具说明、少量示例和固定 policy；把用户输入、检索证据、临时工具结果放在后面。缓存依赖前缀匹配，动态内容越靠前，缓存命中越差。

356. **模型路由应该按什么维度设计？**  
     答：按任务难度、风险等级、上下文长度、结构化输出要求、延迟预算、成本预算和历史成功率路由。简单分类和格式转换走小模型，高风险决策、复杂推理和长上下文 synthesis 走强模型，并保留 fallback 和回滚。

357. **Fallback model 设计的关键难点是什么？**  
     答：难点不是“换一个模型再试”，而是保证输出契约、工具调用能力、安全策略和用户体验一致。fallback 前要定义触发条件，fallback 后要标记 trace，并评估质量差异、成本差异和是否需要人工接管。

358. **Streaming 会改变 LLM 系统设计的哪些部分？**  
     答：Streaming 改善首 token 体验，但要求前端能处理增量输出、取消、重试、部分结果和最终校验。对需要结构化输出或工具调用的任务，不能只看流式文本，还要等最终对象、guardrail 和后端状态确认。

359. **LLM 应用中的缓存分为哪些层？**  
     答：常见层包括 prompt cache、embedding cache、retrieval cache、rerank cache、tool result cache 和 answer cache。越靠近最终答案，越要关注权限、时效、个性化和引用一致性；生产系统通常优先缓存稳定前缀和检索中间结果。

360. **为什么 cost per successful task 比 average request cost 更有意义？**  
     答：Agent 任务可能多轮、多工具、多次重试，只看单次请求成本会低估失败和重试成本。`cost / successful_task` 能把质量、成本和完成率放在一起衡量，更接近业务实际付费意愿。

### 10.2 MCP、Tool Registry 与工具安全

361. **MCP 中 host、client、server 分别是什么？**  
     答：host 是发起连接的 LLM 应用，例如 IDE 或聊天产品；client 是 host 内部连接某个 MCP server 的连接器；server 提供 resources、prompts、tools 等能力。这个拆分让工具和上下文接入标准化，也让权限边界更清楚。

362. **MCP 的 resources、prompts、tools 有什么区别？**  
     答：resources 是给模型或用户使用的上下文和数据；prompts 是可复用的模板化消息或工作流；tools 是模型可请求执行的函数能力。面试中要说明：resources 主要供给信息，tools 可能产生动作，风险等级不同。

363. **MCP 为什么需要 capability negotiation？**  
     答：不同 server 支持的能力不同，例如是否支持 tools、resources subscribe、prompts listChanged。初始化阶段声明 capability 后，host 才能决定展示什么、调用什么、监听什么，避免客户端假设能力存在而导致运行时错误。

364. **MCP resource subscription 适合什么场景？**  
     答：适合文件、文档、配置、任务状态等会变化的上下文。订阅后资源更新可通知 client，Agent 能避免使用过期上下文；但订阅内容仍要做权限控制、脱敏和范围限制。

365. **MCP tool definition 中最关键的字段是什么？**  
     答：至少要有唯一 `name`、清晰 `description`、`inputSchema`，有结构化返回时还应有 `outputSchema`。工具 annotations 和描述可能来自 server，除非 server 可信，否则 host 不能把它们当安全事实。

366. **MCP 工具调用为什么必须要求用户同意和控制？**  
     答：工具可能访问数据或执行代码路径。安全原则要求用户理解数据会被谁访问、动作会造成什么影响，并能授权或拒绝。模型只能提出调用请求，不能绕过 host 的授权和审计。

367. **MCP sampling 有什么特殊风险？**  
     答：sampling 允许 server 触发模型调用，可能造成递归调用、数据外传、成本失控或提示注入扩大化。设计上应让用户控制是否允许 sampling、实际发送的 prompt 以及 server 能看到哪些结果。

368. **工具列表动态变化时，Agent 系统要注意什么？**  
     答：需要处理 `listChanged` 通知、工具版本、schema 变化和 eval 回归。工具新增或重命名可能改变模型选择行为，所以要把 tool registry 变化纳入发布流程，而不是把工具列表当静态 prompt。

369. **MCP server 市场化后，主要安全风险有哪些？**  
     答：风险包括 lookalike tool、过宽权限、恶意工具描述、数据外传、供应链污染和 tool combination attack。host 应做来源信任、权限最小化、工具风险分级、用户确认、审计和安全 eval。

370. **MCP 和普通内部 API 的关系是什么？**  
     答：内部 API 是业务系统接口，MCP 是面向 LLM 应用暴露资源、工具和 prompt 的协议层。生产中通常用 MCP server 包装内部 API，但鉴权、审计、限流、幂等和业务校验仍由企业系统负责。

### 10.3 Agent Runtime、持久化与人工介入

371. **为什么长任务 Agent 需要 durable execution？**  
     答：长任务可能跨分钟到小时，期间会遇到模型超时、工具失败、服务重启和人工审批。durable execution 把状态保存到持久层，使任务可暂停、恢复、重放和审计。

372. **Agent checkpoint 应该保存哪些信息？**  
     答：保存目标、当前阶段、状态变量、工具结果摘要、预算、已完成动作、待审批动作、关键上下文引用、artifact 路径和下一步计划。不要只保存聊天历史，否则恢复后很难判断真实执行状态。

373. **Human-in-the-loop 应放在哪些位置？**  
     答：应放在高风险写操作、外部发送、删除、生产变更、权限升级、低置信度决策和用户明确要求确认的位置。好的设计不是最后统一点“批准”，而是在风险发生前暂停并展示影响、参数和回滚方式。

374. **可恢复 Agent 如何处理有副作用的工具调用？**  
     答：副作用工具必须有 idempotency key、状态查询、去重和审计。恢复或重放时不能盲目再次执行退款、发送邮件或改配置，而应先查上一次动作是否已经成功。

375. **长运行 worker 中 trace 为什么可能需要显式 flush？**  
     答：trace 通常异步批量导出。后台任务、队列 worker 或 serverless 任务结束时，如果进程很快退出，trace 可能还在缓冲区；显式 flush 可提高导出完整性，便于事后排障。

376. **state、memory 和 artifact 的区别是什么？**  
     答：state 是当前任务运行状态；memory 是跨任务、跨会话保留的偏好或事实；artifact 是文件、报告、代码、图表等可独立引用的产物。三者混在聊天上下文里会造成恢复困难和信息丢失。

377. **为什么子 Agent 输出有时应该写入 artifact，而不是只回传给主 Agent？**  
     答：大结果通过主 Agent 口头转述会丢信息、耗 token、引入二次总结误差。让子 Agent 直接写文件、表格或结构化结果，再把引用交给主 Agent，可以降低上下文压力并提高可审查性。

378. **长上下文快满时，Agent 应如何保持连续性？**  
     答：应阶段性压缩已完成工作，把关键事实、决策、待办、证据引用和 artifact 路径写入外部状态；必要时启动新上下文继续执行。不要简单截断历史，否则容易丢掉约束和已做动作。

379. **time travel debugging 对 Agent 有什么价值？**  
     答：它允许从历史 checkpoint 查看、回放或分叉执行，定位哪个决策导致失败。对多轮 Agent 来说，这比只看最终失败输出更有价值，因为很多错误来自早期工具选择或错误状态。

380. **异步 Agent job 应如何向用户呈现进度？**  
     答：前端展示阶段、当前动作、等待原因、可取消入口、已产出 artifact 和预计下一步。后端通过 checkpoint、trace 和事件流驱动 UI，避免用户只能看到一个长时间 spinning 状态。

### 10.4 进阶 Evals 与统计解释

381. **`pass@k` 和 `pass^k` 分别衡量什么？**  
     答：`pass@k` 衡量 k 次尝试中至少一次成功的概率，适合“多试几次有一个能用”的场景；`pass^k` 衡量 k 次全部成功的概率，适合用户每次都期望稳定成功的生产 Agent。

382. **为什么 Agent eval 要重复运行同一个 case？**  
     答：Agent 有采样、工具、检索和环境非确定性，单次通过不代表稳定。重复运行能估计成功率、方差和不稳定失败模式，尤其适合客服、浏览器和研究型 Agent。

383. **offline eval 和 online eval 如何分工？**  
     答：offline eval 用 curated dataset 在发布前比较版本、防回归；online eval 在生产 trace 上抽样监控真实分布、安全和质量漂移。成熟流程是线上失败进入数据集，离线验证修复，再灰度上线。

384. **如何从生产 trace 生成 eval case？**  
     答：先筛失败、低评分、高成本、人工接管和安全事件 trace，脱敏后标注用户目标、初始状态、允许工具、禁止行为、期望 outcome 和评分器。关键是保留轨迹和环境状态，而不只是输入输出。

385. **不同风险等级应选择什么 grader？**  
     答：确定性规则用 code-based grader；开放文本和交互质量用 LLM rubric；高风险、安全、合规和 judge 校准样本用 human grader。越高风险，越不能只依赖模型打分。

386. **如何解释 eval 分数的统计不确定性？**  
     答：看样本量、置信区间、重复 trial、case 难度分布和分层指标。小样本上 2% 的提升可能只是噪声；生产门禁应关注关键场景和高严重度失败，而不是只看平均分。

387. **LLM-as-a-judge 如何校准？**  
     答：用人工标注 golden set 对齐 rubric，定期抽检 judge 输出，比较一致性和偏差；对关键 case 使用双 judge 或 human review。还要固定 judge 模型和 prompt 版本，否则评分基准会漂移。

388. **多轮客服 Agent 为什么需要 end-state eval？**  
     答：同一个目标可能有多条合理路径，逐步匹配固定轨迹会误杀。更好的做法是检查最终 ticket、refund、confirmation、用户状态等是否正确，同时用 transcript rubric 约束语气、轮数和策略遵守。

389. **研究型 Agent 的 eval 难点是什么？**  
     答：研究输出开放、来源会变化、专家可能不同意“完整性”标准。通常需要 groundedness、coverage、source quality、citation support 和人工校准的 LLM rubric 组合评估。

390. **为什么 evaluator 也要版本化？**  
     答：prompt、rubric、规则代码、judge 模型或阈值改变都会改变分数含义。版本化 evaluator 能解释历史趋势，避免把评分器变化误判成模型或 Agent 质量变化。

### 10.5 OWASP、Prompt Injection 与生产安全

391. **OWASP LLM Top 10 对 Agent 面试有什么价值？**  
     答：它提供了 LLM 应用风险分类语言，例如 prompt injection、sensitive information disclosure、supply chain、excessive agency、overreliance 等。面试里可以用它组织威胁建模，而不是零散说“加 guardrail”。

392. **Excessive Agency 是什么？**  
     答：系统给模型过多自主动作能力、过宽权限或缺少确认，导致模型一旦误判就能造成真实影响。缓解方式包括最小权限、工具风险分级、审批、限额、幂等和可回滚设计。

393. **Insecure Output Handling 为什么危险？**  
     答：如果把模型输出直接当代码、SQL、HTML、shell 或业务指令执行，模型错误或被注入的输出会传染到下游系统。必须做解析、转义、schema 校验、权限校验和安全执行环境。

394. **Sensitive Information Disclosure 如何在 LLM 系统里发生？**  
     答：敏感信息可能来自 prompt、RAG context、tool result、memory、trace 或最终输出。防护要覆盖数据最小化、权限过滤、DLP、脱敏、trace 访问控制和输出 guardrail。

395. **Model DoS / cost attack 在 Agent 中如何表现？**  
     答：攻击者可诱导超长上下文、无限循环、多工具重试、高成本模型路由或大量子 Agent。系统要有限步数、预算、速率限制、上下文大小、工具次数和异常成本告警。

396. **LLM supply chain 风险包括哪些？**  
     答：包括第三方模型、embedding 模型、MCP server、插件、prompt 包、数据集、向量库内容和依赖库被污染。缓解方式是来源审查、版本锁定、最小权限、签名、隔离和上线前安全 eval。

397. **为什么检索文档和网页内容必须视为 untrusted content？**  
     答：它们可能包含恶意指令、过期信息或攻击者控制的内容。模型可以把它们当证据，但系统不能让它们覆盖 system/developer 指令，也不能让它们直接决定权限和工具调用。

398. **RAG 能否彻底解决 prompt injection？**  
     答：不能。RAG 反而引入了间接 prompt injection：恶意内容藏在文档或网页中。正确做法是信任域隔离、引用约束、工具前校验、敏感动作审批和安全 regression eval。

399. **安全 eval 应覆盖哪些 case？**  
     答：覆盖直接 injection、间接 injection、跨租户读取、PII 输出、越权工具、高风险写操作、成本攻击、系统 prompt 泄露和恶意工具返回。每次事故都应沉淀为回归 case。

400. **LLM 安全事故复盘应产出什么？**  
     答：产出现象、影响范围、trace、根因层级、修复项、数据/工具/prompt/policy 变更、回滚动作和 regression eval。不要只写“优化 prompt”，否则无法防止同类问题再次发生。

### 10.6 Observability 标准化与 Trace 设计

401. **OpenTelemetry GenAI semantic conventions 解决什么问题？**  
     答：它给模型调用、检索、工具调用、事件、指标和 Agent span 提供统一命名和属性约定，方便不同框架和平台之间交换 telemetry。面试里可把它作为“不要自创不可迁移日志格式”的依据。

402. **模型 inference span 应记录哪些核心字段？**  
     答：记录 operation、provider、request model、response model、token usage、latency、error type 和必要的采样信息。prompt 和输出可做摘要或 opt-in 保存，因为它们可能包含敏感数据。

403. **retrieval span 应记录哪些内容？**  
     答：记录 query 摘要、data source、top-k、filter、retrieved doc id、score、rerank 信息和错误。完整 query 和文档内容可能敏感，应默认摘要、脱敏或受控保存。

404. **tool span 为什么要特别注意敏感字段？**  
     答：tool arguments 和 results 往往包含客户数据、订单、合同、密钥或内部系统返回。trace 要记录可排障的摘要、状态、耗时和风险等级，但敏感参数要脱敏、哈希或按权限隔离。

405. **Agent trace 采样策略如何设计？**  
     答：普通成功请求可低比例采样；失败、高成本、高延迟、人工接管、安全事件和高风险工具调用应全量保留。采样决策最好基于 span 初始属性和最终 outcome 组合。

406. **为什么不应默认保存完整 prompt？**  
     答：完整 prompt 可能含用户隐私、企业知识、检索证据、工具结果和系统策略。默认保存摘要和结构化 metadata，需要排障时再受控开启，并配合保留周期和访问审计。

407. **如何监控 Agent 行为漂移？**  
     答：监控工具调用分布、拒答率、澄清率、任务成功率、policy violation、用户纠错、行为聚类和 eval 分数。行为漂移不是传统错误率能完全覆盖的，需要 trace 和反馈结合。

408. **为什么 trace 要关联 prompt、model、tool schema 和 dataset 版本？**  
     答：Agent 失败可能来自任何一个版本变化。没有版本关联，就无法做 ablation，也无法解释某天质量下降是模型、prompt、工具、检索数据还是 harness 改动造成的。

409. **trace 如何进入 eval 闭环？**  
     答：线上 trace 先用于归因和分类，再脱敏标注为 eval case，最后进入 regression suite。修复后用相同 case 验证，避免线上同类失败重复出现。

410. **Agent observability dashboard 应展示哪些 SLO？**  
     答：展示任务成功率、关键场景成功率、安全违规率、p95/p99 延迟、成本、工具错误率、检索缺失、人工接管率、用户反馈和 eval 趋势。只看 token 和 latency 不足以描述 Agent 质量。

### 10.7 多模态、语音与 Computer-use Agent

411. **语音 Agent 的 trace 和文本 Agent 有什么不同？**  
     答：除 generation、tool、guardrail 外，还要记录 transcription、speech、speech group、音频延迟、打断和识别错误。音频数据通常更敏感，默认不应完整保存。

412. **语音 Agent 的隐私风险有哪些？**  
     答：音频可能包含声纹、背景对话、身份信息和未预期内容。设计上要有录音提示、数据最小化、音频脱敏或不落盘、保留周期、访问控制和用户删除能力。

413. **Computer-use Agent 如何评估是否完成任务？**  
     答：不要只看它说“完成了”，要检查环境最终状态，例如 URL、页面状态、文件系统、数据库、应用配置或订单状态。GUI 路径可以不同，但最终可验证状态必须正确。

414. **浏览器 Agent 中 DOM 工具和截图工具如何取舍？**  
     答：DOM 适合提取大量文本和结构化页面，速度快但 token 可能很大；截图适合视觉布局、商品浏览和非结构化界面，可能更慢但更贴近人类界面。生产 Agent 通常按任务动态选择。

415. **图像输入也会有 prompt injection 吗？**  
     答：会。截图、图片或 OCR 文本可能包含恶意指令。系统应把视觉内容转成 untrusted observation，只提取事实，不允许它覆盖高优先级指令或触发未授权工具。

416. **多模态 RAG 的 ingestion pipeline 有哪些额外步骤？**  
     答：需要版面解析、OCR、表格结构化、图片 caption、图表数据抽取、坐标或页码锚点、跨模态 embedding 和引用定位。评估时要分别检查解析质量、检索质量和引用可验证性。

417. **实时语音 Agent 为什么需要 interruption handling？**  
     答：用户会打断、纠正或改变意图。如果系统不能取消正在生成的语音、停止工具调用或重建状态，就会出现响应滞后和误执行。需要把打断作为一等事件记录到 state 和 trace。

418. **语音 Agent 的关键延迟指标有哪些？**  
     答：包括 speech-to-text 延迟、首 token 延迟、首音频延迟、工具等待时间、端到端响应延迟和打断响应时间。用户感知通常比纯文本更敏感，所以要拆阶段优化。

419. **浏览器 Agent 如何评估工具选择是否正确？**  
     答：构造任务集，标注每一步应该用 DOM、截图、点击、输入、滚动还是搜索工具，并比较实际轨迹。工具选错可能不立刻失败，但会增加 token、延迟和误操作风险。

420. **Computer-use Agent 的 sandbox 应限制什么？**  
     答：限制文件系统、剪贴板、网络域名、下载上传、凭据、系统设置、支付和外部发送。高风险动作前要截图或状态摘要给用户确认，并记录可审计 trace。

### 10.8 作品集、系统设计表达与反追问

421. **面试中介绍 Agent 系统，最稳定的四层结构是什么？**  
     答：先讲用户目标和成功指标，再讲 runtime / tools / data 的架构，然后讲 eval / observability / guardrails，最后讲失败模式和 trade-off。这样能从 demo 叙述升级到生产系统叙述。

422. **面试官问“为什么不用普通 workflow”，怎么回答？**  
     答：先承认固定 workflow 更可控，适合规则明确路径；再说明当前任务是否存在开放步骤、动态工具选择、多轮反馈和未知分解。如果这些条件不足，就应选择 workflow 而不是 Agent。

423. **把 demo Agent 迁到生产，第一阶段应该补什么？**  
     答：补权限、日志/trace、失败处理、工具幂等、eval 集、发布门禁、成本限制和人工接管。不要先追复杂多 Agent，而要先让单 Agent 可测、可控、可恢复。

424. **企业验收 Agent 项目时应看哪些证据？**  
     答：看关键任务成功率、安全违规率、人工接管率、p95 延迟、成本、trace 样本、eval report、权限测试、事故演练和回滚方案。只展示 demo 视频不足以证明可上线。

425. **如何讲一个 Agent 项目的失败复盘？**  
     答：按现象、影响、trace 证据、根因、修复、回归测试和剩余风险来讲。好的复盘要能说明你如何从“模型答错了”定位到具体系统层，而不是泛泛说 prompt 不好。

426. **如何量化 Agent 的业务价值？**  
     答：用任务完成率、平均处理时长、人工节省、用户满意度、转人工率、错误成本、每成功任务成本和 SLA 改善衡量。技术指标要能映射到业务结果，否则很难说服面试官。

427. **没有线上数据时，如何做可信 eval？**  
     答：用历史样例、公开 benchmark、专家合成 case、对抗 case 和小规模人工标注建立初始集；上线后再用真实 trace 迭代。要诚实说明数据来源和覆盖盲区。

428. **面试官问“为什么用这个 Agent 框架”，怎么回答？**  
     答：从需求出发回答：是否需要 durable execution、graph state、human-in-the-loop、多 Agent 编排、observability 集成或生态工具。若只是简单工具循环，直接用 SDK 或少量代码可能更合适。

429. **开源本地模型和闭源 API 模型如何取舍？**  
     答：闭源 API 通常质量、工具能力和维护成本更优；开源本地模型在数据控制、私有化、可定制和单位成本上有优势。生产取舍要看质量门槛、合规、延迟、吞吐、运维能力和供应商风险。

430. **如果只准备一个作品集项目，应该覆盖哪些能力信号？**  
     答：选择一个小而完整的 RAG 或 Agent 项目，必须展示架构图、工具/数据边界、eval dataset、trace、guardrails、失败复盘、成本延迟和 README 讲法。面试官看的是生产意识，不只是功能能跑。
