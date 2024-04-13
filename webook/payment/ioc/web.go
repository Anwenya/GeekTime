package ioc

import (
	"github.com/Anwenya/GeekTime/webook/payment/web"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx/decorator"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

func InitGinServer(hdl *web.WechatHandler) *ginx.Server {
	engine := gin.Default()
	hdl.RegisterRoutes(engine)
	addr := viper.GetString("http.address")
	decorator.InitCounter(
		prometheus.CounterOpts{
			Namespace: "GeekTime",
			Subsystem: "payment",
			Name:      "http",
		},
	)
	return &ginx.Server{
		Engine: engine,
		Addr:   addr,
	}
}
