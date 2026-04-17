# mdBook完整教程 - 从零到部署

> 适合第一次使用mdBook的用户，每一步都有详细说明

**目标**：将《电商系统架构设计与实现》转换为mdBook电子书并部署

---

## 📋 前置检查清单

开始之前，确保您有：
- [ ] macOS系统（您当前使用）
- [ ] 终端（Terminal）基本使用能力
- [ ] GitHub账号
- [ ] 网络连接正常

---

## 第1部分：环境安装（预计30分钟）

### 步骤1：安装Rust工具链

mdBook是用Rust编写的，需要先安装Rust。

```bash
# 1. 打开终端，运行安装命令
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

**安装过程**：
```
Current installation options:
  ...
  
1) Proceed with installation (default)
2) Customize installation
3) Cancel installation
>
```
→ 直接按**回车**选择默认安装

**等待时间**：5-10分钟（取决于网络）

**安装完成标志**：
```
Rust is installed now. Great!
```

```bash
# 2. 让环境变量生效
source $HOME/.cargo/env

# 3. 验证安装
rustc --version
cargo --version
```

**预期输出**：
```
rustc 1.77.0 (aedd173a2 2024-03-17)
cargo 1.77.0 (8ac7396b7 2024-03-01)
```

**✅ 检查点**：如果看到版本号，继续下一步

---

### 步骤2：安装mdBook

```bash
# 安装mdBook（约5-10分钟）
cargo install mdbook

# 验证安装
mdbook --version
```

**预期输出**：
```
mdbook v0.4.37
```

**如果安装失败**：
```bash
# 可能是网络问题，使用国内镜像
mkdir -p ~/.cargo
cat > ~/.cargo/config.toml << 'EOF'
[source.crates-io]
replace-with = 'ustc'
[source.ustc]
registry = "https://mirrors.ustc.edu.cn/crates.io-index"
EOF

# 然后重新安装
cargo install mdbook
```

---

### 步骤3：安装可选插件

```bash
# PDF生成插件（推荐安装）
cargo install mdbook-pdf

# Mermaid图表支持（如果您的书中有流程图）
cargo install mdbook-mermaid
```

**提示**：这些可以之后再装，不影响基本使用

---

## 第2部分：快速体验（预计10分钟）

在迁移您的书之前，先创建一个测试项目熟悉mdBook。

### 步骤4：创建测试项目

```bash
# 1. 进入桌面（或任何您喜欢的位置）
cd ~/Desktop

# 2. 创建测试项目
mdbook init test-book
```

**交互问答**：
```
Do you want a .gitignore to be created? (y/n)
→ 输入 y 回车

What title would you like to give the book?
→ 输入 Test Book 回车
```

```bash
# 3. 进入项目目录
cd test-book

# 4. 查看结构
ls -la
```

**您会看到**：
```
├── book.toml          # 配置文件
└── src/
    ├── SUMMARY.md     # 目录结构（重要！）
    └── chapter_1.md   # 示例章节
```

---

### 步骤5：本地预览

```bash
# 启动开发服务器（会自动打开浏览器）
mdbook serve --open
```

**您会看到浏览器打开并显示**：
- 左侧：目录导航（Chapter 1）
- 中间：章节内容
- 右侧：页面大纲
- 顶部：搜索框、主题切换按钮

**试试这些功能**：
1. 点击左侧目录项
2. 点击右侧大纲跳转
3. 试试搜索功能
4. 切换主题（明亮/暗色）

**🎉 恭喜！您已经成功运行了第一个mdBook项目**

**关闭服务器**：回到终端按 `Ctrl+C`

---

### 步骤6：理解核心文件

#### 6.1 `book.toml` - 配置文件

```bash
# 打开配置文件
cat book.toml
```

**内容说明**：
```toml
[book]
authors = ["您的名字"]        # 作者
language = "en"              # 语言（改为zh-CN）
multilingual = false
src = "src"                  # 源文件目录
title = "Test Book"          # 书名

