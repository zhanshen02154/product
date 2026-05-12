package gorm

import (
	"context"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"gorm.io/gorm"
)

type SupplierRepositoryImpl struct {
	db *gorm.DB
}

// GetSupplierInfoBySkuID 获取指定SKU的供应商信息列表
func (r *SupplierRepositoryImpl) GetSupplierInfoBySkuID(ctx context.Context, skuID int64) ([]*repository.SupplierInfo, error) {
	db := GetDBFromContext(ctx, r.db)

	var supplierProducts []model.SupplierProduct
	err := db.Preload("Supplier").
		Where("sku_id = ?", skuID).
		Find(&supplierProducts).Error

	if err != nil {
		return nil, err
	}

	if len(supplierProducts) == 0 {
		return nil, nil
	}

	result := make([]*repository.SupplierInfo, 0, len(supplierProducts))
	for _, supplierProduct := range supplierProducts {
		if supplierProduct.Supplier == nil {
			continue
		}
		result = append(result, &repository.SupplierInfo{
			SkuID:            skuID,
			SupplierID:       int64(supplierProduct.SupplierID),
			SupplyPrice:      supplierProduct.SupplyPrice,
			MinOrderQuantity: supplierProduct.MinOrderQuantity,
			IsPreferred:      supplierProduct.IsPreferred,
			Supplier: &repository.SupplierDTO{
				ID:            int64(supplierProduct.SupplierID),
				Name:          supplierProduct.Supplier.Name,
				ContactPerson: supplierProduct.Supplier.ContactPerson,
				Phone:         supplierProduct.Supplier.Phone,
				Email:         supplierProduct.Supplier.Email,
				Address:       supplierProduct.Supplier.Address,
				Rating:        supplierProduct.Supplier.Rating,
				LeadTimeDays:  supplierProduct.Supplier.LeadTimeDays,
				PaymentTerms:  supplierProduct.Supplier.PaymentTerms,
			},
		})
	}

	return result, nil
}

// NewSupplierRepository 创建供应商仓储实例
func NewSupplierRepository(db *gorm.DB) repository.SupplierRepository {
	return &SupplierRepositoryImpl{db: db}
}
