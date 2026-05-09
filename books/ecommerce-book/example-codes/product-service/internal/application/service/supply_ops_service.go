package service

import (
	"context"
	"fmt"
	"time"

	"product-service/internal/application/dto"
	"product-service/internal/domain"
)

type SupplyRepository interface {
	NextID(ctx context.Context, prefix string) string
	SaveDraft(ctx context.Context, draft *domain.ProductSupplyDraft) error
	GetDraft(ctx context.Context, draftID string) (*domain.ProductSupplyDraft, error)
	SaveStaging(ctx context.Context, staging *domain.ProductSupplyStaging) error
	GetStaging(ctx context.Context, stagingID string) (*domain.ProductSupplyStaging, error)
	SaveQCReview(ctx context.Context, review *domain.ProductQCReview) error
	GetQCReview(ctx context.Context, reviewID string) (*domain.ProductQCReview, error)
	GetQCReviewByStaging(ctx context.Context, stagingID string) (*domain.ProductQCReview, error)
	AppendLog(ctx context.Context, log domain.SupplyOperationLog) error
	ListLogs(ctx context.Context, traceID string) ([]domain.SupplyOperationLog, error)
}

type SupplyOpsService struct {
	repo          SupplyRepository
	productCenter *ProductCenterService
	inventory     *InventoryService
}

func NewSupplyOpsService(repo SupplyRepository, productCenter *ProductCenterService, inventory *InventoryService) *SupplyOpsService {
	return &SupplyOpsService{
		repo:          repo,
		productCenter: productCenter,
		inventory:     inventory,
	}
}

