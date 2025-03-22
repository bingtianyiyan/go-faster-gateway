//go:build wireinject
// +build wireinject

package wire

import (
	"go-faster-gateway/internal/handler"
	"go-faster-gateway/internal/repository"
	"go-faster-gateway/internal/server"
	"go-faster-gateway/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var ServerSet = wire.NewSet(server.NewServerHTTP)

var RepositorySet = wire.NewSet(
	repository.NewDb,
	repository.NewRepository,
	repository.NewUserRepository,
)

var ServiceSet = wire.NewSet(
	service.NewService,
	service.NewUserService,
)

var HandlerSet = wire.NewSet(
	handler.NewHandler,
	handler.NewUserHandler,
)

func NewWire() (*gin.Engine, func(), error) {
	panic(wire.Build(
		ServerSet,
		RepositorySet,
		ServiceSet,
		HandlerSet,
	))
}
