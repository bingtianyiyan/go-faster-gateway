package plugins

import (
	"go.uber.org/zap/zapcore"
	"os"
)

func (k *ConsoleWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (k *ConsoleWriter) Sync() error {
	return nil
}

type ConsoleWriter struct {
	File *os.File
}

func ConsoleSyncer(file *os.File) *ConsoleWriter {
	return &ConsoleWriter{File: file}
}

func NewConsoleLogger() zapcore.WriteSyncer {
	var cw = ConsoleSyncer(os.Stdout)
	consoleDebugging := zapcore.Lock(cw.File)
	writeSyncer := zapcore.AddSync(consoleDebugging)
	return writeSyncer
}
