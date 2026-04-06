package model

import (
	"database/sql"
	"time"
)

// 补货状态常量
const (
	RestockStatusPending = 1 // 待订货
	RestockStatusPartial = 2 // 部分订货
	RestockStatusOrdered = 3 // 已订货
	RestockStatusFailed  = 4 // 失败
)

// 审核状态常量
const (
	AuditStatusPending  = 1 // 待审核
	AuditStatusApproved = 2 // 审核成功
	AuditStatusRejected = 3 // 审核失败
)

// SkuRestockRecord 补货记录表
type SkuRestockRecord struct {
	ID           int64        `gorm:"primaryKey;autoIncrement;comment:ID" json:"id"`
	UserID       int          `gorm:"column:user_id;not null;default:-1;index:idx_user_id;comment:用户ID" json:"user_id"`
	SkuID        uint64       `gorm:"column:sku_id;not null;default:0;index:idx_sku_id;comment:SKU ID" json:"sku_id"`
	Quantity     int32        `gorm:"column:quantity;not null;comment:补货数量" json:"quantity"`
	Reason       string       `gorm:"column:reason;type:varchar(255);not null;comment:补货原因" json:"reason"`
	Status       uint8        `gorm:"column:status;not null;default:1;index:idx_status;comment:补货状态:1=待订货 2=部分订货 3=已订货 4=失败" json:"status"`
	FailedReason string       `gorm:"column:failed_reason;type:varchar(200);not null;default:'';comment:补货失败原因" json:"failed_reason"`
	CreatedAt    sql.NullTime `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    sql.NullTime `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    sql.NullTime `gorm:"column:deleted_at;index:idx_deleted_at;comment:删除时间" json:"deleted_at"`

	// 关联关系
	Sku    *ProductSku       `gorm:"foreignKey:SkuID;references:ID" json:"sku"`
	Audits []SkuRestockAudit `gorm:"foreignKey:RestockID;references:ID" json:"audits"`
}

// TableName 指定表名
func (SkuRestockRecord) TableName() string {
	return "sku_restock_records"
}

// SkuRestockAudit 补货审核记录
type SkuRestockAudit struct {
	ID                int64        `gorm:"primaryKey;autoIncrement;comment:ID"`
	RestockID         uint64       `gorm:"column:restock_id;not null;default:0;index:idx_restock_id;comment:补货记录ID"`
	AuditUserID       uint         `gorm:"column:audit_user_id;not null;default:0;comment:审核用户ID"`
	AuditStatus       uint8        `gorm:"column:audit_status;not null;default:1;index:idx_audit_status;comment:审核状态:1=待审核 2=审核成功 3=审核失败"`
	AuditFailedReason string       `gorm:"column:audit_failed_reason;type:varchar(200);not null;default:'';comment:审核失败原因"`
	CreatedAt         time.Time    `gorm:"column:created_at;autoCreateTime;comment:创建时间"`
	UpdatedAt         time.Time    `gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
	DeletedAt         sql.NullTime `gorm:"column:deleted_at;index:idx_deleted_at;comment:删除时间"`

	// 关联关系
	Restock *SkuRestockRecord `gorm:"foreignKey:RestockID;references:ID"`
}

// TableName 指定表名
func (SkuRestockAudit) TableName() string {
	return "sku_restock_audit"
}
