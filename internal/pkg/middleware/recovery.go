package middleware

import (
	"go-faster-gateway/internal/pkg/ecode"
	"go-faster-gateway/pkg/log"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

// RecoveryMiddleware 自定义的recovery中间件
func RecoveryMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if r := recover(); r != nil {
				// 发生panic时的处理逻辑
				log.Log.Error("panic", zap.Any("err", r))
				// 返回500 Internal Server Error给客户端
				ctx.Error(ecode.InternalServerErrorErr.Data(), ecode.InternalServerErrorErr.HttpCode)
			}
		}()
		// 调用下一个处理器
		next(ctx)
	}
}
