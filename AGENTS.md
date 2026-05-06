# wxquare.github.io - 个人技术博客

## 项目概述
- 基于Hexo 7.2.0的静态博客
- 主题：hexo-theme-next
- 部署：GitHub Pages (hexo-deployer-git)
- 支持Mermaid图表和PDF生成

## 目录结构
- `source/_posts/AI/` - AI相关文章（包含computer-vision, tensorflow, tvm子目录）
- `source/_posts/system-design/` - 系统设计文章
- `source/_posts/other/` - 其他主题文章
- `source/about/` - 关于页面和面试资料
- `source/diagrams/` - Excalidraw图表文件
- `books/ai-book/` - AI Agent 工程实践 mdBook
- `books/ecommerce-book/` - 电商系统架构设计 mdBook
- `books/system-design-book/` - 程序员系统设计与面试指南 mdBook
- `books/scripts/` - 三本 mdBook 共用的构建辅助脚本
- `resume/` - 简历相关文档

## 开发命令
- 本地预览：`npm run server`
- 生成静态文件：`npm run build`
- 清理缓存：`npm run clean`
- 部署到GitHub Pages：`npm run deploy`

## 写作规范
- 文章必须包含Front Matter（title, date, categories, tags）
- 日期格式：YYYY-MM-DD
- 分类层级：最多2层（如 AI/计算机视觉）
- 标签：用小写，多个词用连字符连接（如 deep-learning）
- 中文文章使用中文标点，英文文章使用英文标点
- 代码块必须指定语言（如 ```python, ```go）
- 中英文之间需要有空格

## 文章命名规范
- 重要系列文章：使用数字前缀（如 `22-ai-system-design.md`）
- 技术笔记：使用描述性名称（如 `tensorflow-model-quantization.md`）
- 日期开头：用于时效性强的文章（如 `2026-03-07-OpenClaw深度调研.md`）

## Hexo 配置参考
- 博客搭建与配置的完整记录：`source/_posts/other/基于Github双分支和Hexo搭建博客.md`
- 遇到 Hexo 配置相关问题（主题设置、插件安装、公式渲染等）时，优先查阅并更新该文章

## 常见陷阱
- 修改_config.yml后必须重启server
- 新增文章后需要运行`hexo clean`清理缓存
- 部署前先运行`npm run build`确保没有错误
- Front Matter的date字段必须是字符串格式，不能是对象
- 图片路径使用相对路径，不要使用绝对路径
- shell 正则中包含反引号时必须使用单引号或转义，避免被 shell 当作命令替换执行

## AI Agent 协作原则（Harness Engineering）
- **上下文精简**：本文件作为入口（<100 行），详细规范见 `.cursorrules`，按需查阅
- **构建即验证**：任何文章变更后必须运行 `npm run clean && npm run build`，构建失败则修复后再提交
- **上下文防火墙**：复杂任务先 Plan 再开新会话执行，避免长会话上下文退化
- **最小操作权限**：不要修改 themes/、db.json、.deploy_git、node_modules
- **错误即规则**：Agent 每犯一个新错误，将其补充到"常见陷阱"中，防止重犯
