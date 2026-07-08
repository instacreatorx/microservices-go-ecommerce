module github.com/ecommerce/api-gateway

go 1.22

require (
	github.com/ecommerce/pkg v0.0.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	go.uber.org/zap v1.27.0
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/ecommerce/pkg => ../../pkg
