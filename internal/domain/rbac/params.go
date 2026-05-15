package rbac

// CreateRoleParams 描述创建角色时需要满足的领域输入。
type CreateRoleParams struct {
	AccountID      int
	Name           string
	IsDefaultOwner bool
}

// CreatePermissionGroupParams 描述创建权限组时需要满足的领域输入。
type CreatePermissionGroupParams struct {
	Code        string
	Name        string
	Description string
	SortOrder   int
}

// CreatePermissionParams 描述创建权限时需要满足的领域输入。
type CreatePermissionParams struct {
	Code        string
	APIPath     string
	Description string
}

// PermissionGroupPermissionsParams 描述权限组绑定权限时需要满足的领域输入。
type PermissionGroupPermissionsParams struct {
	PermissionGroupID int
	PermissionIDs     []int
}

// RolePermissionGroupsParams 描述角色绑定权限组时需要满足的领域输入。
type RolePermissionGroupsParams struct {
	AccountID          int
	RoleID             int
	PermissionGroupIDs []int
}

// UnitAccountRolesParams 描述单元授权账号角色时需要满足的领域输入。
type UnitAccountRolesParams struct {
	OwnerAccountID  int
	UnitID          int
	TargetAccountID int
	RoleIDs         []int
}
