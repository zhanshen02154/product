package model

import (
	"gorm.io/gorm"
	"time"
)

// SpecValue 对应规格属性值表 (spec_values)
type SpecValue struct {
	ID           int64          `gorm:"primaryKey;autoIncrement;comment:属性值ID"`
	SpecID       uint           `gorm:"not null;index:idx_spec_id;comment:规格ID"`
	ValueName    string         `gorm:"type:varchar(100);not null;comment:属性值"`
	ValueImage   *string        `gorm:"type:varchar(500);comment:属性值图片"`
	DisplayOrder int            `gorm:"not null;default:0;index:idx_display_order;comment:显示顺序"`
	IsDeleted    bool           `gorm:"softDelete:flag;default:0;comment:删除标记"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// 关联关系
	Spec *ProductSpec `gorm:"foreignKey:SpecID"`
}
