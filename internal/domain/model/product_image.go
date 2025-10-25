package model

type ProductImage struct {
	Id             int64  `gorm:"type:bigint(20);primaryKey;not null;autoIncrement" json:"id"`
	ImageName      string `gorm:"type:varchar(60);not null;default:''" json:"image_name"`
	ImageCode      string `gorm:"type:varchar(60);not null;default:'';unique_index" json:"image_code"`
	ImageUrl       string `gorm:"type:varchar(150);not null;default:''" json:"image_url"`
	ImageProductId int64  `gorm:"type:bigint(20);not null;default:0" json:"image_product_id"`
}
