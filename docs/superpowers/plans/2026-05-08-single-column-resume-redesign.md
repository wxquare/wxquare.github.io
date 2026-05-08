# Single Column Resume Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite two one-page Chinese resumes in a single-column format with stronger content strategy and visual hierarchy.

**Architecture:** Keep editable Markdown paired with print-ready HTML and exported PDF. Preserve the original `resume-2026*` files, but overwrite the previously generated `resume-backend-architect.*` and `resume-ai-agent.*` files with a full rewrite.

**Tech Stack:** Markdown, static HTML/CSS, headless Chrome PDF export, Hexo build verification.

---

## File Structure

- Modify `source/about/resume-backend-architect.md`: rewritten backend architecture content.
- Modify `source/about/resume-backend-architect.html`: single-column print layout for backend architecture resume.
- Modify `source/about/resume-backend-architect.pdf`: exported one-page PDF.
- Modify `source/about/resume-ai-agent.md`: rewritten AI Agent content.
- Modify `source/about/resume-ai-agent.html`: single-column print layout for AI Agent resume.
- Modify `source/about/resume-ai-agent.pdf`: exported one-page PDF.
- Do not modify `source/about/resume-2026.md`, `source/about/resume-2026.html`, `source/about/resume-2026-ai-agent.md`, or `source/about/resume-2026-ai-agent.html`.

### Task 1: Rewrite Backend Architecture Resume

**Files:**
- Modify: `source/about/resume-backend-architect.md`
- Modify: `source/about/resume-backend-architect.html`

- [ ] **Step 1: Replace Markdown content**

Rewrite the Markdown as a one-page single-column resume with: header, career summary, key metrics, Shopee themed experience blocks, Tencent compressed experience, technical output, education.

- [ ] **Step 2: Replace HTML content**

Rewrite the HTML with single-column A4 print CSS. Use a compact but readable layout with a summary box, metrics row, company blocks, theme blocks, and a technical-output section.

- [ ] **Step 3: Check density**

Run: `wc -l source/about/resume-backend-architect.md source/about/resume-backend-architect.html`

Expected: files exist and are concise enough for manual layout review.

### Task 2: Rewrite AI Agent Resume

**Files:**
- Modify: `source/about/resume-ai-agent.md`
- Modify: `source/about/resume-ai-agent.html`

- [ ] **Step 1: Replace Markdown content**

Rewrite the Markdown as a one-page single-column resume with: header, Agent-focused career summary, key metrics, DoD Agent block, Agent engineering foundation block, compressed ecommerce systems block, Tencent AI engineering block, technical output, education.

- [ ] **Step 2: Replace HTML content**

Rewrite the HTML with the same single-column design language as the backend resume, but tune section content and metrics toward AI Agent roles.

- [ ] **Step 3: Check density**

Run: `wc -l source/about/resume-ai-agent.md source/about/resume-ai-agent.html`

Expected: files exist and are concise enough for manual layout review.

### Task 3: Export and Inspect PDFs

**Files:**
- Modify: `source/about/resume-backend-architect.pdf`
- Modify: `source/about/resume-ai-agent.pdf`

- [ ] **Step 1: Export backend PDF**

Run headless Chrome:

```bash
"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless --disable-gpu --no-sandbox --print-to-pdf=/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/source/about/resume-backend-architect.pdf file:///Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/source/about/resume-backend-architect.html
```

- [ ] **Step 2: Export AI Agent PDF**

Run headless Chrome:

```bash
"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless --disable-gpu --no-sandbox --print-to-pdf=/Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/source/about/resume-ai-agent.pdf file:///Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io/source/about/resume-ai-agent.html
```

- [ ] **Step 3: Verify page count**

Run: `file source/about/resume-backend-architect.pdf source/about/resume-ai-agent.pdf`

Expected: both are PDF documents with `1 pages`.

- [ ] **Step 4: Generate layout screenshots**

Run headless Chrome screenshots to `/private/tmp/resume-backend-architect-redesign.png` and `/private/tmp/resume-ai-agent-redesign.png`, then visually inspect for clipping, overlap, tiny text, or top-heavy whitespace.

### Task 4: Build Verification and Status

**Files:**
- Verify all new resume files.

- [ ] **Step 1: Run Hexo build verification**

Run: `npm run clean && npm run build`

Expected: command exits 0.

- [ ] **Step 2: Review git status**

Run: `git status --short`

Expected: rewritten resume files and this plan are present; unrelated existing dirty files remain untouched.
