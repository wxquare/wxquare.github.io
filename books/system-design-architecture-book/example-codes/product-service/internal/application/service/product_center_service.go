package service

import (
	"context"
	"fmt"
	"time"

	"product-service/internal/application/dto"
	"product-service/internal/domain"
)

type ProductCenterRepository interface {
	NextItemID(ctx context.Context) int64
	GetCurrentByItemID(ctx context.Context, itemID int64) (*domain.PublishedProduct, error)
	GetCurrentBySKUID(ctx context.Context, skuID int64) (*domain.PublishedProduct, error)
	SavePublish(ctx context.Context, product *domain.PublishedProduct, snapshot *domain.ProductSnapshot, outbox *domain.ProductOutboxEvent) error
	GetSnapshot(ctx context.Context, itemID int64, publishVersion int64) (*domain.ProductSnapshot, error)
	ListOutbox(ctx context.Context, status domain.OutboxStatus) ([]domain.ProductOutboxEvent, error)
}

type ProductCenterService struct {
	repo ProductCenterRepository
}

func NewProductCenterService(repo ProductCenterRepository) *ProductCenterService {
	return &ProductCenterService{repo: repo}
}

func (s *ProductCenterService) PublishProductVersion(ctx context.Context, req dto.PublishProductVersionRequest) (*dto.PublishProductVersionResponse, error) {
	cmd := toPublishCommand(req)
	result, err := s.PublishCommand(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return &dto.PublishProductVersionResponse{
		ItemID:         result.ItemID,
		SKUID:          result.SKUID,
		PublishVersion: result.PublishVersion,
		SnapshotID:     result.SnapshotID,
		OutboxEventID:  result.OutboxEventID,
		Status:         string(result.Status),
		InventoryKey:   result.InventoryKey,
		Message:        "商品版本发布成功；搜索、缓存、库存投影通过 Outbox 最终一致刷新",
	}, nil
}

func (s *ProductCenterService) PublishCommand(ctx context.Context, cmd domain.PublishProductVersionCommand) (*domain.PublishProductVersionResult, error) {
	fmt.Printf("\n🚀 [Application Layer] PublishProductVersion called, OperationID=%s\n", cmd.OperationID)

	if cmd.RequestedAt.IsZero() {
		cmd.RequestedAt = time.Now()
	}
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.findExisting(ctx, cmd.Payload.ItemID, cmd.Payload.SKUID)
	if err != nil {
		return nil, err
	}

	itemID := cmd.Payload.ItemID
	nextVersion := int64(1)
	if existing != nil {
		itemID = existing.ItemID
		if cmd.BasePublishVersion != existing.PublishVersion {
			return nil, fmt.Errorf("publish version conflict: base=%d current=%d", cmd.BasePublishVersion, existing.PublishVersion)
		}
		nextVersion = existing.PublishVersion + 1
	} else {
		if cmd.BasePublishVersion != 0 {
			return nil, fmt.Errorf("new product must publish from base version 0, got %d", cmd.BasePublishVersion)
		}
		if itemID <= 0 {
			itemID = s.repo.NextItemID(ctx)
		}
	}

	snapshotID := fmt.Sprintf("snap_%d_%d", itemID, nextVersion)
	eventID := fmt.Sprintf("evt_product_published_%d_%d_%s", itemID, nextVersion, cmd.OperationID)
	product := domain.BuildPublishedProduct(cmd, itemID, nextVersion, snapshotID, cmd.RequestedAt, existing)
	snapshot := domain.BuildProductSnapshot(cmd, itemID, nextVersion, snapshotID, cmd.RequestedAt)
	outbox := domain.BuildProductPublishedOutbox(cmd, itemID, nextVersion, eventID, cmd.RequestedAt)

	if err := s.repo.SavePublish(ctx, product, snapshot, outbox); err != nil {
		return nil, err
	}

	fmt.Printf("✅ [Application Layer] Product published, ItemID=%d, Version=%d, Outbox=%s\n",
		itemID, nextVersion, eventID)

	return &domain.PublishProductVersionResult{
		ItemID:         itemID,
		SKUID:          product.SKUID,
		PublishVersion: product.PublishVersion,
		SnapshotID:     snapshotID,
		OutboxEventID:  eventID,
		Status:         product.Status,
		InventoryKey:   product.StockConfig.InventoryKey,
	}, nil
}

func (s *ProductCenterService) GetProductSnapshot(ctx context.Context, itemID int64, publishVersion int64) (*dto.ProductSnapshotResponse, error) {
	snapshot, err := s.repo.GetSnapshot(ctx, itemID, publishVersion)
	if err != nil {
		return nil, err
	}
	return toProductSnapshotResponse(snapshot), nil
}

func (s *ProductCenterService) ListOutbox(ctx context.Context, status string) (*dto.ProductOutboxListResponse, error) {
	outboxStatus := domain.OutboxStatus(status)
	if status == "" {
		outboxStatus = domain.OutboxPending
	}
	events, err := s.repo.ListOutbox(ctx, outboxStatus)
	if err != nil {
		return nil, err
	}
	resp := &dto.ProductOutboxListResponse{
		Events: make([]dto.ProductOutboxEventDTO, 0, len(events)),
	}
	for _, event := range events {
		resp.Events = append(resp.Events, dto.ProductOutboxEventDTO{
			EventID:        event.EventID,
			AggregateType:  event.AggregateType,
			AggregateID:    event.AggregateID,
			EventType:      event.EventType,
			PublishVersion: event.PublishVersion,
			Status:         string(event.Status),
			Payload:        event.Payload,
			CreatedAt:      event.CreatedAt.Unix(),
		})
	}
	return resp, nil
}

func (s *ProductCenterService) findExisting(ctx context.Context, itemID int64, skuID int64) (*domain.PublishedProduct, error) {
	if itemID > 0 {
		product, err := s.repo.GetCurrentByItemID(ctx, itemID)
		if err == nil {
			return product, nil
		}
	}
	product, err := s.repo.GetCurrentBySKUID(ctx, skuID)
	if err == nil {
		return product, nil
	}
	return nil, nil
}

func toPublishCommand(req dto.PublishProductVersionRequest) domain.PublishProductVersionCommand {
	requestedAt := time.Now()
	return domain.PublishProductVersionCommand{
		OperationID:        req.OperationID,
		SourceType:         req.SourceType,
		SourceID:           req.SourceID,
		OperatorID:         req.OperatorID,
		BasePublishVersion: req.BasePublishVersion,
		RequestedAt:        requestedAt,
		Payload: domain.ProductPublishPayload{
			ItemID:     req.Payload.ItemID,
			SKUID:      req.Payload.SKUID,
			SPUID:      req.Payload.SPUID,
			SKUCode:    req.Payload.SKUCode,
			Title:      req.Payload.Title,
			CategoryID: req.Payload.CategoryID,
			BasePrice: domain.Money{
				Amount:   req.Payload.BasePrice.Amount,
				Currency: req.Payload.BasePrice.Currency,
			},
			Attributes: req.Payload.Attributes,
			Images:     req.Payload.Images,
			Resource: domain.ProductCenterResource{
				ResourceType: req.Payload.Resource.ResourceType,
				ResourceID:   req.Payload.Resource.ResourceID,
				Name:         req.Payload.Resource.Name,
				Attributes:   req.Payload.Resource.Attributes,
			},
			Offer: domain.ProductCenterOffer{
				OfferID:    req.Payload.Offer.OfferID,
				OfferType:  domain.OfferType(req.Payload.Offer.OfferType),
				RatePlanID: req.Payload.Offer.RatePlanID,
				Price: domain.Money{
					Amount:   req.Payload.Offer.Price.Amount,
					Currency: req.Payload.Offer.Price.Currency,
				},
				Channels:   req.Payload.Offer.Channels,
				Attributes: req.Payload.Offer.Attributes,
			},
			StockConfig: domain.ProductStockConfig{
				InventoryKey:      req.Payload.StockConfig.InventoryKey,
				ManagementType:    domain.InventoryManagementType(req.Payload.StockConfig.ManagementType),
				UnitType:          domain.InventoryUnitType(req.Payload.StockConfig.UnitType),
				DeductTiming:      domain.InventoryDeductTiming(req.Payload.StockConfig.DeductTiming),
				InitialStock:      req.Payload.StockConfig.InitialStock,
				OversellAllowed:   req.Payload.StockConfig.OversellAllowed,
				LowStockThreshold: req.Payload.StockConfig.LowStockThreshold,
				Scope: domain.InventoryScope{
					ScopeType:    req.Payload.StockConfig.Scope.ScopeType,
					ScopeID:      req.Payload.StockConfig.Scope.ScopeID,
					CalendarDate: req.Payload.StockConfig.Scope.CalendarDate,
					BatchID:      req.Payload.StockConfig.Scope.BatchID,
					ChannelID:    req.Payload.StockConfig.Scope.ChannelID,
					SupplierID:   req.Payload.StockConfig.Scope.SupplierID,
				},
				SupplierID:   req.Payload.StockConfig.SupplierID,
				SyncStrategy: req.Payload.StockConfig.SyncStrategy,
			},
			InputSchema: domain.InputSchema{
				SchemaID: req.Payload.InputSchema.SchemaID,
				Fields:   toDomainInputFields(req.Payload.InputSchema.Fields),
			},
			Fulfillment: domain.FulfillmentContract{
				Type:       domain.FulfillmentType(req.Payload.Fulfillment.Type),
				Mode:       req.Payload.Fulfillment.Mode,
				TimeoutSec: req.Payload.Fulfillment.TimeoutSec,
				Attributes: req.Payload.Fulfillment.Attributes,
			},
			RefundRule: domain.RefundRule{
				RuleID:      req.Payload.RefundRule.RuleID,
				Refundable:  req.Payload.RefundRule.Refundable,
				NeedReview:  req.Payload.RefundRule.NeedReview,
				Description: req.Payload.RefundRule.Description,
				Attributes:  req.Payload.RefundRule.Attributes,
			},
		},
	}
}

func toDomainInputFields(fields []dto.InputFieldDTO) []domain.InputField {
	result := make([]domain.InputField, 0, len(fields))
	for _, field := range fields {
		result = append(result, domain.InputField{
			Name:        field.Name,
			Label:       field.Label,
			Type:        field.Type,
			Required:    field.Required,
			Pattern:     field.Pattern,
			Description: field.Description,
		})
	}
	return result
}

func toProductSnapshotResponse(snapshot *domain.ProductSnapshot) *dto.ProductSnapshotResponse {
	return &dto.ProductSnapshotResponse{
		SnapshotID:     snapshot.SnapshotID,
		ItemID:         snapshot.ItemID,
		SKUID:          snapshot.SKUID,
		PublishVersion: snapshot.PublishVersion,
		Payload:        toProductPublishPayloadDTO(snapshot.Payload),
		CreatedAt:      snapshot.CreatedAt.Unix(),
	}
}

func toProductPublishPayloadDTO(payload domain.ProductPublishPayload) dto.ProductPublishPayloadDTO {
	return dto.ProductPublishPayloadDTO{
		ItemID:     payload.ItemID,
		SKUID:      payload.SKUID,
		SPUID:      payload.SPUID,
		SKUCode:    payload.SKUCode,
		Title:      payload.Title,
		CategoryID: payload.CategoryID,
		BasePrice:  dto.MoneyDTO{Amount: payload.BasePrice.Amount, Currency: payload.BasePrice.Currency},
		Attributes: payload.Attributes,
		Images:     payload.Images,
		Resource: dto.ProductCenterResourceDTO{
			ResourceType: payload.Resource.ResourceType,
			ResourceID:   payload.Resource.ResourceID,
			Name:         payload.Resource.Name,
			Attributes:   payload.Resource.Attributes,
		},
		Offer: dto.ProductCenterOfferDTO{
			OfferID:    payload.Offer.OfferID,
			OfferType:  string(payload.Offer.OfferType),
			RatePlanID: payload.Offer.RatePlanID,
			Price:      dto.MoneyDTO{Amount: payload.Offer.Price.Amount, Currency: payload.Offer.Price.Currency},
			Channels:   payload.Offer.Channels,
			Attributes: payload.Offer.Attributes,
		},
		StockConfig: dto.ProductStockConfigDTO{
			InventoryKey:      payload.StockConfig.InventoryKey,
			ManagementType:    string(payload.StockConfig.ManagementType),
			UnitType:          string(payload.StockConfig.UnitType),
			DeductTiming:      string(payload.StockConfig.DeductTiming),
			InitialStock:      payload.StockConfig.InitialStock,
			OversellAllowed:   payload.StockConfig.OversellAllowed,
			LowStockThreshold: payload.StockConfig.LowStockThreshold,
			SupplierID:        payload.StockConfig.SupplierID,
			SyncStrategy:      payload.StockConfig.SyncStrategy,
			Scope: dto.InventoryScopeDTO{
				ScopeType:    payload.StockConfig.Scope.ScopeType,
				ScopeID:      payload.StockConfig.Scope.ScopeID,
				CalendarDate: payload.StockConfig.Scope.CalendarDate,
				BatchID:      payload.StockConfig.Scope.BatchID,
				ChannelID:    payload.StockConfig.Scope.ChannelID,
				SupplierID:   payload.StockConfig.Scope.SupplierID,
			},
		},
		InputSchema: dto.InputSchemaDTO{
			SchemaID: payload.InputSchema.SchemaID,
			Fields:   toInputFieldDTOs(payload.InputSchema.Fields),
		},
		Fulfillment: dto.FulfillmentContractDTO{
			Type:       string(payload.Fulfillment.Type),
			Mode:       payload.Fulfillment.Mode,
			TimeoutSec: payload.Fulfillment.TimeoutSec,
			Attributes: payload.Fulfillment.Attributes,
		},
		RefundRule: dto.RefundRuleDTO{
			RuleID:      payload.RefundRule.RuleID,
			Refundable:  payload.RefundRule.Refundable,
			NeedReview:  payload.RefundRule.NeedReview,
			Description: payload.RefundRule.Description,
			Attributes:  payload.RefundRule.Attributes,
		},
	}
}

func toInputFieldDTOs(fields []domain.InputField) []dto.InputFieldDTO {
	result := make([]dto.InputFieldDTO, 0, len(fields))
	for _, field := range fields {
		result = append(result, dto.InputFieldDTO{
			Name:        field.Name,
			Label:       field.Label,
			Type:        field.Type,
			Required:    field.Required,
			Pattern:     field.Pattern,
			Description: field.Description,
		})
	}
	return result
}
