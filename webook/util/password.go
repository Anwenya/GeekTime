package util

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 对明文密码进行hash后返回
// 会校验密码长度不能超过72个字节
// hash后前22个字节是version和随机salt 后边是密码
func HashPassword(Password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword 校验是否匹配
// 先读元信息 比如salt
// 之后对明文进行相同的运算再比较值
func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
