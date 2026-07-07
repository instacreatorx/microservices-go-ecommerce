package router

import (
	"net/http"

	"github.com/ecommerce/api-gateway/internal/config"
	"github.com/ecommerce/api-gateway/internal/handler"
	"github.com/ecommerce/api-gateway/internal/middleware"
	"github.com/ecommerce/pkg/logger"
	"go.uber.org/zap"
)

func New(cfg *config.Config, logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()

	authHandler := handler.NewAuthHandler(cfg.UserServiceURL, logger)
	productHandler := handler.NewProductHandler(cfg.ProductServiceURL, logger)
	orderHandler := handler.NewOrderHandler(cfg.OrderServiceURL, logger)

	authMW := middleware.NewAuthMiddleware(cfg.JWT.Secret)

	mux.HandleFunc("POST /v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)

	mux.HandleFunc("GET /v1/products", productHandler.List)
	mux.HandleFunc("GET /v1/products/{id}", productHandler.GetByID)

	mux.Handle("POST /v1/orders", authMW.Wrap(orderHandler.Create))
	mux.Handle("GET /v1/orders", authMW.Wrap(orderHandler.List))
	mux.Handle("GET /v1/orders/{id}", authMW.Wrap(orderHandler.GetByID))

	return middleware.Chain(mux, logger)
}
