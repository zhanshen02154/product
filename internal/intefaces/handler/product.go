package handler

import (
	"context"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/application/service"
	"github.com/zhanshen02154/product/pkg/swap"
	"github.com/zhanshen02154/product/proto/product"
)

type ProductHandler struct {
	ProductApplicationService service.IProductApplicationService
}

// AddProduct
//
//	@Description: 添加产品
//	@receiver h
//	@param ctx
//	@param req
//	@param resp
//	@return error
func (h *ProductHandler) AddProduct(ctx context.Context, req *product.ProductInfo, resp *product.ResponseProduct) error {
	productAdd := &dto.ProductDto{}
	if err := swap.SwapTo(req, productAdd); err != nil {
		return err
	}
	productRespDto, err := h.ProductApplicationService.AddProduct(ctx, productAdd)
	if err != nil {
		return err
	}
	resp.ProductId = productRespDto.Id
	return nil
}

func (h *ProductHandler) DeductInvetory(ctx context.Context, in *product.OrderDetailReq, out *product.OrderProductResp) error {
	err := h.ProductApplicationService.DeductInvetory(ctx, in)
	if err != nil {
		return err
	}
	out.StatusCode = "0000"
	return nil
}
