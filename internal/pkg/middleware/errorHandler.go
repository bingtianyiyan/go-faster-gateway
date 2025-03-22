package middleware

import (
	"github.com/valyala/fasthttp"
)

func ErrorHandlerMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		next(ctx)
	}
}
