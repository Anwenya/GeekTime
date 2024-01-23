package main

import (
	"log"

	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/Anwenya/GeekTime/webook/internal/web/middleware"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	dbURL := "root:root@tcp(localhost:13306)/webook"
	db, err := gorm.Open(mysql.Open(dbURL),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	if err != nil {
		panic(err)
	}
	// 迁移文件
	migrationURL := "file://db/migration"
	migrationDBURL := "mysql://" + dbURL
	mysqlMigration(migrationURL, migrationDBURL)
	return db
}

func mysqlMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Printf("cannot create new migrate instance:%v", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("failed to run migrate up:%v", err)
	}

	log.Printf("db migration successfully")
}

func initWebServer() *gin.Engine {
	server := gin.Default()
	cors := &middleware.CorsMiddlewareBuilder{}
	session := &middleware.SessionMiddlewareBuilder{}
	login := &middleware.LoginMiddlewareBuilder{}
	server.Use(cors.Cors(), session.Session(), login.CheckLogin())

	return server
}
