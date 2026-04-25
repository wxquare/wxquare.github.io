# Product Supply Operation Section Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite `16.6.1.6 商品供给与运营链路` into a deep, book-ready and interview-ready section based on the governed supply platform design.

**Architecture:** This is a documentation-only change. The new section will compare three supply-operation approaches, recommend the governed supply platform, and connect supply flow with Resource, SPU/SKU, Offer/Rate Plan, stock, search, order, fulfillment, and events.

**Tech Stack:** Markdown, Mermaid-compatible code blocks, Hexo build verification.

---

### Task 1: Replace Section 16.6.1.6

**Files:**
- Modify: `ecommerce-book/src/part3/chapter16.md`

- [ ] **Step 1: Locate the exact section**

Run:

```bash
rg -n "#### 16\\.6\\.1\\.6|#### 16\\.6\\.1\\.7" ecommerce-book/src/part3/chapter16.md
```

Expected: shows the start line for `16.6.1.6` and the next section `16.6.1.7`.

- [ ] **Step 2: Replace the current short section**

Use `apply_patch` to replace the content from `#### 16.6.1.6 商品供给与运营链路` up to, but not including, `#### 16.6.1.7 供应商商品同步链路`.

The replacement must include:

```text
1. 为什么供给运营链路是核心能力
2. 三种方案对比：CRUD、任务化流水线、供给治理平台
3. 推荐方案：供给治理平台
4. 四类供给入口：人工创建、批量导入、供应商同步、运营编辑
5. Listing Task 状态机
6. 校验与审核设计
7. 发布一致性设计
8. 异常治理与可观测性
9. 面试总结
```

- [ ] **Step 3: Verify no company-sensitive words were introduced**

Run:

```bash
rg -n "Shopee|ShopeePay|Hub|DP|dp-admin|airpay" ecommerce-book/src/part3/chapter16.md
```

Expected: no new matches in the rewritten `16.6.1.6` section.

- [ ] **Step 4: Verify Markdown build**

Run:

```bash
npm run clean
npm run build
```

Expected: both commands succeed.

- [ ] **Step 5: Report changed section and verification**

Final response should include:

```text
Updated section path and line number.
Build commands executed.
Any remaining unrelated dirty files were not touched.
```

## Self Review

- The plan covers the approved spec.
- No placeholders remain.
- The scope is a single documentation section.
