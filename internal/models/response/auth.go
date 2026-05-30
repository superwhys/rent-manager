package response

import (
	"git.superwhys.top/rent-manager/rent-manager-server/internal/app/services"
	"github.com/miebyte/goutils/ginutils"
)

// TokenResponse 是认证令牌响应。
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func ToTokenResponseRet(token *services.TokenResult) *ginutils.Ret[TokenResponse] {
	return ResponseWithData(TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
}
