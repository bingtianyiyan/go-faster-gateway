package server

import (
	"go-faster-gateway/internal/handler"
	"go-faster-gateway/internal/middleware"
	"go-faster-gateway/pkg/helper/resp"

	"github.com/gin-gonic/gin"
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
