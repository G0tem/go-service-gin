package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RequireRole проверяет, что роль пользователя совпадает с одной из разрешённых.
// Должен идти ПОСЛЕ middleware.JWTAuth.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access_denied"})
			return
		}

		for _, role := range allowedRoles {
			if claims.Role == role { // точное совпадение (case-sensitive)
				trace.SpanFromContext(c.Request.Context()).SetAttributes(
					attribute.String("auth.required_roles", strings.Join(allowedRoles, ",")),
					attribute.Bool("auth.access_granted", true),
				)
				c.Next()
				return
			}
		}

		trace.SpanFromContext(c.Request.Context()).SetAttributes(
			attribute.String("auth.user_role", claims.Role),
			attribute.String("auth.required_roles", strings.Join(allowedRoles, ",")),
			attribute.Bool("auth.access_granted", false),
		)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient_role"})
	}
}

// RequireScope проверяет, что токен содержит ВСЕ указанные скоупы.
// Должен идти ПОСЛЕ middleware.JWTAuth.
func RequireScope(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access_denied"})
			return
		}

		// O(1) поиск для скоупов
		for _, req := range requiredScopes {
			if !slices.Contains(claims.Scopes, req) {
				trace.SpanFromContext(c.Request.Context()).SetAttributes(
					attribute.String("auth.missing_scope", req),
					attribute.Bool("auth.access_granted", false),
				)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient_scope"})
				return
			}
		}

		trace.SpanFromContext(c.Request.Context()).SetAttributes(
			attribute.String("auth.required_scopes", strings.Join(requiredScopes, ",")),
			attribute.Bool("auth.access_granted", true),
		)
		c.Next()
	}
}
