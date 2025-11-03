package gorm

import (
	"context"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"gorm.io/gorm"
)

// ProductRepository
// @Description: 仓储层
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建productRepository
func NewProductRepository(db *gorm.DB) repository.IProductRepository {
	return &ProductRepository{db: db}
}

// FindProductByID 根据ID查找Product信息
func (u *ProductRepository) FindProductByID(ctx context.Context, id int64) (product *model.Product, err error) {
	db := GetDBFromContext(ctx, u.db)
	product = &model.Product{}
	return product, db.First(product, id).Error
}

// CreateProduct 创建Product信息
func (u *ProductRepository) CreateProduct(ctx context.Context, product *model.Product) (int64, error) {
	db := GetDBFromContext(ctx, u.db)
	return product.Id, db.Create(product).Error
}
