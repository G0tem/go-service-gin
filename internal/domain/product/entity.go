package product

import (
	"errors"
	"time"
)

// Ошибки доменного слоя
var (
	ErrProductNotFound      = errors.New("product not found")
	ErrInvalidPrice         = errors.New("invalid price: must be positive")
	ErrInvalidName          = errors.New("invalid name: cannot be empty")
	ErrInsufficientStock    = errors.New("insufficient stock")
)

// Product — агрегат товара
type Product struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Stock       int
	CategoryID  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewProduct — фабрика для создания товара с валидацией инвариантов
func NewProduct(name string, price float64) (*Product, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	if price <= 0 {
		return nil, ErrInvalidPrice
	}

	now := time.Now()
	return &Product{
		Name:      name,
		Price:     price,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateInfo обновляет информацию о товаре
func (p *Product) UpdateInfo(name, description string) error {
	if name == "" {
		return ErrInvalidName
	}
	p.Name = name
	p.Description = description
	p.UpdatedAt = time.Now()
	return nil
}

// UpdatePrice обновляет цену с валидацией
func (p *Product) UpdatePrice(price float64) error {
	if price <= 0 {
		return ErrInvalidPrice
	}
	p.Price = price
	p.UpdatedAt = time.Now()
	return nil
}

// ReserveStock резервирует товар (уменьшает доступное количество)
func (p *Product) ReserveStock(qty int) error {
	if qty <= 0 {
		return errors.New("quantity must be positive")
	}
	if p.Stock < qty {
		return ErrInsufficientStock
	}
	p.Stock -= qty
	p.UpdatedAt = time.Now()
	return nil
}

// AddStock пополняет склад
func (p *Product) AddStock(qty int) error {
	if qty <= 0 {
		return errors.New("quantity must be positive")
	}
	p.Stock += qty
	p.UpdatedAt = time.Now()
	return nil
}
