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

	"ai_system_oncall/internal/config"
	"ai_system_oncall/internal/database"
	"ai_system_oncall/internal/router"
	"ai_system_oncall/pkg/jwt"
	"ai_system_oncall/pkg/logger"

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

	// Setup router
	r := router.SetupRouter()

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Infof("Starting server on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
	zap.L().Sync()
}
