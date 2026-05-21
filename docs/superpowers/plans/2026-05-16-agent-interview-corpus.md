# Agent Interview Corpus Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a curated internet-sourced LLM / Agent interview question bank while preserving the current Appendix D content.

**Architecture:** Keep Appendix D as the polished tutorial chapter. Add a separate question bank file under `books/ai-book/src/appendix/` with source links, topic tags, summaries, and rewritten interview prompts derived from public pages.

**Tech Stack:** Markdown, web search/open results, existing npm/Hexo build verification.

---

## Files

- Create: `books/ai-book/src/appendix/llm-agent-interview-question-bank.md`
- Modify: `books/ai-book/src/appendix/system-design-interview-portfolio.md`
- Verify with: `rg` and `npm run clean && npm run build`

## Task 1: Create Question Bank

**Files:**
- Create: `books/ai-book/src/appendix/llm-agent-interview-question-bank.md`

- [x] **Step 1: Add source taxonomy**

The question bank must separate:

```text
official job signals
official engineering docs
public interview question banks
community interview reports
derived question bank
```

- [x] **Step 2: Add source entries**

Each entry must include:

```text
source
url
type
tags
summary
usable interview prompts
appendix mapping
quality note
```

- [x] **Step 3: Add derived question bank**

Group rewritten questions by:

```text
RAG
Agent Harness / Coding Agent
Tool Calling / MCP
Evals
Observability
Guardrails / Security
Multi-agent
Portfolio
```

## Task 2: Link Corpus From Appendix

**Files:**
- Modify: `books/ai-book/src/appendix/system-design-interview-portfolio.md`

- [x] **Step 1: Add one short note near the top**

Add a sentence pointing readers to the separate question bank file for raw source signals.

Expected: Appendix content remains intact and the question bank becomes optional extended reading.

## Task 3: Verify

- [x] **Step 1: Check unfinished placeholders**

Run:

```bash
rg -n "TODO|TBD|FIXME" books/ai-book/src/appendix/llm-agent-interview-question-bank.md books/ai-book/src/appendix/system-design-interview-portfolio.md
```

Expected: no matches.

- [x] **Step 2: Build**

Run:

```bash
npm run clean && npm run build
```

Expected: exit code 0.
