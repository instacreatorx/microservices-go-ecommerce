package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ecommerce/pkg/config"
	"github.com/ecommerce/pkg/database"
	"github.com/ecommerce/pkg/logger"
	"github.com/ecommerce/product/internal/delivery/http"
	"github.com/ecommerce/product/internal/repository"
	"github.com/ecommerce/product/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load("product")
	logger := logger.New(cfg.ServiceName, cfg.LogLevel)
	defer logger.Sync()

	ctx := context.Background()
	db, err := database.NewPostgres(ctx, cfg.Database.DSN(), cfg.Database.MaxOpen, cfg.Database.MaxIdle)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	productRepo := repository.NewPostgresProductRepository(db.Pool)
	productUC := usecase.NewProductUsecase(productRepo)
	handler := http.NewProductHandler(productUC)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/products", handler.Create)
	mux.HandleFunc("GET /v1/products", handler.List)
	mux.HandleFunc("GET /v1/products/{id}", handler.GetByID)
	mux.HandleFunc("PUT /v1/products/{id}", handler.Update)
	mux.HandleFunc("DELETE /v1/products/{id}", handler.Delete)
	mux.HandleFunc("PATCH /v1/products/{id}/stock", handler.UpdateStock)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mux,
	}

	go func() {
		logger.Info("starting product service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down")
	srv.Close()
}