# 这是基础配置，后面会添加更多选项
```

#### 6.2 `src/SUMMARY.md` - 目录结构（最重要！）

```bash
# 查看目录文件
cat src/SUMMARY.md
```

**内容说明**：
```markdown
# Summary

- [Chapter 1](./chapter_1.md)
```

**这个文件定义了整本书的结构：**
- 每一行就是一个目录项
- `- [标题](文件路径.md)` 格式
- 支持嵌套（用缩进表示层级）

**示例：多层级目录**
```markdown
# Summary

[Introduction](README.md)

# Part 1

- [Chapter 1](chapter1.md)
  - [Section 1.1](chapter1.md#section-11)
  - [Section 1.2](chapter1.md#section-12)
- [Chapter 2](chapter2.md)

# Part 2

- [Chapter 3](chapter3.md)

---

[Appendix](appendix.md)
```

---

### 步骤7：编辑测试内容

让我们添加一个新章节来理解工作流程。

```bash
# 1. 编辑SUMMARY.md，添加第2章
cat > src/SUMMARY.md << 'EOF'
# Summary

- [Chapter 1](./chapter_1.md)
- [Chapter 2](./chapter_2.md)
EOF

# 2. 创建第2章内容
cat > src/chapter_2.md << 'EOF'
# Chapter 2

这是第二章的内容。

## 2.1 小节标题

这是一个小节。

## 2.2 代码示例

```go
func main() {
    fmt.Println("Hello, mdBook!")
}
```

## 2.3 表格示例

| 特性 | mdBook | 其他工具 |
|------|--------|----------|
| 易用性 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| 性能 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
EOF

# 3. 重新启动服务器
mdbook serve --open
```

**观察变化**：
- 左侧目录多了"Chapter 2"
- 点击可以查看新内容
- 代码高亮自动生效
- 表格正常渲染

**💡 重要发现**：
1. 只要编辑Markdown文件，保存后页面自动刷新
2. 添加新章节必须在SUMMARY.md中注册
3. mdBook会自动处理格式和样式

---

## 第3部分：正式项目（预计2-3小时）

现在我们来创建您的书籍项目。

### 步骤8：创建书籍项目

```bash
# 1. 回到您的博客目录
cd /Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io

# 2. 创建book项目
mdbook init ecommerce-book
```

**交互**：
```
Do you want a .gitignore to be created? (y/n)
→ y

What title would you like to give the book?
→ 电商系统架构设计与实现
```

```bash
# 3. 进入项目
cd ecommerce-book

# 4. 查看结构
tree -L 2
```

---

### 步骤9：配置book.toml

创建完整的配置文件：

```bash
cat > book.toml << 'EOF'
[book]
title = "电商系统架构设计与实现"
description = "从领域建模到工程落地：面向中大型团队的实战指南"
authors = ["wxquare"]
language = "zh-CN"
multilingual = false
src = "src"

[build]
build-dir = "book"

[output.html]
default-theme = "light"
preferred-dark-theme = "navy"
curly-quotes = true
mathjax-support = false
copy-fonts = true
no-section-label = false
git-repository-url = "https://github.com/wxquare/wxquare.github.io"
git-repository-icon = "fa-github"
site-url = "/ecommerce-book/"
cname = ""

[output.html.search]
enable = true
limit-results = 30
teaser-word-count = 30
use-boolean-and = true
boost-title = 2
boost-hierarchy = 1
boost-paragraph = 1
expand = true
heading-split-level = 3

[output.html.fold]
enable = true
level = 1

[output.html.print]
enable = true
page-break = true
EOF
```

---

### 步骤10：创建目录结构

根据您的书籍章节创建目录文件：

```bash
cat > src/SUMMARY.md << 'EOF'
# 《电商系统架构设计与实现》

[书籍介绍](README.md)
[前言](preface.md)

---

# 第一部分：架构方法论与设计原则

- [第1章 架构设计三位一体](part1/chapter1.md)
- [第2章 领域驱动设计战略篇](part1/chapter2.md)
- [第3章 整洁代码与设计模式](part1/chapter3.md)
- [第4章 架构质量保障](part1/chapter4.md)

---

# 第二部分：电商核心系统设计

## Part A：全局架构

- [第5章 电商系统全景图](part2/overview/chapter5.md)
- [第6章 系统集成与一致性设计](part2/overview/chapter6.md)

## Part B：商品供给与运营

- [第7章 商品中心系统](part2/supply/chapter7.md)
- [第8章 库存系统](part2/supply/chapter8.md)
- [第9章 营销系统](part2/supply/chapter9.md)
- [第10章 商品供给与运营管理](part2/supply/chapter10.md)

## Part C：交易链路

- [第11章 计价系统设计与实现](part2/transaction/chapter11.md)
- [第12章 搜索与导购](part2/transaction/chapter12.md)
- [第13章 购物车与结算](part2/transaction/chapter13.md)
- [第14章 订单系统](part2/transaction/chapter14.md)
- [第15章 支付系统](part2/transaction/chapter15.md)

---

# 第三部分：综合案例与落地

- [第16章 B2B2C平台完整架构](part3/chapter16.md)

---

# 附录

- [附录A 技术栈选型指南](appendix/tech-stack.md)
- [附录B 面试题精选](appendix/interview.md)
- [附录C 系统集成模式速查表](appendix/integration.md)
- [附录D 术语表](appendix/glossary.md)
- [附录E 参考资料](appendix/references.md)

---

[后记](postscript.md)
EOF
```

---

### 步骤11：创建目录结构

```bash
# 创建所有需要的目录
mkdir -p src/part1
mkdir -p src/part2/overview
mkdir -p src/part2/supply
mkdir -p src/part2/transaction
mkdir -p src/part3
mkdir -p src/appendix
```

---

### 步骤12：编写内容迁移脚本

创建自动化迁移脚本：

```bash
cat > migrate.sh << 'EOF'
#!/bin/bash

# 从Hexo的source/book迁移到mdBook的src

SOURCE_DIR="../source/book"
TARGET_DIR="./src"

echo "开始迁移内容..."

# 函数：处理单个文件
process_file() {
    local src_file=$1
    local dest_file=$2
    
    echo "处理: $src_file -> $dest_file"
    
    # 移除Hexo Front Matter并处理内容
    sed '1{/^---$/!b;:a;n;/^---$/!ba;d}' "$src_file" | \
    sed '/^\*\*导航\*\*:/d' | \
    sed 's/\.html)/.md)/g' > "$dest_file"
}

# 迁移首页
if [ -f "$SOURCE_DIR/index.md" ]; then
    process_file "$SOURCE_DIR/index.md" "$TARGET_DIR/README.md"
fi

# 迁移第一部分（第1-4章）
for i in {1..4}; do
    if [ -f "$SOURCE_DIR/chapter$i.md" ]; then
        process_file "$SOURCE_DIR/chapter$i.md" "$TARGET_DIR/part1/chapter$i.md"
    fi
done

# 迁移第二部分 - 全局架构
for i in 5 6; do
    if [ -f "$SOURCE_DIR/chapter$i.md" ]; then
        process_file "$SOURCE_DIR/chapter$i.md" "$TARGET_DIR/part2/overview/chapter$i.md"
    fi
done

# 迁移第二部分 - 商品供给
for i in {7..10}; do
    if [ -f "$SOURCE_DIR/chapter$i.md" ]; then
        process_file "$SOURCE_DIR/chapter$i.md" "$TARGET_DIR/part2/supply/chapter$i.md"
    fi
done

# 迁移第二部分 - 交易链路
for i in {11..15}; do
    if [ -f "$SOURCE_DIR/chapter$i.md" ]; then
        process_file "$SOURCE_DIR/chapter$i.md" "$TARGET_DIR/part2/transaction/chapter$i.md"
    fi
done

# 迁移第三部分
if [ -f "$SOURCE_DIR/chapter16.md" ]; then
    process_file "$SOURCE_DIR/chapter16.md" "$TARGET_DIR/part3/chapter16.md"
fi

# 创建占位文件（尚未完成的章节）
touch "$TARGET_DIR/preface.md"
touch "$TARGET_DIR/appendix/tech-stack.md"
touch "$TARGET_DIR/appendix/interview.md"
touch "$TARGET_DIR/appendix/integration.md"
touch "$TARGET_DIR/appendix/glossary.md"
touch "$TARGET_DIR/appendix/references.md"
touch "$TARGET_DIR/postscript.md"

echo "✅ 迁移完成！"
echo "请检查 $TARGET_DIR 目录"
EOF

# 添加执行权限
chmod +x migrate.sh
```

---

### 步骤13：执行迁移

```bash
# 运行迁移脚本
./migrate.sh
```

**预期输出**：
```
开始迁移内容...
处理: ../source/book/index.md -> ./src/README.md
处理: ../source/book/chapter1.md -> ./src/part1/chapter1.md
...
✅ 迁移完成！
```

```bash
# 检查迁移结果
ls -la src/part1/
ls -la src/part2/overview/
ls -la src/part2/supply/
ls -la src/part2/transaction/
ls -la src/part3/
```

---

### 步骤14：测试本地预览

```bash
# 启动服务器
mdbook serve --open
```

**检查项目**：
- [ ] 左侧目录是否正确显示所有章节
- [ ] 点击章节是否能正确跳转
- [ ] 内容是否正确渲染
- [ ] 代码块是否有语法高亮
- [ ] 搜索功能是否可用

**如果发现问题**：
1. 图片不显示 → 检查图片路径
2. 链接失效 → 检查链接格式（应该是`.md`而不是`.html`）
3. 格式错乱 → 检查Markdown语法

---

### 步骤15：自定义样式

创建自定义CSS以优化中文阅读体验：

```bash
# 创建theme目录
mkdir -p theme

# 创建自定义样式
cat > theme/custom.css << 'EOF'
/* 中文字体优化 */
:root {
    --content-max-width: 900px;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", 
                 "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", 
                 "Helvetica Neue", Helvetica, Arial, sans-serif;
    font-size: 16px;
    line-height: 1.8;
}

/* 代码字体 */
code, pre {
    font-family: "SF Mono", Monaco, "Cascadia Code", "Roboto Mono", 
                 Consolas, "Courier New", monospace !important;
}

/* 标题样式 */
h1, h2, h3, h4, h5, h6 {
    font-weight: 600;
    margin-top: 1.5em;
    margin-bottom: 0.8em;
}

h1 {
    font-size: 2em;
    border-bottom: 2px solid #e0e0e0;
    padding-bottom: 0.3em;
}

h2 {
    font-size: 1.5em;
    border-bottom: 1px solid #e0e0e0;
    padding-bottom: 0.3em;
}

h3 {
    font-size: 1.25em;
}

/* 代码块优化 */
.hljs {
    border-radius: 6px;
    padding: 1em;
    overflow-x: auto;
}

/* 引用块样式 */
blockquote {
    border-left: 4px solid #3498db;
    background: #f8f9fa;
    padding: 10px 20px;
    margin: 1.5em 0;
    border-radius: 4px;
}

blockquote p {
    margin: 0;
}

/* 表格样式 */
table {
    border-collapse: collapse;
    width: 100%;
    margin: 1.5em 0;
}

table th {
    background: #f1f3f5;
    font-weight: 600;
}

table th, table td {
    border: 1px solid #dee2e6;
    padding: 10px 15px;
    text-align: left;
}

table tr:hover {
    background: #f8f9fa;
}

/* 内联代码 */
:not(pre) > code {
    background: #f5f7f9;
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 0.9em;
}

/* 目录优化 */
.chapter li.chapter-item {
    line-height: 2;
}

/* 打印优化 */
@media print {
    .nav-chapters {
        display: none;
    }
    
    h1, h2, h3 {
        page-break-after: avoid;
    }
}
EOF
```

**在book.toml中引用样式**：

```bash
# 编辑book.toml，在最后添加
cat >> book.toml << 'EOF'

[output.html]
additional-css = ["theme/custom.css"]
EOF
```

**重新加载查看效果**：
```bash
# 如果服务器还在运行，它会自动重新加载
# 否则重新启动
mdbook serve --open
```

---

## 第4部分：PDF生成（预计30分钟）

### 步骤16：配置PDF导出

```bash
# 在book.toml中添加PDF配置
cat >> book.toml << 'EOF'

[output.pdf]
paper-size = "a4"
margin-top = "20mm"
margin-bottom = "20mm"
margin-left = "25mm"
margin-right = "25mm"
EOF
```

### 步骤17：生成PDF

```bash
# 构建HTML版本
mdbook build

# 使用mdbook-pdf生成PDF
mdbook-pdf

# 或者手动指定输出
mdbook-pdf --standalone --output book/ecommerce-architecture.pdf
```

**生成的PDF位置**：
```
book/ecommerce-architecture.pdf
```

```bash
# 打开PDF查看
open book/ecommerce-architecture.pdf
```

---

## 第5部分：部署到GitHub Pages（预计1小时）

### 步骤18：准备Git仓库

```bash
# 1. 初始化Git（如果还没有）
git init

# 2. 创建.gitignore
cat > .gitignore << 'EOF'
/book/
*.swp
*~
.DS_Store
EOF

# 3. 提交代码
git add .
git commit -m "Initial commit: mdBook project setup"
```

---

### 步骤19：创建GitHub仓库

**在GitHub网站上操作**：

1. 访问 https://github.com/new
2. 填写信息：
   - Repository name: `ecommerce-book`
   - Description: `《电商系统架构设计与实现》电子书`
   - Public（公开）
   - 不要勾选任何初始化选项
3. 点击 "Create repository"

**在终端关联远程仓库**：

```bash
# 使用GitHub给出的命令
git remote add origin https://github.com/wxquare/ecommerce-book.git
git branch -M main
git push -u origin main
```

---

### 步骤20：配置GitHub Actions

创建自动部署配置：

```bash
# 1. 创建目录
mkdir -p .github/workflows

# 2. 创建workflow文件
cat > .github/workflows/deploy.yml << 'EOF'
name: Deploy mdBook

on:
  push:
    branches: [ main ]
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup mdBook
      uses: peaceiris/actions-mdbook@v1
      with:
        mdbook-version: 'latest'
    
    - name: Build book
      run: mdbook build
    
    - name: Upload artifact
      uses: actions/upload-pages-artifact@v3
      with:
        path: ./book

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    
    steps:
    - name: Deploy to GitHub Pages
      id: deployment
      uses: actions/deploy-pages@v4
EOF

# 3. 提交并推送
git add .github/workflows/deploy.yml
git commit -m "Add GitHub Actions workflow"
git push
```

---

### 步骤21：启用GitHub Pages

**在GitHub网站上操作**：

1. 进入您的仓库 `https://github.com/wxquare/ecommerce-book`
2. 点击 **Settings** 标签
3. 左侧菜单点击 **Pages**
4. 在 "Source" 下拉菜单中选择 **GitHub Actions**
5. 等待几秒，页面会显示：
   ```
   Your site is live at https://wxquare.github.io/ecommerce-book/
   ```

---

### 步骤22：验证部署

```bash
# 等待2-3分钟，然后访问
open https://wxquare.github.io/ecommerce-book/
```

**检查项**：
- [ ] 网站可以正常访问
- [ ] 所有章节都显示正常
- [ ] 搜索功能可用
- [ ] 样式正确应用

---

## 第6部分：日常维护

### 更新内容流程

```bash
# 1. 编辑Markdown文件
vim src/part1/chapter1.md

# 2. 本地预览
mdbook serve

# 3. 确认无误后提交
git add .
git commit -m "Update chapter1"
git push

# 4. GitHub Actions自动部署（2-3分钟后生效）
```

### 添加新章节

```bash
# 1. 创建新文件
echo "# 新章节内容" > src/part3/chapter17.md

# 2. 在SUMMARY.md中注册
vim src/SUMMARY.md
# 添加：- [第17章 系统演进与重构](part3/chapter17.md)

# 3. 测试
mdbook serve

# 4. 提交
git add .
git commit -m "Add chapter17"
git push
```

---

## 第7部分：常见问题

### Q1: 安装Rust失败

**症状**：`curl: (7) Failed to connect`

**解决**：
```bash
# 使用镜像源
export RUSTUP_DIST_SERVER=https://mirrors.ustc.edu.cn/rust-static
export RUSTUP_UPDATE_ROOT=https://mirrors.ustc.edu.cn/rust-static/rustup
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

---

### Q2: mdbook serve端口被占用

**症状**：`Error: Address already in use (os error 48)`

**解决**：
```bash
# 指定其他端口
mdbook serve --port 4000
```

---

### Q3: 中文搜索不准确

**解决**：在book.toml中调整搜索配置：
```toml
[output.html.search]
enable = true
use-boolean-and = false  # 改为OR搜索
boost-title = 3
```

---

### Q4: PDF生成失败

**症状**：`Chrome binary not found`

**解决**：
```bash
# mdbook-pdf依赖Chrome，确保系统已安装Chrome
# macOS会自动检测 /Applications/Google Chrome.app
```

---

### Q5: 图片不显示

**原因**：图片路径不正确

**解决**：
```bash
# 图片应放在src目录下
mkdir -p src/images
cp ../source/book/images/* src/images/

# Markdown中引用
![架构图](../images/architecture.png)
```

---

### Q6: 部署后404

**检查**：
1. GitHub Pages是否启用
2. book.toml中的site-url是否正确
   ```toml
   site-url = "/ecommerce-book/"  # 注意斜杠
   ```

---

## 第8部分：进阶技巧

### 技巧1：自定义域名

```bash
# 1. 在book.toml中添加
[output.html]
cname = "book.wxquare.com"

# 2. 在DNS提供商添加CNAME记录
# book.wxquare.com → wxquare.github.io
```

---

### 技巧2：添加Google Analytics

```bash
# 在book.toml中添加
[output.html]
google-analytics = "G-XXXXXXXXXX"
```

---

### 技巧3：自定义404页面

```bash
# 创建404页面
cat > src/404.md << 'EOF'
# 页面未找到

您访问的页面不存在。

[返回首页](/)
EOF

# 在SUMMARY.md中添加（隐藏）
# [404](404.md)
```

---

## 🎉 完成清单

恭喜！如果您完成了所有步骤，现在您应该有：

- [x] 本地mdBook开发环境
- [x] 完整的书籍项目结构
- [x] 迁移完成的所有章节内容
- [x] 自定义样式和主题
- [x] 在线阅读网站（GitHub Pages）
- [x] 自动化部署流程
- [x] PDF导出功能

---

## 📚 快速命令参考

```bash
# 开发
mdbook serve              # 启动开发服务器
mdbook serve --port 4000  # 指定端口
mdbook serve --open       # 启动并打开浏览器

# 构建
mdbook build              # 构建HTML
mdbook build --open       # 构建并打开
mdbook clean              # 清理构建产物

# 测试
mdbook test               # 测试代码示例

# 部署
git add .
git commit -m "Update content"
git push                  # 自动触发部署
```

---

## 🔗 有用的资源

- mdBook官方文档：https://rust-lang.github.io/mdBook/
- Rust安装指南：https://www.rust-lang.org/tools/install
- Markdown语法：https://www.markdownguide.org/
- GitHub Pages文档：https://docs.github.com/en/pages

---

**编写时间**：2026-04-17  
**适用版本**：mdBook v0.4.37+  
**测试环境**：macOS 14.x
