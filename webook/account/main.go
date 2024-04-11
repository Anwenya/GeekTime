package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViperWatch()
	app := Init()
	err := app.server.Serve()
	if err != nil {
		panic(any(err))
	}
}

func initViperWatch() {
	// 从参数中指定配置文件路径
	cfile := pflag.String("config", "config/dev.yaml", "配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	viper.WatchConfig()
	err := viper.ReadInConfig()
	if err != nil {
		panic(any(err))
	}
}
