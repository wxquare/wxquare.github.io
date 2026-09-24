---
title: 使用 Hexo 与 GitHub Pages 搭建博客：双分支部署与日常维护
date: 2023-08-13
description: 介绍 Hexo 博客使用 GitHub Pages 双分支部署的目录约定、发布流程和日常维护方法。
categories:
  - other
tags:
  - hexo
  - github-pages
  - blog
  - mathjax
---

本文记录使用 Hexo、NexT 和 GitHub Pages 搭建博客的完整流程。双分支方案将 Markdown 源码与生成后的静态文件分开管理：`hexo` 分支保存源码，`master` 分支接收 `hexo deploy` 生成的站点文件。实际分支名称可以按仓库设置调整。

## 一、分支与部署模型

```text
hexo 分支：文章、主题配置、图片和 package.json
    |
    | hexo clean && hexo generate && hexo deploy
    v
master 分支：public/ 中的静态站点，由 GitHub Pages 发布
```

源码分支只提交 Markdown、配置和资源文件；`public/` 是构建产物，不直接编辑。若使用 GitHub Actions，也可以让 Actions 发布构建产物，此时不必维护 `master` 分支，关键是明确唯一的发布入口。

## 二、环境准备

建议先确认 Node.js、npm 和 Hexo 版本，再安装项目依赖。项目依赖应优先写入 `package.json`，避免只在全局安装工具导致环境不可复现。

```bash
node --version
npm --version
npx hexo version
npm install
```

新项目可以执行：

```bash
npx hexo init wxquare.github.io
cd wxquare.github.io
npm install
npm install hexo-deployer-git --save
```

Hexo 官方文档：[Hexo 文档](https://hexo.io/zh-cn/docs/)。

## 三、配置 GitHub Pages 部署

在站点 `_config.yml` 中配置部署插件。仓库地址、分支和鉴权方式必须替换为自己的设置：

```yaml
deploy:
  type: git
  repo: https://github.com/wxquare/wxquare.github.io.git
  branch: master
```

使用 HTTPS 时需要 GitHub 支持的凭据方式；使用 SSH 时将 `repo` 改为 SSH 地址，并确认本机密钥已经加入 GitHub。不要把 Token、密码或私钥写入文章和配置文件。

## 四、安装和配置 NexT 主题

主题文档：[NexT Getting Started](https://theme-next.js.org/docs/getting-started/) 和 [NexT Theme Settings](https://theme-next.js.org/docs/theme-settings/)。安装主题后，在主题配置中按需启用菜单、搜索、代码高亮和数学公式等功能。主题版本升级前，先阅读升级说明并保留当前配置备份。

## 五、写作、预览与发布

新文章放在 `source/_posts/` 下，并使用包含 `title`、`date`、`categories` 和至少两个 `tags` 的 Front Matter。

```bash
npx hexo new post "文章标题"
npx hexo clean
npx hexo generate
npx hexo server
```

在浏览器打开 <http://localhost:4000/> 检查页面、图片、代码块和内部链接。确认无误后发布：

```bash
npx hexo deploy
git add source themes package.json package-lock.json _config.yml
git commit -m "docs: update blog posts"
git push origin hexo
```

推荐先提交源码，再执行部署；部署失败时可以重新运行 `hexo clean` 和 `hexo deploy`，或从最近一次源码提交恢复。不要手工修改 `public/` 来修复源码问题。

## 六、启用 LaTeX 数学公式

当前项目已安装 `hexo-filter-mathjax`。服务端插件负责把 Markdown 中的公式转换为 HTML，NexT 的 MathJax 配置负责页面端的公式能力；两者的版本和配置应与当前主题文档保持一致。

在主题配置中按需启用 MathJax：

```yaml
math:
  every_page: false
  mathjax:
    enable: true
    tags: none
```

需要公式的文章在 Front Matter 中声明：

```yaml
mathjax: true
```

正文示例：行内公式 `$E = mc^2$`，块级公式：

```text
$$Agent = Model + Harness$$
```

如果公式无法渲染，先检查插件是否安装、主题配置是否生效，再用 `hexo clean` 重新生成。

## 七、图片、域名与常见故障

- 图片放在 `source/images/`，文章使用站内路径，例如 `/images/example.png`。
- 自定义域名需要在 GitHub Pages 仓库和 DNS 服务商两侧完成配置，并确认 HTTPS 已启用。
- 页面没有更新时，先确认当前分支、部署目标分支和 GitHub Pages 的发布源一致。
- 本地与线上效果不一致时，删除 `public/` 后重新执行 `npx hexo clean && npx hexo generate`。
- 部署鉴权失败时，检查 SSH key、HTTPS 凭据和远端仓库权限，不要把凭据写进命令或文章。

## 参考资料

- [Hexo 官方文档](https://hexo.io/zh-cn/docs/)
- [NexT Getting Started](https://theme-next.js.org/docs/getting-started/)
- [NexT Theme Settings](https://theme-next.js.org/docs/theme-settings/)
- [GitHub Pages 文档](https://docs.github.com/pages)
