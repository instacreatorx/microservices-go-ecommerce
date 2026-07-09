package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ecommerce/order/internal/delivery/consumer"
	deliveryhttp "github.com/ecommerce/order/internal/delivery/http"
	"github.com/ecommerce/order/internal/repository"
	"github.com/ecommerce/order/internal/usecase"
	"github.com/ecommerce/pkg/broker"
	"github.com/ecommerce/pkg/config"
	"github.com/ecommerce/pkg/database"
	"github.com/ecommerce/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load("order")
	logger := logger.New(cfg.ServiceName, cfg.LogLevel)
	defer logger.Sync()

	ctx := context.Background()
	db, err := database.NewPostgres(ctx, cfg.Database.DSN(), cfg.Database.MaxOpen, cfg.Database.MaxIdle)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	rabbitMQ, err := broker.NewRabbitMQ(
		cfg.Broker.Host, cfg.Broker.Port,
		cfg.Broker.User, cfg.Broker.Password, cfg.Broker.VHost,
	)
	if err != nil {
		logger.Fatal("failed to connect to broker", zap.Error(err))
	}
	defer rabbitMQ.Close()

	orderRepo := repository.NewPostgresOrderRepository(db.Pool)
	orderUC := usecase.NewOrderUsecase(orderRepo, rabbitMQ)

	paymentConsumer := consumer.NewPaymentConsumer(orderUC, rabbitMQ, logger)
	go paymentConsumer.Start(ctx)

	handler := deliveryhttp.NewOrderHandler(orderUC)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/orders", handler.Create)
	mux.HandleFunc("GET /v1/orders", handler.List)
	mux.HandleFunc("GET /v1/orders/{id}", handler.GetByID)
	mux.HandleFunc("POST /v1/orders/{id}/cancel", handler.Cancel)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mux,
	}

	go func() {
		logger.Info("starting order service", zap.String("port", cfg.Port))
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
