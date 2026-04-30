package handler

import (
	"net/http"
	"strconv"

	"github.com/G0tem/go-service-gin/internal/application/product"
	"github.com/gin-gonic/gin"
)

// ProductHandler — HTTP обработчик для работы с товарами
type ProductHandler struct {
	service *product.ProductService
}

// NewProductHandler создает новый HTTP обработчик
func NewProductHandler(service *product.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// Create godoc
// @Summary Create a new product
// @Description Creates a new product in the catalog
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body product.CreateProductCmd true "Product creation payload"
// @Success 201 {object} map[string]interface{} "created product"
// @Failure 400 {object} map[string]string "invalid payload"
// @Failure 401 {object} map[string]string "unauthorized"
// @Failure 500 {object} map[string]string "internal error"
// @Router /products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var cmd product.CreateProductCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "details": err.Error()})
		return
	}

	p, err := h.service.Create(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          p.ID,
		"name":        p.Name,
		"description": p.Description,
		"price":       p.Price,
		"stock":       p.Stock,
		"category_id": p.CategoryID,
		"created_at":  p.CreatedAt,
	})
}

// GetByID godoc
// @Summary Get product by ID
// @Description Returns a single product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]interface{} "product details"
// @Failure 400 {object} map[string]string "invalid id"
// @Failure 404 {object} map[string]string "not found"
// @Failure 401 {object} map[string]string "unauthorized"
// @Router /products/{id} [get]
func (h *ProductHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id_required"})
		return
	}

	p, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product_not_found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          p.ID,
		"name":        p.Name,
		"description": p.Description,
		"price":       p.Price,
		"stock":       p.Stock,
		"category_id": p.CategoryID,
		"created_at":  p.CreatedAt,
		"updated_at":  p.UpdatedAt,
	})
}

// Update godoc
// @Summary Update an existing product
// @Description Updates product information
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param product body product.UpdateProductCmd true "Product update payload"
// @Success 200 {object} map[string]interface{} "updated product"
// @Failure 400 {object} map[string]string "invalid payload"
// @Failure 404 {object} map[string]string "not found"
// @Failure 401 {object} map[string]string "unauthorized"
// @Router /products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id_required"})
		return
	}

	var cmd product.UpdateProductCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "details": err.Error()})
		return
	}
	cmd.ID = id

	p, err := h.service.Update(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product_not_found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          p.ID,
		"name":        p.Name,
		"description": p.Description,
		"price":       p.Price,
		"updated_at":  p.UpdatedAt,
	})
}

// Delete godoc
// @Summary Delete a product
// @Description Deletes a product by its ID
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 204 {string} string "no content"
// @Failure 400 {object} map[string]string "invalid id"
// @Failure 404 {object} map[string]string "not found"
// @Failure 401 {object} map[string]string "unauthorized"
// @Router /products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id_required"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product_not_found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// List godoc
// @Summary List products with pagination
// @Description Returns a paginated list of products
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Items per page" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{} "list of products"
// @Failure 400 {object} map[string]string "invalid parameters"
// @Failure 401 {object} map[string]string "unauthorized"
// @Router /products [get]
func (h *ProductHandler) List(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_offset"})
		return
	}

	products, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]map[string]interface{}, len(products))
	for i, p := range products {
		result[i] = map[string]interface{}{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
			"price":       p.Price,
			"stock":       p.Stock,
			"category_id": p.CategoryID,
			"created_at":  p.CreatedAt,
			"updated_at":  p.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items": result,
		"total": len(products),
		"limit": limit,
		"offset": offset,
	})
}

// AddStock godoc
// @Summary Add stock to a product
// @Description Increases the stock quantity of a product
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param qty body map[string]int true "Quantity to add"
// @Success 200 {object} map[string]interface{} "updated stock"
// @Failure 400 {object} map[string]string "invalid payload"
// @Failure 404 {object} map[string]string "not found"
// @Failure 401 {object} map[string]string "unauthorized"
// @Router /products/{id}/stock/add [post]
func (h *ProductHandler) AddStock(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id_required"})
		return
	}

	var body struct {
		Qty int `json:"qty" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "details": err.Error()})
		return
	}

	if err := h.service.AddStock(c.Request.Context(), id, body.Qty); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product_not_found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "stock_added", "qty": body.Qty})
}

// ReserveStock godoc
// @Summary Reserve stock for a product
// @Description Reserves (decreases) the stock quantity of a product
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param qty body map[string]int true "Quantity to reserve"
// @Success 200 {object} map[string]interface{} "reserved stock"
// @Failure 400 {object} map[string]string "invalid payload"
// @Failure 404 {object} map[string]string "not found"
// @Failure 409 {object} map[string]string "insufficient stock"
// @Failure 401 {object} map[string]string "unauthorized"
// @Router /products/{id}/stock/reserve [post]
func (h *ProductHandler) ReserveStock(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id_required"})
		return
	}

	var body struct {
		Qty int `json:"qty" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "details": err.Error()})
		return
	}

	if err := h.service.ReserveStock(c.Request.Context(), id, body.Qty); err != nil {
		if err.Error() == "insufficient stock" {
			c.JSON(http.StatusConflict, gin.H{"error": "insufficient_stock"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "product_not_found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "stock_reserved", "qty": body.Qty})
}
