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
	resp.Id = response.Id
	resp.SkuId = response.SkuId
	resp.Name = response.Name
	resp.Stock = response.Stock
	resp.Status = response.Status
	resp.StockWarn = response.StockWarn
	return nil
}

// CreateRestockApply
//
//	@Description: 提交补货申请
//	@receiver h
//	@param ctx
//	@param req
//	@param resp
//	@return error
func (h *ProductHandler) CreateRestockApply(ctx context.Context, req *product.CreateRestockApplyRequest, resp *product.CreateRestockApplyResponse) error {
	// 构建DTO
	applyDto := &dto.CreateRestockApplyDto{
		SkuID:    req.SkuId,
		UserID:   req.UserId,
		Quantity: req.Quantity,
		Reason:   req.Reason,
	}

	// 调用应用服务
	response, err := h.ProductApplicationService.CreateRestockApply(ctx, applyDto)
	if err != nil {
		return err
	}

	// 构建响应
	if response.RestockRecord != nil {
		resp.RestockRecord = &product.RestockRecordInfo{
			Id:           response.RestockRecord.ID,
			UserId:       response.RestockRecord.UserID,
			SkuId:        response.RestockRecord.SkuID,
			Quantity:     response.RestockRecord.Quantity,
			Reason:       response.RestockRecord.Reason,
			Status:       uint32(response.RestockRecord.Status),
			FailedReason: response.RestockRecord.FailedReason,
			CreatedAt:    response.RestockRecord.CreatedAt,
		}
	}

	if response.SkuInfo != nil {
		resp.SkuInfo = &product.SkuBasicInfo{
			Id:            response.SkuInfo.ID,
			SkuNo:         response.SkuInfo.SkuNo,
			SkuName:       response.SkuInfo.SkuName,
			SpecValueText: response.SkuInfo.SpecValueText,
		}
	}

	return nil
}

// GetRestockApplyInfo
//
//	@Description: 获取补货申请信息
//	@receiver h
//	@param ctx
//	@param req
//	@param resp
//	@return error
func (h *ProductHandler) GetRestockApplyInfo(ctx context.Context, req *product.GetRestockApplyInfoRequest, resp *product.GetRestockApplyInfoResponse) error {
	response, err := h.ProductApplicationService.GetRestockApplyInfo(ctx, req.ApplicationNo, req.UserId)
	if err != nil {
		return err
	}

	resp.Id = response.Id
	resp.SkuId = response.SkuId
	resp.Quantity = response.Quantity
	resp.Reason = response.Reason
	resp.Status = response.Status
	resp.ApplicationNo = response.ApplicationNo

	if response.Audit != nil {
		resp.Audit = &product.RestockAuditInfo{
			Id:                response.Audit.Id,
			RestockId:         response.Audit.RestockId,
			AuditUserId:       response.Audit.AuditUserId,
			AuditStatus:       response.Audit.AuditStatus,
			AuditFailedReason: response.Audit.AuditFailedReason,
			CreatedAt:         response.Audit.CreatedAt,
			UpdatedAt:         response.Audit.UpdatedAt,
		}
	}

	return nil
}

// GetSkuSalesVolume
//
//	@Description: 获取SKU在指定时间范围内的销量和日均销量
//	@receiver h
//	@param ctx
//	@param req
//	@param resp
//	@return error
func (h *ProductHandler) GetSkuSalesVolume(ctx context.Context, req *product.GetSkuSalesVolumeRequest, resp *product.GetSkuSalesVolumeResponse) error {
	// 调用应用层服务
	response, err := h.ProductApplicationService.GetSkuSalesVolume(ctx, req.SkuId, req.StartTime, req.EndTime)
	if err != nil {
		return err
	}

	// 复制响应数据
	resp.SkuId = response.SkuId
	resp.SalesVolume = response.SalesVolume
	resp.DailyAvgSales = response.DailyAvgSales

	return nil
}

// GetSupplierInfo
//
//	@Description: 获取指定SKU的供应商信息列表
//	@receiver h
//	@param ctx
//	@param req
//	@param resp
//	@return error
func (h *ProductHandler) GetSupplierInfo(ctx context.Context, req *product.GetSupplierInfoRequest, resp *product.GetSupplierInfoResponse) error {
	// 调用应用层服务
	response, err := h.ProductApplicationService.GetSupplierInfo(ctx, req.SkuId)
	if err != nil {
		return err
	}

	// 如果没有找到供应商信息，返回空列表
	if response == nil {
		return nil
	}

	// 复制响应数据
	resp.Suppliers = response.Suppliers

	return nil
}

// NewProductHandler 创建Handler
func NewProductHandler(appService service.IProductApplicationService) product.ProductHandler {
	return &ProductHandler{
		ProductApplicationService: appService,
	}
}
