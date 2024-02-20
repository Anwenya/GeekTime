package main

import (
	"github.com/Anwenya/GeekTime/webook/util"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
)

func main() {
	server := InitWebServer()
	err := server.Run(util.Config.HTTPServerAddress)
	if err != nil {
		log.Fatalf("启动失败:%v", err)
	}
	log.Printf("启动成功:%v", util.Config.HTTPServerAddress)
}
