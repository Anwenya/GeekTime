package service

import (
	"context"
	"errors"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	mockrepository "github.com/Anwenya/GeekTime/webook/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestUserService_Login(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		ctx      context.Context
		email    string
		password string

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登陆成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				ur := mockrepository.NewMockUserRepository(ctrl)
				ur.EXPECT().
					FindByEmail(gomock.Any(), "123456789@qq.com").
					Return(
						domain.User{
							Email:    "123456789@qq.com",
							Password: "$2a$10$Di0YpG8xk4zNXAPtqtyEnOwLIhK1r8vU/Lt8F5QSDTis1aLM7Ulia",
							Phone:    "15112345678",
						}, nil)
				return ur
			},
			email:    "123456789@qq.com",
			password: "123456789#&a",
			wantUser: domain.User{
				Email:    "123456789@qq.com",
				Password: "$2a$10$Di0YpG8xk4zNXAPtqtyEnOwLIhK1r8vU/Lt8F5QSDTis1aLM7Ulia",
				Phone:    "15112345678",
			},
		},
		{
			name: "用户未找到",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				ur := mockrepository.NewMockUserRepository(ctrl)
				ur.EXPECT().
					FindByEmail(gomock.Any(), "123456789@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return ur
			},
			email:    "123456789@qq.com",
			password: "123456789#&a",
			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				ur := mockrepository.NewMockUserRepository(ctrl)
				ur.EXPECT().
					FindByEmail(gomock.Any(), "123456789@qq.com").
					Return(domain.User{}, errors.New("异常"))
				return ur
			},
			email:    "123456789@qq.com",
			password: "123456789#&a",
			wantErr:  errors.New("异常"),
		},
		{
			name: "密码不对",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				ur := mockrepository.NewMockUserRepository(ctrl)
				ur.EXPECT().
					FindByEmail(gomock.Any(), "123456789@qq.com").
					Return(
						domain.User{
							Email:    "123456789@qq.com",
							Password: "$2a$10$Di0YpG8xk4zNXAPtqtyEnOwLIhK1r8vU/Lt8F5QSDTis1aLM7Ulia",
							Phone:    "15112345678",
						}, nil)
				return ur
			},
			email:    "123456789@qq.com",
			password: "123456789#&",
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ur := tc.mock(ctrl)
			us := NewUserService(ur)
			user, err := us.Login(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
