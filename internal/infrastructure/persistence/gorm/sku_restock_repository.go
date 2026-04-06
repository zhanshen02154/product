package gorm

import (
	"context"
	"errors"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"gorm.io/gorm"
)

type SkuRestockRepositoryImpl struct {
	db *gorm.DB
}

// NewSkuRestockRepository 创建补货记录仓储实例
func NewSkuRestockRepository(db *gorm.DB) repository.SkuRestockRepository {
	return &SkuRestockRepositoryImpl{db: db}
}

// Create 创建补货记录
func (r *SkuRestockRepositoryImpl) Create(ctx context.Context, record *model.SkuRestockRecord) (*model.SkuRestockRecord, error) {
	db := GetDBFromContext(ctx, r.db)
	if err := db.Create(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

// Update 更新补货记录
func (r *SkuRestockRepositoryImpl) Update(ctx context.Context, record *model.SkuRestockRecord) error {
	db := GetDBFromContext(ctx, r.db)
	return db.Save(record).Error
}

// GetByID 根据ID查询补货记录
func (r *SkuRestockRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.SkuRestockRecord, error) {
	db := GetDBFromContext(ctx, r.db)
	var record model.SkuRestockRecord
	err := db.Where("id = ?", id).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// GetByIDWithSku 根据ID查询补货记录（包含SKU信息）
func (r *SkuRestockRepositoryImpl) GetByIDWithSku(ctx context.Context, id int64) (*model.SkuRestockRecord, error) {
	db := GetDBFromContext(ctx, r.db)
	var record model.SkuRestockRecord
	err := db.Preload("Sku.Product").Where("id = ?", id).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// GetByIDWithDetail 根据ID查询补货记录（包含SKU和审核记录）
func (r *SkuRestockRepositoryImpl) GetByIDWithDetail(ctx context.Context, id int64) (*model.SkuRestockRecord, error) {
	db := GetDBFromContext(ctx, r.db)
	var record model.SkuRestockRecord
	err := db.Preload("Sku.Product").Preload("Audits").Where("id = ?", id).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// ListBySkuID 根据SKU ID查询补货记录列表
func (r *SkuRestockRepositoryImpl) ListBySkuID(ctx context.Context, skuID uint64, offset, limit int) ([]model.SkuRestockRecord, int64, error) {
	db := GetDBFromContext(ctx, r.db)
	var records []model.SkuRestockRecord
	var total int64

	query := db.Model(&model.SkuRestockRecord{}).Where("sku_id = ?", skuID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Sku").Offset(offset).Limit(limit).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// ListByStatus 根据状态查询补货记录列表
func (r *SkuRestockRepositoryImpl) ListByStatus(ctx context.Context, status uint8, offset, limit int) ([]model.SkuRestockRecord, int64, error) {
	db := GetDBFromContext(ctx, r.db)
	var records []model.SkuRestockRecord
	var total int64

	query := db.Model(&model.SkuRestockRecord{}).Where("status = ?", status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Sku").Offset(offset).Limit(limit).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// ListByUserID 根据用户ID查询补货记录列表
func (r *SkuRestockRepositoryImpl) ListByUserID(ctx context.Context, userID int, offset, limit int) ([]model.SkuRestockRecord, int64, error) {
	db := GetDBFromContext(ctx, r.db)
	var records []model.SkuRestockRecord
	var total int64

	query := db.Model(&model.SkuRestockRecord{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Sku").Offset(offset).Limit(limit).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// UpdateStatus 更新补货状态
func (r *SkuRestockRepositoryImpl) UpdateStatus(ctx context.Context, id int64, status uint8, failedReason string) error {
	db := GetDBFromContext(ctx, r.db)
	updates := map[string]interface{}{
		"status": status,
	}
	if failedReason != "" {
		updates["failed_reason"] = failedReason
	}
	return db.Model(&model.SkuRestockRecord{}).Where("id = ?", id).Updates(updates).Error
}

// SkuRestockAuditRepositoryImpl 补货审核记录仓储实现
type SkuRestockAuditRepositoryImpl struct {
	db *gorm.DB
}

// NewSkuRestockAuditRepository 创建补货审核记录仓储实例
func NewSkuRestockAuditRepository(db *gorm.DB) repository.SkuRestockAuditRepository {
	return &SkuRestockAuditRepositoryImpl{db: db}
}

// Create 创建审核记录
func (r *SkuRestockAuditRepositoryImpl) Create(ctx context.Context, audit *model.SkuRestockAudit) (*model.SkuRestockAudit, error) {
	db := GetDBFromContext(ctx, r.db)
	if err := db.Create(audit).Error; err != nil {
		return nil, err
	}
	return audit, nil
}

// Update 更新审核记录
func (r *SkuRestockAuditRepositoryImpl) Update(ctx context.Context, audit *model.SkuRestockAudit) error {
	db := GetDBFromContext(ctx, r.db)
	return db.Save(audit).Error
}

// GetByID 根据ID查询审核记录
func (r *SkuRestockAuditRepositoryImpl) GetByID(ctx context.Context, id int64) (*model.SkuRestockAudit, error) {
	db := GetDBFromContext(ctx, r.db)
	var audit model.SkuRestockAudit
	err := db.Where("id = ?", id).First(&audit).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &audit, nil
}

// GetByRestockID 根据补货记录ID查询审核记录列表
func (r *SkuRestockAuditRepositoryImpl) GetByRestockID(ctx context.Context, restockID uint64) ([]model.SkuRestockAudit, error) {
	db := GetDBFromContext(ctx, r.db)
	var audits []model.SkuRestockAudit
	err := db.Where("restock_id = ?", restockID).Order("created_at DESC").Find(&audits).Error
	if err != nil {
		return nil, err
	}
	return audits, nil
}

// GetLatestByRestockID 根据补货记录ID查询最新审核记录
func (r *SkuRestockAuditRepositoryImpl) GetLatestByRestockID(ctx context.Context, restockID uint64) (*model.SkuRestockAudit, error) {
	db := GetDBFromContext(ctx, r.db)
	var audit model.SkuRestockAudit
	err := db.Where("restock_id = ?", restockID).Order("created_at DESC").First(&audit).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &audit, nil
}

// UpdateAuditStatus 更新审核状态
func (r *SkuRestockAuditRepositoryImpl) UpdateAuditStatus(ctx context.Context, id int64, status uint8, failedReason string) error {
	db := GetDBFromContext(ctx, r.db)
	updates := map[string]interface{}{
		"audit_status": status,
	}
	if failedReason != "" {
		updates["audit_failed_reason"] = failedReason
	}
	return db.Model(&model.SkuRestockAudit{}).Where("id = ?", id).Updates(updates).Error
}
