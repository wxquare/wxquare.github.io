# AI Book Part 2 Chapter Evidence Revision Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Revise all ten Agent-engineering chapters so each has 10,000–12,000 Chinese characters of正文, at least 20 authoritative references, and correctly mapped `[n]` citations in the body, beginning with chapters 8–10 as the first validated batch.

**Architecture:** Keep the existing Markdown chapter structure and independent per-chapter reference numbering. Use an evidence matrix to map claims to sources before expanding prose, then run a deterministic audit that reports Han-character counts, reference counts, citation-number validity, and citation coverage. Deliver chapters in batches of at most three and perform editorial review after each batch.

**Tech Stack:** Markdown, Node.js standard library for audit scripts, existing book build tooling, official papers/specifications/API documentation.

## Global Constraints

- Each chapter正文 targets 10,000–12,000 Unicode Han characters; the audit also reports total characters and non-whitespace characters.
- Each chapter has at least 20 sources under `## 参考资料`, with title, author or institution, year when available, and stable URL.
- Each body citation uses independent per-chapter `[n]` numbering starting at `[1]`; every used number must resolve to that chapter’s bibliography.
- Keep existing chapter order, Markdown conventions, examples, tables, Mermaid blocks, and tone unless correction is required.
- Prefer primary papers, standards, official documentation, official technical reports, and maintained open-source documentation; do not use search snippets or unattributed reposts as authoritative sources.
- Do not modify the unrelated untracked file `books/ai-book/chinese-character-count-by-section.md`.
- Do not change the site theme, global book navigation, or unrelated chapters.

## File Map

- `books/ai-book/src/part2/01-agent-architecture.md`: Chapter 8 architecture and runtime content.
- `books/ai-book/src/part2/02-prompt-engineering.md`: Chapter 9 prompt protocols and structured output.
- `books/ai-book/src/part2/03-context-engineering.md`: Chapter 10 context information architecture.
- `books/ai-book/src/part2/04-harness-engineering.md` through `10-agent-evals-guardrails-observability.md`: Chapters 11–17 revised in later batches.
- `books/ai-book/scripts/audit-part2-chapters.mjs`: Deterministic chapter audit and citation validation.
- `docs/superpowers/research/part2-evidence/`: Committed evidence matrices, one Markdown file per chapter, used to review claim-to-source alignment.
- `docs/superpowers/reports/`: Batch audit reports with counts, unresolved citations, and editorial findings.

---

### Task 1: Build the chapter audit and evidence-matrix format

**Files:**
- Create: `books/ai-book/scripts/audit-part2-chapters.mjs`
- Create: `docs/superpowers/research/part2-evidence/README.md`
- Create: `docs/superpowers/reports/2026-09-21-part2-baseline-audit.md`

**Interfaces:**
- Consumes: chapter Markdown files under `books/ai-book/src/part2/`.
- Produces: one JSON-like console report per chapter containing正文 Han count, total count, bibliography count, body citation numbers, undefined numbers, unused bibliography numbers, and first/last reference lines; exits non-zero when requested chapters violate hard checks.

- [ ] **Step 1: Define the evidence-matrix template**

  Add columns `Claim ID`, `Chapter section`, `Claim or paragraph summary`, `Source number`, `Source type`, `What the source actually supports`, and `Editorial action`. State that an architecture recommendation must be labeled as synthesis when no source directly states it.

- [ ] **Step 2: Implement the audit parser**

  In `audit-part2-chapters.mjs`, read UTF-8 Markdown, split正文 at the first `## 参考资料`, count Han characters with `/\p{Script=Han}/gu`, parse numbered reference lines with `/^(\d+)\.\s+/gm`, parse body citations with `/\[(\d+)\]/g`, and report missing or unused numbers. Ignore code blocks when computing citation coverage, but still report citations inside tables and prose.

