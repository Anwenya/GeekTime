package ioc

import (
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms/localsms"
)

func InitSMSService() sms.SMService {
	return localsms.NewService()
}
