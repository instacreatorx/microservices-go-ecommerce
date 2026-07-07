package main

import (
	"context"

	"github.com/ecommerce/notification/internal/consumer"
	"github.com/ecommerce/pkg/broker"
	"github.com/ecommerce/pkg/config"
	"github.com/ecommerce/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load("notification")
	logger := logger.New(cfg.ServiceName, cfg.LogLevel)
	defer logger.Sync()

	rabbitMQ, err := broker.NewRabbitMQ(
		cfg.Broker.Host, cfg.Broker.Port,
		cfg.Broker.User, cfg.Broker.Password, cfg.Broker.VHost,
	)
	if err != nil {
		logger.Fatal("failed to connect to broker", zap.Error(err))
	}
	defer rabbitMQ.Close()

	ctx := context.Background()
	notificationConsumer := consumer.NewEventConsumer(rabbitMQ, logger)
	notificationConsumer.Start(ctx)
}
