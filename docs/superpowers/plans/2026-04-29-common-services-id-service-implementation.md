# Common Services ID Service Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a production-grade example `common-services/id-service` under `ecommerce-book/example-codes` with governed namespaces, Segment, Snowflake, ULID, business number formatting, HTTP APIs, MySQL persistence, metrics, audit hooks, tests, and documentation.

**Architecture:** `common-services` is an independent Go module. Core ID concepts and algorithms live in `internal/idgen`, infrastructure adapters live in `internal/infrastructure`, HTTP lives in `internal/interfaces/http`, and process assembly lives in `internal/bootstrap` plus `cmd/id-server`.

**Tech Stack:** Go 1.21 standard library, `database/sql`, optional runtime MySQL driver import, `net/http`, `httptest`, in-memory fakes for tests, MySQL schema for production-like persistence.

---

## File Structure

Create these files:

- `ecommerce-book/example-codes/common-services/go.mod`: independent Go module named `common-services`.
- `ecommerce-book/example-codes/common-services/README.md`: runtime guide, API examples, production semantics, and mapping to Appendix H.
- `ecommerce-book/example-codes/common-services/cmd/id-server/main.go`: process entrypoint.
- `ecommerce-book/example-codes/common-services/internal/bootstrap/app.go`: config loading, dependency wiring, server construction.
- `ecommerce-book/example-codes/common-services/internal/idgen/types.go`: shared types, interfaces, errors, request/response models.
- `ecommerce-book/example-codes/common-services/internal/idgen/registry/registry.go`: namespace registry and default configs.
- `ecommerce-book/example-codes/common-services/internal/idgen/registry/registry_test.go`: registry tests.
- `ecommerce-book/example-codes/common-services/internal/idgen/formatter/business_no.go`: base36 business number formatter.
- `ecommerce-book/example-codes/common-services/internal/idgen/formatter/business_no_test.go`: formatter tests.
- `ecommerce-book/example-codes/common-services/internal/idgen/ulid/generator.go`: dependency-free ULID-like generator.
- `ecommerce-book/example-codes/common-services/internal/idgen/ulid/generator_test.go`: ULID tests.
- `ecommerce-book/example-codes/common-services/internal/idgen/segment/generator.go`: Segment generator with local buffer.
- `ecommerce-book/example-codes/common-services/internal/idgen/segment/generator_test.go`: Segment tests.
- `ecommerce-book/example-codes/common-services/internal/idgen/snowflake/generator.go`: Snowflake generator with injectable clock and lease status.
- `ecommerce-book/example-codes/common-services/internal/idgen/snowflake/generator_test.go`: Snowflake tests.
- `ecommerce-book/example-codes/common-services/internal/idgen/router/router.go`: namespace-driven routing across generators and formatter.
- `ecommerce-book/example-codes/common-services/internal/idgen/router/router_test.go`: router tests.
- `ecommerce-book/example-codes/common-services/internal/infrastructure/memory/store.go`: in-memory registry, segment, lease, audit store for tests and no-MySQL demo.
- `ecommerce-book/example-codes/common-services/internal/infrastructure/metrics/metrics.go`: lightweight counters and text exposition.
- `ecommerce-book/example-codes/common-services/internal/infrastructure/audit/audit.go`: audit service wrapper.
- `ecommerce-book/example-codes/common-services/internal/infrastructure/lease/worker.go`: worker lease manager abstraction and background heartbeat.
- `ecommerce-book/example-codes/common-services/internal/infrastructure/mysql/mysql.go`: MySQL schema, namespace seeds, segment store, worker lease, audit store.
- `ecommerce-book/example-codes/common-services/internal/interfaces/http/handler.go`: HTTP routes, request parsing, JSON responses.
- `ecommerce-book/example-codes/common-services/internal/interfaces/http/handler_test.go`: HTTP tests.

Avoid modifying `order-service` and `product-service` in this implementation pass.

## Task 1: Scaffold Module And Core Types

**Files:**
- Create: `ecommerce-book/example-codes/common-services/go.mod`
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/types.go`

- [ ] **Step 1: Create module file**

Create `go.mod`:

```go
module common-services

go 1.21
```

- [ ] **Step 2: Create shared ID types**

Create `internal/idgen/types.go` with this exported surface:

```go
package idgen

import (
    "context"
    "errors"
    "fmt"
    "time"
)

type IDType string
type GeneratorType string
type ExposeScope string
type NamespaceStatus string
type ErrorCode string

const (
    IDTypeInt64      IDType = "INT64"
    IDTypeString     IDType = "STRING"
    IDTypeBusinessNo IDType = "BUSINESS_NO"

    GeneratorSegment   GeneratorType = "SEGMENT"
    GeneratorSnowflake GeneratorType = "SNOWFLAKE"
    GeneratorULID      GeneratorType = "ULID"

    ExposeInternal ExposeScope = "INTERNAL"
    ExposeExternal ExposeScope = "EXTERNAL"
    ExposeMixed    ExposeScope = "MIXED"

    NamespaceEnabled    NamespaceStatus = "ENABLED"
    NamespaceDisabled   NamespaceStatus = "DISABLED"
    NamespaceDeprecated NamespaceStatus = "DEPRECATED"

    ErrNamespaceNotFound ErrorCode = "NAMESPACE_NOT_FOUND"
    ErrNamespaceDisabled ErrorCode = "NAMESPACE_DISABLED"
    ErrGeneratorNotReady ErrorCode = "GENERATOR_NOT_READY"
    ErrSegmentExhausted  ErrorCode = "SEGMENT_EXHAUSTED"
    ErrWorkerLeaseLost   ErrorCode = "WORKER_LEASE_LOST"
    ErrClockRollback     ErrorCode = "CLOCK_ROLLBACK"
    ErrBatchTooLarge     ErrorCode = "BATCH_TOO_LARGE"
    ErrInvalidRequest    ErrorCode = "INVALID_REQUEST"
)

type NamespaceConfig struct {
    Namespace     string
    BizDomain     string
    IDType        IDType
    GeneratorType GeneratorType
    Prefix        string
    ExposeScope   ExposeScope
    Step          int64
    MaxCapacity   int64
    OwnerTeam     string
    Status        NamespaceStatus
}

type IssueRequest struct {
    Namespace string
    Caller    string
    RequestID string
    Count     int
    Now       time.Time
}

type IssueResult struct {
    Namespace string        `json:"namespace"`
    IDType    IDType        `json:"id_type"`
    Generator GeneratorType `json:"generator"`
    RawID     int64         `json:"raw_id,omitempty"`
    ID        string        `json:"id"`
    IssuedAt  time.Time     `json:"issued_at"`
}

type BatchResult struct {
    Namespace string        `json:"namespace"`
    IDType    IDType        `json:"id_type"`
    Generator GeneratorType `json:"generator"`
    Count     int           `json:"count"`
    IDs       []string      `json:"ids"`
    RawIDs    []int64       `json:"raw_ids,omitempty"`
    IssuedAt  time.Time     `json:"issued_at"`
}

type ServiceError struct {
    Code      ErrorCode `json:"code"`
    Message   string    `json:"message"`
    Namespace string    `json:"namespace,omitempty"`
    Retryable bool      `json:"retryable"`
}

func (e *ServiceError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code ErrorCode, namespace, message string, retryable bool) *ServiceError {
    return &ServiceError{Code: code, Namespace: namespace, Message: message, Retryable: retryable}
}

func AsServiceError(err error) (*ServiceError, bool) {
    var svcErr *ServiceError
    if errors.As(err, &svcErr) {
        return svcErr, true
    }
    return nil, false
}

type Registry interface {
    Get(ctx context.Context, namespace string) (NamespaceConfig, error)
    List(ctx context.Context) ([]NamespaceConfig, error)
}

type Service interface {
    Next(ctx context.Context, req IssueRequest) (IssueResult, error)
    Batch(ctx context.Context, req IssueRequest) (BatchResult, error)
}
```

- [ ] **Step 3: Format and compile**

Run:

```bash
cd ecommerce-book/example-codes/common-services
gofmt -w internal/idgen/types.go
go test ./...
```

Expected: `go test` succeeds with no test files.

- [ ] **Step 4: Commit scaffold**

```bash
git add ecommerce-book/example-codes/common-services/go.mod ecommerce-book/example-codes/common-services/internal/idgen/types.go
git commit -m "Add common services ID core types"
```

## Task 2: Registry Defaults

**Files:**
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/registry/registry.go`
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/registry/registry_test.go`

- [ ] **Step 1: Write registry tests**

Create `registry_test.go`:

```go
package registry

import (
    "context"
    "testing"

    "common-services/internal/idgen"
)

func TestStaticRegistryReturnsDefaultNamespace(t *testing.T) {
    reg := NewStatic(DefaultNamespaces())
    cfg, err := reg.Get(context.Background(), "trade.order")
    if err != nil {
        t.Fatalf("get namespace: %v", err)
    }
    if cfg.GeneratorType != idgen.GeneratorSnowflake {
        t.Fatalf("generator = %s", cfg.GeneratorType)
    }
    if cfg.IDType != idgen.IDTypeBusinessNo {
        t.Fatalf("id type = %s", cfg.IDType)
    }
    if cfg.Prefix != "ORD" {
        t.Fatalf("prefix = %s", cfg.Prefix)
    }
}

