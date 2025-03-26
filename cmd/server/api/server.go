package api

import (
	"context"
	"fmt"
	"go-faster-gateway/internal/pkg/balancer"
	db_init "go-faster-gateway/internal/pkg/componentSetup/database"
	"go-faster-gateway/internal/pkg/protocols"
	"go-faster-gateway/internal/pkg/server/fast"
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
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/provider/aggregator"
	"go-faster-gateway/pkg/safe"
)

var (
	defaultConfig string
	configManager *configLoader.ConfigurationManager
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
	fmt.Println(`starting faster-gateway server...`)
	//日志初始化 一个默认
	logger_init.SetupLog(nil)
	configManager = configLoader.NewConfigurationManager(defaultConfig, true)
	staticConfig := configManager.GetStaticConfig()
	if staticConfig == nil {
		log.Log.Error("init preRun fail,get staticConfig fail")
		log.Exit(1)
	} else {
		//日志初始化
		logger_init.SetupLog(staticConfig.Logger)
	}
	fmt.Println("init preRun success")
}

func run() error {

	//Provider 机制是其架构体系中的一个核心概念和独特之处，它允许与各种云原生平台、服务发现工具等进行集成和交互
	providerAggregator := aggregator.NewProviderAggregator(*configManager.GetStaticConfig().Providers)
	ctx := logger.NewContext(context.Background(), log.Log)
	routinesPool := safe.NewPool(ctx)
	//这边可以加入其他文件提供者

	upstreamManager := balancer.NewUpstreamManager()
	httpHandler := protocols.NewHTTPHandler(upstreamManager)
	protocolManager := protocols.NewProtocolFactory([]protocols.ProtocolHandler{
		httpHandler,
	})

	// Watcher
	watcher := configLoader.NewConfigurationWatcher(
		routinesPool,
		providerAggregator,
		"file",
	)
	routerManager := router.NewRouterManager(upstreamManager, protocolManager)
	//httpServer
	fastHttp := fast.NewHttpServer(configManager.GetStaticConfig(), routerManager)
	// Switch router  构建路由
	configManager.SetWatch(watcher)
	//第一次获取动态文件 并且设置，后面都由watch去更新
	dyConfig, err := configManager.GetWatcher().GetConfig()
	if err != nil {
		log.Log.WithError(err).Error("configManager.GetWatcher().GetConfig() fail")
		return err
	} else {
		configManager.SetDynamicConfig(dyConfig)
	}
	routerManager.CreateRouters(ctx, *dyConfig)
	//add listener
	watcher.AddListener(switchRouter(ctx, routerManager, fastHttp))

	db_init.SetupDb(dyConfig.Databases)

	svr := server.NewServer(server.WithConfigurationManager(configManager),
		server.WithFastHttpServer(fastHttp),
		server.WithRoutePool(routinesPool),
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
	log.Exit(0)
	return err
}

func switchRouter(ctx context.Context, routerFactory *router.RouterManager, fastHttp *fast.HttpServer) func(conf dynamic.Configuration) {
	return func(conf dynamic.Configuration) {
		fmt.Println("switch router")
		// http对应的路由信息
		routerFactory.CreateRouters(ctx, conf)
		fastHttp.SwitchRouter()
		log.Log.Info("finish print switchRouter")
	}
}
