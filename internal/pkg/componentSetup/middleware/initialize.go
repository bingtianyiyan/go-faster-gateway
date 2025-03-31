package middleware

import "go-faster-gateway/internal/pkg/middleware"

func SetUpMiddleware() *middleware.MiddlewareManager {
	middlewareManager := middleware.NewMiddlewareManager()
	for _, v := range middleware.MiddlewareHandlerList {
		middlewareManager.Add(v.HandlerName(), v.HandleEvent)
	}
	return middlewareManager
}
