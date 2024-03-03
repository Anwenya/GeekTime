package web

import (
	"github.com/Anwenya/GeekTime/webook/config"
	itoken "github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx"
	"github.com/golang-jwt/jwt/v5"
	"log"
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
	userGroup.POST("/signup", userHandler.SignUp)
	userGroup.POST("/login", userHandler.LoginWithToken)
	userGroup.POST("/logout", userHandler.LogoutWithToken)
	userGroup.POST("/edit", userHandler.Edit)
	userGroup.GET("/profile", userHandler.Profile)
	userGroup.POST("/login_sms/code/send", userHandler.LoginSendSMSCode)
	userGroup.POST("/login_sms", userHandler.LoginWithSMS)
	userGroup.GET("/refresh-token", userHandler.RefreshToken)
}

func (userHandler *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		log.Printf("注册失败:%v", err)
		return
	}

	isEmail, err := userHandler.emailRexExp.MatchString(req.Email)
	if err != nil {
		log.Printf("注册失败:%v", err)
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不一致")
		return
	}

	isPassword, err := userHandler.passwordRexExp.MatchString(req.Password)
	if err != nil {
		log.Printf("注册失败:%v", err)
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	}

	err = userHandler.userService.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		log.Printf("注册失败:%v", err)
		ctx.String(http.StatusOK, "该邮箱已被注册")
	default:
		log.Printf("注册失败:%v", err)
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (userHandler *UserHandler) LoginWithToken(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		log.Printf("token登录失败:%v", err)
		return
	}
	domainUser, err := userHandler.userService.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		err := userHandler.SetLoginToken(ctx, domainUser.Id)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		log.Printf("token登录失败:%v", err)
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		log.Printf("token登录失败:%v", err)
		ctx.String(http.StatusOK, "系统错误")
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

func (userHandler *UserHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		Bio      string `json:"bio"`
	}

	uc, ok := ctx.MustGet("user").(itoken.UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		log.Printf("编辑用户信息失败:%v", err)
		return
	}

	if len(req.Bio) > 4096 || len(req.Nickname) > 128 {
		log.Printf("编辑用户信息失败:%v", "非法格式")
		ctx.String(http.StatusOK, "非法请求")
		return
	}

	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		log.Printf("编辑用户信息失败:%v", err)
		ctx.String(http.StatusOK, "非法生日格式")
		return
	}
	err = userHandler.userService.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       uc.Uid,
		Nickname: req.Nickname,
		Birthday: birthday,
		Bio:      req.Bio,
	})
	if err != nil {
		log.Printf("编辑用户信息失败:%v", err)
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.String(http.StatusOK, "更新成功")
}

func (userHandler *UserHandler) Profile(ctx *gin.Context) {
	uc, ok := ctx.MustGet("user").(itoken.UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	domainUser, err := userHandler.userService.FindById(ctx, uc.Uid)
	if err != nil {
		log.Printf("请求用户信息失败:%v", err)
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	type User struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Bio      string `json:"bio"`
		Birthday string `json:"birthday"`
	}
	ctx.JSON(http.StatusOK, User{
		Nickname: domainUser.Nickname,
		Email:    domainUser.Email,
		Phone:    domainUser.Phone,
		Bio:      domainUser.Bio,
		Birthday: domainUser.Birthday.Format(time.DateOnly),
	})
}

func (userHandler *UserHandler) LoginSendSMSCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "请输入手机号码",
		})
		return
	}
	err := userHandler.codeService.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		log.Printf("短信发送异常:%v", err)
	}
}

func (userHandler *UserHandler) LoginWithSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := userHandler.codeService.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "验证码错误",
		})
		return
	}
	domainUser, err := userHandler.userService.FindOrCreate(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	err = userHandler.SetLoginToken(ctx, domainUser.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "登录成功",
	})
}
