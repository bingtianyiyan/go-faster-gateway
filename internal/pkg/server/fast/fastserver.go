package fast

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"go-faster-gateway/pkg/config/static"
	"go-faster-gateway/pkg/log"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HttpServer struct {
	staticConfig *static.Configuration
	appServer    *fasthttp.Server // 代理服务
	handler      func(ctx *fasthttp.RequestCtx)
}

func NewHttpServer(staticConfig *static.Configuration,
	handler func(ctx *fasthttp.RequestCtx)) *HttpServer {
	return &HttpServer{
		staticConfig: staticConfig,
		handler:      handler,
		appServer: &fasthttp.Server{
			IdleTimeout:  60 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
}

func (s *HttpServer) Start() {
	// 启动http服务代理
	if len(s.staticConfig.EntryPoint.Address) > 0 {
		go s.startHttpProxy()
	}

	//// 启动https服务代理
	//if s.GetStaticConfig().Entrypoints.Websecure != nil {
	//	// 启动https服务代理
	//	//go s.startHTTPSProxy()
	//}
}

func (s *HttpServer) Stop() error {
	sign := make(chan os.Signal)
	signal.Notify(sign, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)
	select {
	case ss := <-sign:
		log.Log.Info("gateway server receive got signal", zap.Any("signal", ss))
		err := s.appServer.Shutdown()
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 5)
		log.Log.Info("gateway http shutdown")
	}
	return nil
}

// startHttpProxy 启动http代理
func (s *HttpServer) startHttpProxy() error {
	err := s.appServer.ListenAndServe(fmt.Sprintf("%s:%d", s.staticConfig.EntryPoint.Address, s.staticConfig.EntryPoint.Port))
	if err != nil {
		log.Log.WithError(err).Error("failed to start http gateway")
	}
	return err
}

// startHTTPSProxy 启动https代理
func (s *HttpServer) startHTTPSProxy() error {
	//err := s.App.ListenAndServeTLS(s.GetStaticConfig().Entrypoints.Websecure.Addr, s.GetStaticConfig().TLS.CertFile, s.GetStaticConfig().TLS.KeyFile)
	//if err != nil {
	//	log.Logger.Error("failed to start https gateway", zap.Error(err))
	//}
	//return err
	return nil
}

func (s *HttpServer) SwitchRouter(handler func(ctx *fasthttp.RequestCtx)) {
	s.appServer.Handler = handler
}
