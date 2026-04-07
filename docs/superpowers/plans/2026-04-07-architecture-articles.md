# Architecture Articles Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a three-article architecture methodology system: deepen 26 (architecture), fill 15 (coding practices), create 27 (checklist).

**Architecture:** Three blog articles with cross-references. 26 provides the architectural "why", 15 provides the coding "how", 27 distills both into an actionable review checklist organized by review stage.

**Tech Stack:** Hexo 7.2.0, Markdown, Mermaid diagrams, Go code examples.

**Spec:** `docs/superpowers/specs/2026-04-07-architecture-articles-design.md`

**Build verify command:** `npm run clean && npm run build` (must pass after every task)

---

## Phase 1: Deepen 26-clean-architecture-ddd-cqrs.md

### Task 1: Deepen Chapter 1 — Clean Architecture (+200 lines)

**Files:**
- Modify: `source/_posts/system-design/26-clean-architecture-ddd-cqrs.md` (append after §1.4)

- [ ] **Step 1: Add §1.5 Architecture Style Comparison**

Insert after the existing §1.4 section (after line 105 `\`\`\``). Add a new subsection comparing Clean Architecture, Hexagonal Architecture, and Onion Architecture:

```markdown
### 1.5 架构风格对比：Clean vs 六边形 vs 洋葱

三种架构风格经常被混用，它们的核心共识都是**依赖反转**，但侧重点不同：

| 维度 | Clean Architecture | 六边形架构 (Hexagonal) | 洋葱架构 (Onion) |
|------|-------------------|----------------------|-----------------|
| 提出者 | Robert C. Martin | Alistair Cockburn | Jeffrey Palermo |
| 核心隐喻 | 同心圆（层层向内） | 六边形端口与适配器 | 洋葱层层剥开 |
| 关键概念 | Entity, Use Case, Adapter | Port, Adapter | Domain Model, Domain Service |
| 外部交互 | 通过 Interface Adapter 层 | 通过 Port（接口）+ Adapter（实现） | 通过 Infrastructure 层 |
| 核心共识 | **依赖方向向内，业务逻辑不依赖外部技术** | 同左 | 同左 |
```

Include a Mermaid diagram showing the three architectures side by side, and a Go code example demonstrating the Port & Adapter pattern.

- [ ] **Step 2: Add §1.6 Dependency Injection in Go**

Add subsection covering:
- Manual DI via constructor injection (recommended for small-medium projects)
- Wire-based DI (for large projects with many dependencies)
- Complete Go code example showing both approaches with the same OrderService

```go
// Manual DI
func NewOrderService(repo domain.OrderRepository, bus EventBus) *OrderService {
    return &OrderService{repo: repo, bus: bus}
}

// In main.go
repo := mysql.NewOrderRepository(db)
bus := kafka.NewEventBus(producer)
svc := NewOrderService(repo, bus)
```

- [ ] **Step 3: Add §1.7 Anti-patterns**

Add subsection with 3 concrete anti-pattern examples in Go:
1. Handler directly importing repository implementation (bypassing domain layer)
2. Circular dependency between domain and infrastructure
3. Infrastructure types leaking into domain layer (e.g., `sql.NullString` in entity)

Each with: problematic code → explanation → corrected code.

- [ ] **Step 4: Build verify**

```bash
cd /Users/xianguiwang/gopath/src/github.com/wxquare/wxquare.github.io
npm run clean && npm run build
```

Expected: Build succeeds with no errors.

- [ ] **Step 5: Commit**

```bash
git add source/_posts/system-design/26-clean-architecture-ddd-cqrs.md
git commit -m "content(26): deepen Clean Architecture — style comparison, DI, anti-patterns"
```

---

### Task 2: Deepen Chapter 2 — DDD (+250 lines)

**Files:**
- Modify: `source/_posts/system-design/26-clean-architecture-ddd-cqrs.md` (append after §2.4)

