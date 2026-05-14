package gorm

import (
	"context"
	"math"
	"time"

	"github.com/zhanshen02154/product/internal/domain/model"
	"github.com/zhanshen02154/product/internal/domain/repository"
	"gorm.io/gorm"
)

type InventoryStockChangeRecordRepositoryImpl struct {
	db *gorm.DB
}

// BatchCreate 批量创建库存变更记录
func (r *InventoryStockChangeRecordRepositoryImpl) BatchCreate(ctx context.Context, records []*model.InventoryStockChangeRecord) error {
	if len(records) == 0 {
		return nil
	}
	db := GetDBFromContext(ctx, r.db)
	return db.CreateInBatches(records, 100).Error
}

// GetSalesVolume 获取SKU在指定时间范围内的销量（订单支付扣减库存的数量总和）和日均销量
// 销量 = SUM(ABS(quantity)) WHERE source_type = 1 (订单支付) AND sku_id = ? AND created_at BETWEEN ? AND ?
// 日均销量 = 总销量 / 天数
func (r *InventoryStockChangeRecordRepositoryImpl) GetSalesVolume(ctx context.Context, skuID int64, startTime, endTime string) (int64, float64, error) {
	db := GetDBFromContext(ctx, r.db)

	var total int64
	query := db.Model(&model.InventoryStockChangeRecord{}).
		Select("COALESCE(SUM(ABS(quantity)), 0)").
		Where("sku_id = ?", skuID).
		Where("source_type = ?", model.SourceTypeOrderPayment)

	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	err := query.Scan(&total).Error
	if err != nil {
		return 0, 0, err
	}

	// 计算天数
	start, _ := time.Parse("2006-01-02 15:04:05", startTime)
	end, _ := time.Parse("2006-01-02 15:04:05", endTime)

	if startTime == "" {
		start = time.Now()
	}
	if endTime == "" {
		end = time.Now()
	}

	days := int64(end.Sub(start).Hours()/24) + 1
	if days <= 0 {
		days = 1
	}

	var dailyAvg float64
	dailyAvg = math.Round((float64(total)/float64(days))*100) / 100

	return total, dailyAvg, nil
}

// GetDailySales 获取SKU在指定时间范围内的每日销量数据
func (r *InventoryStockChangeRecordRepositoryImpl) GetDailySales(ctx context.Context, skuID int64, skuCode string, startDate, endDate string) ([]*repository.DailySalesData, error) {
	db := GetDBFromContext(ctx, r.db)

	var results []*repository.DailySalesData
	err := db.Model(&model.InventoryStockChangeRecord{}).
		Select("DATE_FORMAT(created_at, '%Y-%m-%d') as date, COALESCE(SUM(ABS(quantity)), 0) as sales_volume").
		Where("sku_id = ?", skuID).
		Where("source_type = ?", model.SourceTypeOrderPayment).
		Where("DATE(created_at) >= ?", startDate).
		Where("DATE(created_at) <= ?", endDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 填充 sku_code
	for _, result := range results {
		result.SkuCode = skuCode
	}

	return results, nil
}

// NewInventoryStockChangeRecordRepository 创建库存变更记录仓储实例
func NewInventoryStockChangeRecordRepository(db *gorm.DB) repository.InventoryStockChangeRecordRepository {
	return &InventoryStockChangeRecordRepositoryImpl{db: db}
}
