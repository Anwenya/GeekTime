package startup

import (
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
)

func InitLogger() logger.LoggerV1 {
	return logger.NewNopLogger()
}
