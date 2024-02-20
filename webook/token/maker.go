package token

import (
	"github.com/Anwenya/GeekTime/webook/util"
	"log"
	"time"
)

// Maker is an interface for managing tokens
type Maker interface {
	// CreateToken creates a new token for a specific username and duration
	CreateToken(uid int64, username string, duration time.Duration, ua string) (string, *Payload, error)

	// VerifyToken checks if the token is valid or not
	VerifyToken(token string) (*Payload, error)
}

var TkMaker Maker

func init() {
	var err error
	TkMaker, err = NewPasetoMaker(util.Config.TokenSecretKey)
	if err != nil {
		log.Fatalf("初始化tokenMaker失败:%v", err)
	}
	log.Println("tokenMaker初始化成功")
}
