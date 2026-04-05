package dto

// CreateRestockApplyDto 提交补货申请DTO
type CreateRestockApplyDto struct {
	SkuID    string `json:"sku_id"`   // SKU编号(SkuNo)
	UserID   int32  `json:"user_id"`  // 用户ID（一律为-1）
	Quantity int32  `json:"quantity"` // 补货数量（必须大于0）
	Reason   string `json:"reason"`   // 补货原因
}

// RestockRecordDto 补货记录DTO
type RestockRecordDto struct {
	ID           int64  `json:"id"`
	UserID       int32  `json:"user_id"`
	SkuID        uint64 `json:"sku_id"`
	Quantity     int32  `json:"quantity"`
	Reason       string `json:"reason"`
	Status       uint8  `json:"status"`
	FailedReason string `json:"failed_reason"`
	CreatedAt    string `json:"created_at"`
}

// SkuBasicInfoDto SKU基本信息DTO
type SkuBasicInfoDto struct {
	ID            int64  `json:"id"`
	SkuNo         string `json:"sku_no"`
	SkuName       string `json:"sku_name"`
	SpecValueText string `json:"spec_value_text"`
}

// CreateRestockApplyResponseDto 提交补货申请响应DTO
type CreateRestockApplyResponseDto struct {
	RestockRecord *RestockRecordDto `json:"restock_record"`
	SkuInfo       *SkuBasicInfoDto  `json:"sku_info"`
}
