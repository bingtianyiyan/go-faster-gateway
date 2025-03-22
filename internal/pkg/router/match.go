package router

import (
	"github.com/valyala/fasthttp"
)

type IMatcher interface {
	Match(port int, request interface{}) (IRouterHandler, bool)
}

type IRouterHandler interface {
	ServeHTTP(ctx *fasthttp.RequestCtx)
}
