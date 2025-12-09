package dto

import (
	"github.com/zhanshen02154/product/internal/domain/event/order"
	"github.com/zhanshen02154/product/proto/product"
)

type ProductDto struct {
	Id                 int64              `json:"id"`
	ProductName        string             `json:"product_name"`
	ProductSku         string             `json:"product_sku"`
	ProductPrice       float64            `json:"product_price"`
	ProductDescription string             `json:"product_description"`
	ProductNum         uint64             `json:"product_num"`
	ProductImage       []*ProductImageDto `json:"product_image"`
}

type OrderProductInvetoryItem struct {
	Id    int64 `json:"id"`
	Count int64 `json:"count"`
}

type OrderProductSizeInvetoryItem struct {
	Id    int64 `json:"id"`
	Count int64 `json:"count"`
}

type OrderProductInvetoryDto struct {
	OrderId             int64                           `json:"order_id"`
	ProductInvetory     []*OrderProductInvetoryItem     `json:"product_invetory"`
	ProductSizeInvetory []*OrderProductSizeInvetoryItem `json:"product_size_invetory"`
}

type AddProductResponse struct {
	Id int64 `json:"id"`
}

func (pdto *ProductDto) Reset() {
	pdto.Id = 0
	pdto.ProductName = ""
	pdto.ProductSku = ""
	pdto.ProductPrice = 0
	pdto.ProductDescription = "0"
	pdto.ProductNum = 0
	pdto.ProductImage = make([]*ProductImageDto, 0)
}

// Reset 重置DTO
func (productInvetoryDto *OrderProductInvetoryDto) Reset() {
	productInvetoryDto.OrderId = 0
	productInvetoryDto.ProductInvetory = make([]*OrderProductInvetoryItem, 0)
	productInvetoryDto.ProductSizeInvetory = make([]*OrderProductSizeInvetoryItem, 0)
}

// ConvertTo 将GRPC请求转换为DTO
func (productInvetoryDto *OrderProductInvetoryDto) ConvertTo(req *order.OnPaymentSuccess) {
	productInvetoryDto.OrderId = req.OrderId
	productCountMap := make(map[int64]int64)
	for _, item := range req.Products {
		if _, ok := productCountMap[item.ProductId]; ok {
			productCountMap[item.ProductId] += item.ProductNum
		} else {
			productCountMap[item.ProductId] = item.ProductNum
		}
		productInvetoryDto.ProductSizeInvetory = append(productInvetoryDto.ProductSizeInvetory, &OrderProductSizeInvetoryItem{
			Id:    item.ProductSizeId,
			Count: item.ProductNum,
		})
	}
	for key, val := range productCountMap {
		productInvetoryDto.ProductInvetory = append(productInvetoryDto.ProductInvetory, &OrderProductInvetoryItem{
			Id:    key,
			Count: val,
		})
	}
}

func (o *OrderProductInvetoryDto) ConvertFromOrderDetailReq(req *product.OrderDetailReq) {
	o.OrderId = req.OrderId
	productCountMap := make(map[int64]int64)
	for _, item := range req.Products {
		if _, ok := productCountMap[item.ProductId]; ok {
			productCountMap[item.ProductId] += item.ProductNum
		} else {
			productCountMap[item.ProductId] = item.ProductNum
		}
		o.ProductSizeInvetory = append(o.ProductSizeInvetory, &OrderProductSizeInvetoryItem{
			Id:    item.ProductSizeId,
			Count: item.ProductNum,
		})
	}
	for key, val := range productCountMap {
		o.ProductInvetory = append(o.ProductInvetory, &OrderProductInvetoryItem{
			Id:    key,
			Count: val,
		})
	}
}
