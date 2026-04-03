# AI辅助博客维护系统使用指南

本文档介绍如何使用AI（Claude Code）来辅助维护这个Hexo博客项目。

## 📁 项目结构

```
.
├── CLAUDE.md                    # 项目级记忆系统
├── .claude/
│   ├── settings.json           # Hooks和权限配置
│   ├── skills/                 # 可复用工作流
│   │   ├── new-post/          # 创建新文章
│   │   ├── review-post/       # 文章审查
│   │   ├── organize-posts/    # 文章整理
│   │   ├── generate-summary/  # 生成摘要
│   │   └── link-check/        # 链接检查
│   ├── commands/              # 快捷命令
│   │   ├── publish.md         # 一键发布
│   │   └── stats.md           # 统计报告
│   ├── agents/                # 专业助手
│   │   ├── tech-reviewer.md   # 技术审查
│   │   └── seo-optimizer.md   # SEO优化
│   └── templates/             # 文章模板
│       ├── tech-tutorial.md   # 技术教程模板
│       ├── system-design.md   # 系统设计模板
│       └── interview-prep.md  # 面试准备模板
├── scripts/
│   └── pre-commit-check.sh    # 提交前检查脚本
└── .mcp.json                  # MCP配置文件
```

## 🚀 快速开始

### 1. 创建新文章

```bash
# 使用Skill创建新文章
/new-post
```

AI会询问：
- 文章标题
- 分类（AI/system-design/other）
- 标签
- 是否需要子分类

然后自动创建文件，包含完整的Front Matter和基础结构。

### 2. 审查文章

```bash
# 审查文章质量
/review-post source/_posts/AI/your-article.md
```

AI会检查：
- Front Matter完整性
- 内容结构
- 代码质量
- 格式规范
- 技术准确性
- SEO优化
- 可读性
- 链接有效性

### 3. 发布文章

```bash
# 一键发布流程
/publish
```

AI会自动执行：
1. 检查Git状态
2. 运行构建测试
3. 生成提交信息
4. 提交变更
5. 推送到远程
6. 部署到GitHub Pages

### 4. 查看统计

```bash
# 生成博客统计报告
/stats
```

获得详细的统计信息：
- 文章总数和分类分布
- 热门标签
- 更新活跃度
- 文章长度分析
- 需要更新的文章

## 🎯 核心功能

### Skills（可复用工作流）

#### `/new-post` - 创建新文章
- 智能询问文章信息
- 自动生成文件名
- 创建标准化的Front Matter
- 提供文章模板

#### `/review-post` - 文章审查
- 8大维度全面检查
- 生成详细审查报告
- 提供具体改进建议
- 评分和优先级排序

#### `/organize-posts` - 文章整理
- 扫描所有文章结构
- 识别分类不合理的文章
- 建议合并或拆分
- 生成文章索引

#### `/generate-summary` - 生成摘要
- 三种长度的摘要
- SEO关键词建议
- 社交媒体文案
- 内部链接建议

#### `/link-check` - 链接检查
- 检查内部链接
- 检查外部链接
- 检查图片链接
- 生成失效链接报告

### Commands（快捷命令）

#### `/publish` - 一键发布
完整的发布流程，从检查到部署，一个命令搞定。

#### `/stats` - 统计报告
详细的博客统计分析，帮助了解内容状况。

### Agents（专业助手）

#### Technical Reviewer（技术审查专家）
- 只读权限
- 使用Opus 4.6模型
- 专注技术准确性
- 提供专业建议

使用方法：
```
请Technical Reviewer审查这篇文章
```

#### SEO Optimizer（SEO优化专家）
- 可以编辑文件
- 关键词优化
- 标题优化
- 内部链接建设

使用方法：
```
请SEO Optimizer优化这篇文章
```

### Templates（文章模板）

#### tech-tutorial.md - 技术教程模板
包含：
- 引言和目标读者
- 环境准备
- 核心概念
- 快速开始
- 实战案例
- 常见问题
- 性能优化

