package auth

// WechatLoginReq 表示微信登录请求参数。
type WechatLoginReq struct {
	Code string `json:"code" validate:"required"`
}

// RefreshTokenReq 表示刷新访问令牌请求参数。
type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenResult 描述认证用例签发的访问令牌。
type TokenResult struct {
	AccessToken  string
	RefreshToken string
}

// TokenResponse 是认证令牌响应。
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ToTokenResponse 将令牌结果转换为响应对象。
func ToTokenResponse(token *TokenResult) TokenResponse {
	if token == nil {
		return TokenResponse{}
	}
	return TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
}
