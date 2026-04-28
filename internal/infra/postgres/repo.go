package postgres

import (
	"context"

	"github.com/G0tem/go-service-gin/internal/domain/order"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo { return &OrderRepo{pool: pool} }

func (r *OrderRepo) Create(ctx context.Context, o order.Order) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO orders (id, user_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5)`, o.ID, o.UserID, o.Amount, o.Status, o.CreatedAt)
	return err
}
