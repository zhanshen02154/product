package dto

type ProductImageDto struct {
	Id             int64  `json:"id"`
	ImageName      string `json:"image_name"`
	ImageCode      string `json:"image_code"`
	ImageUrl       string `json:"image_url"`
	ImageProductId int64  `json:"image_product_id"`
}
