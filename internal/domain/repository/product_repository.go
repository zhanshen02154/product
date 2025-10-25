package repository

import (
	model2 "git.imooc.com/zhanshen1614/product/internal/domain/model"
)

type IProductRepository interface {
	FindProductByID(int64) (*model2.Product, error)
	CreateProduct(product *model2.Product) (int64, error)
	DeleteProductByID(int64) error
	UpdateProduct(product *model2.Product) error
	FindAll() ([]model2.Product, error)
}
