package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrDuplicateEmail = errors.New("邮箱重复")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (userDao *UserDAO) Insert(ctx context.Context, user User) error {
	timestamp := time.Now().UnixMilli()
	user.CreateTime, user.UpdateTime = timestamp, timestamp
	err := userDao.db.WithContext(ctx).Create(&user).Error
	//
	if me, ok := err.(*mysql.MySQLError); ok {
		if me.Number == 1062 {
			return ErrDuplicateEmail
		}
	}
	return err
}

func (userDao *UserDAO) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := userDao.db.WithContext(ctx).Where("email=?", email).First(&user).Error
	return user, err
}

func (userDao *UserDAO) UpdateUserById(ctx context.Context, user User) error {
	// 使用map会更新零值
	// 使用struct不会更新零值
	// https://gorm.io/docs/update.html
	return userDao.db.WithContext(ctx).Model(&user).Where("id = ?", user.Id).
		Updates(map[string]any{
			"update_time": time.Now().UnixMilli(),
			"nickname":    user.Nickname,
			"birthday":    user.Birthday,
			"bio":         user.Bio,
		}).Error
}

func (userDao *UserDAO) FindUserById(ctx context.Context, uid int64) (User, error) {
	var user User
	err := userDao.db.WithContext(ctx).Where("id = ?", uid).First(&user).Error
	return user, err
}

type User struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Email      string `gorm:"unique"`
	Password   string
	Nickname   string `gorm:"type=varchar(128)"`
	Birthday   int64
	Bio        string `gorm:"type=varchar(4096)"`
	CreateTime int64
	UpdateTime int64
}

func (User) TableName() string {
	return "users"
}
