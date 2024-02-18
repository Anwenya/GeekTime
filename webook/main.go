package main

import (
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/repository/cache"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms/localsms"
	"github.com/Anwenya/GeekTime/webook/internal/web"
	"github.com/Anwenya/GeekTime/webook/internal/web/middleware"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx/middleware/ratelimit"
	"github.com/Anwenya/GeekTime/webook/token"
	"github.com/Anwenya/GeekTime/webook/util"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/redis/go-redis/v9"
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

	tokenMaker, err := token.NewPasetoMaker(config.TokenSecretKey)
	if err != nil {
		log.Fatalf("初始化tokenMaker失败:%v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisAddress,
	})

	db := initDB(config)
	server := initWebServer(config, tokenMaker, redisClient)

	initUserHandler(db, redisClient, server, config, tokenMaker)
	err = server.Run(config.HTTPServerAddress)
	if err != nil {
		log.Fatalf("启动失败:%v", err)
	}
	log.Printf("启动成功:%v", config.HTTPServerAddress)
}

func initUserHandler(
	db *gorm.DB,
	redisClient redis.Cmdable,
	server *gin.Engine,
	config *util.Config,
	tokenMaker token.Maker,
) {
	userDao := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(redisClient)
	userRepository := repository.NewUserRepository(userDao, userCache)
	userService := service.NewUserService(userRepository)

	codeService := initCodeService(redisClient)

	userHandler := web.NewUserHandler(userService, codeService, config, tokenMaker)
	userHandler.RegisterRoutes(server)
}

func initDB(config *util.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.DBUrlMySQL),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	if err != nil {
		log.Fatalf("数据库连接失败:%v", err)
	}

	mysqlMigration(config)
	log.Println("数据库连接成功")
	return db
}

func initCodeService(redisClient redis.Cmdable) *service.CodeService {
	cc := cache.NewCodeCache(redisClient)
	cr := repository.NewCodeRepository(cc)
	return service.NewCodeService(cr, initMemorySms())
}

func initMemorySms() sms.SMService {
	return localsms.NewService()
}

func mysqlMigration(config *util.Config) {
	migration, err := migrate.New(
		config.MigrationSourceUrl,
		config.MigrationDBUrl,
	)
	if err != nil {
		log.Fatalf("数据库迁移失败:%v", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("数据库迁移失败:%v", err)
	}

	log.Println("数据库迁移成功")
}

func initWebServer(config *util.Config, tokenMaker token.Maker, redisClient redis.Cmdable) *gin.Engine {
	server := gin.Default()
	cors := &middleware.CorsMiddlewareBuilder{}

	//rateLimitBuilder := ratelimit.NewSlideWindowBuilder(redisClient, time.Second, 1)
	rateLimitBuilder := ratelimit.NewTokenBucketBuilder(redisClient, 10, 1)
	session := &middleware.SessionMiddlewareBuilder{}
	login := &middleware.LoginMiddlewareBuilder{}
	server.Use(
		cors.Cors(config),
		rateLimitBuilder.Build(),
		session.Session(config),
		login.CheckLogin(config, tokenMaker),
	)

	return server
}
