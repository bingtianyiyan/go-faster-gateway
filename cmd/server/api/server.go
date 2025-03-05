package api

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go-faster-gateway/cmd/server/wire"
	logger_init "go-faster-gateway/internal/pkg/componentInialize/logger"
	configLoader "go-faster-gateway/pkg/config"
	"go-faster-gateway/pkg/config/static"
	"go-faster-gateway/pkg/http"
	"go-faster-gateway/pkg/log"
)

var (
	envConf string
	conf    *viper.Viper

	StartCmd = &cobra.Command{
		Use:          "server",
		Short:        "Start faster-gateway server",
		Example:      "faster-gateway server -c config/settings.yaml",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			setup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
)

func init() {
	StartCmd.PersistentFlags().StringVarP(&envConf, "config", "c", "config/settings.yml", "Start server with provided configuration file")
	//flag.StringVar(&envConf, "conf", "settings.yml", "config path, eg: -conf config/settings.yml")
	//flag.Parse()
}

func setup() {
	fmt.Println(`starting faster-gateway server...`)
}

func run() error {
	//config inits
	gatewayConfig := NewGatewayConfiguration()
	gatewayConfig.ConfigFile = envConf
	loaders := []configLoader.ResourceLoader{&configLoader.FileLoader{}}
	cmdGateway := &configLoader.Command{
		Name:          "faster-gateway",
		Description:   `HTTP reverse proxy and load balancer`,
		Configuration: gatewayConfig,
		Resources:     loaders,
		Run: func(_ []string) error {
			return runCmd(&gatewayConfig.Configuration)
		},
	}

	//healthcheck check  TODO

	err := configLoader.Execute(cmdGateway)
	if err != nil {
		log.Log.WithError(err).Error("command error")
		log.Exit(1)
	}
	log.Exit(0)
	return err
}

func runCmd(staticConfiguration *static.Configuration) error {
	//配置文件读取
	//conf = config.NewConfig(envConf)
	//日志初始化
	logger_init.SetupLog(staticConfiguration.Logger)
	//wire ioc
	app, cleanup, err := wire.NewWire()
	if err != nil {
		log.Log.Error("register wire fail", err)
		panic(err)
	}
	defer cleanup()

	http.Run(app, fmt.Sprintf(":%d", conf.GetInt("http.port")))

	//srv := &http.Server{
	//	Addr:    fmt.Sprintf("%s:%d", config.Application.Host, config.Application.Port),
	//	Handler: sm(sdk.Runtime.GetEngine()),
	//}
	//
	//go func() {
	//	var dbs = make(map[string]*gorm.DB, 0)
	//	for k, v := range sdk.Runtime.GetDb() {
	//		if k == global.JobDb {
	//			dbs[k] = v
	//		}
	//	}
	//	if (len(dbs)) > 0 {
	//		jobs.Setup(dbs)
	//	}
	//}()
	//
	//go func() {
	//	// 服务连接
	//	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	//		logger.Logger.Fatal("listen err:", zap.Error(err))
	//	}
	//}()
	//
	//fmt.Println("Server run at:")
	//fmt.Printf("-  Local:   http://localhost:%d/ \r\n", config.Application.Port)
	//fmt.Printf("-  Network: http://%s:%d/ \r\n", utils.GetLocaHonst(), config.Application.Port)
	//fmt.Println("Swagger run at:")
	//fmt.Printf("-  Local:   http://localhost:%d/swagger/index.html \r\n", config.Application.Port)
	//fmt.Printf("-  Network: http://%s:%d/swagger/index.html \r\n", utils.GetLocaHonst(), config.Application.Port)
	//fmt.Printf("%s Enter Control + C Shutdown Server \r\n", utils.GetCurrentTimeStr())
	//// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	//quit := make(chan os.Signal)
	//signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	//<-quit
	//
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	//fmt.Printf("%s Shutdown Server ... \r\n", utils.GetCurrentTimeStr())
	//
	//if err := srv.Shutdown(ctx); err != nil {
	//	logger.Logger.Fatal("Server Shutdown:", zap.Error(err))
	//}
	//logger.Logger.Warn("Server exiting")
	return err
}
