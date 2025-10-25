package model

type ProductSize struct {
	Id            int64  `gorm:"type:bigint(20);primaryKey;not null;autoIncrement" json:"id"`
	SizeName      string `gorm:"type:varchar(50);not null;default:''" json:"product_name"`
	SizeCode      string `gorm:"type:varchar(50);not null;default:'';unique_index" json:"size_code"`
	SizeProductId int64  `gorm:"type:bigint(20);not null;default:0" json:"size_product_id"`
}
