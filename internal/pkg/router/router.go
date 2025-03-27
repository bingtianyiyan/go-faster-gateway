package router

import (
	"encoding/json"
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
	BuildRouter([]*dynamic.Service, *middleware.MiddlewareHandler)
	Match(key string) *dynamic.Service
	GetMd5() string
}

// 静态路由处理
type StaticRouter struct {
	router map[string]*dynamic.Service
	mu     sync.RWMutex
}

// NewStaticRouter 静态路由匹配
func NewStaticRouter() *StaticRouter {
	return &StaticRouter{
		router: make(map[string]*dynamic.Service),
	}
}

func (sr *StaticRouter) BuildRouter(apis []*dynamic.Service) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	for _, v := range apis {
		sr.router[v.Service] = v
	}
}

func (sr *StaticRouter) Match(key string) *dynamic.Service {
	if api, ok := sr.router[key]; ok {
		return api
	}
	return nil
}

//type routerHandler func(ctx *fasthttp.RequestCtx, dyConfig *dynamic.Configuration, routerInfo *dynamic.Service)

// DyRouter 动态路由匹配, 将路由规则最终转换成httprouter
type DyRouter struct {
	apis map[string]*dynamic.Service
	//Handler  routerHandler
	ProtocolFactory *protocols.ProtocolFactory
	DyConfig        *dynamic.Configuration
	Router          *fasthttprouter.Router
	Md5             string
	mu              sync.RWMutex
}

func NewDyRouter(protocolFactory *protocols.ProtocolFactory, dyConfig *dynamic.Configuration) *DyRouter {
	return &DyRouter{
		apis:            make(map[string]*dynamic.Service),
		DyConfig:        dyConfig,
		ProtocolFactory: protocolFactory,
	}
}

func (sr *DyRouter) BuildRouter(apis []*dynamic.Service, mwHandler *middleware.MiddlewareHandler) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.Router = fasthttprouter.New()
	for _, v := range apis {
		sr.apis[v.Service] = v
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
			handler.Handle(ctx, sr.DyConfig, temp)
			//sr.Handler(ctx, sr.DyConfig, temp)
		}
		chains := middleware.Chain(h, handlers...)
		reqMethod := v.Routers.Method
		if reqMethod == "*" {
			sr.registerAllMethods(v.Routers.Path, chains)
		} else {
			sr.Router.Handle(v.Routers.Method, v.Routers.Path, chains)
		}
	}
	apiJson, _ := json.Marshal(apis)
	sr.Md5 = md5.MD5(apiJson)
}

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
func (sr *DyRouter) Match(_ string) *dynamic.Service {
	return nil
}

func (sr *DyRouter) GetMd5() string {
	return sr.Md5
}