- [ ] **Step 3: Run the baseline audit**

  Run:

  ```bash
  node books/ai-book/scripts/audit-part2-chapters.mjs books/ai-book/src/part2/*.md
  ```

  Record current counts and known gaps in `docs/superpowers/reports/2026-09-21-part2-baseline-audit.md`; do not edit chapter content in this task.

- [ ] **Step 4: Validate the audit itself**

  Run the command against a temporary copy containing one undefined citation and one skipped reference number, confirm both errors are reported, then remove the temporary copy. Run `node --check books/ai-book/scripts/audit-part2-chapters.mjs` and `git diff --check`.

- [ ] **Step 5: Commit**

  ```bash
  git add books/ai-book/scripts/audit-part2-chapters.mjs docs/superpowers/research/part2-evidence/README.md docs/superpowers/reports/2026-09-21-part2-baseline-audit.md
  git commit -m "chore: add part2 chapter evidence audit"
  ```

### Task 2: Create evidence matrices for chapters 8–10

**Files:**
- Create: `docs/superpowers/research/part2-evidence/chapter-08-agent-architecture.md`
- Create: `docs/superpowers/research/part2-evidence/chapter-09-prompt-engineering.md`
- Create: `docs/superpowers/research/part2-evidence/chapter-10-context-engineering.md`

**Interfaces:**
- Consumes: current chapter headings and the source-selection rules in the design spec.
- Produces: a reviewed set of at least 20 source records and claim mappings that the three chapter revisions can cite without inventing unsupported facts.

- [ ] **Step 1: Map chapter claims to sections**

  For each chapter, list every一级小节 and identify its definitions, research claims, protocol/product facts, engineering recommendations, example claims, and evaluation criteria. Mark claims that are currently uncited or too broad.

- [ ] **Step 2: Select and classify sources**

  For chapter 8, include sources spanning ReAct, MRKL/Toolformer, agent planning or orchestration research, durable execution/state, MCP/A2A, and official runtime tracing/guardrail/handoff documentation. For chapter 9, include instruction-following and few-shot/chain-of-thought research, structured-output/schema documentation, prompt-injection guidance, and evaluation/observability documentation. For chapter 10, include Transformer and long-context research, RAG/DPR and reranking research, memory/context-management work, prompt-injection security research, and official retrieval/tool/context specifications.

- [ ] **Step 3: Record support boundaries**

  For every source, write one sentence describing exactly which claim it supports and one sentence describing what it does not prove. Replace sources that cannot support a precise claim; do not retain a source only to reach the count of 20.

- [ ] **Step 4: Review numbering**

  Number sources per chapter from `[1]`, ensure every planned source will be used in prose, and ensure no claim cites a source outside the source’s support boundary.

- [ ] **Step 5: Commit**

  ```bash
  git add docs/superpowers/research/part2-evidence/chapter-08-agent-architecture.md docs/superpowers/research/part2-evidence/chapter-09-prompt-engineering.md docs/superpowers/research/part2-evidence/chapter-10-context-engineering.md
  git commit -m "docs: map evidence for part2 sample chapters"
  ```

### Task 3: Revise chapters 8–10 using the evidence matrices

**Files:**
- Modify: `books/ai-book/src/part2/01-agent-architecture.md`
- Modify: `books/ai-book/src/part2/02-prompt-engineering.md`
- Modify: `books/ai-book/src/part2/03-context-engineering.md`

**Interfaces:**
- Consumes: the three evidence matrices from Task 2 and the citation parser from Task 1.
- Produces: three standalone chapters with independent numbered bibliographies and claim-level citations.

- [ ] **Step 1: Reshape each chapter before adding prose**

  Keep the current headings where they are coherent, merge only genuinely repetitive paragraphs, and identify missing sections needed to reach the target: a concept boundary, an engineering workflow, a failure/negative case, a concrete example, and an evaluation checklist.

- [ ] **Step 2: Add claim-level citations**

  Place `[n]` directly after definitions, research findings, protocol semantics, product behavior, and externally verifiable claims. Use multiple numbers where a paragraph combines independent claims. Label synthesis as engineering guidance rather than attributing it to a source.

