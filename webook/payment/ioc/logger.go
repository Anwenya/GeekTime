package ioc

import (
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitLogger() logger.LoggerV1 {
	cfg := zap.NewDevelopmentConfig()
	err := viper.UnmarshalKey("log", &cfg)
	if err != nil {
		zap.L().Panic("logger配置初始化失败", zap.Error(err))
	}
	l, err := cfg.Build()
	if err != nil {
		zap.L().Panic("logger初始化失败", zap.Error(err))
	}
	zap.L().Debug("logger初始化成功")
	return logger.NewZapLogger(l)
}