func TestStaticRegistryRejectsMissingNamespace(t *testing.T) {
    reg := NewStatic(DefaultNamespaces())
    _, err := reg.Get(context.Background(), "unknown.namespace")
    svcErr, ok := idgen.AsServiceError(err)
    if !ok {
        t.Fatalf("expected ServiceError, got %T", err)
    }
    if svcErr.Code != idgen.ErrNamespaceNotFound {
        t.Fatalf("code = %s", svcErr.Code)
    }
}

func TestStaticRegistryRejectsDisabledNamespace(t *testing.T) {
    cfgs := DefaultNamespaces()
    cfgs[0].Status = idgen.NamespaceDisabled
    reg := NewStatic(cfgs)
    _, err := reg.Get(context.Background(), cfgs[0].Namespace)
    svcErr, ok := idgen.AsServiceError(err)
    if !ok {
        t.Fatalf("expected ServiceError, got %T", err)
    }
    if svcErr.Code != idgen.ErrNamespaceDisabled {
        t.Fatalf("code = %s", svcErr.Code)
    }
}
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
cd ecommerce-book/example-codes/common-services
go test ./internal/idgen/registry -run TestStaticRegistry -v
```

Expected: fail because `NewStatic` and `DefaultNamespaces` do not exist.

- [ ] **Step 3: Implement registry**

Create `registry.go`:

```go
package registry

import (
    "context"
    "sort"

    "common-services/internal/idgen"
)

type StaticRegistry struct {
    byNamespace map[string]idgen.NamespaceConfig
}

func NewStatic(configs []idgen.NamespaceConfig) *StaticRegistry {
    byNamespace := make(map[string]idgen.NamespaceConfig, len(configs))
    for _, cfg := range configs {
        byNamespace[cfg.Namespace] = cfg
    }
    return &StaticRegistry{byNamespace: byNamespace}
}

func (r *StaticRegistry) Get(ctx context.Context, namespace string) (idgen.NamespaceConfig, error) {
    cfg, ok := r.byNamespace[namespace]
    if !ok {
        return idgen.NamespaceConfig{}, idgen.NewError(idgen.ErrNamespaceNotFound, namespace, "namespace is not registered", false)
    }
    if cfg.Status != idgen.NamespaceEnabled {
        return idgen.NamespaceConfig{}, idgen.NewError(idgen.ErrNamespaceDisabled, namespace, "namespace is not enabled", false)
    }
    return cfg, nil
}

func (r *StaticRegistry) List(ctx context.Context) ([]idgen.NamespaceConfig, error) {
    result := make([]idgen.NamespaceConfig, 0, len(r.byNamespace))
    for _, cfg := range r.byNamespace {
        result = append(result, cfg)
    }
    sort.Slice(result, func(i, j int) bool {
        return result[i].Namespace < result[j].Namespace
    })
    return result, nil
}

func DefaultNamespaces() []idgen.NamespaceConfig {
    return []idgen.NamespaceConfig{
        segment("product.item", "product", "item", 1000, 800000000000),
        segment("product.spu", "product", "spu", 1000, 700000000000),
        segment("product.sku", "product", "sku", 1000, 600000000000),
        ulid("supply.draft", "supply", "draft"),
        ulid("supply.staging", "supply", "staging"),
        ulid("supply.qc_review", "supply", "qc"),
        ulid("checkout.session", "trade", "chk"),
        business("trade.order", "trade", "ORD"),
        business("trade.payment", "trade", "PAY"),
        business("trade.refund", "trade", "RF"),
        snowflake("inventory.ledger", "inventory", "ledger"),
        ulid("inventory.reservation", "inventory", "rsrv"),
        ulid("event.outbox", "event", "evt"),
        ulid("trace.operation", "trace", "op"),
    }
}

func segment(ns, domain, prefix string, step, start int64) idgen.NamespaceConfig {
    return idgen.NamespaceConfig{
        Namespace: ns, BizDomain: domain, IDType: idgen.IDTypeInt64,
        GeneratorType: idgen.GeneratorSegment, Prefix: prefix,
        ExposeScope: idgen.ExposeInternal, Step: step, MaxCapacity: start,
        OwnerTeam: domain + "-team", Status: idgen.NamespaceEnabled,
    }
}

func snowflake(ns, domain, prefix string) idgen.NamespaceConfig {
    return idgen.NamespaceConfig{
        Namespace: ns, BizDomain: domain, IDType: idgen.IDTypeInt64,
        GeneratorType: idgen.GeneratorSnowflake, Prefix: prefix,
        ExposeScope: idgen.ExposeInternal, Step: 0,
        OwnerTeam: domain + "-team", Status: idgen.NamespaceEnabled,
    }
}

func business(ns, domain, prefix string) idgen.NamespaceConfig {
    return idgen.NamespaceConfig{
        Namespace: ns, BizDomain: domain, IDType: idgen.IDTypeBusinessNo,
        GeneratorType: idgen.GeneratorSnowflake, Prefix: prefix,
        ExposeScope: idgen.ExposeMixed, Step: 0,
        OwnerTeam: domain + "-team", Status: idgen.NamespaceEnabled,
    }
}

func ulid(ns, domain, prefix string) idgen.NamespaceConfig {
    return idgen.NamespaceConfig{
        Namespace: ns, BizDomain: domain, IDType: idgen.IDTypeString,
        GeneratorType: idgen.GeneratorULID, Prefix: prefix,
        ExposeScope: idgen.ExposeInternal, Step: 0,
        OwnerTeam: domain + "-team", Status: idgen.NamespaceEnabled,
    }
}
```

- [ ] **Step 4: Run registry tests**

```bash
cd ecommerce-book/example-codes/common-services
gofmt -w internal/idgen/registry
go test ./internal/idgen/registry -v
```

Expected: all registry tests pass.

- [ ] **Step 5: Commit registry**

```bash
git add ecommerce-book/example-codes/common-services/internal/idgen/registry
git commit -m "Add governed ID namespace registry"
```

## Task 3: Formatter And ULID Generator

**Files:**
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/formatter/business_no.go`
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/formatter/business_no_test.go`
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/ulid/generator.go`
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/ulid/generator_test.go`

- [ ] **Step 1: Write formatter tests**

Create `formatter/business_no_test.go`:

```go
package formatter

import (
    "strings"
    "testing"
    "time"
)

func TestBusinessNumberIncludesPrefixDateAndCheckDigit(t *testing.T) {
    f := NewBusinessNumberFormatter()
    got := f.Format("ORD", 1928475629384753152, time.Date(2026, 4, 29, 10, 0, 0, 0, time.UTC))
    if !strings.HasPrefix(got, "ORD20260429") {
        t.Fatalf("business number = %s", got)
    }
    if strings.Contains(got, "1928475629384753152") {
        t.Fatalf("business number exposes raw id: %s", got)
    }
    if len(got) < len("ORD20260429A0") {
        t.Fatalf("business number too short: %s", got)
    }
}
```

- [ ] **Step 2: Implement formatter**

Create `formatter/business_no.go`:

```go
package formatter

import (
    "strconv"
    "strings"
    "time"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

type BusinessNumberFormatter struct{}

func NewBusinessNumberFormatter() *BusinessNumberFormatter {
    return &BusinessNumberFormatter{}
}

func (f *BusinessNumberFormatter) Format(prefix string, rawID int64, now time.Time) string {
    body := strings.ToUpper(strconv.FormatInt(rawID, 36))
    base := prefix + now.Format("20060102") + body
    return base + string(alphabet[checksum(base)%len(alphabet)])
}

func checksum(s string) int {
    sum := 0
    for i := 0; i < len(s); i++ {
        sum += int(s[i]) * (i + 1)
    }
    return sum
}
```

- [ ] **Step 3: Write ULID tests**

Create `ulid/generator_test.go`:

```go
package ulid

import (
    "context"
    "strings"
    "sync"
    "testing"
)

func TestGeneratorAddsPrefixAndULIDLength(t *testing.T) {
    g := NewGenerator()
    id, err := g.Next(context.Background(), "draft")
    if err != nil {
        t.Fatalf("next ulid: %v", err)
    }
    if !strings.HasPrefix(id, "draft_") {
        t.Fatalf("id = %s", id)
    }
    if len(id) != len("draft_")+26 {
        t.Fatalf("length = %d", len(id))
    }
}

func TestGeneratorConcurrentUniqueness(t *testing.T) {
    g := NewGenerator()
    const n = 1000
    seen := sync.Map{}
    var wg sync.WaitGroup
    for i := 0; i < n; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            id, err := g.Next(context.Background(), "evt")
            if err != nil {
                t.Errorf("next ulid: %v", err)
                return
            }
            if _, loaded := seen.LoadOrStore(id, true); loaded {
                t.Errorf("duplicate id: %s", id)
            }
        }()
    }
    wg.Wait()
}
```

- [ ] **Step 4: Implement ULID generator**

Create `ulid/generator.go`:

```go
package ulid

import (
    "context"
    "crypto/rand"
    "encoding/binary"
    "time"
)

