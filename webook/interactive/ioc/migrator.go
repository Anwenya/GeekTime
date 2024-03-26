package ioc

import (
	"github.com/Anwenya/GeekTime/webook/interactive/repository/dao"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx/decorator"
	"github.com/Anwenya/GeekTime/webook/pkg/gormx/connpool"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/Anwenya/GeekTime/webook/pkg/migrator/events"
	"github.com/Anwenya/GeekTime/webook/pkg/migrator/events/fixer"
	"github.com/Anwenya/GeekTime/webook/pkg/migrator/scheduler"
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

func InitGinxServer(
	src SrcDB,
	dst DstDB,
	pool *connpool.DoubleWritePool,
	producer events.Producer,
	l logger.LoggerV1,
) *ginx.Server {
	engine := gin.Default()
	group := engine.Group("/migrator")
	decorator.InitCounter(prometheus.CounterOpts{
		Namespace: "GeekTime",
		Subsystem: "webook_intr_admin",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})
	sch := scheduler.NewScheduler[dao.Interactive](l, src, dst, pool, producer)
	sch.RegisterRoutes(group)
	return &ginx.Server{
		Engine: engine,
		Addr:   viper.GetString("migrator.http.addr"),
	}
}

func InitInteractiveProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer("inconsistent_interactive", p)
}

func InitFixerConsumer(
	client sarama.Client,
	l logger.LoggerV1,
	src SrcDB,
	dst DstDB,
) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, l, "inconsistent_interactive", src, dst)
	if err != nil {
		panic(err)
	}
	return res
}