#### system-design.md - 系统设计模板
包含：
- 需求分析
- 容量估算
- 系统架构
- 数据库设计
- API设计
- 高可用设计
- 性能优化
- 成本估算

#### interview-prep.md - 面试准备模板
包含：
- 知识体系
- 高频面试题
- 手写代码题
- 系统设计题
- 行为面试题
- 准备清单

## 🔧 Hooks（自动化检查）

### PostToolUse Hook
每次编辑Markdown文件后自动提示。

### PreToolUse Hook
Git提交前自动运行检查脚本：
- Front Matter完整性
- 代码块语言标注
- 中英文空格
- Hexo构建测试

## 🔌 MCP集成

### GitHub MCP
- 自动创建PR
- 管理Issues
- 查看仓库信息
- 自动打标签

**设置方法：**
```bash
export GITHUB_TOKEN=your_github_token
```

### Filesystem MCP
- 管理图片文件
- 管理Excalidraw图表
- 批量处理文件

## 📝 使用示例

### 场景1：写新文章

```bash
# 1. 创建文章
/new-post
# 输入：标题、分类、标签

# 2. 编写内容
# （使用模板作为参考）

# 3. 审查文章
/review-post source/_posts/AI/new-article.md

# 4. 修改问题
# （根据审查建议修改）

# 5. 发布
/publish
```

### 场景2：整理旧文章

```bash
# 1. 生成统计报告
/stats

# 2. 整理文章结构
/organize-posts

# 3. 检查链接
/link-check

# 4. 修复问题
# （根据报告修复）

# 5. 提交更新
/publish
```

### 场景3：优化SEO

```bash
# 1. 让SEO专家审查
请SEO Optimizer分析 source/_posts/AI/article.md

# 2. 生成摘要
/generate-summary source/_posts/AI/article.md

# 3. 应用优化建议
# （根据建议修改）

# 4. 发布更新
/publish
```

## 💡 最佳实践

### 1. 定期维护
- 每周运行`/stats`了解博客状况
- 每月运行`/organize-posts`整理结构
- 每季度运行`/link-check`检查链接

### 2. 写作流程
- 使用`/new-post`创建文章
- 参考模板编写内容
- 使用`/review-post`审查
- 使用Technical Reviewer深度审查
- 使用SEO Optimizer优化
- 使用`/publish`发布

### 3. 质量控制
- 每篇文章发布前必须运行`/review-post`
- 技术文章必须经过Technical Reviewer审查
- 重要文章必须经过SEO Optimizer优化

### 4. 协作方式
- CLAUDE.md和Skills配置提交到Git
- 团队成员共享相同的配置
- 定期更新CLAUDE.md（记录新的规则）

## 🎓 学习资源

### 了解Claude Code
- 阅读`source/_posts/AI/claude-code-guide-summary.md`
- 了解Skills、Hooks、MCP的概念
- 学习如何编写自定义Skills

### 扩展功能
- 创建自己的Skills
- 编写自定义Hooks
- 接入更多MCP服务

## 🐛 常见问题

### Q: 如何修改文章模板？
A: 编辑`.claude/templates/`目录下的模板文件。

### Q: 如何添加新的Skill？
A: 在`.claude/skills/`创建新目录，添加`SKILL.md`文件。

### Q: 提交前检查失败怎么办？
A: 查看`scripts/pre-commit-check.sh`的输出，根据提示修复问题。

### Q: 如何禁用某个Hook？
A: 编辑`.claude/settings.json`，删除或注释对应的Hook配置。

### Q: GitHub Token如何设置？
A: 在GitHub生成Personal Access Token，然后`export GITHUB_TOKEN=token`。

## 📚 参考资料

- [Hexo官方文档](https://hexo.io/docs/)
- [Claude Code文档](https://docs.anthropic.com/claude/docs)
- [MCP协议](https://modelcontextprotocol.io/)

---

**更新日志：**
- 2026-04-02：初始版本，完整的AI辅助系统搭建完成
