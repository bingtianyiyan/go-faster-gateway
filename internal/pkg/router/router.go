package router

import (
	"encoding/json"
	"go-faster-gateway/internal/pkg/constants"
	"go-faster-gateway/internal/pkg/middleware"
	"go-faster-gateway/internal/pkg/protocols"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/helper/md5"
	"strings"
	"sync"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

// 路由处理器
type IRouter interface {
	BuildRouter([]*dynamic.ServiceRoute, *middleware.MiddlewareHandler)
	Match(key string) *dynamic.ServiceRoute
	GetMd5() string
}

// 静态路由处理
type StaticRouter struct {
	router map[string]*dynamic.ServiceRoute
	mu     sync.RWMutex
}

// NewStaticRouter 静态路由匹配
func NewStaticRouter() *StaticRouter {
	return &StaticRouter{
		router: make(map[string]*dynamic.ServiceRoute),
	}
}

func (sr *StaticRouter) BuildRouter(apis []*dynamic.ServiceRoute) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	for _, v := range apis {
		sr.router[v.ServiceName] = v
	}
}

func (sr *StaticRouter) Match(key string) *dynamic.ServiceRoute {
	if api, ok := sr.router[key]; ok {
		return api
	}
	return nil
}

//type routerHandler func(ctx *fasthttp.RequestCtx, dyConfig *dynamic.Configuration, routerInfo *dynamic.ServiceName)

// DyRouter 动态路由匹配, 将路由规则最终转换成httprouter
type DyRouter struct {
	apis            map[string]*dynamic.ServiceRoute
	ProtocolFactory *protocols.ProtocolFactory
	MainRouter      *fasthttprouter.Router
	SubRouters      map[string]*fasthttprouter.Router
	handlerLoader   *HandlerLoader
	Md5             string
	mu              sync.RWMutex
}

func NewDyRouter(protocolFactory *protocols.ProtocolFactory) *DyRouter {
	return &DyRouter{
		apis:            make(map[string]*dynamic.ServiceRoute),
		ProtocolFactory: protocolFactory,
		MainRouter:      fasthttprouter.New(),
		SubRouters:      make(map[string]*fasthttprouter.Router),
		handlerLoader:   NewHandlerLoader(),
	}
}

func (sr *DyRouter) BuildRouter(apis []*dynamic.ServiceRoute, mwHandler *middleware.MiddlewareHandler) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	//sr.MainRouter = fasthttprouter.New()
	for _, v := range apis {
		sr.apis[v.ServiceName] = v
		//temp := v
		sr.loadRoute(v, nil, mwHandler)

		//// 每个路由对应的中间件不一样
		//var handlers []middleware.MiddlewareFunc
		//for _, mw := range v.Middlewares {
		//	if h, ok := mwHandler.Handler[strings.ToLower(mw)]; ok {
		//		handlers = append(handlers, h)
		//	}
		//}
		//h := func(ctx *fasthttp.RequestCtx) {
		//	handler := sr.ProtocolFactory.GetHandler(ctx)
		//	//具体处理的事件
		//	handler.Handle(ctx, temp)
		//}
		//chains := middleware.Chain(h, handlers...)
		////1. 静态API路由优先  /api/Account/Login
		////2. 参数路径次之 /static/:param
		////3. 最后定义全局通配符最后 /*path
		////路由匹配优先级：
		////精确路径匹配（完全匹配）
		////正则表达式路由匹配
		////通配符路由（兜底）
		//
		////handler处理
		////1.静态路由
		//sr.registerRoutePattenMode(v.Routers, chains, RouteTypeStatic, v.Handler)
		////2.参数路由
		//sr.registerRoutePattenMode(v.Routers, chains, RouteTypeParam, v.Handler)
		////3.通配符路由
		//sr.registerRoutePattenMode(v.Routers, chains, RouteTypeWildcard, v.Handler)

	}
	apiJson, _ := json.Marshal(apis)
	sr.Md5 = md5.MD5(apiJson)
}

func (sr *DyRouter) loadRoute(routeCfg *dynamic.ServiceRoute, parentRouter *fasthttprouter.Router, mwHandler *middleware.MiddlewareHandler) error {
	for _, routeInfo := range routeCfg.Routers {
		currentRouter := sr.MainRouter
		if parentRouter != nil {
			currentRouter = parentRouter
		}
		//temp := routeCfg
		switch routeInfo.Type {
		case "subrouter":
			sr.loadSubrouter(routeInfo, routeCfg, nil, mwHandler)

		case "wildcard":
			sr.loadWildcardRoute(currentRouter, routeInfo, routeCfg, mwHandler)

		default: // 普通路由
			sr.loadStandardRoute(currentRouter, routeInfo, routeCfg, mwHandler)
		}
	}

	return nil
}

