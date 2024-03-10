package ioc

import (
	"github.com/Anwenya/GeekTime/webook/config"
	"github.com/Anwenya/GeekTime/webook/pkg/gormx"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	gprometheus "gorm.io/plugin/prometheus"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	db, err := gorm.Open(
		mysql.Open(config.Config.DB.MySQL.Url),
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

	// gorm提供的prometheus插件
	// 仅仅记录一些数据库连接的状态
	err = db.Use(
		gprometheus.New(
			gprometheus.Config{
				DBName:          "webook",
				RefreshInterval: 15,
				MetricsCollector: []gprometheus.MetricsCollector{
					&gprometheus.MySQL{
						VariableNames: []string{"thread_running"},
					},
				},
			},
		),
	)
	if err != nil {
		l.Error("数据库插件启动失败")
		panic(any(err))
	}

	// 自定义插件记录sql的执行耗时
	callback := gormx.NewCallbacks(
		prometheus.SummaryOpts{
			Namespace: "GeekTime",
			Subsystem: "webook",
			Name:      "_gorm_db",
			Help:      "统计GORM的数据库操作",
			ConstLabels: map[string]string{
				"instance_id": "my_instance",
			},
			Objectives: map[float64]float64{
				0.5:   0.01,
				0.75:  0.01,
				0.9:   0.01,
				0.99:  0.001,
				0.999: 0.0001,
			},
		},
	)

	err = db.Use(callback)
	if err != nil {
		l.Error("数据库插件启动失败")
		panic(any(err))
	}

	mysqlMigration(l)
	return db
}

func mysqlMigration(l logger.LoggerV1) {
	migration, err := migrate.New(
		config.Config.DB.MySQL.MigrationSourceUrl,
		config.Config.DB.MySQL.MigrationUrl,
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
