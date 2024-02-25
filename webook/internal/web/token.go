package web

import (
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/token"
	"github.com/Anwenya/GeekTime/webook/util"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type tokenHandler struct {
}

func (th *tokenHandler) setToken(ctx *gin.Context, user *domain.User) {
	tokenString, _, err := token.TkMaker.CreateToken(
		user.Id,
		user.Email,
		util.Config.AccessTokenDuration,
		ctx.GetHeader("User-Agent"),
	)
	if err != nil {
		log.Printf("创建token失败:%v", err)
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.Header(util.Config.TokenKey, tokenString)
}
