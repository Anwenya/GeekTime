package main

import (
	"github.com/Anwenya/GeekTime/webook/interactive/config"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	initLogger()
	initPrometheus()

	app := InitApp()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			zap.L().Panic("启动失败", zap.Error(err))
		}
	}

	server := app.server

	err := server.Serve()
	if err != nil {
		zap.L().Panic("启动失败", zap.Error(err))
	}
	zap.L().Info(
		"启动成功",
		zap.String(
			"address",
			config.Config.Grpc.Address,
		),
	)

}

func initPrometheus() {
	go func() {
		// 专门给 prometheus 用的端口
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8082", nil)
		if err != nil {
			zap.L().Panic("启动失败", zap.Error(err))
		}
	}()
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
