package auth

// TokenTypeAccess 表示访问令牌。
const TokenTypeAccess = "access"

// TokenTypeRefresh 表示刷新令牌。
const TokenTypeRefresh = "refresh"

// TokenPair 表示签发的访问令牌和刷新令牌。
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// TokenClaims 表示认证令牌中的应用层声明。
type TokenClaims struct {
	UserID    int
	TokenType string
	Subject   string
}

// TokenManager 提供认证令牌签发和解析能力。
type TokenManager interface {
	GenerateTokenPair(userID int) (*TokenPair, error)
	GenerateAccessToken(userID int, sessionID string) (string, error)
	ParseToken(token string) (*TokenClaims, error)
}
