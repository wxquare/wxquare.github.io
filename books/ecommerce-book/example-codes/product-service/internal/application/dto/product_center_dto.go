package dto

type ProductCenterResourceDTO struct {
	ResourceType string            `json:"resource_type"`
	ResourceID   string            `json:"resource_id"`
	Name         string            `json:"name"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

type ProductCenterOfferDTO struct {
	OfferID    string            `json:"offer_id"`
	OfferType  string            `json:"offer_type"`
	RatePlanID string            `json:"rate_plan_id,omitempty"`
	Price      MoneyDTO          `json:"price"`
	Channels   []string          `json:"channels,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type InventoryScopeDTO struct {
	ScopeType    string `json:"scope_type"`
	ScopeID      string `json:"scope_id"`
	CalendarDate string `json:"calendar_date,omitempty"`
	BatchID      int64  `json:"batch_id,omitempty"`
	ChannelID    string `json:"channel_id,omitempty"`
	SupplierID   int64  `json:"supplier_id,omitempty"`
}

type ProductStockConfigDTO struct {
	InventoryKey      string            `json:"inventory_key"`
	ManagementType    string            `json:"management_type"`
	UnitType          string            `json:"unit_type"`
	DeductTiming      string            `json:"deduct_timing"`
	Scope             InventoryScopeDTO `json:"scope"`
	SupplierID        int64             `json:"supplier_id,omitempty"`
	SyncStrategy      string            `json:"sync_strategy,omitempty"`
	InitialStock      int               `json:"initial_stock"`
	OversellAllowed   bool              `json:"oversell_allowed"`
	LowStockThreshold int               `json:"low_stock_threshold"`
}

type ProductPublishPayloadDTO struct {
	ItemID      int64                    `json:"item_id,omitempty"`
	SKUID       int64                    `json:"sku_id"`
	SPUID       int64                    `json:"spu_id,omitempty"`
	SKUCode     string                   `json:"sku_code"`
	Title       string                   `json:"title"`
	CategoryID  int64                    `json:"category_id"`
	BasePrice   MoneyDTO                 `json:"base_price"`
	Attributes  map[string]string        `json:"attributes,omitempty"`
	Images      []string                 `json:"images,omitempty"`
	Resource    ProductCenterResourceDTO `json:"resource"`
	Offer       ProductCenterOfferDTO    `json:"offer"`
	StockConfig ProductStockConfigDTO    `json:"stock_config"`
	InputSchema InputSchemaDTO           `json:"input_schema"`
	Fulfillment FulfillmentContractDTO   `json:"fulfillment"`
	RefundRule  RefundRuleDTO            `json:"refund_rule"`
}

type PublishProductVersionRequest struct {
	OperationID        string                   `json:"operation_id"`
	SourceType         string                   `json:"source_type"`
	SourceID           string                   `json:"source_id"`
	OperatorID         int64                    `json:"operator_id"`
	BasePublishVersion int64                    `json:"base_publish_version"`
	Payload            ProductPublishPayloadDTO `json:"payload"`
}

type PublishProductVersionResponse struct {
	ItemID         int64  `json:"item_id"`
	SKUID          int64  `json:"sku_id"`
	PublishVersion int64  `json:"publish_version"`
	SnapshotID     string `json:"snapshot_id"`
	OutboxEventID  string `json:"outbox_event_id"`
	Status         string `json:"status"`
	InventoryKey   string `json:"inventory_key"`
	Message        string `json:"message"`
}

type ProductSnapshotResponse struct {
	SnapshotID     string                   `json:"snapshot_id"`
	ItemID         int64                    `json:"item_id"`
	SKUID          int64                    `json:"sku_id"`
	PublishVersion int64                    `json:"publish_version"`
	Payload        ProductPublishPayloadDTO `json:"payload"`
	CreatedAt      int64                    `json:"created_at"`
}

type ProductOutboxEventDTO struct {
	EventID        string                 `json:"event_id"`
	AggregateType  string                 `json:"aggregate_type"`
	AggregateID    string                 `json:"aggregate_id"`
	EventType      string                 `json:"event_type"`
	PublishVersion int64                  `json:"publish_version"`
	Status         string                 `json:"status"`
	Payload        map[string]interface{} `json:"payload"`
	CreatedAt      int64                  `json:"created_at"`
}

type ProductOutboxListResponse struct {
	Events []ProductOutboxEventDTO `json:"events"`
}
