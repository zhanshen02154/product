package dto

type ProductDto struct {
	Id                 int64             `json:"id"`
	ProductName        string            `json:"product_name"`
	ProductSku         string            `json:"product_sku"`
	ProductPrice       float64           `json:"product_price"`
	ProductDescription string            `json:"product_description"`
	ProductNum         uint64            `json:"product_num"`
	ProductImage       []ProductImageDto `json:"product_image"`
}

type AddProductResponse struct {
	Id int64 `json:"id"`
}
