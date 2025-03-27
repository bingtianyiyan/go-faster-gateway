package server

import (
	"context"
	"go-faster-gateway/internal/pkg/router"
	"go-faster-gateway/internal/pkg/server/fast"
	configLoader "go-faster-gateway/pkg/config"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/log"
)

type ServiceManager struct {
	configManager *configLoader.ConfigurationManager
	routeManager  *router.RouterManager
	fastServer    *fast.HttpServer
	ctx           context.Context
}

func NewServiceManager(ctx context.Context, configManager *configLoader.ConfigurationManager, routeManager *router.RouterManager) *ServiceManager {
	return &ServiceManager{
		ctx:           ctx,
		configManager: configManager,
		routeManager:  routeManager,
	}
}

// 构建server资源信息
func (f *ServiceManager) InitBuildServer() {
	//构建 fastHttp
	f.BuildFastHttp()
	//TODO 构建 websocket
}

func (f *ServiceManager) GetConfigManager() *configLoader.ConfigurationManager {
	return f.configManager
}

func (f *ServiceManager) GetRouterManager() *router.RouterManager {
	return f.routeManager
}

func (f *ServiceManager) GetFastServer() *fast.HttpServer {
	return f.fastServer
}

func (f *ServiceManager) BuildFastHttp() *fast.HttpServer {
	dyConfig, err := f.configManager.GetDynamicConfig()
	if err != nil {
		log.Log.WithError(err).Error("BuildFastHttp for GetDynamicConfig fail")
		return nil
	}
	// Switch router  构建路由
	err = f.routeManager.CreateRouters(f.ctx, *dyConfig)
	if err != nil {
		log.Log.WithError(err).Error("BuildFastHttp for CreateRouters fail")
		return nil
	}
	//httpServer
	f.fastServer = fast.NewHttpServer(f.configManager.GetStaticConfig(), f.routeManager.HttpHandler)
	return f.fastServer
}

// TODO BuildWebSocket

// 切换FastHttp的Route
func (f *ServiceManager) SwitchFastHttpRouter(conf dynamic.Configuration) {
	// http对应的路由信息
	err := f.routeManager.CreateRouters(f.ctx, conf)
	if err != nil {
		log.Log.WithError(err).Error("SwitchFastHttpRouter CreateRouters fail")
		return
	}
	f.fastServer.SwitchRouter(f.routeManager.HttpHandler)
	log.Log.Info("SwitchFastHttpRouter success")
}
