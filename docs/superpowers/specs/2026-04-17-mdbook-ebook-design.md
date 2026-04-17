# mdBook电子书展示方案设计文档

**项目**：《电商系统架构设计与实现》电子书展示  
**日期**：2026-04-17  
**作者**：AI Assistant  
**状态**：待审核

---

## 1. 背景与目标

### 1.1 项目背景

用户已完成一本技术专著《电商系统架构设计与实现》的内容创作，包含：
- 18个章节（已完成16章）
- 约150,000字
- 5个附录
- 存储在Hexo博客的 `source/book/` 目录

### 1.2 目标

需要为这本书提供专业的展示方案，满足：
1. **在线阅读**：提供类似专业电子书的阅读体验
2. **PDF下载**：生成高质量PDF供离线阅读
3. **个人品牌传播**：主要通过个人渠道（博客、GitHub）传播
4. **长期维护**：维护成本低，易于更新内容

### 1.3 技术选型

选择 **mdBook** 作为实施方案，理由：
- 专业的书籍阅读体验（左侧目录树、右侧大纲、全文搜索）
- 维护简单（Markdown + 单一配置文件）
- 成熟稳定（Rust官方文档使用）
- 完全免费开源
- 支持多格式导出（HTML、PDF）

---

## 2. 整体架构设计

### 2.1 项目结构

```
ecommerce-book/                    # 新建独立仓库
├── book.toml                      # mdBook配置文件
├── src/                           # 书籍源文件
│   ├── SUMMARY.md                 # 目录结构（自动生成导航）
│   ├── README.md                  # 首页
│   ├── preface.md                 # 前言
│   ├── part1/                     # 第一部分（4章）
│   ├── part2/                     # 第二部分（11章）
│   │   ├── overview/              # 全局架构（2章）
│   │   ├── supply/                # 商品供给（4章）
│   │   └── transaction/           # 交易链路（5章）
│   ├── part3/                     # 第三部分（3章）
│   └── appendix/                  # 附录（5个）
├── theme/                         # 自定义主题
│   ├── custom.css                 # 样式定制
│   └── book.js                    # 自定义脚本
├── .github/workflows/
│   └── deploy.yml                 # GitHub Actions自动部署
└── scripts/
    ├── migrate.sh                 # 内容迁移脚本
    └── build-pdf.sh               # PDF构建脚本
```

### 2.2 部署架构

```
┌─────────────────────────────────────────────────────┐
│  GitHub Repository: ecommerce-book                  │
│                                                       │
│  main branch                                         │
│  ├── src/ (Markdown源文件)                          │
│  ├── book.toml                                      │
│  └── theme/                                         │
│                                                       │
│         ↓ git push                                  │
│         ↓                                           │
│  ┌─────────────────────────────┐                   │
│  │  GitHub Actions             │                   │
│  │  1. mdbook build            │                   │
│  │  2. mdbook-pdf              │                   │
│  │  3. 部署到gh-pages分支      │                   │
│  └─────────────────────────────┘                   │
│         ↓                                           │
│  gh-pages branch                                    │
│  └── book/ (生成的HTML + PDF)                      │
└─────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────┐
│  GitHub Pages                                        │
│  https://wxquare.github.io/ecommerce-book/          │
│  或自定义域名: https://book.wxquare.com              │
└─────────────────────────────────────────────────────┘
```

### 2.3 用户访问流程

```
用户访问 book.wxquare.com
    ↓
加载HTML书籍网站
    ├── 在线阅读（mdBook生成的HTML）
    │   ├── 左侧：目录树导航
    │   ├── 中间：章节内容
    │   ├── 右侧：章节大纲
    │   └── 顶部：搜索、主题切换、PDF下载
    │
    └── 下载PDF（点击下载按钮）
        └── ecommerce-architecture.pdf
```

---

## 3. 核心功能设计

### 3.1 阅读体验功能

#### 3.1.1 导航系统
- **左侧目录树**：
  - 自动从 `SUMMARY.md` 生成
  - 支持折叠/展开章节
  - 高亮当前阅读位置
  - 支持键盘导航（上下箭头）

