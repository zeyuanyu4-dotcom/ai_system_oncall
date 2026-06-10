package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai_system_oncall/internal/cache"
	"ai_system_oncall/internal/config"
	"ai_system_oncall/internal/database"
	"ai_system_oncall/internal/grpcserver"
	"ai_system_oncall/internal/middleware"
	"ai_system_oncall/internal/router"
	"ai_system_oncall/pkg/jwt"
	"ai_system_oncall/pkg/logger"

	toolingv1 "ai_system_oncall/api/proto/tooling/v1"
	"google.golang.org/grpc"

	"go.uber.org/zap"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "configs/config.yaml", "config file path")
}

func main() {
	flag.Parse()

	// Load config
	if err := config.Init(configPath); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	cfg := config.GetConfig()

	// Initialize logger
	if err := logger.Init(&logger.LogConfig{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	}); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Initialize JWT
	jwt.Init(&cfg.JWT)

	// Initialize database
	if err := database.Init(&cfg.Database); err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize Redis cache
	if err := cache.Init(zap.L()); err != nil {
		logger.Warnf("Redis cache initialization failed (running without cache): %v", err)
	}
	defer func() {
		if c := cache.GetCache(); c != nil {
			c.Close()
		}
	}()

	// Initialize rate limiter
	if err := middleware.InitRateLimiter(zap.L()); err != nil {
		logger.Warnf("Rate limiter initialization failed: %v", err)
	}

	// 初始化所有 service（供 HTTP + gRPC 共用）
	services := router.InitServices()

	// Setup router
	r := router.SetupRouterWithServices(services)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 启动 HTTP server
	go func() {
		logger.Infof("Starting server on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 启动 gRPC server（同进程多协议）
	var grpcSrv *grpcserver.Server
	if cfg.Server.GRPCPort > 0 {
		grpcAddr := fmt.Sprintf(":%d", cfg.Server.GRPCPort)
		logInst := zap.L()
		var err error
		grpcSrv, err = grpcserver.NewServer(grpcAddr, logInst,
			[]grpc.UnaryServerInterceptor{grpcserver.JWTAuthInterceptor(logInst)},
			[]grpc.StreamServerInterceptor{grpcserver.JWTAuthStreamInterceptor(logInst)})
		if err != nil {
			logger.Fatalf("Failed to create gRPC server: %v", err)
		}

		// 注册 ToolingService
		toolingv1.RegisterToolingServiceServer(grpcSrv.GetServer(), grpcserver.NewToolingServer(
			services.ServiceService, services.IssueService,
			services.KnowledgeDocService, services.SimulatedLogService,
		))

		go func() {
			logger.Infof("Starting gRPC server on %s", grpcAddr)
			if err := grpcSrv.Start(); err != nil {
				logger.Errorf("gRPC server error: %v", err)
			}
		}()
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("HTTP server forced to shutdown: %v", err)
	}

	if grpcSrv != nil {
		grpcSrv.Stop(ctx)
	}

	logger.Info("Server exited")
	zap.L().Sync()
}
