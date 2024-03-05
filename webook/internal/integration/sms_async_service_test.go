package integration

import (
	"errors"
	"github.com/Anwenya/GeekTime/webook/internal/integration/startup"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	mocksms "github.com/Anwenya/GeekTime/webook/internal/service/sms/mocks"
	"github.com/ecodeclub/ekit/sqlx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"testing"
	"time"
)

type AsyncSMSTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (s *AsyncSMSTestSuite) SetupSuite() {
	l := startup.InitLogger()
	s.db = startup.InitDB(l)
}

func (s *AsyncSMSTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE table `async_sms_tasks`")
}

func (s *AsyncSMSTestSuite) TestSend() {
	t := s.T()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) sms.SMService

		tplId   string
		args    []string
		numbers []string

		wantErr error
	}{
		{
			name: "异步",
			mock: func(ctrl *gomock.Controller) sms.SMService {
				svc := mocksms.NewMockSMService(ctrl)
				return svc
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := startup.InitAsyncSMSService(tc.mock(ctrl))
			err := svc.Send(context.Background(), tc.tplId, tc.args, tc.numbers...)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func (s *AsyncSMSTestSuite) TestAsyncCycle() {
	now := time.Now()
	testCases := []struct {
		name string
		// 虽然是集成测试，但是我们也不想真的发短信，所以用 mock
		mock func(ctrl *gomock.Controller) sms.SMService
		// 准备数据
		before func(t *testing.T)
		after  func(t *testing.T)
	}{
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) sms.SMService {
				svc := mocksms.NewMockSMService(ctrl)
				svc.EXPECT().Send(gomock.Any(), "123",
					[]string{"123456"}, []string{"15212345678"}).
					Return(nil)
				return svc
			},
			before: func(t *testing.T) {
				// 准备一条数据
				err := s.db.Create(
					&dao.AsyncSMSTask{
						Id: 1,
						Config: sqlx.JsonColumn[dao.SMSConfig]{
							Val: dao.SMSConfig{
								TplId:   "123",
								Args:    []string{"123456"},
								Numbers: []string{"15212345678"},
							},
							Valid: true,
						},

						RetryMax:   3,
						Status:     0,
						CreateTime: now.Add(-time.Minute * 2).UnixMilli(),
						UpdateTime: now.Add(-time.Minute * 2).UnixMilli(),
					},
				).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据
				var ast dao.AsyncSMSTask
				err := s.db.Where("id=?", 1).First(&ast).Error
				assert.NoError(t, err)
				assert.Equal(t, uint8(2), ast.Status)
			},
		},
		{
			name: "发送失败 标记为失败",
			mock: func(ctrl *gomock.Controller) sms.SMService {
				svc := mocksms.NewMockSMService(ctrl)
				svc.EXPECT().Send(gomock.Any(), "123",
					[]string{"123456"}, []string{"15212345678"}).
					Return(errors.New("模拟失败"))
				return svc
			},
			before: func(t *testing.T) {
				// 准备一条数据
				err := s.db.Create(
					&dao.AsyncSMSTask{
						Id: 2,
						Config: sqlx.JsonColumn[dao.SMSConfig]{
							Val: dao.SMSConfig{
								TplId:   "123",
								Args:    []string{"123456"},
								Numbers: []string{"15212345678"},
							},
							Valid: true,
						},
						RetryMax:   3,
						RetryCnt:   2,
						Status:     0,
						CreateTime: now.Add(-time.Minute * 2).UnixMilli(),
						UpdateTime: now.Add(-time.Minute * 2).UnixMilli(),
					},
				).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据
				var as dao.AsyncSMSTask
				err := s.db.Where("id=?", 2).First(&as).Error
				assert.NoError(t, err)
				assert.Equal(t, uint8(1), as.Status)
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tc.before(t)
			svc := startup.InitAsyncSMSService(tc.mock(ctrl))
			defer tc.after(t)
			svc.AsyncSend()
		})
	}
}

func TestAsyncSmsService(t *testing.T) {
	suite.Run(t, &AsyncSMSTestSuite{})
}
