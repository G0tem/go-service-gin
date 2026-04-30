package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/G0tem/go-service-gin/internal/domain/order"
	"github.com/G0tem/go-service-gin/internal/domain/ports"
	"github.com/G0tem/go-service-gin/internal/otel"
	"github.com/google/uuid"
)

type CreateOrderHandler struct {
	repo    ports.OrderRepository
	pub     ports.EventPublisher
	metrics *otel.BusinessMetrics
}

func NewCreateOrderHandler(repo ports.OrderRepository, pub ports.EventPublisher, m *otel.BusinessMetrics) *CreateOrderHandler {
	return &CreateOrderHandler{repo: repo, pub: pub, metrics: m}
}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd order.CreateOrderCmd) error {
	o := order.Order{
		ID:     uuid.New().String(),
		UserID: cmd.UserID,
		Amount: cmd.Amount,
		Status: "pending",
	}

	if err := h.repo.Create(ctx, o); err != nil {
		return fmt.Errorf("create order in db: %w", err)
	}

	// 📊 Метрики
	h.metrics.OrdersCreated.Add(ctx, 1)
	h.metrics.OrderAmount.Record(ctx, cmd.Amount)

	// 🏷️ Бизнес-атрибуты в трейс
	otel.SetBusinessAttrs(ctx, cmd.UserID, o.Status, cmd.Amount)

	payload, _ := json.Marshal(o)
	if err := h.pub.Publish(ctx, "order.created", payload); err != nil {
		// В production: Outbox / компенсация
	}
	return nil
}
