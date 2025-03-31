package middleware

import (
	"github.com/valyala/fasthttp"
)

func init() {
	MiddlewareHandlerList = append(MiddlewareHandlerList, NewErrorMiddlewareHandler())
}

type ErrorMiddlewareHandler struct {
	Name string
}

var _ = MiddlewareEventHandler(&ErrorMiddlewareHandler{})

func NewErrorMiddlewareHandler() *ErrorMiddlewareHandler {
	return &ErrorMiddlewareHandler{
		Name: "errorhandler",
	}
}

func (m *ErrorMiddlewareHandler) HandlerName() MiddlewareName {
	return MiddlewareName(m.Name)
}

func (m *ErrorMiddlewareHandler) HandleEvent(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		next(ctx)
	}
}

//
//func ErrorHandlerMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
//	return func(ctx *fasthttp.RequestCtx) {
//		next(ctx)
//	}
//}
