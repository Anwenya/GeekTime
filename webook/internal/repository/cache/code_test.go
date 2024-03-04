package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache/mockredis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRedisCodeCache_Set(t *testing.T) {
	keyFunc := func(biz, phone string) string {
		return fmt.Sprintf("phone_code:%s:%s", biz, phone)
	}
	testCases := []struct {
		name  string
		mock  func(ctrl *gomock.Controller) redis.Cmdable
		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := mockredis.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(nil)
				cmd.SetVal(int64(0))
				res.EXPECT().Eval(
					gomock.Any(),
					luaSetCode,
					[]string{keyFunc("test", "123")},
					[]any{"123"},
				).Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "123",
			code:    "123",
			wantErr: nil,
		},
		{
			name: "redis返回error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := mockredis.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetErr(errors.New("redis异常"))
				res.EXPECT().Eval(
					gomock.Any(),
					luaSetCode,
					[]string{keyFunc("test", "123")},
					[]any{"123"},
				).Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "123",
			code:    "123",
			wantErr: errors.New("redis异常"),
		},
		{
			name: "没有过期时间",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := mockredis.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetVal(int64(-2))
				res.EXPECT().Eval(
					gomock.Any(),
					luaSetCode,
					[]string{keyFunc("test", "123")},
					[]any{"123"},
				).Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "123",
			code:    "123",
			wantErr: ErrCodeInfinite,
		},
		{
			name: "发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				res := mockredis.NewMockCmdable(ctrl)
				cmd := redis.NewCmd(context.Background())
				cmd.SetVal(int64(-1))
				res.EXPECT().Eval(
					gomock.Any(),
					luaSetCode,
					[]string{keyFunc("test", "123")},
					[]any{"123"},
				).Return(cmd)
				return res
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "123",
			code:    "123",
			wantErr: ErrCodeSendTooMany,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cc := NewRedisCodeCache(tc.mock(ctrl))
			err := cc.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