- [ ] **Step 1: Add §2.5 Aggregate Design Principles**

Add subsection covering:
- Rule 1: One transaction modifies only one aggregate
- Rule 2: Inter-aggregate communication via domain events
- Rule 3: Small aggregate vs large aggregate trade-offs
- Decision table: when to merge vs split aggregates
- Go code example: Order aggregate publishing event → Inventory aggregate consuming event

- [ ] **Step 2: Add §2.6 Repository Deep Dive — Unit of Work**

Add subsection covering:
- Why generic CRUD repositories are an anti-pattern
- Unit of Work pattern in Go (transaction boundary management)
- Complete Go implementation:

```go
type UnitOfWork interface {
    OrderRepo() OrderRepository
    InventoryRepo() InventoryRepository
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
}
```

- [ ] **Step 3: Add §2.7 Domain Event Async — Outbox Pattern**

Add subsection covering:
- Problem: dual-write between DB and message bus
- Solution: Outbox Pattern (local transaction table + async relay)
- Complete Go implementation:
  - Outbox table schema
  - Save aggregate + outbox entry in single transaction
  - Relay goroutine that polls outbox and publishes to Kafka
- Mermaid sequence diagram showing the flow

- [ ] **Step 4: Build verify + commit**

```bash
npm run clean && npm run build
git add source/_posts/system-design/26-clean-architecture-ddd-cqrs.md
git commit -m "content(26): deepen DDD — aggregate principles, UoW, Outbox Pattern"
```

---

### Task 3: Deepen Chapter 3 — CQRS (+200 lines)

**Files:**
- Modify: `source/_posts/system-design/26-clean-architecture-ddd-cqrs.md` (append after §3.4)

- [ ] **Step 1: Add §3.5 Event Sourcing**

Add subsection covering:
- Concept: store events instead of state
- Relationship with CQRS (optional combination, not required)
- When to use: audit trails, financial systems, undo/replay
- When NOT to use: simple CRUD, high-frequency updates
- Go code snippet showing event store interface

- [ ] **Step 2: Add §3.6 Eventual Consistency Strategies**

Add subsection covering:
- Compensation transactions (Saga pattern brief)
- Idempotent design (idempotency key pattern)
- UX handling of read model delay (optimistic UI, polling, webhooks)
- Consistency window estimation

- [ ] **Step 3: Add §3.7 Projector Implementation**

Add subsection with complete Go code:
- Projector interface definition
- OrderProjector implementation (event → read model upsert)
- Error handling and retry in projector
- Mermaid diagram showing event flow through projector

- [ ] **Step 4: Build verify + commit**

```bash
npm run clean && npm run build
git add source/_posts/system-design/26-clean-architecture-ddd-cqrs.md
git commit -m "content(26): deepen CQRS — Event Sourcing, eventual consistency, projector"
```

---

### Task 4: Deepen Chapters 4-5 + New Chapter 6 (+450 lines)

**Files:**
- Modify: `source/_posts/system-design/26-clean-architecture-ddd-cqrs.md`

- [ ] **Step 1: Add §4.4 Complete Walk-through**

Insert after existing §4.3. A full code-level walkthrough of an e-commerce order request:
- HTTP request → Handler (Adapter layer)
- Handler → PlaceOrderCommand (Application layer)
- CommandHandler → Order.AddItem() + Order.Place() (Domain layer)
- Save to write DB + Publish OrderPlacedEvent
- Projector receives event → Upsert read model
- GET request → QueryHandler → Read DB → DTO response
- Each step annotated with: `[Clean Arch layer] [DDD concept] [CQRS path]`

- [ ] **Step 2: Deepen §5 — Add overengineering detection + team assessment**

Append to existing Chapter 5:
- §5.3 Overengineering Detection: "If your aggregate has only CRUD, you don't need DDD" decision tree
- §5.4 Team Capability Assessment: tech debt vs architecture investment balance

