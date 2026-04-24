# Product Service Eight-Layer Model Design

## Background

The existing `ecommerce-book/example-codes/product-service` demo already shows a lightweight DDD product service with a `Product` aggregate, repository abstraction, cache, HTTP/gRPC/Event interfaces, and domain events.

The book now introduces an eight-layer product transaction model for heterogeneous digital goods:

1. Product Definition
2. Resource
3. Offer / Rate Plan
4. Availability
5. Input Schema
6. Booking / Lock
7. Fulfillment Contract
8. Refund Rule

The example code should demonstrate how this model can be implemented in a realistic but still readable engineering structure.

## Goals

- Add a complete engineering skeleton for the eight-layer model.
- Show how category capability configuration drives strategy selection.
- Implement example strategies for Topup, Gift Card, Flight, and Hotel.
- Expose an HTTP endpoint that builds a `ProductRuntimeContext` for pre-transaction scenes.
- Add DTOs, repository abstractions, in-memory example persistence, sample data, and documentation.
- Keep the demo runnable without MySQL, Redis, Kafka, or real supplier dependencies.

## Non-Goals

- No real supplier API calls.
- No real MySQL schema migration.
- No real inventory reservation side effects.
- No production-grade pricing, refund, or fulfillment logic.
- No full replacement of the existing `Product` aggregate demo.

## Proposed Structure

```text
ecommerce-book/example-codes/product-service/
├── EIGHT_LAYER_MODEL.md
├── internal/
│   ├── domain/
│   │   ├── category_capability.go
│   │   ├── runtime_context.go
│   │   ├── category_strategy.go
│   │   └── strategy/
│   │       ├── topup_strategy.go
│   │       ├── giftcard_strategy.go
│   │       ├── flight_strategy.go
│   │       └── hotel_strategy.go
│   ├── application/
│   │   ├── dto/runtime_context_dto.go
│   │   └── service/runtime_context_service.go
│   ├── infrastructure/
│   │   └── persistence/capability_repository.go
│   └── interfaces/
│       └── http/runtime_context_handler.go
```

## Domain Design

### CategoryCapability

`CategoryCapability` describes what a category needs at runtime:

- Product model type: single SKU, account-based, resource-based, realtime offer.
- Resource type: none, hotel, flight, gift card brand, biller, merchant.
- Offer type: fixed price, rate plan, realtime quote, bill query.
- Availability type: unlimited, local pool, realtime supplier, seat map.
- Input schema ID.
- Booking mode: none, pre-lock, pay-then-lock, confirm-after-pay.
- Fulfillment type: topup, bill pay, issue code, ticket, booking confirm.
- Refund rule ID.
- Supplier dependency level.

### ProductRuntimeContext

`ProductRuntimeContext` is the read model used by detail, checkout, and order-create preparation. It contains the eight layers as separate fields:

- `ProductDefinition`
- `Resource`
- `Offer`
- `Availability`
- `InputSchema`
- `Booking`
- `Fulfillment`
- `RefundRule`

### CategoryStrategy

`CategoryStrategy` builds category-specific runtime context:

```go
type CategoryStrategy interface {
    CategoryID() int64
    BuildProductDefinition(ctx context.Context, input BuildContextInput) (*ProductDefinition, error)
    BuildResource(ctx context.Context, input BuildContextInput) (*ResourceContext, error)
    BuildOffer(ctx context.Context, input BuildContextInput) (*OfferContext, error)
    CheckAvailability(ctx context.Context, input BuildContextInput) (*AvailabilityContext, error)
    BuildInputSchema(ctx context.Context, input BuildContextInput) (*InputSchema, error)
    BuildBooking(ctx context.Context, input BuildContextInput) (*BookingRequirement, error)
    BuildFulfillment(ctx context.Context, input BuildContextInput) (*FulfillmentContract, error)
    BuildRefundRule(ctx context.Context, input BuildContextInput) (*RefundRule, error)
}
```

The four sample strategies intentionally differ:

| Strategy | Key Behavior |
|----------|--------------|
| Topup | Fixed denomination, unlimited stock, phone-number input, pay-then-topup fulfillment |
| Gift Card | Fixed denomination, local voucher-code pool, issue-code fulfillment, refund before code assignment |
| Flight | Static route/carrier context, realtime quote and availability, pre-lock before order creation |
| Hotel | Static hotel/room context, rate plan, dynamic room availability, confirm-after-pay booking |

## Application Flow

`RuntimeContextService.BuildRuntimeContext` receives `sku_id`, `category_id`, and `scene`.

Flow:

1. Load product sample data.
2. Load category capability.
3. Select `CategoryStrategy` by category ID.
4. Build the eight context layers.
5. Return a DTO suitable for HTTP response and book examples.

The service does not mutate inventory or create orders. It only builds the transaction-preparation context.

## HTTP API

Add:

```http
GET /api/v1/products/:sku_id/runtime-context?category_id=10102&scene=detail
```

Supported sample categories:

| SKU | Category | Product |
|-----|----------|---------|
| 10001 | 10102 | Mobile Topup |
| 10002 | 30105 | Gift Card |
| 40001 | 40102 | Flight |
| 40002 | 40104 | Hotel |

## Persistence

Use an in-memory repository for sample data:

- `CategoryCapabilityRepository`
- `ProductSampleRepository`

This keeps the demo runnable. The documentation will explain how the in-memory data maps to production tables such as:

- `category_capability_tab`
- `product_runtime_rule_tab`
- `product_spu_tab`
- `product_sku_tab`
- `supplier_product_mapping_tab`

## Documentation

Add `EIGHT_LAYER_MODEL.md` with:

- What problem the model solves.
- The eight layers.
- How Topup, Gift Card, Flight, and Hotel differ.
- How to run the API.
- How to evolve from memory repositories to MySQL-backed repositories.

Update `README.md` to link to the new document and mention the runtime-context API.

## Verification

Run from `ecommerce-book/example-codes/product-service`:

```bash
go test ./...
go run cmd/main.go
```

Run from repository root after markdown updates:

```bash
npm run clean
npm run build
```

## Scope Review

This spec is intentionally focused on the example code. It does not require changing the book chapter prose unless the implementation introduces a different structure than the current chapter describes.
