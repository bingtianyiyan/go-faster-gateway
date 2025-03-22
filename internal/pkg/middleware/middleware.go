package middleware

import (
	"sync"

	"github.com/valyala/fasthttp"
)

type MiddlewareFunc func(fasthttp.RequestHandler) fasthttp.RequestHandler

func Chain(handler fasthttp.RequestHandler, middlewares ...MiddlewareFunc) fasthttp.RequestHandler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

type MiddlewareHandler struct {
	Handler map[string]MiddlewareFunc

	mu sync.Mutex
}

type IServer interface {
}