- [ ] **Step 3: Add new Chapter 6 — Progressive Adoption Guide**

Add as new chapter before the existing 总结 chapter. Four evolution stages:
- Stage 0: Standard 3-layer (Handler → Service → Repository)
- Stage 1: + Clean Architecture (introduce interface layer, dependency inversion)
- Stage 2: + DDD (extract aggregate roots, value objects, domain events)
- Stage 3: + CQRS (separate Command/Query handlers, introduce read model)

Each stage includes:
- Trigger condition (when to evolve)
- Before/after directory structure
- Before/after Go code comparison
- Risk control measures
- Mermaid flowchart showing evolution decision tree

- [ ] **Step 4: Renumber existing chapters**

The current "六、总结" becomes "七、总结". Update all references accordingly.

- [ ] **Step 5: Build verify + commit**

```bash
npm run clean && npm run build
git add source/_posts/system-design/26-clean-architecture-ddd-cqrs.md
git commit -m "content(26): add walk-through, overengineering detection, progressive adoption guide"
```

---

## Phase 2: Fill 15-clean-code.md Empty Chapters

### Task 5: Write Chapter 10 — Refactoring Cases (~400 lines)

**Files:**
- Modify: `source/_posts/system-design/15-clean-code.md` (replace empty §10 skeleton at lines 6373-6377)

- [ ] **Step 1: Write §10.1 — Thousand-line Function → Pipeline**

Replace the empty `### 10.1` heading. Write complete case study:
- **Before**: Show a condensed 80-line version of the `CreateOrder` monster function (representing a 1500-line original) with inline comments marking the 12 steps
- **Problem analysis**: Identify violations (SRP, readability, testability). Include metrics: cyclomatic complexity = 45, test coverage = 15%
- **Refactoring steps**: 4 steps with code at each stage:
  1. Identify step boundaries → list of 12 processors
  2. Define `OrderContext` struct
  3. Extract each step into a Processor implementing `func(ctx *OrderContext) error`
  4. Assemble Pipeline with `NewPipeline(processors...)`
- **After**: Complete refactored code showing Pipeline assembly + one sample Processor
- **Effect**: Table comparing before/after metrics

- [ ] **Step 2: Write §10.2 — if-else Hell → Strategy Pattern**

Replace the empty `### 10.2` heading. Write complete case study:
- **Before**: 15-branch price calculation function (~60 lines showing the pattern)
- **Problem analysis**: OCP violation, every new category requires modifying the function
- **Refactoring steps**:
  1. Define `PriceCalculator` interface
  2. Implement TopupCalculator, HotelCalculator, FlightCalculator
  3. Build registry map: `map[CategoryID]PriceCalculator`
  4. Replace if-else with registry lookup
- **After**: Complete code showing registry + one calculator
- **Effect**: "Adding a new category = 1 new file, 1 line in registry"

- [ ] **Step 3: Write §10.3 — Context Explosion → Context Pattern**

Replace the empty `### 10.3` heading. Write complete case study:
- **Before**: Function chain with 12 parameters passed through 5 levels
- **Refactoring steps**:
  1. Group parameters into Input/Intermediate/Output
  2. Define `ProcessContext` struct
  3. Refactor each function to accept `*ProcessContext`
- **After**: Clean function signatures with single context parameter
- **Effect**: Parameter count 12 → 1, adding new intermediate data = add one field

- [ ] **Step 4: Build verify + commit**

```bash
npm run clean && npm run build
git add source/_posts/system-design/15-clean-code.md
git commit -m "content(15): write chapter 10 — three complete refactoring case studies"
```

---

### Task 6: Write Chapter 11 — Performance & Monitoring (~300 lines)

**Files:**
- Modify: `source/_posts/system-design/15-clean-code.md` (replace empty §11 skeleton at lines 6381-6384)

- [ ] **Step 1: Write §11.1 — Performance Optimization Strategies**

