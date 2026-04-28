package handler

import (
	"net/http"

	"github.com/G0tem/go-service-gin/internal/domain/order"
	"github.com/G0tem/go-service-gin/internal/http/middleware"
	usecase "github.com/G0tem/go-service-gin/internal/usecases"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	uc *usecase.CreateOrderHandler
}

func NewOrderHandler(uc *usecase.CreateOrderHandler) *OrderHandler {
	return &OrderHandler{uc: uc}
}

// @Summary Create a new order
// @Description Creates a new order and publishes an event to RabbitMQ
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body order.CreateOrderCmd true "Order creation payload"
// @Success 201 {object} map[string]string "status: order created & event published"
// @Failure 400 {object} map[string]string "invalid payload"
// @Failure 401 {object} map[string]string "unauthorized"
// @Failure 500 {object} map[string]string "internal error"
// @Router /orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth_claims_missing"})
		return
	}

	var cmd order.CreateOrderCmd
	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	// 🔐 Переопределяем UserID из токена (защита от подмены)
	cmd.UserID = claims.UserID

	if err := h.uc.Handle(c.Request.Context(), cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "order_created"})
}
