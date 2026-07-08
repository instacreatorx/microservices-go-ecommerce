module github.com/ecommerce/notification

go 1.22

require (
	github.com/ecommerce/pkg v0.0.0
	github.com/rabbitmq/amqp091-go v1.9.0
	go.uber.org/zap v1.27.0
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/ecommerce/pkg => ../../pkg
