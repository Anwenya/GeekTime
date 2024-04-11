package startup

import (
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	type config struct {
		Url                string `yaml:"url"`
		MigrationUrl       string `yaml:"migrationUrl"`
		MigrationSourceUrl string `yaml:"migrationSourceUrl"`
	}

	cfg := config{
		Url:                "root:root@tcp(192.168.2.130:13306)/webook",
		MigrationUrl:       "mysql://root:root@tcp(192.168.2.130:13306)/webook",
		MigrationSourceUrl: "file://../../db/migration",
	}

	db, err := gorm.Open(
		mysql.Open(cfg.Url),
		&gorm.Config{},
	)
	if err != nil {
		l.Error("数据库连接失败", logger.Error(err))
		panic(any(err))
	}
	l.Info("数据库连接成功")

	// 迁移
	migration, err := migrate.New(
		cfg.MigrationSourceUrl,
		cfg.MigrationUrl,
	)
	if err != nil {
		l.Error("数据库迁移失败", logger.Error(err))
		panic(any(err))
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		l.Error("数据库迁移失败", logger.Error(err))
		panic(any(err))
	}

	l.Info("数据库迁移成功")
	return db
}
