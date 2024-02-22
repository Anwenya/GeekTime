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

func TestFailOverSMService_Send(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) []sms.SMService

		wantErr error
	}{
		{
			name: "一次发送成功",
			mock: func(ctrl *gomock.Controller) []sms.SMService {
				ms := mocksms.NewMockSMService(ctrl)
				ms.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil)
				return []sms.SMService{ms}
			},
		},
		{
			name: "第二次发送成功",
			mock: func(ctrl *gomock.Controller) []sms.SMService {
				ms := mocksms.NewMockSMService(ctrl)
				ms.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(errors.New("发送失败"))

				ms1 := mocksms.NewMockSMService(ctrl)
				ms1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(nil)
				return []sms.SMService{ms, ms1}
			},
		},
		{
			name: "全部失败",
			mock: func(ctrl *gomock.Controller) []sms.SMService {
				ms := mocksms.NewMockSMService(ctrl)
				ms.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(errors.New("发送失败"))

				ms1 := mocksms.NewMockSMService(ctrl)
				ms1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(),
					gomock.Any()).Return(errors.New("发送失败"))
				return []sms.SMService{ms, ms1}
			},
			wantErr: errors.New("短信全部发送失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			foss := NewFailOverSMService(tc.mock(ctrl))
			err := foss.Send(context.Background(), "123", []string{"123"}, "123")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
