package events

import (
	"context"
)

type Producer interface {
	ProduceSyncEvent(ctx context.Context, data BizTags) error
}

type SyncDataEvent struct {
	IndexName string
	DocID     string
	// 这里应该是 BizTags
	Data string
}
