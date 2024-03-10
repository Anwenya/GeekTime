package web

import (
	"github.com/Anwenya/GeekTime/webook/config"
	itoken "github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx/decorator"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"

	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
)

const (
	emailRegexPattern    = `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	bizLogin             = "login"
)

type UserHandler struct {
	itoken.TokenHandler
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	userService    service.UserService
	codeService    service.CodeService
}

func NewUserHandler(
	userService service.UserService,
	codeService service.CodeService,
	th itoken.TokenHandler,
) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		userService:    userService,
		codeService:    codeService,
		TokenHandler:   th,
	}
}

func (userHandler *UserHandler) RegisterRoutes(server *gin.Engine) {
	userGroup := server.Group("/users")
	userGroup.POST("/signup", decorator.WrapBody[SignUpReq](userHandler.SignUp))
	userGroup.POST("/login", decorator.WrapBody[LoginTokenReq](userHandler.LoginWithToken))
	userGroup.POST("/logout", userHandler.LogoutWithToken)
	userGroup.POST("/edit", decorator.WrapBodyAndClaims[UserEditReq, itoken.UserClaims](userHandler.Edit))
	userGroup.GET("/profile", decorator.WrapClaims[itoken.UserClaims](userHandler.Profile))
	userGroup.POST("/login_sms/code/send", decorator.WrapBody[SendSMSCodeReq](userHandler.LoginSendSMSCode))
	userGroup.POST("/login_sms", decorator.WrapBody[LoginSMSReq](userHandler.LoginWithSMS))
	userGroup.GET("/refresh-token", userHandler.RefreshToken)
}

func (userHandler *UserHandler) SignUp(
	ctx *gin.Context,
	req SignUpReq,
) (ginx.Result, error) {

	isEmail, err := userHandler.emailRexExp.MatchString(req.Email)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	if !isEmail {
		return ginx.Result{Code: 5, Msg: "非法邮箱格式"}, err
	}

	if req.Password != req.ConfirmPassword {
		return ginx.Result{Code: 5, Msg: "两次输入密码不一致"}, err
	}

	isPassword, err := userHandler.passwordRexExp.MatchString(req.Password)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	if !isPassword {
		return ginx.Result{
			Code: 5,
			Msg:  "密码必须包含字母、数字、特殊字符，并且不少于八位",
		}, err
	}

	err = userHandler.userService.Signup(
		ctx,
		domain.User{
			Email:    req.Email,
			Password: req.Password,
		},
	)
	switch err {
	case nil:
		return ginx.Result{Msg: "OK"}, nil
	case service.ErrDuplicateEmail:
		return ginx.Result{Code: 5, Msg: "该邮箱已被注册"}, err
	default:
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
}

func (userHandler *UserHandler) LoginWithToken(
	ctx *gin.Context,
	req LoginTokenReq,
) (ginx.Result, error) {
	domainUser, err := userHandler.userService.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		err := userHandler.SetLoginToken(ctx, domainUser.Id)
		if err != nil {
			return ginx.Result{Msg: "系统错误"}, err
		}
		return ginx.Result{
			Msg: "OK",
		}, nil
	case service.ErrInvalidUserOrPassword:
		return ginx.Result{Msg: "用户名或者密码错误"}, nil
	default:
		return ginx.Result{Msg: "系统错误"}, err
	}
}

func (userHandler *UserHandler) LogoutWithToken(ctx *gin.Context) {
	err := userHandler.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{Msg: "退出登录成功"})
}

func (userHandler *UserHandler) RefreshToken(ctx *gin.Context) {
	tokenStr := userHandler.ExtractToken(ctx)
	var rc itoken.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return config.Config.SecretKey.Token, nil
	})
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if token == nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = userHandler.CheckSession(ctx, rc.Ssid)
	if err != nil {
		// token 无效或者 redis 异常
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// 短token
	err = userHandler.SetToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
}

func (userHandler *UserHandler) Edit(
	ctx *gin.Context,
	req UserEditReq,
	uc itoken.UserClaims,
) (ginx.Result, error) {
	if len(req.Bio) > 4096 || len(req.Nickname) > 128 {
		return ginx.Result{
			Code: 4, Msg: "非法参数",
		}, nil
	}

	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		return ginx.Result{
			Code: 4,
			Msg:  "生日格式不对",
		}, err
	}
	err = userHandler.userService.UpdateNonSensitiveInfo(
		ctx,
		domain.User{
			Id:       uc.Uid,
			Nickname: req.Nickname,
			Birthday: birthday,
			Bio:      req.Bio,
		},
	)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (userHandler *UserHandler) Profile(
	ctx *gin.Context,
	uc itoken.UserClaims,
) (ginx.Result, error) {
	domainUser, err := userHandler.userService.FindById(ctx, uc.Uid)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	type User struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Bio      string `json:"bio"`
		Birthday string `json:"birthday"`
	}

	return ginx.Result{
		Data: User{
			Nickname: domainUser.Nickname,
			Email:    domainUser.Email,
			Phone:    domainUser.Phone,
			Bio:      domainUser.Bio,
			Birthday: domainUser.Birthday.Format(time.DateOnly),
		},
	}, nil
}

func (userHandler *UserHandler) LoginSendSMSCode(
	ctx *gin.Context,
	req SendSMSCodeReq,
) (ginx.Result, error) {
	if req.Phone == "" {
		return ginx.Result{
			Code: 4,
			Msg:  "请输入手机号码",
		}, nil
	}
	err := userHandler.codeService.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		return ginx.Result{Msg: "发送成功"}, nil
	case service.ErrCodeSendTooMany:
		return ginx.Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		}, nil
	default:
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
}

func (userHandler *UserHandler) LoginWithSMS(
	ctx *gin.Context,
	req LoginSMSReq,
) (ginx.Result, error) {
	ok, err := userHandler.codeService.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		}, err
	}
	if !ok {
		return ginx.Result{
			Code: 4,
			Msg:  "验证码错误",
		}, nil
	}
	domainUser, err := userHandler.userService.FindOrCreate(ctx, req.Phone)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	err = userHandler.SetLoginToken(ctx, domainUser.Id)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{Msg: "登录成功"}, nil
}
