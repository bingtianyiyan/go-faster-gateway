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
	Router          *fasthttprouter.Router
	Md5             string
	mu              sync.RWMutex
}

func NewDyRouter(protocolFactory *protocols.ProtocolFactory) *DyRouter {
	return &DyRouter{
		apis:            make(map[string]*dynamic.ServiceRoute),
		ProtocolFactory: protocolFactory,
	}
}

func (sr *DyRouter) BuildRouter(apis []*dynamic.ServiceRoute, mwHandler *middleware.MiddlewareHandler) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.Router = fasthttprouter.New()
	for _, v := range apis {
		sr.apis[v.ServiceName] = v
		temp := v
		// 每个路由对应的中间件不一样
		var handlers []middleware.MiddlewareFunc
		for _, mw := range v.Middleware {
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
		//1. 静态API路由优先  /api/Account/Login
		//2. 参数路径次之 /static/:param
		//3. 最后定义全局通配符最后 /*path
		//路由匹配优先级：
		//精确路径匹配（完全匹配）
		//正则表达式路由匹配
		//通配符路由（兜底）

		//handler处理
		//1.静态路由
		sr.registerRoutePattenMode(v.Routers, chains, RouteTypeStatic, v.Handler)
		//2.参数路由
		sr.registerRoutePattenMode(v.Routers, chains, RouteTypeParam, v.Handler)
		//3.通配符路由
		sr.registerRoutePattenMode(v.Routers, chains, RouteTypeWildcard, v.Handler)
		//for _, route := range staticRouteData {
		//	//websocket 特殊处理
		//	if len(route.Methods) == 0 && v.Handler == constants.WebSocket {
		//		sr.Router.GET(route.Path, chains)
		//	} else {
		//		for _, reqMethod := range route.Methods {
		//			if reqMethod == "*" {
		//				sr.registerAllMethods(route.Path, chains)
		//			} else {
		//				sr.Router.Handle(reqMethod, route.Path, chains)
		//			}
		//		}
		//	}
		//}
	}
	apiJson, _ := json.Marshal(apis)
	sr.Md5 = md5.MD5(apiJson)
}

func (sr *DyRouter) registerRoutePattenMode(routes []dynamic.Router, chains fasthttp.RequestHandler, routeType RouteType, webSocketType string) {
	routeDataFilter := utils.Filter(routes, func(u dynamic.Router) bool {
		return ParseRoute(u.Path).Type == routeType
	})

	for _, route := range routeDataFilter {
		//websocket 特殊处理
		if len(route.Methods) == 0 && webSocketType == constants.WebSocket {
			sr.Router.GET(route.Path, chains)
		} else {
			for _, reqMethod := range route.Methods {
				if reqMethod == "*" {
					sr.registerAllMethods(route.Path, chains)
				} else {
					sr.Router.Handle(reqMethod, route.Path, chains)
				}
			}
		}
	}
}

// 处理具体路由
//func setupStaticRouter(routes []dynamic.Router) *fasthttprouter.Router {
//	r := fasthttprouter.New()
//	r.POST("/Account/Login", LoginHandler)
//	// 其他API路由
//	return r
//}

//
//// 处理静态路由
//func setupFileRouter() *fasthttprouter.Router {
//	r := fasthttprouter.New()
//	r.GET("/*filepath", FileServerHandler)
//	return r
//}

func (sr *DyRouter) registerAllMethods(path string, handler fasthttp.RequestHandler) {
	sr.Router.GET(path, handler)
	sr.Router.POST(path, handler)
	sr.Router.PUT(path, handler)
	sr.Router.DELETE(path, handler)
	sr.Router.PATCH(path, handler)
	sr.Router.HEAD(path, handler)
	sr.Router.OPTIONS(path, handler)
}

// Match 动态路由不需要，完全由router去代理
func (sr *DyRouter) Match(_ string) *dynamic.ServiceRoute {
	return nil
}

func (sr *DyRouter) GetMd5() string {
	return sr.Md5
}
