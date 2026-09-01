# Contributing

This repository is a content-heavy technical blog and book workspace. Contributions should keep source-of-truth boundaries clear and leave generated output reproducible from source.

## Scope

Typical contributions include:

- new blog posts
- improvements to existing technical articles
- AI and Agent book chapters
- diagrams and supporting assets
- repository documentation and collaboration workflows

## Local setup

From the repository root:

```bash
npm install
npm run server
```

Local preview:

```text
http://localhost:4000
```

## Where to make changes

- `source/_posts/AI/` for reader-facing AI and Agent blog posts
- `source/_posts/system-design/` for system design articles
- `source/about/` for the minimal contact page; email only; do not add resumes, phone numbers, or work materials
- `books/ai-book/src/` for long-form Agent book chapters
- `books/ai-book/labs/` for runnable Agent experiments
- `docs/` for maintenance notes and internal organization docs

Do not edit generated output directly:

- `public/`
- `.deploy_git/`
- `books/ai-book/book/`

## Writing conventions

Every post should include valid front matter:

```yaml
---
title: 文章标题
date: YYYY-MM-DD
categories:
  - 分类
tags:
  - tag
---
```

Please follow these rules:

- category depth at most 2 levels
- use relative image paths
- specify languages for code fences
- keep spaces between Chinese and English text
- keep filenames descriptive and consistent with the existing series style

## Validate before submitting

Run:

```bash
bash bin/pre-commit-check.sh
```

Today this script performs the required repository validation:

```bash
npm run clean
npm run build
```

## Review checklist

Before opening a PR, confirm:

- source files were edited instead of generated files
- links and image paths still resolve
- new content is stored in the correct canonical directory
- build passes locally
- changes do not introduce duplicate long-form content across blog, books, and docs

## AI collaboration files

If your change affects repository collaboration rules or AI workflows, check:

- `AGENTS.md`
- `.cursorrules`
- `.agents/`
- `docs/agent-development-guide.md`
