package ioc

import (
	"github.com/Anwenya/GeekTime/webook/util"
	"github.com/golang-migrate/migrate/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

func InitDB(config *util.Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.DBUrlMySQL),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	if err != nil {
		log.Fatalf("数据库连接失败:%v", err)
	}

	log.Println("数据库连接成功")
	mysqlMigration(config)
	return db
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
