package jwt

import gojwt "github.com/golang-jwt/jwt/v5"

// accountTokenClaims 表示 JWT 中保存的账号声明。
type accountTokenClaims struct {
	gojwt.RegisteredClaims
	UserID    int    `json:"user_id"`
	TokenType string `json:"token_type"`
}
