package log

import "go-faster-gateway/pkg/log/logger"

type Option func(*options)

type options struct {
	//zap,logrus
	driver string
	//log level
	level string
	//log out tool
	writeTo []*logger.WriteTo
}

func setDefault() options {
	return options{
		driver: "default",
		level:  "warn",
		writeTo: []*logger.WriteTo{
			{Name: "file", Args: logger.File{Path: "temp/logs"}},
		},
	}
}

func WithDriver(s string) Option {
	return func(o *options) {
		o.driver = s
	}
}

func WithLevel(s string) Option {
	return func(o *options) {
		o.level = s
	}
}

func WithWriteTo(s []*logger.WriteTo) Option {
	return func(o *options) {
		o.writeTo = s
	}
}
