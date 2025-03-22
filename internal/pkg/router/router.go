package router

import (
	"encoding/json"
	"go-faster-gateway/internal/pkg/middleware"
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

type routerHandler func(ctx *fasthttp.RequestCtx, routerInfo *dynamic.Service)

// DyRouter 动态路由匹配, 将路由规则最终转换成httprouter
type DyRouter struct {
	apis    map[string]*dynamic.Service
	Handler routerHandler
	Router  *fasthttprouter.Router
	Md5     string
	mu      sync.RWMutex
}

func NewDyRouter(handler routerHandler) *DyRouter {
	return &DyRouter{
		apis:    make(map[string]*dynamic.Service),
		Handler: handler,
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
			sr.Handler(ctx, temp)
		}
		chains := middleware.Chain(h, handlers...)
		sr.Router.Handle(v.Routers.Method, v.Routers.Path, chains)
	}
	apiJson, _ := json.Marshal(apis)
	sr.Md5 = md5.MD5(apiJson)
}

// Match 动态路由不需要，完全由router去代理
func (sr *DyRouter) Match(_ string) *dynamic.Service {
	return nil
}

func (sr *DyRouter) GetMd5() string {
	return sr.Md5
}
