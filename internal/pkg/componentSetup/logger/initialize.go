package logger

import (
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/log/logger"
)

func SetupLog(logConfig *log.Logger) {
	if logConfig != nil {
		log.NewLog(logConfig)
	} else {
		log.NewDefaultLog()
	}
	log.Log = logger.NewHelper(logger.DefaultLogger)
}
