package requests

// CreateOrganizationCommand 是平台管理员创建组织的请求体。
type CreateOrganizationCommand struct {
	Name                    string `json:"name" validate:"required,max=255"`
	AdminPermissionGroupIDs []int  `json:"admin_permission_group_ids" validate:"omitempty,dive,min=1"`
}

// BindAccountToOrganizationCommand 是平台管理员绑定账号到组织的请求参数。
type BindAccountToOrganizationCommand struct {
	OrganizationID int `uri:"organization_id" validate:"required,min=1"`
	AccountID      int `uri:"account_id" validate:"required,min=1"`
}
