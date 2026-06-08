package jwt

import (
	"errors"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"git.superwhys.top/rent-manager/rent-manager-server/config"
	appauth "git.superwhys.top/rent-manager/rent-manager-server/internal/app/auth"
	"git.superwhys.top/rent-manager/rent-manager-server/internal/xerr"
)

// Manager 提供 JWT 签发和解析能力。
type Manager struct {
	signingKey     []byte
	accessTimeout  time.Duration
	refreshTimeout time.Duration
}

// NewManager 创建 JWT 管理器。
func NewManager(jwtConfig *config.JwtConfig) *Manager {
	return &Manager{
		signingKey:     []byte(jwtConfig.JwtSecret),
		accessTimeout:  jwtConfig.JwtTimeout,
		refreshTimeout: jwtConfig.JwtRefreshTimeout,
	}
}

// GenerateTokenPair 生成访问令牌和刷新令牌。
func (m *Manager) GenerateTokenPair(userID int) (*appauth.TokenPair, error) {
	if len(m.signingKey) == 0 {
		return nil, fmt.Errorf("signing key is required")
	}

	sessionID := uuid.NewString()
	now := time.Now()
	accessExpiresAt := now.Add(m.accessTimeout)
	refreshExpiresAt := now.Add(m.refreshTimeout)

	accessClaims := &accountTokenClaims{
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   sessionID,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(accessExpiresAt),
		},
		UserID:    userID,
		TokenType: appauth.TokenTypeAccess,
	}

	refreshClaims := &accountTokenClaims{
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   sessionID,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(refreshExpiresAt),
		},
		UserID:    userID,
		TokenType: appauth.TokenTypeRefresh,
	}

	accessToken, err := gojwt.NewWithClaims(gojwt.SigningMethodHS256, accessClaims).SignedString(m.signingKey)
	if err != nil {
		return nil, err
	}

	refreshToken, err := gojwt.NewWithClaims(gojwt.SigningMethodHS256, refreshClaims).SignedString(m.signingKey)
	if err != nil {
		return nil, err
	}

	return &appauth.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// GenerateAccessToken 生成访问令牌。
func (m *Manager) GenerateAccessToken(userID int, sessionID string) (string, error) {
	if len(m.signingKey) == 0 {
		return "", fmt.Errorf("signing key is required")
	}
	if sessionID == "" {
		return "", fmt.Errorf("session id is required")
	}

	now := time.Now()
	expiresAt := now.Add(m.accessTimeout)
	claims := &accountTokenClaims{
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   sessionID,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(expiresAt),
		},
		UserID:    userID,
		TokenType: appauth.TokenTypeAccess,
	}

	return gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims).SignedString(m.signingKey)
}

// ParseToken 解析 JWT 并返回应用层令牌声明。
func (m *Manager) ParseToken(tokenStr string) (*appauth.TokenClaims, error) {
	claims := &accountTokenClaims{}
	token, err := gojwt.ParseWithClaims(tokenStr, claims, func(token *gojwt.Token) (any, error) {
		if token.Method != gojwt.SigningMethodHS256 {
			return nil, gojwt.ErrSignatureInvalid
		}
		return m.signingKey, nil
	})
	if err != nil {
		if errors.Is(err, gojwt.ErrTokenExpired) {
			return nil, xerr.ErrTokenExpired
		}
		return nil, err
	}

	if !token.Valid {
		return nil, xerr.ErrInvalidToken
	}

	return &appauth.TokenClaims{
		UserID:    claims.UserID,
		TokenType: claims.TokenType,
		Subject:   claims.Subject,
	}, nil
}
