package jwtutils

import "github.com/golang-jwt/jwt/v5"

var ErrTokenExpired = jwt.ErrTokenExpired

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type AccountTokenClaims struct {
	jwt.RegisteredClaims
	UserID    uint   `json:"user_id"`
	TokenType string `json:"token_type"`
}
