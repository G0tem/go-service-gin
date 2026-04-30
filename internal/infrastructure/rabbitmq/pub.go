package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (c *Client) Publish(ctx context.Context, routingKey string, payload []byte) error {
	return c.channel.PublishWithContext(ctx, "", routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
	})
}
