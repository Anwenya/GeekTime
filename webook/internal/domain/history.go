package domain

import "time"

type ReadHistory struct {
	BizId    int64
	Biz      string
	Uid      int64
	ReadTime time.Time
}
