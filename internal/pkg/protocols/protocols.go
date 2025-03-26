package protocols

import (
	"github.com/valyala/fasthttp"
	"go-faster-gateway/pkg/config/dynamic"
)

// 请求的协议处理
type ProtocolHandler interface {
	Handle(ctx *fasthttp.RequestCtx, dyConfig *dynamic.Configuration, routerInfo *dynamic.Service)
	Supports(ctx *fasthttp.RequestCtx) bool
}
