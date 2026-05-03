package service

import (
	"context"
	"fmt"
	"time"

	"product-service/internal/application/dto"
	"product-service/internal/domain"
)

type InventoryRepository interface {
	SaveInventory(ctx context.Context, config *domain.InventoryConfig, balance *domain.InventoryBalance, ledger *domain.InventoryLedger) error
	GetConfig(ctx context.Context, inventoryKey string) (*domain.InventoryConfig, error)
	GetBalance(ctx context.Context, inventoryKey string) (*domain.InventoryBalance, error)
	MutateInventory(ctx context.Context, inventoryKey string, mutate func(record *domain.InventoryRecord) (*domain.InventoryLedger, error)) (*domain.InventoryRecord, error)
	ListLedger(ctx context.Context, inventoryKey string) ([]domain.InventoryLedger, error)
}

type InventoryService struct {
	repo InventoryRepository
}

func NewInventoryService(repo InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) CreateInventory(ctx context.Context, req dto.CreateInventoryRequest) (*dto.CreateInventoryResponse, error) {
	cmd := toCreateInventoryCommand(req)
	result, err := s.CreateInventoryCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return &dto.CreateInventoryResponse{
		InventoryKey: result.InventoryKey,
		Ready:        result.Ready,
		Balance:      toInventoryBalanceDTO(result.Balance),
		Message:      result.Message,
	}, nil
}

func (s *InventoryService) CreateInventoryCommand(ctx context.Context, cmd domain.CreateInventoryCommand) (*domain.CreateInventoryResult, error) {
	fmt.Printf("\n🚀 [Application Layer] CreateInventory called, InventoryKey=%s\n", cmd.Config.InventoryKey)
	if cmd.Config.InventoryKey == "" {
		return nil, fmt.Errorf("inventory_key is required")
	}
	if cmd.Config.ManagementType == "" {
		return nil, fmt.Errorf("management_type is required")
	}
	if cmd.Config.UnitType == "" {
		return nil, fmt.Errorf("unit_type is required")
	}
	now := time.Now()
	config := cmd.Config
	config.SourceType = cmd.SourceType
	config.SourceID = cmd.SourceID
	config.Status = domain.InventoryStatusActive
	config.CreatedAt = now
	config.UpdatedAt = now
	if config.LowStockThreshold == 0 {
		config.LowStockThreshold = 100
	}
	balance := domain.NewInventoryBalance(config, cmd.InitialStock, now)
	before := *balance
	ledger := &domain.InventoryLedger{
		LedgerID:     fmt.Sprintf("ledger_init_%s_%s", config.InventoryKey, cmd.OperationID),
		InventoryKey: config.InventoryKey,
		EventID:      cmd.OperationID,
		ChangeType:   domain.InventoryChangeInbound,
		QtyDelta:     cmd.InitialStock,
		Reason:       cmd.Reason,
		OperatorType: "SUPPLY_OPS",
		CreatedAt:    now,
		Before:       before,
		After:        *balance,
	}

	if err := s.repo.SaveInventory(ctx, &config, balance, ledger); err != nil {
		return nil, err
	}
	fmt.Printf("✅ [Application Layer] Inventory ready, InventoryKey=%s, Available=%d\n",
		config.InventoryKey, balance.AvailableStock)

	return &domain.CreateInventoryResult{
		InventoryKey: config.InventoryKey,
		Ready:        true,
		Balance:      *balance,
		Message:      "库存实例创建成功；后续 Reserve/Confirm/Release 通过库存中心语义 API 执行",
	}, nil
}

func (s *InventoryService) CheckStock(ctx context.Context, req dto.CheckStockRequest) (*dto.CheckStockResponse, error) {
	domainReq := domain.CheckStockRequest{InventoryKey: req.InventoryKey, Qty: req.Qty}
	resp, err := s.CheckStockCommand(ctx, domainReq)
	if err != nil {
		return nil, err
	}
	return &dto.CheckStockResponse{
		InventoryKey: resp.InventoryKey,
		Sellable:     resp.Sellable,
		Available:    resp.Available,
		Message:      resp.Message,
		FromCache:    resp.FromCache,
	}, nil
}

