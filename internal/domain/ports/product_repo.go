package ports

import (
	"context"

	"github.com/G0tem/go-service-gin/internal/domain/product"
)

// ProductRepository — интерфейс репозитория для работы с товарами
type ProductRepository interface {
	Create(ctx context.Context, p *product.Product) error
	GetByID(ctx context.Context, id string) (*product.Product, error)
	Update(ctx context.Context, p *product.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*product.Product, error)
}
