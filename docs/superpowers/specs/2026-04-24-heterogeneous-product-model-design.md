# Heterogeneous Product Model Design

## Context

Chapter 16.6.1 describes the Product Center for a digital goods platform that covers topup, bill payment, gift cards, vouchers, local services, flight, hotel, movie, train, and bus categories. The existing text already explains that these categories differ in SKU shape, pricing, inventory, fulfillment, input forms, and supplier dependency. The missing piece is a decision framework for how an architect decides what belongs in Product Center storage and what should stay in real-time supplier, pricing, order, fulfillment, or after-sale flows.

## Goal

Improve the book section so it is useful both as experience summary and interview material. The design should:

1. Compare multiple industry-style modeling approaches.
2. Explain why standard SPU/SKU alone is insufficient for OTA/O2O/virtual goods.
3. Recommend a hybrid model that combines SPU/SKU, Resource, Offer/Rate Plan, category capability matrix, and runtime context.
4. Clarify how stable data, dynamic supply, and transaction results are stored separately.
5. Avoid company-sensitive names and keep the discussion generic.

## Approaches

### Approach A: Standard SPU/SKU With EAV and ExtInfo

All products are modeled as category, SPU, SKU, attributes, and JSON extensions. This is simple and works well for fixed-denomination categories such as topup, gift cards, and vouchers. It becomes weak for flight, hotel, movie seats, and dynamic bill amounts because real-time quotes, date ranges, and seats are not stable SKUs.

### Approach B: Resource-Centric Product Model

The platform first models business resources, then wraps them into sellable products. Examples include hotel, room type, merchant, outlet, cinema, movie, city, station, airport, biller, and route. This fits OTA and O2O categories better and avoids SKU explosion. However, it does not fully express user input, booking/lock, fulfillment, and refund contracts by itself.

### Approach C: Product Transaction Contract Model

This is the recommended approach. It does not replace SPU/SKU or Resource; it combines them:

```text
Product Definition
  + Resource
  + Offer / Rate Plan
  + Availability
  + Input Schema
  + Booking / Lock
  + Fulfillment Contract
  + Refund Rule
```

The principle is that Product Center unifies business expression and pre-transaction contracts, not all real-time resource states. Stable definitions are stored in product and resource tables. Dynamic supply is resolved by cache or supplier calls. Transaction results are captured in order, fulfillment, and after-sale snapshots.

## Recommended Book Changes

Update `ecommerce-book/src/part3/chapter16.md`:

1. Add a subsection in `16.6.1.2` comparing the three approaches and recommending the hybrid model.
2. Enrich `16.6.1.3` with the recommended layered model:
   - Category
   - Resource
   - SPU
   - SKU
   - Offer / Rate Plan
   - Attribute / EAV
   - ExtInfo JSON
   - Supplier Mapping
   - Category Capability
   - Runtime Context
3. Enrich `16.6.1.4` with resource tables and offer/rate-plan/capability tables.
4. Enrich `16.6.1.5` with category examples that explicitly answer where data is stored:
   - Product/SPU/SKU
   - Resource
   - Offer/Rate Plan
   - Availability
   - Runtime supplier/query/order snapshot

## Scope

This is a writing/design change only. It does not require modifying the example code in this step.

## Self Review

- No placeholders remain.
- The design focuses on Product Center heterogeneous product modeling only.
- It does not introduce company-specific names.
- The recommendation is explicit: use the hybrid transaction contract model.
