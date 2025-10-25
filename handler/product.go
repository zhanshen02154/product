package handler

import (
	"context"
	"git.imooc.com/zhanshen1614/product/common"
	"git.imooc.com/zhanshen1614/product/internal/domain/model"
	"git.imooc.com/zhanshen1614/product/internal/domain/service"
	"git.imooc.com/zhanshen1614/product/proto/product"
)

type Product struct {
	ProductDataService service.IProductDataService
}

// AddProduct
//	@Description: 添加产品
//	@receiver h
//	@param ctx
//	@param req
//	@param resp
//	@return error
func (h *Product) AddProduct(ctx context.Context, req *product.ProductInfo, resp *product.ResponseProduct) error {
	productAdd := &model.Product{}
	if err := common.SwapTo(req, productAdd); err != nil {
		return err
	}
	productId, err := h.ProductDataService.AddProduct(productAdd)
	if err != nil {
		return err
	}
	resp.ProductId = productId
	return nil
}
