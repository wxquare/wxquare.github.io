# 电商搜索与导购文章 Implementation Plan

> **For agentic workers:** 按下方 checkbox 顺序执行；每完成一章可在 Git 中单次或分批提交。本仓库为 Hexo 静态站，验收以 `npm run clean && npm run build` 通过为准。

**Goal:** 在 `source/_posts/system-design/31-ecommerce-search-discovery.md` 发布符合 `docs/superpowers/specs/2026-04-13-ecommerce-search-discovery-design.md` 的系列长文（面试向 + 工程向），并与 `20`～`30` 篇建立交叉引用。

**Architecture:** 单文件 Markdown；主线为「统一导购查询服务（多 scene）」+「Elasticsearch 查询侧专章」；索引形态与同步权威引用 `27` 第 5 章，不写营销算价与推荐 feed 主文。

**Tech Stack:** Hexo 7、Markdown、Mermaid、中文标点与中英文空格规范。

---

## 文件映射

| 文件 | 职责 |
|------|------|
| `source/_posts/system-design/31-ecommerce-search-discovery.md` | 新建正文（Front Matter + 12 章结构 + 系列导航） |
| `docs/superpowers/specs/2026-04-13-ecommerce-search-discovery-design.md` | 已存在，仅当正文偏离 scope 时回溯修订 |

---

### Task 1：Front Matter 与系列导航

**Files:**

- Create: `source/_posts/system-design/31-ecommerce-search-discovery.md`（首屏含 Front Matter + 与 `30` 衔接的系列说明 + 指向 `20` 总索引）

- [ ] **Step 1:** `title` 使用「（十二）」与 `30` 的「（十一）」连续；`date: '2026-04-13'`（字符串）；`categories: [电商系统设计]`；`tags` 与 spec 一致（小写连字符）。

- [ ] **Step 2:** 文首引用块列出 `20`、`27`、`21`、`30`、`25`、`23`、`22`、`28` 的站内路径（与既有文章 `/system-design/slug/` 格式一致）。

**验收:** Front Matter 无 YAML 语法错误；系列篇次不自相矛盾。

---

### Task 2：第 1～2 章（引言、范围、分工表）

**Files:**

- Modify: `source/_posts/system-design/31-ecommerce-search-discovery.md`

- [ ] **Step 1:** 写引言：GMV / 体验、读路径与交易路径差异。

- [ ] **Step 2:** 显式「范围 / 非目标」列表（与 spec 一致：无首页 feed 主文、无营销算价实现）。

- [ ] **Step 3:** Markdown 表：与 `27`/`21`/`30`/`25`/`23`/`22`/`28` 分工（去重规则一段话 + 表）。

**验收:** 读者能一句话说清「本篇不写什么」。

---

### Task 3：第 3 章（统一导购查询服务）

**Files:**

- Modify: `source/_posts/system-design/31-ecommerce-search-discovery.md`

- [ ] **Step 1:** `scene` 对比表：`keyword` / `category` / `shop`（是否有 query、固定 filter、索引范围）。

- [ ] **Step 2:** Mermaid `graph TB` 或 `flowchart`：网关 → 导购查询服务 → ES / 排序 / hydrate → 商品读服务与计价只读等。

- [ ] **Step 3:** 「演进：何时拆 BFF」小节 1 段（可选篇幅 ≤ 15 行）。

**验收:** 面试可手绘等价框图。

---

### Task 4：第 4～5 章（Query、召回、排序、AB）

**Files:**

- Modify: `source/_posts/system-design/31-ecommerce-search-discovery.md`

- [ ] **Step 1:** Query 理解：归一化、同义词、纠错复杂度上限声明。

- [ ] **Step 2:** 粗排 / 精排 / 重排分层 + 合规过滤默认建议（召回后 filter vs 重排：选一种主叙事并写理由）。

- [ ] **Step 3:** `rank_version`、`exp_id`、`query_id` 在日志中的位置说明。