func (s *InventoryService) CheckStockCommand(ctx context.Context, req domain.CheckStockRequest) (*domain.CheckStockResponse, error) {
	if req.Qty <= 0 {
		req.Qty = 1
	}
	config, err := s.repo.GetConfig(ctx, req.InventoryKey)
	if err != nil {
		return nil, err
	}
	balance, err := s.repo.GetBalance(ctx, req.InventoryKey)
	if err != nil {
		return nil, err
	}
	if config.ManagementType == domain.InventoryUnlimited {
		return &domain.CheckStockResponse{
			InventoryKey: req.InventoryKey,
			Sellable:     true,
			Available:    999999,
			Message:      "无限库存：不扣数量，但库存中心仍保留审计入口",
		}, nil
	}
	sellable := config.Status == domain.InventoryStatusActive && balance.AvailableStock >= req.Qty
	message := "可售"
	if !sellable {
		message = fmt.Sprintf("不可售：available=%d requested=%d status=%s", balance.AvailableStock, req.Qty, config.Status)
	}
	return &domain.CheckStockResponse{
		InventoryKey: req.InventoryKey,
		Sellable:     sellable,
		Available:    balance.AvailableStock,
		Message:      message,
	}, nil
}

func (s *InventoryService) ReserveStock(ctx context.Context, req dto.ReserveStockRequest) (*dto.ReserveStockResponse, error) {
	ttl := time.Duration(req.TTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	resp, err := s.ReserveStockCommand(ctx, domain.ReserveStockRequest{
		InventoryKey:   req.InventoryKey,
		OrderID:        req.OrderID,
		Qty:            req.Qty,
		TTL:            ttl,
		IdempotencyKey: req.IdempotencyKey,
		OperatorType:   "ORDER",
	})
	if err != nil {
		return nil, err
	}
	return &dto.ReserveStockResponse{
		ReservationID: resp.ReservationID,
		InventoryKey:  resp.InventoryKey,
		OrderID:       resp.OrderID,
		Status:        string(resp.Status),
		Remaining:     resp.Remaining,
		Idempotent:    resp.Idempotent,
	}, nil
}

func (s *InventoryService) ReserveStockCommand(ctx context.Context, req domain.ReserveStockRequest) (*domain.ReserveStockResponse, error) {
	if req.InventoryKey == "" || req.OrderID == "" {
		return nil, fmt.Errorf("inventory_key and order_id are required")
	}
	if req.Qty <= 0 {
		return nil, fmt.Errorf("qty must be positive")
	}
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = req.OrderID + ":" + req.InventoryKey
	}
	now := time.Now()
	var output domain.ReserveStockResponse
	_, err := s.repo.MutateInventory(ctx, req.InventoryKey, func(record *domain.InventoryRecord) (*domain.InventoryLedger, error) {
		if existing, ok := record.Reservations[req.OrderID]; ok {
			output = domain.ReserveStockResponse{
				ReservationID: existing.ReservationID,
				InventoryKey:  existing.InventoryKey,
				OrderID:       existing.OrderID,
				Status:        existing.Status,
				Remaining:     record.Balance.AvailableStock,
				Idempotent:    true,
			}
			return nil, nil
		}
		before := *record.Balance
		if record.Config.ManagementType != domain.InventoryUnlimited {
			if err := record.Balance.Reserve(req.Qty, record.Config.OversellAllowed, now); err != nil {
				return nil, err
			}
		}
		reservationID := fmt.Sprintf("res_%s_%d", req.OrderID, now.UnixNano())
		reservation := domain.NewInventoryReservation(reservationID, req, now)
		record.Reservations[req.OrderID] = reservation
		after := *record.Balance
		output = domain.ReserveStockResponse{
			ReservationID: reservation.ReservationID,
			InventoryKey:  reservation.InventoryKey,
			OrderID:       reservation.OrderID,
			Status:        reservation.Status,
			Remaining:     record.Balance.AvailableStock,
		}
		return &domain.InventoryLedger{
			LedgerID:     fmt.Sprintf("ledger_reserve_%s_%s_%d", req.OrderID, req.InventoryKey, now.UnixNano()),
			InventoryKey: req.InventoryKey,
			OrderID:      req.OrderID,
			EventID:      req.IdempotencyKey,
			ChangeType:   domain.InventoryChangeReserve,
			QtyDelta:     -req.Qty,
			Reason:       "create order reserve",
			OperatorType: req.OperatorType,
			CreatedAt:    now,
			Before:       before,
			After:        after,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (s *InventoryService) ConfirmStock(ctx context.Context, req dto.ConfirmStockRequest) (*dto.InventoryActionResponse, error) {
	err := s.ConfirmStockCommand(ctx, domain.ConfirmStockRequest{
		InventoryKey:  req.InventoryKey,
		OrderID:       req.OrderID,
		ReservationID: req.ReservationID,
		EventID:       req.EventID,
	})
	if err != nil {
		return nil, err
	}
	return &dto.InventoryActionResponse{Success: true, Message: "库存确认成功"}, nil
}

func (s *InventoryService) ConfirmStockCommand(ctx context.Context, req domain.ConfirmStockRequest) error {
	now := time.Now()
	_, err := s.repo.MutateInventory(ctx, req.InventoryKey, func(record *domain.InventoryRecord) (*domain.InventoryLedger, error) {
		reservation, err := findReservation(record, req.OrderID, req.ReservationID)
		if err != nil {
			return nil, err
		}
		if reservation.Status == domain.ReservationConfirmed {
			return nil, nil
		}
		before := *record.Balance
		if record.Config.ManagementType != domain.InventoryUnlimited {
			if err := record.Balance.Confirm(reservation.Qty, now); err != nil {
				return nil, err
			}
		}
		if err := reservation.Confirm(now); err != nil {
			return nil, err
		}
		return &domain.InventoryLedger{
			LedgerID:     fmt.Sprintf("ledger_confirm_%s_%s_%d", reservation.OrderID, req.InventoryKey, now.UnixNano()),
			InventoryKey: req.InventoryKey,
			OrderID:      reservation.OrderID,
			EventID:      req.EventID,
			ChangeType:   domain.InventoryChangeConfirm,
			QtyDelta:     -reservation.Qty,
			Reason:       "payment confirmed",
			OperatorType: "PAYMENT",
			CreatedAt:    now,
			Before:       before,
			After:        *record.Balance,
		}, nil
	})
	return err
}

func (s *InventoryService) ReleaseStock(ctx context.Context, req dto.ReleaseStockRequest) (*dto.InventoryActionResponse, error) {
	err := s.ReleaseStockCommand(ctx, domain.ReleaseStockRequest{
		InventoryKey:  req.InventoryKey,
		OrderID:       req.OrderID,
		ReservationID: req.ReservationID,
		EventID:       req.EventID,
		Reason:        req.Reason,
	})
	if err != nil {
		return nil, err
	}
	return &dto.InventoryActionResponse{Success: true, Message: "库存释放成功"}, nil
}

func (s *InventoryService) ReleaseStockCommand(ctx context.Context, req domain.ReleaseStockRequest) error {
	now := time.Now()
	_, err := s.repo.MutateInventory(ctx, req.InventoryKey, func(record *domain.InventoryRecord) (*domain.InventoryLedger, error) {
		reservation, err := findReservation(record, req.OrderID, req.ReservationID)
		if err != nil {
			return nil, err
		}
		if reservation.Status == domain.ReservationReleased || reservation.Status == domain.ReservationExpired || reservation.Status == domain.ReservationCancelled {
			return nil, nil
		}
		before := *record.Balance
		if record.Config.ManagementType != domain.InventoryUnlimited {
			if err := record.Balance.Release(reservation.Qty, now); err != nil {
				return nil, err
			}
		}
		if err := reservation.Release(now); err != nil {
			return nil, err
		}
		return &domain.InventoryLedger{
			LedgerID:     fmt.Sprintf("ledger_release_%s_%s_%d", reservation.OrderID, req.InventoryKey, now.UnixNano()),
			InventoryKey: req.InventoryKey,
			OrderID:      reservation.OrderID,
			EventID:      req.EventID,
			ChangeType:   domain.InventoryChangeRelease,
			QtyDelta:     reservation.Qty,
			Reason:       req.Reason,
			OperatorType: "ORDER",
			CreatedAt:    now,
			Before:       before,
			After:        *record.Balance,
		}, nil
	})
	return err
}

func (s *InventoryService) AdjustInventory(ctx context.Context, req dto.AdjustInventoryRequest) (*dto.InventoryActionResponse, error) {
	err := s.AdjustInventoryCommand(ctx, domain.AdjustInventoryCommand{
		InventoryKey:   req.InventoryKey,
		QtyDelta:       req.QtyDelta,
		OperationID:    req.OperationID,
		SourceType:     req.SourceType,
		SourceID:       req.SourceID,
		OperatorID:     req.OperatorID,
		Reason:         req.Reason,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		return nil, err
	}
	return &dto.InventoryActionResponse{Success: true, Message: "库存调整成功"}, nil
}

func (s *InventoryService) AdjustInventoryCommand(ctx context.Context, req domain.AdjustInventoryCommand) error {
	now := time.Now()
	_, err := s.repo.MutateInventory(ctx, req.InventoryKey, func(record *domain.InventoryRecord) (*domain.InventoryLedger, error) {
		before := *record.Balance
		if err := record.Balance.Adjust(req.QtyDelta, now); err != nil {
			return nil, err
		}
		return &domain.InventoryLedger{
			LedgerID:     fmt.Sprintf("ledger_adjust_%s_%d", req.InventoryKey, now.UnixNano()),
			InventoryKey: req.InventoryKey,
			EventID:      req.OperationID,
			ChangeType:   domain.InventoryChangeAdjust,
			QtyDelta:     req.QtyDelta,
			Reason:       req.Reason,
			OperatorType: req.SourceType,
			CreatedAt:    now,
			Before:       before,
			After:        *record.Balance,
		}, nil
	})
	return err
}

func (s *InventoryService) GetLedger(ctx context.Context, inventoryKey string) (*dto.InventoryLedgerListResponse, error) {
	ledgers, err := s.repo.ListLedger(ctx, inventoryKey)
	if err != nil {
		return nil, err
	}
	resp := &dto.InventoryLedgerListResponse{Ledgers: make([]dto.InventoryLedgerDTO, 0, len(ledgers))}
	for _, ledger := range ledgers {
		resp.Ledgers = append(resp.Ledgers, dto.InventoryLedgerDTO{
			LedgerID:     ledger.LedgerID,
			InventoryKey: ledger.InventoryKey,
			OrderID:      ledger.OrderID,
			EventID:      ledger.EventID,
			ChangeType:   string(ledger.ChangeType),
			QtyDelta:     ledger.QtyDelta,
			Reason:       ledger.Reason,
			OperatorType: ledger.OperatorType,
			CreatedAt:    ledger.CreatedAt.Unix(),
			Before:       toInventoryBalanceDTO(ledger.Before),
			After:        toInventoryBalanceDTO(ledger.After),
		})
	}
	return resp, nil
}

func toCreateInventoryCommand(req dto.CreateInventoryRequest) domain.CreateInventoryCommand {
	return domain.CreateInventoryCommand{
		OperationID:        req.OperationID,
		SourceType:         req.SourceType,
		SourceID:           req.SourceID,
		OperatorID:         req.OperatorID,
		Reason:             req.Reason,
		BasePublishVersion: req.BasePublishVersion,
		InitialStock:       req.InitialStock,
		Config: domain.InventoryConfig{
			InventoryKey:      req.InventoryKey,
			ItemID:            req.ItemID,
			SKUID:             req.SKUID,
			ManagementType:    domain.InventoryManagementType(req.ManagementType),
			UnitType:          domain.InventoryUnitType(req.UnitType),
			DeductTiming:      domain.InventoryDeductTiming(req.DeductTiming),
			SupplierID:        req.SupplierID,
			SyncStrategy:      req.SyncStrategy,
			OversellAllowed:   req.OversellAllowed,
			LowStockThreshold: req.LowStockThreshold,
			Scope: domain.InventoryScope{
				ScopeType:    req.Scope.ScopeType,
				ScopeID:      req.Scope.ScopeID,
				CalendarDate: req.Scope.CalendarDate,
				BatchID:      req.Scope.BatchID,
				ChannelID:    req.Scope.ChannelID,
				SupplierID:   req.Scope.SupplierID,
			},
		},
	}
}

func toInventoryBalanceDTO(balance domain.InventoryBalance) dto.InventoryBalanceDTO {
	return dto.InventoryBalanceDTO{
		InventoryKey:   balance.InventoryKey,
		ItemID:         balance.ItemID,
		SKUID:          balance.SKUID,
		TotalStock:     balance.TotalStock,
		AvailableStock: balance.AvailableStock,
		BookingStock:   balance.BookingStock,
		LockedStock:    balance.LockedStock,
		SoldStock:      balance.SoldStock,
		Version:        balance.Version,
		UpdatedAt:      balance.UpdatedAt.Unix(),
	}
}

func findReservation(record *domain.InventoryRecord, orderID string, reservationID string) (*domain.InventoryReservation, error) {
	if orderID != "" {
		if reservation, ok := record.Reservations[orderID]; ok {
			return reservation, nil
		}
	}
	if reservationID != "" {
		for _, reservation := range record.Reservations {
			if reservation.ReservationID == reservationID {
				return reservation, nil
			}
		}
	}
	return nil, fmt.Errorf("reservation not found")
}
