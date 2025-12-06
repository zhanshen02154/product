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
	db, ok := ctx.Value(txKey{}).(*gorm.DB)
	if !ok {
		db = u.db.WithContext(ctx)
	}
	var list []model.Product
	err := db.Debug().Model(model.Product{}).Select("id", "stock").Where("id in ?", productIds).Find(&list).Error
	return list, err
}

// DeductProductSizeInventory 扣减指定规格产品的库存
func (u *ProductRepository) DeductProductSizeInventory(ctx context.Context, id int64, num int64) error {
	db, ok := ctx.Value(txKey{}).(*gorm.DB)
	if !ok {
		db = u.db.WithContext(ctx)
	}
	tx := db.Debug().Model(model.ProductSize{}).Where("id = ? AND stock >= ?", id, num).Update("stock", gorm.Expr("stock - ?", num))
	if err := tx.Error; err != nil {
		return err
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeductProductInventory 扣减产品的库存
func (u *ProductRepository) DeductProductInventory(ctx context.Context, id int64, num int64) error {
	db, ok := ctx.Value(txKey{}).(*gorm.DB)
	if !ok {
		db = u.db.WithContext(ctx)
	}
	tx := db.Debug().Model(model.Product{}).Where("id = ? AND stock >= ?", id, num).Update("stock", gorm.Expr("stock - ?", num))
	if err := tx.Error; err != nil {
		return err
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeductProductSizeInvetoryRevert 扣减指定规格产品的库存补偿
func (u *ProductRepository) DeductProductSizeInvetoryRevert(ctx context.Context, id int64, num int64) error {
	db := GetDBFromContext(ctx, u.db)
	tx := db.Debug().Model(model.ProductSize{}).Where("id = ?", id).Update("stock", gorm.Expr("stock + ?", num))
	if err := tx.Error; err != nil {
		return err
	}
	if tx.RowsAffected == 0 {
		return errors.New("failed to reduce stock")
	}
	return nil
}

// DeductProductInventoryRevert 扣减产品的库存补偿
func (u *ProductRepository) DeductProductInventoryRevert(ctx context.Context, id int64, num int64) error {
	db := GetDBFromContext(ctx, u.db)
	tx := db.Debug().Model(model.Product{}).Where("id = ?", id).Update("stock", gorm.Expr("stock + ?", num))
	if err := tx.Error; err != nil {
		return err
	}
	if tx.RowsAffected == 0 {
		return errors.New("failed to reduce stock")
	}
	return nil
}
