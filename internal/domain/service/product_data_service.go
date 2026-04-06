package service

import (
	"context"
	"errors"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/domain/event/order"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"github.com/zhanshen02154/product/pkg/metadata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IProductDataService interface {
	AddProduct(ctx context.Context, productInfo *model.Product) (int64, error)
	DeductInventory(ctx context.Context, req *order.OnPaymentSuccess) (*dto.OrderSkuDto, error)
	DeductOrderInvetoryRevert(ctx context.Context, req *dto.OrderProductInvetoryDto) error
	FindEventExistsByOrderId(ctx context.Context, orderId int64) (bool, error)
	GetProductSkuDetail(ctx context.Context, skuID int64) (*model.ProductSku, error)
	BatchGetSkuInventoryInfo(ctx context.Context, skuIDs []int64) ([]model.ProductSku, error)
	GetSkuStockBySkuNo(ctx context.Context, skuNo string) (*model.ProductSku, error)
}

// NewProductDataService 创建
func NewProductDataService(productRepository repository.IProductRepository, orderInventoryRepo repository.OrderInventoryEventRepository, skuRepo repository.ProductSkuRepository, stockChangeRepo repository.InventoryStockChangeRecordRepository) IProductDataService {
	return &ProductDataService{productRepository: productRepository, orderInventoryRepo: orderInventoryRepo, skuRepo: skuRepo, stockChangeRepo: stockChangeRepo}
}

type ProductDataService struct {
	productRepository  repository.IProductRepository
	orderInventoryRepo repository.OrderInventoryEventRepository
	skuRepo            repository.ProductSkuRepository
	stockChangeRepo    repository.InventoryStockChangeRecordRepository
}

// AddProduct 插入
func (u *ProductDataService) AddProduct(ctx context.Context, product *model.Product) (int64, error) {
	return u.productRepository.CreateProduct(ctx, product)
}

// DeductInventory 扣减库存
func (u *ProductDataService) DeductInventory(ctx context.Context, req *order.OnPaymentSuccess) (*dto.OrderSkuDto, error) {
	var err error
	skuLenth := len(req.OrderDetails)
	skuIds := make([]int64, 0, skuLenth)
	skuQuantity := make(map[int64]uint32, skuLenth)
	for _, orderDetail := range req.OrderDetails {
		skuIds = append(skuIds, orderDetail.SkuId)
		skuQuantity[orderDetail.SkuId] = orderDetail.Quantity
	}
	skuList, err := u.skuRepo.BatchGetSkuByIDsWithFields(ctx, skuIds)
	if err != nil {
		return nil, status.Error(codes.NotFound, "sku query error:"+err.Error())
	}
	if len(skuList) == 0 || len(skuList) != skuLenth {
		return nil, status.Error(codes.NotFound, "sku not found")
	}
	for _, sku := range skuList {
		if val, ok := skuQuantity[sku.ID]; ok {
			if sku.Stock < val {
				return nil, status.Error(codes.FailedPrecondition, "sku stock out of order")
			}
		} else {
			return nil, status.Error(codes.FailedPrecondition, "sku stock out of order")
		}
	}

	// 执行扣减
	eventId, ok := metadata.GetEventId(ctx)
	orderSkuDto := &dto.OrderSkuDto{
		OrderID: req.OrderId,
		Sku:     make([]dto.OrderSkuItemDto, 0, skuLenth),
	}

	// 准备库存变更记录
	stockChangeRecords := make([]*model.InventoryStockChangeRecord, 0, skuLenth)

	for _, sku := range skuList {
		err = u.skuRepo.DeductInventoryById(ctx, sku.ID, skuQuantity[sku.ID])
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		afterStock := sku.Stock - skuQuantity[sku.ID]
		orderSkuDto.Sku = append(orderSkuDto.Sku, dto.OrderSkuItemDto{
			SkuID:     sku.ID,
			Quantity:  skuQuantity[sku.ID],
			Stock:     afterStock,
			Threshold: sku.StockWarn,
		})

		// 构建库存变更记录
		stockChangeRecords = append(stockChangeRecords, &model.InventoryStockChangeRecord{
			OrderID:     req.OrderId,
			SkuID:       sku.ID,
			SourceType:  model.SourceTypeOrderPayment,
			Quantity:    int64(skuQuantity[sku.ID]),
			BeforeStock: int64(sku.Stock),
			AfterStock:  int64(afterStock),
		})
	}

	// 批量写入库存变更记录
	if len(stockChangeRecords) > 0 {
		if err = u.stockChangeRepo.BatchCreate(ctx, stockChangeRecords); err != nil {
			return nil, status.Error(codes.Internal, "failed to create stock change records: "+err.Error())
		}
	}

	if ok {
		orderInventoryEvent := model.OrderInventoryEvent{
			EventId: eventId,
			OrderId: req.OrderId,
		}
		_, err = u.orderInventoryRepo.Create(ctx, &orderInventoryEvent)
	}

	return orderSkuDto, err
}

// DeductOrderInvetoryRevert 扣减订单库存补偿操作
func (u *ProductDataService) DeductOrderInvetoryRevert(ctx context.Context, req *dto.OrderProductInvetoryDto) error {
	var err error
	if len(req.ProductSizeInvetory) == 0 || len(req.ProductSizeInvetory) == 0 {
		err = errors.New("product data cannot be empty")
	}
	for _, item := range req.ProductInvetory {
		err = u.productRepository.DeductProductInventoryRevert(ctx, item.Id, item.Count)
		if err != nil {
			break
		}
	}
	if err != nil {
		return err
	}
	for _, item := range req.ProductSizeInvetory {
		err = u.productRepository.DeductProductSizeInvetoryRevert(ctx, item.Id, item.Count)
		if err != nil {
			break
		}
	}

	eventId, ok := metadata.GetEventId(ctx)
	if ok {
		orderInventoryEvent := model.OrderInventoryEvent{
			EventId: eventId,
			OrderId: req.OrderId,
		}
		_, err = u.orderInventoryRepo.Create(ctx, &orderInventoryEvent)
	}

	return err
}

// FindEventExistsByOrderId 查找订单库存已被处理过
func (u *ProductDataService) FindEventExistsByOrderId(ctx context.Context, orderId int64) (bool, error) {
	return u.orderInventoryRepo.FindEventExistsByOrderId(ctx, orderId)
}

// GetProductSkuDetail 获取SKU详情
func (u *ProductDataService) GetProductSkuDetail(ctx context.Context, skuID int64) (*model.ProductSku, error) {
	return u.skuRepo.GetSkuDetailByID(ctx, skuID)
}

// BatchGetSkuInventoryInfo 批量获取SKU库存信息
func (u *ProductDataService) BatchGetSkuInventoryInfo(ctx context.Context, skuIDs []int64) ([]model.ProductSku, error) {
	return u.skuRepo.BatchGetSkuInventoryInfo(ctx, skuIDs)
}

// GetSkuStockBySkuNo 根据SKU编号获取SKU库存信息
func (u *ProductDataService) GetSkuStockBySkuNo(ctx context.Context, skuNo string) (*model.ProductSku, error) {
	return u.skuRepo.GetSkuStockBySkuNo(ctx, skuNo)
}
