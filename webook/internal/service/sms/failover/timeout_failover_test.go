package failover

import (
	"context"
	"errors"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	mocksms "github.com/Anwenya/GeekTime/webook/internal/service/sms/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestTimeoutFailoverSMService_Send(t *testing.T) {
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) []sms.SMService
		threshold int32
		idx       int32
		cnt       int32

		wantErr error
		wantCnt int32
		wantIdx int32
	}{
		{
			name: "没有触发切换",
			mock: func(ctrl *gomock.Controller) []sms.SMService {
				ms := mocksms.NewMockSMService(ctrl)
				ms.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(nil)
				return []sms.SMService{ms}
			},
			idx:       0,
			cnt:       12,
			threshold: 15,

			wantCnt: 0,
			wantErr: nil,
		},
		{
			name: "触发切换 成功",
			mock: func(ctrl *gomock.Controller) []sms.SMService {
				ms := mocksms.NewMockSMService(ctrl)
				ms1 := mocksms.NewMockSMService(ctrl)
				ms1.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(nil)
				return []sms.SMService{ms, ms1}
			},
			idx:       0,
			cnt:       15,
			threshold: 15,

			wantIdx: 1,
			wantCnt: 0,
			wantErr: nil,
		},
		{
			name: "触发切换 失败",
			mock: func(ctrl *gomock.Controller) []sms.SMService {
				ms := mocksms.NewMockSMService(ctrl)
				ms1 := mocksms.NewMockSMService(ctrl)
				ms1.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(errors.New("发送失败"))
				return []sms.SMService{ms, ms1}
			},
			idx:       0,
			cnt:       15,
			threshold: 15,

			wantIdx: 1,
			wantCnt: 1,
			wantErr: errors.New("发送失败"),
		},
		{
			name: "触发切换 失败",
			mock: func(ctrl *gomock.Controller) []sms.SMService {
				ms := mocksms.NewMockSMService(ctrl)
				ms1 := mocksms.NewMockSMService(ctrl)
				ms1.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
				return []sms.SMService{ms, ms1}
			},
			idx:       0,
			cnt:       15,
			threshold: 15,

			wantIdx: 1,
			wantCnt: 1,
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tfss := NewTimeoutFailoverSMService(tc.mock(ctrl), tc.threshold)
			tfss.cnt = tc.cnt
			tfss.idx = tc.idx
			err := tfss.Send(context.Background(), "123",
				[]string{"123", "123"}, "123456")
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantCnt, tfss.cnt)
			assert.Equal(t, tc.wantIdx, tfss.idx)
		})
	}
}
