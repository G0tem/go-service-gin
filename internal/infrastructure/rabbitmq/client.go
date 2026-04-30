package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wzy9607/amqp091otel"

	"go.opentelemetry.io/otel"
)

type Client struct {
	conn    *amqp.Connection
	channel *amqp091otel.Channel
}

func NewClient(ctx context.Context, url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}

	// 👇 NewChannel возвращает (*Channel, error) и требует URL + опции
	otelChan, err := amqp091otel.NewChannel(
		ch,
		url, // RabbitMQ URL для атрибутов спана
		amqp091otel.WithTracerProvider(otel.GetTracerProvider()),
		// amqp091otel.WithPropagators(otel.GetTextMapPropagator()), // опционально
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("wrap channel with otel: %w", err)
	}

	return &Client{conn: conn, channel: otelChan}, nil
}

func (c *Client) Close() error {
	var err error
	if c.channel != nil {
		err = c.channel.Close()
	}
	if c.conn != nil {
		if closeErr := c.conn.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}
