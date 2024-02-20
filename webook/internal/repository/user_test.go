package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
	mockcache "github.com/Anwenya/GeekTime/webook/internal/repository/cache/mocks"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	mockdao "github.com/Anwenya/GeekTime/webook/internal/repository/dao/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	now := time.UnixMilli(nowMs)
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)

		ctx context.Context
		uid int64

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "查找成功，缓存未命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uid := int64(123456)
				ud := mockdao.NewMockUserDAO(ctrl)
				uc := mockcache.NewMockUserCache(ctrl)

				// 查找缓存
				uc.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)

				// 查找数据库
				ud.EXPECT().FindUserById(gomock.Any(), uid).Return(
					dao.User{
						Id: uid,
						Email: sql.NullString{
							String: "123456789@qq.com",
							Valid:  true,
						},
						Password: "123456",
						Birthday: 100,
						Bio:      "自我介绍",
						Phone: sql.NullString{
							String: "15112345678",
							Valid:  true,
						},
						CreateTime: nowMs,
						UpdateTime: 102,
					}, nil)

				// 回写缓存
				uc.EXPECT().Set(gomock.Any(), domain.User{
					Id:         uid,
					Email:      "123456789@qq.com",
					Password:   "123456",
					Birthday:   time.UnixMilli(100),
					Bio:        "自我介绍",
					Phone:      "15112345678",
					CreateTime: now,
				}).Return(nil)

				return ud, uc
			},
			uid: 123456,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:         123456,
				Email:      "123456789@qq.com",
				Password:   "123456",
				Birthday:   time.UnixMilli(100),
				Bio:        "自我介绍",
				Phone:      "15112345678",
				CreateTime: now,
			},
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uid := int64(123456)
				ud := mockdao.NewMockUserDAO(ctrl)
				uc := mockcache.NewMockUserCache(ctrl)

				// 查找缓存
				uc.EXPECT().Get(gomock.Any(), uid).Return(domain.User{
					Id:         uid,
					Email:      "123456789@qq.com",
					Password:   "123456",
					Birthday:   time.UnixMilli(100),
					Bio:        "自我介绍",
					Phone:      "15112345678",
					CreateTime: now,
				}, nil)
				return ud, uc
			},
			uid: 123456,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:         123456,
				Email:      "123456789@qq.com",
				Password:   "123456",
				Birthday:   time.UnixMilli(100),
				Bio:        "自我介绍",
				Phone:      "15112345678",
				CreateTime: now,
			},
		},
		{
			name: "未找到用户",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uid := int64(123456)
				ud := mockdao.NewMockUserDAO(ctrl)
				uc := mockcache.NewMockUserCache(ctrl)

				// 查找缓存
				uc.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)

				// 查找数据库
				ud.EXPECT().FindUserById(gomock.Any(), uid).Return(
					dao.User{}, dao.ErrRecordNotFound)

				return ud, uc
			},
			uid:      123456,
			ctx:      context.Background(),
			wantUser: domain.User{},
			wantErr:  dao.ErrRecordNotFound,
		},
		{
			name: "回写缓存失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uid := int64(123456)
				ud := mockdao.NewMockUserDAO(ctrl)
				uc := mockcache.NewMockUserCache(ctrl)

				// 查找缓存
				uc.EXPECT().Get(gomock.Any(), uid).Return(domain.User{}, cache.ErrKeyNotExist)

				// 查找数据库
				ud.EXPECT().FindUserById(gomock.Any(), uid).Return(
					dao.User{
						Id: uid,
						Email: sql.NullString{
							String: "123456789@qq.com",
							Valid:  true,
						},
						Password: "123456",
						Birthday: 100,
						Bio:      "自我介绍",
						Phone: sql.NullString{
							String: "15112345678",
							Valid:  true,
						},
						CreateTime: nowMs,
						UpdateTime: 102,
					}, nil)

				// 回写缓存
				uc.EXPECT().Set(gomock.Any(), domain.User{
					Id:         uid,
					Email:      "123456789@qq.com",
					Password:   "123456",
					Birthday:   time.UnixMilli(100),
					Bio:        "自我介绍",
					Phone:      "15112345678",
					CreateTime: now,
				}).Return(errors.New("redis异常"))

				return ud, uc
			},
			uid: 123456,
			ctx: context.Background(),
			wantUser: domain.User{
				Id:         123456,
				Email:      "123456789@qq.com",
				Password:   "123456",
				Birthday:   time.UnixMilli(100),
				Bio:        "自我介绍",
				Phone:      "15112345678",
				CreateTime: now,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ud, uc := tc.mock(ctrl)
			ur := NewCachedUserRepository(ud, uc)
			user, err := ur.FindById(tc.ctx, tc.uid)
			assert.Equal(t, tc.wantUser, user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
