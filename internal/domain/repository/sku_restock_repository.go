package repository

import (
	"context"
	"github.com/zhanshen02154/product/internal/domain/model"
)

// SkuRestockRepository 补货记录仓储接口
type SkuRestockRepository interface {
	// Create 创建补货记录
	Create(ctx context.Context, record *model.SkuRestockRecord) (*model.SkuRestockRecord, error)

	// Update 更新补货记录
	Update(ctx context.Context, record *model.SkuRestockRecord) error

	// GetByID 根据ID查询补货记录
	GetByID(ctx context.Context, id int64) (*model.SkuRestockRecord, error)

	// GetByIDWithSku 根据ID查询补货记录（包含SKU信息）
	GetByIDWithSku(ctx context.Context, id int64) (*model.SkuRestockRecord, error)

	// GetByIDWithDetail 根据ID查询补货记录（包含SKU和审核记录）
	GetByIDWithDetail(ctx context.Context, id int64) (*model.SkuRestockRecord, error)

	// ListBySkuID 根据SKU ID查询补货记录列表
	ListBySkuID(ctx context.Context, skuID uint64, offset, limit int) ([]model.SkuRestockRecord, int64, error)

	// ListByStatus 根据状态查询补货记录列表
	ListByStatus(ctx context.Context, status uint8, offset, limit int) ([]model.SkuRestockRecord, int64, error)

	// ListByUserID 根据用户ID查询补货记录列表
	ListByUserID(ctx context.Context, userID int, offset, limit int) ([]model.SkuRestockRecord, int64, error)

	// UpdateStatus 更新补货状态
	UpdateStatus(ctx context.Context, id int64, status uint8, failedReason string) error
}

// SkuRestockAuditRepository 补货审核记录仓储接口
type SkuRestockAuditRepository interface {
	// Create 创建审核记录
	Create(ctx context.Context, audit *model.SkuRestockAudit) (*model.SkuRestockAudit, error)

	// Update 更新审核记录
	Update(ctx context.Context, audit *model.SkuRestockAudit) error

	// GetByID 根据ID查询审核记录
	GetByID(ctx context.Context, id int64) (*model.SkuRestockAudit, error)

	// GetByRestockID 根据补货记录ID查询审核记录列表
	GetByRestockID(ctx context.Context, restockID uint64) ([]model.SkuRestockAudit, error)

	// GetLatestByRestockID 根据补货记录ID查询最新审核记录
	GetLatestByRestockID(ctx context.Context, restockID uint64) (*model.SkuRestockAudit, error)

	// UpdateAuditStatus 更新审核状态
	UpdateAuditStatus(ctx context.Context, id int64, status uint8, failedReason string) error
}
