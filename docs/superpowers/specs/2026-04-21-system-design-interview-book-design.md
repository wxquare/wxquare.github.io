# 《程序员系统设计与面试指南》mdBook 设计规格

**项目**：基于现有 Hexo 博文聚合的 mdBook 电子书  
**日期**：2026-04-21  
**状态**：待你审阅  
**代码骨架**：`system-design-book/`（已存在，占位章节可构建）

---

## 1. 背景与目标

### 1.1 背景

仓库中已有成体系的系统设计、计算机基础与 LeetCode 分类长文（`source/_posts/system-design/`、`source/_posts/fundamentals/`）。需要以 **mdBook** 形式提供：

- 单一路径的「书」式阅读体验（目录、搜索、折叠）
- 与现有 `ai-book`、`ecommerce-book` 一致的工程与部署习惯

### 1.2 目标读者

**兼顾两类读者**：

1. **面试准备**：快速定位中间件原理、系统设计题与算法题型  
2. **日常工作**：技术方案、容量与可靠性、基础栈查阅

### 1.3 成功标准

1. `mdbook build` 在 CI 或本地无错误  
2. 各章正文可从源稿完整迁移，**图片与站内链**在 mdBook 站点下可用  
3. **不设全书面试题总附录**；面试题与速查表**下沉到各章末尾**（见 §4.3）  
4. LeetCode 大文 **独立为第六部分多章**，不作为附录 A

### 1.4 非目标（本期不做）

- 不修改 Hexo 主题目录 `themes/`  
- 不要求首版即完成 PDF（可作为后续迭代）  
- 不在规格中承诺固定「N 道」面试题数量

---

## 2. 内容来源与章节映射

| 章 | mdBook 路径 | 主要源稿 |
|----|-------------|----------|
| 1 | `src/part01/chapter01.md` | `source/_posts/system-design/00-system-design-overview.md` |
| 2 | `src/part01/chapter02.md` | `source/_posts/system-design/06-tech-design-methodology.md` |
| 3 | `src/part02/chapter03.md` | `source/_posts/system-design/01-middleware-mysql.md` |
| 4 | `src/part02/chapter04.md` | `source/_posts/system-design/02-middleware-redis.md` |
| 5 | `src/part02/chapter05.md` | `source/_posts/system-design/03-middleware-kafka.md` |
| 6 | `src/part02/chapter06.md` | `source/_posts/system-design/04-middleware-elasticsearch.md` |
| 7 | `src/part02/chapter07.md` | `source/_posts/system-design/05-infrastructure-k8s-docker.md` |
| 8 | `src/part03/chapter08.md` | `source/_posts/system-design/07-system-reliability-engineering.md` |
| 9 | `src/part04/chapter09.md` | `source/_posts/system-design/08-system-design-interview.md` |
| 10 | `src/part05/chapter10.md` | `source/_posts/fundamentals/1-os-fundamentals.md` |
| 11 | `src/part05/chapter11.md` | `source/_posts/fundamentals/2-network-fundamentals.md` |
| 12 | `src/part05/chapter12.md` | `source/_posts/fundamentals/3-bash-shell.md` |
| 13 | `src/part05/chapter13.md` | `source/_posts/fundamentals/4-python-practice.md` |
| 14 | `src/part05/chapter14.md` | `source/_posts/fundamentals/5-cpp-practice.md` |
| 15 | `src/part05/chapter15.md` | `source/_posts/fundamentals/6-golang-practice.md` |
| 16–21 | `src/part06/chapter16.md` … `chapter21.md` | `source/_posts/system-design/09-leetcode.md` **按大节拆分**（见 §3.2） |

---

## 3. 信息架构与叙述约定

### 3.1 全书结构（与 `SUMMARY.md` 一致）

1. **第一部分**：导论与方法论（第 1–2 章）  
2. **第二部分**：核心中间件与基础设施（第 3–7 章）  
3. **第三部分**：系统可靠性工程（第 8 章）  
4. **第四部分**：系统设计面试专题（第 9 章）  
5. **第五部分**：计算机基础（第 10–15 章）  
6. **第六部分**：算法与编码 / LeetCode（第 16–21 章）  
7. **附录**：术语表、参考文献、工具与构建说明（**不**承载「全书面试题汇总」）

### 3.2 LeetCode 源稿拆分（第 16–21 章）

