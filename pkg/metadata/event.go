package metadata

import (
	"context"
	"go-micro.dev/v4/metadata"
)

const eventIdKey = "Event_id"

// GetEventId 获取事件ID
func GetEventId(ctx context.Context) (string, bool) {
	return metadata.Get(ctx, eventIdKey)
}

// GetValueFromMetadata 从元数据里获得指定键的值
func GetValueFromMetadata(ctx context.Context, key string) string {
	val, ok := metadata.Get(ctx, key)
	if !ok {
		return ""
	}
	return val
}