- [ ] **Step 4:** 「可选：向量多路召回」压缩为一节（≤ 一页）。

**验收:** 与 `28` 的边界一句点明（只读消费标签）。

---

### Task 5：第 6 章（Elasticsearch 专题）

**Files:**

- Modify: `source/_posts/system-design/31-ecommerce-search-discovery.md`

- [ ] **Step 1:** 分析链与中文分词原则；`keyword` vs `text`、`doc_values`、nested 反模式。

- [ ] **Step 2:** 至少一个完整 `json` 代码块：典型 `bool` + `filter` + `sort`（语言标签 `json`）。

- [ ] **Step 3:** 深分页：`search_after` 示例字段与面试表述；禁止深 `from/size` 的理由。

- [ ] **Step 4:** 慢查询与容量清单（分片副本、profile、冷热索引一句）。

**验收:** 明确写「索引字段权威见 `27` 5.1」，本篇不出现大段重复 mapping。

---

### Task 6：第 7 章（集成与序列图）

**Files:**

- Modify: `source/_posts/system-design/31-ecommerce-search-discovery.md`

- [ ] **Step 1:** 责任表扩展：商品中心、上架、生命周期、运营配置、计价、库存、营销。

- [ ] **Step 2:** Mermaid `sequenceDiagram`：领域事件 → 索引更新 → ES 可见（与 `27`/`21` 叙述对齐，Topic 名用通用 `product.*` / `listing.*` 类命名或「消息总线」描述，避免与实现漂移）。

- [ ] **Step 3:** Mermaid `sequenceDiagram`：用户列表请求 → ES → hydrate。

- [ ] **Step 4:** 至少一次投递下的幂等：`document_id` + `version` / `updated_at` 比较策略文字 + 可选 `go` 伪代码 ≤ 30 行。

**验收:** 两图可读；幂等段落可单独背诵。

---

### Task 7：第 8～10 章（一致性、可观测、工程清单）

**Files:**

- Modify: `source/_posts/system-design/31-ecommerce-search-discovery.md`

- [ ] **Step 1:** 索引滞后、hydrate 部分失败、ES 故障降级（推荐默认 + 备选权衡）。

- [ ] **Step 2:** 指标表：零结果率、P99、hydrate 成功率、慢查询、实验分桶。

- [ ] **Step 3:** reindex / 发布风险一句；与 `27` 缓存刷新策略呼应一句。

**验收:** 每章有「可落地检查项」列表。

---

### Task 8：第 11～12 章（面试锦囊、总结与索引）

**Files:**

- Modify: `source/_posts/system-design/31-ecommerce-search-discovery.md`

- [ ] **Step 1:** 15～25 条问答式要点，覆盖深分页、相关性 vs 商业化、ES vs DB、列表价一致、幂等、压测。

- [ ] **Step 2:** 总结 + 系列链接 + 「扩展阅读：推荐系统」一句。

**验收:** 条数 ≥ 15。

---

### Task 9：构建与提交

**Files:**

- Shell: 仓库根目录

- [ ] **Step 1:** 运行 `npm run clean && npm run build`

**期望:** 进程退出码 0；无 Hexo 报错。

- [ ] **Step 2:** `git add` 新文章与 plan；`git commit -m "docs: add ecommerce search & discovery post and plan"`

**验收:** 工作区干净或仅剩无关本地改动。

---

## Plan 自检（对照 spec）

| Spec 章节 | Plan 覆盖 |
|-----------|-----------|
| 方案 1 + 3 合并 | Task 3 + 5 |
| 与 `27` 去重 | Task 2、5 |
| 集成与事件 | Task 6 |
| 面试锦囊 | Task 8 |
| 写作约束 / build | Task 1、9 |

无 TBD 占位；Topic 名刻意泛化以降低与代码漂移风险，与 spec「成文时 grep 对齐」兼容：若后续系列统一命名，可在修订提交中替换为精确 Topic。