- [ ] **Step 3: Expand with technical substance**

  Add comparisons, interfaces, state transitions, examples, failure handling, and evaluation criteria that are specific to the chapter topic. Avoid repeating the same explanation of tools, memory, RAG, or guardrails across all three chapters; cross-reference the later chapter conceptually without copying paragraphs.

- [ ] **Step 4: Normalize each bibliography**

  Replace incomplete or generic entries with at least 20 precise entries. Keep the bibliography after the正文 and make every entry’s number match the body.

- [ ] **Step 5: Run chapter-level verification**

  Run:

  ```bash
  node books/ai-book/scripts/audit-part2-chapters.mjs books/ai-book/src/part2/01-agent-architecture.md books/ai-book/src/part2/02-prompt-engineering.md books/ai-book/src/part2/03-context-engineering.md
  git diff --check
  ```

  Resolve every undefined citation, unused source, bibliography count below 20, and正文 count outside 9,800–12,200 before review.

- [ ] **Step 6: Commit the sample batch**

  ```bash
  git add books/ai-book/src/part2/01-agent-architecture.md books/ai-book/src/part2/02-prompt-engineering.md books/ai-book/src/part2/03-context-engineering.md
  git commit -m "docs: revise agent architecture prompt and context chapters"
  ```

### Task 4: Review and report the sample batch

**Files:**
- Create: `docs/superpowers/reports/2026-09-21-part2-sample-batch-audit.md`
- Modify: `books/ai-book/src/part2/01-agent-architecture.md` through `03-context-engineering.md` only if review finds a citation mismatch or factual overreach.

**Interfaces:**
- Consumes: revised chapters, audit output, and evidence matrices.
- Produces: an auditable sample-batch report and a confirmed template for the remaining seven chapters.

- [ ] **Step 1: Perform source-to-claim spot checks**

  For every一级小节, inspect at least one cited paragraph and compare its wording with the linked source. Check protocol/version-sensitive claims against the exact official document, and check research claims against the abstract or stated results rather than a secondary summary.

- [ ] **Step 2: Check editorial quality**

  Search for duplicated paragraphs, unsupported numerical claims, vague “业界普遍认为” statements, uncited definitions, stale product names, and references that are listed but never used.

- [ ] **Step 3: Record results**

  Write the three chapter counts, source-type distribution, spot-check outcomes, unresolved limitations, and the exact acceptance decision in `2026-09-21-part2-sample-batch-audit.md`.

- [ ] **Step 4: Run Markdown/build checks**

  Use the repository’s documented book build command if available; otherwise run the existing Markdown/Hexo validation command discovered in the repository. Confirm no broken internal links or malformed code fences are introduced.

- [ ] **Step 5: Commit the report**

  ```bash
  git add docs/superpowers/reports/2026-09-21-part2-sample-batch-audit.md
  git commit -m "docs: audit part2 sample chapter revision"
  ```

### Task 5: Revise chapters 11–13

**Files:**
- Modify: `books/ai-book/src/part2/04-harness-engineering.md`
- Modify: `books/ai-book/src/part2/05-llm-api-protocol.md`
- Modify: `books/ai-book/src/part2/06-tool-calling-mcp.md`
- Create: `docs/superpowers/research/part2-evidence/chapter-11-harness-engineering.md`
- Create: `docs/superpowers/research/part2-evidence/chapter-12-llm-api-protocol.md`
- Create: `docs/superpowers/research/part2-evidence/chapter-13-tool-calling-mcp.md`

**Interfaces:**
- Consumes: the sample-batch acceptance criteria and audit script.
- Produces: three revised chapters, each with 20+ mapped sources and正文 within the target band.

- [ ] **Step 1: Build three evidence matrices** covering harness layers/evals/observability, model API/message/streaming/tool protocols, and Tool Calling/Skills/MCP/sandbox permissions respectively.
- [ ] **Step 2: Expand and cite the three chapters** while preserving their existing section boundaries and correcting protocol version claims.
- [ ] **Step 3: Run the audit, source spot checks, Markdown checks, and duplicate-content search.**
- [ ] **Step 4: Commit with `git commit -m "docs: revise harness API and tool chapters"` after all checks pass.

