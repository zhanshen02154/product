package model

import (
	"gorm.io/gorm"
	"time"
)

// SkuImage 对应SKU图片表 (sku_images)
type SkuImage struct {
	ID           int64          `gorm:"primaryKey;autoIncrement;comment:图片ID"`
	SkuID        uint           `gorm:"not null;index:idx_sku_id;comment:SKU ID"`
	ImageURL     string         `gorm:"type:varchar(500);not null;comment:图片URL"`
	IsMain       bool           `gorm:"not null;default:0;comment:是否主图：0-否 1-是"`
	DisplayOrder int            `gorm:"not null;default:0;index:idx_display_order;comment:显示顺序"`
	IsDeleted    bool           `gorm:"softDelete:flag;default:0;comment:删除标记"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// 关联关系
	Sku *ProductSku `gorm:"foreignKey:SkuID"`
}
