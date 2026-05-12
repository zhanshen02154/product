package service

import (
	"context"
	"strconv"

	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/domain/event/order"
	productEvent "github.com/zhanshen02154/product/internal/domain/event/product"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/service"
	"github.com/zhanshen02154/product/internal/infrastructure"
	"github.com/zhanshen02154/product/internal/infrastructure/event"
	"github.com/zhanshen02154/product/pkg/swap"
	productProto "github.com/zhanshen02154/product/proto/product"
	"go-micro.dev/v4/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IProductApplicationService interface {
	AddProduct(ctx context.Context, productInfo *dto.ProductDto) (*dto.AddProductResponse, error)
	DeductInventory(ctx context.Context, req *order.OnPaymentSuccess) error
	DeductInvetoryRevert(ctx context.Context, req *dto.OrderProductInvetoryDto) error
	GetProductSkuDetail(ctx context.Context, skuID int64) (*productProto.GetProductSkuDetailResponse, error)
	CheckSkuInventoryThreshold(ctx context.Context, skuIDs []int64, threshold uint32) (*productProto.CheckSkuInventoryThresholdResponse, error)
	GetSkuStockBySkuNo(ctx context.Context, skuNo string) (*productProto.GetSkuStockBySkuNoResponse, error)
	CreateRestockApply(ctx context.Context, req *dto.CreateRestockApplyDto) (*dto.CreateRestockApplyResponseDto, error)
	GetSkuSalesVolume(ctx context.Context, skuID int64, startTime, endTime string) (*productProto.GetSkuSalesVolumeResponse, error)
	GetSupplierInfo(ctx context.Context, skuID int64) (*productProto.GetSupplierInfoResponse, error)
	GetRestockApplyInfo(ctx context.Context, applicationNo string, userID int32) (*productProto.GetRestockApplyInfoResponse, error)
}

// ProductApplicationService 商品服务应用层
type ProductApplicationService struct {
	// 领域服务层
	productDomainService service.IProductDataService
	// 补货领域服务
	skuRestockService service.ISkuRestockService
	// 服务上下文
	serviceContext *infrastructure.ServiceContext
	// 事件总线
	eb event.Listener
}

func NewProductApplicationService(serviceContext *infrastructure.ServiceContext, eb event.Listener) IProductApplicationService {
	return &ProductApplicationService{
		productDomainService: service.NewProductDataService(
			serviceContext.NewProductRepository(),
			serviceContext.NewOrderInventoryEventRepo(),
			serviceContext.NewProductSkuRepository(),
			serviceContext.NewInventoryStockChangeRecordRepository(),
			serviceContext.NewSupplierRepository(),
		),
		skuRestockService: service.NewSkuRestockService(
			serviceContext.NewProductSkuRepository(),
			serviceContext.NewSkuRestockRepository(),
			serviceContext.NewSkuRestockAuditRepository(),
		),
		serviceContext: serviceContext,
		eb:             eb,
	}
}

// AddProduct 添加产品
func (appService *ProductApplicationService) AddProduct(ctx context.Context, productInfo *dto.ProductDto) (*dto.AddProductResponse, error) {
	productModel := &model.Product{}
	if err := swap.SwapTo(productInfo, productModel); err != nil {
		return nil, err
	}
	productId, err := appService.productDomainService.AddProduct(ctx, productModel)
	if err != nil {
		return nil, err
	}
	return &dto.AddProductResponse{Id: productId}, nil
}

// DeductInventory 扣减订单的库存
func (appService *ProductApplicationService) DeductInventory(ctx context.Context, req *order.OnPaymentSuccess) error {
	err := appService.serviceContext.RetryPolicy.Execute(ctx, func() error {
		eventExists, err := appService.productDomainService.FindEventExistsByOrderId(ctx, req.OrderId)
		if err != nil {
			return status.Error(codes.NotFound, "check order inventory event error: "+err.Error())
		}
		if eventExists {
			return nil
		}
		err = appService.serviceContext.TxManager.Execute(ctx, func(txCtx context.Context) error {
			skuDto, err := appService.productDomainService.DeductInventory(txCtx, req)
			if err != nil {
				return status.Error(codes.NotFound, "failed to deduct inventory error:"+err.Error())
			}

			inventoryEvent := productEvent.OnInventoryDeductSuccess{
				OrderId: skuDto.OrderID,
				Sku:     make([]*productEvent.SkuInfo, 0, len(skuDto.Sku)),
			}
			for _, item := range skuDto.Sku {
				inventoryEvent.Sku = append(inventoryEvent.Sku, &productEvent.SkuInfo{
					Id:        item.SkuID,
					Quantity:  item.Quantity,
					Stock:     item.Stock,
					Threshold: item.Threshold,
				})
			}
			err = appService.eb.Publish(txCtx, "ProductEvent", &inventoryEvent, strconv.FormatInt(req.OrderId, 10), "OnInventoryDeductSuccess")
			if err != nil {
				return status.Error(codes.Aborted, "failed to publish event error: "+err.Error())
			}
			return nil
		})

		return err
	})
	return err
}

