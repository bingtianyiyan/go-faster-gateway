package server

import (
	"context"
	"errors"
	"go-faster-gateway/internal/pkg/middleware"
	"go-faster-gateway/pkg/config"
	"go-faster-gateway/pkg/config/static"
	"go-faster-gateway/pkg/log"
	logger2 "go-faster-gateway/pkg/log/logger"
	"go-faster-gateway/pkg/safe"
	"os"
	"os/signal"
	"time"
)

var _ middleware.IServer = (*Server)(nil)

// Server is the reverse-proxy/load-balancer engine.
type Server struct {
	watcher             *config.ConfigurationWatcher
	StaticConfiguration *static.Configuration
	//RouteManager        *router.RouterManager
	//App                 *fasthttp.Server // 代理服务
	signals      chan os.Signal
	stopChan     chan bool
	routinesPool *safe.Pool
}

// Option 参数选项
type Option func(server *Server)

// 配置文件
func WithConfiguration(c *static.Configuration) Option {
	return func(s *Server) {
		s.StaticConfiguration = c
	}
}

// 配置路由信息
//func WithRouteManager(c *router.RouterManager) Option {
//	return func(s *Server) {
//		s.RouteManager = c
//	}
//}

func WithWatch(c *config.ConfigurationWatcher) Option {
	return func(s *Server) {
		s.watcher = c
	}
}

func WithRoutePool(c *safe.Pool) Option {
	return func(s *Server) {
		s.routinesPool = c
	}
}

func WithSignals(c chan os.Signal) Option {
	return func(s *Server) {
		s.signals = c
	}
}

func WithStopChan(c chan bool) Option {
	return func(s *Server) {
		s.stopChan = c
	}
}

// NewServer 初始化服务
func NewServer(options ...Option) *Server {
	srv := &Server{}
	for _, v := range options {
		v(srv)
	}
	srv.configureSignals()
	return srv
}

// Start 启动一个http代理服务
func (s *Server) Start(ctx context.Context) {
	go func() {
		<-ctx.Done()
		slog, _ := logger2.FromContext(ctx)
		slog.Info("I have to go...")
		slog.Info("Stopping server gracefully")
		s.Stop()
	}()

	//s.tcpEntryPoints.Start()
	s.watcher.Start()
	s.routinesPool.GoCtx(s.listenSignals)

	//s.App = &fasthttp.Server{
	//	IdleTimeout:  60 * time.Second,
	//	ReadTimeout:  5 * time.Second,
	//	WriteTimeout: 5 * time.Second,
	//}
	//s.App.Handler = s.RouteManager.Handler
	//
	//// 启动http服务代理
	//if len(s.StaticConfiguration.EntryPoint.Address) > 0 {
	//	go s.startHttpProxy()
	//}
	//
	////// 启动https服务代理
	////if s.GetStaticConfig().Entrypoints.Websecure != nil {
	////	// 启动https服务代理
	////	//go s.startHTTPSProxy()
	////}
	//
	//// 等待关闭
	//return s.WaitStop()
}

// Wait blocks until the server shutdown.
func (s *Server) Wait() {
	<-s.stopChan
}

// Stop stops the server.
func (s *Server) Stop() {
	defer log.Log.Info("Server stopped")

	//s.tcpEntryPoints.Stop()
	s.stopChan <- true
}

// Close destroys the server.
func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	go func(ctx context.Context) {
		<-ctx.Done()
		if errors.Is(ctx.Err(), context.Canceled) {
			return
		} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			panic("Timeout while stopping traefik, killing instance ✝")
		}
	}(ctx)

	s.routinesPool.Stop()

	signal.Stop(s.signals)
	close(s.signals)

	close(s.stopChan)

	cancel()
}

// startHttpProxy 启动http代理
func (s *Server) startHttpProxy() error {
	//err := s.App.ListenAndServe(fmt.Sprintf("%s:%d", s.StaticConfiguration.EntryPoint.Address, s.StaticConfiguration.EntryPoint.Port))
	//if err != nil {
	//	log.Log.WithError(err).Error("failed to start http gateway")
	//}
	//return err
	return nil
}

// startHTTPSProxy 启动https代理
func (s *Server) startHTTPSProxy() error {
	//err := s.App.ListenAndServeTLS(s.GetStaticConfig().Entrypoints.Websecure.Addr, s.GetStaticConfig().TLS.CertFile, s.GetStaticConfig().TLS.KeyFile)
	//if err != nil {
	//	log.Logger.Error("failed to start https gateway", zap.Error(err))
	//}
	//return err
	return nil
}

func (s *Server) configureSignals() {}

func (s *Server) listenSignals(ctx context.Context) {}
