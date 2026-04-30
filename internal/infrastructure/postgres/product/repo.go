package product

import (
	"context"
	"time"

	"github.com/G0tem/go-service-gin/internal/domain/product"
	"github.com/G0tem/go-service-gin/internal/domain/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductRepo — PostgreSQL реализация репозитория товаров
type ProductRepo struct {
	pool *pgxpool.Pool
}

// NewProductRepo создает новый репозиторий товаров
func NewProductRepo(pool *pgxpool.Pool) ports.ProductRepository {
	return &ProductRepo{pool: pool}
}

// Create сохраняет новый товар в БД
func (r *ProductRepo) Create(ctx context.Context, p *product.Product) error {
	query := `
		INSERT INTO products (id, name, description, price, stock, category_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		p.ID, p.Name, p.Description, p.Price, p.Stock, p.CategoryID, p.CreatedAt, p.UpdatedAt)
	return err
}

// GetByID возвращает товар по ID
func (r *ProductRepo) GetByID(ctx context.Context, id string) (*product.Product, error) {
	query := `
		SELECT id, name, description, price, stock, category_id, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	p := &product.Product{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CategoryID, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, product.ErrProductNotFound
	}
	return p, err
}

// Update обновляет существующий товар
func (r *ProductRepo) Update(ctx context.Context, p *product.Product) error {
	query := `
		UPDATE products
		SET name = $2, description = $3, price = $4, stock = $5, category_id = $6, updated_at = $7
		WHERE id = $1
	`
	result, err := r.pool.Exec(ctx, query,
		p.ID, p.Name, p.Description, p.Price, p.Stock, p.CategoryID, p.UpdatedAt)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return product.ErrProductNotFound
	}
	return nil
}

// Delete удаляет товар по ID
func (r *ProductRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return product.ErrProductNotFound
	}
	return nil
}

// List возвращает список товаров с пагинацией
func (r *ProductRepo) List(ctx context.Context, limit, offset int) ([]*product.Product, error) {
	query := `
		SELECT id, name, description, price, stock, category_id, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*product.Product
	for rows.Next() {
		p := &product.Product{}
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CategoryID, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// ensure interfaces
var _ ports.ProductRepository = (*ProductRepo)(nil)