// DeductInvetoryRevert 扣减订单的库存补偿
func (appService *ProductApplicationService) DeductInvetoryRevert(ctx context.Context, req *dto.OrderProductInvetoryDto) error {
	eventExists, err := appService.productDomainService.FindEventExistsByOrderId(ctx, req.OrderId)
	if err != nil {
		return status.Error(codes.NotFound, "check order inventory event error: "+err.Error())
	}
	if eventExists {
		return nil
	}
	lockKey := "deductinvetoryrevert-" + strconv.FormatInt(req.OrderId, 10)
	lock := appService.serviceContext.LockManager.NewLock(lockKey, 15)
	if err := lock.TryLock(ctx); err != nil {
		return err
	}

	defer func() {
		if err := lock.UnLock(ctx); err != nil {
			logger.Error("failed to unlock: ", lock.GetKey(), " reason: ", err)
		}
	}()
	return appService.serviceContext.TxManager.Execute(ctx, func(txCtx context.Context) error {
		return appService.productDomainService.DeductOrderInvetoryRevert(txCtx, req)
	})
}

// GetProductSkuDetail 获取商品SKU详情
func (appService *ProductApplicationService) GetProductSkuDetail(ctx context.Context, skuID int64) (*productProto.GetProductSkuDetailResponse, error) {
	// 调用领域服务获取SKU详情
	sku, err := appService.productDomainService.GetProductSkuDetail(ctx, skuID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get sku detail: "+err.Error())
	}
	if sku == nil {
		return nil, status.Error(codes.NotFound, "sku not found")
	}

	// 转换模型为protobuf响应
	response := &productProto.GetProductSkuDetailResponse{
		Id:            sku.ID,
		ProductId:     int64(sku.ProductID),
		SkuNo:         sku.SkuNo,
		SkuName:       sku.SkuName,
		SpecValueIds:  sku.SpecValueIDs,
		SpecValueText: sku.SpecValueText,
		Price:         sku.Price,
		Stock:         sku.Stock,
		StockWarn:     sku.StockWarn,
		Sales:         int32(sku.Sales),
		Status:        int32(sku.Status),
		CreatedAt:     sku.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     sku.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 处理市场价（可能为nil）
	if sku.MarketPrice != nil {
		response.MarketPrice = *sku.MarketPrice
	}

	// 处理主图（可能为nil）
	if sku.MainImage != nil {
		response.MainImage = *sku.MainImage
	}

	// 处理关联的商品信息
	if sku.Product != nil {
		response.Product = &productProto.ProductBasicInfo{
			Id:          sku.Product.ID,
			ProductNo:   sku.Product.ProductNo,
			ProductName: sku.Product.ProductName,
		}
		if sku.Product.MainImage != nil {
			response.Product.MainImage = *sku.Product.MainImage
		}
		if sku.Product.Description != nil {
			response.Product.Description = *sku.Product.Description
		}
	}

	// 处理SKU图片
	if len(sku.Images) > 0 {
		response.Images = make([]*productProto.SkuImageInfo, 0, len(sku.Images))
		for _, img := range sku.Images {
			imageInfo := &productProto.SkuImageInfo{
				Id:           img.ID,
				ImageUrl:     img.ImageURL,
				IsMain:       img.IsMain,
				DisplayOrder: int32(img.DisplayOrder),
			}
			response.Images = append(response.Images, imageInfo)
		}
	}

	return response, nil
}

// CheckSkuInventoryThreshold 批量检查SKU库存是否小于阈值
func (appService *ProductApplicationService) CheckSkuInventoryThreshold(ctx context.Context, skuIDs []int64, threshold uint32) (*productProto.CheckSkuInventoryThresholdResponse, error) {
	// 调用领域服务获取SKU库存信息
	skuList, err := appService.productDomainService.BatchGetSkuInventoryInfo(ctx, skuIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get sku inventory info: "+err.Error())
	}

	// 构建响应
	response := &productProto.CheckSkuInventoryThresholdResponse{
		Results: make([]*productProto.SkuInventoryCheckResult, 0, len(skuList)),
	}

	// 遍历SKU列表，检查库存阈值
	for _, sku := range skuList {
		// 如果没有指定阈值，使用SKU自己的stock_warn值
		checkThreshold := threshold
		if checkThreshold == 0 {
			checkThreshold = sku.StockWarn
		}

		// 判断库存是否充足
		isSufficient := sku.Stock >= checkThreshold

		result := &productProto.SkuInventoryCheckResult{
			SkuId:        sku.ID,
			SkuNo:        sku.SkuNo,
			SkuName:      sku.SkuName,
			Stock:        sku.Stock,
			Threshold:    checkThreshold,
			IsSufficient: isSufficient,
		}
		response.Results = append(response.Results, result)
	}

	return response, nil
}

// GetSkuStockBySkuNo 根据SKU编号查询SKU库存信息
func (appService *ProductApplicationService) GetSkuStockBySkuNo(ctx context.Context, skuNo string) (*productProto.GetSkuStockBySkuNoResponse, error) {
	if skuNo == "" {
		return nil, status.Error(codes.InvalidArgument, "sku_id cannot be empty")
	}

	sku, err := appService.productDomainService.GetSkuStockBySkuNo(ctx, skuNo)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get sku stock: "+err.Error())
	}
	if sku == nil {
		return nil, status.Error(codes.NotFound, "sku not found")
	}

	return &productProto.GetSkuStockBySkuNoResponse{
		Id:        uint64(sku.ID),
		SkuId:     sku.SkuNo,
		Name:      sku.SkuName,
		Stock:     sku.Stock,
		Status:    int32(sku.Status),
		StockWarn: sku.StockWarn,
	}, nil
}