Replace the empty `### 11.1` heading. Cover 4 strategies with Go code:
1. **sync.Pool in Pipeline**: Reuse ProcessContext objects to reduce GC pressure. Show pool creation, Get/Put lifecycle, benchmark comparison.
2. **Parallel Stage execution**: Fan-out/fan-in pattern for independent processors. Show `errgroup` based implementation.
3. **Timeout control & graceful degradation**: `context.WithTimeout` wrapping each stage. Show fallback behavior.
4. **Zero-copy serialization**: Brief comparison of encoding/json vs protobuf vs flatbuffers for DTO serialization.

- [ ] **Step 2: Write §11.2 — Monitoring & Observability**

Replace the empty `### 11.2` heading. Cover with Go code:
1. **Metrics**: Pipeline stage duration histogram, success/failure counter. Show Prometheus client integration.
2. **Distributed Trace**: OpenTelemetry span per Pipeline stage. Show tracer setup + span creation.
3. **Structured Logging**: slog best practices. Show log format with trace_id, stage_name, duration.
4. **Alert Rules**: Table of key alerts (stage latency P99 > 500ms, error rate > 5%, etc.)

- [ ] **Step 3: Build verify + commit**

```bash
npm run clean && npm run build
git add source/_posts/system-design/15-clean-code.md
git commit -m "content(15): write chapter 11 — performance optimization and observability"
```

---

### Task 7: Write Chapter 12 — Team Adoption (~250 lines)

**Files:**
- Modify: `source/_posts/system-design/15-clean-code.md` (replace empty §12 skeleton at lines 6388-6392)

- [ ] **Step 1: Write §12.1 — Code Review Checklist (compact version)**

Replace the empty `### 12.1` heading. Write a 10-item core checklist:
- 5 coding-level checks (function length, naming, error handling, SOLID, nesting depth)
- 5 design-level checks (dependency direction, aggregate boundary, separation of concerns, pattern fit, testability)
- Add cross-reference: "完整版检查清单见 [27-架构与编码 Code Review Checklist](/system-design/27-architecture-checklist/)"

- [ ] **Step 2: Write §12.2 — Convincing the Team to Refactor**

Replace the empty `### 12.2` heading. Cover:
- ROI quantification method: defect density, change cost, onboarding time (with formula and example numbers)
- Boy Scout Rule: incremental improvement per PR
- Data-driven persuasion: before/after bug rate comparison table

- [ ] **Step 3: Write §12.3 — Refactoring Risk Control**

Replace the empty `### 12.3` heading. Cover:
- Feature flag strategy with Go code example
- Canary deployment: old logic as fallback, new logic gradually rolled out
- Test coverage gate: refactored code must > 80%
- Rollback procedure checklist

- [ ] **Step 4: Build verify + commit**

```bash
npm run clean && npm run build
git add source/_posts/system-design/15-clean-code.md
git commit -m "content(15): write chapter 12 — team adoption, review checklist, risk control"
```

---

### Task 8: Write Chapter 13 — Summary & Outlook (~150 lines)

**Files:**
- Modify: `source/_posts/system-design/15-clean-code.md` (replace empty §13 skeleton at lines 6396-6400)

- [ ] **Step 1: Write §13.1 — Key Takeaways**

Replace the empty `### 13.1` heading. One-sentence summary per chapter (chapters 1-12), formatted as a quick reference table:

| 章节 | 核心要点 |
|------|----------|
| 一、痛点画像 | 复杂业务代码腐化的 5 个根因 |
| ... | ... |

- [ ] **Step 2: Write §13.2 — Learning Path**

Replace the empty `### 13.2` heading. Three tiers:
- Junior (0-2 years): naming → function decomposition → error handling. Books: Clean Code
- Mid (2-5 years): SOLID → design patterns → Pipeline. Books: Design Patterns, Refactoring
- Senior (5+ years): DDD → Clean Architecture → CQRS. Books: DDD (Evans), Clean Architecture (Martin)