func (s *SupplyOpsService) CreateDraft(ctx context.Context, req dto.CreateSupplyDraftRequest) (*dto.CreateSupplyDraftResponse, error) {
	if req.OperationID == "" {
		req.OperationID = s.repo.NextID(ctx, "op")
	}
	sourceType := domain.SupplySourceType(req.SourceType)
	if sourceType == "" {
		sourceType = domain.SupplySourceLocalOps
	}
	payload := publishPayloadDTOToDomain(req.Payload)
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	now := time.Now()
	draft := &domain.ProductSupplyDraft{
		DraftID:            s.repo.NextID(ctx, "draft"),
		OperationID:        req.OperationID,
		SourceType:         sourceType,
		OperatorID:         req.OperatorID,
		Status:             domain.DraftStatusDraft,
		BasePublishVersion: req.BasePublishVersion,
		Payload:            payload,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := s.repo.SaveDraft(ctx, draft); err != nil {
		return nil, err
	}
	_ = s.appendLog(ctx, draft.OperationID, "DRAFT", draft.DraftID, "DRAFT_CREATED", draft.OperatorID, draft.SourceType, "草稿已创建，尚未进入正式商品中心")
	return &dto.CreateSupplyDraftResponse{
		DraftID:     draft.DraftID,
		OperationID: draft.OperationID,
		Status:      string(draft.Status),
		Message:     "Draft 已创建；此阶段不生成正式 item_id",
	}, nil
}

func (s *SupplyOpsService) SubmitDraft(ctx context.Context, draftID string) (*dto.SubmitSupplyDraftResponse, error) {
	draft, err := s.repo.GetDraft(ctx, draftID)
	if err != nil {
		return nil, err
	}
	if draft.Status != domain.DraftStatusDraft {
		return nil, fmt.Errorf("draft %s cannot submit from status %s", draftID, draft.Status)
	}
	qcPolicy := decideQCPolicy(draft.SourceType)
	status := domain.StagingStatusApproved
	if qcPolicy == domain.QCPolicyRequired {
		status = domain.StagingStatusQCPending
	}
	now := time.Now()
	staging := &domain.ProductSupplyStaging{
		StagingID:          s.repo.NextID(ctx, "staging"),
		DraftID:            draft.DraftID,
		OperationID:        draft.OperationID,
		SourceType:         draft.SourceType,
		OperatorID:         draft.OperatorID,
		Status:             status,
		QCPolicy:           qcPolicy,
		BasePublishVersion: draft.BasePublishVersion,
		Payload:            draft.Payload,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	draft.Status = domain.DraftStatusSubmitted
	draft.UpdatedAt = now
	if err := s.repo.SaveDraft(ctx, draft); err != nil {
		return nil, err
	}
	if err := s.repo.SaveStaging(ctx, staging); err != nil {
		return nil, err
	}

	reviewID := ""
	if qcPolicy == domain.QCPolicyRequired {
		review := &domain.ProductQCReview{
			ReviewID:     s.repo.NextID(ctx, "qc"),
			StagingID:    staging.StagingID,
			Status:       domain.QCReviewPending,
			ReviewPolicy: qcPolicy,
			RiskReasons:  []string{"外部商家来源默认需要 QC"},
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := s.repo.SaveQCReview(ctx, review); err != nil {
			return nil, err
		}
		reviewID = review.ReviewID
		_ = s.appendLog(ctx, draft.OperationID, "QC", review.ReviewID, "QC_CREATED", draft.OperatorID, draft.SourceType, "商家来源提交后生成 QC 审核单")
	}

	_ = s.appendLog(ctx, draft.OperationID, "STAGING", staging.StagingID, "SUBMITTED", draft.OperatorID, draft.SourceType, "Draft 提交并冻结为 Staging")
	return &dto.SubmitSupplyDraftResponse{
		DraftID:    draft.DraftID,
		StagingID:  staging.StagingID,
		QCReviewID: reviewID,
		QCPolicy:   string(qcPolicy),
		Status:     string(staging.Status),
		Message:    "Staging 已生成；本地运营自动准入，商家来源等待 QC",
	}, nil
}

func (s *SupplyOpsService) ApproveQC(ctx context.Context, reviewID string, req dto.ApproveQCRequest) (*dto.ApproveQCResponse, error) {
	review, err := s.repo.GetQCReview(ctx, reviewID)
	if err != nil {
		return nil, err
	}
	if review.Status != domain.QCReviewPending && review.Status != domain.QCReviewReviewing {
		return nil, fmt.Errorf("qc review %s cannot approve from status %s", reviewID, review.Status)
	}
	staging, err := s.repo.GetStaging(ctx, review.StagingID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	review.Status = domain.QCReviewApproved
	review.ReviewerID = req.ReviewerID
	review.UpdatedAt = now
	staging.Status = domain.StagingStatusQCApproved
	staging.UpdatedAt = now
	if err := s.repo.SaveQCReview(ctx, review); err != nil {
		return nil, err
	}
	if err := s.repo.SaveStaging(ctx, staging); err != nil {
		return nil, err
	}
	_ = s.appendLog(ctx, staging.OperationID, "QC", review.ReviewID, "QC_APPROVED", req.ReviewerID, staging.SourceType, "QC 通过，允许进入发布事务")
	return &dto.ApproveQCResponse{
		ReviewID:  review.ReviewID,
		StagingID: staging.StagingID,
		Status:    string(review.Status),
		Message:   "QC 已通过；发布仍需执行 Publish",
	}, nil
}

func (s *SupplyOpsService) PublishStaging(ctx context.Context, stagingID string) (*dto.SupplyPublishResponse, error) {
	staging, err := s.repo.GetStaging(ctx, stagingID)
	if err != nil {
		return nil, err
	}
	if !stagingCanPublish(staging) {
		return nil, fmt.Errorf("staging %s cannot publish from status %s", stagingID, staging.Status)
	}
	if staging.QCPolicy == domain.QCPolicyRequired {
		review, err := s.repo.GetQCReviewByStaging(ctx, stagingID)
		if err != nil {
			return nil, err
		}
		if review.Status != domain.QCReviewApproved {
			return nil, fmt.Errorf("qc review %s is not approved", review.ReviewID)
		}
	}

	publishResult, err := s.productCenter.PublishCommand(ctx, domain.PublishProductVersionCommand{
		OperationID:        staging.OperationID,
		SourceType:         string(staging.SourceType),
		SourceID:           staging.StagingID,
		OperatorID:         staging.OperatorID,
		BasePublishVersion: staging.BasePublishVersion,
		Payload:            staging.Payload,
		RequestedAt:        time.Now(),
	})
	if err != nil {
		return nil, err
	}

	inventoryReady := false
	inventoryMessage := ""
	config := inventoryConfigFromPublishedPayload(staging.Payload, publishResult.ItemID)
	_, inventoryErr := s.inventory.CreateInventoryCommand(ctx, domain.CreateInventoryCommand{
		OperationID:        staging.OperationID,
		SourceType:         "PRODUCT_PUBLISH",
		SourceID:           staging.StagingID,
		OperatorID:         staging.OperatorID,
		Reason:             "product published",
		BasePublishVersion: publishResult.PublishVersion,
		Config:             config,
		InitialStock:       staging.Payload.StockConfig.InitialStock,
	})
	if inventoryErr == nil {
		inventoryReady = true
		inventoryMessage = "库存创建完成"
	} else {
		inventoryMessage = inventoryErr.Error()
	}

	now := time.Now()
	staging.Status = domain.StagingStatusPublished
	staging.UpdatedAt = now
	_ = s.repo.SaveStaging(ctx, staging)
	if review, err := s.repo.GetQCReviewByStaging(ctx, stagingID); err == nil && review.Status == domain.QCReviewApproved {
		review.Status = domain.QCReviewPublished
		review.UpdatedAt = now
		_ = s.repo.SaveQCReview(ctx, review)
	}
	if draft, err := s.repo.GetDraft(ctx, staging.DraftID); err == nil {
		draft.Status = domain.DraftStatusArchived
		draft.UpdatedAt = now
		_ = s.repo.SaveDraft(ctx, draft)
	}

	_ = s.appendLog(ctx, staging.OperationID, "PUBLISH", publishResult.SnapshotID, "PUBLISHED", staging.OperatorID, staging.SourceType, "商品中心发布版本、快照和 Outbox 已生成")
	return &dto.SupplyPublishResponse{
		StagingID:      staging.StagingID,
		ItemID:         publishResult.ItemID,
		SKUID:          publishResult.SKUID,
		PublishVersion: publishResult.PublishVersion,
		SnapshotID:     publishResult.SnapshotID,
		OutboxEventID:  publishResult.OutboxEventID,
		InventoryKey:   publishResult.InventoryKey,
		InventoryReady: inventoryReady,
		Message:        "发布完成；库存结果：" + inventoryMessage,
	}, nil
}

func (s *SupplyOpsService) ListLogs(ctx context.Context, traceID string) (*dto.SupplyLogListResponse, error) {
	logs, err := s.repo.ListLogs(ctx, traceID)
	if err != nil {
		return nil, err
	}
	resp := &dto.SupplyLogListResponse{Logs: make([]dto.SupplyOperationLogDTO, 0, len(logs))}
	for _, log := range logs {
		resp.Logs = append(resp.Logs, dto.SupplyOperationLogDTO{
			LogID:      log.LogID,
			TraceID:    log.TraceID,
			ObjectType: log.ObjectType,
			ObjectID:   log.ObjectID,
			Action:     log.Action,
			OperatorID: log.OperatorID,
			SourceType: string(log.SourceType),
			Message:    log.Message,
			CreatedAt:  log.CreatedAt.Unix(),
		})
	}
	return resp, nil
}

func (s *SupplyOpsService) appendLog(ctx context.Context, traceID string, objectType string, objectID string, action string, operatorID int64, sourceType domain.SupplySourceType, message string) error {
	return s.repo.AppendLog(ctx, domain.SupplyOperationLog{
		LogID:      s.repo.NextID(ctx, "log"),
		TraceID:    traceID,
		ObjectType: objectType,
		ObjectID:   objectID,
		Action:     action,
		OperatorID: operatorID,
		SourceType: sourceType,
		Message:    message,
		CreatedAt:  time.Now(),
	})
}

func decideQCPolicy(sourceType domain.SupplySourceType) domain.QCPolicy {
	switch sourceType {
	case domain.SupplySourceMerchant:
		return domain.QCPolicyRequired
	case domain.SupplySourceSupplier:
		return domain.QCPolicyAutoApprove
	default:
		return domain.QCPolicyAutoApprove
	}
}

func stagingCanPublish(staging *domain.ProductSupplyStaging) bool {
	switch staging.Status {
	case domain.StagingStatusApproved, domain.StagingStatusQCApproved, domain.StagingStatusPublishPending:
		return true
	default:
		return false
	}
}

func inventoryConfigFromPublishedPayload(payload domain.ProductPublishPayload, itemID int64) domain.InventoryConfig {
	return domain.InventoryConfig{
		InventoryKey:      payload.StockConfig.InventoryKey,
		ItemID:            itemID,
		SKUID:             payload.SKUID,
		Scope:             payload.StockConfig.Scope,
		ManagementType:    payload.StockConfig.ManagementType,
		UnitType:          payload.StockConfig.UnitType,
		DeductTiming:      payload.StockConfig.DeductTiming,
		SupplierID:        payload.StockConfig.SupplierID,
		SyncStrategy:      payload.StockConfig.SyncStrategy,
		OversellAllowed:   payload.StockConfig.OversellAllowed,
		LowStockThreshold: payload.StockConfig.LowStockThreshold,
	}
}

func publishPayloadDTOToDomain(payload dto.ProductPublishPayloadDTO) domain.ProductPublishPayload {
	return domain.ProductPublishPayload{
		ItemID:     payload.ItemID,
		SKUID:      payload.SKUID,
		SPUID:      payload.SPUID,
		SKUCode:    payload.SKUCode,
		Title:      payload.Title,
		CategoryID: payload.CategoryID,
		BasePrice:  domain.Money{Amount: payload.BasePrice.Amount, Currency: payload.BasePrice.Currency},
		Attributes: payload.Attributes,
		Images:     payload.Images,
		Resource: domain.ProductCenterResource{
			ResourceType: payload.Resource.ResourceType,
			ResourceID:   payload.Resource.ResourceID,
			Name:         payload.Resource.Name,
			Attributes:   payload.Resource.Attributes,
		},
		Offer: domain.ProductCenterOffer{
			OfferID:    payload.Offer.OfferID,
			OfferType:  domain.OfferType(payload.Offer.OfferType),
			RatePlanID: payload.Offer.RatePlanID,
			Price:      domain.Money{Amount: payload.Offer.Price.Amount, Currency: payload.Offer.Price.Currency},
			Channels:   payload.Offer.Channels,
			Attributes: payload.Offer.Attributes,
		},
		StockConfig: domain.ProductStockConfig{
			InventoryKey:      payload.StockConfig.InventoryKey,
			ManagementType:    domain.InventoryManagementType(payload.StockConfig.ManagementType),
			UnitType:          domain.InventoryUnitType(payload.StockConfig.UnitType),
			DeductTiming:      domain.InventoryDeductTiming(payload.StockConfig.DeductTiming),
			InitialStock:      payload.StockConfig.InitialStock,
			OversellAllowed:   payload.StockConfig.OversellAllowed,
			LowStockThreshold: payload.StockConfig.LowStockThreshold,
			SupplierID:        payload.StockConfig.SupplierID,
			SyncStrategy:      payload.StockConfig.SyncStrategy,
			Scope: domain.InventoryScope{
				ScopeType:    payload.StockConfig.Scope.ScopeType,
				ScopeID:      payload.StockConfig.Scope.ScopeID,
				CalendarDate: payload.StockConfig.Scope.CalendarDate,
				BatchID:      payload.StockConfig.Scope.BatchID,
				ChannelID:    payload.StockConfig.Scope.ChannelID,
				SupplierID:   payload.StockConfig.Scope.SupplierID,
			},
		},
		InputSchema: domain.InputSchema{
			SchemaID: payload.InputSchema.SchemaID,
			Fields:   toDomainInputFields(payload.InputSchema.Fields),
		},
		Fulfillment: domain.FulfillmentContract{
			Type:       domain.FulfillmentType(payload.Fulfillment.Type),
			Mode:       payload.Fulfillment.Mode,
			TimeoutSec: payload.Fulfillment.TimeoutSec,
			Attributes: payload.Fulfillment.Attributes,
		},
		RefundRule: domain.RefundRule{
			RuleID:      payload.RefundRule.RuleID,
			Refundable:  payload.RefundRule.Refundable,
			NeedReview:  payload.RefundRule.NeedReview,
			Description: payload.RefundRule.Description,
			Attributes:  payload.RefundRule.Attributes,
		},
	}
}
