package cache

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/account/domain"
)

type AccountCache interface {
	SetUnique(ctx context.Context, cr domain.Credit) error
	GetUnique(ctx context.Context, cr domain.Credit) error
}
