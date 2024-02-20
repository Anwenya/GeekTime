package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"sync"
	"time"
)

type LocalCodeCache struct {
	cache      *bigcache.BigCache
	mutex      sync.Mutex
	expiration time.Duration
}

func NewLocalCodeCache(c *bigcache.BigCache, expiration time.Duration) *LocalCodeCache {
	return &LocalCodeCache{
		cache:      c,
		expiration: expiration,
	}
}

func (lcc *LocalCodeCache) setItem(key string, item any) error {
	val, err := json.Marshal(item)
	if err != nil {
		return err
	}

	err = lcc.cache.Set(key, val)
	if err != nil {
		return err
	}

	return nil
}

func (lcc *LocalCodeCache) getItem(key string) (codeItem, error) {
	val, err := lcc.cache.Get(key)
	if err == nil {
		var item codeItem
		err := json.Unmarshal(val, &item)
		if err != nil {
			return codeItem{}, err
		}
		return item, nil
	}

	return codeItem{}, err
}

func (lcc *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	lcc.mutex.Lock()
	defer lcc.mutex.Unlock()

	key := lcc.key(biz, phone)

	now := time.Now()
	// 先获取一次 查看之前有无缓存
	item, err := lcc.getItem(key)
	newer := codeItem{
		Code:   code,
		Cnt:    3,
		Expire: now.Add(lcc.expiration),
	}

	// 缓存不存在 可以发
	if err == bigcache.ErrEntryNotFound {
		err := lcc.setItem(key, newer)
		return err
	}

	// 缓存存在 根据时间判断是否重发
	if err == nil {
		// 不到一分钟 发送频繁
		if item.Expire.Sub(now) > time.Minute*9 {
			return ErrCodeSendTooMany
		}

		// 重发
		err := lcc.setItem(key, newer)
		return err
	}

	// 缓存异常
	return err
}

func (lcc *LocalCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	lcc.mutex.Lock()
	defer lcc.mutex.Unlock()

	key := lcc.key(biz, phone)
	item, err := lcc.getItem(key)
	// 缓存不存在
	if err == bigcache.ErrEntryNotFound {
		return false, ErrKeyNotExist
	}

	// 缓存异常
	if err != nil {
		return false, err
	}

	// 缓存存在
	// 可验证次数已耗尽
	if item.Cnt <= 0 {
		return false, ErrCodeVerifyTooMany
	}
	// 更新可验证次数
	item.Cnt--
	err = lcc.setItem(key, item)
	if err != nil {
		return false, err
	}

	// 验证成功后删除
	if item.Code == code {
		lcc.cache.Delete(key)
	}

	// 比较验证码
	return item.Code == code, nil
}

func (lcc *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

type codeItem struct {
	Code string
	// 可验证次数
	Cnt int
	// 过期时间
	Expire time.Time
}
