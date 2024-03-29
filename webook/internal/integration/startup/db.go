package startup

import (
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	db, err := gorm.Open(
		mysql.Open("root:root@tcp(192.168.2.128:13306)/webook"),
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
	mysqlMigration(l)
	return db
}

func mysqlMigration(l logger.LoggerV1) {
	migration, err := migrate.New(
		"file://../../db/migration",
		"mysql://root:root@tcp(192.168.2.128:13306)/webook",
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
