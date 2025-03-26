package protocols

import (
	"github.com/valyala/fasthttp"
)

type ProtocolFactory struct {
	handlers []ProtocolHandler
}

func NewProtocolFactory(handlers []ProtocolHandler) *ProtocolFactory {
	return &ProtocolFactory{
		handlers: handlers,
	}
}

func (f *ProtocolFactory) GetHandler(ctx *fasthttp.RequestCtx) ProtocolHandler {
	for _, handler := range f.handlers {
		if handler.Supports(ctx) {
			return handler
		}
	}
	return nil // 没有找到支持的协议
}

func (f *ProtocolFactory) GetDefaultHandler() ProtocolHandler {
	return f.handlers[0]
}

//// 初始化协议工厂
//protocolFactory := factory.NewProtocolFactory()
//if err := protocolFactory.Init(*certFile, *keyFile); err != nil {
//log.Fatalf("Failed to initialize protocol factory: %v", err)
//}
//
//// 创建请求处理器
//handler := func(ctx *fasthttp.RequestCtx) {
//	if handler := protocolFactory.GetHandler(ctx); handler != nil {
//		handler.Handle(ctx)
//	} else {
//		ctx.Error("Unsupported protocol", fasthttp.StatusBadRequest)
//	}
//}
