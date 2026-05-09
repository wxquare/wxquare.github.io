# Two Resume Variants Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build two one-page resume variants, one for backend architecture roles and one for AI Agent engineering roles.

**Architecture:** Keep resume content as plain Markdown for editing, paired with hand-tuned print HTML for PDF export. Preserve existing `resume-2026*` files and create new target-specific files so older versions remain available.

**Tech Stack:** Markdown, static HTML/CSS, browser-based PDF export, Hexo build verification.

---

## File Structure

- Create `source/about/resume-backend-architect.md`: editable backend architecture resume.
- Create `source/about/resume-backend-architect.html`: print-ready backend architecture resume.
- Create `source/about/resume-backend-architect.pdf`: exported one-page PDF.
- Create `source/about/resume-ai-agent.md`: editable AI Agent resume.
- Create `source/about/resume-ai-agent.html`: print-ready AI Agent resume.
- Create `source/about/resume-ai-agent.pdf`: exported one-page PDF.
- Do not modify `source/about/resume-2026.md`, `source/about/resume-2026.html`, `source/about/resume-2026-ai-agent.md`, or `source/about/resume-2026-ai-agent.html`.

### Task 1: Write Backend Architecture Resume

**Files:**
- Create: `source/about/resume-backend-architect.md`
- Create: `source/about/resume-backend-architect.html`

- [ ] **Step 1: Create Markdown content**

Create a one-page resume with these sections: header, education, Shopee work experience, Tencent work experience. Use the role target `后端架构 / 电商核心系统`. Keep Shopee as the main body and include: pricing center, product operation platform, inventory system, OTA delivery, stability governance.

- [ ] **Step 2: Create print HTML**

Create a matching HTML version with A4 print CSS, compact typography, black-and-white styling, and the same content as the Markdown. Use compact line heights and margins so the PDF fits one page.

- [ ] **Step 3: Review content density**

Run: `wc -l source/about/resume-backend-architect.md source/about/resume-backend-architect.html`

Expected: both files exist and are compact enough to inspect manually.

### Task 2: Write AI Agent Resume

**Files:**
- Create: `source/about/resume-ai-agent.md`
- Create: `source/about/resume-ai-agent.html`

- [ ] **Step 1: Create Markdown content**

Create a one-page resume with these sections: header, education, Shopee work experience, Tencent work experience, technical output. Use the role target `AI Agent 工程 / 后端架构`. Put DoD Agent first, then compress ecommerce core systems as production backend evidence.

- [ ] **Step 2: Create print HTML**

Create a matching HTML version with the same print CSS approach as the backend architecture resume. Use slightly tighter spacing if needed because the Agent version has a technical-output section.

- [ ] **Step 3: Review content density**

Run: `wc -l source/about/resume-ai-agent.md source/about/resume-ai-agent.html`

Expected: both files exist and are compact enough to inspect manually.

### Task 3: Export One-Page PDFs

**Files:**
- Create: `source/about/resume-backend-architect.pdf`
- Create: `source/about/resume-ai-agent.pdf`

- [ ] **Step 1: Find an available PDF exporter**

Run: `which wkhtmltopdf || which chromium || which chromium-browser || which google-chrome || which google-chrome-stable`

Expected: one exporter path is printed. If none is printed, check for local Node or bundled browser tooling before stopping.

- [ ] **Step 2: Export PDFs**

Use the available exporter to print:

```bash
source/about/resume-backend-architect.html -> source/about/resume-backend-architect.pdf
source/about/resume-ai-agent.html -> source/about/resume-ai-agent.pdf
```

- [ ] **Step 3: Verify PDF page counts**

Run: `pdfinfo source/about/resume-backend-architect.pdf | rg '^Pages:' && pdfinfo source/about/resume-ai-agent.pdf | rg '^Pages:'`

Expected: both outputs are `Pages:           1`.

### Task 4: Build Verification

**Files:**
- Verify all new resume files.

- [ ] **Step 1: Run Hexo clean and build**

Run: `npm run clean && npm run build`

Expected: command exits 0.

- [ ] **Step 2: Review git status**

Run: `git status --short`

Expected: new resume files and this plan are present; unrelated existing untracked files remain untouched.
