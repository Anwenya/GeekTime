package cache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"gorm.io/gorm/utils/tests"
	"testing"
	"time"
)

const (
	biz   = "login"
	phone = "12345678901"
	code  = "123456"
)

func TestLocalCodeCache(t *testing.T) {

	bc, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Minute*10))
	lcc := NewLocalCodeCache(bc, time.Minute*10)

	// 第一次发送
	err := lcc.Set(context.Background(), biz, phone, code)
	tests.AssertEqual(t, err, nil)

	// 发送频繁
	err = lcc.Set(context.Background(), biz, phone, code)
	tests.AssertEqual(t, err, ErrCodeSendTooMany)

	// 验证通过
	verify, err := lcc.Verify(context.Background(), biz, phone, code)
	tests.AssertEqual(t, verify, true)
	tests.AssertEqual(t, err, nil)

	// 验证失败 验证次数耗尽/频繁
	err = lcc.Set(context.Background(), biz, phone, code)
	tests.AssertEqual(t, err, nil)
	verify, err = lcc.Verify(context.Background(), biz, phone, "code")
	tests.AssertEqual(t, verify, false)
	tests.AssertEqual(t, err, nil)
	verify, err = lcc.Verify(context.Background(), biz, phone, "code")
	tests.AssertEqual(t, verify, false)
	tests.AssertEqual(t, err, nil)
	verify, err = lcc.Verify(context.Background(), biz, phone, "code")
	tests.AssertEqual(t, verify, false)
	tests.AssertEqual(t, err, nil)
	verify, err = lcc.Verify(context.Background(), biz, phone, "code")
	tests.AssertEqual(t, verify, false)
	tests.AssertEqual(t, err, ErrCodeVerifyTooMany)

	// 验证时key不存在
	verify, err = lcc.Verify(context.Background(), biz, "phone", "code")
	tests.AssertEqual(t, verify, false)
	tests.AssertEqual(t, err, ErrKeyNotExist)
}
