package failover

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	"sync/atomic"
)

type TimeoutFailoverSMService struct {
	sss []sms.SMService
	// 当前正在使用节点
	idx int32
	// 连续几个超时了
	cnt       int32
	threshold int32
}

func NewTimeoutFailoverSMService(sss []sms.SMService, threshold int32) *TimeoutFailoverSMService {
	return &TimeoutFailoverSMService{
		sss:       sss,
		threshold: threshold,
	}
}

func (tfss *TimeoutFailoverSMService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&tfss.idx)
	cnt := atomic.LoadInt32(&tfss.cnt)

	// 达到阈值 切换服务
	if cnt >= tfss.threshold {
		newIdx := (idx + 1) % int32(len(tfss.sss))
		if atomic.CompareAndSwapInt32(&tfss.idx, idx, newIdx) {
			atomic.StoreInt32(&tfss.cnt, 0)
		}
		idx = newIdx
	}

	ss := tfss.sss[idx]
	err := ss.Send(ctx, tplId, args, numbers...)
	switch err {
	case nil:
		// 正常 清除计数
		atomic.StoreInt32(&tfss.cnt, 0)
	case context.DeadlineExceeded:
		// 超时 增加计数
		atomic.StoreInt32(&tfss.cnt, 1)
	default:
		// 非超时异常 这里选择增加计数
		// 也可以考虑其他处理方式
		atomic.StoreInt32(&tfss.cnt, 1)
	}
	return err
}
