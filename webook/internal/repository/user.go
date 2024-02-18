package repository

import (
	"context"
	"database/sql"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
	"log"
	"time"

	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository struct {
	userDao *dao.UserDAO
	cache   *cache.UserCache
}

func NewUserRepository(userDao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		userDao: userDao,
		cache:   cache,
	}
}

func (userRepository *UserRepository) Create(ctx context.Context, domainUser domain.User) error {
	return userRepository.userDao.Insert(ctx, userRepository.toEntity(domainUser))
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
		Email:    daoUser.Email.String,
		Phone:    daoUser.Phone.String,
		Password: daoUser.Password,
		Nickname: daoUser.Nickname,
		Birthday: time.UnixMilli(daoUser.Birthday),
		Bio:      daoUser.Bio,
	}
}

func (userRepository *UserRepository) toEntity(domainUser domain.User) dao.User {
	return dao.User{
		Id: domainUser.Id,
		Email: sql.NullString{
			String: domainUser.Email,
			Valid:  domainUser.Email != "",
		},
		Phone: sql.NullString{
			String: domainUser.Phone,
			Valid:  domainUser.Phone != "",
		},
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
	// 先尝试查询缓存
	du, err := userRepository.cache.Get(ctx, uid)
	// 只要err为nil就返回 说明查到了缓存
	if err == nil {
		return du, nil
	}

	// 1.key不存在 说明redis是正常的
	// 2.访问redis时网络或者redis本身有问题
	// 这里可以根据实际情况判断异常是否是ErrKeyNotExist
	// 如果是则 正常查询数据库
	// 不是则 返回零值 避免数据库压力过大
	//if err != cache.ErrKeyNotExist{
	//	return domain.User{}, err
	//}

	// err 不为nil则查询数据库
	daoUser, err := userRepository.userDao.FindUserById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}

	du = userRepository.toDomain(daoUser)
	err = userRepository.cache.Set(ctx, du)
	// 设置缓存失败只记录日志
	if err != nil {
		log.Printf("设置缓存失败:%v", err)
	}

	return du, nil
}

func (userRepository *UserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := userRepository.userDao.FindUserByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return userRepository.toDomain(u), nil
}