- **右侧大纲**：
  - 显示当前章节的二级、三级标题
  - 点击快速跳转
  - 滚动时自动高亮

- **章节切换**：
  - 底部"上一章/下一章"按钮
  - 支持键盘快捷键（← →）

#### 3.1.2 搜索功能
- 全文搜索（支持中文分词）
- 实时搜索结果预览
- 高亮搜索关键词
- 支持正则表达式（可选）

#### 3.1.3 主题切换
- 明亮主题（默认）
- 暗色主题
- 护眼主题
- 用户偏好记忆（localStorage）

#### 3.1.4 代码高亮
- 支持多语言语法高亮
- 代码块一键复制
- 行号显示（可配置）

### 3.2 PDF生成功能

#### 3.2.1 生成方式
使用 `mdbook-pdf` 插件，基于 Chrome Headless 渲染

#### 3.2.2 PDF特性
- 完整的书籍内容
- 自动生成目录（带页码）
- 页眉页脚（章节标题 + 页码）
- 打印优化样式
- 嵌入字体（支持中文）

#### 3.2.3 文件输出
- 文件名：`ecommerce-architecture.pdf`
- 大小：预计 10-15MB
- 页数：预计 400-500页（A4纸）

### 3.3 自定义样式

#### 3.3.1 中文优化
```css
/* 字体栈优化 */
body {
  font-family: -apple-system, "PingFang SC", "Hiragino Sans GB", 
               "Microsoft YaHei", sans-serif;
  font-size: 16px;
  line-height: 1.8;
}

/* 标题样式 */
h1, h2 {
  border-bottom: 1px solid #eee;
  padding-bottom: 0.3em;
}
```

#### 3.3.2 代码块优化
- GitHub风格高亮
- 圆角边框
- 复制按钮

#### 3.3.3 响应式设计
- 桌面端：三栏布局（目录 + 内容 + 大纲）
- 平板端：两栏布局（目录可收起）
- 手机端：单栏布局（目录抽屉式）

---

## 4. 技术实现方案

### 4.1 工具链

#### 4.1.1 核心工具
- **mdBook** (最新版)：核心构建工具
- **Rust/Cargo**：mdBook的依赖

#### 4.1.2 可选插件
- **mdbook-pdf**：PDF生成
- **mdbook-mermaid**：Mermaid图表支持
- **mdbook-toc**：自动生成章节内目录

### 4.2 配置文件

#### 4.2.1 book.toml（核心配置）
```toml
[book]
title = "电商系统架构设计与实现"
authors = ["wxquare"]
language = "zh-CN"
src = "src"

[output.html]
default-theme = "light"
git-repository-url = "https://github.com/wxquare/ecommerce-book"
site-url = "/ecommerce-book/"

[output.html.search]
enable = true
limit-results = 20

[output.pdf]
enable = true
paper-size = "a4"
```

#### 4.2.2 SUMMARY.md（目录结构）
mdBook最重要的文件，定义整本书的结构：
```markdown
# 目录

[书籍介绍](README.md)

# 第一部分：架构方法论

- [第1章 架构设计三位一体](part1/chapter1.md)
  - [1.1 引言](part1/chapter1.md#11-引言)
  - [1.2 Clean Architecture](part1/chapter1.md#12-clean-architecture)
  ...
```

### 4.3 内容迁移策略

#### 4.3.1 自动化迁移
使用脚本处理：
1. 移除Hexo的Front Matter
2. 移除导航链接（mdBook自动生成）
3. 调整内部链接格式（`.html` → `.md`）
4. 重组目录结构

#### 4.3.2 手动调整
需要人工检查的内容：
1. 图片路径
2. 复杂的表格
3. Mermaid图表
4. 特殊格式的代码块

### 4.4 自动化部署

#### 4.4.1 GitHub Actions工作流
```yaml
on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup mdBook
        uses: peaceiris/actions-mdbook@v1
      - name: Build
        run: mdbook build
      - name: Deploy
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./book
```

