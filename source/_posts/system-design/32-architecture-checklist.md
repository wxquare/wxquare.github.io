---
title: 架构与编码 Code Review Checklist
date: 2026-04-07
categories:
  - 系统设计基础
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

## 引言

软件工程里有一句常被引用的话：**好的代码是重构出来的，不是一次写出来的**。它提醒我们：初稿几乎必然欠打磨，真正可靠的质量来自持续、有纪律的迭代。Code Review 正是这种迭代中最关键的一环——它把个人习惯拉平到团队标准，把隐性知识显性化，把缺陷拦截在合并之前。

然而，「随便看看」式的评审往往流于表面：有人只看风格，有人只看有没有明显 bug，有人被 diff 的噪声淹没。结果是：架构层面的失误晚到无法廉价修正，设计层面的模糊在代码里被放大成技术债，上线前才发现性能或可观测性缺口。要对抗这种随机性，需要**分阶段、可重复的 Checklist**：在正确的时机问正确的问题。

本文提供一套按评审阶段组织的清单，建议你**按顺序走完四个阶段**，而不是在单次 PR 里眉毛胡子一把抓：

1. **架构评审**：新项目或新模块启动时，先确认分层、边界、读写路径与技术选型是否站得住脚。
2. **设计评审**：详设与接口冻结阶段，检查聚合、命令查询、事件与模式选型是否与领域一致。
3. **代码评审**：日常 PR，用 SOLID、函数质量、命名、错误处理与依赖方向守住实现细节。
4. **上线前检查**：合并发布窗口，补齐性能、并发、可观测性、测试、回滚与文档。

### 三篇文章如何配合使用

本仓库里与「架构 + 编码」相关的文章可以形成一条学习与实践链路：

| 文章 | 角色 |
|------|------|
| **本文（27）** | **查什么**：各阶段 Review 要问什么、反例长什么样 |
| [复杂业务中的 Clean Code 实践指南](/system-design/31-clean-code/)（15） | **怎么写**：函数、Pipeline、策略、规则引擎等战术 |
| [Clean Architecture、DDD 与 CQRS：三位一体的架构方法论](/system-design/30-clean-architecture-ddd-cqrs/)（26） | **怎么设计**：分层、BC、聚合、CQRS、事件与反模式 |

读 26 建立地图，读 15 练手法，读 27 在评审时逐项打勾——三者互为索引，而不是重复堆砌。

从心理学角度，Checklist 的价值在于**降低认知负荷**：评审者在疲劳、时间压力或上下文切换时，仍有一个外部脚手架防止遗漏。它并不替代经验与判断力——遇到清单未覆盖的灰区，恰恰说明团队应该把新教训**反哺**进清单或 ADR。实践中建议：

- **责任人明确**：架构项由 Tech Lead / 架构负责人主评；设计项由领域 Owner 主评；PR 项由作者与至少一名熟悉该域的审阅者共担；上线前项可与 SRE / On-call 对齐。
- **粒度分层**：巨型 MR 可先要求作者附「自审清单」勾选说明，再在评论里对争议点逐条引用本文章节编号，避免无结构的「感觉不对」。
- **与工具链结合**：复杂度、静态检查、依赖图、覆盖率门槛应作为**门禁**，清单作为**人工语义层**补充（例如：覆盖率够了但测的是 happy path，仍需人眼过业务不变量）。

### 四阶段评审流程（Mermaid）

下面是一张简化的流程图，表示从设计期到合并期的顺序关系（实际项目可在各阶段间迭代，但**问题域**应分开讨论，避免在代码 diff 里硬掰架构决策）。

```mermaid
flowchart LR
  A[架构评审<br/>设计期] --> B[设计评审<br/>详设期]
  B --> C[代码评审<br/>PR 期]
  C --> D[上线前检查<br/>合并期]
  D --> E[发布 / 观测 / 复盘]
```

**使用建议**：

- 架构、设计阶段的结论最好有**可追溯记录**（ADR、RFC 或设计文档），Code Review 时只核对「实现是否背离结论」。
- PR 评论里若发现架构级问题，应**上升**到设计讨论，而不是在局部 hack 里「修掉症状」。
- Checklist 是**最小充分集**的启发工具，团队可按域（支付、搜索、实时链路）扩展专属条目，但不要删掉「依赖方向」「聚合边界」这类高杠杆项。

**与「好的代码是重构出来的」的关系**：清单并不是鼓吹「一次设计完美」，而是规定**在哪些关口必须重构**：当架构评审发现分层倒置，应允许推翻局部实现；当代码评审发现函数失控，应要求拆分而不是堆注释。重构被嵌入流程，而不是留到「有空再说」。

---

## 一、架构评审阶段 — 设计期

**适用时机**：立项、新服务、新子域或大规模模块拆分。目标是在写大量代码之前，把**分层、边界、一致性、读写特征与技术选型**对齐。

