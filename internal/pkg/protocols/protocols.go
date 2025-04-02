package protocols

import (
	"github.com/valyala/fasthttp"
	"go-faster-gateway/pkg/config/dynamic"
)

// 请求的协议处理
type ProtocolHandler interface {
	Handle(ctx *fasthttp.RequestCtx, serviceRoute *dynamic.ServiceRoute, routeInfo dynamic.Router)
	Supports(ctx *fasthttp.RequestCtx) bool
}
