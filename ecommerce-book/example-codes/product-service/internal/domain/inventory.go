package domain

import (
	"errors"
	"fmt"
	"time"
)

// InventoryManagementType describes who owns the inventory fact.
type InventoryManagementType string

const (
	InventorySelfManaged     InventoryManagementType = "SELF_MANAGED"
	InventorySupplierManaged InventoryManagementType = "SUPPLIER_MANAGED"
	InventoryUnlimited       InventoryManagementType = "UNLIMITED"
)

// InventoryUnitType describes how inventory is consumed.
type InventoryUnitType string

const (
	InventoryUnitQuantity InventoryUnitType = "QUANTITY"
	InventoryUnitCode     InventoryUnitType = "CODE"
	InventoryUnitTime     InventoryUnitType = "TIME"
	InventoryUnitBundle   InventoryUnitType = "BUNDLE"
)

// InventoryDeductTiming describes the business point where stock changes.
type InventoryDeductTiming string

const (
	DeductOnOrder           InventoryDeductTiming = "ON_ORDER"
	DeductOnPay             InventoryDeductTiming = "ON_PAY"
	DeductOnFulfillment     InventoryDeductTiming = "ON_FULFILLMENT"
	DeductOnSupplierConfirm InventoryDeductTiming = "ON_SUPPLIER_CONFIRM"
)

type InventoryStatus string

const (
	InventoryStatusActive InventoryStatus = "ACTIVE"
	InventoryStatusLocked InventoryStatus = "LOCKED"
)

// InventoryScope is the reusable scope abstraction from chapter 9.
type InventoryScope struct {
	ScopeType    string `json:"scope_type"`
	ScopeID      string `json:"scope_id"`
	CalendarDate string `json:"calendar_date,omitempty"`
	BatchID      int64  `json:"batch_id,omitempty"`
	ChannelID    string `json:"channel_id,omitempty"`
	SupplierID   int64  `json:"supplier_id,omitempty"`
}

// InventoryConfig is the control-plane contract for one inventory key.
type InventoryConfig struct {
	InventoryKey      string                  `json:"inventory_key"`
	ItemID            int64                   `json:"item_id"`
	SKUID             int64                   `json:"sku_id"`
	Scope             InventoryScope          `json:"scope"`
	ManagementType    InventoryManagementType `json:"management_type"`
	UnitType          InventoryUnitType       `json:"unit_type"`
	DeductTiming      InventoryDeductTiming   `json:"deduct_timing"`
	SupplierID        int64                   `json:"supplier_id,omitempty"`
	SyncStrategy      string                  `json:"sync_strategy,omitempty"`
	OversellAllowed   bool                    `json:"oversell_allowed"`
	LowStockThreshold int                     `json:"low_stock_threshold"`
	Status            InventoryStatus         `json:"status"`
	SourceType        string                  `json:"source_type,omitempty"`
	SourceID          string                  `json:"source_id,omitempty"`
	CreatedAt         time.Time               `json:"created_at"`
	UpdatedAt         time.Time               `json:"updated_at"`
}