const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

type Generator struct{}

func NewGenerator() *Generator {
    return &Generator{}
}

func (g *Generator) Next(ctx context.Context, prefix string) (string, error) {
    var entropy [10]byte
    if _, err := rand.Read(entropy[:]); err != nil {
        return "", err
    }
    var data [16]byte
    ms := uint64(time.Now().UnixMilli())
    data[0] = byte(ms >> 40)
    data[1] = byte(ms >> 32)
    data[2] = byte(ms >> 24)
    data[3] = byte(ms >> 16)
    data[4] = byte(ms >> 8)
    data[5] = byte(ms)
    copy(data[6:], entropy[:])
    encoded := encodeCrockford(data)
    if prefix == "" {
        return encoded, nil
    }
    return prefix + "_" + encoded, nil
}

func encodeCrockford(data [16]byte) string {
    hi := binary.BigEndian.Uint64(data[0:8])
    lo := binary.BigEndian.Uint64(data[8:16])
    value := newUint128(hi, lo)
    var out [26]byte
    for i := 25; i >= 0; i-- {
        rem := value.mod32()
        out[i] = crockford[rem]
        value.div32()
    }
    return string(out[:])
}

type uint128 struct {
    hi uint64
    lo uint64
}

func newUint128(hi, lo uint64) uint128 {
    return uint128{hi: hi, lo: lo}
}

func (u *uint128) mod32() byte {
    return byte(u.lo & 31)
}

func (u *uint128) div32() {
    u.lo = (u.lo >> 5) | (u.hi << 59)
    u.hi >>= 5
}
```

- [ ] **Step 5: Run formatter and ULID tests**

```bash
cd ecommerce-book/example-codes/common-services
gofmt -w internal/idgen/formatter internal/idgen/ulid
go test ./internal/idgen/formatter ./internal/idgen/ulid -v
```

Expected: all tests pass.

- [ ] **Step 6: Commit formatter and ULID**

```bash
git add ecommerce-book/example-codes/common-services/internal/idgen/formatter ecommerce-book/example-codes/common-services/internal/idgen/ulid
git commit -m "Add business number and ULID generators"
```

## Task 4: Segment Generator With Store Interface

**Files:**
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/segment/generator.go`
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/segment/generator_test.go`

- [ ] **Step 1: Write Segment tests**

Create `segment/generator_test.go`:

```go
package segment

import (
    "context"
    "sync"
    "testing"

    "common-services/internal/idgen"
)

func TestGeneratorAllocatesSequentialIDs(t *testing.T) {
    store := NewMemoryStore(map[string]int64{"product.sku": 600000000000}, 3)
    g := NewGenerator(store)
    cfg := idgen.NamespaceConfig{Namespace: "product.sku", Step: 3}
    got := []int64{}
    for i := 0; i < 5; i++ {
        id, err := g.Next(context.Background(), cfg)
        if err != nil {
            t.Fatalf("next: %v", err)
        }
        got = append(got, id)
    }
    want := []int64{600000000001, 600000000002, 600000000003, 600000000004, 600000000005}
    for i := range want {
        if got[i] != want[i] {
            t.Fatalf("got[%d]=%d want %d", i, got[i], want[i])
        }
    }
}

func TestGeneratorConcurrentUnique(t *testing.T) {
    store := NewMemoryStore(map[string]int64{"product.item": 800000000000}, 50)
    g := NewGenerator(store)
    cfg := idgen.NamespaceConfig{Namespace: "product.item", Step: 50}
    const n = 1000
    seen := sync.Map{}
    var wg sync.WaitGroup
    for i := 0; i < n; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            id, err := g.Next(context.Background(), cfg)
            if err != nil {
                t.Errorf("next: %v", err)
                return
            }
            if _, loaded := seen.LoadOrStore(id, true); loaded {
                t.Errorf("duplicate id %d", id)
            }
        }()
    }
    wg.Wait()
}
```

- [ ] **Step 2: Implement Segment generator and memory store**

Create `segment/generator.go`:

```go
package segment

import (
    "context"
    "sync"

    "common-services/internal/idgen"
)

type Range struct {
    Start int64
    End   int64
}

type Store interface {
    Allocate(ctx context.Context, namespace string, step int64) (Range, error)
}

type Generator struct {
    mu       sync.Mutex
    store    Store
    segments map[string]*state
}

type state struct {
    current int64
    end     int64
}

func NewGenerator(store Store) *Generator {
    return &Generator{store: store, segments: make(map[string]*state)}
}

func (g *Generator) Next(ctx context.Context, cfg idgen.NamespaceConfig) (int64, error) {
    g.mu.Lock()
    defer g.mu.Unlock()
    st := g.segments[cfg.Namespace]
    if st == nil || st.current >= st.end {
        step := cfg.Step
        if step <= 0 {
            step = 1000
        }
        allocated, err := g.store.Allocate(ctx, cfg.Namespace, step)
        if err != nil {
            return 0, idgen.NewError(idgen.ErrSegmentExhausted, cfg.Namespace, err.Error(), true)
        }
        st = &state{current: allocated.Start - 1, end: allocated.End}
        g.segments[cfg.Namespace] = st
    }
    st.current++
    return st.current, nil
}

type MemoryStore struct {
    mu    sync.Mutex
    maxID map[string]int64
    step  int64
}

func NewMemoryStore(initial map[string]int64, step int64) *MemoryStore {
    cp := make(map[string]int64, len(initial))
    for k, v := range initial {
        cp[k] = v
    }
    return &MemoryStore{maxID: cp, step: step}
}

func (s *MemoryStore) Allocate(ctx context.Context, namespace string, step int64) (Range, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    if step <= 0 {
        step = s.step
    }
    start := s.maxID[namespace] + 1
    end := s.maxID[namespace] + step
    s.maxID[namespace] = end
    return Range{Start: start, End: end}, nil
}
```

- [ ] **Step 3: Run Segment tests**

```bash
cd ecommerce-book/example-codes/common-services
gofmt -w internal/idgen/segment
go test ./internal/idgen/segment -v
```

Expected: all Segment tests pass.

- [ ] **Step 4: Commit Segment generator**

```bash
git add ecommerce-book/example-codes/common-services/internal/idgen/segment
git commit -m "Add Segment ID generator"
```

## Task 5: Snowflake Generator And Lease Guard

**Files:**
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/snowflake/generator.go`
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/snowflake/generator_test.go`

- [ ] **Step 1: Write Snowflake tests**

Create `snowflake/generator_test.go`:

```go
package snowflake

import (
    "context"
    "testing"
    "time"

    "common-services/internal/idgen"
)

func TestGeneratorProducesIncreasingIDs(t *testing.T) {
    clock := &FakeClock{NowValue: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)}
    lease := &StaticLease{ReadyValue: true, WorkerIDValue: 3, RegionIDValue: 1}
    g := NewGenerator(Config{Epoch: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), MaxWaitRollback: time.Millisecond}, clock, lease)
    first, err := g.Next(context.Background(), idgen.NamespaceConfig{Namespace: "trade.order"})
    if err != nil {
        t.Fatalf("first: %v", err)
    }
    second, err := g.Next(context.Background(), idgen.NamespaceConfig{Namespace: "trade.order"})
    if err != nil {
        t.Fatalf("second: %v", err)
    }
    if second <= first {
        t.Fatalf("ids not increasing: %d <= %d", second, first)
    }
}

func TestGeneratorRejectsLeaseLost(t *testing.T) {
    clock := &FakeClock{NowValue: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)}
    lease := &StaticLease{ReadyValue: false}
    g := NewGenerator(Config{Epoch: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, clock, lease)
    _, err := g.Next(context.Background(), idgen.NamespaceConfig{Namespace: "trade.order"})
    svcErr, ok := idgen.AsServiceError(err)
    if !ok || svcErr.Code != idgen.ErrWorkerLeaseLost {
        t.Fatalf("err = %#v", err)
    }
}

func TestGeneratorRejectsLargeClockRollback(t *testing.T) {
    clock := &FakeClock{NowValue: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)}
    lease := &StaticLease{ReadyValue: true, WorkerIDValue: 1, RegionIDValue: 1}
    g := NewGenerator(Config{Epoch: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), MaxWaitRollback: time.Millisecond}, clock, lease)
    if _, err := g.Next(context.Background(), idgen.NamespaceConfig{Namespace: "trade.order"}); err != nil {
        t.Fatalf("first: %v", err)
    }
    clock.NowValue = clock.NowValue.Add(-10 * time.Millisecond)
    _, err := g.Next(context.Background(), idgen.NamespaceConfig{Namespace: "trade.order"})
    svcErr, ok := idgen.AsServiceError(err)
    if !ok || svcErr.Code != idgen.ErrClockRollback {
        t.Fatalf("err = %#v", err)
    }
}
```

- [ ] **Step 2: Implement Snowflake generator**

Create `snowflake/generator.go`:

```go
package snowflake

import (
    "context"
    "sync"
    "time"

    "common-services/internal/idgen"
)

