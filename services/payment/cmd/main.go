package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ecommerce/payment/internal/delivery/consumer"
	deliveryhttp "github.com/ecommerce/payment/internal/delivery/http"
	"github.com/ecommerce/payment/internal/repository"
	"github.com/ecommerce/payment/internal/usecase"
	"github.com/ecommerce/pkg/broker"
	"github.com/ecommerce/pkg/config"
	"github.com/ecommerce/pkg/database"
	"github.com/ecommerce/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load("payment")
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

	paymentRepo := repository.NewPostgresPaymentRepository(db.Pool)
	paymentUC := usecase.NewPaymentUsecase(paymentRepo, rabbitMQ)

	orderConsumer := consumer.NewOrderConsumer(paymentUC, rabbitMQ, logger)
	go orderConsumer.Start(ctx)

	handler := deliveryhttp.NewPaymentHandler(paymentUC)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/payments", handler.Process)
	mux.HandleFunc("GET /v1/payments/{id}", handler.GetByID)
	mux.HandleFunc("POST /v1/payments/{id}/refund", handler.Refund)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mux,
	}

	go func() {
		logger.Info("starting payment service", zap.String("port", cfg.Port))
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
