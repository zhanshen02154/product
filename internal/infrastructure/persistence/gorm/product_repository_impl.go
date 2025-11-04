package gorm

import (
	"context"
	"errors"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// FindProductSizeListByIds 查找产品规格
func (u *ProductRepository) FindProductSizeListByIds(ctx context.Context, ids []int64) ([]model.ProductSize, error) {
	db := GetDBFromContext(ctx, u.db)
	var productSizeList []model.ProductSize
	err := db.Debug().Model(model.ProductSize{}).Clauses(clause.Locking{Strength: "UPDATE"}).Select("id", "stock").Where("id IN ?", ids).Find(&productSizeList).Error
	return productSizeList, err
}

// FindProductListByIds 根据多个ID查找产品
func (u *ProductRepository) FindProductListByIds(ctx context.Context, productIds []int64) ([]model.Product, error) {
	db := GetDBFromContext(ctx, u.db)
	var list []model.Product
	err := db.Debug().Model(model.Product{}).Clauses(clause.Locking{Strength: "UPDATE"}).Select("id", "stock").Where("id in ?", productIds).Find(&list).Error
	return list, err
}

// DeductProductInvetory 扣减产品库存
func (u *ProductRepository) DeductProductInvetory(ctx context.Context, id int64, num int64) error {
	db := GetDBFromContext(ctx, u.db)
	tx := db.Model(model.Product{}).Where("id = ? AND stock > 0", id).Update("stock", gorm.Expr("stock - ?", num))
	if err := tx.Error; err != nil {
		return err
	}
	if tx.RowsAffected == 0 {
		return errors.New("failed to reduce stock")
	}
	return nil
}

// DeductProductSizeInvetory 扣减指定规格产品的库存
func (u *ProductRepository) DeductProductSizeInvetory(ctx context.Context, id int64, num int64) error {
	db := GetDBFromContext(ctx, u.db)
	tx := db.Model(model.ProductSize{}).Where("id = ? AND stock > 0", id).Update("stock", gorm.Expr("stock - ?", num))
	if err := tx.Error; err != nil {
		return err
	}
	if tx.RowsAffected == 0 {
		return errors.New("failed to reduce stock")
	}
	return nil
}
