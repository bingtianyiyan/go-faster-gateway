package logger

import (
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/log/logger"
)

func SetupLog(logConfig *log.Logger) {
	log.NewLog(logConfig)
	log.Log = logger.NewHelper(logger.DefaultLogger)
}

//func Setup(conf *viper.Viper) {
//	log.NewLogWithViper(conf)
//	log.Log = logger.NewHelper(logger.DefaultLogger)
//}
