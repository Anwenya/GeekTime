package ioc

import (
	"github.com/Anwenya/GeekTime/webook/token"
	"github.com/Anwenya/GeekTime/webook/util"
	"log"
)

func InitConfig() *util.Config {
	var config, err = util.LoadConfig(".")
	if err != nil {
		log.Fatalf("配置文件加载失败:%v", err)
	}
	log.Println("配置文件加载成功")
	return config
}

func InitTokenMaker(config *util.Config) token.Maker {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSecretKey)
	if err != nil {
		log.Fatalf("初始化tokenMaker失败:%v", err)
	}
	log.Println("tokenMaker初始化成功")
	return tokenMaker
}
