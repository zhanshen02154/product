package service

import (
	"git.imooc.com/zhanshen1614/product/internal/domain/model"
	"git.imooc.com/zhanshen1614/product/internal/domain/repository"
)

type IProductDataService interface {
	AddProduct(*model.Product) (int64, error)
}

// 创建
func NewProductDataService(productRepository repository.IProductRepository) IProductDataService {
	return &ProductDataService{productRepository}
}

type ProductDataService struct {
	productRepository repository.IProductRepository
}

// 插入
func (u *ProductDataService) AddProduct(product *model.Product) (int64, error) {
	return u.productRepository.CreateProduct(product)
}
