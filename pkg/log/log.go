package log

import (
	"go-faster-gateway/pkg/log/logger"
	plugins "go-faster-gateway/pkg/log/logplugins"
	plugins_zap "go-faster-gateway/pkg/log/logplugins/zap"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap/zapcore"
)

var Log *logger.Helper

type Logger struct {
	Driver  string            `json:"driver"`
	Level   string            `json:"level"`
	WriteTo []*logger.WriteTo `json:"writeTo"`
}

func NewDefaultLog() logger.Logger {
	op := setDefault()
	var logConfig = Logger{
		Driver:  op.driver,
		Level:   op.level,
		WriteTo: op.writeTo,
	}
	return initZap(
		WithDriver(logConfig.Driver),
		WithLevel(logConfig.Level),
		WithWriteTo(logConfig.WriteTo),
	)
}

func NewLog(logConfig *Logger) logger.Logger {
	return initZap(
		WithDriver(logConfig.Driver),
		WithLevel(logConfig.Level),
		WithWriteTo(logConfig.WriteTo),
	)
}

// initZap
func initZap(opts ...Option) logger.Logger {
	op := setDefault()
	for _, o := range opts {
		o(&op)
	}

	var err error
	var strlen = len(op.writeTo)
	var writeToList = make([]io.Writer, strlen)
	for index, wt := range op.writeTo {
		var output io.Writer
		switch strings.ToLower(wt.Name) {
		case "file":
			var fileArgs logger.File
			err := mapstructure.Decode(wt.Args, &fileArgs)
			if err != nil {
				log.Fatalf("decode file log args error:%v", err)
			}
			if len(fileArgs.Path) == 0 {
				fileArgs.Path = "temp/logs"
			}
			output = plugins.NewFileLogger(fileArgs)
		case "console":
			output = plugins.NewConsoleLogger()
		default:
			output = os.Stdout
		}
		writeToList[index] = output
	}

	var level logger.Level
	level, err = logger.GetLevel(op.level)
	if err != nil {
		log.Fatalf("get logger level error, %s", err.Error())
	}

	switch strings.ToLower(op.driver) {
	case "zap":
		newEncoderConfig := zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "name",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		logger.DefaultLogger, err = plugins_zap.NewLogger(
			logger.WithLevel(level),
			plugins_zap.WithOutput(writeToList),
			plugins_zap.WithCallerSkip(2),
			plugins_zap.WithEncoderConfig(newEncoderConfig))
		if err != nil {
			log.Fatalf("new zap logger error, %s", err.Error())
		}
	default:
		logger.DefaultLogger = logger.NewLogger(logger.WithLevel(level), logger.WithOutput(writeToList))
	}
	return logger.DefaultLogger
}