#### 4.4.2 部署触发条件
- 推送到main分支：自动触发构建和部署
- Pull Request：只构建，不部署（用于预览）
- 手动触发：通过GitHub界面手动运行

---

## 5. 实施计划

### 5.1 阶段划分

#### 阶段1：环境准备（1天）
- [ ] 安装Rust和Cargo
- [ ] 安装mdBook及插件
- [ ] 验证工具链

#### 阶段2：项目初始化（0.5天）
- [ ] 创建项目结构
- [ ] 配置book.toml
- [ ] 编写SUMMARY.md
- [ ] 本地预览测试

#### 阶段3：内容迁移（1-2天）
- [ ] 编写迁移脚本
- [ ] 运行自动迁移
- [ ] 手动调整内容
- [ ] 验证所有链接

#### 阶段4：样式定制（1-2天）
- [ ] 创建custom.css
- [ ] 优化中文排版
- [ ] 测试响应式布局
- [ ] 调整代码高亮

#### 阶段5：PDF生成（1天）
- [ ] 配置mdbook-pdf
- [ ] 优化打印样式
- [ ] 测试PDF生成
- [ ] 调整页面布局

#### 阶段6：自动化部署（1天）
- [ ] 配置GitHub Actions
- [ ] 设置GitHub Pages
- [ ] 配置自定义域名（可选）
- [ ] 测试完整部署流程

#### 阶段7：测试和优化（1天）
- [ ] 功能测试（搜索、导航、主题）
- [ ] 兼容性测试（浏览器、设备）
- [ ] 性能优化
- [ ] 文档完善

### 5.2 时间估算

**总计**：5-7个工作日

**每日任务分配**：
- Day 1: 阶段1 + 阶段2
- Day 2: 阶段3
- Day 3: 阶段4
- Day 4: 阶段5
- Day 5: 阶段6
- Day 6: 阶段7
- Day 7: 缓冲时间

---

## 6. 风险与应对

### 6.1 技术风险

| 风险 | 概率 | 影响 | 应对措施 |
|------|------|------|----------|
| Rust安装失败 | 低 | 中 | 提供多种安装方式（官方脚本、包管理器） |
| mdBook插件不兼容 | 低 | 中 | 使用稳定版本，提供回退方案 |
| PDF生成失败 | 中 | 中 | 使用备选方案（Pandoc） |
| 中文搜索不准确 | 中 | 低 | 调整搜索配置，必要时使用外部搜索 |

### 6.2 内容风险

| 风险 | 概率 | 影响 | 应对措施 |
|------|------|------|----------|
| 图片路径错误 | 高 | 中 | 迁移脚本验证，手动检查 |
| 内部链接失效 | 中 | 中 | 使用mdBook的link-check |
| Markdown格式不兼容 | 低 | 低 | 逐章测试渲染 |

### 6.3 运维风险

| 风险 | 概率 | 影响 | 应对措施 |
|------|------|------|----------|
| GitHub Actions配额耗尽 | 低 | 低 | 公开仓库有免费配额 |
| 部署失败 | 低 | 中 | 保留本地构建，手动部署 |
| 自定义域名配置错误 | 低 | 低 | 提供详细文档，分步验证 |

---

## 7. 成功标准

### 7.1 功能完整性
- [ ] 所有章节正确显示
- [ ] 目录导航功能正常
- [ ] 搜索功能可用
- [ ] PDF可正常下载
- [ ] 代码高亮正确
- [ ] 响应式布局正常

### 7.2 用户体验
- [ ] 页面加载速度 < 2秒
- [ ] 手机端阅读体验良好
- [ ] 搜索响应及时（< 500ms）
- [ ] 无明显UI bug

### 7.3 内容完整性
- [ ] 所有图片正常显示
- [ ] 内部链接无失效
- [ ] 代码块格式正确
- [ ] 表格渲染正常
- [ ] Mermaid图表可见

### 7.4 运维稳定性
- [ ] 自动部署成功率 > 95%
- [ ] 无部署错误残留
- [ ] 构建时间 < 5分钟

