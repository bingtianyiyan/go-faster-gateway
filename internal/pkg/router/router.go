package router

import (
	"encoding/json"
	"go-faster-gateway/internal/pkg/constants"
	"go-faster-gateway/internal/pkg/middleware"
	"go-faster-gateway/internal/pkg/protocols"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/helper/md5"
	"go-faster-gateway/pkg/helper/utils"
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
		sr.router[v.RouteName] = v
	}
}

func (sr *StaticRouter) Match(key string) *dynamic.ServiceRoute {
	if api, ok := sr.router[key]; ok {
		return api
	}
	return nil
}

//type routerHandler func(ctx *fasthttp.RequestCtx, dyConfig *dynamic.Configuration, routerInfo *dynamic.RouteName)

// DyRouter 动态路由匹配, 将路由规则最终转换成httprouter
type DyRouter struct {
	apis            map[string]*dynamic.ServiceRoute
	ProtocolFactory *protocols.ProtocolFactory
	MainRouter      *fasthttprouter.Router
	handlerLoader   *HandlerLoader
	Md5             string
	mu              sync.RWMutex
}

type SubRouter struct {
	ProtocolFactory  *protocols.ProtocolFactory
	ServiceBaseRoute *dynamic.ServiceRoute
}

func NewDyRouter(protocolFactory *protocols.ProtocolFactory) *DyRouter {
	return &DyRouter{
		apis:            make(map[string]*dynamic.ServiceRoute),
		ProtocolFactory: protocolFactory,
		MainRouter:      fasthttprouter.New(),
		handlerLoader:   NewHandlerLoader(),
	}
}

func (sr *DyRouter) BuildRouter(apis []*dynamic.ServiceRoute, mwHandler *middleware.MiddlewareHandler) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	//sr.MainRouter = fasthttprouter.New()
	// 先加载普通路由
	for _, v := range apis {
		sr.apis[v.RouteName] = v
		for _, v2 := range v.Routers {
			sr.loadRoute(v, v2, nil, mwHandler)
		}
		//temp := v
		//// 每个路由对应的中间件不一样
		//var handlers []middleware.MiddlewareFunc
		//for _, mw := range v.Middlewares {
		//	if h, ok := mwHandler.ProtocolName[strings.ToLower(mw)]; ok {
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
		//sr.registerRoutePattenMode(v.Routers, chains, RouteTypeStatic, v.ProtocolName)
		////2.参数路由
		//sr.registerRoutePattenMode(v.Routers, chains, RouteTypeParam, v.ProtocolName)
		////3.通配符路由
		//sr.registerRoutePattenMode(v.Routers, chains, RouteTypeWildcard, v.ProtocolName)

	}
	apiJson, _ := json.Marshal(apis)
	sr.Md5 = md5.MD5(apiJson)
}

func (sr *DyRouter) loadRoute(serviceBaseRoute *dynamic.ServiceRoute, routeCfg dynamic.Router, parentRouter *fasthttprouter.Router, mwHandler *middleware.MiddlewareHandler) error {
	//var isRoot = true
	currentRouter := sr.MainRouter
	if parentRouter != nil {
		currentRouter = parentRouter
		//isRoot = false
	}
	switch routeCfg.Type {
	case constants.Subrouter:
		sr.loadSubrouter(serviceBaseRoute, routeCfg, mwHandler)

	case constants.Wildcard:
		sr.loadWildcardRoute(currentRouter, serviceBaseRoute, routeCfg, mwHandler)

	default: // 普通路由static/param
		sr.loadStandardRoute(currentRouter, serviceBaseRoute, routeCfg, mwHandler)
	}

	return nil
}

