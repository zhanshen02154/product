package handler

import (
	"context"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/application/service"
	"github.com/zhanshen02154/product/pkg/swap"
	"github.com/zhanshen02154/product/proto/product"
)

type ProductHandler struct {
	ProductApplicationService service.IProductApplicationService
}

// AddProduct
//
//	@Description: 添加产品
//	@receiver h
//	@param ctx
//	@param req
//	@param resp
//	@return error
func (h *ProductHandler) AddProduct(ctx context.Context, req *product.ProductInfo, resp *product.ResponseProduct) error {
	productAdd := &dto.ProductDto{}
	if err := swap.SwapTo(req, productAdd); err != nil {
		return err
	}
	productRespDto, err := h.ProductApplicationService.AddProduct(ctx, productAdd)
	if err != nil {
		return err
	}
	resp.ProductId = productRespDto.Id
	return nil
}

// GetProductSkuDetail
//
//	@Description: 获取商品SKU详情
//	@receiver h
//	@param ctx
//	@param req
//	@param resp
//	@return error
func (h *ProductHandler) GetProductSkuDetail(ctx context.Context, req *product.GetProductSkuDetailRequest, resp *product.GetProductSkuDetailResponse) error {
	// 直接调用应用层服务
	response, err := h.ProductApplicationService.GetProductSkuDetail(ctx, req.SkuId)
	if err != nil {
		return err
	}

	// 复制响应数据
	resp.Id = response.Id
	resp.ProductId = response.ProductId
	resp.SkuNo = response.SkuNo
	resp.SkuName = response.SkuName
	resp.SpecValueIds = response.SpecValueIds
	resp.SpecValueText = response.SpecValueText
	resp.Price = response.Price
	resp.MarketPrice = response.MarketPrice
	resp.Stock = response.Stock
	resp.StockWarn = response.StockWarn
	resp.Sales = response.Sales
	resp.MainImage = response.MainImage
	resp.Status = response.Status
	resp.CreatedAt = response.CreatedAt
	resp.UpdatedAt = response.UpdatedAt

	// 复制商品信息
	if response.Product != nil {
		resp.Product = &product.ProductBasicInfo{
			Id:          response.Product.Id,
			ProductNo:   response.Product.ProductNo,
			ProductName: response.Product.ProductName,
			MainImage:   response.Product.MainImage,
			Description: response.Product.Description,
		}
	}

	// 复制图片信息
	if len(response.Images) > 0 {
		resp.Images = make([]*product.SkuImageInfo, 0, len(response.Images))
		for _, img := range response.Images {
			resp.Images = append(resp.Images, &product.SkuImageInfo{
				Id:           img.Id,
				ImageUrl:     img.ImageUrl,
				IsMain:       img.IsMain,
				DisplayOrder: img.DisplayOrder,
			})
		}
	}

	return nil
}

// CheckSkuInventoryThreshold
//
//	@Description: 批量检查商品SKU库存是否小于阈值
//	@receiver h
//	@param ctx
//	@param req
//	@param resp
//	@return error
func (h *ProductHandler) CheckSkuInventoryThreshold(ctx context.Context, req *product.CheckSkuInventoryThresholdRequest, resp *product.CheckSkuInventoryThresholdResponse) error {
	// 调用应用层服务
	response, err := h.ProductApplicationService.CheckSkuInventoryThreshold(ctx, req.SkuIds, req.Threshold)
	if err != nil {
		return err
	}

	// 复制响应数据
	if len(response.Results) > 0 {
		resp.Results = make([]*product.SkuInventoryCheckResult, 0, len(response.Results))
		for _, result := range response.Results {
			resp.Results = append(resp.Results, &product.SkuInventoryCheckResult{
				SkuId:        result.SkuId,
				SkuNo:        result.SkuNo,
				SkuName:      result.SkuName,
				Stock:        result.Stock,
				Threshold:    result.Threshold,
				IsSufficient: result.IsSufficient,
			})
		}
	}

	return nil
}

// GetSkuStockBySkuNo
//
//	@Description: 根据SKU编号查询SKU库存信息
//	@receiver h
//	@param ctx
//	@param req
//	@param resp
//	@return error
func (h *ProductHandler) GetSkuStockBySkuNo(ctx context.Context, req *product.GetSkuStockBySkuNoRequest, resp *product.GetSkuStockBySkuNoResponse) error {
	response, err := h.ProductApplicationService.GetSkuStockBySkuNo(ctx, req.SkuId)
	if err != nil {
		return err
	}
	resp.SkuId = response.SkuId
	resp.Name = response.Name
	resp.Stock = response.Stock
	resp.Status = response.Status
	resp.StockWarn = response.StockWarn
	return nil
}

// NewProductHandler 创建Handler
func NewProductHandler(appService service.IProductApplicationService) product.ProductHandler {
	return &ProductHandler{
		ProductApplicationService: appService,
	}
}
