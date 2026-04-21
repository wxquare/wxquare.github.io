# AI工程实践书籍设计文档

**项目**：《AI工程实践：从编程到Agent的完整指南》  
**日期**：2026-04-21  
**作者**：wxquare + AI Assistant  
**状态**：已批准，开始实施

---

## 1. 项目背景与目标

### 1.1 项目背景

用户已有丰富的AI工程实践博客内容（7篇，约16,700行），涵盖：
- AI编程实践（Vibe Coding、Spec Coding、Harness Engineering）
- Agent系统设计（架构、工具、协作、可观测性）
- 完整实战案例（DoD Agent、知识管理系统）

现需将这些内容系统化整理为一本电子书，形成完整的知识体系。

### 1.2 设计目标

**书名**：《AI工程实践：从编程到Agent的完整指南》

**核心定位**：
- **目标读者**：有经验的AI工程师
- **内容策略**：实战为主（70%有完整素材）+ 基础理论为辅（30%需补充）
- **版本目标**：v1.0结构完整版，每章包含理论→实战→最佳实践→陷阱
- **发布策略**：分阶段，v1.0骨架完整，后续版本逐步深化

**核心价值**：
1. 系统化的AI工程实践方法论（Vibe Coding → Spec Coding → Harness Engineering）
2. 生产级Agent系统设计全流程（从需求分析到部署运维）
3. 真实案例驱动（DoD Agent、知识管理Agent）
4. 工具链完整覆盖（Cursor、Claude Code、MCP、RAG）

### 1.3 设计原则

1. **实战优先**：有完整素材的章节优先，基础理论适度精简
2. **结构完整**：每章包含完整结构（理论→实战→最佳实践→陷阱）
3. **素材复用**：充分利用现有博客内容，避免重复劳动
4. **渐进增强**：v1.0完成骨架，后续版本深化重点章节
5. **工程化规范**：遵循Hexo博客项目的写作规范和构建流程

---

## 2. 整体架构设计

### 2.1 目录结构（11章，4部分）

