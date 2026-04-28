package ports

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type AuthClaims struct {
	UserID string   `json:"sub"`
	Role   string   `json:"role"`
	Scopes []string `json:"scopes"`
	jwt.RegisteredClaims
}

type TokenManager interface {
	GenerateToken(ctx context.Context, claims AuthClaims) (string, error)
	ValidateToken(ctx context.Context, tokenString string) (*AuthClaims, error)
}
