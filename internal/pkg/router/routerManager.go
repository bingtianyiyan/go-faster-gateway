package router

import (
	"context"
	"errors"
	"fmt"
	"go-faster-gateway/internal/pkg/data"
	"go-faster-gateway/internal/pkg/data/provider"
	"go-faster-gateway/internal/pkg/ecode"
	"go-faster-gateway/internal/pkg/middleware"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/poxyResource/balancer"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

// RouterManager
type RouterManager struct {
	Handler           func(ctx *fasthttp.RequestCtx) // http handler --> 代理主处理器
	Upstreams         *balancer.Upstream             // 上游服务，一般路由会保存上游服务的名称，转发到对应的上游服务上去，可以使用负载均衡算法
	MiddlewareHandler *middleware.MiddlewareHandler
	Router            IRouter                 // 路由相关信息
	RouteDataProvider data.IRouteResourceData //路由数据
	Config            *dynamic.Configuration
}

func NewRouterManager() *RouterManager {
	return &RouterManager{
		Upstreams: &balancer.Upstream{
			LB:        make(map[string]balancer.Balancer),
			SyncNodes: make(chan balancer.NodeServer, 1),
		},
	}
}

// CreateRouters creates new TCPRouters
func (f *RouterManager) CreateRouters(ctx context.Context, conf dynamic.Configuration) error {
	fmt.Println("create router")
	//路由数据源初始化
	f.RouteDataProvider = provider.NewRouteResourceFileData(conf.HTTP.Services)
	f.Config = &conf
	//routeData
	routeDataList, err := f.RouteDataProvider.GetAllList(ctx)
	if err != nil {
		return err
	}
	//middleware
	f.RegisterMiddleHandlers(f.Config)
	r := NewDyRouter(f.proxyHandler)
	r.BuildRouter(routeDataList, f.MiddlewareHandler)
	f.Router = r
	handler := r.Router.Handler
	globalMiddlewares := f.Config.GlobalMiddleware
	for i := len(globalMiddlewares) - 1; i >= 0; i-- {
		key := strings.ToLower(globalMiddlewares[i])
		handler = f.MiddlewareHandler.Handler[key](handler)
	}
	f.Handler = handler
	return nil
}

func (f *RouterManager) RegisterMiddleHandlers(config *dynamic.Configuration) {
	var m middleware.MiddlewareHandler
	m.Handler = make(map[string]middleware.MiddlewareFunc)
	// 所有的有配置项的中间件，都会配置在middlewares中
	//for k, v := range s.GetConfig().Middlewares {
	//	tempConfig := v
	//	switch {
	//	case v.IPBlacklist != nil:
	//		m.Handler[k] = middleware.IpBlacklistMiddleware(&tempConfig, s)
	//	case v.IPWhitelist != nil:
	//		m.Handler[k] = middleware.IpWhitelistMiddleware(&tempConfig, s)
	//	case v.RequestID != nil:
	//		m.Handler[k] = middleware.RequestIdMiddleware(&tempConfig, s)
	//	case v.CORS != nil:
	//		m.Handler[k] = middleware.CorsMiddleware(&tempConfig, s)
	//	case v.RemoteAuth != nil:
	//		m.Handler[k] = middleware.RemoteAuthMiddleware(&tempConfig, s)
	//	case v.TokenLimiter != nil:
	//		m.Handler[k] = middleware.TokenLimiterMiddleware(&tempConfig, s)
	//	}
	//}
	// 没配置的(主要是一些全局的中间件)
	for _, v := range config.GlobalMiddleware {
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

// 统一反向代理处理器，构建请求至后端服务
func (f *RouterManager) proxyHandler(ctx *fasthttp.RequestCtx, r *dynamic.Service) {
	if ctx.Err() != nil {
		return
	}

	// 处理完成，转发至后端服务
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// 复制客户端请求的数据
	ctx.Request.CopyTo(req)
	req.SetBody(ctx.PostBody())

	// 获取负载均衡地址
	upstreamServer, err := f.GetLBUpstream(r.Service)
	if err != nil {
		ctx.Error(err.Error(), ecode.InternalServerErrorErr.Code)
		return
	}
	proxy := fasthttp.HostClient{
		Addr:  upstreamServer,
		IsTLS: ctx.IsTLS(),
	}
	req.SetRequestURI("http://" + upstreamServer + r.Routers.ProxyPath)
	// 创建一个新的响应
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// 向目标后端服务器发送请求
	err = proxy.DoTimeout(req, resp, time.Second*5)
	if err != nil {
		log.Log.WithError(err).Error("fasthttp.doTimeout()")
		if errors.Is(err, fasthttp.ErrTimeout) {
			ctx.Error(ecode.BackendTimeoutErr.Data(), ecode.BackendTimeoutErr.HttpCode)
		} else {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}
		return
	}

	// 将目标服务器的响应返回给客户端
	// 将目标服务器的响应头部和主体复制到当前请求对象中
	resp.Header.CopyTo(&ctx.Response.Header)
	ctx.Response.SetBody(resp.Body())
}

// GetLBUpstream 获取负载均衡后的上游服务
func (f *RouterManager) GetLBUpstream(serviceName string) (string, error) {
	var (
		us  string
		err error
	)
	us, err = f.Upstreams.GetNextUpstream(serviceName)
	if err != nil && !errors.Is(err, ecode.UpstreamNotInit) {
		return "", err
	}
	var serviceMap = f.Config.HTTP.Services[serviceName]
	var modelNode = func(model *dynamic.Service) []*balancer.Node {
		nodes := make([]*balancer.Node, 0)
		for _, v := range model.Servers {
			node := &balancer.Node{
				Service: v.Host,
				Port:    uint32(v.Port),
				Weight:  int32(v.Weight),
				Healthy: v.Healthy,
			}
			nodes = append(nodes, node)
		}
		return nodes
	}
	nodes := modelNode(serviceMap)
	f.Upstreams.AddToLB(serviceName, nodes, f.Config.BalanceMode.Balance)
	us, err = f.Upstreams.GetNextUpstream(serviceName)
	return us, err
}

// 获取上游信息
func (f *RouterManager) GetUpstream() *balancer.Upstream {
	return f.Upstreams
}