### 1. 分层结构

**标准**：是否明确定义 **Domain / Application / Adapter / Infrastructure**（或等价四层）？源代码依赖是否**一律指向内层**（Domain 为最内），外层通过接口向内依赖？

**反例（违反依赖方向）**：HTTP Handler 直接 `import` 具体 MySQL 驱动或 ORM 包，绕过应用服务与领域端口。

```go
// BAD: handler depends on concrete DB package
import "github.com/org/repo/infra/mysql"

func HandlePlaceOrder(w http.ResponseWriter, r *http.Request) {
    db := mysql.Default()
    _, _ = db.ExecContext(r.Context(), "INSERT INTO orders ...")
}
```

**合规方向**：Handler 只依赖应用层用例；持久化通过 **Repository 接口**在领域或应用边界声明，由 Infra 实现。

```go
// GOOD: handler -> application port -> domain; infra implements port
type PlaceOrderHandler struct {
    App *application.OrderService
}

func (h *PlaceOrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    cmd, err := decodePlaceOrder(r)
    if err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    if err := h.App.PlaceOrder(r.Context(), cmd); err != nil {
        // map domain/app errors to HTTP
        http.Error(w, err.Error(), http.StatusConflict)
        return
    }
    w.WriteHeader(http.StatusCreated)
}
```

**评审追问**：若团队暂时未引入完整四层，是否至少在包级约定 **adapter 不得被 domain import**，并在 CI 用 `grep` / 自定义 linter 守护？

**参考**：[Clean Architecture、DDD 与 CQRS：三位一体的架构方法论](/system-design/30-clean-architecture-ddd-cqrs/) §1（分层与依赖规则）。

---

### 2. Bounded Context 划分

**标准**：是否识别 **核心域、支撑域、通用域**？每个 BC 是否有清晰的 **Ubiquitous Language** 与对外契约（API / 事件），避免「一个大而全的领域模型」?

**反例**：订单子域与库存子域共用同一个 `Product` 结构体，字段含义在两边互相拉扯（价格、可售库存、展示属性混在同一类型上）。

```go
// BAD: one struct serves two contexts with conflicting meanings
type Product struct {
    ID            string
    Title         string
    PriceCent     int64  // pricing in order context
    WarehouseQty  int    // stock in inventory context — coupling contexts
}
```

**合规 sketch**：不同 BC 使用不同模型与防腐层翻译；集成通过 API、消息或显式 ACL。

```go
// GOOD: separate models + explicit mapping at boundary
type catalog.ProductView struct { ID, Title string }

type ordering.OrderLine struct {
    ProductID   string
    UnitPrice   Money
    SnapshotTitle string // captured at order time, not live catalog coupling
}

type inventory.StockUnit struct {
    SKU string
    OnHand int
}
```

**评审追问**：若两个 BC 必须共享标识符，是共享 **ID** 还是共享 **富模型**？前者常见且可接受，后者往往是边界溃缩的信号。

**参考**：[30-架构方法论](/system-design/30-clean-architecture-ddd-cqrs/) §2.1（限界上下文）。

---

### 3. 聚合边界

**标准**：**一致性边界**是否以聚合为单位设计？是否避免在单个事务中强行修改多个聚合根，除非有显式的领域规则与补偿策略？

**反例**：一个数据库事务内同时更新 `Order` 与 `Inventory` 聚合，绕过领域事件与最终一致性，导致锁竞争与模型腐化。

```go
// BAD: one transaction mutates two aggregates directly
func SaveOrderAndDeductStock(ctx context.Context, tx *sql.Tx, o *Order, inv *Inventory) error {
    if err := persistOrder(tx, o); err != nil {
        return err
    }
    inv.Quantity -= o.LineItems[0].Qty // cross-aggregate invariant hidden in application glue
    return persistInventory(tx, inv)
}
```

**参考**：[30-架构方法论](/system-design/30-clean-architecture-ddd-cqrs/) §2.5（聚合）。

---

### 4. 读写路径评估

**标准**：是否量化 **读写比**、延迟与一致性要求？读路径若存在重 JOIN、宽表、复杂筛选，是否考虑 **独立读模型 / 投影**，而不是全部堆在写模型上？

**反例**：在命令路径（下单）同步执行多表 JOIN 报表查询，拖慢写入尾延迟。

**合规 sketch**：写路径只持久化命令所需最小一致性数据；读路径走物化视图、搜索索引或专用查询服务。

