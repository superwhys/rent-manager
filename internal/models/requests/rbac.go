package requests

type CreateRoleItem struct {
	Name string `json:"name" validate:"required"`
}

type CreateRolesCommand struct {
	Roles []CreateRoleItem `json:"roles" validate:"required,min=1,dive"`
}

type CreatePermissionItem struct {
	Code        string `json:"code" validate:"required"`
	APIPath     string `json:"api_path" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type CreatePermissionsCommand struct {
	Permissions []CreatePermissionItem `json:"permissions" validate:"required,min=1,dive"`
}

type RolePermissionsCommand struct {
	RoleID        int   `uri:"role_id" validate:"required,min=1"`
	PermissionIDs []int `json:"permission_ids" validate:"required,min=1,dive,min=1"`
}

type AccountRolesCommand struct {
	AccountID int   `uri:"account_id" validate:"required,min=1"`
	RoleIDs   []int `json:"role_ids" validate:"required,min=1,dive,min=1"`
}
