package domain

import "time"

type Interactive struct {
	Biz        string
	BizId      int64
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Liked      bool
	Collected  bool
}

type ReadHistory struct {
	BizId    int64
	Biz      string
	Uid      int64
	ReadTime time.Time
}
