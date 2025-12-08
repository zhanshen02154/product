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
