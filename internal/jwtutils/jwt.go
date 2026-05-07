package jwtutils

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/superwhys/rent-manager-server/internal/errcode"
)

func GenerateTokenPair(signingKey []byte, accessTimeout, refreshTimeout time.Duration, userID uint) (string, string, string, error) {
	if len(signingKey) == 0 {
		return "", "", "", fmt.Errorf("signing key is required")
	}

	sessionID := uuid.NewString()
	now := time.Now()
	accessExpiresAt := now.Add(accessTimeout)
	refreshExpiresAt := now.Add(refreshTimeout)

	accessClaims := &AccountTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sessionID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
		},
		UserID:    userID,
		TokenType: TokenTypeAccess,
	}

	refreshClaims := &AccountTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sessionID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
		},
		UserID:    userID,
		TokenType: TokenTypeRefresh,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(signingKey)
	if err != nil {
		return "", "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(signingKey)
	if err != nil {
		return "", "", "", err
	}

	return accessToken, refreshToken, sessionID, nil
}

func GenerateAccessToken(signingKey []byte, timeout time.Duration, userID uint, sessionID string) (string, error) {
	if len(signingKey) == 0 {
		return "", fmt.Errorf("signing key is required")
	}
	if sessionID == "" {
		return "", fmt.Errorf("session id is required")
	}

	now := time.Now()
	expiresAt := now.Add(timeout)
	claims := &AccountTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sessionID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID:    userID,
		TokenType: TokenTypeAccess,
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func ParseToken(tokenStr string, signingKey []byte) (*AccountTokenClaims, error) {
	tc := &AccountTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, tc, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return tc, nil
}

type ContextKey string

const TokenContextKey ContextKey = "user_token_claims"

func TokenClaimsFromContext(ctx context.Context) (*AccountTokenClaims, error) {
	claims, ok := ctx.Value(TokenContextKey).(*AccountTokenClaims)
	if !ok {
		return nil, errcode.ErrUnauthorized
	}
	return claims, nil
}

func SetTokenClaimsToContext(ctx context.Context, claims *AccountTokenClaims) context.Context {
	return context.WithValue(ctx, TokenContextKey, claims)
}
