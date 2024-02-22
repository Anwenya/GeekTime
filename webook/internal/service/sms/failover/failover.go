package failover

import (
	"context"
	"errors"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	"log"
	"sync/atomic"
)

type FailOverSMService struct {
	sss []sms.SMService
	idx uint64
}

func NewFailOverSMService(sss []sms.SMService) *FailOverSMService {
	return &FailOverSMService{
		sss: sss,
	}
}

// Send
// 顺序轮询
// 压力会集中在靠前的服务上
func (foss *FailOverSMService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	for _, ss := range foss.sss {
		err := ss.Send(ctx, tplId, args, numbers...)
		if err == nil {
			return nil
		}
		log.Printf("短信发送失败:%v", err)
	}
	return errors.New("短信全部发送失败")
}

// SendV1
// 每次都改变轮询的起始位置
func (foss *FailOverSMService) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.AddUint64(&foss.idx, 1)
	length := uint64(len(foss.sss))
	for i := idx; i < idx+length; i++ {
		ss := foss.sss[i%length]
		err := ss.Send(ctx, tplId, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.Canceled, context.DeadlineExceeded:
			return err
		}
		log.Printf("短信发送失败:%v", err)
	}
	return errors.New("短信全部发送失败")
}
