package failover

import (
	"context"
	"errors"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	"log"
)

type FailOverSMService struct {
	sss []sms.SMService
	ids uint64
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
		if err != nil {
			return nil
		}
		log.Printf("短信发送失败:%v", err)
	}
	return errors.New("短信全部发送失败")
}
