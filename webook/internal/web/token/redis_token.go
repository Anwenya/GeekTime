package token

import (
	"errors"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type RedisTokenHandler struct {
	client        redis.Cmdable
	signingMethod jwt.SigningMethod
	rcExpiration  time.Duration
}

func NewRedisTokenHandler(client redis.Cmdable) TokenHandler {
	return &RedisTokenHandler{
		client:        client,
		signingMethod: jwt.SigningMethodHS512,
		rcExpiration:  time.Hour * 24 * 7,
	}
}

// ClearToken 清除token 意味着退出登录
func (rth *RedisTokenHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("user").(UserClaims)

	return rth.client.Set(
		ctx,
		fmt.Sprintf("users:ssid:%s", uc.Ssid),
		"",
		rth.rcExpiration).Err()
}

// ExtractToken 从请求头中提取出token
func (rth *RedisTokenHandler) ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		return authCode
	}
	segs := strings.Split(authCode, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}

// SetLoginToken 登录成功设置长短token
func (rth *RedisTokenHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := rth.setRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return rth.SetToken(ctx, uid, ssid)
}

// SetToken 设置短token
func (rth *RedisTokenHandler) SetToken(ctx *gin.Context, uid int64, ssid string) error {
	uc := UserClaims{
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(rth.signingMethod, uc)
	tokenStr, err := token.SignedString(config.Config.SecretKey.Token)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

// 设置长token
func (rth *RedisTokenHandler) setRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(rth.rcExpiration)),
		},
	}
	token := jwt.NewWithClaims(rth.signingMethod, rc)
	tokenStr, err := token.SignedString(config.Config.SecretKey.Token)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

// CheckSession 检查redis中是否存在该token对应的ssid
// 存在则证明该token已被设置为无效
func (rth *RedisTokenHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := rth.client.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err != nil {
		return err
	}
	// 不存在是0
	if cnt > 0 {
		return errors.New("token无效")
	}
	return nil
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Ssid      string
	UserAgent string
}
