package api

import (
	"context"
	"fmt"
	"go-faster-gateway/internal/pkg/balancer"
	db_init "go-faster-gateway/internal/pkg/componentSetup/database"
	"go-faster-gateway/internal/pkg/protocols"
	servers "go-faster-gateway/internal/pkg/server"
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
	// TODO 这边可以加入其他文件提供者

	//构建blance和协议
	upstreamManager := balancer.NewUpstreamManager()
	httpHandler := protocols.NewHTTPHandler(upstreamManager)
	//TODO websocket
	protocolManager := protocols.NewProtocolFactory([]protocols.ProtocolHandler{
		httpHandler,
	})

	// Watcher
	watcher := configLoader.NewConfigurationWatcher(
		routinesPool,
		providerAggregator,
		"file",
	)
	configManager.SetWatch(watcher)

	//第一次获取动态文件 后面都由watch去更新，这边需要在watch设置完之后获取
	dyConfig, err := configManager.GetDynamicConfig()
	if err != nil {
		log.Log.WithError(err).Error("configManager.GetDynamicConfig fail")
		return err
	}

	routerManager := router.NewRouterManager(upstreamManager, protocolManager)
	serviceManager := servers.NewServiceManager(ctx, configManager, routerManager)
	serviceManager.InitBuildServer()

	//add listener
	watcher.AddListener(switchRouter(serviceManager))

	db_init.SetupDb(dyConfig.Databases)

	svr := server.NewServer(server.WithServiceManager(serviceManager),
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

func switchRouter(serviceManager *servers.ServiceManager) func(conf dynamic.Configuration) {
	return func(conf dynamic.Configuration) {
		serviceManager.SwitchFastHttpRouter(conf)
	}
}