```go
// BAD: command handler does heavy read for side UI
func (s *OrderService) PlaceOrder(ctx context.Context, cmd PlaceOrderCommand) error {
    _ = s.db.QueryRowContext(ctx, `
        SELECT ... heavy join for dashboard ...
    `)
    return s.persistOrder(ctx, cmd)
}

// GOOD: split; async projection or query DB
func (s *OrderService) PlaceOrder(ctx context.Context, cmd PlaceOrderCommand) error {
    if err := s.orders.Save(ctx, newOrderFrom(cmd)); err != nil {
        return err
    }
    return s.outbox.Publish(ctx, OrderPlaced{OrderID: cmd.IdempotencyKey})
}
```

**评审追问**：是否测量过 **p99 写延迟** 与 **读 QPS**？若读是写的两个数量级以上，独立读模型往往是经济解。

**参考**：[30-架构方法论](/system-design/30-clean-architecture-ddd-cqrs/) §3（CQRS 与读写分离）。

---

### 5. 技术选型

**标准**：存储与中间件是否与 **访问模式** 匹配（点查、范围扫、全文检索、图关系、流处理）？是否记录选型假设与回退方案？

**反例**：全文搜索需求用 `MySQL LIKE '%keyword%'` 扛流量，缺少倒排索引与相关性能力。

**评审追问**：选型表是否包含 **数据量预估、热点键、一致性级别、运维成本**？是否评估过 **多租户**、**合规留存**、**跨地域** 对存储的影响？

**合规**：为每种访问模式写清「主存储 + 缓存 + 索引」的职责划分，避免所有读都打到 OLTP。

---

### 6. 过度设计检查（YAGNI）

**标准**：是否仅为**已确认**的变更点引入抽象？能否用更简单的模型先交付，再演化？

**反例**：典型 CRUD 后台强行上 **DDD + CQRS + Event Sourcing** 全家桶，团队无力维护投影与版本化事件。

**评审追问**：若去掉 Event Sourcing，业务是否仍成立？若答案是肯定的，则 ES 很可能是 **可选优化** 而非当前必需。同理，CQRS 是否由**观测到的读写不对称**驱动，而不是由「流行架构标签」驱动？

**合规**：从 **Transaction Script + 清晰模块边界** 起步，在出现明确痛点时再引入战术模式；每引入一层，同步引入 **测试与运维** 能力。

**参考**：[30-架构方法论](/system-design/30-clean-architecture-ddd-cqrs/) §5.3（反模式与 YAGNI）。

---

## 二、设计评审阶段 — 详设期

**适用时机**：接口评审、领域模型评审、用例与事件清单冻结前。目标是让 **战术设计**（聚合、Repo、Command/Query、事件）与战略分层一致。

### 1. 聚合根识别

**标准**：**聚合根**是否是外部访问聚合内对象的**唯一入口**？外部代码是否禁止绕过根直接改内部实体状态？

**反例**：`OrderLine` 在包外被直接修改数量，绕过 `Order` 上的库存与金额不变量。

```go
// BAD: line exported and mutated from outside aggregate
type Order struct {
    ID    string
    Lines []*OrderLine // exported slice of mutable lines
}
type OrderLine struct {
    SKU string
    Qty int
}

func SomeHandler() {
    o := &Order{ID: "1", Lines: []*OrderLine{{SKU: "A", Qty: 1}}}
    o.Lines[0].Qty = 999 // invariant broken: no route through Order root
}
```

**合规 sketch**：通过 `Order` 的方法修改行项目，并在方法内校验不变量。

```go
// GOOD: changes go through aggregate root
func (o *Order) ChangeLineQty(sku string, qty int) error {
    if qty < 0 {
        return ErrInvalidQty
    }
    // find line, recompute totals, enforce rules
    return nil
}
```

**评审追问**：若聚合根方法数量爆炸，是 **聚合过大** 还是 **缺少领域服务**？前者考虑拆分聚合与事件协作，后者提取无状态领域服务协调多个根（仍遵守一事务一根的默认）。

**反例补充**：将 `OrderLine` 作为独立聚合根对外暴露 CRUD API，导致订单总额不变量无法封闭。

---

### 2. 实体 vs 值对象

**标准**：**实体**是否有稳定标识且可变（通过受控方法）？**值对象**是否**不可变**、按值语义相等（而非仅按指针）？

**反例**：`Money` 提供 `SetAmount`，被多处共享引用后产生意外修改。

```go
// BAD: value object mutable
type Money struct {
    Currency string
    Amount   int64
}
func (m *Money) SetAmount(a int64) { m.Amount = a }
```

```go
// GOOD: new value instead of mutating
type Money struct {
    currency string
    amount   int64
}
func (m Money) Add(o Money) (Money, error) {
    if m.currency != o.currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{currency: m.currency, amount: m.amount + o.amount}, nil
}
```

**评审追问**：`Equals` 比较是否基于值而非指针？对外暴露的构造函数是否保证 **合法组合**（例如币种非空、金额非负）？

