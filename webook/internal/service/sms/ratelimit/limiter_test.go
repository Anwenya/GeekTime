package ratelimit

import (
	"context"
	"errors"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	mocksms "github.com/Anwenya/GeekTime/webook/internal/service/sms/mocks"
	"github.com/Anwenya/GeekTime/webook/pkg/limiter"
	mocklimitermock "github.com/Anwenya/GeekTime/webook/pkg/limiter/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRateLimitSMService_Send(t *testing.T) {
	// 不需要输入
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (sms.SMService, limiter.Limiter)
		wantErr error
	}{
		{
			name: "不限流",
			mock: func(ctrl *gomock.Controller) (sms.SMService, limiter.Limiter) {
				ss := mocksms.NewMockSMService(ctrl)
				l := mocklimitermock.NewMockLimiter(ctrl)
				// 不用管输入 直接通过限流即可
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				ss.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return ss, l
			},
		},
		{
			name: "限流",
			mock: func(ctrl *gomock.Controller) (sms.SMService, limiter.Limiter) {
				ss := mocksms.NewMockSMService(ctrl)
				l := mocklimitermock.NewMockLimiter(ctrl)
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				return ss, l
			},
			wantErr: errLimited,
		},
		{
			name: "异常",
			mock: func(ctrl *gomock.Controller) (sms.SMService, limiter.Limiter) {
				ss := mocksms.NewMockSMService(ctrl)
				l := mocklimitermock.NewMockLimiter(ctrl)
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, errors.New("异常"))
				return ss, l
			},
			wantErr: errors.New("异常"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ss, l := tc.mock(ctrl)
			lss := NewRateLimitSMService(ss, l)
			err := lss.Send(context.Background(), "123", []string{"123"}, "123")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
