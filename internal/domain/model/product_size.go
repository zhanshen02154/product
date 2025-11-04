package model

import "time"

type ProductSize struct {
	Id            int64     `gorm:"type:bigint(20);primaryKey;not null;autoIncrement" json:"id"`
	SizeName      string    `gorm:"type:varchar(50);not null;default:''" json:"product_name"`
	SizeCode      string    `gorm:"type:varchar(50);not null;default:'';unique_index" json:"size_code"`
	SizeProductId int64     `gorm:"type:bigint(20);not null;default:0" json:"size_product_id"`
	Stock         int64     `gorm:"type:int(11);not null;default:0" json:"stock"`
	Price         float64   `gorm:"type:decimal(18,2);not null;default:0" json:"price"`
	CreatedAt     time.Time `gorm:"type:datetime;comment:'创建时间'" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:datetime;comment:'更新时间'" json:"updated_at"`
}