```go
// GOOD: constructor validates
func NewMoney(currency string, amount int64) (Money, error) {
    if currency == "" {
        return Money{}, ErrInvalidCurrency
    }
    if amount < 0 {
        return Money{}, ErrNegativeAmount
    }
    return Money{currency: currency, amount: amount}, nil
}
```

---

### 3. Repository 接口

**标准**：**Repository 接口**是否定义在**领域层**（或由内层拥有的端口包）？方法名是否表达 **业务需要**（`FindActiveByCustomer`）而非表驱动（`SelectFromOrdersJoin`）？

**反例**：接口放在 `infra` 包，领域层 `import infra` 拉平依赖方向。

```go
// BAD: domain importing infra-defined repository interface
import "github.com/org/repo/infra/persistence"

type OrderService struct {
    Repo persistence.GormOrderRepository // concrete technology leaks inward naming
}
```

**合规**：`domain/repository/order.go` 定义 `OrderRepository`，`infra` 实现。

```go
// GOOD: port owned by domain
package repository

type OrderRepository interface {
    Load(ctx context.Context, id OrderID) (*Order, error)
    Save(ctx context.Context, o *Order) error
}
```

**评审追问**：接口方法是否泄露 **分页实现细节**（offset/limit）到领域？读侧复杂筛选是否应归入 **Query 侧** 而非 `Repository` 万能方法？

---

### 4. Command 设计

**标准**：命令是否表达 **业务意图**（如 `PlaceOrder`、`CancelSubscription`），而不是贫血 CRUD（`CreateOrder` 仅映射 HTTP POST）？

**反例**：`UpdateOrder` 接收任意字段 map，语义不清、不变量无法集中校验。

```go
// BAD: command is just a data bag
type UpdateOrderCommand struct {
    OrderID string
    Patch   map[string]any
}
```

```go
// GOOD: explicit intent
type PlaceOrderCommand struct {
    CustomerID string
    Items      []OrderItemDTO
    IdempotencyKey string
}
```

**参考**：[31-clean-code](/system-design/31-clean-code/) 中与「意图命名」相关的章节（配合 §4 Pipeline 组织用例）。

**评审追问**：命令是否携带 **幂等键**、**版本/乐观锁**、**操作者身份** 等横切要素？失败时是否可映射为明确的业务结果（而非一律 500）？

---

### 5. Query 设计

**标准**：查询是否**直接返回 DTO / 读模型**，**不强行加载完整领域图**？是否避免在查询路径上触发写模型副作用？

**反例**：`GetOrderForReport` 返回 `*Order` 聚合并附带懒加载副作用。

```go
// BAD: query returns rich aggregate used for read-only UI
func (s *QueryService) OrderForUI(ctx context.Context, id string) (*domain.Order, error) {
    return s.orders.LoadFullGraph(ctx, id) // over-fetch, coupling read to write model
}
```

```go
// GOOD: dedicated read DTO
type OrderSummaryDTO struct {
    OrderID     string
    Status      string
    TotalCent   int64
    PlacedAt    time.Time
}
```

**评审追问**：查询是否 **只读**、无副作用？是否避免在 Query 路径开事务写审计表（应下沉到命令或异步）？

---

### 6. 领域事件

**标准**：关键业务状态变更是否发布 **领域事件**？命名是否使用 **过去式**（`OrderPlaced`、`PaymentCaptured`）并携带必要上下文（版本、发生时间）？

**反例**：事件名为 `PlaceOrder`，或事件体只有 ID 无版本，消费者无法安全演进。

```go
// BAD: imperative name
type PlaceOrder struct { OrderID string }

// GOOD: past tense, domain vocabulary
type OrderPlaced struct {
    OrderID   string
    OccurredAt time.Time
    Version   int
}
```

**参考**：[30-架构方法论](/system-design/30-clean-architecture-ddd-cqrs/) §2.7（领域事件与集成）。

---

### 7. 模式选型（决策表）

详设阶段可快速对照下表，避免「每个地方都 if-else」或「每个地方都上框架」。

| 场景特征 | 推荐模式 | 参考 |
|----------|----------|------|
| 多步骤顺序流程 | Pipeline（管道） | [31-clean-code §4](/system-design/31-clean-code/) |
| 同一接口多种实现 | 策略模式 | [31-clean-code §6.1](/system-design/31-clean-code/) |
| 频繁变化的业务规则 | 规则引擎 / 规则表驱动 | [31-clean-code §7](/system-design/31-clean-code/) |
| 跨聚合协作 | 领域事件 + Outbox | [30-架构方法论 §2.7](/system-design/30-clean-architecture-ddd-cqrs/) |

**标准**：选型是否写清 **触发条件、失败语义、测试策略**？是否避免把本应稳定的领域规则埋在 JSON 配置里却无人审核？

**反例**：全系统统一 `RuleEngine.Execute(ctx, ruleSetID, facts)`，但规则集无人版本化与评审，线上等于「可执行的配置漂移」。

