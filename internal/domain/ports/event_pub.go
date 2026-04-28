package ports

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, routingKey string, payload []byte) error
}