| 章 | 建议承载内容（对应 `09-leetcode.md` 一级标题） |
|----|-----------------------------------------------|
| 16 | `## 数据结构` 全文 |
| 17 | `## 基本算法` 全文 |
| 18 | `## 数学 (Mathematics - 核心模式归类)` 全文 |
| 19 | `## 搜索问题核心分类与总结 (Search Strategies)` 全文 |
| 20 | `## DP 问题 (Dynamic Programming - 核心模式归类)` 全文 |
| 21 | `## C++ 字符处理函数速查`；文末 `## 参考` 可并入 **附录 B** 或保留为短「延伸阅读」 |

成书时应对原文 **重复编号的小节**（如多个 `### 4`、`### 28`）在 mdBook 中 **重新编号**，避免目录混乱。

### 3.3 中间件与设计章节的统一体例

每章（第 3–8 章及第 10–15 章建议）按顺序组织：

1. **核心原理与机制**（模型、关键路径、与其他组件关系）  
2. **基础与日常使用**（配置、API、典型场景）  
3. **进阶与生产实践**（调优、故障、案例、监控）  
4. **本章面试题与追问**（题量**不封顶**；迁入原「面试高频 N 题」「速查」等并可持续增补）

第 1 章总览中的「高频题汇总表」迁移时改为 **指向各章 § 本章面试题与追问** 的链接索引，避免与章末内容双份维护。

### 3.4 第 9 章（系统设计面试专题）

保持 **独立成篇**：按源稿「一、二、三…」大节组织；每大节下追问不限量。

---

## 4. 技术方案

### 4.1 工程位置与构建

- **根目录**：`system-design-book/`  
- **配置**：`book.toml`（中文书名、`site-url` 预留为 `/system-design-book/`）  
- **Mermaid**：与 `ai-book` 相同，`book.toml` 同级的 `mermaid.min.js`、`mermaid-init.js`；`mdbook-mermaid` 为 **optional**  
- **验证**：内容变更后执行 `mdbook build`；合并前必须通过

### 4.2 与 Hexo 博客规范对齐（迁移时必须执行）

- 去掉 Hexo **Front Matter**（mdBook 不使用）  
- 代码块 **必须** 带语言标识（与 `.cursorrules` 一致）  
- 中文正文 **中英文之间空格**  
- 图片：博客规则为**相对路径**；迁入后需验证在 mdBook 输出站点下是否解析正确（必要时复制到 `system-design-book/src/assets/` 并改路径）  
- 原站内绝对路径（如 `/2024/03/...`）在书中改为 **相对章节链接** 或 **外链说明**

### 4.3 「附录 B」策略（已定稿）

- **不再**维护单独「全书面试题附录 B」  
- 各源稿中的速查、高频题表，归入对应章的 **「本章面试题与追问」**  
- 全书附录仅保留：**术语表、参考文献汇总、工具与构建**

### 4.4 部署（后续迭代）

- 对齐 `ecommerce-book`：新增 `.github/workflows/deploy-system-design-book.yml`，`paths` 监听 `system-design-book/**`，产出可挂 GitHub Pages 的 `book/`（本规格不强制首版即含 workflow，但推荐与另两本书一致）

---

## 5. 风险与依赖

| 风险 | 缓解 |
|------|------|
| 单篇过长（MySQL、LeetCode） | LeetCode 已拆 16–21 章；MySQL 可按需拆子文件但保持 SUMMARY 一层入口 |
| 图片路径失效 | 迁移时统一走 `src/assets/` 或保留与博客一致的相对资源策略并实测 |
| Mermaid 与 Hexo 插件差异 | 以 mdBook 为准；复杂图可保留为 mermaid 代码块 |

---

## 6. 自检（规格成文时已完成）

- [x] 无「TBD」占位句  
- [x] 章节与源稿映射完整、与现有 `system-design-book/src/SUMMARY.md` 一致  
- [x] 面试题策略与 LeetCode 独立成篇与对话结论一致  
- [x] 非目标与仓库禁止项（不改 `themes/`）已写明  

---

## 7. 审阅通过后下一动作

按 brainstorming 流程：你确认本规格无修改后，再编写 **`docs/superpowers/plans/2026-04-21-system-design-interview-book-implementation.md`**（实现计划：迁移顺序、链接与图片检查、`mdbook build`、可选 CI），然后按计划在仓库中执行迁移与填充正文。
