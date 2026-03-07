package model

import (
	"gorm.io/gorm"
	"time"
)

// ProductSku 对应SKU表 (product_skus)
type ProductSku struct {
	ID            int64          `gorm:"primaryKey;autoIncrement;comment:SKU ID"`
	ProductID     uint           `gorm:"not null;index:idx_product_id;comment:商品ID"`
	SkuNo         string         `gorm:"type:varchar(64);not null;uniqueIndex:uk_sku_no;comment:SKU编号"`
	SkuName       string         `gorm:"type:varchar(255);not null;comment:SKU名称"`
	SpecValueIDs  string         `gorm:"type:varchar(500);not null;comment:规格值ID组合"`
	SpecValueText string         `gorm:"type:varchar(500);not null;comment:规格值文本"`
	Price         float64        `gorm:"type:decimal(18,2);not null;comment:价格"`
	MarketPrice   *float64       `gorm:"type:decimal(18,2);comment:市场价"`
	Stock         uint32         `gorm:"not null;default:0;index:idx_stock;comment:库存"`
	StockWarn     uint32         `gorm:"not null;default:10;comment:库存预警值"`
	Sales         int            `gorm:"not null;default:0;index:idx_sales;comment:销量"`
	MainImage     *string        `gorm:"type:varchar(500);comment:SKU主图"`
	Status        int8           `gorm:"not null;default:1;index:idx_status;comment:状态：0-下架 1-上架"`
	CreatedAt     time.Time      `gorm:"autoCreateTime;index:idx_created_at;comment:创建时间"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	// 关联关系
	Product *Product   `gorm:"foreignKey:ProductID"`
	Images  []SkuImage `gorm:"foreignKey:SkuID"`
}