Include a Mermaid flowchart showing the learning progression.

- [ ] **Step 3: Write §13.3 — From Clean Code to Clean Architecture**

Replace the empty `### 13.3` heading (note: rename from "参考资料" to avoid clash with existing 参考资料 section). Cover:
- Cognitive upgrade path: code-level → module-level → system-level
- Bridge to article 26: "Clean Code 是地基，Clean Architecture 是框架"
- Link: "深入了解架构方法论，请阅读 [Clean Architecture、DDD 与 CQRS](/system-design/26-clean-architecture-ddd-cqrs/)"

- [ ] **Step 4: Build verify + commit**

```bash
npm run clean && npm run build
git add source/_posts/system-design/15-clean-code.md
git commit -m "content(15): write chapter 13 — summary, learning path, bridge to architecture"
```

---

## Phase 3: Create 27-architecture-checklist.md

### Task 9: Create Article + Intro + Chapter 1 — Architecture Review (~300 lines)

**Files:**
- Create: `source/_posts/system-design/27-architecture-checklist.md`

- [ ] **Step 1: Create file with Front Matter + Introduction**

Create the file with:

```yaml
---
title: 架构与编码 Code Review Checklist
date: 2026-04-07
categories:
- 系统设计
tags:
- code-review
- 架构设计
- checklist
- clean-architecture
- ddd
- cqrs
- clean-code
toc: true
---
```

Introduction covering:
- Why systematic review (reference: "好的代码是重构出来的，不是一次写出来的")
- How to use this checklist (walk through stages sequentially)
- Relationship to articles 15 and 26 (this = what to check; those = why and how)
- Mermaid diagram showing the 4 review stages

- [ ] **Step 2: Write Chapter 1 — Architecture Review Stage**

6 checklist items for the design phase. Each item follows the format:

```markdown
- [ ] **1.1 分层结构**
  - **标准**：是否定义了 Domain / Application / Adapter / Infra 四层？依赖方向是否只向内？
  - **反例**：Handler 层直接 import 了 MySQL 包；Domain 层引用了 Redis client
  - **参考**：→ [26-架构方法论 §1 Clean Architecture](/system-design/26-clean-architecture-ddd-cqrs/#一、Clean-Architecture)
```

Items: 分层结构, Bounded Context 划分, 聚合边界, 读写路径评估, 技术选型, 过度设计检查

- [ ] **Step 3: Build verify + commit**

```bash
npm run clean && npm run build
git add source/_posts/system-design/27-architecture-checklist.md
git commit -m "content(27): create architecture checklist — intro + architecture review stage"
```

---

### Task 10: Write Chapter 2 — Design Review Stage (~250 lines)

**Files:**
- Modify: `source/_posts/system-design/27-architecture-checklist.md`

- [ ] **Step 1: Write Chapter 2 — Design Review checklist items**

7 checklist items for the detailed design phase, same format as Chapter 1. Items:
1. 聚合根识别
2. 实体 vs 值对象
3. Repository 接口设计
4. Command 设计
5. Query 设计
6. 领域事件定义
7. 模式选型（Pipeline / 策略 / 规则引擎 decision table）

Each with 标准 + 反例 + 参考 cross-reference to 15 or 26.

The 模式选型 item should include a decision table:

```markdown
| 场景特征 | 推荐模式 | 参考 |
|----------|----------|------|
| 多步骤顺序流程 | Pipeline | → 15 §4 |
| 同一接口多种实现 | 策略模式 | → 15 §6.1 |
| 频繁变化的业务规则 | 规则引擎 | → 15 §7 |
| 跨聚合协作 | 领域事件 | → 26 §2.7 |
```

- [ ] **Step 2: Build verify + commit**

```bash
npm run clean && npm run build
git add source/_posts/system-design/27-architecture-checklist.md
git commit -m "content(27): write design review stage — 7 checklist items"
```

---

