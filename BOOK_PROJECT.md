# 电商书籍项目说明文档

**项目路径**：`/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/books/ecommerce-book/`

**在线地址**：https://wxquare.github.io/ecommerce-book/

**最后更新**：2026-04-17

---

## 📋 项目概况

### 项目结构

```
books/ecommerce-book/
├── book.toml              # mdBook配置文件
├── src/                   # 源文件
│   ├── SUMMARY.md         # 目录结构（重要！）
│   ├── README.md          # 首页
│   ├── part1/             # 第一部分（4章）
│   ├── part2/             # 第二部分（11章）
│   │   ├── overview/      # 全局架构
│   │   ├── supply/        # 商品供给
│   │   └── transaction/   # 交易链路
│   ├── part3/             # 第三部分（1章）
│   └── appendix/          # 附录
├── theme/                 # 自定义样式
│   └── custom.css
└── migrate.sh             # 内容迁移脚本
```

### 内容来源

内容从 `source/book/` 迁移而来（已完成16章）

---

## 🚀 日常使用

### 更新内容

```bash
# 1. 编辑文件
vim books/ecommerce-book/src/part1/chapter1.md

# 2. 提交更改
cd /Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io
git add books/ecommerce-book/
git commit -m "Update chapter1: add new section"
git push origin hexo

# 3. 等待2-3分钟
# GitHub Actions自动构建并部署

# 4. 访问查看
# https://wxquare.github.io/ecommerce-book/
```

### 添加新章节

```bash
# 1. 创建文件
echo "# 新章节标题" > books/ecommerce-book/src/part3/chapter17.md

# 2. 在SUMMARY.md中注册
vim books/ecommerce-book/src/SUMMARY.md
# 添加：- [第17章 标题](part3/chapter17.md)

# 3. 提交
git add .
git commit -m "Add chapter17"
git push
```

---

## ⚙️ 本地开发

### ⚠️ 已知问题

**本地mdBook 0.5.2版本有bug**，会报错：
```
ERROR Rendering failed
  Caused by: Error rendering "index" line 177, col 29: Missing font github
```

### 解决方案

**方案A：不使用本地预览（推荐）**
- 直接编辑Markdown文件
- 推送到GitHub让云端构建
- 2-3分钟后访问 https://wxquare.github.io/ecommerce-book/

**方案B：使用Docker本地预览**
```bash
# 创建脚本
cat > books/ecommerce-book/serve-docker.sh << 'EOF'
#!/bin/bash
docker run --rm -v $(pwd):/book -p 3000:3000 \
  peaceiris/mdbook:v0.4.40 \
  mdbook serve --hostname 0.0.0.0
EOF

chmod +x books/ecommerce-book/serve-docker.sh

# 运行
cd books/ecommerce-book
./serve-docker.sh

# 访问 http://localhost:3000
```

---

## 🔧 配置说明

### book.toml

```toml
[book]
title = "电商系统架构设计与实现"
authors = ["wxquare"]
language = "zh-CN"

[output.html]
site-url = "/ecommerce-book/"        # 重要！部署路径
git-repository-url = "..."           # GitHub仓库链接

[output.html.search]
enable = true                         # 启用搜索
```

### GitHub Actions

**工作流文件**：`.github/workflows/deploy-ecommerce-book.yml`

**触发条件**：
- 推送到 `hexo` 分支
- 修改了 `books/ecommerce-book/` 目录

**构建版本**：mdBook 0.4.40（稳定版本）

---

## 📊 部署流程

```mermaid
graph LR
    A[编辑Markdown] --> B[git commit]
    B --> C[git push]
    C --> D[GitHub Actions触发]
    D --> E[mdBook 0.4.40构建]
    E --> F[部署到gh-pages分支]
    F --> G[GitHub Pages发布]
    G --> H[https://wxquare.github.io/ecommerce-book/]
```

---

## 🐛 故障排查

### 问题1：GitHub Actions失败

**查看日志**：
```
https://github.com/wxquare/wxquare.github.io/actions
```

**常见原因**：
- Markdown语法错误
- 链接格式错误
- SUMMARY.md中的路径不匹配

### 问题2：部署成功但404

**检查**：
1. GitHub Pages是否启用
   ```
   https://github.com/wxquare/wxquare.github.io/settings/pages
   ```
2. 确认Source为"Deploy from a branch"，Branch为"gh-pages"

