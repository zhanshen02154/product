package model

import (
	"gorm.io/gorm"
	"time"
)

// ProductCategory 对应商品分类表 (product_categories)
type ProductCategory struct {
	ID           int64          `gorm:"primaryKey;autoIncrement;comment:分类ID"`
	CategoryName string         `gorm:"type:varchar(100);not null;comment:分类名称"`
	ParentID     uint           `gorm:"index:idx_parent_id;default:0;comment:父分类ID"`
	Level        int            `gorm:"not null;default:1;comment:分类层级"`
	IsDeleted    bool           `gorm:"softDelete:flag;default:0;comment:删除标记"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// 自关联 (树形结构)
	Parent        *ProductCategory  `gorm:"foreignKey:ParentID"`
	SubCategories []ProductCategory `gorm:"foreignKey:ParentID"`
}

// TableName 指定表名
func (ProductCategory) TableName() string {
	return "product_categories"
}
