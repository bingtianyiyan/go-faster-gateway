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
	protocolFactory *protocols.ProtocolFactory
	MainRouter      *fasthttprouter.Router
	handlerLoader   *HandlerLoader
	Md5             string
	mu              sync.RWMutex
}

type SubRouter struct {
	protocolFactory  *protocols.ProtocolFactory
	serviceBaseRoute *dynamic.ServiceRoute
}

func NewDyRouter(protocolFactory *protocols.ProtocolFactory) *DyRouter {
	return &DyRouter{
		apis:            make(map[string]*dynamic.ServiceRoute),
		protocolFactory: protocolFactory,
		MainRouter:      fasthttprouter.New(),
		handlerLoader:   NewHandlerLoader(),
	}
}

func (sr *DyRouter) BuildRouter(apis []*dynamic.ServiceRoute, mwHandler *middleware.MiddlewareHandler) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	// 先加载普通路由
	for _, v := range apis {
		sr.apis[v.RouteName] = v
		for _, v2 := range v.Routers {
			sr.loadRoute(v, v2, nil, mwHandler)
		}
		////1. 静态API路由优先  /api/Account/Login
		////2. 参数路径次之 /static/:param
		////3. 最后定义全局通配符最后 /*path
		////路由匹配优先级：
		////精确路径匹配（完全匹配）
		////正则表达式路由匹配
		////通配符路由（兜底）

	}
	apiJson, _ := json.Marshal(apis)
	sr.Md5 = md5.MD5(apiJson)
}

func (sr *DyRouter) loadRoute(serviceBaseRoute *dynamic.ServiceRoute, routeCfg dynamic.Router, parentRouter *fasthttprouter.Router, mwHandler *middleware.MiddlewareHandler) error {
	currentRouter := sr.MainRouter
	if parentRouter != nil {
		currentRouter = parentRouter
	}
	//校验 routeCfg.Type 如果未设置则根据正则检测是哪种类型
	if len(routeCfg.Routers) > 0 && len(routeCfg.Type) == 0 {
		routeCfg.Type = constants.Subrouter
	} else if len(routeCfg.Type) == 0 {
		routePatern := ParseRoute(routeCfg.Path)
		if routePatern.Type == RouteTypeWildcard {
			routeCfg.Type = constants.Wildcard
		} else if routePatern.Type == RouteTypeParam {
			routeCfg.Type = constants.Param
		} else if routePatern.Type == RouteTypeStatic {
			routeCfg.Type = constants.Static
		} else {
			routeCfg.Type = constants.Subrouter
		}
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
		protocolFactory:  sr.protocolFactory,
		serviceBaseRoute: serviceBaseRoute,
	}
	// 获取基础处理器（已适配为 fasthttp.RequestHandler）
	baseHandler := subSr.AsRequestHandler()
	//全局中间件，服务内全局中间件，路由局部中间件三者中间件
	middlewareList := utils.UnionSlicesUnique(serviceBaseRoute.Middlewares, routeInfo.Middlewares)

	var handlers []middleware.MiddlewareFunc
	if mwHandler != nil {
		for _, mw := range middlewareList {
			if h, ok := mwHandler.Handler[strings.ToLower(mw)]; ok {
				handlers = append(handlers, h)
			}
		}
	}
	wrappedHandler := middleware.Chain(baseHandler, handlers...)

	routeInfo.Path = serviceBaseRoute.RouteGroup + routeInfo.Prefix + "/*path"
	sr.registerRoutePattenByMode(sr.MainRouter, routeInfo, wrappedHandler, serviceBaseRoute.ProtocolName)
	// 其他HTTP方法...
	// 加载子路由
	for _, subRoute := range routeInfo.Routers {
		subRoute.Path = serviceBaseRoute.RouteGroup + routeInfo.Prefix + subRoute.Path
		if err := sr.loadRoute(serviceBaseRoute, subRoute, subRouter, mwHandler); err != nil {
			return err
		}
	}

	return nil
}

// HandleRequest 是您的自定义处理方法
func (sr *SubRouter) HandleRequest(ctx *fasthttp.RequestCtx) {
	handler := sr.protocolFactory.GetHandler(ctx)
	temp := sr.serviceBaseRoute // 假设这是获取临时数据的方法
	handler.Handle(ctx, temp)
}

func (sr *SubRouter) AsRequestHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		sr.HandleRequest(ctx)
	}
}

// loadWildcardRoute 加载通配符路由
func (sr *DyRouter) loadWildcardRoute(currentRoute *fasthttprouter.Router, serviceBaseRoute *dynamic.ServiceRoute, routeInfo dynamic.Router, mwHandler *middleware.MiddlewareHandler) error {
	subSr := &SubRouter{
		protocolFactory:  sr.protocolFactory,
		serviceBaseRoute: serviceBaseRoute,
	}
	// 获取基础处理器（已适配为 fasthttp.RequestHandler）
	baseHandler := subSr.AsRequestHandler()
	//全局中间件，服务内全局中间件，路由局部中间件三者中间件
	middlewareList := utils.UnionSlicesUnique(serviceBaseRoute.Middlewares, routeInfo.Middlewares)

	var handlers []middleware.MiddlewareFunc
	if mwHandler != nil {
		for _, mw := range middlewareList {
			if h, ok := mwHandler.Handler[strings.ToLower(mw)]; ok {
				handlers = append(handlers, h)
			}
		}
	}
	wrappedHandler := middleware.Chain(baseHandler, handlers...)

	// 转换参数路由路径 (如 :id 转换为 :id<regex>)
	routeInfo.Path = serviceBaseRoute.RouteGroup + routeInfo.Prefix + sr.convertParamPath(routeInfo)
	sr.registerRoutePattenByMode(currentRoute, routeInfo, wrappedHandler, serviceBaseRoute.ProtocolName)
	return nil
}

// loadStandardRoute 加载标准路由(静态或参数路由)
func (sr *DyRouter) loadStandardRoute(currentRoute *fasthttprouter.Router, serviceBaseRoute *dynamic.ServiceRoute, routeInfo dynamic.Router, mwHandler *middleware.MiddlewareHandler) error {
	subSr := &SubRouter{
		protocolFactory:  sr.protocolFactory,
		serviceBaseRoute: serviceBaseRoute,
	}
	// 获取基础处理器（已适配为 fasthttp.RequestHandler）
	baseHandler := subSr.AsRequestHandler()
	//全局中间件，服务内全局中间件，路由局部中间件三者中间件
	middlewareList := utils.UnionSlicesUnique(serviceBaseRoute.Middlewares, routeInfo.Middlewares)

	var handlers []middleware.MiddlewareFunc
	if mwHandler != nil {
		for _, mw := range middlewareList {
			if h, ok := mwHandler.Handler[strings.ToLower(mw)]; ok {
				handlers = append(handlers, h)
			}
		}
	}
	wrappedHandler := middleware.Chain(baseHandler, handlers...)
	routeInfo.Path = serviceBaseRoute.RouteGroup + routeInfo.Prefix + routeInfo.Path
	// 注册子路由到主路由
	sr.registerRoutePattenByMode(currentRoute, routeInfo, wrappedHandler, serviceBaseRoute.ProtocolName)
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
