package service

import (
	"context"
	"fmt"
	"regexp"

	"product-service/internal/application/dto"
)

type CategoryActionService struct {
	repo RuntimeContextRepository
}

func NewCategoryActionService(repo RuntimeContextRepository) *CategoryActionService {
	return &CategoryActionService{repo: repo}
}

func (s *CategoryActionService) ValidateTopupAccount(ctx context.Context, req dto.TopupValidateAccountRequest) (*dto.TopupValidateAccountResponse, error) {
	product, err := s.repo.GetProductRuntimeSource(ctx, req.SKUID)
	if err != nil {
		return nil, err
	}
	if product.CategoryID != 10102 {
		return nil, fmt.Errorf("sku %d is not a topup product", req.SKUID)
	}

	operator := product.Attributes["operator"]
	if operator == "" {
		operator = "Unknown Operator"
	}

	matched := regexp.MustCompile(`^\d{8,15}$`).MatchString(req.MobileNumber)
	if !matched {
		return &dto.TopupValidateAccountResponse{
			SKUID:        req.SKUID,
			MobileNumber: req.MobileNumber,
			Valid:        false,
			Operator:     operator,
			Message:      "手机号格式不符合充值要求",
		}, nil
	}

	return &dto.TopupValidateAccountResponse{
		SKUID:        req.SKUID,
		MobileNumber: req.MobileNumber,
		Valid:        true,
		Operator:     operator,
		Message:      "账号可充值；真实系统可在这里继续调用供应商账号校验接口",
	}, nil
}

func (s *CategoryActionService) SearchFlights(ctx context.Context, req dto.FlightSearchRequest) (*dto.FlightSearchResponse, error) {
	if req.From == "" || req.To == "" || req.Date == "" {
		return nil, fmt.Errorf("from, to and date are required")
	}
	if req.Adult <= 0 {
		req.Adult = 1
	}

	product, err := s.repo.GetProductRuntimeSource(ctx, 40001)
	if err != nil {
		return nil, err
	}
	if product.CategoryID != 40102 {
		return nil, fmt.Errorf("sample flight product is not configured")
	}

	capability, err := s.repo.GetCategoryCapability(ctx, product.CategoryID)
	if err != nil {
		return nil, err
	}

	routeCode := fmt.Sprintf("%s-%s", req.From, req.To)
	price := product.BasePrice
	if req.Adult > 1 {
		price.Amount = price.Amount * int64(req.Adult)
	}

	return &dto.FlightSearchResponse{
		RouteCode: routeCode,
		From:      req.From,
		To:        req.To,
		Date:      req.Date,
		Adult:     req.Adult,
		Offers: []dto.FlightOffer{
			{
				OfferToken:       fmt.Sprintf("offer_%s_%s_%d", routeCode, req.Date, req.Adult),
				SKUID:            product.SKUID,
				FlightNo:         "GA-1001",
				Carrier:          product.Attributes["carrier"],
				DepartureTime:    req.Date + "T09:30:00",
				ArrivalTime:      req.Date + "T11:50:00",
				Price:            dto.MoneyDTO{Amount: price.Amount, Currency: price.Currency},
				Seats:            9,
				OfferType:        string(capability.OfferType),
				AvailabilityType: string(capability.AvailabilityType),
				BookingMode:      string(capability.BookingMode),
			},
		},
	}, nil
}
