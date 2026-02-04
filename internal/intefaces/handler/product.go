package handler

import (
	"context"
	"fmt"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/application/service"
	"github.com/zhanshen02154/product/pkg/swap"
	"github.com/zhanshen02154/product/proto/product"
	"go-micro.dev/v4/errors"
	"net/http"
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

// DeductInvetory 扣减订单库存
func (h *ProductHandler) DeductInvetory(ctx context.Context, in *product.OrderDetailReq, out *product.OrderProductResp) error {
	if in.OrderId == 0 || len(in.Products) == 0 {
		return errors.New(fmt.Sprintf("%d", in.OrderId), "data not found", http.StatusPreconditionFailed)
	}
	orderProductInvetoryDto := &dto.OrderProductInvetoryDto{}
	orderProductInvetoryDto.ConvertFromOrderDetailReq(in)
	err := h.ProductApplicationService.DeductInventory(ctx, orderProductInvetoryDto)
	if err != nil {
		return errors.New(fmt.Sprintf("%d", in.OrderId), fmt.Sprintf("deduct invetory error: %v", err), http.StatusPreconditionFailed)
	}
	out.StatusCode = "0000"
	return nil
}

// DeductInvetoryRevert 扣减订单库存事务补偿操作
func (h *ProductHandler) DeductInvetoryRevert(ctx context.Context, in *product.OrderDetailReq, out *product.OrderProductResp) error {
	if in.OrderId == 0 || len(in.Products) == 0 {
		return errors.New(fmt.Sprintf("%d", in.OrderId), "data not found", http.StatusPreconditionFailed)
	}
	orderProductInvetoryDto := &dto.OrderProductInvetoryDto{}
	orderProductInvetoryDto.ConvertFromOrderDetailReq(in)
	err := h.ProductApplicationService.DeductInvetoryRevert(ctx, orderProductInvetoryDto)
	if err != nil {
		return errors.New(fmt.Sprintf("%d", in.OrderId), fmt.Sprintf("deduct invetory revert error: %v", err), http.StatusPreconditionFailed)
	}
	out.StatusCode = "0000"
	return nil
}

// NewProductHandler 创建Handler
func NewProductHandler(appService service.IProductApplicationService) product.ProductHandler {
	return &ProductHandler{
		ProductApplicationService: appService,
	}
}
