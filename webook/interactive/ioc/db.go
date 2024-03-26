package ioc

import (
	"github.com/Anwenya/GeekTime/webook/pkg/gormx/connpool"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

type SrcDB *gorm.DB
type DstDB *gorm.DB

func InitSrcDB(l logger.LoggerV1) SrcDB {
	return InitDB(l, "src")
}

func InitDstDB(l logger.LoggerV1) DstDB {
	return InitDB(l, "dst")
}

func InitDoubleWritePool(
	src SrcDB,
	dst DstDB,
	l logger.LoggerV1,
) *connpool.DoubleWritePool {
	return connpool.NewDoubleWritePool(src, dst, l)
}

func InitBizDB(p *connpool.DoubleWritePool) *gorm.DB {
	doubleWrite, err := gorm.Open(
		mysql.New(mysql.Config{
			Conn: p,
		}),
	)
	if err != nil {
		panic(any(err))
	}
	return doubleWrite
}

type dbConfig struct {
	Url                string `yaml:"url"`
	MigrationUrl       string `yaml:"migrationUrl"`
	MigrationSourceUrl string `yaml:"migrationSourceUrl"`
}

func InitDB(l logger.LoggerV1, key string) *gorm.DB {
	var cfg dbConfig
	err := viper.UnmarshalKey("db.mysql."+key, &cfg)

	db, err := gorm.Open(
		mysql.Open(cfg.Url),
		&gorm.Config{
			Logger: glogger.New(
				gormLoggerFunc(l.Debug),
				glogger.Config{
					// 慢查询
					SlowThreshold: 0,
					LogLevel:      glogger.Info,
				}),
		})
	if err != nil {
		l.Error("数据库连接失败")
		panic(any(err))
	}
	l.Info("数据库连接成功")

	mysqlMigration(l, cfg)
	return db
}

func mysqlMigration(l logger.LoggerV1, cfg dbConfig) {
	migration, err := migrate.New(
		cfg.MigrationSourceUrl,
		cfg.MigrationUrl,
	)
	if err != nil {
		l.Error("数据库迁移失败")
		panic(any(err))
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		l.Error("数据库迁移失败")
		panic(any(err))
	}

	l.Info("数据库迁移成功")
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (glf gormLoggerFunc) Printf(s string, i ...interface{}) {
	glf(s, logger.Field{Key: "args", Val: i})
}
