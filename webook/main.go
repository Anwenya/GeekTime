package main

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/config"
	"github.com/Anwenya/GeekTime/webook/internal/ioc"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	initLogger()
	initPrometheus()

	// 可观测性
	tpCancel := ioc.InitOTEL()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		tpCancel(ctx)
	}()

	app := InitWebServer()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			zap.L().Panic("启动失败", zap.Error(err))
		}
	}

	app.cron.Start()
	defer func() {
		// 等待定时任务退出
		<-app.cron.Stop().Done()
	}()

	server := app.server

	err := server.Run(config.Config.App.HttpServerAddress)
	if err != nil {
		zap.L().Panic("启动失败", zap.Error(err))
	}
	zap.L().Info(
		"启动成功",
		zap.String(
			"address",
			config.Config.App.HttpServerAddress,
		),
	)
}

func initPrometheus() {
	go func() {
		// 专门给 prometheus 用的端口
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8081", nil)
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
