package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// Payload
type Payload struct {
	jwt.RegisteredClaims
	Tid       uuid.UUID `json:"tid"`
	Uid       int64     `json:"uid"`
	Username  string    `json:"username"`
	UserAgent string    `json:"user_agent"`
	IssuedAt  time.Time `json:"issue_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewPayload
func NewPayload(uid int64, username string, duration time.Duration, ua string) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	payload := &Payload{
		Tid:       tokenID,
		Uid:       uid,
		Username:  username,
		UserAgent: ua,
		IssuedAt:  now,
		ExpiresAt: now.Add(duration),
	}
	return payload, nil
}

// Validate
// 自定义校验函数 为了配合paseto
// 如果不使用jwt.RegisteredClaims中的过期时间就需要自己额外做判断
func (payload *Payload) Validate() error {
	if time.Now().After(payload.ExpiresAt) {
		return ErrExpiredToken
	}
	return nil
}
