package middleware

import (
	"fmt"
	"github.com/valyala/fasthttp"
)

func LoggingMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		fmt.Printf("[%s] %s\n", ctx.Method(), ctx.Path())
		next(ctx)
	}
}