**合规**：规则变更走 **PR + 审计 + 影子流量**；核心不变量仍保留在代码与单测中，引擎只编排**可变的参数化策略**。

---

## 三、代码评审阶段 — PR 期

**适用时机**：每次合并请求。本节是清单中最细的部分：把设计约束落到 **Go 代码**的可观察性质上。

### 3.1 SOLID 原则

对每一项，用「一句检查问句」+「违规 vs 合规」最小代码对照。

#### S — 单一职责原则（SRP）

**检查**：该类型是否只有一个变化理由（一个业务职责）？

```go
// BAD: order service also sends email and parses CSV
type OrderService struct{}
func (s *OrderService) PlaceOrder(ctx context.Context, cmd PlaceOrderCommand) error { return nil }
func (s *OrderService) SendPromoEmail(ctx context.Context, userID string) error { return nil }
func (s *OrderService) ImportOrdersFromCSV(ctx context.Context, r io.Reader) error { return nil }
```

```go
// GOOD: split by responsibility
type OrderApplicationService struct { /* deps */ }
func (s *OrderApplicationService) PlaceOrder(ctx context.Context, cmd PlaceOrderCommand) error { return nil }

type NotificationService struct { /* deps */ }
func (s *NotificationService) SendPromoEmail(ctx context.Context, userID string) error { return nil }
```

#### O — 开闭原则（OCP）

**检查**：扩展新行为时，是否**无需修改**原有稳定代码路径（优先组合、接口、策略）？

```go
// BAD: every new payment method edits the same function
func ChargePayment(method string, amount int64) error {
    switch method {
    case "card":
        return chargeCard(amount)
    case "wallet":
        return chargeWallet(amount)
    default:
        return errors.New("unknown")
    }
}
```

```go
// GOOD: open for extension via interface
type PaymentGateway interface {
    Charge(ctx context.Context, amount int64) error
}
```

```go
// GOOD: add new gateway without editing existing orchestration
type StripeGateway struct{}
func (StripeGateway) Charge(ctx context.Context, amount int64) error { return nil }

type PayPalGateway struct{}
func (PayPalGateway) Charge(ctx context.Context, amount int64) error { return nil }

type BillingService struct{ GW PaymentGateway }

func (b *BillingService) Capture(ctx context.Context, amount int64) error {
    return b.GW.Charge(ctx, amount)
}
```

#### L — 里氏替换原则（LSP）

**检查**：子类型/实现是否**可完全替换**接口契约而不破坏调用方假设（不缩小前置条件、不放大后置失败）？

```go
// BAD: implementation surprises caller by doing nothing
type NoOpPaymentGateway struct{}
func (NoOpPaymentGateway) Charge(ctx context.Context, amount int64) error {
    return nil // silently skips payment — violates expectation of "Charge"
}
```

```go
// GOOD: explicit test double with honest behavior
type FakePaymentGateway struct{ Err error }
func (f FakePaymentGateway) Charge(ctx context.Context, amount int64) error {
    return f.Err
}
```

**评审追问**：若接口允许「可选实现」（例如缓存 `MaybeCache`），调用方是否到处 `if impl != nil`？这可能是 ISP 与职责切分不足的信号。

#### I — 接口隔离原则（ISP）

**检查**：接口是否**小而专注**，客户端是否不被迫依赖不需要的方法？

```go
// BAD: fat interface for readers
type Storage interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Put(ctx context.Context, key string, val []byte) error
    Delete(ctx context.Context, key string) error
    List(ctx context.Context, prefix string) ([]string, error)
}
```

```go
// GOOD: segregate by client need
type Reader interface {
    Get(ctx context.Context, key string) ([]byte, error)
}
type Writer interface {
    Put(ctx context.Context, key string, val []byte) error
}
```

#### D — 依赖倒置原则（DIP）

**检查**：高层模块是否依赖**抽象**（接口），而非低层具体实现？

```go
// BAD: application service constructs SQL DB
type App struct{}
func (a *App) Run() {
    db, _ := sql.Open("mysql", dsn)
    _ = db.Ping()
}
```

```go
// GOOD: inject abstraction
type App struct {
    Orders OrderRepository
}
```

```go
// GOOD: wire in main/infra
func main() {
    repo := mysql.NewOrderRepository(db)
    app := &App{Orders: repo}
    _ = app
}
```

**评审追问**：`New*` 构造函数是否把 **具体类型** 泄漏回 domain？理想情况下，domain 只认识接口，具体类型停留在 `cmd/` 或 `infra/` 的组装根。

---

### 3.2 函数质量

1. **函数长度 < 80 行**  
   **检查**：单函数是否可在一屏内理解？超长函数是否可拆为私有步骤函数或 Pipeline 阶段？

