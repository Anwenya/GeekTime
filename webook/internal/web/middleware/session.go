package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type SessionMiddlewareBuilder struct {
}

func (sessionMiddlewareBuilder *SessionMiddlewareBuilder) Session() gin.HandlerFunc {
	// 密钥成对定义以允许密钥轮换，但常见的情况是设置单个身份验证密钥和可选的加密密钥。
	// 密钥对中的第一个密钥用于身份验证，第二个密钥用于加密。加密密钥可以设置为nil或在最后一对中省略，但所有对都需要认证密钥。
	// 建议使用32字节或64字节的认证密钥。加密密钥必须为16、24或32字节，以选择AES-128、AES-192或AES-256的加密方式。

	return sessions.Sessions("ssid", cookie.NewStore([]byte("secret")))
}
