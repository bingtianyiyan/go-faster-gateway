package middleware

import (
	"fmt"
	"github.com/valyala/fasthttp"
)

func init() {
	MiddlewareHandlerList = append(MiddlewareHandlerList, NewLogMiddlewareHandler())
}

type LogMiddlewareHandler struct {
	Name string
}

var _ = MiddlewareEventHandler(&LogMiddlewareHandler{})

func NewLogMiddlewareHandler() *LogMiddlewareHandler {
	return &LogMiddlewareHandler{
		Name: "loghandler",
	}
}

func (m *LogMiddlewareHandler) HandlerName() MiddlewareName {
	return MiddlewareName(m.Name)
}

func (m *LogMiddlewareHandler) HandleEvent(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		fmt.Printf("[%s] %s\n", ctx.Method(), ctx.Path())
		next(ctx)
	}
}

//func LoggingMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
//	return func(ctx *fasthttp.RequestCtx) {
//		fmt.Printf("[%s] %s\n", ctx.Method(), ctx.Path())
//		next(ctx)
//	}
//}
