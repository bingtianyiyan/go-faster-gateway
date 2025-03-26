package server

import (
	"context"
	"errors"
	"go-faster-gateway/internal/pkg/middleware"
	"go-faster-gateway/internal/pkg/server/fast"
	"go-faster-gateway/pkg/config"
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
	configManager  *config.ConfigurationManager
	fastHttpServer *fast.HttpServer
	signals        chan os.Signal
	stopChan       chan bool
	routinesPool   *safe.Pool
}

// Option 参数选项
type Option func(server *Server)

// 配置文件管理
func WithConfigurationManager(c *config.ConfigurationManager) Option {
	return func(s *Server) {
		s.configManager = c
	}
}

func WithFastHttpServer(c *fast.HttpServer) Option {
	return func(s *Server) {
		s.fastHttpServer = c
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

	s.fastHttpServer.Start()
	s.configManager.GetWatcher().Start()
	s.routinesPool.GoCtx(s.listenSignals)
}

// Wait blocks until the server shutdown.
func (s *Server) Wait() {
	<-s.stopChan
}

// Stop stops the server.
func (s *Server) Stop() {
	defer log.Log.Info("Server stopped")

	s.fastHttpServer.Stop()
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

func (s *Server) configureSignals() {}

func (s *Server) listenSignals(ctx context.Context) {}