```
第一部分：AI编程实战（3章）
├── 第1章 从Vibe Coding到Spec Coding：AI编程范式演进
│   ├── 1.1 AI编程工具的三次演进
│   ├── 1.2 Vibe Coding的本质与陷阱
│   ├── 1.3 Spec Coding：规范驱动的工程化方法
│   ├── 1.4 编写高质量Spec的完整指南
│   ├── 1.5 Cursor IDE完整实践
│   └── 1.6 Claude Code工作流与最佳实践
│
├── 第2章 Claude Code：终端原生的AI Agent
│   ├── 2.1 Claude Code的革命性变化
│   ├── 2.2 进阶对话技巧：让AI真正理解你
│   ├── 2.3 Plan模式：先想清楚再动手
│   ├── 2.4 Auto模式与安全防护
│   ├── 2.5 CLAUDE.md：Agent的宪法
│   ├── 2.6 会话管理与上下文优化
│   └── 2.7 完整工作流案例
│
└── 第3章 Harness Engineering：驾驭AI的基础设施
    ├── 3.1 从Prompt到Harness的范式革命
    ├── 3.2 Harness的七大核心组件
    ├── 3.3 上下文工程：CLAUDE.md设计模式
    ├── 3.4 验证回路与自检机制
    ├── 3.5 架构约束与护栏设计
    ├── 3.6 可观测性与调试
    └── 3.7 Harness设计清单

---

第二部分：AI Agent系统设计（4章）
├── 第4章 Agent架构设计与决策框架
│   ├── 4.1 Agent vs 传统后端系统
│   ├── 4.2 主流Agent框架对比
│   ├── 4.3 需求分析：何时需要Agent
│   ├── 4.4 架构设计方法论
│   ├── 4.5 ReACT模式详解
│   ├── 4.6 Plan-and-Execute模式
│   ├── 4.7 状态机与混合架构
│   └── 4.8 数据流与状态管理
│
├── 第5章 工具系统与MCP协议
│   ├── 5.1 Agent工具系统设计原则
│   ├── 5.2 工具抽象与接口设计
│   ├── 5.3 工具发现与动态注册
│   ├── 5.4 MCP协议详解
│   ├── 5.5 工具编排与组合
│   ├── 5.6 错误处理与容错
│   └── 5.7 工具安全与权限控制
│
├── 第6章 多Agent协作与工作流编排
│   ├── 6.1 单Agent vs 多Agent场景
│   ├── 6.2 多Agent协作模式
│   ├── 6.3 Agent间通信机制
│   ├── 6.4 工作流编排策略
│   ├── 6.5 冲突解决与一致性
│   └── 6.6 协调Agent设计
│
└── 第7章 可观测性与成本优化
    ├── 7.1 Agent系统的可观测性挑战
    ├── 7.2 日志、指标与追踪
    ├── 7.3 LLM调用监控
    ├── 7.4 成本分析与优化策略
    ├── 7.5 性能优化技巧
    └── 7.6 调试与故障排查

---

第三部分：完整实战案例（2章）
├── 第8章 DoD Agent：电商告警自动处理系统
│   ├── 8.1 项目背景与设计目标
│   ├── 8.2 需求分析与可行性评估
│   ├── 8.3 架构设计：状态机+ReACT混合模式
│   ├── 8.4 核心组件实现
│   ├── 8.5 知识库与RAG集成
│   ├── 8.6 工具系统设计
│   ├── 8.7 分级决策与渐进式学习
│   ├── 8.8 部署与运维实践
│   └── 8.9 效果评估与持续优化
│
└── 第9章 个人知识管理Agent实践
    ├── 9.1 Karpathy的自进化知识库
    ├── 9.2 知识管理系统架构
    ├── 9.3 OpenClaw框架深度解析
    ├── 9.4 技能系统设计
    ├── 9.5 记忆与上下文管理
    ├── 9.6 多平台集成实践
    └── 9.7 从零搭建个人助手

---

第四部分：基础理论补充（2章）
├── 第10章 LLM能力边界与工程化要点
│   ├── 10.1 大语言模型概览
│   ├── 10.2 LLM的能力边界
│   ├── 10.3 推理与生成机制
│   ├── 10.4 模型选择与评估
│   ├── 10.5 Prompt工程基础
│   └── 10.6 LLM工程化最佳实践
│
└── 第11章 RAG与上下文工程基础
    ├── 11.1 RAG系统架构
    ├── 11.2 文档处理与分块策略
    ├── 11.3 向量检索与召回优化
    ├── 11.4 上下文窗口管理
    ├── 11.5 混合检索策略
    └── 11.6 RAG系统评估与优化

---

附录
├── 附录A 术语表
├── 附录B 参考资料与延伸阅读
└── 附录C 常用工具与框架
```

### 2.2 内容规划

**篇幅估算**：
- 第1-9章：每章8,000-12,000字（基于现有素材改编）
- 第10-11章：每章5,000-8,000字（新写，保持精简）
- 总计：约10-12万字

**内容分布**：
- ✅ 有完整素材：第1-9章（7篇博客，约16,700行）
- ⚠️ 需要补充：第10-11章（基础理论，适度精简）

---

## 3. 素材映射方案

### 3.1 博客到章节的详细映射

