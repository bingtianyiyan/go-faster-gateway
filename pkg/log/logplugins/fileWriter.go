package plugins

import (
	"go-faster-gateway/pkg/log/logger"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

//func (k *FileWriter) Write(p []byte) (n int, err error) {
//	return len(p), nil
//}
//
//func (k *FileWriter) Sync() error {
//	return nil
//}
//
//type FileWriter struct {
//	File *os.File
//}
//
//func FileSyncer(file *os.File) *ConsoleWriter {
//	return &FileWriter{File: file}
//}

func NewFileLogger(fileArgs logger.File) zapcore.WriteSyncer {
	writeSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fileArgs.Path + "." + fileArgs.Suffix, //日志文件存放目录，如果文件夹不存在会自动创建
		MaxSize:    fileArgs.MaxSize,                      //文件大小限制,单位MB
		MaxBackups: fileArgs.MaxBackups,                   //最大保留日志文件数量
		MaxAge:     fileArgs.MaxAge,                       //日志文件保留天数
		Compress:   fileArgs.Compress,                     //是否压缩处理
	})
	return writeSyncer
}