**反例**：一个 `Handle` 内顺序完成：鉴权、解析、校验、调用下游、重试、日志、指标、错误映射——应拆为 **小函数** 或 **Pipeline 阶段**（参见 [31-clean-code §4](/system-design/31-clean-code/)）。

```go
// GOOD: named steps keep the orchestration readable
func (h *PlaceOrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    if err := h.ensureAuth(ctx, r); err != nil {
        h.writeErr(w, err)
        return
    }
    cmd, err := h.decode(r)
    if err != nil {
        h.writeErr(w, err)
        return
    }
    if err := h.app.PlaceOrder(ctx, cmd); err != nil {
        h.writeErr(w, err)
        return
    }
    w.WriteHeader(http.StatusCreated)
}
```

2. **圈复杂度 < 10**  
   **检查**：深层分支是否可表驱动、早返回、策略化？可用 `gocyclo`（或 `golangci-lint` 内置规则）在 CI 中强制执行。

```bash
# example: analyze cyclomatic complexity (install gocyclo if needed)
gocyclo -over 10 ./...
```

3. **嵌套深度 < 3 层**  
   **检查**：是否用 **guard clause** 减少 `if` 金字塔？

```go
// BAD: deep nesting
func Handle(r *http.Request) error {
    if r.Method == http.MethodPost {
        if err := parse(r); err == nil {
            if ok := authorize(r); ok {
                return doWork(r)
            }
        }
    }
    return errors.New("fail")
}
```

```go
// GOOD: flatten with guards
func Handle(r *http.Request) error {
    if r.Method != http.MethodPost {
        return ErrMethodNotAllowed
    }
    if err := parse(r); err != nil {
        return err
    }
    if !authorize(r) {
        return ErrForbidden
    }
    return doWork(r)
}
```

4. **参数个数 < 5**  
   **检查**：超过四个参数时，是否使用 **Options 结构体**或 **functional options**，或按上下文分组？

```go
// BAD: too many parameters
func NewClient(host string, port int, timeout time.Duration, retries int, token string) *Client {
    return &Client{}
}
```

```go
// GOOD: options struct
type ClientOptions struct {
    Host    string
    Port    int
    Timeout time.Duration
    Retries int
    Token   string
}
func NewClient(opt ClientOptions) *Client { return &Client{} }
```

**functional options 补充**（适合可选参数多、未来扩展频繁的场景）：

```go
type clientOption func(*Client)

func WithTimeout(d time.Duration) clientOption {
    return func(c *Client) { c.timeout = d }
}

func NewClient(host string, port int, opts ...clientOption) *Client {
    c := &Client{host: host, port: port, timeout: 5 * time.Second}
    for _, o := range opts {
        o(c)
    }
    return c
}
```

**评审追问**：`context.Context` 是否作为 **第一个参数** 传递 I/O 边界函数，而不是塞进结构体字段隐式携带？

---

### 3.3 命名与通用语言

1. **变量 / 函数名反映业务术语**  
   **检查**：名称是否来自 **Ubiquitous Language**，而非数据库列名的机械翻译？

2. **与团队通用语言一致**  
   **检查**：同一概念是否只有一个词（`Customer` vs `User` 混用要治理）。

3. **不用技术术语代替业务术语**  
   **检查**：是否出现 `SetStatus(1)` 这类**魔法状态**，而不是 `MarkShipped()`？

```go
// BAD: technical + magic number
func (o *Order) SetStatus(s int) { o.status = s }

// GOOD: business verb
func (o *Order) MarkShipped(at time.Time) error {
    if o.status != StatusPaid {
        return ErrInvalidStateTransition
    }
    o.status = StatusShipped
    o.shippedAt = at
    return nil
}
```

---

### 3.4 错误处理

1. **禁止静默忽略错误**  
   **检查**：是否存在 `_ = xxx` 或空白 `if err != nil { }`？

```go
// BAD
_ = os.Remove(path)
```

```go
// GOOD
if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
    return fmt.Errorf("remove temp file: %w", err)
}
```

2. **错误 wrap 携带上下文**  
   **检查**：跨层返回是否使用 `%w` 保留链，并带上**业务动作**语义？

```go
return fmt.Errorf("place order: %w", err)
```

3. **区分业务错误与系统错误**  
   **检查**：调用方能否区分「预期失败」（库存不足）与「应重试 / 告警」的基础设施错误？可用 `errors.Is` / 自定义哨兵错误类型 / `fmt.Errorf` 包装约定。

```go
var ErrOutOfStock = errors.New("out of stock")

func (s *InventoryService) Reserve(ctx context.Context, sku string, qty int) error {
    if qty > available(sku) {
        return fmt.Errorf("reserve %s: %w", sku, ErrOutOfStock)
    }
    return nil
}
```

