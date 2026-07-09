package config

import (
	"os"

	"github.com/ecommerce/pkg/config"
)

type Config struct {
	*config.Config
	UserServiceURL    string
	ProductServiceURL string
	OrderServiceURL   string
	PaymentServiceURL string
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func Load() *Config {
	base := config.Load("api-gateway")
	return &Config{
		Config:            base,
		UserServiceURL:    env("USER_SERVICE_URL", "http://user:8081"),
		ProductServiceURL: env("PRODUCT_SERVICE_URL", "http://product:8082"),
		OrderServiceURL:   env("ORDER_SERVICE_URL", "http://order:8083"),
		PaymentServiceURL: env("PAYMENT_SERVICE_URL", "http://payment:8084"),
	}
}
