package logger

import (
	"github.com/spf13/viper"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/log/logger"
)

func SetupLog(logConfig *log.Logger) {
	log.NewLog(logConfig)
	log.Log = logger.NewHelper(logger.DefaultLogger)
}

func Setup(conf *viper.Viper) {
	log.NewLogWithViper(conf)
	log.Log = logger.NewHelper(logger.DefaultLogger)
}
