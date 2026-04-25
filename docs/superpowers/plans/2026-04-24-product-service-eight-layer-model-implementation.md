# Product Service Eight-Layer Model Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a runnable eight-layer product transaction model demo to `ecommerce-book/example-codes/product-service`.

**Architecture:** Keep the existing DDD demo intact and add a new runtime-context path beside it. Domain files define category capability, eight-layer context, and strategy interfaces; infrastructure supplies in-memory sample repositories; application service orchestrates strategy execution; HTTP exposes the context for Topup, Gift Card, Flight, and Hotel.

**Tech Stack:** Go 1.21 standard library, existing net/http server, in-memory repositories.

---

### Task 1: Domain Model And Strategy Contract

**Files:**
- Create: `ecommerce-book/example-codes/product-service/internal/domain/category_capability.go`
- Create: `ecommerce-book/example-codes/product-service/internal/domain/runtime_context.go`
- Create: `ecommerce-book/example-codes/product-service/internal/domain/category_strategy.go`
- Test: `ecommerce-book/example-codes/product-service/internal/domain/runtime_context_test.go`

- [ ] Write tests for capability lookup expectations and runtime-context layer completeness.
- [ ] Run `go test ./internal/domain` and verify the tests fail because the new types are missing.
- [ ] Add the domain types and interfaces.
- [ ] Run `go test ./internal/domain` and verify the tests pass.

### Task 2: Category Strategies

**Files:**
- Create: `ecommerce-book/example-codes/product-service/internal/domain/strategy/topup_strategy.go`
- Create: `ecommerce-book/example-codes/product-service/internal/domain/strategy/giftcard_strategy.go`
- Create: `ecommerce-book/example-codes/product-service/internal/domain/strategy/flight_strategy.go`
- Create: `ecommerce-book/example-codes/product-service/internal/domain/strategy/hotel_strategy.go`
- Test: `ecommerce-book/example-codes/product-service/internal/domain/strategy/category_strategy_test.go`

- [ ] Write tests that build contexts for Topup, Gift Card, Flight, and Hotel.
- [ ] Run `go test ./internal/domain/strategy` and verify the tests fail because strategies are missing.
- [ ] Implement the four strategies with category-specific availability, input, booking, fulfillment, and refund rules.
- [ ] Run `go test ./internal/domain/strategy` and verify the tests pass.

### Task 3: Runtime Repository And Application Service

**Files:**
- Create: `ecommerce-book/example-codes/product-service/internal/infrastructure/persistence/capability_repository.go`
- Create: `ecommerce-book/example-codes/product-service/internal/application/dto/runtime_context_dto.go`
- Create: `ecommerce-book/example-codes/product-service/internal/application/service/runtime_context_service.go`
- Test: `ecommerce-book/example-codes/product-service/internal/application/service/runtime_context_service_test.go`

- [ ] Write tests for `BuildRuntimeContext` using sample SKU/category pairs.
- [ ] Run `go test ./internal/application/service` and verify the tests fail because the service is missing.
- [ ] Implement in-memory sample data, DTO conversion, and the runtime-context application service.
- [ ] Run `go test ./internal/application/service` and verify the tests pass.

### Task 4: HTTP API And Demo Wiring

**Files:**
- Create: `ecommerce-book/example-codes/product-service/internal/interfaces/http/runtime_context_handler.go`
- Modify: `ecommerce-book/example-codes/product-service/internal/interfaces/http/product_handler.go`
- Modify: `ecommerce-book/example-codes/product-service/cmd/main.go`
- Test: `ecommerce-book/example-codes/product-service/internal/interfaces/http/runtime_context_handler_test.go`

- [ ] Write HTTP handler tests for `/api/v1/products/runtime-context`.
- [ ] Run `go test ./internal/interfaces/http` and verify the tests fail because the route is missing.
- [ ] Wire runtime-context service into the HTTP handler and dependency container.
- [ ] Run `go test ./internal/interfaces/http` and verify the tests pass.

### Task 5: Documentation And Full Verification

**Files:**
- Create: `ecommerce-book/example-codes/product-service/EIGHT_LAYER_MODEL.md`
- Modify: `ecommerce-book/example-codes/product-service/README.md`
- Modify: `ecommerce-book/src/part3/chapter16.md`

- [ ] Document the eight-layer model, sample APIs, and production-table mapping.
- [ ] Update the chapter's example-code mapping table if new files should be visible to readers.
- [ ] Run `go test ./...` from `ecommerce-book/example-codes/product-service`.
- [ ] Run `npm run clean` and `npm run build` from repository root.