---

## 8. 后续优化方向

### 8.1 功能增强
1. **多语言支持**：未来可以提供英文版
2. **评论系统**：集成Giscus或Utterances
3. **版本管理**：支持多版本文档切换
4. **统计分析**：集成Google Analytics或Plausible

### 8.2 内容增强
1. **交互式示例**：嵌入代码演示（CodePen/JSFiddle）
2. **视频教程**：补充关键章节的视频讲解
3. **练习题**：每章末尾添加思考题
4. **勘误系统**：方便读者反馈问题

### 8.3 分发渠道
1. **ePub格式**：通过Pandoc生成ePub
2. **微信读书**：转换为微信读书格式
3. **Kindle**：生成mobi格式
4. **纸质出版**：提供LaTeX版本用于正式出版

---

## 9. 维护计划

### 9.1 日常维护
- **内容更新**：直接编辑Markdown，提交即自动部署
- **Bug修复**：通过GitHub Issues跟踪和修复
- **版本管理**：使用Git标签标记重要版本

### 9.2 定期检查（每季度）
- [ ] 检查所有外部链接有效性
- [ ] 更新依赖版本（mdBook、插件）
- [ ] 审查用户反馈
- [ ] 优化性能和用户体验

### 9.3 重大更新
- 新增章节：更新SUMMARY.md
- 结构调整：重新组织目录
- 样式改版：更新theme目录

---

## 10. 参考案例

### 10.1 成功案例

**The Rust Programming Language**
- URL: https://doc.rust-lang.org/book/
- 特点：官方权威、结构清晰、搜索强大
- 借鉴：目录组织、样式设计

**Comprehensive Rust**
- URL: https://google.github.io/comprehensive-rust/
- 特点：Google出品、练习题丰富、交互性强
- 借鉴：页面布局、互动元素

**Command Line Rust**
- URL: https://www.oreilly.com/library/view/command-line-rust/9781098109424/
- 特点：O'Reilly书籍、代码示例丰富
- 借鉴：代码组织方式

### 10.2 技术文档

- mdBook官方文档：https://rust-lang.github.io/mdBook/
- mdBook用户指南：https://rust-lang.github.io/mdBook/guide/
- GitHub Pages文档：https://docs.github.com/en/pages

---

## 11. 附录

### 11.1 命令速查表

```bash
# 安装
cargo install mdbook
cargo install mdbook-pdf

# 创建项目
mdbook init my-book

# 本地开发
mdbook serve --open          # 启动开发服务器
mdbook serve --port 4000     # 指定端口

# 构建
mdbook build                 # 构建HTML
mdbook build --open          # 构建后打开
mdbook build --dest-dir dist # 指定输出目录

# 测试
mdbook test                  # 测试代码示例

# 清理
mdbook clean                 # 清理构建产物
```

### 11.2 目录结构示例

```
src/
├── SUMMARY.md              # 必需：定义书籍结构
├── README.md               # 可选：首页内容
├── chapter1.md
├── chapter2.md
└── images/                 # 图片资源
    └── architecture.png
```

### 11.3 SUMMARY.md语法

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

## 12. 决策记录

### ADR-001: 选择mdBook而非GitBook
**日期**：2026-04-17  
**状态**：已接受  
**理由**：
- GitBook Legacy已停止维护
- GitBook SaaS需要付费且有锁定风险
- mdBook维护活跃、完全免费、阅读体验接近GitBook

### ADR-002: 使用GitHub Pages部署
**日期**：2026-04-17  
**状态**：已接受  
**理由**：
- 免费托管
- 与GitHub仓库天然集成
- 支持自定义域名
- HTTPS默认开启

### ADR-003: 独立仓库而非Hexo集成
**日期**：2026-04-17  
**状态**：已接受  
**理由**：
- 书籍和博客关注点不同
- 独立维护更清晰
- 避免Hexo配置冲突
- 便于未来迁移或重构

---

**文档版本**：v1.0  
**最后更新**：2026-04-17  
**下次审查**：实施完成后
