package model

import (
	"time"

	"gorm.io/gorm"
)

// Supplier 供应商表
type Supplier struct {
	ID            int64          `gorm:"primaryKey;autoIncrement;comment:供应商ID"`
	Name          string         `gorm:"type:varchar(100);not null;comment:供应商名称"`
	ContactPerson string         `gorm:"type:varchar(50);not null;comment:联系人"`
	Phone         string         `gorm:"type:varchar(20);not null;comment:联系电话"`
	Email         string         `gorm:"type:varchar(100);not null;comment:电子邮件"`
	Address       string         `gorm:"type:text;not null;comment:地址"`
	Rating        float64        `gorm:"type:decimal(3,2);not null;comment:供应商评级"`
	LeadTimeDays  int            `gorm:"not null;comment:交货周期（天）"`
	PaymentTerms  string         `gorm:"type:varchar(50);not null;comment:支付条款"`
	CreatedAt     time.Time      `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt     gorm.DeletedAt `gorm:"index;comment:删除时间"`
}

func (Supplier) TableName() string {
	return "suppliers"
}

// SupplierProduct 供应商商品关联表
type SupplierProduct struct {
	ID               int64     `gorm:"primaryKey;autoIncrement;comment:ID"`
	SupplierID       uint      `gorm:"not null;index:idx_supplier_id;comment:供应商ID"`
	SkuID            uint      `gorm:"not null;index:idx_sku_id;comment:商品SKU ID"`
	SupplyPrice      float64   `gorm:"type:decimal(10,2);not null;default:0.00;comment:供应价"`
	MinOrderQuantity int       `gorm:"not null;default:0;comment:最小起订量"`
	IsPreferred      bool      `gorm:"not null;default:0;comment:是否首选（0=否;1=是）"`
	CreatedAt        time.Time `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime;comment:更新时间"`

	Supplier *Supplier `gorm:"foreignKey:SupplierID"`
}

func (SupplierProduct) TableName() string {
	return "supplier_products"
}
