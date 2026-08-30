# 金融 AI Agent 市场调研

> 调研日期：2026-08-03
> 范围：全球及中国市场中，面向机构投研、投行、财富管理与零售投资者的金融 AI Agent 产品。本文为产品与市场研究，不构成投资建议。

## 执行摘要

金融 AI 正从“聊天式投研搜索”进入“可交付工作成果的工作流自动化”阶段。领先产品已经能够跨数据源检索、拆解多步任务、生成带引用的研究报告、Excel 模型、PPT 或客户服务材料，并在关键节点保留人工复核。

市场的核心竞争并非基础模型本身，而是四项能力：

1. 合法授权且持续更新的金融数据；
2. 可追溯的证据链、权限与审计；
3. 针对投行、PE、私募信贷、研究员及财富顾问的工作流模板；
4. 对 Excel、PowerPoint、CRM、文档库和行情系统的嵌入式集成。

真正端到端自主交易或提供个性化投资建议的公开产品仍然很少。监管责任、适当性义务、利益冲突与可解释性，使“人机协同的工作流 Agent”成为目前更可行的商业化形态。

## 市场定义与成熟度

本文将金融 AI 产品分为四层，避免把所有生成式 AI 功能都称为 Agent：

```text
信息检索/摘要 ──── 研究与材料生成 ──── 顾问工作流 ──── 投资建议/交易执行
  已普及              正在规模化             较成熟              高度受限
```

- **Copilot**：回答问题、摘要财报、检索数据、起草邮件；用户主导每一步。
- **工作流 Agent**：能规划多个步骤、调用数据与工具、输出模型或材料，并支持监控和反复运行。
- **顾问 Agent**：服务持牌顾问，例如会前准备、会议纪要、CRM 更新和客户沟通初稿。
- **建议/交易 Agent**：直接给出个性化建议、下达或执行交易。该层受监管最严格，公开产品通常将自身限制为信息或教育用途。

## 产品地图

| 赛道 | 产品/玩家 | 主要能力 | 现阶段判断 |
| --- | --- | --- | --- |
| 金融数据平台 Agent 化 | FactSet AI for Banking | 多 Agent、触发式生成 pitch、公司画像、备忘录、深度研究和买卖方分析；强调可追溯性 | 机构投行场景，已发布 Alpha |
|  | LSEG Workspace AI Search | 用自然语言跨金融数据和文档提问，提供可验证且带引用的答案 | 研究检索型 Agent，持续扩展 |
|  | S&P Capital IQ Pro / ChatIQ | 多文档分析、自然语言筛选、信用与市场数据、文档智能和审计引用 | 强 Copilot，正在向 Agent 工作流演进 |
| 垂直投研与投行 Agent | Rogo（Felix + Agents） | 从单一提示输出 Excel、PPT、Word、仪表盘和有来源研究；支持企业定制 Agent 与连接器 | 金融分析师工作台路线最清晰 |
|  | Hebbia Matrix | 对大量非结构化文档、表格和图表执行多步推理；结果逐步引用、可审计 | 尽调、私募信贷、对冲基金研究优势明显 |
|  | Finster AI | 与 FactSet 联合面向投行，自动化交易全周期中的材料与研究工作 | 创业公司借大型数据平台分发的代表 |
| 财富管理/投顾赋能 | Morgan Stanley AI @ MS | 内部知识检索、客户会议纪要、行动项、邮件初稿与 CRM 更新 | “顾问增强”成熟案例 |
|  | TIFIN.AI + FactSet | 顾问生产力、客户洞察、个性化服务与财富管理工作流 | 数据商与财富科技结合 |
| 零售投资助手 | Robinhood Cortex | 行情异动解释、投资组合摘要、实时洞察和扫描器 | 明确定位为信息服务，而非投资建议 |
| 中国券商智能体 | 国泰海通、广发、国信、中金、华泰等 | 研报、行情、标的筛选、ETF/基金比较、投顾和运营能力封装为 Skills/MCP | 2026 年进入集中上线期，偏受控工具调用 |

## 重点产品观察

### 1. 数据巨头：从终端问答转向可审计的行业工作流

