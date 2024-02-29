package startup

import (
	"github.com/Anwenya/GeekTime/webook/internal/service/oauth2/wechat"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	return wechat.NewService("", "", l)
}
