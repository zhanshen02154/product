package model

import (
	"database/sql"
)

// OrderInventoryEvent 订单库存事件记录表
type OrderInventoryEvent struct {
	Id        int64        `gorm:"type:bigint(20);primaryKey;not null;autoIncrement" json:"id"`
	EventId   string       `gorm:"type:varchar(50);not null;default:''" json:"event_id"`
	OrderId   int64        `gorm:"type:bigint(20);not null;default:0" json:"order_id"`
	CreatedAt sql.NullTime `gorm:"type:datetime;comment:'创建时间'" json:"created_at"`
	UpdatedAt sql.NullTime `gorm:"type:datetime;comment:'更新时间'" json:"updated_at"`
}