// loadSubrouter 加载子路由
func (sr *DyRouter) loadSubrouter(serviceBaseRoute *dynamic.ServiceRoute, routeInfo dynamic.Router, mwHandler *middleware.MiddlewareHandler) error {
	subRouter := fasthttprouter.New()

	subSr := &SubRouter{
		ProtocolFactory:  sr.ProtocolFactory,
		ServiceBaseRoute: serviceBaseRoute,
	}
	// 获取基础处理器（已适配为 fasthttp.RequestHandler）
	baseHandler := subSr.AsRequestHandler()
	//全局中间件，服务内全局中间件，路由局部中间件三者中间件
	middlewareList := utils.UnionSlicesUnique(serviceBaseRoute.Middlewares, routeInfo.Middlewares)
	wrappedHandler := sr.applyMiddlewares(baseHandler, mwHandler, middlewareList)

	routeInfo.Path = routeInfo.Prefix + "/*path"
	sr.registerRoutePattenByMode(sr.MainRouter, routeInfo, wrappedHandler, serviceBaseRoute.ProtocolName)
	// 其他HTTP方法...
	// 加载子路由
	for _, subRoute := range routeInfo.Routers {
		subRoute.Path = routeInfo.Prefix + subRoute.Path
		if err := sr.loadRoute(serviceBaseRoute, subRoute, subRouter, mwHandler); err != nil {
			return err
		}
	}

	return nil
}

// HandleRequest 是您的自定义处理方法
func (sr *SubRouter) HandleRequest(ctx *fasthttp.RequestCtx) {
	handler := sr.ProtocolFactory.GetHandler(ctx)
	temp := sr.ServiceBaseRoute // 假设这是获取临时数据的方法
	handler.Handle(ctx, temp)
}

func (sr *SubRouter) AsRequestHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		sr.HandleRequest(ctx)
	}
}

// loadWildcardRoute 加载通配符路由
func (sr *DyRouter) loadWildcardRoute(currentRoute *fasthttprouter.Router, serviceBaseRoute *dynamic.ServiceRoute, routeInfo dynamic.Router, mwHandler *middleware.MiddlewareHandler) error {
	temp := serviceBaseRoute
	//全局中间件，服务内全局中间件，路由局部中间件三者中间件
	// 每个路由对应的中间件不一样
	var handlers []middleware.MiddlewareFunc
	middlewareList := utils.UnionSlicesUnique(serviceBaseRoute.Middlewares, routeInfo.Middlewares)
	for _, mw := range middlewareList {
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

	// 转换参数路由路径 (如 :id 转换为 :id<regex>)
	routeInfo.Path = sr.convertParamPath(routeInfo)
	sr.registerRoutePattenByMode(currentRoute, routeInfo, chains, serviceBaseRoute.ProtocolName)
	return nil
}

// loadStandardRoute 加载标准路由(静态或参数路由)
func (sr *DyRouter) loadStandardRoute(currentRoute *fasthttprouter.Router, serviceBaseRoute *dynamic.ServiceRoute, routeInfo dynamic.Router, mwHandler *middleware.MiddlewareHandler) error {
	temp := serviceBaseRoute
	//全局中间件，服务内全局中间件，路由局部中间件三者中间件
	// 每个路由对应的中间件不一样
	var handlers []middleware.MiddlewareFunc
	middlewareList := utils.UnionSlicesUnique(serviceBaseRoute.Middlewares, routeInfo.Middlewares)
	for _, mw := range middlewareList {
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
	sr.registerRoutePattenByMode(currentRoute, routeInfo, chains, serviceBaseRoute.ProtocolName)
	return nil
}

// applyMiddlewares 应用中间件件链
// 参数说明:
// - handler: 基础请求处理器，已经是函数类型不需要指针
// - mwNames: 要应用的中间件名称列表
// 返回值: 包装了中间件的新处理器
func (sr *DyRouter) applyMiddlewares(
	handler fasthttp.RequestHandler,
	mwHandler *middleware.MiddlewareHandler,
	mwNames []string,
) fasthttp.RequestHandler {
	// 从后向前应用中间件（最先添加的中间件最后执行）
	for i := len(mwNames) - 1; i >= 0; i-- {
		if mw, ok := mwHandler.Handler[strings.ToLower(mwNames[i])]; ok {
			handler = mw(handler)
		}
	}
	return handler
}

func (sr *DyRouter) registerRoutePattenByMode(currentRoute *fasthttprouter.Router, route dynamic.Router, chains fasthttp.RequestHandler, webSocketType string) {
	//websocket 特殊处理
	if len(route.Methods) == 0 && webSocketType == constants.WebSocket {
		currentRoute.GET(route.Path, chains)
	} else {
		if len(route.Methods) == 0 {
			route.Methods = append(route.Methods, "*")
		}
		for _, reqMethod := range route.Methods {
			if reqMethod == "*" {
				sr.registerAllMethods(currentRoute, route.Path, chains)
			} else {
				currentRoute.Handle(reqMethod, route.Path, chains)
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