const (
    regionBits   = 5
    workerBits   = 5
    sequenceBits = 12
    maxRegion    = int64(1<<regionBits - 1)
    maxWorker    = int64(1<<workerBits - 1)
    maxSequence  = int64(1<<sequenceBits - 1)
    workerShift  = sequenceBits
    regionShift  = workerBits + sequenceBits
    timeShift    = regionBits + workerBits + sequenceBits
)

type Config struct {
    Epoch           time.Time
    MaxWaitRollback time.Duration
}

type Clock interface {
    Now() time.Time
    Sleep(time.Duration)
}

type Lease interface {
    Ready() bool
    WorkerID() int64
    RegionID() int64
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }
func (RealClock) Sleep(d time.Duration) { time.Sleep(d) }

type Generator struct {
    mu            sync.Mutex
    cfg           Config
    clock         Clock
    lease         Lease
    lastTimestamp int64
    sequence      int64
}

func NewGenerator(cfg Config, clock Clock, lease Lease) *Generator {
    if cfg.Epoch.IsZero() {
        cfg.Epoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
    }
    if cfg.MaxWaitRollback == 0 {
        cfg.MaxWaitRollback = 5 * time.Millisecond
    }
    return &Generator{cfg: cfg, clock: clock, lease: lease, lastTimestamp: -1}
}

func (g *Generator) Next(ctx context.Context, ns idgen.NamespaceConfig) (int64, error) {
    g.mu.Lock()
    defer g.mu.Unlock()
    if !g.lease.Ready() {
        return 0, idgen.NewError(idgen.ErrWorkerLeaseLost, ns.Namespace, "worker lease is not ready", true)
    }
    workerID := g.lease.WorkerID()
    regionID := g.lease.RegionID()
    if workerID < 0 || workerID > maxWorker || regionID < 0 || regionID > maxRegion {
        return 0, idgen.NewError(idgen.ErrWorkerLeaseLost, ns.Namespace, "worker or region is out of range", true)
    }

    timestamp := g.timestamp()
    if timestamp < g.lastTimestamp {
        delta := time.Duration(g.lastTimestamp-timestamp) * time.Millisecond
        if delta > g.cfg.MaxWaitRollback {
            return 0, idgen.NewError(idgen.ErrClockRollback, ns.Namespace, "clock moved backwards", false)
        }
        g.clock.Sleep(delta)
        timestamp = g.timestamp()
        if timestamp < g.lastTimestamp {
            return 0, idgen.NewError(idgen.ErrClockRollback, ns.Namespace, "clock moved backwards after wait", false)
        }
    }

    if timestamp == g.lastTimestamp {
        g.sequence = (g.sequence + 1) & maxSequence
        if g.sequence == 0 {
            timestamp = g.waitNextMillis(g.lastTimestamp)
        }
    } else {
        g.sequence = 0
    }
    g.lastTimestamp = timestamp
    return (timestamp << timeShift) | (regionID << regionShift) | (workerID << workerShift) | g.sequence, nil
}

func (g *Generator) timestamp() int64 {
    return g.clock.Now().Sub(g.cfg.Epoch).Milliseconds()
}

func (g *Generator) waitNextMillis(last int64) int64 {
    ts := g.timestamp()
    for ts <= last {
        g.clock.Sleep(time.Millisecond)
        ts = g.timestamp()
    }
    return ts
}

type FakeClock struct{ NowValue time.Time }

func (c *FakeClock) Now() time.Time { return c.NowValue }
func (c *FakeClock) Sleep(d time.Duration) { c.NowValue = c.NowValue.Add(d) }

type StaticLease struct {
    ReadyValue    bool
    WorkerIDValue int64
    RegionIDValue int64
}

func (l *StaticLease) Ready() bool { return l.ReadyValue }
func (l *StaticLease) WorkerID() int64 { return l.WorkerIDValue }
func (l *StaticLease) RegionID() int64 { return l.RegionIDValue }
```

- [ ] **Step 3: Run Snowflake tests**

```bash
cd ecommerce-book/example-codes/common-services
gofmt -w internal/idgen/snowflake
go test ./internal/idgen/snowflake -v
```

Expected: all Snowflake tests pass.

- [ ] **Step 4: Commit Snowflake generator**

```bash
git add ecommerce-book/example-codes/common-services/internal/idgen/snowflake
git commit -m "Add lease guarded Snowflake generator"
```

## Task 6: Router Service

**Files:**
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/router/router.go`
- Create: `ecommerce-book/example-codes/common-services/internal/idgen/router/router_test.go`

- [ ] **Step 1: Write router tests**

Create `router/router_test.go`:

```go
package router

import (
    "context"
    "strings"
    "testing"
    "time"

    "common-services/internal/idgen"
    "common-services/internal/idgen/formatter"
    "common-services/internal/idgen/registry"
    "common-services/internal/idgen/segment"
    "common-services/internal/idgen/snowflake"
    ulidgen "common-services/internal/idgen/ulid"
)

func newTestService() *Service {
    reg := registry.NewStatic(registry.DefaultNamespaces())
    seg := segment.NewGenerator(segment.NewMemoryStore(map[string]int64{
        "product.sku": 600000000000,
    }, 100))
    clock := &snowflake.FakeClock{NowValue: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)}
    sf := snowflake.NewGenerator(snowflake.Config{Epoch: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, clock, &snowflake.StaticLease{ReadyValue: true, WorkerIDValue: 1, RegionIDValue: 1})
    return New(reg, seg, sf, ulidgen.NewGenerator(), formatter.NewBusinessNumberFormatter(), 1000)
}

func TestNextBusinessOrder(t *testing.T) {
    svc := newTestService()
    got, err := svc.Next(context.Background(), idgen.IssueRequest{Namespace: "trade.order", Caller: "order-service", RequestID: "req-1"})
    if err != nil {
        t.Fatalf("next: %v", err)
    }
    if !strings.HasPrefix(got.ID, "ORD20260429") {
        t.Fatalf("id = %s", got.ID)
    }
    if got.RawID == 0 {
        t.Fatal("raw id is zero")
    }
}

func TestBatchSegmentSKU(t *testing.T) {
    svc := newTestService()
    got, err := svc.Batch(context.Background(), idgen.IssueRequest{Namespace: "product.sku", Count: 3})
    if err != nil {
        t.Fatalf("batch: %v", err)
    }
    if len(got.IDs) != 3 {
        t.Fatalf("len = %d", len(got.IDs))
    }
    if got.IDs[0] != "600000000001" {
        t.Fatalf("first id = %s", got.IDs[0])
    }
}

func TestBatchTooLarge(t *testing.T) {
    svc := newTestService()
    _, err := svc.Batch(context.Background(), idgen.IssueRequest{Namespace: "product.sku", Count: 1001})
    svcErr, ok := idgen.AsServiceError(err)
    if !ok || svcErr.Code != idgen.ErrBatchTooLarge {
        t.Fatalf("err = %#v", err)
    }
}
```

- [ ] **Step 2: Implement router**

Create `router/router.go`:

```go
package router

import (
    "context"
    "strconv"
    "time"

    "common-services/internal/idgen"
    "common-services/internal/idgen/formatter"
    "common-services/internal/idgen/segment"
    "common-services/internal/idgen/snowflake"
    ulidgen "common-services/internal/idgen/ulid"
)

type Service struct {
    registry  idgen.Registry
    segment   *segment.Generator
    snowflake *snowflake.Generator
    ulid      *ulidgen.Generator
    formatter *formatter.BusinessNumberFormatter
    maxBatch  int
}

func New(reg idgen.Registry, seg *segment.Generator, sf *snowflake.Generator, ug *ulidgen.Generator, fmt *formatter.BusinessNumberFormatter, maxBatch int) *Service {
    if maxBatch <= 0 {
        maxBatch = 1000
    }
    return &Service{registry: reg, segment: seg, snowflake: sf, ulid: ug, formatter: fmt, maxBatch: maxBatch}
}

func (s *Service) Next(ctx context.Context, req idgen.IssueRequest) (idgen.IssueResult, error) {
    cfg, err := s.registry.Get(ctx, req.Namespace)
    if err != nil {
        return idgen.IssueResult{}, err
    }
    now := req.Now
    if now.IsZero() {
        now = time.Now()
    }
    raw, id, err := s.issueOne(ctx, cfg, now)
    if err != nil {
        return idgen.IssueResult{}, err
    }
    return idgen.IssueResult{Namespace: cfg.Namespace, IDType: cfg.IDType, Generator: cfg.GeneratorType, RawID: raw, ID: id, IssuedAt: now}, nil
}

func (s *Service) Batch(ctx context.Context, req idgen.IssueRequest) (idgen.BatchResult, error) {
    if req.Count <= 0 {
        return idgen.BatchResult{}, idgen.NewError(idgen.ErrInvalidRequest, req.Namespace, "count must be positive", false)
    }
    if req.Count > s.maxBatch {
        return idgen.BatchResult{}, idgen.NewError(idgen.ErrBatchTooLarge, req.Namespace, "count exceeds max batch size", false)
    }
    cfg, err := s.registry.Get(ctx, req.Namespace)
    if err != nil {
        return idgen.BatchResult{}, err
    }
    now := req.Now
    if now.IsZero() {
        now = time.Now()
    }
    ids := make([]string, 0, req.Count)
    raws := make([]int64, 0, req.Count)
    for i := 0; i < req.Count; i++ {
        raw, id, err := s.issueOne(ctx, cfg, now)
        if err != nil {
            return idgen.BatchResult{}, err
        }
        ids = append(ids, id)
        if raw != 0 {
            raws = append(raws, raw)
        }
    }
    return idgen.BatchResult{Namespace: cfg.Namespace, IDType: cfg.IDType, Generator: cfg.GeneratorType, Count: req.Count, IDs: ids, RawIDs: raws, IssuedAt: now}, nil
}

func (s *Service) issueOne(ctx context.Context, cfg idgen.NamespaceConfig, now time.Time) (int64, string, error) {
    switch cfg.GeneratorType {
    case idgen.GeneratorSegment:
        raw, err := s.segment.Next(ctx, cfg)
        if err != nil {
            return 0, "", err
        }
        return raw, strconv.FormatInt(raw, 10), nil
    case idgen.GeneratorSnowflake:
        raw, err := s.snowflake.Next(ctx, cfg)
        if err != nil {
            return 0, "", err
        }
        if cfg.IDType == idgen.IDTypeBusinessNo {
            return raw, s.formatter.Format(cfg.Prefix, raw, now), nil
        }
        return raw, strconv.FormatInt(raw, 10), nil
    case idgen.GeneratorULID:
        id, err := s.ulid.Next(ctx, cfg.Prefix)
        return 0, id, err
    default:
        return 0, "", idgen.NewError(idgen.ErrInvalidRequest, cfg.Namespace, "unsupported generator type", false)
    }
}
```