### 问题3：内容未更新

**原因**：浏览器缓存

**解决**：
- 强制刷新：Cmd+Shift+R（Mac）或 Ctrl+Shift+R（Windows）
- 或使用无痕模式访问

---

## 📁 内容管理

### 目录结构规则

所有章节必须在 `src/SUMMARY.md` 中注册：

```markdown
# 《电商系统架构设计与实现》

[书籍介绍](README.md)

# 第一部分

- [第1章 标题](part1/chapter1.md)
  - [1.1 小节](part1/chapter1.md#11-小节)
```

### 链接格式

**正确**：
```markdown
[第2章](chapter2.md)
[跨目录链接](../part2/overview/chapter5.md)
[锚点链接](chapter1.md#11-section)
```

**错误**：
```markdown
[第2章](./chapter2.html)  ❌ HTML扩展名
[第2章](chapter2)          ❌ 缺少.md
```

### 图片引用

```markdown
# 方法1：相对路径（推荐）
![架构图](images/architecture.png)

# 方法2：如果图片在其他目录
![架构图](../../source/book/images/architecture.png)
```

---

## 🎨 样式定制

**自定义CSS**：`books/ecommerce-book/theme/custom.css`

**修改后**：推送到GitHub，自动应用

**主要定制**：
- 中文字体优化
- 代码块样式
- 表格美化
- 打印优化

---

## 📝 内容来源关系

### 原始位置
```
source/book/
├── index.md
├── chapter1.md
├── chapter2.md
└── ...
```

### 迁移后
```
books/ecommerce-book/src/
├── README.md          ← source/book/index.md
├── part1/chapter1.md  ← source/book/chapter1.md
├── part1/chapter2.md  ← source/book/chapter2.md
└── ...
```

### ⚠️ 重要：避免双重维护

**选择一个主要位置**：

**方案A（推荐）**：以mdBook为主
- 将 `source/book/` 改为只读（或删除）
- 今后只更新 `books/ecommerce-book/src/`

**方案B**：保持同步
- 保留两边，但需手动同步
- 更新后运行：`cd books/ecommerce-book && ./migrate.sh`

---

## 🔗 重要链接

### 开发相关
- **GitHub仓库**：https://github.com/wxquare/wxquare.github.io
- **Actions日志**：https://github.com/wxquare/wxquare.github.io/actions
- **Pages设置**：https://github.com/wxquare/wxquare.github.io/settings/pages

### 文档
- **在线电子书**：https://wxquare.github.io/ecommerce-book/
- **mdBook官方文档**：https://rust-lang.github.io/mdBook/
- **设计方案文档**：`docs/superpowers/specs/2026-04-17-mdbook-ebook-design.md`
- **使用教程**：`docs/mdbook-tutorial.md`

---

## ✅ 检查清单

### 每次更新内容后

- [ ] 本地编辑Markdown文件
- [ ] git add + commit + push
- [ ] 访问GitHub Actions确认构建成功（绿色✅）
- [ ] 访问在线地址验证内容更新
- [ ] 强制刷新浏览器（Cmd+Shift+R）

### 添加新章节时

- [ ] 创建Markdown文件
- [ ] 在SUMMARY.md中注册
- [ ] 确认路径正确
- [ ] 测试章节链接能否打开

---

## 📞 获取帮助

### 如果遇到问题

1. **查看GitHub Actions日志**
   - 构建失败的详细错误信息
   
2. **检查配置文件**
   - `book.toml` 语法是否正确
   - `SUMMARY.md` 路径是否匹配

3. **验证Markdown语法**
   - 代码块是否有语言标识
   - 链接格式是否正确

---

## 📚 快速命令参考

```bash
# 查看项目状态
cd /Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io
git status books/ecommerce-book/

# 编辑内容
vim books/ecommerce-book/src/part1/chapter1.md

# 提交更改
git add books/ecommerce-book/
git commit -m "Update: ..."
git push origin hexo

# 查看构建日志（浏览器）
open https://github.com/wxquare/wxquare.github.io/actions

# 访问电子书（浏览器）
open https://wxquare.github.io/ecommerce-book/

# Docker本地预览（如果配置了）
cd books/ecommerce-book
./serve-docker.sh
```

---

**维护者**：wxquare  
**创建日期**：2026-04-17  
**最后更新**：2026-04-17
