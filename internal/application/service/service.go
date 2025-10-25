package service

import (
	"context"
	"git.imooc.com/zhanshen1614/product/common"
	"git.imooc.com/zhanshen1614/product/internal/application/dto"
	"git.imooc.com/zhanshen1614/product/internal/domain/model"
	"git.imooc.com/zhanshen1614/product/internal/domain/repository"
	"git.imooc.com/zhanshen1614/product/internal/domain/service"
)

type ProductApplicationService struct {
	productRepo repository.IProductRepository
	productDomainService service.IProductDataService
}

// 添加产品
func (srv *ProductApplicationService) AddProduct(ctx context.Context, productInfo *dto.ProductDto) (*dto.AddProductResponse, error) {
	productModel := &model.Product{}
	if err := common.SwapTo(productInfo, productModel); err != nil {
		return nil, err
	}
	productId, err := srv.productDomainService.AddProduct(productModel)
	if err != nil {
		return nil, err
	}
	return &dto.AddProductResponse{Id: productId}, nil
}