- [ ] **Step 3: Run router tests**

```bash
cd ecommerce-book/example-codes/common-services
gofmt -w internal/idgen/router
go test ./internal/idgen/router -v
```

Expected: all router tests pass.

- [ ] **Step 4: Commit router**

```bash
git add ecommerce-book/example-codes/common-services/internal/idgen/router
git commit -m "Add ID generator router"
```

## Task 7: Metrics, Audit, Memory Infrastructure

**Files:**
- Create: `ecommerce-book/example-codes/common-services/internal/infrastructure/metrics/metrics.go`
- Create: `ecommerce-book/example-codes/common-services/internal/infrastructure/audit/audit.go`
- Create: `ecommerce-book/example-codes/common-services/internal/infrastructure/memory/store.go`

- [ ] **Step 1: Implement metrics**

Create `metrics/metrics.go`:

```go
package metrics

import (
    "fmt"
    "sort"
    "strings"
    "sync"
)

type Recorder struct {
    mu       sync.Mutex
    counters map[string]int64
    gauges   map[string]int64
}

func NewRecorder() *Recorder {
    return &Recorder{counters: make(map[string]int64), gauges: make(map[string]int64)}
}

func (r *Recorder) Inc(name string, labels map[string]string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.counters[key(name, labels)]++
}

func (r *Recorder) SetGauge(name string, labels map[string]string, value int64) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.gauges[key(name, labels)] = value
}

func (r *Recorder) Text() string {
    r.mu.Lock()
    defer r.mu.Unlock()
    lines := make([]string, 0, len(r.counters)+len(r.gauges))
    for k, v := range r.counters {
        lines = append(lines, fmt.Sprintf("%s %d", k, v))
    }
    for k, v := range r.gauges {
        lines = append(lines, fmt.Sprintf("%s %d", k, v))
    }
    sort.Strings(lines)
    return strings.Join(lines, "\n") + "\n"
}

func key(name string, labels map[string]string) string {
    if len(labels) == 0 {
        return name
    }
    names := make([]string, 0, len(labels))
    for k := range labels {
        names = append(names, k)
    }
    sort.Strings(names)
    parts := make([]string, 0, len(names))
    for _, k := range names {
        parts = append(parts, fmt.Sprintf(`%s="%s"`, k, labels[k]))
    }
    return name + "{" + strings.Join(parts, ",") + "}"
}
```

- [ ] **Step 2: Implement audit model**

Create `audit/audit.go`:

```go
package audit

import (
    "context"
    "log"
    "time"
)

type IssueLog struct {
    RequestID    string
    Namespace    string
    Caller       string
    IssueType    string
    IssuedValue  string
    ErrorCode    string
    ErrorMessage string
    CreatedAt    time.Time
}

type Store interface {
    SaveIssueLog(ctx context.Context, log IssueLog) error
}

type Service struct {
    store Store
}

func NewService(store Store) *Service {
    return &Service{store: store}
}

func (s *Service) Record(ctx context.Context, entry IssueLog) {
    if s == nil || s.store == nil {
        return
    }
    if entry.CreatedAt.IsZero() {
        entry.CreatedAt = time.Now()
    }
    if err := s.store.SaveIssueLog(ctx, entry); err != nil {
        log.Printf("[id-audit] save issue log failed: %v", err)
    }
}
```

- [ ] **Step 3: Implement memory infrastructure wrapper**

Create `memory/store.go`:

```go
package memory

import (
    "context"
    "sync"

    "common-services/internal/idgen"
    "common-services/internal/idgen/registry"
    "common-services/internal/idgen/segment"
    "common-services/internal/infrastructure/audit"
)

type Store struct {
    *registry.StaticRegistry
    Segment *segment.MemoryStore
    mu      sync.Mutex
    Logs    []audit.IssueLog
}

func NewStore() *Store {
    initial := map[string]int64{
        "product.item": 800000000000,
        "product.spu":  700000000000,
        "product.sku":  600000000000,
    }
    return &Store{
        StaticRegistry: registry.NewStatic(registry.DefaultNamespaces()),
        Segment:        segment.NewMemoryStore(initial, 1000),
        Logs:           make([]audit.IssueLog, 0),
    }
}

func (s *Store) SaveIssueLog(ctx context.Context, entry audit.IssueLog) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.Logs = append(s.Logs, entry)
    return nil
}

var _ idgen.Registry = (*Store)(nil)
```

- [ ] **Step 4: Run package tests**

```bash
cd ecommerce-book/example-codes/common-services
gofmt -w internal/infrastructure
go test ./internal/infrastructure/... ./internal/idgen/... -v
```

Expected: all existing tests pass.

- [ ] **Step 5: Commit infrastructure helpers**

```bash
git add ecommerce-book/example-codes/common-services/internal/infrastructure
git commit -m "Add ID service metrics audit and memory store"
```

## Task 8: HTTP API

**Files:**
- Create: `ecommerce-book/example-codes/common-services/internal/interfaces/http/handler.go`
- Create: `ecommerce-book/example-codes/common-services/internal/interfaces/http/handler_test.go`

- [ ] **Step 1: Write HTTP tests**

Create `handler_test.go`:

```go
package httpapi

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "common-services/internal/idgen/formatter"
    "common-services/internal/idgen/registry"
    "common-services/internal/idgen/router"
    "common-services/internal/idgen/segment"
    "common-services/internal/idgen/snowflake"
    ulidgen "common-services/internal/idgen/ulid"
    "common-services/internal/infrastructure/metrics"
)

func newHandlerForTest() http.Handler {
    reg := registry.NewStatic(registry.DefaultNamespaces())
    seg := segment.NewGenerator(segment.NewMemoryStore(map[string]int64{"product.sku": 600000000000}, 100))
    clock := &snowflake.FakeClock{NowValue: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)}
    sf := snowflake.NewGenerator(snowflake.Config{Epoch: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, clock, &snowflake.StaticLease{ReadyValue: true, WorkerIDValue: 1, RegionIDValue: 1})
    svc := router.New(reg, seg, sf, ulidgen.NewGenerator(), formatter.NewBusinessNumberFormatter(), 1000)
    return NewHandler(svc, reg, metrics.NewRecorder())
}

func TestNextID(t *testing.T) {
    h := newHandlerForTest()
    req := httptest.NewRequest(http.MethodPost, "/api/v1/ids/next", bytes.NewBufferString(`{"namespace":"trade.order","caller":"order-service","request_id":"req-1"}`))
    rec := httptest.NewRecorder()
    h.ServeHTTP(rec, req)
    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
    }
    if !strings.Contains(rec.Body.String(), "ORD") {
        t.Fatalf("body = %s", rec.Body.String())
    }
}

func TestUnknownNamespaceReturnsStructuredError(t *testing.T) {
    h := newHandlerForTest()
    req := httptest.NewRequest(http.MethodPost, "/api/v1/ids/next", bytes.NewBufferString(`{"namespace":"missing","caller":"test","request_id":"req-2"}`))
    rec := httptest.NewRecorder()
    h.ServeHTTP(rec, req)
    if rec.Code != http.StatusBadRequest {
        t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
    }
    if !strings.Contains(rec.Body.String(), "NAMESPACE_NOT_FOUND") {
        t.Fatalf("body = %s", rec.Body.String())
    }
}

func TestMetricsEndpoint(t *testing.T) {
    h := newHandlerForTest()
    req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
    rec := httptest.NewRecorder()
    h.ServeHTTP(rec, req)
    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d", rec.Code)
    }
}
```

- [ ] **Step 2: Implement HTTP handler**

Create `handler.go`:

```go
package httpapi

import (
    "encoding/json"
    "net/http"

    "common-services/internal/idgen"
    "common-services/internal/infrastructure/metrics"
)

type Handler struct {
    service  idgen.Service
    registry idgen.Registry
    metrics  *metrics.Recorder
    mux      *http.ServeMux
}

type issueRequest struct {
    Namespace string `json:"namespace"`
    Caller    string `json:"caller"`
    RequestID string `json:"request_id"`
    Count     int    `json:"count"`
}

func NewHandler(service idgen.Service, registry idgen.Registry, recorder *metrics.Recorder) *Handler {
    h := &Handler{service: service, registry: registry, metrics: recorder, mux: http.NewServeMux()}
    h.mux.HandleFunc("/api/v1/ids/next", h.next)
    h.mux.HandleFunc("/api/v1/ids/batch", h.batch)
    h.mux.HandleFunc("/api/v1/namespaces", h.namespaces)
    h.mux.HandleFunc("/healthz", h.healthz)
    h.mux.HandleFunc("/readyz", h.readyz)
    h.mux.HandleFunc("/metrics", h.metricsText)
    return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    h.mux.ServeHTTP(w, r)
}

func (h *Handler) next(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeError(w, idgen.NewError(idgen.ErrInvalidRequest, "", "method not allowed", false), http.StatusMethodNotAllowed)
        return
    }
    var req issueRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, idgen.NewError(idgen.ErrInvalidRequest, "", "invalid json", false), http.StatusBadRequest)
        return
    }
    result, err := h.service.Next(r.Context(), idgen.IssueRequest{Namespace: req.Namespace, Caller: req.Caller, RequestID: req.RequestID})
    if err != nil {
        h.metrics.Inc("idgen_requests_total", map[string]string{"namespace": req.Namespace, "result": "error"})
        writeServiceError(w, err)
        return
    }
    h.metrics.Inc("idgen_requests_total", map[string]string{"namespace": req.Namespace, "generator": string(result.Generator), "result": "success"})
    writeJSON(w, http.StatusOK, result)
}

func (h *Handler) batch(w http.ResponseWriter, r *http.Request) {
    var req issueRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, idgen.NewError(idgen.ErrInvalidRequest, "", "invalid json", false), http.StatusBadRequest)
        return
    }
    result, err := h.service.Batch(r.Context(), idgen.IssueRequest{Namespace: req.Namespace, Caller: req.Caller, RequestID: req.RequestID, Count: req.Count})
    if err != nil {
        h.metrics.Inc("idgen_batch_requests_total", map[string]string{"namespace": req.Namespace, "result": "error"})
        writeServiceError(w, err)
        return
    }
    h.metrics.Inc("idgen_batch_requests_total", map[string]string{"namespace": req.Namespace, "result": "success"})
    writeJSON(w, http.StatusOK, result)
}

func (h *Handler) namespaces(w http.ResponseWriter, r *http.Request) {
    configs, err := h.registry.List(r.Context())
    if err != nil {
        writeServiceError(w, err)
        return
    }
    writeJSON(w, http.StatusOK, map[string]any{"namespaces": configs})
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) readyz(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *Handler) metricsText(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; version=0.0.4")
    _, _ = w.Write([]byte(h.metrics.Text()))
}

func writeServiceError(w http.ResponseWriter, err error) {
    if svcErr, ok := idgen.AsServiceError(err); ok {
        status := http.StatusBadRequest
        if svcErr.Retryable {
            status = http.StatusServiceUnavailable
        }
        writeError(w, svcErr, status)
        return
    }
    writeError(w, idgen.NewError(idgen.ErrInvalidRequest, "", err.Error(), false), http.StatusInternalServerError)
}

func writeError(w http.ResponseWriter, err *idgen.ServiceError, status int) {
    writeJSON(w, status, map[string]any{"error": err})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(value)
}
```

- [ ] **Step 3: Run HTTP tests**

```bash
cd ecommerce-book/example-codes/common-services
gofmt -w internal/interfaces/http
go test ./internal/interfaces/http -v
```

Expected: all HTTP tests pass.

- [ ] **Step 4: Commit HTTP API**

```bash
git add ecommerce-book/example-codes/common-services/internal/interfaces/http
git commit -m "Add ID service HTTP API"
```

## Task 9: MySQL Store And Worker Lease

**Files:**
- Create: `ecommerce-book/example-codes/common-services/internal/infrastructure/lease/worker.go`
- Create: `ecommerce-book/example-codes/common-services/internal/infrastructure/mysql/mysql.go`

- [ ] **Step 1: Implement lease manager**

Create `lease/worker.go`:

```go
package lease

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "sync"
    "time"
)

type Store interface {
    AcquireWorker(ctx context.Context, regionID int64, datacenterCode, instanceID string, ttl time.Duration) (workerID int64, leaseToken string, err error)
    RenewWorker(ctx context.Context, regionID, workerID int64, instanceID, leaseToken string, ttl time.Duration) error
    ReleaseWorker(ctx context.Context, regionID, workerID int64, instanceID, leaseToken string) error
}

type Manager struct {
    mu             sync.RWMutex
    store          Store
    regionID       int64
    datacenterCode string
    instanceID     string
    workerID       int64
    leaseToken     string
    ttl            time.Duration
    heartbeatEvery time.Duration
    ready          bool
    stop           chan struct{}
}

func NewManager(store Store, regionID int64, datacenterCode, instanceID string, ttl, heartbeatEvery time.Duration) *Manager {
    if ttl <= 0 {
        ttl = 30 * time.Second
    }
    if heartbeatEvery <= 0 {
        heartbeatEvery = 10 * time.Second
    }
    return &Manager{
        store: store, regionID: regionID, datacenterCode: datacenterCode,
        instanceID: instanceID, workerID: -1, ttl: ttl, heartbeatEvery: heartbeatEvery,
        stop: make(chan struct{}),
    }
}

func (m *Manager) Ready() bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.ready
}

func (m *Manager) WorkerID() int64 {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.workerID
}

func (m *Manager) RegionID() int64 {
    return m.regionID
}

func (m *Manager) MarkReady(workerID int64) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.workerID = workerID
    m.ready = true
}

func (m *Manager) MarkNotReady() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.ready = false
}

func (m *Manager) Start(ctx context.Context) error {
    workerID, token, err := m.store.AcquireWorker(ctx, m.regionID, m.datacenterCode, m.instanceID, m.ttl)
    if err != nil {
        m.MarkNotReady()
        return err
    }
    m.mu.Lock()
    m.workerID = workerID
    m.leaseToken = token
    m.ready = true
    m.mu.Unlock()
    go m.heartbeat()
    return nil
}

func (m *Manager) Stop(ctx context.Context) {
    close(m.stop)
    m.mu.RLock()
    workerID, token := m.workerID, m.leaseToken
    m.mu.RUnlock()
    if workerID >= 0 && token != "" {
        _ = m.store.ReleaseWorker(ctx, m.regionID, workerID, m.instanceID, token)
    }
    m.MarkNotReady()
}

func (m *Manager) heartbeat() {
    ticker := time.NewTicker(m.heartbeatEvery)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            m.mu.RLock()
            workerID, token := m.workerID, m.leaseToken
            m.mu.RUnlock()
            if err := m.store.RenewWorker(context.Background(), m.regionID, workerID, m.instanceID, token, m.ttl); err != nil {
                m.MarkNotReady()
            }
        case <-m.stop:
            return
        }
    }
}

func NewLeaseToken() string {
    var buf [16]byte
    _, _ = rand.Read(buf[:])
    return hex.EncodeToString(buf[:])
}
```

- [ ] **Step 2: Implement MySQL schema and methods**

Create `mysql/mysql.go` with:

```go
package mysql

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "common-services/internal/idgen"
    "common-services/internal/idgen/registry"
    "common-services/internal/idgen/segment"
    "common-services/internal/infrastructure/audit"
    "common-services/internal/infrastructure/lease"
)

type Store struct {
    db *sql.DB
}

func NewStore(db *sql.DB) *Store {
    return &Store{db: db}
}

func (s *Store) InitSchema(ctx context.Context) error {
    statements := []string{
        `CREATE TABLE IF NOT EXISTS id_namespace (
            id BIGINT PRIMARY KEY AUTO_INCREMENT,
            namespace VARCHAR(64) NOT NULL,
            biz_domain VARCHAR(64) NOT NULL,
            id_type VARCHAR(32) NOT NULL,
            generator_type VARCHAR(32) NOT NULL,
            prefix VARCHAR(32) DEFAULT NULL,
            expose_scope VARCHAR(32) NOT NULL,
            step BIGINT NOT NULL DEFAULT 1000,
            max_capacity BIGINT DEFAULT 0,
            owner_team VARCHAR(64) NOT NULL,
            status VARCHAR(32) NOT NULL,
            created_at DATETIME(6) NOT NULL,
            updated_at DATETIME(6) NOT NULL,
            UNIQUE KEY uk_namespace (namespace),
            KEY idx_domain_status (biz_domain, status)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
        `CREATE TABLE IF NOT EXISTS id_segment (
            id BIGINT PRIMARY KEY AUTO_INCREMENT,
            namespace VARCHAR(64) NOT NULL,
            max_id BIGINT NOT NULL,
            step BIGINT NOT NULL,
            version BIGINT NOT NULL,
            updated_at DATETIME(6) NOT NULL,
            UNIQUE KEY uk_namespace (namespace)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
        `CREATE TABLE IF NOT EXISTS id_worker (
            id BIGINT PRIMARY KEY AUTO_INCREMENT,
            worker_id INT NOT NULL,
            region_id INT NOT NULL,
            datacenter_code VARCHAR(32) NOT NULL,
            instance_id VARCHAR(128) NOT NULL,
            lease_token VARCHAR(64) NOT NULL,
            lease_until DATETIME(6) NOT NULL,
            heartbeat_at DATETIME(6) NOT NULL,
            status VARCHAR(32) NOT NULL,
            created_at DATETIME(6) NOT NULL,
            updated_at DATETIME(6) NOT NULL,
            UNIQUE KEY uk_worker_region (worker_id, region_id),
            UNIQUE KEY uk_instance (instance_id),
            KEY idx_status_lease (status, lease_until)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
        `CREATE TABLE IF NOT EXISTS id_issue_log (
            id BIGINT PRIMARY KEY AUTO_INCREMENT,
            request_id VARCHAR(64) NOT NULL,
            namespace VARCHAR(64) NOT NULL,
            caller VARCHAR(128) NOT NULL,
            issue_type VARCHAR(32) NOT NULL,
            issued_value VARCHAR(128) DEFAULT NULL,
            error_code VARCHAR(64) DEFAULT NULL,
            error_message VARCHAR(512) DEFAULT NULL,
            created_at DATETIME(6) NOT NULL,
            UNIQUE KEY uk_request_id (request_id),
            KEY idx_namespace_time (namespace, created_at),
            KEY idx_issue_type_time (issue_type, created_at)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
    }
    for _, stmt := range statements {
        if _, err := s.db.ExecContext(ctx, stmt); err != nil {
            return err
        }
    }
    return s.SeedDefaults(ctx)
}

func (s *Store) SeedDefaults(ctx context.Context) error {
    now := time.Now()
    for _, cfg := range registry.DefaultNamespaces() {
        _, err := s.db.ExecContext(ctx, `INSERT INTO id_namespace(namespace,biz_domain,id_type,generator_type,prefix,expose_scope,step,max_capacity,owner_team,status,created_at,updated_at)
            VALUES(?,?,?,?,?,?,?,?,?,?,?,?)
            ON DUPLICATE KEY UPDATE updated_at=VALUES(updated_at)`,
            cfg.Namespace, cfg.BizDomain, cfg.IDType, cfg.GeneratorType, cfg.Prefix, cfg.ExposeScope, cfg.Step, cfg.MaxCapacity, cfg.OwnerTeam, cfg.Status, now, now)
        if err != nil {
            return err
        }
        if cfg.GeneratorType == idgen.GeneratorSegment {
            _, err = s.db.ExecContext(ctx, `INSERT INTO id_segment(namespace,max_id,step,version,updated_at)
                VALUES(?,?,?,?,?)
                ON DUPLICATE KEY UPDATE updated_at=VALUES(updated_at)`,
                cfg.Namespace, cfg.MaxCapacity, cfg.Step, 0, now)
            if err != nil {
                return err
            }
        }
    }
    return nil
}

func (s *Store) Allocate(ctx context.Context, namespace string, step int64) (segment.Range, error) {
    for attempt := 0; attempt < 3; attempt++ {
        var maxID, version int64
        if err := s.db.QueryRowContext(ctx, `SELECT max_id, version FROM id_segment WHERE namespace=?`, namespace).Scan(&maxID, &version); err != nil {
            return segment.Range{}, err
        }
        result, err := s.db.ExecContext(ctx, `UPDATE id_segment SET max_id=max_id+step, version=version+1, updated_at=? WHERE namespace=? AND version=?`, time.Now(), namespace, version)
        if err != nil {
            return segment.Range{}, err
        }
        affected, err := result.RowsAffected()
        if err != nil {
            return segment.Range{}, err
        }
        if affected == 1 {
            return segment.Range{Start: maxID + 1, End: maxID + step}, nil
        }
    }
    return segment.Range{}, fmt.Errorf("allocate segment conflict: namespace=%s", namespace)
}

func (s *Store) SaveIssueLog(ctx context.Context, entry audit.IssueLog) error {
    _, err := s.db.ExecContext(ctx, `INSERT INTO id_issue_log(request_id,namespace,caller,issue_type,issued_value,error_code,error_message,created_at)
        VALUES(?,?,?,?,?,?,?,?)`,
        entry.RequestID, entry.Namespace, entry.Caller, entry.IssueType, entry.IssuedValue, entry.ErrorCode, entry.ErrorMessage, entry.CreatedAt)
    return err
}
```

Append these registry and worker lease methods to `mysql/mysql.go`:

```go
func (s *Store) Get(ctx context.Context, namespace string) (idgen.NamespaceConfig, error) {
    var cfg idgen.NamespaceConfig
    var prefix sql.NullString
    err := s.db.QueryRowContext(ctx, `SELECT namespace,biz_domain,id_type,generator_type,prefix,expose_scope,step,max_capacity,owner_team,status
        FROM id_namespace WHERE namespace=?`, namespace).Scan(
        &cfg.Namespace, &cfg.BizDomain, &cfg.IDType, &cfg.GeneratorType, &prefix,
        &cfg.ExposeScope, &cfg.Step, &cfg.MaxCapacity, &cfg.OwnerTeam, &cfg.Status,
    )
    if err == sql.ErrNoRows {
        return idgen.NamespaceConfig{}, idgen.NewError(idgen.ErrNamespaceNotFound, namespace, "namespace is not registered", false)
    }
    if err != nil {
        return idgen.NamespaceConfig{}, err
    }
    cfg.Prefix = prefix.String
    if cfg.Status != idgen.NamespaceEnabled {
        return idgen.NamespaceConfig{}, idgen.NewError(idgen.ErrNamespaceDisabled, namespace, "namespace is not enabled", false)
    }
    return cfg, nil
}

func (s *Store) List(ctx context.Context) ([]idgen.NamespaceConfig, error) {
    rows, err := s.db.QueryContext(ctx, `SELECT namespace,biz_domain,id_type,generator_type,prefix,expose_scope,step,max_capacity,owner_team,status
        FROM id_namespace ORDER BY namespace`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    result := make([]idgen.NamespaceConfig, 0)
    for rows.Next() {
        var cfg idgen.NamespaceConfig
        var prefix sql.NullString
        if err := rows.Scan(&cfg.Namespace, &cfg.BizDomain, &cfg.IDType, &cfg.GeneratorType, &prefix, &cfg.ExposeScope, &cfg.Step, &cfg.MaxCapacity, &cfg.OwnerTeam, &cfg.Status); err != nil {
            return nil, err
        }
        cfg.Prefix = prefix.String
        result = append(result, cfg)
    }
    return result, rows.Err()
}

func (s *Store) AcquireWorker(ctx context.Context, regionID int64, datacenterCode, instanceID string, ttl time.Duration) (int64, string, error) {
    now := time.Now()
    until := now.Add(ttl)
    for workerID := int64(0); workerID < 32; workerID++ {
        token := lease.NewLeaseToken()
        result, err := s.db.ExecContext(ctx, `UPDATE id_worker
            SET instance_id=?, lease_token=?, lease_until=?, heartbeat_at=?, status='ACTIVE', updated_at=?
            WHERE region_id=? AND worker_id=? AND (lease_until<? OR status<>'ACTIVE' OR instance_id=?)`,
            instanceID, token, until, now, now, regionID, workerID, now, instanceID)
        if err != nil {
            return 0, "", err
        }
        affected, err := result.RowsAffected()
        if err != nil {
            return 0, "", err
        }
        if affected == 1 {
            return workerID, token, nil
        }

        result, err = s.db.ExecContext(ctx, `INSERT IGNORE INTO id_worker(worker_id,region_id,datacenter_code,instance_id,lease_token,lease_until,heartbeat_at,status,created_at,updated_at)
            VALUES(?,?,?,?,?,?,?,?,?,?)`,
            workerID, regionID, datacenterCode, instanceID, token, until, now, "ACTIVE", now, now)
        if err != nil {
            return 0, "", err
        }
        affected, err = result.RowsAffected()
        if err != nil {
            return 0, "", err
        }
        if affected == 1 {
            return workerID, token, nil
        }
    }
    return 0, "", fmt.Errorf("no worker lease available for region %d", regionID)
}

func (s *Store) RenewWorker(ctx context.Context, regionID, workerID int64, instanceID, leaseToken string, ttl time.Duration) error {
    now := time.Now()
    result, err := s.db.ExecContext(ctx, `UPDATE id_worker
        SET lease_until=?, heartbeat_at=?, updated_at=?
        WHERE region_id=? AND worker_id=? AND instance_id=? AND lease_token=? AND status='ACTIVE' AND lease_until>?`,
        now.Add(ttl), now, now, regionID, workerID, instanceID, leaseToken, now)
    if err != nil {
        return err
    }
    affected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if affected != 1 {
        return fmt.Errorf("worker lease lost: region=%d worker=%d", regionID, workerID)
    }
    return nil
}

func (s *Store) ReleaseWorker(ctx context.Context, regionID, workerID int64, instanceID, leaseToken string) error {
    _, err := s.db.ExecContext(ctx, `UPDATE id_worker
        SET status='EXPIRED', updated_at=?
        WHERE region_id=? AND worker_id=? AND instance_id=? AND lease_token=?`,
        time.Now(), regionID, workerID, instanceID, leaseToken)
    return err
}
```

- [ ] **Step 3: Run compile tests**

```bash
cd ecommerce-book/example-codes/common-services
gofmt -w internal/infrastructure/lease internal/infrastructure/mysql
go test ./...
```

Expected: all tests compile and pass. MySQL code is compile-tested without requiring a live MySQL connection.

- [ ] **Step 4: Commit MySQL and lease**

```bash
git add ecommerce-book/example-codes/common-services/internal/infrastructure/lease ecommerce-book/example-codes/common-services/internal/infrastructure/mysql
git commit -m "Add MySQL persistence and worker lease contracts"
```

## Task 10: Bootstrap And Server Entrypoint

**Files:**
- Create: `ecommerce-book/example-codes/common-services/internal/bootstrap/app.go`
- Create: `ecommerce-book/example-codes/common-services/cmd/id-server/main.go`

- [ ] **Step 1: Implement bootstrap**

Create `bootstrap/app.go`:

```go
package bootstrap

import (
    "context"
    "database/sql"
    "fmt"
    "net/http"
    "os"
    "strconv"
    "time"

    "common-services/internal/idgen/formatter"
    "common-services/internal/idgen/router"
    "common-services/internal/idgen/segment"
    "common-services/internal/idgen/snowflake"
    ulidgen "common-services/internal/idgen/ulid"
    "common-services/internal/infrastructure/memory"
    "common-services/internal/infrastructure/metrics"
    httpapi "common-services/internal/interfaces/http"
)

type App struct {
    Addr   string
    Server *http.Server
    DB     *sql.DB
}

func NewApp(ctx context.Context) (*App, error) {
    addr := getenv("ID_SERVICE_ADDR", ":8090")
    maxBatch := getenvInt("ID_MAX_BATCH_SIZE", 1000)
    regionID := int64(getenvInt("ID_REGION_ID", 1))
    store := memory.NewStore()
    seg := segment.NewGenerator(store.Segment)
    sf := snowflake.NewGenerator(snowflake.Config{Epoch: defaultEpoch()}, snowflake.RealClock{}, &snowflake.StaticLease{ReadyValue: true, WorkerIDValue: 1, RegionIDValue: regionID})
    svc := router.New(store, seg, sf, ulidgen.NewGenerator(), formatter.NewBusinessNumberFormatter(), maxBatch)
    handler := httpapi.NewHandler(svc, store, metrics.NewRecorder())
    server := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
    return &App{Addr: addr, Server: server}, nil
}

func getenv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}

func getenvInt(key string, fallback int) int {
    raw := os.Getenv(key)
    if raw == "" {
        return fallback
    }
    v, err := strconv.Atoi(raw)
    if err != nil {
        return fallback
    }
    return v
}

func defaultEpoch() time.Time {
    t, err := time.Parse(time.RFC3339, getenv("ID_SNOWFLAKE_EPOCH", "2026-01-01T00:00:00Z"))
    if err != nil {
        panic(fmt.Sprintf("invalid ID_SNOWFLAKE_EPOCH: %v", err))
    }
    return t
}
```

The bootstrap starts in memory mode by default. When `ID_MYSQL_DSN` is set, wire the MySQL store and a lease manager from `internal/infrastructure/lease` before constructing the Snowflake generator.

- [ ] **Step 2: Implement main**

Create `cmd/id-server/main.go`:

```go
package main

import (
    "context"
    "log"

    "common-services/internal/bootstrap"
)

func main() {
    ctx := context.Background()
    app, err := bootstrap.NewApp(ctx)
    if err != nil {
        log.Fatalf("bootstrap id service: %v", err)
    }
    log.Printf("id-service listening on %s", app.Addr)
    if err := app.Server.ListenAndServe(); err != nil {
        log.Fatalf("id-service stopped: %v", err)
    }
}
```

- [ ] **Step 3: Run full tests and start compile**

```bash
cd ecommerce-book/example-codes/common-services
gofmt -w internal/bootstrap cmd/id-server
go test ./...
go test ./cmd/id-server
```

Expected: all tests pass and server entrypoint compiles.

- [ ] **Step 4: Commit bootstrap**

```bash
git add ecommerce-book/example-codes/common-services/internal/bootstrap ecommerce-book/example-codes/common-services/cmd/id-server
git commit -m "Add ID service bootstrap and server"
```

## Task 11: README And API Smoke Verification

**Files:**
- Create: `ecommerce-book/example-codes/common-services/README.md`

- [ ] **Step 1: Write README**

Create README with these sections:

```markdown
# Common Services - ID Service

这个示例实现《附录H 全局 ID 体系设计》中的公共 ID 服务。

## 能力

- Registry 统一管理 namespace。
- Segment 生成 `product.item`、`product.spu`、`product.sku`。
- Snowflake 生成 `trade.order`、`trade.payment`、`trade.refund`。
- ULID 生成 `supply.draft`、`checkout.session`、`event.outbox`。
- Business Number Formatter 生成 `ORD20260429...` 这类外部业务单号。
- HTTP API 暴露单个发号、批量发号、namespace 查询、健康检查和指标。

## 运行

```bash
go test ./...
go run ./cmd/id-server
```

默认使用内存模式，便于读者直接运行。配置 `ID_MYSQL_DSN` 后可以切换到 MySQL 持久化模式。

## API 示例

```bash
curl -X POST http://localhost:8090/api/v1/ids/next \
  -H 'Content-Type: application/json' \
  -d '{"namespace":"trade.order","caller":"order-service","request_id":"req-1"}'
```

```bash
curl -X POST http://localhost:8090/api/v1/ids/batch \
  -H 'Content-Type: application/json' \
  -d '{"namespace":"product.sku","caller":"product-service","request_id":"req-2","count":5}'
```

## 生产语义

ID 服务允许跳号，不允许复用。业务表仍然需要唯一索引兜底。对外单号不直接暴露内部递增 ID。
```
```

- [ ] **Step 2: Run full common-services tests**

```bash
cd ecommerce-book/example-codes/common-services
go test ./...
```

Expected: all tests pass.

- [ ] **Step 3: Commit README**

```bash
git add ecommerce-book/example-codes/common-services/README.md
git commit -m "Document common services ID service"
```

## Task 12: Final Verification And Book Build

**Files:**
- Verify: `ecommerce-book/example-codes/common-services/...`
- Verify: repository root build

- [ ] **Step 1: Verify common-services**

Run:

```bash
cd ecommerce-book/example-codes/common-services
go test ./...
```

Expected: all packages pass.

- [ ] **Step 2: Verify server compiles**

Run:

```bash
cd ecommerce-book/example-codes/common-services
go test ./cmd/id-server
```

Expected: package compiles and passes.

- [ ] **Step 3: Verify repository build**

Run from repo root:

```bash
npm run clean && npm run build
```

Expected: Hexo build completes successfully.

- [ ] **Step 4: Inspect git status**

Run:

```bash
git status --short
```

Expected: only unrelated pre-existing product-service changes remain unstaged, or the working tree is clean for files touched by this plan.

- [ ] **Step 5: Final implementation commit if needed**

If verification changed generated files that should be committed, commit only relevant `common-services` and plan-tracking changes:

```bash
git add ecommerce-book/example-codes/common-services docs/superpowers/plans/2026-04-29-common-services-id-service-implementation.md
git commit -m "Add production grade common ID service example"
```

If all task commits already captured the implementation, skip this commit and report the latest commit list.

## Self-Review

Spec coverage:

- Namespace governance is covered by Task 2 and used by Task 6 and Task 8.
- Segment generator is covered by Task 4 and MySQL allocation by Task 9.
- Snowflake generator, region bits, worker bits, lease readiness, and clock rollback are covered by Task 5 and Task 9.
- ULID generation and controlled prefixes are covered by Task 3 and routed in Task 6.
- Business number formatting is covered by Task 3 and Task 6.
- HTTP APIs, health, ready, metrics, and namespace listing are covered by Task 8.
- MySQL schema and default seeding are covered by Task 9.
- README and Appendix H mapping are covered by Task 11.
- Verification is covered by Task 12.

Type consistency:

- Shared request and result types are defined in `internal/idgen/types.go`.
- Router implements `idgen.Service`.
- Registry implementations satisfy `idgen.Registry`.
- Segment store returns `segment.Range`.
- Snowflake lease contract exposes `Ready`, `WorkerID`, and `RegionID`.

Known implementation adjustment:

- Task 3 includes the correct ULID encoding return value in the step text. Use `return string(out[:])` when implementing.
- Task 10 describes memory mode first and MySQL mode when `ID_MYSQL_DSN` is set. Keep MySQL optional so `go test ./...` does not require a local database.
