package job

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/payment/service/wechat"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"time"
)

// SyncWechatOrderJob
// 该任务用于定时扫描本地超时的订单 然后向业务方核对该订单的状态
type SyncWechatOrderJob struct {
	svc *wechat.NativePaymentService
	l   logger.LoggerV1
}

func (s *SyncWechatOrderJob) Name() string {
	return "sync_wechat_order_job"
}

func (s *SyncWechatOrderJob) Run() error {
	// 在发起支付请求的时候设置了订单的超时时间是30分钟
	// 这里扫描31分钟前的订单 避免一些特殊情况
	t := time.Now().Add(-time.Minute * 31)
	offset := 0
	const limit = 100
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		pmts, err := s.svc.FindExpiredPayment(ctx, offset, limit, t)
		cancel()
		if err != nil {
			return err
		}
		for _, pmt := range pmts {
			ctx, cancel = context.WithTimeout(context.Background(), time.Second*3)
			err = s.svc.SyncWechatInfo(ctx, pmt.BizTradeNO)
			cancel()
			if err != nil {
				s.l.Error(
					"同步微信支付状态失败",
					logger.Error(err),
					logger.String("biz_trade_no", pmt.BizTradeNO),
				)
			}
		}

		if len(pmts) < limit {
			return nil
		}

		offset += len(pmts)
	}
}
