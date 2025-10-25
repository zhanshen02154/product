package gorm

import (
	model2 "git.imooc.com/zhanshen1614/product/internal/domain/model"
	"git.imooc.com/zhanshen1614/product/internal/domain/repository"
	"github.com/jinzhu/gorm"
)

// ProductRepository
// @Description: 仓储层
type ProductRepository struct {
	mysqlDb *gorm.DB
}

// FindProductByID 根据ID查找Product信息
func (u *ProductRepository) FindProductByID(productID int64) (product *model2.Product, err error) {
	product = &model2.Product{}
	return product, u.mysqlDb.Preload("ProductImage").Preload("ProductSize").Preload("ProductSeo").First(product, productID).Error
}

// CreateProduct 创建Product信息
func (u *ProductRepository) CreateProduct(product *model2.Product) (int64, error) {
	return product.Id, u.mysqlDb.Create(product).Error
}

// DeleteProductByID 根据ID删除Product信息
func (u *ProductRepository) DeleteProductByID(productID int64) error {
	tx := u.mysqlDb.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	if err := tx.Unscoped().Where("id = ?", productID).Delete(&model2.Product{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Where("images_product_id = ?", productID).Delete(&model2.ProductImage{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Where("size_product_id = ?", productID).Delete(&model2.ProductSize{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Where("seo_product_id = ?", productID).Delete(&model2.ProductSeo{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return u.mysqlDb.Commit().Error
}

// UpdateProduct 更新Product信息
func (u *ProductRepository) UpdateProduct(product *model2.Product) error {
	return u.mysqlDb.Model(product).Update(product).Error
}

// FindAll 获取结果集
func (u *ProductRepository) FindAll() (productAll []model2.Product, err error) {
	return productAll, u.mysqlDb.Preload("ProductImage").Preload("ProductSize").Preload("ProductSeo").Find(&productAll).Error
}

// NewProductRepository 创建productRepository
func NewProductRepository(db *gorm.DB) repository.IProductRepository {
	return &ProductRepository{mysqlDb: db}
}