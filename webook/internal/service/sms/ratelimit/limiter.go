package ratelimit

import (
	"context"
	"errors"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	"github.com/Anwenya/GeekTime/webook/pkg/limiter"
)

var errLimited = errors.New("触发限流")

type RateLimitSMService struct {
	ss      sms.SMService
	limiter limiter.Limiter
	key     string
}

func (rlss *RateLimitSMService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	// 装饰
	limited, err := rlss.limiter.Limit(ctx, rlss.key)
	if err != nil {
		return err
	}
	if limited {
		return errLimited
	}
	// 调用被装饰的服务
	return rlss.ss.Send(ctx, tplId, args, numbers...)
}

func NewRateLimitSMService(ss sms.SMService, limiter limiter.Limiter) *RateLimitSMService {
	return &RateLimitSMService{
		ss:      ss,
		limiter: limiter,
		key:     "sms-limiter",
	}
}
