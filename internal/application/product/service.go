package product

import (
	"context"

	"github.com/G0tem/go-service-gin/internal/domain/product"
	"github.com/G0tem/go-service-gin/internal/domain/ports"
	"github.com/google/uuid"
)

// CreateProductCmd — команда для создания товара
type CreateProductCmd struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"min=0"`
	CategoryID  string  `json:"category_id"`
}

// UpdateProductCmd — команда для обновления товара
type UpdateProductCmd struct {
	ID          string  `json:"-" uri:"id" binding:"required"`
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
}

// ProductService — сервис для управления товарами
type ProductService struct {
	repo ports.ProductRepository
}

// NewProductService создает новый сервис товаров
func NewProductService(repo ports.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// Create создает новый товар
func (s *ProductService) Create(ctx context.Context, cmd CreateProductCmd) (*product.Product, error) {
	p, err := product.NewProduct(cmd.Name, cmd.Price)
	if err != nil {
		return nil, err
	}

	p.ID = uuid.New().String()
	p.Description = cmd.Description
	p.Stock = cmd.Stock
	p.CategoryID = cmd.CategoryID

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

// GetByID возвращает товар по ID
func (s *ProductService) GetByID(ctx context.Context, id string) (*product.Product, error) {
	return s.repo.GetByID(ctx, id)
}

// Update обновляет информацию о товаре
func (s *ProductService) Update(ctx context.Context, cmd UpdateProductCmd) (*product.Product, error) {
	p, err := s.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if err := p.UpdateInfo(cmd.Name, cmd.Description); err != nil {
		return nil, err
	}

	if err := p.UpdatePrice(cmd.Price); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

// Delete удаляет товар
func (s *ProductService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// List возвращает список товаров с пагинацией
func (s *ProductService) List(ctx context.Context, limit, offset int) ([]*product.Product, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

// AddStock пополняет склад товара
func (s *ProductService) AddStock(ctx context.Context, id string, qty int) error {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := p.AddStock(qty); err != nil {
		return err
	}

	return s.repo.Update(ctx, p)
}

// ReserveStock резервирует товар
func (s *ProductService) ReserveStock(ctx context.Context, id string, qty int) error {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := p.ReserveStock(qty); err != nil {
		return err
	}

	return s.repo.Update(ctx, p)
}
