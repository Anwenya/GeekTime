package web

import (
	"github.com/Anwenya/GeekTime/webook/util"
	"log"
	"net/http"
	"time"

	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	emailRegexPattern    = `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

type UserHandler struct {
	config         *util.Config
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	userService    *service.UserService
}

func NewUserHandler(userService *service.UserService, config *util.Config) *UserHandler {
	return &UserHandler{
		config:         config,
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		userService:    userService,
	}
}

func (userHandler *UserHandler) RegisterRoutes(server *gin.Engine) {
	userGroup := server.Group("/users")
	userGroup.POST("/signup", userHandler.SignUp)
	userGroup.POST("/login", userHandler.Login)
	userGroup.POST("/edit", userHandler.Edit)
	userGroup.GET("/profile", userHandler.Profile)
}

func (userHandler *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	isEmail, err := userHandler.emailRexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK, "非法邮箱格式")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入密码不对")
		return
	}

	isPassword, err := userHandler.passwordRexExp.MatchString(req.Password)
	if err != nil {
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
		ctx.String(http.StatusOK, "该邮箱已被注册")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (userHandler *UserHandler) Login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	domainUser, err := userHandler.userService.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("uid", domainUser.Id)
		sess.Options(sessions.Options{
			MaxAge: int(userHandler.config.SessionDuration),
		})
		err = sess.Save()
		if err != nil {
			log.Printf("%v", err)
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (userHandler *UserHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		Bio      string `json:"bio"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	sess := sessions.Default(ctx)
	uid := sess.Get("uid").(int64)

	if len(req.Bio) > 4096 || len(req.Nickname) > 128 {
		ctx.String(http.StatusOK, "非法请求")
		return
	}

	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "非法生日格式")
		return
	}
	err = userHandler.userService.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       uid,
		Nickname: req.Nickname,
		Birthday: birthday,
		Bio:      req.Bio,
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.String(http.StatusOK, "更新成功")
}

func (userHandler *UserHandler) Profile(ctx *gin.Context) {

	sess := sessions.Default(ctx)
	uid := sess.Get("uid").(int64)
	domainUser, err := userHandler.userService.FindById(ctx, uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	type User struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Bio      string `json:"bio"`
		Birthday string `json:"birthday"`
	}
	ctx.JSON(http.StatusOK, User{
		Nickname: domainUser.Nickname,
		Email:    domainUser.Email,
		Bio:      domainUser.Bio,
		Birthday: domainUser.Birthday.Format(time.DateOnly),
	})
}
