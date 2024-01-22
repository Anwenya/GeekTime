package repository

import (
	"context"
	"time"

	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrUserNotFound   = dao.ErrRecordNotFound
)

type UserRepository struct {
	userDao *dao.UserDAO
}

func NewUserRepository(userDao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		userDao: userDao,
	}
}

func (userRepository *UserRepository) Create(ctx context.Context, domainUser domain.User) error {
	return userRepository.userDao.Insert(ctx, dao.User{
		Email:    domainUser.Email,
		Password: domainUser.Password,
	})
}

func (userRepository *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	daoUser, err := userRepository.userDao.FindUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return userRepository.toDomain(daoUser), nil
}

func (userRepository *UserRepository) toDomain(daoUser dao.User) domain.User {
	return domain.User{
		Id:       daoUser.Id,
		Email:    daoUser.Email,
		Password: daoUser.Password,
		Nickname: daoUser.Nickname,
		Birthday: time.UnixMilli(daoUser.Birthday),
		Bio:      daoUser.Bio,
	}
}

func (userRepository *UserRepository) toEntity(domainUser domain.User) dao.User {
	return dao.User{
		Id:       domainUser.Id,
		Email:    domainUser.Email,
		Password: domainUser.Password,
		Nickname: domainUser.Nickname,
		Birthday: domainUser.Birthday.UnixMilli(),
		Bio:      domainUser.Bio,
	}
}

func (userRepository *UserRepository) UpdateNonZeroFields(ctx context.Context, domainUser domain.User) error {
	return userRepository.userDao.UpdateUserById(ctx, userRepository.toEntity(domainUser))
}

func (userRepository *UserRepository) FindById(ctx context.Context, uid int64) (domain.User, error) {
	daoUser, err := userRepository.userDao.FindUserById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}
	return userRepository.toDomain(daoUser), nil
}
