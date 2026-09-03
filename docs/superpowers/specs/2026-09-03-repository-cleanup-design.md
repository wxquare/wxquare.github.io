# 仓库遗留文件清理设计

## 目标

删除已经指向不存在书籍目录的 GitHub Actions Workflow，并清理本地可由构建重新生成的目录和系统噪音，减少目录干扰；不触碰博客内容、资料迁移、主题、源码或现有线上 URL。

## 范围

### 删除

- `.github/workflows/deploy-ecommerce-book.yml`
- `.github/workflows/deploy-system-design-book.yml`
- 本地 `public/` 生成目录
- 本地 `db.json` Hexo 缓存数据库
- 已确认的 `.DS_Store` 文件

### 保留

- `source/`、`books/`、`themes/`、`scripts/` 和 `test/` 中的源码与内容
- 现有 `source/library/`、`source/presentations/` 迁移改动
- `deploy-ai-book.yml` 与 `deploy-system-design-architecture-book.yml`
- 其余配置、依赖和文档

## 实施方式

1. 先记录清理前 Git 状态，并确认待删除 Workflow 的路径确实不存在。
2. 使用 Git 删除两个失效 Workflow，使删除可审查、可恢复。
3. 只删除根目录下当前已确认的 `public/`、`db.json` 和 `.DS_Store` 文件；不使用递归通配符触及其他目录。
4. 检查 Git 状态，确认未提交的资料迁移改动仍保持原状。

## 兼容性与风险控制

- `public/` 和 `db.json` 均是 Hexo 可再生输出，删除不改变 Git 跟踪源码。
- `.DS_Store` 仅为 macOS Finder 元数据，不参与构建。
- 两个被删除的 Workflow 引用 `books/ecommerce-book` 和 `books/system-design-book`，这两个目录当前不存在；删除它们不会影响现有两个书籍 Workflow。
- 本次不移动或重命名任何 `source/` 内容，不改变已有页面、资源和历史 URL。

## 验证

- 删除后确认五类目标均不存在。
- 运行现有测试：`node --test test/legacy-presentation-alias.test.js`。
- 运行站点构建：`npm run build`，确认构建成功后再次执行 `npm run clean`，让工作区回到无生成物状态。
- 最终检查 `git status --short`，确认只出现本设计范围内的 Workflow 删除，且用户已有迁移改动未被覆盖。
