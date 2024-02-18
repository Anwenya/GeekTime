package service

import (
	"context"
	"errors"
	"github.com/Anwenya/GeekTime/webook/util"

	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户不存在或密码不匹配")
)

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(userRepository *repository.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (userService *UserService) Signup(ctx context.Context, domainUser domain.User) error {
	hash, err := util.HashPassword(domainUser.Password)
	if err != nil {
		return err
	}
	domainUser.Password = hash
	return userService.userRepository.Create(ctx, domainUser)
}

func (userService *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	domainUser, err := userService.userRepository.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	err = util.CheckPassword(password, domainUser.Password)
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return domainUser, nil
}

func (userService *UserService) UpdateNonSensitiveInfo(ctx context.Context, domainUser domain.User) error {
	return userService.userRepository.UpdateNonZeroFields(ctx, domainUser)
}

func (userService *UserService) FindById(ctx context.Context, uid int64) (domain.User, error) {
	return userService.userRepository.FindById(ctx, uid)
}

func (userService *UserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 大部分情况应该是已经注册过的用户
	u, err := userService.userRepository.FindByPhone(ctx, phone)
	// 先确保用户不存在
	if err != repository.ErrUserNotFound {
		// 两种情况
		// err == nil 正常查找到了用户信息
		// err != nil 系统异常
		return u, err
	}

	// 用户没有找到
	// 创建用户
	err = userService.userRepository.Create(ctx, domain.User{Phone: phone})

	// 有两种可能
	// 1.err是唯一索引冲突
	// 2.err是其他系统错误
	if err != nil && err != repository.ErrDuplicateUser {
		return domain.User{}, err
	}

	// 有两种可能
	// 1.err == nil 一切顺利
	// 2.索引冲突 但也代表用户存在 再次查询
	// 查询可能存在主从延迟 可以强制走主库
	return userService.userRepository.FindByPhone(ctx, phone)
}
