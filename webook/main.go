package main

import (
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/Anwenya/GeekTime/webook/internal/web/middleware"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := initDB()
	server := initWebServer()
	initUserHandler(db, server)
	server.Run(":8080")
}

func initUserHandler(db *gorm.DB, server *gin.Engine) {
	userDao := dao.NewUserDAO(db)
	userRepository := repository.NewUserRepository(userDao)
	userService := service.NewUserService(userRepository)
	userHandler := web.NewUserHandler(userService)
	userHandler.RegisterRoutes(server)
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(""))
	if err != nil {
		panic(err)
	}
	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	cors := &middleware.CorsMiddlewareBuilder{}
	session := sessions.Sessions("ssid", cookie.NewStore([]byte("secret")))
	login := &middleware.LoginMiddlewareBuilder{}
	server.Use(cors.Cors(), session, login.CheckLogin())

	return server
}
