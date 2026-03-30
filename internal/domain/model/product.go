package model

import (
	"gorm.io/gorm"
	"time"
)

// Product 对应商品表 (products)
type Product struct {
	ID          int64          `gorm:"primaryKey;autoIncrement;comment:商品ID"`
	ProductNo   string         `gorm:"type:varchar(64);not null;uniqueIndex:uk_product_no;comment:商品编号"`
	ProductName string         `gorm:"type:varchar(255);not null;comment:商品名称"`
	CategoryID  uint           `gorm:"not null;index:idx_category;comment:分类ID"`
	BrandID     *uint          `gorm:"index:idx_brand;comment:品牌ID"`
	MainImage   *string        `gorm:"type:varchar(500);comment:主图"`
	Description *string        `gorm:"type:text;comment:商品描述"`
	Status      int8           `gorm:"not null;default:1;index:idx_status;comment:状态：0-下架 1-上架"`
	IsDeleted   bool           `gorm:"softDelete:flag;default:0;comment:删除标记"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;index:idx_created_at;comment:创建时间"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt   gorm.DeletedAt `gorm:"index"` // GORM软删除标准字段，用于查询过滤

	// 关联关系
	Category *ProductCategory `gorm:"foreignKey:CategoryID"`
	Brand    *ProductBrand    `gorm:"foreignKey:BrandID"`
	Specs    []ProductSpec    `gorm:"foreignKey:ProductID"`
	Skus     []ProductSku     `gorm:"foreignKey:ProductID"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}
