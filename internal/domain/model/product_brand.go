package model

import (
	"gorm.io/gorm"
	"time"
)

// ProductBrand 对应品牌表 (product_brands)
type ProductBrand struct {
	ID        int64          `gorm:"primaryKey;autoIncrement;comment:品牌ID"`
	BrandName string         `gorm:"type:varchar(100);not null;comment:品牌名称"`
	Logo      *string        `gorm:"type:varchar(500);comment:品牌Logo"`
	IsDeleted bool           `gorm:"softDelete:flag;default:0;comment:删除标记"`
	CreatedAt time.Time      `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
