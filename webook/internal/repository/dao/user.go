package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrDuplicateEmail = errors.New("邮箱重复")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, user User) error
	FindUserByEmail(ctx context.Context, email string) (User, error)
	UpdateUserById(ctx context.Context, user User) error
	FindUserById(ctx context.Context, uid int64) (User, error)
	FindUserByPhone(ctx context.Context, phone string) (User, error)
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}

func (gud *GORMUserDAO) Insert(ctx context.Context, user User) error {
	timestamp := time.Now().UnixMilli()
	user.CreateTime, user.UpdateTime = timestamp, timestamp
	err := gud.db.WithContext(ctx).Create(&user).Error
	//
	if me, ok := err.(*mysql.MySQLError); ok {
		if me.Number == 1062 {
			return ErrDuplicateEmail
		}
	}
	return err
}

func (gud *GORMUserDAO) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := gud.db.WithContext(ctx).Where("email=?", email).First(&user).Error
	return user, err
}

func (gud *GORMUserDAO) UpdateUserById(ctx context.Context, user User) error {
	// 使用map会更新零值
	// 使用struct不会更新零值
	// https://gorm.io/docs/update.html
	return gud.db.WithContext(ctx).Model(&user).Where("id = ?", user.Id).
		Updates(map[string]any{
			"update_time": time.Now().UnixMilli(),
			"nickname":    user.Nickname,
			"birthday":    user.Birthday,
			"bio":         user.Bio,
		}).Error
}

func (gud *GORMUserDAO) FindUserById(ctx context.Context, uid int64) (User, error) {
	var user User
	err := gud.db.WithContext(ctx).Where("id = ?", uid).First(&user).Error
	return user, err
}

func (gud *GORMUserDAO) FindUserByPhone(ctx context.Context, phone string) (User, error) {
	var user User
	err := gud.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error
	return user, err
}

type User struct {
	Id         int64          `gorm:"primaryKey,autoIncrement"`
	Email      sql.NullString `gorm:"unique"`
	Password   string
	Nickname   string `gorm:"type=varchar(128)"`
	Birthday   int64
	Bio        string         `gorm:"type=varchar(4096)"`
	Phone      sql.NullString `gorm:"unique"`
	CreateTime int64
	UpdateTime int64
}

func (User) TableName() string {
	return "users"
}
