// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/Anwenya/GeekTime/webook/internal/ioc"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	tokenHandler := token.NewRedisTokenHandler(cmdable)
	loggerV1 := ioc.InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, tokenHandler, loggerV1)
	db := ioc.InitDB(loggerV1)
	userDAO := dao.NewGORMUserDAO(db)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCachedCodeRepository(codeCache)
	smService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smService)
	userHandler := web.NewUserHandler(userService, codeService, tokenHandler)
	wechatService := ioc.InitWechatService(loggerV1)
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, tokenHandler)
	articleDAO := dao.NewGORMArticleDAO(db)
	articleCache := cache.NewRedisArticleCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(articleDAO, articleCache, userRepository, loggerV1)
	articleService := service.NewArticleService(articleRepository)
	interactiveDAO := dao.NewGORMInteractiveDAO(db)
	interactiveCache := cache.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	articleHandler := web.NewArticleHandler(loggerV1, articleService, interactiveService)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	return engine
}

// wire.go:

var interactiveServiceSet = wire.NewSet(dao.NewGORMInteractiveDAO, cache.NewRedisInteractiveCache, repository.NewCachedInteractiveRepository, service.NewInteractiveService)
