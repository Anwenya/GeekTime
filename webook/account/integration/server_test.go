package integration

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/account/domain"
	"github.com/Anwenya/GeekTime/webook/account/grpc"
	"github.com/Anwenya/GeekTime/webook/account/integration/startup"
	"github.com/Anwenya/GeekTime/webook/account/repository/dao"
	accountv1 "github.com/Anwenya/GeekTime/webook/api/proto/gen/account/v1"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
	"time"
)

type AccountServiceServerTestSuite struct {
	suite.Suite
	db     *gorm.DB
	server *grpc.AccountServiceServer
}

func (s *AccountServiceServerTestSuite) SetupSuite() {
	l := logger.NopLogger{}

	s.db = startup.InitDB(&l)

	// 创建一个系统账户
	now := time.Now().UnixMilli()
	err := s.db.Create(&dao.Account{
		Type:       domain.AccountTypeSystem,
		Currency:   "CNY",
		CreateTime: now,
		UpdateTime: now,
	}).Error
	require.NoError(s.T(), err)

	s.server = startup.Init()
}

func (s *AccountServiceServerTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE TABLE `accounts`")
	s.db.Exec("TRUNCATE TABLE `account_activities`")
}

func (s *AccountServiceServerTestSuite) TestCredit() {
	testCases := []struct {
		name    string
		before  func(t *testing.T)
		after   func(t *testing.T)
		req     *accountv1.CreditRequest
		wantErr error
	}{
		{
			name:   "用户账号不存在",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()

				var sysAccount dao.Account
				err := s.db.WithContext(ctx).Where("type = ?", uint8(accountv1.AccountType_AccountTypeSystem)).
					First(&sysAccount).Error

				require.NoError(t, err)
				// 比较余额
				require.Equal(t, int64(10), sysAccount.Balance)

				var userAccount dao.Account
				err = s.db.WithContext(ctx).Where("uid = ?", 100).First(&userAccount).Error
				require.NoError(t, err)

				userAccount.Id = 0
				require.True(t, userAccount.CreateTime > 0)
				userAccount.CreateTime = 0
				require.True(t, userAccount.UpdateTime > 0)
				userAccount.UpdateTime = 0

				require.Equal(t, dao.Account{
					Account:  123,
					Uid:      100,
					Type:     uint8(accountv1.AccountType_AccountTypeReward),
					Balance:  100,
					Currency: "CNY",
				}, userAccount)
			},
			req: &accountv1.CreditRequest{
				Biz:   "test",
				BizId: 123,
				Items: []*accountv1.CreditItem{
					{
						Account:     123,
						AccountType: accountv1.AccountType_AccountTypeReward,
						Amount:      100,
						Currency:    "CNY",
						Uid:         100,
					},
					{
						AccountType: accountv1.AccountType_AccountTypeSystem,
						Amount:      10,
						Currency:    "CNY",
					},
				},
			},
		},
		{
			name: "用户账号存在",
			before: func(t *testing.T) {
				err := s.db.Create(&dao.Account{
					Uid:        1025,
					Account:    1234,
					Type:       uint8(accountv1.AccountType_AccountTypeReward),
					Balance:    300,
					Currency:   "CNY",
					CreateTime: 1111,
					UpdateTime: 2222,
				}).Error
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var usrAccount dao.Account
				err := s.db.WithContext(ctx).Where("uid = ?", 1025).
					First(&usrAccount).Error
				require.NoError(t, err)
				usrAccount.Id = 0
				require.True(t, usrAccount.CreateTime > 0)
				usrAccount.CreateTime = 0
				require.True(t, usrAccount.UpdateTime > 0)
				usrAccount.UpdateTime = 0
				require.Equal(t, dao.Account{
					Account:  1234,
					Uid:      1025,
					Type:     uint8(accountv1.AccountType_AccountTypeReward),
					Balance:  400,
					Currency: "CNY",
				}, usrAccount)
			},
			req: &accountv1.CreditRequest{
				Biz:   "test",
				BizId: 321,
				Items: []*accountv1.CreditItem{
					{
						Account:     1234,
						Uid:         1025,
						AccountType: accountv1.AccountType_AccountTypeReward,
						Amount:      100,
						Currency:    "CNY",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			_, err := s.server.Credit(context.Background(), tc.req)
			require.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func TestAccountServiceServer(t *testing.T) {
	suite.Run(t, new(AccountServiceServerTestSuite))
}
