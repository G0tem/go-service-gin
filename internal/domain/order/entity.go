package order

import "time"

type Order struct {
	ID        string
	UserID    string
	Amount    float64
	Status    string
	CreatedAt time.Time
}

type CreateOrderCmd struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}
