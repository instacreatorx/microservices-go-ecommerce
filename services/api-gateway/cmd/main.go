package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ecommerce/api-gateway/internal/config"
	"github.com/ecommerce/api-gateway/internal/router"
	"github.com/ecommerce/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	logger := logger.New(cfg.ServiceName, cfg.LogLevel)
	defer logger.Sync()

	logger.Info("starting API gateway", zap.String("port", cfg.Port))

	r := router.New(cfg, logger)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	if err := srv.Close(); err != nil {
		log.Fatal("failed to close server:", err)
	}
}
