package model

// Product
// @Description: 产品模型
type Product struct {
	Id                 int64   `gorm:"type:bigint(20);primaryKey;not null;autoIncrement" json:"id"`
	ProductName        string  `gorm:"type:varchar(50);not null;default:''" json:"product_name"`
	ProductSku         string  `gorm:"type:varchar(50);not null;default:'';unique_index" json:"product_sku"`
	ProductPrice       float64 `gorm:"type:decimal(18,2);not null;default:0" json:"product_price"`
	ProductDescription string  `gorm:"type:varchar(100);not null;default:0" json:"product_description"`
	ProductNum         uint64  `gorm:"type:int(11);not null;default:0" json:"product_num"`
}
