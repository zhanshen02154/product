package repository

import (
	"context"
)

// SupplierRepository 供应商仓储接口
type SupplierRepository interface {
	// GetSupplierInfoBySkuID 获取指定SKU的供应商信息列表
	GetSupplierInfoBySkuID(ctx context.Context, skuID int64) ([]*SupplierInfo, error)
}

// SupplierInfo 供应商信息（包含供应商和关联的商品信息）
type SupplierInfo struct {
	SkuID            int64        // SKU ID
	SupplierID       int64        // 供应商 ID
	SupplyPrice      float64      // 供应价
	MinOrderQuantity int          // 最小起订量
	IsPreferred      bool         // 是否首选
	Supplier         *SupplierDTO // 供应商信息
}

// SupplierDTO 供应商数据传输对象
type SupplierDTO struct {
	ID            int64   // 供应商 ID
	Name          string  // 名称
	ContactPerson string  // 联系人
	Phone         string  // 联系电话
	Email         string  // 电子邮件
	Address       string  // 地址
	Rating        float64 // 供应商评级
	LeadTimeDays  int     // 交货周期（天）
	PaymentTerms  string  // 支付条款
}
