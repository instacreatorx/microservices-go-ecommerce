package config

import "github.com/ecommerce/pkg/config"

type Config struct {
	*config.Config
	UserServiceURL    string
	ProductServiceURL string
	OrderServiceURL   string
	PaymentServiceURL string
}

func Load() *Config {
	base := config.Load("api-gateway")
	return &Config{
		Config:            base,
		UserServiceURL:    config.GetEnv("USER_SERVICE_URL", "http://user:8081"),
		ProductServiceURL: config.GetEnv("PRODUCT_SERVICE_URL", "http://product:8082"),
		OrderServiceURL:   config.GetEnv("ORDER_SERVICE_URL", "http://order:8083"),
		PaymentServiceURL: config.GetEnv("PAYMENT_SERVICE_URL", "http://payment:8084"),
	}
}
