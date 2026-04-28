//go:build !skip_otel

package otel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func SetBusinessAttrs(ctx context.Context, userID, status string, amount float64) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		attribute.String("business.user.id", userID),
		attribute.String("business.order.status", status),
		attribute.Float64("business.order.amount", amount),
		attribute.String("service.name", "order-service"),
	)
}