| 章节 | 博客来源 | 原始行数 | 改编策略 | 工作量 |
|------|---------|---------|---------|--------|
| 第1章 | 00-vibe-coding-vs-spec-coding.md | 2,526 | 完整保留核心内容，转换为书籍格式，移除Hexo Front Matter | 低 |
| 第2章 | 01-claude-code-practices.md | 545 | 保留核心，扩充实战案例和最佳实践 | 中 |
| 第3章 | 06-harness-engineering.md | 398 | 保留框架，扩充核心组件详解和设计清单 | 中 |
| 第4章 | 02-agent-system-design-guid.md (第1-8章) | ~4,000 | 提取架构设计部分，重新组织 | 低 |
| 第5章 | 02-agent-system-design-guid.md (第11章) | ~800 | 工具系统部分+需补充MCP协议细节 | 中 |
| 第6章 | 02-agent-system-design-guid.md (第6章) | ~1,000 | 多Agent协作部分，适度扩充 | 低 |
| 第7章 | 02-agent-system-design-guid.md (第12章) | ~1,000 | 可观测性部分，完整保留 | 低 |
| 第8章 | 03-dod-agent-design.md | 2,370 | 完整案例，优化呈现结构 | 低 |
| 第9章 | 04-karpathy-evolving-knowledge-base.md + 05-openclaw-research-and-practice.md | 948 + 1,006 | 融合两篇，形成完整知识管理实践 | 中 |
| 第10章 | **需新写** | - | 从现有博客提取LLM相关片段+补充基础理论 | 高 |
| 第11章 | **需新写** | - | 从karpathy文章提取RAG部分+补充基础 | 高 |

### 3.2 内容改编指南

**通用改编原则**：
1. 移除Hexo博客的Front Matter（title, date, categories, tags）
2. 保留核心技术内容和代码示例
3. 调整语气：从博客风格到书籍风格（更系统化、教程化）
4. 添加章节引言和小结
5. 统一术语和格式
6. 添加章节间的衔接段落

**每章标准结构**：
```markdown
# 第X章 [章节标题]

## 引言
- 本章背景和问题
- 为什么重要
- 本章内容概览

## X.1 - X.N 核心内容小节
- 概念讲解
- 实战案例
- 代码示例
- 最佳实践

## 常见陷阱
- 问题描述
- 解决方案

## 本章小结
- 核心要点回顾
- 下一章预告
```

---

## 4. 技术实现方案

### 4.1 工具链

**核心工具**：
- **mdBook**：静态电子书生成工具
- **mdbook-mermaid**：图表支持
- **GitHub Actions**：自动构建和部署
- **GitHub Pages**：在线发布

**项目结构**：
```
ai-book/
├── book.toml              # mdBook配置
├── src/
│   ├── SUMMARY.md         # 目录结构（关键）
│   ├── README.md          # 书籍介绍
│   ├── part1/             # 第一部分（3章）
│   │   ├── chapter1.md
│   │   ├── chapter2.md
│   │   └── chapter3.md
│   ├── part2/             # 第二部分（4章）
│   │   ├── chapter4.md
│   │   ├── chapter5.md
│   │   ├── chapter6.md
│   │   └── chapter7.md
│   ├── part3/             # 第三部分（2章）
│   │   ├── chapter8.md
│   │   └── chapter9.md
│   ├── part4/             # 第四部分（2章）
│   │   ├── chapter10.md
│   │   └── chapter11.md
│   ├── appendix/
│   │   ├── glossary.md
│   │   ├── references.md
│   │   └── tools.md
│   └── images/            # 图片资源
├── theme/
│   └── custom.css         # 自定义样式
├── mermaid.min.js
├── mermaid-init.js
└── book/                  # 构建输出目录
```

### 4.2 配置文件

**book.toml 配置**：
```toml
[book]
title = "AI工程实践：从编程到Agent的完整指南"
description = "面向有经验AI工程师的实战指南"
authors = ["wxquare"]
language = "zh-CN"
src = "src"

[build]
build-dir = "book"

[preprocessor.mermaid]
command = "mdbook-mermaid"

[output.html]
site-url = "/ai-book/"
git-repository-url = "https://github.com/wxquare/wxquare.github.io"
additional-js = ["mermaid.min.js", "mermaid-init.js"]

[output.html.fold]
enable = true
level = 0

[output.html.search]
enable = true
```

### 4.3 部署配置

