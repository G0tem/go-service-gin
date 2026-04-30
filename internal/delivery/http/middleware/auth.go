package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/G0tem/go-service-gin/internal/domain/ports"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type claimsKey struct{}

// JWTAuth возвращает middleware для проверки Bearer-токена
func JWTAuth(tm ports.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_authorization"})
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_authorization_format"})
			return
		}

		claims, err := tm.ValidateToken(c.Request.Context(), parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_or_expired_token"})
			return
		}

		// Безопасная запись в контекст
		ctx := context.WithValue(c.Request.Context(), claimsKey{}, claims)
		c.Request = c.Request.WithContext(ctx)

		// 🏷️ OTel атрибуты
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(
			attribute.String("auth.user_id", claims.UserID),
			attribute.String("auth.role", claims.Role),
		)

		c.Next()
	}
}

// GetClaims извлекает проверенные claims из контекста
func GetClaims(c *gin.Context) *ports.AuthClaims {
	if v, ok := c.Request.Context().Value(claimsKey{}).(*ports.AuthClaims); ok {
		return v
	}
	return nil
}
