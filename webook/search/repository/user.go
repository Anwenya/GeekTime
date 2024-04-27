package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/search/domain"
	"github.com/Anwenya/GeekTime/webook/search/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type userRepository struct {
	dao dao.UserDAO
}

func NewUserRepository(dao dao.UserDAO) UserRepository {
	return &userRepository{dao: dao}
}

func (u *userRepository) InputUser(ctx context.Context, msg domain.User) error {
	return u.dao.InputUser(
		ctx,
		dao.User{
			Id:       msg.Id,
			Email:    msg.Email,
			Nickname: msg.Nickname,
			Phone:    msg.Phone,
		},
	)
}

func (u *userRepository) SearchUser(ctx context.Context, keywords []string) ([]domain.User, error) {
	users, err := u.dao.Search(ctx, keywords)
	if err != nil {
		return nil, err
	}

	return slice.Map[dao.User](users,
		func(idx int, src dao.User) domain.User {
			return domain.User{
				Id:       src.Id,
				Email:    src.Email,
				Nickname: src.Nickname,
				Phone:    src.Phone,
			}
		},
	), nil
}