**GitHub Actions工作流**（`.github/workflows/deploy-ai-book.yml`）：
```yaml
name: Deploy AI Book

on:
  push:
    branches: [ hexo ]
    paths:
      - 'ai-book/**'
      - '.github/workflows/deploy-ai-book.yml'
  workflow_dispatch:

permissions:
  contents: write

jobs:
  deploy:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Rust
      uses: actions-rust-lang/setup-rust-toolchain@v1
    
    - name: Cache cargo packages
      uses: actions/cache@v3
      with:
        path: |
          ~/.cargo/bin/mdbook-mermaid
          ~/.cargo/.crates.toml
          ~/.cargo/.crates2.json
        key: ${{ runner.os }}-cargo-mdbook-mermaid-0.17.0
    
    - name: Setup mdBook
      uses: peaceiris/actions-mdbook@v1
      with:
        mdbook-version: 'latest'
    
    - name: Install and initialize mdbook-mermaid
      run: |
        cargo install mdbook-mermaid --version 0.17.0 --locked --force
        cd ai-book
        mdbook-mermaid install .
    
    - name: Build book
      run: |
        cd ai-book
        mdbook build
    
    - name: Deploy to GitHub Pages
      uses: peaceiris/actions-gh-pages@v3
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
        publish_dir: ./ai-book/book
        publish_branch: master
        destination_dir: ai-book
        keep_files: true
        user_name: 'github-actions[bot]'
        user_email: 'github-actions[bot]@users.noreply.github.com'
```

### 4.4 写作规范

