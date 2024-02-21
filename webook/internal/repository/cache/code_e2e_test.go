package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRedisCodeCache_Set_e2e(t *testing.T) {
	// 访问真实的redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "192.168.133.128:6379",
	})
	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "设置成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:login:12345678911"
				// 检查过期时间
				dur, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, dur > time.Minute*9+time.Second*50)
				// 检查
				code, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", code)
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "12345678911",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				// 先设置一个值用于触发频繁
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:login:12345678911"
				err := rdb.Set(ctx, key, "123456", time.Minute*9+time.Second*30).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:login:12345678911"
				// 检查过期时间
				dur, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, dur > time.Minute*9+time.Second*10)
				// 检查
				code, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", code)
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "12345678911",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
		{
			name: "系统错误",
			before: func(t *testing.T) {
				// 先设置一个没有过期时间的值 触发系统错误
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:login:12345678911"
				err := rdb.Set(ctx, key, "123456", 0).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:login:12345678911"
				// 检查过期时间
				dur, err := rdb.TTL(ctx, key).Result()
				fmt.Println(dur)
				assert.NoError(t, err)
				assert.True(t, dur == time.Nanosecond*-1)
				// 检查
				code, err := rdb.GetDel(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", code)
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "12345678911",
			code:    "123456",
			wantErr: ErrCodeInfinite,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			c := NewCodeCache(rdb)
			err := c.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
