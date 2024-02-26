package main

import (
	"github.com/Anwenya/GeekTime/webook/config"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

func main() {
	initLogger()
	server := InitWebServer()
	err := server.Run(config.Config.App.HttpServerAddress)
	if err != nil {
		zap.L().Panic("启动失败", zap.Error(err))
	}
	zap.L().Info("启动成功",
		zap.Any(
			"address",
			config.Config.App.HttpServerAddress,
		),
	)
}

// 先于其他组件初始化
// 记录启动过程中的日志
func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(any(err))
	}
	zap.ReplaceGlobals(logger)
}
