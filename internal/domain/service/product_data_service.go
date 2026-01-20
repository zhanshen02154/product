package service

import (
	"context"
	"errors"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/domain/event/product"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"github.com/zhanshen02154/product/pkg/metadata"
)

type IProductDataService interface {
	AddProduct(ctx context.Context, productInfo *model.Product) (int64, error)
	DeductInventory(ctx context.Context, req *dto.OrderProductInvetoryDto) (*product.OnInventoryDeductSuccess, error)
	DeductOrderInvetoryRevert(ctx context.Context, req *dto.OrderProductInvetoryDto) error
	FindEventExistsByOrderId(ctx context.Context, orderId int64) (bool, error)
}

// NewProductDataService 创建
func NewProductDataService(productRepository repository.IProductRepository, orderInventoryRepo repository.OrderInventoryEventRepository) IProductDataService {
	return &ProductDataService{productRepository: productRepository, orderInventoryRepo: orderInventoryRepo}
}

type ProductDataService struct {
	productRepository  repository.IProductRepository
	orderInventoryRepo repository.OrderInventoryEventRepository
}

// AddProduct 插入
func (u *ProductDataService) AddProduct(ctx context.Context, product *model.Product) (int64, error) {
	return u.productRepository.CreateProduct(ctx, product)
}

// DeductInventory 扣减库存
func (u *ProductDataService) DeductInventory(ctx context.Context, req *dto.OrderProductInvetoryDto) (*product.OnInventoryDeductSuccess, error) {
	var err error
	for _, item := range req.ProductInvetory {
		err = u.productRepository.DeductProductInventory(ctx, item.Id, item.Count)
		if err != nil {
			break
		}
	}
	for _, item := range req.ProductSizeInvetory {
		err = u.productRepository.DeductProductSizeInventory(ctx, item.Id, item.Count)
		if err != nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	eventId, ok := metadata.GetEventId(ctx)
	if ok {
		orderInventoryEvent := model.OrderInventoryEvent{
			EventId: eventId,
			OrderId: req.OrderId,
		}
		_, err = u.orderInventoryRepo.Create(ctx, &orderInventoryEvent)
	}

	// 发布扣减库存成功事件
	inventoryDeductSuccessEvent := product.OnInventoryDeductSuccess{
		OrderId:      req.OrderId,
		Products:     make([]*product.ProductInventoryItem, len(req.ProductInvetory)),
		ProductSizes: make([]*product.ProductSizeInventoryItem, len(req.ProductSizeInvetory)),
	}
	for _, item := range req.ProductInvetory {
		inventoryDeductSuccessEvent.Products = append(inventoryDeductSuccessEvent.Products, &product.ProductInventoryItem{
			Id:    item.Id,
			Count: item.Count,
		})
	}
	for _, item := range req.ProductSizeInvetory {
		inventoryDeductSuccessEvent.ProductSizes = append(inventoryDeductSuccessEvent.ProductSizes, &product.ProductSizeInventoryItem{
			Id:    item.Id,
			Count: item.Count,
		})
	}
	return &inventoryDeductSuccessEvent, err
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
