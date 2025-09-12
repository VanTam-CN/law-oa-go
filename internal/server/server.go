package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"law-oa-go/internal/common"
	"law-oa-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
	router     *gin.Engine
	logger     *slog.Logger
}

func NewServer() *Server {
	// 初始化结构化日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	// 设置Gin模式
	if common.GetEnv("ENVIRONMENT", "development") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	
	// 添加中间件
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.PrometheusMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.RateLimitMiddleware())
	router.Use(middleware.SecurityMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggingMiddleware(logger))

	// 创建HTTP服务器配置
	httpServer := &http.Server{
		Addr:         ":" + common.GetEnv("PORT", "8080"),
		Handler:      router,
		ReadTimeout:  15 * time.Second,    // 读取超时
		WriteTimeout: 15 * time.Second,    // 写入超时
		IdleTimeout:  120 * time.Second,   // 空闲连接超时
		MaxHeaderBytes: 1 << 20,          // 1MB header限制
	}

	return &Server{
		httpServer: httpServer,
		router:     router,
		logger:     logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", "address", s.httpServer.Addr)
	
	// 启动服务器
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	// 等待中断信号来优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.logger.Info("Shutting down server...")

	// 创建关闭上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	s.logger.Info("Server exited")
	return nil
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

func (s *Server) GetLogger() *slog.Logger {
	return s.logger
}