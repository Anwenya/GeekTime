package main

import (
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/Anwenya/GeekTime/webook/internal/web/middleware"
	"github.com/Anwenya/GeekTime/webook/util"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

func main() {
	var config, err = util.LoadConfig(".")
	if err != nil {
		log.Fatalf("配置文件加载失败:%v", err)
	}
	db := initDB(config)
	server := initWebServer(config)
	initUserHandler(db, server, config)
	err = server.Run(config.HTTPServerAddress)
	if err != nil {
		log.Fatalf("启动失败:%v", err)
	}
}

func initUserHandler(db *gorm.DB, server *gin.Engine, config *util.Config) {
	userDao := dao.NewUserDAO(db)
	userRepository := repository.NewUserRepository(userDao)
	userService := service.NewUserService(userRepository)
	userHandler := web.NewUserHandler(userService, config)
	userHandler.RegisterRoutes(server)
}

func initDB(config *util.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.DBUrlMySQL),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	if err != nil {
		log.Fatal(err)
	}

	mysqlMigration(config)
	return db
}

func mysqlMigration(config *util.Config) {
	migration, err := migrate.New(
		config.MigrationSourceUrl,
		config.MigrationDBUrl,
	)
	if err != nil {
		log.Printf("cannot create new migrate instance:%v", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("failed to run migrate up:%v", err)
	}

	log.Printf("db migration successfully")
}

func initWebServer(config *util.Config) *gin.Engine {
	server := gin.Default()
	cors := &middleware.CorsMiddlewareBuilder{}
	session := &middleware.SessionMiddlewareBuilder{}
	login := &middleware.LoginMiddlewareBuilder{}
	server.Use(
		cors.Cors(config),
		session.Session(config),
		login.CheckLogin(config, nil),
	)

	return server
}