### Task 6: Revise chapters 14–16

**Files:**
- Modify: `books/ai-book/src/part2/07-agent-knowledge-systems.md`
- Modify: `books/ai-book/src/part2/08-agent-memory.md`
- Modify: `books/ai-book/src/part2/09-workflow-orchestration.md`
- Create: `docs/superpowers/research/part2-evidence/chapter-14-agent-knowledge-systems.md`
- Create: `docs/superpowers/research/part2-evidence/chapter-15-agent-memory.md`
- Create: `docs/superpowers/research/part2-evidence/chapter-16-workflow-orchestration.md`

**Interfaces:**
- Consumes: the sample-batch acceptance criteria and audit script.
- Produces: three revised chapters, each with 20+ mapped sources and正文 within the target band.

- [ ] **Step 1: Build source matrices** for RAG/Agentic RAG and knowledge governance, memory lifecycle and deletion, and workflow/state-machine/multi-agent orchestration.
- [ ] **Step 2: Expand each chapter with evidence-backed pipelines, failure modes, trade-off tables, examples, and evaluation criteria; distinguish research evidence from framework-specific behavior.**
- [ ] **Step 3: Run all count, citation, bibliography, duplication, and Markdown checks.
- [ ] **Step 4: Commit with `git commit -m "docs: revise knowledge memory and workflow chapters"` after all checks pass.

### Task 7: Revise chapter 17 and complete the ten-chapter audit

**Files:**
- Modify: `books/ai-book/src/part2/10-agent-evals-guardrails-observability.md`
- Create: `docs/superpowers/research/part2-evidence/chapter-17-agent-evals-guardrails-observability.md`
- Create: `docs/superpowers/reports/2026-09-21-part2-final-audit.md`

**Interfaces:**
- Consumes: all prior chapter conventions and the audit script.
- Produces: the final governance chapter and one report proving the ten-chapter acceptance criteria.

- [ ] **Step 1: Build the evidence matrix** for offline/online evals, guardrail layers, trace grading, red teaming, privacy/safety, and production observability.
- [ ] **Step 2: Expand and cite chapter 17** to the same正文 and bibliography targets, distinguishing metrics, judgment protocols, and operational controls.
- [ ] **Step 3: Run the full audit**

  ```bash
  node books/ai-book/scripts/audit-part2-chapters.mjs books/ai-book/src/part2/*.md
  git diff --check
  ```

- [ ] **Step 4: Perform final editorial review** for cross-chapter duplication, terminology consistency, citation correctness, stale URLs, source diversity, and build compatibility.
- [ ] **Step 5: Record the final report** with one row per chapter: Han count, total count, bibliography count, used citation count, unused citation count, spot-check status, and final disposition.
- [ ] **Step 6: Commit with `git commit -m "docs: complete evidence revision of agent chapters"` only after every chapter passes.

### Task 8: Final verification before handoff

**Files:**
- Modify: only files flagged by the final audit; do not alter unrelated worktree files.

**Interfaces:**
- Consumes: final audit report and repository build command.
- Produces: verified working tree and a concise handoff summary with links to all changed chapter files and reports.

- [ ] **Step 1: Re-run the deterministic audit from a clean process.** Confirm all ten chapters meet 9,800–12,200 Han characters, contain at least 20 references, and have no invalid citation numbers.
- [ ] **Step 2: Re-run the book build or Markdown validation.** Capture the command and result in the final report.
- [ ] **Step 3: Inspect `git diff --stat`, `git diff --check`, and `git status --short`.** Verify the pre-existing `books/ai-book/chinese-character-count-by-section.md` remains untouched.
- [ ] **Step 4: Deliver the final handoff.** Include the exact counts, any accepted deviations, the changed files, and the remaining editorial risks; do not claim completion without the audit output.
