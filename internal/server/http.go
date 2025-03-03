package server

import (
	"github.com/gin-gonic/gin"
	"go-faster-gateway/internal/handler"
	"go-faster-gateway/internal/middleware"
	"go-faster-gateway/pkg/helper/resp"
)

func NewServerHTTP(
	userHandler *handler.UserHandler,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(
		middleware.CORSMiddleware(),
	)
	r.GET("/", func(ctx *gin.Context) {
		resp.HandleSuccess(ctx, map[string]interface{}{
			"say": "Hi Nunu!",
		})
	})
	r.GET("/user", userHandler.GetUserById)

	return r
}
