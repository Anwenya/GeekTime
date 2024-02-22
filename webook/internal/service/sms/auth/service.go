package auth

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	ss  sms.SMService
	key []byte
}

func (as *AuthService) Send(ctx context.Context, tplToken string, args []string, numbers ...string) error {
	var claims SMSClaims
	_, err := jwt.ParseWithClaims(tplToken, &claims, func(t *jwt.Token) (interface{}, error) {
		return as.key, nil
	})
	if err != nil {
		return err
	}
	return as.ss.Send(ctx, claims.Tpl, args, numbers...)
}

type SMSClaims struct {
	jwt.RegisteredClaims
	Tpl string
}
