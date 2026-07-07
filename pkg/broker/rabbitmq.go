package broker

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(host, port, user, password, vhost string) (*RabbitMQ, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, password, host, port, vhost)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("channel: %w", err)
	}

	return &RabbitMQ{Conn: conn, Channel: ch}, nil
}

func (r *RabbitMQ) DeclareExchange(name, kind string) error {
	return r.Channel.ExchangeDeclare(
		name,
		kind,
		true,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) DeclareQueue(name string, args amqp.Table) (amqp.Queue, error) {
	return r.Channel.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		args,
	)
}

func (r *RabbitMQ) BindQueue(queue, key, exchange string) error {
	return r.Channel.QueueBind(
		queue,
		key,
		exchange,
		false,
		nil,
	)
}

func (r *RabbitMQ) Publish(exchange, key string, body []byte) error {
	return r.Channel.Publish(
		exchange,
		key,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (r *RabbitMQ) Consume(queue string) (<-chan amqp.Delivery, error) {
	return r.Channel.Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Conn != nil {
		r.Conn.Close()
	}
}
