# System Design Architecture Book Merge Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a new mdBook under `books/system-design-architecture-book` titled `系统设计与架构实战：原理、工程与电商案例`.

**Architecture:** Build a new book rather than mutating `books/system-design-book` or `books/ecommerce-book`. Use `system-design-book` as the base skeleton, copy ecommerce chapters/assets with minimal content changes, and add only bridge chapters needed by the new table of contents.

**Tech Stack:** mdBook, Markdown, Mermaid preprocessor shared by `books/scripts/mermaid-preprocessor.py`.

---

### Task 1: Create the new book directory

**Files:**
- Create: `books/system-design-architecture-book/`
- Copy from: `books/system-design-book/`

- [ ] **Step 1: Copy `books/system-design-book` to `books/system-design-architecture-book`**

Run:

```bash
cp -R books/system-design-book books/system-design-architecture-book
```

Expected: the new directory contains `book.toml`, `src/`, `mermaid.min.js`, and `mermaid-init.js`.

### Task 2: Copy ecommerce content into the new book

**Files:**
- Create: `books/system-design-architecture-book/src/part04/`
- Create: `books/system-design-architecture-book/src/part05/`
- Create: `books/system-design-architecture-book/images/`
- Create: `books/system-design-architecture-book/example-codes/`

- [ ] **Step 1: Copy ecommerce method chapters into `part01`**

Copy the ecommerce methodology chapters that become the new method spine:

```bash
cp books/ecommerce-book/src/part1/chapter1.md books/system-design-architecture-book/src/part01/03-architecture-combination.md
cp books/ecommerce-book/src/part1/chapter2.md books/system-design-architecture-book/src/part01/04-business-boundary-strategic-design.md
cp books/ecommerce-book/src/part1/chapter3.md books/system-design-architecture-book/src/part01/05-internal-architecture-design.md
cp books/ecommerce-book/src/part1/chapter4.md books/system-design-architecture-book/src/part01/06-integration-consistency-design.md
cp books/ecommerce-book/src/part1/chapter5.md books/system-design-architecture-book/src/part01/07-coding-principles-design-patterns.md
```

- [ ] **Step 2: Copy ecommerce reliability chapter into `part03`**

```bash
cp books/ecommerce-book/src/part1/chapter6.md books/system-design-architecture-book/src/part03/16-architecture-quality-assurance.md
```

- [ ] **Step 3: Copy ecommerce practical chapters into `part04`**

```bash
mkdir -p books/system-design-architecture-book/src/part04
cp books/ecommerce-book/src/part2/overview/chapter5.md books/system-design-architecture-book/src/part04/18-ecommerce-overview.md
cp books/ecommerce-book/src/part2/supply/chapter7.md books/system-design-architecture-book/src/part04/19-product-center.md
cp books/ecommerce-book/src/part2/supply/chapter8.md books/system-design-architecture-book/src/part04/20-inventory-system.md
cp books/ecommerce-book/src/part2/supply/chapter9.md books/system-design-architecture-book/src/part04/21-marketing-system.md
cp books/ecommerce-book/src/part2/supply/chapter10.md books/system-design-architecture-book/src/part04/22-product-supply-ops.md
cp books/ecommerce-book/src/part2/transaction/chapter11.md books/system-design-architecture-book/src/part04/23-pricing-system.md
cp books/ecommerce-book/src/part2/transaction/chapter12.md books/system-design-architecture-book/src/part04/24-search-discovery.md
cp books/ecommerce-book/src/part2/transaction/chapter13.md books/system-design-architecture-book/src/part04/25-cart-checkout.md
cp books/ecommerce-book/src/part2/transaction/chapter14.md books/system-design-architecture-book/src/part04/26-order-system.md
cp books/ecommerce-book/src/part2/transaction/chapter15.md books/system-design-architecture-book/src/part04/27-payment-system.md
cp books/ecommerce-book/src/appendix/supplier-sync.md books/system-design-architecture-book/src/part04/28-supplier-sync.md
cp books/ecommerce-book/src/appendix/product-supply-ops.md books/system-design-architecture-book/src/part04/29-product-supply-governance.md
cp books/ecommerce-book/src/part3/chapter16.md books/system-design-architecture-book/src/part04/30-b2b2c-platform-architecture.md
```

- [ ] **Step 4: Copy ecommerce assets and example code**

```bash
cp -R books/ecommerce-book/images books/system-design-architecture-book/images
cp -R books/ecommerce-book/example-codes books/system-design-architecture-book/example-codes
```

### Task 3: Add bridge chapters

**Files:**
- Create: `books/system-design-architecture-book/src/part02/12-global-id-and-basic-services.md`
- Create: `books/system-design-architecture-book/src/part02/13-tech-stack-selection.md`
- Create: `books/system-design-architecture-book/src/part03/17-reconciliation-compensation-dlq.md`
- Create: `books/system-design-architecture-book/src/part03/18-capacity-planning-resilience.md`
- Create: `books/system-design-architecture-book/src/part05/31-system-design-interview-overview.md`
- Create: `books/system-design-architecture-book/src/part05/32-middleware-reliability-interview.md`
- Create: `books/system-design-architecture-book/src/part05/33-ecommerce-architecture-interview.md`
- Create: `books/system-design-architecture-book/src/part05/34-product-inventory-marketing-pricing-interview.md`
- Create: `books/system-design-architecture-book/src/part05/35-search-cart-order-payment-interview.md`
- Create: `books/system-design-architecture-book/src/part05/36-whiteboard-capacity-estimation.md`

- [ ] **Step 1: Reuse ecommerce appendix content where available**

Copy `id-system.md`, `tech-stack.md`, and `interview.md` into chapter positions, then add short bridge introductions if the copied source is too sparse.

- [ ] **Step 2: Add compact original bridge chapters for missing topics**

Write concise Markdown chapters for reconciliation/compensation/DLQ, capacity/resilience, and interview overview topics.

### Task 4: Rewrite metadata and table of contents

**Files:**
- Modify: `books/system-design-architecture-book/book.toml`
- Modify: `books/system-design-architecture-book/src/README.md`
- Modify: `books/system-design-architecture-book/src/SUMMARY.md`

- [ ] **Step 1: Update book metadata**

Set title to `系统设计与架构实战：原理、工程与电商案例`, description to explain the merged scope, and site URL to `/system-design-architecture-book/`.

- [ ] **Step 2: Rewrite `SUMMARY.md`**

List the new six-part structure and all chapter files in reading order.

- [ ] **Step 3: Rewrite `README.md`**

Describe the new book positioning, reading paths, merged content strategy, and local build command.

### Task 5: Verify the new book

**Files:**
- Verify: `books/system-design-architecture-book/`

- [ ] **Step 1: Run mdBook build**

Run:

```bash
cd books/system-design-architecture-book
mdbook build
```

Expected: build exits successfully.

- [ ] **Step 2: Run full site build**

Run:

```bash
npm run clean
npm run build
```

Expected: Hexo build exits successfully.
