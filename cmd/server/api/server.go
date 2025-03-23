package api

import (
	"context"
	"fmt"
	db_init "go-faster-gateway/internal/pkg/componentSetup/database"
	"go-faster-gateway/pkg/log/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/spf13/cobra"

	logger_init "go-faster-gateway/internal/pkg/componentSetup/logger"
	"go-faster-gateway/internal/pkg/router"
	"go-faster-gateway/internal/server"
	configLoader "go-faster-gateway/pkg/config"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/config/static"

	//"go-faster-gateway/pkg/http"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/provider/aggregator"
	"go-faster-gateway/pkg/safe"
)

var (
	err           error
	defaultConfig string
	staticConfig  *static.Configuration
	StartCmd      = &cobra.Command{
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
	StartCmd.PersistentFlags().StringVarP(&defaultConfig, "config", "c", "config/settings.yml", "Start server with provided configuration file")
}

func setup() {
	fmt.Println(`starting faster-gateway server...,load static config and init log`)
	//static config inits
	staticConfig, err = static.NewStaticConfig(defaultConfig)
	if err != nil {
		//日志初始化 一个默认
		logger_init.SetupLog(nil)
		log.Log.WithError(err).Error("init preRun fail,get staticConfig fail")
		log.Exit(1)
	} else {
		//日志初始化
		logger_init.SetupLog(staticConfig.Logger)
	}
	fmt.Println("init preRun success")
}

func run() error {
	////wire ioc
	//app, cleanup, err := wire.NewWire()
	//if err != nil {
	//	log.Log.Error("register wire fail", err)
	//	panic(err)
	//}
	//defer cleanup()

	//Provider 机制是其架构体系中的一个核心概念和独特之处，它允许与各种云原生平台、服务发现工具等进行集成和交互
	providerAggregator := aggregator.NewProviderAggregator(*staticConfig.Providers)
	ctx := logger.NewContext(context.Background(), log.Log)
	routinesPool := safe.NewPool(ctx)
	//这边可以加入其他文件提供者

	//db
	//db_init.SetupDb(dynamic.Configuration)

	// Watcher

	watcher := configLoader.NewConfigurationWatcher(
		routinesPool,
		providerAggregator,
		"file",
	)

	//1.构建route和route对应的中间件相关组件做关联对应
	routeManger := router.NewRouterManager()
	//2.
	watcher.AddListener(switchDb(ctx))
	// Switch router  构建tcp和udp路由
	watcher.AddListener(switchRouter(ctx, routeManger))

	svr := server.NewServer(server.WithConfiguration(staticConfig),
		server.WithRoutePool(routinesPool),
		server.WithWatch(watcher),
		server.WithSignals(make(chan os.Signal, 1)),
		server.WithStopChan(make(chan bool, 1)))
	ctx, _ = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	//服务入口接收请求
	svr.Start(ctx)
	defer svr.Close()

	sent, err := daemon.SdNotify(false, "READY=1")
	if !sent && err != nil {
		log.Log.WithError(err).Error("Failed to notify")
	}

	//t, err := daemon.SdWatchdogEnabled(false)
	//if err != nil {
	//	log.Log.WithError(err).Error("Could not enable Watchdog")
	//}

	svr.Wait()
	log.Log.Info("Shutting down")

	//http.Run(app, fmt.Sprintf(":%d", conf.GetInt("http.port")))

	return err
	log.Exit(0)
	return err
}

func switchRouter(ctx context.Context, routerFactory *router.RouterManager) func(conf dynamic.Configuration) {
	return func(conf dynamic.Configuration) {
		fmt.Println("switch router")
		// http对应的路由信息
		routerFactory.CreateRouters(ctx, conf)
		log.Log.Info(conf.BalanceMode.Balance)
		log.Log.Info("finish print switchRouter")
	}
}

func switchDb(ctx context.Context) func(conf dynamic.Configuration) {
	return func(conf dynamic.Configuration) {
		fmt.Println("switch db")
		db_init.SetupDb(conf.Databases)
	}
}
