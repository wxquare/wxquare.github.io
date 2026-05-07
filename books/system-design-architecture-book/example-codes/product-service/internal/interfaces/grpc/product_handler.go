package grpc

import (
	"context"
	"fmt"

	"product-service/internal/application/dto"
	"product-service/internal/application/service"
)

// ProductServiceServer gRPC服务端实现
// 注意：这是示例代码，不要求编译通过
// 实际使用需要：protoc --go_out=. --go-grpc_out=. product.proto
type ProductServiceServer struct {
	// UnimplementedProductServiceServer  // protobuf生成的基类
	productService *service.ProductService
}

func NewProductServiceServer(productService *service.ProductService) *ProductServiceServer {
	return &ProductServiceServer{
		productService: productService,
	}
}

// GetProduct 查询单个商品（gRPC接口）
// 数据流转：gRPC Request → Application Service → Domain → Infrastructure → Domain → Application → gRPC Response
func (s *ProductServiceServer) GetProduct(ctx context.Context, req *GetProductRequest) (*GetProductResponse, error) {
	fmt.Printf("\n🌐 [Interface Layer - gRPC] GetProduct called, SKUID=%d\n", req.SkuId)

	// 转换gRPC请求 → DTO
	dtoReq := &dto.GetProductRequest{
		SKUID: req.SkuId,
	}

	// 调用应用服务
	dtoResp, err := s.productService.GetProduct(ctx, dtoReq)
	if err != nil {
		fmt.Printf("❌ [Interface Layer - gRPC] Error: %v\n", err)
		return nil, err
	}

	// 转换DTO → gRPC响应
	grpcResp := &GetProductResponse{
		Product: &ProductInfo{
			SkuId:     dtoResp.SKUID,
			SpuId:     dtoResp.SPUID,
			SkuCode:   dtoResp.SKUCode,
			SkuName:   dtoResp.SKUName,
			BasePrice: dtoResp.BasePrice,
			Specs:     dtoResp.Specs,
			Status:    dtoResp.Status,
			Images:    dtoResp.Images,
			CreatedAt: dtoResp.CreatedAt,
			UpdatedAt: dtoResp.UpdatedAt,
		},
	}

	fmt.Printf("✅ [Interface Layer - gRPC] Response sent\n")
	return grpcResp, nil
}

// BatchGetProducts 批量查询商品（gRPC接口）
func (s *ProductServiceServer) BatchGetProducts(ctx context.Context, req *BatchGetProductsRequest) (*BatchGetProductsResponse, error) {
	fmt.Printf("\n🌐 [Interface Layer - gRPC] BatchGetProducts called, count=%d\n", len(req.SkuIds))

	// 这里可以类似实现批量查询逻辑
	// 为简化示例，这里不实现具体逻辑

	return &BatchGetProductsResponse{
		Products: []*ProductInfo{},
	}, nil
}

// OnShelf 商品上架（gRPC接口）
func (s *ProductServiceServer) OnShelf(ctx context.Context, req *OnShelfRequest) (*OnShelfResponse, error) {
	fmt.Printf("\n🌐 [Interface Layer - gRPC] OnShelf called, SKUID=%d\n", req.SkuId)

	// 这里可以调用应用服务的OnShelf方法
	// 为简化示例，这里不实现具体逻辑

	return &OnShelfResponse{
		Success: true,
		Message: "上架成功",
	}, nil
}

// UpdateBasePrice 更新基础价格（gRPC接口）
func (s *ProductServiceServer) UpdateBasePrice(ctx context.Context, req *UpdateBasePriceRequest) (*UpdateBasePriceResponse, error) {
	fmt.Printf("\n🌐 [Interface Layer - gRPC] UpdateBasePrice called, SKUID=%d, NewPrice=%d\n", req.SkuId, req.NewPrice)

	// 这里可以调用应用服务的UpdateBasePrice方法
	// 为简化示例，这里不实现具体逻辑

	return &UpdateBasePriceResponse{
		Success: true,
		Message: "价格更新成功",
	}, nil
}

// 以下是protobuf生成的消息类型定义（示例，实际由protoc生成）
// 这里手动定义仅作为演示，展示gRPC接口的使用方式

type GetProductRequest struct {
	SkuId int64 `protobuf:"varint,1,opt,name=sku_id,json=skuId,proto3" json:"sku_id,omitempty"`
}

type GetProductResponse struct {
	Product *ProductInfo `protobuf:"bytes,1,opt,name=product,proto3" json:"product,omitempty"`
}

type BatchGetProductsRequest struct {
	SkuIds []int64 `protobuf:"varint,1,rep,packed,name=sku_ids,json=skuIds,proto3" json:"sku_ids,omitempty"`
}

type BatchGetProductsResponse struct {
	Products []*ProductInfo `protobuf:"bytes,1,rep,name=products,proto3" json:"products,omitempty"`
}

type OnShelfRequest struct {
	SkuId int64 `protobuf:"varint,1,opt,name=sku_id,json=skuId,proto3" json:"sku_id,omitempty"`
}

type OnShelfResponse struct {
	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

type UpdateBasePriceRequest struct {
	SkuId    int64 `protobuf:"varint,1,opt,name=sku_id,json=skuId,proto3" json:"sku_id,omitempty"`
	NewPrice int64 `protobuf:"varint,2,opt,name=new_price,json=newPrice,proto3" json:"new_price,omitempty"`
}

type UpdateBasePriceResponse struct {
	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

type ProductInfo struct {
	SkuId     int64             `protobuf:"varint,1,opt,name=sku_id,json=skuId,proto3" json:"sku_id,omitempty"`
	SpuId     int64             `protobuf:"varint,2,opt,name=spu_id,json=spuId,proto3" json:"spu_id,omitempty"`
	SkuCode   string            `protobuf:"bytes,3,opt,name=sku_code,json=skuCode,proto3" json:"sku_code,omitempty"`
	SkuName   string            `protobuf:"bytes,4,opt,name=sku_name,json=skuName,proto3" json:"sku_name,omitempty"`
	BasePrice int64             `protobuf:"varint,5,opt,name=base_price,json=basePrice,proto3" json:"base_price,omitempty"`
	Specs     map[string]string `protobuf:"bytes,6,rep,name=specs,proto3" json:"specs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Status    string            `protobuf:"bytes,7,opt,name=status,proto3" json:"status,omitempty"`
	Images    []string          `protobuf:"bytes,8,rep,name=images,proto3" json:"images,omitempty"`
	CreatedAt int64             `protobuf:"varint,9,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt int64             `protobuf:"varint,10,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}
