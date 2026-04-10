package gorm

import (
	"context"
	"errors"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"gorm.io/gorm"
)

type ProductSkuRepositoryImpl struct {
	db *gorm.DB
}

// BatchGetSkuByIDsWithFields 批量获取多个sku_id的SKU信息（可指定字段）
// 参数：skuIDs - SKU ID列表，fields - 需要查询的字段
// 返回：ProductSku结构体切片和error
func (s *ProductSkuRepositoryImpl) BatchGetSkuByIDsWithFields(ctx context.Context, skuIDs []int64) ([]model.ProductSku, error) {
	// 如果skuIDs为空，直接返回空切片
	if len(skuIDs) == 0 {
		return []model.ProductSku{}, nil
	}
	db := GetDBFromContext(ctx, s.db)
	var results []model.ProductSku

	// 执行查询
	err := db.Model(model.ProductSku{}).
		Select("id", "product_id", "stock", "stock_warn").
		Where("id IN ?", skuIDs).
		Where("status = ?", 1). // 只查询上架状态的SKU
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// DeductInventoryById 根据ID扣减库存并增加销量
func (s *ProductSkuRepositoryImpl) DeductInventoryById(ctx context.Context, id int64, count uint32) error {
	db := GetDBFromContext(ctx, s.db)
	tx := db.Model(model.ProductSku{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"stock": gorm.Expr("stock - ?", count),
			"sales": gorm.Expr("sales + ?", count),
		})
	if err := tx.Error; err != nil {
		return err
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetSkuDetailByID 根据SKU ID获取SKU详情，包括关联的商品信息和图片
func (s *ProductSkuRepositoryImpl) GetSkuDetailByID(ctx context.Context, skuID int64) (*model.ProductSku, error) {
	db := GetDBFromContext(ctx, s.db)
	var sku model.ProductSku

	// 预加载关联关系：Product（商品）、Images（SKU图片）
	// 注意：Product又关联了Category和Brand，根据需要决定是否预加载
	err := db.Model(&model.ProductSku{}).
		Where("id = ?", skuID).
		Preload("Product").
		Preload("Images").
		First(&sku).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 返回nil表示未找到，而不是错误
		}
		return nil, err
	}

	return &sku, nil
}

// BatchGetSkuInventoryInfo 批量获取SKU库存信息（用于库存阈值检查）
// 返回SKU的ID、编码、名称、库存和预警值
func (s *ProductSkuRepositoryImpl) BatchGetSkuInventoryInfo(ctx context.Context, skuIDs []int64) ([]model.ProductSku, error) {
	// 如果skuIDs为空，直接返回空切片
	if len(skuIDs) == 0 {
		return []model.ProductSku{}, nil
	}

	db := GetDBFromContext(ctx, s.db)
	var results []model.ProductSku

	// 只查询需要的字段：ID、编码、名称、库存、预警值
	err := db.Model(model.ProductSku{}).
		Select("id", "sku_no", "sku_name", "stock", "stock_warn").
		Where("id IN ?", skuIDs).
		Where("status = ?", 1). // 只查询上架状态的SKU
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetSkuStockBySkuNo 根据SKU编号获取SKU库存信息
func (s *ProductSkuRepositoryImpl) GetSkuStockBySkuNo(ctx context.Context, skuNo string) (*model.ProductSku, error) {
	db := GetDBFromContext(ctx, s.db)
	var result model.ProductSku

	err := db.Model(model.ProductSku{}).
		Select("id", "sku_no", "sku_name", "stock", "status", "stock_warn").
		Where("sku_no = ?", skuNo).
		First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}

// NewProductSkuRepository 创建商品SKU表仓储层
func NewProductSkuRepository(db *gorm.DB) repository.ProductSkuRepository {
	return &ProductSkuRepositoryImpl{db: db}
}
