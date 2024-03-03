package web

import (
	"fmt"
	"github.com/Anwenya/GeekTime/webook/config"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/service/oauth2/wechat"
	itoken "github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
)

type OAuth2WechatHandler struct {
	itoken.TokenHandler
	s               wechat.Service
	us              service.UserService
	stateCookieName string
}

func NewOAuth2WechatHandler(
	s wechat.Service,
	us service.UserService,
	th itoken.TokenHandler,
) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		s:               s,
		us:              us,
		stateCookieName: "jwt-state",
		TokenHandler:    th,
	}
}

func (oawh *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	authGroup := server.Group("/oauth2/wechat")
	authGroup.GET("/authurl", oawh.Auth2URL)
	authGroup.Any("/callback", oawh.Callback)
}

func (oawh *OAuth2WechatHandler) Auth2URL(ctx *gin.Context) {
	state := uuid.New()
	val, err := oawh.s.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "构造跳转URL失败",
			Code: 5,
		})
	}
	err = oawh.setStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "服务器异常",
			Code: 5,
		})
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: val,
	})
}

func (oawh *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	err := oawh.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "非法请求",
			Code: 4,
		})
		return
	}

	code := ctx.Query("code")
	wechatInfo, err := oawh.s.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "授权码有误",
			Code: 4,
		})
		return
	}

	user, err := oawh.us.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		})
		return
	}

	err = oawh.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
	return
}

func (oawh *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	ck, err := ctx.Cookie(oawh.stateCookieName)
	if err != nil {
		return fmt.Errorf("无法获得 cookie %w", err)
	}

	var sc StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return config.Config.SecretKey.Token, nil
	})

	if err != nil {
		return fmt.Errorf("解析 token 失败 %w", err)
	}

	// 非法操作
	if state != sc.State {
		return fmt.Errorf("state 不匹配")
	}
	return nil
}

func (oawh *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	claims := StateClaims{
		State: state,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES512, claims)
	tokenStr, err := token.SignedString(config.Config.SecretKey.Token)
	if err != nil {
		return err
	}
	ctx.SetCookie(oawh.stateCookieName, tokenStr,
		600, "/oauth2/wechat/callback",
		"", false, true)
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}
