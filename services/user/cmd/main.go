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
	"github.com/ecommerce/user/internal/delivery/http"
	"github.com/ecommerce/user/internal/repository"
	"github.com/ecommerce/user/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load("user")
	logger := logger.New(cfg.ServiceName, cfg.LogLevel)
	defer logger.Sync()

	ctx := context.Background()
	db, err := database.NewPostgres(ctx, cfg.Database.DSN(), cfg.Database.MaxOpen, cfg.Database.MaxIdle)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	userRepo := repository.NewPostgresUserRepository(db.Pool)
	userUC := usecase.NewUserUsecase(userRepo, cfg.JWT.Secret, cfg.JWT.Expiration)
	handler := http.NewUserHandler(userUC)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/auth/register", handler.Register)
	mux.HandleFunc("POST /v1/auth/login", handler.Login)
	mux.HandleFunc("GET /v1/users/{id}", handler.GetByID)
	mux.HandleFunc("PUT /v1/users/{id}", handler.Update)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mux,
	}

	go func() {
		logger.Info("starting user service", zap.String("port", cfg.Port))
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
