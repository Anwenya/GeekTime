package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (

	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode string

	ErrCodeSendTooMany   = errors.New("发送太频繁")
	ErrCodeVerifyTooMany = errors.New("验证太频繁")
	ErrCodeInfinite      = errors.New("验证码存在 但是没有过期时间")
)

type CodeCache struct {
	cmd redis.Cmdable
}

func NewCodeCache(cmd redis.Cmdable) *CodeCache {
	return &CodeCache{
		cmd: cmd,
	}
}

func (cc *CodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := cc.cmd.Eval(ctx, luaSetCode, []string{cc.key(biz, phone)}, code).Int()
	// 调用redis时异常
	if err != nil {
		return err
	}
	switch res {
	// 存在没有设置过期时间的key
	case -2:
		return ErrCodeInfinite
	// 存在key但发生时间间隔小于一分钟
	case -1:
		return ErrCodeSendTooMany
	// 可以正常发送
	default:
		return nil
	}
}

func (cc *CodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	res, err := cc.cmd.Eval(ctx, luaVerifyCode, []string{cc.key(biz, phone)}, code).Int()
	// 调用redis时异常
	if err != nil {
		return false, err
	}

	switch res {
	// 验证码不正确
	case -2:
		return false, nil
	// 验证次数耗尽了 或者 已经过期了
	case -1:
		return false, ErrCodeVerifyTooMany
	// 验证成功
	default:
		return true, nil
	}
}

func (cc *CodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
