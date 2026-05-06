package service

import (
	"context"
	"fmt"

	"product-service/internal/application/dto"
	"product-service/internal/domain"
)

type RuntimeContextRepository interface {
	GetCategoryCapability(ctx context.Context, categoryID int64) (domain.CategoryCapability, error)
	GetProductRuntimeSource(ctx context.Context, skuID int64) (domain.ProductRuntimeSource, error)
}

type RuntimeContextService struct {
	repo       RuntimeContextRepository
	strategies map[int64]domain.CategoryStrategy
}

func NewRuntimeContextService(repo RuntimeContextRepository, strategies []domain.CategoryStrategy) *RuntimeContextService {
	registry := make(map[int64]domain.CategoryStrategy, len(strategies))
	for _, strategy := range strategies {
		registry[strategy.CategoryID()] = strategy
	}
	return &RuntimeContextService{
		repo:       repo,
		strategies: registry,
	}
}

func (s *RuntimeContextService) BuildRuntimeContext(ctx context.Context, skuID int64, categoryID int64, scene string) (*dto.RuntimeContextResponse, error) {
	product, err := s.repo.GetProductRuntimeSource(ctx, skuID)
	if err != nil {
		return nil, err
	}
	if product.CategoryID != categoryID {
		return nil, fmt.Errorf("sku %d belongs to category %d, not %d", skuID, product.CategoryID, categoryID)
	}

	capability, err := s.repo.GetCategoryCapability(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	strategy, ok := s.strategies[categoryID]
	if !ok {
		return nil, fmt.Errorf("category strategy not found: %d", categoryID)
	}

	input := domain.BuildContextInput{
		Product:    product,
		Capability: capability,
		Scene:      parseScene(scene),
	}

	runtimeContext, err := s.buildWithStrategy(ctx, strategy, input)
	if err != nil {
		return nil, err
	}
	if !runtimeContext.IsComplete() {
		return nil, fmt.Errorf("runtime context is incomplete for sku %d", skuID)
	}

	return toRuntimeContextDTO(runtimeContext), nil
}

func (s *RuntimeContextService) buildWithStrategy(ctx context.Context, strategy domain.CategoryStrategy, input domain.BuildContextInput) (*domain.ProductRuntimeContext, error) {
	productDefinition, err := strategy.BuildProductDefinition(ctx, input)
	if err != nil {
		return nil, err
	}
	resource, err := strategy.BuildResource(ctx, input)
	if err != nil {
		return nil, err
	}
	offer, err := strategy.BuildOffer(ctx, input)
	if err != nil {
		return nil, err
	}
	availability, err := strategy.CheckAvailability(ctx, input)
	if err != nil {
		return nil, err
	}
	inputSchema, err := strategy.BuildInputSchema(ctx, input)
	if err != nil {
		return nil, err
	}
	booking, err := strategy.BuildBooking(ctx, input)
	if err != nil {
		return nil, err
	}
	fulfillment, err := strategy.BuildFulfillment(ctx, input)
	if err != nil {
		return nil, err
	}
	refundRule, err := strategy.BuildRefundRule(ctx, input)
	if err != nil {
		return nil, err
	}

	return &domain.ProductRuntimeContext{
		SKUID:             input.Product.SKUID,
		CategoryID:        input.Capability.CategoryID,
		CategoryName:      input.Capability.CategoryName,
		Scene:             input.Scene,
		ProductDefinition: productDefinition,
		Resource:          resource,
		Offer:             offer,
		Availability:      availability,
		InputSchema:       inputSchema,
		Booking:           booking,
		Fulfillment:       fulfillment,
		RefundRule:        refundRule,
	}, nil
}

func parseScene(scene string) domain.Scene {
	switch domain.Scene(scene) {
	case domain.SceneList, domain.SceneDetail, domain.SceneCheckout, domain.SceneCreateOrder:
		return domain.Scene(scene)
	default:
		return domain.SceneDetail
	}
}

func toRuntimeContextDTO(ctx *domain.ProductRuntimeContext) *dto.RuntimeContextResponse {
	return &dto.RuntimeContextResponse{
		SKUID:             ctx.SKUID,
		CategoryID:        ctx.CategoryID,
		CategoryName:      ctx.CategoryName,
		Scene:             string(ctx.Scene),
		ProductDefinition: toProductDefinitionDTO(ctx.ProductDefinition),
		Resource:          toResourceDTO(ctx.Resource),
		Offer:             toOfferDTO(ctx.Offer),
		Availability:      toAvailabilityDTO(ctx.Availability),
		InputSchema:       toInputSchemaDTO(ctx.InputSchema),
		Booking:           toBookingDTO(ctx.Booking),
		Fulfillment:       toFulfillmentDTO(ctx.Fulfillment),
		RefundRule:        toRefundRuleDTO(ctx.RefundRule),
	}
}

func toProductDefinitionDTO(v *domain.ProductDefinition) dto.ProductDefinitionDTO {
	return dto.ProductDefinitionDTO{
		SKUID:      v.SKUID,
		SPUID:      v.SPUID,
		SKUCode:    v.SKUCode,
		Title:      v.Title,
		CategoryID: v.CategoryID,
		BasePrice:  toMoneyDTO(v.BasePrice),
		Attributes: v.Attributes,
	}
}

func toMoneyDTO(v domain.Money) dto.MoneyDTO {
	return dto.MoneyDTO{Amount: v.Amount, Currency: v.Currency}
}

func toResourceDTO(v *domain.ResourceContext) dto.ResourceContextDTO {
	return dto.ResourceContextDTO{
		ResourceType: string(v.ResourceType),
		ResourceID:   v.ResourceID,
		Name:         v.Name,
		Attributes:   v.Attributes,
	}
}

func toOfferDTO(v *domain.OfferContext) dto.OfferContextDTO {
	return dto.OfferContextDTO{
		OfferType:   string(v.OfferType),
		OfferID:     v.OfferID,
		RatePlanID:  v.RatePlanID,
		Price:       toMoneyDTO(v.Price),
		Attributes:  v.Attributes,
		NeedRefresh: v.NeedRefresh,
	}
}

func toAvailabilityDTO(v *domain.AvailabilityContext) dto.AvailabilityContextDTO {
	return dto.AvailabilityContextDTO{
		AvailabilityType: string(v.AvailabilityType),
		Sellable:         v.Sellable,
		Quantity:         v.Quantity,
		Message:          v.Message,
		Attributes:       v.Attributes,
	}
}

func toInputSchemaDTO(v *domain.InputSchema) dto.InputSchemaDTO {
	fields := make([]dto.InputFieldDTO, 0, len(v.Fields))
	for _, field := range v.Fields {
		fields = append(fields, dto.InputFieldDTO{
			Name:        field.Name,
			Label:       field.Label,
			Type:        field.Type,
			Required:    field.Required,
			Pattern:     field.Pattern,
			Description: field.Description,
		})
	}
	return dto.InputSchemaDTO{SchemaID: v.SchemaID, Fields: fields}
}

func toBookingDTO(v *domain.BookingRequirement) dto.BookingRequirementDTO {
	return dto.BookingRequirementDTO{
		Mode:         string(v.Mode),
		Required:     v.Required,
		TTLSeconds:   v.TTLSeconds,
		ProviderSide: v.ProviderSide,
		Attributes:   v.Attributes,
	}
}

func toFulfillmentDTO(v *domain.FulfillmentContract) dto.FulfillmentContractDTO {
	return dto.FulfillmentContractDTO{
		Type:       string(v.Type),
		Mode:       v.Mode,
		TimeoutSec: v.TimeoutSec,
		Attributes: v.Attributes,
	}
}

func toRefundRuleDTO(v *domain.RefundRule) dto.RefundRuleDTO {
	return dto.RefundRuleDTO{
		RuleID:      v.RuleID,
		Refundable:  v.Refundable,
		NeedReview:  v.NeedReview,
		Description: v.Description,
		Attributes:  v.Attributes,
	}
}