FactSet 于 2026 年 3 月发布 FactSet AI for Banking Alpha，定位为面向投行、卖方研究等场景的工作流自动化生态。其公开描述包含多个触发式 Agent，能够产出 pitch 材料、公司画像、备忘录、深度研究及买卖方分析，并强调数据与任务的可追溯性。

LSEG 的 Workspace AI Search 将其金融数据、内容和业务逻辑聚合在会话式入口，主张对复杂问题提供可验证的带引用答案。S&P Capital IQ Pro 则在既有数据工作台中提供 ChatIQ、Document Intelligence 和自然语言筛选等能力。

这类厂商的护城河是授权数据、实体与证券主数据、实时性、已有终端席位、审计能力与客户工作流，而非单一模型能力。其典型商业模式是终端或数据订阅的 AI 增购、企业 API 和私有化/集成服务。

### 2. 垂直创业公司：交付“初级分析师的成果”

Rogo 的 Felix 主张将一个提示转化为可交付的模型、材料、研究和仪表盘；其 Agent Library 覆盖投行、PE、商业银行、私募信贷、对冲基金、股票研究和资产管理等角色。Rogo 的产品方向表明：金融 Agent 的价值不在于一次问答，而在于将企业模板、研究方法、格式规范和机构记忆编码为可重复运行的工作流。

Hebbia Matrix 的核心场景是长文档和跨文档尽调。它支持将内部文档、模型、电话会记录、CRM 数据与外部数据放入同一工作区，以结构化、带来源的方式返回结果。其在私募信贷场景中强调对 covenant、baskets、carve-outs 等条款进行跨文档对比。

此类创业公司的优势是产品迭代速度、垂直交互与“最后一公里”交付；风险是金融数据授权成本、企业安全审查和对大型数据平台渠道的依赖。

### 3. 财富管理：先增强顾问，再讨论替代顾问

Morgan Stanley 的 AI @ Morgan Stanley Assistant 用于检索内部知识；Debrief 在获得客户同意后生成会议笔记、行动项和客户邮件初稿，并写入 Salesforce 供顾问审核。该案例说明，高价值路径是缩短顾问的准备和行政时间，扩大其服务能力，而非让模型独立作出投资决策。

FactSet 与 TIFIN.AI 的合作也体现了这一方向：通过数据和 Agent 工作流帮助顾问做客户服务与个性化沟通，而非取消人工控制。

### 4. 零售侧：信息助手，而非自动荐股

Robinhood Cortex 会基于市场数据、分析师报告、研究报告及平台数据解释个股异动，并提供投资组合摘要等能力。但其协议明确：输出仅供信息和教育用途，不构成购买、卖出、持有证券或采取投资策略的建议。

这反映了零售金融 Agent 的现实边界：产品可以提升理解和参与度，却必须谨慎处理“个性化建议”“预测收益”和“自动执行”之间的法律与信任风险。

### 5. 中国市场：券商将业务能力产品化为 Skills

2026 年，中国多家券商开始把分析师研究框架、研报、行情、筛选、ETF/基金比较、财富和运营能力封装为可调用 Skills。该路线并不追求完全开放的自主 Agent，而是优先把已验证、可审核的业务能力封装为受控工具，再通过统一入口编排。

相对海外，中国市场更需要处理本地数据与研报授权、投资者适当性、内容审核、私有化部署和证券业务责任边界。领先券商的机会在于将独有投研资产与客户服务系统连接起来；其挑战则是将能力做成可复用产品，而非一次性的内部演示。

## 竞争格局与关键判断

### 数据、工作流与信任，构成三层壁垒

1. **数据壁垒**：实时行情、财务指标、研报、电话会、私有市场和内部客户数据，都有授权、时效与版本要求。
2. **工作流壁垒**：投行的模型与 pitch、私募信贷的条款审查、研究员的覆盖框架、财富顾问的客户服务，交付物和审批流程完全不同。
3. **信任壁垒**：引用、数值核验、权限、审批、日志、版本回溯和责任划分，在金融业是产品功能而非后台要求。

### 市场正在从“模型能力竞争”转为“可用结果竞争”

基座模型能力的差距在缩小，而金融客户购买的是：能否使用合规数据、能否嵌入已有工作台、能否输出可审查的材料、能否降低错误和运营风险。因此，平台型数据商与垂直工作流创业公司会长期并存：前者提供数据与分发，后者优化角色体验与任务闭环。

