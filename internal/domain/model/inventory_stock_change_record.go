package model

import (
	"database/sql"
	"gorm.io/gorm"
)

// SourceType 库存变更来源类型常量
const (
	SourceTypeOrderPayment = 1 // 订单支付成功扣减库存
	SourceTypeOrderRefund  = 2 // 订单退款回补库存
	SourceTypeManual       = 3 // 手动调整库存
)

// InventoryStockChangeRecord 库存变更记录
type InventoryStockChangeRecord struct {
	ID          int64          `gorm:"column:id;primaryKey;autoIncrement"`
	OrderID     int64          `gorm:"column:order_id;not null;default:0;comment:订单ID"`
	SkuID       int64          `gorm:"column:sku_id;not null;default:0;comment:SKU ID"`
	SourceType  int32          `gorm:"column:source_type;not null;default:0;comment:来源类型:1-订单支付 2-退款 3-手动调整"`
	Quantity    int64          `gorm:"column:quantity;not null;default:0;comment:变更数量"`
	BeforeStock int64          `gorm:"column:before_stock;not null;default:0;comment:变更前库存"`
	AfterStock  int64          `gorm:"column:after_stock;not null;default:0;comment:变更后库存"`
	CreatedAt   sql.NullTime   `gorm:"column:created_at;autoCreateTime;comment:创建时间"`
	UpdatedAt   sql.NullTime   `gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:删除时间"` // GORM软删除标准字段，用于查询过滤
}

// TableName 指定表名
func (InventoryStockChangeRecord) TableName() string {
	return "inventory_stock_change_records"
}
