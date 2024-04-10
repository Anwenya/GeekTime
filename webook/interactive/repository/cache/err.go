package cache

import (
	"errors"
	"github.com/redis/go-redis/v9"
)

var (
	ErrKeyNotExist   = redis.Nil
	ErrRankingUpdate = errors.New("指定的元素不存在")
)
