package middleware

import (
	"github.com/valyala/fasthttp"
)

// MiddlewareName is the type of an event handler, used as its unique identifier.
type MiddlewareName string

// String returns the string representation of an event handler type.
func (ht MiddlewareName) String() string {
	return string(ht)
}

// MiddlewareEventHandler is a handler
type MiddlewareEventHandler interface {
	// HandlerName is the name of the handler.
	HandlerName() MiddlewareName
	// 事件实际处理 HandleEvent handles an event.
	HandleEvent(next fasthttp.RequestHandler) fasthttp.RequestHandler
}

//// MiddlewareEventHandlerFunc is a function that can be used as a event handler.
//type MiddlewareEventHandlerFunc func(next fasthttp.RequestHandler) fasthttp.RequestHandler
//
//// HandleEvent implements the HandleEvent method of the MiddlewareEventHandler.
//func (f MiddlewareEventHandlerFunc) HandleEvent(next fasthttp.RequestHandler) fasthttp.RequestHandler {
//	return f(next)
//}
//
//// HandlerName implements the HandlerName method of the MiddlewareEventHandler by returning
//func (f MiddlewareEventHandlerFunc) HandlerName() MiddlewareName {
//	return MiddlewareName(middlewareName)
//}