遵循项目`.cursorrules`中的规范：
1. **代码块**：必须指定语言（如 ```python, ```go, ```bash）
2. **中英文空格**：中英文之间必须有空格
3. **图表**：使用Mermaid语法
4. **链接**：使用相对路径，章节链接格式 `[章节名](chapterX.md#section)`
5. **图片**：存放在`src/images/`，使用相对路径引用
6. **术语一致性**：使用统一的术语（Agent、LLM、RAG、MCP等）

---

## 5. 实施计划（8-11天）

### 5.1 阶段划分

**阶段1：结构搭建（第1天）**
- [ ] 更新`ai-book/src/SUMMARY.md`为新目录结构
- [ ] 创建所有章节文件（part1-4目录和11个章节文件）
- [ ] 每个章节文件写入基本骨架（标题+引言占位）
- [ ] 更新`book.toml`配置（书名、描述等）
- [ ] 创建/更新GitHub Actions工作流
- [ ] 本地测试构建：`cd ai-book && mdbook build`

**阶段2：素材迁移（第2-6天）**

*第2天*：
- [ ] 第1章：从`00-vibe-coding-vs-spec-coding.md`迁移
  - 移除Front Matter
  - 调整章节结构为6个小节
  - 处理代码块和图表
  - 添加引言和小结

*第3天*：
- [ ] 第2章：从`01-claude-code-practices.md`迁移并扩充
  - 迁移核心内容
  - 扩充实战案例
  - 补充最佳实践部分
- [ ] 第3章：从`06-harness-engineering.md`迁移并扩充
  - 迁移框架内容
  - 扩充核心组件详解
  - 添加设计清单

*第4天*：
- [ ] 第4章：从`02-agent-system-design-guid.md`提取第1-8章
  - 提取架构设计方法论
  - 重新组织为8个小节
  - 补充决策框架
- [ ] 第5章：从`02-agent-system-design-guid.md`提取第11章
  - 提取工具系统内容
  - 补充MCP协议细节
  - 添加实战案例

*第5天*：
- [ ] 第6章：从`02-agent-system-design-guid.md`提取第6章
  - 提取多Agent协作内容
  - 适度扩充工作流编排
- [ ] 第7章：从`02-agent-system-design-guid.md`提取第12章
  - 提取可观测性内容
  - 补充成本优化策略

*第6天*：
- [ ] 第8章：从`03-dod-agent-design.md`迁移
  - 完整保留案例内容
  - 优化呈现结构
  - 添加代码示例
- [ ] 第9章：融合`04-karpathy`和`05-openclaw`
  - 整合知识管理概念
  - 形成完整实践指南
  - 添加从零搭建步骤

**阶段3：补充内容（第7-9天）**

*第7天*：
- [ ] 第10章：LLM基础理论（新写）
  - 从现有博客提取LLM相关片段
  - 补充能力边界
  - 补充工程化要点
  - 精简保持5,000-8,000字

*第8天*：
- [ ] 第11章：RAG基础理论（新写）
  - 从karpathy文章提取RAG部分
  - 补充系统架构
  - 补充优化策略
  - 精简保持5,000-8,000字

*第9天*：
- [ ] 扩充需要补充的章节
  - 第2章：补充更多Claude Code案例
  - 第3章：补充Harness组件详解
  - 第5章：补充MCP协议细节

**阶段4：优化打磨（第10-11天）**

*第10天*：
- [ ] 内容优化
  - 统一术语表（Agent、LLM、RAG、MCP等）
  - 检查所有代码块格式和语言标注
  - 处理图片路径和引用
  - 添加章节间衔接段落
  - 检查内部链接有效性

*第11天*：
- [ ] 附录完善
  - 更新附录A术语表
  - 更新附录B参考资料
  - 创建附录C工具与框架
- [ ] 最终验证
  - 本地构建测试：`mdbook build`
  - 本地预览测试：`mdbook serve`
  - 检查所有章节渲染正常
  - 检查搜索功能
  - 检查目录导航
- [ ] 部署上线
  - 提交代码到git
  - 触发GitHub Actions
  - 验证在线访问

### 5.2 每日产出

| 天数 | 主要产出 | 验收标准 |
|------|---------|---------|
| Day 1 | 完整目录结构+章节骨架 | 本地构建无错误 |
| Day 2 | 第1章完成 | 章节内容完整，格式正确 |
| Day 3 | 第2-3章完成 | 章节内容完整，格式正确 |
| Day 4 | 第4-5章完成 | 章节内容完整，格式正确 |
| Day 5 | 第6-7章完成 | 章节内容完整，格式正确 |
| Day 6 | 第8-9章完成 | 章节内容完整，格式正确 |
| Day 7 | 第10章完成 | 新写内容完整 |
| Day 8 | 第11章完成 | 新写内容完整 |
| Day 9 | 补充内容完成 | 所有章节内容充实 |
| Day 10 | 内容优化完成 | 术语统一，格式规范 |
| Day 11 | 附录完成+部署上线 | 在线访问正常 |

---

## 6. 质量标准

### 6.1 内容完整性

**必须满足**：
- ✅ 11章全部完成，无占位符
- ✅ 每章包含完整结构（引言、核心内容、最佳实践、陷阱、小结）
- ✅ 所有代码示例完整且可运行
- ✅ 所有图表清晰可见（Mermaid渲染正常）
- ✅ 附录A、B、C全部完成

### 6.2 格式规范

**必须满足**：
- ✅ 所有代码块指定语言
- ✅ 中英文之间有空格
- ✅ 术语统一（Agent、LLM、RAG、MCP等）
- ✅ 章节标题层级正确（# ## ### ####）
- ✅ 链接格式正确（相对路径）
- ✅ 图片引用正确（相对路径）

### 6.3 构建与部署

**必须满足**：
- ✅ 本地构建无错误：`mdbook build`
- ✅ 本地预览正常：`mdbook serve`
- ✅ GitHub Actions自动构建成功
- ✅ 在线访问正常：`https://wxquare.github.io/ai-book/`
- ✅ 搜索功能可用
- ✅ 目录导航流畅
- ✅ 所有章节链接有效

### 6.4 阅读体验

**应该满足**：
- ✅ 左侧目录树折叠展开正常
- ✅ 右侧大纲自动生成
- ✅ 代码块支持复制
- ✅ 图表显示清晰
- ✅ 移动端自适应

---

## 7. 风险与应对

### 7.1 技术风险

| 风险 | 概率 | 影响 | 应对措施 |
|------|------|------|----------|
| 博客内容格式不兼容 | 中 | 中 | 编写格式转换脚本，手动检查 |
| 图片路径处理复杂 | 中 | 低 | 统一存放`src/images/`，使用相对路径 |
| Mermaid图表渲染问题 | 低 | 低 | 测试mdbook-mermaid插件，必要时用图片 |
| 构建时间过长 | 低 | 低 | 使用GitHub Actions缓存 |

### 7.2 内容风险

| 风险 | 概率 | 影响 | 应对措施 |
|------|------|------|----------|
| 博客内容不够系统化 | 中 | 中 | 重新组织结构，添加章节间衔接 |
| 第10-11章写作时间不足 | 中 | 中 | 控制篇幅5,000-8,000字，复用现有素材 |
| 术语不统一 | 高 | 低 | 第10天统一检查和替换 |
| 代码示例过时 | 低 | 中 | 标注代码的时间和版本 |

### 7.3 时间风险

| 风险 | 概率 | 影响 | 应对措施 |
|------|------|------|----------|
| 素材迁移时间超预期 | 中 | 中 | 优先完成有素材的章节，延后第10-11章 |
| 优化打磨时间不足 | 低 | 低 | 预留2天缓冲时间 |

---

## 8. 后续优化方向（v1.1+）

### 8.1 内容深化（v1.1）

**重点章节深化**（每章扩展到15,000-20,000字）：
- 第5章：补充更多MCP协议案例和实现细节
- 第7章：补充完整的监控和调试实践
- 第8章：补充DoD Agent的性能优化和迭代过程

### 8.2 新增内容（v1.2）

**可能新增的章节**：
- 第12章：Agent安全与隐私保护
- 第13章：Agent测试与质量保证
- 第14章：Agent团队协作与CI/CD

### 8.3 格式优化（v1.3）

**增强阅读体验**：
- 添加交互式代码示例（CodePen/JSFiddle）
- 补充视频教程链接
- 添加每章练习题
- 生成PDF/ePub版本

---

## 9. 成功标准

### 9.1 里程碑

**M1：结构搭建完成**（第1天）
- ✅ SUMMARY.md更新
- ✅ 11个章节文件创建
- ✅ 本地构建成功

**M2：素材迁移完成**（第6天）
- ✅ 第1-9章全部完成
- ✅ 所有博客内容迁移

**M3：内容补充完成**（第9天）
- ✅ 第10-11章完成
- ✅ 所有章节内容充实

**M4：优化打磨完成**（第11天）
- ✅ 格式规范统一
- ✅ 附录完成
- ✅ 部署上线

### 9.2 最终验收标准

**内容**：
- ✅ 11章全部完成，每章8,000-12,000字（第10-11章5,000-8,000字）
- ✅ 总字数10-12万字
- ✅ 所有代码示例完整可运行
- ✅ 所有图表清晰可见

**格式**：
- ✅ 代码块全部指定语言
- ✅ 术语统一
- ✅ 链接有效
- ✅ 构建无错误

**部署**：
- ✅ GitHub Actions自动构建成功
- ✅ 在线访问正常
- ✅ 搜索功能可用
- ✅ 移动端适配良好

---

## 10. 参考资料

### 10.1 技术文档

- mdBook官方文档：https://rust-lang.github.io/mdBook/
- mdbook-mermaid插件：https://github.com/badboy/mdbook-mermaid
- GitHub Pages文档：https://docs.github.com/en/pages
- GitHub Actions文档：https://docs.github.com/en/actions

### 10.2 项目规范

- 项目.cursorrules文件
- ecommerce-book项目结构（参考）
- CLAUDE.md（项目级配置）
- BOOK_PROJECT.md（电商书籍项目文档）

### 10.3 博客素材清单

1. `source/_posts/AI/00-vibe-coding-vs-spec-coding.md` (2,526行)
2. `source/_posts/AI/01-claude-code-practices.md` (545行)
3. `source/_posts/AI/02-agent-system-design-guid.md` (8,927行)
4. `source/_posts/AI/03-dod-agent-design.md` (2,370行)
5. `source/_posts/AI/04-karpathy-evolving-knowledge-base.md` (948行)
6. `source/_posts/AI/05-openclaw-research-and-practice.md` (1,006行)
7. `source/_posts/AI/06-harness-engineering.md` (398行)

---

**文档版本**：v1.0  
**最后更新**：2026-04-21  
**下次审查**：实施完成后（预计2026-05-02）
