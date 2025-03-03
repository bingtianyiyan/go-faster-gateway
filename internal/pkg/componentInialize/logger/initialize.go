package logger

import (
	"github.com/spf13/viper"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/log/logger"
)

var Log *logger.Helper

func Setup(conf *viper.Viper) {
	log.NewLog(conf)
	Log = logger.NewHelper(logger.DefaultLogger)
}
