package domain

import (
	"errors"
	"fmt"
	"time"
)

type ProductLifecycleStatus string

const (
	ProductLifecyclePublished ProductLifecycleStatus = "PUBLISHED"
	ProductLifecycleOnline    ProductLifecycleStatus = "ONLINE"
	ProductLifecycleOffline   ProductLifecycleStatus = "OFFLINE"
	ProductLifecycleEnded     ProductLifecycleStatus = "ENDED"
	ProductLifecycleBanned    ProductLifecycleStatus = "BANNED"
	ProductLifecycleArchived  ProductLifecycleStatus = "ARCHIVED"
)

type ProductCenterResource struct {
	ResourceType string            `json:"resource_type"`
	ResourceID   string            `json:"resource_id"`
	Name         string            `json:"name"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

type ProductCenterOffer struct {
	OfferID    string            `json:"offer_id"`
	OfferType  OfferType         `json:"offer_type"`
	RatePlanID string            `json:"rate_plan_id,omitempty"`
	Price      Money             `json:"price"`
	SaleStart  time.Time         `json:"sale_start,omitempty"`
	SaleEnd    time.Time         `json:"sale_end,omitempty"`
	Channels   []string          `json:"channels,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// ProductStockConfig is the product-center view of inventory. It stores how to
// connect to inventory, not the live stock fact.
type ProductStockConfig struct {
	InventoryKey      string                  `json:"inventory_key"`
	ManagementType    InventoryManagementType `json:"management_type"`
	UnitType          InventoryUnitType       `json:"unit_type"`
	DeductTiming      InventoryDeductTiming   `json:"deduct_timing"`
	Scope             InventoryScope          `json:"scope"`
	SupplierID        int64                   `json:"supplier_id,omitempty"`
	SyncStrategy      string                  `json:"sync_strategy,omitempty"`
	InitialStock      int                     `json:"initial_stock"`
	OversellAllowed   bool                    `json:"oversell_allowed"`
	LowStockThreshold int                     `json:"low_stock_threshold"`
}

type ProductPublishPayload struct {
	ItemID      int64                 `json:"item_id,omitempty"`
	SKUID       int64                 `json:"sku_id"`
	SPUID       int64                 `json:"spu_id,omitempty"`
	SKUCode     string                `json:"sku_code"`
	Title       string                `json:"title"`
	CategoryID  int64                 `json:"category_id"`
	BasePrice   Money                 `json:"base_price"`
	Attributes  map[string]string     `json:"attributes,omitempty"`
	Images      []string              `json:"images,omitempty"`
	Resource    ProductCenterResource `json:"resource"`
	Offer       ProductCenterOffer    `json:"offer"`
	StockConfig ProductStockConfig    `json:"stock_config"`
	InputSchema InputSchema           `json:"input_schema"`
	Fulfillment FulfillmentContract   `json:"fulfillment"`
	RefundRule  RefundRule            `json:"refund_rule"`
	EffectiveAt time.Time             `json:"effective_at,omitempty"`
	ExpireAt    time.Time             `json:"expire_at,omitempty"`
}

func (p ProductPublishPayload) Validate() error {
	if p.SKUID <= 0 {
		return errors.New("sku_id is required")
	}
	if p.Title == "" {
		return errors.New("title is required")
	}
	if p.CategoryID <= 0 {
		return errors.New("category_id is required")
	}
	if p.BasePrice.Amount < 0 {
		return errors.New("base_price must not be negative")
	}
	if p.StockConfig.InventoryKey == "" {
		return errors.New("stock_config.inventory_key is required")
	}
	if p.StockConfig.ManagementType == "" {
		return errors.New("stock_config.management_type is required")
	}
	if p.StockConfig.UnitType == "" {
		return errors.New("stock_config.unit_type is required")
	}
	if p.StockConfig.DeductTiming == "" {
		return errors.New("stock_config.deduct_timing is required")
	}
	if p.InputSchema.SchemaID == "" {
		return errors.New("input_schema.schema_id is required")
	}
	if p.Fulfillment.Type == "" {
		return errors.New("fulfillment.type is required")
	}
	if p.RefundRule.RuleID == "" {
		return errors.New("refund_rule.rule_id is required")
	}
	return nil
}

type PublishProductVersionCommand struct {
	OperationID        string                `json:"operation_id"`
	SourceType         string                `json:"source_type"`
	SourceID           string                `json:"source_id"`
	OperatorID         int64                 `json:"operator_id"`
	BasePublishVersion int64                 `json:"base_publish_version"`
	Payload            ProductPublishPayload `json:"payload"`
	RequestedAt        time.Time             `json:"requested_at"`
}

func (c PublishProductVersionCommand) Validate() error {
	if c.OperationID == "" {
		return errors.New("operation_id is required")
	}
	if c.SourceType == "" {
		return errors.New("source_type is required")
	}
	return c.Payload.Validate()
}

type PublishedProduct struct {
	ItemID            int64                  `json:"item_id"`
	SKUID             int64                  `json:"sku_id"`
	SPUID             int64                  `json:"spu_id,omitempty"`
	SKUCode           string                 `json:"sku_code"`
	Title             string                 `json:"title"`
	CategoryID        int64                  `json:"category_id"`
	BasePrice         Money                  `json:"base_price"`
	Status            ProductLifecycleStatus `json:"status"`
	PublishVersion    int64                  `json:"publish_version"`
	CurrentSnapshotID string                 `json:"current_snapshot_id"`
	LastPublishID     string                 `json:"last_publish_id"`
	LastOperationID   string                 `json:"last_operation_id"`
	SourceType        string                 `json:"source_type"`
	StockConfig       ProductStockConfig     `json:"stock_config"`
	EffectiveAt       time.Time              `json:"effective_at,omitempty"`
	ExpireAt          time.Time              `json:"expire_at,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type ProductSnapshot struct {
	SnapshotID     string                `json:"snapshot_id"`
	ItemID         int64                 `json:"item_id"`
	SKUID          int64                 `json:"sku_id"`
	PublishVersion int64                 `json:"publish_version"`
	Payload        ProductPublishPayload `json:"payload"`
	CreatedAt      time.Time             `json:"created_at"`
}

type OutboxStatus string

const (
	OutboxPending OutboxStatus = "PENDING"
	OutboxSending OutboxStatus = "SENDING"
	OutboxSent    OutboxStatus = "SENT"
	OutboxFailed  OutboxStatus = "FAILED"
	OutboxDLQ     OutboxStatus = "DLQ"
)

type ProductOutboxEvent struct {
	EventID        string                 `json:"event_id"`
	AggregateType  string                 `json:"aggregate_type"`
	AggregateID    string                 `json:"aggregate_id"`
	EventType      string                 `json:"event_type"`
	PublishVersion int64                  `json:"publish_version"`
	Status         OutboxStatus           `json:"status"`
	Payload        map[string]interface{} `json:"payload"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

type PublishProductVersionResult struct {
	ItemID         int64
	SKUID          int64
	PublishVersion int64
	SnapshotID     string
	OutboxEventID  string
	Status         ProductLifecycleStatus
	InventoryKey   string
}

func BuildPublishedProduct(cmd PublishProductVersionCommand, itemID int64, version int64, snapshotID string, now time.Time, existing *PublishedProduct) *PublishedProduct {
	status := ProductLifecyclePublished
	createdAt := now
	if existing != nil {
		status = existing.Status
		createdAt = existing.CreatedAt
	}

	return &PublishedProduct{
		ItemID:            itemID,
		SKUID:             cmd.Payload.SKUID,
		SPUID:             cmd.Payload.SPUID,
		SKUCode:           cmd.Payload.SKUCode,
		Title:             cmd.Payload.Title,
		CategoryID:        cmd.Payload.CategoryID,
		BasePrice:         cmd.Payload.BasePrice,
		Status:            status,
		PublishVersion:    version,
		CurrentSnapshotID: snapshotID,
		LastPublishID:     fmt.Sprintf("pub_%d_%d", itemID, version),
		LastOperationID:   cmd.OperationID,
		SourceType:        cmd.SourceType,
		StockConfig:       cmd.Payload.StockConfig,
		EffectiveAt:       cmd.Payload.EffectiveAt,
		ExpireAt:          cmd.Payload.ExpireAt,
		CreatedAt:         createdAt,
		UpdatedAt:         now,
	}
}

func BuildProductSnapshot(cmd PublishProductVersionCommand, itemID int64, version int64, snapshotID string, now time.Time) *ProductSnapshot {
	payload := cmd.Payload
	payload.ItemID = itemID
	return &ProductSnapshot{
		SnapshotID:     snapshotID,
		ItemID:         itemID,
		SKUID:          payload.SKUID,
		PublishVersion: version,
		Payload:        payload,
		CreatedAt:      now,
	}
}

func BuildProductPublishedOutbox(cmd PublishProductVersionCommand, itemID int64, version int64, eventID string, now time.Time) *ProductOutboxEvent {
	return &ProductOutboxEvent{
		EventID:        eventID,
		AggregateType:  "PRODUCT_ITEM",
		AggregateID:    fmt.Sprintf("%d", itemID),
		EventType:      "ProductPublished",
		PublishVersion: version,
		Status:         OutboxPending,
		CreatedAt:      now,
		UpdatedAt:      now,
		Payload: map[string]interface{}{
			"item_id":         itemID,
			"sku_id":          cmd.Payload.SKUID,
			"publish_version": version,
			"operation_id":    cmd.OperationID,
			"source_type":     cmd.SourceType,
			"source_id":       cmd.SourceID,
			"inventory_key":   cmd.Payload.StockConfig.InventoryKey,
			"occurred_at":     now.Unix(),
		},
	}
}