**反例**：`UserID` 在支付域叫 `payer_ref`，在账户域叫 `uid`，在日志里叫 `operator`——评审时应要求统一 **词汇表**（可放在仓库 `docs/glossary.md`）。

---

### 3.5 依赖方向

1. **domain 包不 import adapter / infra**  
   **检查**：`go list -deps` 或 IDE 依赖图是否显示内层干净？

2. **application 只依赖 domain（及标准库 / 通用类型）**  
   **检查**：应用服务是否直接引用 HTTP、ORM、消息 SDK？

3. **无循环依赖**  
   **检查**：包之间是否存在 import 环？出现时应拆接口或提取共享内核类型包。

```bash
# detect import cycles (Go toolchain)
go build ./...
```

**反例**：`domain/order` import `domain/payment` 同时 `domain/payment` import `domain/order`，靠 `interface{}` 或事件总线「糊墙」。

**合规**：提取 **`domain/sharedkernel`** 仅放 ID、金额、时间等最小类型；或把协作上移到 **application** 编排层。

---

### 3.6 DDD 战术模式

1. **聚合根方法保护不变量**  
   **检查**：状态变更是否集中在根上，并在方法内校验规则？

```go
// GOOD: invariant enforced in root method (amounts simplified as int64 cents)
func (o *Order) AddLine(sku string, qty int, unitCent int64) error {
    if qty <= 0 {
        return ErrInvalidQty
    }
    if o.status != StatusDraft {
        return ErrOrderNotEditable
    }
    lineTotal := unitCent * int64(qty)
    if lineTotal < 0 {
        return ErrOverflow
    }
    o.lines = append(o.lines, OrderLine{SKU: sku, Qty: qty, UnitCent: unitCent})
    o.totalCent += lineTotal
    return nil
}
```

2. **值对象不可变（无 setter）**  
   **检查**：值类型字段是否导出写路径？

3. **不在聚合外部直接修改内部实体**  
   **检查**：是否暴露可变内部集合（如 `[]*Line` 直接返回引用）？

```go
// BAD: exposes mutable internal slice
func (o *Order) Lines() []*OrderLine { return o.lines }

// GOOD: return copy or read-only view
func (o *Order) Lines() []OrderLine {
    out := make([]OrderLine, len(o.lines))
    copy(out, o.lines)
    return out
}
```

4. **一个事务一个聚合**  
   **检查**：Repository `Save` 是否在单事务内写入多个根？若必须协作，是否已上升为**事件 + 最终一致性**设计并文档化？

**Saga / 补偿**：若业务强要求跨聚合原子性，是否在架构评审阶段明确 **Saga**、**幂等**、**对账** 而非偷偷用长事务？

---

## 四、上线前检查 — 合并期

**适用时机**：发布分支、灰度前、重大重构合并前。与功能完成度无关的「生产就绪」项在此收敛。

### 1. 性能

**标准**：关键路径是否有 **benchmark**（或等价的压测脚本与基线）？是否排查 **goroutine / channel 泄漏**（长时间运行测试、阻塞 send、未关闭的 worker）？

```go
func BenchmarkPlaceOrder(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        // exercise hot path
    }
}
```

**评审追问**：是否对比过 **alloc/op**？是否在负载下检查 **GC 停顿** 与 **锁竞争**（`mutex` profile）？异步路径是否避免 **无界队列** 导致内存膨胀？

**泄漏排查 sketch**：对长期运行的集成测试使用 `runtime.NumGoroutine()` 采样，或 `go test -race` 暴露 data race 与可疑同步。

---

### 2. 并发安全

**标准**：共享可变状态是否由 **mutex**、**channel 编排**或 **单 goroutine 所有权**保护？map 并发读写是否禁止？

```go
// BAD: map + goroutines without synchronization
var cache = map[string]int{}

func Set(k string, v int) {
    go func() { cache[k] = v }()
}
```

```go
// GOOD: protect shared map
type SafeCache struct {
    mu sync.RWMutex
    m  map[string]int
}

func (c *SafeCache) Set(k string, v int) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.m[k] = v
}
```

**评审追问**：`RWMutex` 的 **读锁重入**、**锁顺序**（多把锁）是否文档化？是否避免在锁内调用可能阻塞的外部 I/O？

---

### 3. 可观测性

**标准**：是否具备 **metrics**（RED/USE）、**trace**（关键 span）、**结构化日志**（带 `request_id`、`order_id` 等关联字段）？

**评审追问**：日志是否 **可查询**（键值字段而非拼接长句）？trace 是否在 **跨服务** 边界传播 `traceparent`？关键指标是否有 **SLO** 与告警阈值（避免「上线了才第一次看监控」）？

```go
// GOOD: structured context in log fields (pseudo API)
logger.Info("order_placed",
    "order_id", orderID,
    "customer_id", customerID,
    "duration_ms", elapsed.Milliseconds(),
)
```

