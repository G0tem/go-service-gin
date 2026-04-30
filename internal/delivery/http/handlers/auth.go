package handler

import (
	"net/http"

	"github.com/G0tem/go-service-gin/internal/domain/ports"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	tm ports.TokenManager
	// userRepo ports.UserRepository // TODO здесь будет проверка email/password + bcrypt
}

func NewAuthHandler(tm ports.TokenManager) *AuthHandler {
	return &AuthHandler{tm: tm}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// @Summary User login
// @Description Authenticate user and return JWT access token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User credentials"
// @Success 200 {object} map[string]interface{} "access_token, token_type, expires_in"
// @Failure 400 {object} map[string]string "invalid payload"
// @Failure 401 {object} map[string]string "invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload"})
		return
	}

	// 🔄 TODO: user, err := h.userRepo.FindByEmail(ctx, req.Email)
	// bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(req.Password))
	// эмулируем успешную проверку
	userID := "usr_" + req.Email[:3] // mock
	role := "user"

	token, err := h.tm.GenerateToken(c.Request.Context(), ports.AuthClaims{
		UserID: userID,
		Role:   role,
		Scopes: []string{"orders:read", "orders:write"}, // TODO: загружается из БД/кэша
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token_generation_failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   900,
	})
}
