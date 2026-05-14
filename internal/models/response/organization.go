package response

import domainaccount "github.com/superwhys/rent-manager-server/internal/domain/account"

// OrganizationResponse 是组织响应。
type OrganizationResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

// ToOrganizationResponse 将组织领域对象转换为响应对象。
func ToOrganizationResponse(organization *domainaccount.Organization) *OrganizationResponse {
	if organization == nil {
		return nil
	}
	return &OrganizationResponse{
		ID:       organization.ID,
		Name:     organization.Name,
		IsActive: organization.IsActive,
	}
}