// CreateRestockApply 提交补货申请
func (appService *ProductApplicationService) CreateRestockApply(ctx context.Context, req *dto.CreateRestockApplyDto) (*dto.CreateRestockApplyResponseDto, error) {
	var response *dto.CreateRestockApplyResponseDto

	err := appService.serviceContext.TxManager.Execute(ctx, func(txCtx context.Context) error {
		var txErr error
		response, txErr = appService.skuRestockService.CreateRestockApply(txCtx, req)
		return txErr
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create restock apply: "+err.Error())
	}

	return response, nil
}

// GetSkuSalesVolume 获取SKU在指定时间范围内的销量和日均销量
func (appService *ProductApplicationService) GetSkuSalesVolume(ctx context.Context, skuID int64, startTime, endTime string) (*productProto.GetSkuSalesVolumeResponse, error) {
	if skuID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sku_id must be greater than 0")
	}

	salesVolume, dailyAvgSales, err := appService.productDomainService.GetSkuSalesVolume(ctx, skuID, startTime, endTime)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get sku sales volume: "+err.Error())
	}

	return &productProto.GetSkuSalesVolumeResponse{
		SkuId:         skuID,
		SalesVolume:   salesVolume,
		DailyAvgSales: dailyAvgSales,
	}, nil
}

// GetSupplierInfo 获取指定SKU的供应商信息列表
func (appService *ProductApplicationService) GetSupplierInfo(ctx context.Context, skuID int64) (*productProto.GetSupplierInfoResponse, error) {
	if skuID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "sku_id must be greater than 0")
	}

	supplierInfoList, err := appService.productDomainService.GetSupplierInfo(ctx, skuID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get supplier info: "+err.Error())
	}

	if supplierInfoList == nil || len(supplierInfoList) == 0 {
		return nil, nil
	}

	suppliers := make([]*productProto.SupplierInfoItem, 0, len(supplierInfoList))
	for _, supplierInfo := range supplierInfoList {
		supplierDTO := &productProto.SupplierDTO{
			Id:            supplierInfo.Supplier.ID,
			Name:          supplierInfo.Supplier.Name,
			ContactPerson: supplierInfo.Supplier.ContactPerson,
			Phone:         supplierInfo.Supplier.Phone,
			Email:         supplierInfo.Supplier.Email,
			Address:       supplierInfo.Supplier.Address,
			Rating:        supplierInfo.Supplier.Rating,
			LeadTimeDays:  int32(supplierInfo.Supplier.LeadTimeDays),
			PaymentTerms:  supplierInfo.Supplier.PaymentTerms,
		}
		suppliers = append(suppliers, &productProto.SupplierInfoItem{
			SkuId:            supplierInfo.SkuID,
			SupplierId:       supplierInfo.SupplierID,
			SupplyPrice:      supplierInfo.SupplyPrice,
			MinOrderQuantity: int32(supplierInfo.MinOrderQuantity),
			IsPreferred:      supplierInfo.IsPreferred,
			Supplier:         supplierDTO,
		})
	}

	return &productProto.GetSupplierInfoResponse{
		Suppliers: suppliers,
	}, nil
}

// GetRestockApplyInfo 获取补货申请信息
func (appService *ProductApplicationService) GetRestockApplyInfo(ctx context.Context, applicationNo string, userID int32) (*productProto.GetRestockApplyInfoResponse, error) {
	if applicationNo == "" {
		return nil, status.Error(codes.InvalidArgument, "application_no cannot be empty")
	}

	record, err := appService.skuRestockService.GetRestockApplyInfo(ctx, applicationNo, int(userID))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get restock apply info: "+err.Error())
	}
	if record == nil {
		return nil, status.Error(codes.NotFound, "restock apply not found")
	}

	// 构建审核信息
	var audit *productProto.RestockAuditInfo
	if len(record.Audits) > 0 {
		latestAudit := record.Audits[0]
		audit = &productProto.RestockAuditInfo{
			Id:                latestAudit.ID,
			RestockId:         int64(latestAudit.RestockID),
			AuditUserId:       uint32(latestAudit.AuditUserID),
			AuditStatus:       uint32(latestAudit.AuditStatus),
			AuditFailedReason: latestAudit.AuditFailedReason,
			CreatedAt:         latestAudit.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:         latestAudit.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return &productProto.GetRestockApplyInfoResponse{
		Id:            record.ID,
		SkuId:         int64(record.SkuID),
		Quantity:      record.Quantity,
		Reason:        record.Reason,
		Status:        uint32(record.Status),
		ApplicationNo: record.ApplicationNo,
		Audit:         audit,
	}, nil
}
