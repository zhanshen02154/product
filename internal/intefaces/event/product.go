package event

import (
	"github.com/zhanshen02154/product/internal/application/service"
)

// ProductEventHandler 订单事件处理器接口
type ProductEventHandler interface {
}

// ProductEventHandlerImpl 订单事件处理器实现类
type ProductEventHandlerImpl struct {
	orderAppService service.IProductApplicationService
}

// NewHandler 新建Handler
func NewHandler(appService service.IProductApplicationService) ProductEventHandler {
	return &ProductEventHandlerImpl{orderAppService: appService}
}
