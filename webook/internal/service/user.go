package service

import (
	"context"
	"errors"
	"github.com/Anwenya/GeekTime/webook/util"

	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateEmail
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
