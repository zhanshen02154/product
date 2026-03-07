package model

import (
	"gorm.io/gorm"
	"time"
)

// ProductSpec 对应商品规格属性表 (product_specs)
type ProductSpec struct {
	ID           int64          `gorm:"primaryKey;autoIncrement;comment:规格ID"`
	ProductID    uint           `gorm:"not null;index:idx_product_id;comment:商品ID"`
	SpecName     string         `gorm:"type:varchar(100);not null;comment:规格名称"`
	DisplayOrder int            `gorm:"not null;default:0;index:idx_display_order;comment:显示顺序"`
	IsDeleted    bool           `gorm:"softDelete:flag;default:0;comment:删除标记"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// 关联关系
	Product    *Product    `gorm:"foreignKey:ProductID"`
	SpecValues []SpecValue `gorm:"foreignKey:SpecID"`
}
