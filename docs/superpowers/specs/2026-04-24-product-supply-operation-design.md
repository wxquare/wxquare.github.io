# Product Supply and Operation Design

## Context

Chapter 16.6.1 explains the Product Center for a digital goods platform. Sections 16.6.1.2 to 16.6.1.5 now establish the heterogeneous product model: Category, Resource, SPU, SKU, Offer/Rate Plan, Capability Matrix, Runtime Context, and supplier mappings.

Section 16.6.1.6 currently describes product supply and operation at a high level. It should be redesigned so it connects with the previous data model and also works as interview material for explaining how a senior engineer designs B-side product supply in OTA, O2O, and virtual goods platforms.

## Goal

Rewrite `16.6.1.6 商品供给与运营链路` as a deep design section that:

1. Compares three supply-operation design approaches.
2. Recommends a governed supply platform approach.
3. Connects product supply with Resource, SPU/SKU, Offer/Rate Plan, stock, search, order, and payment consistency.
4. Explains the full supply lifecycle: creation, batch import, supplier sync, operation edit, review, publish, search/cache refresh, event integration, and quality governance.
5. Provides interview-ready summaries and trade-offs.

## Approaches

### Approach A: Lightweight CRUD With Review Flow

This treats supply operation as a back-office CRUD problem: create products, edit fields, approve, publish, and off-shelf.

It is simple and useful for small teams or fixed-SKU categories. It is not enough for a multi-category platform because it does not handle batch scale, supplier sync, high-risk changes, publish consistency, downstream refresh, and data quality feedback loops.

### Approach B: Task-Based Supply Pipeline

This uses `Listing Task` as a unified asynchronous execution unit. Manual upload, batch import, supplier sync, and operation edit are represented as tasks with status, progress, error files, retries, and audit trail.

It is a good implementation baseline, but task execution alone is not sufficient. The system still needs risk review, publishing consistency, event delivery, quality inspection, and compensation.

### Approach C: Governed Supply Platform

This is the recommended design. It builds on the task-based pipeline and adds governance:

```text
Supply Entry
  → Listing Task
  → Standardization and Validation
  → Risk Detection
  → Review and Approval
  → Publishing
  → Search / Cache / Event Refresh
  → Quality Inspection and Compensation
```

The key idea is that product supply is not just CRUD. It is a controlled supply pipeline with idempotency, validation, differentiated review, publish snapshots, downstream events, retries, compensation, and quality monitoring.

## Recommended Section Structure

`16.6.1.6 商品供给与运营链路` should contain:

1. **Why Supply Operation Is Core**
   - Digital goods are not created once and forgotten.
   - Supply comes from manual upload, batch import, supplier sync, and operation edit.
   - Supply must coordinate product master data, resource mapping, offer/rate plan, stock, search, order snapshots, and fulfillment contracts.

2. **Three Approaches**
   - A: Lightweight CRUD with review flow.
   - B: Task-based supply pipeline.
   - C: Governed supply platform, recommended.

3. **Recommended Architecture**
   - Supply entry layer.
   - Listing Task layer.
   - Validation layer.
   - Risk and review layer.
   - Publish layer.
   - Integration/event layer.
   - Quality governance layer.

4. **Four Supply Entry Types**
   - Manual creation: small batch, high quality, strong review.
   - Batch import: large batch, async, progress tracking, error file.
   - Supplier sync: full sync, incremental sync, provider push, idempotent mapping.
   - Operation edit: single item edit, batch edit, high-risk change review.

5. **State Machine**
   - New supply flow:
     `DRAFT → VALIDATING → REVIEWING → APPROVED → PUBLISHING → PUBLISHED`
   - Rejection and failure:
     `VALIDATING/REVIEWING/PUBLISHING → REJECTED/FAILED`
   - Edit flow:
     `PUBLISHED → EDITING → REVIEWING → PUBLISHING → PUBLISHED`
   - Retirement flow:
     `PUBLISHED → OFF_SHELF → ARCHIVED`

6. **Validation and Review**
   - Schema validation.
   - Master-data validation.
   - Product transaction contract validation.
   - Stock and sellability validation.
   - Risk validation.
   - Auto-review, manual review, and high-risk forced review.

7. **Publishing Consistency**
   - Write product master data, resource mapping, offer/rate plan, stock, input schema, fulfillment rule, and refund rule.
   - Generate publish version and product snapshot.
   - Use Outbox for product events.
   - Refresh search index and cache asynchronously.
   - Notify marketing, pricing, order, fulfillment, and data platform.

8. **Quality Governance**
   - Task progress and failure reasons.
   - Product quality reports.
   - Search index failure detection.
   - Cache refresh failure detection.
   - Supplier sync delay.
   - Price and stock anomaly detection.
   - Retry, compensation, alerting, and manual repair.

9. **Interview Summary**
   - Product supply is a governed pipeline, not back-office CRUD.
   - The hard parts are idempotency, validation, risk review, publish consistency, downstream event consistency, and operational observability.

## Scope

This design is for rewriting the book section only. It does not require changing the example code in this step.

## Self Review

- No placeholders remain.
- The design is focused on section 16.6.1.6 only.
- It avoids company-specific names.
- It connects with the already redesigned heterogeneous product model.
- It makes the recommended approach explicit: Governed Supply Platform.
