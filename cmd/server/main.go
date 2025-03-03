package main

import (
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"go-faster-gateway/cmd/server/wire"
	logger_init "go-faster-gateway/internal/pkg/componentInialize/logger"
	"go-faster-gateway/internal/pkg/config"
	"go-faster-gateway/pkg/http"
)

var (
	envConf string
	conf    *viper.Viper
)

func init() {
	flag.StringVar(&envConf, "conf", "settings.yml", "config path, eg: -conf config/settings.yml")
	flag.Parse()
}

func main() {
	//配置文件读取
	conf = config.NewConfig(envConf)
	//日志初始化
	logger_init.Setup(conf)
	//wire ioc
	app, cleanup, err := wire.NewWire(conf)
	if err != nil {
		logger_init.Log.Error("register wire fail", err)
		panic(err)
	}
	defer cleanup()

	http.Run(app, fmt.Sprintf(":%d", conf.GetInt("http.port")))
}
