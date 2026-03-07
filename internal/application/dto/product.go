package dto

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

// OrderSkuDto 订单SKU-库存的DTO
type OrderSkuDto struct {
	OrderID int64             `json:"order_id"`
	Sku     []OrderSkuItemDto `json:"sku"`
}

type OrderSkuItemDto struct {
	SkuID    int64  `json:"sku_id"`
	Quantity uint32 `json:"quantity"`
	Stock    uint32 `json:"stock"`
}
