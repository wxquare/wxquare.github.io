# LLM Foundation Part 4 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a new fourth part to `books/ai-book` that teaches large-model foundations with the structure requested by the user: macro concepts first, then industrial practice, then current research.

**Architecture:** Create `books/ai-book/src/part4/` with seven focused chapters. Move the existing KV cache appendix into chapter 24 and remove its appendix entry, so foundational material lives in the book body while appendices remain reference-oriented.

**Tech Stack:** mdBook Markdown, Mermaid diagrams, existing repo build commands.

---

### Task 1: Update Book Structure

**Files:**
- Modify: `books/ai-book/src/SUMMARY.md`

- [ ] Insert `# 第四部分：大模型基础专题` after 第三部分 B and before 附录.
- [ ] Add chapter links 21 through 27 exactly as requested.
- [ ] Remove the old appendix entry for `appendix/llm-inference-performance-kv-cache.md`.

### Task 2: Create Fourth-Part Chapters

**Files:**
- Create: `books/ai-book/src/part4/01-llm-foundation-roadmap.md`
- Create: `books/ai-book/src/part4/02-tokenizer-embedding-context.md`
- Create: `books/ai-book/src/part4/03-transformer-decoder-attention.md`
- Create: `books/ai-book/src/part4/04-llm-inference-kv-cache.md`
- Create: `books/ai-book/src/part4/05-training-alignment.md`
- Create: `books/ai-book/src/part4/06-finetuning-quantization-serving.md`
- Create: `books/ai-book/src/part4/07-embedding-rerank-rag-basics.md`

- [ ] Each chapter starts with macro understanding.
- [ ] Each chapter includes industrial practice.
- [ ] Each chapter includes research status as of 2026-05.
- [ ] Each chapter includes interview/engineering takeaways.
- [ ] Use citations to primary sources or official docs where claims are time-sensitive or technical.

### Task 3: Migrate KV Cache Content

**Files:**
- Modify: `books/ai-book/src/part4/04-llm-inference-kv-cache.md`
- Delete: `books/ai-book/src/appendix/llm-inference-performance-kv-cache.md`

- [ ] Preserve and improve the existing KV cache explanation.
- [ ] Reframe it as chapter 24, not an appendix.
- [ ] Keep the clarification that KV means Transformer Key / Value vectors, not database key-value.

### Task 4: Verify

**Commands:**
- [ ] Run `mdbook build books/ai-book`.
- [ ] Run `npm run clean && npm run build`.
- [ ] Search generated book for the new part and KV cache page.
