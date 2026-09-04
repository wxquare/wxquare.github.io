# 资料目录合并与整理设计

## 背景

当前仓库将资料分散在 `source/library/` 和 `source/presentations/` 两个源目录：前者保存第三方公开参考资料，后者保存本人创作或确认可公开的第一方演示资料。两个目录的内容边界清楚，但索引和入口分散；其中 `source/presentations/` 下的资料也缺少统一的资料库组织方式。

## 目标

- 物理上合并为一个资料根目录 `source/library/`。
- 保留资料来源和内容类型边界，避免把第一方演示误认为第三方资料。
- 统一资料索引和元数据格式。
- 保留已有公开 URL 的兼容访问能力。
- 不删除任何资料载荷，不修改与本任务无关的现有工作区改动。

## 目标结构

```text
source/library/
├── index.md
├── books/          # 第三方书籍
├── papers/         # 第三方论文
├── slides/         # 第三方演示资料
├── presentations/  # 第一方演示资料
└── other/          # 已有的受控补充资料
```

不再保留空的 `tutorials/` 分类及其 `.gitkeep` 文件。`other/` 继续保留，用于当前 DaSiamRPN 的 README、LICENSE 和结果图等确实无法归入其他类型、但来源和再分发边界明确的补充资料。

## 方案

### 源文件迁移

- 将 `source/presentations/ddia-reading-share-2020.pdf`、`ddia-reading-share-2022.pptx` 和 `k8s-network.pdf` 移动到 `source/library/presentations/`。
- 将 `source/presentations/index.md` 的条目合并进 `source/library/index.md`，并为 `k8s-network.pdf` 补充完整元数据。
- 删除 `source/presentations/index.md`、原目录下的资料载荷以及空的 `source/library/tutorials/.gitkeep`；不删除 `source/library/other/` 的现有内容。
- 保留 `source/library/slides/` 作为第三方演示资料目录，与第一方 `presentations/` 分开。

### 文档与规则

同步更新 `README.md`、`AGENTS.md` 和 `docs/library-policy.md`：统一说明 `source/library/` 是资料根目录，并定义 `slides/` 与 `presentations/` 的来源边界。README 的入口改为统一资料库索引，不再把 presentations 描述为独立源目录。

### URL 兼容

迁移后新的规范 URL 为 `/library/...`。扩展现有构建期 alias generator，使旧的 `/presentations/` 索引和旧的 `/presentations/<filename>` 文件地址继续可访问；已有 `/pdf/k8s-network.pdf` alias 继续从新的规范源读取。alias 只生成构建产物，不新增重复 Git 源文件、符号链接或伪装格式的文件。

### 大文件处理

本次只做目录和索引整理，不删除现有二进制资料。索引继续保留 SHA-256 和公开说明；对超过 10 MiB 的 `ddia-reading-share-2022.pptx`，在整理时明确记录其第一方公开确认及存储例外状态，避免与默认仅外链规则冲突。

## 验证

- 测试 alias generator 能从 `source/library/presentations/` 读取规范文件，并生成旧 URL。
- 检查仓库内不再存在活动的 `source/presentations/` 源文件或 `source/library/tutorials/.gitkeep`。
- 运行 `npm test`。
- 运行 `npm run clean && npm run build`，确认 Hexo 构建成功。
- 运行 `npm run check:links`，确认生成站点的内部链接没有因迁移失效。

## 不在本次范围内

- 不整理 `source/booklist/`、`source/pdf/` 或 `source/diagrams/`。
- 不删除、压缩或重新生成任何已有资料载荷。
- 不修改当前工作区中与本任务无关的文章、图片和图表变更。
