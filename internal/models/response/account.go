package response

import (
	"time"

	domainaccount "git.superwhys.top/rent-manager/rent-manager-server/internal/domain/account"
)

// AccountResponse 是账号信息响应。
type AccountResponse struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	WechatOpenID string    `json:"wechat_openid"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ToAccountResponse 将账号领域对象转换为响应对象。
func ToAccountResponse(account *domainaccount.Account) AccountResponse {
	if account == nil {
		return AccountResponse{}
	}
	return AccountResponse{
		ID:           account.ID,
		Name:         account.Name,
		Email:        account.Email,
		WechatOpenID: account.WechatOpenID,
		IsActive:     account.IsActive,
		CreatedAt:    account.CreatedAt,
		UpdatedAt:    account.UpdatedAt,
	}
}
