package service

import (
	"context"
	"errors"

	"github.com/zhanshen02154/product/internal/application/dto"
	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
)

type ISkuRestockService interface {
	CreateRestockApply(ctx context.Context, req *dto.CreateRestockApplyDto) (*dto.CreateRestockApplyResponseDto, error)
	GetRestockApplyInfo(ctx context.Context, applicationNo string, userID int) (*model.SkuRestockRecord, error)
}

// NewSkuRestockService 创建补货服务
func NewSkuRestockService(
	skuRepo repository.ProductSkuRepository,
	restockRepo repository.SkuRestockRepository,
	auditRepo repository.SkuRestockAuditRepository,
) ISkuRestockService {
	return &SkuRestockService{
		skuRepo:     skuRepo,
		restockRepo: restockRepo,
		auditRepo:   auditRepo,
	}
}

type SkuRestockService struct {
	skuRepo     repository.ProductSkuRepository
	restockRepo repository.SkuRestockRepository
	auditRepo   repository.SkuRestockAuditRepository
}

// CreateRestockApply 提交补货申请
func (s *SkuRestockService) CreateRestockApply(ctx context.Context, req *dto.CreateRestockApplyDto) (*dto.CreateRestockApplyResponseDto, error) {
	// 1. 参数验证
	if req.SkuID == "" {
		return nil, errors.New("sku_id不能为空")
	}
	if req.Quantity <= 0 {
		return nil, errors.New("补货数量必须大于0")
	}
	if req.Reason == "" {
		return nil, errors.New("补货原因不能为空")
	}

	// 2. 检查SKU是否存在
	sku, err := s.skuRepo.GetSkuStockBySkuNo(ctx, req.SkuID)
	if err != nil {
		return nil, err
	}
	if sku == nil {
		return nil, errors.New("SKU不存在")
	}

	// 3. 检查SKU状态（禁止对已下架的SKU发起补货申请）
	if sku.Status == 0 {
		return nil, errors.New("该型号的商品已经下架")
	}

	if sku.Stock > sku.StockWarn {
		return nil, errors.New("该型号的商品库存充足")
	}

	// 4. 创建补货记录
	restockRecord := &model.SkuRestockRecord{
		UserID:       int(req.UserID),
		SkuID:        uint64(sku.ID),
		Quantity:     req.Quantity,
		Reason:       req.Reason,
		Status:       model.RestockStatusPending, // 待订货
		FailedReason: "",
	}

	// 5. 保存补货记录
	createdRecord, err := s.restockRepo.Create(ctx, restockRecord)
	if err != nil {
		return nil, err
	}

	// 6. 构建响应
	response := &dto.CreateRestockApplyResponseDto{
		RestockRecord: &dto.RestockRecordDto{
			ID:           createdRecord.ID,
			UserID:       int32(createdRecord.UserID),
			SkuID:        createdRecord.SkuID,
			Quantity:     createdRecord.Quantity,
			Reason:       createdRecord.Reason,
			Status:       createdRecord.Status,
			FailedReason: createdRecord.FailedReason,
		},
		SkuInfo: &dto.SkuBasicInfoDto{
			ID:            sku.ID,
			SkuNo:         sku.SkuNo,
			SkuName:       sku.SkuName,
			SpecValueText: sku.SpecValueText,
		},
	}

	// 设置创建时间
	if createdRecord.CreatedAt.Valid {
		response.RestockRecord.CreatedAt = createdRecord.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}

	return response, nil
}

// GetRestockApplyInfo 获取补货申请信息
func (s *SkuRestockService) GetRestockApplyInfo(ctx context.Context, applicationNo string, userID int) (*model.SkuRestockRecord, error) {
	// 1. 根据业务流水号和用户ID查询补货记录
	record, err := s.restockRepo.GetByApplicationNo(ctx, applicationNo, userID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}

	// 2. 查询最新审核记录
	audit, err := s.auditRepo.GetLatestByRestockID(ctx, uint64(record.ID))
	if err != nil {
		return nil, err
	}

	// 3. 设置审核记录
	if audit != nil {
		record.Audits = []model.SkuRestockAudit{*audit}
	}

	return record, nil
}