---

### 4. 测试覆盖

**标准**：核心业务规则覆盖率是否 **> 80%**（按团队约定工具统计）？是否有 **集成测试** 覆盖仓储、消息、外部 HTTP 的 fake / 容器测试？

**评审追问**：表格驱动测试是否覆盖 **边界与错误路径**？是否用 **黄金文件** 或 **属性测试**（可选）补强复杂规则？ flaky 测试是否标记并修复，而不是 `t.Skip` 永久化？

```go
func TestPlaceOrder_OutOfStock(t *testing.T) {
    t.Parallel()
    // arrange inventory with 0 stock, expect ErrOutOfStock
}
```

---

### 5. 回滚方案

**标准**：是否有 **feature flag** 或配置开关？**数据库迁移**是否可回滚或具备向前兼容的双写/双读阶段？

**评审追问**：配置变更是否 **版本化**？破坏性 API 是否 **并行双版本** 一段时间？事件 schema 是否 **向后兼容** 或采用 **双写新字段** 策略？

---

### 6. 文档更新

**标准**：架构变更（新 BC、事件契约、SLA）是否同步到 **README / ADR / 运维手册**？Review 链接是否可追溯到决策记录？

**评审追问**：On-call 是否知道 **如何降级**、**如何重放消息**、**如何解读关键告警**？新人能否仅凭文档跑起 **本地依赖**（docker-compose / makefile 目标）？

---

## 附录：快速参考卡片

下列 20 条是各阶段「若只能记五条」时的**高杠杆**提醒；完整项仍以正文为准。

| 阶段 | #1 | #2 | #3 | #4 | #5 |
|------|----|----|----|----|-----|
| 架构评审 | 依赖向内 | BC 划分 | 聚合边界 | 读写评估 | YAGNI |
| 设计评审 | 聚合根入口 | 值对象不可变 | Repo 在领域层 | Command 表达意图 | 领域事件 |
| 代码评审 | SRP | 函数 < 80 行 | 业务命名 | 错误 wrap | 依赖方向 |
| 上线前 | Benchmark | 并发安全 | 可观测性 | 测试 > 80% | 回滚方案 |

**用法**：打印或放进 MR 模板描述区；负责人对勾选结果负责，避免形式主义勾选。

### MR 描述区模板示例（可复制）

将下列 Markdown 粘到 Merge Request 正文，作者先自评，审阅者补勾或评论编号。

```markdown
## Self review (author)
- [ ] 3.1 SOLID: no obvious SRP/OCP violations in new types
- [ ] 3.2 Function size / complexity / nesting / arity
- [ ] 3.3 Naming aligns with glossary
- [ ] 3.4 Errors wrapped, no silent `_ = err`
- [ ] 3.5 Dependency direction respected
- [ ] 3.6 DDD tactical: aggregate invariants, VO immutability

## Release readiness (if applicable)
- [ ] Benchmark or load evidence linked
- [ ] Concurrency / race checked
- [ ] Metrics + logs + traces for new paths
- [ ] Tests: core coverage & integration
- [ ] Rollback / migration plan
- [ ] Docs / ADR updated

## Design links
- ADR / RFC: ...
```

### 按角色的「最小阅读路径」

| 角色 | 建议优先阅读 |
|------|----------------|
| 作者（提 PR） | 第三节全文 + 附录卡片 |
| 审阅者（同域） | 3.3–3.6 + 第二节与本文冲突点 |
| Tech Lead（新模块） | 第一、二节 + 第四节 |
| SRE / On-call | 第四节 + 事件与迁移说明 |

---

## 参考资料

### 站内文章

- [复杂业务中的 Clean Code 实践指南](/system-design/31-clean-code/)
- [Clean Architecture、DDD 与 CQRS：三位一体的架构方法论](/system-design/30-clean-architecture-ddd-cqrs/)

### 外部资料

- Robert C. Martin, *Clean Architecture: A Craftsman's Guide to Software Structure and Design*
- Eric Evans, *Domain-Driven Design: Tackling Complexity in the Heart of Software*
- Martin Fowler, [CQRS](https://martinfowler.com/bliki/CQRS.html)（模式概述与适用边界）

---

## 总结

系统化的 Code Review 不是挑剔，而是**把重构前移到成本最低的阶段**。按 **架构 → 设计 → 代码 → 上线前** 四段清单推进，并与 [31-clean-code](/system-design/31-clean-code/)、[30-架构方法论](/system-design/30-clean-architecture-ddd-cqrs/) 交叉引用，团队可以在一致的语言下讨论分层、边界与实现细节。建议把本文的「附录快速参考」嵌入 MR 模板，并在复盘时根据失效案例**增补你们自己的第 21 条**——最好的 Checklist 永远是活文档。
