package ioc

import (
	"github.com/Anwenya/GeekTime/webook/internal/service/oauth2/wechat"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	//appID, ok := os.LookupEnv("WECHAT_APP_ID")
	//if !ok {
	//	panic(any("找不到环境变量 WECHAT_APP_ID"))
	//}
	//appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	//if !ok {
	//	panic(any("找不到环境变量 WECHAT_APP_SECRET"))
	//}
	return wechat.NewService("", "", l)
}