### 近 12–24 个月最可能规模化的场景

- 投研、私募信贷和投行的文档尽调、条款提取、指标标准化与 IC memo 初稿；
- 研究报告、财报电话会和事件驱动的持续监控；
- Excel 模型更新、来源对账和 PowerPoint 材料生成；
- 财富顾问的会前准备、会议记录、客户分层、合规沟通初稿；
- Agent 治理：数据权限、引用核验、审批、评测与留痕。

## 主要风险与合规边界

- **事实与时点风险**：模型可能出现幻觉、数字抄错、旧数据引用或忽略关键限定条件。
- **数据权利风险**：行情、研报、卖方研究、客户资料及内部模型均有严格授权边界。
- **适当性与责任风险**：个性化推荐、利益冲突、账户适当性和投顾牌照责任不能因使用 AI 而消失。
- **权限风险**：连接 CRM、邮箱、文档库和交易系统后，Agent 的越权调用与数据外泄风险显著上升。
- **可解释性风险**：对于投资结论、信用判断和对外材料，必须能够说明来源、计算过程和人工审批人。

美国证券交易委员会已将基金或投顾使用 AI Agent 视为重要讨论方向，但传统的投顾义务与对客户的责任并未被替代。

## 产品机会建议

若从零切入，优先级建议如下：

1. **投研/私募信贷尽调 Agent**：聚焦文档抽取、指标标准化、条款比较、IC memo 初稿与证据链；价值易衡量，且不触及交易执行。
2. **券商财富管理中间层**：在合规边界内连接研究内容、客户画像、产品池和 CRM，作为顾问助手而不是自动投顾。
3. **金融 Agent 治理层**：提供权限、数据血缘、引用验证、审批、评测和留痕；这是所有高价值金融 Agent 的共同刚需。
4. **Excel/PPT 原生 Agent**：完成“拉数—计算—复核—排版—引用”的闭环，减少分析师在交付物之间切换的成本。

不建议以自动荐股或自动交易作为首个切入点：合规、责任与信任成本最高，且大型券商、数据商和交易平台拥有更强的渠道与数据优势。

## 参考来源

### 产品与市场

- [FactSet：AI for Banking 发布公告](https://investor.factset.com/news-releases/news-release-details/factset-accelerates-innovation-banking-launch-new-ai-native/)
- [FactSet：金融服务中 Agent 的数据准确性与 RAG](https://insight.factset.com/what-makes-ai-agents-work)
- [LSEG：Workspace AI Search](https://www.lseg.com/en/data-analytics/products/workspace/updates/act-with-the-same-confidence-at-a-new-speed-introducing-lseg-workspace-ai-search)
- [S&P Global：Capital IQ Pro](https://www.spglobal.com/market-intelligence/en/solutions/products/sp-capital-iq-pro)
- [Rogo：Agent Library](https://rogo.ai/news/agent-library)
- [Rogo：Felix、连接器与自定义 Agent 更新](https://rogo.ai/news/may-product-update)
- [Hebbia：Matrix 产品页](https://www.hebbia.com/product)
- [Hebbia：股票研究团队使用场景](https://www.hebbia.com/blog/5-ways-equity-research-teams-use-hebbia-to-drive-speed-and-insight)
- [Morgan Stanley：AI @ MS Debrief](https://www.morganstanley.com/press-releases/ai-at-morgan-stanley-debrief-launch)
- [Robinhood：Cortex 产品说明](https://robinhood.com/gb/en/learn/articles/cortex-digests-is-here/)
- [Robinhood：Cortex Assistant 协议](https://cdn.robinhood.com/assets/robinhood/legal/cortex_agreement.pdf)
- [中国金融信息网：券商 Skills 产品化](https://www.cnfin.com/yw-lb/detail/20260602/4420519_1.html)

### 监管与责任

- [SEC：人工智能与投资管理的未来](https://www.sec.gov/newsroom/speeches-statements/daly-020326-artificial-intelligence-future-investment-management)
- [SEC：对数字化投顾的观察](https://www.sec.gov/newsroom/whats-new/observations-examinations-advisers-provide-electronic-investment-advice)
