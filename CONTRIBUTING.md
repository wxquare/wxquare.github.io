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

- `AGENTS.md` is the normative source for post placement, writing conventions, category mappings, and AI collaboration rules.
- Use `.agents/config/post-categories.json` only as the machine-readable category directory mapping used by tooling.
- `source/about/` for the minimal contact page; email only; do not add resumes, phone numbers, or work materials
- `books/ai-book/src/` for long-form Agent book chapters
- `books/ai-book/labs/llm-from-scratch/` is a retained legacy experiment; do not add new runnable experiments here. New experiments must use an independent Git repository under `/Users/xianguiwang/Projects/`, with GitHub visibility Private by default.
- `docs/` for maintenance notes and internal organization docs

Do not edit generated output directly:

- `public/`
- `.deploy_git/`
- `books/ai-book/book/`

## Writing conventions

Read the writing and content-placement rules in `AGENTS.md`, especially sections 3, 4, and 10. That document defines Front Matter, categories, tags, image paths, code blocks, naming, and content-type boundaries.

## Validate before submitting

Run:

```bash
bash tools/pre-commit-check.sh
```

Today this script performs the required repository validation:

```bash
npm test
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
- `docs/library-policy.md`
