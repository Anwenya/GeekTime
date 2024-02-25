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

type UserService interface {
	Signup(ctx context.Context, domainUser domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	UpdateNonSensitiveInfo(ctx context.Context, domainUser domain.User) error
	FindById(ctx context.Context, uid int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}

type userService struct {
	ur repository.UserRepository
}

func NewUserService(ur repository.UserRepository) UserService {
	return &userService{
		ur: ur,
	}
}

func (us *userService) Signup(ctx context.Context, domainUser domain.User) error {
	hash, err := util.HashPassword(domainUser.Password)
	if err != nil {
		return err
	}
	domainUser.Password = hash
	return us.ur.Create(ctx, domainUser)
}

func (us *userService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	domainUser, err := us.ur.FindByEmail(ctx, email)
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

func (us *userService) UpdateNonSensitiveInfo(ctx context.Context, domainUser domain.User) error {
	return us.ur.UpdateNonZeroFields(ctx, domainUser)
}

func (us *userService) FindById(ctx context.Context, uid int64) (domain.User, error) {
	return us.ur.FindById(ctx, uid)
}

func (us *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 大部分情况应该是已经注册过的用户
	u, err := us.ur.FindByPhone(ctx, phone)
	// 先确保用户不存在
	if err != repository.ErrUserNotFound {
		// 两种情况
		// err == nil 正常查找到了用户信息
		// err != nil 系统异常
		return u, err
	}

	// 用户没有找到
	// 创建用户
	err = us.ur.Create(ctx, domain.User{Phone: phone})

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
	return us.ur.FindByPhone(ctx, phone)
}

func (us *userService) FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	// 大部分情况应该是已经注册过的用户
	u, err := us.ur.FindByWechat(ctx, wechatInfo.OpenId)
	// 先确保用户不存在
	if err != repository.ErrUserNotFound {
		// 两种情况
		// err == nil 正常查找到了用户信息
		// err != nil 系统异常
		return u, err
	}

	// 用户没有找到
	// 创建用户
	err = us.ur.Create(ctx, domain.User{WechatInfo: wechatInfo})

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
	return us.ur.FindByWechat(ctx, wechatInfo.OpenId)
}
