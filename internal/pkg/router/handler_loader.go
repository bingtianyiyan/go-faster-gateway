package router

import (
	"fmt"
	"github.com/valyala/fasthttp"
)

type HandlerLoader struct {
	handlers map[string]fasthttp.RequestHandler
}

func NewHandlerLoader() *HandlerLoader {
	loader := &HandlerLoader{
		handlers: make(map[string]fasthttp.RequestHandler),
	}

	// 注册内置处理器
	loader.Register("api.v1.userList", apiV1UserListHandler)
	loader.Register("api.v1.userDetail", apiV1UserDetailHandler)
	loader.Register("static.fileHandler", staticFileHandler)
	loader.Register("admin.dashboard", adminDashboardHandler)
	loader.Register("admin.assets", adminAssetsHandler)
	loader.Register("default.handler", defaultHandler)

	return loader
}

func (l *HandlerLoader) Register(name string, handler fasthttp.RequestHandler) {
	l.handlers[name] = handler
}

func (l *HandlerLoader) Load(name string) (fasthttp.RequestHandler, error) {
	if handler, ok := l.handlers[name]; ok {
		return handler, nil
	}
	return nil, fmt.Errorf("handler %s not found", name)
}

// 示例处理器实现
func apiV1UserListHandler(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("API v1 User List")
}

func apiV1UserDetailHandler(ctx *fasthttp.RequestCtx) {
	id := ctx.UserValue("id").(string)
	ctx.WriteString("API v1 User Detail: " + id)
}

func staticFileHandler(ctx *fasthttp.RequestCtx) {
	filepath := ctx.UserValue("filepath").(string)
	ctx.WriteString("Static File: " + filepath)
}

func adminDashboardHandler(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("Admin Dashboard")
}

func adminAssetsHandler(ctx *fasthttp.RequestCtx) {
	assetpath := ctx.UserValue("assetpath").(string)
	ctx.WriteString("Admin Asset: " + assetpath)
}

func defaultHandler(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("Default Handler")
}