### Task 11: Write Chapter 3 — Code Review Stage (~400 lines)

**Files:**
- Modify: `source/_posts/system-design/27-architecture-checklist.md`

- [ ] **Step 1: Write §3.1 SOLID Principles checklist**

5 items (SRP, OCP, LSP, ISP, DIP), each with:
- 标准: concrete Go-specific check question
- 反例: Go code snippet showing violation
- 正例: Go code snippet showing compliance
- 参考: → 15 §3.1

- [ ] **Step 2: Write §3.2 Function Quality checklist**

4 items (length < 80 lines, cyclomatic complexity < 10, nesting < 3, params < 5).
Include specific Go tooling commands to check:
- `gocyclo` for cyclomatic complexity
- `gocognit` for cognitive complexity

- [ ] **Step 3: Write §3.3-3.4 Naming + Error Handling checklist**

Naming items (3): business terminology, ubiquitous language consistency, anti-pattern detection (SetStatus vs Place).
Error handling items (3): no ignored errors, wrapped errors with context, business vs system error separation.

- [ ] **Step 4: Write §3.5-3.6 Dependency Direction + DDD Tactical checklist**

Dependency items (3): domain imports clean, application only depends on domain, no circular deps. Include `go vet` / import analysis commands.
DDD items (4): aggregate root protects invariants, value objects immutable, no external mutation of aggregate internals, one transaction one aggregate.

- [ ] **Step 5: Build verify + commit**

```bash
npm run clean && npm run build
git add source/_posts/system-design/27-architecture-checklist.md
git commit -m "content(27): write code review stage — SOLID, function quality, naming, DDD checks"
```

---

### Task 12: Write Chapter 4 + Appendix — Pre-merge + Quick Reference (~250 lines)

**Files:**
- Modify: `source/_posts/system-design/27-architecture-checklist.md`

- [ ] **Step 1: Write Chapter 4 — Pre-merge Checklist**

6 items for the merge phase:
1. Performance (benchmark existence, goroutine/channel leak risk)
2. Concurrency safety (mutex, atomic, concurrent data structures)
3. Observability (metrics, trace, structured logs)
4. Test coverage (core logic > 80%, integration tests exist)
5. Rollback plan (feature flag, DB migration reversibility)
6. Documentation (architecture changes documented)

- [ ] **Step 2: Write Appendix — Quick Reference Card**

One-page summary: 5 most critical items per stage (total 20 items), formatted as a compact table:

```markdown
| 阶段 | Top 5 检查项 |
|------|-------------|
| 架构评审 | 1. 依赖向内 2. BC 划分 3. 聚合边界 4. 读写评估 5. YAGNI |
| 设计评审 | ... |
| 代码评审 | ... |
| 上线前 | ... |
```

- [ ] **Step 3: Write 参考资料 section**

Add references:
- Cross-links to articles 15 and 26
- External references: Clean Architecture book, DDD book, CQRS pattern docs

- [ ] **Step 4: Build verify + commit**

```bash
npm run clean && npm run build
git add source/_posts/system-design/27-architecture-checklist.md
git commit -m "content(27): write pre-merge checks, quick reference card, references"
```

---

## Phase 4: Final Verification

### Task 13: Cross-reference Verification + Final Build

**Files:**
- Read: all three articles
- Modify: any broken cross-references

- [ ] **Step 1: Verify all cross-references**

Check that all `→ 15 §X` and `→ 26 §Y` references in article 27 point to actual existing sections. Check that article 15 §12.1 links to article 27. Check that article 15 §13.3 links to article 26.

- [ ] **Step 2: Verify internal Hexo links**

All cross-article links should use Hexo path format: `/system-design/XX-name/`. Verify these will resolve correctly.

- [ ] **Step 3: Final build**

```bash
npm run clean && npm run build
```

Expected: Build succeeds. All three articles generate HTML.

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "content: finalize architecture article trilogy — cross-references verified"
```
