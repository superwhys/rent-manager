package requests

type CreateRoleItem struct {
	Name           string `json:"name" validate:"required"`
	IsDefaultOwner bool   `json:"is_default_owner"`
}

type CreateRolesCommand struct {
	OrganizationID int              `json:"organization_id" validate:"required,min=1"`
	Roles          []CreateRoleItem `json:"roles" validate:"required,min=1,dive"`
}

type CreatePermissionGroupItem struct {
	Code        string `json:"code" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type CreatePermissionGroupsCommand struct {
	PermissionGroups []CreatePermissionGroupItem `json:"permission_groups" validate:"required,min=1,dive"`
}

type CreatePermissionItem struct {
	Code        string `json:"code" validate:"required"`
	APIPath     string `json:"api_path" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type CreatePermissionsCommand struct {
	Permissions []CreatePermissionItem `json:"permissions" validate:"required,min=1,dive"`
}

type PermissionGroupPermissionsCommand struct {
	PermissionGroupID int   `uri:"permission_group_id" validate:"required,min=1"`
	PermissionIDs     []int `json:"permission_ids" validate:"required,min=1,dive,min=1"`
}

type RolePermissionGroupsCommand struct {
	OrganizationID     int   `json:"organization_id" validate:"required,min=1"`
	RoleID             int   `uri:"role_id" validate:"required,min=1"`
	PermissionGroupIDs []int `json:"permission_group_ids" validate:"required,min=1,dive,min=1"`
}

type AccountRolesCommand struct {
	OrganizationID int   `json:"organization_id" validate:"required,min=1"`
	AccountID      int   `uri:"account_id" validate:"required,min=1"`
	RoleIDs        []int `json:"role_ids" validate:"required,min=1,dive,min=1"`
}

type PermissionCheckRequest struct {
	AccountID      int
	OrganizationID int
	APIPath        string
	Version        int64
}
