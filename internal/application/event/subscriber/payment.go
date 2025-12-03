package subscriber

import (
	"context"
	"errors"
	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/application/service"
	"github.com/zhanshen02154/product/proto/product"
	"sync"
)

// PaymentEventHandler 支付事件处理器接口
type PaymentEventHandler interface {
	OnPaySuccess(ctx context.Context, req *product.OrderDetailReq) error
}

// PaymentEventHandlerImpl 支付事件处理器实现类
type PaymentEventHandlerImpl struct {
	productAppService service.IProductApplicationService
	orderInvetoryPool sync.Pool
}

// NewPaymentEventHandler 新建Handler
func NewPaymentEventHandler(appService service.IProductApplicationService) PaymentEventHandler {
	return &PaymentEventHandlerImpl{productAppService: appService, orderInvetoryPool: sync.Pool{
		New: func() interface{} {
			return &dto.OrderProductInvetoryDto{
				ProductInvetory:     []*dto.OrderProductInvetoryItem{},
				ProductSizeInvetory: []*dto.OrderProductSizeInvetoryItem{},
			}
		},
	}}
}

// OnPaySuccess 支付成功
func (h *PaymentEventHandlerImpl) OnPaySuccess(ctx context.Context, req *product.OrderDetailReq) error {
	if req.OrderId == 0 || len(req.Products) == 0 {
		return errors.New("order_id or products cannot be empty")
	}
	return nil
}
