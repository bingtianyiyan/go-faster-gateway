package router

import (
	"context"
	"go-faster-gateway/internal/pkg/balancer"
	"go-faster-gateway/internal/pkg/constants"
	"go-faster-gateway/internal/pkg/data"
	"go-faster-gateway/internal/pkg/data/provider"
	"go-faster-gateway/internal/pkg/middleware"
	"go-faster-gateway/internal/pkg/protocols"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/helper/utils"
	"strings"

	"github.com/valyala/fasthttp"
)

// RouterManager
type RouterManager struct {
	HttpHandler       func(ctx *fasthttp.RequestCtx) // http handler --> 代理主处理器
	upstreamsManager  *balancer.UpstreamManager      // 上游服务，一般路由会保存上游服务的名称，转发到对应的上游服务上去，可以使用负载均衡算法
	protocolManager   *protocols.ProtocolFactory
	middlewareHandler *middleware.MiddlewareHandler
	middlewareManager *middleware.MiddlewareManager
	router            IRouter                 // 路由相关信息
	routeDataProvider data.IRouteResourceData //路由数据
}

func NewRouterManager(upstreamsManager *balancer.UpstreamManager,
	protocolManager *protocols.ProtocolFactory,
	middlewareManager *middleware.MiddlewareManager) *RouterManager {
	return &RouterManager{
		upstreamsManager:  upstreamsManager,
		protocolManager:   protocolManager,
		middlewareManager: middlewareManager,
	}
}

// CreateRouters creates new TCPRouters
func (f *RouterManager) CreateRouters(ctx context.Context, conf dynamic.Configuration) error {
	// TODO 路由数据源初始化(后期可能http+websocket+tcp 这边需要修改 成配置，抽象
	f.routeDataProvider = provider.NewRouteResourceFileData(conf)
	//routeData
	routeDataList, err := f.routeDataProvider.GetAllList(ctx)
	if err != nil {
		return err
	}
	//过滤出是http,https,websocket类型的数据
	filteredRouteDataList := utils.Filter(routeDataList, func(u *dynamic.ServiceRoute) bool {
		return u.ProtocolName == constants.Http || u.ProtocolName == constants.Https || u.ProtocolName == constants.WebSocket
	})
	//middleware
	f.RegisterMiddleHandlers(conf)
	r := NewDyRouter(f.protocolManager)
	//这边只需要把http,https,websocket的
	r.BuildRouter(filteredRouteDataList, f.middlewareHandler)
	f.router = r
	handler := r.MainRouter.Handler
	if len(f.middlewareHandler.Handler) > 0 {
		for i := len(conf.GlobalMiddleware) - 1; i >= 0; i-- {
			key := strings.ToLower(conf.GlobalMiddleware[i])
			fc, ok := f.middlewareHandler.Handler[key]
			if ok {
				handler = fc(handler)
			}
		}
	}
	f.HttpHandler = handler
	return nil
}

func (f *RouterManager) RegisterMiddleHandlers(conf dynamic.Configuration) {
	var m middleware.MiddlewareHandler
	m.Handler = make(map[string]middleware.MiddlewareFunc)
	// 所有需要的都会配置在middlewares中
	for _, v := range conf.Middlewares {
		tempHandler, ok := f.middlewareManager.Get(v)
		if ok {
			m.Handler[v] = tempHandler
		}
	}
	// 全局中间件
	for _, v := range conf.GlobalMiddleware {
		v1 := strings.ToLower(v)
		if _, ok := m.Handler[v]; ok {
			continue
		}
		switch {
		case v1 == "recoveryhandler" || v1 == "errorhandler":
			tempHandler, ok := f.middlewareManager.Get(v)
			if ok {
				m.Handler[v] = tempHandler
			}
		}
	}
	f.middlewareHandler = &m
}