// loadSubrouter 加载子路由
func (sr *DyRouter) loadSubrouter(routeInfo dynamic.Router, routeCfg *dynamic.ServiceRoute, parentRouter *fasthttprouter.Router, mwHandler *middleware.MiddlewareHandler) error {
	temp := routeCfg
	subRouter := fasthttprouter.New()
	sr.SubRouters[routeCfg.ServiceName] = subRouter

	// 每个路由对应的中间件不一样
	var handlers []middleware.MiddlewareFunc
	for _, mw := range routeCfg.Middlewares {
		if h, ok := mwHandler.Handler[strings.ToLower(mw)]; ok {
			handlers = append(handlers, h)
		}
	}
	h := func(ctx *fasthttp.RequestCtx) {
		handler := sr.ProtocolFactory.GetHandler(ctx)
		//具体处理的事件
		handler.Handle(ctx, temp)
	}
	chains := middleware.Chain(h, handlers...)

	// 注册子路由到主路由
	sr.registerRoutePattenByMode(subRouter, routeInfo, chains, routeCfg.Handler)
	// 其他HTTP方法...
	//// 加载子路由
	//for _, subRoute := range routeCfg.Routes {
	//	if err := rl.loadRoute(subRoute, subRouter); err != nil {
	//		return err
	//	}
	//}

	return nil
}

// loadWildcardRoute 加载通配符路由
func (sr *DyRouter) loadWildcardRoute(currentRoute *fasthttprouter.Router, routeInfo dynamic.Router, routeCfg *dynamic.ServiceRoute, mwHandler *middleware.MiddlewareHandler) error {
	temp := routeCfg
	// 每个路由对应的中间件不一样
	var handlers []middleware.MiddlewareFunc
	for _, mw := range routeCfg.Middlewares {
		if h, ok := mwHandler.Handler[strings.ToLower(mw)]; ok {
			handlers = append(handlers, h)
		}
	}
	h := func(ctx *fasthttp.RequestCtx) {
		handler := sr.ProtocolFactory.GetHandler(ctx)
		//具体处理的事件
		handler.Handle(ctx, temp)
	}
	chains := middleware.Chain(h, handlers...)

	// 注册子路由到主路由
	// 转换参数路由路径 (如 :id 转换为 :id<regex>)
	path := sr.convertParamPath(routeInfo)
	routeInfo.Path = path
	sr.registerRoutePattenByMode(currentRoute, routeInfo, chains, routeCfg.Handler)
	return nil
}

// loadStandardRoute 加载标准路由(静态或参数路由)
func (sr *DyRouter) loadStandardRoute(currentRoute *fasthttprouter.Router, routeInfo dynamic.Router, routeCfg *dynamic.ServiceRoute, mwHandler *middleware.MiddlewareHandler) error {
	temp := routeCfg
	// 每个路由对应的中间件不一样
	var handlers []middleware.MiddlewareFunc
	for _, mw := range routeCfg.Middlewares {
		if h, ok := mwHandler.Handler[strings.ToLower(mw)]; ok {
			handlers = append(handlers, h)
		}
	}
	h := func(ctx *fasthttp.RequestCtx) {
		handler := sr.ProtocolFactory.GetHandler(ctx)
		//具体处理的事件
		handler.Handle(ctx, temp)
	}
	chains := middleware.Chain(h, handlers...)

	// 注册子路由到主路由
	sr.registerRoutePattenByMode(currentRoute, routeInfo, chains, routeCfg.Handler)
	return nil
}

func (sr *DyRouter) registerRoutePattenByMode(currentRoute *fasthttprouter.Router, route dynamic.Router, chains fasthttp.RequestHandler, webSocketType string) {
	//websocket 特殊处理
	if len(route.Methods) == 0 && webSocketType == constants.WebSocket {
		currentRoute.GET(route.Path, chains)
	} else {
		for _, reqMethod := range route.Methods {
			//// 转换参数路由路径 (如 :id 转换为 :id<regex>)
			//path := sr.convertParamPath(route)
			if reqMethod == "*" {
				sr.registerAllMethods(currentRoute, route.Prefix+route.Path, chains)
			} else {
				currentRoute.Handle(reqMethod, route.Prefix+route.Path, chains)
			}
		}
	}
}

func (sr *DyRouter) registerAllMethods(currentRoute *fasthttprouter.Router, path string, handler fasthttp.RequestHandler) {
	currentRoute.GET(path, handler)
	currentRoute.POST(path, handler)
	currentRoute.PUT(path, handler)
	currentRoute.DELETE(path, handler)
	currentRoute.PATCH(path, handler)
	currentRoute.HEAD(path, handler)
	currentRoute.OPTIONS(path, handler)
}

func (sr *DyRouter) convertParamPath(route dynamic.Router) string {
	if route.Type != "param" || len(route.Params) == 0 {
		return route.Path
	}

	path := route.Path
	for param, pattern := range route.Params {
		// 将 :id 转换为 :id<regex>
		path = strings.Replace(path, ":"+param, ":"+param+"<"+pattern+">", 1)
	}
	return path
}

// Match 动态路由不需要，完全由router去代理
func (sr *DyRouter) Match(_ string) *dynamic.ServiceRoute {
	return nil
}

func (sr *DyRouter) GetMd5() string {
	return sr.Md5
}
