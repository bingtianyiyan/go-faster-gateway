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
	UpstreamsManager  *balancer.UpstreamManager      // 上游服务，一般路由会保存上游服务的名称，转发到对应的上游服务上去，可以使用负载均衡算法
	ProtocolManager   *protocols.ProtocolFactory
	MiddlewareHandler *middleware.MiddlewareHandler
	Router            IRouter                 // 路由相关信息
	RouteDataProvider data.IRouteResourceData //路由数据
}

func NewRouterManager(upstreamsManager *balancer.UpstreamManager,
	protocolManager *protocols.ProtocolFactory) *RouterManager {
	return &RouterManager{
		UpstreamsManager: upstreamsManager,
		ProtocolManager:  protocolManager,
	}
}

// CreateRouters creates new TCPRouters
func (f *RouterManager) CreateRouters(ctx context.Context, conf dynamic.Configuration) error {
	// TODO 路由数据源初始化(后期可能http+websocket+tcp 这边需要修改 成配置，抽象
	f.RouteDataProvider = provider.NewRouteResourceFileData(conf.EasyServiceRoute.Services)
	//routeData
	routeDataList, err := f.RouteDataProvider.GetAllList(ctx)
	if err != nil {
		return err
	}
	//过滤出是http,https,websocket类型的数据
	filteredRouteDataList := utils.Filter(routeDataList, func(u *dynamic.ServiceRoute) bool {
		return u.Handler == constants.Http || u.Handler == constants.Https || u.Handler == constants.WebSocket
	})
	//middleware
	f.RegisterMiddleHandlers(conf)
	r := NewDyRouter(f.ProtocolManager)
	//这边只需要把http,https,websocket的
	r.BuildRouter(filteredRouteDataList, f.MiddlewareHandler)
	f.Router = r
	handler := r.MainRouter.Handler
	if len(f.MiddlewareHandler.Handler) > 0 {
		for i := len(conf.GlobalMiddleware) - 1; i >= 0; i-- {
			key := strings.ToLower(conf.GlobalMiddleware[i])
			fc, ok := f.MiddlewareHandler.Handler[key]
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
	// 所有的有配置项的中间件，都会配置在middlewares中
	//for k, v := range s.GetConfig().Middlewares {
	//	tempConfig := v
	//	switch {
	//	case v.IPBlacklist != nil:
	//		m.HttpHandler[k] = middleware.IpBlacklistMiddleware(&tempConfig, s)
	//	case v.IPWhitelist != nil:
	//		m.HttpHandler[k] = middleware.IpWhitelistMiddleware(&tempConfig, s)
	//	case v.RequestID != nil:
	//		m.HttpHandler[k] = middleware.RequestIdMiddleware(&tempConfig, s)
	//	case v.CORS != nil:
	//		m.HttpHandler[k] = middleware.CorsMiddleware(&tempConfig, s)
	//	case v.RemoteAuth != nil:
	//		m.HttpHandler[k] = middleware.RemoteAuthMiddleware(&tempConfig, s)
	//	case v.TokenLimiter != nil:
	//		m.HttpHandler[k] = middleware.TokenLimiterMiddleware(&tempConfig, s)
	//	}
	//}
	// 没配置的(主要是一些全局的中间件)
	for _, v := range conf.GlobalMiddleware {
		v := strings.ToLower(v)
		if _, ok := m.Handler[v]; ok {
			continue
		}
		switch {
		case v == "recovery":
			m.Handler[v] = middleware.RecoveryMiddleware
		case v == "errorhandler":
			m.Handler[v] = middleware.ErrorHandlerMiddleware
		}
	}
	f.MiddlewareHandler = &m
}
