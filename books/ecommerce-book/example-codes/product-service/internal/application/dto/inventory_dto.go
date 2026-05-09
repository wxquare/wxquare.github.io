package dto

type CreateInventoryRequest struct {
	OperationID        string            `json:"operation_id"`
	SourceType         string            `json:"source_type"`
	SourceID           string            `json:"source_id"`
	OperatorID         int64             `json:"operator_id"`
	Reason             string            `json:"reason"`
	BasePublishVersion int64             `json:"base_publish_version"`
	InventoryKey       string            `json:"inventory_key"`
	ItemID             int64             `json:"item_id"`
	SKUID              int64             `json:"sku_id"`
	ManagementType     string            `json:"management_type"`
	UnitType           string            `json:"unit_type"`
	DeductTiming       string            `json:"deduct_timing"`
	Scope              InventoryScopeDTO `json:"scope"`
	SupplierID         int64             `json:"supplier_id,omitempty"`
	SyncStrategy       string            `json:"sync_strategy,omitempty"`
	InitialStock       int               `json:"initial_stock"`
	OversellAllowed    bool              `json:"oversell_allowed"`
	LowStockThreshold  int               `json:"low_stock_threshold"`
}

type InventoryBalanceDTO struct {
	InventoryKey   string `json:"inventory_key"`
	ItemID         int64  `json:"item_id"`
	SKUID          int64  `json:"sku_id"`
	TotalStock     int    `json:"total_stock"`
	AvailableStock int    `json:"available_stock"`
	BookingStock   int    `json:"booking_stock"`
	LockedStock    int    `json:"locked_stock"`
	SoldStock      int    `json:"sold_stock"`
	Version        int64  `json:"version"`
	UpdatedAt      int64  `json:"updated_at"`
}

type CreateInventoryResponse struct {
	InventoryKey string              `json:"inventory_key"`
	Ready        bool                `json:"ready"`
	Balance      InventoryBalanceDTO `json:"balance"`
	Message      string              `json:"message"`
}

type CheckStockRequest struct {
	InventoryKey string `json:"inventory_key"`
	Qty          int    `json:"qty"`
}

type CheckStockResponse struct {
	InventoryKey string `json:"inventory_key"`
	Sellable     bool   `json:"sellable"`
	Available    int    `json:"available"`
	Message      string `json:"message"`
	FromCache    bool   `json:"from_cache"`
}

type ReserveStockRequest struct {
	InventoryKey   string `json:"inventory_key"`
	OrderID        string `json:"order_id"`
	Qty            int    `json:"qty"`
	TTLSeconds     int    `json:"ttl_seconds"`
	IdempotencyKey string `json:"idempotency_key"`
}

type ReserveStockResponse struct {
	ReservationID string `json:"reservation_id"`
	InventoryKey  string `json:"inventory_key"`
	OrderID       string `json:"order_id"`
	Status        string `json:"status"`
	Remaining     int    `json:"remaining"`
	Idempotent    bool   `json:"idempotent"`
}

type ConfirmStockRequest struct {
	InventoryKey  string `json:"inventory_key"`
	OrderID       string `json:"order_id"`
	ReservationID string `json:"reservation_id"`
	EventID       string `json:"event_id"`
}

type ReleaseStockRequest struct {
	InventoryKey  string `json:"inventory_key"`
	OrderID       string `json:"order_id"`
	ReservationID string `json:"reservation_id"`
	EventID       string `json:"event_id"`
	Reason        string `json:"reason"`
}

type AdjustInventoryRequest struct {
	InventoryKey   string `json:"inventory_key"`
	QtyDelta       int    `json:"qty_delta"`
	OperationID    string `json:"operation_id"`
	SourceType     string `json:"source_type"`
	SourceID       string `json:"source_id"`
	OperatorID     int64  `json:"operator_id"`
	Reason         string `json:"reason"`
	IdempotencyKey string `json:"idempotency_key"`
}

type InventoryActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type InventoryLedgerDTO struct {
	LedgerID     string              `json:"ledger_id"`
	InventoryKey string              `json:"inventory_key"`
	OrderID      string              `json:"order_id,omitempty"`
	EventID      string              `json:"event_id,omitempty"`
	ChangeType   string              `json:"change_type"`
	QtyDelta     int                 `json:"qty_delta"`
	Reason       string              `json:"reason,omitempty"`
	OperatorType string              `json:"operator_type"`
	CreatedAt    int64               `json:"created_at"`
	Before       InventoryBalanceDTO `json:"before"`
	After        InventoryBalanceDTO `json:"after"`
}

type InventoryLedgerListResponse struct {
	Ledgers []InventoryLedgerDTO `json:"ledgers"`
}
