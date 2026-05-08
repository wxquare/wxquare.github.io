# 两份一页简历设计说明

## 目标

基于现有 `source/about/` 简历与面试材料，产出两份可投递的一页中文简历：

1. 后端架构版：面向后端架构、电商核心系统、交易平台、平台化方向。
2. AI Agent 版：面向 AI Agent 工程、Agent 应用后端、AI 平台后端方向。

每份简历都需要提供 `Markdown + HTML + PDF` 三种产物，并控制在 A4 一页内。

## 现有材料

- `source/about/resume-2026.md`：后端架构基础版本。
- `source/about/resume-2026.html`：后端架构 HTML 打印版本。
- `source/about/resume-2026-ai-agent.md`：AI Agent 方向基础版本。
- `source/about/resume-2026-ai-agent.html`：AI Agent HTML 打印版本。
- `source/about/resume-project-description.md`：统一商品运营管理平台项目深讲。
- `source/about/interview-talking-points.md`：自我介绍、核心项目话术与追问材料。

## 后端架构版设计

定位为 `后端架构 / 电商核心系统`。

核心叙事是“复杂业务平台化 + 高并发交易链路 + 资金安全 + 稳定性治理”。正文优先保留 Shopee 阶段的商品、库存、价格、OTA 与稳定性成果：

- 统一商品运营平台：突出 10+ 品类统一建模、策略模式、状态机、异步上架、批量操作框架、新品类 2 天接入。
- 多品类统一库存系统：突出二维库存模型、Redis Lua 原子扣减、2 万 QPS、日均 200 万订单、库存一致性与对账。
- 统一价格计算引擎：突出 DDD、四层计价模型、5000 万+ 日调用、双快照、灰度迁移、0 资损、P99 优化。
- OTA 品类从 0 到 1：作为复杂业务交付能力补充，压缩描述 Hotel / Ferry / Flight 全链路与复用收益。
- 系统稳定性：只保留告警治理、大促保障、限流降级、资金安全等强相关内容。

AI Agent 内容在后端架构版中只作为告警治理或工程效率的一句补充，不作为主项目展开。

## AI Agent 版设计

定位为 `AI Agent 工程 / 后端架构`。

核心叙事是“把后端工程能力迁移到 Agent 生产化”。正文第一项目放 DoD 告警处理 Agent，后续用电商核心系统证明候选人具备真实生产系统经验：

- DoD 告警处理 Agent：突出状态机 + ReACT 混合架构、RAG、MCP 工具调用、风险分级、审计、trace、自动处理率与响应时间改善。
- Agent 工程知识体系：作为技术输出放在简历末尾，突出 AI Agent 系统设计文档、DoD Agent 设计文档、工具链实践。
- 电商核心系统架构：压缩商品、库存、价格三个项目为生产级后端能力支撑，强调异步任务、状态机、幂等、一致性、可观测性和灰度迁移。
- 腾讯 AI 广告经历：保留 AI 原生广告植入与 TVM 模型优化，增强 AI 工程连续性。

OTA 内容在 AI Agent 版中只保留一条，作为复杂业务背景，不占用主要篇幅。

## 文件产物

后端架构版：

- `source/about/resume-backend-architect.md`
- `source/about/resume-backend-architect.html`
- `source/about/resume-backend-architect.pdf`

AI Agent 版：

- `source/about/resume-ai-agent.md`
- `source/about/resume-ai-agent.html`
- `source/about/resume-ai-agent.pdf`

保留现有 `resume-2026*` 文件，不强行覆盖，避免破坏已有版本。

## 排版约束

- A4 一页。
- 中文简历，面向国内技术岗位。
- 简洁黑白打印风格，避免过度装饰。
- 页眉包含姓名、电话、邮箱、工作经验、求职意向。
- 教育背景放在前部，但压缩为两行。
- Shopee 为主体，腾讯为补充。
- 每条项目经历使用“动作 + 技术方案 + 结果指标”的结构，避免只有技术名词堆叠。
- 指标优先保留与岗位最匹配的数字，删除重复或弱相关指标。

## 验证方式

- 生成 HTML 后检查 A4 打印样式。
- 生成 PDF 后确认每份只有 1 页。
- 运行博客构建验证：`npm run clean && npm run build`。
- 不修改 `themes/`、`db.json`、`.deploy_git/`、`node_modules/`。

## 风险与取舍

- 一页限制会压缩项目细节，因此保留深讲材料在 `source/about/resume-project-description.md` 与 `source/about/interview-talking-points.md`。
- 后端架构版优先系统复杂度与业务结果，AI Agent 版优先 Agent 生产化能力，两者不追求内容完全一致。
- PDF 导出依赖本地可用的浏览器或 HTML 转 PDF 工具；若环境缺少依赖，需要使用可用工具替代并说明。
