package api

import (
	"context"
	"fmt"
	"go-faster-gateway/pkg/log/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	//"go-faster-gateway/cmd/server/wire"
	db_init "go-faster-gateway/internal/pkg/componentSetup/database"
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
}

func setup() {
	fmt.Println(`starting faster-gateway server...`)
}

func run() error {
	//config inits
	gatewayConfig := configLoader.NewGatewayConfiguration()
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
	//日志初始化
	logger_init.SetupLog(staticConfiguration.Logger)
	////wire ioc
	//app, cleanup, err := wire.NewWire()
	//if err != nil {
	//	log.Log.Error("register wire fail", err)
	//	panic(err)
	//}
	//defer cleanup()

	//Provider 机制是其架构体系中的一个核心概念和独特之处，它允许与各种云原生平台、服务发现工具等进行集成和交互
	providerAggregator := aggregator.NewProviderAggregator(*staticConfiguration.Providers)
	ctx := logger.NewContext(context.Background(), log.Log)
	routinesPool := safe.NewPool(ctx)
	//这边可以加入其他文件提供者

	//db
	db_init.SetupDb(staticConfiguration.Databases)

	// Watcher

	watcher := configLoader.NewConfigurationWatcher(
		routinesPool,
		providerAggregator,
		"internal",
	)

	//1.构建route和route对应的中间件相关组件做关联对应
	routeManger := router.NewRouterManager()
	//2.
	// Switch router  构建tcp和udp路由
	watcher.AddListener(switchRouter(ctx, routeManger))

	svr := server.NewServer(server.WithConfiguration(staticConfiguration),
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
}

func switchRouter(ctx context.Context, routerFactory *router.RouterManager) func(conf dynamic.Configuration) {
	return func(conf dynamic.Configuration) {
		fmt.Println("switch router")
		// http对应的路由信息
		routerFactory.CreateRouters(ctx, conf)
		fmt.Println(conf.BalanceMode.Balance)
		fmt.Println("finish print")
	}
}