// InventoryBalance is the aggregate view used by Check/Reserve/Confirm/Release.
type InventoryBalance struct {
	InventoryKey     string    `json:"inventory_key"`
	ItemID           int64     `json:"item_id"`
	SKUID            int64     `json:"sku_id"`
	TotalStock       int       `json:"total_stock"`
	AvailableStock   int       `json:"available_stock"`
	BookingStock     int       `json:"booking_stock"`
	LockedStock      int       `json:"locked_stock"`
	SoldStock        int       `json:"sold_stock"`
	SupplierStock    int       `json:"supplier_stock"`
	SupplierSyncTime time.Time `json:"supplier_sync_time,omitempty"`
	Version          int64     `json:"version"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func NewInventoryBalance(config InventoryConfig, initialStock int, now time.Time) *InventoryBalance {
	if initialStock < 0 {
		initialStock = 0
	}
	return &InventoryBalance{
		InventoryKey:   config.InventoryKey,
		ItemID:         config.ItemID,
		SKUID:          config.SKUID,
		TotalStock:     initialStock,
		AvailableStock: initialStock,
		Version:        1,
		UpdatedAt:      now,
	}
}

func (b *InventoryBalance) Reserve(qty int, oversellAllowed bool, now time.Time) error {
	if qty <= 0 {
		return errors.New("reserve qty must be positive")
	}
	if !oversellAllowed && b.AvailableStock < qty {
		return fmt.Errorf("insufficient stock: available=%d requested=%d", b.AvailableStock, qty)
	}
	b.AvailableStock -= qty
	b.BookingStock += qty
	b.Version++
	b.UpdatedAt = now
	return nil
}

func (b *InventoryBalance) Confirm(qty int, now time.Time) error {
	if qty <= 0 {
		return errors.New("confirm qty must be positive")
	}
	if b.BookingStock < qty {
		return fmt.Errorf("insufficient booking stock: booking=%d requested=%d", b.BookingStock, qty)
	}
	b.BookingStock -= qty
	b.SoldStock += qty
	b.Version++
	b.UpdatedAt = now
	return nil
}

func (b *InventoryBalance) Release(qty int, now time.Time) error {
	if qty <= 0 {
		return errors.New("release qty must be positive")
	}
	if b.BookingStock < qty {
		return fmt.Errorf("insufficient booking stock: booking=%d requested=%d", b.BookingStock, qty)
	}
	b.BookingStock -= qty
	b.AvailableStock += qty
	b.Version++
	b.UpdatedAt = now
	return nil
}

func (b *InventoryBalance) Adjust(delta int, now time.Time) error {
	if delta == 0 {
		return errors.New("adjust delta must not be zero")
	}
	if delta < 0 && b.AvailableStock+delta < 0 {
		return fmt.Errorf("adjust would make available stock negative: available=%d delta=%d", b.AvailableStock, delta)
	}
	b.TotalStock += delta
	b.AvailableStock += delta
	b.Version++
	b.UpdatedAt = now
	return nil
}

func (b *InventoryBalance) Lock(qty int, now time.Time) error {
	if qty <= 0 {
		return errors.New("lock qty must be positive")
	}
	if b.AvailableStock < qty {
		return fmt.Errorf("insufficient stock to lock: available=%d requested=%d", b.AvailableStock, qty)
	}
	b.AvailableStock -= qty
	b.LockedStock += qty
	b.Version++
	b.UpdatedAt = now
	return nil
}

func (b *InventoryBalance) InvariantHolds() bool {
	return b.TotalStock == b.AvailableStock+b.BookingStock+b.LockedStock+b.SoldStock
}

type ReservationStatus string

const (
	ReservationReserved  ReservationStatus = "RESERVED"
	ReservationConfirmed ReservationStatus = "CONFIRMED"
	ReservationReleased  ReservationStatus = "RELEASED"
	ReservationExpired   ReservationStatus = "EXPIRED"
	ReservationCancelled ReservationStatus = "CANCELLED"
)

type InventoryReservation struct {
	ReservationID  string            `json:"reservation_id"`
	InventoryKey   string            `json:"inventory_key"`
	OrderID        string            `json:"order_id"`
	Qty            int               `json:"qty"`
	Status         ReservationStatus `json:"status"`
	ExpireAt       time.Time         `json:"expire_at,omitempty"`
	IdempotencyKey string            `json:"idempotency_key"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

func NewInventoryReservation(reservationID string, req ReserveStockRequest, now time.Time) *InventoryReservation {
	return &InventoryReservation{
		ReservationID:  reservationID,
		InventoryKey:   req.InventoryKey,
		OrderID:        req.OrderID,
		Qty:            req.Qty,
		Status:         ReservationReserved,
		ExpireAt:       now.Add(req.TTL),
		IdempotencyKey: req.IdempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (r *InventoryReservation) Confirm(now time.Time) error {
	if r.Status == ReservationConfirmed {
		return nil
	}
	if r.Status != ReservationReserved {
		return fmt.Errorf("reservation %s cannot confirm from status %s", r.ReservationID, r.Status)
	}
	r.Status = ReservationConfirmed
	r.UpdatedAt = now
	return nil
}

func (r *InventoryReservation) Release(now time.Time) error {
	if r.Status == ReservationReleased || r.Status == ReservationExpired || r.Status == ReservationCancelled {
		return nil
	}
	if r.Status != ReservationReserved {
		return fmt.Errorf("reservation %s cannot release from status %s", r.ReservationID, r.Status)
	}
	r.Status = ReservationReleased
	r.UpdatedAt = now
	return nil
}

type InventoryChangeType string

const (
	InventoryChangeInbound InventoryChangeType = "INBOUND"
	InventoryChangeReserve InventoryChangeType = "RESERVE"
	InventoryChangeConfirm InventoryChangeType = "CONFIRM"
	InventoryChangeRelease InventoryChangeType = "RELEASE"
	InventoryChangeRefund  InventoryChangeType = "REFUND"
	InventoryChangeLock    InventoryChangeType = "LOCK"
	InventoryChangeAdjust  InventoryChangeType = "ADJUST"
)

type InventoryLedger struct {
	LedgerID     string              `json:"ledger_id"`
	InventoryKey string              `json:"inventory_key"`
	OrderID      string              `json:"order_id,omitempty"`
	EventID      string              `json:"event_id,omitempty"`
	ChangeType   InventoryChangeType `json:"change_type"`
	QtyDelta     int                 `json:"qty_delta"`
	Reason       string              `json:"reason,omitempty"`
	OperatorType string              `json:"operator_type"`
	CreatedAt    time.Time           `json:"created_at"`
	Before       InventoryBalance    `json:"before"`
	After        InventoryBalance    `json:"after"`
}

type InventoryRecord struct {
	Config       *InventoryConfig
	Balance      *InventoryBalance
	Reservations map[string]*InventoryReservation
}

type CreateInventoryCommand struct {
	OperationID        string
	SourceType         string
	SourceID           string
	OperatorID         int64
	Reason             string
	BasePublishVersion int64
	Config             InventoryConfig
	InitialStock       int
}

type CreateInventoryResult struct {
	InventoryKey string
	Ready        bool
	Balance      InventoryBalance
	Message      string
}

type CheckStockRequest struct {
	InventoryKey string
	Qty          int
}

type CheckStockResponse struct {
	InventoryKey string
	Sellable     bool
	Available    int
	Message      string
	FromCache    bool
}

type ReserveStockRequest struct {
	InventoryKey   string
	OrderID        string
	Qty            int
	TTL            time.Duration
	IdempotencyKey string
	OperatorType   string
}

type ReserveStockResponse struct {
	ReservationID string
	InventoryKey  string
	OrderID       string
	Status        ReservationStatus
	Remaining     int
	Idempotent    bool
}

type ConfirmStockRequest struct {
	InventoryKey  string
	OrderID       string
	ReservationID string
	EventID       string
}

type ReleaseStockRequest struct {
	InventoryKey  string
	OrderID       string
	ReservationID string
	EventID       string
	Reason        string
}

type AdjustInventoryCommand struct {
	InventoryKey   string
	QtyDelta       int
	OperationID    string
	SourceType     string
	SourceID       string
	OperatorID     int64
	Reason         string
	IdempotencyKey string
}
